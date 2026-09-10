# kan-484-flow-add-no-delegation-guard-to-implementer

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** every leaf agent the conductor dispatches — implementer, panel slot, panel-fix, verifier — is told, in its own prompt, to do the work itself and never call the `Agent` tool, and `check-dispatch-paragraphs.sh` fails if any of those four sites loses that paragraph.
**Architecture:** one verbatim blockquote added at four sites in three skill files (design.md sections 1–2); one table entry in the existing paragraph guard plus harness cases in its existing assertion harness (sections 3–4). No store, spec, budget or planner-dispatch changes.
**Spec:** this change's own `design.md` (the repo carries no `spectre/specs/` entries; the skill markdown is the contract surface).

Baseline for every `Baseline:` below is the guard-harness count, `ls scripts/test-*.sh | wc -l`,
the suite `scripts/run-guard-tests.sh` discovers: 59 harnesses before task 1, and no task adds
one — task 2 adds cases inside an existing harness.
<!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-484-flow-add-no-delegation-guard-to-implementer, 2026-09-10 -->

Before task 1, `scripts/test-check-dispatch-paragraphs.sh` carries 45 cases and prints `all cases
passed`; `scripts/check-dispatch-paragraphs.sh` prints `DISPATCH-PARAGRAPHS-OK: … — 17 site(s)
validated`.
<!-- measured: grep -c '^# Case [0-9]*:' scripts/test-check-dispatch-paragraphs.sh; scripts/test-check-dispatch-paragraphs.sh | tail -1; scripts/check-dispatch-paragraphs.sh @ branch spectre/kan-484-flow-add-no-delegation-guard-to-implementer, 2026-09-10 -->

**Normative-inventory discipline, task 1:** before the task's first edit run
`scripts/check-normative-inventory.sh > .superpowers/sdd/normative-before-1.txt` (the
`.superpowers/` tree is git-ignored); after its last edit run it again to
`normative-after-1.txt` and check `diff normative-before-1.txt normative-after-1.txt | grep '^<'`
prints nothing — added lines are this change's own new sentences, a removed line is a cut or
reworded existing requirement and is restored before the commit.

**The paragraph, verbatim — every site in task 1 and the `DELEGATION_BLOCK` literal in task 2 carry exactly this text:**

```markdown verified:authored in-tree for this change; design.md section 1 is the source
> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the conductor's closed list (**Dispatch sites — the
> conductor's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.
```

- [x] 1. Carry the NO DELEGATION paragraph at the four leaf dispatch sites
**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`
**Tests:** none — prose only; the verification is the guard set in the verify step
**Regression:** reverting leaves the implementer, panel slot, panel-fix and verifier prompts with no instruction against forking — the kan-30 fix run's implementer grandchild is exactly what a conductor propagating those prompts verbatim would allow again
**Baseline:** before=59 after=59
<!-- predicted: no harness is added by this task; confirmed by ls scripts/test-*.sh | wc -l at the verify step -->

  - [x] **Step 1: implementer dispatch** — in `skills/flow/implement.md` section 4, directly after the implementer's TOOLS blockquote (the one that opens `> **TOOLS:** Every tool you need …`, line 534 at `3bb9b83`) and before the TARGETED TESTS blockquote, insert one blank line and the paragraph from the header, verbatim. Do not add it to the conductor's own section-4 restatement (the FOREGROUND BUILDS / REPRODUCE, DON'T READ pair near line 644) — the conductor is bound by the closed list, not by this paragraph.
  - [x] **Step 2: the closed list's one sentence** — in `skills/flow/implement.md`, **Dispatch sites — the conductor's closed list**, after the paragraph that opens `**The self-check.**` (line 139 at `3bb9b83`), add one paragraph: the four rows are the whole run's dispatch tree, because every row's prompt carries the NO DELEGATION paragraph (section **4** below, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`) — a leaf never dispatches, so nothing exists below these rows. One paragraph, no restated blockquote.
  - [x] **Step 3: panel slot dispatch** — in `skills/flow/review-panel.md`, after the slot TOOLS blockquote (the block under `**Every slot's dispatch prompt also carries the TOOLS paragraph**:`, line 396–401 at `3bb9b83`) and before `**Every slot carries the MODEL HANDSHAKE paragraph**`, add a blank line, the lead-in `**Every slot's dispatch prompt also carries the NO DELEGATION paragraph**:`, a blank line, and the paragraph verbatim.
  - [x] **Step 4: panel-fix dispatch** — in `skills/flow/review-panel.md`, after the fix TOOLS blockquote (the block under `**Every fix subagent's dispatch prompt also carries the TOOLS paragraph**:`, line 859–864 at `3bb9b83`) and before `**Every fix subagent's dispatch prompt also carries the MODEL HANDSHAKE paragraph**`, add the lead-in `**Every fix subagent's dispatch prompt also carries the NO DELEGATION paragraph**:` and the paragraph verbatim, same spacing as step 3.
  - [x] **Step 5: experimental slot enumeration** — in `skills/flow/review-panel.md`, **Experimental slot**, the sentence `carrying the same REPORT FILE / REPRODUCER / CONTEXT BUNDLE / WORKTREES / TOOLS / FOREGROUND BUILDS / MODEL HANDSHAKE / REPRODUCE, DON'T READ paragraphs every slot's dispatch already carries above.` (lines 163–165 at `3bb9b83`): insert `NO DELEGATION / ` immediately after `TOOLS / `. Nothing else in the sentence changes.
  - [x] **Step 6: verifier dispatch** — in `skills/flow/verify-and-handoff.md`, **The verifier dispatch**, after the TOOLS blockquote (under `**The prompt also carries the TOOLS paragraph**:`, lines 132–137 at `3bb9b83`) and before `**The prompt also carries the MODEL HANDSHAKE paragraph**:`, add the lead-in `**The prompt also carries the NO DELEGATION paragraph**:` and the paragraph verbatim.
  - [x] **Step 7: Verify** — from the worktree root, each must exit 0: `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-dispatch-paragraphs.sh` (still `17 site(s) validated` — the guard row is task 3's), `scripts/check-contract-budget.sh` (no row moves: `wc -c` on the three files stays under 52398, 73554 and 42943 respectively, the `budgets()` rows in `scripts/check-contract-budget.sh` at `3bb9b83`), `scripts/check-stage-mark-calls.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`; `grep -c 'NO DELEGATION:\*\*' skills/flow/implement.md skills/flow/review-panel.md skills/flow/verify-and-handoff.md` prints `1`, `2`, `1`; the normative-inventory diff from the header prints nothing; `ls scripts/test-*.sh | wc -l` prints 59.
**Commit:** `docs(flow): carry a NO DELEGATION paragraph at every leaf dispatch site`

- [x] 2. Harness cases for the NO DELEGATION guard entry
**Build:** red
**Squash-with:** Task 3
**Files:** `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `case 46`, `case 47`, `case 48`, `case 49`, `case 50`, `case 51`
**Regression:** reverting removes the six cases, so dropping the `delegation` entry from `scripts/check-dispatch-paragraphs.sh` — or any one of its three phrases, or its min-2 count on `review-panel.md` — leaves the harness green; `case 46`: the label absent from `implement.md` goes unreported; `case 47`, `case 48`, `case 49`: a block missing one shared phrase passes; `case 50`: the label absent from `verify-and-handoff.md` goes unreported; `case 51`: `review-panel.md` with one block instead of two passes
**Baseline:** before=59 after=59
<!-- predicted: no harness file is added — six cases inside the existing harness; confirmed by ls scripts/test-*.sh | wc -l at the verify step -->
**After:** Task 1

  - [x] **Step 1: the literal and the fixtures** — in `scripts/test-check-dispatch-paragraphs.sh`: after `INDEPENDENT_BLOCK=` (line 433 at `3bb9b83`) add `DELEGATION_BLOCK=` holding the header paragraph verbatim, single-quoted with the file's `'"'"'` convention for every apostrophe (`you start`, `conductor's`). In `new_root` (line 145), change the `verify-and-handoff.md` seed from `printf '%s\n\n%s\n' "$TOOLS_BLOCK" "$HANDSHAKE_BLOCK"` to `printf '%s\n\n%s\n\n%s\n' "$TOOLS_BLOCK" "$HANDSHAKE_BLOCK" "$DELEGATION_BLOCK"`; leave the `brainstorm.md` seed unchanged (the planner is not a site). Append `\n\n$DELEGATION_BLOCK` once to case 1's inline `implement.md` list (after its last `$HANDSHAKE_BLOCK`, line 496) and to `CLEAN_IMPLEMENT` (line 1361), and twice — `\n\n$DELEGATION_BLOCK\n\n$DELEGATION_BLOCK` — to case 1's inline `review-panel.md` list (after `$INDEPENDENT_BLOCK`, line 479) and to `CLEAN_REVIEW_PANEL` (line 1339). Extend case 1's comment (lines 444–456) with one sentence naming the NO DELEGATION blocks (one in `implement.md`, two in `review-panel.md`, one seeded in `verify-and-handoff.md`; KAN-484). Every other case asserts exit 1 or 2 on a `case "$OUT" in *…*)` pattern and is unaffected by an additional violation, so no other fixture changes.
  - [x] **Step 2: three phrase-dropped variants** — beside `DELEGATION_BLOCK`, add `DELEGATION_BLOCK_NO_NEVER_CALL_AGENT` (the sentence `Never call the \`Agent\` tool, and never spawn a subagent,` becomes `Do not use the \`Agent\` tool, and never spawn a subagent,`), `DELEGATION_BLOCK_NO_NEVER_SPAWN` (`and never spawn a subagent, background agent or helper` becomes `and start no subagent, background agent or helper`) and `DELEGATION_BLOCK_NO_LEAF` (`you are the leaf of this run` becomes `you are the last agent in this chain`). Each keeps every other phrase intact and stays a plausible paragraph, matching how `VERBATIM_BLOCK_NO_*` and the TOOLS variants (line 389 onward) are built.
  - [x] **Step 3: cases 46–51** — after case 45 and before the `if [ "$FAILURES" -ne 0 ]` footer (line 1684), add six cases in the file's existing shape (`# ===` banner comment, `new_root`, `write_site` calls, `run_guard`, one `[ "$RC" -eq 1 ] && pass … || fail …` line, one `case "$OUT" in` pattern assertion), each using `$CLEAN_REVIEW_PANEL` / `$CLEAN_IMPLEMENT` for whichever file the case is not exercising:
    - `case 46`: `implement.md` written as `$CLEAN_IMPLEMENT` with its trailing `$DELEGATION_BLOCK` removed — write the full list inline, identical to `CLEAN_IMPLEMENT` minus that block; `review-panel.md` `$CLEAN_REVIEW_PANEL`. Assert exit 1 and `*"implement.md"*"NO DELEGATION"*`.
    - `case 47`, `case 48`, `case 49`: `review-panel.md` as `$CLEAN_REVIEW_PANEL` with its second `$DELEGATION_BLOCK` replaced by `$DELEGATION_BLOCK_NO_NEVER_CALL_AGENT`, `$DELEGATION_BLOCK_NO_NEVER_SPAWN`, `$DELEGATION_BLOCK_NO_LEAF` respectively (full inline list each time, first block correct); `implement.md` `$CLEAN_IMPLEMENT`. Assert exit 1 and, respectively, `*"review-panel.md"*'missing the required phrase: "Never call the `Agent` tool"'*`, `*'missing the required phrase: "never spawn a subagent"'*`, `*'missing the required phrase: "the leaf of this run"'*` — mind the quoting: the guard prints the phrase inside double quotes, and the backticks around `Agent` are literal characters in the pattern.
    - `case 50`: `write_site "skills/flow/verify-and-handoff.md" "$TOOLS_BLOCK

$HANDSHAKE_BLOCK"` (overriding `new_root`'s seed, keeping the other two required blocks so this is the only violation); the other two files clean. Assert exit 1 and `*"verify-and-handoff.md"*"NO DELEGATION"*`.
    - `case 51`: `review-panel.md` as `$CLEAN_REVIEW_PANEL` with one `$DELEGATION_BLOCK` removed (exactly one correct block present); `implement.md` `$CLEAN_IMPLEMENT`. Assert exit 1 and `*"review-panel.md"*'requires at least 2 block(s) carrying the label "**NO DELEGATION:**", found 1'*` — the min-blocks message `report_line` prints, the threshold cases 20–21 had to add after review for FOREGROUND BUILDS.
  - [x] **Step 4: header comment** — extend the harness's header comment (after the `Case 45` paragraph, line 98–101) with a `Cases 46-51 cover KAN-484's NO DELEGATION paragraph …` paragraph in the shape of the `Cases 31-34` one: required once in `implement.md`, twice in `review-panel.md`, once in `verify-and-handoff.md`; which fixtures gained the block; what each case drops.
  - [x] **Step 5: run it RED** — `scripts/test-check-dispatch-paragraphs.sh 2>&1 | tail -20`. Expected: `FAIL: case 46: expected exit 1, got rc=0 …` through `FAIL: case 51: …` (the guard has no `delegation` entry yet, so every new fixture is clean to it and exits 0), case 1 still `ok`, footer `6 case(s) failed`, exit 1. Record the tail in the report file. Do not commit — this task squashes into task 3.
**Commit:** `feat(scripts): guard the NO DELEGATION dispatch paragraph`

- [x] 3. The NO DELEGATION entry in check-dispatch-paragraphs.sh
**Build:** green
**Files:** `scripts/check-dispatch-paragraphs.sh`
**Tests:** none — task 2's six cases are this task's tests
**Regression:** reverting drops the `delegation` entry and its four site rows, so a later prose edit trimming the paragraph from any of the four sites passes `scripts/check-dispatch-paragraphs.sh` silently, and task 2's cases 46–51 fail
**Baseline:** before=59 after=59
<!-- predicted: no harness is added by this task; confirmed by ls scripts/test-*.sh | wc -l at the verify step -->
**After:** Task 2

  - [x] **Step 1: the entry** — in `scripts/check-dispatch-paragraphs.sh`, add to each associative array (lines 165–199 at `3bb9b83`), after the `independent` key: `ENTRY_LABEL[delegation]="**NO DELEGATION:**"`; `ENTRY_SHARED_PHRASES[delegation]="Never call the \`Agent\` tool${US}never spawn a subagent${US}the leaf of this run"` (inside the existing double-quoted style — the backticks must be escaped as `\`` so the shell does not run a command); `ENTRY_VARIANTS[delegation]=""`. Append three site rows — one per file, `review-panel.md`'s row carrying min 2 for its two dispatches — to the four parallel arrays, keeping index alignment: `SITE_ENTRY` gains `delegation delegation delegation`, `SITE_PATHS` gains `"skills/flow/implement.md" "skills/flow/review-panel.md" "skills/flow/verify-and-handoff.md"` in that order, `SITE_MIN_BLOCKS` gains `1 2 1`, `SITE_VARIANTS` gains `"" "" ""`. Each array then holds 20 entries; a misaligned array shows up in step 3's run as a wrong site named or a `bad array subscript` error.
  - [x] **Step 2: header comment** — in the header comment: append a `KAN-484 added a ninth required paragraph — NO DELEGATION, which tells every leaf agent the conductor dispatches to do the work itself and never call the Agent tool or spawn a subagent, closing the conductor's closed list one level down — at four sites: the implementer dispatch in implement.md, the panel slot dispatch and the panel-fix subagent dispatch in review-panel.md, and the verifier dispatch in verify-and-handoff.md.` sentence to the history paragraph (after the INDEPENDENT PASSES sentence ending `min 1 block.`, line 53); add three rows to the paragraph table (`**NO DELEGATION:**` at `skills/flow/implement.md` 1, `skills/flow/review-panel.md` 2, `skills/flow/verify-and-handoff.md` 1, `(none)` variants); add a `NO DELEGATION shared phrases (no variants — every block carrying the label must carry all three): "Never call the \`Agent\` tool", "never spawn a subagent", "the leaf of this run". Required once in implement.md (implementer dispatch), twice in review-panel.md (panel slot dispatch, panel-fix subagent dispatch), once in verify-and-handoff.md (verifier dispatch) — the conductor dispatch and the planner dispatch are not sites: the conductor is bound by its closed list and the planner is the parent's child.` paragraph after the INDEPENDENT PASSES phrases paragraph.
  - [x] **Step 3: run it GREEN** — `scripts/test-check-dispatch-paragraphs.sh 2>&1 | tail -3` prints `all cases passed`, exit 0; `scripts/check-dispatch-paragraphs.sh` prints `DISPATCH-PARAGRAPHS-OK: … — 20 site(s) validated` against this worktree (17 before plus the three new rows), exit 0.
  - [x] **Step 4: mutation check, by hand** — temporarily delete the `verify-and-handoff.md` site row's trio (`delegation` from `SITE_ENTRY`, the path, its `1`, its `""`), run the harness: `case 50` must fail; restore. Temporarily change `review-panel.md`'s `2` to `1`: `case 51` must fail; restore. Both restorations confirmed by `git diff --stat scripts/check-dispatch-paragraphs.sh` showing only the intended additions and the harness green again.
  - [x] **Step 5: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-markdown-integrity.py`, `scripts/test-check-dispatch-paragraphs.sh` (`all cases passed`), `scripts/check-dispatch-paragraphs.sh` (`20 site(s) validated`); `grep -c '^# Case [0-9]*:' scripts/test-check-dispatch-paragraphs.sh` prints 51; `ls scripts/test-*.sh | wc -l` prints 59. Commit tasks 2 and 3 together as one commit carrying `Task-Id: 3` and this task's `**Commit:**` subject.
**Commit:** `feat(scripts): guard the NO DELEGATION dispatch paragraph`
