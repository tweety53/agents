package harvest

// ZCode writes every API call its sessions make as one JSON line --
// type "model_io" -- under ~/.zcode/cli/rollout/model-io-sess_<uuid>.jsonl.
// This file parses those lines into the same Record and CommandRecord shapes
// the Claude parser (transcript.go) produces, so attribution, session-token
// binding and pricing work unchanged over either source. The shapes are
// measured against real rollout files (2026-09-09), not assumed; every field
// this package does not need (requestId, durationMs, querySource, traceId,
// turnId, attempt, the request body's bulk) is simply never named, the same
// tolerance rawLine's doc comment commits to for the Claude format.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultZcodeRolloutRootEnv, when set, overrides DefaultZcodeRolloutRoot --
// a test's own isolated root, the same pattern
// DefaultTranscriptsRootEnv's FLOW_TRANSCRIPTS_DIR already uses. Like
// FLOW_TRANSCRIPTS_DIR, this variable is deliberately not workspace-isolated
// (kan-388's recorded decision): the rollout root is a read-only input, and
// every worktree's harvested data lands in that worktree's own isolated
// database, so the source stays shared while the written metrics do not.
const DefaultZcodeRolloutRootEnv = "FLOW_ZCODE_ROLLOUTS_DIR"

// DefaultZcodeRolloutRoot returns the directory ZCode writes its model-io
// rollout files under: DefaultZcodeRolloutRootEnv when set, otherwise
// "~/.zcode/cli/rollout" resolved against the current user's home directory.
func DefaultZcodeRolloutRoot() (string, error) {
	if v := os.Getenv(DefaultZcodeRolloutRootEnv); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("harvest: resolve home directory: %w", err)
	}
	return filepath.Join(home, ".zcode", "cli", "rollout"), nil
}

// recordTypeModelIO is the only top-level "type" this package extracts
// anything from. The measured rollout files carry no other value, but the
// set is not closed and a shape this package does not recognise is skipped,
// never an error -- the same discipline ParseAssistantRecords applies to the
// Claude transcript format.
const recordTypeModelIO = "model_io"

// rolloutSubagentSessionPrefix is the per-dispatch session id prefix ZCode
// stamps on every rollout line a subagent dispatch produces (measured
// 2026-09-12 against a live capture, model-io-sess_subagent_agent_<uuid>.jsonl):
// one rollout file per dispatch, sessionId sess_subagent_agent_<uuid> on every
// line, querySource "subagent", and the <uuid> byte-identical to the agentId
// the dispatcher's Agent tool result returns -- which `flow record dispatch
// begin -agent-id` already records on the dispatch row. That session id is
// the per-dispatch identity cost attribution keys on, exactly, with no
// timestamp inference involved (KAN-506): a record carrying it IS that
// dispatch's spend, no matter how many sibling dispatches run concurrently.
const rolloutSubagentSessionPrefix = "sess_subagent_agent_"

// agentIDFromRolloutSessionID reports whether sessionID names a subagent
// dispatch's own session and returns that dispatch's agent id -- the
// dispatcher-side form (agent_<uuid>) the dispatches.agent_id column carries.
// The mapping is prefix-stripping, not parsing: sess_ + subagent_ + the agent
// id, so a main-session id (sess_<uuid>) and anything else simply do not
// match.
func agentIDFromRolloutSessionID(sessionID string) (string, bool) {
	rest, ok := strings.CutPrefix(sessionID, rolloutSubagentSessionPrefix)
	if !ok || rest == "" {
		return "", false
	}
	return "agent_" + rest, true
}

// rawRolloutLine and its nested types are the minimal decode shape this
// package needs from one rollout line. encoding/json drops everything else
// silently; DisallowUnknownFields is deliberately not used, for the same
// reason transcript.go's rawLine declines it -- this is a format the package
// only reads and never writes.
type rawRolloutLine struct {
	Type        string              `json:"type"`
	SessionID   string              `json:"sessionId"`
	CompletedAt string              `json:"completedAt"`
	Model       *rawRolloutModel    `json:"model"`
	Request     *rawRolloutRequest  `json:"request"`
	Response    *rawRolloutResponse `json:"response"`
}

type rawRolloutModel struct {
	ModelID string `json:"modelId"`
}

// rawRolloutRequest is the history half of the line: every earlier message of
// the session, replayed on every request. Its tool calls are one of the two
// places a stage mark's command text appears.
type rawRolloutRequest struct {
	Messages []rawRolloutMessage `json:"messages"`
}

type rawRolloutMessage struct {
	ToolCalls []rawRolloutToolCall `json:"toolCalls"`
}

// rawRolloutResponse is the response half: this call's own usage, and the
// tool calls the model issued -- the other place a mark's command appears.
type rawRolloutResponse struct {
	Usage     *rawRolloutUsage     `json:"usage"`
	ToolCalls []rawRolloutToolCall `json:"toolCalls"`
}

type rawRolloutToolCall struct {
	Name  string                   `json:"name"`
	Input *rawRolloutToolCallInput `json:"input"`
}

type rawRolloutToolCallInput struct {
	Command string `json:"command"`
}

type rawRolloutUsage struct {
	InputTokens      int64 `json:"inputTokens"`
	OutputTokens     int64 `json:"outputTokens"`
	CacheReadTokens  int64 `json:"cacheReadTokens"`
	CacheWriteTokens int64 `json:"cacheWriteTokens"`
}

// ParseRolloutRecords decodes every model_io record found in complete (as
// SplitCompleteLines defines "complete") into a Record, in file order. One
// line is one API response -- there is no per-content-block dedup here, and
// none is needed: unlike Claude's transcript, which writes one line per
// content block of a single response repeating its message.id, a rollout
// line's usage is one call's own usage and each line is read exactly once
// through the offset.
//
// Every unparseable or unrecognised line is skipped rather than fatal, the
// same tolerance ParseAssistantRecords applies: the format is written by a
// harness across versions this package does not control. A line whose
// completedAt fails to parse as RFC 3339 is skipped too -- it cannot be
// attributed to any window without a timestamp.
//
// The cache split is never known on this format: ZCode reports a single
// collapsed cacheWriteTokens with no 5m/1h split, so every record lands with
// CacheSplitKnown false and its cache-creation total in the unknown split
// (Bucket.add), where store.Store.Price can see it -- and prices it exactly
// under the flat-rate rule the store's own pricing rule provides.
//
// A line carrying a subagent dispatch's per-dispatch session id
// (sess_subagent_agent_<uuid>, rolloutSubagentSessionPrefix) yields a record
// marked IsSidechain with that dispatch's own agent id -- the identity
// attribution keys on instead of the dispatch's time window, so concurrent
// dispatches neither lose nor blend their figures (KAN-506). A main-session
// line marks neither field, unchanged.
func ParseRolloutRecords(complete []byte) []Record {
	var out []Record
	forEachRolloutLine(complete, func(raw rawRolloutLine) {
		if raw.Type != recordTypeModelIO || raw.Response == nil || raw.Response.Usage == nil {
			return
		}
		ts, err := time.Parse(time.RFC3339, raw.CompletedAt)
		if err != nil {
			return
		}
		u := raw.Response.Usage
		var model string
		if raw.Model != nil {
			model = raw.Model.ModelID
		}
		rec := Record{
			Timestamp: ts,
			SessionID: raw.SessionID,
			Model:     model,
			Usage: Usage{
				InputTokens:              u.InputTokens,
				CacheCreationInputTokens: u.CacheWriteTokens,
				CacheReadInputTokens:     u.CacheReadTokens,
				OutputTokens:             u.OutputTokens,
			},
		}
		if agentID, ok := agentIDFromRolloutSessionID(raw.SessionID); ok {
			rec.IsSidechain = true
			rec.AgentID = agentID
		}
		out = append(out, rec)
	})
	return out
}

// ParseRolloutCommandRecords decodes every Bash tool-call command found in
// complete's lines, from both places the format records them:
// request.messages[*].toolCalls[*] -- the session's history, replayed on
// every request, which is why a mark's command text keeps appearing line
// after line -- and response.toolCalls[*], the call's own tool use. Both
// carry the same line-level sessionId, which is what CommandRecord needs.
//
// Like ParseCommandRecords, it does not require the line to carry a usage
// object: session-token resolution must work from any model_io line. A tool
// call that is not Bash, or whose input does not decode, contributes nothing.
func ParseRolloutCommandRecords(complete []byte) []CommandRecord {
	var out []CommandRecord
	forEachRolloutLine(complete, func(raw rawRolloutLine) {
		if raw.Type != recordTypeModelIO {
			return
		}
		collect := func(calls []rawRolloutToolCall) {
			for _, call := range calls {
				if call.Name != "Bash" || call.Input == nil || call.Input.Command == "" {
					continue
				}
				out = append(out, CommandRecord{SessionID: raw.SessionID, Command: call.Input.Command})
			}
		}
		if raw.Request != nil {
			for _, msg := range raw.Request.Messages {
				collect(msg.ToolCalls)
			}
		}
		if raw.Response != nil {
			collect(raw.Response.ToolCalls)
		}
	})
	return out
}

// ReadRolloutNewRecords reads path from offset to EOF -- the rollout
// source's counterpart of ReadNewRecords (transcript.go), sharing that
// function's offset discipline verbatim: a partial trailing line is split
// off and never counted as consumed, and a file shorter than offset is
// ErrOffsetBeyondEOF, never silently reread from 0.
func ReadRolloutNewRecords(path string, offset int64) ([]Record, []CommandRecord, int64, error) {
	f, err := openAt(path, offset)
	if err != nil {
		return nil, nil, offset, err
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, offset, fmt.Errorf("harvest: read %s from offset %d: %w", path, offset, err)
	}

	complete, _ := SplitCompleteLines(raw)
	return ParseRolloutRecords(complete), ParseRolloutCommandRecords(complete), offset + int64(len(complete)), nil
}

// ReadRolloutAllCommands reads path from byte 0 to its current EOF and
// returns every Bash command found in it -- the rollout source's
// counterpart of ReadAllCommands (transcript.go), for exactly that
// function's one caller: the retried-give-up scan must find a mark sitting
// in a rollout file whose committed offset already equals its full length.
// Like its counterpart it computes no Record, so there is nothing here a
// caller could re-attribute, and it neither reads nor advances
// harvest_offsets for path.
func ReadRolloutAllCommands(path string) ([]CommandRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("harvest: open %s: %w", path, err)
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("harvest: read %s: %w", path, err)
	}

	complete, _ := SplitCompleteLines(raw)
	return ParseRolloutCommandRecords(complete), nil
}

// forEachRolloutLine walks complete line by line -- complete always ends in
// '\n' -- decoding each into rawRolloutLine and handing every successfully
// decoded line to fn. A line that fails to decode is skipped, never fatal.
func forEachRolloutLine(complete []byte, fn func(rawRolloutLine)) {
	start := 0
	for start < len(complete) {
		idx := bytes.IndexByte(complete[start:], '\n')
		if idx < 0 {
			break // complete always ends in '\n'; unreachable in practice.
		}
		end := start + idx + 1
		line := bytes.TrimSpace(complete[start:end])
		start = end
		if len(line) == 0 {
			continue
		}
		var raw rawRolloutLine
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		fn(raw)
	}
}
