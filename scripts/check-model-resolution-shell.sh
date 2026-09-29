#!/usr/bin/env bash
# check-model-resolution-shell.sh — proves the documented bash block that
# resolves /flow's run-level dispatch model actually resolves it correctly,
# rather than trusting the prose by eye: skills/flow/SKILL.md's "Model
# resolution" block (DEFAULT_MODEL and MODEL_SOURCE). The three project
# toggles that block once resolved were removed: execution mode, implementer
# effort and the review panel are always the plan's decision.
# SELF_REVIEW_MODEL is resolved nowhere any more (KAN-854): the archive-phase
# self-review pass runs inline on the session's own model, so a resolution
# only cost two subprocesses and a possible exit-2 stop for a value nothing
# read. This guard's drift checks refuse one re-added to either file.
#
# WHY THIS EXISTS. That block is documentation — prose describing what a
# `/flow` run executes, not itself a script this repository runs in CI — so
# no guard previously caught a change that silently broke its logic. The
# concrete failure this closes: flipping `[ -n "$DEFAULT_MODEL" ]` to
# `[ -z "$DEFAULT_MODEL" ]` is a one-character change with no compiler and
# no type system behind it, and it would forcibly overwrite a real,
# non-empty resolved value with the `opus` fallback instead of only filling
# in the empty case. Extracted and run for real, that mutation fails this
# script. PLANNING_MODEL was covered here too until kan-488 removed it end
# to end — planning runs inline on the session's own model now.
#
# HOW IT AVOIDS DUPLICATING THE BLOCK. The exact fenced ```bash block under
# skills/flow/SKILL.md's "## Model resolution" heading is extracted verbatim
# and executed — never retyped here — so this guard cannot go stale relative
# to what the skill actually says to run. archive.md is only drift-checked. `flow` is
# stubbed, via a throwaway PATH entry, answering both `settings get`
# (canned settings JSON) and `settings models` (the fixed vocabulary);
# `project-get.sh` is the real script, on `PATH` from this repository's
# own scripts/, run for real against a per-case MAIN_CHECKOUT fixture.
#
# CHECK_MODEL_RESOLUTION_SKILL_MD and CHECK_MODEL_RESOLUTION_ARCHIVE_MD, when
# set, name the files the block is
# extracted from and the drift check reads, instead of the real skills — test-check-model-resolution-shell.sh's
# sandbox override, so its mutation cases never write the real tree
# (KAN-376: a concurrent run-guard-tests.sh run fingerprints that tree by
# mtime, and a restore landing inside the fingerprint window failed
# test-setup.sh's containment case). The default is unchanged: production
# reads the real skill.
#
# Usage: check-model-resolution-shell.sh
# Exit 0 the extracted block resolves both variables correctly for every
# case below, 1 a case resolved wrong, 2 cannot answer at all (either file
# missing/unreadable, the block not found, a drift check refused, or a
# temp-file failure).
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

[[ -n "$SKILL_BLOCK" ]] || die "no \`\`\`bash block found under '## Model resolution' in $SKILL_MD"
echo "$SKILL_BLOCK" | grep -q 'SELF_REVIEW_MODEL' && die "SKILL.md's block mentions SELF_REVIEW_MODEL — no run resolves it (KAN-854), this guard's own drift check"
grep -q 'SELF_REVIEW_MODEL=' "$ARCHIVE_MD" && die "archive.md assigns SELF_REVIEW_MODEL — no run resolves it (KAN-854), this guard's own drift check"
echo "$SKILL_BLOCK" | grep -q 'DEFAULT_MODEL' || die "extracted block does not mention DEFAULT_MODEL — heading or fence shape changed"
echo "$SKILL_BLOCK" | grep -q '_TOGGLE' && die "SKILL.md's block still resolves a toggle — the decision is always the planner's, this guard's own drift check"
echo "$SKILL_BLOCK" | grep -q 'PLANNING_MODEL' && die "extracted block still mentions PLANNING_MODEL — it is no longer resolved anywhere, this guard's own drift check"

BLOCK="$SKILL_BLOCK"

STUB_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-model-resolution-shell.XXXXXX")" \
  || die "cannot create a temp dir for the flow stub"
trap 'rm -rf "$STUB_DIR"' EXIT

# run_case <settings-json> <expected DEFAULT_MODEL> <expected MODEL_SOURCE> <case name> [project.md body]
# The fifth argument, when non-empty, is written verbatim to a fresh
# MAIN_CHECKOUT fixture's .flow/project.md; when omitted or empty, the
# fixture carries no .flow/project.md at all (project-get.sh's "no file"
# exit 1). An empty settings JSON makes the stub's `settings get` exit 3 —
# the store unreachable. Every expectation is passed explicitly, so no case
# inherits another's.
FAILURES=0
CASES=0
run_case() {
  local json="$1" expect_default="$2" expect_source="$3" name="$4" project_body="${5:-}"
  CASES=$((CASES + 1))

  cat >"$STUB_DIR/flow" <<EOF
#!/usr/bin/env bash
case "\$1 \$2" in
  "settings get") if [ -z '$json' ]; then exit 3; else printf '%s' '$json'; fi ;;
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

  local out
  out="$(MAIN_CHECKOUT="$checkout" PATH="$STUB_DIR:$SCRIPT_DIR:$PATH" bash -c "$BLOCK"$'\n''printf "%s\t%s\n" "$DEFAULT_MODEL" "$MODEL_SOURCE"' 2>/dev/null)"
  local rc=$?
  rm -rf "$checkout"
  if [[ $rc -ne 0 ]]; then
    echo "check-model-resolution-shell: case '$name' — block exited non-zero: $out" >&2
    FAILURES=$((FAILURES + 1))
    return
  fi

  local got_default got_source
  IFS=$'\t' read -r got_default got_source <<<"$out"

  if [[ "$got_default" != "$expect_default" ]]; then
    echo "check-model-resolution-shell: case '$name' — DEFAULT_MODEL resolved to '$got_default', want '$expect_default'" >&2
    FAILURES=$((FAILURES + 1))
  fi
  if [[ "$got_source" != "$expect_source" ]]; then
    echo "check-model-resolution-shell: case '$name' — MODEL_SOURCE resolved to '$got_source', want '$expect_source'" >&2
    FAILURES=$((FAILURES + 1))
  fi
}

# Case 1: the store answers, no project file — the store's value.
run_case '{"defaultModel":"sonnet","reviewers":[]}' \
  "sonnet" "store" "store value, no project file"

# Case 2: a `## self review model` key never reaches DEFAULT_MODEL — the
# dispatch key is `## model` alone.
run_case '{"defaultModel":"sonnet","reviewers":[]}' \
  "sonnet" "store" "a self review model key never reaches DEFAULT_MODEL" \
  $'## self review model\n\n`haiku`\n'

# Case 3: the project's `## model` key is present and valid — it beats the
# store's value, and MODEL_SOURCE says so.
run_case '{"defaultModel":"sonnet","reviewers":[]}' \
  "opus" "project" "project-model-wins" \
  $'## model\n\n`opus`\n'

# Case 4: the key is present but invalid — reported and dropped, so the
# store's value stands.
run_case '{"defaultModel":"sonnet","reviewers":[]}' \
  "sonnet" "store" "project-model-invalid-drops" \
  $'## model\n\n`gpt-9`\n'

# Case 5: the store is unreachable — the project key still resolves the
# dispatch model, the outage property the in-repo key exists to provide.
run_case '' \
  "opus" "project" "store-down-project-wins" \
  $'## model\n\n`opus`\n'

# Case 6: the store is unreachable and no project key — the literal
# fallback, named as one.
run_case '' \
  "opus" "fallback" "store-down-no-project-falls-back"

# Case 7: the store answers JSON null for defaultModel — the literal
# fallback fires rather than the string "null" becoming the model.
run_case '{"defaultModel":null,"reviewers":[]}' \
  "opus" "fallback" "store-null-falls-back"

if [[ "$FAILURES" -gt 0 ]]; then
  echo "check-model-resolution-shell: $FAILURES failure(s)" >&2
  exit 1
fi

echo "MODEL-RESOLUTION-SHELL-OK: $CASES case(s) checked"
exit 0
