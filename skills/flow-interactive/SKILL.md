---
name: flow-interactive
description: Conversational /flow session — one invocation is one change, one Jira issue and one worktree, and a stream of requests worked inline. Each request is looked at live, implemented test-first, verified with the project's lint and targeted tests, landed by merge and push, reinstalled and, when it touches ui paths, checked on the live page; an explicit stop runs one self-review pass, cleanup and FINISHED. Every request is recorded in flowd; no spectre artifacts or state file. Use for /flow-interactive.
allowed-tools: Bash(flow:*), Bash(land-self-review-report.sh:*)
license: MIT
---

One `/flow-interactive` invocation is one **session**: one store change, one Jira issue, one git
worktree, and a stream of **requests**, each shipped to the base branch on its own. The session
marks existing `flow.*` stage keys, so labels, phases, cost by stage and the timeline read as they
do for `/flow` and `/flow-fast`. Each request is a row the store keeps (`flow record request`),
and every stage run made for it carries its request number (`-request <n>`). This session does the
work inline: no review panel per request, no spectre artifacts, no plan file, no decision roll, no
state file, and no workspace isolation beyond the worktree.

**Announce at start:** "Using flow-interactive for change `<name>`."

**No flags.** The only argument is empty, a first request, a Jira key, or the exact name of an open
interactive change; report anything else rather than ignoring it. The base is the branch
`origin/HEAD` names; from the worktree's creation on, `<base>` is what
`resolve-base-branch.sh <worktree>` prints.

**Every mark is a literal call.** `<harness>` is this harness's name, `fi-<literal-token>` is one
token generated once per invocation, shaped as **Stage marks** (`skills/flow-contracts/pipeline.md`)
states, and typed the same at every `stage begin` and `stage mark` — never a shell substitution,
per the CLI's own usage text. `<n>` is the request's number in the session.

**Guardrails, the whole list.** Never dispatch a subagent, save the self-review pass's own
dispatches (`skills/flow-self-review/SKILL.md`) and the in-run fix loop of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`). Never write `<project>/spectre/`, save those
passes' own `self-review-<name>` and `pipeline-fix-<name>-<k>` changes, or a state file. Never push
to a branch other than `<name>` and `<name>:<base>`. Never set up workspace isolation.

## Stage keys

Every stage `/flow-interactive` marks uses a `flow.*` key — the same keys `/flow` marks. The full
key list, in the order each section below marks them:

| Step | Keys |
|------|------|
| **Session start** | `flow.kickoff` |
| **One request** | `flow.load-context`, `flow.brainstorm`, `flow.sdd-tdd`, `flow.verify`, `flow.landing-routes`, `flow.refresh-main-checkout`, `flow.visual-verify` |
| **Explicit stop** | `flow.self-review`, `flow.cleanup`, `flow.write-finished` |

Every **One request** stage is a `flow stage begin … -request <n>` / `flow stage end` pair; the
stop's stages carry no `-request`.

## Session start

**Argument.** A key or name matching an open `/flow-interactive` change resumes it (**Resume**,
below); one matching a finished change ends the run with one line saying so. Anything else starts
a new session, a non-empty argument being its first request.

**Jira.** Resolve the issue per **Resolution (how `jiraIssue` is decided)**
(`skills/flow-contracts/jira-integration.md`): with no key given, file a Task through its creation
path, summary `Interactive session <YYYY-MM-DD>`, labelled per **Labels on issues the pipeline
creates** (`skills/flow-contracts/jira-integration-finish.md`), unless the argument declines an
issue ("no jira"). Move it to **In Progress** per **Transitions** there. Every Jira failure is the
one `⚠ Jira: skipped — <reason>` line of **Never blocking**.

**Name.** `<lowercased key>-<slug>` per **Change naming**; with no issue,
`interactive-<YYYYMMDD-HHMM>` (UTC).

```bash
flow stage mark  -command '/flow-interactive' -stages flow.kickoff -harness <harness> -session-token fi-<literal-token> <name>
```

**Landing route.** `project-get.sh <project> "default landing route" --enum "pull request" "merge
and push" manual` must resolve `merge and push`. Absent (exit 1), malformed (exit 3) or any other
value ends the run with one line naming `## default landing route` as the section to declare, and
so does `project-get.sh <project> handoff --enum required none` resolving `required` — that
project asks to test before anything lands.

**Worktree.** One per session, `<project>-worktrees/<name>` on branch `<name>`, created as section
1 of `/flow-fast` creates it (`skills/flow-fast/SKILL.md`), its `## worktree setup` run once — with
no `--base`. Then `mkdir -p <changeRoot>`, where `<changeRoot>` is
`<project>-worktrees/<name>/.superpowers/sdd/<name>/`: the record verbs find the change's checkout by that
directory, and it holds the session's edit scripts. The main checkout is never checked out,
staged, committed or written.

**Hook registrations.** Grep the harness's hook config for every identifier `<agents repo>/setup.sh`'s
`warn_unregistered_hook` checks for that harness (on Claude Code: `enforce-agent-baseline`,
`enforce-loop-dispatch`, `flow-active-change`, `flow-session-activity`, `flow-stop-guard`, `await-own-agents`,
`protect-main-checkout` in `~/.claude/settings.json`). Any missing gets one line naming them and
`<agents repo>/setup.sh global`, which prints each registration snippet. Never edit that file, ask
nothing about it, and continue.

Read `<project>/.flow/project.md`'s `## lint`, `## test`, `## visual verification` and `## run`
once, with `project-get.sh <project> <key>`. The project's instruction file is already in this
session's context; never read it here.

## One request

Every operator message that asks for something is a request — a question included ("why don't I
see X") — recorded at once, `<n>` one past the highest recorded:

```bash
flow record request begin -change <name> -n <n> -title '<one line>'
```

A message that changes the in-flight request's own ask is applied to that request instead.
Outside a brainstorm, a question carrying a recommended option is taken, never asked
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). A pipeline
defect met mid-session is fixed per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`), as `/flow-fast` follows it.

**1. Look.** The live state behind what the operator saw first — the rows the page reads, the page
itself, the log — then the code. Never answer from code alone.

```bash
flow stage begin -command '/flow-interactive' -stage flow.load-context -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow stage end   -command '/flow-interactive' -stage flow.load-context -outcome completed <name>
```

**2. Brainstorm, rarely** — only when the request needs a new table, endpoint or layout, spans
more than one module, or two readings lead to materially different work. Record the reason, then
ask the operator through **AskUserQuestion**, per **Unanswered mid-run asks**
(`skills/flow-contracts/pipeline.md`):

```bash
flow stage begin -command '/flow-interactive' -stage flow.brainstorm -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow record request begin -change <name> -n <n> -brainstorm-reason '<why>'
flow stage end   -command '/flow-interactive' -stage flow.brainstorm -outcome completed <name>
```

**3. Implement.** A run-loaded prose edit the verbatim-moves guard refuses is acknowledged in
`<changeRoot>/verbatim-moves.txt`. Rebase the worktree onto `origin/<base>` on a clean tree; a failing test first
for each behaviour change; edits through the Edit tool, or a script written to `<changeRoot>` and
run by path that asserts each anchor matched exactly once; never a symlinked build input (`make -s
web-dist` runs `npm ci` itself); commit with the module as scope, no trailer; push the branch —
otherwise as section 4 of `/flow-fast` commits (`skills/flow-fast/SKILL.md`). No learning note is
composed.

```bash
flow stage begin -command '/flow-interactive' -stage flow.sdd-tdd -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow stage end   -command '/flow-interactive' -stage flow.sdd-tdd -outcome completed <name>
```

**4. Verify.** `<agents repo>/scripts/run-verify-list.sh <worktree> lint`, the `## test` commands
scoped to the packages and files the request touched, and `check-fast-route-record.sh <worktree>
<base>` — as section 5 of `/flow-fast` states them. The request never lands red.

```bash
flow stage begin -command '/flow-interactive' -stage flow.verify -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow stage end   -command '/flow-interactive' -stage flow.verify -outcome completed <name>
```

**5. Land.** Fetch and rebase — a conflict per the **Conflict** bullet of **Sync the branch onto
the base** (`skills/flow/sync-onto-base.md`); a rebase that moved the branch re-runs step 4. On
the request's first landing, note `git -C <worktree> rev-parse origin/<base>` right after this
rebase: it is the request's **base**, the parent of its first landed commit. Push with `git -C
<worktree> push origin <name>:<base>` as its own git-only call. A non-fast-forward rejection
fetches, rebases, re-runs step 4 and pushes again; any other rejection leaves the request open and
the session says why (⛔ blocked). Then fetch, and `git -C <worktree> merge-base --is-ancestor HEAD
origin/<base>` must pass before anything else runs.

```bash
flow stage begin -command '/flow-interactive' -stage flow.landing-routes -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow stage end   -command '/flow-interactive' -stage flow.landing-routes -outcome completed <name>
```

**6. Reinstall.** `git -C <project> pull --ff-only` when the main checkout is on `<base>` and
clean, then the reinstall the project's instruction file requires after a landing, each command
its own call with its exit code checked; a failed one is re-run once, and a service it restarts
must answer before anything reads it. A skipped pull is one line; on a request step 7 would
check, the request stays open (⛔ blocked), since the live page cannot show the landed code.

```bash
flow stage begin -command '/flow-interactive' -stage flow.refresh-main-checkout -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow stage end   -command '/flow-interactive' -stage flow.refresh-main-checkout -outcome completed <name>
```

**7. UI check**, when `check-visual-trigger.sh <worktree>` matches the request's diff (`git diff
--name-only <base>..HEAD` on its stdin): open the live page — the link the operator gave, else the
URL `## run` serves, by the app's own run-view route, keyed by the project key and `<name>`, never a display name —
`location.reload()`, screenshot, and read the screenshot against the ask. A pass ends the stage
run `-outcome completed`. A defect ends it `-outcome failed` with `-reason` naming the defect, and is a fix commit and steps 4–7
again inside the same request. The stage's end is its own call, made after the screenshot is read —
never chained to its begin.

```bash
flow stage begin -command '/flow-interactive' -stage flow.visual-verify -request <n> -harness <harness> -session-token fi-<literal-token> <name>
flow stage end   -command '/flow-interactive' -stage flow.visual-verify -outcome completed|failed [-reason '<the defect, one line>'] <name>
```

**8. End.** One of:

```bash
flow record request end -change <name> -n <n> -outcome landed -commit <last landed sha> -base <base sha>
flow record request end -change <name> -n <n> -outcome answered
flow record request end -change <name> -n <n> -outcome dropped
```

`landed` runs the **Request-end guard**, below. Then a short report carrying neither
`**Learning:**` nor `**Prompt tuning:**`.

A resumed request whose base was never noted passes the parent of its earliest commit found on
`origin/<base>`, or any older ancestor: a wider range can only demand a UI check, never skip one.

A **mid-turn message** is recorded as the next request at once and worked, in the same turn,
after the in-flight one ends — never a stop. A turn ends when the queue is empty.

## Request-end guard

`flow record request end -outcome landed` checks, in-process and before the write:

1. The worktree is the checkout carrying `<changeRoot>` — its first carrying linked worktree, else
   the main checkout when it carries it; none is refused. `<base>` is what the
   `resolve-base-branch` guard prints for it.
2. `-commit` must be an ancestor of `origin/<base>` — the local ref a successful push updates — and
   `-base` an ancestor of `-commit`.
3. `git diff --name-only <base>..<commit>` is fed to the `check-visual-trigger` guard against the
   worktree's checkout of `<project>/.flow/project.md`. A `ui paths` match (exit 0) requires the request's latest
   `/flow-interactive` `flow.visual-verify` stage run, by start, to have ended `completed`. Exit 1,
   or exit 2 with `VISUAL-TRIGGER-NOT-CONFIGURED`, requires nothing. Exit 2 with
   `VISUAL-TRIGGER-CANNOT-ANSWER` refuses, quoting its cause.
4. With the store unreachable, checks 1–3's git work still runs; the stage-run read is skipped
   with one `⚠` line, and the write is journalled like every record verb.

Exit codes, those of every record verb: 0 written or journalled; 1 refused, one line naming the
failed check; 2 usage. `-commit` and `-base` are required with `landed` and refused with
`answered` and `dropped`; an `-outcome` outside the three is a usage error. A refusal is fixed —
land the commit, run the UI check and end it `completed` — and the end re-run, never worked around.

## Explicit stop

Only a message whose intent is to end the session; anything that could also be a request is kept
as a request and the session stays open, saying so in one line. Stop:

1. The in-flight request lands (unless the stop drops it); queued requests not started end
   `dropped` and are listed in the stop report.
2. **Self-review**: the self-review paragraphs of section 5 of `/flow-fast`
   (`skills/flow-fast/SKILL.md`) as written, with `-command '/flow-interactive'` and the `fi-`
   token. `flow self-review bundle -change <name>` renders the session's requests in number order
   as its git-log source. Its fix commits and its report land on `<name>`, the report through the
   landing chain:

   ```bash
   land-self-review-report.sh "<worktree>" "<name>" \
     "docs(self-review): <name> self-review report" \
     docs/self-review/<name>-self-review.md \
     --push "<name>"
   ```

   ```bash
   flow stage begin -command '/flow-interactive' -stage flow.self-review -harness <harness> -session-token fi-<literal-token> <name>
   flow stage end   -command '/flow-interactive' -stage flow.self-review -outcome completed <name>
   ```

3. Steps 4–6 of **One request** land those commits, their stage runs with no `-request`.
4. **Cleanup**: the tree clean and `git -C <worktree> merge-base --is-ancestor <name>
   origin/<base>` passing, then `git -C <project> worktree remove <project>-worktrees/<name>`
   without `--force`, `git -C <project> branch -D <name>` and `git -C <project> push origin
   --delete <name>`.

   ```bash
   flow stage begin -command '/flow-interactive' -stage flow.cleanup -harness <harness> -session-token fi-<literal-token> <name>
   flow stage end   -command '/flow-interactive' -stage flow.cleanup -outcome completed <name>
   ```

5. Jira: the description gains a `## Session summary` section — one line per request: number,
   title, outcome, commit — per **Description sync** (`skills/flow-contracts/jira-integration.md`);
   the issue moves to **Done**.
6. Move the change to `FINISHED`:

   ```bash
   flow stage mark  -command '/flow-interactive' -stages flow.write-finished -harness <harness> -session-token fi-<literal-token> <name>
   ```

7. The stop report carries `**Learning:** NONE — /flow-interactive composes no note` and the
   `**Prompt tuning:**` line the self-review pass leaves.

## Resume

`/flow-interactive <key or name>` on an open session reuses its worktree — recreated from the
remote branch, else `origin/<base>`, when gone, with `mkdir -p <changeRoot>` run in it since
`<project>/.superpowers/` is gitignored — marks `flow.kickoff` under a new token, continues
numbering after the highest recorded request,
and takes an open request from the dead session as the current one: its stages re-run as new
attempts under the same number.
