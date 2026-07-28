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

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, git
boundaries, the finish contract, and the handoff output shape.

## State gate

Accepts **`IN_PROGRESS`**. Run 1 ends at `IN_PROGRESS`; run 2 ends at `FINISHED`.

At `STARTED` there is nothing to integrate — emit the wrong-state handoff and recommend
`/myflow-do`. At `FINISHED` the change is already archived; say so and stop.

## Deciding which run this is

**`pipeline.md`'s Finish contract is canonical for every procedure below.** Load it and follow it
there — the base-branch resolution, the four preflight checks, the removal sequence and their
rationales live in that one file. This skill carries only what is specific to *executing* it, and
deliberately does not restate it: a second copy of a procedure that deletes worktrees is a second
copy that can drift.

Which run happens is decided by one thing: whether the change's branch has already reached the
base branch. No field records "integration started" — a field could disagree with git.

- **not merged** → run 1 (integrate)
- **merged** → run 2 (archive and clean up)

A PR a human merged on the forge, a colleague's merge, and run 1's own merge are indistinguishable
here, and that is correct — all three mean the same thing.

**The base branch is resolved, never assumed, and never derived from the current branch.** Follow
the resolution in **Finish contract**; if it does not resolve to a branch distinct from the one
checked out, **stop and ask**. This skill runs inside the apply worktree, where `HEAD` is the
change's own branch — so any resolution that consults `HEAD`'s upstream would compare the branch
against itself and report every pushed branch as merged.

---

# Run 1 — integrate

## 1.1 Ask how the branch should land, before any git action

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Having asked once, run to completion without asking again. **Never** remember the answer between
runs, and never infer it from anything else.

**Report an existing PR before asking.** Run 1 is re-entered whenever the branch is not merged, so
a previous attempt may already have opened one. If a PR exists for this branch — from `prUrl`, or
from a PR CLI when one is usable — say so, including whether it is open or closed-unmerged, before
the operator answers. Re-asking cold invites a duplicate PR.

## 1.2 Commit the staged work

All three routes commit first — implementation, `docs/manual-test/<name>.md`, and the `openspec/`
planning artifacts. The planning artifacts were hidden from the review diff, not from the commit.

Run `git add -A` first rather than assuming everything is already staged: the operator may have
edited the worktree at the human gate without staging.

The state file is **not** committed — it lives outside the repo.

## 1.3 Take the chosen route

Per **Finish contract** → run 1. Push with `-u` so the branch has an upstream; the unpushed-commits
check in run 2 treats a missing upstream as unknown and refuses to clean up without one.

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

## 1.4 No verification gate

**Run no tests, no linters, and no spec-coverage check** — see **Finish contract**. Correctness was
established during `/myflow-do` and by the human gate.

## 1.5 State and handoff

Write the state file with `state` unchanged at `IN_PROGRESS`, `prUrl` set if a PR was opened, and
every other field carried forward.

**Transition the issue to In Review** once a PR is confirmed open, per **Jira integration**
(`skills/myflow-contracts/jira-integration.md`) — after the state write, never before, and never
blocking.

```
## Branch integrated — waiting on the merge

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it

<what the operator must do before the next run>

Next:
/myflow-finish <name>
```

The last line is this same command, because that is what the operator runs once the branch is
merged.

---

# Run 2 — archive and clean up

Follow **Finish contract** → run 2 for the full procedure. In outline, and stopping at the first
step that fails:

1. **Verify the merge** — a PR CLI when usable, otherwise `git merge-base --is-ancestor`, which
   must stay reachable on its own as the only evidence on a non-GitHub forge. Fetch first so the
   tracking ref is current. Not merged → this is not run 2; fall back to run 1 and **archive
   nothing**.
2. **Sync delta specs, then archive.** Assess each delta in `<changeRoot>/specs/` against
   `openspec/specs/`, show a summary, and offer: sync now (recommended), archive without syncing,
   or cancel. Apply `## ADDED` by appending (creating the capability spec if absent), `## MODIFIED`
   by replacing the block matched on its `### Requirement:` heading whitespace-insensitively,
   `## REMOVED` by deleting it, `## RENAMED` in place preserving the body. Then move the change to
   `openspec/changes/archive/<YYYY-MM-DD>-<name>/`, taking any nested `<name>-fix-N` with it.
3. **Commit and push the archive** on the base branch in the main checkout.
4. **Remove the worktrees** — the four gating checks, the ignored-file disclosure, and the removal
   sequence are all in **Finish contract**. The disclosure is not optional: `--force` destroys
   every ignored file, so show the list and get confirmation. Verify each removal succeeded.
5. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — a failed removal keeps its entry, so it stays findable. Carry `artifactUrl`, `jiraIssue` and `prUrl` forward. The state file stays at its
   user-scoped path as the terminal record — it is **never** moved into the archive.

**Transition the issue to Done** after the state write, per **Jira integration**.

```
## Finished

**Change:** <name>
**Specs:** synced | skipped | none
**Archived:** openspec/changes/archive/<date>-<name>/ (committed and pushed)
**Worktrees:** removed | left alone — <reason>
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
```

Run 2 is terminal and names **no** next command.

## Guardrails

- **Never** archive a change whose branch has not reached the base branch.
- **Never** run tests, linters, or a coverage check.
- **Never** hardcode `main` or `develop`, and **never** resolve the base branch from `HEAD`'s
  upstream — this skill runs on the change's own branch, so that compares it against itself.
- **Never** restate the finish procedure here; `pipeline.md` is canonical and this skill points
  at it.
- **Never** let a git failure pass silently — report its output and stop at `IN_PROGRESS`.
- **Never** merge the change branch in run 2; step 2.1 already proved it merged.
- **Never** use `git branch -D`.
- **Never** remove any worktree if any check failed for any worktree.
- **Never** `git add` the state file, and never move it into the archive.
- **Never** let a Jira call block the archive — one skipped-with-reason line.
- **No flags.** The only argument is the optional change name; report anything else.
