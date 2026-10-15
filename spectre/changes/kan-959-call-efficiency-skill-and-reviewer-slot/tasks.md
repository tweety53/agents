# kan-959-call-efficiency-skill-and-reviewer-slot

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a `/call-efficiency-audit` measuring skill, a `call-efficiency` review-panel slot, `test-sanity`
promoted from experimental, and one always-on `test-sanity+call-efficiency` pass-1 dispatch on every
panel, its model a reproducible 50/50 `sonnet`/`opus` roll at `low`.

**Spec:** `spectre/changes/kan-959-call-efficiency-skill-and-reviewer-slot/design.md` — read it
before any task; each task's `**Decision:**` names the entry it implements.

## Global constraints

- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before
  `gofmt -l .`.
- No test contacts the dev workspace's `flowd`; a `flow record` test passes `-addr` explicitly.
- The slot ids are exactly `test-sanity` and `call-efficiency`; the bundle is written
  `test-sanity+call-efficiency`, in that order, everywhere — plan-class output, `decision.json`'s
  `extra_dispatch.slots`, `flow record dispatch -slot`, `render-slot-prompt.sh`'s slot argument.
- The extra pair's effort is `low`, always; its model is `sonnet` when the roll is below 50, else
  `opus`.
- Every new or reworded sentence in a run-loaded file (`skills/**/*.md`, `rules/*`) is listed in
  `spectre/changes/kan-959-call-efficiency-skill-and-reviewer-slot/verbatim-moves.txt`, exactly as
  `scripts/check-verbatim-moves.sh`'s FAIL line prints it after `::`, one per line, under a `#`
  comment naming the task. That file is never staged by a task's commit (it rides the planning and
  archive commits); a task appends to it and leaves it unstaged.
- Never edit `~/.claude/CLAUDE.md` or `~/.zcode/` from a task — the Models-rule carve-out
  (decision `extra-dispatch-sonnet-carve-out`) is the parent's step after landing.

## Review Focus

- A decision written before this change (no `extra_dispatch` key) must still render and record — a
  resumed or fix run reads an old `decision.json`. Pinned in Task 2.
- `extra_dispatch` naming an off-policy model (`haiku`, `fable`, empty) is refused by the store like
  every other pair. Pinned in Task 2.
- The `micro` class prints the `extra:` line too — it is the one class whose tree consults no roll.
  Pinned in Task 1.
- Roll boundaries 49/50 pick `sonnet`/`opus`. Pinned in Task 1.
- The experimental directory left empty after Task 4 must read `none available`, never an error.
  Already pinned by `TestPlanClassTree`'s `nil` candidates; Task 4 runs it against the real tree.

- [x] 1. `plan-class.sh` prints the always-on extra dispatch's line
  - [x] **Step 1: Write the failing test.** Append to `stats/internal/guard/plan_class_test.go`:

```go unverified:compile-against-pcExtra
// TestPlanClassExtra pins the always-on test-sanity+call-efficiency
// dispatch's line (kan-959, decision extra-roll-own-line): sonnet below 50,
// opus from 50, always low, its roll printed — on every class, micro
// included, since the line is printed beside the tree, never inside it.
func TestPlanClassExtra(t *testing.T) {
	t.Parallel()
	for roll, want := range map[uint64]string{
		0:  "extra: test-sanity+call-efficiency · roll 0 · sonnet low\n",
		49: "extra: test-sanity+call-efficiency · roll 49 · sonnet low\n",
		50: "extra: test-sanity+call-efficiency · roll 50 · opus low\n",
		99: "extra: test-sanity+call-efficiency · roll 99 · opus low\n",
	} {
		if got := pcExtra(roll); got != want {
			t.Errorf("pcExtra(%d) = %q, want %q", roll, got, want)
		}
	}
}
```

    Run `cd stats && go test ./internal/guard/ -run 'TestPlanClassExtra' -count=1` — FAIL (`undefined: pcExtra`).
  - [x] **Step 2: Implement.** In `stats/internal/guard/planclass.go`, add after `pcTree`:

```go unverified:compile
// pcExtra is the always-on test-sanity+call-efficiency dispatch's line:
// its roll and the pair it picks — sonnet below 50, opus otherwise, always
// at low. Printed on every class, micro included (kan-959).
func pcExtra(roll uint64) string {
	model := "opus"
	if roll < 50 {
		model = "sonnet"
	}
	return fmt.Sprintf("extra: test-sanity+call-efficiency · roll %d · %s low\n", roll, model)
}
```

    and directly after `fmt.Fprint(stdout, pcTree(effective, roll(name), roll(name+"exp"), roll(name+"bundle"), cands))` add `fmt.Fprint(stdout, pcExtra(roll(name+"extra")))`.
  - [x] **Step 3: Update every end-to-end pin.** Every test in `plan_class_test.go` that compares the guard's whole stdout gains the eighth line for its change name: run each such case's guard, read the `extra:` line it now prints, and append it to the expected output literally (never recompute it in the test). `TestPlanClassTree`'s `pcTree` comparisons are unchanged.
  - [x] **Step 4: Header.** In `scripts/plan-class.sh`'s header comment, change `# Prints exactly seven lines to stdout:` to `# Prints exactly eight lines to stdout:`, add after the `experimental:` usage line `#   extra: test-sanity+call-efficiency · roll N · sonnet|opus low`, and add one sentence to the paragraph after the usage block: `# The extra line is printed on every class, micro included: the always-on test-sanity+call-efficiency dispatch's pair, sonnet when roll(<name>extra) < 50, else opus, always at low (kan-959).`
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestPlanClass' -count=1` — all exit 0, `gofmt -l` prints nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`, `scripts/plan-class.sh`
**Tests:** `TestPlanClassExtra`
**Regression:** reverting the commit removes `pcExtra`: the test fails to compile and the guard prints no `extra:` line.
**Baseline:** before=4 after=5
<!-- measured: grep -c '^func Test' stats/internal/guard/plan_class_test.go @ c5ff691c -->
**Commit:** feat(guard): plan-class prints the always-on extra dispatch's rolled pair
**After:** none

**Decision:** extra-roll-own-line

- [x] 2. The decision renders and the store checks `extra_dispatch`
  - [x] **Step 1: Write the failing render test.** Append to `stats/cmd/flow/decision_test.go`:

```go unverified:compile-against-ExtraDispatch
// TestDecisionRenderExtraDispatch pins kan-959's extra_dispatch: present,
// it prints one Extra dispatch line with its bundle and pair, on a micro
// decision too; absent (a decision written before kan-959), the render is
// unchanged and still exits 0.
func TestDecisionRenderExtraDispatch(t *testing.T) {
	micro := `{"class":"micro","classMechanical":"micro","override":null,` +
		`"inputs":{"tasks":1,"files":1,"repos":1,"migration":false,"spec":false,"red":false,"unverified":false},` +
		`"rolls":{"compact":65,"experimental":0,"bundle":43,"effort":96},"execution":"inline",` +
		`"implementer":"skipped — inline","fixer":"skipped — inline","panel":"default",` +
		`"groups":null,"groups_mechanical":null,"groups_override":null,"groups_reason":null`
	with := micro + `,"extra_dispatch":{"slots":["test-sanity","call-efficiency"],"model":"sonnet","effort":"low","reason":"extra roll 12"}}`
	rc, out, errOut := decisionRun(t, with, "render", "-file", "<file>", "-reviewers", "primary=opus, principles=opus")
	if rc != 0 {
		t.Fatalf("exit %d: %s", rc, errOut)
	}
	if want := "- **Extra dispatch:** test-sanity+call-efficiency sonnet/low\n"; !strings.Contains(out, want) {
		t.Errorf("render lacks %q:\n%s", want, out)
	}
	rc, out, errOut = decisionRun(t, micro+"}", "render", "-file", "<file>", "-reviewers", "primary=opus, principles=opus")
	if rc != 0 {
		t.Fatalf("a decision without extra_dispatch: exit %d: %s", rc, errOut)
	}
	if strings.Contains(out, "Extra dispatch") {
		t.Errorf("a decision without extra_dispatch renders an Extra dispatch line:\n%s", out)
	}
}
```

    Run `cd stats && go test ./cmd/flow/ -run 'TestDecisionRenderExtraDispatch' -count=1` — FAIL (no Extra dispatch line).
  - [x] **Step 2: Implement the render.** In `stats/cmd/flow/decision.go`, add to `dDecision` after `Visual`:

```go unverified:compile
	ExtraDispatch    *struct {
		Slots  []string `json:"slots"`
		Model  string   `json:"model"`
		Effort string   `json:"effort"`
	} `json:"extra_dispatch"`
```

    and in `renderDecision`, directly after the `if panelDefault != "" { … } else { … }` block that writes the Panel line:

```go unverified:compile
	if x := d.ExtraDispatch; x != nil {
		line("Extra dispatch", strings.Join(x.Slots, "+")+" "+dPair{Model: x.Model, Effort: x.Effort}.pair())
	}
```

    `extra_dispatch` stays out of `dRequired`: an older decision has none.
  - [x] **Step 3: Write the failing store test.** In `stats/internal/api/records_test.go`, append:

```go unverified:compile
// TestRecordDecisionChecksExtraDispatchModel pins kan-959: the always-on
// extra_dispatch pair is checked like every other pair — an off-policy model
// is refused naming extra_dispatch.model; sonnet and opus record.
func TestRecordDecisionChecksExtraDispatchModel(t *testing.T) {
	body := func(model string) string {
		return `{"class":"micro","execution":"inline","implementer":"skipped — inline","fixer":"skipped — inline","panel":"default","groups":null,` +
			`"extra_dispatch":{"slots":["test-sanity","call-efficiency"],"model":"` + model + `","effort":"low","reason":"r"}}`
	}
	for _, bad := range []string{"haiku", "fable", ""} {
		t.Run("refuses "+bad, func(t *testing.T) {
			_, fs := recordTestServer(t, "proj", "kan-1")
			_, _, err := api.ApplyDecisionRecord(context.Background(), fs, "proj", "kan-1",
				records.Decision{SessionToken: "mf-decide-1", Decision: json.RawMessage(body(bad))})
			if !errors.Is(err, api.ErrInvalidRecord) {
				t.Fatalf("ApplyDecisionRecord error = %v, want ErrInvalidRecord", err)
			}
			if want := fmt.Sprintf("extra_dispatch.model %q", bad); !strings.Contains(err.Error(), want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), want)
			}
			if fs.recordCalls != 0 {
				t.Errorf("the store was reached for an off-policy extra_dispatch")
			}
		})
	}
	for _, good := range []string{"sonnet", "opus"} {
		t.Run("records "+good, func(t *testing.T) {
			_, fs := recordTestServer(t, "proj", "kan-1")
			if _, _, err := api.ApplyDecisionRecord(context.Background(), fs, "proj", "kan-1",
				records.Decision{SessionToken: "mf-decide-1", Decision: json.RawMessage(body(good))}); err != nil {
				t.Fatalf("ApplyDecisionRecord: %v", err)
			}
		})
	}
}
```

    Run `cd stats && go test ./internal/api/ -run 'TestRecordDecisionChecksExtraDispatchModel' -count=1` — the `refuses` subtests FAIL.
  - [x] **Step 4: Implement the check.** In `checkDecisionModels` (`stats/internal/api/records.go`), after the `fixer` check add `if err := checkPairModel("extra_dispatch", d["extra_dispatch"], false); err != nil { return err }`, and add `extra_dispatch.model` to the doc comment's list of checked pairs.
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ ./internal/api/ && go test ./cmd/flow/ -run 'TestDecisionRender' -count=1 && go test ./internal/api/ -run 'TestRecordDecision' -count=1` — all exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go`, `stats/internal/api/records.go`, `stats/internal/api/records_test.go`
**Tests:** `TestDecisionRenderExtraDispatch`, `TestRecordDecisionChecksExtraDispatchModel`
**Regression:** reverting the commit drops the Extra dispatch line from the render and lets `extra_dispatch.model` `haiku` reach the store.
**Baseline:** before=55 after=57
<!-- measured: cat stats/cmd/flow/decision_test.go stats/internal/api/records_test.go | grep -c '^func Test' @ c5ff691c -->
**Commit:** feat(flow-cli): render and check the decision's extra_dispatch pair
**After:** none

**Decision:** extra-dispatch-top-level-pair

Correction (2026-10-15): the plan's render test called `decision render` without `-session-model`, which the command requires (exit 2, usage); the shipped test passes `-session-model opus` on both calls. Assertions unchanged.

- [x] 3. `test-sanity` becomes a regular slot template
  - [x] **Step 1: Move.** `git mv skills/flow/experimental/test-sanity.md skills/flow/test-sanity-reviewer-prompt.md`.
  - [x] **Step 2: Drop the experimental framing.** In the moved file: delete line 1 (`description: …`) and the blank line after it; retitle `# Test-Sanity Reviewer Prompt Template`; replace the intro paragraph (the one naming the `exp-test-sanity` slot and **Experimental slot**) with `The panel's \`test-sanity\` slot, one of the two roles of the always-on extra dispatch — see **The always-on extra dispatch** (\`skills/flow/review-panel.md\`).`; change `You are an EXPERIMENTAL test-sanity reviewer.` to `You are a test-sanity reviewer.`; in the `[DIFF_PATH]` placeholder note change `slot-delta-<round>-exp-test-sanity.diff` to `slot-delta-<round>-test-sanity.diff`. Change nothing else in the template body.
  - [x] **Step 3: Experimental directory.** If `skills/flow/experimental/` is now empty, leave it absent (git tracks no empty directory) — `pcCandidates` already reads an absent directory as no candidates.
  - [x] **Step 4: Verify.** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh` exit 0; `scripts/check-verbatim-moves.sh` exits 0 once this task's FAIL sentences are listed in `verbatim-moves.txt`; `cd stats && go test ./internal/guard/ -run 'TestPlanClass|TestRenderSlotPrompt' -count=1` exits 0 (the real tree now has no experimental candidate). `grep -rn 'exp-test-sanity\|experimental/test-sanity' skills/ scripts/ stats/` prints only `render_slot_prompt_test.go` hits, which Task 5 owns. Commit.

**Build:** green
**Files:** `skills/flow/test-sanity-reviewer-prompt.md`, `skills/flow/experimental/test-sanity.md`, `stats/internal/guard/render_slot_prompt_test.go`
**Tests:** none — repairs existing tests: none; the moved file is a prompt template, not a test, and the guards in step 4 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** refactor(flow): promote test-sanity from experimental to a regular slot template
**After:** none

**Decision:** test-sanity-promoted

Correction (2026-10-15): step 4 required `TestRenderSlotPrompt` green, but its exp-slot subtest read the moved `experimental/test-sanity.md` from the real tree; the subtest now renders an `exp-probe` candidate from a copy of the skill's top-level files, so it moved into this task (`**Files:**` widened) and Task 5's step-2 update has nothing left to do. The new intro's **The always-on extra dispatch** citation resolves only once Task 7 adds that heading; `check-references.sh` reports it until then.

- [x] 4. The `call-efficiency` reviewer prompt
  - [x] **Step 1: Write `skills/flow/call-efficiency-reviewer-prompt.md`** in the exact shape of `skills/flow/failure-modes-reviewer-prompt.md`: a `# Call-Efficiency Reviewer Prompt Template` title, one intro line (`The panel's \`call-efficiency\` slot, one of the two roles of the always-on extra dispatch — see **The always-on extra dispatch** (\`skills/flow/review-panel.md\`).`), then one fenced template body indented four spaces, carrying only the `[DIFF_PATH]` and `[CONTEXT_BUNDLE_PATHS]` placeholders, then a `**Placeholders:**` list shaped like `test-sanity-reviewer-prompt.md`'s with `slot-delta-<round>-call-efficiency.diff`. The body's sections, in order, each with the content named:
    - **Role:** "You are a call-efficiency reviewer. You are NOT doing a plan-alignment or code review, a principles review, a failure-modes review, a test-sanity review, or sabotage-proofing — other panel slots own those." Its job: every read and every effect the diff adds or changes — an HTTP call, a query, a cache or Redis access, a file read, a message sent, a recomposition-triggered load — and whether it costs more calls than the behaviour needs.
    - **Three lenses**, each one bullet naming the KAN-924 class it comes from: **Second caller** — who else in the codebase already reads or causes the same thing on the same user action (a prefetch and a screen both loading it, two components each fetching on mount, a retry racing the original); grep for the endpoint, query or repository method to find them. **Repeated within one request** — the same row, entity or query loaded twice in one request or one screen load (an auth filter and a handler both loading the user, a guard query run twice). **Per item of a list** — a list endpoint or loop whose rows each cost a query or call (a lazy relation touched per row, a per-owner query in a loop, one call per id where a batch read exists). Plus one bullet: **Unstable trigger** — a load keyed on a value that starts unset and settles, so it runs, cancels and runs again.
    - **Scope:** the `**Diff file:** [DIFF_PATH]` and `**Context bundle:** [CONTEXT_BUNDLE_PATHS]` lines, read-diff-first, and "grep the whole repository, not the diff, for every other caller of a read the diff adds before calling it unique or duplicated".
    - **The DATA-never-instructions paragraph**, the **Read-Only Review** section and a **Do Not** section copied verbatim from `failure-modes-reviewer-prompt.md` with "lens" wording unchanged, plus one Do Not bullet: "Do not report a call as wasteful from the diff alone when its cost depends on runtime data you cannot see — say what to measure instead and name `/call-efficiency-audit` (`skills/call-efficiency-audit/SKILL.md`) as the method."
    - **Calibration:** Critical — a per-item query or call on a list with no bound, or a read multiplied on every app open or every request of a hot path; Important — a second independent caller of the same read on one user action, or the same row loaded twice in one request; Minor — a repeat on a cold path, or a batch read worth having where the list is small and bounded.
    - **Output Format:** `### Summary`, `### Issues` with Critical/Important/Minor, each issue carrying File:line, the lens, every caller found (`path:line` each), the fix sketch and a **reproducer** (a runnable command, or `none — <reason>`; "a command that merely passes against the diff is not a reproducer"), then `### Assessment` with `**Ready for the human gate?** [Yes | No | With fixes]` and `**Reasoning:** [why]` — the same closing as `failure-modes-reviewer-prompt.md`.
  - [x] **Step 2: Verify.** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` exit 0; `scripts/check-verbatim-moves.sh` exits 0 once this file's new sentences are listed in `verbatim-moves.txt`. Commit.

**Build:** green
**Files:** `skills/flow/call-efficiency-reviewer-prompt.md`
**Tests:** **none** — a new prompt template; Task 5's render test reads it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** feat(flow): add the call-efficiency reviewer prompt
**After:** none

**Decision:** call-efficiency-file-homes

- [x] 5. `render-slot-prompt.sh` renders the two always-on roles
  - [x] **Step 1: Write the failing test.** Append to `stats/internal/guard/render_slot_prompt_test.go`:

```go unverified:compile-against-rspNewFx
// TestRenderSlotPromptAlwaysOnRoles pins kan-959: the always-on extra
// dispatch's bundle renders one PASS section per role from the real
// templates, each pointed at its own report, and the old exp- id for
// test-sanity no longer resolves.
func TestRenderSlotPromptAlwaysOnRoles(t *testing.T) {
	t.Parallel()
	skill := rspSkillDir(t)
	fx := rspNewFx(t)
	sdd := fx.canon + "/.superpowers/sdd/"
	r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "test-sanity+call-efficiency", "-diff", "final", "--", fx.canon)
	if r.rc != 0 {
		t.Fatalf("exit %d\n%s", r.rc, r.out)
	}
	for _, want := range []string{
		"## PASS test-sanity", "## PASS call-efficiency",
		"`" + sdd + "panel-report-0-test-sanity.md`", "`" + sdd + "panel-report-0-call-efficiency.md`",
		"You are a test-sanity reviewer.", "You are a call-efficiency reviewer.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("render lacks %q", want)
		}
	}
	if r, _ := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "exp-test-sanity", "-diff", "final", "--", fx.canon); r.rc != 2 {
		t.Errorf("exp-test-sanity: exit %d, want 2 — its template moved", r.rc)
	}
}
```

    Before running, read `rspNewFx`/`fx.run` in the same file and match their exact signatures; adjust only the call shape, never the assertions. Run `cd stats && go test ./internal/guard/ -run 'TestRenderSlotPromptAlwaysOnRoles' -count=1` — FAIL (`no reviewer template for slot "test-sanity"`).
  - [x] **Step 2: Implement.** In `stats/internal/guard/renderslotprompt.go`, `rspSlots` gains `"test-sanity": true, "call-efficiency": true`; update its comment to name them. Update any existing test in the file that rendered `exp-test-sanity` from the real skill directory to render `test-sanity` instead.
  - [x] **Step 3: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1` — all exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`
**Tests:** `TestRenderSlotPromptAlwaysOnRoles`
**Regression:** reverting the commit refuses `test-sanity+call-efficiency` with exit 2: no brief can be rendered for the extra dispatch.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/render_slot_prompt_test.go @ c5ff691c -->
**Commit:** feat(guard): render-slot-prompt knows the test-sanity and call-efficiency slots
**After:** Task 3, 4

**Decision:** test-sanity-promoted

- [x] 6. The `/call-efficiency-audit` measuring skill
  - [x] **Step 1: Write `skills/call-efficiency-audit/SKILL.md`.** Frontmatter `name: call-efficiency-audit` and a `description:` naming the trigger (`/call-efficiency-audit`, "audit call efficiency", "duplicate requests", "N+1") and that it measures a running dev stack. The body generalises `gymie/docs/audits/kan-924-call-efficiency.md`'s Method and checklists — no gymie paths, endpoints or ids — in these sections:
    - **Never production** — the stack is a dev or worktree stack; every setting changed (statement logging) is set on that stack's own database only and reset afterwards; cite `rules/never-touch-production.mdc`.
    - **Scope** — the layers to cover (client app, gateway, database, cache) and naming what is out of scope in the report.
    - **Browser request log** — a fresh browser context (no site data, empty offline store), signed in by API, app loaded, every request to the backend recorded for a fixed window with start and end times; three cold opens; then a fixed walk across the main screens; per-path count table with ids folded to `{id}`; the repeats table — read, calls per run, whether the second call started before the first answered, and the callers found in code. A CDP `Network.requestWillBeSent` log of one more open counts CORS preflights.
    - **Per-endpoint statement count** — statement logging set on the stack's database only (`ALTER DATABASE … SET log_statement = 'all'`, reset afterwards), each endpoint called once warm, statements logged in the window after the call counted, `BEGIN`/`COMMIT`/`SET` excluded and transactions counted from `BEGIN`; index use checked with `EXPLAIN` on the same database.
    - **Cache** — `redis-cli MONITOR` (or the cache's equivalent) over one full walk: repeated key reads, missing TTLs, `KEYS`/`SCAN` on hot paths, per-request limiter cost.
    - **Checklists** — one table per layer, each row a check with a verdict and evidence column, carrying every check row of the KAN-924 report's Frontend, Postgres, Redis and Gateway tables generalised (repeated reads on open, waterfalls, refresh stampede, uncancelled reads, polling while hidden, retries without backoff, effects keyed on unstable values, prefetch of unopened data, dead reads, wide reads on every visit; lazy relations in lists, entity loaded twice per request, one transaction per repository call, per-owner queries, repeated selects in one read, missing FK indexes, save-in-loop, transactions spanning external calls, `OFFSET`/`COUNT(*)` paging, scheduled jobs scanning tables; cache rows as above; CORS preflights and `maxAge`, identity resolved twice, per-row presigned URLs, outbound calls in a transaction).
    - **Findings** — the table shape `ID | Area | Evidence | Severity | Action`, every finding with measured evidence, and a **Prevention** table grouping recurring classes (two or more instances); the three KAN-924 classes named as the ones the `call-efficiency` panel slot (`skills/flow/call-efficiency-reviewer-prompt.md`) asks about.
    - **Before/after** — after a fix, the same measurement on the same stack and the before/after table.
    - **Commit nothing measuring** — the tooling is throwaway; only the report is kept.
  - [x] **Step 2: Index it.** Add one row to the skill index table in `CLAUDE.md` and the same row in `AGENTS.md` (each after the `skills/flow-self-review/` row): `| \`skills/call-efficiency-audit/\` | \`/call-efficiency-audit\` | Measure a running dev stack's call efficiency — browser request log, per-endpoint SQL statement counts, cache \`MONITOR\` — and report repeated reads, repeated queries within a request and per-item queries. Standalone, not a pipeline stage |`. Confirm `setup.sh global` installs every `skills/*/` directory without a list to extend (read its skills loop); if it keeps a list, add `call-efficiency-audit` to it and add that file to this task's `**Files:**`.
  - [x] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-hand-notes-in-step.sh` exit 0; `scripts/check-verbatim-moves.sh` exits 0 once the new sentences are listed in `verbatim-moves.txt`. Commit.

**Build:** green
**Files:** `skills/call-efficiency-audit/SKILL.md`, `CLAUDE.md`, `AGENTS.md`, `stats/internal/guard/references.go`
**Tests:** **none** — a new standalone skill and two index rows; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** feat(call-efficiency-audit): add the call-efficiency-audit measuring skill
**After:** Task 4

**Decision:** call-efficiency-file-homes

Correction (2026-10-15): `check-references.sh` refused the new skill as `0 checked, and not declared expected-zero` — its citations are plain parenthetical paths, never bold-adjacent — so it joins the contract/index-doc category of `crExpectedZero` (`stats/internal/guard/references.go`, `**Files:**` widened). The skill and its slash command are named `call-efficiency-audit` at the operator's instruction (2026-10-15); the panel slot stays `call-efficiency`.

- [x] 7. The pipeline dispatches the always-on extra bundle
  - [x] **Step 1: Decide (`skills/flow/brainstorm-planner.md`).** In **Decide**:
    - after step 3 (**review panel**) insert a step, renumbering the later steps and every in-file "step N" reference to them:

```markdown verified:wording-from-design-decisions
4. **extra dispatch** — on every class, `micro` included: record `plan-class.sh`'s `extra:` line as
   `extra_dispatch` `{slots, model, effort, reason}` — `slots` `["test-sanity", "call-efficiency"]`,
   `model` and `effort` the line's pair, `reason` naming its roll. It is never chosen: the roll
   decides the model and the effort is `low`, always.
```

    - in **The micro row**'s sentence, after "the decision record itself," add "the extra dispatch,".
    - in **Model and effort**'s bullet list, after the `every pass-1 panel dispatch` bullet, add `- **the extra dispatch** — its rolled pair, \`sonnet\` or \`opus\` at \`low\`: the one pass-1 panel dispatch that may run on \`sonnet\` (decision \`extra-dispatch-sonnet-carve-out\`, kan-959);` and change the preceding bullet to `- **every other pass-1 panel dispatch** — \`opus\`, always;`.
    - in the JSON field paragraph ("The JSON carries: …"), after the `visual` entry add `` `extra_dispatch` (an object `{slots, model, effort, reason}`, recorded on every class), ``.
  - [x] **Step 2: Review panel (`skills/flow/review-panel.md`).** Add a `### The always-on extra dispatch` subsection directly before `### The docs-only reduction`:

```markdown verified:wording-from-design-decisions
### The always-on extra dispatch

**Every pass 1 dispatches one more bundle, `test-sanity+call-efficiency`, on the decision's
`extra_dispatch` pair** — on a decided and a `default` panel alike, beside the dispatches the
roster resolves, and outside their `≤2 × ≤3` cap. No automatic reduction removes it: the docs-only
reduction, the late-fix reduction and the small append each still dispatch it. A decision with no
`extra_dispatch` — written before kan-959 — dispatches it on `opus` at `low`.

| id | Slot | How to spawn |
|---|------|---------------|
| `test-sanity` | **Test-sanity** — would each test the diff adds or changes fail if its behaviour broke | `flow-low` + `test-sanity-reviewer-prompt.md` |
| `call-efficiency` | **Call-efficiency** — second callers, a read repeated in one request, a query per list item | `flow-low` + `call-efficiency-reviewer-prompt.md` |

It is rendered, recorded and its findings handled like any other bundle (**Bundled dispatch**
below): one `dispatches` row with `-slot test-sanity+call-efficiency` and the pair's `-model` and
`-effort`. A role from it that raised a finding re-runs like any other role (**Panel re-runs**,
`skills/flow/review-panel-fix-round.md`).
```

    - In the opening paragraph, after "This file owns dispatch: mapping each resolved id to its slot and spawning it." add "Every pass 1 also dispatches **The always-on extra dispatch** below."
    - In **The docs-only reduction**'s first paragraph change "reduces pass 1 to `primary` alone**, plus every slot" to "reduces pass 1 to `primary` alone**, plus **The always-on extra dispatch** and every slot".
    - In the paragraph opening "**This, the late-fix reduction below and the small append", after "No slot is ever added by diff size, touched area, or any other automatic trigger" add ", and none removes the always-on extra dispatch".
  - [x] **Step 3: Late-fix and fix-round files.** In `skills/flow/review-panel-late-fix.md`, wherever the late-fix reduction or the small append names the single dispatch it runs, add that the always-on extra dispatch still runs beside it (cite **The always-on extra dispatch**, `skills/flow/review-panel.md`). In `skills/flow/review-panel-fix-round.md`'s **Panel re-runs**, add one sentence: "A `test-sanity` or `call-efficiency` role re-runs in the round's rerun bundle like any other role, on `opus` on a `default` panel."
  - [x] **Step 4: `/flow-fast` (`skills/flow-fast/SKILL.md`).** Change the review-panel bullet's "A `default` panel (the `micro` class) runs no panel." to "A `default` panel (the `micro` class) runs **The always-on extra dispatch** (`skills/flow/review-panel.md`) alone."
  - [x] **Step 5: Experimental-slot file.** `grep -n 'test-sanity' skills/flow/review-panel-experimental-slot.md` — replace any example naming `test-sanity` with the generic `exp-<name>`; touch nothing else.
  - [x] **Step 6: Verify.** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-ask-silence.sh` exit 0; `scripts/check-normative-inventory.sh` output unchanged against a capture taken before step 1; `scripts/check-verbatim-moves.sh` exits 0 once every new or reworded sentence is listed in `verbatim-moves.txt`. `scripts/plan-class.sh spectre/changes/kan-959-call-efficiency-skill-and-reviewer-slot/tasks.md 1` prints an `extra:` line. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/review-panel.md`, `skills/flow/review-panel-late-fix.md`, `skills/flow/review-panel-fix-round.md`, `skills/flow-fast/SKILL.md`, `skills/flow/SKILL-rationale.md`, `scripts/plan-dispatch-groups.sh`, `stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/guardsymlinks.go`
**Tests:** **none** — pipeline prose; Tasks 1, 2 and 5 pin the mechanism it describes, and the guards in step 6 cover the prose.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** feat(flow): dispatch the always-on test-sanity+call-efficiency bundle on every panel
**After:** Task 1, 2, 3, 4, 5

**Decision:** extra-dispatch-always
**Decision:** extra-rerun-unchanged
**Decision:** extra-dispatch-sonnet-carve-out

Correction (2026-10-15): the Decide renumbering also moved two cross-file step citations — `scripts/plan-dispatch-groups.sh`'s header (step 4 → 5) and `skills/flow/SKILL-rationale.md`'s visual-verification heading (step 5 → 6). The full lint set found the Task 6 skill undeclared in two more coverage guards — `check-stage-mark-calls.sh` and `check-guard-symlinks.sh` — so it is declared expected-zero in `stagemarkcalls.go` and `guardsymlinks.go` here. Step 5 found no `test-sanity` example in `review-panel-experimental-slot.md`; nothing changed there.
