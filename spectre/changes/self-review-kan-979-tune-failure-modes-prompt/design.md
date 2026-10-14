## Context

- Three non-`big` `flow-fix` findings from the kan-979 self-review pass, planned as one change per item 1
  of **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task.
- Verified at `acc2a4b8`: the kan-979 narrative's in-run fix line carries no sha and no blast radius; no
  file under `skills/`, `scripts/`, `hooks/` or `rules/` runs `flow record summary`; the store holds
  implementer rows for kan-979's planned tasks and none for its fix-round task.
- New run-loaded sentences each carry a `(`skills/….md`)` citation, so `check-verbatim-moves.sh` passes
  without an acknowledgement file (spiked at plan time).

## Decisions

### Narrative format check lives in integrate-change.sh commit

**ID:** narrative-check-in-commit
**Status:** active
**Chosen:** an inline check at the top of `cmd_commit` in `scripts/integrate-change.sh`, stopping
`STOP: narrative` before the reshape — the narrative is appended just before `commit`, so that is the
first point it exists and nothing has been rewritten yet.
**Considered:** a new `check-narrative-*.sh` guard plus its `test-check-*.sh` companion and a lint row —
three files and a guard for one grep with one caller; a check in `prepare` — the narrative is not
written yet when `prepare` runs.

### The sha may be backticked and "file" singular

**ID:** narrative-format-tolerance
**Status:** active
**Chosen:** accept `` `<sha>` `` as well as `<sha>`, a 7- to 40-hex sha or `deferred`, and
`(blast radius <N> file)` as well as `files` — the readable variants of the contract's own format.
**Considered:** the contract's literal text only — refuses a line a reader and the self-review pass both
read correctly.

### Summary recording stated once, in pipeline.md

**ID:** summary-record-in-pipeline
**Status:** active
**Chosen:** the `flow record summary` call is stated in **Summary and live-stack line, before every
handoff** (`skills/flow-contracts/pipeline.md`); `skills/flow/verify-and-handoff.md` cites it at the
`IN_PROGRESS` block, where the summary is printed.
**Considered:** the command written in both files — the same instruction stated twice.

### An unrecorded fix round is caught by a missing implementer row

**ID:** undispatched-task-signal
**Status:** active
**Chosen:** signal four of `check-unfinished-work` (`stats/internal/guard/unfinishedwork.go`): a ticked
column-0 task `- [x] <n>. ` whose id no `task-<ids>-implementer…` dispatch key names is OUTSTANDING, the
breakdown naming the task ids. The guard already reads the dispatch rows and the plan, integrate's gate
already runs it, and OUTSTANDING already routes to the operator through `--accept-outstanding`.
**Considered:** a `prepare` check that every `## Fix round <n>` heading has a pass row — the heading is
no contract and `flow record` has no verb that reads pass rows back; comparing commit times against the
newest panel dispatch — clock-dependent and false-positive on a Minor fixed inline after a clean close.

### Group keys cover every task they name

**ID:** group-key-coverage
**Status:** active
**Chosen:** a key `task-<n>[+<n>…]-implementer` with any suffix (`-fix-<k>`, `-r`) covers each id it
names; the row's `taskId` is not read — a group row carries one id, and the test fixtures carry none.
**Considered:** matching `taskId` — misses every task but the first of a group.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — task 3 adds a verdict signal to a gate every integrate run passes through, and
task 1 a stop in that run's `commit`.
**Considered:** `low` — under-reads a gate change whose false positive blocks every landing.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, tasks 1, 2 and 3, one fixer in plan order — tasks 1 and 3 share
`scripts/test-integrate-change.sh`, and task 2 is a two-sentence docs edit not worth a dispatch of its own.
**Considered:** one fixer per task — three dispatches, and tasks 1 and 3 serialise on their shared file anyway.

## Open questions

## Live check

None — the change edits a guard, a script and prose; no running service or store state changes. The
harnesses exercise the guard and the script.
