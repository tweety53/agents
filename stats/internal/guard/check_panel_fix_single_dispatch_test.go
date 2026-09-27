package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Every case of scripts/test-check-panel-fix-single-dispatch.sh at d71a2327,
// one subtest per ok: label. The harness's stub `flow` answered
// `record dispatches` and `record findings` from canned JSON; here those are
// Env.Dispatches and Env.Findings, and the "real flow" subtests exec a stub
// `flow` on Env's PATH so the production read runs too.

// pfdDispatches is the harness's dispatch_json: key, role, token triples.
func pfdDispatches(vals ...string) []byte {
	out := []map[string]string{}
	for i := 0; i+2 < len(vals); i += 3 {
		out = append(out, map[string]string{"key": vals[i], "role": vals[i+1], "sessionToken": vals[i+2]})
	}
	b, _ := json.Marshal(out)
	return b
}

// pfdFindings is the harness's findings_json: n fixed rows raised in round.
func pfdFindings(n, round int) []byte {
	out := []map[string]any{}
	for i := 1; i <= n; i++ {
		out = append(out, map[string]any{"ref": "F" + strconv.Itoa(i), "round": round, "status": "fixed"})
	}
	b, _ := json.Marshal(out)
	return b
}

type pfdCase struct {
	dispatches, findings []byte
	dErr, fErr           error
	args                 []string // nil means the harness's run_guard defaults
}

func pfdRun(t *testing.T, c pfdCase) (int, string, string) {
	t.Helper()
	fn := Registry["check-panel-fix-single-dispatch"]
	if fn == nil {
		t.Fatal("check-panel-fix-single-dispatch is not registered")
	}
	if c.findings == nil {
		c.findings = []byte("[]")
	}
	args := c.args
	if args == nil {
		args = []string{t.TempDir(), "demo", "mf-tok"}
	}
	env := Env{Getenv: pathEnv(t.TempDir()), Dir: t.TempDir(),
		Dispatches: func(string) ([]byte, error) { return c.dispatches, c.dErr },
		Findings:   func(string) ([]byte, error) { return c.findings, c.fErr }}
	var out, errb bytes.Buffer
	code := fn(args, env, &out, &errb)
	return code, out.String(), errb.String()
}

func pfdExpect(t *testing.T, c pfdCase, want int, needle string) {
	t.Helper()
	got, out, errs := pfdRun(t, c)
	if got != want {
		t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", want, got, out, errs)
	}
	if !strings.Contains(out+errs, needle) {
		t.Fatalf("expected output to name %q, got:\n%s%s", needle, out, errs)
	}
}

func pfdExpectNoStderr(t *testing.T, c pfdCase) {
	t.Helper()
	got, out, errs := pfdRun(t, c)
	if got != 0 || errs != "" {
		t.Fatalf("expected exit 0 and no stderr, got %d\nstdout: %s\nstderr: %s", got, out, errs)
	}
	if out != "PANEL-FIX-SINGLE-DISPATCH-OK: every panel-fix dispatch of this run is one per chunked round (bounds and retries included)\n" {
		t.Fatalf("stdout: %q", out)
	}
}

func TestCheckPanelFixSingleDispatch(t *testing.T) {
	t.Parallel()
	const pf = "panel-fix"
	d := pfdDispatches
	cases := []struct {
		label  string
		c      pfdCase
		want   int
		needle string // "" with want 0 means: no stderr at all
	}{
		{"case 1: one canonical dispatch per round exits 0 with no stderr",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-2", pf, "mf-tok")}, 0, ""},
		{"case 2: original plus -retry exits 0",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-retry", pf, "mf-tok")}, 0, "OK"},
		{"case 3: two originals in one round exit 1, naming the round",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1", pf, "mf-tok")}, 1, "panel-fix-1"},
		{"case 4: invented key exits 1, naming it verbatim",
			pfdCase{dispatches: d("panel-fix-f1", pf, "mf-tok")}, 1, "panel-fix-f1"},
		{"case 5: retry-only round exits 1, naming the round",
			pfdCase{dispatches: d("panel-fix-1-retry", pf, "mf-tok")}, 1, "panel-fix-1"},
		{"case 6: other roles ignored, foreign token rows ignored",
			pfdCase{dispatches: d("task-1-implementer", "implementer", "mf-tok", "slot-primary", "reviewer", "mf-tok",
				"panel-fix-1", pf, "mf-tok", "panel-fix-1", pf, "mf-other", "panel-fix-1", pf, "mf-other")}, 0, "OK"},
		{"case 7a: missing session-token argument exits 2",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok"), args: []string{"/", "demo", ""}}, 2, "session token is required"},
		{"case 7b: non-directory worktree exits 2",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok"), args: []string{"/no-such-dir", "demo", "mf-tok"}}, 2, "not a directory: /no-such-dir"},
		{"case 7c: store unreachable exits 2",
			pfdCase{dErr: errors.New("exit status 1")}, 2, "flow record dispatches failed for 'demo' -- cannot answer"},
		{"case 7d: change name outside the allowlist exits 2",
			pfdCase{dispatches: []byte("[]"), args: []string{"/", "../evil", "mf-tok"}}, 2, "change name '../evil' is not a plain change name"},
		{"case 8: chunked-legal-round-passes",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-2", pf, "mf-tok", "panel-fix-1-3", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 0, ""},
		{"case 9: chunk-retry-pair-passes",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-2", pf, "mf-tok", "panel-fix-1-2-retry", pf, "mf-tok"), findings: pfdFindings(20, 0)}, 0, "OK"},
		{"case 10: legacy-round-with-findings-verb-passes",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-retry", pf, "mf-tok")}, 0, "OK"},
		{"case 11: chunk-without-bare-key-fails, naming the round",
			pfdCase{dispatches: d("panel-fix-1-2", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 1, "panel-fix-1"},
		{"case 12: chunk-gap-fails, naming the gapped chunk",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-2", pf, "mf-tok", "panel-fix-1-4", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 1, "panel-fix-1-4"},
		{"case 13: chunk-count-over-bound-fails, naming the round",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-2", pf, "mf-tok", "panel-fix-1-3", pf, "mf-tok"), findings: pfdFindings(15, 0)}, 1, "panel-fix-1"},
		{"case 14: chunk-retry-alone-fails, naming the chunk",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-2-retry", pf, "mf-tok"), findings: pfdFindings(20, 0)}, 1, "panel-fix-1-2"},
		{"case 15: two-originals-on-chunk-fails, naming the chunk",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-2", pf, "mf-tok", "panel-fix-1-2", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 1, "panel-fix-1-2"},
		{"case 16: invented-chunk-key-shape-fails, naming it verbatim",
			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-f2", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 1, "panel-fix-1-f2"},
		{"case 17: findings-store-unreachable-exits-2",
			pfdCase{dispatches: []byte("[]"), fErr: errors.New("exit status 1")}, 2, "flow record findings failed for 'demo' -- cannot answer"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			if tc.want == 0 && tc.needle == "" {
				pfdExpectNoStderr(t, tc.c)
				return
			}
			pfdExpect(t, tc.c, tc.want, tc.needle)
		})
	}

	// The exact stderr lines the bash printed at d71a2327 on this input, in its order:
	// out-of-shape keys as rows arrive, then per base, then per round.
	t.Run("violation lines byte for byte", func(t *testing.T) {
		t.Parallel()
		got, out, errs := pfdRun(t, pfdCase{
			dispatches: d("panel-fix-2-2", pf, "mf-tok", "panel-fix-x", pf, "mf-tok", "panel-fix-1", pf, "mf-tok",
				"panel-fix-1", pf, "mf-tok", "panel-fix-1-retry", pf, "mf-tok", "panel-fix-3-retry", pf, "mf-tok",
				"panel-fix-1-2", pf, "mf-tok", "panel-fix-1-4", pf, "mf-tok", "panel-fix-1-3", pf, "mf-tok",
				"panel-fix-1-3-retry", pf, "mf-tok", "panel-fix-1-3-retry", pf, "mf-tok"),
			findings: pfdFindings(11, 0)})
		want := "check-panel-fix-single-dispatch: out-of-shape panel-fix key 'panel-fix-x' -- the canonical shape is panel-fix-<round>[-<chunk>][-retry]\n" +
			"check-panel-fix-single-dispatch: round panel-fix-1 carries 2 panel-fix dispatches -- the fix goes to ONE subagent as the combined list, never one per reviewer, slot or finding\n" +
			"check-panel-fix-single-dispatch: round panel-fix-3 carries 1 -retry dispatch(es) and no original -- a retry is never a round's only dispatch\n" +
			"check-panel-fix-single-dispatch: round panel-fix-1-3 carries 3 panel-fix dispatches -- at most the original plus the handshake's one -retry\n" +
			"check-panel-fix-single-dispatch: round panel-fix-2 carries chunked dispatches but no panel-fix-2 original -- chunk numbering starts at the round's own key\n" +
			"check-panel-fix-single-dispatch: round panel-fix-1 carries 4 chunk dispatch(es) for 11 finding(s) raised in earlier rounds -- at most ceil(11/10)=2, never one per finding\n"
		if got != 1 || out != "" || errs != want {
			t.Fatalf("exit %d\nstdout: %q\nstderr:\n%s\nwant:\n%s", got, out, errs, want)
		}
	})

	// CANONICAL_KEY_RE as RE2: each key's verdict is what bash's
	// [[ =~ ]] gave under LC_ALL=C, measured at d71a2327.
	t.Run("canonical key regexp matches bash", func(t *testing.T) {
		t.Parallel()
		for key, ok := range map[string]bool{
			"panel-fix-1": true, "panel-fix-01": true, "panel-fix-1-2": true, "panel-fix-1-2-retry": true,
			"panel-fix-1-retry-retry": false, "panel-fix-f1": false, "panel-fix-": false, "panel-fix-1\nx": false,
			"panel-fix-1-2-3": false, "PANEL-FIX-1": false, "xpanel-fix-1": false, "panel-fix-١": false,
		} {
			if pfdKeyRE.MatchString(key) != ok {
				t.Errorf("%q: want %v", key, ok)
			}
		}
	})

	// The production reads: a stub `flow` on PATH, hooks nil.
	t.Run("real flow reads on PATH", func(t *testing.T) {
		t.Parallel()
		fn := Registry["check-panel-fix-single-dispatch"]
		if fn == nil {
			t.Fatal("check-panel-fix-single-dispatch is not registered")
		}
		for _, sc := range []struct {
			stub string
			want int
			line string
		}{
			{"case \"${2:-}\" in dispatches) echo '[{\"key\":\"panel-fix-1\",\"role\":\"panel-fix\",\"sessionToken\":\"mf-tok\"},{\"key\":\"panel-fix-1\",\"role\":\"panel-fix\",\"sessionToken\":\"mf-tok\"}]' ;; findings) echo '[]' ;; *) exit 2 ;; esac",
				1, "round panel-fix-1 carries 2 panel-fix dispatches"},
			{"echo 'flow: connect: connection refused' >&2; exit 1", 2, "flow record dispatches failed for 'demo' -- cannot answer"},
			{"case \"${2:-}\" in dispatches) echo '[]' ;; *) echo refused >&2; exit 1 ;; esac", 2, "flow record findings failed for 'demo' -- cannot answer"},
		} {
			bin := t.TempDir()
			writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\n"+sc.stub+"\n")
			var out, errb bytes.Buffer
			got := fn([]string{t.TempDir(), "demo", "mf-tok"}, Env{Getenv: pathEnv(bin), Dir: t.TempDir()}, &out, &errb)
			if got != sc.want || !strings.Contains(errb.String(), sc.line) || strings.Contains(errb.String(), "refused") {
				t.Errorf("%s: exit %d stderr %q", sc.stub, got, errb.String())
			}
		}
	})

	// bash's `cd -wt` read the worktree as an option and failed; its refusal
	// printed $WORKTREE after the failed substitution had emptied it.
	t.Run("dash-led worktree refused as vanished", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mkdir(t, filepath.Join(dir, "-wt"))
		var out, errb bytes.Buffer
		got := Registry["check-panel-fix-single-dispatch"]([]string{"-wt", "demo", "mf-tok"},
			Env{Getenv: pathEnv(t.TempDir()), Dir: dir}, &out, &errb)
		if got != 2 || errb.String() != "check-panel-fix-single-dispatch: worktree vanished: \n" {
			t.Fatalf("exit %d stderr %q", got, errb.String())
		}
	})

	t.Run("unreadable dispatch rows exit 2", func(t *testing.T) {
		t.Parallel()
		for _, rows := range []string{"not json", "null", "[1]", `[{"key":"panel-fix-1"}] 5`} {
			pfdExpect(t, pfdCase{dispatches: []byte(rows)}, 2, "dispatch rows were not readable JSON -- cannot answer")
		}
		pfdExpect(t, pfdCase{dispatches: []byte("[]"), findings: []byte("{")}, 2, "findings rows were not readable JSON -- cannot answer")
	})

	// Beyond the harness (panel review of task 9): mechanisms the bash got
	// from jq and $(...) are pinned against the bash at d71a2327 run live,
	// with a stub `flow` on PATH serving the same payloads to both sides.
	t.Run("port: implicit jq and command-substitution mechanisms match the bash at d71a2327", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("jq"); err != nil {
			t.Skip("jq is not on PATH; the bash at d71a2327 requires it")
		}
		script := filepath.Join(t.TempDir(), "check-panel-fix-single-dispatch.sh")
		src, err := exec.Command("git", "-C", "../../..", "show", "d71a2327:scripts/check-panel-fix-single-dispatch.sh").Output()
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, script, string(src))
		row := func(key string) string {
			return `{"key":` + key + `,"role":"panel-fix","sessionToken":"mf-tok"}`
		}
		chunked := "[" + row(`"panel-fix-1"`) + "," + row(`"panel-fix-1-2"`) + "," + row(`"panel-fix-1-3"`) + "]"
		for _, fx := range []struct{ label, dispatches, findings string }{
			{"a NUL in a key is dropped", "[" + row(`"panel-fix-1\u0000"`) + "]", "[]"},
			{"a key's trailing newlines are stripped", "[" + row(`"panel-fix-1\n\n"`) + "]", "[]"},
			{"a true round sorts below every number", "[" + row(`"panel-fix-0"`) + "," + row(`"panel-fix-0-2"`) + "," + row(`"panel-fix-0-3"`) + "]", `[{"round":true},{"round":false}]`},
			{"an empty findings read is an empty count", chunked, ""},
			{"a round number wraps at 64 bits", "[" + row(`"panel-fix-99999999999999999999-2"`) + "]", "[]"},
			{"a null dispatch row selects nothing", "[null," + row(`"panel-fix-1"`) + "]", "[]"},
		} {
			bin := t.TempDir()
			writeFile(t, filepath.Join(bin, "d.json"), fx.dispatches)
			writeFile(t, filepath.Join(bin, "f.json"), fx.findings)
			writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\ncase \"${2:-}\" in dispatches) cat \""+bin+"/d.json\" ;; findings) cat \""+bin+"/f.json\" ;; *) exit 2 ;; esac\n")
			wt := t.TempDir()
			got, gout, gerr := pfdRun(t, pfdCase{dispatches: []byte(fx.dispatches), findings: []byte(fx.findings), args: []string{wt, "demo", "mf-tok"}})
			cmd := exec.Command("bash", script, wt, "demo", "mf-tok")
			cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"))
			var bout, berr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &bout, &berr
			want := 0
			if err := cmd.Run(); err != nil {
				var ee *exec.ExitError
				if !errors.As(err, &ee) {
					t.Fatal(err)
				}
				want = ee.ExitCode()
			}
			// bash 5's own "ignored null byte" warning (3.2 prints none)
			// names its script path; a Go binary cannot emit it (recorded in
			// task 9's Correction).
			bashErr := regexp.MustCompile(`(?m)^.*: warning: command substitution: ignored null byte in input\n`).ReplaceAllString(berr.String(), "")
			if got != want || gout != bout.String() || gerr != bashErr {
				t.Errorf("%s:\nport rc=%d stdout=%q stderr=%q\nbash rc=%d stdout=%q stderr=%q", fx.label, got, gout, gerr, want, bout.String(), berr.String())
			}
		}
	})

	// Declared divergences, pinned on the port's side: a findings stream of
	// several values on a chunked round is the header's cannot-answer 2
	// (the bash passed the round after an arithmetic error), and a worktree
	// the caller cannot search is refused as the bash's `cd` refused it.
	t.Run("port: a multi-value findings stream cannot answer", func(t *testing.T) {
		t.Parallel()
		d := pfdDispatches("panel-fix-1", "panel-fix", "mf-tok", "panel-fix-1-2", "panel-fix", "mf-tok")
		pfdExpect(t, pfdCase{dispatches: d, findings: []byte("[] []")}, 2, "findings rows were not readable JSON -- cannot answer")
	})
	t.Run("port: a worktree without search permission is refused as vanished", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("running as root, permission bits are not enforced")
		}
		wt := t.TempDir()
		if err := os.Chmod(wt, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(wt, 0o755) })
		code, _, errb := pfdRun(t, pfdCase{dispatches: []byte("[]"), args: []string{wt, "demo", "mf-tok"}})
		if code != 2 || errb != "check-panel-fix-single-dispatch: worktree vanished: \n" {
			t.Fatalf("exit %d stderr %q", code, errb)
		}
	})
}
