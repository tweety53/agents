# Review panel record — kan-19-finish-safety-records-and-effort

Worktree: `/Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort`
Merge base: `a85a729ff120ba820688178e64ea3ac5eb58e607`

## Pass 1 — full roster

Mode: **full** (pass 1 always runs the whole selected roster).
Diff read by every slot: `.superpowers/sdd/final-review.diff` — `git diff --cached <merge-base> -- . ':(exclude)openspec/'`,
15 files, 970 insertions, 31 deletions.

| Slot | Lens / mandate | Model | Dispatched as |
|------|----------------|-------|---------------|
| 0 | Primary — plan alignment + code quality | `sonnet` | general-purpose, following `requesting-code-review/code-reviewer.md` |
| 1 | Defect hunt | `sonnet` | general-purpose (see substitution note) |
| 2 | Principles — **Merged** + hard invariants | `sonnet` | general-purpose, following `principles-reviewer-prompt.md` |
| 3 | Security | `sonnet` | general-purpose (see substitution note) |
| 4 | Adversarial | `sonnet` | general-purpose, following `adversarial-reviewer-prompt.md` |
| 5 | Principles — **Lens B — simplicity & state** | `sonnet` | general-purpose, same template |
| 6 | Principles — **Lens C — robustness & ops** | `sonnet` | general-purpose, same template |

No two principle reviewers share a lens. No slot was collapsed into another.

**Substitution note, slots 1 and 3.** This harness exposes no `bugbot` or `security-review`
`subagent_type` — the available agent types are `claude`, `claude-code-guide`, `Explore`,
`general-purpose`, `Plan` and `statusline-setup`. Both slots were therefore dispatched as
`general-purpose` on Sonnet carrying the equivalent mandate, rather than skipped. That is a
harness limitation, recorded here so the record is not read as a clean run of the named agents.

### Optional slot selection, evaluated against `final-review.diff` before dispatch

| Slot | Included? | Why |
|------|-----------|-----|
| 3 — Security | **yes** | The diff adds shell scripts doing path handling and file copying from caller-supplied arguments, and changes a tracked config file (`.myflow/project.md`) whose `## standards` entries are resolved and read by a review subagent. |
| 4 — Adversarial | **yes** | ~970 changed lines (>~300), two new assertion harnesses added, and behaviour changes to a contract that gates a destructive operation. |
| 5 — Lens B (simplicity & state) | **yes** | >~200 changed lines, and four new modules (two guards, two harnesses). |
| 6 — Lens C (robustness & ops) | **yes** | Error handling, exit-status protocols, a degradation path for a repository without the scripts, and the project's declared lint/test configuration. |

Nothing was excluded: every optional trigger fired.

### Resolved reviewer inputs

- `[PRINCIPLES_PATH]` = `/Users/tweety53/.claude/skills/myflow-do/engineering-principles.md` — confirmed to exist before dispatch (8,395 bytes).
- `[STANDARDS_PATHS]` — from `## standards` in `.myflow/project.md`: `CLAUDE.md` and `AGENTS.md`.
  Both are **form 2** entries (bare filenames containing no `/`), so both resolve to the project's
  own file under the project root, which is the apply worktree. Normalized, both are existing
  regular files whose parent is exactly the project root, so both pass containment. Neither entry
  was dropped; no entry failed resolution.
  - `/Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort/CLAUDE.md`
  - `/Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort/AGENTS.md`

### Deferred minors carried into the panel

Passed to slot 0 for triage, from the per-task reviews:

- Task 3 report says 29 assertions; a live run shows 28 `ok:` lines (cosmetic).
- `scripts/test-preserve-session-records.sh`'s unwritable-destination case is a no-op as root.
- Change names containing glob metacharacters are not sanitized before a `find -name` pattern
  (not reachable under openspec's naming).
- Uneven line wrapping in some `state-file.md` / `state-self-heal.md` prose.
- Pre-existing and untouched: a guardrail in `skills/myflow-finish/SKILL.md` cites "step 2.1",
  which is not a heading in that file.

### Results

**Three passes ran. Pass 3 is clean across the full roster: 0 Critical, 0 Important from every slot.**

| Slot | Pass 1 | Pass 2 | Pass 3 |
|------|--------|--------|--------|
| 0 — Primary | 1 Important (model-policy contradiction) | clean | **clean** |
| 1 — Defect hunt | 1 Important (unguarded `git status`) | 2 Important (sandbox leak, TOCTOU) | **clean** |
| 2 — Principles (Merged) | clean | clean | **clean** |
| 3 — Security | 1 Important (symlinked destination) | **1 Critical** (unvalidated sources), 2 Important | **clean** |
| 4 — Adversarial | 2 Important (panel diff scope, harness blind spot) | 1 Important (stale plan record) | **clean** |
| 5 — Lens B | 1 Important (effort SSoT) | 1 Important (spec/contract duplication) | **clean** (withdrew it on the merits) |
| 6 — Lens C | 1 Important (exit-code contract) | 1 Important (undocumented failure semantics) | **clean** |

Pass 1 diff: `git diff --cached <merge-base> -- . ':(exclude)openspec/'` — **wrong**; that pathspec belongs
to the human gate, not the panel, so no pass-1 slot saw the plan or the five spec deltas. Caught by slot 4
and corrected for passes 2 and 3, which read the unfiltered branch diff.

Pass 2 and pass 3 were both **full** re-runs, escalated automatically rather than asked: a new Critical
surfaced, the fixes altered a delta spec and planning artifacts, and they touched files outside the
findings' named sets.

Two fix waves ran, each a single subagent given the whole deduped list. Three of pass 2's findings were
defects *in pass 1's own repairs* — including the Critical, where hardening the destination left the
source side unvalidated.

### Deferred minors, none blocking

- The two guards are enforced by contract text alone, as every guard in this repository is — not a
  regression this change introduces.
- `specs/myflow-finish-cleanup/spec.md` does not carry the script's own "narrowed, not closed" caveat
  about the residual destination race; the caveat is in the script header and in `tasks.md` Task 10.
- `scripts/test-preserve-session-records.sh`'s unwritable-destination case is a no-op if run as root.
- Uneven line wrapping in some `state-file.md` / `state-self-heal.md` prose.
- Pre-existing and untouched: a guardrail in `skills/myflow-finish/SKILL.md` cites "step 2.1", which is
  not a heading in that file.

### Parked with a ruling

`skills/myflow-start/SKILL.md` still names `medium` literally in its AskUserQuestion block, its revision
sentence and its handoff template. All three are mandated verbatim by the plan, and the operator ruled
that only the *derived* claims elsewhere were to be de-duplicated. Real, deliberate, recorded.
