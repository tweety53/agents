package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// changePlanRef, changePlanPath and changePlanDir -- the three functions the
// guards call -- began as a port of a bash library whose bash twin was
// deleted (KAN-887); its header moved here and is canonical for the
// resolution order and the containment rules. Each returns the bash
// function's exit status as an int (0 resolved, 1 unresolvable, 3 an
// ambiguous store answer) and writes its stderr lines to stderr.
//
// One owner for "where is this change's tasks.md, really" — KAN-363 task 7,
// called by check-unfinished-work (task 8) and check-task-commit-fields
// (task 9) in place of composing `<worktree>/<spec-root>/changes/<name>`
// directly. KAN-343 shipped a cross-repository change by hand: the second
// repository's worktree carried no plan at all, and every guard that looked
// for one there reported the absence as a verdict — `OUTSTANDING: no plan at
// <path>`, or exit 2 with nothing further — because "no tasks.md" and "this
// change's plan lives somewhere else" were the same observation to them. This
// file is what tells the two apart, by reading `link.md` (spectre task 1's
// grammar) the way a caller of `spectre validate` would.
//
// Resolution order, per design.md's `guards-take-the-canonical-worktree-path`:
//
//   1. <worktree>/<spec-root>/changes/<name>/tasks.md exists → that path;
//      or no <worktree>/<spec-root>/changes/<name>/ directory at all and
//      <worktree>/<spec-root>/changes/archive/<name>/tasks.md exists → that
//      path. Integrate's run 1 archives the change on its branch before the
//      route, and a fix run after it writes into the archived directory, so
//      the archived plan is this change's plan, not a stale copy. Every
//      worktree read in the steps below resolves a change's directory the
//      same way (_change_plan_change_dir).
//   2. Otherwise, no local tasks.md and no local link.md for <name> here at
//      all (the cross-repo shape where the plan lives only in the canonical
//      repo — the change directory absent entirely, or an empty scaffold a
//      refused or aborted link step left behind, KAN-430) and a canonical
//      worktree argument was passed → the same-named plan under the
//      canonical worktree's own <spec-root>/changes/, if it exists. The gate
//      is BOTH files being absent, not merely an absent tasks.md or an
//      absent directory, so a satellite's link resolution below is never
//      preempted (design.md's no-fallback-for-linked-satellites — whose
//      letter named the absent directory; the empty scaffold is the recorded
//      widening, an existing-but-empty directory carrying no signal of its
//      own).
//   3. Otherwise, <worktree>/<spec-root>/changes/<name>/link.md exists and
//      carries `## Part of`:
//        - a canonical worktree argument was passed → its own
//          <spec-root>/changes/<canonical-id>/tasks.md, if that exists;
//          no fallback to peers when the argument was passed but the plan
//          is not there — a guard that named its own canonical worktree
//          already resolved it through flow's own worktree set, and a
//          wrong answer from that set is a fact worth failing loudly on,
//          not a cue to go looking somewhere else.
//        - no canonical worktree argument → resolve the peer name through
//          <worktree>/<spec-root>/peers and print the plan there, if it
//          exists. A peer that is declared but not present on disk is not
//          a finding (design.md's `peer-absence-is-not-a-finding`) and,
//          here, not a refusal either — it is simply unresolvable, and the
//          caller decides what that means.
//   4. None of the above, and no canonical worktree argument was passed →
//      the store step (KAN-267): `flow state find` for the change's own
//      name, or for the link's canonical id when the change is a
//      satellite, and walk each matching record's `worktrees` map — the
//      one place that names every tree a cross-repo change affects, which
//      is what lets a guard resolve the plan from the satellite side with
//      no caller-supplied argument at all. See changePlanStoreDir's
//      own comment for the refusal and never-blocks rules.
//
//   5. None of the above → unresolvable.
//
// WHY A PEER ROOT DOES NOT FALL BACK TO changes/archive/ THE WAY spectre's
// own `internal/check` does for the peer side of a link (task 2's
// `independent-archive`). That fallback answers "is the counterpart change
// still findable at all", which matters to `validate` reporting archive skew.
// This function answers a narrower question for a guard running inside a
// live worktree: a peer root is a sibling repository's main checkout, where
// a change's archive appears only once the change has landed and nothing is
// left for a guard to judge, and a `peers`-relative archive lookup would
// require duplicating spectre's own two-step search for a case this
// repository's guards do not reach. Narrower is deliberate here, not an
// oversight. A worktree is different: integrate's run 1 archives on the
// change branch in a worktree that stays open until run 2, which is why
// step 1's archived half applies to every worktree this function reads.
//
// CONTAINMENT applies to every name concatenated into a path in this
// function: the change name (the caller's own argument, from a
// pull-request-editable state file), and the peer name and canonical change
// id (both read out of link.md, itself a pull-request-editable file this
// function parses). All three are checked against
// check-unfinished-work.sh's own allowlist — start with a letter or digit,
// then only letters, digits, '.', '_' and '-' — character for character, so
// `../escape` in any of the three positions cannot walk this function
// outside the worktree(s) it was given. #
// SYMLINKS: the file and directory tests follow symlinks, so a symlink at
// `<worktree>/<spec-root>/changes/<allowlisted-name>` is followed and its
// content returned as the change's plan — the same accepted tradeoff
// check-unfinished-work.sh's own header already ships and argues.
//
// Paths are composed by string concatenation, as the bash did, never
// filepath.Join, so every printed path is byte-identical to the helper's.
// ponytail: relative worktree paths are tested against the process cwd, as
// the bash helper tested them against its own; env.Dir is only the cwd for
// the git and flow subprocesses. Callers pass absolute paths today.

// changePlanLinkPartOf is _change_plan_link_part_of: the first code span
// under link.md's `## Part of` heading, stopping at the next `## ` heading.
func changePlanLinkPartOf(file string) (string, bool) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", false
	}
	inSection := false
	ref := ""
	for _, line := range lines(b) {
		if inSection {
			if strings.HasPrefix(line, "## ") {
				break
			}
			if _, rest, ok := strings.Cut(line, "`"); ok {
				ref, _, _ = strings.Cut(rest, "`")
				if ref != "" {
					break
				}
			}
		}
		if line == "## Part of" {
			inSection = true
		}
	}
	return ref, ref != ""
}

// changePlanMainCheckout is _change_plan_main_checkout: the main checkout
// beside which a peers file's relative paths resolve.
func changePlanMainCheckout(env Env, dir string) (string, bool) {
	common, ok := gitOut(env, "-C", dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if !ok || common == "" {
		return "", false
	}
	if cw, _ := gitOut(env, "config", "--file", common+"/config", "core.worktree"); cw != "" {
		if strings.HasPrefix(cw, "/") {
			return cw, true
		}
		return common + "/" + cw, true
	}
	if strings.HasSuffix(common, "/.git") {
		return filepath.Dir(common), true
	}
	gitdir, ok := gitOut(env, "-C", dir, "rev-parse", "--absolute-git-dir")
	if ok && gitdir == common {
		return dir, true
	}
	return "", false
}

// changePlanPeerRoot is _change_plan_peer_root: the peer's checkout, read
// from <worktree>/<spec_root>/peers and resolved beside the main checkout.
func changePlanPeerRoot(env Env, worktree, specRoot, peerName string) (string, bool) {
	b, err := os.ReadFile(worktree + "/" + specRoot + "/peers")
	if err != nil {
		return "", false
	}
	resolved := ""
	for _, line := range lines(b) {
		// `read -r pname prel rest` with the default IFS: space and tab.
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' })
		if len(f) == 0 {
			continue
		}
		if f[0] == peerName {
			if len(f) > 1 {
				resolved = f[1]
			}
			break
		}
	}
	if resolved == "" {
		return "", false
	}
	base, ok := changePlanMainCheckout(env, worktree)
	if !ok {
		base = worktree
	}
	// `( cd "$base/$resolved" && pwd )`: the logical path, `..` taken
	// lexically, provided it is a directory.
	p := base + "/" + resolved
	if !filepath.IsAbs(p) {
		p = filepath.Join(env.Dir, p)
	}
	p = filepath.Clean(p)
	if !isDir(p) {
		return "", false
	}
	return p, true
}

// changePlanStoreDir is _change_plan_store_dir: the change directory the
// state store's `flow state find <key>` records point at. 0 exactly one
// match, 1 none or the store is unusable, 3 more than one (named on stderr).
// jq is required only because the bash helper requires it; this port parses
// the JSON itself.
func changePlanStoreDir(env Env, stderr io.Writer, key string) (string, int) {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return "", 1
	}
	if _, ok := lookPath(env, "jq"); !ok {
		return "", 1
	}
	cmd := exec.Command(flow, "state", "find", key)
	cmd.Dir = env.Dir
	out, err := cmd.Output()
	if err != nil {
		return "", 1
	}
	var found struct {
		Records []struct {
			ProjectKey json.RawMessage            `json:"projectKey"`
			Worktrees  map[string]json.RawMessage `json:"worktrees"`
		} `json:"records"`
	}
	if json.Unmarshal(out, &found) != nil {
		return "", 1
	}
	var matches, projects []string
	for _, r := range found.Records {
		var proj string
		if json.Unmarshal(r.ProjectKey, &proj) != nil || proj == "" {
			continue
		}
		trees := make([]string, 0, len(r.Worktrees))
		for tree := range r.Worktrees {
			trees = append(trees, tree)
		}
		sort.Strings(trees) // jq's `keys[]` order
		for _, tree := range trees {
			if tree == "" {
				continue
			}
			dir := ChangeDir(tree+"/"+specRootLeaf(tree, io.Discard)+"/changes", key)
			if isFile(dir + "/tasks.md") {
				matches = append(matches, dir)
				projects = append(projects, proj)
			}
		}
	}
	switch {
	case len(matches) > 1:
		for j := range matches {
			fmt.Fprintf(stderr, "change-plan: ambiguous state-record resolution for '%s': project %s resolves to %s\n", key, projects[j], matches[j])
		}
		return "", 3
	case len(matches) == 1:
		return matches[0], 0
	}
	return "", 1
}

// ChangeDir is a change's own directory under changes (a tree's
// <spec-root>/changes): changes/<name> while it exists, else
// changes/archive/<name> when that holds a tasks.md -- integrate's run 1
// archives the change on its branch before the route, and a fix run after
// it writes into the archived directory. The name is the caller's to have
// allowlisted. Every worktree changePlanDir reads -- the judged one, the
// canonical worktree argument, the store's worktrees -- goes through it, as
// _change_plan_change_dir in the bash twin; exported for `flow tasks`, which
// resolves no further than this.
func ChangeDir(changes, name string) string {
	if !isDir(changes+"/"+name) && isFile(changes+"/archive/"+name+"/tasks.md") {
		return changes + "/archive/" + name
	}
	return changes + "/" + name
}

// changePlanDir is change_plan_dir (_change_plan_resolve_dir): the directory
// holding the change's canonical tasks.md.
func changePlanDir(env Env, stderr io.Writer, worktree, name, canonical string) (string, int) {
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "change-plan: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)
		return "", 1
	}
	specRoot := specRootLeaf(worktree, stderr)
	dir := ChangeDir(worktree+"/"+specRoot+"/changes", name)
	if isFile(dir + "/tasks.md") {
		return dir, 0
	}
	link := dir + "/link.md"
	if !isFile(link) && canonical != "" {
		absentDir := ChangeDir(canonical+"/"+specRootLeaf(canonical, stderr)+"/changes", name)
		if isFile(absentDir + "/tasks.md") {
			return absentDir, 0
		}
	}
	if !isFile(link) {
		if canonical == "" {
			return changePlanStoreDir(env, stderr, name)
		}
		return "", 1
	}
	ref, ok := changePlanLinkPartOf(link)
	if !ok {
		if canonical == "" {
			return changePlanStoreDir(env, stderr, name)
		}
		return "", 1
	}
	peer, _, _ := strings.Cut(ref, ":")
	changeID := ref
	if _, after, found := strings.Cut(ref, ":"); found {
		changeID = after
	}
	if !plainChangeName(peer) {
		fmt.Fprintf(stderr, "change-plan: peer name '%s' in %s is not a plain peer name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", peer, link)
		return "", 1
	}
	if !plainChangeName(changeID) {
		fmt.Fprintf(stderr, "change-plan: canonical change id '%s' in %s is not a plain change id — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", changeID, link)
		return "", 1
	}
	if canonical != "" {
		canonDir := ChangeDir(canonical+"/"+specRootLeaf(canonical, stderr)+"/changes", changeID)
		if isFile(canonDir + "/tasks.md") {
			return canonDir, 0
		}
		return "", 1
	}
	peerRoot, ok := changePlanPeerRoot(env, worktree, specRoot, peer)
	if !ok {
		return changePlanStoreDir(env, stderr, changeID)
	}
	peerDir := peerRoot + "/" + specRootLeaf(peerRoot, stderr) + "/changes/" + changeID
	if isFile(peerDir + "/tasks.md") {
		return peerDir, 0
	}
	return changePlanStoreDir(env, stderr, changeID)
}

// changePlanPath is change_plan_path: changePlanDir's answer plus /tasks.md.
func changePlanPath(env Env, stderr io.Writer, worktree, name, canonical string) (string, int) {
	dir, rc := changePlanDir(env, stderr, worktree, name, canonical)
	if rc != 0 {
		return "", rc
	}
	return dir + "/tasks.md", 0
}

// changePlanRef is change_plan_ref: a satellite's `<peer>:<change-id>` from
// its link.md, or false for a local plan, a missing link.md, or no
// `## Part of`.
func changePlanRef(stderr io.Writer, worktree, name string) (string, bool) {
	if !plainChangeName(name) {
		return "", false
	}
	dir := worktree + "/" + specRootLeaf(worktree, stderr) + "/changes/" + name
	if isFile(dir+"/tasks.md") || !isFile(dir+"/link.md") {
		return "", false
	}
	return changePlanLinkPartOf(dir + "/link.md")
}
