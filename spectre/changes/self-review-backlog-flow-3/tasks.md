# self-review-backlog-flow-3

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close the live Backlog Flow findings — KAN-609's unregistered zcode spawn type, KAN-629's
unrecorded partial stack, KAN-631's schema check constraints, KAN-633's wire-contract verification.

**Architecture:** a two-line hook change pinned by its hook test, and prose additions where their
readers already look — the harness mapping, the roster, the run-instructions stage, the principles
file and the visual-verification pair.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-backlog-flow-3-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-3`, branch `self-review-backlog-flow-3`; never touch a main checkout.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root — applied and checked in order in a scratch clone at `9637bf73` (`design.md`, **Context**). Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- `spectre/changes/self-review-backlog-flow-3/verbatim-moves.txt`, committed with this plan, acknowledges every sentence the diffs add or reword, so `scripts/check-verbatim-moves.sh` passes once each task lands.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`.

## Review Focus

- On Claude Code the hook's pinned pairs stay `flow-low`/`flow-medium` on `opus` — cases 3–7 unchanged; only a `~/.zcode/hooks` run maps to `general-purpose`.
- The out-of-worktree exception must not cover an application of a repository the change does hold a worktree in — that refusal still ends the run.
- The wire-contract line is added only when every matched path draws nothing; a diff touching one component keeps the screenshot comparison.

- [ ] 1. Dispatch ZCode on general-purpose and record a substituted spawn type (KAN-609)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'ZCODE_PAIR = ' hooks/enforce-agent-baseline.py`
    prints `64:ZCODE_PAIR = ("flow-high", "glm-5.3-flash")`. Save Step 2's diff, apply its test hunk alone
    (`git apply --include='scripts/test-enforce-agent-baseline-hook.sh' <the saved diff>`) and run
    `scripts/test-enforce-agent-baseline-hook.sh`: cases 8 and 8b fail, and the run ends `FAIL: test-enforce-agent-baseline-hook.sh (2)`.
    Restore with `git checkout -- scripts/test-enforce-agent-baseline-hook.sh`.
    <!-- measured: both commands in a scratch clone @ 9637bf73 → line 64 flow-high; FAIL 8, FAIL 8b, (2) -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9637bf73; test-enforce-agent-baseline-hook.sh passes, and fails cases 8 and 8b with the hook hunk reverted; check-references.sh, check-installed-citations.sh, check-python-suppressions.sh, check-markdown-integrity.py and check-verbatim-moves.sh 9637bf73 exit 0
diff --git a/hooks/enforce-agent-baseline.py b/hooks/enforce-agent-baseline.py
index 119ec67c..37237a81 100755
--- a/hooks/enforce-agent-baseline.py
+++ b/hooks/enforce-agent-baseline.py
@@ -59,9 +59,10 @@ PINNED_ROLES = (
 )
 
 # Harness mapping (skills/flow-contracts/model-policy.md): on zcode every pinned pair
-# is dispatched as flow-high / glm-5.3-flash. The zcode registration runs the hook from
+# is dispatched as general-purpose / glm-5.3-flash -- ZCode's Agent tool offers no
+# flow-<effort> type (KAN-609). The zcode registration runs the hook from
 # ~/.zcode/hooks, which is how the hook tells the harness.
-ZCODE_PAIR = ("flow-high", "glm-5.3-flash")
+ZCODE_PAIR = ("general-purpose", "glm-5.3-flash")
 
 # flow-high is dispatched in exactly two cases (skills/flow/implement.md): the no-progress
 # escalation -- marked by its opening line, or a self-review or pipeline-fix loop key whose
diff --git a/scripts/test-enforce-agent-baseline-hook.sh b/scripts/test-enforce-agent-baseline-hook.sh
index 73428be4..32e6764f 100755
--- a/scripts/test-enforce-agent-baseline-hook.sh
+++ b/scripts/test-enforce-agent-baseline-hook.sh
@@ -15,7 +15,9 @@
 #      file) on flow-medium / opus                         -> allow
 #   7. tooling-analyst prompt on flow-high / opus          -> deny, names flow-medium
 #   8. verifier prompt run from a ~/.zcode/hooks path on
-#      flow-high / glm-5.3-flash                           -> allow
+#      general-purpose / glm-5.3-flash                     -> allow
+#  8b. the same from ~/.zcode/hooks on flow-high /
+#      glm-5.3-flash, a type ZCode does not offer (KAN-609) -> deny, names general-purpose
 #   9. the same from ~/.zcode/hooks on flow-low / opus     -> deny, names glm-5.3-flash
 #  10. a fork dispatch carrying the verifier marker        -> allow
 #  11. malformed JSON                                      -> allow, exit 0
@@ -113,9 +115,12 @@ expect_allow "6 analyst on flow-medium/opus"
 run_hook "$HOOK" "$(payload "$ANALYST" flow-high opus)"
 expect_deny "7 analyst on flow-high" 'subagent_type: \"flow-medium\"'
 
-run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER" flow-high glm-5.3-flash)"
+run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER" general-purpose glm-5.3-flash)"
 expect_allow "8 zcode verifier on its mapped pair"
 
+run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER" flow-high glm-5.3-flash)"
+expect_deny "8b zcode verifier on a type ZCode does not offer" 'subagent_type: \"general-purpose\"'
+
 run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER" flow-low opus)"
 expect_deny "9 zcode verifier on the claude pair" 'model: \"glm-5.3-flash\"'
 
diff --git a/skills/flow-contracts/model-policy.md b/skills/flow-contracts/model-policy.md
index 869d926b..5f0601bc 100644
--- a/skills/flow-contracts/model-policy.md
+++ b/skills/flow-contracts/model-policy.md
@@ -91,7 +91,9 @@ inline row.** Where the session does not state one, `<parent model>` is `unknown
 `VERIFY_MODEL`, the literal `opus` of a dispatch with no recorded pair, a decision's implementer,
 fixer, group and panel pairs, and an operator override alike resolve and are recorded as they would be on Claude Code,
 and are replaced at the dispatch: the Agent tool's `model` parameter is `glm-5.3-flash`, the
-`subagent_type` is `flow-high` (or the site's own non-flow type, unchanged), and the dispatch's
+`subagent_type` is `general-purpose` (or the site's own non-flow type, unchanged) — ZCode's Agent tool offers no
+`flow-<effort>` type, `<agents repo>/setup.sh` installing those definitions into `~/.claude/agents/` alone, so the NO
+DELEGATION paragraph every such prompt carries is what keeps the dispatch from forking — and the dispatch's
 ledger line records `-model glm-5.3-flash -effort high` — the model the dispatch actually ran
 on, never the pre-mapping value. A reply's `Model:` line is not compared on this harness — the recorded mapping satisfies the handshake, per the paragraph below. The harness is
 the same value the run's `-harness` marks carry (**Stage marks**,
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 46e1898b..a8d2feca 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -180,6 +180,9 @@ A `default` panel, which records no pair, runs every slot on the literal `opus`
 (**Model and effort**, `skills/flow/brainstorm-planner.md`). Every slot in this table
 is dispatched on `flow-low` (`agents/flow-low.md`, on
 a `default` panel) and carries the same model rule.
+**A spawn type this harness's Agent tool does not offer is substituted on the record, never
+silently**: the slot runs on `general-purpose`, and the round records `flow record pass -round <round> -note 'roster: <slot ids> on general-purpose — <type> not offered'`.
+A harness **Harness mapping** (`skills/flow-contracts/model-policy.md`) maps to `general-purpose` makes no such substitution.
 
 **Check for an operator-named id at two points**: at the start of this stage (has the operator, in
 this run's own argument or in the session before this stage, named an id the resolved list does not
````

  - [ ] **Step 3: Verify.** `scripts/test-enforce-agent-baseline-hook.sh` prints `PASS: test-enforce-agent-baseline-hook.sh`;
    `scripts/check-python-suppressions.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `scripts/test-enforce-agent-baseline-hook.sh`, `skills/flow-contracts/model-policy.md`, `skills/flow/review-panel.md`
**Tests:** `8b zcode verifier on a type ZCode does not offer`
**Regression:** `8b zcode verifier on a type ZCode does not offer` — reverting the commit restores `ZCODE_PAIR` to `flow-high`, so the hook again admits a zcode verifier on a type ZCode never offers and refuses the `general-purpose` one that can run.
**Commit:** fix(hooks): dispatch ZCode on general-purpose and record a substituted spawn type
**After:** none

**Decision:** zcode-general-purpose
**Decision:** roster-substitution-recorded

- [ ] 2. Hand off the in-scope stack when an out-of-worktree app's port is held (KAN-629)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "outside this change's worktrees" skills/flow/verify-and-handoff.md; echo $?`: it prints only `1` —
    no rule or handoff line covers an out-of-worktree application whose start was refused.
    <!-- measured: the grep above in a scratch clone @ 9637bf73 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9637bf73 after Task 1's diff; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh 9637bf73 exit 0
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 09f9c0e9..b6721b17 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -358,7 +358,8 @@ Resolve the run instructions the start below uses, the handoff's `**Running:**`
   Exit 2 adds `Freshness: unverified — <the guard's stderr reason>` instead, the same
   visible-gap rule the visual-verification procedure's own step 6 (`skills/flow/visual-verify.md`) runs on a missing row. A refused start
   (below) already ends the run, so this check only ever runs on a start that succeeded or was
-  skipped by the protected-service rule.
+  skipped by the protected-service rule, or on the rest of a stack the out-of-worktree exception
+  below hands off.
 
 - **A refused start is relayed, never resolved.** When the start command above exits non-zero
   because a port it needs is already held, this stage does not retry, does not pick another port,
@@ -369,6 +370,15 @@ Resolve the run instructions the start below uses, the handoff's `**Running:**`
   <name>` as the next command — the same change, re-run once the holder is stopped. It never stops
   a stack this run did not start.
 
+  **An application outside the change's worktrees is the one exception.** When the refused start
+  is that of an application whose repository the change holds no worktree in — its code is not
+  under review — the rest of the stack is handed off, and the `Running:` section names it in place
+  of a URL bullet:
+
+  ```
+  Not started: <app> (<url>) — port <port> held by <the holder the start's output names>; outside this change's worktrees.
+  ```
+
 ```bash
 flow stage end -command '/flow' -stage flow.run-instructions -outcome completed <name>
 ```
@@ -473,6 +483,7 @@ flow stage end -command '/flow' -stage flow.write-in-progress -outcome completed
 - <app>: <the URL this worktree resolved> — one bullet per started application
 - Login: <app> — <user> / <password> — one bullet per `## credentials` row, when the section declares any
 - Not started: <app> (<url>) — protected, see <project>/CLAUDE.md. | Nothing to start — no runnable application; <the `## lint` and `## test` commands>
+- Not started: <app> (<url>) — port <port> held by <holder>; outside this change's worktrees. — one bullet per application **A refused start is relayed, never resolved** above hands off unstarted
 
 **Decisions:** none | <one line each:>
 - ? <an open question the operator must answer>
````

  - [ ] **Step 3: Verify.** `grep -c "outside this change's worktrees" skills/flow/verify-and-handoff.md` prints `2`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 9637bf73 → 2 -->
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — a prose rule and a handoff line; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): hand off the in-scope stack when an out-of-worktree app's port is held
**After:** none

**Decision:** out-of-worktree-refused-start

- [ ] 3. Name a schema-enforceable invariant held in code alone as a robustness violation (KAN-631)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "check constraint" skills/flow/engineering-principles.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 9637bf73 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9637bf73 after Tasks 1 and 2; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh 9637bf73 exit 0
diff --git a/skills/flow/engineering-principles.md b/skills/flow/engineering-principles.md
index 5dc37f2f..78ebb86e 100644
--- a/skills/flow/engineering-principles.md
+++ b/skills/flow/engineering-principles.md
@@ -156,7 +156,10 @@ observable.
 
 **Violation looks like:** an empty `catch` block, a swallowed exception logged at debug,
 an external call with no timeout or bounded retry, a fallback that silently returns empty
-data, or an error message that omits the identifier needed to investigate it.
+data, an error message that omits the identifier needed to investigate it, or an invariant
+the schema could enforce — an ordered pair, a non-negative amount — held in application code
+alone, with no check constraint to refuse a bad insert where it happens — a test's own setup
+included — rather than let it surface later as a wrong read or an intermittent failure.
 
 ### Principle of Least Astonishment / Surprise
 
````

  - [ ] **Step 3: Verify.** `grep -c "check constraint" skills/flow/engineering-principles.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/engineering-principles.md`
**Tests:** **none** — a principle's example list; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): name a schema-enforceable invariant held in code alone as a robustness violation
**After:** none

**Decision:** schema-check-constraints

- [ ] 4. Verify a flagged diff that draws nothing on its wire contract (KAN-633)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -rn "wire contract" skills/flow; echo $?`: it prints only `1` — neither the parent nor the
    verifier has a mode for a flagged diff that draws nothing.
    <!-- measured: the grep above in a scratch clone @ 9637bf73 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9637bf73 after Tasks 1 to 3; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh 9637bf73 exit 0
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 3123c52d..b1e6f292 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -530,4 +530,11 @@ sampling skipped that instant is no departure.
     index and what it shows — or `clean`.
 
+**The wire-contract check — when the prompt names one.** With no `wire contract:` line, the
+report's `wire contract:` line reads n/a. Otherwise, signed in as an account
+`<project>/.flow/project.md`'s `## credentials` declares, call each named endpoint on the running
+API at its worktree-resolved URL, read the response's shape and wire names against the line, then
+drive the stated acceptance behaviour through the API and read the next response — a screenshot
+comparison is no substitute for it.
+
 **No required step is left undone for want of time or context.** The report's `steps:` line gives
 each of steps 4, 7–11 and the motion step a status `check-verify-report.sh` admits — its header
@@ -568,4 +575,5 @@ or a measurement the pixels could not give is step 10 undone.
 - motion: <named motions recorded>/<named> — <each motion id: its strip directory, frame count, and `clean` or each departing frame's index and what it shows> | n/a — no motions named
   - <a strip recorded beyond the named motions>
+- wire contract: <method> <path> — <the shape and wire names read, and the acceptance behaviour seen, or each departure> | n/a — no wire contract named
 - visual-verification.md: written | not written — <reason>
 - steps: 4 <status> | 7 <status> | 8 <status> | 9 <status> | 10 <status> | 11 <status> | motion <status>
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 204fb1c3..f6549ba0 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -171,4 +171,13 @@ it is gone. A scoped re-run names the subset **The loop** (`skills/flow/verify-f
 states.
 
+**A flagged diff that draws nothing is named in the prompt, with the wire contract to check.** When
+every path step 2 matched is a web-layer file that draws nothing on screen — a DTO, a converter, a
+filter, no component, style or layout — the parent adds `wire contract:` and one `<method> <path>:
+<the payload fields and wire names the diff adds or changes>; <the change's stated acceptance
+behaviour>` line per endpoint the diff serves, read from `proposal.md`, `design.md` and the diff.
+The `ui paths` flag stands; the verification follows the diff's substance. A departure the
+verifier reports on a `wire contract` line blocks as a defect in a captured screenshot does
+(**Blocking** below).
+
 3. **Pre-flight the workspace, before anything is dispatched.** The verifier's one re-dispatch
    cannot repair an environment that cannot pass. With the parent's own Bash call, before the
````

  - [ ] **Step 3: Verify.** `grep -c "wire contract" skills/flow/visual-verify.md skills/flow/visual-verify-verifier.md`
    prints `3` and `3`; `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — a prompt rule, a verifier paragraph and a report-template line; `check-verbatim-moves.sh` covers the acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(visual-verify): verify a flagged diff that draws nothing on its wire contract
**After:** none

**Decision:** wire-contract-verification
