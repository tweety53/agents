# myflow pipeline

The twelve-stage pipeline itself: stage definitions, the command→stage transition table, gate
semantics, stage boundaries, and the opt-out flags.

**Load this file when running any `/myflow-*` command or `openspec-*-superpowers` skill.** It is
split out of `rules/myflow-manual-review.mdc` so the always-on rule layer carries only the trigger,
not the whole state machine — the same reason the four contract files beside it were split out.

This file is **canonical** for everything in it. Where a skill or command disagrees with it, this
file wins. The headings below keep the names they had in the rule, so an existing reference to a
section by name still resolves — now to this file.

## Model policy

`/myflow-start` should run on **Opus** (or the harness's strongest available model) — brainstorming and design benefit most from stronger reasoning. Every other `/myflow-*` command (`do`, `do-fix`, `manual-test`, `review`, `finish`, `full`) should run on **Sonnet** (or the harness's standard default).

- **Claude Code**: enforced via `model: opus` / `model: sonnet` in each command's frontmatter (`commands-claude/*.md`) — no manual action needed.
- **Cursor**: not enforceable yet (no per-command model frontmatter support as of this writing) — each `.cursor/commands/myflow-*.md` file carries an explicit note; switch models manually in the composer/chat picker.
- **Codex**: no per-command/skill model override mechanism either — model is a session/profile-level setting; switch manually before starting a new proposal.
- **`/myflow-full`**: runs as a single command/session, so its frontmatter model (Sonnet) applies even during Phase A brainstorming — the per-phase Opus switch doesn't happen automatically anywhere. For brainstorming-heavy new work, run `/myflow-start` standalone (Opus) first, then `/myflow-full <name> skip-propose` for the rest.

## Change name resolution (all `/myflow-*` commands)

`<name>` is **optional** on every `/myflow-*` command. When omitted:

- Run `openspec list --json` and filter to changes relevant to that command's stage (not yet archived).
- Exactly one match → use it automatically; announce which change was picked.
- Multiple matches → **AskUserQuestion** listing each (name, status, last modified) — never guess.
- Zero matches → fall back to that command's normal "no change" handling (e.g. `/myflow-start` asks what to build; others suggest the prior stage's command).

## Pipeline stages

A change is always in exactly one of twelve stages, recorded in its state file (see **State file**):

| Stage | Set by | Waiting on |
|-------|--------|------------|
| `awaiting-proposal-review` | `/myflow-start`, `/myflow-start-fix` | human — read the proposal artifact |
| `proposal-done` | `/myflow-start-done` | `/myflow-do` |
| `awaiting-do-review` | `/myflow-do`; also `/myflow-fast-path` in `checkpoint` mode (resume by re-invoking `/myflow-fast-path <name>` — **not** `/myflow-do-done`) | human — review staged diff |
| `do-review-started` | `/myflow-do-manual-review` | human — mid-review |
| `do-done` | `/myflow-do-done` | `/myflow-manual-test` |
| `awaiting-fix-review` | `/myflow-do-fix` | human — review the fix |
| `fix-review-started` | `/myflow-do-fix-manual-review` | human — mid-review |
| `awaiting-manual-test` | `/myflow-manual-test` | human — run the apps |
| `manual-test-done` | `/myflow-manual-test-done` | `/myflow-review` |
| `awaiting-pr-review` | `/myflow-review` | human — review + merge the PR |
| `review-done` | `/myflow-review-done`, or `/myflow-review automerge` | `/myflow-finish` |
| `finished` | `/myflow-finish` | — terminal |

**`*-done` and `*-manual-review` commands are pure state writes.** They update `stage`,
`updatedAt`, and `updatedBy` and nothing else — no verification, no artifact reading, no git.
This keeps the *claim* ("I am done reviewing") separate from the *fact* ("the PR merged"), which
`/myflow-finish` still verifies independently. `/myflow-review-done` therefore checks nothing;
a change can reach `review-done` with an unmerged PR, and `/myflow-finish` will refuse to archive it.

## State file

The contract governing where a change's state file lives, its full JSON shape, monotonic gate
values, and carry-forward rules. **State file** (`skills/myflow-contracts/state-file.md`) — load
it before reading or writing a state file.

## State self-heal

The contract governing how a state file is validated against on-disk artifacts, and how a
missing, unparseable, or contradicted file is corrected.
**State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — load it before self-healing
a state file.

## Project configuration

The contract governing `.myflow/project.md` — its optional keys, how a `## standards` entry
resolves to a file, and the containment rules that keep resolution safe.
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) — load it before
resolving project configuration.

## Jira integration

The contract governing how a change is linked to a Jira issue, transitioned through the pipeline,
and has its description synced — including that Jira is never a gate and never blocks a stage
write. **Jira integration** (`skills/myflow-contracts/jira-integration.md`) — load it before any
Jira-related step.

## Stage transitions

Each command declares the stage(s) it accepts:

| Command | Accepts stage(s) |
|---------|------------------|
| `/myflow-start` | *(none — creates the change)* |
| `/myflow-start-fix` | `awaiting-proposal-review` |
| `/myflow-start-done` | `awaiting-proposal-review` |
| `/myflow-do` | `proposal-done` |
| `/myflow-do-manual-review` | `awaiting-do-review` |
| `/myflow-do-done` | `awaiting-do-review`, `do-review-started` |
| `/myflow-do-fix` | `awaiting-do-review`, `do-review-started`, `do-done`, `awaiting-manual-test`, `manual-test-done`, `awaiting-pr-review` |
| `/myflow-do-fix-manual-review` | `awaiting-fix-review` |
| `/myflow-do-fix-done` | `awaiting-fix-review`, `fix-review-started` |
| `/myflow-manual-test` | `do-done` (advance) · `awaiting-manual-test` (refresh) |
| `/myflow-manual-test-done` | `awaiting-manual-test` |
| `/myflow-review` | `manual-test-done` |
| `/myflow-review-done` | `awaiting-pr-review` |
| `/myflow-finish` | `review-done` |
| `/myflow-fast-path` | *(none — creates the change)* · `proposal-done` · `awaiting-do-review` **with `fastPath: true`** (checkpoint resume only) |
| `/myflow-status`, `/myflow-info` | any — read-only, never block |
| `/myflow-full` | *(composite — drives the sequence, always stops at Gate D)* |

Keep the existing mismatch-handoff prose and the AskUserQuestion-default-No rule unchanged.

**This table is authoritative.** Every command file — in **both** agents-repo command trees
(`agents/commands/` and `agents/commands-claude/`) — must state exactly the stage(s) its row lists,
and must agree with the skill it delegates to. The repo-root docs (`README.md`, `AGENTS.md`,
`CLAUDE.md`) must agree with it too. When a command and its skill disagree, whichever the agent
reads first wins, which is non-determinism in the one layer that must be deterministic. A renamed
stage is therefore not done until the command layer is swept too.

**`scripts/check-vocabulary.sh` in the agents repo is the regression check for this.** Run it
**with no arguments** — it is cwd-independent, so run it from whichever affected worktree carries
`scripts/`:

```bash
scripts/check-vocabulary.sh
```

**Pass no path list.** The set of trees and files to scan is the script's own default scan set — see
`DEFAULT_TARGETS` in the script — so that scope is defined in exactly **one** place. Restating the
list here, or at any other call site, would give the guard two scopes that drift apart, which is how
the repo-root docs came to sit outside it in the first place.

**What it actually does, and what it does not.** It greps for a fixed list of *known-retired
literals* — the retired stage values (`awaiting-review`, `awaiting-test`, bare `code-review`) and the <!-- vocab-guard:allow -->
retired review-panel roster spellings — reports every hit as `file:line`, and exits non-zero, so it
fails loudly rather than warning. That makes it a **regression check against vocabulary that has
already been retired once**, not a validator of the tables above. It matches literals only: it exits
clean on any paraphrase of a retired term (`primary plus five additional reviewers`, `six-agent <!-- vocab-guard:allow -->
panel`, `Senior engineer`), and it cannot know about the *next* rename until that rename's literals <!-- vocab-guard:allow -->
are added to it — the same blind spot as the incident it commemorates. **A clean run is therefore
not proof that a rename is complete.** It proves only that the specific strings recorded in it are
gone. Sweeping every layer by hand is still what makes a rename done.

`requesting-code-review` is a real external Superpowers skill and is exempt — never rename it. A
line that must quote a retired token to do its job (the ones above, and the legacy-stage migration
mapping) carries a trailing `vocab-guard:allow` marker; the marker is per-line on purpose, so real
drift cannot hide behind one documentary sentence. **`/myflow-review` runs this script as part of
its verification gate**, which is what keeps the guard reachable — a guard nothing invokes can never
fail.

**`scripts/test-setup.sh` is the agents repo's second guard, and the same rules apply to it.** It is
the regression harness for `setup.sh`, which writes into the user's home directory and has already
shipped two data-loss defects — reversed delimiters that silently deleted content, and an append
branch that could never converge. It runs every case against a sandboxed `HOME` under `/tmp`, never
touches the real `~/.claude`, `~/.cursor` or `~/.codex`, takes no arguments, and exits non-zero if
any assertion fails. Run it from whichever affected worktree carries `scripts/` (both scripts are cwd-independent):

```bash
scripts/test-setup.sh
```

**`/myflow-review` runs it in the same verification gate**, whenever the agents repo is among the
affected worktrees — not only when the diff touched `setup.sh`, because the rules and skills the
installer reads are its inputs. Treat a non-zero exit exactly like a failing test: fix `setup.sh` or
the input that broke it, never the assertion. **What it covers, and what it does not:** the
installer's observable filesystem effects — what exists after a run, what it points at, what it
contains, its mode, and the exit status — for the shapes listed in its own header. It says nothing
about whether the installed content is correct, and nothing about agent behaviour.

**`/myflow-manual-test` refresh mode.** When the incoming stage is already `awaiting-manual-test`, the command enters **refresh mode** automatically — no mismatch, no prompt, no override. Refresh mode does not re-ask the skip question (that decision was made on the advance run, which is what makes `"skipped"` unclobberable), does not rewrite `gates.reviewed`, preserves every already-checked box, re-emits `stage: awaiting-manual-test` with all gates carried forward, and announces "refreshed after fix round N".

**On mismatch, stop.** Report the actual stage, the expected stage, and the command that should run instead, then **AskUserQuestion** for an explicit override with **"No — run the suggested command instead"** as the default and recommended answer. Only proceed when the user explicitly chooses to override. Never advance the stage silently from a wrong starting point.

Handoff format on mismatch:

```
## Wrong stage for this command

**Change:** <name>
**Current stage:** <actual> (set by <updatedBy>, <updatedAt>)
**This command expects:** <expected>
**Suggested instead:** /myflow-<other> <name>
```

## Fix re-entry

`/myflow-do-fix` records `originStage` in the state file — the stage it was invoked from.
`/myflow-do-fix-done` returns the change to that stage, then clears `originStage`.

| Fix raised at | Returns to |
|---------------|-----------|
| `awaiting-do-review` | `awaiting-do-review` |
| `do-review-started` | `awaiting-do-review` |
| `do-done` | `do-done` |
| `awaiting-manual-test` | `awaiting-manual-test` |
| `manual-test-done` | `manual-test-done` |
| `awaiting-pr-review` | `awaiting-pr-review` |

`do-review-started` returns to `awaiting-do-review`, not to itself: the diff changed, so the
prior review is stale and review restarts. Every other origin returns to itself. A second fix
round overwrites `originStage` with the stage that round began from.

**Why `do-done` and `manual-test-done` are origins.** Both mean "the work is complete and the
human confirmed it, but something was found before the next command ran" — exactly when a fix is
wanted. Accepting `manual-test-done` is what lets `/myflow-review`'s own coverage check recommend
`/myflow-do-fix` without forcing a stage-mismatch override, since the change sits at
`manual-test-done` while that check runs. Accepting `do-done` closes the symmetric hole right
after `/myflow-do-done`.

`/myflow-do-fix` still commits and pushes **only** when `originStage` is `awaiting-pr-review`
(a PR exists remotely, so staging alone would leave the fix invisible). At every other origin it
stages without committing.

## IntelliJ commands

Every stage that waits on a human must print a copy-paste command in its handoff:

```bash
open -na "IntelliJ IDEA" --args "<absolute path>"
```

Use `open -na`, not the `idea` shim — that shim is not on this machine's PATH. `open` resolves
the app by name (bundle `com.jetbrains.intellij`), returns immediately instead of blocking the
shell, and reuses a running instance.

| Stage | Path to open |
|-------|--------------|
| `awaiting-proposal-review` | main checkout (artifacts live there) |
| `awaiting-do-review`, `do-review-started` | apply worktree root |
| `awaiting-fix-review`, `fix-review-started` | apply worktree root |
| `awaiting-manual-test` | apply worktree root, plus the guide's absolute path |
| `awaiting-pr-review` | apply worktree root |

Paths are absolute, resolved from `git worktree list`. Never emit a relative path.

## Stage boundaries

| Stage | Command | Allowed git actions |
|-------|---------|---------------------|
| Proposal | `/myflow-start` | None (planning artifacts only) |
| Do | `/myflow-do` | Branch/worktree + **`git add` (stage)** — no commits, push, merge, or PR |
| Manual review | User (Gate B) | User inspects **staged** diff in worktree IDE; agent waits |
| Fix at Gate B/C | `/myflow-do-fix` | Resumes **existing** worktree + **`git add`** — no commits, push, merge, or PR |
| Manual test | `/myflow-manual-test` | Generate/refresh guide + **`git add`** — no commits |
| Review | `/myflow-review` | Verifies Gate C, checks coverage, tests/linters, **commits**, **pushes**, **opens PR** — merges only with `automerge` (see **Auto-merge (opt-in)**) |
| PR review | User (Gate D) | User reviews the PR on the forge and merges it — unless `automerge` was used, in which case there is no PR to review |
| Fix at Gate D | `/myflow-do-fix` | **Commits and pushes** to the existing PR branch — the one exception (see **Fix rounds**) |
| Finish | `/myflow-finish` | Verifies the PR merged, syncs delta specs, archives — no commits, tests, or merges |

## Do (#2–#6 only)

- Run Superpowers steps **#2–#6** only. **Never** run `finishing-a-development-branch` (#7) during do.
- **Never** `git commit`, `git push`, merge, or open a PR during do — even if SDD implementer prompts say to commit.
- **`git add` is required** at end of do (and allowed during do) so Gate B shows staged changes in the IDE Source Control view.
- Record `MERGE_BASE` at worktree setup; before each SDD task record `TASK_BASE=$(git rev-parse HEAD)`.
- Per-task and final automated review (#6): review **working tree** diffs (`git diff TASK_BASE`, `git diff MERGE_BASE`), not commit ranges.
- Final #6 is a **strict multi-agent panel** with **three required slots** — primary `requesting-code-review`, Bugbot, and a **principles reviewer** — plus conditional slots (Security Review, Adversarial, extra principle lenses B/C) selected by the trigger table in `openspec-apply-superpowers` (**Optional slot selection**), which is canonical. The principles agent runs `principles-reviewer-prompt.md` on the **parent model** (no `model` override), checks the diff against `engineering-principles.md`, and owns the project's hard invariants read out of its standards files — core purity, new suppressions, weakened lint config. The Adversarial reviewer (slot 4) and the optional lens reviewers (slots 5+) use the provider's economy-tier model (e.g. Grok parent → `composer-2.5-fast`); slots 0 and 2 inherit the parent model. Do not skip or collapse any dispatched slot.
- **Panel pass 1 always runs the full roster selected for this change.** Re-runs after a fix round are **targeted** by default — primary plus the agents that raised the findings, over the fix-scoped diff — and escalate to the full whole-branch panel on the triggers in `openspec-apply-superpowers` (**Panel re-runs**), which is canonical. Gate D-origin fixes and the `full-panel` flag always run full. Handoff still requires a non-stale clean result from every slot in this run's roster.
- Progress ledger: `Task N: complete (uncommitted, review clean)` — not commit SHAs.
- Before handoff: in every affected repo/worktree, `git add -A` (respects `.gitignore`), confirm `git status` shows changes as **staged**, then hand off worktree path + review commands; tell user to run `/myflow-manual-test <name>` after manual review, then `/myflow-review` after testing.
- **Guard against a mistaken re-run:** if `<name>` already has an apply worktree and looks like it's already past do (a manual-test guide exists, or a clean final-review-panel record exists, or every original task is checked), **AskUserQuestion** before proceeding — did the user mean `/myflow-do-fix` instead? Default/recommended answer: **No, use `/myflow-do-fix`**. Only continue with a fresh/expanded `/myflow-do` run if the user explicitly picks **Yes**.

## Manual review (Gate B)

After do, **stop** unless the user explicitly skips review (`skip-review` flag on `/myflow-full` only).

Provide:

- Worktree path and branch name (open that folder in the IDE to see Source Control)
- Confirmation that changes are **staged and uncommitted**
- `git status`, `git diff --cached --stat`, and how to open full staged diff (`git diff --cached`)
- Hint: `/myflow-manual-test <name>` when code looks good; `/myflow-do-fix <name>` if changes needed (repeatable — loop as many rounds as needed)

Do not commit or integrate at Gate B.

## Manual test (Gate C)

After Gate B (and any Gate B fix rounds) are satisfied, user runs `/myflow-manual-test <name>` (or full-cycle continues into it).

- Agent writes/refreshes `docs/manual-test/<name>.md` (how to run involved apps + functionality checklist from specs)
- **Run commands must use absolute apply-worktree roots** for every app in scope — the apps listed under `## apps` in the project's `.myflow/project.md`, or the apps auto-detected from the repository when that file or key is absent (see **Project configuration**). Resolve each root from the progress ledger / `git worktree list`; never a relative sibling path (`../<other-app>`) and never a main-branch checkout while an apply worktree for that app exists
- Stages the file; replies with a **link to the file only** (no guide body in chat); **stops**
- User opens the link, runs the apps, and checks off items; does not move to review until satisfied
- Hint: `/myflow-review <name>` when testing done; `/myflow-do-fix <name>` for fixes (repeatable — refresh the guide after each fix, then re-test)
- `/myflow-manual-test` **always asks** (AskUserQuestion, default and recommended **No**): "Skip manual testing for this change?"
  - **No** → normal guide, unchecked checklist, `gates.tested: false`
  - **Yes** → same guide marked `**Manual test status:** SKIPPED`, every box left unchecked, `gates.tested: "skipped"`
- Either answer advances the stage to `awaiting-manual-test`. The distinction lives in `gates.tested`, which `/myflow-review` reads to tell an intentional bypass from an incomplete checklist.
- The skip prompt is asked **only on the advance run** (incoming stage `do-done`). In **refresh mode** (incoming stage already `awaiting-manual-test`, i.e. after a fix round) it is not re-asked and `gates.tested` is carried forward untouched.
- `gates.reviewed` is set `true` only when Gate B actually happened. Under `/myflow-full skip-review`, write `gates.reviewed: false` and note in the summary that Gate B was explicitly skipped — the state must never claim a review that never happened.
- `/myflow-manual-test-skip` is **retired** — the prompt replaces it.

Do not commit or integrate at Gate C.

## Fix rounds (Gate B / Gate C findings)

When Gate B or Gate C surfaces something that needs fixing, use `/myflow-do-fix <name>` — **not** a bare re-run of `/myflow-do <name>`. This is what keeps the change's proposal from going stale after review/test rounds: the fix is documented (appended to `proposal.md`/`tasks.md`, or tracked as a linked nested `<name>-fix-N` sub-change — the skill asks which) **before** it is implemented. Both gates can loop through as many fix rounds as needed before continuing.

- Resumes the **existing** apply worktree/branch — never a new one.
- No commits, push, merge, or PR — same constraints as do (commits now happen in `/myflow-review`, later than before), unless `originStage` is `awaiting-pr-review` (see **Fix re-entry**).
- Re-runs the strict review panel before handing back to Gate B — **targeted by default** (primary
  plus the agents that raised the findings, over the fix-scoped diff), escalating to the **full**
  whole-branch panel on the triggers in `openspec-apply-superpowers` (**Panel re-runs**), which is
  canonical. A Gate D-origin fix and the `full-panel` flag always run full. Either way, handoff
  requires a non-stale clean result from every slot in this run's roster.
- A Gate C fix invalidates prior checked-off boxes in `docs/manual-test/<name>.md` for anything the fix touched — refresh via `/myflow-manual-test <name>` before re-testing.
- A nested `<name>-fix-N` sub-change is never archived standalone — only together with its parent via `/myflow-finish <name>`.

## Review (commit + PR; merge only with `automerge`)

Requires stage `manual-test-done`. After Gate B and Gate C are both satisfied, run `/myflow-review <name>` (optionally `/myflow-review <name> automerge`).

- **Verify Gate C first:** read `gates.tested` and `docs/manual-test/<name>.md`. `"skipped"` means intentionally bypassed — record and continue, and **never** overwrite it. Otherwise, if any box is still `- [ ]`, notify the user (count + what's open) and get explicit confirmation before proceeding. If `gates.tested` is `false` and every checklist box is ticked, **promote it to `true`** before writing state — `/myflow-review` is the only writer of `gates.tested: true`.
- **Check test coverage** against every delta-spec scenario for `<name>` and any nested `<name>-fix-N`. Report gaps and offer `/myflow-do-fix` (recommended). Never write the missing tests in this stage.
- Run verification (tests/linters) with evidence shown.
- **Commit** all implementation changes, including `docs/manual-test/<name>.md`.
- **Push the branch and open a PR.** By default, **do not merge** — see **Auto-merge (opt-in)** for the one exception.
- **PR opening is forge-agnostic.** Branch on actual capability, never on the assumption that `gh` exists:
  1. `gh` installed **and** the remote is a GitHub host → `gh pr create` (and `gh pr list` to detect/reuse an existing PR) → `stage: awaiting-pr-review`, `gates.prOpened: true`.
  2. A remote exists but there is no usable PR CLI for that host (e.g. Bitbucket, or `gh` missing) → push, print the forge's create-PR URL, and **AskUserQuestion**: "Have you opened the PR?" (default **No**). **Yes** → `stage: awaiting-pr-review`, `gates.prOpened: true` (record the PR URL if the user supplies one; do not block if not). **No** → stay at `manual-test-done`, `gates.prOpened: false`, and say plainly what to do next. Human confirmation is a legitimate substitute for an API probe — the same trust model already used at Gates B and C.
  3. No remote at all → push is impossible; stop, stay at `manual-test-done`, `gates.prOpened: false`.
- The state file is **not** committed or pushed — it lives outside the repo (see **State file**).

## Auto-merge (opt-in)

**Never merge unless the user explicitly passed `automerge` to `/myflow-review`.** This is the
only exception in the system and it is defined here; skills reference this section rather than
each restating a conditional.

- **Default (no flag):** commit, push, open a PR (or print the forge URL and confirm), stop at
  `awaiting-pr-review`. The human reviews and merges.
- **`/myflow-review automerge`:** commit, push, merge into the base branch, push the merge. There
  is no PR for a human to review, so `awaiting-pr-review` is skipped and the stage goes directly
  to `review-done`. Announce plainly that auto-merge was used.

`automerge` must be typed by the user on that invocation. It is never inferred, never remembered
between runs, and never defaulted on.

## PR review (Gate D)

By default, the human reviews the PR on the forge, optionally loops `/myflow-do-fix <name>`, and
**merges it manually**. The agent never merges unless the user explicitly passed `automerge` to
`/myflow-review` (see **Auto-merge (opt-in)**) — in which case this gate was skipped entirely and
the change is already at `review-done`. Nothing in myflow automates a default merge.

## Finish (verify merged + sync + archive)

Requires stage `review-done`.

- **Verify the PR actually merged**, or — when `automerge` was used — that the merge commit landed on the base branch. If it is still open (non-automerge path), **stop** — archiving a change whose code never reached the base branch is a real drift risk.
- Archive any nested `<name>-fix-N` sub-changes together with `<name>`.
- Delta sync (if chosen), archive move, then write `stage: finished`, `gates.prMerged: true` to the user-scoped state file. The state file is **not** moved into the archive — it stays at its user-scoped path as the terminal record.
- `gh` may be unavailable; the `git merge-base --is-ancestor` check must remain reachable on its own and is the merge evidence on non-GitHub forges.
- **Never** commit, run tests, or merge here.

## Full cycle gates

| Gate | When | Default action |
|------|------|----------------|
| A | After proposal | User approves plan before do (`/myflow-start-done` is a separate human confirmation — `/myflow-full` never invokes it) |
| B | After do | **Manual review** — stop; inline `/myflow-do-fix` loop available (`/myflow-do-done` and `/myflow-do-manual-review` are human confirmations — `/myflow-full` never invokes them) |
| C | After `/myflow-manual-test` | **Manual testing** — stop; inline `/myflow-do-fix` loop available |
| D | After `/myflow-review` | **Human PR review + merge** — stop; `/myflow-full` always ends here, unless the user explicitly ran it with `automerge`, in which case there is no PR to stop at and the cycle ends at `review-done` |
| E | `/myflow-finish` | Verify merged + sync + archive (user invoked) |

**`/myflow-full` must never auto-invoke a `*-done` or `*-manual-review` command.** Those are
human confirmations of a stage a human actually reviewed; a full-cycle run that invokes them
itself would self-certify a review nobody performed. `/myflow-full` drives the sequence up to and
including `/myflow-do`, `/myflow-manual-test`, and `/myflow-review`, and always stops at a human
gate — it never runs a `*-done` command itself.

**Crossing Gate B/C on an explicit "Continue".** The prohibition above stops the agent from
self-certifying a review *nobody performed*. It does not apply when the human is asked, the cycle
stops, and the human answers **Continue to manual test** (Gate B) or **Continue to review**
(Gate C): that answer *is* the confirmation `/myflow-do-done` / `/myflow-manual-test-done` would
record, made in the moment, inside the cycle. So on Continue, `/myflow-full` writes the `*-done`
stage directly (`do-done` / `manual-test-done`) with honest gate values and proceeds. It still
never invokes the `*-done` command, still stops and waits for the answer, and still never writes
`gates.tested: true` (only `/myflow-review` writes that). **"Stop here", an aborted prompt, or any
other point in the cycle never writes a `*-done` stage.** This is the same reasoning the flags
below rest on — see the paragraph in **Opt-out (explicit only)**.

## Opt-out (explicit only)

- `skip-review` on `/myflow-full` — skip Gate B only; Gate C still runs unless also skipped. Records `gates.reviewed: false` (Gate B explicitly skipped), never `true`.
- `skip-manual-test` on `/myflow-full` — pre-answers the Gate C skip prompt with **Yes**, then carries the change through to `manual-test-done` with `gates.tested: "skipped"` (the Gate C question is not asked); review still runs and still checks coverage.
- `full-panel` on `/myflow-do` and `/myflow-do-fix` (and `/myflow-full`, which forwards it) — force every review-panel re-run to use every roster slot — including both extra principle lenses — over the whole-branch diff, disabling targeted re-runs. Opt-in; the default is targeted with automatic escalation.
- `checkpoint` on `/myflow-fast-path` — adds a Gate B stop on the staged diff after verification, before anything is pushed. On **Continue**, the cycle writes `do-done` directly with `gates.reviewed: false` (the same in-the-moment-confirmation reasoning as **Crossing Gate B/C on an explicit "Continue"**); on **Stop here** the change is left at `awaiting-do-review` and re-invoking the command resumes it. `/myflow-fast-path` itself always records `gates.reviewed: false` and `gates.tested: "skipped"` — it never claims a gate that nobody ran — and never accepts `automerge`.
- `commit-during-apply` — legacy SDD per-task commits; only when the user explicitly requests it.
- `automerge` on `/myflow-review` (and therefore on `/myflow-full`, which forwards it) — see **Auto-merge (opt-in)**. Opt-in only; never implied by any other flag.

**How `skip-review` and `skip-manual-test` reach a `*-done` stage without `/myflow-full` invoking a `*-done` command.** Elsewhere this file forbids `/myflow-full` from auto-invoking any `*-done` or `*-manual-review` command, because those are human confirmations of a review that actually happened — self-writing one would self-certify a review nobody did. `skip-review` and `skip-manual-test` are the deliberate exception, not a contradiction of that rule: typing the flag **at invocation time** is itself the human's explicit, in-the-moment opt-out decision — a stated instruction, not the agent inferring or assuming consent after the fact. So when `skip-review` is passed, the cycle writes `stage: do-done` directly (with `gates.reviewed: false`, honestly recording that Gate B was skipped, never `true`) as it moves past the Gate B stop — this is the flag being honored, not a `/myflow-do-done` invocation, and no review is claimed. The identical logic applies to `skip-manual-test` reaching `manual-test-done` (with `gates.tested: "skipped"`): the flag at invocation is the human's decision, so writing the stage forward is executing that decision, not self-certifying testing nobody did. And the same logic again covers a human answering **Continue** at a Gate B/C prompt — see **Crossing Gate B/C on an explicit "Continue"** under **Full cycle gates**. In all three cases a human's explicit instruction authorizes the stage write; in none of them does `/myflow-full` invoke a `*-done` command.

`no-archive` is **removed** — `/myflow-full` now always ends at Gate D (or, with `automerge`, at `review-done`) and never reaches archiving.
