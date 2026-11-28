---
name: flow-interactive
description: Conversational /flow session — one invocation is one change, one Jira issue and one session worktree, and a stream of requests the main session coordinates. Each request is looked at live by the coordinator, then one pair-judged operator-asked subagent — up to three at once, each in its request's own worktree — implements it test-first, verifies it, lands it by merge and push, reinstalls and, when it touches ui paths, checks the live page; an explicit stop runs one self-review pass, cleanup and FINISHED. Every request is recorded in flowd; no spectre artifacts or state file. Use for /flow-interactive.
allowed-tools: Bash(flow:*)
license: MIT
---

One `/flow-interactive` invocation is one **session**: one store change, one Jira issue, one git
session worktree, and a stream of **requests**, each shipped to the base branch on its own from a
worktree of its own. The session
marks existing `flow.*` stage keys, so labels, phases, cost by stage and the timeline read as they
do for `/flow`. Each request is a row the store keeps (`flow record request`),
and every stage run made for it carries its request number (`-request <n>`). The main session is
the **coordinator**: it looks, records, dispatches and reports, and does nothing else. A request that
changes code is executed whole — implement, verify, land, reinstall, UI check — by one
`operator-asked` subagent (**3. Execute**), up to three of them at once. No review panel per
request, no spectre artifacts, no plan file, no decision, no state file, and no workspace isolation
beyond each request's worktree.

**Announce at start:** "Using flow-interactive for change `<name>`."

**No flags.** The only argument is empty, a first request, a Jira key, or the exact name of an open
interactive change; report anything else rather than ignoring it. The base is the branch
`origin/HEAD` names; from the worktree's creation on, `<base>` is what
`flow-guard resolve-base-branch <worktree>` prints.

**Every mark is a literal call**, carrying no session token: the CLI infers the calling agent's,
per **Stage marks** (`skills/flow-contracts/pipeline.md`). `<n>` is the request's number in the
session.

**Guardrails, the whole list.** Never use git stash pop/apply/drop; if a worktree looks wrong, stop and report BLOCKED with the exact `git status` output. Never dispatch a subagent, save the request's own executor
(**3. Execute**), the self-review pass's own
dispatches (`skills/flow-self-review/SKILL.md`), the in-run fix loop of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`), and a subagent the operator asks for in their
own words (**Operator-asked subagents**, below). Never write `<project>/spectre/`, save those
passes' own `self-review-<name>` and `pipeline-fix-<name>-<k>` changes, or a state file. Never push
to a branch other than `<name>`, `<name>:<base>` and `<name>-r<n>:<base>`. Never set up workspace
isolation for the session worktree; a request worktree gets its own through the project's
`## workspace isolation` (`skills/flow-contracts/workspace-isolation.md`), removed with it.

## Stage keys

Every stage `/flow-interactive` marks uses a `flow.*` key — the same keys `/flow` marks. The full
key list, in the order each section below marks them:

| Step | Keys |
|------|------|
| **Session start** | `flow.plan` |
| **One request** | `flow.plan`, `flow.implement`, `flow.verify`, `flow.land` |
| **Explicit stop** | `flow.self-review`, `flow.cleanup` |

Every **One request** stage is a `flow stage begin … -request <n>` / `flow stage end … -request <n>`
pair; the stop's stages carry no `-request`. Each executor makes its request's pairs from that
request's worktree; a begin supersedes and an end closes only its own request's run, so executors
in flight together never close each other's runs.

## Session start

**Argument.** A key or name matching an open `/flow-interactive` change resumes it (**Resume**,
below); one matching a finished change ends the run with one line saying so. Anything else starts
a new session, save as **Name.** below states, a non-empty argument being its first request.

**Session start comes first, whatever the argument says.** A non-empty argument — a question
included — is the session's first request and is looked at, answered or implemented only once
**Session start** (Jira, **Name.**, **Worktree.**, the `flow.plan` mark) is complete and the request
is recorded. A turn that has answered the argument without a session behind it has skipped the
skill.

The invocation's first tool call is the open-session lookup (`flow state list`); no answer, look,
edit or other tool call comes before the `flow.plan` mark and the recorded request row. A session
started late is never repaired backwards — no hand-made worktree, no backfilled marks, no direct
push to the base: stop and run **Session start** from its beginning. A plain message mid-session is
recorded as a request before anything else, and code changes go to the executor, never the
coordinator.

**Jira.** Resolve the issue per **Resolution (how `jiraIssue` is decided)**
(`skills/flow-contracts/jira-integration.md`): with no key given, reuse the In Progress `Interactive session` issue, whoever it is assigned to, when one exists, else file a Task through its creation
path, summary `Interactive session <YYYY-MM-DD>`, labelled per **Labels and body of issues the pipeline
creates** (`skills/flow-contracts/jira-integration-finish.md`), unless the argument declines an
issue ("no jira"). Move it to **In Progress** per **Transitions** there. Every Jira failure is the
one `⚠ Jira: skipped — <reason>` line of **Never blocking**.

**Name.** `<lowercased key>-<slug>` per **Change naming**; with no issue,
`interactive-<YYYYMMDD-HHMM>` (UTC). **An open `/flow-interactive` change whose name is, or starts
with `<lowercased key>-`, the resolved issue's key is resumed whatever its slug** (**Resume**, below) —
the reused `Interactive session` issue's included, so a topic-slugged request never opens a second
change on the same issue — never a new session numbering from 1. Find it with `flow state list`: every record whose name is, or starts with `<lowercased key>-`, and whose state is not `FINISHED` — the most recently updated when several. (`flow state find` takes one exact name and never matches by key.) **Whatever state the record shows is open: a `STARTED` row whose `updatedBy` is `flow stage begin (synthetic)` is a session's own row — a stage mark creates it, as no state write exists before the stop — so it is resumed, never judged another command's and never left beside a new name**. Looking before creating is not optional: a session whose worktree is not the one found is a second session on the issue.

```bash
flow stage mark  -command '/flow-interactive' -stages flow.plan <name>
```

**Landing route.** `flow-guard project-get <project> "default landing route" --enum "pull request" "merge
and push" manual` must resolve `merge and push`. Absent (exit 1), malformed (exit 3) or any other
value ends the run with one line naming `## default landing route` as the section to declare, and
so does `flow-guard project-get <project> handoff --enum required none` resolving `required` — that
project asks to test before anything lands.

**Worktree.** One session worktree, `<project>-worktrees/<name>` on branch `<name>`, created from the
main checkout, with `<base>` the branch `origin/HEAD` points at — with no `--base`:

```bash
git fetch origin
git worktree add <project>-worktrees/<name> -b <name> origin/<base>
git -C <project>-worktrees/<name> push -u origin <name>
```

Then the project's declared `## worktree setup` (**Project configuration**,
`skills/flow-contracts/project-configuration.md`) runs once, from the worktree root, before
anything else touches the tree — the fenced command lines only, in order, in the foreground; the
key's exit 1 (absent) continues with nothing run, its exit 2 stops the run, and a command's
non-zero exit ends the run naming the command and its output — each only for a cause out of the
run's reach; any other is cured and the setup re-run (**Fewest operator actions**,
`skills/flow-contracts/pipeline.md`). A re-run reusing an existing
worktree runs nothing here — the creating run already ran it. Then `mkdir -p <changeRoot>`, where `<changeRoot>` is
`<project>-worktrees/<name>/.superpowers/sdd/<name>/`: the record verbs find the change's checkout
by that directory, and it holds the session's edit scripts. The main checkout is never checked out, staged, committed or written.
The session worktree holds the coordinator's briefs and the explicit stop's self-review fix
commits; each code request gets a worktree of its own (**3. Execute**).

**Hook registrations.** Grep the harness's hook config for every identifier `<agents repo>/setup.sh`'s
`warn_unregistered_hook` checks for that harness (on Claude Code: `enforce-agent-baseline`,
`enforce-loop-dispatch`, `flow-active-change`, `flow-session-activity`, `await-own-agents`,
`block-unsafe-commands` in `~/.claude/settings.json`). Any missing gets one line naming them and
`<agents repo>/setup.sh global`, which prints each registration snippet. Never edit that file, ask
nothing about it, and continue.

Read `<project>/.flow/project.md`'s `## lint`, `## test`, `## visual verification` and `## run`
once, with `flow-guard project-get <project> <key>`. The project's instruction file is already in this
session's context; never read it here.

## One request

**The request in order.** Each step is stated in its own paragraph below; this list only orders
them.

| Step | Run by | Next |
|---|---|---|
| Record the request | the coordinator | 1 |
| 1. Look | the coordinator | big → **Big requests**; needs a brainstorm → 2; changes code → 3; else 8, `answered` |
| 2. Brainstorm, rarely | the coordinator | 3 |
| 3. Execute | the coordinator: request worktree, brief, `pair-judge`, dispatch | 3a |
| 3a. Implement | the executor | 4 |
| 4. Verify | the executor | holding the landing grant → 5; else `ready:` and wait for it |
| 5. Land | the executor, under the grant | 6 |
| 6. Reinstall | the executor, under the grant | `check-visual-trigger` matches → 7; else its report |
| 7. UI check | the executor, under the grant | a defect → a fix commit, then 4; a pass → its report |
| 8. End | the coordinator | the worktree removed, the grant passed, the queue's head dispatched |

Every operator message that asks for something is a request — a question included ("why don't I
see X") — recorded at once, `<n>` one past the highest recorded:

```bash
flow record request begin -change <name> -n <n> -title '<one line>'
```

A message carrying several distinct asks — a pasted list, numbered or bulleted — is recorded at
intake as one request per item, its `-title` that item's text on one line, before anything is
dispatched; an item needing no code ends `answered`. A request is ended only `landed`, `answered`
with what was found, or `dropped` with `-reason`, stored on the row for the operator to read; that a list is
worked through, done or all landed is derived from those rows — `flow-guard check-open-requests
<worktree>` printing nothing — never from memory.

A message that changes an in-flight request's own ask is applied to that request instead.
Outside a brainstorm, a question carrying a recommended option is taken, never asked
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). A pipeline
defect met mid-session is fixed per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`).

**1. Look.** The live state behind what the operator saw first — the rows the page reads, the page
itself, the log — then the code. Never answer from code alone.

```bash
flow stage begin -command '/flow-interactive' -stage flow.plan -request <n> <name>
flow stage end   -command '/flow-interactive' -stage flow.plan -request <n> -outcome completed <name>
```

**Big requests.** Once looked at, judge the request's size. It is big when it would not land as
one small step: it restructures how a module works, spans several modules, adds a stage or a
schema, or needs a plan of its own to be done well. A big request is put to the operator through
**AskUserQuestion**, one line saying why it reads as big, with two options and neither recommended
— **Run it here** or **File a Jira task** — so **Auto-resolution** never takes it. **A Jira task is
never filed unprompted:** only the operator's answer or words create one. A request whose work lands
in another repository of the operator's (gymie's `.flow/project.md`, say) is not big for that reason:
it is done there in a short-lived git worktree of that repository, off its base, committed with the
module as scope, tested as that repository's own instructions say, pushed to its base and the
worktree removed — the one push this session makes outside `<name>`; the report names the repository
and commit. Filed,
it is one Task, created and labelled as **Session start**'s **Jira** paragraph creates one, its
description the ask and what the look found; the request ends `answered`, naming the key. Run
here, it continues at step 2.

**2. Brainstorm, rarely** — only when the request needs a new table, endpoint or layout, spans
more than one module, or two readings lead to materially different work. Record the reason, then
ask the operator through **AskUserQuestion**, per **Unanswered mid-run asks**
(`skills/flow-contracts/pipeline.md`):

```bash
flow stage begin -command '/flow-interactive' -stage flow.plan -request <n> <name>
flow record request begin -change <name> -n <n> -brainstorm-reason '<why>'
flow stage end   -command '/flow-interactive' -stage flow.plan -request <n> -outcome completed <name>
```

**3. Execute.** A request that changes code is executed by **one fresh subagent**, never by the
coordinator: it implements, verifies, lands, reinstalls and UI-checks, steps 3a–7 below. A request
needing no code — a question, a lookup — ends `answered` with no dispatch and no judge call,
whatever is in flight.

**Request worktree.** Each code request `<n>` is executed in its own worktree `<wt-n>`,
`<project>-worktrees/<name>-r<n>` on the local branch `<name>-r<n>`, which the coordinator creates
from the main checkout at dispatch:

```bash
git fetch origin
git worktree add <project>-worktrees/<name>-r<n> -b <name>-r<n> origin/<base>
```

The executor runs the project's `## worktree setup` there, exactly as **Worktree.** runs it, then
`mkdir -p <wt-n>/.superpowers/sdd/<name>/` — that request's `<changeRoot>`, holding its edit
scripts and its step-7 screenshot. Its environment is its own: `run-verify-list` and
`flow-guard prepare-workspace` derive the request's database, bucket and ports from the branch
`<name>-r<n>` per `skills/flow-contracts/workspace-isolation.md`, and whatever starts the
applications creates them on demand. Once the request is ended (step 8, any outcome), the
coordinator removes it all — the worktree's processes, its workspace, the worktree and the branch:

```bash
flow-guard workspace-remove <project> <name>-r<n>
```

Exit 1 (a dirty or unlanded worktree, or a failed removal) leaves the request ⛔ blocked with the
guard's lines relayed.

**In flight, cap and queue.** The in-flight count is the number of `git worktree list` entries on
`refs/heads/<name>-r*` — derived, never stored, so it survives compaction and **Resume**. A code
request is dispatched, in the background, when the count is below 3 and its step-1 paths to touch
overlap no in-flight request's brief; otherwise it waits. The queue is the recorded code requests
with no outcome and no `<name>-r<n>` worktree, in number order; each worktree removal dispatches
the queue's head, when it now meets both conditions.

The coordinator writes the brief to `<abs-worktree>/.superpowers/sdd/pair-judge/operator-<k>.md`
(the ask, what step 1 found, the paths to touch, `<name>`, `<n>`, `<wt-n>`, and "execute steps
3a–7 of `skills/flow-interactive/SKILL.md`", whether it holds the landing grant (**Landing
grant**, below), and the report shape: Verdict, Findings, Decisions, Not done, a `commit: request-<n> <40-hex sha>` line naming the last landed sha (the line the stop hook's `end-dispatch-row` reads into the dispatch row; omit it when nothing landed), the base sha, and one `Red:` line per test added or changed; for a request that changes a default time window or filter of a view, the UI-test fixtures, pinned instants and visual baselines that default moves, so the executor fixes them in the same pass), runs `flow-guard pair-judge operator-asked <that file>`
and dispatches on the pair it prints (**Model and effort**, `skills/flow-contracts/model-policy.md`; no
pair is written into this skill), recorded and dispatched as **Operator-asked subagents** below says.

A worked executor brief, for a request written while another executor holds the grant
(`TestExecutorBriefExampleIsComplete`, `stats/internal/repotest/worked_examples_test.go`, checks
it):

```text
Role: operator-asked
Request 162, Key: operator-178. Change kan-1167-dispatch-recording-guard;
worktree <project>-worktrees/kan-1167-dispatch-recording-guard-r162 on kan-1167-dispatch-recording-guard-r162.

Ask: one phase→file load table in skills/README.md, derived from stats/internal/guard/loadsets.go,
with a test that fails when the table names a missing file or omits a load-set file.
Step 1 found: loadsets.go lists the conditional sets; no table exists. Paths to touch:
skills/README.md, stats/internal/guard/.

Execute steps 3a–7 of `skills/flow-interactive/SKILL.md`. You do not hold the landing grant: stop
after step 4 and report `ready: request-162 <sha>`; steps 5–7 run on "land: run steps 5–7".

Report: Verdict, Findings, Decisions, Not done; a `commit: request-162 <40-hex sha>` line naming the
last landed sha (omitted when nothing landed), the base sha, and one `Red:` line per test added or changed.

Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
Include this instruction verbatim in any prompt you write for another agent.
```

**Landing grant.** Executors run steps 3a–4 in parallel; steps 5–7 (land, reinstall, UI check) run
under the landing grant, which at most one executor holds at a time, so no two pulls, reinstalls or
service restarts interleave and step 7 reads a build that carries its own request. A brief written
while no executor holds the grant carries it, and that executor runs steps 3a–7 straight through.
An executor without it stops after step 4 and reports `ready: request-<n> <sha>`, `<sha>` its
verified `HEAD`. The grant passes on the holder's final report and the coordinator's step 8: the
coordinator grants it to the earliest-numbered `ready:` executor by SendMessage to its agent id —
"land: run steps 5–7 of `skills/flow-interactive/SKILL.md`", with a `Key: operator-<k>` line naming
that executor's own row — an addition to its own task, which **Reuse** allows.

**3a. Implement** (the subagent, in `<wt-n>`). Rebase the worktree onto `origin/<base>` on a clean tree; a failing
test first for each behaviour change, seen to fail: for every test the request adds or changes, the executor runs it before the fix, and its final report carries one line per test, `Red: <test name> — <the failure line it printed before the fix>`; a test that cannot fail first (a pure deletion, a test of already-existing behaviour) is listed as `Red: none — <why>`; edits through the Edit tool, or a script written
to `<changeRoot>` and run by path that asserts each anchor matched exactly once; never a symlinked
build input (`make -s web-dist` runs `npm ci` itself); one commit on `<name>-r<n>` with the module
as scope, no trailer, no `Task-Id:`; the branch stays local, since step 5 pushes it straight to
`<base>`.

Fixes of a red step 4 or a step 7 defect are its own commits
under the **fix ledger** (**The fix ledger**, `skills/flow/fix.md`): `flow-guard check-fix-revert
<abs-worktree>` runs after each.

```bash
flow stage begin -command '/flow-interactive' -stage flow.implement -request <n> <name>
flow stage end   -command '/flow-interactive' -stage flow.implement -request <n> -outcome completed <name>
```

**4. Verify** (the subagent runs it itself).
Run `flow-guard run-verify-list <wt-n> lint` — it runs every command of
`## lint`'s fence, never the fence lines `flow-guard project-get` prints with them — and the `## test` commands scoped to what the
request touched — the packages, modules or test files the diff names, never the whole list; the full suite,
`flow-guard run-verify-list -full <wt-n> test`, runs only when the operator asked for it, under the operator's own `FLOW_OPERATOR_FULL_SUITE=1` (**Inline verify**, `skills/flow/verify-and-handoff.md`).

When the diff names nothing those commands can scope to, none of them
runs — the report states that empty-scope reason explicitly, the stated reason
being what makes the skip auditable rather than lazy, never a reflexive run of the whole list in its
place.

Then `flow-guard run-verify-list <wt-n> visual` runs the `## visual verification` `verify` row when the diff touches `ui paths` — the committed baselines a `lint` list never checks — under **A request's visual list** (`skills/flow/visual-verify.md`), loaded only then.

A failure is fixed, committed and re-run; the request never lands red, a failing test it did not introduce included unless its guard proves it pre-existing and unrelated (**the sweep**,
`skills/flow-contracts/known-bugs.md`).

The `lint` list,
exiting 0 on a tree clean before and after it, writes `HEAD` to the worktree's `git rev-parse
--git-path flow-lint-green`: the evidence the **Request-end guard** reads, so step 4 runs on the
request's last commit.

```bash
flow stage begin -command '/flow-interactive' -stage flow.verify -request <n> <name>
flow stage end   -command '/flow-interactive' -stage flow.verify -request <n> -outcome completed <name>
```

**5. Land** (under the landing grant). One call lands the request and marks `flow.land` around its
push:

```bash
flow interactive land -C <wt-n> -n <n>
```

Exit 0 prints `commit: <sha>` and `base: <sha>`; on the request's first landing that `base:` is the
request's **base**, the parent of its first landed commit. A rebase conflict is resolved per the
**Conflict** bullet of **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`), then
step 4 re-runs the `lint` list plus the `## test` commands scoped to the union of the request's diff
and the conflicted paths — not the whole `## test` list that file's **After a rebase that needed
resolution** runs, the full suite being the operator's alone — and the call is made again. A refusal naming `flow-lint-green` re-runs step 4 on the rebased `HEAD`, then the
call; a rejected push leaves the request open and the session says why (⛔ blocked).

**6. Reinstall** (under the landing grant). `git -C <project> pull --ff-only` when the main checkout is on `<base>` and
clean, then the reinstall the project's instruction file requires after a landing, each command
its own call with its exit code checked; a failed one is re-run once, and a service it restarts
must answer before anything reads it. A skipped pull is one line; on a request step 7 would
check, the request stays open (⛔ blocked), since the live page cannot show the landed code.

**7. UI check** (under the landing grant), when `flow-guard check-visual-trigger <wt-n>` matches the request's diff (`git diff
--name-only <base>..HEAD` on its stdin): open the live page — the link the operator gave, else the
URL `## run` serves, by the app's own run-view route, keyed by the project key and `<name>`, never a display name —
the executor opens it with the project's own installed Playwright through Bash — a short Node script written under `<changeRoot>` that navigates, reloads, takes the screenshot into `<changeRoot>` and exits — never an MCP browser tool, which a `flow-*` subagent does not carry —
`location.reload()`, screenshot, and read the screenshot against the ask.

When the request changes a view's default (a window, a filter), the executor also reads that view against real live data on the new default and confirms it is non-empty; fixtures alone do not pass.

A pass ends the stage
run `-outcome completed`. That end carries `-screenshot <path>`, the screenshot just read, saved
under `<changeRoot>`. A defect ends it `-outcome failed` with `-reason` naming the defect, and is a fix commit and steps 4–7
again inside the same request.

The stage's end is its own call, made after the screenshot is read —
never chained to its begin. The order is fixed: the stage's begin is made first, the page opened and the screenshot taken only after it, the screenshot read, and the end made last; a screenshot taken before the begin, or an end made before the screenshot, is refused by the Request-end guard (check 3) and the check is redone.

```bash
flow stage begin -command '/flow-interactive' -stage flow.verify -request <n> <name>
flow stage end   -command '/flow-interactive' -stage flow.verify -request <n> -outcome completed|failed [-screenshot <path>] [-reason '<the defect, one line>'] <name>
```

**8. End** (the coordinator, on the subagent's report: Verdict, the last landed sha, the base sha, the
screenshot path). Steps 1–7 are not optional and never reported as left undone: a request whose reinstall, UI check or record is outstanding is not ended — run them first, and a report never carries a "Not done" for a step this section requires. One of:

```bash
flow record request end -change <name> -n <n> -outcome landed -commit <last landed sha> -base <base sha>
flow record request end -change <name> -n <n> -outcome answered
flow record request end -change <name> -n <n> -outcome dropped -reason '<why, one line>'
```

`landed` runs the **Request-end guard**, below. Then the request's worktree is removed (**Request
worktree**), the landing grant passes and the queue's head is dispatched (**3. Execute**), and a
short report. A report may not call a list done while `flow-guard check-open-requests <worktree>`
exits 1; it names the `OPEN` rows instead.

A request that implements a self-review finding is not ended `landed` until its executor ran `flow self-review finding-implement <id> -commit <last landed sha>` (`skills/flow-self-review/SKILL.md`).

Before `landed`, when `git diff --name-only <base>..<commit>` matches `_test.go`, `.test.ts` or `.test.tsx`, the coordinator checks the executor's report carries the `Red:` lines of step 3a. A report without them is sent back to the executor exactly as the Request-end guard's refusal text says (**Reuse**, in **Operator-asked subagents**), never accepted and never worked around by the coordinator.

A resumed request whose base was never noted passes the parent of its earliest commit found on
`origin/<base>`, or any older ancestor: a wider range can only demand a UI check, never skip one.

A **mid-turn message** is recorded as the next request at once, looked at while executors work, and
dispatched or queued per **In flight, cap and queue**, in the same turn — never a stop. A turn ends
when the queue is empty and no executor is in flight.

## Operator-asked subagents

**The executor of step 3 is this section's subagent**, its pair from `flow-guard pair-judge
operator-asked` unless the operator names one. When the operator asks for a subagent in their own
words ("run it with an opus xhigh subagent"), the request dispatches it and counts it under this run: inside the request's open `flow.implement`
stage (the executor's own marks open it), record the row (role `operator-asked`, key `operator-<k>`), dispatch, and end the row when the agent reports — never a fork.
Each agent gets its own row: `<k>` is one past the highest `operator-` key the change records, so
two agents of one request never share a key. A dispatch the hook refuses ends its row `aborted` at
once.

```bash
flow record dispatch begin -change <name> -role operator-asked -model <model> -effort <level> -key operator-<k> -request <n>
```

The Agent call's `description` opens with `operator-<k>`, its prompt carries `Role:
operator-asked` and the baseline pointer, and its `subagent_type` / `model` is the judged pair, one the
`operator-asked` row of **Allowed pairs** (`skills/flow-contracts/model-policy.md`) admits; an effort
the operator names above the agent type's own (xhigh, max) rides on the Agent call's `effort`. The agent works in its request's
worktree `<wt-n>`, commits on `<name>-r<n>`; the executor of step 3 runs steps 3a–7 itself, an
operator-named agent's commits land through steps 4–8 as any other, under the landing grant. Its report is relayed, not re-told.

```bash
flow record dispatch end -change <name> -key operator-<k> -outcome <outcome> -agent-id <the agent's id> [-commit <last sha it landed>]
```

**Reuse.** An agent, finished or idle, is sent a message (SendMessage to its agent id) only for an addition to its own earlier task: the fix of its own refusal, a scope change to the same request, a follow-up to the same request, the landing grant of its own request. A new request is a fresh dispatch with its own brief, `operator-<k>` row and pair-judge pair, never a resumed agent, even while an earlier agent is idle in the same worktree.

Every end names the agent's id, an aborted row's included, so its tokens are measured; a `completed` end of a request that landed carries `-commit`, the last sha that agent landed. A row whose agent is given up on or replaced is ended `aborted` before the replacement's begin, so no row stays open with no outcome.

## Request-end guard

`flow record request end -outcome landed` checks, in-process and before the write:

1. The worktree is the checkout carrying `<changeRoot>` — the carrying linked worktree on branch
   `<name>-r<n>` for request `<n>`, else its first carrying linked worktree, else the main checkout
   when it carries it; none is refused. `<base>` is what the `resolve-base-branch`
   guard prints for it. `<project>-worktrees/<name>/.superpowers/` must be gitignored there, and
   `-commit` must track no path under it.
2. `-commit` must be an ancestor of `origin/<base>` — the local ref a successful push updates — and
   `-base` an ancestor of `-commit`. The request's latest `/flow-interactive` `flow.verify` stage
   run carrying `-request <n>`, by start, must have ended `completed`, whatever the diff touches.
   `-commit` must be the sha the worktree's `flow-lint-green` holds (step 4).
3. `git diff --name-only <base>..<commit>` is fed to the `check-visual-trigger` guard against the
   worktree's checkout of `<project>/.flow/project.md`. A `ui paths` match (exit 0) requires the request's latest
   `/flow-interactive` `flow.verify` stage run, by start, to have ended `completed`. Exit 1,
   or exit 2 with `VISUAL-TRIGGER-NOT-CONFIGURED`, requires nothing. Exit 2 with
   `VISUAL-TRIGGER-CANNOT-ANSWER` refuses, quoting its cause. That run's `-screenshot` must name a
   file whose modification time falls between its begin and its end (step 7).
4. With the store unreachable, checks 1–3's git work still runs; the stage-run reads are skipped
   with one `⚠` line, and the write is journalled like every record verb.

Exit codes, those of every record verb: 0 written or journalled; 1 refused, one line naming the
failed check; 2 usage. `-commit` and `-base` are required with `landed` and refused with
`answered` and `dropped`; `-reason` is required with `dropped` and refused otherwise; an `-outcome` outside the three is a usage error. A refusal is fixed —
land the commit, run the verify and the UI check and end each `completed` — and the end re-run, never worked around. The coordinator never repairs a refusal itself: a `flow record request end` refusal (verify or UI-check evidence, lint-green, ancestry) is sent back to the executor that ran the request — per **Reuse** in **Operator-asked subagents**, a fresh dispatch's brief being the refusal text — and the end is re-run on its new report.

## Explicit stop

Only a message whose intent is to end the session; anything that could also be a request is kept
as a request and the session stays open, saying so in one line. Stop:

1. Every in-flight request lands, one landing grant at a time (unless the stop drops it); queued
   requests not started end `dropped` and are listed in the stop report, each with its reason.
   `flow-guard check-open-requests <worktree>` must then exit 0 — an exit 1 lists the requests with
   no outcome, and the stop goes no further until each is ended. Every `<name>-r<n>`
   worktree is removed with `flow-guard workspace-remove` (**Request worktree**) before the
   session worktree's cleanup (item 4); an exit 1 stops the stop there, ⛔ blocked.
2. **Self-review**, with `-command '/flow-interactive'`.
   A `flow.self-review` stage already completed for `<name>` skips the pass: a re-run never
   records its findings twice.
   Assemble the bundle in `<sr-dir>/bundle.md` (**2. Read the bundle and run the seven angles**,
   `skills/flow-self-review/SKILL.md`), never printed, never committed, as the stdout of `flow
   self-review bundle -change <name>` run from the worktree — of the sources `/flow`'s bundle
   judges a change by, a `/flow-interactive` session can never have tasks.md, design.md,
   narrative.md, visual-verification.md and live-verification.md (no spectre artifacts), so those
   five report `skipped` by design, never silently; the ledger and panel render from the store
   when the session wrote rows.
   Over that bundle run steps 2–5 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`),
   canonical for the seven angles, the fix, where the fixes land, the filing and the store rows.
   The landing push carries the fix commits' shas per item 2 of step 3 of `/flow-self-review`.
   A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes the
   mark below `-outcome stopped` instead, and continues straight into the fix **the sweep**
   (`skills/flow-contracts/known-bugs.md`) names, in the same invocation. `flow self-review bundle -change <name>` renders the session's requests in number order
   as its git-log source. Its fix commits land on `<name>`; no report is written.
   The pass runs on every stop, a stop that landed no new commit included: it reviews the whole
   session's requests, never the stop invocation's diff.

   ```bash
   flow stage begin -command '/flow-interactive' -stage flow.self-review <name>
   flow stage end   -command '/flow-interactive' -stage flow.self-review -outcome completed <name>
   ```

3. Steps 4–6 of **One request** land those commits, their stage runs with no `-request`.
4. **Cleanup**: the tree clean and `git -C <worktree> merge-base --is-ancestor <name>
   origin/<base>` passing, then `flow change save -C <worktree> <name>` (the self-review bundle's
   only copy; a non-zero exit stops, removing nothing), `flow-guard stop-worktree-processes <project>-worktrees/<name>` (a dev stack or Gradle daemon started from the worktree), `git -C <project> worktree remove <project>-worktrees/<name>`
   without `--force`, `git -C <project> branch -D <name>` and `git -C <project> push origin
   --delete <name>`.

   ```bash
   flow stage begin -command '/flow-interactive' -stage flow.cleanup <name>
   flow stage end   -command '/flow-interactive' -stage flow.cleanup -outcome completed <name>
   ```

5. Jira: the description gains a `## Session summary` section — one line per request: number,
   title, outcome, commit — per **Description sync** (`skills/flow-contracts/jira-integration.md`);
   the issue moves to **Done**.
6. Move the change to `FINISHED` — the one state write a session makes, from the row the stage marks
   created, with the synthetic marker the read adds dropped (`state set` refuses it): `flow state get
   <name> | jq '.state="FINISHED" | .updatedBy="/flow-interactive close" | del(.synthetic)' | flow state
   set <name>`, then `flow state get <name>` printing `"state":"FINISHED"`. A stop that ends without
   it leaves a closed session's card on the board as `STARTED` for good.

## Resume

`/flow-interactive <key or name>` on an open session reuses its worktree — recreated from the
remote branch, else `origin/<base>`, when gone, with
`mkdir -p <changeRoot>` run in it — marks `flow.plan`, continues
numbering after the highest recorded request,
and takes the open requests from the dead session as in flight or queued: each is dispatched to a
fresh executor per **3. Execute**, its stages re-run as new attempts under the same number. An open
request's surviving `<name>-r<n>` worktree is reused; one gone is recreated from its local branch
when that survives, else as **Request worktree** creates it, with `## worktree setup` and
`mkdir -p <changeRoot>` run in it.
