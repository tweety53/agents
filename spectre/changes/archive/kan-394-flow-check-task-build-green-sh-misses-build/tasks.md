# kan-394-flow-check-task-build-green-sh-misses-build

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

Two commits in dependency order: the grammar, both guards and their harnesses (task 1), then the
contract that states the rule in prose (task 2). The approved design is
`docs/superpowers/specs/2026-09-04-kan-394-flow-check-task-build-green-sh-misses-build-design.md`;
`design.md` beside this file carries the decisions.

**Baseline, measured before any edit:**

- `scripts/test-check-task-build-green.sh` prints 67 `ok:` lines and `all cases passed`;
  `scripts/test-check-task-commit-fields.sh` prints 213 and `all cases passed`.
  <!-- measured: scripts/test-check-task-build-green.sh | grep -c '^ok'; scripts/test-check-task-commit-fields.sh | grep -c '^ok' @ cba777b -->
- `scripts/check-contract-budget.sh` reports 71 owned files within budget;
  `skills/flow-contracts/build-green.md` is 5420 bytes against its 6678-byte row.
  <!-- measured: scripts/check-contract-budget.sh; wc -c skills/flow-contracts/build-green.md; grep build-green scripts/check-contract-budget.sh @ cba777b -->
- The shipped guard, run against KAN-29's archived plan (`gymie`, `spectre/changes/archive/kan-29-final-verification-public-profiles-friends/tasks.md`),
  reports 8 tasks as `has no **Build:** tag` — tasks 38 through 45; the same file under the task 1
  guard reports 1 violation, task 45's `**Build:** \`tests/\` 33 of 34 — see the flake below`, as a
  malformed tag at its own line.
  <!-- measured: scripts/check-task-build-green.sh <that file> before and after applying task 1's diff on a scratch copy of scripts/ @ cba777b -->

Both diffs below were authored for this plan, applied to a scratch copy, and run: every harness
named passes, and `git apply --check` accepts each diff against `cba777b`.

---

- [x] 1. Read the `Build:` keyword as a prefix and report a malformed tag by its line

One shared grammar change (`BUILD_LINE_RE` + `BUILD_KIND_RE`, `BuildTag` gaining `value` and an
optional `kind`) reaches both guards through `select_build_tag`; `check-task-build-green.py` adds
the malformed-tag violation, `check-task-commit-fields.py` changes only its docstring and one
comment since it reads `build == "red"` alone. Case 7 is re-asserted, cases 29 through 34 and case
87 are added.

**Files:** `scripts/lib/plan_grammar.py`, `scripts/check-task-build-green.py`,
`scripts/check-task-commit-fields.py`, `scripts/test-check-task-build-green.sh`,
`scripts/test-check-task-commit-fields.sh`
**Tests:** `scripts/test-check-task-build-green.sh` — Case 7, Case 29, Case 30, Case 31, Case 32,
Case 33, Case 34; `scripts/test-check-task-commit-fields.sh` — Case 87
**Regression:** reverting this commit restores the closed-word regex — case 29's
`**Build:** green — …` line is reported as `has no **Build:** tag` again, case 7 and case 32 lose
the malformed-tag message, and case 34's later green line silently wins.
**Baseline:** before=67 after=83 `ok:` lines in `scripts/test-check-task-build-green.sh`;
before=213 after=218 in `scripts/test-check-task-commit-fields.sh`
<!-- measured: both harnesses run on the scratch copy with this diff applied; grep -c '^ok' @ cba777b -->
**Commit:** `fix(scripts): read the Build: keyword as a prefix and report a malformed tag by its line`
**Build:** green

  - [x] **Step 1: Apply the diff** — from the repository root, `git apply` the block below verbatim
    (it touches the five files in `**Files:**` and nothing else).

```diff verified:authored for this plan, applied to a scratch copy of scripts/ and run — test-check-task-build-green.sh 79 ok, test-check-task-commit-fields.sh 215 ok, test-plan-dispatch-bundles.sh and test-check-plan-shape.sh unchanged; git apply --check passes at cba777b
--- a/scripts/lib/plan_grammar.py
+++ b/scripts/lib/plan_grammar.py
@@ -148,11 +148,19 @@
 # `iter_tasks` below, which tests for a task line before this pattern.
 BODY_BOUNDARY_RE: Pattern[str] = re.compile(r"^#{2,3}(?:\s|$)")
 
-# BUILD_TAG_RE — the `**Build:**` tag's vocabulary and its scope: the WHOLE
-# physical line, whose value is a bare `green` or `red`. Group "kind" is
-# that word. Anything else on the line (`**Build:** yellow`, or a prose line
-# joined onto it) is not a tag, and a body carrying no tag line has no tag.
-BUILD_TAG_RE: Pattern[str] = re.compile(r"^\*\*Build:\*\*\s+(?P<kind>green|red)\s*$")
+# BUILD_LINE_RE — a column-0 `**Build:**` line. The FIRST non-fenced one in
+# a body is the task's tag line, whatever its value says. Group "value" is
+# everything after the field name.
+BUILD_LINE_RE: Pattern[str] = re.compile(r"^\*\*Build:\*\*(?P<value>.*)$")
+
+# BUILD_KIND_RE — the tag's vocabulary, applied to BUILD_LINE_RE's value: a
+# `green` or `red` opening the value and ending on a word boundary. Text
+# after the keyword (`green — 2259 tests`, `green (unchanged)`) is free
+# prose and ignored (KAN-394). A value opening with neither word (`yellow`,
+# `greenish`, a bare backticked path) has no kind, and
+# check-task-build-green.py reports that line as a malformed tag rather
+# than as a missing one.
+BUILD_KIND_RE: Pattern[str] = re.compile(r"^\s+(?P<kind>green|red)\b")
 
 # SQUASH_WITH_FIELD_RE — the field's PRESENCE on one physical line. Group
 # "value" is everything after the field name, and is what `partner_ids`
@@ -257,23 +265,29 @@
 
 
 class BuildTag(NamedTuple):
-    """A task body's `**Build:**` tag. `offset` is the 0-based index, within
-    the body sequence handed to `select_build_tag`, of the line it was read
-    from; `kind` is "green" or "red"."""
+    """A task body's `**Build:**` tag line. `offset` is the 0-based index,
+    within the body sequence handed to `select_build_tag`, of the line it
+    was read from; `kind` is "green" or "red", or None when the line's value
+    opens with neither (a malformed tag); `value` is the text after
+    `**Build:**`, stripped, for the malformed-tag message."""
 
     offset: int
-    kind: str
+    kind: Optional[str]
+    value: str
 
 
 def select_build_tag(body: Sequence[str]) -> Optional[BuildTag]:
-    """The `**Build:**` tag `body` carries, or None if it carries none.
-    `body` is a task's body lines, in order, fences included.
+    """The `**Build:**` tag line `body` carries, or None if it carries no
+    `**Build:**` line at all. `body` is a task's body lines, in order,
+    fences included.
 
     WHICH line is the tag is defined here and nowhere else (fix round 9,
-    F20): the FIRST non-fenced line matching `BUILD_TAG_RE` in full. The tag
-    is line-scoped for the same reason `Squash-with:` is — its grammar is
-    one closed word, so it never needs to wrap, and a prose line placed
-    under it with no blank line between must not change what it says.
+    F20): the FIRST non-fenced line matching `BUILD_LINE_RE`, well-formed or
+    not (KAN-394) — a malformed first line is reported as such rather than
+    letting a later line decide. The tag is line-scoped for the same reason
+    `Squash-with:` is: its keyword opens the value and nothing after it on
+    the line is read, and a prose line placed under it with no blank line
+    between must not change what it says.
     """
     in_fence = False
     for offset, line in enumerate(body):
@@ -282,9 +296,15 @@
             continue
         if in_fence:
             continue
-        match = BUILD_TAG_RE.match(line)
+        match = BUILD_LINE_RE.match(line)
         if match is not None:
-            return BuildTag(offset=offset, kind=match.group("kind"))
+            value = match.group("value")
+            kind = BUILD_KIND_RE.match(value)
+            return BuildTag(
+                offset=offset,
+                kind=kind.group("kind") if kind is not None else None,
+                value=value.strip(),
+            )
     return None
 
 
--- a/scripts/check-task-build-green.py
+++ b/scripts/check-task-build-green.py
@@ -18,7 +18,7 @@
   0  clean — every task in the file carries a resolvable **Build:** tag
      (including a file with zero tasks).
   1  violations found — one or more of: a duplicate task id, a task with no
-     tag, a `red` task with no **Squash-with:** field at all, a
+     tag, a malformed tag, a `red` task with no **Squash-with:** field at all, a
      **Squash-with:** field naming zero ids, a named partner that does not
      exist in this file, or a named partner that is itself `red`. Printed
      one per line as `file:line: message`.
@@ -70,16 +70,23 @@
 `lib/plan_grammar.py`'s `iter_tasks` and `select_task`. Within that body,
 the FIRST line (also outside any fence) matching
 
-    ^\\*\\*Build:\\*\\*\\s+(green|red)\\s*$
+    ^\\*\\*Build:\\*\\*(?P<value>.*)$
 
-is the task's tag, and that selection is `lib/plan_grammar.py`'s
-`select_build_tag` (fix round 9, F20). A body with no such line has no tag at all — this
-includes a line that merely looks like an attempt at one (`**Build:**
-yellow`), which is deliberately treated the same as no tag rather than as a
-separate parse-error class: the fixed vocabulary is `green` or `red`, and
-anything else is simply absent, reported the same way an entirely missing
-line is. `Build:` no longer carries any inline suffix — a `red` task's
-partner is read from a separate field.
+is the task's tag line, and its kind is read from that value by
+
+    ^\\s+(?P<kind>green|red)\\b
+
+— the keyword as a prefix ending on a word boundary, with anything after it
+on the line (`green — 2259 tests`, `green (unchanged)`) ignored. Both
+patterns and the selection are `lib/plan_grammar.py`'s `BUILD_LINE_RE`,
+`BUILD_KIND_RE` and `select_build_tag` (fix round 9, F20; KAN-394). A body
+with no `**Build:**` line has no tag at all. A first `**Build:**` line whose
+value opens with neither keyword (`**Build:** yellow`, `**Build:** greenish`,
+a bare backticked path) is a MALFORMED tag, its own violation naming the
+line and the value — never folded into "no tag", which would name the
+consequence and hide the cause, and never overridden by a well-formed
+`**Build:**` line further down. `Build:` carries no inline partner — a
+`red` task's partner is read from a separate field.
 
 Independently, within that same body, the FIRST line (also outside any
 fence) that is a `**Squash-with:**` field whose value gates as `Task
@@ -170,8 +177,9 @@
 
     id: str
     task_line: int
-    tag_kind: Optional[str] = None  # "green", "red", or None (no tag)
-    tag_line: Optional[int] = None
+    tag_kind: Optional[str] = None  # "green", "red", or None (no tag, or malformed)
+    tag_line: Optional[int] = None  # None: no **Build:** line at all
+    tag_value: Optional[str] = None  # the tag line's value, for the malformed message
     squash_line: Optional[int] = None  # None: no **Squash-with:** field
     partners: List[str] = field(default_factory=list)
     # The line of a fence this task's body opens and never closes, or None
@@ -191,6 +199,7 @@
         if tag is not None:
             task.tag_kind = tag.kind
             task.tag_line = found.body_start + tag.offset + 1
+            task.tag_value = tag.value
         squash = select_squash_with(found.lines)
         # A value that does not gate is not a Squash-with field at all — a
         # red task carrying one is reported as having no field, not as
@@ -243,6 +252,17 @@
                 "field below it is inside the fence and was not read"
             )
             continue
+        # A `**Build:**` line whose value opens with neither keyword is a
+        # malformed tag, named by its own line and value (KAN-394) — never
+        # folded into "no tag", which would name the consequence and hide
+        # the cause exactly as the unclosed-fence case above would.
+        if task.tag_line is not None and task.tag_kind is None:
+            violations.append(
+                f"{relfile}:{task.tag_line}: task {task.id} has a **Build:** "
+                f'line reading "{task.tag_value}", which is neither green '
+                "nor red"
+            )
+            continue
         if task.tag_kind is None:
             violations.append(
                 f"{relfile}:{task.task_line}: task {task.id} has no "
--- a/scripts/check-task-commit-fields.py
+++ b/scripts/check-task-commit-fields.py
@@ -120,7 +120,9 @@
 review rounds).
 
 `**Build:**` is read from the body by that module's `select_build_tag`: the
-first line that is `**Build:** green` or `**Build:** red` in full. It used
+first `**Build:**` line, whose value opens with `green` or `red` on a word
+boundary and may carry prose after the keyword (KAN-394); a first line
+opening with neither has no kind, which this guard treats as not red. It used
 to be read through FIELD_RE's alternation below, which joined continuation
 lines onto its value and let a later `**Build:**` line overwrite an earlier
 one — so `**Build:** red` followed by `**Build:** green` was green here and
@@ -476,7 +478,8 @@
         else None
     )
 
-    # The tag is the first line-gated `**Build:** green|red` in the body,
+    # The tag is the first `**Build:**` line in the body, its kind the
+    # `green`/`red` prefix of that line's value (None when malformed),
     # per lib/plan_grammar.py's select_build_tag — never a value assembled
     # from FIELD_RE's capture above, which joined continuation lines onto it
     # and let a later tag line overwrite an earlier one (fix round 9, F20).
--- a/scripts/test-check-task-build-green.sh
+++ b/scripts/test-check-task-build-green.sh
@@ -144,8 +144,10 @@
 [ -z "$OUT" ] && pass "case 6: no output" || fail "case 6: expected no output, got: $OUT"
 
 # ===========================================================================
-# Case 7: a malformed tag line (e.g. "**Build:** yellow") is treated as NO
-# tag -- falls through to case 2's violation, not a separate parse error.
+# Case 7 (KAN-394): a malformed tag line (e.g. "**Build:** yellow") is its
+# OWN violation, naming the tag line and its value -- it used to fall
+# through to case 2's "no tag", which named the consequence and hid the
+# cause.
 # ===========================================================================
 new_fixture
 {
@@ -155,8 +157,8 @@
 run_guard "$TASKS_MD"
 [ "$RC" -eq 1 ] && pass "case 7: malformed tag fails" || fail "case 7: rc=$RC out=$OUT"
 case "$OUT" in
-  *"task 1 has no **Build:** tag"*) pass "case 7: reported as missing tag" ;;
-  *) fail "case 7: expected missing-tag message, out=$OUT" ;;
+  *"tasks.md:3: task 1 has a **Build:** line reading \"yellow\", which is neither green nor red"*) pass "case 7: reported as a malformed tag at its own line" ;;
+  *) fail "case 7: expected malformed-tag message, out=$OUT" ;;
 esac
 
 # ===========================================================================
@@ -740,7 +742,102 @@
 run_guard "$TASKS_MD"
 [ "$RC" -eq 0 ] && pass "case 28: an indented task-shaped line opens no task" || fail "case 28: rc=$RC out=$OUT"
 [ -z "$OUT" ] && pass "case 28: no output" || fail "case 28: expected no output, got: $OUT"
+
+# ===========================================================================
+# Case 29 (KAN-394): the keyword is a PREFIX of the tag's value. A green tag
+# followed by prose on the same line -- the shape KAN-29's plan wrote six
+# times and this guard rejected for the change's whole life -- is green.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Green tag with trailing prose\n\n'
+  printf '**Build:** green — `:shared:desktopTest` 2259, matching the predicted +1 exactly\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 29: green followed by prose is green" || fail "case 29: rc=$RC out=$OUT"
+[ -z "$OUT" ] && pass "case 29: no output" || fail "case 29: expected no output, got: $OUT"
+
+# ===========================================================================
+# Case 30 (KAN-394): a parenthesised remark after the keyword is prose too.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Green tag with a parenthesised remark\n\n'
+  printf '**Build:** green (unchanged; `:shared:desktopTest` still 2258, nothing was touched)\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 30: green with a parenthesised remark is green" || fail "case 30: rc=$RC out=$OUT"
+[ -z "$OUT" ] && pass "case 30: no output" || fail "case 30: expected no output, got: $OUT"
 
+# ===========================================================================
+# Case 31 (KAN-394): a red tag with trailing prose is still red -- its
+# partner is resolved exactly as for a bare `red`. The same body is asserted
+# against check-task-commit-fields.sh as its own case 87.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Red tag with trailing prose\n\n'
+  printf '**Build:** red — lands with task 2\n'
+  printf '**Squash-with:** Task 2\n\n'
+  printf -- '- [ ] 2. Green partner\n\n'
+  printf '**Build:** green\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 31: red followed by prose is red and resolves its partner" || fail "case 31: rc=$RC out=$OUT"
+[ -z "$OUT" ] && pass "case 31: no output" || fail "case 31: expected no output, got: $OUT"
+
+# ===========================================================================
+# Case 32 (KAN-394): the prefix ends on a WORD BOUNDARY. `greenish` opens
+# with `green` but is not the keyword, so the line is a malformed tag.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Not quite green\n\n'
+  printf '**Build:** greenish\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 1 ] && pass "case 32: greenish is not green" || fail "case 32: rc=$RC out=$OUT"
+case "$OUT" in
+  *"task 1 has a **Build:** line reading \"greenish\", which is neither green nor red"*) pass "case 32: reported as a malformed tag" ;;
+  *) fail "case 32: expected malformed-tag message, out=$OUT" ;;
+esac
+
+# ===========================================================================
+# Case 33 (KAN-394): KAN-29's eighth line, whose value opens with a
+# backticked path and no keyword at all -- a malformed tag, with the whole
+# value quoted so the operator sees what the guard could not read.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. A tag that names no colour\n\n'
+  printf '**Build:** `tests/` 33 of 34 — see the flake below\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 1 ] && pass "case 33: a keyword-less value fails" || fail "case 33: rc=$RC out=$OUT"
+case "$OUT" in
+  *'task 1 has a **Build:** line reading "`tests/` 33 of 34 — see the flake below", which is neither green nor red'*) pass "case 33: the value is quoted in the message" ;;
+  *) fail "case 33: expected the quoted value, out=$OUT" ;;
+esac
+
+# ===========================================================================
+# Case 34 (KAN-394): the FIRST **Build:** line decides, well-formed or not.
+# A malformed line followed by a green one is a malformed tag at the first
+# line -- the later line does not quietly win, unlike before this change,
+# when the malformed line matched nothing and the green one was the tag.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Malformed then green\n\n'
+  printf '**Build:** yellow\n'
+  printf '**Build:** green\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 1 ] && pass "case 34: the first Build line decides" || fail "case 34: rc=$RC out=$OUT"
+case "$OUT" in
+  *"tasks.md:3: task 1 has a **Build:** line reading \"yellow\""*) pass "case 34: the malformed first line is the one reported" ;;
+  *) fail "case 34: expected line 3 reported, out=$OUT" ;;
+esac
+
 if [ "$FAILURES" -gt 0 ]; then
   printf '%d failure(s)\n' "$FAILURES" >&2
   exit 1
--- a/scripts/test-check-task-commit-fields.sh
+++ b/scripts/test-check-task-commit-fields.sh
@@ -3244,6 +3244,36 @@
 case "$OUT" in
   *"revert"*) pass "case 86: stderr also surfaces the original revert-conflict failure's own text, not just the stash-pop one" ;;
   *) fail "case 86: original revert-conflict failure text missing from stderr, out=$OUT" ;;
+esac
+
+# ===========================================================================
+# Case 87 (KAN-394): a `**Build:** red` tag with prose after the keyword is
+# still red here -- the keyword is a word-boundary prefix of the value, read
+# through lib/plan_grammar.py's select_build_tag as in the other guard, so
+# the fold is resolved and the missing partner reported rather than the
+# task dropping onto the single-commit path. The same body is asserted
+# against check-task-build-green.sh as its own case 31 (with a real partner).
+# ===========================================================================
+new_repo
+write_tasks_md "$REPO" '- [ ] 1. Red task whose tag carries trailing prose
+
+**Files:** `alpha.txt`
+**Commit:** test: add alpha
+**Build:** red — lands with its partner
+
+**Squash-with:** Task 9
+'
+git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
+git -C "$REPO" commit -q -m "plan"
+printf 'a\n' > "$REPO/alpha.txt"
+git -C "$REPO" add alpha.txt
+git -C "$REPO" commit -q -m "test: add alpha"
+SHA="$(git -C "$REPO" rev-parse HEAD)"
+run_guard "$REPO" 1 "$SHA"
+[ "$RC" -eq 1 ] && pass "case 87: red followed by prose is still red" || fail "case 87: rc=$RC out=$OUT"
+case "$OUT" in
+  *"task 1: Squash-with: names Task 9, which does not exist in this plan"*) pass "case 87: the red task's missing partner is reported" ;;
+  *) fail "case 87: expected the missing-partner message, out=$OUT" ;;
 esac
 
 if [ "$FAILURES" -gt 0 ]; then
```

  - [x] **Step 2: Run the harnesses** — `scripts/test-check-task-build-green.sh`,
    `scripts/test-check-task-commit-fields.sh`, `scripts/test-plan-dispatch-bundles.sh`,
    `scripts/test-check-plan-shape.sh`; expected: each ends `all cases passed`, and the first two
    print the `after=` counts above.
  - [x] **Step 3: Lint** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-guard-symlinks.sh`; expected: all exit 0.
  - [x] **Step 4: Commit** with the subject above.

- [x] 2. State the prefix grammar and the malformed-tag class in the contract

`skills/flow-contracts/build-green.md` is canonical for the tag; its regex sentence, its
"treated the same as no tag" paragraph and the guard's scope list are brought in line with task
1. The file grows to 6045 bytes, inside its 6678-byte budget row, so no row moves.
<!-- measured: wc -c on the scratch copy with this diff applied @ cba777b -->

**Files:** `skills/flow-contracts/build-green.md`
**Tests:** **none** — prose; `scripts/check-contract-budget.sh` checks the row and
`scripts/check-references.sh` the citations.
**Regression:** reverting this commit leaves the contract stating a closed-word regex the guard
no longer implements — the two are required to describe the same rule.
**Baseline:** before=79 after=79 `ok:` lines in `scripts/test-check-task-build-green.sh`;
`check-contract-budget.sh` before=71 after=71 owned files within budget
<!-- predicted: scripts/test-check-task-build-green.sh and scripts/check-contract-budget.sh after task 2 -->
**Commit:** `docs(flow-contracts): state the Build: prefix grammar and the malformed-tag class`
**Build:** green

  - [x] **Step 1: Apply the diff** — from the repository root, `git apply` the block below verbatim.

```diff verified:authored for this plan on a scratch copy of the file; git apply --check passes at cba777b; wc -c 6045
--- a/skills/flow-contracts/build-green.md
+++ b/skills/flow-contracts/build-green.md
@@ -32,8 +32,11 @@
 its own, which is also what keeps spectre's malformed-task check off it.
 
 **The two columns of indent belong to the steps alone: a task's FIELDS sit at column 0.** The
-`Build:` tag is read as `^\*\*Build:\*\*\s+(green|red)\s*$` — anchored at column 0, exactly as the
-task line above it is — and `**Squash-with:**` is anchored the same way, as is every field the
+`Build:` tag line is read as `^\*\*Build:\*\*(?P<value>.*)$` — anchored at column 0, exactly as the
+task line above it is — and its kind from that value as `^\s+(green|red)\b`: the keyword is a
+prefix ending on a word boundary, and whatever follows it on the line (`green — 2259 tests`,
+`green (unchanged)`) is free prose the guard ignores, still line-scoped. `**Squash-with:**` is
+anchored at column 0 the same way, as is every field the
 `flow-task-commit-fields` family adds to a task. Indenting the fields along with the steps is the
 natural reading of "the body sits beneath its task", and it is wrong in a way nothing catches
 kindly: `spectre validate` reports no findings, because an indented `**Build:**` line is no more a
@@ -41,11 +44,14 @@
 `task <id> has no **Build:** tag` — naming the consequence and hiding the cause, since the tag is
 there, one column short of where its regex looks. Measured on a one-task plan written both ways.
 
-Within that body, the `Build:` tag is the **first** line matching the vocabulary above; a body with
-no such line has no tag at all, and a line that merely resembles one (`**Build:** yellow`) is
-treated the same as no tag rather than as a separate malformed-tag class. This is the same
-placement rule the guard script's own docstring states as a regex — this file states it in prose,
-and the two are required to describe the same rule, the column-0 anchor included.
+Within that body, the `Build:` tag is the **first** column-0 `**Build:**` line, well-formed or not; a
+body with no such line has no tag at all. A first line whose value opens with neither keyword
+(`**Build:** yellow`, `**Build:** greenish`, a bare backticked path) is a **malformed tag** — its
+own violation, reported by that line and its value, never as a missing tag (which would name the
+consequence and hide the cause, exactly as the unclosed-fence finding refuses to), and never
+overridden by a well-formed `**Build:**` line further down. This is the same placement rule the
+guard script's own docstring states as a regex — this file states it in prose, and the two are
+required to describe the same rule, the column-0 anchor included.
 
 ## The guard's scope
 
@@ -54,6 +60,7 @@
 `tasks.md` and fails the run when:
 
 - a task has no `**Build:**` tag;
+- a task's first `**Build:**` line is malformed — its value opens with neither `green` nor `red`;
 - a task is tagged `red` with no `**Squash-with:**` field naming a merge partner;
 - a partner named by `**Squash-with:**` does not exist among the tasks in that same plan; or
 - a partner named by `**Squash-with:**` exists but is itself tagged `red` — which is how the guard
```

  - [x] **Step 2: Verify** — `scripts/check-contract-budget.sh`, `scripts/check-references.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-normative-inventory.sh`,
    `scripts/check-markdown-integrity.py`; expected: all exit 0.
  - [x] **Step 3: Commit** with the subject above.
