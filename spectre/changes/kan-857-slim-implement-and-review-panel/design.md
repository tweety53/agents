# Design — kan-857-slim-implement-and-review-panel

Row ids are the audits' (`audit-implement.md` and `audit-review-panel.md`, `## Candidates`). The
audits ran at `700e184`. Every row was re-located by quoted text on `b8cc6e75`.

## Decisions

### D1 — one file per load condition, siblings in `skills/flow/`

**ID:** split-files-by-load-condition
**Status:** active
**Chosen:** seven new files, each behind a `**Load …** only when` directive at the point where its
text used to sit:

- `document-fix.md`
- `cross-repo-worktrees.md`
- `sdd-dispatch.md`
- `gated-review-fix.md`
- `review-panel-late-fix.md`
- `review-panel-fix-round.md`
- `review-panel-experimental-slot.md`

They live in `skills/flow/`, so `check-guard-symlinks` rule 2 still sees their guard citations.
**Considered:**
- One "conditional" file per phase file. Rejected: the fix-run condition and the `sdd` condition
  are independent. A merged file would load sdd text into every fix run, and the reverse.

### D2 — headings stay where they were cited

**ID:** stub-headings-stay
**Status:** active
**Chosen:**
- `## 3. Documenting a fix, before implementing it` and `### The late-fix reduction` keep their
  heading in the original file, with the load directive as their body.
- `## Panel re-runs` keeps its heading and its triage rules (Critical/Important to the fix, Minor
  deferral, the deferral reason, rules that change mid-run, the stale definition).
- `review-panel-fix-round.md` opens its own `## Panel re-runs`, so a citation of either file
  resolves.
- The router in `SKILL.md`, `/flow-fast`'s "**Check base movement first** through **Panel
  re-runs**" range and every `**The late-fix reduction**` citer need no edit.

The Panel re-runs citers whose content moved are repointed to `review-panel-fix-round.md`:

- the commit route: `implement.md`'s full-suite fix, `gated-review-fix.md`, and
  `verify-and-handoff.md`'s stage-diff;
- the re-run diff rules: the three reviewer prompts' `[DIFF_PATH]`;
- the rerun policy: two citations in `brainstorm-planner.md`.

### D3 — the fix-subagent paragraphs and the panel-fix record stay in `review-panel.md`

**ID:** pinned-stay
**Status:** active
**Chosen:** these stay in `review-panel.md`, because `dpSites` pins them there:

- VERBATIM, FINDINGS, PLAN FIELDS, ROUND SCOPE, FOREGROUND, TOOLS, NO DELEGATION, the handshake
  pointer, TARGETED, OUTPUT BUDGET, MUTATION PROOF, PIXEL PROBE and REPORT FILE;
- the panel-fix `flow record dispatch` pair, so `check-stage-mark-calls` keeps scanning it.

The fix-round file's header says so. The chunking paragraph and the non-convergence loop move,
because no guard pins them.
**Considered:**
- Re-pointing `dpSites` (RP02, 7.4 KB more). Rejected for this change: it edits a guard's
  contract, not a prompt.

### D4 — the omission rule stays in `review-panel.md`

**ID:** omission-rule-stays
**Status:** active
**Chosen:** "A slot that supplies nothing for a finding has not supplied a legal exemption"
applies whenever a finding is recorded, including in a Minor-only round. It moved to **Recording
findings** instead of into the fix-round file.

### D5 — rows not done

**ID:** rows-deferred
**Status:** active
**Chosen:** these rows are not done.

| Rows | Reason |
|---|---|
| R19, R37, RP31, RP32 | The audit itself scored them low or not counted. |
| R24, R27, R28 | Cutting the clause needs more acknowledged rewording than it saves. R27 and R28 also sit on a `git checkout <sha> -- .` span that the sentence splitter cuts mid-command. |
| RP11 | The slot handshake pointer. KAN-853 already fixed its drift, and it is the only place that names the slot's retry key. |
| RP08 | Its canonical copy in `SKILL.md` no longer exists. |
| RP13–RP17, RP43, RP44 | Duplicates inside `review-panel-fix-round.md`. That file now loads only with a fix round, so the saving is conditional. |
| D40, RP18, RP19, R25, R26 | Their text is already gone from `b8cc6e75` (KAN-853 and the bugbot/security retirement). |
| RP02, RP06 | Guard-blocked. |
| MECHANICS rows | Out of scope. |

### D6 — optional-slots splits by slot

**ID:** optional-slots-by-slot
**Status:** active
**Chosen:** `review-panel-optional-slots.md` now loads for `mutation` only. It carries the
mutating-role bundle sentence (RP05), MUTATION ENTRY CONTEXT (RP04) and **The throwaway
worktree**. `review-panel-experimental-slot.md` loads for an `exp-` slot. `review-panel.md`
carries one directive for each.

## Load-condition review

Each directive's condition was checked against every place the moved text applies.

| File | Condition | Review |
|---|---|---|
| `document-fix.md` | fix run: `IN_PROGRESS` plus instructions, or a plain message at `IN_PROGRESS` | The moved text opens "**Fix runs only**". The router sends both fix-run shapes through `implement.md`, which loads it at section 3 before section 1. A resumed fix run is again a fix run. |
| `cross-repo-worktrees.md` | a linked peer, an `## apps` entry with no worktree, a `regression checkout`, or a resolved set spanning more than one repository | Worktree setup and `spectre link` act only on a created worktree: a peer or an app. The apps recipe acts only on an entry with no worktree. The toolchain acts only on the regression checkout, created or resumed. The merge-order record needs more than one repository. The fourth clause was added in review: a fix run whose app worktrees already exist still extends the record. |
| `sdd-dispatch.md` | `execution` is `sdd` | Every moved paragraph dispatches or picks an implementer group; inline dispatches none. The panel's own gather passes `""` for the sixth argument in its own command. "Never read the bundle back" survives inline as the Read-discipline bullet. |
| `gated-review-fix.md` | a gated reviewer pass closes `fix` | The moved text is exactly that path. The clean and Minor-only cases stay in `implement.md`. |
| `review-panel-late-fix.md` | fix run, before the first round | The section fires only on a fix run ("never on a creating run"). The staleness carve-out concerns only a delta the reduction read. |
| `review-panel-fix-round.md` | a round recorded a Critical or Important finding, or a close guard sends the run to the handback loop | A Minor-only round defers and closes without it. Rounds after pass 1 exist only after a fix. The reproducer guards and runs happen "before dispatching the fix subagent". The second clause was added in review: `check-panel-findings-closed.sh` exit 1 returns to the handback loop, which now lives in this file. Citers outside the panel stage (the full-suite fix, the gated fix and the stage-diff check) name the file at the point of use. |
| `review-panel-optional-slots.md` | roster carries `mutation` | Everything left in it concerns the mutating slot. |
| `review-panel-experimental-slot.md` | roster carries an `exp-` slot | Unchanged text; the old combined condition, split. |
| `model-policy.md` **Harness mapping** | harness `zcode` | `implement.md`'s handshake keeps a one-line pointer naming the file at the point of use. |

## Measured

`scripts/load-sets.sh . /dev/null`, before (`b8cc6e75`) → after:
<!-- measured: scripts/load-sets.sh @ branch claude/brave-hamilton-533aad -->

| Session | Before | After |
|---|---|---|
| Implementation, definite | 278,660 B | 215,026 B |
| Implementation, incl. cited | 322,629 B | 259,591 B |
| Implementation, worst case | 427,211 B | 422,480 B |
| Planning, definite | 108,514 B | 108,534 B |
| Finish, definite | 182,449 B | 182,449 B |

Per-file sizes:

| File | Before | After |
|---|---|---|
| `implement.md` | 84,561 B | 62,391 B |
| `review-panel.md` | 102,912 B | 61,438 B |

A first inline run that records no Critical or Important finding therefore loads 63.6 KB less. A
fix run adds `document-fix.md` (9.0 KB) and `review-panel-late-fix.md` (4.4 KB). A round with a
Critical or Important finding adds `review-panel-fix-round.md` (34.2 KB).

`check-normative-inventory.sh` output is byte-identical before and after. `go test
./internal/guard/...` fails the same cases on `b8cc6e75` as on this branch (environment: root,
locale); no new failure.

## Open questions
