#!/usr/bin/env bash
# check-contract-budget.sh — fail when an owned Markdown file outgrows its
# declared budget.
#
# Budgets are a RATCHET AGAINST REGROWTH, not a target. Each is the size its
# file actually had when the change that added its row landed, plus 25%.
# Nothing here was ever a number a change had to hit: a target-first budget
# creates pressure to push normative text into an appendix to make a number,
# which is the one way a trim could do harm rather than merely disappoint.
#
# The table lives HERE rather than in the files being measured, so that raising
# a budget is a visible diff in the guard rather than an edit to the file that
# outgrew it. Bytes rather than lines, because a long-line file cheats a line
# count; bytes rather than tokens, because that would need a tokenizer this
# repository does not have and could not pin.
#
# COVERAGE IS EVERY OWNED MARKDOWN FILE, not only the files a `/myflow-*` run
# loads. The guard used to stop at skills/myflow-contracts/*.md plus each
# skills/*/SKILL.md and its rationale sibling — 13 rows — on the reasoning that
# a byte cut is only paid for where it is loaded per run. Regrowth does not
# respect that boundary: when this guard was widened, 79 owned files sat outside
# it, openspec/specs/ alone accounting for 462 KB of text nothing held back.
#
# The corpus is resolved by scripts/lib/owned-corpus.sh, which
# check-normative-inventory.sh calls too. One implementation, because the
# inventory is the only thing standing between a corpus-wide prose trim and a
# silently deleted requirement: if the two guards disagreed about ownership, a
# file could be ratcheted here and invisible there, or the reverse. That library
# is canonical for the scope roots and the structural exclusions; this guard
# states neither a second time.
#
# THE ONE THING THIS GUARD ADDS TO THE LIBRARY'S ANSWER IS A SYMLINK REFUSAL.
# owned_corpus_files enumerates with `find -type f`, so a `.md` symlink is not
# owned content and never reaches the corpus — the right answer for ownership,
# since a symlink has no bytes of its own to ratchet. But silence would be the
# wrong answer here: replacing a ratcheted file with a symlink would drop it out
# of the corpus AND out of the ratchet with every check still green, which is
# the "escapes silently" failure this guard exists to prevent, arrived at from
# the other direction. So a `.md` or `.mdc` symlink standing where a covered
# file could stand is named and counted as a violation, before anything reads
# it. The sweep below reuses the library's own scope roots and its exclusion
# predicate, so it can never disagree with the corpus about WHERE a covered file
# may live — only about what to do with a symlink found there.
#
# EVERY COVERED FILE NEEDS A ROW. A file with no row is a violation rather than
# a pass — otherwise a file added later escapes the ratchet silently, which is
# exactly how the file inventory in myflow-contract-distribution went stale
# through three changes before anyone noticed.
#
# THE FROZEN openspec/specs/ TREE CARRIES NO ROWS ANY MORE, though it once did:
# it was the largest single piece of the 79-file, 462-KB widening described
# above, 33 rows' worth. Those rows are gone now because the tree they measured
# is no longer part of the corpus at all: scripts/lib/owned-corpus.sh's scope
# roots name spectre/specs/, not openspec/specs/, so owned_corpus_files no
# longer enumerates a single file under the old tree, and "every covered file
# needs a row" above already excuses it, the same way it excuses everything
# else outside the library's scope roots. A ratchet exists to catch REGROWTH,
# and a frozen tree cannot regrow — nothing under openspec/ is edited again,
# by this guard's own corpus definition. Do not re-add its rows: that tree is
# history now, not something this guard measures, no matter how large or
# small it happens to sit.
#
# THE TABLE IS HAND-MAINTAINED AND HAS NO REGENERATE MODE. A guard that can
# rewrite its own bound from current sizes makes regeneration the path of least
# resistance, and a ratchet whose bound is recomputed on demand does not
# ratchet. The table is regenerated once, by hand, as the last action of a
# change that alters any covered file — a table computed while later edits are
# still landing goes stale silently, and it did so three times in one change
# before that rule was written.
#
# ROWS ARE PER FILE, NEVER PER DIRECTORY. A directory total lets one file
# balloon while another shrinks with the sum unchanged, and the guard says
# nothing.
#
# The table is keyed on the path RELATIVE TO THE REPOSITORY ROOT, never on the
# bare basename. Every skill directory has a file literally named SKILL.md —
# skills/myflow-do/SKILL.md and skills/myflow-start/SKILL.md would collide on
# "SKILL.md" and silently share one row, so whichever budget that row held
# would govern both files rather than either one's own size.
#
# Exit codes: 0 every file within budget; 1 a file over budget, with no row, or
# standing where a covered file should be as a symlink; 2 the guard cannot
# answer at all.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# REPO_ROOT is resolved from this script's own location, exactly like
# check-references.sh and check-vocabulary.sh. That is what makes the guard
# argument-free and self-scoped from any cwd: running it from /tmp still scans
# the real repository instead of silently scanning nothing and reporting a false
# clean.
#
# CHECK_CONTRACT_BUDGET_ROOT is an explicit, opt-in override honored only when
# set. It exists solely so the companion harness can point the guard at a
# sandboxed fixture tree — never set it for a normal invocation. Set-but-empty
# is refused rather than treated as unset: falling back there would make the
# harness scan the real repository while believing it scanned a fixture.
if [ -n "${CHECK_CONTRACT_BUDGET_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_CONTRACT_BUDGET_ROOT"
elif [ "${CHECK_CONTRACT_BUDGET_ROOT+set}" = "set" ]; then
  printf 'CHECK_CONTRACT_BUDGET_ROOT is set but empty\n' >&2
  exit 2
else
  REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

[ -d "$REPO_ROOT" ] || { printf 'not a directory: %s\n' "$REPO_ROOT" >&2; exit 2; }

# shellcheck source=lib/owned-corpus.sh
source "$SCRIPT_DIR/lib/owned-corpus.sh"

# path-relative-to-repo-root<space>max-bytes, one row per covered file.
budgets() {
  cat <<'EOF'
.myflow/project.md 20396
AGENTS.md 18648
CLAUDE.md 15195
README.md 47306
commands-claude/myflow-do.md 2240
commands-claude/myflow-fast.md 2536
commands-claude/myflow-finish.md 3163
commands-claude/myflow-start.md 1815
commands-claude/myflow-status.md 1637
commands/myflow-do.md 2738
commands/myflow-fast.md 3040
commands/myflow-finish.md 3507
commands/myflow-research.md 1038
commands/myflow-start.md 2178
commands/myflow-status.md 1981
rules/agent-baseline.md 6883
rules/be-brief.mdc 7782
rules/build-the-simplest-thing.mdc 3660
rules/commit-scope-is-the-module.mdc 2998
rules/context7.mdc 2376
rules/dependency-versions.mdc 2331
rules/design-mockups-are-specs.mdc 3176
rules/dispatch-carries-the-baseline.mdc 3918
rules/kotlin-backend-development-standard.mdc 9641
rules/lint-fix-priority.mdc 2961
rules/myflow-manual-review.mdc 5630
rules/never-touch-production.mdc 2336
rules/no-direct-pushes-to-main.mdc 2416
skills/README.md 4781
skills/myflow-contracts/SKILL.md 9665
skills/myflow-contracts/artifacts-registry-rationale.md 6981
skills/myflow-contracts/artifacts-registry.md 7620
skills/myflow-contracts/build-green.md 4698
skills/myflow-contracts/finish-contract.md 54286
skills/myflow-contracts/git-boundaries-rationale.md 2317
skills/myflow-contracts/git-boundaries.md 5535
skills/myflow-contracts/handoff-blocks-rationale.md 13086
skills/myflow-contracts/handoff-blocks.md 20240
skills/myflow-contracts/jira-followups.md 45385
skills/myflow-contracts/jira-integration-rationale.md 3661
skills/myflow-contracts/jira-integration.md 19932
skills/myflow-contracts/model-policy-rationale.md 7963
skills/myflow-contracts/model-policy.md 8010
skills/myflow-contracts/operator-prompts.md 2432
skills/myflow-contracts/pipeline-rationale.md 14943
skills/myflow-contracts/pipeline.md 36155
skills/myflow-contracts/plan-provenance.md 30695
skills/myflow-contracts/project-configuration-rationale.md 16982
skills/myflow-contracts/project-configuration.md 48175
skills/myflow-contracts/session-records-rationale.md 2365
skills/myflow-contracts/session-records.md 2388
skills/myflow-contracts/state-file.md 33172
skills/myflow-contracts/workspace-isolation-rationale.md 14317
skills/myflow-contracts/workspace-isolation.md 31471
skills/myflow-contracts/worktree-resolution-rationale.md 595
skills/myflow-contracts/worktree-resolution.md 2343
skills/myflow-do/SKILL-rationale.md 27551
skills/myflow-do/SKILL.md 103578
skills/myflow-do/adversarial-reviewer-prompt.md 3545
skills/myflow-do/bug-hunter-reviewer-prompt.md 1437
skills/myflow-do/engineering-principles.md 10737
skills/myflow-do/principles-reviewer-prompt.md 14156
skills/myflow-do/security-reviewer-prompt.md 1540
skills/myflow-fast/SKILL-rationale.md 3241
skills/myflow-fast/SKILL.md 25636
skills/myflow-finish/SKILL-rationale.md 5786
skills/myflow-finish/SKILL.md 50207
skills/myflow-research/SKILL.md 4600
skills/myflow-start/SKILL-rationale.md 7030
skills/myflow-start/SKILL.md 38342
skills/myflow-status/SKILL.md 22856
EOF
}

# budget_for <path-relative-to-repo-root> — echo the declared budget, or nothing.
#
# Deliberately NOT `awk -v n="$rel"`: a filename containing a newline makes
# awk's -v assignment fail, and under `set -e` that aborts the whole run with
# exit 2 — indistinguishable from "the guard cannot answer at all", and it
# abandons every other file in the corpus rather than reporting one bad name.
# Comparing in bash keeps a pathological filename a per-file violation, which is
# what it is.
budget_for() {
  local want="$1" brel bmax
  while read -r brel bmax; do
    if [ "$brel" = "$want" ]; then
      printf '%s' "$bmax"
      return 0
    fi
  done <<<"$(budgets)"
  return 0
}

violations=0
checked=0

# check_one <path> — measure one covered file against its budget row.
check_one() {
  local path="$1" rel max actual
  checked=$((checked + 1))
  rel="${path#"$REPO_ROOT"/}"

  # `wc -c` is guarded rather than assigned bare. Under `set -e` a plain
  # assignment from a failing command aborts with THAT command's status — 1 for
  # an unreadable file — which is the code this contract reserves for "over
  # budget or undeclared", and would send a caller off to edit budgets() when
  # the guard itself could not read the file.
  # Readability is tested BEFORE the redirect rather than relying on
  # `2>/dev/null` to mop up after it. When `<` cannot open the file, bash
  # reports that itself, to the shell's stderr, before any redirection on the
  # command is established — so the suppression never applies to the one
  # failure this guard is handling, and the operator sees bash's raw line
  # followed by ours. Testing first means one message, in this guard's own
  # format.
  if [ ! -r "$path" ]; then
    printf 'cannot read %s\n' "$path" >&2
    exit 2
  fi
  if ! actual="$(wc -c < "$path" 2>/dev/null)"; then
    printf 'cannot read %s\n' "$path" >&2
    exit 2
  fi
  if ! actual="$(printf '%s' "$actual" | tr -d '[:space:]')"; then
    printf 'cannot normalise the size of %s\n' "$rel" >&2
    exit 2
  fi

  max="$(budget_for "$rel")"

  if [ -z "$max" ]; then
    printf '%s: no budget declared — add a row to budgets() in %s\n' \
      "$rel" "$(basename "${BASH_SOURCE[0]}")"
    violations=$((violations + 1))
  elif [ "$actual" -gt "$max" ]; then
    printf '%s: %s bytes exceeds budget %s — trim it back under budget, or raise its row in budgets() in %s if the growth is deliberate\n' \
      "$rel" "$actual" "$max" "$(basename "${BASH_SOURCE[0]}")"
    violations=$((violations + 1))
  fi
}

WORK="$(mktemp -d)" || { printf 'cannot create a temporary directory\n' >&2; exit 2; }
trap 'rm -rf "$WORK"' EXIT

# The covered set. owned_corpus_files returns 2, having said why on stderr, when
# it cannot answer — a scope root missing or unreadable, or a find that failed —
# and that is this guard's exit 2 unchanged.
owned_corpus_files "$REPO_ROOT" > "$WORK/files" || exit 2

# The symlink sweep the header describes. Following a link instead of refusing
# it would be worse than either silence or refusal: it would size whatever the
# link points at, anywhere this process can read, so a `.md`-named symlink added
# in a pull request turns the guard into an existence-and-size oracle for
# arbitrary paths on whoever's machine runs it. The content never leaks, but the
# size does, and none of it is this guard's business.
#
# Every scope root is known to exist and be readable by this point:
# owned_corpus_files validated them and exited 2 otherwise.
link_rc=0
{
  find "$REPO_ROOT" -maxdepth 1 -type l \( -name '*.md' -o -name '*.mdc' \) -print0 || link_rc=1
  for scope_dir in "${OWNED_CORPUS_SCOPE_DIRS[@]}"; do
    find "$REPO_ROOT/$scope_dir" -type l \( -name '*.md' -o -name '*.mdc' \) -print0 || link_rc=1
  done
} > "$WORK/links"
if [ "$link_rc" -ne 0 ]; then
  printf 'cannot enumerate symlinks under %s\n' "$REPO_ROOT" >&2
  exit 2
fi

# Both lists are sorted before they are read, so the violation lines come out in
# a stable order rather than in whatever order `find` walked the directories.
# Exit codes and counts are order-independent either way; this is presentation
# only, and it exists so this guard does not differ from
# check-normative-inventory.sh — which sorts its own output — on whether two runs
# over the same tree print the same bytes. `sort -z` keeps the NUL delimiter, so
# a filename containing a newline stays one entry.
if ! LC_ALL=C sort -z "$WORK/files" > "$WORK/files.sorted"; then
  printf 'cannot sort the covered files under %s\n' "$REPO_ROOT" >&2
  exit 2
fi
if ! LC_ALL=C sort -z "$WORK/links" > "$WORK/links.sorted"; then
  printf 'cannot sort the symlinks under %s\n' "$REPO_ROOT" >&2
  exit 2
fi

paths=()
while IFS= read -r -d '' path; do
  paths+=("$path")
done < "$WORK/files.sorted"

links=()
while IFS= read -r -d '' path; do
  rel="${path#"$REPO_ROOT"/}"
  if owned_corpus_excluded "$rel"; then
    continue
  fi
  links+=("$path")
done < "$WORK/links.sorted"

# An empty corpus is refused, not reported as a clean run. The two are
# indistinguishable on stdout, and one of them is "the self-scoped root resolved
# somewhere that holds none of this repository's text" — the false clean the
# root resolution above exists to prevent.
if [ "${#paths[@]}" -eq 0 ] && [ "${#links[@]}" -eq 0 ]; then
  printf 'no .md or .mdc file found under %s — refusing to report a clean run\n' \
    "$REPO_ROOT" >&2
  exit 2
fi

for path in ${links[@]+"${links[@]}"}; do
  checked=$((checked + 1))
  printf '%s: is a symlink — a covered file must be a regular file\n' "${path#"$REPO_ROOT"/}"
  violations=$((violations + 1))
done

for path in ${paths[@]+"${paths[@]}"}; do
  check_one "$path"
done

if [ "$violations" -gt 0 ]; then
  printf 'BUDGET-FAIL: %d file(s) over budget or undeclared\n' "$violations"
  exit 1
fi

# The count is part of the verdict deliberately: a bare "BUDGET-OK" cannot be
# told apart from a run that scanned nothing, which is the failure the
# self-scoped REPO_ROOT above exists to prevent. The number is the evidence that
# the guard looked.
printf 'BUDGET-OK: %d owned Markdown file(s) within budget\n' "$checked"
