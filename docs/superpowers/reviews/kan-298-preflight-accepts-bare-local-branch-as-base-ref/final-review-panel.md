# Review panel — kan-298-preflight-accepts-bare-local-branch-as-base-ref

Resolved roster (from the settings store, `flow settings get`): `primary`, `principles`,
`code-review-low`, `bugbot`. Model for every slot taking an override: `sonnet`.

**Substituted:** Bugbot. This harness's Agent tool offers no `bugbot` agent type, so the slot ran as
a **general-purpose** subagent carrying the Bugbot brief with mutation testing made mandatory. Its
dispatches record `-slot Bugbot` and `-model sonnet` — never `unknown (agent-defined)`, which would
falsely claim the real agent definition ran.

**Added this run:** none — the resolved list ran alone, in every pass.

**Diff size:** `check-panel-diff-size.sh` measured 36 changed lines against the cap, exit 0, under
cap. No operator question was reached.

---

## Pass 0 — initial panel

**Mode:** initial. All four resolved slots. Diff read: `.superpowers/sdd/final-review.diff`
(`git diff ca7b003`), branch at `6e4340b4`/`dc890c44`.

| Slot | Verdict |
|------|---------|
| Primary | PASS — also searched for other silent callers of `resolve_remote_base`; only the two guards this change covers |
| Principles | PASS — no findings at any severity |
| Code review (low) | PASS — no high-confidence defects |
| Bugbot (substituted) | FINDINGS — two surviving mutants, F1 and F2 |

**Reproducer refusal.** `check-panel-reproducers.sh` exited 1 on F1: *"F1's reproducer command
carries a shell metacharacter — a runnable reproducer is a bare path optionally followed by plain
arguments, never a shell command line."* Per the contract this is a refusal, not a repair: F1's
reproducer was recorded unverifiable and put to the operator rather than silently rewritten. The
dispatcher independently confirmed F1 by reading `test-check-finish-preflight.sh:40-45` (`run_guard`
merges streams with `2>&1`) and `:317-321` (the missing-argument case asserts only `$RC`). The
operator directed the fix to proceed on those verified merits. F2 carried the legal `none — <reason>`
exemption and was dispatched without a run. No finding was bounced.

**Operator decision:** fix F1 and F2 together in one round.

## Fix round 1

Test-only; no production code touched. One case added to each harness asserting that the
no-arguments path leaves stdout empty and writes exactly the usage message to stderr. Two fixup
commits, folded by `git rebase --autosquash`.

The dispatcher's own independent mutation proof of the two behaviours the fix added:

```text
fix-mutation: scripts/check-finish-preflight.sh — dropped `>&2` from the usage heredoc — missing arguments: stdout is empty
fix-mutation: scripts/check-finish-preflight.sh — deleted the `exit 2` after `EOF` — missing arguments: stderr is exactly the usage message
fix-mutation: scripts/check-base-moved.sh — dropped `>&2` from the usage heredoc — missing arguments: stdout is empty
fix-mutation: scripts/check-base-moved.sh — deleted the `exit 2` after `EOF` — missing arguments: stderr is exactly the usage message
fix-mutations-total: 4
```

**Disclosed rather than passed over:** the fix's diff touches neither path F1 and F2 named
(`check-finish-preflight.sh:54`, `check-base-moved.sh:50`) — both findings are coverage gaps whose
fix belongs in the harness, which is what the raising slot itself prescribed. The "fix must touch a
path the finding named" check is therefore satisfied here by mutation evidence rather than by the
diff's paths, recorded as such rather than left to look like an ordinary pass.

## Pass 1 — Full mode

**Escalated automatically, not asked:** the fix touched files outside the set the findings named.
Primary read the rewritten `final-review.diff`; Principles and Code review (low) read
`slot-delta-1-<slot>.diff` from `dc890c44`; Bugbot reads no diff file. Branch at
`4c5d3b1d`/`a7c533ef`.

| Slot | Verdict |
|------|---------|
| Primary | PASS |
| Principles | Compliant — one Minor, F3 |
| Code review (low) | PASS |
| Bugbot (substituted) | PASS — F1 and F2 confirmed closed |

Bugbot's table also showed the new exact-match assertion catches two things the design never planned
for: a reworded usage line that keeps the grepped phrase intact, and an indentation-only change.

**Operator decision:** fix F3 rather than withdraw it.

## Fix round 2

The expected usage text extracted into `scripts/lib/base-ref-usage.sh`, exposing
`base_ref_usage_message <guard-name>`, sourced by both harnesses. Two fixup commits, folded.

The dispatcher's own independent mutation proof, including a mutation whose only job is to catch the
extraction becoming self-referential — deriving the expected value from the guard would make both
assertions vacuous:

```text
fix-mutation: scripts/check-finish-preflight.sh — dropped `>&2` — missing arguments: stdout is empty
fix-mutation: scripts/check-finish-preflight.sh — deleted `exit 2` — missing arguments: stderr is exactly the usage message
fix-mutation: scripts/check-finish-preflight.sh — reworded a usage line, grepped phrase left intact — missing arguments: stderr is exactly the usage message
fix-mutation: scripts/check-finish-preflight.sh — changed the usage text visibly (tautology probe) — missing arguments: stderr is exactly the usage message
fix-mutation: scripts/check-base-moved.sh — dropped `>&2` — missing arguments: stdout is empty
fix-mutation: scripts/check-base-moved.sh — deleted `exit 2` — missing arguments: stderr is exactly the usage message
fix-mutation: scripts/check-base-moved.sh — reworded a usage line, grepped phrase left intact — missing arguments: stderr is exactly the usage message
fix-mutation: scripts/check-base-moved.sh — changed the usage text visibly (tautology probe) — missing arguments: stderr is exactly the usage message
fix-mutations-total: 8
```

## Pass 2 — Full mode

**Escalated automatically:** the fix added a new file outside the set F3 named. Branch at
`01a6d144`/`cc186b6b`.

| Slot | Verdict |
|------|---------|
| Primary | FINDINGS — F4 and F5 |
| Principles | Compliant, no findings — reversed its own pass-1 position on the merits, judging the extraction sound now that it exists |
| Code review (low) | PASS |
| Bugbot (substituted) | PASS — 15-row mutation table, no surviving mutants; L1 proves the golden value is independent, not derived |

**A dispatch error of the dispatcher's own, recorded rather than buried.** Bugbot and Primary ran
concurrently against one shared worktree, and Bugbot's brief is to mutate files in place. Primary
observed Bugbot's in-flight mutations and reverted them, which produced F5 and is the most likely
cause of the single non-reproducing suite failure Principles reported here and Code review (low)
reported in pass 1. Contention of this shape can only turn a caught mutant into a *surviving* one,
never the reverse, so no slot's "caught" result is called into question by it — and the dispatcher's
own uncontended mutation runs corroborate every one. The lesson is scheduling: a mutation-testing
slot must not share a worktree with slots that assert worktree cleanliness.

## Fix round 3

Dispatched **alone**, with no concurrent agent, and told not to accept the dispatcher's own
hypothesis. It searched every commit's tree, all history via `git log --all -S`, the reflog and
dangling objects, the index, the working tree and untracked files. **It found no residue and
correctly made no change.**

```text
fix-mutation: none — the round found no defect to fix, so there was nothing to mutation-prove
fix-mutations-total: 0
```

Cause: leaning toward a concurrent reviewer's in-flight mutation, at moderate-to-high confidence,
**explicitly not conclusive** — it could not correlate against Primary's own pass-2 report, which
the dispatcher had failed to write to disk at the time. That omission was the round's own finding
about the dispatcher, and the file has since been written.

Flake question settled separately by the dispatcher, with nothing else touching the worktree:
20 consecutive runs, 10 per harness, all exit 0. Counting the dispatcher's earlier runs, 30 clean.

**No pass 3 was dispatched.** Fix round 3 changed no code, so every slot's delta is empty and the
whole-branch diff is byte-identical to the one all four slots passed in pass 2 — *not re-run,
nothing new since its last read*.

## Outcome

F1 fixed, F2 fixed, F3 fixed, F4 fixed, F5 withdrawn by the operator after investigation.
Zero findings open at any severity.
