# kan-279-give-the-parent-a-mutation-helper

## Why

The review panel's fix-round mutation-proof duty (`skills/flow/review-panel.md`, "The fix round
mutation-proves what it changed") is hand-run: back up a file, apply a change, run a harness, count
`FAIL:` lines, restore, verify the tree is clean. Per KAN-279's own account, KAN-77 ran "eleven fix
rounds" and "roughly 40 mutations" by hand. One was wrong: a mutation that broke the whole function
instead of one mechanism produced "extremely strong coverage" that proved nothing — "61 and 23
failures" across the two harnesses. The mechanics are identical every time; only the judgment (which
mechanism to mutate, whether a survivor is real or equivalent) needs to stay the parent's.

## What changes

- A new script, `scripts/mutate-and-verify.sh`, mechanizes the loop for a `check-*.sh` guard and one
  or more `test-*.sh` harnesses: apply a unified-diff mutation, run the harness(es) before and after,
  report the new-failure set per harness, flag a suspiciously broad blast radius rather than reporting
  it as a strong pass, then unconditionally restore and verify the touched files are clean.
- `skills/flow/review-panel.md`'s mutation-proof section points the parent at the script for the
  mechanical steps, keeping which-mechanism and real-vs-equivalent-survivor judgment explicit as the
  parent's own.
- The script is installed alongside the skill the same way every other guard/util script is (a
  symlink from `skills/flow/scripts/`), and added to `skills/flow/SKILL.md`'s guard-presence-check
  list.

No normative pipeline requirement changes — the existing "a fix round mutation-proves every behaviour
it changes" requirement is unchanged; this only mechanizes how the parent carries it out. No spec
edit.
