# kan-798-flow-fix-protect-main-checkout-py-false — session narrative

## 2026-09-28 — creating run

- Reachability check at checklist open: mode 1 of the filed finding (unexpanded `$VAR` redirect
  denied) no longer reproduced at the base — `resolve()` already let `$`/backtick paths through —
  so the run planned on the two live modes plus the issue's prescribed expansion as an improvement
  over the let-through.
- The shlex-whitespace mechanism for mode 3 was spiked at plan time and failed (`\n` glues into
  words), which moved the design to the logical-line pre-split; recorded as a decision-considered
  line rather than discovered mid-task.
- The base moved 5 commits while the panel stage opened; the no-overlap automatic rebase refused
  once (tasks.md carried uncommitted tick edits) — resolved by making the reviewer-dispatch
  planning commit first, then rebasing clean and force-with-lease pushing.
- check-task-commit-fields refused task 1's commit: the plan's `**Files:**` fields were written
  comma-separated but the guard parses backtick-quoted tokens; fields rewritten (the record
  carries its own corrections) and the guard re-run green.
- The store rejected an inline dispatch row recorded at effort `unknown` and then `default` — the
  daemon enforces the zcode mapping (`glm-5.3-flash`/`high` only); all inline rows recorded at
  that pair.
- Panel pass 0 caught a real regression the plan's own narrow rule introduced (the `.worktrees`
  exemption also exempted the root, so `rm -rf <main>/.worktrees` sailed through) — the fix
  round's `!= d` arm closed it, mutation-proved by case 35.
- The fix round initially fixed F6 (two-site convention) with only a cross-reference comment; the
  slot's reproducer demands single-siting the literal. Withdrawn with reason instead of chasing
  the route: a personal hook reading a repo lib path across arbitrary protected repos trades a
  style duplication for a deployment fragility.
- check-panel-fix-single-dispatch flagged the gated per-task reviewer's fix key
  (`task-1+2-implementer-fix-1`) as out-of-shape at panel close — that key is implement.md's own
  prescription for the sdd-stage fix round; auto-resolved Continue, recorded.
- Store hiccup: one stage-end write journalled mid-run (`store unreachable` warning, exit 0);
  per the state contract the journal replays.

## 2026-09-28 — integrate run

- Preflight RUN1; unfinished-work gate CLEAR; visual-verify OK (no UI paths).
- The base had moved 39 commits since the panel's rebase, overlapping KNOWN-BUGS.md — the sync
  rebase stopped on that file, resolved in place (upstream's five new deferred entries kept, this
  change's entry appended after), and the full lint+test lists ran green afterwards, per the
  resolution-requiring-rebase rule.
- Landing route taken from the project's configured default (merge and push), not asked.
