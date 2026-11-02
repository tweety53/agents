package guard

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
	"slices"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/kickoff-worktree.sh shim (KAN-1058).
//
// flow-guard kickoff-worktree — /flow's kickoff steps 1–4 (skills/flow/brainstorm.md
// **A**), run in order in one call so no session re-types them.
//
// Usage: flow-guard kickoff-worktree [--from <worktree>] <project> <name> [<base>]
//
// <base>, when given, names the branch the change is cut from and lands on
// instead of origin/HEAD's: an origin/<base> this repository's origin lacks is
// first pushed from its origin default branch (one stderr line), so every
// repository of the change records the same base; a fresh branch is created
// from it, and either form records
// it as `git config branch.spectre/<name>.flowBase <base>`, which
// flow-guard resolve-base-branch reads ahead of origin/HEAD. A name failing
// resolve-base-branch's validation is exit 2.
//
// --from <worktree> names one of this change's existing worktrees, for every
// worktree a change adds after kickoff. With no <base>, the base is that
// worktree's `branch.spectre/<name>.flowBase` — none recorded, the default
// branch, exactly as without --from, and no flowBase is written. An explicit
// <base> wins over the recorded one. The new worktree is persisted into the
// change's state record through `flow state add-worktree -C <worktree>`, so
// the record is found under the change's own project. An empty or missing
// <worktree> is exit 2.
//
// With --from and `<worktree>/spectre/changes/<name>/` present, the merge-order
// record is written there too (kan-1017: a hand-written one failed `spectre
// validate` twice): this repository's name is its entry in
// `<worktree>/spectre/peers` — one `<name> <path>` line per peer, the path
// relative to <worktree>'s main checkout — resolved before step 1, a
// repository the file does not declare (or an unreadable file) exit 2 with
// nothing added. After step 4, `link.md` there names every repository the
// change's state record holds (`flow state get -C <worktree> <name>`), each
// worktree resolved to its `spectre/peers` name through its main checkout,
// <worktree>'s own repository and a worktree no peer names skipped, in
// `spectre/peers` order: a new file is written as `## Merge order` then the
// entries `.` and the first name, numbered, each name in backticks; an
// existing one whose `## Merge order` lacks a name gets it as the next
// numbered entry, nothing reordered; one already naming it, or holding no
// such section, is left as it is. The file is never
// committed.
//
//   0. `flow` must be on PATH, checked first: a worktree step 2 could not
//      persist is never created.
//   1. flow-guard check-worktree-location <project> — its exit 1 or 2 stops the run
//      with its own lines, as this script's exit 1 or 2.
//   2. `git fetch origin`; when origin/spectre/<name> exists, `git worktree
//      add <project>-worktrees/<name> spectre/<name>`, a local branch
//      tracking it, otherwise `-b spectre/<name> origin/<default-branch>`,
//      the default branch read from refs/remotes/origin/HEAD, never HEAD —
//      or origin/<base> when <base> is given.
//      The merge base is `git -C <worktree> rev-parse HEAD` right after the
//      add, persisted before anything else runs by `flow state add-worktree
//      -C <project> <name> <worktree> <merge-base>` (`-C <worktree>` of
//      --from when given). <project>-worktrees/ is
//      the sibling of <project>'s physical path, outside the repository, so
//      nothing needs ignoring (design.md: sibling-worktrees-layout).
//   3. flow-guard project-get <worktree> "worktree setup": exit 0 runs every line
//      inside the body's ``` fences, in order, from the worktree root, in the
//      foreground — never the fence markers or the prose around them; exit 1
//      (no such section) is one stderr line and the run continues; exit 2
//      stops the run.
//   4. `git -C <worktree> push -u origin spectre/<name>`.
//
// On success stdout carries exactly:
//   worktree: <project>-worktrees/<name>
//   merge-base: <sha>
// then `merge-order: <link.md>` when the record above was written.
//
// Exit codes: 0 all four steps done; 1 a step stopped the run — the location
// guard's refusal, a failed git call (creating a missing origin/<base>
// among them), `flow state add-worktree` refusing (no
// STARTED record), a setup command's non-zero exit (named, with its output)
// or a failed push — its lines relayed on stderr, and the worktree already
// persisted when step 2 got that far; 2 usage, `flow` absent from PATH, no
// resolvable default branch, the location guard's or project-get's own
// cannot-answer exit, a project path that no longer resolves after it, or a
// sibling this shim sets unset.
//
// Go, in flow-guard (stats/internal/guard/kickoffworktree.go). flow-guard is
// built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
// derives it, and exits 2 with the cause when it cannot.

// kickoffWorktree is scripts/kickoff-worktree.sh: `/flow`'s kickoff steps
// 1–4 (skills/flow/brainstorm.md **A**) in order — the location guard, the
// worktree add at the sibling <project>-worktrees/ root (design.md:
// sibling-worktrees-layout) persisted through `flow state add-worktree`, the
// project's `## worktree setup` commands, the branch push. The shim's header
// is the contract: 0 all four done (`worktree:` and `merge-base:` on
// stdout), 1 a step stopped the run (its lines relayed; the worktree is
// already persisted when step 2 got that far), 2 usage or it cannot answer.
func init() {
	Registry["kickoff-worktree"] = kickoffWorktree
}

// kwRun runs name args from dir (or env's directory), returning its exit
// code and combined output.
func kwRun(env Env, dir, name string, args ...string) (int, string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = env.Dir
	if dir != "" {
		cmd.Dir = dir
	}
	return kwCombined(cmd)
}

// kwCombined runs cmd, returning its exit code and combined output.
func kwCombined(cmd *exec.Cmd) (int, string) {
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	return exitCode(cmd.Run()), out.String()
}

// kwFenced is every line inside a ``` fence of body, in order: the commands
// of a `## worktree setup` section, never its fence markers or prose.
func kwFenced(body string) []string {
	var cmds []string
	in := false
	for _, l := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			in = !in
			continue
		}
		if in && strings.TrimSpace(l) != "" {
			cmds = append(cmds, l)
		}
	}
	return cmds
}

func kickoffWorktree(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "kickoff-worktree"
	const usage = "usage: flow-guard kickoff-worktree [--from <worktree>] <project> <name> [<base>]\n"
	// --from names this change's existing worktree: the base defaults to the
	// flowBase it records, and the state record is found through it.
	from := ""
	if len(args) > 0 && args[0] == "--from" {
		if len(args) < 2 || args[1] == "" {
			fmt.Fprint(stderr, usage)
			return 2
		}
		from, args = smcAbs(env, args[1]), args[2:]
	}
	if len(args) < 2 || len(args) > 3 || args[0] == "" || args[1] == "" || (len(args) == 3 && args[2] == "") {
		fmt.Fprint(stderr, usage)
		return 2
	}
	project, name := smcAbs(env, args[0]), args[1]
	// An optional base names the branch the change is cut from and lands
	// on; it flows into refs and a git config value, so it passes the
	// name validation resolve-base-branch applies.
	base := ""
	if len(args) == 3 {
		base = args[2]
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	if base == "" && from != "" {
		// Outside a repository `config --get` still reads the global config
		// and exits 1, so the worktree is checked first.
		if rc, _ := kwRun(env, "", git, "-C", from, "rev-parse", "--git-dir"); rc != 0 {
			fmt.Fprintf(stderr, "%s: --from %s is not a git worktree — cannot read its base, so none is added\n", self, from)
			return 2
		}
		// Exit 1 is no flowBase: the default branch, as without --from; any
		// other failure cannot answer the base.
		switch rc, out := kwRun(env, "", git, "-C", from, "config", "--get", "branch.spectre/"+name+".flowBase"); rc {
		case 0:
			base = strings.TrimSpace(out)
		case 1:
		default:
			fmt.Fprint(stderr, out)
			return 2
		}
	}
	if base != "" {
		for i := 0; i < len(base); i++ {
			if !rbbNameByte(base[i], i > 0) {
				fmt.Fprintf(stderr, "%s: the base branch name is invalid\n", self)
				return 2
			}
		}
	}
	// sr15 (kan-1017): with --from and the change's directory in that
	// worktree, the new repository's name in its spectre/peers is resolved
	// before anything is cut — a repository the file does not declare stops
	// here, nothing added or recorded.
	link, peerName := "", ""
	if from != "" && isDir(filepath.Join(from, "spectre/changes", name)) {
		link = filepath.Join(from, "spectre/changes", name, "link.md")
		var msg string
		if peerName, msg = kwPeerName(env, git, from, project); msg != "" {
			fmt.Fprintf(stderr, "%s: %s — cannot write %s's merge order, so no worktree is added\n", self, msg, link)
			return 2
		}
	}
	// Before step 1: a worktree step 2 could not persist is never created.
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintf(stderr, "%s: no flow binary on PATH — step 2 could not persist the worktree, so none is created\n", self)
		return 2
	}
	relay := func(rc int, out string) int {
		fmt.Fprint(stderr, out)
		return rc
	}

	// 1. The location guard: its exit 1 or 2 stops the run with its lines.
	if rc, out := kwCombined(siblingCmd(env, "FLOW_GUARD_WORKTREE_LOCATION", "check-worktree-location", project)); rc != 0 {
		if rc != 1 {
			rc = 2
		}
		return relay(rc, out)
	}

	// 2. Fetch, add in one of two forms, persist before anything else runs.
	// The worktree lives beside the main checkout's physical path — the
	// root the location guard resolves — outside the repository, so nothing
	// needs ignoring; `git worktree add` creates the root on first use.
	physical, err := filepath.EvalSymlinks(project)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot resolve %s: %v\n", self, project, err)
		return 2
	}
	branch := "spectre/" + name
	wt := filepath.Join(siblingRoot(physical), name)
	if rc, out := kwRun(env, "", git, "-C", project, "fetch", "origin"); rc != 0 {
		return relay(1, out)
	}
	// The default branch by name, never HEAD: the main checkout may be on
	// any branch and is never moved.
	defaultBranch := func() (string, bool) {
		rc, out := kwRun(env, "", git, "-C", project, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
		if rc != 0 {
			fmt.Fprintf(stderr, "%s: cannot resolve the default branch — refs/remotes/origin/HEAD is unset (git remote set-head origin --auto)\n", self)
			return "", false
		}
		return strings.TrimSpace(out), true
	}
	// A base this origin lacks is created from its default branch, so every
	// repository of the change records the same base.
	if base != "" {
		if rc, _ := kwRun(env, "", git, "-C", project, "rev-parse", "-q", "--verify", "refs/remotes/origin/"+base); rc != 0 {
			def, ok := defaultBranch()
			if !ok {
				return 2
			}
			if rc, out := kwRun(env, "", git, "-C", project, "push", "origin", "refs/remotes/"+def+":refs/heads/"+base); rc != 0 {
				return relay(1, out)
			}
			if rc, out := kwRun(env, "", git, "-C", project, "fetch", "origin", base); rc != 0 {
				return relay(1, out)
			}
			fmt.Fprintf(stderr, "%s: origin/%s did not exist in %s — created it from %s\n", self, base, project, def)
		}
	}
	add := []string{"-C", project, "worktree", "add", wt, branch}
	if rc, _ := kwRun(env, "", git, "-C", project, "rev-parse", "-q", "--verify", "origin/"+branch); rc != 0 {
		start := "origin/" + base
		if base == "" {
			def, ok := defaultBranch()
			if !ok {
				return 2
			}
			start = def
		}
		add = []string{"-C", project, "worktree", "add", wt, "-b", branch, start}
	} else if base != "" {
		// Resuming a pushed branch: it must already sit on <base>, or
		// flowBase would record a base the branch was never cut from.
		if rc, _ := kwRun(env, "", git, "-C", project, "merge-base", "--is-ancestor", "origin/"+base, "origin/"+branch); rc != 0 {
			fmt.Fprintf(stderr, "%s: origin/%s does not descend from origin/%s in %s — stop and ask the operator: rebase it with git rebase --onto origin/%s <its merge base> %s and push, then re-run\n", self, branch, base, project, base, branch)
			return 1
		}
	}
	if rc, out := kwRun(env, "", git, add...); rc != 0 {
		return relay(1, out)
	}
	rc, out := kwRun(env, "", git, "-C", wt, "rev-parse", "HEAD")
	if rc != 0 {
		return relay(1, out)
	}
	mergeBase := strings.TrimSpace(out)
	stateDir := project
	if from != "" {
		stateDir = from
	}
	if rc, out := kwRun(env, "", flow, "state", "add-worktree", "-C", stateDir, name, wt, mergeBase); rc != 0 {
		return relay(1, out)
	}
	// Recorded on the branch, so resolve-base-branch and every guard it
	// feeds land the change back on <base> rather than on origin/HEAD.
	if base != "" {
		if rc, out := kwRun(env, "", git, "-C", wt, "config", "branch."+branch+".flowBase", base); rc != 0 {
			return relay(1, out)
		}
	}

	// 3. The project's worktree setup, fenced lines only, in order.
	if rc := kwSetup(env, wt, self, stderr); rc != 0 {
		return rc
	}

	// 4. The branch exists on the remote from its first minute.
	if rc, out := kwRun(env, "", git, "-C", wt, "push", "-u", "origin", branch); rc != 0 {
		return relay(1, out)
	}
	fmt.Fprintf(stdout, "worktree: %s\nmerge-base: %s\n", wt, mergeBase)
	if link != "" {
		// Every repository the state record holds is listed, not only the
		// one added now: a worktree added before the change directory
		// existed would otherwise be left out of the merge order.
		names := kwRecordedPeers(env, git, flow, from, name)
		if !slices.Contains(names, peerName) {
			names = append(names, peerName)
		}
		changed := false
		for _, n := range names {
			wrote, err := kwMergeOrder(link, n)
			if err != nil {
				fmt.Fprintf(stderr, "%s: cannot write %s: %v\n", self, link, err)
				return 1
			}
			changed = changed || wrote
		}
		if changed {
			fmt.Fprintf(stdout, "merge-order: %s\n", link)
		}
	}
	return 0
}

// kwRecordedPeers is the spectre/peers name of every worktree from's state
// record holds for name, in spectre/peers order; from's own repository and a
// worktree no peer names are skipped. A record that cannot be read yields
// none.
func kwRecordedPeers(env Env, git, flow, from, name string) []string {
	// stdout alone: a fallback warning on stderr is not part of the record.
	get := exec.Command(flow, "state", "get", "-C", from, name)
	get.Dir = env.Dir
	out, err := get.Output()
	if err != nil {
		return nil
	}
	var rec struct {
		Worktrees map[string]json.RawMessage `json:"worktrees"`
	}
	if json.Unmarshal(out, &rec) != nil {
		return nil
	}
	peers, own, msg := kwPeers(env, git, from)
	if msg != "" {
		return nil
	}
	held := map[string]bool{}
	for wt := range rec.Worktrees {
		rc, out := kwRun(env, "", git, "-C", wt, "rev-parse", "--path-format=absolute", "--git-common-dir")
		if rc != 0 {
			continue
		}
		if m := kwReal(filepath.Dir(strings.TrimSpace(out))); m != own {
			held[m] = true
		}
	}
	var names []string
	for _, p := range peers {
		if held[p[1]] {
			names = append(names, p[0])
		}
	}
	return names
}

// kwReal is p with symlinks resolved, or cleaned when it cannot be.
func kwReal(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// kwPeers is from's spectre/peers as {name, resolved path} pairs in file
// order, and from's own main checkout, resolved. A non-empty message is why
// they cannot be read.
func kwPeers(env Env, git, from string) ([][2]string, string, string) {
	body, err := os.ReadFile(filepath.Join(from, "spectre/peers"))
	if err != nil {
		return nil, "", "cannot read " + filepath.Join(from, "spectre/peers")
	}
	rc, out := kwRun(env, "", git, "-C", from, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if rc != 0 {
		return nil, "", "cannot resolve " + from + "'s main checkout"
	}
	main := filepath.Dir(strings.TrimSpace(out))
	var peers [][2]string
	for _, line := range strings.Split(string(body), "\n") {
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		p := f[1]
		if !filepath.IsAbs(p) {
			p = filepath.Join(main, p)
		}
		peers = append(peers, [2]string{f[0], kwReal(p)})
	}
	return peers, kwReal(main), ""
}

// kwPeerName is project's name in from's spectre/peers: one `<name> <path>`
// line per peer, the path relative to from's main checkout. A non-empty
// message is why it cannot be named.
func kwPeerName(env Env, git, from, project string) (string, string) {
	peers, _, msg := kwPeers(env, git, from)
	if msg != "" {
		return "", msg
	}
	want := kwReal(project)
	for _, p := range peers {
		if p[1] == want {
			return p[0], ""
		}
	}
	return "", filepath.Join(from, "spectre/peers") + " declares no entry for " + project
}

// kwMergeOrder writes link with `## Merge order` naming `.` then peer, or
// appends peer as the next numbered entry of an existing section that lacks
// it; a file holding no such section is left alone. wrote is whether the
// file changed.
func kwMergeOrder(link, peer string) (bool, error) {
	entry := "`" + peer + "`"
	body, err := os.ReadFile(link)
	if errors.Is(err, fs.ErrNotExist) {
		return true, os.WriteFile(link, []byte("## Merge order\n\n1. `.`\n2. "+entry+"\n"), 0o644)
	}
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(body), "\n")
	in, last, n := false, -1, 0
	for i, l := range lines {
		if strings.HasPrefix(l, "## ") {
			if in {
				break
			}
			in = strings.TrimSpace(l) == "## Merge order"
			continue
		}
		if num, rest, ok := strings.Cut(l, ". "); in && ok {
			if v, err := strconv.Atoi(num); err == nil {
				if strings.TrimSpace(rest) == entry {
					return false, nil
				}
				last, n = i, v
			}
		}
	}
	if last < 0 {
		return false, nil
	}
	lines = slices.Insert(lines, last+1, fmt.Sprintf("%d. %s", n+1, entry))
	return true, os.WriteFile(link, []byte(strings.Join(lines, "\n")), 0o644)
}

// kwSetup runs wt's `## worktree setup` fenced lines, in order, from wt:
// 0 done or none declared, 1 a command failed (named, its output relayed),
// 2 project-get cannot answer. stdout alone is the section body; a warning
// project-get prints on stderr is relayed, never parsed as a command.
func kwSetup(env Env, wt, self string, stderr io.Writer) int {
	get := siblingCmd(env, "FLOW_GUARD_PROJECT_GET", "project-get", wt, "worktree setup")
	var body, getErr bytes.Buffer
	get.Stdout, get.Stderr = &body, &getErr
	switch rc := exitCode(get.Run()); rc {
	case 0:
		fmt.Fprint(stderr, getErr.String())
		for _, c := range kwFenced(body.String()) {
			if rc, cout := kwRun(env, wt, "bash", "-c", c); rc != 0 {
				fmt.Fprintf(stderr, "%s: worktree setup command failed (exit %d): %s\n", self, rc, c)
				fmt.Fprint(stderr, cout)
				return 1
			}
		}
	case 1:
		fmt.Fprintf(stderr, "%s: no ## worktree setup declared — skipped\n", self)
	default:
		fmt.Fprint(stderr, getErr.String())
		return 2
	}
	return 0
}
