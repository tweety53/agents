#!/usr/bin/env bash
# Assertion harness for state-advance.sh. Every case runs against a state file
# in a sandboxed directory; the real state tree under ~/Agents/myflow is never
# read or written.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ADVANCE="$SCRIPT_DIR/../skills/myflow-state-advance/state-advance.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

new_state() {
  SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/state-test.XXXXXX")"
  STATE_FILE="$SANDBOX/demo.json"
  cat > "$STATE_FILE" <<'JSON'
{
  "stage": "awaiting-do-review",
  "gates": { "reviewed": true, "tested": "skipped", "prOpened": null, "prMerged": null },
  "worktree": null,
  "branch": "openspec/demo",
  "originStage": null,
  "artifactUrl": "https://example.invalid/a",
  "jiraIssue": "KAN-10",
  "fastPath": true,
  "REVIEWED_TREE": null,
  "MERGE_BASE": { "/tmp/wt": "abc123" },
  "updatedAt": "2026-07-01T00:00:00Z",
  "updatedBy": "/myflow-do"
}
JSON
}

run_advance() {
  set +e
  OUT="$(MYFLOW_STATE_FILE="$STATE_FILE" "$ADVANCE" "$@" 2>&1)"
  RC=$?
  set -e
}

# 1. Happy path writes stage, updatedAt, updatedBy.
new_state
run_advance --name demo --target do-done \
  --accepted awaiting-do-review,do-review-started --by /myflow-do-done
[ "$RC" -eq 0 ] && pass "happy path exits 0" || fail "happy path: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "do-done" ] \
  && pass "stage written" || fail "stage not written"
[ "$(jq -r '.updatedBy' "$STATE_FILE")" = "/myflow-do-done" ] \
  && pass "updatedBy written" || fail "updatedBy not written"
[ "$(jq -r '.updatedAt' "$STATE_FILE")" != "2026-07-01T00:00:00Z" ] \
  && pass "updatedAt refreshed" || fail "updatedAt not refreshed"

# 2. Unowned fields are carried forward verbatim.
[ "$(jq -r '.gates.tested' "$STATE_FILE")" = "skipped" ] \
  && pass "gates.tested carried" || fail "gates.tested lost"
[ "$(jq -r '.gates.reviewed' "$STATE_FILE")" = "true" ] \
  && pass "gates.reviewed carried" || fail "gates.reviewed lost"
[ "$(jq -r '.jiraIssue' "$STATE_FILE")" = "KAN-10" ] \
  && pass "jiraIssue carried" || fail "jiraIssue lost"
[ "$(jq -r '.artifactUrl' "$STATE_FILE")" = "https://example.invalid/a" ] \
  && pass "artifactUrl carried" || fail "artifactUrl lost"
[ "$(jq -r '.fastPath' "$STATE_FILE")" = "true" ] \
  && pass "fastPath carried" || fail "fastPath lost"
[ "$(jq -r '.MERGE_BASE["/tmp/wt"]' "$STATE_FILE")" = "abc123" ] \
  && pass "MERGE_BASE carried" || fail "MERGE_BASE lost"

# 3. Stage mismatch exits 4 and writes nothing.
new_state
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target review-done --accepted awaiting-pr-review --by /myflow-review-done
[ "$RC" -eq 4 ] && pass "mismatch exits 4" || fail "mismatch: rc=$RC"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "mismatch writes nothing" || fail "mismatch mutated the file"

# 4. Missing state file exits 3.
new_state
rm -f "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "missing state exits 3" || fail "missing state: rc=$RC"
[ ! -f "$STATE_FILE" ] && pass "missing state creates nothing" || fail "file was created"

# 5. Unparseable state file exits 3.
new_state
printf 'not json' > "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "unparseable exits 3" || fail "unparseable: rc=$RC"

# 6. Stale worktree exits 3 and writes nothing.
new_state
jq '.worktree = "/nonexistent/worktree"' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "stale worktree exits 3" || fail "stale worktree: rc=$RC"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "stale worktree writes nothing" || fail "stale worktree mutated the file"

# 7. Dynamic target: do-review-started resolves to awaiting-do-review, clears originStage.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = "do-review-started"' \
  "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target originStage \
  --accepted awaiting-fix-review,fix-review-started --by /myflow-do-fix-done
[ "$RC" -eq 0 ] && pass "dynamic target exits 0" || fail "dynamic: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "awaiting-do-review" ] \
  && pass "do-review-started resolves" || fail "dynamic resolution wrong"
[ "$(jq -r '.originStage' "$STATE_FILE")" = "null" ] \
  && pass "originStage cleared" || fail "originStage not cleared"

# 8. Dynamic target: the other five origins target themselves.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = "manual-test-done"' \
  "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target originStage --accepted awaiting-fix-review --by /myflow-do-fix-done
[ "$(jq -r '.stage' "$STATE_FILE")" = "manual-test-done" ] \
  && pass "self-targeting origin" || fail "self-targeting origin wrong"

# 9. Null originStage exits 5.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = null' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target originStage --accepted awaiting-fix-review --by /myflow-do-fix-done
[ "$RC" -eq 5 ] && pass "null originStage exits 5" || fail "null originStage: rc=$RC"

# 10. Corrupt originStage exits 6 and is never repaired.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = "awaiting-fix-review"' \
  "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target originStage --accepted awaiting-fix-review --by /myflow-do-fix-done
[ "$RC" -eq 6 ] && pass "corrupt originStage exits 6" || fail "corrupt originStage: rc=$RC"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "corrupt originStage writes nothing" || fail "corrupt originStage mutated the file"

# 11. The written file is valid JSON with every original key present.
new_state
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
jq -e . "$STATE_FILE" >/dev/null 2>&1 \
  && pass "output is valid JSON" || fail "output is not valid JSON"
for key in stage gates worktree branch originStage artifactUrl jiraIssue fastPath \
           REVIEWED_TREE MERGE_BASE updatedAt updatedBy; do
  jq -e "has(\"$key\")" "$STATE_FILE" >/dev/null 2>&1 \
    || fail "key dropped: $key"
done
pass "all keys retained"

# 12. Every value-taking flag missing its value exits 2 with a non-empty message.
new_state
run_advance --target do-done --accepted awaiting-do-review --by /myflow-do-done --name
[ "$RC" -eq 2 ] && pass "--name with no value exits 2" || fail "--name with no value: rc=$RC"
[ -n "$OUT" ] && pass "--name with no value prints a message" || fail "--name with no value: empty output"

run_advance --name demo --accepted awaiting-do-review --by /myflow-do-done --target
[ "$RC" -eq 2 ] && pass "--target with no value exits 2" || fail "--target with no value: rc=$RC"
[ -n "$OUT" ] && pass "--target with no value prints a message" || fail "--target with no value: empty output"

run_advance --name demo --target do-done --by /myflow-do-done --accepted
[ "$RC" -eq 2 ] && pass "--accepted with no value exits 2" || fail "--accepted with no value: rc=$RC"
[ -n "$OUT" ] && pass "--accepted with no value prints a message" || fail "--accepted with no value: empty output"

run_advance --name demo --target do-done --accepted awaiting-do-review --by
[ "$RC" -eq 2 ] && pass "--by with no value exits 2" || fail "--by with no value: rc=$RC"
[ -n "$OUT" ] && pass "--by with no value prints a message" || fail "--by with no value: empty output"

# 13. Real path resolution (no MYFLOW_STATE_FILE override) matches the State file
# contract. The expected path is derived here by re-running the same algorithm
# skills/myflow-contracts/state-file.md documents. That algorithm is transcribed
# here rather than imported from the script, so a change to the script alone
# breaks this assertion — but a change made to BOTH in the same way would not be
# caught, since the two copies are transcriptions of one contract, not
# independent derivations of it.
REALHOME="$(mktemp -d "${TMPDIR:-/tmp}/state-test-home.XXXXXX")"
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
REAL_STATE_FILE="$REALHOME/Agents/myflow/state/$PROJECT_KEY/realpath-demo.json"
mkdir -p "$(dirname "$REAL_STATE_FILE")"
cat > "$REAL_STATE_FILE" <<'JSON'
{
  "stage": "awaiting-do-review",
  "gates": { "reviewed": true, "tested": "skipped", "prOpened": null, "prMerged": null },
  "worktree": null,
  "branch": "openspec/realpath-demo",
  "originStage": null,
  "artifactUrl": null,
  "jiraIssue": null,
  "fastPath": null,
  "REVIEWED_TREE": null,
  "MERGE_BASE": null,
  "updatedAt": "2026-07-01T00:00:00Z",
  "updatedBy": "/myflow-do"
}
JSON
set +e
OUT="$(env -u MYFLOW_STATE_FILE HOME="$REALHOME" "$ADVANCE" \
  --name realpath-demo --target do-done --accepted awaiting-do-review --by /myflow-do-done 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "real path resolution exits 0" || fail "real path resolution: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$REAL_STATE_FILE")" = "do-done" ] \
  && pass "real path resolution wrote the contract-derived path" \
  || fail "real path resolution did not write $REAL_STATE_FILE"

# 14. A worktree the state file names, probed from a cwd that is not in ANY
# repository, still passes — the probe runs against the repository that owns the
# recorded path, not against cwd's. This is the assertion the whole worktree
# check was missing: every case above uses "worktree": null, so a regression that
# escalated EVERY non-null worktree would have passed all of them.
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
new_state
jq --arg wt "$REPO_ROOT" '.worktree = $wt' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/state-test-cwd.XXXXXX")"
set +e
OUT="$(cd "$OUTSIDE" && MYFLOW_STATE_FILE="$STATE_FILE" "$ADVANCE" \
  --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "a live worktree passes when probed from an unrelated cwd" \
  || fail "live worktree from unrelated cwd: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "do-done" ] \
  && pass "live worktree run wrote the stage" || fail "live worktree run did not write"

# 15. A worktree path that exists but whose repository cannot answer (not a git
# repository at all) is UNKNOWN, not a contradiction — the contract's "a check
# that cannot be performed is not a contradiction". It must not escalate.
new_state
NOTAREPO="$(mktemp -d "${TMPDIR:-/tmp}/state-test-norepo.XXXXXX")"
jq --arg wt "$NOTAREPO" '.worktree = $wt' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 0 ] && pass "an unanswerable worktree probe does not escalate" \
  || fail "unanswerable worktree probe escalated: rc=$RC out=$OUT"

# 16. A top-level JSON array parses fine and passes `jq -e .`, but `.stage`
# against it is a jq RUNTIME error (status 5 — this script's "originStage is
# null" code). It must be caught as a structural problem and escalate as 3.
new_state
printf '["stage"]' > "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "a top-level array exits 3" || fail "top-level array: rc=$RC out=$OUT"

new_state
printf '"awaiting-do-review"' > "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "a top-level string exits 3" || fail "top-level string: rc=$RC out=$OUT"

# 17. A partial state file must escalate rather than be perpetuated. Before this
# check the jq merge turned {"stage": …} into a three-field file and exited 0.
new_state
printf '{"stage": "awaiting-do-review"}' > "$STATE_FILE"
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "a partial state file exits 3" || fail "partial state file: rc=$RC out=$OUT"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "a partial state file is not rewritten" || fail "partial state file was rewritten"

# 17b. An OPTIONAL field the contract allows to be absent rather than null must
# still be accepted. `skills/myflow-contracts/state-file.md` documents `fastPath`
# as "`null`/absent", so a required-key check demanding all twelve fields would
# escalate a file the canonical contract calls legitimate.
new_state
jq 'del(.fastPath)' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 0 ] && pass "an absent optional field is accepted" \
  || fail "absent optional field: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "do-done" ] \
  && pass "absent optional field still writes the stage" || fail "absent optional field did not write"
jq -e 'has("fastPath") | not' "$STATE_FILE" >/dev/null 2>&1 \
  && pass "an absent optional field is not invented" || fail "fastPath was invented"

# 18. `gates` present but missing a gate key is the same class of partial file.
new_state
jq 'del(.gates.prMerged)' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "a missing gate key exits 3" || fail "missing gate key: rc=$RC out=$OUT"

# 19. --name is validated before it reaches a path. A traversing name must exit
# 2 and create nothing outside the state directory.
SANDBOX_HOME="$(mktemp -d "${TMPDIR:-/tmp}/state-test-name.XXXXXX")"
set +e
OUT="$(env -u MYFLOW_STATE_FILE HOME="$SANDBOX_HOME" "$ADVANCE" \
  --name '../../../../victim/other' --target do-done \
  --accepted awaiting-do-review --by /myflow-do-done 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "a traversing --name exits 2" || fail "traversing --name: rc=$RC out=$OUT"
[ ! -e "$SANDBOX_HOME/../../../../victim" ] \
  && pass "a traversing --name creates nothing outside the state directory" \
  || fail "traversing --name created something outside"

new_state
run_advance --name 'Demo Change' --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 2 ] && pass "a --name outside [a-z0-9-] exits 2" || fail "bad --name: rc=$RC"

# 20. An unknown argument exits 2 — the branch no case exercised before.
new_state
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done --oops
[ "$RC" -eq 2 ] && pass "an unknown argument exits 2" || fail "unknown argument: rc=$RC out=$OUT"
[ -n "$OUT" ] && pass "an unknown argument prints a message" || fail "unknown argument: empty output"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "an unknown argument writes nothing" || fail "unknown argument mutated the file"

# 21. A typo'd --target must not be written as a stage at exit 0.
new_state
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target do-dune --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 2 ] && pass "a --target that is not a stage exits 2" || fail "bogus --target: rc=$RC out=$OUT"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "a bogus --target writes nothing" || fail "bogus --target mutated the file"

new_state
run_advance --name demo --target do-done --accepted awaiting-do-review,nonsense --by /myflow-do-done
[ "$RC" -eq 2 ] && pass "a bogus --accepted element exits 2" || fail "bogus --accepted: rc=$RC"

# 22. --accepted matches a NON-FIRST element, and tolerates whitespace around
# the separator. Every case above happened to match element one.
new_state
jq '.stage = "do-review-started"' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target do-done \
  --accepted 'awaiting-do-review, do-review-started' --by /myflow-do-done
[ "$RC" -eq 0 ] && pass "--accepted matches a non-first element with whitespace" \
  || fail "non-first --accepted element: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "do-done" ] \
  && pass "non-first --accepted element wrote the stage" || fail "non-first element did not write"

if [ "$FAILURES" -ne 0 ]; then
  printf '\n%d assertion(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '\nAll state-advance assertions passed\n'
