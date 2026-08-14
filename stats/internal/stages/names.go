// Package stages holds the documented stage vocabulary for every myflow
// pipeline command -- the names a `myflow stage begin`/`stage end` mark is
// allowed to carry.
//
// The vocabulary is not authored here first: it is a transcription of
// README.md's "Level 1 -- the stages of each command" table, the one place
// this pipeline's stage names are documented. TestStageNamesMatchReadmeLevelOne
// (names_test.go) re-derives the same list mechanically from README.md at
// test time and fails if this package disagrees with it, so the database's
// vocabulary and the documentation stay one list rather than two that can
// drift -- exactly the property task 5's TestProjectKeyMatchesStateFileContract
// already holds for the project-key recipe.
package stages

import (
	"fmt"
	"strings"
)

// Command is a myflow pipeline command name, exactly as README.md's Level 1
// table spells it in its "Command" column (including the leading slash).
type Command string

const (
	MyflowStart  Command = "/myflow-start"
	MyflowDo     Command = "/myflow-do"
	MyflowFinish Command = "/myflow-finish"
	MyflowFast   Command = "/myflow-fast"
)

// Names holds, for each command, the stage names documented in README.md's
// "Level 1 -- the stages of each command" table, transcribed from its
// "Stages, in order" column by the same mechanical rule
// TestStageNamesMatchReadmeLevelOne applies at test time:
//
//   - the cell is split on "→" and ";" (";" only appears where a row's
//     cell branches -- `/myflow-finish`'s two runs, `/myflow-fast`'s three
//     routes -- and both characters mark "what comes next in this
//     command's stage sequence" in the prose);
//   - each resulting segment has every " ▸" substructure marker removed
//     (▸ means "expanded at level 2 below", not part of the name) and is
//     then trimmed.
//
// `/myflow-status` is deliberately absent from this map: its row states
// plainly that it is "read-only -- no stages, no state write", so it has
// no entry rather than an entry mapping to an empty (and therefore
// silently-always-valid-looking) slice.
var Names = map[Command][]string{
	MyflowStart: {
		"resolve the change",
		"ask the planning effort, the three model choices and the review panel roster *(creating run only)*",
		"brainstorm",
		"design approval",
		"create the OpenSpec artifacts",
		"writing-plans",
		"publish the proposal artifact",
		"write `STARTED`",
	},
	MyflowDo: {
		"state gate",
		"load context and validate the plan",
		"isolate the workspace *(first run only)*",
		"document the fix *(re-runs only)*",
		"SDD + TDD per task",
		"the review panel",
		"resolve the run instructions",
		"validate the project's `## workspace isolation` section, then export what it declares",
		"run the project's lint and test commands",
		"stage, excluding the planning paths",
		"write `IN_PROGRESS`",
	},
	MyflowFinish: {
		"the preflight verdict, taken once per worktree in the resolved set, decides which run follows — *run 1:* the unfinished-work gate",
		"the landing question",
		"preserve the session records",
		"two commits, implementation first",
		"the landing routes",
		"write `IN_PROGRESS`",
		"move the issue to In Review",
		"*run 2:* verify the merge",
		"sync delta specs and archive",
		"commit and push the archive",
		"cleanup",
		"verify the cleanup",
		"write `FINISHED`",
		"self-review",
	},
	MyflowFast: {
		"state gate",
		"*(creating run)* brainstorming chained straight into implementation and the review panel, no gate in between — the full sequence, by cited section, is **No state file — brainstorm into implementation** (`skills/myflow-fast/SKILL.md`)",
		"*(at `IN_PROGRESS`, argument present)* a fix, chained the same way — **At `IN_PROGRESS`** (`skills/myflow-fast/SKILL.md`)",
		"*(bare at `IN_PROGRESS`)* the landing question, and merge-and-push alone continues into the archive sequence within the same invocation — same section, plus **After merge-and-push specifically** (`skills/myflow-fast/SKILL.md`)",
	},
}

// Valid reports whether stage is a documented stage name for command.
func Valid(command Command, stage string) bool {
	for _, s := range Names[command] {
		if s == stage {
			return true
		}
	}
	return false
}

// ErrUnknownStage is returned by Validate when stage is not a documented
// stage of command -- either because command itself has no entry in Names
// (including `/myflow-status`, which marks no stages at all) or because
// its documented stages do not include stage.
type ErrUnknownStage struct {
	Command Command
	Stage   string
}

func (e *ErrUnknownStage) Error() string {
	alternatives, ok := Names[e.Command]
	if !ok || len(alternatives) == 0 {
		return fmt.Sprintf("myflow: %q names no documented stage for %s -- %s has no documented stages in README.md's Level 1 table",
			e.Stage, e.Command, e.Command)
	}
	return fmt.Sprintf("myflow: %q is not a documented stage of %s -- documented stages are: %s",
		e.Stage, e.Command, strings.Join(alternatives, "; "))
}

// Validate returns nil when stage is a documented stage of command, and an
// *ErrUnknownStage naming the documented alternatives otherwise. This is
// the check that keeps a mark from ever recording an undocumented stage
// name: a mark naming a stage absent from README.md's Level 1 table is
// rejected here, before it ever reaches the store.
func Validate(command Command, stage string) error {
	if Valid(command, stage) {
		return nil
	}
	return &ErrUnknownStage{Command: command, Stage: stage}
}
