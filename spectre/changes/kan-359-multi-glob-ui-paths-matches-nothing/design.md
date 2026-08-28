# A multi-glob `ui paths` matches nothing — design

**Jira:** [KAN-359](https://tweety53.atlassian.net/browse/KAN-359)
**Change:** `kan-359-multi-glob-ui-paths-matches-nothing`
**Date:** 2026-08-28

## The defect, reproduced

```text verified:run against scripts/check-visual-trigger.sh at 182ad31, three fixtures built under TMPDIR
one glob   `src/gateway/src/main/**`              -> matching path exits 0 (MATCH)
two globs  `a/**`, `src/gateway/src/main/**`      -> same path exits 1 (NO-MATCH)
two globs  a/**, src/gateway/src/main/**          -> same path exits 0 (MATCH)
```

The third line isolates it: identical globs, no backticks, and matching works.

**Root cause.** The cell parser strips the backticks surrounding the **whole cell**, then the value
is split on commas. `` `a/**`, `b/**` `` becomes ``a/**`, `b/**`` and splitting leaves a stray
backtick on each element, so no element matches anything.

## The fix

**Split first, strip per element.** Reversing the order is the whole correction: each
comma-separated element gets its own backtick and whitespace stripping, exactly as a reader would
expect from a list whose elements are individually formatted.

**The validator must agree with the trigger.** `check-visual-verification.sh` parses `ui paths` the
same way and reports a value that yields no usable glob. Today it returns `VISUAL-OK` on a value the
trigger guard cannot use — and the validator is the guard an author runs to check their own
declaration, so it is the worse of the two to be wrong.

## Decisions

### Split on commas first, then strip each element

**ID:** `split-then-strip`
**Status:** active
**Chosen:** parse `ui paths` by splitting on commas and stripping each element's own backticks and
surrounding whitespace.
**Considered:** stripping every backtick anywhere in the cell — rejected, it would silently mangle a
glob legitimately containing one; requiring declarations to drop the backticks — rejected, the
backticked form is this repository's markdown convention for a value and both real declarations use
it, so the parser should read what authors actually write.

### The validator rejects what the trigger cannot parse

**ID:** `validator-agrees-with-trigger`
**Status:** active
**Chosen:** `check-visual-verification.sh` applies the same parse and reports a `ui paths` value
yielding no usable glob.
**Considered:** leaving the validator as a shape check only — rejected: it blessed the value that
caused this, and an author checking their own declaration would be told it is fine.

### The parser is not extracted into `scripts/lib/` — yet

**ID:** `no-shared-parser-yet`
**Status:** superseded by `shared-trim-glob-element`
**Chosen:** both guards parse `ui paths` independently, made to agree by giving both the same rule
and pinning both with harness cases.
**Considered:** extracting a shared parser into `scripts/lib/`, as the BOM handling was — a real
option, chosen against by the operator as larger than the defect warrants. **Recorded because the
argument for it is on the record**: the BOM fix reached one guard of three and had to be extracted
afterwards. If these two parsers diverge again, extraction is the answer rather than a third
alignment.
**Reversed by task 5** on the review panel's evidence and a corrected cost estimate — see
`shared-trim-glob-element` below. The estimate this decision was made on was wrong in one specific
way: it treated extraction as needing new plumbing, when both guards already `source`
`scripts/lib/sanitize-display.sh` and `scripts/lib/strip-bom.sh`, and already share
`scripts/lib/visual-table-cells.awk` for this same table's cells. Against that baseline, extracting
one 18-line, byte-identical function was one new lib file plus two `source` lines — smaller than the
duplication it replaced, not larger. This decision itself is left in place, not deleted, because the
argument it recorded — extraction is the answer if these two parsers diverge again — is exactly the
argument task 5 acted on early, once the cost side of that argument turned out to already be paid.

### `trim_glob_element` is shared, and strips in one pass

**ID:** `shared-trim-glob-element`
**Status:** active
**Chosen:** `trim_glob_element` — byte-identical in `check-visual-trigger.sh` and
`check-visual-verification.sh` — moves into `scripts/lib/trim-glob-element.sh`, sourced by both,
following the exact pattern `sanitize-display.sh` and `strip-bom.sh` already established for these
two guards.
**Considered:** leaving the two copies in place, per `no-shared-parser-yet` — rejected once the
plumbing both guards already share made extraction cheaper than the decision it reverses estimated.

**Severity of the char-at-a-time strip, stated plainly, before describing the fix**: the growth is
quadratic in the padding length, but the constant is small and the worst case measured below (the
ORIGINAL implementation) is sub-second. This is a shape worth fixing while the code is open, not a
denial of service — and it is not new: the same quadratic shape existed before this change too, via
pattern trimming rather than a character loop. It is recorded here because it was found and fixed,
not because it was ever an urgent risk.

**Two bash-native single-pass designs were tried first and MEASURED SLOWER than the original char
loop, not faster** — recorded because both look like the obvious fix and both are a trap in this
bash (5.3.15; reproduced under `LC_ALL=C`, so not a locale artifact):

1. A two-pointer walk using `${s:i:1}` to find the first/last non-padding byte offset, then one
   substring extraction — no repeated copy of `s`, in theory the right fix.
2. The classic no-explicit-loop parameter-expansion idiom,
   `lead="${s%%[!$ws]*}"; s="${s#"$lead"}"` (and the mirror for the trailing run).

Both were benchmarked, not assumed correct by inspection, before being discarded: indexed access and
`%%`/`##` pattern matching against a large string are each quadratic INSIDE bash's own string engine
in this bash version, regardless of how few statements the script itself writes. "No loop written in
the script" is not the same claim as "linear time," and neither delivered it.

**The design that measured linear: one `awk` process, one `sub()` each end.** `sanitize_display`
(`lib/sanitize-display.sh`) already shells out to `awk` for the identical reason — a single-pass,
C-implemented regex substitution instead of bash's own string engine — so this follows an established
convention in this codebase rather than introducing a new one:

```bash verified:quoted verbatim from scripts/lib/trim-glob-element.sh, this worktree
trim_glob_element() {
  LC_ALL=C awk '
    { sub(/^[ \t`]+/, ""); sub(/[ \t`]+$/, ""); printf "%s", $0 }
  ' <<< "$1"
}
```

**Measured**, wall-clock around a single `trim_glob_element` call, a hostile value built from N
backticks of padding on each side of one glob element (`bash 5.3.15`, this machine, `LC_ALL=C`):

```text verified:wall-clock timing in this worktree, both the rejected designs and the shipped one
                                8,000/side (16 KB)   20,000/side (39 KB)   60,000/side (117 KB)
original char-at-a-time              0.05s                 0.19s                 0.84s
two-pointer ${s:i:1} walk            0.55s                 4.81s                47.09s
parameter-expansion %%/## idiom      0.48s                 2.69s                24.64s
shipped: one awk sub() each end     0.005s                0.005s                 0.010s
```

The shipped design is not merely "a single pass" in the sense of one linear scan written in the
script — it is the only one of the three single-pass candidates that measured as one: both
bash-native attempts, despite writing no explicit per-character loop that copies `s`, turned out to
be quadratic inside bash's own indexing and pattern-matching implementation at this scale, and
measured SLOWER than the char-loop they were meant to replace. The `awk` design is faster than the
original at every size tested, including the smallest, and is flat rather than growing across the
three sizes.

**The per-call process-spawn cost this adds to the ORDINARY case (2-8 short elements, one call each)
is a few milliseconds total** — negligible next to either guard's own `awk`-based table parse, which
already spawns at least one `awk` process per invocation.

**The existing harnesses passed unchanged** after the extraction — `scripts/test-check-visual-trigger.sh`
and `scripts/test-check-visual-verification.sh` needed no edits, confirming the rewrite preserves the
char-at-a-time strip's exact boundary behaviour (interior spaces and backticks untouched, an
all-padding element reducing to empty, an empty element unchanged). One new case was added to
`scripts/test-check-visual-trigger.sh`, pinning the strip against a padded element.

## Task 3 — proof against both real declarations

**This repository — triggers.** `stats/web/src/App.tsx` (a real changed path) against this
repository's own `.flow/project.md` (single glob `` `stats/web/src/**` ``):
`scripts/check-visual-trigger.sh .` exits 0, printing `` MATCH: stats/web/src/App.tsx — matched
`stats/web/src/**` ``.

**Gymie — did not trigger for six of eight globs when this section was first written; fixed by task
4, see below.** Read via `git show flow/visual-verification-declaration:.flow/project.md` in
`/Users/tweety53/Projects/gymie` — read-only, that checkout was never switched off its own branch
and nothing was written or committed there. Gymie's `ui paths` declares eight individually
backticked globs. Copying that same text into a throwaway root (never running either guard against
the Gymie checkout itself) and testing two real changed paths from it, against the guard as it stood
after tasks 1–2 (commit `a7e6904`):

- gateway — `src/gateway/src/main/resources/application.yml` against `` `src/gateway/src/main/**` ``
  — `VISUAL-TRIGGER-NO-MATCH`, exit 1.
- a domain module —
  `src/app/auth/src/main/kotlin/com/gymie/auth/infrastructure/presentation/web/AuthController.kt`
  against `` `src/app/*/src/main/kotlin/**/infrastructure/presentation/web/**` `` —
  `VISUAL-TRIGGER-NO-MATCH`, exit 1.

Both globs match correctly in isolation, alone in a two-row declaration — the failure appears only
against Gymie's full eight-glob list, which rules out the `split-then-strip` fix (task 1 already
covers a mixed backticked/bare, three-element and space-containing multi-glob value; those all
pass) and points at something specific to a longer list.

**Root cause — a second, pre-existing bug, independent of the backtick defect.** `glob_to_ere`'s own
declaration line,

```bash verified:quoted verbatim from scripts/check-visual-trigger.sh's glob_to_ere, unchanged by tasks 1-2
local g="$1" out="" i=0 n=${#g} c nc
```

expands every word on a `local` command — `${#g}` included — before any assignment in that same
command takes effect, so `${#g}` reads whatever `g` already held coming into the call, never the
`$1` being bound in the same statement. Minimal reproduction, bash 5.3.15:

```bash verified:run in bash 5.3.15, output shown in the trailing comments
f() { local g="$1" n=${#g}; echo "g=$g n=$n"; }
g="preexisting-longer-value"
f "short"                              # g=short n=24 — 24 is len("preexisting-longer-value")
f "a-much-longer-string-value-here"    # g=a-much-longer-string-value-here n=24 — same stale 24
```

`check-visual-trigger.sh`'s own per-element split-and-strip loop — present before task 1's fix and
carried forward by it, since reusing the loop's own variable name was never the bug — uses `g` as
its loop variable (`for g in "${RAW_GLOBS[@]}"; do g="$(trim_glob_element "$g")"; … done`) and
leaves it holding the LAST raw glob's trimmed value once the loop ends: 19 characters for Gymie's
declaration, the length of its own last glob `` `src/app/src/main/**` ``. Every `glob_to_ere` call
made afterward, while matching, reads `n=19` regardless of which glob is actually being converted,
so any glob longer than 19 characters is silently truncated mid-conversion into a regex built from
its own first 19 characters — which can never match the real path behind it. Six of Gymie's eight
globs exceed 19 characters — `` `src/gateway/src/main/**` `` (23), the two domain-module globs (61
and 63), the two `:common` globs (54 and 64), and `` `gymie-admin-frontend/**` `` (23) — and all six
are silently truncated. Only the two globs at or under 19 characters work, and only by coincidence:
`` `gymie-frontend/**` `` (17) and the trailing `` `src/app/src/main/**` `` itself, which happens to
equal the stale value because it IS the stale value. Confirmed directly: a changed path matching
either of those two DOES trigger; changed paths matching any of the other six do not.

This is why the existing harness never caught it, task 1's new cases included: `test-check-visual-
trigger.sh`'s case 15 (two bare globs) and every new multi-glob case task 1 added put the longer
glob LAST — so the stale leftover was always at least as long as every glob converted before it, and
nothing was ever truncated. Gymie's own ordering puts several long globs before its short last one,
which is what exposes the collision.

**Not fixed by tasks 1–2.** Out of their scope (`split-then-strip` is about backtick-stripping
order, not this variable collision) and out of task 3's own scope as originally written
(`design.md` only) — recorded per this task's own instruction that a continued Gymie failure after
tasks 1 and 2 is a finding to report, not a step to skip. Task 2's validator fix did not catch it
either: it reports only a `ui paths` value with ZERO usable elements after split-then-strip, and
every one of Gymie's eight elements is individually a non-empty string — so
`check-visual-verification.sh` still reported `VISUAL-OK` on a declaration six-eighths of which
could not actually trigger.

**The operator overruled "report, don't fix" once this finding landed**, because the change's own
proposal states both real declarations are verified to trigger — shipping with Gymie still failing
six of eight globs would ship a fix that does not fix the reported problem. Task 4 (commit
`114b16b`) closed it: `n=${#g}` moved into its own `local` statement, evaluated after `g` is bound,
so it reads the freshly-assigned `$1` rather than the split loop's stale leftover. The same sweep
covered every other `local` in `check-visual-trigger.sh` — `trim_glob_element`'s `local s="$1" c`
and `glob_match`'s `local path="$1" glob="$2" regex` — and found neither assigns a variable derived
from an earlier one in the same statement, so neither carries the quirk. Pinned by
`test-check-visual-trigger.sh` case 24: a 76-character glob listed BEFORE a 4-character one, RED
against the pre-task-4 guard (the stale leftover, 4, truncated the 76-character glob to its first 4
characters) and GREEN after.

## Task 4 — all eight of Gymie's globs verified to trigger

Re-run against the fixed guard (commit `114b16b`), the same throwaway copy of Gymie's declaration,
every one of its eight globs, plus a negative control:

```text verified:run against scripts/check-visual-trigger.sh at 114b16b, Gymie's declaration copied into a throwaway root
gymie-frontend/webApp/src/Main.kt                                    -> MATCH `gymie-frontend/**`                exit 0
gymie-admin-frontend/src/App.tsx                                     -> MATCH `gymie-admin-frontend/**`          exit 0
src/app/auth/.../infrastructure/presentation/web/AuthController.kt   -> MATCH `.../presentation/web/**`          exit 0
src/app/admin/.../infrastructure/configurations/AdminSecurityConfig.kt -> MATCH `.../configurations/**`          exit 0
src/app/common/.../common/web/SomeFile.kt                            -> MATCH `.../common/web/**`                exit 0
src/app/common/.../common/observability/CorrelationIdFilter.kt       -> MATCH `.../common/observability/**`      exit 0
src/gateway/src/main/resources/application.yml                       -> MATCH `src/gateway/src/main/**`          exit 0
src/app/src/main/resources/application.yml                           -> MATCH `src/app/src/main/**`              exit 0
README.md (negative control, matches no glob)                        -> NO-MATCH                                 exit 1
```

`scripts/check-visual-verification.sh` still reports `VISUAL-OK` on the same declaration —
unsurprising, since task 2's fix targets zero-usable-glob values and every one of Gymie's eight
globs was always individually well-formed; the defect was in `glob_to_ere`'s own conversion, not in
what `check-visual-verification.sh` parses out of the cell. This repository's own single-glob
declaration was re-confirmed unaffected: `stats/web/src/App.tsx` still matches `` `stats/web/src/**` ``,
exit 0.

## Task 6 — whether the harness can fail spuriously

A review panel slot observed ONE cold run of `scripts/test-check-visual-trigger.sh` fail its
multi-glob cases with the same symptom shape as the original defect — a second glob in a list not
matching — followed by many clean runs, sequential and parallel; it correctly excluded the
observation from its findings, being unable to reproduce it. A later attempt could not reproduce it
either. This task investigated systematically rather than by re-running and hoping.

**Stress testing, to establish a base rate.** 60 sequential runs, 40 concurrent parallel runs, and 80
concurrent parallel runs under artificial CPU load (three `yes > /dev/null` background processes)
were run against the harness as it stands after tasks 1-5 (which include the new case 25, itself a
padded multi-glob case, and the new `scripts/lib/trim-glob-element.sh` extraction):

```text verified:run in this worktree, three separate stress runs against scripts/test-check-visual-trigger.sh
sequential, 60 runs                 0 failures
parallel, 40 concurrent runs        0 failures
parallel, 80 concurrent runs, CPU-loaded   0 failures
```

180 runs, zero failures. This raises confidence but is not, by itself, an answer — "could not
reproduce" is exactly where this investigation started, and a rare enough flake can survive 180
tries. The rest of this section establishes why each category the task named cannot happen, from the
script's own text and from guarantees the operating system and the shell make, rather than from
absence of an observed failure.

**Fixture isolation — cannot collide, not merely does not.** Every `new_root` call in the harness
(`scripts/test-check-visual-trigger.sh`) creates its fixture via `mktemp -d
"${TMPDIR:-/tmp}/check-visual-trigger-test.XXXXXX"`. `mktemp -d` is backed by `mkdtemp(3)`, which
atomically creates a directory under a kernel lock and is specified to retry on a name collision —
it cannot return a name that collides with an existing entry. No two `new_root` calls, whether in the
same process, in two sequential runs, or in two concurrent runs, can ever be handed the same
directory. Every one of the harness's 25 cases (and every parallel copy of the whole harness) reads
and writes only its own `$ROOT/.flow/project.md`, so no case can observe another case's fixture
regardless of timing.

**State leaking between cases — no shared process, so nothing to leak through.** `run_guard` invokes
the real guard as `"$GUARD" "$ROOT" <<< paths` inside a command substitution, which forks a brand new
`bash` process for every single call. Bash keeps no state — no variable, no cache, no compiled form —
between separate process invocations of a script; each call re-reads `check-visual-trigger.sh` from
disk and starts a fresh interpreter with fresh variables. Within that fresh process, every function
that does its own string work (`glob_to_ere`, `glob_match`) declares every variable it touches
`local` — verified by grepping every `local` statement in both guards (below) — so no function call
can leave residue for a later call inside the SAME process either. `trim_glob_element` no longer
does any of its own variable bookkeeping at all: it is now one `awk` subprocess per call
(`scripts/lib/trim-glob-element.sh`), so it has no bash-level local state to leak in the first place.

```text verified:grep -n "^\s*local " scripts/check-visual-trigger.sh scripts/check-visual-verification.sh scripts/lib/trim-glob-element.sh, this worktree
check-visual-trigger.sh:243  local g="$1" out="" i=0 c nc     (glob_to_ere — g, out, i, c, nc)
check-visual-trigger.sh:244  local n=${#g}                    (glob_to_ere — n, its OWN statement,
                                                                 task 4's fix; the only line in
                                                                 either guard where a `local`
                                                                 declaration is split across two
                                                                 statements, and it is split for
                                                                 exactly this reason)
check-visual-trigger.sh:275  local path="$1" glob="$2" regex  (glob_match — none derived from
                                                                 another on the same line)
check-visual-verification.sh:388  local tag="$1" key="$2" rtag rline rkey rval
check-visual-verification.sh:420  local tag="$1" key="$2" label="$3"
```

Every one of these either assigns only from a positional parameter, never from a sibling variable on
the same statement (task 4's own hazard — see Task 3/4 above), or is task 4's fix itself, still
correctly split. No other instance of that hazard exists in either guard today.

**Environment leaking into the guard — not applicable to this guard, and closed for its sibling.**
`check-visual-trigger.sh` calls no `git` command at all — confirmed by grep; its only mentions of
`git` are in comments describing its caller's own `git diff --name-only` step — so `GIT_*` variables
have nothing to reach in this guard. `check-visual-verification.sh` DOES call `git`, and neutralizes
every ambient `GIT_*` variable before doing so via the sourced `git_clean` (`scripts/lib/git-clean.sh`,
pinned by that guard's own case 22) for the identical class of hazard. Neither guard, nor its test
harness, ever `export`s a variable — confirmed by grep — so nothing either harness sets is visible to
a child process beyond what the process's own argument list and stdin already carry. `IFS` is set
in exactly two places in `check-visual-trigger.sh` (`IFS=',' read -r -a RAW_GLOBS <<< ...` and
`while IFS= read -r changed`), both as a command-prefix assignment — POSIX scopes that assignment to
the single command it prefixes, not to the shell that runs it, so neither can leak into any later
statement, function call, or subshell.

**A stale or cached copy — structurally impossible within one worktree.** The harness resolves the
guard under test as `"$SCRIPT_DIR/check-visual-trigger.sh"`, where `SCRIPT_DIR` is derived from the
TEST SCRIPT's own `${BASH_SOURCE[0]}`, not from `$PATH` or any symlink-farm entry. Bash holds no
bytecode cache between invocations of an interpreted script — every `bash file.sh` (or every
subprocess exec of one) re-reads the file's current bytes from disk at exec time. Whatever is on
disk at that path when a case runs is what runs; there is no caching layer inside bash for this
script, in this worktree, to serve a stale copy from. The only way a stale copy could run is
invoking the harness from a DIFFERENT checkout or worktree while editing this one — an operator
mistake, not a property of the script the task asked to be ruled in or out.

**Conclusion.** The four mechanisms the task named are each ruled out by a specific, checkable
property of the current code, not merely by repeated non-reproduction: fixture collision is
prevented by `mkdtemp`'s own atomicity guarantee; inter-case state leakage is prevented by every
call being a fresh process with every working variable declared `local` (and `trim_glob_element` no
longer having any bash-level state to leak at all); environment leakage is prevented by
`check-visual-trigger.sh` never calling `git` and never `export`ing anything, and by `IFS` being
scoped to single commands throughout; and a stale copy cannot run because bash re-reads the script
from disk on every invocation and the harness resolves it by the test file's own location, not
`$PATH`. Combined with 180 stress-test runs producing zero failures, this harness, as it stands
after tasks 1-5, cannot fail spuriously by any of the named mechanisms. No code change was made for
this task — the earlier observation is best explained as a symptom-shape match against task 4's
OWN bug (the `local`/`${#g}` same-line expansion hazard, live in the code between commits `a7e6904`
and `114b16b`), most plausibly observed against an in-flight or intermediate copy of the code from
that window rather than against the harness's own mechanics — offered as the most parsimonious
account available, not as a second confirmed reproduction: nothing above depends on it, and it
changes no conclusion if it is wrong.

## Open questions

None. Task 4 closed the finding it opened: all eight of Gymie's globs, and this repository's own
single glob, now trigger for a real changed path each. Task 6 closed the harness-determinism finding
the review panel raised: see above for the ruled-out mechanisms.

## Task 9 — the cost was mislocated a third time

Task 7's corrected measurement left the guard still costing seconds on hostile input after
`trim_glob_element` was linear. That remaining cost was attributed to `split_cells`'s
character-at-a-time accumulation. **It was mostly not there.**

`split_cells` was genuinely quadratic and was fixed — 0.64s to 0.01s isolated. But roughly 14.8s of
the 15.44s came from **`trimcell`'s `$`-anchored gsub**, which is pathologically slow on any
*computed* string under this machine's awk (`awk version 20200816`, the one true awk), independent
of match complexity. Fixed by finding the trailing run without a `$` anchor.

```text verified:/usr/bin/time -p against check-visual-trigger.sh, 60,000 interior padding per side, 120 KB config, this worktree
before   15.44s
after     0.05s
```

**Task 9 therefore touched `trimcell` and `foldcell` beyond its stated `split_cells` scope**, and
that is recorded rather than quietly absorbed: fixing `split_cells` alone left the guard's own
measured number unchanged, and all three functions live in the file the task already named.

**This is the third mislocated diagnosis in this change**, and the pattern is the point. A panel slot
attributed the cost to `trim_glob_element`; the coordinator attributed it to the wrong measurement
method entirely, and that false number shipped into a file header marked "verified"; this task
attributed the remainder to `split_cells`. Each was corrected only by measuring the specific thing
rather than reasoning from the shape of the code. **The recurring error is not any one wrong guess —
it is trusting a diagnosis that was never measured in isolation.**

## Task 10 — the same error, in a verification claim

Task 10 pinned `split_cells`'s escape semantics with four committed cases, and reported each one RED
against a deliberately broken splitter. Re-running those mutations independently found the mapping
wrong: forcing backslash detection off reds cases 27 and 28, **not** case 30. Case 30's fixture is
`` `a\\|b/**` ``, where the real `|` splits whether or not the preceding `\\` pair is understood, so
the observable outcome is identical under that mutation.

Case 30 is not vacuous — it needed a mutation nothing had written yet:

```text verified:mutations applied to a scratch copy of scripts/lib/visual-table-cells.awk, reverted before commit, lib confirmed byte-identical
bs = 0 (backslash detection off)                  -> reds 27, 28
if (1) in place of nc == "|" || nc == "\\"        -> reds 29 alone
\\ pair also swallows a following |               -> reds 30 alone
```

The four cases all earn their place, and the commit message was amended to state which mutation
actually catches which. **This is the fourth instance of the pattern above and the first in a
verification claim rather than a performance one** — the previous three mislocated where a cost
lived; this one mislocated what a test proves. Both fail the same way: a claim about behaviour,
believed from the shape of the code, that nobody ran. A test asserted to be RED under a mutation
that in fact leaves it green protects nothing, and reads in the record as though it does.

## Round 4 — the panel found a mutation nothing caught

The final panel round ran four slots against task 10's diff. Three returned clean. The
mutation-testing slot enumerated single-point mutations of `split_cells` rather than reading it, and
found one that **all three harnesses stayed green under**:

```text verified:mutation applied to scripts/lib/visual-table-cells.awk, all three harnesses run, lib restored and hash-checked against HEAD
rest = substr(rest, bs + 1)  ->  rest = substr(rest, bs + 2)   # lone-backslash branch
```

That drops the character following a backslash which forms no escape pair. Case 29 covers the
branch, but places its backslash at the end of the glob, immediately before the cell's closing
backtick — a byte `trimcell` strips whether the splitter consumed it or not, so the correct and
buggy paths **coincide there by construction**. The branch's advance distance was untested at every
position where it is observable.

Case 31 fixes that by putting a lone backslash mid-glob (`a\bc/**` against `a\bc/App.tsx`). It reds
under that mutation and is the only one of the 31 that does.

**The lesson is about where a fixture sits, not whether it exists.** Case 29 was a real test of a
real branch, written deliberately, and it still proved nothing about that branch's arithmetic,
because its fixture sat at a boundary another function normalises away. This is the same trap task
8's finding flagged — the boundary-vs-interior distinction in `trimcell` — reappearing one level
down, in a case written by someone who had been told about it.

**Cross-guard exposure, accepted rather than fixed.** `test-check-visual-verification.sh` and
`test-resolve-visual-screenshots.sh` contain no backslash fixtures at all, so the shared lib's escape
semantics are pinned by `test-check-visual-trigger.sh` alone. That is adequate — all three harnesses
are in `## test` and every one of them runs, so a regression is caught once, which is what coverage
owes. It is recorded because "the trigger harness alone protects a lib three guards depend on" is
a fact worth knowing before someone deletes a case from it.

**One process note.** The code-review slot edited `scripts/lib/visual-table-cells.awk` in the
worktree while mutation-testing, against its read-only instruction, then restored it and disclosed
the violation unprompted. Verified independently: both that file and the harness hash-match `HEAD`,
so nothing was lost. The disclosure is the reason it was cheap to check, and is noted here as the
behaviour to keep.

