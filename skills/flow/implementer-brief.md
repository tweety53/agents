# Implementer brief

Every implementer and fixer dispatch of **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) reads
this file in full first: its prompt is a stub naming this file, never these paragraphs. On an
`inline` run the parent loads it and is bound by it in the same words (**Inline — the parent
implements**, `skills/flow/implement.md`).

> **FLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the guard runs on it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer. The trailer identifies the task;
> the subject is this task's declared `**Commit:**` field, reproduced exactly. **The commit is
> pathspec-scoped** — `-- <this task's files>` — so it carries only the paths this task names,
> whatever else the index holds; a plain commit sweeps a pre-staged foreign tree in with the
> task's work. **Never weaken or
> bypass a project's commit validation to fit** — no `--no-verify`. Stage for that commit only
> through the guarded sequence below, in this order — the clearing pass runs first, before any
> `git add`, because a `:(exclude)` governs what an `add` adds and cannot retract what an earlier
> step already staged:
>
> ```bash
> stage-task-files.sh <abs-worktree> <this task's files> \
>   && git -C <abs-worktree> commit -m "<the task's declared subject>" -m "Task-Id: <n>" -- <this task's files>
> ```
>
> `stage-task-files.sh <abs-worktree> <file>…` runs that clearing reset, then the add with the excludes
> ahead of the paths, then exits 1 when git left any of this task's paths unstaged: report the lines it printed and
> commit nothing (kan-1012: `git add` exited 0 with 15 new paths untracked, and the commit failed).
> Its header (`<agents repo>/scripts/stage-task-files.sh`) is canonical for the sequence.
> Paths are relative to the worktree root, which `-C <abs-worktree>` anchors whatever directory
> the shell is in — from a subdirectory of it, an unanchored `git add` reads them
> against that directory and fails with a pathspec error. Both spec-tree leaf spellings are named because the
> leaf is the project's — `spectre` or `openspec` — and a path the project does not use is a
> reset that clears nothing and an exclude that matches nothing. **Never a bare
> `git add -A`, `git add .`, or `git commit -a`**:
> any of them sweeps already-staged planning artifacts into the task commit, which
> `check-task-commit-planning-paths.sh` fails at the boundary. Never
> `<project>/spectre/changes/` either — the excludes do not licence a second, deliberate add of it.
> **A capability spec under `<project>/spectre/specs/` is your work, not theirs**: when this task's
> `**Files:**` names one, edit it and commit it here, in this task's own commit. **A withdrawn
> task is backed out, not erased**: when work an earlier commit landed is withdrawn, the
> backing-out commit's body carries the rejection reason — what was wrong, and what would have to
> change for it to come back (**Revert and back-out commits**,
> `skills/flow-contracts/git-boundaries.md`). **Your commit's message describes the tree it lands**: an attempt you
> wrote and undid before committing is never what its subject or body describes (**Revert and
> back-out commits**, `skills/flow-contracts/git-boundaries.md`).

**A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
implementer runs the red task first — writes its tests, runs them, and reports the failing
output — then the green partner, and makes **one** commit for the pair: the partner's declared
`**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.

> **REQUIRED SUB-SKILL:** Read and follow superpowers:test-driven-development from its file — on
> Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/test-driven-development/SKILL.md`
> — RED-GREEN-REFACTOR for this task. Delete any code written before its test.

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

> **REQUIRED READING:** the engineering principles section of the context bundle below — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them.

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context-group-<g>.md` carries this
> change's proposal, design, engineering principles, and — under `## tasks.md` — the plan header
> plus your group's own task(s) only, gathered for you. You **must** still read the
> actual diff and the actual code — the bundle is shared *input*, never a substitute for the source.
> It also carries this project's `## lint`/`## test`/`## run` commands, already resolved — you do
> not need to open `<project>/.flow/project.md` yourself for them.

> **PROJECT HAZARDS:** the bundle's `## hazards` section carries this project's recorded
> warnings, filtered to this change's shape — each one names a way this project specifically
> bites, recorded after it cost time. They are binding: read them before your first edit and
> never argue one away without measuring.

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. An assumption tagged
> `unverified:` in your task is a guess to confirm before you build on it. When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-amendment.md`). Report a correction your
> measurement produced — a tag to retag, a number to replace, a `**Files:**` entry to add — as the
> exact field and its corrected value, so it can be transcribed into the record in place
> (The record carries its own corrections, `skills/flow/implement.md`).

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

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
> run the module or repository suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. A plan step that asks for the module or repository
> suite is run as this task's targeted tests instead. Pipe a test run's output through `tail` so a
> green run costs lines of context, not a build log.

> **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For
> every behaviour, name, path, flag or exit code the diff changes, grep for the comments, file
> headers, usage lines, docs and cross-file citations that describe it, and bring each one in
> line in this same commit. Every refusal or exit path the diff adds gets a test case of its
> own.

> **PROOF RUNS:** Every build, test suite or probe you start runs under
> `<agents repo>/scripts/run-capped.sh --`, which kills its whole process tree past the project's
> `## resource cap` (default 4096 MB). The cap is per command; a run-capped.sh started under another
> joins the outer cap. Gradle runs with `--no-daemon` (run-capped.sh sets it), since a daemon
> reparents out of the tree, and with bounded `--max-workers` and `-Xmx`. A dev stack detaches, so
> it is bounded by its own config — worker counts and heap — not by run-capped.sh. A flake fix
> carries one of two proofs. Either its cause is reproduced — preferably in one isolated run, a delay
> or fault injected at the seam, the test alone or with the one test whose leftover state it reads;
> a load reproduction (backend or frontend) is allowed only
> under run-capped.sh with an explicit bounded worker count sized under the cap: Gradle `--max-workers`/`maxParallelForks` and `-Xmx`,
> Playwright `--workers`, vitest `--maxWorkers`, at most 5 capped runs — and the fix then shows the
> check failing with the mechanism removed (`~/.claude/rules/fix-determinism-at-the-source.md`). Or
> the flake does not reproduce within the cap and goes straight to code: no further reproduction
> loops; fix it on deeper analysis of the code and the failing run's saved artifacts, and the commit
> body says so — that alone waives the mechanism-removed run. Your report states which, with the
> forced condition, the cap, the worker count and the run count.

> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your
> whole context — a large output is paid for again on every turn after it. Read a file over 200
> lines by line range — `grep -n` for the symbol, then `sed -n '<a>,<b>p'` or Read with
> `offset`/`limit` — and never re-read a file already in your context unless you have edited it
> since. Cap every search (`| head -40`) and every build, lint or install run (`| tail -30`), and
> reproduce a failing block from its log rather than printing the whole log. Never print a
> generated file — a lockfile, a snapshot, a bundle, a build artifact.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.

> **SCRATCH HARNESS:** A throwaway harness — a measuring spec, a repro script, a probe — lives in
> your session scratchpad or in this change's state dir (the path `flow state dir` prints), never
> in a new worktree or branch of any repository. Cleanup removes only the change's own worktrees
> and branches, so one you create outlives the run (kan-960 left a `kan-960-measure` worktree and
> branch in `gymie-playwright`).

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
> commit and your final test run — carrying each commit's sha, the failing RED output you saw,
> and anything the plan's `unverified:` tags asked you to establish. The report opens with this
> summary head, and your final message is the head alone — the dispatcher reads it from your
> completion notification and opens the report body only where a judgement needs it:
>
> ```text
> Model: <the model named in your own system prompt>
> status: DONE | BLOCKED
> commit: <task-id> <full sha>
> correction: <task-id> <field> — <corrected value>
> undeclared: <task-id> <path>[,<path>…] — <one-line reason>
> fixed: F<n> <full sha>
> not-fixed: F<n> — <reason>
> fix-mutation: <path> — <what was mutated> — <the test that failed>
> question: <the question, verbatim>
> report: <this report file's absolute path>
> ```
>
> `Model:`, `status:` and `report:` appear once; every other line once per item — a `commit:` per
> task, its sha a `<worktree>=<sha>[,<worktree>=<sha>…]` map on a change spanning repositories, a
> `fixed:` or `not-fixed:` per finding a FIX INPUT names — and not at all when there is none.

> **FIX DISPATCH:** a prompt carrying a `FIX INPUT:` line makes you a fixer, not a task's
> implementer: fix exactly what that line names — reviewer reports and finding refs, a
> `close-task.sh` refusal's lines, a wave member's commits to re-apply, or a failing suite block —
> at its root cause, under every other paragraph of this brief save the commit subject and trailer
> FLOW — COMMIT-PER-TASK names, which this paragraph's take the place of. A finding's fix is one
> commit per task at the branch tip, carrying that task's `Task-Id:` trailer,
> `-m "Flow-Key: <your key>"` and a `fix(<module>): …` subject; a re-commit rewrites only the
> refused, unpushed task commit, without the paths the refusal names, keeping its declared
> `**Commit:**` subject and `Task-Id:` trailer; a pick re-applies the member's commits onto the
> branch tip in order, each keeping its subject and trailer; a full-suite fix commits with
> `-m "Flow-Key: full-suite-fix-<n>"` and no `Task-Id:`. When the line names MUTATION PROOF, follow
> that paragraph (`skills/flow/review-panel.md`) in its exact words before you end your turn.
> Whether a finding stands is the dispatcher's call, never yours: one you cannot fix is a
> `not-fixed:` line, never a silent skip.

> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded
> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not
> name — is reported, never decided in code. Where the plan can proceed, implement it as written
> and carry the question in your REPORT FILE; where it cannot proceed without the answer, report
> BLOCKED. A decision you make silently is a decision nobody recorded.
