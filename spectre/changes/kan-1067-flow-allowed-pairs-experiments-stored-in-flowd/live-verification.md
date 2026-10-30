# Live verification — kan-1067

On `scripts/live-store-copy.sh up kan1067` (copy of the dev `flow` database as `flow_kan1067_live`, branch flowd on :4175, branch `flow` binary). The name drops the hyphen: the script refuses `kan-1067`.

1. `flow experiment open -name t -cell 'planner=opus/flow-planner-xhigh' -cell 'plan-designer=opus/Plan@xhigh' -cell 'explorer=sonnet/Explore@low'` → `experiment 1 open`, exit 0. A second open naming `explorer` → exit 1, `role "explorer" is held by open experiment 1`. Matched.
2. `scripts/plan-agent-pair.sh x 1`: before figure, measured from a detached worktree at merge base 52c67c31 → `opus low`; on the branch with no experiment open → `opus low`; with experiment 1 open → `opus xhigh`; after close → `opus low`. Matched.
3. check-dispatch-pair: `Role: explorer` on `sonnet/Explore@low` → `DISPATCH-PAIR-OK explorer`; `Role: planner` on `opus/flow-planner-xhigh` → `DISPATCH-PAIR-OK planner`; the same planner dispatch with `FLOW_ADDR`/`FLOW_RECORDS_ADDR` on a dead port → `DISPATCH-PAIR-REFUSED` (table fallback). Matched.
4. `flow experiment stats 1` right after opening: runs before 0, after 0. The before window has the after window's length by design (Task 6), so a just-opened experiment has an empty before window. Proxy on the scratch copy only: `started_at` moved 14 days back → runs before 84, after 76; before means reviewRounds 1.345, round0CritImp 1.226 (non-zero, from real history). tokens and wallClockMin before are null: the copied history in that window carries no token or wall-clock metrics. planCorrections is null in both windows, since no older change recorded corrections. Matched on the proxy.
5. `flow experiment close 1 -verdict negative -summary s` → `experiment 1 closed negative`; `flow experiment overrides` → `{}`. Matched.
6. The Experiments view's data at the copy's port: `GET /api/v1/roles` lists explorer, plan-designer and planner each with an `override` naming experiment 1 beside the table pairs (what the view strikes through). The rendering itself is checked by visual verification against the prototype.

Failure look (same pair before and after; all-zero before figures): not observed.
