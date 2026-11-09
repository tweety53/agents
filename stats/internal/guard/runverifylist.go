// runverifylist.go is run-verify-list.sh ported to Go (KAN-1020).
//
// run-verify-list.sh — run one of a project's verify lists the way **Inline
// verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
// `## Report` lines, so no run hand-builds the runner (the
// self-review-backlog-flow-6 self-review pass found the parent writing one
// per run, its stdin detach patched in on a second call).
//
// Usage: run-verify-list.sh <worktree> lint|test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
//
// RECORDING (kan-1033 self-review: the parent hand-recorded the verify row and
// back-filled a forgotten begin with a guessed start). Given the five trailing
// arguments, the list records its own `verifier` row through `flow record
// dispatch`: `begin` (`-agent-id inline`, `-diff-base` the HEAD of
// <worktree>, which check-verify-green judges the row by) before its first command, and `end`
// after its last: `-outcome completed` when the list exited 0, `-outcome
// blocked -cause test-failure` when a command failed, `-outcome skipped` when
// every command of a `test` list was skipped. No `flow`
// on PATH exits 2 before any command runs. The optional sixth recording
// argument, <canonical-worktree>, is the `-C` of both calls — the worktree
// holding spectre/changes/<change> when <worktree> is another checkout of a
// multi-repo change (kan-1014 self-review); it defaults to <worktree>. A `begin` the store refuses prints
// `run-verify-list: flow record dispatch begin refused — no command ran: <its
// output>` on stderr and exits 4 before any command (kan-1032 self-review: the
// list ran in full, its row unrecorded, and was re-run only to get the row); a
// failed `end` prints `run-verify-list: flow record dispatch end failed: <its
// output>` on stderr and never changes the exit code.
//
// THE LIST is every non-blank, non-`#` line inside a ```bash fence of
// `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
// appends `flow-guard check-spec-reach <worktree>`. A `# inputs: <pathspec>...`
// line of the `test` list declares what the one command below it reads, and a
// `# suite: <name>` line names the one command below it as that suite: it runs
// under `flow suite record -C <worktree> -suite <name> --`, so its runtime
// reaches the store `flow suite list` reads, and runs bare when no `flow` is
// on PATH; every other `#` line is a comment. A line holding a `<word>`
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
// A `test` list then ends with the `## visual verification` `verify` row,
// run once in the foreground after every other command has exited, `<specs>`
// replaced by nothing, from the section's `regression checkout` when declared
// and <worktree> otherwise (verify-list-runs-visual, kan-1047: the run never
// ran the visual verify by hand) -- when the row is declared,
// `<worktree>/.superpowers/sdd/decision.json` does not record
// `visual.verify` `skipped`, HEAD is not the green sha, and `git diff
// --name-only <merge-base>..HEAD` touches a `ui paths` glob, the merge base
// that of HEAD and resolve-base-branch's base (its `origin/` ref when
// fetched). A base that does not resolve runs it. It reports and fails the
// list like any other command; a `lint` list never runs it.
// A `visual` list runs that `verify` row alone, as a `test` list's tail
// would, and never the `## lint` or `## test` commands; when the conditions
// above skip it, it prints `- visual verify — not run: …` and exits 0
// (kan-1052 self-review, F2: /flow-fast and /flow-interactive had no command
// that ran the row). Unlike a `test` list's tail, it never falls back to a
// `regression checkout`'s main checkout: one holding no worktree on
// <worktree>'s branch prints `- visual verify — not run: …` naming it and
// exits 1 (kan-1052 self-review, R1: the main checkout's build is not the
// change's).
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
// THE LINT STAMP (kan-1081 self-review, F1: a request ended landed with no
// evidence its lint list ran green). A `lint` run — the appended
// check-spec-reach included — that exits 0 on a tree clean before and after
// it writes HEAD plus a newline to the worktree's own git dir (`git rev-parse
// --git-path flow-lint-green`), never into the tree; a failing or dirty run
// leaves an earlier stamp as it was. Nothing is skipped on it: `flow record
// request end -outcome landed` and `flow interactive land` read it, refusing
// a commit the stamp does not hold.
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
//     `result files: <dir>` — beside it, when the run wrote JUnit `TEST-*.xml`
//     files under the worktree: each copied into a per-run
//     `<git-path flow-verify-failed>/<list>-<n>-results-<ns>/` no later run
//     removes, so a flake's first failing run stays traceable
//   - the visual verify, when it ran, last, as `- `<verify row>` — exit <n>`,
//     its kept output `test-visual.log`
//
// Exit codes:
//   0  every command exited 0
//   1  at least one command exited non-zero, or a `visual` list refused a
//      `regression checkout` holding no worktree on the branch
//   2  usage (not `<worktree> lint|test|visual` plus none, all five, or all six recording
//      arguments, no flow binary on PATH for those, or <worktree> is not a
//      directory), a list line holding a `<word>` placeholder — its stderr
//      line names it — or project-get.sh exit 2 — its stderr line says why,
//      or a recording <key> that does not fold, by check-verify-green's
//      vgFixSuffix, to `verify` or `verify-<s>`, <s> the worktree's basename
//      or its repository's name (kan-1067 self-review, 1-C: `verify-fix-1b`
//      formed its own group and verify stayed NOT-GREEN)
//   3  the list is not declared (project-get.sh exit 1): the caller
//      auto-detects, as **Inline verify** says
//   4  the store refused the recording begin (stderr names why): no command ran
//
// project-get, check-spec-reach and run-capped run in this same binary
// (runSibling, selfExe). The test
// list runs on rgtRun, a job pool, with one job per command so
// every one starts at once (decision verify-list-on-rgtrun).

package guard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// The contract below moved here from the header of the deleted
// scripts/run-verify-list.sh shim (KAN-1058).
//
// flow-guard run-verify-list — run one of a project's verify lists the way **Inline
// verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
// `## Report` lines.
//
// Usage: flow-guard run-verify-list <worktree> lint|test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
//
// Exit codes:
//   0  every command exited 0
//   1  at least one command exited non-zero, or a `visual` list refused a
//      `regression checkout` holding no worktree on the branch
//   2  usage (not `<worktree> lint|test|visual` plus none, all five, or all six recording
//      arguments, no flow binary on PATH for those, or <worktree> is not a
//      directory), a list line holding a `<word>` placeholder — its stderr
//      line names it — or flow-guard project-get exit 2 — its stderr line says why
//   3  the list is not declared (flow-guard project-get exit 1): the caller
//      auto-detects, as **Inline verify** says
//   4  the store refused the recording begin (stderr names why): no command ran
//
// With the recording arguments the list ends the row it began with one of
// three outcomes: blocked with cause test-failure when a command failed,
// skipped when every command of a test list was skipped, and completed
// otherwise — a lint list always runs flow-guard check-spec-reach, so it never ends
// skipped. The begin carries `-diff-base <sha>`, the HEAD of <worktree> (the
// verified worktree, never <canonical-worktree>), which flow-guard check-verify-green
// judges the row by; a <worktree> git cannot resolve a HEAD in records none,
// and that row is never green (decision verify-row-records-head).
//
// A `# suite: <name>` line in the `test` list names the one command below it
// as that suite: it runs under `flow suite record -C <worktree> -suite <name>
// --`, so its runtime reaches the store `flow suite list` reads; with no flow
// on PATH it runs bare.
//
// A `test` list ends with the `## visual verification` `verify` row, run once
// after every other command, when the worktree's diff from its base branch
// touches a `ui paths` glob and decision.json does not record `visual.verify`
// `skipped`; a base that does not resolve runs it, and it fails the list like
// any other command (kan-1047). A declared `specs` row runs first, its paths
// put in `<specs>`: no path runs no spec, a non-zero exit or an unresolved
// base fails the list, and the whole suite never runs in its place. With no
// `specs` row, `<specs>` is emptied and the suite runs whole.
//
// A `visual` list runs that `verify` row alone, as a `test` list's tail would,
// and never the `## lint` or `## test` commands; when the conditions above
// skip it, it prints `- visual verify — not run: …` and exits 0 (kan-1052
// self-review, F2).
//
// The mechanism, and its full contract — the list, the green sha, the output
// — is stats/internal/guard/runverifylist.go; its test is
// stats/internal/guard/run_verify_list_test.go.

// keepResults copies every JUnit result file (`TEST-*.xml`) under wt written
// since the run started into dir, at its path relative to wt, and returns dir
// -- or "" when there was none. A failing suite's result files are what a
// flake is traced from, and the next run overwrites them (kan-1017's
// self-review: an 8-test failure left nothing to trace). dir is per run and
// never removed by a later one. .git and node_modules are not walked.
func keepResults(wt, dir string, since time.Time) string {
	kept := ""
	filepath.WalkDir(wt, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); n == ".git" || n == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasPrefix(d.Name(), "TEST-") || !strings.HasSuffix(d.Name(), ".xml") {
			return nil
		}
		if fi, err := d.Info(); err != nil || fi.ModTime().Before(since) {
			return nil
		}
		rel, _ := filepath.Rel(wt, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		dst := filepath.Join(dir, rel)
		if os.MkdirAll(filepath.Dir(dst), 0o777) == nil && os.WriteFile(dst, b, 0o666) == nil {
			kept = dir
		}
		return nil
	})
	return kept
}

func init() { Registry["run-verify-list"] = runVerifyList }

var (
	rvlFenceOpen = regexp.MustCompile("^```bash[[:space:]]*$")
	rvlInputs    = regexp.MustCompile(`^[[:space:]]*# inputs:(.*)$`)
	rvlSuite     = regexp.MustCompile(`^[[:space:]]*# suite:[[:space:]]*([A-Za-z0-9._-]+)[[:space:]]*$`)
	rvlTemplate  = regexp.MustCompile(`<[A-Za-z][A-Za-z0-9_-]*>`)
)

// rvlVisualVerify is the `## visual verification` `verify` row a `test` list
// ends with, the directory it runs from, and -- when the section declares a
// `specs` row -- that row with `<frontend-root>` and `<merge-base>`
// substituted, whose printed paths the caller puts in `<specs>`; with no
// `specs` row, `<specs>` is emptied and the row runs whole. ok false when the
// row is not declared, decision.json records `visual.verify` `skipped`, the
// green sha covers the tree, or the diff from the base branch touches no
// `ui paths` glob (verify-list-runs-visual). A base that does not resolve
// runs it: running more is the safe side -- but a `specs` row then has no
// merge base to scope to, and specs comes back "" with ok true and cmd ""
// so the caller fails the list rather than run the whole suite
// (verify-list-runs-targeted). fallback is the `regression checkout` dir
// fell back to when that checkout holds no worktree on wt's branch, "" otherwise.
func rvlVisualVerify(env Env, wt string, covered bool) (cmd, specs, dir, fallback string, ok bool) {
	b, err := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	if err != nil {
		return "", "", "", "", false
	}
	cells := vvCells(string(b))
	if cells["verify"] == "" || covered {
		return "", "", "", "", false
	}
	var d struct {
		Visual struct {
			Verify string `json:"verify"`
		} `json:"visual"`
	}
	if raw, err := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "decision.json")); err == nil && json.Unmarshal(raw, &d) == nil && d.Visual.Verify == "skipped" {
		return "", "", "", "", false
	}
	mb := ""
	changed := func() ([]string, error) {
		var base bytes.Buffer
		if Registry["resolve-base-branch"]([]string{wt}, env, &base, io.Discard) != 0 {
			return nil, fmt.Errorf("no base branch")
		}
		git := envGit(env)
		ref := resolveRemoteBase(git, wt, strings.TrimSpace(base.String()))
		m, ok := capture(git("-C", wt, "merge-base", "HEAD", ref))
		if !ok {
			return nil, fmt.Errorf("no merge base")
		}
		mb = strings.TrimSpace(m)
		names, ok := capture(git("-C", wt, "diff", "--no-renames", "--name-only", mb+"..HEAD"))
		if !ok {
			return nil, fmt.Errorf("no diff")
		}
		return strings.Split(names, "\n"), nil
	}
	if visualTrigger(env, wt, changed, io.Discard, io.Discard) == 1 {
		return "", "", "", "", false
	}
	dir = wt
	if c := cells["regression checkout"]; c != "" {
		dir = c
		if !filepath.IsAbs(c) {
			dir = filepath.Join(wt, c)
		}
		// The cell names a main checkout; the change's worktree of it, on
		// wt's branch, is where the change's code and captures live.
		git := envGit(env)
		fallback = dir
		if branch, _ := capture(git("-C", wt, "branch", "--show-current")); branch != "" {
			if w := rvsWorktreeOn(git("-C", dir, "worktree", "list", "--porcelain"), branch); w != "" {
				dir, fallback = w, ""
			}
		}
	}
	if s := cells["specs"]; s != "" {
		if mb == "" {
			return "", "", dir, fallback, true
		}
		return cells["verify"], strings.NewReplacer("<frontend-root>", wt, "<merge-base>", mb).Replace(s), dir, fallback, true
	}
	return strings.ReplaceAll(cells["verify"], "<specs>", ""), "", dir, fallback, true
}

func runVerifyList(args []string, env Env, stdout, stderr io.Writer) int {
	// A closed reader (`| head`) must not kill the list before its row's
	// `end`: writes return EPIPE instead (decision f6-ignore-sigpipe, kan-1037).
	// Notify, not Ignore: an ignored SIGPIPE is inherited by every list
	// command, where Go resets a handled one to the default in each child.
	sigpipe := make(chan os.Signal, 1)
	signal.Notify(sigpipe, syscall.SIGPIPE)
	defer signal.Stop(sigpipe)
	abs := func(p string) string {
		if env.Dir != "" && !filepath.IsAbs(p) {
			return filepath.Join(env.Dir, p)
		}
		return p
	}
	const usage = "usage: flow-guard run-verify-list <worktree> lint|test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]"
	if (len(args) != 2 && len(args) != 7 && len(args) != 8) || (args[1] != "lint" && args[1] != "test" && args[1] != "visual") {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if fi, err := os.Stat(abs(args[0])); err != nil || !fi.IsDir() {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	record := func(string, ...string) bool { return true }
	blockedMark := ""
	if len(args) >= 7 {
		flow, ok := lookPath(env, "flow")
		if !ok {
			fmt.Fprintln(stderr, "run-verify-list: no flow binary on PATH")
			return 2
		}
		wt, change, verified := args[0], args[2], args[0]
		if len(args) == 8 {
			wt = args[7]
		}
		key, token, model, effort := args[3], args[4], args[5], args[6]
		if !verifyKeyOK(env, abs(args[0]), key) {
			fmt.Fprintf(stderr, "run-verify-list: key %q does not fold to verify or verify-<repository name> (skills/flow/verify-and-handoff.md, **Recording.**) — no command ran\n", key)
			return 2
		}
		blockedMark = "flow-verify-blocked-" + key + "-" + token
		record = func(verb string, flags ...string) bool {
			argv := []string{"record", "dispatch", verb, "-C", wt, "-change", change}
			if verb == "begin" {
				argv = append(argv, "-role", "verifier", "-model", model, "-effort", effort, "-agent-id", "inline")
				// The verified worktree's HEAD, which check-verify-green judges
				// the row by (decision verify-row-records-head); a worktree git
				// cannot resolve records none, and that row is never green.
				if head, ok := capture(envGit(env)("-C", abs(verified), "rev-parse", "HEAD")); ok && head != "" {
					flags = append(flags, "-diff-base", head)
				}
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
	wt, list := args[0], args[1]
	childEnv := icChildEnv(env)

	// A visual list runs the `verify` row alone: no list to read.
	var body []byte
	if list != "visual" {
		get := runSibling(env, "project-get", wt, list)
		get.Env, get.Stderr = childEnv, stderr
		var err error
		body, err = get.Output()
		switch exitCode(err) {
		case 0:
		case 1:
			return 3
		default:
			return 2
		}
	}

	// inputs[i] is the pathspec list the `# inputs:` line above cmds[i] named,
	// suites[i] the `# suite:` line's name.
	var cmds, inputs, suites []string
	pending, suite, fenced := "", "", false
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
		if m := rvlSuite.FindStringSubmatch(line); m != nil {
			suite = m[1]
			continue
		}
		if strings.HasPrefix(fields[0], "#") {
			continue
		}
		cmds = append(cmds, line)
		inputs = append(inputs, pending)
		suites = append(suites, suite)
		pending, suite = "", ""
	}
	for _, c := range cmds {
		if rvlTemplate.MatchString(c) {
			fmt.Fprintf(stderr, "run-verify-list: %s line is a template, not a command — move it out of the fence: %s\n", list, c)
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
	lintSHA, lintFile := "", ""
	if list == "lint" {
		if h, ok := gitOut("rev-parse", "-q", "--verify", "HEAD"); ok && clean() {
			if g, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-lint-green"); ok {
				lintSHA, lintFile = h, g
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

	started := time.Now().Add(-time.Second)
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
			if dir := keepResults(abs(wt), filepath.Join(keep, fmt.Sprintf("%s-%s-results-%d", list, key, time.Now().UnixNano())), started); dir != "" {
				fmt.Fprintf(stdout, "  result files: %s\n", dir)
			}
		}
	}
	// Each command runs as `(cd "$WT" && bash -c …)` did: from the worktree,
	// PWD its logical path, so `pwd` prints the path the caller named.
	cmdEnv := append(childEnv[:len(childEnv):len(childEnv)], "PWD="+abs(wt))
	// Every command runs under run-capped, which kills its process tree
	// past the project's `## resource cap` (kan-1044).
	self, err := selfExe()
	if err != nil {
		fmt.Fprintf(stderr, "run-verify-list: cannot resolve flow-guard's own executable: %v\n", err)
		return 2
	}
	capped := []string{self, "run-capped", "--", "bash", "-c"}
	// run runs one command in the foreground.
	run := func(shown, cmd, key string) {
		results := rgtRun(context.Background(), [][]string{append(capped[:len(capped):len(capped)], cmd)}, 1, outDir, abs(wt), cmdEnv)
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
			if flow, ok := lookPath(env, "flow"); ok && suites[i] != "" {
				argvs = append(argvs, append([]string{flow, "suite", "record", "-C", abs(wt), "-suite", suites[i], "--"}, append(capped[:len(capped):len(capped)], c)...))
			} else {
				argvs = append(argvs, append(capped[:len(capped):len(capped)], c))
			}
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
	visualRan := false
	if list != "lint" {
		if cmd, specs, dir, fallback, ok := rvlVisualVerify(env, abs(wt), unchanged("")); !ok && list == "visual" {
			fmt.Fprintln(stdout, "- visual verify — not run: no `verify` row, `visual.verify` skipped, or no `ui paths` change")
		} else if ok && list == "visual" && fallback != "" {
			// /flow-fast and /flow-interactive set up no regression-checkout
			// worktree; the main checkout holds a build that is not the
			// change's (kan-1052 self-review, R1).
			failed = 1
			fmt.Fprintf(stdout, "- visual verify — not run: `regression checkout` %s holds no worktree on this branch; create it, and start the stack the `verify` row needs from it, then re-run\n", fallback)
		} else if ok {
			visualRan = true
			// A `specs` row scopes the suite to the specs it maps; one that
			// cannot run, or fails, blocks the list, and one that maps
			// nothing runs no spec -- never the whole suite, which runs only
			// when the operator asks (verify-list-runs-targeted).
			switch {
			case cmd == "":
				failed = 1
				fmt.Fprintln(stdout, "- `specs` — not run, no merge base to scope the visual verify to; the whole suite runs only when the operator asks")
			case specs != "":
				r := rgtRun(context.Background(), [][]string{{"bash", "-c", specs}}, 1, outDir, dir, cmdEnv)[0]
				out, _ := os.ReadFile(r.out)
				paths := strings.Fields(string(out))
				switch {
				case r.rc != 0:
					report(specs, r.rc, r.out, "specs")
					cmd = ""
				case len(paths) == 0:
					fmt.Fprintf(stdout, "- `%s` — exit 0, mapped no spec: the visual verify ran no spec\n", specs)
					cmd = ""
				default:
					cmd = strings.ReplaceAll(cmd, "<specs>", strings.Join(paths, " "))
				}
			}
			if cmd != "" {
				results := rgtRun(context.Background(), [][]string{{"bash", "-c", cmd}}, 1, outDir, dir, cmdEnv)
				report(cmd, results[0].rc, results[0].out, "visual")
			}
		}
	}
	if headSHA != "" && failed == 0 && clean() {
		os.WriteFile(greenFile, []byte(headSHA+"\n"), 0o666)
	}
	if list == "lint" {
		self, _ := selfExe()
		run("flow-guard check-spec-reach "+wt, `"`+self+`" check-spec-reach "`+wt+`"`, "spec-reach")
		if lintSHA != "" && failed == 0 && clean() {
			os.WriteFile(lintFile, []byte(lintSHA+"\n"), 0o666)
		}
	}
	// Lint and test share one dispatch key, run at once, and the later
	// group's end overwrites the row (skills/flow/verify-and-handoff.md,
	// Recording): each group marks its own failure, clears its own mark when
	// green, and ends blocked while any group's mark stands. The ends run
	// under one lock, held through the record, so whichever group ends last
	// sees the other's mark — a blocked group is never masked.
	anyBlocked := false
	if blockedMark != "" {
		if mark, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", blockedMark); ok {
			if lock, err := os.OpenFile(mark+".lock", os.O_CREATE|os.O_RDWR, 0o666); err == nil {
				defer lock.Close()
				syscall.Flock(int(lock.Fd()), syscall.LOCK_EX)
			}
			if failed != 0 {
				os.WriteFile(mark+"-"+list, nil, 0o666)
			} else {
				os.Remove(mark + "-" + list)
			}
			for _, l := range []string{"lint", "test", "visual"} {
				if _, err := os.Stat(mark + "-" + l); err == nil {
					anyBlocked = true
				}
			}
		}
	}
	switch {
	case failed != 0 || anyBlocked:
		record("end", "-outcome", "blocked", "-cause", "test-failure")
	case list != "lint" && !visualRan && len(cmds) > 0 && !slices.Contains(skip, false):
		// A lint list always ran check-spec-reach.sh, so it never ends skipped.
		record("end", "-outcome", "skipped")
	default:
		record("end", "-outcome", "completed")
	}
	return failed
}

// verifyKeyOK: key folds, by check-verify-green's vgFixSuffix, to `verify` or
// `verify-<s>`, <s> wt's basename or its repository's name (the directory
// holding its git common dir); a git failure admits only the basename.
func verifyKeyOK(env Env, wt, key string) bool {
	g := vgFixSuffix.ReplaceAllString(key, "verify$2")
	if g == "verify" || g == "verify-"+filepath.Base(wt) {
		return true
	}
	out, ok := capture(envGit(env)("-C", wt, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	return ok && g == "verify-"+filepath.Base(filepath.Dir(out))
}

// rgtJob is one harness's run: its exit code, its wall time and the file
// its merged stdout+stderr was captured into.
type rgtJob struct {
	rc, elapsed int
	out         string
}

// rgtRun runs each argv in dir — `bash
// -c <command>` for run-verify-list's test list — at most jobs at a time,
// starting them in list order, stdin /dev/null. Results are indexed by list position,
// so every later walk is in list order by construction. stdout and stderr
// share one capture file -- one descriptor, so the harness's own write order
// is kept -- and a file rather than a pipe, so a background process a
// harness leaves holding its output never delays the run.
func rgtRun(ctx context.Context, argvs [][]string, jobs int, tmpDir, dir string, childEnv []string) []rgtJob {
	results := make([]rgtJob, len(argvs))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	for i, argv := range argvs {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() { <-sem; wg.Done() }()
			r := &results[i]
			r.out = filepath.Join(tmpDir, strconv.Itoa(i)+".out")
			f, err := os.Create(r.out)
			if err != nil {
				r.rc = 126
				return
			}
			defer f.Close()
			cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
			cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, childEnv, f, f
			// Per-harness wall time is whole Unix seconds end minus start,
			// what the script's `$(date +%s)` arithmetic computed -- never
			// a truncated time.Since, which differs whenever a run
			// straddles a second boundary.
			s := time.Now().Unix()
			r.rc = exitCode(cmd.Run())
			r.elapsed = int(time.Now().Unix() - s)
		}()
	}
	wg.Wait()
	return results
}
