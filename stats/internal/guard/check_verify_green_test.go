package guard

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// vgRepo makes a git worktree with one commit and returns it with that
// commit's sha.
func vgRepo(t *testing.T) (string, string) {
	t.Helper()
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, "src/a.go"), "package a\n")
	gitRun(t, wt, "init", "-q")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-qm", "a")
	return wt, vgHead(t, wt)
}

func vgHead(t *testing.T, wt string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", wt, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func vgCommit(t *testing.T, wt, path string) string {
	t.Helper()
	writeFile(t, filepath.Join(wt, path), "changed\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-qm", "edit "+path)
	return vgHead(t, wt)
}

func vgRow(key, sha string) string {
	return `{"key":"` + key + `","role":"verifier","sessionToken":"mf-tok","outcome":"completed","diffBase":"` + sha + `"}`
}

func vgRun(t *testing.T, rows string, args ...string) (int, string) {
	t.Helper()
	env := Env{Getenv: os.Getenv, Dir: args[0], Dispatches: func(string) ([]byte, error) { return []byte(rows), nil }}
	var out, errb bytes.Buffer
	rc := Registry["check-verify-green"](args, env, &out, &errb)
	return rc, out.String() + errb.String()
}

func TestCheckVerifyGreenNeedsAHead(t *testing.T) {
	t.Parallel()
	wt, _ := vgRepo(t)
	rc, out := vgRun(t, "["+vgRow("verify", "")+"]", wt, "demo", "mf-tok")
	if rc != 1 || !strings.Contains(out, "VERIFY-NOT-GREEN:") || !strings.Contains(out, "verify (no HEAD recorded)") {
		t.Fatalf("exit %d\n%s", rc, out)
	}
}

func TestCheckVerifyGreenSpansPlanningCommits(t *testing.T) {
	t.Parallel()
	wt, a := vgRepo(t)
	vgCommit(t, wt, "spectre/changes/demo/tasks.md")
	rc, out := vgRun(t, "["+vgRow("verify", a)+"]", wt, "demo", "mf-tok")
	if rc != 0 || !strings.Contains(out, "VERIFY-GREEN:") || !strings.Contains(out, "verify@"+a[:7]) {
		t.Fatalf("exit %d\n%s", rc, out)
	}
}

func TestCheckVerifyGreenRefusesACodeCommitSinceTheRow(t *testing.T) {
	t.Parallel()
	wt, a := vgRepo(t)
	c := vgCommit(t, wt, "src/a.go")
	rc, out := vgRun(t, "["+vgRow("verify", a)+"]", wt, "demo", "mf-tok")
	if rc != 1 || !strings.Contains(out, "VERIFY-NOT-GREEN:") || !strings.Contains(out, c[:7]+" since "+a[:7]) {
		t.Fatalf("exit %d\n%s", rc, out)
	}
}

func TestCheckVerifyGreenJudgesEachWorktree(t *testing.T) {
	t.Parallel()
	wa, a := vgRepo(t)
	wb, _ := vgRepo(t)
	b := vgCommit(t, wb, "src/b.go")
	vgCommit(t, wb, "spectre/changes/demo/design.md")
	rows := "[" + vgRow("verify-a", a) + "," + vgRow("verify-b", b) + "]"
	if rc, out := vgRun(t, rows, wa, "demo", "mf-tok", wb); rc != 0 || !strings.Contains(out, "verify-a@"+a[:7]) || !strings.Contains(out, "verify-b@"+b[:7]) {
		t.Fatalf("both resolve: exit %d\n%s", rc, out)
	}
	wx, _ := vgRepo(t)
	x := vgCommit(t, wx, "src/x.go")
	if rc, out := vgRun(t, "["+vgRow("verify-a", a)+","+vgRow("verify-b", x)+"]", wa, "demo", "mf-tok", wb); rc != 1 || !strings.Contains(out, "verify-b (HEAD resolves in no worktree)") {
		t.Fatalf("a sha in neither: exit %d\n%s", rc, out)
	}
}

func TestCheckVerifyGreen(t *testing.T) {
	t.Parallel()
	wt, head := vgRepo(t)
	row := func(key, token, outcome string) string {
		return `{"key":"` + key + `","role":"verifier","sessionToken":"` + token + `","outcome":"` + outcome + `","diffBase":"` + head + `"}`
	}
	for _, c := range []struct {
		name, rows string
		err        error
		want       int
		needle     string
	}{
		{"every inline verify row completed", "[" + row("verify-fe", "mf-tok", "completed") + "," + row("verify-be", "mf-tok", "completed") + "]", nil, 0, "VERIFY-GREEN"},
		{"a cache-skipped verify is green", "[" + row("verify-2-gymie", "mf-tok", "skipped") + "," + row("verify-2-gymie-frontend", "mf-tok", "skipped") + "]", nil, 0, "VERIFY-GREEN"},
		{"a verify still running", "[" + row("verify", "mf-tok", "") + "]", nil, 1, "verify (running)"},
		{"a blocked verify", "[" + row("verify", "mf-tok", "blocked") + "]", nil, 1, "verify (blocked)"},
		{"a blocked verify re-run green by a fix round", "[" + row("verify", "mf-tok", "blocked") + "," + row("verify-fix-1", "mf-tok", "completed") + "]", nil, 0, "1 inline verify row(s) green"},
		{"a fix round's verify judged per worktree", "[" + row("verify-fe", "mf-tok", "blocked") + "," + row("verify-be", "mf-tok", "completed") + "," + row("verify-fix-1-fe", "mf-tok", "completed") + "]", nil, 0, "2 inline verify row(s) green"},
		{"a blocked verify re-run green after a new commit", "[" + row("verify", "mf-tok", "blocked") + "," + row("verify-2", "mf-tok", "completed") + "]", nil, 0, "1 inline verify row(s) green"},
		{"a re-run of a digit-led repository's blocked verify is green", "[" + row("verify-3d-viewer", "mf-tok", "blocked") + "," + row("verify-2-3d-viewer", "mf-tok", "completed") + "]", nil, 0, "1 inline verify row(s) green"},
		{"a re-run judged per worktree", "[" + row("verify-fe", "mf-tok", "blocked") + "," + row("verify-2-fe", "mf-tok", "completed") + "]", nil, 0, "1 inline verify row(s) green"},
		{"a fix round's blocked verify after a green one", "[" + row("verify", "mf-tok", "completed") + "," + row("verify-fix-1", "mf-tok", "blocked") + "]", nil, 1, "verify-fix-1 (blocked)"},
		{"another run's green and a visual verifier do not count", "[" + row("verify", "mf-old", "completed") + "," + row("visual-verify", "mf-tok", "completed") + "]", nil, 1, "no inline verify"},
		{"an unreadable store cannot answer", "", errors.New("down"), 2, "cannot answer"},
		{"no row for the change names the canonical worktree", "[]", nil, 2, "pass the change's canonical worktree"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			env := Env{Getenv: os.Getenv, Dir: wt, Dispatches: func(string) ([]byte, error) { return []byte(c.rows), c.err }}
			var out, errb bytes.Buffer
			got := Registry["check-verify-green"]([]string{env.Dir, "demo", "mf-tok"}, env, &out, &errb)
			if got != c.want || !strings.Contains(out.String()+errb.String(), c.needle) {
				t.Fatalf("exit %d, want %d naming %q\n%s%s", got, c.want, c.needle, out.String(), errb.String())
			}
		})
	}
}
