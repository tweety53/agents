# Self-review context bundle for kan-619-flow-fix-check-references-reads-any-backticked

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-619-flow-fix-check-references-reads-any-backticked/tasks.md (absent)
skipped: spectre/changes/archive/kan-619-flow-fix-check-references-reads-any-backticked/design.md (absent)
skipped: spectre/changes/archive/kan-619-flow-fix-check-references-reads-any-backticked/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-619-flow-fix-check-references-reads-any-backticked.md

# SDD ledger — kan-619-flow-fix-check-references-reads-any-backticked

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T18:10:36Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-619-flow-fix-check-references-reads-any-backticked-panel.md

# Review panel — kan-619-flow-fix-check-references-reads-any-backticked

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | minor | scripts/check-installed-citations.sh:208 | the new hint (mark the line citations-guard:allow) also prints on coverage-only violation runs, where marking cannot fix anything, including the exact state the feature creates (a member whose only citation line was just marked exits 1 with 0 checked, and not declared expected-zero plus this hint) |   |
| F2 | primary+principles | minor | scripts/test-check-installed-citations.sh:741 | the section comment claims every exempt case is paired with the same line unmarked and reported, but allow-marker-exempts-a-command-shape and allow-marker-in-shell-comment-is-exempt have no unmarked counterparts, so a future classifier change silently passing those shapes keeps all marker cases green |   |
| F3 | primary+principles | minor | scripts/check-installed-citations.py:121 | the claimed marker-to-zero declared-zero interaction is real (verified live) but pinned by no harness case |   |
| F4 | primary | minor | scripts/check-installed-citations.py:24 | the module docstring exit-code contract still says Fix the named file:line by prefixing the citation; never suppress., now stale against the declared marker this diff adds and the wrapper advertises |   |
| F5 | principles | minor | scripts/check-installed-citations.sh:208 | the marker literal is duplicated as prose in the wrapper while its definition lives in CITATION_ALLOW_MARKER (check-installed-citations.py:124), and no case pins the hint text, so a rename drifts the guidance silently |   |

findings-total: 5
finding-status: F1 deferred the wrapper prints one hint beside every violation class; scoping it to citation rows means class-aware reporting the wrapper does not have
finding-status: F2 deferred two exempt cases lack unmarked twins; closing it adds two new harness cases
finding-status: F3 deferred the marker-to-zero interaction is pinned by no case; closing it adds a coverage-checker case
finding-status: F4 fixed
finding-status: F5 deferred the hint is wrapper-owned prose; deriving it from the Python constant crosses the wrapper-classifier split

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- roster: compact — 34
- no addition this round — the resolved list ran alone
- diff size: within cap — proceeding
- docs-only: no — first non-doc path scripts/check-installed-citations.py; resolved roster primary+principles runs
## git log --stat

commit 371b007ea8c39b6c9d8b582a9709c6c77bb5d683
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:26:48 2026 +0300

    docs(scripts): state the declared marker beside the guard's exit-code contract

 scripts/check-installed-citations.py | 4 +++-
 1 file changed, 3 insertions(+), 1 deletion(-)

commit 01ccdbcff60a5e91cc0220a96b06977f1e28ffa0
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:07:13 2026 +0300

    test(scripts): pin the installed-citations guard's declared non-path marker
    
    Five cases: the kan-548 n/a shape reported unmarked and exempt marked,
    the kan-561 command shape exempt marked, the marker proven line-scoped
    against an unmarked unrooted citation, and a shell-fence comment line
    exempt. The unmarked/report twin is the mutation proof that the marker,
    not a classifier change, turns the report off.

 scripts/test-check-installed-citations.sh | 53 +++++++++++++++++++++++++++++++
 1 file changed, 53 insertions(+)

commit d2de210a7101ce504101e6636ae8e08e8cee72cc
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:07:01 2026 +0300

    feat(scripts): exempt a declared non-path marker line in the installed-citations guard
    
    A line carrying citations-guard:allow is skipped wholesale — the
    declared convention for non-path tokens (KAN-619), sibling of
    check-references.sh's refs-guard:allow. It retires the per-token
    rewording commits: kan-548's backticked 'n/a — no frame' marker and
    kan-561's backticked 'git add -- <the spec, ...>' command shape both
    classified as citations whose first segment names no root, and each
    trip cost a rewording. The classifier's own exclusions are untouched.

 scripts/check-installed-citations.py | 28 ++++++++++++++++++++++++++++
 scripts/check-installed-citations.sh |  1 +
 2 files changed, 29 insertions(+)

## Session narrative

This run fixed the citation-guard false-positive class KAN-619 records. The issue attributes the trips to check-references.sh, but the archaeology (commits ec18a67 "write the no-frame marker unquoted so the citation guard passes" and 3f3f77d, which split a backticked command shape into prose) showed both recorded false positives — kan-548's backticked `n/a — no frame` and kan-561's backticked `git add -- <the spec, its PNGs, .../>` placeholder list — came from check-installed-citations' word-level span classifier, so the declared exemption lands there: a line carrying `citations-guard:allow` is skipped wholesale, the sibling of check-references' existing `refs-guard:allow`. Implementation was test-first: five harness cases written before the guard change, three red against the recorded shapes, then the classifier skip, the wrapper hint line and the docstring amendment. The review panel (compact roster, primary+principles in one dispatch) raised five Minors: the docstring's stale never-suppress line was fixed inline; the other four — the hint also printing beside coverage-only violations, two exempt cases lacking unmarked twins, the unpinned marker-to-zero interaction, and the marker literal duplicated as wrapper prose — deferred with category and mechanism per the Minor default, no slot re-running. Where it struggled: the line-scoped harness case first used a `../`-rooted token, which classify_token excludes outright, so the case switched to a plain unrooted path; and the worktree needed `make web-build` before go vet could compile the embedded SPA.
