package guard

import "testing"

// The prepare and commit cases of integrate-change's former bash harness,
// run as integrate_change_test.go runs the happy path.
func TestIntegrateChangePrepareCommit(t *testing.T) {
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
check "untasked: the stray commit is listed" 'printf "%s" "$OUT" | grep -q "^  [0-9a-f]\{12\} test(b): a stray edit$"'
sut prepare "$WORK/a" demo --accept-outstanding
check "untasked: --accept-outstanding passes the gate" '[ "$RC" -eq 0 ]'`},
		{"archived_rerun", `# --- archived_rerun ---
fixture
to_land "archived_rerun: "
printf 'more\n' >"$WORK/a-worktrees/demo/more.txt"
g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): more" -m "Task-Id: 2"
sut prepare "$WORK/a" demo
check "archived_rerun: prepare passes the gate" '[ "$RC" -eq 0 ]'
check "archived_rerun: b's reshaped commit is not listed" '! printf "%s" "$OUT" | grep -q "feat(b): do the work"'`},
		{"archived_rerun_self_review", `# --- archived_rerun_self_review: the self-review loop's commits sit below the base ---
fixture
to_land "archived_rerun_self_review: " no-report
A="$WORK/a-worktrees/demo"
mkdir -p "$A/spectre/changes/self-review-demo"
printf '# tasks\n' >"$A/spectre/changes/self-review-demo/tasks.md"
g -C "$A" add spectre/changes/self-review-demo && g -C "$A" commit -qm "chore(spectre): plan self-review-demo"
printf 'fix\n' >"$A/sr-fix.txt"
g -C "$A" add sr-fix.txt && g -C "$A" commit -qm "fix(a): a self-review fix"
sut prepare "$WORK/a" demo
check "archived_rerun_self_review: a resumed run 1 holds no new work" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — without new work"'
g -C "$A" commit -q --allow-empty -m "chore(spectre): plan demo — fix-run tasks"
sut prepare "$WORK/a" demo
check "archived_rerun_self_review: a fix run's plan commit above the loop is new work" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — with new work"'
g -C "$A" commit -q --allow-empty -m "chore(spectre): archive self-review-demo"
sut prepare "$WORK/a" demo
check "archived_rerun_self_review: the loop's archive commit re-covers the plan above it" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — without new work"'
g -C "$A" commit -q --allow-empty -m "chore(spectre): plan self-review-demo"
g -C "$A" commit -q --allow-empty -m "chore(spectre): plan demo"
g -C "$A" commit -q --allow-empty -m "fix(a): fix-run work"
sut prepare "$WORK/a" demo
check "archived_rerun_self_review: a bare fix-run plan commit closes the loop" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — with new work"'`},
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
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: sync-conflict — .*sync-onto-base.sh. --resume, the whole ## lint and ## test lists, then re-run prepare$"'`},
		{"guard_test", `# --- guard_test ---
# A scoped re-verification is <agents repo>/scripts/test-<stem>.sh, the repo
# derived from the guards' own path: run a copy of scripts/ whose stats/ is
# this checkout's, holding a failing test-gt.sh, so no test lands in the tree.
fixture
mkdir -p "$WORK/agents"
cp -R "$SCRIPT_DIR" "$WORK/agents/scripts"
ln -s "$(cd "$SCRIPT_DIR/../stats" && pwd -P)" "$WORK/agents/stats"
printf '#!/usr/bin/env bash\necho gt-test-own-output\nexit 1\n' >"$WORK/agents/scripts/test-gt.sh"
printf 'same\n' >"$WORK/b-worktrees/demo/gt.txt"
g -C "$WORK/b-worktrees/demo" add gt.txt && g -C "$WORK/b-worktrees/demo" commit -qm "feat(b): gt" -m "Task-Id: 2"
printf 'same\n' >"$WORK/b/gt.txt"
g -C "$WORK/b" add gt.txt && g -C "$WORK/b" commit -qm "base: gt" && g -C "$WORK/b" push -q origin main
SUT_SAVED="$SUT"; SUT="$WORK/agents/scripts/integrate-change.sh"
sut prepare "$WORK/a" demo
SUT="$SUT_SAVED"
check "guard_test: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "guard_test: the stop is STOP: guard-test" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: guard-test"'
check "guard_test: the failing test's own output is relayed" 'printf "%s" "$OUT" | grep -qx "gt-test-own-output"'`},
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
		{"rebase_lint_dirty", `# --- rebase_lint_dirty: an uncommitted plan edit the rebase set aside and
# restored is no rewrite of the list's (kan-1010's self-review) ---
fixture
printf '\nCorrection: narrative.\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/tasks.md"
printf '\n## lint\n\n` + "`" + `` + "`" + `` + "`" + `bash\ntrue\n` + "`" + `` + "`" + `` + "`" + `\n' >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" add .flow/project.md && g -C "$WORK/a" commit -qm "base: lint list" && g -C "$WORK/a" push -q origin main
sut prepare "$WORK/a" demo
check "rebase_lint_dirty: the rebase ran" 'printf "%s" "$OUT" | grep -q "^REBASED: $WORK/a-worktrees/demo"'
check "rebase_lint_dirty: the list passed" 'printf "%s" "$OUT" | grep -q "^REBASE-LINT-PASSED: $WORK/a-worktrees/demo"'
check "rebase_lint_dirty: no rebase-lint stop" '! printf "%s" "$OUT" | grep -q "^STOP: rebase-lint"'
check "rebase_lint_dirty: the edit is still uncommitted" 'grep -q "^Correction: narrative.$" "$WORK/a-worktrees/demo/spectre/changes/demo/tasks.md" && [ -n "$(g -C "$WORK/a-worktrees/demo" status --porcelain --untracked-files=no)" ]'`},
		{"nested_link", `# --- nested_link ---
# commit-split adds ` + "`" + `<spec>/changes/*/link.md` + "`" + ` only when a change directory
# holds a link.md, so b's base carries one; the branch adds a nested one.
fixture
mkdir -p "$WORK/b/spectre/changes/old"
printf 'old\n' >"$WORK/b/spectre/changes/old/link.md"
g -C "$WORK/b" add -A && g -C "$WORK/b" commit -qm "base: old link" && g -C "$WORK/b" push -q origin main
g -C "$WORK/b-worktrees/demo" reset -q --hard main
mkdir -p "$WORK/b-worktrees/demo/spectre/changes/archive/y"
printf 'y\n' >"$WORK/b-worktrees/demo/spectre/changes/archive/y/link.md"
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): nested link"
sut prepare "$WORK/a" demo
check "nested_link: prepare's NEXT asks a subject for b" \
  '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>"'
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b-worktrees/demo=feat(b): nest a link"
check "nested_link: commit exits 0" '[ "$RC" -eq 0 ]'
check "nested_link: b's implementation commit carries the nested link.md" \
  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/y/link.md)" = "feat(b): nest a link" ]'`},
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
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): notes"
sut prepare "$WORK/a" demo
check "planning_only: prepare exits 0" '[ "$RC" -eq 0 ]'
check "planning_only: prepare's NEXT asks a subject for a and none for b" \
  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo=<impl subject>" && ! printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo="'
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work"
check "planning_only: commit with no subject for b exits 0" '[ "$RC" -eq 0 ]'
check "planning_only: b's branch is the base plus its planning and archive commits" \
  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD | tr "\n" "|")" = "chore(spectre): archive demo|chore(spectre): plan demo|" ]'`},
		{"satellite_archive", `# --- satellite_archive ---
fixture
mkdir -p "$WORK/b-worktrees/demo/spectre/changes/demo"
printf '## Part of\n\n` + "`" + `a:demo` + "`" + `\n' >"$WORK/b-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): link demo"
to_land "satellite_archive: " no-report
check "satellite_archive: b's change directory is archived" \
  '[ ! -d "$WORK/b-worktrees/demo/spectre/changes/demo" ] && [ -f "$WORK/b-worktrees/demo/spectre/changes/archive/demo/link.md" ]'
check "satellite_archive: b's branch ends on its own archive commit" \
  '[ "$(git -C "$WORK/b-worktrees/demo" log -1 --format=%s)" = "chore(spectre): archive demo" ]'`},
		{"link_only_canon", `# --- link_only_canon ---
fixture
g -C "$WORK/a-worktrees/demo" reset -q --hard main
c="$WORK/a-worktrees/demo/spectre/changes/demo"
mkdir -p "$c"
printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
printf '# demo\n\n## Merge order\n\n1. ` + "`" + `b` + "`" + `\n2. ` + "`" + `.` + "`" + `\n' >"$c/link.md"
g -C "$WORK/a-worktrees/demo" add -A && g -C "$WORK/a-worktrees/demo" commit -qm "chore(spectre): merge order"
sut prepare "$WORK/a" demo
check "link_only_canon: prepare exits 0" '[ "$RC" -eq 0 ]'
check "link_only_canon: prepare's NEXT asks a subject for b and none for a" \
  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>" && ! printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo="'
printf '\n## integrate run\n\nnarrative\n' >>"$c/narrative.md"
sut commit "$WORK/a" demo "$WORK/b-worktrees/demo=feat(b): do the work"
check "link_only_canon: commit with no subject for a exits 0" '[ "$RC" -eq 0 ]'
check "link_only_canon: a's link.md lands in a commit subjected chore(spectre): merge order" \
  '[ "$(git -C "$WORK/a-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/demo/link.md spectre/changes/demo/link.md)" = "chore(spectre): merge order" ]'`},
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
check "outstanding: the planning commit keeps its own subject and carries the text in its body" \
  '[ "$(g -C "$WORK/a-worktrees/demo" log --format=%s --grep="^Outstanding: " -1)" = "chore(spectre): plan demo" ] && [ "$(g -C "$WORK/a-worktrees/demo" log --format=%b --grep="^Outstanding: " -1)" = "Outstanding: Task 3 is unticked" ]'`},
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
g -C "$WORK/a-worktrees/demo" add -A && g -C "$WORK/a-worktrees/demo" commit -q --amend -m "chore(spectre): plan-gate demo"
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
g -C "$WORK/a-worktrees/demo" add -A && g -C "$WORK/a-worktrees/demo" commit -q --amend -m "chore(spectre): plan-gate demo"
sut prepare "$WORK/a" demo
check "live_check_none: a wrapped no-runtime section needs no record" '[ "$RC" -eq 0 ]'`},
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
  printf 'more\n' >"$WORK/a-worktrees/demo/more.txt"
  g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): more" -m "Task-Id: 2"
  sut prepare "$WORK/a" demo
  [ "$RC" -eq 0 ] || halt prepare
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): more work" "$WORK/b=feat(b): do the work"
  check "${1}the re-run commit exits 0" '[ "$RC" -eq 0 ]'
  check "${1}a's pin reaches both Task-Id commits" \
    '[ "$(git -C "$WORK/a" log --format=%s --grep="^Task-Id: [12]$" main..refs/flow/pre-reshape/demo | sort | tr "\n" ,)" = "feat(a): more,feat(a): task 1," ]'
  check "${1}origin holds the advanced pin" \
    '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$(git -C "$WORK/a" rev-parse refs/flow/pre-reshape/demo)" ]'
}
`

// TestIntegrateChangeRequiresInRunFixLine pins kan-1032's self-review: a
// fixed flow-fix row the store lists needs an in-run fix line naming its sha.
func TestIntegrateChangeRequiresInRunFixLine(t *testing.T) {
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
