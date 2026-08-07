# Tasks — kan-87-cut-per-command-load-further

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** cut what each `/myflow-*` command loads per run, by moving the finish-only half of
`pipeline.md` into its own file and splitting the three pipeline skills into a core and a rationale
appendix.

**Architecture:** the same strict partition KAN-82 established — every line lands in exactly one
file, byte-for-byte, with exactly two permitted edits. What is new here is a **hoist before the
cut**: two passages inside the finish block are depended on by `/myflow-do`, and they are promoted
into the core before anything moves, so no citation ends up pointing at a file its reader never
loads.

**Repository:** one — `/Users/tweety53/Projects/agents`. No runnable application; verification is
the guard scripts declared in `.myflow/project.md` plus the assertions written into each task.

**Order:** group 1 hoists, because every later step depends on the core being complete first.
Group 2 makes the finish-only cut. Group 3 splits the three skills. Group 4 updates the index, the
rule layer and the entry points. Group 5 widens the guard and regenerates budgets **last**, because
its numbers are the sizes groups 1-4 produced.

**Global constraints, carried into every dispatch:**

- **This change moves text. It does not change a rule.** No requirement is added, removed or
  reworded; no `/myflow-*` command behaves differently afterwards. A dispatch that finds itself
  improving a rule's wording has left this plan's scope and must stop and report.
- **A moved line keeps its heading level, its wording and its wrapping.** Exactly **two** edits are
  permitted to a moved line, and no others:
  1. **repointing a citation** — its backticked path and the bold token beside it; and
  2. **deleting a position word** — `above` or `below` — that the move made false, because what it
     locates is now in a different file. A deletion, never a substitution.

  **Edit 1 may force a re-wrap, and that is permitted.** Replacing `**Finish contract**` with
  `**Preserving the session records**` lengthens the line by fifteen characters, past the corpus's
  wrap width — so the paragraph holding the citation may be re-wrapped. Confine it to that
  paragraph; its lines then appear in the line-multiset check as `<`/`>` pairs differing only by
  wrapping, and are accounted for exactly as the repointed line is. **The lines around a *moved*
  passage are still never re-wrapped** — that rule is what the paragraph-granularity constraint
  rests on, and this does not touch it. Task 1.1 hit this and it is recorded here so tasks 2.1 and
  3.1-3.3 do not each rediscover it.
- **The unit that moves is a whole blank-line-delimited paragraph, never part of one.** A paragraph
  mixing a rule with its justification stays in the core whole, and the lines around any moved
  passage are never re-wrapped. Both rules are inherited from KAN-82 and are stated in
  **The split is a verbatim partition, with citation repointing as its only edit**
  (`openspec/specs/myflow-contract-economy/spec.md`).
- **`scripts/check-references.sh` cannot verify what this change most needs verified.** It resolves
  a citation whenever a heading of that name exists, and heading mirroring guarantees one does. A
  green run is **not** evidence that a citation still reaches its substance. Every task that moves a
  cited section names its citations individually and checks each by reading.
- **`scripts/check-references.sh` is also line-scoped**: it checks a citation only when the bold
  token and the backticked path sit on the **same physical line**. Never split them, and verify by
  mutating the bold token in a scratch copy and confirming the guard fails.
- **No suppression markers, ever.** There is no auto-fix command in this repository.
- Every fenced block and numeric claim in a planning artifact carries its provenance tag, per
  **The four tags** (`skills/myflow-contracts/plan-provenance.md`). Every task carries one
  `**Build:**` tag, per **The build-green tag** (`skills/myflow-contracts/build-green.md`).
- **Every path in a command is absolute.** Resolve the apply worktree once from
  `git -C /Users/tweety53/Projects/agents worktree list`; `<WT>` below stands for it.

**Baseline, measured before any task runs:**

```text verified:authored in-tree from the wc output below
skills/myflow-contracts/pipeline.md   1348 lines   88253 bytes   every command
skills/myflow-do/SKILL.md              609 lines   40529 bytes   every /myflow-do run
skills/myflow-start/SKILL.md           493 lines   28944 bytes   every /myflow-start run
skills/myflow-finish/SKILL.md          467 lines   28086 bytes   every /myflow-finish run
```

<!-- measured: wc -l -c on each file @ 63fa6ee -->

---

## 1. Hoist what `/myflow-do` depends on

### 1.1 Promote the preserve-session-records outcome table into its own core section

**Files:**
- Modify: `<WT>/skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Produces: a new `##`-level section in `pipeline.md` holding the three-outcome table, which
  `skills/myflow-do/SKILL.md` and `skills/myflow-finish/SKILL.md` both cite.

The table beginning `| Outcome | What it means | What you do |`, together with the paragraph
**A non-zero exit is never silent and never a stop** and the paragraph
**That ordering is deliberate, and deliberately differs from `/myflow-do`'s**, currently sits inside
`## Finish contract` → `### Run 1 — the branch is not merged`. It is 1,951 bytes.

<!-- measured: python3 range extraction over pipeline.md @ 63fa6ee -->

`skills/myflow-do/SKILL.md` cites it as canonical for all three outcomes on its `prUrl` commit path,
and `/myflow-do` does not load a finish-only file — so it must not move out.

- [x] Cut those three passages out of `### Run 1 — the branch is not merged` and place them under a
      new `## Preserving the session records` section in `pipeline.md`, **byte for byte**. Position
      it immediately after `## Git boundaries`, which is the section its two callers already read.

- [x] Add one sentence under the new heading naming which commands read it — `/myflow-do` on its
      `prUrl` path and `/myflow-finish` run 1 — and stating that the script invocation itself is
      described by each caller. That sentence is newly authored; everything else is moved.

- [x] Repoint the citations in `<WT>/skills/myflow-do/SKILL.md` and
      `<WT>/skills/myflow-finish/SKILL.md` from `**Finish contract**` to
      `**Preserving the session records**`, keeping the bold token and the path on one line.

- [x] Verify, and prove the guard actually looked:

```bash unverified:run in the apply worktree after the hoist
cd "<WT>"
BASE="$(git rev-parse HEAD)"
diff <(git show "$BASE:skills/myflow-contracts/pipeline.md" | grep -v '^[[:space:]]*$' | sort) \
     <(grep -v '^[[:space:]]*$' skills/myflow-contracts/pipeline.md | sort)
scripts/check-references.sh
```

Expected: the `diff` shows as `>` only the new heading and its one authored sentence, and as `<`
only the pre-repoint form of any citation line. **No other line may appear as `<`** — this task moves
text *within* one file, so a moved line appears on both sides and produces no output at all. Then
mutate the new bold token in a scratch copy and confirm `check-references.sh` **fails**.

**Build:** green

### 1.2 Confirm `Temporary artifacts registry` stays, and record why

**Files:**
- Modify: `<WT>/skills/myflow-contracts/pipeline.md`

`## Temporary artifacts registry` is 5,301 bytes and is cited by
`skills/myflow-contracts/project-configuration.md` and
`skills/myflow-contracts/workspace-isolation.md` — both loaded by `/myflow-do`.

<!-- measured: python3 range extraction plus grep -rln over skills/ @ 63fa6ee -->

- [x] Leave the section in `pipeline.md`. Add one sentence to its opening paragraph stating that it
      stays in the core **because two contracts `/myflow-do` loads cite into it**, so a reader of
      those contracts never meets a citation to a file their command does not load. Newly authored.

- [x] Confirm no citation into it needs repointing:

```bash unverified:run in the apply worktree
cd "<WT>"
grep -rn '\*\*Temporary artifacts registry\*\*' skills/ rules/ CLAUDE.md AGENTS.md \
  | grep -v pipeline-rationale
```

Expected: every hit names `pipeline.md`, and none names `finish-contract.md`.

**Build:** green

### 1.3 Hoist the worktree scan the self-heal contract depends on

**Files:**
- Modify: `<WT>/skills/myflow-contracts/pipeline.md`
- Modify: `<WT>/skills/myflow-contracts/finish-contract.md`
- Modify: `<WT>/skills/myflow-contracts/state-self-heal.md`

**This task was added during implementation.** Task 2.1's implementer hit it and stopped rather than
freelancing, which was the right call. It is not new scope: the ADDED requirement
**A passage another command depends on is hoisted before a single-command move**
(`openspec/changes/kan-87-cut-per-command-load-further/specs/myflow-contract-economy/spec.md`)
already governs it — group 1 simply failed to enumerate this instance.

`skills/myflow-contracts/state-self-heal.md` reuses the `git worktree list --porcelain` scan that
lives inside `### Worktree cleanup` to rebuild a state file's `worktrees` map. That file is loaded
by `/myflow-status` directly and by `/myflow-do` and `/myflow-start` through `state-file.md` — and
none of them loads `finish-contract.md`. Repointing the citation there would create the exact
cross-command dangling reference this change exists to prevent; leaving it pointing at
`pipeline.md` was a real stale reference that `scripts/check-references.sh` caught.

- [x] Hoist the scan — its introductory sentence, the fenced `git worktree list --porcelain`
      snippet, **Never guess a path** and **The path is taken with `substr`, never `$2`** — out of
      `### Worktree cleanup` into a new `## Resolving a change's worktrees` section in
      `pipeline.md`, byte for byte.

- [x] Leave a citation in its place in `finish-contract.md`, and repoint
      `state-self-heal.md`'s citation to the new core section.

- [x] Verify `scripts/check-references.sh` exits 0 — it was failing on this exact line before the
      hoist, which is what makes the fix checkable rather than asserted.

**Build:** green

---

## 2. Move the finish-only half out

### 2.1 Move `Finish contract` and `Worktree cleanup` into `finish-contract.md`

**Files:**
- Create: `<WT>/skills/myflow-contracts/finish-contract.md`
- Modify: `<WT>/skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: the core left complete by tasks 1.1 and 1.2.
- Produces: `finish-contract.md`, carrying its rules **and** its reasoning inline, per
  **A section reachable from only one command lives in its own file**
  (`openspec/specs/myflow-contract-economy/spec.md`).

After the hoists, what remains finish-only is **25,760 bytes**: `## Finish contract` with both run
procedures and the base-branch resolution, and `### Worktree cleanup`.

<!-- measured: 33012 block minus 1951 minus 5301 @ 63fa6ee -->

- [x] Extract those sections by line range with `sed -n`, **never by retyping**, into
      `<WT>/skills/myflow-contracts/finish-contract.md`. Delete the same ranges from `pipeline.md`.

- [x] Give the new file a `#` title, a bolded canonical-authority sentence, and one sentence naming
      `/myflow-finish` as the only command that loads it. Leave every moved heading at its original
      level.

- [x] Leave a **pointer stub** in `pipeline.md` where `## Finish contract` was: the heading, one
      sentence naming what the contract governs, and the exact path.

- [x] **Check every `Finish contract` citation by reading, one at a time** — found by re-running the
      `grep`, not by a count fixed here. At the time of writing they were in
      `skills/myflow-finish/SKILL.md`, `skills/myflow-do/SKILL.md`, `skills/myflow-start/SKILL.md`
      and `skills/myflow-status/SKILL.md`. For each, decide whether it names substance that moved:

      - `myflow-do`'s was the outcome table — already repointed by task 1.1. Confirm nothing else in
        that file cites into the moved sections.
      - `myflow-start`'s and `myflow-status`'s are rationale citations. Repoint the path and delete
        any position word the move falsified.
      - `myflow-finish`'s repoint to the new file.

      **A green `check-references.sh` is not evidence here** — it resolves against the stub's
      heading. State, per citation, what you read and why it now reaches its substance.

- [x] Verify:

```bash unverified:run in the apply worktree after the move
cd "<WT>"
BASE="$(git rev-parse HEAD)"
diff <(git show "$BASE:skills/myflow-contracts/pipeline.md" | grep -v '^[[:space:]]*$' | sort) \
     <(cat skills/myflow-contracts/pipeline.md skills/myflow-contracts/finish-contract.md \
       | grep -v '^[[:space:]]*$' | sort)
scripts/check-references.sh && scripts/check-vocabulary.sh
wc -c skills/myflow-contracts/pipeline.md skills/myflow-contracts/finish-contract.md
```

Expected: as `>` only the new file's title and header, the stub's lines, and the post-edit form of
each repointed or position-word-edited line; as `<` only their pre-edit forms. No other `<` line.
Both guards exit 0. Record the `wc -c` output for task 5.1.

- [x] Confirm `/myflow-info` is still satisfiable from `pipeline.md` alone:

```bash unverified:run in the apply worktree after the move
cd "<WT>"
grep -c '^stateDiagram-v2' skills/myflow-contracts/pipeline.md
grep -c '^#### ' skills/myflow-contracts/pipeline.md
grep -c '^| `/myflow-finish` |' skills/myflow-contracts/pipeline.md
```

Expected: `1` for the diagram, at least `8` for the level-2 expansions, and `1` for the
`/myflow-finish` row of the level-1 stage table — so the core still describes both runs and every
stage.

**Build:** green

---

## 3. Split the three pipeline skills

### 3.1 Split `skills/myflow-do/SKILL.md`

**Files:**
- Create: `<WT>/skills/myflow-do/SKILL-rationale.md`
- Modify: `<WT>/skills/myflow-do/SKILL.md`

The largest per-command file at 40,529 bytes. Apply the same partition test: **core if removing the
passage would change what an agent does; rationale if removing it changes no agent behaviour.**

Worked examples from this file, so the boundary is not left to taste:

| Passage | Goes to | Why |
|---------|---------|-----|
| The panel slot table, the optional-slot selection table, the re-run modes table | core | an agent reads them to act |
| Every `**Never**` guardrail | core | prohibitions |
| The four required dispatch blocks, quoted verbatim | core | an agent pastes them |
| The handoff block | core | an agent renders it |
| "the previous shape asked a hand-rolled table parser to recover one fact from a grammar defined in prose, and it failed **open** six distinct ways" | rationale | history |
| "That is the point of the split" and similar closing justifications | rationale | removing them changes nothing |

- [x] Extract rationale passages by line range, whole paragraphs only, into
      `<WT>/skills/myflow-do/SKILL-rationale.md`.

- [x] Give the appendix a `#` title, a bolded sentence naming it the reasoning behind
      `skills/myflow-do/SKILL.md`, and a statement that **a `/myflow-*` run never loads it**. Mirror
      the core's `##`/`###` heading tree in order; a wholly normative section keeps its appendix
      heading with no body.

- [x] Add one line to the core's header naming the appendix and the no-load rule.

- [x] **Do not touch the YAML frontmatter.** The harness resolves the skill through it; an appendix
      beside `SKILL.md` is an ordinary file and is not referenced from frontmatter.

- [x] Verify:

```bash unverified:run in the apply worktree after the split
cd "<WT>"
BASE="$(git rev-parse HEAD)"
diff <(git show "$BASE:skills/myflow-do/SKILL.md" | grep -v '^[[:space:]]*$' | sort) \
     <(cat skills/myflow-do/SKILL.md skills/myflow-do/SKILL-rationale.md \
       | grep -v '^[[:space:]]*$' | sort)
headings() { awk '/^```/{f=!f; next} !f && /^#{2,4} /' "$1"; }
diff <(headings skills/myflow-do/SKILL.md) <(headings skills/myflow-do/SKILL-rationale.md)
head -12 skills/myflow-do/SKILL.md
scripts/check-references.sh
wc -c skills/myflow-do/SKILL.md skills/myflow-do/SKILL-rationale.md
```

Expected: the first `diff` shows as `>` only the appendix's title, header and mirrored headings plus
the core's added header line, and as `<` only pre-edit forms of edited lines. The second `diff` is
**empty**. The `head` shows the frontmatter unchanged. The guard exits 0. Record `wc -c`.

**Build:** green

### 3.2 Split `skills/myflow-start/SKILL.md`

**Files:**
- Create: `<WT>/skills/myflow-start/SKILL-rationale.md`
- Modify: `<WT>/skills/myflow-start/SKILL.md`

28,944 bytes. Its `## Convergence` section is unusually rationale-dense — the two prompts and the
threshold are core; the paragraphs explaining *why the two prompts recommend opposite courses* and
*why the threshold counts rounds rather than questions* are rationale.

Note that `pipeline.md` cites `**Convergence** (skills/myflow-start/SKILL.md)` for the threshold and
the prompts. Those stay in the core, so that citation still reaches its substance — confirm by
reading rather than by the guard.

- [x] Same procedure as task 3.1: extract whole rationale paragraphs by line range, author the
      appendix's title/header, mirror the heading tree, add the core header line, leave frontmatter
      untouched.

- [x] Verify with the same five checks as task 3.1, substituting this file pair, **plus**: confirm
      `pipeline.md`'s `**Convergence**` citation still names content in `SKILL.md` and not in the
      appendix.

**Build:** green

### 3.3 Split `skills/myflow-finish/SKILL.md`

**Files:**
- Create: `<WT>/skills/myflow-finish/SKILL-rationale.md`
- Modify: `<WT>/skills/myflow-finish/SKILL.md`

28,086 bytes. This file's citations were repointed by task 2.1; the split must not disturb them —
check they still sit on one physical line afterwards.

- [x] Same procedure as task 3.1.

- [x] Verify with the same five checks, **plus**: confirm every citation this file makes to
      `finish-contract.md` survived with its bold token and path on one line.

**Build:** green

---

## 4. Update the index, the rule layer and the entry points

### 4.1 Name the new files everywhere a file is discovered

**Files:**
- Modify: `<WT>/skills/myflow-contracts/SKILL.md`
- Modify: `<WT>/rules/myflow-manual-review.mdc`
- Modify: `<WT>/CLAUDE.md`
- Modify: `<WT>/AGENTS.md`

- [x] Add `finish-contract.md` to the index table in `skills/myflow-contracts/SKILL.md`, stating it
      is loaded by `/myflow-finish` and no other command.

- [x] Extend the **Rationale appendices** section of that file so it names the skill-level
      appendices as well, and states that a skill's appendix lives beside its `SKILL.md` rather than
      in the contracts directory.

- [x] Add `finish-contract.md` to the narrower-contracts table in `rules/myflow-manual-review.mdc`.
      **Do not add a count** — that file already states why it carries none.

- [x] Update the `skills/myflow-contracts/` row in `CLAUDE.md` and `AGENTS.md` identically.

- [x] Verify the index and the directory agree, and no skill loads an appendix:

```bash unverified:run in the apply worktree after the index edits
cd "<WT>"
diff <(ls skills/myflow-contracts/*.md | xargs -n1 basename | grep -v '^SKILL.md$' | sort) \
     <(grep -oE '\]\([a-z0-9-]+\.md\)' skills/myflow-contracts/SKILL.md \
        | sed -E 's/^\]\(//; s/\)$//' | grep -v '^SKILL.md$' | sort -u)
grep -rn 'rationale\.md' skills/myflow-*/SKILL.md | grep -iv 'never loads\|does not load\|reasoning behind'
scripts/check-references.sh && scripts/check-vocabulary.sh
```

Expected: the first `diff` is **empty**. The `grep` finds nothing — every mention of an appendix in
a skill file is the statement that it is *not* loaded, never a load instruction. Both guards exit 0.

- [x] Confirm only `/myflow-finish` is told to load the finish contract — the delta spec's
      **A start or do run does not load the finish contract** scenario:

```bash unverified:run in the apply worktree after the index edits
cd "<WT>"
grep -ln 'finish-contract\.md' skills/myflow-*/SKILL.md
```

Expected: five hits, and **every one must be read rather than counted**:

| Hit | Why it is legitimate |
|-----|----------------------|
| `skills/myflow-finish/SKILL.md` | the loader |
| `skills/myflow-contracts/SKILL.md` | the index row naming the loader — not itself a load instruction |
| `skills/myflow-status/SKILL.md` | a reasoning pointer for the merge-status test |
| `skills/myflow-start/SKILL.md` | a reasoning pointer — "whose reasoning is stated under **Finish contract** … and is not re-argued here" |
| `skills/myflow-info/SKILL.md` | the pointer line decision `myflow-info-core-only` calls for; that file states plainly that it does **not** read the contract |

**A hit in `myflow-do` is a defect** — that command never loads `finish-contract.md`, so a citation
from it is the failure task 1.1's hoist exists to prevent, reappearing by another route.

The last two are legitimate under the ADDED requirement's own test: *a citation whose only role is to
point at reasoning is not a real dependency*. **This expectation originally named only
`myflow-finish`**, and the box was ticked against a command that printed four lines — the same
stale-Expected defect KAN-82's panel found twice.

**It then recurred a third time, inside this very correction.** The table was written naming four
hits, and a later fix in the same round added the fifth — `myflow-info/SKILL.md`'s pointer line,
which decision `myflow-info-core-only` itself calls for. A hand-authored enumeration of `grep`
hits goes stale the moment any later task adds one, exactly as the budget table does.

**So this block carries the same instruction task 5.1 carries: re-run the command as the task's last
action and read what it prints, rather than asserting a count fixed at design time.** That is the
rule; the table above is its output at the moment of writing, not a substitute for running it.

**Build:** green

---

## 5. Widen the guard and regenerate budgets

### 5.1 Cover the skill files, then regenerate the table last

**Files:**
- Modify: `<WT>/scripts/check-contract-budget.sh`
- Modify: `<WT>/scripts/test-check-contract-budget.sh`

**This task runs last**, and its final step runs after every other task in the plan is complete.
KAN-82's finding F23 was that this table drifted three times because it was recomputed while later
edits were still landing.

- [x] Widen the scanned set from `$CONTRACT_DIR/*.md` to that plus `$REPO_ROOT/skills/*/SKILL.md`
      and `$REPO_ROOT/skills/*/SKILL-rationale.md`. Keep every existing property: the symlink
      refusal, the readability pre-test, the `budget_for` bash lookup, the count in the verdict, and
      the three exit codes.

- [x] Budget keys must now distinguish files that share a basename — every skill has a `SKILL.md`.
      Key the table on the path **relative to the repository root**, not on `basename`, and update
      the existing rows to match.

```bash unverified:authored for this change; run the harness to confirm
# name -> path-relative-to-repo-root, so skills/myflow-do/SKILL.md and
# skills/myflow-start/SKILL.md cannot collide on a bare basename.
rel="${path#"$REPO_ROOT"/}"
max="$(budget_for "$rel")"
```

- [x] Add harness cases for the new glob: a `skills/<x>/SKILL.md` over budget, one with no row, and
      two skills whose `SKILL.md` files have different budgets — the last proving the basename
      collision is actually fixed.

- [x] **Last action of the whole change**: regenerate the table from the tree.

```bash unverified:run once every other task in this plan is complete
cd "<WT>"
for f in skills/myflow-contracts/*.md skills/*/SKILL.md skills/*/SKILL-rationale.md; do
  [ -e "$f" ] || continue
  printf '%s %d\n' "$f" "$(( $(wc -c < "$f") * 125 / 100 ))"
done | sort
```

Expected: one row per covered file, each the achieved size plus 25%. Paste them into `budgets()`,
then confirm every row matches `floor(actual * 1.25)` exactly before finishing.

**Build:** green

---

## 6. Whole-repository verification

### 6.1 Run every declared guard and harness, and measure the result

**Files:**
- Modify: none — this task changes nothing and only proves the change.

- [x] Run every command in `.myflow/project.md`'s `## lint` and `## test` sections.

- [x] Prove the new files ship to all three harnesses with no installer edit:

```bash unverified:run once every other task is complete
cd "<WT>"
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global >/dev/null
for h in .claude .cursor .codex; do
  test -r "$SANDBOX/$h/skills/myflow-contracts/finish-contract.md" || echo "MISSING $h/finish-contract.md"
  for s in myflow-do myflow-start myflow-finish; do
    test -r "$SANDBOX/$h/skills/$s/SKILL-rationale.md" || echo "MISSING $h/$s/SKILL-rationale.md"
  done
done
git diff --stat -- setup.sh
rm -rf "$SANDBOX"
```

Expected: no `MISSING` line, and `git diff --stat -- setup.sh` prints nothing.

- [x] Record the saving:

```bash unverified:run once every other task is complete
cd "<WT>"
printf 'pipeline.md   %6d -> %6d\n' 88253 "$(wc -c < skills/myflow-contracts/pipeline.md)"
for s in myflow-do myflow-start myflow-finish; do
  printf '%-14s %6s -> %6d\n' "$s" "$(git show HEAD:skills/$s/SKILL.md | wc -c)" "$(wc -c < skills/$s/SKILL.md)"
done
```

Expected: every figure smaller than its baseline. Report the actual numbers in the handoff rather
than a target.

- [x] Confirm no rule changed. For each of the four split pairs, every `<` line in the line-multiset
      diff is a repointed citation or a deleted position word — reviewed one at a time, not skimmed.
      **A `<` line that is neither is lost text.**

**Build:** green
