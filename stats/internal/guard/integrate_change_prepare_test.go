package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// The prepare and commit cases of integrate-change's former bash harness,
// run as integrate_change_test.go runs the happy path.
func TestIntegrateChangePrepareCommit(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"foreign_staged", `# --- foreign_staged ---
fixture
printf 'stray\n' >"$WORK/b/stray.txt"
g -C "$WORK/b" add stray.txt
sut prepare "$WORK/a" demo
check "staged work in a main checkout stops prepare with exit 1" '[ "$RC" -eq 1 ]'
check "the stop is STOP: main-checkout" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: main-checkout"'
check "the guard's own DRIFT-DIRTY line is relayed" 'printf "%s" "$OUT" | grep -q "^DRIFT-DIRTY: $WORK/b"'
check "the preflight begins no mark" '! grep -q "stage begin" "$WORK/stages.log"'`},
		{"no_verdict", `# --- no_verdict ---
fixture
: >"$WORK/fail-findings"
sut prepare "$WORK/a" demo
check "no_verdict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "no_verdict: the stop is STOP: no-verdict" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: no-verdict — check-unfinished-work"'
check "no_verdict: the land mark closes stopped" 'tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.land -outcome stopped"'`},
		{"sync_conflict", `# --- sync_conflict ---
fixture
printf 'base side\n' >"$WORK/b/work-b.txt"
g -C "$WORK/b" add work-b.txt && g -C "$WORK/b" commit -qm "base: work-b" && g -C "$WORK/b" push -q origin main
sut prepare "$WORK/a" demo
check "sync_conflict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "sync_conflict: the stop line names the --resume call and what follows it" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: sync-conflict — .*flow-guard sync-onto-base --resume., the whole ## lint and ## test lists, then re-run prepare$"'`},
		{"rebase_lint", `# --- rebase_lint ---
# b's base gains a .flow/project.md whose ## lint list fails and nothing this
# change touches: the rebase is clean, and the rebased tree's lint list runs.
fixture
mkdir -p "$WORK/b/.flow"
printf '# project\n\n## lint\n\n` + "`" + `` + "`" + `` + "`" + `bash\necho lint-own-output; exit 1\n` + "`" + `` + "`" + `` + "`" + `\n' >"$WORK/b/.flow/project.md"
g -C "$WORK/b" add .flow/project.md && g -C "$WORK/b" commit -qm "base: lint list" && g -C "$WORK/b" push -q origin main
sut prepare "$WORK/a" demo
check "rebase_lint: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "rebase_lint: the rebase was clean" 'printf "%s" "$OUT" | grep -q "^REBASED: $WORK/b-worktrees/demo"'
check "rebase_lint: the stop is STOP: rebase-lint" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — the ## lint list failed on $WORK/b-worktrees/demo"'
check "rebase_lint: the list's own output is relayed" 'printf "%s" "$OUT" | grep -q "lint-own-output"'`},
		{"rebase_lint_env", `# --- rebase_lint_env: the list runs under prepare-workspace.sh's exports ---
fixture
mkdir -p "$WORK/b/.flow"
printf '# project\n\n## lint\n\n` + "`" + `` + "`" + `` + "`" + `bash\necho "lint-env:${DEMO_ADDR:-unset}"; exit 1\n` + "`" + `` + "`" + `` + "`" + `\n\n## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|----------|----------|---------|----------------|\n| ` + "`" + `url` + "`" + ` | ` + "`" + `DEMO_ADDR` + "`" + ` | ` + "`" + `http://127.0.0.1:1` + "`" + ` | ` + "`" + `http://127.0.0.1:2` + "`" + ` |\n' >"$WORK/b/.flow/project.md"
g -C "$WORK/b" add .flow/project.md && g -C "$WORK/b" commit -qm "base: lint list" && g -C "$WORK/b" push -q origin main
sut prepare "$WORK/a" demo
check "rebase_lint_env: the list saw the workspace's exported variable" 'printf "%s" "$OUT" | grep -q "lint-env:http://127.0.0.1:2"'`},
		{"rebase_lint_rewrite", `# --- rebase_lint_rewrite: a passing list that rewrote tracked files stops ---
fixture
mkdir -p "$WORK/b/.flow"
printf '# project\n\n## lint\n\n` + "`" + `` + "`" + `` + "`" + `bash\necho reformatted >>.flow/project.md\n` + "`" + `` + "`" + `` + "`" + `\n' >"$WORK/b/.flow/project.md"
g -C "$WORK/b" add .flow/project.md && g -C "$WORK/b" commit -qm "base: lint list" && g -C "$WORK/b" push -q origin main
sut prepare "$WORK/a" demo
check "rebase_lint_rewrite: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "rebase_lint_rewrite: the stop names the rewrite" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — the ## lint list rewrote tracked files on $WORK/b-worktrees/demo"'`},
		{"rebase_lint_dirty", `# --- rebase_lint_dirty: an uncommitted plan file the rebase never touches
# is no rewrite of the list's (kan-1010's self-review) ---
fixture
printf 'Correction: narrative.\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/notes.md"
printf '\n## lint\n\n` + "`" + `` + "`" + `` + "`" + `bash\ntrue\n` + "`" + `` + "`" + `` + "`" + `\n' >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" add .flow/project.md && g -C "$WORK/a" commit -qm "base: lint list" && g -C "$WORK/a" push -q origin main
sut prepare "$WORK/a" demo
check "rebase_lint_dirty: the rebase ran" 'printf "%s" "$OUT" | grep -q "^REBASED: $WORK/a-worktrees/demo"'
check "rebase_lint_dirty: the list passed" 'printf "%s" "$OUT" | grep -q "^REBASE-LINT-PASSED: $WORK/a-worktrees/demo"'
check "rebase_lint_dirty: no rebase-lint stop" '! printf "%s" "$OUT" | grep -q "^STOP: rebase-lint"'
check "rebase_lint_dirty: the edit is still uncommitted" 'grep -q "^Correction: narrative.$" "$WORK/a-worktrees/demo/spectre/changes/demo/notes.md" && [ -n "$(g -C "$WORK/a-worktrees/demo" status --porcelain --ignored -- spectre/changes/demo/notes.md)" ]'`},
		{"nested_link", `# --- nested_link ---
# A link.md outside the change's own folder is implementation: b's base
# carries one; the branch adds a nested one.
fixture
mkdir -p "$WORK/b/spectre/changes/old"
printf 'old\n' >"$WORK/b/spectre/changes/old/link.md"
g -C "$WORK/b" add -f -A && g -C "$WORK/b" commit -qm "base: old link" && g -C "$WORK/b" push -q origin main
g -C "$WORK/b-worktrees/demo" reset -q --hard main
mkdir -p "$WORK/b-worktrees/demo/spectre/changes/archive/y"
printf 'y\n' >"$WORK/b-worktrees/demo/spectre/changes/archive/y/link.md"
g -C "$WORK/b-worktrees/demo" add -f -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): nested link"
sut prepare "$WORK/a" demo
check "nested_link: prepare's NEXT asks a subject for b" \
  '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>"'
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b-worktrees/demo=feat(b): nest a link"
check "nested_link: commit exits 0" '[ "$RC" -eq 0 ]'
check "nested_link: b's implementation commit carries the nested link.md" \
  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/y/link.md)" = "feat(b): nest a link" ]'`},
		{"archive_deletion", `# --- archive_deletion ---
# commitTwo plans only the change's own folders, so an archive deletion is
# implementation and needs a subject (kan-1072 self-review, F1).
fixture
mkdir -p "$WORK/b/spectre/changes/archive/old"
printf 'old\n' >"$WORK/b/spectre/changes/archive/old/tasks.md"
g -C "$WORK/b" add -f -A && g -C "$WORK/b" commit -qm "base: old archive" && g -C "$WORK/b" push -q origin main
g -C "$WORK/b-worktrees/demo" reset -q --hard main
g -C "$WORK/b-worktrees/demo" rm -q -r spectre/changes/archive && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): remove the change archive"
sut prepare "$WORK/a" demo
check "archive_deletion: prepare's NEXT asks a subject for b" \
  '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>"'`},
		{"fix_folder_deletion", `# --- fix_folder_deletion ---
# commitTwo excludes only the fix folders on disk, so deleting one is
# implementation and needs a subject (kan-1072 self-review, M1).
fixture
mkdir -p "$WORK/b/spectre/changes/demo-fix-1"
printf 'fix\n' >"$WORK/b/spectre/changes/demo-fix-1/tasks.md"
g -C "$WORK/b" add -f -A && g -C "$WORK/b" commit -qm "base: fix folder" && g -C "$WORK/b" push -q origin main
g -C "$WORK/b-worktrees/demo" reset -q --hard main
g -C "$WORK/b-worktrees/demo" rm -q -r spectre/changes/demo-fix-1 && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): drop demo-fix-1"
sut prepare "$WORK/a" demo
check "fix_folder_deletion: prepare's NEXT asks a subject for b" \
  '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>"'`},
		{"mixed_roots", `# --- mixed_roots ---
fixture
mkdir -p "$WORK/b-worktrees/demo/openspec/changes/demo"
printf 'notes\n' >"$WORK/b-worktrees/demo/openspec/changes/demo/notes.md"
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(openspec): notes"
sut prepare "$WORK/a" demo
check "mixed_roots: prepare exits 0" '[ "$RC" -eq 0 ]'
check "mixed_roots: b's diffstat lists no openspec/changes/ path" '! printf "%s" "$OUT" | grep -q "openspec/changes/"'`},
		{"planning_only", `# --- planning_only ---
fixture
g -C "$WORK/b-worktrees/demo" reset -q --hard main
mkdir -p "$WORK/b-worktrees/demo/spectre/changes/demo"
printf 'notes\n' >"$WORK/b-worktrees/demo/spectre/changes/demo/notes.md"
sut prepare "$WORK/a" demo
check "planning_only: prepare exits 0" '[ "$RC" -eq 0 ]'
check "planning_only: prepare's NEXT asks a subject for a and none for b" \
  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo=<impl subject>" && ! printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo="'
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work"
check "planning_only: commit with no subject for b exits 0" '[ "$RC" -eq 0 ]'
check "planning_only: b's branch is the base, its change directory uncommitted" \
  '[ -z "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD)" ] && [ -f "$WORK/b-worktrees/demo/spectre/changes/demo/notes.md" ]'`},
		{"link_only_canon", `# --- link_only_canon ---
fixture
g -C "$WORK/a-worktrees/demo" reset -q --hard main
c="$WORK/a-worktrees/demo/spectre/changes/demo"
mkdir -p "$c"
printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
printf '# demo\n\n## Merge order\n\n1. ` + "`" + `b` + "`" + `\n2. ` + "`" + `.` + "`" + `\n' >"$c/link.md"
sut prepare "$WORK/a" demo
check "link_only_canon: prepare exits 0" '[ "$RC" -eq 0 ]'
check "link_only_canon: prepare's NEXT asks a subject for b and none for a" \
  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>" && ! printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo="'
printf '\n## integrate run\n\nnarrative\n' >>"$c/narrative.md"
sut commit "$WORK/a" demo "$WORK/b-worktrees/demo=feat(b): do the work"
check "link_only_canon: commit with no subject for a exits 0" '[ "$RC" -eq 0 ]'
check "link_only_canon: a's branch is the base, its change directory uncommitted" \
  '[ -z "$(git -C "$WORK/a-worktrees/demo" log --format=%s main..HEAD)" ] && [ -f "$c/link.md" ]'`},
		{"commit_stop_closes_land", `# --- commit_stop_closes_land ---
fixture
sut prepare "$WORK/a" demo
sut commit "$WORK/a" demo
check "commit_stop_closes_land: commit stops with exit 1" '[ "$RC" -eq 1 ]'
check "commit_stop_closes_land: the land mark closes stopped" 'tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.land -outcome stopped"'
sut commit "$WORK/a" demo
check "commit_stop_closes_land: the retry reopens flow.land" '[ "$(grep -c "stage begin .*-stage flow.land" "$WORK/stages.log")" = 2 ]'`},
		{"stale_land_stopped", `# --- stale_land_stopped: a stopped prepare's marker does not outlive the next prepare ---
fixture
: >"$WORK/fail-findings"
sut prepare "$WORK/a" demo
rm "$WORK/fail-findings"
sut prepare "$WORK/a" demo
sut commit "$WORK/a" demo
check "stale_land_stopped: no second flow.land run opens after the second prepare" '[ "$(grep -c "stage begin .*-stage flow.land" "$WORK/stages.log")" = 2 ]'`},
		{"commit_pushes", `# --- commit_pushes ---
fixture
for r in a b; do g -C "$WORK/$r-worktrees/demo" push -q -u origin spectre/demo; done
to_land "commit_pushes: " no-report
check "commit_pushes: each remote branch is the reshaped HEAD" \
  '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)" ] && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)" ]'`},
		{"reshape_rerun", `# --- reshape_rerun ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work"
check "reshape_rerun: commit with no subject for b stops after the reshape" \
  'printf "%s" "$OUT" | grep -q "^STOP: commit-two" && [ "$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)" = "$(git -C "$WORK/b" rev-parse main)" ]'
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "reshape_rerun: the re-run commit exits 0" '[ "$RC" -eq 0 ]'
check "reshape_rerun: no pre-reshape pin is made" \
  '! git -C "$WORK/a" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null && ! git --git-dir="$WORK/b-origin.git" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null'`},
		{"base_mismatch", `# --- base_mismatch ---
fixture
for r in a b; do g -C "$WORK/$r" push -q origin main:refs/heads/feature/x; done
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "base_mismatch: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "base_mismatch: the stop is STOP: base-mismatch naming b's worktree and feature/x" \
  'printf "%s" "$OUT" | tail -1 | grep "^STOP: base-mismatch — " | grep -F "$WORK/b-worktrees/demo" | grep -q feature/x'
check "base_mismatch: the preflight begins no mark" '! grep -q "stage begin" "$WORK/stages.log"'
g -C "$WORK/b-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "base_mismatch: b recording feature/x too passes the gate" '! printf "%s" "$OUT" | grep -q "STOP: base-mismatch" && grep -q "stage begin .*-stage flow.land" "$WORK/stages.log"'
fixture
g -C "$WORK/a" push -q origin main:refs/heads/feature/x
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "base_mismatch: feature/x on a's origin only passes the gate" '! printf "%s" "$OUT" | grep -q "STOP: base-mismatch" && grep -q "stage begin .*-stage flow.land" "$WORK/stages.log"'
check "base_mismatch: feature/x on a's origin only is created on b's origin from main" \
  'printf "%s" "$OUT" | grep -q "^BASE-CREATED: origin/feature/x from main" && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse feature/x)" = "$(git --git-dir="$WORK/b-origin.git" rev-parse main)" ]'
check "base_mismatch: b is re-pointed onto feature/x" \
  '[ "$(git -C "$WORK/b-worktrees/demo" config branch.spectre/demo.flowBase)" = feature/x ] && PATH="$WORK/bin:$PATH" flow state get -C "$WORK/a" demo | grep -qF "$(git -C "$WORK/b-worktrees/demo" merge-base HEAD origin/feature/x)"'
fixture
g -C "$WORK/a" push -q origin main:refs/heads/feature/x
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
g -C "$WORK/b-worktrees/demo" config branch.spectre/demo.flowBase main
sut prepare "$WORK/a" demo
check "base_mismatch: b recording another base stops" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: base-mismatch — "'`},
		{"outstanding", `# --- outstanding ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
for flag in --plan-message --outstanding; do
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" "$flag" "Task 3 is unticked"
  check "outstanding: $flag is refused" '[ "$RC" -ne 0 ] && grep -q "unknown argument .$flag." "$WORK/stderr"'
done`},
		{"drift_switch", `# --- drift_switch: a clean main checkout left on another branch is switched
# back to its default branch, the branch keeping its commits ---
fixture
g -C "$WORK/a" checkout -q -b side
printf 'side\n' >"$WORK/a/side.txt"
g -C "$WORK/a" add side.txt && g -C "$WORK/a" commit -qm "side work"
side="$(g -C "$WORK/a" rev-parse side)"
sut prepare "$WORK/a" demo
check "drift_switch: prepare exits 0" '[ "$RC" -eq 0 ]'
check "drift_switch: one DRIFT-SWITCHED line names side, its sha and its one commit" \
  '[ "$(printf "%s\n" "$OUT" | grep -c "^DRIFT-SWITCHED:")" -eq 1 ] && printf "%s\n" "$OUT" | grep -qxF "DRIFT-SWITCHED: $WORK/a — side -> main, side kept at $(git -C "$WORK/a" rev-parse --short "$side") (1 commits ahead of main)"'
check "drift_switch: a's main checkout is on main" '[ "$(git -C "$WORK/a" branch --show-current)" = main ]'
check "drift_switch: side keeps its commit" '[ "$(git -C "$WORK/a" rev-parse side)" = "$side" ]'
check "drift_switch: the re-read prints DRIFT-CLEAN" 'printf "%s\n" "$OUT" | grep -qxF "DRIFT-CLEAN: $WORK/a — main"'
check "drift_switch: no main-checkout stop and the preflight ran" '! printf "%s" "$OUT" | grep -q "STOP: main-checkout" && grep -q "stage begin .*-stage flow.land" "$WORK/stages.log"'`},
		{"drift_switch_stops", `# --- drift_switch_stops: each repository is judged on its own; a dirty or
# detached checkout is never switched ---
fixture
g -C "$WORK/a" checkout -q -b side
g -C "$WORK/b" checkout -q -b other
printf 'dirty\n' >"$WORK/b/base.txt"
sut prepare "$WORK/a" demo
check "drift_switch_stops: a dirty b stops prepare at STOP: main-checkout" \
  '[ "$RC" -eq 1 ] && printf "%s\n" "$OUT" | tail -1 | grep -q "^STOP: main-checkout"'
check "drift_switch_stops: a is switched though b stops the run" \
  '[ "$(git -C "$WORK/a" branch --show-current)" = main ] && printf "%s\n" "$OUT" | grep -qF "DRIFT-SWITCHED: $WORK/a "'
check "drift_switch_stops: dirty b is not switched" \
  '[ "$(git -C "$WORK/b" branch --show-current)" = other ] && [ "$(cat "$WORK/b/base.txt")" = dirty ] && ! printf "%s\n" "$OUT" | grep -qF "DRIFT-SWITCHED: $WORK/b "'
fixture
g -C "$WORK/a" checkout -q --detach
sut prepare "$WORK/a" demo
check "drift_switch_stops: a detached HEAD stops and stays detached" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | grep -q "STOP: main-checkout" && [ -z "$(git -C "$WORK/a" branch --show-current)" ] && ! printf "%s" "$OUT" | grep -q "DRIFT-SWITCHED"'`},
		{"drift_switch_refused", `# --- drift_switch_refused: a switch git refuses changes nothing and stops ---
fixture
g -C "$WORK/a" checkout -q -b side
g -C "$WORK/a" worktree add -q "$WORK/a-main" main
sut prepare "$WORK/a" demo
check "drift_switch_refused: a switch git refuses stops and leaves a on side" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | grep -q "STOP: main-checkout" && printf "%s\n" "$OUT" | grep -qxF "DRIFT-SWITCH-REFUSED: $WORK/a — git switch main failed; left on side" && [ "$(git -C "$WORK/a" branch --show-current)" = side ]'
fixture
printf 'tracked\n' >"$WORK/a/.env"
g -C "$WORK/a" add .env && g -C "$WORK/a" commit -qm "track .env" && g -C "$WORK/a" push -q origin main
g -C "$WORK/a" checkout -q -b side
g -C "$WORK/a" rm -q --cached .env && printf '.env\n' >"$WORK/a/.gitignore"
g -C "$WORK/a" add .gitignore && g -C "$WORK/a" commit -qm "ignore .env"
printf 'SECRET local\n' >"$WORK/a/.env"
sut prepare "$WORK/a" demo
check "drift_switch_refused: an ignored file main tracks is never overwritten" \
  '[ "$RC" -eq 1 ] && printf "%s\n" "$OUT" | grep -qF "DRIFT-SWITCH-REFUSED: $WORK/a " && [ "$(git -C "$WORK/a" branch --show-current)" = side ] && [ "$(cat "$WORK/a/.env")" = "SECRET local" ]'`},
	})
}

// TestIntegrateChangePrepareRerunReadsThePlanOnDisk: a re-run of prepare
// after run 1's commit finds the plan in the worktree, never in a commit --
// the change directory is still on disk, uncommitted, and the re-run is a
// first pass naming its live narrative.
func TestIntegrateChangePrepareRerunReadsThePlanOnDisk(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"rerun_on_disk", `# --- rerun_on_disk ---
fixture
to_land "rerun_on_disk: " no-report
A="$WORK/a-worktrees/demo"
check "rerun_on_disk: the change directory is on disk and untracked" \
  '[ -f "$A/spectre/changes/demo/tasks.md" ] && [ -z "$(git -C "$A" ls-files spectre/changes)" ]'
sut prepare "$WORK/a" demo
check "rerun_on_disk: prepare exits 0" '[ "$RC" -eq 0 ]'
check "rerun_on_disk: its NEXT is the commit call" 'printf "%s" "$OUT" | tail -1 | grep -qF "NEXT: flow-guard integrate-change commit "'
printf '\n## fix run\n\nfixed\n' >>"$A/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "rerun_on_disk: commit exits 0" '[ "$RC" -eq 0 ]'`},
		{"rerun_appended_task", `# --- rerun_appended_task ---
fixture
A="$WORK/a-worktrees/demo"
printf '**Tasks appended:** 1\n\n- [x] 1. Do the work\n- [x] 2. The fix\n' >"$A/spectre/changes/demo/tasks.md"
printf 'fix\n' >"$A/fix-a.txt"
g -C "$A" add fix-a.txt && g -C "$A" commit -qm "fix(a): the appended task" -m "Task-Id: 2"
echo '[{"key":"task-1-implementer","outcome":"completed","seq":1},{"key":"task-2-implementer","taskId":"2","outcome":"completed","seq":2},{"key":"panel-0-gated","role":"reviewer","outcome":"completed","seq":3},{"key":"verify","role":"verifier","outcome":"completed","seq":4,"startedAt":"2030-01-02T00:00:00Z"}]' >"$WORK/dispatches.json"
to_land "rerun_appended_task: " no-report
T="$(git -C "$A" log -1 --format='%(trailers:key=Task-Id,valueonly,separator=%x2C)' --grep='^feat(a): do the work$')"
check "rerun_appended_task: the implementation commit keeps the squashed Task-Id trailers" '[ "$T" = "2,1" ]'
sut prepare "$WORK/a" demo
check "rerun_appended_task: a re-run of prepare passes the unfinished-work gate" '[ "$RC" -eq 0 ]'`},
		{"rerun_reshape_base", `# --- rerun_reshape_base ---
fixture
A="$WORK/a-worktrees/demo"
to_land "rerun_reshape_base: " no-report
split="$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
check "rerun_reshape_base: the reshape base is run 1's implementation commit" 'printf "%s" "$OUT" | grep -qF "WORKTREE: $A — base main, reshape base $split"'`},
	})
}
