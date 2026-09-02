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
# stale relative to what the skill actually says to run. Only `flow` (the
# block's one external command) is stubbed, via a throwaway PATH entry, so
# the block runs unmodified against representative canned settings JSON.
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

# run_case <settings-json> <expected self_review_model> <expected planning_model> <case name>
FAILURES=0
run_case() {
  local json="$1" expect_srm="$2" expect_pm="$3" name="$4"

  cat >"$STUB_DIR/flow" <<EOF
#!/usr/bin/env bash
printf '%s' '$json'
EOF
  chmod +x "$STUB_DIR/flow"

  local out
  out="$(PATH="$STUB_DIR:$PATH" bash -c "$BLOCK"$'\n''printf "%s\t%s\n" "$SELF_REVIEW_MODEL" "$PLANNING_MODEL"' 2>&1)" \
    || { echo "check-model-resolution-shell: case '$name' — block exited non-zero: $out" >&2; FAILURES=$((FAILURES + 1)); return; }

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

if [[ "$FAILURES" -gt 0 ]]; then
  echo "check-model-resolution-shell: $FAILURES failure(s)" >&2
  exit 1
fi

echo "MODEL-RESOLUTION-SHELL-OK: 3 case(s) checked"
exit 0
