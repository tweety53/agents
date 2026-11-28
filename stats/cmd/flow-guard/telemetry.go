package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/records"
)

// telemetry.go records every guard this binary dispatches into flowd's
// guard_runs table, so the flow-health dashboard can answer which guards ever
// fire -- see stats/internal/store/migrations/0032_guard_runs.sql. It appends
// the run as one JSON line to records.GuardRunsPath with
// records.AppendJSONLine, the file `flow-guard hook` buffers its runs in
// beside, and spawns nothing on exit: no git call, no flowd round trip. flowd
// ingests the file (cmd/flowd/hookruns.go's ingestGuardRuns), resolving each
// line's project from its worktree.
//
// It is strictly best-effort and never changes a guard's answer: a file that
// cannot be written is ignored, and a line whose worktree names no project
// the store holds (a guard run inside a test fixture's throwaway repository)
// is dropped at ingestion.

// telemetryOff names FLOW_GUARD_TELEMETRY's opt-out values.
var telemetryOff = map[string]bool{"0": true, "off": true, "false": true, "no": true}

// guardRun is records.GuardRun's wire shape, with guard.Headless for its
// headless call.
type guardRun struct {
	Guard      string `json:"guard"`
	Worktree   string `json:"worktree"`
	ExitCode   int    `json:"exitCode"`
	Outcome    string `json:"outcome"`
	DurationMS int    `json:"durationMs"`
	// WaitMS is the part of DurationMS the guard spent waiting, not
	// working (guard.Reasons.Wait).
	WaitMS int `json:"waitMs,omitempty"`
	// BootstrapMS is the time from scripts/flow-guard's start to this
	// binary's dispatch (bootstrapMS), absent when unstamped.
	BootstrapMS *int      `json:"bootstrapMs,omitempty"`
	RecordedAt  time.Time `json:"recordedAt"`
	// Reasons is the run's distinct reason ids, posted only when the run is
	// not clear; absent, the store records NULL.
	Reasons []string `json:"reasons,omitempty"`
	// Headless is the run's headless model call, absent when it made none.
	Headless *core.Headless `json:"headless,omitempty"`
}

// outcomeFor classifies guard name's exit code and the first line it printed
// on stdout: 0 is clear -- save check-unfinished-work's first line opening
// `OUTSTANDING:`, the verdict it prints on exit 0, which is the guard firing;
// any other guard printing that prefix on exit 0 is clear --
// the guard's own could-not-answer code is cannot-answer, and anything else
// -- a violation, a refusal -- is the guard firing.
func outcomeFor(name string, code int, first string) string {
	switch {
	case code == 0 && name == "check-unfinished-work" && strings.HasPrefix(first, "OUTSTANDING:"):
		return "fired"
	case code == 0:
		return "clear"
	case code == core.CannotAnswer(name):
		return "cannot-answer"
	default:
		return "fired"
	}
}

// gitCheckout reports whether dir lies in a git checkout: it or an ancestor
// holds a .git entry -- a stat walk, so the check forks no git.
func gitCheckout(dir string) bool {
	for d := filepath.Clean(dir); ; d = filepath.Dir(d) {
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			return true
		}
		if filepath.Dir(d) == d {
			return false
		}
	}
}

// firstLine is an io.Writer tee that keeps the first line written through it,
// up to firstLineMax bytes -- what outcomeFor reads.
type firstLine struct {
	w    io.Writer
	line []byte
	done bool
}

const firstLineMax = 256

func (f *firstLine) Write(p []byte) (int, error) {
	if !f.done {
		rest := p
		if i := bytes.IndexByte(rest, '\n'); i >= 0 {
			rest, f.done = rest[:i], true
		}
		if room := firstLineMax - len(f.line); room > 0 {
			f.line = append(f.line, rest[:min(len(rest), room)]...)
		}
	}
	return f.w.Write(p)
}

// recordRun records one dispatched guard's run, and reports nothing: every
// failure is swallowed, per this file's header.
func recordRun(getenv func(string) string, dir, name string, code int, first string, reasons []string, headless *core.Headless, wait, elapsed time.Duration) {
	if telemetryOff[strings.ToLower(getenv("FLOW_GUARD_TELEMETRY"))] {
		return
	}
	now := time.Now()
	run := guardRun{
		Guard: name, Worktree: dir, ExitCode: code, Outcome: outcomeFor(name, code, first),
		DurationMS: int(elapsed.Milliseconds()), WaitMS: int(wait.Milliseconds()), RecordedAt: now.UTC(), Headless: headless,
		BootstrapMS: bootstrapMS(getenv("FLOW_GUARD_START"), now.Add(-elapsed)),
	}
	if run.Outcome != "clear" {
		run.Reasons = reasons
	}
	records.AppendJSONLine(records.GuardRunsPath(getenv), run)
}

// bootstrapMS is the milliseconds from stamp -- FLOW_GUARD_START, the
// $EPOCHREALTIME scripts/flow-guard exports on its first line (seconds, a
// locale's decimal point or comma, microseconds) -- to start, the guard's
// dispatch: the bootstrap's hashing, any build, and the exec. nil when the
// stamp is absent (a bash without $EPOCHREALTIME, a direct call of the
// binary), will not parse, or lies after start.
func bootstrapMS(stamp string, start time.Time) *int {
	secs, err := strconv.ParseFloat(strings.Replace(stamp, ",", ".", 1), 64)
	if err != nil {
		return nil
	}
	d := start.Sub(time.UnixMicro(int64(secs * 1e6)))
	if d < 0 {
		return nil
	}
	ms := int(d.Milliseconds())
	return &ms
}

// dispatched reports whether args name a registered guard -- the only runs
// recorded; a usage error is not a guard's answer. A `hook` run is never
// recorded here: a hook records its own wall time (guard/hook/hook.go).
func dispatched(args []string) (string, bool) {
	if len(args) == 0 || args[0] == "hook" {
		return "", false
	}
	if _, ok := registry[args[0]]; !ok {
		return "", false
	}
	return args[0], true
}

// workingDir is the directory the guard ran in, or "" when it cannot be
// read (run itself then answered cannot-answer, and there is no project to
// record against).
func workingDir() string {
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	return d
}

// guardWorktree is the worktree a guard run is recorded against (decision
// guard-run-worktree-arg): the first of args[1:] that is an existing
// directory whose cleaned absolute path carries a `-worktrees/<change>`
// segment -- a guard driven from the main checkout at a change's worktree --
// else cwd, the directory the guard ran in.
func guardWorktree(args []string, cwd string) string {
	for _, a := range args[min(1, len(args)):] {
		p, err := filepath.Abs(a)
		if err != nil || !underWorktrees(p) {
			continue
		}
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	return cwd
}

// underWorktrees reports whether path p has a segment ending `-worktrees`
// followed by a further segment.
func underWorktrees(p string) bool {
	parts := strings.Split(filepath.ToSlash(p), "/")
	for i, s := range parts[:max(0, len(parts)-1)] {
		if strings.HasSuffix(s, "-worktrees") && len(s) > len("-worktrees") && parts[i+1] != "" {
			return true
		}
	}
	return false
}
