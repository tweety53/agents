// Package guard holds the Go ports of the flow guard scripts. Each guard
// keeps its bash script's CLI contract; scripts/<name>.sh execs flow-guard.
package guard

import (
	"io"
	"time"
)

// Env is everything a guard reads from its process, injected so tests run
// in-process and in parallel.
type Env struct {
	Getenv func(string) string // os.Getenv in production
	Dir    string              // working directory the bash guard would run in
	// Deadlines override the guards' integer-second env knobs in tests only;
	// zero means "parse the env knob as the bash guard does".
	ReproducerBound, ReproducerGrace, SurvivorsTimeout, SurvivorsKillGrace time.Duration
	// SurvivorsExpire, when non-nil, replaces the survivors bound's timer:
	// the bound fires when it receives, so a test decides when (or never).
	SurvivorsExpire <-chan time.Time
	// ProcTable, when non-nil, replaces run-reproducer's process-table read
	// (rrProcTable), so a test can make the read fail.
	ProcTable func() ([]rrProc, error)
	// Findings returns `flow record findings -change <name>`'s JSON; nil
	// means exec the `flow` CLI on PATH, as the bash guard does.
	Findings func(change string) ([]byte, error)
}

// Func is one guard's entry point: args exclude the guard name; the return
// value is the process exit code.
type Func func(args []string, env Env, stdout, stderr io.Writer) int

// Registry maps a guard's basename (without .sh) to its Func. Each port
// task adds its own entry in its own file's init().
var Registry = map[string]Func{}
