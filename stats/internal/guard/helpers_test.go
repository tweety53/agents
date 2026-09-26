package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
)

// execFixtures backs every executable fixture with an inode already
// assessed. macOS assesses a newly written executable on its first exec --
// 0.16-0.31s measured on this machine, serialised machine-wide (twenty fresh
// scripts took 3.06s run concurrently, 3.33s one after another) -- so the
// ~210 executable fixtures this package writes cost ~30s of wall however
// parallel it runs. The assessment holds per inode, across a rewrite with
// other content. So each distinct body gets one master inode here, every
// fixture carrying that body is a hard link to it, and a master no fixture
// links any more is rewritten for the next new body: the cost is bounded by
// the distinct bodies alive at once. Every fixture still carries exactly the
// bytes and mode its test wrote; a test changes one only through writeExec,
// which unlinks it first.
//
// dir is the one path this package writes outside t.TempDir(): a master must
// outlive the test that first wrote its body, and a t.TempDir() is removed
// when its test ends. TestMain removes it after the run.
var execFixtures struct {
	sync.Mutex
	dir     string            // created by TestMain
	masters map[string]string // body -> master inode's path in dir
	n       int               // masters ever named
}

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "guard-exec-fixtures")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	execFixtures.dir, execFixtures.masters = dir, map[string]string{}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// The rows below pin the bash helpers' own outputs: scripts/lib/sha256-hex.sh
// (known SHA-256 vectors), scripts/lib/spec-root.sh (its four documented
// branches) and scripts/lib/change-plan.sh (every case of
// scripts/test-lib-change-plan.sh, named after its ok: label).

func TestSHA256Hex(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ in, want string }{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	} {
		if got := sha256Hex(tc.in); got != tc.want {
			t.Errorf("sha256Hex(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}

	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	nul := filepath.Join(dir, "nul")
	writeFile(t, empty, "")
	writeFile(t, nul, "a\x00b")
	for _, tc := range []struct{ path, want string }{
		{empty, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		// A NUL byte survives: the file's bytes are hashed, never a string copy.
		{nul, shasum(t, nul)},
	} {
		got, err := sha256HexFile(tc.path)
		if err != nil {
			t.Fatalf("sha256HexFile(%s): %v", tc.path, err)
		}
		if got != tc.want {
			t.Errorf("sha256HexFile(%s) = %s, want %s", tc.path, got, tc.want)
		}
	}
	if _, err := sha256HexFile(filepath.Join(dir, "absent")); err == nil {
		t.Error("sha256HexFile on an absent file: want an error")
	}
}

// shasum is the bash helper's own first tool, run on the real file, so the
// NUL-byte vector is the tool's answer rather than one typed in here.
func shasum(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("shasum", "-a", "256", path).Output()
	if err != nil {
		t.Fatalf("shasum: %v", err)
	}
	return strings.Fields(string(out))[0]
}

func TestSpecRoot(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, want, stderr string
		trees              []string
	}{
		{"neither tree is spectre", "spectre", "", nil},
		{"spectre only", "spectre", "", []string{"spectre"}},
		{"openspec only", "openspec", "", []string{"openspec"}},
		{"both: spectre wins and the other is reported", "spectre",
			"spec-root: DIR carries both spectre/changes/ and openspec/changes/ — using spectre/; every change under openspec/changes/ is invisible to this guard until that tree is moved\n",
			[]string{"spectre", "openspec"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, tree := range tc.trees {
				mkdir(t, dir+"/"+tree+"/changes")
			}
			var stderr bytes.Buffer
			if got := specRootLeaf(dir, &stderr); got != tc.want {
				t.Errorf("specRootLeaf = %q, want %q", got, tc.want)
			}
			if want := strings.ReplaceAll(tc.stderr, "DIR", dir); stderr.String() != want {
				t.Errorf("stderr = %q, want %q", stderr.String(), want)
			}
		})
	}
}

func TestChangePlan(t *testing.T) {
	t.Parallel()
	work := t.TempDir()
	// The store step must never reach the real flow CLI: this default stub
	// fails like an unreachable store, as the bash harness's DEFAULT_BIN does.
	defaultBin := work + "/default-bin"
	writeExec(t, defaultBin+"/flow", "#!/usr/bin/env bash\nexit 1\n")
	env := Env{Getenv: pathEnv(defaultBin)}

	satellite := func(dir, ref string) {
		mkdir(t, dir+"/spectre/changes/sat-change")
		writeFile(t, dir+"/spectre/changes/sat-change/link.md", "## Part of\n\n`"+ref+"`\n")
	}
	plan := func(tree, id string) string {
		d := tree + "/spectre/changes/" + id
		writeFile(t, d+"/tasks.md", "# plan\n\n- [ ] 1. do a thing\n")
		return d
	}
	peers := func(tree, line string) { writeFile(t, tree+"/spectre/peers", line+"\n") }

	// Case 1: a plain change.
	plain := work + "/case1-worktree"
	plainDir := plan(plain, "plain-change")
	// Case 2: a satellite with the canonical worktree passed.
	sat2, canon2 := work+"/case2-satellite", work+"/case2-canonical"
	satellite(sat2, "peerx:canon-change")
	canon2Dir := plan(canon2, "canon-change")
	// Case 3: a satellite resolving through peers.
	sat3, peer3 := work+"/case3-parent/sat-tree", work+"/case3-parent/peer-tree"
	satellite(sat3, "peery:canon-change")
	peers(sat3, "peery ../peer-tree")
	peer3Dir := plan(peer3, "canon-change")
	// Case 2b: canonical worktree passed but empty — never falls back to peers.
	sat2b, canon2b := work+"/case2b-parent/sat-tree", work+"/case2b-canonical"
	mkdir(t, canon2b+"/spectre/changes")
	satellite(sat2b, "peerw:canon-change")
	peers(sat2b, "peerw ../peer-tree")
	plan(work+"/case2b-parent/peer-tree", "canon-change")
	// Case 4: a peer declared but not checked out.
	sat4 := work + "/case4-parent/sat-tree"
	satellite(sat4, "ghost:canon-change")
	peers(sat4, "ghost ../not-checked-out")
	// Case 5: a link.md with `## Parts`, no `## Part of`.
	canon5 := work + "/case5-canonical"
	plan(canon5, "canon-change")
	writeFile(t, canon5+"/spectre/changes/canon-only-satellite-dir/link.md",
		"## Parts\n\n`peerz:some-part`\n\n## Merge order\n\n1. `.`\n2. `peerz`\n")
	// Case 5b: an empty `## Part of` must not read past the next heading.
	sat5b, peer5b := work+"/case5b-worktree", work+"/case5b-peer"
	writeFile(t, sat5b+"/spectre/changes/sat-change/link.md", "## Part of\n\n## Parts\n\n`peerz:some-part`\n")
	peers(sat5b, "peerz ../case5b-peer")
	plan(peer5b, "some-part")
	// Case 6a-6c: containment — change name, peer name, canonical change id.
	bad6 := work + "/case6-worktree"
	mkdir(t, bad6+"/spectre/changes")
	writeFile(t, bad6+"/secret/tasks.md", "# secret\n")
	sat6b := work + "/case6b-worktree"
	satellite(sat6b, "../escape:canon-change")
	peers(sat6b, "../escape ../elsewhere")
	plan(work+"/elsewhere", "canon-change")
	sat6c, canon6c := work+"/case6c-worktree", work+"/case6c-canonical"
	satellite(sat6c, "peerx:../escape")
	writeFile(t, canon6c+"/spectre/escape/tasks.md", "# escaped\n")
	// Case 10-10f: the absent-local-dir branch.
	sat10, canon10 := work+"/case10-no-dir", work+"/case10-canonical"
	mkdir(t, sat10+"/spectre/changes")
	canon10Dir := plan(canon10, "x-repo-change")
	sat10b := work + "/case10b-no-dir"
	mkdir(t, sat10b+"/spectre/changes")
	sat10c, canon10c := work+"/case10c-satellite", work+"/case10c-canonical"
	satellite(sat10c, "peerx:canon-change")
	plan(canon10c, "sat-change")
	sat10d, canon10d := work+"/case10d-empty-scaffold", work+"/case10d-canonical"
	mkdir(t, sat10d+"/spectre/changes/x-repo-change")
	canon10dDir := plan(canon10d, "x-repo-change")
	sat10e := work + "/case10e-empty-scaffold"
	mkdir(t, sat10e+"/spectre/changes/x-repo-change")
	sat10f, canon10f := work+"/case10f-parts-only-link", work+"/case10f-canonical"
	writeFile(t, sat10f+"/spectre/changes/sat-change/link.md", "## Parts\n\n`peerq:some-part`\n")
	plan(canon10f, "sat-change")
	// Case 11-11b: peers resolve beside the MAIN checkout, from a linked worktree.
	main11, peer11 := work+"/case11-main", work+"/case11-peer"
	gitRun(t, "", "init", "-q", main11)
	satellite(main11, "peerv:canon-change")
	peers(main11, "peerv ../case11-peer")
	plan(peer11, "canon-change")
	gitRun(t, main11, "add", "spectre")
	gitRun(t, main11, "commit", "-qm", "peers")
	wt11 := work + "/case11-wtroot/kan-11-worktree"
	mkdir(t, work+"/case11-wtroot")
	gitRun(t, main11, "worktree", "add", "-q", wt11, "-b", "case11-wtb")
	peer11Resolved, err := filepath.EvalSymlinks(peer11)
	if err != nil {
		t.Fatal(err)
	}
	// Case 11d: a separate-git-dir primary, and a linked worktree of it.
	sgd11, peer11d := work+"/case11d-sgd", work+"/case11d-peer"
	mkdir(t, work+"/case11d-elsewhere/gitdirs")
	gitRun(t, "", "init", "-q", "--separate-git-dir="+work+"/case11d-elsewhere/gitdirs/repo.git", sgd11)
	satellite(sgd11, "peeru:canon-change")
	peers(sgd11, "peeru ../case11d-peer")
	peer11dDir := plan(peer11d, "canon-change")
	gitRun(t, sgd11, "add", "spectre")
	gitRun(t, sgd11, "commit", "-qm", "x")
	mkdir(t, work+"/case11d-wtroot")
	gitRun(t, sgd11, "worktree", "add", "-q", work+"/case11d-wtroot/wt", "-b", "case11d-wtb")
	plan(work+"/case11d-elsewhere/case11d-peer", "canon-change") // the decoy

	// A nil wantErr means stderr is not asserted (the bash case discarded it).
	type row struct {
		name, fn                    string // fn: path, dir or ref
		worktree, change, canonical string
		want                        string
		rc                          int
		env                         Env
		wantErr                     *string
	}
	none := ""
	rows := []row{
		{name: "case 1: a plain change resolves at exit 0", fn: "path", worktree: plain, change: "plain-change", want: plainDir + "/tasks.md"},
		{name: "case 1 dir: a plain change's directory resolves at exit 0", fn: "dir", worktree: plain, change: "plain-change", want: plainDir},
		{name: "case 2: a satellite with the canonical worktree passed resolves at exit 0", fn: "path", worktree: sat2, change: "sat-change", canonical: canon2, want: canon2Dir + "/tasks.md"},
		{name: "case 2 dir: a satellite's directory resolves at exit 0", fn: "dir", worktree: sat2, change: "sat-change", canonical: canon2, want: canon2Dir},
		{name: "case 3: a satellite resolving through peers resolves at exit 0", fn: "path", worktree: sat3, change: "sat-change", want: peer3Dir + "/tasks.md"},
		{name: "case 2b: a canonical worktree passed with no plan there does not fall back to peers", fn: "path", worktree: sat2b, change: "sat-change", canonical: canon2b, rc: 1},
		{name: "case 4: a satellite whose peer is absent cannot resolve", fn: "path", worktree: sat4, change: "sat-change", rc: 1},
		{name: "case 5: a link.md with no ## Part of cannot resolve", fn: "path", worktree: canon5, change: "canon-only-satellite-dir", rc: 1},
		{name: "case 5b: an empty ## Part of does not read past the next heading", fn: "path", worktree: sat5b, change: "sat-change", rc: 1},
		{name: "case 6a: a change name outside the allowlist is rejected", fn: "path", worktree: bad6, change: "../../secret", rc: 1},
		{name: "case 6b: a peer name outside the allowlist is rejected", fn: "path", worktree: sat6b, change: "sat-change", rc: 1},
		{name: "case 6c: a canonical change id outside the allowlist is rejected", fn: "path", worktree: sat6c, change: "sat-change", canonical: canon6c, rc: 1},
		{name: "case 7: a local plan has no ref", fn: "ref", worktree: plain, change: "plain-change", rc: 1},
		{name: "case 8: a satellite's ref resolves at exit 0", fn: "ref", worktree: sat2, change: "sat-change", want: "peerx:canon-change"},
		{name: "case 8b: the peers-only satellite's ref resolves too", fn: "ref", worktree: sat3, change: "sat-change", want: "peery:canon-change"},
		{name: "case 9: a ## Parts-only link.md is not a satellite and has no ref", fn: "ref", worktree: canon5, change: "canon-only-satellite-dir", rc: 1},
		{name: "case 10: absent local dir resolves the canonical worktree's same-named plan", fn: "path", worktree: sat10, change: "x-repo-change", canonical: canon10, want: canon10Dir + "/tasks.md"},
		{name: "case 10 dir: the absent-dir resolution prints the canonical change directory", fn: "dir", worktree: sat10, change: "x-repo-change", canonical: canon10, want: canon10Dir},
		{name: "case 10b: absent local dir without a canonical worktree stays unresolvable", fn: "path", worktree: sat10b, change: "x-repo-change", rc: 1},
		{name: "case 10c: satellite dir with link.md and no tasks.md never takes the absent-dir branch", fn: "path", worktree: sat10c, change: "sat-change", canonical: canon10c, rc: 1},
		{name: "case 10d: empty local change dir with a canonical worktree resolves the same-named plan", fn: "path", worktree: sat10d, change: "x-repo-change", canonical: canon10d, want: canon10dDir + "/tasks.md"},
		{name: "case 10e: empty local change dir without a canonical worktree stays unresolvable", fn: "path", worktree: sat10e, change: "x-repo-change", rc: 1},
		{name: "case 10f: local dir with a link.md never takes the empty-dir branch", fn: "path", worktree: sat10f, change: "sat-change", canonical: canon10f, rc: 1},
		{name: "case 11: peer path resolves from a worktree whose parent holds no peer checkout", fn: "path", worktree: wt11, change: "sat-change", want: peer11Resolved + "/spectre/changes/canon-change/tasks.md"},
		{name: "case 11b: peer path still resolves from the main checkout itself", fn: "path", worktree: main11, change: "sat-change", want: peer11Resolved + "/spectre/changes/canon-change/tasks.md"},
		{name: "case 11c: unreadable git common dir falls back to worktree-relative resolution", fn: "path", worktree: sat3, change: "sat-change", want: peer3Dir + "/tasks.md"},
		{name: "case 11d: a linked worktree of a separate-git-dir primary refuses rather than resolving the gitdir-side decoy", fn: "path", worktree: work + "/case11d-wtroot/wt", change: "sat-change", rc: 1},
		{name: "case 11d: the separate-git-dir primary itself resolves the real peer", fn: "path", worktree: sgd11, change: "sat-change", want: peer11dDir + "/tasks.md"},
		{name: "case 6a stderr: the rejected change name is explained", fn: "path", worktree: bad6, change: "../../secret", rc: 1,
			wantErr: ptr("change-plan: change name '../../secret' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n")},
		{name: "case 6b stderr: the rejected peer name is explained", fn: "path", worktree: sat6b, change: "sat-change", rc: 1,
			wantErr: ptr("change-plan: peer name '../escape' in " + sat6b + "/spectre/changes/sat-change/link.md is not a plain peer name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n")},
		{name: "case 6c stderr: the rejected canonical change id is explained", fn: "path", worktree: sat6c, change: "sat-change", canonical: canon6c, rc: 1,
			wantErr: ptr("change-plan: canonical change id '../escape' in " + sat6c + "/spectre/changes/sat-change/link.md is not a plain change id — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n")},
	}

	// The store cases: a stub `flow` answering `state find` from a fixture,
	// logging every call. Each case gets its own stub directory.
	store := func(name, fixture string) (Env, string) {
		bin := work + "/store-" + name
		calls := bin + "/calls"
		// Paths named from the stub's own, so every store stub shares one of
		// two bodies (writeExec).
		body := "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"${0%/*}/calls\"\n"
		if fixture == "" {
			body += "echo 'flow: store unreachable' >&2\nexit 1\n"
		} else {
			writeFile(t, bin+"/fixture.json", fixture)
			body += "cat \"${0%/*}/fixture.json\"\n"
		}
		writeExec(t, bin+"/flow", body)
		return Env{Getenv: pathEnv(bin)}, calls
	}
	record := func(project, tree string) string {
		return `{"projectKey":"` + project + `","name":"x","worktrees":{"` + tree + `":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`
	}
	case11s, canon11s := work+"/case11s-no-dir", work+"/case11s-canonical"
	mkdir(t, case11s+"/spectre/changes")
	canon11sDir := plan(canon11s, "x-repo-plan")
	env11, calls11 := store("11", `{"source":"store","complete":true,"records":[{"projectKey":"proj-a","name":"x-repo-plan","worktrees":{"`+case11s+`":null,"`+canon11s+`":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}]}`)
	case12, canon12 := work+"/case12-satellite", work+"/case12-canonical"
	satellite(case12, "storepeer:canon-plan")
	canon12Dir := plan(canon12, "canon-plan")
	env12, calls12 := store("12", `{"records":[`+record("proj-a", canon12)+`]}`)
	case13, tree13a, tree13b := work+"/case13-no-dir", work+"/case13-tree-a", work+"/case13-tree-b"
	mkdir(t, case13+"/spectre/changes")
	plan(tree13a, "ambig-plan")
	plan(tree13b, "ambig-plan")
	env13, _ := store("13", `{"records":[`+record("proj-a", tree13a)+","+record("proj-b", tree13b)+`]}`)
	case14, canon14, peer14 := work+"/case14-satellite", work+"/case14-canonical-empty", work+"/case14-peer"
	satellite(case14, "skippeer:canon-change")
	peers(case14, "skippeer ../case14-peer")
	mkdir(t, canon14+"/spectre/changes")
	plan(peer14, "canon-change")
	env14, calls14 := store("14", `{"records":[`+record("proj-a", peer14)+`]}`)
	case14b, canon14b := work+"/case14b-no-dir", work+"/case14b-canonical-empty"
	mkdir(t, case14b+"/spectre/changes")
	mkdir(t, canon14b+"/spectre/changes")
	env14b, calls14b := store("14b", `{"records":[`+record("proj-a", work+"/case14b-elsewhere")+`]}`)
	case15 := work + "/case15-no-dir"
	mkdir(t, case15+"/spectre/changes")
	env15, _ := store("15", "")

	rows = append(rows,
		row{name: "store-resolves-absent-dir: the record's worktrees map resolves the plan's directory", fn: "dir", worktree: case11s, change: "x-repo-plan", want: canon11sDir, env: env11},
		row{name: "store-resolves-satellite-without-peers: the record under the link's canonical id resolves", fn: "path", worktree: case12, change: "sat-change", want: canon12Dir + "/tasks.md", env: env12},
		row{name: "store-ambiguity-refuses: two resolving projects refuse with the ambiguity code 3", fn: "dir", worktree: case13, change: "ambig-plan", rc: 3, env: env13,
			wantErr: ptr("change-plan: ambiguous state-record resolution for 'ambig-plan': project proj-a resolves to " + tree13a + "/spectre/changes/ambig-plan\n" +
				"change-plan: ambiguous state-record resolution for 'ambig-plan': project proj-b resolves to " + tree13b + "/spectre/changes/ambig-plan\n")},
		row{name: "store-skipped-when-canonical-arg: a supplied canonical worktree never falls back to the store", fn: "path", worktree: case14, change: "sat-change", canonical: canon14, rc: 1, env: env14},
		row{name: "store-skipped-when-canonical-arg: the absent-dir site never falls back to the store behind a supplied canonical worktree", fn: "path", worktree: case14b, change: "x-repo-change", canonical: canon14b, rc: 1, env: env14b},
		row{name: "store-unavailable-stays-unresolvable: an unreachable store cannot resolve", fn: "path", worktree: case15, change: "lonely-plan", rc: 1, env: env15, wantErr: &none},
	)

	// The rows share fixtures read-only; git fixtures above are built before
	// any row runs, so the subtests run in parallel with nothing to race on.
	t.Run("rows", func(t *testing.T) {
		for _, r := range rows {
			t.Run(r.name, func(t *testing.T) {
				t.Parallel()
				e := r.env
				if e.Getenv == nil {
					e = env
				}
				var stderr bytes.Buffer
				var got string
				var rc int
				switch r.fn {
				case "path":
					got, rc = changePlanPath(e, &stderr, r.worktree, r.change, r.canonical)
				case "dir":
					got, rc = changePlanDir(e, &stderr, r.worktree, r.change, r.canonical)
				case "ref":
					var ok bool
					got, ok = changePlanRef(&stderr, r.worktree, r.change)
					if !ok {
						rc = 1
					}
				}
				if got != r.want || rc != r.rc {
					t.Errorf("got (%q, %d), want (%q, %d); stderr=%q", got, rc, r.want, r.rc, stderr.String())
				}
				if r.wantErr != nil && stderr.String() != *r.wantErr {
					t.Errorf("stderr = %q, want %q", stderr.String(), *r.wantErr)
				}
			})
		}
	})

	for _, c := range []struct{ name, calls, want string }{
		{"store-resolves-absent-dir: the store was asked for the change's own name", calls11, "state find x-repo-plan\n"},
		{"store-resolves-satellite-without-peers: the store was asked under the link's canonical id", calls12, "state find canon-plan\n"},
		{"store-skipped-when-canonical-arg: the store was never consulted", calls14, ""},
		{"store-skipped-when-canonical-arg: the absent-dir site never consulted the store", calls14b, ""},
	} {
		got, _ := os.ReadFile(c.calls)
		if string(got) != c.want {
			t.Errorf("%s: calls = %q, want %q", c.name, got, c.want)
		}
	}
}

func ptr(s string) *string { return &s }

// pathEnv is an Env.Getenv whose PATH puts bin first, the way the bash
// harness prepends its stub directory; every other name reads the process.
func pathEnv(bin string) func(string) string {
	return func(k string) string {
		if k == "PATH" {
			return bin + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}
}

func mkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	// Unlinked first: a write through a writeExec fixture's hard link would
	// land in the shared master and change every sibling's bytes.
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeExec links path to the master inode carrying body (execFixtures).
func writeExec(t *testing.T, path, body string) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	execFixtures.Lock()
	defer execFixtures.Unlock()
	m, ok := execFixtures.masters[body]
	if !ok {
		m = idleMaster()
		if err := os.WriteFile(m, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(m, 0o755); err != nil {
			t.Fatal(err)
		}
		execFixtures.masters[body] = m
	}
	if err := os.Link(m, path); err != nil {
		t.Fatal(err)
	}
}

// fixtureRoot is a bash expression for the tree a fixture at rel sits in,
// derived from the script's own path: a reproducer is exec'd by its
// resolved absolute path, so $0 always carries a directory.
func fixtureRoot(rel string) string {
	return "${0%/*}/" + strings.Repeat("../", strings.Count(rel, "/"))
}

// idleMaster takes a master no fixture links any more, else names a new one.
func idleMaster() string {
	for body, m := range execFixtures.masters {
		var st syscall.Stat_t
		if syscall.Stat(m, &st) == nil && st.Nlink == 1 {
			delete(execFixtures.masters, body)
			return m
		}
	}
	execFixtures.n++
	return filepath.Join(execFixtures.dir, strconv.Itoa(execFixtures.n))
}

// fixtureGitEnv is what every fixture-building git runs under: the user's
// global config masked, as the bash harnesses' GIT_CONFIG_GLOBAL=/dev/null
// did, no `git maintenance run --auto` child spawned after each commit, and
// the files ref backend that tcfRepo.head()'s loose-ref read assumes, whatever
// init.defaultRefFormat says.
var fixtureGitEnv = []string{"GIT_CONFIG_GLOBAL=/dev/null",
	"GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=maintenance.auto", "GIT_CONFIG_VALUE_0=false",
	"GIT_DEFAULT_REF_FORMAT=files"}

// gitRun runs git under fixtureGitEnv with a fixed identity, as the bash
// harness's GIT_CONFIG_GLOBAL=/dev/null plus its `config user.*` do.
func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command("git", args...)
	cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestPlainNameRefusesEmpty pins the empty-name refusal plainChangeName took over
// from change-plan.sh's _change_plan_name_ok when the package's two copies of
// the allowlist became one: a satellite ref such as `:change` hands it an
// empty peer name, which must be refused rather than indexed.
func TestPlainNameRefusesEmpty(t *testing.T) {
	t.Parallel()
	if plainChangeName("") {
		t.Fatal(`plainChangeName("") = true, want false`)
	}
}
