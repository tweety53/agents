> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Make an unchecked claim in a plan visible at the proposal gate and refusable at the
implementation gate, enforced by a guard for the part a script can actually enforce.

**Architecture:** One new contract defining four tags, one guard script plus a sandboxed harness in
this repository's established guard style, two skill edits, and a config wiring change. No installer
change — `setup.sh` copies the contracts directory wholesale.

**Tech stack:** Bash (guards), Markdown (skills, contracts, rules). No build system, no package
manager.

**This plan is its own first test case.** Every fenced block below carries `verified:` or
`unverified:`; every number carries `measured:` or `predicted:`. Task 3 runs the finished guard over
this file, and it must pass.

## Global constraints

- **No commits.** `/myflow-do` stages only — `git add`, never `git commit`.
- **Lint:** `scripts/check-vocabulary.sh` and `scripts/check-references.sh`. Both are green on
  this branch.
  <!-- measured: ./scripts/check-vocabulary.sh && ./scripts/check-references.sh @ branch openspec/kan-14-plan-provenance -->
  **There is no auto-fix command in this repository** — every hit is fixed by editing the offending
  line, never by weakening a guard or adding a suppression marker.
- **Test:** `scripts/test-setup.sh` and `scripts/test-check-references.sh`.
- **`scripts/` currently holds 4 scripts**; this change adds 3 (the post-review reshape split the
  guard into `check-plan-provenance.sh` plus `check-plan-provenance.py`, alongside
  `test-check-plan-provenance.sh` — see design.md's "Post-review reshape").
  <!-- measured: ls scripts/ | wc -l @ merge-base d38372a (the count BEFORE this change; the
       three scripts it adds do not exist at that ref, which is the point being measured) -->
- **Match the house style.** Guard scripts carry a header comment stating the rule *and the failure
  that motivated it*, report `file:line`, and accept an env-var root override so a harness can point
  them at a fixture tree. Read `scripts/check-references.sh` before writing the new guard.
- **Never add a suppression marker** to make a guard pass.

## 1. The contract and its discoverability

Defines the four tags. Nothing enforces them yet — that is task 2 — but the guard's rules must be
written down before code encodes them.

**Files:**
- Create: `skills/myflow-contracts/plan-provenance.md`
- Modify: `rules/myflow-manual-review.mdc` — the contracts table (the table with the
  `| Contract | Load it before |` header)
- Modify: `skills/myflow-contracts/SKILL.md` — its list of contained contracts

**Interfaces produced:** the tag vocabulary every later task references —
`verified:<how>`, `unverified:<what-to-check>`, `measured:<command> @ <ref>`,
`predicted:<what-confirms-it>`.

- [x] 1.1 Read `skills/myflow-contracts/state-file.md` and `skills/myflow-contracts/jira-integration.md`
  in full before writing. They set the house voice for a contract file: a canonical-authority
  statement near the top, dense rationale, and worked examples. Match it.

- [x] 1.2 Create `skills/myflow-contracts/plan-provenance.md`. It MUST contain, in this order: a
  one-line statement that the file is canonical for plan provenance; the four tags with their exact
  syntax; the asymmetry rule (every code block needs a tag; a number with no tag may not appear at
  all) **with the reason** — "194 tests" was invented, not mislabelled, so labelling alone would not
  <!-- measured: ./gradlew test @ c515c42 (intellij-review-queue) -->
  have caught it; the guard's scope and the reason for it; and the implementer's duty.

- [x] 1.3 Write the tag syntax section using these exact examples, so the guard in task 2 and the
  contract cannot drift:

````markdown verified:authored in-tree for this change
```kotlin verified:javap intellij.platform.diff.jar
override val toolWindowIds: Array<String>
```

```kotlin unverified:confirm the member is a property, not a function
override fun getToolWindowIds(): Array<String>
```

Baseline: 197 tests, 0 failures
<!-- measured: ./gradlew test @ c515c42 -->

After the deletion: 186 tests
<!-- predicted: ./gradlew test after task 1 -->
````

- [x] 1.4 State the guard's scope in the contract, and **why it is narrow**: only a change's
  `tasks.md`, excluding archived changes. Cite `scripts/check-references.sh`'s own header, which
  records that an over-firing guard in this repository produced 28 false failures on its own tree,
  after which suppression markers were the only way to silence them — and those markers then
  switched off the real checks sharing those lines.

- [x] 1.5 State plainly what the guard does **not** do: it checks that provenance is *stated*, never
  whether it is true. No script can confirm a verification was performed.

- [x] 1.6 Add a row to the contracts table in `rules/myflow-manual-review.mdc`. Match the existing
  rows' shape exactly — bold contract name, backticked path, and the "load it before" trigger:

````markdown verified:authored in-tree for this change
| **Plan provenance** (`skills/myflow-contracts/plan-provenance.md`) | writing or checking a plan's provenance tags |
````

- [x] 1.7 Add the new contract to `skills/myflow-contracts/SKILL.md`'s list of contained contracts,
  matching how the other five are listed there.

- [x] 1.8 Run `./scripts/check-references.sh`. It polices exactly what this task adds —
  cross-references between bold section names and backticked paths. Expected: `all referenced
  sections resolve`. If it fails, the new file's references are wrong; fix the reference, never the
  guard.

- [x] 1.9 Run `./scripts/check-vocabulary.sh`. Expected: both guards clean.

- [x] 1.10 Stage: `git add skills/myflow-contracts/ rules/myflow-manual-review.mdc`

## 2. The guard and its harness

TDD applies: the harness is written first and must fail before the guard exists.

**Files:**
- Create: `scripts/check-plan-provenance.sh`
- Create: `scripts/test-check-plan-provenance.sh`

**Interfaces:**
- Consumes: the tag vocabulary from task 1.
- Produces: `scripts/check-plan-provenance.sh`, honouring `CHECK_PLAN_PROVENANCE_ROOT` as its
  scan-root override; exit 0 clean, non-zero with `file:line` output on any violation.

- [x] 2.1 Read `scripts/check-references.sh` and `scripts/test-check-references.sh` in full. The
  guard must mirror their structure: `set -euo pipefail`, a `REPO_ROOT` resolved from
  `BASH_SOURCE` by default with a single explicit env override, and `file:line` output. The harness
  must mirror its fixture-sandbox approach — fixtures under `TMPDIR`, never touching the real tree.

- [x] 2.2 Write the harness first, at `scripts/test-check-plan-provenance.sh`, covering **both**
  directions and all four scope rules. These are the cases; write them as fixtures in the
  `test-check-references.sh` style:

  | Fixture | Expect |
  |---|---|
  | `tasks.md` with a block tagged `verified:x` | exit 0 |
  | `tasks.md` with a block tagged `unverified:x` | exit 0 |
  | `tasks.md` with an untagged fenced block | exit non-zero, names the file and line |
  | `tasks.md` with `197 tests` and a `measured:` comment on the next line | exit 0 |
  | `tasks.md` with `197 tests` and no provenance comment | exit non-zero |
  | `tasks.md` with `197 tests` and a `predicted:` comment | exit 0 | <!-- measured: ./gradlew test @ c515c42 (one attribution for the one real "197 tests" figure quoted across the three rows above — not one identical copy per row, which is what the pass-6 fix wave flagged as a copy-paste artifact, including on the row describing the ABSENCE of a comment) -->
  | untagged block under `openspec/changes/archive/x/tasks.md` | exit 0 — archives are not scanned |
  | untagged block in `openspec/changes/x/proposal.md` | exit 0 — only `tasks.md` is scanned |
  | untagged block in `skills/foo/SKILL.md` | exit 0 — out of scope |
  | a tagged 4-backtick block containing a 3-backtick block | exit 0 — nesting must not miscount |

- [x] 2.3 Run the harness. Expected: it fails because `scripts/check-plan-provenance.sh` does not
  exist. Confirm the failure is "guard not found", not a harness bug.

- [x] 2.4 Write `scripts/check-plan-provenance.sh`. The fence walk must follow CommonMark's rule
  that a closing fence is at least as long as its opener, or the last row of 2.2's table fails:

````bash unverified:this is the intended shape; verify the regexes against real fixtures before relying on them
# Walk lines. Outside a fence, a line matching ^(`{3,}|~{3,})(.*) opens one:
#   - remember the fence character and its LENGTH
#   - the remainder is the info string; it must contain verified: or unverified:
# Inside a fence, only a line of the SAME character, at least as long, with no
# info string, closes it. A shorter or different fence is content, which is what
# makes a 4-backtick block quoting a 3-backtick block work.
````

- [x] 2.5 Implement the numeric rule **narrowly**, because this is where over-firing lives. A line
  containing a digit group immediately followed by one of a small closed set of unit words is a
  numeric claim; anything else is not:

````bash unverified:tune the unit list against this repo's real plans before finalising
UNITS='tests|failures|errors|files|lines|seconds|minutes|ms'
# A claim line matches:  [0-9][0-9,]*[[:space:]]+($UNITS)\b
# It is satisfied when a `measured:` or `predicted:` HTML comment appears on the
# same line or within the following 2 lines.
````

  Deliberately NOT matched: `step 1.2`, `line 452`, `IU-262`, `KAN-14`, version numbers. A guard that
  flags those is the over-firing failure this scope exists to avoid.

- [x] 2.6 Write the guard's header comment. It states the rule **and the failure that motivated it**
  — that is the house convention, and `check-references.sh`'s header is the model. Name the KAN-6
  `DiffRequest` case: a plan snippet that would have made the guard permanently false and the feature
  a no-op every unit test still passed.

- [x] 2.7 Run the harness. Expected: every case passes. Fix the guard, never the assertions — an
  assertion loosened to make a guard pass is the guard deleting itself.

- [x] 2.8 Run the guard against the real tree with no override:
  `./scripts/check-plan-provenance.sh`. Expected: it scans `openspec/changes/*/tasks.md`, excluding
  `archive/`, and reports what it finds. Do not act on failures yet — task 3 does.

- [x] 2.9 `chmod +x` both scripts, matching the existing four.

- [x] 2.10 Stage: `git add scripts/`

> **Post-implementation note, added during the fix wave:** this task group's Bash pseudocode above
> (2.4, 2.5) describes the guard as it was first shipped. Five review panel passes and seven fix
> waves later, the classifier was rewritten from Bash to a real block-structure parser in Python 3
> (`scripts/check-plan-provenance.py`), keeping `scripts/check-plan-provenance.sh`'s CLI contract
> byte-identical as a thin `exec` wrapper. This task's steps are not rewritten to match — the
> Bash-pseudocode snippets above are history, not current instructions — see design.md's
> "Post-review reshape" section for the decision record and `check-plan-provenance.py`'s own module
> docstring for the full defect history that motivated it.

> **Further fix waves, added after the above note was written:** passes 9 and 10 each found and
> fixed additional defects in the same container-prefix machinery (a blockquote-to-list-marker gap
> tolerance, and its knock-on effects on the recorded list content column and the abort-eligibility
> check). Not restated here as a count or a list — `check-plan-provenance.py`'s own module
> docstring and `.superpowers/sdd/fix-wave-pass9.md`/`fix-wave-pass10.md` are the source of truth
> for what each wave found and fixed, for the same reason the pass-6/7/8 history above points at the
> docstring rather than duplicating it.

## 3. Wire the guard into this repository's configuration

**Files:**
- Modify: `.myflow/project.md` — `## lint` and `## test`

- [x] 3.1 Add `scripts/check-plan-provenance.sh` to the `## lint` block and
  `scripts/test-check-plan-provenance.sh` to the `## test` block in `.myflow/project.md`. Keep the
  existing statement that no auto-fix command exists — it stays true.

- [x] 3.2 Run the full declared lint set exactly as a myflow run would:

````bash verified:these are the commands declared in .myflow/project.md at d38372a, plus the new guard
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
````

- [x] 3.3 Run the full declared test set:

````bash verified:these are the commands declared in .myflow/project.md at d38372a, plus the new harness
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
````

- [x] 3.4 **The guard now scans this change's own `tasks.md` and `kan-8-myflow-updates/tasks.md`.**
  This file was written to satisfy the rule, so it must pass.

  **Correction, made during the fix wave:** an earlier draft of this task measured `kan-8`'s plan
  with `grep -c '^```'` and concluded it "contains zero fenced code blocks, so the block rule cannot
  fire on it." That command is anchored to column 0; every one of `kan-8`'s fences is indented under
  a list item, which is exactly the shape column-0 anchoring cannot see. The measurement was true
  about what it counted and false about what it was taken to mean — the earlier figure came from a
  measurement blind to the shape that mattered.

  The real count, taken with a leading-whitespace-tolerant pattern:
  52 fence lines (26 blocks, one pair correctly collapsed as a 4-backtick-wrapping-3-backtick case)
  <!-- measured: grep -cE '^[[:space:]]*(`{3,}|~{3,})' openspec/changes/kan-8-myflow-updates/tasks.md @ branch openspec/kan-14-plan-provenance -->
  and the guard, once it could see indentation, in fact fires **25 times** on that file
  <!-- measured: ./scripts/check-plan-provenance.sh 2>&1 | grep -c 'kan-8-myflow-updates/tasks.md' @ branch openspec/kan-14-plan-provenance -->
  — the block rule does fire on it, repeatedly. **Do not tag `kan-8`'s plan** — see the
  handoff note in `.myflow/project.md` for why the fix is to let those hits clear when `kan-8`
  archives, never to narrow the guard or add a suppression marker.

- [x] 3.5 Stage: `git add .myflow/project.md openspec/`

## 4. `/myflow-start` writes the tags

**Files:**
- Modify: `skills/myflow-start/SKILL.md` — section D (the writing-plans step, around the line reading
  "paths, verification commands, bite-sized steps, no placeholders")

- [x] 4.1 Read section D of `skills/myflow-start/SKILL.md` in full, plus its Guardrails list, before
  editing. The edit adds a duty to an existing step; it must not restate the step.

- [x] 4.2 Add the tagging duty to section D: while enriching `tasks.md`, tag every fenced block and
  every numeric claim per **Plan provenance** (`skills/myflow-contracts/plan-provenance.md`). State
  that code which cannot be verified is tagged `unverified:` and **kept** — a plan without the
  snippet is worse than a plan with a labelled guess.

- [x] 4.3 Add the guard run to section D, before the artifact is published: run the project's
  configured plan-provenance guard if it declares one, and fix any hit before publishing.

- [x] 4.4 Add one Guardrail line to that skill's Guardrails list, matching the existing terse style:
  never publish a plan carrying an untagged block or an unsourced number.

- [x] 4.5 Run `./scripts/check-references.sh` — this edit adds a bold-token-plus-path reference to
  the new contract, which is exactly what that guard checks. Expected: clean.

- [x] 4.6 Run `./scripts/check-vocabulary.sh`. Expected: clean.

- [x] 4.7 Stage: `git add skills/myflow-start/SKILL.md`

## 5. `/myflow-do` makes the implementer act on the tags

This is where the tag converts into saved time. Without it the tag is decoration.

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — section 4 (Execute), the block of standing clauses every
  implementer dispatch must carry

- [x] 5.1 Read section 4 of `skills/myflow-do/SKILL.md`. It currently carries three standing clauses
  in blockquote form — **MYFLOW — NO COMMITS**, **REQUIRED SUB-SKILL**, and **REQUIRED READING**.
  The new clause joins them and must match their voice and formatting exactly.

- [x] 5.2 Add the fourth clause, verbatim:

````markdown verified:authored in-tree for this change
> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. A block tagged
> `verified:<how>` was checked as stated; if it does not compile, report that — do not contort the
> code to match it.
````

- [x] 5.3 Add a Guardrail line to that skill's Guardrails list: never dispatch an implementer
  without the provenance clause.

- [x] 5.4 Run `./scripts/check-references.sh` and `./scripts/check-vocabulary.sh`. Expected: both
  clean.

- [x] 5.5 Stage: `git add skills/myflow-do/SKILL.md`

## 6. Verify the whole change, including the installer

**Files:**
- No new edits expected. This task is verification; any fix it forces belongs to the task that
  introduced the fault.

- [x] 6.1 Confirm `setup.sh` needs no change by running it against a sandbox home, exactly as
  `.myflow/project.md`'s `## run` key documents:

````bash verified:this is the command declared under ## run in .myflow/project.md at d38372a
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
````

- [x] 6.2 Assert the new contract landed in the sandbox home. It must appear in **all three** harness
  trees, because `setup.sh` copies the contracts directory wholesale and the contract is only useful
  where the skills that read it are installed:

````bash verified:paths taken from the harness list in rules/myflow-manual-review.mdc at d38372a
ls "$SANDBOX/.claude/skills/myflow-contracts/plan-provenance.md"
ls "$SANDBOX/.cursor/skills/myflow-contracts/plan-provenance.md"
ls "$SANDBOX/.codex/skills/myflow-contracts/plan-provenance.md"
````

  If any is missing, `setup.sh` does **not** copy the directory wholesale and the proposal's
  "no installer change" claim is wrong — report that rather than editing `setup.sh` silently.

- [x] 6.3 Run every guard and harness one final time and record the real output:

````bash verified:the full declared set from .myflow/project.md after task 3
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
````

- [x] 6.4 Confirm the guard is not vacuous: temporarily strip a provenance tag from one block in this
  file, re-run `./scripts/check-plan-provenance.sh`, confirm it **fails** naming that line, then
  restore the tag. A guard that passes on a broken input is worse than no guard. Do not leave the
  file modified.

- [x] 6.5 Confirm no suppression marker was added anywhere in the change:
  `git diff --cached | grep -i "suppress\|noqa\|nolint"` — expected: no output.

- [x] 6.6 Stage everything: `git add -A`

## 7. Self-heal must carry forward the fields it did not infer

Added to this change during implementation, at the operator's direction, after a live state file was
damaged during this very run.

**What happened.** `/myflow-start` wrote `kan-14-plan-provenance.json` with `artifactUrl` and
`jiraIssue` populated. Twenty-five seconds later `/myflow-status` rewrote it with both fields `null`.
<!-- measured: jq -r .updatedAt "$STATE_DIR/kan-14-plan-provenance.json" and its .damaged-backup
     copy — 07:27:52 then 07:28:17 @ the live state file during this run's /myflow-start and
     /myflow-status, which is outside the repository and therefore not re-runnable after the fact;
     the backup file is the only surviving artifact -->
Another change's file, `kan-8-myflow-updates.json`, shows the same `updatedBy` and kept its fields,
so the loss is intermittent — consistent with a race in which the prior file was momentarily
unreadable.

**Why it is not simply "a read-only command wrote".** `skills/myflow-status/SKILL.md:12` sanctions
self-heal as the one exception, and `:51` instructs it to *rewrite the state file* from inferred
truth. The gap is that neither that line nor **State self-heal**
(`skills/myflow-contracts/state-self-heal.md`) restates the carry-forward duty that **State file**
(`skills/myflow-contracts/state-file.md`) imposes on every write — so an inference-only rewrite has
no source for `artifactUrl` or `jiraIssue` and lands them as `null`. The contract forbids the
outcome; nothing in the path prevents it; nothing detects it.

**Files:**
- Modify: `skills/myflow-contracts/state-self-heal.md`
- Modify: `skills/myflow-status/SKILL.md` — the rewrite instruction
- Create: the delta spec at
  `openspec/changes/kan-14-plan-provenance/specs/myflow-state-integrity/spec.md` (already written by
  the parent; read it as your requirements)

- [x] 7.1 Read `skills/myflow-contracts/state-self-heal.md` and `skills/myflow-contracts/state-file.md`
  in full, plus `skills/myflow-status/SKILL.md` lines 10-65. The fix is a contract clarification, not
  a behaviour invention — the duty already exists in `state-file.md` and must be restated where the
  rewrite happens.

- [x] 7.2 In `state-self-heal.md`, state explicitly that a self-heal rewrite **carries forward every
  field it did not itself infer**. Self-heal infers `state` only; `artifactUrl`, `jiraIssue`,
  `prUrl`, `branch` and `worktrees` are re-emitted exactly as read.

- [x] 7.3 In the same file, handle the case that produced this bug: when the prior file is missing or
  unparseable, the unowned fields have **no source**. Writing `null` silently is what destroyed the
  published proposal link. Require instead that the correction announcement names every field that
  could not be recovered, so the loss is visible rather than silent.

  A worked example belongs here, matching the file's existing `⚠ state corrected:` shape.

- [x] 7.4 In `skills/myflow-status/SKILL.md`, at the line instructing the rewrite, restate the
  carry-forward duty and reference the contract. Do not restate the whole contract — one sentence and
  a pointer, matching how that file references its other contracts.

- [x] 7.5 **Say plainly in your report that this task has no automated test.** This repository's
  executable surface is its guard scripts; a skill's runtime behaviour is not reachable from them.
  Verification is `./scripts/check-references.sh` (the new cross-references),
  `./scripts/check-vocabulary.sh`, and human review. Do not invent a test that cannot fail.

- [x] 7.6 Run `./scripts/check-references.sh` and `./scripts/check-vocabulary.sh`. Expected: both
  clean.

## 8. REVERTED — the `openspec/specs` widening was tried and rejected

> **This task group was implemented, reviewed, and then deliberately undone.** It is kept rather than
> deleted because the reasoning is the useful part.
>
> **What it did:** added `openspec/specs` to `check-vocabulary.sh`'s `DEFAULT_TARGETS`, so a dead
> capability spec could not hide from the vocabulary guard. It worked — it immediately surfaced two
> dead capabilities.
>
> **Why it was reverted:** the adversarial reviewer built fixtures from `kan-8-myflow-updates`'s
> delta bodies — the text that *becomes* the live spec once that change archives — and ran the real
> guard against them. **19 hits across 5 files**
> <!-- measured: ./scripts/check-vocabulary.sh $(find openspec/changes/kan-8-myflow-updates/specs -iname spec.md) 2>&1 | grep -cE '^openspec/changes/kan-8-myflow-updates/specs/[^:]+:[0-9]+:' @ branch openspec/kan-14-plan-provenance -->
> (`agents-repo-verification`, `myflow-command-surface`, `myflow-contract-distribution`,
> `myflow-state-advance`, `myflow-state-machine`). The cause is not kan-8's: **a requirement that
> forbids a term must name that term.** One reads verbatim `SHALL NOT contain `gates`, `tested`,
> `originStage`, `REVIEWED_TREE`, `fastPath``. Scanning live specs for retired vocabulary flags
> requirements for doing their job, and there is no honest repair — rewording deletes the
> requirement, and a suppression marker is forbidden by `CLAUDE.md`.
>
> **What survived:** nothing from this group. The two dead capabilities it found are already removed
> by kan-8's own deltas, which is why this change's duplicate deltas were also dropped.
>
> **What replaced it:** a short comment at `DEFAULT_TARGETS` recording that this was tried and
> rejected, so the next person does not re-propose it.

## 9. Remove the remaining legacy, with no backward compatibility

Added during implementation at the operator's direction: *"remove all legacy things here too,
backward compatibility is not needed for now."*

Everything below was classified by inspection, not assumption. **Three hits were examined and
deliberately kept** — see 9.5.

**Files:**
- Modify: `skills/myflow-contracts/state-file.md` — the dead-field tombstone paragraph
- Modify: `skills/myflow-do/SKILL.md` — the `originStage` half of one sentence
- Modify: `scripts/check-references.sh` — one comment example
- The capability removal originally cited a delta spec at
  `openspec/changes/kan-14-plan-provenance/specs/myflow-review-panel-economics/spec.md` — that
  delta was dropped after pass 1 as a duplicate of `kan-8`'s identical delta (see task 8's note
  above, "this change's duplicate deltas were also dropped"); this change's `specs/` directory
  carries no such capability. `/myflow-finish` deletes `openspec/specs/myflow-review-panel-economics/`
  at archive time via `kan-8`'s own delta, not this change's.

- [x] 9.1 Read the tombstone at `skills/myflow-contracts/state-file.md` (the paragraph beginning
  "There is no `gates` object…", carrying three `vocab-guard:allow` markers) and the one at
  `skills/myflow-do/SKILL.md` (the sentence containing "There is no `originStage` and no fix
  re-entry table", carrying one).

- [x] 9.2 **Rewrite, do not simply delete.** Each tombstone carries a live rule welded to a dead
  vocabulary. Keep the rule, drop the vocabulary:

  - `state-file.md`: the live content is *why* the fields went — nothing records a human confirmation
    now that the `*-done` commands are gone, and a fix never moves the state so there is no origin to
    return to. State that positively, naming no removed field.
  - `myflow-do/SKILL.md`: the live rule is "a fix never moves the state". Keep exactly that; drop
    "There is no `originStage` and no fix re-entry table".

  **The mechanical ban survives regardless.** `scripts/check-vocabulary.sh` already fails on every one
  of these terms, and that guard — not a paragraph — is what prevents reintroduction. The prose was
  redundant enforcement that cost four suppression markers.

- [x] 9.3 Remove the four now-unnecessary `vocab-guard:allow` markers from those two files. A
  suppression marker that outlives its reason is the thing this repository's own guard history warns
  about.

- [x] 9.4 In `scripts/check-references.sh`, one comment uses `fastPath: true` as its worked example
  of a bold-wrapped code span and carries a `vocab-guard:allow` for it. Replace the example with a
  live field — `prUrl: null` reads identically for the purpose — and drop that marker too. **Change
  the comment only; do not touch the extraction logic**, which is load-bearing and covered by
  `test-check-references.sh`.

- [x] 9.5 **Three hits were examined and are deliberately NOT changed.** Verify each is still true
  and leave it alone:

  | Hit | Why it stays |
  |---|---|
  | `scripts/check-vocabulary.sh`'s ban list | It is the enforcement mechanism. It must name the dead terms to forbid them. |
  | `scripts/test-setup.sh` `mktemp -d /tmp/myflow-test-setup.XXXXXX` | Coincidental substring in a temp-dir name, not a reference to the deleted `/myflow-test`. |
  | `openspec/changes/**` and `docs/**` history | Archived plans, design docs and review records legitimately quote the old vocabulary. Rewriting history to remove a word is not legacy removal. |

- [x] 9.6 Run `./scripts/check-vocabulary.sh` and `./scripts/check-references.sh`. Expected after
  tasks 8 and 9 together: `check-references.sh` clean; `check-vocabulary.sh` failing **only** on the
  two dead spec directories that `/myflow-finish` deletes at archive time. Report the exact hits.

- [x] 9.7 Run `./scripts/test-check-references.sh` — 9.4 edits that guard's source, so its own
  harness must still pass.

## 10. Implementer model policy (added after implementation, at the operator's direction)

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` — the `## Model policy` section
- Modify: `skills/myflow-do/SKILL.md` — step 4, where implementers are actually dispatched
- Create: `openspec/changes/kan-14-plan-provenance/specs/myflow-model-policy/spec.md`
- Modify: `commands/myflow-do.md` — its `**Model:**` line said the opposite of the new rule
  ("Opus is reserved for `/myflow-start`'s brainstorming stage"), and `pipeline.md` designates that
  file as how this policy reaches Cursor. `commands/myflow-{finish,status,info}.md` carry the same
  sentence and are **correct** — they dispatch no implementers — so they are deliberately untouched.
- Modify: `skills/README.md`, `CLAUDE.md`, `AGENTS.md`, `README.md` — each summarises the model
  policy and was silent on the implementer rule.
- Modify: `openspec/changes/kan-14-plan-provenance/design.md` — the decision block
  (`implementers-at-the-ceiling`), with the rejected alternative named.

**Why this is here and not in a follow-up:** the gap was found while auditing this change's own
records, and the fix is three prose edits plus a delta spec. Deferring it would leave the pipeline
silently inheriting subagent-driven-development's economy default — the behaviour that produced an
unrecorded, unverifiable model history for this change.

- [x] 10.1 In `pipeline.md`'s `## Model policy`, state that implementer subagents dispatched by
  `/myflow-do` run on Opus (or the strongest available model), and that this **explicitly overrides**
  superpowers:subagent-driven-development's "least powerful model that can handle each role"
  guidance for this pipeline. Name the conflict rather than leaving two documents to disagree.

- [x] 10.2 In the same section, note that review-panel slots remain on Sonnet — as non-normative
  context only, since `myflow-review-panel-economics` (kan-8) already owns that requirement and
  restating it as an ADDED requirement in a new capability would put two SHALLs for one obligation
  into the live specs, which `openspec validate --strict` cannot detect. State that an explicit
  operator instruction overrides either default in either direction. Also name the two further
  upstream instructions this pipeline overrides (dispatch the final review on the most capable
  model; escalate the model in fix rounds 4-5), and scope the "enforced via frontmatter" bullet:
  frontmatter sets the *session's* model and cannot set a subagent's.

- [x] 10.3 In `myflow-do/SKILL.md` step 4 (Execute), add one line pointing at that rule where
  implementers are dispatched — the panel table already names its own models, so the implementer
  rule must be visible at the same distance.

- [x] 10.4 Require the SDD ledger to record the model used for each dispatch. Without it the policy
  is unverifiable, which is the defect that surfaced it. State the record's scope honestly — it
  lives under gitignored `.superpowers/` in a worktree `/myflow-finish` removes, so it is
  session-scoped and does not answer an after-the-fact audit. Make `unknown (agent-defined)` a
  legal value for slots dispatched by `subagent_type`, whose model the dispatcher cannot observe.
  Backfill this change's own ledger, which held zero model entries.

- [x] 10.5 Write the delta spec `specs/myflow-model-policy/spec.md` with requirements for the
  implementer default, the operator override, and the ledger record — **three**, not four: the
  panel default is kan-8's and is carried here only as rationale.

- [x] 10.6 Run `openspec validate --strict kan-14-plan-provenance`; then the full declared lint set.
