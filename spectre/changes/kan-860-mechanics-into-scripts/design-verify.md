# Design — KAN-860, group `verify` (WORK IN PROGRESS — no code landed)

Stopped on the coordinator's instruction (usage limits) before any code or prompt edit. This file
records the investigation and the chosen designs so a follow-up run can implement them directly.
**Nothing below is implemented; no prompt file changed; no `verbatim-moves.txt` lines exist.**

Prompt bytes (unchanged): `skills/flow/visual-verify.md` 17,655; `skills/flow/verify-and-handoff.md`
26,082; `skills/flow-contracts/workspace-isolation.md` 22,723 (rationale 13,894).

## VH-25 — `check-visual-trigger.sh` distinct exit-2 token (designed, not implemented)

- Go: `stats/internal/guard/visualtrigger.go`, `visualTrigger()`. Keep every exit code and every
  existing stderr line byte-identical; **append one token line to stderr on every exit 2**:
  - `VISUAL-TRIGGER-NOT-CONFIGURED: <root> — no visual verification section` for the two
    "not configured" cases (no `.flow/project.md`; file declares no `## visual verification`).
  - `VISUAL-TRIGGER-CANNOT-ANSWER: <root> — <short cause>` for every other exit 2 (root not a dir,
    not a regular file, unreadable, read failure, duplicate section, `ui paths` absent/empty or no
    usable glob, stdin unreadable, usage).
- Tests: `check_visual_trigger_test.go` pins stderr whole — every exit-2 case's `err:` gains the
  token line (cases 1–2 NOT-CONFIGURED, the rest CANNOT-ANSWER). `visualverifydispatched.go`
  calls `visualTrigger` with `io.Discard` stderr, so it is unaffected.
- Shim header (`scripts/check-visual-trigger.sh`): add the two tokens to the exit-contract block.
- Prompt (`verify-and-handoff.md` steps 1–2): step 1 stops reading project.md by hand; the one
  guard call serves both steps — exit 2 + `VISUAL-TRIGGER-NOT-CONFIGURED:` → `Visual: not
  configured`, skip; exit 1 → `Visual: no UI paths touched`; exit 2 + `VISUAL-TRIGGER-CANNOT-ANSWER:`
  → report stderr, skip. Keep step numbering (visual-verify.md cites "steps 1 and 2", "steps 1–3").

## VH-24 — `flow record handoff-lines` (designed, not implemented)

- Verb in `stats/cmd/flow/record.go` (+ `main.go` usage line), tests in `record_test.go` using
  `gitRepo`/`isolatedStateRoot`/`httptest` fakes like `TestCostStatusPrintsOneLine`.
- Call: `flow record handoff-lines -change <name> -C <canonical-worktree> [-worktree <abs-worktree> ...]`
  (`-worktree` repeatable for the other affected worktrees). Worktrees are **deduplicated by
  `fallback.ProjectKey`** — the journal and findings are keyed per project, so two worktrees of one
  project must not double-count (the prose's "one call per affected worktree" never said how to
  combine; this is the chosen aggregation).
- Output, stdout, always exit 0 (caller mistake → exit 2, like every record verb), block-ready:
  ```
  **Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
  **Deferred:** <count> | unknown — the findings could not be read
  **Costs:** <formatCostStatusLine or "unknown">    (canonical worktree only)
  ### Deferred minors
  F<n> <location> — <note> — <reason>   (one per finding whose status starts with "deferred";
                                          reason = status minus "deferred" and one space)
  none                                   (when the count is 0)
  ```
  Records = sum of `countRecordJournalEntries` over the deduped keys; any uncountable → unknown.
  Deferred = findings from `GetRunRecord` per key (404 = none); any read failure → the `unknown`
  spellings above (new vocabulary: the prose had no answer for a failed findings read — note for
  review). Diagnostics go to stderr.
- Refactor: extract the count/line/findings helpers from `runRecordJournalCount`,
  `runRecordCostStatus`, `runRecordFindings` so both paths share them (parity by construction), and
  table-test the formatter against the prose's three Records spellings, the Deferred filter
  (`deferred` prefix only; `open`/`fixed`/`withdrawn …` excluded) and the `none` case.
- Prompt: the three call blocks + the filter/format paragraph in `verify-and-handoff.md`
  ("**Produce the handoff's `Records:` count** …" through "… reads `none` when the count is `0`.")
  become one call line plus "exits 0 always; render each line exactly as printed".
  `skills/flow-contracts/handoff-blocks.md:137` also cites `journal-count` (not this group's file).

## VV-18 — `check-visual-preflight.sh` (designed, not implemented)

- Go `stats/internal/guard/visualpreflight.go` registered `check-visual-preflight`; shim
  `scripts/check-visual-preflight.sh` (`flow_guard_exec check-visual-preflight 2 …`); setup.sh
  symlink + guard-symlink rules; tests `check_visual_preflight_test.go`.
- Call: `prepare-workspace.sh <worktree> | check-visual-preflight.sh <worktree> <app-root>=<resolved-url> ...`
  — stdin the `KEY=value` lines, one argument per matched app (its package root, relative to the
  worktree, and its worktree-resolved URL, both already resolved by the parent for the verifier
  prompt). Exit 0 all pass; 1 at least one `FAIL:` line (stage ends `-outcome stopped`); 2 cannot
  answer (lsof missing, unparsable section) — treat as a failure.
- Check 1 ports: ports = `port` rows' exported values ∪ each app URL's port. `lsof -nP
  -iTCP:<port> -sTCP:LISTEN` (exit 0 + output = held; exit 1 = free; other = exit 2). Held → probe:
  HTTP GET the app URL (any response within 5 s answers) when the port came through one, else a TCP
  dial to 127.0.0.1/::1. Unanswered → FAIL naming port, lsof lines, probe. Commands resolved via
  `lookPath(env, …)` (changeplan.go) so tests inject fake `lsof`/`node` through `Env.Getenv("PATH")`.
- Check 4 Playwright: `node -e "console.log(require.resolve('@playwright/test/package.json',
  {paths: [root]}))"`; the result outside `EvalSymlinks(worktree)` → FAIL naming the path. Not
  resolvable / node absent → an info line, pass (fresh worktrees lack node_modules; `setup`
  installs).
- Checks 2–3 are where behaviour could drift from the prose's judgment. Chosen mechanical rules:
  - base-URL: for each isolation row whose exported value ≠ Default, scan non-Markdown text files
    under each app root (skip `.git node_modules dist build coverage test-results
    playwright-report testdata vendor .superpowers`, binaries, >1 MiB) for the Default literal
    (`port` rows: `:<default>` not followed by a digit — a bare number matches comments like
    playwright.config.ts's "daemon on 4173"). A hit line naming the row's Variable is overridden;
    a row whose `start` command names the Variable or the resolved value is overridden. Else FAIL
    `file:line`.
  - origins: configuration-shaped files (`.env*`, `*.json`, `*.y*ml`, `*.toml`, `*.ini`,
    `*.properties`, `*.conf`, `*.config.*`) in the worktree; a line (or bracketed list) naming an
    `allowed[_-]?origins` key and at least one URL must include every app's resolved origin, else
    FAIL naming the file. Source files are excluded so guard tests/fixtures do not self-trigger.
  - Measured against this repo: no hit (only `stats/web/playwright.config.ts` carries 4173/4174,
    and not as `:4173`).
- Prompt: step 3's four bullets become the call line + exit contract; the "Any failing check ends
  the stage here" paragraph stays (judgment).

## Rows left

- **VH-26** (stop/run resolution, `Running:` extraction) — low confidence; the protected-service and
  refused-start rules would have to be encoded exactly. Not attempted, per the task.
- **WI-12** (a project-declared `claim` verb) — low confidence and a design change to the
  `## workspace isolation` command table. Not attempted, per the task.
- **VV-18, VH-24, VH-25** — designed above, **not implemented** (stopped for usage limits).
