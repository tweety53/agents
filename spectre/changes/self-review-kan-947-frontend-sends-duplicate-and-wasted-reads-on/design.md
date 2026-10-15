## Context

- Five non-`big` findings from the kan-947 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task. All were verified against the tree at `e47bb947`:
  - 1a: `rrDefaultBnd = "20"` (`stats/internal/guard/runreproducer.go`); `rrDeadline` accepts any
    all-digit `RUN_REPRODUCER_BOUND_SECONDS`, larger or smaller, so the comment's "can only ever
    shrink the wait" is false. Only `docs/self-review/kan-571-…-self-review.md` names the variable;
    no run-loaded file does. An exit-contract run that times out is that guard's exit 2, which
    stops the run.
  - 1b: `pcResolveCitation` and `pcDeclPath` (`stats/internal/guard/panelexitcontract.go`) resolve a
    `<repo>:` prefix on both declaration kinds; the WORKTREES paragraph of
    `skills/flow/review-panel.md` asks the prefix of a finding's `file:line` only.
  - 1c: spiked — `flow state get no-such-change-xyz -C <worktree>` exits 1 with nothing on stdout
    and `flow: no state recorded for agents-a740d89c/no-such-change-xyz` on **stderr**
    (`stats/cmd/flow/state.go:378`), not on stdout as the finding guessed; `state.go` also exits 1
    on a project-key failure, which is not no state.
  - 1e: `/Users/tweety53/Projects/gymie/.flow/project.md`'s `ui paths` holds `gymie-frontend/**`;
    `gymie-frontend/.flow/project.md` declares no `## visual verification`, so its worktree's
    trigger exits 2 not-configured and the canonical worktree's diff (its change directory only)
    exits 1. Two callers ask the per-worktree question: `flow.visual-verify` step 1–2
    (`skills/flow/verify-and-handoff.md`) and the full-suite visual run (`skills/flow/implement.md`).
  - 2a: **Inline verify** (`skills/flow/verify-and-handoff.md`) runs per worktree of the resolved
    set with no skip. `check-verify-green.sh` needs at least one completed row, which the
    non-skipped worktree still records.
- Every task was spiked together at `e47bb947`: Go vet, gofmt and the whole `go test ./...` pass;
  task 4's test fails without its guard change; `check-verbatim-moves.sh` (with this change's
  `verbatim-moves.txt`), `check-references.sh`, `check-markdown-integrity.py`,
  `check-guard-symlinks.sh`, `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`,
  `check-stage-mark-calls.sh` and `check-normative-inventory.sh` exit 0.
- No finding was dropped; 1c's premise about stdout was corrected to stderr.

## Decisions

### Document the bound override rather than raise the default

**ID:** reproducer-bound-override
**Status:** active
**Chosen:** correct the comment (the override is unclamped; a larger bound only lets a run finish)
and state in `review-panel-fix-round.md` that every run of a test-building reproducer — the
exit-contract guard, the per-finding run and its re-run — carries `RUN_REPRODUCER_BOUND_SECONDS`.
**Considered:** raising `rrDefaultBnd` — every hung reproducer then waits the longer bound;
clamping the override to shrink only, as the comment claimed — removes the one way a Gradle
reproducer reaches a verdict.

### Prefix reproducer declarations in the WORKTREES paragraph

**ID:** prefix-reproducer-declarations
**Status:** active
**Chosen:** one clause in the WORKTREES paragraph: `# demonstrates:` and `# premise:` paths citing a
worktree other than the canonical one carry the same `<repo>:` prefix. That paragraph is the one a
cross-repo slot receives.
**Considered:** a clause in the reproducer authoring rule (`review-panel.md`'s `# demonstrates:`
paragraph) — read by single-worktree runs too, where no prefix exists.

### No state is named by its stderr line

**ID:** no-state-stderr-line
**Status:** active
**Chosen:** **Reading the state** runs `flow state get` as its own Bash call and names no state as
exit 1 with `flow: no state recorded for <project>/<name>` on stderr; an exit 1 without that line
is a failure to report.
**Considered:** "the no-record line on stdout", the finding's wording — the spike shows stdout
empty and the line on stderr.

### A `--repo` flag on the visual trigger

**ID:** visual-trigger-repo-prefix
**Status:** active
**Chosen:** `check-visual-trigger.sh [--repo <name>] <root>` matches every stdin path as
`<name>/<path>`. In a run whose resolved set holds more than one worktree, every other worktree's
diff is piped through it against the canonical worktree, `<name>` that worktree's main checkout's
basename — the name `ui paths` spells it with — and its exit 0 makes the canonical worktree survive
step 2. `implement.md`'s full-suite visual clause counts the same run.
**Considered:** prefixing in prose with `sed` — the glob question belongs to the guard, which owns
its semantics; deriving the name inside the guard from a `--from <worktree>` argument — git
plumbing in a guard that today reads no git, for a name the WORKTREES paragraph already defines;
matching against the canonical project only when the worktree's own project is not configured —
misses a frontend whose own project declares a section.

### Skip by a change-directory diff, not check-panel-docs-only.sh

**ID:** verify-skip-change-dir-only
**Status:** active
**Chosen:** in a run whose resolved set holds more than one worktree, **Inline verify** skips a
worktree for which `git -C <worktree> diff --name-only <merge-base>..HEAD ':(exclude)spectre/changes/<name>' ':(exclude)spectre/changes/archive/<name>'`
prints nothing; the skipped worktree records no row and the `## Report` names it.
**Considered:** `check-panel-docs-only.sh`, the finding's suggestion — it exits 0 for any diff of
only `.md`/`.mdc` files, so it would also skip a worktree whose real documentation edits its own
`## lint` judges (this repository's markdown guards); skipping in single-worktree runs too — such a
run's only worktree carries the implementation, so the case does not arise and the narrower rule
changes nothing there.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — task 4 changes a guard's argument parsing and two pipeline call sites, and
task 5 changes which worktrees verify runs for; both alter run behaviour a low-effort read could
pass over.
**Considered:** `low` — three of the five tasks are one prose clause each, but the two that are not
carry the review risk.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, tasks 1–5, one fixer in plan order — small tasks; tasks 4 and 5 share
`skills/flow/verify-and-handoff.md` and run in that order.
**Considered:** two groups splitting the Go task from the prose — two dispatches for work one fixer
finishes inside its context.

## Open questions

### Integrate's visual-verify check is still per-worktree

**ID:** integrate-visual-check-cross-repo
**Status:** open
**Why it is open:** `check-visual-verify-dispatched.sh` (`stats/internal/guard/visualverifydispatched.go`)
asks the trigger for one worktree and one merge base, so for a frontend-only cross-repo change it
still answers `VISUAL-VERIFY-OK: no UI paths touched` without a verifier; closing it needs the
other worktrees and their merge bases passed in — a larger change than this pass's findings.
**What it affects:** whether integrate's run 1 catches a skipped visual verify on such a change.

## Live check

None — the change edits one guard's argument handling and pipeline prose; the guard's Go test drives
it end to end, and no daemon, store or other runtime state changes.
