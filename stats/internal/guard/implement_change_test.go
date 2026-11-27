package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// imHelpers drives implement-change on integrate-change's two-repository
// fixture, its spectre a stub logging each call with its directory:
// `validate` exits by `$WORK/validate-rc`, `list --json` prints
// `$WORK/list.json`, two of three tasks done by default.
const imHelpers = `im_fixture() {
  fixture
  export IC_GUARD=implement-change
  printf '#!/usr/bin/env bash\nW="${0%%/bin/*}"\necho "$PWD $*" >>"$W/spectre.log"\ncase "$1" in\n  validate) exit "$(cat "$W/validate-rc" 2>/dev/null || echo 0)" ;;\n  list) cat "$W/list.json" ;;\nesac\n' >"$WORK/bin/spectre"
  chmod +x "$WORK/bin/spectre"
  echo '{"changes":[{"id":"other","done":0,"total":4},{"id":"demo","done":2,"total":3}]}' >"$WORK/list.json"
}
`

// Cost: 2.1 s — five cases over the two-repository fixture, each running the sibling guards the driver sequences.
func TestImplementChangePrepare(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"tasks_left_is_execute", imHelpers + `im_fixture
sut prepare "$WORK/a" demo
check "exits 0" '[ "$RC" -eq 0 ]'
check "begins flow.implement" 'grep -q "stage begin .*-stage flow.implement" "$WORK/stages.log"'
check "ends no stage" '! grep -q "stage end" "$WORK/stages.log"'
check "validates in the canonical worktree" 'grep -qx "$WORK/a-worktrees/demo validate demo" "$WORK/spectre.log"'
check "prints the workspace id" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ID: demo-id"'
check "prints this change's tasks" 'printf "%s\n" "$OUT" | grep -qx "TASKS: 2/3"'
check "runs the base check" 'printf "%s\n" "$OUT" | grep -q "^BASE-CURRENT: "'
check "ends on NEXT: execute" '[ "$(printf "%s\n" "$OUT" | tail -1)" = "NEXT: execute" ]'`},
		{"every_task_done_is_review_panel", imHelpers + `im_fixture
echo '{"changes":[{"id":"demo","done":3,"total":3}]}' >"$WORK/list.json"
sut prepare "$WORK/a" demo
check "exits 0" '[ "$RC" -eq 0 ]'
check "ends flow.implement completed" 'grep -q "stage end .*-stage flow.implement -outcome completed" "$WORK/stages.log"'
check "ends on NEXT: review-panel" '[ "$(printf "%s\n" "$OUT" | tail -1)" = "NEXT: review-panel" ]'`},
		{"validate_findings_stop", imHelpers + `im_fixture
echo 1 >"$WORK/validate-rc"
sut prepare "$WORK/a" demo
check "exits 1" '[ "$RC" -eq 1 ]'
check "ends on STOP: validate" 'printf "%s\n" "$OUT" | tail -1 | grep -q "^STOP: validate — "'
check "closes flow.implement stopped" 'grep -q "stage end .*-stage flow.implement -outcome stopped -reason STOP: validate" "$WORK/stages.log"'
check "reads no task list" '! grep -q " list --json" "$WORK/spectre.log"'`},
		{"no_plan_stops", imHelpers + `im_fixture
echo '{"changes":[]}' >"$WORK/list.json"
sut prepare "$WORK/a" demo
check "exits 1" '[ "$RC" -eq 1 ]'
check "ends on STOP: no-plan" 'printf "%s\n" "$OUT" | tail -1 | grep -q "^STOP: no-plan — "'`},
		{"missing_name_is_usage", imHelpers + `im_fixture
sut prepare "$WORK/a"
check "exits 2" '[ "$RC" -eq 2 ]'
check "marks nothing" '[ ! -e "$WORK/stages.log" ]'`},
	})
}
