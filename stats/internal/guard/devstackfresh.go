package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-dev-stack-fresh.sh shim (KAN-1058).
//
// flow-guard check-dev-stack-fresh — verify, before a handoff prints a running dev
// stack's URL, that the stack is serving the worktree's own build rather
// than one committed before HEAD.
//
// Usage: flow-guard check-dev-stack-fresh <worktree>
//
// ONE ARGUMENT, THE WORKTREE — the directory holding `.flow/project.md` and
// the build the stack ought to be serving, matching flow-guard check-visual-verification's
// own project-root convention.
//
// KAN-334's failure, measured on kan-333: a handoff printed `localhost:3000`
// while the dev server behind it was serving a bundle eleven commits stale —
// the watcher had stopped recompiling, the log had stopped growing, and
// nothing in the pipeline looked. The start rule in the run-instructions
// stage rebuilds and restarts what it can, but a stack it only skips or
// merely starts (a start command that does not rebuild, a watcher that has
// died) is trusted rather than checked. This guard is the check.
//
// THE CHECK IS THE PROJECT'S OWN DECLARED FINGERPRINT. A project tells the
// pipeline once, in `.flow/project.md`'s `## visual verification` section,
// how to prove a served bundle is the worktree's build: the `fingerprint`
// command row, which rebuilds first and then compares the served document
// against the rebuild (this repository's own row is the worked example).
// Reusing that row is the whole design: no second per-project mechanism is
// invented, no new project-configuration vocabulary is added, and a project
// that already declared the row for visual verification needs to write
// nothing. The row's own exit code is the verdict.
//
// THREE EXIT CODES, flow-guard project-get's convention:
//   0  fresh — a `fingerprint` row is declared and its command exited 0.
//   1  stale (or not serving) — the command is declared and exited non-zero.
//      The restart action is the same for both, so the caller words one line.
//   2  cannot answer — usage; `<worktree>` is not a directory; no
//      `.flow/project.md`; no `## visual verification` section; the section
//      carries no readable `fingerprint` row. The caller prints its
//      freshness-unverified gap line from this exit, so the reason goes to
//      stderr and the exit is never a silent pass.
//
// Exit 1 carries the fingerprint command's own output, streamed verbatim —
// it is the evidence, and a diff or a curl failure names itself.
//
// THE DECLARED COMMAND IS EXECUTED — that is this guard's purpose, and the
// one thing that separates it from flow-guard check-visual-verification, which
// validates the same section and deliberately runs nothing. The command
// comes from the project's own tracked `.flow/project.md`, the same trust
// every `## run` command the pipeline executes already carries.
//
// The row is read by the same table helpers every visual guard uses —
// vtSplitCells, vtTrimCell and vtFoldCell, after stripBOM — and code fences
// are skipped so a worked example inside the section can never declare a
// row. The mechanism is stats/internal/guard/devstackfresh.go; its test is
// stats/internal/guard/check_dev_stack_fresh_test.go.

// checkDevStackFresh is `flow-guard check-dev-stack-fresh`. The comment above states the contract — the project's declared `fingerprint` row as
// the check, and the three exit codes; this file is the mechanism.
func init() {
	Registry["check-dev-stack-fresh"] = checkDevStackFresh
}

var (
	dsfClose = regexp.MustCompile("^##[" + cSpace + "]+")
	dsfBlank = regexp.MustCompile("^[" + cSpace + "]*$")
)

func checkDevStackFresh(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "check-dev-stack-fresh: usage: flow-guard check-dev-stack-fresh <worktree>")
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
	cmd.Env = append(os.Environ(), "PATH="+env.Getenv("PATH"))
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
