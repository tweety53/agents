// Package guard holds the Go ports of the flow guard scripts. Each guard
// keeps its bash script's CLI contract; scripts/<name>.sh execs flow-guard.
package guard

import (
	"io"
	"slices"
	"sync"
	"time"
)

// Env is everything a guard reads from its process, injected so tests run
// in-process and in parallel.
type Env struct {
	Getenv func(string) string // os.Getenv in production
	Dir    string              // working directory the bash guard would run in
	Stdin  io.Reader           // os.Stdin in production; nil reads as empty input
	// LookupEnv is os.LookupEnv in production: bash's `${NAME+set}`, for a
	// guard whose contract tells a set-but-empty variable from an unset one.
	LookupEnv func(string) (string, bool)
	// Environ is os.Environ in production (nil reads as it): the whole
	// environment a guard hands its children.
	Environ func() []string
	// Deadlines override the guards' integer-second env knobs in tests only;
	// zero means "parse the env knob as the bash guard does".
	ReproducerBound, ReproducerGrace, SurvivorsTimeout, SurvivorsKillGrace time.Duration
	// LoadAvg returns the one-minute load average run-reproducer names on a
	// timeout; nil reads it from the machine.
	LoadAvg func() (float64, error)
	// SurvivorsExpire, when non-nil, replaces the survivors bound's timer:
	// the bound fires when it receives, so a test decides when (or never).
	SurvivorsExpire <-chan time.Time
	// MutateExpire, when non-nil, replaces mutate-and-verify's per-run
	// timer: the running pass is killed as timed out when it receives, so a
	// test decides which pass hits the bound and the baseline never races it.
	MutateExpire <-chan time.Time
	// ProcTable, when non-nil, replaces run-reproducer's process-table read
	// (rrProcTable), so a test can make the read fail.
	ProcTable func() ([]rrProc, error)
	// Findings returns `flow record findings -change <name>`'s JSON; nil
	// means exec the `flow` CLI on PATH, as the bash guard does.
	Findings func(change string) ([]byte, error)
	// Dispatches returns `flow record dispatches -change <name>`'s JSON; nil
	// means exec the `flow` CLI on PATH, as the bash guard does.
	Dispatches func(change string) ([]byte, error)
	// Passes returns `flow record passes -change <name>`'s JSON; nil means
	// exec the `flow` CLI on PATH.
	Passes func(change string) ([]byte, error)
	// Decisions returns `flow record decisions -change <name>`'s JSON; nil
	// means exec the `flow` CLI on PATH.
	Decisions func(change string) ([]byte, error)
	// Verdict records `flow record verdict -change <name> -guard <guard>
	// -worktree <worktree> -verdict <line>`; Verdicts returns `flow record
	// verdicts -guard <guard> -false-positive`'s JSON. nil means exec the
	// `flow` CLI on PATH, as the bash guard does.
	Verdict  func(change, guard, worktree, verdict string) error
	Verdicts func(guard string) ([]byte, error)
	// StageKeys returns what `go run ./cmd/flow stage keys` prints from the
	// repository root's stats module; nil means exec it, as the bash guard
	// does.
	StageKeys func() (string, error)
	// Reasons is the run's reason collector (see Reason); nil, as in every
	// test's Env{…} literal, makes Reason a no-op.
	Reasons *Reasons
}

// Reasons collects the reason ids one guard run records: distinct ids in
// first-seen order, safe under concurrent calls.
type Reasons struct {
	mu  sync.Mutex
	ids []string
}

// IDs returns a copy of the recorded ids, in first-seen order.
func (r *Reasons) IDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.ids)
}

// Reason records id as a reason the guard refuses or cannot answer. A guard
// calls it once at each refusal and cannot-answer exit, before returning;
// the collector keeps distinct ids in first-seen order and is safe under
// concurrent calls. An in-process sub-guard gets the same env, so it records
// into the same run. flow-guard posts the ids with the run's telemetry.
func Reason(env Env, id string) {
	r := env.Reasons
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !slices.Contains(r.ids, id) {
		r.ids = append(r.ids, id)
	}
}

// quietReasons is env with no collector, for a guard run in process as a
// probe: its non-zero exit is an answer to the caller, not the caller's
// refusal, so it records nothing; the caller tags its own exit.
func quietReasons(env Env) Env {
	env.Reasons = nil
	return env
}

// Func is one guard's entry point: args exclude the guard name; the return
// value is the process exit code.
type Func func(args []string, env Env, stdout, stderr io.Writer) int

// Registry maps a guard's basename (without .sh) to its Func. Each port
// task adds its own entry in its own file's init().
var Registry = map[string]Func{}
