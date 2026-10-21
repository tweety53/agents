package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// checkDevStackFresh is scripts/check-dev-stack-fresh.sh's guard. The shim's
// header states the contract — the project's declared `fingerprint` row as
// the check, and the three exit codes; this file is the mechanism.
func init() {
	Registry["check-dev-stack-fresh"] = checkDevStackFresh
}

var (
	dsfClose = regexp.MustCompile(`^##[ \t\n\v\f\r]+`)
	dsfBlank = regexp.MustCompile(`^[ \t\n\v\f\r]*$`)
)

func checkDevStackFresh(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "check-dev-stack-fresh: usage: check-dev-stack-fresh.sh <worktree>")
		return 2
	}
	worktree := args[0]
	if !isDir(smcAbs(env, worktree)) {
		fmt.Fprintf(stderr, "check-dev-stack-fresh: %s is not a directory\n", worktree)
		return 2
	}
	cfg := worktree + "/.flow/project.md"
	if _, err := os.Lstat(smcAbs(env, cfg)); err != nil {
		fmt.Fprintf(stderr, "check-dev-stack-fresh: %s does not exist — no fingerprint is declared, so freshness cannot be checked\n", cfg)
		return 2
	}
	if !isFile(smcAbs(env, cfg)) {
		fmt.Fprintf(stderr, "check-dev-stack-fresh: %s is not a regular file\n", cfg)
		return 2
	}
	// An unreadable file reaches the bash's awk as empty input: no row.
	text, _ := os.ReadFile(smcAbs(env, cfg))
	fpr, found := dsfFingerprint(string(text))
	if !found {
		fmt.Fprintf(stderr, "check-dev-stack-fresh: %s declares no readable `fingerprint` row under `## visual verification` — freshness cannot be checked\n", cfg)
		return 2
	}
	if fpr == "" {
		fmt.Fprintf(stderr, "check-dev-stack-fresh: the fingerprint row in %s resolves to an empty command — freshness cannot be checked\n", cfg)
		return 2
	}

	// Run the declared command from the worktree, streaming its output — the
	// evidence on a stale or dead stack is the command's own diff or curl
	// noise, not a verdict line.
	fmt.Fprintf(stdout, "check-dev-stack-fresh: fingerprint: %s\n", fpr)
	bash := "bash"
	if p, ok := lookPath(env, "bash"); ok {
		bash = p
	}
	cmd := exec.Command(bash, "-c", fpr)
	cmd.Dir = smcAbs(env, worktree)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = env.Stdin, stdout, stderr
	if cmd.Run() == nil {
		fmt.Fprintln(stdout, "stack-fresh: the fingerprint exited 0 — the stack is serving the worktree's build")
		return 0
	}
	fmt.Fprintln(stderr, "stack-stale: the fingerprint exited non-zero — the stack is not serving the worktree's build (stale bundle, or nothing serving)")
	return 1
}

// dsfFingerprint is the bash's row-extraction awk: the first `fingerprint`
// row of a `| Command | Runs |` table under `## visual verification`, code
// fences skipped so a worked example never declares a row.
func dsfFingerprint(text string) (string, bool) {
	insec, infence, incmd := false, false, false
	for _, l := range vvLines(text) {
		switch {
		case strings.HasPrefix(l, "```"):
			infence = !infence
			continue
		case infence:
			continue
		case !insec:
			insec = vvHeading.MatchString(l)
			continue
		case dsfClose.MatchString(l):
			return "", false
		case dsfBlank.MatchString(l), !strings.HasPrefix(l, "|"):
			incmd = false
			continue
		}
		cells := vtSplitCells(l)
		if len(cells) < 2 {
			continue
		}
		key := vtFoldCell(cells[0])
		if key == "command" && vtFoldCell(cells[1]) == "runs" {
			incmd = true
			continue
		}
		if incmd && key == "fingerprint" {
			return vtTrimCell(cells[1]), true
		}
	}
	return "", false
}
