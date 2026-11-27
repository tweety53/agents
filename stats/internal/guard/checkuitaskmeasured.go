package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// flow-guard check-ui-task-measured — the measured gate of a UI task: answer
// whether a task whose **Files:** touch the project's `ui paths` was measured
// against its frames after its last source commit, and measured clean.
// `close-task` runs it in-process for every task it closes, so "the
// implementer measures" is a refusal rather than a sentence (kan-1155: the
// fixers skipped it round after round).
//
// Usage: flow-guard check-ui-task-measured <worktree> <task-id> [<peer worktree>…]
//
// <worktree> is the change's canonical worktree, on `spectre/<name>`; the
// plan is <worktree>/<spec root>/changes/<name>/tasks.md. A task file touches
// `ui paths` when it matches one of the `## visual verification` globs as
// written, or as `<repository name>/<file>` for the repository of <worktree>
// or of a <peer worktree> the file exists in — a cross-repository task's
// files are relative to their own worktree, and the globs name that
// repository by its main checkout's basename.
//
// A file named `*Samples.kt` or `Fake*` draws no frame and never makes a task a UI task.
//
// The record is <changeRoot>/visual-verification/task-<task-id>-measure.json,
// the manifest `flow-guard visual-verify-pass measure <worktree> <name>
// task-<task-id> <capture spec>` writes there beside its own. It must
//   - exist and carry at least one pair,
//   - be no older than the newest commit touching the task's files in any of
//     the worktrees, and
//   - carry no flagged band, seam or `over` row and no failed compose, after
//     the declared tolerance and the whole-frame shift fold.
//
// A UI task is not measured per task when a LATER task of the plan lists a
// `*.mockups` capture-spec sidecar in its **Files:** — no capture spec exists
// until then, and flow.verify's visual stage measures the frames.
//
// Prints ONE verdict line:
//
//	UI-TASK-MEASURED: <task-id> — <why it passes>            (stdout)
//	UI-TASK-UNMEASURED: <task-id> — <reason> — run: <command> (stderr)
//
// Exit 0 measured clean, or not a UI task (no `ui paths`, no plan, or no file
// of the task matches); 1 the record is missing, stale or flagged; 2 a usage
// error, a worktree off a `spectre/<name>` branch, or a task the plan lacks.
// Its verdict, utmJudge, runs in process from close-task; no guard.Registry key.
func checkUITaskMeasured(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard check-ui-task-measured <worktree> <task-id> [<peer worktree>…]")
		return 2
	}
	wt := pcAbs(env, args[0])
	branch, _ := capture(envGit(env)("-C", wt, "rev-parse", "--abbrev-ref", "HEAD"))
	name, onChange := strings.CutPrefix(strings.TrimSpace(branch), "spectre/")
	if !onChange {
		Reason(env, "not-a-change-worktree")
		fmt.Fprintf(stderr, "check-ui-task-measured: %s is not on a spectre/<name> branch\n", wt)
		return 2
	}
	worktrees := []string{wt}
	for _, p := range args[2:] {
		worktrees = append(worktrees, pcAbs(env, p))
	}
	id, msg := utmJudge(env, worktrees, name, args[1])
	switch id {
	case "":
		fmt.Fprintf(stdout, "UI-TASK-MEASURED: %s — %s\n", args[1], msg)
		return 0
	case "task-not-in-plan":
		Reason(env, id)
		fmt.Fprintf(stderr, "check-ui-task-measured: %s\n", msg)
		return 2
	}
	Reason(env, id)
	fmt.Fprintf(stderr, "UI-TASK-UNMEASURED: %s — %s\n", args[1], msg)
	return 1
}

// utmJudge is the gate's verdict for one task: an empty id passes, msg saying
// why; `task-not-in-plan` cannot be answered; `measure-missing`, `measure-stale`
// and `measure-not-clean` refuse, msg ending in the command that repairs it.
// worktrees is the canonical worktree, then the task's other worktrees.
func utmJudge(env Env, worktrees []string, name, taskID string) (id, msg string) {
	wt := worktrees[0]
	changeRoot := filepath.Join(wt, specRootLeaf(wt, io.Discard), "changes", name)
	raw, err := os.ReadFile(filepath.Join(changeRoot, "tasks.md"))
	if err != nil {
		return "", "no plan at " + filepath.Join(changeRoot, "tasks.md")
	}
	cfg, _ := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	var globs []string
	for _, g := range strings.Split(vvCells(string(cfg))["ui paths"], ",") {
		if g = strings.Trim(strings.TrimSpace(g), "`"); g != "" {
			globs = append(globs, g)
		}
	}
	if len(globs) == 0 {
		return "", "the project declares no `ui paths`"
	}
	lines, err := tcfPlanLines(raw)
	if err != nil {
		return "task-not-in-plan", "cannot read " + filepath.Join(changeRoot, "tasks.md") + ": " + err.Error()
	}
	task, found := tcfParseTask(lines, taskID)
	if !found {
		return "task-not-in-plan", "task " + taskID + " is not in " + filepath.Join(changeRoot, "tasks.md")
	}
	git := envGit(env)
	ui := false
	for _, f := range task.files {
		f = strings.TrimPrefix(f, "./")
		if b := filepath.Base(f); strings.HasSuffix(b, "Samples.kt") || strings.HasPrefix(b, "Fake") {
			continue // a sample or fake draws no frame
		}
		names := []string{f}
		for _, w := range worktrees {
			if _, err := os.Stat(filepath.Join(w, f)); err != nil {
				continue // not this repository's file
			}
			repo := filepath.Base(w)
			if common, ok := capture(git("-C", w, "rev-parse", "--path-format=absolute", "--git-common-dir")); ok {
				repo = filepath.Base(filepath.Dir(strings.TrimSpace(common)))
			}
			names = append(names, repo+"/"+f)
		}
		for _, g := range globs {
			re := vtGlobRegexp(g)
			for _, n := range names {
				ui = ui || re.MatchString(vtBytesAsRunes(n))
			}
		}
	}
	if !ui {
		return "", "no file of the task matches `ui paths`"
	}

	// A plan that adds its capture spec (a `.mockups` sidecar) in a later task
	// leaves nothing to measure against yet; flow.verify measures these frames.
	after := false
	for _, t := range tcfIterTasks(lines) {
		if t.id == taskID {
			after = true
			continue
		}
		if !after {
			continue
		}
		later, _ := tcfParseTask(lines, t.id)
		for _, f := range later.files {
			if strings.HasSuffix(f, ".mockups") {
				return "", "the plan adds its capture spec in task " + t.id + "; flow.verify measures this task's frames"
			}
		}
	}

	record := filepath.Join(changeRoot, "visual-verification", "task-"+taskID+"-measure.json")
	run := fmt.Sprintf("run: flow-guard visual-verify-pass measure %s %s task-%s <the task's capture spec>", wt, name, taskID)
	st, err := os.Stat(record)
	var m vvmManifestOut
	if body, rerr := os.ReadFile(record); err != nil || rerr != nil || json.Unmarshal(body, &m) != nil || len(m.Pairs) == 0 {
		return "measure-missing", "no measure record with a pair at " + record + " — commit the task, then " + run
	}
	var last int64 // the newest commit touching the task's files, in any worktree
	for _, w := range worktrees {
		out, _ := capture(git(append([]string{"-C", w, "log", "-1", "--format=%ct", "--"}, task.files...)...))
		if ct, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64); err == nil {
			last = max(last, ct)
		}
	}
	if st.ModTime().Unix() < last {
		return "measure-stale", "the measure record " + record + " is older than the task's last source commit — " + run
	}
	var flagged []string
	for _, p := range m.Pairs {
		n := len(p.Bands) + len(p.Seams) + len(p.Over)
		if p.Exits.Compose == nil || *p.Exits.Compose != 0 {
			n++
		}
		if n > 0 {
			flagged = append(flagged, fmt.Sprintf("%s %d", p.Frame, n))
		}
	}
	if len(flagged) > 0 {
		return "measure-not-clean", "flagged rows remain (" + strings.Join(flagged, ", ") + ") in " + record + " — fix them at their source, commit, then " + run
	}
	return "", "measured clean after its last source commit: " + record
}
