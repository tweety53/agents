package guard

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Every case of the check-hand-notes-in-step contract, one subtest per
// verdict: the guard diffs the hand-maintained content — everything outside
// the `<!-- flow:begin -->` … `<!-- flow:end -->` managed block — of every
// harness file setup.sh's `local managed_files=(...)` declaration names, and
// fails when one file's hand section carries an edit the other lacks. Like
// TestCheckInstalledRules, every case runs against this checkout's own
// setup.sh unless a case overrides it, and no case reads or writes the
// operator's real $HOME: CHECK_HAND_NOTES_HOME names a t.TempDir().
//
// The fixtures deliberately give the two files DIFFERENT managed blocks — the
// installer rewrites `~/.claude/` pointers to `~/.zcode/` ones — pinning that
// block content is never read as drift.

// hnhRoot is this checkout, the root the shim exports.
var hnhRoot = cirRoot

// hnhFixtureHome is a known-good pair: identical hand sections around
// managed blocks that differ the way the installer's pointer rewrites differ.
func hnhFixtureHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	hnhWritePair(t, home,
		"<!-- flow:begin -->\n<!-- rule: be-brief.mdc -->\n\nFull rule: `~/.claude/rules/be-brief.md`.\n\n<!-- flow:end -->\n",
		"<!-- flow:begin -->\n<!-- rule: be-brief.mdc -->\n\nFull rule: `~/.zcode/rules/be-brief.md`.\n\n<!-- flow:end -->\n")
	return home
}

// hnhWritePair writes <home>/.claude/CLAUDE.md and <home>/.zcode/AGENTS.md
// with the given managed blocks around one shared hand section.
func hnhWritePair(t *testing.T, home, claudeBlock, agentsBlock string) {
	t.Helper()
	hand := "# Personal notes\n\nOnly the operator writes here.\n"
	writeFile(t, home+"/.claude/CLAUDE.md", hand+claudeBlock+hand)
	writeFile(t, home+"/.zcode/AGENTS.md", hand+agentsBlock+hand)
}

func TestCheckHandNotesInStep(t *testing.T) {
	t.Parallel()
	root, err := hnhRoot()
	if err != nil {
		t.Fatal(err)
	}

	const p = "check-hand-notes-in-step: "
	run := func(t *testing.T, home, setupSh, repoRoot string) (int, string, string) {
		t.Helper()
		getenv := func(k string) string {
			switch k {
			case "CHECK_HAND_NOTES_HOME":
				return home
			case "CHECK_HAND_NOTES_SETUP_SH":
				return setupSh
			case "FLOW_GUARD_REPO_ROOT":
				return repoRoot
			}
			return ""
		}
		var stdout, stderr bytes.Buffer
		code := checkHandNotesInStep(nil, Env{Getenv: getenv, LookupEnv: os.LookupEnv, Dir: t.TempDir()}, &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}

	t.Run("in_step: identical hand sections, differing managed blocks", func(t *testing.T) {
		t.Parallel()
		home := hnhFixtureHome(t)
		code, stdout, stderr := run(t, home, "", root)
		if code != 0 || stderr != "" {
			t.Fatalf("code = %d, stderr = %q; want 0, \"\"", code, stderr)
		}
		want := "HAND-NOTES-OK: " + home + " — 2 harness file(s), hand-maintained sections in step\n"
		if stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
	})

	t.Run("one_sided_append: the drift is named at its line", func(t *testing.T) {
		t.Parallel()
		home := hnhFixtureHome(t)
		agents := home + "/.zcode/AGENTS.md"
		b, err := os.ReadFile(agents)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, agents, string(b)+"# Flake bisection\n\nBisect on the smallest unit first.\n")
		code, stdout, stderr := run(t, home, "", root)
		if code != 1 {
			t.Fatalf("code = %d, want 1", code)
		}
		wantVerdict := "HAND-NOTES-DRIFT: " + home + " — 1 violation(s); make every harness file's hand-maintained sections identical\n"
		if stdout != wantVerdict {
			t.Fatalf("stdout = %q, want %q", stdout, wantVerdict)
		}
		wantViolation := p + agents + " differs from " + home + "/.claude/CLAUDE.md" +
			" at hand-section line 7: " + strconv.Quote("# Flake bisection\n") + " vs " + strconv.Quote("<EOF>") + "\n"
		if stderr != wantViolation {
			t.Fatalf("stderr = %q, want %q", stderr, wantViolation)
		}
	})

	t.Run("missing_final_newline: the comparison is byte-exact", func(t *testing.T) {
		t.Parallel()
		home := hnhFixtureHome(t)
		claude := home + "/.claude/CLAUDE.md"
		b, err := os.ReadFile(claude)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, claude, strings.TrimSuffix(string(b), "\n"))
		code, _, stderr := run(t, home, "", root)
		if code != 1 {
			t.Fatalf("code = %d, want 1", code)
		}
		agents := home + "/.zcode/AGENTS.md"
		if !strings.Contains(stderr, agents+" differs from "+claude+" at hand-section line 6: ") ||
			!strings.Contains(stderr, strconv.Quote("Only the operator writes here.\n")+" vs "+strconv.Quote("Only the operator writes here.")) {
			t.Fatalf("stderr = %q, want a line-numbered byte difference naming both sides", stderr)
		}
	})

	t.Run("none: no global install is not stale", func(t *testing.T) {
		t.Parallel()
		home := t.TempDir()
		code, stdout, stderr := run(t, home, "", root)
		if code != 0 || stderr != "" {
			t.Fatalf("code = %d, stderr = %q; want 0, \"\"", code, stderr)
		}
		want := "HAND-NOTES-NONE: " + home + " — no global install found, nothing to check\n"
		if stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
	})

	t.Run("single: one harness file has no pair to compare", func(t *testing.T) {
		t.Parallel()
		home := t.TempDir()
		writeFile(t, home+"/.zcode/AGENTS.md", "# notes\n")
		code, stdout, _ := run(t, home, "", root)
		if code != 0 {
			t.Fatalf("code = %d, want 0", code)
		}
		want := "HAND-NOTES-SINGLE: " + home + " — one harness file present, no pair to compare\n"
		if stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
	})

	t.Run("unterminated: a file ending inside its managed block", func(t *testing.T) {
		t.Parallel()
		home := t.TempDir()
		hand := "# notes\n"
		writeFile(t, home+"/.claude/CLAUDE.md", hand+"<!-- flow:begin -->\nnever closed\n")
		writeFile(t, home+"/.zcode/AGENTS.md", hand)
		code, stdout, stderr := run(t, home, "", root)
		if code != 1 {
			t.Fatalf("code = %d, want 1", code)
		}
		if !strings.Contains(stderr, p+home+"/.claude/CLAUDE.md ends inside an unterminated managed block — repair the file\n") {
			t.Fatalf("stderr = %q, want the unterminated-block violation", stderr)
		}
		if !strings.Contains(stdout, "HAND-NOTES-DRIFT: "+home) {
			t.Fatalf("stdout = %q, want the DRIFT verdict", stdout)
		}
	})

	t.Run("no_declaration: a setup.sh without the declaration refuses", func(t *testing.T) {
		t.Parallel()
		home := hnhFixtureHome(t)
		setupSh := home + "/setup.sh"
		writeFile(t, setupSh, "#!/usr/bin/env bash\n# no managed_files here\n")
		code, stdout, stderr := run(t, home, setupSh, root)
		if code != 1 || stdout != "" {
			t.Fatalf("code = %d, stdout = %q; want 1, \"\"", code, stdout)
		}
		want := p + "no 'local managed_files=(' declaration in " + setupSh + " — cannot resolve the managed-block targets\n"
		if stderr != want {
			t.Fatalf("stderr = %q, want %q", stderr, want)
		}
	})

	t.Run("no_repo_root: the default installer path cannot resolve", func(t *testing.T) {
		t.Parallel()
		home := hnhFixtureHome(t)
		code, stdout, stderr := run(t, home, "", "")
		if code != 1 || stdout != "" {
			t.Fatalf("code = %d, stdout = %q; want 1, \"\"", code, stdout)
		}
		want := p + "FLOW_GUARD_REPO_ROOT is unset and CHECK_HAND_NOTES_SETUP_SH is not set — run scripts/check-hand-notes-in-step.sh, which sets it\n"
		if stderr != want {
			t.Fatalf("stderr = %q, want %q", stderr, want)
		}
	})
}
