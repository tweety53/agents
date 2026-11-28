package guard

// loadSet is a conditional load set: the files a session loads only when condition holds.
type loadSet struct {
	condition string
	files     []string
}

// conditionalLoadSets are the planning, implementation and finish sessions' conditional load sets
// under repo root r, in that order — prune-candidates annotates a
// never-loaded file with its condition.
func conditionalLoadSets(r string) []loadSet {
	f, c := r+"/skills/flow/", r+"/skills/flow-contracts/"
	return []loadSet{
		{"flow-fix, resume", inDir(f, "resume.md")},
		{"fix run, cross-repo, sdd, live check, fix round, optional slots, isolation, pipeline defects, progress, UI",
			catSets(inDir(f, "fix.md", "fix-visual.md", "cross-repo-worktrees.md", "fixer-brief.md", "slot-brief.md", "review-panel-fix-round.md", "review-panel-optional-slots.md", "live-check.md"),
				inDir(c, "workspace-isolation.md", "pipeline-defects.md", "progress-visibility.md", "quiet-progress.md"), inDir(f, "visual-verify.md"),
				inDir(c, "git-boundaries-commit-chain.md", "project-configuration-visual.md", "guard-verdict-verification.md"))},
		{"finish contracts, registry, gate, base moved, own tooling, state file, config, verdicts",
			catSets(inDir(c, "finish-contract-run1.md", "finish-contract-run2.md", "artifacts-registry.md"), inDir(f, "unfinished-work-gate.md", "sync-onto-base.md", "sync-onto-base-visual.md", "integrate-own-tooling.md"),
				inDir(c, "state-file.md", "project-configuration.md", "guard-verdict-verification.md"))},
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
