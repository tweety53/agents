# myflow-self-review Specification

## Purpose
TBD - created by archiving change kan-23-myflow-self-review. Update Purpose after archive.
## Requirements
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

### Requirement: The operator rates the run, 1 to 5

After the report and the filing asks, self-review SHALL ask the operator to rate the run on a 1
(rough) to 5 (excellent) scale. The rating SHALL be recorded only in the self-review report file,
never in the state file.

#### Scenario: Rating is recorded in the report only

- **WHEN** the operator rates the run
- **THEN** the rating appears in `docs/self-review/<name>-self-review.md` and is never written to
  the change's `state.json`

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

### Requirement: One combined reasoning pass over all five angles

Self-review SHALL answer all **five** angles below in a single reasoning pass, fed the gathered
bundle plus the live session's own context — never as five separate subagent dispatches:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `myflow-fix` |
| 2 | Token/time cost, and what would reduce it without quality loss | `myflow-cost` |
| 3 | What went well, and how to reproduce it | `myflow-improvement` |
| 4 | What could be automated or moved to a script | `myflow-automation` |
| 5 | What could move to the Go app or its persistent storage | `myflow-stats-app` |

Angle 5 SHALL cover both **records** — which artifacts the pipeline writes to files today would be
better held in the stats daemon's database and queryable across runs — and **logic** — which
derivation work now performed by a Bash guard or by the agent itself could move into the `myflow`
CLI or the daemon. It SHALL NOT extend to what the SPA should display.

Each angle SHALL produce zero or more findings. An angle that produces none SHALL say so
**explicitly**, the way `## Decisions` and `## Open questions` are present-but-empty rather than
absent. An angle that is silent SHALL NOT be treated as an angle that yielded nothing: the two are
indistinguishable to a reader, and that is how a write-only angle passes unnoticed.

#### Scenario: One pass produces all five angles

- **WHEN** self-review runs
- **THEN** a single reasoning pass produces findings for all five angles, not five separate
  dispatches

#### Scenario: An angle with nothing to report says so

- **WHEN** one angle produces no findings at all
- **THEN** that angle's section is present and carries an explicit statement that it produced none,
  rather than being omitted or left empty

### Requirement: The filing ask covers every angle's findings

The filing ask SHALL cover the findings of **every** angle, not only the problems angle.

Before any prompt fires, self-review SHALL explain **each** finding in the message body, stating
three things: what was observed, what breaks because of it, and what the fix would be. A prompt's
option text SHALL NOT be relied on to carry that explanation. The operator SHALL have the
explanation in front of them before the decision is recorded, because a filed issue is durable and
already on the board — an explanation that arrives after the filing describes something the operator
did not agree to.

The decision SHALL then be recorded as **one multi-select prompt per angle**, listing that angle's
findings, defaulting to filing none of them. A bare observation with no concrete change implied
SHALL NOT appear as an option.

A filed issue SHALL carry every label on the change's linked issue, plus `AI-generated`, **plus the
label naming the angle that produced it** from the table above, and SHALL link to that issue when
one exists, per **Labels on issues the pipeline creates**
(`skills/myflow-contracts/jira-integration.md`). A filing failure SHALL degrade to the standard
`⚠ Jira: skipped — <reason>` line and SHALL NOT stop self-review, per **Never blocking**
(`skills/myflow-contracts/jira-integration.md`).

This explain-before-filing rule SHALL bind every filing ask the pipeline offers, not only
self-review's — `/myflow-finish` run 1's follow-up filing included.

#### Scenario: A cost or what-went-well finding is offered for filing

- **WHEN** an angle other than problems produces a finding naming a concrete pipeline or script
  change
- **THEN** that finding is explained in full and appears as an option in its angle's filing prompt,
  exactly as a problems-angle finding does

#### Scenario: The explanation precedes the prompt

- **WHEN** self-review has findings to offer
- **THEN** every finding's observation, consequence and proposed fix are stated in the message body
  before the first filing prompt is presented

#### Scenario: A filed issue carries its angle's label

- **WHEN** the operator selects a finding from the automation angle for filing
- **THEN** the created issue carries `myflow-automation` in addition to the linked issue's labels
  and `AI-generated`

#### Scenario: Operator declines every finding in an angle

- **WHEN** the operator selects nothing in an angle's filing prompt
- **THEN** no issue is created for that angle, and self-review continues to the next angle

### Requirement: The report records each finding in a parseable form

`docs/self-review/<name>-self-review.md` SHALL carry one section per angle, all five present, in the
order the angle table states.

Each finding SHALL be recorded as a single line naming, in a fixed form, the label of the angle that
produced it, the finding itself, and its disposition — the issue key when it was filed, or an
explicit declined marker when it was not:

```markdown
- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201
- **[myflow-fix]** Preflight compares against a stale local base ref — declined
```

An angle that produced no findings SHALL carry an explicit none-marker in place of finding lines.

The angle label on a finding line SHALL match the section the line sits under. A finding recorded as
filed SHALL name an issue key.

#### Scenario: A filed finding names its issue

- **WHEN** a finding was filed as KAN-201
- **THEN** its line in the report names `myflow-cost` and `filed: KAN-201`

#### Scenario: An empty angle carries a marker

- **WHEN** the automation angle produced no findings
- **THEN** its section is present and carries the none-marker rather than being empty or omitted

### Requirement: A guard checks every self-review report

The repository SHALL carry a guard that checks every report under `docs/self-review/` against the
report shape above: all five angle sections present; each section carrying either at least one
finding line or the none-marker; each finding line parseable, with its angle label matching its
section and a filed finding naming an issue key.

The guard SHALL report per-member coverage and SHALL exit 0 when clean, 1 when it finds violations,
and 2 when it cannot answer at all — the exit-code contract every guard in this repository carries.

Reports written before this requirement existed SHALL be **declared by name, with a reason**, in the
guard's own source, exactly as a legitimately-zero corpus member is declared under **Requirement:
Zero coverage SHALL be declared or SHALL be a violation** (`openspec/specs/agents-repo-verification/spec.md`).
A report absent from that declared set SHALL satisfy every rule. The guard SHALL NOT infer that a
report predates the rule from the report's own content — a marker a new report can forget to write
is a mechanism for passing without being checked, which is the outcome the declaration exists to
prevent.

#### Scenario: A new report omits an angle

- **WHEN** a report not on the declared pre-rule list is missing the angle-5 section
- **THEN** the guard names the report and the missing section, and exits 1

#### Scenario: A pre-rule report is declared

- **WHEN** the guard runs against the reports that predate this requirement
- **THEN** each is reported as declared, with its reason, and the guard exits 0

#### Scenario: A finding claims to be filed but names no issue

- **WHEN** a finding line is marked filed and carries no issue key
- **THEN** the guard names that line and exits 1

