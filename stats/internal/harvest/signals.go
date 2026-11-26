package harvest

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SignalKind names which observability signal a Signal carries -- the
// same closed set design.md's "Harvesting" section lists for
// attribution (task 5): a compaction, a completed turn, a tool call
// starting, a tool call's result (including a denial), or a synthetic
// API-error assistant line.
type SignalKind string

const (
	SignalCompaction SignalKind = "compaction"
	SignalTurn       SignalKind = "turn"
	SignalToolUse    SignalKind = "tool_use"
	SignalToolResult SignalKind = "tool_result"
	SignalAPIError   SignalKind = "api_error"
)

// Compaction is a compact_boundary system line's own figures -- what
// triggered it and the token counts and wall-clock duration it cost.
type Compaction struct {
	Trigger    string
	PreTokens  int64
	PostTokens int64
	DurationMs int64
}

// Signal is one observability event ParseSignalRecords extracted from a
// transcript line, riding on the same Record every usage record uses --
// only the fields its Kind actually carries are populated, the rest
// stay zero.
type Signal struct {
	Kind     SignalKind
	ToolName string // SignalToolUse
	IsError  bool   // SignalToolResult
	// ToolUseID, Tool, Input, Error and Fix describe a failed call a
	// SignalToolResult carries (see toolErrorFields); empty on a success or a
	// failure whose text or input is withheld.
	ToolUseID, Tool, Input, Error, Fix string
	Denied                             bool  // SignalToolResult
	DurationMs                         int64 // SignalTurn
	Messages                           int64 // SignalTurn
	Compaction                         *Compaction
}

// CompactionEvent is one compaction's own figures, keyed in Signals by
// its record's timestamp (RFC3339Nano) so repeated compactions within one
// stage run each get their own entry rather than overwriting each other.
type CompactionEvent struct {
	Trigger    string `json:"trigger"`
	PreTokens  int64  `json:"pre_tokens"`
	PostTokens int64  `json:"post_tokens"`
	DurationMs int64  `json:"duration_ms"`
}

// ToolErrorEvent is one failed tool call a self-review bundle shows: the call,
// its trimmed error and, when the same tool succeeded next in the same read,
// the call that fixed it.
type ToolErrorEvent struct {
	Tool  string `json:"tool"`
	Input string `json:"input"`
	Error string `json:"error"`
	Fix   string `json:"fix,omitempty"`
}

// maxToolErrorEvents bounds the events one Signals bucket accumulates per read.
const maxToolErrorEvents = 5

// Signals is one main/sidechain bucket's (or one dispatch's) observability
// counters, accumulated by add from a batch's signal and usage records --
// SignalsDelta's own doc comment says how the two buckets combine into a
// stage run's metrics bag.
type Signals struct {
	Compactions      int64                      `json:"compactions,omitempty"`
	CompactionEvents map[string]CompactionEvent `json:"compaction_events,omitempty"` // key: RFC3339Nano timestamp
	Turns            int64                      `json:"turns,omitempty"`
	TurnDurationMs   int64                      `json:"turn_duration_ms,omitempty"`
	TurnMessages     int64                      `json:"turn_messages,omitempty"`
	ToolCalls        map[string]int64           `json:"tool_calls,omitempty"`
	ToolCallsTotal   int64                      `json:"tool_calls_total,omitempty"`
	ToolErrors       int64                      `json:"tool_errors,omitempty"`
	ToolErrorEvents  map[string]ToolErrorEvent  `json:"tool_error_events,omitempty"` // key: tool_use id
	Denials          int64                      `json:"denials,omitempty"`
	APIErrors        int64                      `json:"api_errors,omitempty"`
	ContextEnd       string                     `json:"context_end,omitempty"`
	ServedModels     map[string]int64           `json:"served_models,omitempty"`
	ServedEfforts    map[string]int64           `json:"served_efforts,omitempty"`
	contextEndAt     time.Time
}

// SignalsDelta splits Signals by IsSidechain exactly as TokenDelta splits
// tokens -- Main for a stage run's own thread, Sidechain for every
// subagent transcript line attributed to it.
type SignalsDelta struct {
	Main      Signals `json:"main"`
	Sidechain Signals `json:"sidechain"`
}

// add folds one record into s. A usage record contributes its served
// model and effort and, when it is the latest seen, the context size at
// that message; a signal record contributes its own counters.
//
// A synthetic API-error line (Model == "<synthetic>") is excluded from
// both ServedModels and ContextEnd -- it carries an all-zero Usage, so
// letting it win the "latest seen" race would zero out context_end for
// good the moment it is a batch's last record: jsonb_deep_add's
// last-write-wins string replace has no earlier value to fall back to.
func (s *Signals) add(r Record) {
	if r.Signal == nil {
		if r.Model != "" && r.Model != "<synthetic>" {
			bump(&s.ServedModels, strings.ToLower(r.Model))
		}
		if r.Effort != "" {
			bump(&s.ServedEfforts, r.Effort)
		}
		if r.Model != "<synthetic>" && r.Timestamp.After(s.contextEndAt) {
			s.contextEndAt = r.Timestamp
			u := r.Usage
			s.ContextEnd = strconv.FormatInt(u.InputTokens+u.CacheReadInputTokens+u.CacheCreationInputTokens, 10)
		}
		return
	}
	switch r.Signal.Kind {
	case SignalCompaction:
		s.Compactions++
		if c := r.Signal.Compaction; c != nil {
			if s.CompactionEvents == nil {
				s.CompactionEvents = map[string]CompactionEvent{}
			}
			s.CompactionEvents[r.Timestamp.UTC().Format(time.RFC3339Nano)] = CompactionEvent{
				Trigger: c.Trigger, PreTokens: c.PreTokens, PostTokens: c.PostTokens, DurationMs: c.DurationMs,
			}
		}
	case SignalTurn:
		s.Turns++
		s.TurnDurationMs += r.Signal.DurationMs
		s.TurnMessages += r.Signal.Messages
	case SignalToolUse:
		bump(&s.ToolCalls, r.Signal.ToolName)
		s.ToolCallsTotal++
	case SignalToolResult:
		if r.Signal.IsError {
			s.ToolErrors++
			if r.Signal.Tool != "" && r.Signal.Error != "" && len(s.ToolErrorEvents) < maxToolErrorEvents {
				if s.ToolErrorEvents == nil {
					s.ToolErrorEvents = map[string]ToolErrorEvent{}
				}
				s.ToolErrorEvents[r.Signal.ToolUseID] = ToolErrorEvent{
					Tool: r.Signal.Tool, Input: r.Signal.Input, Error: r.Signal.Error, Fix: r.Signal.Fix,
				}
			}
		}
		if r.Signal.Denied {
			s.Denials++
		}
	case SignalAPIError:
		s.APIErrors++
	}
}

func bump(m *map[string]int64, key string) {
	if *m == nil {
		*m = map[string]int64{}
	}
	(*m)[key]++
}

func (d *SignalsDelta) bucket(sidechain bool) *Signals {
	if sidechain {
		return &d.Sidechain
	}
	return &d.Main
}

// DenialPrefixes are the literal openings of a tool_result a harness writes
// when a call was refused rather than run. Extend only with a transcript
// sample added to testdata/transcripts/signals.jsonl beside it.
var DenialPrefixes = []string{
	"Permission for this action was denied",
	"The user doesn't want to proceed",
	"PreToolUse hook",
}

func isDenial(text string) bool {
	for _, p := range DenialPrefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}
	return false
}

// toolResultText flattens a tool_result's content, which is either a
// string or an array of {type:"text", text} blocks, into one string for
// the denial prefix check.
func toolResultText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var blocks []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var b strings.Builder
	for _, blk := range blocks {
		b.WriteString(blk.Text)
	}
	return b.String()
}

// ParseSignalRecords is ParseAssistantRecords' sibling over the same bytes:
// it yields one Record per observability signal -- a compaction, a turn,
// a tool call, a tool result, an API error -- with Usage zero and Signal
// set. tool_use blocks are deduplicated by block id and tool_result
// blocks by tool_use_id, since one API response is written as one line
// per content block and each block id appears once in a healthy
// transcript; a line repeated verbatim (a re-read overlap) is then counted
// once.
func ParseSignalRecords(complete []byte) []Record {
	var out []Record
	seenTool := map[string]bool{}
	uses := map[string]rawContentBlock{}
	// ponytail: a fix is paired only within one read of the transcript; add
	// cross-read state if a failure's fix lands in a later batch.
	pending := map[string]int{} // tool name -> index in out of its latest unfixed failure
	start := 0
	for start < len(complete) {
		idx := bytes.IndexByte(complete[start:], '\n')
		if idx < 0 {
			break
		}
		end := start + idx + 1
		line := bytes.TrimSpace(complete[start:end])
		start = end
		if len(line) == 0 {
			continue
		}
		var raw rawLine
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			continue
		}
		base := Record{Timestamp: ts, SessionID: raw.SessionID, IsSidechain: raw.IsSidechain, AgentID: raw.AgentID}
		emit := func(s Signal) {
			r := base
			r.Signal = &s
			out = append(out, r)
		}

		switch raw.Type {
		case "system":
			switch raw.Subtype {
			case "compact_boundary":
				c := &Compaction{}
				if raw.CompactMetadata != nil {
					c.Trigger = raw.CompactMetadata.Trigger
					c.PreTokens = raw.CompactMetadata.PreTokens
					c.PostTokens = raw.CompactMetadata.PostTokens
					c.DurationMs = raw.CompactMetadata.DurationMs
				}
				emit(Signal{Kind: SignalCompaction, Compaction: c})
			case "turn_duration":
				emit(Signal{Kind: SignalTurn, DurationMs: raw.DurationMs, Messages: raw.MessageCount})
			}
		case recordTypeAssistant:
			if raw.IsAPIErrorMessage {
				emit(Signal{Kind: SignalAPIError})
			}
			if raw.Message == nil {
				continue
			}
			for _, blk := range contentBlocks(raw.Message.Content) {
				if blk.Type != "tool_use" || blk.ID == "" || seenTool["use:"+blk.ID] {
					continue
				}
				seenTool["use:"+blk.ID] = true
				uses[blk.ID] = blk
				emit(Signal{Kind: SignalToolUse, ToolName: blk.Name})
			}
		case "user":
			if raw.Message == nil {
				continue
			}
			for _, blk := range contentBlocks(raw.Message.Content) {
				if blk.Type != "tool_result" || blk.ToolUseID == "" || seenTool["result:"+blk.ToolUseID] {
					continue
				}
				seenTool["result:"+blk.ToolUseID] = true
				text := toolResultText(blk.Content)
				sig := Signal{Kind: SignalToolResult, IsError: blk.IsError, Denied: blk.IsError && isDenial(text)}
				use, known := uses[blk.ToolUseID]
				if blk.IsError && known {
					sig.ToolUseID, sig.Tool = blk.ToolUseID, use.Name
					sig.Input, sig.Error = toolErrorFields(use.Input, text)
				}
				if !blk.IsError && known {
					if i, ok := pending[use.Name]; ok {
						if in, ok := sanitizeToolText(toolInputSummary(use.Input)); ok {
							out[i].Signal.Fix = in
						}
						delete(pending, use.Name)
					}
				}
				emit(sig)
				if sig.Error != "" {
					pending[sig.Tool] = len(out) - 1
				}
			}
		}
	}
	return out
}

const (
	maxToolInputChars = 200
	maxToolErrorChars = 300
)

// toolInputSummary is the Bash command or, for any other tool, the call's
// input JSON.
func toolInputSummary(raw json.RawMessage) string {
	var b rawBashInput
	if json.Unmarshal(raw, &b) == nil && b.Command != "" {
		return b.Command
	}
	return string(raw)
}

// toolErrorFields is a failed call's input summary and error text, each
// redacted and trimmed; both empty when either touches a credential file,
// so the pair is withheld whole.
func toolErrorFields(input json.RawMessage, errText string) (in, msg string) {
	in, ok1 := sanitizeToolText(toolInputSummary(input))
	msg, ok2 := sanitizeToolText(errText)
	if !ok1 || !ok2 {
		return "", ""
	}
	return trimChars(in, maxToolInputChars), trimChars(msg, maxToolErrorChars)
}

var (
	// sensitivePath matches the credential locations a bundle never carries.
	sensitivePath = regexp.MustCompile(`(?i)(^|[/\s'"=~])\.env\b|\.claude\b|\.config\b|credentials|\.pem\b|id_rsa|\.netrc|\.aws\b`)
	// secretValue matches an assignment or header carrying a secret value.
	secretValue = regexp.MustCompile(`(?i)((token|secret|passw(or)?d|api[_-]?key|authorization|bearer)["']?\s*[=:]?\s*(?:bearer\s+)?)[^\s'"]+|sk-[A-Za-z0-9_-]{12,}|gh[pousr]_[A-Za-z0-9]{20,}`)
)

// sanitizeToolText redacts secret values and reports false when s names a
// credential, .env, ~/.claude or ~/.config path.
func sanitizeToolText(s string) (string, bool) {
	if sensitivePath.MatchString(s) {
		return "", false
	}
	return secretValue.ReplaceAllStringFunc(s, func(m string) string {
		if loc := secretValue.FindStringSubmatchIndex(m); loc != nil && loc[3] > 0 {
			return m[:loc[3]] + "[redacted]"
		}
		return "[redacted]"
	}), true
}

func trimChars(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}
