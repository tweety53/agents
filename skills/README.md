# Gymie Cursor skills — myflow

**myflow** = OpenSpec + Superpowers **Basic Workflow** bridge with **manual review** and **manual test** gates (each with an inline fix loop) between apply and code review.

```text
start → do (#2–#6) → manual review (Gate B, optional do-fix×N) → manual test (Gate C, optional do-fix×N) → code review (commit + #7) → finish (archive)
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
| 7 | finishing-a-development-branch | `/myflow-code-review` (always) |

| Command | Skill | Stage |
|---------|-------|-------|
| `/myflow-start <name>` | `openspec-propose-superpowers` | #1 + OpenSpec artifacts + #3 |
| `/myflow-do <name>` | `openspec-apply-superpowers` | #2–#6 (stage; no commits) |
| *(manual review)* | User (Gate B) | Inspect **staged** diff in worktree IDE |
| `/myflow-do-fix <name>` | `openspec-apply-fix-superpowers` | Gate B/C finding → document in proposal (append or nested) → #4–#6 in existing worktree (stage; no commits) |
| `/myflow-manual-test <name>` | `openspec-manual-test-superpowers` | Gate C — run guide + functionality checklist MD |
| `/myflow-manual-test-skip <name>` | `openspec-manual-test-superpowers` (skip mode) | Gate C — same guide, marked `SKIPPED`, no boxes checked |
| `/myflow-code-review <name>` | `openspec-code-review-superpowers` | Verify Gate C → test coverage check → tests/linters → commit → #7 |
| `/myflow-finish <name>` | `openspec-archive-superpowers` | Verify merged into base branch → delta sync → archive |
| `/myflow-full <name>` | `openspec-full-cycle-superpowers` | Full cycle with gates A + B + C + D + E |

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
# /myflow-do-fix) → tests/linters → commit apply work → #7 finish branch

/myflow-finish add-my-feature
# verify the branch actually landed on main/develop → OpenSpec delta sync → archive
```

One-shot with approval gates:

```text
/myflow-full add-my-feature
# start → do (#2–#6) → Gate B review (+ do-fix loop) → Gate C manual test (+ do-fix loop)
# → code review (commit + #7) → finish (archive)
```

**Full cycle flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test`, `no-archive`, `commit-during-apply` (legacy)

**Resume:** `/myflow-do <name>` (partial) · `/myflow-do-fix <name>` (Gate B/C finding) · `/myflow-manual-test <name>` (Gate C) · `/myflow-manual-test-skip <name>` (bypass Gate C) · `/myflow-code-review <name>` (after both gates) · `/myflow-finish <name>` (after code review) · `/myflow-full <name> skip-propose` (do → review → test → code review → finish)

## OpenSpec only (lighter loop)

Invoke by skill name — no myflow alias; **does not enforce Basic Workflow #1–#7 or manual review/test gates**:

- `openspec-propose`, `openspec-apply-change`, `openspec-archive-change`
- `openspec-explore`, `openspec-update-change`, `openspec-sync-specs`
