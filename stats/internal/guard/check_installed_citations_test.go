package guard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Every case of scripts/test-check-installed-citations.sh at 3e48ecac, one
// subtest per ok: label. Each fixture is a miniature agents repository with
// its own runnable setup.sh, so a case changes what "installed" means
// without touching this repository; the guard runs in-process with
// CHECK_INSTALLED_CITATIONS_ROOT on Env and really runs that setup.sh.
//
// Three of the harness's labels asserted on the harness itself, not the
// guard: "CHECK PHASE replays every case in ascending build order" and the
// two EXIT-trap-chaining cases, which pinned how the harness combined its
// own cleanup trap with scripts/lib/parallel.sh's. Neither mechanism exists
// once the harness is gone; the three rows standing in for them pin the
// guard-side facts they protected — the guard's own sandbox is gone after a
// clean run and after a refused one, and setup.sh ran inside that sandbox.

// cicSetup is write_fixture_setup_sh's minimal installer: `global` symlinks
// every skills/<name>/ and every always-on rules/*.mdc into
// $HOME/.claude/{skills,rules} (and docs/ when present); `claude-code <proj>`
// and `zcode <proj>` copy CLAUDE.md into <proj>. Every fixture links one inode (writeExec).
const cicSetup = `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="${1:-}"
PROJ="${2:-$SCRIPT_DIR}"
: "${HOME:?HOME must be set}"
case "$MODE" in
  global)
    mkdir -p "$HOME/.claude/skills" "$HOME/.claude/rules"
    for d in "$SCRIPT_DIR"/skills/*/; do
      [ -d "$d" ] || continue
      ln -sfn "${d%/}" "$HOME/.claude/skills/$(basename "$d")"
    done
    for f in "$SCRIPT_DIR"/rules/*.mdc; do
      [ -f "$f" ] || continue
      if grep -q 'alwaysApply: true' "$f"; then
        ln -sfn "$f" "$HOME/.claude/rules/$(basename "$f")"
      fi
    done
    if [ -d "$SCRIPT_DIR/docs" ]; then
      ln -sfn "$SCRIPT_DIR/docs" "$HOME/.claude/docs"
    fi
    ;;
  claude-code|zcode)
    mkdir -p "$PROJ"
    if [ -f "$SCRIPT_DIR/CLAUDE.md" ] && [ ! -f "$PROJ/CLAUDE.md" ]; then
      cp "$SCRIPT_DIR/CLAUDE.md" "$PROJ/CLAUDE.md"
    fi
    ;;
  *)
    echo "fixture setup.sh: unknown mode '$MODE'" >&2
    exit 1
    ;;
esac
`

// cicSafe always classifies and always judges clean, so a fixture file a
// case is not testing never trips an undeclared-zero finding.
const cicSafe = "`skills/other/SKILL.md`"

// cicFixture is new_fixture_repo: one always-installed skill pair, one
// always-on rule, one opt-in rule (never installed, never scanned), a
// CLAUDE.md (a copy target) and a README.md (a real repository-root file).
func cicFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeExec(t, root+"/setup.sh", cicSetup)
	writeFile(t, root+"/skills/flow/SKILL.md", "# flow fixture\n"+cicSafe+"\n")
	writeFile(t, root+"/skills/other/SKILL.md", "# other fixture\n`skills/flow/SKILL.md`\n")
	writeFile(t, root+"/rules/always-on-rule.mdc", "---\nalwaysApply: true\n---\n# always-on rule fixture\n\n"+cicSafe+"\n")
	writeFile(t, root+"/rules/opt-in-rule.mdc", "---\nalwaysApply: false\n---\n# opt-in rule fixture\n\nplaceholder\n")
	writeFile(t, root+"/CLAUDE.md", "# CLAUDE.md fixture\n"+cicSafe+"\n")
	writeFile(t, root+"/README.md", "# README fixture\n")
	return root
}

type cicResult struct {
	rc          int
	out, errOut string
}

func cicRun(t *testing.T, root string, vars map[string]string) cicResult {
	t.Helper()
	all := map[string]string{"CHECK_INSTALLED_CITATIONS_ROOT": root}
	for k, v := range vars {
		all[k] = v
	}
	var out, errOut bytes.Buffer
	rc := checkInstalledCitations(nil, crEnv(t.TempDir(), all), &out, &errOut)
	return cicResult{rc, out.String(), errOut.String()}
}

// hasLine is `grep -aq -- "^<prefix>"` over out's lines.
func hasLine(out, prefix string) bool {
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, prefix) {
			return true
		}
	}
	return false
}

func TestCheckInstalledCitations(t *testing.T) {
	t.Parallel()
	const flow = "skills/flow/SKILL.md"
	// register_case: content, then SAFE_CITATION, written with the trailing
	// newline the bash `$(printf …)` stripped.
	type classifier struct {
		label, rel, content string
		reported            bool
	}
	cases := []classifier{
		{"bare-root-file", flow, "# flow fixture\n`README.md`\n", true},
		{"bare-generic-filename", flow, "# flow fixture\n`tasks.md`\n", false},
		{"installed-root", flow, "# flow fixture\n`skills/other/SKILL.md`\n", false},
		{"agents-repo-prefix", flow, "# flow fixture\n`<agents repo>/README.md`\n", false},
		{"project-prefix", flow, "# flow fixture\n`<project>/.flow/project.md`\n", false},
		{"unrooted-directory", flow, "# flow fixture\n`.flow/project.md`\n", true},
		{"unrooted-script", flow, "# flow fixture\n`scripts/check-references.sh`\n", true},
		{"fenced-command", flow, "# flow fixture\n\n```bash\ngit -C <abs-worktree> reset -- spectre/\n```\n", false},
		{"fenced-comment", flow, "# flow fixture\n\n```bash\n# see spectre/\n```\n", true},
		{"fenced-comment-agents-repo-prefix", flow, "# flow fixture\n\n```bash\n# see <agents repo>/README.md for details\n```\n", false},
		{"absolute-path", flow, "# flow fixture\n`/etc/hosts`\n", false},
		{"home-path", flow, "# flow fixture\n`~/.claude/skills/`\n", false},
		{"url", flow, "# flow fixture\n`https://example.test/a/b`\n", false},
		{"shell-variable", flow, "# flow fixture\n`$SCRIPT_DIR/lib/x.sh`\n", false},
		{"git-ref", flow, "# flow fixture\n`origin/main`\n", false},
		{"regex-fragment", flow, "# flow fixture\n`[A-Za-z0-9._-]+/x`\n", false},
		{"opt-in-rule-out-of-scope", "rules/opt-in-rule.mdc", "---\nalwaysApply: false\n---\n# opt-in rule fixture\n\n`README.md`\n", false},
		{"parent-relative-path", flow, "# flow fixture\n`../<other-app>`\n", false},
		{"placeholder-rooted", flow, "# flow fixture\n`<state-dir>/<name>-proposal-artifact.html`\n", false},
		{"quoted-program-output", flow, "# flow fixture\n`'spectre/<name>':`\n", false},
		{"unrecognised-placeholder-is-reported", flow, "# flow fixture\n`<foo>/spectre/specs/x.md`\n", true},
		{"git-branch-name-is-not-a-citation", flow, "# flow fixture\nBranch `spectre/<name>`.\n", false},
		{"file-line-reference-is-not-a-citation", flow, "# flow fixture\n| F1 | Bugbot | Minor | `src/Foo.kt:42` | replaced the silent catch |\n", false},
		{"origin-with-extension-is-a-citation", flow, "# flow fixture\n`origin/README.md`\n", true},
		{"origin-ref-with-nested-path-is-not-a-citation", flow, "# flow fixture\n`origin/spectre/<name>`\n", false},
		{"origin-directory-is-a-citation", flow, "# flow fixture\n`origin/rules/`\n", true},
		{"second-word-in-backtick-span-is-seen", flow, "# flow fixture\n`see .flow/project.md`\n", true},
		{"shell-example-second-word-not-path-shaped", flow, "# flow fixture\n`skills/other/SKILL.md verbose`\n", false},
		{"skill-dir-rooted", flow, "# flow fixture\n`<skill-dir>/scripts/check-references.sh`\n", false},
		{"sibling-worktrees-rooted", flow, "# flow fixture\n`<project>-worktrees/<name>/spectre/changes/<name>/`\n", false},
		{"agents-repo-sibling-worktrees-rooted", flow, "# flow fixture\n`<agents repo>-worktrees/<slug>`\n", false},
		{"sibling-worktrees-lookalike-is-reported", flow, "# flow fixture\n`<project>-worktrees-old/<name>/x.md`\n", true},
		{"spectre-branch-with-change-name-is-not-a-citation", flow, "# flow fixture\n`spectre/<change-name>`\n", false},
		{"spectre-shape-with-real-path-is-still-reported", flow, "# flow fixture\n`spectre/specs/x.md`\n", true},
		{"spectre-shape-with-trailing-path-is-still-reported", flow, "# flow fixture\n`spectre/changes/<name>/`\n", true},
		{"chore-archive-branch-name-is-not-a-citation", flow, "# flow fixture\nBranch `chore/archive-<name>`.\n", false},
		{"chore-archive-shape-with-trailing-path-is-still-reported", flow, "# flow fixture\n`chore/archive-<name>/spec.md`\n", true},
		{"fraction-is-not-a-citation", flow, "# flow fixture\nStatus `Task 22/22` and `Task <x>/<n>`.\n", false},
		{"fraction-shape-with-trailing-path-is-still-reported", flow, "# flow fixture\n`22/22/spec.md`\n", true},
		{"multi-letter-placeholder-pair-is-still-reported", flow, "# flow fixture\n`<name>/<file>`\n", true},
		{"chore-archive-bracket-cannot-smuggle-a-slash", flow, "# flow fixture\n`chore/archive-<a/b>`\n", true},
		{"html-comment-is-not-a-citation", flow, "# flow fixture\n`<!-- measured: ./gradlew test @ c515c42 -->`\n", false},
		{"html-comment-with-leading-whitespace-is-not-a-citation", flow, "# flow fixture\n` <!-- measured: ./gradlew test @ c515c42 -->`\n", false},
		{"colon-segment-is-not-a-citation", flow, "# flow fixture\n`measured:/predicted:`\n", false},
		{"final-colon-segment-is-a-citation", flow, "# flow fixture\n`.flow/project.md:`\n", true},
		{"redirection-is-not-merged-into-a-citation", flow, "# flow fixture\nRun `some-cmd < skills/other/SKILL.md > output.log` to reproduce.\n", false},
		{"na-marker-line-without-marker-is-reported", flow, "# flow fixture\nThe sweep reads `n/a — no frame` beside the other frame-bound readings.\n", true},
		{"na-marker-line-is-exempt", flow, "# flow fixture\nThe sweep reads `n/a — no frame` beside the other frame-bound readings. <!-- citations-guard:allow -->\n", false},
		{"allow-marker-exempts-a-command-shape", flow, "# flow fixture\n`git add -- <the spec, its PNGs, <changeRoot>/visual-verification/>` first <!-- citations-guard:allow -->\n", false},
		{"allow-marker-is-line-scoped", flow, "# flow fixture\n`somewhere/else.md` is unrooted and reported\nThe sweep reads `n/a — no frame` <!-- citations-guard:allow -->\n", true},
		{"allow-marker-in-shell-comment-is-exempt", flow, "# flow fixture\n\n```bash\n# the sweep reads n/a — no frame <!-- citations-guard:allow -->\n```\n", false},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			root := cicFixture(t)
			writeFile(t, filepath.Join(root, c.rel), c.content+"\n"+cicSafe)
			r := cicRun(t, root, nil)
			wantRC := 0
			if c.reported {
				wantRC = 1
			}
			if r.rc != wantRC || hasLine(r.out, c.rel+":") != c.reported {
				t.Fatalf("want reported=%v: rc=%d out=%q err=%q", c.reported, r.rc, r.out, r.errOut)
			}
		})
	}

	// recognised: the citation was seen — its member's coverage count is 1
	// — not merely unreported.
	for _, c := range []struct{ label, content string }{
		{"agents-repo-prefix-bogus-path: recognised (coverage count 1), not silently dropped", "# flow fixture\n`<agents repo>/does-not-exist/nested/path.md`\n"},
		{"placeholder-rooted-is-recognised: recognised (coverage count 1), not silently dropped", "# flow fixture\n`<state-dir>/does-not-exist.html`\n"},
		{"abs-worktree-rooted-is-recognised: recognised (coverage count 1), not silently dropped", "# flow fixture\n`<abs-worktree>/.superpowers/sdd/does-not-exist.diff`\n"},
		{"skill-dir-rooted-is-recognised: recognised (coverage count 1), not silently dropped", "# flow fixture\n`<skill-dir>/scripts/does-not-exist.sh`\n"},
	} {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			root := cicFixture(t)
			writeFile(t, root+"/"+flow, c.content)
			r := cicRun(t, root, nil)
			if r.rc != 0 || hasLine(r.out, flow+":") || !strings.Contains(r.out, flow+" 1") {
				t.Fatalf("rc=%d out=%q err=%q", r.rc, r.out, r.errOut)
			}
		})
	}

	t.Run("newly-installed-directory", func(t *testing.T) {
		t.Parallel()
		root := cicFixture(t)
		writeFile(t, root+"/docs/somefile.md", "# docs fixture\n\n`README.md`\n")
		r := cicRun(t, root, nil)
		if r.rc != 1 || !hasLine(r.out, "docs/somefile.md:") {
			t.Fatalf("rc=%d out=%q err=%q", r.rc, r.out, r.errOut)
		}
	})

	// Refusals: exit 2, a reason on stderr, nothing on stdout.
	refused := func(t *testing.T, r cicResult, needle string) {
		t.Helper()
		if r.rc != 2 || r.out != "" || !strings.Contains(r.errOut, needle) {
			t.Fatalf("rc=%d out=%q err=%q, want 2, no stdout, stderr naming %q", r.rc, r.out, r.errOut, needle)
		}
	}
	t.Run("CHECK_INSTALLED_CITATIONS_ROOT set but empty: exit 2, stderr names it, stdout empty", func(t *testing.T) {
		t.Parallel()
		refused(t, cicRun(t, "", nil), "CHECK_INSTALLED_CITATIONS_ROOT")
	})
	t.Run("root is not a directory: exit 2, stdout empty", func(t *testing.T) {
		t.Parallel()
		refused(t, cicRun(t, filepath.Join(t.TempDir(), "does-not-exist"), nil), "")
	})
	failingSetup := func(t *testing.T) string {
		root := t.TempDir()
		writeExec(t, root+"/setup.sh", "#!/usr/bin/env bash\necho \"fixture setup.sh: deliberately failing\" >&2\nexit 1\n")
		mkdir(t, root+"/skills")
		mkdir(t, root+"/rules")
		return root
	}
	t.Run("the fixture's setup.sh exits non-zero: exit 2, stderr names the failed install, stdout empty", func(t *testing.T) {
		t.Parallel()
		refused(t, cicRun(t, failingSetup(t), nil), "setup.sh")
	})
	t.Run("the guard is asked to install with HOME outside its own sandbox: refused, no setup.sh invocation", func(t *testing.T) {
		t.Parallel()
		root, outside := t.TempDir(), t.TempDir()
		marker := root + "/ran"
		writeExec(t, root+"/setup.sh", "#!/usr/bin/env bash\ntouch \"${0%/*}/ran\"\n")
		_, err := cicRunSetup(context.Background(), root, root, outside, "", "global")
		var refusal *cicSandboxRefusal
		if !errors.As(err, &refusal) {
			t.Fatalf("err = %v, want a sandbox refusal", err)
		}
		if _, err := os.Stat(marker); err == nil {
			t.Fatal("setup.sh ran despite the refusal")
		}
	})

	// Report shape.
	t.Run("a clean fixture: exit 0, a verdict line naming the root and the file count", func(t *testing.T) {
		t.Parallel()
		root := cicFixture(t)
		writeFile(t, root+"/"+flow, "# flow fixture\n`skills/other/SKILL.md`\n")
		r := cicRun(t, root, nil)
		first, _, _ := strings.Cut(r.out, "\n")
		if r.rc != 0 || !strings.HasPrefix(first, "INSTALLED-CITATIONS-OK:") {
			t.Fatalf("rc=%d out=%q err=%q", r.rc, r.out, r.errOut)
		}
	})
	t.Run("a clean fixture: the coverage fragment names the scanned member", func(t *testing.T) {
		t.Parallel()
		root := cicFixture(t)
		writeFile(t, root+"/"+flow, "# flow fixture\n`skills/other/SKILL.md`\n")
		r := cicRun(t, root, nil)
		_, rest, _ := strings.Cut(r.out, "\n")
		if !strings.Contains(rest, flow) {
			t.Fatalf("no coverage fragment in %q", r.out)
		}
	})
	t.Run("a fixture with two violations: exit 1, two path:line lines, a verdict naming 2", func(t *testing.T) {
		t.Parallel()
		root := cicFixture(t)
		writeFile(t, root+"/"+flow, "# flow fixture\n`.flow/project.md`\n`scripts/check-references.sh`\n")
		r := cicRun(t, root, nil)
		if r.rc != 1 || strings.Count("\n"+r.out, "\n"+flow+":") != 2 || !strings.Contains(r.out, "2 violation(s)") {
			t.Fatalf("rc=%d out=%q err=%q", r.rc, r.out, r.errOut)
		}
	})
	t.Run("undeclared zero coverage is a violation, not a silent pass", func(t *testing.T) {
		t.Parallel()
		root := cicFixture(t)
		writeFile(t, root+"/"+flow, "# flow fixture, genuinely no citations at all\n\nJust prose.\n")
		r := cicRun(t, root, nil)
		if r.rc != 1 || !hasLine(r.out, flow+":0: 0 checked, and not declared expected-zero") {
			t.Fatalf("rc=%d out=%q err=%q", r.rc, r.out, r.errOut)
		}
	})

	// The stand-ins for the harness-only labels (see the file comment).
	noSandboxLeft := func(t *testing.T, tmp string) {
		t.Helper()
		left, err := filepath.Glob(tmp + "/check-installed-citations.*")
		if err != nil || len(left) != 0 {
			t.Fatalf("sandbox left behind: %v %v", left, err)
		}
	}
	t.Run("a clean run leaves no sandbox behind", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		if r := cicRun(t, cicFixture(t), map[string]string{"TMPDIR": tmp}); r.rc != 0 {
			t.Fatalf("rc=%d out=%q err=%q", r.rc, r.out, r.errOut)
		}
		noSandboxLeft(t, tmp)
	})
	t.Run("a refused run leaves no sandbox behind", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		refused(t, cicRun(t, failingSetup(t), map[string]string{"TMPDIR": tmp}), "setup.sh")
		noSandboxLeft(t, tmp)
	})
	t.Run("setup.sh runs with HOME and its cwd inside the guard's own sandbox", func(t *testing.T) {
		t.Parallel()
		root, tmp := t.TempDir(), t.TempDir()
		log := root + "/log"
		writeExec(t, root+"/setup.sh", "#!/usr/bin/env bash\nprintf '%s %s %s\\n' \"$1\" \"$(cd \"$HOME\" && pwd -P)\" \"$(pwd -P)\" >> \"${0%/*}/log\"\n")
		writeFile(t, root+"/skills/x.md", "`skills/x.md`\n")
		cicRun(t, root, map[string]string{"TMPDIR": tmp})
		got, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(got), "\n"), "\n")
		if len(lines) != 3 || !strings.HasPrefix(lines[0], "global ") || !strings.HasPrefix(lines[1], "claude-code ") ||
			!strings.HasPrefix(lines[2], "zcode ") {
			t.Fatalf("setup.sh runs = %q, want global, then claude-code, then zcode", lines)
		}
		tmpPhys, _ := filepath.EvalSymlinks(tmp)
		for _, l := range lines {
			f := strings.Fields(l)
			if !strings.HasPrefix(f[2], tmpPhys+"/check-installed-citations.") || f[1] != f[2] {
				t.Errorf("run %q: want HOME the sandbox and cwd the sandbox under %s", l, tmp)
			}
		}
	})
}

// TestCicDecodeErrorMatchesPython holds cicDecodeError, a hand-written
// replica of CPython's UnicodeDecodeError text, to python3's own text for
// every malformed-sequence shape: a bad start byte, a bad second byte, a bad
// later continuation byte, and a truncated sequence.
func TestCicDecodeErrorMatchesPython(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	inputs := []string{
		"\xff", "a\x80b", "\xe2 ", "\xe0\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80",
		"\xe2\x82 ", "\xf0\x9f\x98 ", "\xf0\x9f ", "ab\xe2\x82", "\xf0\x9f\x98", "\xe2",
	}
	for _, in := range inputs {
		t.Run(fmt.Sprintf("%q", in), func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("python3", "-c", `import sys
try: sys.stdin.buffer.read().decode("utf-8")
except UnicodeDecodeError as e: print(str(e).removeprefix("'utf-8' codec "), end="")`)
			cmd.Stdin = strings.NewReader(in)
			want, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}
			if got := cicDecodeError([]byte(in)); got != string(want) {
				t.Fatalf("cicDecodeError(%q) = %q, python3 says %q", in, got, want)
			}
		})
	}
}

// TestCheckInstalledCitationsSIGINT interrupts the guard while its sandboxed
// setup.sh runs: the sandbox is gone and the guard died of that SIGINT, as
// the Python's try/finally left it on KeyboardInterrupt. The guard runs in a
// child copy of this test binary, since the SIGINT ends its process.
func TestCheckInstalledCitationsSIGINT(t *testing.T) {
	if root := os.Getenv("CIC_SIGINT_ROOT"); root != "" {
		os.Exit(checkInstalledCitations(nil, Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: root}, io.Discard, io.Discard))
	}
	t.Parallel()
	root := cicFixture(t)
	writeExec(t, root+"/setup.sh", "#!/usr/bin/env bash\ntouch \"$HOME/started\"\nexec sleep 30\n")
	tmp := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCheckInstalledCitationsSIGINT$")
	cmd.Env = append(os.Environ(), "CIC_SIGINT_ROOT="+root, "CHECK_INSTALLED_CITATIONS_ROOT="+root, "TMPDIR="+tmp)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	sandboxes := func() []string {
		m, _ := filepath.Glob(filepath.Join(tmp, cicName+".*"))
		return m
	}
	for deadline := time.Now().Add(20 * time.Second); ; time.Sleep(10 * time.Millisecond) {
		if m, _ := filepath.Glob(filepath.Join(tmp, cicName+".*", "started")); len(m) > 0 {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			t.Fatal("setup.sh never started")
		}
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	err := cmd.Wait()
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("guard exited %v, want death by SIGINT", err)
	}
	if ws := exit.Sys().(syscall.WaitStatus); !ws.Signaled() || ws.Signal() != syscall.SIGINT {
		t.Errorf("guard exit status %v, want death by SIGINT", ws)
	}
	if left := sandboxes(); len(left) > 0 {
		t.Errorf("SIGINT left the sandbox behind: %v", left)
	}
}
