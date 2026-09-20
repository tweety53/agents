# Self-review context bundle for kan-613-flow-fix-flow-fast-s-plan-shape-substitution

found: 1 of 7 sources; skipped: 6 of 7 sources
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-613-flow-fix-flow-fast-s-plan-shape-substitution.md (absent)
skipped: .superpowers/sdd/reviews/kan-613-flow-fix-flow-fast-s-plan-shape-substitution-panel.md (absent)
skipped: spectre/changes/archive/kan-613-flow-fix-flow-fast-s-plan-shape-substitution/tasks.md (absent)
skipped: spectre/changes/archive/kan-613-flow-fix-flow-fast-s-plan-shape-substitution/design.md (absent)
skipped: spectre/changes/archive/kan-613-flow-fix-flow-fast-s-plan-shape-substitution/narrative.md (absent)

## git log --stat

commit bb188110859da304b88d66c04087789a5e109f4e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 01:23:17 2026 +0300

    fix(scripts): wait for the hook test stub to bind before the first case

 scripts/test-flow-active-change-hook.sh | 21 +++++++++++++++++++++
 1 file changed, 21 insertions(+)

commit e0b45a65fe42a4d98caf70e0fc9c9aa85b2c0c54
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 01:10:26 2026 +0300

    docs(flow-fast): name the Build tag the build-green close guard requires

 skills/flow-fast/SKILL.md | 4 +++-
 1 file changed, 3 insertions(+), 1 deletion(-)

## Session narrative

This run resolved KAN-613 — flow-fast's plan-shape substitution names the `**Build:**` tag the
build-green close guard requires — as a dynamic-toggled flow-fast run: plan classified `micro`, so
decide recorded the micro row's defaults (inline, no panel) and the fix landed as a three-line
addition to `skills/flow-fast/SKILL.md`'s writing-plans bullet, citing `build-green.md` rather
than copying its rule. Verification surfaced two obstacles. The first was ordinary: the fresh
worktree lacked the SPA build, so `go vet` and `tsc -b` failed until the project's own
`make web-build` ran. The second was a real find: the guard suite's one red harness,
`test-flow-active-change-hook.sh`, passed standalone every time but failed under the suite's
76-way parallel load; measurement showed the stub store refuses connections for ~145 ms after
spawn, and case 1's request fires inside that window, with the hook's fail-open turning the
refusal into a silently empty answer. The fix is the readiness wait now sitting between stub
spawn and the cases, committed as its own task; the struggle was resisting the first read of the
failure as "store unreachable" before the empty stub log proved no request was ever made.
