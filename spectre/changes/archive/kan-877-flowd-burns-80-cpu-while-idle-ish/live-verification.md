# kan-877-flowd-burns-80-cpu-while-idle-ish — live verification

The `## Live check` in `design.md` needs the dev flowd restarted on the new build, which no agent may do
(`CLAUDE.md`). In its place the run drove the real `Watcher` over this machine's real transcript root,
2026-10-07, at the branch tip (2c38a406):

- Setup: a throwaway in-package test (deleted after the run, never committed) built a `Watcher` over
  `~/.claude/projects` (2,246 `.jsonl` files), a fake sink holding every file's committed offset at its
  current size, and 104 pending retried give-up tokens matching nothing — the shape `## Context` measured.
- Before (from `design.md` `## Context`): ~90 s CPU per 5 s cycle for as long as a retried token stays
  pending; ~0.17 s per idle cycle.
- After: cycle 1 30.3 s wall (the one retried-token scan pass), cycles 2–4 22 ms, 20 ms, 21 ms wall.
  <!-- measured: LIVE_ROOT=$HOME/.claude/projects go test ./internal/harvest -run TestZZLiveMeasure -v, 2026-10-07 — machine-local -->
- Failure look (`design.md` step 4): cycles past the first costing seconds each. Not seen: matched.

Still for the operator after landing: restart flowd and sample `ps -o time=` per the `## Live check` steps.
