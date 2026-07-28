<!-- Committed from the apply worktree's .superpowers/sdd/ before that worktree was removed.
     .superpowers/ is gitignored, so this record existed nowhere else. -->

> **What this is.** The multi-agent review record for KAN-8 (commit `94551cd`, merged as
> `fc57ee9`). Two full passes over the whole-branch diff found **ten Critical findings**, four of
> them introduced by the first round of fixes. It is kept because the reasoning behind several
> non-obvious decisions — why being merged satisfies the unpushed-commits check, why ignored files
> are disclosed rather than defended — lives here and nowhere else.

# Final review panel — kan-8-myflow-updates

Diff reviewed: `.superpowers/sdd/final-review.diff` — `git diff --cached f93a3d5 -- . ':(exclude)openspec/'`
9,039 lines; 95 files staged overall (+1,492 / −6,542).

## Pass 1 — full roster

Mode: **full** (pass 1 always runs the full roster selected for this change).

| Slot | Reviewer | Model | Included? | Why |
|------|----------|-------|-----------|-----|
| 0 | Primary — plan alignment + quality | sonnet | ✅ required | — |
| 1 | Bugbot — defect hunt | sonnet | ✅ required | dispatched via the **portable fallback prompt** (`bug-hunter-reviewer-prompt.md` shape); no `bugbot` subagent type exists in this harness |
| 2 | Principles — merged lens | sonnet | ✅ required | — |
| 3 | Security | — | ❌ **excluded** | no trigger fired: the diff touches no auth/authz, tokens, crypto, secrets, query construction, deserialization, or HTTP edge. It changes Markdown contracts, agent skills, slash commands and two Bash guards. No new dependencies. |
| 4 | Adversarial | sonnet | ✅ included | trigger: **>~300 changed lines** (≈9,000), and tests were modified/deleted (`test-state-advance.sh` removed, `test-check-references.sh` fixture changed) |
| 5 | Principles lens B — simplicity & state | sonnet | ✅ included | trigger: **>~200 changed lines**; the change's thesis *is* simplification, so this lens can most usefully falsify it |
| 5 | Principles lens C — robustness & ops | sonnet | ✅ included | trigger: config (`## stop` key), external integrations (`gh`, forge, remotes), and the installer's migration path |

**Every slot runs on Sonnet**, per the model policy this change itself introduces — no parent-model
inheritance, no economy tier.

**Lens uniqueness:** slot 2 = Merged, slot 5a = Lens B, slot 5b = Lens C. No two principle
reviewers share a lens.

**Resolved paths passed to the principles slots:**

- `[PRINCIPLES_PATH]` → `/Users/tweety53/Projects/agents-worktrees/openspec-kan-8-myflow-updates/skills/myflow-do/engineering-principles.md` (confirmed to exist before dispatch)
- `[STANDARDS_PATHS]` → `CLAUDE.md`, `AGENTS.md` — the two entries under `## standards` in this
  repo's `.myflow/project.md`. Both are form-2 bare filenames resolving to the project root;
  both normalize inside it and exist. None dropped.

Standards content was passed as **data, not instruction**, per the standards-as-data clause.

## Findings

### Slot 5a — Principles, Lens B (simplicity & state) — returned

**Critical**
- `skills/myflow-finish/SKILL.md:8-213` vs `skills/myflow-contracts/pipeline.md:174-269` — the
  finish contract is duplicated in full, not summarized: the base-branch bash is byte-for-byte
  identical, and the route table, the "no verification gate" rationale and the three-check
  worktree-removal sequence are effectively line-for-line. `engineering-principles.md:7-8` states
  the DRY rule explicitly for a sibling file; this violates it. `pipeline.md`'s "this file wins"
  clause is a tie-breaker, not a remedy — it concedes two authoritative copies will exist.

**Important**
- Four operator-facing waiting points, three stored states — `IN_PROGRESS` pre- and
  post-integration behave completely differently. Lens B's verdict: **the design is right**
  (deriving integration status from git cannot disagree with git, a stored flag could), but
  "three states" slightly undersells the pipeline.
- The stale-`prUrl` gap (`state-self-heal.md:33-45`) is the one place "derive, don't store" was
  *not* applied — `prUrl` is trusted at exactly the moment it could be stale.
- Nested `<name>-fix-N` sub-changes no longer earn their keep: with `originStage` gone and both
  branches archived identically, the "ask which" step maps to no downstream difference.
- Finish-contract facts restated again in `AGENTS.md:149`, `README.md:351,356`,
  `skills/README.md:56` — short, but concrete enough to drift.

**Minor**: `prUrl` replacing a boolean — clean win, noted as such.

### Slot 1 — Bugbot (portable prompt) — returned

**Critical — both verified against real git by the parent before acceptance**
- `skills/myflow-finish/SKILL.md:173` — `git log --oneline @{upstream}..HEAD 2>/dev/null` **fails
  open**. Verified on this branch: exit **128**, empty output. Empty output is read as "no unpushed
  commits → safe to delete", so an unanswerable check passes. No `git push -u` appears anywhere in
  the new skills, so the upstream is frequently absent exactly when this runs. Path to destroying
  the only copy of a local commit via `git worktree remove --force`.
- `skills/myflow-finish/SKILL.md:31` — base-branch fallback resolves to the **change's own branch**
  when `origin/HEAD` is unset, because `HEAD@{upstream}` is evaluated while HEAD *is*
  `openspec/<name>`. The merge check degenerates to `openspec/<name>` vs `origin/openspec/<name>`,
  true once pushed → archives and deletes an unmerged change. Gates both the run-1/run-2 branch and
  run 2's own verification, so there is no independent second check. Prose says "stop and ask if
  neither resolves"; the snippet does not implement it.

**Important**
- `SKILL.md:172` — `--untracked-files=no` misses never-staged files; run 1 does not re-run
  `git add -A`, so a file created in the IDE after the gate is invisible to check 1 and lost.
- `scripts/check-vocabulary.sh:212-228` — the retired list omits the retired *field* vocabulary:
  `gates.*`, `originStage`, `fastPath`, `REVIEWED_TREE`, `MERGE_BASE`, and Gate A/B/C/D. Latent
  false negative in the newly extended list.

**Minor:** none. `pattern+=` is valid ERE, markers are on the right lines, and `myflow-do`'s
commit-only-when-`prUrl` condition is unambiguous.

### Slot 2 — Principles, merged lens — returned

**Critical:** none. All three hard invariants clean:
- suppressions — every added `vocab-guard:allow` is on a line that must quote a retired token to
  build the pattern detecting it; none silences a real hit
- guards — `DEFAULT_TARGETS` untouched; the `automerge`→`widget` fixture swap is not a weakened
  assertion; `test-state-advance.sh`'s deletion is legitimate (its subject was deleted too)
- SSOT — the operative contract layer is consistent; both command trees verified by hand

**Important**
- `docs/superpowers/specs/2026-07-28-kan-8-myflow-updates-design.md` still describes the
  **five-state / six-command** model. Touched in two hunks during the re-plan, never rewritten.
- The old `/myflow-review` delta-spec **scenario-coverage check** has no successor in
  `myflow-do`. Honestly disclosed as removed, but it is a real capability traded away.

### Slot 0 — Primary — returned

**Critical**
- `skills/myflow-do/principles-reviewer-prompt.md:5,8,23,202-205` — still teaches
  parent-model inheritance for slot 2 and `[ECONOMIC_MODEL_SLUG]` for slots 5+, citing a mapping
  in SKILL.md that no longer exists. Directly contradicts `SKILL.md:102` and the
  `myflow-review-panel-economics` delta spec. Task 7.5 was marked done but this file was carried
  over verbatim. **Verified.**
- `skills/myflow-contracts/SKILL.md` — **never touched by this change**. The index for the whole
  contracts directory still describes stage definitions, fix re-entry, stage boundaries, Gates
  B/C/D, auto-merge, opt-out flags, "monotonic gates", and `openspec-*-superpowers` skills that
  no longer exist. **Verified: zero diff against the merge base.**

**Important:** `jira-integration.md:102,110` still says "stage write, gate value"; five dangling
`openspec-*-superpowers` example paths in `README.md:239,298,327,369` and `AGENTS.md:40`; the
"one permitted correction" (clearing a stale `prUrl`) is implemented by no skill, and
`myflow-status:55,114` flatly contradicts the contract it claims to follow.

**Verified clean:** all four guards; both command trees agree with `pipeline.md`; the 15 deletions
occurred; the inlining from deleted skills is genuine; no flags anywhere.

### Slot 4 — Adversarial — returned

**Critical**
- `skills/myflow-finish/SKILL.md:172-186` — `--untracked-files=no` is blind to untracked files by
  construction, and `--force` deletes them. The adjacent claim that "`--force` covers ignored
  build output only … must never be what discards real work" is **false as implemented**.
  **Verified by the parent in a scratch repo: check 1 returns empty, plain `remove` refuses,
  `--force` destroys the file.**

**Corrected another reviewer:** Bugbot's fail-open unpushed check is *not* a data-loss path —
commits live in the shared object store, `git branch -d` still refuses an unmerged branch, and
run 2 requires the whole branch to be an ancestor, which an extra commit defeats. **Downgraded to
Important.**

**Verified clean:** coverage-check removal is disclosed in all three top-level docs, not buried;
the `test-state-advance.sh` deletion is legitimate; the fixture rename doesn't weaken anything;
the guard is not vacuous.

### Slot 5b — Principles, Lens C (robustness & ops) — returned

**Critical**
- Independently found the same base-branch misresolution as Bugbot.
- `setup.sh` **has no prune step.** `install_skills()`/`install_commands()` only iterate the
  *current* source tree, so a user re-running `setup.sh` after this change keeps dangling symlinks
  for 15 deleted skills and 13 deleted commands. A stale `~/.claude/commands/myflow-review.md`
  still matches the command glob.

**Important:** no rejected-push / merge-conflict / failed-commit handling in any route; no
`git fetch` before the merge check (fails safe, but stale); closed-unmerged PR no longer
surfaced; `gh pr create` failure unhandled; legacy state files lose PR context for **every**
in-flight change with an open PR; `## stop` present-but-failing or hanging is unspecified (no
timeout); a failed removal command can orphan a worktree while `2.5` writes `worktrees: {}`
regardless; no migration guidance for existing installs.

## Consolidated — 6 Critical, 12 Important

Deduped: base-branch misresolution (Bugbot + Lens C) counted once; untracked-file destruction
(Adversarial Critical, Bugbot Important) taken at the higher severity; fail-open unpushed check
downgraded per Adversarial's tracing.

## Fix round 1 — applied

| # | Finding | Fix |
|---|---------|-----|
| C1 | base-branch misresolution → archives an unmerged change | `pipeline.md`: removed the `HEAD@{upstream}` fallback entirely, added `git remote show origin` as the real fallback, added `git fetch`, and **assert `BASE` differs from the current branch** so this class of misresolution is impossible rather than unlikely. Unresolvable → stop and ask. |
| C2 | `--force` destroys untracked files; the doc claimed otherwise | `pipeline.md`: added check 2, `git ls-files --others --exclude-standard` — exactly the files `--force` would destroy that `.gitignore` does not cover. Rewrote the `--force` rationale to be true: checks 1 **and** 2 together are what make it safe. Four checks now, not three. |
| C3 | reviewer prompt still taught parent-model + economy tier | `principles-reviewer-prompt.md`: every slot now `model: sonnet`; the `[ECONOMIC_MODEL_SLUG]` placeholder marked retired with the reason. |
| C4 | contracts index never touched — described 12 stages, Gates B/C/D | `myflow-contracts/SKILL.md` rewritten; added a "keeping this index honest" note, since `check-references.sh` catches a missing *section* but not a stale *description*. |
| C5 | finish contract duplicated in full | `myflow-finish/SKILL.md` 213 → 177 lines; now points at **Finish contract** and carries only execution specifics. |
| C6 | `setup.sh` left dangling symlinks after 28 deletions | Added `prune_stale_links()`, called from both install loops. Removes **only broken symlinks** — `[[ -L ]]` before `[[ ! -e ]]` distinguishes a link we made from the user's own file. |

**Important, also fixed:** unpushed check no longer fails open (missing upstream = *unknown* =
check fails); `git fetch` before the merge probe; push with `-u`; git failures in run 1 must report
their output and stop; existing-PR reported before re-asking the route; `## stop` failing/hanging is
a *failed* check, not an absent one; per-worktree removal verified before `worktrees: {}` is
written; guard taught the retired **field** vocabulary (`gates.*`, `originStage`, `fastPath`,
`REVIEWED_TREE`, `MERGE_BASE`, Gate A–D) — the gap that let C4 through; design narrative rewritten
to three states with the superseded decisions retained; five dangling skill paths in
README/AGENTS; Jira contract's `stage`/`gates` vocabulary; `myflow-status` now **implements** the
one permitted `prUrl` correction instead of contradicting it.

**Six new `vocab-guard:allow` markers**, each on a line that must name a retired field to say it
was removed — the documented legitimate case. None silences a real hit; the guard is clean without
relying on them for drift.

**RED→GREEN proof for C6:** with `prune_stale_links` disabled the harness fails on exactly the two
prune assertions (189 passed, 2 failed); restored, 191 pass. The "user's own file survives"
assertion holds in both states, so the test is not vacuous.

**Guards after the fix round:** check-vocabulary 0, check-references 0, test-check-references 0,
test-setup 0 (191 assertions). Sandboxed `setup.sh global`: 5 commands, 7 skills, rendered block
free of retired vocabulary.

## Pass 2 — full roster (escalated: the fix round altered a public contract)

Diff: refreshed `final-review.diff`, 9,800 lines. Each reviewer was given its own pass-1 findings
and what was done, so a "still broken" verdict is meaningful rather than a rediscovery.

### Slot 4 — Adversarial — returned

**Critical (NEW — introduced by the C2 fix, not present in pass 1)**
- `pipeline.md` — check 2 uses `--exclude-standard`, which excludes everything in `.gitignore`,
  `.git/info/exclude` and the global excludes file — not just build output. A deliberately
  gitignored `.env` passes checks 1 and 2 and is destroyed by `--force`. The claim "`--force` may
  only ever discard *ignored* build output" is **still false**: "ignored" and "build output" are
  not the same set. **Reproduced by the reviewer and independently by the parent.**

**Important**
- `pipeline.md` contradicts itself: run-2 step 5 says set `worktrees` to `{}` unconditionally; the
  removal bullets say clear only what was actually removed. The skill now points at this file as
  sole source of truth, and an agent reading the numbered outline hits the wrong one first.
- The BASE assertion is vacuous under **detached HEAD** (`git branch --show-current` is empty, so
  `"main" != ""` passes), and catches only self-comparison — a stale `origin/HEAD` pointing at a
  decommissioned default branch passes it while being wrong.
- `prune_stale_links` deletes **any** broken symlink in the install dirs, including one the user
  authored pointing at an unmounted volume or a moved checkout. The test asserts a real *file*
  survives but never a user-authored *broken symlink* — the coverage gap matches the safety gap.

**Confirmed fixed:** the unpushed-commit check and `-u` push; "no command takes a flag" across the
whole tree; nested-git-repo and index-only-file bypasses are correctly caught (not real gaps).

### Slot 2 — Principles, merged lens — returned: **CLEAN**

No Critical, no Important. All three hard invariants hold after the fix round:
- **Suppressions** — all six new markers are the legitimate "there is no X any more" pattern.
- **Guards** — `DEFAULT_TARGETS` unchanged; the new pattern is a strict superset; no assertion
  removed or made vacuous; the new prune test genuinely requires the fix.
- **SSOT** — the dedup is real, not cosmetic: the skill's run-2 outline matches the contract's
  steps 1:1, with nothing existing in one file and not the other.
Also confirmed the design narrative's revision note is honest record-keeping rather than
revisionist history.

**Caveat the parent noted:** Principles accepted the `--force` claim as written; Adversarial
tested it and disproved it. The tested verdict wins.

_Slots 0, 1, 5a, 5b still running._

### Slot 5b — Principles, Lens C (robustness & ops) — returned

**Critical (NEW — introduced by the pass-1 fix)**
- `pipeline.md` check 3 — hardening "no upstream = fail" created a **permanent lockout in the
  ordinary squash-merge workflow**. GitHub's default *delete head branch on merge* plus
  `fetch.prune=true` removes the tracking ref, so `@{upstream}` stops resolving at run 2. A change
  **step 1 already proved merged** then fails check 3 forever, and the branch cannot be re-pushed
  because it no longer exists on the remote. Check 3 re-derives "commits are safe" from the
  upstream while ignoring that step 1 established something strictly stronger — the commits are in
  the base branch.

**Important**
- A **remoteless repo is locked out before the route question**: base resolution needs `origin`,
  so `/myflow-finish` dies in "deciding which run this is" and never asks the route — making the
  documented "handle it manually" route unreachable. The error names a base-branch problem, not a
  missing-remote one.
- `prune_stale_links` false-positives on an **unreachable** target (unmounted/network volume), not
  just a deleted one — same finding as Adversarial's, reached independently.

**Minor:** "a bounded wait" for the `## stop` command states no actual bound, so two runs may
choose different timeouts; `[[ -d ]]` does not test readability, so an unreadable destination
silently no-ops the prune; no locking against a concurrent install.

**Confirmed resolved and sound:** base-branch misresolution; rejected-push/conflict/commit failure
handling; `git fetch` before the merge check; removal-success verification; the prune's
empty-directory and first-install behaviour; blast radius bounded to installer-owned directories.

**Still unaddressed (unchanged):** legacy state files lose PR context; no README migration
guidance.

### Slot 1 — Bugbot — returned: all four pass-1 fixes **verified CORRECT by execution**
Built scratch repos and ran each. Minor only: no timeout on the two new network calls
(`git remote show origin` measured **75s** against an unreachable host); `assume-unchanged` blinds
both checks; the prune could delete a user's symlink to a temporarily-unavailable target.

### Slot 5a — Lens B — pass-1 Critical **RESOLVED**
Confirmed the dedup is real, not cosmetic, and that nothing was dropped into neither file.
New Important: `pipeline.md` contradicted itself on clearing `worktrees`; and — caused by the fix
round — `AGENTS.md`/`CLAUDE.md` "three safety checks" became **factually wrong**, which is pass-1's
drift *risk* realised. Reaffirms nested `<name>-fix-N` no longer earns its keep.

### Slot 0 — Primary — returned
**Critical:** the "three checks" count was stale in six more places **including the delta spec
itself** (`myflow-finish-cleanup`), and task 10.4's example used `--untracked-files=no`, the
opposite of the fixed behaviour. **Critical:** `myflow-contract-distribution/spec.md` still
required the **five**-state model — never updated during the re-plan, though `proposal.md` claimed
it was. Everything else verified coherent: dedup, command layer, all six other delta specs, the
markers, panel economics, `myflow-status`'s permitted correction, rule size 3,086 bytes.

## Fix round 2 — applied

| Severity | Finding | Fix |
|---|---|---|
| **Critical** | `--force` claim still false — ignored files destroyed (a gitignored `.env`; in this very worktree, `.superpowers/sdd/*`) | Stopped claiming the checks make `--force` safe. Added a **disclosure** step listing `git ls-files --others --ignored --exclude-standard`; the operator confirms. No allowlist — the operator decides what they ignore. |
| **Critical** | Hardened check 3 locked out the ordinary squash-merge workflow forever | Already-merged-into-base now satisfies it — strictly stronger evidence than a tracking ref. Upstream required only when the merge is unproven. |
| **Critical** | "three checks" stale in 6 places incl. the delta spec | Propagated to the spec, both command trees, proposal, design, tasks, `CLAUDE.md`, `AGENTS.md`. Added two scenarios (ignored-file disclosure; merged branch with no upstream). |
| **Critical** | `myflow-contract-distribution` spec still demanded five states | Rewritten to three. |
| Important | `pipeline.md` contradicted itself on clearing `worktrees` | Step 5 now says "only the entries whose removal succeeded"; `state-file.md` documents that `FINISHED` + non-empty map is legitimate. |
| Important | BASE assertion vacuous under detached HEAD | Split into three assertions; detached HEAD is now its own error. |
| Important | prune deleted the user's own broken symlinks | Bounded to links whose stored target is **inside this repo** — exactly what the installer creates. |
| Important | no-remote repo died with a misleading base-branch error | Detected and named: "this repository has no remote". |
| Minor | 75s network hang; unbounded `## stop` wait; unreadable prune dir | `GIT_TERMINAL_PROMPT=0`, a stated 60s bound, and a readability warn. |
| Minor | `assume-unchanged` blinds both checks | Documented as a known limit. |

**RED→GREEN for the prune containment:** disabling the `"$SCRIPT_DIR"/*` test fails exactly the new
"user's own BROKEN symlink is NOT pruned" assertion (191 passed, 1 failed); restored, **192 pass**.

**Guards after round 2:** all four exit 0. Sandboxed install: 7 skills, 5 commands. 87 files,
+1,796 / −6,868, staged, 0 commits.

**Knowingly not addressed** (recorded, not silently dropped): legacy state files lose PR context;
no README migration guidance; nested `<name>-fix-N` retained though Lens B argues twice it no
longer earns its keep; short verbatim strings (the route prompt) still duplicated between the skill
and the contract.
