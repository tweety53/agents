package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckTaskBuildGreen ports scripts/test-check-task-build-green.sh at
// bbfe685e: one subtest per ok: label, the guard run in-process against real
// fixture files (and a real git repository for the unlanded-archive case).
// The harness's wrapper cases — a missing sibling, a flow-guard that cannot
// be built, the no-argument scan of this repository — exec the real shim.
func TestCheckTaskBuildGreen(t *testing.T) {
	t.Parallel()
	type check struct{ label, fail string }
	var checks []check
	want := func(label string, ok bool, r guardResult) {
		c := check{label: label}
		if !ok {
			c.fail = fmt.Sprintf("rc=%d out=%q", r.rc, r.out)
		}
		checks = append(checks, c)
	}
	has := func(r guardResult, s string) bool { return strings.Contains(r.out, s) }
	env := Env{Getenv: os.Getenv}
	one := func(path string) guardResult { return runGuard("check-task-build-green", []string{path}, env) }
	scan := func(root string) guardResult {
		return runGuard("check-task-build-green", nil, Env{Getenv: psGetenv(map[string]string{"CHECK_TASK_BUILD_GREEN_ROOT": root})})
	}
	run := func(body string) (string, guardResult) {
		md := psFixture(t, body)
		return md, one(md)
	}
	clean := func(n string, r guardResult) {
		want("case "+n+": no output", r.out == "", r)
	}

	_, r := run("## 1 Group\n\n- [ ] 1. First task\n\n**Build:** green\n\n- [ ] 2. Second task\n\n**Build:** green\n")
	want("case 1: all green passes", r.rc == 0, r)
	clean("1", r)

	_, r = run("- [ ] 1. Untagged task\n\nSome body text, no build tag at all.\n")
	want("case 2: missing tag fails", r.rc == 1, r)
	want("case 2: names the task id", has(r, "task 1 has no **Build:** tag"), r)

	_, r = run("- [ ] 1. Red with no partner\n\n**Build:** red\n\n**Squash-with:** Task ,\n")
	want("case 3: red with no partner fails", r.rc == 1, r)
	want("case 3: names the task id", has(r, "task 1 is red with no merge partner named"), r)

	_, r = run("- [ ] 1. Red with nonexistent partner\n\n**Build:** red\n\n**Squash-with:** Task 9\n")
	want("case 4: nonexistent partner fails", r.rc == 1, r)
	want("case 4: names the missing partner", has(r, "task 1 merges with Task 9, which does not exist in this plan"), r)

	mutual := "- [ ] 1. First task\n\n**Build:** red\n\n**Squash-with:** Task 2\n\n- [ ] 2. Second task\n\n**Build:** red\n\n**Squash-with:** Task 1\n"
	_, r = run(mutual)
	want("case 5: partner itself red fails", r.rc == 1, r)
	want("case 5: reported against task 1", has(r, "task 1 merges with Task 2, which is itself red"), r)

	_, r = run("- [ ] 1. First task\n\n**Build:** red\n\n**Squash-with:** Task 2\n\n- [ ] 2. Second task\n\n**Build:** green\n")
	want("case 6: partner green passes", r.rc == 0, r)
	clean("6", r)

	_, r = run("- [ ] 1. Malformed tag\n\n**Build:** yellow\n")
	want("case 7: malformed tag fails", r.rc == 1, r)
	want("case 7: reported as a malformed tag at its own line", has(r, `tasks.md:3: task 1 has a **Build:** line reading "yellow", which is neither green nor red`), r)

	_, r = run("- [ ] 1. Untagged real task\n\nExample of the wrong way:\n\n```markdown\n**Build:** green\n```\n")
	want("case 8: fenced example tag does not count as real", r.rc == 1, r)
	want("case 8: still reports missing tag", has(r, "task 1 has no **Build:** tag"), r)

	_, r = run("- [ ] 1. Tagged real task\n\nExample of the wrong way:\n\n```markdown\n- [ ] 9. Example\n**Build:** red\n\n**Squash-with:** Task 9\n```\n\n**Build:** green\n")
	want("case 9: fenced task line+tag does not corrupt real task", r.rc == 0, r)
	clean("9", r)

	_, r = run("- [ ] 1. First\n**Build:** green\n- [ ] 1. Duplicate id\n**Build:** green\n")
	want("case 10: duplicate id fails", r.rc == 1, r)
	want("case 10: names the duplicate and first line", has(r, "task 1 is defined more than once (first at line 1)"), r)

	_, r = run("- [ ] 1. \nSome body text, no tag.\n")
	want("case 11: titleless task line recognised as a task", r.rc == 1, r)
	want("case 11: reports the missing tag for 1", has(r, "task 1 has no **Build:** tag"), r)

	_, r = run("- [ ] 1. First\n\n**Build:** red\n\n**Squash-with:** Task 2, 2\n\n- [ ] 2. Second\n\n**Build:** red\n\n**Squash-with:** Task 1\n")
	want("case 12: duplicate partner id fails", r.rc == 1, r)
	want("case 12: exactly one violation line for the duplicated partner", strings.Count(r.out, "task 1 merges with Task 2, which is itself red") == 1, r)

	// Case 13: the shim's no-argument scan of this repository's own changes.
	r = psExec(t, []string{"CHECK_TASK_BUILD_GREEN_ROOT=", "FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, tbgShim)
	want("case 13: no-arg scan of the real repo tree passes", r.rc == 0, r)

	// Case 14: the no-argument scan over CHECK_TASK_BUILD_GREEN_ROOT; the
	// fixture is no git worktree, so its archived change stays out.
	root := t.TempDir()
	writeFile(t, root+"/spectre/changes/change-a/tasks.md", "- [ ] 1. Clean task\n\n**Build:** green\n")
	writeFile(t, root+"/spectre/changes/change-b/tasks.md", "- [ ] 1. Untagged task\n\nNo tag here.\n")
	writeFile(t, root+"/spectre/changes/archive/2024-01-01-old-change/tasks.md", "- [ ] 1. Archived, also untagged\n\nNo tag here either -- must never be scanned.\n")
	r = scan(root)
	want("case 14: aggregation reports non-zero when any file violates", r.rc != 0, r)
	want("case 14: surfaces change-b's own violation", has(r, "change-b/tasks.md:1: task 1 has no **Build:** tag"), r)
	want("case 14: archived tasks.md excluded from the scan", !has(r, "archive/"), r)

	// Case 14b: an archived change joins the scan only while the base lacks it.
	root = t.TempDir()
	untagged := "- [ ] 1. Archived, untagged\n\nNo tag here.\n"
	writeFile(t, root+"/spectre/changes/archive/landed-change/tasks.md", untagged)
	gitRun(t, root, "init", "-q", "-b", "main")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "chore(spectre): archive landed-change")
	writeFile(t, root+"/spectre/changes/archive/new-change/tasks.md", untagged)
	r = scan(root)
	want("case 14b: an archived change the base lacks is scanned", r.rc == 1, r)
	want("case 14b: names the unlanded archived change's violation", has(r, "archive/new-change/tasks.md:1: task 1 has no **Build:** tag"), r)
	want("case 14b: an archived change the base carries stays out of the scan", !has(r, "landed-change"), r)

	_, r = run("- [ ] 1. Red task\n\n**Build:** red\n\n**Squash-with:** Task 2\n\n- [ ] 2. Green partner\n\n**Build:** green\n")
	want("case 15: partner read from Squash-with: passes", r.rc == 0, r)
	clean("15", r)

	_, r = run("- [ ] 1. Red with no Squash-with field\n\n**Build:** red\n")
	want("case 16: red with no Squash-with: field fails", r.rc == 1, r)
	want("case 16: names the missing-field violation", has(r, "task 1 is red with no **Squash-with:** field"), r)

	_, r = run(mutual)
	want("case 17: Squash-with: partner must be green fails", r.rc == 1, r)
	want("case 17: reported against task 1", has(r, "task 1 merges with Task 2, which is itself red"), r)

	_, r = run("- [ ] 1. Red task whose Squash-with value carries free text\n\n**Build:** red\n\n**Squash-with:** Task 2 (see step 3)\n\n- [ ] 2. Green partner\n\n**Build:** green\n")
	want("case 18: a Squash-with value that does not gate fails", r.rc == 1, r)
	want("case 18: an ungated value is reported as no field at all", has(r, "task 1 is red with no **Squash-with:** field"), r)

	_, r = run("- [ ] 3. Red task whose Squash-with is followed by prose\n\n**Build:** red\n\n**Squash-with:** Task 4\nThe fold is described in the paragraph above.\n\n- [ ] 4. Green partner\n\n**Build:** green\n")
	want("case 18: the field is line-scoped, so a following prose line is not part of its value", r.rc == 0, r)

	_, r = run("- [ ] 1. Red task whose first Squash-with line does not gate\n\n**Build:** red\n\n**Squash-with:** Task 2 (see note below)\n**Squash-with:** Task 2\n\n- [ ] 2. Green partner named by the gating line\n\n**Build:** green\n")
	want("case 19: the gating line is the field, so the partner resolves", r.rc == 0, r)
	clean("19", r)

	// Case 20: the shim's own sibling missing exits 2 naming it — the
	// sibling is lib/flow-guard.sh now that no Python module is loaded.
	md, _ := run("- [ ] 1. Clean task\n\n**Build:** green\n")
	stripped := t.TempDir()
	psCopy(t, tbgShim, stripped+"/check-task-build-green.sh")
	r = psExec(t, nil, stripped+"/check-task-build-green.sh", md)
	want("case 20: a guard copy with no lib/ sibling exits 2", r.rc == 2, r)
	want("case 20: the message names the missing module path", has(r, "cannot load lib/flow-guard.sh beside "+stripped+"/check-task-build-green.sh"), r)

	_, r = run("- [ ] 1. Red task carrying two Build lines\n\n**Build:** red\n**Build:** green\n\n**Squash-with:** Task 9\n")
	want("case 21: the first Build line is the tag, so the task is red", r.rc == 1, r)
	want("case 21: the red task's missing partner is reported", has(r, "task 1 merges with Task 9, which does not exist in this plan"), r)

	_, r = run("- [ ] 1. Red task whose tag line is followed by prose\n\n**Build:** red\nthis sentence explains the tag and is not part of it\n\n**Squash-with:** Task 9\n")
	want("case 22: a prose line under the tag does not unset it", r.rc == 1, r)
	want("case 22: the red task's missing partner is reported", has(r, "task 1 merges with Task 9, which does not exist in this plan"), r)

	_, r = run("- [ ] 1. Real task with a worked example in its body\n\n**Build:** green\n\nExample of a fold, shown but never declared:\n\n```\n- [ ] 9. Example red task\n**Build:** red\n**Squash-with:** Task 8 (see the note)\n```\n")
	want("case 23: a fenced example task line opens no task", r.rc == 0, r)
	clean("23", r)

	_, r = run("- [ ] 1. First task line for this id\n\n**Build:** green\n\n- [ ] 1. Second task line reusing the id\n\n**Build:** red\n\n- [ ] 3. Red task folding into Task 1\n\n**Build:** red\n**Squash-with:** Task 1\n")
	want("case 24: the duplicate id is still reported", r.rc == 1, r)
	want("case 24: names the duplicate and first line", has(r, "task 1 is defined more than once (first at line 1)"), r)
	want("case 24: a duplicated id resolves to its first task line", !has(r, "task 3 merges with Task 1, which is itself red"), r)

	_, r = run("- [ ] 1. Task whose body opens a fence it never closes\n\n~~~\n\n**Build:** green\n**Files:** `alpha.txt`\n**Commit:** test: add alpha\n")
	want("case 25: an unclosed fence in a task body fails", r.rc == 1, r)
	want("case 25: names the task and the line the fence opened on", has(r, ":3: task 1 opens a code fence here that is never closed"), r)
	want("case 25: the untagged-task noise is replaced, not accompanied", !has(r, "has no **Build:** tag"), r)

	_, r = run("- [ ] 1. Task whose body closes the fence it opens\n\n**Build:** green\n**Files:** `alpha.txt`\n**Commit:** test: add alpha\n\n~~~\nexample\n~~~\n")
	want("case 25: a closed fence is not reported as unclosed", r.rc == 0, r)

	_, r = run("- [ ] 1. Untagged task\n\nSome body text, no build tag at all.\n")
	want("case 26a: a checkbox line opens a task", r.rc == 1, r)
	want("case 26a: names the task id read off the checkbox line", has(r, "task 1 has no **Build:** tag"), r)

	_, r = run("- [x] 2. Untagged done task\n\nSome body text, no build tag at all.\n")
	want("case 26b: a checked task line is still a task", r.rc == 1, r)
	want("case 26b: names the checked task's id", has(r, "task 2 has no **Build:** tag"), r)

	_, r = run("- [ ] 3. Tagged task with steps\n\n**Build:** green\n\n  - [ ] **Step 1: a step, never a task**\n  - [x] **Step 2: a done step, still never a task**\n")
	want("case 26c: step checkboxes are not tasks", r.rc == 0, r)
	clean("26c", r)

	_, r = run("- [ ] 5. The only real task\n\n**Build:** green\n\n### 4 A heading, not a task\n\nNo tag here.\n")
	want("case 26d: a level-3 heading opens no task", r.rc == 0, r)
	clean("26d", r)

	_, r = run("- [ ] 1. Green task\n\n**Build:** green\n\n- [ ] 1.1. A dotted task line, which is no task\n\n- [ ] 2. Red task whose partner id is dotted\n\n**Build:** red\n\n**Squash-with:** Task 1.1\n")
	want("case 27: a dotted partner id fails", r.rc == 1, r)
	want("case 27a: a dotted task line opens no task", !has(r, "has no **Build:** tag"), r)
	want("case 27b: the dotted partner id names no task, by its own name", has(r, "task 2 merges with Task 1.1, which does not exist in this plan"), r)

	_, r = run("- [ ] 1. Green task\n\n**Build:** green\n\n  - [ ] 2. An indented task-shaped line, which is no task\n")
	want("case 28: an indented task-shaped line opens no task", r.rc == 0, r)
	clean("28", r)

	_, r = run("- [ ] 1. Green tag with trailing prose\n\n**Build:** green — `:shared:desktopTest` 2259, matching the predicted +1 exactly\n")
	want("case 29: green followed by prose is green", r.rc == 0, r)
	clean("29", r)

	_, r = run("- [ ] 1. Green tag with a parenthesised remark\n\n**Build:** green (unchanged; `:shared:desktopTest` still 2258, nothing was touched)\n")
	want("case 30: green with a parenthesised remark is green", r.rc == 0, r)
	clean("30", r)

	_, r = run("- [ ] 1. Red tag with trailing prose\n\n**Build:** red — lands with task 2\n**Squash-with:** Task 2\n\n- [ ] 2. Green partner\n\n**Build:** green\n")
	want("case 31: red followed by prose is red and resolves its partner", r.rc == 0, r)
	clean("31", r)

	_, r = run("- [ ] 1. Not quite green\n\n**Build:** greenish\n")
	want("case 32: greenish is not green", r.rc == 1, r)
	want("case 32: reported as a malformed tag", has(r, `task 1 has a **Build:** line reading "greenish", which is neither green nor red`), r)

	_, r = run("- [ ] 1. A tag that names no colour\n\n**Build:** `tests/` 33 of 34 — see the flake below\n")
	want("case 33: a keyword-less value fails", r.rc == 1, r)
	want("case 33: the value is quoted in the message", has(r, "task 1 has a **Build:** line reading \"`tests/` 33 of 34 — see the flake below\", which is neither green nor red"), r)

	_, r = run("- [ ] 1. Malformed then green\n\n**Build:** yellow\n**Build:** green\n")
	want("case 34: the first Build line decides", r.rc == 1, r)
	want("case 34: the malformed first line is the one reported", has(r, `tasks.md:3: task 1 has a **Build:** line reading "yellow"`), r)

	_, r = run("- [ ] 1. No space before the keyword\n\n**Build:**green\n")
	want("case 35: no space before the keyword fails", r.rc == 1, r)
	want("case 35: reported as a malformed tag, not accepted as green", has(r, `task 1 has a **Build:** line reading "green", which is neither green nor red`), r)

	_, r = run("- [ ] 1. Some task\n\n**Build:** yellow\n")
	want("case 36: malformed tag fails", r.rc == 1, r)
	want("case 36: no duplicate no-tag violation", !has(r, "has no **Build:** tag"), r)

	// Unlanded archives: a shim copy with no stats/ beside it cannot build
	// flow-guard; the same copy with no lib/flow-guard.sh cannot load it.
	u := t.TempDir()
	mkdir(t, u+"/root/spectre/changes/some-change")
	psCopy(t, tbgShim, u+"/scripts/check-task-build-green.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", u+"/scripts/lib/flow-guard.sh")
	uenv := []string{"CHECK_TASK_BUILD_GREEN_ROOT=" + u + "/root", "FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
	r = psExec(t, uenv, u+"/scripts/check-task-build-green.sh")
	want("unlanded archives: flow-guard that cannot be built exits 2", r.rc == 2 && has(r, "cannot build flow-guard"), r)
	if err := os.Remove(u + "/scripts/lib/flow-guard.sh"); err != nil {
		t.Fatal(err)
	}
	r = psExec(t, uenv, u+"/scripts/check-task-build-green.sh")
	want("unlanded archives: a missing lib/flow-guard.sh exits 2", r.rc == 2 && has(r, "cannot load lib/flow-guard.sh beside "+u+"/scripts/check-task-build-green.sh"), r)

	// Beyond the harness: the outputs it never pinned, as the wrapper
	// printed them at bbfe685e.
	sandbox := t.TempDir()
	psCopy(t, tbgShim, sandbox+"/scripts/check-task-build-green.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", sandbox+"/scripts/lib/flow-guard.sh")
	psLink(t, filepath.Dir(psScripts)+"/stats", sandbox+"/stats")
	psLink(t, "../../../scripts/check-task-build-green.sh", sandbox+"/skills/flow/scripts/check-task-build-green.sh")
	psLink(t, "../../../scripts/lib", sandbox+"/skills/flow/scripts/lib")
	writeFile(t, sandbox+"/spectre/changes/dirty-change/tasks.md", "- [ ] 1. Dirty\n")
	r = psExec(t, []string{"CHECK_TASK_BUILD_GREEN_ROOT=", "FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, sandbox+"/skills/flow/scripts/check-task-build-green.sh")
	want("pin: the symlinked shim scans the real root, not skills/flow", r.rc == 1 &&
		strings.HasSuffix(r.stdout, sandbox+"/spectre/changes/dirty-change/tasks.md:1: task 1 has no **Build:** tag\n") &&
		strings.Count(r.stdout, "\n") == 1, r)
	missing := t.TempDir() + "/nope.md"
	r = one(missing)
	want("pin: a missing file exits 2 with the OSError on stderr", r.rc == 2 && r.stdout == "" &&
		r.err == missing+": cannot read file: [Errno 2] No such file or directory: '"+missing+"'\n", r)
	r = one(t.TempDir())
	want("pin: a directory reads as Errno 21", r.rc == 2 && strings.Contains(r.err, ": cannot read file: [Errno 21] Is a directory: '"), r)
	// Python let the UnicodeDecodeError escape as a traceback (exit 1); the
	// header's "the file cannot be read" is exit 2, and the header wins.
	_, r = run("- [ ] 1. \xff\n")
	want("pin: a plan that is not UTF-8 exits 2", r.rc == 2 && has(r, "cannot read file: UnicodeDecodeError"), r)
	r = runGuard("check-task-build-green", []string{"a", "b"}, env)
	want("pin: the usage line", r.rc == 2 && r.err == "usage: check-task-build-green.sh [path-to-tasks.md]\n", r)
	md, r = run("- [ ] 1. A\n\n- [ ] 2. B\n\n**Build:** red\n\n- [ ] 3. C\n\n**Build:** red\n\n**Squash-with:** Task 2, 9\n\n- [ ] 1. A again\n")
	want("pin: violations go to stdout in order, duplicates first", r.rc == 1 && r.err == "" && r.stdout == md+":13: task 1 is defined more than once (first at line 1)\n"+
		md+":1: task 1 has no **Build:** tag\n"+
		md+":5: task 2 is red with no **Squash-with:** field\n"+
		md+":11: task 3 merges with Task 2, which is itself red\n"+
		md+":11: task 3 merges with Task 9, which does not exist in this plan\n"+
		md+":13: task 1 has no **Build:** tag\n", r)
	root = t.TempDir()
	writeFile(t, root+"/spectre/changes/dirty-change/tasks.md", "- [ ] 1. Dirty\n")
	r = runGuard("check-task-build-green", nil, Env{Getenv: psGetenv(map[string]string{"FLOW_GUARD_REPO_ROOT": root})})
	want("pin: with no override the root is FLOW_GUARD_REPO_ROOT", r.rc == 1 && r.out == root+"/spectre/changes/dirty-change/tasks.md:1: task 1 has no **Build:** tag", r)
	r = runGuard("check-task-build-green", nil, Env{Getenv: psGetenv(nil)})
	want("pin: with no override and no FLOW_GUARD_REPO_ROOT it exits 2", r.rc == 2 &&
		r.err == "check-task-build-green: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-task-build-green.sh, which sets it\n", r)
	r = scan(t.TempDir())
	want("pin: a root with no changes directory passes silently", r.rc == 0 && r.out == "", r)
	root = t.TempDir()
	writeFile(t, root+"/spectre/changes/bad/tasks.md", "- [ ] 1. \xff\n")
	writeFile(t, root+"/spectre/changes/dirty/tasks.md", "- [ ] 1. Dirty\n")
	r = scan(root)
	want("pin: an unreadable file makes the scan exit 2 and the rest is still checked", r.rc == 2 && has(r, "dirty/tasks.md:1: task 1 has no **Build:** tag"), r)

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}

var tbgShim = psScripts + "/check-task-build-green.sh"
