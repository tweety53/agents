# Finish gate, Jira projection, and commit hygiene — Implementation Plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Make Jira behaviour normative and correct its two defects; refuse to integrate silently
over unfinished work; make review findings countable; take planning artifacts out of the reviewed
diff; and put every temporary artifact in one registry.

**Architecture:** Two new guard scripts land first, each with its own harness, because they are the
only mechanically testable parts and every later task references them by name. The contract text
then follows in dependency order — the canonical contracts before the skills that point at them,
and the README last, so no task leaves a dangling cross-reference for `check-references.sh` to
catch.

**Tech Stack:** POSIX shell under `bash` for both guards and both harnesses; Markdown for every
contract, skill and spec file. No new dependency of any kind.

## Global Constraints

- **No suppression markers, and no weakening of any guard's configuration.** This repository's lint
  policy is fix-first; a false hit is fixed by correcting the text or the classifier, never by
  silencing a line.
- **No auto-fix command exists in this repository.** The Lint Fix Priority rule's auto-fix step is
  inapplicable here rather than skipped.
- **No per-task commits.** `pipeline.md`'s git boundaries give `/myflow-do` `git add` only, unless a
  `prUrl` is already recorded. This deliberately overrides the writing-plans template's per-task
  commit step — do not add commits.
- **All three lint guards are expected to exit zero** at the end of every task:
  `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-plan-provenance.sh`.
- **Every fenced block and numeric claim added to a planning artifact carries a provenance tag**,
  per `skills/myflow-contracts/plan-provenance.md`. The guard scans this change's own `tasks.md`,
  `design.md` and `proposal.md`.
- **`bash` 3.2 compatibility.** `test-check-finish-preflight.sh`'s header records that only indexed
  arrays are available; associative arrays are not. Both new harnesses follow it.
- **Both new guards print exactly one verdict line on stdout, and exit non-zero with no verdict line
  when they cannot read what they were pointed at.** The verdict carries the answer, not the exit
  status — `check-finish-preflight.sh`'s header states why, and both new scripts follow that
  contract rather than inventing another.
- **A cross-reference is written in one of the shapes `check-references.sh` recognises** —
  `**Section** (`path`)`, `**Section** in `path`, or `see/per/under **Section** … `path``.
  A reference written in another shape is not checked, which is worse than not writing it.

## File structure

| File | Responsibility | Tasks |
|---|---|---|
| `scripts/check-unfinished-work.sh` | The run-1 gate: four signals, one verdict | 1 |
| `scripts/test-check-unfinished-work.sh` | Its harness; one case per signal and per refusal | 1 |
| `scripts/check-cleanup-complete.sh` | Run-2 verification: registry rows that should be gone | 2 |
| `scripts/test-check-cleanup-complete.sh` | Its harness | 2 |
| `skills/myflow-contracts/jira-integration.md` | The Jira projection contract | 3 |
| `skills/myflow-start/SKILL.md` | Earlier transition; the planning gate | 4 |
| `skills/myflow-do/SKILL.md` | Staging split; findings table; the guide's new section | 5, 6, 7 |
| `skills/myflow-finish/SKILL.md` | The run-1 gate; two commits; In Review; run-2 cleanup | 8, 9 |
| `skills/myflow-contracts/pipeline.md` | Git boundaries; the cleanup registry | 5, 8, 9 |
| `README.md` | The command table and the pipeline summary | 10 |

---

### Task 1: The run-1 gate script

**Files:**
- Create: `scripts/check-unfinished-work.sh`
- Create: `scripts/test-check-unfinished-work.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: `check-unfinished-work.sh <worktree> <change-name>`, printing one line beginning
  `CLEAR:` or `OUTSTANDING:` and exiting zero, or printing nothing on stdout and exiting two.
  Tasks 8 and 9 reference this exact invocation and these exact verdict tokens.

- [x] **Step 1: Write the failing harness**

Create `scripts/test-check-unfinished-work.sh`, following the sandbox and assertion idiom of the
existing harness:

```bash verified:idiom copied from scripts/test-check-finish-preflight.sh:11-43, read in full
#!/usr/bin/env bash
# Assertion harness for check-unfinished-work.sh. Builds throwaway worktree
# fixtures under a sandboxed TMPDIR and asserts the guard's verdict and exit
# status. Never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in openspec/specs/myflow-finish-cleanup/spec.md, never against
# observed output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-unfinished-work.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOXES=()
cleanup() {
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

run_guard() {
  set +e
  OUT="$("$GUARD" "$1" "$2" 2>&1)"
  RC=$?
  set -e
}
```

Then a fixture builder that writes a complete, clean change — every box ticked, every plan item
checked, no open findings, and a `## Known incomplete` section reading `None.`:

```bash unverified:confirm the guide and panel-record paths match what myflow-do writes in Tasks 6 and 7
# new_fixture -> sets WT, and writes a fully finished change named "demo"
new_fixture() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/unfinished-work-test.XXXXXX")"
  SANDBOXES+=("$WT")
  mkdir -p "$WT/docs/manual-test" "$WT/openspec/changes/demo" "$WT/.superpowers/sdd"
  cat > "$WT/docs/manual-test/demo.md" <<'GUIDE'
# Manual test — demo

- [x] the app starts

## Known incomplete

None.
GUIDE
  printf -- '- [x] 1.1 done\n' > "$WT/openspec/changes/demo/tasks.md"
  cat > "$WT/.superpowers/sdd/final-review-panel.md" <<'PANEL'
| Slot | Severity | Location | Status | Note |
|---|---|---|---|---|
| Bugbot | Minor | a.sh:1 | fixed | corrected |
PANEL
}
```

And the cases — one per signal, plus the two refusal shapes:

```bash unverified:assertion strings must match the verdict text Step 3 writes
new_fixture; run_guard "$WT" demo
[ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^CLEAR:' \
  && pass "a finished change is CLEAR" || fail "clean: rc=$RC out=$OUT"

new_fixture; printf -- '- [ ] the app starts\n' > "$WT/docs/manual-test/demo.md"
run_guard "$WT" demo
[ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^OUTSTANDING:' \
  && pass "an unticked guide box is OUTSTANDING" || fail "guide box: rc=$RC out=$OUT"

new_fixture; printf -- '- [ ] 1.1 not done\n' > "$WT/openspec/changes/demo/tasks.md"
run_guard "$WT" demo
printf '%s' "$OUT" | grep -q '^OUTSTANDING:' \
  && pass "an unchecked plan item is OUTSTANDING" || fail "plan item: rc=$RC out=$OUT"

new_fixture
printf '| Bugbot | Minor | a.sh:1 | open | unfixed |\n' \
  >> "$WT/.superpowers/sdd/final-review-panel.md"
run_guard "$WT" demo
printf '%s' "$OUT" | grep -q '^OUTSTANDING:' \
  && pass "an open finding is OUTSTANDING" || fail "open finding: rc=$RC out=$OUT"

new_fixture
printf '# Manual test — demo\n\n- [x] ok\n\n## Known incomplete\n\n- the fix is not written\n' \
  > "$WT/docs/manual-test/demo.md"
run_guard "$WT" demo
printf '%s' "$OUT" | grep -q '^OUTSTANDING:' \
  && pass "a non-empty Known incomplete section is OUTSTANDING" || fail "known: rc=$RC out=$OUT"

new_fixture; rm "$WT/docs/manual-test/demo.md"
run_guard "$WT" demo
printf '%s' "$OUT" | grep -q '^OUTSTANDING:' \
  && pass "a missing guide is OUTSTANDING, not CLEAR" || fail "missing guide: rc=$RC out=$OUT"

new_fixture
printf '# Manual test — demo\n\n- [x] ok\n' > "$WT/docs/manual-test/demo.md"
run_guard "$WT" demo
printf '%s' "$OUT" | grep -q '^OUTSTANDING:' \
  && pass "a missing Known incomplete section is OUTSTANDING" || fail "no section: rc=$RC out=$OUT"

run_guard "/nonexistent/worktree" demo
[ "$RC" -ne 0 ] && ! printf '%s' "$OUT" | grep -qE '^(CLEAR|OUTSTANDING):' \
  && pass "an unreadable worktree exits non-zero with no verdict" \
  || fail "unreadable: rc=$RC out=$OUT"

[ "$FAILURES" -eq 0 ] && echo "All check-unfinished-work assertions passed" || exit 1
```

- [x] **Step 2: Run the harness and watch every case fail**

Run: `chmod +x scripts/test-check-unfinished-work.sh && ./scripts/test-check-unfinished-work.sh`
Expected: every case FAILs — the guard does not exist yet, so `$GUARD` is not executable and `RC`
is non-zero for all of them, including the cases that expect a verdict line.

- [x] **Step 3: Write the guard**

Create `scripts/check-unfinished-work.sh`. Argument handling and the unreadable-tree refusal mirror
`check-finish-preflight.sh:32-49`; the four signals are counted independently and reported together
so the operator sees the whole picture in one prompt rather than one signal at a time:

```bash unverified:run the harness in Step 4; the counting expressions are written, not yet executed
#!/usr/bin/env bash
# check-unfinished-work.sh — report whether a change carries unfinished work,
# for /myflow-finish's run-1 gate.
#
# Usage: check-unfinished-work.sh <worktree> <change-name>
#
# Prints ONE verdict line to stdout:
#   CLEAR: <reason>          nothing outstanding
#   OUTSTANDING: <breakdown> one or more signals fired
#
# Exit 0 whenever a verdict was reached; exit 2 when the worktree cannot be
# read. The VERDICT carries the answer, not the exit status — see
# check-finish-preflight.sh's header for why this repository separates them.
#
# A MISSING FILE COUNTS AS OUTSTANDING, NOT CLEAR. Treating an absent guide or
# an absent "## Known incomplete" section as clearance is the failure this
# guard exists to prevent: a branch merged over unfinished work because nothing
# was written down.
set -euo pipefail

WORKTREE="${1:-}"
NAME="${2:-}"

if [ -z "$WORKTREE" ] || [ -z "$NAME" ]; then
  echo "usage: check-unfinished-work.sh <worktree> <change-name>" >&2
  exit 2
fi
if [ ! -d "$WORKTREE" ]; then
  echo "check-unfinished-work: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

GUIDE="$WORKTREE/docs/manual-test/$NAME.md"
PANEL="$WORKTREE/.superpowers/sdd/final-review-panel.md"
REASONS=""
add() { REASONS="${REASONS:+$REASONS; }$1"; }

# Signal 1 and 4 — the guide's boxes, and its Known incomplete section.
if [ ! -f "$GUIDE" ]; then
  add "no manual test guide at $GUIDE"
else
  BOXES="$(grep -c -- '- \[ \]' "$GUIDE" || true)"
  [ "${BOXES:-0}" -gt 0 ] && add "$BOXES unticked box(es) in the manual test guide"
  if ! grep -q '^## Known incomplete' "$GUIDE"; then
    add "the manual test guide has no '## Known incomplete' section"
  else
    KNOWN="$(sed -n '/^## Known incomplete/,/^## /p' "$GUIDE" \
      | sed '1d;/^## /d' | grep -v '^[[:space:]]*$' | grep -v '^None\.$' || true)"
    [ -n "$KNOWN" ] && add "the guide records work known to be incomplete"
  fi
fi

# Signal 2 — the plan, including any nested fix sub-change.
PLAN_OPEN=0
for f in "$WORKTREE/openspec/changes/$NAME/tasks.md" \
         "$WORKTREE/openspec/changes/$NAME"-fix-*/tasks.md; do
  [ -f "$f" ] || continue
  N="$(grep -c -- '- \[ \]' "$f" || true)"
  PLAN_OPEN=$((PLAN_OPEN + ${N:-0}))
done
[ "$PLAN_OPEN" -gt 0 ] && add "$PLAN_OPEN unchecked plan item(s)"

# Signal 3 — findings whose recorded status is open.
if [ ! -f "$PANEL" ]; then
  add "no review panel record at $PANEL"
else
  OPEN="$(grep -c -E '^\|[^|]*\|[^|]*\|[^|]*\|[[:space:]]*open[[:space:]]*\|' "$PANEL" || true)"
  [ "${OPEN:-0}" -gt 0 ] && add "$OPEN open finding(s) in the review panel record"
fi

if [ -z "$REASONS" ]; then
  echo "CLEAR: every checklist box is ticked, every plan item is checked, no finding is open, and nothing is recorded as incomplete"
else
  echo "OUTSTANDING: $REASONS"
fi
exit 0
```

- [x] **Step 4: Run the harness and watch every case pass**

Run: `chmod +x scripts/check-unfinished-work.sh && ./scripts/test-check-unfinished-work.sh`
Expected: `All check-unfinished-work assertions passed`, exit zero.

- [x] **Step 5: Confirm the repository's own guards stay clean**

Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-provenance.sh`
Expected: all three exit zero.

---

### Task 2: The run-2 cleanup verification script

**Files:**
- Create: `scripts/check-cleanup-complete.sh`
- Create: `scripts/test-check-cleanup-complete.sh`

**Interfaces:**
- Consumes: nothing from Task 1 — the two guards share no code, deliberately, since they answer
  different questions at different points in the run.
- Produces: `check-cleanup-complete.sh <repo> <change-name> <state-dir>`, printing one line
  beginning `COMPLETE:` or `LEFTOVER:` and exiting zero, or printing nothing on stdout and exiting
  two. Task 9 references this invocation and these tokens.

- [x] **Step 1: Write the failing harness**

Create `scripts/test-check-cleanup-complete.sh` with the same header, sandbox and assertion helpers
as Task 1's harness, then a fixture that is a repository after a successful cleanup — no worktree,
no local branch, no remote branch, no proposal artifact source, and the change directory moved under
`openspec/changes/archive/`:

```bash unverified:confirm `git worktree list --porcelain` output shape on this git; run the harness in Step 4
new_fixture() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/cleanup-complete-test.XXXXXX")"
  STATE="$(mktemp -d "${TMPDIR:-/tmp}/cleanup-complete-state.XXXXXX")"
  SANDBOXES+=("$REPO" "$STATE")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  mkdir -p "$REPO/openspec/changes/archive/2026-01-01-demo"
  echo x > "$REPO/f.txt"
  git -C "$REPO" add -A
  git -C "$REPO" commit -qm base
}
```

Cases, one per registry row the guard checks:

```bash unverified:assertion strings must match the verdict text Step 3 writes
new_fixture; run_guard "$REPO" demo "$STATE"
[ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^COMPLETE:' \
  && pass "a fully cleaned repository is COMPLETE" || fail "clean: rc=$RC out=$OUT"

new_fixture; git -C "$REPO" branch openspec/demo
run_guard "$REPO" demo "$STATE"
printf '%s' "$OUT" | grep -q '^LEFTOVER:' \
  && pass "a surviving local branch is LEFTOVER" || fail "local branch: rc=$RC out=$OUT"

new_fixture; touch "$STATE/demo-proposal-artifact.html"
run_guard "$REPO" demo "$STATE"
printf '%s' "$OUT" | grep -q '^LEFTOVER:' \
  && pass "a surviving proposal artifact source is LEFTOVER" || fail "artifact: rc=$RC out=$OUT"

new_fixture; mkdir -p "$REPO/openspec/changes/demo"
run_guard "$REPO" demo "$STATE"
printf '%s' "$OUT" | grep -q '^LEFTOVER:' \
  && pass "an unarchived change directory is LEFTOVER" || fail "change dir: rc=$RC out=$OUT"

run_guard "/nonexistent/repo" demo "$STATE"
[ "$RC" -ne 0 ] && ! printf '%s' "$OUT" | grep -qE '^(COMPLETE|LEFTOVER):' \
  && pass "an unreadable repository exits non-zero with no verdict" \
  || fail "unreadable: rc=$RC out=$OUT"
```

- [x] **Step 2: Run the harness and watch every case fail**

Run: `chmod +x scripts/test-check-cleanup-complete.sh && ./scripts/test-check-cleanup-complete.sh`
Expected: every case FAILs — the guard does not exist yet.

- [x] **Step 3: Write the guard**

```bash unverified:run the harness in Step 4; the git queries are written, not yet executed
#!/usr/bin/env bash
# check-cleanup-complete.sh — verify that everything the cleanup registry says
# should be gone after /myflow-finish run 2 actually is.
#
# Usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>
#
# Prints ONE verdict line to stdout:
#   COMPLETE: <reason>    every registry row whose lifetime ends at run 2 is gone
#   LEFTOVER: <breakdown> one or more are still present
#
# Exit 0 whenever a verdict was reached; exit 2 when the repository cannot be
# read. Reporting a leftover is the whole point: run 2 previously assumed its
# own removals succeeded.
set -euo pipefail

REPO="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"

if [ -z "$REPO" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>" >&2
  exit 2
fi
if [ ! -d "$REPO" ] || ! git -C "$REPO" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-cleanup-complete: $REPO is not a git repository — cannot determine anything" >&2
  exit 2
fi

LEFT=""
add() { LEFT="${LEFT:+$LEFT; }$1"; }

# A command substitution, never a scratch file: a verifier that writes into the
# tree it is verifying can leave behind exactly the class of leftover it exists
# to report.
WT_LIST="$(git -C "$REPO" worktree list --porcelain 2>/dev/null \
  | awk -v b="refs/heads/openspec/$NAME" '/^worktree /{w=$2} /^branch /{if ($2==b) print w}' || true)"
[ -n "$WT_LIST" ] && add "worktree(s) still registered for openspec/$NAME"

git -C "$REPO" show-ref --verify --quiet "refs/heads/openspec/$NAME" \
  && add "local branch openspec/$NAME still exists"

git -C "$REPO" show-ref --verify --quiet "refs/remotes/origin/openspec/$NAME" \
  && add "remote-tracking ref origin/openspec/$NAME still exists"

[ -d "$REPO/openspec/changes/$NAME" ] \
  && add "openspec/changes/$NAME was never moved into the archive"

[ -f "$STATE_DIR/$NAME-proposal-artifact.html" ] \
  && add "the proposal artifact source is still in the state directory"

if [ -z "$LEFT" ]; then
  echo "COMPLETE: no worktree, branch, remote ref, change directory or artifact source remains for $NAME"
else
  echo "LEFTOVER: $LEFT"
fi
exit 0
```

- [x] **Step 4: Run the harness and watch every case pass**

Run: `chmod +x scripts/check-cleanup-complete.sh && ./scripts/test-check-cleanup-complete.sh`
Expected: `All check-cleanup-complete assertions passed`, exit zero.

- [x] **Step 5: Add both harnesses to the project's test list**

Modify the `## test` section of `.myflow/project.md` to list
`scripts/test-check-unfinished-work.sh` and `scripts/test-check-cleanup-complete.sh` alongside the
existing five entries. Do **not** add either guard to `## lint`: both need a change in flight and
would fail on an unrelated invocation, which is why `check-finish-preflight.sh` is not a lint step
either.

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit zero.

---

### Task 3: The Jira projection contract

**Files:**
- Modify: `skills/myflow-contracts/jira-integration.md` — the **Transitions** section, and a new
  section for labels on created issues
- Test: `scripts/check-references.sh`, `scripts/check-vocabulary.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: the section names Tasks 4, 8 and 9 reference — **Transitions**, **Unrecognised
  statuses**, **Labels on issues the pipeline creates**, and **Never blocking**. Those exact bold
  tokens are what `check-references.sh` matches against this file's headings, so the headings and
  the references must be written together.

- [x] **Step 1: Replace the transition table's timing rows**

In the **Transitions** table, change `/myflow-start`'s row from `end of the run, after the state
write` to `start of the run, immediately after the key resolves`, and change `/myflow-finish`'s
In Review row from `run 1, after the PR is confirmed open` to `run 1, after the chosen route
completes — every route`.

Immediately below the table, replace the paragraph beginning "No other command transitions the
issue" so it keeps that sentence and adds:

> `/myflow-start` transitions **before** brainstorming rather than after the state write. That
> ordering does not weaken **Never blocking**: a failed transition is still one line and the run
> still writes its state at the end exactly as it would have. What the old ordering protected was
> the state write, and nothing about an earlier call makes the state write depend on Jira.

- [x] **Step 2: Add the unrecognised-status rule**

Add a new `### Unrecognised statuses` section after **Transitions**:

> The order is To Do, In Progress, In Review, Done, matched by name. A status outside those four
> has **no position** in it. Do not infer one — in particular do not infer one from Jira's
> `statusCategory`, which groups a custom `TO DO URGENT` with `In Progress` under `indeterminate`
> and would report the issue as already at the target, freezing the board for the whole change.
>
> Show the operator the issue key, its current status and the intended target, and ask whether to
> transition. Only an explicit yes transitions it; anything else leaves the status untouched and
> emits one skipped-with-reason line.

- [x] **Step 3: Add the labels rule**

Add a `### Labels on issues the pipeline creates` section:

> An issue any `/myflow-*` command creates carries **every label on the change's linked issue, plus
> `AI-generated`**. No label is invented: the parent's labels exist by construction, and
> `AI-generated` is applied only because the project already uses it. With no linked issue, the
> created issue carries `AI-generated` alone. Link the created issue to the change's issue whenever
> one exists.

- [x] **Step 4: Confirm the two requirements this change only codifies**

`myflow-jira-projection` states two requirements that describe behaviour this contract **already**
has: "Jira is never a gate" and "The issue description is appended to, never rewritten". They need
no edit, only proof that the existing text still satisfies them and was not disturbed by Steps 1
through 3.

Run: `grep -n 'Never blocking' skills/myflow-contracts/jira-integration.md`
Expected: the section survives, and still requires exactly one `⚠ Jira: skipped — <reason>` line on
every failure path.

Run: `grep -n 'exact prefix' skills/myflow-contracts/jira-integration.md`
Expected: the pre-write assertion under **Description sync** survives, with both conditions — exact
prefix and strictly longer — intact.

- [x] **Step 5: Verify the contract's own cross-references still resolve**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-plan-provenance.sh`
Expected: all three exit zero. `check-references.sh` is the real test here — every skill that says
**Jira integration** (`skills/myflow-contracts/jira-integration.md`) is checked against this file's
headings, so a section renamed without its referrers fails the run.

---

### Task 4: `/myflow-start` — the earlier transition and the planning gate

**Files:**
- Modify: `skills/myflow-start/SKILL.md` — section **A. Resolve the change**, section **F. Write
  state and hand off**, and **Guardrails**
- Test: `scripts/check-references.sh`

**Interfaces:**
- Consumes: the section names Task 3 produced.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Move the transition to section A**

At the end of **A. Resolve the change**, after the key is resolved and before the effort question,
add:

> **Transition the issue to In Progress now**, per **Transitions** in **Jira integration**
> (`skills/myflow-contracts/jira-integration.md`) — before brainstorming, so the board is correct
> while planning runs. A failure is one skipped-with-reason line and planning continues; nothing
> about this call may delay or alter the proposal.

- [x] **Step 2: Remove the transition from section F**

In **F. Write state and hand off**, delete the sentence beginning "**Transition the issue to In
Progress after the state write**" and replace it with:

> The In Progress transition already happened in section **A**. Sync added scope here, per
> **Description sync** in **Jira integration** (`skills/myflow-contracts/jira-integration.md`) —
> only when this run added scope the issue does not already describe.

- [x] **Step 3: Add the planning-gate guardrails**

Append to **Guardrails**:

> - **Never** resolve an open question by assumption. Put it to the operator, at the point the
>   answer is first needed, and do everything that does not depend on it in the meantime.
> - **Never** ask for an approval in open prose. Offer named options, mark the recommended one, and
>   say what each one will do.

- [x] **Step 4: Verify**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit zero.
Run: `grep -n 'after the state write' skills/myflow-start/SKILL.md`
Expected: no match for the In Progress transition — the only surviving uses, if any, concern
description sync.

---

### Task 5: The staging split

**Files:**
- Modify: `skills/myflow-do/SKILL.md:220-228` (the staging block) and its handoff block
- Modify: `skills/myflow-contracts/pipeline.md` — the **Git boundaries** table and the
  **Handoff output** bullet about the review diff
- Test: `scripts/check-references.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: the exclusion pathspec that Task 8 stages the complement of.

- [x] **Step 1: Replace `/myflow-do`'s staging block**

Replace the three-command staging block with:

```bash verified:the pathspec was run in a scratch repository; it staged only the non-excluded file
git -C <worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
git -C <worktree> status
git -C <worktree> diff --cached --stat
```

and add below it:

> **Those three paths are never staged.** The exclusion is what keeps them out of the diff, rather
> than a filter applied when the diff is displayed: a filtered display leaves them in the staging
> area, where the IDE's staged-changes pane and `git status` show them again. The list is fixed —
> the pipeline chooses these paths itself, so no project can differ. `/myflow-finish` stages and
> commits them separately, so nothing is lost by leaving them unstaged here.

- [x] **Step 2: Drop the filter from the handoff**

In the handoff block, change the review command from
`git -C <absolute worktree path> diff --cached -- . ':(exclude)openspec/'` to
`git -C <absolute worktree path> diff --cached`, because there is now nothing staged to exclude.

- [x] **Step 3: Update `pipeline.md`**

In **Git boundaries**, change `/myflow-do`'s two `git add` rows to read
`git add` **excluding the planning paths**, and change `/myflow-finish` run 1's row to
`**Commits twice** — implementation, then planning artifacts`.

In **Handoff output**, replace the bullet beginning "**The review diff excludes `openspec/`**" with
a bullet stating that the three paths are never staged before finish, keeping the existing
explanation that the plan was read at `STARTED`.

- [x] **Step 4: Verify**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit zero.
Run: `grep -rn "exclude)openspec" skills/ README.md`
Expected: matches only where a *commit* or an *archive* path legitimately still excludes it — no
surviving instance describes the reviewed diff.

---

### Task 6: The findings table and the every-severity bar

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — section **5. The review panel**, its **Panel re-runs**
  subsection, and **Guardrails**
- Test: `scripts/check-vocabulary.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: the panel record's table shape that Task 1's guard already counts —
  column four is the status, and `open` is the blocking value. The guard's expression matches a row
  whose fourth cell is exactly `open`, so the column order here is load-bearing.

- [x] **Step 1: Require the findings table**

In **5. The review panel**, after the slot table, add:

> **Every finding is recorded as a row** in `.superpowers/sdd/final-review-panel.md`:
>
> | Slot | Severity | Location | Status | Note |
> |---|---|---|---|---|
> | Bugbot | Minor | `src/Foo.kt:42` | fixed | replaced the silent catch |
>
> `Status` is exactly one of `open`, `fixed` or `withdrawn`. `withdrawn` means the finding was
> retracted with a stated reason in the note, never that it was ignored. Free prose is not a record
> of a finding's state: a state that cannot be counted cannot be enforced, and
> `scripts/check-unfinished-work.sh` counts this column.

- [x] **Step 2: Widen the bar to every severity**

In **Panel re-runs**, replace "handoff still requires **zero** open Critical/Important findings from
every agent that has run" with "handoff still requires **zero open findings at any severity** from
every agent that has run".

In the same subsection, widen the fix-wave sentence too: "Union all Critical/Important findings,
dedupe by file:line + theme" becomes "Union all **open** findings, dedupe by file:line + theme".
A bar that blocks on every severity while the fix wave still collects only two of them would leave
the run unable to clear its own gate.

Add immediately after it:

> A minor finding blocks the handoff exactly as a critical one does. The escalation ladder is what
> makes that terminate: when fix rounds do not converge the run hands back to the operator, who
> resolves the disagreement — including by marking a row `withdrawn` with a reason. That handback is
> the existing human gate, not a routine way to defer a finding.

- [x] **Step 3: Update the guardrail**

Replace the guardrail "**Never** hand off with an open Critical/Important finding, or a stale clean
result" with "**Never** hand off with an open finding of any severity, or a stale clean result".

- [x] **Step 4: Verify**

Run: `scripts/check-vocabulary.sh && scripts/check-references.sh`
Expected: both exit zero.
Run: `grep -rn 'Critical/Important' skills/myflow-do/SKILL.md`
Expected: **no match.** All three occurrences — the handoff bar, the fix-wave union, and the
guardrail — are widened by this task. The severity words still appear in the reviewer prompt files
beside `SKILL.md`, which is their own vocabulary and is not touched here.

---

### Task 7: `## Known incomplete` in the manual test guide

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — section **6. Write the manual test guide**
- Test: `scripts/check-references.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: the `## Known incomplete` heading Task 1's guard greps for, spelled exactly.

- [x] **Step 1: Add the section's duty**

Append to section **6**:

> - **Always write a `## Known incomplete` section.** Either the single word `None.` or a bullet per
>   item the run knows is unfinished — a defect instrumented but not fixed, a case deliberately left
>   for later, a box that cannot be ticked yet. Refresh it on every fix run.
>
>   Finish runs in a different session and has no memory of this one, so anything not written here
>   is invisible at the integration gate. `scripts/check-unfinished-work.sh` reads this section, and
>   treats its **absence** as outstanding rather than as clear.

- [x] **Step 2: Verify**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit zero.
Run: `grep -n '## Known incomplete' skills/myflow-do/SKILL.md scripts/check-unfinished-work.sh`
Expected: the heading is spelled identically in both files.

---

### Task 8: Finish run 1 — the gate, two commits, and In Review

**Files:**
- Modify: `skills/myflow-finish/SKILL.md` — sections **1.1** through **1.5**
- Modify: `skills/myflow-contracts/pipeline.md` — **Run 1 — the branch is not merged**
- Test: `scripts/check-references.sh`

**Interfaces:**
- Consumes: `check-unfinished-work.sh <worktree> <change-name>` from Task 1, and the exclusion
  pathspec from Task 5.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Put the gate before the landing question**

Insert a new section **1.0 Check for unfinished work** ahead of **1.1**:

> Run `scripts/check-unfinished-work.sh <worktree> <name>` once per worktree in the state file's
> `worktrees` map, **before the landing question and before any git action**.
>
> - `CLEAR:` from every worktree → continue to **1.1** with no extra prompt.
> - `OUTSTANDING:` → show the breakdown and offer exactly three courses: **Continue — integrate
>   anyway**, **Stop — I'll finish it first**, and **File a Jira task, then continue**.
> - **No verdict line, and a non-zero exit** → stop and ask the operator. Never read missing output
>   as either verdict, and check the exit code as well as the line.
>
> **Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed.
> **File a Jira task** creates an issue carrying the outstanding items, labelled per **Labels on
> issues the pipeline creates** in **Jira integration**
> (`skills/myflow-contracts/jira-integration.md`), then continues.
>
> Asking the landing question first and only then reporting unfinished work would make the operator
> choose a route for a branch they have not yet been told is incomplete.

- [x] **Step 2: Split the commit in 1.2**

Replace the single `git add -A` instruction in **1.2** with:

```bash unverified:the second add is unconstrained on purpose; confirm the first leaves nothing else unstaged
git -C <worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
git -C <worktree> commit -m "<type>(<name>): <what the implementation does>"
git -C <worktree> add -A
git -C <worktree> commit -m "chore(<name>): plan, test guide and session records"
```

and add:

> **Implementation first, planning artifacts second.** The newest commit on a branch is the one a
> forge shows first, and that should be the code. The second commit's message lists anything the
> operator chose to integrate over at **1.0** — the git history is then the durable record that the
> transcript is not.
>
> `scripts/preserve-session-records.sh` still runs **before** the first `add`, unchanged:
> `docs/superpowers/` is one of the excluded paths, so its files are picked up by the second staging
> pass.

- [x] **Step 3: Make In Review route-independent in 1.5**

Replace "**Transition the issue to In Review** once a PR is confirmed open" with:

> **Transition the issue to In Review** at the end of a successful run 1, whichever route was taken
> — pull request, merge and push, or manual. Per **Transitions**
> (`skills/myflow-contracts/jira-integration.md`): after the state write, never before, never
> blocking. A run that stopped on a failed push does **not** transition; the branch never left the
> operator's hands.

Note the reference shape above: `**Transitions** (`path`)`, with **no** second bold token between
the section name and the path. Task 4 proved by mutation that the older
`**Transitions** in **Jira integration** (`path`)` shape checks only the file title, leaving the
section name unguarded. Mutation-test whatever you write.

- [x] **Step 4: Mirror both changes in `pipeline.md`**

In **Run 1 — the branch is not merged**, add the gate ahead of the landing question, and change
"All three routes first commit the staged work" to describe the two commits in order.

- [x] **Step 5: Correct a claim Task 5 made impossible**

`pipeline.md`'s ordering-asymmetry paragraph under **Finish contract** justifies `/myflow-do`'s
preserve-call position by saying that hoisting it "would create **and stage** `docs/superpowers/`
files on every ordinary staged-only run". Since Task 5, `docs/superpowers/` is excluded from
`/myflow-do`'s staging pass, so the staging half is no longer possible — the paragraph now asserts
something false rather than merely stale.

Drop the two words so it reads "would create `docs/superpowers/` files on every ordinary
staged-only run". The rest of the paragraph, including the instruction not to harmonise the two
orderings, stays exactly as it is: creating the files unnecessarily is still the reason.

- [x] **Step 6: Verify**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit zero.
Run: `grep -n 'PR is confirmed open' skills/ -r`
Expected: no match — the PR-conditioned In Review rule is gone from every file.

---

### Task 9: Finish run 2 — the registry, the remote branch, and verification

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` — **Worktree cleanup**, plus a new registry section
- Modify: `skills/myflow-finish/SKILL.md` — run 2's numbered outline
- Test: `scripts/check-references.sh`

**Interfaces:**
- Consumes: `check-cleanup-complete.sh <repo> <change-name> <state-dir>` from Task 2.
- Produces: the registry section name later contract edits point at.

- [x] **Step 1: Add the registry**

Add a `## Temporary artifacts registry` section to `pipeline.md` with one row per artifact —
the per-task and review diffs, the panel record, the session ledger, the proposal artifact source,
each worktree, the local branch, the remote branch, the change directory, and the state file —
naming for each what creates it, where it lives, and what removes it. Take the rows verbatim from
the table in this change's `design.md`.

Then replace every scattered restatement of a removal rule with a pointer to this section, so the
rules are stated once.

- [x] **Step 2: Delete the remote branch**

In **Worktree cleanup**, after the `git branch -d` step, add:

```bash unverified:confirm the exact stderr wording git emits for an already-deleted remote branch
git -C "$REPO" push origin --delete "openspec/<name>" || true
```

> The remote branch is deleted without a further prompt. Run 2 has already proved the branch is an
> ancestor of the base branch, so its commits are in the base branch and nothing can be lost — which
> is why this is not gated the way the ignored-file disclosure is. A branch the forge already deleted
> on merge is **success**, not an error. Report the outcome either way.

- [x] **Step 3: Fix the worktree-scan snippet's path truncation**

`pipeline.md`'s fallback scan — the one used when the state file's `worktrees` map is absent —
reads the worktree path as awk's `$2`:

```bash verified:run against awk on this machine; the truncation below is its actual output
git -C "$REPO" worktree list --porcelain \
  | awk '/^worktree /{w=$2} /^branch /{if ($2=="refs/heads/openspec/<name>") print w}'
```

`worktree list --porcelain` emits the path **raw**, so a path containing a space is truncated at the
first space: fed `worktree /tmp/my worktree/x`, that snippet prints `/tmp/my`. Run 2 would then
`git worktree remove --force` a path that is not the worktree, or fail having reported the wrong one.
Replace `w=$2` with `w=substr($0,10)`, which prints the full path — `10` is one past the length of
the literal `worktree ` prefix.

`scripts/check-cleanup-complete.sh` already parses this stream correctly; this step brings the
contract's own snippet into line with the guard that verifies its result. Leaving the two
disagreeing is how the wrong one gets copied next.

- [x] **Step 4: Verify the cleanup**

Add a step after the removals:

> Run `scripts/check-cleanup-complete.sh <repo> <name> <state-dir>`.
> `COMPLETE:` → report the cleanup as verified. `LEFTOVER:` → name what remains in the handoff
> rather than assuming the removals worked. A non-zero exit with no verdict line → report it and
> leave the affected `worktrees` entries in the state file.

- [x] **Step 5: Mirror it in the skill's run-2 outline**

Add the verification as a step in `myflow-finish/SKILL.md`'s run 2 list, and extend its step about
removing worktrees to name the remote branch.

- [x] **Step 6: Verify**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit zero.
Run: `./scripts/test-check-cleanup-complete.sh && ./scripts/test-check-unfinished-work.sh`
Expected: both harnesses pass — the contract edits must not have changed the guards.

---

### Task 10: README and the final sweep

**Files:**
- Modify: `README.md`, `AGENTS.md`, `CLAUDE.md` — the `/myflow` commands reference table (one copy each)
- Modify: `commands/myflow-finish.md` and `commands-claude/myflow-finish.md` — stale run-1 description
- Modify: `.myflow/project.md` — the `## lint` rationale
- Modify: `skills/myflow-start/SKILL.md` and `skills/myflow-do/SKILL.md` — the description-echo slot
- Modify: any file carrying an unchecked subsection cross-reference
- Test: all three guards and both new harnesses

**Interfaces:**
- Consumes: every preceding task.
- Produces: nothing.

- [x] **Step 1: Update the command table**

In the `/myflow-*` commands reference, change the `/myflow-do` row to say staging excludes the
planning paths and the panel requires zero open findings at any severity; change the
`/myflow-finish` row to say run 1 checks for unfinished work before asking how to land, commits
twice, and moves the issue to In Review on every route, and that run 2 removes the remote branch and
verifies the cleanup.

**The same table exists three times.** `AGENTS.md` and `CLAUDE.md` each carry their own copy of the
command reference, and both still summarise `/myflow-do` as ending in `git add -A`, which Task 5
made false. Update all three, and check they agree afterwards:

```bash verified:run in the apply worktree after Task 5; both files matched, README did not
grep -n 'git add -A' README.md AGENTS.md CLAUDE.md
```

These two files are also the repository's own `## standards` entries, so a stale command table there
is what the principles reviewer reads as this project's stated rules.

**A fourth and fifth copy live in the command trees.** `commands/myflow-finish.md` and
`commands-claude/myflow-finish.md` both still describe run 1 as asking up front how the branch
should land and committing the staged work — one commit, no gate. `pipeline.md` is explicit that
this is not cosmetic: every command file, in **both** trees, must agree with the skill it delegates
to, because when a command and its skill disagree, whichever the agent reads first wins — which is
non-determinism in the one layer that must be deterministic.

**All four outer copies also restate cleanup rules, and the registry requirement forbids that.**
Task 9's spec requires every removal rule to be stated once, in the registry, with everything else
pointing at it. Task 9 could only enforce that inside the two files it owned; these four still carry
their own summaries of what run 2 removes:

| File | What it restates |
|---|---|
| `commands/myflow-finish.md:15` | run 2 "removes the worktrees and branches" |
| `commands-claude/myflow-finish.md:11` | the same sentence |
| `AGENTS.md:156` | the four gating checks, the ignored-file disclosure, the preserved-copy condition |
| `CLAUDE.md:110` | the same, verbatim |

A one-line summary that *names* what run 2 does is fine — these are reference tables, not the
contract. What must go is any restatement of a **rule** (a gate, a condition, a safety check),
which becomes a pointer at the registry instead. The requirement is not satisfied until it does.

Bring both command files into line with `skills/myflow-finish/SKILL.md` as Task 8 left it, and check
the other command files in both trees, reporting which needed no change:

```bash verified:run in the apply worktree after Task 8; only the two myflow-finish files matched
grep -ln 'Commits the staged work\|git add -A\|Asks up front\|after the state write\|Critical/Important' commands/*.md commands-claude/*.md
```

- [x] **Step 2: Name the two new guards in the `## lint` rationale**

`.myflow/project.md`'s `## lint` section explains why `check-finish-preflight.sh` and
`preserve-session-records.sh` are deliberately **not** lint steps: both need a change in flight and
a real worktree, so they would fail on every unrelated invocation. That same reasoning now covers
`check-unfinished-work.sh` and `check-cleanup-complete.sh`, which Task 2 added to `## test` without
extending this paragraph. Add them to it by name.

An unnamed exclusion reads as an oversight, and the next person to tidy the file adds them to
`## lint` and breaks every run that has no change in flight.

- [x] **Step 3: Give the handoffs a slot for the echoed description**

`myflow-jira-projection`'s requirement "The issue description is appended to, never rewritten" ends:
the pre-edit description SHALL be echoed verbatim into the handoff on any run that writes it, so the
transcript is the recovery path when no local backup exists. `jira-integration.md` states this, but
neither `/myflow-start`'s nor `/myflow-do`'s handoff block has anywhere to put it — an implementer
following those templates literally would sync a description and echo nothing.

Add one conditional line to the handoff block in `skills/myflow-start/SKILL.md` and in
`skills/myflow-do/SKILL.md`, in each block's existing voice: on a run that wrote the description,
the pre-edit text is reproduced verbatim in a fenced block, inside `<details>` when long, and
nothing is echoed on a run that did not. Do not restructure either handoff.

This is the only requirement in the change's delta specs that no earlier task implements — found by
Task 4's reviewer, which is why it lands here rather than in the task that owns either file.

- [x] **Step 4: Sweep the subsection cross-references into the shape that is actually checked**

Task 4 established, by mutation, that `check-references.sh` associates a bold token with a path only
when **no second bold token sits in the gap between them**, and only when that gap is at most sixty
characters **and** at most three words. So the repository's pervasive shape —

```text verified:mutation-tested in Task 4; renaming the subsection left the guard exiting 0
**Subsection** in **Jira integration** (`skills/myflow-contracts/jira-integration.md`)
```

— checks only the **file title**. Every subsection name referenced that way is an unguarded
interface: rename the section and nothing fails.

Rewrite each such reference as either of these, both of which mutation-test correctly:

```text verified:the shape Task 4 landed on and proved by renaming the target heading
**Subsection** in Jira integration (`skills/myflow-contracts/jira-integration.md`)
**Subsection** (`skills/myflow-contracts/jira-integration.md`)
```

The file title becomes plain text, or is dropped where the sentence does not need it — plain text in
the gap is fine, only a second **bold** token breaks the association. Mind the three-word budget:
`in Jira integration` is exactly three words and fits; `in the Jira integration contract` does not,
and silently restores the bug being swept.

Find them with the pattern below, then **prove each rewrite** by renaming its target heading, seeing
`check-references.sh` fail and name the line, and restoring the heading byte-for-byte. A reference
you have not seen fail is a reference you have not checked.

```bash verified:run against this worktree after Task 4; the count is this pattern's output there
grep -rn '\*\*[^*]*\*\* in \*\*[^*]*\*\* (`[^`]*\.mdc\?`)' skills/ rules/ README.md
```
<!-- measured: the grep above, run in the apply worktree after Task 4 landed; it reported two lines -->

One of the two is `skills/myflow-start/SKILL.md:37`, which Task 4 deliberately left alone as
belonging to this sweep.

- [x] **Step 5: Run everything**

Run:

```bash verified:the five existing commands are the `## test` and `## lint` lists in .myflow/project.md, read in full
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
```

Expected: every command exits zero.

- [x] **Step 6: Confirm the installer still works against a sandbox**

Run: `SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global`
Expected: the run completes and links the two new scripts nowhere — `setup.sh` installs skills,
commands and rules, and `scripts/` is used from the repository. Confirm no error is reported.
