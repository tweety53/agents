package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Every case of the retired scripts/test-check-plan-provenance.sh, one leaf
// subtest per ok: label (the bash harness's own text, a timing label's
// measured figure dropped), grouped under the harness's case number and in
// the harness's own order, so its section comments still locate a case.
// Each case builds its own fixture tree in t.TempDir() -- the harness's
// new_fixture -- and runs the guard in-process with
// CHECK_PLAN_PROVENANCE_ROOT pointed at it; stdout and stderr share one
// buffer, as the harness's 2>&1 read them.
//
// READ THIS BEFORE ADDING OR "FIXING" A FIXTURE (carried from the harness).
// That suite four times encoded one of the guard's own defects as its
// specification -- asserting the buggy output as correct, which then made
// the defect look verified and hid it through several further review
// passes. Check a fixture against the CommonMark spec, with the section
// cited in a comment, before believing it has found a bug; observed output
// is never the justification for an assertion.

const (
	ppDemo    = "spectre/changes/demo-change/"
	ppClean   = "```bash verified:ran it locally\necho hi\n```\n" // clean_tasks_md
	ppStated1 = "check-plan-provenance: 1 file(s) scanned, all provenance stated"
	ppNoTag   = "fenced code block has no verified:/unverified: tag"
	ppNoProv  = "numeric claim with no measured:/predicted: provenance comment"
)

// ppFixture is the harness's new_fixture: one live change, one archived
// change and an out-of-scope skills/ directory.
func ppFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{ppDemo, "spectre/changes/archive/old-change", "skills/foo"} {
		mkdir(t, filepath.Join(root, d))
	}
	return root
}

// ppWrite writes body to root/rel, creating its directory.
func ppWrite(t *testing.T, root, rel, body string) {
	t.Helper()
	writeFile(t, filepath.Join(root, rel), body)
}

// ppRun is the harness's run_guard: the guard in-process against root.
func ppRun(root string) tcfRes {
	return ppRunEnv(map[string]string{"CHECK_PLAN_PROVENANCE_ROOT": root})
}

// ppRunEnv runs the guard with vars as its whole environment.
func ppRunEnv(vars map[string]string) tcfRes {
	var out bytes.Buffer
	rc := Registry["check-plan-provenance"](nil, crEnv(os.TempDir(), vars), &out, &out)
	return tcfRes{rc, out.String()}
}

// ppTasks is new_fixture, a demo-change/tasks.md carrying body, run_guard.
func ppTasks(t *testing.T, body string) tcfRes {
	t.Helper()
	root := ppFixture(t)
	ppWrite(t, root, ppDemo+"tasks.md", body)
	return ppRun(root)
}

// ppOut is $OUT as the harness's $(...) captured it: trailing newlines
// stripped.
func ppOut(r tcfRes) string { return strings.TrimRight(r.out, "\n") }

// ppChmod is chmod(1) for a fixture the case locks and later restores.
func ppChmod(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func ppSymlink(t *testing.T, target, link string) {
	t.Helper()
	mkdir(t, filepath.Dir(link))
	_ = os.Remove(link)
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// ppShim runs scripts/check-plan-provenance.sh itself through real bash
// with env as its whole environment.
func ppShim(t *testing.T, env ...string) tcfRes {
	t.Helper()
	return ppShimAt(t, tcfScriptsDir(t)+"/check-plan-provenance.sh", "", env...)
}

// ppShimAt runs the shim at script from working directory dir ("" is the
// test's own).
func ppShimAt(t *testing.T, script, dir string, env ...string) tcfRes {
	t.Helper()
	cmd := exec.Command("/bin/bash", script)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if cmd.ProcessState == nil {
		t.Fatal(err)
	}
	return tcfRes{cmd.ProcessState.ExitCode(), string(out)}
}

// ppTimed runs the guard against root and reports how long it took.
func ppTimed(root string) (tcfRes, time.Duration) {
	start := time.Now()
	r := ppRun(root)
	return r, time.Since(start)
}

type ppCase struct {
	name string
	fn   func(t *testing.T)
}

func TestCheckPlanProvenance(t *testing.T) {
	t.Parallel()
	if Registry["check-plan-provenance"] == nil {
		t.Fatal("check-plan-provenance is not registered")
	}
	var cases []ppCase
	for _, part := range [][]ppCase{
		ppCasesSyntaxToPass6(), ppCasesPass6To8(), ppCasesPass8To10(), ppCasesPass10To11(),
		ppCasesPass11Containment(), ppCasesLadderBanners(), ppCasesPass13(), ppCasesPass14(),
		ppCasesAnchorQuotes(), ppCasesScopeToEnd(), ppCasesShim(), ppCasesUnicodeMatchers(),
	} {
		cases = append(cases, part...)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}

// ppCasesSyntaxToPass6 is the harness from its first section to the
// pass-6 fix wave's invalid-UTF-8 case.
func ppCasesSyntaxToPass6() []ppCase {
	return []ppCase{
		// SECTION: Provenance syntax — verified:/unverified: tag vocabulary.
		{"case 1", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:ran it locally\necho hi\n```\n")
			ok(t, "verified: block passes", r.rc == 0, r)
		}},
		{"case 2", func(t *testing.T) {
			r := ppTasks(t, "```bash unverified:confirm the flag name\necho hi\n```\n")
			ok(t, "unverified: block passes", r.rc == 0, r)
		}},
		// SECTION: Fence detection — a fenced block with no tag is flagged.
		{"case 3", func(t *testing.T) {
			r := ppTasks(t, "intro\n```bash\necho hi\n```\n")
			ok(t, "untagged block fails", r.rc != 0, r)
			ok(t, "untagged block reports file:line", has(r.out, "spectre/changes/demo-change/tasks.md:2"), r)
		}},
		// SECTION: Claim boundaries — numeric provenance (measured:/predicted:).
		{"case 4", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- measured: ./gradlew test @ c515c42 -->\n")
			ok(t, "measured: comment satisfies numeric claim", r.rc == 0, r)
		}},
		{"case 5", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n")
			ok(t, "unattributed numeric claim fails", r.rc != 0, r)
			ok(t, "unattributed claim reports file:line", has(r.out, "spectre/changes/demo-change/tasks.md:1"), r)
		}},
		{"case 6", func(t *testing.T) {
			r := ppTasks(t, "After the deletion: 186 tests\n<!-- predicted: ./gradlew test after task 1 -->\n")
			ok(t, "predicted: comment satisfies numeric claim", r.rc == 0, r)
		}},
		// SECTION: Scope — which files/directories the guard scans. Cases 7-9
		// carry the clean tasks.md, so each exercises ONLY the out-of-scope
		// behaviour under test — not the "zero tasks.md scanned" refusal
		// (case 13), which would otherwise mask it with a false rc=2.
		{"case 7", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			ppWrite(t, root, "spectre/changes/archive/old-change/tasks.md", "```bash\necho hi\n```\n")
			r := ppRun(root)
			ok(t, "archived tasks.md is not scanned", r.rc == 0, r)
		}},
		{"case 7b", func(t *testing.T) {
			root := vmRepo(t, map[string]string{
				ppDemo + "tasks.md":                           ppClean,
				"spectre/changes/archive/old-change/tasks.md": "```bash\necho hi\n```\n",
			})
			r := ppRun(root)
			ok(t, "an archived change the base carries is not scanned", r.rc == 0, r)
			ppWrite(t, root, "spectre/changes/archive/new-change/tasks.md", "```bash\necho hi\n```\n")
			r = ppRun(root)
			ok(t, "an archived change the base lacks is scanned", r.rc == 1 && has(r.out, "spectre/changes/archive/new-change/tasks.md:1") && !has(r.out, "old-change"), r)
		}},
		{"case 8", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			ppWrite(t, root, ppDemo+"proposal.md", "```bash\necho hi\n```\n")
			r := ppRun(root)
			ok(t, "proposal.md is in scope", r.rc == 1, r)
		}},
		{"case 9", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			ppWrite(t, root, "skills/foo/SKILL.md", "```bash\necho hi\n```\n")
			r := ppRun(root)
			ok(t, "SKILL.md is out of scope", r.rc == 0, r)
		}},
		// SECTION: Fence detection (cont'd) — nested/quoted fence boundaries.
		// Case 10: the inner 3-backtick fence is not a new block boundary and
		// the outer 4-backtick close is recognised past it — CommonMark's
		// same-character/at-least-as-long closing rule.
		{"case 10", func(t *testing.T) {
			r := ppTasks(t, "````markdown verified:authored in-tree for this change\n```kotlin verified:javap intellij.platform.diff.jar\noverride val toolWindowIds: Array<String>\n```\n````\n")
			ok(t, "4-backtick block containing a 3-backtick block passes", r.rc == 0, r)
		}},
		// SECTION: Claim boundaries (cont'd) — over-firing guards and the
		// lookahead window.
		{"case 11", func(t *testing.T) {
			r := ppTasks(t, "See step 1.2 for details.\nError reported at line 452.\nTracked as IU-262 and KAN-14.\nUpgrade to version 2.1.0 of the library.\n")
			ok(t, "step/line/ticket/version numbers do not over-fire", r.rc == 0, r)
		}},
		// 11b: a unit word must not match as a mere PREFIX of a longer word.
		{"case 11b-1", func(t *testing.T) {
			r := ppTasks(t, "We migrated 10 filesystems last week.\n")
			ok(t, "'10 filesystems' does not match 'files' as a prefix", r.rc == 0, r)
		}},
		{"case 11b-2", func(t *testing.T) {
			r := ppTasks(t, "It took 5 minuteswalk to get there.\n")
			ok(t, "'5 minuteswalk' does not match 'minutes' as a prefix", r.rc == 0, r)
		}},
		{"case 11b-3", func(t *testing.T) {
			r := ppTasks(t, "Found 10 errorsome behaviors.\n")
			ok(t, "'10 errorsome' does not match 'errors' as a prefix", r.rc == 0, r)
		}},
		{"case 11b-4", func(t *testing.T) {
			r := ppTasks(t, "The batch processed 20 testsuite runs overnight.\n")
			ok(t, "'20 testsuite' does not match 'tests' as a prefix", r.rc == 0, r)
		}},
		// 11c: the lookahead window is the claim's line and the two after it.
		{"case 11c-1", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\na blank note line\n<!-- measured: ./gradlew test @ c515c42 -->\n")
			ok(t, "a provenance comment 2 lines after the claim satisfies it", r.rc == 0, r)
		}},
		{"case 11c-2", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\na blank note line\nanother blank note line\n<!-- measured: ./gradlew test @ c515c42 -->\n")
			ok(t, "a provenance comment 3 lines after the claim does not satisfy it", r.rc != 0, r)
		}},
		// SECTION: Exit codes — environment (exit 2).
		{"case 12", func(t *testing.T) {
			r := ppRun(filepath.Join(t.TempDir(), "plan-prov-does-not-exist"))
			ok(t, "a nonexistent CHECK_PLAN_PROVENANCE_ROOT exits 2", r.rc == 2, r)
		}},
		// 13: a changes tree with zero matching tasks.md (one at the wrong
		// depth) never reports a clean run.
		{"case 13", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"nested/tasks.md", "```bash\necho hi\n```\n")
			r := ppRun(root)
			ok(t, "zero matching tasks.md refuses to report a clean run", r.rc == 2, r)
		}},
		// SECTION: Fence detection (cont'd) — indentation.
		{"case 14", func(t *testing.T) {
			r := ppTasks(t, "1. Do the thing:\n  ```bash\n  echo hi\n  ```\n")
			ok(t, "an indented untagged fence is detected", r.rc != 0, r)
			ok(t, "indented fence reports the correct line", has(r.out, "tasks.md:2"), r)
		}},
		{"case 14b", func(t *testing.T) {
			r := ppTasks(t, "1. Do the thing:\n  ```bash verified:ran it locally\n  echo hi\n  ```\n")
			ok(t, "an indented tagged fence passes", r.rc == 0, r)
		}},
		// SECTION: Claim boundaries (cont'd) — provenance-comment negation and
		// CLAIM_RE's left boundary.
		{"case 15-1", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- not actually measured: just a guess -->\n")
			ok(t, "'not actually measured:' does not satisfy provenance", r.rc != 0, r)
		}},
		{"case 15-2", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- unmeasured: revisit -->\n")
			ok(t, "'unmeasured:' does not satisfy provenance", r.rc != 0, r)
		}},
		{"case 16-1", func(t *testing.T) {
			r := ppTasks(t, "Run sha256 tests to confirm the binary.\n")
			ok(t, "'sha256 tests' does not over-fire", r.rc == 0, r)
		}},
		{"case 16-2", func(t *testing.T) {
			r := ppTasks(t, "Normalize utf8 lines before comparing.\n")
			ok(t, "'utf8 lines' does not over-fire", r.rc == 0, r)
		}},
		{"case 16-3", func(t *testing.T) {
			r := ppTasks(t, "Track ticket KAN256 for the follow-up.\n")
			ok(t, "'KAN256' (no unit word) does not over-fire", r.rc == 0, r)
		}},
		{"case 16-4", func(t *testing.T) {
			r := ppTasks(t, "It found 256 failures overnight.\n")
			ok(t, "a genuine boundary-satisfying numeric claim still fires", r.rc != 0, r)
		}},
		// SECTION: Scope (cont'd) — "nothing in flight" steady state.
		{"case 17", func(t *testing.T) {
			root := ppFixture(t)
			os.RemoveAll(filepath.Join(root, ppDemo))
			r := ppRun(root)
			ok(t, "zero non-archived changes exits 0 (nothing in flight)", r.rc == 0, r)
		}},
		{"case 17b", func(t *testing.T) {
			root := ppFixture(t)
			os.RemoveAll(filepath.Join(root, ppDemo))
			ppWrite(t, root, "spectre/changes/archive/old-change/tasks.md", "```bash\necho hi\n```\n")
			r := ppRun(root)
			ok(t, "only archive/ present still exits 0", r.rc == 0, r)
		}},
		// SECTION: Exit codes (cont'd) — environment (exit 2).
		{"case 18", func(t *testing.T) {
			root := ppFixture(t)
			os.RemoveAll(filepath.Join(root, "spectre"))
			r := ppRun(root)
			ok(t, "missing spectre/changes/ exits 2 (cannot determine)", r.rc == 2, r)
		}},
		// SECTION: Containment / security — symlink escapes (exit 3).
		{"case 19", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, "real-changes/evil-change/tasks.md", "```bash\necho hi\n```\n")
			os.RemoveAll(filepath.Join(root, ppDemo))
			ppSymlink(t, filepath.Join(root, "real-changes/evil-change"), filepath.Join(root, "spectre/changes/evil-change"))
			r := ppRun(root)
			ok(t, "a symlinked change directory is not bypassed", r.rc != 0, r)
		}},
		{"case 20", func(t *testing.T) {
			root := ppFixture(t)
			ppSymlink(t, "/dev/zero", filepath.Join(root, ppDemo+"tasks.md"))
			r := ppRun(root)
			ok(t, "a symlinked tasks.md is refused, not opened (exit 3, containment)", r.rc == 3, r)
		}},
		// SECTION: Provenance syntax (cont'd) — fence-tag left boundary.
		{"case 21", func(t *testing.T) {
			r := ppTasks(t, "```bash preverified:not a real tag\necho hi\n```\n")
			ok(t, "'preverified:' does not satisfy the fence-tag rule", r.rc != 0, r)
		}},
		// SECTION: Fence detection (cont'd) — container-prefix parsing.
		{"case 22", func(t *testing.T) {
			r := ppTasks(t, "> ```bash\n> rm -rf /\n> ```\n")
			ok(t, "a blockquoted untagged fence is detected", r.rc != 0, r)
			ok(t, "blockquoted fence reports the correct line", has(r.out, "tasks.md:1"), r)
		}},
		{"case 22b", func(t *testing.T) {
			r := ppTasks(t, "> ```bash verified:ran it locally\n> echo hi\n> ```\n")
			ok(t, "a blockquoted tagged fence passes", r.rc == 0, r)
		}},
		{"case 23", func(t *testing.T) {
			r := ppTasks(t, "- ```bash\n  echo hi\n  ```\nBaseline: 42 tests\n")
			ok(t, "a bulleted untagged fence is detected", r.rc != 0, r)
			ok(t, "bulleted fence reports the opening line, not the close", has(r.out, "tasks.md:1: "+ppNoTag), r)
			ok(t, "a claim after the bulleted fence's close is still reported, not swallowed", has(r.out, "tasks.md:4: "+ppNoProv), r)
		}},
		// 24: at column 0 indentation is never the cause (Important 4,
		// pass-10), so the remedy never says dedent.
		{"case 24", func(t *testing.T) {
			r := ppTasks(t, "- - ```bash\n  echo hi\n  ```\n")
			ok(t, "a doubled list marker on a fence line aborts loudly (exit 4, content-classification)", r.rc == 4, r)
			ok(t, "doubled marker at column 0: remedy names the doubled marker, never dedent",
				!has(r.out, "dedent") && has(r.out, "doubled", "marker"), r)
		}},
		{"case 25", func(t *testing.T) {
			r := ppTasks(t, "1. 2. ```bash\n   echo hi\n   ```\n")
			ok(t, "a doubled ordered-list marker on a fence line aborts loudly (exit 4, content-classification)", r.rc == 4, r)
		}},
		// 25b: `--` is not a marker at all; the later bare fence is unclosed.
		{"case 25b", func(t *testing.T) {
			r := ppTasks(t, "-- ```bash\necho hi\n```\n")
			ok(t, "glued '--' with no space is prose, not a doubled marker; the later bare fence is reported unclosed (rc=1)", r.rc == 1, r)
		}},
		// 26: the checkbox-fence regression; continuation at the true
		// content column `- [ ] ` establishes (6).
		{"case 26", func(t *testing.T) {
			r := ppTasks(t, "- [ ] ```bash verified:ran it locally\n      echo hi\n      ```\nBaseline: 500 tests\n```bash\necho second block, untagged\n```\n")
			ok(t, "checkbox-fence regression: guard still fails the file", r.rc != 0, r)
			ok(t, "checkbox-fence regression: tagged fence at line 1 is not a false positive", !has(r.out, "tasks.md:1"), r)
			ok(t, "checkbox-fence regression: the untagged claim on line 4 is reported, not swallowed", has(r.out, "tasks.md:4: "+ppNoProv), r)
			ok(t, "checkbox-fence regression: the untagged block opened at line 5 is reported, not swallowed", has(r.out, "tasks.md:5: "+ppNoTag), r)
		}},
		{"case 27", func(t *testing.T) {
			r := ppTasks(t, "- [ ] ```bash\n      echo hi\n      ```\n")
			ok(t, "an untagged checkbox-fence is detected", r.rc != 0, r)
			ok(t, "untagged checkbox-fence reports the opening line", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		{"case 28", func(t *testing.T) {
			r := ppTasks(t, "- [x] ```bash unverified:confirm the flag name\n      echo hi\n      ```\n")
			ok(t, "a tagged checked-checkbox-fence passes", r.rc == 0, r)
		}},
		{"case 29", func(t *testing.T) {
			r := ppTasks(t, "> - [ ] ```bash verified:ran it locally\n>       echo hi\n>       ```\n")
			ok(t, "a blockquoted checkbox-fence, tagged, passes", r.rc == 0, r)
		}},
		{"case 30", func(t *testing.T) {
			r := ppTasks(t, "- > ```bash verified:ran it locally\n  > echo hi\n  > ```\n")
			ok(t, "a list-item-containing-a-blockquote fence, tagged, passes", r.rc == 0, r)
		}},
		{"case 31", func(t *testing.T) {
			r := ppTasks(t, "[ ] ```bash\necho hi\n```\n")
			ok(t, "a bare checkbox with no list marker aborts loudly (exit 4, content-classification)", r.rc == 4, r)
		}},
		{"case 32", func(t *testing.T) {
			r := ppTasks(t, "= ```bash``` is an assignment-like example, not a real fence.\n")
			ok(t, "'= ```bash```' is read as prose, not aborted", r.rc == 0, r)
		}},
		{"case 33", func(t *testing.T) {
			r := ppTasks(t, "Note that ```bash``` marks a fence inline, not as a real block.\n")
			ok(t, "a mid-sentence mention of a fence run does not abort", r.rc == 0, r)
		}},
		{"case 34", func(t *testing.T) {
			r := ppTasks(t, "3.4. ```bash verified:ran it locally\n     echo hi\n     ```\n")
			ok(t, "a compound-numbered (3.4.) tagged fence passes", r.rc == 0, r)
		}},
		{"case 34b", func(t *testing.T) {
			r := ppTasks(t, "3.4. ```bash\n     echo hi\n     ```\n")
			ok(t, "a compound-numbered (3.4.) untagged fence is detected", r.rc != 0, r)
			ok(t, "compound-numbered untagged fence reports the opening line", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		{"case 34c", func(t *testing.T) {
			r := ppTasks(t, "1.2.3) ```bash verified:ran it locally\n       echo hi\n       ```\n")
			ok(t, "a compound-numbered (1.2.3)) tagged fence passes", r.rc == 0, r)
		}},
		{"case 34d", func(t *testing.T) {
			r := ppTasks(t, "3.4 explains how the ```bash``` fence marker works inline, not as a real block.\n")
			ok(t, "an unterminated compound number mentioning a fence mid-sentence does not abort", r.rc == 0, r)
		}},
		{"case 34e-1", func(t *testing.T) {
			r := ppTasks(t, "= ```bash\necho hi\n```bash verified:closes a fresh opener, not the = line\nmore\n```\n")
			ok(t, "'= ```bash' is prose; the real fence below it is still parsed normally", r.rc == 0, r)
		}},
		{"case 34e-2", func(t *testing.T) {
			r := ppTasks(t, "[ ] ```bash\necho hi\n```\n")
			ok(t, "a bare checkbox with no list marker still aborts loudly (exit 4, content-classification)", r.rc == 4, r)
		}},
		// 35: prose that mentions fence syntax behind punctuation.
		{"case 35-1", func(t *testing.T) {
			r := ppTasks(t, "`git diff` shows changes; wrap snippets in ```bash fences with a tag.\n")
			ok(t, "a line led by a single-backtick code span does not abort", r.rc == 0, r)
		}},
		{"case 35-2", func(t *testing.T) {
			r := ppTasks(t, "\"```\" is the fence marker used here, see docs.\n")
			ok(t, "a line led by a double-quote does not abort", r.rc == 0, r)
		}},
		{"case 35-3", func(t *testing.T) {
			r := ppTasks(t, "(wrap snippets in ```lang blocks with a tag)\n")
			ok(t, "a line led by a parenthesis does not abort", r.rc == 0, r)
		}},
		{"case 35-4", func(t *testing.T) {
			r := ppTasks(t, "**Note:** wrap the snippet in ```bash blocks with a tag.\n")
			ok(t, "a line led by bold markdown (**) does not abort", r.rc == 0, r)
		}},
		{"case 36", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:test\necho \"start\"\n# See ```example``` for illustration\necho \"end\"\n```\n")
			ok(t, "a fence run behind punctuation as CONTENT inside an open fence does not abort", r.rc == 0, r)
		}},
		// SECTION: Containment / security (cont'd) — quadratic-DoS regression
		// guard. Timed directly, so a regression back to quadratic behaviour
		// fails an assertion rather than merely running the suite slower.
		{"case 37", func(t *testing.T) {
			root := ppFixture(t)
			many := strings.Repeat("> ", 4000)
			ppWrite(t, root, ppDemo+"tasks.md", many+"```bash\necho hi\n"+many+"```\n")
			r, took := ppTimed(root)
			ok(t, "a 4000-marker blockquote-padded untagged fence is still detected", r.rc != 0, r)
			ok(t, "4000 repeated blockquote markers scan in under 5s", took < 5*time.Second, tcfRes{r.rc, fmt.Sprintf("took %v", took)})
		}},
		// SECTION: Pass-5 Criticals — container scoping on the CLOSING test.
		{"case 38", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:test\necho \"start\"\n> ```\necho \"ran 200 tests\"\n```\n")
			ok(t, "a blockquote-shaped fence run as CONTENT inside a bare-opened fence does not falsely close it", r.rc == 0, r)
		}},
		{"case 39-1", func(t *testing.T) {
			r := ppTasks(t, "- Confirmed the fence style uses ```: markers, not tildes.\n")
			ok(t, "bullet-prefixed prose mentioning a fence run does not abort", r.rc == 0, r)
		}},
		{"case 39-2", func(t *testing.T) {
			r := ppTasks(t, "> See the ``` marker used below for illustration.\n")
			ok(t, "blockquote-prefixed prose mentioning a fence run does not abort", r.rc == 0, r)
		}},
		{"case 39b", func(t *testing.T) {
			r := ppTasks(t, "- 42 tests pass in the suite, see ```bash for reference\n")
			ok(t, "bullet-prefixed prose with an accompanying numeric claim is not swallowed by an abort", r.rc != 0, r)
			ok(t, "the '42 tests' claim is still reported", has(r.out, "tasks.md:1: "+ppNoProv), r)
		}},
		// 40: the bash guard's O(N^2) array walk (10k lines 0.68s, 160k lines
		// 47.56s, measured against it) must stay linear.
		{"case 40", func(t *testing.T) {
			root := ppFixture(t)
			var b strings.Builder
			for i := 1; i <= 40000; i++ {
				fmt.Fprintf(&b, "Task %d: ordinary prose line with no fence and no numeric claim.\n", i)
			}
			b.WriteString("```bash verified:generated fixture\necho hi\n```\n")
			ppWrite(t, root, ppDemo+"tasks.md", b.String())
			r, took := ppTimed(root)
			ok(t, "a 40,000-line clean tasks.md scans clean", r.rc == 0, r)
			ok(t, "a 40,000-line tasks.md scans in under 5s", took < 5*time.Second, tcfRes{r.rc, fmt.Sprintf("took %v", took)})
		}},
		// SECTION: Pass-6 fix wave — Critical 1: blockquote depth is a COUNT.
		{"case 41", func(t *testing.T) {
			r := ppTasks(t, "> ```bash verified:test\n> echo \"start\"\n```\n> echo \"ran 111 tests\"\n> ```\n")
			ok(t, "depth-1 fence: a bare (zero-level) line does not close it; the matching one-level line does", r.rc == 0, r)
		}},
		{"case 42", func(t *testing.T) {
			r := ppTasks(t, "> > ```bash verified:test\n> > echo \"start\"\n> ```\n> > echo \"ran 999 tests\"\n> > ```\n")
			ok(t, "depth-2 fence: a one-level-deep line does not falsely close it (the pass-6 brief repro)", r.rc == 0, r)
			ok(t, "depth-2 fence: zero violations reported", ppOut(r) == ppStated1, r)
		}},
		{"case 43", func(t *testing.T) {
			r := ppTasks(t, "> > > ```bash verified:test\n> > > echo \"start\"\n> ```\n> > ```\n> > > echo \"ran 777 tests\"\n> > > ```\n")
			ok(t, "depth-3 fence: neither a one-level nor a two-level line falsely closes it; the matching three-level line does", r.rc == 0, r)
		}},
		{"case 44", func(t *testing.T) {
			r := ppTasks(t, "> > ```bash\n> > echo hi\n> > ```\n")
			ok(t, "depth-2 fence: a genuinely untagged block, correctly closed at depth 2, is still flagged", r.rc != 0, r)
			ok(t, "depth-2 untagged fence reports the opening line", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		// SECTION: Pass-6 fix wave — Critical 2: a leading UTF-8 BOM.
		{"case 45", func(t *testing.T) {
			r := ppTasks(t, "\xef\xbb\xbf```bash verified:test\necho hi\n```\nBaseline: 999 tests\n")
			ok(t, "a BOM-prefixed fence opens/closes correctly; the trailing unattributed claim is reported", r.rc != 0, r)
			ok(t, "BOM fixture: the '999 tests' claim is reported, not silently swallowed", has(r.out, "tasks.md:4: "+ppNoProv), r)
		}},
		// SECTION: Pass-6 fix wave — Critical 3: quadratic ReDoS in CLAIM_RE
		// (measured through the real wrapper before the fix: 4,000 -> 0.22s,
		// 8,000 -> 0.66s, 16,000 -> 2.39s).
		{"case 46", func(t *testing.T) {
			root := ppFixture(t)
			nums := make([]string, 0, 16000)
			for i := 1; i < 16000; i++ {
				nums = append(nums, fmt.Sprint(i))
			}
			ppWrite(t, root, ppDemo+"tasks.md", strings.Join(nums, ",")+" widgets\n")
			r, took := ppTimed(root)
			ok(t, "a large comma-separated-digits line with no unit word is not a claim (exit 0)", r.rc == 0, r)
			ok(t, "a 16,000-number comma-separated line scans well under a second", took < 2*time.Second, tcfRes{r.rc, fmt.Sprintf("took %v", took)})
		}},
		// SECTION: Pass-6 fix wave — Critical 4: an unreadable file after
		// containment is classified, never exit 1.
		{"case 47", func(t *testing.T) {
			root := ppFixture(t)
			p := filepath.Join(root, ppDemo+"tasks.md")
			ppWrite(t, root, ppDemo+"tasks.md", "Baseline: 197 tests\n<!-- measured: ./gradlew test @ c515c42 -->\n")
			ppChmod(t, p, 0)
			r := ppRun(root)
			ppChmod(t, p, 0o644)
			ok(t, "chmod 000 on tasks.md exits classified, never rc=1 (violations-found)", r.rc != 1, r)
			ok(t, "chmod 000: a classified message is printed, not a traceback", !has(r.out, "Traceback") && has(r.out, "cannot read file"), r)
		}},
		// Case 48 pinned the Python wrapper's missing-python3 exit. The Go
		// port has no interpreter to miss; the one missing piece a shipped
		// guard can still meet is the toolchain the shim builds flow-guard
		// with (KAN-760's guard-binary-built-from-checkout): exit 2, never
		// bash's own 127, run through real bash with an empty build cache.
		{"case 48", func(t *testing.T) {
			r := ppShim(t, "PATH=/usr/bin:/bin", "FLOW_GUARD_CACHE_DIR="+t.TempDir(), "CHECK_PLAN_PROVENANCE_ROOT="+ppFixture(t))
			ok(t, "missing go on PATH exits 2", r.rc == 2, r)
			ok(t, "missing go: message names go", has(r.out, "check-plan-provenance: cannot build flow-guard from", "no go on PATH"), r)
		}},
		// SECTION: Pass-6 fix wave — Important 6: invalid UTF-8 is refused.
		{"case 49", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- measured: ./gradlew test @ c515c42 -->\n\xff\xfe not valid utf-8\n")
			ok(t, "invalid UTF-8 exits 2 (same taxonomy code as the chmod-000 OSError case)", r.rc == 2, r)
			ok(t, "invalid UTF-8: a classified message is printed", has(r.out, "cannot decode as UTF-8"), r)
		}},
	}
}

// ppCasesPass6To8 runs from the pass-6 info-string case to the pass-8
// fix wave's Critical 1.
func ppCasesPass6To8() []ppCase {
	return []ppCase{
		// SECTION: Pass-6 fix wave — Minor: a numeric claim on a fence's
		// info-string line.
		{"case 50", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:ran 500 tests\necho hi\n```\n")
			ok(t, "a numeric claim in a fence's info string is reported", r.rc != 0, r)
			ok(t, "info-string claim reports the fence's opening line", has(r.out, "tasks.md:1: "+ppNoProv), r)
		}},
		{"case 50b", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:ran 500 tests\n<!-- measured: ./gradlew test @ c515c42 -->\necho hi\n```\n")
			ok(t, "an info-string claim with a measured: comment in the lookahead window passes", r.rc == 0, r)
		}},
		// SECTION: Pass-7 fix wave — Critical 1: an unreadable/broken scan
		// entry is a scan-integrity failure, never a silent skip.
		{"case 51", func(t *testing.T) {
			root := ppFixture(t)
			locked := filepath.Join(root, "spectre/changes/locked-change")
			ppWrite(t, root, "spectre/changes/locked-change/tasks.md", "Baseline: 123 tests\n")
			ppChmod(t, locked, 0)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			r := ppRun(root)
			ppChmod(t, locked, 0o755)
			ok(t, "a locked (chmod 000) change dir beside a clean one aborts classified, never a clean run", r.rc == 2, r)
			ok(t, "locked dir: no false 'all provenance stated' clean run", !has(r.out, "all provenance stated"), r)
		}},
		{"case 52", func(t *testing.T) {
			root := ppFixture(t)
			os.RemoveAll(filepath.Join(root, ppDemo))
			ppSymlink(t, filepath.Join(root, "spectre/changes/does-not-exist"), filepath.Join(root, "spectre/changes/dangling-change"))
			r := ppRun(root)
			ok(t, "a dangling symlink change dir aborts classified, never 'nothing in flight'", r.rc == 2, r)
			ok(t, "dangling symlink (alone): no false 'nothing in flight'", !has(r.out, "nothing in flight"), r)
		}},
		{"case 53", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			ppSymlink(t, filepath.Join(root, "spectre/changes/does-not-exist"), filepath.Join(root, "spectre/changes/dangling-change"))
			r := ppRun(root)
			ok(t, "a dangling symlink beside a real change aborts classified, never a clean run", r.rc == 2, r)
			ok(t, "dangling symlink (beside real): no false 'all provenance stated' clean run", !has(r.out, "all provenance stated"), r)
		}},
		{"case 54", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			ppSymlink(t, filepath.Join(root, "spectre/changes/circular-a"), filepath.Join(root, "spectre/changes/circular-b"))
			ppSymlink(t, filepath.Join(root, "spectre/changes/circular-b"), filepath.Join(root, "spectre/changes/circular-a"))
			r := ppRun(root)
			ok(t, "a circular symlink change dir aborts classified, never silently skipped", r.rc == 2, r)
			ok(t, "circular symlink: no false 'all provenance stated' clean run", !has(r.out, "all provenance stated"), r)
		}},
		// SECTION: Pass-7 fix wave — Critical 2: an UN-CONTAINED fence's own
		// indentation is capped at 3 columns (CommonMark).
		{"case 55", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:ci\necho hi\n```\n")
			ok(t, "0-column fence still opens/closes normally", r.rc == 0, r)
		}},
		{"case 56", func(t *testing.T) {
			r := ppTasks(t, "   ```bash verified:ci\n   echo hi\n   ```\n")
			ok(t, "3-column fence is still a real fence", r.rc == 0, r)
		}},
		// 57: re-pointed by pass 8's Critical 1 — a bare 4+-column fence-like
		// line aborts rather than being read either way; the claim later in
		// the file is no longer reached, which is the point.
		{"case 57", func(t *testing.T) {
			r := ppTasks(t, "    ``` verified:ci\n    echo hi\n    Suite finished: ran 500 tests.\n    ```\n")
			ok(t, "a 4-column bare fence-like run aborts (content-classification), not read as an indented code block", r.rc == 4, r)
			ok(t, "4-column pseudo-fence: the abort names the opening line", has(r.out, "tasks.md:1: fence-like run indented 4+ columns"), r)
		}},
		{"case 58", func(t *testing.T) {
			r := ppTasks(t, "        ~~~ verified:ci\n        echo hi\n        Suite finished: ran 777 tests.\n        ~~~\n")
			ok(t, "an 8-column bare fence-like run (tilde) aborts identically to the 4-column backtick case", r.rc == 4, r)
		}},
		{"case 59", func(t *testing.T) {
			r := ppTasks(t, "    ```\n    Baseline: ran 321 tests.\n    ```\n")
			ok(t, "a 4-column untagged bare fence-like run aborts rather than being misread either as a fence or as prose", r.rc == 4, r)
			ok(t, "4-column untagged pseudo-fence: not misreported as a fence violation", !has(r.out, ppNoTag), r)
		}},
		// SECTION: Pass-7 fix wave — Critical 3: CLAIM_RE's boundary is
		// negative (not word-constituent), not an ASCII-punctuation allowlist.
		{"case 60", func(t *testing.T) {
			r := ppTasks(t, "Result—42 tests\n")
			ok(t, "em dash (left, glued) is caught", r.rc != 0, r)
		}},
		{"case 61", func(t *testing.T) {
			r := ppTasks(t, "Result‘42 tests\n")
			ok(t, "curly quote (left, glued) is caught", r.rc != 0, r)
		}},
		{"case 62", func(t *testing.T) {
			r := ppTasks(t, "（42 tests\n")
			ok(t, "fullwidth paren (left, glued) is caught", r.rc != 0, r)
		}},
		{"case 63", func(t *testing.T) {
			r := ppTasks(t, "Result: 42 tests—confirmed\n")
			ok(t, "em dash (right, glued) is caught", r.rc != 0, r)
		}},
		{"case 64", func(t *testing.T) {
			r := ppTasks(t, "Result: “42 tests’\n")
			ok(t, "curly quote (right, glued) is caught", r.rc != 0, r)
		}},
		{"case 65", func(t *testing.T) {
			r := ppTasks(t, "Result: (42 tests）\n")
			ok(t, "fullwidth paren (right, glued) is caught", r.rc != 0, r)
		}},
		{"case 66-1", func(t *testing.T) {
			r := ppTasks(t, "Result — 42 tests\n")
			ok(t, "em dash (left, spaced) is still caught", r.rc != 0, r)
		}},
		{"case 66-2", func(t *testing.T) {
			r := ppTasks(t, "Result-42 tests\n")
			ok(t, "ASCII hyphen (left, glued) is still caught", r.rc != 0, r)
		}},
		// SECTION: Pass-7 fix wave — Important 7: the 10 MiB size cap.
		{"case 67", func(t *testing.T) {
			r := ppTasks(t, strings.Repeat("x", 10*1024*1024+1024))
			ok(t, "a tasks.md over the 10 MiB cap is refused (exit 2)", r.rc == 2, r)
			ok(t, "size cap: the message names the file and the cap", has(r.out, "tasks.md", "exceeds the", "byte cap"), r)
		}},
		// SECTION: Pass-7 fix wave — Important 8: a per-file failure never
		// masks a real violation in ANOTHER file, nor stops the scan.
		{"case 68", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, "spectre/changes/bad-encoding-change/tasks.md", "\xff\xfe not valid utf-8\n")
			ppWrite(t, root, ppDemo+"tasks.md", "Baseline: 197 tests\n")
			r := ppRun(root)
			ok(t, "a bad-encoding file elsewhere does not mask a real violation (exit 1, not 2)", r.rc == 1, r)
			ok(t, "mixed bad-encoding + violation: the real violation is still reported", has(r.out, "demo-change/tasks.md:1: "+ppNoProv), r)
			ok(t, "mixed bad-encoding + violation: the bad file is still named", has(r.out, "cannot decode as UTF-8"), r)
		}},
		{"case 69", func(t *testing.T) {
			root := ppFixture(t)
			os.RemoveAll(filepath.Join(root, ppDemo))
			ppWrite(t, root, "spectre/changes/bad-encoding-change/tasks.md", "\xff\xfe not valid utf-8\n")
			r := ppRun(root)
			ok(t, "a bad-encoding file with no other violations still refuses a clean run (exit 2)", r.rc == 2, r)
		}},
		// SECTION: Pass-7 fix wave — Critical 2 REGRESSION FIX: the cap is
		// measured against the fence's CONTAINER content column.
		{"case 70", func(t *testing.T) {
			r := ppTasks(t, ">```bash verified:t\n>ran 500 tests\n>```\n")
			ok(t, "blockquote rel indent 0: real fence recognized (no false claim report)", r.rc == 0, r)
		}},
		{"case 71", func(t *testing.T) {
			r := ppTasks(t, ">   ```bash verified:t\n>   ran 500 tests\n>   ```\n")
			ok(t, "blockquote rel indent 2 (coordinator's exact repro): real fence recognized", r.rc == 0, r)
			ok(t, "blockquote rel indent 2: zero violations reported", ppOut(r) == ppStated1, r)
		}},
		{"case 72", func(t *testing.T) {
			r := ppTasks(t, ">    ```bash verified:t\n>    ran 500 tests\n>    ```\n")
			ok(t, "blockquote rel indent 3: real fence recognized", r.rc == 0, r)
		}},
		{"case 73", func(t *testing.T) {
			r := ppTasks(t, ">     ```bash verified:t\n>     ran 500 tests\n>     ```\n")
			ok(t, "blockquote rel indent 4: over budget, claim reported (not a fence)", r.rc != 0, r)
			ok(t, "blockquote rel indent 4: the '500 tests' claim is reported", has(r.out, ppNoProv), r)
		}},
		{"case 74", func(t *testing.T) {
			r := ppTasks(t, "- item\n\n  ```bash verified:t\n  ran 500 tests\n  ```\n")
			ok(t, "list rel indent 0 (absolute column 2): bare fence within the 0-3 budget, unconditionally a fence", r.rc == 0, r)
		}},
		// 75: reversed from pass 7 on purpose — the ambient list tracker is
		// deleted, so a bare 4+-column line refuses to guess.
		{"case 75", func(t *testing.T) {
			r := ppTasks(t, "- item\n\n     ```bash verified:t\n     echo hi\n     ```\n")
			ok(t, "list rel indent 3 (absolute column 5): now aborts (content-classification) instead of silently resolving via list context", r.rc == 4, r)
			ok(t, "list rel indent 3: the abort names the opening line", has(r.out, "tasks.md:3: fence-like run indented 4+ columns"), r)
		}},
		{"case 76", func(t *testing.T) {
			r := ppTasks(t, "- item\n\n      ```bash verified:t\n      ran 500 tests\n      ```\n")
			ok(t, "list rel indent 4 (absolute column 6): aborts (content-classification)", r.rc == 4, r)
		}},
		{"case 77-1", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:t\nran 500 tests\n```\n")
			ok(t, "un-contained rel/abs indent 0: still a real fence", r.rc == 0, r)
		}},
		{"case 77-2", func(t *testing.T) {
			r := ppTasks(t, "   ```bash verified:t\n   ran 500 tests\n   ```\n")
			ok(t, "un-contained rel/abs indent 3: still a real fence", r.rc == 0, r)
		}},
		{"case 77-3", func(t *testing.T) {
			r := ppTasks(t, "    ```bash verified:t\n    ran 500 tests\n    ```\n")
			ok(t, "un-contained rel/abs indent 4: aborts (content-classification), no active list/blockquote to be relative to", r.rc == 4, r)
		}},
		{"case 77-4", func(t *testing.T) {
			r := ppTasks(t, "        ```bash verified:t\n        ran 500 tests\n        ```\n")
			ok(t, "un-contained rel/abs indent 8: aborts (content-classification)", r.rc == 4, r)
		}},
		{"case 78", func(t *testing.T) {
			r := ppTasks(t, "\t```bash verified:t\n\tran 500 tests\n\t```\n")
			ok(t, "a tab-indented un-contained fence (tab = 4 columns): aborts (content-classification)", r.rc == 4, r)
		}},
		// SECTION: Pass-12 fix wave — Critical 1: a tab AFTER a container
		// prefix token expands relative to its TRUE absolute column. Content
		// columns are hand-computed by CommonMark's tab-expansion rule.
		// 78a: ' -<TAB>' — ' ' 0->1, '-' 1->2, tab 2->4.
		{"case 78a-1", func(t *testing.T) {
			r := ppTasks(t, " -\t```bash verified:t\n    ran 1 tests\n    ```\n")
			ok(t, "tab after bullet marker, tagged, true content col 4: exit 0", r.rc == 0, r)
		}},
		{"case 78a-2", func(t *testing.T) {
			r := ppTasks(t, " -\t```bash\n    ran 1 tests\n    ```\n")
			ok(t, "tab after bullet marker, untagged, true content col 4: exit 1", r.rc == 1, r)
			ok(t, "tab after bullet marker, untagged: opening line reported, not 'never closed'", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		// 78b: ' 1.<TAB>' — ' ' 0->1, '1.' 1->3, tab 3->4.
		{"case 78b-1", func(t *testing.T) {
			r := ppTasks(t, " 1.\t```bash verified:t\n    ran 1 tests\n    ```\n")
			ok(t, "tab after ordered marker, tagged, true content col 4: exit 0", r.rc == 0, r)
		}},
		{"case 78b-2", func(t *testing.T) {
			r := ppTasks(t, " 1.\t```bash\n    ran 1 tests\n    ```\n")
			ok(t, "tab after ordered marker, untagged, true content col 4: exit 1", r.rc == 1, r)
			ok(t, "tab after ordered marker, untagged: opening line reported, not 'never closed'", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		// 78c: '- [ ]<TAB>' — '- ' 0->2, '[ ]' 2->5, tab 5->8.
		{"case 78c-1", func(t *testing.T) {
			r := ppTasks(t, "- [ ]\t```bash verified:t\n        ran 1 tests\n        ```\n")
			ok(t, "tab after checkbox, tagged, true content col 8: exit 0", r.rc == 0, r)
		}},
		{"case 78c-2", func(t *testing.T) {
			r := ppTasks(t, "- [ ]\t```bash\n        ran 1 tests\n        ```\n")
			ok(t, "tab after checkbox, untagged, true content col 8: exit 1", r.rc == 1, r)
			ok(t, "tab after checkbox, untagged: opening line reported, not 'never closed'", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		// 78d: '> <TAB>' prose, then a SEPARATE untagged fence — the silent-miss
		// shape; the tab sits after the marker's one space.
		{"case 78d", func(t *testing.T) {
			r := ppTasks(t, "> \tordinary blockquoted text, not a fence\n\n```bash\necho hi\n```\n")
			ok(t, "tab after blockquote marker + later untagged fence: exit 1, not silently swallowed", r.rc == 1, r)
			ok(t, "tab after blockquote marker + later untagged fence: the fence itself is flagged, not silently skipped", has(r.out, "tasks.md:3: "+ppNoTag), r)
		}},
		{"case 79", func(t *testing.T) {
			r := ppTasks(t, ">```bash\n>ran 500 tests\n>```\n")
			ok(t, "untagged fence inside a blockquote (rel indent 0) is still flagged", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		{"case 80", func(t *testing.T) {
			r := ppTasks(t, "- item\n  >```bash verified:t\n  >ran 500 tests\n  >```\n")
			ok(t, "nested blockquote-inside-list, rel indent 0: real fence recognized", r.rc == 0, r)
		}},
		// SECTION: Pass-8 fix wave — Critical 1: the ambient list-column
		// tracker is DELETED; ambiguous list context fails loud (exit 4).
		{"case 81", func(t *testing.T) {
			r := ppTasks(t, "- item\n  ```bash verified:t\n  echo hi\n  ```\n")
			ok(t, "'- item' then a 2-space-indented tagged fence: still exit 0 (unambiguous)", r.rc == 0, r)
		}},
		{"case 82", func(t *testing.T) {
			r := ppTasks(t, "   ```bash verified:t\n   echo hi\n   ```\n")
			ok(t, "a 3-space-indented bare tagged fence: exit 0", r.rc == 0, r)
		}},
		{"case 83", func(t *testing.T) {
			r := ppTasks(t, "    ```bash verified:t\n    echo hi\n    ```\n")
			ok(t, "a 4-space-indented bare fence run aborts (exit 4)", r.rc == 4, r)
			ok(t, "4-space bare fence abort: message names the opening line", has(r.out, "tasks.md:1: fence-like run indented 4+ columns"), r)
		}},
		// 84: a thematic break opened a phantom list scope under the deleted
		// tracker (a silent miss); it now aborts like any bare 4+ run.
		{"case 84", func(t *testing.T) {
			r := ppTasks(t, "- - -\n    ```bash\n    echo hi\n    ```\n")
			ok(t, "'- - -' (thematic break) then a 4-space fence: no longer a silent miss — aborts (exit 4)", r.rc == 4, r)
		}},
		{"case 84b", func(t *testing.T) {
			r := ppTasks(t, "* * *\n    ```bash\n    echo hi\n    ```\n")
			ok(t, "'* * *' (thematic break) then a 4-space fence: aborts (exit 4)", r.rc == 4, r)
		}},
		{"case 85", func(t *testing.T) {
			r := ppTasks(t, "-      item\n    ```bash\n    echo hi\n    ```\n")
			ok(t, "'-      item' (wide gap) then a 4-space fence: no longer a silent miss — aborts (exit 4)", r.rc == 4, r)
		}},
		{"case 86", func(t *testing.T) {
			r := ppTasks(t, "- outer\n  - inner\n  ```bash verified:t\n  echo hi\n  ```\n")
			ok(t, "nested list, dedent, fence at 2-space indent: exit 0", r.rc == 0, r)
		}},
		// 87: the abort is scoped to fence RUNS, not to indentation in
		// general.
		{"case 87", func(t *testing.T) {
			r := ppTasks(t, "    Baseline: ran 500 tests with no fence in sight.\n")
			ok(t, "4-space-indented ordinary prose with an untagged claim: exit 1, not an abort", r.rc == 1, r)
			ok(t, "4-space prose claim: the claim is reported", has(r.out, "tasks.md:1: "+ppNoProv), r)
			ok(t, "4-space prose claim: no false abort", !has(r.out, "fence-like run"), r)
		}},
		{"case 88", func(t *testing.T) {
			r := ppTasks(t, "- ```bash verified:t\n  echo hi\n  ```\n")
			ok(t, "list-marker-prefixed fence (relative 0, same line): exit 0", r.rc == 0, r)
		}},
	}
}

// ppMarkerWidth is the harness's marker_width_case: a fence opened on
// `<marker>` with its continuation and closer at the marker's TRUE content
// column passes tagged, and is still caught untagged, naming the opening
// line.
func ppMarkerWidth(name, marker string, col int, label string) ppCase {
	return ppCase{name, func(t *testing.T) {
		pad := strings.Repeat(" ", col)
		r := ppTasks(t, marker+"```bash verified:ran it locally\n"+pad+"echo hi\n"+pad+"```\n")
		ok(t, fmt.Sprintf("%s: tagged fence, continuation/closer at true content column (%d): exit 0", label, col), r.rc == 0, r)
		r = ppTasks(t, marker+"```bash\n"+pad+"echo hi\n"+pad+"```\n")
		ok(t, label+": untagged fence at the same content column is still caught (exit 1)", r.rc != 0, r)
		ok(t, label+": untagged fence reports the opening line, not 'never closed'", has(r.out, "tasks.md:1: "+ppNoTag), r)
	}}
}

// ppBqListGap is the harness's bq_list_gap_case: an untagged fence opened
// behind bqPrefix, gap columns and a bullet, a claim after it closes, and a
// second untagged bare fence — all three reported whatever the gap. The
// continuation/closer lines sit at the TRUE content column: the gap plus
// the marker's own 2 (pass-10 fix wave, Critical 2).
func ppBqListGap(name, bqPrefix string, gap int, label string) ppCase {
	return ppCase{name, func(t *testing.T) {
		gapPad := strings.Repeat(" ", gap)
		content := bqPrefix + gapPad + "  "
		r := ppTasks(t, bqPrefix+gapPad+"- ```python\n"+content+"echo untagged fence 1\n"+content+"```\nBaseline: 194 tests\n```python\necho untagged fence 2\n```\n")
		l := fmt.Sprintf("%s (gap=%d)", label, gap)
		ok(t, l+": exit 1, not masked", r.rc == 1, r)
		ok(t, l+": first untagged fence reported", has(r.out, "tasks.md:1: "+ppNoTag), r)
		ok(t, l+": the '194 tests' claim is reported, not masked", has(r.out, "tasks.md:4: "+ppNoProv), r)
		ok(t, l+": second untagged fence reported, not swallowed", has(r.out, "tasks.md:5: "+ppNoTag), r)
	}}
}

// ppViolatingSibling is the harness's violating_sibling_case_dir: a
// "<prefix>-violating-change" carrying one unattributed claim.
func ppViolatingSibling(t *testing.T, root, prefix string) {
	t.Helper()
	ppWrite(t, root, "spectre/changes/"+prefix+"-violating-change/tasks.md", "Baseline: 321 tests\n")
}

// ppBare is new_fixture with demo-change removed.
func ppBare(t *testing.T) string {
	t.Helper()
	root := ppFixture(t)
	os.RemoveAll(filepath.Join(root, ppDemo))
	return root
}

const ppIndented4 = "    ```bash\n    echo hi\n    ```\n"

// ppCasesPass8To10 runs from the pass-8 fix wave's Critical 2 to the
// pass-10 harness blind spot's circular-symlink pair.
func ppCasesPass8To10() []ppCase {
	return []ppCase{
		// SECTION: Pass-8 fix wave — Critical 2: an abort in one change
		// directory never skips a later one; 1 outranks 4. Both sort orders.
		{"case 89", func(t *testing.T) {
			root := ppBare(t)
			ppWrite(t, root, "spectre/changes/aaa-abort-change/tasks.md", ppIndented4)
			ppWrite(t, root, "spectre/changes/zzz-violation-change/tasks.md", "Baseline: 197 tests\n")
			r := ppRun(root)
			ok(t, "abort dir sorts first: violation elsewhere still wins (exit 1, not 4)", r.rc == 1, r)
			ok(t, "abort dir sorts first: the later directory's violation is still reported, not skipped", has(r.out, "zzz-violation-change/tasks.md:1: "+ppNoProv), r)
			ok(t, "abort dir sorts first: the abort itself is still reported", has(r.out, "aaa-abort-change/tasks.md:1: fence-like run indented 4+ columns"), r)
		}},
		{"case 90", func(t *testing.T) {
			root := ppBare(t)
			ppWrite(t, root, "spectre/changes/aaa-violation-change/tasks.md", "Baseline: 197 tests\n")
			ppWrite(t, root, "spectre/changes/zzz-abort-change/tasks.md", ppIndented4)
			r := ppRun(root)
			ok(t, "violation dir sorts first: the exit code is still 1, not overwritten to 4 by a later abort", r.rc == 1, r)
			ok(t, "violation dir sorts first: the violation is reported", has(r.out, "aaa-violation-change/tasks.md:1: "+ppNoProv), r)
			ok(t, "violation dir sorts first: the later directory's abort is still scanned and reported", has(r.out, "zzz-abort-change/tasks.md:1: fence-like run indented 4+ columns"), r)
		}},
		{"case 91", func(t *testing.T) {
			root := ppBare(t)
			ppWrite(t, root, "spectre/changes/aaa-abort-change/tasks.md", ppIndented4)
			ppWrite(t, root, "spectre/changes/zzz-clean-change/tasks.md", "```bash verified:t\necho hi\n```\n")
			r := ppRun(root)
			ok(t, "an abort beside a CLEAN directory (no violations anywhere): exit 4", r.rc == 4, r)
		}},
		// SECTION: Pass-8 fix wave — Important 3: a closing fence is capped
		// at the same 0-3 column budget as an opener.
		{"case 92", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:ran it\necho start\n      ```\nRan 500 tests and it passed.\n```\n")
			ok(t, "an over-indented (6-column) closer does not close the fence early; the real closer does", r.rc == 0, r)
			ok(t, "over-indented closer: zero violations reported (no false 'never closed', no false claim)", ppOut(r) == ppStated1, r)
		}},
		{"case 92b", func(t *testing.T) {
			r := ppTasks(t, "> ```bash verified:ran it\n> echo start\n>       ```\n> Ran 500 tests and it passed.\n> ```\n")
			ok(t, "blockquoted over-indented closer does not close the fence early; the real closer does", r.rc == 0, r)
		}},
		// SECTION: Pass-9 fix wave — Critical 1: FenceContext records the
		// list's content column and the closing test requires it exactly.
		ppMarkerWidth("case 93", "- ", 2, "bullet '- '"),
		ppMarkerWidth("case 94", "* ", 2, "bullet '* '"),
		ppMarkerWidth("case 95", "- [ ] ", 6, "checkbox '- [ ] '"),
		ppMarkerWidth("case 96", "1. ", 3, "ordered '1. '"),
		ppMarkerWidth("case 97", "10. ", 4, "two-digit ordered '10. '"),
		ppMarkerWidth("case 98", "1) ", 3, "ordered-paren '1) '"),
		{"case 99", func(t *testing.T) {
			r := ppTasks(t, "- [ ] ```bash verified:ran it locally\n  echo hi\n  ```\n")
			ok(t, "checkbox fence, closer at col 2 (short of true content column 6): no longer a false close", r.rc != 0, r)
			ok(t, "checkbox fence, closer at col 2: correctly reported as never closed", has(r.out, "tasks.md:1: fenced code block never closed"), r)
		}},
		{"case 100", func(t *testing.T) {
			r := ppTasks(t, "10. ```bash verified:ran it locally\n  echo hi\n  ```\n")
			ok(t, "ordered '10. ' fence, closer at col 2 (short of true content column 4): no longer a false close", r.rc != 0, r)
			ok(t, "ordered '10. ' fence, closer at col 2: correctly reported as never closed", has(r.out, "tasks.md:1: fenced code block never closed"), r)
		}},
		// SECTION: Pass-9 fix wave — Critical 2: a 0-3 column gap between a
		// blockquote marker and a list marker no longer defeats the
		// classifier.
		ppBqListGap("case 101", "> ", 0, "depth-1 blockquote, list-marker gap"),
		ppBqListGap("case 102", "> ", 1, "depth-1 blockquote, list-marker gap"),
		ppBqListGap("case 103", "> ", 2, "depth-1 blockquote, list-marker gap"),
		ppBqListGap("case 104", "> ", 3, "depth-1 blockquote, list-marker gap"),
		ppBqListGap("case 105", "> > ", 0, "depth-2 blockquote, list-marker gap"),
		ppBqListGap("case 106", "> > ", 1, "depth-2 blockquote, list-marker gap"),
		ppBqListGap("case 107", "> > ", 2, "depth-2 blockquote, list-marker gap"),
		ppBqListGap("case 108", "> > ", 3, "depth-2 blockquote, list-marker gap"),
		// 108b-108d: the pass-10 Critical 2 repro — true content column 3 =
		// gap 1 + marker width 2.
		{"case 108b", func(t *testing.T) {
			r := ppTasks(t, ">  - ```bash verified:manual\n>    code line one\n>    ```\n")
			ok(t, "gap-1 fixture: closer at the TRUE content column (gap+marker) closes correctly, exit 0", r.rc == 0, r)
		}},
		{"case 108c", func(t *testing.T) {
			r := ppTasks(t, ">  - ```bash verified:manual\n>    code line one\n>    ```\nBaseline: 55 tests\n")
			ok(t, "gap-1 fixture: claim after the real closer is reported, not masked", r.rc == 1, r)
			ok(t, "gap-1 fixture: exactly the '55 tests' claim is reported", has(r.out, "tasks.md:4: "+ppNoProv), r)
		}},
		{"case 108d", func(t *testing.T) {
			r := ppTasks(t, ">  - ```bash verified:manual\n>    code line one\n>   ```\n")
			ok(t, "gap-1 fixture: a closer one column short of the true content column does not falsely close", r.rc != 0, r)
			ok(t, "gap-1 fixture: short closer correctly leaves the fence unclosed, not silently closed", has(r.out, "fenced code block never closed"), r)
		}},
		// SECTION: Pass-10 fix wave — Critical 3: a 4+ column gap between a
		// blockquote marker and a list marker refuses loudly (the operator's
		// adjudication: Slot 1's reading governs).
		{"case 109", func(t *testing.T) {
			r := ppTasks(t, ">     - ```python\n>     echo x\n>     ```\nBaseline: 42 tests\n")
			ok(t, "4+ column gap, list marker, untagged fence: loud abort (exit 4), never silent prose", r.rc == 4, r)
			ok(t, "4+ column gap after blockquote: abort names the ambiguous line", has(r.out, "tasks.md:1: line has a fence-like run", "refusing to guess"), r)
			ok(t, "4+ column gap after blockquote: no false clean run", !has(r.out, "all provenance stated"), r)
		}},
		{"case 109b", func(t *testing.T) {
			r := ppTasks(t, ">     Confirmed the fence style uses ``` markers, all good.\n")
			ok(t, "genuine prose mentioning a fence behind a 4+ column blockquote gap: still prose, no abort", r.rc == 0, r)
		}},
		// SECTION: Pass-10 fix wave — the harness blind spot: each failure
		// mode paired with a VIOLATING sibling, both sort orders; the
		// violation is reported and the exit is 1.
		{"case 110", func(t *testing.T) {
			root := ppBare(t)
			locked := filepath.Join(root, "spectre/changes/locked-change")
			ppWrite(t, root, "spectre/changes/locked-change/tasks.md", "irrelevant\n")
			ppChmod(t, locked, 0)
			ppViolatingSibling(t, root, "a")
			r := ppRun(root)
			ppChmod(t, locked, 0o755)
			ok(t, "locked dir + violating sibling (violation sorts first): exit 1, not masked", r.rc == 1, r)
			ok(t, "locked dir + violation (violation first): violation reported", has(r.out, "321 tests", "no measured") || has(r.out, "numeric claim with no measured"), r)
		}},
		{"case 111", func(t *testing.T) {
			root := ppBare(t)
			locked := filepath.Join(root, "spectre/changes/locked-change")
			ppWrite(t, root, "spectre/changes/locked-change/tasks.md", "irrelevant\n")
			ppChmod(t, locked, 0)
			ppViolatingSibling(t, root, "z")
			r := ppRun(root)
			ppChmod(t, locked, 0o755)
			ok(t, "locked dir + violating sibling (violation sorts last): exit 1, not masked", r.rc == 1, r)
			ok(t, "locked dir + violation (violation last): violation reported", has(r.out, "numeric claim with no measured"), r)
		}},
		{"case 112", func(t *testing.T) {
			root := ppBare(t)
			ppSymlink(t, filepath.Join(root, "spectre/changes/does-not-exist"), filepath.Join(root, "spectre/changes/dangling-change"))
			ppViolatingSibling(t, root, "a")
			r := ppRun(root)
			ok(t, "dangling symlink + violating sibling (violation sorts first): exit 1, not masked", r.rc == 1, r)
			ok(t, "dangling symlink + violation (violation first): violation reported", has(r.out, "numeric claim with no measured"), r)
		}},
		{"case 113", func(t *testing.T) {
			root := ppBare(t)
			ppSymlink(t, filepath.Join(root, "spectre/changes/does-not-exist"), filepath.Join(root, "spectre/changes/dangling-change"))
			ppViolatingSibling(t, root, "z")
			r := ppRun(root)
			ok(t, "dangling symlink + violating sibling (violation sorts last): exit 1, not masked", r.rc == 1, r)
			ok(t, "dangling symlink + violation (violation last): violation reported", has(r.out, "numeric claim with no measured"), r)
		}},
		{"case 114", func(t *testing.T) {
			root := ppBare(t)
			ppSymlink(t, filepath.Join(root, "spectre/changes/circular-a"), filepath.Join(root, "spectre/changes/circular-b"))
			ppSymlink(t, filepath.Join(root, "spectre/changes/circular-b"), filepath.Join(root, "spectre/changes/circular-a"))
			ppViolatingSibling(t, root, "a")
			r := ppRun(root)
			ok(t, "circular symlink + violating sibling (violation sorts first): exit 1, not masked", r.rc == 1, r)
			ok(t, "circular symlink + violation (violation first): violation reported", has(r.out, "numeric claim with no measured"), r)
		}},
		{"case 115", func(t *testing.T) {
			root := ppBare(t)
			ppSymlink(t, filepath.Join(root, "spectre/changes/circular-a"), filepath.Join(root, "spectre/changes/circular-b"))
			ppSymlink(t, filepath.Join(root, "spectre/changes/circular-b"), filepath.Join(root, "spectre/changes/circular-a"))
			ppViolatingSibling(t, root, "z")
			r := ppRun(root)
			ok(t, "circular symlink + violating sibling (violation sorts last): exit 1, not masked", r.rc == 1, r)
			ok(t, "circular symlink + violation (violation last): violation reported", has(r.out, "numeric claim with no measured"), r)
		}},
	}
}

// ppSiblingCase is one of cases 116-123: a failing change directory
// written by fail, beside a violating sibling sorting before ("a") or after
// ("z") it; the violation is reported and wins exit 1.
func ppSiblingCase(name, prefix, label, reported string, fail func(t *testing.T, root string) (restore func())) ppCase {
	return ppCase{name, func(t *testing.T) {
		root := ppFixture(t)
		restore := fail(t, root)
		ppViolatingSibling(t, root, prefix)
		r := ppRun(root)
		if restore != nil {
			restore()
		}
		ok(t, label, r.rc == 1, r)
		ok(t, reported, has(r.out, "numeric claim with no measured"), r)
	}}
}

func ppUnreadableChange(t *testing.T, root string) func() {
	p := filepath.Join(root, "spectre/changes/unreadable-change/tasks.md")
	ppWrite(t, root, "spectre/changes/unreadable-change/tasks.md", "irrelevant\n")
	ppChmod(t, p, 0)
	return func() { ppChmod(t, p, 0o644) }
}

func ppBadEncodingChange(t *testing.T, root string) func() {
	ppWrite(t, root, "spectre/changes/bad-encoding-change/tasks.md", "\xff\xfe not valid utf-8\n")
	return nil
}

func ppOversizedChange(t *testing.T, root string) func() {
	ppWrite(t, root, "spectre/changes/oversized-change/tasks.md", strings.Repeat("a", 10*1024*1024+1))
	return nil
}

func ppAbortChange(t *testing.T, root string) func() {
	ppWrite(t, root, "spectre/changes/abort-change/tasks.md", "- - ```bash\necho hi\n```\n")
	return nil
}

// ppOverIndentedMessage is the harness's
// assert_over_indented_marker_message (Minor 6, pass 13): a line that
// carries a marker but exceeds the outer-indent cap names the outer
// indentation, never the bare line's "no container prefix" text.
func ppOverIndentedMessage(t *testing.T, label string, r tcfRes) {
	t.Helper()
	ok(t, label+": abort message names the over-indented container prefix as the cause", has(r.out, "container prefix starting 4+ columns in"), r)
	ok(t, label+": abort message does not claim the line has no container prefix", !has(r.out, "with no container prefix on this line"), r)
}

// ppIndentCloser is the harness's indent_closer_case: a `- ` bullet fence
// indented indent columns, continuation/closer at the true content column
// (indent + 2, the width of the "- " literal below) plus offset. Indent
// 0-3 closes normally; indent 4-5 refuses (exit 4) tagged or not — CORRECTED
// at pass 12 (Critical 2): CommonMark caps UN-CONTAINED indentation at 0-3
// columns before ANY container marker is recognised, and this parser cannot
// tell "indent 4 inside an open list" from "indent 4 at the top level".
func ppIndentCloser(indent, offset int, label string) ppCase {
	return ppCase{fmt.Sprintf("case indent=%d offset=%d", indent, offset), func(t *testing.T) {
		const bulletMarkerWidth = 2
		markerPad := strings.Repeat(" ", indent)
		pad := strings.Repeat(" ", indent+bulletMarkerWidth+offset)
		shape := fmt.Sprintf("indent=%d, closer=content_col+%d", indent, offset)
		tagged := ppTasks(t, markerPad+"- ```bash verified:ran it locally\n"+pad+"echo hi\n"+pad+"```\n")
		if indent > 3 {
			ok(t, label+": tagged, "+shape+": exit 4 (outer indent capped before container recognition)", tagged.rc == 4, tagged)
			ppOverIndentedMessage(t, label+": tagged, "+shape, tagged)
		} else {
			ok(t, label+": tagged, "+shape+": exit 0", tagged.rc == 0, tagged)
		}
		untagged := ppTasks(t, markerPad+"- ```bash\n"+pad+"echo hi\n"+pad+"```\n")
		if indent > 3 {
			ok(t, label+": untagged, "+shape+": exit 4 (tag is irrelevant to a refusal)", untagged.rc == 4, untagged)
			ppOverIndentedMessage(t, label+": untagged, "+shape, untagged)
			return
		}
		ok(t, label+": untagged, "+shape+": exit 1", untagged.rc == 1, untagged)
		ok(t, label+": untagged, "+shape+": opening line reported, not 'never closed'", has(untagged.out, "tasks.md:1: "+ppNoTag), untagged)
	}}
}

// ppCasesPass10To11 runs from the pass-10 blind spot's unreadable-file pair
// through the pass-11 marker-indent x closer-column matrix.
func ppCasesPass10To11() []ppCase {
	cases := []ppCase{
		ppSiblingCase("case 116", "a", "unreadable tasks.md + violating sibling (violation sorts first): exit 1, not masked",
			"unreadable tasks.md + violation (violation first): violation reported", ppUnreadableChange),
		ppSiblingCase("case 117", "z", "unreadable tasks.md + violating sibling (violation sorts last): exit 1, not masked",
			"unreadable tasks.md + violation (violation last): violation reported", ppUnreadableChange),
		ppSiblingCase("case 118", "a", "bad encoding + violating sibling (violation sorts first): exit 1, not masked",
			"bad encoding + violation (violation first): violation reported", ppBadEncodingChange),
		ppSiblingCase("case 119", "z", "bad encoding + violating sibling (violation sorts last): exit 1, not masked",
			"bad encoding + violation (violation last): violation reported", ppBadEncodingChange),
		ppSiblingCase("case 120", "a", "oversized tasks.md + violating sibling (violation sorts first): exit 1, not masked",
			"oversized tasks.md + violation (violation first): violation reported", ppOversizedChange),
		ppSiblingCase("case 121", "z", "oversized tasks.md + violating sibling (violation sorts last): exit 1, not masked",
			"oversized tasks.md + violation (violation last): violation reported", ppOversizedChange),
		ppSiblingCase("case 122", "a", "classification abort + violating sibling (violation sorts first): exit 1, not masked by exit 4",
			"abort + violation (violation first): violation reported", ppAbortChange),
		ppSiblingCase("case 123", "z", "classification abort + violating sibling (violation sorts last): exit 1, not masked by exit 4",
			"abort + violation (violation last): violation reported", ppAbortChange),
	}
	// SECTION: Pass-11 fix wave — Critical 1: the outermost leading
	// indentation in front of a list marker is folded into
	// list_content_col. Indents 0-5 at closer offsets 0-3.
	for indent := 0; indent <= 5; indent++ {
		for offset := 0; offset <= 3; offset++ {
			cases = append(cases, ppIndentCloser(indent, offset, "marker-indent x closer-column (bullet '- ')"))
		}
	}
	return cases
}

// ppBqIndent is the harness's blockquote_indent_case: indent_closer_case's
// blockquote equivalent. The marker repeats on every line (no lazy
// continuation), then offset columns before the fence run. Indent 4-5
// refuses for the same CommonMark indented-code-block reason as the bullet.
func ppBqIndent(indent, offset int, label string) ppCase {
	return ppCase{fmt.Sprintf("case bq indent=%d offset=%d", indent, offset), func(t *testing.T) {
		markerPad := strings.Repeat(" ", indent)
		pad := strings.Repeat(" ", offset)
		shape := fmt.Sprintf("indent=%d, closer=+%d", indent, offset)
		body := func(info string) string {
			return markerPad + "> ```bash" + info + "\n" + markerPad + "> " + pad + "echo hi\n" + markerPad + "> " + pad + "```\n"
		}
		tagged := ppTasks(t, body(" verified:ran it locally"))
		if indent > 3 {
			ok(t, label+": tagged, "+shape+": exit 4 (outer indent capped before container recognition)", tagged.rc == 4, tagged)
			ppOverIndentedMessage(t, label+": tagged, "+shape, tagged)
		} else {
			ok(t, label+": tagged, "+shape+": exit 0", tagged.rc == 0, tagged)
		}
		untagged := ppTasks(t, body(""))
		if indent > 3 {
			ok(t, label+": untagged, "+shape+": exit 4 (tag is irrelevant to a refusal)", untagged.rc == 4, untagged)
			ppOverIndentedMessage(t, label+": untagged, "+shape, untagged)
			return
		}
		ok(t, label+": untagged, "+shape+": exit 1", untagged.rc == 1, untagged)
		ok(t, label+": untagged, "+shape+": opening line reported, not 'never closed'", has(untagged.out, "tasks.md:1: "+ppNoTag), untagged)
	}}
}

func ppMkfifo(t *testing.T, path string) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Fatal(err)
	}
}

// ppContainmentSibling is one of cases 124-131: a containment failure built
// by fail beside a violating sibling; exit 3 wins and the scan continues.
// message, when set, is the containment message's own ok: line.
func ppContainmentSibling(name, prefix, label, reported, message, want string, fail func(t *testing.T, root string)) ppCase {
	return ppCase{name, func(t *testing.T) {
		root := ppFixture(t)
		fail(t, root)
		ppViolatingSibling(t, root, prefix)
		r := ppRun(root)
		ok(t, label, r.rc == 3, r)
		ok(t, reported, has(r.out, "numeric claim with no measured"), r)
		if message != "" {
			parts := strings.Split(want, "*")
			ok(t, message, has(r.out, parts...), r)
		}
	}}
}

func ppSymlinkedTasks(t *testing.T, root string) {
	os.RemoveAll(filepath.Join(root, ppDemo))
	ppSymlink(t, "/dev/zero", filepath.Join(root, "spectre/changes/symlinked-tasks-change/tasks.md"))
}

func ppEscapedChange(t *testing.T, root string) {
	os.RemoveAll(filepath.Join(root, ppDemo))
	ppWrite(t, root, "outside-changes/escaped-change/tasks.md", "irrelevant\n")
	ppSymlink(t, filepath.Join(root, "outside-changes/escaped-change"), filepath.Join(root, "spectre/changes/escaped-change"))
}

func ppAliasedChange(t *testing.T, root string) {
	os.RemoveAll(filepath.Join(root, ppDemo))
	ppWrite(t, root, "spectre/changes/real-target-change/tasks.md", "irrelevant\n")
	ppSymlink(t, filepath.Join(root, "spectre/changes/real-target-change"), filepath.Join(root, "spectre/changes/aliased-change"))
}

func ppFifoChange(t *testing.T, root string) {
	ppMkfifo(t, filepath.Join(root, "spectre/changes/fifo-change/tasks.md"))
}

// ppCasesPass11Containment runs the pass-11 blockquote matrix and the
// containment-beside-a-violation pairs (cases 124-131).
func ppCasesPass11Containment() []ppCase {
	var cases []ppCase
	for indent := 0; indent <= 5; indent++ {
		for offset := 0; offset <= 3; offset++ {
			cases = append(cases, ppBqIndent(indent, offset, "marker-indent x closer-column (blockquote '> ')"))
		}
	}
	// SECTION: Pass-11 fix wave — Critical 2: a containment refusal (exit 3)
	// never stops the scan and still outranks a violation elsewhere. The
	// symlinked-change-directory and directory-escape fixtures are one source
	// branch (a symlink's realpath is never its own path); both shapes are
	// built for honesty about what is constructed. The two "cannot resolve"
	// branches need a directory to vanish between lstat and stat — a race,
	// not a fixture — and are not exercised (the TOCTOU known limitation).
	return append(cases,
		ppContainmentSibling("case 124", "a", "symlinked tasks.md + violating sibling (violation sorts first): exit 3, scan continues",
			"symlinked tasks.md + violation (violation first): sibling violation reported",
			"symlinked tasks.md + violation (violation first): containment message reported", "tasks.md is a symlink", ppSymlinkedTasks),
		ppContainmentSibling("case 125", "z", "symlinked tasks.md + violating sibling (violation sorts last): exit 3, scan continues",
			"symlinked tasks.md + violation (violation last): sibling violation reported", "", "", ppSymlinkedTasks),
		ppContainmentSibling("case 126", "a", "directory escape + violating sibling (violation sorts first): exit 3, scan continues",
			"directory escape + violation (violation first): sibling violation reported",
			"directory escape + violation (violation first): containment message reported", "outside*symlink escape", ppEscapedChange),
		ppContainmentSibling("case 127", "z", "directory escape + violating sibling (violation sorts last): exit 3, scan continues",
			"directory escape + violation (violation last): sibling violation reported", "", "", ppEscapedChange),
		ppContainmentSibling("case 128", "a", "symlinked change directory (in-tree target) + violating sibling (violation sorts first): exit 3",
			"symlinked change directory (in-tree target) + violation (violation first): sibling violation reported", "", "", ppAliasedChange),
		ppContainmentSibling("case 129", "z", "symlinked change directory (in-tree target) + violating sibling (violation sorts last): exit 3",
			"symlinked change directory (in-tree target) + violation (violation last): sibling violation reported", "", "", ppAliasedChange),
		ppContainmentSibling("case 130", "a", "non-regular tasks.md (FIFO) + violating sibling (violation sorts first): exit 3, scan continues",
			"FIFO tasks.md + violation (violation first): sibling violation reported",
			"FIFO tasks.md + violation (violation first): containment message reported", "not a regular file after resolution", ppFifoChange),
		ppContainmentSibling("case 131", "z", "non-regular tasks.md (FIFO) + violating sibling (violation sorts last): exit 3, scan continues",
			"FIFO tasks.md + violation (violation last): sibling violation reported", "", "", ppFifoChange),
	)
}

// ppMSymlinked is the "m-symlinked-tasks-change" containment fixture cases
// 132-132b share.
func ppMSymlinked(t *testing.T, root string) {
	ppSymlink(t, "/dev/zero", filepath.Join(root, "spectre/changes/m-symlinked-tasks-change/tasks.md"))
}

const ppAbortBanner = "fence-like run of backticks/tildes behind a prefix"

// ppCasesLadderBanners runs case 132 through the pass-13 blockquote-tab
// cases 137a-138.
func ppCasesLadderBanners() []ppCase {
	return []ppCase{
		{"case 132", func(t *testing.T) {
			root := ppBare(t)
			ppAbortChange(t, root)
			ppMSymlinked(t, root)
			ppViolatingSibling(t, root, "a")
			r := ppRun(root)
			ok(t, "containment + abort + violation, all three directories: exit 3 wins", r.rc == 3, r)
			ok(t, "containment + abort + violation: violation still reported", has(r.out, "numeric claim with no measured"), r)
			ok(t, "containment + abort + violation: abort still reported", has(r.out, ppAbortBanner), r)
			ok(t, "containment + abort + violation: containment message still reported", has(r.out, "tasks.md is a symlink"), r)
		}},
		// SECTION: Minor 8 (pass-12 fix wave) — the 3-over-2 and 3-over-4
		// ladder edges, with no violation anywhere.
		{"case 132a", func(t *testing.T) {
			root := ppBare(t)
			p := filepath.Join(root, "spectre/changes/unreadable-change/tasks.md")
			ppWrite(t, root, "spectre/changes/unreadable-change/tasks.md", "no claims, no fences\n")
			ppChmod(t, p, 0)
			ppMSymlinked(t, root)
			r := ppRun(root)
			ppChmod(t, p, 0o644)
			ok(t, "containment + environment-failure-only sibling: exit 3 wins", r.rc == 3, r)
			ok(t, "containment + environment-only: environment failure still reported", has(r.out, "cannot read file"), r)
			ok(t, "containment + environment-only: containment message still reported", has(r.out, "tasks.md is a symlink"), r)
		}},
		{"case 132b", func(t *testing.T) {
			root := ppBare(t)
			ppAbortChange(t, root)
			ppMSymlinked(t, root)
			r := ppRun(root)
			ok(t, "containment + abort-only sibling: exit 3 wins", r.rc == 3, r)
			ok(t, "containment + abort-only: abort still reported", has(r.out, ppAbortBanner), r)
			ok(t, "containment + abort-only: containment message still reported", has(r.out, "tasks.md is a symlink"), r)
		}},
		// SECTION: Important 4 (pass-12 fix wave) — the EXACT wording of all
		// four summary banners, each produced in isolation.
		{"case 133", func(t *testing.T) {
			r := ppTasks(t, "intro\n```bash\necho hi\n```\n")
			ok(t, "exit 1 banner fixture: rc=1", r.rc == 1, r)
			ok(t, "exit 1 banner: exact wording", has(r.out, "1 plan-provenance violation(s) found."), r)
			ok(t, "exit 1 remedy line: exact wording", has(r.out, "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment — or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."), r)
		}},
		{"case 134", func(t *testing.T) {
			r := ppTasks(t, "- - ```bash\necho hi\n```\n")
			ok(t, "exit 4 banner fixture: rc=4", r.rc == 4, r)
			ok(t, "exit 4 banner: exact wording", has(r.out, "1 line(s) could not be classified (content-classification, see above) — refusing to report a clean run"), r)
		}},
		{"case 135", func(t *testing.T) {
			root := ppFixture(t)
			p := filepath.Join(root, ppDemo+"tasks.md")
			ppWrite(t, root, ppDemo+"tasks.md", "no claims, no fences here\n")
			ppChmod(t, p, 0)
			r := ppRun(root)
			ppChmod(t, p, 0o644)
			ok(t, "exit 2 banner fixture: rc=2", r.rc == 2, r)
			ok(t, "exit 2 banner: exact wording, distinct from exit 4's 'could not be classified'", has(r.out, "1 file(s) could not be read (environment failure — a permission, encoding or size problem on disk, see above) — refusing to report a clean run"), r)
			ok(t, "exit 2 banner: does not reuse exit 4's phrasing", !has(r.out, "could not be classified"), r)
		}},
		// 136: the noun is "path(s)" since pass 14's Critical 5 — the counter
		// also carries spectre/changes ITSELF, a directory.
		{"case 136", func(t *testing.T) {
			root := ppFixture(t)
			ppSymlink(t, "/dev/zero", filepath.Join(root, ppDemo+"tasks.md"))
			r := ppRun(root)
			ok(t, "exit 3 banner fixture: rc=3", r.rc == 3, r)
			ok(t, "exit 3 banner: exact wording", has(r.out, "1 path(s) failed containment (symlinked, escaping their change directory, or non-regular — see above) — refusing to report anything but a security failure. Every violation, abort or file error listed above from OTHER directories was still found and still needs fixing; fix the containment escape(s) first."), r)
		}},
		// SECTION: Pass-13 fix wave — Critical 1: a blockquote marker
		// consumes at most ONE COLUMN of the whitespace after its `>`
		// (CommonMark block quotes, and tabs spec example 6).
		// 137a: '>' col 0; the tab spans 1-3, the marker takes col 1; 2 + 3
		// spaces = 5 columns inside the blockquote — indented code, so line 1
		// opens no fence and line 3's untagged fence is reported.
		{"case 137a", func(t *testing.T) {
			r := ppTasks(t, ">\t   ```text verified:x\nhello\n```bash\necho hi\n```\n> ```\n")
			ok(t, "tab after '>' + trailing close: exit 1, no longer a silent clean run", r.rc == 1, r)
			ok(t, "tab after '>' + trailing close: the swallowed untagged fence on line 3 is reported", has(r.out, "tasks.md:3: "+ppNoTag), r)
			ok(t, "tab after '>' + trailing close: no false clean run", !has(r.out, "all provenance stated"), r)
		}},
		// 137b: the tab replaced by the columns it is worth ('>' + 6 spaces).
		{"case 137b", func(t *testing.T) {
			r := ppTasks(t, ">      ```text verified:x\nhello\n```bash\necho hi\n```\n> ```\n")
			ok(t, "6 spaces after '>' (same column width as the tab shape): exit 1", r.rc == 1, r)
			ok(t, "6 spaces after '>': the untagged fence on line 3 is reported, exactly as in the tab shape", has(r.out, "tasks.md:3: "+ppNoTag), r)
		}},
		{"case 137c", func(t *testing.T) {
			r := ppTasks(t, ">\t   ```text verified:x\nhello\n```bash\necho hi\n```\n")
			ok(t, "tab after '>' without a trailing close: exit 1", r.rc == 1, r)
			ok(t, "tab after '>' without a trailing close: line 3's untagged fence is reported, not line 1's phantom fence", has(r.out, "tasks.md:3: "+ppNoTag), r)
		}},
		{"case 138", func(t *testing.T) {
			r := ppTasks(t, ">\t   ```bash\n")
			ok(t, "no-list-marker false positive ('>' + tab + 3 spaces + fence run): not a fence, exit 0", r.rc == 0, r)
			ok(t, "no-list-marker false positive: no phantom fence opened", !has(r.out, "never closed"), r)
		}},
	}
}

// ppBqTabListCloser is the harness's bq_tab_list_closer_case: '>' col 0,
// the tab spans 1-3 and the marker takes col 1, so the bullet sits at col 4
// and the item's content column is 6; the closer's run sits at 2 + n, so n
// 4-7 close it and 3 (short) or 8 (past the budget) do not.
func ppBqTabListCloser(n, expect int, label string) ppCase {
	return ppCase{fmt.Sprintf("case 139 n=%d", n), func(t *testing.T) {
		r := ppTasks(t, ">\t- ```text verified:x\n> hello\n> "+strings.Repeat(" ", n)+"```\n")
		ok(t, fmt.Sprintf("%s: closer at bq-relative column %d: exit %d", label, n, expect), r.rc == expect, r)
	}}
}

// ppBq2TabListCloser is the harness's bq2_tab_list_closer_case: content
// column 10 at depth 2; the closer's run sits at 8 + n, so n 2-5 close it
// and 1 does not.
func ppBq2TabListCloser(n, expect int, label string) ppCase {
	return ppCase{fmt.Sprintf("case 142 n=%d", n), func(t *testing.T) {
		r := ppTasks(t, ">\t>\t- ```text verified:x\n>\t> hello\n>\t>\t"+strings.Repeat(" ", n)+"```\n")
		ok(t, fmt.Sprintf("%s: closer with %d extra spaces: exit %d", label, n, expect), r.rc == expect, r)
	}}
}

// ppCasesPass13 runs cases 139-145.
func ppCasesPass13() []ppCase {
	const l139, l142 = "tab after '>' then a bullet (depth 1)", "tabs after both '>'s then a bullet (depth 2)"
	return []ppCase{
		ppBqTabListCloser(4, 0, l139), ppBqTabListCloser(5, 0, l139), ppBqTabListCloser(6, 0, l139),
		ppBqTabListCloser(7, 0, l139), ppBqTabListCloser(3, 1, l139), ppBqTabListCloser(8, 1, l139),
		{"case 140", func(t *testing.T) {
			r := ppTasks(t, ">\t- ```text\n> hello\n>     ```\n")
			ok(t, "tab after '>' then a bullet, untagged: exit 1", r.rc == 1, r)
			ok(t, "tab after '>' then a bullet, untagged: opening line reported, not 'never closed'", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		// 141: depth 2 — quoted content begins at col 6, the run sits at col
		// 11, 5 columns in: indented code.
		{"case 141", func(t *testing.T) {
			r := ppTasks(t, ">\t>\t   ```bash\n")
			ok(t, "depth-2 tabs, run 5 columns inside the quote: not a fence, exit 0", r.rc == 0, r)
		}},
		{"case 141b-1", func(t *testing.T) {
			r := ppTasks(t, ">\t> ```text verified:x\n>\t> echo hi\n>\t> ```\n")
			ok(t, "depth-2 tabs, run at relative indent 0: real fence, opens and closes, exit 0", r.rc == 0, r)
		}},
		{"case 141b-2", func(t *testing.T) {
			r := ppTasks(t, ">\t> ```text\n>\t> echo hi\n>\t> ```\n")
			ok(t, "depth-2 tabs, run at relative indent 0, untagged: opening line reported", has(r.out, "tasks.md:1: "+ppNoTag), r)
		}},
		ppBq2TabListCloser(2, 0, l142), ppBq2TabListCloser(3, 0, l142), ppBq2TabListCloser(4, 0, l142),
		ppBq2TabListCloser(5, 0, l142), ppBq2TabListCloser(1, 1, l142),
		// SECTION: Pass-13 fix wave — Important 2: the outer-indent cap on
		// the CLOSING side, open and close indents deliberately mismatched.
		{"case 143a", func(t *testing.T) {
			r := ppTasks(t, "> ```text verified:x\n> echo hi\n    > ```\n")
			ok(t, "open at outer indent 0, close at outer indent 4: not a close, fence reported as never closed", r.rc == 1, r)
			ok(t, "open outer 0 / close outer 4: reported as never closed, not accepted silently", has(r.out, "never closed"), r)
		}},
		{"case 143b", func(t *testing.T) {
			r := ppTasks(t, "> ```text verified:x\n> echo hi\n   > ```\n")
			ok(t, "open at outer indent 0, close at outer indent 3: still a valid close, exit 0", r.rc == 0, r)
		}},
		// 143c: CommonMark budgets 0-3 outer columns per line independently.
		{"case 143c", func(t *testing.T) {
			r := ppTasks(t, "   > ```text verified:x\n   > echo hi\n> ```\n")
			ok(t, "open at outer indent 3, close at outer indent 0: still a valid close, exit 0", r.rc == 0, r)
		}},
		{"case 143d", func(t *testing.T) {
			r := ppTasks(t, "   > ```text verified:x\n   > echo hi\n    > ```\n")
			ok(t, "open at outer indent 3, close at outer indent 4: not a close", r.rc == 1, r)
		}},
		// SECTION: Pass-13 fix wave — Minor 6: the abort message for a line
		// that HAS a container marker but exceeds the outer-indent cap.
		{"case 144", func(t *testing.T) {
			r := ppTasks(t, "\t- \t```bash\n")
			ok(t, "over-indented line that DOES carry a marker: exit 4", r.rc == 4, r)
			ppOverIndentedMessage(t, "over-indented line that DOES carry a marker", r)
			ok(t, "over-indented line with a marker: remedy does not ask for a marker that is already present", !has(r.out, "adding an explicit list/blockquote marker on this line"), r)
		}},
		{"case 145", func(t *testing.T) {
			r := ppTasks(t, "\t```bash\n")
			ok(t, "bare over-indented line: exit 4", r.rc == 4, r)
			ok(t, "bare over-indented line: keeps the bare-case message", has(r.out, "with no container prefix on this line"), r)
		}},
	}
}

// ppCasesPass14 runs cases 146-183.
func ppCasesPass14() []ppCase {
	var cases []ppCase
	// SECTION: Pass-14 fix wave — Critical 1: whitespace runs bounded WHERE
	// THEY ARE USED. List items: rule 1 (1 <= W <= 4) puts the content
	// column W past the marker; rule 2 (W >= 5) puts it ONE past, the rest
	// indented code. Block quotes: 0-3 columns of indent before a `>`.
	// 146-151: a bullet + W spaces + a TAGGED fence, a claim next line.
	for w := 1; w <= 6; w++ {
		cases = append(cases, ppCase{fmt.Sprintf("case 146-151 w=%d", w), func(t *testing.T) {
			r := ppTasks(t, "-"+strings.Repeat(" ", w)+"```bash verified:x\ntook 5 ms\n"+strings.Repeat(" ", w+1)+"```\n")
			if w <= 4 {
				ok(t, fmt.Sprintf("list marker + %d space(s): rule 1, fence opens, claim is content", w), r.rc == 0, r)
				return
			}
			ok(t, fmt.Sprintf("list marker + %d space(s): rule 2, indented code, claim reported", w), r.rc == 1, r)
			ok(t, fmt.Sprintf("list marker + %d space(s): the claim is named at line 2", w), has(r.out, "tasks.md:2: numeric claim"), r)
		}})
	}
	cases = append(cases, []ppCase{
		{"case 152", func(t *testing.T) {
			r := ppTasks(t, "took 5 ms\n")
			ok(t, "claim-alone control: reported", r.rc == 1, r)
		}},
		// 153: the tab runs cols 1-4, W=3 — rule 1.
		{"case 153", func(t *testing.T) {
			r := ppTasks(t, "-\t```bash verified:x\ntook 5 ms\n\t```\n")
			ok(t, "list marker + one tab (W=3): rule 1, fence opens", r.rc == 0, r)
		}},
		// 154: CommonMark Tabs example 7 — W=7, rule 2.
		{"case 154", func(t *testing.T) {
			r := ppTasks(t, "-\t\t```bash verified:x\ntook 5 ms\n\t\t```\n")
			ok(t, "list marker + two tabs (W=7): rule 2 (spec Tabs ex. 7), claim reported", r.rc == 1, r)
		}},
	}...)
	// 155-160: the same bound on a checkbox's trailing whitespace.
	for w := 1; w <= 6; w++ {
		cases = append(cases, ppCase{fmt.Sprintf("case 155-160 w=%d", w), func(t *testing.T) {
			r := ppTasks(t, "- [ ]"+strings.Repeat(" ", w)+"```bash verified:x\ntook 5 ms\n"+strings.Repeat(" ", w+5)+"```\n")
			if w <= 4 {
				ok(t, fmt.Sprintf("checkbox + %d space(s): rule 1, fence opens", w), r.rc == 0, r)
				return
			}
			ok(t, fmt.Sprintf("checkbox + %d space(s): rule 2, claim reported", w), r.rc == 1, r)
		}})
	}
	// 161-166: `>` + N spaces + `>`: the first marker credits one column, so
	// N-1 columns sit before the second. N <= 4 is depth 2; N = 5 ends the
	// run at depth 1 and the guard refuses.
	for n := 0; n <= 5; n++ {
		cases = append(cases, ppCase{fmt.Sprintf("case 161-166 n=%d", n), func(t *testing.T) {
			pre := ">" + strings.Repeat(" ", n) + "> ```"
			r := ppTasks(t, pre+"bash\nx\n"+pre+"\n")
			if n <= 4 {
				ok(t, fmt.Sprintf("bq gap %d (<=3 cols before the nested >): depth 2, untagged fence reported", n), r.rc == 1, r)
				return
			}
			ok(t, fmt.Sprintf("bq gap %d (4+ cols): run ends at depth 1, guard refuses", n), r.rc == 4, r)
		}})
	}
	// SECTION: Pass-14 fix wave — Criticals 2 and 3: a fence carrying BOTH
	// bq_pre and bq_post. `> - > ```` puts the run at column 6; a closer
	// `>` + N spaces + `> ```` is legal for N = 1 + 2 + (0..3) = 3..6.
	for _, n := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 10, 15, 20} {
		cases = append(cases, ppCase{fmt.Sprintf("case 167-178 n=%d", n), func(t *testing.T) {
			r := ppTasks(t, "> - > ```text verified:manual\ninside\n>"+strings.Repeat(" ", n)+"> ```\n")
			if n >= 3 && n <= 6 {
				ok(t, fmt.Sprintf("bq_pre+bq_post closer at N=%d: legal, fence closes", n), r.rc == 0, r)
				return
			}
			ok(t, fmt.Sprintf("bq_pre+bq_post closer at N=%d: illegal, fence stays open", n), r.rc == 1, r)
			ok(t, fmt.Sprintf("bq_pre+bq_post closer at N=%d: named as unclosed", n), has(r.out, "never closed"), r)
		}})
	}
	return append(cases, []ppCase{
		{"case 179", func(t *testing.T) {
			r := ppTasks(t, "> - > ```text verified:manual\ninside\n>   > ```\n```untagged\nx\n```\ntook 7 ms\n")
			ok(t, "bq_pre+bq_post: downstream findings survive the fence", r.rc == 1, r)
			ok(t, "bq_pre+bq_post: the later untagged fence is reported", has(r.out, "tasks.md:4: fenced code block has no verified"), r)
			ok(t, "bq_pre+bq_post: the later claim is reported", has(r.out, "tasks.md:7: numeric claim"), r)
		}},
		// 180: the gap before the nested `>` is budgeted 0-3 on both sides.
		{"case 180", func(t *testing.T) {
			r := ppTasks(t, "> -     > ```bash\n")
			ok(t, "bq_post gap 4+ as an OPENER: refused, same budget as the closer", r.rc == 4, r)
		}},
		// SECTION: Pass-14 fix wave — Critical 4: the wrapper's interpreter
		// gate. Cases 181-183 pinned the Python wrapper's python3 probe; the
		// shim's equivalent gate is the go toolchain it builds flow-guard
		// with, whose every failure is exit 2 (flow_guard_exec), never 1.
		{"case 181", func(t *testing.T) {
			bin := t.TempDir()
			writeExec(t, bin+"/go", "#!/bin/sh\necho \"xcrun: error: invalid active developer path\" >&2\nexit 1\n")
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", "clean\n")
			r := ppShim(t, "PATH="+bin+":/usr/bin:/bin", "FLOW_GUARD_CACHE_DIR="+t.TempDir(), "CHECK_PLAN_PROVENANCE_ROOT="+root)
			ok(t, "broken go: exit 2, not exit 1", r.rc == 2, r)
			ok(t, "broken go: names what actually happened", has(r.out, "check-plan-provenance: could not build flow-guard from"), r)
			ok(t, "broken go: the toolchain's own stderr is passed through", has(r.out, "invalid active developer path"), r)
		}},
		{"case 182", func(t *testing.T) {
			r := ppShim(t, "PATH=/usr/bin:/bin", "FLOW_GUARD_CACHE_DIR="+t.TempDir(), "CHECK_PLAN_PROVENANCE_ROOT="+ppFixture(t))
			ok(t, "absent go: still exit 2", r.rc == 2, r)
			ok(t, "absent go: keeps its own distinct message", has(r.out, "no go on PATH"), r)
		}},
		// 183: a working toolchain is unaffected. Run through the real shim,
		// which builds flow-guard from this checkout.
		{"case 183", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"tasks.md", "clean\n")
			r := ppShim(t, "PATH="+os.Getenv("PATH"), "HOME="+os.Getenv("HOME"), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "CHECK_PLAN_PROVENANCE_ROOT="+root)
			ok(t, "working go: unchanged, exit 0", r.rc == 0, r)
		}},
	}...)
}

// ppCasesAnchorQuotes runs cases 184-199.
func ppCasesAnchorQuotes() []ppCase {
	untagged := "```untagged\nx\n```\n"
	return []ppCase{
		// SECTION: Pass-14 fix wave — Critical 5: `spectre/changes` may itself
		// be a symlink; the anchor is validated before anything is anchored
		// on it.
		{"case 184", func(t *testing.T) {
			root := t.TempDir()
			mkdir(t, root+"/spectre")
			mkdir(t, root+"/decoy")
			ppWrite(t, root, "real/plan-change/tasks.md", untagged)
			ppSymlink(t, root+"/decoy", root+"/spectre/changes")
			r := ppRun(root)
			ok(t, "changes/ symlinked to a decoy: exit 3, not a clean run", r.rc == 3, r)
			ok(t, "changes/ decoy symlink: names the anchor", has(r.out, "spectre/changes is a symlink"), r)
		}},
		{"case 185", func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			mkdir(t, root+"/spectre")
			ppWrite(t, outside, "evil/tasks.md", untagged)
			ppSymlink(t, outside, root+"/spectre/changes")
			r := ppRun(root)
			ok(t, "changes/ symlinked outside the root: exit 3", r.rc == 3, r)
		}},
		{"case 186", func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			mkdir(t, root+"/spectre/changes")
			ppWrite(t, outside, "evil/tasks.md", untagged)
			ppSymlink(t, outside+"/evil", root+"/spectre/changes/evil")
			r := ppRun(root)
			ok(t, "changes/<name> symlink control: still exit 3", r.rc == 3, r)
			ok(t, "changes/<name> symlink control: keeps its own per-file message", has(r.out, "refusing to scan a path reached through a symlink escape"), r)
		}},
		{"case 187", func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			ppWrite(t, outside, "changes/demo-change/tasks.md", "clean\n")
			ppSymlink(t, outside, root+"/spectre")
			r := ppRun(root)
			ok(t, "spectre/ itself symlinked: exit 3 via the realpath check", r.rc == 3, r)
		}},
		{"case 188", func(t *testing.T) {
			r := ppTasks(t, "clean\n")
			ok(t, "real changes/ directory: anchor check does not fire", r.rc == 0, r)
		}},
		// SECTION: Claim boundaries (cont'd) — an issue key is not a quantity.
		{"case 189", func(t *testing.T) {
			r := ppTasks(t, "catches the KAN-6 errors in one pass\n")
			ok(t, "issue key followed by a unit word is not a claim", r.rc == 0, r)
		}},
		{"case 190", func(t *testing.T) {
			r := ppTasks(t, "the suite reported 6 errors\n")
			ok(t, "a bare digit run before a unit word is still a claim", r.rc == 1, r)
		}},
		// SECTION: Claim boundaries (cont'd) — a quoted number is reproduced,
		// not claimed.
		{"case 191", func(t *testing.T) {
			r := ppTasks(t, "a baseline of \"194 tests\" was invented, not measured\n")
			ok(t, "straight-quoted number is not a claim", r.rc == 0, r)
		}},
		// 192: a code span is NOT a delimiter class (pass 17,
		// quotation-quotes-only); this case is that removal's specification.
		{"case 192", func(t *testing.T) {
			r := ppTasks(t, "the offending line read `85 lines` with no tag\n")
			ok(t, "a number in a code span is an ordinary claim, not a quotation", r.rc == 1, r)
			ok(t, "code span: an unquoted claim carries no veto note", !has(r.out, ": note: the quotation exemption was withdrawn"), r)
		}},
		{"case 192b", func(t *testing.T) {
			r := ppTasks(t, "the log read ``a ` tick and 7 errors`` verbatim\n")
			ok(t, "a number in a multi-backtick span is an ordinary claim too", r.rc == 1, r)
		}},
		{"case 193", func(t *testing.T) {
			r := ppTasks(t, "quoting \342\200\234194 tests\342\200\235 from the earlier plan\n")
			ok(t, "curly-quoted number is not a claim", r.rc == 0, r)
		}},
		{"case 194", func(t *testing.T) {
			r := ppTasks(t, "it ran \"and then reported 12 failures\n")
			ok(t, "unmatched delimiter does not exempt", r.rc == 1, r)
		}},
		{"case 195", func(t *testing.T) {
			r := ppTasks(t, "the guard's own suite reported 12 failures\n")
			ok(t, "apostrophe is not a quotation delimiter", r.rc == 1, r)
		}},
		{"case 196", func(t *testing.T) {
			r := ppTasks(t, "we quoted \"194 tests\" but then asserted 12 failures\n")
			ok(t, "a bare claim beside a quoted one is still reported", r.rc == 1, r)
		}},
		// SECTION: Claim boundaries (cont'd) — the quotation exemption fails
		// CLOSED: a class whose delimiters do not balance yields no region.
		{"case 197", func(t *testing.T) {
			r := ppTasks(t, "a \"quote\" and a stray \" mark, then 99 tests ran, \"done\"\n")
			ok(t, "stray quote plus a later pair does not exempt", r.rc == 1, r)
			ok(t, "class-wide veto names itself in the output", has(r.out, "quotation exemption was withdrawn on this line (unbalanced quotation delimiters)"), r)
			ok(t, "class-wide veto names the remedy and the contract", has(r.out, "balance them or reword the line", "skills/flow-contracts/plan-provenance-guard.md"), r)
		}},
		{"case 198", func(t *testing.T) {
			r := ppTasks(t, "a `span` and a stray ` then 50 files were read `ok`\n")
			ok(t, "a backtick run does not create a quotation", r.rc == 1, r)
			ok(t, "a backtick run vetoes nothing", !has(r.out, ": note: the quotation exemption was withdrawn"), r)
		}},
		{"case 199", func(t *testing.T) {
			r := ppTasks(t, "a matched \"194 tests\" pair, then a dangling \"\n")
			ok(t, "an opener unclosed at end of line exempts nothing", r.rc == 1, r)
		}},
	}
}

// ppCasesScopeToEnd runs cases 200-238.
func ppCasesScopeToEnd() []ppCase {
	withClean := func(t *testing.T) string {
		root := ppFixture(t)
		ppWrite(t, root, ppDemo+"tasks.md", ppClean)
		return root
	}
	return []ppCase{
		{"case 200", func(t *testing.T) {
			r := ppTasks(t, "the log read \"a tick\" then \"7 errors\" then \"done\" verbatim\n")
			ok(t, "an enclosed claim between unrelated pairs stays exempt", r.rc == 0, r)
		}},
		// SECTION: Scan scope — design.md and proposal.md alongside tasks.md.
		{"case 201", func(t *testing.T) {
			root := withClean(t)
			ppWrite(t, root, ppDemo+"design.md", "the suite reported 197 tests\n")
			r := ppRun(root)
			ok(t, "design.md numeric claim is in scope", r.rc == 1, r)
		}},
		// 202: specs/ legislates rather than describes.
		{"case 202", func(t *testing.T) {
			root := withClean(t)
			ppWrite(t, root, ppDemo+"specs/some-capability/spec.md", "a baseline of 197 tests\n")
			r := ppRun(root)
			ok(t, "change specs/ is not scanned", r.rc == 0, r)
		}},
		{"case 203", func(t *testing.T) {
			root := withClean(t)
			ppWrite(t, root, "spectre/changes/archive/old-change/design.md", "the suite reported 197 tests\n")
			r := ppRun(root)
			ok(t, "archived design.md is not scanned", r.rc == 0, r)
		}},
		{"case 204", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, ppDemo+"design.md", ppClean)
			r := ppRun(root)
			ok(t, "design.md alone is a scanned change, not a broken glob", r.rc == 0, r)
		}},
		{"case 205", func(t *testing.T) {
			root := withClean(t)
			ppSymlink(t, "/etc/hosts", filepath.Join(root, ppDemo+"design.md"))
			r := ppRun(root)
			ok(t, "a symlinked design.md is a containment refusal", r.rc == 3, r)
		}},
		{"case 206", func(t *testing.T) {
			root := withClean(t)
			ppWrite(t, root, ppDemo+"design.md", "ran the suite, 3 tests <!-- measured: pytest -q -->\n")
			ppWrite(t, root, ppDemo+"proposal.md", "ran the suite, 3 tests <!-- measured: pytest -q -->\n")
			r := ppRun(root)
			ok(t, "every scanned planning artifact is counted", ppOut(r) == "check-plan-provenance: 3 file(s) scanned, all provenance stated", r)
		}},
		// 207a: a containment refusal skips exactly ONE CANDIDATE — the
		// refused tasks.md sorts before its violating sibling design.md. The
		// symlink target is fixture-local and the preconditions are asserted
		// separately (pass-15 fix wave, Important 4), so a broken fixture is
		// never blamed on the guard.
		{"case 207a", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, "symlink-target.txt", "not a planning artifact\n")
			ppSymlink(t, filepath.Join(root, "symlink-target.txt"), filepath.Join(root, ppDemo+"tasks.md"))
			ppWrite(t, root, ppDemo+"design.md", "the suite reported 197 tests\n")
			lst, lerr := os.Lstat(filepath.Join(root, ppDemo+"tasks.md"))
			st, serr := os.Stat(filepath.Join(root, ppDemo+"design.md"))
			ok(t, "sibling candidate: fixture built (symlinked tasks.md, non-empty design.md)",
				lerr == nil && lst.Mode()&os.ModeSymlink != 0 && serr == nil && st.Size() > 0, tcfRes{})
			r := ppRun(root)
			ok(t, "sibling candidate: rc=3 (containment outranks the violation)", r.rc == 3, r)
			ok(t, "sibling candidate: the violating design.md was still scanned and reported", has(r.out, "spectre/changes/demo-change/design.md:1: "+ppNoProv), r)
		}},
		{"case 207b", func(t *testing.T) {
			root := ppFixture(t)
			ppWrite(t, root, "symlink-target.txt", "not a planning artifact\n")
			ppWrite(t, root, ppDemo+"tasks.md", "the suite reported 197 tests\n")
			ppSymlink(t, filepath.Join(root, "symlink-target.txt"), filepath.Join(root, ppDemo+"design.md"))
			ppWrite(t, root, ppDemo+"proposal.md", "ran the suite, 3 tests <!-- measured: pytest -q -->\n")
			lst, lerr := os.Lstat(filepath.Join(root, ppDemo+"design.md"))
			st, serr := os.Stat(filepath.Join(root, ppDemo+"tasks.md"))
			ok(t, "sibling candidate (mirror): fixture built",
				lerr == nil && lst.Mode()&os.ModeSymlink != 0 && serr == nil && st.Size() > 0, tcfRes{})
			r := ppRun(root)
			ok(t, "sibling candidate (mirror): rc=3", r.rc == 3, r)
			ok(t, "sibling candidate (mirror): the violating tasks.md was still reported", has(r.out, "spectre/changes/demo-change/tasks.md:1: "+ppNoProv), r)
		}},
		// 208: containment confirms the BASENAME. No CLI fixture can reach a
		// mismatch (every candidate is built from ppScanned), so the check is
		// called directly.
		{"case 208", func(t *testing.T) {
			root := withClean(t)
			changes := filepath.Join(root, "spectre", "changes")
			var stderr bytes.Buffer
			g := ppGuard{stdout: &stderr, stderr: &stderr, dir: os.TempDir()}
			got := g.verifyContainment(filepath.Join(changes, "demo-change", "tasks.md"), changes, "design.md")
			ok(t, "containment confirms the basename, not only the directory", !got, tcfRes{0, stderr.String()})
		}},
		// SECTION: The quotation exemption is BLIND to backticks (209-212): a
		// backtick neither exempts (209/210) nor vetoes (211/212).
		{"case 209", func(t *testing.T) {
			r := ppTasks(t, "`a``b` and then we saw 99 tests happen ``\n")
			ok(t, "a backtick run neither exempts nor vetoes (leftover run)", r.rc == 1, r)
		}},
		{"case 210", func(t *testing.T) {
			r := ppTasks(t, "the log said ``a ` b`` and 42 files changed ` done\n")
			ok(t, "a backtick run neither exempts nor vetoes (trailing run)", r.rc == 1, r)
		}},
		// 211: inverted by pass 17 deliberately — the quotes pair and enclose.
		{"case 211", func(t *testing.T) {
			r := ppTasks(t, "## Task\n\n`a` \"b` and 7 tests failed\" c\n")
			ok(t, "a matched quotation is read without consulting backticks", r.rc == 0, r)
		}},
		// 212: the accepted cost of quotation-quotes-only, as an assertion.
		{"case 212", func(t *testing.T) {
			r := ppTasks(t, "the tool printed `\342\200\234` before 8 failures and \342\200\235 after, plus ` more\n")
			ok(t, "a curly pair spanning a backtick run exempts (the accepted cost)", r.rc == 0, r)
		}},
		// 213: one "cannot determine whether <filename> exists" per scanned
		// filename, each naming a different artifact.
		{"case 213", func(t *testing.T) {
			root := ppFixture(t)
			locked := filepath.Join(root, "spectre/changes/locked-change")
			mkdir(t, locked)
			ppChmod(t, locked, 0)
			ppWrite(t, root, ppDemo+"tasks.md", ppClean)
			r := ppRun(root)
			ppChmod(t, locked, 0o755)
			var names []string
			for _, line := range strings.Split(ppOut(r), "\n") {
				i := strings.Index(line, "cannot determine whether ")
				if i < 0 {
					continue
				}
				rest := line[i+len("cannot determine whether "):]
				if j := strings.Index(rest, " exists"); j >= 0 {
					names = append(names, rest[:j])
				}
			}
			ok(t, "a locked change dir emits one 'cannot determine' per scanned filename", len(names) == 3, r)
			sort.Strings(names)
			ok(t, "each 'cannot determine' names a distinct, correct planning artifact", strings.Join(names, " ") == "design.md proposal.md tasks.md", r)
		}},
		// SECTION: The quotation exemption, BACKSLASH ESCAPES (214-221):
		// CommonMark §2.4 makes \" a literal quote.
		{"case 214", func(t *testing.T) {
			r := ppTasks(t, "Escape it as \\` in the shell, and `grep -c` reported 42 files after the \\` fix\n")
			ok(t, "an escaped backtick is an ordinary character now", r.rc == 1, r)
			ok(t, "an escaped backtick vetoes nothing", !has(r.out, ": note: the quotation exemption was withdrawn"), r)
		}},
		{"case 215", func(t *testing.T) {
			r := ppTasks(t, "Escape it as X in the shell, and `grep -c` reported 42 files after the X fix\n")
			ok(t, "the unescaped control for 214 reports identically", r.rc == 1, r)
		}},
		{"case 216", func(t *testing.T) {
			r := ppTasks(t, "Write \\` to escape a backtick; the run then reported 99 tests and exited \\` cleanly\n")
			ok(t, "a pair of escaped backticks exempts nothing", r.rc == 1, r)
		}},
		{"case 217", func(t *testing.T) {
			r := ppTasks(t, "Use \\\" for a literal quote; the run reported 99 tests, then printed \\\" done\n")
			ok(t, "a pair of escaped straight quotes is not a quotation", r.rc == 1, r)
			ok(t, "escape veto names itself in the output", has(r.out, "quotation exemption was withdrawn on this line (a backslash-escaped quotation delimiter)"), r)
			ok(t, "escape veto names the remedy and the contract", has(r.out, "unescape it or reword the line", "skills/flow-contracts/plan-provenance-guard.md"), r)
		}},
		{"case 218", func(t *testing.T) {
			r := ppTasks(t, "the suite ran \\`99 tests\\` today\n")
			ok(t, "escaped backticks hugging the claim do not exempt it", r.rc == 1, r)
		}},
		// 219: \\ is an escaped BACKSLASH, so the quote after it is live —
		// parity, not adjacency.
		{"case 219", func(t *testing.T) {
			r := ppTasks(t, "notes: \\\\\"99 tests\\\\\" recorded\n")
			ok(t, "an escaped backslash leaves the following quote live", r.rc == 0, r)
		}},
		{"case 220", func(t *testing.T) {
			r := ppTasks(t, "notes: \\\\\\\"99 tests\\\\\\\" recorded\n")
			ok(t, "an odd backslash run escapes the delimiter after it", r.rc == 1, r)
		}},
		{"case 221", func(t *testing.T) {
			r := ppTasks(t, "the baseline was \"99 tests\", and a literal quote is written \\\"\n")
			ok(t, "an escaped quote withdraws a sound quotation elsewhere on the line", r.rc == 1, r)
		}},
		// SECTION: The quotation exemption and the `<` character (222-228):
		// a `<` anywhere on the line withdraws the exemption.
		{"case 222", func(t *testing.T) {
			r := ppTasks(t, "See <!--\"--> the benchmark ran 77 tests <!--\"-->.\n")
			ok(t, "quotes inside an HTML comment are not live delimiters", r.rc == 1, r)
			ok(t, "angle-bracket veto names itself in the output", has(r.out, "quotation exemption was withdrawn on this line (a `<` character on the line)"), r)
			ok(t, "angle-bracket veto names the remedy and the contract", has(r.out, "remove the `<` or reword the line", "skills/flow-contracts/plan-provenance-guard.md"), r)
		}},
		{"case 223", func(t *testing.T) {
			r := ppTasks(t, "the baseline was \"99 tests\", per <a href=\"x\">the report</a>\n")
			ok(t, "an HTML tag on the line withdraws the whole line", r.rc == 1, r)
		}},
		{"case 224", func(t *testing.T) {
			r := ppTasks(t, "the baseline was \"194 tests\" <!-- see the earlier plan -->\n")
			ok(t, "a `<` with no delimiter in it still withdraws the exemption", r.rc == 1, r)
		}},
		// 225 is written with printf '%s\n', so its \342... stay literal
		// backslash text, exactly as the harness wrote them.
		{"case 225", func(t *testing.T) {
			r := ppTasks(t, `See <!--\342\200\234--> the benchmark ran 77 tests <!--\342\200\235-->.`+"\n")
			ok(t, "curly quotes inside an HTML comment are not live delimiters", r.rc == 1, r)
		}},
		{"case 226", func(t *testing.T) {
			r := ppTasks(t, "See \" the benchmark ran 77 tests \".\n")
			ok(t, "the angle-bracket-free control for 222 exempts the claim", r.rc == 0, r)
		}},
		{"case 227", func(t *testing.T) {
			r := ppTasks(t, "<a b='>\"'> the benchmark ran 77 tests \"\n")
			ok(t, "a `>` inside a single-quoted attribute cannot hide a delimiter", r.rc == 1, r)
		}},
		{"case 228", func(t *testing.T) {
			r := ppTasks(t, "the range a <b and c covers 5 tests\n")
			ok(t, "a bare `<` withdraws the exemption", r.rc == 1, r)
			ok(t, "the note on a delimiter-free line names the `<`, not a delimiter", has(r.out, "quotation exemption was withdrawn on this line (a `<` character on the line)"), r)
		}},
		// SECTION: The exit-1 remedy banner.
		{"case 229", func(t *testing.T) {
			r := ppTasks(t, "a \"quote\" and a stray \" mark, then 99 tests ran, \"done\"\n")
			ok(t, "banner case: the vetoed claim is reported", r.rc == 1, r)
			ok(t, "the exit-1 banner offers rewording as a remedy", has(r.out, "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment — or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."), r)
		}},
		// SECTION: ppRegionAt's precondition is enforced (case 230). No CLI
		// fixture reaches it — the sole producer builds the list correctly —
		// so the function is called directly.
		{"case 230", func(t *testing.T) {
			overlapping := []ppRegion{{0, 10}, {4, 20}}
			unsorted := []ppRegion{{10, 20}, {0, 5}}
			inverted := []ppRegion{{9, 3}}
			_, a := ppRegionAt(6, overlapping)
			_, b := ppRegionAt(2, unsorted)
			_, c := ppRegionAt(5, inverted)
			got := []bool{!a, !b, !c,
				ppRegionsSortedDisjoint(overlapping), ppRegionsSortedDisjoint(unsorted), ppRegionsSortedDisjoint(inverted),
				ppRegionsSortedDisjoint([]ppRegion{{0, 4}, {4, 9}, {9, 9}})}
			want := []bool{true, true, true, false, false, false, true}
			ok(t, "_region_at fails closed on a violated precondition, and the producer check agrees",
				fmt.Sprint(got) == fmt.Sprint(want), tcfRes{0, fmt.Sprint(got)})
		}},
		// SECTION: The evidence rule — a tag with an empty payload is a
		// violation (231-238).
		{"case 231", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:\necho hi\n```\n")
			ok(t, "empty-payload verified: fails", r.rc != 0, r)
			ok(t, "empty-payload verified: reports file:line and the evidence message", has(r.out, "spectre/changes/demo-change/tasks.md:1", "carries", "no evidence"), r)
		}},
		{"case 232", func(t *testing.T) {
			r := ppTasks(t, "```bash unverified:\necho hi\n```\n")
			ok(t, "empty-payload unverified: fails", r.rc != 0, r)
		}},
		{"case 233", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:   \necho hi\n```\n")
			ok(t, "whitespace-only payload fails", r.rc != 0, r)
		}},
		{"case 234", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- measured: -->\n")
			ok(t, "empty-payload measured: fails", r.rc != 0, r)
			ok(t, "empty measured: costs one finding, claim stays satisfied", !has(r.out, "numeric claim with no"), r)
		}},
		{"case 235", func(t *testing.T) {
			r := ppTasks(t, "After the deletion: 186 tests\n<!-- predicted: -->\n")
			ok(t, "empty-payload predicted: fails", r.rc != 0, r)
		}},
		{"case 236", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- measured:\n./gradlew test @ c515c42 -->\n")
			ok(t, "comment closing on a later line is not evidence-checked", r.rc == 0, r)
		}},
		{"case 237", func(t *testing.T) {
			r := ppTasks(t, "Baseline: 197 tests\n<!-- measured: wc -l on the machine-local log; no ref, the file is machine-local -->\n")
			ok(t, "payload without a ref passes", r.rc == 0, r)
		}},
		{"case 238", func(t *testing.T) {
			r := ppTasks(t, "```bash verified:ran it locally\necho \"<!-- measured: -->\"\n```\n")
			ok(t, "empty comment inside fence content is not flagged", r.rc == 0, r)
		}},
	}
}

// ppCasesShim is the shim's own contract past the harness: with
// CHECK_PLAN_PROVENANCE_ROOT unset it scans the checkout it sits in, from
// any working directory — the default the Python guard took from its own
// file's location. A copy of the shim and lib/flow-guard.sh in a temporary
// tree, with stats/ linked to this checkout's, is a checkout whose planning
// artifacts the case controls.
func ppCasesShim() []ppCase {
	return []ppCase{
		{"shim default root", func(t *testing.T) {
			tree := t.TempDir()
			scripts := tcfScriptsDir(t)
			for _, rel := range []string{"check-plan-provenance.sh", "lib/flow-guard.sh"} {
				body, err := os.ReadFile(filepath.Join(scripts, rel))
				if err != nil {
					t.Fatal(err)
				}
				ppWrite(t, tree, "scripts/"+rel, string(body))
			}
			stats, err := filepath.Abs("../..")
			if err != nil {
				t.Fatal(err)
			}
			ppSymlink(t, stats, filepath.Join(tree, "stats"))
			ppWrite(t, tree, ppDemo+"tasks.md", "Baseline: 197 tests\n")
			r := ppShimAt(t, filepath.Join(tree, "scripts/check-plan-provenance.sh"), t.TempDir(),
				"PATH="+os.Getenv("PATH"), "HOME="+os.Getenv("HOME"), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
			ok(t, "unset CHECK_PLAN_PROVENANCE_ROOT: the shim scans its own checkout, from any working directory",
				r.rc == 1 && has(r.out, "spectre/changes/demo-change/tasks.md:1: "+ppNoProv), r)
		}},
		{"root override unset or empty", func(t *testing.T) {
			r := ppRunEnv(nil)
			ok(t, "unset CHECK_PLAN_PROVENANCE_ROOT without the shim: exit 2, names the unset variable and the shim",
				r.rc == 2 && r.out == "CHECK_PLAN_PROVENANCE_ROOT is unset — run scripts/check-plan-provenance.sh, which sets it\n", r)
			r = ppRun("")
			ok(t, "empty CHECK_PLAN_PROVENANCE_ROOT: exit 2, set but empty",
				r.rc == 2 && r.out == "CHECK_PLAN_PROVENANCE_ROOT is set but empty\n", r)
		}},
	}
}

// ppCasesUnicodeMatchers pins the classes of every hand-written matcher that
// stands in for a Python pattern over Unicode \s or \w, or for a bounded
// repetition: CLAIM_RE's whitespace run, (?<!\w), (?<![A-Z]-), {0,20} and
// (?!\w); PROVENANCE_RE's \s*; FENCE_TAG_RE's (^|\s); the list, checkbox
// and blockquote container prefixes and closers; and the str.strip() sites.
// The whitespace rows use NBSP, U+3000 and U+001C/U+001F, which Python's \s
// matches and an ASCII-only class would not; each container-prefix and
// closer row was found by running a mutant with one call site's class made
// ASCII-only. Every row's exit code and output lines are the Python guard's
// at 3e48ecac, run over the same tasks.md.
func ppCasesUnicodeMatchers() []ppCase {
	rows := []struct {
		name, body string
		rc         int
		lines      []string
	}{
		{"claim: NBSP separates the number from its unit", "ran 5\u00a0tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"provenance comment: NBSP after <!-- satisfies", "ran 5 tests <!--\u00a0measured: x -->\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"fence tag: NBSP before the tag", "```bash\u00a0verified:ran it\necho hi\n```\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"list marker: NBSP after - opens a tagged fence", "-\u00a0```bash verified:x\n  echo hi\n  ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"checkbox marker: NBSP after [ ] opens a tagged fence", "- [ ]\u00a0```bash verified:x\n      echo hi\n      ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"ordered list marker: NBSP after 1. opens a tagged fence", "1.\u00a0```bash verified:x\n   echo hi\n   ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: NBSP after > opens a tagged fence", ">\u00a0```bash verified:x\n>\u00a0echo hi\n>\u00a0```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: NBSP indents a tagged fence", "\u00a0```bash verified:x\n\u00a0echo hi\n\u00a0```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"list item: NBSP after - carries a claim", "-\u00a0ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: NBSP after > carries a claim", ">\u00a0ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"closer tail: NBSP after the closing fence", "```bash verified:x\necho hi\n```\u00a0\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"comment payload: only NBSP is empty", "ran 5 tests <!-- measured:\u00a0-->\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: measured:/predicted: comment carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"tag payload: only NBSP is empty", "```bash verified:\u00a0\necho hi\n```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: verified:/unverified: tag on the info string carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: U+3000 separates the number from its unit", "ran 5\u3000tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"provenance comment: U+3000 after <!-- satisfies", "ran 5 tests <!--\u3000measured: x -->\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"fence tag: U+3000 before the tag", "```bash\u3000verified:ran it\necho hi\n```\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"list marker: U+3000 after - opens a tagged fence", "-\u3000```bash verified:x\n  echo hi\n  ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"checkbox marker: U+3000 after [ ] opens a tagged fence", "- [ ]\u3000```bash verified:x\n      echo hi\n      ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"ordered list marker: U+3000 after 1. opens a tagged fence", "1.\u3000```bash verified:x\n   echo hi\n   ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: U+3000 after > opens a tagged fence", ">\u3000```bash verified:x\n>\u3000echo hi\n>\u3000```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: U+3000 indents a tagged fence", "\u3000```bash verified:x\n\u3000echo hi\n\u3000```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"list item: U+3000 after - carries a claim", "-\u3000ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: U+3000 after > carries a claim", ">\u3000ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"closer tail: U+3000 after the closing fence", "```bash verified:x\necho hi\n```\u3000\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"comment payload: only U+3000 is empty", "ran 5 tests <!-- measured:\u3000-->\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: measured:/predicted: comment carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"tag payload: only U+3000 is empty", "```bash verified:\u3000\necho hi\n```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: verified:/unverified: tag on the info string carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: U+001C separates the number from its unit", "ran 5\u001ctests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"provenance comment: U+001C after <!-- satisfies", "ran 5 tests <!--\u001cmeasured: x -->\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"fence tag: U+001C before the tag", "```bash\u001cverified:ran it\necho hi\n```\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"list marker: U+001C after - opens a tagged fence", "-\u001c```bash verified:x\n  echo hi\n  ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"checkbox marker: U+001C after [ ] opens a tagged fence", "- [ ]\u001c```bash verified:x\n      echo hi\n      ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"ordered list marker: U+001C after 1. opens a tagged fence", "1.\u001c```bash verified:x\n   echo hi\n   ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: U+001C after > opens a tagged fence", ">\u001c```bash verified:x\n>\u001cecho hi\n>\u001c```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: U+001C indents a tagged fence", "\u001c```bash verified:x\n\u001cecho hi\n\u001c```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"list item: U+001C after - carries a claim", "-\u001cran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: U+001C after > carries a claim", ">\u001cran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"closer tail: U+001C after the closing fence", "```bash verified:x\necho hi\n```\u001c\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"comment payload: only U+001C is empty", "ran 5 tests <!-- measured:\u001c-->\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: measured:/predicted: comment carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"tag payload: only U+001C is empty", "```bash verified:\u001c\necho hi\n```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: verified:/unverified: tag on the info string carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: U+001F separates the number from its unit", "ran 5\u001ftests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"provenance comment: U+001F after <!-- satisfies", "ran 5 tests <!--\u001fmeasured: x -->\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"fence tag: U+001F before the tag", "```bash\u001fverified:ran it\necho hi\n```\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"list marker: U+001F after - opens a tagged fence", "-\u001f```bash verified:x\n  echo hi\n  ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"checkbox marker: U+001F after [ ] opens a tagged fence", "- [ ]\u001f```bash verified:x\n      echo hi\n      ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"ordered list marker: U+001F after 1. opens a tagged fence", "1.\u001f```bash verified:x\n   echo hi\n   ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: U+001F after > opens a tagged fence", ">\u001f```bash verified:x\n>\u001fecho hi\n>\u001f```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: U+001F indents a tagged fence", "\u001f```bash verified:x\n\u001fecho hi\n\u001f```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"list item: U+001F after - carries a claim", "-\u001fran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: U+001F after > carries a claim", ">\u001fran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"closer tail: U+001F after the closing fence", "```bash verified:x\necho hi\n```\u001f\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"comment payload: only U+001F is empty", "ran 5 tests <!-- measured:\u001f-->\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: measured:/predicted: comment carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"tag payload: only U+001F is empty", "```bash verified:\u001f\necho hi\n```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: verified:/unverified: tag on the info string carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: space separates the number from its unit", "ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"provenance comment: space after <!-- satisfies", "ran 5 tests <!-- measured: x -->\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"fence tag: space before the tag", "```bash verified:ran it\necho hi\n```\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"list marker: space after - opens a tagged fence", "- ```bash verified:x\n  echo hi\n  ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"checkbox marker: space after [ ] opens a tagged fence", "- [ ] ```bash verified:x\n      echo hi\n      ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"ordered list marker: space after 1. opens a tagged fence", "1. ```bash verified:x\n   echo hi\n   ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: space after > opens a tagged fence", "> ```bash verified:x\n> echo hi\n> ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: space indents a tagged fence", " ```bash verified:x\n echo hi\n ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"list item: space after - carries a claim", "- ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote: space after > carries a claim", "> ran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"closer tail: space after the closing fence", "```bash verified:x\necho hi\n``` \nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"comment payload: only space is empty", "ran 5 tests <!-- measured: -->\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: measured:/predicted: comment carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"tag payload: only space is empty", "```bash verified: \necho hi\n```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: verified:/unverified: tag on the info string carries no evidence after the colon", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: \u00e9 before the number is a word boundary miss", "x\u00e95 tests\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: \u00e9 after the unit is a word boundary miss", "ran 5 tests\u00e9\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: \u00b2 before the number is a word boundary miss", "x\u00b25 tests\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: \u00b2 after the unit is a word boundary miss", "ran 5 tests\u00b2\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: \u0663 before the number is a word boundary miss", "x\u06635 tests\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: \u0663 after the unit is a word boundary miss", "ran 5 tests\u0663\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: _ before the number is a word boundary miss", "x_5 tests\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: _ after the unit is a word boundary miss", "ran 5 tests_\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: \u00c9- before the number is still a claim", "\u00c9-5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: KAN- before the number is not a claim", "KAN-5 tests\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: a 21-digit run is a claim", "111111111111111111111 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"claim: a 22-digit run is not a claim", "1111111111111111111111 tests\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"claim: a 21-digit comma run is a claim", "1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: a nested list marker's NBSP gap (tagged fence)", "-\u00a0-\u00a0```bash verified:x\n-\u00a0-\u00a0echo hi\n-\u00a0-\u00a0```\nran 5 tests\n", 4, []string{"spectre/changes/demo-change/tasks.md:1: line has a fence-like run of backticks/tildes behind a prefix this guard does not recognise \u2014 a doubled/glued container marker, a bare checkbox with no preceding list marker, or a list marker sitting behind a gap wider than 3 columns \u2014 refusing to guess whether it opens or closes a fence. Fix by removing the doubled marker, adding the missing list marker before the checkbox, or narrowing the gap to 0-3 columns, so the container is unambiguous. Scanning of this file stopped here; re-run after fixing to see anything further on in it.", "1 line(s) could not be classified (content-classification, see above) \u2014 refusing to report a clean run"}},
		{"container prefix: a nested list marker's NBSP gap (untagged fence)", "-\u00a0-\u00a0```bash\n-\u00a0-\u00a0echo hi\n-\u00a0-\u00a0```\n", 4, []string{"spectre/changes/demo-change/tasks.md:1: line has a fence-like run of backticks/tildes behind a prefix this guard does not recognise \u2014 a doubled/glued container marker, a bare checkbox with no preceding list marker, or a list marker sitting behind a gap wider than 3 columns \u2014 refusing to guess whether it opens or closes a fence. Fix by removing the doubled marker, adding the missing list marker before the checkbox, or narrowing the gap to 0-3 columns, so the container is unambiguous. Scanning of this file stopped here; re-run after fixing to see anything further on in it.", "1 line(s) could not be classified (content-classification, see above) \u2014 refusing to report a clean run"}},
		{"container prefix: a bare checkbox's NBSP gap (tagged fence)", "[ ]\u00a0```bash verified:x\n[ ]\u00a0echo hi\n[ ]\u00a0```\nran 5 tests\n", 4, []string{"spectre/changes/demo-change/tasks.md:1: line has a fence-like run of backticks/tildes behind a prefix this guard does not recognise \u2014 a doubled/glued container marker, a bare checkbox with no preceding list marker, or a list marker sitting behind a gap wider than 3 columns \u2014 refusing to guess whether it opens or closes a fence. Fix by removing the doubled marker, adding the missing list marker before the checkbox, or narrowing the gap to 0-3 columns, so the container is unambiguous. Scanning of this file stopped here; re-run after fixing to see anything further on in it.", "1 line(s) could not be classified (content-classification, see above) \u2014 refusing to report a clean run"}},
		{"container prefix: a bare checkbox's NBSP gap (untagged fence)", "[ ]\u00a0```bash\n[ ]\u00a0echo hi\n[ ]\u00a0```\n", 4, []string{"spectre/changes/demo-change/tasks.md:1: line has a fence-like run of backticks/tildes behind a prefix this guard does not recognise \u2014 a doubled/glued container marker, a bare checkbox with no preceding list marker, or a list marker sitting behind a gap wider than 3 columns \u2014 refusing to guess whether it opens or closes a fence. Fix by removing the doubled marker, adding the missing list marker before the checkbox, or narrowing the gap to 0-3 columns, so the container is unambiguous. Scanning of this file stopped here; re-run after fixing to see anything further on in it.", "1 line(s) could not be classified (content-classification, see above) \u2014 refusing to report a clean run"}},
		{"container prefix: a list checkbox's NBSP gap (tagged fence)", "-\u00a0[ ]\u00a0```bash verified:x\n-\u00a0[ ]\u00a0echo hi\n-\u00a0[ ]\u00a0```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: a list checkbox's NBSP gap (untagged fence)", "-\u00a0[ ]\u00a0```bash\n-\u00a0[ ]\u00a0echo hi\n-\u00a0[ ]\u00a0```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: NBSP after > counts toward the blockquote indent (untagged fence)", ">\u00a0\u00a0\u00a0\u00a0```bash\n>\u00a0\u00a0\u00a0\u00a0echo hi\n>\u00a0\u00a0\u00a0\u00a0```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block has no verified:/unverified: tag", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: NBSP after > counts toward the blockquote indent (untagged fence)", ">\u00a0   ```bash\n>\u00a0   echo hi\n>\u00a0   ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block has no verified:/unverified: tag", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: leading NBSP before a list marker with a wide gap (tagged fence)", "\u00a0-     ```bash verified:x\n\u00a0-     echo hi\n\u00a0-     ```\nran 5 tests\n", 1, []string{"spectre/changes/demo-change/tasks.md:4: numeric claim with no measured:/predicted: provenance comment", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"container prefix: leading NBSP before a list marker with a wide gap (untagged fence)", "\u00a0-     ```bash\n\u00a0-     echo hi\n\u00a0-     ```\n", 0, []string{"check-plan-provenance: 1 file(s) scanned, all provenance stated"}},
		{"blockquote closer: NBSP\u00d74 before > is not a closer", "> ```bash verified:x\n> echo hi\n\u00a0\u00a0\u00a0\u00a0> ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"nested blockquote closer: NBSP\u00d77 before > is not a closer", "> - > ```bash verified:x\n> - > echo hi\n>\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0> ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote closer: U+3000\u00d74 before > is not a closer", "> ```bash verified:x\n> echo hi\n\u3000\u3000\u3000\u3000> ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"nested blockquote closer: U+3000\u00d77 before > is not a closer", "> - > ```bash verified:x\n> - > echo hi\n>\u3000\u3000\u3000\u3000\u3000\u3000\u3000> ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"blockquote closer: U+001C\u00d74 before > is not a closer", "> ```bash verified:x\n> echo hi\n\u001c\u001c\u001c\u001c> ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
		{"nested blockquote closer: U+001C\u00d77 before > is not a closer", "> - > ```bash verified:x\n> - > echo hi\n>\u001c\u001c\u001c\u001c\u001c\u001c\u001c> ```\n", 1, []string{"spectre/changes/demo-change/tasks.md:1: fenced code block never closed", "1 plan-provenance violation(s) found.", "For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment \u2014 or, where a note above says the quotation exemption was withdrawn, by rewording the line; never suppress."}},
	}
	cases := make([]ppCase, 0, len(rows))
	for _, row := range rows {
		cases = append(cases, ppCase{"unicode matchers: " + row.name, func(t *testing.T) {
			r := ppTasks(t, row.body)
			ok(t, "exit code is the Python guard's", r.rc == row.rc, r)
			for _, l := range row.lines {
				ok(t, "output carries the Python guard's line "+l, has(r.out, l), r)
			}
		}})
	}
	return cases
}

// TestPpReprMatchesPython holds ppRepr, which the guard's
// `[Errno N] …: '<path>'` refusals render a path through, to python3's own
// repr() over every escaping class: quotes, backslash, C0, DEL, C1, a
// non-ASCII space, a zero-width and a line separator, a private-use and an
// astral character.
func TestPpReprMatchesPython(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	for _, in := range []string{
		"plain", "it's", `say "hi"`, `both ' and "`, `back\slash`, "tab\tnl\ncr\r",
		"nul\x00del\x7f", "c1\u0085\u009f", "nb sp", "zw​sp", "ls ps ",
		"pua", "astral\U0001f600", "astral-pua\U000f0000", "é日本",
	} {
		t.Run(fmt.Sprintf("%q", in), func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("python3", "-c", `import sys; print(repr(sys.stdin.buffer.read().decode("utf-8")), end="")`)
			cmd.Stdin = strings.NewReader(in)
			want, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}
			if got := ppRepr(in); got != string(want) {
				t.Fatalf("ppRepr(%q) = %s, python3 says %s", in, got, want)
			}
		})
	}
}
