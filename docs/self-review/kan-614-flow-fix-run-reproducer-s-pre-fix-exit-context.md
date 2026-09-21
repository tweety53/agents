# Self-review context bundle for kan-614-flow-fix-run-reproducer-s-pre-fix-exit

found: 4 of 7 sources; skipped: 3 of 7 sources
skipped: spectre/changes/archive/kan-614-flow-fix-run-reproducer-s-pre-fix-exit/tasks.md (absent)
skipped: spectre/changes/archive/kan-614-flow-fix-run-reproducer-s-pre-fix-exit/design.md (absent)
skipped: spectre/changes/archive/kan-614-flow-fix-run-reproducer-s-pre-fix-exit/narrative.md (absent)

## change summary

# kan-614-flow-fix-run-reproducer-s-pre-fix-exit — change summary

## What changed and why

KAN-614: for the third recorded time (kan-542, kan-546, kan-556) a fix-round caller passed
`run-reproducer.sh` the raw reproducer exit where its `--pre-fix-exit` flag wanted the
dispatch-time verdict, producing a spurious ambiguity refusal before the correct call. A numeric
value cannot say which convention's verdict it means — a raw `1` (generic demonstrator) and a
verdict `1` (not demonstrated) are the same token — so the flag now takes the verdict by name.

- `scripts/run-reproducer.sh` — `--pre-fix-exit <0|1>` renamed to `--pre-fix-verdict <demonstrated|not-demonstrated>`; any other value, every number included, is a usage failure (exit 4) before the reproducer executes. The ambiguity refusal (KAN-524) compares the named verdict word-for-word and its messages name verdicts, not exit codes; behaviour is otherwise unchanged under both the generic and mutation-reproducer conventions.
- `scripts/test-run-reproducer.sh` — the ambiguity/flip cases (19–22, 26, 27) carry the named verdicts; the usage-failure loop (case 23) grew `0` and `1`, the exact misfeed values, so the KAN-614 shape is pinned as a never-executed refusal. Panel finding F1 added the RAN-marker assertion to case 20 its siblings 19/26 already had.
- `skills/flow/review-panel.md`, `scripts/check-panel-reproducer-exit-contract.sh`, `scripts/test-check-panel-reproducer-exit-contract.sh` — the fix-round re-run instruction and two comments carry the renamed flag; the prose tells the parent how to name the verdict from its `F<n>: exit $?` record (0 → `demonstrated`, 1 → `not-demonstrated`).

## What was verified and how

- Full `## lint` list green in the worktree (all guard scripts, gofmt, go vet — after the SPA deps/dist were installed and built in the worktree — and `tsc -b`); normative inventory byte-identical across the prose edit.
- Scoped `## test`: both touched harnesses green (`test-run-reproducer.sh` 105 assertions, `test-check-panel-reproducer-exit-contract.sh` 49).
- Review panel (compact roster, primary+principles on one dispatch): zero Critical/Important; one Minor (F1) fixed inline and proved by its own reproducer (exit 0 pre-fix → 1 post-fix); findings closed; principles pass compliant.

## Deliberately left out

- No back-compat alias: `--pre-fix-exit` is refused outright, which is the fix — accepting the old numeric form would keep the misfeed expressible.
- The historical `docs/self-review/*` reports keep their original `--pre-fix-exit` mentions; they are records of past runs, not instructions to current callers.
## .superpowers/sdd/ledgers/kan-614-flow-fix-run-reproducer-s-pre-fix-exit.md

# SDD ledger — kan-614-flow-fix-run-reproducer-s-pre-fix-exit

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-1-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T17:55:20Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-614-flow-fix-run-reproducer-s-pre-fix-exit-panel.md

# Review panel — kan-614-flow-fix-run-reproducer-s-pre-fix-exit

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | scripts/test-run-reproducer.sh:519-522 | case 20 asserts the ambiguity refusal but, unlike sibling cases 19/26, never asserts the fixture's RAN marker, so a refuse-without-running regression passes it clean. Pre-existing at the merge base but on an assertion line this diff rewrote. |   |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/reproducers/1-primary-1.sh

## Pass log

### Round 0

- roster: compact — 78
- diff size: 146 under cap — exit 0, proceeding
- docs-only: exit 1 — resolved roster runs (first non-doc path: scripts/check-panel-reproducer-exit-contract.sh); no operator-named slot this round

### Round 1

fix-mutation: scripts/test-run-reproducer.sh — case 20's refuse-without-running gap (stub runner refusing exit-2-naming-convention without executing the fixture) — F1 reproducer .superpowers/sdd/reproducers/1-primary-1.sh: exit 0 pre-fix → exit 1 post-fix
fix-mutations-total: 1
## git log --stat

commit f78ed2b3f0760ed129388d66ca0c56d86af79580
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:16:42 2026 +0300

    test(reproducer): make case 20 assert the reproducer ran before the refusal

 scripts/test-run-reproducer.sh | 6 ++++++
 1 file changed, 6 insertions(+)

commit 6eee79a66f4d7dc24aaca0da4ca2fbbc82dce8ed
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:50:19 2026 +0300

    docs(flow): carry the renamed verdict flag in the panel prose

 scripts/check-panel-reproducer-exit-contract.sh      | 2 +-
 scripts/test-check-panel-reproducer-exit-contract.sh | 4 ++--
 skills/flow/review-panel.md                          | 7 ++++---
 3 files changed, 7 insertions(+), 6 deletions(-)

commit 0440c2e862043fc7d6c42ce855c280e53dbd20e2
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:47:50 2026 +0300

    fix(reproducer): take the pre-fix verdict by name, not as a raw exit

 scripts/run-reproducer.sh | 54 ++++++++++++++++++++++++++++-------------------
 1 file changed, 32 insertions(+), 22 deletions(-)

commit 0ce7d784700b27e775ce62514da654f19aba96fd
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:45:20 2026 +0300

    test(reproducer): expect --pre-fix-verdict and refuse raw-exit values

 scripts/test-run-reproducer.sh | 79 +++++++++++++++++++++++-------------------
 1 file changed, 43 insertions(+), 36 deletions(-)

## Session narrative

This `/flow-fast` run resolved KAN-614 to the rename branch of the issue's own fix options —
`--pre-fix-exit <0|1>` became `--pre-fix-verdict <demonstrated|not-demonstrated>` — because the
validate-the-number alternative is unimplementable: at re-run time the runner cannot distinguish a
raw `1` from a verdict `1`, so only a value vocabulary that names the printed verdicts catches the
misfeed at the flag. All three dynamic toggles fired, so the run wrote a plan-shaped `tasks.md`
(two `check-plan-shape.sh` rejections first: multi-line `**After:**` values, then the same field
used as prose where the grammar wants `Task <ids>`/`none` — dependency edges, not outcome notes),
ran `plan-class.sh` to `class: small`, decided inline execution with a compact
`primary+principles` panel, and dispatched the panel as one subagent with both role briefs
inlined, this harness registering no `flow-high` type. Implementation went red-first (cases 19–27
rewritten for the named verdicts, misfeed values `0`/`1` added to the usage loop), then the
runner rename, then the prose cross-references with the normative inventory diffed clean. The
panel raised one Minor — case 20 lacked the never-executed assertion its siblings carry — fixed
inline per the re-runs contract's triviality bar and proved by the finding's own reproducer
flipping 0 → 1. Where it struggled: the worktree needed `stats/web`'s npm install and dist build
before `go vet`'s embed pattern answered, one compound shell call lost its cwd and re-run with
explicit paths, and the panel noted the pre-panel citation scan subshell could not find
`check-references` on its PATH (the direct run exits 0; informational only).
