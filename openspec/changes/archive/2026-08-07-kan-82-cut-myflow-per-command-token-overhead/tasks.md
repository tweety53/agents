# Tasks — kan-82-cut-myflow-per-command-token-overhead

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** cut what every `/myflow-*` command loads before it does any work, by splitting the two
largest contract files into a normative core and a rationale appendix, moving the one section that
only `/myflow-finish` needs into its own file, and adding a byte-budget guard so the cores cannot
grow back.

**Architecture:** a **strict partition**. Every sentence of the two files lands in exactly one of
the resulting five, byte-for-byte unchanged; the only permitted edits are repointing a citation whose
target moved and deleting a position word the move made false — the two the global constraints below
state in full. Nothing is authored, so the split cannot produce two independently authored statements
of one rule — the failure the linked issue names as its constraint. Appendices carry the same
heading tree as their cores, so a citation resolves against whichever file its path names.

**Repository:** one — `/Users/tweety53/Projects/agents`. It has no runnable application; verification
is the guard scripts declared in `.myflow/project.md` plus targeted assertions written into the tasks
below.

**Order:** group 1 is a pure file move and needs no partition judgment, so it lands first and proves
the citation-repointing loop. Group 2 partitions the smaller of the two files, proving the partition
mechanics and the line-multiset check. Group 3 applies both to the large file. Group 4 updates the
index, the rule layer and the two entry-point files. Group 5 writes the budget guard **last**,
because its numbers are the sizes groups 1-3 produced.

**Global constraints, carried into every dispatch:**

- **This change moves text. It does not change a rule.** No requirement is added, removed or
  reworded; no `/myflow-*` command behaves differently afterwards. A dispatch that finds itself
  improving a rule's wording has left this plan's scope and must stop and report.
- **A moved line keeps its heading level, its wording and its wrapping.** **Two** edits are permitted
  to a moved line, and no others — not a heading level, not a typo, not a rewording:
  1. **repointing a citation** — its backticked path and the bold token beside it; and
  2. **deleting a position word** — `above` or `below` — that the move made false, because the thing
     it locates is now in a different file.

  Only a **new** file's own title and header lines are authored, and only an appendix's mirrored
  headings are written fresh. Where a task and this constraint appear to disagree, this constraint
  and the delta spec behind it win, and the disagreement is reported rather than resolved silently.

  **The second edit is a deletion, never a substitution.** `above` is removed; it does not become
  "in `jira-integration.md`". A deletion cannot alter a rule — it removes a locator that is now
  false and leaves the citation's path to do the locating, which is this corpus's existing
  convention for a cross-file citation: `**State file** (`skills/myflow-contracts/state-file.md`)`
  carries no position word at all.

  **This reverses an earlier decision, and the reversal is recorded rather than quiet.** Permitting
  only path repointing was chosen during planning to keep the diff a pure move. The review panel
  found what that cost: a split moves passages across a file boundary that did not exist before, so
  roughly a dozen `above`/`below` words were left pointing across it, and one core term — "the
  pre-check" — was left with its defining paragraph in a file no run loads. Those are not slips; they
  are what the two rules produce together. A deleted position word still shows as a `<`/`>` pair in
  the line-multiset check, so it is reviewed exactly as a repointed citation is.
- **The unit that moves is a whole blank-line-delimited paragraph, never part of one.** A paragraph
  mixing a rule with its justification **stays in the core, whole**. Do not split it, and do not
  re-wrap the survivors to make it fit.

  **This is a consequence of the verification, and it is accepted rather than worked around.** These
  files are hard-wrapped at roughly 100 characters, so a sentence boundary almost never lands on a
  line break; splitting mid-paragraph means re-wrapping the lines that stay, and a re-wrapped line is
  indistinguishable from a lost one in the line-multiset check that proves nothing was dropped.
  Keeping the check trustworthy is worth more than the extra bytes a sentence-level split would move
  — the check is the only thing standing between this change and silent rule loss across the whole
  corpus being partitioned.
  The cost is that rationale welded into a rule's own sentence group stays in the core; `pipeline.md`
  carries 64% of its bytes as whole pure-prose paragraphs, so the cost is small there and was large
  only in `jira-integration.md`, which is unusually rule-dense.
- **No suppression markers, ever.** Fix the offending line; never silence a guard, never weaken a
  guard's configuration or thresholds. There is no auto-fix command in this repository.
- **`scripts/check-references.sh` is line-scoped**: it checks a citation only when the bold token and
  the backticked path sit on the **same physical line**. A citation split across two lines by prose
  wrapping passes unchecked. Every task that writes or repoints a citation keeps the token and the
  path on one line, and verifies by **mutating the bold token in a scratch copy and confirming the
  guard fails** — a passing guard is not evidence it looked.
- Every fenced block added to any change's `tasks.md`, `design.md` or `proposal.md` carries
  `verified:`/`unverified:` on its info string, and every numeric claim carries a
  `measured:`/`predicted:` comment within two lines, per **The four tags**
  (`skills/myflow-contracts/plan-provenance.md`).
- Every task carries exactly one `**Build:**` tag, per **The build-green tag**
  (`skills/myflow-contracts/build-green.md`).
- **Every path in a command is absolute.** The apply worktree, not the main checkout, holds this
  work; resolve it once from `git -C /Users/tweety53/Projects/agents worktree list` and use the path
  it prints. `<WT>` below stands for that absolute path.

**Baseline, measured before any task runs:**

```text verified:authored in-tree from the wc output below
skills/myflow-contracts/pipeline.md            1540 lines   103326 bytes
skills/myflow-contracts/jira-integration.md     732 lines    52783 bytes
skills/myflow-contracts/project-configuration.md 542 lines   45236 bytes
skills/myflow-contracts/workspace-isolation.md  417 lines    31079 bytes
skills/myflow-contracts/plan-provenance.md      363 lines    24466 bytes
skills/myflow-contracts/state-file.md           236 lines    15951 bytes
skills/myflow-contracts/state-self-heal.md      188 lines    15236 bytes
skills/myflow-contracts/build-green.md           58 lines     3646 bytes
skills/myflow-contracts/SKILL.md                 41 lines     3062 bytes
```

<!-- measured: wc -l -c skills/myflow-contracts/*.md @ adedf66 -->

---

## 1. Move the follow-up contract into its own file

### 1.1 Move `### Follow-up issues` whole into `skills/myflow-contracts/jira-followups.md`

**Files:**
- Create: `<WT>/skills/myflow-contracts/jira-followups.md`
- Modify: `<WT>/skills/myflow-contracts/jira-integration.md`

**Interfaces:**
- Produces: the file `skills/myflow-contracts/jira-followups.md` and the headings later tasks and
  `/myflow-finish` cite — `# Follow-up issues`, and beneath it the `###` headings the moved section
  already carries.

`### Follow-up issues` is the **last** section of `jira-integration.md`: it opens at line 276 and runs
to line 732, 457 lines and 35467 bytes — 67% of the file.

<!-- measured: awk '/^### Follow-up issues/{f=1} f{n++; b+=length($0)+1} END{print n, b}' skills/myflow-contracts/jira-integration.md @ adedf66 -->

It is reachable only from `/myflow-finish` run 1's unfinished-work gate. Moving it whole — its rules
**and** its reasoning together — is what stops `/myflow-start` and `/myflow-do` paying for it.

- [x] Cut lines 276 to the end of `<WT>/skills/myflow-contracts/jira-integration.md` into the new
      file `<WT>/skills/myflow-contracts/jira-followups.md`, **byte for byte**. Do not reword, do not
      reflow, do not renumber.

- [x] Give the new file a title and a canonical-authority header in the same shape the other contract
      files use — `# Follow-up issues`, then a bolded sentence naming it canonical for follow-up
      issues and stating that where it and a skill disagree, this file wins. **Leave every moved
      heading at its original level** and let the new `#` title sit above them.

      **This bullet previously offered to promote the moved `### Follow-up issues` heading to `##`,
      and that clause is struck** — it authorised an edit the delta spec forbids. **The split is a
      verbatim partition, with citation repointing as its only edit**
      (`openspec/changes/kan-82-cut-myflow-per-command-token-overhead/specs/myflow-contract-economy/spec.md`)
      permits exactly one change to a moved line, and a heading-level change is not it. The
      implementer and the task reviewer reached opposite conclusions from the struck clause, which is
      what surfaced the contradiction; the spec is what settles it. The cost is accepted: this file
      goes from its `#` title straight to an `###`, where every sibling contract goes title → `##`.

- [x] Add, immediately under the header, one sentence stating **which command loads this file**:
      `/myflow-finish` run 1 and no other. That sentence is the file's whole reason for existing, and
      a reader who does not know it will fold the file back into `jira-integration.md`.

- [x] In `<WT>/skills/myflow-contracts/jira-integration.md`, replace the removed section with a
      **pointer stub**: a `### Follow-up issues` heading, one sentence naming what the contract
      governs, and the exact path `skills/myflow-contracts/jira-followups.md`. This is the same stub
      shape `rules/myflow-manual-review.mdc` already uses for the contracts it no longer holds, and it
      is what keeps a reader of the core from concluding the rules were deleted.

- [x] Repoint every citation that pointed **into** the moved section. Find them:

```bash unverified:run in the apply worktree after the move
cd "<WT>"
grep -rn 'Follow-up issues\|jira-integration\.md' --include='*.md' --include='*.mdc' --include='*.sh' . \
  | grep -v '/openspec/changes/archive/'
```

      For each hit whose bold token names a heading that is now in `jira-followups.md`, change the
      backticked path to `skills/myflow-contracts/jira-followups.md` and change **nothing else on the
      line**. Keep the bold token and the path on one physical line.

- [x] Verify the move lost nothing, and verify the guard actually looked:

```bash unverified:run in the apply worktree after the move
cd "<WT>"
BASE="$(git rev-parse HEAD)"
diff <(git show "$BASE:skills/myflow-contracts/jira-integration.md" | grep -v '^[[:space:]]*$' | sort) \
     <(cat skills/myflow-contracts/jira-integration.md skills/myflow-contracts/jira-followups.md \
       | grep -v '^[[:space:]]*$' | sort)
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: the `diff` prints as `>` only the new file's title and header lines, the stub's lines, and
the repointed form of any citation line; and as `<` only the pre-repoint form of those same citation
lines. **No other line may appear as `<`.** Both guards exit 0. Then copy one repointed citation into
a scratch file, mutate its bold token, and confirm `scripts/check-references.sh` **fails** on it.

**A repointed citation is a `<`/`>` pair, and an earlier wording of this expectation said "nothing at
all as `<`", which forbade the one edit the split permits.** The exclusion is stated correctly in
`design.md`'s verification table and was dropped when this task was written; the task reviewer found
its consequence. Read the `<` side as *every removed line must be one you deliberately repointed*,
not as *no line was removed*.

**Build:** green

---

## 2. Partition `jira-integration.md`

### 2.1 Split the remaining `jira-integration.md` into core and rationale

**Files:**
- Create: `<WT>/skills/myflow-contracts/jira-integration-rationale.md`
- Modify: `<WT>/skills/myflow-contracts/jira-integration.md`

**Interfaces:**
- Consumes: the reduced `jira-integration.md` task 1.1 left behind.
- Produces: `skills/myflow-contracts/jira-integration-rationale.md`, carrying the same `##`/`###`
  heading sequence as the core.

Apply the partition test from **A contract file separates its normative core from its rationale**
(`openspec/changes/kan-82-cut-myflow-per-command-token-overhead/specs/myflow-contract-economy/spec.md`),
paragraph by paragraph: **core if removing it would change what an agent does; rationale if removing
it changes no agent behaviour.**

Worked examples from this file, so the boundary is not left to taste:

| Passage | Goes to | Why |
|---------|---------|-----|
| The `\| Command \| When \| Target status \|` transition table | core | an agent reads it to pick a transition |
| "Resolve transitions by name, never by identifier" | core | a prohibition, stated in prose |
| The pre-write assertion's two numbered conditions | core | an agent performs them |
| "That narrows the window; it does not close it, and this contract does not pretend otherwise" | rationale | removing it changes no behaviour |
| "This assumes `TO DO URGENT` means *not yet started* in every project…" | rationale | names an accepted limit; the mapping itself is stated elsewhere and stays |
| "The alternative was to infer a position… that is the thing this section exists to forbid" | rationale | justification of a rule already stated above it |

- [x] Create `<WT>/skills/myflow-contracts/jira-integration-rationale.md` with a `#` title, a bolded
      sentence stating it is the reasoning behind `skills/myflow-contracts/jira-integration.md` and
      that **a `/myflow-*` run never loads it**, and then the core's `##`/`###` headings reproduced in
      the same order.

- [x] Move each rationale passage under its mirrored heading, **byte for byte**. A heading whose
      section is wholly normative keeps its appendix heading with **no body beneath it** — present,
      not absent.

- [x] Repoint every citation whose target moved into the appendix, changing the path and nothing else
      on the line.

- [x] Add to the core's header one line stating that its reasoning lives in
      `skills/myflow-contracts/jira-integration-rationale.md` and that a `/myflow-*` run does not load
      it.

- [x] Verify no line was lost and the heading trees match:

```bash unverified:run in the apply worktree after the partition
cd "<WT>"
BASE="$(git rev-parse HEAD)"
diff <(git show "$BASE:skills/myflow-contracts/jira-integration.md" | grep -v '^[[:space:]]*$' | sort) \
     <(cat skills/myflow-contracts/jira-integration.md skills/myflow-contracts/jira-integration-rationale.md \
       | grep -v '^[[:space:]]*$' | sort)
headings() { awk '/^```/{f=!f; next} !f && /^#{2,4} /' "$1"; }
diff <(headings skills/myflow-contracts/jira-integration.md) \
     <(headings skills/myflow-contracts/jira-integration-rationale.md)
scripts/check-references.sh
wc -c skills/myflow-contracts/jira-integration.md skills/myflow-contracts/jira-integration-rationale.md
```

Expected: the first `diff` shows as `>` the appendix's title/header lines, the core's added header
line, and the repointed form of any citation line; and as `<` only the pre-repoint form of those same
citation lines. **No other line may appear as `<`.** The second `diff` is **empty** — identical
heading sequences. The guard exits 0. The `wc -c` output is recorded for task 5.1.

**A line that merely moved to the appendix produces no diff output at all**, because this `diff`
compares the original against core **plus** appendix — the line is present on both sides. So a `<`
line is never evidence that a passage was moved; it is evidence the passage was **lost**, unless it
is the pre-repoint form of a citation. A repointed citation is a `<`/`>` pair, and it is the one edit
**The split is a verbatim partition, with citation repointing as its only edit**
(`openspec/changes/kan-82-cut-myflow-per-command-token-overhead/specs/myflow-contract-economy/spec.md`)
permits.

- [x] Mutate one repointed citation's bold token in a scratch copy and confirm
      `scripts/check-references.sh` **fails**. A passing guard is not evidence it looked.

**Build:** green

---

## 3. Partition `pipeline.md`

### 3.1 Split `pipeline.md` into core and rationale

**Files:**
- Create: `<WT>/skills/myflow-contracts/pipeline-rationale.md`
- Modify: `<WT>/skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: the partition mechanics proved in task 2.1.
- Produces: `skills/myflow-contracts/pipeline-rationale.md`, heading tree mirrored.

This is the same operation as task 2.1 on a file three times the size, cited by 132 backticked
references across the corpus.

<!-- measured: grep -roh '(`[a-zA-Z0-9./_-]*pipeline\.md`)' --include='*.md' --include='*.mdc' --include='*.sh' . | wc -l @ adedf66 -->

Worked examples from this file, so the boundary is not left to taste:

| Passage | Goes to | Why |
|---------|---------|-----|
| The state table, the level-1 stage table, the transition table | core | an agent reads them to act |
| The guarded two-commit `&&` chain, and the worktree-cleanup checks 1-5 | core | an agent runs them |
| Every handoff template under **The block each state renders** | core | an agent renders them |
| "Never substitute a commit count" and the paragraph giving `rev-list --count`'s failure | **split** | the prohibition is core; the paragraph explaining why the count cannot separate the two states is rationale |
| "Signal 1 precedes signal 2, and that ordering is the point" | rationale | the order is already stated by the numbered list above it |
| "**Interleaving the removal into the worktree half was considered and rejected**" | rationale | a rejected alternative |
| "**This reverses an earlier decision, and the reversal is recorded rather than quiet**" | rationale | history |
| "Both of those facts are Claude Code's, and the rule is stated against the mechanism" | rationale | meta-commentary about the document |

**Where a passage states a rule and then justifies it in the same paragraph, the whole paragraph
stays in the core.** This instruction previously said to split it at the sentence boundary, and that
is **struck**: it is unexecutable against the line-multiset check below, for the reason the
paragraph-granularity global constraint gives. Task 2.1's implementer found this by trying it. Move
whole blank-line-delimited paragraphs, and leave a mixed paragraph whole.

The fourth row of the worked-examples table above — "Never substitute a commit count", marked
**split** — is struck with it. That paragraph stays in the core entire.

- [x] Create `<WT>/skills/myflow-contracts/pipeline-rationale.md` with a `#` title, a bolded sentence
      naming it the reasoning behind `skills/myflow-contracts/pipeline.md` and stating that **a
      `/myflow-*` run never loads it**, then the core's `##`/`###` headings in the same order.

- [x] Move each rationale passage under its mirrored heading, byte for byte. A wholly normative
      section keeps an empty appendix heading.

- [x] Add to the core's header one line naming the appendix path and the no-load-at-runtime rule.

- [x] Repoint every citation whose target moved. `pipeline.md` cites **itself** throughout —
      `**Git boundaries** (`pipeline.md`)` — so a self-citation whose target is now in the appendix
      becomes `(`pipeline-rationale.md`)`, and one whose target stayed is left untouched.

- [x] Verify:

```bash unverified:run in the apply worktree after the partition
cd "<WT>"
BASE="$(git rev-parse HEAD)"
diff <(git show "$BASE:skills/myflow-contracts/pipeline.md" | grep -v '^[[:space:]]*$' | sort) \
     <(cat skills/myflow-contracts/pipeline.md skills/myflow-contracts/pipeline-rationale.md \
       | grep -v '^[[:space:]]*$' | sort)
headings() { awk '/^```/{f=!f; next} !f && /^#{2,4} /' "$1"; }
diff <(headings skills/myflow-contracts/pipeline.md) \
     <(headings skills/myflow-contracts/pipeline-rationale.md)
scripts/check-references.sh
scripts/check-vocabulary.sh
wc -c skills/myflow-contracts/pipeline.md skills/myflow-contracts/pipeline-rationale.md
```

Expected: the first `diff` shows as `>` only the appendix's title/header lines, the core's added
header line, and the repointed form of any citation line; and as `<` only the pre-repoint form of
those same citation lines. **No other line may appear as `<`** — read task 2.1's note on why a moved
line produces no diff output at all, which applies here unchanged. The second `diff` is empty. Both
guards exit 0. The `wc -c` output is recorded for task 5.1.

- [x] Mutate one repointed citation's bold token in a scratch copy and confirm
      `scripts/check-references.sh` **fails**.

- [x] Confirm `/myflow-info` is still satisfiable: the state diagram, the level-1 stage table and the
      level-2 expansions are **all in the core**, because that command reads `pipeline.md` at
      invocation time and is forbidden from answering from memory.

```bash unverified:run in the apply worktree after the partition
cd "<WT>"
grep -c '^stateDiagram-v2' skills/myflow-contracts/pipeline.md
grep -c '^#### ' skills/myflow-contracts/pipeline.md
```

Expected: `1` for the diagram, and at least `8` — one per level-2 expansion the distribution spec
names.

<!-- measured: the eight-stage table in openspec/specs/myflow-contract-distribution/spec.md @ adedf66 -->

**Build:** green

---

## 4. Update the index, the rule layer and the entry-point files

### 4.1 Name the new files everywhere a contract file is discovered

**Files:**
- Modify: `<WT>/skills/myflow-contracts/SKILL.md`
- Modify: `<WT>/rules/myflow-manual-review.mdc`
- Modify: `<WT>/CLAUDE.md`
- Modify: `<WT>/AGENTS.md`

**Interfaces:**
- Consumes: the five files groups 1-3 produced.
- Produces: an index that agrees with the directory in both directions — the delta spec's
  **The index and the directory agree** scenario.

- [x] In `<WT>/skills/myflow-contracts/SKILL.md`, add a row for each of
      `pipeline-rationale.md`, `jira-integration-rationale.md` and `jira-followups.md`, naming when
      each is needed. Confirm the index also names the three files that were missing from it before
      this change — `build-green.md`, `plan-provenance.md` and `workspace-isolation.md` — adding rows
      for any that are absent, since the delta spec requires the index and the directory to agree in
      both directions.

- [x] In the same file, state the runtime rule under its own heading: **a `/myflow-*` run never loads
      a rationale appendix**; appendices are for whoever edits a contract. State plainly that nothing
      enforces it.

- [x] In `<WT>/rules/myflow-manual-review.mdc`, add `jira-followups.md` to the table of narrower
      contracts with the command that needs it (`/myflow-finish` run 1), and add one line carrying the
      runtime rule. **Do not add the appendices to that table** — it lists what a command loads, and a
      command loads no appendix.

- [x] In `<WT>/CLAUDE.md` and `<WT>/AGENTS.md`, update the `skills/myflow-contracts/` row of each
      skill index so it names the new files. Make the **same** edit in both: a project read by two
      harnesses that disagree is the failure `setup.sh` renders both files to avoid.

- [x] Verify the index and the directory agree, and that no appendix is named in a load instruction:

```bash unverified:run in the apply worktree after the index edits
cd "<WT>"
# Extract the markdown LINK targets, not every `[a-z-]*.md`-shaped substring: a bare
# pattern also matches `myflow-review.md` inside `rules/myflow-manual-review.mdc` and
# `project.md` inside the literal text `.myflow/project.md`, neither of which is a file
# in this directory, and both of which make the diff below non-empty for no reason.
diff <(ls skills/myflow-contracts/*.md | xargs -n1 basename | grep -v '^SKILL.md$' | sort) \
     <(grep -oE '\]\([a-z0-9-]+\.md\)' skills/myflow-contracts/SKILL.md \
        | sed -E 's/^\]\(//; s/\)$//' | grep -v '^SKILL.md$' | sort -u)
grep -rn 'rationale\.md' skills/myflow-start/SKILL.md skills/myflow-do/SKILL.md \
  skills/myflow-finish/SKILL.md skills/myflow-status/SKILL.md skills/myflow-info/SKILL.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: the `diff` is **empty**. The `grep` finds **nothing** — no skill instructs an appendix
load. Both guards exit 0.

**Build:** green

---

## 5. The budget guard

### 5.1 Write `scripts/check-contract-budget.sh` and its harness

**Files:**
- Create: `<WT>/scripts/check-contract-budget.sh`
- Create: `<WT>/scripts/test-check-contract-budget.sh`
- Modify: `<WT>/.myflow/project.md`

**Interfaces:**
- Consumes: the byte sizes tasks 2.1, 3.1 and 4.1 printed with `wc -c`.
- Produces: `scripts/check-contract-budget.sh`, argument-free, honouring
  `CHECK_CONTRACT_BUDGET_ROOT` for its harness alone; exit 0 all within budget, 1 over budget or
  missing an entry, 2 cannot answer.

This task runs **last** because its numbers are what the four tasks before it produced. Budgets for
the files this change does not touch are computable now — achieved size plus 25%, rounded down:

| File | bytes today | budget |
|------|------------|--------|
| `build-green.md` | 3646 | 4557 |
| `plan-provenance.md` | 24466 | 30582 |
| `project-configuration.md` | 45236 | 56545 |
| `state-file.md` | 15951 | 19938 |
| `state-self-heal.md` | 15236 | 19045 |
| `workspace-isolation.md` | 31079 | 38848 |

<!-- measured: wc -c skills/myflow-contracts/*.md @ adedf66, each multiplied by 1.25 and truncated -->

- [x] Generate the remaining rows from the tree as it now stands, rather than transcribing them:

```bash unverified:run in the apply worktree after tasks 1.1 through 4.1
cd "<WT>"
for f in skills/myflow-contracts/*.md; do
  printf '%s %d\n' "$(basename "$f")" "$(( $(wc -c < "$f") * 125 / 100 ))"
done | sort
```

Expected: nine or more rows, one per `.md` in the directory, each a filename and a budget.

- [x] Write `<WT>/scripts/check-contract-budget.sh`. Resolve the repository root from the script's own
      location exactly as `scripts/check-references.sh` does, honouring an explicit
      `CHECK_CONTRACT_BUDGET_ROOT` override only when set, and rejecting it when set but empty:

```bash unverified:authored for this change; run its harness to confirm
#!/usr/bin/env bash
# check-contract-budget.sh — fail when a contract file outgrows its declared budget.
#
# Budgets are a ratchet against regrowth, NOT a target: each was set to the size
# the file actually had when the core/rationale split landed, plus 25%. Raising
# one is a deliberate, visible diff in this table — which is the whole reason the
# table lives here rather than in the file being measured.
#
# Exit codes: 0 every file within budget; 1 a file over budget or with no entry;
# 2 the guard cannot answer at all.
set -euo pipefail

if [ -n "${CHECK_CONTRACT_BUDGET_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_CONTRACT_BUDGET_ROOT"
elif [ "${CHECK_CONTRACT_BUDGET_ROOT+set}" = "set" ]; then
  printf 'CHECK_CONTRACT_BUDGET_ROOT is set but empty\n' >&2
  exit 2
else
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi
[ -d "$REPO_ROOT" ] || { printf 'not a directory: %s\n' "$REPO_ROOT" >&2; exit 2; }

CONTRACT_DIR="$REPO_ROOT/skills/myflow-contracts"
[ -d "$CONTRACT_DIR" ] || { printf 'no contracts directory: %s\n' "$CONTRACT_DIR" >&2; exit 2; }

# name<space>max-bytes. A file in CONTRACT_DIR with no row here is a violation,
# so a contract added later cannot silently escape the ratchet.
budgets() {
  cat <<'EOF'
build-green.md 4557
plan-provenance.md 30582
project-configuration.md 56545
state-file.md 19938
state-self-heal.md 19045
workspace-isolation.md 38848
EOF
}

violations=0
checked=0
for path in "$CONTRACT_DIR"/*.md; do
  [ -e "$path" ] || { printf 'no .md files in %s\n' "$CONTRACT_DIR" >&2; exit 2; }
  checked=$((checked + 1))
  name="$(basename "$path")"
  max="$(budgets | awk -v n="$name" '$1 == n { print $2; exit }')"
  actual="$(wc -c < "$path" | tr -d ' ')"
  if [ -z "$max" ]; then
    printf '%s: no budget declared — add a row to budgets() in %s\n' \
      "skills/myflow-contracts/$name" "$(basename "${BASH_SOURCE[0]}")"
    violations=$((violations + 1))
  elif [ "$actual" -gt "$max" ]; then
    printf '%s: %s bytes exceeds budget %s\n' \
      "skills/myflow-contracts/$name" "$actual" "$max"
    violations=$((violations + 1))
  fi
done

if [ "$violations" -gt 0 ]; then
  printf 'BUDGET-FAIL: %d file(s) over budget or undeclared\n' "$violations"
  exit 1
fi
printf 'BUDGET-OK: %d contract file(s) within budget\n' "$checked"
```

      **The success line carries the count deliberately.** A guard that reports a bare `BUDGET-OK`
      cannot be told apart from one that scanned nothing, which is the failure the self-scoped
      `REPO_ROOT` above exists to prevent — so the count is the evidence that it looked.

- [x] Fill `budgets()` with the rows the generator above printed, so every `.md` in the directory has
      one — the six unchanged files from the table, plus `SKILL.md`, `pipeline.md`,
      `pipeline-rationale.md`, `jira-integration.md`, `jira-integration-rationale.md` and
      `jira-followups.md`.

- [x] `chmod +x <WT>/scripts/check-contract-budget.sh`.

- [x] Write `<WT>/scripts/test-check-contract-budget.sh`, following the shape of the harnesses already
      in `scripts/`: build a fixture tree under `mktemp -d`, point the guard at it with
      `CHECK_CONTRACT_BUDGET_ROOT`, and assert each outcome:

```bash unverified:authored for this change; run it to confirm
#!/usr/bin/env bash
# test-check-contract-budget.sh — harness for check-contract-budget.sh.
set -euo pipefail
GUARD="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-contract-budget.sh"
failures=0

expect() {  # expect <label> <want-exit> <root>
  local label="$1" want="$2" root="$3" got=0
  CHECK_CONTRACT_BUDGET_ROOT="$root" "$GUARD" >/dev/null 2>&1 || got=$?
  if [ "$got" -ne "$want" ]; then
    printf 'FAIL %s: exit %d, wanted %d\n' "$label" "$got" "$want"
    failures=$((failures + 1))
  else
    printf 'ok   %s\n' "$label"
  fi
}

FIX="$(mktemp -d)"
trap 'rm -rf "$FIX"' EXIT

mkdir -p "$FIX/within/skills/myflow-contracts"
printf 'x\n' > "$FIX/within/skills/myflow-contracts/build-green.md"
expect 'a file well under budget passes' 0 "$FIX/within"

mkdir -p "$FIX/over/skills/myflow-contracts"
head -c 5000 /dev/zero | tr '\0' 'x' > "$FIX/over/skills/myflow-contracts/build-green.md"
expect 'a file over budget fails' 1 "$FIX/over"

mkdir -p "$FIX/undeclared/skills/myflow-contracts"
printf 'x\n' > "$FIX/undeclared/skills/myflow-contracts/brand-new-contract.md"
expect 'a file with no budget row fails' 1 "$FIX/undeclared"

expect 'a missing contracts directory cannot be answered' 2 "$FIX/absent"

CHECK_CONTRACT_BUDGET_ROOT='' "$GUARD" >/dev/null 2>&1 && rc=0 || rc=$?
if [ "$rc" -ne 2 ]; then
  printf 'FAIL an empty override exits %d, wanted 2\n' "$rc"
  failures=$((failures + 1))
else
  printf 'ok   an empty override is refused\n'
fi

[ "$failures" -eq 0 ] || { printf '%d failure(s)\n' "$failures"; exit 1; }
printf 'all checks passed\n'
```

      **The over-budget fixture uses `build-green.md` at 5000 bytes against its 4557-byte budget**, so
      the assertion is real rather than nominal.

- [x] `chmod +x <WT>/scripts/test-check-contract-budget.sh`.

- [x] Add `scripts/check-contract-budget.sh` to `## lint` and `scripts/test-check-contract-budget.sh`
      to `## test` in `<WT>/.myflow/project.md`. Add one sentence there stating what the guard is —
      a ratchet set from achieved size plus 25%, raised only by a deliberate edit to its table —
      matching how that file already annotates `check-workspace-isolation.sh`.

- [x] Verify the guard against the real tree and prove it can fail:

```bash unverified:run in the apply worktree after the guard is written
cd "<WT>"
scripts/test-check-contract-budget.sh
scripts/check-contract-budget.sh
printf '\n%s\n' "$(head -c 200000 /dev/zero | tr '\0' 'x')" >> skills/myflow-contracts/build-green.md
scripts/check-contract-budget.sh; echo "exit=$?"
git checkout -- skills/myflow-contracts/build-green.md
scripts/check-contract-budget.sh
```

Expected: the harness prints `all checks passed`; the guard prints `BUDGET-OK` on the clean tree;
after the append it prints the `build-green.md` over-budget line and `exit=1`; after the checkout it
prints `BUDGET-OK` again.

- [x] Prove the guard is self-scoped rather than cwd-dependent, which is the delta spec's
      **The guard runs from any working directory** scenario. A guard that silently scanned nothing
      would report a false clean:

```bash unverified:run in the apply worktree after the guard is written
cd /tmp && "<WT>/scripts/check-contract-budget.sh"
```

Expected: `BUDGET-OK`, naming the same file count as the run from inside the worktree — not a clean
result reached by finding no files.

**Build:** green

---

## 6. Whole-repository verification

### 6.1 Run every declared guard and the installer sandbox

**Files:**
- Modify: none — this task changes nothing and only proves the change.

**Interfaces:**
- Consumes: everything tasks 1.1 through 5.1 produced.

- [x] Run every command in `.myflow/project.md`'s `## lint` and `## test` sections, in order:

```bash unverified:run in the apply worktree once every other task is complete
cd "<WT>"
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-uncommitted-review-package.sh
scripts/test-check-task-build-green.sh
scripts/test-check-workspace-isolation.sh
scripts/test-check-contract-budget.sh
```

Expected: every one exits 0. A non-zero exit is a real hit on this change — fix the offending line,
never the guard.

- [x] Prove the new files reach all three harnesses with **no installer edit**, which is the delta
      spec's **Files added to the contracts skill ship without an installer edit** scenario:

```bash unverified:run in the apply worktree once every other task is complete
cd "<WT>"
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global >/dev/null
for h in .claude .cursor .codex; do
  for f in pipeline.md pipeline-rationale.md jira-integration.md \
           jira-integration-rationale.md jira-followups.md; do
    test -r "$SANDBOX/$h/skills/myflow-contracts/$f" \
      || echo "MISSING $h/$f"
  done
done
git diff --stat -- setup.sh
rm -rf "$SANDBOX"
```

Expected: no `MISSING` line, and `git diff --stat -- setup.sh` prints **nothing** — the installer was
not touched.

- [x] Record the saving, which is the change's whole point:

```bash unverified:run in the apply worktree once every other task is complete
cd "<WT>"
BASE="$(git rev-parse HEAD)"
before=$(( $(git show "$BASE:skills/myflow-contracts/pipeline.md" | wc -c) \
         + $(git show "$BASE:skills/myflow-contracts/jira-integration.md" | wc -c) ))
after=$(( $(wc -c < skills/myflow-contracts/pipeline.md) \
        + $(wc -c < skills/myflow-contracts/jira-integration.md) ))
printf 'a /myflow-start run loaded %d bytes of these two files, now loads %d\n' "$before" "$after"
```

Expected: `before` is 156109 and `after` is smaller; report the actual figure in the handoff rather
than a target.

<!-- measured: 103326 + 52783, from the baseline wc above @ adedf66 -->

- [x] Confirm no rule changed. Every difference between the old files and the new five is a moved
      line or a repointed citation path:

```bash unverified:run in the apply worktree once every other task is complete
cd "<WT>"
BASE="$(git rev-parse HEAD)"
for pair in "pipeline.md pipeline-rationale.md" \
            "jira-integration.md jira-integration-rationale.md jira-followups.md"; do
  set -- $pair
  orig="$1"
  diff <(git show "$BASE:skills/myflow-contracts/$orig" | grep -v '^[[:space:]]*$' | sort) \
       <(cd skills/myflow-contracts && cat "$@" | grep -v '^[[:space:]]*$' | sort)
done
```

Expected: for each pair, **every `<` line is one of the two permitted edits** — the pre-repoint form
of a repointed citation, or the pre-deletion form of a line whose stale `above`/`below` was removed —
and every `>` line is a new file's own title or header, a pointer stub, or the post-edit form of a
`<` line. Review them one by one, not skimmed; a `<` line that is neither is lost text.

**This expectation previously read "no `<` lines at all", and that is struck.** It predates the
reversal that made deleting a stale position word a permitted edit, so it contradicted the global
constraint stated at the top of this file — and a run that took it literally would have reported a
failure on correct work. The count is not zero and was never going to be: it is 14 for the pipeline
pair and 13 for the jira triple, one per permitted edit.

**Build:** green
