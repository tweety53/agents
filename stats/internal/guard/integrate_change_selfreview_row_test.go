package guard

import (
	"github.com/tweety53/agents/stats/internal/smoke"
	"testing"
)

// The self-review pass's land marker is flowd's completed flow.self-review
// stage run, never a committed report (KAN-1073): the fixture's flow stub
// answers `flow stage completed` from $WORK/sr-completed, and
// $WORK/sr-unreachable stands in for a flowd that cannot answer.

// TestIntegrateChangeLandStopsWithoutACompletedSelfReview pins land's
// check: no completed run stops land before any push, even with a report
// committed the old way; flowd unreachable is exit 2; a completed run lands.
func TestIntegrateChangeLandStopsWithoutACompletedSelfReview(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"land_on_stage_row", `# --- land_on_stage_row ---
fixture
to_land "land_on_stage_row: " no-report
mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
sut land "$WORK/a" demo
check "land_on_stage_row: a committed report without a completed run stops land" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: self-review-report — flow.self-review has not completed for demo"'
check "land_on_stage_row: nothing was landed" \
  '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ]'
touch "$WORK/sr-unreachable"
sut land "$WORK/a" demo
check "land_on_stage_row: flowd unreachable is exit 2" '[ "$RC" -eq 2 ] && grep -q "flowd cannot say whether flow.self-review completed for demo" "$WORK/stderr"'
rm "$WORK/sr-unreachable"
touch "$WORK/sr-completed"
sut land "$WORK/a" demo
check "land_on_stage_row: a completed run lands" '[ "$RC" -eq 0 ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`}})
}

// TestIntegrateChangeSkipsThePassWhenTheStageCompleted pins commit's NEXT
// line: a completed flow.self-review run skips the pass; none runs it.
func TestIntegrateChangeSkipsThePassWhenTheStageCompleted(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{{"skip_on_stage_row", `# --- skip_on_stage_row ---
fixture
to_land "skip_on_stage_row: " no-report
check "skip_on_stage_row: no completed run asks for the pass" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run the self-review pass .*end flow.self-review completed"'
touch "$WORK/sr-completed"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "skip_on_stage_row: a completed run skips the pass" '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: flow.self-review has already completed — skip the pass"'
check "skip_on_stage_row: the skip opens no flow.self-review run" '[ "$(grep -c "stage begin .*flow.self-review" "$WORK/stages.log")" -eq 1 ]'`}})
}
