# Tasks — cut `pipeline.md`'s load cost by splitting it by consumer

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** move five sections out of `skills/myflow-contracts/pipeline.md` into files loaded only by
the commands that reach them, give each its own rationale appendix, sweep inline argument prose into
those appendices, and prove nothing was lost.

**Architecture:** twelve files where there were two. Each extraction is atomic — core, appendix,
citation repointing and budget rows land in one commit, so every guard exits clean at every task
boundary.

**Spec:** `design.md` beside this file, and `specs/myflow-contract-economy/spec.md`.

## Global constraints

- **This is a relocation. No contract's meaning changes.** A sentence that reads differently after
  the move is a defect, not an improvement. Cut nothing; move it or leave it.
- **Never a paraphrase.** Passages move byte-for-byte. The only permitted edit to a moved passage is
  repointing a citation whose target file changed.
- **Every task ends with every guard in `.myflow/project.md`'s `## lint` exiting clean.** No task
  may leave `check-references.sh`, `check-installed-citations.sh` or `check-contract-budget.sh`
  failing "until the next task fixes it". There is no auto-fix for `scripts/check-*`; fix the
  offending line.
- **`setup.sh` is not edited.** `install_skills` symlinks whole skill directories, so files added
  under `skills/myflow-contracts/` install with no installer change. Verify this rather than
  assuming it (task 8).
- **`rules/myflow-manual-review.mdc` gains no row**, per decision `load-declaration-site` in
  `design.md`. Its contract table and its `pipeline.md`-first instruction stay byte-identical.
- **Every task emits a per-move ledger**, per **Requirement: A move or eviction is recorded in a
  per-move ledger** (`openspec/specs/myflow-contract-economy/spec.md`). Four columns, one row per
  removed passage. The review panel checks it against the diff in both directions.

### The shape every extracted core takes

Copied from `skills/myflow-contracts/handoff-blocks.md`, which is the established precedent — not
invented here.

```markdown verified:this is handoff-blocks.md's own opening, with its subject substituted
# <Section title>

<one line naming what the file governs>

**Loaded by `<command>`, `<command>` and `<command>`.** <one line on when.>

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/myflow-contracts/<name>-rationale.md`;
**a `/myflow-*` run never loads it.**

## <the original `##` heading, verbatim>

<the original body, verbatim>
```

### The shape every appendix takes

```markdown verified:this is handoff-blocks-rationale.md's own opening, with its subject substituted
# <Section title> — rationale

This file is the reasoning behind `skills/myflow-contracts/<name>.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## <the same `##` heading as the core>

<the reasoning moved out of pipeline-rationale.md's matching section>
```

A wholly normative section leaves its appendix heading **present with no body**, per **Requirement:
An appendix mirrors its core's heading tree** (`openspec/specs/myflow-contract-economy/spec.md`).

---

### 1 Capture the verification baselines

**Build:** green

**Files:**
- Create: `openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/normative-baseline.txt`
- Create: `openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/sentence-baseline.txt`
- Create: `openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md`
- Create: `/tmp/kan-295-sentence-set.py` *(throwaway — never committed)*

**Interfaces:**
- Consumes: nothing.
- Produces: the two baselines tasks 2-7 are measured against, and the script task 8 re-runs.

**Nothing may be edited before this task lands.** A baseline captured after the first edit proves
nothing. Run every capture against `HEAD` with a clean working tree.

- [x] **Step 1: Capture the normative inventory**

```bash unverified:the redirect target does not exist until this step runs
scripts/check-normative-inventory.sh \
  > openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/normative-baseline.txt
```

Its contract: exit `0` the inventory printed, exit `2` it cannot answer. It has no violation code.
A non-zero exit here is an environment failure — stop and fix it; do not proceed with a partial
baseline.

- [x] **Step 2: Write the throwaway sentence-set script**

`/tmp/kan-295-sentence-set.py`, Python 3 standard library only. It takes one or more file paths,
splits their Markdown into sentences, normalises each sentence's internal whitespace to single
spaces, strips surrounding whitespace, drops empty results, and prints one sentence per line,
sorted. It normalises **nothing else** — case, punctuation and wording are the content being
protected, exactly as `check-normative-inventory.sh` states for its own inventory.

It is not committed and gets no harness, no `## lint` row and no `budgets()` row, per decision
`no-cut-proof` in `design.md`: it answers a question about this change alone, and a future
relocation would have no baseline for it to diff against.

- [x] **Step 3: Capture the sentence baseline**

```bash unverified:the script is written in step 2 of this same task
python3 /tmp/kan-295-sentence-set.py \
  skills/myflow-contracts/pipeline.md \
  skills/myflow-contracts/pipeline-rationale.md \
  > openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/sentence-baseline.txt
```

- [x] **Step 4: Re-derive the citation inventory, per section**

The counts in `design.md` were measured at `15f11cc`. Re-derive them against `HEAD` before relying
on them, and record the per-section lists in `verification.md`. Search **every** citation shape
`check-references.sh` recognises, not only the parenthesised one — the guard's own header enumerates
four:

```bash verified:these are the four shapes check-references.sh's is_associated documents
grep -rn --include='*.md' --include='*.mdc' 'myflow-contracts/pipeline\.md' \
  skills rules commands commands-claude openspec/specs \
  CLAUDE.md AGENTS.md README.md .myflow
```

`design.md` records **57** citations into `pipeline.md`, **33** of them into a moved section. A
count that differs at `HEAD` replaces the recorded one — in `verification.md` and in `design.md`
both — rather than being reconciled to it.

- [x] **Step 5: Record the four cross-boundary determinations**

`verification.md` carries `design.md`'s cross-boundary table verbatim and, for each of the four,
confirms by reading that the verdict still holds at `HEAD`. Per **Requirement: A passage another
command depends on is hoisted before a single-command move**
(`openspec/specs/myflow-contract-economy/spec.md`), a green `check-references.sh` is **not**
evidence here and must not be offered as one.

**Tests:** no new test file. `scripts/check-normative-inventory.sh` and the throwaway script are the
verification instruments this task installs, not things it tests.

**Regression:** Reverting this task destroys the only record of what the corpus said before the
move, making every later claim that nothing was lost unfalsifiable.

**Baseline:** `scripts/check-normative-inventory.sh` before=exit 0 after=exit 0; the repository's
`## lint` and `## test` lists gain no entry.

**Commit:** `docs(openspec): capture the pre-move normative and sentence baselines`

---

### 2 Extract `worktree-resolution.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/worktree-resolution.md`
- Create: `skills/myflow-contracts/worktree-resolution-rationale.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-status/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `README.md`, `CLAUDE.md`, `AGENTS.md`
- Modify: `commands/myflow-finish.md`, `commands-claude/myflow-finish.md`
- Modify: `openspec/specs/myflow-finish-cleanup/spec.md` (citation repoint only; left unstaged, per
  the git-boundaries prohibition on this worktree committing anything under `openspec/`)
- **Allowed-collateral:** `skills/myflow-contracts/finish-contract.md`

**Interfaces:**
- Consumes: task 1's citation inventory.
- Produces: the extraction pattern tasks 3-6 repeat. Get the shape right here.

The smallest section (1517 bytes) goes first, so the pattern is settled on the cheapest case.

- [x] **Step 1: Move the section**

`## Resolving a change's worktrees` moves out of `pipeline.md` into
`worktree-resolution.md`, body verbatim, wrapped in the core shape above. Its opening names its
consumers: `/myflow-do`, `/myflow-finish`, `/myflow-status` and `/myflow-fast`.

- [x] **Step 2: Move the matching rationale**

`## Resolving a change's worktrees` moves out of `pipeline-rationale.md` into
`worktree-resolution-rationale.md`. If that section is empty in `pipeline-rationale.md`, the
appendix carries the heading with no body — present, not absent.

- [x] **Step 3: Repoint every citation naming the section**

Task 1's list. Each becomes
``**Resolving a change's worktrees** (`skills/myflow-contracts/worktree-resolution.md`)``.

- [x] **Step 4: Declare the load in each consuming `SKILL.md`**

One line at the step that needs it — `/myflow-do`'s workspace-isolation gate, `/myflow-finish`'s
preflight verdict and run 2 removal, `/myflow-status`'s merge-status report, `/myflow-fast`'s
citation of all three. Not a line at the top of the file: the point of the split is that a command
loads it when it reaches the step, not on principle.

- [x] **Step 5: Add the two `budgets()` rows**

Landing size plus 25%, keyed on the path relative to the repository root. `pipeline.md`'s and
`pipeline-rationale.md`'s rows are **not** lowered here — task 8 does that once, after the sweep, so
the number is not re-derived five times.

- [x] **Step 6: Emit the ledger and verify**

```bash verified:these are the guards .myflow/project.md's ## lint section names
scripts/check-references.sh \
  && scripts/check-installed-citations.sh \
  && scripts/check-contract-budget.sh \
  && scripts/check-markdown-integrity.py
```

**Tests:** no new test file; the four guards above are the verification, plus the ledger the review
panel reads.

**Regression:** Reverting this task returns the section to `pipeline.md`, so `/myflow-start` pays
1517 bytes per run for a section it can never reach.

**Baseline:** `scripts/check-references.sh`, `check-installed-citations.sh`,
`check-contract-budget.sh`, `check-markdown-integrity.py` before=exit 0 after=exit 0.

**Commit:** `refactor(myflow-contracts): move worktree resolution into its own contract`

---

### 3 Extract `session-records.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/session-records.md`
- Create: `skills/myflow-contracts/session-records-rationale.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- **Allowed-collateral:** `skills/myflow-contracts/finish-contract.md`

**Interfaces:**
- Consumes: task 2's established pattern.
- Produces: the one pair in this change whose core and appendix headings deliberately differ.

- [x] **Step 1: Move the section and its rationale**

Core heading `## Rendering the session records`; appendix heading `## Preserving the session
records`. **The two differ on purpose.** `pipeline.md` states in prose that the appendix "keeps its
former name"; that sentence moves to `session-records.md` alongside the section it explains.
Renaming either heading is out of scope, per decision `session-records-heading` in `design.md`.

- [x] **Step 2: Repoint, declare, budget, verify**

As task 2 steps 3-6. `/myflow-do` reads this on its `prUrl` commit path; `/myflow-finish` reads it
in run 1. `/myflow-start` and `/myflow-status` render no record and get no load line.

**Tests:** no new test file; the four guards from task 2 step 6, plus the ledger.

**Regression:** Reverting this task restores the heading mismatch to `pipeline.md`/`pipeline-rationale.md`
and puts 1486 bytes back into every `/myflow-start` and `/myflow-status` run.

**Baseline:** the four guards before=exit 0 after=exit 0.

**Commit:** `refactor(myflow-contracts): move session-record rendering into its own contract`

---

### 4 Extract `git-boundaries.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/git-boundaries.md`
- Create: `skills/myflow-contracts/git-boundaries-rationale.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-start/SKILL.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-finish/SKILL-rationale.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `README.md`
- Modify: `openspec/specs/myflow-command-surface/spec.md` (citation repoint only; left unstaged, per
  the git-boundaries prohibition on this worktree committing anything under `openspec/`)

**Interfaces:**
- Consumes: task 1 step 5's determination that `myflow-start/SKILL.md`'s citation is a pointer.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Move the section and its rationale**

The `##` heading, the git-boundaries table, the `&&` chain fence, the planning-path rules and the
tracked-symlink prohibition move whole. `pipeline-rationale.md`'s `## Git boundaries` section moves
to the appendix.

- [x] **Step 2: Repoint `myflow-start/SKILL.md:537` without hoisting**

`/myflow-start` does not load `git-boundaries.md`. Its guardrail —
"**Never** commit anything. Stage the planning artifacts and leave the commit to `/myflow-finish`" —
already states in full the rule `/myflow-start` obeys, so the citation is a pointer to detail, not a
dependency on substance. Repoint it and change nothing else. **Do not hoist the table row into
`pipeline.md`**: a one-row fragment of a table another file owns is exactly the copy that goes
wrong silently.

- [x] **Step 3: Repoint the rest, declare, budget, verify**

As task 2 steps 3-6. `/myflow-do`, `/myflow-finish` and `/myflow-fast` get load lines at their own
staging and commit steps.

**Tests:** no new test file; the four guards from task 2 step 6, plus the ledger.

**Regression:** Reverting this task puts 3946 bytes back into every `/myflow-start` and
`/myflow-status` run, neither of which commits anything.

**Baseline:** the four guards before=exit 0 after=exit 0.

**Commit:** `refactor(myflow-contracts): move git boundaries into their own contract`

---

### 5 Extract `artifacts-registry.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/artifacts-registry.md`
- Create: `skills/myflow-contracts/artifacts-registry-rationale.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-contracts/workspace-isolation.md`
- Modify: `skills/myflow-contracts/workspace-isolation-rationale.md`
- Modify: `skills/myflow-contracts/project-configuration-rationale.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `commands/myflow-finish.md`, `commands-claude/myflow-finish.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `README.md`, `CLAUDE.md`, `AGENTS.md`
- **Allowed-collateral:** `skills/myflow-contracts/project-configuration.md`,
  `skills/myflow-contracts/finish-contract.md`

**Interfaces:**
- Consumes: task 2's pattern.
- Produces: nothing later tasks depend on.

The widest citation surface of the five: ten citations across two command trees, two rationale
appendices, and both repository instruction files.

- [x] **Step 1: Move the section and its rationale**

The registry table and every per-row prose block move whole — including the conditional
proposal-artifact row, the workspace row's "asking, not looking" rule, the archive-branch leftovers
naming five real branches, and the claimed-cache-index rule. The `pipeline.md` sentence declaring
this table "the one place a cleanup rule is stated" moves with it and keeps its wording; it names the
table, not the file it sits in.

- [x] **Step 2: Repoint, declare, budget, verify**

As task 2 steps 3-6. Both `commands/` trees must be edited, per **State transitions**
(`skills/myflow-contracts/pipeline.md`): a command file that disagrees with its skill is
non-determinism in the one layer that must be deterministic.

**Tests:** no new test file; the four guards from task 2 step 6, plus the ledger.

**Regression:** Reverting this task puts 6183 bytes back into every `/myflow-start` and
`/myflow-status` run, neither of which removes an artifact.

**Baseline:** the four guards before=exit 0 after=exit 0.

**Commit:** `refactor(myflow-contracts): move the temporary artifacts registry into its own contract`

---

### 6 Extract `model-policy.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/model-policy.md`
- Create: `skills/myflow-contracts/model-policy-rationale.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `skills/myflow-start/SKILL.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `README.md`, `CLAUDE.md`, `AGENTS.md`
- Modify: `skills/README.md`, `commands/myflow-do.md`, `commands/myflow-fast.md`

**Interfaces:**
- Consumes: task 1 step 5's determination that `state-file.md:166` is a pointer.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Move the section and its rationale**

The whole section, including the three-role table, the subagent-model overrides of
`superpowers:subagent-driven-development`, the `unknown (agent-defined)` rule and the per-harness
enforcement notes. `pipeline-rationale.md`'s `## Model policy` section moves to the appendix.

The opening sentence "Change the capability first and bring this section with it: a section that
contradicts the OpenSpec requirement is this file's defect, not the spec's" moves too, and now names
`model-policy.md` as the file whose defect it would be.

- [x] **Step 2: Repoint `state-file.md:166` without hoisting**

`state-file.md` is loaded by every command; `model-policy.md` is loaded by three. `/myflow-finish`
and `/myflow-status` carry `models` forward verbatim — which `state-file.md` states itself, in the
same sentence — so they need the roles' definitions not at all. Repoint and change nothing else.

- [x] **Step 3: Repoint the rest, declare, budget, verify**

As task 2 steps 3-6. `/myflow-start` loads it at its model questions; `/myflow-do` at each
implementer and panel dispatch; `/myflow-fast` at both.

**Tests:** no new test file; the four guards from task 2 step 6, plus the ledger.

**Regression:** Reverting this task puts 6661 bytes back into every `/myflow-finish` and
`/myflow-status` run, neither of which dispatches a subagent.

**Baseline:** the four guards before=exit 0 after=exit 0.

**Commit:** `refactor(myflow-contracts): move model policy into its own contract`

---

### 7 Sweep inline rationale, and name the family

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-contracts/git-boundaries.md`, `git-boundaries-rationale.md`
- Modify: `skills/myflow-contracts/model-policy.md`, `model-policy-rationale.md`
- Modify: `skills/myflow-contracts/artifacts-registry.md`, `artifacts-registry-rationale.md`
- Modify: `skills/myflow-contracts/session-records.md`, `session-records-rationale.md`
- Modify: `skills/myflow-contracts/worktree-resolution.md`, `worktree-resolution-rationale.md`

**Interfaces:**
- Consumes: all five extractions, complete.
- Produces: the final shape task 8 measures and budgets.

- [x] **Step 1: Apply the core/rationale test to every section in the family**

One test, from **Requirement: A contract file separates its normative core from its rationale**
(`openspec/specs/myflow-contract-economy/spec.md`), applied to sentence groups — a paragraph, a
bullet, a table — and **never to clauses inside a sentence**:

> A passage is **core** if removing it would change what an agent does. A passage is **rationale**
> if removing it changes no agent behaviour.

**The test is behavioural, not typographic.** A prose paragraph stating a rule is core. Core keeps
every table, ordered step list, handoff template, code block an agent runs, and every prohibition.

- [x] **Step 2: Move, never rewrite**

Each moved passage lands in its appendix under the heading it sat under, verbatim. Where the core
must retain a pointer, the pointer is a new sentence — that is a **rule extraction**, and its ledger
row reads `— (rule extracted)` with the new core sentence quoted, per the ledger requirement.

The largest single candidate is `Stage marks`' argument for why the session token must be a literal
rather than a shell substitution — roughly 1.5 KB restated across four sentences and three
enforcement sites. **The prohibition itself stays in the core**, as does the enumeration of the
three enforcement sites: an agent that loses either writes a substitution.

- [x] **Step 3: Name the family in `pipeline.md`**

One short block, added only now that all five files exist — `check-references.sh` skips a backticked
path that resolves to no file, so an earlier addition would have been unverified. It names the five
siblings and what each governs. It is a discovery aid, not a stub: it claims no authority over text
`pipeline.md` no longer contains, per **Requirement: Canonical authority moves with the contract
text** (`openspec/specs/myflow-contract-distribution/spec.md`).

- [x] **Step 4: Re-mirror every heading tree**

Each of the six appendices carries its core's `##` and `###` headings, in the same order. A section
that is wholly normative leaves its appendix heading present with **no body** — missing rather than
dropped, so it is visible the section was examined rather than skipped.

**Tests:** no new test file; the four guards from task 2 step 6, plus the ledger, which is largest
for this task.

**Regression:** Reverting this task restores roughly 6-8 KB of argument prose to files `/myflow-do`
loads on every run, and re-breaks the heading mirroring the extractions established.

**Baseline:** the four guards before=exit 0 after=exit 0.

**Commit:** `refactor(myflow-contracts): sweep inline rationale into the appendices`

---

### 8 Lower the budgets and prove nothing was lost

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh`
- Modify: `scripts/test-check-cleanup-complete.sh` (fixed a hardcoded `pipeline.md` path task 5's
  move broke)
- Modify: `openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md`
- Modify: `openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/design.md`

**Interfaces:**
- Consumes: tasks 1-7, complete.
- Produces: the evidence the review panel reads alongside the diff.

- [x] **Step 1: Lower `pipeline.md`'s and `pipeline-rationale.md`'s budget rows**

To each file's actual landing size plus 25%. **Left at 62862 and 28107, the ratchet silently permits
regrowth to today's size** — which is the whole failure this change exists to stop, reintroduced by
omission.

- [x] **Step 2: Diff the normative inventory**

```bash unverified:normative-after.txt does not exist until this step runs
scripts/check-normative-inventory.sh > /tmp/kan-295-normative-after.txt
diff openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/normative-baseline.txt \
     /tmp/kan-295-normative-after.txt
```

The inventory is sorted and whitespace-normalised, so a reflowed paragraph reads as unchanged and a
moved sentence reads as unchanged. **A difference is resolved by restoring the sentence, never by
accepting the new inventory.**

- [x] **Step 3: Diff the sentence set**

```bash unverified:the after-set does not exist until this step runs
python3 /tmp/kan-295-sentence-set.py \
  skills/myflow-contracts/pipeline.md \
  skills/myflow-contracts/pipeline-rationale.md \
  skills/myflow-contracts/git-boundaries.md \
  skills/myflow-contracts/git-boundaries-rationale.md \
  skills/myflow-contracts/model-policy.md \
  skills/myflow-contracts/model-policy-rationale.md \
  skills/myflow-contracts/artifacts-registry.md \
  skills/myflow-contracts/artifacts-registry-rationale.md \
  skills/myflow-contracts/session-records.md \
  skills/myflow-contracts/session-records-rationale.md \
  skills/myflow-contracts/worktree-resolution.md \
  skills/myflow-contracts/worktree-resolution-rationale.md \
  > /tmp/kan-295-sentence-after.txt
comm -23 openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/sentence-baseline.txt \
     /tmp/kan-295-sentence-after.txt
```

`comm -23` prints every baseline sentence absent from the after-set. **Each one is either restored
or justified by name in `verification.md`.** A sentence a citation repoint legitimately changed is a
justified residue; a sentence nothing accounts for is a cut, and is restored.

This check does **not** substitute for the ledger, which the ledger requirement forbids: it is exact
for a pure move and blind to a rule extraction, which authors a line that did not previously exist.
Both run.

- [x] **Step 4: Verify the installer needs no change**

```bash verified:this is .myflow/project.md's ## run recipe for exercising the installer
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
ls "$SANDBOX/.claude/skills/myflow-contracts/" \
   "$SANDBOX/.cursor/skills/myflow-contracts/" \
   "$SANDBOX/.codex/skills/myflow-contracts/"
```

All ten new files must be present in all three harnesses with no edit to `setup.sh`. If any is
absent, that is a real installer gap — fix `setup.sh` and say so; the non-goal was an *unnecessary*
edit, not a refusal to make a necessary one.

- [x] **Step 5: Record the measured result**

`verification.md` and `design.md` carry the actual before/after byte counts per command, replacing
`design.md`'s predicted figures and retagging them `measured:`. A prediction left in place after the
measurement exists is a stale number nothing checks.

- [x] **Step 6: Full lint and test**

Every command in `.myflow/project.md`'s `## lint`, then `## test`. Split the test run across more
than one invocation — measured at roughly 144s total against a 120000ms tool timeout.

**Tests:** no new test file; the two baseline diffs and the sandboxed installer run are this task's
verification.

**Regression:** Reverting this task leaves the budgets ratchet permitting `pipeline.md` to regrow to
50 KB, and removes the only evidence that the move dropped nothing.

**Baseline:** `check-normative-inventory.sh` before=exit 0 after=exit 0, inventories byte-identical;
`comm -23` before=n/a after=empty or fully justified; every `## lint` guard before=exit 0
after=exit 0.

**Commit:** `chore(scripts): lower the pipeline budgets and record the move's evidence`
