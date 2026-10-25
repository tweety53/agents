// runverifylist.go is run-verify-list.sh ported to Go (KAN-1020).
//
// run-verify-list.sh — run one of a project's verify lists the way **Inline
// verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
// `## Report` lines, so no run hand-builds the runner (the
// self-review-backlog-flow-6 self-review pass found the parent writing one
// per run, its stdin detach patched in on a second call).
//
// Usage: run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort>]
//
// RECORDING (kan-1033 self-review: the parent hand-recorded the verify row and
// back-filled a forgotten begin with a guessed start). Given the five trailing
// arguments, the list records its own `verifier` row through `flow record
// dispatch`: `begin` (`-agent-id inline`) before its first command, and `end
// -outcome completed` after its last only when the list exited 0. No `flow`
// on PATH exits 2 before any command runs. A `begin` the store refuses prints
// `run-verify-list: flow record dispatch begin refused — no command ran: <its
// output>` on stderr and exits 4 before any command (kan-1032 self-review: the
// list ran in full, its row unrecorded, and was re-run only to get the row); a
// failed `end` prints `run-verify-list: flow record dispatch end failed: <its
// output>` on stderr and never changes the exit code.
//
// THE LIST is every non-blank, non-`#` line inside a ```bash fence of
// `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
// appends `check-spec-reach.sh <worktree>`. A `# inputs: <pathspec>...`
// line of the `test` list declares what the one command below it reads; every
// other `#` line is a comment. A line holding a `<word>`
// placeholder — `./gradlew verifyChange :<module>:test --tests '<selectors>'`
// — is a template, not a command: the shell reads `<module>` as a redirect,
// so the list is refused before anything runs (kan-983 self-review). Each command runs from
// <worktree> in a fresh `bash -c`, stdin detached (</dev/null) so none waits
// on a terminal, and the whole list always runs: a failure never stops it.
// A `test` list's commands run concurrently, each into its own output file,
// reported in list order once every one has exited (kan-920 self-review: the
// guard harnesses and `go test -race` ran one after the other, 6½ minutes; run
// together they took 307 s, each slowed by the other's CPU). A `lint` list
// runs in order, one command at a time.
// The caller exports the KEY=value lines prepare-workspace.sh printed for
// <worktree> first; every command inherits them. This script retries
// nothing — the one inline re-run of a failing command stays the caller's,
// and a race-marked command takes **The loop** whatever its re-run exits.
//
// THE GREEN SHA (KAN-887's self-review: verify re-ran the whole list minutes
// after a green full-suite run, the delta a test rename). A `test` run that
// exits 0 on a tree clean before and after it — `git status --porcelain
// --untracked-files=all` empty — writes HEAD to the worktree's own git dir
// (`git rev-parse --git-path flow-test-green`), never into the tree. A later
// `test` run on a clean tree skips every command when HEAD is that sha, and
// otherwise skips each command whose `# inputs:` pathspecs, plus
// `.flow/project.md` itself (an edited command must run), `git diff --quiet
// <sha> HEAD` finds untouched. A command declaring no inputs reads everything
// and always runs; a dirty tree, no git, no recorded sha or a failing diff
// runs the whole list. A skip passes and prints instead of output:
//   - `<command>` — skipped, no input changed since <sha> passed the list
// ponytail: the sha ignores the exported KEY=value lines; a run under other
// isolation values reuses the green, so key the file by them if they ever vary.
//
// OUTPUT, one block per command in list order, stdout and stderr together:
//   - `<command>` — exit <n>
//     <each output line, indented two spaces; past 20 lines,
//     `(truncated — last 20 of <N> lines)` and then the last 20>
//     `data race — never environmental` — when the command's whole output,
//     not only its tail, carries `WARNING: DATA RACE` or `race detected
//     during execution of test`
//     `full output: <path>` — on a non-zero exit inside a git worktree, the
//     command's whole output copied to `<git-path flow-verify-failed>/<list>-<n>.log`
//     (<n> its 0-based list position, `spec-reach` for the appended
//     check-spec-reach.sh), never into the tree; each run first removes that
//     list's earlier logs, so a printed path holds this run's output
//
// Exit codes:
//   0  every command exited 0
//   1  at least one command exited non-zero
//   2  usage (not `<worktree> lint|test` plus none or all five recording
//      arguments, no flow binary on PATH for those, or <worktree> is not a
//      directory), a list line holding a `<word>` placeholder — its stderr
//      line names it — or project-get.sh exit 2 — its stderr line says why
//   3  the list is not declared (project-get.sh exit 1): the caller
//      auto-detects, as **Inline verify** says
//   4  the store refused the recording begin (stderr names why): no command ran
//
// project-get.sh and check-spec-reach.sh are exec'd from
// FLOW_GUARD_SCRIPT_DIR, the directory the shim exports as its own. The test
// list runs on rgtRun, run-guard-tests' job pool, with one job per command so
// every one starts at once (decision verify-list-on-rgtrun).

package guard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func init() { Registry["run-verify-list"] = runVerifyList }

var (
	rvlFenceOpen = regexp.MustCompile("^```bash[[:space:]]*$")
	rvlInputs    = regexp.MustCompile(`^[[:space:]]*# inputs:(.*)$`)
	rvlTemplate  = regexp.MustCompile(`<[A-Za-z][A-Za-z0-9_-]*>`)
)

func runVerifyList(args []string, env Env, stdout, stderr io.Writer) int {
	abs := func(p string) string {
		if env.Dir != "" && !filepath.IsAbs(p) {
			return filepath.Join(env.Dir, p)
		}
		return p
	}
	const usage = "usage: run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort>]"
	if (len(args) != 2 && len(args) != 7) || (args[1] != "lint" && args[1] != "test") {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if fi, err := os.Stat(abs(args[0])); err != nil || !fi.IsDir() {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	record := func(string, ...string) bool { return true }
	if len(args) == 7 {
		flow, ok := lookPath(env, "flow")
		if !ok {
			fmt.Fprintln(stderr, "run-verify-list: no flow binary on PATH")
			return 2
		}
		wt, change := args[0], args[2]
		key, token, model, effort := args[3], args[4], args[5], args[6]
		record = func(verb string, flags ...string) bool {
			argv := []string{"record", "dispatch", verb, "-C", wt, "-change", change}
			if verb == "begin" {
				argv = append(argv, "-role", "verifier", "-model", model, "-effort", effort, "-agent-id", "inline")
			}
			argv = append(argv, "-key", key, "-session-token", token)
			argv = append(argv, flags...)
			rc, out := kwRun(env, "", flow, argv...)
			if rc == 0 {
				return true
			}
			if verb == "begin" {
				fmt.Fprintf(stderr, "run-verify-list: flow record dispatch begin refused — no command ran: %s\n", strings.TrimSpace(out))
			} else {
				fmt.Fprintf(stderr, "run-verify-list: flow record dispatch %s failed: %s\n", verb, strings.TrimSpace(out))
			}
			return false
		}
		args = args[:2]
	}
	scriptDir := env.Getenv("FLOW_GUARD_SCRIPT_DIR")
	if scriptDir == "" {
		fmt.Fprintln(stderr, "run-verify-list: FLOW_GUARD_SCRIPT_DIR is unset — run scripts/run-verify-list.sh, which sets it")
		return 2
	}
	wt, list := args[0], args[1]
	childEnv := icChildEnv(env)

	get := exec.Command(filepath.Join(scriptDir, "project-get.sh"), wt, list)
	get.Dir, get.Env, get.Stderr = env.Dir, childEnv, stderr
	body, err := get.Output()
	switch exitCode(err) {
	case 0:
	case 1:
		return 3
	default:
		return 2
	}

	// inputs[i] is the pathspec list the `# inputs:` line above cmds[i] named.
	var cmds, inputs []string
	pending, fenced := "", false
	for _, line := range strings.Split(strings.TrimRight(string(body), "\n"), "\n") {
		switch {
		case rvlFenceOpen.MatchString(line):
			fenced = true
			continue
		case strings.HasPrefix(line, "```"):
			fenced = false
			continue
		}
		fields := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' })
		if !fenced || len(fields) == 0 {
			continue
		}
		if m := rvlInputs.FindStringSubmatch(line); m != nil {
			pending = m[1]
			continue
		}
		if strings.HasPrefix(fields[0], "#") {
			continue
		}
		cmds = append(cmds, line)
		inputs = append(inputs, pending)
		pending = ""
	}
	for _, c := range cmds {
		if rvlTemplate.MatchString(c) {
			fmt.Fprintf(stderr, "run-verify-list.sh: %s line is a template, not a command — move it out of the fence: %s\n", list, c)
			return 2
		}
	}

	git := envGit(env)
	gitOut := func(args ...string) (string, bool) {
		return capture(git(append([]string{"-C", wt}, args...)...))
	}
	clean := func() bool {
		st, ok := gitOut("status", "--porcelain", "--untracked-files=all")
		return ok && st == ""
	}
	headSHA, since, greenFile := "", "", ""
	if list == "test" {
		if h, ok := gitOut("rev-parse", "-q", "--verify", "HEAD"); ok && clean() {
			if g, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-test-green"); ok {
				headSHA, greenFile = h, g
				b, _ := os.ReadFile(g)
				since = strings.TrimRight(string(b), "\n")
			}
		}
	}
	// unchanged reports whether a recorded green sha covers a command.
	unchanged := func(in string) bool {
		if since == "" {
			return false
		}
		if since == headSHA {
			return true
		}
		if strings.ReplaceAll(in, " ", "") == "" {
			return false
		}
		specs := strings.FieldsFunc(in, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' })
		diff := append([]string{"diff", "--quiet", since, headSHA, "--"}, specs...)
		return git(append([]string{"-C", wt}, append(diff, ".flow/project.md")...)...).Run() == nil
	}

	tmp := env.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	outDir, err := os.MkdirTemp(tmp, "run-verify-list.")
	if err != nil {
		fmt.Fprintf(stderr, "run-verify-list: cannot create a temporary directory: %v\n", err)
		return 2
	}
	defer os.RemoveAll(outDir)
	keep, _ := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-verify-failed")
	if keep != "" {
		os.MkdirAll(keep, 0o777)
		old, _ := filepath.Glob(filepath.Join(keep, list+"-*.log"))
		for _, f := range old {
			os.Remove(f)
		}
	}

	failed := 0
	report := func(shown string, rc int, file, key string) {
		if rc != 0 {
			failed = 1
		}
		fmt.Fprintf(stdout, "- `%s` — exit %d\n", shown, rc)
		out, _ := os.ReadFile(file)
		var lines []string
		if len(out) > 0 {
			lines = strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
		}
		if n := len(lines); n > 20 {
			fmt.Fprintf(stdout, "  (truncated — last 20 of %d lines)\n", n)
			lines = lines[n-20:]
		}
		for _, l := range lines {
			fmt.Fprintf(stdout, "  %s\n", l)
		}
		if bytes.Contains(out, []byte("WARNING: DATA RACE")) || bytes.Contains(out, []byte("race detected during execution of test")) {
			fmt.Fprintln(stdout, "  data race — never environmental")
		}
		if rc != 0 && keep != "" {
			log := filepath.Join(keep, list+"-"+key+".log")
			os.WriteFile(log, out, 0o666)
			fmt.Fprintf(stdout, "  full output: %s\n", log)
		}
	}
	// Each command runs as `(cd "$WT" && bash -c …)` did: from the worktree,
	// PWD its logical path, so `pwd` prints the path the caller named.
	cmdEnv := append(childEnv[:len(childEnv):len(childEnv)], "PWD="+abs(wt))
	// run runs one command in the foreground.
	run := func(shown, cmd, key string) {
		results := rgtRun(context.Background(), [][]string{{"bash", "-c", cmd}}, 1, outDir, abs(wt), cmdEnv)
		report(shown, results[0].rc, results[0].out, key)
	}

	if !record("begin") {
		return 4
	}
	skip := make([]bool, len(cmds))
	var argvs [][]string
	for i, c := range cmds {
		switch {
		case unchanged(inputs[i]):
			skip[i] = true
		case list == "test":
			argvs = append(argvs, []string{"bash", "-c", c})
		default:
			run(c, c, fmt.Sprint(i))
		}
	}
	if list == "test" {
		results := rgtRun(context.Background(), argvs, max(len(argvs), 1), outDir, abs(wt), cmdEnv)
		for i, c := range cmds {
			if skip[i] {
				fmt.Fprintf(stdout, "- `%s` — skipped, no input changed since %s passed the list\n", c, since)
				continue
			}
			r := results[0]
			results = results[1:]
			report(c, r.rc, r.out, fmt.Sprint(i))
		}
	}
	if headSHA != "" && failed == 0 && clean() {
		os.WriteFile(greenFile, []byte(headSHA+"\n"), 0o666)
	}
	if list == "lint" {
		run("check-spec-reach.sh "+wt, `"`+scriptDir+`/check-spec-reach.sh" "`+wt+`"`, "spec-reach")
	}
	if failed == 0 {
		record("end", "-outcome", "completed")
	}
	return failed
}
