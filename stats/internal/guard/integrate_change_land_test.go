package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// The land and cleanup cases of integrate-change's former bash
// harness, run as integrate_change_test.go runs the happy path.
func TestIntegrateChangeLandCleanup(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"rebase_lint_land", `# --- rebase_lint_land: land's re-sync after a rejected push runs the list ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
[ "$RC" -eq 0 ] || halt commit
touch "$WORK/sr-completed"
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
sut cleanup "$WORK/a" demo
check "gh_open_merged: cleanup runs to the Finished handoff" '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
check "gh_open_merged: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"ignored_files", `# --- ignored_files ---
fixture
for r in a b; do
  printf 'secret.txt\n' >>"$WORK/$r/.git/info/exclude"
  printf 'ignored\n' >"$WORK/$r-worktrees/demo/secret.txt"
done
to_land "ignored_files: "
sut land "$WORK/a" demo
sut cleanup "$WORK/a" demo
check "ignored_files: ignored files go with the worktrees, never asked about" \
  '[ "$RC" -eq 0 ] && [ ! -d "$WORK/a-worktrees/demo" ] && [ ! -d "$WORK/b-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"cleanup_failure", `# --- cleanup_failure ---
fixture
to_land "cleanup_failure: "
printf 'dirty\n' >>"$WORK/b-worktrees/demo/base.txt"
sut land "$WORK/a" demo
sut cleanup "$WORK/a" demo
check "cleanup_failure: a refused removal stops at STOP: worktree-cleanup" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: worktree-cleanup — remove-change-worktrees exit 1 for $WORK/b "'
check "cleanup_failure: b's worktree is left alone" '[ -f "$WORK/b-worktrees/demo/base.txt" ]'
g -C "$WORK/b-worktrees/demo" checkout -q -- base.txt
sut cleanup "$WORK/a" demo
check "cleanup_failure: the re-run removes b and writes FINISHED" \
  '[ "$RC" -eq 0 ] && [ ! -d "$WORK/b-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
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
		{"refresh_on_recorded_base", `# --- refresh_on_recorded_base ---
fixture
g -C "$WORK/a" push -q origin main:refs/heads/feature/x
g -C "$WORK/a" checkout -q -b feature/x --track origin/feature/x
g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
sut prepare "$WORK/a" demo
check "refresh_on_recorded_base: a clean main checkout on its recorded base passes prepare unswitched" \
  '[ "$RC" -eq 0 ] && [ "$(git -C "$WORK/a" branch --show-current)" = feature/x ] && ! printf "%s" "$OUT" | grep -q "^DRIFT-SWITCHED:"'
to_land "refresh_on_recorded_base: "
sut land "$WORK/a" demo
sut cleanup "$WORK/a" demo
check "refresh_on_recorded_base: a's main checkout on its recorded base is refreshed, not refused" \
  'printf "%s\n" "$OUT" | grep -qE "^REFRESH-(DONE|CURRENT): $WORK/a( |\$)" && ! printf "%s\n" "$OUT" | grep -q "^REFRESH-REFUSED: $WORK/a "'
sut cleanup "$WORK/a" demo
check "refresh_on_recorded_base: a cleanup re-run after worktree removal still refreshes onto feature/x" \
  'printf "%s\n" "$OUT" | grep -qE "^REFRESH-(DONE|CURRENT): $WORK/a( |\$)" && ! printf "%s\n" "$OUT" | grep -q "^REFRESH-REFUSED: $WORK/a "'`},
		{"orphan_sweep", `# --- orphan_sweep ---
fixture
printf '\n## workspace isolation\n\n| Command | Runs |\n|---------|------|\n| ` + "`" + `remove` + "`" + ` | ` + "`" + `echo <id> >>%s/removed.log; [ <id> != orph-bad ]` + "`" + ` |\n| ` + "`" + `orphans` + "`" + ` | ` + "`" + `printf "orph-1\\norph-bad\\n"` + "`" + ` |\n' "$WORK" >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" commit -qam "chore: workspace commands" && g -C "$WORK/a" push -q origin main
to_land "orphan_sweep: "
sut land "$WORK/a" demo
sut cleanup "$WORK/a" demo
check "orphan_sweep: the change's own workspace is removed" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-REMOVED: $WORK/a"'
check "orphan_sweep: each orphan id goes through the remove command" '[ "$(cat "$WORK/removed.log")" = "$(printf "demo-id\norph-1\norph-bad")" ]'
check "orphan_sweep: a removed orphan is reported" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVED: orph-1"'
check "orphan_sweep: a failed orphan removal is reported and cleanup still finishes" \
  'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVE-FAILED: orph-bad — exit 1" && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
		{"padded_orphan", `# --- padded_orphan ---
fixture
printf '\n## workspace isolation\n\n| Command | Runs |\n|---------|------|\n| ` + "`" + `remove` + "`" + ` | ` + "`" + `echo "[<id>]" >>%s/removed.log` + "`" + ` |\n| ` + "`" + `orphans` + "`" + ` | ` + "`" + `printf " orph-1\\n"` + "`" + ` |\n' "$WORK" >>"$WORK/a/.flow/project.md"
g -C "$WORK/a" commit -qam "chore: workspace commands" && g -C "$WORK/a" push -q origin main
to_land "padded_orphan: "
sut land "$WORK/a" demo
sut cleanup "$WORK/a" demo
check "padded_orphan: cleanup exits 0" '[ "$RC" -eq 0 ]'
check "padded_orphan: a padded orphan id is refused and never removed" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVE-FAILED:  orph-1 — refused, not an id" && ! grep -q "orph-1" "$WORK/removed.log"'`},
	})
}
