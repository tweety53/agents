# KAN-860 — implement group (WIP: design only, no code landed)

Status: **stopped early on the coordinator's instruction (usage limits).** No script, Go guard,
test or prompt edit landed from this group. This file records the located passages and the design
already settled, so a follow-up can implement it without redoing the survey.

## Where the passages live now (audit line numbers are stale)

- MX1 review gate: `skills/flow/implement.md`, paragraph "**The review gate.** After the guard
  passes a task's commit…". It is still inline there.
- MX3 task-close call: `skills/flow/implement.md`, step 2 of "**The next implementer overlaps the
  guard.**" ("One Bash call: the implementer's `record dispatch end`, the guard on every commit…").
- MX2/OS05 throwaway snapshot: the wave-group copy has **moved to `skills/flow/sdd-dispatch.md`**
  (the "**Waves — concurrent dispatch of ready groups.**" fence, plus the `worktree remove --force`
  sentence under "**As wave members return**"). The slot copy is
  `skills/flow/review-panel-optional-slots.md` "## The throwaway worktree" (create fence, then the
  fold-back + remove fence).
- MX7 fixup fold: implement.md no longer carries its own copy. `skills/flow/gated-review-fix.md`
  and implement.md's full-suite fix both cite **Panel re-runs**
  (`skills/flow/review-panel-fix-round.md`), and review-panel.md carries no fold. **Drift #10 is
  therefore already reconciled at this base**: the one copy (review-panel-fix-round.md,
  "**Rewrite-based folding is for unpushed history only**" through "**A fixup whose fold empties
  its target commit is dropped…**") holds the union: aside wrapping, keep-both-sides conflicts, a
  refusal stops the round, the empty-fold drop, and "clean autosquash is not evidence".

## Settled design (not implemented)

All new logic is Go in flow-guard, with a `flow_guard_exec` shim, a `skills/flow/scripts/` symlink
and table tests.

1. **MX1 `check-review-gate.sh <worktree> <task-id> <sha|wt=sha,…> <canonical-worktree> <change-name> [<refused-path>…]`.**
   The plan comes from the same resolution as check-task-commit-fields. Factor that resolution
   out of `checkTaskCommitFields` into one helper so the two cannot drift. The declared set comes
   from `tcfResolveFolded` (Files ∪ Allowed-collateral, fold-widened). The facts come from the
   prose's own commands, `git diff --numstat` and `--name-only <sha>^..<sha>`, summed or unioned
   across map pairs. A binary `-` counts as 0 lines. Output is one line:
   `FIRE: task <id> — …` (more than 40 lines, and/or the undeclared paths) or
   `QUIET: task <id> — <n> changed lines, every path declared`. Exit 0 means a verdict was
   reached. Exit 2 means it could not judge. The trailing refused paths keep the rule "a
   `Files:` widening cannot disarm the gate". The 40 constant moves into Go, so the header of
   check-dispatch-paragraphs.sh must stop naming implement.md as "the one site to re-tune".
2. **MX3 `close-task.sh <canonical-worktree> <name> <merge-base> <task-id>:<sha|map> …`**
   (optional flags: `-end-key/-session-token/-end-commit` for `flow record dispatch end`, and
   `-undeclared <id>=<paths>`). It is a guard, not a `flow` verb, because it composes guards
   in-process through `guard.Registry`. The steps run in this order:
   1. check-task-commit-fields for every task, in single or map form.
   2. check-task-commit-planning-paths.
   3. If any step exited 2, the script exits 2 (stop). If any exited 1, it exits 1 (re-commit,
      with no tick and no push).
   4. The gate runs for each task.
   5. `flow tasks tick -C <canonical> <name> <id>` runs for each task the gate left QUIET.
   6. `git push origin spectre/<name>` runs for each distinct worktree. A failed push or tick
      exits 2.

   The one deliberate change of order: the push moves after the planning-paths guard, so a
   refusal never follows a push.
3. **MX2/OS05 `throwaway-worktree.sh create <worktree> <copy> [--sdd]` | `remove <worktree> <copy> [--fold-back]`.**
   It ports the fences byte for byte, including the `status --porcelain -z` token loop (the
   rename-origin token quirk as well), the `.superpowers/sdd` copy and the `-nt` fold-back.
   Both callers use it.
4. **MX7 `fold-fixup.sh <worktree> <task-sha> <tasks-md> <path>…` and `fold-fixup.sh --finish <worktree> <task-sha> <tasks-md>`.**
   The steps run in this order:
   1. `git add --`
   2. `guard-autosquash targets`
   3. `commit --fixup -- paths`
   4. aside
   5. `GIT_SEQUENCE_EDITOR=: git rebase -i --autosquash <sha>^`. This needs `-i`: git 2.43, the
      version in this container, ignores a non-interactive `--autosquash`.
   6. restore
   7. the empty-fold drop: `reset --hard <parent>` when the emptied commit is the tip, otherwise
      `rebase --onto`
   8. `guard-autosquash after <sha>^ <tasks-md>`

   Exit codes: 0 folded or dropped; 1 a guard-autosquash refusal (the round stops); 2 could not
   answer; 3 a conflict is left in progress (resolve by hand keeping both sides,
   `git rebase --continue`, then `--finish`).

Rows left: all of MX1, MX2/OS05, MX3 and MX7 (not started, for lack of budget), plus MX5
(low confidence, per instruction). No verbatim-moves lines were added.
