## ADDED Requirements

### Requirement: Self-review runs only after FINISHED is written

`/myflow-finish` run 2 SHALL run self-review as its step 8, immediately after step 7 (write
`FINISHED`) succeeds, and only then. A run 2 that stops earlier — at a cleanup leftover, or on any
failure before the `FINISHED` write — SHALL NOT run self-review.

Self-review SHALL NOT be able to prevent, delay, or undo the `FINISHED` write. A failure inside
self-review SHALL be reported and SHALL NOT reopen the change or move its state.

#### Scenario: Self-review follows a completed archive

- **WHEN** run 2 reaches step 7 and writes `FINISHED`
- **THEN** step 8 (self-review) runs next, before the terminal handoff prints

#### Scenario: A cleanup leftover skips self-review entirely

- **WHEN** run 2 stops at step 6 because `check-cleanup-complete.sh` reported `LEFTOVER:`
- **THEN** self-review does not run, and the change stays at `IN_PROGRESS` exactly as it does today

#### Scenario: A self-review failure never reopens the change

- **WHEN** any part of self-review fails after `FINISHED` has already been written
- **THEN** the failure is reported in the handoff and the change remains `FINISHED`

### Requirement: A skip prompt defaults to running self-review

On entering step 8, `/myflow-finish` SHALL ask:

> **Run self-review for this change?**
> - **Yes — run it** *(default, recommended)*
> - **No — skip**

Only an explicit **No** SHALL skip the remainder of step 8. Silence, a stalled prompt, or a session
that cannot ask SHALL run self-review rather than skip it, since nothing external is written before
the later per-finding and rating asks.

#### Scenario: Explicit No skips

- **WHEN** the operator answers **No — skip**
- **THEN** step 8 stops immediately, the handoff prints `Self-review: skipped`, and nothing else in
  the run-2 handoff changes

#### Scenario: Silence runs it

- **WHEN** the session cannot ask, or the prompt is not answered
- **THEN** self-review runs, exactly as an explicit **Yes** would

### Requirement: Context is gathered deterministically, not by the model re-reading files

`/myflow-finish` SHALL invoke `scripts/gather-self-review-context.sh <archived-change-path> <name>
<state-dir>` to assemble the self-review input bundle before any reasoning pass runs. The script
SHALL collect, without performing any judgement:

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

`<archived-change-path>` SHALL be validated by resolving it TWICE, independently, and comparing the
two results, rather than by a bounded, fixed-depth series of individual symlink checks:

1. A trusted repository root SHALL be derived from the script's own process working directory via
   `git rev-parse --git-common-dir` (never `--show-toplevel`, which returns a *worktree's* root
   rather than the main repository's when run inside one) — and never from
   `<archived-change-path>` itself, since that would trust the very path being validated. The
   script's documented caller (`skills/myflow-finish/SKILL.md`) always invokes it with the process
   working directory at the repository root.
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

### Requirement: One combined reasoning pass, not four separate dispatches

Self-review SHALL answer all four angles below in a single reasoning pass, fed the gathered bundle
plus the live session's own context — never as four separate subagent dispatches:

1. problems encountered, and what pipeline change would avoid them
2. token/time cost, and what would reduce it without quality loss
3. what went well, and how to reproduce it
4. what could be automated or moved to a script

#### Scenario: One pass produces all four angles

- **WHEN** self-review runs
- **THEN** a single reasoning pass produces findings for all four angles, not four separate
  dispatches

### Requirement: Each actionable finding gets its own Jira filing ask

For each finding that names a concrete pipeline or script change (as opposed to a bare
observation), self-review SHALL ask once:

> **File `<one-line finding>` as a Jira issue?**
> - **No — don't file** *(default, recommended)*
> - **Yes — file it**

Only an explicit **Yes** SHALL file. A filed issue SHALL carry every label on the change's linked
issue plus `AI-generated`, and SHALL link to that issue when one exists, per **Labels on issues the
pipeline creates** (`skills/myflow-contracts/jira-integration.md`). A filing failure SHALL degrade
to the standard `⚠ Jira: skipped — <reason>` line and SHALL NOT stop self-review, per **Never
blocking** (`skills/myflow-contracts/jira-integration.md`).

#### Scenario: Operator declines a finding

- **WHEN** the operator answers **No — don't file** for a finding
- **THEN** no Jira issue is created for it, and self-review continues to the next finding

#### Scenario: Operator files a finding

- **WHEN** the operator answers **Yes — file it**
- **THEN** a new Jira issue is created, titled from the finding, carrying `AI-generated` plus the
  linked issue's labels, and linked to the change's issue when one exists

#### Scenario: A bare observation is never asked about

- **WHEN** a finding states an observation with no concrete change implied
- **THEN** no filing ask is made for it

### Requirement: The operator rates the run, 1 to 5

After the report and the filing asks, self-review SHALL ask the operator to rate the run on a 1
(rough) to 5 (excellent) scale. The rating SHALL be recorded only in the self-review report file,
never in the state file.

#### Scenario: Rating is recorded in the report only

- **WHEN** the operator rates the run
- **THEN** the rating appears in `docs/self-review/<name>-self-review.md` and is never written to
  the change's `state.json`

### Requirement: The report is committed to a fixed path

Self-review SHALL write `docs/self-review/<name>-self-review.md`, containing the four-angle report,
the rating, and which findings were filed versus declined, then commit it on the base branch (in
the main checkout) and push.

#### Scenario: Report path is fixed per change

- **WHEN** self-review completes for change `<name>`
- **THEN** the report exists at `docs/self-review/<name>-self-review.md`, committed and pushed on
  the base branch

### Requirement: The run-2 handoff names the self-review outcome

Run 2's terminal `## Finished` block SHALL carry one additional line:

```
**Self-review:** docs/self-review/<name>-self-review.md (rating: <n>/5) | skipped
```

#### Scenario: Self-review ran

- **WHEN** self-review completed and the operator rated the run
- **THEN** the handoff's `Self-review` line names the report path and the rating

#### Scenario: Self-review was skipped

- **WHEN** the operator answered **No — skip** at the step-8 prompt
- **THEN** the handoff's `Self-review` line reads `skipped`
