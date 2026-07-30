## Context

`/myflow-finish` is a two-run command, and which run happens is decided by one thing: whether the
change's branch has already reached the base branch. `skills/myflow-contracts/pipeline.md`'s Finish
contract implements that as `git merge-base --is-ancestor HEAD origin/$BASE`.

That test already carries one guard, added earlier and working: never resolve the base branch from
`HEAD@{upstream}`, because `/myflow-finish` runs inside the apply worktree where `HEAD` *is* the
change's own branch, so that resolution compares the branch with itself. On KAN-14 the guard did its
job — the base resolved to `main` and was asserted distinct from the current branch.

The hole is elsewhere. **A branch with no commits of its own is an ancestor of every branch.** When
`/myflow-do` has staged work but committed nothing — the normal state at `IN_PROGRESS`, since
`/myflow-do` is forbidden from committing until a PR exists — `HEAD` is still the merge base, the
ancestor test is trivially true, and run 2 proceeds to archive the change, push the archive, and run
`git worktree remove --force`. Base resolution was right; the merge check was still wrong.

Two adjacent gaps live in the same command. The SDD ledger that records which model ran each
subagent dispatch sits at `.superpowers/sdd/tasks/progress.md`; `.gitignore` excludes
`.superpowers/`, and run 2 removes the worktree. `myflow-model-policy` states that limitation
honestly and defers durability to whichever change owns the archive step. And `/myflow-start` writes
the published proposal's HTML source to `<state-dir>/<name>-proposal-artifact.html` so a revision
round republishes to the same URL — a file the finish contract never mentions, so across three
finished changes two had it removed by hand and one did not.

A fourth item was added by the operator during this planning run. `/myflow-start` currently spends
the same effort on every change: the full brainstorming checklist, the full plan enrichment,
regardless of whether the change is a one-line contract fix or a parser rewrite. There is no way to
say "this one is small" without the operator hand-steering each question, and no record of what was
chosen.

## Goals / Non-Goals

**Goals:**

- Make the run-1/run-2 decision impossible to get wrong on a zero-commit branch, and make that
  decision testable rather than reviewable-only.
- Give a "cannot determine" input an honest refusal with an operator override, not a guess.
- Make the model record survive the change, without creating pressure to fill in values nothing
  measured.
- Account for the proposal artifact source so its fate stops depending on whether someone noticed.
- Let the operator size a change's planning effort once, at the only point where the change is
  created, and record what was chosen.

**Non-Goals:**

- Letting the effort level weaken any gate. Brainstorming, the design approval gate, writing-plans
  and the review panel are not effort-scaled; only how much reasoning is spent inside them is.
- Effort governing `/myflow-do` or `/myflow-finish`. The recorded value is a record and a default for
  a revision round, not an input those commands act on.

- Item 1 of KAN-19 — widening `scripts/check-plan-provenance.py`'s scan scope beyond `tasks.md`. It
  shares no file with this work and is a separate change.
- Commit messages. The issue notes one false claim in KAN-14's own final commit message; commit
  messages are outside any guard's reach and this change does not try to bring them inside it.
- Preserving the rest of `.superpowers/sdd/` — fix-wave briefs, per-task diffs, per-task reports.
- Re-introducing a verification gate to `/myflow-finish`. The preflight is a safety check on the
  branch's own state, not a test or lint run; the contract's "no verification gate" rule stands.
- KAN-15, serialization of independent tasks in `/myflow-do`.

## Decisions

### Split KAN-19, shipping the finish items first

**ID:** split-finish-first
**Status:** active
**Chosen:** items 2, 3 and 4 in this change; item 1 in a separate one — they share no file and no
requirement, and item 3 destroys work, so it must not wait behind a parser rewrite.
**Considered:** one change covering all four items — a single review panel and one merge, but a plan
mixing a Python parser rewrite with a git-safety contract change, with the destructive bug held
behind it. Item 3 alone as a minimal hotfix — fastest removal of the destructive path, but leaves
three follow-ups open and re-opens `/myflow-finish` three times for work that could land together.

### A missing recorded merge base refuses and asks

**ID:** missing-merge-base-refuses
**Status:** active
**Chosen:** stop, report `HEAD`, the base and the uncommitted count, and require explicit operator
confirmation — no recorded merge base is an honest unknown, and this whole change exists because a
check returned a confident wrong answer where an unknown was available.
**Considered:** fall back to the ancestor test plus a clean-tree check — catches the KAN-14 shape
without operator input, but a worktree that was committed and then reset would pass it silently.
Refuse outright with no override — safest, but locks out every change whose state file predates this
fix with no in-band route forward, which invites hand-editing the state file instead.

### Run 1 preserves the session records, not run 2

**ID:** preserve-at-run-1
**Status:** active
**Chosen:** `/myflow-finish` run 1, before its `git add -A` — the records land in the same commit as
the work they describe and merge with it, they survive a change that is merged but never archived,
and run 2 gains no new write before its irreversible steps.
**Considered:** run 2 before worktree removal, which the issue suggested and KAN-14's manual
precedent followed — but it adds a write to the one command performing irreversible operations, and
loses the record entirely for a change that merges without being archived. `/myflow-do` authoring
them under `docs/superpowers/` from the start — maximum durability, but it puts per-task ledger churn
into the staged diff the human reads at the `IN_PROGRESS` gate.

### Preserve the ledger and the review panel record, and nothing else

**ID:** preserve-ledger-and-panel
**Status:** active
**Chosen:** exactly what was kept by hand on both prior changes — the ledger answers "which model ran
this task", the panel record answers "what was found and how it was resolved", and nothing else
reproduces either.
**Considered:** the ledger alone — smallest footprint and the only thing `myflow-model-policy`
strictly requires, but it drops a record the operator chose to keep by hand twice. Everything under
`.superpowers/sdd/` — nothing to decide per change, but the per-task diffs duplicate commits already
in git history.

### The proposal artifact source is preserved in-repo, then removed from the state directory

**ID:** artifact-source-preserve-then-remove
**Status:** active
**Chosen:** copy it into the repository at run 1 alongside the other records, and remove the
state-directory copy in run 2's disclosed cleanup — the recorded `artifactUrl` stays republishable
from git history, and the state directory stops accumulating one file per finished change.
**Considered:** delete at run 2 with no preservation, matching what `kan-8` and `kan-10` did by hand —
simplest, but the terminal state file advertises an `artifactUrl` forever while the only source that
could republish it is gone. Retain it where it is and say so in the contract — zero machinery, but
the state directory grows without bound and the file keeps looking like litter.

### The run-2 decision becomes an executable script with a test harness

**ID:** preflight-as-script
**Status:** active
**Chosen:** `scripts/check-finish-preflight.sh` plus `scripts/test-check-finish-preflight.sh`,
matching this repository's three existing `check-*`/`test-*` pairs — the one item with real branching
logic becomes the one item a test can hold to account.
**Considered:** prose contract only, extending `pipeline.md`'s snippet — consistent with how every
other finish rule is written and adds nothing to install, but the fix would be verified only by
review, exactly as the broken version was. A script for the run-1/run-2 decision with prose for the
preservation and removal steps — smallest executable surface covering the dangerous branch, but it
leaves the copy step, which touches three source paths that can each be absent, untestable.

### Effort governs this `/myflow-start` run's reasoning, and nothing further

**ID:** effort-scopes-this-run
**Status:** active
**Chosen:** the answer sets how much reasoning `/myflow-start` spends on its own brainstorming and
plan enrichment — later commands decide their own depth and do not read it as an instruction.
**Considered:** pipeline depth across all three commands, with `/myflow-do` scaling its review panel
and fix rounds from the same answer — one question sizes the whole change, but it lets a value chosen
before anything was known silently thin the panel that catches implementation defects, which is the
one place this pipeline deliberately does not economise. Both as a single scale — simplest to answer,
but it conflates a planning-time budget with a review-time one, and the same change may want them
different.

### The chosen effort is recorded in the state file

**ID:** effort-in-state-file
**Status:** active
**Chosen:** a new `effort` field, written by `/myflow-start` and carried forward verbatim by every
other command — exactly the shape `jiraIssue` already has, so the carry-forward duty needs no new
mechanism.
**Considered:** recording it in the proposal artifact and `design.md` only — human-readable with no
schema change, but a later command would have to parse markdown to read it, which nothing in the
pipeline does. Not storing it at all and asking fresh each run — no schema change and no stale value,
but the operator answers repeatedly and a revision round could silently contradict the first run.

## The preflight's shape

The script takes three positional arguments — worktree path, resolved base ref, recorded merge base
(or `-` when absent) — and prints exactly one verdict word on stdout followed by a reason. It exits
`0` whenever it reached a verdict and `2` on environment failure.

**The verdict, not the exit status, carries the answer.** An exit-code-only protocol would make
"cannot determine" indistinguishable from "violation", which is the confusion this change exists to
remove. `2` keeps its established meaning from this repository's other guards: the script cannot
determine anything about the tree.

| # | Condition | Verdict |
|---|-----------|---------|
| a | recorded merge base is `-` or empty | `REFUSE` — no recorded merge base |
| b | `HEAD` equals the recorded merge base | `RUN1` — the branch has no commits of its own |
| c | `HEAD` is not an ancestor of the base ref | `RUN1` — not merged |
| d | the worktree has uncommitted tracked changes or untracked-unignored files | `REFUSE` — merged by ancestry, but entries are uncommitted |
| e | otherwise | `RUN2` |

**Check b runs before check c, and that ordering is the fix.** A branch with no commits of its own is
an ancestor of everything, so c returns "merged" on precisely the most dangerous input.

**Check d runs after check c**, because dirty-and-unmerged is the ordinary in-flight state and must
stay `RUN1` rather than becoming a refusal the operator sees on every normal finish.

Both sides of check b are normalised through `git rev-parse --verify <ref>^{commit}`, so a shortened
sha in the state file compares correctly against a full `HEAD`.

**Base-branch resolution stays in the contract, not in the script.** The script receives the resolved
base as an argument. The existing `HEAD@{upstream}` guard works and is well argued in `pipeline.md`;
copying it into a script would create a second copy of a procedure that gates a destructive
operation, which is exactly what `myflow-finish`'s own guardrails forbid.

On a multi-repo change the skill runs the script once per `worktrees` key and proceeds to run 2 only
if every worktree returns `RUN2`.

Run 2's existing cleanup checks 1 and 2 would already have caught KAN-14's staged entries — they run
at step 4, after the irreversible archive-and-push at step 3. Check d is those same checks hoisted to
a precondition.

## The preservation step's shape

`scripts/preserve-session-records.sh` copies three sources into the repository:

| Source | Destination |
|--------|-------------|
| `.superpowers/sdd/tasks/progress.md` | `docs/superpowers/ledgers/<date>-<name>.md` |
| `.superpowers/sdd/final-review-panel.md` | `docs/superpowers/reviews/<date>-<name>-panel.md` |
| `<state-dir>/<name>-proposal-artifact.html` | `docs/superpowers/artifacts/<date>-<name>.html` |

`<date>` is fixed at the first copy. The script globs each destination for an existing `*-<name>.*`
and reuses that path when one is found, so a re-copy overwrites in place instead of accumulating one
dated duplicate per fix round.

**A missing source is reported in one line and is never a failure.** A change may legitimately have
no panel record. A preservation step that can block an integration would be a worse failure than the
gap it closes.

It is invoked from `/myflow-finish` run 1 and from `/myflow-do`'s commit path — the `prUrl` case, the
one place `/myflow-do` is permitted to commit — so a fix round raised after a PR is open refreshes
the records rather than leaving them a round stale.

**The `unknown (agent-defined)` rule survives untouched.** Panel slots dispatched by `subagent_type`
resolve their model from their own agent definition, which the dispatcher never reads. Durability
must not create pressure to fill those entries with a plausible value nothing measured — that is the
failure the model-record requirement exists to police, and KAN-14's own panel record already made it
once, listing two such slots as `sonnet`.

## The effort ask, and the schema hazard it walks into

`/myflow-start` asks **once per change, on the run that creates it**. A revision round — the command
re-entered at `STARTED` — reads the recorded value, states it, and does not ask again. That is what
"the first time" means operationally: the state file's absence, not a guess about the operator's
familiarity.

The ask uses the same interactive mechanism as `/myflow-finish`'s integration question, so the
no-flags rule is untouched: effort is never an argument, and an argument that is not a change name
is still reported rather than interpreted.

Three levels, and what each may change:

| Level | What it changes |
|-------|-----------------|
| `low` | Questions are batched rather than asked one at a time; the design is presented once; `tasks.md` is enriched to plan quality but grouped more coarsely |
| `medium` | The default. The checklist is followed with related questions grouped |
| `high` | Every checklist item worked separately, alternatives enumerated per open question, each design section approved on its own |

**What no level may change.** Brainstorming still runs, the design approval gate still holds,
writing-plans still runs, and `tasks.md` is never left a thin scaffold — those are `/myflow-start`'s
standing guardrails, and an effort level that could switch one off would make the field a way to skip
the gates rather than a way to size the thinking inside them. `low` means fewer rounds, not fewer
gates.

### The closed schema is the real hazard here

`skills/myflow-contracts/state-self-heal.md` closes the state-file schema **in both directions**: a
file missing a documented field is unparseable, and a file carrying a key the contract does not
document is unparseable too. Adding `effort` naively therefore breaks in both directions at once —
every state file written before this change lacks the key, and every file written after it carries
one that older installed copies of the contract do not know.

`effort` is therefore documented as a field whose **absence is a legal value**, meaning "not
recorded", read as `null`. This is the first field added since the schema closed, so the carve-out is
stated explicitly in the contract rather than left as an inference from the field's nullability —
`jiraIssue`, `artifactUrl` and `prUrl` are all documented as *present and nullable*, which is a
different thing from *absent*.

Two alternatives were rejected. Requiring the key and treating older files as unparseable would route
all three existing finished changes through self-heal, which announces unrecovered fields and
rewrites from artifact inference — a loud, alarming correction for a field nobody had a chance to
set. Versioning the state file with a schema-version key would solve it generally, but it introduces
a migration mechanism for one optional field and would itself be a new undocumented key on every file
that predates it, reproducing the same problem one level up.

## Risks / Trade-offs

- **The preflight adds a way for a legitimate finish to be refused** → verdict `REFUSE` never acts on
  its own; it stops and asks, with `HEAD`, the base and the uncommitted count on screen. The operator
  can override. A refusal costs one prompt; the failure it replaces costs the change.
- **A second copy of the merge check could drift from `pipeline.md`** → the contract points at the
  script and states the verdicts; it does not restate the checks. The script's header carries the
  rationale, as this repository's other guards do.
- **The preserved panel record is large** — KAN-14's was 119 KB → accepted. It is text, it compresses,
  and it is written once per change. The alternative is losing the only narrative record of what the
  panel found.
- **`docs/superpowers/artifacts/` will hold HTML that references a live URL** → the file is a source
  for republishing, not a served page. Nothing in the repository renders it.
- **`.superpowers/sdd/tasks/progress.md` is unconfirmed** → the path comes from
  `myflow-model-policy`'s prose and the issue text; no worktree existed while planning. `tasks.md`
  tags it for confirmation at implementation time rather than asserting it, and the copy step treats
  a missing source as a reported skip, so a wrong path degrades to a visible no-op rather than a
  failure.
- **An effort level chosen before anything is known could under-size a change** → it governs only how
  `/myflow-start` reasons, never whether a gate runs, and the operator sees every design section
  regardless. A change that turns out larger than expected is revised by re-running `/myflow-start`,
  which is already the supported path.
- **A recorded `effort` invites a later command to start acting on it** → the requirement states that
  the field is a record and a revision-round default, not an input `/myflow-do` or `/myflow-finish`
  reads. Widening it is a change to those commands' specs, not a reinterpretation of this field.
- **Older installed copies of the contract will see an undocumented `effort` key** → the schema
  carve-out is added to `state-file.md` and `state-self-heal.md`, which `setup.sh` installs into every
  harness. Until an installation is refreshed, a stale copy would read a new file as unparseable and
  self-heal it — announcing the correction, which is the designed loud path rather than a silent one.
- **The two new scripts are repository-local and are not installed elsewhere by `setup.sh`** → the
  contract text that points at them is installed. A harness running `/myflow-finish` against a
  repository without the script must be told what to do; the contract states the fallback in prose,
  so the check is never silently skipped.

## Migration Plan

No data migration. Changes already at `STARTED` or `IN_PROGRESS` continue to work unchanged, because
`/myflow-do` has always written the merge base into `worktrees`. A change whose state file carries no
recorded merge base — self-healed, or hand-edited — receives the `REFUSE` path with an explicit
operator override, which is the designed behaviour rather than a migration gap.

Rollback is reverting the commit: the contract text and the scripts land together, and no state file
format changes.

## Open Questions

None. The one unconfirmed fact — the ledger's exact path — is handled in the plan as a
confirm-before-use step rather than left open here.
