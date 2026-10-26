# self-review-kan-1035-flow-self-review-analyse-visual-verify-prompts

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close the three non-`big` findings of the kan-1035 self-review pass — visual-verify cases
with no severity or slot, verifier cases `/flow-prompt-tuning` cannot replay, and `/flow-fast`
running its lint list's fence lines.

**Architecture:** prose only — one sentence in angle 8, one sentence in step 2a and one replay
bullet in `/flow-prompt-tuning` step 4, and `/flow-fast` section 5 pointing at the existing
`scripts/run-verify-list.sh`.

**Spec:** `design.md` beside this file.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1035-flow-self-review-analyse-visual-verify-prompts`, branch `kan-1035-flow-self-review-analyse-visual-verify-prompts`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its step gives, applied with `git apply` from the worktree's root, in task order. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.

## Review Focus

- Task 2: the replay bullet sends a verifier-target case to a verifier dispatch over `capture.png` and `mockup.png`, and step 2a names both files.
- Task 3: `run-verify-list.sh` runs only the fenced lint commands; the `## test` scoping sentence is unchanged.

- [ ] 1. Judge a visual-verify departure as an Important in angle 8
  - [ ] **Step 1: Apply the fix:**

````diff verified:applied with git apply in this worktree at f9428445, Tasks 1-3 in order; check-verbatim-moves.sh then printed exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 05a09a78..97ed1c4f 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -94,7 +94,9 @@ the bundle names.
 Angle 8 reads the run's review-panel findings and its visual-verify findings, each judged alone on its own severity — no severity
 needs another alongside it. A Critical finding qualifies; an Important finding qualifies; a Minor
 finding qualifies only when the pass judges it genuinely useful, meaning it is in a class the
-prompts should handle and would recur. A finding is a case only when its root cause is a generic
+prompts should handle and would recur. The verifier grades no severity: a visual-verify departure blocks the handoff
+(**Blocking**, `skills/flow/visual-verify.md`), so it is judged as an Important, and its case records
+`source.slot` `visual-verify` and `source.severity` `Important`. A finding is a case only when its root cause is a generic
 gap in instructions:
 
 - **`implementer`** — a rule in the implementer instructions (`skills/flow/implement.md`,
````

  - [ ] **Step 2: Verify:** run `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh`; the sentence this task adds is reported `ok acknowledged new text`, and the two FAIL lines left are Tasks 2 and 3's until they land.

**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** none — prose; `check-verbatim-moves.sh` checks the sentence.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `fix(flow-self-review): grade a visual-verify departure as Important`

- [ ] 2. Give a verifier case its capture and mockup, and replay it with a verifier
  - [ ] **Step 1: Apply the fix:**

````diff verified:applied with git apply in this worktree at f9428445, Tasks 1-3 in order; check-verbatim-moves.sh then printed exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 05a09a78..97ed1c4f 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -196,7 +198,10 @@ worktree — `<agents repo>-worktrees/<name>/docs/prompt-tuning/<case-id>/` on t
 | `index.html` | rendered by `flow prompt-tuning render <case-dir>` from the files above, never written by hand |
 
 `bad.diff`, `good.diff` and `task.md` are extracted with `git -C <repo> show` exactly as angle 7's
-extraction does. Once a case's files are written, run `flow prompt-tuning render <case-dir>`, then
+extraction does. A `reviewer` case whose `target` is `skills/flow/visual-verify-verifier.md` or
+`skills/flow/visual-verify.md` also carries `capture.png`, the capture the verifier judged, and
+`mockup.png`, the mockup frame it judged that capture against, both copied into the case folder.
+Once a case's files are written, run `flow prompt-tuning render <case-dir>`, then
 `<agents repo>/scripts/check-prompt-tuning-cases.sh` over the new folders. A red line is fixed by
 rewriting the field it names; a case no generic rewrite passes is dropped, and its report line says
 so, disposition `noted`.
diff --git a/skills/flow-prompt-tuning/SKILL.md b/skills/flow-prompt-tuning/SKILL.md
index c7dd9845..65e2268a 100644
--- a/skills/flow-prompt-tuning/SKILL.md
+++ b/skills/flow-prompt-tuning/SKILL.md
@@ -70,6 +70,9 @@ text with the update applied.
 
 - **`reviewer` case** — a reviewer dispatch over `bad.diff` (a miss) or `good.diff` plus the
   finding's location (a false positive), with the target prompt as its instructions.
+- **verifier case** — a `reviewer` case whose `target` is `skills/flow/visual-verify-verifier.md`
+  or `skills/flow/visual-verify.md` is replayed instead by a verifier dispatch judging the case's
+  `capture.png` against its `mockup.png`, with the target prompt as its instructions.
 - **`implementer` case** — an implementer dispatch given `task.md` and the pre-change code,
   writing in a scratch directory under the session's scratchpad, never in a repository.
 - **Grader** — a grader dispatch judges each run against the case's `finding.md`: was the defect
````

  - [ ] **Step 2: Verify:** run `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh`; this task's two sentences are reported `ok acknowledged new text`.

**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-prompt-tuning/SKILL.md`
**Tests:** none — prose; `check-verbatim-moves.sh` checks the sentences.
**Regression:** none — no test is declared.
**Build:** green
**After:** Task 1
**Commit:** `fix(flow-prompt-tuning): replay a verifier case over its capture and mockup`

- [ ] 3. Run /flow-fast's lint list through run-verify-list.sh
  - [ ] **Step 1: Apply the fix:**

````diff verified:applied with git apply in this worktree at f9428445, Tasks 1-3 in order; check-verbatim-moves.sh then printed exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
index 3aafd793..f35af6df 100644
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -289,4 +289,5 @@
-Run every command in `## lint`, in the worktree, and the `## test` commands scoped to what the
+Run `<agents repo>/scripts/run-verify-list.sh <worktree> lint` — it runs every command of
+`## lint`'s fence, never the fence lines `project-get.sh` prints with them — and the `## test` commands scoped to what the
 change touched — the packages, modules or test files the diff names, never the full suite unless
 the operator asked for it. When the diff names nothing those commands can scope to, none of them
 runs — the change summary below states that empty-scope reason explicitly, the stated reason
````

  - [ ] **Step 2: Verify:** run `scripts/check-verbatim-moves.sh` (exit 0) and `scripts/check-references.sh` (exit 0), then the absence sweep `git grep -n -F -e 'Run every command in' -- skills/flow-fast/SKILL.md`, which prints nothing.

**Files:** `skills/flow-fast/SKILL.md`
**Tests:** none — prose; `run-verify-list.sh` is already covered by `stats/internal/guard/run_verify_list_test.go`.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `fix(flow-fast): run the lint list through run-verify-list.sh`
