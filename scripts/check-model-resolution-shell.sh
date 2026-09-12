#!/usr/bin/env bash
# check-model-resolution-shell.sh — proves the two documented bash blocks
# that resolve /flow's run-level variables actually resolve them correctly,
# rather than trusting the prose by eye: skills/flow/SKILL.md's "Model
# resolution" block for the three toggles (EXECUTION_MODE_TOGGLE,
# IMPLEMENTER_MODEL_TOGGLE, REVIEW_PANEL_TOGGLE), and skills/flow/archive.md
# step 9's block for SELF_REVIEW_MODEL. The two are extracted and run
# together, in that order, so the cases below assert all four variables at
# once exactly as they did when one block resolved all four. SELF_REVIEW_MODEL
# moved to archive.md because only the archive-phase self-review pass reads it
# and it governs no dispatch — resolving it on every run cost two subprocess
# calls a run that stops earlier never needs.
#
# WHY THIS EXISTS. That block is documentation — prose describing what a
# `/flow` run executes, not itself a script this repository runs in CI — so
# no guard previously caught a change that silently broke its logic. The
# concrete failure this closes: flipping `[ -z "$SELF_REVIEW_MODEL" ]` to
# `[ -n "$SELF_REVIEW_MODEL" ]` is a one-character change with no compiler
# and no type system behind it, and it would forcibly overwrite a real,
# non-empty resolved value with the `fable` fallback instead of only
# filling in the empty case. Extracted and run for real, that mutation
# fails this script. PLANNING_MODEL was covered here too until kan-488
# removed it end to end — planning runs inline on the session's own model
# now, with no PLANNING_MODEL variable left in the block to resolve.
#
# HOW IT AVOIDS DUPLICATING THE BLOCKS. The exact fenced ```bash block under
# skills/flow/SKILL.md's "## Model resolution" heading, and the one under
# archive.md's "Resolve `SELF_REVIEW_MODEL` here" marker line, are extracted
# verbatim and executed — never retyped here — so this guard cannot go
# stale relative to what the skills actually say to run. `flow` is
# stubbed, via a throwaway PATH entry, answering both `settings get`
# (canned settings JSON) and `settings models` (the fixed vocabulary);
# `project-get.sh` is the real script, on `PATH` from this repository's
# own scripts/, run for real against a per-case MAIN_CHECKOUT fixture.
#
# CHECK_MODEL_RESOLUTION_SKILL_MD and CHECK_MODEL_RESOLUTION_ARCHIVE_MD, when
# set, name the files the blocks are
# extracted from instead of the real skills — test-check-model-resolution-shell.sh's
# sandbox override, so its mutation cases never write the real tree
# (KAN-376: a concurrent run-guard-tests.sh run fingerprints that tree by
# mtime, and a restore landing inside the fingerprint window failed
# test-setup.sh's containment case). The default is unchanged: production
# reads the real skill.
#
# Usage: check-model-resolution-shell.sh
# Exit 0 the extracted blocks resolve every variable correctly for every
# case below, 1 a case resolved wrong, 2 cannot answer at all (either file
# missing/unreadable, a block not found in it, or a temp-file failure).
set -uo pipefail

die() {
  echo "check-model-resolution-shell: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILL_MD="$REPO_ROOT/skills/flow/SKILL.md"
ARCHIVE_MD="$REPO_ROOT/skills/flow/archive.md"
# Honoured only when set, and refused when set but empty — the
# RUN_GUARD_TESTS_ROOT idiom: a silent fallback to the real skill would run
# the cases against a file the caller never chose.
if [ "${CHECK_MODEL_RESOLUTION_SKILL_MD+set}" = set ]; then
  [[ -n "$CHECK_MODEL_RESOLUTION_SKILL_MD" ]] || die "CHECK_MODEL_RESOLUTION_SKILL_MD is set but empty"
  SKILL_MD="$CHECK_MODEL_RESOLUTION_SKILL_MD"
fi
if [ "${CHECK_MODEL_RESOLUTION_ARCHIVE_MD+set}" = set ]; then
  [[ -n "$CHECK_MODEL_RESOLUTION_ARCHIVE_MD" ]] || die "CHECK_MODEL_RESOLUTION_ARCHIVE_MD is set but empty"
  ARCHIVE_MD="$CHECK_MODEL_RESOLUTION_ARCHIVE_MD"
fi

[[ -r "$SKILL_MD" ]] || die "cannot read $SKILL_MD"
[[ -r "$ARCHIVE_MD" ]] || die "cannot read $ARCHIVE_MD"

# Extract the first ```bash ... ``` fence after the "## Model resolution"
# heading, verbatim, body lines only (fences excluded).
SKILL_BLOCK="$(awk '
  $0 == "## Model resolution" { seen_heading = 1; next }
  seen_heading && /^```bash$/ { in_block = 1; next }
  in_block && /^```$/ { exit }
  in_block { print }
' "$SKILL_MD")"

# And the first ```bash fence after archive.md step 9's marker line. That
# fence is indented three spaces, as a numbered list item's block must be;
# the common indent is stripped so the body executes as written.
ARCHIVE_BLOCK="$(awk '
  /^ *\*\*Resolve `SELF_REVIEW_MODEL` here, where it is consumed\*\*/ { seen_marker = 1; next }
  seen_marker && /^ *```bash$/ { in_block = 1; next }
  in_block && /^ *```$/ { exit }
  in_block { sub(/^   /, ""); print }
' "$ARCHIVE_MD")"

[[ -n "$SKILL_BLOCK" ]] || die "no \`\`\`bash block found under '## Model resolution' in $SKILL_MD"
[[ -n "$ARCHIVE_BLOCK" ]] || die "no \`\`\`bash block found under the 'Resolve SELF_REVIEW_MODEL here' marker in $ARCHIVE_MD"
echo "$ARCHIVE_BLOCK" | grep -q 'SELF_REVIEW_MODEL' || die "extracted archive block does not mention SELF_REVIEW_MODEL — marker or fence shape changed"
echo "$SKILL_BLOCK" | grep -q 'SELF_REVIEW_MODEL' && die "SKILL.md's block still mentions SELF_REVIEW_MODEL — it resolves in archive.md now, this guard's own drift check"
echo "$SKILL_BLOCK" | grep -q 'EXECUTION_MODE_TOGGLE' || die "extracted block does not mention EXECUTION_MODE_TOGGLE — heading or fence shape changed"
echo "$SKILL_BLOCK" | grep -q 'IMPLEMENTER_MODEL_TOGGLE' || die "extracted block does not mention IMPLEMENTER_MODEL_TOGGLE — heading or fence shape changed"
echo "$SKILL_BLOCK" | grep -q 'REVIEW_PANEL_TOGGLE' || die "extracted block does not mention REVIEW_PANEL_TOGGLE — heading or fence shape changed"
echo "$SKILL_BLOCK" | grep -q 'PLANNING_MODEL' && die "extracted block still mentions PLANNING_MODEL — kan-488 removed it end to end, this guard's own drift check"

# Run order matches the run's own: the toggles first, then archive step 9.
BLOCK="$SKILL_BLOCK
$ARCHIVE_BLOCK"

STUB_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-model-resolution-shell.XXXXXX")" \
  || die "cannot create a temp dir for the flow stub"
trap 'rm -rf "$STUB_DIR"' EXIT

# run_case <settings-json> <expected self_review_model> <case name> \
#          [project.md body] [expected execution mode toggle] [expected implementer model toggle] \
#          [expected review panel toggle] [expected stderr substring]
# The fourth argument, when non-empty, is written verbatim to a fresh
# MAIN_CHECKOUT fixture's .flow/project.md; when omitted or empty, the
# fixture carries no .flow/project.md at all (project-get.sh's "no file"
# exit 1), reproducing the pre-task-6 behaviour cases 1-3 still expect. The
# three toggle arguments default to "default" — the no-key/absent-key
# resolution every pre-task-6 case still exercises. The eighth argument, when
# non-empty, must appear in the block's stderr.
FAILURES=0
run_case() {
  local json="$1" expect_srm="$2" name="$3" project_body="${4:-}" \
        expect_exec="${5:-default}" expect_impl="${6:-default}" expect_panel="${7:-default}" \
        expect_stderr="${8:-}"

  cat >"$STUB_DIR/flow" <<EOF
#!/usr/bin/env bash
case "\$1 \$2" in
  "settings get") printf '%s' '$json' ;;
  "settings models") printf 'sonnet\nopus\nhaiku\nfable\n' ;;
  *) echo "flow stub: unexpected arguments: \$*" >&2; exit 2 ;;
esac
EOF
  chmod +x "$STUB_DIR/flow"

  local checkout
  checkout="$(mktemp -d "${TMPDIR:-/tmp}/check-model-resolution-shell-checkout.XXXXXX")" \
    || die "cannot create a temp MAIN_CHECKOUT dir"
  if [[ -n "$project_body" ]]; then
    mkdir -p "$checkout/.flow"
    printf '%s\n' "$project_body" >"$checkout/.flow/project.md"
  fi

  local out stderr_file stderr_out
  stderr_file="$(mktemp "${TMPDIR:-/tmp}/check-model-resolution-shell-stderr.XXXXXX")" \
    || die "cannot create a temp stderr file"
  out="$(MAIN_CHECKOUT="$checkout" PATH="$STUB_DIR:$SCRIPT_DIR:$PATH" bash -c "$BLOCK"$'\n''printf "%s\t%s\t%s\t%s\n" "$SELF_REVIEW_MODEL" "$EXECUTION_MODE_TOGGLE" "$IMPLEMENTER_MODEL_TOGGLE" "$REVIEW_PANEL_TOGGLE"' 2>"$stderr_file")"
  local rc=$?
  stderr_out="$(cat "$stderr_file")"
  rm -rf "$checkout" "$stderr_file"
  if [[ $rc -ne 0 ]]; then
    echo "check-model-resolution-shell: case '$name' — block exited non-zero: $out" >&2
    FAILURES=$((FAILURES + 1))
    return
  fi

  local got_srm got_exec got_impl got_panel
  IFS=$'\t' read -r got_srm got_exec got_impl got_panel <<<"$out"

  if [[ "$got_srm" != "$expect_srm" ]]; then
    echo "check-model-resolution-shell: case '$name' — SELF_REVIEW_MODEL resolved to '$got_srm', want '$expect_srm'" >&2
    FAILURES=$((FAILURES + 1))
  fi
  if [[ "$got_exec" != "$expect_exec" ]]; then
    echo "check-model-resolution-shell: case '$name' — EXECUTION_MODE_TOGGLE resolved to '$got_exec', want '$expect_exec'" >&2
    FAILURES=$((FAILURES + 1))
  fi
  if [[ "$got_impl" != "$expect_impl" ]]; then
    echo "check-model-resolution-shell: case '$name' — IMPLEMENTER_MODEL_TOGGLE resolved to '$got_impl', want '$expect_impl'" >&2
    FAILURES=$((FAILURES + 1))
  fi
  if [[ "$got_panel" != "$expect_panel" ]]; then
    echo "check-model-resolution-shell: case '$name' — REVIEW_PANEL_TOGGLE resolved to '$got_panel', want '$expect_panel'" >&2
    FAILURES=$((FAILURES + 1))
  fi
  if [[ -n "$expect_stderr" ]] && [[ "$stderr_out" != *"$expect_stderr"* ]]; then
    echo "check-model-resolution-shell: case '$name' — stderr did not contain '$expect_stderr': got '$stderr_out'" >&2
    FAILURES=$((FAILURES + 1))
  fi
}

# Case 1: the field empty — the fallback path (`-z` true).
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":""}' \
  "fable" "empty"

# Case 2: the field set — the non-empty path (`-z` false, value preserved).
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus"}' \
  "opus" "set"

# Case 3: the field absent from the JSON entirely — jq's `// empty` path.
run_case '{"defaultModel":"sonnet","reviewers":[]}' \
  "fable" "absent"

# Case 4: the project's self-review-model key is present and valid — it
# wins over a set, different store value.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus"}' \
  "haiku" "project key valid wins over store" \
  $'## self review model\n\n`haiku`\n'

# Case 5: the project's self-review-model key is present but invalid —
# reported and dropped, so the store's value survives untouched.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus"}' \
  "opus" "project key invalid is dropped, store survives" \
  $'## self review model\n\n`not-a-model`\n'

# Case 6: the project declares no key — falls through to the store's own
# set value, exactly as if project-get.sh were never consulted.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus"}' \
  "opus" "project key absent falls through to store"

# Case 7: the project's self-review-model key is present and valid, and
# the store's own field is empty — the project value wins over the
# `fable` fallback.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":""}' \
  "sonnet" "project key valid wins over fable fallback" \
  $'## self review model\n\n`sonnet`\n'

# Case 8: all three toggle keys declared `dynamic` — each resolves to
# `dynamic`, independently of the model key.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus"}' \
  "opus" "all three toggles dynamic" \
  $'## execution mode\n\n`dynamic`\n\n## implementer model\n\n`dynamic`\n\n## review panel\n\n`dynamic`\n' \
  "dynamic" "dynamic" "dynamic"

# Case 9: `## review panel` holds a literal that is neither `default` nor
# `dynamic` — reported and dropped, resolving as `default`, with the
# warning on stderr.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus"}' \
  "opus" "review panel toggle invalid falls back to default" \
  $'## review panel\n\n`sometimes`\n' \
  "default" "default" "default" \
  "'## review panel' body 'sometimes' is not 'default' or 'dynamic' — dropped"

if [[ "$FAILURES" -gt 0 ]]; then
  echo "check-model-resolution-shell: $FAILURES failure(s)" >&2
  exit 1
fi

echo "MODEL-RESOLUTION-SHELL-OK: 9 case(s) checked"
exit 0
