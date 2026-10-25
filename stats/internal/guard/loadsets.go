package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// load-sets [<repo>] [<rendered-global-CLAUDE.md>] — the static instruction bytes each /flow
// session loads from this repository (KAN-852). A report, never a verdict: it always exits 0.
//
// Sums the bytes of each session's load set as the flow files' load directives define it —
// planning, implementation and finish, each its own session (/clear between them) — plus the
// files cited at the point of use and the conditional loads. Tokens are approximated as
// bytes/4. It excludes the superpowers skills, MEMORY.md and the project's own code.
//
// <repo> defaults to FLOW_GUARD_REPO_ROOT, the checkout scripts/load-sets.sh lives in. The
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
// Ported from the prototype loadsets.sh in docs/prompt-audit-2026-09-29/tools.md.
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
	in := func(dir string, names ...string) []string {
		out := make([]string, len(names))
		for i, n := range names {
			out[i] = dir + n
		}
		return out
	}
	cat := func(sets ...[]string) []string {
		var out []string
		for _, s := range sets {
			out = append(out, s...)
		}
		return out
	}
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
	pd := cat(in(f, "brainstorm.md", "brainstorm-planner.md"), []string{r + "/skills/spectre-planning/SKILL.md"}, in(c, "jira-integration.md", "plan-provenance.md", "build-green.md"))
	pc := in(c, "worktree-resolution.md", "git-boundaries.md", "operator-prompts.md", "operator-prompts-auto-resolution.md")
	px := in(f, "withdrawal.md", "seeded-note.md", "resume.md")
	row("  phase files + load directives", pd...)
	row("  + cited at point of use", pc...)
	row("  TOTAL definite", cat(always, router, pd)...)
	row("  + conditional (flow-fix/flow-cost/flow-speed, seeded note, resume)", px...)
	row("  TOTAL incl. cited", cat(always, router, pd, pc)...)

	fmt.Fprintln(stdout, "--- implementation session (enters via Resuming at STARTED) ---")
	id := cat(in(f, "resume.md", "implement.md"), in(c, "worktree-resolution.md"), in(f, "review-panel.md", "verify-and-handoff.md"), in(c, "session-records.md", "git-boundaries.md"))
	ic := cat(in(c, "operator-prompts.md", "operator-prompts-auto-resolution.md", "project-configuration-standards.md", "model-policy.md", "known-bugs.md"),
		in(f, "primary-reviewer-prompt.md", "principles-reviewer-prompt.md", "failure-modes-reviewer-prompt.md"), in(c, "plan-amendment.md"))
	ix := cat(in(f, "document-fix.md", "cross-repo-worktrees.md", "sdd-dispatch.md", "gated-review-fix.md", "review-panel-late-fix.md", "review-panel-fix-round.md", "review-panel-optional-slots.md", "review-panel-experimental-slot.md"),
		in(c, "workspace-isolation.md"), in(f, "visual-verify.md", "visual-verify-tooling-analysis.md"),
		in(c, "git-boundaries-commit-chain.md", "project-configuration-visual.md", "guard-verdict-verification.md"))
	row("  phase files + load directives", id...)
	row("  + cited / reviewer templates", ic...)
	row("  + conditional (fix run, cross-repo, sdd, gated fix, fix round, optional slots, isolation, UI)", ix...)
	row("  TOTAL definite", cat(always, router, id)...)
	row("  TOTAL incl. cited", cat(always, router, id, ic)...)
	row("  TOTAL worst case", cat(always, router, id, ic, ix)...)

	fmt.Fprintln(stdout, "--- finish session (merge-and-push: run 1 chained into run 2) ---")
	fd := cat(in(f, "integrate.md"), in(c, "git-boundaries-commit-chain.md", "jira-integration.md", "jira-integration-finish.md"), in(f, "cleanup.md"))
	fc := in(c, "operator-prompts.md", "model-policy.md")
	fx := cat(in(c, "finish-contract-run1.md", "finish-contract-run2.md", "artifacts-registry.md"), in(f, "unfinished-work-gate.md", "sync-onto-base.md"),
		in(c, "finish-hand-fallbacks.md", "jira-followups.md", "jira-followups-join.md", "state-file.md", "project-configuration.md", "guard-verdict-verification.md"))
	row("  phase files + load directives", fd...)
	row("  + cited", fc...)
	row("  + conditional (finish contracts, registry, gate, base moved, hand fallbacks, follow-ups, state file, config, verdicts)", fx...)
	row("  TOTAL definite", cat(always, router, fd)...)
	row("  TOTAL incl. cited", cat(always, router, fd, fc)...)

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
