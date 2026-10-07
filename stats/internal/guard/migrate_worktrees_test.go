package guard

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mwFx is a main checkout with two worktrees in the retired in-repo layout,
// <repo>/.worktrees/{free,held}, and a fake `flow` on PATH serving the state
// records the guard rewrites. The fake prints each record in the exact shape
// `flow state get` prints the store's (api.toDTO: projectKey, name, state,
// branch, worktrees, jiraIssue, repos, updatedAt, updatedBy; `synthetic`
// added by markSyntheticIfNeeded), and records every argv it receives plus
// each `state set` body -- the real CLI is never run against the real store.
type mwFx struct {
	dir, repo, free, held, sibling, fake, mergeBase string
	self                                            string // FLOW_GUARD_SELF; "" is the real scripts/ dir
	g                                               fxGit
}

func mwNewFx(t *testing.T) *mwFx {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fx := &mwFx{dir: dir, repo: dir + "/repo", fake: dir + "/fake", sibling: dir + "/repo-worktrees"}
	fx.free, fx.held = fx.repo+"/.worktrees/free", fx.repo+"/.worktrees/held"
	fx.g.git("", "init", "-q", "-b", "main", fx.repo)
	fx.g.write(fx.repo+"/.gitignore", ".worktrees/")
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.mergeBase = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.git(fx.repo, "worktree", "add", "-q", fx.free, "-b", "spectre/free")
	fx.g.git(fx.repo, "worktree", "add", "-q", fx.held, "-b", "spectre/held")
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	mkdir(t, fx.fake+"/records")
	mkdir(t, fx.fake+"/set")
	mkdir(t, fx.fake+"/bin")
	fx.list(t, true)
	fx.record(t, "kan-1-free", `{"projectKey":"repo","name":"kan-1-free","state":"IN_PROGRESS","branch":"spectre/free",`+
		`"worktrees":{"`+fx.free+`":"`+fx.mergeBase+`","/elsewhere/peer":null},"jiraIssue":"KAN-1",`+
		`"repos":[{"repoRoot":"`+fx.repo+`","mergeBase":"`+fx.mergeBase+`"}],`+
		`"updatedAt":"2026-10-07T10:00:00.123456789Z","updatedBy":"/flow"}`)
	fx.record(t, "kan-2-held", `{"projectKey":"repo","name":"kan-2-held","state":"IN_PROGRESS",`+
		`"worktrees":{"`+fx.held+`":"`+fx.mergeBase+`"},"updatedAt":"2026-10-07T10:00:00Z","updatedBy":"/flow"}`)
	fx.record(t, "kan-3-synthetic", `{"name":"kan-3-synthetic","projectKey":"repo","state":"STARTED","synthetic":true,`+
		`"updatedAt":"2026-10-07T10:00:00Z","updatedBy":"flow stage begin (synthetic)","worktrees":{"`+fx.free+`":null}}`)
	fx.record(t, "kan-4-none", `{"projectKey":"repo","name":"kan-4-none","state":"FINISHED","updatedAt":"2026-10-07T10:00:00Z","updatedBy":"/flow"}`)
	fx.flowBin(t, "0")
	return fx
}

// list writes the `flow state list` answer: complete from the store, or the
// partial on-disk fallback.
func (fx *mwFx) list(t *testing.T, complete bool) {
	t.Helper()
	out := `{"source":"store","complete":true,"records":[{"name":"kan-1-free","state":"IN_PROGRESS"},` +
		`{"name":"kan-2-held","state":"IN_PROGRESS"},{"name":"kan-3-synthetic","state":"STARTED"},{"name":"kan-4-none","state":"FINISHED"}]}`
	if !complete {
		out = `{"source":"fallback","complete":false,"records":[]}`
	}
	writeFile(t, fx.fake+"/list.json", out+"\n")
}

func (fx *mwFx) record(t *testing.T, name, body string) {
	t.Helper()
	writeFile(t, fx.fake+"/records/"+name+".json", body)
}

// flowBin writes the fake CLI; setRC is what `state set` exits, and a
// successful set becomes the record the next `state get` serves, as the
// store's does.
func (fx *mwFx) flowBin(t *testing.T, setRC string) {
	t.Helper()
	body := `#!/bin/bash
F='` + fx.fake + `'
echo "$*" >> "$F/args.log"
# another project's records, when a test serves one, live under by/<basename of -C>
[ "$3" = -C ] && [ -d "$F/by/${4##*/}" ] && F="$F/by/${4##*/}"
case "$1 $2" in
"state list") cat "$F/list.json" ;;
"state get") [ -e "$F/fallback" ] && echo "⚠ flow: store unreachable — read local fallback" >&2
  cat "$F/records/${!#}.json" || exit 1
  # a .next record is another session's write landing after this read
  if [ -e "$F/records/${!#}.next" ]; then mv "$F/records/${!#}.next" "$F/records/${!#}.json"; fi ;;
"state set") cat > "$F/set/${!#}.json"; rc=` + setRC + `
  [ "$rc" = 0 ] && cp "$F/set/${!#}.json" "$F/records/${!#}.json"
  exit $rc ;;
*) echo "fake flow: unexpected $*" >&2; exit 2 ;;
esac
`
	if err := os.WriteFile(fx.fake+"/bin/flow", []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

// hold starts a process whose cwd is inside wt, as a running stack would.
func (fx *mwFx) hold(t *testing.T, wt string) *exec.Cmd {
	t.Helper()
	held := exec.Command("sleep", "60")
	held.Dir = wt
	if err := held.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Process.Kill(); _ = held.Wait() })
	return held
}

// run runs migrate-worktrees in-process with the real
// check-worktree-processes.sh beside it, as the shim exports FLOW_GUARD_SELF,
// and the fake flow first on PATH.
func (fx *mwFx) run(t *testing.T, path string, args ...string) (int, string, string) {
	t.Helper()
	self := tcfScriptsDir(t) + "/migrate-worktrees.sh"
	if fx.self != "" {
		self = fx.self
	}
	env := Env{Dir: fx.dir, Getenv: func(k string) string {
		switch k {
		case "FLOW_GUARD_SELF":
			return self
		case "PATH":
			return path
		}
		return os.Getenv(k)
	}}
	var out, errb bytes.Buffer
	if args == nil {
		args = []string{fx.repo}
	}
	code := migrateWorktrees(args, env, &out, &errb)
	return code, out.String(), errb.String()
}

func (fx *mwFx) path() string { return fx.fake + "/bin:" + os.Getenv("PATH") }

func (fx *mwFx) sets(t *testing.T) []string {
	t.Helper()
	b, _ := os.ReadFile(fx.fake + "/args.log")
	var got []string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "state set ") {
			got = append(got, l)
		}
	}
	return got
}

func (fx *mwFx) worktreeList(t *testing.T) string {
	t.Helper()
	out := fx.g.git(fx.repo, "worktree", "list", "--porcelain")
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	return out
}

func TestMigrateWorktrees(t *testing.T) {
	t.Parallel()

	t.Run("moves the free worktree, skips the held one, rewrites the record", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		held := fx.hold(t, fx.held)
		ob, _ := os.ReadFile(fx.fake + "/records/kan-1-free.json")
		code, out, errb := fx.run(t, fx.path())
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		newFree := fx.sibling + "/free"
		if l := rcwLines(out, "MIGRATED: "); len(l) != 1 || l[0] != "MIGRATED: "+fx.free+" -> "+newFree {
			t.Errorf("MIGRATED lines %q, want one for %s", l, fx.free)
		}
		if l := rcwLines(out, "HELD: "); len(l) != 1 || !strings.HasPrefix(l[0], "HELD: "+fx.held+" — ") {
			t.Errorf("HELD lines %q, want one for %s", l, fx.held)
		}
		list := fx.worktreeList(t)
		for _, want := range []string{"worktree " + newFree + "\n", "worktree " + fx.held + "\n"} {
			if !strings.Contains(list+"\n", want) {
				t.Errorf("git worktree list lacks %q:\n%s", want, list)
			}
		}
		if strings.Contains(list, "worktree "+fx.free+"\n") || rcwExists(fx.free) {
			t.Errorf("%s is still registered or on disk", fx.free)
		}
		if !rcwExists(fx.held) || !rcwExists(fx.repo+"/.worktrees") {
			t.Error("the held worktree or its .worktrees parent was removed")
		}

		// Exactly one record rewritten: the synthetic one is dropped, the
		// held one and the one naming no worktree have nothing to rewrite.
		if s := fx.sets(t); len(s) != 1 || s[0] != "state set -C "+fx.repo+" kan-1-free" {
			t.Fatalf("state set calls %q, want one for kan-1-free", s)
		}
		var orig, got map[string]json.RawMessage
		gb, _ := os.ReadFile(fx.fake + "/set/kan-1-free.json")
		if json.Unmarshal(ob, &orig) != nil || json.Unmarshal(gb, &got) != nil {
			t.Fatalf("set body is not a JSON object: %s", gb)
		}
		if len(got) != len(orig) {
			t.Errorf("set body has %d fields, the record %d: %s", len(got), len(orig), gb)
		}
		for k, v := range orig {
			if k != "worktrees" && string(got[k]) != string(v) {
				t.Errorf("field %s: set %s, record %s", k, got[k], v)
			}
		}
		var wts map[string]*string
		if err := json.Unmarshal(got["worktrees"], &wts); err != nil {
			t.Fatal(err)
		}
		if len(wts) != 2 || wts[newFree] == nil || *wts[newFree] != fx.mergeBase {
			t.Errorf("worktrees %s, want %s -> %s plus the peer", got["worktrees"], newFree, fx.mergeBase)
		}
		if v, ok := wts["/elsewhere/peer"]; !ok || v != nil {
			t.Errorf("the peer worktree's null merge base was not kept: %s", got["worktrees"])
		}

		// Released, a re-run moves it, rewrites its record and removes the
		// emptied .worktrees.
		_ = held.Process.Kill()
		_ = held.Wait()
		code, out, errb = fx.run(t, fx.path())
		if code != 0 {
			t.Fatalf("re-run exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "MIGRATED: "); len(l) != 1 || l[0] != "MIGRATED: "+fx.held+" -> "+fx.sibling+"/held" {
			t.Errorf("re-run MIGRATED lines %q", l)
		}
		if rcwExists(fx.repo + "/.worktrees") {
			t.Error("the emptied .worktrees was left behind")
		}
		if s := fx.sets(t); len(s) != 2 || s[1] != "state set -C "+fx.repo+" kan-2-held" {
			t.Errorf("state set calls %q, want kan-2-held's second", s)
		}
	})

	t.Run("a cross-repo record is rewritten for every repository named", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		// Repository b's worktree is recorded only in the primary project's
		// record, as a cross-repo change's is; b's own project lists none.
		b, bwt := fx.dir+"/b", fx.dir+"/b/.worktrees/x"
		fx.g.git("", "init", "-q", "-b", "main", b)
		fx.g.git(b, "commit", "-q", "--allow-empty", "-m", "base")
		fx.g.git(b, "worktree", "add", "-q", bwt, "-b", "spectre/x")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		mkdir(t, fx.fake+"/by/b/set")
		mkdir(t, fx.fake+"/by/b/records")
		writeFile(t, fx.fake+"/by/b/list.json", `{"source":"store","complete":true,"records":[]}`+"\n")
		fx.record(t, "kan-1-free", `{"projectKey":"repo","name":"kan-1-free","state":"IN_PROGRESS",`+
			`"worktrees":{"`+fx.free+`":"`+fx.mergeBase+`","`+bwt+`":null},"updatedAt":"2026-10-07T10:00:00Z","updatedBy":"/flow"}`)
		code, out, errb := fx.run(t, fx.path(), b, fx.repo)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "MIGRATED: "+bwt+" -> "+fx.dir+"/b-worktrees/x"); len(l) != 1 {
			t.Errorf("b's worktree did not move:\n%s", out)
		}
		gb, _ := os.ReadFile(fx.fake + "/records/kan-1-free.json")
		if !strings.Contains(string(gb), `"`+fx.dir+`/b-worktrees/x":null`) || strings.Contains(string(gb), bwt) ||
			!strings.Contains(string(gb), `"`+fx.sibling+`/free":"`+fx.mergeBase+`"`) {
			t.Errorf("kan-1-free keeps an old key: %s", gb)
		}
	})

	t.Run("the guard's own scripts directory moves with the worktree it sits in", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		// The shim's SCRIPT_DIR inside a migrated worktree, reached through a
		// symlink as a logical cwd would be; a second repository's worktree,
		// named after it, is checked once that worktree has moved.
		mkdir(t, fx.free+"/scripts")
		writeFile(t, fx.free+"/scripts/check-worktree-processes.sh", "#!/bin/sh\necho \"CLEAR: $1\"\n")
		if err := os.Chmod(fx.free+"/scripts/check-worktree-processes.sh", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(fx.free, fx.dir+"/link"); err != nil {
			t.Fatal(err)
		}
		fx.self = fx.dir + "/link/scripts/migrate-worktrees.sh"
		b, bwt := fx.dir+"/b", fx.dir+"/b/.worktrees/x"
		fx.g.git("", "init", "-q", "-b", "main", b)
		fx.g.git(b, "commit", "-q", "--allow-empty", "-m", "base")
		fx.g.git(b, "worktree", "add", "-q", bwt, "-b", "spectre/x")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		mkdir(t, fx.fake+"/by/b/records")
		writeFile(t, fx.fake+"/by/b/list.json", `{"source":"store","complete":true,"records":[]}`+"\n")
		code, out, errb := fx.run(t, fx.path(), fx.repo, b)
		if code != 0 || strings.Contains(out, "FAILED: ") {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		for _, want := range []string{fx.free + " -> " + fx.sibling + "/free", fx.held + " -> " + fx.sibling + "/held", bwt + " -> " + fx.dir + "/b-worktrees/x"} {
			if len(rcwLines(out, "MIGRATED: "+want)) != 1 {
				t.Errorf("no MIGRATED: %s\n%s", want, out)
			}
		}
	})

	t.Run("nothing to move", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		fx.g.git(fx.repo, "worktree", "remove", "--force", fx.free)
		fx.g.git(fx.repo, "worktree", "remove", "--force", fx.held)
		code, out, errb := fx.run(t, fx.path())
		if code != 0 || out != "" || len(fx.sets(t)) != 0 {
			t.Fatalf("exit %d, sets %q\nstdout:\n%s\nstderr:\n%s", code, fx.sets(t), out, errb)
		}
	})

	t.Run("a move that fails exits 1 and leaves the worktree", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		mkdir(t, fx.sibling+"/free/occupied")
		code, out, _ := fx.run(t, fx.path())
		if code != 1 || len(rcwLines(out, "FAILED: "+fx.free+" — ")) != 1 {
			t.Fatalf("exit %d, stdout %q; want 1 and a FAILED line for %s", code, out, fx.free)
		}
		if !rcwExists(fx.free) || len(rcwLines(out, "MIGRATED: ")) != 1 {
			t.Errorf("the failed worktree moved, or the other did not:\n%s", out)
		}
		for _, s := range fx.sets(t) {
			if strings.HasSuffix(s, " kan-1-free") {
				t.Errorf("a record was rewritten for a worktree that did not move: %q", s)
			}
		}
	})

	t.Run("a record rewrite that fails exits 1", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		fx.flowBin(t, "1")
		code, out, _ := fx.run(t, fx.path())
		if code != 1 || len(rcwLines(out, "FAILED: kan-1-free — ")) != 1 {
			t.Fatalf("exit %d, stdout %q; want 1 and a FAILED line for kan-1-free", code, out)
		}
	})

	t.Run("a re-run repairs a record whose rewrite failed", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		fx.flowBin(t, "1")
		if code, out, _ := fx.run(t, fx.path()); code != 1 || len(rcwLines(out, "MIGRATED: ")) != 2 {
			t.Fatalf("first run exit %d, stdout %q; want 1 with both moved", code, out)
		}
		fx.flowBin(t, "0")
		code, out, errb := fx.run(t, fx.path())
		if code != 0 || len(rcwLines(out, "MIGRATED: ")) != 0 {
			t.Fatalf("re-run exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		gb, _ := os.ReadFile(fx.fake + "/records/kan-1-free.json")
		if !strings.Contains(string(gb), `"`+fx.sibling+`/free":"`+fx.mergeBase+`"`) || strings.Contains(string(gb), fx.free) {
			t.Errorf("kan-1-free not repaired on the re-run: %s", gb)
		}
	})

	t.Run("the write starts from a fresh read, keeping a concurrent write", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		writeFile(t, fx.fake+"/records/kan-1-free.next", `{"projectKey":"repo","name":"kan-1-free","state":"IN_PROGRESS",`+
			`"worktrees":{"`+fx.free+`":"`+fx.mergeBase+`","/elsewhere/added":null},"updatedAt":"2026-10-07T10:05:00Z","updatedBy":"/flow"}`)
		if code, out, errb := fx.run(t, fx.path()); code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		gb, _ := os.ReadFile(fx.fake + "/set/kan-1-free.json")
		if !strings.Contains(string(gb), `"/elsewhere/added":null`) || !strings.Contains(string(gb), `"`+fx.sibling+`/free"`) {
			t.Errorf("the set dropped the write made after the first read: %s", gb)
		}
	})

	t.Run("a failed state get cannot answer, and nothing moves", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		if err := os.Remove(fx.fake + "/records/kan-2-held.json"); err != nil {
			t.Fatal(err)
		}
		code, out, errb := fx.run(t, fx.path())
		if code != 2 || out != "" || !strings.Contains(errb, "flow state get kan-2-held failed") {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 2, nothing, the failed get", code, out, errb)
		}
		if !rcwExists(fx.free) || !rcwExists(fx.held) {
			t.Error("a worktree moved although a record could not be read")
		}
	})

	t.Run("a get answered from the local fallback cannot answer, and nothing moves", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		writeFile(t, fx.fake+"/fallback", "")
		code, out, errb := fx.run(t, fx.path())
		if code != 2 || out != "" || !strings.Contains(errb, "local fallback") {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 2, nothing, a fallback reason", code, out, errb)
		}
		if !rcwExists(fx.free) {
			t.Error("a worktree moved on a fallback read")
		}
	})

	t.Run("a process check that cannot answer leaves the worktree", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		mkdir(t, fx.dir+"/self")
		if err := os.WriteFile(fx.dir+"/self/check-worktree-processes.sh", []byte("#!/bin/sh\nexit 2\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		fx.self = fx.dir + "/self/migrate-worktrees.sh"
		code, out, _ := fx.run(t, fx.path())
		if code != 1 || len(rcwLines(out, "FAILED: "+fx.free+" — check-worktree-processes.sh could not answer")) != 1 {
			t.Fatalf("exit %d, stdout %q; want 1 and a FAILED line for %s", code, out, fx.free)
		}
		if !rcwExists(fx.free) || !rcwExists(fx.held) || len(fx.sets(t)) != 0 {
			t.Errorf("a worktree moved or a record was rewritten: sets %q", fx.sets(t))
		}
	})

	t.Run("a git worktree move that fails exits 1 and leaves the worktree", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		fx.g.git(fx.repo, "worktree", "lock", fx.free)
		code, out, _ := fx.run(t, fx.path())
		if code != 1 || len(rcwLines(out, "FAILED: "+fx.free+" — git worktree move: ")) != 1 {
			t.Fatalf("exit %d, stdout %q; want 1 and a git worktree move FAILED line", code, out)
		}
		if !rcwExists(fx.free) || rcwExists(fx.sibling+"/free") {
			t.Error("the locked worktree moved")
		}
	})

	t.Run("a nested path keeps its relative path", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		nested := fx.repo + "/.worktrees/grp/x"
		fx.g.git(fx.repo, "worktree", "add", "-q", nested, "-b", "spectre/x")
		code, out, errb := fx.run(t, fx.path())
		if code != 0 || len(rcwLines(out, "MIGRATED: "+nested+" -> "+fx.sibling+"/grp/x")) != 1 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
	})

	t.Run("a worktree holding another worktree is refused", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		inner := fx.free + "/inner"
		fx.g.git(fx.repo, "worktree", "add", "-q", inner, "-b", "spectre/inner")
		code, out, _ := fx.run(t, fx.path())
		if code != 1 || len(rcwLines(out, "FAILED: "+fx.free+" — holds the worktree "+inner)) != 1 {
			t.Fatalf("exit %d, stdout %q; want 1 and a FAILED line for %s", code, out, fx.free)
		}
		if strings.Contains(fx.worktreeList(t), "prunable") {
			t.Errorf("a worktree was orphaned:\n%s", fx.worktreeList(t))
		}
	})

	t.Run("a record keyed through a symlinked parent is rewritten", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		link := fx.dir + "/link"
		if err := os.Symlink(fx.dir, link); err != nil {
			t.Fatal(err)
		}
		logical := link + "/repo/.worktrees/free"
		fx.record(t, "kan-1-free", `{"projectKey":"repo","name":"kan-1-free","state":"IN_PROGRESS","worktrees":{"`+logical+`":"`+fx.mergeBase+`"},"updatedAt":"2026-10-07T10:00:00Z","updatedBy":"/flow"}`)
		code, out, errb := fx.run(t, fx.path())
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		gb, _ := os.ReadFile(fx.fake + "/records/kan-1-free.json")
		if !strings.Contains(string(gb), `"`+fx.sibling+`/free":"`+fx.mergeBase+`"`) {
			t.Errorf("kan-1-free keyed by %s was not rewritten: %s", logical, gb)
		}
	})

	t.Run("a partial state list cannot answer, and nothing moves", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		fx.list(t, false)
		code, out, errb := fx.run(t, fx.path())
		if code != 2 || out != "" || !strings.Contains(errb, "partial") {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 2, nothing, a partial-list reason", code, out, errb)
		}
		if !rcwExists(fx.free) || !rcwExists(fx.held) {
			t.Error("a worktree moved although the records could not be read")
		}
	})

	t.Run("cannot answer", func(t *testing.T) {
		t.Parallel()
		fx := mwNewFx(t)
		noFlow := filepath.Dir(fixtureGit) + ":/bin:/usr/bin"
		for name, c := range map[string]struct {
			path string
			args []string
			want string
		}{
			"usage":      {fx.path(), []string{}, "usage: migrate-worktrees.sh <main-checkout> [<main-checkout>...]"},
			"not a repo": {fx.path(), []string{fx.fake}, "is not a git repository"},
			"no flow":    {noFlow, nil, "no flow binary on PATH"},
		} {
			code, out, errb := fx.run(t, c.path, c.args...)
			if code != 2 || out != "" || !strings.Contains(errb, c.want) {
				t.Errorf("%s: exit %d, stdout %q, stderr %q; want 2, nothing, %q", name, code, out, errb, c.want)
			}
		}
		if !rcwExists(fx.free) || !rcwExists(fx.held) {
			t.Error("a worktree moved on a cannot-answer path")
		}
	})
}
