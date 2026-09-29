# Design — kan-858-slim-verify-contracts-and-reviewer-prompts

Row ids come from the audits (`audit-verify-and-contracts.md` and `audit-subagent-prompts.md`,
`## Candidates`). The audits ran at `700e184`. Every row was re-located by quoted text on
`e6db4832`.

## Decisions

### D1 — the verifier reads its own steps from a sibling file

**ID:** verifier-steps-file
**Status:** active
**Chosen:** `skills/flow/visual-verify-verifier.md` holds step 4, steps 7–11 and the `## Report`
template, verbatim. The prompt contract in **Steps 3–13** names the file's absolute path, beside
`visual-verify.md`, in place of "run steps 4 and 7–11 below".

Two clauses inside step 10 are parent duties: the `frames:` reconciliation and the `matrix:` check.
Both sentences stay whole in the verifier file, because they define the report lines the verifier
writes. Verbatim copies also sit before **Blocking** in `visual-verify.md`, so the parent still sees
both duties.

**Considered:**
- Keeping the steps in the parent file and telling the verifier to read that. Rejected: the parent
  loads the 40 KB it never acts on, which is the cost VV-05 exists to remove.

### D2 — the tooling analyst gets its own file, and `dpSites` follows it

**ID:** tooling-analyst-file
**Status:** active
**Chosen:**
- `skills/flow/visual-verify-tooling-analysis.md` holds the analyst's dispatch paragraphs, prompt,
  recording, re-run and abort.
- The miss definition and classification stay in `visual-verify.md` under the unchanged heading
  **A missed defect — the tooling analysis**, so no heading citer moves. A load directive follows
  them.
- `dpSites` now requires TOOLS, MODEL HANDSHAKE and NO DELEGATION at min 1 in each file. The
  harness fixture writes both files, case 50a tests the new file at its own threshold, and the
  clean-run site count is 30.

### D3 — `workspace-isolation.md` is read by section, not split

**ID:** workspace-isolation-by-section
**Status:** active
**Chosen:** the directive in **Verify** names the one section each trigger needs, read with
`grep -n` plus `sed -n`:

| Trigger | Section read |
|---|---|
| exit 0, stderr names a `cache index` row | **The cache index** |
| exit 1 | **The empty id** |
| exit 2 | nothing |

The hand-apply case (the script cannot be located) keeps its citation of the whole contract.

**Considered:**
- A physical split. Rejected: `check_cleanup_complete_test` reads the file's fenced id blocks, and
  about 20 citers point into its sections.

### D4 — the two-commit chain is its own contract file

**ID:** commit-chain-file
**Status:** active
**Chosen:** `skills/flow-contracts/git-boundaries-commit-chain.md` holds the guarded chain, the
skip and fail rules and the symlink case, verbatim. Its heading, **The guarded two-commit chain**,
is what three places now cite:

- `integrate.md`'s run-as-one-command line;
- `finish-contract-run1.md`'s "is the chain … gives";
- `git-boundaries.md`'s pathspec-default exception.

`integrate.md` loads the file beside `git-boundaries.md`. `verify-and-handoff.md` loads it only on
the `prUrl` path.

**Cost:** the finish session now loads both files, which adds 646 B of header and directive. GB-02,
moving Planning commits out of the finish session, would recover it. It is out of scope here.

### D5 — calibration is one file two reviewers read

**ID:** calibration-file
**Status:** active
**Chosen:** `skills/flow/reviewer-calibration.md` is the single copy of the severity scale.

- The per-task reviewer's REPORT FILE paragraph cites it instead of the whole primary template
  (5.8 KB → 0.7 KB).
- The primary prompt's `## Calibration` tells the slot to read `[CALIBRATION_PATH]`.
- `review-panel.md`'s placeholder table fills that path the same way it fills `[PRINCIPLES_PATH]`.

**Considered:**
- A verbatim second copy. Rejected: two copies of a severity scale drift.

**Cost:** the primary slot makes one extra read of 0.7 KB.

### D6 — rows not done

**ID:** rows-deferred
**Status:** active
**Chosen:** these rows are not done.

| Rows | Reason |
|---|---|
| VH-21 | Its canonical copy in `SKILL.md` was removed by KAN-855. `implement.md` states only that the parent does the work itself, not that the handoff is printed directly in the same turn, so the paragraph stays. |
| VV-07, VV-09, VV-13, VV-14 | Low confidence in the audit. Each carries a scope note, or the next sentence depends on it ("The report's `sweeps:` line then…"). |
| VV-01–VV-03 | Not in KAN-858's scope. VV-01 is the hook-enforced baseline pointer. |
| WI-01 | Deleting the rationale pointers needs a repository convention decision; the audit asked the dispatcher. |
| WI-05, WI-10 | Not in scope. |
| GB-02 | High churn: 13 citer lines for 3.5 KB saved, in the finish session only. |
| SR-01, WR-01, VH-01, VH-02, VH-04, VH-16 | Not named by KAN-858. VH-01, VH-02 and VH-16 are also low confidence with a drift caveat. |
| MECHANICS rows | Out of scope. |

## Load-condition review

Each new directive's condition was checked against every place the moved text applies.

| File | Condition | Review |
|---|---|---|
| `visual-verify-verifier.md` | read by the verifier `flow.visual-verify` dispatches | Every moved step was already "run by one verifier per worktree" (**Steps 3–13**). No parent step moved: steps 3, 5, 6, 12 and 13 and **Blocking** stay. |
| `visual-verify-tooling-analysis.md` | fix run with at least one miss | The moved text opened "On a fix run with at least one miss, the parent dispatches…". A worktree with no miss "dispatches none", and that sentence stays in `visual-verify.md`. |
| `workspace-isolation.md` by section | exit-0-with-cache-index, exit 1 or exit 2 | Exit 1 is a dropped row, which **The empty id** covers (relay the row, stop). Exit 2 stops with the script's own lines. The cache index procedure is its own section. |
| `git-boundaries-commit-chain.md` | integrate run 1, or a recorded `prUrl` | The chain runs through `commit-split.sh` only at those two call sites (`finish-contract-run1.md`: "at both this call site and the implement phase's PR-exception path"). |
| `artifacts-registry.md` (no longer loaded by `implement.md`) | — | No implementation step creates and then removes an artifact by the registry's table. The throwaway copies and the render target carry their own removal. The archive run still loads it. |

## Measured

`scripts/load-sets.sh . /dev/null`, before (`e6db4832`) → after:
<!-- measured: scripts/load-sets.sh @ branch claude/brave-hamilton-533aad -->

| Session | Before | After |
|---|---|---|
| Implementation, definite | 215,378 B | 204,487 B |
| Implementation, incl. cited | 259,878 B | 244,286 B |
| Implementation, worst case | 423,128 B | 368,165 B |
| Planning, incl. cited | 127,823 B | 125,574 B |
| Finish, definite | 182,449 B | 183,095 B |
| Reviewer: baseline + primary | 11,966 B | 10,240 B (+0.7 KB calibration read) |
| Reviewer: baseline + principles + EP | 24,181 B | 20,657 B |
| Reviewer: baseline + failure modes | 12,836 B | 11,495 B |

`visual-verify.md`, which the parent loads on every UI run, drops from 61,653 B to 17,489 B.

`check-normative-inventory.sh` output is byte-identical before and after. The failing set of
`go test ./...` is the same on `e6db4832` as on this branch (10 guard tests, caused by the
environment: root and locale). No test fails that did not fail before.

The transcript of one real UI run, which KAN-858's "Done when" asks for, cannot be produced in this
environment: it has no UI project and no `flow`/`spectre` CLIs. Structurally, the verifier's prompt
contract names every item it named before and adds the file path.

## Open questions
