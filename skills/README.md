# Gymie Cursor skills — myflow

**myflow** = OpenSpec + Superpowers **Basic Workflow** bridge with a **twelve-stage** state
machine: a proposal gate, **manual review** and **manual test** gates (each with an inline fix
loop) between apply and review, and a PR review gate before finish.

```text
awaiting-proposal-review (Gate A) → proposal-done → awaiting-do-review (Gate B) → do-review-started → do-done →
[awaiting-fix-review → fix-review-started] → awaiting-manual-test (Gate C) → manual-test-done →
awaiting-pr-review (Gate D) → review-done → finished
```

**`*-done` and `*-manual-review` commands are pure state writes** — they call
`myflow-state-advance` to update `stage`/`updatedAt`/`updatedBy` only, with no verification, no
artifact reading, and no git. They record a human confirmation as a discrete fact, separate from
`/myflow-finish` independently verifying the PR merged.

See also: `.cursor/rules/myflow-manual-review.mdc`

`<name>` is **optional** on every `/myflow-*` command below — if omitted, the sole active (non-archived) change relevant to that stage is used automatically; if there are multiple, you're asked which.

**Model:** `/myflow-start` → Opus (enforced via frontmatter in Claude Code; manual switch elsewhere). Every other command → Sonnet. See "Model policy" in `myflow-manual-review.mdc`.

## Superpowers Basic Workflow map

| Step | Skill | myflow stage |
|------|-------|--------------|
| 1 | brainstorming | `/myflow-start` |
| 2 | using-git-worktrees | `/myflow-do` |
| 3 | writing-plans | `/myflow-start` (+ validated at do) |
| 4 | subagent-driven-development | `/myflow-do` (+ `/myflow-do-fix` rounds) |
| 5 | test-driven-development | `/myflow-do` (+ `/myflow-do-fix` rounds, every task) |
| 6 | requesting-code-review + strict panel (Bugbot, Security, Adversarial, Senior, Economic Senior) | `/myflow-do` (per-task + final **1+5 agents**; full re-run on every `/myflow-do-fix` round) |
| 7 | finishing-a-development-branch | `/myflow-review` (commit + push + open PR, never merges) |

| Command | Skill | Stage |
|---------|-------|-------|
| `/myflow-start <name>` | `openspec-propose-superpowers` | #1 + OpenSpec artifacts + #3 → publishes proposal artifact → `awaiting-proposal-review` |
| *(Gate A)* | User | Read the proposal artifact |
| `/myflow-start-fix <name>` | `openspec-propose-fix-superpowers` | Revise proposal, republish artifact to same URL, stay at `awaiting-proposal-review` |
| `/myflow-start-done <name>` | `myflow-state-advance` | *Pure state write* → `proposal-done` |
| `/myflow-do <name>` | `openspec-apply-superpowers` | #2–#6 (stage; no commits) → `awaiting-do-review` |
| *(Gate B)* | User | Inspect **staged** diff in worktree IDE |
| `/myflow-do-manual-review <name>` | `myflow-state-advance` | *Pure state write* → `do-review-started` |
| `/myflow-do-done <name>` | `myflow-state-advance` | *Pure state write* → `do-done` |
| `/myflow-do-fix <name>` | `openspec-apply-fix-superpowers` | Gate B/C/D finding → document in proposal (append or nested) → #4–#6 in existing worktree (stage at Gate B/C; commit+push at Gate D PR-fix mode) → records `originStage`, sets `awaiting-fix-review` |
| `/myflow-do-fix-manual-review <name>` | `myflow-state-advance` | *Pure state write* → `fix-review-started` |
| `/myflow-do-fix-done <name>` | `myflow-state-advance` | *Pure state write* → returns to `originStage`, clears it |
| `/myflow-manual-test <name>` | `openspec-manual-test-superpowers` | Gate C — run guide + functionality checklist MD; always asks whether to skip (default No) → `awaiting-manual-test` |
| *(Gate C)* | User | Run the apps, check off items in the guide |
| `/myflow-manual-test-done <name>` | `myflow-state-advance` | *Pure state write* → `manual-test-done` |
| `/myflow-review <name>` | `openspec-review-superpowers` | Verify Gate C → test coverage check → tests/linters → commit + push → #7 open PR (never merges unless `automerge`) → `awaiting-pr-review` (or `review-done` with `automerge`) |
| *(Gate D)* | User | Review and merge the PR on the forge — skipped when `automerge` was used |
| `/myflow-review-done <name>` | `myflow-state-advance` | *Pure state write* → `review-done` |
| `/myflow-finish <name>` | `openspec-archive-superpowers` | Verify the PR merged → delta sync → archive |
| `/myflow-full <name>` | `openspec-full-cycle-superpowers` | Full cycle through Gate D (PR open, stop) — or `review-done` with `automerge`; `/myflow-finish` is a separate human-initiated step |
| `/myflow-status <name>` | — (read-only) | Stage report for open changes |
| `/myflow-info` | — (read-only) | Reads the rule file and explains the pipeline |

### Typical flow

```text
/myflow-start add-my-feature
# #1 brainstorm → OpenSpec proposal/specs/design → #3 writing-plans → tasks.md
# publishes proposal artifact → stage: awaiting-proposal-review

# Gate A — read the proposal artifact
# Want changes first? /myflow-start-fix add-my-feature — revises + republishes to the same URL

/myflow-start-done add-my-feature
# pure state write — confirms the proposal was reviewed → stage: proposal-done

/myflow-do add-my-feature
# #2 worktree → #4 SDD → #5 TDD → #6 review panel (primary+Bugbot+Security+Adversarial+Senior+Economic Senior) → git add -A (staged; no #7)
# stage: awaiting-do-review

# Gate B — open worktree in IDE; review staged changes
#   cd <worktree> && git diff --cached
# /myflow-do-manual-review add-my-feature — pure state write, stage: do-review-started (optional, marks review in progress)

# Found something to fix? /myflow-do-fix add-my-feature
# documents the fix in proposal.md/tasks.md (or a linked <name>-fix-N sub-change),
# resumes the SAME worktree, #4-#6 again (full panel re-run), stage; no commits
# records originStage, stage: awaiting-fix-review
# /myflow-do-fix-manual-review add-my-feature — pure state write, stage: fix-review-started (optional)
# /myflow-do-fix-done add-my-feature — pure state write, returns to originStage, clears it
# — loop as many rounds as needed, then continue

/myflow-do-done add-my-feature
# pure state write — confirms the implementation diff was reviewed → stage: do-done

/myflow-manual-test add-my-feature
# writes docs/manual-test/<name>.md (how to run apps + checklist); link only + stop
# stage: awaiting-manual-test

# Gate C — follow the guide; check off items in the MD
# Found something to fix? /myflow-do-fix add-my-feature, then refresh the guide
# and re-test — loop as many rounds as needed

/myflow-manual-test-done add-my-feature
# pure state write — confirms manual testing is complete → stage: manual-test-done

/myflow-review add-my-feature
# verify Gate C checklist (or SKIPPED marker) → test coverage check (routes gaps to
# /myflow-do-fix) → tests/linters → commit apply work → #7 push + open PR (never merges unless automerge)
# stage: awaiting-pr-review (or, with automerge, commits+pushes+merges → stage: review-done, no PR)

# Gate D — review the PR on the forge and merge it (human step; nothing in myflow merges
# unless automerge was used, in which case this gate was already skipped)

/myflow-review-done add-my-feature
# pure state write — confirms the PR was reviewed (and merged) → stage: review-done

/myflow-finish add-my-feature
# verify the PR actually merged → OpenSpec delta sync → archive
```

One-shot with approval gates:

```text
/myflow-full add-my-feature
# start → do (#2–#6) → Gate B review (+ do-fix loop) → Gate C manual test (+ do-fix loop)
# → review (commit + push + open PR) → stop at Gate D (PR open) — or, with automerge, ends at
# stage: review-done (no PR to stop at)
# never auto-invokes any *-done/*-manual-review command — those stay separate human confirmations
# /myflow-finish is always a separate, human-initiated step after the PR merges
```

**Full cycle flags:** `skip-propose`, `propose-only`, `skip-review` (skips Gate B only; typing the flag at invocation is the human's explicit opt-out, so the cycle writes stage: do-done with gates.reviewed: false rather than self-certifying a review nobody did), `skip-manual-test` (pre-answers the Gate C skip prompt with Yes, writing stage: manual-test-done with gates.tested: "skipped" for the same reason; review still runs and still checks coverage), `automerge` (opt-in only, on `/myflow-review`/`/myflow-full` — commits, pushes, and merges, skipping Gate D and ending at review-done; never implied by any other flag), `commit-during-apply` (legacy)

**Resume:** `/myflow-do <name>` (partial) · `/myflow-do-fix <name>` (Gate B/C/D finding) · `/myflow-manual-test <name>` (Gate C) · `/myflow-review <name>` (after both gates) · `/myflow-finish <name>` (after the PR merges) · `/myflow-status <name>` (find where a change is) · `/myflow-full <name> skip-propose` (do → review → test → review → Gate D)

## OpenSpec only (lighter loop)

Invoke by skill name — no myflow alias; **does not enforce Basic Workflow #1–#7 or manual review/test gates**:

- `openspec-propose`, `openspec-apply-change`, `openspec-archive-change`
- `openspec-explore`, `openspec-update-change`, `openspec-sync-specs`
