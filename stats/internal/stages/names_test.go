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
// skills/flow-contracts/state-file.md.
const readmePath = "../../../README.md"

// levelOneHeading locates the section this test extracts from.
// stagesTableHeaderMarker locates the first line of the *stages* table
// specifically -- the section also holds a second table (the per-command
// gate), whose header ("| Command") must not be mistaken for the stages
// table's own.
const (
	levelOneHeading         = "### Level 1 — the stages of each command"
	stagesTableHeaderMarker = "| Key | Name | Commands |"
)

// extractedStage mirrors stages.Stage but keeps Commands as plain strings,
// so this file needs no second copy of stages.Command's value set to parse
// a table cell into it.
type extractedStage struct {
	Key      string
	Name     string
	Commands []string
}

// extractReadmeLevelOneStages reads README.md and mechanically rebuilds the
// same table names.go transcribes by hand: one row per documented stage,
// key/name/commands, in the order README.md lists them. It is the
// test-time half of the tie between the code and the document -- see
// names.go's package doc comment for the other half.
//
// It fails loudly (via t.Fatalf) if the heading, the table header, or any
// row's command cell go missing, and if it ever finds the same key twice --
// a parser that silently produced nothing, or that silently let a
// duplicate key through, would defeat exactly the property this extraction
// exists to enforce (tasks.md, "Uniqueness is asserted, not assumed").
func extractReadmeLevelOneStages(t *testing.T) []extractedStage {
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

	tableIdx := strings.Index(rest, stagesTableHeaderMarker)
	if tableIdx == -1 {
		t.Fatalf("%s: no %q table found under %q", readmePath, stagesTableHeaderMarker, levelOneHeading)
	}
	rest = rest[tableIdx:]

	seen := map[string]bool{}
	var out []extractedStage
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "|") {
			// The stages table ends at the first blank line or non-table
			// line after its header -- exactly where the markdown table
			// itself ends, and before the second (gate) table begins.
			break
		}

		cells := strings.Split(line, "|")
		if len(cells) < 5 {
			continue
		}
		keyCell := strings.TrimSpace(cells[1])
		if keyCell == "Key" {
			continue // the header row
		}
		if strings.Trim(keyCell, "-: ") == "" {
			continue // the "|---|---|---|" separator row
		}
		key := strings.Trim(keyCell, "`")

		name := strings.TrimSpace(cells[2])

		var commands []string
		for _, tok := range strings.Split(cells[3], ",") {
			tok = strings.Trim(strings.TrimSpace(tok), "`")
			if tok == "" {
				continue
			}
			commands = append(commands, tok)
		}

		if seen[key] {
			t.Fatalf("%s: key %q appears more than once in the stages table -- a duplicate key would silently merge two stages in every statistic", readmePath, key)
		}
		seen[key] = true

		out = append(out, extractedStage{Key: key, Name: name, Commands: commands})
	}

	if len(out) == 0 {
		t.Fatalf("%s: extraction matched no table rows at all -- the table's shape changed in a way this parser no longer understands", readmePath)
	}
	return out
}

// stageTableAsExtracted re-keys stages.Table into the plain-string shape
// extractReadmeLevelOneStages produces, so the two can be compared directly
// without this file needing its own second copy of stages.Command's values.
func stageTableAsExtracted() []extractedStage {
	out := make([]extractedStage, len(stages.Table))
	for i, s := range stages.Table {
		cmds := make([]string, len(s.Commands))
		for j, c := range s.Commands {
			cmds[j] = string(c)
		}
		out[i] = extractedStage{Key: s.Key, Name: s.Name, Commands: cmds}
	}
	return out
}

// TestStagesMatchReadmeLevelOne is the mechanical tie between stages.Table
// and README.md: it re-derives the table from the document at test time
// and fails if the two disagree -- in content or in order -- for exactly
// the reason names.go's package doc comment states: a hand-copied list is
// exactly what this test exists to catch drifting.
func TestStagesMatchReadmeLevelOne(t *testing.T) {
	extracted := extractReadmeLevelOneStages(t)
	want := stageTableAsExtracted()

	if !reflect.DeepEqual(extracted, want) {
		t.Errorf("README.md's Level 1 stages table disagrees with stages.Table\n  README:  %#v\n  code:    %#v", extracted, want)
	}
}

// TestTableKeysAreUnique guards stages.Table's own internal consistency,
// independent of README.md: a hand-authored duplicate key would merge two
// stages in every statistic, silently, so it must fail loudly here even if
// TestStagesMatchReadmeLevelOne's README-side check were somehow bypassed.
func TestTableKeysAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range stages.Table {
		if seen[s.Key] {
			t.Errorf("duplicate stage key %q in stages.Table", s.Key)
		}
		seen[s.Key] = true
	}
}

// stageKeyRE is the naming convention every documented stage key must
// follow: lowercase, dotted once (namespace.stage), hyphen-separated
// within each segment. A key that does not match this shape is not
// self-describing in a shell invocation or a dashboard's group-by, which
// is the entire reason keys exist rather than numeric ids (design.md,
// "Numeric ids. Stable and unreadable.").
var stageKeyRE = regexp.MustCompile(`^[a-z][a-z-]*\.[a-z][a-z0-9-]*$`)

// maxPlausibleKeyLen is the guidance's own bound ("under ~30 characters")
// plus a few characters of headroom for a legitimately long two-word
// stage segment -- not a bound anyone should need to approach in practice.
const maxPlausibleKeyLen = 32

// TestStageKeysFollowNamingConvention pins the key shape the guidance
// prescribes: lowercase, dotted, hyphen-separated, namespaced by the
// command that defines the stage, and short.
func TestStageKeysFollowNamingConvention(t *testing.T) {
	for _, s := range stages.Table {
		if !stageKeyRE.MatchString(s.Key) {
			t.Errorf("stage key %q does not match %s (lowercase, dotted, hyphen-separated)", s.Key, stageKeyRE)
		}
		if len(s.Key) > maxPlausibleKeyLen {
			t.Errorf("stage key %q is %d characters, want at most %d", s.Key, len(s.Key), maxPlausibleKeyLen)
		}
	}
}

// maxPlausibleStageNameLen bounds a documented stage's Name. It is picked
// from the longest legitimate name in stages.Table (62 characters, "Ask
// planning effort, models & panel roster (creating run only)") plus
// generous headroom for reasonable rewording -- not from any inherent
// limit on prose. This guards against the specific defect this task
// exists to fix: a stage "name" that is really a 400-character sentence
// describing an entire command's run, which is what let the pipeline's
// former composite command record one giant "stage" per invocation
// instead of several real ones
// (tasks.md, "Step 4: The test that would catch the next one"). A name
// anywhere near this bound is a description standing in for a stage list,
// not a stage name, whichever command it was written for.
const maxPlausibleStageNameLen = 90

// TestNoDocumentedStageNameIsImplausiblyLong is the guard tasks.md's step 4
// asks for: no documented stage's Name may be long enough to plausibly be
// a sentence describing a whole command rather than one discrete stage.
func TestNoDocumentedStageNameIsImplausiblyLong(t *testing.T) {
	for _, s := range stages.Table {
		if n := len(s.Name); n > maxPlausibleStageNameLen {
			t.Errorf("stage %q: name is %d characters (%q), want at most %d -- "+
				"a name this long is a command description standing in for a stage, not a stage name",
				s.Key, n, s.Name, maxPlausibleStageNameLen)
		}
	}
}

// TestUndocumentedStageKeyIsRejected pins Validate's rejection of a key
// absent from the documented table, and that the rejection names the
// documented alternatives -- exactly what the CLI needs to report a useful
// error instead of a bare "invalid".
func TestUndocumentedStageKeyIsRejected(t *testing.T) {
	err := stages.Validate(stages.Flow, "flow.a-key-nobody-documented")
	if err == nil {
		t.Fatal("Validate: got nil error for an undocumented stage key, want a rejection")
	}
	msg := err.Error()
	if !strings.Contains(msg, "flow.a-key-nobody-documented") {
		t.Errorf("error %q does not name the rejected key", msg)
	}
	for _, s := range stages.Table {
		if !stages.Valid(stages.Flow, s.Key) {
			continue
		}
		if !strings.Contains(msg, s.Key) {
			t.Errorf("error %q does not name documented alternative %q", msg, s.Key)
		}
	}

	var unknown *stages.ErrUnknownStage
	if !asErrUnknownStage(err, &unknown) {
		t.Fatalf("Validate's error is not an *ErrUnknownStage: %v (%T)", err, err)
	}
	if unknown.Command != stages.Flow || unknown.Stage != "flow.a-key-nobody-documented" {
		t.Errorf("ErrUnknownStage = %+v, want Command=%s Stage=%q", unknown, stages.Flow, "flow.a-key-nobody-documented")
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

// TestValidStageKeyIsAccepted breaks the opposite direction: a key taken
// verbatim from stages.Table must be accepted for every command it lists,
// or Validate would reject every legitimate mark and the CLI could never
// mark a stage at all.
func TestValidStageKeyIsAccepted(t *testing.T) {
	for _, s := range stages.Table {
		for _, cmd := range s.Commands {
			if err := stages.Validate(cmd, s.Key); err != nil {
				t.Errorf("Validate(%s, %q) = %v, want nil", cmd, s.Key, err)
			}
		}
	}
}

// TestStageKeyIsRejectedForTheWrongCommand pins that a key documented for
// one command is not silently valid for another -- Valid must key on
// (command, key) together, not merely check key against the table as a
// whole.
func TestStageKeyIsRejectedForTheWrongCommand(t *testing.T) {
	// "flow.brainstorm" is documented for /flow only -- never for the
	// read-only /flow-status, which marks no stages at all.
	if stages.Valid(stages.Command("/flow-status"), "flow.brainstorm") {
		t.Error(`Valid("/flow-status", "flow.brainstorm") = true, want false -- that key is not documented for /flow-status`)
	}
	if !stages.Valid(stages.Flow, "flow.brainstorm") {
		t.Error(`Valid(Flow, "flow.brainstorm") = false, want true`)
	}
}

// TestFlowStatusHasNoDocumentedStages pins that /flow-status -- which
// README.md documents as read-only, marking nothing -- validates no key at
// all.
func TestFlowStatusHasNoDocumentedStages(t *testing.T) {
	for _, s := range stages.Table {
		for _, c := range s.Commands {
			if c == "/flow-status" {
				t.Errorf("stage %q lists /flow-status in Commands, want none -- it is read-only and marks no stages", s.Key)
			}
		}
	}
	if err := stages.Validate("/flow-status", "flow.resolve-change"); err == nil {
		t.Error("Validate(/flow-status, ...) = nil, want a rejection naming that the command has no documented stages")
	}
}

// TestNameLooksUpDocumentedKey pins Name's two outcomes: a documented key
// returns its README-declared name and true; an undocumented one returns
// the zero value and false, rather than panicking or returning a
// misleadingly-present empty name.
func TestNameLooksUpDocumentedKey(t *testing.T) {
	got, ok := stages.Name("flow.review-panel")
	if !ok {
		t.Fatal(`Name("flow.review-panel") ok = false, want true`)
	}
	if got != "The review panel ▸" {
		t.Errorf(`Name("flow.review-panel") = %q, want "The review panel ▸"`, got)
	}

	if _, ok := stages.Name("flow.a-key-nobody-documented"); ok {
		t.Error("Name: ok = true for an undocumented key, want false")
	}
}
