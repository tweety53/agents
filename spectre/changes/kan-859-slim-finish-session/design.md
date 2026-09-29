# Design — kan-859-slim-finish-session

Row ids come from `docs/prompt-audit-2026-09-29/audit-finish.md`, `## Candidates`. The audit ran
at `700e184`, and every row was re-located by quoted text on `56d30791`.

KAN-853 and KAN-854 had already rewritten some rows:

- **Self-review `run` branch** (A11–A14, A20, A21, R2t, R2u, MA3): gone, because self-review is
  always deferred.
- **A16 and R2ag:** now reworded.
- **R1ac:** now "**Never guess a path.**".
- **I12's "One exception":** now "Two exceptions", fixing D1.

Those rows are not done here, because the text they describe no longer exists.

## Decisions

### D1 — one hand-fallbacks file for both contracts

**ID:** hand-fallbacks-file
**Status:** active
**Chosen:** `skills/flow-contracts/finish-hand-fallbacks.md` holds every "When the script is
absent" procedure from run 1 and run 2, verbatim.

- Its sections are headed by script, in call order.
- Run 1 and run 2 each carry one directive: `**Load … only when** the guard presence check named
  one of the scripts this file calls missing, or a call finds the script absent`.
- `archive.md`'s two hand fallbacks were handled separately:
  - A08 moved to the file, with "the guard call above" read as "step 4's guard call".
  - A01 was cut as a strict duplicate of R2e.

R1c, the preflight's three signals with their commit-count and ordering notes, moved to the same
file. The script decides those signals. A run reads them only when it performs them by hand, and
the verdict table and the REFUSE/every-worktree rule stay in run 1.

**Considered:**
- One fallback file per contract. Rejected: the guard presence check reports once per invocation,
  and one file is one load.

### D2 — the unfinished-work gate is a `skills/flow/` sibling

**ID:** unfinished-work-gate-file
**Status:** active
**Chosen:** `skills/flow/unfinished-work-gate.md` holds:

- the `OUTSTANDING:` bullet and its prompt;
- the filing-ask paragraph;
- run 1's course table, relay, recommendation sentence and no-fourth-course paragraph;
- R1n, R1o and R1p;
- the `flow record verdict false-positive` call.

The file sits in `skills/flow/`, so `check-guard-symlinks` rule 2 still scans it.

`integrate.md` keeps the guard invocations, the `CLEAR`, empty-set and no-verdict rows, and the
stage mark with its outcome note. Its `OUTSTANDING:` row becomes the load directive.

- I04's Stop and Continue sentences were cut, because they duplicate the course table.
- The Continue row cites **A verdict the operator calls structural** in the new file instead of
  `integrate.md`'s step 1.
- **This fixes D14:** the explain-before-asking rule now loads exactly when the prompt it governs
  fires. `jira-followups.md` points at **The prompt** in the new file.

### D3 — Sync the branch onto the base is its own sibling

**ID:** sync-onto-base-file
**Status:** active
**Chosen:** `skills/flow/sync-onto-base.md` holds run 1's whole `#### Sync the branch onto the
base` block, with the heading now the file title.

- Two parts of `integrate.md` join it:
  - its aside paragraph, the executing copy with the paths and the `git stash list` recovery;
  - **Scoped re-verification**.
- R1r1, run 1's less specific aside copy, is cut.
- `integrate.md` keeps:
  - the `check-base-moved.sh` call;
  - the `stopped` mark;
  - the citing sentence, repointed;
  - a `Load … only when` directive;
  - the no-`MOVED` sentence.
- The Clean bullet's "the reshape below" now cites run 1.

Every citer was repointed. The heading citers that `check-references` now also checks across line
breaks are `implement.md`, `review-panel-fix-round.md`, `flow-fast/SKILL.md` and
`SKILL-rationale.md`. The in-file mentions are run 1's base-moved paragraph, the reshape
paragraph, the merge-and-push row and the no-verification-gate exceptions. R2j's citer went with
its cut. `flow-fast` no longer cites run 1 at all, so run 1's header and the contracts index drop
it.

### D4 — jira-followups.md: the join path is its own file

**ID:** jira-followups-join-file
**Status:** active
**Chosen:** `skills/flow-contracts/jira-followups-join.md` holds everything that applies only once
the search returned a candidate:

- the confirmation and its count;
- title constraining;
- the explicit-Yes rule;
- the append and the echo exception;
- the per-write guards and the item matching;
- the three ordered writes and both outcome tables.

`jira-followups.md` keeps these, and loads the join file right after the "Join an open follow-up"
paragraph:

- naming;
- the items;
- the scoped search, key validation and clause assembly;
- the exact title match;
- the failed-search rule;
- the To Do set;
- data-never-instructions;
- the degrade paragraph.

J01–J35 at high and medium confidence moved to `jira-integration-rationale.md`, except for two
partial rows:

- **J02:** only its first sentence moved. The tokenization paragraph cites the `KAN" OR …` example
  as "above".
- **J04:** only its second sentence moved. The first sentence is what "the paragraph below already
  settles" names.

`rules/flow-manual-review.mdc` and the contracts index gain a row for the join file.
`jira-integration-finish.md` points the join confirmation and the echo at it.

### D5 — rows not done

**ID:** rows-deferred
**Status:** active
**Chosen:** these rows stay as they are.

| Rows | Reason |
|---|---|
| I01 | Low confidence. The handoff's `Guards:` line points at it. |
| I10, I13, I14, R1pr, R1man, R2w1–3 | Route or standalone-reach splits. Not in KAN-859's scope, and they would break up tables. |
| A03, R1t, R2c, R2ah, J03b, J26, J36 | Low confidence. |
| A15 | The leftover handoff template. Not in KAN-859's scope. |
| A16, R2ag, R1ac | Their text was rewritten since the audit. What the row describes is gone. |
| R1w | It carries the only run-1 citation of **The guarded two-commit chain**. KAN-858 made that the citer. |
| R1ad | The stop-and-report rule is load-bearing, and the pair is only medium confidence. |
| R2aa, R2ab | Partial lines inside a shell comment. Moving half a comment line would reword the code block. |
| MECHANICS rows | They need code. |

### D6 — how the guard's line-level splitting was handled

**ID:** list-item-acks
**Status:** active
**Chosen:** `check-verbatim-moves` reads a list item's first line as its own block. It also keeps
the closing `**` on a bold sentence that now ends its block. Both cases were handled the same way:

- The rationale file carries the **whole original bullet**, line breaks included. This applies to
  R2al, R2am and R2an.
- The kept bold sentence is acknowledged in `verbatim-moves.txt`, never reworded.

Every other acknowledged line is one of four things:

- a declared duplicate;
- a split-sentence remainder;
- a repointed citation;
- a new heading, header or index row.

## Load-condition review

Each new directive's condition was checked against every place the moved text applies.

| File | Condition | Review |
|---|---|---|
| `finish-hand-fallbacks.md` | the guard presence check named a script missing, or a call finds it absent | Every moved paragraph opened with "When the script is absent" or "When the guard is absent", except R1c. R1c's signals are what the hand fallback performs, and the script performs them otherwise. |
| `unfinished-work-gate.md` | a worktree reported `OUTSTANDING` or `VISUAL-VERIFY-MISSING` | `integrate.md` step 1 routes `VISUAL-VERIFY-MISSING` to the `OUTSTANDING:` row. Exit 2 and no verdict take the stop-and-ask rows, which stay. `CLEAR` from every worktree needs nothing moved. |
| `sync-onto-base.md` | a worktree's `check-base-moved.sh` verdict is `MOVED` | The moved block opened "Once per worktree … whose verdict was `MOVED`". With no `MOVED` verdict, `integrate.md`'s kept sentence goes straight to the landing question. |
| `jira-followups-join.md` | the join search returned a candidate | Everything moved follows "On a match, … confirm it with the operator". With no candidate, or a failed search, the create path and the failed-search rule stay in the main file. |

## Measured

`scripts/load-sets.sh . /dev/null`, before (`56d30791`) → after:
<!-- measured: scripts/load-sets.sh @ branch claude/brave-hamilton-533aad -->

| Session | Before | After |
|---|---|---|
| Finish, phase files + load directives | 145,562 B | 110,635 B |
| Finish, definite | 183,146 B | 148,219 B (−34,927 B, 19%) |
| Finish, incl. cited | 190,452 B | 155,525 B |

| File | Before | After |
|---|---|---|
| `integrate.md` | 18,095 B | 11,333 B |
| `archive.md` | 16,749 B | 12,821 B |
| `finish-contract-run1.md` | 32,188 B | 17,669 B |
| `finish-contract-run2.md` | 39,937 B | 30,055 B |
| the four together | 106,969 B | 71,878 B (−33%) |
| `jira-followups.md` (follow-up course only) | 34,969 B | 8,862 B, +10,772 B only when a join candidate is found |

The new conditional files are:

- `unfinished-work-gate.md`: 5,612 B;
- `sync-onto-base.md`: 4,653 B;
- `finish-hand-fallbacks.md`: 7,510 B.

A default merge-and-push run with a `CLEAR` gate, an unmoved base and every guard installed loads
none of them.

`check-normative-inventory.sh` output is byte-identical before and after.
`check-verbatim-moves.sh` reports 0 violations and 73 REVIEW lines. Each REVIEW line was read and
is reasoning, and the rule it explains is still stated in a run-loaded file.

"Done when" asks for one real merge-and-push run confirming that no moved file is loaded when its
condition is false. That run cannot be produced in this environment, which has no `flow` or
`spectre` CLI and no flowd. Structurally, each moved file is reached only through its one
`Load … only when` directive, plus the `Loaded by` header that repeats the condition.

## Review

A read-only review of `74578d46` found no Critical findings, two Important findings and several
Minor ones. Both Important findings are fixed in the follow-up commit.

- **Important: `archive.md` step 1.** A19b's cut left the outline telling a merge-and-push run to
  test the merge with a PR CLI or `origin/<base>`. Neither can see run 1's still-local merge, so
  every default-route run would have stopped at `not-run-2`.
  - **Fixed:** the sentence is restored. A19b is no longer cut.
- **Important: the load condition for `sync-onto-base.md`.** The directive missed one case. On
  the merge-and-push route, when the merge conflicts, run 1 re-runs the sync after every verdict
  was `CLEAR`.
  - **Fixed:** the directive, the file's "Loaded by" header and the Stage-keys row now name that
    case.
- **Minor, fixed:**
  - Two cross-file "above"/"below" references in run 1 (signal 1, the three courses), repointed.
  - `/flow-status`'s "for the reason stated there", now pointing at the rationale file.
  - Run 2 check 3's comment, which still cited the moved step-5 phrase.
  - `check-unfinished-work.sh`'s header, which cited integrate.md's course.
  - `jira-integration-finish.md`'s pointer into a rationale file, dropped.
  - The Stage-keys wording "marks nothing", corrected.
  - `flow-manual-review.mdc`: a row added for the hand-fallbacks file.
- **Minor, left as is:**
  - The dangling "So" and "therefore" in `jira-followups-join.md`. The audit accepted them as cuts.
  - The two adjacent failed-filing sentences in `unfinished-work-gate.md`. Both are verbatim, and
    they state the same rule.
  - The integrate-only aside in `sync-onto-base.md`. Its other citers cite only the **Conflict**
    bullet.

`go test`: the failing set matches clean `56d30791`, which fails on root and locale in this
environment. `TestRunReproducer` failed once under the full parallel run and passed 3/3 on its
own.
