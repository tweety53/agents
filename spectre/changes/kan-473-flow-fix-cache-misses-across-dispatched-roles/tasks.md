# kan-473-flow-fix-cache-misses-across-dispatched-roles

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Four tasks: one script, two prose edits, one guard row with its harness cases. `design.md` is
canonical for every decision and for the TOOLS paragraph's wording; a task below names the lines it
changes and what the change must say, never a second copy of prose that `design.md` already carries.

**Baseline, measured before any edit:**

- Every guard in `.flow/project.md`'s `## lint` that scans owned Markdown or plans exits 0 on the
  clean branch: `check-vocabulary.sh`, `check-references.sh`, `check-markdown-integrity.py`,
  `check-installed-citations.sh`, `check-stage-mark-calls.sh`, `check-contract-budget.sh` (71 files
  within budget), `check-dispatch-paragraphs.sh` (8 sites validated).
  <!-- measured: each script run from the worktree root @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->
- `scripts/test-check-dispatch-paragraphs.sh` passes 30 cases (64 `ok:` lines); the `## test`
  runner `scripts/run-guard-tests.sh` discovers 55 `scripts/test-*.sh` harnesses.
  <!-- measured: scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' and ls scripts/test-*.sh | wc -l @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->
- Byte sizes against `scripts/check-contract-budget.sh`'s `budgets()` rows: `skills/flow/implement.md`
  31,948 of 32,500; `skills/flow/review-panel.md` 50,049 of 58,732; `skills/flow/brainstorm.md`
  11,908 of 35,015; `skills/flow/verify-and-handoff.md` 31,337 of 33,199.
  <!-- measured: wc -c on each file and the budgets() rows at scripts/check-contract-budget.sh:190-198 @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->
- Line anchors on the clean branch, all in `skills/flow/`: `implement.md` 46–49 (the conductor relay
  contract's never-in-flight sentence), 407–411 (the implementer's FOREGROUND BUILDS block), 492–512
  (**Never end a turn with a child in flight** and **Turn discipline** with the `seq 1 110` loop at
  502); `review-panel.md` 313–317 (the slot's FOREGROUND BUILDS block), 737–741 (the fix subagent's);
  `brainstorm.md` 112–126 (the planner dispatch prompt and relay contract); `verify-and-handoff.md`
  84–92 (the verifier's baseline pointer and relay contract). `scripts/check-dispatch-paragraphs.sh`
  136–175 (the entry and site tables); `scripts/test-check-dispatch-paragraphs.sh` 256 (case 1's
  fixtures) and 1061 (last line, before which every case ends).
  <!-- measured: sed -n on each file @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` in its verify
step.** Task 3 is the one expected to trip it, on `implement.md`, and raises that row in its own
commit; no other row moves.

---

- [x] 1. Add `scripts/cache-misses.py`, the transcript analysis

Create `scripts/cache-misses.py`, Python 3 standard library only, executable, with the module
docstring stating what it reports and why (the KAN-473 finding: a zero-cache-read turn is a full
re-price, and this lists each one with what preceded it). Behaviour, per `design.md`'s
"`scripts/cache-misses.py`" section:

- Arguments: one or more paths. A `.jsonl` file is one transcript, labelled by its basename; a
  directory expands to its `*.jsonl` files sorted by name, each labelled by the `description` field
  of the sibling `<stem>.meta.json` when that file exists, else the basename.
- Per transcript, walk the entries in file order. Track the most recent tool names since the last
  usage-bearing assistant row (every `tool_use` block's `name` in assistant entries). For each
  entry with `type == "assistant"` whose `message.usage` exists, deduplicate on
  `(timestamp[:19], usage.input_tokens, usage.cache_read_input_tokens)` — a streamed message
  writes several rows carrying identical usage — and count the survivors as turns. A turn whose
  `cache_read_input_tokens` is 0 (or absent) and `input_tokens` > 5000 prints one line:

```text verified:authored in-tree for this change; the shape is the ad-hoc analysis design.md's table was read from
turn#<n> <HH:MM:SS> gap=<seconds, one decimal>s in=<input_tokens> prev_user=<text|tool_result> tools_since_last_usage=[<names>]
```

  `gap` is the seconds since the previous surviving turn (0.0 for the first). `prev_user` is
  `text` when the nearest earlier `type == "user"` entry's `message.content` is a string or a list
  whose first block is `type == "text"`, else `tool_result`. Timestamps are parsed with
  `datetime.datetime.fromisoformat` after replacing a trailing `Z` with `+00:00`.
- Each transcript is preceded by a `===== <label>` line. A file that cannot be parsed as JSONL
  (any line failing `json.loads`, or a missing file) is named on stderr and skipped; the exit
  code is 0 always — this is a report, not a guard (`design.md`'s `cache-misses-report-not-guard`).
- `-h`/`--help` through `argparse`, no other options.

  - [ ] **Step 1: Write the script** as specified, `chmod +x scripts/cache-misses.py`, shebang
    `#!/usr/bin/env python3`.
  - [ ] **Step 2: Run it on the KAN-445 session** —
    `scripts/cache-misses.py ~/.claude/projects/-Users-tweety53-Projects-agents/a7273417-c46a-4128-9f73-619e4a99c84d.jsonl ~/.claude/projects/-Users-tweety53-Projects-agents/a7273417-c46a-4128-9f73-619e4a99c84d/subagents`
    — and confirm the output reproduces `design.md`'s table: 18 lines in all, among them
    <!-- predicted: the run in this step; the ad-hoc analysis that produced design.md's table printed these 18 rows -->
    `turn#23 11:58:32 … tools_since_last_usage=['ToolSearch']` under "Conduct kan-445
    implementation" and `turn#8 12:28:29 … ['ToolSearch']` under "Panel slot: principles".
  - [ ] **Step 3: Run it on a non-transcript** — `scripts/cache-misses.py scripts/README.md` if
    present, else `scripts/cache-misses.py CLAUDE.md` — and confirm one stderr line naming the file
    and exit 0.
  - [ ] **Step 4: Verify** — from the worktree root run `scripts/check-vocabulary.sh` and
    `scripts/check-references.sh` (the `scripts` target is in both guards' scan); both exit 0.

**Files:** `scripts/cache-misses.py`
**Tests:** none — a report script whose check is the KAN-445 session run in Step 2; a synthetic
fixture would be the hand-built value REPRODUCE, DON'T READ warns against (`design.md`'s
`cache-misses-report-not-guard`)
**Regression:** reverting this commit removes the only repeatable way to tell, on the next `/flow`
run's transcripts, whether tasks 2–4 removed the schema-load and idle child-wait misses.
**Baseline:** before=0 after=0 — no harness changes; `scripts/run-guard-tests.sh` still discovers 55
<!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->
**Commit:** `feat(scripts): add cache-misses.py, the zero-cache-read turn report`
**Build:** green
**After:** none

- [x] 2. Bound the wait loop under the prompt-cache TTL and add the parent `continue` backstop

Edit `skills/flow/implement.md` in three places, per `design.md`'s two `implement.md` sections.

**Turn discipline** (lines 495–512). In the fenced loop at line 502, `seq 1 110` becomes
`seq 1 48`. In the paragraph after the fence, the sentence "The loop is bounded under the Bash
tool's ten-minute cap" becomes one stating that the loop is bounded at 240 s, under the prompt-cache
TTL rather than the Bash tool's cap: a wait longer than the TTL re-prices the whole context on
return, while a bounded wait's `still-running` turn reads it at the cache rate and keeps it warm.
The rest of the paragraph — `still-running` re-issues the wait, the ceiling is tracked across
calls, the two other files state their own batches — stays.

**The relay contract** (lines 46–49). After "it waits for every implementer, reviewer, slot and fix
subagent it launched first", add the cost in one sentence: a turn that ends with a child running
idles this role and the parent until the child finishes, and both re-price their whole context on
resume.

**The parent backstop.** After the three-block list (after line 59, before "The first line of its
first reply"), add one paragraph: a conductor return that names a child in flight — `awaiting`,
`in flight`, a dispatch key with no `## Stage` end mark behind it — is not one of the three blocks;
the parent resumes it with `continue` in its very next action and never waits on a grandchild
itself.

  - [ ] **Step 1: Change the loop bound and its sentence** under **Turn discipline**.
  - [ ] **Step 2: Add the cost sentence** to the relay contract.
  - [ ] **Step 3: Add the backstop paragraph** after the three-block list.
  - [ ] **Step 4: Verify** — from the worktree root run `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-stage-mark-calls.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh` and
    `scripts/check-contract-budget.sh`; all exit 0, and `grep -n 'seq 1 110' skills/flow/` prints
    nothing. `implement.md` is expected to stay under 32,500 bytes here — the added prose is under
    550 bytes; if the budget guard trips, cut nothing and raise the row as task 3 describes, in
    this task's commit instead.
    <!-- predicted: wc -c skills/flow/implement.md after this task's edit -->

**Files:** `skills/flow/implement.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — prose edit; the verify step's guard scripts are the check
**Regression:** reverting this commit restores a 9m10s silent wait per call, longer than any
prompt-cache TTL, so every child that runs past five minutes costs the conductor one full re-price
of its context on the wait's return — and leaves the parent free to idle on a conductor's
"awaiting" return, the 13–16 minute gaps design.md's table records.
<!-- measured: the idle child-wait rows of design.md's table @ machine-local transcript, cannot be re-run from a ref -->
**Baseline:** before=0 after=0 — no harness changes; the guards above exit 0 before and after
<!-- measured: the guard runs recorded in the baseline block above @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->
**Commit:** `fix(flow): bound the child wait under the prompt-cache TTL and resume an awaiting conductor`
**Build:** green

- [x] 3. Carry the TOOLS paragraph at the six dispatch sites and raise `implement.md`'s budget row

Add the TOOLS blockquote — its wording is `design.md`'s "The TOOLS paragraph" section, reproduced
byte-for-byte at every site, wrapped at the surrounding file's column — at six sites, each
introduced in the shape its neighbours use:

- `skills/flow/implement.md`, the conductor dispatch: after the relay contract's three-block list
  and task 2's backstop paragraph, a paragraph "The prompt also carries the TOOLS paragraph:"
  followed by the block.
- `skills/flow/implement.md`, the implementer dispatch: directly after the FOREGROUND BUILDS block
  under "Every implementer dispatch **must** also carry:" (lines 407–411 on the clean branch).
- `skills/flow/review-panel.md`, the panel slot: after the slot's FOREGROUND BUILDS block (313–317),
  introduced "**Every slot's dispatch prompt also carries the TOOLS paragraph**:".
- `skills/flow/review-panel.md`, the fix subagent: after its FOREGROUND BUILDS block (737–741),
  introduced "**Every fix subagent's dispatch prompt also carries the TOOLS paragraph**:".
- `skills/flow/brainstorm.md`, the planner: after the relay-contract paragraph (123–126),
  introduced "**The prompt also carries the TOOLS paragraph**:".
- `skills/flow/verify-and-handoff.md`, the verifier: after the relay-contract paragraph (86–92),
  introduced the same way.

Then measure `wc -c skills/flow/implement.md` and set its row at
`scripts/check-contract-budget.sh:193` to that size times 1.25, rounded down to the byte — the
header's own definition of a row. No other row changes: the three other files stay within theirs
by more than the ~450 bytes a block adds.
<!-- predicted: wc -c on each edited file after this task; the block is ~420 bytes plus its introducing line -->

  - [ ] **Step 1: Add the six blocks**, one per site above, byte-identical text.
  - [ ] **Step 2: Confirm the count** — `grep -c '^> \*\*TOOLS:\*\*' skills/flow/implement.md
    skills/flow/review-panel.md skills/flow/brainstorm.md skills/flow/verify-and-handoff.md` prints
    `2`, `2`, `1`, `1`.
  - [ ] **Step 3: Raise the budget row** for `skills/flow/implement.md` as measured.
  - [ ] **Step 4: Verify** — from the worktree root run `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-stage-mark-calls.sh`,
    `scripts/check-dispatch-paragraphs.sh` (still 8 sites — task 4 registers the new ones),
    `scripts/check-installed-citations.sh` and `scripts/check-contract-budget.sh`; all exit 0.

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`, `skills/flow/brainstorm.md`,
`skills/flow/verify-and-handoff.md`, `scripts/check-contract-budget.sh`
**Tests:** none — prose edit plus one budget row; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves every dispatched role free to `ToolSearch "."` mid-run,
the schema load that re-priced 67,917 and 23,490 tokens in the KAN-445 run — and task 4's guard,
once it lands, then fails on all four files.
<!-- measured: the schema-load rows of design.md's table @ machine-local transcript, cannot be re-run from a ref -->
**Baseline:** before=0 after=0 — no harness changes; the guards above exit 0 before and after
<!-- measured: the guard runs recorded in the baseline block above @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles -->
**Commit:** `feat(flow): carry a TOOLS paragraph on every dispatch prompt`
**Build:** green

- [x] 4. Register the TOOLS paragraph in `scripts/check-dispatch-paragraphs.sh` and its harness

Extend the guard's tables (`scripts/check-dispatch-paragraphs.sh` lines 136–175) and its harness
(`scripts/test-check-dispatch-paragraphs.sh`), following the MUTATION PROOF entry KAN-464 added as
the pattern.

**The guard.** Add entry key `tools`: `ENTRY_LABEL[tools]="**TOOLS:**"`, shared phrases
`in your first turn`, `never a wildcard query`, `re-prices your whole context` (US-joined like the
others), `ENTRY_VARIANTS[tools]=""`. Append four sites to the parallel arrays in order —
`skills/flow/implement.md` min 2, `skills/flow/review-panel.md` min 2, `skills/flow/brainstorm.md`
min 1, `skills/flow/verify-and-handoff.md` min 1 — each with an empty variant string. Extend the
header comment's paragraph table and the shared-phrases list with the TOOLS rows, naming KAN-473 as
the change that added the sixth paragraph and the two new site files.

**The harness.** Add `TOOLS_BLOCK`, the paragraph verbatim from `design.md`, and three variants each
dropping exactly one shared phrase while staying a plausible paragraph:
`TOOLS_BLOCK_NO_FIRST_TURN` ("in your first turn" → "before your first tool call"),
`TOOLS_BLOCK_NO_WILDCARD` ("never a wildcard query" → "never a broad query"),
`TOOLS_BLOCK_NO_REPRICES` ("re-prices your whole context" → "costs your whole context again").
`new_root` and every fixture asserted clean (case 1 and every case whose site files are written
correct) gain `skills/flow/brainstorm.md` and `skills/flow/verify-and-handoff.md` carrying one
`TOOLS_BLOCK` each, and two `TOOLS_BLOCK`s appended to each of `implement.md` and `review-panel.md`
— every existing case must stay green, since the guard now hard-fails (exit 2) on a missing site
file. Then four new cases after case 30:

- case 31: the TOOLS label absent from `skills/flow/brainstorm.md` (its file written with prose and
  no block) — exit 1, output names `brainstorm.md` and `TOOLS`;
- cases 32–34: `implement.md` carrying one correct `TOOLS_BLOCK` plus one variant, one variant per
  case — exit 1, output names the dropped phrase.

  - [ ] **Step 1: Write the harness changes first** — `TOOLS_BLOCK`, the three variants, the
    fixture additions, cases 31–34.
  - [ ] **Step 2: Run it to see RED** — `scripts/test-check-dispatch-paragraphs.sh 2>&1 | tail -8`;
    cases 31–34 fail with `expected exit 1, got rc=0` (the unregistered label is invisible to the
    guard) and the final line reports failures.
    <!-- predicted: the harness run in this step, before the guard's tables change -->
  - [ ] **Step 3: Add the entry and the four sites** to the guard's tables, and the header rows.
  - [ ] **Step 4: Run it to see GREEN** — `scripts/test-check-dispatch-paragraphs.sh 2>&1 | tail -3`
    ends `all cases passed`, and `scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:'` prints
    72 (64 before, plus two `ok:` lines per new case).
    <!-- predicted: the harness run in this step; every case above emits an exit-code ok and a names-the-element ok -->
  - [ ] **Step 5: Verify** — from the worktree root run `scripts/check-dispatch-paragraphs.sh`
    (ends `12 site(s) validated`), `scripts/check-vocabulary.sh`, `scripts/check-references.sh`
    and `scripts/check-markdown-integrity.py`; all exit 0.
    <!-- predicted: 8 sites on the clean branch plus the four this task appends -->

**Files:** `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `case 31`, `case 32`, `case 33`, `case 34` in `scripts/test-check-dispatch-paragraphs.sh`
**Regression:** reverting this commit leaves the TOOLS paragraph unguarded — a later trim of any
of its three phrases, or of a whole block, passes `check-dispatch-paragraphs.sh` clean, the silent
failure the guard's header names as its reason to exist; cases 31–34 are what fail if the entry
or the sites are dropped from the tables.
**Baseline:** before=30 after=34
<!-- measured: before — the case count of scripts/test-check-dispatch-paragraphs.sh @ branch spectre/kan-473-flow-fix-cache-misses-across-dispatched-roles; after — predicted by the four cases this task adds -->
**Commit:** `feat(scripts): guard the TOOLS dispatch paragraph at its six sites`
**Build:** green
