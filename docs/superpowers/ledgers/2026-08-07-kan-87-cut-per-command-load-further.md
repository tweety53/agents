# SDD ledger — kan-87-cut-per-command-load-further

Models from the state file: `models.implementation` = Sonnet, `models.reviewPanel` = Sonnet,
`models.panelFix` = Sonnet. All three are operator choices recorded at the creating
`/myflow-start` run, overriding the pipeline defaults of Opus / Sonnet / Opus.

| Task | Dispatch | Model | Outcome |
|------|----------|-------|---------|
| 1.1 + 1.2 | general-purpose | **sonnet** | complete (uncommitted); reported the re-wrap consequence of a lengthened citation token |
| 1.3 | parent session | **opus** | added mid-run; hoisted the worktree scan `state-self-heal.md` depends on |
| 2.1 | general-purpose | **sonnet** | complete; **stopped and reported** the `state-self-heal.md` dependency rather than creating a dangling citation |
| 3.1 + 3.2 + 3.3 | general-purpose | **sonnet** | complete; reported a paragraph-atomicity conflict and a hollowed `Convergence` citation |
| 4.1 | general-purpose | **sonnet** | complete; reported stale canonical-authority prose in three files |
| 5.1 | general-purpose | **sonnet** | complete; proved the basename collision with a scoped mutant rather than asserting it |
| 6.1 | parent session | **opus** | verification only |

## Review panel

Four slots, all on **sonnet** from `models.reviewPanel`: adversarial, defect-hunt,
principles (Merged), principles (Lens C — robustness & ops).

**Roster note.** The `bugbot` and `security-review` agent types do not exist in this harness, so the
defect-hunt slot ran as `general-purpose` reading `bug-hunter-reviewer-prompt.md`. Its ledger entry
reads `sonnet (prompt-file substitute)` rather than `unknown (agent-defined)`, because the
dispatcher chose the model. A separate security slot was **not** run: the only executable surface is
one guard script and its harness, and the security-relevant properties of that script — symlink
refusal before read, no `awk -v` injection path, no false clean — were carried forward from KAN-82
and re-verified by the defect-hunt and Lens C slots. That is a deviation from the panel contract's
conditional-slot table, recorded here rather than left to be inferred from the agent count.

All fix waves were applied by the parent session on **opus**, not by a fix subagent.

## What the panel changed about the change

**Five implementer dispatches, five real plan-or-spec defects reported rather than worked around.**
None was catchable by a guard. Each is recorded in `tasks.md` at the task that found it.

**The panel then found five more**, four of them Major, and two of those were **test-coverage gaps
proven by mutation** — the `SKILL-rationale.md` glob and the double-count prefix skip each had zero
coverage, and mutating them left all 17 harness cases green.

**`scripts/check-references.sh` was green throughout and proved nothing.** It resolves a citation
whenever a heading of that name exists, and the mirroring rule guarantees one exists in both files.
Three of this change's defects lived in exactly that gap — a hollowed `Convergence` citation, a
dropped citation clause, and a cross-command dependency on a moved scan. All three were found by an
agent reading.
