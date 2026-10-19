# self-review-kan-998-bugfixes-superset-header-focus-mode

## Context

- Six findings of kan-998's self-review pass, none `big`, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm
  design: `.superpowers/sdd/2026-10-19-self-review-kan-998-bugfixes-superset-header-focus-mode-design.md`
  in this worktree. Every question was auto-resolved to its recommended option
  (`skills/flow-contracts/operator-prompts-auto-resolution.md`). No finding is dropped: each was
  confirmed against the tree or a recorded run before it was planned.
- **F-1.1 — Task 1.** Step 3 of `skills/flow/visual-verify.md` names the `regression checkout`
  recipe as "the `kickoff-worktree.sh --from` call, then the regression checkout's `setup`" — the
  merge-order record, which `skills/flow/cross-repo-worktrees.md` requires to name every
  repository of the set and `skills/flow/implement.md` lists as part of that recipe, is left out.
  `peerLogs` (`stats/internal/selfreview/git.go`) reads peers only from the archived `link.md`'s
  `## Merge order`. Blast radius: one file.
- **F-1.2 — Task 2.** The absence-sweep paragraph of `skills/flow/brainstorm-planner.md` scopes
  the `git grep` to "the paths the task touches". kan-998 removed labels gymie-playwright still
  drove, and its regression run failed 83 tests on them. Blast radius: one file.
  <!-- measured: count from the kan-998 self-review finding F-1.2's evidence, a recorded run, not re-runnable at a ref -->
- **F-1.3 — Task 3.** No `scripts/*.sh` is on `PATH`; `~/.local/bin` is, and holds `flow`. A
  symlink into it does not work: the scripts load their siblings from `dirname
  "${BASH_SOURCE[0]}"`, which a symlink resolves to the bin directory — a symlinked
  `project-get.sh` exits with `cannot load lib/flow-guard.sh beside <bin>/project-get.sh`, while a
  two-line `exec` shim answers `merge and push`. Blast radius: two files (`setup.sh` and one
  setuptest file), plus the fingerprint helper.
  <!-- measured: ln -s scripts/project-get.sh <scratch>/bin/ and a printf'd exec shim beside it, each run with <worktree> 'default landing route', in this worktree @ 46db35e0 → the symlink's lib error, the shim's "merge and push" -->
- **F-1.4 — Task 4.** kan-998's `live-verification.md` carries `Before figure: not measured on a
  merge-base stack`, and **Live check** (`skills/flow/verify-and-handoff.md`) requires the before
  figure from a detached merge-base worktree. The rule chosen (decision
  `live-record-before-rule`) passes every archived record that names a before figure and refuses
  kan-998's. Blast radius: three files (guard, shim, test) plus the skill sentence.
  <!-- measured: per archived live-verification.md under agents and gymie whose design.md Live check names a before figure, grep -iw before minus lines matching not measured|unmeasured|not taken, in the main checkouts @ develop and main on 2026-10-19 → every record at least one line, none carrying a marker; kan-998's record (gymie-worktrees/kan-998… @ 6b1e0745) carries one marker line -->
- **F-4.1 — Task 5.** `compose-mockup-frames.sh` exits 1 on a sidecar line no captured PNG
  matches, but only at the verifier's step 10, after the capture. The rule chosen (decision
  `sidecar-reach-literals`) is static: over gymie-playwright it passes every one of 231 sidecar
  lines, and over kan-998's specs with the sidecars as they stood before their fix it names the
  four `pinned-set-strip` lines whose tests were gone. Blast radius: three files plus the skill
  sentence.
  <!-- measured: the rule as a python3 spike (scratchpad sr.py) over ~/Projects/gymie-playwright @ e864161 → 231 lines, 0 misses; over git archive d951d4e fidelity with pinned-set-strip.spec.ts.mockups and quick-mode.spec.ts.mockups taken from d951d4e^ (gymie-playwright-worktrees/kan-998…) → 226 lines, 4 misses -->
- **F-4.2 — Task 6.** Playwright writes a spec's baselines into `<spec>-snapshots/`, and
  `resolve-visual-screenshots.sh` prints them. Byte-identical groups exist in today's
  gymie-playwright too (`kan-346-route-e`, `kan-753-offline-l2`), some intended — so the guard
  names groups for the verifier to explain rather than failing the stage on sight. Blast radius:
  three files plus the skill sentences.
  <!-- measured: shasum of every *.png per *-snapshots dir in ~/Projects/gymie-playwright @ e864161 → 63 dirs, 3 groups (one of 3 in kan-346-route-e, one of 3 and one of 2 in kan-753-offline-l2) -->
- **New guards are Go.** `.flow/project.md` requires every new guard in `flow-guard`
  (`stats/internal/guard/`, registered in `Registry`, Go table tests) behind a thin
  `flow_guard_exec` shim; `run-guard-tests.sh` accepts `stats/internal/guard/<name>_test.go` as the
  companion.
- **Prose acknowledgement.** Every new run-loaded sentence is listed in this change's
  `verbatim-moves.txt`, committed with the plan; `check-verbatim-moves.sh` against `46db35e0`
  with all five prose diffs applied fails on exactly those sentences.
  <!-- measured: scripts/check-verbatim-moves.sh 46db35e0 in a scratch clone @ 46db35e0 with the prose diffs of Tasks 1, 2, 4, 5, 6 applied → 11 FAIL lines, all new-text, each listed in verbatim-moves.txt -->

## Decisions

### Route step 3 to the recipe's merge-order record

**ID:** merge-order-in-visual-verify
**Status:** active
**Chosen:** one sentence in step 3 of `skills/flow/visual-verify.md` stating that the recipe's merge-order record runs with the worktree it adds, its planning commit included, citing **Isolate the workspace — additional worktrees** — the rule lives once, in `cross-repo-worktrees.md`.
**Considered:** restating the extend-or-write rule in step 3 — a second copy of a rule `cross-repo-worktrees.md` states canonically; teaching `peerLogs` to read the state record's `worktrees` map instead — the bundle would then disagree with run 1, which still sequences from `## Merge order` and would never land the omitted repository's branch.

### Extend the absence sweep to the regression checkout

**ID:** sweep-regression-checkout
**Status:** active
**Chosen:** a paragraph in the same blockquote as the absence-sweep rule: a removed UI tag, test id or label a declared `regression checkout` drives is swept there too with the same `-e` list, and the plan carries the task that updates every hit.
**Considered:** sweeping every repository of the worktree set — a backend identifier shares names with unrelated frontend text and would plan edits nobody needs; a guard over the plan — whether a removed string is a label the checkout drives is a planning judgment no file states.

### Exec shims in ~/.local/bin, not symlinks

**ID:** script-shims-local-bin
**Status:** active
**Chosen:** `setup.sh global` writes, for every `scripts/*.sh` but the `test-*.sh` harnesses, a `#!/bin/sh` shim into `~/.local/bin` carrying a marker line and `exec '<abs script path>' "$@"`; each run removes marked shims first and rewrites the current set, so a deleted script's shim goes with it; a same-named file without the marker is left alone and counted SKIPPED. `realHomeFingerprint` gains `~/.local/bin`, so a test leaking a shim into the real home fails.
**Considered:** symlinks through `link_into` — every script loads `lib/` from `dirname "${BASH_SOURCE[0]}"`, which a symlink resolves to the bin dir (spiked: `cannot load lib/flow-guard.sh`); rewriting every bare name in the skills to `<agents repo>/scripts/…` — dozens of call sites across skills, and the next bare name reintroduces the defect; adding `<agents repo>/scripts` to `PATH` in the shell profile — a dotfile edit per shell, and harness sessions that do not source the profile still miss it; linking only the entry scripts the finding named — the skills invoke dozens of distinct names, and any list goes stale.

### Refuse a live record without a measured before figure

**ID:** live-record-before-rule
**Status:** active
**Chosen:** `check-live-verification.sh <record>` fails when no line names a before figure (whole word `before`, any case) or when a line naming one also says `not measured`, `unmeasured` or `not taken`; run by **Live check** after the record is written.
**Considered:** a per-cell rule over a `Before` table column — kan-929's and kan-930's records carry `—` in that column for after-only checks, legitimately; gating on `design.md`'s `## Live check` naming a before — kan-980's one-line no-runtime section contains the word "before", and **Live check** already decides whether the check runs, so the guard runs only where it did.

### A sidecar line is reached by a string or template literal of its spec

**ID:** sidecar-reach-literals
**Status:** active
**Chosen:** `check-sidecar-reach.sh <capture root>` walks every `*.mockups` under the root (pruning `node_modules`, `.git`, `.worktrees`), and each line's screenshot name must equal a `'…'`, `"…"` or `` `…` `` literal of the spec beside it, or fully match a template literal with each `${…}` read as one or more characters; run in step 3's pre-flight.
**Considered:** reading only `toHaveScreenshot(` arguments — five call sites pass a helper's `name` parameter, the literal sitting at the helper's caller; asking Playwright (`--list`) as `check-spec-reach.sh` does — it lists test titles, never screenshot names; checking the baseline PNG exists — a sidecar line the change adds has no baseline until the capture writes it.

### Name identical baselines for the verifier to explain

**ID:** identical-baselines-explained
**Status:** active
**Chosen:** `check-identical-baselines.sh` reads the resolver's PNG paths on stdin, groups byte-identical files per directory, prints one `identical:` line per group and exits 1 when any exists; the verifier's step 9 runs it and its report carries one `- identical:` line per group — an explanation, or the defect, which blocks.
**Considered:** blocking on any group — gymie-playwright's `kan-346-route-e` and `kan-753-offline-l2` hold groups that are intended; comparing across directories — a fresh capture and its own baseline are identical by design; a pixel-tolerance comparison — the defect is a capture that never left the first state, which byte equality catches exactly.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — Task 3 changes what `setup.sh global` writes into the operator's home, and Tasks 4–6 add three guards whose verdicts gate live stages.
**Considered:** flow-low — Tasks 1 and 2 are one paragraph each, but a low-effort pass could under-read the shim lifecycle and the guards' false-positive edges.

### Two fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** two groups in series — group 1, Tasks 1–3 (`self-review-kan-998-bugfixes-superset-header-focus-mode-fix-g1`): the two prose fixes and the setup.sh shims; group 2, Tasks 4–6 (`self-review-kan-998-bugfixes-superset-header-focus-mode-fix-g2`): the three Go guards, which share one package, one shim shape and one test pattern. Task 5 runs after Task 1, both editing `skills/flow/visual-verify.md`.
**Considered:** one group — six tasks, three of them new guards, in one fixer's context; one group per task — six dispatches for work two fixers finish in series.

## Open questions

## Live check

No running service or persistent state is touched: Tasks 4–6 add `flow-guard` subcommands built
per call from the checkout by `scripts/lib/flow-guard.sh`, Task 3 changes what `setup.sh global`
writes and is exercised in the sandboxed setuptest harness, and Tasks 1 and 2 are prose.
