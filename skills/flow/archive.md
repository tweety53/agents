# Archive and clean up (run 2)

Loaded either by `skills/flow/integrate.md`'s merge-and-push route, in the same invocation, or by a
fresh bare `/flow <name>` invocation once `check-finish-preflight.sh` returns `RUN2` from every
worktree.

**Load `skills/flow-contracts/artifacts-registry.md`** — every removal below is a row in it.

**`skills/flow-contracts/finish-contract-run2.md` is canonical for the full procedure.** In outline,
each numbered step below is bracketed by its own mark, with three exceptions that run inside the
mark of the step before: step 2 (positioning) inside step 3's `flow.sync-archive`; step 6 (remove
the proposal artifact source) inside step 5's `flow.cleanup`; step 11 (remove the landing
worktree) inside step 10's `flow.push-archive`. **Eleven steps, eight marks.**

**Generate this run's own session token here, before this first mark** — a separate invocation from
the integrate phase above (unless chained straight through from merge-and-push, in which case reuse
that same run's token — see **Model resolution** and the token note in `skills/flow/SKILL.md`).

```bash
flow stage begin -command '/flow' -stage flow.verify-merge -harness <harness> -session-token mf-<literal-token> <name>
```

1. **Verify the merge** — a PR CLI when usable, otherwise `git merge-base --is-ancestor`. Fetch
   first. Not merged → this is not run 2; fall back to `skills/flow/integrate.md` and **archive
   nothing** — end this mark `-outcome not-run-2` and stop.

```bash
flow stage end   -command '/flow' -stage flow.verify-merge -outcome completed <name>
flow stage begin -command '/flow' -stage flow.sync-archive -harness <harness> -session-token mf-<literal-token> <name>
```

2. **Position the landing worktree on the archive branch**, before anything else touches it.
   Resolve `<base>` with `resolve-base-branch.sh` against the apply worktree, run
   `classify-untracked.sh <project>/.worktrees/_landing-<name>` first when the worktree already
   exists — the archive pre-flight settling its untracked entries, **Run 2 — the branch is
   merged** (`skills/flow-contracts/finish-contract-run2.md`), step 2, canonical for the classes
   and for settling a reported asset through the operator — then invoke
   `prepare-archive-branch.sh <project>/.worktrees/_landing-<name> <base> chore/archive-<name>`.
   Exit `0` → `<landing-worktree>` is on `chore/archive-<name>`, cut from a fast-forwarded `<base>`;
   continue to step 3. Anything else stops run 2 here, with nothing staged, committed, pushed or
   removed. The four exit codes are **Run 2 — the branch is merged**
   (`skills/flow-contracts/finish-contract-run2.md`), step 2. The main checkout itself is never
   read, checked out, or written by this step — the guard creates `<landing-worktree>` under it when
   absent and positions it when present, per its own header.

   **When the guard is absent**, perform the same positioning by hand, in the same order, against
   `<landing-worktree>`.

3. **Archive the change** — under the mark `flow.sync-archive`. Run `spectre archive "<name>"` in
   `<landing-worktree>`. It `git mv`s `<project>/spectre/changes/<name>/` into
   `<project>/spectre/changes/archive/<name>/` and leaves the rename staged; it does not commit —
   step 4 does.

   **One call per change, parent and each `<name>-fix-N` sibling alike** — run `spectre archive
   "<name>-fix-N"` for every sub-change in this same step.

   **`spectre archive` refuses four things; `--force` overrides three of them.** Unchecked tasks, a
   missing `tasks.md`, and a `tasks.md` carrying no task at all each exit `1` and say `(use --force
   to archive anyway)`. A destination that already exists also exits `1`, and `--force` does
   **not** override that one. **`/flow` never passes `--force`.** Report the refusal, stop, and
   leave the change at `IN_PROGRESS`.

```bash
flow stage end   -command '/flow' -stage flow.sync-archive -outcome completed <name>
flow stage begin -command '/flow' -stage flow.commit-archive -harness <harness> -session-token mf-<literal-token> <name>
```

4. **Commit the archive** on `chore/archive-<name>` in `<landing-worktree>` — no push here; step 10
   carries it. The rendered ledger and panel record are preserved into this commit first — each
   when present, an absent file copying nothing; `<canonical-worktree>` resolves as run 1 resolves
   it (**Finish contract**, `skills/flow-contracts/finish-contract-run1.md`):

   ```bash
   for pair in "ledgers/<name>.md ledger.md" "reviews/<name>-panel.md panel.md"; do
     set -- $pair
     [ -f "<canonical-worktree>/.superpowers/sdd/$1" ] \
       && cp "<canonical-worktree>/.superpowers/sdd/$1" \
             "<landing-worktree>/spectre/changes/archive/<name>/$2"
   done
   [ "$(git -C <landing-worktree> branch --show-current)" = "chore/archive-<name>" ] \
     && git -C <landing-worktree> add -A \
     && check-archive-scope.sh <landing-worktree> "spectre/changes/" \
     && { git -C <landing-worktree> diff --cached --quiet \
          || git -C <landing-worktree> commit -m "chore(spectre): archive <name>"; }
   ```

   **The copy loop runs before the `add -A`, so the preserved copies ride the archive commit under
   the scope `check-archive-scope.sh` verifies.** Why the preservation exists — step 5 destroys
   the worktree the renders live in, and step 9's bundle serves the copies when the store renders
   report skipped — is step 4 of **Run 2 — the branch is merged**
   (`skills/flow-contracts/finish-contract-run2.md`), canonical for it.

   **`check-archive-scope.sh`** refuses a `git add -A` that staged more than the archive move — see
   step 4 of **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`) for
   why. A `SCOPE-VIOLATION` stops the commit exactly as a branch mismatch does. **An exit 2 with
   nothing on stdout** — a landing worktree that cannot be read — stops the commit the same way:
   the guard's header makes an inability to answer a non-verdict, so it is never read
   as `SCOPE-OK`. **When absent**, run
   `git -C <landing-worktree> diff --cached --name-only` by hand and refuse any path outside
   `<agents repo>/spectre/changes/`.

   A branch mismatch is reported, naming the branch found, and stops the commit, leaving the change
   at `IN_PROGRESS`. The subject is the fixed literal shown, per **Commit scopes name the module**
   (`<agents repo>/rules/commit-scope-is-the-module.mdc`). The self-review bundle (`flow
   self-review bundle`, step 9) resolves this commit by matching that subject line whole —
   **reproduce it exactly.**

```bash
flow stage end   -command '/flow' -stage flow.commit-archive -outcome completed <name>
flow stage begin -command '/flow' -stage flow.cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket.** The workspace half runs the `remove` row of the command table
   `project-get.sh <main-checkout> "workspace isolation"` prints (exit 1: the skipped-not-failed
   case below), from the **main checkout**, with `<id>` from `flow workspace-id <name>` — never
   handed to this run. A project declaring no `## workspace isolation` section, or no `remove`
   command, has this half **skipped, not failed**. A removal that fails is **the one exception to the
   stop-at-the-first-failure rule**: report it and carry on to step 7, which decides the verdict
   from the project's survivor report, never from this command's exit code.
6. **Remove the proposal artifact source** — always the literal `null`/absent under `/flow`, since
   `publish-proposal-removed` means no `/flow` change ever publishes one, per **Temporary artifacts
   registry** (`skills/flow-contracts/artifacts-registry.md`)'s row for it. This step is
   unconditionally a no-op skip; it still runs, and still reports the skip.

Steps 5 and 6 together are the one `flow.cleanup` stage:

```bash
flow stage end   -command '/flow' -stage flow.cleanup -outcome completed <name>
flow stage begin -command '/flow' -stage flow.verify-cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once per
   repository, after every removal above. `COMPLETE:` → report the cleanup as verified, **relay
   every clause the line carries after ` — ` word for word**, and go on to step 8 — a `SKIPPED:`
   clause there is never a pass. `LEFTOVER:` → name what remains and **stop without writing
   `FINISHED`**, leaving the change at `IN_PROGRESS`. **No verdict line at all, and a non-zero
   exit** → report it, leave the affected `worktrees` entries in the state file, and treat it as
   `LEFTOVER`.

```bash
flow stage end -command '/flow' -stage flow.verify-cleanup -outcome completed <name>
```

(`completed` on `COMPLETE:`, `leftover` on `LEFTOVER:` or on a missing verdict line — either way the
run stops here, at `IN_PROGRESS`, and nothing below runs.)

```bash
flow stage begin -command '/flow' -stage flow.write-finished -harness <harness> -session-token mf-<literal-token> <name>
```

8. **Write `FINISHED`** — reached only on `COMPLETE:` — clearing from `worktrees` **only the
   entries whose removal actually succeeded**. Carry `artifactUrl` (`null`), `jiraIssue`,
   `planningEffort` (`null`), `models.default` (`null`) and `prUrl` forward. This is the terminal
   write.

```bash
flow stage end -command '/flow' -stage flow.write-finished -outcome completed <name>
flow stage begin -command '/flow' -stage flow.self-review -harness <harness> -session-token mf-<literal-token> <name>
```

**Load `skills/flow-contracts/jira-integration.md`.** **Transition the issue to Done** after the state write, per **Jira integration**
(`skills/flow-contracts/jira-integration.md`). A run that stopped at step 7 transitions nothing.

9. **Run self-review.** The procedure — skippable per run with running it the default, fetching
   input as a served bundle rather than an inline re-read, one combined reasoning pass across all
   five angles plus the rating, the per-angle filing ask, and the report path — is **Run 2 — the
   branch is merged** (`skills/flow-contracts/finish-contract-run2.md`), step 9, canonical for it.
   What is specific to *executing* it here: `flow self-review bundle -change <name>` fetches the
   whole bundle — flowd serves the ledger and panel record from the store, falling back, per
   record, to the copies step 4 committed onto the archive branch when the store yields no render
   for it, reads the
   archived change's files out of the `chore/archive-<name>` branch of the repository the command
   resolves from its own location (the main checkout its working directory sits in), and derives
   the finish-run commits' git log — so nothing is rendered into the landing worktree first and no
   landing-worktree path is passed in or baked into the bundle.

   **Resolve `SELF_REVIEW_MODEL` here, where it is consumed** — `/flow`'s **Model resolution**
   (`skills/flow/SKILL.md`) deliberately does not, since no run that stops before archive reads it:

   ```bash
   MAIN_CHECKOUT="${MAIN_CHECKOUT:-$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)}"
   SELF_REVIEW_MODEL="$(flow settings get | jq -r '.selfReviewModel // empty')"
   PROJECT_SRM="$(project-get.sh "$MAIN_CHECKOUT" 'self review model' 2>/dev/null | tr -d '`' | xargs)"
   if [ -n "$PROJECT_SRM" ]; then
     if flow settings models | grep -qx -- "$PROJECT_SRM"; then SELF_REVIEW_MODEL="$PROJECT_SRM"
     else echo "⚠ flow: .flow/project.md '## self review model' body '$PROJECT_SRM' is not a valid model — dropped" >&2; fi
   fi
   [ -z "$SELF_REVIEW_MODEL" ] && SELF_REVIEW_MODEL=fable
   ```

   `<project>/.flow/project.md`'s `## self review model` key, when present and a valid `ValidModels`
   member, wins over the store's `selfReviewModel` field; when both are empty, or `flow settings
   get` cannot reach the store at all, `SELF_REVIEW_MODEL` falls back to the literal `fable`, named
   as a fallback rather than a resolved value exactly as `DEFAULT_MODEL`'s own `sonnet` literal is.
   On harness `zcode` the dispatch runs on `glm-5.3-flash` / `high` regardless (**Harness
   mapping**, `skills/flow-contracts/model-policy.md`).

   Run `project-get.sh <main-checkout> "self review"` (exit 1: absent) and match the body against
   the three literals `run` / `skip` / `defer` byte-for-byte after trimming leading/trailing
   whitespace; a body matching none is reported by name and dropped, resolving as absent. `skip`
   ends step 9 here, the handoff's `Self-review` line reading `skipped — project default`. `run`
   proceeds to the reasoning pass below with no prompt. `defer` proceeds straight to the bundle
   write below, with no prompt and no reasoning pass.

   When the key is absent, the skip prompt fires first:

   > **Run self-review for this change?**
   > - **Yes — run it** *(default, recommended)*
   > - **Defer — save the bundle for `/flow-self-review`**
   > - **No — skip**

   An explicit **No** stops step 9 here; the handoff's `Self-review` line reads `skipped`. A session
   with no interactive channel to present this prompt still runs self-review, exactly as an explicit
   **Yes** would.

   **On `defer`** — by key or by the prompt's third option — fetch the bundle exactly as above
   (`flow self-review bundle -change <name>`) and write its stdout, then `## Session narrative`
   (one paragraph this session writes for run 2 itself — the archived `design.md` and
   `narrative.md` are bundle sections now, and a change predating the narrative rule has the
   bundle report it skipped), to
   `<project>/docs/self-review/<name>-context.md` physically under `<landing-worktree>`; commit
   with the report's own landing script, path and subject swapped:

   ```bash
   land-self-review-report.sh "<landing-worktree>" "chore/archive-<name>" \
     "docs(self-review): <name> self-review context bundle" \
     docs/self-review/<name>-context.md
   ```

   Then straight to the `flow stage end … flow.self-review -outcome completed` mark below; no
   reasoning pass runs, and the handoff's `Self-review` line reads `deferred —
   docs/self-review/<name>-context.md`.

   **On `run` (or the skip prompt's explicit Yes), this session runs the combined reasoning pass
   itself, inline — no subagent, no dispatch.**
   `SELF_REVIEW_MODEL` still resolves, purely as a recorded value, but governs
   nothing here: there is no dispatch left to send it to. Feed the bundle's content and the
   five angles cited below directly into this session's own reasoning, then continue straight into
   the filing-and-rating prompt below — the same session already driving `AskUserQuestion`.

   The five angles and their labels are **Run 2 — the branch is merged**
   (`skills/flow-contracts/finish-contract-run2.md`), step 9, canonical for them.

   **One combined pass** — never five separate dispatches. Every finding is explained in the message
   body first, before any prompt fires. The filing ask and the rating are **one `AskUserQuestion`
   call** — one multi-select question per three findings, each option `<label>: <finding>`, the
   rating last:

   > **File any of these findings as Jira issues?**
   > - **`<label>`: <finding 1>**
   > - **`<label>`: <finding 2>**
   > - **`<label>`: <finding 3>**
   > - **None — file nothing** *(default, recommended)*
   >
   > **Rate this flow run:**
   > - **5 — excellent**
   > - **4 — good**
   > - **3 — fine**
   > - **2 — rough** — a `1` is typed through the tool's free-text "Other"

   More than nine findings roll the overflow into one further call without the rating.

   Write `<project>/docs/self-review/<name>-self-review.md`, physically under `<landing-worktree>`
   — one section per angle, all five present; each finding one line naming its angle's label, the
   finding, and its disposition;
   an angle with no findings carrying an explicit none-marker — plus the rating — and commit it on
   `chore/archive-<name>` **in `<landing-worktree>`**, not pushing here:

   ```bash
   land-self-review-report.sh "<landing-worktree>" "chore/archive-<name>" \
     "docs(self-review): <name> self-review report" \
     docs/self-review/<name>-self-review.md
   ```

   A branch mismatch or a commit that FAILS is reported and stops this commit. The change stays
   `FINISHED` regardless.

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>
flow stage begin -command '/flow' -stage flow.push-archive -harness <harness> -session-token mf-<literal-token> <name>
```

10. **Push the archive branch and land it.** The procedure — the two-row route table keyed on how
    this run of `archive.md` was reached, the guarantee that the archive commit and step 9's
    self-review report or context bundle always land together, and the failure-reporting rules —
    is **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`), step
    10, canonical for it.
11. **Remove the landing worktree.** Successful or not — never leave it behind for a later run to
    trip over:

    ```bash
    git -C <main-checkout> worktree remove --force <landing-worktree>
    git -C <main-checkout> worktree prune
    ```

    ```bash
    refresh-main-checkout.sh <main-checkout> <base>
    ```

    Runs inside step 10's mark. The refresh, and why the main checkout needs one although no step
    edited it, are step 11 of **Run 2 — the branch is merged**
    (`skills/flow-contracts/finish-contract-run2.md`); a `REFRESH-REFUSED` line goes into the
    handoff verbatim.

```bash
flow stage end -command '/flow' -stage flow.push-archive -outcome completed <name>
```

```
## Finished

**Change:** <name>
**Archived:** spectre/changes/archive/<name>/ (committed on chore/archive-<name>)
**Archive PR:** <prUrl> (merged) | none — merged directly into <base> | <prUrl> — open, not yet merged: <reason>; land it with: gh pr merge <prUrl> --merge | not pushed — <reason>; land it with: git -C <main-checkout> push -u origin chore/archive-<name>, then open and merge a PR against <base>
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Cleanup:** verified
**Self-review:** <path> (rating: <n>/5) | deferred — docs/self-review/<name>-context.md | skipped | skipped — project default
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
```

A run 2 that **completes step 10** is terminal and names **no** next command.

On a leftover — or on no verdict at all — the run stops at step 7 instead:

```
## Cleanup incomplete — not finished

**Change:** <name>
**Archived:** spectre/changes/archive/<name>/ (committed on chore/archive-<name>)
**Remaining:** <what the guard named> | unverified — <what the guard reported on stderr>
**State:** IN_PROGRESS — FINISHED is not written while anything remains

<what the operator must clear>

Next:
/flow <name>
```

## Worktree cleanup

For each worktree, run **every** check below before removing anything — the six-check sequence is
**Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`), canonical for it:

**Never ask check 4's ignored-files confirmation before removing a worktree.** Report what
`--force` will destroy — how many ignored files, which are build output, and which are irreplaceable
together with whether they were already preserved — and proceed. This is a scoped override of the
disclosure ask in **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`); it is safe
here because the records worth keeping are already out of the worktree by this point: the change's
own work is committed and step 1 proved it merged, and the rendered ledger and panel record ride
the archive commit step 4 made before any removal. **Checks 1, 2, 3, 5 and 6 remain gates.** Check
6, the live-process check, is named explicitly because it is the one this override could plausibly
be read as reaching: a live process is not a preserved record, so `HELD:` and the guard's exit 2
both stop `/flow` exactly as they stop the base contract. Check 4 turning up something genuinely
irreplaceable and *unpreserved* is not this override's case: stop and ask.

## Guardrails

- **Never** merge the change branch in run 2; step 1 already proved it.
- **Never** report a cleanup as done without the verdict that says so, and **never write
  `FINISHED` over a leftover or an unverified cleanup**.
- **Never** `git add` the state file, and never move it into the archive.
- **Never** let a Jira call block the archive — one skipped-with-reason line.
- **Never** let self-review block, delay, or undo the `FINISHED` write.
- **Never** ask the self-review skip prompt or the filing-and-rating prompt, nor resolve the
  `## self review` key, before `FINISHED` has been written.
