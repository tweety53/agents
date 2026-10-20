## Context

Found by the self-review pass of `/flow` change `kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`
(a gymie change spanning gymie, gymie-frontend and gymie-playwright; canonical worktree
`/Users/tweety53/Projects/gymie-worktrees/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`).
The agents fixes ride `self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`, based at
`origin/main` (`9adea2ca`), in this worktree. F3 and F6 are gymie project configuration and F5 is
gymie-playwright test harness, outside the agents repository: per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`) they commit on the change's own branch
`spectre/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy` — F3 and F6 in the gymie worktree
`/Users/tweety53/Projects/gymie-worktrees/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`,
F5 in the gymie-playwright worktree
`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`.
Each finding is carried by exactly one task: Task 1 F1, Task 2 F2, Task 3 F4, Task 4 F7, Task 5 F9,
Task 6 F3, Task 7 F6, Task 8 F5. None was disproved; none is `big`.

Facts measured at plan time:

- **F1.** `skills/flow/verify-fix-loop.md` step 3 reads "run one dispatch: `primary` alone" and names
  no key; `skills/flow/review-panel-late-fix.md` runs `primary` "beside **The always-on extra
  dispatch**" and voids the reduction on any Critical or Important. Signal four of
  `check-unfinished-work` accepts only reviewer rows keyed `panel-<round>-<slot>`;
  `flow record next-round` prints a run's first round (`skills/flow/review-panel.md`).
- **F2.** `git ls-tree` of the kan-995 archive commit `3cf7aaaa` lists no `link.md`. The state record
  names three worktrees (gymie `0ff7ff0a`, gymie-frontend `ee12dd85`, gymie-playwright `2c97c418`).
  `scripts/check-unfinished-work.sh` over the gymie worktree with those three ranges exits 0 with an
  `OUTSTANDING` verdict whose only reason is signal four's `Task-Id:` reason — no merge-order reason.
  `changePlanMainCheckout` and `changePlanPeerRoot` (`stats/internal/guard/changeplan.go`) already
  resolve a worktree's main checkout and a `spectre/peers` entry's checkout.
- **F3.** `echo ${GATEWAY_PORT-unset}` in an agent's shell prints `unset`. The kan-995 backend
  worktree's ports are per-app files `.dev-stack/ports/gateway.port` (`9480`) and
  `.dev-stack/ports/frontend.port` (`4400`). From the gymie-playwright worktree, the prefix Task 6
  writes yields `BACKEND_URL=http://localhost:9480` and `FRONTEND_URL=http://localhost:4400`, under
  zsh and bash; from the gymie-playwright main checkout, 8080 and 3000. No guard in agents or gymie
  parses the rows' `${GATEWAY_PORT…}` text.
- **F4/F7.** `skills/flow/visual-verify-verifier.md`'s motion step offers `page.screenshot` "where no
  screencast is available" in a step that mandates a Chromium page, and names one directory per
  motion with no rule for a re-record.
- **F5.** `fidelity/kan-940-cardio-rows.spec.ts:134` is `cardio-R1f`: `openHome(…, 'day_group_header_friends')`,
  `clickTag(page, 'day_group_header_mine')`, then `unfold(page, 'friends', 'cardio_row_<endurance>')`,
  whose tap is `clickTag` → `settledTag` — the helper `dadc7396` ("settle a tag across rendered
  frames") changed. `dadc7396` is no longer on the reshaped branch but its object is present, and
  `git show dadc7396 -- fixtures.ts | git apply -R --check` succeeds at the worktree's HEAD
  (`8d10052`). The kan-995 stack is up: 9480 and 4400 are listening. `playwright.config.ts` runs 5
  workers.
- **F6.** `npx playwright test --help` lists `--output <dir>`; `npx playwright test <spec> --list
  --output test-results/plan-spike` exits 0. `test-results/` is gitignored in gymie-playwright.
- **F9.** Root cause pinned (see the brainstorm record). A hook-denied Agent call's user line
  carries `"toolUseResult": "Error: PreToolUse:Agent hook error: …"`, a string, while
  `rawLine.ToolUseResult` (`stats/internal/harvest/transcript.go`) is `*rawToolUseResult`; the
  line fails to decode and `ParseDispatchEvents` skips it, so no denial is ever produced for a
  hook denial. Replaying the parent transcript's lines 3880–3990 one `RunOnce` per line reproduces
  the stored stamps exactly; with the decode tolerating a string, both denials parse and only
  `verify-flake-1` ← `a0c94877856e38f6d` is stamped, and `go test ./internal/harvest/` passes.
  `stats/internal/harvest/signals.go` decodes the same `rawLine`, so its denial count drops these
  lines too.

## Decisions

### Step 3 names its keys, the extra dispatch and the voiding

**ID:** fix-round-review-keys
**Status:** active
**Chosen:** rewrite step 3 of `verify-fix-loop.md` so the re-review runs `primary` beside **The always-on extra dispatch**, keyed `panel-<round>-primary` and `panel-<round>-test-sanity+call-efficiency` with `<round>` from `flow record next-round` and the round's pass rows recorded, and state in the step that once a Critical or Important is raised every later round opens the decided full roster.
**Considered:** citing `review-panel-late-fix.md` alone and deleting step 3's own description — the step would no longer say what to dispatch, and the run read the step, not the citation; a guard refusing a `review-*` reviewer key — signal four already refuses it at the gate, late, and the defect is the contract's wording.

### A multi-repository change must name every repository in its merge order

**ID:** merge-order-signal
**Status:** active
**Chosen:** signal ten of `check-unfinished-work`, read where signals three and nine are (the worktree holding the plan, `planRC == 0`): for every range whose worktree's main checkout (`changePlanMainCheckout`) differs from the plan worktree's, the plan directory's `link.md` must hold a `## Merge order` section with a numbered entry naming, in backticks, a peer whose `changePlanPeerRoot` resolves to that main checkout. Each repository not covered is one reason, `merge order does not name <main checkout>: <planDir>/link.md — write the change's ## Merge order per skills/flow/cross-repo-worktrees.md ("Then make the merge-order record cover the whole set") before landing`. A single-repository change, or a call with one range, reads nothing.
**Considered:** `kickoff-worktree.sh` writing the record (the pass's option B) — the resolved set grows after kickoff, in **2. Isolate the workspace**, so a kickoff-time write cannot cover a peer added later, while the gate sees the final set; counting entries against ranges — passes a record naming the wrong repository.

### The verify and capture rows read the worktree's ports

**ID:** rows-derive-ports
**Status:** active
**Chosen:** each row opens `GW=$(cat ../../gymie-worktrees/"${PWD##*/}"/.dev-stack/ports/gateway.port 2>/dev/null); FE=$(cat ../../gymie-worktrees/"${PWD##*/}"/.dev-stack/ports/frontend.port 2>/dev/null);` and sets `FRONTEND_URL=http://localhost:${FRONTEND_PORT:-${FE:-3000}} BACKEND_URL=http://localhost:${GATEWAY_PORT:-${GW:-8080}}` — the regression checkout's worktree and the backend worktree share the change's directory name, an exported variable still wins, and the main checkout, which has no such sibling, falls back to 8080/3000.
**Considered:** a new `<backend-root>` placeholder — no agents contract substitutes one into `verify`, so it would need a contract change in agents for a gymie row; `|| echo 8080` — a `|` splits the Markdown table cell; reading `.dev-stack/port-block` — one more parse for the same numbers the per-app files hold.

### Every Playwright run keeps its own artifacts

**ID:** per-run-output-dir
**Status:** active
**Chosen:** `verify` and `capture` pass `--output test-results/$(date -u +%Y%m%dT%H%M%SZ)`, so no later run wipes an earlier run's traces and error contexts; `test-results/` stays gitignored and is removed with the worktree.
**Considered:** a prose rule to copy `test-results/` aside after a non-zero exit — a step the agent forgets, the failure F6 records; `capture` left as is — its failures lose their traces the same way.

### Motion is recorded by screencast only, and a re-record replaces its strip

**ID:** motion-screencast-only
**Status:** active
**Chosen:** the motion step drops the `page.screenshot` fallback and its sampled-clock `<ms>` clause and says never to sample with `page.screenshot` (Task 3); a later round's recording of a motion deletes `motion/<motion id>/` first and writes there, never into a suffixed sibling, and `visual-verification.md` cites an earlier round's strip by its round number (Task 4).
**Considered:** keeping the fallback for a non-Chromium page — the step mandates Chromium, so the clause only ever lowered the sampling rate; a guard refusing `motion/*-r<k>/` at the archive commit — one more guard for a rule one sentence states, where the verifier's own report already lists what it wrote.

### cardio-R1f opens on a trace under load

**ID:** r1f-trace-first
**Status:** active
**Chosen:** Task 8 runs `cardio-R1f` alone at `--repeat-each 20` on the config's 5 workers — the load the full suite applied — first at HEAD, then with `dadc7396`'s `fixtures.ts` hunk reversed. Failing only without it: `dadc7396` already covers the flake; the task commits nothing and its report carries both runs' counts. Failing at HEAD too: the kept `--output` trace is read, the cause fixed at its source in `fixtures.ts` or the spec, and the fix proven by removing it and watching the repeat run fail. Neither run failing: the task grows the load to the full fidelity project once, per `docs/flake-bisection.md`'s bounded exception, before concluding.
**Considered:** a full-suite bisect from the start — gymie's flake brief makes the smallest reproducing subset the first move; closing on the SettledTag hypothesis unmeasured — the pass recorded it as a hypothesis, not a verdict.

### Decoding tolerates a string toolUseResult

**ID:** tool-use-result-string
**Status:** active
**Chosen:** `rawToolUseResult` gains an `UnmarshalJSON` that leaves the struct zero for any non-object value, so a line whose `toolUseResult` is a string decodes; every parser of `rawLine` (`ParseDispatchEvents`, `ParseAgentLaunches`, `ParseCommandRecords`, `ParseAssistantRecords`, `ParseSignalRecords`) gains those lines at once.
**Considered:** `json.RawMessage` for the field and decoding at each reader — every reader repeats the check; matching denials by tool_use_id from a second, lenient decode inside `ParseDispatchEvents` only — leaves the signal count blind to the same lines.

### kan-995's three stamped rows are not rewritten

**ID:** k995-rows-left
**Status:** active
**Chosen:** the fix is forward only; rows 35–37 of kan-995 keep their stamps, and this design and the pass record are where their error is recorded.
**Considered:** re-running `flow record dispatch end -agent-id <right id>` on rows 35–37 — `EndDispatch`'s `COALESCE` would replace the ids, but the token figures the harvester credited from each agent's file stay on the row they were credited to, so each row would name one agent and carry another's tokens; a store verb that re-attributes a historical row's tokens — new store machinery for three rows of one landed change.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 2 adds a signal to the landing gate every run passes, and Task 5 changes how every transcript line is decoded; a slip in either blocks every later run or silently mis-stamps every dispatch.
**Considered:** `flow-low` — each task is small and test-pinned, but two of them sit on every run's path.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** three fixer dispatches in series: `fix-g1` Tasks 1, 3, 4 — run-loaded prose in `skills/flow/`, sharing one `verbatim-moves.txt`; `fix-g2` Tasks 2, 5 — Go in `stats/internal/guard` and `stats/internal/harvest`; `fix-g3` Tasks 6, 7, 8 — the gymie and gymie-playwright worktrees against the running kan-995 stack, in that order, so Task 8's runs use the rows' derived ports and keep their traces.
**Considered:** one fixer for all eight — one context carrying the prose guards, the Go suite and a load-reproduction loop in two other repositories; one per task — eight dispatches, two pairs of them editing the same file in series.

## Open questions

## Live check

Every live step reads; none writes the store. From this worktree, with `stats/` built:

- `scripts/check-unfinished-work.sh` over the kan-995 gymie worktree, with the three ranges of its state record (`<gymie-wt>=0ff7ff0aa5dcf012b85ae71994fb2f9b673699e4`, `<frontend-wt>=ee12dd85d101a8cb06623903b03c27b047cfa2da`, `<playwright-wt>=2c97c418024ef37bfa39008ae23c9b13eed5811a`), prints two `merge order does not name` reasons — one for the gymie-frontend main checkout, one for the gymie-playwright main checkout — beside the `Task-Id:` reason it prints today. Not working looks like: the verdict unchanged from plan time.
- The real parent transcript `~/.claude/projects/-Users-tweety53-Projects-gymie/c135a1ae-4003-4fd0-98e6-5b321b434e4a.jsonl`, lines 3880–3990, fed to `harvest.ParseDispatchEvents` by a throwaway test (deleted afterwards, never committed) yields two denials, at `00:36:03.314Z` and `00:36:11.666Z`. Not working looks like: zero denials, the figure measured at plan time.
- From the kan-995 gymie-playwright worktree, the `verify` row's prefix (its `GW`/`FE` reads and URLs, ending in `env`) prints `BACKEND_URL=http://localhost:9480` and `FRONTEND_URL=http://localhost:4400` while that stack is up; from `/Users/tweety53/Projects/gymie-playwright`, 8080 and 3000. Not working looks like: 8080 from the worktree.
