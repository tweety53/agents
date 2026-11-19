package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// flow-guard fix-slot-judge <worktree> <since-sha> <fix-instructions-file> <decision.json>
// — pick which of the decided review slots must review one fix run's delta.
//
// The decided slots are every `slots` entry of decision.json's `panel`
// bundles, in bundle order, each once. `primary` always runs and is never
// put to the judge. Every other slot is put to a headless `claude -p` --
// pjRun, on the fix-run-judge role's first base pair read from `flow policy
// get` -- whose system prompt lists each slot with its line from the
// reviewer catalog of skills/flow/decider-prompt.md under
// FLOW_GUARD_REPO_ROOT (left out when unreadable), and whose stdin is the fix
// instructions, `git diff --stat <since-sha>` and `git diff <since-sha>` in
// <worktree> (the working tree, as write-panel-diff late-fix writes it), cut
// at 200 KB as pair-judge cuts its digest.
//
// Prints one line per decided slot, `<slot>: run|skip — <reason>`, in the
// decided order, and saves the same lines, overwriting, beside the
// instructions file as `<file less its extension>.slots`; a failed save is
// one stderr line and leaves the exit 0. A decision whose only slot is
// primary runs no judge.
//
// The fallback: every decided slot prints `run — judge fallback`, and one
// stderr line names the cause -- no fix-run-judge pair, the command missing
// or failing, the deadline, an answer that is not JSON, a line naming a slot
// that was not put to it, or a put slot left unanswered.
//
// Exit 0 on the printed lines, whatever the judge did. Exit 2, nothing on
// stdout: usage (reason usage), an unreadable instructions file
// (instructions-unreadable), or a decision.json that is unreadable or names
// no slot (decision-unreadable).
//
// Go, in flow-guard (stats/internal/guard/fixslotjudge.go).
func init() { Registry["fix-slot-judge"] = fixSlotJudge }

const fsjPrompt = `You decide which code reviewers must review one fix made to an already-reviewed branch.

The user message is the fix instructions, then the fix's diff stat and diff. The primary reviewer always reads the fix. Each reviewer below runs only when this fix gives its lens something to find; a reviewer whose lens the fix does not touch is skipped.

The reviewers, one per line:
%s
Answer with exactly one line per reviewer above and nothing else:
<id>: run|skip — <one-line reason>
`

var (
	fsjAnswer  = regexp.MustCompile(`^\s*` + "`?" + `([a-z][a-z0-9-]*)` + "`?" + `:\s*(run|skip)\b\s*[—-]*\s*(.*)$`)
	fsjCatalog = regexp.MustCompile("(?m)^\\| `([a-z][a-z0-9-]*)` \\| (.+?) \\|\\s*$")
)

func fixSlotJudge(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 4 {
		fmt.Fprintln(stderr, "usage: flow-guard fix-slot-judge <worktree> <since-sha> <fix-instructions-file> <decision.json>")
		Reason(env, "usage")
		return 2
	}
	wt, since, instrFile, decFile := args[0], args[1], args[2], args[3]
	instr, err := os.ReadFile(instrFile)
	if err != nil {
		fmt.Fprintf(stderr, "fix-slot-judge: %v\n", err)
		Reason(env, "instructions-unreadable")
		return 2
	}
	slots, err := fsjSlots(decFile)
	if err != nil {
		fmt.Fprintf(stderr, "fix-slot-judge: %s: %v\n", decFile, err)
		Reason(env, "decision-unreadable")
		return 2
	}
	put := slices.DeleteFunc(slices.Clone(slots), func(s string) bool { return s == "primary" })
	verdict := map[string]string{"primary": "run — always runs"}
	if len(put) > 0 {
		answers, cause := fsjAsk(env, wt, since, instr, put)
		if cause != "" {
			fmt.Fprintf(stderr, "fix-slot-judge: judge fallback — %s\n", strings.TrimSpace(pjSpace.ReplaceAllString(cause, " ")))
			for _, s := range put {
				answers[s] = "run — judge fallback"
			}
		}
		for s, a := range answers {
			verdict[s] = a
		}
	}
	var out strings.Builder
	for _, s := range slots {
		fmt.Fprintf(&out, "%s: %s\n", s, verdict[s])
	}
	file := strings.TrimSuffix(instrFile, filepath.Ext(instrFile)) + ".slots"
	if err := os.WriteFile(file, []byte(out.String()), 0o644); err != nil {
		fmt.Fprintf(stderr, "fix-slot-judge: could not save %s: %v\n", file, err)
	}
	fmt.Fprint(stdout, out.String())
	return 0
}

// fsjSlots is decision.json's decided slots, bundle order, each once.
func fsjSlots(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d struct {
		Panel []struct {
			Slots []string `json:"slots"`
		} `json:"panel"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	var slots []string
	for _, b := range d.Panel {
		for _, s := range b.Slots {
			if !slices.Contains(slots, s) {
				slots = append(slots, s)
			}
		}
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("its panel names no slot")
	}
	return slots, nil
}

// fsjAsk puts the slots to the judge: each slot's `run|skip — <reason>`, or
// a non-empty cause why the answer cannot be used.
func fsjAsk(env Env, wt, since string, instr []byte, put []string) (map[string]string, string) {
	answers := map[string]string{}
	rows, err := policyTable(env)
	if err != nil {
		return answers, err.Error()
	}
	git := func(a ...string) (string, error) {
		b, err := exec.Command("git", append([]string{"-C", wt}, a...)...).Output()
		return string(b), err
	}
	stat, err := git("diff", "--no-renames", "--stat", since)
	if err != nil {
		return answers, fmt.Sprintf("git diff --stat %s: %v", since, err)
	}
	diff, err := git("diff", "--no-renames", since)
	if err != nil {
		return answers, fmt.Sprintf("git diff %s: %v", since, err)
	}
	digest := []byte("## Fix instructions\n\n" + string(instr) + "\n\n## Diff stat\n\n" + stat + "\n## Diff\n\n" + diff)
	if len(digest) > pjDigestCap {
		digest = append(digest[:pjDigestCap:pjDigestCap], pjTruncated...)
	}
	catalog := map[string]string{}
	if root := env.Getenv("FLOW_GUARD_REPO_ROOT"); root != "" {
		if b, err := os.ReadFile(filepath.Join(root, "skills", "flow", "decider-prompt.md")); err == nil {
			for _, m := range fsjCatalog.FindAllStringSubmatch(string(b), -1) {
				catalog[m[1]] = m[2]
			}
		}
	}
	var list strings.Builder
	for _, s := range put {
		list.WriteString("- " + s)
		if c := catalog[s]; c != "" {
			list.WriteString(" — " + c)
		}
		list.WriteString("\n")
	}
	result, spend, cause := pjRun(env, rows, "fix-run-judge", fmt.Sprintf(fsjPrompt, list.String()), digest)
	if cause != "" {
		return answers, cause
	}
	for _, l := range strings.Split(result, "\n") {
		m := fsjAnswer.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if m[1] == "primary" {
			continue
		}
		if !slices.Contains(put, m[1]) {
			return map[string]string{}, fmt.Sprintf("judge answered for %s, which was not put to it", m[1]) + spend
		}
		answers[m[1]] = m[2] + " — " + strings.TrimSpace(pjSpace.ReplaceAllString(m[3], " "))
	}
	for _, s := range put {
		if answers[s] == "" {
			return map[string]string{}, fmt.Sprintf("judge left %s unanswered: %s", s, strings.TrimSpace(pjSpace.ReplaceAllString(result, " "))) + spend
		}
	}
	return answers, ""
}
