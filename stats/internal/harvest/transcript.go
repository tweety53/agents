// Package harvest turns Claude Code's own session transcripts into the
// token, model and effort metrics design.md's "Harvesting" section
// describes: it parses ~/.claude/projects/*/*.jsonl (and its per-session
// subagents/*.jsonl files, which carry sidechain messages under the same
// top-level sessionId), attributes each assistant message to the open
// stage window whose session matches and whose [started_at, ended_at)
// interval contains the message's timestamp, and commits the resulting
// per-batch token deltas back through internal/store's
// CommitHarvestBatch, atomically alongside the transcript's newly
// consumed byte offset.
//
// This package never imports internal/store. It depends on two narrow,
// consumer-defined interfaces (WindowSource, attribute.go; HarvestSink,
// watcher.go) that a real *store.Store happens to satisfy once the
// daemon wires it in, so TestHarvestNeedsNoDatabase can exercise the
// whole attribution and commit path against fakes, with no PostgreSQL
// running -- exactly what tasks.md's "internal/harvest must not depend on
// the store directly" and go-interface-design's consumer-side-interface
// rule both require.
package harvest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// ErrOffsetBeyondEOF is returned by ReadNewRecords when offset is past the
// current end of path -- the file was rotated, truncated, or replaced out
// from under the watcher. Re-reading from 0 in that case would attribute
// the same bytes twice under a new identity; refusing surfaces the
// anomaly instead.
var ErrOffsetBeyondEOF = errors.New("harvest: offset is beyond the end of the transcript file")

// recordTypeAssistant is the only top-level "type" this package extracts
// anything from. Every other value -- "user", "attachment",
// "queue-operation", "file-history-snapshot", "system", "mode",
// "permission-mode", "ai-title", "last-prompt", "file-history-delta", and
// whatever future type Claude Code adds -- is skipped, never an error:
// design.md's transcript shape is confirmed only for assistant records,
// and the harvester must tolerate an open-ended set of everything else
// (this task's own instructions: "do not assume the set is closed").
const recordTypeAssistant = "assistant"

// Usage is one assistant message's token accounting, read verbatim from
// its "usage" object. A key legitimately absent from the JSON (for
// example output_tokens_details on a message with no thinking) decodes
// to zero here -- and that zero is a real, measured token count for that
// message, not a stand-in for "not measured". The unmeasured case this
// change's absence-is-not-a-value rule actually guards is per-harness
// (task 10's tokens_available flag on a harness that writes no
// transcript at all), not per-field on a message the harvester did read.
//
// CacheCreationInputTokens keeps its original meaning -- the collapsed
// total task 9 already read -- and stays populated whether or not the
// split below is present, so nothing that only ever read the total loses
// it. CacheCreation5mTokens and CacheCreation1hTokens are the two rates
// Anthropic actually charges (task 23): a 5-minute cache write costs
// 1.25x base input, a 1-hour write costs 2x, and the collapsed total
// alone cannot say which applied. CacheSplitKnown is false when the
// message's "usage" object carried no "cache_creation" split object at
// all -- a real, older or differently-shaped transcript line, not zero
// cache-creation usage -- and is the fact store.Store.Price (via
// attribute.go's Bucket) relies on to refuse pricing that portion rather
// than guessing which rate applied.
type Usage struct {
	InputTokens              int64
	CacheCreationInputTokens int64
	CacheCreation5mTokens    int64
	CacheCreation1hTokens    int64
	CacheSplitKnown          bool
	CacheReadInputTokens     int64
	OutputTokens             int64
	ThinkingTokens           int64
	// Speed is the message's own "speed" field verbatim -- "standard",
	// "fast", or "" when the transcript line carried none (an older
	// format, or a harness that never sets it). Fast mode doubles Opus's
	// per-token rate (task 23) and is otherwise indistinguishable from
	// standard speed at the same model, so it must survive from the raw
	// transcript line through to pricing rather than being discarded the
	// way it was before this task.
	Speed string
}

// Record is the one shape this package extracts from a transcript line:
// exactly the fields design.md's "Harvesting" section names --
// timestamp, message.model, effort, isSidechain and usage -- plus the
// session id attribution matches against. Every non-assistant line
// parses to nothing (ParseAssistantRecords simply omits it), so Record
// never needs to represent "not an assistant message".
type Record struct {
	Timestamp   time.Time
	SessionID   string
	IsSidechain bool
	Model       string
	Effort      string
	Usage       Usage
}

// rawLine and its nested types are the minimal decode shape this package
// needs from a transcript line. Every other field the real format carries
// (parentUuid, uuid, cwd, gitBranch, toolUseResult, and dozens more) is
// simply never named here, so encoding/json drops it silently -- exactly
// the tolerance an evolving, unversioned transcript format needs. This
// deliberately does not call DisallowUnknownFields: that guard belongs to
// this project's own wire formats (internal/api's request bodies), not to
// a format this package only reads and never writes.
type rawLine struct {
	Type        string      `json:"type"`
	Timestamp   string      `json:"timestamp"`
	SessionID   string      `json:"sessionId"`
	IsSidechain bool        `json:"isSidechain"`
	Effort      string      `json:"effort"`
	Message     *rawMessage `json:"message"`
}

type rawMessage struct {
	Model string    `json:"model"`
	Usage *rawUsage `json:"usage"`
}

type rawUsage struct {
	InputTokens              int64                   `json:"input_tokens"`
	CacheCreationInputTokens int64                   `json:"cache_creation_input_tokens"`
	CacheCreation            *rawCacheCreation       `json:"cache_creation"`
	CacheReadInputTokens     int64                   `json:"cache_read_input_tokens"`
	OutputTokens             int64                   `json:"output_tokens"`
	OutputTokensDetails      *rawOutputTokensDetails `json:"output_tokens_details"`
	Speed                    string                  `json:"speed"`
}

type rawOutputTokensDetails struct {
	ThinkingTokens int64 `json:"thinking_tokens"`
}

// rawCacheCreation is usage.cache_creation's shape -- present only when
// the harness recorded which of the two cache-write rates applied.
// Confirmed against a live session file (this task's own verified
// comment, transcript.go's package doc) rather than assumed: the two
// keys are exactly ephemeral_5m_input_tokens and
// ephemeral_1h_input_tokens, nested one level under "cache_creation",
// alongside (not instead of) the top-level, collapsed
// cache_creation_input_tokens total.
type rawCacheCreation struct {
	Ephemeral5mInputTokens int64 `json:"ephemeral_5m_input_tokens"`
	Ephemeral1hInputTokens int64 `json:"ephemeral_1h_input_tokens"`
}

// SplitCompleteLines splits raw into the leading span that ends in a
// complete, newline-terminated line ("complete") and whatever trails the
// last newline ("tail"). Only complete is ever parsed as records.
//
// This is the same reasoning internal/reconcile's splitCompleteLines
// applies to the write-ahead journal (task 6), reused rather than
// reinvented: a transcript is being appended to by a live harness while
// this package reads it, one JSON object plus its trailing newline per
// write, so the only way a byte range after the last newline can be
// non-empty is a writer that stopped mid-write -- a live process still
// composing its next line, not a corrupt file. Treating that trailing
// span as data to parse would either fail to decode or, worse, succeed
// on a truncated-but-coincidentally-valid fragment and be attributed as
// if it were a complete message. Leaving it alone is correct either way:
// it is never counted as consumed, so the byte offset this package
// persists never advances past it, and the next read naturally starts
// from the same place and picks up the rest of that line once the writer
// finishes it.
func SplitCompleteLines(raw []byte) (complete, tail []byte) {
	idx := bytes.LastIndexByte(raw, '\n')
	if idx < 0 {
		return nil, raw
	}
	return raw[:idx+1], raw[idx+1:]
}

// ParseAssistantRecords decodes every assistant record found in complete
// (as SplitCompleteLines defines "complete") into a Record, in file
// order. Every other line -- a different "type", or a line that fails to
// decode as JSON at all -- is skipped rather than treated as fatal:
// unlike a journal entry (this package's own format, task 6), a
// transcript line is written by the harness across versions this package
// does not control, so a shape it does not recognise is exactly the
// "must be tolerated" case these instructions call out, not a corruption
// signal. A record whose timestamp fails to parse as RFC 3339 is skipped
// for the same reason: it cannot be attributed to any window without a
// timestamp, so keeping it would only ever produce a zero-value time
// that silently sorted into whichever window happened to start at the
// Unix epoch.
func ParseAssistantRecords(complete []byte) []Record {
	var out []Record
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

		var raw rawLine
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		if raw.Type != recordTypeAssistant || raw.Message == nil || raw.Message.Usage == nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			continue
		}

		u := raw.Message.Usage
		var thinking int64
		if u.OutputTokensDetails != nil {
			thinking = u.OutputTokensDetails.ThinkingTokens
		}
		var (
			cache5m, cache1h int64
			splitKnown       bool
		)
		if u.CacheCreation != nil {
			splitKnown = true
			cache5m = u.CacheCreation.Ephemeral5mInputTokens
			cache1h = u.CacheCreation.Ephemeral1hInputTokens
		}
		out = append(out, Record{
			Timestamp:   ts,
			SessionID:   raw.SessionID,
			IsSidechain: raw.IsSidechain,
			Model:       raw.Message.Model,
			Effort:      raw.Effort,
			Usage: Usage{
				InputTokens:              u.InputTokens,
				CacheCreationInputTokens: u.CacheCreationInputTokens,
				CacheCreation5mTokens:    cache5m,
				CacheCreation1hTokens:    cache1h,
				CacheSplitKnown:          splitKnown,
				CacheReadInputTokens:     u.CacheReadInputTokens,
				OutputTokens:             u.OutputTokens,
				ThinkingTokens:           thinking,
				Speed:                    u.Speed,
			},
		})
	}
	return out
}

// ReadNewRecords reads path from offset to EOF, splits off any partial
// trailing line (SplitCompleteLines), parses the assistant records found
// in the complete portion, and returns them alongside the byte offset a
// caller should persist as "consumed" -- offset + len(complete), never
// len(raw): a partial final line must never be counted as read, or a
// restart that catches it mid-write would skip the rest of that same
// line forever once the writer finishes it.
//
// A file shorter than offset (rotated or truncated out from under the
// watcher) is reported via ErrOffsetBeyondEOF rather than silently
// reread from 0, which would double-count everything already attributed
// from it under the old identity.
func ReadNewRecords(path string, offset int64) (records []Record, newOffset int64, err error) {
	f, err := openAt(path, offset)
	if err != nil {
		return nil, offset, err
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, offset, fmt.Errorf("harvest: read %s from offset %d: %w", path, offset, err)
	}

	complete, _ := SplitCompleteLines(raw)
	return ParseAssistantRecords(complete), offset + int64(len(complete)), nil
}

// openAt opens path and seeks to offset, checking first that offset does
// not exceed the file's current size (ErrOffsetBeyondEOF).
func openAt(path string, offset int64) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("harvest: open %s: %w", path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("harvest: stat %s: %w", path, err)
	}
	if offset > info.Size() {
		_ = f.Close()
		return nil, fmt.Errorf("%w: %s has %d bytes, offset is %d", ErrOffsetBeyondEOF, path, info.Size(), offset)
	}
	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("harvest: seek %s to %d: %w", path, offset, err)
		}
	}
	return f, nil
}
