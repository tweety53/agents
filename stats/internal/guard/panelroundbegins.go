package guard

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// panel-round-begins (scripts/panel-round-begins.sh) records every slot's
// begin of one review-panel round, each under the literal key
// panel-<round>-<slot> -- the round's begins block the parent hand-wrote
// before, where kan-1026's `-key panel-$R-$s` from `R=$(flow record
// next-round ...)` left the daemon nothing to read the key from
// (skills/flow/review-panel.md, "Every slot's dispatch is recorded").
// Exit 0 every begin recorded; 1 a begin failed (its output relayed, the
// rest still recorded); 2 usage, or no flow binary on PATH.
func init() {
	Registry["panel-round-begins"] = panelRoundBegins
}

const prbUsage = "usage: panel-round-begins.sh <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...\n"

var prbRound = regexp.MustCompile(`^[0-9]+$`)

func panelRoundBegins(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 5 || !prbRound.MatchString(args[2]) {
		fmt.Fprint(stderr, prbUsage)
		return 2
	}
	wt, change, round, token := args[0], args[1], args[2], args[3]
	type slot struct{ name, model, effort, diffBase string }
	var slots []slot
	for _, spec := range args[4:] {
		f := strings.Split(spec, ":")
		if len(f) < 3 || len(f) > 4 || f[0] == "" || f[1] == "" || f[2] == "" {
			fmt.Fprintf(stderr, "panel-round-begins: slot %q is not <slot>:<model>:<effort>[:<diff-base>]\n%s", spec, prbUsage)
			return 2
		}
		s := slot{name: f[0], model: f[1], effort: f[2]}
		if len(f) == 4 {
			s.diffBase = f[3]
		}
		slots = append(slots, s)
	}
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintln(stderr, "panel-round-begins: no flow binary on PATH")
		return 2
	}
	code := 0
	for _, s := range slots {
		key := "panel-" + round + "-" + s.name
		argv := []string{"record", "dispatch", "begin", "-C", wt, "-change", change, "-role", "reviewer",
			"-slot", s.name, "-model", s.model, "-effort", s.effort}
		if s.diffBase != "" {
			argv = append(argv, "-diff-base", s.diffBase)
		}
		argv = append(argv, "-key", key, "-session-token", token)
		rc, out := kwRun(env, "", flow, argv...)
		fmt.Fprintf(stdout, "%s: %s\n", key, strings.TrimSpace(out))
		if rc != 0 {
			code = 1
		}
	}
	return code
}
