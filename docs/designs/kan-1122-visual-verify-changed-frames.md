# kan-1122-visual-verify-changed-frames
> Design record of change kan-1122-visual-verify-changed-frames (KAN-1122), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1122 (labels `flow`, `flow-speed`, `flow-tests`). Observed on base `f3e458be`: every fix round
re-verifies the change's whole frame list (`skills/flow/visual-verify-verifier.md`, "**The list is
the change's, never the round's**"), and a clean scoped round is followed by an unscoped
`-fix-<k>-full` pass (`skills/flow/verify-fix-loop.md`, **One full run closes the loop**). Both
are enforced in code: `inRunFixRefusal` (`stats/cmd/flow/record.go`) refuses a new visual-verify
key until a `-fix-<k>-full` row has completed, and `flow-guard check-visual-verify-dispatched`
(`stats/internal/guard/visualverifydispatched.go`) answers `closing-run-missing` while the newest
qualifying verifier row carries a round base. The ticket decided: "changed frames, script-gated,
with the full pass dropped on that condition."

The ticket's note that KAN-1109 makes `measure` compose only new or modified PNGs does not hold in
this repository: `measure` deletes and recomposes every frame of its spec on every call
(`stats/internal/guard/visualverifypass_measure.go`, the sidecar loop), so "the script re-composes
every frame each round" already holds and nothing waits on KAN-1109.

What constrains the design:

- Capture stays the verifier's (its steps 8–9). A Playwright `capture` runs with
  `--update-snapshots=all`, which rewrites every PNG its spec takes, so a re-captured PNG's mtime
  is always this round's (`skills/flow/visual-verify-verifier.md` step 8).
- `prepare` begins the verifier row before the verifier runs, so anything derived from the
  verifier's captures can only be recorded at the row's end, which `close` already makes on
  every path (`endRow`, `stats/internal/guard/visualverifypass_close.go`).
- `measure` is called by the verifier as `measure <worktree> <name> <key> <spec path>`; it knows
  the row's worktree and key, never the canonical worktree the stage file lives in.
- The dispatch ledger is the only state `flow record dispatch begin` reads, so the closing signal
  must live on a dispatch row.

This is one change because the signal, its writer and its two readers only work together: a field
nobody writes, or a guard that accepts a field nobody can set, closes nothing.

Nothing in this change is visible to a user of the stats SPA, so no prototype was built.

## Decisions

### Where the closing signal is recorded

**ID:** closes-loop-on-verifier-row
**Status:** active
**Chosen:** a new `closesLoop` field on the scoped verifier's own dispatch row, set by the
script — operator, round 1 Q1: "A. New field on verifier dispatch row (Recommended)". The script
derives it, flowd keeps it auditable, and both guards already read verifier rows.
**Considered:** B, a `frames-moved: none` line in the stage file or the verifier report that
`check-visual-verify-dispatched` reads from disk — fails because `flow record dispatch begin`
reads only the ledger and has no path to such a file, and the gitignored `.superpowers` file can
be gone when the guard runs. C, `close` recording a completed `-fix-<k>-full` row with no verifier
behind it — fails the rule that a row is never back-filled with instants not read from its
agent's transcript, and every cost and stage stat would count a full run that never ran.

### Where the changed-frame list is computed

**ID:** changed-frames-in-measure
**Status:** active
**Chosen:** in `flow-guard visual-verify-pass measure`, at the verifier's step 9, which prints a
`changed:` line the verifier judges from — operator, round 1 Q2: "A. In `measure`, at verifier
step 9 (Recommended)". Capture stays where it is.
**Considered:** B, in `prepare` before dispatch, with the parent running every capture spec and
the bundle carrying the list — fails because capture output and its failures land in the main
session's context, and capture moves out of the verifier's steps 8–9, a larger rewrite of
`visual-verify-verifier.md` and `prepare`.

### How the script shows every frame was re-captured this round

**ID:** capture-freshness-by-mtime
**Status:** active
**Chosen:** freshness by timestamp, checked in `close` — operator, round 2 Q3: "A. Timestamp
check in close (Recommended)". `prepare` stamps the row with the instant it begins it; a frame is
fresh when its manifest pair's capture PNG has an mtime after that instant. A missing or stale
frame leaves `closesLoop` unset, and the `-fix-<k>-full` run stays. It is derived from files the
round already writes and mutates no tracked file. Its failure mode — a capture tool that rewrites
a PNG only when its bytes changed, keeping the old mtime — reads as stale and fails safe: the
full run stays and only the saving is lost.
**Considered:** B, manifest coverage only — fails under a spec the verifier never re-ran whose
old PNG is still on disk: it composes, reads as unchanged and closes the loop on a frame no one
re-captured after the fix, the defect the full run exists to catch. C, `prepare` deleting every
frame's capture PNG before dispatch — fails because tracked files are deleted in the worktree
before the verifier runs, and a verifier aborted midway leaves the tree dirty with deletions that
the next round's diff, `flow-guard check-task-commit-fields` and every later stage's git status
all see.

### Freshness is read in close, never stored by measure

**ID:** freshness-derived-in-close
**Status:** active
**Chosen:** `close` stats each manifest pair's `capture` path itself and compares its mtime with
the row's instant; `measure` gains no manifest field — the capture path is already in every pair.
**Considered:** `measure` writing a per-pair `fresh` flag into the manifest, as the round-2 Q3
option text sketched — a stored copy of a fact `close` can read directly from the file
(derive, don't store), and one more manifest field for `check-verify-report`'s readers to carry.

### What counts as a changed frame

**ID:** changed-by-blob-hash
**Status:** active
**Chosen:** a frame is changed when its manifest pair's capture PNG, as it stands in the working
tree, differs from the blob at the round base: `git -C <capture's dir> hash-object <capture>`
against `git -C <capture's dir> rev-parse <round base>:./<capture's base name>`. A capture absent
at the round base, or a `rev-parse` that fails (a regression checkout in another repository that
does not hold the sha), is changed. A pair with no capture is changed. One helper serves `measure`
and `close`.
**Considered:** `git diff --name-only <round base> -- <screenshots root>`, as rounds 1–2
sketched — `git diff` lists no untracked file, so a capture this round wrote for the first time
reads as unchanged and is never judged; and the round base's sha does not resolve in a regression
checkout of another repository. Comparing per-frame `measure` figures between manifests — the
figures can stay inside tolerance while pixels move, and a manifest carries no prior-round copy.
Widening `vvp.changed` — its two callers, `trigger` and `roundDiff`, want its `..HEAD` semantics
unchanged.

### How the round base and the row's instant reach measure and close

**ID:** round-file-per-row
**Status:** active
**Chosen:** when `prepare` begins a verifier row with a round base, it writes
`<row worktree>/.superpowers/sdd/visual-round-<key>.txt` holding one line,
`<round base> <RFC 3339 instant with nanoseconds>`, the instant read right before its
`flow record dispatch begin`. `measure` reads it by its `<worktree>` and `<key>`, `close` by the
worktree the stage file's `row <key>` line names. No file means an unscoped row: `measure` prints
no `changed:` line and `close` sets no flag. The verifier passes no new argument.
**Considered:** writing both onto the stage file's `row <key> <worktree>` line, as round 2
sketched — `measure` is passed the row's worktree, never the canonical worktree that holds the
stage file, so a sibling repository's row could not find it; and three parsers
(`close`, `rowWorktree`, `endStage`) read the rest of a `row` line as the worktree path. A new
`measure` flag the verifier must pass — one more argument a verifier can get wrong, while
`measure` already takes the key.

### Which frames close must find fresh

**ID:** closing-frame-list
**Status:** active
**Chosen:** the union of the frames `check-frame-list` declares for the change (when the worktree
declares `mockups`), every frame of `<changeRoot>/visual-verification/prototype.mockups` (when it
exists), and every frame of the key's manifest pairs. An empty union sets no flag. The declared
list is read through a helper extracted from `checkFrameList` (`stats/internal/guard/framelist.go`),
its second caller.
**Considered:** the manifest's pairs alone — a spec the verifier skipped leaves no pair, so its
frames would never be asked for. The declared list alone — a prototype-specified change, this
repository's own case, declares no `mockups` root and would never close early.

### Unjudged frames stay on the frames line

**ID:** unverified-frames-on-frames-line
**Status:** active
**Chosen:** frames not judged this round stay on the report's `frames:` line with the reason
`capture unchanged since <round base>` — the existing "named, with the reason" path of
`visual-verify-verifier.md`, so `check-frame-list` needs no scoped mode.
**Considered:** a scoped mode for `check-frame-list` — a guard change the existing path already
makes unnecessary.

### The full run stays as the fallback

**ID:** full-run-kept-as-fallback
**Status:** active
**Chosen:** the `-fix-<k>-full` key, its refusals and its paragraph stay for every round whose
condition fails.
**Considered:** deleting the full run outright — the ticket drops it only "on that condition".

### Moved frames stay named by the parent

**ID:** moved-frames-kept-in-prompt
**Status:** active
**Chosen:** the parent keeps naming `moved:` frames in the verifier's prompt, and the verifier
judges them beside the `changed:` frames.
**Considered:** deriving them from the changed set — the ticket names them explicitly; a shared
fix's recapture is already inside the changed set, so the two never conflict.

### Who writes the field

**ID:** closes-loop-writer
**Status:** active
**Chosen:** `close` passes `-closes-loop` to the `flow record dispatch end` its `endRow` already
runs, only on the `completed` path (step 5 of its contract).
**Considered:** `begin` — written before the verifier runs. `measure` writing the row — it has no
row access and runs once per spec.

### Close tells the parent whether the full run follows

**ID:** close-prints-closes-loop-line
**Status:** active
**Chosen:** on the `completed` path of a row with a round file, `close` prints
`CLOSES-LOOP: yes` or `CLOSES-LOOP: no — <first failing frame and why>` before `NEXT: done`,
and the stage prose has the parent run the `-fix-<k>-full` run only on `no`.
**Considered:** leaving the parent to learn it from `flow record dispatch begin`'s refusal of the
next round's key — the parent would dispatch or skip the full run on a refusal it hits later,
and the reason a round did not close would be printed nowhere.

### Close derives the condition itself

**ID:** closes-loop-derived-in-close
**Status:** active
**Chosen:** `close` recomputes the changed set through the same helper `measure` uses.
**Considered:** `close` trusting a `changed:` file the verifier or `measure` left on disk — a
stored copy the verifier could edit (derive, don't store).

### The field's shape

**ID:** closes-loop-field-shape
**Status:** active
**Chosen:** `ClosesLoop bool` / JSON `closesLoop` on both `records.Dispatch` and
`records.DispatchEnd`, column `closes_loop boolean NOT NULL DEFAULT false` added by migration
`0066_dispatch_closes_loop.sql`, written by `EndDispatch`'s `UPDATE` and read through
`dispatchColumns` and `scanDispatchRow`; CLI flag `flow record dispatch end -closes-loop`,
refused with any outcome other than `completed`.
**Considered:** a text column — the signal is yes/no, and the round base already sits in
`round_base`. Adding it to the begin `INSERT … ON CONFLICT DO UPDATE` list, as round 2 sketched —
nothing writes the field at begin, and a begin re-run must not clear an end's value.

### A round that moves a frame's mapping keeps the full run

**ID:** frame-map-change-keeps-full-run
**Status:** active
**Chosen:** `close` sets no flag when the `<spec>.mockups` sidecar of a manifest pair's spec
exists and differs from its blob at the round base (the changed-by-blob-hash compare, so an
untracked sidecar differs), or when `git diff --name-only <round base> -- <mockups root>` in the
row's worktree names a path; it prints `CLOSES-LOOP: no — <path> changed since <round base>`. A
frame whose mockup or sidecar line moved while its capture did not is not in the changed set, so
no verifier of the round judges it; the `-fix-<k>-full` run does. Taken at writing-plans: the
capture-only changed set left this case unjudged once the full run is dropped.
**Considered:** adding the frame's mockup to the blob compare of changed-by-blob-hash — a rendered
or prototype frame lives in a gitignored or untracked path, so every such frame would read changed
and no round of a prototype-specified change would close early. `git diff <round base>..HEAD`
alone — it misses a sidecar the verifier writes this round, which is never committed before
`close` reads the diff.

### A prototype frame that changed since the last verifier keeps the full run

**ID:** prototype-change-keeps-full-run
**Status:** active
**Chosen:** every `completed` close of a verifier row, scoped or not, in a change with
`<changeRoot>/visual-verification/prototype.mockups`, writes
`<row worktree>/.superpowers/sdd/visual-prototype.txt` after the row's end: one
`<sha-256 hex> <path relative to visual-verification/>` line for `prototype.mockups`, then one per
`prototype/*.png` in name order. A scoped row's close sets no flag when that map exists and the
file is absent or differs from the same lines computed now, printing `CLOSES-LOOP: no —
<visual-prototype.txt path> differs from prototype.mockups and prototype/*.png now`. The file holds
what the last completed verifier of that worktree judged against, since a verifier renders the
prototype before its close; it is stored rather than derived because the verifier overwrites
`prototype/*.png` on every run, so no earlier render survives to derive it from. A lost write only
keeps the next round's full run, so its error is not fatal. Taken at plan review 2 (F1): in
prototype mode `measure` maps frames through `prototype.mockups` and `prototype/`
(`stats/internal/guard/visualverifypass_measure.go`), which frame-map-change-keeps-full-run's
sidecar and mockups-root compares never read, so a re-rendered prototype PNG beside a
byte-identical capture closed the loop with its frame unjudged.
**Considered:** keeping the full run when `prototype.mockups` or a `prototype/` file has an mtime
after the round base, plan review 2's suggested fix — the verifier renders the prototype and
rewrites the map on every run (`skills/flow/visual-verify-verifier.md`, "An approved prototype is
the frame where no mockup is"), so every prototype file is always newer and no prototype-specified
change would close early, the outcome closing-frame-list rejected. A blob compare against the
round base — `spectre/changes/` is gitignored (`.gitignore`), so every prototype file reads
changed, the same outcome. Recording the gap as an accepted risk — a fix round that corrects the
prototype render (an injected width rule, a re-drawn state) is exactly the round whose capture can
stay byte-identical. A snapshot written by `prepare` — a fixer that re-renders before `prepare`
would land in the snapshot and hide its own change.

### An unjudged view is marked on its own view line

**ID:** unchanged-view-line
**Status:** active
**Chosen:** the verifier writes an unjudged view's line as `- <view>: <absolute PNG path> —
capture unchanged since <round base>`, and `flow-guard check-verify-report` counts such a view as
accounted, with no frame line and no `per view, no frame composed:` line. Taken at writing-plans:
`check-verify-report` refuses a view with neither line (`vrUnaccounted`,
`stats/internal/guard/verifyreport.go`), which unverified-frames-on-frames-line did not cover.
**Considered:** a `per view, no frame composed: <view> sweeps:` line with every sweep `n/a` — it
states that no frame is composed for a view that has one. Keeping the frame's composite line and
exempting it from the over-tolerance checks — the composite line is the per-frame line `close`
counts as judged, so `close` would have to parse its tail to tell judged from not.

### The changed line names its round base

**ID:** changed-line-names-round-base
**Status:** active
**Chosen:** `measure` prints `changed: <frame id> … — round base <sha>` or
`changed: none — round base <sha>`, so the verifier writes `capture unchanged since <round base>`
from what it read, with no new bundle line. Display format, decided at writing-plans.
**Considered:** a `round base:` line in the verifier's bundle — a second place `prepare` writes
the same sha, and the verifier would join two outputs to write one reason.

### A remainder row gets no round file

**ID:** no-round-file-for-rest
**Status:** active
**Chosen:** `prepare` writes the round file only for a key with no `-rest[-<n>]` segment, and
removes `visual-round-<key>.txt` for a row begun with no round base, so a re-dispatch voided to
unscoped never reads a stale one. A remainder's `measure` still reads its base key's file, since
the manifest key drops the segment (`skills/flow/visual-verify.md`), so the remainder judges the
same changed set; its `close` finds no file of its own and prints no `CLOSES-LOOP:` line, which
keeps the full run, as the `-rest` non-goal states.
**Considered:** `close` refusing the flag for a `-rest` key while `prepare` writes its file — a
file nothing acts on.

## Open questions

## Recommended approach

1. **Every round captures every frame.** A fix round's verifier still runs every capture spec the
   change's frame list needs; the rule that specs are chosen to reach every frame on the list
   stays. `measure` recomposes every frame, as it already does.
2. **`prepare` stamps the scoped row** (round-file-per-row): beside `flow record dispatch begin
   … -round-base <sha>`, it writes `visual-round-<key>.txt` with the round base and the instant.
3. **`measure` names the changed frames** (changed-frames-in-measure, changed-by-blob-hash,
   changed-line-names-round-base): with a round file for its key, after writing the manifest it
   prints `changed: <frame> … — round base <sha>` — the frames of this call's pairs whose capture
   differs from the round base — or `changed: none — round base <sha>`, before its `MANIFEST:`
   line. Without one it prints no `changed:` line.
4. **The verifier judges the `changed:` frames plus the `moved:` frames.** Every other frame on
   the change's list is named on the `frames:` line with `capture unchanged since <round base>`
   (unverified-frames-on-frames-line), and its view line carries the same reason
   (unchanged-view-line). A view no frame is composed for is read as today. With no `changed:`
   line it judges every frame, as today.
5. **`close` sets the signal** (closes-loop-writer, closes-loop-derived-in-close): on the
   `completed` path, for a row with a round file, it passes `-closes-loop` when both hold:
   - every frame of the closing frame list (closing-frame-list) has a pair in
     `visual-manifest-<key>.json` whose capture exists with an mtime after the round file's
     instant (capture-freshness-by-mtime, freshness-derived-in-close);
   - every frame those pairs show changed has its own per-frame line in the report — the
     `- <frame id>:` line `check-frame-list` already recognises, not only a mention on the
     `frames:` line;
   - no `<spec>.mockups` sidecar of those pairs and nothing under the `mockups` root changed since
     the round base (frame-map-change-keeps-full-run);
   - with a prototype map, `visual-prototype.txt` matches the map and renders now
     (prototype-change-keeps-full-run).
   It prints `CLOSES-LOOP: yes` or `CLOSES-LOOP: no — <first failing frame and why>` before
   `NEXT: done`, so the parent knows whether the `-fix-<k>-full` run follows. Every `completed`
   close, scoped or not, then rewrites `visual-prototype.txt` when the change has a prototype map.
6. **Both guards accept the signal** (closes-loop-on-verifier-row): `inRunFixRefusal` counts a
   completed visual-verify verifier row with `closesLoop` keyed `-fix-<j>` as closing every fix
   up to `<j>`, beside the `-fix-<j>-full` rows it counts today; `check-visual-verify-dispatched`
   does not answer `closing-run-missing` when the newest qualifying row carries `closesLoop`.
   When the signal is absent, both behave exactly as now (full-run-kept-as-fallback).
7. **Prose**: the loop, the verifier, the stage and the handoff state the new close, the README's
   guard listing names it, and `SKILL-rationale.md` records the reason under the passages it
   changes.

Rejected alternatives and the decision that rejected each: a ledger-free signal or a synthetic
full row (closes-loop-on-verifier-row); parent-side capture (changed-frames-in-measure);
coverage-only or delete-before-dispatch freshness (capture-freshness-by-mtime); a stored fresh
flag (freshness-derived-in-close); `git diff --name-only` (changed-by-blob-hash); the stage-file
row line or a new flag (round-file-per-row); an mtime or blob compare of the prototype files, or a
`prepare` snapshot (prototype-change-keeps-full-run).

## Critical files

- `stats/internal/store/migrations/0066_dispatch_closes_loop.sql` — new: adds `dispatches.closes_loop`.
- `stats/internal/records/types.go` — `Dispatch.ClosesLoop`, `DispatchEnd.ClosesLoop`.
- `stats/internal/store/records.go` — `dispatchColumns`, `scanDispatchRow`, `EndDispatch`'s `UPDATE`.
- `stats/internal/store/records_test.go` — the end-and-read-back test of the field.
- `stats/cmd/flow/record.go` — `runRecordDispatchEnd`'s `-closes-loop` flag and its refusal; `inRunFixRefusal`.
- `stats/cmd/flow/record_test.go` — `inRunFixRefusal` and the flag's tests.
- `stats/internal/guard/visualverifydispatched.go` — the row loader reads `closesLoop`; `closing-run-missing` exempts it.
- `stats/internal/guard/check_visual_verify_dispatched_test.go` — its test.
- `stats/internal/guard/visualverifypass.go` — the header comment that is `visual-verify-pass`'s contract (prepare's round file, measure's `changed:` line, close's `-closes-loop` and `CLOSES-LOOP:` line).
- `stats/internal/guard/visualverifypass_prepare.go` — `dispatch` writes the round file.
- `stats/internal/guard/visualverifypass_measure.go` — prints the `changed:` line.
- `stats/internal/guard/visualverifypass_close.go` — computes the condition, passes `-closes-loop`.
- `stats/internal/guard/framelist.go` — the declared-frames, report-section and per-frame-line helpers extracted from `checkFrameList`; its existing tests (`check_frame_list_test.go`) pin the extraction unchanged.
- `stats/internal/guard/verifyreport.go` — `vrUnaccounted` accounts for a view line marked `capture unchanged since <round base>`.
- `stats/internal/guard/visual_verify_pass_prepare_test.go`, `visual_verify_pass_measure_test.go`, `visual_verify_pass_close_test.go`, `check_verify_report_test.go` — their tests.
- `skills/flow/verify-fix-loop.md` — **The loop**'s step 3 and step 4, **One full run closes the loop**.
- `skills/flow/visual-verify-verifier.md` — the frame-list paragraph, step 9 and the report's view line.
- `skills/flow/visual-verify.md` — the stage's close and its `NEXT: done` handling.
- `skills/flow/verify-and-handoff.md` — the handoff's closing-run passage.
- `skills/flow/SKILL-rationale.md` — the reasons of the passages changed above.
- `README.md` — the pipeline's guard listing.

## Existing code to reuse

- `vvp.readStage`, `vvp.rowWorktree`, `vvp.endRow` (`stats/internal/guard/visualverifypass*.go`) — the stage file, the row's worktree, the row's end.
- `vvp.git` / `vvp.gitChild` — git calls in a given directory.
- `vvmManifestOut`, `vvmPair` (`visualverifypass_measure.go`) — the manifest `close` reads; `capture` and `frame` per pair.
- `cmfMatchCapture` (`composemockupframes.go`) — the sidecar-line-to-PNG match `measure` already uses.
- `checkFrameList`'s declared-list derivation and its per-frame-line regexps (`framelist.go`) — extracted, not rewritten.
- `inRunFixRefusal`'s `completed` predicate and `fullRunKey`/`fixRoundKey` (`stats/cmd/flow/record.go`).
- `vvdVerifier` and its row loader (`visualverifydispatched.go`).
- `0065_pair_judge.sql` and its `pair_pick` plumbing — the template for adding a dispatch column.

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean, and the guard suite of
  `.flow/project.md`'s `## test`, plus `go test ./cmd/flow/... ./internal/store/...`.
- A `visual-verify-pass` scenario test drives prepare → measure → close on a fixture worktree
  with a round base: (a) every spec re-captured, one frame changed and judged → `close` passes
  `-closes-loop` and prints `CLOSES-LOOP: yes`; (b) one spec not re-run (its PNG's mtime before
  the instant) → no flag, `CLOSES-LOOP: no — <frame> …`; (c) a changed frame named only on the
  `frames:` line → no flag; (d) a new untracked capture → `measure` names it under `changed:`;
  (e) a sidecar changed since the round base → no flag; (f) a prototype map whose renders match
  `visual-prototype.txt` → the flag, and one whose render differs from it, or with no such file →
  no flag.
- `scripts/flow-guard check-references` and every `## lint` guard clean over the edited skill files.
- The store and CLI legs run against real state under `## Live check` below.

## Live check

Against a live-store copy, from this worktree: `scripts/flow-guard live-store-copy up kan-1122`
(it builds `flowd` from this checkout, so the copy migrates through `0066`), its `FLOW_ADDR` and
`FLOW_RECORDS_ADDR` exports set, and the `flow` CLI built from this worktree,
`cd stats && go build -o "$(flow state dir kan-1122-visual-verify-changed-frames)/flow" ./cmd/flow`,
put first on `PATH`. The scratch change `kan-1122` gets a plan at
`<worktree>/spectre/changes/kan-1122/tasks.md` holding `- [x] 1. In-run fix 1 — flow.visual-verify: x`;
every begin below carries `-change kan-1122 -role verifier -model opus -effort low -session-token
mf-live-kan1122 -diff-base <HEAD> -agent-id none`, and every end waits 3 s or more after its begin,
the store's minimum window for a subagent row.
<!-- measured: begin, end, `flow record dispatches -change kan-1122` and the refusal below run against `live-store-copy up kan-1122` @ merge-base f3e458be (an end 1.08 s after its begin was refused "dispatch window too short") -->

1. Begin `-key visual-verify-fix-1 -round-base <HEAD>`, then
   `flow record dispatch end -change kan-1122 -key visual-verify-fix-1 -session-token
   mf-live-kan1122 -outcome completed -closes-loop`. Record the row's `closesLoop` from
   `flow record dispatches -change kan-1122`.
2. Begin `-key visual-verify-2` — record its exit code and stderr.
3. `scripts/flow-guard live-store-copy down kan-1122`, `up kan-1122` again, the plan rewritten, and
   steps 1 and 2 repeated with step 1's end carrying no `-closes-loop`: record step 2's exit code
   and stderr.

Before (the merge base's `flow`, built the same way from a detached worktree at the merge base):
step 1's end exits 2 printing `flag provided but not defined: -closes-loop`. After: step 1's row
reads `"closesLoop":true`, step 2 exits 0, and step 3's begin exits 2 printing `-key
"visual-verify-2" is not scoped to In-run fix 1: … until the -fix-1-full run has completed`.
<!-- measured: the before figure and step 3's refusal, against `live-store-copy up kan-1122` with the installed flow @ merge-base f3e458be -->
**Not working looks like** step 2 refused with that same line while step 1's row reads
`closesLoop` true — the field is stored but the refusal ignores it — or step 1's row carrying no
`closesLoop` key, the daemon having dropped the field. `flow-guard
check-visual-verify-dispatched` is not run here: no worktree of this change touches a `ui paths`
entry, so it answers `VISUAL-VERIFY-OK: no UI paths touched` with or without the field; its
acceptance is Task 3's unit test. The check ends with `scripts/flow-guard live-store-copy down
kan-1122` and `rm -rf <worktree>/spectre/changes/kan-1122`.

## Non-goals

- Motions are not narrowed further: a scoped round's `motions:` stays as step 4 scopes it, and a
  motion the round diff does not touch is not re-checked once the full run is dropped — the
  ticket's condition is about frames.
- A `-rest` remainder row is judged against its own key's manifest only; a round split by a
  remainder does not set the signal and keeps its full run.
- The stage-analysis brief (`docs/briefs/flow-stage-analysis.html`) is a dated record and is not
  rewritten.
- The compose-only round is unchanged; it already closes with no `-full` run.

## Risks

- The brief recorded dropping the full run as WON'T-FIX: "41 of 83 closing full runs were
  followed by another fix round — half of them found something the scoped run missed". Under this
  design a frame the scoped round does not judge is byte-identical to one an earlier verifier
  judged — a frame the earlier verifier reported defective included, which the fix round must name for judging (panel round 0, F1) — so what the full run caught there was either a frame the scoped run never
  captured — now closed by capture-freshness-by-mtime and closing-frame-list — or a second look
  at an unchanged capture (judgment variance), which this change gives up. The ticket decided
  the trade.
- A capture tool that keeps a PNG's old mtime when its bytes did not change reads as stale and
  keeps the full run; the saving is lost, never the check.
- A prototype render that is not byte-stable across runs of the same HTML reads as changed and
  keeps the full run (prototype-change-keeps-full-run); the saving is lost, never the check.
- Accepted: plan review 2, F2 — `prepare` writes `visual-round-<key>.txt` before `flow record
  dispatch begin`, so a refused begin leaves a round file for a row that never began — no reader
  runs for a key whose row did not begin, and the next `prepare` of that key rewrites or removes
  the file.
- A daemon built before the field refuses the `closesLoop` payload (`DisallowUnknownFields`) and
  the end falls to the journal; `make -C stats restart` after landing, per `CLAUDE.md`'s reinstall
  rule, rebuilds it.
- mtime is a wall-clock comparison on one machine: the instant and the captures are taken on the
  same host, so clock skew between hosts does not arise.
- KAN-1124's plan also adds a `0066_` migration. Migrations are recorded by filename and applied in
  lexical order (`RunMigrations`, `stats/internal/store/migrations.go`), so two `0066_` files with
  different names both apply; neither change renumbers.
