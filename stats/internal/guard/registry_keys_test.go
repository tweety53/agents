package guard

import (
	"maps"
	"slices"
	"testing"
)

// registryKeys is every guard name flow-guard answers. The guards live in
// guard's subpackages, each registering its own from init(); guard imports
// every one of them, so a subpackage this package stops importing — or a
// registration lost in a move — drops its names from guard.Registry and
// fails here.
var registryKeys = []string{
	"affected-specs", "await-detached", "await-rows", "break-and-prove", "build-slot-prompt",
	"check-base-behind", "check-dev-stack-fresh", "check-finish-preflight", "check-fix-revert",
	"check-frame-constants", "check-guard-calls", "check-loop-dispatch", "check-open-requests",
	"check-panel-recorded", "check-unfinished-work", "check-verify-env", "check-visual-trigger",
	"close-task", "compose-mockup-frames", "design-repo", "diff-groups", "end-dispatch-row", "hook",
	"implement-change", "integrate-change", "kickoff-worktree", "live-store-copy",
	"measure-visual-properties", "merge-panel-findings", "pair-judge", "parallel-groups",
	"pipeline-fix", "plan-check", "pre-existing-failure", "prepare-workspace", "project-get",
	"prune-candidates", "rebuild-full-app-suite-zip", "recapture", "recover-guard-incident",
	"refresh-main-checkout", "remove-change-worktrees", "render-frames", "resolve-base-branch",
	"resolve-visual-screenshots", "round-ceiling", "run-detached", "run-reproducer",
	"run-verify-list", "smoke-keep", "stop-worktree-processes", "sync-base-behind", "sync-onto-base",
	"test-cost", "throwaway-worktree", "unchanged", "verify-change", "visual-verify-pass",
	"workspace", "workspace-remove", "write-panel-diff",
}

func TestRegistryKeySet(t *testing.T) {
	t.Parallel()
	got := slices.Sorted(maps.Keys(Registry))
	if !slices.Equal(got, registryKeys) {
		t.Errorf("guard.Registry keys = %q\nwant %q", got, registryKeys)
	}
}
