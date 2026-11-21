# Review panel

Loaded by `skills/flow/SKILL.md` immediately after `skills/flow/implement.md`'s `flow.implement`
stage closes, on every implementation run — creating, resumed, or fix. Dispatches the panel: the
`panel` bundles of this run's newest decision (`<abs-worktree>/.superpowers/sdd/decision.json`, per
**Plan check gate**, `skills/flow/brainstorm-planner.md`), each bundle one reviewer dispatch. This file owns
dispatch: mapping each slot id to its prompt and spawning each bundle.

```bash
flow stage begin -command '/flow' -stage flow.review -harness <harness> -session-token mf-<literal-token> <name>
```

**The pass log is store rows, rendered — never a hand-written file.** Every fact this file records
about a panel run the parent records as it arises with `flow record pass` or `flow record mutation`
(`-change <name> -round <n>`, the round `0` for the initial panel and `1..n` for a fix round).
`flow record render -kind panel` renders them into the panel record's pass-log section under
`<abs-worktree>/.superpowers/sdd/reviews/` in the canonical worktree, beside the findings.
**A run's first round is the number `flow record next-round -change <name>` prints**, never one
carried from session memory or an earlier run's handoff, and each later round of the run is one
more — so a later run never reuses a round an earlier run of the change already recorded under.

**Everything from the panel-diff write to the throwaway worktrees and the `[PRINCIPLES_PATH]`
and `[STANDARDS_PATHS]` resolution below is one Bash call**, under the turn discipline of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`): every verdict printed, and only the dispatch depends on them — an
absent principles file or an operator prompt is read off that call's
output and handled before any slot launches.

**A guard's exit status is captured bare — a reader command's is not the guard's.** A guard's
exit status read through a pipe is the pipe's, not the guard's: `flow-guard write-panel-diff final
<abs-worktree> <worktree> <merge-base> | tail` exits with `tail`'s 0 whatever verdict the guard reached. Capture
the output to a file and read `$?` after, use `set -o pipefail`, or run the guard
bare — so a red verdict cannot be laundered into green by the reader command.

## The panel

Every slot id a `panel` bundle names maps to one prompt:

| id | Slot | Prompt |
|---|------|---------------|
| `primary` | **Primary** — plan alignment and senior code review | `primary-reviewer-prompt.md`: `final-review.diff` against `proposal.md`, `design.md` and each task's `**Files:**`/`**Tests:**`/`**Commit:**` fields in `tasks.md`, plus code quality, architecture, testing and production readiness |
| `principles` | **Principles** | `principles-reviewer-prompt.md`; all three principle groups always apply <!-- refs-guard:allow --> |
| `test-sanity` | **Test-sanity** — would each test the diff adds or changes fail if its behaviour broke | `test-sanity-reviewer-prompt.md` |
| `call-efficiency` | **Call-efficiency** — second callers, a read repeated in one request, a query per list item | `call-efficiency-reviewer-prompt.md` |
| `failure-modes` | **Failure-modes** — error return, timeout, partial write, concurrent re-entry | `failure-modes-reviewer-prompt.md` |
| `security` | **Security** — injection, authorisation, secrets, trust boundary; added by `flow-guard plan-check` when a task's `**Files:**` matches the project's `## security paths`, never picked by the decider | `security-reviewer-prompt.md` |
| `mutation` | **Mutation** — sabotage-proofing | the mutation-testing brief below, own throwaway worktree copy per repository (see **The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`) |

A role a bundle drops leaves the bundle dispatching the roles it still names, and a bundle left
naming none is not dispatched.

**A subagent-facing file is passed by absolute path, never read into this context.** This repository's
`primary-reviewer-prompt.md` and `reviewer-calibration.md` (Primary), `principles-reviewer-prompt.md` and
`engineering-principles.md` (Principles),
`failure-modes-reviewer-prompt.md` (Failure-modes), `security-reviewer-prompt.md` (Security), and
`<project>/.flow/project.md`'s standards files are inputs to the slot that reads them; the
dispatcher resolves their paths, confirms each exists, and names them in the prompt.

**Fill every template placeholder from this table — never by reading the template.** Each is
substituted in the dispatch prompt with the value named here:

| Placeholder | Template | The parent substitutes |
|---|---|---|
| `[DIFF_PATH]` | all six | the absolute path of the diff this dispatch reads, in `<abs-worktree>/.superpowers/sdd/` of the canonical worktree: `final-review.diff` in pass 1; `late-fix.diff` on a fix's review (`skills/flow/fix.md`, step 4); on a re-run, the round's `fix-round-N.diff` |
| `[ARTIFACT_PATHS]` | primary | the absolute paths of the change's `proposal.md`, `design.md` and `tasks.md` in the plan directory (`<changeRoot>`, or a satellite's canonical change directory) |
| `[CALIBRATION_PATH]` | primary | the **absolute** path of `reviewer-calibration.md` beside this file, resolved as `[PRINCIPLES_PATH]` is |
| `[PRINCIPLES_PATH]` | principles | as the Principles section below resolves it |
| `[STANDARDS_PATHS]` | principles | as the Principles section below resolves it; empty when none resolve |
| `[GLOBAL_CONSTRAINTS]` | principles | the absolute path of the change's `design.md`, or `none` when it has none |

**A spawn type this harness's Agent tool does not offer is substituted on the record, never
silently**: the slot runs on the closest `flow-<effort>` type the harness offers — never `general-purpose`, which can delegate — and the round records `flow record pass -round <round> -note 'panel: <slot ids> on <the type it ran on> — <type> not offered'`.

**Check for an operator-named id at two points**: at the start of this stage (has the operator, in
this run's own argument or in the session before this stage, named an id the decided panel does not
carry?), and again at the start of every fix round below — an operator may ask mid-run, after seeing
pass 1's result, and that request adds the slot starting from the round it was made, never
retroactively to a pass already closed. Record which
slots were added this way and why (the operator's own words) with `flow record pass -round <round>`.

**Model and effort belong to the dispatch's role, not the slot**: every pass-1 panel dispatch runs
on role `panel-reviewer`, and every fix-round re-run on role `panel-rerun`
(**Panel re-runs**), each on the pair `flow-guard pair-judge` picks for it (**Model and effort**,
`skills/flow/brainstorm-planner.md`) — the dispatch's `subagent_type` is the pair's agent type, the
model passed explicitly as the Agent tool's own `model` parameter — and both `model` and `-effort`
are recorded. A slot carries no model of its own.

**Load `skills/flow/review-panel-optional-slots.md`** before dispatching any round that
dispatches `mutation` — it carries the mutation slot's prompt and **The throwaway worktree**.

**A panel runs as decided.** No slot outside the decided panel is ever added by diff
size, touched area, or any other automatic trigger: beyond the decided panel, anything reaches the
panel only through an explicit per-run operator instruction, for that run only. A change whose every
task's `**Files:**` is documentation is decided `primary` alone by `plan-check`, never reduced here.

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` (the canonical worktree's) and its
`[TOUCHED_FILES]` list once per round from **every** worktree in the change's resolved set
(**Resolving a change's worktrees**, `skills/flow-contracts/worktree-resolution.md`), in resolved
order, with one call:

```bash
flow-guard write-panel-diff final <abs-worktree> <worktree> <merge-base> [<worktree> <merge-base>…]
```

Its header (`stats/internal/guard/writepaneldiff.go`) is canonical for both files' shape. Exit 0
prints the two paths; exit 2 — it cannot answer, and nothing was written — stops the run with its
stderr. Then dispatch the round's
`panel` bundles in the canonical worktree, each reading the whole combined file; a role is
never dispatched once per worktree: one pass reads every worktree's section, so a seam between two repositories is in one pass's view. **Bundled dispatch** below states how a round's roles ride
those dispatches.

### Bundled dispatch

Before dispatching any panel round,
re-check the decision's `panel` bundles.

**Each `panel` bundle is one dispatch**, one **PASS `<id>`** section per role in the
bundle's order — findings still recorded per role. **The one-per-role
shape — one dispatch per role — is the explicit fallback, taken when a bundled dispatch's findings
need separation** — a finding the bundle cannot attribute to its role, or a bundled read that must
separate to hold review quality — and recorded with `flow record pass -round <round>` naming why.

**One `dispatches` row per bundle** — the same `flow record dispatch begin`/`end` pair below, with
`-slot` the bundle's roles `+`-joined in the bundle's order (`primary+principles`) and
`-model`/`-effort` the bundle's role's pair. Every finding still records its own single role in `-slot`, with the
bundle's `-dispatch-seq`.

An incident a round's slot caused is saved as **4. Execute (SDD + TDD)** saves one
(`skills/flow/implement.md`).

**Every dispatch's brief is assembled by hand, once the round's diff is written** — one file per
dispatch, a one-role dispatch included, at
`<canon>/.superpowers/sdd/slot-prompt-<round>-<slots>[-<repo>].md` (`<slots>` the bundle's ids
`+`-joined, `-<repo>` only on one repository's dispatch of a multi-repository panel), made of, in order:

1. the line `Read <skill-dir>/slot-brief.md in full first — every paragraph binds you`, then the resolved set its WORKTREES paragraph lists, and on one repository's dispatch the repository its SECTION paragraph reads; on a fix round, the FIX-ROUND SCOPE blockquote of `skills/flow/review-panel-fix-round.md`, naming the `<fix report>` path(s);
2. one `## PASS <id>` section per slot in the bundle's order: the fenced body of that slot's `<id>-reviewer-prompt.md`, dedented, with the six values of the placeholder table above filled with absolute paths, then that slot's REPORT FILE line (`panel-report-<round>-<id>[-<repo>].md`). `mutation` has no template: its section is **The mutation-testing brief** below, built last, naming its throwaway copies, `<worktree>-mutation-<round>` per repository.

The Agent call's description is `panel-<round>-<slots>[-<repo>]` and its prompt opens `Role: panel-reviewer` (`Role: panel-rerun` on a fix round). Its prompt carries only that Role line
(**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`), the baseline
pointer, `read <the assembled path> in full first — it is your brief`, and each role's REPORT FILE path — so a slot that answers without reading its brief still owes its file. The return message carries one findings summary per role under a heading naming the
role; the parent records each finding under that role.

**A defect raised by several roles is recorded per role and fixed once.** Findings from different
roles in the same round that cite the same `file:line` and name the same root cause are one
defect: the first in the decision's `panel` order is recorded as usual, and every other raising role's row is
recorded under its own `-slot` with `-supersedes F<first>`. Before the fix dispatch the parent folds them: the fix
subagent receives the defect as one structured block, the first row's, listing every linked
`F<n>`; only the first row's role re-runs on it (**Panel re-runs**,
`skills/flow/review-panel-fix-round.md`); and each linked row takes the first row's verdict — one
`flow record status` call per linked `-ref`, made in the same step as the first row's, with the
same status. Findings that differ in
location or in root cause are never folded, however alike their prose, and neither is a row whose
`-supersedes` names a row of an earlier round: that is the same defect re-raised by a later fix
round, and its own role re-runs on it.

**A finding recorded under a joined slot that a per-role re-run must cover is withdrawn with the chain named, and re-recorded under each raising role.** The joined row's `-slot` names the bundle and no single role, so a fix round's slot coverage (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`) has no row it can close. Withdraw it — `flow record status -change <name> -ref F<n> -status 'withdrawn <reason>'`, the reason naming the chain: the joined row it leaves and the per-role re-run that needs single-role rows — then record the defect once per raising role, `flow record finding` under each role's own single-role `-slot` with the bundle's `-dispatch-seq`.

**Re-runs are re-grouped by the same bundles**, carrying only the roles re-running this round — a
bundle whose other members are clean dispatches with its re-running members only; the one-per-role fallback above separates them one per role when the
bundled findings need separation (**Panel re-runs**).

**Every slot's dispatch is recorded** — on claude-code by the hooks alone, which begin each row at
launch and end it when its agent stops (**Dispatch sites — the parent's closed list**,
`skills/flow/implement.md`); on any other harness by the parent, with this pair:

```bash
flow record dispatch begin -change <name> -role panel-reviewer -slot <slot|slot+slot+slot> -model <m> -effort <e> \
  -key panel-<round>-<slot|slot+slot+slot> \
  -session-token mf-<literal-token>
flow record dispatch end -change <name> -key panel-<round>-<slot|slot+slot+slot> \
  -session-token mf-<literal-token> -outcome completed \
  [-ended-at "$(date -u -r <report-file> +%Y-%m-%dT%H:%M:%SZ)"]
```

Every dispatch of the round is launched in one message; the round's wait is the harness's completion notifications, never a poll (the turn discipline of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`).
A parent-recorded `end` — any harness but claude-code, whose daemon reads no agent transcript — passes its own slot's report-file modification time as `-ended-at`; a bundled dispatch's `<report-file>` is the newest of its slots' reports.

**A live check's before figures are measured before the round's launch:** when
`design.md`'s `## Live check` names something to exercise, the parent measures each before figure
from the merge base the way **Live check** (`skills/flow/verify-and-handoff.md`) states and writes
it to `<abs-worktree>/.superpowers/sdd/live-before.md` with the merge-base sha it was measured at — never into the change's directory, a
tracked path the suite's tree gate reads.

Every round brackets itself with `git -C <worktree> status --porcelain`, captured in the Bash call
before the round's launch and again before any report is read or finding recorded: the two must
match, paths under `<project>/.superpowers/` aside, and a mismatch names every path a slot left
changed or untracked and stops the round.

`-slot` names the dispatch's roles from **The panel** table above. `-role` is
`reviewer` for every one; `-task` is omitted. `-model` and `-effort` are the dispatch's role's pair
— `panel-reviewer` in pass 1, `panel-rerun` on a fix-round re-run — or this run's override.

**`-agent-id` is never typed on a slot's row — never invented** — the daemon captures the launch identifier,
stamping each launch onto the row its Agent-tool `description`'s first word names in the launching session — each slot's description opens with its row's `-key` (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`) — and pairing a launch whose description names no row with the begin nearest it in time; never reuse a `-key`.

**Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
printed — `recorded: dispatch <seq>` — into each of that slot's `flow record finding` calls as
`-dispatch-seq <seq>`.

**A reproducer for a finding whose defect may be target-specific is authored and run on the target
the defect manifests on — the target `flow.verify`'s verifier drives, never the test
target alone.** A reported behaviour that will not reproduce on the JVM/desktop test target after
honest attempts is not thereby nonexistent: the same defect can live only in the browser/wasmJs
build the verifier drives. Author the reproducer to drive that target the way the finding
describes — real input events (`page.mouse.wheel()`), a fresh account, data seeded past the
resting state — and a finding closes on that target's evidence, never on the test target's clean
exit alone.

### No forking

No panel slot is dispatched onto a skill or agent that forks its own background agent — a forked
agent reports to nobody the dispatcher is tracking. Repair it by dispatching the slot on a shape
that reports back directly; never drop the slot.

A slot whose dispatch failed or returned no parsable report is `not run`, never zero findings, and
its round is never clean.


### The findings merger

**The parent reads a round's findings through one findings merger, never the reports one by one.**
**A section a report leaves unread is read in the same round, before the merger.** When a
round-`<r>` report's `Unread:` line names a section, the parent dispatches one more dispatch of that role in round `<r>`, keyed
`panel-<r>-<id>-unread` with `<id>` the report's own id, its prompt carrying the SECTION paragraph
for each section the line names, and its report written to `panel-report-<r>-<id>-unread.md`.
Once every dispatch of the round has ended — its `end` recorded, every role's report file
confirmed under **Recording findings**, and the worktree-status comparison passed — the parent dispatches one merger: a one-shot that
reads the round's report files from disk, writes one summary and exits.

**The merger never judges, fixes or dispatches.** It decides no finding real, records nothing,
and edits no file but its summary. Whether each finding stands, the fold of a
defect raised by several roles, every `flow record finding`, `status` and `incident` call, the fix
and every re-run decision stay the parent's.

**The merger runs on the `panel-merger` role's pair** — its work is mechanical, and every review
stays on its slot's own pair. Its row is begun and ended by the hooks.

It is launched with `run_in_background: true`, its `description` opening with its key, and a
prompt opening with the line `Role: panel-merger`, its pair its **Allowed pairs** row
(`skills/flow-contracts/model-policy.md`) — followed by the baseline pointer, the line `Read <skill-dir>/slot-brief.md` for its TOOLS and NO DELEGATION paragraphs, the
MERGER BRIEF paragraph below, every role's report path, and one line per dispatch of the round —
its `-key`, the `-outcome` its row was closed with, and its elapsed seconds.

> **MERGER BRIEF:** read each report file listed below — never the diff, never the code — and
> write `<abs-worktree>/.superpowers/sdd/panel-summary-<round>.md`, then return its content and
> stop. One line per fact, in exactly these shapes:
> `dispatch: <key> — <outcome> — <elapsed>s` for each dispatch listed below, as given;
> `slot: <id> — <report path> — Critical <n> — Important <n> — Minor <n>`, the counts its
> report's `Critical:` and `Important:` lines give and its own Minor count, or `placeholder` in
> place of the counts for a `no verbatim report captured` file;
> `finding: <id>#<n> — <severity> — <file:line> — <task id|none> — <the reviewer's own sentence naming the defect, quoted>`, `<n>` the finding's
> 1-based index in its report;
> `duplicate: <file:line> — <id>#<n>, <id>#<n>[, …]` grouping findings of different roles that
> cite one `file:line`. You judge nothing: never decide a finding real, join findings at different
> locations, drop one, re-word one, or fix anything.

**The summary is what the parent reads**, not the reports: its `finding:` lines carry each
finding's severity, location, task and the reviewer's own sentence, the `-note` of
**Recording findings**. A `duplicate:`
line is a candidate fold the parent confirms or rejects under **Bundled dispatch**'s
several-roles rule.

### The mutation-testing brief

Wherever the panel dispatches Mutation, the dispatch prompt carries a mutation-testing
brief: for each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a
boundary, remove a branch, move an interaction off its target, overlay an earlier commit's tree and
run the tests — and establish whether an existing test fails. A mutation only counts once its edit
landed: every mutation must confirm its edit landed before the tests run — the target changed
where the slot intended, not nowhere and not somewhere else. An edit that never applied must be
redone with a working mechanism: it is a refusal, never a **surviving mutant**, and it never buys
a test. A mutation no test catches is a
**surviving mutant**, an ordinary finding that blocks the handoff exactly as any other, unless it is
withdrawn with a reason under the fix round's handback. A reproducer never `cd`s — it runs each tool against its directory with `go -C`, `make -C` or `npm --prefix` and names every file as `$PWD/<path>` — and makes its edit through a patch file or a fixed-string replacement, never through shell quoting nested inside another quoted string. A reproducer adds no file under `$PWD`: a Go test it adds is written under `mktemp -d` and passed to `go test -overlay <json>`, whose `Replace` map names the test's path in the package, a Go mutation goes through the same overlay, its `Replace` map naming the mutated source's path and its mutated copy under `mktemp -d`, and every other file it writes lives in its own `mktemp -d` directory. Its test command filters by the name prefix the mutated code's test family
shares (`-run '^TestCheckPanel'`), never by a full test name, so a test the fix adds under that
prefix still kills the mutant and the run's cost does not grow with the package; only when the
family shares no prefix does it run every test of the package, file or module the mutated code
belongs to.
Every test run of a mutant passes the test runner's no-cache flag — `-count=1` for `go test` — because a cached result never sees a mutated file a test reads through a subprocess.
The slot's own mutant runs are scoped before the first mutation too: it measures one run of the mutated code's whole package, and when that run's seconds times the number of mutants it plans passes 15 minutes, every mutant run filters by the name prefix the mutated code's test family shares (`-run '^TestCheckPanel'`), never a full test name (group-allowed-pairs-roles self-review: the slot ran the whole guard package per mutant and timed out at 15 minutes; the prefix-filtered retry finished in 88 s).
The slot never forces a full rebuild of its copy — no `--rerun-tasks`, `clean`, cache-busting flag or deleting build output — whatever the build tool; it measures one scoped run before its first mutation; and it runs one build at a time in its copy, never starting another while an earlier one, a backgrounded one included, still runs.
A Gradle run that fails with `Timeout waiting to lock journal` is waited out and retried with backoff, a few minutes in total — the lock sits in the shared Gradle home, held by another worktree's daemon; never run `gradle --stop` or kill a daemon another worktree owns, and if the lock is still held the mutant is reported `unproven (journal lock held by <pid from the message>)`, never skipped silently.
Never symlink a shared `node_modules` into your copy: run `npm ci` inside the copy, which installs its own — over a symlink it empties the shared directory.
Never run `git worktree add`: you work only in the copies your prompt lists, which the parent creates and removes (**The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`); a comparison against the base reads `git show <base>:<path>` or `git diff <base> -- <path>`, never a checkout of your own.

### Principles

Principles, when dispatched, is the panel's judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files.

**Resolve `[PRINCIPLES_PATH]` before dispatching the principles slot.** It is the **absolute** path
of `engineering-principles.md` **beside this file** — `skills/flow/`, always. Under the global install that is
`~/.claude/skills/flow/engineering-principles.md`; under a
project-local install it is `<project>/.claude/skills/…`.
Resolve it from where this file was actually read — never hardcode a repo-relative
`skills/…` path: the subagent's working directory is the project worktree, which has no
`skills/` tree, so a relative path fails to open and the reviewer loses its principle
list. Confirm the file
exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching the principles slot**, from the entries
`flow-guard project-get <worktree> standards` prints (exit 1: none declared), resolved per the entry-form
and containment rules of **Project configuration — standards** (`skills/flow-contracts/project-configuration-standards.md`),
never by reading the template. Pass an **empty** value when none resolve.
Record which standards files were passed, or that none resolved.

## Recording findings, and the record's format

**The slot writes its own report.** Each slot writes
`<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` itself, per the REPORT FILE paragraph
its prompt carries — `<round>` the same value that round's findings carry on `-round` (`0` initial,
`1..n` fix rounds), `<id>` the resolved reviewer id, never the slot display name. As each slot's
`flow record dispatch end` is recorded, run **The throwaway worktree**'s (`skills/flow/review-panel-optional-slots.md`) fold-back for that slot's
copies first, then confirm the file exists and is non-empty (`test -s`); when it is not, write it
yourself carrying the single line `no verbatim report captured — <reason>`.
Every dispatched slot ends up with one, a slot that raised nothing included. **Never re-emit a
slot's report from this context** — record its `F<n>` rows and cite the file, per **Read
discipline**'s never-`cat`-a-report rule (`skills/flow/implement.md`).

**Every finding is a row in the store. The panel record is rendered from those rows.** The parent
records every finding itself. Every
finding a round raised is recorded in one Bash call, `flow record finding -change <name> -from
<abs-worktree>/.superpowers/sdd/findings-<round>.jsonl`, over a file the parent writes first: one
JSON object per finding, its keys the flags below without their dash, a flag the finding omits
left out — never a shell loop of single `flow record finding` calls.

```bash
flow record finding -change <name> -ref F<n> -round <r> -slot <slot> -severity <sev> \
  -location <file:line> -status open [-reproducer <command>] \
  -dispatch-seq <seq> [-task <id>] -note <the finding>
```

`-round` is `0` for the initial panel and `1..n` for a fix round. `-task` is the task id the
finding names, omitted where it names none. `-ref` is unique within the change, and a finding
recorded under a ref the change already holds rewrites that row. **A run's first finding ref is the
one `flow record next-ref -change <name>` prints**, never `F1` carried from session memory or an
earlier run's numbering, and each later finding of the run is one more. **`-note` carries the reviewer's own sentence naming the
defect, not a dispatcher restatement.** Where the reviewer's wording runs long, quote the sentence
that names the defect and leave the rest to the report file. **A fix round updates the finding it
resolved rather than appending a second row:**

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**`-status` carries the whole status text the marker line shows** — a withdrawal passes its reason
with it: `-status 'withdrawn <reason>'`.

**Render the record when the panel closes** — every slot's result clean, no finding open — into
the canonical worktree, never into whichever worktree the closing pass runs in, so a
multi-worktree run leaves one copy:

```bash
flow record render -change <name> -kind panel -repo <canonical-worktree>
```

**A panel that raised nothing still renders**, declaring `findings-total: 0`. There is no
skip-the-render shortcut.

**A `withdrawn` marker's reason is checked for being there at all.**

## Panel re-runs

**Load `skills/flow/review-panel-fix-round.md` only when** a round has recorded a Critical or
Important finding, before that round's fix opens, or a section of it is cited at the point of use (the full-suite fix, the gated
reviewer's fix, the stage-diff check) — it carries every fix round's procedure: the
fix commit routes, which slots re-run and on what, the round close, the fix chunks and the
non-convergence loop, while the fix subagent's dispatch paragraphs and dispatch record stay in
this file.

**Every Critical and Important goes to the fix; every Minor is fixed too, and a Minor never
causes a fix round of its own.** Every Critical and Important the round raised goes to the fix
subagent below, closed by the verification that follows it — the fix diff touches a path the finding named, or one of the path condition's accepted
alternatives below closes the finding. **A round that raised a Critical or Important sends every
Minor it raised to that same fix**, closed by the same verification. **A round that raised no
Critical and no Important has its Minors fixed by one fresh subagent at the round's close** — role
`minor-fix`, key `panel-<round>-minor-fix` (**Dispatch sites — the parent's closed list**,
`skills/flow/implement.md`), its pair from `flow-guard pair-judge minor-fix`, its prompt carrying
every fix subagent paragraph below but REPORT FILE, its open Minors as the
findings, and `REPORT FILE: <abs-worktree>/.superpowers/sdd/minor-fix-report-<round>.md` — no
slot re-run. It edits each Minor's named location, stages the changed paths and makes one
commit per worktree at the branch tip —
`git -C <worktree> commit -m "fix(<module>): review Minors" -m "Flow-Key: flow.review-panel" -- <the changed paths>`, the scope
naming the module the edits moved. On its return the parent pushes that commit per **Branch
backup** (`skills/flow-contracts/git-boundaries.md`), ends the row with it as `-commit`, and
records each finding `flow record status -change <name> -ref F<n> -status fixed`. `flow.verify`'s lint and tests,
which run after this stage, are that fix's verification. **A Minor is never deferred, and it names a defect only
when it completes this sentence with a Y that is true at HEAD**: given `<an input, a state or a
reader>`, `<location>` `<does or says Y>`; it should `<Z>`. One that cannot — a suggestion, a
hardening against a future edit, a tradeoff the code already states, a test remark naming no
change that would survive, a layout the project's formatter and lint accept — is withdrawn with
`-status 'withdrawn names no defect: <the half of the sentence it lacks>'`, as is one no change to
the tree can resolve, with `-status 'withdrawn <reason>'`, the reason one clause naming the
mechanism — out of scope, pre-existing and cosmetic are not reasons. **Where Y is a result shipped
code produces at run time — an output, a stored value, an exit code, a permission, a request made —
the Minor fix carries one test that fails without it, in the same commit**; where the cause cannot
be forced in one isolated run, the commit body says so. Then proceed to
the stage close. **No reviewer is dispatched against a
`review Minors` commit alone**: a further round opens only when `git log <the round's dispatch
sha>..HEAD` holds a commit with another subject — a task, an in-run fix, a Critical or Important
fix. A Minor a delta re-run raises on a line its fix round wrote is closed the same way, at that
round's close. A fixed finding that fails verification takes the handback below, and that loop
re-runs no slot either.

**A rule that changes mid-run governs from the round it lands in, and each round's pass log names
the rule it decided under.** When the wording of a rule this file states — the Minor
rule above included — changes while a panel is in flight, by an operator instruction or by an
edit to this file, the new wording governs from the first round decided after it lands: a round
the old rule already closed stands as closed, never re-decided retroactively. Each round decided
under wording that changed during the run records that wording beside its decisions, with the
round's own pass-log rows — `flow record pass -change <name> -round <round> -note 'decided under:
<the rule as this round applied it>'` — so the rendered pass log shows which rule governed what.

Handoff still requires **zero open findings at any severity** from every agent that has run, and
no stale result — where **a slot's clean result is stale when the rule above required that slot to
re-run and it has not, or when any commit or working-tree change to source landed after that slot's
last read, from any stage — `flow.verify` included**. An unrecorded edit after the panel closes is
stale by definition, not only one a fix round produced. A fix against which a slot raised no
finding leaves that slot's result current: a fix's own report covers what it changed. A fix of Mutation's own findings leaves Mutation's result current too (**Panel re-runs**,
`skills/flow/review-panel-fix-round.md`). **The `review Minors` commit leaves every slot's result current** — a source change
after a slot's last read that does not make that result stale, and neither does a fix's delta
for a slot its one reviewer did not read (`skills/flow/fix.md`).

**Each fix chunk's prompt is the fixer stub** (**Every fix is a fresh one-shot**,
`skills/flow/implement.md`): `Role: panel-fix`, its `FIX INPUT:` the chunk's structured blocks, each
naming its slot's report path, and its REPORT FILE
`<abs-worktree>/.superpowers/sdd/panel-fix-report-<round>[-<n>].md`; the brief it reads is
`skills/flow/fixer-brief.md`. **A restore that would discard unstaged edits is denied** (**Restoring
a path with unstaged edits**, `skills/flow-contracts/git-boundaries.md`).

**The fix subagent's own dispatch is recorded too** — the hooks begin and end its row, `-role
panel-fix`, the end's `-commit` the fix commit its summary head names, carrying the round's
`Flow-Key:` trailer.

Then the stage ends:

```bash
flow stage end -command '/flow' -stage flow.review -outcome completed <name>
```

Once this stage ends clean — zero open findings at any severity, no stale result as **Panel
re-runs** defines it — continue into
`skills/flow/verify-and-handoff.md`.
