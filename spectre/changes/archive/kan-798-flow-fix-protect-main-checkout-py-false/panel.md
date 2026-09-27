# Review panel — kan-798-flow-fix-protect-main-checkout-py-false

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | critical | hooks/protect-main-checkout.py:90-94 | the .worktrees exemption also exempts the existing <main>/.worktrees root itself, contradicting its own comment — rm -rf <main>/.worktrees and cd <main>/.worktrees && git reset regressed deny to allow |   |
| F2 | primary | minor | hooks/protect-main-checkout.py:196-205 | expand_vars greedy dollar-NAME replacement has no word boundary; D=<main>/docs then $Dy/f.txt is falsely denied though bash resolves $Dy unset |   |
| F3 | primary | minor | hooks/protect-main-checkout.py:181-193 | global assignment collection expands use-before-set — echo x > $C/f.txt before C=<main> is falsely denied; documented design tradeoff, recorded anyway |   |
| F4 | primary | minor | spectre/changes/kan-798-flow-fix-protect-main-checkout-py-false/proposal.md | proposal says six harness cases, the diff adds nine (26-34); tasks Tests fields omit 32-34 added by the test(hooks) fix commit |   |
| F5 | primary | minor | hooks/protect-main-checkout.py:181-193 | the module-docstring documentation Step 1 specified landed in collect_assignments docstring instead |   |
| F6 | principles | minor | hooks/protect-main-checkout.py:88-94 | the .worktrees root convention is now enforced in two unlinked places — check-worktree-location.sh:64 and the hook — with no cross-reference |   |
| F7 | primary+principles | minor | hooks/protect-main-checkout.py:187 | collect_assignments is caller-less dead code after the fix inlined collection into bash_hits — the def is its only remaining occurrence |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 withdrawn — the unlinked-ness defect is resolved by the round-0 cross-reference comment; the reproducer's single-site route would couple the personal hook's runtime to a repository lib path across arbitrary protected repos, trading a style duplication for a deployment fragility
finding-status: F7 fixed

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-worktrees-root-self-target-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-expansion-boundary-false-deny-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-use-before-set-false-deny-3.sh
finding-reproducer: F4 none — plan-artifact wording drift, nothing runnable
finding-reproducer: F5 none — docstring placement, nothing runnable
finding-reproducer: F6 .superpowers/sdd/reproducers/0-worktrees-convention-second-site-4.sh
finding-reproducer: F7 [ "$(grep -c collect_assignments hooks/protect-main-checkout.py)" -gt 1 ]

## Pass log

### Round 0

- roster: compact — 54
- no addition this round — the resolved list ran alone
- base moved 5 commits, no overlap — auto-rebased onto origin/main at 675a55a9, re-check CLEAR
- diff-size: 407 changed lines, under cap — proceed
- docs-only: no — first non-doc path hooks/protect-main-checkout.py; resolved roster runs
- fix round 0: inline fix for F1-F6; fix-mutations 3 flips + 2 none; commit 2e8ddc37
fix-mutation: hooks/protect-main-checkout.py — dropped the path!=d arm of the .worktrees exemption — case 35 rm of the .worktrees root itself stays denied
fix-mutation: hooks/protect-main-checkout.py — reverted boundary-aware expansion to greedy substring replace — case 36 unset near-name variable is not expanded greedily
fix-mutation: hooks/protect-main-checkout.py — reverted positional assignment collection to global pre-collection — case 37 use before set is not expanded
fix-mutation: spectre/changes/kan-798-flow-fix-protect-main-checkout-py-false/proposal.md — none — wording-only plan correction, no executable behaviour
fix-mutation: hooks/protect-main-checkout.py — none — docstring placement (F5) and cross-reference comment (F6) are comment-only changes
fix-mutations-total: 5

### Round 1

- fix round 0 dispatched inline — parent-applied (execution inline); agents ran: none — inline; why: decided execution inline; diff path: .superpowers/sdd/fix-round-0.diff
- cap check: fix-round diff 174 lines, under cap — proceed
- rerun: primary and principles each re-run solo — both raised the critical F1; rerun pair glm-5.3-flash/high (zcode mapping of opus/low)
- re-run: primary — F1 F2 F3 F5 fixed; F4 not fixed (proposal says nine cases, branch gains twelve); N1 new minor dead code; principles — F1 fixed, F6 withdrawal accepted; NF1 same dead-code defect deduped into F7
- auto-decided: F4 verification failed — another fix round (round 1) on F4, F7 riding it; no Critical/Important raised so no slot re-runs after round 1
- fix round 1: F4 proposal wording corrected to twelve, F7 dead helper deleted; reproducible proof: F7 instrument exits not-demonstrated post-fix; commit 3b4cf3d6
fix-mutation: spectre/changes/kan-798-flow-fix-protect-main-checkout-py-false/proposal.md — none — wording-only plan correction, no executable behaviour
fix-mutation: hooks/protect-main-checkout.py — none — deletion of an uncalled function changes no executable behaviour; the finding's own instrument (grep count) is the verification
fix-mutations-total: 2
