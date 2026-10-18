# self-review-backlog-flow-4

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close the live Backlog Flow findings — KAN-691's unreachable error-path fix, KAN-698's
prerequisite measured off the worktree, KAN-652's tautological calendar-marks colour check.

**Architecture:** two prose additions where their readers already look — the testing principles
and the plan-provenance contract — and one test-only change in gymie-frontend.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-backlog-flow-4-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1 and 2 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-4`; Task 3 runs in gymie-frontend's `/Users/tweety53/Projects/gymie-frontend-worktrees/self-review-backlog-flow-4`, which its Step 1 creates. Both are on branch `self-review-backlog-flow-4`; never touch a main checkout.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from its worktree's root — Tasks 1 and 2 checked in order in a scratch clone at `54033573`, Task 3 in a scratch gymie-frontend worktree at `5ce0d237` (`design.md`, **Context**). Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. A prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- The KAN-691 paragraph stays a testing principle: it asks for the real trigger and an interceptor trace, and names no project's client.
- The KAN-698 check runs in the change's own worktree, never against `origin/<base>`; a stale base is reported, never rebased by the implementer.
- Task 3's colour read must fail on a swapped and on an inverted mark and pass on the unchanged code; the tag-literal check is gone.

- [x] 1. Require an error-path fix's test to drive its real trigger (KAN-691)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -i "interceptor" skills/flow/engineering-principles.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 54033573 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 54033573; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh 54033573 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/engineering-principles.md b/skills/flow/engineering-principles.md
index 78ebb86e..7773e03c 100644
--- a/skills/flow/engineering-principles.md
+++ b/skills/flow/engineering-principles.md
@@ -192,6 +192,13 @@ on a measured or reported bound (Compose's semantics bounds, coerced to the pare
 the property under test is what is drawn — an overflow is visible only in the rendered
 pixels, so the assertion passes on the unfixed code.
 
+**An error-path fix is tested against its real trigger.** A test that hands the fixed handler a
+payload the real backend never sends on that path — a bare `403` where every real one carries an
+error code — proves nothing about reachability: a global interceptor, a client-wide handler for
+that status or code, can take the real response first and leave the local handler dead while its
+test stays green (KAN-691). Before the test is written, trace the real response through every
+interceptor to the handler, and drive the test or reproducer with that real response, end to end.
+
 ### Twelve-Factor App
 
 Config comes from the environment, backing services are attached resources, processes are
````

  - [x] **Step 3: Verify.** `grep -c "interceptor" skills/flow/engineering-principles.md` prints `2`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 54033573 → 2 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/engineering-principles.md`
**Tests:** **none** — a principle paragraph; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): require an error-path fix's test to drive its real trigger
**After:** none

**Decision:** error-path-real-trigger

- [x] 2. Measure a plan's prerequisite as an ancestor of the worktree's HEAD (KAN-698)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "is-ancestor" skills/flow-contracts/plan-provenance.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 54033573 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 54033573 after Task 1's diff; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh 54033573 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow-contracts/plan-provenance.md b/skills/flow-contracts/plan-provenance.md
index 20f8d2e9..e3f668ed 100644
--- a/skills/flow-contracts/plan-provenance.md
+++ b/skills/flow-contracts/plan-provenance.md
@@ -73,3 +73,15 @@ URL, or the output that shows its check ran — or it is not written at all, the
 being the `unverified:`/`predicted:` tag. An evidence-free verification tag — a bare `verified:`
 or `measured:` naming nothing after the colon — is a violation, the same way an `unverified:` with
 no `<what-to-check>` is, and both are reported as such.
+
+## A prerequisite is measured on the worktree's own HEAD
+
+A plan that rests on another change being present — "KAN-361 is merged before task 7" — measures
+it in the worktree the change is implemented in, never against a remote branch:
+`git -C <worktree> merge-base --is-ancestor <prerequisite sha> HEAD`, its `measured:` comment
+naming that worktree's HEAD as the ref. A remote branch carrying the prerequisite says nothing
+about a worktree based before it (KAN-698: confirmed against `origin/develop` while the backend
+worktree's HEAD lacked KAN-361, so the feature was silently absent on the stack). The first step of
+every task that depends on the prerequisite runs the same check in its own worktree; an exit 1 is a
+stale base the implementer reports and never builds past — rebasing the worktree is the
+operator's call.
````

  - [x] **Step 3: Verify.** `grep -c "is-ancestor" skills/flow-contracts/plan-provenance.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 54033573 → 1 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/plan-provenance.md`
**Tests:** **none** — a contract section; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow-contracts): measure a plan's prerequisite as an ancestor of the worktree's HEAD
**After:** none

**Decision:** prerequisite-worktree-ancestor

- [x] 3. Assert each calendar mark's colour beneath the disc in gymie-frontend (KAN-652)
  - [x] **Step 1: Create the worktree, then reproduce before touching anything; do not guess.**
    In `/Users/tweety53/Projects/gymie-frontend` run `git fetch origin`, then
    `git worktree add -b self-review-backlog-flow-4 /Users/tweety53/Projects/gymie-frontend-worktrees/self-review-backlog-flow-4 origin/develop`;
    in the new worktree run the `## worktree setup` line of its `.flow/project.md` (it copies the
    gitignored `local.properties`). Then
    `grep -n 'tag.contains("inverted")' shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt`
    prints line `320`, the check on the test's own tag literal, and
    `grep -c 'colors.accent2' shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt` prints `0`.
    <!-- measured: both greps @ gymie-frontend 5ce0d237 → 320, 0 -->
  - [x] **Step 2: Apply the fix** in that worktree (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch gymie-frontend worktree at 5ce0d237; ./gradlew ktlintFormat, then ./gradlew verifyChange :shared:desktopTest --tests 'com.gymie.shared.ui.HomeCalendarMarksUiTest' exit 0; the test fails with HomeCalendarGrid.kt's borrowed mark drawn in colors.accent and with its own mark drawn in colors.neutral100
diff --git a/shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt b/shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt
index c3cbba6b..81670cae 100644
--- a/shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt
+++ b/shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt
@@ -1,5 +1,6 @@
 package com.gymie.shared.ui
 
+import androidx.compose.ui.graphics.Color
 import androidx.compose.ui.graphics.toPixelMap
 import androidx.compose.ui.test.ComposeUiTest
 import androidx.compose.ui.test.ExperimentalTestApi
@@ -40,7 +41,6 @@ import kotlinx.serialization.json.Json
 import kotlin.math.abs
 import kotlin.test.Test
 import kotlin.test.assertEquals
-import kotlin.test.assertFalse
 import kotlin.test.assertTrue
 
 /**
@@ -305,19 +305,30 @@ class HomeCalendarMarksUiTest {
     }
 
     /**
-     * Asserts [tag]'s node exists, carries no `_inverted` infix, and sits inside [dayTag]'s cell, below
-     * [wrapTag]'s 32dp disc/ring centred on the 32dp numeral box — exactly [gap] below it when one
-     * is given.
+     * Asserts [tag]'s node exists, draws in [color] — its ink lies nearest [color], so a mark inverted to the
+     * disc's ink or drawn in another kind's colour fails — and sits inside
+     * [dayTag]'s cell, below [wrapTag]'s 32dp disc/ring centred on the 32dp numeral box — exactly [gap]
+     * below it when one is given.
      */
     private fun ComposeUiTest.assertMarkBeneathDisc(
         tag: String,
+        color: Color,
         wrapTag: String,
         dayTag: String,
         gap: Dp? = null,
     ) {
         val disc = 32.dp
-        onNodeWithTag(tag, useUnmergedTree = true).assertExists()
-        assertFalse(tag.contains("inverted"), "Expected $tag to carry no _inverted infix")
+        val pixels = onNodeWithTag(tag, useUnmergedTree = true).captureToImage().toPixelMap()
+        // A 6dp mark's 1dp ring anti-aliases, so no pixel reads a token exactly: the pixel furthest from
+        // the page background is the mark's ink, and it must lie nearest [color] among every kind's
+        // colour and the disc's own ink — an inverted or swapped mark lies nearest another.
+        val colors = BroadsheetColors()
+        val ink =
+            (0 until pixels.width)
+                .flatMap { x -> (0 until pixels.height).map { y -> pixels[x, y] } }
+                .maxBy { it.distanceTo(colors.background) }
+        val nearest = listOf(colors.accent, colors.accent2, colors.text, colors.neutral100).minBy { ink.distanceTo(it) }
+        assertEquals(color, nearest, "Expected $tag to draw in $color, but its ink $ink lies nearest $nearest")
         val discPx = with(density) { disc.toPx() }
         val wrap = unclippedBounds(wrapTag, useUnmergedTree = true)
         assertTrue(abs(wrap.width - discPx) <= 0.5f, "Expected $wrapTag to be $disc wide, but it was ${wrap.width}px")
@@ -374,11 +385,16 @@ class HomeCalendarMarksUiTest {
             }
             awaitHomeReady()
 
-            assertMarkBeneathDisc("cal_mark_own_planned_$markDay", "cal_day_pill_$markDay", "cal_day_$markDay")
-            assertMarkBeneathDisc("cal_mark_borrowed_planned_$markDay", "cal_day_pill_$markDay", "cal_day_$markDay")
-            assertMarkBeneathDisc("cal_mark_session_planned_$markDay", "cal_day_pill_$markDay", "cal_day_$markDay")
-            assertMarkBeneathDisc("cal_mark_own_planned_$selectedDay", "cal_day_selected_$selectedDay", "cal_day_$selectedDay")
-            assertMarkBeneathDisc("cal_mark_borrowed_planned_$selectedDay", "cal_day_selected_$selectedDay", "cal_day_$selectedDay")
+            // Each kind's own colour (KAN-574 tasks.md): own accent, borrowed accent-2, session text.
+            val colors = BroadsheetColors()
+            val kindColor = mapOf("own" to colors.accent, "borrowed" to colors.accent2, "session" to colors.text)
+            val todayWrap = "cal_day_pill_$markDay"
+            val selectedWrap = "cal_day_selected_$selectedDay"
+            assertMarkBeneathDisc("cal_mark_own_planned_$markDay", colors.accent, todayWrap, "cal_day_$markDay")
+            assertMarkBeneathDisc("cal_mark_borrowed_planned_$markDay", colors.accent2, todayWrap, "cal_day_$markDay")
+            assertMarkBeneathDisc("cal_mark_session_planned_$markDay", colors.text, todayWrap, "cal_day_$markDay")
+            assertMarkBeneathDisc("cal_mark_own_planned_$selectedDay", colors.accent, selectedWrap, "cal_day_$selectedDay")
+            assertMarkBeneathDisc("cal_mark_borrowed_planned_$selectedDay", colors.accent2, selectedWrap, "cal_day_$selectedDay")
 
             val seventeenth = unclippedBounds("cal_day_$markDay")
             val eighteenth = unclippedBounds("cal_day_$betweenDay")
@@ -393,7 +409,7 @@ class HomeCalendarMarksUiTest {
                 val wrap = if (day == markDay.toString()) "cal_day_pill_$day" else "cal_day_selected_$day"
                 val kinds = if (day == markDay.toString()) listOf("own", "borrowed", "session") else listOf("own", "borrowed")
                 kinds.forEach { kind ->
-                    assertMarkBeneathDisc("cal_mark_${kind}_planned_$day", wrap, "cal_day_$day", gap = 6.dp)
+                    assertMarkBeneathDisc("cal_mark_${kind}_planned_$day", kindColor.getValue(kind), wrap, "cal_day_$day", gap = 6.dp)
                 }
             }
             val plainMarkTop = unclippedBounds("cal_mark_own_planned_$plainDay", useUnmergedTree = true).top
@@ -616,3 +632,11 @@ class HomeCalendarMarksUiTest {
             assertEquals("rotated-refresh", session.refreshToken())
         }
 }
+
+/** Squared RGB distance — enough to rank which of a few tokens a pixel lies nearest. */
+private fun Color.distanceTo(other: Color): Float {
+    val r = red - other.red
+    val g = green - other.green
+    val b = blue - other.blue
+    return r * r + g * g + b * b
+}
````

  - [x] **Step 3: Verify.** In that worktree: `./gradlew ktlintFormat`, then
    `./gradlew verifyChange :shared:desktopTest --tests 'com.gymie.shared.ui.HomeCalendarMarksUiTest'` is
    `BUILD SUCCESSFUL`. Prove the read: in `shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/HomeCalendarGrid.kt`
    replace `color = calendarMarkColor(colors.accent, inMonth, colors),` with
    `color = calendarMarkColor(colors.neutral100, inMonth, colors),` and run
    `./gradlew :shared:desktopTest --tests 'com.gymie.shared.ui.HomeCalendarMarksUiTest.todayAndSelectedKeepEveryMarkBeneathTheDiscInItsOwnColour'`:
    it fails on `Expected cal_mark_own_planned_2026-09-17 to draw in …`; restore with
    `git checkout -- shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/HomeCalendarGrid.kt`.
    `git status --short` then shows only the test file.
    <!-- measured: that mutation and run in a scratch gymie-frontend worktree @ 5ce0d237 → exit 1, the own mark's assertion -->
  - [x] **Step 4: Commit** the file in gymie-frontend with the subject below. The session lands it
    by gymie-frontend's `## default landing route` (`design.md`, `kan-652-gymie-frontend-worktree`).

**Build:** green
**Files:** `shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt`
**Tests:** **none** — the task strengthens the existing `todayAndSelectedKeepEveryMarkBeneathTheDiscInItsOwnColour`; Step 3's mutation proves its colour read.
**Regression:** none — no test is declared; reverting the commit restores the tag-literal check, which passes on an inverted mark.
**Commit:** test(shared): assert each calendar mark's colour beneath the disc
**After:** none

**Decision:** marks-colour-nearest-token
**Decision:** kan-652-gymie-frontend-worktree
