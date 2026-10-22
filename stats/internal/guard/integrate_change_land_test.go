package guard

import "testing"

// The land, cleanup and visual-gate cases of integrate-change's former bash
// harness, run as integrate_change_test.go runs the happy path; peer_fix_case is the
// helper both peer-fix cases call.
func TestIntegrateChangeLandCleanup(t *testing.T) {
	icRunCases(t, []icCaseDef{
		{"no_report", `# --- no_report ---
fixture
to_land "no_report: " no-report
pushed="$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)"
printf 'later\n' >"$WORK/a-worktrees/demo/later.txt"
g -C "$WORK/a-worktrees/demo" add later.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): later" -m "Task-Id: 2"
sut land "$WORK/a" demo
check "no_report: land stops with exit 1" '[ "$RC" -eq 1 ]'
check "no_report: the stop is STOP: self-review-report" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: self-review-report"'
check "no_report: nothing was landed" \
  '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ]'
check "no_report: a's remote branch still sits at commit's HEAD" \
  '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$pushed" ]'`},
		{"loop_landing", `# --- loop_landing ---
fixture
to_land "loop_landing: "
printf '[{"seq":1,"key":"self-review-demo-plan","outcome":"","sessionToken":"self-review-demo"}]\n' >"$WORK/loop-rows"
sut land "$WORK/a" demo
check "loop_landing: an open own-branch self-review loop stops land at STOP: loop-landing" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: loop-landing"'
check "loop_landing: nothing landed on either main" \
  '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ]'`},
		{"rebase_lint_land", `# --- rebase_lint_land: land's re-sync after a rejected push runs the list ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
[ "$RC" -eq 0 ] || halt commit
mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
printf '\n## lint\n\n` + "`" + `` + "`" + `` + "`" + `bash\necho land-lint-output; exit 1\n` + "`" + `` + "`" + `` + "`" + `\n' >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" add .flow/project.md && g -C "$WORK/a" commit -qm "base: lint list" && g -C "$WORK/a" push -q origin main
sut land "$WORK/a" demo
check "rebase_lint_land: land stops at rebase-lint naming what was already pushed" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — already pushed: $WORK/b-worktrees/demo; the ## lint list failed on $WORK/a-worktrees/demo"'
check "rebase_lint_land: the list's own output is relayed" 'printf "%s" "$OUT" | grep -q "land-lint-output"'
check "rebase_lint_land: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: lint list" ]'`},
		{"gh_open_merged", `# --- gh_open_merged ---
fixture
printf '#!/usr/bin/env bash\n[ "$1 $2" = "pr view" ] && { echo OPEN; exit 0; }\nexit 1\n' >"$WORK/bin/gh"
chmod +x "$WORK/bin/gh"
jq -c '.prUrl = "https://forge.invalid/pr/1"' "$WORK/state.json" >"$WORK/state.tmp" && mv "$WORK/state.tmp" "$WORK/state.json"
to_land "gh_open_merged: "
sut land "$WORK/a" demo
check "gh_open_merged: land runs to the Finished handoff" '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
check "gh_open_merged: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"disclose", `# --- disclose ---
fixture
for r in a b; do
  printf 'secret.txt\n' >>"$WORK/$r/.git/info/exclude"
  printf 'irreplaceable\n' >"$WORK/$r-worktrees/demo/secret.txt"
done
to_land "disclose: "
sut land "$WORK/a" demo
check "disclose: land's run 2 stops at STOP: disclose for a" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: disclose — relay $WORK/a.s .*--proceed $WORK/a "'
sut cleanup "$WORK/a" demo --proceed "$WORK/a"
check "disclose: --proceed a removes a" '[ ! -d "$WORK/a-worktrees/demo" ]'
check "disclose: b is not removed on a's answer" '[ -f "$WORK/b-worktrees/demo/secret.txt" ]'
check "disclose: the re-run stops at STOP: disclose for b" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: disclose — relay $WORK/b.s "'
sut cleanup "$WORK/a" demo --proceed "$WORK/a" --proceed "$WORK/b"
check "disclose: --proceed for both finishes" '[ "$RC" -eq 0 ] && [ ! -d "$WORK/b-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"rerun_leftover", `# --- rerun_leftover ---
fixture
mkdir -p "$WORK/a/spectre/changes/demo"
to_land "rerun_leftover: "
sut land "$WORK/a" demo
check "rerun_leftover: run 2 stops at STOP: leftover" '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: leftover"'
check "rerun_leftover: both worktrees were removed first" '[ ! -d "$WORK/a-worktrees/demo" ] && [ ! -d "$WORK/b-worktrees/demo" ]'
rmdir "$WORK/a/spectre/changes/demo"
g -C "$WORK/b" push -q origin main:refs/heads/spectre/demo
sut cleanup "$WORK/a" demo
check "rerun_leftover: the re-run reaches b and deletes its remote branch" \
  '! git --git-dir="$WORK/b-origin.git" rev-parse -q --verify refs/heads/spectre/demo >/dev/null'
check "rerun_leftover: b's cleanup check ran" 'printf "%s" "$OUT" | grep -q "^COMPLETE: .*$WORK/b"'
check "rerun_leftover: the re-run writes FINISHED" '[ "$RC" -eq 0 ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"no_state_dir", `# --- no_state_dir ---
fixture
to_land "no_state_dir: "
rm -rf "$WORK/statedir"
sut land "$WORK/a" demo
check "no_state_dir: run 2 reaches COMPLETE, not STOP: leftover" \
  '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | grep -q "^COMPLETE: " && ! printf "%s" "$OUT" | grep -q "^STOP: leftover"'
check "no_state_dir: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"pr_url", `# --- pr_url ---
fixture
printf '#!/usr/bin/env bash\necho "gh: no forge" >&2\nexit 1\n' >"$WORK/bin/gh"
chmod +x "$WORK/bin/gh"
to_land "pr_url: "
sut land "$WORK/a" demo --route "pull request"
check "pr_url: no usable gh stops at STOP: pr for b first" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: pr — no usable gh for $WORK/b-worktrees/demo "'
sut land "$WORK/a" demo --route "pull request" --pr-url "$WORK/b-worktrees/demo=https://forge.invalid/b/1"
check "pr_url: b's answer does not answer for a" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: pr — no usable gh for $WORK/a-worktrees/demo "'
sut land "$WORK/a" demo --route "pull request" --pr-url "$WORK/b-worktrees/demo=https://forge.invalid/b/1" --pr-url "$WORK/a-worktrees/demo=https://forge.invalid/a/1"
check "pr_url: with both answered, IN_PROGRESS records the canonical PR" \
  '[ "$RC" -eq 0 ] && [ "$(jq -r "[.state, .prUrl] | join(\" \")" "$WORK/state.json")" = "IN_PROGRESS https://forge.invalid/a/1" ]'`},
		{"peer_fix", `# --- peer_fix ---
fixture
peer_fix_case "peer_fix: "`},
		{"peer_fix_canon_first", `# --- peer_fix_canon_first ---
fixture
printf '# demo\n\n## Parts\n\n` + "`" + `b:demo` + "`" + `\n\n## Branch\n\nmain\n\n## Merge order\n\n1. ` + "`" + `.` + "`" + `\n2. ` + "`" + `b` + "`" + `\n' \
  >"$WORK/a-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
peer_fix_case "peer_fix_canon_first: "`},
		{"visual_stop", `# --- visual_stop ---
fixture
ui_a
A="$WORK/a-worktrees/demo"
sut prepare "$WORK/a" demo
check "visual_stop: an unverified UI HEAD stops prepare with exit 1" '[ "$RC" -eq 1 ]'
check "visual_stop: the stop is STOP: visual-verify" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
sut prepare "$WORK/a" demo --accept-outstanding
check "visual_stop: --accept-outstanding does not pass it" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
check "visual_stop: no worktree is judged after the MISSING one, whose row stays the newest to flag" \
  '! printf "%s\n" "$OUT" | sed -n "/^VISUAL-VERIFY-MISSING:/,\$p" | sed 1d | grep -q "^VISUAL-VERIFY-"'
check "visual_stop: the gate mark closes stopped" 'tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.unfinished-work-gate -outcome stopped"'
flag "$A" "$(git -C "$A" rev-parse HEAD~1)"
sut prepare "$WORK/a" demo --accept-outstanding
check "visual_stop: a false-positive flag on an earlier HEAD does not pass it" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
flag "$WORK/b-worktrees/demo" "$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
check "visual_stop: a flag on another worktree's row does not pass it" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
flag "$A" "$(git -C "$A" rev-parse HEAD)" other
sut prepare "$WORK/a" demo
check "visual_stop: another change's flag does not pass it" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
check "visual_stop: a flag on this worktree's row for this HEAD passes the gate" '[ "$RC" -eq 0 ]'
rm "$WORK/verdicts.json"
printf '[{"key":"task-1-implementer","outcome":"completed"},{"key":"verify","role":"verifier","outcome":"completed","startedAt":"2026-01-02T00:00:00Z"},{"key":"visual-verify","role":"verifier","outcome":"completed","diffBase":"%s"}]' \
  "$(git -C "$A" rev-parse HEAD)" >"$WORK/dispatches.json"
sut prepare "$WORK/a" demo
check "visual_stop: a verifier of the HEAD passes the gate" '[ "$RC" -eq 0 ]'`},
		{"visual_archived", `# --- visual_archived: an archived re-run with no new work runs the gate ---
fixture
ui_a
A="$WORK/a-worktrees/demo"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$A/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "visual_archived: commit archives, then stops before a's unverified branch push" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
sut prepare "$WORK/a" demo
check "visual_archived: the re-run is archived without new work" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — without new work"'
check "visual_archived: a flag on the pre-archive HEAD does not pass the gate" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'`},
		{"visual_rebased", `# --- visual_rebased: a base move in prepare re-runs the gate ---
fixture
ui_a
A="$WORK/a-worktrees/demo"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
printf 'moved\n' >"$WORK/a/moved.txt"
g -C "$WORK/a" add moved.txt && g -C "$WORK/a" commit -qm "base: moved" && g -C "$WORK/a" push -q origin main
sut prepare "$WORK/a" demo
check "visual_rebased: a was rebased" 'printf "%s\n" "$OUT" | grep -q "^REBASED: "'
check "visual_rebased: the rebased HEAD no verifier saw stops prepare" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — rebased onto origin/main after the gate: .*a-worktrees/demo: "'`},
		{"visual_resync", `# --- visual_resync: land's re-sync re-runs the gate before that push ---
fixture
ui_a
A="$WORK/a-worktrees/demo"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$A/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "visual_resync: commit's reshaped HEAD stops before a's branch push" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "visual_resync: commit re-run with that HEAD flagged exits 0" '[ "$RC" -eq 0 ]'
mkdir -p "$A/docs/self-review"
printf '# demo self-review\n' >"$A/docs/self-review/demo-self-review.md"
g -C "$A" add docs/self-review/demo-self-review.md
g -C "$A" commit -qm "docs(self-review): demo self-review report"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
printf 'moved\n' >"$WORK/a/moved.txt"
g -C "$WORK/a" add moved.txt && g -C "$WORK/a" commit -qm "base: moved" && g -C "$WORK/a" push -q origin main
sut land "$WORK/a" demo
check "visual_resync: a's re-synced HEAD no verifier saw stops land" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — .*already pushed: $WORK/b-worktrees/demo"'
check "visual_resync: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'
sut land "$WORK/a" demo
check "visual_resync: re-running land after the stop still stops at visual-verify" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
check "visual_resync: the re-run did not push a's unverified HEAD" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'`},
		{"visual_resync_reread", `# --- visual_resync_reread: land's re-sync rewrites the canonical report first ---
# a is canonical and its own fix sha is rebased by the re-sync, so
# reread_fix_shas commits the rewrite into a; a flag on the pre-rewrite
# HEAD must not pass the gate for the rewrite commit pushed after it. A
# fixed committer date makes the re-sync's HEAD predictable to flag.
fixture
ui_a
A="$WORK/a-worktrees/demo"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$A/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
[ "$RC" -eq 0 ] || halt commit
old="$(git -C "$A" log --format=%h --grep="^feat(a): do the work$" -1 HEAD)"
mkdir -p "$A/docs/self-review"
printf '# demo self-review\n\n- **[flow-fix]** the fix — fixed: %s\n' "$old" >"$A/docs/self-review/demo-self-review.md"
g -C "$A" add docs/self-review/demo-self-review.md
g -C "$A" commit -qm "docs(self-review): demo self-review report"
printf 'moved\n' >"$WORK/a/moved.txt"
g -C "$WORK/a" add moved.txt && g -C "$WORK/a" commit -qm "base: moved" && g -C "$WORK/a" push -q origin main
export GIT_COMMITTER_DATE="2026-01-01T00:00:00Z"
g clone -q "$A" "$WORK/pred" && g -C "$WORK/pred" fetch -q "$WORK/a-origin.git" main &&
  g -C "$WORK/pred" rebase -q --onto FETCH_HEAD "$(git -C "$A" merge-base HEAD "$(git --git-dir="$WORK/a-origin.git" rev-parse main~1)")" >/dev/null 2>&1
flag "$A" "$(git -C "$A" rev-parse HEAD)"
cur="$(sed 's/^\[//; s/\]$//' "$WORK/verdicts.json")"
flag "$A" "$(git -C "$WORK/pred" rev-parse HEAD)"
printf '[%s,%s]' "$cur" "$(sed 's/^\[//; s/\]$//' "$WORK/verdicts.json")" >"$WORK/verdicts.json"
sut land "$WORK/a" demo
unset GIT_COMMITTER_DATE
check "visual_resync_reread: the re-sync rewrote a's own fix sha" 'printf "%s\n" "$OUT" | grep -q "^FIXED-SHA: $old -> "'
check "visual_resync_reread: a flag on the pre-rewrite HEAD does not pass the rewrite commit" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
check "visual_resync_reread: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'`},
		{"visual_reread", `# --- visual_reread: reread_fix_shas's canonical re-push runs the gate ---
# The report is a's declared UI path, so the rewrite commit alone — all a's
# HEAD adds past its landed push — is a UI change no verifier saw.
fixture
printf '# demo\n\n## Parts\n\n` + "`" + `b:demo` + "`" + `\n\n## Branch\n\nmain\n\n## Merge order\n\n1. ` + "`" + `.` + "`" + `\n2. ` + "`" + `b` + "`" + `\n' \
  >"$WORK/a-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
printf '\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| ` + "`" + `ui paths` + "`" + ` | ` + "`" + `docs/self-review/demo-self-review.md` + "`" + ` |\n' >>"$WORK/a-worktrees/demo/.flow/project.md"
g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
A="$WORK/a-worktrees/demo"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$A/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
[ "$RC" -eq 0 ] || halt commit
printf 'fix\n' >"$WORK/b-worktrees/demo/fix-b.txt"
g -C "$WORK/b-worktrees/demo" add fix-b.txt && g -C "$WORK/b-worktrees/demo" commit -qm "fix(b): the peer fix"
old="$(git -C "$WORK/b-worktrees/demo" rev-parse --short HEAD)"
mkdir -p "$A/docs/self-review"
printf '# demo self-review\n\n- **[flow-fix]** the peer fix — fixed: %s\n' "$old" >"$A/docs/self-review/demo-self-review.md"
g -C "$A" add docs/self-review/demo-self-review.md
g -C "$A" commit -qm "docs(self-review): demo self-review report"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
printf 'moved\n' >"$WORK/b/moved.txt"
g -C "$WORK/b" add moved.txt && g -C "$WORK/b" commit -qm "base: moved" && g -C "$WORK/b" push -q origin main
sut land "$WORK/a" demo
check "visual_reread: the canonical push landed first" 'git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $old$"'
check "visual_reread: the rewrite of b's fix sha stops the canonical re-push" \
  'printf "%s\n" "$OUT" | grep -q "^FIXED-SHA: $old -> " && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — the rewritten report was not pushed"'
check "visual_reread: neither a's branch nor its base carries the rewritten report" \
  'git --git-dir="$WORK/a-origin.git" show spectre/demo:docs/self-review/demo-self-review.md | grep -q "fixed: $old$" && git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $old$"'`},
		{"visual_uncommitted", `# --- visual_uncommitted: an uncommitted UI edit folded in by commit stops every push ---
for route in "merge and push" "pull request" manual; do
  fixture
  ui_a
  A="$WORK/a-worktrees/demo"
  printf '#!/bin/sh\necho https://forge.invalid/pr/1\n' >"$WORK/bin/gh"; chmod +x "$WORK/bin/gh"
  printf '[{"key":"task-1-implementer","outcome":"completed"},{"key":"verify","role":"verifier","outcome":"completed","startedAt":"2026-01-02T00:00:00Z"},{"key":"visual-verify","role":"verifier","outcome":"completed","diffBase":"%s"}]' \
    "$(git -C "$A" rev-parse HEAD)" >"$WORK/dispatches.json"
  printf 'edited\n' >>"$A/work-a.txt"
  sut prepare "$WORK/a" demo
  [ "$RC" -eq 0 ] || halt prepare
  printf '\n## integrate run\n\nnarrative\n' >>"$A/spectre/changes/demo/narrative.md"
  sut commit "$WORK/a" demo "$A=feat(a): do the work" "$WORK/b=feat(b): do the work"
  check "visual_uncommitted ($route): commit stops at visual-verify" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
  check "visual_uncommitted ($route): commit did not push a's branch" '! git --git-dir="$WORK/a-origin.git" rev-parse -q --verify spectre/demo >/dev/null'
  mkdir -p "$A/docs/self-review"
  printf '# demo self-review\n' >"$A/docs/self-review/demo-self-review.md"
  g -C "$A" add docs/self-review/demo-self-review.md
  g -C "$A" commit -qm "docs(self-review): demo self-review report"
  sut land "$WORK/a" demo --route "$route"
  check "visual_uncommitted ($route): land stops at visual-verify" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
  check "visual_uncommitted ($route): land pushed neither a's branch nor its base" '! git --git-dir="$WORK/a-origin.git" rev-parse -q --verify spectre/demo >/dev/null && [ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ]'
done`},
		{"refresh_on_recorded_base", `# --- refresh_on_recorded_base ---
fixture
g -C "$WORK/a" push -q origin main:refs/heads/feature/x
g -C "$WORK/a" checkout -q -b feature/x --track origin/feature/x
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "refresh_on_recorded_base: prepare never switches a checkout on its recorded base" \
  '[ "$RC" -eq 1 ] && [ "$(git -C "$WORK/a" branch --show-current)" = feature/x ] && ! printf "%s" "$OUT" | grep -q "^DRIFT-SWITCHED:"'
# a's main checkout off main is foreign-staged drift; the operator continues.
sut prepare "$WORK/a" demo --accept-foreign
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
[ "$RC" -eq 0 ] || halt commit
mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
sut land "$WORK/a" demo
check "refresh_on_recorded_base: a's main checkout on its recorded base is refreshed, not refused" \
  'printf "%s\n" "$OUT" | grep -qE "^REFRESH-(DONE|CURRENT): $WORK/a( |\$)" && ! printf "%s\n" "$OUT" | grep -q "^REFRESH-REFUSED: $WORK/a "'
# The re-run a leftover stop asks for comes after the worktrees are gone.
sut cleanup "$WORK/a" demo
check "refresh_on_recorded_base: a cleanup re-run after worktree removal still refreshes onto feature/x" \
  'printf "%s\n" "$OUT" | grep -qE "^REFRESH-(DONE|CURRENT): $WORK/a( |\$)" && ! printf "%s\n" "$OUT" | grep -q "^REFRESH-REFUSED: $WORK/a "'`},
		{"orphan_sweep", `# --- orphan_sweep ---
fixture
printf '\n## workspace isolation\n\n| Command | Runs |\n|---------|------|\n| ` + "`" + `remove` + "`" + ` | ` + "`" + `echo <id> >>%s/removed.log; [ <id> != orph-bad ]` + "`" + ` |\n| ` + "`" + `orphans` + "`" + ` | ` + "`" + `printf "orph-1\\norph-bad\\n"` + "`" + ` |\n' "$WORK" >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" commit -qam "chore: workspace commands" && g -C "$WORK/a" push -q origin main
to_land "orphan_sweep: "
sut land "$WORK/a" demo
check "orphan_sweep: the change's own workspace is removed" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-REMOVED: $WORK/a"'
check "orphan_sweep: each orphan id goes through the remove command" '[ "$(cat "$WORK/removed.log")" = "$(printf "demo-id\norph-1\norph-bad")" ]'
check "orphan_sweep: a removed orphan is reported" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVED: orph-1"'
check "orphan_sweep: a failed orphan removal is reported and cleanup still finishes" \
  'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVE-FAILED: orph-bad — exit 1" && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"indented_fixed_row", `# --- indented_fixed_row ---
fixture
to_land "indented_fixed_row: " no-report
printf 'fix\n' >"$WORK/b-worktrees/demo/fix-b.txt"
g -C "$WORK/b-worktrees/demo" add fix-b.txt && g -C "$WORK/b-worktrees/demo" commit -qm "fix(b): the peer fix"
old="$(git -C "$WORK/b-worktrees/demo" rev-parse --short HEAD)"
mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
printf '# demo self-review\n\n  - **[flow-fix]** the peer fix — fixed: %s\n' "$old" >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
sut land "$WORK/a" demo
check "indented_fixed_row: land exits 0" '[ "$RC" -eq 0 ]'
check "indented_fixed_row: an indented report row is not recorded as fixed" '! grep -q -- "-disposition fixed" "$WORK/fixed.log" 2>/dev/null'`},
		{"padded_orphan", `# --- padded_orphan ---
fixture
printf '\n## workspace isolation\n\n| Command | Runs |\n|---------|------|\n| ` + "`" + `remove` + "`" + ` | ` + "`" + `echo "[<id>]" >>%s/removed.log` + "`" + ` |\n| ` + "`" + `orphans` + "`" + ` | ` + "`" + `printf " orph-1\\n"` + "`" + ` |\n' "$WORK" >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" commit -qam "chore: workspace commands" && g -C "$WORK/a" push -q origin main
to_land "padded_orphan: "
sut land "$WORK/a" demo
check "padded_orphan: land exits 0" '[ "$RC" -eq 0 ]'
check "padded_orphan: a padded orphan id is refused and never removed" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVE-FAILED:  orph-1 — refused, not an id" && ! grep -q "orph-1" "$WORK/removed.log"'`},
	})
}
