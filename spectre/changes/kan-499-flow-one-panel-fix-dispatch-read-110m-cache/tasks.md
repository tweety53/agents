# kan-499-flow-one-panel-fix-dispatch-read-110m-cache — Implementation Plan

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** bound each `/flow` panel-fix dispatch to at most 10 findings so a finding-heavy round
can no longer balloon one subagent's context.

**Architecture:** the fix-dispatch contract in `skills/flow/review-panel.md` moves from one
dispatch per round carrying every open finding to sequential per-chunk dispatches (first chunk
keeps key `panel-fix-<round>`, later chunks `panel-fix-<round>-<n>` contiguous from 2);
`scripts/check-panel-fix-single-dispatch.sh` learns the chunked shape, per-chunk retry rules, and
a chunk-count bound derived from the findings store (`ceil(findings raised in earlier rounds / 10)`)
so per-finding dispatch stays a violation; `skills/flow/implement.md`'s dispatch-site closed list
row follows.

**Tech Stack:** Bash (guard + harness, `/bin/bash` 3.2 — no associative arrays), Markdown
contracts, `jq` for store reads.

## Global Constraints

- Chunk cap: at most 10 findings per panel-fix dispatch — KAN-499's number.
  <!-- predicted: KAN-499 states "around 10"; the panel record shows each chunk's finding list -->
- The guard never reads a rendered document — only `flow record dispatches` and, newly,
  `flow record findings` (the same verb `check-panel-findings-closed.sh` reads).
- A findings-store read that fails is exit 2 ("cannot answer"), never a clean verdict.
- `/bin/bash` is 3.2: parallel indexed arrays, never associative arrays, in the guard and harness.
- Bash for this repo is `shellcheck`-free by history; the guard's own harness is its test suite.

---

- [x] 1. Guard tests for the chunked fix-dispatch contract

**Build:** red
**Squash-with:** Task 2
**Files:**
- Modify: `scripts/test-check-panel-fix-single-dispatch.sh`
**Tests:** `chunked-legal-round-passes`, `chunk-retry-pair-passes`, `legacy-round-with-findings-verb-passes`, `chunk-without-bare-key-fails`, `chunk-gap-fails`, `chunk-count-over-bound-fails`, `chunk-retry-alone-fails`, `two-originals-on-chunk-fails`, `invented-chunk-key-shape-fails`, `findings-store-unreachable-exits-2`
**Regression:** reverting this commit silences the added cases — a chunked round would be rejected
(or an over-bound one accepted) with no failing test naming it.
**Baseline:** before=10 after=20
<!-- measured: bash scripts/test-check-panel-fix-single-dispatch.sh @ main dd4e2f9 -->
<!-- measured: bash scripts/test-check-panel-fix-single-dispatch.sh @ branch spectre/kan-499-flow-one-panel-fix-dispatch-read-110m-cache, task 1 red run -->
**Commit:** fix(scripts): accept and bound chunked panel-fix dispatch keys

**Interfaces:**
- Consumes: the harness's existing sandbox + stub-`flow` helpers (dispatch_json, run_case shape).
- Produces: a stub `flow` that answers `record findings` with a per-case canned JSON array
  (default `[]`), so task 2's guard can read both verbs in every case.

  - [ ] **Step 1: Extend the stub `flow` to answer `record findings`**

The stub currently matches only `record dispatches`. Add a third canned payload variable
(`FINDINGS_JSON`, default `[]`) and a branch for `record findings -change <name>` that prints it,
mirroring the existing `dispatches` branch. Keep every existing case working unchanged — they now
exercise the findings read implicitly with `[]`.

```bash unverified:run the full harness before committing — every existing case must stay ok
# inside the stub's case dispatch, beside the existing record-dispatches branch:
record\ findings*)
  printf '%s\n' "${FINDINGS_JSON:-[]}"
  exit 0
  ;;
```

  - [ ] **Step 2: Add the legal-shape and bound cases**

Each case builds its dispatch rows via the existing `dispatch_json` helper and, where the case
exercises the bound, sets `FINDINGS_JSON` to findings raised in earlier rounds. Round 0 carries
the initial findings; the round-1 fix dispatches chunks over them.

```bash unverified:confirm jq parses the canned arrays exactly as the real store prints them
# chunked-legal-round-passes: 24 findings raised in round 0, round 1 chunks 24 into 3 dispatches
FINDINGS_JSON='[{"ref":"F1","round":0,"status":"fixed"},{"ref":"F2","round":0,"status":"fixed"}]'
rows: panel-fix-1, panel-fix-1-2, panel-fix-1-3   → expect exit 0   (3 ≤ ceil(24/10)=3)

# chunk-retry-pair-passes: chunk 2 handshake-failed once, retried clean
rows: panel-fix-1, panel-fix-1-2, panel-fix-1-2-retry → expect exit 0 (prior findings ≥ 20)

# legacy-round-with-findings-verb-passes: single dispatch, empty findings — today's shape, unchanged
rows: panel-fix-1 (+ panel-fix-1-retry) → expect exit 0
```

  - [ ] **Step 3: Add the violation cases**

```bash unverified:exit codes only — the harness asserts status, not stderr wording
# chunk-without-bare-key-fails:  panel-fix-1-2 with no panel-fix-1          → exit 1
# chunk-gap-fails:               panel-fix-1, panel-fix-1-2, panel-fix-1-4  → exit 1
# chunk-count-over-bound-fails:  panel-fix-1, panel-fix-1-2, panel-fix-1-3 with only 15 round-0
#                                findings (bound ceil(15/10)=2)             → exit 1
# chunk-retry-alone-fails:       panel-fix-1-2-retry, no panel-fix-1-2      → exit 1
# two-originals-on-chunk-fails:  panel-fix-1-2 twice                        → exit 1
# invented-chunk-key-shape-fails: panel-fix-1-f2 (one per finding, KAN-482) → exit 1
# findings-store-unreachable-exits-2: stub exits non-zero on `record findings` → exit 2
```

  - [ ] **Step 4: Run the harness — expect the red drivers to fail**

Run: `bash scripts/test-check-panel-fix-single-dispatch.sh`
Expected: chunked-legal-round-passes, chunk-retry-pair-passes and
findings-store-unreachable-exits-2 FAIL (today's guard rejects any second dispatch and never
reads findings); every legacy case and every violation case still passes (violation cases expect
exit 1, which today's guard produces for the coarser one-dispatch reason). Overall FAIL is the
expected RED.

  - [ ] **Step 5: No commit — the red commit folds into task 2's**

This task's work lands in task 2's commit via the `**Squash-with:**` pairing.

---

- [x] 2. Guard accepts and bounds the chunked panel-fix key shape

**Build:** green
**Files:**
- Modify: `scripts/check-panel-fix-single-dispatch.sh`
**Tests:** the full `test-check-panel-fix-single-dispatch.sh` harness (all cases from task 1)
**Regression:** reverting reintroduces the one-dispatch-only rule — a legal chunked round is
rejected at panel close, and the findings-store outage reads as a clean verdict again.
**Baseline:** before=10 after=20
<!-- measured: bash scripts/test-check-panel-fix-single-dispatch.sh @ main dd4e2f9 -->
<!-- measured: task 1's red run at the worktree branch names the same case set -->
**Commit:** fix(scripts): accept and bound chunked panel-fix dispatch keys
**After:** Task 1

**Interfaces:**
- Consumes: `flow record dispatches -change <name> -C <worktree>` (unchanged) and
  `flow record findings -change <name> -C <worktree>` — rows carry at least `ref`, `round`
  (int), `status` (stats/internal/records/types.go's `Finding`).
- Produces: exit semantics unchanged (0 clean / 1 violations named on stderr / 2 cannot-answer);
  new canonical key shape accepted, documented in the header comment.

  - [ ] **Step 1: Widen the key regex and parse the chunk**

```bash unverified:diff against task 1's failing cases — every case green before committing
CANONICAL_KEY_RE='^panel-fix-[0-9]+(-[0-9]+)?(-retry)?$'
```

Parse each row's key into round (first integer), chunk (second integer, or `0` when the bare
`panel-fix-<round>` form), and retry flag. Group rows by the full base (`panel-fix-<round>` or
`panel-fix-<round>-<n>`, retry stripped) in the existing parallel-array style; the per-base checks
(retry never alone, ≤1 original, ≤2 total with exactly one retry) are unchanged and now apply per
chunk key, not per round.

  - [ ] **Step 2: Read the findings store once, derive the per-round bound**

Before the row loop (or after it, before the round analysis), call
`flow record findings -change "$NAME" -C "$WORKTREE"`; on failure exit 2 with the same
cannot-answer posture as the dispatches read. For each round R that owns panel-fix rows, the
chunk ceiling is `ceil(prior/10)` where `prior` is the count of findings whose `round` is
strictly less than R. A round whose prior is 0 may carry no chunks at all (bound 0).

```bash unverified:confirm the jq filter against a real store read once the daemon is up
prior="$(printf '%s' "$FINDINGS_JSON" | jq --argjson r "$round" \
  '[.[] | select((.round // 0) < $r)] | length')"
```

  - [ ] **Step 3: Add the per-round chunk-shape checks**

For each round that owns at least one chunked base (`panel-fix-<round>-<n>`):

```bash unverified:authored for this change — the harness cases from task 1 are its proof
# 1. the bare key must exist with at least one original dispatch;
# 2. the chunk numbers of the original dispatches must be exactly 2..k, contiguous, no gaps;
# 3. k must be ≤ the round's bound (ceil(prior/10)) — over the bound is the per-finding abuse.
```

Each violation names the round, what was found, and the rule, in the existing
`check-panel-fix-single-dispatch: <violation>` stderr shape.

  - [ ] **Step 4: Update the header comment**

The header's key-shape and count paragraphs describe the new canonical shape, the per-chunk retry
rule, the contiguity requirement, and the findings-store bound (KAN-499) — including why the
bound exists (a well-formed per-finding key sequence must stay caught, KAN-482).

  - [ ] **Step 5: Run the harness — expect full green**

Run: `bash scripts/test-check-panel-fix-single-dispatch.sh`
Expected: PASS, 20 ok cases
<!-- measured: task 1's red run carried 20 cases; task 2's green run asserts the same count -->

  - [ ] **Step 6: Commit**

```bash verified:git -C <worktree> log --oneline confirms subject and folding
git add scripts/check-panel-fix-single-dispatch.sh scripts/test-check-panel-fix-single-dispatch.sh
git commit -m "fix(scripts): accept and bound chunked panel-fix dispatch keys"
```

  - [ ] **Step 7: Verify step**

Run: `bash -n scripts/check-panel-fix-single-dispatch.sh && bash scripts/test-check-panel-fix-single-dispatch.sh`
Expected: syntax clean, harness PASS.

---

- [x] 3. Contract prose — review-panel.md fix step and implement.md closed-list row

**Build:** green
**Files:**
- Modify: `skills/flow/review-panel.md` (the fix step, ~line 924–956: the REPORT FILE paragraph,
  the give-the-findings paragraph, the key-shape sentence; the close-guard handback wording,
  ~line 1000)
- Modify: `skills/flow/implement.md` (the dispatch-site closed-list table row, line 52)
**Tests:** none — prose only
**Regression:** reverting leaves the guard accepting a shape no contract states — the prose and
the guard would disagree about what a legal fix round is.
**Baseline:** before=0 after=0
<!-- measured: prose-only task — no test counts move -->
**Commit:** docs(flow): chunk the panel-fix dispatch contract at 10 findings
**After:** Task 2

  - [ ] **Step 1: Rewrite the give-the-findings paragraph**

The paragraph beginning "Give the surviving findings to **one** fix subagent" (review-panel.md,
~line 933) becomes the chunked contract, keeping every existing constraint it generalizes: at
most 10 findings per dispatch; chunks dispatched sequentially, each awaited in the foreground
before the next begins; never one dispatch per reviewer, slot, or finding; no fix subagent left
in flight at turn end. Key rule: the round's first chunk is keyed `panel-fix-<round>`; each
further chunk appends `-<n>` (`panel-fix-<round>-2`, `-3`, …), contiguous from 2; the handshake
retry suffixes `-retry` onto the chunk's own key. The number of chunks a round may carry is
bounded — `check-panel-fix-single-dispatch.sh` refuses a round whose chunk count exceeds
`ceil(findings raised in earlier rounds / 10)` (KAN-499).

  - [ ] **Step 2: Rewrite the REPORT FILE and recording paragraphs**

`panel-fix-report-<round>.md` becomes one report per dispatch: the bare key's chunk writes
`panel-fix-report-<round>.md`; chunk `n` writes `panel-fix-report-<round>-<n>.md`. The dispatcher
waits on each dispatch's own file. The `flow record dispatch` block's `-key` examples gain the
chunked form.

  - [ ] **Step 3: Rewrite the close-guard handback wording**

The `## Question` block after `check-panel-fix-single-dispatch.sh`'s exit 1 (~line 1000) describes
the violation set the guard now reports: over-bound chunk counts, non-contiguous chunks, a chunked
round with no bare key, per-chunk count/retry violations — not only "more than one panel-fix
dispatch".

  - [ ] **Step 4: Update implement.md's closed-list row**

Line 52 of `skills/flow/implement.md`:

```markdown verified:read at the time this plan was written; re-read before editing
| panel-fix, one per chunk of at most 10 findings | `panel-fix` | `panel-fix-<round>[-<chunk>]` (`-retry` once per chunk) | `skills/flow/review-panel.md`, the fix step |
```

  - [ ] **Step 5: Verify step**

Run: `scripts/check-vocabulary.sh skills/flow/review-panel.md skills/flow/implement.md && scripts/check-references.sh`
Expected: both clean.

  - [ ] **Step 6: Commit**

```bash verified:git -C <worktree> log --oneline confirms subject
git add skills/flow/review-panel.md skills/flow/implement.md
git commit -m "docs(flow): chunk the panel-fix dispatch contract at 10 findings"
```
