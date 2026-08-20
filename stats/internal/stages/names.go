// Package stages holds the documented stage vocabulary for every myflow
// pipeline command -- the keys a `myflow stage begin`/`stage end` mark is
// allowed to carry -- plus the small set of constants the daemon and the
// CLI must agree on to interpret a stage mark's effects, such as
// SyntheticChangeUpdatedBy (synthetic.go). Those constants are not stage
// vocabulary themselves, but they exist only because a stage mark does,
// and one package for both keeps the daemon's writer and the CLI's reader
// from ever drifting onto two copies of the same literal.
//
// The vocabulary is not authored here first: it is a transcription of
// README.md's "Level 1 -- the stages of each command" table, the one place
// this pipeline's stages are documented. TestStagesMatchReadmeLevelOne
// (names_test.go) re-derives the same table mechanically from README.md at
// test time and fails if this package disagrees with it, so the database's
// vocabulary and the documentation stay one list rather than two that can
// drift -- exactly the property task 5's TestProjectKeyMatchesStateFileContract
// already holds for the project-key recipe.
//
// A stage has two identifiers that deliberately do different jobs
// (design.md, "a declared key, and a name that may change"):
//
//   - Key is what a mark carries, what the store groups by, and what never
//     changes. It is declared here, never derived from Name -- a key
//     slugified from the name would change every time the name is
//     reworded, which defeats the entire reason a key exists.
//   - Name is prose for a human reading a dashboard or a mark's rejection
//     message. It may be reworded freely without splitting a stage's
//     recorded history, because nothing is ever keyed on it.
package stages

import (
	"fmt"
	"strings"
)

// Command is a myflow pipeline command name, exactly as README.md's Level 1
// table spells it (including the leading slash).
type Command string

const (
	MyflowStart  Command = "/myflow-start"
	MyflowDo     Command = "/myflow-do"
	MyflowFinish Command = "/myflow-finish"
	MyflowFast   Command = "/myflow-fast"
)

// Stage is one documented pipeline stage: a stable key, a human-readable
// name, and every command that runs it.
//
// Commands is namespaced by the command that *defines* the stage, not by
// the command that merely runs it: `/myflow-fast` chains `/myflow-start`,
// `/myflow-do` and `/myflow-finish` without minting a single name of its
// own, so a stage `/myflow-fast` can run simply lists MyflowFast alongside
// whichever of the other three defines it. That is what makes a fast run
// directly comparable, stage for stage, against the equivalent
// start->do->finish sequence -- design.md's "`/myflow-fast` reuses the
// chained commands' stage names".
type Stage struct {
	Key      string
	Name     string
	Commands []Command
}

// Table is the ordered, hand-authored transcription of README.md's
// "Level 1 -- the stages of each command" stages table. Row order here
// matches the README table's row order; TestStagesMatchReadmeLevelOne
// checks the two are identical, not merely set-equal.
//
// Every /myflow-start, /myflow-do and /myflow-finish stage also lists
// MyflowFast in Commands: design.md's decision is that `/myflow-fast`'s
// allowed stage set is the *union* of the three chained commands',
// expressed here in the Commands column rather than restated as a fourth
// set of names. /myflow-status marks no stages at all and so contributes
// no rows.
var Table = []Stage{
	// /myflow-start
	{Key: "start.resolve-change", Name: "Resolve the change", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.ask-options", Name: "Ask planning effort, models & panel roster (creating run only)", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.brainstorm", Name: "Brainstorm ▸", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.design-approval", Name: "Design approval", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.create-artifacts", Name: "Create the OpenSpec artifacts", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.writing-plans", Name: "Writing-plans ▸", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.publish-proposal", Name: "Publish the proposal artifact", Commands: []Command{MyflowStart, MyflowFast}},
	{Key: "start.write-started", Name: "Write `STARTED`", Commands: []Command{MyflowStart, MyflowFast}},

	// /myflow-do
	{Key: "do.state-gate", Name: "State gate", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.load-context", Name: "Load context and validate the plan", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.isolate-workspace", Name: "Isolate the workspace (first run only)", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.document-fix", Name: "Document the fix (re-runs only)", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.sdd-tdd", Name: "SDD + TDD per task ▸", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.review-panel", Name: "The review panel ▸", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.run-instructions", Name: "Resolve the run instructions", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.workspace-export", Name: "Validate and export workspace isolation", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.lint-and-test", Name: "Run the project's lint and test commands", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.stage-diff", Name: "Stage, excluding the planning paths", Commands: []Command{MyflowDo, MyflowFast}},
	{Key: "do.write-in-progress", Name: "Write `IN_PROGRESS`", Commands: []Command{MyflowDo, MyflowFast}},

	// /myflow-finish
	{Key: "finish.preflight", Name: "Preflight verdict (decides run 1 vs run 2) ▸", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.unfinished-work-gate", Name: "Unfinished-work gate (run 1) ▸", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.landing-question", Name: "The landing question (run 1)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.preserve-sessions", Name: "Preserve the session records (run 1)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.commit-two", Name: "Two commits, implementation first (run 1)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.landing-routes", Name: "The landing routes (run 1) ▸", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.write-in-progress", Name: "Write `IN_PROGRESS` (run 1)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.move-in-review", Name: "Move the issue to In Review (run 1)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.verify-merge", Name: "Verify the merge (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.sync-archive", Name: "Position the checkout, sync delta specs and archive (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.commit-archive", Name: "Commit the archive (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.cleanup", Name: "Cleanup (run 2) ▸", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.verify-cleanup", Name: "Verify the cleanup (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.write-finished", Name: "Write `FINISHED` (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.self-review", Name: "Self-review (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
	{Key: "finish.push-archive", Name: "Push the archive branch and open its PR (run 2)", Commands: []Command{MyflowFinish, MyflowFast}},
}

// byKey indexes Table by Key. byCommand indexes Table by every command that
// runs each stage, preserving Table's row order within each command's
// slice. Both are built once, from the literal Table above, and panic on a
// duplicate key: a duplicate is not a state this package can recover from
// meaningfully, since it would silently merge two different stages in
// every statistic (tasks.md, "Uniqueness is enforced, not assumed").
// TestStagesMatchReadmeLevelOne (names_test.go) is the loud, test-time half
// of this same guard, checked against README.md rather than against
// Table's own internal consistency.
var (
	byKey     map[string]Stage
	byCommand map[Command][]Stage
)

func init() {
	byKey = make(map[string]Stage, len(Table))
	byCommand = make(map[Command][]Stage)
	for _, s := range Table {
		if _, dup := byKey[s.Key]; dup {
			panic(fmt.Sprintf("stages: duplicate key %q in Table -- two stages sharing a key would merge two different things in every statistic", s.Key))
		}
		byKey[s.Key] = s
		for _, c := range s.Commands {
			byCommand[c] = append(byCommand[c], s)
		}
	}
}

// Valid reports whether key is a documented stage key of command.
func Valid(command Command, key string) bool {
	s, ok := byKey[key]
	if !ok {
		return false
	}
	for _, c := range s.Commands {
		if c == command {
			return true
		}
	}
	return false
}

// Name returns the documented human-readable name for key, and whether key
// is documented at all. This is the one place the interface should ever go
// from a stored key to prose -- a mark itself never carries a name, only a
// key.
func Name(key string) (string, bool) {
	s, ok := byKey[key]
	if !ok {
		return "", false
	}
	return s.Name, true
}

// ErrUnknownStage is returned by Validate when key is not a documented
// stage key of command -- either because command itself runs no documented
// stages (including `/myflow-status`, which marks none at all) or because
// key is not among them.
type ErrUnknownStage struct {
	Command Command
	Stage   string
}

func (e *ErrUnknownStage) Error() string {
	alternatives := byCommand[e.Command]
	if len(alternatives) == 0 {
		return fmt.Sprintf("myflow: %q names no documented stage key for %s -- %s has no documented stages in README.md's Level 1 table",
			e.Stage, e.Command, e.Command)
	}
	keys := make([]string, len(alternatives))
	for i, s := range alternatives {
		keys[i] = s.Key
	}
	return fmt.Sprintf("myflow: %q is not a documented stage key of %s -- documented stage keys are: %s",
		e.Stage, e.Command, strings.Join(keys, "; "))
}

// Validate returns nil when key is a documented stage key of command, and
// an *ErrUnknownStage naming the documented alternatives otherwise. This is
// the check that keeps a mark from ever recording an undocumented stage: a
// mark naming a key absent from README.md's Level 1 table is rejected
// here, before it ever reaches the store.
func Validate(command Command, key string) error {
	if Valid(command, key) {
		return nil
	}
	return &ErrUnknownStage{Command: command, Stage: key}
}
