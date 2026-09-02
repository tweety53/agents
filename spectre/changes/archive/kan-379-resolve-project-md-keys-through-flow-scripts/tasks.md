# kan-379-resolve-project-md-keys-through-flow-scripts

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** yes — tasks 3 and 5 move existing prose verbatim between
> `skills/flow-contracts/` files; no normative sentence, exit-code contract, scenario or
> rejected-alternative rationale is reworded.

## Global constraints

- `~/.claude/rules/be-brief.md` — cut, never paraphrase. A "**Load `…`**" sentence is deleted, not
  reworded; a moved passage is byte-identical at its destination.
- Every command under `## lint` in `.flow/project.md` exits clean after every task;
  `scripts/run-guard-tests.sh` and `cd stats && go test ./... -race -count=1` pass after every
  task that touches `scripts/` or `stats/`.
- `scripts/check-guard-symlinks.sh` rule 4: no shipped script derives a repository root from
  `$SCRIPT_DIR/..`; `project-get.sh` takes its root as an argument.
- Baselines: 49 Bash harnesses under `scripts/test-*.sh`; 128 test functions in
  `stats/cmd/flow`.
  <!-- measured: ls scripts/test-*.sh | wc -l @ ed2fbf8 -->
  <!-- measured: cd stats && go test ./cmd/flow -count=1 -list '.*' | grep -c '^Test' @ ed2fbf8 -->

- [x] 1. Shared section extractor and `project-get.sh`

**Build:** green

**Files:**
- Create: `scripts/lib/project-section.sh`
- Create: `scripts/project-get.sh`
- Create: `scripts/test-project-get.sh`
- Create: `skills/flow/scripts/project-get.sh` (symlink)
- Modify: `scripts/gather-dispatch-context.sh`
- Modify: `scripts/check-model-keys.sh`

**Tests:** `scripts/test-project-get.sh` — Case 1: usage with one argument exits 2; Case 2: root not
a directory exits 2; Case 3: no `.flow/project.md` exits 1 with a stderr line naming the path;
Case 4: key absent exits 1 naming the key; Case 5: key present, body printed verbatim including a
fenced block and a `### ` subheading, trailing blank lines removed, exit 0; Case 6: key declared
twice exits 2; Case 7: a UTF-8 BOM before the first heading does not hide a `## <key>` on line 1;
Case 8: a key whose body is empty prints nothing and exits 0; Case 9: a multi-word key passed as
one quoted argument resolves
**Regression:** Case 5 fails if `project_section` stops at a `### ` line or keeps trailing blanks;
Case 6 fails if the twice-declared check is dropped; Case 7 fails if `strip_bom_cat` is bypassed;
Case 3/4 fail if the two exit-1 reasons collapse into one
**Baseline:** before=49 after=50
<!-- measured: ls scripts/test-*.sh | wc -l @ ed2fbf8; the after count is this task's one new harness -->
**Commit:** `feat(scripts): add project-get.sh over a shared .flow/project.md section extractor`

  - [ ] **Step 1: Write `scripts/lib/project-section.sh`**

  A sourced library in the shape of `scripts/lib/within-root.sh`: a header naming its three
  callers and why the extractor is defined once, then one function. It sources
  `scripts/lib/strip-bom.sh` relative to its own `${BASH_SOURCE[0]}` directory.

  ```bash unverified:confirm strip_bom_cat's calling convention against scripts/lib/strip-bom.sh before relying on it
  # project_section <file> <key> -> prints the body of the first "## <key>"
  # heading: every line after it up to, not including, the next "^## " line
  # or EOF ("### " subheadings included), leading and trailing blank lines
  # removed, nothing else normalised. Prints nothing when <key> is absent.
  project_section() {
    local file="$1" key="$2"
    strip_bom_cat "$file" | awk -v key="## $key" '
      $0 == key { grab = 1; next }
      /^## / { if (grab) exit }
      grab { lines[++n] = $0 }
      END {
        s = 1; e = n
        while (s <= e && lines[s] ~ /^[[:space:]]*$/) s++
        while (e >= s && lines[e] ~ /^[[:space:]]*$/) e--
        for (i = s; i <= e; i++) print lines[i]
      }'
  }
  ```

  - [ ] **Step 2: Write `scripts/project-get.sh`**

  Header in the shape of `scripts/check-visual-trigger.sh`'s: usage, the three exit codes, and the
  twice-declared refusal. Body:

  ```bash unverified:run scripts/test-project-get.sh once written
  #!/usr/bin/env bash
  set -euo pipefail
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  source "$SCRIPT_DIR/lib/project-section.sh"
  [ "$#" -eq 2 ] || { echo "usage: project-get.sh <project-root> <key>" >&2; exit 2; }
  ROOT="$1"; KEY="$2"
  [ -d "$ROOT" ] || { echo "project-get: $ROOT is not a directory" >&2; exit 2; }
  CFG="$ROOT/.flow/project.md"
  [ -f "$CFG" ] || { echo "project-get: $CFG does not exist" >&2; exit 1; }
  COUNT="$(strip_bom_cat "$CFG" | awk -v key="## $KEY" '$0 == key { n++ } END { print n + 0 }')"
  case "$COUNT" in
    0) echo "project-get: $CFG declares no '## $KEY' section" >&2; exit 1 ;;
    1) project_section "$CFG" "$KEY" ;;
    *) echo "project-get: $CFG declares $COUNT '## $KEY' sections — a second declaration is ambiguous, so neither was read" >&2; exit 2 ;;
  esac
  ```

  `chmod +x scripts/project-get.sh`.

  - [ ] **Step 3: Symlink it into the skill**

  Run: `ln -s ../../../scripts/project-get.sh skills/flow/scripts/project-get.sh`
  Expected: `readlink skills/flow/scripts/project-get.sh` prints the relative target, matching
  every other entry in that directory (`ls -l skills/flow/scripts | head -3` shows the shape).

  - [ ] **Step 4: Write `scripts/test-project-get.sh`**

  Same harness shape as `scripts/test-check-visual-trigger.sh` (fixture roots under `TMPDIR`, the
  real script as a subprocess, `pass`/`fail` counters, `# Case N:` comments) covering the nine
  cases in this task's **Tests:** field. Case 5's fixture body is a `## lint` section holding a
  fenced block, a `### note` subheading and two trailing blank lines; the expected output is
  asserted byte for byte against a here-doc.

  Run: `scripts/test-project-get.sh`
  Expected: nine `ok:` lines, exit 0.

  - [ ] **Step 5: Point `gather-dispatch-context.sh` at the library**

  Delete the `extract_project_section()` function (the comment block above it stays, its
  `extract_project_section <file> <key>` description line replaced by one line naming
  `project_section` from `lib/project-section.sh`). Add `source "$SCRIPT_DIR/lib/project-section.sh"`
  beside the four existing `source` lines. Replace the one call
  `extract_project_section "$PROJECT_FILE_RESOLVED" "$key"` with
  `project_section "$PROJECT_FILE_RESOLVED" "$key"`.

  Run: `scripts/test-gather-dispatch-context.sh`
  Expected: exit 0 — the `## project commands` cases still see the lint/test/run bodies.

  - [ ] **Step 6: Point `check-model-keys.sh` at the library**

  Replace the first `awk` of `extract_section_body` (the heading-to-next-heading grab) with a
  call to `project_section "$file" "$heading"`, keeping the second `awk` (the blank-trim and
  backtick-strip) as the pipe it feeds. Add `source "$SCRIPT_DIR/lib/project-section.sh"` after
  `SCRIPT_DIR` is set.

  Run: `scripts/test-check-model-keys.sh && scripts/check-model-keys.sh`
  Expected: both exit 0.

  - [ ] **Step 7: Guards and commit**

  Run: `scripts/check-guard-symlinks.sh && scripts/run-guard-tests.sh`
  Expected: exit 0; the runner reports 50 harnesses.
  <!-- predicted: scripts/run-guard-tests.sh after this task; 49 at ed2fbf8 plus test-project-get.sh -->

  ```bash verified:the commit subject is this task's Commit field
  git add scripts/lib/project-section.sh scripts/project-get.sh scripts/test-project-get.sh skills/flow/scripts/project-get.sh scripts/gather-dispatch-context.sh scripts/check-model-keys.sh
  git commit -m "feat(scripts): add project-get.sh over a shared .flow/project.md section extractor"
  ```

- [x] 2. Repoint the phase files

**Build:** green

**Files:**
- Modify: `skills/flow/review-panel.md`
- Modify: `skills/flow/verify-and-handoff.md`
- Modify: `skills/flow/integrate.md`
- Modify: `skills/flow/archive.md`
- Modify: `skills/flow/implement.md`

**Tests:** **none** — prose edits; `check-references.sh`, `check-guard-symlinks.sh` and
`check-contract-budget.sh` are the gates
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): resolve .flow/project.md keys through project-get.sh`

  - [ ] **Step 1: `review-panel.md` — the citation-check key (line 21)**

  Delete the sentence `**Load `skills/flow-contracts/project-configuration.md`** — the key below
  is canonical there.` In the paragraph after it, replace `read `<project>/.flow/project.md`'s
  `## review panel citation check` key` with `run `project-get.sh <worktree> "review panel citation
  check"`` and, after `If declared, run its command`, state the exit mapping in one sentence: exit 1
  is the "key absent — skip silently" case the paragraph already names; exit 2 is reported and the
  worktree skipped, the same way the guard's own exit 2 is. The **Guard resolution** and **Project
  configuration** cites stay.

  - [ ] **Step 2: `review-panel.md` — the standards entries (line 288)**

  Replace `from the `## standards` entries in `<project>/.flow/project.md`, per **Project
  configuration** (`skills/flow-contracts/project-configuration.md`)` with `from the entries
  `project-get.sh <worktree> standards` prints (exit 1: none declared), resolved per the entry-form
  and containment rule the `[STANDARDS_PATHS]` step of `skills/flow/principles-reviewer-prompt.md`
  carries`.

  - [ ] **Step 3: `verify-and-handoff.md` — lint/test (line 66) and the cache-index load (line 48)**

  Delete `**Load `skills/flow-contracts/project-configuration.md`** — the `## lint`/`## test` keys
  below are canonical there.` Replace `Run the `## lint` and `## test` commands from
  `<project>/.flow/project.md` (auto-detect if absent)` with `Run the commands `project-get.sh
  <worktree> lint` and `project-get.sh <worktree> test` print (auto-detect on exit 1)`.

  Replace `**Load `skills/flow-contracts/workspace-isolation.md`** — it is canonical for the cache
  index row below.` with `**Load `skills/flow-contracts/workspace-isolation.md` only when
  `prepare-workspace.sh` exited non-zero, or exited 0 with stderr naming a `cache index` row** — the
  procedures for both live there; the ordinary exit-0 run loads nothing.` The line-57 fallback
  paragraph is untouched.

  - [ ] **Step 4: `integrate.md` — the landing route (line 133)**

  Delete `**Load `skills/flow-contracts/project-configuration.md`** — the `## default landing
  route` key is canonical there.` Replace `Read `<project>/.flow/project.md`'s `## default landing
  route` section, if present,` with `Run `project-get.sh <main-checkout> "default landing route"`
  (exit 1: absent)`. The byte-for-byte literal-match sentences stay.

  - [ ] **Step 5: `archive.md` — the `remove` command (line 80)**

  Delete `**Load `skills/flow-contracts/project-configuration.md`** — the `remove` command below is
  canonical there.` In step 5, replace `runs the project's `remove` command from **Project
  configuration** (`skills/flow-contracts/project-configuration.md`), run from the **main
  checkout**, with the workspace id re-derived here from the change name — never handed to this
  run` with `runs the `remove` row of the command table `project-get.sh <main-checkout> "workspace
  isolation"` prints (exit 1: the skipped-not-failed case below), from the **main checkout**, with
  `<id>` from `flow workspace-id <name>` — never handed to this run`.

  - [ ] **Step 6: `implement.md` — the workspace id (lines 99-104)**

  Replace the two paragraphs from `**Load `skills/flow-contracts/workspace-isolation.md`**` through
  `on a fix run exactly as on the first.` with one: `**Then run `flow workspace-id <name>` for this
  worktree's workspace id**, once per run, on a fix run exactly as on the first.`

  - [ ] **Step 7: Guards**

  Run: `scripts/check-references.sh && scripts/check-guard-symlinks.sh && scripts/check-contract-budget.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-vocabulary.sh`
  Expected: all exit 0. Then `grep -n "Load .skills/flow-contracts/project-configuration.md" skills/flow/*.md`
  prints nothing, and `grep -n "Load .skills/flow-contracts/workspace-isolation.md" skills/flow/*.md`
  prints only `verify-and-handoff.md`'s conditional sentence.

  - [ ] **Step 8: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow/review-panel.md skills/flow/verify-and-handoff.md skills/flow/integrate.md skills/flow/archive.md skills/flow/implement.md
  git commit -m "docs(flow): resolve .flow/project.md keys through project-get.sh"
  ```

- [x] 3. Split `plan-provenance.md` by audience

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/plan-provenance.md`
- Create: `skills/flow-contracts/plan-provenance-guard.md`
- Modify: `skills/flow-contracts/build-green.md`
- Modify: `skills/flow-contracts/SKILL.md`
- Modify: `rules/flow-manual-review.mdc`
- Modify: `scripts/check-plan-provenance.py`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `scripts/check-installed-citations.sh`
- Modify: `scripts/check-references.sh`
- Modify: `scripts/test-check-plan-provenance.sh`

**Tests:** `test-check-plan-provenance.sh` — no new test added; its 3 existing assertions were
repointed to the guard's relocated cite as necessary collateral of the move; `check-references.sh`
and `check-contract-budget.sh` are the other gates
**Regression:** `test-check-plan-provenance.sh` fails if its 3 repointed assertions still expect the
pre-move cite
**Baseline:** before=passing after=passing (assertions repointed, not added)
**Commit:** `docs(flow-contracts): split plan-provenance.md by audience`

  - [ ] **Step 1: Create `plan-provenance-guard.md`**

  Title `# Plan provenance — the guard`, one opening paragraph stating the file is canonical for
  what `scripts/check-plan-provenance.py` enforces and that the tag vocabulary itself is **The
  four tags** (`skills/flow-contracts/plan-provenance.md`). Then move — cut from
  `plan-provenance.md`, paste byte-identical — the sections **The guard's scope, and why it is
  narrow**, **The quotation exemption** with all four `###` subsections, and **What the guard does
  not do**, in that order.

  Run: `wc -c skills/flow-contracts/plan-provenance.md skills/flow-contracts/plan-provenance-guard.md`
  Expected: about 6000 and about 19500 bytes.
  <!-- predicted: the per-section byte sums measured at ed2fbf8 (5822 run-applied, 19195 guard-facing) plus the new file's title paragraph -->

  - [ ] **Step 2: Repoint the citers**

  - `skills/flow-contracts/build-green.md` lines 69-73: the two `**What the guard does not do**`
    cites change their path to `skills/flow-contracts/plan-provenance-guard.md`.
  - `skills/flow-contracts/SKILL.md` line 31: the `plan-provenance.md` row's description becomes
    `Write a plan's provenance tags: the four tags, the asymmetry rule, the implementer's duty`;
    add a row beneath it for `plan-provenance-guard.md`: `What check-plan-provenance.py enforces:
    the guard's scope, the quotation exemption and its vetoes, what the guard does not do`.
  - `rules/flow-manual-review.mdc` line 57: add a table row for `plan-provenance-guard.md` with the
    purpose `what the plan-provenance guard enforces`. The row sits outside the `core` markers, so
    the managed block does not change.
  - `scripts/check-plan-provenance.py` lines 591-592, 904, 2634: each cite of
    `plan-provenance.md` that names the quotation exemption, a veto, or "What the guard does not do"
    now names `plan-provenance-guard.md`; line 73's "canonical definition" cite stays on
    `plan-provenance.md` (it names the tag rule).

  - [ ] **Step 3: Budget rows**

  In `scripts/check-contract-budget.sh`'s `budgets()` table set the
  `skills/flow-contracts/plan-provenance.md` row to the file's landed size plus 25 percent, and add
  a `skills/flow-contracts/plan-provenance-guard.md` row at its landed size plus 25 percent, both
  rounded up to the integer, in the table's sorted position.

  - [ ] **Step 4: Guards and commit**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-installed-citations.sh && scripts/check-installed-rules.sh && scripts/test-check-plan-provenance.sh && scripts/check-plan-provenance.sh`
  Expected: all exit 0.

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow-contracts/plan-provenance.md skills/flow-contracts/plan-provenance-guard.md skills/flow-contracts/build-green.md skills/flow-contracts/SKILL.md rules/flow-manual-review.mdc scripts/check-plan-provenance.py scripts/check-contract-budget.sh
  git commit -m "docs(flow-contracts): split plan-provenance.md by audience"
  ```

- [x] 4. `flow state resolve`

**Build:** green

**Files:**
- Modify: `stats/cmd/flow/state.go`
- Modify: `stats/cmd/flow/state_test.go`
- Modify: `stats/cmd/flow/main.go`

**Tests:** `TestStateResolveStoreDropsFinished`, `TestStateResolveFallbackUnionsChangesDir`,
`TestStateResolveFallbackDropsArchived`, `TestStateResolveFallbackNamesUnreadable`,
`TestStateResolveTakesNoPositionalArguments`
**Regression:** `TestStateResolveStoreDropsFinished` fails if a `FINISHED` row reaches
`candidates`; `TestStateResolveFallbackUnionsChangesDir` fails if the changes-directory scan is
dropped or `archive` is listed as a change; `TestStateResolveFallbackDropsArchived` fails if the
`archive/<name>` exclusion is dropped; `TestStateResolveFallbackNamesUnreadable` fails if an
unparseable fallback file is silently dropped; `TestStateResolveTakesNoPositionalArguments` fails
if the flag parser accepts one
**Baseline:** before=128 after=133
<!-- measured: cd stats && go test ./cmd/flow -count=1 -list '.*' | grep -c '^Test' @ ed2fbf8; the after count adds this task's five -->
**Commit:** `feat(flow): add state resolve for the change-name candidate set`

  - [ ] **Step 1: Write the failing tests**

  Beside `TestStateListFallbackReportsLocalRecords` in `state_test.go`, using the same
  `httptest` server and fallback-directory fixtures it uses. The fallback cases set `-C` to a temp
  main checkout carrying `spectre/changes/<a>/`, `spectre/changes/<b>/`,
  `spectre/changes/archive/<b>/` and `spectre/changes/archive/` itself; the expected `candidates`
  carry `<a>` and every readable fallback record's name, never `<b>` and never `archive`.

  Run: `cd stats && go test ./cmd/flow -run TestStateResolve -count=1`
  Expected: compile failure — `runStateResolve` undefined.

  - [ ] **Step 2: Implement `runStateResolve`**

  In `state.go`: a `case "resolve":` in `runState`'s switch; a `stateResolveOutput` struct
  (`Source string`, `Complete bool`, `Candidates []stateListRecord`, `Unreadable []string`, JSON
  tags `source`, `complete`, `candidates`, `unreadable`); `runStateResolve` reusing
  `parseStateListFlags` (the usage text names `state resolve` too), `fallback.ProjectKey` for the
  key and the main checkout, `listStateBoard` for the store path, and
  `fallbackStateListRecords` for the fallback path.

  ```go unverified:compile and run this task's tests once written
  func resolveCandidates(rows []stateListRecord, changesDir string, fromStore bool) (cands []stateListRecord, unreadable []string, err error) {
  	seen := map[string]bool{}
  	for _, r := range rows {
  		switch {
  		case r.Unreadable:
  			unreadable = append(unreadable, r.Name)
  		case fromStore && r.State == "FINISHED":
  		case !seen[r.Name]:
  			seen[r.Name] = true
  			cands = append(cands, r)
  		}
  	}
  	if fromStore {
  		return cands, unreadable, nil
  	}
  	entries, err := os.ReadDir(changesDir)
  	for _, e := range entries {
  		if e.IsDir() && e.Name() != "archive" && !seen[e.Name()] {
  			seen[e.Name()] = true
  			cands = append(cands, stateListRecord{Name: e.Name()})
  		}
  	}
  	cands = slices.DeleteFunc(cands, func(r stateListRecord) bool {
  		_, statErr := os.Stat(filepath.Join(changesDir, "archive", r.Name))
  		return statErr == nil
  	})
  	return cands, unreadable, err
  }
  ```

  Add `path/filepath` to `state.go`'s imports. A `ReadDir` error is printed on stderr as `flow: state resolve: read <dir>: <err>` and the
  output still prints, exit 0 — the never-block rule `state list` already follows. `candidates`
  and `unreadable` are always present in the JSON (empty arrays, never `null`).

  - [ ] **Step 3: Usage lines**

  `main.go`'s `usage` gains `  state resolve       print the change-name candidate set: source, complete, candidates, unreadable`
  under `state list`; `stateUsage` in `state.go` gains the `flow state resolve [-addr url] [-timeout dur] [-C dir]`
  line and one sentence per source stating what `candidates` holds.

  - [ ] **Step 4: Run the Go gates and commit**

  Run: `cd stats && gofmt -w . && go vet ./... && gofmt -l . && go test ./... -race -count=1`
  Expected: `gofmt -l` prints nothing; every package passes.

  ```bash verified:the commit subject is this task's Commit field
  git add stats/cmd/flow/state.go stats/cmd/flow/state_test.go stats/cmd/flow/main.go
  git commit -m "feat(flow): add state resolve for the change-name candidate set"
  ```

- [x] 5. `pipeline.md` stage marks and change-name resolution

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/pipeline.md`
- Modify: `skills/flow-contracts/pipeline-rationale.md`
- Modify: `skills/flow-status/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`

**Tests:** **none** — prose relocation; `check-references.sh`, `check-stage-mark-calls.sh`,
`check-contract-budget.sh` and the normative-inventory comparison in step 5 are the gates
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): resolve the change name through flow state resolve`

  - [ ] **Step 1: Rewrite **Change name resolution** in `pipeline.md` (line 443 onward)**

  Keep, verbatim: the opening paragraph's first sentence (`<name>` is optional on `/flow` and
  `/flow-status`); the "unreadable is reported and skipped — never silently dropped" paragraph; the
  "every command reports which source produced it" paragraph's first sentence; the three-outcome
  list; the "defined once, here" paragraph; the Jira naming line. Replace the store-first, the
  `"source":"store"`, the `"source":"fallback"` and the "From that union" paragraphs with:

  ```markdown verified:authored in-tree for this change; the JSON fields are task 4's
  **The candidate set is `flow state resolve -C <main-checkout>`'s `candidates` — never a
  hand-written HTTP call, never a directory listing of your own.** It prints one JSON object:
  `"source"` (`"store"` or `"fallback"`), `"complete"` (`true` only for `"source":"store"`),
  `"candidates"` (each carrying `name`, `state`, `updatedAt`, `updatedBy`) and `"unreadable"` (the
  names of fallback records that could not be read). A `FINISHED` change is never a candidate, and
  neither is one already under `<project>/spectre/changes/archive/`.
  ```

  Move each deleted paragraph's explanatory sentences — why the CLI rather than `curl`, why the
  filesystem source exists — verbatim under `pipeline-rationale.md`'s existing **Change name
  resolution (all `/flow*` commands)** heading, after its current text.

  - [ ] **Step 2: Trim **Stage marks** in `pipeline.md` (lines 160-235) to the caller's obligations**

  Keep, verbatim: the first paragraph's first two sentences (marks at the boundaries, key not
  prose name); the sentence `**A `stage begin` mark MUST carry `-session-token` and `-harness`.**`
  with its two definitions (a literal per-run token reused at every mark; `-harness` names the
  harness running the mark); the placeholder paragraph's first sentence (neither is ever
  hardcoded); the fenced example; the first sentence of the literal-token paragraph; the
  never-block paragraph's first and last sentences; the `stage end` paragraph; the `/flow-status`
  paragraph. Move every other sentence — what the CLI rejects, what the README-parsing test
  guarantees, what `check-stage-mark-calls.sh` rejects, the transcript-binding explanation, the
  "a reader will improve the literal" warning — verbatim under `pipeline-rationale.md` **Stage
  marks**, after its current text.

  Run: `wc -c skills/flow-contracts/pipeline.md skills/flow-contracts/pipeline-rationale.md`
  Expected: `pipeline.md` about 22000 bytes; `pipeline-rationale.md` about 19000 bytes.
  <!-- predicted: 29059 minus about 3000 (stage marks) minus about 2800 (resolution) for pipeline.md; 12063 plus the same moved text for the rationale, at ed2fbf8 sizes -->

  - [ ] **Step 3: Repoint `skills/flow-status/SKILL.md` (lines 32-49)**

  `flow state list [-C dir]` becomes `flow state resolve [-C dir]` in the sentence and in the
  `BOARD=` line; the `$BOARD` description names `candidates` and `unreadable` in place of
  `records` and the per-record `unreadable` flag; step 1's table is built from `candidates`, and
  every `unreadable` name is printed by name above it. Nothing else in the file changes.

  - [ ] **Step 4: Raise the `pipeline-rationale.md` budget row**

  Set `skills/flow-contracts/pipeline-rationale.md` in `budgets()` to its landed size plus 25
  percent, rounded up. `pipeline.md`'s row is not lowered.

  - [ ] **Step 5: Guards, normative inventory, commit**

  Run: `scripts/check-references.sh && scripts/check-stage-mark-calls.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh`
  Expected: all exit 0.

  Run `scripts/check-normative-inventory.sh` here and in a `git worktree add <tmp> ed2fbf8`
  checkout; every line of the base output for `pipeline.md` and `pipeline-rationale.md` appears
  in the branch output under one of the two files.

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow-contracts/pipeline.md skills/flow-contracts/pipeline-rationale.md skills/flow-status/SKILL.md scripts/check-contract-budget.sh
  git commit -m "docs(flow-contracts): resolve the change name through flow state resolve"
  ```

- [x] 6. `flow settings models` and the model-resolution block

**Build:** green

**Files:**
- Modify: `stats/cmd/flow/settings.go`
- Modify: `stats/cmd/flow/settings_test.go`
- Modify: `stats/cmd/flow/main.go`
- Modify: `skills/flow/SKILL.md`
- Modify: `scripts/check-model-resolution-shell.sh`

**Tests:** `TestSettingsCmd_Models`; `scripts/check-model-resolution-shell.sh` Case 4: project
key valid wins over a set store value; Case 5: project key invalid is reported on stderr and the
store value survives; Case 6: project key absent falls through to the store; Case 7: project key
valid with the store empty wins over the `fable` fallback — each case run for both keys
**Regression:** `TestSettingsCmd_Models` fails if the command prints anything but `ValidModels`
sorted one per line; Case 4/7 fail if the block stops reading the keys; Case 5 fails if an invalid
body is applied instead of dropped; Case 6 fails if an absent key overwrites the store value
**Baseline:** before=133 after=134
<!-- measured: cd stats && go test ./cmd/flow -count=1 -list '.*' | grep -c '^Test' @ ed2fbf8 was 128; task 4 adds five, this task one -->
**Commit:** `feat(flow): read the project's model keys in model resolution`

  - [ ] **Step 1: `flow settings models`**

  Test first: `TestSettingsCmd_Models` in `settings_test.go` drives `run` with
  `[]string{"settings", "models"}` and expects stdout equal to the sorted keys of
  `store.ValidModels`, one per line, exit 0. Then a `case "models":` in `runSettings` calling
  `runSettingsModels(stdout)`: collect `store.ValidModels` keys, `slices.Sort`, print. No flags,
  no store call. `settingsUsage` and `main.go`'s `usage` each gain one line.

  Run: `cd stats && gofmt -w . && go vet ./... && gofmt -l . && go test ./cmd/flow -count=1`
  Expected: pass, `gofmt -l` silent.

  - [ ] **Step 2: Extend the `## Model resolution` block in `skills/flow/SKILL.md`**

  Replace the two comment lines (`# <project>/.flow/project.md's ... wins`) with the reads. The
  block becomes:

  ```bash unverified:run scripts/check-model-resolution-shell.sh once step 3's cases exist
  MAIN_CHECKOUT="${MAIN_CHECKOUT:-$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)}"
  SETTINGS_JSON="$(flow settings get)"
  DEFAULT_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.defaultModel')"
  REVIEWERS="$(printf '%s' "$SETTINGS_JSON" | jq -r '.reviewers[]')"
  SELF_REVIEW_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.selfReviewModel // empty')"
  PROJECT_SRM="$(project-get.sh "$MAIN_CHECKOUT" 'self review model' 2>/dev/null | tr -d '`' | xargs)"
  if [ -n "$PROJECT_SRM" ]; then
    if flow settings models | grep -qx -- "$PROJECT_SRM"; then SELF_REVIEW_MODEL="$PROJECT_SRM"
    else echo "⚠ flow: .flow/project.md '## self review model' body '$PROJECT_SRM' is not a valid model — dropped" >&2; fi
  fi
  [ -z "$SELF_REVIEW_MODEL" ] && SELF_REVIEW_MODEL=fable
  PLANNING_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.planningModel // empty')"
  PROJECT_PM="$(project-get.sh "$MAIN_CHECKOUT" 'planning model' 2>/dev/null | tr -d '`' | xargs)"
  if [ -n "$PROJECT_PM" ]; then
    if flow settings models | grep -qx -- "$PROJECT_PM"; then PLANNING_MODEL="$PROJECT_PM"
    else echo "⚠ flow: .flow/project.md '## planning model' body '$PROJECT_PM' is not a valid model — dropped" >&2; fi
  fi
  [ -z "$PLANNING_MODEL" ] && PLANNING_MODEL=fable
  ```

  The prose paragraphs for `SELF_REVIEW_MODEL` and `PLANNING_MODEL` below the block already state
  the precedence; nothing there changes.

  - [ ] **Step 3: Extend `scripts/check-model-resolution-shell.sh`**

  The `flow` stub becomes a two-verb script: `settings get` prints the case's JSON, `settings
  models` prints `sonnet`, `opus`, `haiku`, `fable` one per line. `run_case` gains a fifth argument,
  the fixture `.flow/project.md` body (empty string for "no file"), written under a per-case
  `MAIN_CHECKOUT` temp dir that is exported to the block; the real `scripts/project-get.sh` is put
  on `PATH` by prepending `$SCRIPT_DIR`. Cases 1-3 pass an empty fixture and keep their
  expectations. Add Cases 4-7 from this task's **Tests:** field, each with `## planning model` and
  `## self review model` bodies written as backticked literals the way this repository's own
  `.flow/project.md` writes them. The final `MODEL-RESOLUTION-SHELL-OK: N case(s) checked` line
  reflects the new count.

  Run: `scripts/check-model-resolution-shell.sh`
  Expected: `MODEL-RESOLUTION-SHELL-OK: 7 case(s) checked`, exit 0.
  <!-- predicted: 3 cases at ed2fbf8 plus the four this task adds -->

  - [ ] **Step 4: Gates and commit**

  Run: `scripts/check-model-resolution-shell.sh && scripts/check-model-keys.sh && scripts/check-references.sh && scripts/check-contract-budget.sh && cd stats && go test ./... -race -count=1`
  Expected: all exit 0.

  ```bash verified:the commit subject is this task's Commit field
  git add stats/cmd/flow/settings.go stats/cmd/flow/settings_test.go stats/cmd/flow/main.go skills/flow/SKILL.md scripts/check-model-resolution-shell.sh
  git commit -m "feat(flow): read the project's model keys in model resolution"
  ```

- [x] 7. Measure to the leaf and run every guard

**Build:** green

**Files:**
- Modify: `spectre/changes/kan-379-resolve-project-md-keys-through-flow-scripts/design.md`

**Tests:** **none** — measurement and verification, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(spectre): record kan-379 after-figures per load set`

  - [ ] **Step 1: Run kan-378's measurement script with this change's load sets**

  Extract the `measure.sh` block from
  `spectre/changes/archive/kan-378-flow-deletion-only-cut-of-stale-restated-and/tasks.md` task 8
  into the session scratchpad (never the repository). Edit its `after` branch so set A omits
  `$C/project-configuration.md` and `$C/workspace-isolation.md` (neither is on the ordinary path
  any more) and set C keeps `$C/plan-provenance.md` (now the run half). Run
  `measure.sh <repo-root> after` at the branch tip.

  - [ ] **Step 2: Record the after column in `design.md`**

  In `## Measurement`, replace each `recorded by the measurement task` cell with the script's
  output for that set and add a provenance comment under the existing one:
  `<!-- measured: kan-378's measure.sh with task 7's set edits, run as measure.sh <repo> after @ branch spectre/kan-379-resolve-project-md-keys-through-flow-scripts -->`.
  Replace the `unverified:` tag on the `state resolve` JSON example with `verified:matches
  stateResolveOutput in stats/cmd/flow/state.go` once checked, and the two `unverified:` tags in
  `## 5` with `verified:scripts/check-model-resolution-shell.sh passes` — or leave them
  `unverified:` and say why.

  - [ ] **Step 3: Every lint command, every test command**

  Run every command under `## lint` in `.flow/project.md`, then every command under `## test`.
  Expected: every command exits 0.

  - [ ] **Step 4: Confirm the diff touches only the files this plan names**

  Run: `git diff --stat ed2fbf8 -- . ':!spectre/changes' ':!docs/superpowers'`
  Expected: exactly the union of every task's `**Files:**` field above.

  - [ ] **Step 5: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add spectre/changes/kan-379-resolve-project-md-keys-through-flow-scripts/design.md
  git commit -m "docs(spectre): record kan-379 after-figures per load set"
  ```
