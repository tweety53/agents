package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// The self-review pass runs after land and gates cleanup: its marker is
// flowd's completed flow.self-review stage run, never a committed report
// (KAN-1073). The fixture's flow stub answers `flow stage completed` from
// $WORK/sr-completed, and $WORK/sr-unreachable stands in for a flowd that
// cannot answer.

// TestCommitPrintsLandNext pins commit's NEXT line: land, and no
// flow.self-review run begun.
func TestCommitPrintsLandNext(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"commit_next_land", `# --- commit_next_land ---
fixture
to_land "commit_next_land: " no-report
check "commit_next_land: NEXT names land" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: flow-guard integrate-change land"'
check "commit_next_land: no self-review run begun" '! grep -q "flow.self-review" "$WORK/stages.log"'`}})
}

// TestLandRunsWithoutSelfReview pins that land on the pull request route
// pushes and records the PR with no completed flow.self-review run.
func TestLandRunsWithoutSelfReview(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"land_pr_without_self_review", `# --- land_pr_without_self_review ---
fixture
to_land "land_pr: " no-report
sut land "$WORK/a" demo --route "pull request" --pr-url "$WORK/b-worktrees/demo=https://forge.invalid/b/1" --pr-url "$WORK/a-worktrees/demo=https://forge.invalid/a/1"
check "land_pr: land exits 0 and records the PR" \
  '[ "$RC" -eq 0 ] && [ "$(jq -r "[.state, .prUrl] | join(\" \")" "$WORK/state.json")" = "IN_PROGRESS https://forge.invalid/a/1" ]'
check "land_pr: its NEXT names the self-review pass" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run the self-review pass .*waiting on the merge"'`}})
}

// TestLandMergeAndPushDoesNotCleanUp pins that merge and push lands, writes
// IN_PROGRESS and leaves the worktrees, naming the pass then cleanup.
func TestLandMergeAndPushDoesNotCleanUp(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"land_merge_keeps_worktree", `# --- land_merge_keeps_worktree ---
fixture
to_land "land_merge: " no-report
sut land "$WORK/a" demo
check "land_merge: land exits 0 with IN_PROGRESS" '[ "$RC" -eq 0 ] && [ "$(jq -r .state "$WORK/state.json")" = IN_PROGRESS ]'
check "land_merge: flowLandBase is recorded" '[ -n "$(git -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowLandBase)" ]'
check "land_merge: the worktrees stay" '[ -d "$WORK/a-worktrees/demo" ] && [ -d "$WORK/b-worktrees/demo" ]'
check "land_merge: NEXT names the pass then cleanup" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run the self-review pass .*then flow-guard integrate-change cleanup"'`}})
}

// TestReLandKeepsFirstLandBase pins that a second land leaves the recorded
// flowLandBase alone (it would otherwise become HEAD itself).
func TestReLandKeepsFirstLandBase(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"reland_keeps_base", `# --- reland_keeps_base ---
fixture
to_land "reland: " no-report
git -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowLandBase first-base
sut land "$WORK/a" demo
check "reland: flowLandBase kept" '[ "$(git -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowLandBase)" = first-base ]'`}})
}

// TestCleanupRefusesWithoutSelfReview pins cleanup's gate: no completed run
// is exit 1, flowd unreachable exit 2, and nothing is removed.
func TestCleanupRefusesWithoutSelfReview(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"cleanup_refuses", `# --- cleanup_refuses ---
fixture
to_land "cleanup_refuses: " no-report
sut land "$WORK/a" demo
sut cleanup "$WORK/a" demo
check "cleanup_refuses: no completed run stops cleanup" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: self-review-report — flow.self-review has not completed for demo"'
check "cleanup_refuses: nothing was removed" '[ -d "$WORK/a-worktrees/demo" ] && [ -d "$WORK/b-worktrees/demo" ]'
touch "$WORK/sr-unreachable"
sut cleanup "$WORK/a" demo
check "cleanup_refuses: flowd unreachable is exit 2" '[ "$RC" -eq 2 ] && grep -q "flowd cannot say whether flow.self-review completed for demo" "$WORK/stderr"'`}})
}

// TestCleanupRunsAfterSelfReview pins that a completed run lets cleanup
// reach its merge check and finish.
func TestCleanupRunsAfterSelfReview(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"cleanup_runs", `# --- cleanup_runs ---
fixture
to_land "cleanup_runs: " no-report
sut land "$WORK/a" demo
touch "$WORK/sr-completed"
sut cleanup "$WORK/a" demo
check "cleanup_runs: cleanup finishes" '[ "$RC" -eq 0 ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`}})
}

// TestLandPrintsNoInReviewTransition pins that no route's land output asks
// the parent for a Jira In Review transition.
func TestLandPrintsNoInReviewTransition(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"land_no_in_review", `# --- land_no_in_review ---
fixture
to_land "no_in_review: " no-report
for r in "merge and push" manual; do
  sut land "$WORK/a" demo --route "$r"
  check "no_in_review: $r prints no In Review" '[ "$RC" -eq 0 ] && ! printf "%s" "$OUT" | grep -q "to In Review"'
done`}})
}
