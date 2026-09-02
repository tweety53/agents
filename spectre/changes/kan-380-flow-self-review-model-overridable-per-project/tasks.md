# kan-380-flow-self-review-model-overridable-per-project

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Two tasks, both prose. Task 1 rewords the three `stats/` comment sites that say an empty
`selfReviewModel` inherits `defaultModel` — Go doc comment, CLI flag help string, migration
header — with no logic change, so `go vet`, `gofmt -l` and the existing tests prove nothing moved.
Task 2 is the skill layer: the resolver, the project key, the handshake, the `/flow-settings`
option and this repository's own key. `design.md` is canonical for the resolution order, the
handshake and the decisions; `proposal.md` for why. Every Go command below runs from `stats/`;
every guard from the repo root.

**Baseline, measured before any edit:**

- `stats/internal/store/settings_test.go`: 6 tests.
  <!-- measured: grep -c '^func Test' stats/internal/store/settings_test.go @ branch main (472c8a2) -->
- `stats/cmd/flow/settings_test.go`: 5 tests.
  <!-- measured: grep -c '^func Test' stats/cmd/flow/settings_test.go @ branch main (472c8a2) -->
- `skills/flow-settings/SKILL.md` is 7457 bytes against a `check-contract-budget.sh` row of 8010;
  every other file task 2 touches has more than 2500 bytes of headroom. Task 2's edits to that
  file are rewordings plus one dropped paragraph, so the row is expected to hold; if the guard
  trips, raise that row in the same commit — a budget raise is a visible diff by the guard's own
  design.
  <!-- measured: wc -c skills/flow-settings/SKILL.md against its row in scripts/check-contract-budget.sh @ branch main (472c8a2) -->
- `jq -r '.selfReviewModel // empty'` prints the empty string for both `""` and an absent key, and
  the value itself otherwise — so the resolver's `[ -z … ]` test needs no `null` case.
  <!-- measured: printf '{"selfReviewModel":""}' | jq -r '.selfReviewModel // empty' | wc -c @ branch main (472c8a2) -->

- [x] 1. Reword the store-side "empty inherits" comments to "empty means fable"

No logic changes. `ValidateSettings` already accepts `""` and every `ValidModels` member,
`fable` included, so the store needs nothing new to hold the meaning task 2 gives an empty value.

**`stats/internal/store/settings.go`** — the `Settings` struct's doc comments only:
- `SelfReviewModel` (the comment that reads "Unlike DefaultModel, empty is a valid value here --
  it means "inherit DefaultModel", not "unset""): empty is valid and means the literal `fable`,
  the same meaning `PlanningModel`'s empty carries; resolution — project key, then this field,
  then `fable` — is `skills/flow/SKILL.md`'s Model resolution, not the store's.
- `PlanningModel` (the comment that reads "unlike SelfReviewModel -- it does not mean "inherit
  DefaultModel""): drop the contrast; both fields' empty now means `fable`.

**`stats/cmd/flow/settings.go`** — the `-self-review-model` flag's help string
(`fset.String("self-review-model", "", "self-review's model, e.g. opus; empty inherits -model")`):
change the trailing clause to `empty resolves to the store's default`, matching the
`-planning-model` flag's own wording beside it.

**`stats/internal/store/migrations/0016_flow_settings_self_review_model.sql`** — header comment
only, the SQL statement untouched: "inherit default_model" and "resolves to "inherit"" become the
literal `fable` the skill resolver falls back to, stated once; the NOT-NULL-DEFAULT-empty reason
stays. Migrations are recorded by filename alone (`stats/internal/store/migrations.go` keeps a
`filename` primary key and no checksum), so editing an applied migration's comment changes
nothing for a database that already ran it.

  - [ ] **Step 1: Edit the three comment sites above.**
  - [ ] **Step 2: Prove nothing moved.**

```bash verified:the commands stats/README.md and .flow/project.md name for this module
cd stats && gofmt -l . && go vet ./... \
  && go test ./internal/store/... ./cmd/flow/... -race -count=1
grep -rn -i "inherit" stats/internal/store/settings.go stats/cmd/flow/settings.go \
  stats/internal/store/migrations/0016_flow_settings_self_review_model.sql ; test $? -eq 1
```

`gofmt -l .` prints nothing; both test packages pass with unchanged counts; the `grep` finds no
`inherit` left in the three files.

  - [ ] **Step 3: Commit.**

**Files:** `stats/internal/store/settings.go`, `stats/cmd/flow/settings.go`,
`stats/internal/store/migrations/0016_flow_settings_self_review_model.sql`
**Tests:** none
**Regression:** reverting this commit leaves the store's own comments claiming an empty
`selfReviewModel` inherits `defaultModel` while task 2's resolver reads it as `fable` — two
sources of truth for one field's empty value.
**Baseline:** before=6 after=6 store tests, before=5 after=5 cmd/flow settings tests
<!-- predicted: grep -c '^func Test' per file after task 1 -->
**Commit:** `docs(store): an empty selfReviewModel means fable, not inherit`
**Build:** green

- [x] 2. `SELF_REVIEW_MODEL`: three-tier resolution, the project key, the handshake, the `/flow-settings` option

Prose only — no Go, no tests. `design.md`'s **Resolution**, **The project key**, **The
handshake** and **What "empty" means** are canonical; this task writes them into the places a
run reads.

**`skills/flow/SKILL.md`** Model resolution:
- In the bash block, replace the two `SELF_REVIEW_MODEL` lines with:

```bash verified:jq behaviour measured in the baseline above; shape copied from the PLANNING_MODEL lines in the same block
SELF_REVIEW_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.selfReviewModel // empty')"
# <project>/.flow/project.md's `## self review model` body, when present and a ValidModels member, wins
[ -z "$SELF_REVIEW_MODEL" ] && SELF_REVIEW_MODEL=fable
```

- Rewrite the `**\`SELF_REVIEW_MODEL\` resolves independently of \`DEFAULT_MODEL\`**` paragraph
  in `PLANNING_MODEL`'s paragraph's shape: it governs the archive-phase self-review subagent
  (**Run self-review**, `skills/flow/archive.md`); `<project>/.flow/project.md`'s `## self review
  model` key, when present and a valid `ValidModels` member, wins over the store's
  `selfReviewModel`; when both are empty, or the store is unreachable, it falls back to the
  literal `fable`, naming this a fallback exactly as `DEFAULT_MODEL`'s `sonnet` literal is; a
  plain-language session instruction overrides it for that run only, recorded with the dispatch
  and never written back. Delete the sentence that says inheriting `DEFAULT_MODEL` is the field's
  "unset" meaning — it is no longer true, and a false sentence is not a restatement to keep.
- The unreachable-store paragraph above the table already names `DEFAULT_MODEL`'s `sonnet`
  fallback; add one clause naming that `SELF_REVIEW_MODEL` falls back to `fable` on the same
  failure, exactly as `PLANNING_MODEL`'s paragraph does.

**`skills/flow/archive.md`** **Run self-review**, the paragraph beginning "On anything but No,
the combined reasoning pass runs as a subagent, on `SELF_REVIEW_MODEL`": add the handshake —
the subagent's report opens with `Model: <the model named in its own system prompt>`; the session
compares that line with `SELF_REVIEW_MODEL`; a mismatch re-dispatches once on `opus`; a second
mismatch proceeds on whatever model answered and names it in the run's own output; the `Model:`
line is stripped before the five-angle body is used. State plainly that no dispatch record is
written for either dispatch — self-review records none today and this change adds none — and
that the handshake never blocks: a self-review that runs on the wrong model still runs.

**`skills/flow-contracts/finish-contract-run2.md`** **Run self-review**: the parenthetical
"(empty inherits `defaultModel` — **Model resolution**, `skills/flow/SKILL.md`)" becomes a
pointer to the same section with the three-tier order named in one clause — project key, then
the store's `selfReviewModel`, then `fable`, with `opus` on a verified mismatch — so run 2's
contract and the skill agree.

**`skills/flow-contracts/project-configuration.md`**:
- Add a `## self review model` row to the optional keys table directly under `## planning
  model`'s, same single-line-literal shape: one member of
  `<agents repo>/stats/internal/store/settings.go`'s `ValidModels` map and nothing else, never
  free-form prose; when present and valid it wins over the store's `selfReviewModel` in
  `SELF_REVIEW_MODEL` resolution (**Model resolution**, `skills/flow/SKILL.md`); absent falls
  through to the store; a body matching no member is reported by name (quoting what was found)
  and dropped, resolving as if the key were absent.
- Under the `## planning model` matching paragraph, add one sentence: `## self review model`'s
  body is matched exactly the same way.

**`skills/flow-settings/SKILL.md`**:
- **Read current settings**: `selfReviewModel` "(a string, empty meaning "inherit
  `defaultModel`")" becomes "(a string, empty meaning "the store's own default, `fable`")", and
  the printed line `self-review model:  <selfReviewModel, or "(inherits default model)" when
  empty>` becomes `<selfReviewModel, or "(fable — store default)" when empty>`, matching the
  planning-model line beneath it.
- **Offer to change each field**, the **Self-review model** bullet: the explicit option
  **"Inherit default model"** becomes **"Store default (fable)"** (still mapping to the empty
  string, `-self-review-model ""`); the sentence contrasting it with `defaultModel` stays — an
  empty value is still a first-class choice — and it now names what empty resolves to: `fable`,
  unless `<project>/.flow/project.md`'s `## self review model` key or a per-run session
  instruction overrides it, per **Model resolution** (`skills/flow/SKILL.md`).
- The **Planning model** bullet's clause "unlike `selfReviewModel`, an empty value does not mean
  "inherit `defaultModel`" — it resolves to the literal `fable`" loses its contrast: the two
  fields now agree, so the clause reads "an empty value resolves to the literal `fable`".
- The intro paragraph's description of `selfReviewModel` gains the same override clause the
  `planningModel` description already carries.

**`.flow/project.md`** (this repository): add, directly after the `## default landing route`
section and before `## workspace isolation`:

```markdown verified:shape copied from this file's own `## default landing route` and `## jira` sections
## self review model

`fable`
```

**`scripts/check-contract-budget.sh`**: only if `check-contract-budget.sh` trips on
`skills/flow-settings/SKILL.md` (see the baseline) — raise that row to the file's new size plus
its headroom, in this same commit.

  - [ ] **Step 1: Edit the six prose files above.**
  - [ ] **Step 2: Run the prose guards.**

```bash verified:the guards .flow/project.md's lint list names for skill prose
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh \
  && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py \
  && scripts/check-stage-mark-calls.sh && scripts/check-dispatch-paragraphs.sh
grep -rn -i "inherit" skills/flow/SKILL.md skills/flow-settings/SKILL.md \
  skills/flow-contracts/finish-contract-run2.md ; test $? -eq 1
```

Every guard exits 0 and the `grep` finds no `inherit` left in the three reworded skill files.

  - [ ] **Step 3: Commit.**

**Files:** `skills/flow/SKILL.md`, `skills/flow/archive.md`,
`skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/project-configuration.md`,
`skills/flow-settings/SKILL.md`, `.flow/project.md`, `scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** reverting this commit puts `SELF_REVIEW_MODEL` back on one store field
inheriting `DEFAULT_MODEL`, with no project key, no `fable` default and no verified `opus`
fallback; `## self review model` in `.flow/project.md` becomes a key no skill reads.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `docs(flow): resolve the self-review model per project with a verified opus fallback`
**Build:** green
