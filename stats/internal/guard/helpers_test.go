package guard

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
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
	// runSibling execs this very executable with a guard's name first: in a
	// test binary that is the test binary, so it dispatches like flow-guard.
	if len(os.Args) > 1 {
		if fn, ok := Registry[os.Args[1]]; ok {
			cwd, _ := os.Getwd()
			os.Exit(fn(os.Args[2:], Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: cwd, Stdin: os.Stdin}, os.Stdout, os.Stderr))
		}
	}
	os.Setenv("FLOW_GUARD_TELEMETRY", "off")
	// A real `flow` on PATH would read the dev store, so a test would follow
	// whatever the operator has stored; a dead records address makes any
	// such read fail fast. A test wanting an answer puts its own fake `flow`
	// on PATH.
	os.Setenv("FLOW_RECORDS_ADDR", "http://127.0.0.1:1")
	// A test binary started with SIGINT or SIGHUP ignored -- a `&` job of a
	// non-interactive shell, or nohup -- hands that ignore to every child it
	// execs, so the signal tests watch a child that cannot die of the signal.
	// Catching and dropping the signal keeps this process as deaf as it
	// started, while exec resets a caught signal to its default in each child.
	for _, sig := range []os.Signal{syscall.SIGINT, syscall.SIGHUP} {
		if signal.Ignored(sig) {
			signal.Notify(make(chan os.Signal, 1), sig)
		}
	}
	dir, err := os.MkdirTemp("", "guard-exec-fixtures")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	execFixtures.dir, execFixtures.masters = dir, map[string]string{}
	// Every allowed-pairs reader runs `flow policy get` (policyTable). A
	// real `flow` would answer from the dev store, or its cache; a `flow`
	// first on PATH answering testPolicy -- the seeded pairs -- and handing
	// every other call to the next `flow` on PATH keeps every test on one
	// known table. A test wanting another answer puts its own fake `flow`
	// on PATH ahead of it.
	policyBin := filepath.Join(dir, "policy-bin")
	if err := os.MkdirAll(policyBin, 0o755); err == nil {
		err = os.WriteFile(filepath.Join(policyBin, "flow"), []byte("#!/bin/sh\n"+policyStub+"PATH=${PATH#*:} exec flow \"$@\"\n"), 0o755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	os.Setenv("PATH", policyBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	// Every sibling a guard under test runs is the flow-guard these sources
	// build, as in production; TestRunSiblingSelfExec covers the default.
	bin, err := guardCacheBuilt()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	siblingExe = bin
	// Every guard a test runs through scripts/flow-guard builds into, and
	// finds in, the cache above — never the operator's.
	os.Setenv("FLOW_GUARD_CACHE_DIR", filepath.Dir(filepath.Dir(bin)))
	repo, _ := filepath.Abs("../../..")
	before, watched := treeStatus(repo)
	code := m.Run()
	_ = os.RemoveAll(dir)
	if watched {
		if leak := treeLeak(repo, before); leak != "" {
			fmt.Fprint(os.Stderr, leak)
			code = 1
		}
	}
	os.Exit(code)
}

// treeStatus is the repository's `git status --porcelain`; false when it
// cannot be read (no git work tree), which skips the tree gate.
func treeStatus(repo string) (string, bool) {
	out, err := exec.Command("git", "-C", repo, "status", "--porcelain", "--untracked-files=normal").Output()
	return string(out), err == nil
}

// treeLeak is the tree gate of the deleted run-guard-tests (kan-584,
// kan-1051 F5): a suite that leaves the repository's status lines changed --
// a test or a reproducer writing into the checkout -- fails, so the leak is
// caught by the run that made it, never by the next guard reading the tree.
// It compares status lines, not content: an edit restored inside the run, or
// one to a file already modified or untracked, is not caught. It returns the
// failure message, or "" when the status is unchanged.
func treeLeak(repo, before string) string {
	after, ok := treeStatus(repo)
	if !ok || after == before {
		return ""
	}
	return fmt.Sprintf("guard tests changed the tree of %s:\nbefore:\n%safter:\n%s", repo, before, after)
}

// The rows below pin the helpers' own outputs: sha256.go (known SHA-256
// vectors, the outputs the retired scripts/lib/sha256-hex.sh produced), specroot.go (the retired scripts/lib/spec-root.sh's four documented
// branches) and changeplan.go (every case of its deleted bash twin's harness
// (KAN-887), named after its ok: label).

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
		{"spectre only", "spectre", "", []string{"spectre/changes"}},
		{"openspec only", "openspec", "", []string{"openspec/changes"}},
		{"both: spectre wins and the other's open change is reported", "spectre",
			"spec-root: DIR carries both spectre/changes/ and openspec/changes/ — using spectre/; every change under openspec/changes/ is invisible to this guard until that tree is moved\n",
			[]string{"spectre/changes", "openspec/changes/archive", "openspec/changes/open-change"}},
		{"both: an openspec tree of archive/ alone is not reported", "spectre", "",
			[]string{"spectre/changes", "openspec/changes/archive/old-change"}},
		{"both: an empty openspec/changes/ is not reported", "spectre", "",
			[]string{"spectre/changes", "openspec/changes"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, tree := range tc.trees {
				mkdir(t, dir+"/"+tree)
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
	smoke.Heavy(t)
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
	// The same record with PATH holding the stub alone: no jq on it.
	_, calls11n := store("11n", `{"source":"store","complete":true,"records":[{"projectKey":"proj-a","name":"x-repo-plan","worktrees":{"`+case11s+`":null,"`+canon11s+`":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}]}`)
	bin11n := filepath.Dir(calls11n)
	env11n := Env{Getenv: func(k string) string {
		if k == "PATH" {
			return bin11n
		}
		return os.Getenv(k)
	}}
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
	// Case 16: only an archived copy of the change, no live directory -- it
	// does not resolve: integrate drops a change's directory, and no reader
	// falls back to the archive; 16b: a live directory resolves beside an
	// archived one; 16c: an archive holding other changes only.
	case16 := work + "/case16-archived"
	plan(case16, "archive/arch-change")
	case16b := work + "/case16b-both"
	case16bDir := plan(case16b, "arch-change")
	plan(case16b, "archive/arch-change")
	case16c := work + "/case16c-neither"
	plan(case16c, "archive/other-change")
	// 16d: a satellite whose canonical worktree holds the change archived
	// only -- unresolved too.
	sat16d, canon16d := work+"/case16d-satellite", work+"/case16d-canonical"
	satellite(sat16d, "peerx:canon-change")
	plan(canon16d, "archive/canon-change")

	rows = append(rows,
		row{name: "store-resolves-absent-dir: the record's worktrees map resolves the plan's directory", fn: "dir", worktree: case11s, change: "x-repo-plan", want: canon11sDir, env: env11},
		row{name: "store-resolves-without-jq: the store resolves with no jq on PATH", fn: "dir", worktree: case11s, change: "x-repo-plan", want: canon11sDir, env: env11n},
		row{name: "store-resolves-satellite-without-peers: the record under the link's canonical id resolves", fn: "path", worktree: case12, change: "sat-change", want: canon12Dir + "/tasks.md", env: env12},
		row{name: "store-ambiguity-refuses: two resolving projects refuse with the ambiguity code 3", fn: "dir", worktree: case13, change: "ambig-plan", rc: 3, env: env13,
			wantErr: ptr("change-plan: ambiguous state-record resolution for 'ambig-plan': project proj-a resolves to " + tree13a + "/spectre/changes/ambig-plan\n" +
				"change-plan: ambiguous state-record resolution for 'ambig-plan': project proj-b resolves to " + tree13b + "/spectre/changes/ambig-plan\n")},
		row{name: "store-skipped-when-canonical-arg: a supplied canonical worktree never falls back to the store", fn: "path", worktree: case14, change: "sat-change", canonical: canon14, rc: 1, env: env14},
		row{name: "store-skipped-when-canonical-arg: the absent-dir site never falls back to the store behind a supplied canonical worktree", fn: "path", worktree: case14b, change: "x-repo-change", canonical: canon14b, rc: 1, env: env14b},
		row{name: "store-unavailable-stays-unresolvable: an unreachable store cannot resolve", fn: "path", worktree: case15, change: "lonely-plan", rc: 1, env: env15, wantErr: &none},
		row{name: "case 16: an archived change with no live directory does not resolve", fn: "path", worktree: case16, change: "arch-change", rc: 1, env: env15, wantErr: &none},
		row{name: "case 16 dir: an archived change's directory does not resolve", fn: "dir", worktree: case16, change: "arch-change", rc: 1, env: env15, wantErr: &none},
		row{name: "case 16b: a live change directory resolves beside an archived one", fn: "dir", worktree: case16b, change: "arch-change", want: case16bDir, env: env15},
		row{name: "case 16d: a satellite does not resolve its canonical worktree's archived plan", fn: "dir", worktree: sat16d, change: "sat-change", canonical: canon16d, rc: 1, env: env15, wantErr: &none},
		row{name: "case 16c: neither a live nor an archived directory cannot resolve", fn: "dir", worktree: case16c, change: "arch-change", rc: 1, env: env15, wantErr: &none},
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

// guardCache is the FLOW_GUARD_CACHE_DIR every test shares whose shim only
// has to reach the guard, the binary built into it once per run. A cache of
// its own costs each such test a `go build` of flow-guard and the first exec
// of a new binary, which macOS assesses serialised machine-wide
// (execFixtures). The key is the sources' hash, so a shim run from any
// checkout of these sources finds the one binary -- under this process's
// locale: flow_guard_key's glob sorts by collation, so a shim run with no
// LANG (check_plan_provenance's) keys and builds a second. A test about the
// build itself -- no go, a broken go -- keeps an empty cache of its own.
// Lives in execFixtures.dir, which TestMain removes after the run.
func guardCache(t *testing.T) string {
	t.Helper()
	return filepath.Dir(filepath.Dir(guardBinary(t)))
}

// guardBinary is the one flow-guard guardCache built. A test that execs the
// binary itself takes this path, never a glob of the cache: a no-LANG shim
// sharing the cache adds a second key's binary beside it at any time.
func guardBinary(t *testing.T) string {
	t.Helper()
	bin, err := guardCacheBuilt()
	if err != nil {
		t.Fatal(err)
	}
	return bin
}

var guardCacheBuilt = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(execFixtures.dir, "flow-guard-cache")
	boot, err := filepath.Abs("../../../scripts/flow-guard")
	if err != nil {
		return "", err
	}
	// The bootstrap builds under its key, then execs the binary;
	// spec-root's verdict is not read.
	cmd := exec.Command(boot, "spec-root")
	cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+dir)
	out, _ := cmd.CombinedOutput()
	bins, _ := filepath.Glob(dir + "/*/flow-guard")
	if len(bins) != 1 {
		return "", fmt.Errorf("building the shared flow-guard cache left %d binaries in %s:\n%s", len(bins), dir, out)
	}
	return bins[0], nil
})

// fixtureGit is the git binary every fixture runs, and every git stub a test
// writes execs as the real one: the binary `git` on PATH runs, resolved once
// through its own --exec-path. On macOS /usr/bin/git is an xcrun trampoline
// that spawns the real binary, which doubles each call's cost -- measured
// here at 14.5ms CPU and 26ms wall against 8ms and 12ms direct -- and the
// fixtures spawn thousands. The guards under test still resolve `git` from
// PATH themselves. Plain "git" when the lookup fails.
var fixtureGit = func() string {
	out, err := exec.Command("git", "--exec-path").Output()
	if err != nil {
		return "git"
	}
	p, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)) + "/git")
	if err != nil {
		return "git"
	}
	return p
}()

// fixtureGitPath is PATH with a directory holding only `git`, a link to
// fixtureGit, put first: a guard under test that resolves git over its PATH
// then skips the xcrun trampoline too, which halves each of its git calls.
func fixtureGitPath(t *testing.T) string {
	t.Helper()
	dir, err := fixtureGitDir()
	if err != nil {
		t.Fatal(err)
	}
	return dir + string(os.PathListSeparator) + os.Getenv("PATH")
}

var fixtureGitDir = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(execFixtures.dir, "git-bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, os.Symlink(fixtureGit, filepath.Join(dir, "git"))
})

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
	if err := fixtureGitExec(dir, args...); err != nil {
		t.Fatal(err)
	}
}

// fixtureGitExec is gitRun for code with no *testing.T: the shape templates below.
func fixtureGitExec(dir string, args ...string) error {
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command(fixtureGit, args...)
	cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v: %v\n%s", args, err, out)
	}
	return nil
}

// repoShape is one recurring fixture repository: git init [-b Branch] [--bare], plus one
// empty commit with message Commit when Commit is set.
type repoShape struct {
	Branch, Commit string
	Bare           bool
}

// shapeTemplates holds each shape built once per package run under execFixtures.dir, through
// fixtureGitExec's fixed identity, and copied per test: a `git init` plus a commit is a handful of
// execs, and the package's fixtures make one per test. A template is a directory; for the
// origin shapes it holds origin.git (bare) and clone/ (Commit pushed to it).
var shapeTemplates struct {
	sync.Mutex
	plain, origin map[repoShape]func() (string, error)
}

func shapeTemplate(s repoShape, withOrigin bool) (string, error) {
	shapeTemplates.Lock()
	m := &shapeTemplates.plain
	if withOrigin {
		m = &shapeTemplates.origin
	}
	if *m == nil {
		*m = map[repoShape]func() (string, error){}
	}
	build := (*m)[s]
	if build == nil {
		build = sync.OnceValues(func() (string, error) { return buildShape(s, withOrigin) })
		(*m)[s] = build
	}
	shapeTemplates.Unlock()
	return build()
}

func buildShape(s repoShape, withOrigin bool) (string, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "shape")
	if err != nil {
		return "", err
	}
	init := func(repo string, bare bool) error {
		args := []string{"init", "-q"}
		if s.Branch != "" {
			args = append(args, "-b", s.Branch)
		}
		if bare {
			args = append(args, "--bare")
		}
		return fixtureGitExec("", append(args, repo)...)
	}
	if !withOrigin {
		if err := init(dir+"/repo", s.Bare); err != nil {
			return "", err
		}
		if s.Commit != "" {
			// ponytail: no bare shape carries a commit yet; build it through a clone when one does.
			if s.Bare {
				return "", fmt.Errorf("repoShape %+v: a bare shape with a commit is not built", s)
			}
			if err := fixtureGitExec(dir+"/repo", "commit", "-q", "--allow-empty", "-m", s.Commit); err != nil {
				return "", err
			}
		}
		return dir + "/repo", nil
	}
	if s.Commit == "" || s.Branch == "" || s.Bare {
		return "", fmt.Errorf("repoShape %+v: an origin shape names a Branch and a Commit and is not bare", s)
	}
	origin, clone := dir+"/origin.git", dir+"/clone"
	for _, step := range []func() error{
		func() error { return init(origin, true) },
		func() error { return init(clone, false) },
		func() error { return fixtureGitExec(clone, "commit", "-q", "--allow-empty", "-m", s.Commit) },
		func() error { return fixtureGitExec(clone, "remote", "add", "origin", origin) },
		func() error { return fixtureGitExec(clone, "push", "-q", "-u", "origin", s.Branch) },
		// origin/HEAD, as `git clone` records it.
		func() error { return fixtureGitExec(clone, "remote", "set-head", "origin", s.Branch) },
	} {
		if err := step(); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// copyShape copies shape s into dir: a bare shape's repository is dir itself, any other's .git
// is dir/.git; dir is created if missing and files already in it are kept. It takes no
// *testing.T, so a sync.OnceValues master can build on it.
func copyShape(dir string, s repoShape) error {
	tpl, err := shapeTemplate(s, false)
	if err != nil {
		return err
	}
	if s.Bare {
		return copyTree(tpl, dir)
	}
	return copyTree(tpl+"/.git", dir+"/.git")
}

// fixtureRepo is copyShape, failing t on error.
func fixtureRepo(t *testing.T, dir string, s repoShape) {
	t.Helper()
	if err := copyShape(dir, s); err != nil {
		t.Fatal(err)
	}
}

// fixtureRepoWithOrigin copies the origin shape: a bare repository at origin and a clone at
// dir with s.Commit on s.Branch pushed to it and tracking it, origin/HEAD naming s.Branch as
// a clone's does, its origin URL rewritten to
// origin in Go. Worktrees and further commits are made after the copy: gitdir paths are
// absolute.
func fixtureRepoWithOrigin(t *testing.T, dir, origin string, s repoShape) {
	t.Helper()
	tpl, err := shapeTemplate(s, true)
	if err == nil {
		err = copyTree(tpl+"/origin.git", origin)
	}
	if err == nil {
		err = copyTree(tpl+"/clone/.git", dir+"/.git")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := dir + "/.git/config"
	b, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	from, to := "\turl = "+tpl+"/origin.git\n", "\turl = "+origin+"\n"
	if !bytes.Contains(b, []byte(from)) {
		t.Fatalf("%s carries no %q", cfg, from)
	}
	if err := os.WriteFile(cfg, bytes.Replace(b, []byte(from), []byte(to), 1), 0o644); err != nil {
		t.Fatal(err)
	}
}

// copyTree copies src to dst — regular files, directories and symlinks — keeping each one's
// mode and mtime; dst and its directories are created when missing, other files in them kept.
func copyTree(src, dst string) error {
	type dirTime struct {
		path string
		mt   time.Time
	}
	var dirs []dirTime
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		to := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			if err := os.MkdirAll(to, 0o755); err != nil {
				return err
			}
			dirs = append(dirs, dirTime{to, info.ModTime()})
			return os.Chmod(to, info.Mode().Perm())
		case d.Type()&fs.ModeSymlink != 0:
			l, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(l, to)
		case d.Type().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if err := os.WriteFile(to, b, info.Mode().Perm()); err != nil {
				return err
			}
			if err := os.Chmod(to, info.Mode().Perm()); err != nil {
				return err
			}
			return os.Chtimes(to, info.ModTime(), info.ModTime())
		}
		return fmt.Errorf("copyTree: %s is neither a file, a directory nor a symlink", p)
	})
	if err != nil {
		return err
	}
	// Children set their directory's mtime as they are written, so directories go last,
	// deepest first.
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := os.Chtimes(dirs[i].path, dirs[i].mt, dirs[i].mt); err != nil {
			return err
		}
	}
	return nil
}

// TestPlainNameRefusesEmpty pins the empty-name refusal plainChangeName took over
// from the bash change-plan library's name check when the package's two copies of
// the allowlist became one: a satellite ref such as `:change` hands it an
// empty peer name, which must be refused rather than indexed.
func TestPlainNameRefusesEmpty(t *testing.T) {
	t.Parallel()
	if plainChangeName("") {
		t.Fatal(`plainChangeName("") = true, want false`)
	}
}

// hermeticGitLookup is os.LookupEnv answering GIT_CONFIG_GLOBAL=/dev/null, so
// a guard that forwards it runs git without the operator's ~/.gitconfig.
func hermeticGitLookup(k string) (string, bool) {
	if k == "GIT_CONFIG_GLOBAL" {
		return "/dev/null", true
	}
	return os.LookupEnv(k)
}

// TestChildEnvTurnsGuardTelemetryOff pins TestMain's telemetry switch: no
// guard this package execs posts a guard_runs row to the dev store.
func TestChildEnvTurnsGuardTelemetryOff(t *testing.T) {
	if got := os.Getenv("FLOW_GUARD_TELEMETRY"); got != "off" {
		t.Fatalf("FLOW_GUARD_TELEMETRY = %q, want off", got)
	}
	cmd := exec.Command("sh", "-c", "printf %s \"$FLOW_GUARD_TELEMETRY\"")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "off" {
		t.Fatalf("child FLOW_GUARD_TELEMETRY = %q, want off", out)
	}
}

// historicalFile is <rev>:<path> read from the repository at root — the bash
// a parity test pins its port against, as it stood at <rev>. A clone that
// does not carry <rev>, a shallow one, skips the calling test naming the
// commit, rather than failing the whole package on history it never had
// (KAN-905).
func historicalFile(t *testing.T, root, rev, path string) []byte {
	t.Helper()
	if exec.Command(fixtureGit, "-C", root, "cat-file", "-e", rev+"^{commit}").Run() != nil {
		t.Skipf("%s is not in this clone's history (a shallow clone?) — the parity check against the bash at it cannot run", rev)
	}
	src, err := exec.Command(fixtureGit, "-C", root, "show", rev+":"+path).Output()
	if err != nil {
		t.Fatalf("git show %s:%s: %v", rev, path, err)
	}
	return src
}

// TestChangeDirIsLiveOnly pins that ChangeDir never resolves into
// changes/archive/: integrate drops the change directory instead of moving it.
func TestChangeDirIsLiveOnly(t *testing.T) {
	changes := filepath.Join(t.TempDir(), "changes")
	if err := os.MkdirAll(filepath.Join(changes, "archive", "kan-1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changes, "archive", "kan-1", "tasks.md"), []byte("- [ ] 1. One\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := ChangeDir(changes, "kan-1"), changes+"/kan-1"; got != want {
		t.Fatalf("ChangeDir = %s; want %s", got, want)
	}
}

// bashUsageRe matches a shim-era usage line's script name.
var bashUsageRe = regexp.MustCompile(`usage: ([a-z][a-z0-9-]*)\.sh\b`)

// bashUsage respells a bash-at-base usage line the way the guard now spells
// it, so a parity test compares everything else byte for byte.
func bashUsage(s string) string {
	return bashUsageRe.ReplaceAllString(s, "usage: flow-guard $1")
}

// guardShim writes an executable that runs `scripts/flow-guard <name>` of
// this checkout with its arguments, for a test that hands a guard to code
// expecting one executable path — a sibling override, a PATH entry.
func guardShim(t *testing.T, name string) string {
	t.Helper()
	boot, err := filepath.Abs("../../../scripts/flow-guard")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), name)
	writeExec(t, p, "#!/bin/sh\nexec '"+boot+"' "+name+" \"$@\"\n")
	return p
}

func TestTreeLeak(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	before, ok := treeStatus(repo)
	if !ok {
		t.Fatal("cannot read a fresh repository's status")
	}
	if leak := treeLeak(repo, before); leak != "" {
		t.Fatalf("an unchanged tree reported: %s", leak)
	}
	if err := os.WriteFile(filepath.Join(repo, "leaked_test.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if leak := treeLeak(repo, before); !strings.Contains(leak, "?? leaked_test.go") {
		t.Fatalf("a leaked file not reported: %q", leak)
	}
}

// TestSpecRootFrozenOpenspecIsSpectre pins kan-1086 self-review, F5: an
// openspec/changes/ holding archive/ alone is frozen history, so a spectre/
// tree beside it — even with no changes/ yet — wins; an open openspec change
// or no spectre/ keeps openspec.
func TestSpecRootFrozenOpenspecIsSpectre(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, want string
		trees      []string
	}{
		{"a frozen openspec tree beside spectre/ is spectre", "spectre", []string{"openspec/changes/archive/old", "spectre"}},
		{"a frozen openspec tree alone is openspec", "openspec", []string{"openspec/changes/archive/old"}},
		{"an open openspec change beside spectre/ is openspec", "openspec", []string{"openspec/changes/open-change", "spectre"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, tree := range tc.trees {
				mkdir(t, dir+"/"+tree)
			}
			var stderr bytes.Buffer
			if got := specRootLeaf(dir, &stderr); got != tc.want || stderr.Len() != 0 {
				t.Errorf("specRootLeaf = %q stderr %q, want %q and no stderr", got, stderr.String(), tc.want)
			}
		})
	}
}

// TestFixtureRepoCopiesTheShape pins the repository-shape templates: each copy
// is the shape it names, on its own branch with a clean tree, every copy of a
// shape is the one template, an origin copy pushes to its own origin, and
// copyTree keeps every copied file's mode and mtime.
func TestFixtureRepoCopiesTheShape(t *testing.T) {
	t.Parallel()
	g := &fxGit{}
	check := func(what, got, want string) {
		t.Helper()
		if g.err != nil {
			t.Fatalf("%s: %v", what, g.err)
		}
		if got != want {
			t.Errorf("%s = %q, want %q", what, got, want)
		}
	}
	root := t.TempDir()
	shape := repoShape{Branch: "main", Commit: "init"}
	a, b := root+"/a", root+"/b"
	fixtureRepo(t, a, shape)
	fixtureRepo(t, b, shape)
	for _, d := range []string{a, b} {
		check(d+" HEAD", g.git(d, "symbolic-ref", "HEAD"), "refs/heads/main")
		check(d+" log", g.git(d, "log", "--format=%s"), "init")
		check(d+" status", g.git(d, "status", "--porcelain"), "")
	}
	check("b's HEAD", g.git(b, "rev-parse", "HEAD"), g.git(a, "rev-parse", "HEAD"))

	bare := root + "/bare.git"
	fixtureRepo(t, bare, repoShape{Bare: true, Branch: "develop"})
	check("bare HEAD", g.git(bare, "symbolic-ref", "HEAD"), "refs/heads/develop")
	check("bare", g.git(bare, "rev-parse", "--is-bare-repository"), "true")

	unborn, fresh := root+"/unborn", root+"/fresh"
	fixtureRepo(t, unborn, repoShape{})
	g.git("", "init", "-q", fresh)
	check("unborn HEAD", g.git(unborn, "symbolic-ref", "HEAD"), g.git(fresh, "symbolic-ref", "HEAD"))
	if exec.Command(fixtureGit, "-C", unborn, "rev-parse", "-q", "--verify", "HEAD").Run() == nil {
		t.Error("repoShape{} copied a repository with a commit, want an unborn one")
	}

	clone, origin := root+"/clone", root+"/origin.git"
	fixtureRepoWithOrigin(t, clone, origin, shape)
	check("origin URL", g.git(clone, "remote", "get-url", "origin"), origin)
	check("origin/main", g.git(clone, "rev-parse", "origin/main"), g.git(clone, "rev-parse", "HEAD"))
	check("origin/HEAD", g.git(clone, "symbolic-ref", "refs/remotes/origin/HEAD"), "refs/remotes/origin/main")
	g.git(clone, "commit", "-q", "--allow-empty", "-m", "next")
	g.git(clone, "push", "-q", "origin", "main")
	check("origin's main after a push", g.git(origin, "rev-parse", "main"), g.git(clone, "rev-parse", "HEAD"))

	src, dst := root+"/src", root+"/dst"
	mkdir(t, src+"/sub")
	if err := os.WriteFile(src+"/sub/x", []byte("x"), 0o751); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(src+"/sub/x", 0o751); err != nil {
		t.Fatal(err)
	}
	mtime := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes(src+"/sub/x", mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("sub/x", src+"/link"); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(src, dst); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(dst + "/sub/x")
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o751 || !st.ModTime().Equal(mtime) {
		t.Errorf("copied file mode %v mtime %v, want -rwxr-x--x and %v", st.Mode().Perm(), st.ModTime(), mtime)
	}
	if l, err := os.Readlink(dst + "/link"); err != nil || l != "sub/x" {
		t.Errorf("copied symlink = %q, %v; want sub/x", l, err)
	}
}

// testPolicy is `flow policy get` as 0061_role_pairs.sql seeds it: the
// Allowed pairs table as it stood when the pairs moved into flowd
// (testdata/policy.json, which internal/repotest's hook tests read too).
// Embedded, not read off the working directory: a test binary self-exec'd
// as a guard (runSibling) runs from wherever its caller is, and -trimpath
// makes a runtime.Caller path module-relative, so neither locates the file
// (kan-1090 self-review planning).
//
//go:embed testdata/policy.json
var testPolicyJSON string

var testPolicy = strings.TrimSpace(testPolicyJSON)

// policyStub is the arm of a fake `flow` answering `policy get` with
// testPolicy; every fake flow of this package's policy readers opens with it.
var policyStub = "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '" + testPolicy + "'; exit 0; }\n"

// noFlowEnv is an env whose only `flow` answers `policy get` with
// testPolicy and fails everything else: the seeded pairs alone.
func noFlowEnv(t *testing.T) Env {
	t.Helper()
	return fakeFlowEnv(t, "exit 1\n")
}

// fakeFlowEnv is an env whose PATH holds only a `flow` answering `policy
// get` with testPolicy and running script for everything else.
func fakeFlowEnv(t *testing.T, script string, over ...string) Env {
	t.Helper()
	return fakeFlowEnvBare(t, policyStub+script, over...)
}

// policyFlowEnv is an env whose only `flow` answers `policy get` with
// policy and fails everything else.
func policyFlowEnv(t *testing.T, policy string, over ...string) Env {
	t.Helper()
	return fakeFlowEnvBare(t, "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '"+policy+"'; exit 0; }\nexit 1\n", over...)
}

// fakeFlowEnvBare is an env whose PATH holds only a `flow` running script.
func fakeFlowEnvBare(t *testing.T, script string, over ...string) Env {
	t.Helper()
	bin := t.TempDir()
	writeExec(t, bin+"/flow", "#!/bin/sh\n"+script)
	return envWith(t, append([]string{"PATH=" + bin}, over...))
}

var (
	psScripts, _ = filepath.Abs("../../../scripts")
	psBoot       = psScripts + "/flow-guard"
)

func psFixture(t *testing.T, body string) string {
	t.Helper()
	p := t.TempDir() + "/tasks.md"
	writeFile(t, p, body)
	return p
}

func psGetenv(vars map[string]string) func(string) string {
	return func(k string) string {
		if v, ok := vars[k]; ok {
			return v
		}
		if k == "PATH" {
			return os.Getenv(k)
		}
		return ""
	}
}

func psCopy(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, dst, string(b))
	if err := os.Chmod(dst, 0o755); err != nil {
		t.Fatal(err)
	}
}

func psLink(t *testing.T, target, link string) {
	t.Helper()
	mkdir(t, filepath.Dir(link))
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// psExec runs a script with bash from a directory outside any repository.
func psExec(t *testing.T, extra []string, args ...string) guardResult {
	t.Helper()
	cmd := exec.Command("bash", args...)
	cmd.Dir = t.TempDir()
	cmd.Env = append(append(os.Environ(), "CHECK_PLAN_SHAPE_ROOT="), extra...)
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	rc := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		rc = ee.ExitCode()
	}
	return guardResult{rc, strings.TrimRight(o.String()+e.String(), "\n"), o.String(), e.String()}
}

// bt turns the ¤ placeholder into a backtick: a Go raw string cannot hold one.
func bt(s string) string { return strings.ReplaceAll(s, "¤", "`") }

// tcfGit runs git -C dir under fixtureGitEnv, returning
// stdout without its trailing newline.
func tcfGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimRight(string(out), "\n")
}

// has is the harness's `case "$OUT" in *a*b*)`: every part, in order.
func has(s string, parts ...string) bool {
	for _, p := range parts {
		i := strings.Index(s, p)
		if i < 0 {
			return false
		}
		s = s[i+len(p):]
	}
	return true
}

func tcfScriptsDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../../scripts")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// pjJudgeRow is the pair-judge role's seeded row, as 0065 seeds it.
const pjJudgeRow = `"pair-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`

// pjPolicy is a `flow policy get` answer carrying the pair-judge row beside
// the given role rows.
func pjPolicy(rows ...string) string {
	return "{" + strings.Join(append(rows, pjJudgeRow), ",") + "}"
}

// pjStub is a judge command in a fresh directory: it saves its argv
// (NUL-separated) to argv and its stdin to stdin, then runs answer.
type pjStub struct{ dir string }

func newPJStub(t *testing.T, answer string) pjStub {
	t.Helper()
	s := pjStub{t.TempDir()}
	writeExec(t, s.cmd(), "#!/bin/sh\nprintf '%s\\0' \"$@\" > '"+s.dir+"/argv'\n/bin/cat > '"+s.dir+"/stdin'\n"+answer)
	return s
}

func (s pjStub) cmd() string { return s.dir + "/judge" }

// ran is whether the stub was invoked at all.
func (s pjStub) ran() bool {
	_, err := os.Stat(s.dir + "/argv")
	return err == nil
}

func (s pjStub) argv(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(s.dir + "/argv")
	if err != nil {
		t.Fatalf("the judge was never run: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(b), "\x00"), "\x00")
}

func (s pjStub) stdin(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(s.dir + "/stdin")
	if err != nil {
		t.Fatalf("the judge's stdin was never saved: %v", err)
	}
	return string(b)
}

// systemPrompt is the value the stub was given after --system-prompt.
func (s pjStub) systemPrompt(t *testing.T) string {
	t.Helper()
	a := s.argv(t)
	i := slices.Index(a, "--system-prompt")
	if i < 0 || i+1 >= len(a) {
		t.Fatalf("argv %q carries no --system-prompt value", a)
	}
	return a[i+1]
}

// pjRepoRoot is the agents checkout these tests run in, whose roles table
// the guard reads the Does line from.
func pjRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// pjEnv is an env whose `flow` answers policy with policy, whose judge
// command is stub, and whose repository root is this checkout.
func pjEnv(t *testing.T, policy string, stub pjStub) Env {
	t.Helper()
	return policyFlowEnv(t, policy, "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd())
}

// pjDigest writes body to a digest file and returns its path.
func pjDigest(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "digest.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// pjJSON is a stub answer printing the judge's text as `claude -p
// --output-format json` does, with extra JSON fields (a cost) appended.
func pjJSON(text, extra string) string {
	b, _ := json.Marshal(text)
	return "printf '%s\\n' '{\"type\":\"result\",\"result\":" + string(b) + extra + "}'\n"
}

func gdcCopyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := dst + strings.TrimPrefix(p, src)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, fi.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

// crEnv is an Env whose variables are vars alone, plus the process's PATH
// and locale for the `sort` the guard runs; a name absent from vars is unset.
func crEnv(dir string, vars map[string]string) Env {
	lookup := func(k string) (string, bool) {
		if v, ok := vars[k]; ok {
			return v, true
		}
		switch k {
		case "PATH", "LANG", "LC_ALL", "LC_COLLATE", "LC_CTYPE":
			return os.LookupEnv(k)
		}
		return "", false
	}
	return Env{Dir: dir, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
}
