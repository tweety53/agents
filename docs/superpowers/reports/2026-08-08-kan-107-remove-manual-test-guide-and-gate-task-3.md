# Task 3 report: Drop `docs/manual-test/` from the staging exclusions

## Status

DONE_WITH_CONCERNS (see "Grep sweep beyond the declared Files list" below — Step 5's grep is not
fully empty, and I deliberately left a set of out-of-scope hits rather than expand past the brief's
declared `Files:` list).

## What was test-driven vs. contract-only

- **Test-driven (RED→GREEN):** Step 1 amended `scripts/test-uncommitted-review-package.sh` to drop
  the `docs/manual-test` fixtures and assertions from the two combined-exclusion cases (the
  `4.11 checkpoint …` and `4.12 uncommitted-review-package …` blocks), reducing both to the two
  surviving paths (`openspec/`, `docs/superpowers/`). Step 2 ran the harness — it was green
  immediately, exactly as the brief predicted, because the harness exercises the helper scripts
  (`checkpoint`, `uncommitted-review-package`), not the contract prose. No RED phase was possible
  here in the classic sense since the assertions were being *removed*, not added — this was
  harness-reduction, not behavior-driven addition. I then also edited the two real scripts
  `skills/myflow-do/scripts/checkpoint` and `skills/myflow-do/scripts/uncommitted-review-package`
  (git add exclusion pathspecs and their header comments) to actually drop `docs/manual-test/` from
  the staging exclusion — the harness stayed green after this because with no `docs/manual-test`
  fixtures created, the pathspec change is behavior-invisible to the current test, but it is the
  real fix the task title describes.
- **Contract-only, no test to drive them:** Steps 3–4, all Markdown edits — `pipeline.md`,
  `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, `finish-contract.md`,
  `pipeline-rationale.md`. These are prose/contract text; `check-references.sh`,
  `check-plan-provenance.sh`, `check-vocabulary.sh`, `check-contract-budget.sh` and
  `check-task-build-green.sh` are the closest thing to guards here and all pass, but none of them
  assert the specific two-vs-three-path content — that's confirmed by the Step 5 grep instead.

## Files changed

- `scripts/test-uncommitted-review-package.sh` — dropped `docs/manual-test` fixtures/writes/asserts
  from the two combined-exclusion test cases; reduced the grep alternation to
  `^(openspec/|docs/superpowers/)`.
- `skills/myflow-contracts/pipeline.md` — "the three that Handoff output names" → "the two that…";
  git chain (lines ~145-146) dropped `docs/manual-test/` from both `reset` and `add` pathspecs, and
  the second commit subject changed `"chore(<name>): plan, test guide and session records"` →
  `"chore(<name>): plan and session records"`; symlink-rule sentence "any of the three is a symlink"
  → "either of the two is a symlink" (its example `git add` invocation already used
  `docs/superpowers/`, so no swap was needed there); Handoff-output bullet (~line 245) dropped
  `docs/manual-test/` from the "never stages …" list.
- `skills/myflow-do/SKILL.md` — NO-COMMITS block (line ~131) dropped `docs/manual-test/`; the two
  bare staging lines (~481-482) dropped it from both `reset` and `add`; "Those three paths are never
  staged" → "Those two paths…"; "picks up the three excluded paths" → "…the two excluded paths".
- `skills/myflow-finish/SKILL.md` — "All three routes commit — implementation, `docs/manual-test/…`,
  the `openspec/` planning artifacts…" dropped the `docs/manual-test/` clause (the "all three"
  there names the three *landing routes* — PR/merge/manual — not the planning-path count, so that
  phrase itself was left alone); git chain (~158-159) same two-path + commit-subject treatment as
  `pipeline.md`.
- `skills/myflow-contracts/finish-contract.md` — line ~126 dropped `docs/manual-test/<name>.md,`
  from the "the implementation, then …" sentence; "Those three planning paths" → "Those two planning
  paths"; "picks the three paths up" → "…the two paths up". ("All three routes" earlier in the same
  file is again the three landing routes, left alone.)
- `skills/myflow-contracts/pipeline-rationale.md` (not in the brief's `Files:` list, but carries the
  identical "never stages `openspec/`, `docs/manual-test/` or `docs/superpowers/`" sentence as
  `pipeline.md`'s Handoff-output bullet) — dropped `docs/manual-test/`.
- `skills/myflow-do/scripts/checkpoint` and `skills/myflow-do/scripts/uncommitted-review-package`
  (also not in the brief's `Files:` list) — dropped `docs/manual-test/` from both the header comments
  ("the three NO-COMMITS planning paths" → "the two NO-COMMITS planning paths") and the actual
  `git add -A -- . ':(exclude)…'` pathspecs. These are the scripts that physically implement "the
  staging exclusions" the task title names, so I judged them in-scope even though the brief's
  `Files:` list didn't cite them; `scripts/test-uncommitted-review-package.sh` (which invokes both
  via `$CHECKPOINT`) still passes.

## Test summary

`scripts/test-uncommitted-review-package.sh` passes (30/30 assertions). Full project test list
(`.myflow/project.md`'s `## test` block, 12 scripts) and full lint list (`## lint` block, 6 guards)
all pass, including `check-contract-budget.sh` (`BUDGET-OK: 24 contract file(s) within budget` — no
row needed raising, since every touched contract file shrank).

## Concerns / open items

**Step 5's grep is not fully empty**, contrary to the brief's stated acceptance bar. After my edits:

```
skills/myflow-do/SKILL-rationale.md:79:In the same run, write or refresh `docs/manual-test/<name>.md`. …
skills/myflow-do/SKILL.md:327:In the same run, write or refresh `docs/manual-test/<name>.md`. See …
skills/myflow-do/SKILL.md:536:Test guide: <absolute path to docs/manual-test/<name>.md>
scripts/test-check-unfinished-work.sh:192,196,198,777   (explicitly protected — untouched)
scripts/test-check-references.sh:111:printf 'see **Whatever** in `docs/manual-test/<name>.md`\n' \
scripts/check-references.sh:414:      # docs/manual-test/<name>.md are legitimate … (Task 6's comment site)
commands/myflow-do.md:14:Produces **both** the staged diff **and** `docs/manual-test/<name>.md`, …
commands-claude/myflow-do.md:10:Produces **both** the staged diff **and** `docs/manual-test/<name>.md`, …
```

I left these deliberately rather than edit them, for two different reasons:

1. `scripts/test-check-unfinished-work.sh` — explicitly protected by the brief's own instruction
   ("Tasks 1 and 2 have already landed uncommitted changes in … `scripts/test-check-unfinished-work.sh`
   … Do not touch any of them").
2. `skills/myflow-do/SKILL-rationale.md:79`, `skills/myflow-do/SKILL.md:327` and `:536`,
   `commands/myflow-do.md:14`, `commands-claude/myflow-do.md:10`, `scripts/test-check-references.sh:111`
   — all six are about the manual-test-**guide-generation** feature itself (the "6. Write the manual
   test guide" step, the handoff's "Test guide:" line, the command description, and an arbitrary
   fixture path string), not about the "staging exclusions" this task's title and `Files:`/`Produces:`
   lines scope it to. None of these five files appear in the brief's `Files:` list, and fixing them
   would mean removing or rewriting the guide-generation instruction itself — a behavior change past
   what Task 3's declared interface (`the planning paths are exactly openspec/ and docs/superpowers/,
   everywhere`) commits to. That reads as another numbered task's job in the eight-task plan (guide
   generation removal), and `scripts/check-references.sh:414`'s companion two-comment-site carve-out
   already tells me Task 6 does a further `docs/`-related sweep later, which is consistent with these
   being left for a later task rather than mine.

The brief's Context section does say plainly "anything else it hits is yours," which taken literally
would include these six sites. I chose not to expand into them because doing so means changing what
`/myflow-do` actually produces (dropping the guide-writing step / its handoff line), which contradicts
the task's own `Files:` list and `Produces:` line, and — if a later task in this plan owns guide-
generation removal with its own brief and its own line numbers — pre-empting it here risks a
conflicting edit. Flagging this explicitly per plan-provenance guidance ("confirm each fits its
surroundings … report anything that does not") rather than silently picking a side. If the intent
really was for Task 3 to also strip these six, they're a small, mechanical follow-up.

No other concerns. `check-contract-budget.sh` needed no ratchet change. The `docs/manual-test/`
directory itself was left untouched, as instructed (Task 6 deletes it).

## Fix round (review finding)

Review adjudicated the six guide-generation sites (`skills/myflow-do/SKILL-rationale.md:79`,
`skills/myflow-do/SKILL.md:327` and `:536`, `commands/myflow-do.md:14`,
`commands-claude/myflow-do.md:10`, `scripts/test-check-references.sh:111`) as correctly left alone —
they belong to Tasks 4, 6 and 7, and the brief's unqualified "no output" wording for Step 5 was the
brief's own defect, not a gap in this task's work. No change made for those; confirmed still present
and still intentionally untouched.

One finding required a fix: **`skills/myflow-contracts/pipeline-rationale.md:30`** still read "a fix
touching only **the three planning paths** leaves the implementation commit empty" — a path *count*
stated in words, which Step 5's literal `docs/manual-test` grep can't catch since the sentence names
no path. Fixed to "a fix touching only the two planning paths leaves the implementation commit
empty."

**Sweep performed:** every file this task touched (`scripts/test-uncommitted-review-package.sh`,
`skills/myflow-contracts/pipeline.md`, `skills/myflow-contracts/pipeline-rationale.md`,
`skills/myflow-contracts/finish-contract.md`, `skills/myflow-do/SKILL.md`,
`skills/myflow-do/scripts/checkpoint`, `skills/myflow-do/scripts/uncommitted-review-package`,
`skills/myflow-finish/SKILL.md`) was grepped for `three` and cross-checked against `path`,
`planning path` and `excluded path` context, not just the literal directory string.

Covering command run after the fix:

```bash
for f in scripts/test-uncommitted-review-package.sh skills/myflow-contracts/pipeline.md \
         skills/myflow-contracts/pipeline-rationale.md skills/myflow-contracts/finish-contract.md \
         skills/myflow-do/SKILL.md skills/myflow-do/scripts/checkpoint \
         skills/myflow-do/scripts/uncommitted-review-package skills/myflow-finish/SKILL.md; do
  grep -niE "three" "$f" | grep -iE "path"
done
```

Output: one line, from `skills/myflow-contracts/pipeline-rationale.md`:

```
53:sources are still attempted after any one failure, so a single bad path costs one record, not three.
```

**Left deliberately.** That "three" counts the three preserved *session-record sources* (the SDD
ledger, the review panel record, and the proposal artifact source, per the "Preserving the session
records" section a few lines above) — it is not a planning-path count, and the two-path fact it
would need to disagree with does not appear in that sentence. Confirmed by reading the surrounding
paragraph (`skills/myflow-contracts/pipeline-rationale.md:46-53`).

Every other `three` hit in the eight swept files was independently confirmed unrelated to planning
paths before being left alone: pipeline states/commands/model-roles/per-state-templates in
`pipeline.md`; landing routes (PR/merge/manual), `OUTSTANDING` courses, and
preserve-session-records outcomes in `pipeline-rationale.md`, `finish-contract.md`,
`skills/myflow-do/SKILL.md` and `skills/myflow-finish/SKILL.md`; review-panel required slots and
fix-round counts in `skills/myflow-do/SKILL.md`. None of those name a planning-path count.

**Re-verification after the fix:**

```
$ bash scripts/test-uncommitted-review-package.sh   → uncommitted-review-package: all cases pass
$ bash scripts/check-references.sh                  → check-references: all referenced sections resolve
$ bash scripts/check-contract-budget.sh              → BUDGET-OK: 24 contract file(s) within budget
```

`skills/myflow-contracts/pipeline-rationale.md` staged via `git add` (own work only, no commit, per
NO-COMMITS).
