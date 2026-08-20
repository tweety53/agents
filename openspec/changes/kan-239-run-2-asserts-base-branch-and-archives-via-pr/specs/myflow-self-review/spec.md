## MODIFIED Requirements

### Requirement: The report is committed to a fixed path

Self-review SHALL write `docs/self-review/<name>-self-review.md`, containing the five-section report,
the rating, and which findings were filed versus declined, then commit it on the archive branch
`chore/archive-<name>` in the main checkout — the branch run 2's positioning step put that checkout
on, and the same branch the archive commit already sits on.

Self-review SHALL NOT push. The push is run 2's final step, which carries the archive commit and this
report in the same pull request.

The commit SHALL assert the branch rather than assume it: naming the directory with
`git -C <main-checkout>` fixes the directory and not the branch, which is what stranded the reports
for KAN-197 and KAN-200 on an already-merged archive branch.

#### Scenario: Report path is fixed per change

- **WHEN** self-review completes for change `<name>`
- **THEN** the report exists at `docs/self-review/<name>-self-review.md`, committed on
  `chore/archive-<name>`

#### Scenario: The report rides in the archive pull request

- **WHEN** run 2 pushes the archive branch and opens its pull request
- **THEN** that pull request carries both the archive commit and the self-review report, so the
  report cannot be stranded behind an already-merged archive

#### Scenario: The commit is refused on the wrong branch

- **WHEN** the main checkout is not on `chore/archive-<name>` when the report commit is attempted
- **THEN** the commit is not made, the mismatch is reported naming the branch found and the branch
  required, and the change stays `FINISHED`

### Requirement: Context is gathered deterministically, not by the model re-reading files

`/myflow-finish` SHALL invoke `scripts/gather-self-review-context.sh <archived-change-path> <name>
<state-dir> [<repo-root>]` to assemble the self-review input bundle before any reasoning pass runs.
The script SHALL collect, without performing any judgement:

- the SDD ledger, `docs/superpowers/ledgers/<YYYY-MM-DD>-<name>.md` — date-prefixed, exactly as
  `scripts/preserve-session-records.sh` writes it; the date is not known in advance, so the script
  SHALL locate it the same way that script locates an existing destination: a digit-anchored search
  for a file ending `-<name>.md` under `docs/superpowers/ledgers/`
- the review-panel record, `docs/superpowers/reviews/<YYYY-MM-DD>-<name>-panel.md` — located the
  same way, searching for a file ending `-<name>-panel.md`
- `tasks.md` from the archived change
- `git log --stat` covering the change's two finish-run-1 commits and the archive commit

A source that does not exist SHALL be reported as `skipped: <src> (absent)` on stdout, matching
`preserve-session-records.sh`'s own `skipped:`/`preserved:` vocabulary, and gathering SHALL continue
with the remaining sources — a change may legitimately have no panel record. The script SHALL exit 2
on an invocation error (a missing argument or an invalid change name); otherwise it SHALL always
exit 0 — a missing source is never a pass/fail determination and never changes the exit code.

A source that resolves outside the repository via a symlink SHALL be reported as `refused: <src>
(resolves outside the repository)` on stdout instead of `skipped:` — a refusal is never reported as
an absence — and SHALL NOT be read; the script SHALL still exit 0 for this case.

`<repo-root>` is **optional**. When it is omitted, the trusted repository root SHALL be derived
exactly as before, from the script's own process working directory. When it is supplied, it SHALL
become the trusted repository root, and SHALL first be validated by the same tests
`<archived-change-path>` receives: it SHALL be an absolute path, its lexically-collapsed form and its
symlink-resolved form SHALL be identical, and it SHALL be the root of a git repository. A supplied
`<repo-root>` failing any of those SHALL be an invocation error, exit 2 — the same class as a
missing argument or an invalid change name. **This deliberately differs from a malformed
`<archived-change-path>`**, which is not an invocation error: that case prints a `note:` line,
reports every source `skipped:`, and still exits 0, per the paragraph below. The two are treated
differently because a caller-supplied root that cannot be trusted leaves nothing safe to gather
against, whereas an untrustworthy archived path only means those sources go unread.

Trusting a caller-supplied root does not weaken the containment argument that follows: the
prohibition being defended is deriving the anchor from `<archived-change-path>` — the untrusted input
under test — not accepting one from the caller.

`<archived-change-path>` SHALL be validated by resolving it TWICE, independently, and comparing the
two results, rather than by a bounded, fixed-depth series of individual symlink checks:

1. A trusted repository root SHALL be `<repo-root>` when supplied, and otherwise SHALL be derived
   from the script's own process working directory via `git rev-parse --git-common-dir` (never
   `--show-toplevel`, which returns a *worktree's* root rather than the main repository's when run
   inside one) — and in neither case from `<archived-change-path>` itself, since that would trust the
   very path being validated.
2. `<archived-change-path>` SHALL be normalized into an absolute, lexically-collapsed path — `.` and
   `..` components resolved by pure string manipulation, never touching the filesystem — so this step
   alone cannot be fooled by a symlink at any component. Call this the LEXICAL path.
3. The LEXICAL path SHALL sit exactly one path segment past
   `<trusted-repository-root>/openspec/changes/archive/` — no deeper nesting, no shallower, checked
   as a real path boundary rather than a string prefix. Any other shape (missing, wrong location,
   extra nesting) is an invalid invocation.
4. `<archived-change-path>` SHALL ALSO be resolved semantically, following every symlink at every
   path component. Call this the REAL path.
5. The LEXICAL and REAL paths SHALL be compared for exact equality. Any divergence between them —
   whether from the leaf itself being a symlink, an ancestor at any depth being a symlink, a trailing
   slash, or a `..` component that would desync a purely lexical walk from the real filesystem — means
   a symlink sat somewhere between the trusted repository root and the leaf, and the invocation SHALL
   be refused: never used to derive anything — the repository root, the `tasks.md` trust boundary, or
   any other downstream path.

When `<archived-change-path>` fails validation for any of these reasons (no enclosing repository
found, the path does not resolve to the expected `openspec/changes/archive/<leaf>` shape, or the
lexical and real resolutions diverge), the script SHALL treat the invocation the same invalid-
invocation way: print a distinct `note:` line, report every source `skipped:`, and still exit 0 —
never resolve a symlink and use its target as a trust boundary. This single mechanism subsumes every
bypass shape a bounded, fixed-hop-count check could miss, since it makes no assumption about how many
ancestor levels exist or what shape a deviation takes — it looks for ANY divergence between what the
path string says and what the filesystem actually resolves to, not for a specific bypass pattern.

**A supplied `<repo-root>` does not change any of the above.** It replaces only where the trusted
root comes from; every containment, divergence and refusal rule stated here applies unchanged.

#### Scenario: The repo root is omitted

- **WHEN** the script is invoked with three arguments, from a process working directory inside the
  repository
- **THEN** the trusted repository root is derived via `git rev-parse --git-common-dir` exactly as
  before, and every existing call site behaves unchanged

#### Scenario: The repo root is supplied

- **WHEN** the script is invoked with a fourth argument naming an absolute, symlink-free git
  repository root
- **THEN** that path becomes the trusted repository root, and the archived path is validated against
  it regardless of the process working directory or the checked-out branch

#### Scenario: A malformed repo root is an invocation error

- **WHEN** the fourth argument is relative, reaches through a symlink, or is not a git repository
  root
- **THEN** the script exits 2 and gathers nothing

#### Scenario: All sources present

- **WHEN** the ledger, panel record and `tasks.md` all exist for the change
- **THEN** the script prints a bundle containing all four sources' content

#### Scenario: A source is missing

- **WHEN** the change has no review-panel record (no optional slot ever fired)
- **THEN** the script reports `skipped: docs/superpowers/reviews/<name>-panel.md (absent)` — the
  reported source name is `<name>-panel.md` without the date prefix, since no dated file was found
  to name — and still returns the other sources, exiting 0

#### Scenario: A source resolves outside the repository via a symlink

- **WHEN** a symlink is planted at a tracked source location (the ledger directory, the panel
  directory, or `<archived-change-path>/tasks.md`) and its target resolves outside the repository
  root
- **THEN** the script reports `refused: <src> (resolves outside the repository)` on stdout, does
  NOT read the symlink's target content into the bundle, does NOT report it as `skipped: ...
  (absent)`, and still exits 0

#### Scenario: The archived-change-path itself is a symlink

- **WHEN** `<archived-change-path>` is itself a symlink (e.g. a tracked symlink at
  `openspec/changes/archive/<date>-<name>` pointing outside the repository, with a `tasks.md` at
  the target)
- **THEN** the lexical and real resolutions of `<archived-change-path>` diverge, the script prints a
  distinct `note:` line naming the divergence, does NOT read the target's `tasks.md` (or anything
  else derived from resolving the symlink) into the bundle, reports every source `skipped:` rather
  than `refused:`, and still exits 0

#### Scenario: A symlinked leaf with a trailing slash

- **WHEN** `<archived-change-path>` is invoked with a trailing slash (e.g.
  `openspec/changes/archive/<date>-<name>/`, the exact shape `skills/myflow-finish/SKILL.md`
  documents), and the slash-stripped leaf is itself a symlink to a directory outside the repository
  with a `tasks.md` at the target
- **THEN** lexical normalization collapses the trailing slash on its own, the lexical and real
  resolutions still diverge because the leaf is a symlink, the script prints the same distinct
  `note:` line as the no-trailing-slash case, does NOT read the target's `tasks.md` into the bundle,
  reports every source `skipped:` rather than `refused:`, and still exits 0

#### Scenario: A symlinked ancestor at any depth

- **WHEN** any ancestor path component of `<archived-change-path>` — `openspec/changes/archive/`,
  `openspec/changes/`, `openspec/`, or any other component between the repository root and the
  leaf — is itself a symlink to a directory outside the repository, and the leaf itself
  (`<date>-<name>`) is an ordinary, non-symlink directory found underneath that symlinked ancestor,
  with a `tasks.md` present at that location
- **THEN** the lexical and real resolutions of `<archived-change-path>` diverge regardless of which
  ancestor, or how many levels up, carried the symlink, the script prints a distinct `note:` line,
  does NOT read the `tasks.md` found underneath it into the bundle, reports every source `skipped:`
  rather than `refused:`, and still exits 0 — the same as if the leaf itself had been the symlink

#### Scenario: A `..` path component that would desync a bounded lexical walk

- **WHEN** `<archived-change-path>` contains a literal `..` path component (e.g.
  `openspec/changes/archive/extra/../<date>-<name>`), such that the leaf it resolves to is reachable
  and not itself a symlink, but a real ancestor of that leaf (e.g. `openspec/` itself) is a symlink
  to a directory outside the repository
- **THEN** lexical normalization collapses the `..` component by pure string manipulation, landing on
  the ordinary-looking `openspec/changes/archive/<date>-<name>` shape and passing the shape check —
  but the real, symlink-following resolution lands somewhere else entirely, so the lexical and real
  paths diverge and the script refuses the invocation via the same mechanism as any other symlinked
  ancestor: a distinct `note:` line, every source `skipped:` rather than `refused:`, never reading the
  `tasks.md` reachable through the symlinked ancestor into the bundle, still exiting 0

#### Scenario: An archived-change-path with extra directory nesting

- **WHEN** `<archived-change-path>` has one or more extra directory levels between
  `openspec/changes/archive/` and the `<date>-<name>` leaf (e.g.
  `openspec/changes/archive/<subdir>/<date>-<name>`) — a shape that never matches
  `openspec/changes/archive/<leaf>` regardless of whether any component along the way is a symlink
- **THEN** the lexical path fails the shape check (more than one segment past the trusted archive
  root) and the script refuses the invocation before it even reaches the lexical-vs-real symlink
  comparison: a distinct `note:` line naming the unexpected location, every source `skipped:` rather
  than `refused:`, and still exits 0

### Requirement: Self-review runs only after FINISHED is written

`/myflow-finish` run 2 SHALL run self-review as its step 9, immediately after step 8 (write
`FINISHED`) succeeds, and only then. A run 2 that stops earlier — at a cleanup leftover, or on any
failure before the `FINISHED` write — SHALL NOT run self-review.

Self-review SHALL NOT be able to prevent, delay, or undo the `FINISHED` write. A failure inside
self-review SHALL be reported and SHALL NOT reopen the change or move its state.

**Only the step numbers change here.** Inserting the checkout-positioning step ahead of the sync
shifted every later run-2 step by one; the ordering this requirement exists to guarantee — self-review
strictly after the terminal write, never before it, and never able to affect it — is unchanged, which
is why the archive push was placed after self-review rather than the write being moved.

#### Scenario: Self-review follows a completed archive

- **WHEN** run 2 reaches step 8 and writes `FINISHED`
- **THEN** step 9 (self-review) runs next, before the terminal handoff prints

#### Scenario: A cleanup leftover skips self-review entirely

- **WHEN** run 2 stops at step 7 because `check-cleanup-complete.sh` reported `LEFTOVER:`
- **THEN** self-review does not run, and the change stays at `IN_PROGRESS` exactly as it does today

#### Scenario: A self-review failure never reopens the change

- **WHEN** any part of self-review fails after `FINISHED` has already been written
- **THEN** the failure is reported in the handoff and the change remains `FINISHED`

### Requirement: A skip prompt defaults to running self-review

On entering step 9, `/myflow-finish` SHALL ask:

> **Run self-review for this change?**
> - **Yes — run it** *(default, recommended)*
> - **No — skip**

Only an explicit **No** SHALL skip the remainder of step 9. Silence, a stalled prompt, or a session
that cannot ask SHALL run self-review rather than skip it, since nothing external is written before
the later per-finding and rating asks.

#### Scenario: Explicit No skips

- **WHEN** the operator answers **No — skip**
- **THEN** step 9 stops immediately, the handoff prints `Self-review: skipped`, and nothing else in
  the run-2 handoff changes

#### Scenario: Silence runs it

- **WHEN** the session cannot ask, or the prompt is not answered
- **THEN** self-review runs, exactly as an explicit **Yes** would
