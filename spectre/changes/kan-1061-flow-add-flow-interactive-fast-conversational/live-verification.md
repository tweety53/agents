# Live verification — kan-1061-flow-add-flow-interactive-fast-conversational

Scratch session against `scripts/live-store-copy.sh up` (a copy of the dev `flow` database), the `flow` CLI built from the checkout under test, in a scratch repository with a bare origin whose `.flow/project.md` declares `merge and push` and `ui paths` `web/**`. Harness: the parent's scratch `live/run.sh`, steps 1–5 of design.md's `## Live check`.

## Before — merge base 35f477f8d6f697deb4913c2409acddef40979ab7 (detached worktree, copy on port 4176)

- Kickoff and stop marks: refused — `"flow.write-finished" names no documented stage key for /flow-interactive`.
- `flow record request …`: exit 2 (no such verb); requests endpoint absent.
- Request 3's end without a UI check: not guarded (verb absent, exit 2).
- Stage runs carrying `request`: 0. Change state after the stop mark: none (never created).
- Failure look matched on every figure — this is the before.

## After — HEAD 9825b70e (change worktree, copy on port 4175)

- Kickoff mark accepted; `flow.load-context` run with `-request 1` accepted; request 1 `answered`, exit 0.
- Request 2 (docs-only commit pushed to the base): `request end -outcome landed`, exit 0.
- Request 3 (`web/x.ts`, pushed): no UI check → exit 1 (`touched ui paths and has no UI check`); latest check `failed` → exit 1; a later `completed` check → exit 0.
- Requests read back: `[{n:1,answered},{n:2,landed,base set},{n:3,landed,base set}]` — `base_sha` set on both landed rows.
- Stage runs carrying `request`: 3. State after `flow.write-finished`: `FINISHED`.
- No failure look: no mark refused, request 3's end refused until a passing UI check, the change `FINISHED` after the stop.
- Board pill and card: covered by Task 9's captures and `flow.visual-verify`, not re-checked here — the copy's board drops a `FINISHED` change.
