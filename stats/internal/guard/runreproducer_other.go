//go:build !darwin

package guard

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// rrProcTable is one read of the process table; a failed read is an error,
// never an empty table (runreproducer_darwin.go says why). Off darwin it is
// read through ps(1); runreproducer_darwin.go says why darwin reads it
// in-process.
func rrProcTable() ([]rrProc, error) {
	out, err := exec.Command("ps", "-A", "-o", "pid=", "-o", "ppid=", "-o", "pgid=").Output()
	if err != nil {
		return nil, fmt.Errorf("ps -A: %w", err)
	}
	var procs []rrProc
	for _, l := range strings.Split(string(out), "\n") {
		f := strings.Fields(l)
		if len(f) != 3 {
			continue
		}
		var p rrProc
		var e1, e2, e3 error
		p.pid, e1 = strconv.Atoi(f[0])
		p.ppid, e2 = strconv.Atoi(f[1])
		p.pgid, e3 = strconv.Atoi(f[2])
		if e1 == nil && e2 == nil && e3 == nil {
			procs = append(procs, p)
		}
	}
	return procs, nil
}
