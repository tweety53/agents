#!/usr/bin/env bash
# test-prompt-tuning-line.sh — the Finished handoff's `**Prompt tuning:**`
# line keeps its operator-requested shape (pipeline-fix 35): the value is an
# `index.html` path or the literal NONE with no reason, and it and the Learning
# line are shown only in the Finished handoff, never an earlier one (kan-1038).
# Exit codes: 0 pass, 1 fail.
set -u
cd "$(dirname "$0")/.." || exit 1
rc=0
fail() { printf 'FAIL: %s\n' "$1"; rc=1; }
grep -qF '**Prompt tuning:** <each case'"'"'s `index.html` absolute path in `<agents repo>`'"'"'s main checkout> | NONE' skills/flow/cleanup.md \
  || fail "skills/flow/cleanup.md: Finished block lacks the '**Prompt tuning:** <path> | NONE' line"
grep -qF '| NONE —' <<<"$(grep -F 'Prompt tuning:' skills/flow/cleanup.md)" \
  && fail "skills/flow/cleanup.md: Prompt tuning NONE carries a reason"
for f in skills/flow/integrate.md skills/flow/verify-and-handoff.md skills/flow-contracts/handoff-blocks.md; do
  grep -qE '^\*\*(Learning|Prompt tuning):\*\* <' "$f" \
    && fail "$f: an earlier handoff block prints a Learning or Prompt tuning line"
done
grep -qF "**The \`Learning:\` and \`Prompt tuning:\` lines are shown only once the change is fully done**" skills/flow/cleanup.md \
  || fail "skills/flow/cleanup.md: no only-at-the-end sentence for the Learning and Prompt tuning lines"
grep -qF '`**Prompt tuning:**`' skills/flow-fast/SKILL.md \
  || fail "skills/flow-fast/SKILL.md: does not cite the '**Prompt tuning:**' line"
if grep -rnF 'Prompt-tuning:' skills rules scripts stats --include='*.md' --include='*.mdc' --include='*.sh' --include='*.go' | grep -v 'scripts/test-prompt-tuning-line.sh'; then
  fail "a 'Prompt-tuning:' spelling remains"
fi
[ "$rc" -eq 0 ] && printf 'all cases passed\n'
exit "$rc"
