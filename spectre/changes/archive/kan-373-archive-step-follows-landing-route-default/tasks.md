# kan-373-archive-step-follows-landing-route-default

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Three tasks. Task 1 defines the new project-configuration key and sets this repo's own value —
nothing reads it yet, so it lands first and inertly. Task 2 makes the landing question read it.
Task 3 makes the archive step mirror the parent route; it depends on nothing task 2 touches, but
follows it so a reader sees "the default is chosen" before "the archive step follows whatever was
chosen."

- [x] 1. Add the `## default landing route` project-configuration key, and set it for this repo

  In `skills/flow-contracts/project-configuration.md`, add one row to the `| Key | Supplies |`
  table (immediately after the `## jira` row), matching that row's own style — "Optional." prefix,
  what it supplies, where it's read:

  ```markdown verified:authored in-tree for this change
  | `## default landing route` | Optional. One of the literal bodies `pull request`, `merge and push` or `manual` — this section holds that value and nothing else, never free-form prose, matching `## jira`'s own single-line-literal shape. Used as `skills/flow/integrate.md`'s landing question's own `(default, recommended)` option; absent, or a body matching none of the three literals exactly, is reported by name and dropped, falling back to `pull request` as today. |
  ```

  Then add a short subsection (after the `## jira` row's own prose, before `## Where the agents
  repository is`) spelling out the exact match rule, mirroring `## jira`'s "must match... in its
  entirety" phrasing:

  ```markdown verified:authored in-tree for this change
  **`## default landing route`'s body is matched byte-for-byte against the three literals above,
  leading/trailing whitespace trimmed and nothing else normalized** — no case-folding, no synonym
  list. A body that does not match exactly one of them is a malformed row: report it by name
  (quoting what was found) and drop it, resolving as if the key were absent.
  ```

  **In this repository's own `<agents repo>/.flow/project.md`**, add the new section right after
  the existing `## jira` section (before `## workspace isolation`):

  ```markdown verified:authored in-tree for this change
  ## default landing route

  `merge and push`
  ```

  Verify: `scripts/check-plan-shape.sh spectre/changes/kan-373-archive-step-follows-landing-route-default/tasks.md`
  exits 0 (checked once this task's own commit lands); `scripts/check-references.sh` exits 0 — the
  new key row must resolve under that guard's citation rules exactly like the `## jira` row beside
  it. Then run the project's `## lint` list and confirm it is clean.

**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`, `.flow/project.md`
**Tests:** none — this task only adds a documented key and this repo's own value for it; nothing
reads either yet (task 2 is the first reader).
**Regression:** n/a — no code path exercises this key before task 2 lands.
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow-contracts): add the default landing route project-configuration key`

- [x] 2. Resolve the project default in `integrate.md`'s landing question

  In `skills/flow/integrate.md`'s `## 2. Ask how the branch should land`, immediately before the
  existing `> **How should this branch land?**` prompt block, add a short paragraph and resolution
  step:

  > **Load `skills/flow-contracts/project-configuration.md`** — the `## default landing route` key
  > is canonical there. Read `<project>/.flow/project.md`'s `## default landing route` section, if
  > present, and resolve it against the three literals `pull request` / `merge and push` /
  > `manual`, byte-for-byte after trimming leading/trailing whitespace. A body matching none of
  > them exactly is reported by name and dropped, resolving as absent. Absent → `pull request`
  > stays the recommended default, exactly as today.

  Then change the prompt block itself so the recommended marker moves to whichever option the
  resolution above named, e.g. for a resolved `merge and push`:

  ```markdown verified:authored in-tree for this change
  > **How should this branch land?**
  > - **Open a pull request**
  > - **Merge and push** *(default, recommended)*
  > - **Handle it manually**
  ```

  Word it so the marker's placement is described generically (it moves to match the resolved
  option; `pull request` keeps it when nothing resolved), not hardcoded to one option in the prose.

  Leave everything else in this section unchanged — the base-branch-moved check above it, the
  "having asked once, run to completion" sentence below it, and the existing-PR report all stay
  exactly as they are; only which option carries `(default, recommended)` changes, and the question
  is still always asked.

  Verify: `grep -n "default landing route" skills/flow/integrate.md` finds the new resolution step;
  `scripts/check-references.sh` and `scripts/check-plan-shape.sh` both exit 0. Then run the
  project's `## lint` list and confirm it is clean.

**Build:** green
**Files:** `skills/flow/integrate.md`
**Tests:** none — prose-only change to a prompt's default option; task 1's key is inert until this
task reads it, and there is no automated harness for an interactive prompt's wording.
**Regression:** reverting this task returns the landing question to always recommending "Open a
pull request" regardless of `.flow/project.md`'s declared default — a silent behavior loss with no
test to catch it, since none exists for this prose.
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow): read the project's default landing route in the landing question`

- [x] 3. Archive mirrors the merge-and-push route, same-invocation only

  **In `skills/flow/archive.md`**, step 10 currently reads:

  > 10. **Push the archive branch and open its pull request.** Push `chore/archive-<name>`; open a
  > pull request against `<base>` via a PR CLI when usable, else print the forge's create-PR URL
  > and ask whether it was opened. This push carries both the archive commit and the self-review
  > report, so there is no window in which the archive PR merges while the report is still
  > unwritten. Never pushes to `<base>`.
  >
  > A failed push, or a failed PR creation, is reported and never moves the change off `FINISHED`.

  Replace it with a route split, keyed on whether this run of `archive.md` was reached as the
  direct, same-invocation continuation of `skills/flow/integrate.md`'s merge-and-push route (see
  that file's own "After merge-and-push specifically" section) or as a standalone invocation:

  ```markdown verified:authored in-tree for this change
  10. **Push the archive branch and land it**, per **Run 2** (`skills/flow-contracts/finish-contract-run2.md`)'s
      own route table for this step:

      | Reached via | Then |
      |---|---|
      | the merge-and-push continuation, same invocation as `skills/flow/integrate.md` | push `chore/archive-<name>`; merge it into `<base>`; push `<base>` — no PR |
      | a standalone invocation (the PR or manual routes always defer archiving this way, and so does any resumed run) | push `chore/archive-<name>`; open a pull request against `<base>` via a PR CLI when usable, else print the forge's create-PR URL and ask whether it was opened |

      Either way this push carries both the archive commit and the self-review report, so there is
      no window in which the archive lands while the report is still unwritten. Never pushes
      anything but `chore/archive-<name>` and, on the merge-and-push row, `<base>` itself.

      A failed push, a failed merge, or a failed PR creation is reported and never moves the change
      off `FINISHED`.
  ```

  **In `skills/flow-contracts/finish-contract-run2.md`**, step 10 currently reads:

  > 10. **Push the archive branch and open its pull request.** Push `chore/archive-<name>`; open a
  > pull request against `<base>` via a PR CLI when usable for the host, else print the forge's
  > create-PR URL and ask whether it was opened — the same shape Run 1's pull-request route
  > (`skills/flow-contracts/finish-contract-run1.md`) uses. This push carries both the archive
  > commit and the self-review report from step 9, so there is no window in which the archive pull
  > request merges while the report is still unwritten. Run 2 never pushes to `<base>`.
  >
  > A failed push, or a failed pull-request creation, is reported with the command's own output. It
  > never moves the change off `FINISHED` — the change is already terminal by step 8 — and the
  > handoff names the unpushed branch `chore/archive-<name>` and prints the exact push and
  > pull-request commands needed to land it by hand. The archive is never reported as landed until
  > this step actually lands it.

  Add the same two-row split as the canonical statement (this file is canonical for run 2's
  procedure; `archive.md` above cites it rather than restating it):

  ```markdown verified:authored in-tree for this change
  10. **Push the archive branch and land it — the route depends on how this run of archive.md was
      reached.**

      | Reached via | Then |
      |---|---|
      | the merge-and-push continuation, same invocation as run 1 | push `chore/archive-<name>`; merge it into `<base>`; push `<base>` — the same three sub-steps Run 1's own merge-and-push route (`skills/flow-contracts/finish-contract-run1.md`) performs, applied to the archive branch instead |
      | a standalone invocation | push `chore/archive-<name>`; open a pull request against `<base>` via a PR CLI when usable for the host, else print the forge's create-PR URL and ask whether it was opened — the same shape Run 1's pull-request route uses |

      This push carries both the archive commit and the self-review report from step 9 either way,
      so there is no window in which the archive lands while the report is still unwritten. Run 2
      never pushes anything but `chore/archive-<name>` and, on the merge-and-push row, `<base>`
      itself.

      A failed push, a failed merge, or a failed pull-request creation is reported with the
      command's own output. It never moves the change off `FINISHED` — the change is already
      terminal by step 8. On the standalone row, the handoff names the unpushed branch
      `chore/archive-<name>` and prints the exact push and pull-request commands needed to land it
      by hand. The archive is never reported as landed until this step actually lands it, on either
      row.
  ```

  **Explain why no new state is needed**, as a short paragraph directly after the split in
  `finish-contract-run2.md` (this is `design.md`'s own `no-route-field` decision, stated here for
  the reader of the contract rather than only in the change's own design doc):

  > **This split reads no persisted field.** The merge-and-push row is recognized because this run
  > of archive.md is executing as `skills/flow/integrate.md`'s own same-invocation continuation —
  > a fact already in scope for that one call path, never written to the state file. A standalone
  > invocation (the PR and manual routes always defer archiving this way) has no way to know what
  > the original route was, or whether the operator merged the PR through some mechanism this
  > pipeline never chose — so it always takes the standalone row, which is the same behavior this
  > step already had before this change.

  **In `skills/flow/archive.md`'s own handoff block**, extend the `**Archive PR:**` line to cover
  the new no-PR case:

  ```markdown verified:authored in-tree for this change
  **Archive PR:** <prUrl> | none — merged directly into <base> | not pushed — <reason>; land it with: git -C <main-checkout> push -u origin chore/archive-<name>, then open a PR against <base>
  ```

  Verify: `grep -n "merge-and-push continuation" skills/flow/archive.md skills/flow-contracts/finish-contract-run2.md`
  finds both new route splits; `scripts/check-references.sh` and `scripts/check-plan-shape.sh` both
  exit 0. Then run the project's `## lint` list and confirm it is clean.

**Build:** green
**Files:** `skills/flow/archive.md`, `skills/flow-contracts/finish-contract-run2.md`
**Tests:** none — prose-only change to two contract files describing an interactive/git procedure;
there is no automated harness for either file's own text, matching kan-288's task 2 precedent for
this same pair of files.
**Regression:** reverting this task returns the archive step to always opening a PR regardless of
how it was reached — the original bug this whole change exists to fix, with no test to catch a
silent revert since none exists for this prose.
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow): make the archive step follow the change's own merge-and-push route`
