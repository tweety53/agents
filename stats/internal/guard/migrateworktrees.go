package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// migrateWorktrees is scripts/migrate-worktrees.sh: that script's header is
// the contract. It moves every registered worktree under <main>/.worktrees/,
// the retired in-repo layout, of every <main> named, to
// <dirname main>/<basename main>-worktrees/
// (design.md: migrate-all-now, sibling-worktrees-layout) with `git worktree
// move`, skipping one check-worktree-processes.sh reports HELD, and rewrites
// the worktrees key of every flow state record that named a moved path.
func init() { Registry["migrate-worktrees"] = migrateWorktrees }

func migrateWorktrees(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "migrate-worktrees: "+format+"\n", a...)
		return 2
	}
	if len(args) == 0 {
		return refuse("usage: migrate-worktrees.sh <main-checkout> [<main-checkout>...]")
	}
	git := envGit(env)
	// Every repository is moved and its records read in one run: a
	// cross-repo change's record lives in one project and names another
	// repository's worktree, so the rename map spans them all.
	type mwRepo struct{ main, oldRoot, sibRoot string }
	var repos []mwRepo
	seen := map[string]bool{}
	for _, a := range args {
		// Physical form, as git reports every worktree path.
		main, err := filepath.EvalSymlinks(pcAbs(env, a))
		if err != nil || !isDir(main) || git("-C", main, "rev-parse", "--git-dir").Run() != nil {
			return refuse("%s is not a git repository", a)
		}
		if !seen[main] {
			seen[main] = true
			repos = append(repos, mwRepo{main, main + "/.worktrees", siblingRoot(main)})
		}
	}
	all := map[string][]string{}
	for _, r := range repos {
		porcelain, ok := capture(git("-C", r.main, "worktree", "list", "--porcelain"))
		if !ok {
			return refuse("cannot list the worktrees of %s", r.main)
		}
		all[r.main] = mwWorktrees(porcelain)
	}
	scriptDir, ok := guardSelfDir(env, stderr, "migrate-worktrees: ", "migrate-worktrees")
	if !ok {
		return 2
	}
	// Physical, as the worktree paths are: the guard may run from a worktree
	// it moves, and its scripts directory then moves with it.
	if p, err := filepath.EvalSymlinks(scriptDir); err == nil {
		scriptDir = p
	}
	flow, ok := lookPath(env, "flow")
	if !ok {
		return refuse("no flow binary on PATH — the state records naming a moved worktree could not be rewritten, so nothing is moved")
	}

	// Every record is read before anything moves: a store that cannot
	// enumerate them all would leave a moved worktree's record stale. The
	// record pass runs even with nothing left to move, so a re-run repairs a
	// record an earlier run moved the worktree of but failed to rewrite.
	var records []mwRecord
	for _, r := range repos {
		rs, msg := mwRecords(env, flow, r.main)
		if msg != "" {
			return refuse("%s", msg)
		}
		records = append(records, rs...)
	}

	failed := false
	for _, r := range repos {
		for _, wt := range all[r.main] {
			if !strings.HasPrefix(wt, r.oldRoot+"/") {
				continue
			}
			// Moving a worktree carries any worktree nested inside it, whose git
			// registration then names a path that no longer exists.
			if inner := mwNested(all[r.main], wt); inner != "" {
				failed = true
				fmt.Fprintf(stdout, "FAILED: %s — holds the worktree %s\n", wt, inner)
				continue
			}
			procs := exec.Command(scriptDir+"/check-worktree-processes.sh", wt)
			procs.Stderr = stderr
			verdict, err := procs.Output()
			switch v := string(verdict); {
			case err == nil && strings.HasPrefix(v, "HELD:"):
				_, _ = io.WriteString(stdout, v)
				continue
			case err != nil || !strings.HasPrefix(v, "CLEAR:"):
				failed = true
				fmt.Fprintf(stdout, "FAILED: %s — check-worktree-processes.sh could not answer\n", wt)
				continue
			}
			dest := r.sibRoot + "/" + strings.TrimPrefix(wt, r.oldRoot+"/")
			// `git worktree move` into an existing directory moves the worktree
			// inside it, as mv does, so an occupied destination is refused.
			if _, err := os.Lstat(dest); err == nil {
				failed = true
				fmt.Fprintf(stdout, "FAILED: %s — %s already exists\n", wt, dest)
				continue
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				failed = true
				fmt.Fprintf(stdout, "FAILED: %s — %s\n", wt, rcwFlat(err.Error()))
				continue
			}
			if out, err := git("-C", r.main, "worktree", "move", wt, dest).CombinedOutput(); err != nil {
				failed = true
				fmt.Fprintf(stdout, "FAILED: %s — git worktree move: %s\n", wt, rcwFlat(string(out)))
				continue
			}
			fmt.Fprintf(stdout, "MIGRATED: %s -> %s\n", wt, dest)
			if rel, ok := strings.CutPrefix(scriptDir, wt+"/"); ok {
				scriptDir = dest + "/" + rel
			}
		}
	}

	// The rename map comes from git, not from this run's moves: every
	// worktree registered under a sibling root answers for the path it
	// would have had under its repository's .worktrees/.
	moved := map[string]string{}
	for _, r := range repos {
		porcelain, ok := capture(git("-C", r.main, "worktree", "list", "--porcelain"))
		if !ok {
			fmt.Fprintf(stdout, "FAILED: %s — cannot list its worktrees after the moves\n", r.main)
			return 1
		}
		for _, w := range mwWorktrees(porcelain) {
			if rel, ok := strings.CutPrefix(w, r.sibRoot+"/"); ok {
				moved[r.oldRoot+"/"+rel] = w
			}
		}
	}

	for _, r := range records {
		if !mwRenames(r.worktrees, moved) {
			continue
		}
		// Rewritten from a fresh read, never the snapshot taken before the
		// moves: a write another session made in between is kept.
		fresh, msg := mwGet(env, flow, r.main, r.name)
		if msg != "" {
			failed = true
			fmt.Fprintf(stdout, "FAILED: %s — %s\n", r.name, msg)
			continue
		}
		if fresh == nil || !mwRenames(fresh.worktrees, moved) {
			continue
		}
		r = *fresh
		renamed := map[string]json.RawMessage{}
		for k, v := range r.worktrees {
			if dest, ok := moved[mwCanon(k)]; ok {
				k = dest
			}
			renamed[k] = v
		}
		r.body["worktrees"], _ = json.Marshal(renamed)
		body, _ := json.Marshal(r.body)
		set := exec.Command(flow, "state", "set", "-C", r.main, r.name)
		set.Dir, set.Stdin = env.Dir, bytes.NewReader(body)
		if out, err := set.CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "FAILED: %s — flow state set: %s\n", r.name, rcwFlat(string(out)))
		} else {
			_, _ = stderr.Write(out) // the CLI's fallback warning, if any
		}
	}

	for _, r := range repos {
		_ = os.Remove(r.oldRoot) // only when empty
	}
	if failed {
		return 1
	}
	return 0
}

type mwRecord struct {
	main      string // the main checkout whose project holds the record
	name      string
	body      map[string]json.RawMessage
	worktrees map[string]json.RawMessage
}

// mwRecords reads every state record `flow state list` names for main, as
// `flow state get` prints it. A synthetic record (a stage mark's bootstrap)
// is dropped: `state get` adds a "synthetic" field the store's closed PUT
// schema refuses, and no STARTED write has given it a worktree to rewrite.
// A non-empty message is a reason the records cannot be read in full.
func mwRecords(env Env, flow, main string) ([]mwRecord, string) {
	list := exec.Command(flow, "state", "list", "-C", main)
	list.Dir = env.Dir
	out, err := list.Output()
	var l struct {
		Source   string `json:"source"`
		Complete bool   `json:"complete"`
		Records  []struct {
			Name string `json:"name"`
		} `json:"records"`
	}
	if err != nil || json.Unmarshal(out, &l) != nil {
		return nil, "flow state list failed"
	}
	if !l.Complete {
		return nil, fmt.Sprintf("flow state list is partial (source %q) — a record naming a moved worktree could be missed", l.Source)
	}
	var records []mwRecord
	for _, e := range l.Records {
		r, msg := mwGet(env, flow, main, e.Name)
		if msg != "" {
			return nil, msg
		}
		if r != nil {
			records = append(records, *r)
		}
	}
	return records, ""
}

// mwGet reads one record as `flow state get` prints it: nil when it names no
// worktree, is synthetic or empty. A get answered from the on-disk fallback
// is refused: written back, it would overwrite the store with a stale copy.
func mwGet(env Env, flow, main, name string) (*mwRecord, string) {
	get := exec.Command(flow, "state", "get", "-C", main, name)
	var errb bytes.Buffer
	get.Dir, get.Stderr = env.Dir, &errb
	out, err := get.Output()
	if err != nil {
		return nil, fmt.Sprintf("flow state get %s failed", name)
	}
	if strings.Contains(errb.String(), "store unreachable") {
		return nil, fmt.Sprintf("flow state get %s read the local fallback, not the store", name)
	}
	r := mwRecord{main: main, name: name}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, "" // nothing recorded, nothing to rewrite
	}
	if json.Unmarshal(out, &r.body) != nil || r.body == nil {
		return nil, fmt.Sprintf("flow state get %s printed no JSON object", name)
	}
	if _, synthetic := r.body["synthetic"]; synthetic {
		return nil, ""
	}
	if raw, ok := r.body["worktrees"]; ok && json.Unmarshal(raw, &r.worktrees) != nil {
		return nil, fmt.Sprintf("flow state get %s: worktrees is not an object", name)
	}
	if len(r.worktrees) == 0 {
		return nil, ""
	}
	return &r, ""
}

// mwWorktrees is every worktree path `git worktree list --porcelain` names.
func mwWorktrees(porcelain string) []string {
	var all []string
	for _, line := range strings.Split(porcelain, "\n") {
		if w, ok := strings.CutPrefix(line, "worktree "); ok {
			all = append(all, w)
		}
	}
	return all
}

// mwNested is a registered worktree strictly inside wt, or "".
func mwNested(all []string, wt string) string {
	for _, w := range all {
		if strings.HasPrefix(w, wt+"/") {
			return w
		}
	}
	return ""
}

// mwRenames reports whether any of a record's worktree keys is renamed.
func mwRenames(worktrees map[string]json.RawMessage, moved map[string]string) bool {
	for k := range worktrees {
		if _, ok := moved[mwCanon(k)]; ok {
			return true
		}
	}
	return false
}

// mwCanon is p with its nearest existing ancestor resolved physically, as git
// reports worktree paths: a record keyed through a symlinked parent still
// matches, and so does an old path that no longer exists after its move.
func mwCanon(p string) string {
	rest := ""
	for d := p; ; d = filepath.Dir(d) {
		if r, err := filepath.EvalSymlinks(d); err == nil {
			return filepath.Join(r, rest)
		}
		if filepath.Dir(d) == d {
			return p
		}
		rest = filepath.Join(filepath.Base(d), rest)
	}
}
