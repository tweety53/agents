package guard

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"time"
)

// loadGate holds a heavy run -- flow-guard run-verify-list's `test` and
// `visual` lists, flow-guard recapture -- until the one-minute load average
// is at or under the core count, as F30 holds a plan's timing gate
// (skills/flow/brainstorm-planner.md; kan-1091: eight visual runs took 31-49
// minutes each beside other work, 3.2 where the machine was free). Over it,
// name prints one `waiting` line on stderr and the load is read again every
// 15 seconds for up to FLOW_LOAD_WAIT seconds, 1800 unset; still over it
// then, loadGate returns false and the caller runs nothing, its reason
// `load-over-cores`. A load that
// cannot be read, and an Env whose LoadGate is unset, hold nothing.
func loadGate(env Env, name string, stderr io.Writer) bool {
	if !env.LoadGate {
		return true
	}
	read := env.LoadAvg
	if read == nil {
		read = rrLoadAvg
	}
	wait := 1800 * time.Second
	if s, err := strconv.Atoi(env.Getenv("FLOW_LOAD_WAIT")); err == nil && s >= 0 {
		wait = time.Duration(s) * time.Second
	}
	poll := env.LoadPoll
	if poll == 0 {
		poll = 15 * time.Second
	}
	cores := float64(runtime.NumCPU())
	for start, told := time.Now(), false; ; time.Sleep(poll) {
		load, err := read()
		if err != nil || load <= cores {
			return true
		}
		if time.Since(start) >= wait {
			fmt.Fprintf(stderr, "%s: load average %.2f still over %.0f cores after %s — no command ran; re-run once the load average is at or under %.0f\n", name, load, cores, wait, cores)
			return false
		}
		if !told {
			fmt.Fprintf(stderr, "%s: load average %.2f over %.0f cores — waiting up to %s for it to drop\n", name, load, cores, wait)
			told = true
		}
	}
}
