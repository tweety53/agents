# Manual test guide — kan-19-finish-safety-records-and-effort

**Worktree (every command below runs here):**
`/Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort`

This repository has **no runnable application** — it is the source of the myflow skills, commands
and rules, installed elsewhere by `setup.sh`. "Running the apps" here means running the guard
scripts, the assertion harnesses, and a sandboxed installer pass. Everything else in this change is
contract and skill text, checked by reading it.

## How to run everything in scope

Open the worktree:

```bash
open -na "IntelliJ IDEA" --args "/Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort"
```

The declared lint commands (`.myflow/project.md` `## lint`) — all three must exit 0:

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

The declared test commands (`.myflow/project.md` `## test`) — all five must exit 0:

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
./scripts/test-check-finish-preflight.sh
./scripts/test-preserve-session-records.sh
```

The installer, against a sandboxed home (`.myflow/project.md` `## run`):

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

The two new scripts are repository-local and are **not** expected to appear in the sandbox — only
skills, commands and rules are installed. The contract text that points at them is.

## Checklist

### The finish preflight — `scripts/check-finish-preflight.sh`

Run it by hand against this very worktree, which is the ordinary `IN_PROGRESS` shape (work present,
nothing committed). It must say `RUN1`, never `RUN2`:

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort
./scripts/check-finish-preflight.sh "$PWD" main a85a729ff120ba820688178e64ea3ac5eb58e607; echo "exit=$?"
```

- [ ] The line above prints `RUN1: …` and exits 0. **This is the whole point of the change** — the
      old ancestor-only test answered "merged" here, and the archive run then `--force`-removed the
      worktree holding all the work.
- [ ] `./scripts/check-finish-preflight.sh "$PWD" main -` prints `REFUSE: …` — no recorded merge
      base is refused, never guessed.
- [ ] `./scripts/check-finish-preflight.sh "$PWD" no-such-base a85a729` still prints `RUN1` — an
      unresolvable base branch cannot mask a branch that has no commits of its own.
- [ ] `./scripts/check-finish-preflight.sh /tmp main -` exits 2 (not a git worktree) — "cannot
      determine" is distinguishable from a verdict.
- [ ] `./scripts/test-check-finish-preflight.sh` prints `all cases pass`. Its cases cover a merged
      branch with a clean tree (`RUN2`), a merged branch with a dirty worktree (`REFUSE`), unmerged
      and dirty (`RUN1`, the ordinary in-flight state — it must not prompt on every finish), and a
      shortened recorded sha comparing correctly against a full `HEAD`.

### Session-record preservation — `scripts/preserve-session-records.sh`

Exercise it against a scratch copy so the repository tree is untouched:

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-19-finish-safety-records-and-effort
SCRATCH="$(mktemp -d)"; mkdir -p "$SCRATCH/wt/.superpowers/sdd/tasks" "$SCRATCH/state"
cp .superpowers/sdd/tasks/progress.md "$SCRATCH/wt/.superpowers/sdd/tasks/progress.md"
./scripts/preserve-session-records.sh "$SCRATCH/wt" demo "$SCRATCH/state"
find "$SCRATCH/wt/docs/superpowers" -type f
```

- [ ] The ledger is reported as `preserved:` and lands at
      `docs/superpowers/ledgers/<date>-demo.md`.
- [ ] The two absent sources are each reported as `skipped:` on their own line, and the exit status
      is still 0 — a missing source is never fatal.
- [ ] Running it a second time overwrites in place: still exactly one ledger file, not one dated
      duplicate per round.
- [ ] Replace `"$SCRATCH/wt/.superpowers/sdd/tasks/progress.md"` with a symlink to any file outside
      `$SCRATCH/wt` and re-run: the ledger is **refused** — a message on stderr naming the source and
      what it resolved to, a non-zero exit, and no copy of that file's content under
      `docs/superpowers/`. It is not reported as `skipped:`; that word means only "absent".
- [ ] Invoke it with the change name `'*'`: it exits 2 with a message about the name and creates
      nothing. A name outside the allowlist is rejected before any directory is touched.
- [ ] `./scripts/test-preserve-session-records.sh` prints `all cases pass`, including the case that
      a *different* change's preserved record is never adopted and overwritten, the source- and
      destination-containment refusals, and the change-name allowlist.

### The contract text — read it in place

- [ ] `skills/myflow-contracts/pipeline.md`, `## Finish contract`: the three signals are listed in
      order, signal 1 (HEAD against the recorded merge base) precedes signal 2 (the ancestor test),
      and the text says why. Substituting a commit count is explicitly forbidden.
- [ ] Signal 2 describes what the script actually does — `git merge-base --is-ancestor` and nothing
      else. The PR-CLI option appears only in the script-absent fallback, attributed to the human
      doing the check by hand.
- [ ] A `REFUSE` stops before anything is touched and asks; a multi-repo change reaches run 2 only
      when **every** worktree returns `RUN2`.
- [ ] `### Run 1 — the branch is not merged`: all routes commit the session records under
      `docs/superpowers/` alongside the implementation, the manual test guide and the `openspec/`
      artifacts, and the records are copied **before** staging.
- [ ] `### Run 2 — the branch is merged`: the new step removes the proposal artifact source from the
      state directory — **only** when a preserved copy exists under `docs/superpowers/artifacts/` —
      and it sits after the worktree cleanup and before the `FINISHED` write. The surrounding
      numbering is intact.
- [ ] `skills/myflow-finish/SKILL.md` **points at** the contract for all of the above and restates
      no procedure. Its guardrails forbid deciding run 1 versus run 2 from the ancestor test or a
      commit count alone, and forbid deleting the artifact source without a preserved copy.
- [ ] `skills/myflow-do/SKILL.md`, the "**The one commit exception.**" path: the preserve call
      happens **before** the staging that the commit picks up. (If it ran after `git add -A`, the
      records would miss the commit.)

### The effort field

- [ ] `skills/myflow-contracts/state-file.md`: `effort` appears in the JSON shape and in the field
      list, documented as governing `/myflow-start`'s own reasoning depth and nothing else.
- [ ] The same file states that a state file **omitting** `effort` is valid and reads as `null`, and
      explains why that is different from the *present and nullable* `artifactUrl` / `jiraIssue` /
      `prUrl`. Without this, every state file written before the change would be routed through
      self-heal.
- [ ] `state-file.md` `## Effort`: three levels with `medium` the default, and the rule that **no
      level may switch a gate off** — brainstorming, the design approval gate, writing-plans, and a
      fully enriched `tasks.md` all still happen at `low`.
- [ ] `skills/myflow-contracts/state-self-heal.md` names `effort` as the one documented exception to
      the closed schema, and includes it in the fields a rewrite re-emits as read.
- [ ] `skills/myflow-start/SKILL.md`: the ask happens only on the run that **creates** the change
      (the state file does not exist), uses AskUserQuestion, is never an argument, and a revision
      round reads the recorded level and says which one it is reusing. The level table is not
      restated here — it points at the contract.
- [ ] `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md` and `skills/myflow-status/SKILL.md`
      all carry `effort` forward the way they carry `jiraIssue`.

### Declared configuration and the docs

- [ ] `.myflow/project.md` `## test` lists all five harnesses; `## lint` still lists the three
      guards and explains why the two new scripts are tested but not linted (both need a worktree, a
      branch and a resolved base ref).
- [ ] `.myflow/project.md` `## lint`'s note about `check-plan-provenance.sh` matches reality — it
      exits **0** now that `kan-8-myflow-updates` is archived.
- [ ] `CLAUDE.md` and `AGENTS.md` describe the effort ask, the preflight verdict and the
      artifact-source removal, and remain in step with each other.

### The change itself

- [ ] `openspec validate kan-19-finish-safety-records-and-effort --strict` reports the change is
      valid.
