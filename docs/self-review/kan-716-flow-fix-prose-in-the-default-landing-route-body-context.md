# Self-review context bundle for kan-716-flow-fix-prose-in-the-default-landing-route-body

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-716-flow-fix-prose-in-the-default-landing-route-body/tasks.md (absent)
skipped: spectre/changes/archive/kan-716-flow-fix-prose-in-the-default-landing-route-body/design.md (absent)
skipped: spectre/changes/archive/kan-716-flow-fix-prose-in-the-default-landing-route-body/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-716-flow-fix-prose-in-the-default-landing-route-body.md

# SDD ledger — kan-716-flow-fix-prose-in-the-default-landing-route-body

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-06T23:03:23Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: bec88075
- Outcome: completed
- Started: 2026-10-06T23:19:20Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 79058f34f81a2f4bfe96a6a6d80c14c4e6f7562c
- Outcome: completed
- Started: 2026-10-06T23:30:41Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-716-flow-fix-prose-in-the-default-landing-route-body-panel.md

# Review panel — kan-716-flow-fix-prose-in-the-default-landing-route-body

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | stats/internal/guard/project_get_test.go:146 | task 1's **Tests:** field names a nonempty word-extension case; the diff adds 'merge and pushed' instead and nonempty appears nowhere in the suite — add the named case or amend the task text (behavior verified correct, exit 3) |   |
| F2 | primary | minor | stats/internal/guard/projectget.go:98 | --enum is first-match-wins, so a literal that word-boundary-prefixes another shadows it; cannot fire with the four documented vocabularies — settle the tie-break in the contract only if a vocabulary ever overlaps |   |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh

## Pass log

### Round 0

- roster: compact — 87
- no addition this round — the resolved list ran alone
- diff size: 51 lines, cap not exceeded — proceed
- docs-only reduction: no (first non-documentation path scripts/project-get.sh) — resolved roster runs: primary+principles

### Round 1

- FIX_BASE 79058f34f81a2f4bfe96a6a6d80c14c4e6f7562c — round-boundary sync rebased onto f527f924 (no overlap), branch rewritten
- panel-fix-1 ran: F1 (important) and F2 (minor) fixed in commit 49ddf520 — longest-match tie-break in pgEnum plus the named nonempty and punctuation cases; read fix-round-1.diff; no finding bounced
- cap check from primary-held sha 170dbc76: under cap; docs-only: no — re-run reads fix-round-1.diff
- re-run: primary re-runs — it raised F1 and F2, both touched by the fix; principles raised nothing and keeps its result
- re-run clean: F1 fixed, F2 fixed, no new finding — the round closes on it
fix-mutation: stats/internal/guard/projectget.go — the longest-match tie-break comparison ( > flipped to <) — TestProjectGetEnum/the_fix_diffs_tie-break_case (shadow case: merge and push over merge)
fix-mutation: stats/internal/guard/projectget.go — the pgWordByte boundary guard (!pgWordByte flipped to pgWordByte) — TestProjectGetEnum nonempty handoff-vocabulary case
fix-mutations-total: 2
## git log --stat

commit 49ddf5205fa7823747a974e816f5f0dd9eaf71be
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Oct 7 02:27:01 2026 +0300

    fix(guard): longest match wins the --enum tie-break; add the named nonempty case
    
    F1: task 1 names a nonempty word-extension case against the handoff
    literals (required, none); the suite carried only route-vocabulary
    extensions. TestProjectGetEnum now carries it as a subtest, plus a
    punctuation-after-literal case.
    
    F2: pgEnum was first-match-wins, so a literal word-boundary-prefixing
    a longer one shadowed it; the longest matching literal now wins.

 stats/internal/guard/project_get_test.go | 11 +++++++++++
 stats/internal/guard/projectget.go       | 20 +++++++++++++-------
 2 files changed, 24 insertions(+), 7 deletions(-)

commit 79058f34f81a2f4bfe96a6a6d80c14c4e6f7562c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Oct 7 01:58:42 2026 +0300

    docs(flow): state the leading-literal match for single-line-literal keys
    
    The resolution paragraph in project-configuration.md now says a head matches
    a row literal on equality or when the literal leads it at a word boundary —
    prose after the literal on the head line, like the lines below it, is
    documentation for the reader, never read — and the four key rows defer to it
    instead of restating a byte-exact clause the match no longer has. The
    project-get.sh header states the same rule in place of its byte-for-byte
    wording. The reworded and new sentences are listed in the change's
    verbatim-moves.txt as the guard prints them (kan-716).

 scripts/project-get.sh                         |  6 ++++--
 skills/flow-contracts/project-configuration.md | 22 ++++++++++++----------
 2 files changed, 16 insertions(+), 12 deletions(-)

commit bec8807533f0d61c2791638038423d6a806ccfb6
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Oct 7 01:55:49 2026 +0300

    fix(project-get): match a single-line-literal key on its leading literal
    
    pgEnum matched the head byte-for-byte against the literals, so prose the
    operator writes after the literal on the head line — 'merge and push — the
    standing choice since kan-512' — exited 3 and dropped the configured value,
    sending every run back to its ask or fallback (kan-716; prose below the head
    was already never read). The head now matches on equality or on a leading
    literal whose next byte cannot extend it into a longer word (letter, digit,
    underscore): 'manually' and 'merge and pushed' still match nothing, and the
    exit-3 stderr line is unchanged.

 stats/internal/guard/project_get_test.go |  5 +++++
 stats/internal/guard/projectget.go       | 18 ++++++++++++++----
 2 files changed, 19 insertions(+), 4 deletions(-)

## Session narrative

The session resolved KAN-716 — prose defeating the single-line-literal resolution of `.flow/project.md` keys — and found the defect narrower than the ticket states: the `--enum` head resolution that landed in 5b64e1a9 already ignores prose *below* the head, so what remained was prose typed onto the head line itself, which still exited 3 and silently dropped the configured route. The fix took the ticket own first proposal — a leading literal, matched at a word boundary — over the ticket alternative (documenting that the section must carry the bare literal only), because documenting a footgun leaves the silent loss of automation in place; the word-boundary guard (letters, digits, underscore) was chosen over a bare prefix so `manually` never resolves to `manual`. Two approaches were tried and abandoned before the round closed: the first pass-1 decision recorded the compact panel as two single-role dispatches, re-grouped to the one bundled `primary+principles` dispatch the plan-class tree printed after the planner re-read the grouping rule; and the fix round initially had only the route-vocabulary word-extension case in hand until the primary reviewer pinned that the plan named `nonempty` — vocabulary for `## handoff`, not the landing route — which forced the handoff-vocabulary subtest and, with it, the first-match-wins shadowing defect (F2) the code had carried all along. The run paid for two round-boundary auto-rebases (origin/main moved twice mid-review, no overlap either time) and one reproducer repair: the reviewer scripts declared premises as bare paths until the exit-contract guard demanded the full `path:line:content` form. Minors fixed inline alongside the Important in one commit; both reproducers flipped demonstrated to not-demonstrated, and the delta re-run came back clean.
