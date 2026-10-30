// runverifylist.go is run-verify-list.sh ported to Go (KAN-1020).
//
// run-verify-list.sh — run one of a project's verify lists the way **Inline
// verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
// `## Report` lines, so no run hand-builds the runner (the
// self-review-backlog-flow-6 self-review pass found the parent writing one
// per run, its stdin detach patched in on a second call).
//
// Usage: run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
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
// appends `check-spec-reach.sh <worktree>`. A `# inputs: <pathspec>...`
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
//     `result files: <dir>` — beside it, when the run wrote JUnit `TEST-*.xml`
//     files under the worktree: each copied into a per-run
//     `<git-path flow-verify-failed>/<list>-<n>-results-<ns>/` no later run
//     removes, so a flake's first failing run stays traceable
//   - the visual verify, when it ran, last, as `- `<verify row>` — exit <n>`,
//     its kept output `test-visual.log`
//
// Exit codes:
//   0  every command exited 0
//   1  at least one command exited non-zero
//   2  usage (not `<worktree> lint|test` plus none, all five, or all six recording
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
// project-get.sh and check-spec-reach.sh are exec'd from
// FLOW_GUARD_SCRIPT_DIR, the directory the shim exports as its own. The test
// list runs on rgtRun, run-guard-tests' job pool, with one job per command so
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
	"strings"
	"syscall"
	"time"
)

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
// (verify-list-runs-targeted).
func rvlVisualVerify(env Env, wt string, covered bool) (cmd, specs, dir string, ok bool) {
	b, err := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	if err != nil {
		return "", "", "", false
	}
	cells := vvCells(string(b))
	if cells["verify"] == "" || covered {
		return "", "", "", false
	}
	var d struct {
		Visual struct {
			Verify string `json:"verify"`
		} `json:"visual"`
	}
	if raw, err := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "decision.json")); err == nil && json.Unmarshal(raw, &d) == nil && d.Visual.Verify == "skipped" {
		return "", "", "", false
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
		return "", "", "", false
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
		if branch, _ := capture(git("-C", wt, "branch", "--show-current")); branch != "" {
			if w := rvsWorktreeOn(git("-C", dir, "worktree", "list", "--porcelain"), branch); w != "" {
				dir = w
			}
		}
	}
	if s := cells["specs"]; s != "" {
		if mb == "" {
			return "", "", dir, true
		}
		return cells["verify"], strings.NewReplacer("<frontend-root>", wt, "<merge-base>", mb).Replace(s), dir, true
	}
	return strings.ReplaceAll(cells["verify"], "<specs>", ""), "", dir, true
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
	const usage = "usage: run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]"
	if (len(args) != 2 && len(args) != 7 && len(args) != 8) || (args[1] != "lint" && args[1] != "test") {
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
	// Every command runs under run-capped.sh, which kills its process tree
	// past the project's `## resource cap` (kan-1044).
	capped := filepath.Join(scriptDir, "run-capped.sh")
	// run runs one command in the foreground.
	run := func(shown, cmd, key string) {
		results := rgtRun(context.Background(), [][]string{{capped, "--", "bash", "-c", cmd}}, 1, outDir, abs(wt), cmdEnv)
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
				argvs = append(argvs, []string{flow, "suite", "record", "-C", abs(wt), "-suite", suites[i], "--", capped, "--", "bash", "-c", c})
			} else {
				argvs = append(argvs, []string{capped, "--", "bash", "-c", c})
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
	if list == "test" {
		if cmd, specs, dir, ok := rvlVisualVerify(env, abs(wt), unchanged("")); ok {
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
		run("check-spec-reach.sh "+wt, `"`+scriptDir+`/check-spec-reach.sh" "`+wt+`"`, "spec-reach")
	}
	// Lint and test share one dispatch key, and the later group's end
	// overwrites the row (skills/flow/verify-and-handoff.md, Recording): a
	// blocked lint group is carried into the test group's end, never masked.
	lintBlocked := false
	if blockedMark != "" {
		if mark, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", blockedMark); ok {
			switch {
			case list != "lint":
				_, err := os.Stat(mark)
				lintBlocked = err == nil
			case failed != 0:
				os.WriteFile(mark, nil, 0o666)
			default:
				os.Remove(mark)
			}
		}
	}
	switch {
	case failed != 0 || lintBlocked:
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
