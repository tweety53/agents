# Review panel — kan-201-reduce-context-rediscovery-across-review-panel

**Pass 1.** Roster: `light`. Panel model: Sonnet. Diff: `.superpowers/sdd/final-review.diff`,
2,224 changed lines across 18 files, against merge base `003fb8b`.

## Diff size

`check-panel-diff-size.sh` measured **2224** against a cap of **2000** — exit 1. The operator was
asked and chose **proceed with the panel anyway**.

## Slots

| Slot | Status |
|---|---|
| 0 — Primary | ran; no findings |
| 2 — Principles (Merged lens) | ran; 1 finding |
| 3 — Code review (low) | ran; 6 findings, 2 pruned by the slot itself before reporting |
| 4 — Security | trigger fired (path/file handling, deserialization) — **declined by operator** |
| 5 — Adversarial | trigger fired (>300 lines, behaviour change to tested code) — **declined by operator** |
| 6 — Lens B | trigger fired (>200 lines) — **declined by operator** |
| 6 — Lens C | trigger fired (error handling, external integrations) — **declined by operator** |

Standards files resolved for slot 2: `CLAUDE.md`, `AGENTS.md`.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low), extended by dispatcher | Critical | `scripts/gather-dispatch-context.sh:186` | the documented principles-path is refused because a global install reaches every skill through a symlink, so the bundle is never produced |
| F2 | Code review (low) | Major | `scripts/gather-dispatch-context.sh:186` | a relative or trailing-slash path argument is refused as a symlink divergence |
| F3 | Principles | Minor | `stats/web/src/views/RunDetail.tsx:213` | hand-rolls DataTable's expand affordance as a flat list, losing per-row key uniqueness |
| F4 | Code review (low) | Minor | `stats/internal/harvest/watcher.go:379` | a sidecar landing after its transcript is harvested loses its descriptors permanently |
| F5 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:249` | `spawn_depth,omitempty` drops a genuine depth 0, making it indistinguishable from absent |
| F6 | Code review (low) | Minor | `stats/internal/store/pricing.go:422` | a malformed `dispatches` value aborts the whole Price call, losing already-computed model costs |
| F7 | Code review (low) | Minor | `stats/web/src/metrics.ts:206` | `dispatches` missing from `KNOWN_TOP_LEVEL_KEYS`, so the raw bag is dumped alongside the formatted table |
| F8 | Code review (low) | Minor | `stats/web/src/views/RunDetail.tsx:155` | the dispatch table has no `.data-table-scroll` wrapper, against the project's own stated convention |

findings-total: 39
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed
finding-status: F27 fixed
finding-status: F28 fixed
finding-status: F29 fixed
finding-status: F30 fixed
finding-status: F31 fixed
finding-status: F32 fixed
finding-status: F33 withdrawn — the operator judged the repetition correct: each of the three sites is an independently re-enterable stage that starts its own stage mark, so a run resumed at section 5 or at a fix round would meet a cross-reference to text it never read, and the CONTEXT BUNDLE block is literal prompt text copied verbatim into a subagent dispatch, where a pointer cannot resolve at all
finding-status: F34 fixed
finding-status: F35 fixed
finding-status: F36 fixed
finding-status: F37 fixed
finding-status: F38 fixed
finding-status: F39 fixed

reproducers-total: 39
finding-reproducer: F1 scripts/gather-dispatch-context.sh . openspec/changes/kan-201-reduce-context-rediscovery-across-review-panel kan-201-reduce-context-rediscovery-across-review-panel CLAUDE.md
finding-reproducer: F2 scripts/gather-dispatch-context.sh . openspec/changes/kan-201-reduce-context-rediscovery-across-review-panel kan-201-reduce-context-rediscovery-across-review-panel CLAUDE.md
finding-reproducer: F3 none — a component-structure finding with no runnable check
finding-reproducer: F4 none — depends on the harness's write-order guarantee between a subagent transcript flush and its sidecar write, not observable from the diff
finding-reproducer: F5 none — verified by inspection; encoding/json omitempty fires on the Go zero value unconditionally
finding-reproducer: F6 none — no current writer produces a malformed dispatches value; would require seeding a stage run's metrics bag directly
finding-reproducer: F7 none — triggering requires rendering MetricsDetail for a run whose bag carries dispatches, a component test rather than a shell command
finding-reproducer: F8 none — reproducing the visual overflow requires a browser render
finding-reproducer: F9 none — the defect is a missing precondition at the source; the shared function's misbehaviour is shown by sourcing it and calling it, which needs a shell construct the reproducer shape forbids
finding-reproducer: F10 none — a byte-identical duplication, established by reading both definitions
finding-reproducer: F11 none — requires a pricer-wired variant of the existing late-sidecar test, which the current test deliberately does not wire
finding-reproducer: F12 none — requires two CommitHarvestBatch calls against a live Postgres and a read-back, not a single command
finding-reproducer: F13 none — requires an injected non-terminal failure for the dispatches-loop lookup alone, which no seam currently allows
finding-reproducer: F14 none — an inert fixture value the SPA never reads; established by comparing the fixture against the producer's own MarshalJSON
finding-reproducer: F15 none — requires a throwaway Go test in package harvest_test; the raising slot ran one and deleted it
finding-reproducer: F16 none — resolve_file is a shell function with no CLI wrapper, so demonstrating it needs a sourcing construct the reproducer shape forbids
finding-reproducer: F17 grep -c meta-sidecar-was-found-count stats/web/src/views/RunDetail.test.tsx
finding-reproducer: F18 none — demonstrating it requires injecting a commit failure into the test file, which mutates the file under review
finding-reproducer: F19 none — a standing duplication risk, established by comparing the fake's merge rule against the migration's
finding-reproducer: F20 none — demonstrating it needs a shell redirect into a not-yet-existing directory, a construct the reproducer shape forbids
finding-reproducer: F21 none — demonstrating it needs a symlinked spec created inside the change directory, a filesystem mutation the reproducer shape forbids
finding-reproducer: F22 none — a byte-identical duplication, established by reading both definitions
finding-reproducer: F23 none — established by grepping for non-test readers of the field, which finds only the two call sites already holding a plain int
finding-reproducer: F24 none — a structural duplication, established by reading the two accumulation blocks alongside the third in encodePatches
finding-reproducer: F25 none — a stale plan figure, established by running the package's own test count
finding-reproducer: F26 test ! -L skills/myflow-fast/scripts/gather-self-review-context.sh
finding-reproducer: F27 none — a stale plan enumeration, established by counting the test functions the file defines
finding-reproducer: F28 none — requires a throwaway Go test seeding an unpriceable models bag beside a priceable dispatch model; the raising slot ran one in a scratch copy and deleted it
finding-reproducer: F29 none — requires three staged RunOnce cycles with interleaved non-assistant lines; established by code trace
finding-reproducer: F30 grep -c func-DispatchBucket-MarshalJSON stats/internal/harvest/attribute.go
finding-reproducer: F31 none — established by comparing the backfill path against the bounded session-token retry path in the same file
finding-reproducer: F32 none — a structural duplication, established by reading the two pricing loops side by side
finding-reproducer: F33 none — a deliberate-repetition judgement, established by reading the three dispatch sites
finding-reproducer: F34 none — a structural duplication, established by reading both normalization implementations
finding-reproducer: F35 none — established by reading the two encodePatches call sites, whose maps start nil on every call
finding-reproducer: F36 none — established by reading the failure routing in add_fixed_source
finding-reproducer: F37 none — a documentation-locality finding, established by reading the two files' comments
finding-reproducer: F38 none — a stale plan figure, established by counting the harness's cases and assertions
finding-reproducer: F39 grep -c func-TestPriceDispatch stats/internal/store/stageruns_test.go

## Dispatcher's note on F1

Slot 3 raised this as a Major scoped to a trailing slash. The dispatcher verified the reproducer and
found a broader, more severe root cause, so the finding is recorded at Critical: `~/.claude/skills/myflow-do`
is itself a symlink created by `setup.sh global`, so `[PRINCIPLES_PATH]` — which
`skills/myflow-do/SKILL.md` requires be the absolute path of the file beside the running `SKILL.md` —
always diverges lexically from its real path in a global install. `validate_path` refuses it, the
script exits 2, and the skill's documented fallback proceeds without a bundle. Half A therefore never
produces a bundle on a global install.

`scripts/test-gather-dispatch-context.sh` case 13 asserts that a symlinked principles-path exits 2,
encoding this defect as the specification — the failure mode
`scripts/test-check-plan-provenance.sh`'s own header records, and which each reviewer was briefed
against.

Root cause is the dispatcher's own task-1 brief, which separated containment-under-worktree from the
symlink-divergence check for the principles path but kept the latter applying to it.

## Bounce log — pass 1

**F1 and F2** (defect identity: `scripts/gather-dispatch-context.sh:186`, theme *lexical/real
divergence refusal applied to a legitimately symlinked or non-canonical path*) were bounced once.
The reproducer the slot declared, `scripts/test-gather-dispatch-context.sh`, exited **0** under
`run-reproducer.sh`, carrying `gather-dispatch-context: all cases pass`. That pass is itself
evidence for the finding rather than against it: harness case 13 asserts a symlinked
principles-path exits 2, so the suite encodes the defect as its specification and can never fail on
it.

The reproducer was replaced with a direct invocation that does demonstrate the defect, verified by
the dispatcher to exit 2 with `worktree '.' resolves through a symlink somewhere between the
repository root and the leaf`, no symlink being involved. One bounce consumed; no fix round consumed,
per the guard-class accounting.

## Pass 2 — full escalation

**Mode: Full, escalated automatically, not asked.** The pass-1 fix altered a guard's behaviour —
`scripts/gather-dispatch-context.sh`'s path validation, under F1 and F2 — which the escalation ladder
names as an automatic trigger. Every **required** slot re-runs against a rewritten
`.superpowers/sdd/final-review.diff`. The four conditional slots were declined by the operator at
pass 1 and are not re-run; their status stays **declined**, distinct from a slot whose trigger never
fired and from one not re-run because its subject was unchanged.

All eight pass-1 findings are marked `fixed`. Verified by the dispatcher before re-running:

- F1's real global-install shape now succeeds — invoking the script with
  `~/.claude/skills/myflow-do/engineering-principles.md`, whose parent directory is a symlink created
  by `setup.sh global`, exits 0 and emits a bundle where it previously exited 2.
- F1/F2's recorded reproducer now exits 0, so `run-reproducer.sh` reports the defect no longer
  demonstrated.

**Footprint change recorded:** the F3 fix nested the dispatch rows inside `StageRunTable`'s existing
per-row detail rather than keeping the flat toggle list, so task 5 now touches
`stats/web/src/components/StageRunTable.tsx` and `stats/web/src/metrics.ts`. `tasks.md`'s task 5
`**Files:**` was updated to declare both, and `check-task-commit-fields.sh` passes against the
rewritten commit. The SPA suite moved 164 → 163: the toggle-uniqueness test was removed along with
the workaround it existed to cover.

## Pass 2 results — five new findings

Slot 0 (Primary): no findings; verified each pass-1 fix independently rather than from this record.
Slot 2 (Principles): F9, F10. Slot 3 (Code review low): F11, F12, F13 — every one a regression
introduced by pass 1's own fixes, which is what a full re-run exists to catch.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F9 | Principles | Important | `scripts/lib/resolve-file.sh:56` | the F2 defect was fixed only at its caller; the shared function keeps the bug and states no precondition |
| F10 | Principles | Minor | `stats/web/src/components/StageRunTable.tsx` | `asRecord` duplicates `metrics.ts`'s `asObject` byte for byte |
| F11 | Code review (low) | Major | `stats/internal/harvest/watcher.go:712` | a late sidecar's backfill never re-prices, so that dispatch's `cost_usd` stays permanently absent |
| F12 | Code review (low) | Major | `stats/internal/harvest/attribute.go:249` | `spawn_depth` is a JSON number, and `jsonb_deep_add` sums numbers, so a re-sent depth accumulates instead of replacing |
| F13 | Code review (low) | Major | `stats/internal/store/pricing.go:476` | a non-`ErrPricingNotFound` error in the dispatches loop discards the models loop's completed result |
| F14 | Dispatcher | Minor | `stats/web/src/views/RunDetail.test.tsx:501` | the SPA fixtures kept `spawn_depth` numeric after F12's fix made the producer write a string |

**Dispatcher verification before dispatching the fix round.** F9 confirmed by calling the shared
function directly: `resolve_file "."` returns `/private/tmp/.` and `resolve_file "/tmp/"` returns
`//`. The exposure is wider than the finding states — `scripts/check-guard-symlinks.sh:250` passes
`$entry`, a path from a directory scan, not only `${BASH_SOURCE[0]}`. F12 confirmed by reading
`0005_jsonb_deep_add.sql:61-62`, which sums two JSON numbers at one key and falls to last-write-wins
only for other types, so the doc comment claiming descriptors resolve last-write-wins holds for the
three string fields and not for `spawn_depth`. The worktree was confirmed clean after slot 3's
throwaway reproducer tests were deleted.

**F12 is the notable one: pass 1's F5 fix introduced it, and `fakeHarvestSink` conceals it** by
hand-coding replace-on-write semantics the real SQL does not have. A fake that diverges from the
store it stands in for cannot catch this class of defect at all.

## Fix round 2, and one finding the dispatcher raised

F9-F13 all fixed. `spawn_depth` is now marshalled as a JSON **string**, which falls into
`jsonb_deep_add`'s non-numeric `ELSE b` branch and so resolves last-write-wins exactly as the three
descriptor strings already did — surviving re-sends unchanged while a genuine depth 0 stays
distinguishable from an absent sidecar. `jsonb_deep_add` itself was not touched, correctly: summing
numeric leaves is what every token figure depends on.

`fakeHarvestSink`'s descriptor merge was corrected to reproduce the SQL's real rule rather than
hard-coding replace-on-write, which is what let F12 hide in the first place.

**F14 — Minor — `stats/web/src/views/RunDetail.test.tsx`, raised by the dispatcher after fix round 2.**
The F12 fix changed the wire shape of `spawn_depth` from a number to a string, but the SPA fixtures
still encoded the numeric form. The SPA reads the field nowhere, so nothing behaved incorrectly —
but a fixture that misrepresents what the producer writes is the same trap that concealed F12, one
layer up. Fixed by moving the fixtures to the string form and recording, at the fixture, why the
producer writes a string at all. Folded into task 5's commit.

## Pass 3 results

Slot 0 (Primary): no findings. Slot 2 (Principles): compliant, one documentation-discoverability
Minor, folded into F16's fix. Slot 3 (Code review low): F15, F16, F17.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F15 | Code review (low) | Major | `stats/internal/harvest/watcher_test.go:131` | the corrected fake still mis-merges the one field F12 introduced, and errors on the ordinary two-cycle case |
| F16 | Code review (low) | Minor | `scripts/lib/resolve-file.sh:63` | the round-2 header asserts the root `/` is left alone, which the code does not do |
| F17 | Code review (low) | Minor | `stats/web/src/views/RunDetail.test.tsx:464` | F14's edit left a duplicated sentence fragment in the comment |

**Dispatcher adjudication — the two slots disagreed and the disagreement was resolved by measuring.**
Slot 0 examined the same `resolve_file` behaviour and ruled it pre-existing and out of scope; slot 3
raised it as a Major regression. Running both versions settles it:

| input | at merge base `003fb8b` | after round 2 |
|---|---|---|
| `/` | `//` | `//` |
| `///` | `//` | `//` |
| `/tmp/` | `//` — leaf dropped entirely | `//private/tmp` |

The double leading slash on a root-parented path is **pre-existing**, so slot 0 was right that it is
not a behavioural regression, and round 2 in fact improved `/tmp/` from dropping the leaf outright to
resolving it. What slot 3 is right about is narrower and still real: round 2 added a header comment
asserting the root case is handled, and it is not. F16 is therefore recorded at **Minor** — a false
claim in new documentation with a small fix — rather than at the Major slot 3 proposed, and the
pre-existing `//` behaviour is fixed alongside it because the fix is cheap and makes the comment true.

**F15 stands as a Major and is the notable one.** It is a third consecutive round in which a fix
introduced a regression, and it defeats precisely the protection round 2 claimed to add: the
corrected fake still cannot catch a wire-representation defect in `spawn_depth`, the very field that
motivated correcting it.

**F17 is the dispatcher's own defect**, introduced by the F14 edit.

## Fix round 3, and a placebo test it uncovered

F15, F16 and F17 fixed. `mergeDeepAddFake` now classifies a leaf by inspecting the raw JSON token's
first byte, matching `jsonb_typeof`, instead of decoding into `json.Number` — whose Go kind is
`string`, which is why it accepted a quoted `"0"` and summed it. `resolve_file` now returns `/` for
`/` and `///`, and `/private/tmp` for both `/tmp` and `/tmp/`, verified under bash 5.3 and bash 3.2;
that also closes the pre-existing doubled-leading-slash behaviour the dispatcher's adjudication
identified, so the header comment is now true rather than merely corrected.

**The notable outcome of this round is not a finding — it is what the fix revealed about round 2's
own test.** `TestSpawnDepthStaysConstantAcrossOrdinaryReSends`, added in round 2 to pin F12, stayed
green against the buggy code: it drives `Watcher.RunOnce`, which logs and swallows a per-path commit
error, so cycles 2 and 3 failed silently while the assertion only ever observed cycle 1's correct
result. The replacement calls `fakeHarvestSink.CommitHarvestBatch` directly, fails against the old
code and passes against the fix.

That is F12's own defect class one layer up: a test that cannot observe the failure it exists to
catch. Three separate instances of it have now appeared in this change — the fake that contradicted
the store, the suite that encoded a guard's defect as its specification (harness case 13, F1), and
this placebo. Each was found by a different mechanism, and none by the test suite itself.

## Pass 4 results

Slot 0 (Primary): **no findings**. Slot 3 (Code review low): **no findings** — and it reports the
harness offered no `code-review` skill this pass, so it **substituted** a general-purpose reviewer
briefed to high-confidence defects only. The substitution is named here as the contract requires; the
slot was not dropped and the panel did not fall back to two required slots.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F18 | Principles | Important | `stats/internal/harvest/watcher_test.go:785` | the test that revealed the placebo pattern was never hardened against it, and still asserts only terminal state across cycles |
| F19 | Principles | Minor | `stats/internal/harvest/watcher_test.go` | the fake carries a second implementation of the migration's per-leaf merge rule, which has drifted twice in three rounds |

**Two independent placebo sweeps, both clean.** Slot 2's Important asked whether the change's testing
approach keeps producing tests that cannot observe the failure they exist to catch. Slot 0 and slot 3
each swept every other `RunOnce`-driven test in the package independently, and both found none: the
outage test asserts `commitCount` by deliberate design, and the backfill tests assert a value that
changes on the exact cycle a silent commit failure would leave stale. The blind spot is therefore
**one test, not a systemic hole** — which is why F18 is recorded at Important rather than higher.

`Watcher.RunOnce`'s log-and-continue on a commit error is confirmed by both slots at
`watcher.go:433-437`, and `resolveSessionTokens` applies the same pattern to `BindSession` errors —
consistent design rather than a second defect.

**Slot 3 mutation-tested its own clean result** rather than asserting it: it reverted
`mergeDeepAddFake`'s classifier to the pre-round-3 shape and confirmed
`TestCommitHarvestBatchPreservesSpawnDepthAcrossTwoCommits` fails while the older test still passes,
then restored it.

**Process observation, recorded because a future panel run can hit it.** Slot 0 observed that
revert appear and disappear in the worktree mid-review, correctly identified it as a concurrent
slot's throwaway edit rather than its own, and verified the tree was clean afterwards. No review was
contaminated here, but panel slots share one worktree, so a slot that mutates it — even briefly, even
to prove a finding — can change what a concurrent slot reads. A slot proving a finding by mutation
should do so outside the shared worktree.

## Fix rounds 4 and 5

**F18 fixed, and the hardening was mutation-verified rather than asserted.** Injecting two failed
commits after cycle 1 left the old assertion passing — demonstrating directly that it had been a
placebo — while the new `commitCount` check failed with `commitCount = 2, want 4`. The generalised
rule sits at `fakeHarvestSink`'s doc comment, beside the existing `failNextCommits` documentation
where a future test author meets it: any `RunOnce`-driven test asserting a value's stability across
cycles must also assert a call-count signal.

**F19 fixed by removal, at the operator's explicit direction.** The operator was offered withdrawal
with a follow-up ticket and chose to fix it in-change instead. `mergeDeepAddFake` and
`isJSONNumberLiteral` are deleted outright, along with the field supporting them; the fake's
descriptor handling is now last-write-wins by presence, so **no JSON-type-aware merge logic remains
in Go to drift from the migration**. The merge-semantics coverage moved to
`stats/internal/store/harvest_test.go`, where the real `jsonb_deep_add` performs the merge:
one test proving numeric leaves sum across two commits at `dispatches.<id>.tokens`, and one proving a
`spawn_depth` string leaf is replaced rather than summed. Tests that never exercised the merge rule
stayed on the in-memory fake, correctly.

**Both defects were proved caught, in a scratch tree outside this worktree.** Reintroducing F12
(marshalling `spawn_depth` as a bare JSON number) failed the relocated test with
`spawn_depth = 2, want 1`; an SQL-layer analogue of F15 — summing any two leaves whose text merely
looks numeric, mirroring the old fake's confusion — failed it on a corrupted decode.

**The fixer caught a methodological flaw in its own first draft**, which is worth recording because
it would have made the proof vacuous: the store-level tests initially built patches from hand-written
JSON literals, bypassing the very `MarshalJSON` that F12 breaks. They were rebuilt to marshal through
the real production types. A test that reproduces a defect only if it exercises the broken code path
is worth nothing when it takes a different path.

`tasks.md` task 4's `**Files:**` now declares `stats/internal/store/harvest_test.go`, and
`check-task-commit-fields.sh` passes against the rewritten commits.

## Pass 5 results

Slot 2: **clean at every severity**. Slot 0: one Minor (F25). Slot 3: five findings, one Major — and
it ran the harness's real `code-review` skill this pass, where pass 4 had to substitute.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F20 | Code review (low) | Major | `skills/myflow-do/SKILL.md:186` | the first bundle redirect writes into `.superpowers/sdd/` before anything creates it, so a fresh worktree's first implementer dispatch never gets a bundle |
| F21 | Code review (low) | Minor | `scripts/gather-dispatch-context.sh:351` | the refusal message says "outside the repository" when the boundary actually checked is the change directory |
| F22 | Code review (low) | Minor | `scripts/gather-dispatch-context.sh:128` | `within_root` is a byte-for-byte copy rather than living in `scripts/lib/`, which this script already sources |
| F23 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:294` | a shadow wire struct and two custom methods exist to string-encode one field no production code reads as an int |
| F24 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:456` | the per-model and per-dispatch accumulation blocks are structurally identical, and the shape repeats a third time in `encodePatches` |
| F25 | Primary | Minor | `openspec/changes/…/tasks.md:442` | task 4's `**Baseline:** after=86` was stale; the package now runs 90 |

**F20 is the second dead-on-arrival defect in this feature, and it was masked by this very run.**
`skills/myflow-do/SKILL.md` invokes `superpowers:subagent-driven-development` at line 205, and that
skill's own workspace script is what creates `.superpowers/sdd/`. The first gather redirect sits at
line 186, nineteen lines earlier, and no `mkdir` appears anywhere in the file — verified by grep. On
a fresh worktree the redirect therefore fails at the shell before the gather script runs, and the
first implementer dispatch falls back to the pre-capability prompt shape. The two later redirects are
safe only because the directory exists by then.

The dispatcher's own run could not have surfaced this: `.superpowers/sdd/` was created by hand here
before `final-review.diff` was first written. **A capability whose failure mode is a silent fallback
needs a test that the bundle was actually produced, not merely that the script can produce one.**

**F25 is already fixed** — the dispatcher corrected task 4's baseline to 90 and attached a
`measured:` provenance comment naming the command that produces it.

**A correction to this record's own earlier framing.** Findings per pass read 8, 5, 3, 2, 5 — not a
clean convergence. Slot 3 reported the harness's `code-review` skill unavailable on pass 4 and
substituted a general-purpose reviewer; this pass the real skill ran and found five. Part of the
apparent convergence was tool availability rather than a diminishing defect population, which is
exactly why the substitution is recorded per pass rather than left implicit.

**Slot 3 refuted one of its own tool's candidates rather than passing it through**: a claim that
`pendingDispatchMeta` loses all but one `agentID` per `(path, stageRunID)`. That contradicts the
documented and planning-verified invariant that one subagent transcript carries exactly one
`agentId`, so the scenario cannot arise. Recording the refusal matters as much as recording the
findings.

## Fix round 6

F20-F24 fixed.

**F20** — `mkdir -p <worktree>/.superpowers/sdd` now precedes all three
`gather-dispatch-context.sh` redirects in `skills/myflow-do/SKILL.md` (sections 4, 5 and the
fix-round dispatch), so each site creates its own target directory instead of depending on an
earlier stage's side effect. Each site also now directs the run to `test -f` the bundle after
writing it and report plainly if it is missing, closing the silent-fallback class rather than only
this instance. Proved on a fresh worktree outside this repository: replaying the pre-fix shell
redirect against a `.superpowers/sdd/`-less directory failed at the shell with "no such file or
directory" (matching the finding's own description) before the gather script ever ran; with `mkdir
-p` first, the same invocation exited 0 and produced a real bundle.

**F21** — the refusal message and its header-comment description in `gather-dispatch-context.sh`
now both read `(resolves outside the change directory)`, naming the boundary `within_root
"$resolved" "$CHANGE_ROOT_REAL"` actually enforces. The boundary itself is unchanged.
`test-gather-dispatch-context.sh`'s two assertions on the old string were updated to match.

**F22** — `within_root` was factored into new `scripts/lib/within-root.sh`, sourced by
`gather-dispatch-context.sh` alongside `lib/resolve-file.sh`; the inline copy and its "Copied
verbatim" comment were deleted. `gather-self-review-context.sh` was left untouched: per
`scripts/lib/resolve-file.sh`'s own header, only a guard that ships through the
`skills/*/scripts/` symlink farm can safely assume a sibling `lib/` travels with it, and that
script explicitly does not (confirmed: no `skills/*/scripts/gather-self-review-context.sh` symlink
exists, unlike `gather-dispatch-context.sh` which is symlinked from both `skills/myflow-do/scripts/`
and `skills/myflow-fast/scripts/`). The new file's own header names this as the same open question
`resolve-file.sh`'s header raises and resolves it the same way. Every harness covering the touched
scripts passed: `test-gather-dispatch-context.sh` (all cases), `test-gather-self-review-context.sh`
(all cases, unaffected), `test-lib-coverage.sh`, `test-plan-dispatch-bundles.sh`,
`test-check-guard-symlinks.sh`.

**F23** — grepped `stats/internal` and `stats/web` excluding tests for a production reader of
`DispatchBucket.SpawnDepth` as an int: none found. The only two writers are
`watcher.go:525`/`:532` and `watcher.go:877`/`:884`, both already holding `meta.SpawnDepth` as a
plain `int` locally (from `DispatchMeta.SpawnDepth`, `transcript.go`). `SpawnDepth` is now declared
`*string` directly on `DispatchBucket`, tagged `json:"spawn_depth,omitempty"`; the `dispatchBucketWire`
shadow struct and its `MarshalJSON`/`UnmarshalJSON` methods are deleted (attribute.go). Both
watcher.go call sites now convert with `strconv.Itoa(meta.SpawnDepth)` at construction.
`encoding/json`'s own type-checking still rejects a corrupted (unquoted-number) `spawn_depth` on
decode, so that guarantee survives the method removal for free. `jsonb_deep_add` untouched.
`stats/internal/store/harvest_test.go`'s `TestCommitHarvestBatchReplacesSpawnDepthStringAcrossTwoOrdinaryCommits`
(the Postgres-backed test pinning both properties) still passes and still drives the real production
type through `buildDispatchPatch`'s `json.Marshal`.

**F24** — extracted a generic `upsertBucket[K comparable, V any]` helper (attribute.go, beside
`TokenDelta.bucket`): lazily initializes the destination map, looks up the current value for a key
(the zero value if absent), passes it to a `combine` closure, and stores the result back. Used at
all four sites that shared the nil-check/lazy-make/get-or-zero/mutate/store-back shape: `Attribute`'s
per-model and per-dispatch token accumulation (attribute.go), where `combine` is an accumulator
(`bucket().add()`), and `encodePatches`'s per-model and per-dispatch patch construction (watcher.go),
where `combine` builds a fresh `ModelBucket`/`DispatchBucket` and ignores the zero value it is
handed. Neither map's own meaning changed and the two are still built independently — `upsertBucket`
only factors the surrounding mechanics, never merges `Models` and `Dispatches` together.

**Verification.** `go build ./...` and `go vet ./internal/harvest/... ./internal/store/...` clean
(the pre-existing, out-of-scope `internal/web/embed.go:29` `pattern all:dist` failure is
unaffected); `gofmt -l .` silent. `go test ./internal/... -race -count=1`: every package passes
except `internal/web` (the same out-of-scope embed failure, a setup failure, not a test failure).
`stats/web`: `npm test` 163 passed (unchanged from the recorded baseline), `npx tsc -b` clean.
`scripts/check-vocabulary.sh`, `check-references.sh`, `check-guard-symlinks.sh`,
`check-contract-budget.sh`, `check-markdown-integrity.py` and `check-stage-mark-calls.sh` all exit 0.

## Pass 6 results

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F26 | Primary, Principles and Code review (low) — all three | Major | `scripts/lib/within-root.sh:14` | the new file justifies leaving a duplicate boundary check by a claim about the symlink farm that is false |
| F27 | Primary | Minor | `openspec/changes/…/tasks.md:434` | task 4 enumerated four `TestPriceDispatch*` tests where five exist |

**F26 is the first finding all three slots raised independently**, at the same file and line and on
the same theme, at severities Major, Important and Minor. Recorded at the highest.

The claim — that `gather-self-review-context.sh` does not ship through the `skills/*/scripts/`
symlink farm — is false and was checkable in one command. Both `skills/myflow-fast/scripts/` and
`skills/myflow-finish/scripts/` symlink that script, and each carries a sibling `lib` symlink, which
is precisely the criterion the header itself states as sufficient. Fix round 6 copied the claim from
`scripts/lib/resolve-file.sh`'s pre-existing header into a brand-new file instead of testing it, and
the dispatcher relayed it onward without checking it against evidence already in this session.

**This is the third instance in this change of one defect class: a claim asserted in a comment that
nothing verifies.** The others were harness case 13, which encoded the guard's own defect as its
specification, and `TestSpawnDepthStaysConstantAcrossOrdinaryReSends`, which stayed green against
broken code. Two of the three were caught only because a reviewer checked a stated fact against the
filesystem rather than reading on.

**Fixed by removal, not by rewording alone.** `gather-self-review-context.sh` now sources
`scripts/lib/within-root.sh` and its inline copy is deleted, so one definition serves both callers;
both headers now state what is actually true. `preserve-session-records.sh` was confirmed to qualify
under the corrected criterion as well — it ships through three skill directories, each with a sibling
`lib`, and carries its own inline `resolve_file` — and is **deliberately not migrated here**: it is
outside this change's scope and its callers are not exercised by it. It is recorded as a follow-up
candidate instead.

**A process note, recorded because it bears on what a clean pass is worth.** Slot 3's tooling behaved
differently in each of three passes: unavailable and substituted on pass 4, available and correctly
scoped on pass 5, and on pass 6 available but pointed at the working tree's git-status diff rather
than `final-review.diff` — so its automated result described two untracked planning files and nothing
of the change. The slot noticed and reviewed by hand, and reported the mis-scoping. Had it not, that
pass would have recorded a clean automated result that examined none of the code. A slot's clean
result is only worth what its tooling actually read.

## Pass 7 results — the pass is not clean

Slot 2: one Minor (F37). Slot 0: one Minor (F38, already fixed). Slot 3: a Major with a proven
reproducer, an Important, and eight quality findings — **after correcting its own earlier report**:
its tooling result returned late and proved to have been correctly scoped after all, so it superseded
a "no findings" conclusion it had already given. A slot that revises its own verdict on new evidence
is behaving correctly.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F28 | Code review (low) | Major | `stats/internal/store/pricing.go:412` | `Price` returns at `priced == 0` before the dispatches loop, so a dispatch whose own model is priceable is never priced when the models bag is entirely unpriceable |
| F29 | Code review (low) | Important | `stats/internal/harvest/watcher.go:454` | the pending-meta delete clears every entry for a path, not only those the batch covered, so a stage run's descriptors can be lost permanently |
| F30 | Code review (low) | Minor | `stats/web/src/views/RunDetail.test.tsx:465` | the comment claims a `MarshalJSON` that F23 deleted |
| F31 | Code review (low) | Minor | `stats/internal/harvest/watcher.go:496` | the backfill retry has no cycle cap or give-up tracking, unlike the bounded session-token path beside it |
| F32 | Code review (low) | Minor | `stats/internal/store/pricing.go:427` | the dispatches pricing loop hand-copies the models loop's sequence |
| F33 | Code review (low) | Minor | `skills/myflow-do/SKILL.md` | the gather procedure and CONTEXT BUNDLE block are each repeated three times |
| F34 | Code review (low) | Minor | `scripts/gather-dispatch-context.sh:96` | `lexical_normalize` duplicates the same walk inlined in `gather-self-review-context.sh` |
| F35 | Code review (low) | Minor | `stats/internal/harvest/watcher.go:871` | `upsertBucket`'s two `encodePatches` closures read as accumulation but always build fresh |
| F36 | Code review (low) | Minor | `scripts/gather-dispatch-context.sh:291` | any `resolve_file` failure is reported as a containment breach |
| F37 | Principles | Minor | `scripts/gather-self-review-context.sh:329` | one helper sourced and a lookalike inlined, with nothing local explaining the asymmetry |
| F38 | Primary | Minor | `openspec/changes/…/tasks.md:94` | task 1's baseline said 17 cases / 43 assertions; the harness has 20 and 50 |

**F28 is the third consecutive dead-feature-class defect**, after F1 (the guard refused the documented
principles path in every global install) and F20 (the first redirect wrote into a directory nothing
had created). All three share a shape: the feature degrades silently, every guard and test passes,
and nothing reports that it did nothing. F28's variant is narrower — it needs the models bag and the
dispatch's own sidecar model to diverge — but that divergence is exactly what per-dispatch
attribution exists to expose.

**F38 is the third stale-baseline finding of this change**, after F25 and F27, all of them the
dispatcher's own plan bookkeeping and all found by a reviewer rather than by the dispatcher. Each is
trivial; the pattern is not. Plan figures written at planning time do not survive fix rounds that add
tests, and nothing in the pipeline re-checks them.

**The operator was given the stop-or-continue decision at this point** — the dispatcher had committed
to stopping rather than spending an eighth round on its own judgement — and chose to fix F28, F29 and
the cheap minors, then re-run only the slots that raised findings rather than a full pass.

## Fix round 7

F28, F29, F30, F31, F32, F34, F35, F36 and F37 fixed. F33 judged and left **open**, for the
operator.

**F28** (`stats/internal/store/pricing.go`) — `Price` no longer returns at `priced == 0` before the
dispatches loop even runs. The models loop still returns early via `return err` on a genuine
non-`ErrPricingNotFound` store error (unchanged), but an unpriceable "models" bag no longer skips
dispatch pricing: a dispatch resolves its own rate from its own sidecar-declared model, independent
of whatever the "models" bag itself could price. `patchFields` now writes a "models" key only when
at least one model bucket actually priced (mirroring how "dispatches" was already conditional), so an
entirely-unpriceable run with a priceable dispatch writes `{"dispatches": {...}}` alone rather than
an empty `"models": {}`. `ErrPricingNotFound` is still returned naming every unpriceable model, and
the top-level `cost_usd` is still gated on `len(missing) == 0`, both unchanged in meaning.
`TestPriceDispatchIsPricedWhenEveryModelBucketIsUnpriceable` (`stats/internal/store/stageruns_test.go`)
pins the exact scenario the raising slot proved in its own scratch copy; confirmed red against the
pre-fix code (`dispatches.agent-1.cost_usd = 0, want 1.6`) and green after.

**F29** (`stats/internal/harvest/watcher.go`, `RunOnce`) — the `hasMeta` branch now deletes only the
`stageRunID` entries this batch's own `deltas` actually carried dispatch entries for, never the whole
`w.pendingDispatchMeta[path]` map. `TestBackfillSurvivesAnInterveningBatchWithNoDispatchDescriptors`
(`stats/internal/harvest/watcher_test.go`) drives the exact staged cycle the finding describes — a
sidecar arrives, then an intervening batch reads new bytes that parse to zero assistant records (a
non-`"assistant"`-typed line, mirroring interleaved tool_use/tool_result), then a final "nothing new"
cycle that must still be able to backfill the earlier, still-pending entry. Confirmed red against the
pre-fix code (descriptors stayed permanently absent) and green after.

**F31** — `maybeBackfillDispatchMeta` is now bounded exactly like the sibling session-token retry
path: new `dispatchMetaCycles`/`gaveUpDispatchMeta` fields (per-path) and a new
`maxDispatchMetaBackfillCycles` constant (= `maxSessionTokenResolutionCycles`, 60), following that
path's own conventions rather than inventing new ones — same warn-once-and-give-up shape, and
`RunOnce`'s own pending-entry recording now skips a path once given up on, so a give-up actually stops
looking, not merely stops logging (the same distinction `TestSessionTokenStopsBeingScannedAfterBoundedGiveUp`
already pins for the sibling path). No new test added — this mirrors an existing, already-tested
pattern rather than introducing new observable behaviour a scratch reproduction could pin more
precisely than reading the mirrored code already does.

**F32** — extracted `priceCandidate` (`stats/internal/store/pricing.go`), the shared
resolve-rate-then-price-tokens sequence both the "models" and "dispatches" loops need. It returns
`(rate, cost, ok, err)`; `err` is left for the caller to decide what to do with, since the two loops
still disagree there (models aborts the whole call on a non-`ErrPricingNotFound` error, dispatches
skips just that dispatch, per F13) — only the part that was actually identical between the two loops
was factored out.

**F34** — judged: **extract**, and done. `lexically_collapse` (new `scripts/lib/lexical-normalize.sh`)
factors out the component-stack walk that was byte-for-byte identical between
`gather-dispatch-context.sh`'s `lexical_normalize` and `gather-self-review-context.sh`'s
`validate_archived_path()` step 2 — each function's own doc comment already said the only real
difference was what a relative input gets joined onto *before* the walk (this process's own cwd vs a
separately-derived trusted root), which is exactly the half left local to each caller. Both callers
already qualify to source from `scripts/lib/` under `scripts/lib/resolve-file.sh`'s own stated
criterion (both ship through the `skills/*/scripts/` symlink farm, each with its own `lib` symlink —
confirmed, the same check F26 ran), so this follows the identical path `within_root`'s own extraction
(F22) already took, including sourcing the same way. Every dependent harness run and passing:
`test-gather-dispatch-context.sh`, `test-gather-self-review-context.sh`, `test-lib-coverage.sh`,
`test-plan-dispatch-bundles.sh`, `test-check-guard-symlinks.sh`, plus the full required guard suite
below.

**F35** — `encodePatches`'s two `mp.Models`/`mp.Dispatches` construction sites
(`stats/internal/harvest/watcher.go`) now use a plain nil-check-and-assign (`make` once, then a
direct map index per entry) instead of `upsertBucket`, which never actually read the zero value it
was handed there. `upsertBucket`'s own doc comment (`attribute.go`) was corrected to describe only its
two genuine remaining call sites (`Attribute`'s per-model and per-dispatch accumulation), which do
read and mutate an existing value.

**F36** — `add_fixed_source` (`scripts/gather-dispatch-context.sh`) now tracks a parallel
`REFUSED_REASONS` array (bash 3.2 has no associative arrays) distinguishing `resolve_file` failure
("unresolvable") from a genuine `within_root` containment breach ("outside"), and the two print
different diagnostics. The source stays excluded from the bundle either way, per the finding. Added
case 21 to `test-gather-dispatch-context.sh`, but not as a reproducer of the "unresolvable" branch
itself: proved directly that `[ -f ]` (add_fixed_source's own absence gate, checked before
`resolve_file` is ever called) already returns false for a genuine symlink cycle, routing it to
"skipped", never "refused" — so a cycle cannot reach the branch this fix changes, and the case instead
pins that the fix left that ordinary disposition alone. The "unresolvable" branch's own reachability
(a `resolve_file`-internal failure — e.g. a component vanishing between the `-f` check and
`resolve_file`'s own walk — with `-f` having already succeeded) is a genuine TOCTOU race a synchronous
single-process test cannot force deterministically; recorded at the fix site rather than claimed
covered.

**F37** — added a one-line, non-restating cross-reference at `gather-self-review-context.sh`'s inline
`resolve_file` definition, pointing at `scripts/lib/resolve-file.sh`'s own header for why this
particular copy stays inline. Corrected while writing it: that header does not say this script fails
the sourcing criterion (it ships through the farm exactly like `within_root` does) — it says migrating
it was left out of scope for the change that migrated `within_root`. The added comment states that
correctly rather than inventing a disqualification reason that isn't there.

**F30** — corrected `stats/web/src/views/RunDetail.test.tsx`'s comment: it no longer claims
`spawn_depth` is written by `harvest.DispatchBucket`'s own `MarshalJSON` (F23 deleted that method).
States what the code now does instead — `SpawnDepth` is a `*string` field with a plain
`json:"spawn_depth,omitempty"` tag, converted via `strconv.Itoa` at construction, needing no custom
Marshal/Unmarshal — while keeping the still-true *reason* (jsonb_deep_add sums numeric leaves) intact.

**F33 — judged, left open.** Read all three sites (section 4 "Execute", section 5 "The review
panel", and the fix-round dispatch under "Panel re-runs") in `skills/myflow-do/SKILL.md`. Each is an
independently re-enterable stage — `/myflow-do` is re-run to apply a fix, and each of these three
sections begins its own `myflow stage begin`, meaning a session executing the fix-round section is not
guaranteed to have sections 4 or 5 anywhere in its own context. Each occurrence already
cross-references the others for the *reasoning* ("same as section 4", "per section 4's rule above")
while repeating the *literal* `mkdir -p` / `gather-dispatch-context.sh` invocation and the literal
CONTEXT BUNDLE blockquote verbatim — the blockquote in particular is not documentation prose but text
copied directly into every dispatched subagent's own prompt, which a cross-reference cannot supply at
dispatch time. This reads as the same "each dispatching stage's instructions must stand alone"
principle `CLAUDE.md`'s own myflow section states elsewhere, applied consistently, not as
copy-paste drift — the three sites have not diverged from each other anywhere they were checked.
Judged correct as written; left open per the instruction not to "fix" a deliberate choice — for the
operator to confirm or override.

**Commit discipline for this round.** `scripts/*` changes folded into `--fixup=d3fc599`;
`stats/internal/*` changes folded into `--fixup=6632d5e`; `stats/web/*` changes folded into
`--fixup=327aa7f`. No `skills/myflow-do/SKILL.md` change this round (F33 left open, not fixed), so no
fixup against `c542ef3`. Rebased with `--autosquash` onto `003fb8b`; five commits remain, same order,
same subjects.

**Verification.** `cd stats && go build ./... `, `go vet ./internal/store/... ./internal/harvest/...`
and `gofmt -l .` all clean. `go test ./internal/... -race -count=1`: every package passes except the
pre-existing, out-of-scope `internal/web` embed failure (`pattern all:dist: no matching files found`).
`scripts/test-gather-dispatch-context.sh`, `scripts/test-gather-self-review-context.sh`,
`scripts/test-lib-coverage.sh`, `scripts/test-plan-dispatch-bundles.sh` and
`scripts/test-check-guard-symlinks.sh` all pass. `scripts/check-vocabulary.sh`,
`check-references.sh`, `check-guard-symlinks.sh`, `check-contract-budget.sh`,
`check-markdown-integrity.py` and `check-stage-mark-calls.sh` all exit 0.

## Fix round 8, and F33's withdrawal

Nine of ten findings fixed. F28 and F29 were each pinned by a test confirmed red before its fix —
F28's returning `dispatches.agent-1.cost_usd = 0, want 1.6`. `lexical_normalize` became the third
helper extracted from a security-relevant validation path in this change, after `resolve_file` and
`within_root`, and now lives in `scripts/lib/lexical-normalize.sh` shared by both gather scripts.

**F36's coverage was reported honestly rather than claimed.** The fixer added a case proving the
ordinary symlink-cycle path is unaffected, and stated plainly that the "unresolvable" branch itself
is a TOCTOU race it could not force deterministically in a synchronous test — rather than writing a
test that appears to cover it. Given that three of this change's defects were tests or comments
asserting something they did not establish, saying "not covered, and here is why" is the more
valuable outcome.

**F33 withdrawn by the operator**, reason on its marker line. The fixer judged the three repeated
`SKILL.md` blocks correct and left the finding open rather than acting on its own judgement, which is
what a fix subagent should do with a finding it believes is wrong: leave it open, say why, and let
the operator decide. It never wrote `withdrawn` itself — only the operator's answer does that.

## Pass 8 — targeted re-run

Slot 0's composition checks all passed. F28 and F32 compose: `ErrPricingNotFound` still names only
unpriceable `models` entries, dispatch failures routing through a separate accumulator that never
reaches it; `cost_usd` is still gated on every model bucket having priced; dispatch costs still reach
only `dispatchesPatch`. F29 and F31 compose too: entries the narrowed delete leaves behind are swept
by the next backfill, which re-reads the sidecar and walks the whole per-path map, and the give-up
path clears map and counter together while `gaveUpDispatchMeta` blocks re-adding — neither stranded
nor leaked. F34's extraction leaves both callers normalising identically, with each keeping its own
make-absolute half.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F39 | Primary | Major | `openspec/changes/…/tasks.md:95` and `:450` | task 1 and task 4 baselines stale again after round 8's own test additions |

**F39 is the fourth stale-baseline finding in this change**, after F25, F27 and F38 — every one the
dispatcher's own plan bookkeeping, every one found by a reviewer, and every one introduced by the
same fix round that added the tests. Measured now: 21 cases and 54 assertions in the gather harness,
91 test functions in `internal/harvest`, and six `TestPriceDispatch*` tests.

**Patching the fourth instance is not the fix.** The cause is structural: a plan's `**Baseline:**`
figures are written once at planning time, every fix round that adds a test invalidates them, and
nothing in the pipeline re-checks them — `check-task-commit-fields.sh` reports `Baseline: skipped,
not verified` precisely because the project's test command cannot be run as a single targetable
command. So the only thing standing between a plan figure and reality is someone remembering, which
has now failed four times in one change. A guard that re-derives declared baselines from the commands
the plan itself names would close it; that belongs in a follow-up, not here.

## Pass 8 closes the panel

Slot 3: **no findings** — all nine round-8 fixes verified, F29 proved load-bearing by reverting it in
a scratch copy and watching the test go red, and the new `lexical-normalize.sh` exercised under both
bash 5.3 and system bash 3.2 with the edge shapes (`/`, `///`, `/tmp/`, `/a/b/../../c`, `//a/b`,
`/../..`) matching the deleted inline implementations exactly. Slot 0: composition clean, one finding
(F39), since fixed. Slot 2 was not re-run: its own F37 was a one-line comment correction, and slot 3
verified that correction directly.

**Why this constitutes a non-stale clean result.** F39 was a stale figure in `tasks.md`, a planning
path excluded from `final-review.diff` and committed only at finish — so correcting it changed
nothing either slot had read. The reviewed artifact is byte-identical to what both slots examined.

**Slot 3's automated tooling has now behaved differently in four of the passes it ran**: unavailable
and substituted (pass 4), correctly scoped (pass 5), scoped to the working tree's git-status diff
(pass 6), and scoped to the main checkout rather than the worktree (pass 8). On pass 7 its result
arrived late and it correctly retracted a "no findings" once it did — which is the only reason F28
was ever found. The slot reported its own state every time, which is what made each of those results
interpretable; had it not, three passes would have recorded clean automated results that examined
none of this change.

**Final tally: 39 findings across eight passes and eight fix rounds.** 38 fixed, one withdrawn by the
operator with a recorded reason. Three were dead-on-arrival defects that made the capability a silent
no-op (F1, F20, F28), and each was found by a different slot at a different pass. Three were claims
asserted in comments or tests that nothing verified (harness case 13, the placebo re-send test, the
symlink-farm claim). Four were stale plan baselines, all the dispatcher's.
