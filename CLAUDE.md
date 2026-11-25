# Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.claude/CLAUDE.md` from `<agents repo>/rules/lint-fix-priority.mdc`.

What is project-specific is which commands it means. Full list in `<project>/.flow/project.md`'s `## lint`
section — named here so this file states them rather than leaving a placeholder:

```bash
cd stats && gofmt -w .                                  # auto-fix, Go source only — nothing else in
                                                          # this repository has an auto-fix command
scripts/flow-guard check-guard-calls                              # plus every other scripts/flow-guard check-* guard
                                                          # named in <project>/.flow/project.md's `## lint`
cd stats && make -s web-dist && go vet ./... && gofmt -l .  # must exit clean before claiming Go work done
cd stats && make -s web-dist && cd web && npx tsc -b      # must exit clean before claiming SPA work done
```

---

### No incident history in instruction files

Rules, skills, contracts, templates and rationale appendices state the rule and its reason, never the incident behind it; incidents are recorded with `flow record incident`.

---

### No scripts in bash

Every new script, and any new logic added to an existing one, is Go in `flow-guard`:
`stats/internal/guard/<name>.go`, registered in `guard.Registry`, tested beside it, called as
`flow-guard <name>`. Every non-zero exit of a guard first records `guard.Reason(env, <id>)`,
`<id>` a kebab-case name of the verdict, never of its subject (`untasked`, not `fresh`) —
`TestEveryExitSiteRecordsAReason` and `TestReasonIDsNameTheReason` (`stats/internal/guard/`) fail
on a missing or subject-named id. `setup.sh` and `scripts/flow-guard` are the only bash files in
this repository and stay that way; no guard checks it. The full rule is `rules/no-bash-scripts.mdc`.

---

### Never stop the dev workspace's stats service or its storage

`flowd` on `127.0.0.1:4173`, the `flow-postgres` container on host port 5433, and the default
`flow` database inside it are the **dev workspace's** service and storage — the store every
`flow` call in every project writes state, stage marks and records through. No agent action stops
or drops them: not `docker compose down`, not `launchctl unload`, not a `kill` on the daemon's pid,
and not to make a later step succeed. Bringing them back up does not repair a run that already fell
through to the on-disk journal.

**Only those.** A worktree's own derived `flow_<id>` database and bucket are per-change artifacts,
and `<project>/scripts/flow-guard workspace remove <id>` drops them during the cleanup run as it should.

The reasoning and that boundary are the `## stop` section of `<project>/.flow/project.md` — canonical for it, and not
restated here. Stopping the dev stack is an operator action; the commands live in that file's
`## run` section and in `<project>/stats/README.md`.

### Reinstall after a change to this repository is fully done

Once a change to this repository has fully landed — on `/flow`, after run 2's cleanup; on
`/flow-fast`, after its cleanup; on `/flow-interactive`, after each request lands; on a
pipeline-fix or self-review loop, after `flow-guard pipeline-fix land` — run, from the main checkout:

```bash
./setup.sh global
make -C stats restart
```

`setup.sh global` reinstalls the rules, skills, hooks and guard symlinks every session reads, and
`make restart` rebuilds the `flow` CLI and restarts `flowd`, so the next run acts on what just
landed instead of the copies installed before it. This restart is the one stop of `flowd` the
rule above allows: it brings the service straight back on the new build.

---

## Project Skills (spectre / /flow workflow)

These skills live in `skills/` next to this file (or in `<project>/.claude/skills/` if installed there).
To invoke a skill, type its slash command (`/flow`, `/flow-fast`, …); the harness loads its
`SKILL.md` through the Skill tool.

Every skill below but `flow-plan` and `flow-contracts` requires the `spectre` CLI to be
installed. Those two need none — reading a spectre tree, or a contract file, is reading markdown.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/flow/` | `/flow` | Single-command pipeline: brainstorming behind a design gate, implementation under SDD + TDD behind the decided review panel, and integrate/cleanup across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/flow-fast/` | `/flow-fast` | Minimal-ceremony `/flow` variant: one invocation from Jira key to landed change. A git worktree for isolation only, implementation and review panel as `flow-guard plan-check … -decide-only` decides, project lint plus targeted tests, the project's default landing route, cleanup. Marks every `flow.*` stage `/flow` marks and keeps the Jira transitions; no spectre artifacts or state file |
| `skills/flow-interactive/` | `/flow-interactive` | Conversational session: one invocation is one change, one Jira issue and one worktree, and a stream of requests the main session coordinates — it looks at each live, records, dispatches and reports, and one pair-judged `operator-asked` subagent per code request implements it test-first, verifies it with the project's lint and targeted tests, lands it by merge and push, reinstalls and, when it touches `ui paths`, checks the live page — until an explicit stop runs one self-review pass, cleanup and `FINISHED`. Every request is recorded in flowd; no spectre artifacts or state file |
| `skills/flow-plan/` | `/flow-plan` | Thinking-partner mode — explore ideas, investigate, no implementation; a captured session creates the change at `STARTED` for `/flow` to resume |
| `skills/flow-learn-quiz/` | `/flow-learn-quiz` | Teacher mode, only when asked — restates, checklists and quizzes you (AskUserQuestion) on a change until you show you understand it. Standalone, not a pipeline stage |
| `skills/flow-self-review/` | *(run by the pipeline)* | Runs a change's self-review pass, a one-shot dispatch — inside `/flow`'s integrate, `/flow-fast`'s verify and `/flow-interactive`'s stop, or standalone from the context bundle flowd saved at cleanup; fixes and lands every `flow-fix` finding that is not big, suggests the impactful big ones and other angles' findings (no Jira issue), records each in the flow store. Standalone, not a pipeline stage |
| `skills/flow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for `/flow`) plus the state file, project configuration and Jira contracts, `jira-followups.md` when `/flow`'s integrate run 1 files or joins a follow-up, `finish-contract-run1.md`/`finish-contract-run2.md` for `/flow`'s two-run integrate/cleanup procedure, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need — and never a `-rationale.md` appendix, which carries a contract's or a skill's reasoning for whoever edits it and is not loaded by a run |

### Superpowers general skills

The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD,
subagent-driven-development, etc.). These are referenced by the `/flow` skill above.

Install it per `<agents repo>/README.md`'s Claude Code section.
