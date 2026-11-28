package integrate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/integrate-change.sh shim (KAN-1058).
//
// flow-guard integrate-change — the mechanical half of a bare /flow at IN_PROGRESS:
// run 1 (integrate) and run 2 (cleanup), sequenced through the existing
// guards.
//
// Usage:
//   flow-guard integrate-change prepare <main-checkout> <name>
//   flow-guard integrate-change commit  <main-checkout> <name> [<worktree>=<impl subject>]...
//   flow-guard integrate-change land    <main-checkout> <name> [--route <route>] [--pr-url <worktree>=<url|none>]...
//   flow-guard integrate-change cleanup <main-checkout> <name>
//
// Go, in flow-guard: integrateChange's doc comment in
// stats/internal/guard/integrate/integratechange.go is the contract — every
// subcommand, its flags, the NEXT:/STOP:/JIRA: closing line and exit 0/1/2.

func init() {
	core.Registry["integrate-change"] = integrateChange
}

// ic is the bash's global state: its flags, the worktree set load_set
// reads, and run's OUT/RC.
type ic struct {
	env            core.Env
	prog           string // the guard name its own stderr lines open with
	stdout, stderr io.Writer
	childEnv       []string

	main, name, harness, token string
	route                      string
	squashed                   map[int][]string // by order index: the Task-Id trailers each reshape folds
	prArgs, subjects           []string

	openStage, stopOutcome string
	out                    string
	rc                     int

	wt, mb, repo, mains []string
	prURLState, jiraKey string
	canon, spec         string
	order               []int
	bases, rb           map[int]string
	base, canonBase     string
}

// integrateChange is integrate-change.sh — the mechanical half of a bare /flow at IN_PROGRESS:
// run 1 (integrate) and run 2 (cleanup), sequenced through the existing
// guards so the parent agent does only what needs judgement — the commit
// subjects, the self-review pass's dispatches and
// filing prompt, the Jira calls, and every stop that needs the operator.
//
// skills/flow-contracts/finish-contract-run1.md and finish-contract-run2.md
// are canonical for every step this script runs, and each guard's own header
// is canonical for its verdicts: this script re-implements none of them. It
// calls them in the contracts' order, relays their lines unchanged, and stops
// where a verdict needs a decision. skills/flow/integrate.md and
// skills/flow/cleanup.md say what the parent does at each stop.
//
// Usage:
//
//	integrate-change.sh prepare <main-checkout> <name>
//	integrate-change.sh commit  <main-checkout> <name> [<worktree>=<impl subject>]...
//	integrate-change.sh land    <main-checkout> <name> [--route <route>] [--pr-url <worktree>=<url|none>]...
//	integrate-change.sh cleanup <main-checkout> <name>
//
// Every subcommand also takes --harness <harness> and --session-token
// <literal token> (else EnvHarness and EnvToken: $FLOW_SESSION_TOKEN, else the calling
// agent's token, internal/caller) for its stage
// marks; a mark the store cannot take is a warning on stderr, never a stop.
//
// prepare  resolves the worktree set (the state file's `worktrees` map, else
//
//	the porcelain scan of <main-checkout>), runs
//	check-main-checkout-drift per distinct main checkout,
//	switching a checkout whose only drift is DRIFT-BRANCH back to its
//	default branch (switch_back below) and re-reading its drift; any
//	DRIFT-DIRTY, unswitched DRIFT-BRANCH or exit 2 is
//	STOP: main-checkout. Then check-finish-preflight per worktree. Every
//	worktree RUN2 -> NEXT is `cleanup`. Otherwise run 1, which first
//	stops at STOP: base-mismatch when one worktree records a flowBase
//	another resolves away from while that one's origin carries it, then
//	the unfinished-work gate (check-unfinished-work, each worktree's
//	range from `git merge-base HEAD origin/$BASE`), check-base-moved and
//	sync-onto-base on MOVED (then
//	the rebased worktree's ## lint list — rebase_lint below), and the
//	default landing route (project-get.sh). Prints per worktree its
//	reshape base and the diffstat the commit subjects are written from,
//	and records the reshape base in the worktree's own git dir
//	(`git rev-parse --git-path flow-reshape-base`) for `commit`. Its
//	NEXT names `<worktree>=<impl subject>` only for a worktree whose
//	implementation commit would not be empty (impl_paths below) and
//	is not the canonical worktree's link.md alone (link_only below).
//
// commit   the reshape per worktree (`git reset --soft <reshape base>`,
//
//	the Task-Id trailers it folds carried into the implementation
//	commit), the ledger render, commitTwo per
//	worktree in `## Merge order`, then the push of each reshaped
//	branch. A
//	worktree with nothing on the implementation side takes no
//	subject; the canonical one whose only implementation path is a
//	link.md takes `chore(spectre): merge order`; any other without
//	one stops `commit-two`.
//	Then the --force-with-lease branch push per worktree, in merge
//	order — the reshape's push (Branch backup, git-boundaries.md).
//
// land     the route (merge and push: the --force-with-lease branch push then
//
//	`push origin HEAD:<base>` per worktree in merge order, one re-sync
//	and retry after a rejection; pull request: the branch push and
//	`gh pr create --fill`, or that worktree's --pr-url when there is
//	no usable gh; manual: the branch push) and the IN_PROGRESS write.
//	It does not call `cleanup`: the self-review pass runs after land.
//
// cleanup  run 2: verify the merge (merged when gh reports the recorded PR
//
//	MERGED or `merge-base --is-ancestor` holds), `flow change save` in the
//	canonical worktree, remove-change-worktrees
//	per repository, the workspace `remove` row, the
//	proposal artifact, the FINISHED write (clearing every worktree no longer on disk) and
//	refresh-main-checkout per repository (onto origin/HEAD's branch,
//	or onto a worktree's resolved base when the main checkout has that
//	base checked out).
//
// Prints the guards' own lines, one verdict line per step of its own, and
// ends on exactly one machine-readable line:
//
//	NEXT: <what the parent does next>          exit 0
//	STOP: <kind> — <what the parent does>      exit 1
//	JIRA: transition <KEY> to <status>         before NEXT, when an issue is linked
//
// Exit 0 the subcommand ran to its end; 1 a stop — a guard verdict, a
// refusal or a failed git step needs the parent or the operator, and the
// stage mark open at that point is closed `stopped` (`not-run-2` where
// run 2's contract names it), its end carrying
// `-reason "STOP: <kind>"`; 2 usage, a guard it
// sequences missing beside it (named on stderr, nothing run), or the state
// record cannot be read.
//
// Every sibling guard below runs in this same binary (RunSibling). The
// process moves to `/` once its arguments are resolved, so it never holds a
// worktree it removes; every child runs in `/` too, save the calls the bash
// ran in a subshell `cd`, which keep their own directory.
func integrateChange(args []string, env core.Env, stdout, stderr io.Writer) int {
	c := &ic{env: env, stdout: stdout, stderr: stderr, prog: "integrate-change", stopOutcome: "stopped",
		bases: map[int]string{}, rb: map[int]string{}, squashed: map[int][]string{}}
	c.childEnv = core.ChildEnv(env)
	return core.ExitCodeOf(c.integrate(args))
}

func (c *ic) integrate(args []string) error {
	if err := c.parse(args); err != nil {
		return err
	}
	if c.env.Dir != "/" {
		// The process's own cwd: an in-process caller already runs in `/`.
		if err := os.Chdir("/"); err != nil {
			return c.die("cannot-cd", "cannot cd /: %v", err)
		}
	}
	switch args[0] {
	case "prepare":
		return c.prepare()
	case "commit":
		return c.commit()
	case "land":
		return c.land()
	case "cleanup":
		return c.cleanup()
	}
	return c.die("unknown-subcommand", "unknown subcommand '%s'", args[0])
}

func (c *ic) die(id, format string, a ...any) error {
	core.Reason(c.env, id)
	fmt.Fprintf(c.stderr, c.prog+": "+format+"\n", a...)
	return core.Exit(2)
}

func (c *ic) say(s string) { fmt.Fprintln(c.stdout, s) }

func (c *ic) parse(args []string) error {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	sub, mainArg, name := arg(0), arg(1), arg(2)
	if sub == "" || mainArg == "" || name == "" {
		return c.die("usage", "usage: flow-guard integrate-change prepare|commit|land|cleanup <main-checkout> <name> [flags]")
	}
	// A relative argument resolves against the directory it was called from.
	m, ok := icPhys(c.env.Dir, mainArg)
	if !ok {
		return c.die("not-a-directory", "%s is not a directory", mainArg)
	}
	c.main = m
	if !changeNamePlain(name) {
		return c.die("change-name-not-plain", "change name '%s' is not a plain change name", name)
	}
	c.name = name
	c.harness, c.token = core.EnvHarness(c.env), core.EnvToken(c.env)
	rest := args[3:]
	for i := 0; i < len(rest); i++ {
		next := func() string {
			if i+1 < len(rest) {
				return rest[i+1]
			}
			return ""
		}
		a := rest[i]
		switch a {
		case "--harness":
			c.harness = next()
			i++
		case "--session-token":
			c.token = next()
			i++
		case "--route":
			c.route = next()
			i++
		case "--pr-url":
			if !icPairArg(next()) {
				return c.die("bad-pr-url", "--pr-url takes <worktree>=<url|none>")
			}
			c.prArgs = append(c.prArgs, next())
			i++
		default:
			if !strings.Contains(a, "=") {
				return c.die("unknown-argument", "unknown argument '%s'", a)
			}
			c.subjects = append(c.subjects, a)
		}
	}
	if c.harness == "" || c.token == "" {
		return c.die("harness-or-token-missing", "--harness and --session-token are required (or FLOW_HARNESS or CLAUDECODE=1 / FLOW_SESSION_TOKEN or a Claude Code session)")
	}
	return nil
}

// icPairArg is the bash's `?*=?*`: an `=` with a character either side.
func icPairArg(s string) bool {
	for j := 1; j < len(s)-1; j++ {
		if s[j] == '=' {
			return true
		}
	}
	return false
}

// changeNamePlain is the bash's `case "$NAME" in *[!a-z0-9._-]*|-*|”)`.
func changeNamePlain(n string) bool {
	if n == "" || n[0] == '-' {
		return false
	}
	for _, r := range n {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// icPhys is `cd <p> && pwd -P` from dir: `..` taken logically, then the
// physical path.
func icPhys(dir, p string) (string, bool) {
	if p == "" {
		p = dir // `cd ""` stays put
	}
	if !filepath.IsAbs(p) {
		if dir == "" {
			dir = "/"
		}
		p = filepath.Join(dir, p)
	}
	r, err := filepath.EvalSymlinks(filepath.Clean(p))
	if err != nil || !core.IsDir(r) {
		return "", false
	}
	return r, true
}

// cmd is a child process: run in dir with the child environment, found
// over env's PATH, its stderr passed straight through.
func (c *ic) cmd(dir, name string, args ...string) *exec.Cmd {
	path := name
	found := true
	if !strings.Contains(name, "/") {
		path, found = core.LookPath(c.env, name)
	}
	cmd := exec.Command(path, args...)
	if !found {
		cmd.Path = name
		cmd.Err = &exec.Error{Name: name, Err: exec.ErrNotFound}
	}
	cmd.Dir = dir
	cmd.Env = c.childEnv
	cmd.Stderr = c.stderr
	return cmd
}

// rcOf is `$?`: 127 for a command never found, as bash gives it.
func rcOf(err error) int {
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist) {
		return 127
	}
	return core.ExitCode(err)
}

// capture is `"$(cmd)"`: stdout, trailing newlines stripped, and `$?`.
func (c *ic) capture(cmd *exec.Cmd) (string, int) {
	var b bytes.Buffer
	cmd.Stdout = &b
	rc := rcOf(cmd.Run())
	return strings.TrimRight(b.String(), "\n"), rc
}

// git runs git in `/`, its stdout captured and its stderr passed through.
func (c *ic) git(args ...string) (string, bool) {
	out, rc := c.capture(c.cmd("/", "git", args...))
	return out, rc == 0
}

// gitQuiet is c.git with `2>/dev/null`.
func (c *ic) gitQuiet(args ...string) (string, bool) {
	cmd := c.cmd("/", "git", args...)
	cmd.Stderr = nil
	out, rc := c.capture(cmd)
	return out, rc == 0
}

// gitPass runs git with both streams passed through, as a bare call did.
func (c *ic) gitPass(args ...string) bool {
	cmd := c.cmd("/", "git", args...)
	cmd.Stdout = c.stdout
	return cmd.Run() == nil
}

// sibling is sibling guard name's child process, run like c.cmd's.
func (c *ic) sibling(name string, args ...string) *exec.Cmd {
	cmd := core.RunSibling(c.env, name, args...)
	cmd.Dir = "/"
	cmd.Env = c.childEnv
	cmd.Stderr = c.stderr
	return cmd
}

// run is the bash's run: OUT the guard's stdout, relayed; RC its exit
// code. Its stderr goes straight through.
func (c *ic) run(guard string, args ...string) {
	c.out, c.rc = c.capture(c.sibling(guard, args...))
	if c.out != "" {
		fmt.Fprintln(c.stdout, c.out)
	}
}

// runIn is run for a guard called in process: its stdout captured and relayed, its stderr
// straight through, in `/` as a sibling runs.
func (c *ic) runIn(fn core.Func, args ...string) {
	env := c.env
	env.Dir = "/"
	var b bytes.Buffer
	c.rc = fn(args, env, &b, c.stderr)
	c.out = strings.TrimRight(b.String(), "\n")
	if c.out != "" {
		fmt.Fprintln(c.stdout, c.out)
	}
}

// verdict is `grep -m1 -E "^(<alts>):" <<<"$OUT"`.
func (c *ic) verdict(alts ...string) string {
	for _, l := range strings.Split(c.out, "\n") {
		for _, a := range alts {
			if strings.HasPrefix(l, a+":") {
				return l
			}
		}
	}
	return ""
}

func (c *ic) markBegin(stage string) {
	c.openStage = stage
	cmd := c.cmd("/", "flow", "stage", "begin", "-C", c.main, "-command", "/flow", "-stage", stage,
		"-harness", c.harness, "-session-token", c.token, c.name)
	if cmd.Run() != nil {
		fmt.Fprintf(c.stderr, "%s: warning: flow stage begin %s failed\n", c.prog, stage)
	}
}

// markEnd ends stage with outcome; a reason, when given, is the end's
// -reason (metrics.reason).
func (c *ic) markEnd(stage, outcome string, reason ...string) {
	args := []string{"stage", "end", "-C", c.main, "-command", "/flow", "-stage", stage, "-outcome", outcome}
	for _, r := range reason {
		args = append(args, "-reason", r)
	}
	cmd := c.cmd("/", "flow", append(args, c.name)...)
	if cmd.Run() != nil {
		fmt.Fprintf(c.stderr, "%s: warning: flow stage end %s failed\n", c.prog, stage)
	}
	c.openStage = ""
}

func (c *ic) landStoppedFile() string { return c.canon + "/.superpowers/sdd/land-stopped" }

// reopenLand begins flow.land again when a stop ended it, so a retried commit
// or land can end it completed; otherwise prepare's run is still open.
func (c *ic) reopenLand() {
	if f := c.landStoppedFile(); core.IsFile(f) {
		_ = os.Remove(f)
		c.markBegin("flow.land")
	}
}

func (c *ic) stop(kind, text string) error { return c.stopAs(kind, kind, text) }

// stopAs is stop for a printed STOP kind the skills name that is not itself a
// reason: id is the reason id recorded, kind the word printed after STOP:.
func (c *ic) stopAs(id, kind, text string) error {
	if c.openStage != "" {
		if c.openStage == "flow.land" && c.canon != "" {
			// A stop ends the run; the retry of commit or land reopens it (reopenLand).
			_ = os.MkdirAll(c.canon+"/.superpowers/sdd", 0o755)
			_ = os.WriteFile(c.landStoppedFile(), nil, 0o644)
		}
		c.markEnd(c.openStage, c.stopOutcome, "STOP: "+kind)
	}
	c.say("STOP: " + kind + " — " + text)
	core.Reason(c.env, id)
	return core.Exit(1)
}

func (c *ic) flags() string { return "--harness " + c.harness + " --session-token " + c.token }

func (c *ic) mainOf(w string) (string, bool) {
	d, ok := c.gitQuiet("-C", w, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if !ok {
		return "", false
	}
	return icPhys("/", d+"/..")
}

// mainOfGone: the main checkout of a worktree no longer on disk, derived
// from the layout every worktree is created in (<main>-worktrees/<name>,
// git-boundaries.md) — accepted only when it is the top of a git
// repository that is not itself a linked worktree.
func (c *ic) mainOfGone(w string) (string, bool) {
	p := core.Dirname(w)
	if !strings.HasSuffix(p, "-worktrees") || len(p) == len("-worktrees") {
		return "", false
	}
	m, ok := icPhys("/", strings.TrimSuffix(p, "-worktrees"))
	if !ok {
		return "", false
	}
	d, ok := c.gitQuiet("-C", m, "rev-parse", "--path-format=absolute", "--git-dir")
	if !ok {
		return "", false
	}
	if common, _ := c.git("-C", m, "rev-parse", "--path-format=absolute", "--git-common-dir"); d != common {
		return "", false
	}
	if top, _ := c.git("-C", m, "rev-parse", "--show-toplevel"); top != m {
		return "", false
	}
	return m, true
}

// loadSet: the worktree set — wt[i] its path, mb[i] its recorded merge base
// (- when none), repo[i] its main checkout (mainOfGone's for one already
// gone, empty when that cannot be derived), mains the distinct main
// checkouts, <main-checkout> always among them.
func (c *ic) loadSet() error {
	state, rc := c.capture(c.cmd("/", "flow", "state", "get", "-C", c.main, c.name))
	if rc != 0 {
		return c.die("flow-state-get-failed", "flow state get %s failed", c.name)
	}
	st := core.ParseState(state)
	c.prURLState, c.jiraKey = st.PrURL, st.Jira
	c.wt, c.mb, c.repo, c.mains = nil, nil, nil, []string{c.main}
	for i, k := range st.Keys {
		if k != "" {
			c.wt, c.mb = append(c.wt, k), append(c.mb, st.Values[i])
		}
	}
	if len(c.wt) == 0 {
		porcelain, _ := c.git("-C", c.main, "worktree", "list", "--porcelain")
		w := ""
		for _, l := range strings.Split(porcelain, "\n") {
			if strings.HasPrefix(l, "worktree ") {
				w = l[len("worktree "):]
			}
			if f := strings.Fields(l); len(f) > 1 && f[0] == "branch" && f[1] == "refs/heads/spectre/"+c.name && w != "" {
				c.wt, c.mb = append(c.wt, w), append(c.mb, "-")
			}
		}
	}
	for _, w := range c.wt {
		r, ok := c.mainOf(w)
		if !ok {
			r, _ = c.mainOfGone(w)
		}
		c.repo = append(c.repo, r)
		if r != "" && !icHas(c.mains, r) {
			c.mains = append(c.mains, r)
		}
	}
	return nil
}

func icHas(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// findCanonical: canon the worktree whose change directory holds tasks.md
// on disk; spec its tree.
func (c *ic) findCanonical() bool {
	c.canon, c.spec = "", ""
	for _, w := range c.wt {
		if !core.IsDir(w) {
			continue
		}
		s := core.SpecRootLeaf(w, io.Discard)
		if core.IsFile(w + "/" + s + "/changes/" + c.name + "/tasks.md") {
			c.canon, c.spec = w, s
			return true
		}
	}
	return false
}

func (c *ic) idxOf(w string) int {
	for i, x := range c.wt {
		if x == w {
			return i
		}
	}
	return -1
}

var icMergeItem = regexp.MustCompile("^[0-9]+\\. `")

// mergeOrder: order the worktree indexes in the canonical link.md's
// `## Merge order` (`.` the canonical repository, a peer through
// <spec>/peers), every worktree it does not name after them in map order.
func (c *ic) mergeOrder() {
	c.order = nil
	b, err := os.ReadFile(c.canon + "/" + c.spec + "/changes/" + c.name + "/link.md")
	if err == nil {
		on := false
		for _, l := range core.ReadLines(string(b)) {
			if strings.HasPrefix(l, "## ") {
				on = l == "## Merge order"
			}
			if !on || !icMergeItem.MatchString(l) {
				continue
			}
			item := l[strings.Index(l, "`")+1:]
			if j := strings.Index(item, "`"); j >= 0 {
				item = item[:j]
			}
			var root string
			if item == "." {
				root, _ = c.mainOf(c.canon)
			} else {
				rel := icPeer(c.canon+"/"+c.spec+"/peers", item)
				if rel == "" {
					continue
				}
				m, _ := c.mainOf(c.canon)
				var ok bool
				if root, ok = icPhys("/", m+"/"+rel); !ok {
					continue
				}
			}
			for i := range c.wt {
				if c.repo[i] == root {
					c.order = append(c.order, i)
				}
			}
		}
	}
	for i := range c.wt {
		have := false
		for _, j := range c.order {
			have = have || j == i
		}
		if !have {
			c.order = append(c.order, i)
		}
	}
}

// icPeer is `awk -v n=<item> '$1==n {print $2; exit}' <peers>`.
func icPeer(peers, item string) string {
	b, err := os.ReadFile(peers)
	if err != nil {
		return ""
	}
	for _, l := range core.ReadLines(string(b)) {
		if f := strings.Fields(l); len(f) > 0 && f[0] == item {
			if len(f) > 1 {
				return f[1]
			}
			return ""
		}
	}
	return ""
}

// baseOf: c.base the worktree's base branch, resolve-base-branch.sh's
// answer, always fetched: land pushes onto origin/<base>.
func (c *ic) baseOf(i int) error {
	var rc int
	c.base, rc = c.capture(c.sibling("resolve-base-branch", "--fetch", c.wt[i]))
	if rc == 3 {
		return c.stop("no-remote", c.wt[i]+": this repository has no remote, so there is nothing to push to or merge into")
	}
	if rc != 0 {
		return c.stopAs("base-unresolved", "base", fmt.Sprintf("%s: no base branch resolved (resolve-base-branch exit %d) — ask the operator", c.wt[i], rc))
	}
	return nil
}

// reshapeBase: a re-run's reshape base in worktree w — the newest
// `Flow-Key: flow.commit-two` commit in mb..HEAD (run 1's implementation
// commit, already reshaped), else mb.
func (c *ic) reshapeBase(w, mb string) string {
	log, _ := c.git("-C", w, "log", "--format=%H %(trailers:key=Flow-Key,valueonly,separator=%x20)", mb+"..HEAD")
	for _, l := range core.ReadLines(log) {
		h, keys, _ := strings.Cut(l, " ")
		if slices.Contains(strings.Fields(keys), "flow.commit-two") {
			return h
		}
	}
	return mb
}

func (c *ic) reshapeFile(w string) string {
	f, _ := c.git("-C", w, "rev-parse", "--path-format=absolute", "--git-path", "flow-reshape-base")
	return f
}

// implPaths: the paths commitTwo would put in the implementation commit
// after a reshape to <rev> — every change against <rev>, committed, staged,
// unstaged or untracked, outside the change's own <spec>/changes/<name>/ and
// <name>-fix-<N>/ folders, plus its own link.md: commitTwo's pathspecs
// (kan-1072 self-review, F1). Empty means that commit is skipped, so the
// worktree needs no implementation subject.
func (c *ic) implPaths(w, rev string) []string {
	p := core.SpecRootLeaf(w, c.stderr) + "/changes/"
	diff, _ := c.capture(c.cmd("/", "git", "-C", w, "diff", "--no-renames", "--name-only", rev))
	others, _ := c.capture(c.cmd("/", "git", "-C", w, "ls-files", "--others", "--exclude-standard"))
	var out []string
	for _, l := range append(core.ReadLines(diff), core.ReadLines(others)...) {
		dir, _, _ := strings.Cut(strings.TrimPrefix(l, p), "/")
		// A fix folder counts only while the worktree holds it, as in
		// commitTwo, which lists it from disk.
		planning := strings.HasPrefix(l, p) && (dir == c.name || isFixOf(dir, c.name) && core.IsDir(w+"/"+p+dir))
		if !planning || l == p+c.name+"/link.md" {
			out = append(out, l)
		}
	}
	return out
}

// linkOnly: worktree <i> is the canonical one and every path implPaths
// gives it is a link.md — the merge-order record alone, whose
// implementation commit takes the literal subject `chore(spectre): merge
// order` (git-boundaries.md, Planning commits) rather than one asked for.
func (c *ic) linkOnly(i int, rev string) bool {
	if c.wt[i] != c.canon {
		return false
	}
	p := c.implPaths(c.wt[i], rev)
	if len(p) == 0 {
		return false
	}
	for _, l := range p {
		if !strings.HasSuffix(l, "/link.md") {
			return false
		}
	}
	return true
}

// writeState: the record with its state set, updatedBy /flow, prUrl when
// pr is given, and its worktrees narrowed to keep when keep is non-nil.
func (c *ic) writeState(state string, keep []string, pr string) bool {
	j, rc := c.capture(c.cmd("/", "flow", "state", "get", "-C", c.main, c.name))
	if rc != 0 {
		return false
	}
	var rec map[string]json.RawMessage
	if json.Unmarshal([]byte(j), &rec) != nil {
		return false
	}
	set := func(k string, v any) {
		b, _ := json.Marshal(v)
		rec[k] = b
	}
	set("state", state)
	set("updatedBy", "/flow")
	if pr != "" {
		set("prUrl", pr)
	}
	if keep != nil {
		var wts map[string]json.RawMessage
		if json.Unmarshal(rec["worktrees"], &wts) != nil || wts == nil {
			return false // jq's with_entries on a non-object fails
		}
		for k := range wts {
			if !icHas(keep, k) {
				delete(wts, k)
			}
		}
		set("worktrees", wts)
	}
	b, _ := json.Marshal(rec)
	cmd := c.cmd("/", "flow", "state", "set", "-C", c.main, c.name)
	cmd.Stdin = bytes.NewReader(b)
	cmd.Stdout = c.stdout
	return cmd.Run() == nil
}

// rebaseLint: a textually clean rebase can still break the build — one
// dropped a Go import no exit code caught (KAN-900) — so the rebased tree
// runs its project's ## lint list (run-verify-list.sh) under the KEY=value
// lines prepare-workspace.sh prints for it. run-verify-list applies them
// itself, the caller's values winning; this call reads them first only to
// stop with its own reason when prepare-workspace fails. A failure stops with the list's report above the stop
// line; a project declaring no list (exit 3) has nothing to run. A passing
// list that rewrote tracked files (a formatter pre-step) stops too: the
// rewrite would otherwise ride into the reshape unreviewed or be left
// behind a push. What the list rewrote is the tracked diff's change across
// it, never the diff itself: the rebase sets the planning paths aside and
// restores them, so an uncommitted narrative edit is already there
// (kan-1010's self-review). The prefix leads the stop text — land passes
// what it already pushed.
func (c *ic) rebaseLint(w, prefix string) error {
	before, _ := c.git("-C", w, "diff", "--no-renames", "HEAD")
	out, rc := c.capture(c.sibling("prepare-workspace", w))
	if rc != 0 {
		fmt.Fprintln(c.stdout, out)
		return c.stopAs("rebase-workspace-failed", "rebase-lint", prefix+"flow-guard prepare-workspace failed on "+w+" after its rebase, so its ## lint list did not run — report its output above; the change stays IN_PROGRESS")
	}
	cmd := c.sibling("run-verify-list", w, "lint")
	cmd.Env = append([]string{}, c.childEnv...)
	for _, kv := range core.ReadLines(out) {
		// `export "$kv"` refused a name that is no identifier.
		if core.Exportable.MatchString(kv) {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	var b bytes.Buffer
	cmd.Stdout, cmd.Stderr = &b, &b
	rc = rcOf(cmd.Run())
	switch rc {
	case 0:
		if after, _ := c.git("-C", w, "diff", "--no-renames", "HEAD"); after != before {
			return c.stopAs("rebase-lint-changed-files", "rebase-lint", prefix+"the ## lint list rewrote tracked files on "+w+" after its rebase — commit or revert them; the change stays IN_PROGRESS")
		}
		c.say("REBASE-LINT-PASSED: " + w)
	case 3:
		c.say("REBASE-LINT-NONE: " + w + " — its project declares no ## lint list")
	default:
		fmt.Fprintln(c.stdout, strings.TrimRight(b.String(), "\n"))
		return c.stopAs("rebase-lint-failed", "rebase-lint", fmt.Sprintf("%sthe ## lint list failed on %s after its rebase (run-verify-list exit %d) — report its output above; the change stays IN_PROGRESS", prefix, w, rc))
	}
	return nil
}

func (c *ic) pushBranch(w string) error {
	if !c.gitPass("-C", w, "push", "--force-with-lease", "-u", "origin", "spectre/"+c.name) {
		return c.stopAs("push-rejected", "push", "the branch push from "+w+" was rejected — report git's output; the change stays IN_PROGRESS")
	}
	return nil
}

// isFixOf is whether dir is a fix sub-change of name: <name>-fix-<digits>.
func isFixOf(dir, name string) bool {
	n, ok := strings.CutPrefix(dir, name+"-fix-")
	return ok && n != "" && strings.Trim(n, "0123456789") == ""
}
