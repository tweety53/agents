> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

# kan-516 — log unprefixed fixes in the store as `/flow <name>` fix runs

KAN-516 · implementation plan.

**Goal:** a problem report typed as a plain message, in a session whose last `/flow <name>` run
left the change at `IN_PROGRESS`, runs as the fix run `/flow <name> <message>` would — planning
pass, implementation, panel, verify, handoff — under a fresh session token, so the store records
it as one more fix run. A `UserPromptSubmit` hook names the change on every plain prompt from the
transcript's last `/flow` mark; the `/flow` skill carries the rule. `docs/research/kan-516.md`
(sections 2–5) is canonical for every decision a task implements until `/flow` folds it into this
change's `design.md`.

Three tasks, one repository. No spec, no migration.

**Baseline, measured before any edit, on `main` at `ad45fa5`:** `test-setup.sh` carries 146
assertions.
<!-- measured: grep -c '^assert_' scripts/test-setup.sh @ main ad45fa5 -->

**Every task's verify step is the lint guards its own files need plus its own harness**, never
`scripts/run-guard-tests.sh` — that run is the last bundle's FULL SUITE paragraph and
`flow.verify`'s.

---

- [ ] 1. The prompt hook and its harness

**Build:** green
**Files:** `hooks/flow-active-change.py`, `scripts/test-flow-active-change-hook.sh`
**Tests:** `test-flow-active-change-hook.sh` — `Case 1: one /flow mark — the context line names
its change`, `Case 2: two /flow marks for different changes — the last one wins`, `Case 3: only
/flow-plan and /flow-fast marks — no output, exit 0`, `Case 4: a prompt starting with / — no
output, exit 0`, `Case 5: malformed stdin — no output, exit 0`, `Case 6: no transcript_path — the
rollout under FLOW_ZCODE_ROLLOUTS_DIR named by session_id is read`
**Regression:** reverting this commit removes the hook, so no plain prompt ever carries the
change name and the rule in task 3 rests on the session's own memory alone.
**Baseline:** `test-flow-active-change-hook.sh` before=0 after=6
<!-- predicted: grep -c '^# Case [0-9]' scripts/test-flow-active-change-hook.sh after task 1 -->
**Commit:** `feat(hooks): inject the session's active flow change on every plain prompt`

  - [ ] **Step 1: The hook.** `hooks/flow-active-change.py`, Python 3 standard library only, the
    module docstring stating the contract as `hooks/enforce-agent-baseline.py`'s does: reads the
    stdin JSON; returns 0 with no output when the payload is not an object, when `prompt` starts
    with `/`, or on any exception. Transcript: `transcript_path` when present, else
    `<root>/model-io-sess_<session_id>.jsonl` with `<root>` = `FLOW_ZCODE_ROLLOUTS_DIR` or
    `~/.zcode/cli/rollout` and `<session_id>` = the payload's `session_id` or the
    `CLAUDE_SESSION_ID` environment variable. Read the file as bytes decoded with
    `errors="replace"`, take the last match of

  ```python unverified:confirm the pattern matches a real Claude Code transcript line carrying a fix run's flow.document-fix mark — the mark's flag order is the one skills/flow/implement.md section 3 prints
  re.compile(r"flow stage begin -command '/flow' [^;|&\n]*?-session-token mf-[A-Za-z0-9_-]+ ([A-Za-z0-9._-]+)")
  ```

    and print `{"hookSpecificOutput": {"hookEventName": "UserPromptSubmit", "additionalContext":
    LINE}}` via `json.dumps`, `LINE` being `flow: this session's last /flow run marked change
    <name>. A plain message reporting a problem or asking for a change to it is a fix run of
    <name> — **A plain message at IN_PROGRESS** (skills/flow/SKILL.md).` No match, no file: exit
    0, no output.
  - [ ] **Step 2: Harness.** `scripts/test-flow-active-change-hook.sh`, `set -euo pipefail`, a
    `mktemp -d` sandbox removed on exit, one `# Case N: …` comment per case exactly as
    `**Tests:**` names them, a `run_hook <payload-json>` helper that pipes the payload to
    `python3 "$REPO/hooks/flow-active-change.py"` capturing stdout and rc. Fixture transcripts
    are JSONL lines whose `"command"` string carries the mark, written with `printf` so the
    single quotes survive. Case 1 asserts rc 0 and stdout containing `marked change
    fixture-one`; Case 2 writes marks for `fixture-one` then `fixture-two` and asserts
    `fixture-two`; Case 3 writes `-command '/flow-plan'` and `-command '/flow-fast'` marks only and
    asserts empty stdout; Case 4 sends `"prompt": "/flow x"` over Case 1's transcript and asserts
    empty; Case 5 pipes `not json` and asserts rc 0 and empty; Case 6 omits `transcript_path`,
    sets `FLOW_ZCODE_ROLLOUTS_DIR` to the sandbox holding `model-io-sess_abc.jsonl` with a mark,
    sends `"session_id": "abc"` and asserts `marked change`. Print `PASS`/`FAIL` per case and exit
    non-zero on any failure, the shape `scripts/test-check-self-review-report.sh` uses.
  - [ ] **Step 3: Verify.**

  ```bash verified:the harness is this task's own file; check-vocabulary.sh is a .flow/project.md ## lint row and scans scripts/ and hooks/
  scripts/test-flow-active-change-hook.sh
  scripts/check-vocabulary.sh
  ```

- [ ] 2. Registration reporting, README, setup tests and the live check

**Build:** green
**Files:** `setup.sh`, `scripts/test-setup.sh`, `README.md`
**Tests:** `test-setup.sh` — the assertions `the flow-active-change hook is installed`, `an
unregistered flow-active-change hook is reported` and `the zcode flow-active-change hook is
linked`
**Regression:** reverting this commit makes `setup.sh` report a fresh install as fully
registered while the prompt hook is linked and never fires.
**Baseline:** `test-setup.sh` before=146 after=149
<!-- predicted: grep -c '^assert_' scripts/test-setup.sh after task 2 -->
**Commit:** `feat(setup): report and document the flow-active-change hook's registration`
**After:** Task 1

  - [ ] **Step 1: `install_hooks`.** After the existing `enforce-agent-baseline` grep, a second
    `grep -q 'flow-active-change' "$settings"` under the same `require_grep_ok` discipline; return
    early only when both are registered. When the prompt hook is missing, print `⚠ The
    flow-active-change hook is installed but NOT registered, so a plain problem report is never
    named as a fix of the open change. Add this to "hooks" in $settings:` and the snippet

  ```json verified:the shape is the PreToolUse snippet install_hooks already prints, with the event and command swapped
      "UserPromptSubmit": [
        {
          "hooks": [
            { "type": "command", "command": "python3 \"$HOME/.claude/hooks/flow-active-change.py\"" }
          ]
        }
      ]
  ```

  - [ ] **Step 2: `install_hooks_zcode`.** The same second grep against `$config` and the
    ZCode-shaped snippet: a `UserPromptSubmit` entry beside `PreToolUse` under `events`, command
    `python3 "$HOME/.zcode/hooks/flow-active-change.py"`.
  - [ ] **Step 3: README.** In the tree listing, under `hooks/`, add
    `flow-active-change.py ← UserPromptSubmit hook: names the session's last /flow change on
    every plain prompt, so a problem report runs as its fix run`; in the ZCode **Step 2**
    snippet, add the `UserPromptSubmit` event beside `PreToolUse`.
  - [ ] **Step 4: Setup tests.** In `make_fixture_repo`, beside the baseline fixture hook, write
    `$d/hooks/flow-active-change.py` with the same two-line stub. Beside the three existing hook
    assertions add the three `**Tests:**` names: `assert_exists` on
    `$home/.claude/hooks/flow-active-change.py`, `assert_contains "$RUN_LOG"
    'flow-active-change.py'` in the global group, `assert_symlink` on
    `$home/.zcode/hooks/flow-active-change.py` in the ZCode group.
  - [ ] **Step 5: Live check.** Pick a real transcript:
    `grep -l "flow stage begin -command '/flow' " ~/.claude/projects/-Users-tweety53-Projects-agents/*.jsonl | head -1`
    — the grep confirms the file holds a `/flow` mark, so an empty result below is a failure and
    not an unmarked transcript. Pipe `{"prompt":"the button is misaligned","transcript_path":"<that
    file>"}` to `python3 hooks/flow-active-change.py`; success is one JSON line whose
    `additionalContext` names the change that transcript's last mark carried.
  - [ ] **Step 6: Verify.**

  ```bash verified:test-setup.sh is this task's own harness; the three guards are .flow/project.md ## lint rows covering setup.sh's fixture and README.md
  scripts/test-setup.sh
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  ```

- [ ] 3. The router rule

**Build:** green
**Files:** `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md`, `commands/flow.md`, `commands-claude/flow.md`, `scripts/check-contract-budget.sh`
**Tests:** **none** — a skill, a contract and two stubs carry prose; the guards in the verify
step are their checks.
**Regression:** reverting this commit leaves the hook naming a change on every prompt with no
rule that turns the message into a fix run, so the store stays as empty as today.
**Baseline:** `test-setup.sh` before=149 after=149 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 3 -->
**Commit:** `feat(flow): run a plain problem report at IN_PROGRESS as a fix run of the session's change`
**After:** none

  - [ ] **Step 1: Transition table.** In `skills/flow-contracts/pipeline.md`, after the
    `IN_PROGRESS, with an argument` row:
    ``| *(none — a plain message)* | `IN_PROGRESS`, in the session whose last `/flow` run marked the change | fix run; state unchanged — **A plain message at `IN_PROGRESS`** (`skills/flow/SKILL.md`) |``.
    In **Every invocation is re-entrant**, after the "with an argument" bullet: `**At
    IN_PROGRESS, with no invocation at all** — a plain message reporting a problem or asking for
    a change, typed in the session that ran the last /flow <name> — is that same fix run, the
    message its instructions.` The bare-invocation paragraph that starts integrate is untouched.
  - [ ] **Step 2: The subsection.** In `skills/flow/SKILL.md`, under **Reading the state**,
    after the state bullets and before **Check guard presence**, add `### A plain message at
    IN_PROGRESS` carrying, in this order: (a) the trigger — a message with no `/flow` invocation
    that reports a problem or asks for a change to the change's code or artifacts, in a session
    whose most recent `/flow` run marked `<name>`; `<name>` comes from the `flow: this session's
    last /flow run marked change <name>` context line `hooks/flow-active-change.py` injects, or
    from this session's own context when no hook is registered; (b) the gate — `flow state get
    <name>` must answer `IN_PROGRESS`; any other answer, or an unreachable store, makes the
    message an ordinary turn, never a wrong-state handoff, since nothing was invoked; (c) the
    ambiguity prompt — a message that reads as either a question or a change request asks once,
    shape per **The shape** (`skills/flow-contracts/operator-prompts.md`): `**Run this as a fix
    of <name>?**` — **Yes** *(recommended)* / **No — ordinary turn**; silence takes Yes and the
    handoff carries the ⚠ line; (d) the run — announce `Using flow for change <name> — fix run
    from a plain message`, generate a fresh `mf-` token exactly as the paragraph below generates
    one, and continue at **3. Documenting a fix, before implementing it**
    (`skills/flow/implement.md`) with the verbatim message as the fix instructions — nothing
    else about the fix run changes; (e) what is not a trigger — a question, a discussion, an
    unrelated task, a session that never ran `/flow`, and a Jira key named in prose without the
    slash. Add `— or a plain message, per **A plain message at IN_PROGRESS** below` to the
    `IN_PROGRESS, an argument present` bullet.
  - [ ] **Step 3: Stubs.** In both `commands/flow.md` and `commands-claude/flow.md`, after the
    sentence ending "the argument is fix instructions.": `A plain problem report typed with no
    /flow at all, in the session that ran the last /flow <name>, is the same fix run (**A plain
    message at IN_PROGRESS**, skills/flow/SKILL.md).`
  - [ ] **Step 4: Budget.** `skills/flow/SKILL.md` sits at its `budgets()` row exactly (17812);
    set the row to the file's new size plus 25%. Check `commands/flow.md`,
    `commands-claude/flow.md` and `skills/flow-contracts/pipeline.md` against their rows after the
    edits and raise any the growth crosses.
  - [ ] **Step 5: Verify.**

  ```bash verified:each guard is a .flow/project.md ## lint row and runs without arguments over skills/, commands/ and commands-claude/
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  scripts/check-stage-mark-calls.sh
  scripts/check-installed-citations.sh
  ```
