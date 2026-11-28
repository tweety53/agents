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
	// PairJudgeTimeout is pair-judge's model-call deadline; zero means 60 s.
	PairJudgeTimeout time.Duration
	// PairJudgeCoin is the 50/50 draw of the implementer's opus/flow-low pick; nil is a fair coin.
	PairJudgeCoin func() bool
	// SlotGate turns heavySlot on: flow-guard's main sets it, so a test's
	// zero Env never waits on flowd. SlotPoll is the gate's re-ask interval,
	// zero 2 s; SlotRenew a held slot's renewal interval, zero 20 s.
	SlotGate  bool
	SlotPoll  time.Duration
	SlotRenew time.Duration
	// HookRuns is the append-only file `flow-guard hook` records each run's
	// wall time in (records.HookRunsPath); flow-guard's main sets it, so a
	// test's zero Env writes nothing.
	HookRuns string
	// SurvivorsExpire, when non-nil, replaces the survivors bound's timer:
	// the bound fires when it receives, so a test decides when (or never).
	SurvivorsExpire <-chan time.Time
	// Findings returns `flow record findings -change <name>`'s JSON; nil
	// means exec the `flow` CLI on PATH, as the bash guard does.
	Findings func(change string) ([]byte, error)
	// Dispatches returns `flow record dispatches -change <name>`'s JSON; nil
	// means exec the `flow` CLI on PATH, as the bash guard does.
	Dispatches func(change string) ([]byte, error)
	// Reasons is the run's reason collector (see Reason); nil, as in every
	// test's Env{…} literal, makes Reason a no-op.
	Reasons *Reasons
}

// Reasons collects the reason ids one guard run records: distinct ids in
// first-seen order, safe under concurrent calls.
type Reasons struct {
	mu       sync.Mutex
	ids      []string
	headless *Headless
	wait     time.Duration
}

// Wait returns the time the run spent waiting rather than working -- a
// heavy slot's acquisition, await-rows blocked between reads -- which
// flow-guard posts as the run's wait_ms.
func (r *Reasons) Wait() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.wait
}

// recordWait adds d to the run's wait; a nil collector makes it a no-op,
// as Reason.
func recordWait(env Env, d time.Duration) {
	if r := env.Reasons; r != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.wait += d
	}
}

// Headless is the one headless model call a guard run made (pair-judge's
// judge, plan-check's decider): flow-guard posts it with the run, and the
// Role costs view counts it as a run of Role. CostUSD is nil when the
// answer carried none.
type Headless struct {
	Role       string   `json:"role"`
	Model      string   `json:"model"`
	Effort     string   `json:"effort"`
	CostUSD    *float64 `json:"costUsd,omitempty"`
	DurationMS int      `json:"durationMs"`
}

// Headless returns the run's headless call, nil when it made none.
func (r *Reasons) Headless() *Headless {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.headless
}

// recordHeadless keeps h as the run's headless call; a nil collector makes
// it a no-op, as Reason.
func recordHeadless(env Env, h Headless) {
	if r := env.Reasons; r != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.headless = &h
	}
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
