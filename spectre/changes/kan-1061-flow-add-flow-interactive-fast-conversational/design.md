## Context

- Ticket KAN-1061 and its two operator comments; the brief
  `docs/briefs/flow-interactive-session-2026-10-29.md` — the hand-run session this command
  replaces, its working loop and its failure list.
- Design approved after two planning rounds (round 1: the operator took the recommended option on
  the board, self-review, implementer and guard questions; round 2: convergence approved).
- **Approved prototype — the specification of the board change:**
  `.superpowers/sdd/planner/prototype-board.html` in this change's worktree
  (`/Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational/.superpowers/sdd/planner/prototype-board.html`).
  The board's first card opens the run view. It is the source of truth: every departure the
  implementation, a reviewer or `flow.visual-verify` finds is fixed to match it. Every capture of a
  view it draws drives each interactive state it shows — the running and the waiting card, the
  card's hover and focus border, and the run view.
- `/flow-fast` (`skills/flow-fast/SKILL.md`) is the nearest command: a worktree for isolation,
  `flow.*` stage marks under one token, no spectre artifacts, its in-run self-review. A
  `/flow-fast` change row is the synthetic `STARTED` row a stage mark bootstraps, and its
  `flow.write-finished` mark is what moves it to `FINISHED` (`ApplyBeginStageMark`,
  `stats/internal/api/stages.go`).
- Every flow project's `.flow/project.md` declares `merge and push` as its default landing route
  except `gymie-playwright` and `spectre-e2e`, which declare none; none declares `## handoff`
  `required`. `origin/HEAD` names the landing branch in each (`develop` in the three gymie
  repositories, `main` elsewhere).
- Both prompt hooks ask the store for the session's latest stage run with `command=/flow`
  (`hooks/enforce-agent-baseline.py` `session_change`, `hooks/flow-active-change.py`
  `last_flow_change`), so a `/flow-fast` session's dispatches skip the begun-row check today.

## Design

### Shape

One `/flow-interactive` invocation is one **session**: one store change, one Jira issue, one git
worktree, and a stream of **requests**, each shipped to the base branch on its own. The command
marks existing `flow.*` stage keys, so labels, phases, cost by stage and the timeline already read
correctly. Each request is a row of a new store table, `requests`, and every stage run made for it
carries its request number. The main session does the work inline; no review panel per request, no
spectre artifacts, no plan file, no decision roll. The command takes no flags: the base is the
branch `origin/HEAD` names.

### Session start (`flow.kickoff`)

- **Argument**: empty, a first request, a Jira key, or the exact name of an open interactive
  change. A key or name matching an open `/flow-interactive` change resumes it (**Resume**,
  below); one matching a finished one ends the run with one line saying so.
- **Jira**: with no key given, file a Task through the existing creation path
  (**Resolution**, `skills/flow-contracts/jira-integration.md`), summary `Interactive session
  <YYYY-MM-DD>`, labelled per **Labels on issues the pipeline creates**
  (`skills/flow-contracts/jira-integration-finish.md`), unless the argument declines an issue
  ("no jira"). Move it to In Progress. Every Jira failure is the one `⚠ Jira: skipped — <reason>`
  line of **Never blocking**.
- **Name**: `<lowercased key>-<slug>` per **Change naming**; with no issue,
  `interactive-<YYYYMMDD-HHMM>` (UTC).
- **Session token** `fi-<key or slug stem>-<UTC start stamp>`, the stamp shaped as **Stage marks**
  (`skills/flow-contracts/pipeline.md`) states, generated once per invocation (a resume gets a new
  one).
- **Landing route**: `project-get.sh <project> "default landing route" --enum "pull request"
  "merge and push" manual` must resolve `merge and push`. Absent (exit 1), malformed (exit 3) or
  any other value ends the run with one line naming the section to declare, and so does
  `## handoff` `required` — that project asks to test before anything lands.
- **Worktree**: one per session, `<project>-worktrees/<name>` on branch `<name>`, created as
  `/flow-fast` section 1 creates it, `## worktree setup` run once. `mkdir -p <changeRoot>`, where
  `<changeRoot>` = `<worktree>/.superpowers/sdd/<name>/`: the record verbs find the change's
  checkout by that directory (`changeCheckouts`, `stats/cmd/flow/record.go`), and it holds the
  session's edit scripts.
- **Hook registrations**: grep the harness's hook config for every identifier `setup.sh`'s
  `warn_unregistered_hook` checks for that harness (on Claude Code: `enforce-agent-baseline`,
  `enforce-loop-dispatch`, `flow-active-change`, `flow-session-activity`, `flow-stop-guard`,
  `protect-main-checkout` in `~/.claude/settings.json`). Any missing gets one line naming them and
  `<agents repo>/setup.sh global`, which prints each registration snippet. The session never edits
  that file, asks nothing about it, and continues.
- Read `## lint`, `## test`, `## visual verification` and `## run` once.

### One request

Every operator message that asks for something is a request — a question included ("why don't I
see X") — recorded at once with `flow record request begin -n <n> -title '<one line>'`. A message
that changes the in-flight request's own ask is applied to that request instead. Outside a
brainstorm, a question carrying a recommended option is taken, never asked (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). A pipeline defect met mid-session is
fixed per **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`), as `/flow-fast`
follows it. The cycle, each step its own stage run carrying `-request <n>`:

1. **Look** (`flow.load-context`): the live state behind what the operator saw first — the rows
   the page reads, the page itself, the log — then the code. Never answer from code alone.
2. **Brainstorm**, rarely (`flow.brainstorm`): only when the request needs a new table, endpoint or
   layout, spans more than one module, or two readings lead to materially different work. The
   reason is recorded on the request (`flow record request begin -n <n> -brainstorm-reason
   '<why>'`); questions go to the operator through **AskUserQuestion**, per **Unanswered mid-run
   asks** (`skills/flow-contracts/pipeline.md`).
3. **Implement** (`flow.sdd-tdd`): rebase the worktree onto `origin/<base>` on a clean tree; a
   failing test first for each behaviour change; edits through the Edit tool, or a script written
   to `<changeRoot>` and run by path that asserts each anchor matched exactly once; never a
   symlinked build input (`make -s web-dist` runs `npm ci` itself); commit with the module as
   scope, no trailer; push the branch.
4. **Verify** (`flow.verify`): `run-verify-list.sh <worktree> lint` (every lint and guard
   command), the `## test` commands scoped to the packages and files the request touched,
   `check-fast-route-record.sh <worktree> <base>` — as `/flow-fast` section 5 states them.
5. **Land** (`flow.landing-routes`): fetch and rebase — a conflict per the **Conflict** bullet of
   **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`); a rebase that moved the
   branch re-runs step 4. On the request's first landing, note `git -C <worktree> rev-parse
   origin/<base>` right after this rebase: it is the request's **base**, the parent of its first
   landed commit. `git -C <worktree> push origin <name>:<base>` as its own git-only call. A
   non-fast-forward rejection fetches, rebases, re-runs step 4 and pushes again; any other
   rejection leaves the request open and the session says why (⛔ blocked). Then fetch, and
   `merge-base --is-ancestor HEAD origin/<base>` must pass before anything else runs.
6. **Reinstall** (`flow.refresh-main-checkout`): `git -C <project> pull --ff-only` when the main
   checkout is on the base and clean, then the reinstall the project's instruction file requires
   after a landing, each command its own call with its exit code checked; a failed one is re-run
   once, and a service it restarts must answer before anything reads it. A skipped pull is one
   line; on a request step 7 would check, the request stays open (⛔ blocked), since the live page
   cannot show the landed code.
7. **UI check** (`flow.visual-verify`), when `check-visual-trigger.sh` matches the request's diff:
   open the live page — the link the operator gave, else the URL `## run` serves, by the app's own
   route (`#/run/<project_key>/<name>`, never a display name) — `location.reload()`, screenshot,
   and read the screenshot against the ask. A pass ends the stage run `-outcome completed`. A
   defect ends it `-outcome failed`, and is a fix commit and steps 4–7 again inside the same
   request.
8. **End**: `flow record request end -n <n> -outcome landed -commit <last landed sha> -base <base>`
   (**Request-end guard**, below), or `-outcome answered|dropped`, which take neither sha; then a
   short report carrying neither `**Learning:**` nor `**Prompt tuning:**`.

A resumed request whose base was never noted passes the parent of its earliest commit found on
`origin/<base>`, or any older ancestor: a wider range can only demand a UI check, never skip one.

A **mid-turn message** is recorded as the next request at once and worked, in the same turn,
after the in-flight one ends — never a stop. A turn ends when the queue is empty.

### Request-end guard

`flow record request end -outcome landed` checks, in-process and before the write — the pattern of
`flow stage end`'s `gateVisualVerifyTrailers` (`stats/cmd/flow/stage.go`):

1. The worktree is the checkout carrying `<changeRoot>` (`changeCheckouts`, its first carrying
   linked worktree, else the main checkout when it carries it; none → refused); `<base>` is what
   the `resolve-base-branch` guard prints for it.
2. `-commit` must be an ancestor of `origin/<base>` — the local ref a successful push updates — and
   `-base` an ancestor of `-commit`.
3. `git diff --name-only <base>..<commit>` is fed to the `check-visual-trigger` guard in-process
   against the worktree's `.flow/project.md`. Exit 0 (a `ui paths` match) requires the request's
   latest `flow.visual-verify` stage run — the latest by start of the change's `/flow-interactive`
   `flow.visual-verify` runs carrying the request — to have ended `completed`. Exit 1, or exit 2
   with `VISUAL-TRIGGER-NOT-CONFIGURED`, requires nothing. Exit 2 with
   `VISUAL-TRIGGER-CANNOT-ANSWER` refuses, quoting its cause.
4. With the store unreachable, checks 1–3's git work still runs; the stage-run read is skipped
   with one `⚠` line, and the write is journalled like every record verb.

Exit codes, those of every record verb: 0 written or journalled; 1 refused, one line naming the
failed check; 2 usage. `-commit` and `-base` are required with `landed` and refused with
`answered` and `dropped`; an `-outcome` outside the three is a usage error. `-base` is stored as
`requests.base_sha`.

### Explicit stop

Only a message whose intent is to end the session; anything that could also be a request is kept
as a request and the session stays open, saying so in one line. Stop:

1. The in-flight request lands (unless the stop drops it); queued requests not started end
   `dropped` and are listed in the stop report.
2. **Self-review** (`flow.self-review`, no `-request`): `/flow-fast` section 5's self-review
   paragraphs as written, with `-command '/flow-interactive'` and the `fi-` token. Its bundle's
   git-log source, for a change with requests, renders each request in number order —
   `### Request <n>: <title> — <outcome>`, then for a landed one `git log --stat
   <base_sha>..<commit_sha>` — since the session branch's own log is empty by stop time. Its fix
   commits and its report land on `<name>`.
3. Steps 4–6 of **One request** land those commits with no `-request`.
4. **Cleanup** (`flow.cleanup`): tree clean and branch an ancestor of `origin/<base>`, then
   `git worktree remove` without `--force`, branch deleted locally and on the remote.
5. Jira: the description gains a `## Session summary` section — one line per request: number,
   title, outcome, commit — under the existing append-only assertion; the issue moves to Done.
6. `flow stage mark flow.write-finished` moves the change to `FINISHED`.
7. The stop report carries `**Learning:** NONE — /flow-interactive composes no note` and the
   `**Prompt tuning:**` line the self-review pass leaves.

### Resume

`/flow-interactive <key or name>` reuses the worktree (recreated from the remote branch, else
`origin/<base>`, when gone, then `mkdir -p <changeRoot>` in it, since `.superpowers/` is
gitignored), marks `flow.kickoff` under a new token, continues numbering after the
highest recorded request, and takes an open request from the dead session as the current one — its
stages re-run as new attempts under the same number.

### Store, CLI and API

- Migration `0056_requests.sql`: table `requests` (`change_id`, `n`, `title`, `brainstorm_reason`,
  `received_at`, `ended_at`, `outcome` landed|answered|dropped, `commit_sha`, `base_sha`), unique
  on `(change_id, n)`; column `stage_runs.request_n`. A request's state is derived on read, never
  stored: ended → its outcome; any stage run carrying it → `in progress`; otherwise `queued`.
- `flow record request begin -change <name> -n <n> [-title '<one line>'] [-brainstorm-reason
  '<why>']` and `flow record request end -change <name> -n <n> -outcome
  landed|answered|dropped [-commit <sha> -base <sha>]`: each an upsert by `(change, n)` — an
  omitted title or reason leaves the stored value, and `received_at` is kept from the first write
  — journalled on an unreachable store like every record verb and replayed by reconcile, the CLI
  stamping each write's instant so a replay keeps it; `begin` also bumps `changes.updated_at`, so a
  session older than the board's period chip stays on the board; `end -outcome landed` runs the
  guard above. `-n` is a positive integer; `begin` with neither `-title` nor
  `-brainstorm-reason` is a usage error.
- `flow stage begin -request <n>`: stored on the stage run; refused with any command but
  `/flow-interactive`, by the CLI (exit 2) and by the API (400).
- `stats/internal/stages/names.go`: `FlowInteractive`, added to the rows of `flow.kickoff`,
  `flow.load-context`, `flow.brainstorm`, `flow.sdd-tdd`, `flow.verify`, `flow.visual-verify`,
  `flow.landing-routes`, `flow.refresh-main-checkout`, `flow.self-review`, `flow.cleanup`,
  `flow.write-finished`; README's Level 1 table follows; a pin test ties the skill's own Stage keys
  table to the rows, as `TestStageKeysMatchFlowFastSkillTable` does for `/flow-fast`.
- `stats/internal/api/stages.go`: the `flow.write-finished` → `FINISHED` rule, today `/flow-fast`
  only, also fires for `/flow-interactive`; the synthetic `STARTED` bootstrap already covers any
  command. `flow stage end`'s self-review gate (`gateBrainstormDispatchesClosed`,
  `stats/cmd/flow/stage.go`) covers `/flow-interactive`'s `flow.self-review` as it covers
  `/flow-fast`'s.
- Board API: `LiveStateBoard` reads the change's requests and its latest stage run's command; the
  row DTO gains `command`, and `tasks` is served from requests when the change has any — each
  request a task whose `state` is `done` when ended, `in progress` or `pending`, carrying
  `outcome`, `commit` and `reason` when set; the stage-run DTO gains `request`.
- Self-review bundle (`stats/internal/selfreview/git.go`, `Bundle`): the requests source of
  **Explicit stop** step 2, in place of `changeBranchLog` for a change with requests, read by the
  bundle handler (`stats/internal/api/selfreview.go`) from the store.

### Board

As the prototype draws it: the existing card and run view, made interactive-aware for a change
whose `command` is `/flow-interactive`.

- **Card**: pill `Interactive`; `Request <n>: <title>` (the in-progress request) as the running
  line; `requests <ended>/<total>` in the meta line and the progress bar at ended/total. Between
  requests — no request in progress and no stage open — a ring glyph and `Waiting for the next
  request · <span since the last stage ended>` in place of `Paused after …`, a muted `Last:
  Request <n> <outcome> — <title>` line for the latest ended request, no activity badge, no
  agent-count line and no needs-you border. The card's hover and focus border is the prototype's.
- **Run view**: the pill and `Now running` as on the card; a `Requests` panel in the Tasks slot,
  its note `<ended>/<total>` — landed `✓ <n>. <title>` with the short commit, answered `✓` with
  `answered`, dropped `–` struck through with `dropped`, the current request `●` in bold
  (`aria-current="step"`), queued `○` with `queued`, and a request's brainstorm reason on its own
  line under it as `brainstormed — <reason>`; the Main panel's note `Request <n> · <span since the
  request's first stage run began>` and its steps cut to the per-request stages of the current
  request — `Loading context`, `Implementing tasks`, `Running lint and tests`, `Landing the
  change`, `Updating the main checkout`, `Checking the UI`, each done, current or todo; the
  timeline's upcoming lanes cut to the same stages; the Stage runs table with a first `Request`
  column reading `R<n>`, its rows newest first with no session group rows.

### Hooks

- `hooks/enforce-agent-baseline.py`: `FLOW_COMMAND` gains `-interactive`, and `session_change`
  reads the session's latest stage run of any command — closing the `/flow-fast` fail-open
  **Context** names.
- `hooks/flow-active-change.py`: one GET of the session's latest stage run of any command. `/flow`
  keeps today's line; `/flow-interactive`, unless that run is `flow.write-finished`, injects
  "flow: this session runs /flow-interactive change `<name>`. A plain message is its next request
  — or its stop (**One request**, **Explicit stop**, `skills/flow-interactive/SKILL.md`)."; any
  other command injects nothing.
- `hooks/flow-stop-guard.py` stays `/flow`-only: a turn ending between requests is the session's
  normal rest. Its handoff check reads any session's last turn text, so per-request reports carry
  neither handoff line and the stop report carries both.

### Registration

`skills/flow-interactive/SKILL.md` with a `## Stage keys` table and the sections above;
`skills/flow-interactive/scripts/` symlinks for every guard it invokes — the set
`check-guard-symlinks.sh` requires, `lib` included, as `skills/flow-fast/scripts/` carries them;
`commands-claude/flow-interactive.md`; index rows in `CLAUDE.md`, `AGENTS.md`,
`templates/CLAUDE.md`, `templates/AGENTS.md`, `README.md`, `skills/README.md`; `CLAUDE.md` and
`AGENTS.md`'s **Reinstall after a change to this repository is fully done** gains "on
`/flow-interactive`, after each request lands"; `skills/flow-contracts/jira-integration.md` gains
the default-create Resolution line, the start/stop Transitions rows and the session-summary
Description line; `skills/flow-contracts/pipeline.md`'s **Command surface** names the command.

### The brief's failure list

| Failure | Covered by |
|---------|-----------|
| Worktree force-removed before its commit landed | One session worktree, removed only at stop, never `--force`, only when clean and an ancestor of `origin/<base>`; step 5's ancestor check after every push; the request-end guard's check 2 |
| `make web-dist` vs a symlinked `node_modules` | Step 3: never symlink a build input; `make -s web-dist` runs `npm ci` |
| Struct-field insert missed silently | Step 3: Edit tool, or a script asserting each anchor matched exactly once |
| `python3 -` heredoc failed to parse | Step 3: scripts are files under `<changeRoot>`, run by path |
| Old bundle after a restart | Step 7: `location.reload()` before every screenshot |
| Wrong URL (display name, not project key) | Step 7: the app's own route, keyed by project key |
| UI shipped without looking | Step 7 on every trigger match; the request-end guard's check 3 |
| `make -C stats restart` exited 2 | Step 6: exit code checked, re-run once, the service must answer |
| Session-activity hook never registered | Session start: registration check, one line naming the fix |
| Answering by assumption | Step 1: live state before code |

## Decisions

### One change per session, requests as their own record

**ID:** session-change-request-record
**Status:** active
**Chosen:** one store change per session; a `requests` table plus `stage_runs.request_n` — the
ask, its brainstorm reason, receipt, end, outcome, base and commit exist nowhere else, and the
operator asked for enough recorded data to analyse the lifecycle.
**Considered:** one change per request — N cards and N issues, contradicting the per-run issue,
resume-by-id and "current request" the ticket asks for; each request as an inline
`task-<n>-implementer` dispatch with the ask in `dispatches.notes` (zero migrations) — the ask
and reason in a generic text column, an inline window that must close with a commit (a question
has none) and starts at the previous request's end, and no queued state: a proxy record;
requests as `tasks.md` lines — lost with the worktree and invisible to the store.

### Reuse the `flow.*` stage keys

**ID:** reuse-flow-stage-keys
**Status:** active
**Chosen:** add `/flow-interactive` to eleven existing rows — labels, phases, cost by stage and
`autoStageFor` already read them, and stage runs carry their command.
**Considered:** new `interactive.*` keys — new `STAGE_PHASE`, label and README rows for no gain in
what can be analysed.

### One worktree per session

**ID:** session-worktree
**Status:** active
**Chosen:** one worktree for the whole session, rebased before every request — `## worktree setup`
(`npm ci`, Go cache) runs once and nothing is removed per request.
**Considered:** a worktree per request, as the brief's session did — a setup per request and the
removal step that lost a request's edits.

### Merge and push, declared

**ID:** merge-and-push-only
**Status:** active
**Chosen:** every request lands by merge and push, and the project must declare that route; an
absent, malformed or other route, or `## handoff` `required`, ends the run at start naming the
section.
**Considered:** honouring `pull request` — no project declares it, and per-request PRs from one
session branch would stack on each other; treating an absent route as merge and push —
`/flow-fast` reads absent as `pull request`, so this would push to the base of a project that never
chose it; honouring `handoff required` by stopping before each landing — the session would ship
nothing without a turn per request, the opposite of what the command is for.

### Main session implements inline

**ID:** inline-implementation
**Status:** active
**Chosen:** the main session does each request itself — what made the 2026-10-29 session fast.
**Considered:** a fresh implementer subagent per request — main's context stays small, but every
one-line fix pays a dispatch and a hand-back; inline with a subagent only for a brainstormed
request — two paths to maintain.

### The board, made interactive-aware

**ID:** board-interactive-aware
**Status:** active
**Chosen:** the existing card and run view, as the prototype draws them.
**Considered:** a dedicated interactive card and run layout — more SPA code and new baselines for
the same information; no SPA change — requests read as `Task n`, the `/flow` phase strip shows
Decisions, Review and Integrate as todo, and the card reads `Paused after …` between requests.

### Self-review once, at the stop

**ID:** self-review-at-stop
**Status:** active
**Chosen:** one self-review pass over the whole session at the explicit stop, through `/flow-fast`'s
in-run self-review, its fixes landed before the session finishes.
**Considered:** never — the cheapest, but the session's process lessons are never mined; after
every request — one opus dispatch per request slows the loop the command exists for.

### The request-end guard

**ID:** request-end-guard
**Status:** active
**Chosen:** `flow record request end -outcome landed` refuses unless the commit is on
`origin/<base>` and, when the request's diff matches `ui paths`, its latest `flow.visual-verify`
run ended `completed`; in-process in the CLI, reusing the `resolve-base-branch` and
`check-visual-trigger` guards. A defect-finding UI check ends `failed`, so only a passing re-check
after the fix satisfies it. With the store unreachable the git checks still run and only the
stage-run read is skipped.
**Considered:** steps only — nothing stops a skipped check, and those are the brief's two costliest
failures; a separate `flow-guard` script the skill calls before `end` — a step the session can
skip, which is steps only again; any completed `flow.visual-verify` run carrying the request — the
run that found the defect would satisfy it; letting the call through when the store is unreachable
(round 1) — the git checks need no store, so skipping them gains nothing.

### The request's base is fixed at its first landing

**ID:** request-base-first-landing
**Status:** active
**Chosen:** `-base` is `origin/<base>` right after the request's first landing rebase — the parent
of its first landed commit — so `<base>..<commit>` holds exactly that request's commits on a
single landing, and on a fix-and-reland adds only other changes' commits, which can demand a UI
check the request already ran, never skip one.
**Considered:** the base at implement start — a concurrent landing before the rebase widens the
range with another change's files and demands a UI check for a request that touched none; the
previous request's `commit_sha`, derived with no flag — the same widening whenever a concurrent run
lands between requests; a `Flow-Request:` trailer on every commit — a line in the base's whole
history for one guard's range.

### No learning note

**ID:** no-learning-note
**Status:** active
**Chosen:** `/flow-interactive` composes no learning note; the stop report's `**Learning:**` line
reads `NONE — /flow-interactive composes no note`, and `skills/flow-contracts/learning.md`, which
lists `/flow`'s implementation run and `/flow-fast` alone, needs no edit.
**Considered:** one note per session at kickoff — it delays the first request, which the command
exists to answer at once.

### The prompt hook names the interactive session

**ID:** prompt-names-interactive-session
**Status:** active
**Chosen:** `flow-active-change.py` reads the session's latest stage run of any command and, for an
unfinished `/flow-interactive` change, reminds main that a plain message is its next request or
its stop.
**Considered:** the hook stays `/flow`-only — after a compaction in a long session nothing points a
plain message back at the request protocol, and an unrecorded request is lost from the lifecycle
record the ticket exists for; two GETs, `/flow` first — a second store call on every prompt of
every session, where one GET branching on the command answers both.

### Hook registration is a notice

**ID:** hook-registration-notice
**Status:** active
**Chosen:** at kickoff, one line naming every unregistered hook and `setup.sh global`, which prints
the snippets; the session continues.
**Considered:** asking through **AskUserQuestion**, as the brief put it — no answer changes what the
session does next; editing `~/.claude/settings.json` — the operator's own file, which the brief's
session was blocked from editing.

### Visual check on the live page after landing

**ID:** live-ui-check-after-landing
**Status:** active
**Chosen:** the main session reloads and screenshots the live page the operator is watching, after
the reinstall.
**Considered:** `/flow`'s `flow.visual-verify` verifier — an opus dispatch and capture specs per
UI request, against seeded data, minutes and dollars each; a check before landing on the UI-test
stack or a store copy — seeded or copied data, not the rows the operator is looking at, and a
store copy per UI request.

### Resume by key or by name

**ID:** resume-by-key-or-name
**Status:** active
**Chosen:** resume accepts a Jira key or the exact change name, so a session without an issue can
still be resumed and stopped.
**Considered:** key only, as the ticket's comment states ("a run with no id assigned is not
resumable") — a keyless session left without a stop could then never be ended and would sit on
Active runs forever; the change name is the id it has.

### Brainstorm in-session

**ID:** brainstorm-in-session
**Status:** active
**Chosen:** a rare brainstorm runs in the session under `flow.brainstorm`, its reason on the
request.
**Considered:** handing a big request to `/flow` — the operator's comment keeps it in the session.

### Fix the begun-row fail-open for `/flow-fast` here

**ID:** session-change-any-command
**Status:** active
**Chosen:** widen `session_change` to any command in this change — the interactive session's
own dispatches need it, and it closes the `/flow-fast` gap with one test.
**Considered:** a separate change — the same edit, made twice.

### No flags

**ID:** no-base-flag
**Status:** active
**Chosen:** `/flow-interactive` takes no `--base`; the base is the branch `origin/HEAD` names.
**Considered:** `--base <branch>`, as `/flow-fast` takes it — `origin/HEAD` already names the
landing branch in every flow project.

### Requests reach the board as its tasks

**ID:** requests-served-as-tasks
**Status:** active
**Chosen:** the board row's `tasks` carries the requests when the change has any, each task
gaining optional `outcome`, `commit` and `reason` fields, so the card's count, its progress bar
and the Tasks slot read one list.
**Considered:** a separate `requests` field on the row — every reader of `tasks` (count, progress,
running line, panel) would branch on which list to read.

### The request verb lives in its own file

**ID:** request-verb-own-file
**Status:** active
**Chosen:** `flow record request` and its guard in `stats/cmd/flow/request.go`, dispatched from
`runRecord`.
**Considered:** inside `stats/cmd/flow/record.go` — already the CLI's largest file, and the guard
shares nothing with the other record verbs but their write helpers, which it calls.

### Interactive display states the prototype does not draw

**ID:** interactive-display-states
**Status:** active
**Chosen:** a `FINISHED` interactive change keeps the `Finished` pill every finished change shows;
the Main steps add `Designing the change` after `Loading context` only when the current request
ran `flow.brainstorm`; a request recorded but not started while nothing is in progress reads as
waiting; a stage run carrying no request shows an empty `Request` cell.
**Considered:** an `Interactive` pill on a finished session — it would read as still running; a
`Designing the change` pill on every request — the prototype draws none, and most requests never
brainstorm.

### The UI-test fixture carries the interactive sessions

**ID:** seed-interactive-in-fixture
**Status:** active
**Chosen:** `stats/cmd/uitest-seed/seed.go` seeds the prototype's running and waiting sessions
inside the pinned board period, and every capture the seed moves is recaptured and read.
**Considered:** seeding them outside the pinned period so no existing capture moves — the open
experiment's after window and the weekly and trend views still count stage runs outside it, so
captures move either way, and the board captures would then never show an interactive card.

### The self-review gate covers the interactive stop

**ID:** self-review-gate-interactive
**Status:** active
**Chosen:** `flow stage end` refuses a `/flow-interactive` `flow.self-review` end while a dispatch
row of the change is open, as it does for `/flow` and `/flow-fast`.
**Considered:** leaving the gate to `/flow` and `/flow-fast` — the pass's own row, opened by
`enforce-loop-dispatch.py`, would stay running on the board after the stop.

## Open questions

## Recommended approach

Store first — the `requests` table, `stage_runs.request_n` and their store methods — then the
stage vocabulary and the `FINISHED` rule, the API, client, journal replay and `flow stage begin
-request`, then `flow record request` with its guard, the board API and the self-review bundle's
requests source. The SPA follows: the card, then the run view, then the fixture and the capture
specs against the prototype. The two hook changes are independent. The skill and its command file
come once the CLI they call exists, and the index and contract registration last. Rejected
alternatives: a change per request or a dispatch-row proxy (session-change-request-record), new
stage keys (reuse-flow-stage-keys), a worktree per request (session-worktree), honouring other
landing routes (merge-and-push-only), an implementer subagent (inline-implementation), a
dedicated board layout (board-interactive-aware), a self-review per request (self-review-at-stop),
steps without a guard (request-end-guard), a flag-derived base (request-base-first-landing), a
separate `requests` row field (requests-served-as-tasks), a seed outside the pinned period
(seed-interactive-in-fixture).

## Critical files

- `stats/internal/store/migrations/0056_requests.sql` — the `requests` table and
  `stage_runs.request_n`.
- `stats/internal/store/requests.go` — begin/end upserts, the list with derived state.
- `stats/internal/store/stageruns.go` — `BeginStageInput.RequestN`, `StageRun.RequestN`, the insert
  and `QueryStageRuns`.
- `stats/internal/records/types.go` — the request wire types.
- `stats/internal/store/aggregate.go` — `LiveStateBoard`'s command and requests.
- `stats/internal/stages/names.go`, `README.md` — the stage vocabulary and its documented table.
- `stats/internal/api/stages.go` — `-request` on a begin mark, the `FINISHED` rule.
- `stats/internal/api/requests.go`, `records.go`, `server.go` — the request routes and
  `RecordStore` methods.
- `stats/internal/api/stats.go` — board row `command`, requests as `tasks`, stage-run `request`.
- `stats/internal/api/selfreview.go`, `stats/internal/selfreview/git.go` — the bundle's requests
  source.
- `stats/internal/client/client.go` — request writes, `BeginStageRequest.Request`,
  `StageRunRow.Request`.
- `stats/internal/reconcile/reconcile.go` — replay of the request journal kinds and of a begin
  mark's request.
- `stats/cmd/flow/request.go`, `record.go`, `stage.go` — the request verb and its guard, the
  `-request` flag, the self-review gate.
- `stats/web/src/api.ts`, `lib/interactive.ts`, `views/StateBoard.tsx`, `views/RunDetail.tsx`,
  `components/StageRunTable.tsx`, `styles.css` — the board.
- `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/interactive.spec.ts` and its
  `.mockups` sidecar — fixture and captures.
- `hooks/enforce-agent-baseline.py`, `hooks/flow-active-change.py` — the two hook changes.
- `skills/flow-interactive/SKILL.md`, `skills/flow-interactive/scripts/`,
  `commands-claude/flow-interactive.md` — the command.
- `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`, `templates/AGENTS.md`, `skills/README.md`,
  `skills/flow-contracts/jira-integration.md`, `skills/flow-contracts/pipeline.md` — registration.

## Existing code to reuse

- `change_task_counts` (`stats/internal/store/taskcounts.go`) and its record route — the
  change-scoped record shape; `RecordIncident` → `ApplyIncidentRecord` → reconcile's `incident`
  case — the write, journal and replay path every record verb shares (`callRecord`,
  `classifyRecordWrite`, `journalRecordWrite`, `stats/cmd/flow/record.go`).
- `gateVisualVerifyTrailers` (`stats/cmd/flow/stage.go`) — in-process `guard.Registry` calls with a
  `guard.Env`; `changeCheckouts` (`stats/cmd/flow/record.go`); `resolve-base-branch` and
  `check-visual-trigger` (`stats/internal/guard/`); `client.ListStageRuns`.
- `LiveStateBoard`'s `jsonb_agg` lateral for subagents (`stats/internal/store/aggregate.go`).
- `changeBranchLog`, `dropPNGStatLines`, `trimmedOutput` (`stats/internal/selfreview/git.go`).
- `TestStageKeysMatchFlowFastSkillTable` and `extractFlowFastSkillKeys`
  (`stats/internal/stages/names_test.go`).
- `runningTask`, `taskCount`, `stageLabel`, `effectiveActivity` (`stats/web/src/lib/`),
  `TasksPanel`, `WorkPanel` (`stats/web/src/views/RunDetail.tsx`), `StageRunTable`.
- `seedFixture`, `fixtureNow` (`stats/cmd/uitest-seed/seed.go`); `experiments.spec.ts` and its
  `.mockups` sidecar (`stats/web/tests/visual/`).
- `runHook`, `openRowFixture`, `newStageStore` (`stats/internal/repotest/`).
- `skills/flow-fast/SKILL.md` — its worktree creation, verify, self-review and landing sections,
  cited rather than copied.

## Verification

- `cd stats && gofmt -l . && make -s web-dist && go vet ./...` clean; the touched Go packages'
  tests green; `cd stats/web && npx tsc -b` clean and the touched vitest files green; the
  interactive capture spec matches the prototype; `run-verify-list.sh <worktree> lint` clean.
- The live check below.

## Live check

The change touches `flowd` and its store. `flow.verify` runs a scratch session against
`scripts/live-store-copy.sh up kan-1061`, with the `flow` CLI built from this worktree
(`cd stats && go build -o <scratch>/flow ./cmd/flow`) and the copy's `FLOW_ADDR` and
`FLOW_RECORDS_ADDR` exports, in a scratch repository with a bare origin whose `.flow/project.md`
declares `merge and push` and a `ui paths` glob:

1. Kickoff: `flow stage mark -command '/flow-interactive' -stages flow.kickoff …` for a scratch
   change.
2. Request 1: `request begin`, a `flow.load-context` run carrying `-request 1`, `request end
   -outcome answered`.
3. Request 2: a commit touching no `ui paths` file, pushed to the base; `request end -outcome
   landed -commit <sha> -base <base>` exits 0.
4. Request 3: a commit touching a `ui paths` file, pushed; `request end -outcome landed` exits 1
   with no `flow.visual-verify` run, exits 1 after a run carrying `-request 3` ends `failed`, and
   exits 0 once a later one ends `completed`.
5. Stop: `flow stage mark -command '/flow-interactive' -stages flow.write-finished …`.

Record before/after: the change on Active runs with `command` `/flow-interactive`, pill
`Interactive` on the copy's board page, `tasks` 3/3 from requests, stage runs carrying `request`,
`requests.base_sha` set on both landed rows, `FINISHED` after the stop mark. Not working looks
like: marks refused as an unknown stage for the command, the change absent from the board,
`tasks` null, request 3's end accepted without a passing UI check, or the change still `STARTED`
after `flow.write-finished`. `scripts/live-store-copy.sh down kan-1061` ends it.
