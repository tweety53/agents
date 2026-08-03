# Tasks — kan-23-myflow-self-review

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Repository:** `/Users/tweety53/Projects/agents` — a documentation-and-guards repository. Most
tasks below edit Markdown contracts/skills; one task adds a new Bash script plus its test harness.
There is no compiled artifact and no application to run. Verification is the three lint guards named
in `.myflow/project.md` (`check-vocabulary.sh`, `check-references.sh`, `check-plan-provenance.sh`),
targeted `grep` assertions, and the new script's own test harness.

**Order:** the script and its test (group 1) have no dependency on the contract edits and can be
built first; the contract edits (group 2) define the invocation the script is called from, so group
2 cites the script's usage line rather than guessing it; the skill edit (group 3) implements what
groups 1-2 specify; group 4 is bookkeeping (the project's test-command list).

**Global constraints, carried into every dispatch:**

- **No suppression markers, ever.** Fix the offending line; never silence a guard.
- **Citation over copy.** Where a rule already lives in `pipeline.md` or `jira-integration.md`, cite
  it with a bold token plus its backticked path rather than restating it, so
  `scripts/check-references.sh` checks the pointer.
- Every path written into a contract, skill or handoff is **absolute** where it names something the
  operator opens, exactly as the rest of this pipeline's output already is.
- Self-review must never be able to block, delay, or undo the `FINISHED` write — every task below
  that touches step 8 must preserve that ordering.

---

## 1. `scripts/gather-self-review-context.sh` and its test

### 1.1 Write the script

- [x] Create `scripts/gather-self-review-context.sh`, modeled on the existing
      `scripts/preserve-session-records.sh` (argument-driven, no reliance on the caller's cwd,
      `set -euo pipefail`, one clearly documented usage line, a header comment stating the contract
      it implements).
- [x] Usage: `gather-self-review-context.sh <archived-change-path> <name> <state-dir>`. Accept the
      **archived** change path (`openspec/changes/archive/<date>-<name>/`) rather than a worktree
      path — run 2 has already removed the worktree by the time step 8 runs (Requirement: "The
      script has no worktree to run against", `design.md` → Risks / Trade-offs).
- [x] Collect four sources, each independently, none blocking on another:
      1. `docs/superpowers/ledgers/<YYYY-MM-DD>-<name>.md` — date-prefixed, exactly as
         `preserve-session-records.sh` writes it; located via the same digit-anchored search that
         script uses for an existing destination, not a literal `<name>.md` path.
      2. `docs/superpowers/reviews/<YYYY-MM-DD>-<name>-panel.md` — located the same way.
      3. `<archived-change-path>/tasks.md`
      4. `git log --stat` for the change's two finish-run-1 commits plus the archive commit — resolve
         these by walking `git log --oneline` on the base branch for commits matching
         `^(feat|fix|chore)\(<name>\): ` (implementation and planning) and
         `^chore\(<name>\): .*archive` (the archive commit — real convention confirmed by grepping
         `git log --oneline -- openspec/changes/archive/`: word order varies, e.g. "sync delta specs
         and archive the change" vs. "archive change and sync delta specs", but the commit is always
         scoped `chore(<name>): ...` and always contains "archive") rather than a fixed commit count
         back from `HEAD`.
- [x] A missing source (1-3) is **not fatal**: print `skipped: <src> (absent)` on stdout — matching
      `preserve-session-records.sh`'s own `skipped:`/`preserved:` vocabulary so a reader of both
      scripts' output recognizes the pattern — and continue gathering the rest.
- [x] Print the gathered bundle to stdout as one document: a header line naming which sources were
      found vs. skipped, then each found source's content under its own subheading.
- [x] The script exits 2 on an invocation error (a missing argument or an invalid change name);
      otherwise it always exits 0 — a missing source is never fatal, and the script performs no
      judgement and reports no pass/fail otherwise, per **Requirement: Context is gathered
      deterministically** (`specs/myflow-self-review/spec.md`).
- [x] Apply the same change-name allowlist `preserve-session-records.sh` already applies (one
      leading alphanumeric, then letters/digits/`.`/`_`/`-`) before interpolating `<name>` into any
      path, for the same reason that script states it: a name containing `/` or a glob
      metacharacter must not be used to build a path.

For the script's actual header and usage comment, read `scripts/gather-self-review-context.sh`
itself rather than a copy here: across the fix rounds recorded in
`.superpowers/sdd/final-review-panel.md` (F4, F12, F13, F14) the header grew substantially past what
any snippet frozen at plan-writing time could stay byte-identical with, and a plan snippet that has
to be kept in lockstep with actively-evolving code across every future fix round is the wrong
contract — the script's own header is the single source of truth for its documented behavior.

**Verify:**

```bash verified:ran against the implemented script; bash -n clean, executable bit set
bash -n scripts/gather-self-review-context.sh
chmod +x scripts/gather-self-review-context.sh
```

### 1.2 Write `scripts/test-gather-self-review-context.sh`

- [x] Model the harness on `scripts/test-preserve-session-records.sh`: sandboxed trees under
      `mktemp -d "${TMPDIR:-/tmp}/..."`, an indexed `TREES` array with a `cleanup` trap (never a
      space-separated string — the same bash-3.2-and-spaces-in-`mktemp`-paths hazard that script's
      header documents), `fail`/`pass` helpers, a `FAILURES` counter, exit non-zero at the end iff
      `FAILURES > 0`.
- [x] Assert against the stated contract — **Requirement: Context is gathered deterministically**
      (`specs/myflow-self-review/spec.md`) — never against whatever the script happens to print;
      `test-check-plan-provenance.sh`'s own header (see `scripts/test-check-plan-provenance.sh`)
      records this exact mistake happening before and is why this instruction is repeated in every
      new test harness in this repository.
- [x] Cases to cover:
      - all four sources present → bundle contains all four, script exits 0
      - ledger absent, others present → skipped, reported, script still exits 0, other three
        sources still present in the bundle
      - panel record absent (the common case — a change with no optional review slot) → same
        skip-and-continue behavior
      - `tasks.md` absent from the archived path → same skip-and-continue behavior
      - all four sources absent → the bundle reports all four as skipped, script still exits 0
      - a change name containing `/` or `*` is rejected before any path is touched (mirrors
        `test-preserve-session-records.sh`'s own path-injection case)
- [x] Register the new test in `.myflow/project.md`'s `## test` list (task 4.1 below) so it runs
      wherever this repository's tests run.

**Verify:**

```bash verified:scripts/test-gather-self-review-context.sh @ branch openspec/kan-23-myflow-self-review — 90 assertions, all ok, exit 0 (post-F27: validate_archived_path() resolves the trusted repo root via git-common-dir, not show-toplevel; see scripts/gather-self-review-context.sh header for the full mechanism and its review history)
bash -n scripts/test-gather-self-review-context.sh
scripts/test-gather-self-review-context.sh
```

Expect every case to print `ok: <case>` and the script to exit 0.

---

## 2. Contract edits — `skills/myflow-contracts/pipeline.md`

### 2.1 Level 1 stage table — `/myflow-finish` row

- [x] In **Level 1 — the stages of each command**, the `/myflow-finish` row's run-2 stage list
      currently ends `… → cleanup ▸ → verify the cleanup → write FINISHED`. Add `→ self-review ▸`
      after `write FINISHED`, so the row reads `… → verify the cleanup → write FINISHED →
      self-review ▸`.
- [x] Leave the row's "Gate after it" column text about run 2 (currently: "after run 2, nothing —
      the state is terminal") unchanged: self-review runs after the terminal state is already
      written, so it does not add a new human gate to this table's own claim about what happens
      after run 2.

**Verify:**

```bash verified:grep run against this tree; the Level 1 row edit is present at pipeline.md:79
grep -n "self-review" skills/myflow-contracts/pipeline.md
```

Expect at least the Level 1 table row edit to appear.

### 2.2 Level 2 expansion — new "Self-review" subsection

- [x] Add a new **Level 2** expansion, `#### Self-review — /myflow-finish run 2`, placed
      immediately after the existing `#### Cleanup — /myflow-finish run 2` subsection (so the two
      run-2 expansions sit together, matching run 2's own step order).
- [x] State the **structure** only, per this section's own stated rule ("Each expansion states the
      structure … and cites the file that owns the tuned values"): self-review runs only after
      `FINISHED`, is skippable per-run (default: run it), gathers context via the new script rather
      than re-reading files inline, runs one combined reasoning pass covering all four angles from
      KAN-23, offers a per-finding Jira filing ask, asks the operator for a 1-5 rating, and commits a
      report to `docs/self-review/<name>-self-review.md`.
- [x] Cite `skills/myflow-self-review/SKILL.md`'s own detail is unnecessary here — self-review has
      no dedicated skill file, it is a step inside `skills/myflow-finish/SKILL.md`; cite instead
      **Requirement: Self-review runs only after FINISHED is written** and the other
      `myflow-self-review` requirements in `openspec/specs/myflow-self-review/spec.md` (present only
      once this change reaches archive — until then this is a forward reference exactly as other
      Level 2 expansions in this file already make to specs not yet landed, e.g. the Planning-effort
      expansion's reference above it).

**Verify:**

```bash verified:grep and guard run against this tree; the expansion is present at pipeline.md:231, check-references.sh exits 0
grep -n "#### Self-review" skills/myflow-contracts/pipeline.md
scripts/check-references.sh
```

### 2.3 Finish contract — run 2 outline gains step 8

- [x] In **Run 2 — the branch is merged** (inside the Finish contract section), the numbered outline
      currently ends at step 7 ("Write `FINISHED`"). Add step 8: "Run self-review — see **Self-review
      — /myflow-finish run 2** above — after `FINISHED` is written; a skip, a failure, or a decline
      never moves the change off `FINISHED`."
- [x] Do not renumber the Jira `Done` transition paragraph that currently follows step 7's
      description; state explicitly that the Done transition still fires immediately after the state
      write, **before** step 8 runs — self-review must not be able to delay the Jira transition
      either.

**Verify:**

```bash verified:grep run against this tree; step 8 is present at pipeline.md:1060
grep -n "Run self-review" skills/myflow-contracts/pipeline.md
```

### 2.4 Terminal handoff template gains the `Self-review` line

- [x] In the `## Finished` block template (the fenced block immediately under **Run 2 — archive and
      clean up** → the terminal-block section), add one line after `**Cleanup:** verified`:
      ``**Self-review:** <path> (rating: <n>/5) | skipped``.
- [x] In the `## Cleanup incomplete — not finished` template (the leftover-stop block), add **no**
      `Self-review` line — that block is printed only when run 2 stops *before* step 7, so
      self-review never ran; adding the field there would misstate a run that never reached it.
- [x] Add one sentence to the surrounding prose stating why the field is absent from the
      leftover-stop block specifically (mirrors the existing "run-only" reasoning already used
      elsewhere in **The block each state renders**), so a later editor does not "fix" the omission
      by copying the line into both templates.

**Verify:**

```bash verified:grep run against this tree; exactly one occurrence, at pipeline.md:723, in the Finished-block prose and not in the leftover-stop template
grep -n "Self-review:" skills/myflow-contracts/pipeline.md
```

Expect exactly one occurrence, in the `## Finished` template, not in the leftover-stop template.

---

## 3. `skills/myflow-finish/SKILL.md` — implement step 8

### 3.1 Add `## 1.6`-equivalent step under Run 2, after 6 ("Verify cleanup") and after "Write
      FINISHED"

- [x] The existing Run 2 outline in this file states steps 1-7 as bullet items ("In outline, and
      stopping at the first step that fails: 1. … 7. **Write FINISHED**"). Add item 8: **Run
      self-review.** Point at **Self-review — /myflow-finish run 2**
      (`skills/myflow-contracts/pipeline.md`) as canonical for the procedure, following this file's
      own established pattern of pointing at `pipeline.md` rather than restating steps
      (`skills/myflow-finish/SKILL.md`'s own "Deciding which run this is" section already states this
      convention: "This skill carries only what is specific to *executing* it").
- [x] State what is specific to *executing* it here (mirroring how step 6's cleanup-verification
      bullet states its own specifics): the script invocation
      `scripts/gather-self-review-context.sh <archived-change-path> <name> <state-dir>`, resolving
      `<archived-change-path>` as `openspec/changes/archive/<YYYY-MM-DD>-<name>/` using the same date
      step 2 (sync + archive) already used when it moved the change there.
- [x] State the skip-prompt wording verbatim (matches
      **Requirement: A skip prompt defaults to running self-review**,
      `openspec/changes/kan-23-myflow-self-review/specs/myflow-self-review/spec.md`):

  ```markdown verified:copied verbatim from the approved design and the delta spec written earlier in this change
  > **Run self-review for this change?**
  > - **Yes — run it** *(default, recommended)*
  > - **No — skip**
  ```

- [x] State the per-finding filing-ask wording verbatim (same source):

  ```markdown verified:copied verbatim from the approved design and the delta spec written earlier in this change
  > **File `<one-line finding>` as a Jira issue?**
  > - **No — don't file** *(default, recommended)*
  > - **Yes — file it**
  ```

- [x] State explicitly that the reasoning step is **one combined pass** answering all four angles
      (problems/fixes, cost, what-went-well, automation candidates) plus the rating request — never
      four separate dispatches — matching **Requirement: One combined reasoning pass, not four
      separate dispatches** (`specs/myflow-self-review/spec.md`). Name the risk this guards against:
      a self-review mechanism that itself burns disproportionate tokens would be its own finding
      under angle #2.
- [x] State the rating ask verbatim: `**Rate this myflow run, 1 (rough) to 5 (excellent):**`.
- [x] State the report path verbatim: `docs/self-review/<name>-self-review.md`, and that it is
      committed and pushed on the base branch in the main checkout (not the removed worktree) — as
      one guarded commit, using the same skip-if-empty-but-never-silent-on-failure pattern already
      documented under **Git boundaries** (`skills/myflow-contracts/pipeline.md`):

  ```bash verified:matches skills/myflow-finish/SKILL.md's implemented commit+push snippet, including the M2 fix (commit && push grouped together so push never fires on an empty diff)
  git -C <main-checkout> add -- docs/self-review/<name>-self-review.md \
    && { git -C <main-checkout> diff --cached --quiet \
         || { git -C <main-checkout> commit -m "docs(<name>): self-review report" \
              && git -C <main-checkout> push; }; }
  ```

  A commit that FAILS (hook rejection, push rejected) is reported with git's own output; the change
  stays `FINISHED` regardless, per **Requirement: Self-review runs only after FINISHED is written**
  (`specs/myflow-self-review/spec.md`) — a report that failed to commit is a self-review failure to
  report in the handoff, never a reason to reopen the change.
- [x] Cross-reference **Labels on issues the pipeline creates** and **Never blocking**
      (`skills/myflow-contracts/jira-integration.md`) for the filing ask's labelling and failure
      handling, rather than restating either rule — matches this file's own established citation
      style for every other Jira interaction it already documents.

### 3.2 Add the new handoff field to this file's own `## Finished` block copy

- [x] This file's `## Finished` block (currently ending `**Cleanup:** verified`) is this command's
      own rendering of the template defined in `pipeline.md`. Add the same
      `**Self-review:** <path> (rating: <n>/5) | skipped` line here, immediately after
      `**Cleanup:** verified`, matching what task 2.4 added to the canonical template — this file's
      own prose already states the rule this has to follow: "The block below is **not** a second
      definition of the handoff … Change the template first and bring this block with it."

### 3.3 Guardrails list

- [x] Add to this file's **Guardrails** section: "**Never** let self-review block, delay, or undo the
      `FINISHED` write — it runs only after that write succeeds, and a failure or a skip inside it
      never moves the change off `FINISHED`."
- [x] Add: "**Never** ask the self-review skip prompt, a per-finding filing ask, or the rating
      question before `FINISHED` has been written."

**Verify:**

```bash verified:grep run against this tree — 13 occurrences in skills/myflow-finish/SKILL.md; both guards exit 0
grep -n "self-review\|Self-review" skills/myflow-finish/SKILL.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

---

## 4. Project bookkeeping

### 4.1 `.myflow/project.md` — register the new test

- [x] Add `scripts/test-gather-self-review-context.sh` to the `## test` fenced block's list,
      alongside the existing seven test commands, in the same style (one command per line, no
      surrounding prose changes needed).

**Verify:**

```bash verified:grep run against this tree; line 46 of .myflow/project.md
grep -n "test-gather-self-review-context.sh" .myflow/project.md
```

---

## 5. Full-repository verification

- [x] Run every lint guard named in `.myflow/project.md`'s `## lint` list.
- [x] Run every test named in `.myflow/project.md`'s `## test` list, including the new one added in
      task 4.1.

```bash verified:ran on branch openspec/kan-23-myflow-self-review — all three lint guards and all eight test scripts exited 0
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
scripts/test-gather-self-review-context.sh
```

Expect every guard to exit 0 and every test script to report all cases `ok:` with no `FAIL:` lines.
