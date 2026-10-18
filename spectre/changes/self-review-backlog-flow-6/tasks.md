# self-review-backlog-flow-6

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close the one live finding of the next five Backlog Flow tickets — KAN-702's
hand-picked satellite merge base in the cross-repo visual trigger.

**Architecture:** one prose sentence in the visual-verification step that already runs the match.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-backlog-flow-6-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Task 1 runs in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6`, branch `spectre/self-review-backlog-flow-6`; never touch a main checkout.
- The task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root, checked in a scratch clone at `cfb9b564`. Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. The sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- The sentence binds `<merge-base>` to the working notes, the source **Inline verify** in the same file already names, and never to the state file's `worktrees` map.

- [ ] 1. Bind the cross-repo visual trigger's merge base to the working notes (KAN-702)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "own merge base from this run's working notes" skills/flow/verify-and-handoff.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ cfb9b564 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at cfb9b564; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh cfb9b564 fails only on the sentence verbatim-moves.txt lists
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index b6721b17..532f4b43 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -218,6 +218,10 @@ this pipeline restates it. Resolve once per worktree in this run's resolved set,
    Its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
    does.
 
+   `<merge-base>` in both calls is that worktree's own merge base from this run's working notes —
+   never `origin/<base>` or another ref picked by hand, which matches every path the base gained
+   since the worktree was cut as this change's own (KAN-702).
+
    **On a fix run, both calls read the fix run's own diff.** A worktree that has a since-close sha this run — the one **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`) read for it — feeds them `git -C <worktree> diff --name-only <since-close-sha>..HEAD` in place of `<merge-base>..HEAD`, so a fix run whose delta touches no `ui paths` skips this stage at step 2 as exit 1 does.
    A worktree whose since-close sha is `-` keeps `<merge-base>..HEAD`, and so does every worktree when the change's newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify` — the rows `<agents repo>/scripts/check-visual-verify-dispatched.sh` counts) does not have `outcome` `completed` — an earlier run's UI change that was never verified may lie before the since-close sha.
 
````

  - [ ] **Step 3: Verify.** `grep -c "own merge base from this run's working notes" skills/flow/verify-and-handoff.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ cfb9b564 → 1 -->
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — a prose sentence; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): bind the cross-repo visual trigger's merge base to the working notes
**After:** none

**Decision:** visual-trigger-working-notes-base
