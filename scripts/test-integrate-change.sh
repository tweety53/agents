#!/usr/bin/env bash
# test-integrate-change.sh — assertion harness for integrate-change.sh.
#
# Every fixture is a real two-repository change under mktemp -d: two main
# checkouts `a` (canonical, its link.md's `## Merge order` naming `b`,
# resolved through spectre/peers, before `.`) and `b`, each with a bare origin, an apply
# worktree on spectre/demo beside it at <repo>-worktrees/demo, and a task
# commit there. The REAL integrate-change.sh runs the REAL guards it
# sequences; only the two external CLIs are stubs on PATH — `flow` (the
# state record a JSON file, every stage mark appended to a log) and
# `spectre` (`archive` a git mv) — since a store and the spectre binary are
# external dependencies no harness case may have.
#
# Cases:
#   happy_path      prepare -> commit -> land on the configured merge-and-push
#                   route, chained into cleanup: both repositories' origin
#                   main carries the implementation commit (a's also the
#                   planning and archive commits), the worktrees are gone,
#                   the record is FINISHED with an empty map, and the stage
#                   log runs from flow.preflight to flow.refresh-main-checkout.
#   foreign_staged  a main checkout with staged work stops prepare at
#                   STOP: foreign-staged, exit 1, before the preflight.
#   no_report       land with no committed self-review report stops at
#                   STOP: self-review-report and pushes nothing.
#   no_verdict      a store outage leaves check-unfinished-work without a
#                   verdict: STOP: no-verdict, which --accept-outstanding
#                   does not pass.
#   sync_conflict   a conflicting base stops prepare at STOP: sync-conflict,
#                   its line naming the --resume call whole.
#   guard_test      a failing scoped re-verification stops prepare at
#                   STOP: guard-test with the test's own output relayed.
#   gh_open_merged  gh reports the recorded PR OPEN but the branch is on
#                   origin/main: run 2 still runs to FINISHED.
#   disclose        both apply worktrees hold an unpreserved ignored file:
#                   --proceed <a> removes a and stops again for b, which
#                   stays on disk until --proceed <b>.
#   rerun_leftover  a run 2 stopped on a leftover after removing both
#                   worktrees; the re-run still reaches repository b.
#   pr_url          the pull-request route with no usable gh: a --pr-url for
#                   b does not answer for a.
#
# Bash 3.2 is the floor: indexed arrays only.
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUT="$SCRIPT_DIR/integrate-change.sh"
PASS=0
FAIL=0
pass() { PASS=$((PASS + 1)); printf 'ok - %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf 'FAIL - %s\n' "$1" >&2; }
check() { if eval "$2"; then pass "$1"; else fail "$1"; fi; }

WORK=""
cleanup_work() { [ -n "$WORK" ] && rm -rf "$WORK"; return 0; }
trap cleanup_work EXIT

g() { git -c user.name=Test -c user.email=test@example.invalid -c commit.gpgsign=false "$@"; }

# fixture: builds $WORK with repos a and b, the stubs, and the state record.
fixture() {
  cleanup_work
  WORK="$(cd "$(mktemp -d)" && pwd -P)"
  local r
  for r in a b; do
    g init -q --bare -b main "$WORK/$r-origin.git"
    g init -q -b main "$WORK/$r"
    printf 'base\n' >"$WORK/$r/base.txt"
    if [ "$r" = a ]; then
      mkdir -p "$WORK/a/.flow" "$WORK/a/spectre/changes"
      printf '# project\n\n## default landing route\n\nmerge and push\n' >"$WORK/a/.flow/project.md"
      printf 'b ../b\n' >"$WORK/a/spectre/peers"
    fi
    g -C "$WORK/$r" add -A && g -C "$WORK/$r" commit -qm base
    g -C "$WORK/$r" remote add origin "$WORK/$r-origin.git"
    g -C "$WORK/$r" push -q -u origin main
    g -C "$WORK/$r" remote set-head origin main
    g -C "$WORK/$r" worktree add -q -b spectre/demo "$WORK/$r-worktrees/demo" main
    printf 'work in %s\n' "$r" >"$WORK/$r-worktrees/demo/work-$r.txt"
    g -C "$WORK/$r-worktrees/demo" add -A
    g -C "$WORK/$r-worktrees/demo" commit -qm "feat($r): task 1"
  done
  local c="$WORK/a-worktrees/demo/spectre/changes/demo"
  mkdir -p "$c"
  printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
  printf '# demo\n\n## Parts\n\n`b:demo`\n\n## Branch\n\nmain\n\n## Merge order\n\n1. `b`\n2. `.`\n' >"$c/link.md"
  g -C "$WORK/a-worktrees/demo" add -A
  g -C "$WORK/a-worktrees/demo" commit -qm "chore(spectre): plan-gate demo"

  mkdir -p "$WORK/bin" "$WORK/statedir"
  cat >"$WORK/bin/flow" <<EOF
#!/usr/bin/env bash
case "\$1 \$2" in
  "state get") cat "$WORK/state.json" ;;
  "state set") cat >"$WORK/state.json" ;;
  "state dir") echo "$WORK/statedir" ;;
  "state list") echo '{"source":"store","complete":true,"records":[{"name":"demo"}]}' ;;
  "stage begin"|"stage end") echo "\$*" >>"$WORK/stages.log" ;;
  "record findings") [ -f "$WORK/fail-findings" ] && exit 1; echo '[]' ;;
  "self-review findings") echo '[]' ;;
  "workspace-id "*) echo demo-id ;;
  *) : ;;
esac
EOF
  cat >"$WORK/bin/spectre" <<'EOF'
#!/usr/bin/env bash
[ "$1" = archive ] || exit 1
mkdir -p spectre/changes/archive && git mv "spectre/changes/$2" "spectre/changes/archive/$2"
EOF
  chmod +x "$WORK/bin/flow" "$WORK/bin/spectre"
  printf '{"state":"IN_PROGRESS","branch":"spectre/demo","worktrees":{"%s":"%s","%s":"%s"},"jiraIssue":"KAN-1","prUrl":null,"updatedBy":"/flow"}\n' \
    "$WORK/a-worktrees/demo" "$(git -C "$WORK/a" rev-parse main)" \
    "$WORK/b-worktrees/demo" "$(git -C "$WORK/b" rev-parse main)" >"$WORK/state.json"
}

# halt <step>: a step that did not exit 0 leaves nothing for the next to
# judge — print its output and end the run failed.
halt() {
  printf '%s\n%s\n' "$OUT" "$(cat "$WORK/stderr")" >&2
  fail "$1 did not complete; the happy path stops here"
  printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
  exit 1
}

sut() {
  OUT="$(PATH="$WORK/bin:$PATH" GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.invalid \
    GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.invalid \
    "$SUT" "$@" --harness test --session-token mf-test 2>"$WORK/stderr")"
  RC=$?
}

# to_land [<prefix>]: prepare and commit, checked under <prefix>, then the
# self-review report committed in the canonical worktree — the state land
# expects.
to_land() {
  sut prepare "$WORK/a" demo
  check "${1}prepare exits 0" '[ "$RC" -eq 0 ]'
  [ "$RC" -eq 0 ] || halt prepare
  printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
  check "${1}commit exits 0" '[ "$RC" -eq 0 ]'
  [ "$RC" -eq 0 ] || halt commit
  [ "${2:-}" = no-report ] && return 0
  mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
  printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
  g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
  g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
}

# --- happy_path ---
fixture
sut prepare "$WORK/a" demo
check "prepare exits 0" '[ "$RC" -eq 0 ]'
check "prepare names the configured route" 'printf "%s" "$OUT" | grep -q "^ROUTE: merge and push — from this project"'
check "prepare prints each worktree's reshape base" '[ "$(printf "%s\n" "$OUT" | grep -c "^WORKTREE: ")" -eq 2 ]'
check "prepare ends on NEXT: commit" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: append this run.s narrative"'
[ "$RC" -eq 0 ] || halt prepare

printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "commit exits 0" '[ "$RC" -eq 0 ]'
check "commit ends on NEXT: self-review" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run the self-review pass"'
[ "$RC" -eq 0 ] || halt commit
mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"

sut land "$WORK/a" demo
check "land exits 0" '[ "$RC" -eq 0 ]'
[ "$RC" -eq 0 ] || halt land
check "a's origin main carries implementation, plan, archive and report commits" \
  '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -4 | tr "\n" "|")" = "docs(self-review): demo self-review report|chore(spectre): archive demo|chore(spectre): plan demo|feat(a): do the work|" ]'
check "b's origin main carries its implementation commit" \
  '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = "feat(b): do the work" ]'
check "b was pushed before a (merge order)" \
  'printf "%s\n" "$OUT" | grep "^PUSHED: " | head -1 | grep -q "/b-worktrees/demo"'
check "both worktrees are removed" '[ ! -d "$WORK/a-worktrees/demo" ] && [ ! -d "$WORK/b-worktrees/demo" ]'
check "the record is FINISHED with an empty worktrees map" \
  '[ "$(jq -c "[.state, .worktrees]" "$WORK/state.json")" = "[\"FINISHED\",{}]" ]'
check "Jira In Review then Done are handed to the parent" \
  '[ "$(printf "%s\n" "$OUT" | grep "^JIRA: " | tr "\n" "|")" = "JIRA: transition KAN-1 to In Review|JIRA: transition KAN-1 to Done|" ]'
check "the stage log runs preflight through refresh-main-checkout" \
  'head -1 "$WORK/stages.log" | grep -q "stage begin .*-stage flow.preflight" && tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.refresh-main-checkout -outcome completed"'
check "land ends on NEXT: Finished handoff" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'

# --- foreign_staged ---
fixture
printf 'stray\n' >"$WORK/b/stray.txt"
g -C "$WORK/b" add stray.txt
sut prepare "$WORK/a" demo
check "staged work in a main checkout stops prepare with exit 1" '[ "$RC" -eq 1 ]'
check "the stop is STOP: foreign-staged" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: foreign-staged"'
check "the guard's own STAGED-FOREIGN line is relayed" 'printf "%s" "$OUT" | grep -q "^STAGED-FOREIGN: $WORK/b"'
check "no preflight ran" '! grep -q flow.preflight "$WORK/stages.log" 2>/dev/null'

# --- no_report ---
fixture
to_land "no_report: " no-report
sut land "$WORK/a" demo
check "no_report: land stops with exit 1" '[ "$RC" -eq 1 ]'
check "no_report: the stop is STOP: self-review-report" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: self-review-report"'
check "no_report: nothing was pushed" \
  '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ] && ! git --git-dir="$WORK/b-origin.git" rev-parse -q --verify refs/heads/spectre/demo >/dev/null'

# --- no_verdict ---
fixture
: >"$WORK/fail-findings"
sut prepare "$WORK/a" demo --accept-outstanding
check "no_verdict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "no_verdict: the stop is STOP: no-verdict, not passed by --accept-outstanding" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: no-verdict — check-unfinished-work"'
check "no_verdict: the gate mark closes stopped" 'tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.unfinished-work-gate -outcome stopped"'

# --- sync_conflict ---
fixture
printf 'base side\n' >"$WORK/b/work-b.txt"
g -C "$WORK/b" add work-b.txt && g -C "$WORK/b" commit -qm "base: work-b" && g -C "$WORK/b" push -q origin main
sut prepare "$WORK/a" demo
check "sync_conflict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "sync_conflict: the stop line names the --resume call and what follows it" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: sync-conflict — .*sync-onto-base.sh. --resume, the whole ## lint and ## test lists, then re-run prepare$"'

# --- guard_test ---
# A scoped re-verification is <agents repo>/scripts/test-<stem>.sh, the repo
# derived from the guards' own path: run a copy of scripts/ whose stats/ is
# this checkout's, holding a failing test-gt.sh, so no test lands in the tree.
fixture
mkdir -p "$WORK/agents"
cp -R "$SCRIPT_DIR" "$WORK/agents/scripts"
ln -s "$(cd "$SCRIPT_DIR/../stats" && pwd -P)" "$WORK/agents/stats"
printf '#!/usr/bin/env bash\necho gt-test-own-output\nexit 1\n' >"$WORK/agents/scripts/test-gt.sh"
printf 'same\n' >"$WORK/b-worktrees/demo/gt.txt"
g -C "$WORK/b-worktrees/demo" add gt.txt && g -C "$WORK/b-worktrees/demo" commit -qm "feat(b): gt"
printf 'same\n' >"$WORK/b/gt.txt"
g -C "$WORK/b" add gt.txt && g -C "$WORK/b" commit -qm "base: gt" && g -C "$WORK/b" push -q origin main
SUT_SAVED="$SUT"; SUT="$WORK/agents/scripts/integrate-change.sh"
sut prepare "$WORK/a" demo
SUT="$SUT_SAVED"
check "guard_test: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
check "guard_test: the stop is STOP: guard-test" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: guard-test"'
check "guard_test: the failing test's own output is relayed" 'printf "%s" "$OUT" | grep -qx "gt-test-own-output"'

# --- gh_open_merged ---
fixture
printf '#!/usr/bin/env bash\n[ "$1 $2" = "pr view" ] && { echo OPEN; exit 0; }\nexit 1\n' >"$WORK/bin/gh"
chmod +x "$WORK/bin/gh"
jq -c '.prUrl = "https://forge.invalid/pr/1"' "$WORK/state.json" >"$WORK/state.tmp" && mv "$WORK/state.tmp" "$WORK/state.json"
to_land "gh_open_merged: "
sut land "$WORK/a" demo
check "gh_open_merged: land runs to the Finished handoff" '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
check "gh_open_merged: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'

# --- disclose ---
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
check "disclose: --proceed for both finishes" '[ "$RC" -eq 0 ] && [ ! -d "$WORK/b-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'

# --- rerun_leftover ---
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
check "rerun_leftover: the re-run writes FINISHED" '[ "$RC" -eq 0 ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'

# --- pr_url ---
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
  '[ "$RC" -eq 0 ] && [ "$(jq -r "[.state, .prUrl] | join(\" \")" "$WORK/state.json")" = "IN_PROGRESS https://forge.invalid/a/1" ]'

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
