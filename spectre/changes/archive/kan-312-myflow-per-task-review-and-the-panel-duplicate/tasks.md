# kan-312-myflow-per-task-review-and-the-panel-duplicate

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

**Spec:** `docs/superpowers/specs/2026-09-03-kan-312-myflow-per-task-review-and-the-panel-duplicate-design.md`
— sections 1–4 map onto tasks 1–3 below; `design.md` beside this file carries the decisions.

**Baseline for every task's `Baseline:` field:** `scripts/run-guard-tests.sh` on `main` reports
50 harnesses, 47 passed, 3 failed. The three failures are pre-existing and unrelated:
`test-check-self-review-report.sh` case 7, `test-check-task-build-green.sh` case 13 (this checkout
carries both `spectre/changes/` and `openspec/changes/`), `test-run-reproducer.sh`. The counts
below are the **passed** count.
<!-- measured: scripts/run-guard-tests.sh @ 6c463b5 (main) -->

- [x] 1. Add `check-panel-docs-only.sh` and its test harness

**Build:** green
**Files:** `scripts/check-panel-docs-only.sh`, `scripts/test-check-panel-docs-only.sh`, `skills/flow/scripts/check-panel-docs-only.sh`, `scripts/lib/panel-touched-paths.sh` (added, review round 0 F1: the GIT_BIN resolution, worktree/merge-base validation and COMMITTED/STAGED/UNSTAGED collection this guard and `check-panel-citation-trigger.sh` shared as a line-for-line copy, extracted into one sourced library — no `skills/flow/scripts/` symlink needed, since `lib` is already symlinked there as a whole directory), `scripts/check-panel-citation-trigger.sh` and `scripts/test-check-panel-citation-trigger.sh` (both pre-existing, edited in the same fixup: the guard now sources the library above, and the test's structural/trace checks were repointed at the library file the logic moved into)
**Tests:** `scripts/test-check-panel-docs-only.sh`
**Regression:** `scripts/test-check-panel-docs-only.sh` — reverting this commit removes the guard, so the harness's every case fails with "No such file" from `run_guard`.
**Baseline:** before=47 after=48
<!-- predicted: scripts/run-guard-tests.sh after task 1 — 51 harnesses, 48 passed, the same 3 pre-existing failures -->
**Commit:** `feat(scripts): add check-panel-docs-only.sh, the panel's docs-only guard`

  - [x] **Step 1: Write the harness first**, at `scripts/test-check-panel-docs-only.sh`, in
    `scripts/test-check-panel-citation-trigger.sh`'s shape (indexed `REPOS` array, EXIT trap,
    `run_guard` setting `RC` and `OUT`, `new_repo` creating a `main` repo whose merge-base
    commit carries `base.go` and `tracked.md`). Cases, each asserting the exit code and, where
    named, stdout:

```bash unverified:confirm against the guard once step 3 lands — the cases assert the contract in the spec's section 1, never observed output
# 1. committed .md change only                         -> exit 0, empty stdout
# 2. committed .mdc change only                        -> exit 0, empty stdout
# 3. committed .md plus committed main.go              -> exit 1, stdout is "main.go"
# 4. committed .md plus staged main.go (not committed) -> exit 1, stdout is "main.go"
# 5. committed .md plus unstaged edit to base.go       -> exit 1, stdout is "base.go"
# 6. staged-only .md change, nothing committed         -> exit 0
# 7. unstaged-only edit to tracked.md                  -> exit 0
# 8. no changes at all                                 -> exit 1, empty stdout
# 9. missing arguments                                 -> exit 2
# 10. <worktree> not a directory                       -> exit 2
# 11. <worktree> not a git repository                  -> exit 2
# 12. <merge-base> not resolving                       -> exit 2
# 13. flag-shaped merge base "--evil"                  -> exit 2, OUT contains "does not resolve"
# 14. structural: every `-C "$WORKTREE"` call site in the guard carries the literal
#     `"$GIT_BIN" -C "$WORKTREE"` (the positive check case 11b of the citation harness uses)
# 15. mutation: replace the guard's `\.mdc?$` pattern with `\.never$` via sed into a copy,
#     run case 1 against the copy, assert it now exits 1; then run case 3 against the copy,
#     assert it still exits 1 — proving the pattern is what decides case 1
```

    Case 15 mutates a **copy** of the guard (`cp "$GUARD" "$REPO/guard-mutant.sh"`), never
    the checked-in file, so the harness leaves the tree untouched even when interrupted.
    End with the citation harness's summary shape: `check-panel-docs-only: all cases pass`
    on zero failures, else exit 1.

  - [x] **Step 2: Run the harness and confirm it fails** — `bash scripts/test-check-panel-docs-only.sh`
    exits non-zero with every case reporting the missing guard.

  - [x] **Step 3: Write the guard** at `scripts/check-panel-docs-only.sh`. The header comment
    states the usage, the three exit codes with their meaning (0 every touched path ends `.md` or
    `.mdc`; 1 at least one does not, that first path printed to stdout, an empty path set
    included; 2 cannot answer), that the path set is the union of committed-since-merge-base,
    staged and unstaged paths exactly as `check-panel-citation-trigger.sh` collects it, and why
    the guard exists (KAN-312: the whole-branch panel re-reads a docs-only branch every per-task
    reviewer already read). Body:

```bash unverified:confirm `awk` prints the first non-matching non-empty line and `exit`s on it under macOS awk and gawk alike
set -euo pipefail

GIT_BIN="$(type -P git)" || {
  echo "check-panel-docs-only: no git binary found on PATH" >&2
  exit 2
}

WORKTREE="${1:-}"
MERGEBASE="${2:-}"

if [ -z "$WORKTREE" ] || [ -z "$MERGEBASE" ]; then
  echo "check-panel-docs-only: usage: check-panel-docs-only.sh <worktree> <merge-base>" >&2
  exit 2
fi
if [ ! -d "$WORKTREE" ]; then
  echo "check-panel-docs-only: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi
if ! "$GIT_BIN" -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-panel-docs-only: $WORKTREE is not a git worktree — cannot determine anything" >&2
  exit 2
fi
if ! "$GIT_BIN" -C "$WORKTREE" rev-parse --verify --end-of-options "${MERGEBASE}^{commit}" >/dev/null 2>&1; then
  echo "check-panel-docs-only: merge base '$MERGEBASE' does not resolve in $WORKTREE" >&2
  exit 2
fi

COMMITTED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only --end-of-options "${MERGEBASE}..HEAD" 2>/dev/null)" || {
  echo "check-panel-docs-only: cannot list this change's committed paths in $WORKTREE" >&2
  exit 2
}
STAGED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only --cached 2>/dev/null)" || {
  echo "check-panel-docs-only: cannot list this change's staged paths in $WORKTREE" >&2
  exit 2
}
UNSTAGED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only 2>/dev/null)" || {
  echo "check-panel-docs-only: cannot list this change's unstaged paths in $WORKTREE" >&2
  exit 2
}

PATHS="$(printf '%s\n%s\n%s\n' "$COMMITTED" "$STAGED" "$UNSTAGED" | awk 'NF' | sort -u)"

if [ -z "$PATHS" ]; then
  exit 1
fi

NONDOC="$(printf '%s\n' "$PATHS" | awk '$0 !~ /\.mdc?$/ { print; exit }')"

if [ -n "$NONDOC" ]; then
  printf '%s\n' "$NONDOC"
  exit 1
fi
exit 0
```

    `chmod +x scripts/check-panel-docs-only.sh`.

  - [x] **Step 4: Run the harness and confirm every case passes** —
    `bash scripts/test-check-panel-docs-only.sh` prints `check-panel-docs-only: all cases pass`.

  - [x] **Step 5: Ship the symlink** — `ln -s ../../../scripts/check-panel-docs-only.sh
    skills/flow/scripts/check-panel-docs-only.sh`, the same relative target every sibling in that
    directory carries; `scripts/check-guard-symlinks.sh` exits 0.

  - [x] **Step 6: Run the lint list and the guard suite**, fixing any hit:

```bash unverified:run after steps 1-5 land; predicts a clean run since the two new scripts follow the citation trigger's shape and add no owned Markdown
scripts/check-guard-symlinks.sh
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/run-guard-tests.sh
```

    `run-guard-tests.sh` reports 51 harnesses, 48 passed, the same 3 pre-existing failures.
<!-- predicted: scripts/run-guard-tests.sh after task 1 -->

- [x] 2. Reduce the panel's pass 1 roster on a docs-only branch in `review-panel.md`

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none
**Regression:** none
**Baseline:** before=48 after=48
<!-- predicted: scripts/run-guard-tests.sh after task 2 — unchanged from task 1, this task edits prose only -->
**Commit:** `feat(review-panel): dispatch primary alone on a docs-only branch`

  - [x] **Step 1: Capture the normative inventory before editing** —
    `scripts/check-normative-inventory.sh > "$TMPDIR/inventory-before.txt"`.

  - [x] **Step 2: Insert the guard call and a new subsection** directly after the paragraph that
    begins "Exit 2 stops the run. Record the measured count" (the `check-panel-diff-size.sh`
    block under **The roster**) and before "Write `<abs-worktree>/.superpowers/sdd/final-review.diff`":

    First a one-line lead-in, "Then run", followed by a bash fence carrying exactly this call:

```bash unverified:confirm the symlink task 1 shipped resolves from skills/flow/scripts/ so check-guard-symlinks.sh rule 2 accepts this invocation
check-panel-docs-only.sh <worktree> <merge-base>
```

    then the subsection:

```markdown unverified:confirm the subsection heading level matches its neighbours (### under ## The roster) so check-references.sh resolves the cite task 3 adds
### The docs-only reduction

Per design.md's `docs-only-reduces-to-primary` (narrowing `roster-from-settings`): **exit 0 —
every path this branch touched, committed since the merge base, staged or unstaged, ends `.md` or
`.mdc` — reduces pass 1 to `primary` alone**, plus every slot the operator's per-run instruction
named at this stage's start. Every other resolved slot is recorded in
`<abs-worktree>/.superpowers/sdd/final-review-panel.md` as `not dispatched — docs-only
reduction`. `primary` is the reduced roster even when the resolved list does not carry it — the
same shape **Model resolution** (`skills/flow/SKILL.md`) already defines for an empty store list.
On a docs-only branch the whole-branch read covers the text every per-task reviewer already read
against the same plan; there is no code seam between commits for a second slot to find (KAN-312).

**Exit 1 runs the resolved roster unchanged**; the path the guard printed — the first
non-documentation path — is recorded beside the verdict. An empty touched-path set is exit 1 too.
**Exit 2 reports the guard's stderr and runs the resolved roster unchanged**: an unanswered
question never reduces a panel.

Record the verdict, the printed path where there is one, and the roster actually dispatched in
`<abs-worktree>/.superpowers/sdd/final-review-panel.md` on **every** run, beside the diff-size
fields above.

**This is the one automatic reduction, and it only ever removes.** No slot is ever added by diff
size, touched area, or any other automatic trigger: beyond the reduced or resolved roster,
anything reaches the panel only through an explicit per-run operator instruction, for that run
only.
```

  - [x] **Step 3: Cut the sentence the subsection replaces** under **The roster**: the paragraph
    beginning "**No slot is ever added by diff size, touched area, or any other automatic
    trigger: the roster is exactly the resolved list above" keeps its two check-points and its
    recording rule, and loses its final clause — from "— a documentation-, prompt-, or test-only
    diff with no operator addition runs the resolved list alone" to "never a silently skipped
    review." — ending instead at `"no addition this round — the resolved list ran alone."` Its
    opening sentence, now stated in the subsection, is cut here too: the paragraph opens at
    "Check for one at two points:". Delete, never reword.

  - [x] **Step 4: Rewrite the pass-1 sentence under Panel re-runs.** Replace
    "**Pass 1 always runs the resolved roster plus every slot the operator named at this stage's
    start that the resolved roster did not already carry.** Only re-runs after a fix are scoped."
    with:

```markdown unverified:re-read in place after the edit — the bold cite must name the heading step 2 adds, verbatim
**Pass 1 runs the roster **The docs-only reduction** chose — the resolved roster, or `primary`
alone on a docs-only branch — plus every slot the operator named at this stage's start that it
did not already carry.** Only re-runs after a fix are scoped.
```

  - [x] **Step 5: Add re-classification to the re-run cap check.** After the paragraph beginning
    "**The cap check on a re-run** is `check-panel-diff-size.sh <worktree> <sha> <cap>`", add:

```markdown unverified:confirm "no last-reviewed sha is held" is still the phrase the delta rule above uses, so this paragraph cites the rule it relies on by its own words
**The docs-only guard runs again beside that cap check**, `check-panel-docs-only.sh <worktree>
<merge-base>`, per design.md's `fix-rounds-reclassify`. A branch that stays docs-only keeps the
reduced roster, and `primary` re-runs on its delta as above. A branch the fix round made no
longer docs-only — exit 1 or 2 where pass 1 saw exit 0 — dispatches, in this round, every
resolved slot not yet dispatched this run, each reading the whole `final-review.diff` under the
no-last-reviewed-sha rule above; Bugbot and Security among them take their pass-1 shape, throwaway
worktree included. Record which slots joined this way and the path the guard printed.
```

  - [x] **Step 6: Diff the normative inventory** —
    `diff <(scripts/check-normative-inventory.sh) "$TMPDIR/inventory-before.txt"` is empty: the
    cut clauses carry no SHALL/MUST keyword, and the added text carries none. A non-empty diff is
    fixed by restoring the sentence, never by accepting the new inventory.

  - [x] **Step 7: Run the lint list**, fixing any hit:

```bash unverified:run after steps 2-5 land; predicts a clean run — every fenced block is tagged, no bare numeric claim is added, the required dispatch paragraphs are untouched, and review-panel.md's 45213-byte budget row has room for the roughly 2.5 KB added
scripts/check-markdown-integrity.py
scripts/check-dispatch-paragraphs.sh
scripts/check-references.sh
scripts/check-guard-symlinks.sh
scripts/check-vocabulary.sh
scripts/check-contract-budget.sh
scripts/check-stage-mark-calls.sh
```

    If `check-contract-budget.sh` reports `skills/flow/review-panel.md` over budget, that is a
    measurement contradicting this plan: stop and report per **When a measurement contradicts
    the plan** (`skills/flow-contracts/plan-provenance.md`) rather than raising the row.

- [x] 3. Mirror the reduction in `SKILL.md`'s guardrails and the `Panel:` handoff line

**Build:** green
**Files:** `skills/flow/SKILL.md`, `skills/flow/verify-and-handoff.md`
**Tests:** none
**Regression:** none
**Baseline:** before=48 after=48
<!-- predicted: scripts/run-guard-tests.sh after task 3 — unchanged, prose only -->
**Commit:** `docs(flow): mirror the docs-only panel reduction in the guardrails and handoff`

  - [x] **Step 1: Extend the guardrail bullet** in `skills/flow/SKILL.md`'s **Guardrails**.
    Replace the bullet beginning "**Never** add a slot beyond the resolved roster automatically"
    with:

```markdown unverified:re-read the cite after editing — **The docs-only reduction** must match task 2's heading verbatim for check-references.sh
- **Never** add a slot beyond the resolved roster automatically, by diff size, touched area, or any
  other trigger — only an explicit operator instruction adds one, for that run only, checked at the
  start of the panel stage and at every fix round. The one automatic change to the roster is a
  reduction — `check-panel-docs-only.sh`'s docs-only verdict dispatches `primary` alone — and it
  only ever removes; see **The docs-only reduction** (`skills/flow/review-panel.md`).
```

  - [x] **Step 2: Add the guard to the presence list** in `skills/flow/SKILL.md`'s "Check guard
    presence" paragraph: insert `check-panel-docs-only.sh`, between
    `check-panel-citation-trigger.sh` and `check-panel-diff-size.sh`, keeping the list's
    alphabetical order.

  - [x] **Step 3: Extend the `Panel:` handoff line** in `skills/flow/verify-and-handoff.md`'s
    handoff block. Replace the `**Panel:**` line with:

```markdown unverified:confirm the handoff block is a fenced ``` block whose lines are rendered verbatim, so the added field appears in the operator-facing output
**Panel:** clean — roster: <the slot list this run dispatched>; reduced: <"docs-only — " followed by the resolved slot(s) not dispatched, or "no">; substituted: <slot(s) dispatched as general-purpose in place of their own agent type, or "none">; added this run: <slot(s) an explicit operator instruction added beyond the resolved list, or "none — resolved list ran alone">
```

    In the prose paragraph below it that begins "**The `Panel:` line is `/flow`'s own**", after
    "it states what **Review panel** (`skills/flow/review-panel.md`) actually dispatched this
    run:" the enumeration becomes "the resolved roster or its docs-only reduction to `primary`
    (**The docs-only reduction**, `skills/flow/review-panel.md`), any slot substituted as a
    general-purpose agent per `unspawnable-id-substitutes`, and any slot an explicit operator
    instruction added beyond the resolved list."

  - [x] **Step 4: Run the lint list**, fixing any hit:

```bash unverified:run after steps 1-3 land; predicts a clean run — SKILL.md sits at 13859 bytes against a 16278 budget and verify-and-handoff.md at 21735 against 24862, both with room for the lines added
scripts/check-markdown-integrity.py
scripts/check-references.sh
scripts/check-guard-symlinks.sh
scripts/check-vocabulary.sh
scripts/check-contract-budget.sh
scripts/check-dispatch-paragraphs.sh
scripts/check-normative-inventory.sh
```

    The inventory is diffed against a capture taken before step 1, exactly as task 2 step 1 and
    step 6 do; the expected difference is empty.
