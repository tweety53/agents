package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/load-sets.sh shim (KAN-1058).
//
// <repo> defaults to this script's repository, passed on as FLOW_GUARD_REPO_ROOT.

// load-sets [<repo>] [<rendered-global-CLAUDE.md>] — the static instruction bytes each /flow
// session loads from this repository (KAN-852). A report, never a verdict: it always exits 0.
//
// Sums the bytes of each session's load set as the flow files' load directives define it —
// planning, implementation and finish, each its own session (/clear between them) — plus the
// files cited at the point of use and the conditional loads. Tokens are approximated as
// bytes/4. It excludes the superpowers skills, MEMORY.md and the project's own code.
//
// <repo> defaults to FLOW_GUARD_REPO_ROOT, the checkout scripts/flow-guard exports. The
// second argument defaults to ~/.claude/CLAUDE.md, the always-on block setup.sh installs;
// render a sandboxed one with
//
//	SB="$(mktemp -d)"; HOME="$SB" ./setup.sh global   # then pass "$SB/.claude/CLAUDE.md"
//
// The lists below are the load sets, by hand: edit them when a change moves what a session
// loads. A file that does not exist is printed as MISSING on stderr — once per row that sums
// it — and counted as zero bytes, so a stale list shows up in the report rather than aborting
// it.
//
// Ported from the prototype loadsets.sh.
func init() { Registry["load-sets"] = loadSets }

func loadSets(args []string, env Env, stdout, stderr io.Writer) int {
	r := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if len(args) > 0 && args[0] != "" {
		r = args[0]
	}
	global := env.Getenv("HOME") + "/.claude/CLAUDE.md"
	if len(args) > 1 && args[1] != "" {
		global = args[1]
	}
	f, c := r+"/skills/flow/", r+"/skills/flow-contracts/"
	cond := conditionalLoadSets(r)
	row := func(name string, files ...string) {
		b := 0
		for _, p := range files {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				b += int(st.Size())
			} else {
				fmt.Fprintln(stderr, "load-sets: MISSING "+strings.TrimPrefix(p, r+"/"))
			}
		}
		fmt.Fprintf(stdout, "%-50s %7d bytes  ~%6d tok\n", name, b, b/4)
	}

	always := []string{global, r + "/CLAUDE.md"}
	router := []string{r + "/commands-claude/flow.md", f + "SKILL.md", c + "pipeline.md"}
	row("always-on (global block + project CLAUDE.md)", always...)
	row("router (command + SKILL.md + pipeline.md)", router...)

	fmt.Fprintln(stdout, "--- planning session (creating run) ---")
	pd := catSets(inDir(f, "brainstorm.md", "brainstorm-planner.md"), inDir(c, "jira-integration.md"))
	pc := inDir(c, "worktree-resolution.md", "git-boundaries.md", "operator-prompts.md", "operator-prompts-auto-resolution.md")
	px := cond[0].files
	row("  phase files + load directives", pd...)
	row("  + cited at point of use", pc...)
	row("  TOTAL definite", catSets(always, router, pd)...)
	row("  + conditional ("+cond[0].condition+")", px...)
	row("  TOTAL incl. cited", catSets(always, router, pd, pc)...)

	fmt.Fprintln(stdout, "--- implementation session (enters via Resuming at STARTED) ---")
	id := catSets(inDir(f, "resume.md", "implement.md"), inDir(c, "worktree-resolution.md"), inDir(f, "review-panel.md", "verify-and-handoff.md"), inDir(c, "session-records.md", "git-boundaries.md"))
	ic := catSets(inDir(c, "operator-prompts.md", "operator-prompts-auto-resolution.md", "project-configuration-standards.md", "model-policy.md", "known-bugs.md"),
		inDir(f, "primary-reviewer-prompt.md", "principles-reviewer-prompt.md", "failure-modes-reviewer-prompt.md"), inDir(c, "plan-amendment.md"))
	ix := cond[1].files
	row("  phase files + load directives", id...)
	row("  + cited / reviewer templates", ic...)
	row("  + conditional ("+cond[1].condition+")", ix...)
	row("  TOTAL definite", catSets(always, router, id)...)
	row("  TOTAL incl. cited", catSets(always, router, id, ic)...)
	row("  TOTAL worst case", catSets(always, router, id, ic, ix)...)

	fmt.Fprintln(stdout, "--- finish session (merge-and-push: run 1 chained into run 2) ---")
	fd := catSets(inDir(f, "integrate.md"), inDir(c, "git-boundaries-commit-chain.md", "jira-integration.md", "jira-integration-finish.md"), inDir(f, "cleanup.md"))
	fc := inDir(c, "operator-prompts.md", "model-policy.md")
	fx := cond[2].files
	row("  phase files + load directives", fd...)
	row("  + cited", fc...)
	row("  + conditional ("+cond[2].condition+")", fx...)
	row("  TOTAL definite", catSets(always, router, fd)...)
	row("  TOTAL incl. cited", catSets(always, router, fd, fc)...)

	fmt.Fprintln(stdout, "--- subagents (each also inherits whatever always-on context the harness gives it) ---")
	baseline := r + "/rules/agent-baseline.md"
	row("  reviewer: baseline + primary template", baseline, f+"primary-reviewer-prompt.md")
	row("  reviewer: baseline + principles + EP", baseline, f+"principles-reviewer-prompt.md", f+"engineering-principles.md")
	row("  reviewer: baseline + failure modes", baseline, f+"failure-modes-reviewer-prompt.md")
	row("  verifier: baseline + verifier steps (UI runs)", baseline, f+"visual-verify-verifier.md")
	row("  project-configuration.md (cited everywhere)", c+"project-configuration.md")
	row("  state-file.md (load before any state r/w)", c+"state-file.md")
	return 0
}

// loadSet is a conditional load set: the files a session loads only when condition holds.
type loadSet struct {
	condition string
	files     []string
}

// conditionalLoadSets are the planning, implementation and finish sessions' conditional load sets
// under repo root r, in that order — load-sets sums them and prune-candidates annotates a
// never-loaded file with its condition.
func conditionalLoadSets(r string) []loadSet {
	f, c := r+"/skills/flow/", r+"/skills/flow-contracts/"
	return []loadSet{
		{"flow-fix, resume", inDir(f, "resume.md")},
		{"fix run, cross-repo, sdd, live check, fix round, optional slots, isolation, pipeline defects, progress, UI",
			catSets(inDir(f, "fix.md", "fix-visual.md", "cross-repo-worktrees.md", "fixer-brief.md", "slot-brief.md", "review-panel-fix-round.md", "review-panel-optional-slots.md", "live-check.md"),
				inDir(c, "workspace-isolation.md", "pipeline-defects.md", "progress-visibility.md", "quiet-progress.md"), inDir(f, "visual-verify.md"),
				inDir(c, "git-boundaries-commit-chain.md", "project-configuration-visual.md", "guard-verdict-verification.md"))},
		{"finish contracts, registry, gate, base moved, stops, own tooling, follow-ups, state file, config, verdicts",
			catSets(inDir(c, "finish-contract-run1.md", "finish-contract-run2.md", "artifacts-registry.md"), inDir(f, "unfinished-work-gate.md", "sync-onto-base.md", "sync-onto-base-visual.md", "integrate-stops.md", "integrate-own-tooling.md"),
				inDir(c, "jira-followups.md", "jira-followups-join.md", "state-file.md", "project-configuration.md", "guard-verdict-verification.md"))},
	}
}

// inDir is each of names prefixed with dir.
func inDir(dir string, names ...string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = dir + n
	}
	return out
}

// catSets is sets concatenated.
func catSets(sets ...[]string) []string {
	var out []string
	for _, s := range sets {
		out = append(out, s...)
	}
	return out
}
