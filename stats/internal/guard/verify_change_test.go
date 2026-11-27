package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// vcHelpers drives verify-change on integrate-change's two-repository
// fixture: run-verify-list is a stub (FLOW_GUARD_RUN_VERIFY_LIST) logging
// each call and exiting by `$WORK/rvl-<list>-rc`, so no case waits on the
// heavy slot or runs a real list.
const vcHelpers = `vc_fixture() {
  fixture
  export IC_GUARD=verify-change FLOW_GUARD_RUN_VERIFY_LIST="$WORK/bin/rvl"
  printf '#!/usr/bin/env bash\nW="${0%%/bin/*}"\necho "$*" >>"$W/rvl.log"\necho "report $2"\nexit "$(cat "$W/rvl-$2-rc" 2>/dev/null || echo 0)"\n' >"$WORK/bin/rvl"
  chmod +x "$WORK/bin/rvl"
}
`

// Cost: 3.74 s — eight cases over the two-repository fixture, each running the sibling guards the driver sequences.
func TestVerifyChange(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"happy_path", vcHelpers + `vc_fixture
sut "$WORK/a" demo --model opus --effort flow-low
check "exits 0" '[ "$RC" -eq 0 ]'
check "begins flow.verify" 'grep -q "stage begin .*-stage flow.verify" "$WORK/stages.log"'
check "ends no stage" '! grep -q "stage end" "$WORK/stages.log"'
check "runs lint and test per worktree" '[ "$(wc -l <"$WORK/rvl.log")" -eq 4 ]'
check "suffixes the key with the repository name" 'grep -q "^$WORK/b-worktrees/demo test demo verify-b mf-test opus flow-low $WORK/a-worktrees/demo$" "$WORK/rvl.log"'
check "passes no canonical for the canonical worktree" 'grep -q "^$WORK/a-worktrees/demo lint demo verify-a mf-test opus flow-low$" "$WORK/rvl.log"'
check "prints each list's exit" '[ "$(printf "%s\n" "$OUT" | grep -c "^VERIFY: .* — exit 0$")" -eq 4 ]'
check "ends on NEXT: visual-verification" '[ "$(printf "%s\n" "$OUT" | tail -1)" = "NEXT: visual-verification" ]'`},
		{"red_list_stops_and_leaves_the_stage_open", vcHelpers + `vc_fixture
echo 1 >"$WORK/rvl-test-rc"
sut "$WORK/a" demo --model opus --effort flow-low
check "exits 1" '[ "$RC" -eq 1 ]'
check "still runs every worktree" '[ "$(wc -l <"$WORK/rvl.log")" -eq 4 ]'
check "ends on STOP: verify-failed" 'printf "%s\n" "$OUT" | tail -1 | grep -q "^STOP: verify-failed — "'
check "ends no stage" '! grep -q "stage end" "$WORK/stages.log"'`},
		{"undeclared_list_is_auto_detect", vcHelpers + `vc_fixture
echo 3 >"$WORK/rvl-lint-rc"
sut "$WORK/a" demo --model opus --effort flow-low
check "exits 0" '[ "$RC" -eq 0 ]'
check "names each auto-detect first" '[ "$(printf "%s\n" "$OUT" | tail -1)" = "NEXT: auto-detect $WORK/a-worktrees/demo lint; auto-detect $WORK/b-worktrees/demo lint; visual-verification" ]'`},
		{"worktree_with_nothing_outside_its_change_directory_is_skipped", vcHelpers + `vc_fixture
g -C "$WORK/b-worktrees/demo" reset -q --hard main
sut "$WORK/a" demo --model opus --effort flow-low
check "exits 0" '[ "$RC" -eq 0 ]'
check "prints the skip" 'printf "%s\n" "$OUT" | grep -q "^SKIPPED: $WORK/b-worktrees/demo — nothing outside its change directory$"'
check "runs only the other worktree" '! grep -q "b-worktrees" "$WORK/rvl.log"'`},
		{"live_check_section_comes_before_visual", vcHelpers + `vc_fixture
printf '# Design\n\n## Live check\n\nCurl the endpoint and read the count.\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/design.md"
sut "$WORK/a" demo --model opus --effort flow-low
check "names live-check" '[ "$(printf "%s\n" "$OUT" | tail -1)" = "NEXT: live-check; visual-verification" ]'`},
		{"no_runtime_live_check_is_skipped", vcHelpers + `vc_fixture
printf '# Design\n\n## Live check\n\nNone — a prose change.\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/design.md"
sut "$WORK/a" demo --model opus --effort flow-low
check "names only visual" '[ "$(printf "%s\n" "$OUT" | tail -1)" = "NEXT: visual-verification" ]'`},
		{"rerun_key_skips_the_base_check", vcHelpers + `vc_fixture
sut "$WORK/a" demo --model opus --effort flow-low --key verify-fix-1
check "exits 0" '[ "$RC" -eq 0 ]'
check "carries the key" 'grep -q " verify-fix-1-a " "$WORK/rvl.log"'
check "runs no base check" '! printf "%s\n" "$OUT" | grep -q "^BASE-"'`},
		{"missing_pair_is_usage", vcHelpers + `vc_fixture
sut "$WORK/a" demo
check "exits 2" '[ "$RC" -eq 2 ]'
check "runs nothing" '[ ! -e "$WORK/rvl.log" ]'`},
	})
}
