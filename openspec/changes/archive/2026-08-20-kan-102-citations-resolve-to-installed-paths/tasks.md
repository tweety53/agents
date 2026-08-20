> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Every path citation in a file `setup.sh` installs or copies names the root it resolves
against, and a guard that derives "installed" from the installer itself keeps it that way.

**Architecture:** One new guard — a Python classifier (`scripts/check-installed-citations.py`) behind
a thin shell wrapper (`scripts/check-installed-citations.sh`) — plus its fixture harness. The guard
learns the installed set by running `setup.sh` into a sandbox and reading back what appeared, so it
cannot drift from the installer. `scripts/check-references.sh` gains prefix stripping so the rewrite
does not silently delete its coverage. Then 200 citations across 31 installed files are rewritten in
three waves, and the guard is registered in `.myflow/project.md`.

**Tech Stack:** Bash (wrapper, harness; Bash 3.2 is the floor, as
`scripts/test-check-finish-preflight.sh`'s header records for this repository) and Python 3 standard
library only, via `/usr/bin/python3` — the toolchain `scripts/check-plan-provenance.py` already
established and `.myflow/project.md` records. Nothing is added to either.

**Spec:** `openspec/changes/kan-102-citations-resolve-to-installed-paths/design.md` and the three
delta specs beside it under `specs/`. The approved design is
`docs/superpowers/specs/2026-08-20-kan-102-citations-resolve-to-installed-paths-design.md`.

## Global Constraints

- **The three recognised roots are fixed by the spec and are not extended by any task.** Installed
  roots (`skills/`, `rules/`, `commands/`, `commands-claude/`, `hooks/`) stay bare; `<agents repo>/`
  names this checkout; `<project>/` names the project the command runs against. A task that invents a
  fourth prefix has exceeded its scope.
- **`skills/…` citations are never rewritten.** All 554 of them keep their bare form. This is
  decision `installed-roots-stay-bare` in `design.md`; prefixing them would delete
  `scripts/check-references.sh`'s entire coverage.
- **The installed set is derived by running `setup.sh`, never by re-implementing its globs or by a
  written list.** This is decision `derive-by-running-setup`. A task that hardcodes an installed path
  has exceeded its scope.
- **The rewrite is a judgment call per site, not a `sed` over a root map.** Which prefix a citation
  takes is decided by what its sentence means. `openspec/specs/myflow-model-policy/spec.md` is this
  repository; `openspec/changes/<name>/` is the target project; both begin `openspec/`.
- **No task edits `openspec/` or `docs/superpowers/`.** Those are the planning paths
  `/myflow-do` never stages.
- **A guard is named by basename in every invoking position**, never as a repository-relative
  `scripts/<name>` path — `scripts/check-guard-symlinks.sh`'s rule 3 rejects the latter.
- **A guard does not derive a repository root as a fixed depth above itself**
  (`scripts/check-guard-symlinks.sh`'s rule 4). Derive it from the resolved physical location.
- **Every grep in a new guard carries `-a` and `--`, and splits `rc > 1` from `rc == 1`**, per the
  three disciplines `scripts/check-guard-symlinks.sh`'s header records.
- **A refusal writes to stderr and nothing to stdout.** No caller may read an absent report as a
  clean one.

## Baseline

All measured 2026-08-20 against `332e593`.

- `scripts/check-references.sh` reports `check-references: all referenced sections resolve`.
  <!-- measured: scripts/check-references.sh @ 332e593 -->
- `scripts/check-guard-symlinks.sh` reports
  `GUARD-SYMLINKS-OK: /Users/tweety53/Projects/agents — 60 guard(s) across 6 skill(s) validated`.
  <!-- measured: scripts/check-guard-symlinks.sh @ 332e593 -->
- `scripts/check-contract-budget.sh` reports `BUDGET-OK: 27 contract file(s) within budget`.
  <!-- measured: scripts/check-contract-budget.sh @ 332e593 -->
- `scripts/check-vocabulary.sh` reports `✓ Panel-vocabulary guard: clean`.
  <!-- measured: scripts/check-vocabulary.sh @ 332e593 -->
- 236 citations across 32 installed files name no root: 41 in the copied root files and always-on
  rules, 121 under `skills/myflow-contracts/`, 74 in the command skills and `commands/`.
  <!-- measured: scripts/check-installed-citations.sh @ 070179e, the guard task 2 built; it supersedes the planning session's throwaway probe, which reported 200 across 31 files because it filtered path-shaped tokens more aggressively than the real classifier does -->
- `skills/myflow-do/SKILL.md` is 70768 bytes against a budget row of 71317 — **549 bytes of
  headroom**, the tightest row this change touches. Its 32 sites add about 355 bytes.
  <!-- measured: wc -c skills/myflow-do/SKILL.md and the budgets table in scripts/check-contract-budget.sh @ 332e593 -->
- `skills/myflow-contracts/pipeline.md` is 46216 bytes against 55728 — 9512 bytes of headroom, and
  its 35 sites add about 435 bytes.
  <!-- measured: wc -c skills/myflow-contracts/pipeline.md and the budgets table in scripts/check-contract-budget.sh @ 332e593 -->
- A sandboxed `HOME="$SANDBOX" ./setup.sh global` completes in 0.588s and produces 54 symlinks.
  <!-- measured: time ( SB="$(mktemp -d)"; HOME="$SB" ./setup.sh global >/dev/null 2>&1; find "$SB" -type l | wc -l ) @ 332e593 -->
- `scripts/check-installed-citations.sh`, `scripts/check-installed-citations.py` and
  `scripts/test-check-installed-citations.sh` do not exist.
  <!-- measured: ls scripts/check-installed-citations.sh scripts/check-installed-citations.py scripts/test-check-installed-citations.sh @ 332e593 -->

---

### 1 `scripts/test-check-installed-citations.sh` — the harness, before the guard exists

**Build:** red

**Squash-with:** Task 2

**Files:**
- Add: `scripts/test-check-installed-citations.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable definition of the guard's contract — its exit codes, its report shape, and
  one named case per classifier decision. Task 2 satisfies it; every later task depends on it holding.

The guard's whole value is that its classifier draws two boundaries correctly: what counts as a
citation, and what counts as a root. Written after the guard, a harness asserts what the guard
happens to do. Written first, it asserts what `specs/myflow-citation-roots/spec.md` says it must do.

Model it on `scripts/test-check-guard-symlinks.sh`: a fixture tree under `TMPDIR`, one `ok:`/`FAIL:`
line per assertion, a `check-installed-citations: all cases pass` line at the end, and a non-zero
exit on any failure.

- [x] **Step 1: Build the fixture-tree helper**

A helper that materialises a miniature agents repository under `TMPDIR` — a real `setup.sh` capable
of installing, a `skills/<name>/SKILL.md`, an always-on rule, an opt-in rule, a `CLAUDE.md`, and a
`README.md` at its root — then points the guard at it through
`CHECK_INSTALLED_CITATIONS_ROOT`. The fixture carries its own `setup.sh` rather than the real one, so
a case can change what is installed without touching this repository.

```bash unverified:confirm scripts/test-check-guard-symlinks.sh's fixture helper is named make_fixture_repo and can be followed rather than re-derived
# follow the existing helper's shape; do not invent a second fixture idiom
grep -n 'fixture' scripts/test-check-guard-symlinks.sh | head -20
```

- [x] **Step 2: Write the classifier cases**

Each case is one fixture file carrying one citation, and one assertion on whether the guard reports
it.

| Case | Citation in an installed fixture file | Expects |
|------|----------------------------------------|---------|
| `bare-root-file` | `` `README.md` ``, with `README.md` present at the fixture root | reported |
| `bare-generic-filename` | `` `tasks.md` ``, with no `tasks.md` at the fixture root | not reported |
| `installed-root` | `` `skills/other/SKILL.md` `` | not reported |
| `agents-repo-prefix` | `` `<agents repo>/README.md` `` | not reported |
| `project-prefix` | `` `<project>/.myflow/project.md` `` | not reported |
| `unrooted-directory` | `` `.myflow/project.md` `` | reported |
| `unrooted-script` | `` `scripts/check-vocabulary.sh` `` | reported |
| `fenced-command` | `openspec/` on a non-comment line inside a ```bash fence | not reported |
| `fenced-comment` | `# see openspec/` inside a ```bash fence | reported |
| `absolute-path` | `` `/etc/hosts` `` | not reported |
| `home-path` | `` `~/.claude/skills/` `` | not reported |
| `url` | `` `https://example.test/a/b` `` | not reported |
| `shell-variable` | `` `$SCRIPT_DIR/lib/x.sh` `` | not reported |
| `git-ref` | `` `origin/main` `` | not reported |
| `regex-fragment` | `` `[A-Za-z0-9._-]+/x` `` | not reported |
| `opt-in-rule-out-of-scope` | `` `README.md` `` inside a rule the fixture's `setup.sh` does not install | not reported |
| `newly-installed-directory` | `` `README.md` `` inside a fixture file under a directory the fixture's `setup.sh` is edited to start installing | reported, with no edit to the guard |

`newly-installed-directory` is what proves the derivation is real rather than a re-implementation.
The case edits the **fixture's** `setup.sh` to install a directory it did not before, and asserts the
guard picks the new files up. A guard carrying its own copy of the installer's globs passes every
other case in this table and fails this one.

`fenced-comment` is the case that keeps the fence exclusion honest. A comment inside a fence is prose
about a path, not a shell argument, so excluding the whole fenced block wholesale would let a real
citation hide there.

- [x] **Step 3: Write the refusal cases**

| Case | Expects |
|------|---------|
| `CHECK_INSTALLED_CITATIONS_ROOT` set but empty | exit `2`, stderr names it, stdout empty |
| root is not a directory | exit `2`, stderr names it, stdout empty |
| the fixture's `setup.sh` exits non-zero | exit `2`, stderr names the failed install, stdout empty |
| the guard is asked to install with `HOME` outside its own sandbox | exit `2`, no `setup.sh` invocation |

Every refusal case asserts **stdout is empty as well as** the exit code. A guard that refuses while
printing a partial report is the failure this discipline exists to stop.

- [x] **Step 4: Write the report-shape cases**

| Case | Expects |
|------|---------|
| a clean fixture | exit `0`, a verdict line naming the root and the file count |
| a fixture with two violations | exit `1`, two `path:line: message` lines, a verdict naming `2` |
| any run | a `scripts/lib/coverage.sh` per-member fragment on the OK line |

- [x] **Step 5: Mutation-prove every classifier case**

Per KAN-197, a case that still passes with the line it covers removed proves nothing. Once task 2
lands, delete each classifier exclusion in turn — the absolute-path test, the URL test, the shell
variable test, the git-ref test, the regex test, the fence test, the slash-less resolution test — and
confirm the matching case fails. Restore, and record each mutation in this task's SDD ledger entry.

- [x] **Step 6: Verify**

```bash verified:run it and read the failure — the guard does not exist yet, so every case must fail for that reason and no other
bash scripts/test-check-installed-citations.sh; echo "exit=$?"
```

**Tests:** the cases enumerated in steps 2, 3 and 4, in `scripts/test-check-installed-citations.sh`.

**Regression:** Reverting this task leaves the guard with no executable statement of its contract.
Its classifier is the whole change — a boundary that moves silently turns the guard either into
noise nobody reads or into a pass that checks nothing, and only a named case per decision catches
that.

**Baseline:** before=0 after=23 cases in `scripts/test-check-installed-citations.sh`.
<!-- measured: bash scripts/test-check-installed-citations.sh @ 070179e — 17 classifier, 4 refusal, 2 report-shape -->

**Commit:** `test(kan-102-citations-resolve-to-installed-paths): pin the citation classifier's two boundaries and the guard's exit contract`

---

### 2 `scripts/check-installed-citations.py` and its wrapper — the guard

**Build:** green

**Files:**
- Add: `scripts/check-installed-citations.py`
- Add: `scripts/check-installed-citations.sh`
- Add: `scripts/test-check-installed-citations.sh` — task 1's file. Task 1 is `Build: red` with
  `**Squash-with:** Task 2`, so its commit folds into this one and this commit is the squash unit's
  only commit. The field declares what the commit touches, which is all three files.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's harness, which defines the contract this task satisfies.
- Produces: `scripts/check-installed-citations.sh`, argument-free, honouring
  `CHECK_INSTALLED_CITATIONS_ROOT`, exiting `0` clean / `1` violations / `2` cannot answer. Tasks 4,
  5 and 6 run it to watch their own violation count fall; task 8 registers it.

- [x] **Step 1: Write the wrapper**

A thin `exec`, matching `scripts/check-plan-provenance.sh`. It resolves its own physical location —
never a fixed depth above itself, per `scripts/check-guard-symlinks.sh`'s rule 4 — and hands the root
to the Python.

```bash unverified:confirm scripts/check-plan-provenance.sh's wrapper shape, and copy it rather than inventing a second one
cat scripts/check-plan-provenance.sh
```

- [x] **Step 2: Derive the installed set by running the installer**

Create one sandbox directory. Refuse to invoke `setup.sh` unless both the `HOME` it will be given and
the project directory it will be given lie inside that sandbox — the refusal
`scripts/test-setup.sh` already implements, adopted rather than restated. Run the installer for
`global` and for `all`, then walk the sandbox: every symlink resolves to a path inside the repository
root, which is that file's repository-relative source; every regular file that matches a file at the
repository root by content is a copy.

The installed **roots** are the first path segment of each derived source path.

```bash verified:this is the sandbox invocation measured in the Baseline above, 0.588s and 54 symlinks for the global mode
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" "$REPO_ROOT/setup.sh" global
HOME="$SANDBOX" "$REPO_ROOT/setup.sh" all "$SANDBOX/proj"
```

An installer run that exits non-zero is exit `2`, not an empty installed set. An empty set would make
every citation a violation and the report unreadable; worse, an empty *root* set would make every
citation pass.

- [x] **Step 3: Write the citation classifier**

Parse each installed `.md`/`.mdc` as a block structure first, then as tokens — the ordering that
`scripts/check-plan-provenance.py`'s docstring records as the reason it is not a regular expression.

Exclude, in this order: a non-comment line inside a fenced `bash`/`sh`/`zsh` block. Then, per
backticked token, exclude an absolute path, a `~`-rooted path, a URL, a token whose first segment is
a shell variable reference, a git ref shape, and a regular-expression or glob fragment.

```python unverified:confirm the exact exclusion set against scripts/check-plan-provenance.py's own classifier before copying its idiom
# A token with no "/" is a citation only when it names a real file at the repository root.
# This is what catches `README.md` and `setup.sh` without flagging `SKILL.md` or `tasks.md`,
# which are shapes rather than files. See specs/myflow-citation-roots/spec.md.
if "/" not in token:
    return (repo_root / token).is_file()
```

- [x] **Step 4: Judge each citation against the roots**

A citation passes when its first segment is an installed root, or when it begins with the literal
`<agents repo>/` or `<project>/`. Anything else is a violation, reported as `path:line: message`
naming the token and both prefixes as the fix.

- [x] **Step 5: Report and exit**

One violation line per finding, then a verdict line naming the root and the count, then the
`scripts/lib/coverage.sh` per-member fragment — the discipline
`openspec/specs/agents-repo-verification/spec.md` already requires of a corpus-scanning guard. Exit
`0` clean, `1` violations, `2` cannot answer, with a refusal writing to stderr and nothing to stdout.

- [x] **Step 6: Verify the harness passes**

```bash unverified:the case count is task 1's predicted 21; confirm against the harness's own tail line
bash scripts/test-check-installed-citations.sh; echo "exit=$?"
```

- [x] **Step 7: Verify it is red against the real corpus**

The guard must find the defect it was built for. It is not registered in `## lint` yet — task 8 does
that, after tasks 4-6 drive the count to zero — so a non-zero exit here is the expected result.

```bash unverified:the exact count is the Baseline's 200, produced by a throwaway probe; this run is what makes it reproducible, and a materially different number means the classifier disagrees with the probe and must be reconciled before task 4
scripts/check-installed-citations.sh; echo "exit=$?"
```

**Tests:** task 1's cases, now passing, in `scripts/test-check-installed-citations.sh`.

**Regression:** Reverting this task leaves the convention declared in three delta specs and enforced
by nothing. Every citation the waves fix could be reintroduced by the next edit to any installed
file, which is exactly how 200 of them accumulated.

**Baseline:** before=0 after=236 violations reported against this repository's own corpus.
<!-- measured: scripts/check-installed-citations.sh @ 070179e; the planning probe's 200 was the estimate this run reconciled, and the guard's number stands -->

**Commit:** `feat(kan-102-citations-resolve-to-installed-paths): report every citation in an installed file that names no root`

---

### 3 `scripts/check-references.sh` — strip `<agents repo>/` before resolving

**Build:** green

**Files:**
- Modify: `scripts/check-references.sh`
- Modify: `scripts/test-check-references.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from tasks 1-2. This task is independent of the guard and deliberately lands
  **before** the rewrite.
- Produces: a `check-references.sh` that resolves a prefixed citation. Tasks 4, 5 and 6 depend on it:
  without this, every citation they prefix silently leaves that guard's coverage.

This task is ordered before the rewrite for one reason. If the waves land first,
`scripts/check-references.sh` keeps exiting `0` while checking strictly less than it did — a guard
that stops checking without failing, which is the worst failure shape a guard has, and one no
subsequent run would surface.

- [x] **Step 1: Add the failing harness cases**

Two cases in `scripts/test-check-references.sh`:

| Case | Expects |
|------|---------|
| a line referencing **Some section** in `` `<agents repo>/<fixture>.md` `` whose heading was removed | reported, exit non-zero |
| a line referencing a path beginning `<project>/` | not resolved, not reported, and **not** a containment refusal |

The second case is the one that matters most. `check-references.sh` treats a path normalising outside
the repository root as a **failure**, not a note — so a `<project>/` prefix read as a traversal would
turn every project-relative citation into a lint failure the moment task 4 lands.

- [x] **Step 2: Run them to verify they fail**

```bash unverified:confirm the harness reports per-case, so the two new cases are individually visible in its output
bash scripts/test-check-references.sh; echo "exit=$?"
```

Expected: the first case fails because the prefixed path does not resolve and is therefore skipped;
the second fails or errors depending on how the containment test reads the `<` in the prefix.

- [x] **Step 3: Strip the prefix before resolution**

Strip a leading literal `<agents repo>/` from a cited path, then resolve as before. The containment
test still runs on the **stripped** path, and still decides from shape before any existence test — the
guard's existing rule, unchanged.

Leave `<project>/` unstripped. It resolves to nothing, falls through the existing
does-not-resolve path, and is not checked — the spec's stated outcome.

- [x] **Step 4: Run the harness**

```bash unverified:confirm the harness's tail line reads "check-references: all cases pass" rather than another wording
bash scripts/test-check-references.sh; echo "exit=$?"
```

Expected: PASS, both new cases included.

- [x] **Step 5: Confirm the corpus is unchanged**

Nothing carries the prefix yet apart from the ten pre-existing `<agents repo>/` sites, so the guard's
verdict on this repository must be identical to the Baseline.

```bash verified:this is the Baseline's own invocation, re-run to compare against it
scripts/check-references.sh; echo "exit=$?"
```

**Tests:** `a prefixed citation is still checked` and `a project-prefixed citation is neither
resolved nor refused`, in `scripts/test-check-references.sh`.

**Regression:** Reverting this task means every `<agents repo>/`-prefixed `.md` citation the waves
create stops being checked for a live section, with no failure anywhere to say so. The guard's
per-member coverage report would show the drop only to a reader who compared two runs by hand.

**Baseline:** before=2 after=4 — the two new cases added to `scripts/test-check-references.sh`.
<!-- predicted: bash scripts/test-check-references.sh after this task; before=2 counts the two cases this task adds as absent, not the harness's full case count, which this task does not change otherwise -->

**Commit:** `fix(kan-102-citations-resolve-to-installed-paths): resolve a citation that names the agents repository explicitly`

---

### 4 Wave A — the copied root files and the always-on rules

**Build:** green

**Files:**
- Modify: `CLAUDE.md`
- Modify: `AGENTS.md`
- Modify: `rules/myflow-manual-review.mdc`
- Modify: `rules/lint-fix-priority.mdc`
- Modify: `rules/agent-baseline.md`
- Modify: `rules/dispatch-carries-the-baseline.mdc`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 2's guard, to count what is left; task 3's fix, so prefixed `.md` citations stay
  checked.
- Produces: 41 fewer violations. Nothing later depends on this wave's text.

This wave is first because it is the most severe. `CLAUDE.md` is **copied into the project root** and
auto-loads into every session there, so no command invocation is needed to reach a bad citation —
its bare `` `README.md` `` is read as canonical myflow documentation while naming a file a
contributor wrote. `AGENTS.md` is the same for Codex. The always-on rules are here for the same
reason: every one of them is loaded into every session of every project the installer touches.

Run `scripts/check-installed-citations.sh` first and work from its own report. It names every site
by `path:line`, and it is the authority on which sites exist — the planning probe that produced this
plan's earlier numbers was an estimate the guard has since superseded.

- [x] **Step 1: Rewrite the citations, deciding each by what its sentence means**

Not a `sed`. Two examples from these very files, both beginning with the same segment:

| Site | Means | Becomes |
|------|-------|---------|
| `CLAUDE.md`'s `` `README.md` `` in "`README.md`'s 'How the pipeline works' section" | this repository's README | `` `<agents repo>/README.md` `` |
| `CLAUDE.md`'s `` `.myflow/project.md` `` in "Full list in `.myflow/project.md`'s `## lint` section" | the project being worked on | `` `<project>/.myflow/project.md` `` |

`agents/rules/lint-fix-priority.mdc` in both root files names this repository's rule source and
becomes `<agents repo>/rules/lint-fix-priority.mdc` — the redundant `agents/` segment goes, because
the prefix now says what it was standing in for.

- [x] **Step 2: Run the guard and confirm the count fell by exactly this wave's share**

```bash unverified:the arithmetic assumes task 2's step 7 confirmed 200; if it found a different total, subtract this wave's own share from that total instead
scripts/check-installed-citations.sh; echo "exit=$?"
```

Expected: 195 violations remain, and none of them in these six files.

- [x] **Step 3: Run the reference guard**

```bash verified:the Baseline's own invocation, re-run so a prefixed citation that stopped resolving is caught here rather than three tasks later
scripts/check-references.sh; echo "exit=$?"
```

Expected: `check-references: all referenced sections resolve`.

- [x] **Step 4: Run the vocabulary and markdown guards**

```bash verified:both are in .myflow/project.md's ## lint list and both scan these files
scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py; echo "exit=$?"
```

**Tests:** none added. This wave is covered by the cases task 1 wrote and by the guard's own verdict
on the corpus, asserted in step 2.

**Regression:** Reverting this task restores the highest-severity half of the defect: a bare
`` `README.md` `` in a file that auto-loads into every session of every project it is copied into,
with no command invocation needed to reach it.

**Baseline:** before=236 after=195 violations.
<!-- measured: scripts/check-installed-citations.sh @ 070179e reported 41 across these six files — CLAUDE.md 12, AGENTS.md 20, myflow-manual-review.mdc 2, lint-fix-priority.mdc 2, agent-baseline.md 4, dispatch-carries-the-baseline.mdc 1 -->

**Commit:** `docs(kan-102-citations-resolve-to-installed-paths): name the root of every citation in the copied root files and always-on rules`

---

### 5 Wave B — the contract files

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `skills/myflow-contracts/project-configuration.md`
- Modify: `skills/myflow-contracts/project-configuration-rationale.md`
- Modify: `skills/myflow-contracts/plan-provenance.md`
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `skills/myflow-contracts/handoff-blocks.md`
- Modify: `skills/myflow-contracts/handoff-blocks-rationale.md`
- Modify: `skills/myflow-contracts/build-green.md`
- Modify: `skills/myflow-contracts/jira-integration.md`
- Modify: `skills/myflow-contracts/jira-followups.md`
- Modify: `skills/myflow-contracts/operator-prompts.md`
- Modify: `skills/myflow-contracts/SKILL.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 2's guard and task 3's fix, as wave A does.
- Produces: 121 fewer violations. Task 7 edits `pipeline.md` again, at a different section.

- [x] **Step 1: Rewrite the citations**

Same per-site judgment as wave A. The distinctions that recur here:

| Site | Means | Becomes |
|------|-------|---------|
| `openspec/specs/myflow-run-telemetry/spec.md` | this repository's spec | `<agents repo>/openspec/specs/…` |
| `openspec/changes/<name>/` | the target project's change directory | `<project>/openspec/changes/<name>/` |
| `stats/cmd/myflow/stage.go` | this repository's source | `<agents repo>/stats/cmd/myflow/stage.go` |
| `docs/superpowers/ledgers/` | the target project's planning directory | `<project>/docs/superpowers/ledgers/` |
| `scripts/check-stage-mark-calls.sh` in prose | this repository's guard | `<agents repo>/scripts/check-stage-mark-calls.sh` |

Two relative fragments in `pipeline.md` are completed as well as prefixed:
`internal/api/stages.go` becomes `<agents repo>/stats/internal/api/stages.go`, and
`lib/panel-record.sh` becomes `<agents repo>/scripts/lib/panel-record.sh`. Both are unresolvable as
written today, which is why neither has a first segment worth prefixing on its own.

- [x] **Step 2: Leave every `skills/…` citation alone**

There are hundreds of them in these files. They are already correct, and prefixing them would delete
`scripts/check-references.sh`'s coverage — Global Constraint 2, decision `installed-roots-stay-bare`.

- [x] **Step 3: Run the guard**

```bash unverified:the arithmetic assumes 200 at task 2 step 7 and 171 after task 4; carry forward whatever those runs actually reported
scripts/check-installed-citations.sh; echo "exit=$?"
```

Expected: 74 violations remain, and none of them under `skills/myflow-contracts/`.

- [x] **Step 4: Run the reference and budget guards**

```bash verified:both are Baseline invocations; the budget one matters here because pipeline.md is the largest file this wave touches
scripts/check-references.sh && scripts/check-contract-budget.sh; echo "exit=$?"
```

Expected: `all referenced sections resolve` and `BUDGET-OK: 27 contract file(s) within budget`.
`pipeline.md` gains about 435 bytes against 9512 of headroom.

**Tests:** none added; covered as wave A is.

**Regression:** Reverting this task restores the largest share of the defect and the one with the
longest reach — `pipeline.md` alone is loaded by every `/myflow-*` command, so each of its 35 bad
citations is read on every run of every command in every installed project.

**Baseline:** before=195 after=74 violations.
<!-- measured: scripts/check-installed-citations.sh @ 070179e reported 121 across these fourteen files, pipeline.md 36 and project-configuration.md 16 the largest -->

**Commit:** `docs(kan-102-citations-resolve-to-installed-paths): name the root of every citation in the contract files`

---

### 9 Three classifier exclusions the corpus proved necessary

**Build:** green

**Files:**
- Modify: `scripts/check-installed-citations.py`
- Modify: `scripts/test-check-installed-citations.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: tasks 2 and 5. Wave B is what proved these three shapes exist in the real corpus.
- Produces: a guard reporting 67 rather than 78, with every remaining site a real one. Task 6 drives
  that 67 to zero; task 8 cannot register the guard until it does.

Wave B rewrote 121 sites and left **four** the implementer refused to force, each a token the
classifier calls a citation that is not one. Forcing a root onto any of them would have written
something false into the corpus, so they were reported instead — correctly. But a guard that reports
four permanent false positives cannot be registered in `## lint`, so the classifier has to stop
calling them citations.

All four fall into three mechanical classes. **None of them is fixed with a suppression marker**: a
marker would silence a real hit and a false one identically, and this repository's lint rule forbids
adding one. These are classifier defects, and they are fixed in the classifier.

- [x] **Step 1: A `../`-rooted token is not a citation**

`pipeline.md:333`'s `../<other-app>` and `project-configuration.md:68`'s
`../../../../Users/victim/.ssh/config.mdc`. A path beginning `../` is relative to something the
document is not naming, so it **cannot** take a root prefix — there is no correct answer to write.
Both are also negative examples: the first sits in "never a relative path, never `../<other-app>`",
the second is a traversal-attack illustration in the containment discussion. Rooting either would
corrupt the thing it is demonstrating.

- [x] **Step 2: A token whose first segment is a RECOGNISED placeholder is already rooted**

`pipeline.md:525`'s `<state-dir>/<name>-proposal-artifact.html`. `<state-dir>` is the machine-level
myflow state directory — genuinely neither this repository nor the target project — and the
placeholder **is** the root. The corpus already uses this idiom widely: `<abs-worktree>/`,
`<changeRoot>/`, `<the running command's own skill directory>/scripts/`.

Generalise rather than special-casing `<state-dir>` — but generalise to a **closed set**, not to any
bracket shape. The recognised placeholder roots are enumerated in
`specs/myflow-citation-roots/spec.md`: `<agents repo>/`, `<project>/`, `<abs-worktree>/`,
`<changeRoot>/`, `<state-dir>/`. The first two are simply the ones this convention gives fixed
meanings to, which makes a coherent vocabulary rather than an exception list.

**Accepting any `<…>`-shaped first segment fails open**, and a guard that fails open is worse than no
guard, because it reports clean while checking nothing. `<foo>/openspec/specs/x.md` would pass while
naming an unrooted path, and so would a typo of a real placeholder — `<changeroot>/`,
`<change-root>/`, `<>/`. An unrecognised placeholder is a **violation**. Adding a placeholder root is
a deliberate act that extends the spec's table; it is not something an author does by accident.

- [x] **Step 3: A token carrying quote punctuation is quoted output, not a citation**

`finish-contract.md:500`'s `` `error: unable to delete 'openspec/<name>': remote ref does not exist` ``
— a git error message quoted verbatim, and labelled in the surrounding prose as measured against a
scratch remote on git 2.50.1. Prefixing the path inside it would misrepresent what git printed.

A real citation is never wrapped in `'` or `"` **inside** its own backtick span; a quoted fragment of
program output is. Exclude a token carrying a leading or trailing quote character.

- [x] **Step 4: Add one harness case per exclusion, and mutation-prove each**

Five cases: `parent-relative-path`, `placeholder-rooted`, `quoted-program-output`,
`unrecognised-placeholder-is-reported` — the fail-closed case, asserting `<foo>/openspec/specs/x.md`
**is** reported — and one asserting that a recognised placeholder-rooted token is **recognised**
rather than merely absent — the same
recognised-versus-never-seen distinction `agents-repo-prefix-bogus-path` already draws, because that
collapse is what hid this change's critical defect.

Then disable each of the three exclusions in turn and confirm exactly its own case fails. An
exclusion that no case catches is not load-bearing, and one that two cases catch is shadowing a
neighbour — the defect already found once in this guard, when `$` was matched by both the
shell-variable and the regex/glob rules.

- [x] **Step 5: Verify**

```bash verified:run at 405792b — the count is 67, not the 78-minus-4 the task first predicted, because seven further corpus sites already carried these same three shapes; every one of the eleven was traced individually and none is a classifier over-reach
scripts/check-installed-citations.sh; echo "exit=$?"
bash scripts/test-check-installed-citations.sh; echo "exit=$?"
```

Expected: 67 violations, none under `skills/myflow-contracts/`, and every harness case passing.

**Tests:** `parent-relative-path`, `placeholder-rooted`, `placeholder-rooted-is-recognised`,
`unrecognised-placeholder-is-reported` and `quoted-program-output`, in
`scripts/test-check-installed-citations.sh`.

**Regression:** Reverting this task leaves four permanent false positives in the guard's report.
Task 8 could then never register it in `## lint` without failing this repository's own lint on every
unrelated invocation — and the pressure would be to silence them with a marker, which is how a guard
stops distinguishing a real hit from a false one.

**Baseline:** before=78 after=67 violations.
<!-- measured: scripts/check-installed-citations.sh @ 405792b; the four sites wave B named are closed, and so are seven more the corpus already carried in the same three shapes — three `<abs-worktree>/` CONTEXT BUNDLE repeats, `<changeRoot>/specs/`, `<changeRoot>/tasks.md`, and two further `../` negative examples -->

**Commit:** `fix(kan-102-citations-resolve-to-installed-paths): stop classifying parent-relative, placeholder-rooted and quoted-output tokens as citations`

---

### 6 Wave C — the command skills and `commands/`

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`
- Modify: `skills/myflow-do/principles-reviewer-prompt.md`
- Modify: `skills/myflow-do/adversarial-reviewer-prompt.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-finish/SKILL-rationale.md`
- Modify: `skills/myflow-start/SKILL.md`
- Modify: `skills/myflow-start/SKILL-rationale.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `skills/myflow-status/SKILL.md`
- Modify: `skills/openspec-explore/SKILL.md`
- Modify: `commands/opsx-explore.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 2's guard, task 3's fix, and task 9's three classifier exclusions — without those,
  four false positives would remain and zero would be unreachable.
- Produces: zero remaining violations, which is what lets task 8 register the guard in `## lint`.
  Work from `scripts/check-installed-citations.sh`'s own `path:line` report, as waves A and B do.

- [x] **Step 1: Rewrite the citations**

Same per-site judgment. Two shapes recur only here:

| Site | Means | Becomes |
|------|-------|---------|
| `.superpowers/sdd/final-review.diff` in the reviewer prompts | a file inside the apply worktree | `<abs-worktree>/.superpowers/sdd/final-review.diff` — corrected by task 10; this wave first wrote `<project>/`, which is wrong because the directory is worktree-scoped |
| `specs/<capability>/spec.md` in `openspec-explore` and `commands/opsx-explore.md` | the target project's own OpenSpec specs | `<project>/openspec/specs/<capability>/spec.md` |

The second is completed as well as prefixed: `specs/<capability>/spec.md` is an OpenSpec-relative
fragment, and writing it out is what makes the prefix meaningful.

- [x] **Step 2: Watch `skills/myflow-do/SKILL.md`'s budget**

This is the tightest row the change touches: 549 bytes of headroom against about 355 added. Run the
budget guard immediately after editing that file, before editing the rest.

```bash verified:the Baseline's own invocation; this file is the one row where the outcome is genuinely in doubt
scripts/check-contract-budget.sh; echo "exit=$?"
```

If it trips, raise **that one row** in `scripts/check-contract-budget.sh`'s `budgets()` table to the
file's new size plus 25%, exactly as `.myflow/project.md` prescribes for a genuine addition — and add
`scripts/check-contract-budget.sh` to this task's `Files:` list when you do. Never narrow the guard's
scope and never delete a row.

- [x] **Step 3: Run the guard and confirm zero**

```bash unverified:the arithmetic assumes 69 remaining after task 5; carry forward whatever that run reported
scripts/check-installed-citations.sh; echo "exit=$?"
```

Expected: exit `0`, and a verdict line naming the file count with no violations. The 67 sites are
spread across exactly this task's twelve files — `myflow-do/SKILL.md` 30, `myflow-finish/SKILL.md` 8,
`myflow-do/SKILL-rationale.md` 5, `principles-reviewer-prompt.md` 5, and a tail of one to three each.
<!-- measured: scripts/check-installed-citations.sh @ 405792b, grouped by file -->

- [x] **Step 4: Run the guard half of the lint list**

The three `stats/` entries (`gofmt -l .`, `go vet ./...`, `npx tsc -b`) are omitted deliberately: no
task in this plan touches Go or TypeScript.

```bash verified:these are exactly the eleven guard entries in .myflow/project.md's ## lint block at 332e593, in order
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-provenance.sh \
  && scripts/check-task-build-green.sh && scripts/check-workspace-isolation.sh \
  && scripts/check-uitest-overrides.sh && scripts/check-contract-budget.sh \
  && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh \
  && scripts/check-guard-symlinks.sh && scripts/check-self-review-report.sh; echo "exit=$?"
```

**Tests:** none added; covered as waves A and B are.

**Regression:** Reverting this task leaves the guard unable to be registered — task 8's `## lint`
entry would fail the repository's own lint on every unrelated invocation, which is precisely the
condition `.myflow/project.md` gives for keeping a check out of that list.

**Baseline:** before=67 after=0 violations.
<!-- measured: scripts/check-installed-citations.sh @ 405792b reported 67 across these twelve files, myflow-do/SKILL.md 30 the largest -->

**Commit:** `docs(kan-102-citations-resolve-to-installed-paths): name the root of every citation in the command skills`

---

### 7 `pipeline.md` — replace Guard resolution's prose exemption

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: wave B, which already prefixed this file's other citations.
- Produces: a **Guard resolution** section consistent with the convention. Nothing later depends on
  it.

This is the one contract edit the change makes beyond mechanical prefixing, and it is separate from
wave B on purpose: a reviewer could reasonably approve 102 mechanical rewrites and still reject a
change to what a contract says.

- [x] **Step 1: Replace the exemption paragraph**

`pipeline.md`'s **Guard resolution** currently carries a paragraph beginning "The exemption is a
shape, not a list of names", which classifies a `scripts/<name>` citation by whether it sits in an
invoking position and lets everything else keep a repository-relative path. That classification is
what wave B removed the need for.

Replace it with the rule from `specs/myflow-contract-distribution/spec.md`: prose describing this
repository's own guards is not an invocation and names the guard as `<agents repo>/scripts/<name>`.
Keep the section's existing statements about basename resolution and about
`skills/myflow-contracts/` carrying no `scripts/` directory — neither changes.

- [x] **Step 2: Confirm `check-guard-symlinks.sh` rule 3 still holds**

Rule 3 rejects a repository-relative `scripts/<name>` in an **invoking** position, and that is
unaffected: an invoking position still uses the basename form, and a `<agents repo>/`-prefixed path
is not a bare repository-relative one.

```bash verified:the Baseline's own invocation, whose OK line names 60 guards across 6 skills
scripts/check-guard-symlinks.sh; echo "exit=$?"
```

- [x] **Step 3: Run the guard, references and budget**

```bash verified:the three guards that can respond to an edit of this file
scripts/check-installed-citations.sh && scripts/check-references.sh \
  && scripts/check-contract-budget.sh; echo "exit=$?"
```

Expected: exit `0` from all three.

**Tests:** none added. The rule this task states is asserted by the `unrooted-script` case task 1
wrote, which reports a bare `scripts/check-vocabulary.sh` regardless of position.

**Regression:** Reverting this task leaves `pipeline.md` telling a reader that a prose
`scripts/<name>` citation may keep its repository-relative path while the guard reports it — a
contract and its enforcement disagreeing, which per `pipeline.md`'s own framing is non-determinism in
the one layer that must be deterministic.

**Baseline:** before=1 after=0 paragraphs in `skills/myflow-contracts/pipeline.md` exempting prose
from the citation-root rule.
<!-- measured: grep -n "The exemption is a shape, not a list of names" skills/myflow-contracts/pipeline.md @ 332e593 gave one hit -->

**Commit:** `docs(kan-102-citations-resolve-to-installed-paths): drop Guard resolution's prose exemption for the citation-root rule`

---

### 10 Cross-wave root corrections — a syntactically valid root is not a correct one

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/principles-reviewer-prompt.md`
- Modify: `skills/myflow-do/adversarial-reviewer-prompt.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `skills/myflow-contracts/handoff-blocks-rationale.md`
- Modify: `skills/myflow-contracts/project-configuration.md`
- Modify: `scripts/check-installed-citations.py`
- Modify: `scripts/test-check-installed-citations.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: waves B and C, and task 9's classifier.
- Produces: a corpus where every root is *correct*, not merely present. Task 8 registers the guard
  after this.

Wave C's review found three classes of site that satisfy the guard and are still wrong. **The guard
reaching zero proves every citation acquired a root; it cannot prove any of them acquired the right
one.** These span two waves, so they are corrected once, here, rather than as fixups against three
commits.

- [x] **Step 1: A git branch name is not a filesystem path — five sites**

`skills/myflow-do/SKILL.md:115`, `skills/myflow-contracts/finish-contract.md:169,177,178` and
`skills/myflow-contracts/state-file.md:140` render the branch `openspec/<name>` as
`<project>/openspec/<name>`. `SKILL.md:115` is an **instruction** — "Branch
`<project>/openspec/<name>`" — that an agent would follow literally, creating a branch of that name.

Revert all five to bare `openspec/<name>`, then add one **exact-literal** exclusion for that token to
the classifier, extending the precedent `GIT_BRANCH_RE` already sets for `origin/main` and
`refs/heads/…`.

Two alternatives were tested and rejected on evidence, not taste:

- **Writing them as full refs** (`refs/heads/openspec/<name>`) is *actively wrong*, not merely
  verbose: `git branch "refs/heads/openspec/foo"` creates a branch literally named
  `refs/heads/openspec/foo`. It also contradicts the repository's own literal usage —
  `finish-contract.md` runs `git branch -d "openspec/<name>"` and `state-file.md`'s JSON schema
  stores `"branch": "openspec/<name>"`, both bare.
- **A shape rule** — "no extension and no trailing slash means not a citation" — has real collateral
  damage: it would silently stop checking `` `<agents repo>/rules/<name>` `` and
  `` `<project>/gradlew workspaceRemove` `` among six corpus hits, and would create a permanent blind
  spot for any future bare directory-shaped citation. That is the fail-open outcome task 9 exists to
  prevent.

- [x] **Step 2: `.superpowers/sdd/` is worktree-scoped — eighteen sites**

Every `<project>/.superpowers/sdd/…` becomes `<abs-worktree>/.superpowers/sdd/…`: seven in
`skills/myflow-do/SKILL.md`, five in `skills/myflow-contracts/pipeline.md`, two in
`principles-reviewer-prompt.md`, and one each in `adversarial-reviewer-prompt.md`,
`finish-contract.md`, `handoff-blocks-rationale.md` and `project-configuration.md`.

The evidence that `<project>/` is wrong here is the corpus's own:

- **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`) lists these artifacts as
  living in `.superpowers/sdd/` **in the worktree**, removed with it at run 2.
- `skills/myflow-do/SKILL.md` **already** cites `<abs-worktree>/.superpowers/sdd/dispatch-context.md`
  at three sites, so the file uses both roots for one directory.
- `skills/myflow-do/SKILL.md:429` reads "against `<project>/.superpowers/sdd/final-review.diff` **in
  the worktree**" — the sentence contradicts its own citation.
- `specs/myflow-citation-roots/spec.md`'s worked scenario for a recognised placeholder root is
  literally `` `<abs-worktree>/.superpowers/sdd/final-review.diff` `` — this change's own spec
  disagrees with this change's own corpus.

**The two reviewer-prompt files matter most.** They ship verbatim as subagent prompts, and
`principles-reviewer-prompt.md` states that the subagent's working directory is the project
worktree, so a `[DIFF_PATH]` handed to a fresh subagent must be worktree-rooted.

- [x] **Step 3: A fabricated `file:line` illustration is not a citation — one site**

`skills/myflow-do/SKILL.md:485`'s example findings row cites `<project>/src/Foo.kt:42`. Every other
`file:line` reference in that same file is bare, because a findings table's Location column is taken
verbatim from `git diff` output and is diff-relative. Rooting the illustration teaches a format real
findings never use.

Revert it to bare `src/Foo.kt:42`. If the guard then reports it, exclude the `file:line` shape — a
token ending `:<digits>` names a line in a file, not a path to cite — rather than re-rooting it.

- [x] **Step 4: Harness cases and mutation**

One case per exclusion added in steps 1 and 3, plus one asserting `<abs-worktree>/…` is **recognised**
rather than merely absent. Then disable each new exclusion in turn and confirm exactly its own case
fails, per the discipline every exclusion in this guard is held to.

- [x] **Step 5: Verify**

```bash unverified:the corpus count should stay at zero — every change here either re-roots a site to a different recognised placeholder or reverts one to bare behind a new exclusion; a nonzero count means a revert landed without its exclusion
scripts/check-installed-citations.sh; echo "exit=$?"
bash scripts/test-check-installed-citations.sh; echo "exit=$?"
```

**Tests:** `git-branch-name-is-not-a-citation`, `file-line-reference-is-not-a-citation` and
`abs-worktree-rooted-is-recognised`, in `scripts/test-check-installed-citations.sh`.

**Regression:** Reverting this task leaves an instruction telling an agent to create a branch named
`<project>/openspec/<name>`, two dispatched reviewer prompts pointing at a path that does not exist
in the directory the reviewer runs in, and this change's own spec contradicting this change's own
corpus on the worked example it chose to illustrate the rule.

**Baseline:** before=0 after=0 violations; 24 sites re-rooted or reverted across eight files.
<!-- measured: grep -rn over the worktree at c849b58 found 5 `<project>/openspec/<name>` sites, 18 `<project>/.superpowers/sdd` sites in installed files, and 1 `<project>/src/Foo.kt:42` -->

**Commit:** `fix(kan-102-citations-resolve-to-installed-paths): correct the roots wave B and C got syntactically right and semantically wrong`

---

### 8 Register the guard in `.myflow/project.md`

**Build:** green

**Files:**
- Modify: `.myflow/project.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: tasks 1-7. Every one of them must have landed: the guard must exist, and the corpus must
  be clean, or this entry fails the repository's own lint on every unrelated invocation.
- Produces: nothing. This is the last task.

- [x] **Step 1: Add the guard to `## lint`**

`scripts/check-installed-citations.sh` scans the repository tree and runs against a bare tree, so it
belongs in `## lint` — the criterion the existing requirement states, and the same one that keeps
`check-finish-preflight.sh` out.

- [x] **Step 2: Add the harness to `## test`**

`scripts/test-check-installed-citations.sh`, beside its siblings.

- [x] **Step 3: Note the guard's own runtime beside the section's timing note**

`## test` already records that its list runs close to this harness's default tool timeout and must be
split across invocations. This guard runs `setup.sh` twice per invocation, so say what that costs.

```bash unverified:measure it rather than copying the Baseline's single-mode figure — this guard runs both modes, and the Baseline measured only global
time scripts/check-installed-citations.sh >/dev/null
```

- [x] **Step 4: Run the guard half of the lint list, now including the new guard**

```bash unverified:the same eleven entries as task 6 step 4 with the new guard appended; it is the first run where the guard is part of the project's declared lint rather than invoked by hand
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-provenance.sh \
  && scripts/check-task-build-green.sh && scripts/check-workspace-isolation.sh \
  && scripts/check-uitest-overrides.sh && scripts/check-contract-budget.sh \
  && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh \
  && scripts/check-guard-symlinks.sh && scripts/check-self-review-report.sh \
  && scripts/check-installed-citations.sh; echo "exit=$?"
```

- [x] **Step 5: Run the guard harnesses**

```bash unverified:confirm both harness tail lines read "all cases pass" before claiming the task done
bash scripts/test-check-installed-citations.sh && bash scripts/test-check-references.sh; echo "exit=$?"
```

**Tests:** none added. The registration itself is asserted by
`openspec/specs/agents-repo-verification/spec.md`'s existing requirement that `## lint` names every
guard scanning the repository tree.

**Regression:** Reverting this task leaves the guard present and never run. Nothing would report a new
unrooted citation, and the 200 fixed by waves A, B and C would begin accumulating again from the next
edit to any installed file.

**Baseline:** before=14 after=15 entries in `.myflow/project.md`'s `## lint` block, and before=30
after=31 in its `## test` block. Eleven of the fourteen lint entries are guard scripts; the other
three are `stats/`'s Go and TypeScript checks, which no task here touches.
<!-- measured: awk over the ## lint and ## test fenced blocks in .myflow/project.md @ 332e593 gave 14 and 30 non-empty lines -->

**Commit:** `chore(kan-102-citations-resolve-to-installed-paths): run the citation-root guard in this repository's lint`

---

### 11 `myflow-fast` carries the sibling files it dispatches with

**Build:** green

**Files:**
- Add: `skills/myflow-fast/engineering-principles.md` *(relative symlink to `../myflow-do/engineering-principles.md`)*
- Add: `skills/myflow-fast/principles-reviewer-prompt.md` *(relative symlink to `../myflow-do/principles-reviewer-prompt.md`)*
- Add: `skills/myflow-fast/adversarial-reviewer-prompt.md` *(relative symlink to `../myflow-do/adversarial-reviewer-prompt.md`)*
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md` — carries the reasoning, so `SKILL.md` states the rule alone and stays under budget
- Modify: `scripts/check-installed-citations.sh` — one expected-zero declaration for the new symlink

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: wave C (task 6), which prefixed `skills/myflow-do/SKILL.md`'s other citations and left
  this paragraph untouched — it states a *resolution rule* rather than carrying a bare path, so no
  wave's mechanical rewrite could reach it.
- Produces: a `[PRINCIPLES_PATH]` rule stated against the running command's own skill directory.
  Nothing later depends on it.

**Why this belongs to KAN-102.** It is the same defect the change exists for, in the one shape
neither the new guard nor any wave can see. `[PRINCIPLES_PATH]` is defined as "the absolute path of
`engineering-principles.md` in the directory you are reading this file from — under a global install,
`~/.claude/skills/myflow-do/engineering-principles.md`". Both halves are unrooted. `/myflow-fast`
runs `skills/myflow-do/SKILL.md`'s section text verbatim, so "the file you are reading" names the
**cited** skill rather than the **running** one; and the hardcoded path names Claude Code's install
root inside a file that also installs into `~/.cursor/skills/` and `~/.codex/skills/`.

**The consequence was observed, not predicted.** A `/myflow-fast` run on another project resolved
`skills/myflow-fast/engineering-principles.md` for an implementer's REQUIRED READING and found
nothing there — the dispatch proceeded regardless, which is exactly the silent failure the delta
spec's new requirement names.

- [x] **Step 1: Link the three files myflow-fast dispatches with**

Relative symlinks into `../myflow-do/`, matching the convention `skills/*/scripts/` already uses for
guards. Only the three that current skill text actually names — `bug-hunter-reviewer-prompt.md` and
`security-reviewer-prompt.md` are named by no `SKILL.md` (both slots dispatch by `subagent_type`), so
linking them would propagate dead files.

```bash verified:run in this worktree; all three resolve, and `ls -la` shows each as a relative symlink into ../myflow-do/
cd skills/myflow-fast && for f in engineering-principles.md principles-reviewer-prompt.md \
  adversarial-reviewer-prompt.md; do ln -s "../myflow-do/$f" "$f"; done
```

- [x] **Step 2: Restate `[PRINCIPLES_PATH]` against the running command's own skill**

In `skills/myflow-do/SKILL.md`, replace "in the directory you are reading this file from — under a
global install, `~/.claude/skills/myflow-do/engineering-principles.md`" with a resolution against
**the running command's own skill directory**, citing **Guard resolution**
(`skills/myflow-contracts/pipeline.md`) rather than restating it — that section already states this
exact rule for guards, and its reason (a contract loaded by more than one command must resolve
inside whichever command is actually running) is the reason here too.

Keep the requirement that the resolved path is **absolute**, and that a missing file stops the run
rather than dispatching a blind reviewer. Neither changes.

- [x] **Step 3: Run the symlink and reference guards**

```bash verified:run after steps 1 and 2 in this worktree; GUARD-SYMLINKS-OK named 60 guards across 6 skills, check-references reported all sections resolve
scripts/check-guard-symlinks.sh && scripts/check-references.sh; echo "exit=$?"
```

Expected: exit `0` from both. `check-guard-symlinks.sh` rule 1 scans `skills/*/scripts/` only, so
these three top-level symlinks fall outside its scan — recorded here so a later reader does not
mistake its silence for coverage of them.

- [x] **Step 4: Run the citation, budget and markdown guards**

```bash verified:run in this worktree; all three exit 0 after the two corrections below
scripts/check-installed-citations.sh && scripts/check-contract-budget.sh \
  && scripts/check-markdown-integrity.py; echo "exit=$?"
```

**Both guards failed on the first attempt, and the failures are recorded here rather than smoothed
over — each one names a rule this task had to satisfy rather than a mistake to hide.**

**The budget row did go over**, exactly as this step anticipated: `skills/myflow-do/SKILL.md` reached
71376 bytes against its 71317 budget, 59 over, because the first draft of step 2 argued its case in
`SKILL.md` itself. The fix is the split this repository already prefers — `SKILL.md` states the rule
and cites `SKILL-rationale.md`; the rationale file carries why the *running* skill is the right root,
why absolute, and the observed failure. The row now measures 71131, with 186 bytes of headroom.

**The new symlink needed an expected-zero declaration.** `check-installed-citations.sh` reported
`skills/myflow-fast/engineering-principles.md:0: 0 checked, and not declared expected-zero
(coverage)` — its coverage rule requires a member citing nothing to say so, precisely so a file that
silently stops being scanned cannot pass as clean. It now carries the same declaration
`skills/myflow-do/engineering-principles.md` has, noting it is a symlink to that copy. The other two
symlinks needed nothing: they carry real citations and contribute coverage of their own
(`adversarial-reviewer-prompt.md` 1, `principles-reviewer-prompt.md` 11).

The guard's scanned-member count moves **56 → 59** as a result, which is the honest signal that these
three files are now part of the installed set the change governs.

**Tests:** none added. The convention is asserted by the delta spec's requirement **A sibling-file
citation resolves against the running command's own skill**. No existing harness scans a skill's
non-`scripts/` siblings, and adding one is a larger change than this task: the guard would have to
learn which sibling files a `SKILL.md` names *for a dispatch*, which is a prose-classification
problem rather than a path-shape one.

**Regression:** Reverting this task returns `/myflow-fast` to dispatching implementers and the
principles slot with a REQUIRED READING path that does not resolve inside its own skill directory.
The failure is silent — the dispatch still happens, and the subagent simply never reads the
principles it was told to satisfy.

**Baseline:** before=0 after=3 sibling `.md` files in `skills/myflow-fast/`; `skills/myflow-do/`
carries 5 and is unchanged. `check-installed-citations.sh` scans before=56 after=59 members.
`skills/myflow-do/SKILL.md` measures before=71376 after=71131 bytes against a 71317 budget — the
before figure is the over-budget first draft of step 2, not the pre-task file.
<!-- measured: ls skills/myflow-*/[a-z]*.md | grep -Ec "engineering-principles|reviewer-prompt" per directory, wc -c on SKILL.md, and the scanned-member count in check-installed-citations.sh's own OK line @ branch openspec/kan-102-citations-resolve-to-installed-paths -->

**Commit:** `fix(kan-102-citations-resolve-to-installed-paths): myflow-fast carries the sibling files it dispatches with`
