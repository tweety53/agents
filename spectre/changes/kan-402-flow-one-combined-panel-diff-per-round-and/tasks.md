# kan-402-flow-one-combined-panel-diff-per-round-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

**Spec:** `design.md` beside this file. Every task is a `Build: green` task with its own commit.
Task 1 is the only code task and is independent of tasks 2 and 3; task 2 depends on task 1's
sixth argument existing; task 3 is documentation only and depends on neither.

- [x] 1. Scope `gather-dispatch-context.sh`'s plan section to named task ids

**Build:** green
**Files:** `scripts/gather-dispatch-context.sh`, `scripts/test-gather-dispatch-context.sh`
**Tests:** `scoped to one id: header plus that task's block only`, `scoped to two ids given in reverse: document order`, `a fenced task-line lookalike is not a task: exit 2`, `an unknown id exits 2 naming it`, `the five-argument call keeps the whole plan` — five cases appended to `scripts/test-gather-dispatch-context.sh`
**Regression:** reverting this commit makes the six-argument call exit 2 with the usage message, so every scoped case fails with `call exited 2`; the five-argument case still passes on its own, which is why the first four are the ones that pin the feature.
**Baseline:** before=60 after=65
<!-- measured: ./scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok: ' @ branch main -->
<!-- predicted: ./scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok: ' after this task, on branch spectre/kan-402-flow-one-combined-panel-diff-per-round-and -->
**Commit:** `feat(scripts): scope the dispatch bundle's plan section to named task ids`

  - [x] **Step 1: Append the five failing cases to `scripts/test-gather-dispatch-context.sh`.** Insert
    them after case 36 (the trailing-newline case) and before the final `if [ "$FAILURES" -ne 0 ]`
    block. They need a six-argument capture and a plan fixture with real task lines; both are
    local to the new block so no existing case changes.

```sh unverified:confirm `capture()`'s ERR/OUT contract at scripts/test-gather-dispatch-context.sh:80-98 is unchanged before copying its shape
# ===========================================================================
# CASES 37-41 (kan-402): the optional sixth argument scopes the ## tasks.md
# section to the named task ids. design.md's scope-tasks-not-files and
# unknown-task-id-exit-2 decisions are the contract; plan-dispatch-bundles.py
# is the task-line and fence grammar.
# ===========================================================================

# capture_scoped <task-ids> -> RC, ERR, OUT, invoking with the six arguments.
capture_scoped() {
  set +e
  ERR="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" demo "$PRINCIPLES" "$OUTPUT_PATH" "$1" 2>&1 1>/dev/null)"
  RC=$?
  set -e
  if [ -f "$OUTPUT_PATH" ]; then
    OUT="$(cat "$OUTPUT_PATH")
$ERR"
  else
    OUT="$ERR"
  fi
}

# write_plan_fixture -> overwrites CHANGE_ROOT/tasks.md with three tasks, a
# step line under task 1, and a fenced task-line lookalike inside task 2.
# The fence lines are built from a variable so this suite's own file carries
# no fence at column 0 inside a fenced plan block (the plan-provenance and
# build-green guards toggle fence state on any such line).
write_plan_fixture() {
  local fence='```'
  printf '%s\n' \
    '# demo plan' '' '> **Execution:** header line' '' \
    '- [ ] 1. First' '**Files:** `a`' '  - [ ] **Step 1: one**' '' \
    '- [ ] 2. Second' '**Files:** `b`' '' \
    "${fence}sh" '- [ ] 9. Not a task' "$fence" '' \
    '- [ ] 3. Third' '**Files:** `c`' > "$CHANGE_ROOT/tasks.md"
}

# CASE 37: one id -> the header, task 1's block (its step included), nothing else.
new_repo
write_plan_fixture
capture_scoped 1
if [ "$RC" -ne 0 ]; then
  fail "scoped one id: call exited $RC: $OUT"
elif ! printf '%s' "$OUT" | grep -q '^> \*\*Execution:\*\* header line$'; then
  fail "scoped one id: plan header missing from the section: $OUT"
elif ! printf '%s' "$OUT" | grep -q '^- \[ \] 1\. First$'; then
  fail "scoped one id: task 1's line missing: $OUT"
elif ! printf '%s' "$OUT" | grep -q '^  - \[ \] \*\*Step 1: one\*\*$'; then
  fail "scoped one id: task 1's step missing: $OUT"
elif printf '%s' "$OUT" | grep -q 'Second\|Third\|Not a task'; then
  fail "scoped one id: another task's block leaked into the section: $OUT"
else
  pass "scoped to one id: header plus that task's block only"
fi

# CASE 38: ids "3,1" -> both blocks, task 1 before task 3, task 2 absent.
new_repo
write_plan_fixture
capture_scoped 3,1
if [ "$RC" -ne 0 ]; then
  fail "scoped two ids: call exited $RC: $OUT"
elif printf '%s' "$OUT" | grep -q 'Second'; then
  fail "scoped two ids: task 2 leaked: $OUT"
else
  LINE_1_38="$(printf '%s\n' "$OUT" | grep -n '^- \[ \] 1\. First$' | cut -d: -f1)"
  LINE_3_38="$(printf '%s\n' "$OUT" | grep -n '^- \[ \] 3\. Third$' | cut -d: -f1)"
  if [ -z "$LINE_1_38" ] || [ -z "$LINE_3_38" ]; then
    fail "scoped two ids: a named task's line is missing: $OUT"
  elif [ "$LINE_1_38" -gt "$LINE_3_38" ]; then
    fail "scoped two ids: task 3 printed before task 1 — not document order: $OUT"
  else
    pass "scoped to two ids given in reverse: document order"
  fi
fi

# CASE 39: the fenced `- [ ] 9.` line is worked-example text, never a task.
new_repo
write_plan_fixture
capture_scoped 9
if [ "$RC" -ne 2 ]; then
  fail "fenced lookalike: expected exit 2, got $RC: $OUT"
elif ! printf '%s' "$ERR" | grep -q 'task 9 not found in tasks.md'; then
  fail "fenced lookalike: stderr does not name the missing task: $ERR"
else
  pass "a fenced task-line lookalike is not a task: exit 2"
fi

# CASE 40: an id with no task line at all.
new_repo
write_plan_fixture
capture_scoped 7
if [ "$RC" -ne 2 ]; then
  fail "unknown id: expected exit 2, got $RC: $OUT"
elif ! printf '%s' "$ERR" | grep -q 'task 7 not found in tasks.md'; then
  fail "unknown id: stderr does not name the missing task: $ERR"
else
  pass "an unknown id exits 2 naming it"
fi

# CASE 41: without the sixth argument the whole plan is carried, lookalike included.
new_repo
write_plan_fixture
run_it
if [ "$RC" -ne 0 ]; then
  fail "five-argument call: exited $RC: $OUT"
elif ! printf '%s' "$OUT" | grep -q '^- \[ \] 2\. Second$' \
  || ! printf '%s' "$OUT" | grep -q '^- \[ \] 9\. Not a task$'; then
  fail "five-argument call: the plan section is no longer the whole file: $OUT"
else
  pass "the five-argument call keeps the whole plan"
fi
```

  - [x] **Step 2: Run the suite and confirm cases 37–40 fail on exit 2 with the usage message.**
    Run: `./scripts/test-gather-dispatch-context.sh 2>&1 | tail -8`
    Expected: four `FAIL:` lines whose text ends in `usage: gather-dispatch-context.sh <worktree>
    <change-root> <name> <principles-path> <output-path>` (the script rejects nothing about a
    sixth argument today — it never reads `$6` — so confirm the exact failure mode you see and
    report it if it differs), case 41 passing, and `4 case(s) failed`.

  - [x] **Step 3: Read and validate the sixth argument.** Below the existing five `${N:-}`
    assignments and their usage check, add:

```sh unverified:confirm the usage line in the script's `if [ -z "$WORKTREE" ] …` block is the one the header comment quotes, and update both to show `[<task-ids>]`
TASK_IDS="${6:-}"
case "$TASK_IDS" in
  "" | [0-9]* ) ;;
  *) echo "gather-dispatch-context: task ids '$TASK_IDS' must be a comma-separated list of integers" >&2; exit 2 ;;
esac
if [ -n "$TASK_IDS" ] && ! printf '%s' "$TASK_IDS" | grep -qE '^[0-9]+(,[0-9]+)*$'; then
  echo "gather-dispatch-context: task ids '$TASK_IDS' must be a comma-separated list of integers" >&2
  exit 2
fi
```

    Update the usage echo and the header's `# Usage:` line to
    `gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids>]`,
    and add one header paragraph stating what `<task-ids>` does, that a named id absent from the
    plan is exit 2, and that the grammar is `plan-dispatch-bundles.py`'s — pointing at that
    script's docstring rather than restating it.

  - [x] **Step 4: Add `scope_tasks` and hook it in after `add_fixed_source "$TASKS_FILE" "tasks.md"`.**
    The function prints the plan header (every line before the first task line) plus each named
    task's block in document order; it exits 1 after naming every id it did not find.

```sh verified:run with macOS awk against the case-37 fixture on 2026-09-04 — ids 3,1 print the header, task 1 and task 3 in document order and exit 0; id 9 exits 1 naming it; still confirm the fence regex against plan-dispatch-bundles.py's before committing
# scope_tasks <tasks-file> <ids> — print the plan header and the named tasks'
# blocks, in document order. Task-line, body-boundary and fence grammar are
# plan-dispatch-bundles.py's (see its docstring), so the ids that script
# printed are the ids this finds. Exits 1, naming each missing id on stderr.
scope_tasks() {
  awk -v ids="$2" '
    BEGIN {
      gsub(/ /, "", ids)
      n = split(ids, want, ",")
      for (i = 1; i <= n; i++) w[want[i]] = 1
      fence = 0; header = 1; cur = ""
    }
    {
      if ($0 ~ /^(```|~~~)/) { fence = !fence; if (header || cur != "") print; next }
      if (!fence && $0 ~ /^- \[[ x]\] [0-9]+\. /) {
        id = $0; sub(/^- \[[ x]\] /, "", id); sub(/\..*$/, "", id)
        header = 0
        cur = (id in w) ? id : ""
        if (cur != "") seen[cur] = 1
      } else if (!fence && !header && $0 ~ /^##(#)?([ \t]|$)/) {
        cur = ""
      }
      if (header || cur != "") print
    }
    END {
      bad = 0
      for (i = 1; i <= n; i++) if (!(want[i] in seen)) {
        printf "gather-dispatch-context: task %s not found in tasks.md\n", want[i] > "/dev/stderr"
        bad = 1
      }
      exit bad
    }
  ' "$1"
}
```

    Then, immediately after the three `add_fixed_source` calls:

```sh unverified:confirm FOUND_LABELS/FOUND_PATHS run parallel and that "" in FOUND_PATHS is the only sentinel render_body currently reads
TASKS_SCOPED_BODY=""
if [ -n "$TASK_IDS" ]; then
  last=$(( ${#FOUND_LABELS[@]} - 1 ))
  if [ "$last" -lt 0 ] || [ "${FOUND_LABELS[$last]}" != "tasks.md" ]; then
    echo "gather-dispatch-context: task ids given but tasks.md is absent or refused" >&2
    exit 2
  fi
  TASKS_SCOPED_BODY="$(scope_tasks "${FOUND_PATHS[$last]}" "$TASK_IDS")" || exit 2
  FOUND_LABELS[$last]="tasks.md (scoped to task(s) $TASK_IDS)"
  FOUND_PATHS[$last]="@scoped-tasks"
fi
```

    In `render_body`, replace the `if [ -n "${FOUND_PATHS[$i]}" ]; then cat …; else printf … fi`
    with a three-way `case "${FOUND_PATHS[$i]}" in "@scoped-tasks") printf '%s\n' "$TASKS_SCOPED_BODY" ;; "") printf '%s' "$PROJECT_COMMANDS_BODY" ;; *) cat "${FOUND_PATHS[$i]}" ;; esac`.
    The scoped body enters `BODY` like any other section, so the hash sidecar and
    skip-when-unchanged need no change.

  - [x] **Step 5: Run the suite to green.**
    Run: `./scripts/test-gather-dispatch-context.sh 2>&1 | tail -3`
    Expected: `all cases passed`, and `./scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok: '` prints 65.

  - [x] **Step 6: Run the project lint and commit.**
    Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-provenance.sh && scripts/check-task-build-green.sh`
    Expected: every guard exits 0. Then `git add scripts/gather-dispatch-context.sh scripts/test-gather-dispatch-context.sh` and commit with this task's `**Commit:**` subject and a `Task-Id: 1` trailer.

- [x] 2. Gather one bundle per dispatch bundle in `implement.md`

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — prose; `scripts/check-references.sh` and `scripts/check-vocabulary.sh` are the checks that run against it
**Regression:** reverting this commit leaves `implement.md` gathering one full-plan bundle before any implementer, so task 1's sixth argument has no caller and every implementer keeps reading the whole plan — the KAN-29 cost this change exists to cut.
**Baseline:** before=65 after=65
<!-- predicted: ./scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok: ' — unchanged by a documentation task; task 1 sets the count -->
**Commit:** `docs(flow): gather the implementer bundle once per dispatch bundle`

  - [x] **Step 1: Replace the pre-dispatch gather paragraph.** In section 4 (`## 4. Execute (SDD +
    TDD)`), the paragraph opening **Gather the dispatch context bundle before dispatching any
    implementer.** and its fenced `gather-dispatch-context.sh` call currently run once, before the
    `plan-dispatch-bundles.sh` call. Move that paragraph to sit **after** the `plan-dispatch-bundles.sh`
    block and its exit-code sentences, and reword its opening to:

```markdown unverified:confirm against the current section-4 text that the `<changeRoot>` and `<principles-path>` definitions, the never-gates sentence, the `test -f` check and the stderr-line reporting sentence are all carried over unchanged
**Gather one bundle per dispatch bundle, immediately before that bundle's implementer goes out.**
Take `<k>` and the ids from the `bundle <k>: <ids>` line `plan-dispatch-bundles.sh` printed for
it, comma-separated:
```

    followed by a `bash` fence carrying:

```sh unverified:confirm `plan-dispatch-bundles.py` prints the ids space-separated, so the caller joins them with commas
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context-bundle-<k>.md <id>[,<id>…]
```

    and then this closing paragraph:

```markdown unverified:confirm design.md's decision id `scope-tasks-not-files` is spelled exactly so
The sixth argument scopes the bundle's `## tasks.md` section to the plan header and the named
tasks' blocks (per design.md's `scope-tasks-not-files`); a named id the plan does not carry is
exit 2, a plan defect reported like a missing `**Files:**` field. The panel's and the fix
subagent's bundles (`skills/flow/review-panel.md`) keep the five-argument call and the whole plan.
```

    Keep every other sentence of the existing paragraph (`<changeRoot>` and `<principles-path>`
    definitions, "the bundle never gates a run", the `test -f` check on the per-bundle path,
    "never read the bundle back", and the stderr-line report) verbatim, substituting
    `dispatch-context-bundle-<k>.md` for `dispatch-context.md` in the `test -f` command.

  - [x] **Step 2: Repoint the CONTEXT BUNDLE paragraph.** Replace its first sentence so the
    blockquote reads:

```markdown unverified:confirm the remaining three sentences of the paragraph (must still read the diff and code; lint/test/run commands carried; no need to open project.md) are kept verbatim after this one
> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context-bundle-<k>.md` carries this
> change's proposal, design, engineering principles, and — under `## tasks.md` — the plan header
> plus your bundle's own task(s) only, gathered for you.
```

  - [x] **Step 3: Run the guards and commit.**
    Run: `scripts/check-vocabulary.sh && scripts/check-references.sh`
    Expected: both exit 0. `git add skills/flow/implement.md`, commit with this task's `**Commit:**` subject and `Task-Id: 2`.

- [x] 3. One combined `final-review.diff` and one dispatch per slot per round in `review-panel.md`

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — prose; `scripts/check-references.sh` and `scripts/check-vocabulary.sh` are the checks that run against it
**Regression:** reverting this commit restores "one per included slot, in every affected worktree", the per-worktree doubling KAN-402 reports (KAN-29 dispatches 72–79).
**Baseline:** before=65 after=65
<!-- predicted: ./scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok: ' — unchanged by a documentation task; task 1 sets the count -->
**Commit:** `docs(flow): review one combined panel diff per round, one dispatch per slot`

  Every edit below is in `skills/flow/review-panel.md`; the paragraph each names is quoted by its
  opening words as the file reads on `main` today. Section names refer to design.md.

  - [x] **Step 1: The combined diff (design.md "The combined diff").** Replace the paragraph
    opening `Write <abs-worktree>/.superpowers/sdd/final-review.diff from git diff <merge-base>` with:

```markdown unverified:confirm the surrounding paragraphs ("This is the one automatic reduction…" above, "Every slot's dispatch is recorded" below) are untouched
Write `<canonical-worktree>/.superpowers/sdd/final-review.diff` once per round from **every**
worktree in the change's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), in resolved order — each worktree's section
opened by a header naming it and its own working-notes merge base, then that worktree's
`git diff <merge-base>` (staged and unstaged):
```

    followed by a `bash` fence carrying:

```sh unverified:confirm `git diff <merge-base>` with no second ref is the staged-plus-unstaged form the paragraph being replaced already used
: > <canonical-worktree>/.superpowers/sdd/final-review.diff
# for each <worktree> in the resolved set, in order:
printf '# worktree: %s — merge base %s\n' "<abs-worktree>" "<merge-base>" \
  >> <canonical-worktree>/.superpowers/sdd/final-review.diff
git -C <worktree> diff <merge-base> >> <canonical-worktree>/.superpowers/sdd/final-review.diff
```

    and then this closing paragraph:

```markdown unverified:confirm design.md's decision id `combined-diff-per-round` is spelled exactly so
A single-worktree change writes the same shape with one header. Then dispatch **separate** review
subagents — **one per included slot per round**, in the canonical worktree, each reading the whole
combined file. Never merge two slots into one prompt, and never dispatch a slot once per worktree:
one slot reads every worktree's section, so a seam between two repositories is in one reviewer's
view (design.md's `combined-diff-per-round`).
```

  - [x] **Step 2: `-diff-base` and the record (design.md "One dispatch per slot").** In the paragraph
    opening `` `-slot` names the slot from **The roster** table above ``, replace the sentence
    beginning `` `-diff-base <sha>` is the sha a delta-slot's delta starts from `` with:

```markdown unverified:confirm the rest of that paragraph (the `-model` sentences and the "narrowed by" clause) is kept verbatim
`-diff-base <sha>` is passed on a slot dispatched against a delta and on no other; it takes one
sha, so it carries the **canonical worktree's** held last-reviewed sha for that slot, and the
panel record names every worktree's sha beside the delta path (design.md's
`diff-base-canonical-sha`).
```

  - [x] **Step 3: Cap and docs-only across worktrees (design.md "Cap and docs-only across
    worktrees").** After the `check-panel-diff-size.sh <worktree> <merge-base>` fence and before
    `Exit 0 proceeds.`, insert:

```markdown unverified:confirm the exit-0/1/2 sentences that follow still read correctly with "the sum" as their subject
once per worktree in the resolved set, unchanged. **The gating count is the sum across
worktrees** — one slot now reads every section — and the over-cap prompt names the sum and each
worktree's own count (design.md's `cap-sum-across-worktrees`).
```

    In **The docs-only reduction**, change `**exit 0 — every path this branch touched … reduces
    pass 1 to primary alone**` to open `**exit 0 from every worktree in the resolved set — …**`,
    and change `the path the guard printed — the first non-documentation path — is recorded` to
    `the first non-documentation path any worktree's run printed is recorded`. Add one sentence
    after the exit-1 sentence: `One worktree at exit 1 or 2 runs the resolved roster unchanged for
    the whole change.`

  - [x] **Step 4: Bugbot and Security (design.md "Bugbot and Security").** In **Bugbot's throwaway
    worktree**, before the `git -C <worktree> worktree add --detach` fence, add: `Run the sequence
    below once per worktree in the resolved set, producing one <worktree>-bugbot-<round> per
    repository.` Replace `Dispatch Bugbot with Full Repository Path: <worktree>-bugbot-<round> in
    place of <worktree>.` with:

```markdown unverified:confirm the removal fence that follows is reworded to "for each copy" rather than duplicated
Dispatch Bugbot **once**, its prompt listing every `<worktree>-bugbot-<round>` copy as the
repository paths to mutate and test in, in place of `<worktree>` (design.md's
`bugbot-security-one-dispatch`).
```

    Reword the removal sentence to `Remove every copy unconditionally once that dispatch closes`
    and prefix the `git worktree remove` fence with `# for each copy:`. In the sentence `Security
    is **not** isolated this way — … it keeps sharing <worktree> with the reading slots.`, append:
    `It too is dispatched once, its prompt naming every worktree in the resolved set.`

  - [x] **Step 5: Findings across worktrees (design.md "Findings across worktrees").** After the
    REPORT FILE paragraph's blockquote, add a new paragraph and blockquote:

```markdown unverified:confirm the CITATION CHECK paragraph that follows is then updated per step 6 rather than left naming a single file
**Every slot's dispatch prompt also carries the WORKTREES paragraph**, listing the resolved set
and, when it holds more than one worktree, the qualification rule:

> **WORKTREES:** this change spans `<abs-worktree-1>`, `<abs-worktree-2>`, …; `final-review.diff`
> is sectioned by worktree, each section headed `# worktree: <path> — merge base <sha>`. When more
> than one is listed, prefix every finding's `file:line` with that worktree's basename
> (`gymie-frontend:src/Foo.tsx:42`) and write its reproducer to run from that worktree.
```

  - [x] **Step 6: Per-worktree inputs named per file.** In the CONTEXT BUNDLE sentence `the same one
    skills/flow/implement.md's implementer dispatch carries`, change to `the same shape
    skills/flow/implement.md's implementer dispatch carries, naming the panel's own five-argument
    bundle <abs-worktree>/.superpowers/sdd/dispatch-context.md`. In the relocation sentence, change
    `When <abs-worktree>/.superpowers/sdd/relocation-comparison.md exists` to `For every worktree
    whose <abs-worktree>/.superpowers/sdd/relocation-comparison.md exists`. In the CITATION CHECK
    paragraph's lead sentence, change `present only when the citation pre-check above wrote
    <abs-worktree>/.superpowers/sdd/citation-check.md` to `present for every worktree whose
    citation pre-check above wrote <abs-worktree>/.superpowers/sdd/citation-check.md, one path each`.

  - [x] **Step 7: Deltas on re-runs (design.md "Deltas on re-runs").** In **Panel re-runs**, replace
    the paragraph opening `**When the round raised anything above Minor, re-run on deltas.**` up to
    and including `so a round right after one falls under this same no-held-sha rule. Then:` with:

```markdown unverified:confirm the three bullets that follow ("every diff-reading slot…", "Bugbot and Security…", "a slot the operator has not named…") are kept, with the first bullet's "on its delta" now meaning the combined delta
**When the round raised anything above Minor, re-run on deltas.** A slot's last-reviewed sha is
held **per slot per worktree**: each dispatch sets that slot's sha in every worktree to the HEAD it
was dispatched against, and a slot not dispatched in a round keeps the shas it had. A delta is
`<canonical-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`, combined exactly as
`final-review.diff` is — one `# worktree:` header per worktree, followed by that worktree's `git
diff <held-sha> HEAD`; a worktree in which the slot holds no sha contributes its whole `git diff
<merge-base>` section. Every slot's dispatch prompt names the path it was given and, for a delta,
each worktree's starting sha. **Check base movement first** above clears every slot's held sha in
the rebased worktree on a clean panel-entry rebase, so that worktree's section falls under the
no-held-sha rule in the next round. Then:
```

    In the first bullet, change `**A slot whose delta is empty is not dispatched**` to `**A slot
    whose delta is empty in every worktree is not dispatched**`.

  - [x] **Step 8: The cap on a re-run.** Replace the paragraph opening `**The cap check on a re-run**`
    up to `naming the gating count and, when it differs, the full-branch count too.` with:

```markdown unverified:confirm the trailing "Record both in final-review-panel.md…" sentence of that paragraph is kept
**The cap check on a re-run** is `check-panel-diff-size.sh <worktree> <sha> <cap>` once per
worktree per **distinct** held sha among the diff-reading slots dispatched this round (two slots
sharing a sha in a worktree need one call there, not two); a slot with no held sha in a worktree
counts from that worktree's merge base. **The gating count is the largest per-slot sum across
worktrees** — the largest single combined read any one slot this round faces — and an exit-1
result from a call contributing to it puts the over-cap choice to the operator (**The roster**,
above), naming the gating sum, its per-worktree counts and, when it differs, the full-branch sum.
```

  - [x] **Step 9: Run the guards and commit.**
    Run: `scripts/check-vocabulary.sh && scripts/check-references.sh`
    Expected: both exit 0. `git add skills/flow/review-panel.md`, commit with this task's `**Commit:**` subject and `Task-Id: 3`.
