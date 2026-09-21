# Self-review context bundle for kan-616-harness-case-pins-flow-records-addr

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-616-harness-case-pins-flow-records-addr/tasks.md (absent)
skipped: spectre/changes/archive/kan-616-harness-case-pins-flow-records-addr/design.md (absent)
skipped: spectre/changes/archive/kan-616-harness-case-pins-flow-records-addr/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-616-harness-case-pins-flow-records-addr.md

# SDD ledger — kan-616-harness-case-pins-flow-records-addr

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T18:01:44Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 576f6213a2ede8135868bfbb84f736428098ccf0
- Outcome: completed
- Started: 2026-09-21T18:43:56Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T18:44:32Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T18:44:32Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-616-harness-case-pins-flow-records-addr-panel.md

# Review panel — kan-616-harness-case-pins-flow-records-addr

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | scripts/test-flow-addr-declaration.sh:113-118 | Assertion 5 counts registrations but cannot attribute them: an inline registration through resolveDefaultAddr() inside a record-family file — a new flow record verb resolving FLOW_ADDR where the declaration says the whole flow record family resolves FLOW_RECORDS_ADDR — passes all six cases |   |
| F2 | primary+principles | minor | scripts/test-flow-addr-declaration.sh:183 | the m-decl sed eats the closing paren, so the mutated sentence stops matching and assertion 1 fires before assertion 3 is reached — no drift escapes, but the suite never proves assertion 3 detects a declared-side shrink, the property the case name claims |   |
| F3 | primary+principles | minor | scripts/test-flow-addr-declaration.sh:109 | Assertion 4 greps the literal list record.go selfreview.go while assertion 3 derives the identical fact from $declared_files — one fact, two representations; a declared family file calling both seams then passes assertions 2-5 silently |   |
| F4 | primary+principles | minor | scripts/test-flow-addr-declaration.sh:162,169,176,183,188 | the mutations use sed -i "" the BSD-only form where sibling harnesses use the portable sed -i.bak; on GNU the mutations cannot apply and a clean tree reports drift — fail-visible, but darwin-locked when the portable form is the established convention two files over |   |
| F5 | primary+principles | minor | scripts/test-flow-addr-declaration.sh:91,101,115,116 | the $srcs expansions are unquoted, so a REAL_ROOT whose path contains a space splits find output into wrong grep arguments and a clean tree is reported dirty — fail-visible, but a legitimate root layout is rejected |   |
| F6 | principles | minor | scripts/test-flow-addr-declaration.sh:214 | the rewritten m-decl comment says the case fails through assertion 4's set comparison; the set comparison is assertion 3 — introduced by the fix diff, behavior correct, cross-reference wrong |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 deferred quote-hardening is an array-vs-xargs mechanism choice across five expansions, and the failure is fail-visible on the space-free paths every real run of this repo takes
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 none — no GNU sed on this host, the divergence cannot be run here
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F6 none — comment-only cross-reference, no executable behaviour

## Pass log

### Round 0

- roster: compact — 53
- no addition this round — the resolved list ran alone
- diff size 195 lines, under cap — proceeding
- docs-only: no — scripts/test-flow-addr-declaration.sh; resolved roster runs
- reproducers re-authored before dispatch: added the # demonstrates declaration the KAN-606 instrument audit requires to 0-primary-1/2/3/4.sh; fresh verdicts and shas taken below

### Round 1

- panel-fix ran inline (execution mode inline): F1-F4 fixed by the parent, F5 deferred cosmetic; reproducers re-verified — F1/F2/F3 not demonstrated post-fix, F4 exempt (no GNU sed); fix diff touches scripts/test-flow-addr-declaration.sh as every finding named
fix-mutation: scripts/test-flow-addr-declaration.sh — inline resolveDefaultAddr() -addr registration through StringVar(new(string)) appended to a record-family file (the zap shape) — 1->0 reproducer verdict: harness green -> rejected on that tree
fix-mutation: scripts/test-flow-addr-declaration.sh — m-decl mutation sed ate vs kept the closing paren — assertion-1-parse -> assertion-3-set-inequality tripped reason
fix-mutation: scripts/test-flow-addr-declaration.sh — suite.go declared into the record family while calling both seams — pass-unmutated green -> red on the leak tree
fix-mutation: scripts/test-flow-addr-declaration.sh — none — no GNU sed on this host — portability not measurable here; mutation sites rewritten to the sibling sed -i.bak convention
fix-mutations-total: 4
## git log --stat

commit d3cb7a08544c456fb4a4b991f7e92a4e21646677
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:49:39 2026 +0300

    fix(harness): correct the m-decl comment's assertion cross-reference

 scripts/test-flow-addr-declaration.sh | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)

commit 576f6213a2ede8135868bfbb84f736428098ccf0
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:43:53 2026 +0300

    fix(harness): attribute -addr registrations to the declared record family's files

 scripts/test-flow-addr-declaration.sh | 78 +++++++++++++++++++++++++----------
 1 file changed, 57 insertions(+), 21 deletions(-)

commit 0c7699b8b61b9a614c430aca0b72e5ac6d4a427c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:58:26 2026 +0300

    test(harness): pin the addr-resolution command sets against the isolation table

 scripts/test-flow-addr-declaration.sh | 195 ++++++++++++++++++++++++++++++++++
 1 file changed, 195 insertions(+)

## Session narrative

A `/flow-fast` run against KAN-616: one new guard harness, `scripts/test-flow-addr-declaration.sh`, pinning the set of `flow` CLI commands whose `-addr` default resolves `FLOW_RECORDS_ADDR` (the record family: `flow record`, `flow self-review bundle`) against `.flow/project.md`'s `## workspace isolation` sentence, with mutation cases proving each assertion detects the drift it names. All three project toggles are `dynamic`, so the run wrote a plan-shaped `tasks.md`, classified the change `small` via `plan-class.sh`, and recorded a compact panel (primary+principles, one dispatch). The panel raised one Important — the registration census could not attribute registrations, so an inline `resolveDefaultAddr()` registration inside a record-family file passed every case — plus four Minors; the parent fixed F1–F4 inline and deferred one quoting nit. Where it struggled: the declaration's hard-wrapped prose defeated two parse anchors before the mid-line one landed; F1's first fix anchored on the `&f.addr` registration shape and the reviewer's own `StringVar(new(string), ...)` reproducer defeated it, forcing the variable-agnostic `"addr", resolve...` pattern; the F2 reproducer had to be re-authored to lift the mutation the harness actually commits (its first hard-coded copy kept demonstrating the pre-fix behaviour, and its lifted command referenced the harness's own `$SANDBOX` variable and `m-decl` sandbox layout); and `run-reproducer.sh`'s re-run flag is `--pre-fix-verdict demonstrated`, not the `--pre-fix-exit <code>` shape the panel contract's prose suggests. Both re-run slots confirmed every listed finding fixed with no new defect above a comment cross-reference, fixed in `d3cb7a0`. Full lint list and the scoped test green at `576f621` + `d3cb7a0`.
