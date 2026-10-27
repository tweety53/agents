#!/usr/bin/env bash
# test-prompt-tuning-line.sh — the Finished handoff's `**Prompt tuning:**`
# line keeps its operator-requested shape (pipeline-fix 35): the value is an
# `index.html` path or the literal NONE with no reason, and the line is stated
# as present in every /flow and /flow-fast run's final message.
# Exit codes: 0 pass, 1 fail.
set -u
cd "$(dirname "$0")/.." || exit 1
rc=0
fail() { printf 'FAIL: %s\n' "$1"; rc=1; }
grep -qF '**Prompt tuning:** <each case'"'"'s `index.html` absolute path in `<agents repo>`'"'"'s main checkout> | NONE' skills/flow/cleanup.md \
  || fail "skills/flow/cleanup.md: Finished block lacks the '**Prompt tuning:** <path> | NONE' line"
for f in skills/flow/cleanup.md skills/flow/integrate.md skills/flow/verify-and-handoff.md skills/flow-contracts/handoff-blocks.md; do
  grep -qF '**Prompt tuning:** <each case'"'"'s `index.html` absolute path in `<agents repo>`'"'"'s main checkout> | NONE' "$f" \
    || fail "$f: handoff block lacks the '**Prompt tuning:** <path> | NONE' line"
  grep -F 'Prompt tuning:' "$f" | grep -qF '| NONE —' \
    && fail "$f: Prompt tuning NONE carries a reason"
done
grep -qF "**The \`Prompt tuning:\` line is present in every \`/flow\` and \`/flow-fast\` run's final message.**" skills/flow/cleanup.md \
  || fail "skills/flow/cleanup.md: no always-present sentence for the Prompt tuning line"
grep -qF '`**Prompt tuning:**`' skills/flow-fast/SKILL.md \
  || fail "skills/flow-fast/SKILL.md: does not cite the '**Prompt tuning:**' line"
if grep -rnF 'Prompt-tuning:' skills rules scripts stats --include='*.md' --include='*.mdc' --include='*.sh' --include='*.go' | grep -v 'scripts/test-prompt-tuning-line.sh'; then
  fail "a 'Prompt-tuning:' spelling remains"
fi
[ "$rc" -eq 0 ] && printf 'all cases passed\n'
exit "$rc"
