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

	firstPass, _, offsetAfterFirst, err := harvest.ReadNewRecords(path, 0)
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

	secondPass, _, offsetAfterSecond, err := harvest.ReadNewRecords(path, offsetAfterFirst)
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

	_, _, _, err := harvest.ReadNewRecords(path, 10_000)
	if !errors.Is(err, harvest.ErrOffsetBeyondEOF) {
		t.Fatalf("ReadNewRecords error = %v, want ErrOffsetBeyondEOF", err)
	}
}

// TestParseCommandRecordsFindsBashToolUse is KAN-172 task 2's parsing
// guard: a `stage begin ... -session-token mf-abc123 ...` command a skill
// runs through the Bash tool lands in an assistant line's
// message.content[].input.command, exactly as a real transcript records
// it (confirmed by reading a live ~/.claude/projects/*.jsonl file before
// writing this fixture, not assumed). ParseCommandRecords must recover
// that literal string, alongside the session it was recorded under.
func TestParseCommandRecordsFindsBashToolUse(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1},"content":[{"type":"text","text":"running the mark"},{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -stage do.tests -session-token mf-abc123 -harness claude-code","description":"begin stage"}}]}}` + "\n")

	got := harvest.ParseCommandRecords(line)
	if len(got) != 1 {
		t.Fatalf("got %d command records, want 1: %+v", len(got), got)
	}
	if got[0].SessionID != "session-x" {
		t.Errorf("SessionID = %q, want %q", got[0].SessionID, "session-x")
	}
	want := "myflow stage begin -stage do.tests -session-token mf-abc123 -harness claude-code"
	if got[0].Command != want {
		t.Errorf("Command = %q, want %q", got[0].Command, want)
	}
}

// TestParseCommandRecordsSkipsNonBashToolsAndTextBlocks is the negative
// companion: a text block, a non-Bash tool_use (for example Read), and a
// Bash-named block whose input carries no "command" key must all
// contribute nothing -- silently, the same tolerance
// ParseAssistantRecords already extends to shapes this package does not
// recognise.
func TestParseCommandRecordsSkipsNonBashToolsAndTextBlocks(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1},"content":[{"type":"text","text":"hello"},{"type":"tool_use","name":"Read","input":{"file_path":"/tmp/foo"}},{"type":"tool_use","name":"Bash","input":{"description":"no command key here"}}]}}` + "\n")

	got := harvest.ParseCommandRecords(line)
	if len(got) != 0 {
		t.Fatalf("got %d command records, want 0: %+v", len(got), got)
	}
}

// TestParseCommandRecordsIgnoresMissingUsage is the reason
// ParseCommandRecords does not reuse ParseAssistantRecords' per-line
// gate: an assistant line carrying a Bash tool_use but no "usage" object
// at all (a shape this package has not observed live, but not one this
// parser should depend on to find a session token) must still yield its
// command.
func TestParseCommandRecordsIgnoresMissingUsage(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-xyz"}}]}}` + "\n")

	got := harvest.ParseCommandRecords(line)
	if len(got) != 1 || got[0].Command != "myflow stage begin -session-token mf-xyz" {
		t.Fatalf("got %+v, want one command record for mf-xyz", got)
	}
}

// TestParseAssistantRecordsCarriesAgentID is KAN-201 task 3's positive
// case: a subagent transcript line carries the dispatch's own "agentId"
// field (confirmed against a real agent-<id>.jsonl file during planning,
// tasks.md's "Facts this plan rests on"), and it must decode straight
// into Record.AgentID -- the same one-line addition IsSidechain already
// has.
func TestParseAssistantRecordsCarriesAgentID(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s1","agentId":"abc123","isSidechain":true,"message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1}}}` + "\n")
	records := harvest.ParseAssistantRecords(line)
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if records[0].AgentID != "abc123" {
		t.Errorf("AgentID = %q, want %q", records[0].AgentID, "abc123")
	}
}

// TestParseAssistantRecordsAgentIDAbsent is the negative companion: a
// parent-session line carries no "agentId" at all, and must decode to
// AgentID == "" -- never a placeholder, per this change's own
// absence-is-never-a-value rule.
func TestParseAssistantRecordsAgentIDAbsent(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s1","message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1}}}` + "\n")
	records := harvest.ParseAssistantRecords(line)
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if records[0].AgentID != "" {
		t.Errorf("AgentID = %q, want empty (no agentId field on this line)", records[0].AgentID)
	}
}

// TestReadDispatchMeta covers the positive case for the meta sidecar:
// a subagent transcript at .../subagents/agent-x.jsonl with a
// well-formed sibling agent-x.meta.json returns all four descriptors and
// true. The fixture's keys match a real meta file's shape exactly
// (tasks.md's verified plan-provenance note): agentType, description,
// toolUseId, spawnDepth, model -- toolUseId is read by nobody here and
// must be silently dropped, the same tolerance rawLine's own doc comment
// commits to for the transcript format.
func TestReadDispatchMeta(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subagents")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir subagents: %v", err)
	}
	transcriptPath := filepath.Join(sub, "agent-x.jsonl")
	if err := os.WriteFile(transcriptPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	metaPath := filepath.Join(sub, "agent-x.meta.json")
	metaJSON := `{"agentType":"general-purpose","description":"Implement Task 3 checkpoint mode","toolUseId":"toolu_01DuDwNWDG136mB5nqhkgQQU","spawnDepth":1,"model":"haiku"}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	got, ok := harvest.ReadDispatchMeta(transcriptPath)
	if !ok {
		t.Fatalf("ReadDispatchMeta ok = false, want true")
	}
	want := harvest.DispatchMeta{
		AgentType:   "general-purpose",
		Description: "Implement Task 3 checkpoint mode",
		Model:       "haiku",
		SpawnDepth:  1,
	}
	if got != want {
		t.Errorf("ReadDispatchMeta = %+v, want %+v", got, want)
	}
}

// TestReadDispatchMetaAbsent covers both ways a dispatch's tokens must
// still be attributable with no descriptors: a subagent transcript with
// no sidecar at all, and a transcript whose parent directory is not
// named "subagents" even though a validly-named sidecar sits right next
// to it -- proving the parent-directory guard governs, not merely
// whichever file happens to exist on disk.
func TestReadDispatchMetaAbsent(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subagents")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir subagents: %v", err)
	}
	noSidecar := filepath.Join(sub, "agent-y.jsonl")
	if err := os.WriteFile(noSidecar, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	if _, ok := harvest.ReadDispatchMeta(noSidecar); ok {
		t.Errorf("ReadDispatchMeta ok = true with no sidecar present, want false")
	}

	main := filepath.Join(dir, "main")
	if err := os.MkdirAll(main, 0o755); err != nil {
		t.Fatalf("mkdir main: %v", err)
	}
	notSubagent := filepath.Join(main, "agent-z.jsonl")
	if err := os.WriteFile(notSubagent, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	sidecarNextToIt := filepath.Join(main, "agent-z.meta.json")
	if err := os.WriteFile(sidecarNextToIt, []byte(`{"agentType":"general-purpose","description":"d","spawnDepth":0,"model":"m"}`), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	if _, ok := harvest.ReadDispatchMeta(notSubagent); ok {
		t.Errorf("ReadDispatchMeta ok = true for a transcript whose parent is not subagents/, want false regardless of a present sidecar")
	}
}

// TestReadDispatchMetaMalformed covers a sidecar that exists but does not
// decode as JSON -- also false, never an error, per the same rule: a
// dispatch's tokens must still be attributable without its descriptors.
func TestReadDispatchMetaMalformed(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subagents")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir subagents: %v", err)
	}
	transcriptPath := filepath.Join(sub, "agent-w.jsonl")
	if err := os.WriteFile(transcriptPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	metaPath := filepath.Join(sub, "agent-w.meta.json")
	if err := os.WriteFile(metaPath, []byte("{not valid json"), 0o644); err != nil {
		t.Fatalf("write malformed meta: %v", err)
	}

	if _, ok := harvest.ReadDispatchMeta(transcriptPath); ok {
		t.Errorf("ReadDispatchMeta ok = true for malformed sidecar JSON, want false")
	}
}

// TestReadNewRecordsAlsoReturnsCommands is the wiring guard for
// ReadNewRecords' extended signature: the same read that returns token
// records must also return the Bash commands found in the identical
// byte range, so Watcher.RunOnce's session-token resolution (KAN-172,
// task 2) never needs a second read of the file.
func TestReadNewRecordsAlsoReturnsCommands(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	content := `{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-abc123"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	records, commands, newOffset, err := harvest.ReadNewRecords(path, 0)
	if err != nil {
		t.Fatalf("ReadNewRecords: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d token records, want 1", len(records))
	}
	if len(commands) != 1 || commands[0].Command != "myflow stage begin -session-token mf-abc123" {
		t.Fatalf("got %+v, want one command record for mf-abc123", commands)
	}
	if int(newOffset) != len(content) {
		t.Errorf("newOffset = %d, want %d", newOffset, len(content))
	}
}
