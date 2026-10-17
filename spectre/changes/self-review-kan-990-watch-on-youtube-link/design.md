## Context

- Three non-`big` findings from the kan-990 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task; none is dropped. Checked against this worktree at `04a3a161` and the gymie
  worktree `/Users/tweety53/Projects/gymie-worktrees/kan-990-watch-on-youtube-link` at `909ade0d`:
  - A (Task 1): gymie `.flow/project.md` line 18, the `Regression suite` row, reads "Include when
    the change adds or edits a capture spec, a `.mockups` sidecar, or otherwise touches this
    checkout." **2. Isolate the workspace** (`skills/flow/implement.md`) resolves `## apps` entries
    by that column. Blast radius 1.
  - B (Task 2): `impl_paths` (`scripts/integrate-change.sh`) keeps every `<spec>/changes/*/link.md`;
    `prepare` asks `<worktree>=<impl subject>` whenever it is non-empty and `commit` stops at
    `STOP: commit-split` without one. `git-boundaries.md` **Planning commits** already names
    `chore(spectre): merge order` as the subject of the merge-order record's write. The existing
    `nested_link` case pins that a non-canonical worktree whose only delta is a `link.md` is still
    asked a subject, so the default is limited to the canonical worktree. `skills/flow/integrate.md`
    states the `NEXT:` subject rule too, so it gains a pointer clause: blast radius 4, not the 3
    the pass counted.
  - C (Task 3): no script or Go code builds a `verify`/`visual-verify`/`tooling-analysis` key — the
    suffix is prose the parent follows (`visual-verify.md` l.71, 87, 101; `verify-and-handoff.md`
    l.115; `verify-fix-loop.md` l.61, 88; `visual-verify-tooling-analysis.md` l.55) plus two guard
    header comments (`check-verify-green.sh` l.12, `check-visual-verify-dispatched.sh` l.73).
    `stats/cmd/flow/record.go` documents `-key` as what makes `begin` idempotent, and
    `visualverifydispatched.go` counts a verifier row only for its `diffBase`, so one key shared by
    two worktrees leaves one row and one `diffBase`. **Disproved for this finding:**
    `stats/internal/guard/check_panel_reproducer_exit_contract_test.go`, which the pass listed — its
    `<worktree-basename>` is the panel *citation* prefix, a different mechanism that already
    resolves repository-name prefixes (its case 45, kan-932). Blast radius 6.
- **Task 1 reaches outside `<agents repo>`.** Per **Pipeline defects found mid-run**
  (`skills/flow-contracts/pipeline.md`) it commits on the change's own branch in the gymie
  worktree above — `spectre/kan-990-watch-on-youtube-link` — never in this agents worktree.
- Spiked at plan time, every edit reverted: Task 1's diff applied in the gymie worktree
  (`git apply --check` passes at `909ade0d`). Task 2's diff applied here:
  `scripts/test-integrate-change.sh` prints `165 passed, 0 failed`; with `integrate-change.sh`
  reverted it prints `162 passed, 3 failed`, the three `link_only_canon` checks after `prepare
  exits 0`. Tasks 2 and 3 applied together: `check-references.sh`, `check-markdown-integrity.py`,
  `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`, `check-ask-silence.sh` and
  `check-normative-inventory.sh` exit 0, and `check-verbatim-moves.sh` reports 0 violations with
  this change's `verbatim-moves.txt`.

## Decisions

### The regression suite row names ui paths

**ID:** regression-row-names-ui-paths
**Status:** active
**Chosen:** extend gymie's `Regression suite` inclusion rule to any change touching a `## visual
verification` `ui paths` path — the one-cell edit the pass named.
**Considered:** `flow.visual-verify` creating the playwright worktree before `capture` — a pipeline
change across several stage files for what one project row states; narrowing to "a path whose
screen a capture covers" — no planner can tell that at kickoff, and an unneeded worktree costs
nothing.

### A link.md-only canonical worktree takes the merge-order subject

**ID:** link-only-canon-merge-order-subject
**Status:** active
**Chosen:** a `link_only` helper in `integrate-change.sh`: the canonical worktree whose every
`impl_paths` entry is a `link.md` is left out of `prepare`'s `NEXT:` subjects and committed by
`commit` as `chore(spectre): merge order` when no subject was passed.
**Considered:** every worktree whose only paths are `link.md` — breaks the `nested_link` case, where a
satellite's lone `link.md` is its real implementation and keeps its asked subject; moving the
canonical `link.md` to the planning side — changes `commit-split`'s pathspec in Go for every
project, larger than the defect.

### The key suffix is the repository name

**ID:** key-suffix-repository-name
**Status:** active
**Chosen:** `-<repository name>`, the worktree's main checkout's basename — the name
`check-visual-trigger.sh --repo` and the panel's WORKTREES prefix already use — in every key-suffix
sentence and the two guard headers that echo it; no new test, since nothing mechanical builds the
key and `check-verbatim-moves.sh` pins the reworded sentences.
**Considered:** the worktree's absolute path — unreadable in a key and never used elsewhere; making
`flow record dispatch begin` refuse a re-begin whose `-diff-base` differs under one key — a store
change on the replay path the finding does not need; a text guard refusing `-<worktree basename>`
— asserts file text, not the property.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** first review on `flow-low` (effort `low`); one fixer dispatch (key
`self-review-kan-990-watch-on-youtube-link-fix`) carrying Tasks 1–3 — each task is a verified diff
applied with `git apply`, three file-disjoint edits across two repositories.
**Considered:** `flow-medium` for the first review — Task 2's shell branch is pinned by a test that
fails without it, nothing else is code; one fixer per task — three dispatches for three
`git apply`s.

## Open questions

## Live check

None for this change's run: Task 2's behaviour is pinned by its harness case against real git
repositories, and Tasks 1 and 3 are prose.
