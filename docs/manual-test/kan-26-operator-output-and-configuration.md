# Manual test — operator-facing output and configuration (KAN-26)

This repository declares **no runnable application** in `.myflow/project.md`: it is the source of the
myflow skills, commands and rules, installed elsewhere by `setup.sh`. There is nothing to start, no
port and no URL. Each check below is therefore the command to run, or the file to read.

## How to run what is in scope

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration

# the three lint guards
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh

# the seven assertion harnesses
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
./scripts/test-check-finish-preflight.sh
./scripts/test-preserve-session-records.sh
./scripts/test-check-unfinished-work.sh
./scripts/test-check-cleanup-complete.sh

# the installer, against a throwaway HOME — never the real one
SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global
```

Open the worktree:

```bash
open -na "IntelliJ IDEA" --args "/Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration"
```

## Verification

- [ ] all three guards exit 0
- [ ] all seven harnesses exit 0
- [ ] the sandboxed installer exits 0

## myflow-planning-effort

- [ ] `skills/myflow-contracts/state-file.md` has a `## Planning effort` section whose table offers `low` / `default` / `detailed`, with `default` named as the recommendation
- [ ] no `"effort":` key survives — `grep -rn '"effort":' skills rules commands commands-claude scripts README.md AGENTS.md CLAUDE.md` returns nothing
- [ ] a state file carrying the retired `effort` key is read as its equivalent level (`medium` → `default`, `high` → `detailed`, `low` → `low`), is not unparseable, and gets no announced correction
- [ ] every consumer performs the fallback — `/myflow-status`'s jq reads `(.planningEffort // .effort)`, and `/myflow-start`, `/myflow-do` and `/myflow-finish` each say the level is carried forward as the **mapped** level
- [ ] a file carrying **both** keys resolves to `planningEffort`, and a retired-key value outside the mapped three reads as **not recorded** rather than making the file unparseable
- [ ] the two protected false positives are untouched — `jira-integration.md`'s "best-effort reconstruction" and `test-preserve-session-records.sh`'s `kan-19-…-and-effort` slug
- [ ] the guard's comment claims exactly what the guard does, and no more: it matches a **quoted** key (`"effort"` or `'effort'`) that either stands where a field stands (line start, after `{`, or straight after an opening backtick) or carries a JSON value (`"…"`, `null`, `<…>`), and it says plainly that an **unquoted** key is matched by neither alternative and never was
- [ ] appended to a **sandbox copy** of the scan set, one line at a time: `{ "effort": null }`, `{ "effort" : null }`, `{ 'effort': null }`, `"effort": "low"`, `{ "a": 1, "effort": null }` and `` The `"effort": null` entry `` each exit 1, while `effort: low`, `- effort: low`, `` `effort: null` ``, `` the `effort` key `` and `Three settings exist, 'effort': low, medium, high are the names.` each exit 0 — the last of those being the false positive round 7 removed
- [ ] no `vocab-guard:allow` marker was added to silence anything, and the entry claims no completeness the header does not already limit

## myflow-state-machine

- [ ] `state-file.md`'s JSON example carries `planningEffort` and a `models` object with `implementation`, `reviewPanel`, `panelFix`
- [ ] an **absent** `planningEffort` or `models` reads as "not recorded" and does not make the file unparseable
- [ ] `state-self-heal.md`'s "**There is no legacy-value migration**" is scoped to the retired `stage` field, with the collapse-vs-rename distinction stated, and lists both new fields among the fields a rewrite carries forward
- [ ] a self-heal rebuild recovers the `worktrees` **keys** from `git worktree list` rather than emptying the map, records `null` where the merge base cannot be recovered, and names `worktrees` in the announcement as recovered without merge bases

## myflow-model-policy

- [ ] `pipeline.md`'s `## Model policy` names the three roles and their defaults, with panel fixes defaulting to the strongest model rather than Sonnet
- [ ] `skills/myflow-start/SKILL.md` asks three separate model questions on the creating run only, each naming its default as the recommendation
- [ ] `skills/myflow-do/SKILL.md` dispatches implementers on `models.implementation` and the panel on `models.reviewPanel`, each with its documented default
- [ ] Bugbot and Security Review still take no model override and still record `unknown (agent-defined)`

## myflow-review-panel-economics

- [ ] Sonnet is stated as the panel's **default** rather than an absolute, pointing at the recorded override
- [ ] the slot roster and the optional-slot trigger table are unchanged

## myflow-handoff-output

- [ ] `pipeline.md` carries a per-state handoff template under `### The block each state renders`
- [ ] `IN_PROGRESS` has two templates — one for a change after `/myflow-do`, one for a branch after `/myflow-finish` run 1
- [ ] a value the state file does not carry is reported as **missing**; a **run-only** value is omitted instead, and the two are distinguished
- [ ] `skills/myflow-status/SKILL.md` regenerates the block for a named change and stores nothing
- [ ] `/myflow-status` with no argument still prints the table, and `FINISHED` changes stay omitted
- [ ] the three producing skills **cite** the template rather than carrying uncited copies of it
- [ ] **merge status governs when it is known** — a change whose branch is already merged does not get a "waiting on the merge" block from the same invocation whose table says "it will archive"
- [ ] the block never instructs the operator to run a command that cannot work — no `git diff --cached` for a branch already committed — and the `Git:` field has a true option on every landing route
- [ ] the `STARTED` template does not promise a Jira issue URL
- [ ] `/myflow-start`, `/myflow-do` and `/myflow-finish` print `/rename <change-name>` and `/color cyan` after their announcement; `/myflow-status` and `/myflow-info` print neither
- [ ] the contract records **why** those two commands are printed rather than invoked
- [ ] the `STARTED` template's **Jira** line is marked `(run-only)`, and the contract says why `/myflow-status` cannot reproduce it — the state file records no transition history and the report may not call Jira
- [ ] a branch landed by **handle it manually** is not rendered as a staged diff awaiting review — merge status alone decides that row, not `prUrl`

## myflow-progress-visibility

- [ ] `pipeline.md` has a `## Progress visibility` section stating the rule harness-neutrally, with a per-command step mapping
- [ ] the task list is described as a **view, never a record**, and `tasks.md` remains the single source of completion state
- [ ] no third checkbox marker was added to `tasks.md`
- [ ] `/myflow-status` and `/myflow-info` register nothing

## myflow-manual-test-guide

- [ ] `skills/myflow-do/SKILL.md` section 6 describes the guide as a capability-scoped behaviour checklist, not a per-task transcript
- [ ] it states that the checkbox syntax and the known-incomplete section are unchanged because `scripts/check-unfinished-work.sh` reads both
- [ ] it carries the no-runnable-application clause
- [ ] **this guide** is itself an instance of the new register — read it and judge whether it is what you asked for

## myflow-contract-distribution

- [ ] `pipeline.md` has a `## Pipeline flow` section with the state diagram, a level-1 row per command, and eight level-2 expansions
- [ ] `README.md` contains no `stateDiagram-v2` block and points at `pipeline.md` instead
- [ ] no level-2 expansion restates a tuned threshold another file owns
- [ ] `skills/myflow-info/SKILL.md` may present the diagram and stage table it read during the invocation, never a remembered one

## myflow-jira-projection

- [ ] `jira-integration.md` has a `### Follow-up issues` section giving the title rule `<KEY> follow-up`, or `myflow follow-up` with no linked issue
- [ ] the join search enumerates exactly `To Do` and `TO DO URGENT` by name and states it is **not** derived from `statusCategory`
- [ ] **joining asks first** — the candidate's key, title and status are shown and only an explicit yes joins; declining files a new follow-up
- [ ] the joined issue's description is **not** echoed verbatim into the handoff
- [ ] a failed **search** has a defined outcome, and it is not "no match, therefore create"
- [ ] a partial join is reported as such, not as success
- [ ] the retry guard re-attempts the writes not yet satisfied, and matches independent of date
- [ ] the join search carries an explicit **project constraint**, resolved the same way the filing site resolves it; a search that cannot name a project is not performed
- [ ] the contract states how the `## jira` body is split into candidate keys, and requires `[A-Z]{2,10}` of **the whole** candidate rather than of a substring — its own counter-example `KAN" OR project != "KAN` is shown failing under that reading, and `none` and the several-keys case both still work
- [ ] the candidate's title is sanitised in four **ordered** steps — `Cc`/`Cf` fold, whitespace collapse, fence-delimiter fold, truncate — and the contract states the order is itself a requirement
- [ ] a crafted title cannot close the block it is displayed in, and the stated reason the fold runs last holds: neither earlier step deletes anything to zero width, so neither can manufacture a delimiter run behind it
- [ ] a change does **not** retitle its own follow-up
- [ ] a partial append followed by a failed retitle has a reported form — the row is chosen by the later writes, the clause reports what the append did

## Known incomplete

**Nine findings are open: one Critical, four Important, four Minor.** The review panel ran eight full
passes and seven fix rounds; 86 findings were raised, 75 fixed and 2 withdrawn by the operator.

**The run was stopped deliberately at the operator's direction**, who chose to hand off with these
nine open and carry them into a Jira follow-up at `/myflow-finish` rather than open an eighth fix
round. The escalation ladder allows five rounds; it was already overridden twice. Findings per pass —
21, 10, 8, 9, 12, 10, 7, 9 — did not converge, and this pass's Critical was **introduced by the
previous round's fix**. Read `.superpowers/sdd/final-review-panel.md` before deciding whether to
integrate.

**Everything mechanical passes**: all three lint guards, all seven test harnesses, and the
change-scoped `openspec validate` are green, and three of seven panel slots returned clean.

### Open Critical

- **F80 — the Jira project-key check does not defeat the attack it names.** `jira-integration.md:280`
  names "a bare `OR` between two keys" as the threat. `:283-286` drops failing keys **per candidate**,
  treating the section as naming none only if *every* key is dropped. Measured: `OR` fullmatches
  `[A-Z]{2,10}`, so it survives and is accepted as a project key. The worked example claiming "not one
  of which is uppercase letters and nothing else" is false, and `:307-310` states the same rule as
  all-or-nothing, contradicting `:283-286`. **No injection is achievable** — a letters-only value
  cannot escape the quotes it is interpolated into — so the failure is scope integrity, gated behind
  the join confirmation prompt. Fourth round on this clause.

### Open Important

- **F79** — F71's disclosure reached two of the three trigger files. `rules/myflow-manual-review.mdc`
  is unchanged across all seven rounds, and it is `setup.sh`'s source for the **global** managed
  blocks and Cursor's only rule layer. Proven by running the installer into a throwaway `$HOME`.
- **F81** — `check-vocabulary.sh` alternative 2 prefix-matches `null` and `<`, so
  `"effort": nullify the previous value` trips it, contradicting round 7's "and nothing more" claim.
- **F83** — `specs/myflow-jira-projection/spec.md:50` still folds only `Cc`/`Cf`; the contract has
  folded `Cc`, `Cf`, `Zl` and `Zp` since round 6. The spec archives verbatim.
- **F84** — a `## jira` body mixing `none` with a real key is undefined, and the two readings scope
  the search to different projects.

### Open Minor

- **F78** — the F71 disclosure runs five near-duplicated sentences where one or two would carry it.
- **F82** — the guard's backtick anchor matches after any backtick, not only an opening one.
- **F85** — a `verified:` tag claims 24 payloads for a 23-row table.
- **F86** — a `verified:` tag claims commands the block does not run.

### Carried forward, not fixed

- **`commands/myflow-info.md:18`** and its `commands-claude` twin (line 14) describe the input as a
  change name, while `skills/myflow-info/SKILL.md:81` says the only argument is an optional *topic*.
- **`skills/README.md:36-37`** carries two near-duplicate "See Model policy" citations. Predates this
  change.

### Accepted residuals, not defects

- **The Jira join's forged-section residue.** Provenance is unobtainable from the tools available;
  the contract states this and bounds it — the outstanding list reaches the planning commit's message
  and the handoff on every route, so a forgery costs the tracker copy, never the record.
- **The vocabulary guard does not catch unquoted spellings**, including this repository's own
  bare-backtick house style. Never did; the claim was corrected rather than the pattern widened,
  because unquoted `effort` is the ordinary English word.
- **The trigger digest paraphrases the canonical block** rather than mirroring it, so `/myflow-info`
  and the session-start file read differently. Disclosed in two of three copies — see F79.
- **A partial join does not survive the merge.** The retry lives in run 1 only.
- **Multi-repo self-heal recovers one repository.** A rebuild cannot read the `worktrees` key set
  that records the others.
- **A partial `models` object is undefined** — no writer produces one.
- **The Jira lost-update window is narrowed, not closed.**
- **Nothing mechanically detects handoff-template drift**; adding no new guard script is a declared
  non-goal.
- **`check-references.sh` has a blind spot** — a citation wrapped across two physical lines is
  silently skipped.
