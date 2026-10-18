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
#   narrative_format  an `In-run pipeline fix:` line off the pipeline.md
#                   format stops commit at STOP: narrative, exit 1; the line
#                   rewritten to the format passes.
#   foreign_staged  a main checkout with staged work stops prepare at
#                   STOP: foreign-staged, exit 1, before the preflight.
#   no_report       land with no committed self-review report stops at
#                   STOP: self-review-report, lands nothing and pushes no
#                   commit made after `commit`.
#   loop_landing    land with an own-branch self-review loop's rows refusing
#                   the landing stops at STOP: loop-landing, exit 1, before
#                   the first landing push; nothing lands on either main.
#   untasked        a trailerless code commit in b stops prepare at
#                   STOP: unfinished-work, its subject listed; the
#                   --accept-outstanding re-run passes the gate.
#   archived_rerun  prepare re-run after commit and a later task commit:
#                   b's reshaped implementation commit carries commit-split's
#                   Flow-Key, so the gate does not list it.
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
#   no_state_dir    the state directory `flow state dir` prints does not
#                   exist (a store-only project): run 2 still reaches
#                   COMPLETE and writes FINISHED, not STOP: leftover.
#   pr_url          the pull-request route with no usable gh: a --pr-url for
#                   b does not answer for a.
#   nested_link     b's branch adds only spectre/changes/archive/y/link.md,
#                   which commit-split's `*/link.md` stages: prepare asks a
#                   subject for b, and commit with it lands that link.md in
#                   b's implementation commit.
#   mixed_roots     b's tree is openspec/ while a's is spectre/: prepare's
#                   diffstat for b leaves out b's openspec/changes/.
#   planning_only   b's branch changes nothing outside spectre/changes/:
#                   prepare's NEXT asks no subject for b, and commit with no
#                   subject for b commits b's planning delta alone.
#   satellite_archive b's branch carries a satellite's
#                   spectre/changes/demo/link.md: commit archives it in b
#                   and commits that as b's own `chore(spectre): archive
#                   demo` (KAN-815).
#   link_only_canon a's branch changes nothing outside spectre/changes/ but
#                   its change directory's link.md, the merge-order record:
#                   prepare's NEXT asks no subject for a, and commit with no
#                   subject for a lands that link.md in an implementation
#                   commit subjected `chore(spectre): merge order`.
#   commit_pushes   both branches were pushed before the reshape: commit
#                   leaves each remote spectre/demo at its reshaped HEAD, so
#                   a later plain branch-backup push is a fast-forward.
#   peer_fix        the report names a fix committed in b, the
#                   non-canonical worktree, and b's base moves so its land
#                   push is rejected and the re-sync rewrites the fix: the
#                   report on a's origin main names the rewritten sha, and
#                   land records a `fixed` row for it.
#   peer_fix_canon_first  peer_fix with the merge order `.` then `b`: the
#                   rewrite lands after a was pushed, and land pushes a's
#                   branch and base again in the same run.
#   pre_reshape_pin  commit stopped after the reshape, then re-run: each
#                   repository's refs/flow/pre-reshape/demo stays at its
#                   pre-reshape HEAD, reaches the Task-Id: 1 commit, and is
#                   on origin; a pin on origin the local one does not
#                   descend from stops commit at STOP: pre-reshape, untouched. happy_path checks
#                   the pin survives run 2.
#   rerun_pin       an archived re-run with a new Task-Id: 2 commit in a:
#                   commit advances a's pin so it reaches both Task-Id
#                   commits and pushes it to origin; with the local pin
#                   deleted first (a re-clone), commit fetches origin's and
#                   does the same rather than stopping.
#   visual_stop     a declares `ui paths` its task commit touches and no
#                   verifier saw that HEAD: prepare stops at
#                   STOP: visual-verify, and --accept-outstanding does not
#                   pass it; neither does a false-positive flag on an
#                   earlier HEAD, on another worktree's row, or on another
#                   change. A flag on this worktree's row for this HEAD, or
#                   a verifier row whose diffBase is that HEAD, passes.
#                   The stop comes at the first MISSING worktree, no later
#                   one judged, so its row is the newest a flag lands on.
#   visual_archived an archived re-run with no new work still runs the
#                   gate: a flag on the pre-archive HEAD does not pass it.
#   visual_rebased  a base move in prepare rebases a: the gate runs again
#                   on the rebased HEAD and stops.
#   visual_resync   land's push of a is rejected and the re-sync makes a
#                   HEAD no verifier saw: land stops before a's push,
#                   naming b as already pushed; re-running land stops again.
#   visual_uncommitted an uncommitted UI edit on a verified HEAD is folded in
#                   by commit: on each route, commit stops before a's branch
#                   push and land before any push from a.
#   base_mismatch   a recording flowBase feature/x while b resolves main and
#                   b's origin has feature/x stops prepare at
#                   STOP: base-mismatch, exit 1, before the preflight; b
#                   recording it too, or feature/x on a's origin only, passes.
#   outstanding     commit --outstanding keeps the planning commit's own
#                   subject and puts the text in its body; --plan-message,
#                   which replaced the subject whole, is refused.
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

g() { GIT_AUTHOR_DATE=2026-01-01T00:00:00Z GIT_COMMITTER_DATE=2026-01-01T00:00:00Z git -c user.name=Test -c user.email=test@example.invalid -c commit.gpgsign=false "$@"; }

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
    g -C "$WORK/$r-worktrees/demo" commit -qm "feat($r): task 1" -m "Task-Id: 1"
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
  "self-review bundle") printf 'found: 1 of 1 sources; skipped: 0 of 1 sources\n' ;;
  "self-review finding") echo "\$*" >>"$WORK/fixed.log" ;;
  "record dispatches") [ -f "$WORK/loop-rows" ] && exec cat "$WORK/loop-rows"; cat "$WORK/dispatches.json" 2>/dev/null || echo '[{"key":"task-1-implementer","outcome":"completed"},{"key":"verify","role":"verifier","outcome":"completed","startedAt":"2026-01-02T00:00:00Z"}]' ;;
  "record decisions") echo '[]' ;;
  "record verdicts") cat "$WORK/verdicts.json" 2>/dev/null || echo '[]' ;;
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

# --- narrative_format ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
nar="$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
printf '\n## integrate run\n\n- In-run pipeline fix: the renderer could not render an exp- slot\n' >>"$nar"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_format: an off-format in-run fix line stops commit with exit 1" '[ "$RC" -eq 1 ]'
check "narrative_format: the stop names the off-format line" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*the renderer could not render an exp- slot"'
# narrative_passes <label> <line>: a fresh prepare, <line> as the narrative's
# in-run fix entry, and commit must pass.
narrative_passes() {
  fixture
  sut prepare "$WORK/a" demo
  [ "$RC" -eq 0 ] || halt prepare
  printf '\n## integrate run\n\n%s\n' "$2" >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
  check "narrative_format: $1 passes commit" '[ "$RC" -eq 0 ]'
}
narrative_passes "a line in the format" '- In-run pipeline fix: 1a2b3c4 — the renderer could not render an exp- slot (blast radius 2 files)'
narrative_passes "a prose mention of the token" 'The `In-run pipeline fix:` line was missing a sha.'
narrative_passes "a well-formed line with trailing whitespace" '- In-run pipeline fix: 1a2b3c4 — the defect (blast radius 2 files)   '
narrative_passes "a deferred line" '- In-run pipeline fix: deferred — the defect (blast radius 1 file)'
narrative_passes "a backticked sha" '- In-run pipeline fix: `1a2b3c4` — the defect (blast radius 3 files).' 

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
check "run 2 keeps a's pre-reshape pin, locally and on origin" \
  'git -C "$WORK/a" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null && git --git-dir="$WORK/a-origin.git" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null'
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
pushed="$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)"
printf 'later\n' >"$WORK/a-worktrees/demo/later.txt"
g -C "$WORK/a-worktrees/demo" add later.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): later" -m "Task-Id: 2"
sut land "$WORK/a" demo
check "no_report: land stops with exit 1" '[ "$RC" -eq 1 ]'
check "no_report: the stop is STOP: self-review-report" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: self-review-report"'
check "no_report: nothing was landed" \
  '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ]'
check "no_report: a's remote branch still sits at commit's HEAD" \
  '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$pushed" ]'

# --- loop_landing ---
fixture
to_land "loop_landing: "
printf '[{"seq":1,"key":"self-review-demo-plan","outcome":"","sessionToken":"self-review-demo"}]\n' >"$WORK/loop-rows"
sut land "$WORK/a" demo
check "loop_landing: an open own-branch self-review loop stops land at STOP: loop-landing" \
  '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: loop-landing"'
check "loop_landing: nothing landed on either main" \
  '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ]'

# --- untasked ---
fixture
printf 'stray\n' >"$WORK/b-worktrees/demo/stray.txt"
g -C "$WORK/b-worktrees/demo" add stray.txt && g -C "$WORK/b-worktrees/demo" commit -qm "test(b): a stray edit"
sut prepare "$WORK/a" demo
check "untasked: prepare stops at STOP: unfinished-work" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: unfinished-work"'
check "untasked: the stray commit is listed" 'printf "%s" "$OUT" | grep -q "^  [0-9a-f]\{12\} test(b): a stray edit$"'
sut prepare "$WORK/a" demo --accept-outstanding
check "untasked: --accept-outstanding passes the gate" '[ "$RC" -eq 0 ]'

# --- archived_rerun ---
fixture
to_land "archived_rerun: "
printf 'more\n' >"$WORK/a-worktrees/demo/more.txt"
g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): more" -m "Task-Id: 2"
sut prepare "$WORK/a" demo
check "archived_rerun: prepare passes the gate" '[ "$RC" -eq 0 ]'
check "archived_rerun: b's reshaped commit is not listed" '! printf "%s" "$OUT" | grep -q "feat(b): do the work"'

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
g -C "$WORK/b-worktrees/demo" add gt.txt && g -C "$WORK/b-worktrees/demo" commit -qm "feat(b): gt" -m "Task-Id: 2"
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

# --- no_state_dir ---
fixture
to_land "no_state_dir: "
rm -rf "$WORK/statedir"
sut land "$WORK/a" demo
check "no_state_dir: run 2 reaches COMPLETE, not STOP: leftover" \
  '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | grep -q "^COMPLETE: " && ! printf "%s" "$OUT" | grep -q "^STOP: leftover"'
check "no_state_dir: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'

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

# --- nested_link ---
# commit-split adds `<spec>/changes/*/link.md` only when a change directory
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
  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/y/link.md)" = "feat(b): nest a link" ]'

# --- mixed_roots ---
fixture
mkdir -p "$WORK/b-worktrees/demo/openspec/changes/demo"
printf 'notes\n' >"$WORK/b-worktrees/demo/openspec/changes/demo/notes.md"
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(openspec): notes"
sut prepare "$WORK/a" demo
check "mixed_roots: prepare exits 0" '[ "$RC" -eq 0 ]'
check "mixed_roots: b's diffstat lists no openspec/changes/ path" '! printf "%s" "$OUT" | grep -q "openspec/changes/"'

# --- planning_only ---
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
  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD | tr "\n" "|")" = "chore(spectre): archive demo|chore(spectre): plan demo|" ]'

# --- satellite_archive ---
fixture
mkdir -p "$WORK/b-worktrees/demo/spectre/changes/demo"
printf '## Part of\n\n`a:demo`\n' >"$WORK/b-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): link demo"
to_land "satellite_archive: " no-report
check "satellite_archive: b's change directory is archived" \
  '[ ! -d "$WORK/b-worktrees/demo/spectre/changes/demo" ] && [ -f "$WORK/b-worktrees/demo/spectre/changes/archive/demo/link.md" ]'
check "satellite_archive: b's branch ends on its own archive commit" \
  '[ "$(git -C "$WORK/b-worktrees/demo" log -1 --format=%s)" = "chore(spectre): archive demo" ]'

# --- link_only_canon ---
fixture
g -C "$WORK/a-worktrees/demo" reset -q --hard main
c="$WORK/a-worktrees/demo/spectre/changes/demo"
mkdir -p "$c"
printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
printf '# demo\n\n## Merge order\n\n1. `b`\n2. `.`\n' >"$c/link.md"
g -C "$WORK/a-worktrees/demo" add -A && g -C "$WORK/a-worktrees/demo" commit -qm "chore(spectre): merge order"
sut prepare "$WORK/a" demo
check "link_only_canon: prepare exits 0" '[ "$RC" -eq 0 ]'
check "link_only_canon: prepare's NEXT asks a subject for b and none for a" \
  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>" && ! printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo="'
printf '\n## integrate run\n\nnarrative\n' >>"$c/narrative.md"
sut commit "$WORK/a" demo "$WORK/b-worktrees/demo=feat(b): do the work"
check "link_only_canon: commit with no subject for a exits 0" '[ "$RC" -eq 0 ]'
check "link_only_canon: a's link.md lands in a commit subjected chore(spectre): merge order" \
  '[ "$(git -C "$WORK/a-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/demo/link.md spectre/changes/demo/link.md)" = "chore(spectre): merge order" ]'

# --- commit_pushes ---
fixture
for r in a b; do g -C "$WORK/$r-worktrees/demo" push -q -u origin spectre/demo; done
to_land "commit_pushes: " no-report
check "commit_pushes: each remote branch is the reshaped HEAD" \
  '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)" ] && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)" ]'

# peer_fix_case <prefix>: the report names a fix committed in b, b's base
# moves, and land runs — checked under <prefix>.
peer_fix_case() {
  to_land "$1" no-report
  printf 'fix\n' >"$WORK/b-worktrees/demo/fix-b.txt"
  g -C "$WORK/b-worktrees/demo" add fix-b.txt && g -C "$WORK/b-worktrees/demo" commit -qm "fix(b): the peer fix"
  old="$(git -C "$WORK/b-worktrees/demo" rev-parse --short HEAD)"
  mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
  printf '# demo self-review\n\n- **[flow-fix]** the peer fix — fixed: %s\n' "$old" >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
  g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
  g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
  printf 'moved\n' >"$WORK/b/moved.txt"
  g -C "$WORK/b" add moved.txt && g -C "$WORK/b" commit -qm "base: moved" && g -C "$WORK/b" push -q origin main
  sut land "$WORK/a" demo
  new="$(git --git-dir="$WORK/b-origin.git" log --format=%h --grep="^fix(b): the peer fix$" -1 main)"
  check "${1}land exits 0" '[ "$RC" -eq 0 ]'
  check "${1}the re-sync rewrote b's fix sha" '[ -n "$new" ] && [ "$new" != "$old" ]'
  check "${1}the report on a's origin main names the rewritten sha" \
    'git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $new$"'
  check "${1}a fixed row is recorded for b's fix" \
    'grep -q -- "-disposition fixed -ref $new " "$WORK/fixed.log"'
}

# --- peer_fix ---
fixture
peer_fix_case "peer_fix: "

# --- peer_fix_canon_first ---
fixture
printf '# demo\n\n## Parts\n\n`b:demo`\n\n## Branch\n\nmain\n\n## Merge order\n\n1. `.`\n2. `b`\n' \
  >"$WORK/a-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
peer_fix_case "peer_fix_canon_first: "

# --- pre_reshape_pin ---
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
  'printf "%s" "$OUT" | grep -q "^STOP: pre-reshape" && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$stray" ]'

# --- rerun_pin ---
rerun_pin_case() {
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
rerun_pin_case "rerun_pin: "
rerun_pin_case "rerun_pin reclone: " reclone

# --- visual_stop ---
# ui_a: a declares `ui paths` its task commit touches.
ui_a() {
  printf '\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `work-a.txt` |\n' >>"$WORK/a-worktrees/demo/.flow/project.md"
  g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
}
# flag <worktree> <sha> [<change>]: an operator's false-positive flag on
# <worktree>'s MISSING verdict row naming HEAD <sha>.
flag() {
  printf '[{"change":"%s","guard":"check-visual-verify-dispatched","worktree":"%s","verdict":"VISUAL-VERIFY-MISSING: no verifier — HEAD %s","falsePositive":true,"falsePositiveReason":"hand-verified","flaggedAt":"2026-01-01T00:00:00Z"}]' \
    "${3:-demo}" "$(cd "$1" && pwd -P)" "$2" >"$WORK/verdicts.json"
}
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
check "visual_stop: a verifier of the HEAD passes the gate" '[ "$RC" -eq 0 ]'

# --- visual_archived: an archived re-run with no new work runs the gate ---
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
check "visual_archived: a flag on the pre-archive HEAD does not pass the gate" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'

# --- visual_rebased: a base move in prepare re-runs the gate ---
fixture
ui_a
A="$WORK/a-worktrees/demo"
flag "$A" "$(git -C "$A" rev-parse HEAD)"
printf 'moved\n' >"$WORK/a/moved.txt"
g -C "$WORK/a" add moved.txt && g -C "$WORK/a" commit -qm "base: moved" && g -C "$WORK/a" push -q origin main
sut prepare "$WORK/a" demo
check "visual_rebased: a was rebased" 'printf "%s\n" "$OUT" | grep -q "^REBASED: "'
check "visual_rebased: the rebased HEAD no verifier saw stops prepare" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — rebased onto origin/main after the gate: .*a-worktrees/demo: "'

# --- visual_resync: land's re-sync re-runs the gate before that push ---
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
check "visual_resync: the re-run did not push a's unverified HEAD" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'

# --- visual_resync_reread: land's re-sync rewrites the canonical report first ---
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
check "visual_resync_reread: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'

# --- visual_reread: reread_fix_shas's canonical re-push runs the gate ---
# The report is a's declared UI path, so the rewrite commit alone — all a's
# HEAD adds past its landed push — is a UI change no verifier saw.
fixture
printf '# demo\n\n## Parts\n\n`b:demo`\n\n## Branch\n\nmain\n\n## Merge order\n\n1. `.`\n2. `b`\n' \
  >"$WORK/a-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
printf '\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `docs/self-review/demo-self-review.md` |\n' >>"$WORK/a-worktrees/demo/.flow/project.md"
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
  'git --git-dir="$WORK/a-origin.git" show spectre/demo:docs/self-review/demo-self-review.md | grep -q "fixed: $old$" && git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $old$"'

# --- visual_uncommitted: an uncommitted UI edit folded in by commit stops every push ---
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
done

# --- base_mismatch ---
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

# --- outstanding ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" --plan-message "Task 3 is unticked"
check "outstanding: --plan-message is refused" '[ "$RC" -ne 0 ] && grep -q "unknown argument .--plan-message." "$WORK/stderr"'
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" --outstanding "Task 3 is unticked"
check "outstanding: commit exits 0" '[ "$RC" -eq 0 ]'
check "outstanding: the planning commit keeps its own subject and carries the text in its body" \
  '[ "$(g -C "$WORK/a-worktrees/demo" log --format=%s --grep="^Outstanding: " -1)" = "chore(spectre): plan demo" ] && [ "$(g -C "$WORK/a-worktrees/demo" log --format=%b --grep="^Outstanding: " -1)" = "Outstanding: Task 3 is unticked" ]'

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
