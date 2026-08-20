## ADDED Requirements

### Requirement: Run 2 positions the main checkout before it archives

Before the delta-spec sync and the archive move, `/myflow-finish` run 2 SHALL put the main checkout
on an archive branch named `chore/archive-<name>`, cut from a base branch that has been
fast-forwarded to `origin/<base>`. `<base>` SHALL be whatever `resolve-base-branch.sh` printed for
the apply worktree, never a guess and never derived from the currently checked-out branch.

The positioning SHALL be performed by a guard,
`prepare-archive-branch.sh <main-checkout> <base> <archive-branch>`, rather than by prose, for the
same reason `check-finish-preflight.sh` owns the run-1-versus-run-2 verdict: the decision sits in
front of an irreversible step. Its exit codes SHALL be:

| Exit | Meaning |
|------|---------|
| `0` | the main checkout is on `<archive-branch>`, cut from an up-to-date `<base>`; one line on stdout names the branch it started from and the branch it is now on |
| `1` | a named refusal — a dirty working tree, **on `<base>` or off it**, a detached `HEAD`, or an existing `<archive-branch>` that is not descended from `origin/<base>` |
| `2` | `<main-checkout>` is missing, unreadable, or not a git worktree |
| `3` | `<base>` cannot be fast-forwarded to `origin/<base>` — the local branch has diverged |

A main checkout that is already on `<base>` with a **clean** working tree SHALL be fast-forwarded and
used. A main checkout on any other branch with a **clean** working tree SHALL be switched to `<base>`
and fast-forwarded.

**A dirty working tree SHALL be refused wherever it is found — on `<base>` as well as off it.** The
refusal SHALL name both the branch found and the branch required. Refusing only the off-`<base>` case
would let uncommitted changes ride onto the archive branch unremarked, and would leave the
fast-forward to fail on its own terms rather than at a gate that says why.

An existing `<archive-branch>` that is descended from `origin/<base>` SHALL be reused rather than
refused, so that a run 2 which stopped after this step can be re-run.

Anything but exit `0` SHALL stop run 2 before the archive move, with nothing staged, committed,
pushed or removed.

**When the script is absent** — a harness whose repository does not carry it — the same checks SHALL
be performed by hand, in the same order, and the handoff SHALL say the positioning was done
manually. The check is never skipped for want of the script.

#### Scenario: The checkout is already on the base branch

- **WHEN** run 2 starts with the main checkout on `<base>` and a clean working tree
- **THEN** `<base>` is fast-forwarded to `origin/<base>`, `chore/archive-<name>` is created from it,
  and the archive move is staged there

#### Scenario: The checkout is on the base branch and dirty

- **WHEN** run 2 starts with the main checkout on `<base>` and uncommitted changes
- **THEN** run 2 stops with a refusal, exactly as it does for a dirty tree on any other branch, and
  nothing is staged, committed, pushed or removed

#### Scenario: The checkout is elsewhere and clean

- **WHEN** run 2 starts with the main checkout on an unrelated branch and a clean working tree
- **THEN** the checkout is switched to `<base>`, fast-forwarded, and put on `chore/archive-<name>`

#### Scenario: The checkout is elsewhere and dirty

- **WHEN** run 2 starts with the main checkout on an unrelated branch and uncommitted changes
- **THEN** run 2 stops with a refusal naming the branch found and `<base>`, and nothing is staged,
  committed, pushed or removed

#### Scenario: A re-run reuses the existing archive branch

- **WHEN** run 2 is re-run and `chore/archive-<name>` already exists and is descended from
  `origin/<base>`
- **THEN** it is reused rather than refused, and run 2 continues

### Requirement: Run 2 pushes the archive branch and opens its pull request

As its final step, after `FINISHED` is written and after self-review has committed its report onto
the archive branch, run 2 SHALL push `chore/archive-<name>` and open a pull request against
`<base>`. It SHALL open the pull request via a PR CLI when one is usable for the host, and otherwise
SHALL print the forge's create-PR URL and ask whether it was opened — the same shape run 1's
pull-request route already uses.

Run 2 SHALL NOT push to `<base>` at any point.

The push SHALL carry both the archive commit and the self-review report, so there is no window in
which the archive pull request is merged while the report is still unwritten.

A failed push, or a failed pull-request creation, SHALL be reported with the underlying command's
own output and SHALL NOT move the change off `FINISHED`. The handoff SHALL then name the unpushed
archive branch and print the exact commands to complete the landing by hand, and SHALL NOT report
the archive as landed.

After the push, run 2 SHALL restore the main checkout to `<base>`.

#### Scenario: The archive lands via a pull request

- **WHEN** run 2 completes successfully
- **THEN** `chore/archive-<name>` is pushed, a pull request against `<base>` is open, that branch
  carries both the archive commit and the self-review report, and nothing was pushed to `<base>`

#### Scenario: The push fails after FINISHED

- **WHEN** the archive branch cannot be pushed
- **THEN** the failure is reported with git's own output, the change stays `FINISHED`, and the
  handoff names the local archive branch and the commands needed to land it

#### Scenario: The checkout is returned to the base branch

- **WHEN** run 2 finishes, successfully or with a failed push
- **THEN** the main checkout is left on `<base>`

## MODIFIED Requirements

### Requirement: Run 2 commits and pushes the archive

After the delta-spec sync and the archive move, `/myflow-finish` SHALL commit the resulting changes
on the archive branch `chore/archive-<name>` that the positioning step put the main checkout on, and
SHALL NOT push at this point — the push is the final step of run 2, per **Requirement: Run 2 pushes
the archive branch and opens its pull request**.

There is no merge to do: the change branch was already merged, which the verification step proved.
Run 2 SHALL NOT merge anything into the base branch, and SHALL NOT commit the archive on the base
branch itself.

Every commit run 2 makes after the positioning step — the archive commit and the self-review report
alike — SHALL be made on `chore/archive-<name>`, and SHALL assert that branch rather than assume it:
naming the directory with `git -C <main-checkout>` fixes the directory and not the branch.

A finished change SHALL NOT leave the archive move uncommitted in the working tree.

#### Scenario: The archive move is committed on the archive branch

- **WHEN** `/myflow-finish` completes run 2 successfully
- **THEN** `git status` in the main checkout shows no pending archive changes, and the archive
  commit is present on `chore/archive-<name>` and on no other local branch

#### Scenario: Nested fix sub-changes are archived together

- **WHEN** a change has nested `<name>-fix-N` sub-changes that are not yet archived
- **THEN** they are archived in the same operation and included in the same commit

#### Scenario: A non-base branch is never merged into the base branch

- **WHEN** run 2 is invoked with the main checkout on an unrelated branch carrying unmerged work
- **THEN** that branch is never committed to and never merged into the base branch — the positioning
  step either switches away from it or refuses

### Requirement: Run 2 invokes self-review after writing FINISHED

Immediately after `FINISHED` is written, run 2 SHALL invoke self-review as step 9. Self-review's own
behavior — the skip prompt, context gathering, the reasoning pass, per-finding Jira filing, the
operator rating, the report and the handoff line — is defined in full by the `myflow-self-review`
capability and is not restated here. This requirement states only the invocation point: after
`FINISHED`, and before the archive branch is pushed.

Self-review SHALL NOT be able to prevent the `FINISHED` write, since it never runs before that write
succeeds, and a failure inside it SHALL NOT move the change off `FINISHED`.

**Only the step numbers and the trailing clause change here.** The positioning step shifted every
later run-2 step by one, and self-review is no longer the last thing run 2 does — the archive push
follows it, so that the report and the archive land in one pull request.

#### Scenario: Step 9 follows the FINISHED write

- **WHEN** run 2 completes step 8 and `FINISHED` is written
- **THEN** step 9 (self-review, per the `myflow-self-review` capability) runs before the archive
  branch is pushed

#### Scenario: A run 2 that never reaches FINISHED never runs self-review

- **WHEN** run 2 stops at step 7 on a cleanup leftover
- **THEN** step 9 does not run, and the change remains `IN_PROGRESS`
