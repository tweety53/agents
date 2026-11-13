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
		{"narrative_format", `# --- narrative_format ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
nar="$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
printf '\n## integrate run\n\n- In-run pipeline fix: the renderer could not render an exp- slot\n' >>"$nar"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_format: an off-format in-run fix line stops commit with exit 1" '[ "$RC" -eq 1 ]'
check "narrative_format: the stop names the off-format line" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*the renderer could not render an exp- slot"'`},
		{"narrative_format_in_format", icNarrativePasses + `# --- narrative_format_in_format ---
narrative_passes "a line in the format" '- In-run pipeline fix: 1a2b3c4 — the renderer could not render an exp- slot (blast radius 2 files)'`},
		{"narrative_format_prose_mention", icNarrativePasses + `# --- narrative_format_prose_mention ---
narrative_passes "a prose mention of the token" 'The ` + "`" + `In-run pipeline fix:` + "`" + ` line was missing a sha.'`},
		{"narrative_format_trailing_space", icNarrativePasses + `# --- narrative_format_trailing_space ---
narrative_passes "a well-formed line with trailing whitespace" '- In-run pipeline fix: 1a2b3c4 — the defect (blast radius 2 files)   '`},
		{"narrative_format_deferred", icNarrativePasses + `# --- narrative_format_deferred ---
narrative_passes "a deferred big line" '- In-run pipeline fix: deferred — the defect (blast radius 60 files)'`},
		{"narrative_format_deferred_small", `# --- narrative_format_deferred_small ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\n- In-run pipeline fix: deferred — the hook denied a planner (blast radius 3 files)\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_format_deferred_small: a deferred fix under 60 files stops commit" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*defers an in-run fix of blast radius 3"'`},
		{"narrative_format_backticked_sha", icNarrativePasses + `# --- narrative_format_backticked_sha ---
narrative_passes "a backticked sha" '- In-run pipeline fix: ` + "`" + `1a2b3c4` + "`" + ` — the defect (blast radius 3 files).'`},
		{"narrative_runs", `# --- narrative_runs ---
fixture
echo 1 >"$WORK/fix-sessions"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_runs: a fix run with no section stops commit at STOP: narrative" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — NARRATIVE-RUNS-MISSING"'
printf '\n## 2026-10-19 — fix run\n\nfixed\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_runs: with the fix run's section commit no longer stops at narrative" \
  '! printf "%s" "$OUT" | grep -q "^STOP: narrative"'
rm -f "$WORK/fix-sessions"
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
touch "$WORK/fail-sessions"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_runs: a failed stage-sessions read names the failure in the stop" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — check-narrative-runs: flow stage sessions"'
rm -f "$WORK/fail-sessions"`},
		{"foreign_staged", `# --- foreign_staged ---
fixture
printf 'stray\n' >"$WORK/b/stray.txt"
g -C "$WORK/b" add stray.txt
sut prepare "$WORK/a" demo
check "staged work in a main checkout stops prepare with exit 1" '[ "$RC" -eq 1 ]'
check "the stop is STOP: foreign-staged" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: foreign-staged"'
check "the guard's own STAGED-FOREIGN line is relayed" 'printf "%s" "$OUT" | grep -q "^STAGED-FOREIGN: $WORK/b"'
check "no preflight ran" '! grep -q flow.preflight "$WORK/stages.log" 2>/dev/null'`},
		{"untasked", `# --- untasked ---
fixture
printf 'stray\n' >"$WORK/b-worktrees/demo/stray.txt"
g -C "$WORK/b-worktrees/demo" add stray.txt && g -C "$WORK/b-worktrees/demo" commit -qm "test(b): a stray edit"
sut prepare "$WORK/a" demo
check "untasked: prepare stops at STOP: unfinished-work" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: unfinished-work"'
check "untasked: the stop says to finish the work, asking nothing" 'printf "%s" "$OUT" | tail -1 | grep -qF "finish the unfinished work, asking nothing, then re-run prepare"'
check "untasked: the stray commit is listed" 'printf "%s" "$OUT" | grep -q "^  [0-9a-f]\{12\} test(b): a stray edit$"'
sut prepare "$WORK/a" demo --accept-outstanding
check "untasked: --accept-outstanding passes the gate" '[ "$RC" -eq 0 ]'`},
		{"archived_rerun", `# --- archived_rerun ---
fixture
to_land "archived_rerun: "
A="$WORK/a-worktrees/demo"
printf 'more\n' >"$WORK/a-worktrees/demo/more.txt"
g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): more" -m "Task-Id: 2"
sut prepare "$WORK/a" demo
check "archived_rerun: prepare passes the gate" '[ "$RC" -eq 0 ]'
check "archived_rerun: b's reshaped commit is not listed" '! printf "%s" "$OUT" | grep -q "feat(b): do the work"'`},
		{"no_verdict", `# --- no_verdict ---
fixture
: >"$WORK/fail-findings"
sut prepare "$WORK/a" demo --accept-outstanding
check "no_verdict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "no_verdict: the stop is STOP: no-verdict, not passed by --accept-outstanding" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: no-verdict — check-unfinished-work"'
check "no_verdict: the gate mark closes stopped" 'tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.unfinished-work-gate -outcome stopped"'`},
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
# commit-split plans only the change's own folders, so an archive deletion is
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
# commit-split excludes only the fix folders on disk, so deleting one is
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
		{"commit_pushes", `# --- commit_pushes ---
fixture
for r in a b; do g -C "$WORK/$r-worktrees/demo" push -q -u origin spectre/demo; done
to_land "commit_pushes: " no-report
check "commit_pushes: each remote branch is the reshaped HEAD" \
  '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)" ] && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)" ]'`},
		{"pre_reshape_pin", `# --- pre_reshape_pin ---
fixture
pa="$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)"; pb="$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work"
check "pre_reshape_pin: commit with no subject for b stops after the reshape" \
  'printf "%s" "$OUT" | grep -q "^STOP: commit-split" && [ "$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)" = "$(git -C "$WORK/b" rev-parse main)" ]'
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "pre_reshape_pin: the re-run commit exits 0" '[ "$RC" -eq 0 ]'
check "pre_reshape_pin: each pin is the pre-reshape HEAD, not the reshaped one" \
  '[ "$(git -C "$WORK/a" rev-parse refs/flow/pre-reshape/demo)" = "$pa" ] && [ "$(git -C "$WORK/b" rev-parse refs/flow/pre-reshape/demo)" = "$pb" ]'
check "pre_reshape_pin: b's Task-Id: 1 commit is found through its pin" \
  '[ "$(git -C "$WORK/b" log --format=%s --grep="^Task-Id: 1$" main..refs/flow/pre-reshape/demo)" = "feat(b): task 1" ]'
check "pre_reshape_pin: each origin holds its pin" \
  '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$pa" ] && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$pb" ]'
fixture
stray="$(g -C "$WORK/b" commit-tree "main^{tree}" -m stray)"
g -C "$WORK/b" push -q origin "$stray:refs/flow/pre-reshape/demo"
g -C "$WORK/b" update-ref refs/flow/pre-reshape/demo "$(g -C "$WORK/b-worktrees/demo" rev-parse HEAD)"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "pre_reshape_pin: a differing pin on origin stops commit, untouched" \
  'printf "%s" "$OUT" | grep -q "^STOP: pre-reshape" && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$stray" ]'`},
		{"rerun_pin", icRerunPinCase + `# --- rerun_pin ---
rerun_pin_case "rerun_pin: "`},
		{"rerun_pin_reclone", icRerunPinCase + `# --- rerun_pin_reclone ---
rerun_pin_case "rerun_pin reclone: " reclone`},
		{"base_mismatch", `# --- base_mismatch ---
fixture
for r in a b; do g -C "$WORK/$r" push -q origin main:refs/heads/feature/x; done
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "base_mismatch: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "base_mismatch: the stop is STOP: base-mismatch naming b's worktree and feature/x" \
  'printf "%s" "$OUT" | tail -1 | grep "^STOP: base-mismatch — " | grep -F "$WORK/b-worktrees/demo" | grep -q feature/x'
check "base_mismatch: no preflight ran" '! grep -q flow.preflight "$WORK/stages.log" 2>/dev/null'
g -C "$WORK/b-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "base_mismatch: b recording feature/x too passes the gate" '! printf "%s" "$OUT" | grep -q "STOP: base-mismatch" && grep -q flow.preflight "$WORK/stages.log"'
fixture
g -C "$WORK/a" push -q origin main:refs/heads/feature/x
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "base_mismatch: feature/x on a's origin only passes the gate" '! printf "%s" "$OUT" | grep -q "STOP: base-mismatch" && grep -q flow.preflight "$WORK/stages.log"'
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
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" --plan-message "Task 3 is unticked"
check "outstanding: --plan-message is refused" '[ "$RC" -ne 0 ] && grep -q "unknown argument .--plan-message." "$WORK/stderr"'
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" --outstanding "Task 3 is unticked"
check "outstanding: commit exits 0" '[ "$RC" -eq 0 ]'
check "outstanding: the implementation commit keeps its own subject and carries the text in its body" \
  '[ "$(g -C "$WORK/a-worktrees/demo" log --format=%s --grep="^Outstanding: " -1)" = "feat(a): do the work" ] && g -C "$WORK/a-worktrees/demo" log --format=%b --grep="^Outstanding: " -1 | grep -qx "Outstanding: Task 3 is unticked"'`},
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
check "drift_switch: no foreign-staged stop and the preflight ran" '! printf "%s" "$OUT" | grep -q "STOP: foreign-staged" && grep -q flow.preflight "$WORK/stages.log"'`},
		{"drift_switch_stops", `# --- drift_switch_stops: each repository is judged on its own; a dirty or
# detached checkout is never switched ---
fixture
g -C "$WORK/a" checkout -q -b side
g -C "$WORK/b" checkout -q -b other
printf 'dirty\n' >"$WORK/b/base.txt"
sut prepare "$WORK/a" demo
check "drift_switch_stops: a dirty b stops prepare at STOP: foreign-staged" \
  '[ "$RC" -eq 1 ] && printf "%s\n" "$OUT" | tail -1 | grep -q "^STOP: foreign-staged"'
check "drift_switch_stops: a is switched though b stops the run" \
  '[ "$(git -C "$WORK/a" branch --show-current)" = main ] && printf "%s\n" "$OUT" | grep -qF "DRIFT-SWITCHED: $WORK/a "'
check "drift_switch_stops: dirty b is not switched" \
  '[ "$(git -C "$WORK/b" branch --show-current)" = other ] && [ "$(cat "$WORK/b/base.txt")" = dirty ] && ! printf "%s\n" "$OUT" | grep -qF "DRIFT-SWITCHED: $WORK/b "'
fixture
g -C "$WORK/a" checkout -q --detach
sut prepare "$WORK/a" demo
check "drift_switch_stops: a detached HEAD stops and stays detached" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | grep -q "STOP: foreign-staged" && [ -z "$(git -C "$WORK/a" branch --show-current)" ] && ! printf "%s" "$OUT" | grep -q "DRIFT-SWITCHED"'`},
		{"drift_switch_refused", `# --- drift_switch_refused: a switch git refuses changes nothing and stops ---
fixture
g -C "$WORK/a" checkout -q -b side
g -C "$WORK/a" worktree add -q "$WORK/a-main" main
sut prepare "$WORK/a" demo
check "drift_switch_refused: a switch git refuses stops and leaves a on side" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | grep -q "STOP: foreign-staged" && printf "%s\n" "$OUT" | grep -qxF "DRIFT-SWITCH-REFUSED: $WORK/a — git switch main failed; left on side" && [ "$(git -C "$WORK/a" branch --show-current)" = side ]'
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
		{"live_check", `# --- live_check ---
fixture
d="$WORK/a-worktrees/demo/spectre/changes/demo"
printf '# demo\n\n## Live check\n\nRun prepare twice and compare the two runs.\n\n## Decisions\n' >"$d/design.md"
sut prepare "$WORK/a" demo
check "live_check: a named live check with no record stops prepare at STOP: live-check" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: live-check — "'
printf 'Before figure: not measured\n' >"$d/live-verification.md"
sut prepare "$WORK/a" demo --accept-outstanding
check "live_check: a refused record stops prepare, --accept-outstanding or not" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: live-check — "'
chmod 000 "$d/live-verification.md"
sut prepare "$WORK/a" demo
chmod 644 "$d/live-verification.md"
check "live_check: an unreadable record stops prepare with the guard's could-not-answer text" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: live-check — check-live-verification could not answer "'
printf 'Before: two lines. After: two lines.\n' >"$d/live-verification.md"
sut prepare "$WORK/a" demo
check "live_check: a passing record lets prepare through" '[ "$RC" -eq 0 ]'`},
		{"live_check_none", `# --- live_check_none ---
fixture
printf '# demo\n\n## Live check\n\nNone — the change runs per invocation;\nno service holds its state.\n\n## Decisions\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/design.md"
sut prepare "$WORK/a" demo
check "live_check_none: a wrapped no-runtime section needs no record" '[ "$RC" -eq 0 ]'`},
	})
}

// TestIntegrateChangeCommitRefusesAnOpenTestFailure: a narrative line that
// leaves a failing or flaky test open stops commit; one that says it was fixed,
// or records an operator's decision, or names no test, passes (kan-1103
// self-review, F2).
func TestIntegrateChangeCommitRefusesAnOpenTestFailure(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"narrative_open_test", `# --- narrative_open_test ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\n%s\n' '- Open, not mine: ` + "`" + `TestX` + "`" + ` failed once under load; it passes on re-run. Not fixed here.' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_open_test: a line leaving a failing test open stops commit" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*leaves a failing or flaky test open"'`},
		{"narrative_open_test_fixed", icNarrativePasses + `# --- narrative_open_test_fixed ---
narrative_passes "a failing test the run fixed" '- ` + "`" + `TestX` + "`" + ` failed once under load; fixed by the self-review (F1).'`},
		{"narrative_open_test_preexisting_fixed", icNarrativePasses + `# --- narrative_open_test_preexisting_fixed ---
narrative_passes "a pre-existing flaky test the run fixed" '- Pre-existing flaky test ` + "`" + `TestX` + "`" + ` fixed at its source in abc1234 (F1).'`},
		{"narrative_open_test_operator", icNarrativePasses + `# --- narrative_open_test_operator ---
narrative_passes "a flake the operator chose to proceed past" '- Pre-existing flake in the visual spec: the operator chose to proceed past it.'`},
		{"narrative_open_test_no_test", icNarrativePasses + `# --- narrative_open_test_no_test ---
narrative_passes "a pre-existing failure naming no test" '- admin-frontend failed to start — a pre-existing port conflict.'`},
	})
}

// icNarrativePasses is narrative_format's pass case, one call per case.
const icNarrativePasses = `# narrative_passes <label> <line>: a fresh prepare, <line> as the narrative's
# in-run fix entry, and commit must pass.
narrative_passes() {
  fixture
  sut prepare "$WORK/a" demo
  [ "$RC" -eq 0 ] || halt prepare
  printf '\n## integrate run\n\n%s\n' "$2" >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
  check "narrative_format: $1 passes commit" '[ "$RC" -eq 0 ]'
}
`

// icRerunPinCase is rerun_pin's case, one call per case.
const icRerunPinCase = `rerun_pin_case() {
  fixture
  to_land "${1}"
  [ "${2:-}" = reclone ] && git -C "$WORK/a" update-ref -d refs/flow/pre-reshape/demo
  A="$WORK/a-worktrees/demo"
  printf 'more\n' >"$WORK/a-worktrees/demo/more.txt"
  g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): more" -m "Task-Id: 2"
  sut prepare "$WORK/a" demo
  [ "$RC" -eq 0 ] || halt prepare
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): more work" "$WORK/b=feat(b): do the work"
  check "${1}the re-run commit exits 0" '[ "$RC" -eq 0 ]'
  check "${1}a's pin reaches both Task-Id commits" \
    '[ "$(git -C "$WORK/a" log --format=%s --grep="^Task-Id: [12]$" main..refs/flow/pre-reshape/demo | sort | tr "\n" ,)" = "feat(a): do the work,feat(a): more,feat(a): task 1," ]'
  check "${1}origin holds the advanced pin" \
    '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$(git -C "$WORK/a" rev-parse refs/flow/pre-reshape/demo)" ]'
}
`

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
check "rerun_on_disk: its NEXT names the live narrative" 'printf "%s" "$OUT" | tail -1 | grep -qF "append this run'"'"'s narrative to $A/spectre/changes/demo/narrative.md"'
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
check "rerun_reshape_base: with no fixed row the reshape base is run 1's implementation commit" 'printf "%s" "$OUT" | grep -qF "WORKTREE: $A — base main, reshape base $split"'
printf 'fix\n' >"$A/sr-fix.txt"
g -C "$A" add sr-fix.txt && g -C "$A" commit -qm "fix(a): a self-review fix" -m "Flow-Key: self-review-fix-1"
fix="$(git -C "$A" rev-parse HEAD)"
printf '[{"angle":"code","disposition":"fixed","ref":"%s"}]' "$(git -C "$A" rev-parse --short=12 HEAD)" >"$WORK/sr-findings.json"
sut prepare "$WORK/a" demo
check "rerun_reshape_base: a fixed row's commit is the reshape base" 'printf "%s" "$OUT" | grep -qF "WORKTREE: $A — base main, reshape base $fix"'`},
	})
}

// TestIntegrateChangeRequiresInRunFixLine pins kan-1032's self-review: a
// fixed flow-fix row the store lists needs an in-run fix line naming its sha.
func TestIntegrateChangeDerivesInRunFixLine(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"fix_line_derived", `# --- fix_line_derived ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
echo '[{"angle":"flow-fix","disposition":"fixed","ref":"1a2b3c4d5e6f","note":"the renderer could not render an exp- slot","blastRadius":2}]' >"$WORK/sr-findings.json"
printf '\n## integrate run\n\nnarrative' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "fix_line_derived: a row with its note and blast radius passes commit" '[ "$RC" -eq 0 ]'
check "fix_line_derived: the line is appended to the narrative" \
  'grep -qxF -- "- In-run pipeline fix: 1a2b3c4 — the renderer could not render an exp- slot (blast radius 2 files)" "$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"'
check "fix_line_derived: the append is reported" \
  'printf "%s" "$OUT" | grep -q "^narrative: appended "'
rm -f "$WORK/sr-findings.json"`},
		{"fix_line_multiline_note", `# --- fix_line_multiline_note ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '%s\n' '[{"angle":"flow-fix","disposition":"fixed","ref":"1a2b3c4d5e6f","note":"two\nlines","blastRadius":2}]' >"$WORK/sr-findings.json"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "fix_line_multiline_note: a multi-line note stops commit with exit 1" '[ "$RC" -eq 1 ]'
check "fix_line_multiline_note: the stop names the sha" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*1a2b3c4"'
rm -f "$WORK/sr-findings.json"`},
	})
}

func TestIntegrateChangeRequiresInRunFixLine(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"fix_line_missing", `# --- fix_line_missing ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
echo '[{"angle":"flow-fix","disposition":"fixed","ref":"1a2b3c4d5e6f"}]' >"$WORK/sr-findings.json"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "fix_line_missing: a fixed flow-fix row with no line stops commit with exit 1" '[ "$RC" -eq 1 ]'
check "fix_line_missing: the stop names the sha" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*1a2b3c4"'
rm -f "$WORK/sr-findings.json"`},
		{"fix_line_present", `# --- fix_line_present ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
echo '[{"angle":"flow-fix","disposition":"fixed","ref":"1a2b3c4d5e6f"}]' >"$WORK/sr-findings.json"
printf '\n## integrate run\n\n%s\n' '- In-run pipeline fix: 1a2b3c4 — the renderer could not render an exp- slot (blast radius 2 files)' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "fix_line_present: a fixed row with its line passes commit" '[ "$RC" -eq 0 ]'
rm -f "$WORK/sr-findings.json"`},
		{"fix_line_other_rows", `# --- fix_line_other_rows ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
echo '[{"angle":"flow-cost","disposition":"fixed","ref":"1a2b3c4d5e6f"},{"angle":"flow-fix","disposition":"noted","ref":"2b3c4d5e6f7a"}]' >"$WORK/sr-findings.json"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "fix_line_other_rows: a flow-cost fixed row and a flow-fix noted row pass commit" '[ "$RC" -eq 0 ]'
rm -f "$WORK/sr-findings.json"`},
	})
}

// icDesignRecord writes a design.md into the canonical change directory and
// sets rec to the record's path; design_files <worktree> prints the paths
// the worktree's branch commits since main list.
const icDesignRecord = `design() {
  rec="$WORK/a-worktrees/demo/docs/designs/demo.md"
  printf '## Context\n\n- one\n\n## Decisions\n\n**Status:** active\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/design.md"
}
design_files() {
  git -C "$1" log --format= --name-only main..HEAD
}
`

// TestIntegrateChangeDesignRecord pins run 1's design record: commit writes
// the canonical change directory's design.md, under its two-line header, into
// the canonical worktree's implementation commit, and prepare asks for that
// worktree's subject whenever a design.md exists.
func TestIntegrateChangeDesignRecord(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"record_in_impl_commit", icDesignRecord + `# --- record_in_impl_commit ---
fixture
design
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "record_in_impl_commit: commit exits 0" '[ "$RC" -eq 0 ]'
check "record_in_impl_commit: a's implementation commit lists the record" \
  'git -C "$WORK/a-worktrees/demo" log --format= --name-only --grep="^feat(a): do the work$" main..HEAD | grep -qx docs/designs/demo.md'
check "record_in_impl_commit: line 1 is the title" '[ "$(sed -n 1p "$rec")" = "# demo" ]'
check "record_in_impl_commit: line 2 is the header with the Jira key" \
  '[ "$(sed -n 2p "$rec")" = "> Design record of change demo (KAN-1), written at landing. Each decision'"'"'s own **Status:** line is the status." ]'
check "record_in_impl_commit: line 3 is empty" '[ -z "$(sed -n 3p "$rec")" ]'
check "record_in_impl_commit: the rest is design.md byte for byte" \
  'tail -n +4 "$rec" | cmp -s - "$WORK/a-worktrees/demo/spectre/changes/demo/design.md"'
check "record_in_impl_commit: b's commit lists no record" '! design_files "$WORK/b-worktrees/demo" | grep -q "^docs/designs/"'`},
		{"no_design_no_record", icDesignRecord + `# --- no_design_no_record ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "no_design_no_record: commit exits 0" '[ "$RC" -eq 0 ]'
check "no_design_no_record: no commit lists a record" \
  '! design_files "$WORK/a-worktrees/demo" | grep -q "^docs/designs/" && ! design_files "$WORK/b-worktrees/demo" | grep -q "^docs/designs/"'`},
		{"record_only_canon", icDesignRecord + `# --- record_only_canon ---
fixture
g -C "$WORK/a-worktrees/demo" reset -q --hard main
c="$WORK/a-worktrees/demo/spectre/changes/demo"
mkdir -p "$c"
printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
printf '# demo\n\n## Merge order\n\n1. ` + "`" + `b` + "`" + `\n2. ` + "`" + `.` + "`" + `\n' >"$c/link.md"
design
sut prepare "$WORK/a" demo
check "record_only_canon: prepare exits 0" '[ "$RC" -eq 0 ]'
check "record_only_canon: prepare's NEXT asks a subject for a" \
  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo=<impl subject>"'
printf '\n## integrate run\n\nnarrative\n' >>"$c/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): record the design" "$WORK/b-worktrees/demo=feat(b): do the work"
check "record_only_canon: commit exits 0" '[ "$RC" -eq 0 ]'
check "record_only_canon: a's implementation commit lists the record" \
  'git -C "$WORK/a-worktrees/demo" log --format= --name-only --grep="^feat(a): record the design$" main..HEAD | grep -qx docs/designs/demo.md'`},
		{"record_write_fails", icDesignRecord + `# --- record_write_fails ---
fixture
design
mkdir -p "$WORK/a-worktrees/demo/docs" && printf 'x\n' >"$WORK/a-worktrees/demo/docs/designs"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "record_write_fails: an unwritable record stops commit naming its path" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -qF "STOP: design-record — cannot create the directory of $WORK/a-worktrees/demo/docs/designs/demo.md"'`},
		{"no_jira_key", icDesignRecord + `# --- no_jira_key ---
fixture
sed -i.bak 's/"jiraIssue":"KAN-1",//' "$WORK/state.json"
design
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "no_jira_key: commit exits 0" '[ "$RC" -eq 0 ]'
check "no_jira_key: line 2 drops the parenthesis" \
  '[ "$(sed -n 2p "$rec")" = "> Design record of change demo, written at landing. Each decision'"'"'s own **Status:** line is the status." ]'`},
	})
}
