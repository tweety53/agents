> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Apply the existing `myflow-contract-economy` partition machinery (rule extraction,
per-move ledger, budget re-anchor) to the pipeline files KAN-95 left untouched, consolidate the
"named options + marked recommendation + safe default" pattern into one contract, remove three
duplicated procedures, and fold in two approved `/myflow-fast` behavior changes.

**Architecture:** No new partition mechanism. Every eviction/move follows
`myflow-contract-economy`'s existing rules verbatim: a passage that is pure justification moves to
its file's existing `-rationale.md` (all three `SKILL-rationale.md` siblings for `myflow-do`,
`myflow-finish`, `myflow-start` already exist; `myflow-fast/SKILL-rationale.md` is created here); a
mixed passage is handled by rule extraction under the bounded carve-out (original moves verbatim,
core gains a stated rule, the pair is quoted in that task's per-move ledger). New procedural
dedup — `operator-prompts.md`, `commit-split.sh`, `prepare-workspace.sh` — is new shared surface,
not a split of an existing file.

**Tech Stack:** Markdown skills/contracts, Bash guard scripts, the `openspec` CLI. No runnable
application in this repository; verification is grep, the guard scripts under `## lint`/`## test`
in `.myflow/project.md`, and reading the diff.

## Global Constraints

- **Per-move ledger, mandatory.** Every task that evicts, moves, or extracts prose from a file the
  pipeline loads emits the four-column ledger table (`myflow-contract-economy`'s "A move or
  eviction is recorded in a per-move ledger") in its own output, before the diff reaches the review
  panel. A deletion in the diff with no ledger row is a review-panel finding.
- **Verbatim partition + bounded carve-out only.** A moved passage is byte-for-byte unchanged
  except the two permitted edits (repointing a citation, deleting a stale position word). A mixed
  passage is handled only by rule extraction (original to the appendix verbatim; core gains a
  stated rule, not a restated argument; the ledger quotes both).
- **No new capability, no rule change** to `myflow-contract-economy` itself — every task here
  *applies* its existing requirements.
- **`myflow-fast-command`'s two requirements change** (proposal's Modified Capabilities) — Task 3
  is where that spec delta becomes true in the skill file.
- **Item G (table-driving the oversized guard test suites) is out of scope for this plan.**
  `/myflow-finish`'s integrate run files it as a follow-up Jira issue; no task here attempts it.
- **No suppression markers, no guard weakening.** A lint hit is fixed by editing the offending
  line.
- **Budget re-anchor is the last content task (Task 12), after every other task lands** — per
  `myflow-contract-economy`'s "regenerated once, as the last action" rule. No earlier task edits
  `scripts/check-contract-budget.sh`.

## Baseline

Files this plan edits most, measured before any task runs:

<!-- verified: wc -c on each path, working tree at the start of this change -->

| File | Bytes now | Budget in `budgets()` |
|------|-----------|------------------------|
| `skills/myflow-contracts/pipeline.md` | 34347 | 40161 |
| `skills/myflow-do/SKILL.md` | 44816 | 44781 |
| `skills/myflow-finish/SKILL.md` | 27405 | 32050 |
| `skills/myflow-start/SKILL.md` | 28352 | 32465 |
| `skills/myflow-fast/SKILL.md` | 10616 | 12671 |

**`skills/myflow-do/SKILL.md` is already 35 bytes over its declared budget of 44781.** This plan's
own edits to it (Tasks 2, 6, 8, 10) are net-negative (evicting more than they add), so Task 12
re-anchors its row from whatever the file measures after Task 11 — not from today's already-over
figure. A task that finds itself *adding* net bytes to this file before Task 12 runs is a signal to
re-check that task's own eviction accounting, not a reason to pre-emptively raise the budget here.

`skills/myflow-contracts/operator-prompts.md` and `skills/myflow-fast/SKILL-rationale.md` are new;
they get budget rows in Task 12 from their own measured size, not estimated in advance.

---

### 1 New contract: `skills/myflow-contracts/operator-prompts.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/operator-prompts.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the shared contract Task 4 points the five call sites at.

- [x] **Step 1: Write the shape**

State the shape every operator-facing prompt in this corpus already follows (drawn from
`myflow-planning-gate`'s existing "Every approval or choice is offered as options, not as open
prose" requirement — this file consolidates that requirement's prose pattern, it does not add a new
rule):

```markdown unverified:authored in-tree for this change; matches the shape myflow-planning-gate's existing requirement already states in prose
A prompt in this shape states:
- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent — the safe default, always the recommended option
- a ⚠ marker in the handoff when that silent default actually fired
```

- [x] **Step 2: State the doctrine line**

Add: "Every call site below cites this contract for the mechanics and states only its own
question text and options — never the shape itself."

- [x] **Step 3: Verify**

```bash unverified:no existing guard checks a brand-new file's content; this confirms only that it was created and is discoverable
test -f skills/myflow-contracts/operator-prompts.md
grep -n "myflow-contracts/SKILL.md" skills/myflow-contracts/SKILL.md
```

Expected: the file exists. (Task 4's own verify step confirms the five call sites cite it; this
step only confirms creation.)

**Tests:** none — new contract file, no executable surface. Verified by the file's existence and by
Task 4's call-site citations resolving via the check-references guard.
**Regression:** not applicable — no executable behavior to revert.
**Baseline:** not applicable — no test harness for a prose contract file.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): add the operator-prompts contract`

---

### 2 Anti-restatement boilerplate collapse — contracts and three pipeline `SKILL.md` files

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`, `skills/myflow-contracts/pipeline-rationale.md`,
  `skills/myflow-contracts/jira-integration.md`, `skills/myflow-contracts/handoff-blocks.md`,
  `skills/myflow-contracts/handoff-blocks-rationale.md`,
  `skills/myflow-contracts/project-configuration.md`,
  `skills/myflow-contracts/project-configuration-rationale.md`,
  `skills/myflow-contracts/workspace-isolation.md`,
  `skills/myflow-contracts/workspace-isolation-rationale.md`, `skills/myflow-do/SKILL.md`,
  `skills/myflow-do/SKILL-rationale.md`, `skills/myflow-finish/SKILL.md`,
  `skills/myflow-finish/SKILL-rationale.md`, `skills/myflow-start/SKILL.md`,
  `skills/myflow-start/SKILL-rationale.md` — the exact set located by Step 1's grep (not
  pre-enumerated before the task ran; recorded here after the fact from that task's own report)

**Interfaces:**
- Consumes: nothing from earlier tasks (Task 1's contract is a citation target for a *different*
  pattern — the operator-prompts shape — and is unrelated to this task's doctrine-line pattern).
- Produces: the collapsed citation form (`see <Section> (<file>.md)`) every later task's own edits
  in these files should match.

- [x] **Step 1: Locate every occurrence**

```bash verified:run against the working tree at task start
grep -rn "is canonical\|not restated here\|see the rationale\|never act on a remembered\|Never act on a remembered" skills/myflow-contracts/*.md skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md skills/myflow-start/SKILL.md
```

Read each hit. A hit that is *itself* the one-time doctrine statement this task is about to add
(none yet exist) is not a collapse target; every other hit is a per-citation restatement.

- [x] **Step 2: Add the doctrine line to each file's top**

Near the top of each touched file (after its title/purpose paragraph, before its first `##`
section), add: "Every citation below is canonical at its target. Never restate its content here
and never act on a remembered version of it — read it fresh each time it is needed."

- [x] **Step 3: Rule-extract each per-citation paragraph**

For each hit from Step 1: the original paragraph moves to that file's `-rationale.md`, appended
under the heading it sat under in the core (mirroring the appendix's existing heading tree — add
the heading there if the mirror is currently empty-bodied). The core keeps only the bare citation
form (`see Model policy (pipeline.md)` or equivalent) at that site.

- [x] **Step 4: Emit the per-move ledger**

One row per paragraph moved in Step 3: Removed passage (first 8 words), Source heading,
Destination (`<file>-rationale.md` + heading), Pointer left (the bare citation form left in the
core).

- [x] **Step 5: Verify**

```bash verified:both declared under ## lint in .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

```bash unverified:confirms no doctrine-eligible hit was missed; exact zero-count depends on Step 1's actual hit list
grep -rn "is canonical\|not restated here\|see the rationale\|never act on a remembered" skills/myflow-contracts/*.md skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md skills/myflow-start/SKILL.md | grep -v "Every citation below is canonical"
```

Expected: both guards exit 0; the second grep returns nothing beyond the doctrine lines themselves.

**Tests:** none — prose/citation-only change. Verified by the guards above and by the ledger being
checked against the diff at review.
**Regression:** not applicable — no test harness for prose content; `check-references.sh` is the
regression signal (a broken citation fails it).
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): collapse anti-restatement boilerplate in the contracts and three pipeline skills`

---

### 3 `myflow-fast`: boilerplate collapse, silent defaults, dropped design-approval gate

**Build:** green

**Files:**
- Modify: `skills/myflow-fast/SKILL.md`
- Create: `skills/myflow-fast/SKILL-rationale.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the two `myflow-fast-command` requirement changes the proposal's delta spec declares
  (Modified Capabilities), and the rationale sibling Task 5 may also append to.

- [x] **Step 1: Boilerplate collapse in this file**

Same mechanism as Task 2 Step 1–4, scoped to `skills/myflow-fast/SKILL.md` alone, appending moved
paragraphs to the new `SKILL-rationale.md` (Step 2 creates it first if a paragraph needs to move
before the file otherwise would exist — order Steps so the file exists before anything is appended
to it).

- [x] **Step 2: Create `SKILL-rationale.md` with the mirrored heading tree**

Per `myflow-contract-economy`'s "An appendix mirrors its core's heading tree" — every `##`/`###`
heading in `SKILL.md` gets a same-named, same-order heading here, empty-bodied where nothing moves
to it.

- [x] **Step 3: Rewrite "Recorded defaults favor speed" → silent, no ask**

Replace the section's current text (which describes four `AskUserQuestion` calls, two of them with
an overridden recommendation) with: the question round is **not asked** on a creating run;
`planningEffort: default`, `models.implementation: sonnet`, `models.reviewPanel: sonnet`,
`models.panelFix: sonnet`, and `reviewPanelRoster: light` are recorded directly. An explicit
session instruction naming a different value for one of these fields still overrides it, recorded
with the dispatch exactly as `pipeline.md`'s Model policy section already permits for any operator
override.

- [x] **Step 4: Update the "No state file" section's bullet list**

The bullet citing "Ask the planning effort, the models, and the review panel roster — creating runs
only" (`skills/myflow-start/SKILL.md`) currently says "the same four questions, exactly as
written." Change it to state this skill does **not** run that question round — it records the
defaults per Step 3 instead — while every other cited section (A, B, C, D) still runs exactly as
written.

- [x] **Step 5: Drop the design-approval confirm**

In the same "No state file" section, the citation to brainstorming (workflow #1) currently says
"run in full, including its design-approval gate; brainstorming stays fully interactive here,
exactly as under `/myflow-start`, with no auto-answering." Rewrite: the clarifying questions and
design presentation stay fully interactive, exactly as under `/myflow-start` — but the separate
explicit post-design confirm is **not** run here; `/myflow-fast` proceeds directly into artifact
creation once the design is presented, unless the operator raises an objection during that
presentation. State plainly that this is a scoped override of `superpowers:brainstorming`'s hard
gate, `/myflow-fast` only, and that `/myflow-start` is unaffected.

- [x] **Step 6: Update the Guardrails list**

Remove "Never continue from brainstorming into implementation before the design is approved" (now
false) and replace with a line stating the design-approval confirm is not run here (clarifying
questions and presentation remain interactive; the confirm specifically is skipped). Add a line
stating the setup question round is never asked on a creating run.

- [x] **Step 7: Write the rationale for both changes into `SKILL-rationale.md`**

Under the two sections touched in Steps 3–6, record why: the operator flagged the interactive
round as friction against a speed-oriented command; the design-approval confirm is replaced by the
downstream `IN_PROGRESS` staged-diff review as the actual checkpoint; both are scoped to
`/myflow-fast` only, `/myflow-start` is deliberately unchanged.

- [x] **Step 8: Verify**

```bash unverified:confirms the two rewritten sections no longer describe an ask, and the guardrails list no longer states the removed rule
grep -n "recommended\|Recommend" skills/myflow-fast/SKILL.md
grep -n "AskUserQuestion" skills/myflow-fast/SKILL.md
```

Expected: no remaining reference to asking the four setup questions or to a separate design-approval
`AskUserQuestion` call; the file still cites `myflow-start/SKILL.md`'s clarifying-question and
design-presentation sections as run in full.

```bash verified:declared under ## lint
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: `check-references.sh` exits 0. `check-contract-budget.sh` fails here — expected, since
`SKILL-rationale.md` has no row yet; Task 12 adds it. Do not add the row in this task.

**Tests:** none — prose/behavior-description change in a skill file, no executable surface added.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `feat(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): myflow-fast silent setup defaults and dropped design-approval gate`

---

### 4 Adopt `operator-prompts.md` at the five call sites

**Build:** green

**Files:**
- Modify: `skills/myflow-start/SKILL.md` (convergence confirm), `skills/myflow-start/SKILL-rationale.md`
- Modify: `skills/myflow-do/SKILL.md` (review panel handback, §5; fix-documenting ask, §3)
- Modify: `skills/myflow-finish/SKILL.md` (run-1 unfinished-work gate; self-review skip prompt — the
  fifth site, located in this file's step 8, not `finish-contract.md`)
- Modify: wherever the self-review skip prompt is currently inlined (located by Step 1)

**Interfaces:**
- Consumes: `skills/myflow-contracts/operator-prompts.md` (Task 1).
- Produces: five call sites shrunk to question text + options + citation.

- [x] **Step 1: Locate the fifth call site**

```bash unverified:the self-review skip prompt's exact home file is not confirmed from this session's reading; locate it before editing
grep -rln "self-review" skills/myflow-do/SKILL.md skills/myflow-start/SKILL.md skills/myflow-contracts/*.md
```

Read the hits to find the actual skip-prompt text (distinct from the self-review *procedure*
described elsewhere).

- [x] **Step 2: Rewrite each of the five sites**

For each site: keep the question text and its named options verbatim (these are call-site-specific
and not part of the shared shape); remove the inline restatement of "what happens if silent" and
"mark the recommendation" mechanics; add one citation: `— shape per Operator prompts
(skills/myflow-contracts/operator-prompts.md)`.

- [x] **Step 3: Move call-site-specific rationale, if any, to that file's own `-rationale.md`**

Where a site's prose beyond the mechanics is *why this particular default* was chosen (not part of
the shared shape), it stays at the site if it's load-bearing, or moves to that file's
`-rationale.md` if it's pure justification — same lever-1 test as Task 2.

- [x] **Step 4: Emit the per-move ledger** for anything moved in Step 3.

- [x] **Step 5: Verify**

```bash verified:declared under ## lint
scripts/check-references.sh
```

```bash unverified:confirms all five sites now cite the shared contract
grep -rn "operator-prompts.md" skills/myflow-start/SKILL.md skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md
```

Expected: `check-references.sh` exits 0; the citation grep shows at least four hits in these three
files (the fifth site's file is confirmed by Step 1 and may be one of these three or elsewhere).

**Tests:** none — prose-only change.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): adopt the operator-prompts contract at its five call sites`

---

### 5 Move remaining justification prose to existing rationale appendices

**Build:** green

**Files:**
- Modify: `skills/myflow-start/SKILL.md`, `skills/myflow-start/SKILL-rationale.md`
- Modify: `skills/myflow-do/SKILL.md`, `skills/myflow-do/SKILL-rationale.md`
- Modify: `skills/myflow-finish/SKILL.md`, `skills/myflow-finish/SKILL-rationale.md`
- Modify: `skills/myflow-contracts/pipeline.md`, `skills/myflow-contracts/pipeline-rationale.md`

**Interfaces:**
- Consumes: nothing from earlier tasks (independent lever-1 pass; ordered after Task 2 per the
  ticket's own stated sequencing, not because it depends on Task 2's output).
- Produces: the trimmed files Task 6 (report-only prompt compression) depends on being already
  slimmed, per the ticket's "sequencing: this item runs after items 1 and 2" note.

- [x] **Step 1: `myflow-start`'s `## Convergence`**

Apply the core/rationale behavioral test to the section. Core keeps: loop until an explicit "move
on"; rounds 1–2 open silently, round 3+ is offered; an unanswerable question is recorded rather than
guessed. Everything defending those choices (why the confirm and the offer recommend opposite
courses, the tuned-threshold history) moves to `SKILL-rationale.md` under the same heading.

- [x] **Step 2: Sweep the other three files for the same pattern**

```bash unverified:the exact passage list is not pre-enumerated; located by reading each file section by section against the core/rationale behavioral test ("would removing it change what an agent does?")
```

Read `myflow-do/SKILL.md`, `myflow-finish/SKILL.md`, and `pipeline.md` section by section. A
section whose removal would change no agent behavior — measurement narrative, design history,
argument against a rejected alternative — moves to that file's `-rationale.md`. A section stating an
obligation, a table, a code block an agent runs, or a step list stays in the core.

- [x] **Step 3: Emit the per-move ledger** for every passage moved in Steps 1–2.

- [x] **Step 4: Verify**

```bash verified:declared under ## lint
scripts/check-references.sh
```

Expected: exits 0; every appendix's heading tree still mirrors its core's (spot-check with a diff
of `grep -n "^##"` between each core/appendix pair).

**Tests:** none — prose-only change.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): move remaining justification prose to the existing rationale appendices`

---

### 6 Stop mirroring `check-unfinished-work.sh`'s grammar in `myflow-do`

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing consumed later.

- [x] **Step 1: Confirm the guard already reports its own reject reasons**

```bash verified:confirmed by reading scripts/check-unfinished-work.sh during design
grep -n "add \"" scripts/check-unfinished-work.sh
```

Expected: multiple `add "..."` calls, each a human-readable reject reason (marker span, malformed
marker line, etc.) — already true, this step is a re-confirmation before editing, not new work.

- [x] **Step 2: Trim §5's marker-parse rules to the emit format only**

Keep only what a caller writing a marker line needs to produce (the format itself). The parse/
rejection rules that mirror the guard's own logic move to `SKILL-rationale.md`, or are dropped
outright where they add nothing beyond what the guard's own error text already says (dropping, not
moving, is correct here per `myflow-contract-economy`'s "rule extraction ... not ... a licence to
shorten a passage that carries no argument" — a passage that is *only* a duplicate of the guard's
own text carries no argument of its own to preserve).

- [x] **Step 3: Replace §7's exit-code table with a citation**

Replace the prose table with: "run the guard and read its own output" plus a citation to the
guard's own header comment, which already states its exit-code contract.

- [x] **Step 4: Emit the per-move ledger** for anything moved (not dropped) in Steps 2–3.

- [x] **Step 5: Verify**

```bash verified:declared under ## lint
scripts/check-references.sh
```

```bash unverified:confirms the exit-code table is gone from section 7's prose
grep -n "exit 0\|exit 1\|exit 2" skills/myflow-do/SKILL.md
```

Expected: `check-references.sh` exits 0; the exit-code grep shows no standalone table describing
`check-unfinished-work.sh`'s codes (a citation to the guard's own header is fine).

**Tests:** none — prose-only change; the guard's own behavior is unchanged.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): stop mirroring check-unfinished-work.sh's grammar in myflow-do`

---

### 7 Dedup the empty-worktree-set stop

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` (sections 2 and 7)
- Modify: `skills/myflow-finish/SKILL.md` (both occurrences)

**Interfaces:**
- Consumes: `pipeline.md`'s existing "Resolving a change's worktrees" section (already the rule's
  home per KAN-95 part 4 — not edited by this task).
- Produces: nothing consumed later.

- [x] **Step 1: Locate the three restatements**

```bash verified:run against the working tree at task start
grep -n "empty resolved set\|comes back empty\|worktree.*never a vacuous" skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md
```

- [x] **Step 2: Replace each with a citation**

Each site keeps its own local consequence (what the calling step does — stop and report, ask the
operator) but cites "Resolving a change's worktrees" (`skills/myflow-contracts/pipeline.md`) for the
shared reasoning instead of restating it.

- [x] **Step 3: Emit the per-move ledger** for each restatement removed.

- [x] **Step 4: Verify**

```bash verified:declared under ## lint
scripts/check-references.sh
```

```bash unverified:confirms exactly one restatement source remains (pipeline.md itself) and the three call sites now cite it
grep -c "empty resolved set is never a vacuous" skills/myflow-contracts/pipeline.md skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md
```

Expected: `check-references.sh` exits 0; the phrase (or its equivalent) appears in `pipeline.md`
and not restated in the other two.

**Tests:** none — prose-only change.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): dedup the empty-worktree-set stop behind a citation`

---

### 8 `scripts/commit-split.sh` — dedup the two-commit chain

**Build:** green

**Files:**
- Create: `scripts/commit-split.sh`
- Create: `scripts/test-commit-split.sh`
- Modify: `skills/myflow-do/SKILL.md` (§7 PR-exception path)
- Modify: `skills/myflow-finish/SKILL.md` (§1.2)
- Modify: `.myflow/project.md` (`## test` list)

**Interfaces:**
- Consumes: `pipeline.md`'s existing Git boundaries section — the guarded bash chain there is the
  spec this script implements; that section is **not** rewritten to a citation (it stays the
  canonical spec of the chain's behavior, per Task design decision: a spec is not a restatement).
- Produces: `commit-split.sh <worktree> <name> <impl-msg> <plan-msg>`, called by both remaining
  sites.

- [x] **Step 1: Write `scripts/commit-split.sh`**

```bash unverified:authored in-tree for this change; wraps the exact guarded chain already specified in pipeline.md's Git boundaries section
#!/usr/bin/env bash
set -euo pipefail
worktree="$1" name="$2" impl_msg="$3" plan_msg="$4"
git -C "$worktree" reset -q -- openspec/ docs/superpowers/
git -C "$worktree" add -A -- . ':(exclude)openspec/' ':(exclude)docs/superpowers/'
git -C "$worktree" diff --cached --quiet \
  || git -C "$worktree" commit -m "$impl_msg"
git -C "$worktree" add -A
git -C "$worktree" diff --cached --quiet \
  || git -C "$worktree" commit -m "$plan_msg"
```

A planning path that is a tracked symlink SHALL stop the script rather than being worked around —
`git add` on a symlinked pathspec exits 128 with a message naming the path; the script propagates
that exit code and message rather than catching and reinterpreting it (`set -e` already gives this
for free; no additional handling is added).

- [x] **Step 2: Write `scripts/test-commit-split.sh`**

Cases: (1) both commits happen when both staging areas have changes; (2) the implementation commit
is skipped when only planning paths changed (guarded-empty-commit skip); (3) the planning commit is
skipped when only implementation paths changed; (4) a tracked symlink at `openspec/` stops the
script with exit 128 and neither commit is made.

- [x] **Step 3: Replace both call sites**

`myflow-do` §7's PR-exception path and `myflow-finish` §1.2 both currently spell out the guarded
bash chain inline. Replace each with a call to `scripts/commit-split.sh <worktree> <name>
"<impl-msg>" "<plan-msg>"`, keeping each site's own message-text derivation (they differ) as prose
around the call.

- [x] **Step 4: Add the test entry**

Add `scripts/test-commit-split.sh` to `.myflow/project.md`'s `## test` list, alphabetically ordered
among the existing entries.

- [x] **Step 5: Emit the per-move ledger** for the two inline chains removed from the skill files.

- [x] **Step 6: Verify**

```bash unverified:new script and harness, run for the first time in this task
chmod +x scripts/commit-split.sh
scripts/test-commit-split.sh
```

```bash verified:declared under ## lint
scripts/check-references.sh
```

Expected: the new harness's 4 cases pass; `check-references.sh` exits 0.

**Tests:** Case 1: both commits happen; Case 2: implementation commit skipped when empty; Case 3:
planning commit skipped when empty; Case 4: tracked symlink at `openspec/` stops the script, exit
128, no commit made.
**Regression:** Case 4 (symlink stop): removing the `set -e` propagation and substituting a caught-
and-continued path would let a symlinked planning path silently land in the implementation commit —
this is the one failure mode `pipeline.md`'s Git boundaries section names as never worked around.
**Baseline:** before=0 after=4 cases in `scripts/test-commit-split.sh` (new harness).
**Commit:** `feat(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): add commit-split.sh and dedup the two-commit chain`

---

### 9 Dedup the `planningEffort` retired-key fallback

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `skills/myflow-start/SKILL.md`, `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing consumed later.

- [x] **Step 1: Locate the three explanations**

```bash verified:run against the working tree at task start
grep -n "planningEffort" skills/myflow-start/SKILL.md skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md skills/myflow-contracts/state-file.md
```

- [x] **Step 2: Confirm or write the single description in `state-file.md`**

`state-file.md` already carries the retired-key fallback as part of its state-shape documentation
(it is the canonical file for the shape). Confirm the description there is complete enough that the
other two sites need only cite it — extend it if a detail currently exists only at a call site and
not there.

- [x] **Step 3: Replace the other two explanations with a citation**

`skills/myflow-do/SKILL.md` and `skills/myflow-finish/SKILL.md` (the two that are not `state-file
.md` itself) keep only "read the level through the retired-key fallback, per **State file**
(`skills/myflow-contracts/state-file.md`)" at their sites.

- [x] **Step 4: Emit the per-move ledger** for both replaced explanations.

- [x] **Step 5: Verify**

```bash verified:declared under ## lint
scripts/check-references.sh
```

```bash unverified:confirms the fallback is explained in full in exactly one place
grep -c "retired.key\|retired key" skills/myflow-contracts/state-file.md skills/myflow-start/SKILL.md skills/myflow-do/SKILL.md skills/myflow-finish/SKILL.md
```

Expected: `check-references.sh` exits 0; `state-file.md` carries the full explanation, the other
three carry only a citation (a bare mention of "retired key" in a citation sentence is expected and
is not a restatement).

**Tests:** none — prose-only change.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): dedup the planningEffort retired-key fallback into state-file.md`

---

### 10 `scripts/prepare-workspace.sh`

**Build:** green

**Files:**
- Create: `scripts/prepare-workspace.sh`
- Create: `scripts/test-prepare-workspace.sh`
- Modify: `skills/myflow-do/SKILL.md` (§7)
- Modify: `.myflow/project.md` (`## test` list)

**Interfaces:**
- Consumes: `workspace-isolation.md`'s existing derivation rules (spec, not rewritten — same
  treatment as Task 8's relationship to `pipeline.md`'s Git boundaries section).
- Produces: `prepare-workspace.sh <worktree>`, called by `myflow-do` §7.

- [x] **Step 1: Write `scripts/prepare-workspace.sh`**

Wraps exactly what `myflow-do` §7 currently does in prose: run `check-workspace-isolation.sh`
against the worktree (fail loudly and stop if it fails); derive and export the workspace variables
per `workspace-isolation.md`'s existing rules; print what was exported, one `KEY=value` line per
variable, so the caller's remaining prose (lint/test/staging/state-write) can read "run the script,
then continue with what it exported" instead of carrying the derivation rules inline.

- [x] **Step 2: Write `scripts/test-prepare-workspace.sh`**

Cases: (1) a project declaring no `## workspace isolation` section prints nothing and exits 0
(mirrors `check-workspace-isolation.sh`'s own no-op case); (2) a project declaring the section
exports the derived variables and prints them; (3) `check-workspace-isolation.sh` failing stops the
script before any export, non-zero exit.

- [x] **Step 3: Rewrite `myflow-do` §7**

Replace the workspace-isolation-guard-plus-export prose with: run `prepare-workspace.sh
<worktree>`, source or read its printed exports, then continue with lint/test/staging/state-write
exactly as before (those four jobs are unaffected — only the guard+export pair moves into the
script).

- [x] **Step 4: Add the test entry** to `.myflow/project.md`'s `## test` list, alphabetically
ordered.

- [x] **Step 5: Emit the per-move ledger** for the prose removed from §7.

- [x] **Step 6: Verify**

```bash unverified:new script and harness, run for the first time in this task
chmod +x scripts/prepare-workspace.sh
scripts/test-prepare-workspace.sh
```

```bash verified:declared under ## lint
scripts/check-references.sh
scripts/check-workspace-isolation.sh
```

Expected: the new harness's 3 cases pass; both guards exit 0 (this repository declares no
`## workspace isolation` section, so `check-workspace-isolation.sh` reports its existing no-op
pass, per `.myflow/project.md`'s own documented note).

**Tests:** Case 1: no `## workspace isolation` section — no-op, exit 0; Case 2: section declared —
variables exported and printed; Case 3: `check-workspace-isolation.sh` failure stops before export.
**Regression:** Case 3 (guard failure stops before export): removing the fail-fast check would let
a workspace's variables export even when the isolation guard rejects the project's declaration —
the one failure this script exists to prevent silently.
**Baseline:** before=0 after=3 cases in `scripts/test-prepare-workspace.sh` (new harness).
**Commit:** `feat(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): add prepare-workspace.sh and fold myflow-do section 7's guard+export into it`

---

### 11 Compress the report-only subagent prompts

**Reverted.** Implemented, reviewed clean at the per-task pass, then reverted after the whole-branch
panel's re-check (Principles reviewer) found the task's own premise wrong: `principles-reviewer-
prompt.md` and `adversarial-reviewer-prompt.md` are persisted, repo-committed dispatch templates —
read by every future `/myflow-do` run — not the subagent's transient chat output the caveman
Boundaries rule licenses compressing. No other file in this change was compressed this way, which
was itself the signal. Both files were restored byte-for-byte to their pre-task content and the
task's commit was dropped from history entirely (`git rebase -i` with the commit marked `drop`) —
this task lands no diff and is excluded from the final branch.

**Build:** green

**Files:**
- Modify: `skills/myflow-do/principles-reviewer-prompt.md`
- Modify: `skills/myflow-do/adversarial-reviewer-prompt.md`
- Modify: `skills/myflow-do/SKILL.md` (locate-and-report dispatches in §4 only)

**Interfaces:**
- Consumes: Task 5's slimmed `myflow-do/SKILL.md` (this task's own note: runs after the meta-prose
  and justification-prose passes, per the ticket's stated sequencing, so what remains here is
  load-bearing content worth compressing carefully rather than duplication that would otherwise
  hide inside it).
- Produces: nothing consumed later.

- [x] **Step 1: Compress the two prompt files**

Caveman-style compression (technical substance intact; articles, filler, and pleasantries
dropped) — these are report-only prompts whose subagent's output is chat, not a persisted artifact,
per the caveman Boundaries rule already in force corpus-wide.

- [x] **Step 2: Locate and compress the locate-and-report dispatches in `myflow-do` §4**

```bash unverified:the exact dispatch clauses eligible for this treatment are not pre-enumerated from this session's reading of section 4; located by reading the section against the ticket's own exclusion list
```

A dispatch clause is eligible only if it is locate-and-report (the subagent's output is read and
discarded, never persisted). Skip anything matching the ticket's exclusion list: implementer
dispatch clauses (contracts, must be read exactly), negation-built prohibition text, guard grammar
compared byte-for-byte, and anything that becomes a handoff, commit message, spec, or Jira text.

- [x] **Step 3: Verify**

```bash unverified:confirms no persisted-artifact-producing clause was compressed; read each touched clause and check it does not appear on the exclusion list
```

Read the diff against the exclusion list above, clause by clause. This is a judgment check, not a
mechanical one — no guard scans for compression correctness.

**Tests:** none — prompt-text compression, no executable surface.
**Regression:** not applicable.
**Baseline:** not applicable.
**Commit:** `docs(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): compress the report-only reviewer prompts`

---

### 12 Budget guard re-anchor

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh` (the `budgets()` table)

**Interfaces:**
- Consumes: the final measured size of every file Tasks 1–11 touched or created.
- Produces: nothing consumed later — this is the last content task.

- [x] **Step 1: Measure every touched or created file**

```bash unverified:exact figures depend on Tasks 1-11's actual final content
for f in skills/myflow-contracts/operator-prompts.md skills/myflow-contracts/pipeline.md skills/myflow-contracts/pipeline-rationale.md skills/myflow-contracts/jira-integration.md skills/myflow-do/SKILL.md skills/myflow-do/SKILL-rationale.md skills/myflow-finish/SKILL.md skills/myflow-finish/SKILL-rationale.md skills/myflow-start/SKILL.md skills/myflow-start/SKILL-rationale.md skills/myflow-fast/SKILL.md skills/myflow-fast/SKILL-rationale.md skills/myflow-contracts/state-file.md; do printf '%s %s\n' "$f" "$(wc -c < "$f")"; done
```

Include every other `skills/myflow-contracts/*.md` core Task 2 Step 1 actually found and edited —
that hit list, not re-derived here.

- [x] **Step 2: Add rows for the two new files**

`operator-prompts.md` and `myflow-fast/SKILL-rationale.md` each get a new row: their Step 1 measured
size, plus 25%.

- [x] **Step 3: Re-anchor every touched file's existing row**

Each existing row for a file this plan touched is re-anchored to its Step 1 measured size, plus
25% — not raised speculatively, not left at its pre-change figure.

- [x] **Step 4: Verify**

```bash verified:declared under ## lint and ## test
scripts/check-contract-budget.sh
scripts/test-check-contract-budget.sh
```

Expected: both exit 0.

**Tests:** none — configuration-table update, covered by the existing test-check-contract-budget
harness (unmodified by this task).
**Regression:** not applicable — this task adds no new guard logic, only table data.
**Baseline:** not applicable.
**Commit:** `chore(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): re-anchor the contract budget table`

---

### 13 Full guard and test sweep

**Build:** green

**Files:**
- None (verification only; no source changes)

**Interfaces:**
- Consumes: every prior task's final state.
- Produces: nothing — terminal task.

- [x] **Step 1: Run every declared lint command**

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

- [x] **Step 2: Run every declared test command**

```bash verified:declared under ## test in .myflow/project.md, plus the two new harnesses this plan adds
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-check-task-build-green.sh
scripts/test-check-task-commit-fields.sh
scripts/test-check-workspace-isolation.sh
scripts/test-check-contract-budget.sh
scripts/test-check-vocabulary.sh
scripts/test-commit-split.sh
scripts/test-prepare-workspace.sh
```

Expected: every command exits 0.

**Tests:** none — this task adds no new test, it runs the full existing suite plus this plan's two
new harnesses.
**Regression:** not applicable.
**Baseline:** not applicable — this task's own verification *is* the baseline check for the whole
plan.
**Commit:** `chore(kan-106-slim-the-myflow-skills-cut-meta-prose-extract): full guard and test sweep`
