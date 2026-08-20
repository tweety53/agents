# Final review panel — kan-102-citations-resolve-to-installed-paths

**Roster:** `light` (recorded). Required slots: Primary · Principles · Code review (low).
**Panel model:** sonnet (recorded `models.reviewPanel`).
**Diff:** `git diff 332e593` — 1968 lines, under the cap (`check-panel-diff-size.sh` exit 0, no
operator prompt needed).
**Optional slots:** none — no trigger fired. The change adds no dependency, touches no auth, no
network boundary and no user data; it is one lint-time guard plus a documentation rewrite.


## Findings

| ID | Slot | Severity | Site | Finding | Resolution |
|----|------|----------|------|---------|------------|
| F1 | Primary | Important | `check-installed-citations.py` | `origin` excluded on the first segment alone, so `` `origin/README.md` `` passed unseen | narrowed: remainder must carry no file extension |
| F2 | Primary | Important | `check-installed-citations.py` | a multi-word backtick span dropped every word but the first, so `` `see .myflow/project.md` `` was never seen | every word is now classified |
| F3 | Primary | Important | `check-installed-citations.py` | `FILE_LINE_RE` excludes any `path:line`, including a real unrooted one | narrowing forces a fabricated findings-table example to be re-rooted; kept and **documented in the spec as an accepted hole** |
| F4 | Primary | Important | `pipeline-rationale.md:330` | `<project>/.superpowers/` on a sentence that says "in a worktree" — task 10 missed this site of its own defect class | corrected to `<abs-worktree>/`; first fix was **reported and never made**, caught by diff |
| F5 | Primary | Important | `specs/myflow-citation-roots/spec.md` | `GIT_BRANCH_LITERALS` and `FILE_LINE_RE` implemented with no requirement covering them | spec amended to cover every exclusion |
| F6 | Primary | Minor | `.myflow/project.md:110` | prose named a `project` mode; `setup.sh` has no such mode | corrected to `all` |
| F7 | Code quality | Minor | `check-installed-citations.sh:53` | unprotected `mktemp -d` exits 1 — the guard's *"violations found"* code — on an environment failure | refuses with exit 2, stdout empty |
| F8 | Code quality | Minor | `test-check-installed-citations.sh` | two hardcoded `/tmp` paths where siblings honour `${TMPDIR:-/tmp}` | both use `${TMPDIR:-/tmp}` and are registered for cleanup |
| F9 | Code quality | Important | `check-installed-citations.py` | the generalised bracket merge was greedy, turning `` `some-cmd < scripts/input.txt > output.log` `` into a garbled citation | merge rejected when an intervening word carries `/` |
| F10 | Primary | Important | `check-installed-citations.py` | `origin/rules/` excluded as a "ref" though it is a directory | `origin` now also requires no trailing `/` |
| F11 | Primary | Important | `check-installed-citations.py` | the colon rule dropped `` `.myflow/project.md:` ``, a real citation | fires only on a **non-final** segment |
| F12 | Primary | Minor | `check-installed-citations.py` | a mid-span HTML comment orphans a slash-bearing fragment, which over-reports | fails closed, absent from the corpus; **documented in the spec as accepted residue** |
| F13 | Code quality | Important | `check-installed-citations.py` | the comment-span skip was bound to the literal first four characters, so a leading space defeated it | `span.lstrip().startswith("<!--")` |

finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 withdrawn documented in specs/myflow-citation-roots/spec.md as an accepted hole; the narrowing was tried and forces a fabricated findings-table example to be re-rooted
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 withdrawn documented in specs/myflow-citation-roots/spec.md as accepted residue; it fails closed by over-reporting and occurs nowhere in the corpus
finding-status: F13 fixed
findings-total: 13

**Zero open.** Eleven fixed, two withdrawn — each withdrawal written into
`specs/myflow-citation-roots/spec.md` rather than merely decided here, because an undocumented hole is
one an author eventually drives a real citation through.

## Rounds

### Slot 2 — Principles (Merged lens), model sonnet — **CLEAN, zero findings**

Verified by running, not by re-reading: the guard (exit 0, zero violations), the harness, both
`check-references.sh` and its harness including the two new prefix-stripping cases,
`check-contract-budget.sh`, and a grep for new suppression markers or lint-config weakening (none).

Ruled on the five questions it was asked to judge rather than defend:

1. **The nine exclusions are a coherent classifier, not an accreted pile.** Each is a distinct named
   shape carrying a doc comment stating *why*, and each is mutation-proven individually. It singled
   out the deliberate exclusion of `$` from `REGEX_GLOB_CHARS` — kept out precisely so the
   shell-variable rule stays independently testable — as evidence of a designed classifier rather
   than an accumulated one.
2. **Keeping the space in `<agents repo>` was right.** The merge is narrowly scoped, case-sensitive,
   isolated to one function and covered by dedicated cases; minting new vocabulary while rewriting
   236 sites would have compounded risk inside one change.
3. **The closed five-root set is the right shape.** The rejected alternative fails open, and a guard
   reporting clean while checking nothing is the single failure mode this change exists to prevent.
4. **Tasks 9 and 10 are legitimate SDD discovery, not a design gap.** "Acquired *a* root" and
   "acquired the *right* root" are different questions, and no purely syntactic classifier can answer
   the second — that gap is inherent, and it was caught by review rather than shipped.
5. **Nothing over- or under-built.** Every non-trivial mechanism is justified by measurement or by an
   in-repo precedent: the sandboxed installer run is the literal requirement, the Python classifier
   follows `check-plan-provenance.py` and the 723-candidate measurement, coverage reuses
   `scripts/lib/coverage.sh`, and the harness follows `test-check-guard-symlinks.sh`'s idiom.

**One inaccuracy in its report, checked and dismissed:** it counted 29 harness cases. The harness
emits **35** `ok:` assertions and all pass. No defect — a miscount in the reviewer's prose, not in the
code.

### Slot 3 — Code review (low), model sonnet — **2 minor findings**

Ran the harness, both guards against the real tree, and hand-built reproducers against all four
scrutiny areas.

- **Minor — `scripts/check-installed-citations.sh:53`.** `WORK="$(mktemp -d …)"` is unprotected under
  `set -euo pipefail`, so a `mktemp` failure kills the script with mktemp's own status **1** — this
  guard's *"violations found"* code — with nothing on stdout. An environment failure is then
  indistinguishable from an empty violation report, which contradicts the guard's own contract that
  **2** means "cannot answer". **Reproduced by the parent:**
  `TMPDIR=/nonexistent-dir-xyz bash scripts/check-installed-citations.sh` → exit 1, stdout 0 bytes.

  The reviewer noted this is **inherited, not introduced** — `scripts/check-guard-symlinks.sh:122`
  carries the identical unprotected pattern. Precedent does not make it right, and this guard is the
  one being added, so it is fixed here. Whether the pre-existing guard should be fixed too is a
  follow-up, not this change's scope.

- **Minor — `scripts/test-check-installed-citations.sh:454,458`.** The empty-root case hardcodes
  `/tmp/…` where every sibling path in the same file honours `${TMPDIR:-/tmp}`. Only bites in a
  sandbox where `/tmp` is not writable, but it is an inconsistency inside one file.

**Clean on all four scrutiny areas, each probed rather than read:**

- **The sandbox** — `HOME` and the project dir are always subdirectories of a fresh `mkdtemp()`,
  gated by `ensure_within_sandbox` before every `subprocess.run`; `setup.sh` writes only under those;
  `shutil.rmtree` sits in a `finally` and survives any exception mid-install. No escape path found,
  and the refusal is proven to fire with **zero** `subprocess.run` calls.
- **Silent failure** — every refusal path exits 2 with an empty stdout, including non-UTF-8 input,
  reproduced directly. The wrapper escalates any Python exit outside `{0,2}` to 2, so even an
  uncaught exception cannot read as clean.
- **Shell correctness** — grep's `rc >= 2` split correct, `--` before paths, arrays not strings for
  space-safe paths.
- **Python correctness** — token slicing, the fence state machine and every exclusion validated
  against the suite and the live 56-file corpus; the sole encoding assumption fails closed.

The `check-references.sh` prefix strip is confirmed correct: the original `$path` is preserved for
error messages, only `resolve_path` is stripped, and `<project>/` is deliberately untouched.

### Slot 1 — Primary (plan alignment + code quality), model sonnet

**Round 1 — 5 findings.** Three were live classifier bypasses, found by *hunting* rather than
reading: `` `origin/README.md` ``, `` `see .myflow/project.md` `` and `` `.myflow/project.md:42` ``
all passed without ever being counted as checked — invisible rather than judged clean, the same
recognised-versus-never-seen collapse that hid this change's original critical defect. Plus
`pipeline-rationale.md:330`, a live instance of task 10's own defect class that task 10 missed, and
two exclusions implemented with no requirement covering them — a divergence this change had already
made once.

**Round 2 — finding 4 was reported fixed and was not.** The implementer's report said
`pipeline-rationale.md:330` had been corrected; `git diff --stat` showed the file untouched. The
reviewer caught it by diffing rather than believing either the implementer or the parent. **A report
that says a change landed when it did not is worse than one that says it was skipped, because it
stops anyone looking.** Two further dormant holes found on renewed hunting: `origin/rules/` excluded
as a "ref" though it is a directory, and `.myflow/project.md:` dropped because the colon rule did not
distinguish a final segment from a non-final one.

**Round 3 — CLEAN.** Every finding verified closed by diff and by independently re-running both
mutations. One dormant edge remained — a mid-span HTML comment orphaning a slash-bearing fragment —
which **fails closed** (over-reports) and occurs nowhere in the corpus. Recorded in the spec as
accepted residue rather than left undocumented.

### Slot 3 — Code review (low), round 2 and 3

**Round 2 — 1 new finding**, in exactly the function flagged for re-scrutiny: the generalised bracket
merge was greedy, turning `` `some-cmd < scripts/input.txt > output.log` `` into a citation called
`< scripts/input.txt >`. Unlike every other defect in this change, this one **over-reports** rather
than blinding a signal — still a defect, because a guard that emits unreadable diagnostics on
ordinary prose is one people stop reading.

**Round 3 — 1 new finding.** The comment-span skip was bound to the literal first four characters, so
a span with leading whitespace fell through. It closed the shape the corpus happens to use rather
than the class — the same narrowness this change had already been bitten by three times.

### The parent's own two errors, both caught by this panel

1. **The bound I specified broke a rule I had approved two rounds earlier.** Rejecting a merge when
   an intervening word carries `/` kills the garbled redirection token — and also kills
   `<!-- measured: ./gradlew … -->`. The implementer implemented it exactly as stated, hit the
   conflict, and **reported instead of special-casing around it**. The fix was not a better bound but
   moving the comment rule to the span level, where being a comment is a property of the span rather
   than of any word inside it.
2. **I claimed `scripts/` was an installed root. It is not** — `setup.sh` installs `skills/`,
   `rules/`, `commands/`, `commands-claude/` and `hooks/`, never a bare `scripts/`. Corrected by the
   implementer and passed on to both reviewers so neither inherited the error.

### The determinism scare, and what it actually was

The code-quality reviewer reported that its **first** run showed violations at two real corpus sites
while 41 later runs were clean, could find no mechanism, and honestly declined to file it. A guard
that answers differently on identical input would be worse than every defect this change fixed, so it
was checked rather than assumed.

**The guard is deterministic** — 20 consecutive `rc=0` runs on a clean tree here, 30 in an isolated
detached checkout, on top of 45 earlier identical results across five `PYTHONHASHSEED` values. The
variance was **the scanned tree changing between invocations**: an agent that had been told to review
only had written into this worktree — three symlinks under `skills/myflow-fast/`, edits to
`skills/myflow-do/SKILL.md` and `SKILL-rationale.md`, and a `declare_if_present` line added to
`scripts/check-installed-citations.sh` itself.

Those edits fix a **real** bug — `/myflow-fast`'s skill directory carries no
`engineering-principles.md`, so `[PRINCIPLES_PATH]` cannot resolve there — but they are **not this
change's scope**, and a citation-rooting commit is not where they belong. They were preserved to
`out-of-scope-principles-path.patch` in the session scratchpad, the worktree was restored, and the
guard verified still at zero afterwards. Recorded as a follow-up.

## Verdict — panel clean

Three required slots under the `light` roster, all reporting clean at close. **Zero open findings at
any severity.** Every hole the classifier has is written down in
`specs/myflow-citation-roots/spec.md`: the `file:line` exclusion, the `origin/README` residue, and
the mid-span comment edge. That was deliberate — an undocumented hole is one an author eventually
drives a real citation through, and this change was bitten three times by a rule that existed only in
code.
