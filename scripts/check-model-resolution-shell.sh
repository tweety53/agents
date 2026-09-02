#!/usr/bin/env bash
# check-model-resolution-shell.sh — proves skills/flow/SKILL.md's "Model
# resolution" bash block actually resolves SELF_REVIEW_MODEL and
# PLANNING_MODEL correctly, rather than trusting the prose by eye.
#
# WHY THIS EXISTS. That block is documentation — prose describing what a
# `/flow` run executes, not itself a script this repository runs in CI — so
# no guard previously caught a change that silently broke its logic. The
# concrete failure this closes: flipping either `[ -z "$X" ]` test to
# `[ -n "$X" ]` is a one-character change with no compiler and no type
# system behind it, and it would forcibly overwrite a real, non-empty
# resolved value with the `fable` fallback instead of only filling in the
# empty case. Extracted and run for real, that mutation fails this script.
#
# HOW IT AVOIDS DUPLICATING THE BLOCK. The exact fenced ```bash block under
# skills/flow/SKILL.md's "## Model resolution" heading is extracted
# verbatim and executed — never retyped here — so this guard cannot go
# stale relative to what the skill actually says to run. `flow` is
# stubbed, via a throwaway PATH entry, answering both `settings get`
# (canned settings JSON) and `settings models` (the fixed vocabulary);
# `project-get.sh` is the real script, on `PATH` from this repository's
# own scripts/, run for real against a per-case MAIN_CHECKOUT fixture.
#
# Usage: check-model-resolution-shell.sh
# Exit 0 the extracted block resolves both variables correctly for every
# case below, 1 a case resolved wrong, 2 cannot answer at all (SKILL.md
# missing/unreadable, the block not found in it, or a temp-file failure).
set -uo pipefail

die() {
  echo "check-model-resolution-shell: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILL_MD="$REPO_ROOT/skills/flow/SKILL.md"

[[ -r "$SKILL_MD" ]] || die "cannot read $SKILL_MD"

# Extract the first ```bash ... ``` fence after the "## Model resolution"
# heading, verbatim, body lines only (fences excluded).
BLOCK="$(awk '
  $0 == "## Model resolution" { seen_heading = 1; next }
  seen_heading && /^```bash$/ { in_block = 1; next }
  in_block && /^```$/ { exit }
  in_block { print }
' "$SKILL_MD")"

[[ -n "$BLOCK" ]] || die "no \`\`\`bash block found under '## Model resolution' in $SKILL_MD"
echo "$BLOCK" | grep -q 'SELF_REVIEW_MODEL' || die "extracted block does not mention SELF_REVIEW_MODEL — heading or fence shape changed"
echo "$BLOCK" | grep -q 'PLANNING_MODEL' || die "extracted block does not mention PLANNING_MODEL — heading or fence shape changed"

STUB_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-model-resolution-shell.XXXXXX")" \
  || die "cannot create a temp dir for the flow stub"
trap 'rm -rf "$STUB_DIR"' EXIT

# run_case <settings-json> <expected self_review_model> <expected planning_model> <case name> [project.md body]
# The fifth argument, when non-empty, is written verbatim to a fresh
# MAIN_CHECKOUT fixture's .flow/project.md; when omitted or empty, the
# fixture carries no .flow/project.md at all (project-get.sh's "no file"
# exit 1), reproducing the pre-task-6 behaviour cases 1-3 still expect.
FAILURES=0
run_case() {
  local json="$1" expect_srm="$2" expect_pm="$3" name="$4" project_body="${5:-}"

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

  local out stderr_file
  stderr_file="$(mktemp "${TMPDIR:-/tmp}/check-model-resolution-shell-stderr.XXXXXX")" \
    || die "cannot create a temp stderr file"
  out="$(MAIN_CHECKOUT="$checkout" PATH="$STUB_DIR:$SCRIPT_DIR:$PATH" bash -c "$BLOCK"$'\n''printf "%s\t%s\n" "$SELF_REVIEW_MODEL" "$PLANNING_MODEL"' 2>"$stderr_file")"
  local rc=$?
  rm -rf "$checkout" "$stderr_file"
  if [[ $rc -ne 0 ]]; then
    echo "check-model-resolution-shell: case '$name' — block exited non-zero: $out" >&2
    FAILURES=$((FAILURES + 1))
    return
  fi

  local got_srm="${out%%$'\t'*}"
  local got_pm="${out#*$'\t'}"

  if [[ "$got_srm" != "$expect_srm" ]]; then
    echo "check-model-resolution-shell: case '$name' — SELF_REVIEW_MODEL resolved to '$got_srm', want '$expect_srm'" >&2
    FAILURES=$((FAILURES + 1))
  fi
  if [[ "$got_pm" != "$expect_pm" ]]; then
    echo "check-model-resolution-shell: case '$name' — PLANNING_MODEL resolved to '$got_pm', want '$expect_pm'" >&2
    FAILURES=$((FAILURES + 1))
  fi
}

# Case 1: both fields empty — the fallback path (`-z` true).
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"","planningModel":""}' \
  "fable" "fable" "both empty"

# Case 2: both fields set — the non-empty path (`-z` false, value preserved).
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus","planningModel":"haiku"}' \
  "opus" "haiku" "both set"

# Case 3: both fields absent from the JSON entirely — jq's `// empty` path.
run_case '{"defaultModel":"sonnet","reviewers":[]}' \
  "fable" "fable" "both absent"

# Case 4: the project's model keys are present and valid — they win over a
# set, different store value, for both keys.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus","planningModel":"haiku"}' \
  "haiku" "opus" "project key valid wins over store" \
  $'## self review model\n\n`haiku`\n\n## planning model\n\n`opus`\n'

# Case 5: the project's model keys are present but invalid — reported and
# dropped, so the store's value survives untouched.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus","planningModel":"haiku"}' \
  "opus" "haiku" "project key invalid is dropped, store survives" \
  $'## self review model\n\n`not-a-model`\n\n## planning model\n\n`also-not-a-model`\n'

# Case 6: the project declares neither key — falls through to the store's
# own set value, exactly as if project-get.sh were never consulted.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"opus","planningModel":"haiku"}' \
  "opus" "haiku" "project key absent falls through to store"

# Case 7: the project's model keys are present and valid, and the store's
# own fields are empty — the project value wins over the `fable` fallback.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":"","planningModel":""}' \
  "sonnet" "haiku" "project key valid wins over fable fallback" \
  $'## self review model\n\n`sonnet`\n\n## planning model\n\n`haiku`\n'

if [[ "$FAILURES" -gt 0 ]]; then
  echo "check-model-resolution-shell: $FAILURES failure(s)" >&2
  exit 1
fi

echo "MODEL-RESOLUTION-SHELL-OK: 7 case(s) checked"
exit 0
