# KAN-77 — The SDD ledger gets one canonical path

**Date:** 2026-08-21 · **Jira:** KAN-77 · **Change:** `kan-77-sdd-ledger-canonical-path`

## The problem

The SDD ledger has no path any component agrees on:

| Component | Says |
|-----------|------|
| `skills/myflow-do/SKILL.md` | names `dispatch-context.md`, `final-review-panel.md`, `final-review.diff`, `fix-round-N.diff` flat under `.superpowers/sdd/` — and **never names the ledger** |
| `scripts/preserve-session-records.sh:227` | reads only `<worktree>/.superpowers/sdd/tasks/progress.md` |
| `pipeline.md` Temporary artifacts registry | names the *directory* `.superpowers/sdd/` — settles nothing |

On disagreement `preserve()` prints `skipped: <src> (absent)`, exit 0 — a legitimate outcome per the
table in **Preserving the session records** (`pipeline.md`), since a change may truly have no such
record. So *never written* and *written elsewhere* are indistinguishable; the ledger dies with the
worktree at run 2.

Why it matters: **Model policy** (`pipeline.md`) requires every dispatch to record its model and
names the preserved ledger as what makes that answerable after archiving. Nobody keeps the
transcript.

## What is NOT the problem

Both verified against the current tree:

- **`gather-self-review-context.sh` already searches the dated shape.** `find_dated()` (line 411)
  globs `[0-9]{4}-[0-9]{2}-[0-9]{2}-${NAME}${suffix}` — exactly what preservation writes.
  `docs/superpowers/ledgers/<name>.md` is only `LEDGER_LABEL` (line 402), a display string; the
  script's header says so. This change does not touch it.
- **The `tasks/` component is not a mistake.** `sdd-workspace` prints
  `<repo-root>/.superpowers/sdd/<plan-basename>/`; myflow's plan is always `tasks.md`; and
  `skills/myflow-do/SKILL.md:214` invokes that skill. Preservation reads the path upstream actually
  produces. myflow-do's silence is what lets a run produce a different one.

**Out of scope:** KAN-155 — that nothing *writes* the ledger. This change names the path and asserts
presence; it does not specify the writing.

## Decisions

### The canonical path is upstream's, not a new flat one

**ID:** canonical-path-is-upstream
**Status:** active
**Chosen:** `<abs-worktree>/.superpowers/sdd/tasks/progress.md` — what `sdd-workspace tasks.md`
already creates and what preservation already reads, so exactly one ledger exists and neither script
moves.
**Considered:**
- *Flat `.superpowers/sdd/ledger.md`*, matching every other myflow artifact and most observed runs —
  ruled out: upstream still creates `tasks/progress.md`, leaving two ledgers and a controller needing
  to be told which is real, the same ambiguity this removes.
  **Stronger than that**: upstream's `subagent-driven-development/SKILL.md:141-146` calls
  `.superpowers/sdd/progress.md` "the old flat path" and instructs a resuming controller to treat a
  ledger found there as *another plan's* and start fresh. A flat ledger is therefore not merely
  misfiled — a resumed run disowns it.
- *Accept both, canonical `tasks/`* — ruled out: two legal paths forever is the tolerance being
  ended, and the rescue covers the transition without blessing a second path.

### A misfiled record is rescued, not merely reported

**ID:** rescue-not-report
**Status:** active
**Chosen:** on canonical-absent, walk an ordered allowlist of known-wrong paths, first hit wins, copy
to the canonical destination, name every path tried.
**Considered:** *report only, operator moves it by hand* — ruled out: the failure being fixed is that
nobody reads the skip line in time, and a report needing a human preserves that dependency.
*Name the paths tried but keep the skip* — same reason: honest message, record still lost.

### The rescue matches by name, from an ordered list

**ID:** allowlist-not-glob
**Status:** active
**Chosen:** fixed, ordered literal relative paths per source; first existing wins.
**Considered:**
- *Glob minus a denylist of myflow's own artifacts* — the denylist must track myflow-do's artifact
  list, and the two drift the moment one is added, reproducing this defect class elsewhere.
- *Glob plus a first-line header check* — a ledger written without the header is exactly the sloppy
  run being rescued, so the check fails in its only case.

### Two new outcomes, both exit 0

**ID:** outcomes-exit-zero
**Status:** active
**Chosen:** `rescued:` and `MISSING:` both exit 0; non-zero keeps its one meaning — a copy attempted
and refused or failed.
**Considered:** *MISSING exits non-zero* — ruled out: the contract defines non-zero as an
attempted-and-failed copy, and a source that was never there is not that. Visibility comes from a
distinct word instead.

### Rescue and MISSING cover two sources, not three

**ID:** two-sources-not-three
**Status:** active
**Chosen:** ledger and panel record; the proposal artifact keeps the plain skip.
**Considered:** *all three* — `/myflow-fast` publishes no proposal artifact, so MISSING would fire on
every fast run. An alarm that always fires is one nobody reads.

### The panel record's allowlist is seeded

**ID:** panel-allowlist-seeded
**Status:** active
**Chosen:** `panel.md`, `review-panel.md`, `final-review.md` — catching a misfile the first time
rather than after.
**Considered:** *empty list, MISSING only, until a real misfile is observed* — the
simplest-thing reading, since no misfiled panel record exists here. Overruled by the operator: the
ledger's own history shows the cost of learning the wrong path only after a loss.

### The writer asserts presence but does not specify writing

**ID:** assert-not-write
**Status:** active
**Chosen:** myflow-do names the path at its upstream-skill invocation, and § 7 adds a non-gating
`test -f`.
**Considered:** *name the path only, leave presence to KAN-155* — a path nobody checks is one a run
can still silently skip; the assert costs one line.

## Open questions

*(none)*

## Design

### 1. The canonical path, in three places

| Component | Change |
|-----------|--------|
| `skills/myflow-do/SKILL.md` § 4 | names `<abs-worktree>/.superpowers/sdd/tasks/progress.md` at the point it invokes `superpowers:subagent-driven-development` |
| `pipeline.md` registry | the `SDD ledger` row's Location becomes the filename, not the directory |
| `scripts/preserve-session-records.sh` | path unchanged; its source-list comment gains a citation to the registry row |

### 2. Rescue in `preserve()`

Fifth parameter: a newline-separated, ordered list of worktree-relative fallbacks, empty for the
proposal artifact. Replaces only the leading `[ ! -f "$src" ]` early return:

1. canonical present → unchanged, `preserved:`
2. canonical absent, fallback exists → that file passes **every existing protection unchanged** and
   is copied to the canonical destination → `rescued: <dest> (found at <path>)`
3. canonical absent, none found, list non-empty → `MISSING: <canonical> — tried <paths>`, return 0
4. canonical absent, list empty → `skipped: <src> (absent)`, unchanged

| Source | Canonical | Fallbacks, in order |
|--------|-----------|---------------------|
| SDD ledger | `.superpowers/sdd/tasks/progress.md` | `.superpowers/sdd/ledger.md`, `.superpowers/sdd/progress.md`, `.superpowers/sdd/tasks/ledger.md` |
| Panel record | `.superpowers/sdd/final-review-panel.md` | `.superpowers/sdd/panel.md`, `.superpowers/sdd/review-panel.md`, `.superpowers/sdd/final-review.md` |
| Proposal artifact | `<state-dir>/<name>-proposal-artifact.html` | *(none)* |

Ledger order from KAN-77's occurrence list: `ledger.md` for KAN-15, KAN-58, KAN-73, KAN-82, KAN-129,
KAN-131; `progress.md` for KAN-110. `tasks/ledger.md` is the remaining combination.

A rescued file is a source like any other — the three header protections apply in full, and a
fallback resolving outside the worktree is a **refusal** (stderr, non-zero), never a silent skip.
Existing rule, new source; not a new rule.

### 3. The outcome table

**Preserving the session records** (`pipeline.md`) gains two rows:

| Outcome | What it means | What you do |
|---------|---------------|-------------|
| `rescued: <dest> (found at <path>)`, exit 0 | written to a non-canonical path, now copied to the canonical destination | **Report it.** The record is safe; the writer is not. Proceed. |
| `MISSING: <canonical> — tried <paths>`, exit 0 | a record that should exist for every change was at none of its known paths | **Report it, naming the paths tried.** Proceed. |

The paragraph below the table gains: neither new outcome is non-zero, so a caller branching on exit
status cannot read them as failures.

`finish-contract.md` is deliberately not edited — it cites this table rather than restating it,
which is what lets the row be added in one place.

### 4. The writer-side assert

`skills/myflow-do/SKILL.md` § 7, beside the existing verification:

```bash unverified:confirm the exact surrounding wording in section 7 when the file is open
test -f <abs-worktree>/.superpowers/sdd/tasks/progress.md
```

Non-gating, reported at its call site — same shape and reasoning as § 4's dispatch-context check. **No
new handoff field**: the block is shared with `/myflow-fast` and `/myflow-finish`, so a field added
for one changes all three.

### 5. The two misnamed preserved ledgers

`2026-08-20-kan-102-citations-resolve-to-installed-paths-ledger.md` and
`2026-08-08-kan-95-slim-the-myflow-contract-files-per-move-ledgers.md` carry a hand-typed suffix
`find_dated()` cannot match. Each first line is read to confirm its change before `git mv`. The
kan-95 file sits beside `2026-08-08-kan-95-slim-the-myflow-contract-files.md`, so it may be a
genuinely separate record — if its first line says so, leave it and record why.

**Resolved.** The kan-102 file's first line reads `# SDD ledger — kan-102-…`: a real ledger with a
typo'd suffix, renamed. The kan-95 file's reads `# Per-move ledgers — kan-95-…`: it is the per-move
ledger **A move or eviction is recorded in a per-move ledger** (`myflow-contract-economy`) requires,
not an SDD ledger, merely filed in the same directory. Its change's real SDD ledger already sits
beside it under the canonical name, so renaming would have destroyed one of the two. Left in place.

**This task produces no task commit.** `docs/superpowers/` is a planning path — pipeline.md's Git
boundaries and section 4's COMMIT-PER-TASK clause both forbid a task commit from touching it. The
rename is left in the working tree and lands in finish run 1's `chore(openspec): plan and session
records` commit.

### 6. Error handling

Nothing here can block an integration: every new outcome exits 0, `RC` accumulation is untouched, and
remaining sources are still attempted after any failure — the guarantee **Preserving the session
records** (`pipeline.md`) already states. The § 7 `test -f` reports and continues.

### 7. Testing

`scripts/test-preserve-session-records.sh` gains cases. The harness header's rule holds: assert
against the delta spec, never against observed output. They are written **before** the script changes
and confirmed to fail against the unmodified script, so none can pass vacuously:

- canonical present → `preserved:`, no rescue
- each ledger fallback in isolation → `rescued:` naming that path, content actually arrives
- two fallbacks present → the earlier one wins, deterministically
- no ledger anywhere → `MISSING:` naming every path tried, exit 0
- the same set for the panel record
- proposal artifact absent → `skipped:` unchanged, no `MISSING:`
- a fallback symlinked out of the worktree → refused on stderr, non-zero, other sources still preserved
- idempotency: a rescue overwrites an existing dated file rather than adding a second

`check-references.sh` must stay green (the registry row and table rows add citations).
`check-contract-budget.sh` needs no table edit: `pipeline.md` is 47,565 B against 55,728 and
`myflow-do/SKILL.md` 71,528 against 89,566.
<!-- measured: wc -c on both files; scripts/check-contract-budget.sh budgets() @ branch main -->
