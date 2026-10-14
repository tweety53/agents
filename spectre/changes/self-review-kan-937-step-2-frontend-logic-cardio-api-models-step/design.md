## Context

- Five non-`big` findings from the kan-937 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task. All five were verified against the tree at `7ba0c7b6`:
  - `skills/flow/verify-and-handoff.md` **Write `IN_PROGRESS`** closes open dispatch rows but runs no
    unfinished-work check; only integrate's `prepare` runs `check-unfinished-work.sh`.
  - `stats/internal/guard/unfinishedwork.go` signal four marks a task covered by any row naming it,
    whatever its role, so an implementer-only appended task is CLEAR.
  - `spectre validate` on a scratch tree refuses `1. .` (`"1. ." is not a numbered "." or peer entry`)
    and accepts ``1. `.` `` / ``2. `peer` ``; gymie's archived `link.md` files use the backticked form.
  - `selfreview.Bundle` builds its git-log source from the one archive repository only; a satellite
    (gymie-frontend) carries its change commits on `spectre/<name>` with no planning or archive commit,
    and the canonical repository's tracked `spectre/peers` maps each peer name to a path relative to it.
  - `skills/flow/implement.md`'s **Records:** bullet forbids begin-and-end-together for an inline fix row
    only; `EndDispatch` (`stats/internal/store/records.go`) closes any row whatever its window.
- The new run-loaded sentences are acknowledged in this change's `verbatim-moves.txt`, spiked at plan time
  with every task's prose applied: `check-verbatim-moves.sh`, `check-references.sh`,
  `check-installed-citations.sh` and `check-markdown-integrity.py` all exit 0. Each task copies its prose
  verbatim from `tasks.md`, or the acknowledgement no longer matches.
- No finding was dropped.

## Decisions

### The fix run checks itself at the IN_PROGRESS handoff

**ID:** fix-run-self-check
**Status:** active
**Chosen:** a paragraph in **Write `IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`), after the
open-row close, running `check-unfinished-work.sh` per resolved worktree with integrate's arguments; an
`OUTSTANDING:` line is finished by the run itself and the guard re-run, a missing verdict or exit 2 stops
the run.
**Considered:** running it on every run reaching `IN_PROGRESS` — the finding concerns fix runs, and a
creating run's panel already closes its own findings; a new guard specific to fix runs — the existing
guard already names exactly the missing rows.

### Signal four's reviewer row is any reviewer row begun after the task's first implementer row

**ID:** reviewer-after-implementer
**Status:** active
**Chosen:** a ticked appended task needs a row naming it whose role is not `reviewer` (unchanged
coverage) and a `reviewer`-role row of the change whose `seq` is greater than the earliest such row's.
The reviewer row need not name the task: an appended task is panel-checked by the fix run's panel
(`skills/flow/document-fix.md`), whose `panel-<round>-<slot>` keys name no task. Seq is the dispatch's
append order within its change (`stats/internal/records/types.go`).
**Considered:** requiring a `task-<n>-reviewer` row — the panel's rows never name the task, so every
properly reviewed append would refuse; measuring after the latest non-reviewer row — a Minor-only
round's inline `panel-fix` row naming the task would then demand a re-review the contract waives;
comparing `startedAt` — ties at the seconds a caller writes; seq is strictly ordered.

### Merge order syntax stated as a literal block

**ID:** merge-order-literal
**Status:** active
**Chosen:** a short paragraph and a fenced ``1. `.` `` / ``2. `<peer>` `` block after the merge-order
paragraph of `skills/flow/cross-repo-worktrees.md`.
**Considered:** prose description only — the defect is that prose alone was misread.

### Peer logs appended inside the git-log source

**ID:** peer-logs-in-git-log
**Status:** active
**Chosen:** `selfreview.Bundle` reads the archived `link.md`'s `## Merge order` and `spectre/peers` out of
the archive revision, and for each entry other than `.` appends `changeBranchLog` of the peer's
`spectre/<name>` under a `### <peer>` heading to the existing `git log --stat` source. A peer that
resolves nothing is skipped silently, the bundle's "a missing source is never fatal" rule. A relative
peer path joins the canonical repository root; an absolute one is used as is.
**Considered:** one bundle source per peer — changes the found/skipped totals every existing test pins
for no reader benefit; deriving the satellite's commits by subject — satellites carry no planning or
archive commit to anchor on.

### Inline window floor enforced in the store

**ID:** inline-window-floor
**Status:** active
**Chosen:** `EndDispatch` refuses, with a new `ErrInlineWindowTooShort` mapped to 409, a row whose agent
id is `inline` (begin's or end's) when the end carries no token report and closes under ten seconds after
the row's `started_at`. Ten seconds: a begin and end in one Bash call land milliseconds apart, while real
inline work takes minutes. A token report exempts the row: its cost is measured by the report, not the
window. The store is where the row's begin time and agent id are known; the CLI knows neither.
**Considered:** a CLI-side check — would need a second read of the run record per end call; no
enforcement, prose only — the finding's recurrence is the evidence prose alone does not hold.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — task 2 changes a verdict every integrate run passes through, task 5 adds a refusal
to a record write every run makes, and task 4 changes the daemon's bundle.
**Considered:** `low` — under-reads a gate change whose false positive blocks every landing.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, tasks 1–5, one fixer in plan order — the tasks are file-disjoint and small, and
three of them are two-paragraph docs edits not worth a dispatch of their own.
**Considered:** one fixer per Go task plus one for the docs — three dispatches for work one fixer
finishes inside its context.

## Open questions

## Live check

None — the change edits a guard, the self-review bundle assembler, a store write's refusal and prose; the
Go tests exercise the guard, the assembler against real git repositories and the store against the
flow-postgres test database. The running flowd picks the store and bundle changes up on its next restart.
