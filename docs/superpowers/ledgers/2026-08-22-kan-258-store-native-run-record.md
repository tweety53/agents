# SDD ledger — kan-258-store-native-run-record

Recorded models: implementation `opus`, review panel `sonnet`, panel fixes `opus` — operator
override at the design gate, per **Model policy** (`skills/myflow-contracts/pipeline.md`).
Roster: `light`. Planning effort: `default`.

Tasks 1+2 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 1+2: implemented and folded into one commit `01bf42a`
`feat(store): add the dispatches and findings tables and their write path` (`Task-Id: 2`).
Model: opus. `./internal/store` 128 → 135 top-level tests, all passing against the live
`myflow-postgres` stack. Commit-fields guard could not run — it refuses while
`openspec/changes/` holds two unarchived changes (kan-77's directory alongside this one) — so the
check was performed by hand: subject and trailer correct, one `Files:` deviation
(`internal/records/types.go`), which was a plan defect and is now corrected in the plan.

Five plan defects reported by the implementer and corrected in `tasks.md`/`design.md` rather than
worked around: `DispatchWindowsForSession` was unbuildable in task 2 (its return type belongs to
task 10) and moved; `records.Dispatch` gained `ID` so the metrics key is reachable typed;
`records.Finding` gained `DispatchSeq` so `dispatch_id` is writable; finding order corrected from
lexical to natural; task 19 now builds the SPA before `go vet ./...`, which cannot pass in a fresh
worktree.

Tasks 1+2 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 1+2 review: clean — one Minor, the `internal/records/types.go` plan-boundary deviation already
corrected in the plan. No Critical or Major. Checkboxes ticked.

Tasks 3+4 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 3+4: implemented and folded into one commit `6a19368`
`feat(api): serve the run record routes and add their client methods` (`Task-Id: 4`).
Model: opus. `./internal/api` 63 → 69, `./internal/store` 135 → 138, `./internal/client` 40 → 45
top-level tests, all passing; no store test skipped. Commit-fields guard still cannot run (two
unarchived changes under `openspec/changes/`); hand-checked instead.

Five further plan defects reported and corrected in `tasks.md`: `RecordStore` has four methods not
five (`MergeDispatchMetrics` has no route); `UpsertFinding` needed a `created bool` to answer 201 vs
200; the client signature now carries that bool through, and task 6's CLI prints `recorded:` vs
`updated:` — an operator ruling, taken here rather than deferred, because the split already exists
server-side and discarding it at the client boundary would make it dead weight; task 4 now owns the
four client methods' own tests rather than declaring none; task 3's file list gains
`internal/api/changes_test.go`, since Go cannot split `fakeStore` across files.

Tasks 3+4 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 3+4 review: 2 Major, 1 Minor.
  - Major: `UpsertFinding`'s `created` flag is not concurrency-safe — the reviewer reproduced 2–3
    spurious `created=true` results per run across 5 runs of 20 concurrent upserts of one new ref.
    The `existing` CTE reads the statement-start snapshot, which `ON CONFLICT DO UPDATE`'s conflict
    resolution never re-evaluates. Fix: `RETURNING (xmax = 0)`, plus the concurrency test that was
    missing entirely.
  - Major: checkboxes for tasks 3–4 unticked. Parent bookkeeping, ticked once the fix passes.
  - Minor: the shipped client `RecordFinding` dropped the `created` return the amended plan
    requires, so the recorded decision was not actually implemented.
Fix round 1: dispatched, model: opus, folded into `6a19368` as a fixup.
Fix round 1: complete, folded into `11936ae` (still one commit on top of `01bf42a`, `Task-Id: 4`
intact). `created` now derives from `RETURNING (xmax = 0)`; the wrong doc comment was rewritten
rather than left beside corrected code. The new concurrency test needed a warm connection pool to be
a real detector — `store.Open`'s `MaxConns=10, MinConns=2` made 20 writers serialise behind
connection dialling, so the naive test caught the bug only 3 runs in 15; with a warm-up barrier the
RED is 15/15 and the fix is 20/20 green. Client `RecordFinding` now returns `created`.
`./internal/store` 138 → 139 (the extra is that test). Tasks 3+4 checkboxes ticked.

Tasks 5+6 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 5+6: implemented and folded into one commit `3aa9a4a`
`feat(myflow): add the record verb and its journal fallback` (`Task-Id: 6`). Model: opus.
`./cmd/myflow` 38 → 44 top-level tests, matching the plan's baseline exactly; `./internal/client`
unchanged at 45. `validateSessionToken` reused verbatim, not copied.

Three flag-set departures reported and folded into `tasks.md` and `design.md`: `-task` is optional
(a required flag makes design.md's own nullable `task_id` case unreachable); `-dispatch-seq` was
added (without it nothing could ever populate the column task 4 taught the store to resolve);
`-round` defaults to 0 (0 is the initial panel and a real value, so there is no "not given").
`record render` exits non-zero while unimplemented — it rendered nothing, and the never-block
guarantee covers an unreachable store, not an absent subcommand.

Tasks 5+6 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 5+6 review: clean — no Critical or Major. The reviewer walked every error path in `record.go`
and confirmed the never-block guarantee holds: no path exits non-zero on a transport failure, and no
path journals a daemon rejection. One Minor, an operator ruling taken here: the delta spec
enumerated three legitimate non-zero exits as if exhaustive, and a request the daemon itself refuses
(400/404) is a fourth. `specs/myflow-run-record/spec.md` now names both kinds of caller mistake —
one the CLI judges before any network call, one the daemon judges having been reached — and states
that neither is journalled, with its own scenario. Checkboxes ticked.

Tasks 7+8 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 7+8: implemented and folded into one commit `1e79913`
`feat(reconcile): replay record journal entries into the store` (`Task-Id: 8`). Model: opus.
`./internal/reconcile` 14 → 17 top-level tests, matching the plan's baseline. Mutation-checked by
hand: disabling the `.journal.record` branch fails all three new tests.

Six plan gaps reported and corrected in `tasks.md`: task 8's Files list missed `cmd/myflow/journal.go`,
`reconcile.New`'s third call site; task 7's tests belong in their own `record_test.go`, following the
package's own structure; task 7 step 4's stated RED was unreachable from an external test package;
step 1 exercises three journal kinds, not two, because `status` has its own journalled shape; and the
embed directory the plan named was wrong — it is `stats/internal/web/dist`, not `stats/web/dist`.
`applyRecordEntry` deliberately does not re-run the API handlers' field checks, on the argument that
the CLI validates before journalling; flagged to the reviewer to probe rather than accepted on the
implementer's word.

Tasks 7+8 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 7+8 review: 1 Major, nothing else. `applyRecordEntry` bypassed the API layer's required-field
checks — and the same file already establishes the opposite pattern twice (`replayFile` calls
`api.DecodeChangeBody`, `replayStageFile` calls `api.ApplyBeginStageMark`). The reviewer demonstrated
the consequence rather than asserting it: the schema's `NOT NULL` columns are satisfied by an empty
string, so a journalled entry carrying `"role":""` inserts a garbage row, returns nil, is read as
success and is retired. Silent corruption of the record this change exists to make trustworthy.
Confirmed clean: the retire-vs-stop split, the suffix dispatch ordering, `status` replay, and the
mutation check (disabling the branch fails all three new tests).
Fix round 2: dispatched, model: opus, folded into `1e79913` as a fixup.
Fix round 2: complete, folded into `9415fbe` (one commit on top of `3aa9a4a`, `Task-Id: 8` intact).
Three exported apply functions in `internal/api` now hold the required-field rule, called by both the
handlers and `applyRecordEntry` — one statement of the rule, in the package that owns it. An invalid
body is retired rather than requeued, because a body the daemon refuses would be refused identically
forever and would block every valid entry behind it. The closed role list stayed in the CLI
deliberately: the delta spec classes an unrecognised role as the CLI's own pre-network judgment.
`reconcile.RecordStore` deleted; `reconcile.New` now takes `api.RecordWriter`, as it already takes
`api.StageStore`. `./internal/reconcile` 17 → 19; both tasks' baselines corrected in the plan.
Mutation check: reverting the three calls fails all 9 required-field subcases. Checkboxes ticked.

Tasks 9+10 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 9+10: implemented and folded into one commit `872cef2`
`feat(harvest): attribute sidechain usage to dispatch windows` (`Task-Id: 10`). Model: opus.
`./internal/harvest` 71 → 76, `./internal/store` 139 → 140; zero skips.

Three plan defects fixed rather than worked around, one of them serious: `MergeDispatchMetrics` used
`jsonb_deep_merge`, which **replaces** numeric leaves — the harvester sends a per-batch delta and
holds no running total, so any dispatch outliving one harvest cycle would have recorded only the last
batch. Switched to `jsonb_deep_add`, which migration 0005 already provides for this exact reason. The
plan also had the session token joining through `changes` (it lives on `stage_runs`) and claimed a
`store`→`harvest` import edge already existed (it did not; the stage source is an adapter in the
daemon).

Tasks 9+10 review: clean, one Minor. The reviewer queried `jsonb_deep_add` directly against the live
database rather than reasoning about it, walked every retry path for double-counting, and confirmed
the additivity assertion is structural. Minor: at an exact `StartedAt` tie, `bestDispatchWindow`'s
"started last" is meaningless and the windows query has no secondary sort key, so the winner is not
reproducible.

Operator ruling: per-slot cost is wrong whenever panel slots run concurrently, because their windows
overlap on one parent session. Asked, and answered — fix it in this change rather than record it.
Tasks 20 and 21 added, running immediately after task 10 and before the renderer, since they change
the `dispatches` shape: a nullable `agent_id`, an optional `-agent-id` flag, attribution matching on
it with the window rule as a fallback that stays the normal path on Cursor and Codex, and the tie-break
made stable. Recorded as decision `agent-id-separates-concurrent-slots` in `design.md`, with two new
scenarios in the delta spec. Checkboxes for 9+10 ticked.

Tasks 20+21 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 20+21: implemented and folded into one commit `b1d3839`
`feat(harvest): attribute sidechain usage by agent id, with a stable tie-break` (`Task-Id: 21`).
Model: opus. `./internal/harvest` 76 → 80, `./internal/store` 140 → 142, `./cmd/myflow` 44 → 45;
zero skips. `0011_dispatch_agent_id.sql` is a new migration; `0010` untouched.

The tie-break test is a real detector rather than a formality: it runs all six permutations of three
same-instant windows and asserts the lowest id wins in each. Deleting the tie-break fails 4 of 6,
deterministically — repeating one order would prove nothing, since the old behaviour is stable for a
fixed input order and unstable only across orders. The "absent never matches absent" rule was
verified by mutation too: the naive `w.AgentID == agentID` fails in both orders, because the cases
put the id-less window earlier so a `"" == ""` match visibly pulls the record away from where the
interval rule places it.

One extra store test beyond the plan's list, and the reason is the defect class this change keeps
meeting: every harvest test uses a fake window source, so `agent_id` missing from
`DispatchWindowsForSession`'s SELECT would leave agent-id attribution silently dead against a real
database while all four harvest tests stayed green. Both baselines corrected to 142.

Tasks 20+21 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 20+21 review: clean, one Minor. The reviewer reproduced both mutation claims rather than
accepting them (tie-break deleted → 4 of 6 permutations fail; naive id comparison → fails in both
orders), applied `0011` end-to-end against a throwaway database, and confirmed the NULL ↔ `""`
round-trip. Minor: the partial index on `agent_id` is dead weight — established by `EXPLAIN` against
a live migrated database, which plans a hash semi join over a sequential scan and references no such
index, not by inspection. Removed, with the reason recorded in the migration itself and in the plan
so the next reader does not add it back. `./internal/store` and `./internal/harvest` green after the
removal. Folded into `b61c5fb`; still one commit on top of `872cef2`. Checkboxes ticked.

Tasks 11+12 (red/green pair, folded): dispatched, model: opus, review: combined (roster `light`).

Tasks 11+12: implemented and folded into one commit `8d48982`
`feat(records): render the ledger and panel record from the store` (`Task-Id: 12`). Model: opus.
`./internal/records` 0 → 7, `./cmd/myflow` 45 → 50; zero skips. The real
`scripts/check-unfinished-work.sh` is invoked from a test as a subprocess, three times: CLEAR with
every finding fixed, OUTSTANDING with one open, and CLEAR again for a record whose note reads
`finding-status: F9 fixed` and whose location reads `findings-total: 99`. No test re-implements the
guard's parsing. The rendered record also passes `check-panel-reproducers.sh`.

Neutralisation replaces the label-forming `:` with U+A789 on the way out only, and flattens line
breaks in the same pass — an embedded newline can carry the second half of a note to the start of a
line, which the colon substitution alone would not cover. `Destination` walks to the nearest existing
ancestor before `EvalSymlinks`, because resolving only the leaf leaves a symlinked parent unresolved
whenever the leaf does not yet exist, which is exactly a first render's state.

Two operator rulings, both taken here rather than deferred:
  - The panel record's filename is `-panel.md`, matching the ~40 records already in
    `docs/superpowers/reviews/` and the script this change retires, not `design.md`'s
    `-review-panel.md`. A second naming convention in one archive directory is the drift this change
    exists to remove.
  - A panel that raised zero findings still renders; `MISSING:` is `ledger`-only. Settled by an
    existing normative statement rather than preference: **The panel record declares how many
    findings it carries** already says zero findings is a declaration and silence is not. Under the
    shipped rule a clean panel wrote no record and the guard reported OUTSTANDING for it.
Both recorded in `design.md`, the delta spec and the plan. Fix round 3: dispatched, model: opus.
Tasks 11+12 review: 1 Major, nothing else. The reviewer built and ran an adversarial pass against the
real guards rather than reasoning about them — tab, LF, CR, CRLF, form feed, vertical tab, NEL,
U+2028, U+2029, a zero-width space splitting a label, an RTL override, and a raw table-row injection,
each in Note, Location, Slot, Severity and Reproducer, all correctly reported OUTSTANDING. Containment
was attacked the same way (symlinked `docs/`, symlinked `docs/superpowers`, symlinked repo root,
relative root, `.` and `..` as change names); nothing wrote outside the root. The retired script's own
glob incident — one change's record overwriting another's — is closed and pinned by a test.

Major: `tokenLine` treats a dispatch as unmeasured only when `len(raw) == 0`, but `insertDispatch`
always defaults empty metrics to `{}` — two bytes — so an unattributed dispatch renders
`input 0, output 0, …`, indistinguishable from a measured zero, which task 12 step 1 and the
function's own comment both forbid. Not an edge case: `{}` is the state of every dispatch before the
harvester runs and the permanent state on Cursor and Codex, which write no transcript, so every
ledger entry on two of three harnesses would misreport zero tokens forever.

**Fourth occurrence in this change of one defect class**: a test backed by a hand-built value or a
fake passes while the real integration is broken. Here the test used a composite literal `Dispatch{}`
with `Metrics` never set — `len == 0`, a shape the store never produces. The fix dispatch requires the
failing test to be built from a value read back out of the real store, and the composite-literal test
kept alongside it rather than replaced.
Fix round 4: dispatched, model: opus.
Fix round 4: complete, folded into `eea61ae` (one commit on top of `b61c5fb`, `Task-Id: 12` intact).
`dispatchMetrics.Tokens` is now a pointer and detection is key presence, not byte length, so `{}`,
`null`, `{"tokens":null}`, whitespace and an absent key all read `not measured`, while a present
`tokens` object renders its figures including explicit zeros. The new failing test was built from a
value read back out of the real store, and the composite-literal test was kept alongside rather than
replaced. `./internal/records` 8 → 9, `./internal/store` 142 → 143. Mutation check: reverting only
the detection fails both. The implementer checked every neighbouring render path for the same
absent-vs-zero assumption and found none, matching `internal/api/stats.go`'s existing key-presence
pattern. Recorded in `design.md`, including that `omitempty` still never fires on `Metrics`, so a
future consumer testing it by byte length inherits the trap. Checkboxes ticked.

Tasks 13+14+15 (contract half, one commit each): dispatched, model: opus, review: combined.

Tasks 13+14+15: three commits, one per task — `d6f4fc3`
`chore(scripts): retire the session-record preservation script` (`Task-Id: 13`), `5b0adeb`
`docs(pipeline): replace session-record preservation with rendering from the store` (`Task-Id: 14`),
`8e773bd` `docs(finish-contract): render the ledger at run 1 instead of copying it` (`Task-Id: 15`).
Model: opus. `pipeline.md` shrank, 49810 → 49751 bytes against a 62262 budget; no budget row is over,
so task 19 has none to raise from these three.

Plan defects reported: task 13 step 2 named `setup.sh`, which never referenced the script — the
symlink set is three committed `scripts/` symlinks, all deleted, since leaving one would be
`check-guard-symlinks.sh` rule 1's own violation. Task 13 step 4's expectation was inverted:
`check-references.sh` matches backticked `.md` paths against headings and never a `.sh` path, so
deleting the script was invisible to it — what broke it was task 14's section rename.

Two gaps the retirement opened, neither owned by any task, both closed by new task 22:
  - **The retired script copied three sources; the render replaces two.** The proposal artifact's
    source is now preserved by nothing, so `pipeline.md`'s `Proposal artifact source` registry row —
    "run 2, only if a preserved copy exists" — can never fire, and no durable copy of the approved
    proposal reaches the repository. Operator ruling: run 1 copies it inline, stated in
    `finish-contract.md`, with the same change-name allowlist `records.Destination` carries. Not a
    new script, which would undo the retirement; not `record render`, which renders from the store and
    has no business copying a file that was never in it.
  - Twenty-five citations across eleven `scripts/` files name the deleted script's header as the
    canonical statement of the change-name allowlist. The rule moved into `records.Destination`; the
    citations are repointed there rather than deleted, since the allowlist's history is why it exists.
Task 22 also covers two stale rationale passages and one stale Go comment.

Tasks 13+14+15 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 13+14+15 review: 1 Major, 1 Minor, both already owned by task 22 — the reviewer reached the
same two gaps independently. Major: `myflow-finish/SKILL-rationale.md:41` describes the deleted script
as running, present tense. Minor: `pipeline-rationale.md` still asserts the copy-vs-staging ordering
asymmetry as current and load-bearing, after task 14 deleted `pipeline.md`'s own copy of it for being
false — and **no grep can catch that one**, because it never names the script literally. Found by
reading. Confirmed clean: the registry's three touched rows, the Model policy claim (backed by real
`dispatches_model`/`dispatches_role` indexes), the stage key unchanged in `names.go`, `README.md`'s
Level 1 table and `myflow-finish/SKILL.md`, both scripts and all three symlinks gone, and
`.myflow/project.md` updated in both lists. Checkboxes ticked.

Task 22: dispatched, model: opus, review: combined (roster `light`).

Task 22: one commit `1c09b1a`
`docs(finish-contract): copy the proposal artifact at run 1 and repoint the retired citations`
(`Task-Id: 22`), 16 files. Model: opus. No budget row over.

Four things the implementer found that the plan had not: `test-check-cleanup-complete.sh` was
**already failing** at `8e773bd` — task 14's registry edit left its declarations stale, and tasks
13–15's review missed it because the harness lives in `## test` while the review ran `## lint`; the
artifact copy needed the destination-containment check too, since `docs/superpowers/artifacts/` is
tracked and run 1 pushes what it wrote; a repointed comment placed inside
`check-workspace-isolation.sh`'s single-quoted awk program broke the guard outright on an apostrophe,
caught by `bash -n`; and task 22's own step 5 grep could not print `clean` at its position, since two
hits live in task 17's file. Plan corrected: task 17 now owns `skills/myflow-do/SKILL-rationale.md`,
which no task owned and which carries the last `check-references.sh` hit, and task 19 now states that
a lint-only check is insufficient, with this change's own evidence for it.

Task 22 review: clean, no findings. 29 of 30 harnesses pass; the one failure is task 17's known
citations. The reviewer verified the moved collating-range measurement against the deleted script's
own content through git history rather than trusting it was carried over, confirmed the citations
genuinely split two ways rather than repointing uniformly, and cross-tallied
`check-cleanup-complete.sh`'s fourteen registry markers 1:1 against `pipeline.md`'s registry table.
Checkboxes ticked.

Tasks 16+17: two commits — `6c3ff64`
`docs(handoff-blocks): report journalled record writes at the IN_PROGRESS gate` (`Task-Id: 16`) and
`891fc64` `docs(myflow-do): record dispatches and findings in the store instead of files`
(`Task-Id: 17`). Model: opus. All guards exit 0, the retirement grep prints `clean`, no budget row is
over (`myflow-do/SKILL.md` 71620 → 77658 against 89525).

**A blocker found, not worked around.** The renderer writes the panel record to
`docs/superpowers/reviews/<date>-<change>-panel.md`; `check-unfinished-work.sh` and
`check-panel-reproducers.sh` both read `<worktree>/.superpowers/sdd/final-review-panel.md`. After
task 17 nothing writes a findings table or a marker block into that file — it survives as the pass
log alone, which carries no `findings-total:` line. Every finish run 1 would report OUTSTANDING and
every fix round would fail its reproducer guard, while every unit test stayed green.
`render_test.go` masked it by copying the rendering into the sdd path, which is why task 12's review
did not catch it, and `proposal.md` still asserts both guards are unaffected. The implementer
declined to invent a bridge, which was correct.

Operator ruling: **one file, at the rendered destination** — the guards resolve
`<worktree>/docs/superpowers/reviews/*-<change>-panel.md`, anchored. Writing both paths was rejected:
it reintroduces the two-locations-that-can-disagree problem this change exists to remove, and run 1
commits from `docs/superpowers/reviews/` anyway. The pass log stays where it is; it is not a findings
record. Recorded in `design.md` and owned by new task 23, which also amends the delta spec's "the
guard SHALL require no change" to say what is actually true — the marker contract is unchanged, the
location moves — and fixes `myflow-fast/SKILL.md`'s stale mechanism sentence.

Tasks 16+17 review: dispatched, model: sonnet, review: combined (roster `light`).
Tasks 16+17 review: 1 Critical, 1 Major, 1 Minor.
  - **Critical: the review panel's own dispatches are never recorded.** `myflow-do/SKILL.md` calls
    `myflow record dispatch` at one code site and one prose mention; section 5 — Primary, Bugbot,
    Principles, the low-effort reviewer, Security, Adversarial, the lens slots and the panel-fix
    subagent — records nothing. Those are the majority of a run's dispatches. It falsifies
    `myflow-model-policy`'s "**Every** dispatch records the model it used", makes the operator-approved
    `agent-id-separates-concurrent-slots` decision dead code (the column exists to separate
    concurrently dispatched panel slots, and no panel-slot row is ever created), leaves
    `findings.dispatch_id` permanently NULL for every panel finding, and leaves `-role red-partner`
    with no call site. No guard asserts dispatch-row coverage, so it would have shipped silently.
  - **Major: the `Records:` line has no way to be computed.** The journal path is an internal Go
    function; hand-deriving it is the exact recipe `state-file.md` withdrew into the CLI. New task 24
    adds `myflow record journal-count`.
  - Minor: `handoff-blocks.md` calls the record journal "the third beside the fallback file and the
    state journal", omitting `.journal.stage` and coining a name `state-file.md` does not use.
Confirmed clean: every format rule survives verbatim and is repointed to bind the renderer, section
7's assert is non-gating, the `prUrl` push path is correct and the git boundary unchanged, and the
`Records:` line is defined once and cited rather than restated.
Fix round 5 + task 24: dispatched, model: opus.
Fix round 5 + task 24: complete. `25f0405` `docs(myflow-do): record dispatches and findings in the
store instead of files` (`Task-Id: 17`, amended in place, subject unchanged) and `2d93c1a`
`feat(myflow): report the journalled record-write count` (`Task-Id: 24`). Model: opus.
`./cmd/myflow` 52 → 55, matching task 24's baseline exactly; no budget row moves.

Every subagent dispatch now has a call site: each panel slot with `-role reviewer -slot <name>
-model -agent-id`, the fix subagent with `-role panel-fix` on `models.panelFix`, and a red task's
partner with `-role red-partner` — the role's first call site anywhere. `record finding` carries
`-dispatch-seq`, with the ordering rule that makes it available. The `unknown (agent-defined)` rule
and the never-invent-an-agent-id rule are both restated at the new call sites.

Two judgements worth keeping: the `handoff-blocks.md` edit was moved out of task 17's commit into
task 24's, because task 17's declared `Files:` list does not name it and
`check-task-commit-fields.sh` reads that as a mismatch — the guard was respected rather than worked
around. And the plan had no rendering for `unknown`, which `journal-count` is required to produce;
a value with no shape in the contract is one an agent invents. Added as the line's third alternative.

Task 23: one commit `f7b0c58` `fix(check-unfinished-work): read the panel record where it is now
rendered` (`Task-Id: 23`). Model: opus. `test-check-unfinished-work.sh` 128 → 136 `ok:` cases,
`test-check-panel-reproducers.sh` 58 → 67; `./internal/records` unchanged at 21 (the plan's "9" was
stale and is corrected).

Both harnesses prove the old path is not read **by mutation, in two directions** — a record present
only at the sdd path reports OUTSTANDING, and an open finding written into the sdd path beside a clean
rendered record still reports CLEAR. Injecting a silent fallback into each guard fails exactly the
case written to catch it. Anchoring is pinned too: another change's dated record, an undated
`demo-panel.md`, and which of two dated files wins.

`render_test.go` **would not** have passed unchanged: with the guards fixed and its helper left alone,
all three guard-running cases fail with `no review panel record`. That was confirmed as the RED step
before any guard was touched — and it is why this mismatch survived task 12's review, since the helper
had been copying the rendering into the guard's own path before checking it.

Plan errors found: `check-panel-reproducers.sh` took no change name at all, so it gained one as a
required second argument with its sibling's allowlist copied byte for byte — forcing one edit outside
the declared file list, to the guard's only call site in `myflow-do/SKILL.md` section 5, since leaving
it would make the documented invocation exit 2. And task 24's commit had introduced a rootless
citation in `handoff-blocks.md` that failed `check-installed-citations.sh`; folded into that commit as
a fixup rather than carried to task 19. Checkboxes for 16, 17, 24 and 23 ticked.

Tasks 18+19: two commits — `82f02b5`
`feat(check-stage-mark-calls): cover the record verb's session token` (`Task-Id: 18`) and `24e993b`
`chore(scripts): raise contract budgets for the store-native run record` (`Task-Id: 19`). Model: opus.
Harness 53 → 59 `ok:` cases; the guard's bare run 38 → 42 call sites. Nine budget rows regenerated to
size + 25% — all nine were still inside their old budgets; the table records the size a file had when
the change touching it landed. Three of the nine are rationale files the plan never listed, written
before tasks 20–24 existed.

**One real defect found by the full run:** `go vet ./...` failed — `internal/web/embed_test.go`'s fake
store never gained the four `api.RecordStore` methods, although task 4 updated its `api.New` call site
to pass four stores. That package had not compiled since task 4, and no earlier task could have seen
it: a fresh worktree cannot vet at all until `stats/internal/web/dist` is built, so every earlier task
vetted only the packages it touched. Fixed in the file's existing zero-value idiom.

**All 14 lint commands exit 0** (the plan listed 13 — it omitted `check-normative-inventory.sh`).
`gofmt -l .` empty, `go vet ./...` clean, `npx tsc -b` clean. **All 31 guard harnesses pass**;
`go test ./... -race -count=1` passes all 15 packages; `npm test` 163 tests across 9 files. Step 4's
grep prints `clean`. Go top-level tests 393 → 459; the plan's `after=411` was stale and is corrected.

One flake, not a regression: `test-check-cleanup-complete.sh` failed once in a run deliberately
overlapped with `go test -race` and the SPA build, and passed 4/4 standalone and in a clean rerun. Its
cases measure wall clock against a bound and the harness's own comment says a case pressed against
that bound would fail on a loaded machine. Neither commit touches that guard.

`proposal.md`'s Impact section was audited against what actually shipped and rewritten by the
dispatcher: it claimed `setup.sh` was modified (it never was — retiring a guard deletes committed
symlinks, since the installer symlinks whole directories), listed `check-unfinished-work.sh` as
unaffected (task 23 edited it), and omitted migration 0011, three Go files, nine guard scripts, three
deleted symlinks, three rationale files and `.myflow/project.md`.

All 24 tasks complete; every checkbox ticked.

## Review panel — NOT RUN

`check-panel-diff-size.sh <worktree> 7262812` measured **9762 lines against the 2000-line cap** and
exited 1. Put to the operator per **5. The review panel** (`skills/myflow-do/SKILL.md`), whose
over-cap prompt offers exactly two courses. **The operator chose to stop and split the change.**

No panel slot was dispatched. This change carries **no whole-branch review**, and
`.superpowers/sdd/final-review-panel.md` deliberately carries no `findings-total:` line —
`findings-total: 0` would assert a clean panel that never ran, and `check-unfinished-work.sh`
correctly reads a record without that line as outstanding. Every one of the 17 commits did pass its
own combined spec-and-quality review on `sonnet` as it landed, and five of those raised findings that
were fixed and re-verified; what is absent is the cross-commit pass.

Stage `do.review-panel` closed with outcome `abandoned`.

**A telemetry blemish the dispatcher caused and could not clean up.** Probing for the valid
`-outcome` values through the CLI rather than reading `stage.go` wrote a `stage end` carrying the
literal outcome `bogus` into `kan-258-store-native-run-record.journal.stage`. The field is free text
and validated nowhere, so the CLI accepted it. Removing the entry was blocked by the permission
classifier and was not worked around; a corrective `stage end -outcome abandoned` was issued instead,
and file order means the later entry wins on replay. Both entries are still journalled — the daemon
is down — so nothing has reached the store yet.

## Review panel — RUN (operator reversed the split decision)

The operator reversed the earlier stop-and-split answer and chose to proceed with the panel over the
full diff. Stage `do.review-panel` reopened as a new attempt.

Three slots dispatched, roster `light`, all on `sonnet` (`models.reviewPanel`):
  - Slot 0 Primary (plan alignment + code quality), model: sonnet — 4 findings, 2 Major.
  - Slot 2 Principles, lens Merged, model: sonnet — 1 finding, Minor.
  - Slot 3 Code review (low), model: sonnet — 10 findings, 4 Major.

**15 findings total, 6 Major.** Every commit had already passed its own per-task review; these are the
cross-commit findings that per-task review cannot reach by construction.

The most serious is an interlocking cluster (F6, F7, F8, F15) that traces to a **dispatcher design
decision that was wrong**. `design.md` chose "one call per dispatch, at close", reasoning that a
begin/end pair "would double the call count and buy nothing". False: the harvester commits its
transcript offset every 5s and never re-reads past it, so a row that does not exist until close means
every tick before that close finds nothing to attribute to — usage dropped or credited to a sibling.
Compounding it, no production path ever set `EndedAt`, so every window stayed open forever and stale
siblings kept claiming later usage; and `RecordDispatch` had no idempotency key, so a replayed
journalled write inserted a second row for one logical dispatch. Per-dispatch cost — one of the four
records this change exists to make trustworthy — would have been quietly wrong.

Slot 0's F2 is the other one worth naming: `openspec/specs/agents-repo-verification/spec.md` is a
**live** capability spec requiring `.myflow/project.md`'s `## test` list to name the harness task 13
deleted, and this change carries no delta spec for that capability — so after merge the repository
would normatively violate its own spec. No per-task review could have seen it: task 13 was correct,
and the spec it broke lives outside this change's directory.

Fix round 1 (code: F4, F6, F7, F8, F10–F15): dispatched, model: opus (`models.panelFix`).

Fix round 1 (code): complete. Branch head `ee2e18a`, still 17 commits, subjects and trailers
unchanged. Model: opus. F6/F7/F8/F15 fixed together as one cluster — `myflow record dispatch` is now a
`begin`/`end` pair, `begin` writing the row as the dispatch starts and printing the allocated seq;
migration `0012_dispatch_key.sql` adds the idempotency key with `UNIQUE (change_id, session_token,
dispatch_key)` so a replayed begin returns the original row; `Window.contains` and
`DispatchWindow.contains` both delegate to one `intervalContains`. F4, F10, F12, F13 fixed; F14
migrated properly rather than withdrawn, resolving a 409 divergence that would have journalled a
conflict for a replay that could never succeed. **F11 withdrawn with the reason recorded in
`groupBySession`'s own doc comment**: the duplicated index-building code is gone, and the second pass
is kept deliberately, because sharing the grouped structure would couple both attributors' signatures
and call order to save one O(n) walk.

One consequence the implementer decided rather than hid: a 404 from `dispatch end` is journalled
rather than refused, because the begin it closes may still be queued ahead of it in the same journal.
`ErrDispatchNotFound` is explicitly non-definitive, with a contrast test proving a finding-ref 404
stays definitive.

Mutation check: reverting only the begin insert made
`TestUsageDuringAnUnfinishedDispatchIsAttributedToIt` fail with `deltas = map[]` — the exact
production symptom, with no error anywhere.

`design.md` gained decision `dispatch-begin-and-end`; the superseded "one call per dispatch, at close"
paragraph is kept and annotated with why its reasoning was wrong.

Fix round 2 (specs and contracts): complete. Branch head `0965779`, still 17 commits. Model: opus.
All five fixed. F2 and F3 required **new delta specs for two live capabilities** —
`agents-repo-verification` and `myflow-self-review` — because a live spec is not editable in place by
a change. Both restate the whole requirement deliberately: OpenSpec's archive replaces a MODIFIED
requirement block wholesale by header match, so a partial restatement would silently delete the rest.
F3 also corrected a citation to a vocabulary word nothing emits any more (`preserved:`). F9 restored
the source-path containment the retirement dropped, stated as the write-side half of the pair
`check_boundary` applies on the read side.

**Panel closed at zero open findings.** `check-unfinished-work.sh` CLEAR; `check-panel-reproducers.sh`
REPRODUCERS-OK, 15 declared. The reproducers were rewritten once: the guard requires a bare path with
plain arguments or the `none — <reason>` exemption, and the dispatcher had written them as shell
command lines.

Final verification, whole branch: `gofmt -l .` empty, `go vet ./...` clean, `go test ./... -race
-count=1` all 15 packages, every `scripts/test-*.sh` harness passes, every lint guard exits 0,
`npx tsc -b` clean, `npm test` 163 passed.
