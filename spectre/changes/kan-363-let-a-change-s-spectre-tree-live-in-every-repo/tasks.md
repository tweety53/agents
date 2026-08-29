# kan-363-let-a-change-s-spectre-tree-live-in-every-repo

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Fourteen tasks across two repositories. Tasks 1–5 are in `spectre` and ship the tree format, the
command and the validation. Task 6 links this change's own two trees with what tasks 1–5 just
built. Tasks 7–11 are in `agents` and make the guards and the contracts read the link. Tasks 12 and
13 close a peer-resolution defect task 6's real run exposed — 12 in `spectre`, 13 in `agents`.
Task 14 settles a boundary collision the review panel's own commits exposed, in `agents`.
`design.md` is canonical for the decisions; `proposal.md` for the KAN-343 measurements behind them.

**Every task states its own `**Repository:**`, and every `**Files:**` path is relative to that
repository's root.** The convention is KAN-343's, which is the change this one exists because of.

**Baseline, measured before any edit:**

- `spectre`: 145 top-level `func Test…` across the module.
  <!-- measured: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l @ 42d69c6 -->
- `spectre`: `go test ./...` clean, 7 packages with tests, 2 with none.
  <!-- measured: go test ./... -count=1 @ 42d69c6 -->
- `agents`: 42 harnesses matched by `scripts/test-*.sh`.
  <!-- measured: ls scripts/test-*.sh | wc -l @ 00344a6 -->
- Neither repository has a `spectre/peers` file; no tree in either declares a peer.
  <!-- measured: ls ~/Projects/spectre/spectre/peers ~/Projects/agents/spectre/peers @ 42d69c6 and 00344a6 -->
- `internal/check/check.go`'s `Structural` already receives `map[string]tree.ResolvedPeer`, so the
  link checks need no new plumbing to reach peer resolution.
  <!-- measured: grep -n 'func Structural' internal/check/check.go @ 42d69c6 -->

**Global constraints, both repositories:**

- `spectre` targets the Go version pinned in `go.mod` (`go 1.26.5`) and has **no external
  dependencies** — `go.mod` carries no `require` block, and this change adds none.
- `spectre`'s exit codes are a contract: `0` success, `1` findings or a content refusal, `2` usage
  or IO error. `spectre link`'s refusals are exit 1; a malformed invocation is exit 2.
- `spectre`'s checks must all be clean before a task is done: `gofmt -l .`, `go vet ./...`,
  `go test ./...`.
- `agents`' guard scripts target bash 3.2 — indexed arrays only, no associative arrays, no
  `wait -n` — and every guard exports `LC_ALL=C` before any `case` range or `sort`.
- Per KAN-197, every guard assertion added here carries a mutation test: prove the assertion fails
  when the behaviour it checks is broken.

---

- [x] 1. `internal/parse/link.go` — the `link.md` grammar

**Repository:** `/Users/tweety53/Projects/spectre`

The parser and its model type for the five sections, per `design.md`'s `no-free-text-in-link-md`.
Parsing only — no peer resolution, no filesystem access, no findings.

`model.Link` carries: `PartOf` (a `model.LinkRef` or the zero value), `Parts` (`[]model.LinkRef`),
`Branch` (string), `MergeOrder` (`[]string`, where `"."` names this tree) and `TasksHere`
(`[]int`, expanded from the written ranges). `model.LinkRef` is `{Peer, ChangeID string}`.

The grammar, strict in every section:

```go unverified:confirm against internal/parse/spec.go's existing section-splitting idiom before writing
// Section headings recognised, each at most once; a repeated heading is a parse error
// naming both lines, the way peers already errors on a repeated name.
"## Part of"     // exactly one `<peer>:<change-id>` in a code span
"## Parts"       // one `<peer>:<change-id>` code span per line, at least one
"## Branch"      // exactly one non-empty line
"## Merge order" // a numbered list; each item `.` or a peer name, in a code span
"## Tasks here"  // one line: comma-separated ints and `N-M` ranges
```

- `<peer>` must be non-empty and must match the same character rule spectre already applies to a
  peers name; `<change-id>` must satisfy `isValidID`.
- `## Branch`'s line must match the git ref-name shape `<agents repo>/scripts/resolve-base-branch.sh`
  enforces: the first character in `[A-Za-z0-9._]`, every character in `[A-Za-z0-9._/-]`. Copy the
  rule, not the script — `spectre` has no dependency on `agents`.
- `## Tasks here` must be ascending overall, every range ascending, no overlaps, every number a
  positive integer. `4-12` and `1,3,7-9` parse; `12-4`, `3,1` and `2-5,4` are errors naming the
  offending item.
- A file carrying both `## Part of` and `## Parts` is an error: links are one level deep.
- A file carrying `## Parts` or `## Merge order` alongside `## Part of`, or `## Tasks here`
  alongside `## Parts`, is an error naming the section that does not belong on that side.

Write `internal/parse/link_test.go` **first, RED before GREEN**, table-driven per this repository's
`go-test-table-driven` skill, with a case per rule above and a case per rejection.

**Files:** `internal/parse/link.go`, `internal/parse/link_test.go`, `internal/model/model.go`
**Tests:** `TestParseLink`, `TestParseLinkRejects`
**Regression:** reverting this commit removes the only definition of the `link.md` grammar, so every
later task's checks have no shape to check against and `spectre link` has nothing to write.
**Baseline:** before=145 after=147 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 1 -->
**Commit:** `feat(parse): parse the link.md grammar`
**Build:** green

- [x] 2. `internal/check` — the link findings

**Repository:** `/Users/tweety53/Projects/spectre`

Wire the link into `Structural`, which already receives `map[string]tree.ResolvedPeer`.

Behaviour, per `design.md`'s `pointer-tree-not-full-tree`, `independent-archive` and
`peer-absence-is-not-a-finding`:

- A change directory holding `link.md` **and nothing else** is a satellite: skip
  `ProposalFindings`, `TaskFindings` and `DesignFindings` entirely for it. A directory holding
  `link.md` *and* a `proposal.md` or `tasks.md` keeps every existing heading check and additionally
  gets the link checks.
- The peer named by `## Part of` or by a `## Parts` entry must be **declared** in `peers`. Not
  declared is a finding.
- When the peer is declared **and present**, the counterpart change must exist there — searched in
  `changes/<id>/` and then `changes/archive/<id>/` — and its `link.md` must name this change back.
  A one-sided link is a finding naming both sides.
- When the peer is declared and **not present**, report nothing. Not a finding, per
  `peer-absence-is-not-a-finding`.
- Archive skew — the counterpart resolved under `changes/archive/` while this side is still under
  `changes/` — is a finding stating which side is archived. Never a refusal.
- `## Merge order` must name `.` exactly once and every entry of `## Parts` exactly once, and
  nothing else. Missing, duplicated and unknown entries are three distinct findings.
- Both sides' `## Branch` must be byte-identical when the peer is present; disagreement is a finding
  quoting both.
- Every number in `## Tasks here` must match a task id in the canonical `tasks.md` when the peer is
  present; a number matching none is a finding naming it.

Write `internal/check/link_test.go` **first, RED before GREEN**, building two- and three-tree
fixtures with `internal/testtree`. Cover each finding above and, separately, the cases that must
produce **no** finding: peer declared but absent, canonical archived with the part archived too, a
satellite directory that is link-only.

**Files:** `internal/check/link.go`, `internal/check/link_test.go`, `internal/check/check.go`,
`internal/testtree/testtree.go`
**Tests:** `TestLinkFindings`, `TestLinkFindingsQuiet`, `TestStructuralSkipsHeadingsForSatellite`
**Regression:** reverting this commit makes `spectre validate` accept a one-sided link, a merge
order naming a part that does not exist, and a satellite whose `## Tasks here` cites tasks the
canonical plan does not have — every disagreement this change exists to surface.
**Baseline:** before=147 after=150 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 2 -->
**Commit:** `feat(check): report link resolution, agreement and merge-order findings`
**Build:** green

- [x] 3. `spectre link` — the guarded two-sided write

**Repository:** `/Users/tweety53/Projects/spectre`

```text verified:the shape agreed during brainstorming; matches internal/cmd/cmd.go's flagSet convention
spectre link [--root <path>] [--force] <peer>:<canonical-id>
```

Run in the satellite tree. Per `design.md`'s `one-command-writes-both-sides` and
`guarded-cross-tree-write`, it writes the **peer's side first**, then its own, so a refused peer
write leaves nothing half-created.

Refusals, each exit 1 and each naming the check that failed:

- the peer is not declared in `peers`, or is declared and not present;
- the canonical change id does not exist in the peer tree;
- the canonical change is itself a satellite (it carries `## Part of`) — links are one level deep;
- this link already exists on either side;
- the peer's own change directory has uncommitted modifications.

`--force` overrides the **last** refusal only, and no other. A malformed invocation — a missing
argument, an argument with no `:`, an unparseable `--root` — is exit 2.

**A sixth refusal, found while implementing:** the peer must have a declared name for *this* tree in
its own `peers` file. Without one there is no name to write into the peer's `## Parts` entry, so the
entry would be unresolvable the moment it was written. Exit 1, naming the check, like the other
five — and covered by its own test, since a refusal nothing exercises is a refusal nobody knows
still fires.

**The satellite's change id is the canonical id**, not a separate argument. Task 6's worked example
is the statement of this: `spectre link agents:<id>` creates the satellite directory under that same
`<id>`.

What it writes: the peer's `link.md` gains `## Parts` and `## Merge order` (seeded as `.` then the
new part) and, if absent, `## Branch` read from the satellite's current branch; the satellite's
`link.md` gains `## Part of` and the same `## Branch`. Re-ordering `## Merge order` is a hand edit
afterwards — the command never reorders an existing list.

Register the command in `cmd/spectre/main.go` alongside the existing six.

**Also fix `validate`'s peer resolution, found while implementing task 2.**
`internal/cmd/validate.go` resolves `peers` only when `changeID == ""`, on the stated grounds that
"ref checking only runs for a whole-tree validate". Task 2's link checks consult `peers` for a
single change too, so that premise no longer holds and `spectre validate <change-id>` runs every
link check against a nil peer map — silently reporting nothing. This is the form `/flow` itself
invokes (`spectre validate "<name>"`, implement step 1), so the checks would never fire in the
pipeline that needs them. Resolve peers unconditionally and correct the comment. Cover it with a
test that a single-change validate reports a link finding the whole-tree validate also reports.

Write `internal/cmd/link_test.go` **first, RED before GREEN**, one case per refusal, one for
`--force` overriding exactly the dirty-tree refusal and not the others, one asserting the peer side
is written before the local side (make the local write fail and assert the peer side landed), and
one asserting a clean two-tree link round-trips through task 2's checks with no findings.

**Files:** `internal/cmd/link.go`, `internal/cmd/link_test.go`, `cmd/spectre/main.go`,
`internal/cmd/validate.go`, `internal/cmd/validate_test.go`
**Tests:** `TestLink`, `TestLinkRefuses`, `TestLinkForce`, `TestLinkWriteOrder`,
`TestValidateSingleChangeResolvesPeers`
**Regression:** reverting this commit leaves the `link.md` format with nothing that writes it, so
every link would be hand-authored and the two sides could disagree from the moment they were
created.
**Baseline:** before=150 after=155 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 3 -->
**Commit:** `feat(cmd): add the guarded two-sided spectre link command`
**Build:** green

- [x] 4. `spectre list` renders a satellite

**Repository:** `/Users/tweety53/Projects/spectre`

A satellite change has no `tasks.md`, so `list` shows it today — if at all — with no progress. Per
`design.md`'s `two-way-link` it renders as the change it is part of, with the canonical's progress:

```text unverified:confirm against internal/cmd/list.go's existing column alignment before writing
kan-343-route-c-...  ->  gymie  7/12
```

When the peer is declared and not present, the progress column reads `-` and the row still renders
— `peer-absence-is-not-a-finding` applies to `list` as much as to `validate`. `--json` gains
`"partOf": {"peer": ..., "changeId": ...}` on a satellite row and `"parts": [...]` on a canonical
row; a change with no link carries neither key.

Write the `list` cases **first, RED before GREEN**, in `internal/cmd/list_test.go` beside the
existing ones: a satellite with the peer present, the same with the peer absent, a canonical row
carrying `parts`, and the `--json` shape for each.

**Files:** `internal/cmd/list.go`, `internal/cmd/list_test.go`
**Tests:** `TestListSatellite`, `TestListSatellitePeerAbsent`, `TestListJSONLinks`
**Regression:** reverting this commit makes a satellite change indistinguishable in `list` from a
change whose plan was never written — which is the same confusion in the CLI that
`check-unfinished-work.sh` was making in the gate.
**Baseline:** before=154 after=157 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 4 -->
**Commit:** `feat(cmd): show a linked change and its canonical progress in list`
**Build:** green

- [x] 5. Document the link

**Repository:** `/Users/tweety53/Projects/spectre`

`docs/links.md`, new, canonical for the `link.md` format: the five sections and their grammar in
full, the satellite/canonical distinction, what `validate` checks and what it deliberately does not
(peer absent), the archive-skew rule, and `spectre link`'s five refusals with `--force`'s single
scope.

`README.md` gains: `link.md` in the tree diagram, a `link` row in the Commands table, a `link.md`
row in File templates, and a `## Links across repositories` section pointing at `docs/links.md` in
the same shape as the existing `## References across trees` pointer.

No prose in either file restates what the other states canonically.

**Files:** `docs/links.md`, `README.md`
**Tests:** **none** — documentation only; the behaviour it describes is covered by tasks 1–4.
**Regression:** reverting this commit leaves a tree format and a command that only this change's own
`design.md` explains, in a repository whose README is the entry point for every other user.
**Baseline:** before=157 after=157 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 5 -->
**Commit:** `docs(links): document link.md and spectre link`
**Build:** green

- [x] 6. Link this change's own two trees

**Repository:** `/Users/tweety53/Projects/agents`

Per `design.md`'s `dogfood-once-green`. Reinstall the binary from the branch first
(`go build -o ~/go/bin/spectre ./cmd/spectre` in the spectre worktree) so `spectre link` is the
version tasks 1–5 just built, and say in the commit body which commit it was built from.

Declare the peers — neither repository has a `peers` file today:

```text unverified:confirm both paths resolve from each tree's parent directory before committing
# agents/spectre/peers
spectre ../spectre

# spectre/spectre/peers
agents ../agents
```

Then, from the **spectre** worktree, run
`spectre link agents:kan-363-let-a-change-s-spectre-tree-live-in-every-repo`. `agents` is canonical
— the plan lives here — so the spectre repo gains a link-only change directory and this tree gains
`## Parts`, `## Merge order` and `## Branch`.

Set `## Merge order` to `spectre` then `.`: the guards in `agents` call a binary that must already
carry the feature, so spectre lands first. Set `## Tasks here` to `1-5` on the satellite.

Then run `spectre validate` in **both** trees and confirm no findings — this is the end-to-end proof
the decision asked for, and the first time the two sides are checked against each other for real.

**Files:** `spectre/peers`,
`spectre/changes/kan-363-let-a-change-s-spectre-tree-live-in-every-repo/link.md`
**Allowed-collateral:** `spectre/changes/kan-363-let-a-change-s-spectre-tree-live-in-every-repo/*`
**Tests:** **none** — this task's verification is `spectre validate` reporting no findings in both
trees, which is a run rather than a test this repository keeps.
**Regression:** reverting this commit unlinks this change's own trees, so tasks 7–10's guards would
first meet a real link on someone else's change rather than on the one that built them.
**Baseline:** before=42 after=42 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 6 -->
**Commit:** `chore(spectre): declare the peer and link this change's two trees`
**Build:** green

- [x] 7. `scripts/lib/change-plan.sh` — resolve a change's canonical plan

**Repository:** `/Users/tweety53/Projects/agents`

One owner for "where is this change's `tasks.md`, really", sourced by tasks 8 and 9 the way
`scripts/lib/spec-root.sh` is sourced by seven guards today.

```bash unverified:confirm the signature against how check-unfinished-work.sh composes CHANGES before writing
# change_plan_path <worktree> <change-name> [canonical-worktree]
#
# Prints the absolute path of the change's tasks.md and returns 0.
# Returns 1 when the change is a satellite and no canonical plan could be
# reached — the caller decides whether that is a refusal or a verdict.
```

Resolution order, per `design.md`'s `guards-take-the-canonical-worktree-path`:

1. `<worktree>/<spec-root>/changes/<name>/tasks.md` exists → print it.
2. Otherwise, if `<worktree>/<spec-root>/changes/<name>/link.md` exists and carries `## Part of`:
   - when `<canonical-worktree>` was passed, print
     `<canonical-worktree>/<spec-root>/changes/<canonical-id>/tasks.md` if it exists;
   - otherwise resolve the peer name through `<worktree>/<spec-root>/peers` — relative to the
     tree's parent directory — and print the plan there if it exists.
3. Neither → return 1.

Containment applies to every name concatenated into a path here: the change name, the peer name and
the canonical change id all arrive from a pull-request-editable file. Apply
`check-unfinished-work.sh`'s own allowlist — start with a letter or digit, then only letters,
digits, `.`, `_` and `-` — to each of the three, character for character, and `export LC_ALL=C` so
the accepted set does not move with the operator's locale.

Write `scripts/test-lib-change-plan.sh` **first, RED before GREEN**: a plain change, a satellite
with the canonical worktree passed, a satellite resolving through `peers`, a satellite whose peer is
absent, a `link.md` with no `## Part of`, and one rejected name per containment rule. Per KAN-197,
prove each assertion fails when the behaviour is broken.

`.flow/project.md` needs no edit: its `## test` section states that `scripts/run-guard-tests.sh`
discovers every `scripts/test-*.sh` by glob and picks a new harness up automatically, "with no
edit needed here". Name the new caller in `scripts/lib/spec-root.sh`'s header instead, which
already lists the guards that source it.

**Files:** `scripts/lib/change-plan.sh`, `scripts/test-lib-change-plan.sh`,
`scripts/lib/spec-root.sh`
**Tests:** `scripts/test-lib-change-plan.sh`
**Regression:** reverting this commit removes the only resolver that reaches a canonical plan from a
satellite worktree, so both guards below fall back to reporting an absence as a verdict — the exact
KAN-343 failure.
**Baseline:** before=42 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 7 -->
**Commit:** `feat(scripts): resolve a change's canonical plan through its link`
**Build:** green

- [x] 8. `check-unfinished-work.sh` follows the link

**Repository:** `/Users/tweety53/Projects/agents`

Take an optional third-position canonical worktree argument — the guard's existing positionals are
`<worktree> <change-name>`, so the new one is `$3`, matching task 7's own
`change_plan_path <worktree> <change-name> [canonical-worktree]` signature — and source task 7's
resolver in place
of composing `PRIMARY_PLAN` directly.

Changed behaviour, and nothing else:

- A satellite worktree whose canonical plan resolves counts its unticked items against **that**
  plan, and its verdict line still names the worktree it judged. It no longer adds
  `no plan at <path>`.
- A satellite whose canonical plan cannot be reached is exit 2, `cannot determine anything`, naming
  which resolution step failed — never `OUTSTANDING`, and never `CLEAR`.
- A plain change with no `link.md` behaves byte-identically to today, including the fix-sub-change
  sweep and the findings signal.

The header gains a paragraph on why a satellite's missing plan is not the missing-record case the
existing header rejects: there the absence is evidence, here the plan is somewhere else and reached.

**The fix-sub-change sweep follows the plan, not the worktree.** The sweep globs every `tasks.md`
under the worktree's own `changes/`, which for a satellite is the wrong tree: a canonical
`<canonical-id>-fix-N` sibling lives beside the canonical change in the **canonical** repository and
is never swept, so a satellite reports `CLEAR` while the canonical change carries an unchecked
fix-round item. That is the silent clearance this guard's own header exists to prevent. Sweep the
directory the plan actually resolved from: `scripts/lib/change-plan.sh` gains a companion that
returns the canonical change **directory** alongside the plan path, and the sweep runs there.
Cover it with a fixture whose canonical tree carries an unchecked `-fix-1` sibling — asserting
`OUTSTANDING`, not `CLEAR`.

Extend `scripts/test-check-unfinished-work.sh` with the satellite cases and their mutation tests.

**Files:** `scripts/check-unfinished-work.sh`, `scripts/test-check-unfinished-work.sh`,
`scripts/lib/change-plan.sh`
**Tests:** `scripts/test-check-unfinished-work.sh`
**Regression:** reverting this commit restores `OUTSTANDING: <worktree> — no plan at …` for every
secondary worktree, which is what needed a human to wave the KAN-343 gate through.
**Baseline:** before=43 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 8 -->
**Commit:** `fix(scripts): count a satellite worktree against its canonical plan`
**Build:** green

- [x] 9. `check-task-commit-fields.sh` follows the link

**Repository:** `/Users/tweety53/Projects/agents`

Its wrapper resolves *which* `tasks.md` the named task lives in, and today refuses when the glob
matches more than one root change or none. Two changes:

- Take an optional canonical-worktree argument, passed straight to task 7's resolver, and use the
  resolver's answer when the worktree's own glob finds no `tasks.md` — the "no `tasks.md` found
  under `<dir>`" exit 2 becomes a resolution attempt first.
- A change directory that is link-only is **not** counted toward the root-change ambiguity test. A
  satellite is not a second root change, exactly as a `<name>-fix-N` sibling is not.

The Python guard is untouched: it already takes a resolved `tasks.md`, and the wrapper's whole job
is resolving one.

**Also add the symlink sentence task 7's review asked for.** `scripts/lib/change-plan.sh`'s header
restates `check-unfinished-work.sh`'s containment reasoning but is silent where that guard's own
header is explicit: `-f`/`-d` follow symlinks, so a symlink at
`<worktree>/<spec-root>/changes/<allowlisted-name>` is followed and its content returned as the
change's plan. That is the same accepted tradeoff the sibling guard already ships and argues — not
a hole this change opens — and the header should say so in one sentence rather than leave the next
reader to rediscover it. Header comment only; no behaviour changes.

Extend `scripts/test-check-task-commit-fields.sh` with a satellite worktree resolving to the
canonical plan, a satellite alongside a root change (which must not read as ambiguity), and the
unresolvable case. Mutation tests per KAN-197.

**Files:** `scripts/check-task-commit-fields.sh`, `scripts/test-check-task-commit-fields.sh`,
`scripts/lib/change-plan.sh`
**Tests:** `scripts/test-check-task-commit-fields.sh`
**Regression:** reverting this commit returns the guard to exiting 2 without a verdict in both
repositories of a cross-repo change — the state in which all 12 of KAN-343's tasks had their
`Files:`/`Tests:`/`Commit:` fields checked by hand.
**Baseline:** before=43 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 9 -->
**Commit:** `fix(scripts): resolve a linked change's plan in the commit-fields guard`
**Build:** green

- [x] 10. `check-cleanup-complete.sh` sees a satellite's `link.md`

**Repository:** `/Users/tweety53/Projects/agents`

A satellite's change directory is a change artifact like any other and must be gone after archive.

**This task's original premise was measured and found false, so it carries no production change.**
The plan asserted that the guard's change-artifact check looks for the files a `spectre new` scaffold
leaves, and that a link-only directory would therefore report `COMPLETE`. It does not: row four
(`scripts/check-cleanup-complete.sh:465`) is a plain `[ -d ... ]` existence test, content-agnostic,
so it already reports `LEFTOVER` for a link-only leftover and `COMPLETE` once archived.

What was missing is coverage. That shape was pinned only by coincidence, through an existing case
whose fixture happens to be a scaffold-shaped empty directory. Add cases naming it directly, in the
guard's own `COMPLETE`/`LEFTOVER` vocabulary, and mutation-prove them by changing row four to the
scaffold-file check the plan assumed — which must flip them.

Extend `scripts/test-check-cleanup-complete.sh` with a leftover satellite directory, an archived one
(which must read `COMPLETE`), and the mutation test for each.

**Files:** `scripts/test-check-cleanup-complete.sh`
**Tests:** `scripts/test-check-cleanup-complete.sh`
**Regression:** reverting this commit removes the only test naming a link-only leftover directly, so
a later narrowing of row four to a scaffold-file check would let a satellite directory survive archive
unreported — the second repository keeping a dangling link to a change that no longer exists there.
**Baseline:** before=43 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 10 -->
**Commit:** `fix(scripts): report a leftover satellite change directory`
**Build:** green

- [x] 11. The flow contracts create the link and land the repos in order

**Repository:** `/Users/tweety53/Projects/agents`

Four edits, each in the file canonical for what it states, and none restating another:

- **`skills/flow/implement.md`**, `flow.isolate-workspace`: after creating each worktree beyond the
  canonical one, run `spectre link <canonical-peer>:<name>` in it, and record what was written
  alongside the merge base in the run's working notes. A failure is reported and the run continues —
  the link is not a gate, and a change with one worktree runs nothing.
- **`skills/flow-contracts/finish-contract.md`**, run 1: the routes iterate the resolved worktree
  set **in the canonical `link.md`'s `## Merge order`**, not in map order, stopping on the first
  failure with that repository's own output. State that a change with no `link.md` has one worktree
  and one trivially-ordered route, so nothing about the single-repository path changes.
- **`skills/flow-contracts/state-file.md`**, under **A change spanning repositories is one
  record**: one sentence saying the *ordering* of those repositories lives in the canonical
  `link.md`, not in `worktrees`, and citing the finish contract for what reads it.
- **`skills/flow/SKILL.md`**'s guard list and the **Guard presence check** union: add
  `change-plan.sh` as a `lib/` sibling reached by tasks 8 and 9, per the sibling-dependency rule
  `check-guard-symlinks.sh`'s rule 2 already applies.

Then run the repository's own reference and vocabulary guards, and `./setup.sh global`, so the
installed copies under `~/.claude/skills/` carry the new lib. That script is **run**, not edited —
it is deliberately absent from this task's `**Files:**` for that reason.

**Files:** `skills/flow/implement.md`, `skills/flow-contracts/finish-contract.md`,
`skills/flow-contracts/state-file.md`, `skills/flow/SKILL.md`
**Tests:** **none** — contract prose; the behaviour it directs is covered by tasks 7–10's harnesses,
and the citations it adds are checked by `scripts/check-references.sh` and
`scripts/check-guard-symlinks.sh`, which must both exit clean before this task is done.
**Regression:** reverting this commit leaves the guards able to follow a link that nothing ever
creates, and leaves run 1 landing a multi-repository change in map order — the hand-ordered merge
KAN-343 needed.
**Baseline:** before=43 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 11 -->
**Commit:** `feat(flow-contracts): create the link and land the repos in merge order`
**Build:** green

- [x] 12. `Peers()` resolves a peer by probing for a tree, not by comparing a basename

**Repository:** `/Users/tweety53/Projects/spectre`

Task 6's real run exposed this, and `internal/tree/tree.go`'s `Peers()` is where it lives:

```go verified:read from internal/tree/tree.go at c987d80, and reproduced on disk by task 6's reviewer
if filepath.Base(p) != "spectre" {
    p = filepath.Join(p, "spectre")
}
```

A `peers` entry names the **repository**, and `Peers()` appends the tree leaf to reach the tree. When
the repository's own directory is named `spectre`, the basename test sees `spectre`, skips the
append, and resolves one level short — at the repository root. `ResolvePeer` then reports `found`,
because a missing `config.md` is a legal default, while `ChangesDir()` points at a directory that
does not exist. Every change in the real peer tree is invisible and `validate` reports a spurious
one-sided link. Reproduced on disk: a marker change in the real tree could not be seen through the
mis-resolved path.

**Probe instead of guessing.** Append the tree leaf only when the given path is not already a tree —
decide by looking for `changes/` inside it, which is what `spectre init` actually creates and what
every caller goes on to read. `<agents repo>/scripts/lib/spec-root.sh` resolves the same question the
same way and its header states why probing `changes/` beats probing the tree root: a bare directory
can exist for reasons unrelated to a tree, while `changes/` is evidence a caller can use.

Both `../spectre` and `../spectre/spectre` must resolve correctly afterwards — the second already
points straight at a tree, so it needs no append and must not gain one.

Write the tests **first, RED before GREEN**, in `internal/tree/peer_test.go`: a peer repository
directory named `spectre` addressed as `../spectre`; the same addressed as `../spectre/spectre`; an
ordinary peer needing the append; and a declared path that is neither, which must still resolve to
`PeerNotPresent` rather than silently succeeding.

Update `docs/references.md` where it states how a `peers` path resolves.

**Files:** `internal/tree/tree.go`, `internal/tree/peer_test.go`, `docs/references.md`
**Tests:** `TestPeersResolvesRepoNamedLikeTheTree`, `TestPeersResolvesPathAlreadyATree`
**Regression:** reverting this commit returns `spectre` to being undeclarable as a peer under its own
documented form, so the one repository most likely to use links on itself resolves to a tree that is
not there and validates against nothing.
**Baseline:** before=165 after=167 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 12 -->
**Commit:** `fix(tree): resolve a peer by probing for a tree`
**Build:** green

- [x] 13. Write the documented peer form, now that it resolves

**Repository:** `/Users/tweety53/Projects/agents`

Task 6 had to write `spectre ../spectre/spectre` into `spectre/peers` to work around task 12's
defect. With that fixed, write the documented form — `spectre ../spectre` — and prove it resolves by
running `spectre validate` in both trees against a binary built from task 12.

The satellite's `## Tasks here` also moves: tasks 1–5 **and 12** now live in `spectre`, so it becomes
`1-5,12`. `validate` checks every number there against a real task in the canonical `tasks.md`, so a
stale value is a finding — which is the mechanism this change built, catching this change's own
drift.

**Files:** `spectre/peers`
**Allowed-collateral:** `spectre/changes/kan-363-let-a-change-s-spectre-tree-live-in-every-repo/link.md`
**Tests:** **none** — the verification is `spectre validate` reporting no findings in both trees,
which is a run rather than a test this repository keeps.
**Regression:** reverting this commit restores a peer path that only resolves because of the very
defect task 12 removes, leaving the documented form broken for the next reader who copies it.
**Baseline:** before=43 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 13 -->
**Commit:** `chore(spectre): declare the peer in its documented form`
The satellite's `link.md` edit lands in the `spectre` repository under its own accurate subject,
`chore(spectre): tasks here now includes 12` — this task spans two repositories and each commit
names what it actually did, rather than one subject being reused where it would be false.
**Build:** green

- [x] 14. `link.md` is implementation, not planning

**Repository:** `/Users/tweety53/Projects/agents`

**The collision, found by this change's own commits.** `skills/flow-contracts/git-boundaries.md`
states that within the spectre tree the boundary is the **directory** — `<project>/spectre/changes/`
is planning, `<project>/spectre/specs/` is not — so nothing under `changes/` may enter a task commit,
and the integrate phase commits it instead. But `link.md` lives at
`<spec-root>/changes/<id>/link.md`, and in a satellite repository it is the **only** content the
change has. Under the rule as written, a satellite's branch carries no implementation commit at all,
and the file that says what the branch is part of arrives only at integration — after the moment
someone cloning that repository alone would want it.

**Carve it out by file, on the reasoning that already exempts `specs/`.** A capability spec is
implementation because it states what the system must do. `link.md` is implementation for the same
kind of reason: it states which repositories this change spans, which branch carries it, and in what
order they land — structural fact about the change, not the planning narrative `proposal.md`,
`design.md` and `tasks.md` carry. The boundary stops being purely directory-shaped, and the file that
makes it so is named.

**`commit-split.sh` needs a real change, not just prose.** It clears the planning paths from the
index and then stages with them excluded:

```bash verified:read from scripts/commit-split.sh at 1841d12
git -C "$worktree" reset -q -- "$plan_dir" docs/superpowers/
git -C "$worktree" add -A -- . ":(exclude)$plan_dir" ':(exclude)docs/superpowers/'
```

**A git `:(exclude)` pathspec always wins over a positive one**, so `link.md` cannot be re-included by
adding a pattern alongside the exclusion — it needs its own `add` after the excluded one, and the
`reset` must not strip it back out. Establish that behaviour by running git rather than reasoning
about it, and say what you found.

Extend `scripts/test-commit-split.sh`: a worktree whose change directory holds `link.md` alongside
`proposal.md`/`design.md`/`tasks.md` must put `link.md` in the **implementation** commit and the
other three in the planning commit. Per KAN-197, mutation-prove it — drop the new `add` and confirm
the case fails.

**Do not restate the rule in a second file.** `git-boundaries.md` is canonical for it; **Handoff
output** (`skills/flow-contracts/pipeline.md`) names the planning paths and must stay consistent
without carrying a second statement of the carve-out.

**Files:** `skills/flow-contracts/git-boundaries.md`, `scripts/commit-split.sh`,
`scripts/test-commit-split.sh`, `skills/flow-contracts/pipeline.md`,
`scripts/check-contract-budget.sh`
**Allowed-collateral:** `scripts/check-contract-budget.sh` — `git-boundaries.md` outgrows its
budget row, and raising that row is the mechanism the guard's own message prescribes for
deliberate growth ("trim it back under budget, or raise its row in `budgets()`"), never a
weakening of the check
**Tests:** `scripts/test-commit-split.sh`
**Regression:** reverting this commit puts `link.md` back under the planning exclusion, so a
satellite repository's branch carries no implementation commit and its link reaches the tree only at
integration.
**Baseline:** before=43 after=43 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 14 -->
**Commit:** `fix(flow-contracts): treat link.md as implementation, not planning`
**Build:** green
