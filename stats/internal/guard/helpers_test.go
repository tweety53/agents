package guard

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

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
		if got := core.Sha256Hex(tc.in); got != tc.want {
			t.Errorf("sha256Hex(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}

	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	nul := filepath.Join(dir, "nul")
	guardtest.WriteFile(t, empty, "")
	guardtest.WriteFile(t, nul, "a\x00b")
	for _, tc := range []struct{ path, want string }{
		{empty, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		// A NUL byte survives: the file's bytes are hashed, never a string copy.
		{nul, guardtest.Shasum(t, nul)},
	} {
		got, err := core.Sha256HexFile(tc.path)
		if err != nil {
			t.Fatalf("sha256HexFile(%s): %v", tc.path, err)
		}
		if got != tc.want {
			t.Errorf("sha256HexFile(%s) = %s, want %s", tc.path, got, tc.want)
		}
	}
	if _, err := core.Sha256HexFile(filepath.Join(dir, "absent")); err == nil {
		t.Error("sha256HexFile on an absent file: want an error")
	}
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
				guardtest.Mkdir(t, dir+"/"+tree)
			}
			var stderr bytes.Buffer
			if got := core.SpecRootLeaf(dir, &stderr); got != tc.want {
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
	guardtest.WriteExec(t, defaultBin+"/flow", "#!/usr/bin/env bash\nexit 1\n")
	env := core.Env{Getenv: guardtest.PathEnv(defaultBin)}

	satellite := func(dir, ref string) {
		guardtest.Mkdir(t, dir+"/spectre/changes/sat-change")
		guardtest.WriteFile(t, dir+"/spectre/changes/sat-change/link.md", "## Part of\n\n`"+ref+"`\n")
	}
	plan := func(tree, id string) string {
		d := tree + "/spectre/changes/" + id
		guardtest.WriteFile(t, d+"/tasks.md", "# plan\n\n- [ ] 1. do a thing\n")
		return d
	}
	peers := func(tree, line string) { guardtest.WriteFile(t, tree+"/spectre/peers", line+"\n") }

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
	guardtest.Mkdir(t, canon2b+"/spectre/changes")
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
	guardtest.WriteFile(t, canon5+"/spectre/changes/canon-only-satellite-dir/link.md",
		"## Parts\n\n`peerz:some-part`\n\n## Merge order\n\n1. `.`\n2. `peerz`\n")
	// Case 5b: an empty `## Part of` must not read past the next heading.
	sat5b, peer5b := work+"/case5b-worktree", work+"/case5b-peer"
	guardtest.WriteFile(t, sat5b+"/spectre/changes/sat-change/link.md", "## Part of\n\n## Parts\n\n`peerz:some-part`\n")
	peers(sat5b, "peerz ../case5b-peer")
	plan(peer5b, "some-part")
	// Case 6a-6c: containment — change name, peer name, canonical change id.
	bad6 := work + "/case6-worktree"
	guardtest.Mkdir(t, bad6+"/spectre/changes")
	guardtest.WriteFile(t, bad6+"/secret/tasks.md", "# secret\n")
	sat6b := work + "/case6b-worktree"
	satellite(sat6b, "../escape:canon-change")
	peers(sat6b, "../escape ../elsewhere")
	plan(work+"/elsewhere", "canon-change")
	sat6c, canon6c := work+"/case6c-worktree", work+"/case6c-canonical"
	satellite(sat6c, "peerx:../escape")
	guardtest.WriteFile(t, canon6c+"/spectre/escape/tasks.md", "# escaped\n")
	// Case 10-10f: the absent-local-dir branch.
	sat10, canon10 := work+"/case10-no-dir", work+"/case10-canonical"
	guardtest.Mkdir(t, sat10+"/spectre/changes")
	canon10Dir := plan(canon10, "x-repo-change")
	sat10b := work + "/case10b-no-dir"
	guardtest.Mkdir(t, sat10b+"/spectre/changes")
	sat10c, canon10c := work+"/case10c-satellite", work+"/case10c-canonical"
	satellite(sat10c, "peerx:canon-change")
	plan(canon10c, "sat-change")
	sat10d, canon10d := work+"/case10d-empty-scaffold", work+"/case10d-canonical"
	guardtest.Mkdir(t, sat10d+"/spectre/changes/x-repo-change")
	canon10dDir := plan(canon10d, "x-repo-change")
	sat10e := work + "/case10e-empty-scaffold"
	guardtest.Mkdir(t, sat10e+"/spectre/changes/x-repo-change")
	sat10f, canon10f := work+"/case10f-parts-only-link", work+"/case10f-canonical"
	guardtest.WriteFile(t, sat10f+"/spectre/changes/sat-change/link.md", "## Parts\n\n`peerq:some-part`\n")
	plan(canon10f, "sat-change")
	// Case 11-11b: peers resolve beside the MAIN checkout, from a linked worktree.
	main11, peer11 := work+"/case11-main", work+"/case11-peer"
	guardtest.GitRun(t, "", "init", "-q", main11)
	satellite(main11, "peerv:canon-change")
	peers(main11, "peerv ../case11-peer")
	plan(peer11, "canon-change")
	guardtest.GitRun(t, main11, "add", "spectre")
	guardtest.GitRun(t, main11, "commit", "-qm", "peers")
	wt11 := work + "/case11-wtroot/kan-11-worktree"
	guardtest.Mkdir(t, work+"/case11-wtroot")
	guardtest.GitRun(t, main11, "worktree", "add", "-q", wt11, "-b", "case11-wtb")
	peer11Resolved, err := filepath.EvalSymlinks(peer11)
	if err != nil {
		t.Fatal(err)
	}
	// Case 11d: a separate-git-dir primary, and a linked worktree of it.
	sgd11, peer11d := work+"/case11d-sgd", work+"/case11d-peer"
	guardtest.Mkdir(t, work+"/case11d-elsewhere/gitdirs")
	guardtest.GitRun(t, "", "init", "-q", "--separate-git-dir="+work+"/case11d-elsewhere/gitdirs/repo.git", sgd11)
	satellite(sgd11, "peeru:canon-change")
	peers(sgd11, "peeru ../case11d-peer")
	peer11dDir := plan(peer11d, "canon-change")
	guardtest.GitRun(t, sgd11, "add", "spectre")
	guardtest.GitRun(t, sgd11, "commit", "-qm", "x")
	guardtest.Mkdir(t, work+"/case11d-wtroot")
	guardtest.GitRun(t, sgd11, "worktree", "add", "-q", work+"/case11d-wtroot/wt", "-b", "case11d-wtb")
	plan(work+"/case11d-elsewhere/case11d-peer", "canon-change") // the decoy

	// A nil wantErr means stderr is not asserted (the bash case discarded it).
	type row struct {
		name, fn                    string // fn: path, dir or ref
		worktree, change, canonical string
		want                        string
		rc                          int
		env                         core.Env
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
			wantErr: guardtest.Ptr("change-plan: change name '../../secret' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n")},
		{name: "case 6b stderr: the rejected peer name is explained", fn: "path", worktree: sat6b, change: "sat-change", rc: 1,
			wantErr: guardtest.Ptr("change-plan: peer name '../escape' in " + sat6b + "/spectre/changes/sat-change/link.md is not a plain peer name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n")},
		{name: "case 6c stderr: the rejected canonical change id is explained", fn: "path", worktree: sat6c, change: "sat-change", canonical: canon6c, rc: 1,
			wantErr: guardtest.Ptr("change-plan: canonical change id '../escape' in " + sat6c + "/spectre/changes/sat-change/link.md is not a plain change id — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n")},
	}

	// The store cases: a stub `flow` answering `state find` from a fixture,
	// logging every call. Each case gets its own stub directory.
	store := func(name, fixture string) (core.Env, string) {
		bin := work + "/store-" + name
		calls := bin + "/calls"
		// Paths named from the stub's own, so every store stub shares one of
		// two bodies (WriteExec).
		body := "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"${0%/*}/calls\"\n"
		if fixture == "" {
			body += "echo 'flow: store unreachable' >&2\nexit 1\n"
		} else {
			guardtest.WriteFile(t, bin+"/fixture.json", fixture)
			body += "cat \"${0%/*}/fixture.json\"\n"
		}
		guardtest.WriteExec(t, bin+"/flow", body)
		return core.Env{Getenv: guardtest.PathEnv(bin)}, calls
	}
	record := func(project, tree string) string {
		return `{"projectKey":"` + project + `","name":"x","worktrees":{"` + tree + `":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`
	}
	case11s, canon11s := work+"/case11s-no-dir", work+"/case11s-canonical"
	guardtest.Mkdir(t, case11s+"/spectre/changes")
	canon11sDir := plan(canon11s, "x-repo-plan")
	env11, calls11 := store("11", `{"source":"store","complete":true,"records":[{"projectKey":"proj-a","name":"x-repo-plan","worktrees":{"`+case11s+`":null,"`+canon11s+`":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}]}`)
	// The same record with PATH holding the stub alone: no jq on it.
	_, calls11n := store("11n", `{"source":"store","complete":true,"records":[{"projectKey":"proj-a","name":"x-repo-plan","worktrees":{"`+case11s+`":null,"`+canon11s+`":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}]}`)
	bin11n := filepath.Dir(calls11n)
	env11n := core.Env{Getenv: func(k string) string {
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
	guardtest.Mkdir(t, case13+"/spectre/changes")
	plan(tree13a, "ambig-plan")
	plan(tree13b, "ambig-plan")
	env13, _ := store("13", `{"records":[`+record("proj-a", tree13a)+","+record("proj-b", tree13b)+`]}`)
	case14, canon14, peer14 := work+"/case14-satellite", work+"/case14-canonical-empty", work+"/case14-peer"
	satellite(case14, "skippeer:canon-change")
	peers(case14, "skippeer ../case14-peer")
	guardtest.Mkdir(t, canon14+"/spectre/changes")
	plan(peer14, "canon-change")
	env14, calls14 := store("14", `{"records":[`+record("proj-a", peer14)+`]}`)
	case14b, canon14b := work+"/case14b-no-dir", work+"/case14b-canonical-empty"
	guardtest.Mkdir(t, case14b+"/spectre/changes")
	guardtest.Mkdir(t, canon14b+"/spectre/changes")
	env14b, calls14b := store("14b", `{"records":[`+record("proj-a", work+"/case14b-elsewhere")+`]}`)
	case15 := work + "/case15-no-dir"
	guardtest.Mkdir(t, case15+"/spectre/changes")
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
			wantErr: guardtest.Ptr("change-plan: ambiguous state-record resolution for 'ambig-plan': project proj-a resolves to " + tree13a + "/spectre/changes/ambig-plan\n" +
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

// TestPlainNameRefusesEmpty pins the empty-name refusal PlainChangeName took over
// from the bash change-plan library's name check when the package's two copies of
// the allowlist became one: a satellite ref such as `:change` hands it an
// empty peer name, which must be refused rather than indexed.
func TestPlainNameRefusesEmpty(t *testing.T) {
	t.Parallel()
	if core.PlainChangeName("") {
		t.Fatal(`plainChangeName("") = true, want false`)
	}
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

func TestTreeLeak(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	before, ok := guardtest.TreeStatus(repo)
	if !ok {
		t.Fatal("cannot read a fresh repository's status")
	}
	if leak := guardtest.TreeLeak(repo, before); leak != "" {
		t.Fatalf("an unchanged tree reported: %s", leak)
	}
	if err := os.WriteFile(filepath.Join(repo, "leaked_test.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if leak := guardtest.TreeLeak(repo, before); !strings.Contains(leak, "?? leaked_test.go") {
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
				guardtest.Mkdir(t, dir+"/"+tree)
			}
			var stderr bytes.Buffer
			if got := core.SpecRootLeaf(dir, &stderr); got != tc.want || stderr.Len() != 0 {
				t.Errorf("specRootLeaf = %q stderr %q, want %q and no stderr", got, stderr.String(), tc.want)
			}
		})
	}
}

// TestFixtureRepoCopiesTheShape pins the repository-shape templates: each copy
// is the shape it names, on its own branch with a clean tree, every copy of a
// shape is the one template, an origin copy pushes to its own origin, and
// CopyTree keeps every copied file's mode and mtime.
func TestFixtureRepoCopiesTheShape(t *testing.T) {
	t.Parallel()
	g := &guardtest.FxGit{}
	check := func(what, got, want string) {
		t.Helper()
		if g.Err != nil {
			t.Fatalf("%s: %v", what, g.Err)
		}
		if got != want {
			t.Errorf("%s = %q, want %q", what, got, want)
		}
	}
	root := t.TempDir()
	shape := guardtest.RepoShape{Branch: "main", Commit: "init"}
	a, b := root+"/a", root+"/b"
	guardtest.FixtureRepo(t, a, shape)
	guardtest.FixtureRepo(t, b, shape)
	for _, d := range []string{a, b} {
		check(d+" HEAD", g.Git(d, "symbolic-ref", "HEAD"), "refs/heads/main")
		check(d+" log", g.Git(d, "log", "--format=%s"), "init")
		check(d+" status", g.Git(d, "status", "--porcelain"), "")
	}
	check("b's HEAD", g.Git(b, "rev-parse", "HEAD"), g.Git(a, "rev-parse", "HEAD"))

	bare := root + "/bare.git"
	guardtest.FixtureRepo(t, bare, guardtest.RepoShape{Bare: true, Branch: "develop"})
	check("bare HEAD", g.Git(bare, "symbolic-ref", "HEAD"), "refs/heads/develop")
	check("bare", g.Git(bare, "rev-parse", "--is-bare-repository"), "true")

	unborn, fresh := root+"/unborn", root+"/fresh"
	guardtest.FixtureRepo(t, unborn, guardtest.RepoShape{})
	g.Git("", "init", "-q", fresh)
	check("unborn HEAD", g.Git(unborn, "symbolic-ref", "HEAD"), g.Git(fresh, "symbolic-ref", "HEAD"))
	if exec.Command(guardtest.FixtureGit, "-C", unborn, "rev-parse", "-q", "--verify", "HEAD").Run() == nil {
		t.Error("repoShape{} copied a repository with a commit, want an unborn one")
	}

	clone, origin := root+"/clone", root+"/origin.git"
	guardtest.FixtureRepoWithOrigin(t, clone, origin, shape)
	check("origin URL", g.Git(clone, "remote", "get-url", "origin"), origin)
	check("origin/main", g.Git(clone, "rev-parse", "origin/main"), g.Git(clone, "rev-parse", "HEAD"))
	check("origin/HEAD", g.Git(clone, "symbolic-ref", "refs/remotes/origin/HEAD"), "refs/remotes/origin/main")
	g.Git(clone, "commit", "-q", "--allow-empty", "-m", "next")
	g.Git(clone, "push", "-q", "origin", "main")
	check("origin's main after a push", g.Git(origin, "rev-parse", "main"), g.Git(clone, "rev-parse", "HEAD"))

	src, dst := root+"/src", root+"/dst"
	guardtest.Mkdir(t, src+"/sub")
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
	if err := guardtest.CopyTree(src, dst); err != nil {
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

var (
	psScripts, _ = filepath.Abs("../../../scripts")
	psBoot       = psScripts + "/flow-guard"
)

func psFixture(t *testing.T, body string) string {
	t.Helper()
	p := t.TempDir() + "/tasks.md"
	guardtest.WriteFile(t, p, body)
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
	guardtest.WriteFile(t, dst, string(b))
	if err := os.Chmod(dst, 0o755); err != nil {
		t.Fatal(err)
	}
}

func psLink(t *testing.T, target, link string) {
	t.Helper()
	guardtest.Mkdir(t, filepath.Dir(link))
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// psExec runs a script with bash from a directory outside any repository.
func psExec(t *testing.T, extra []string, args ...string) guardtest.GuardResult {
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
	return guardtest.GuardResult{Rc: rc, Out: strings.TrimRight(o.String()+e.String(), "\n"), Stdout: o.String(), Err: e.String()}
}

// tcfGit runs git -C dir under FixtureGitEnv, returning
// stdout without its trailing newline.
func tcfGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(guardtest.FixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), guardtest.FixtureGitEnv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimRight(string(out), "\n")
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
	guardtest.WriteExec(t, s.cmd(), "#!/bin/sh\nprintf '%s\\0' \"$@\" > '"+s.dir+"/argv'\n/bin/cat > '"+s.dir+"/stdin'\n"+answer)
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
func pjEnv(t *testing.T, policy string, stub pjStub) core.Env {
	t.Helper()
	return guardtest.PolicyFlowEnv(t, policy, "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd())
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

// crEnv is an Env whose variables are vars alone, plus the process's PATH
// and locale for the `sort` the guard runs; a name absent from vars is unset.
func crEnv(dir string, vars map[string]string) core.Env {
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
	return core.Env{Dir: dir, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
}
