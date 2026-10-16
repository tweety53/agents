# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-16

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** the explain-then-ask turn split applies only to `/flow-self-review`'s flow-eval
candidate mining ask; every other operator ask explains in the message body and asks in the same
turn, as before b8e0e063.

**Architecture:** prose only — the rule moves from `skills/flow-contracts/operator-prompts.md` into
`skills/flow-self-review/SKILL.md` step 4; `skills/flow/unfinished-work-gate.md`'s citation goes.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/` — this change's `verbatim-moves.txt` is already
  committed with the plan.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`. A diff that does not apply is reported, never hand-merged by guesswork.

## Review Focus

- The restored sentences match `git show bb273948`'s pre-image word for word.
- No run-loaded file cites **Explain before asking** after the change.
- Step 4 still states the angle-7 split, the same-turn case with no candidate, both rejected
  alternatives and the **Unanswered mid-run asks** citation.

- [x] 1. Scope the explain-then-ask turn split to self-review's mining ask
  - [x] **Step 1: Apply.** Save the diff below to `/tmp/pf16-task1.diff` and run
    `git apply /tmp/pf16-task1.diff` from the worktree root:

````diff verified:applied with git apply to this worktree at a618d04f; check-verbatim-moves 0 violations with this change's verbatim-moves.txt, check-references and check-ask-silence OK
diff --git a/skills/flow-contracts/operator-prompts.md b/skills/flow-contracts/operator-prompts.md
index 533fe4f3..edb84d2f 100644
--- a/skills/flow-contracts/operator-prompts.md
+++ b/skills/flow-contracts/operator-prompts.md
@@ -13,28 +13,6 @@ A prompt in this shape states:
 - a 🤖 marker in the handoff when that silent default actually fired, or when **Auto-resolution**
   below took it
 
-## Explain before asking
-
-**A flow-eval candidate's explanation is written in chat as the turn's final message, and the turn
-ends there.** That covers each flow-eval candidate in its kind's template exactly as step 2 of
-`skills/flow-self-review/SKILL.md` defines it (impl, review or visual, every field). The
-`AskUserQuestion` that offers the candidates fires only in a later turn, after the operator has seen
-the explanation; the operator may instead answer in chat directly.
-
-**Every other ask explains and asks in the same turn.** A finding, a filing item, a gate or any other
-question the operator must read up on first is explained in chat — observed, what breaks, the fix —
-and the `AskUserQuestion` follows in that same turn; the turn is never ended just to split the two.
-
-Either way, the explanation is never carried only inside the `AskUserQuestion`'s question, option or
-preview text. Option text records the decision only.
-
-Why: a harness focus mode shows only a turn's final message, so a flow-eval candidate's template
-written before a call in the same turn is never seen. Carrying the explanation inside the prompt
-instead was rejected by the operator on 2026-10-16: the explanation belongs in chat, in the agreed format, not in the prompt. Splitting every ask across two turns was
-rejected by the operator the same day: the split is for flow-eval candidates alone.
-An unanswered call of this shape follows **Unanswered mid-run asks**
-(`skills/flow-contracts/pipeline.md`).
-
 ## Every item offered
 
 **A prompt that asks the operator to choose from a list — findings to file, items to act on —
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index f65474a1..02799b36 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -274,13 +274,24 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
 
 ### 4. Explain, then ask
 
-Every offered finding is explained before the operator answers — what was observed, what breaks,
-and what the fix would be — in chat ahead of the ask, per **Explain before asking**
-(`skills/flow-contracts/operator-prompts.md`): a filed issue is durable, and an explanation
-arriving afterward describes something the operator did not agree to. When the call offers an
-angle-7 candidate, the whole explanation is the turn's final message and the call fires in the next
-turn; with none, the call follows the explanation in the same turn. Every fixed finding is one
-line in the explanation message, naming its landed sha; it is not offered. The filing ask,
+Every finding is explained in the message body first, before any prompt fires — what was
+observed, what breaks, and what the fix would be. Every fixed finding is one line in the same
+body, naming its landed sha; it is not offered. A prompt's option text cannot carry that
+explanation, so the prompt records the decision only: a filed issue is durable, and an
+explanation arriving afterward describes something the operator did not agree to.
+
+When the call offers an angle-7 candidate, the whole explanation is the turn's final message and
+the call fires in the next turn; with none, the call follows the explanation in the same turn. The
+`AskUserQuestion` that offers the candidates fires only in a later turn, after the operator has seen
+the explanation; the operator may instead answer in chat directly. Why: a harness focus mode shows
+only a turn's final message, so a flow-eval candidate's template written before a call in the same
+turn is never seen. Carrying the explanation inside the prompt instead was rejected by the operator
+on 2026-10-16: the explanation belongs in chat, in the agreed format, not in the prompt. Splitting
+every ask across two turns was rejected by the operator the same day: the split is for flow-eval
+candidates alone. An unanswered call of this shape follows **Unanswered mid-run asks**
+(`skills/flow-contracts/pipeline.md`).
+
+The filing ask,
 the mining ask and the rating are **one `AskUserQuestion` call**, shape per **Operator prompts**
 (`skills/flow-contracts/operator-prompts.md`): up to three multi-select questions of
 three offered items each — findings to file or angle-7 candidates to mine, never both in one
diff --git a/skills/flow/unfinished-work-gate.md b/skills/flow/unfinished-work-gate.md
index 9a68c6e4..2949a179 100644
--- a/skills/flow/unfinished-work-gate.md
+++ b/skills/flow/unfinished-work-gate.md
@@ -23,10 +23,11 @@ for the guards.
 
 **The `OUTSTANDING:` prompt above is a filing ask** — its third course files exactly these items,
 per **Follow-up issues** (`skills/flow-contracts/jira-followups.md`).
-**Every filing ask explains before it asks.** The filing prompt explains each item the run would
-file — what was observed, what breaks because of it, and what the fix would be — in chat, ahead of the
-ask in the same turn, per **Explain before asking** (`skills/flow-contracts/operator-prompts.md`), and
-otherwise follows the shape that contract defines. A filed issue is durable and already on the board; an explanation reaching the
+**Every filing ask explains before it asks.** Before the filing prompt fires, the message body
+explains each item the run would file — what was observed, what breaks because of it, and what the
+fix would be — never leaving that explanation to the prompt's option text. The prompt itself follows
+the shape **Operator prompts** (`skills/flow-contracts/operator-prompts.md`) defines and records
+only the decision. A filed issue is durable and already on the board; an explanation reaching the
 operator afterward would describe something they never agreed to.
 
 ## The three courses
````

  - [x] **Step 2: Verify.** `grep -rln "Explain before asking" --exclude-dir=.git --exclude-dir=archive skills rules commands-claude`
    prints nothing; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-ask-silence.sh .` and `scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 3: Commit** the three files with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/operator-prompts.md`, `skills/flow-self-review/SKILL.md`, `skills/flow/unfinished-work-gate.md`, `spectre/changes/pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-16/verbatim-moves.txt`
**Tests:** **none** — prose contract; guarded by check-verbatim-moves, check-references and check-ask-silence.
**Regression:** none
**Commit:** fix(flow-contracts): scope the explain-then-ask turn split to self-review's mining ask
**After:** none

**Decision:** rule-lives-in-step-4
**Decision:** restore-original-text
**Decision:** dedupe-moved-sentences
