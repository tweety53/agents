# kan-1121-visual-verify-parallel-frame-shards
> Design record of change kan-1121-visual-verify-parallel-frame-shards (KAN-1121), written at landing. Each decision's own **Status:** line is the status.

## Context

The operator settled three points, verbatim:

- Round 1, Q1 (who runs step 8 on a sharded pass): "**A. prepare captures in Go (recommended)**".
- Round 1, Q2 (which passes shard): "**A. Unscoped, single-worktree only (recommended)**".
- Round 1, Q3 (above 100 frames): "**A. Keep 4 shards, slices grow (recommended)**".
- Round 2, convergence: "**Nothing unclear — approve and move on (Recommended)**".

**Settled by the ticket.** Frames are split by count, about 25 per verifier and up to 4 at once;
each verifier owns its slice of the manifest; the parent takes the one merge-base capture before
any shard starts.

**Defect reachability.** The base has no shards: `prepare`'s `dispatch()` makes one row per
worktree (`stats/internal/guard/visualverifypass_prepare.go:499-571`), and a `-rest` remainder
starts only after the first verifier's `close` (`skills/flow/visual-verify.md:93-102`).

**Frame unit.** One frame is one `<screenshot> <frame id>` sidecar line, which becomes one
manifest pair (`stats/internal/guard/visualverifypass_measure.go:28-43,141-151,533-543`).
`check-frame-list` counts design.md's declared frames instead
(`stats/internal/guard/framelist.go:30-62,153`).

**`measure` always works on the whole spec.** It has no subset flag, and its manifest merge drops
every pair of the same spec (`visualverifypass_measure.go:254-269`), so two calls on one spec
clobber each other even when run one after the other. The `.frame.png` deletion (`:150`),
`matrix/` and the composites all sit in the shared `<changeRoot>/visual-verification/`. The
`--prototype` path already composes from a filtered sidecar under `visual-measure-<key>/`
(`:143-155`).

**The guards refuse a report that covers only a slice.** `check-verify-report` refuses through
`vrUnaccounted` (`stats/internal/guard/verifyreport.go:414-433`) and `check-frame-list` through its
frames count and per-frame lines (`framelist.go:179-196`). The manifest key is the dispatch key
minus `-rest` (`verifyreport.go:179-180,388`).

**`close` assumes one row per worktree.** `endStage` ends the stage when the newest row per
worktree closes (`stats/internal/guard/visualverifypass_close.go:373-403`); the blocked count
`prior` is kept per worktree (`:84-89`); the redispatch key is `visual-verify-<prior+2><suffix>`
(`:92`); `attempts()` takes the worktree's last earlier row (`visualverifypass_prepare.go:724-753`).

**`check-visual-verify-dispatched` accepts any completed `visual-verify*` row**
(`stats/internal/guard/visualverifydispatched.go:733-781`), so one completed shard would pass the
gate. `flow record`'s `inRunFixRefusal` (`stats/cmd/flow/record.go:4653-4706`) likewise treats one
completed `-full` shard as closing the fix, which leaves this gate as the only whole-set check.

**Nothing blocks four shards from running at once.** `flow record` refuses no second open
verifier row (`stats/cmd/flow/record.go:1394-1407`); the hook admits any `visual-verify-*` key
(`hooks/enforce-agent-baseline.py:136`); the parent's turn discipline already waits on any number
of background children (`skills/flow/implement.md:945-970`).

**Shared writes inside each verifier today.** Step 8 runs specs and the whole-spec
`capture --update-snapshots=all`, authors spec and sidecar lines, and updates the full app suite
and its zip (`skills/flow/visual-verify-verifier.md:12-98`). Step 10's merge-base capture runs
lazily in a fixed `../<name>-merge-base` worktree on its own stack (`:230-265`). Step 11 writes the
one `<changeRoot>/visual-verification.md`. The `prototype/` and `motion/<id>/` directories are
shared too.

**`vvcExit` reads every integer on the verdict's `exits:` line** (`visualverifypass_close.go:293,306-310`),
so a shard key such as `s2of4` on that line would read as a non-zero exit.

**Two existing defects, confirmed by reading.** `close` builds the redispatch key as
`visual-verify-<prior+2><suffix>` (`visualverifypass_close.go:92`), dropping `-fix-<k>[-full]`;
while In-run fix k is ticked and not closed, `record.go:4687-4692` refuses a key without
`-fix-<k>`, so `prepare` stops with `STOP: row`. No test covers it
(`visual_verify_pass_close_test.go:166,178` check only the bare key). And
`skills/flow/visual-verify.md:189-191` says the bundle carries the step-9 `measure` call, but
`bundle()` (`visualverifypass_prepare.go:621-708`) writes no such line.

## Decisions

### Redispatch key keeps the fix-round segment

**ID:** redispatch-key-keeps-round
**Status:** active
**Chosen:** `close` builds the redispatch key by putting `-<prior+2>` after `visual-verify` and keeping every later segment of the closing key: `-fix-<k>[-full]`, the shard segment and the repository suffix — one close test covers a `visual-verify-fix-3` redispatch. Confirmed by reading: `visualverifypass_close.go:92` builds `visual-verify-<n><suffix>`, and while In-run fix k is ticked and not closed, `record.go:4687-4692` refuses a key without `-fix-<k>`.
**Considered:** filing it as a separate ticket — rejected because now is always taken over later, and shard keys depend on the same function.

### Every bundle names the measure call

**ID:** bundle-names-measure-call
**Status:** active
**Chosen:** `bundle()` writes the step-9 `flow-guard visual-verify-pass measure` call with its arguments into every verifier bundle, sharded or not, with `--frames <slice>` on a shard's bundle — `skills/flow/visual-verify.md:189-191` already says the bundle carries this call, but `bundle()` (`visualverifypass_prepare.go:621-708`) writes no such line (confirmed by grep).
**Considered:** adding the line to shard bundles only — rejected because it would leave the unsharded bundle out of line with its own contract.

### Merge-base capture only for specs missing a merge-base baseline

**ID:** merge-base-capture-when-baseline-missing
**Status:** active
**Chosen:** the parent's up-front capture runs only for specs with at least one PNG absent at the merge base (`git cat-file -e`); every other attribution reads `git show <merge-base>:<png>`. An unsharded pass keeps today's lazy in-verifier capture.
**Considered:** a whole-spec capture on every sharded pass — rejected because it starts a stack even when every baseline exists. Moving the unsharded pass's capture to the parent too — rejected because the ticket's reason, no second stack per shard, does not apply to a single verifier.

### Change capture runs before the merge-base capture

**ID:** capture-before-merge-base
**Status:** active
**Chosen:** `prepare` runs `setup` and the change's capture first, then lists the specs missing a merge-base baseline and prints `NEXT: merge-base-capture` — the PNG list the `git cat-file -e` test needs exists only after the change's capture, and a capture failure falls back to one verifier before the parent has started a merge-base stack for nothing. Auto-taken in round 2 under the recommended option.
**Considered:** the merge-base capture first, then `setup` and capture under `--merge-base-captured` (design-1) — rejected because a new test's PNG does not yet exist to be tested, and a failed capture wastes the merge-base stack already run.

### Prepare runs step 8's capture on a sharded pass

**ID:** prepare-captures-sharded
**Status:** active
**Chosen:** the operator's Q1 answer A (round 1): `prepare` runs `setup` and `capture --update-snapshots=all` per spec in Go before any shard; no shard edits a spec or sidecar; s1 updates the full app suite and its zip.
**Considered:** a serial `<key>-capture` verifier before the shards — rejected by the operator (one extra dispatch of latency on every sharded pass; a capture failure stops the round before any judging). Each shard capturing its own tests with `-g` — rejected by the operator (concurrent Playwright runs race on the PNGs, spec and sidecar).

### Shard only unscoped single-worktree passes

**ID:** shard-unscoped-single-worktree
**Status:** active
**Chosen:** the operator's Q2 answer A (round 1): only the first pass and a `-fix-<k>-full` run of a single-worktree set shard. `CLOSES-LOOP`, `closesLoop`, the repository suffixes and the `-fix-<k>` refusals in `record.go` stay unchanged; `close` keys its blocked count and `endStage` by shard index alone, with no stage-file lock and no loop-verdict aggregation.
**Considered:** every pass and every set, with the 4-shard cap per worktree, `CLOSES-LOOP` aggregated across shards, per-(worktree, shard) lanes and a stage-file lock — rejected by the operator (about 600 lines of Go against about 350, and the aggregated loop verdict is the most error-prone part).
<!-- predicted: git diff --stat of either design once implemented -->

### Four shards at most, slices grow

**ID:** four-shards-slices-grow
**Status:** active
**Chosen:** the operator's Q3 answer A (round 1): n = min(4, ⌈F/25⌉); above 100 frames each slice grows past 25 and an over-long shard relies on the `-rest` remainder.
**Considered:** slices of exactly 25 with at most 4 running and the next dispatched as each closes — rejected by the operator (dispatch-on-close logic in the parent and `prepare --continue`).

### Each shard owns its own manifest file

**ID:** manifest-per-shard
**Status:** active
**Chosen:** `visual-manifest-<shard key minus -rest>.json`, one writer per file — the same-spec merge clobber cannot occur across shards.
**Considered:** one shared manifest with a lock and a per-frame merge — rejected as a lock and a merge rewrite for no reader that needs one file.

### Shard key shape

**ID:** shard-key-segment
**Status:** active
**Chosen:** the segment `-s<i>of<n>`, carried by `close`'s `suffix()` — it matches neither the redispatch regex (`visualverifypass_prepare.go:467`) nor `fixRoundKey`/`fullRunKey`, and putting n in the key lets the gate derive the full set without stored state.
**Considered:** `-s<i>` — rejected because it could collide with a repository suffix and does not carry n. `-shard<i>of<n>` — rejected as longer, with no gain.

### Slice unit and cut

**ID:** slice-by-frame-id
**Status:** active
**Chosen:** distinct frame ids, in contiguous slices, never cut between two frames of one screenshot — the unit the ticket counts, and one writer per composite.
**Considered:** capture-weighted units — rejected as an extra weighting rule for the same result. A free cut — rejected because two shards would then write one screenshot's composite.

### Lead shard owns shared, non-frame work

**ID:** lead-shard-owns-shared
**Status:** active
**Chosen:** s1 owns the views no frame maps, the motions, the `identical` lines, and the full app suite and its zip — one writer for every shared file.
**Considered:** round-robin across shards — rejected because `motion/<id>/` and the zip are shared files, and two writers race on them.

### Record parts merged by close

**ID:** record-parts-merged-by-close
**Status:** active
**Chosen:** each shard writes `record-s<i>of<n>.md`, and `close` rebuilds `visual-verification.md` from the parts in shard order — self-review and `flow change save` keep reading the one file (`stats/internal/selfreview/git.go:390`).
**Considered:** the parent writing the record from the manifests — rejected because the record carries judged prose no manifest holds.

### Closes run one at a time, no lock

**ID:** closes-serial-no-lock
**Status:** active
**Chosen:** the parent runs each shard's `close` as its own Bash call, one after another, and the prose says so — no concurrency exists for a lock to guard.
**Considered:** a stage-file lock in `close` and `prepare` — rejected as code against a concurrency the parent never creates.

### Capture failure falls back to one verifier

**ID:** capture-failure-falls-back-serial
**Status:** active
**Chosen:** a non-zero `setup` or `capture` exit on a sharded pass prints `SHARDS: none — <cmd> exit <n>` and dispatches today's single verifier, which re-runs and reports the failure — no new failure path.
**Considered:** blocking the stage — rejected because it adds a new failure path, where the single verifier already handles this one.

### A shard's writing state check uses an account of its own

**ID:** shard-state-writes-own-account
**Status:** active
**Chosen:** on a sharded dispatch, a step-10 state check that writes to the stack — a submitted form, or data seeded to reach a comparison branch — runs signed in as an account the shard registers fresh for it, as a capture spec that registers fresh does (`skills/flow/visual-verify.md:378`), never the seeded account the other shards' hand captures read; a check that only reads, or types without submitting, stays on the seeded account. The verifier's sharded-dispatch paragraph says so.
**Considered:** one seeded account per shard, named in the bundle — rejected because the dev seed is the project's and holds one account, so n accounts would be a new project setting. s1 running every writing check — rejected because a check belongs to its frame, which another shard owns. Accepting the risk — rejected because a disturbed hand capture reads as a departure and becomes a fix task against a defect that does not exist.

### One fix task, opener without the shard segment

**ID:** fix-opener-pass-key
**Status:** active
**Chosen:** one task per round, transcribing every shard report with a defect; `**Opener:**` is the pass key without the shard segment — unique per round, so F24 (`stats/internal/guard/planshape.go:680-707`) is unchanged.
**Considered:** listing the shard keys comma-separated — rejected because the string changes with n and adds nothing F24 needs.

## Open questions

## Recommended approach

1. **When `prepare` shards.** In the preflight run (`visualverifypass_prepare.go:266-291`) it
   counts F, the frames on the change's list: distinct frame ids on the capture specs' sidecar
   lines, plus declared frames (`flDeclared`, `framelist.go:208`) that have no sidecar line.
   - It shards only when all of these hold: F > 25, `mockups` is declared, the set has one
     worktree, no `--round-base` was passed (the pass is unscoped), and the stage file has no
     `row` line yet.
   - Otherwise it writes nothing new, and the run is today's path byte for byte.
   - When it shards, n = min(4, ⌈F/25⌉). The cut unit is a group: frame ids joined, transitively,
     by a shared screenshot, across every spec — so a frame id lies in one group however many
     specs map it, and its `bands/<frame>.json` and `matrix/<frame>.json`, keyed by frame id alone
     (`visualverifypass_measure.go:181-182,197,206`), have one writer, as each composite does. Groups are
     taken in order of first appearance (`specs` order, then line order); slices s1…s(n-1) each
     close at the first group boundary at or past ⌈F/n⌉ frames — a cut that would split a group
     moves forward to its end — and the last slice takes the rest. No slice is empty: when the
     groups run out before n slices, n is the number of slices made, and n = 1 is today's
     unsharded path. Above 100 frames n stays 4 and each slice grows past 25.
   - It writes one `spec <abs>` line per capture spec and one `slice s<i>of<n> <frame,…>` line per
     shard to the stage file, and deletes every `<changeRoot>/visual-verification/record-s*.md` —
     at the start of every sharded pass, a `-fix-<k>-full` one included, and nowhere else. A
     capture fallback (step 2) runs after this deletion; an unsharded pass leaves old parts on
     disk, which no unsharded `close` reads.
2. **Step 8 in `prepare`.** Once the stack is up (after `probe` and `fingerprint`, before any row
   begins), a sharded `prepare` runs `setup`, then `capture <spec> --update-snapshots=all` for
   each `spec` line, from `captureRoot`, through `kwRun` as it runs `specs` today
   (`prepare.go:270`), and appends `captured` to the stage file.
   - A non-zero exit prints `SHARDS: none — <cmd> exit <n>`, drops the `slice` lines and
     dispatches today's single row; that verifier re-runs the capture and reports the failure as
     it does now.
   - No shard edits a capture spec or a sidecar and no shard re-captures a spec. A declared frame
     with no sidecar line goes into s1's slice, `check-frame-list` reports its `no sidecar line`
     gap, and the pass reaches the loop.
3. **The parent's merge-base capture.** After the capture, `prepare` names each spec with at least
   one PNG absent at the merge base (`git cat-file -e <merge-base>:<png>`).
   - None: it dispatches the shards at once.
   - Some: it prints one `MERGE-BASE-SPEC: <spec>` line each, then `NEXT: merge-base-capture`, and
     begins no row. The parent runs step 10's merge-base recipe
     (`visual-verify-verifier.md:245-261`, moved into a `## The merge-base capture` section of `visual-verify.md`) once for every named spec
     in its own Bash calls — worktree, derived workspace, stack, `capture`, stop, remove — copies
     every PNG it wrote to `<changeRoot>/visual-verification/merge-base/<path relative to the
     screenshots root>`, then runs `prepare --continue --merge-base-captured` (adding `--started`
     again if it started the main stack). A `captured` stage line means the capture is not re-run.
   - A failed merge-base capture is passed as `prepare --continue --merge-base-failed "<reason>"`,
     which puts `merge-base capture failed — <reason>` into every shard bundle; the departures then
     count as this change's, as the rule already says.
   - Every other attribution reads `git show <merge-base>:<png>`, as today.
4. **The shards.**
   - Keys: `<stage key>-s<i>of<n>` — `visual-verify-s2of4`, `visual-verify-fix-3-full-s1of3`; a
     remainder `visual-verify-rest[-<m>]-s2of4`, a redispatch `visual-verify-<m>-s2of4`. No
     repository suffix, since a sharded set has one worktree.
   - A shard key passed to `prepare --continue` makes exactly one row for it, reading its slice
     from the stage file; `attempts()` counts an earlier row only when its shard segment matches.
   - Manifests: each shard owns `visual-manifest-<key minus -rest>.json`, so the same-spec merge
     clobber cannot occur.
   - Bundle: names the slice and the exact call `flow-guard visual-verify-pass measure … --frames
     <slice>`. `--frames` writes the filtered sidecar to `visual-measure-<key>/<spec
     basename>.mockups` through the prototype seam, so composites, `.frame.png` and matrix files
     touch only the slice's own files. T1 proves `compose-mockup-frames` accepts a filtered sidecar
     without `--prototype`.
   - The lead shard s1 also owns the full app suite update, its `-g` capture, the zip rebuild and
     `check-spec-reach`, every view no frame is composed for (prototype render and
     `measure --prototype` included), the motions and the `identical` lines. Every other shard's
     prompt reads `motions: none`, its `close` gets `0` motions, and its `exits:` line says
     `full app suite: not run`.
5. **`close` per shard.**
   - It reads the slice from the stage file and passes `--frames <slice>` to `check-verify-report`
     and `check-frame-list`. `check-frame-list --frames` checks only declared frames in the slice
     and compares the frames line's `m` to that count. `check-verify-report --frames` drops from
     `vrUnaccounted` a PNG the sidecar maps only to frames outside the slice; a PNG no sidecar line
     maps is required only of an `-s1of<n>` report (key read from the filename, as at
     `verifyreport.go:388`).
   - The shard segment joins `suffix()`, so `restKey` and the redispatch key keep it. The blocked
     count `prior` counts only rows of the same shard index, and `endStage` tracks the newest row
     per shard index, ending the stage once every one is closed — `stopped` if any is blocked.
   - Each shard writes `<changeRoot>/visual-verification/record-s<i>of<n>.md`, never
     `visual-verification.md`. On the `judge`, `loop` and `done` paths `close` rewrites
     `visual-verification.md` as the parts s1…sn joined in order.
   - Step 12's commit and `ROUND-BASE:` run on each closing row; a commit with nothing staged is
     skipped (`close.go:464-466`), and the parent takes the last `ROUND-BASE:` printed.
     `CLOSES-LOOP` never applies, because sharded passes are unscoped.
   - The parent runs each shard's `close` as its own Bash call, one at a time. It acts on `rest`,
     `redispatch` or `cure-environment` per shard at once, and on `loop` or `judge` only once every
     shard of the pass has closed.
6. **`check-visual-verify-dispatched`.** `vvdCompletedVerifiers` also collects shard rows of every
   outcome. For the n of the newest shard row, every index 1..n must have its newest
   `-s<i>of<n>` row completed, and then only those newest rows are kept; otherwise every shard row is dropped before the newest-row and head
   loop (`visualverifydispatched.go:420-431`). The rule reads shard rows alone: an unsharded row —
   a scoped `-fix-<k>` row, or a capture fallback's single row — is never part of a shard set, and
   the loop then compares every surviving row by `startedAt` as today, so an unsharded row newer
   than the shard set is the newest row and governs. Shard rows are not grouped by `diffBase`, since an
   earlier shard's commit moves HEAD, and `newestRoundBase` stays empty for them.
   `verify-and-handoff.md:308`'s "newest row" becomes "the newest row of each shard index".
7. **Fix loop.** One round makes one task, transcribing every shard report that names a defect. Its
   `**Opener:**` is the pass key without the shard segment, which keeps F24 unchanged.
8. **Prose.** `skills/flow/visual-verify.md` (the sharded dispatch, `NEXT: merge-base-capture` and
   the parent's merge-base recipe, waiting for every shard), `skills/flow/visual-verify-verifier.md`
   (one "sharded dispatch" paragraph over steps 8, 9, 10's merge-base paragraph and its writing
   state checks (`shard-state-writes-own-account`), and 11, plus the `exits:` digit rule — no shard
   key on that line), `skills/flow/verify-fix-loop.md`,
   `skills/flow/verify-and-handoff.md`, `skills/flow/implement.md` (closed-list key shape, and a
   one-mutator exception for the shards of one pass), `skills/flow/SKILL-rationale.md`, and the
   inline `visual-verify-pass` description in `README.md`'s `**Verify:**` paragraph.

Rejected alternatives: a serial capture verifier or per-shard `-g` capture
(`prepare-captures-sharded`); sharding every pass and set with an aggregated loop verdict
(`shard-unscoped-single-worktree`); a queue of 25-frame slices (`four-shards-slices-grow`); a
shared locked manifest (`manifest-per-shard`); a stage-file lock (`closes-serial-no-lock`); the
merge-base capture before the change's capture (`capture-before-merge-base`); a whole-spec
merge-base capture on every sharded pass (`merge-base-capture-when-baseline-missing`); the parent
writing the record from manifests (`record-parts-merged-by-close`); a seeded account per shard, or
s1 running every writing state check (`shard-state-writes-own-account`).

## Critical files

- `stats/internal/guard/visualverifypass.go` — header usage text and flag parsing (`:425-530`):
  gain `--frames`, `--merge-base-captured`, `--merge-base-failed`; stage-file reading of `spec`,
  `slice`, `captured` lines.
- `stats/internal/guard/visualverifypass_measure.go` — `measure --frames <slice>`: filtered sidecar
  through the prototype seam (`:140-155`).
- `stats/internal/guard/visualverifypass_prepare.go` — preflight count and slice (`:266-291`),
  setup and capture with fallback, merge-base step, `dispatch()` shard rows (`:499-571`),
  `bundle()` measure-call and slice lines (`:621-708`), `attempts()` shard filter (`:724-753`),
  old `record-s*.md` deletion.
- `stats/internal/guard/visualverifypass_close.go` — `suffix()` shard segment (`:331-337`),
  redispatch key (`:92`), `prior` per shard index (`:84-89`), `--frames` to both guards (`:52,61`),
  `endStage` per shard index (`:373-403`), record-part merge.
- `stats/internal/guard/verifyreport.go` — `check-verify-report --frames`, `vrUnaccounted`
  (`:414-433`).
- `stats/internal/guard/framelist.go` — `check-frame-list --frames` (`:153,176-196`).
- `stats/internal/guard/visualverifydispatched.go` — shard-set rule in `vvdCompletedVerifiers`
  (`:733-781`) and the newest-row loop (`:420-431`).
- `stats/internal/guard/visual_verify_pass_measure_test.go`, `visual_verify_pass_prepare_test.go`,
  `visual_verify_pass_close_test.go`, `check_verify_report_test.go`, `check_frame_list_test.go`,
  `check_visual_verify_dispatched_test.go` — the tests beside each.
- `skills/flow/visual-verify.md` — sharded dispatch, parent's merge-base recipe, waiting on every
  shard; its `**The relay contract.**` paragraph stays blank-line free
  (`visualverifypass_prepare.go:573-594` copies it).
- `skills/flow/visual-verify-verifier.md` — the sharded-dispatch paragraph and `exits:` digit rule.
- `skills/flow/verify-fix-loop.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/implement.md`,
  `skills/flow/SKILL-rationale.md` — key shape, newest-row-per-shard wording, one-mutator
  exception, fix-task opener, rationale.
- `README.md` — the `**Verify:**` paragraph (`:264`), which describes `visual-verify-pass` inline.
- Depends on, unchanged: `stats/cmd/flow/record.go` (`inRunFixRefusal`, open-row refusal,
  `fixRoundKey`/`fullRunKey`), `hooks/enforce-agent-baseline.py:136`,
  `stats/internal/guard/planshape.go` (F24), `stats/internal/selfreview/git.go:390`.

## Existing code to reuse

- `kwRun` (`stats/internal/guard/kickoffworktree.go:117`) — runs `setup` and `capture` as `prepare`
  already runs `specs`.
- The `--prototype` filtered-sidecar path in `visualverifypass_measure.go:140-155` and
  `mockupsLines` (`:535`) — the `--frames` filter.
- `flDeclared` (`framelist.go:208`) — declared frames for the count and the slice check.
- `restKey` and `suffix` (`visualverifypass_close.go:331-357`) — carry the shard segment.
- `appendStage` (`visualverifypass.go:620`) and the stage file's `row <key> <w>` lines — the new
  `spec`, `slice`, `captured` lines.
- `dispatch()`'s abort-on-refused-begin unwind (`visualverifypass_prepare.go:558-565`) and its
  `NEXT: dispatch <n>` line — unchanged for n shard rows.
- `commit()` (`visualverifypass_close.go:415-479`) — skips an empty commit, so per-shard closes
  need no new commit logic.
- `vrRestSegment` (`verifyreport.go:179-180`) — manifest key derivation keeps the shard segment.

## Verification

- `cd stats && go test ./internal/guard -run 'TestVisualVerifyPass|TestCheckVerifyReport|TestCheckFrameList|TestCheckVisualVerifyDispatched'`, covering:
  - `prepare` on a fixture with 25 frames writes no `slice` line and dispatches one row, output
    identical to today's; with 60 frames writes three `slice` lines of 20 and, once captured,
    dispatches `visual-verify-s1of3`..`-s3of3`; with 130 frames four slices; a two-worktree set
    and a `--round-base` pass never shard;
  - a cut that would split two frames of one screenshot moves forward to keep them in one slice;
    a frame id two specs map lands in one slice with every frame their screenshots join; groups
    that run out before n slices make fewer slices, never an empty one;
  - a sharded pass deletes old `record-s*.md` parts at its start, a `-fix-<k>-full` one
    included;
  - a failing `capture` prints `SHARDS: none — …` and dispatches one unsharded row;
  - a spec with a PNG absent at the merge base prints `MERGE-BASE-SPEC:` and
    `NEXT: merge-base-capture` and begins no row; `--merge-base-failed` puts the failure line in
    every bundle;
  - every bundle, sharded or not, carries the `measure` call, a shard's with `--frames`;
  - `measure --frames` composes only the slice's frames and leaves other frames' `.frame.png`
    untouched;
  - `check-frame-list --frames` and `check-verify-report --frames` accept a slice-only report and
    refuse one missing a slice frame; a sidecar-less PNG is required only of `-s1of<n>`;
  - `close` on `visual-verify-fix-3` blocked prints `redispatch visual-verify-2-fix-3`
    — `flow record`'s `inRunFixRefusal` admits that key while In-run fix 3 is open and refuses
    today's `visual-verify-2` ("is not scoped to In-run fix 3"), and admits
    `visual-verify-fix-3-full-s1of3`, `visual-verify-2-fix-3-full-s2of3` and
    `visual-verify-fix-3-full-rest-s2of3` begun with no round base;
    <!-- measured: a throwaway go test in stats/cmd/flow calling inRunFixRefusal([]string{"3"}, records.Run{}, <key>, <round base>) for each key @ main 0419a52b --> on a shard
    blocked prints `visual-verify-2-s2of3`; the stage ends only once every shard index's newest row
    is closed; `visual-verification.md` is the parts joined in shard order;
  - `check-visual-verify-dispatched` refuses a set with one shard completed and another aborted or
    open, and accepts a set with every index's newest row completed; an incomplete shard set
    followed by a newer completed unsharded row is judged by that row, as today.
- The project's `## lint` commands from `.flow/project.md` exit clean.

## Live check

None — the change edits `flow-guard` subcommands and skill prose: no daemon, store schema or
scheduler changes, the guards keep writing dispatch rows through the unchanged `flow record` CLI,
and a sharded pass needs a project with `mockups` declared, which this repository is not; the Go
tests above run the guards on fixture projects end to end.

## Non-goals

- Sharding a scoped fix round or a multi-worktree set (`shard-unscoped-single-worktree`).
- Aggregating `CLOSES-LOOP` across shards, or changing `record.go`'s `-fix-<k>` refusals
  (`shard-unscoped-single-worktree`).
- A queue of 25-frame slices above 100 frames (`four-shards-slices-grow`).
- A stage-file lock (`closes-serial-no-lock`).
- Moving an unsharded pass's merge-base capture to the parent
  (`merge-base-capture-when-baseline-missing`).
- Sharding a prototype-only change (its F is 0).
- A hook change: `hooks/enforce-agent-baseline.py:136` already admits every shard key.

## Risks

- Shards share one live stack. Frame captures are `prepare`'s, taken before any shard starts, and
  a state check that writes runs on an account of its own (`shard-state-writes-own-account`); on an
  app with no way to register an account, a writing check still runs on the seeded account, where
  another shard's write can disturb its hand captures.
- Comparisons across shards are lost: sweep 1's flow-wide diff and `identical` groups spanning a
  directory are not compared when a flow is split; the contiguous cut in `specs` order limits this.
- A frame with no sidecar line now costs one loop round, since no shard adds sidecar lines.
- A failed capture falls back to one serial verifier; each sharded pass, `-full` ones included,
  may start a merge-base stack again.
- `flow record`'s `inRunFixRefusal` treats one completed `-full` shard as closing the fix;
  `check-visual-verify-dispatched` is the only whole-set check.
- Above 100 frames each shard runs past 25 frames and leans on the `-rest` remainder.
- Accepted: `vvdShardSeg` (Task 5) re-declares Task 3's `vvpShardSeg` regexp in the same `guard`
  package — Task 5 is `After: none` and runs beside Task 3; reading Task 3's `vvpShardSeg` would serialize it behind Task 3, and both regexps quote decision shard-key-segment's one pattern, which Task 5's test pins.
