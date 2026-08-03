# Manual test guide — kan-23-myflow-self-review

This repository declares no runnable application (`.myflow/project.md`'s `## apps` section). Run
the checks below from the repository root in this worktree:

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/test-gather-self-review-context.sh
```

## Checks

### `scripts/gather-self-review-context.sh` — context gathering

- [x] Running the script against a change with all four sources present (ledger, panel record,
      `tasks.md`, git log) returns a bundle containing all four, exits 0.
- [x] Running it against a change missing one or more sources reports each missing source as
      `skipped: <src> (absent)` and still returns the others, exits 0.
- [x] A missing required argument, or a change name containing `/` or a glob character, exits 2 with
      a usage/error message on stderr — every other case exits 0.
- [x] A tracked symlink planted at a ledger, panel, or `tasks.md` location — at any depth, with or
      without a trailing slash, with or without a `..` component in the invocation path — is refused
      (`refused:`/`note:` line naming why) and its target's content never appears in the bundle.
- [x] The archived-change path itself being a symlink, or one of its fixed ancestor path components
      (`openspec/`, `openspec/changes/`, `openspec/changes/archive/`) being a symlink, is refused the
      same way.
- [x] Invoking the script with cwd inside a git worktree (not the main checkout) still resolves the
      trusted repository root to the main checkout.
- [x] A literal `*`, `?`, or `[...]` in the archived-change-path argument is never glob-expanded.

### `scripts/test-gather-self-review-context.sh` — the script's own test harness

- [x] Running it directly reports every case `ok:` and exits 0.

### Pipeline documentation — `/myflow-finish` run 2 gains a self-review step

- [x] `skills/myflow-contracts/pipeline.md`'s Level 1 stage table names `self-review` as the last
      step of `/myflow-finish` run 2, after `write FINISHED`.
- [x] Its Finish contract's run-2 outline states step 8 (self-review) runs after `FINISHED` is
      written, and that a skip, a failure, or a decline never moves the change off `FINISHED`.
- [x] The `## Finished` terminal block in `skills/myflow-finish/SKILL.md` carries a `Self-review:`
      line; the leftover-stop (`## Cleanup incomplete`) block does not.
- [x] `skills/myflow-finish/SKILL.md` states the skip prompt, the per-finding Jira filing prompt, and
      the rating prompt verbatim, and states the report is committed to
      `docs/self-review/<name>-self-review.md` with `push` never firing when nothing was committed.

### Bookkeeping

- [x] `.myflow/project.md`'s `## test` list includes `scripts/test-gather-self-review-context.sh`.

## Known incomplete

None. All three lint guards and the new script's full test harness pass; the review panel converged
across 8 rounds with zero open findings at any severity. Note for the operator: the self-review
*feature itself* (the skip/filing/rating prompts actually firing during a real `/myflow-finish` run
2) cannot be exercised here — that requires a full pipeline run against a genuinely finished change,
which is out of scope for this change's own manual test pass. The context-gathering script and every
documented contract edit are independently checkable above and were.
