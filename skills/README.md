# Gymie Cursor skills — myflow

**myflow** = OpenSpec + Superpowers **Basic Workflow** bridge with **manual review** and **manual test** gates (each with an inline fix loop) between apply and code review.

```text
start → do (#2–#6) → manual review (Gate B, optional do-fix×N) → manual test (Gate C, optional do-fix×N) → code review (commit + push + open PR) → PR review (Gate D, human-merged) → finish (archive)
```

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
| 7 | finishing-a-development-branch | `/myflow-code-review` (commit + push + open PR, never merges) |

| Command | Skill | Stage |
|---------|-------|-------|
| `/myflow-start <name>` | `openspec-propose-superpowers` | #1 + OpenSpec artifacts + #3 |
| `/myflow-do <name>` | `openspec-apply-superpowers` | #2–#6 (stage; no commits) |
| *(manual review)* | User (Gate B) | Inspect **staged** diff in worktree IDE |
| `/myflow-do-fix <name>` | `openspec-apply-fix-superpowers` | Gate B/C/D finding → document in proposal (append or nested) → #4–#6 in existing worktree (stage at Gate B/C; commit+push at Gate D PR-fix mode) |
| `/myflow-manual-test <name>` | `openspec-manual-test-superpowers` | Gate C — run guide + functionality checklist MD; always asks whether to skip (default No) |
| `/myflow-code-review <name>` | `openspec-code-review-superpowers` | Verify Gate C → test coverage check → tests/linters → commit + push → #7 open PR (never merges) |
| *(PR review)* | User (Gate D) | Review and merge the PR on the forge — nothing in myflow merges |
| `/myflow-finish <name>` | `openspec-archive-superpowers` | Verify the PR merged → delta sync → archive |
| `/myflow-full <name>` | `openspec-full-cycle-superpowers` | Full cycle through Gate D (PR open, stop); `/myflow-finish` is a separate human-initiated step |
| `/myflow-status <name>` | — (read-only) | Stage report for open changes |
| `/myflow-info` | — (read-only) | Reads the rule file and explains the pipeline |

### Typical flow

```text
/myflow-start add-my-feature
# #1 brainstorm → OpenSpec proposal/specs/design → #3 writing-plans → tasks.md

/myflow-do add-my-feature
# #2 worktree → #4 SDD → #5 TDD → #6 review panel (primary+Bugbot+Security+Adversarial+Senior+Economic Senior) → git add -A (staged; no #7)

# Gate B — open worktree in IDE; review staged changes
#   cd <worktree> && git diff --cached

# Found something to fix? /myflow-do-fix add-my-feature
# documents the fix in proposal.md/tasks.md (or a linked <name>-fix-N sub-change),
# resumes the SAME worktree, #4-#6 again (full panel re-run), stage; no commits
# — loop as many rounds as needed, then continue

/myflow-manual-test add-my-feature
# writes docs/manual-test/<name>.md (how to run apps + checklist); link only + stop

# Gate C — follow the guide; check off items in the MD
# Found something to fix? /myflow-do-fix add-my-feature, then refresh the guide
# and re-test — loop as many rounds as needed

/myflow-code-review add-my-feature
# verify Gate C checklist (or SKIPPED marker) → test coverage check (routes gaps to
# /myflow-do-fix) → tests/linters → commit apply work → #7 push + open PR (never merges)

# Gate D — review the PR on the forge and merge it (human step; nothing in myflow merges)

/myflow-finish add-my-feature
# verify the PR actually merged → OpenSpec delta sync → archive
```

One-shot with approval gates:

```text
/myflow-full add-my-feature
# start → do (#2–#6) → Gate B review (+ do-fix loop) → Gate C manual test (+ do-fix loop)
# → code review (commit + push + open PR) → stop at Gate D (PR open)
# /myflow-finish is always a separate, human-initiated step after the PR merges
```

**Full cycle flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test` (pre-answers the Gate C skip prompt with Yes and must announce it), `commit-during-apply` (legacy)

**Resume:** `/myflow-do <name>` (partial) · `/myflow-do-fix <name>` (Gate B/C/D finding) · `/myflow-manual-test <name>` (Gate C) · `/myflow-code-review <name>` (after both gates) · `/myflow-finish <name>` (after the PR merges) · `/myflow-status <name>` (find where a change is) · `/myflow-full <name> skip-propose` (do → review → test → code review → Gate D)

## OpenSpec only (lighter loop)

Invoke by skill name — no myflow alias; **does not enforce Basic Workflow #1–#7 or manual review/test gates**:

- `openspec-propose`, `openspec-apply-change`, `openspec-archive-change`
- `openspec-explore`, `openspec-update-change`, `openspec-sync-specs`
