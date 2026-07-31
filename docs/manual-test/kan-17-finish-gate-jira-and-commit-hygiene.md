# Manual test — kan-17-finish-gate-jira-and-commit-hygiene

**Worktree (every command below runs from here):**
`/Users/tweety53/Projects/agents-worktrees/openspec-kan-17-finish-gate-jira-and-commit-hygiene`

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-17-finish-gate-jira-and-commit-hygiene
```

## There is no application to run

This repository declares no runnable application — it is the source of the myflow skills, installed
elsewhere by `setup.sh`. There is nothing to start, no port, no URL. "Running the app" here means
running the guards and their harnesses, plus a sandboxed installer run.

Two of this change's surfaces cannot be exercised without a live pipeline run: the Jira transitions
and the operator prompts. Those are checked by reading, and by using this change on the next
change you make.

## First: the automated evidence

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
```

Expected: every command exits zero. The last two are new in this change.

## Checklist

### The run-1 gate refuses to integrate over unfinished work

- [ ] **A finished change is CLEAR and a broken one is OUTSTANDING.** Point the guard at this very
      worktree — its own guide and plan are real inputs.
      ```bash
      scripts/check-unfinished-work.sh "$PWD" kan-17-finish-gate-jira-and-commit-hygiene
      ```
      Expected: one line, beginning `OUTSTANDING:` while boxes in this guide are unticked, and
      `CLEAR:` once you have ticked them all. That transition is the feature.

- [ ] **Each of the four signals fires on its own.** Untick one box here, or add a line under
      `## Known incomplete`, and re-run. The verdict names which signal fired.

- [ ] **Silence is not clearance.** Rename the guide away and re-run: the verdict is `OUTSTANDING`,
      naming the missing guide — not `CLEAR`.

- [ ] **An unreadable target refuses rather than guesses.**
      ```bash
      scripts/check-unfinished-work.sh /nonexistent demo; echo "exit=$?"
      ```
      Expected: nothing on stdout, a message on stderr, `exit=2`. A caller must not be able to read
      silence as either verdict.

### Run 2's cleanup is verified, not assumed

- [ ] **The verifier reports what is left.** Against this repository, where the change is still in
      flight, every registry row is legitimately present.
      ```bash
      scripts/check-cleanup-complete.sh "$PWD" kan-17-finish-gate-jira-and-commit-hygiene \
        /Users/tweety53/Agents/myflow/state/agents-a740d89c
      ```
      Expected: one `LEFTOVER:` line naming the worktree, the local branch and the unarchived change
      directory. That is correct here — nothing has been cleaned up yet.

- [ ] **An unreadable repository refuses.** Same shape as above: no stdout, exit 2.

- [ ] **A worktree path containing a space survives the scan.** This is the defect the change fixed;
      `pipeline.md`'s snippet and the guard now both use `substr($0,10)` rather than `$2`.

### Planning artifacts leave the reviewed diff

- [ ] **The staged diff is implementation only.** Read what `/myflow-do` now instructs, in
      `skills/myflow-do/SKILL.md` section 7: a `git reset` of the three planning paths, then a
      staging `add` that excludes them.
- [ ] **The exclusion is enforced, not asserted.** The `reset` is there because a pathspec `add`
      cannot retract prior staging — satisfy yourself that the text says why, so a later editor
      does not "simplify" it away.
- [ ] **Finish commits both, in order.** `skills/myflow-finish/SKILL.md` section 1.2: implementation
      first, planning artifacts and session records second.

### Findings became countable — by a marker line, not a parsed table

The panel record is
`/Users/tweety53/Projects/agents-worktrees/openspec-kan-17-finish-gate-jira-and-commit-hygiene/.superpowers/sdd/final-review-panel.md`.
It carries a human-readable row per finding (`ID | Slot | Severity | Location | Note`) and, beside
them, one machine-readable `finding-status: F<n> open|fixed|withdrawn <reason>` line each, plus a
`findings-total: <n>` checksum. **The marker is the only statement of a finding's state** — the
table deliberately has no status cell, because a second surface can silently disagree with the one
that governs.

- [ ] **Flipping a marker to `open` blocks the gate.** Change one `finding-status:` line to `open`
      and re-run the gate against this worktree — the verdict gains an open-findings reason. Change
      it back.
- [ ] **The checksum catches a deleted marker.** Delete one marker line and re-run: the verdict
      reports the total disagreeing with the marker count. Restore it.
- [ ] **A fenced marker does not count.** Wrap a marker in a code fence: the markers no longer form
      one unbroken block, and the gate says so. This was a real attack that worked before the fix.
- [ ] **A reused identifier is refused.** Give two rows the same `F<n>` and duplicate its marker:
      the gate names the reused identifier rather than passing on matching counts.
- [ ] **Every severity blocks.** `skills/myflow-do/SKILL.md` no longer says Critical/Important
      anywhere in its handoff bar, its fix-wave union, or its guardrails.

### Jira behaves as a projection

These need a live run to exercise; read them now and watch them on the next change.

- [ ] **In Progress fires when planning begins** — `skills/myflow-start/SKILL.md` section A, before
      the effort question, not after the state write.
- [ ] **In Review fires on every landing route** — `skills/myflow-finish/SKILL.md` section 1.5. The
      phrase "PR is confirmed open" appears nowhere under `skills/`.
- [ ] **An unrecognised status is asked about, never inferred from `statusCategory`.** This is the
      rule that unblocks `TO DO URGENT`; KAN-17 itself sat in it.
- [ ] **A created issue inherits the parent's labels plus `AI-generated`.**

### The registry states each cleanup rule once

- [ ] **`pipeline.md` has a `## Temporary artifacts registry`** with a row per artifact, including
      the remote branch, which nothing deleted before this change.
- [ ] **Nothing restates a rule.** `AGENTS.md`, `CLAUDE.md` and both `commands*/myflow-finish.md`
      name what run 2 does but no longer carry the gating checks or conditions.

### The installer still works

- [ ] **A sandboxed install completes.**
      ```bash
      SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global; echo "exit=$?"; rm -rf "$SANDBOX"
      ```
      Expected: exit 0. `setup.sh` installs skills, commands and rules; the two new scripts live in
      `scripts/` and are used from the repository, so it installs neither.

## Known incomplete

- **Two residuals in the findings gate that no document format can close.** A marker that simply
  states `fixed` for a finding that is genuinely open, or a finding removed from its row, its marker
  and the total together, both read as clear. Nothing detects a finding that was never written down.
  Recorded here because the panel's own standard is that a record living only in a transcript does
  not satisfy a durability requirement.
- **The panel record's own path is not symlink-checked.** A symlink pointing at a clean decoy is
  read as the record. Unchanged by this change, not derived from the PR-controlled change name, and
  it requires worktree write access — at which point the record can simply be rewritten. Fixing it
  belongs with the containment work, not here.
- **`## Known incomplete work` is accepted as this section** by prefix match. Pre-existing.
- **A third family of unchecked cross-references is documented but not fixed.**
  `skills/myflow-contracts/state-file.md:45` says `See **Multi-repo shape** below`, but no heading
  of that name exists — the target is a bold paragraph lead-in at `:76`. Applying this change's own
  sweep rule there makes `scripts/check-references.sh` **fail**, so closing it needs a structural
  decision (promote the paragraph to a heading, or reword the pointer) rather than a locator fix.
  The panel and a re-reviewer both confirmed the shape; it is the only real instance in the guard's
  scan set.
- **Two of this change's surfaces have no automated coverage** and cannot get any here: the Jira
  transitions (MCP-only, no CLI on this machine) and the three operator prompts. They are verified
  by reading now, and by using the pipeline on the next change.
