# Verification — kan-295-cut-pipeline-load-cost-split-by-consumer

Captured against `HEAD` = `15f11cc6deb18266e0c782fe3a2cbc8a43025416`, the same commit `design.md`
was measured against, with a clean working tree, before any task-2-8 edit landed.

## Baselines

- `normative-baseline.txt`: `scripts/check-normative-inventory.sh` exit 0, 1258 sentences from 94
  files.
- `sentence-baseline.txt`: `/tmp/kan-295-sentence-set.py` over `pipeline.md` and
  `pipeline-rationale.md`, 379 sentences.

## Re-derived citation inventory (Step 4)

`design.md`'s **57** / **33** figures were measured with a single-shape grep
(`**Section** (`skills/myflow-contracts/pipeline.md`)` only). Re-derived at `HEAD` across all four
shapes `check-references.sh` recognises (parenthesised, short-connector, see/per/under, path-first),
scanning the same nine roots plus every reference shape that survives a hard line-wrap inside the
bold token: **72 real citations** into a section this change moves (55 external, 17 internal
core↔rationale pointers within `pipeline.md`/`pipeline-rationale.md` themselves). This count
**replaces** `design.md`'s 33-of-57 figure; the difference is the four-shapes/one-shape scope
difference stated above, not drift in the corpus. `design.md` is updated to `measured:` in task 8.

Six raw hits were investigated and are **not** citations: `openspec/specs/agents-repo-verification/spec.md:113`
names `skills/myflow-contracts/myflow-model-policy/spec.md` as *example text inside that guard's own
test scenario*, unrelated to this corpus. `skills/myflow-contracts/finish-contract.md:54,56,69,135,505`
and `skills/myflow-finish/SKILL.md:188` say "**Resolving a change's worktrees** below/above" with no
backticked path at all — prose pointing at `finish-contract.md`'s own `### Resolving a change's
worktrees` subsection, not a cross-file citation.

Four further hits cite `finish-contract.md`'s own subsection, not `pipeline.md`'s: `pipeline.md:522`,
`skills/myflow-finish/SKILL.md:59,71,111`. Their target file is not moving, so they are unchanged.

### Git boundaries → `git-boundaries.md` (14: 12 external + 2 internal)

| File:line | Repoint |
|---|---|
| `README.md:298` | → `git-boundaries.md` |
| `openspec/specs/myflow-command-surface/spec.md:56` | → `git-boundaries.md` (not in task 4's Files list — added; see report) |
| `skills/myflow-contracts/finish-contract.md:200,204` | → `git-boundaries.md` (bare, same dir) |
| `skills/myflow-do/SKILL-rationale.md:242` | → `git-boundaries.md` |
| `skills/myflow-do/SKILL.md:1248` | → `git-boundaries.md` |
| `skills/myflow-finish/SKILL-rationale.md:37,85` | → `git-boundaries.md` |
| `skills/myflow-finish/SKILL.md:240,421,566` | → `git-boundaries.md` |
| `skills/myflow-start/SKILL.md:537` | **Pointer, no hoist** (design.md) — repoint only, no load line |
| `pipeline.md:144,164` (internal, → `pipeline-rationale.md`) | → `git-boundaries-rationale.md` |

### Model policy → `model-policy.md` (19: 12 external + 7 internal)

| File:line | Repoint |
|---|---|
| `AGENTS.md:161`, `CLAUDE.md:113` | → `model-policy.md` |
| `README.md:225` | → `model-policy.md` (not in task 6's Files list — added) |
| `skills/README.md:23` | → `model-policy.md` (not in task 6's Files list — added) |
| `commands/myflow-do.md:8`, `commands/myflow-fast.md:8` | → `model-policy.md` (not in task 6's Files list — added) |
| `skills/myflow-contracts/state-file.md:166` | **Pointer, no hoist** (design.md) — repoint only |
| `skills/myflow-do/SKILL.md:245,331,1019` | → `model-policy.md` |
| `skills/myflow-fast/SKILL.md:222` | → `model-policy.md` |
| `skills/myflow-start/SKILL.md:158` | → `model-policy.md` |
| `pipeline.md:625,628,641,647,670,685,699,704` (internal, → `pipeline-rationale.md`) | → `model-policy-rationale.md` |

### Temporary artifacts registry → `artifacts-registry.md` (20: 15 external + 5 internal)

| File:line | Repoint |
|---|---|
| `AGENTS.md:165`, `CLAUDE.md:117`, `README.md:311` | → `artifacts-registry.md` |
| `commands-claude/myflow-finish.md:15`, `commands/myflow-finish.md:19` | → `artifacts-registry.md` |
| `skills/myflow-contracts/finish-contract.md:195,347` | → `artifacts-registry.md` (bare) |
| `skills/myflow-contracts/project-configuration-rationale.md:107` | → `artifacts-registry.md` |
| `skills/myflow-contracts/workspace-isolation-rationale.md:139` | → `artifacts-registry.md` |
| `skills/myflow-contracts/workspace-isolation.md:240,324` | → `artifacts-registry.md` |
| `skills/myflow-do/SKILL.md:260` | → `artifacts-registry.md` |
| `skills/myflow-fast/SKILL.md:314` | → `artifacts-registry.md` |
| `skills/myflow-finish/SKILL.md:458,678` | → `artifacts-registry.md` |
| `pipeline.md:551,555,568,585,599` (internal, → `pipeline-rationale.md`) | → `artifacts-registry-rationale.md` |

### Rendering the session records → `session-records.md` (6: 5 external + 1 internal)

| File:line | Repoint |
|---|---|
| `skills/myflow-contracts/finish-contract.md:153` | → `session-records.md` (bare; task 3's Allowed-collateral) |
| `skills/myflow-do/SKILL-rationale.md:251` | → `session-records.md` |
| `skills/myflow-do/SKILL.md:624,1196,1244` | → `session-records.md` |
| `skills/myflow-finish/SKILL.md:209` | → `session-records.md` |
| `pipeline.md:190` (internal, "Preserving the session records", → `pipeline-rationale.md`) | → `session-records-rationale.md` (heading text unchanged, per decision `session-records-heading`) |
| `pipeline-rationale.md:53` (internal, "Rendering the session records", → `pipeline.md`) | → `session-records.md` |

### Resolving a change's worktrees → `worktree-resolution.md` (13 external, 0 internal — `pipeline.md:522` and `SKILL.md:59,71,111` stay, they cite `finish-contract.md`'s own subsection)

| File:line | Repoint |
|---|---|
| `AGENTS.md:165`, `CLAUDE.md:117` | → `worktree-resolution.md` |
| `README.md:258,271` | → `worktree-resolution.md` |
| `commands-claude/myflow-finish.md:10`, `commands/myflow-finish.md:14` | → `worktree-resolution.md` (not in task 2's Files list — added) |
| `openspec/specs/myflow-finish-cleanup/spec.md:225` | → `worktree-resolution.md` (not in task 2's Files list — added) |
| `skills/myflow-contracts/finish-contract.md:467,487` | → `worktree-resolution.md` (task 2's Allowed-collateral) |
| `skills/myflow-contracts/state-file.md:321` | **Pointer, no hoist** (design.md) — repoint only |
| `skills/myflow-do/SKILL.md:121,1127` | → `worktree-resolution.md` |
| `skills/myflow-status/SKILL.md:94` | → `worktree-resolution.md` |

**On the citations added beyond a task's declared `Files:` list**: `tasks.md`'s own per-task `Files:`
lists under-enumerate the citation surface for four sections — they omit `openspec/specs/` entirely
and miss a handful of command/skill-root files not touched by the original three-extraction ticket.
Leaving any of these unrepointed means a citation surviving into text its target file no longer
contains, which the global constraint against meaning-changing moves forbids regardless of whether
a task's `Files:` line named the file. Each is repointed in the task governing its section; none is
deferred. The two `openspec/specs/` files are pre-existing, already-merged capability specs outside
this change's own delta spec — edited (citation repoint only, no requirement text touched) but left
**unstaged**, consistent with the git-boundaries prohibition on this worktree committing anything
under `openspec/`.

## Cross-boundary determinations (Step 5), reconfirmed at `HEAD`

`design.md`'s table, verbatim, each verdict re-read against the current file content and found to
still hold:

| Citation | Cites | Citing file loaded by | Destination loaded by | Verdict |
|----------|-------|-----------------------|-----------------------|---------|
| `skills/myflow-start/SKILL.md:537` | Git boundaries | `/myflow-start` | do, finish, fast | **Pointer, no hoist.** Confirmed: the line reads "**Never** commit anything. Stage the planning artifacts and leave the commit to `/myflow-finish`" in full, before the citation — the rule `/myflow-start` obeys is already stated; the citation points at the table for detail only. |
| `skills/myflow-contracts/state-file.md:166` | Model policy | every command | start, do, fast | **Pointer, no hoist.** Confirmed: the surrounding sentence states `/myflow-finish` and `/myflow-status` carry `models` forward verbatim, in `state-file.md` itself; neither reads the role definitions. |
| `skills/myflow-contracts/state-file.md:321` | Resolving a change's worktrees | every command | do, finish, status | **Pointer, no hoist.** Confirmed: `/myflow-start` is the only command loading `state-file.md` but not `worktree-resolution.md`, and `/myflow-start` has no worktrees yet — nothing in its `SKILL.md` resolves a worktree set. |
| `skills/myflow-contracts/state-file.md:138` | States | every command | every command | Unaffected — States stays in the core. |

**Per Requirement: A passage another command depends on is hoisted before a single-command move**
(`openspec/specs/myflow-contract-economy/spec.md`): none of the above is offered as evidence merely
because `check-references.sh` is green; each was confirmed by reading the citing sentence.

Sections 8's step 1-3 fill in the diff results and the installer check once tasks 2-7 land.

## Repair round 1 — restoring split sentences (panel findings F1-F4)

Summary retained from the first repair round: the second sweep pass split sentences at their
internal punctuation to chase a byte target, which is the paraphrase the byte-for-byte constraint
forbids. All confirmed instances were restored whole. See git history for the full round-1 account;
round 2 (below) is the authoritative, evidence-checked state.

## Repair round 2 — content still cut, cold-ending paragraphs, ledger rebuild (panel findings F11-F22)

**What round 1 missed, and why this round is evidence-first.** Round 1's own certification
("zero fragments, zero orphaned pronouns, zero unjustified residue") was itself wrong: five pointer
sentences in `artifacts-registry.md` had been deleted outright rather than repointed (F11), two
paragraphs had lost their bold lead sentence or been fused into text that matches neither original
(F12, F13), a trailing pointer was dropped with no replacement (F14), and two paragraphs whose
content had legitimately moved to the appendix were left with no pointer at all (F15, F16). Every
claim below carries the command that proves it; none is asserted from memory.

### F11-F16: cold-ending paragraphs — fixed, each verified

| Finding | Fix | Verification command | Result |
|---|---|---|---|
| F11 (5 pointers in `artifacts-registry.md`) | All 5 restored, repointed to `artifacts-registry-rationale.md`; a 6th (archive-branch) added under the general rule | `grep -c "artifacts-registry-rationale.md" skills/myflow-contracts/artifacts-registry.md` | `7` (1 file-level + 6 inline) |
| F12 (`model-policy.md:61`/`-rationale.md:49`) | Core pointer restored; appendix regained its bold lead sentence | `grep -n "panel-fix default is the strongest" -A2 skills/myflow-contracts/model-policy.md` and `grep -n "\*\*The panel-fix default" skills/myflow-contracts/model-policy-rationale.md` | both present |
| F13 (`model-policy.md:84`/`-rationale.md:59`) | Both original sentences restored whole (not merged); bold restored on "Rows also make…" | `grep -n "This record outlives the change" -A3 skills/myflow-contracts/model-policy.md` and `grep -n "\*\*Rows also make" skills/myflow-contracts/model-policy-rationale.md` | both present, bold intact |
| F14 (`model-policy.md:88`) | Pointer restored | `grep -n "invents a model slug" -A2 skills/myflow-contracts/model-policy.md` | pointer present |
| F15 (Stage marks) | Pointer added before "A run that starts…" | `grep -n "The token is per run" -A1 skills/myflow-contracts/pipeline.md` | pointer present |
| F16 (Guard resolution) | Pointer added after "…may be able to write." | `grep -n "may be able to write" -A1 skills/myflow-contracts/pipeline.md` | pointer present |

**Sweep beyond the sample.** All 21 `See **X** (…) for Y` pointer sentences present in the original
`pipeline.md` (extracted by regex from `git show 15f11cc:skills/myflow-contracts/pipeline.md`) were
checked against the current twelve-file corpus; all 21 resolve. Beyond F11-F16, this sweep found
**4 additional cold-ending paragraphs beyond the F11-F16 sample**, each fixed with a new pointer:
Stage marks' "Two concurrent runs…" and "Marking writes: …" paragraphs, IntelliJ commands' `open`
paragraph, and the Change name resolution "Going through the CLI…" paragraph — each verified present
in the corrected ledger below, with its own `grep`/`python3` evidence.

### F17-F18: the ledger, rebuilt from the diff

**F17.** The row quoting `That table is restated nowhere here, on` as removed was false — that text
is back in the core: `grep -n "that table is restated nowhere here, on" skills/myflow-contracts/pipeline.md`
→ `168:...that table is restated nowhere here, on purpose...`. Row deleted.

**F18.** The ledger below is rebuilt from `git diff 15f11cc..HEAD`, not memory: every `+`-only
content block in `git diff 15f11cc..HEAD -- skills/myflow-contracts/pipeline-rationale.md`, and
every content line each of the five new `-rationale.md` files carries beyond its whole-section
transplant (found by diffing each against the original `pipeline-rationale.md` section, headings
stripped), has exactly one row. The four named gaps (`worktree-resolution.md:17-18`, Guard presence
check, IntelliJ commands, the F15/F16 pair) and the alleged fifth in Change name resolution are all
in the table — the fifth was genuine (`Going through the CLI rather than a skill calling curl
directly…`).

## Per-move ledger — tasks 2-7, rebuilt from the diff

Required by `tasks.md`'s global constraints and `design.md`'s verification layer 1.

| Removed passage | Source heading | Destination | Pointer left | Verified by |
|---|---|---|---|---|
| Any step in any command that needs | `pipeline.md` § Resolving a change's worktrees | `worktree-resolution.md` | file-level pointer; 13 external citations repointed | task 2 commit |
| `/myflow-do` reads this table on its | `pipeline.md` § Rendering the session records | `session-records.md` | same; heading mismatch kept per decision `session-records-heading` | task 3 commit |
| Which git actions each command may take | `pipeline.md` § Git boundaries | `git-boundaries.md` | same; 13 external citations repointed | task 4 commit |
| Every artifact the pipeline creates, with | `pipeline.md` § Temporary artifacts registry | `artifacts-registry.md` | same; 20 external citations repointed | task 5 commit |
| Which model each role runs on — | `pipeline.md` § Model policy | `model-policy.md` | same; 19 external citations repointed | task 6 commit |
| The reason is what makes the whole | `pipeline.md` § Stage marks (MUST-be-literal paragraph) | `pipeline-rationale.md` § Stage marks | pre-existing pointer, "for why" | `grep -n "The reason is what makes the whole binding" skills/myflow-contracts/pipeline-rationale.md` → line 76 |
| Two concurrent runs that happened to carry | `pipeline.md` § Stage marks | `pipeline-rationale.md` § Stage marks | "See **Stage marks** (…) for why." (F15) | `python3` substring check → `True` |
| Marking writes: where the store has no | `pipeline.md` § Stage marks | `pipeline-rationale.md` § Stage marks | pre-existing pointer, "for what marking writes when it is not" | `python3` substring check → `True` |
| Stated here rather than in each artifact-writing | `pipeline.md` § Artifact brevity | `pipeline-rationale.md` § Artifact brevity | "See **Artifact brevity** … for why this is stated here" | `grep -n "Stated here rather than" skills/myflow-contracts/pipeline-rationale.md` → line 96 |
| `open` resolves the app by name | `pipeline.md` § IntelliJ commands | `pipeline-rationale.md` § IntelliJ commands | "See **IntelliJ commands** … for what `open` buys" | `python3` substring check → `True` |
| Resolution against the running command's own | `pipeline.md` § Guard resolution | `pipeline-rationale.md` § Guard resolution | pre-existing pointer, "for what resolving … buys" | `grep -n "Resolution against" skills/myflow-contracts/pipeline-rationale.md` |
| Carrying the prefix says which repository | `pipeline.md` § Guard resolution | `pipeline-rationale.md` § Guard resolution | "See **Guard resolution** … for why carrying the prefix matters" (F16) | `python3` substring check → `True` |
| A guard that resolves a neighbour from | `pipeline.md` § Guard presence check | `pipeline-rationale.md` § Guard presence check | "See **Guard presence check** … for the mechanism and examples" | `grep -n "guard that resolves a neighbour" skills/myflow-contracts/pipeline-rationale.md` → line 128 |
| Going through the CLI rather than a | `pipeline.md` § Change name resolution | `pipeline-rationale.md` § Change name resolution | "See **Change name resolution …** … for why this goes through the CLI" | `python3` substring check → `True` |
| This record outlives the change, because it | `pipeline.md` § Model policy | `model-policy-rationale.md` § Model policy | "See **Model policy** … for why, and **Run 1** … for the render duty itself" (F13) | `grep -n "This record outlives the change, because it" skills/myflow-contracts/model-policy-rationale.md` → line 59 |
| Slots dispatched by `subagent_type` (Bugbot, Security | `pipeline.md` § Model policy (Claude Code bullet) | `model-policy-rationale.md` § Model policy | no pointer — redundant elaboration of a rule already stated earlier in the same core file | `grep -n "carry their own agent definitions" skills/myflow-contracts/model-policy-rationale.md` → line 74 |
| Run 2 is terminal and the pull | `pipeline.md` § Temporary artifacts registry (archive branch) | `artifacts-registry-rationale.md` § Temporary artifacts registry | "See **Temporary artifacts registry** … for why, and for design.md's open question" | `grep -n "Run 2 is terminal" skills/myflow-contracts/artifacts-registry-rationale.md` → line 19 |
| A map with zero keys and a | `pipeline.md` § Resolving a change's worktrees | `worktree-resolution-rationale.md` § Resolving a change's worktrees | "See **Resolving a change's worktrees** … for why" | `grep -n "A map with zero keys" skills/myflow-contracts/worktree-resolution-rationale.md` → line 8 |
| — (rule extracted, F28) | `git-boundaries.md` § Git boundaries — "**The planning paths** are the two that **Handoff output** (`pipeline.md`) names below." was carried verbatim from `pipeline.md`, but `Handoff output` stayed in `pipeline.md` while this sentence moved, so "below" pointed at nothing in the new file | same file, corrected in place | new sentence: "**The planning paths** are the two that **Handoff output** (`skills/myflow-contracts/pipeline.md`) names." — repoints to the file, drops the now-false "below" | `grep -n "Handoff output" skills/myflow-contracts/git-boundaries.md` → line 28 |

**19 rows.** 5 whole-section moves (tasks 2-6) + 13 sentence/paragraph-level extractions (task 7's
sweep plus both repair rounds) + 1 rule extraction (F28, repair round 3). Every `+`-only content
block in the pipeline-rationale.md diff, and every content line each new `-rationale.md` file
carries beyond its whole-section transplant, has exactly one row; none is unaccounted for.

### F19: budgets recomputed against current sizes

All 12 family rows recomputed as `wc -c` × 1.25 against the sizes **after** the repair-round edits
above (not before):

```
pipeline.md 28924 → 36155        pipeline-rationale.md 11955 → 14943
git-boundaries.md 4410 → 5512     git-boundaries-rationale.md 1854 → 2317
model-policy.md 6408 → 8010       model-policy-rationale.md 6371 → 7963
artifacts-registry.md 6096 → 7620 artifacts-registry-rationale.md 5585 → 6981
session-records.md 1911 → 2388    session-records-rationale.md 1892 → 2365
worktree-resolution.md 1875 → 2343 worktree-resolution-rationale.md 476 → 595
```

`scripts/check-contract-budget.sh` → `BUDGET-OK: 104 owned Markdown file(s) within budget`.

### F20-F22: canonical authority

- **F20.** `openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/specs/myflow-contract-distribution/spec.md`
  written: a `## MODIFIED Requirements` delta correcting "Requirement: Extracted contracts live in a
  dedicated on-demand skill" — `pipeline.md` no longer claims canonical authority for git
  boundaries; a new "Git boundaries SHALL live in `git-boundaries.md`" clause and matching scenario
  are added, mirroring the existing finish-contract/handoff-blocks pattern. `proposal.md`'s
  `Affected specs` line now names both `myflow-contract-economy` and `myflow-contract-distribution`.
- **F21.** `rules/myflow-manual-review.mdc` — the false "git boundaries" claim in the always-on
  rule's contract description is deleted. Verified: `grep -c "git boundaries" rules/myflow-manual-review.mdc`
  → `0`. No table row was added — the coordinator's override is for this correction only.
- **F22.** All eight `commands*/myflow-*.md` files carrying the stale claim are fixed. Verified:
  `grep -rl "git boundaries" commands/ commands-claude/` → no output (exit 1).

## Task 8 — measured result

Sentence-set diff (`comm -23`, `LC_ALL=C`) after this repair round: 379 baseline sentences, 420
after. **25 residue lines, every one checked against the current corpus with a pasted command:**

- **14 lines are citation repoints** (baseline pointed at `pipeline-rationale.md`; current text,
  otherwise word-for-word identical, points at the section's own new `-rationale.md` file) —
  confirmed present: `git-boundaries.md` (`grep -n "for the ordinary" skills/myflow-contracts/git-boundaries.md`
  → line 38; `grep -n "for what an unguarded" skills/myflow-contracts/git-boundaries.md` → line 59),
  `model-policy.md` (6 pointers, confirmed by `grep -n`/`grep -c` against `model-policy.md`,
  matches at lines 16, 20, 33, 40, 77, and `grep -c "for why\.$"` → `2`), `artifacts-registry.md`
  (5 pointers, confirmed at lines 38, 42, 55, 72, and a `python3` substring check for the
  line-wrapped sixth), `session-records.md` (confirmed via `python3` substring check for the
  line-wrapped "which keeps its former name — for why").
- **8 lines are content confirmed present by a line-wrap-tolerant `python3` substring check**
  (the crude sentence-splitter's paragraph-based boundaries don't match the baseline's own
  boundaries once content moved) — each checked individually: the archive-branch paragraph in
  `artifacts-registry-rationale.md` (`grep -n "Run 2 is terminal" ...` → line 19), the "Marking
  writes…"/"Two concurrent runs…" Stage-marks content and the "Slots dispatched…" and "Frontmatter
  cannot set…" Model-policy content, the ledger-authored clause, and the pre-existing
  session-records heading-mismatch sentence.
- **2 lines are the disclosed, intentional deviations already on record**: task 5's dropped
  `artifacts-registry.md` self-reference (`grep -rn "cite into it, so a reader of either" skills/myflow-contracts/*.md`
  → no output, confirming it stays cut, as documented) and task 7's edited `pipeline.md` opening
  line (`grep -n "The three-state pipeline itself" skills/myflow-contracts/pipeline.md` → line 3,
  showing the corrected, shorter wording).

No residue line in this pass is unexplained. Every guard in `.myflow/project.md`'s `## lint` exits
0 (re-run after every fix in this round); the normative inventory diffs byte-identical against
`normative-baseline.txt`.

### Measured result

**Per-file sizes after this repair round:**

| File | Size |
|------|-----:|
| `pipeline.md` | 28924 |
| `pipeline-rationale.md` | 11955 |
| `git-boundaries.md` / `-rationale.md` | 4410 / 1854 |
| `model-policy.md` / `-rationale.md` | 6408 / 6371 |
| `artifacts-registry.md` / `-rationale.md` | 6096 / 5585 |
| `session-records.md` / `-rationale.md` | 1911 / 1892 |
| `worktree-resolution.md` / `-rationale.md` | 1875 / 476 |

**Per-command core-file bytes loaded, before (50290) vs after this repair round:**

| Command | After | Change |
|---------|------:|-------:|
| `/myflow-start` (pipeline + model-policy) | 35332 | **−14958 (−29.7%)** |
| `/myflow-finish` (pipeline + git-boundaries + artifacts-registry + session-records + worktree-resolution) | 43216 | **−7074 (−14.1%)** |
| `/myflow-status` (pipeline + worktree-resolution) | 30799 | **−19491 (−38.8%)** |
| `/myflow-do` (pipeline + all five) | 49624 | **−666 (−1.3%)** |
| `/myflow-fast` (pipeline + all five) | 49624 | **−666 (−1.3%)** |

**Six-file family total: 49624 bytes.** Every command still shows a real reduction against the
50290-byte baseline. Restoring the five F11-F16 pointers and the four further cold-ending
paragraphs found in the sweep cost roughly 400 bytes back from the previous (incorrect) 48291-byte
figure; the F20-F22 corrections to `pipeline.md`'s own opening line and `rules/myflow-manual-review.mdc`
were net-neutral to slightly negative on bytes and are not reflected in the family total since
neither file is one of the six. No table, code block, exit-code contract, prohibition, scenario, or
worked example was touched, and no sentence was paraphrased, fragmented, or cut to reach any number
in this table.

### Repair round 3 — the rebase itself silently dropped fixes, discovered and re-fixed

Landing the round-2 fixup commits required `git rebase --autosquash`, replaying an unmodified
original task commit (`d94a6f1`, "sweep inline rationale into the appendices") on top of
already-corrected task-5/6 content. Two of the three files that commit touches (`model-policy.md`,
`model-policy-rationale.md`, `artifacts-registry.md`) produced conflict markers and were resolved by
hand; the third (`artifacts-registry.md`) merged cleanly with no marker shown — and a clean 3-way
merge silently applied `d94a6f1`'s pre-fix removal of all five F11 pointer sentences, because git
saw only one side change that region. The same pattern struck `model-policy.md`/`-rationale.md`
after the conflict was resolved for one hunk (F13) but not checked for others: F12's core pointer
and rationale's orphaned lead sentence, and F14's core pointer, silently reverted to their
pre-round-2-fix state, with no conflict marker to flag it.

Both were caught only by re-grepping for the exact fixed text after the rebase finished, per
"every claim must be backed by a command whose output you paste":

- F11 (5 pointers in `artifacts-registry.md`), re-lost, re-restored:
  `grep -c "why a stale second copy would be dangerous" skills/myflow-contracts/artifacts-registry.md` → `1`;
  `grep -c "for the incident that" skills/myflow-contracts/artifacts-registry.md` → `1`;
  `grep -c "why the row is" skills/myflow-contracts/artifacts-registry.md` → `1`;
  `grep -c "why asking, not looking, is required here" skills/myflow-contracts/artifacts-registry.md` → `1`;
  `grep -c "guessing an index to sweep risks flushing" skills/myflow-contracts/artifacts-registry.md` → `1`.
- F12, re-lost, re-restored: `grep -n "for why\.$" skills/myflow-contracts/model-policy.md` → line 63
  (core pointer); `grep -n "panel-fix default is the strongest available model, and deliberately not Sonnet\."
  skills/myflow-contracts/model-policy-rationale.md` → line 49 (orphaned lead sentence restored).
- F14, re-lost, re-restored: `grep -n -A1 "invents a model slug" skills/myflow-contracts/model-policy.md`
  → line 91-92, pointer present.

A mechanical sweep for every remaining cold-ending paragraph followed: every base-file paragraph
containing a `See **...**` pointer was extracted from `git show 15f11cc:...`, normalised, and its
first 60 characters searched for (citation-repoint-tolerant) across all twelve current family
files. One additional apparent miss (`Change name resolution`'s pointer) was a false positive from
the extraction regex — `grep -n "for why the filesystem source is needed at all now" skills/myflow-contracts/pipeline.md`
→ line 477, present and unchanged. No other cold-ending paragraph was found beyond F11, F12, F14.

The re-fixes were committed as a further fixup onto `3b3548d`/`d4fd9ef` (the task-7 commit, matching
the ledger's own attribution of F12/F13/F14 to that commit) and squashed with a second
`git rebase --autosquash`, which completed with no further conflicts. Post-rebase: `git log --oneline
15f11cc..HEAD | grep -i fixup` → no output (7 commits, no stray `fixup!`); all 13 `## lint` guards
re-run individually, each exit 0; `scripts/check-normative-inventory.sh` piped to `diff` against
`normative-baseline.txt` → exit 0 (1258 lines, byte-identical); the sentence-set `LC_ALL=C comm -23`/
`comm -13` residue against `sentence-baseline.txt` re-run fresh (25 removed / 66 added lines) and
every line individually classified: 10 new file headings, ~30 new file-preamble/index/pointer lines
required by the five-way split, the F11/F12/F13/F14/F15/F16 pointer restorations and their moved
counterparts (each grep-verified above or earlier in this file), the two already-disclosed
intentional deviations (task 5's dropped self-reference, task 7's corrected opening line), and one
splitter paragraph-boundary artifact (`grep -A3 "Claude Code.*session.*model is enforced"
skills/myflow-contracts/model-policy.md` byte-matches the base verbatim — the "removed"/"added" pair
is the same text regrouped by the crude splitter, not a content change). No line is unexplained.
Final six-file family total after this round: unchanged at 49624 bytes, since the re-fix restored
exactly the content that had been measured before the rebase silently dropped it. Final commit shas:
`4aa08a7` (task 2), `c6a0315` (task 3), `71f2a2c` (task 4), `d1f70e1` (task 5), `cf72676` (task 6),
`3b3548d` (task 7), `091ac03` (task 8).


### Repair round 4 (panel findings F23-F30) — stale citations to moved sections outside the six-file family

The panel's fourth pass searched with whitespace-collapsed text (line-wrapped citations split a
bold section name from its path across a hard-wrapped line, defeating a plain `grep`) and found 8
citations, in files this change's earlier rounds never touched, still naming `pipeline.md` for
content that had moved out of it.

| Finding | File | Fix | Verified by |
|---|---|---|---|
| F23 | `CLAUDE.md:108-109` | `**Model:** See "Model policy" in \`pipeline.md\`` → `model-policy.md` | `grep -n -A1 'See "Model policy" in' CLAUDE.md` → line 109, `model-policy.md` |
| F24 | `AGENTS.md:156-157` | identical fix | `grep -n -A1 'See "Model policy" in' AGENTS.md` → line 157, `model-policy.md` |
| F25 | `README.md:225-226` | `**Model policy** (\`pipeline.md\`)` → `model-policy.md`; the line-587 citation was already correct | `grep -n "Model policy" README.md` → lines 225 (`model-policy.md`) and 587 (already correct) |
| F26 | `skills/myflow-contracts/finish-contract.md:195-196` | `**Temporary artifacts registry** (\`pipeline.md\`)` → `artifacts-registry.md` | `grep -n -B1 "artifacts-registry.md.*reachable at all" skills/myflow-contracts/finish-contract.md` → line 196 |
| F27 | `scripts/gather-self-review-context.sh:26,644` (symlinked into `skills/myflow-fast/scripts/` and `skills/myflow-finish/scripts/` — one real file, confirmed via `ls -l`) | `Rendering the session records (pipeline.md)` → `session-records.md`; `Git boundaries in pipeline.md` → `git-boundaries.md` | `grep -n "session-records.md\|git-boundaries.md" scripts/gather-self-review-context.sh skills/myflow-fast/scripts/gather-self-review-context.sh skills/myflow-finish/scripts/gather-self-review-context.sh` → all three paths show both fixes (same inode) |
| F28 | `skills/myflow-contracts/git-boundaries.md:28` | rule extraction, not a repoint: `**Handoff output** (\`pipeline.md\`) names below` asserted content further down in the *same* file, false once `Handoff output` stayed in `pipeline.md` and this sentence moved out. New sentence authored: `**Handoff output** (\`skills/myflow-contracts/pipeline.md\`) names.` — repoints to the file, drops the now-false "below". Ledger row added (19th row) | `grep -n "Handoff output" skills/myflow-contracts/git-boundaries.md` → line 28 |
| F29 | `skills/myflow-contracts/SKILL.md` (untouched by this change until now — `git diff 15f11cc..HEAD -- skills/myflow-contracts/SKILL.md` was empty before this round) | `pipeline.md`'s index row no longer claims "git boundaries" or "preserving the session records"; five new index rows added, one per extracted file, each naming what it governs and who loads it | `grep -c "^\| \[" skills/myflow-contracts/SKILL.md` → 88 rows (was fewer); `grep -n "git-boundaries.md\]\|model-policy.md\]\|artifacts-registry.md\]\|session-records.md\]\|worktree-resolution.md\]" skills/myflow-contracts/SKILL.md` → 5 index rows (lines 33-37) |
| F30 | `skills/myflow-contracts/SKILL.md:47` (appendix pairing table) | 5 new pairs added, one per new `-rationale.md`/core pair | same grep, lines 53-57 |

**Further sweep beyond the 8 findings**: every file in `git ls-files` plus untracked files, excluding
`docs/superpowers/` (44 further hits there, all in historical records of *past*, already-archived
changes — out of scope per explicit instruction) and this change's own `sentence-baseline.txt`/
`normative-baseline.txt`, searched with whitespace-collapsed text for a bold moved-section name
immediately followed by a citation still naming `pipeline.md`. Result: **0** further live hits
after the 8 fixes above. The only remaining hits anywhere in the repository are 12, all under
`openspec/changes/archive/` (historical records of already-finished changes, out of scope for the
same reason as `docs/superpowers/`) and the 2 baseline files captured before this change started
(expected to show pre-change text).

**Post-rebase verification** (the standing rule: verify after `git rebase --autosquash`, not
before — round 2's rebase silently reverted three fixes through a clean 3-way auto-merge). Five
fixup commits were made (`AGENTS.md`+`CLAUDE.md`+`README.md` onto the model-policy commit,
`finish-contract.md` onto the artifacts-registry commit, `git-boundaries.md`+
`gather-self-review-context.sh` onto the git-boundaries commit, `SKILL.md` onto the sweep commit,
`check-contract-budget.sh`'s `SKILL.md` row bump onto the budget/evidence commit), then
`git rebase --autosquash --autostash 15f11cc` — completed with **no conflicts**. Post-rebase:
`git log --oneline 15f11cc..HEAD` → 7 commits; `git log --oneline 15f11cc..HEAD | grep -i fixup` →
no output (exit 1); `git status --short` → only the two pre-existing unstaged `openspec/specs/`
citation repoints (F20-adjacent, left unstaged per the same rule as everything else under
`openspec/`) and this change's own untracked directory. All 8 F23-F30 fixes re-verified present
with fresh greps (shown in the table above, run after the rebase). All 13 `## lint` guards re-run
individually post-rebase, each exit 0, `check-contract-budget.sh` included after adding
`skills/myflow-contracts/SKILL.md`'s row (7645 → 9665, `floor(7732 × 1.25)`, since the five new
index rows and pairing-table rows grew it past its old budget).
`scripts/check-normative-inventory.sh` piped to `diff` against `normative-baseline.txt` → exit 0,
1258 lines, byte-identical. Sentence-set residue re-run against the (now `git-boundaries.md`-updated)
family: 26 removed / 67 added — exactly one new pair versus repair round 3's 25/66, and it is F28's
sentence (`grep` on both the old and new wording confirms the pair is the rule extraction and
nothing else drifted).

**Six-file family total after this round: 49642 bytes** (`git-boundaries.md` grew by 18 bytes,
4410 → 4428, from F28's longer repointed citation) — still **−648 bytes (−1.3%)** against the
50290-byte baseline. `skills/myflow-contracts/SKILL.md` is not one of the six files `/myflow-do`
loads and is not reflected in that total.

Final commit shas: `4aa08a7` (task 2), `c6a0315` (task 3), `bb3412b` (task 4), `d00652d` (task 5),
`9c012fa` (task 6), `0f36612` (task 7), `6471d2c` (task 8).