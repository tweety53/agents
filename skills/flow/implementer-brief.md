# Implementer brief

Every implementer and fixer dispatch of **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) reads
this file in full first: its prompt is a stub naming this file, never these paragraphs.

> **FLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the guard runs on it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer. The trailer identifies the task;
> the subject is this task's declared `**Commit:**` field, reproduced exactly. **The commit is
> pathspec-scoped** — `-- <this task's files>` — so it carries only the paths this task names,
> whatever else the index holds; a plain commit sweeps a pre-staged foreign tree in with the
> task's work. **Never weaken or
> bypass a project's commit validation to fit** — no `--no-verify`:
>
> ```bash
> git -C <abs-worktree> add -- <this task's files> \
>   && git -C <abs-worktree> commit -m "<the task's declared subject>" -m "Task-Id: <n>" -- <this task's files>
> ```
>
> Paths are relative to the worktree root, which `-C <abs-worktree>` anchors whatever directory
> the shell is in — from a subdirectory of it, an unanchored `git add` reads them
> against that directory and fails with a pathspec error. **Never a bare
> `git add -A`, `git add .`, or `git commit -a`**:
> any of them sweeps already-staged planning artifacts into the task commit. Never
> `<project>/spectre/changes/` either.
> **A withdrawn
> task is backed out, not erased**: when work an earlier commit landed is withdrawn, the
> backing-out commit's body carries the rejection reason — what was wrong, and what would have to
> change for it to come back (**Revert and back-out commits**,
> `skills/flow-contracts/git-boundaries.md`). **Your commit's message describes the tree it lands**: an attempt you
> wrote and undid before committing is never what its subject or body describes (**Revert and
> back-out commits**, `skills/flow-contracts/git-boundaries.md`).

> **CONFIDENCE ONCE GREEN:** Only a finished group is estimated: a `BLOCKED` or `PARTIAL` report files
> nothing. A group is not finished while a guard it runs exits non-zero or its measure carries a
> flagged row (**MEASURE A UI TASK BEFORE YOU REPORT**): fix it, or report `BLOCKED` naming the
> guard — never `DONE` carrying it as a gap, whatever the number. Once its tests are green, before your commit, whoever turns the group green — its
> implementer, or a fixer closing what its implementer's report left open — files one estimate, never
> acted on. Answer for the tasks of this dispatch: how confident are you,
> 1–100%, that you executed them as the plan writes them, and what would move you to 100% — one gap
> per thing still open, each naming what would close it and ending with the command that would
> verify or close it (`flow-guard check-frame-constants …`, `flow-guard visual-verify-pass measure …`);
> a gap without its command is incomplete. Then run `flow confidence tags`, reuse a
> printed tag only when it names the same kind of gap — a near fit hides two unrelated gaps under one
> name — and write a new lowercase kebab-case tag only when none does (a tag names the kind of gap,
> never the change), and file it:
>
> ```bash
> flow confidence record -C <abs-worktree> -change <name> -stage implement -task '<the ids of the group, e.g. 2+3>' \
>   -score <n> -gap '<one gap, in full>' [-gap …] -tag <tag> [-tag …]
> ```
>
> `<name>` is the last path element of your `Plan:` line's change directory. At 100% pass `-score
> 100` with the first three flags and no `-gap` or `-tag`: it is kept as a plain 100%. An estimate
> below 90% reaches the operator's review list; one of 90% or more is stored and shown beside your
> group only. **Whatever the number, continue** — a low estimate is never a question, never a `BLOCKED`
> and never a reason to change the plan; a gap that is a question only the operator can settle still
> goes through **REPORT, DON'T DECIDE**, as it would unestimated. A question your report carries names
> only a choice **What still stops** (`skills/flow-contracts/operator-prompts-auto-resolution.md`)
> lets a run stop for: a red guard, an open flag or a failed check is your next step, never a
> question. A warning or a non-zero exit of
> either command is noted in your REPORT FILE and does not stop you. **A gap that reading code settles
> is settled by reading it — "unread" is never a gap.** **A gap a command inside your tasks can settle is settled by running it** — a test, a
> query against real data, a build — and is never left to stand as a gap; name that file or command in your REPORT FILE. A
> scenario no test covers is closed by writing the test inside your task. **A gap is never one of
> these, whatever the number:** existing code that contradicts a recorded decision or the plan (the
> plan's decision is applied, or the task reports BLOCKED — **REPORT, DON'T DECIDE**; "kept existing"
> is a silent decision), a test you wrote and have not run (run it; if it cannot run here the
> task is `BLOCKED` naming the command and why, never `DONE`), and a seam you left inert — a default
> no-op, a stub or a placeholder that nothing in your tasks replaces (wire the real one inside your
> task, or carry it as a question in your REPORT FILE for the plan), a guard, gate, branch or
> condition you added that no test fails without (write the test, delete the guard, watch it fail,
> restore it), and work a later task of the plan
> names — a capture run, a wiring step — which your REPORT FILE carries as `owned by Task <n>` and your
> estimate leaves out. Below 100, a gap there is one no command or file read in your tasks could
> settle, the REPORT FILE saying why. Any other fixer skips this paragraph.

> **REQUIRED SUB-SKILL:** Read and follow superpowers:test-driven-development from its file — on
> Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/test-driven-development/SKILL.md`
> — RED-GREEN-REFACTOR for this task. Delete any code written before its test.

> **GO CHANGES:** When this task's files include a `*.go` file, read and follow the `go-coding-standards`
> and `go-error-handling` skills first — a `*_test.go` file adds `go-test-quality`. A task with no `.go`
> file skips this.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, read and
> follow superpowers:systematic-debugging from its file — on Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/systematic-debugging/SKILL.md`
> — before writing a fix. An expected RED step needs no such read.

> **PROVE THE GUARD BITES:** When this task's tests assert on configuration or file content — a
> guard script, an embedded config, a fixture file — a passing run alone is not evidence: break
> the property the assertion protects, run the guard or the test against the broken state and
> capture its failure, then restore and capture the pass. Report both runs. "The pattern is now
> stricter" is intent, not evidence — the failing run against the broken state is the evidence.

> **PIN BEFORE REFACTOR:** A task that refactors code whose observable behaviour includes
> produced output — a guard's warnings, a CLI's printed lines, an emitted file — pins that
> output before any code moves: its RED step first writes harness cases asserting the current
> outputs exactly as produced today, including the ones nothing asserted yet, and only then
> refactors. The refactor lands with two proofs, both recorded before the change closes: the
> pinned suite passing, and one end-to-end old-vs-new comparison that runs the whole producer
> against the same fixture on both sides of the refactor and diffs the outputs byte-for-byte.
> The suite proves only what it asserts; the old-vs-new diff is the proof for every output it
> never named.

> **REQUIRED READING:** `engineering-principles.md` beside this file — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them. **Scope lifetime** is the one that bites async work a control starts.

> **PLAN AMENDMENT:** When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-amendment.md`).

> **MEASURE A UI TASK BEFORE YOU REPORT:** when a file under your task's `**Files:**` matches the
> project's `ui paths` (`## visual verification`, `.flow/project.md`), commit, then run
> `flow-guard visual-verify-pass measure <canonical-worktree> <name> task-<task-id> <the task's capture spec>`
> and fix every row it flags at its source — commit and re-run until none is left. It writes
> `<changeRoot>/visual-verification/task-<task-id>-measure.json`; `flow-guard close-task` refuses
> the task (`flow-guard check-ui-task-measured <canonical-worktree> <task-id>`) while that record is
> missing, older than your last commit to the task's files, or carries a flagged row, and prints
> this command. A length delta within the project's `tolerance` is not flagged. A task whose plan
> adds the capture spec in a later task is not measured per task; **the task that adds that spec
> measures every earlier UI task's frame** with it, on the same command, and fixes each flagged row at
> the earlier task's source before it reports.

> **NO FRAME VALUE WITHOUT A MODEL:** a mockup frame's value — a setting's current value, a count,
> a name, a date — is drawn only where a model, store or API field holds it. A frame element whose
> value no model holds is never drawn with the frame's literal value: a control behind it then
> shows made-up data that nothing changes. Report it as a departure, or as a
> question under **REPORT, DON'T DECIDE**, and leave the element out of the code until the
> operator settles it.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop. A `flow-guard` run counts: write your report file last, after
> every guard has exited, and hand back only then.

> **STASH:** Never use git stash pop/apply/drop; if a worktree looks wrong, stop and report BLOCKED with the exact `git status` output.

> **FILE EDITS:** change a file with the Edit or Write tool, never with an inline script — `python3 -c`/heredoc, `sed -i`, `perl -i`, a `cat >` redirect. A file-path allow rule covers the Edit tool; a shell edit falls to the permission classifier, which may deny it and stop the run.

> **GIT CALLS:** run `git commit` and `git push` each as its own Bash call — never chained with `&&`, `;` or a pipe to another command. Allow rules match a single command; a chained one falls to the permission classifier, which may deny it and stop the run.

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build
> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t
> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. A plan step that asks for the module or repository
> suite is run as this task's targeted tests instead. Pipe a test run's output through `tail` so a
> green run costs lines of context, not a build log.

> **A PASS CONDITION IS MET BEFORE YOU REPORT:** a plan step whose command must be green, or each
> failure dispositioned — a regression or capture sweep included — is done only when the command is
> green or every failure is fixed or dispositioned as the step says (re-baselined with the departure
> recorded, restored). TARGETED TESTS above decides which command runs; this decides when its step
> is done. Anything less is `status: BLOCKED`, never `DONE`, with a `question:` line naming the
> command and the count of failures left.

> **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For
> every behaviour, name, path, flag or exit code the diff changes, grep for the comments, file
> headers, usage lines, docs and cross-file citations that describe it, and bring each one in
> line in this same commit. Every refusal or exit path the diff adds gets a test case of its
> own. A diff that removes a UI control, its tag or its accessibility id also greps every test
> and spec for that tag — tests that tap, find or assert on it — and updates them in this same
> task, never leaving them to fail in the suite.

> **PROOF RUNS:** Keep runs bounded — never start an unbounded load, a leaking loop or a suite
> that can exhaust the machine's memory; stop a run that does. A flake fix
> carries one of two proofs. Either its cause is reproduced — preferably in one isolated run, a delay
> or fault injected at the seam, the test alone or with the one test whose leftover state it reads;
> a load reproduction (backend or frontend) is allowed only
> with an explicit bounded worker count: Gradle `--max-workers`/`maxParallelForks` and `-Xmx`,
> Playwright `--workers`, vitest `--maxWorkers`, at most 5 runs — and the fix then shows the
> check failing with the mechanism removed (`~/.claude/rules/fix-determinism-at-the-source.md`). Or
> the flake does not reproduce within those runs and goes straight to code: no further reproduction
> loops; fix it on deeper analysis of the code and the failing run's saved artifacts, and the commit
> body says so — that alone waives the mechanism-removed run. Your report states which, with the
> forced condition, the worker count and the run count.

> **REWRITTEN BASELINES:** a capture that rewrites a baseline PNG in this task follows **No rewritten baseline is staged unread** (**The sweep**, `skills/flow-contracts/known-bugs.md`) before you stage it.

> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your
> whole context — a large output is paid for again on every turn after it. Read a file over 200
> lines by line range — Read with `offset`/`limit`, or `sed -n '<a>,<b>p'`, over the range a
> `grep -n` found — and never re-read a file already in your context unless you have edited it
> since. **A search whose hit needs its surrounding lines is one call** — `grep -n -C<k>`, or the
> Grep tool with context — never a `grep` followed by a `sed -n` window on the same hit. Cap every
> search (`| head -40`) and every build, lint or install run (`| tail -30`), and
> reproduce a failing block from its log rather than printing the whole log. Never print a
> generated file — a lockfile, a snapshot, a bundle, a build artifact.

> **AFTER A COMPACTION:** Your context is compacted near its window's limit: a summary takes the
> place of your earlier turns. When your context opens with such a summary, re-read this brief
> before your next action — the summary does not carry it.

> **SCRUB A REPRODUCED ARTIFACT:** A real artifact reproduced into a committed fixture — a transcript, a log, a database dump — is reduced to the lines the test needs and scrubbed of every identifier, path, prompt text and personal datum before it is staged; the raw copy stays under `<abs-worktree>/.superpowers/sdd/`, never in the tree.

> **SCRATCH HARNESS:** A throwaway harness — a measuring spec, a repro script, a probe — lives in
> your session scratchpad or in this change's state dir (the path `flow state dir` prints), never
> in a new worktree or branch of any repository. Cleanup removes only the change's own worktrees
> and branches, so one you create outlives the run.

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
> commit and your final test run — carrying each commit's sha, the failing RED output you saw and, under
> `## Open gaps`, each `-gap` of your estimate verbatim. The report opens with this
> summary head, and your final message is the head alone — the dispatcher reads the head alone
> and never opens the report body:
>
> ```text
> status: DONE | BLOCKED
> confidence: <n>% — filed
> commit: <task-id> <full sha>
> fixed: F<n> <full sha>
> not-fixed: F<n> — <reason>
> question: <the question, verbatim>
> report: <this report file's absolute path>
> ```
>
> `status:` and `report:` appear once, and so does `confidence:` — the CONFIDENCE ONCE GREEN estimate you made, `filed` after your `flow confidence record` call, at 100% too (omitted by a `BLOCKED` or `PARTIAL` head and by a fixer that turned no group green); a `DONE` implementer head without it says the estimate was skipped; every other line once per item — a `commit:` per
> task, its sha a `<worktree>=<sha>[,<worktree>=<sha>…]` map on a change spanning repositories, a
> `fixed:` or `not-fixed:` per finding a FIX INPUT names — and not at all when there is none.

> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded
> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not
> name — is reported, never decided in code. Where the plan can proceed, implement it as written
> and carry the question in your REPORT FILE; where it cannot proceed without the answer, report
> BLOCKED. A decision you make silently is a decision nobody recorded. **The mechanism the design
> states is the one you build.** One you judge better, or one that suits a single code path, is not
> yours to substitute: build the design's, or report BLOCKED naming why it cannot work, never both
> change it and carry on.

> **AN EDIT DENIED IS REPORTED, NEVER ROUTED AROUND:** when a permission classifier denies an edit,
> report `status: BLOCKED` with the `question:` line `edit denied — <file> — <the exact edit: the old
> and new text, or the line to add>`, and never reach the file another way. The parent makes the
> edit itself where **What still stops**' **Out of the run's reach** bullet
> (`skills/flow-contracts/operator-prompts-auto-resolution.md`) lets it.

> **A FLOW TOOLING DEFECT IS REPORTED, NEVER FIXED HERE:** a guard, hook, script, skill or contract
> of the `agents` repository that wrongly blocks your task is named in your REPORT FILE — the
> command, its output and the defect — for the parent's in-run pipeline-fix loop (**Pipeline
> defects found mid-run**, `skills/flow-contracts/pipeline.md`), and never edited in this worktree
> unless your task's own `**Files:**` names that file.
