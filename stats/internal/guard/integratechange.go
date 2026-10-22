package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	Registry["integrate-change"] = integrateChange
}

// icExit unwinds a subcommand to integrateChange with its exit code: the
// bash's `exit` from inside any helper.
type icExit int

// ic is the bash's global state: its flags, the worktree set load_set
// reads, and run's OUT/RC.
type ic struct {
	env            Env
	stdout, stderr io.Writer
	childEnv       []string

	scriptDir, main, name, harness, token string
	acceptForeign, acceptOutstanding      bool
	planMsg, route                        string
	prArgs, proceed, subjects             []string

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
// subjects, the narrative paragraph, the self-review pass's dispatches and
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
//	integrate-change.sh prepare <main-checkout> <name> [--accept-foreign] [--accept-outstanding]
//	integrate-change.sh commit  <main-checkout> <name> [<worktree>=<impl subject>]... [--outstanding <text>]
//	integrate-change.sh land    <main-checkout> <name> [--route <route>] [--pr-url <worktree>=<url|none>]...
//	integrate-change.sh cleanup <main-checkout> <name> [--proceed <main-checkout>]...
//
// Every subcommand also takes --harness <harness> and --session-token
// <literal token> (else $FLOW_HARNESS and $FLOW_SESSION_TOKEN) for its stage
// marks; a mark the store cannot take is a warning on stderr, never a stop.
//
// prepare  resolves the worktree set (the state file's `worktrees` map, else
//
//	the porcelain scan of <main-checkout>), runs check-foreign-staged
//	and check-main-checkout-drift per distinct main checkout,
//	switching a checkout whose only drift is DRIFT-BRANCH back to its
//	default branch (switch_back below) and re-reading its drift,
//	migrate-worktrees, then check-finish-preflight per worktree. Every
//	worktree RUN2 -> NEXT is `cleanup`. Otherwise run 1, which first
//	stops at STOP: base-mismatch when one worktree records a flowBase
//	another resolves away from while that one's origin carries it: the
//	uncommitted-archive undo, the archived decision, the unfinished-work
//	gate (check-unfinished-work, check-visual-verify-dispatched against
//	`git merge-base HEAD origin/$BASE`, check-untasked-commits from that
//	base or an archived re-run's), check-base-moved and
//	sync-onto-base on MOVED (running each GUARD-TEST it names, then
//	the rebased worktree's ## lint list — rebase_lint below), and the
//	default landing route (project-get.sh). Prints per worktree its
//	reshape base and the diffstat the commit subjects are written from,
//	and records the reshape base in the worktree's own git dir
//	(`git rev-parse --git-path flow-reshape-base`) for `commit`. Its
//	NEXT names `<worktree>=<impl subject>` only for a worktree whose
//	implementation commit would not be empty (impl_paths below) and
//	is not the canonical worktree's link.md alone (link_only below).
//	--accept-foreign / --accept-outstanding: the operator chose
//	Continue at that stop; its lines still print. A
//	VISUAL-VERIFY-MISSING is STOP: visual-verify, which
//	--accept-outstanding never passes — only an operator's
//	false-positive flag on that worktree's verdict row naming its
//	current HEAD. The gate runs on an archived re-run with no new
//	work too, and again on each worktree a base move rebased here.
//	commit and land run the same gate before every push they make,
//	on every route, and again after land's re-sync.
//
// commit   the pre-reshape pin per worktree (refs/flow/pre-reshape/<name>
//
//	at its HEAD; kept when it already reaches HEAD, so a re-run
//	after a stop keeps the first, else advanced to a merge of the
//	old pin and HEAD; fetched from origin when absent locally; run 2
//	never deletes it), reshape-branch per
//	worktree, the ledger render, commit-split per
//	worktree in `## Merge order`, `spectre archive` (each
//	<name>-fix-N too) unless already archived, commit-archive,
//	then in every other worktree whose <spec>/changes/<name>/ still
//	exists `spectre archive <name>` and commit-archive. A
//	worktree with nothing on the implementation side takes no
//	subject; the canonical one whose only implementation path is a
//	link.md takes `chore(spectre): merge order`; any other without
//	one stops `commit-split`. An
//	archived change without new work skips straight to commit-archive.
//	Before any of it, a narrative lacking the section of a fix or
//	integrate run the store marked stops `narrative`
//	(check-narrative-runs.sh).
//	Then, each behind the visual-verify gate, the --force-with-lease
//	branch push per worktree, in merge
//	order — the reshape's push (Branch backup, git-boundaries.md) —
//	each followed by a plain push of that worktree's pin, refused
//	(STOP: pre-reshape) when origin holds one the local pin does
//	not descend from.
//	--outstanding <text> puts `Outstanding: <text>` in the planning
//	commit's body, under this script's own subject.
//	The parent appends its narrative before calling this; an
//	`In-run pipeline fix:` line in it off the format pipeline.md's
//	Pipeline defects found mid-run fixes stops first, STOP: narrative.
//
// land     the route (merge and push: the --force-with-lease branch push then
//
//	`push origin HEAD:<base>` per worktree in merge order, one re-sync
//	and retry after a rejection; pull request: the branch push and
//	`gh pr create --fill`, or that worktree's --pr-url when there is
//	no usable gh; manual: the branch push), the self-review report's
//	sha re-read before each landing push and its `fixed` rows after,
//	and the IN_PROGRESS write. Every push, on every route (the
//	canonical re-push after a report rewrite too), runs the
//	visual-verify gate on that worktree's HEAD first. Refuses to push anything until the
//	self-review report is committed in the canonical worktree; only
//	`commit`'s branch push precedes it. Merge
//	and push continues into `cleanup`.
//
// cleanup  run 2: verify the merge (merged when gh reports the recorded PR
//
//	MERGED or `merge-base --is-ancestor` holds), remove-change-worktrees
//	per repository (--proceed passed only to each repository named by
//	one), the workspace `remove` row, the
//	proposal artifact, check-cleanup-complete per repository, the
//	FINISHED write (clearing every worktree no longer on disk) and
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
// stage mark open at that point is closed `stopped` (`not-run-2` and
// `leftover` where run 2's contract names those); 2 usage, a guard it
// sequences missing beside it (named on stderr, nothing run), or the state
// record cannot be read.
//
// Ported to Go (KAN-1019): scripts/integrate-change.sh is now the shim that
// execs this guard, and every sibling below is exec'd from
// FLOW_GUARD_SCRIPT_DIR, the directory the shim exports as its own. Every
// child runs in `/` — the bash's `cd /` — save the calls the bash ran in a
// subshell `cd`, which keep their own directory.
func integrateChange(args []string, env Env, stdout, stderr io.Writer) (code int) {
	c := &ic{env: env, stdout: stdout, stderr: stderr, stopOutcome: "stopped",
		bases: map[int]string{}, rb: map[int]string{}}
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(icExit)
			if !ok {
				panic(r)
			}
			code = int(e)
		}
	}()
	c.scriptDir = env.Getenv("FLOW_GUARD_SCRIPT_DIR")
	if c.scriptDir == "" {
		fmt.Fprintln(stderr, "integrate-change: FLOW_GUARD_SCRIPT_DIR is unset — run scripts/integrate-change.sh, which sets it")
		return 2
	}
	c.childEnv = icChildEnv(env)
	c.parse(args)
	switch args[0] {
	case "prepare":
		c.prepare()
	case "commit":
		c.commit()
	case "land":
		c.land()
	case "cleanup":
		c.cleanup()
	default:
		c.die("unknown subcommand '%s'", args[0])
	}
	return 0
}

// icChildEnv is what every child runs under: the process environment with
// each variable the harness's `sut` set taken from env instead, so a test's
// stubs and identity reach the guards as the bash's exported ones did.
func icChildEnv(env Env) []string {
	names := []string{"PATH", "HOME", "TMPDIR", "FLOW_GUARD_CACHE_DIR", "FLOW_GUARD_TELEMETRY", "FLOW_GUARD_SCRIPT_DIR",
		"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL",
		"GIT_AUTHOR_DATE", "GIT_COMMITTER_DATE"}
	lookup := env.LookupEnv
	if lookup == nil {
		lookup = func(n string) (string, bool) { v := env.Getenv(n); return v, v != "" }
	}
	var out []string
	for _, kv := range os.Environ() {
		n, _, _ := strings.Cut(kv, "=")
		keep := true
		for _, m := range names {
			keep = keep && n != m
		}
		if keep {
			out = append(out, kv)
		}
	}
	for _, n := range names {
		if v, ok := lookup(n); ok {
			out = append(out, n+"="+v)
		}
	}
	return out
}

func (c *ic) die(format string, a ...any) {
	fmt.Fprintf(c.stderr, "integrate-change: "+format+"\n", a...)
	panic(icExit(2))
}

func (c *ic) say(s string) { fmt.Fprintln(c.stdout, s) }

func (c *ic) parse(args []string) {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	// The bash's presence loop now lives in the shim, ahead of this exec.
	sub, mainArg, name := arg(0), arg(1), arg(2)
	if sub == "" || mainArg == "" || name == "" {
		c.die("usage: integrate-change.sh prepare|commit|land|cleanup <main-checkout> <name> [flags]")
	}
	m, ok := icPhys(c.env.Dir, mainArg)
	if !ok {
		c.die("%s is not a directory", mainArg)
	}
	c.main = m
	if !changeNamePlain(name) {
		c.die("change name '%s' is not a plain change name", name)
	}
	c.name = name
	c.harness, c.token = c.env.Getenv("FLOW_HARNESS"), c.env.Getenv("FLOW_SESSION_TOKEN")
	c.planMsg = "chore(spectre): plan " + name
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
		case "--accept-foreign":
			c.acceptForeign = true
		case "--accept-outstanding":
			c.acceptOutstanding = true
		case "--outstanding":
			// The planning commit's subject is always this script's own: the
			// outstanding work rides in its body, so a text with no Conventional
			// Commits subject never reaches a commit-msg hook as one (it stopped
			// commit-split after the reshape when it replaced the subject whole).
			if next() == "" {
				c.die("--outstanding takes the outstanding work's text")
			}
			c.planMsg = fmt.Sprintf("chore(spectre): plan %s\n\nOutstanding: %s", name, next())
			i++
		case "--route":
			c.route = next()
			i++
		case "--pr-url":
			if !icPairArg(next()) {
				c.die("--pr-url takes <worktree>=<url|none>")
			}
			c.prArgs = append(c.prArgs, next())
			i++
		case "--proceed":
			p, ok := icPhys(c.env.Dir, next())
			if !ok {
				c.die("--proceed takes the main checkout its disclosure stop named")
			}
			c.proceed = append(c.proceed, p)
			i++
		default:
			if !strings.Contains(a, "=") {
				c.die("unknown argument '%s'", a)
			}
			c.subjects = append(c.subjects, a)
		}
	}
	if c.harness == "" || c.token == "" {
		c.die("--harness and --session-token are required (or FLOW_HARNESS / FLOW_SESSION_TOKEN)")
	}
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
	if err != nil || !isDir(r) {
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
		path, found = lookPath(c.env, name)
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
	if errors.Is(err, exec.ErrNotFound) {
		return 127
	}
	return exitCode(err)
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

func (c *ic) sibling(name string) string { return c.scriptDir + "/" + name + ".sh" }

// run is the bash's run: OUT the guard's stdout, relayed; RC its exit
// code. Its stderr goes straight through.
func (c *ic) run(guard string, args ...string) {
	c.out, c.rc = c.capture(c.cmd("/", c.sibling(guard), args...))
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
		fmt.Fprintf(c.stderr, "integrate-change: warning: flow stage begin %s failed\n", stage)
	}
}

func (c *ic) markEnd(stage, outcome string) {
	cmd := c.cmd("/", "flow", "stage", "end", "-C", c.main, "-command", "/flow", "-stage", stage, "-outcome", outcome, c.name)
	if cmd.Run() != nil {
		fmt.Fprintf(c.stderr, "integrate-change: warning: flow stage end %s failed\n", stage)
	}
	c.openStage = ""
}

func (c *ic) stop(kind, text string) {
	if c.openStage != "" {
		c.markEnd(c.openStage, c.stopOutcome)
	}
	c.say("STOP: " + kind + " — " + text)
	panic(icExit(1))
}

func (c *ic) flags() string { return "--harness " + c.harness + " --session-token " + c.token }

// visualFalsePositive: an operator flagged a false positive on <worktree>'s
// own check-visual-verify-dispatched verdict row for this change, the row
// naming its current HEAD (the verdict line ends "— HEAD <sha>") — so a
// flag never covers a later commit, or another worktree. A failed or
// unreadable read is no flag.
func (c *ic) visualFalsePositive(w string) bool {
	pw, ok := icPhys("/", w)
	if !ok {
		return false
	}
	h, ok := c.git("-C", w, "rev-parse", "HEAD")
	if !ok {
		return false
	}
	cmd := c.cmd("/", "flow", "record", "verdicts", "-guard", "check-visual-verify-dispatched", "-false-positive", "-C", c.canon)
	cmd.Stderr = nil
	out, _ := c.capture(cmd)
	var rows []struct {
		Change        string `json:"change"`
		FalsePositive any    `json:"falsePositive"`
		Worktree      string `json:"worktree"`
		Verdict       string `json:"verdict"`
	}
	if json.Unmarshal([]byte(out), &rows) != nil {
		return false
	}
	for _, r := range rows {
		if r.Change == c.name && r.FalsePositive == true && r.Worktree == pw && strings.HasSuffix(r.Verdict, " — HEAD "+h) {
			return true
		}
	}
	return false
}

// visualOK: check-visual-verify-dispatched passes <worktree>'s current
// HEAD, or its MISSING verdict for that HEAD is flagged a false positive.
// No verdict is its own stop.
func (c *ic) visualOK(w, base string) bool {
	mb, _ := c.git("-C", w, "merge-base", "HEAD", "origin/"+base)
	c.run("check-visual-verify-dispatched", w, c.name, mb, c.canon)
	v := fmt.Sprintf("%d:%s", c.rc, c.verdict("VISUAL-VERIFY-OK", "VISUAL-VERIFY-MISSING"))
	switch {
	case strings.HasPrefix(v, "0:VISUAL-VERIFY-OK:"):
		return true
	case strings.HasPrefix(v, "1:VISUAL-VERIFY-MISSING:"):
		return c.visualFalsePositive(w)
	}
	c.stop("no-verdict", fmt.Sprintf("check-visual-verify-dispatched reached no verdict for %s (exit %d) — report its output and ask the operator; --accept-outstanding does not pass this", w, c.rc))
	return false
}

// visualGate: visualOK on each worktree, stopping at the first MISSING —
// so that MISSING row is the newest verdict, the one
// `flow record verdict false-positive` flags.
func (c *ic) visualGate() {
	for i := range c.wt {
		if !c.visualOK(c.wt[i], c.bases[i]) {
			c.stop("visual-verify", visualStopText(c.wt[i], "prepare"))
		}
	}
}

func visualStopText(w, step string) string {
	return w + ": no verifier saw its current HEAD — dispatch flow.visual-verify on it, fix every departure in-run, commit, then re-run " + step + "; --accept-outstanding does not pass this. A false-positive flag is recorded now, while this worktree's verdict is the newest"
}

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
	p := gdcDirname(w)
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

// icState is the record `flow state get` prints, its worktrees in the
// record's own key order (jq's to_entries).
type icState struct {
	keys, values []string
	prURL, jira  string
}

func icParseState(s string) icState {
	var st icState
	var top map[string]json.RawMessage
	if json.Unmarshal([]byte(s), &top) != nil {
		return st
	}
	var pr, jira any
	_ = json.Unmarshal(top["prUrl"], &pr)
	_ = json.Unmarshal(top["jiraIssue"], &jira)
	st.prURL, st.jira = icJQText(pr), icJQText(jira)
	dec := json.NewDecoder(bytes.NewReader(top["worktrees"]))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return st
	}
	for dec.More() {
		k, err := dec.Token()
		if err != nil {
			return st
		}
		var v any
		if dec.Decode(&v) != nil {
			return st
		}
		val := icJQText(v)
		if val == "" {
			val = "-" // `.value // "-"`
		}
		st.keys, st.values = append(st.keys, k.(string)), append(st.values, val)
	}
	return st
}

// icJQText is `jq -r '.x // empty'` of one value.
func icJQText(v any) string {
	switch x := v.(type) {
	case nil, bool:
		if x == true {
			return "true"
		}
		return ""
	case string:
		return x
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// loadSet: the worktree set — wt[i] its path, mb[i] its recorded merge base
// (- when none), repo[i] its main checkout (mainOfGone's for one already
// gone, empty when that cannot be derived), mains the distinct main
// checkouts, <main-checkout> always among them.
func (c *ic) loadSet() {
	state, rc := c.capture(c.cmd("/", "flow", "state", "get", "-C", c.main, c.name))
	if rc != 0 {
		c.die("flow state get %s failed", c.name)
	}
	st := icParseState(state)
	c.prURLState, c.jiraKey = st.prURL, st.jira
	c.wt, c.mb, c.repo, c.mains = nil, nil, nil, []string{c.main}
	for i, k := range st.keys {
		if k != "" {
			c.wt, c.mb = append(c.wt, k), append(c.mb, st.values[i])
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
}

func icHas(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// findCanonical: canon the worktree whose change directory holds tasks.md;
// spec its tree.
func (c *ic) findCanonical() bool {
	c.canon, c.spec = "", ""
	for _, w := range c.wt {
		if !isDir(w) {
			continue
		}
		s := specRootLeaf(w, io.Discard)
		if isFile(w+"/"+s+"/changes/"+c.name+"/tasks.md") || isFile(w+"/"+s+"/changes/archive/"+c.name+"/tasks.md") {
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
	link := c.canon + "/" + c.spec + "/changes/" + c.name + "/link.md"
	if !isFile(link) {
		link = c.canon + "/" + c.spec + "/changes/archive/" + c.name + "/link.md"
	}
	if b, err := os.ReadFile(link); err == nil {
		on := false
		for _, l := range icLines(string(b)) {
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
	for _, l := range icLines(string(b)) {
		if f := strings.Fields(l); len(f) > 0 && f[0] == item {
			if len(f) > 1 {
				return f[1]
			}
			return ""
		}
	}
	return ""
}

// icLines splits text into lines as `read` sees them.
func icLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// baseOf: c.base the worktree's base branch, resolve-base-branch.sh's
// answer (it fetches).
func (c *ic) baseOf(i int) {
	var rc int
	c.base, rc = c.capture(c.cmd("/", c.sibling("resolve-base-branch"), c.wt[i]))
	if rc == 3 {
		c.stop("no-remote", c.wt[i]+": this repository has no remote, so there is nothing to push to or merge into")
	}
	if rc != 0 {
		c.stop("base", fmt.Sprintf("%s: no base branch resolved (resolve-base-branch exit %d) — ask the operator", c.wt[i], rc))
	}
}

// archBase: an archived re-run's base.
func (c *ic) archBase(w, mb string) string {
	log, _ := c.git("-C", w, "log", "--format=%H %s", mb+"..HEAD")
	if s := icFirstSubject(log, "docs(self-review): "+c.name+" self-review report", "docs(self-review): "+c.name+" self-review context bundle"); s != "" {
		return s
	}
	log, _ = c.git("-C", w, "log", "--format=%H %s", mb+"..HEAD")
	if s := icFirstSubject(log, "chore(spectre): archive "+c.name); s != "" {
		return s
	}
	return mb
}

// icFirstSubject: the hash of the first `<hash> <subject>` line whose
// subject is one of subjects.
func icFirstSubject(log string, subjects ...string) string {
	for _, l := range icLines(log) {
		h, s, _ := strings.Cut(l, " ")
		if icHas(subjects, s) {
			return h
		}
	}
	return ""
}

func (c *ic) reshapeFile(w string) string {
	f, _ := c.git("-C", w, "rev-parse", "--path-format=absolute", "--git-path", "flow-reshape-base")
	return f
}

var icLinkUnder = regexp.MustCompile(`/link\.md$`)

// implPaths: the paths commit-split would put in the implementation commit
// after a reshape to <rev> — every change against <rev>, committed, staged,
// unstaged or untracked, outside <spec>/changes/ but for a link.md in any
// directory below it — commit-split's `*/link.md` pathspec, whose `*`
// crosses `/`. Empty means that commit is skipped, so the worktree needs no
// implementation subject.
func (c *ic) implPaths(w, rev string) []string {
	p := specRootLeaf(w, c.stderr) + "/changes/"
	diff, _ := c.capture(c.cmd("/", "git", "-C", w, "diff", "--no-renames", "--name-only", rev))
	others, _ := c.capture(c.cmd("/", "git", "-C", w, "ls-files", "--others", "--exclude-standard"))
	var out []string
	for _, l := range append(icLines(diff), icLines(others)...) {
		if !strings.HasPrefix(l, p) || icLinkUnder.MatchString(l[len(p):]) {
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

// guardTests: run each GUARD-TEST the last sync-onto-base named; a failure
// stops with the test's own output above the stop line.
func (c *ic) guardTests() {
	for _, l := range icLines(c.out) {
		if !strings.HasPrefix(l, "GUARD-TEST: ") {
			continue
		}
		t := l
		if k := strings.LastIndex(t, " — "); k >= 0 {
			t = t[k+len(" — "):]
		}
		cmd := c.cmd("/", "bash", t)
		var b bytes.Buffer
		cmd.Stdout, cmd.Stderr = &b, &b
		if cmd.Run() != nil {
			fmt.Fprintln(c.stdout, strings.TrimRight(b.String(), "\n"))
			c.stop("guard-test", "the scoped re-verification "+t+" failed — report its output above; the change stays IN_PROGRESS")
		}
		c.say("GUARD-TEST-PASSED: " + t)
	}
}

// rebaseLint: a textually clean rebase can still break the build — one
// dropped a Go import no exit code caught (KAN-900) — so the rebased tree
// runs its project's ## lint list (run-verify-list.sh) under the KEY=value
// lines prepare-workspace.sh prints for it, as that script's caller
// contract asks. A failure stops with the list's report above the stop
// line; a project declaring no list (exit 3) has nothing to run. A passing
// list that rewrote tracked files (a formatter pre-step) stops too: the
// rewrite would otherwise ride into the reshape unreviewed or be left
// behind a push. What the list rewrote is the tracked diff's change across
// it, never the diff itself: the rebase sets the planning paths aside and
// restores them, so an uncommitted narrative edit is already there
// (kan-1010's self-review). The prefix leads the stop text — land passes
// what it already pushed.
func (c *ic) rebaseLint(w, prefix string) {
	before, _ := c.git("-C", w, "diff", "--no-renames", "HEAD")
	out, rc := c.capture(c.cmd("/", c.sibling("prepare-workspace"), w))
	if rc != 0 {
		fmt.Fprintln(c.stdout, out)
		c.stop("rebase-lint", prefix+"prepare-workspace.sh failed on "+w+" after its rebase, so its ## lint list did not run — report its output above; the change stays IN_PROGRESS")
	}
	cmd := c.cmd("/", c.sibling("run-verify-list"), w, "lint")
	cmd.Env = append([]string{}, c.childEnv...)
	for _, kv := range icLines(out) {
		if kv != "" && (kv[0] == '_' || kv[0] >= 'A' && kv[0] <= 'Z' || kv[0] >= 'a' && kv[0] <= 'z') && strings.Contains(kv, "=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	var b bytes.Buffer
	cmd.Stdout, cmd.Stderr = &b, &b
	rc = rcOf(cmd.Run())
	switch rc {
	case 0:
		if after, _ := c.git("-C", w, "diff", "--no-renames", "HEAD"); after != before {
			c.stop("rebase-lint", prefix+"the ## lint list rewrote tracked files on "+w+" after its rebase — commit or revert them; the change stays IN_PROGRESS")
		}
		c.say("REBASE-LINT-PASSED: " + w)
	case 3:
		c.say("REBASE-LINT-NONE: " + w + " — its project declares no ## lint list")
	default:
		fmt.Fprintln(c.stdout, strings.TrimRight(b.String(), "\n"))
		c.stop("rebase-lint", fmt.Sprintf("%sthe ## lint list failed on %s after its rebase (run-verify-list exit %d) — report its output above; the change stays IN_PROGRESS", prefix, w, rc))
	}
}

// pinPreReshape: refs/flow/pre-reshape/<name> at the worktree's HEAD before
// the reshape folds its per-task commits — the one ref that still reaches
// each `Task-Id:` commit and the pre-task tasks.md afterwards. Absent
// locally, origin's is fetched first (a re-clone). Kept when it already
// reaches HEAD (a re-run after a stop already reshaped); otherwise advanced
// to a merge of the old pin and HEAD, so an archived re-run's new tasks
// stay reachable beside the earlier ones.
func (c *ic) pinPreReshape(w string) {
	r := "refs/flow/pre-reshape/" + c.name
	if _, ok := c.git("-C", w, "rev-parse", "-q", "--verify", r); !ok {
		c.gitQuiet("-C", w, "fetch", "-q", "origin", r+":"+r)
	}
	if p, ok := c.git("-C", w, "rev-parse", "-q", "--verify", r); ok {
		if _, ok := c.git("-C", w, "merge-base", "--is-ancestor", "HEAD", p); ok {
			return
		}
		if p, ok = c.git("-C", w, "commit-tree", "HEAD^{tree}", "-p", p, "-p", "HEAD", "-m", "pre-reshape "+c.name); ok {
			if _, ok := c.git("-C", w, "update-ref", r, p); ok {
				return
			}
		}
	} else if _, ok := c.git("-C", w, "update-ref", r, "HEAD"); ok {
		return
	}
	c.stop("pre-reshape", "pinning "+w+"'s HEAD as "+r+" failed — report git's output; this worktree was not reshaped")
}

// pushPin: the pin to origin, plain, when the worktree has one; origin's is
// replaced only when the local pin descends from it — any other differing
// pin is refused, never overwritten.
func (c *ic) pushPin(w string) {
	r := "refs/flow/pre-reshape/" + c.name
	l, ok := c.git("-C", w, "rev-parse", "-q", "--verify", r)
	if !ok {
		return
	}
	rem, ok := c.git("-C", w, "ls-remote", "origin", r)
	if !ok {
		c.stop("push", "reading origin's "+r+" from "+w+" failed — report git's output; the change stays IN_PROGRESS")
	}
	if k := strings.IndexAny(rem, " \t\n"); k >= 0 {
		rem = rem[:k]
	}
	if rem == l {
		return
	}
	if rem != "" {
		if _, ok := c.gitQuiet("-C", w, "merge-base", "--is-ancestor", rem, l); !ok {
			c.stop("pre-reshape", "origin's "+r+" is "+rem+", not an ancestor of "+w+"'s pin "+l+" — report it; neither was overwritten")
		}
	}
	if !c.gitPass("-C", w, "push", "-q", "origin", r+":"+r) {
		c.stop("push", "the push of "+r+" from "+w+" was rejected — report git's output; the change stays IN_PROGRESS")
	}
}

func (c *ic) pushBranch(w string) {
	if !c.gitPass("-C", w, "push", "--force-with-lease", "-u", "origin", "spectre/"+c.name) {
		c.stop("push", "the branch push from "+w+" was rejected — report git's output; the change stays IN_PROGRESS")
	}
}
