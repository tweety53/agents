// narrativeruns.go is check-narrative-runs.sh ported to Go (KAN-1020).
//
// check-narrative-runs.sh <worktree> <change> <narrative.md>
//
// Compares a change's narrative with the runs the flow store marked for it
// (decision narrative-runs-from-stage-marks). The fix runs are the sessions
// `flow stage sessions -stage flow.document-fix` counts, the integrate runs
// those of `-stage flow.preserve-sessions`; the narrative must hold at least
// as many `## <date> — fix run` and `## <date> — integrate run` sections. A
// narrative that does not exist holds zero sections.
//
// Exit 0: NARRATIVE-RUNS-OK — every marked run has its section.
// Exit 1: NARRATIVE-RUNS-MISSING — a fix or integrate run lacks its section.
// Exit 2: usage error, a store read that did not print one integer, or a
//         narrative that exists and cannot be read.
//
// `flow` is exec'd from env's PATH, as the bash did, so a stub on PATH drives
// it (decision narrative-runs-exec-flow).

package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func init() { Registry["check-narrative-runs"] = checkNarrativeRuns }

var (
	nrInteger   = regexp.MustCompile(`^[0-9]+$`)
	nrFixRun    = regexp.MustCompile(`^## .* — fix run`)
	nrIntegrate = regexp.MustCompile(`^## .* — integrate run`)
)

func checkNarrativeRuns(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: check-narrative-runs.sh <worktree> <change> <narrative.md>")
		return 2
	}
	wt, change, nf := args[0], args[1], args[2]

	sessions := func(stage string) (string, bool) {
		flow := "flow"
		if p, ok := lookPath(env, "flow"); ok {
			flow = p
		}
		cmd := exec.Command(flow, "stage", "sessions", "-C", wt, "-stage", stage, change)
		cmd.Dir = env.Dir
		cmd.Stderr = stderr
		out, ok := capture(cmd)
		if !ok {
			fmt.Fprintf(stderr, "check-narrative-runs: flow stage sessions -stage %s failed\n", stage)
			return "", false
		}
		if !nrInteger.MatchString(out) {
			fmt.Fprintf(stderr, "check-narrative-runs: flow stage sessions -stage %s printed '%s', not one integer\n", stage, out)
			return "", false
		}
		return out, true
	}
	f, ok := sessions("flow.document-fix")
	if !ok {
		return 2
	}
	i, ok := sessions("flow.preserve-sessions")
	if !ok {
		return 2
	}

	hf, hi := 0, 0
	path := nf
	if env.Dir != "" && !filepath.IsAbs(path) {
		path = filepath.Join(env.Dir, path)
	}
	if _, err := os.Stat(path); err == nil {
		body, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "check-narrative-runs: cannot read %s\n", nf)
			return 2
		}
		for _, line := range strings.Split(string(body), "\n") {
			if nrFixRun.MatchString(line) {
				hf++
			}
			if nrIntegrate.MatchString(line) {
				hi++
			}
		}
	}

	fn, _ := strconv.Atoi(f)
	in, _ := strconv.Atoi(i)
	if hf < fn || hi < in {
		fmt.Fprintf(stdout, "NARRATIVE-RUNS-MISSING: %s holds %d '— fix run' section(s) for %s fix run(s) and %d '— integrate run' section(s) for %s integrate run(s) the store marked\n", nf, hf, f, hi, i)
		return 1
	}
	fmt.Fprintf(stdout, "NARRATIVE-RUNS-OK: %s fix run(s), %s integrate run(s), each with its section\n", f, i)
	return 0
}
