# SDD ledger — kan-102-citations-resolve-to-installed-paths

Run: `/myflow-fast`, harness `claude-code`, session token `mf-k102a1`.

Recorded model roles (myflow-fast's speed defaults, no operator override given):
`models.implementation: sonnet` · `models.reviewPanel: sonnet` · `models.panelFix: sonnet` ·
`reviewPanelRoster: light` · `planningEffort: default`.

Dispatch bundles, from `plan-dispatch-bundles.sh`: `1 2` · `3` · `4` · `5 7` · `6` · `8`.

---

## Tasks 1 and 2 — the harness and the guard

- **Implementer:** general-purpose, model: **sonnet** (the recorded `models.implementation`).
- **Commit:** `070179e`, one commit for the squash unit — task 1 is `Build: red` with
  `Squash-with: Task 2`, and its commit was folded in with `--fixup` / `--autosquash` as planned.
  Trailers `Task-Id: 1` and `Task-Id: 2` both present.
- **Files:** `scripts/check-installed-citations.py`, `scripts/check-installed-citations.sh`,
  `scripts/test-check-installed-citations.sh`.
- **`check-task-commit-fields.sh . 2 070179e 332e593`:** exit 0. `Regression:` and `Baseline:`
  reported as skipped-not-verified, which is this project's standing case — its `## test` command
  cannot target a single named test.
- **Harness:** `check-installed-citations: all cases pass`, 23 cases (17 classifier, 4 refusal,
  2 report-shape) against the 21 the plan predicted.
- **Guard against the real corpus:** exit 1, **236 violations across 32 files**, 0.82s.
- **Per-task review:** combined (spec + quality), model **sonnet**, per the `light` roster.

### Two defects the implementer found and fixed while implementing

1. A single-backtick span holding a shell example with arguments — `` `agents/setup.sh global` `` —
   was classified as one garbled multi-word citation rather than as its leading path-like word.
   Fixed by taking the first whitespace-delimited word of a backtick span, the same base-token
   extraction `check-guard-symlinks.sh` already uses. Dropped the count 246 → 236.
2. The shell-variable exclusion (`$SCRIPT_DIR/…`) was **shadowed** by the regex/glob exclusion,
   since both matched on `$`. The mutation test for the shell-variable rule therefore passed with
   that rule disabled — exactly the untested-mutation shape KAN-197 names. Fixed by removing `$`
   from the regex/glob character set, making both rules independently load-bearing.

### Plan repair this task forced

The plan's Baseline claimed **200 violations across 31 files**, from a throwaway classifier probe run
during planning. The real guard reports **236 across 32**, because the probe filtered path-shaped
tokens more aggressively than the classifier the spec actually requires. The guard is the authority,
so the plan's Baseline and the three waves' arithmetic were repaired against it before wave A was
dispatched: wave A 29 → 41 sites (and `rules/dispatch-carries-the-baseline.mdc` added to its file
list), wave B 102 → 121, wave C 69 → 74. Task 1's predicted case count was likewise replaced with
the measured 23.

The residual gap the implementer could not attribute to a classifier defect — `` `N/M` ``,
`` `src/Foo.kt:42` `` in an illustrative table row, `` `../<other-app>` `` used as a negative
example — was reported rather than absorbed, and is left for the review panel to rule on.

### Per-task review round 1 — not clean

The reviewer ran the guard rather than reading it, which is what found the critical defect.

- **Critical.** `extract_backtick_tokens()` keeps only the first whitespace-delimited word of a
  backtick span — the fix for defect 1 above. But `<agents repo>/…` is *itself* two words, so every
  citation carrying that prefix was discarded before `classify_token`/`judges_ok` ever ran. The
  guard reported a bogus `<agents repo>/does-not-exist/nested/path.md` as clean. Fatal to the
  change: waves 4-6 rewrite ~121 sites into exactly that form, so the guard would have enforced
  nothing on any of them while exiting 0.

  The `agents-repo-prefix` harness case passed **for the wrong reason** — deleting the
  `<agents repo>/` branch from `judges_ok` outright left it green, because the case never reached
  that function. A second untested-mutation, found the same way KAN-197 says to look.
- **Important.** The same root cause inverted: fence *comment* lines are split on whitespace rather
  than by backtick extraction, so an un-backticked `<agents repo>/README.md` in a comment splits and
  the fragment `repo>/README.md` is reported as a citation naming no root. No corpus occurrence yet;
  waves 4-6 are about to create the conditions.
- **Important.** `agents-repo-verification`'s requirement **Zero coverage SHALL be declared or SHALL
  be a violation** is unimplemented. Twelve members report `0` and are silently accepted.

- **Clean and verified by the reviewer, not merely asserted:** the installer derivation genuinely
  runs `setup.sh` with no re-implemented globs, and `newly-installed-directory` is a real test of
  that; the sandbox refusal cannot be reached with a `HOME` outside the sandbox; the exit-code and
  stdout/stderr contract holds live in all three states; the `$`-shadowing fix is real, and all six
  token-level exclusions were mutation-tested independently with no shadowing left among them.

**Fix round dispatched to the same implementer** (never a fresh one, never the reviewer), model
**sonnet**, folding into `070179e` by `--fixup` / `--autosquash`.

**A design consequence worth recording.** `<agents repo>` carries a space, and that is what made the
critical defect possible at all. The term is pre-existing — ten sites and a resolution procedure in
`project-configuration.md` — so the change keeps it rather than minting `<agents-repo>`, and pays
for it with a tokenizer that must special-case the literal. Any future tool that reads these
citations inherits the same obligation.

### Per-task review round 2 — clean

Same reviewer, verifying its own findings rather than accepting the implementer's account.

- **Critical, closed.** `merge_agents_repo_prefix()` rejoins the two-word placeholder before the
  leading-word cut, in both tokenizers. Re-run mutation: deleting the `<agents repo>/` / `<project>/`
  branch from `judges_ok` now *fails* the `agents-repo-prefix` case, where before the fix it stayed
  green. The new `agents-repo-prefix-bogus-path` case distinguishes "recognised" (coverage count 1)
  from "never seen" (absent), so the two can no longer collapse into one observable.
- **Important, closed.** The fence-comment repro now reports OK with coverage 1.
- **Important, closed.** The wrapper sources `scripts/lib/coverage.sh`, records per member, declares
  ten genuinely-citation-free files, and folds `coverage_verdict`'s answer into the violation report.
  The reviewer spot-checked all ten declarations by content and confirmed each is honest. The refusal
  contract survives the restructure: empty root, non-directory root and installer failure all still
  exit 2 with **zero bytes on stdout**, even though the Python's own exit code no longer carries the
  verdict.
- **Commit:** `377b7dd`. 27 harness cases. Corpus unchanged at 236 violations across 32 files.

**One concern I raised was wrong, and is recorded because the reasoning is worth keeping.** I
reported that the guard appeared to report clean over an empty scan set. It does not: my fixture was
malformed, and `scripts/lib/coverage.sh`'s own `coverage_verdict()` already fails an empty corpus —
protection every guard sourcing that library inherits, pre-dating this change and covered by
`scripts/test-lib-coverage.sh` case 4. The reviewer rebuilt the case with two well-formed fixtures
(one installing nothing, one installing only a non-Markdown file) and both exit 1 with
`coverage:0: no member was ever recorded`. No new harness case was warranted.

**Tasks 1 and 2: cleared for their checkboxes.**

---

## Task 3 — `check-references.sh` strips the prefix

- **Implementer:** general-purpose, model: **sonnet**. Task base `377b7dd`.
- Dispatched with the ordering rationale spelled out — this task lands before the rewrite so that a
  guard which stops checking without failing cannot happen — and with the `<project>/` containment
  trap called out explicitly, since reading that prefix as a traversal would turn every
  project-relative citation into a lint failure the moment wave A lands.

- **Commit:** `eeb4235`. `check-task-commit-fields.sh . 3 eeb4235 377b7dd` exit 0. Harness 37
  assertions pass; `check-references.sh` verdict unchanged; `check-installed-citations.sh` still 236.
- **Plan defect the implementer found:** step 4's `unverified:` comment guessed the harness tail line
  reads `check-references: all cases pass`. It actually reads `All check-references assertions
  passed`. The guess was wrong and the tag is what made it safe to act on — the block was a
  hypothesis, and the implementer checked it rather than transcribing it.
- **The RED step confirmed the task's own rationale.** Before the fix, the prefixed-citation case
  failed not with a stale-heading report but with `0 checked, and not declared expected-zero
  (coverage)` — direct evidence the prefixed path was being silently *skipped* rather than checked,
  which is precisely the silent coverage loss this task was ordered before the rewrite to prevent.

### Per-task review round 1 on task 3 — one important test-only finding

Production code verified correct on every point the review was pointed at, live rather than by
reading: the strip happens once before the candidate list is built; `contained()` still decides
lexically before any existence test, and still refuses
`<agents repo>/../../../../etc/passwd.md`; `<project>/` falls through the ordinary
does-not-resolve path; error messages report the citation **as written**, prefix intact.

- **Important, test only.** Case 27 is **unfalsifiable**. Its fixture never places a file at the path
  that would result from stripping `<project>/`, so "correctly left alone" and "incorrectly stripped"
  are indistinguishable — both leave the citation unresolved, `RC=0`, no escape message. Patching
  `check-references.sh` to also strip `<project>/` — the exact regression the case claims to lock
  out — leaves the harness reporting `All check-references assertions passed`. The reviewer proved
  the gap is the fixture's construction rather than an inherent limit, by adding a mismatched-heading
  file at the post-strip path and watching the same mutation become detectable.

  The case's comment also claims the containment test "must never trip on it", which no realistic
  mutation of `<project>/` handling could ever falsify — a comment asserting coverage nothing tests.

- **Judged sound:** citing the case from `skills/openspec-explore/SKILL.md`, a declared expected-zero
  member. A non-declared-zero location produces an unrelated coverage-guard failure instead. The gap
  was the missing colliding target, not the citing file.

**Fix round dispatched to the same implementer**, model **sonnet**, folding into `eeb4235`.

**This is the second case in two tasks that passed for the wrong reason**, and both were found by
mutation rather than by reading. The pattern is worth carrying into the panel: in this change a test
asserting "nothing was reported" is nearly always weaker than it looks, because *not being seen at
all* produces the same observable as *being seen and judged fine*.

### Task 3 fix round — clean, verified by the parent rather than by a third review round

`7ba797f`. The fixture gained a colliding target at the post-strip path (`.myflow/project.md` under
the fixture root) carrying a heading that does not match the cited section, so a wrongly-stripped
`<project>/` now resolves to a real file and is reported.

The parent re-ran the mutation independently rather than accepting the implementer's account:
applying the `<project>/`-also-stripped mutation makes case 27 fail with
`no bold token resolves to a heading in <project>/.myflow/project.md`; restoring makes it pass, and
`git diff scripts/check-references.sh` is empty afterwards, so the production file is byte-identical
to the reviewed one. A third review round was not spent: the finding was test-only, and the fix's
evidence is self-verifying in a way a reading pass could not improve on.

The implementer also **reworded the case's comment** rather than leaving it claiming a guarantee
nothing tests — the containment half is now stated as an unfalsifiable sanity check.

**Task 3: cleared for its checkbox.**

---

## Task 4 — Wave A, the copied root files and the always-on rules

- **Implementer:** general-purpose, model: **sonnet**. Task base `7ba797f`. 41 sites, six files.
- Dispatched with the guard's own `path:line` report named as the authority over both this prompt and
  the plan's per-file counts, and with the per-site judgment made explicit — the same citation root
  (`openspec/`, `README.md`) resolves to this repository in one sentence and to the target project in
  the next, so a `sed` over a root map would be wrong in both directions.
- Carried the `<agents repo>`-contains-a-space warning forward: it must sit inside the same backtick
  span as its path, never as two spans, because that is what the critical defect was made of.

- **Commit:** `c7b939c`, 6 files, 38 insertions / 38 deletions. 236 → **195**, matching the repaired
  plan exactly, with none of the 195 remaining in the six files.
  `check-task-commit-fields.sh . 4 c7b939c 7ba797f` exit 0. `check-references.sh` clean with
  `CLAUDE.md` and `AGENTS.md` coverage unchanged at 6 each — no coverage silently dropped, which is
  the specific thing task 3 was ordered early to make observable.

### Parent-caught defect — a site silenced rather than rooted

The implementer flagged two edits as a "plain typo fix": adding `~/` to `` `.claude/skills/` ``
(`CLAUDE.md:53`) and `` `.codex/skills/` `` (`AGENTS.md:99`). Both are wrong, and wrong in the one
shape this change exists to prevent.

The sentence reads "These skills live in `skills/` next to this file (or in `.claude/skills/` if
installed there)" — **"next to this file" is the project root**, so the sentence describes a
per-project install. `setup.sh` confirms it: `install_claude_code` runs
`install_skills "$PROJECT_DIR/.claude/skills"` (setup.sh:203) and `install_codex` runs
`install_skills "$PROJECT_DIR/.codex/skills"` (setup.sh:220). The home-directory form at
setup.sh:766-775 belongs to `install_global`, a different mode the sentence is not describing. The
correct citations are `<project>/.claude/skills/` and `<project>/.codex/skills/`.

**The severity is not the two words.** `classify_token` treats any `~`-rooted token as *not a
citation at all*, so adding `~/` did not root these sites — it removed them from the guard's view.
The observable effect is a site that stopped being reported without being fixed, which is exactly
the outcome the whole change is built to make impossible. **A citation must leave the guard's report
by acquiring a root, never by ceasing to look like a citation.** The fix round carries that as a
general rule, and asks the implementer to re-read its own diff for any other edit of that shape — a
removed backtick span, a path moved into a fence, an added `~/`, `/` or `$`.

This is the third defect in this change found by asking "could this pass for the wrong reason?"
rather than by reading the change for correctness, and the first one where the *implementer's* fix,
not a test, was what would have passed for the wrong reason.

### Pre-existing defect recorded, deliberately not fixed here

`AGENTS.md:49` cites `specs/myflow-global-install/spec.md`. No such spec exists — not at that path
(this repository's specs live under `openspec/specs/`) and not under `openspec/specs/` either. It is
stale, and `check-references.sh` never caught it because the line carries no bold token adjacent to
the path, so it passed both before and after. The implementer resolved the shorthand to
`<agents repo>/openspec/specs/myflow-global-install/spec.md`, which is the right *form*; the missing
target is a pre-existing defect outside this change's scope and is left for a follow-up rather than
quietly fixed inside a citation-rewrite commit.

### Wave A fix round 1 — the silenced sites, corrected

`09d1922`. Both became `<project>/.claude/skills/` and `<project>/.codex/skills/`. The implementer
verified the correction the right way — against the guard's own `classify_token`/`judges_ok` rather
than against the violation count: the `~/` form returns `classify_token=False` (invisible, never
counted), the `<project>/` form returns `classify_token=True, judges_ok=True` (seen, and judged
rooted). That distinction is the whole point, and a count alone cannot show it.

Asked directly whether any other edit had the silencing shape, the implementer re-read the full diff
line by line and reported none.

### Per-task review on wave A — clean on substance, two minor prose findings

The reviewer was pointed at what the guard structurally *cannot* check: whether each site got the
**right** root. The count landing on 195 proves only that every site acquired *a* root.

- **Every site correct.** ~30 citations judged independently against their own sentences, then
  compared with the implementer's choice. No site the reviewer would have decided differently.
  `openspec/` and `docs/` appear under both roots in these files, so the first segment carried no
  information and each one had to be read in context.
- **No silencing shape**, independently re-audited rather than accepted. The only `~/`-rooted tokens
  left in the diff (`~/.codex/AGENTS.md`, `~/.claude/CLAUDE.md`) are genuine home-relative paths,
  unchanged, and correctly left bare.
- **The three `agents/` drops are correct** — there is no `agents/` subdirectory in this repository;
  `setup.sh` and the rules sit at the root, so the old informal clone-dir prefix was standing in for
  exactly what `<agents repo>/` now says.
- **Minor ×2.** `AGENTS.md:18` and `rules/agent-baseline.md:52` each pair a possessive with the
  prefix that now states the same thing — "a project … in its `<project>/.myflow/project.md`".
  Cosmetic, and the reviewer called them non-blocking; sent back anyway, because myflow's bar is zero
  open findings **at any severity** and a bar that bends for cosmetics is not a bar. The fix round
  asks the implementer to sweep its own six files for the same shape rather than fixing only the two
  the reviewer happened to name.

### Wave A fix round 2 — clean

`76ddb1e`. Both possessive doublings fixed, the first reworded to match the existing house style two
lines away (`` `<project>/.myflow/project.md` ``'s `## lint` section).

Asked to sweep its own six files for the same shape, the implementer found no further instances and
**named six candidates it had considered and deliberately left alone**, each with the reason the
possessive is doing separate work: "the project's own, plus `~/.codex/AGENTS.md` globally" contrasts
project against global; "the project's configured lint commands (see …)" attaches the possessive to
*commands*, not to the citation; "its own guard list" and "its own contract file" are
self-referential to the file rather than to a project. Discriminating between a redundant possessive
and a load-bearing one is the judgment this task wanted, and getting the negative cases explicitly
enumerated is better evidence than a clean sweep with no reasoning.

**Task 4: cleared for its checkbox.** Final: `76ddb1e`, 195 violations, every guard clean.

---

## Tasks 5 and 7 — Wave B, and the Guard resolution edit

- **Implementer:** general-purpose, model: **sonnet**. Task base `76ddb1e`. Two separate commits.
- Bundled by `plan-dispatch-bundles.sh` because both edit `pipeline.md`, at different sections. The
  plan deliberately keeps them as separate tasks: a reviewer could reasonably approve 121 mechanical
  rewrites and still reject a change to what a contract *says*.
- 121 sites across fourteen files — the largest wave, and the one where `openspec/` and `docs/`
  appear under both roots most densely, so the first path segment carries no information anywhere in
  it.
- The dispatch carries the two rules this change has already paid for in defects: a citation leaves
  the report by acquiring a root and never by ceasing to look like one; and `<agents repo>` contains
  a space, so it must sit inside the same backtick span as its path.
- It also asks for `check-references.sh`'s **per-member coverage compared before and after**, not
  just its verdict. A member whose count silently drops to zero is exactly the failure task 3 was
  ordered early to make visible, and the verdict line alone would not show it.

- **Commits:** `c92f2c8` (task 5, wave B) and `838859d` (task 7, the Guard resolution edit). Both
  `check-task-commit-fields.sh` runs exit 0.
- **195 → 78**, not the predicted 74. The gap is four sites the implementer **refused to force**, and
  refusing was the right call: each is a token the classifier calls a citation that is not one, and a
  root written onto any of them would be false.
- **`check-references.sh` per-member coverage compared before and after**, captured across a
  `git stash` to get the pre-edit baseline: every count identical. Zero coverage lost — the thing
  task 3 was sequenced early to make observable, actually observed rather than assumed.
- **No evasion shape used.** All four residual sites remain fully visible in the guard's report,
  exactly as written.

### Judgment calls the implementer surfaced rather than buried

- Git **branch-name** citations (`openspec/<name>` as a literal git ref, not a file path) at
  `finish-contract.md:169,177,178` and `state-file.md:140` took `<project>/` for consistency with the
  guard's blanket treatment. Flagged as the closest call in the set — a branch name is not strictly a
  filesystem citation, and the convention has nothing to say about refs.
- Bare `scripts/` fragments naming a *per-command* skill directory were resolved with the corpus's
  own existing multi-word placeholder idiom, `` `<the running command's own skill directory>/scripts/` ``,
  rather than by inventing a technique. Task 9's placeholder rule below retroactively makes that
  idiom principled rather than incidental.
- `state-file.md:220`: `"/myflow-do"` unquoted to `/myflow-do` — a slash-command name, not a path;
  the quotes were defeating the leading-`/` exclusion every other command mention relies on.

---

## Task 9 — three classifier exclusions the corpus proved necessary *(added mid-run)*

Wave B's four residual sites are classifier defects, not corpus defects, and they block task 8: a
guard reporting four permanent false positives can never be registered in `## lint`. The plan gained
a ninth task rather than absorbing them.

**A suppression marker was considered and rejected.** `check-references.sh` has the precedent
(`refs-guard:allow`), so the idiom exists here. But a marker silences a real hit and a false one
identically, and this repository's lint rule forbids introducing one — the correct reading is that
these are defects *in the classifier*, and a classifier defect is fixed in the classifier.

Three mechanical classes close all four, with **no fourth root and no exception list**:

1. A `../`-rooted token is relative to something the document does not name, so no correct root
   exists to write. Closes the two negative examples.
2. A token whose first segment is a `<…>` placeholder is **already rooted** — the placeholder is the
   root. Stated as a generalisation rather than a special case for `<state-dir>`, which makes
   `<agents repo>/` and `<project>/` simply the two placeholders this convention gives fixed
   meanings to. That is a coherent rule; an exception list is not.
3. A token carrying a leading or trailing quote character **inside its own backtick span** is quoted
   program output. Closes the verbatim git error message.

Closing all four restores the plan's original arithmetic exactly: 78 → 74 → 0.

- **Implementer:** the same agent that built the guard, model **sonnet**, task base `838859d`.
- Dispatched with the instruction that a **fifth** shape must be reported rather than swallowed by
  widening an exclusion — every exclusion is a hole in the guard, and a hole nobody can articulate is
  one an author will later drive a real citation through.

- **Commit:** `405792b`. 31 harness cases (was 27). `check-task-commit-fields.sh . 9 405792b 838859d`
  exit 0. All three exclusions implemented in `classify_token`/`judges_ok`, **no suppression marker**.
- **Mutation, all three load-bearing:** disabling `../` breaks only `parent-relative-path`; disabling
  the quote rule breaks only `quoted-program-output`; disabling the `judges_ok` placeholder branch
  breaks both `placeholder-rooted` and `placeholder-rooted-is-recognised` — two angles on one branch
  (not-reported, and recognised-via-coverage) rather than an unrelated rule catching it by accident.

### The count came in at 67, not 74, and the implementer refused to force it

This is the second planning-time estimate this change has had to repair against the guard's own
measurement, and it was handled the same way: **repair the baseline, never narrow the rule.**

My predicted 74 assumed only the four sites wave B named carried these shapes. Seven more already
did — three `<abs-worktree>/.superpowers/sdd/dispatch-context.md` repeats (the CONTEXT BUNDLE
paragraph, verbatim 3×), `<changeRoot>/specs/`, `<changeRoot>/tasks.md`, and two further `../`
negative examples (`../../../etc/passwd`, a second `../<other-app>`). All eleven were traced
individually and every one is a genuine instance of the three classes.

Narrowing the rules to hit 74 would have meant re-introducing exactly the `<state-dir>` special-case
the task rejected — trading a coherent rule for an exception list, to defend a number that was only
ever an estimate. The implementer flagged it for a decision rather than guessing, which is what the
`unverified:` provenance tag on that step existed to provoke.

Parent verification before accepting: 67 confirmed, spread across exactly wave C's twelve files, and
**no `<…>`-rooted or `../`-rooted token remains anywhere in the report** — the exclusions leave no
residue that would suggest they are over-broad.

### Per-task review dispatched — pointed at the one risk that matters

Task 9 is the riskiest commit on the branch: **every exclusion is a hole in the guard**, and the
guard is about to become this repository's lint. A hole too wide silently swallows real violations.
The review is asked to try to *smuggle* a genuinely unrooted citation through each new exclusion — a
fake placeholder like `<foo>/openspec/specs/x.md`, a `../`-prefixed path really naming a repo file, a
quoted citation that is not program output — and to independently re-verify all eleven dropped sites,
since any one of them that is really a citation is a violation the guard has stopped reporting
without it being fixed.

### Per-task review on task 9 — one critical finding: the guard failed open

The review was pointed at the single risk that matters for this commit — every exclusion is a hole,
and the guard is about to become this repository's lint — and asked to *smuggle* real violations
through each new hole rather than to read the code for correctness. It got four through.

- **Critical.** The placeholder rule accepted any first segment starting `<` and ending `>`, so
  `<foo>/openspec/specs/x.md`, `<>/openspec/specs/x.md` and `<a><b>/openspec/specs/x.md` all passed
  while naming unrooted paths — as did a plausible **typo** of a real placeholder, `<changeroot>/`
  or `<change-root>/`. **A guard that fails open is worse than no guard: it reports clean while
  checking nothing.** Same principle that made wave A's `~/` edits a defect.
- **Also critical, and the same defect seen from the contract side.** `specs/myflow-citation-roots/spec.md`
  said "exactly three roots are recognised" while the code recognised an unbounded fourth category,
  and no task amended that spec. A contract disagreeing with its enforcement is what `pipeline.md`
  calls non-determinism in the one layer that must be deterministic — and task 7, in this very
  change, quotes that framing.
- **Important.** The reported mutation count was wrong: disabling the placeholder branch breaks
  **six** cases, not two. The design claim survived — all six exercise one branch from different
  angles, which is the expected consequence of collapsing an exception list into a general rule, not
  shadowing — but the true number is what would have shown how much test surface had come to rest on
  one unbounded, spec-uncovered branch.

### The fix: a closed set, and the spec amended to match

`3fe4243`. `PLACEHOLDER_ROOTS` is now a five-member frozenset tested by exact membership —
`<agents repo>`, `<project>`, `<abs-worktree>`, `<changeRoot>`, `<state-dir>` — and anything else in
bracket shape is a violation. The spec was amended **by the parent**, not the implementer, since a
contract change is the planner's call: it now carries the closed table, the fail-open reasoning, and
two new scenarios (`An unrecognised placeholder root is a violation`, `A recognised placeholder root
is rooted`). The `../` and quote exclusions were added to the classifier requirement's `SHALL NOT`
list with their bounds stated — `./` and bare `..` deliberately not excluded.

**Two holes accepted deliberately, and recorded as accepted rather than left undocumented.**
`` `../openspec/specs/x.md` `` and `` `"openspec/specs/x.md"` `` remain unreportable. Neither is a
shape an author produces by accident when meaning to cite a repository file. The placeholder hole
differed in kind precisely because a plausible *typo* landed in it, which is an ordinary mistake
rather than a contrivance.

**Parent verification, run against the guard's own module rather than through a review:** all five
smuggle attempts now REPORTED, all five recognised placeholders ROOTED, including the case-typo
`<changeroot>` failing closed. Corpus unchanged at 67, violation list byte-identical before and
after narrowing, and **no placeholder outside the five exists anywhere in the corpus** — so no sixth
root is owed. A third review round was not spent: the evidence is direct and a reading pass could
not improve on it.

**Task 9: cleared for its checkbox.**

---

## Task 6 — Wave C, the command skills and `commands/`

- **Implementer:** general-purpose, model: **sonnet**. Task base `3fe4243`. 67 sites, twelve files.
- The dispatch now carries all three earned rules, and states plainly that `~/`, a leading `/`, `$`,
  `../`, a quote wrapper, a split backtick span and a fence move **all** suppress a site silently —
  which after task 9 makes silent suppression the easiest way to produce a wrong result that looks
  right, and therefore the thing the reviewer must hunt for.
- `skills/myflow-do/SKILL.md` takes 30 of the 67 against 549 bytes of budget headroom, so the budget
  guard is run immediately after that file and before the other eleven, with the one-row raise
  pre-authorised.

### Per-task review on wave C — the guard reached zero and three classes of site were still wrong

`c849b58` drove the corpus to **zero violations** with every lint guard green. The review found three
classes of site that satisfy the guard and are semantically wrong, none of them detectable by any
guard. **A syntactically valid root is not a correct one.**

The reviewer also **disproved both fixes the parent had proposed** for the branch-name defect, with
evidence rather than preference — the kind of correction a review is for:

- *Write them as full refs* (`refs/heads/openspec/<name>`) is **actively wrong**, not merely verbose:
  `git branch "refs/heads/openspec/foo"` creates a branch literally named `refs/heads/openspec/foo`.
  It also contradicts the repository's own usage — `finish-contract.md` runs
  `git branch -d "openspec/<name>"` bare, and `state-file.md`'s JSON schema stores the bare form.
- *A shape rule* ("no extension, no trailing slash → not a citation") would silently stop checking
  six real corpus citations, including `` `<agents repo>/rules/<name>` ``, and would open a permanent
  blind spot for any future bare directory-shaped citation — the fail-open outcome task 9 exists to
  prevent.

Its own recommendation — an exact-literal exclusion, extending the precedent `GIT_BRANCH_RE` already
sets for `origin/main` — was taken.

### The three classes, and why each was wrong

1. **A git branch name given a filesystem root**, 5 sites. `skills/myflow-do/SKILL.md:115` read
   "Branch `<project>/openspec/<name>`" — an **instruction**, which an agent would follow literally.
2. **`.superpowers/sdd/` rooted at `<project>/` when it is worktree-scoped**, 18 sites across seven
   files. The corpus contained its own disproof in four places: the **Temporary artifacts registry**
   places these artifacts in the worktree; `myflow-do/SKILL.md` already cited
   `<abs-worktree>/.superpowers/sdd/dispatch-context.md` at three sites, so one file used both roots
   for one directory; `SKILL.md:429` read "against `<project>/…/final-review.diff` **in the
   worktree**", contradicting itself inside one sentence; and **this change's own delta spec** used
   `<abs-worktree>/.superpowers/sdd/final-review.diff` as its worked example of a recognised
   placeholder root. The change disagreed with itself.

   The two reviewer-prompt sites were the highest-stakes: those files ship **verbatim** as subagent
   prompts, and `principles-reviewer-prompt.md` states the subagent's working directory *is* the
   worktree, so a `<project>/`-rooted `[DIFF_PATH]` sends a fresh reviewer to a path that may not
   exist where it stands.
3. **A fabricated `file:line` illustration rooted**, 1 site — teaching a format real findings never
   use, since a findings table's Location column comes verbatim from `git diff` and is diff-relative.

**Judged clean, no action:** the `` `N/M` `` → `` `N` of `M` `` reword. It was never a path citation,
it is a single non-recurring notational collision, and a permanent classifier rule for one site is
the wrong trade — unlike task 9's three exclusions, each of which recurred across multiple sites and
earned its surface.

## Task 10 — cross-wave root corrections *(added mid-run)*

`8944113`, 24 sites across ten files, guard still at zero. Added `GIT_BRANCH_LITERALS` (exact-literal)
and `FILE_LINE_RE` (`:<digits>$`) exclusions, plus three harness cases — 35 in total — each
mutation-proven to fail alone. The rejected alternatives are preserved in the classifier's own
docstring, so the next reader meets the reasoning rather than re-deriving it.

Corrected as a **new task rather than fixups**, because the defect spanned waves B and C and the
classifier: three commits would have needed rewriting to express one coherent decision.

The parent also corrected wave C's own guidance table in `tasks.md`, which still taught the
`<project>/` root task 10 had just overturned — a plan that keeps teaching a superseded rule is a
plan that will be followed again.

---

## Task 8 — register the guard in this repository's lint

- **Implementer:** general-purpose, model: **sonnet**. Task base `8944113`. The last task.
- This is what makes the change self-sustaining: without it the guard exists and nothing runs it, and
  the 236 rooted citations would start accumulating again from the next edit to any installed file.
- Dispatched with an explicit instruction to **measure** this guard's runtime rather than copy the
  Baseline's 0.588s, which timed the `global` mode alone — this guard runs the installer **twice**,
  and it is the only lint entry that shells out to `setup.sh` at all, which the section's prose
  should say.

---

## Close

**10 tasks, 10 commits, panel clean at zero open findings.** 236 citations rooted across three waves;
the guard runs in this repository's lint and reports zero, deterministically (20 consecutive `rc=0`
here, 30 in an isolated checkout, 45 earlier across five `PYTHONHASHSEED` values).

**Full project test suite green — 28 harnesses**, not only the two this change adds.

### What this run should be remembered for

**Every defect that mattered was found by asking "could this pass for the wrong reason?", never by
reading the code for correctness.** The count of them is the point:

- the tokenizer discarded every `<agents repo>/…` citation, and a green test proved nothing;
- two `~/` edits made sites vanish from the report without rooting them — the target count was hit
  with every guard green;
- the placeholder rule accepted any `<…>` shape, so `<foo>/openspec/specs/x.md` passed;
- a git branch name took a filesystem root, satisfying the guard while producing an instruction an
  agent would follow literally;
- three more bypasses — `origin/README.md`, `see .myflow/project.md`, `.myflow/project.md:42` — were
  invisible rather than judged clean;
- a fix was **reported as landed and was never made**, caught only by diffing the file.

In every one of those the visible signal was correct and the underlying work was wrong. That is the
same defect class the change itself exists to fix, recurring inside the process building the fix.

### What the guard cannot do, stated so nobody assumes otherwise

**Reaching zero proves every citation acquired *a* root. It cannot prove any acquired the *right*
one.** Task 10 corrected 24 sites that satisfied the guard and were semantically wrong, including two
reviewer prompts that ship verbatim as subagent prompts and pointed a fresh reviewer at a path that
may not exist where it stands. Only reading each sentence finds those.

### Every hole is written down

`specs/myflow-citation-roots/spec.md` documents the `file:line` exclusion, the `origin/README`
residue, and the mid-span comment edge. An undocumented hole is one an author eventually drives a
real citation through — this change was bitten three times by a rule that existed only in code.

### Follow-ups, deliberately not fixed here

1. **`AGENTS.md` cites `openspec/specs/myflow-global-install/spec.md`, which does not exist** — not at
   that path and not under `openspec/specs/`. Pre-existing; `check-references.sh` never caught it
   because the line carries no bold token adjacent to the path.
2. **`scripts/check-guard-symlinks.sh:122` carries the same unprotected `mktemp -d`** this change
   fixed in its own guard: a failure exits 1, which is that guard's "violations found" code.
3. **`/myflow-fast`'s skill directory has no `engineering-principles.md`**, so `[PRINCIPLES_PATH]`
   cannot resolve there. A real bug, found when an agent told to review only wrote a fix into this
   worktree. Preserved as `out-of-scope-principles-path.patch` in the session scratchpad and reverted
   — a citation-rooting commit is not where that belongs.
