#!/usr/bin/env bash
# check-contract-budget.sh — fail when a contract file outgrows its declared budget.
#
# Budgets are a RATCHET AGAINST REGROWTH, not a target. Each was set to the size
# the file actually had when the core/rationale split landed, plus 25%. Nothing
# here was ever a number the split had to hit: a target-first budget creates
# pressure to push normative text into an appendix to make a number, which is the
# one way that change could have done harm rather than merely disappoint.
#
# The table lives HERE rather than in the files being measured, so that raising a
# budget is a visible diff in the guard rather than an edit to the file that
# outgrew it. Bytes rather than lines, because a long-line file cheats a line
# count; bytes rather than tokens, because that would need a tokenizer this
# repository does not have and could not pin.
#
# EVERY .md in skills/myflow-contracts/ needs a row, appendices and SKILL.md
# alike. A file with no row is a violation rather than a pass — otherwise a
# contract added later escapes the ratchet silently, which is exactly how the
# file inventory in myflow-contract-distribution went stale through three
# changes before anyone noticed.
#
# Exit codes: 0 every file within budget; 1 a file over budget or with no row;
# 2 the guard cannot answer at all.
set -euo pipefail

# REPO_ROOT is resolved from this script's own location, exactly like
# check-references.sh and check-vocabulary.sh. That is what makes the guard
# argument-free and self-scoped from any cwd: running it from /tmp still scans
# the real repository instead of silently scanning nothing and reporting a false
# clean.
#
# CHECK_CONTRACT_BUDGET_ROOT is an explicit, opt-in override honored only when
# set. It exists solely so the companion harness can point the guard at a
# sandboxed fixture tree — never set it for a normal invocation.
if [ -n "${CHECK_CONTRACT_BUDGET_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_CONTRACT_BUDGET_ROOT"
elif [ "${CHECK_CONTRACT_BUDGET_ROOT+set}" = "set" ]; then
  printf 'CHECK_CONTRACT_BUDGET_ROOT is set but empty\n' >&2
  exit 2
else
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

[ -d "$REPO_ROOT" ] || { printf 'not a directory: %s\n' "$REPO_ROOT" >&2; exit 2; }

CONTRACT_DIR="$REPO_ROOT/skills/myflow-contracts"
[ -d "$CONTRACT_DIR" ] || {
  printf 'no contracts directory: %s\n' "$CONTRACT_DIR" >&2
  exit 2
}

# name<space>max-bytes, one row per .md in the contracts directory.
budgets() {
  cat <<'EOF'
SKILL.md 6073
build-green.md 4557
jira-followups.md 44633
jira-integration-rationale.md 3090
jira-integration.md 19488
pipeline-rationale.md 20572
pipeline.md 110316
plan-provenance.md 30582
project-configuration.md 56572
state-file.md 19938
state-self-heal.md 19045
workspace-isolation.md 38848
EOF
}

# budget_for <filename> — echo the declared budget, or nothing.
#
# Deliberately NOT `awk -v n="$name"`: a filename containing a newline makes
# awk's -v assignment fail, and under `set -e` that aborts the whole run with
# exit 2 — indistinguishable from "the guard cannot answer at all", and it
# abandons every other file in the directory rather than reporting one bad name.
# Comparing in bash keeps a pathological filename a per-file violation, which is
# what it is.
budget_for() {
  local want="$1" bname bmax
  while read -r bname bmax; do
    if [ "$bname" = "$want" ]; then
      printf '%s' "$bmax"
      return 0
    fi
  done <<<"$(budgets)"
  return 0
}

violations=0
checked=0

for path in "$CONTRACT_DIR"/*.md; do
  # An unmatched glob expands to the literal pattern, so a directory holding no
  # .md at all would otherwise be reported clean. `-e` alone follows symlinks and
  # is false for a broken one, so a single dangling `*.md` symlink would abort
  # here claiming the directory holds no .md files while real ones sit beside it
  # unexamined — a false report from a guard, which is the thing this repository
  # least tolerates. `-L` catches the symlink itself, whatever it points at.
  { [ -e "$path" ] || [ -L "$path" ]; } \
    || { printf 'no .md file matched %s/*.md\n' "$CONTRACT_DIR" >&2; exit 2; }
  checked=$((checked + 1))
  name="$(basename "$path")"

  # A contract is a REGULAR FILE tracked in git. A symlink here is refused rather
  # than followed, and it is refused before anything reads it.
  #
  # Following one would size whatever it points at, anywhere this process can
  # read: a `.md`-named symlink added in a pull request turns this guard into an
  # existence-and-size oracle for arbitrary paths on whoever's machine runs it,
  # and if its name matches a budget row the target's exact byte count is printed
  # in the violation line. The content never leaks, but the size does, and none of
  # it is this guard's business. Refusing is also the honest answer: a symlink has
  # no size of its own to ratchet.
  if [ -L "$path" ]; then
    printf 'skills/myflow-contracts/%s: is a symlink — a contract must be a regular file\n' "$name"
    violations=$((violations + 1))
    continue
  fi

  max="$(budget_for "$name")"

  # `wc -c` is guarded rather than assigned bare. Under `set -e` a plain
  # assignment from a failing command aborts with THAT command's status — 1 for
  # an unreadable file — which is the code this contract reserves for "over
  # budget or undeclared", and would send a caller off to edit budgets() when
  # the guard itself could not read the file.
  # Readability is tested BEFORE the redirect rather than relying on `2>/dev/null`
  # to mop up after it. When `<` cannot open the file, bash reports that itself,
  # to the shell's stderr, before any redirection on the command is established —
  # so the suppression never applies to the one failure this guard is handling,
  # and the operator sees bash's raw line followed by ours. Testing first means
  # one message, in this guard's own format.
  if [ ! -r "$path" ]; then
    printf 'cannot read %s\n' "$path" >&2
    exit 2
  fi
  if ! actual="$(wc -c < "$path" 2>/dev/null)"; then
    printf 'cannot read %s\n' "$path" >&2
    exit 2
  fi
  actual="$(printf '%s' "$actual" | tr -d '[:space:]')"

  if [ -z "$max" ]; then
    printf 'skills/myflow-contracts/%s: no budget declared — add a row to budgets() in %s\n' \
      "$name" "$(basename "${BASH_SOURCE[0]}")"
    violations=$((violations + 1))
  elif [ "$actual" -gt "$max" ]; then
    printf 'skills/myflow-contracts/%s: %s bytes exceeds budget %s — trim it back under budget, or raise its row in budgets() in %s if the growth is deliberate\n' \
      "$name" "$actual" "$max" "$(basename "${BASH_SOURCE[0]}")"
    violations=$((violations + 1))
  fi
done

if [ "$violations" -gt 0 ]; then
  printf 'BUDGET-FAIL: %d file(s) over budget or undeclared\n' "$violations"
  exit 1
fi

# The count is part of the verdict deliberately: a bare "BUDGET-OK" cannot be
# told apart from a run that scanned nothing, which is the failure the
# self-scoped REPO_ROOT above exists to prevent. The number is the evidence that
# the guard looked.
printf 'BUDGET-OK: %d contract file(s) within budget\n' "$checked"
