package stages_test

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/stages"
)

// readmePath is README.md's location relative to this package --
// stats/internal/stages is three directories below the repository root
// (stages -> internal -> stats -> root), the same relative depth
// internal/fallback/journal_test.go's contractPath already uses for
// skills/myflow-contracts/state-file.md.
const readmePath = "../../../README.md"

// levelOneHeading and tableHeaderMarker locate, in order, the section this
// test extracts from and the first line of its table -- mirroring how a
// human would find it, rather than a line number that would silently go
// stale the next time README.md grows or shrinks above this section.
const (
	levelOneHeading   = "### Level 1 — the stages of each command"
	tableHeaderMarker = "| Command"
)

// noStagesMarker is the literal phrase README.md's own `/myflow-status`
// row uses to say it marks no stages. Extraction treats a cell containing
// this phrase as "zero stages" explicitly, rather than mechanically
// splitting it into two prose fragments that happen to look like stage
// names.
const noStagesMarker = "no stages, no state write"

// tableRowSplitRE finds the top-level separators inside a "Stages, in
// order" cell that mark "what comes next" in that command's sequence: "→"
// for the ordinary case, and ";" where a row's cell branches into more
// than one run or route (`/myflow-finish`'s two runs, `/myflow-fast`'s
// three routes).
var tableRowSplitRE = regexp.MustCompile(`→|;`)

// extractReadmeLevelOneStages reads README.md and mechanically rebuilds
// the same map names.go transcribes by hand: command name -> its ordered
// list of documented stage names, split from the "Stages, in order"
// column exactly as names.go's own doc comment states the rule. It is the
// test-time half of the tie between the code and the document -- see
// names.go's package doc comment for the other half.
//
// It fails loudly (via t.Fatalf) rather than returning an empty map if the
// section heading, the table header, or every row's command cell go
// missing: a parser that silently produced nothing would let this test
// compare two empty maps and pass despite testing nothing, which is
// exactly the failure mode this extraction exists to catch instead of
// commit.
func extractReadmeLevelOneStages(t *testing.T) map[string][]string {
	t.Helper()

	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read %s: %v", readmePath, err)
	}
	text := string(data)

	headingIdx := strings.Index(text, levelOneHeading)
	if headingIdx == -1 {
		t.Fatalf("%s no longer contains the heading %q -- the section this test extracts from has moved or been renamed", readmePath, levelOneHeading)
	}
	rest := text[headingIdx:]

	tableIdx := strings.Index(rest, tableHeaderMarker)
	if tableIdx == -1 {
		t.Fatalf("%s: no %q table found under %q", readmePath, tableHeaderMarker, levelOneHeading)
	}
	rest = rest[tableIdx:]

	out := map[string][]string{}
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "|") {
			// The table ends at the first blank line or non-table line
			// after its header -- exactly where the markdown table itself
			// ends.
			break
		}

		cells := strings.Split(line, "|")
		if len(cells) < 4 {
			continue
		}
		cmdCell := strings.TrimSpace(cells[1])
		if cmdCell == "Command" {
			continue // the header row
		}
		if strings.Trim(cmdCell, "-: ") == "" {
			continue // the "|---|---|---|" separator row
		}
		cmd := strings.Trim(cmdCell, "`")

		stagesCell := strings.TrimSpace(cells[2])
		if strings.Contains(stagesCell, noStagesMarker) {
			out[cmd] = nil
			continue
		}

		var names []string
		for _, tok := range tableRowSplitRE.Split(stagesCell, -1) {
			tok = strings.ReplaceAll(tok, " ▸", "")
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			names = append(names, tok)
		}
		out[cmd] = names
	}

	if len(out) == 0 {
		t.Fatalf("%s: extraction matched no table rows at all -- the table's shape changed in a way this parser no longer understands", readmePath)
	}
	return out
}

// TestStageNamesMatchReadmeLevelOne is the mechanical tie between
// stages.Names and README.md: it re-derives the table from the document at
// test time and fails if the two disagree, for exactly the reason
// names.go's package doc comment states -- a hand-copied list is exactly
// what this test exists to catch drifting.
func TestStageNamesMatchReadmeLevelOne(t *testing.T) {
	extracted := extractReadmeLevelOneStages(t)

	// Guard against the parser and the hand-authored map silently agreeing
	// on nothing: every pipeline command this package actually validates
	// against must have a non-empty extracted stage list, or the equality
	// check below would pass by both sides being empty rather than by
	// both sides genuinely agreeing.
	for _, cmd := range []stages.Command{stages.MyflowStart, stages.MyflowDo, stages.MyflowFinish, stages.MyflowFast} {
		if len(extracted[string(cmd)]) == 0 {
			t.Fatalf("extraction found no stages for %s -- the extractor or the README table is broken, not merely empty", cmd)
		}
	}

	// /myflow-status is documented as marking no stages, and Names has no
	// entry for it. Confirm the README row actually says so explicitly,
	// rather than this test's absence-equals-absence check passing because
	// the row went missing entirely.
	if got, ok := extracted["/myflow-status"]; !ok {
		t.Fatalf("extraction found no /myflow-status row at all")
	} else if len(got) != 0 {
		t.Errorf("/myflow-status: extracted %v, want zero stages (its row states \"no stages, no state write\")", got)
	}

	for cmd, want := range Names_stringKeyed() {
		got, ok := extracted[cmd]
		if !ok {
			t.Errorf("stages.Names has %q, but README.md's Level 1 table has no such row", cmd)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: README.md's Level 1 table disagrees with stages.Names\n  README:  %#v\n  code:    %#v", cmd, got, want)
		}
	}

	for cmd := range extracted {
		if cmd == "/myflow-status" {
			continue
		}
		if _, ok := Names_stringKeyed()[cmd]; !ok {
			t.Errorf("README.md's Level 1 table has a row for %q, but stages.Names has no entry for it", cmd)
		}
	}
}

// Names_stringKeyed re-keys stages.Names by plain string, so it can be
// compared directly against extractReadmeLevelOneStages' output without
// this test file needing its own second copy of the command list.
func Names_stringKeyed() map[string][]string {
	out := make(map[string][]string, len(stages.Names))
	for cmd, names := range stages.Names {
		out[string(cmd)] = names
	}
	return out
}

// TestUndocumentedStageNameIsRejected pins Validate's rejection of a stage
// name absent from the documented table, and that the rejection names the
// documented alternatives -- exactly what the CLI needs to report a useful
// error instead of a bare "invalid".
func TestUndocumentedStageNameIsRejected(t *testing.T) {
	err := stages.Validate(stages.MyflowDo, "a stage nobody documented")
	if err == nil {
		t.Fatal("Validate: got nil error for an undocumented stage name, want a rejection")
	}
	msg := err.Error()
	if !strings.Contains(msg, "a stage nobody documented") {
		t.Errorf("error %q does not name the rejected stage", msg)
	}
	for _, want := range stages.Names[stages.MyflowDo] {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not name documented alternative %q", msg, want)
		}
	}

	var unknown *stages.ErrUnknownStage
	if !asErrUnknownStage(err, &unknown) {
		t.Fatalf("Validate's error is not an *ErrUnknownStage: %v (%T)", err, err)
	}
	if unknown.Command != stages.MyflowDo || unknown.Stage != "a stage nobody documented" {
		t.Errorf("ErrUnknownStage = %+v, want Command=%s Stage=%q", unknown, stages.MyflowDo, "a stage nobody documented")
	}
}

func asErrUnknownStage(err error, target **stages.ErrUnknownStage) bool {
	e, ok := err.(*stages.ErrUnknownStage)
	if !ok {
		return false
	}
	*target = e
	return true
}

// TestValidStageNameIsAccepted breaks the opposite direction: a stage name
// taken verbatim from Names must be accepted for its own command, or
// Validate would reject every legitimate mark and the CLI could never mark
// a stage at all.
func TestValidStageNameIsAccepted(t *testing.T) {
	for cmd, names := range stages.Names {
		for _, name := range names {
			if err := stages.Validate(cmd, name); err != nil {
				t.Errorf("Validate(%s, %q) = %v, want nil", cmd, name, err)
			}
		}
	}
}

// TestStageNameIsRejectedForTheWrongCommand pins that a stage name
// documented for one command is not silently valid for another -- Valid
// must key on (command, stage) together, not merely check stage against
// the union of every command's stages.
func TestStageNameIsRejectedForTheWrongCommand(t *testing.T) {
	// "brainstorm" is documented for /myflow-start, never for /myflow-do.
	if stages.Valid(stages.MyflowDo, "brainstorm") {
		t.Error(`Valid(MyflowDo, "brainstorm") = true, want false -- "brainstorm" is only documented for /myflow-start`)
	}
	if !stages.Valid(stages.MyflowStart, "brainstorm") {
		t.Error(`Valid(MyflowStart, "brainstorm") = false, want true`)
	}
}

// TestMyflowStatusHasNoDocumentedStages pins that /myflow-status -- which
// README.md documents as read-only, marking nothing -- has no entry in
// Names, so a mark naming it as a command is always rejected.
func TestMyflowStatusHasNoDocumentedStages(t *testing.T) {
	if _, ok := stages.Names["/myflow-status"]; ok {
		t.Error(`stages.Names has an entry for "/myflow-status", want none -- it is read-only and marks no stages`)
	}
	if err := stages.Validate("/myflow-status", "resolve the change"); err == nil {
		t.Error("Validate(/myflow-status, ...) = nil, want a rejection naming that the command has no documented stages")
	}
}
