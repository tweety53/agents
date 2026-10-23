## Context

- kan-959 added the always-on `test-sanity+call-efficiency` dispatch: `pcExtra` in `stats/internal/guard/planclass.go`, `extra_dispatch` in `decision.json` (render in `stats/cmd/flow/decision.go`, model check in `stats/internal/api/records.go`, slot membership in `stats/internal/store/aggregate.go`), and **The always-on extra dispatch** in `skills/flow/review-panel.md`. This change adds a second bundle on the same rails.
- `mutation`'s pass is never rendered — its brief and throwaway copy stay parent-typed wherever it is bundled (`renderslotprompt.go`), so bundling it beside `failure-modes` needs no guard change.

## Decisions

### Second bundle, its own roll and its own line

**ID:** mf-own-roll
**Status:** active
**Chosen:** a second bundle `mutation+failure-modes`, its pair from `roll(<name>extramf)` — `sonnet` below 50, else `opus`, always `low` — printed as a second `extra:` line after the first; plan-class prints nine lines. Operator's answer (2026-10-23).
**Considered:** sharing the `test-sanity+call-efficiency` pair — rejected by the operator; one line carrying both bundles — rejected, the parser and the decision record read one bundle per line.

### Recorded as `extra_dispatch_mutation`

**ID:** extra-dispatch-mutation-key
**Status:** active
**Chosen:** a new decision key `extra_dispatch_mutation` `{slots, model, effort, reason}` beside `extra_dispatch`, optional like it — every decision recorded before this change stays valid with no compatibility code.
**Considered:** turning `extra_dispatch` into an array — rejected, it breaks every recorded decision and the `$.extra_dispatch.slots[*]` query, and needs a dual reader.

### failure-modes and mutation leave the roster

**ID:** mf-out-of-roster
**Status:** active
**Chosen:** the Decide tree's roster is `primary; principles` on every class and shape; a `default` panel's resolved roster (settings store) drops any role an always-on extra dispatch carries, recorded as a not-dispatched pass. Operator's answer (2026-10-23).
**Considered:** keeping both in the full roster — rejected by the operator, the role would run twice on a full panel.

### Experimental slot keeps its gate, joins the first dispatch

**ID:** exp-slot-first-dispatch
**Status:** active
**Chosen:** an experimental slot is still admitted only on a `full`-shape `regular`/`big` panel (`room` unchanged), and now rides the first dispatch, `primary+principles+exp-<name>`; elsewhere it stays `skipped — bundle cap`.
**Considered:** admitting it on every non-micro panel, since the first dispatch now has room — rejected as an unrequested change to how often experimental prompts run.

### Bundle name order

**ID:** mf-bundle-order
**Status:** active
**Chosen:** `mutation+failure-modes`, the operator's own spelling, everywhere: plan-class output, `extra_dispatch_mutation.slots`, `flow record dispatch -slot`, `render-slot-prompt.sh`'s slot argument.
**Considered:** `failure-modes+mutation` (old roster order, mutating role last) — rejected, the bundle is its own dispatch so the overflow ordering it served no longer applies.

### Learning note is a file beside the repository

**ID:** learning-note-file
**Status:** active
**Chosen:** the note is written to `<project>-learning/<name>.md`, `<project>` the main checkout's absolute path, the folder created on first use and never inside a git repository; its absolute path is printed the moment the note is composed (`flow.load-context` on `/flow`, the same stage on `/flow-fast`), and the handoff's `**Learning:**` line repeats the path in place of the `## Learning` section. Recorded with `flow learning add` when the file is written. Operator's request (2026-10-23).
**Considered:** keeping the section inline in the handoff — rejected by the operator, it arrives only at the end of the run; writing it under the worktree or `flow state dir` — rejected, the operator asked for a sibling folder like `<project>-worktrees`, outside git.

## Open questions

## Live check

None — the change alters a guard's printed lines, a decision field and skill text; `flow.verify` runs `plan-class.sh` against this change's own `tasks.md` and `flow decision render` against its own `decision.json`, which exercise every runtime path the change touches.
