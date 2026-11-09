package main

// interactive.go — `flow interactive land` (kan-1081 self-review, F4: the
// session landed each request through a hand-chained fetch, rebase and push
// in a bash helper, nothing checking the commit it pushed).
//
// Usage: flow interactive land [-addr url] [-timeout dur] [-C worktree]
//          -change name -n request -harness name -session-token token
//
// -C is the session's worktree (default: cwd); every other flag but -addr and
// -timeout is required, -n a positive integer. In order, each refusal one
// stderr line and exit 1:
//   1. `git status --porcelain --untracked-files=all` non-empty — refused;
//   2. the base is flow-guard resolve-base-branch's answer; `git fetch
//      origin`, then `git rebase origin/<base>` — a failing rebase is
//      refused and left in place for the **Conflict** bullet of **Sync the
//      branch onto the base**;
//   3. base := `git rev-parse origin/<base>`;
//   4. HEAD tracking a path under .superpowers/ — refused, naming it;
//   5. a declared `## lint` list whose flow-lint-green does not hold HEAD —
//      refused: a rebase that moved HEAD needs the list run again;
//   6. `flow stage begin -command /flow-interactive -stage
//      flow.landing-routes -request <n>`, then `git push origin
//      <branch>:<base>`, `git fetch origin` and `git merge-base
//      --is-ancestor HEAD origin/<base>` — a failure of any ends the stage
//      `failed` with a one-line -reason, exit 1;
//   7. otherwise the stage ends `completed` and stdout carries
//      `commit: <HEAD>` and `base: <base>`, exit 0.
// Exit 2 is usage.

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard"
)

const interactiveUsage = `usage: flow interactive land [-addr url] [-timeout dur] [-C worktree]
                              -change name -n request -harness name -session-token token
`

// runInteractive routes `flow interactive`.
func runInteractive(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "land" {
		fmt.Fprint(stderr, interactiveUsage)
		return 2
	}
	return runInteractiveLand(ctx, args[1:], stdout, stderr)
}

func runInteractiveLand(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow interactive land", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var addr string
	registerRecordsAddrFlag(fset, &addr)
	timeout := fset.Duration("timeout", defaultTimeout, "store request timeout before falling back")
	wt := fset.String("C", "", "the session's worktree (default: cwd)")
	change := fset.String("change", "", "the change name (required)")
	n := fset.Int("n", 0, "the request's number in the session, a positive integer (required)")
	harness := fset.String("harness", "", "the harness running the session (required)")
	token := fset.String("session-token", "", "the session's literal token (required)")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprint(stderr, interactiveUsage)
		return 2
	}
	if fset.NArg() != 0 || *change == "" || *n <= 0 || *harness == "" || *token == "" {
		fmt.Fprintln(stderr, "flow: -change, a positive -n, -harness and -session-token are required, and no argument")
		fmt.Fprint(stderr, interactiveUsage)
		return 2
	}
	if *wt == "" {
		d, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "flow: %v\n", err)
			return 2
		}
		*wt = d
	}

	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "flow: interactive land refused: "+format+"\n", a...)
		return 1
	}
	git := func(a ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", *wt}, a...)...).CombinedOutput()
		s := strings.TrimSpace(string(out))
		if err != nil {
			return s, fmt.Errorf("git %s: %v: %s", strings.Join(a, " "), err, s)
		}
		return s, nil
	}

	st, err := git("status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return refuse("%v", err)
	}
	if st != "" {
		return refuse("%s is not clean — commit or remove: %s", *wt, strings.SplitN(st, "\n", 2)[0])
	}
	var out, errBuf strings.Builder
	if code := guard.Registry["resolve-base-branch"]([]string{*wt}, guard.Env{Getenv: os.Getenv, Dir: *wt}, &out, &errBuf); code != 0 {
		return refuse("resolve-base-branch exited %d on %s: %s", code, *wt, strings.TrimSpace(errBuf.String()))
	}
	baseBranch := strings.TrimSpace(out.String())
	if _, err := git("fetch", "origin"); err != nil {
		return refuse("%v", err)
	}
	if _, err := git("rebase", "origin/"+baseBranch); err != nil {
		return refuse("%v — resolve it per the Conflict bullet of Sync the branch onto the base, then land again", err)
	}
	base, err := git("rev-parse", "origin/"+baseBranch)
	if err != nil {
		return refuse("%v", err)
	}
	head, err := git("rev-parse", "HEAD")
	if err != nil {
		return refuse("%v", err)
	}
	if first, err := scratchTracked(*wt, head); err != nil {
		return refuse("%v", err)
	} else if first != "" {
		return refuse("HEAD %s tracks %s — untrack .superpowers/ (git rm -r --cached .superpowers), commit, and land again", head, first)
	}
	if lintDeclared(*wt) && lintGreenAt(*wt) != head {
		return refuse("HEAD %s is not the commit a green lint list stamped (flow-lint-green holds %q) — run flow-guard run-verify-list %s lint (step 4), then land again", head, lintGreenAt(*wt), *wt)
	}

	conn := []string{"-addr", addr, "-timeout", timeout.String(), "-C", *wt, "-command", "/flow-interactive", "-stage", "flow.landing-routes"}
	begin := append(append([]string{}, conn...), "-request", strconv.Itoa(*n), "-harness", *harness, "-session-token", *token, *change)
	if code := runStageBegin(ctx, begin, stderr); code != 0 {
		return refuse("flow stage begin flow.landing-routes exited %d", code)
	}
	end := func(outcome string, extra ...string) int {
		return runStageEnd(ctx, append(append(append([]string{}, conn...), "-outcome", outcome), append(extra, *change)...), stderr)
	}
	branch, err := git("rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		_, err = git("push", "origin", branch+":"+baseBranch)
	}
	if err == nil {
		_, err = git("fetch", "origin")
	}
	if err == nil {
		_, err = git("merge-base", "--is-ancestor", "HEAD", "origin/"+baseBranch)
	}
	if err != nil {
		reason := strings.Join(strings.Fields(err.Error()), " ")
		end("failed", "-reason", reason)
		return refuse("%s", reason)
	}
	if code := end("completed"); code != 0 {
		return refuse("flow stage end flow.landing-routes exited %d after the push", code)
	}
	fmt.Fprintf(stdout, "commit: %s\nbase: %s\n", head, base)
	return 0
}
