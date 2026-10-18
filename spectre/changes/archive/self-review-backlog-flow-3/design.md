## Context

- Five Backlog Flow findings planned as one change per item 1 of **3. Fix every finding that is
  not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-3-design.md` in this worktree. Every
  question was auto-resolved to its recommended option (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **KAN-609 — Task 1.** **Harness mapping** (`skills/flow-contracts/model-policy.md:94` at
  `9637bf73`) dispatches every zcode dispatch on `subagent_type` `flow-high`, and
  `hooks/enforce-agent-baseline.py:64` pins the verifier and the tooling analyst to
  `("flow-high", "glm-5.3-flash")` on zcode. `setup.sh` runs `install_agents` into
  `~/.claude/agents/` alone (lines 224 and 1168) — no ZCode directory. Across the 45 ZCode sessions
  recorded under `~/.zcode/cli/`, the Agent tool's available types are `general-purpose`, `Explore`,
  `documents`, `pdf` and `presentations`, never a `flow-*` type, and all 78 recorded dispatches
  carry `general-purpose`; `ZCode.app` references no `.zcode/agents` directory. Nothing in the
  contract records a substituted spawn type. Blast radius: four files; ten files cite **Harness
  mapping** or **The roster** and none needs a change.
- **KAN-629 — Task 2.** **A refused start is relayed, never resolved**
  (`skills/flow/verify-and-handoff.md`) ends the run on every held-port refusal and has no case for
  an application whose repository the change holds no worktree in; the `Running:` template has no
  line for one. Blast radius: one file.
- **KAN-631 — Task 3.** `grep -n "check constraint" skills/flow/engineering-principles.md` finds
  nothing; the principles reviewer reads that file. The ticket's gymie instance needs no fix — its
  constraint already exists and caught the insert. Blast radius: one file.
- **KAN-633 — Task 4.** `grep -rn "wire contract" skills/flow` finds nothing; nothing tells the
  parent or the verifier how to verify a flagged diff that draws nothing. Blast radius: two files.
- **KAN-645 — already fixed, no task.** `skills/flow-contracts/project-configuration-visual.md:35`
  requires `verify` to "run its whole suite when `<specs>` is empty" (since `7c1c0052`,
  2026-09-28), and gymie's `verify` row changed from `npm run test:visual` to
  `npx playwright test <specs>` in `e7539736` (2026-09-28; gymie `.flow/project.md:365` at
  `3c57055d`). That command names no `--project`, so it runs both the `functional` and `fidelity`
  projects `gymie-playwright/playwright.config.ts` declares (at `d416244`), and
  `scripts/specs-for-diff.mjs` lists specs from the whole checkout, `fidelity/` included. The
  kan-574 report that filed it is dated 2026-09-22, before both.
- Every task's diff was applied in order to a scratch clone at `9637bf73` (this worktree's base):
  `scripts/test-enforce-agent-baseline-hook.sh` passes, and fails cases 8 and 8b with the hook
  hunk reverted; `check-references.sh`, `check-installed-citations.sh`,
  `check-dispatch-paragraphs.sh`, `check-ask-silence.sh`, `check-hand-notes-in-step.sh`,
  `check-stage-mark-calls.sh`, `check-guard-symlinks.sh`, `check-python-suppressions.sh`,
  `check-markdown-integrity.py` and `check-verbatim-moves.sh 9637bf73` (with this change's
  `verbatim-moves.txt`) exit 0.

## Decisions

### ZCode dispatches on general-purpose

**ID:** zcode-general-purpose
**Status:** active
**Chosen:** align the contract with the harness — **Harness mapping**'s zcode paragraph and the hook's `ZCODE_PAIR` name `general-purpose`, the NO DELEGATION paragraph carrying the no-forking rule the `flow-<effort>` definitions' tool allowlist carries on Claude Code; hook case 8 moves to the new pair and case 8b denies `flow-high` on zcode. The recorded `-effort high` is unchanged.
**Considered:** registering the type — installing `agents/*.md` into a ZCode directory — no evidence ZCode loads agent definitions from any directory (no reference in `ZCode.app`, no `flow-*` type in 45 sessions' Agent tool), so the install would be a guess; changing the recorded effort as well — the fixed zcode pair is what flowd and every `glm-5.3-flash / high` sentence key on (KAN-610 is the store-side follow-up), a wider change than the ticket asks.

### A substituted spawn type goes on the pass log

**ID:** roster-substitution-recorded
**Status:** active
**Chosen:** one sentence in **The roster** (`skills/flow/review-panel.md`): a slot whose spawn type the harness does not offer runs on `general-purpose` and the round records `flow record pass -round <round> -note 'roster: <slot ids> on general-purpose — <type> not offered'`; a harness whose mapping already names `general-purpose` makes no substitution.
**Considered:** a new `flow record dispatch` field for the requested type — a CLI and store change for a fact one pass-log line carries; recording it on every dispatch site — the ticket names the panel's pass log, and `/flow-settings` already warns at write time.

### An out-of-worktree application's refused start does not end the run

**ID:** out-of-worktree-refused-start
**Status:** active
**Chosen:** an exception under **A refused start is relayed, never resolved**: when the refused application's repository holds no worktree of the change, the rest of the stack is handed off and `Running:` carries `Not started: <app> (<url>) — port <port> held by <holder>; outside this change's worktrees.`
**Considered:** keeping the stop for every refusal and only recording it — kan-361's handoff of the in-scope stack was the right call, and a stop makes the operator stop another session's process to review code that process does not serve; a per-app status line for every application — the started ones already have their URL bullet, so only the unstarted one is news.

### Schema-enforceable invariants are a robustness principle

**ID:** schema-check-constraints
**Status:** active
**Chosen:** extend **Robustness and failure principles**' "Violation looks like" list in `skills/flow/engineering-principles.md` with an invariant the schema could enforce held in application code alone, no check constraint refusing a bad insert — a test's own setup included.
**Considered:** a writing-plans rule — plans do not decide schema shape task by task, while the principles reviewer reads every diff that adds a table; a gymie-side brief — the lesson holds for any project with a schema.

### A flagged diff that draws nothing is verified on its wire contract

**ID:** wire-contract-verification
**Status:** active
**Chosen:** the parent adds `wire contract:` lines to the verifier prompt when every matched path draws nothing (`skills/flow/visual-verify.md`); the verifier calls each endpoint on the running API signed in through `## credentials`, checks shape, wire names and the acceptance behaviour, and reports a `wire contract:` line (`skills/flow/visual-verify-verifier.md`); a departure blocks as a screenshot defect does. The steps and `check-verify-report.sh` are unchanged.
**Considered:** narrowing `ui paths` so such diffs skip the stage — the ticket keeps the flag, and the live check is the verification the change needs; a new required `steps:` item — `check-verify-report.sh` would need a Go change for a line the parent already reads.

### KAN-645 is already closed

**ID:** verify-whole-suite-exists
**Status:** active
**Chosen:** no task — the whole-suite rule on `verify` and gymie's `npx playwright test <specs>` row already put the fidelity project inside the green-suite claim.
**Considered:** a guard refusing a `verify` scoped to one Playwright project — `npm run test:visual` named no project; the scope hid in `package.json`, and Playwright's own `--list` per project would be a second `check-spec-reach.sh` for a defect no tree carries.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one two-line hook change pinned by a failing hook case, and four prose additions, every diff applied and checked in a scratch clone.
**Considered:** `flow-medium` — the hook gates every zcode verifier dispatch, but case 8b and case 9 pin both directions of the pair.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-backlog-flow-3-fix`, Tasks 1, 2, 3 and 4 in that order.
**Considered:** one fixer per task — each task is one `git apply` of a checked diff over disjoint files, so four dispatches buy nothing but four cold starts.

## Open questions
