---
name: myflow-finish
description: Two-run finish — integrate the branch (open a PR, merge, or leave it manual), then once merged sync delta specs, archive, push, and remove the worktrees. Use for /myflow-finish.
allowed-tools: Bash(openspec:*)
license: MIT
---

Finish a change. This command runs **twice**, and which run happens is decided by one thing:
whether the change's branch has already reached the base branch.

**Announce at start:** "Using myflow-finish for change `<name>` — run 1 (integrate)." or
"— run 2 (archive)."

Immediately after that line, print these two commands for the operator to paste, per
**Handoff output** (`skills/myflow-contracts/pipeline.md`) — that section fixes the colour and
records why they are printed rather than invoked; do not restate its reasoning here:

```text
/rename <change-name>
/color cyan
```

**Load `skills/myflow-contracts/pipeline.md` first.**

**Then register this run's steps** with the harness's task-list mechanism and keep each entry's
status current as the run proceeds, per
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) — that section names which steps
this command registers and is the one to read. What is specific to this command, and so stated
here: run 1 and run 2 have different step lists, so the entries registered are the steps of the run
`<agents repo>/scripts/check-finish-preflight.sh`'s verdict named, and none are registered before that verdict is
in hand. A run that registered run 1's steps and then archived would show the operator a list that
never matched the work.

The reasoning behind this file lives in `skills/myflow-finish/SKILL-rationale.md`; **a
`/myflow-*` run never loads it.**

## State gate

Accepts **`IN_PROGRESS`**. Run 1 ends at `IN_PROGRESS`; run 2 ends at `FINISHED`.

At `STARTED` there is nothing to integrate — emit the wrong-state handoff and recommend
`/myflow-do`. At `FINISHED` the change is already archived; say so and stop.

## Deciding which run this is

**`pipeline.md`'s Finish contract is canonical for every procedure below.** Load it and follow it
there — the base-branch resolution, the preflight checks, the removal sequence and their
rationales live in that one file. This skill carries only what is specific to *executing* it.

No field records "integration started" — a field could disagree with git.

**Check guard presence.** Per **Guard presence check** (`skills/myflow-contracts/pipeline.md`),
confirm every guard this command invokes — `check-finish-preflight.sh`, `resolve-base-branch.sh`,
`prepare-archive-branch.sh`, `check-unfinished-work.sh`, `preserve-session-records.sh`,
`commit-split.sh`, `check-cleanup-complete.sh` and `gather-self-review-context.sh` — is present in
`skills/myflow-finish/scripts/`. A complete set prints nothing; any absence prints that section's
block once, and the run continues under each guard's own hand-run fallback.

Run `check-finish-preflight.sh` once per worktree in the set found by **Resolving a change's
worktrees** (`skills/myflow-contracts/finish-contract.md`) — never a raw read of the state file's
`worktrees` map, which a `{}` map would make pass having checked nothing — and act on the verdict.
Its `<base-ref>` argument is `origin/$BASE`, `$BASE` being what `resolve-base-branch.sh` prints for
that worktree — the remote-tracking ref, never the bare local name: `resolve-base-branch.sh`'s fetch
refreshes only remote-tracking refs, so a stale local branch of the same name would silently feed
this verdict.

- **`RUN1`** → run 1 (integrate)
- **`RUN2`** from every worktree → run 2 (archive and clean up)
- **`REFUSE`** → stop, report what the script reported, and ask the operator before anything else
- **A resolved set that comes back empty** → stop and ask the operator, exactly as `REFUSE`, per
  **Resolving a change's worktrees** (`skills/myflow-contracts/finish-contract.md`)
- **No verdict line at all, and exit 2** → the script could not read the tree, which is a fourth
  outcome and not a verdict. Treat it exactly as `REFUSE`: stop, report the message it printed on
  stderr, and ask the operator. Never re-run it hoping for a verdict, and never read the missing
  line as either run — a caller that greps for `RUN2` in empty output finds nothing, which is why
  the exit code has to be checked as well as the line.

**The base branch comes from `resolve-base-branch.sh`, never assumed and never derived from the
current branch.** Run it against the apply worktree; its exit contract is the one **Finish contract**
(`skills/myflow-contracts/finish-contract.md`) is canonical for. Anything but exit `0` — **stop and
ask**. This skill runs inside the apply worktree, where `HEAD` is the change's own branch — so any
resolution that consults `HEAD`'s upstream would compare the branch against itself and report every
pushed branch as merged. Run 2 resolves the base the same way, from the apply worktree, **before
cleanup removes it**: the archive commit is run 2's step 4, cleanup is step 5, so the worktree is
still there when resolution runs.

---

# Run 1 — integrate

**On a `RUN1` verdict**, mark the Level 1 table's first two stages: the preflight verdict just
taken above (`finish.preflight`, closed immediately since the verdict is already in hand), then
run 1's unfinished-work gate (`finish.unfinished-work-gate`), per **Stage marks**
(`skills/myflow-contracts/pipeline.md`). A `RUN2` verdict marks nothing here; its own first mark is
`finish.verify-merge`, below.

**Generate this run's session token once, right here, before the first mark — a short, unique
literal string — and reuse that exact same value, unchanged, at every `stage begin` run 1 makes
below.** Never mint a fresh token per mark (design.md's "one token per session, not one per mark").
Run 2, a separate invocation, generates its own token at its own first mark, below.

```bash
myflow stage begin -command '/myflow-finish' -stage finish.preflight -harness <harness> -session-token mf-<literal-token> <name>
myflow stage end   -command '/myflow-finish' -stage finish.preflight -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.unfinished-work-gate -harness <harness> -session-token mf-<literal-token> <name>
```

## 1.0 Check for unfinished work

Run `check-unfinished-work.sh <worktree> <name>` once per worktree in the set found by
**Resolving a change's worktrees** (`skills/myflow-contracts/finish-contract.md`) — never a raw read
of the state file's `worktrees` map, for the same reason **Deciding which run this is** above does
not read it raw — **before the landing question and before any git action**. See
**Finish contract** (`skills/myflow-contracts/finish-contract.md`).

- **`CLEAR:` from every worktree** → continue to **1.1** with no extra prompt.
- **A resolved set that comes back empty** → stop and ask the operator, per **Resolving a change's
  worktrees** (`skills/myflow-contracts/finish-contract.md`)
- **`OUTSTANDING:`** → show the breakdown the script printed, and offer exactly three courses —
  shape per Operator prompts (`skills/myflow-contracts/operator-prompts.md`):

  > **This change carries unfinished work — how should run 1 proceed?**
  > - **Stop — I'll finish it first** *(recommended)*
  > - **Continue — integrate anyway**
  > - **File or join a Jira follow-up, then continue**

  There is no fourth, and in particular none that hands back to `/myflow-do` inline.
- **No verdict line at all, and a non-zero exit** → stop and ask the operator. Never read the
  missing line as either verdict — a caller that greps for `CLEAR` in empty output finds nothing,
  which is why the exit code has to be checked as well as the line.

**Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed.
**Continue** carries the outstanding list into **1.2**'s planning commit and into **1.5**'s handoff.
**File or join a Jira follow-up** puts the outstanding items on a **follow-up** issue and then
continues — joining an open one when the search finds a candidate and the operator confirms it, and
filing a new one otherwise. See **1.0 Check for unfinished work**
(`skills/myflow-finish/SKILL-rationale.md`) for why the course is labelled this way. What that
follow-up is titled, the search, the confirmation, what a failed search does, and how it is
labelled are all
**Follow-up issues** (`skills/myflow-contracts/jira-followups.md`), and none of it is restated
here; **a filing that fails is one skipped-with-reason line and the run still continues**, per
**Never blocking** (`skills/myflow-contracts/jira-integration.md`) — the outstanding list still
reaches the planning commit and the handoff, so the durable record does not depend on the tracker.
Why **Stop** is the marked recommendation is stated under
**Finish contract** (`skills/myflow-contracts/finish-contract.md`) and is not re-argued here.

Close the `finish.unfinished-work-gate` mark opened above, whichever way this gate resolved:
`completed` on **Continue** or **File or join a Jira follow-up, then continue**, `stopped` on
**Stop** (the run ends here, at `IN_PROGRESS`, with nothing staged, committed or pushed):

```bash
myflow stage end -command '/myflow-finish' -stage finish.unfinished-work-gate -outcome completed <name>
```

## 1.1 Ask how the branch should land, before any git action

```bash
myflow stage begin -command '/myflow-finish' -stage finish.landing-question -harness <harness> -session-token mf-<literal-token> <name>
```

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Having asked once, run to completion without asking again. **Never** remember the answer between
runs, and never infer it from anything else.

**Report an existing PR before asking.** Run 1 is re-entered whenever the branch is not merged, so
a previous attempt may already have opened one. If a PR exists for this branch — from `prUrl`, or
from a PR CLI when one is usable — say so, including whether it is open or closed-unmerged, before
the operator answers. See **1.1 Ask how the branch should land, before any git action**
(`skills/myflow-finish/SKILL-rationale.md`) for why.

```bash
myflow stage end -command '/myflow-finish' -stage finish.landing-question -outcome completed <name>
```

## 1.2 Commit the staged work

```bash
myflow stage begin -command '/myflow-finish' -stage finish.preserve-sessions -harness <harness> -session-token mf-<literal-token> <name>
```

**Before any route commits, reshape the branch.** Run
`git -C <worktree> reset --soft <recorded-merge-base>`, where `<recorded-merge-base>` is the
merge base recorded in the state file's `worktrees` map for this worktree — the same merge base
**Resolving a change's worktrees** and the finish-preflight verdict above both reference. This
collapses every per-task and fixup commit `/myflow-do` made on the branch back into the working
tree, uncommitted, so the branch carries no history for the two-commit chain below to inherit —
that chain then commits from this reshaped state exactly as it always has.

All three routes commit — implementation, the `<project>/openspec/` planning
artifacts, and the session records preserved under `<project>/docs/superpowers/` — as **two** commits, never
one.

**Preserve the session records first.** Run
`preserve-session-records.sh <worktree> <name> <state-dir>` before staging, so the SDD ledger,
the review panel record and the proposal artifact source are committed with the change — in the
planning commit, beside the plan they describe — rather than lost with the worktree that holds them.
A source that does not exist is reported and skipped — never a failure, and never a reason
to stop the integration. **A non-zero exit means a copy was attempted and refused or failed:** report
it with the script's own stderr message and continue the integration, per the outcome table under
**Preserving the session records** in `skills/myflow-contracts/pipeline.md`. Say in the handoff
which records were preserved and which were not.

```bash
myflow stage end   -command '/myflow-finish' -stage finish.preserve-sessions -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.commit-two -harness <harness> -session-token mf-<literal-token> <name>
```

Then stage and commit twice, in this order, rather than assuming everything is already staged: the
operator may have edited the worktree at the human gate without staging.

```bash
commit-split.sh <worktree> <name> \
  "<type>(<module>): <what the implementation does>" \
  "chore(openspec): plan and session records"
```

`<type>`, `<module>` and `<what the implementation does>` are derived from the reshaped diff — this
run's own read of what the branch changed and which module carries it — since the `reset --soft`
above collapsed every task and fixup commit back into the working tree, so there is no per-task
subject left to reuse. The planning message is a **fixed literal**, never derived: every planning
commit stages the same two trees in every change, so nothing about it varies.

**Run that as one command.** The guards, the skipped-empty rule, the stop-on-failure rule and the
symlinked-planning-path case are all stated under
**Git boundaries** (`skills/myflow-contracts/pipeline.md`) and are not re-argued here — this call
implements that chain rather than restating it. In short:
an empty commit is **skipped, not an error** — a fix that touched only the planning paths, a fix
that touched only implementation, and a re-run after a rejected push all reach this block with one
side or both already satisfied — while a commit that FAILS stops the chain and is reported with
git's own output, so a rejected first commit can never fall through into a single commit carrying
everything.

**Implementation first, planning artifacts second.** The second commit's message lists anything the
operator chose to integrate over at **1.0** — the git history is then the durable record that the
transcript is not. Say in the handoff when a commit was skipped as empty; a silently missing commit
looks like a lost one. See **1.2 Commit the staged work**
(`skills/myflow-finish/SKILL-rationale.md`) for why.

The state file is **not** committed — it lives outside the repo.

```bash
myflow stage end -command '/myflow-finish' -stage finish.commit-two -outcome completed <name>
```

## 1.3 Take the chosen route

```bash
myflow stage begin -command '/myflow-finish' -stage finish.landing-routes -harness <harness> -session-token mf-<literal-token> <name>
```

Per **Finish contract** (`skills/myflow-contracts/finish-contract.md`) → run 1. Push with `-u` so the
branch has an upstream; the unpushed-commits check in run 2 treats a missing upstream as unknown
and refuses to clean up without one.

**Every git step here can fail, and none of them may fail silently.** A rejected push
(non-fast-forward), a merge conflict, a commit blocked by a hook, or `gh pr create` erroring
(expired auth, rate limit, an existing PR for the head branch) must be **reported with the command's
own output**, and the run must **stop** leaving the change at `IN_PROGRESS`. Say which step
succeeded and which did not, so re-running `/myflow-finish` resumes from a known place. A `gh`
failure is distinct from `gh` being absent: absence falls through to the print-the-URL path,
failure stops and reports.

If there is no remote at all, pushing is impossible and no route can complete — say exactly that
("this repository has no remote, so there is nothing to push to or merge into"), not that the base
branch failed to resolve, and change nothing else. The work stays staged at `IN_PROGRESS`; nothing
is lost, because nothing was pushed.

**Human confirmation is a legitimate substitute for an API probe** on a forge with no usable CLI;
it is the same trust model the human gate rests on. If the answer is No, leave `prUrl` null and say
what to do next.

```bash
myflow stage end -command '/myflow-finish' -stage finish.landing-routes -outcome completed <name>
```

## 1.4 No verification gate

**Run no tests, no linters, and no spec-coverage check** — see
**Finish contract** (`skills/myflow-contracts/finish-contract.md`). Correctness was established during
`/myflow-do` and by the human gate.

## 1.5 State and handoff

**This section carries the Level 1 table's last two run-1 stages, in the order the table lists
them**: the state write happens first, and the Jira transition follows it — never before, per
**Transitions** (`skills/myflow-contracts/jira-integration.md`), cited below. See **Stage marks**
(`skills/myflow-contracts/pipeline.md`).

```bash
myflow stage begin -command '/myflow-finish' -stage finish.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

Write the state file with `state` unchanged at `IN_PROGRESS`, `prUrl` set if a PR was opened, and
every other field carried forward.

```bash
myflow stage end -command '/myflow-finish' -stage finish.write-in-progress -outcome completed <name>
```

```bash
myflow stage begin -command '/myflow-finish' -stage finish.move-in-review -harness <harness> -session-token mf-<literal-token> <name>
```

**Transition the issue to In Review** at the end of a successful run 1, whichever route was taken —
pull request, merge and push, or manual. Per
**Transitions** (`skills/myflow-contracts/jira-integration.md`): after the state write, never
before, never blocking. A run that stopped on a failed push does **not** transition; the branch
never left the operator's hands.

```bash
myflow stage end -command '/myflow-finish' -stage finish.move-in-review -outcome completed <name>
```

```
## Branch integrated — waiting on the merge | merged and waiting on run 2

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it
**Outstanding:** <what 1.0 reported and the operator integrated over> | none
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)

<what the operator must do before the next run>

Next:
/myflow-finish <name>
```

**Select the heading from the route this run actually took**, not from a claim about run 1 in
general:

| Route taken in 1.3 | Heading |
|--------------------|---------|
| pull request | *waiting on the merge* — the PR is open and unmerged |
| **merge and push** | *merged and waiting on run 2* — 1.3 merged it into the base branch itself |
| manual | *waiting on the merge* — the operator has the branch and has not merged it yet |

Where the route is not certain — a run resumed after a partial
failure — take the answer from the merge-status test in
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) rather than assuming; it is
the same test `/myflow-status` uses to regenerate this block for a change whose branch has since
been merged, a run stopped at a run-2 cleanup leftover most often. See **1.5 State and handoff**
(`skills/myflow-finish/SKILL-rationale.md`) for why the heading depends on the route taken.

---

# Run 2 — archive and clean up

Follow **Finish contract** (`skills/myflow-contracts/finish-contract.md`) → run 2 for the full
procedure. In outline, and stopping at the first step that fails, each numbered step below is
bracketed by its own mark per **Stage marks** (`skills/myflow-contracts/pipeline.md`), using that
step's exact name from the Level 1 table — with three exceptions, each of which opens no mark of its
own and runs inside the mark of the step before it: step 2 (positioning) inside step 3's
`finish.sync-archive`, per that mark's own reworded Level 1 Name; step 6 (remove the proposal
artifact source) inside step 5's `finish.cleanup`, as it already did before this change; and step 11
(restore the checkout) inside step 10's `finish.push-archive`. **Eleven steps, eight marks.** A
failed mark never blocks, delays or alters the step it brackets.

**Generate this run's own session token once, right here, before this first mark — a short, unique
literal string — and reuse that exact same value, unchanged, at every `stage begin` run 2 makes
below.** This is a separate invocation from run 1 above, so it carries its own, freshly-generated
token, never run 1's.

```bash
myflow stage begin -command '/myflow-finish' -stage finish.verify-merge -harness <harness> -session-token mf-<literal-token> <name>
```

1. **Verify the merge** — a PR CLI when usable, otherwise `git merge-base --is-ancestor`, which
   must stay reachable on its own as the only evidence on a non-GitHub forge. Fetch first so the
   tracking ref is current. Not merged → this is not run 2; fall back to run 1 and **archive
   nothing** — end this mark `-outcome not-run-2` and stop; nothing below runs.

```bash
myflow stage end   -command '/myflow-finish' -stage finish.verify-merge -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.sync-archive -harness <harness> -session-token mf-<literal-token> <name>
```

2. **Position the main checkout on the archive branch**, before anything else touches it — the same
   mark as step 3 below, since positioning is what that mark's name now covers. Resolve `<base>`
   with `resolve-base-branch.sh` against the apply worktree, exactly as **1.3 Take the chosen
   route** above resolves it for run 1, then invoke
   `prepare-archive-branch.sh <main-checkout> <base> chore/archive-<name>`. Exit `0` → the checkout
   is on `chore/archive-<name>`, cut from a fast-forwarded `<base>`; continue to step 3. Anything
   else stops run 2 here, with nothing staged, committed, pushed or removed — report the guard's
   own message and leave the change at
   `IN_PROGRESS`. The four exit codes are **Run 2 — the branch is merged**
   (`skills/myflow-contracts/finish-contract.md`), step 2, and are not restated here.

   **When the guard is absent**, perform the same positioning by hand, in the same order, and say in
   the handoff that it was done manually — `prepare-archive-branch.sh`'s own header is the authority
   for the exact commands, per that same step 2.

3. **Sync delta specs, then archive.** Assess each delta in `<changeRoot>/specs/` against
   `<project>/openspec/specs/`, show a summary, and offer: sync now (recommended), archive without syncing,
   or cancel. Apply `## ADDED` by appending (creating the capability spec if absent), `## MODIFIED`
   by replacing the block matched on its `### Requirement:` heading whitespace-insensitively,
   `## REMOVED` by deleting it, `## RENAMED` in place preserving the body. Then move the change to
   `<project>/openspec/changes/archive/<YYYY-MM-DD>-<name>/`, taking any nested `<name>-fix-N` with it.

```bash
myflow stage end   -command '/myflow-finish' -stage finish.sync-archive -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.commit-archive -harness <harness> -session-token mf-<literal-token> <name>
```

4. **Commit the archive** on `chore/archive-<name>` in the main checkout — no push here; step 10
   below carries it. Assert the branch in the same guarded `&&` chain rather than assuming it,
   mirroring the shape **Git boundaries** (`skills/myflow-contracts/pipeline.md`) already documents:

   ```bash
   [ "$(git -C <main-checkout> branch --show-current)" = "chore/archive-<name>" ] \
     && git -C <main-checkout> add -A \
     && { git -C <main-checkout> diff --cached --quiet \
          || git -C <main-checkout> commit -m "chore(openspec): archive change, sync delta specs"; }
   ```

   A branch mismatch is reported, naming the branch found, and stops the commit, leaving the change
   at `IN_PROGRESS` — `git -C <main-checkout>` fixes the directory, never the branch.

   The subject is the fixed literal shown, scope included, per **Finish run 2's two commits carry
   module scopes** (`<agents repo>/openspec/specs/myflow-commit-scope/spec.md`).

```bash
myflow stage end   -command '/myflow-finish' -stage finish.commit-archive -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket.** The workspace half runs the project's `remove` command, read from the
   command table **Project configuration** (`skills/myflow-contracts/project-configuration.md`)
   defines and run from the **main checkout**, which that same table states, with the workspace id
   substituted into its text. **Nothing hands this run that id** — it is
   re-derived here from the change name, which is the only thing it ever comes from, per
   **The workspace id** (`skills/myflow-contracts/workspace-isolation.md`). A project declaring no
   `## workspace isolation` section, or no `remove` command in it, has this half **skipped, not
   failed**. A removal that fails is **the one exception to the stop-at-the-first-failure rule
   above**: report it and carry on to step 7, which decides the verdict from the project's survivor
   report and never from this command's exit code, per
   **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`). A worktree half that
   stops on a failed check takes the removal with it, and why the removal follows that half at all —
   the stack is down only once its check 5 has run — is
   **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`);
   see **Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`).
6. **Remove the proposal artifact source** from the state directory, on the condition its row in
   **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`) gives.

Steps 5 and 6 together are the Level 1 table's one `cleanup` stage:

```bash
myflow stage end   -command '/myflow-finish' -stage finish.cleanup -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.verify-cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once
   per repository, after every removal above. `COMPLETE:` → report the cleanup as verified, **relay
   every clause the line carries after ` — ` word for word**, and go on to step 8 — a `SKIPPED:`
   clause there says a registry row was not verified, so reporting only "cleanup verified" tells the
   operator the opposite of what the guard said, and a skip is never a pass. `LEFTOVER:` → name what
   remains and **stop without writing `FINISHED`**, leaving
   the change at `IN_PROGRESS`. **No verdict line at all, and a non-zero exit** → report it, leave
   the affected `worktrees` entries in the state file, and treat it as `LEFTOVER`; the exit code is
   checked as well as the line, because a caller that greps for `COMPLETE` in empty output finds
   nothing. Why a leftover blocks the write, and why run 2 is safe to re-enter afterwards, is
   canonical under **Finish contract** (`skills/myflow-contracts/finish-contract.md`).

Close this mark with the verdict step 7 reached: `completed` on `COMPLETE:`, `leftover` on
`LEFTOVER:` or on a missing verdict line — either way the run stops here, at `IN_PROGRESS`, and
nothing below runs:

```bash
myflow stage end -command '/myflow-finish' -stage finish.verify-cleanup -outcome completed <name>
```

```bash
myflow stage begin -command '/myflow-finish' -stage finish.write-finished -harness <harness> -session-token mf-<literal-token> <name>
```

8. **Write `FINISHED`** — reached only on `COMPLETE:` — clearing from `worktrees` **only the entries
   whose removal actually succeeded**. Carry `artifactUrl`, `jiraIssue`, `planningEffort`, `models`,
   `reviewPanelRoster` and `prUrl` forward — read the planning effort through the retired-key
   fallback, per **State file** (`skills/myflow-contracts/state-file.md`). This is the terminal
   write, so a field dropped here is dropped for good. The state file stays at its user-scoped path
   as the terminal record — it is **never** moved into the archive.

```bash
myflow stage end -command '/myflow-finish' -stage finish.write-finished -outcome completed <name>
myflow stage begin -command '/myflow-finish' -stage finish.self-review -harness <harness> -session-token mf-<literal-token> <name>
```

**Transition the issue to Done** after the state write, per
**Jira integration** (`skills/myflow-contracts/jira-integration.md`). A run that stopped at step 7
transitions nothing — the change is not done.

9. **Run self-review.** The procedure — skippable per run with running it the default, gathering
   input via a script rather than an inline re-read, one combined reasoning pass across all five
   angles plus the rating, the per-angle filing ask, and the report path — see
   **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`), step 9. The
   requirement to change first when that procedure changes is
   **Requirement: Self-review runs only after FINISHED is written**
   (`<agents repo>/openspec/specs/myflow-self-review/spec.md`). What is specific to *executing* it here: the
   script invocation `gather-self-review-context.sh
   <archived-change-path> <name> <state-dir> <main-checkout>`, resolving `<archived-change-path>` as
   `<project>/openspec/changes/archive/<YYYY-MM-DD>-<name>/` using the same date step 3 (sync + archive)
   already used when it moved the change there, and passing `<main-checkout>` as the trust anchor so
   the script's `--git-common-dir` derivation no longer depends on this run's own working directory.

   The skip prompt fires first, and reads — shape per Operator prompts
   (`skills/myflow-contracts/operator-prompts.md`):

   > **Run self-review for this change?**
   > - **Yes — run it** *(default, recommended)*
   > - **No — skip**

   An explicit **No** stops step 9 here; the handoff's `Self-review` line reads `skipped`. A
   session with no interactive channel to present this prompt at all is a distinct case from
   silence — silence still gets the prompt and defaults per Operator prompts
   (`skills/myflow-contracts/operator-prompts.md`), but a session that cannot ask still runs
   self-review, exactly as an explicit **Yes** would.

   The reasoning step that follows is **one combined pass** — never five separate dispatches —
   answering all five angles per the table in **Run 2 — the branch is merged**
   (`skills/myflow-contracts/finish-contract.md`), step 9, plus the rating request, fed the script's
   bundle and the live session's own context. Angle 5's remit covers the records the pipeline writes
   to files today and the derivation work now done in Bash or by the agent — **not** what the SPA
   should display. See **Run 2 — archive and clean up**
   (`skills/myflow-finish/SKILL-rationale.md`) for why this runs as one combined pass.

   Every finding is explained in the message body first — what was observed, what breaks, and what
   the fix would be — before any prompt fires. The filing ask is then **one multi-select prompt per
   angle**, listing that angle's findings and defaulting to filing none, shape per Operator prompts
   (`skills/myflow-contracts/operator-prompts.md`):

   > **File any of this angle's findings as Jira issues?**
   > - **<finding 1>**
   > - **<finding 2>**
   > - **None — file nothing** *(default, recommended)*

   An angle with no findings gets no filing ask, and a bare observation with no concrete change
   implied does not appear as an option. Labelling a filed issue and handling a filing failure
   follow **Labels on issues the pipeline creates** and **Never blocking**
   (`skills/myflow-contracts/jira-integration.md`) verbatim.

   Then ask the operator to rate the run:

   **Rate this myflow run, 1 (rough) to 5 (excellent):**

   Write `<project>/docs/self-review/<name>-self-review.md` — one section per angle, all five present; each
   finding one line naming its angle's label, the finding, and its disposition, the issue key when
   filed or an explicit declined marker when not; an angle with no findings carrying an explicit
   none-marker instead — plus the rating — and commit it on `chore/archive-<name>` **in the main
   checkout**, never the removed worktree, asserting the branch rather than assuming it and **not
   pushing here** — step 10 below pushes it in the same PR as the archive commit — as one guarded
   commit mirroring the shape **Git boundaries** (`skills/myflow-contracts/pipeline.md`) already
   documents:

   ```bash
   [ "$(git -C <main-checkout> branch --show-current)" = "chore/archive-<name>" ] \
     && git -C <main-checkout> add -- docs/self-review/<name>-self-review.md \
     && { git -C <main-checkout> diff --cached --quiet \
          || git -C <main-checkout> commit -m "docs(self-review): <name> self-review report"; }
   ```

   The subject is the fixed literal shown, scope included, per **Finish run 2's two commits carry
   module scopes** (`<agents repo>/openspec/specs/myflow-commit-scope/spec.md`).

   A branch mismatch reports the branch found and stops this commit, and so does a commit that FAILS
   (hook rejection) — reported with git's own output either way. The change stays `FINISHED`
   regardless — a report that failed to commit is a self-review failure to report in the handoff,
   never a reason to reopen the change.

```bash
myflow stage end -command '/myflow-finish' -stage finish.self-review -outcome completed <name>
```

The change is `FINISHED` before this mark closes, and a failure inside self-review — a declined
run, a failed filing, a failed commit — is a self-review failure to report, never a reason this mark
records anything but `completed`: self-review's own outcome is carried in the handoff text, not in
this mark's `-outcome` value. See **Stage marks** (`skills/myflow-contracts/pipeline.md`) — a failed
*mark* is a different thing from a failed self-review, and only the former ever changes what this
call records.

```bash
myflow stage begin -command '/myflow-finish' -stage finish.push-archive -harness <harness> -session-token mf-<literal-token> <name>
```

10. **Push the archive branch and open its pull request.** Push `chore/archive-<name>`; open a pull request against `<base>` via a PR CLI when
    usable for the host, else print the forge's create-PR URL and ask whether it was opened — the
    same shape **1.3 Take the chosen route** above already uses. This push carries both the archive
    commit from step 4 and the self-review report from step 9, so there is no window in which the
    archive pull request merges while the report is still unwritten. Run 2 never pushes to `<base>`.

    A failed push, or a failed pull-request creation, is reported with the command's own output. It
    never moves the change off `FINISHED` — the change is already terminal by step 8 — and the
    handoff names the unpushed branch `chore/archive-<name>` and the exact push and pull-request
    commands needed to land it by hand. The archive is never reported as landed until this step
    actually lands it.

11. **Restore the main checkout to `<base>`.** Successful or not — a failed push does not skip
    this — never leave the checkout on the archive branch. This step runs inside step 10's
    `finish.push-archive` mark rather than opening one of its own.

```bash
myflow stage end -command '/myflow-finish' -stage finish.push-archive -outcome completed <name>
```

```
## Finished

**Change:** <name>
**Specs:** synced | skipped | none
**Archived:** openspec/changes/archive/<date>-<name>/ (committed on chore/archive-<name>)
**Archive PR:** <prUrl> | not pushed — <reason>; land it with: git -C <main-checkout> push -u origin chore/archive-<name>, then open a PR against <base>
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Cleanup:** verified
**Self-review:** <path> (rating: <n>/5) | skipped
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
```

On a failed push or PR
creation, the `Archive PR` line is the local branch and the exact commands to land it by hand — never
a claim that the archive merged. A run 2 that **completes step 10** is terminal and names **no** next
command.

On a leftover — or on no verdict at all — the run stops at step 7 instead, and its handoff names
what remains and points back at itself, because clearing the leftover and re-running is what
finishes the change:

```
## Cleanup incomplete — not finished

**Change:** <name>
**Specs:** synced | skipped | none
**Archived:** openspec/changes/archive/<date>-<name>/ (committed on chore/archive-<name>)
**Remaining:** <what the guard named> | unverified — <what the guard reported on stderr>
**State:** IN_PROGRESS — FINISHED is not written while anything remains

<what the operator must clear>

Next:
/myflow-finish <name>
```

## Guardrails

- **Never** ask how the branch should land before the unfinished-work gate has been answered, and
  never run a git command before it either.
- **Never** mix the implementation and the planning artifacts in one commit, and never leave the
  planning commit's message silent about work the operator chose to integrate over.
- **Never** archive a change whose branch has not reached the base branch.
- **Never** decide run 1 versus run 2 from the ancestor test alone, and never from a commit count —
  both answer wrongly on a branch that was never committed.
- **Never** run tests, linters, or a coverage check.
- **Never** hardcode `main` or `develop`, and **never** resolve the base branch from `HEAD`'s
  upstream — this skill runs on the change's own branch, so that compares it against itself.
- **Never** restate the finish procedure here; `pipeline.md` is canonical and this skill points
  at it.
- **Never** let a git failure pass silently — report its output and stop at `IN_PROGRESS`. A commit
  **skipped** because nothing was staged is not a failure; say it was skipped and carry on.
- **Never** stage past a symlinked planning path with a bare `git add -A` — that is the one
  workaround that puts the plan into the implementation commit.
- **Never** merge the change branch in run 2; run 2's step 1, "Verify the merge", already proved it.
- **Never** state a cleanup rule here. Which artifact is removed, when, and on what condition is
  **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`); how it is removed is
  **Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`). A second copy is the one
  that goes stale, and a stale copy of a removal rule deletes the wrong thing.
- **Never** report a cleanup as done without the verdict that says so, and **never write `FINISHED`
  over a leftover or an unverified cleanup** — `FINISHED` is terminal, so the console line would be
  the only record of it, which is exactly the transcript-only record this pipeline refuses.
- **Never** `git add` the state file, and never move it into the archive.
- **Never** let a Jira call block the archive — one skipped-with-reason line.
- **Never** let self-review block, delay, or undo the `FINISHED` write — it runs only after that
  write succeeds, and a failure or a skip inside it never moves the change off `FINISHED`.
- **Never** ask the self-review skip prompt, a per-angle filing ask, or the rating question
  before `FINISHED` has been written.
- **No flags.** The only argument is the optional change name; report anything else.
