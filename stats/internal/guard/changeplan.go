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

// The port of scripts/lib/change-plan.sh's change_plan_ref, change_plan_path
// and change_plan_dir -- the three functions the ported guards call. That
// file's header is canonical for the resolution order and the containment
// rules; this is its body in Go, returning the bash function's exit status
// as an int (0 resolved, 1 unresolvable, 3 an ambiguous store answer) and
// writing its stderr lines to stderr.
//
// Paths are composed by string concatenation, as the bash does, never
// filepath.Join, so every printed path is byte-identical to the helper's.
// ponytail: relative worktree paths are tested against the process cwd, as
// the bash helper tests them against its own; env.Dir is only the cwd for
// the git and flow subprocesses. Callers pass absolute paths today.

// lines splits a file's content the way `while IFS= read -r line || [ -n
// "$line" ]` reads it: a trailing newline does not add an empty last line.
func lines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

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

// gitOut runs git in env.Dir and returns its stdout with trailing newlines
// stripped, as command substitution does.
func gitOut(env Env, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = env.Dir
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimRight(string(out), "\n"), true
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

// lookPath is `command -v <name>` over env's PATH, not the process's.
func lookPath(env Env, name string) (string, bool) {
	for _, dir := range filepath.SplitList(env.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		p := dir + "/" + name
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() && fi.Mode()&0o111 != 0 {
			return p, true
		}
	}
	return "", false
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
			dir := tree + "/" + specRootLeaf(tree, io.Discard) + "/changes/" + key
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

// changePlanDir is change_plan_dir (_change_plan_resolve_dir): the directory
// holding the change's canonical tasks.md.
func changePlanDir(env Env, stderr io.Writer, worktree, name, canonical string) (string, int) {
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "change-plan: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)
		return "", 1
	}
	specRoot := specRootLeaf(worktree, stderr)
	dir := worktree + "/" + specRoot + "/changes/" + name
	if isFile(dir + "/tasks.md") {
		return dir, 0
	}
	link := dir + "/link.md"
	if !isFile(link) && canonical != "" {
		absentDir := canonical + "/" + specRootLeaf(canonical, stderr) + "/changes/" + name
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
		canonDir := canonical + "/" + specRootLeaf(canonical, stderr) + "/changes/" + changeID
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
