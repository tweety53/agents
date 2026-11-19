// Package stages holds the documented stage vocabulary for every flow
// pipeline command -- the keys a `flow stage begin`/`stage end` mark is
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
// Table is also the vocabulary's one SERVED source: Keys() hands the keys
// to `flow stage keys`, so no served consumer of the vocabulary keeps a
// copy that could drift from it. (Table itself
// stays an exported slice deliberately -- the package's own external test
// reads its rows from here.)
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

// Command is a flow pipeline command name, exactly as README.md's Level 1
// table spells it (including the leading slash).
type Command string

const (
	Flow     Command = "/flow"
	FlowFast Command = "/flow-fast"
	FlowPlan Command = "/flow-plan"

	FlowInteractive Command = "/flow-interactive"
)

// Stage is one documented pipeline stage: a stable key, a human-readable
// name, and every command that runs it.
//
// Commands is namespaced by the command that *defines* the stage.
type Stage struct {
	Key      string
	Name     string
	Commands []Command
}

// Table is the ordered, hand-authored transcription of README.md's
// "Level 1 -- the stages of each command" stages table. Row order here
// matches the README table's row order; TestStagesMatchReadmeLevelOne
// checks the two are identical, not merely set-equal.
var Table = []Stage{
	// /flow -- mints its own flow.* namespace rather than reusing start./do./finish.,
	// per design.md's flow-rename-content-split (see README.md's Level 1 table for why).
	{Key: "flow.plan", Name: "Plan — kickoff to the plan gate", Commands: []Command{Flow, FlowFast, FlowInteractive, FlowPlan}},
	{Key: "flow.implement", Name: "Implement", Commands: []Command{Flow, FlowFast, FlowInteractive}},
	{Key: "flow.review", Name: "Review", Commands: []Command{Flow, FlowFast}},
	{Key: "flow.verify", Name: "Verify: workspace isolation, lint and test", Commands: []Command{Flow, FlowFast, FlowInteractive}},
	{Key: "flow.land", Name: "Land", Commands: []Command{Flow, FlowFast, FlowInteractive}},
	{Key: "flow.self-review", Name: "Run the self-review pass (run 1)", Commands: []Command{Flow, FlowFast, FlowInteractive}},
	{Key: "flow.pipeline-fix", Name: "Pipeline fix — a fix to the flow tooling a run makes", Commands: []Command{Flow, FlowFast}},
	{Key: "flow.cleanup", Name: "Cleanup (run 2) ▸", Commands: []Command{Flow, FlowFast, FlowInteractive}},
	// /flow-plan -- plan.session is its own stage: the research session up to
	// a capture's kickoff, recorded so its cost joins the change it seeds. A
	// capture's planning marks flow.plan, which lists FlowPlan for that
	// reason.
	{Key: "plan.session", Name: "Plan session — `/flow-plan` research up to a capture's kickoff", Commands: []Command{FlowPlan}},
}

// Retired maps a stage key only historical rows carry to the Table key that
// replaced it, so a reader counts old rows with the new stage.
var Retired = map[string]string{
	"flow.design-approval":       "flow.plan",
	"flow.plan-review":           "flow.plan",
	"flow.kickoff":               "flow.plan",
	"flow.brainstorm":            "flow.plan",
	"flow.create-artifacts":      "flow.plan",
	"flow.writing-plans":         "flow.plan",
	"flow.decide":                "flow.plan",
	"flow.load-context":          "flow.implement",
	"flow.isolate-workspace":     "flow.implement",
	"flow.document-fix":          "flow.implement",
	"flow.sdd-tdd":               "flow.implement",
	"flow.review-panel":          "flow.review",
	"flow.visual-verify":         "flow.verify",
	"flow.stage-diff":            "flow.verify",
	"flow.run-instructions":      "flow.verify",
	"flow.write-in-progress":     "flow.verify",
	"flow.preflight":             "flow.land",
	"flow.unfinished-work-gate":  "flow.land",
	"flow.landing-question":      "flow.land",
	"flow.preserve-sessions":     "flow.land",
	"flow.commit-two":            "flow.land",
	"flow.landing-routes":        "flow.land",
	"flow.verify-merge":          "flow.cleanup",
	"flow.write-finished":        "flow.cleanup",
	"flow.refresh-main-checkout": "flow.cleanup",
	"flow.commit-archive":        "flow.land",
	"flow.verify-cleanup":        "flow.cleanup",
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

// Keys returns every documented stage key, in Table's row order, on a
// slice freshly allocated for the caller -- mutating the result never
// reaches package state. This is the serving half of the vocabulary: the
// `flow stage keys` subcommand prints it, so every served consumer reads
// the same list this package validates marks against.
func Keys() []string {
	out := make([]string, len(Table))
	for i, s := range Table {
		out[i] = s.Key
	}
	return out
}

// ForCommand returns command's stage keys in Table order -- the order a run
// of that command marks them -- or nil for an unknown command.
func ForCommand(command Command) []string {
	out := make([]string, 0, len(byCommand[command]))
	for _, s := range byCommand[command] {
		out = append(out, s.Key)
	}
	return out
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
// stages (including `/flow-status`, which marks none at all) or because
// key is not among them.
type ErrUnknownStage struct {
	Command Command
	Stage   string
}

func (e *ErrUnknownStage) Error() string {
	alternatives := byCommand[e.Command]
	if len(alternatives) == 0 {
		return fmt.Sprintf("flow: %q names no documented stage key for %s -- %s has no documented stages in README.md's Level 1 table",
			e.Stage, e.Command, e.Command)
	}
	keys := make([]string, len(alternatives))
	for i, s := range alternatives {
		keys[i] = s.Key
	}
	return fmt.Sprintf("flow: %q is not a documented stage key of %s -- documented stage keys are: %s",
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
