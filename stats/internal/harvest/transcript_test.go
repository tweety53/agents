package harvest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

const (
	mainThreadFixture = "../../testdata/transcripts/main-thread.jsonl"
	sidechainFixture  = "../../testdata/transcripts/sidechain.jsonl"
	truncatedFixture  = "../../testdata/transcripts/truncated.jsonl"
)

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return data
}

// TestParseUsageFromLiveFixture parses main-thread.jsonl and checks the
// resulting records against the exact values the fixture was built with,
// so this test would fail if the parser silently mis-mapped a JSON field
// to the wrong Go field -- exactly the class of bug a fixture built to
// already match whatever shape the parser expects could never catch.
//
// The fixture's own content is synthetic -- an earlier version of this
// fixture copied a real transcript excerpt verbatim, which turned out to
// carry real cwd paths, a real git branch, and a subagent prompt quoting
// an unrelated project's task brief (task 9's post-commit review, finding
// F5). What matters, and what this fixture still preserves exactly as
// discovered against a live ~/.claude/projects/ tree, is its *structure*:
// every record type actually found there (assistant, user, attachment,
// system, file-history-snapshot, file-history-delta, ai-title, mode,
// permission-mode, last-prompt), the exact usage shape including
// output_tokens_details.thinking_tokens, and the fact that a subagent's
// messages live in a separate subagents/agent-*.jsonl file while sharing
// the parent session's own sessionId (see
// TestDiscoverTranscriptsFindsNestedSubagentFiles, watcher_test.go).
// This test's own name keeps its
// original wording ("LiveFixture") because it is one of the eight test
// names tasks.md commits to by name; only the fixture's content changed.
func TestParseUsageFromLiveFixture(t *testing.T) {
	complete, tail := harvest.SplitCompleteLines(readFixture(t, mainThreadFixture))
	if len(tail) != 0 {
		t.Fatalf("main-thread fixture should have no partial trailing line, got %d trailing bytes", len(tail))
	}

	records := harvest.ParseAssistantRecords(complete)
	if len(records) != 7 {
		t.Fatalf("got %d assistant records, want 7", len(records))
	}

	first := records[0]
	wantTS, err := time.Parse(time.RFC3339Nano, "2026-01-01T00:00:00.500Z")
	if err != nil {
		t.Fatalf("parse want timestamp: %v", err)
	}
	if !first.Timestamp.Equal(wantTS) {
		t.Errorf("first record timestamp = %v, want %v", first.Timestamp, wantTS)
	}
	if first.SessionID != mainSessionID {
		t.Errorf("first record sessionID = %q, want %q", first.SessionID, mainSessionID)
	}
	if first.IsSidechain {
		t.Errorf("first record IsSidechain = true, want false (main thread)")
	}
	if first.Model != "claude-opus-5" {
		t.Errorf("first record model = %q, want claude-opus-5", first.Model)
	}
	if first.Effort != "medium" {
		t.Errorf("first record effort = %q, want medium", first.Effort)
	}
	wantUsage := harvest.Usage{
		InputTokens:              2,
		CacheCreationInputTokens: 10597,
		// The fixture's cache_creation object is present on every
		// assistant record (built to match a real transcript's own shape
		// -- this task's own verified plan-provenance note), and every
		// main-thread record here is entirely a 1-hour write.
		CacheCreation1hTokens: 10597,
		CacheSplitKnown:       true,
		CacheReadInputTokens:  12673,
		OutputTokens:          302,
		ThinkingTokens:        0,
	}
	if first.Usage != wantUsage {
		t.Errorf("first record usage = %+v, want %+v", first.Usage, wantUsage)
	}

	// The seventh record is the one fixture line carrying
	// output_tokens_details.thinking_tokens -- verifying it decodes
	// confirms the nested field is actually read, not merely present in
	// the JSON and silently ignored.
	last := records[6]
	if last.Usage.ThinkingTokens != 37 {
		t.Errorf("last record thinking tokens = %d, want 37", last.Usage.ThinkingTokens)
	}
}

// TestNonAssistantAndUnknownTypesAreSkipped asserts the tolerance this
// task's instructions require directly: every non-assistant type in the
// fixture (last-prompt, mode, permission-mode, attachment, system,
// file-history-snapshot, user, ai-title, file-history-delta) is present
// in the file yet contributes no Record and causes no error.
func TestNonAssistantAndUnknownTypesAreSkipped(t *testing.T) {
	complete, _ := harvest.SplitCompleteLines(readFixture(t, mainThreadFixture))
	records := harvest.ParseAssistantRecords(complete)
	if len(records) != 7 {
		t.Fatalf("got %d records from a fixture with 7 assistant lines among many other types, want 7", len(records))
	}

	// An entirely-unrecognised type (never documented anywhere in this
	// package) must not error either -- "do not assume the set is
	// closed".
	unknown := []byte(`{"type":"some-future-type-nobody-has-invented-yet","timestamp":"2026-01-01T00:00:00Z"}` + "\n")
	if got := harvest.ParseAssistantRecords(unknown); len(got) != 0 {
		t.Errorf("unknown type produced %d records, want 0", len(got))
	}
}

// TestTruncatedFinalLineIsResumedNotFailed exercises the exact scenario
// this task calls out: a transcript file that is being appended to while
// it is read, so its last line can be incomplete. truncated.jsonl is
// main-thread.jsonl with its final record's bytes cut in half and no
// trailing newline -- a genuine partial write, not a corrupt file.
func TestTruncatedFinalLineIsResumedNotFailed(t *testing.T) {
	raw := readFixture(t, truncatedFixture)
	complete, tail := harvest.SplitCompleteLines(raw)
	if len(tail) == 0 {
		t.Fatalf("truncated fixture should have a non-empty partial trailing line")
	}

	records := harvest.ParseAssistantRecords(complete)
	if len(records) != 6 {
		t.Fatalf("got %d records from the truncated fixture, want 6 (the 7th record's bytes are incomplete)", len(records))
	}

	// Simulate the writer finishing its write: a fresh copy of the file
	// starts truncated, then gains the rest of main-thread.jsonl's bytes.
	// A resume from the previously returned offset must pick up exactly
	// the one record that was missing, never re-emit the six already
	// consumed.
	dir := t.TempDir()
	path := filepath.Join(dir, "growing.jsonl")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write growing fixture: %v", err)
	}

	firstPass, offsetAfterFirst, err := harvest.ReadNewRecords(path, 0)
	if err != nil {
		t.Fatalf("ReadNewRecords (truncated): %v", err)
	}
	if len(firstPass) != 6 {
		t.Fatalf("first pass got %d records, want 6", len(firstPass))
	}
	if int(offsetAfterFirst) != len(complete) {
		t.Fatalf("offset after first pass = %d, want %d (exactly the complete portion)", offsetAfterFirst, len(complete))
	}

	full := readFixture(t, mainThreadFixture)
	if err := os.WriteFile(path, full, 0o644); err != nil {
		t.Fatalf("complete the write: %v", err)
	}

	secondPass, offsetAfterSecond, err := harvest.ReadNewRecords(path, offsetAfterFirst)
	if err != nil {
		t.Fatalf("ReadNewRecords (resumed): %v", err)
	}
	if len(secondPass) != 1 {
		t.Fatalf("second pass got %d records, want exactly 1 (the record that was missing)", len(secondPass))
	}
	if secondPass[0].Usage.ThinkingTokens != 37 {
		t.Errorf("resumed record thinking tokens = %d, want 37", secondPass[0].Usage.ThinkingTokens)
	}
	if int(offsetAfterSecond) != len(full) {
		t.Errorf("offset after second pass = %d, want %d (end of file)", offsetAfterSecond, len(full))
	}
}

// TestCacheCreationSplitAbsentIsRecordedAsUnknown covers the case task 23
// exists to guard against directly: a transcript line carrying the
// collapsed cache_creation_input_tokens total but no "cache_creation"
// split object at all -- an older or differently-shaped line, not zero
// cache-creation usage. The parser must not guess a split (both zero
// would silently read as "no 1-hour write happened", which is a real,
// distinct fact this line does not actually report); it must say the
// split is unknown, per Usage.CacheSplitKnown's own doc comment.
func TestCacheCreationSplitAbsentIsRecordedAsUnknown(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s1","message":{"model":"claude-opus-5","usage":{"input_tokens":2,"cache_creation_input_tokens":900,"cache_read_input_tokens":10,"output_tokens":5}}}` + "\n")
	records := harvest.ParseAssistantRecords(line)
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	u := records[0].Usage
	if u.CacheSplitKnown {
		t.Errorf("CacheSplitKnown = true, want false: this line carries no cache_creation split object")
	}
	if u.CacheCreationInputTokens != 900 {
		t.Errorf("CacheCreationInputTokens = %d, want 900 (the collapsed total is still recorded)", u.CacheCreationInputTokens)
	}
	if u.CacheCreation5mTokens != 0 || u.CacheCreation1hTokens != 0 {
		t.Errorf("split tokens = (5m=%d, 1h=%d), want both 0 when no split was ever recorded", u.CacheCreation5mTokens, u.CacheCreation1hTokens)
	}
}

// TestCacheCreationSplitAndFastSpeedAreParsed is the positive companion:
// a line carrying both the cache_creation split and a "fast" speed
// (task 23's third defect) must decode both, alongside the same
// collapsed total.
func TestCacheCreationSplitAndFastSpeedAreParsed(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s1","message":{"model":"claude-opus-5","usage":{"input_tokens":2,"cache_creation_input_tokens":900,"cache_creation":{"ephemeral_5m_input_tokens":300,"ephemeral_1h_input_tokens":600},"cache_read_input_tokens":10,"output_tokens":5,"speed":"fast"}}}` + "\n")
	records := harvest.ParseAssistantRecords(line)
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	u := records[0].Usage
	if !u.CacheSplitKnown {
		t.Fatalf("CacheSplitKnown = false, want true")
	}
	if u.CacheCreation5mTokens != 300 || u.CacheCreation1hTokens != 600 {
		t.Errorf("split tokens = (5m=%d, 1h=%d), want (300, 600)", u.CacheCreation5mTokens, u.CacheCreation1hTokens)
	}
	if u.CacheCreationInputTokens != 900 {
		t.Errorf("CacheCreationInputTokens = %d, want 900 (unchanged collapsed total)", u.CacheCreationInputTokens)
	}
	if u.Speed != "fast" {
		t.Errorf("Speed = %q, want %q", u.Speed, "fast")
	}
}

// TestReadNewRecordsOffsetBeyondEOFIsReported guards against silently
// re-reading a rotated or truncated file from byte 0, which would
// re-attribute usage this package already reported as consumed.
func TestReadNewRecordsOffsetBeyondEOFIsReported(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "short.jsonl")
	if err := os.WriteFile(path, []byte(`{"type":"user"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, _, err := harvest.ReadNewRecords(path, 10_000)
	if !errors.Is(err, harvest.ErrOffsetBeyondEOF) {
		t.Fatalf("ReadNewRecords error = %v, want ErrOffsetBeyondEOF", err)
	}
}
