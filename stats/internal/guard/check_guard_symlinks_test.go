package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-guard-symlinks.sh at d71a2327, one subtest
// per ok: label, nested under the harness's own case id. The bash harness ran
// the guard with CHECK_GUARD_SYMLINKS_ROOT on a sandboxed fixture tree and
// asserted on its exit, its stdout and its stderr; here the guard runs
// in-process with the same override on Env. Each fixture's case also carries
// an "output pinned" subtest: the whole stdout, stderr and exit the bash
// guard printed for that fixture at d71a2327, the fixture's paths replaced by
// <repo> (as given), <repo-phys> (resolved) and <outside>.

// gsMD turns a fixture written with ' for ` into the markdown the harness
// wrote: every harness body was a single-quoted bash string, so none carries
// a real '.
func gsMD(s string) string { return strings.ReplaceAll(s, "'", "`") }

const (
	gsPlainGuard = `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo ok`
	gsFixedDepthGuard = `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
echo "$REPO_ROOT"`
	// F5 — the same fixed-depth defect, spelled two other ways. dirname() and
	// a two-step cd chain answer the identical "one level above $SCRIPT_DIR"
	// question the literal `$SCRIPT_DIR/..` form does, and rule 4 must catch
	// all three.
	gsFixedDepthDirname = `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
echo "$REPO_ROOT"`
	gsFixedDepthCdChain = `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" && cd ..
REPO_ROOT="$(pwd)"
echo "$REPO_ROOT"`
	gsCleanFlowMD = `# flow fixture

Run the guard:

'''bash
check-foo.sh <worktree>
'''
`
)

// gsNew is new_repo: an empty root carrying scripts/ and skills/.
func gsNew(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mkdir(t, repo+"/scripts")
	mkdir(t, repo+"/skills")
	return repo
}

// gsGuard is add_real_guard.
func gsGuard(t *testing.T, repo, name, body string) {
	t.Helper()
	writeFile(t, repo+"/scripts/"+name, body+"\n")
}

// gsLink is link_guard: skills/<skill>/scripts/<name> -> ../../../scripts/<name>.
func gsLink(t *testing.T, repo, skill, name string) {
	t.Helper()
	mkdir(t, repo+"/skills/"+skill+"/scripts")
	symlink(t, "../../../scripts/"+name, repo+"/skills/"+skill+"/scripts/"+name)
}

// gsSkill is write_skill_md.
func gsSkill(t *testing.T, repo, skill, body string) {
	t.Helper()
	writeFile(t, repo+"/skills/"+skill+"/SKILL.md", body+"\n")
}

type gsRes struct {
	rc       int
	out, err string
}

// gsRun is run_guard: the guard over root, stdout and stderr apart.
func gsRun(t *testing.T, root string) gsRes {
	t.Helper()
	return gsRunEnv(t, root, nil)
}

// gsRunEnv is gsRun with vars set on top of the override.
func gsRunEnv(t *testing.T, root string, vars map[string]string) gsRes {
	t.Helper()
	all := map[string]string{"CHECK_GUARD_SYMLINKS_ROOT": root}
	for k, v := range vars {
		all[k] = v
	}
	var out, errb bytes.Buffer
	rc := checkGuardSymlinks(nil, crEnv(t.TempDir(), all), &out, &errb)
	return gsRes{rc, out.String(), errb.String()}
}

func gsCheck(t *testing.T, label string, ok bool, format string, a ...any) {
	t.Helper()
	t.Run(label, func(t *testing.T) {
		if !ok {
			t.Fatalf(format, a...)
		}
	})
}

func gsFirstLine(s string) string { return strings.SplitN(s, "\n", 2)[0] }

// gsOK is assert_ok: exit 0 and the OK verdict on the first line.
func gsOK(t *testing.T, r gsRes, label string) {
	t.Helper()
	gsCheck(t, label, r.rc == 0 && strings.HasPrefix(gsFirstLine(r.out), "GUARD-SYMLINKS-OK:"),
		"expected exit 0 and a GUARD-SYMLINKS-OK verdict, got rc=%d out=%s err=%s", r.rc, r.out, r.err)
}

// gsSilent is assert_silent: assert_ok, and at most a coverage breakdown
// after the verdict — never a violation line.
func gsSilent(t *testing.T, r gsRes, label string) {
	t.Helper()
	gsOK(t, r, label)
	n := strings.Count(r.out, "\n")
	gsCheck(t, label+": verdict plus at most a coverage breakdown", n == 1 || n == 2,
		"expected 1 or 2 stdout lines, got %d: %s", n, r.out)
}

// gsInvalid is assert_invalid: exit 1 and the INVALID verdict last.
func gsInvalid(t *testing.T, r gsRes, label string) {
	t.Helper()
	lines := strings.Split(strings.TrimRight(r.out, "\n"), "\n")
	gsCheck(t, label, r.rc == 1 && strings.HasPrefix(lines[len(lines)-1], "GUARD-SYMLINKS-INVALID:"),
		"expected exit 1 and a GUARD-SYMLINKS-INVALID verdict, got rc=%d out=%s err=%s", r.rc, r.out, r.err)
}

// gsReports is assert_reports.
func gsReports(t *testing.T, r gsRes, needle, label string) {
	t.Helper()
	gsCheck(t, label, strings.Contains(r.out, needle), "the report does not name %q: %s", needle, r.out)
}

// gsRefuses is assert_refuses: exit 2, nothing on stdout, the guard's own
// name on stderr.
func gsRefuses(t *testing.T, r gsRes, label string) {
	t.Helper()
	gsCheck(t, label+": exits 2", r.rc == 2, "expected exit 2, got rc=%d out=%s", r.rc, r.out)
	gsCheck(t, label+": writes nothing to stdout", r.out == "", "emitted a verdict line: %s", r.out)
	gsCheck(t, label+": names the failure on stderr", strings.Contains(r.err, "check-guard-symlinks: "),
		"no named message on stderr: %s", r.err)
}

// gsPin compares the whole result against the bash guard's at d71a2327,
// gsPins[<this subtest>].
func gsPin(t *testing.T, r gsRes, repo string, outside ...string) {
	t.Helper()
	got := r.out + "--- stderr\n" + r.err + fmt.Sprintf("--- exit %d\n", r.rc)
	for _, o := range outside {
		got = strings.ReplaceAll(got, o, "<outside>")
	}
	if phys, err := filepath.EvalSymlinks(repo); err == nil {
		got = strings.ReplaceAll(got, phys, "<repo-phys>")
	}
	got = strings.ReplaceAll(got, repo, "<repo>")
	want := gsPins[t.Name()]
	gsCheck(t, "output pinned", got == want, "output differs from the bash guard's\n got:\n%s\nwant:\n%s", got, want)
}

// gsBashParity runs the bash guard at d71a2327 over repo and requires the
// port's stdout, stderr and exit to match it byte for byte.
func gsBashParity(t *testing.T, repo string) {
	t.Helper()
	scripts := filepath.Join(t.TempDir(), "scripts")
	for _, rel := range []string{"check-guard-symlinks.sh", "lib/resolve-file.sh", "lib/coverage.sh"} {
		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/"+rel).Output()
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(scripts, rel), string(src))
	}
	cmd := exec.Command("bash", filepath.Join(scripts, "check-guard-symlinks.sh"))
	cmd.Env = append(os.Environ(), "CHECK_GUARD_SYMLINKS_ROOT="+repo)
	var bout, berr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &bout, &berr
	brc := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		brc = ee.ExitCode()
	}
	r := gsRun(t, repo)
	gsCheck(t, "matches the bash", r.rc == brc && r.out == bout.String() && r.err == berr.String(),
		"port rc=%d out=%q err=%q\nbash rc=%d out=%q err=%q", r.rc, r.out, r.err, brc, bout.String(), berr.String())
}

// gsCountLines is `grep -c <needle>` over the report.
func gsCountLines(out, needle string) int {
	n := 0
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, needle) {
			n++
		}
	}
	return n
}

func TestCheckGuardSymlinks(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		// 1. A clean tree — one real guard with no sibling dependency and no
		// fixed-depth root, invoked with a placeholder argument inside a bash
		// fence, correctly symlinked into the one skill that cites it.
		{"1", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-foo.sh")
			gsSkill(t, repo, "flow", gsMD(gsCleanFlowMD))
			r := gsRun(t, repo)
			gsSilent(t, r, "a clean tree passes with exactly one verdict line")
			gsPin(t, r, repo)
		}},
		// 2a. Rule 1: an entry that is a regular file, not a symlink.
		{"2a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			writeFile(t, repo+"/skills/flow/scripts/check-bogus.sh", "not a symlink\n")
			gsSkill(t, repo, "flow", "# fixture, no citations")
			r := gsRun(t, repo)
			gsInvalid(t, r, "a non-symlink entry under skills/*/scripts/ is a rule 1 violation")
			gsReports(t, r, "check-bogus.sh", "rule 1: names the offending entry")
			gsReports(t, r, "rule 1", "rule 1: names the rule")
			gsPin(t, r, repo)
		}},
		// 2a′. A __pycache__ directory is skipped under skills/*/scripts/ (not a
		// rule 1 violation) and under scripts/ (not counted as a guard).
		{"2a′", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-foo.sh")
			gsSkill(t, repo, "flow", gsMD(gsCleanFlowMD))
			clean := gsRun(t, repo).out
			writeFile(t, repo+"/skills/flow/scripts/__pycache__/check-foo.cpython-314.pyc", "bytecode\n")
			writeFile(t, repo+"/scripts/__pycache__/check-foo.cpython-314.pyc", "bytecode\n")
			r := gsRun(t, repo)
			gsSilent(t, r, "a __pycache__ directory under skills/*/scripts/ is not a rule 1 violation")
			gsCheck(t, "rule 1: __pycache__ under scripts/ and skills/*/scripts/ leaves the verdict and its count unchanged",
				r.out == clean, "__pycache__ changed the verdict: %q vs clean %q", r.out, clean)
			gsPin(t, r, repo)
		}},
		// 2b. A dangling symlink.
		{"2b", func(t *testing.T) {
			repo := gsNew(t)
			mkdir(t, repo+"/skills/flow/scripts")
			symlink(t, "../../../scripts/does-not-exist.sh", repo+"/skills/flow/scripts/check-missing.sh")
			gsSkill(t, repo, "flow", "# fixture, no citations")
			r := gsRun(t, repo)
			gsInvalid(t, r, "a dangling symlink under skills/*/scripts/ is a rule 1 violation")
			gsReports(t, r, "check-missing.sh", "rule 1: names the dangling entry")
			gsReports(t, r, "does not resolve", "rule 1: says it does not resolve")
			gsPin(t, r, repo)
		}},
		// 2c. An absolute-target symlink that still resolves.
		{"2c", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			mkdir(t, repo+"/skills/flow/scripts")
			symlink(t, repo+"/scripts/check-foo.sh", repo+"/skills/flow/scripts/check-foo.sh")
			gsSkill(t, repo, "flow", "# fixture, no citations")
			r := gsRun(t, repo)
			gsInvalid(t, r, "an absolute symlink target under skills/*/scripts/ is a rule 1 violation")
			gsReports(t, r, "is absolute", "rule 1: says the target is absolute")
			gsPin(t, r, repo)
		}},
		// 2d. F9 — an absolute target OUTSIDE the repository stops at the rule 1
		// violation: an unreadable off-repo target read by rule 4 would add a
		// second violation line. flow-settings, the declared expected-zero
		// skill, keeps the count about F9 alone.
		{"2d", func(t *testing.T) {
			repo := gsNew(t)
			mkdir(t, repo+"/skills/flow-settings/scripts")
			outside := t.TempDir() + "/check-guard-symlinks-outside"
			writeFile(t, outside, "unrelated content, unreadable\n")
			if err := os.Chmod(outside, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(outside, 0o644) })
			symlink(t, outside, repo+"/skills/flow-settings/scripts/check-outside.sh")
			gsSkill(t, repo, "flow-settings", "# fixture, no citations")
			r := gsRun(t, repo)
			gsInvalid(t, r, "an absolute off-repo symlink target is a rule 1 violation")
			gsReports(t, r, "is absolute", "rule 1 (F9): says the target is absolute")
			const label = "F9: exactly one violation line, not a second from rule 4 reading outside the repo"
			if os.Geteuid() == 0 {
				t.Run(label, func(t *testing.T) { t.Skip("running as root; mode 000 is still readable") })
			} else {
				n := 0
				for _, l := range strings.Split(strings.TrimRight(r.out, "\n"), "\n") {
					if !strings.HasPrefix(l, "GUARD-SYMLINKS-") {
						n++
					}
				}
				gsCheck(t, label, n == 1, "expected exactly 1 violation line, got %d: %s", n, r.out)
			}
			gsPin(t, r, repo, outside)
		}},
		// 3a. Rule 2: invoked in a bash fence, never symlinked in at all.
		{"3a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-baz.sh", gsPlainGuard)
			mkdir(t, repo+"/skills/flow/scripts")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
check-baz.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "an invoked guard with no symlink is a rule 2 violation")
			gsReports(t, r, "check-baz.sh", "rule 2: names the missing guard")
			gsReports(t, r, "rule 2", "rule 2: names the rule")
			gsPin(t, r, repo)
		}},
		// 3b. A sibling dependency, resolved from the required guard's own
		// source, missing from the skill that needs the guard beside it.
		{"3b", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-with-lib.sh", `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/helper.sh"`)
			writeFile(t, repo+"/scripts/lib/helper.sh", "true\n")
			gsLink(t, repo, "flow", "check-with-lib.sh")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
check-with-lib.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a required sibling with no symlink is a rule 2 violation")
			gsReports(t, r, "lib", "rule 2: names the missing sibling")
			gsReports(t, r, "sibling dependency", "rule 2: says it is a sibling dependency")
			gsPin(t, r, repo)
		}},
		// 3c. F6 — `./guard.sh <args>` and `bash guard.sh <args>`, both never
		// symlinked, are both invocations.
		{"3c", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-dotslash.sh", gsPlainGuard)
			gsGuard(t, repo, "check-viabash.sh", gsPlainGuard)
			mkdir(t, repo+"/skills/flow/scripts")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
./check-dotslash.sh <worktree>
'''

'''bash
bash check-viabash.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a ./guard.sh or bash guard.sh invocation with no symlink is a rule 2 violation")
			gsReports(t, r, "check-dotslash.sh", "rule 2 (F6): names the ./guard.sh form")
			gsReports(t, r, "check-viabash.sh", "rule 2 (F6): names the bash guard.sh form")
			gsPin(t, r, repo)
		}},
		// 3d. F7 — a stray, unpaired backtick earlier on the line must not drop
		// the real citation after it.
		{"3d", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-strayed.sh", gsPlainGuard)
			mkdir(t, repo+"/skills/flow/scripts")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

A stray ' mark appears here, then Run 'check-strayed.sh' for the real thing.
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a real citation after a stray backtick on the same line is still a rule 2 violation")
			gsReports(t, r, "check-strayed.sh", "rule 2 (F7): the citation after the stray backtick was not dropped")
			gsPin(t, r, repo)
		}},
		// 3e. F10 — two required guards sharing one sibling: the sibling is
		// required once, and both citing guards are still enforced.
		{"3e", func(t *testing.T) {
			repo := gsNew(t)
			const body = `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/shared.sh"`
			gsGuard(t, repo, "check-one.sh", body)
			gsGuard(t, repo, "check-two.sh", body)
			writeFile(t, repo+"/scripts/lib/shared.sh", "true\n")
			gsLink(t, repo, "flow", "check-one.sh")
			gsLink(t, repo, "flow", "check-two.sh")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
check-one.sh <worktree>
check-two.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "one sibling shared by two required guards is still a rule 2 violation")
			n := gsCountLines(r.out, "sibling dependency")
			gsCheck(t, "F10: the shared sibling is reported exactly once, not once per citing guard", n == 1,
				"expected exactly 1 sibling-dependency violation, got %d: %s", n, r.out)
			gsPin(t, r, repo)
		}},
		// 3f. Any token of a fenced command line: a pipeline segment, an &&
		// continuation and a command substitution.
		{"3f", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-pipe.sh", gsPlainGuard)
			gsGuard(t, repo, "check-cont.sh", gsPlainGuard)
			gsGuard(t, repo, "check-subst.sh", gsPlainGuard)
			mkdir(t, repo+"/skills/flow/scripts")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
git -C <worktree> diff --name-only <base>..HEAD | check-pipe.sh <worktree>
'''

'''bash
some-command <worktree> \
  && check-cont.sh <worktree> <project>
'''

'''bash
BASE="$(check-subst.sh <worktree>)"
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a guard named as a pipeline segment, && continuation or command substitution is a rule 2 violation")
			gsReports(t, r, "check-pipe.sh", "rule 2 (3f): names the pipeline-segment form")
			gsReports(t, r, "check-cont.sh", "rule 2 (3f): names the && continuation form")
			gsReports(t, r, "check-subst.sh", "rule 2 (3f): names the command-substitution form")
			gsPin(t, r, repo)
		}},
		// 3g. KAN-530 F6 — a prose Run of a .sh basename matching no guard.
		{"3g", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", gsMD(`# flow-settings fixture

Run 'check-typo.sh' before committing.
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a prose invocation of a .sh basename matching no guard is flagged (KAN-530 F6)")
			gsReports(t, r, "check-typo.sh", "rule 2 (KAN-530 F6): names the unknown basename")
			gsReports(t, r, "rule 2", "rule 2 (KAN-530 F6): names the rule")
			gsPin(t, r, repo)
		}},
		// 3h. KAN-530 F6 — the <placeholder> prose shape.
		{"3h", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", gsMD(`# flow-settings fixture

The invocation is 'check-ph-typo.sh <worktree>' as written above.
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a placeholder invocation of a .sh basename matching no guard is flagged (KAN-530 F6)")
			gsReports(t, r, "check-ph-typo.sh", "rule 2 (KAN-530 F6): names the unknown placeholder-form basename")
			gsPin(t, r, repo)
		}},
		// 3i. KAN-530 F6 — in the fence: a bare leading token, and the second
		// token after an interpreter.
		{"3i", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", gsMD(`# flow-settings fixture

'''bash
check-lead-typo.sh <worktree>
'''

'''bash
bash check-via-typo.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a fence invocation of a .sh basename matching no guard is flagged (KAN-530 F6)")
			gsReports(t, r, "check-lead-typo.sh", "rule 2 (KAN-530 F6): names the unknown leading-token basename")
			gsReports(t, r, "check-via-typo.sh", "rule 2 (KAN-530 F6): names the unknown interpreter-form basename")
			gsPin(t, r, repo)
		}},
		// 3j. KAN-530 F6 — the boundary the .sh shape buys: a non-invoking
		// unknown .sh name, and an invoking name without the .sh shape.
		{"3j", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", gsMD(`# flow-settings fixture

The 'check-nowhere.sh' guard reads the marker block. Run 'flow status' to see it.
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "a non-invoking unknown .sh name and an invoking non-.sh name stay prose (KAN-530 F6)")
			gsPin(t, r, repo)
		}},
		// 3k. KAN-530 F6 — unknown-name rows from two skills' files, each
		// reported exactly once at its own citing file.
		{"3k", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-known.sh", gsPlainGuard)
			gsLink(t, repo, "flow-zed", "check-known.sh")
			gsSkill(t, repo, "flow-settings", gsMD(`# flow-settings fixture

Run 'check-typo.sh' before committing.
`))
			gsSkill(t, repo, "flow-zed", gsMD(`# flow-zed fixture

'''bash
check-known.sh <worktree>
bash check-other-typo.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "unknown-name rows across two skills are each reported exactly once, at their own citing file (KAN-530 F6)")
			n := gsCountLines(r.out, "check-typo.sh is invoked")
			gsCheck(t, "3k: the flow-settings row is reported exactly once", n == 1, "expected exactly 1 check-typo.sh row, got %d: %s", n, r.out)
			n = gsCountLines(r.out, "check-other-typo.sh is invoked")
			gsCheck(t, "3k: the flow-zed row is reported exactly once", n == 1, "expected exactly 1 check-other-typo.sh row, got %d: %s", n, r.out)
			gsReports(t, r, "skills/flow-settings/SKILL.md:3: check-typo.sh", "3k: the flow-settings row names its own citing file")
			gsReports(t, r, "skills/flow-zed/SKILL.md:5: check-other-typo.sh", "3k: the flow-zed row names its own citing file")
			gsPin(t, r, repo)
		}},
		// 3l. A subcommand word between the basename and its first
		// <placeholder> — the shape skills/flow/implement.md and
		// review-panel.md use for guard-autosquash.sh
		// ('guard-autosquash.sh targets <worktree> <task-sha>'), opening a
		// line with no Run/Invoke before it. The placeholder form used to
		// demand the '<' right after the basename, so this citation fell out
		// of the required set and the missing symlink read GUARD-SYMLINKS-OK.
		{"3l", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-subcmd.sh", gsPlainGuard)
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'check-subcmd.sh targets <worktree> <task-sha>' asserts the target.
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a placeholder invocation with a subcommand word, symlink absent, is a rule 2 violation")
			gsReports(t, r, "skills/flow/scripts/check-subcmd.sh", "rule 2 (3l): names the missing symlink")
			gsPin(t, r, repo)
		}},
		// 3m. KAN-860 F6 — a placeholder invocation whose backtick span wraps
		// onto the next line (review-panel-fix-round.md's
		// check-task-commit-fields.sh call). The per-line scan found no closing
		// backtick and dropped the span, so the deleted symlink read
		// GUARD-SYMLINKS-OK.
		{"3m", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-wrap.sh", gsPlainGuard)
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

The close re-runs the guard:
'check-wrap.sh <worktree> <task-id>
<name>' for every task.
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a wrapped placeholder invocation, symlink absent, is a rule 2 violation")
			gsReports(t, r, "skills/flow/scripts/check-wrap.sh", "rule 2 (3m): names the missing symlink")
			gsPin(t, r, repo)
		}},
		// 4a. Rule 3: the repository-relative form inside a bash fence.
		{"4a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-qux.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-qux.sh")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
scripts/check-qux.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a repository-relative path inside a bash fence is a rule 3 violation")
			gsReports(t, r, "rule 3", "rule 3: names the rule")
			gsPin(t, r, repo)
		}},
		// 4b. The same path in a plain descriptive sentence is prose.
		{"4b", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-qux.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-qux.sh")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

Run the guard:

'''bash
check-qux.sh <worktree>
'''

'scripts/check-qux.sh' reads only the marker block. It never parses the table.
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "a repository-relative path in a descriptive sentence is prose, not a rule 3 violation")
			gsPin(t, r, repo)
		}},
		// 4c. F3 — the project-configured guards are exempt from rule 3 in a
		// fence and in imperative prose.
		{"4c", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-references.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-references.sh")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

Run 'scripts/check-plan-provenance.sh' before committing.

'''bash
scripts/check-references.sh
'''
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "a project-configured guard keeps its repository-relative path without tripping rule 3 (F3)")
			gsPin(t, r, repo)
		}},
		// 4d. F8 — a shellsession fence is not scanned as bash/sh/zsh.
		{"4d", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", gsMD(`# flow-settings fixture

'''shellsession
scripts/check-qux.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "a ```shellsession fence is not scanned as bash/sh/zsh (F8)")
			gsPin(t, r, repo)
		}},
		// 5a. Rule 4: a shipped guard deriving $SCRIPT_DIR/.. .
		{"5a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-fixed-depth.sh", gsFixedDepthGuard)
			gsLink(t, repo, "flow", "check-fixed-depth.sh")
			gsSkill(t, repo, "flow", gsMD(strings.ReplaceAll(gsCleanFlowMD, "check-foo.sh", "check-fixed-depth.sh")))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a shipped guard deriving $SCRIPT_DIR/.. is a rule 4 violation")
			gsReports(t, r, "check-fixed-depth.sh", "rule 4: names the offending guard")
			gsReports(t, r, "rule 4", "rule 4: names the rule")
			gsPin(t, r, repo)
		}},
		// 5b. An unshipped guard may keep $SCRIPT_DIR/.. .
		{"5b", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-project-only.sh", gsFixedDepthGuard)
			gsSkill(t, repo, "flow-settings", "# fixture, no citations")
			r := gsRun(t, repo)
			gsSilent(t, r, "an unshipped guard keeping $SCRIPT_DIR/.. is not a rule 4 violation")
			gsPin(t, r, repo)
		}},
		// 5c. F5 — the dirname() spelling.
		{"5c", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-fixed-dirname.sh", gsFixedDepthDirname)
			gsLink(t, repo, "flow", "check-fixed-dirname.sh")
			gsSkill(t, repo, "flow", gsMD(strings.ReplaceAll(gsCleanFlowMD, "check-foo.sh", "check-fixed-dirname.sh")))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a shipped guard deriving dirname($SCRIPT_DIR) is a rule 4 violation (F5)")
			gsReports(t, r, "check-fixed-dirname.sh", "rule 4 (F5): names the offending guard (dirname form)")
			gsReports(t, r, "rule 4", "rule 4 (F5): names the rule (dirname form)")
			gsPin(t, r, repo)
		}},
		// 5d. F5 — the `cd $SCRIPT_DIR && cd ..` spelling.
		{"5d", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-fixed-cdchain.sh", gsFixedDepthCdChain)
			gsLink(t, repo, "flow", "check-fixed-cdchain.sh")
			gsSkill(t, repo, "flow", gsMD(strings.ReplaceAll(gsCleanFlowMD, "check-foo.sh", "check-fixed-cdchain.sh")))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a shipped guard deriving cd $SCRIPT_DIR && cd .. is a rule 4 violation (F5)")
			gsReports(t, r, "check-fixed-cdchain.sh", "rule 4 (F5): names the offending guard (cd-chain form)")
			gsReports(t, r, "rule 4", "rule 4 (F5): names the rule (cd-chain form)")
			gsPin(t, r, repo)
		}},
		// F4a. A delegating skill whose delegate requires a guard it does not
		// carry.
		{"F4a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-deleg.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-deleg.sh")
			gsSkill(t, repo, "flow", gsMD(gsPresenceMD))
			gsSkill(t, repo, "flow-deleg", gsMD(gsDelegMD))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a delegating skill whose delegate requires a guard it does not carry is a rule 2 violation (F4)")
			gsReports(t, r, "check-deleg.sh", "rule 2 (F4): names the guard required by delegation")
			gsReports(t, r, "flow-deleg", "rule 2 (F4): names the delegating skill")
			gsPin(t, r, repo)
		}},
		// F4b. The same, carrying the symlink.
		{"F4b", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-deleg.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-deleg.sh")
			gsLink(t, repo, "flow-deleg", "check-deleg.sh")
			gsSkill(t, repo, "flow", gsMD(gsPresenceMD))
			gsSkill(t, repo, "flow-deleg", gsMD(gsDelegMD))
			r := gsRun(t, repo)
			gsSilent(t, r, "a delegating skill carrying the delegated guard is not a rule 2 violation (F4)")
			gsPin(t, r, repo)
		}},
		// F4c. An ordinary cross-command citation outside the presence
		// paragraph is not delegation.
		{"F4c", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-deleg.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-deleg.sh")
			gsSkill(t, repo, "flow", gsMD(gsPresenceMD))
			gsSkill(t, repo, "flow-settings", gsMD(`# fixture standing in for flow-settings

This command explains what '/flow' does elsewhere. It invokes nothing
of its own.
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "an ordinary cross-command citation outside the presence paragraph is not delegation (F4)")
			gsPin(t, r, repo)
		}},
		// F12. A skill carrying a real symlink but citing nothing rule 2 can
		// see is a coverage violation, not silence.
		{"F12", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-shadow.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-shadow.sh")
			gsLink(t, repo, "flow-shadow", "check-shadow.sh")
			gsSkill(t, repo, "flow", gsMD(strings.ReplaceAll(gsPresenceMD, "check-deleg.sh", "check-shadow.sh")))
			gsSkill(t, repo, "flow-shadow", gsMD(`# flow-shadow fixture — KANs own shape, generalized

See '/flow' for details on this behavior. This skill delegates, but not
in a form the classifier above recognizes.
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a skill carrying a real symlink but citing nothing rule 2 can see is a coverage violation, not silence (F12)")
			gsReports(t, r, "flow-shadow", "F12: names the under-covered skill")
			gsReports(t, r, "coverage", "F12: reports it as a coverage finding")
			gsReports(t, r, "0 checked", "F12: reports the zero count")
			gsReports(t, r, "not declared expected-zero", "F12: says it was never declared")
			gsPin(t, r, repo)
		}},
		// F13. A declared expected-zero member reports its zero without failing.
		{"F13", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", "# flow-settings fixture, no citations — declared expected-zero in the guard's own source")
			r := gsRun(t, repo)
			gsSilent(t, r, "a declared expected-zero member reports its zero without failing (F13)")
			gsReports(t, r, "declared", "F13: the coverage breakdown marks the zero as declared")
			gsPin(t, r, repo)
		}},
		// 6a. An unreadable skills/ directory — no verdict line, ever.
		{"6a", func(t *testing.T) {
			const label = "an unreadable skills/ directory"
			if os.Geteuid() == 0 {
				t.Run(label, func(t *testing.T) { t.Skip("running as root; mode 000 is still readable") })
				return
			}
			repo := gsNew(t)
			if err := os.Chmod(repo+"/skills", 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(repo+"/skills", 0o755) })
			r := gsRun(t, repo)
			gsRefuses(t, r, label)
			gsPin(t, r, repo)
		}},
		// 6b. A root that is not a directory at all.
		{"6b", func(t *testing.T) {
			notADir := t.TempDir() + "/check-guard-symlinks-notadir"
			writeFile(t, notADir, "")
			r := gsRun(t, notADir)
			gsRefuses(t, r, "a root that is not a directory")
			gsPin(t, r, notADir)
		}},
		// 6c–6e. resolve_file's own contract (F9, F16), on the Go twin the
		// guard calls (resolvefile.go; TestResolveFileParity pins it to the
		// bash library).
		{"6c", func(t *testing.T) {
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			want, _ := filepath.EvalSymlinks(cwd)
			got, _ := resolveFile(".")
			gsCheck(t, "resolve_file '.' resolves to the cwd itself, no spurious trailing '.' (F9)", got == want,
				"resolve_file '.' = %q, want %q", got, want)
			dir := t.TempDir()
			noSlash, _ := resolveFile(dir)
			withSlash, _ := resolveFile(dir + "/")
			gsCheck(t, "resolve_file with a trailing slash matches the same path without one, leaf not dropped (F9)",
				withSlash == noSlash && withSlash != "", "resolve_file with a trailing slash = %q, want %q", withSlash, noSlash)
		}},
		{"6d", func(t *testing.T) {
			root, _ := resolveFile("/")
			gsCheck(t, "resolve_file '/' resolves to '/', not '//' (F16)", root == "/", "resolve_file '/' = %q, want '/'", root)
			multi, _ := resolveFile("///")
			gsCheck(t, "resolve_file '///' collapses to '/', not '//' (F16)", multi == "/", "resolve_file '///' = %q, want '/'", multi)
		}},
		{"6e", func(t *testing.T) {
			const label = "resolve_file /tmp (a root-parented symlink) resolves without a doubled leading slash (F16)"
			if fi, err := os.Lstat("/tmp"); err != nil || fi.Mode()&os.ModeSymlink == 0 {
				t.Run(label, func(t *testing.T) { t.Skip("/tmp is not a symlink on this platform") })
				return
			}
			got, _ := resolveFile("/tmp")
			want, _ := filepath.EvalSymlinks("/tmp")
			gsCheck(t, label, got == want && !strings.HasPrefix(got, "//"), "resolve_file /tmp = %q, want %q", got, want)
		}},
		// 7. The agents repository itself, read from its real path: the guard
		// must exit 0 against it.
		{"7", func(t *testing.T) {
			root, err := filepath.Abs("../../..")
			if err != nil {
				t.Fatal(err)
			}
			gsOK(t, gsRun(t, root), "the agents repository's own tree validates cleanly")
		}},
		// 8a. Rule 5: a symlink directly under a skill directory's top level.
		{"8a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsLink(t, repo, "flow-self-review", "check-foo.sh")
			gsSkill(t, repo, "flow-self-review", gsMD(strings.ReplaceAll(strings.ReplaceAll(gsPresenceMD,
				"check-deleg.sh", "check-foo.sh"), "# flow fixture", "# flow-self-review fixture")))
			writeFile(t, repo+"/skills/flow-self-review/engineering-principles.md", "principles\n")
			gsSkill(t, repo, "flow-settings", "# fixture, no citations")
			symlink(t, "../flow-self-review/engineering-principles.md", repo+"/skills/flow-settings/engineering-principles.md")
			r := gsRun(t, repo)
			gsInvalid(t, r, "a symlink directly under a skill directory's top level is a rule 5 violation")
			gsCheck(t, "rule 5: the report names both the symlink's path and its target",
				strings.Contains(r.out, "engineering-principles.md") && strings.Contains(r.out, "../flow-self-review/engineering-principles.md"),
				"the report does not name both the symlink's path and its target: %s", r.out)
			gsPin(t, r, repo)
		}},
		// 8b. A symlink under skills/<skill>/scripts/ is rule 1's, not rule 5's.
		{"8b", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsLink(t, repo, "flow-status", "check-foo.sh")
			gsSkill(t, repo, "flow-status", gsMD(strings.ReplaceAll(gsCleanFlowMD, "# flow fixture", "# flow-status fixture")))
			r := gsRun(t, repo)
			gsSilent(t, r, "a symlink under skills/<skill>/scripts/ is not a rule 5 violation")
			gsPin(t, r, repo)
		}},
		// 8c. Only regular files and directories: no rule 5 finding.
		{"8c", func(t *testing.T) {
			repo := gsNew(t)
			gsSkill(t, repo, "flow-settings", "# fixture, no citations")
			r := gsRun(t, repo)
			gsSilent(t, r, "a skill directory with only regular files and directories is not a rule 5 violation")
			gsPin(t, r, repo)
		}},
		// 8d. Rule 5 sees skills/flow-contracts/ too.
		{"8d", func(t *testing.T) {
			repo := gsNew(t)
			writeFile(t, repo+"/skills/flow-contracts/pipeline.md", "contract\n")
			writeFile(t, repo+"/other-file.txt", "other\n")
			symlink(t, "../../other-file.txt", repo+"/skills/flow-contracts/evil-link")
			r := gsRun(t, repo)
			gsInvalid(t, r, "a symlink directly under skills/flow-contracts/ is a rule 5 violation")
			gsReports(t, r, "evil-link", "rule 5: names the symlink under skills/flow-contracts/")
			gsPin(t, r, repo)
		}},
		// 8e. Rules 1-4 keep ignoring flow-contracts.
		{"8e", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-foo.sh")
			gsSkill(t, repo, "flow", gsMD(gsCleanFlowMD))
			writeFile(t, repo+"/skills/flow-contracts/scripts/not-a-symlink.sh", "")
			gsSkill(t, repo, "flow-contracts", "# fixture, no citations")
			r := gsRun(t, repo)
			gsSilent(t, r, "a non-symlink entry under skills/flow-contracts/scripts/ is not a rule 1 violation — flow-contracts stays out of rules 1-4's scan")
			gsPin(t, r, repo)
		}},
		// 9a. Rule 6: a carried .sh symlink no citation requires.
		{"9a", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-listed.sh", gsPlainGuard)
			gsLink(t, repo, "flow-plan", "check-listed.sh")
			gsSkill(t, repo, "flow-plan", gsMD(`# flow-plan fixture

The guard presence list: 'check-listed.sh' (kickoff).
`))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a carried .sh symlink no citation requires is a rule 6 violation")
			gsReports(t, r, "rule 6", "rule 6: names the rule")
			gsReports(t, r, "skills/flow-plan/scripts/check-listed.sh", "rule 6: names the carried symlink")
			gsPin(t, r, repo)
		}},
		// 9b. The same symlink, cited in an invoking position.
		{"9b", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-listed.sh", gsPlainGuard)
			gsLink(t, repo, "flow-plan", "check-listed.sh")
			gsSkill(t, repo, "flow-plan", gsMD(`# flow-plan fixture

Every guard this skill runs: 'check-listed.sh <project>' (kickoff).
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "a carried .sh symlink its own citation requires is not a rule 6 violation")
			gsPin(t, r, repo)
		}},
		// 9c. Sibling-required.
		{"9c", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-with-sib.sh", `#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/check-sibling.sh"`)
			gsGuard(t, repo, "check-sibling.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-with-sib.sh")
			gsLink(t, repo, "flow", "check-sibling.sh")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
check-with-sib.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "a carried .sh symlink a sibling dependency requires is not a rule 6 violation")
			gsPin(t, r, repo)
		}},
		// 9d. The lib/ directory symlink and a .py twin are out of rule 6.
		{"9d", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-py.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-py.sh")
			writeFile(t, repo+"/scripts/lib/helper.sh", "true\n")
			symlink(t, "../../../scripts/lib", repo+"/skills/flow/scripts/lib")
			writeFile(t, repo+"/scripts/twin.py", "x\n")
			symlink(t, "../../../scripts/twin.py", repo+"/skills/flow/scripts/twin.py")
			gsSkill(t, repo, "flow", gsMD(`# flow fixture

'''bash
check-py.sh <worktree>
'''
`))
			r := gsRun(t, repo)
			gsSilent(t, r, "an uncited lib directory symlink or .py twin is not a rule 6 violation")
			gsPin(t, r, repo)
		}},
		// 9e. A dangling .sh symlink is rule 1's finding, never rule 6's.
		{"9e", func(t *testing.T) {
			repo := gsNew(t)
			mkdir(t, repo+"/skills/flow/scripts")
			symlink(t, "../../../scripts/check-absent.sh", repo+"/skills/flow/scripts/check-absent.sh")
			gsSkill(t, repo, "flow", "# fixture, no citations")
			r := gsRun(t, repo)
			gsInvalid(t, r, "a dangling .sh symlink is a rule 1 violation")
			gsCheck(t, "rule 6: a dangling symlink is rule 1's finding alone", !strings.Contains(r.out, "rule 6"),
				"a dangling symlink is rule 1's finding alone: %s", r.out)
			gsPin(t, r, repo)
		}},
		// 9f. DECLARED_RULE6: a declared pair is exempt, an undeclared
		// neighbour still violates.
		{"9f", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-visual-verification.sh", gsPlainGuard)
			gsGuard(t, repo, "check-undeclared.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-visual-verification.sh")
			gsLink(t, repo, "flow", "check-undeclared.sh")
			gsSkill(t, repo, "flow", "# fixture, no citations")
			r := gsRun(t, repo)
			gsInvalid(t, r, "a declared pair is exempt from rule 6 while an undeclared sibling still violates (9f)")
			gsCheck(t, "rule 6 (9f): the declared pair is not reported", !strings.Contains(r.out, "check-visual-verification.sh"),
				"the declared pair must not be reported: %s", r.out)
			gsReports(t, r, "check-undeclared.sh", "rule 6 (9f): names the undeclared sibling")
			gsPin(t, r, repo)
		}},
		// Locale-sensitive ordering: every `sort` the bash ran orders the
		// report under the caller's collation — mixed-case skill names put
		// the coverage fragment in a different order under en_US.UTF-8 than
		// under C.
		{"collation", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			for _, s := range []string{"flow-b", "Flow-a", "flow-C"} {
				gsLink(t, repo, s, "check-foo.sh")
				gsSkill(t, repo, s, gsMD(gsCleanFlowMD))
			}
			for _, lcAll := range []string{"en_US.UTF-8", "C"} {
				t.Run(lcAll, func(t *testing.T) {
					gsPin(t, gsRunEnv(t, repo, map[string]string{"LC_ALL": lcAll}), repo)
				})
			}
		}},
		// An empty CHECK_GUARD_SYMLINKS_ROOT refuses rather than falling back.
		{"empty root override", func(t *testing.T) {
			r := gsRun(t, "")
			gsRefuses(t, r, "an empty CHECK_GUARD_SYMLINKS_ROOT")
		}},
		// Beyond the harness (panel review of task 4): awk ended a line at its
		// first NUL, so a citation, fenced call or delegate after one was never
		// seen. Each fixture runs the bash at d71a2327 live beside the port.
		{"port: a prose citation after a NUL is not seen", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsGuard(t, repo, "check-bar.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-foo.sh")
			gsSkill(t, repo, "flow", gsMD(gsCleanFlowMD+"\nRun\x00 'check-bar.sh <worktree>' now."))
			gsBashParity(t, repo)
		}},
		{"port: a fenced call after a NUL is not seen", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-foo.sh", gsPlainGuard)
			gsGuard(t, repo, "check-bar.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-foo.sh")
			gsSkill(t, repo, "flow", gsMD(gsCleanFlowMD+"\n'''bash\ntrue\x00 check-bar.sh <worktree>\n'''"))
			gsBashParity(t, repo)
		}},
		{"port: a delegate after a NUL is not seen", func(t *testing.T) {
			repo := gsNew(t)
			gsGuard(t, repo, "check-deleg.sh", gsPlainGuard)
			gsLink(t, repo, "flow", "check-deleg.sh")
			gsSkill(t, repo, "flow", gsMD(gsPresenceMD))
			gsSkill(t, repo, "flow-deleg", gsMD("# flow-deleg fixture\n\n**Check guard presence.** Confirm every guard named in\x00 '/flow' own\npresence checks is present."))
			gsBashParity(t, repo)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}

const (
	gsPresenceMD = `# flow fixture

**Check guard presence.** Confirm every guard this command invokes:

'''bash
check-deleg.sh <worktree>
'''
`
	gsDelegMD = `# flow-deleg fixture, a flow-fast stand-in

**Check guard presence.** Confirm every guard named in '/flow' own
presence checks is present in 'skills/flow-deleg/scripts/'.
`
)

// gsPins is each fixture's whole result from the bash guard at d71a2327,
// captured by running scripts/check-guard-symlinks.sh over the same fixture.
var gsPins = map[string]string{
	"TestCheckGuardSymlinks/1":                     "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/2a":                    "<repo>/skills/flow/scripts/check-bogus.sh:0: not a symlink — every entry under skills/*/scripts/ must be a relative symlink into the repository's scripts/ directory (rule 1)\nflow:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/2a′":                   "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/2b":                    "<repo>/skills/flow/scripts/check-missing.sh:0: symlink does not resolve (target '../../../scripts/does-not-exist.sh') (rule 1)\nflow:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/2c":                    "<repo>/skills/flow/scripts/check-foo.sh:0: symlink target is absolute ('<repo>/scripts/check-foo.sh') — every entry under skills/*/scripts/ must be a RELATIVE symlink, or it bakes this machine's checkout path into the repository (rule 1)\nflow:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/2d":                    "<repo>/skills/flow-settings/scripts/check-outside.sh:0: symlink target is absolute ('<outside>') — every entry under skills/*/scripts/ must be a RELATIVE symlink, or it bakes this machine's checkout path into the repository (rule 1)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3a":                    "<repo>/skills/flow/SKILL.md:4: invokes check-baz.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-baz.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3b":                    "<repo>/scripts/check-with-lib.sh:0: lib is a sibling dependency this guard resolves from its own directory, but skill \"flow\" carries no symlink at skills/flow/scripts/lib (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3c":                    "<repo>/skills/flow/SKILL.md:4: invokes check-dotslash.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-dotslash.sh (rule 2)\n<repo>/skills/flow/SKILL.md:8: invokes check-viabash.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-viabash.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3d":                    "<repo>/skills/flow/SKILL.md:3: invokes check-strayed.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-strayed.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3e":                    "<repo>/scripts/check-one.sh:0: lib is a sibling dependency this guard resolves from its own directory, but skill \"flow\" carries no symlink at skills/flow/scripts/lib (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3f":                    "<repo>/skills/flow/SKILL.md:4: invokes check-pipe.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-pipe.sh (rule 2)\n<repo>/skills/flow/SKILL.md:9: invokes check-cont.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-cont.sh (rule 2)\n<repo>/skills/flow/SKILL.md:13: invokes check-subst.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-subst.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 3 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3g":                    "<repo>/skills/flow-settings/SKILL.md:3: check-typo.sh is invoked here, but no guard named check-typo.sh exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3h":                    "<repo>/skills/flow-settings/SKILL.md:3: check-ph-typo.sh is invoked here, but no guard named check-ph-typo.sh exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3i":                    "<repo>/skills/flow-settings/SKILL.md:4: check-lead-typo.sh is invoked here, but no guard named check-lead-typo.sh exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)\n<repo>/skills/flow-settings/SKILL.md:8: check-via-typo.sh is invoked here, but no guard named check-via-typo.sh exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3j":                    "GUARD-SYMLINKS-OK: <repo> — 0 guard(s) across 1 skill(s) validated\n  flow-settings 0 (declared: invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage)\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/3k":                    "<repo>/skills/flow-settings/SKILL.md:3: check-typo.sh is invoked here, but no guard named check-typo.sh exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)\n<repo>/skills/flow-zed/SKILL.md:5: check-other-typo.sh is invoked here, but no guard named check-other-typo.sh exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3l":                    "<repo>/skills/flow/SKILL.md:3: invokes check-subcmd.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-subcmd.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/3m":                    "<repo>/skills/flow/SKILL.md:4: invokes check-wrap.sh here, but skill \"flow\" carries no symlink at skills/flow/scripts/check-wrap.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/4a":                    "<repo>/skills/flow/SKILL.md:4: a repository-relative scripts/<name> path appears in an invoking position (a bash-fenced command, or an imperative Run/Invoke/Execute) — name the guard by basename instead, per the resolution rule in skills/flow-contracts/pipeline.md (rule 3)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/4b":                    "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/4c":                    "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/4d":                    "GUARD-SYMLINKS-OK: <repo> — 0 guard(s) across 1 skill(s) validated\n  flow-settings 0 (declared: invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage)\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/5a":                    "<repo-phys>/scripts/check-fixed-depth.sh:4: derives a repository root as $SCRIPT_DIR/.. — this guard is reachable from more than one directory once shipped, and a fixed number of levels above itself silently answers a skill directory instead of the repository root (rule 4)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/5b":                    "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow-settings 0 (declared: invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage)\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/5c":                    "<repo-phys>/scripts/check-fixed-dirname.sh:4: derives a repository root as $SCRIPT_DIR/.. — this guard is reachable from more than one directory once shipped, and a fixed number of levels above itself silently answers a skill directory instead of the repository root (rule 4)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/5d":                    "<repo-phys>/scripts/check-fixed-cdchain.sh:4: derives a repository root as $SCRIPT_DIR/.. — this guard is reachable from more than one directory once shipped, and a fixed number of levels above itself silently answers a skill directory instead of the repository root (rule 4)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/6a":                    "--- stderr\ncheck-guard-symlinks: <repo>/skills is not a readable directory — cannot scan\n--- exit 2\n",
	"TestCheckGuardSymlinks/6b":                    "--- stderr\ncheck-guard-symlinks: <repo> is not a readable directory — cannot scan\n--- exit 2\n",
	"TestCheckGuardSymlinks/8a":                    "<repo>/skills/flow-settings/engineering-principles.md:0: symlink directly under a skill directory (target '../flow-self-review/engineering-principles.md') — every symlink a skill carries must sit under skills/<skill>/scripts/, never at the skill directory's own top level (rule 5)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/8b":                    "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow-status 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/8c":                    "GUARD-SYMLINKS-OK: <repo> — 0 guard(s) across 1 skill(s) validated\n  flow-settings 0 (declared: invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage)\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/8d":                    "<repo>/skills/flow-contracts/evil-link:0: symlink directly under a skill directory (target '../../other-file.txt') — every symlink a skill carries must sit under skills/<skill>/scripts/, never at the skill directory's own top level (rule 5)\ncoverage:0: no member was ever recorded — this guard checked nothing (empty corpus)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/8e":                    "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/9a":                    "<repo>/skills/flow-plan/scripts/check-listed.sh:0: carried but required by nothing — no citation in this skill's own text, no delegation and no sibling dependency names check-listed.sh, so this symlink is dead weight the next cleanup will prune (rule 6)\nflow-plan:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/9b":                    "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 1 skill(s) validated\n  flow-plan 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/9c":                    "GUARD-SYMLINKS-OK: <repo> — 2 guard(s) across 1 skill(s) validated\n  flow 2\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/9d":                    "GUARD-SYMLINKS-OK: <repo> — 3 guard(s) across 1 skill(s) validated\n  flow 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/9e":                    "<repo>/skills/flow/scripts/check-absent.sh:0: symlink does not resolve (target '../../../scripts/check-absent.sh') (rule 1)\nflow:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/9f":                    "<repo>/skills/flow/scripts/check-undeclared.sh:0: carried but required by nothing — no citation in this skill's own text, no delegation and no sibling dependency names check-undeclared.sh, so this symlink is dead weight the next cleanup will prune (rule 6)\nflow:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/F12":                   "<repo>/skills/flow-shadow/scripts/check-shadow.sh:0: carried but required by nothing — no citation in this skill's own text, no delegation and no sibling dependency names check-shadow.sh, so this symlink is dead weight the next cleanup will prune (rule 6)\nflow-shadow:0: 0 checked, and not declared expected-zero (coverage)\nGUARD-SYMLINKS-INVALID: <repo> — 2 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/F13":                   "GUARD-SYMLINKS-OK: <repo> — 0 guard(s) across 1 skill(s) validated\n  flow-settings 0 (declared: invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage)\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/F4a":                   "<repo>/skills/flow/SKILL.md:6: invokes check-deleg.sh here, and skill \"flow-deleg\" delegates to \"flow\"'s own guard presence check (task 4's convention), but carries no symlink at skills/flow-deleg/scripts/check-deleg.sh (rule 2)\nGUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)\n--- stderr\n--- exit 1\n",
	"TestCheckGuardSymlinks/F4b":                   "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 2 skill(s) validated\n  flow 1 · flow-deleg 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/F4c":                   "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 2 skill(s) validated\n  flow 1 · flow-settings 0 (declared: invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage)\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/collation/C":           "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 3 skill(s) validated\n  Flow-a 1 · flow-C 1 · flow-b 1\n--- stderr\n--- exit 0\n",
	"TestCheckGuardSymlinks/collation/en_US.UTF-8": "GUARD-SYMLINKS-OK: <repo> — 1 guard(s) across 3 skill(s) validated\n  Flow-a 1 · flow-b 1 · flow-C 1\n--- stderr\n--- exit 0\n",
}

// TestShimSiblingsDeclared runs rule 2 over each of the ten shims KAN-841
// wrote, as they stand in this checkout's scripts/: a skill carrying the shim
// alone must be told it also needs lib/ — the loader every shim sources — and
// check-finish-preflight's check-worktree-location.sh, which its Go guard
// execs from beside the shim. Rule 2 derives both from the shim's
// `$SCRIPT_DIR/<name>` spellings; a shim reaching a sibling any other way
// ships a skill that passes this guard and then cannot run.
func TestShimSiblingsDeclared(t *testing.T) {
	t.Parallel()
	scripts := tcfScriptsDir(t)
	shims := map[string][]string{
		"check-stage-mark-calls.sh":          {"lib"},
		"check-guard-symlinks.sh":            {"lib"},
		"check-dispatch-paragraphs.sh":       {"lib"},
		"mutate-and-verify.sh":               {"lib"},
		"check-base-moved.sh":                {"lib"},
		"check-panel-fix-single-dispatch.sh": {"lib"},
		"prove-reproducer.sh":                {"lib"},
		"check-finish-preflight.sh":          {"lib", "check-worktree-location.sh"},
		"kickoff-worktree.sh":                {"lib", "check-worktree-location.sh", "project-get.sh"},
		"fold-fixup.sh":                      {"lib", "guard-autosquash.sh"},
		"commit-archive.sh":                  {"lib"},
		"remove-change-worktrees.sh":         {"lib", "check-worktree-processes.sh"},
	}
	for shim, siblings := range shims {
		t.Run(shim, func(t *testing.T) {
			t.Parallel()
			b, err := os.ReadFile(scripts + "/" + shim)
			if err != nil {
				t.Fatal(err)
			}
			repo := gsNew(t)
			gsGuard(t, repo, shim, string(b))
			gsLink(t, repo, "flow", shim)
			gsSkill(t, repo, "flow", gsMD("# flow fixture\n\n'''bash\n"+shim+" <worktree>\n'''\n"))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a skill carrying the shim alone is a rule 2 violation")
			for _, sib := range siblings {
				gsReports(t, r, sib+" is a sibling dependency", "names the missing "+sib)
			}
		})
	}
}

// TestEveryFlowGuardShimRequiresLib is KAN-854: every real shim that loads
// lib/flow-guard.sh -- by $SCRIPT_DIR or by $(dirname -- "${BASH_SOURCE[0]}")
// -- makes lib a rule 2 sibling dependency of a skill that carries it. Before
// gsSelfDirSibling, every dirname-spelled shim passed with no lib symlink.
func TestEveryFlowGuardShimRequiresLib(t *testing.T) {
	t.Parallel()
	scripts := tcfScriptsDir(t)
	entries, err := os.ReadDir(scripts)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, e := range entries {
		shim := e.Name()
		if !strings.HasSuffix(shim, ".sh") || strings.HasPrefix(shim, "test-") {
			continue
		}
		b, err := os.ReadFile(scripts + "/" + shim)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `/lib/flow-guard.sh"`) {
			continue
		}
		found++
		t.Run(shim, func(t *testing.T) {
			t.Parallel()
			repo := gsNew(t)
			gsGuard(t, repo, shim, string(b))
			gsLink(t, repo, "flow", shim)
			gsSkill(t, repo, "flow", gsMD("# flow fixture\n\n'''bash\n"+shim+" <worktree>\n'''\n"))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a skill carrying the shim alone is a rule 2 violation")
			gsReports(t, r, "lib is a sibling dependency", "names the missing lib")
		})
	}
	if found == 0 {
		t.Fatal("no shim in scripts/ loads lib/flow-guard.sh -- the walk found nothing to check")
	}
}

// TestSelfDirSiblingSpellings pins gsSelfDirSibling's two spellings and its
// refusal to read a `/..` walk as a sibling.
func TestSelfDirSiblingSpellings(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, body string }{
		{"bash-source", `. "$(dirname -- "${BASH_SOURCE[0]}")/lib/helper.sh"`},
		{"dollar-zero", `. "$(dirname "$0")/lib/helper.sh"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			repo := gsNew(t)
			writeFile(t, repo+"/scripts/lib/helper.sh", "true\n")
			gsGuard(t, repo, "check-x.sh", "#!/usr/bin/env bash\n"+c.body)
			gsLink(t, repo, "flow", "check-x.sh")
			gsSkill(t, repo, "flow", gsMD("# flow fixture\n\n'''bash\ncheck-x.sh <worktree>\n'''\n"))
			r := gsRun(t, repo)
			gsInvalid(t, r, "a skill carrying the guard without lib is a rule 2 violation")
			gsReports(t, r, "lib is a sibling dependency", "names the missing lib")
		})
	}
	t.Run("parent-walk", func(t *testing.T) {
		t.Parallel()
		repo := gsNew(t)
		gsGuard(t, repo, "check-x.sh", "#!/usr/bin/env bash\n"+`ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"`)
		gsLink(t, repo, "flow", "check-x.sh")
		gsSkill(t, repo, "flow", gsMD("# flow fixture\n\n'''bash\ncheck-x.sh <worktree>\n'''\n"))
		r := gsRun(t, repo)
		gsSilent(t, r, "a /.. walk names no sibling")
		gsCheck(t, "a /.. walk is no sibling", !strings.Contains(r.out, ".. is a sibling dependency"), "output: %s", r.out)
	})
}
