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
	multiBlockFixture = "../../testdata/transcripts/multi-block.jsonl"
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

	firstPass, err := harvest.ReadNewRecords(path, 0, nil)
	if err != nil {
		t.Fatalf("ReadNewRecords (truncated): %v", err)
	}
	if len(firstPass.Records) != 6 {
		t.Fatalf("first pass got %d records, want 6", len(firstPass.Records))
	}
	if int(firstPass.NewOffset) != len(complete) {
		t.Fatalf("offset after first pass = %d, want %d (exactly the complete portion)", firstPass.NewOffset, len(complete))
	}

	full := readFixture(t, mainThreadFixture)
	if err := os.WriteFile(path, full, 0o644); err != nil {
		t.Fatalf("complete the write: %v", err)
	}

	secondPass, err := harvest.ReadNewRecords(path, firstPass.NewOffset, nil)
	if err != nil {
		t.Fatalf("ReadNewRecords (resumed): %v", err)
	}
	if len(secondPass.Records) != 1 {
		t.Fatalf("second pass got %d records, want exactly 1 (the record that was missing)", len(secondPass.Records))
	}
	if secondPass.Records[0].Usage.ThinkingTokens != 37 {
		t.Errorf("resumed record thinking tokens = %d, want 37", secondPass.Records[0].Usage.ThinkingTokens)
	}
	if int(secondPass.NewOffset) != len(full) {
		t.Errorf("offset after second pass = %d, want %d (end of file)", secondPass.NewOffset, len(full))
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

	_, err := harvest.ReadNewRecords(path, 10_000, nil)
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
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1},"content":[{"type":"text","text":"running the mark"},{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -stage do.tests -session-token mf-abc123 -harness claude-code","description":"begin stage"}}]}}` + "\n")

	got := harvest.ParseCommandRecords(line)
	if len(got) != 1 {
		t.Fatalf("got %d command records, want 1: %+v", len(got), got)
	}
	if got[0].SessionID != "session-x" {
		t.Errorf("SessionID = %q, want %q", got[0].SessionID, "session-x")
	}
	want := "flow stage begin -stage do.tests -session-token mf-abc123 -harness claude-code"
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
	line := []byte(`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-xyz"}}]}}` + "\n")

	got := harvest.ParseCommandRecords(line)
	if len(got) != 1 || got[0].Command != "flow stage begin -session-token mf-xyz" {
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

// TestParseAssistantRecordsDeduplicatesByMessageID pins that a
// multi-block API response -- one JSONL line per content block, every
// line carrying the same message.id and identical usage -- is counted
// once, and that a line carrying no id at all is never collapsed into a
// neighbour (design.md's harvest-dedupe-by-message-id).
func TestParseAssistantRecordsDeduplicatesByMessageID(t *testing.T) {
	complete, _ := harvest.SplitCompleteLines(readFixture(t, multiBlockFixture))
	records := harvest.ParseAssistantRecords(complete)
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (three lines share msg_dup, one has no id, one is msg_two)", len(records))
	}
	if got := records[0].Usage.InputTokens; got != 5 {
		t.Errorf("first record input = %d, want 5 (msg_dup, counted once)", got)
	}
	if got := records[1].Usage.InputTokens; got != 11 {
		t.Errorf("second record input = %d, want 11 (the id-less line is kept)", got)
	}
	if got := records[2].Usage.InputTokens; got != 13 {
		t.Errorf("third record input = %d, want 13 (msg_two)", got)
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
	content := `{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","message":{"model":"claude-opus-5","usage":{"input_tokens":1,"output_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-abc123"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	batch, err := harvest.ReadNewRecords(path, 0, nil)
	if err != nil {
		t.Fatalf("ReadNewRecords: %v", err)
	}
	if len(batch.Records) != 1 {
		t.Fatalf("got %d token records, want 1", len(batch.Records))
	}
	if len(batch.Commands) != 1 || batch.Commands[0].Command != "flow stage begin -session-token mf-abc123" {
		t.Fatalf("got %+v, want one command record for mf-abc123", batch.Commands)
	}
	if len(batch.Launches) != 0 {
		t.Errorf("got %d launches, want 0 (no launch in this fixture)", len(batch.Launches))
	}
	if int(batch.NewOffset) != len(content) {
		t.Errorf("newOffset = %d, want %d", batch.NewOffset, len(content))
	}
}

// TestParseAgentLaunchesExtractsAsyncLaunches is KAN-322's positive case:
// an async agent launch's tool result is a user-type line whose
// top-level "toolUseResult" object carries status "async_launched" and a
// non-empty "agentId" (confirmed against a live parent transcript before
// writing this fixture, not assumed). ParseAgentLaunches must recover the
// agent id, the session it was recorded under, and the line's index
// within complete, so the watcher can pair it with the dispatch begin
// command that precedes it in file order.
func TestParseAgentLaunchesExtractsAsyncLaunches(t *testing.T) {
	complete := []byte(
		`{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"session-x","toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a68cee7239419a7e7"}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:01Z","sessionId":"session-x","message":{"content":[{"type":"tool_result","tool_use_id":"u1","content":"plain result"}]}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:02Z","sessionId":"session-x","toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a8884ead3c626d980"}}` + "\n")

	got := harvest.ParseAgentLaunches(complete)
	if len(got) != 2 {
		t.Fatalf("got %d launches, want 2: %+v", len(got), got)
	}
	if got[0].AgentID != "a68cee7239419a7e7" || got[0].SessionID != "session-x" {
		t.Errorf("first launch = %+v, want agent a68cee7239419a7e7 in session-x", got[0])
	}
	if got[0].Line != 0 {
		t.Errorf("first launch Line = %d, want 0", got[0].Line)
	}
	if got[1].AgentID != "a8884ead3c626d980" || got[1].Line != 2 {
		t.Errorf("second launch = %+v, want agent a8884ead3c626d980 at line 2", got[1])
	}
}

// TestParseAgentLaunchesIgnoresNonLaunchResults is the negative
// companion: a tool result with no agentId, one with an empty agentId, a
// tool result whose status is not async_launched (a resumed agent's
// result, whose id belongs to its original dispatch and must never be
// re-stamped onto whatever begin happens to precede it), an assistant
// line, and a line that is not JSON at all must all contribute nothing.
func TestParseAgentLaunchesIgnoresNonLaunchResults(t *testing.T) {
	complete := []byte(
		`{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"s","message":{"content":[{"type":"tool_result","tool_use_id":"u1","is_error":true,"content":"PreToolUse hook denied"}]}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:01Z","sessionId":"s","toolUseResult":{"status":"async_launched"}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:02Z","sessionId":"s","toolUseResult":{"status":"completed","agentId":"a68cee7239419a7e7"}}` + "\n" +
			`{"type":"assistant","timestamp":"2026-01-01T00:00:03Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[]}}` + "\n" +
			`not json at all` + "\n")

	got := harvest.ParseAgentLaunches(complete)
	if len(got) != 0 {
		t.Fatalf("got %d launches, want 0: %+v", len(got), got)
	}
}

// TestParseAgentLaunchesOnePerLine pins the one launch per line
// rule: each line carries at most one toolUseResult, so two lines that
// each carry one yield two launches, one per line -- and a single line
// never yields more than one.
func TestParseAgentLaunchesOnePerLine(t *testing.T) {
	line := `{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"s","toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a68cee7239419a7e7"}}` + "\n"
	got := harvest.ParseAgentLaunches([]byte(line + line))
	if len(got) != 2 {
		t.Fatalf("got %d launches, want 2 (two distinct lines, each its own launch): %+v", len(got), got)
	}
	if got[0].Line != 0 || got[1].Line != 1 {
		t.Errorf("launch lines = %d, %d, want 0, 1", got[0].Line, got[1].Line)
	}
}

// TestParseCommandRecordsCarriesLineNumbers pins CommandRecord.Line:
// the watcher pairs a launch with the most recent dispatch-begin command
// at a smaller line, which needs each command's index within complete,
// not just its text.
func TestParseCommandRecordsCarriesLineNumbers(t *testing.T) {
	complete := []byte(
		`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"text","text":"thinking"}]}}` + "\n" +
			`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"flow record dispatch begin -key task-1-implementer -session-token mf-abc123 -started-at 2026-01-01T00:00:00Z"}}]}}` + "\n" +
			`{"type":"assistant","timestamp":"2026-01-01T00:00:02Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-xyz"}}]}}` + "\n")

	got := harvest.ParseCommandRecords(complete)
	if len(got) != 2 {
		t.Fatalf("got %d commands, want 2: %+v", len(got), got)
	}
	if got[0].Line != 1 {
		t.Errorf("first command Line = %d, want 1", got[0].Line)
	}
	if got[1].Line != 2 {
		t.Errorf("second command Line = %d, want 2", got[1].Line)
	}
}

// TestParseDispatchEventsJoinsDeniedAgentCall is KAN-322's denial fix:
// an errored tool_result whose tool_use_id names an Agent/Task tool call
// is a dispatch attempt that never launched, and the watcher retires the
// begin it left pending. A denied Bash call says nothing about any
// dispatch and yields nothing; a successful agent result resolves its
// call without a denial.
func TestParseDispatchEventsJoinsDeniedAgentCall(t *testing.T) {
	complete := []byte(
		`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","id":"uA","name":"Agent","input":{}}]}}` + "\n" +
			`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","id":"uB","name":"Bash","input":{"command":"ls"}}]}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:02Z","sessionId":"s","message":{"content":[{"type":"tool_result","tool_use_id":"uB","is_error":true,"content":"PreToolUse hook denied"}]}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:03Z","sessionId":"s","message":{"content":[{"type":"tool_result","tool_use_id":"uA","is_error":true,"content":"PreToolUse hook denied the Agent call"}]}}` + "\n")

	denials, stillOpen := harvest.ParseDispatchEvents(complete, nil)
	if len(denials) != 1 {
		t.Fatalf("got %d denials, want 1: %v", len(denials), denials)
	}
	wantTS, _ := time.Parse(time.RFC3339Nano, "2026-01-01T00:00:03Z")
	if !denials[0].Equal(wantTS) {
		t.Errorf("denial timestamp = %v, want %v", denials[0], wantTS)
	}
	if len(stillOpen) != 0 {
		t.Errorf("stillOpen = %v, want empty (every call resolved)", stillOpen)
	}
}

// TestParseDispatchEventsCarriesOpenCallsAcrossBatches pins the
// straddled-call join: a batch whose Agent tool_use carries no result
// reports the id in stillOpen, and the NEXT read -- handed that set back
// -- joins a denial landing there to the same call. Without the carry,
// a dispatch denied across a batch boundary would leave its begin
// pending and free to steal a later launch.
func TestParseDispatchEventsCarriesOpenCallsAcrossBatches(t *testing.T) {
	batch1 := []byte(
		`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","id":"uA","name":"Agent","input":{}}]}}` + "\n")

	denials, stillOpen := harvest.ParseDispatchEvents(batch1, nil)
	if len(denials) != 0 {
		t.Fatalf("got %d denials before any result, want 0", len(denials))
	}
	if len(stillOpen) != 1 || stillOpen[0] != "uA" {
		t.Fatalf("stillOpen = %v, want [uA]", stillOpen)
	}

	carried := map[string]bool{"uA": true}
	batch2 := []byte(
		`{"type":"user","timestamp":"2026-01-01T00:00:05Z","sessionId":"s","message":{"content":[{"type":"tool_result","tool_use_id":"uA","is_error":true,"content":"Permission for this action was denied"}]}}` + "\n")
	denials, stillOpen = harvest.ParseDispatchEvents(batch2, carried)
	if len(denials) != 1 {
		t.Fatalf("got %d denials across the boundary, want 1", len(denials))
	}
	if len(stillOpen) != 0 {
		t.Errorf("stillOpen = %v, want empty (the carried call resolved)", stillOpen)
	}
}
