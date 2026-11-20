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
> **A capability spec under `<project>/spectre/specs/` is your work, not theirs**: when this task's
> `**Files:**` names one, edit it and commit it here, in this task's own commit. **A withdrawn
> task is backed out, not erased**: when work an earlier commit landed is withdrawn, the
> backing-out commit's body carries the rejection reason — what was wrong, and what would have to
> change for it to come back (**Revert and back-out commits**,
> `skills/flow-contracts/git-boundaries.md`). **Your commit's message describes the tree it lands**: an attempt you
> wrote and undid before committing is never what its subject or body describes (**Revert and
> back-out commits**, `skills/flow-contracts/git-boundaries.md`).

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
> against them.

> **PLAN AMENDMENT:** When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-amendment.md`).

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop. A `flow-guard` run counts: write your report file last, after
> every guard has exited, and hand back only then.

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

> **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For
> every behaviour, name, path, flag or exit code the diff changes, grep for the comments, file
> headers, usage lines, docs and cross-file citations that describe it, and bring each one in
> line in this same commit. Every refusal or exit path the diff adds gets a test case of its
> own.

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

> **AFTER A COMPACTION:** Your context is compacted near 150,000 tokens: a summary takes the
> place of your earlier turns. When your context opens with such a summary, re-read this brief
> before your next action — the summary does not carry it.

> **SCRUB A REPRODUCED ARTIFACT:** A real artifact reproduced into a committed fixture — a transcript, a log, a database dump — is reduced to the lines the test needs and scrubbed of every identifier, path, prompt text and personal datum before it is staged; the raw copy stays under `<abs-worktree>/.superpowers/sdd/`, never in the tree (kan-1132 self-review, F6).

> **SCRATCH HARNESS:** A throwaway harness — a measuring spec, a repro script, a probe — lives in
> your session scratchpad or in this change's state dir (the path `flow state dir` prints), never
> in a new worktree or branch of any repository. Cleanup removes only the change's own worktrees
> and branches, so one you create outlives the run.

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
> commit and your final test run — carrying each commit's sha and the failing RED output you saw. The report opens with this
> summary head, and your final message is the head alone — the dispatcher reads it from your
> completion notification and never opens the report body:
>
> ```text
> status: DONE | BLOCKED
> commit: <task-id> <full sha>
> fixed: F<n> <full sha>
> not-fixed: F<n> — <reason>
> question: <the question, verbatim>
> report: <this report file's absolute path>
> ```
>
> `status:` and `report:` appear once; every other line once per item — a `commit:` per
> task, its sha a `<worktree>=<sha>[,<worktree>=<sha>…]` map on a change spanning repositories, a
> `fixed:` or `not-fixed:` per finding a FIX INPUT names — and not at all when there is none.

> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded
> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not
> name — is reported, never decided in code. Where the plan can proceed, implement it as written
> and carry the question in your REPORT FILE; where it cannot proceed without the answer, report
> BLOCKED. A decision you make silently is a decision nobody recorded.

> **A FLOW TOOLING DEFECT IS REPORTED, NEVER FIXED HERE:** a guard, hook, script, skill or contract
> of the `agents` repository that wrongly blocks your task is named in your REPORT FILE — the
> command, its output and the defect — for the parent's in-run pipeline-fix loop (**Pipeline
> defects found mid-run**, `skills/flow-contracts/pipeline.md`), and never edited in this worktree
> unless your task's own `**Files:**` names that file.
