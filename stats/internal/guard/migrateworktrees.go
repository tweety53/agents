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

// The contract below moved here from the header of the deleted
// scripts/migrate-worktrees.sh shim (KAN-1058).
//
// flow-guard migrate-worktrees — move every registered worktree under
// <main-checkout>/.worktrees/, the retired in-repo layout, of every
// repository named, to the sibling layout
// <dirname main>/<basename main>-worktrees/<name> (kan-916 design.md's
// migrate-all-now and sibling-worktrees-layout decisions), and rewrite every
// flow state record that named a moved path.
//
// Usage: flow-guard migrate-worktrees <main-checkout> [<main-checkout>...]
//
// Name every repository at once: a cross-repo change's record lives in one
// project and names another repository's worktree, which only a run that
// moved both and reads every project's records can rewrite. A repository
// named twice is migrated once.
//
// Every state record of every repository named is read first —
// `flow state list -C <main>` per repository, then `flow state get` per
// change; a synthetic record (a stage mark's bootstrap)
// is dropped, never written back; a get answered from the local fallback
// ("store unreachable") is refused. Then, per worktree under .worktrees/:
//   one holding another registered worktree is refused (FAILED), since the
//   move would carry the inner one away from its registration;
//   flow-guard check-worktree-processes says HELD → its HELD line is printed and the
//   worktree is left in place; otherwise `git worktree move <old> <new>`.
// The renames are then read from every repository's git — every worktree
// registered under a sibling root answers for its old .worktrees/ path,
// whichever run moved it —
// and each record whose `worktrees` map names a renamed path (keys compared
// physically, so a key through a symlinked parent matches) is re-read and
// written back with `flow state set -C <main> <name>`: every other field and
// every other key unchanged, the moved key renamed with its merge-base value
// kept, <main> being the repository whose project holds the record. Finally
// each <main>/.worktrees is removed when it is empty. Re-running
// after a HELD worktree is released moves it; re-running after a failed
// record rewrite repairs the record.
//
// Prints, per worktree or record:
//   MIGRATED: <old> -> <new>
//   HELD: <old> — <pid> <cwd>[; ...]        flow-guard check-worktree-processes's line
//   FAILED: <old> — <reason>                the worktree holds another one, the
//                                           process check failed, the destination
//                                           already exists, or the move failed
//   FAILED: <change> — flow state set: <output>
//   FAILED: <change> — <reason>             the re-read before the write failed
//
// Exit 0 when every worktree was moved or held (nothing to move included),
// 1 when a move or a record rewrite failed, 2 with NOTHING on stdout when it
// cannot answer — no argument, a <main-checkout> that is not a git
// repository, no `flow` on PATH, a `flow state list` that fails or is partial
// (source fallback), or a `flow state get` that fails or reads the local
// fallback. Nothing is moved on
// exit 2.

// migrateWorktrees is `flow-guard migrate-worktrees`: the comment above is the contract. It moves every registered worktree under <main>/.worktrees/,
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
		return refuse("usage: flow-guard migrate-worktrees <main-checkout> [<main-checkout>...]")
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
			procs := siblingCmd(env, "FLOW_GUARD_CHECK_WORKTREE_PROCESSES", "check-worktree-processes", wt)
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
