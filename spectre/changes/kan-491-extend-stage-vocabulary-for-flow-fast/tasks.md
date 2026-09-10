# kan-491-extend-stage-vocabulary-for-flow-fast

> **Execution:** `/flow-fast` implements this plan. Mark a task's own
> checkbox when its commit lands and its verify step is green.
> **Relocation:** no

## Global constraints

- `stats/internal/stages/names_test.go` (package `stages_test`) re-derives
  `stages.Table` from README.md at test time; README's Commands cells and
  the Go rows move together, in this task, never separately.
- The README Commands cell parser splits on commas and trims backticks,
  preserving order; `TestStagesMatchReadmeLevelOne` compares with
  `reflect.DeepEqual`, so cell order must match the Go rows' `Commands`
  order: `` `/flow`, `/flow-fast` ``.
- flow-fast marks exactly the keys its SKILL.md Stage keys table names —
  never `flow.design-approval`, `flow.visual-verify`,
  `flow.verify-cleanup`, `flow.self-review`.

---

- [x] 1. Vocabulary carries /flow-fast, pinned to the SKILL table by a drift guard

**Files:**
- `stats/internal/stages/names.go`
- `stats/internal/stages/names_test.go`
- `README.md`

**Build:** green
**Tests:** `TestStageKeysMatchFlowFastSkillTable`, `TestStagesMatchReadmeLevelOne`
**Regression:** reverting this commit empties the `/flow-fast` vocabulary:
`TestStageKeysMatchFlowFastSkillTable` fails again, and every `/flow-fast`
stage mark is rejected by the CLI — the kan-357 defect returns.
**Baseline:** before=10 after=11
<!-- measured: go test ./internal/stages/ -count=1 -v @ branch spectre/kan-491-extend-stage-vocabulary-for-flow-fast (worktree, from main 9dbc70e) -->
<!-- predicted: go test ./internal/stages/ -count=1 -v after this task -->
**Commit:** feat(stats): stage vocabulary carries /flow-fast

  - [ ] **Step 1: Write the failing drift guard**

Add to `stats/internal/stages/names_test.go`, beside
`TestStagesMatchReadmeLevelOne`:

```go verified:authored in-tree for this change
// skillPath locates flow-fast's SKILL.md relative to this package, at the
// same depth as readmePath above.
const skillPath = "../../../skills/flow-fast/SKILL.md"

// stageKeysHeading locates the section the drift guard extracts from.
const stageKeysHeading = "## Stage keys"

// extractFlowFastSkillKeys reads flow-fast's own Stage keys table and
// returns every backticked flow.* key it names -- the authoritative
// statement of which keys /flow-fast marks, so the vocabulary can be
// pinned to the skill rather than to a hand-copied list.
func extractFlowFastSkillKeys(t *testing.T) map[string]bool {
	t.Helper()

	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read %s: %v", skillPath, err)
	}
	text := string(data)
	headingIdx := strings.Index(text, stageKeysHeading)
	if headingIdx == -1 {
		t.Fatalf("%s no longer contains the heading %q", skillPath, stageKeysHeading)
	}

	keys := map[string]bool{}
	for _, line := range strings.Split(text[headingIdx:], "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if len(keys) > 0 {
				break // the table has ended
			}
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		if strings.TrimSpace(cells[1]) == "Phase file" {
			continue // the header row
		}
		for _, tok := range strings.Split(cells[2], ",") {
			tok = strings.Trim(strings.TrimSpace(tok), "`")
			if strings.HasPrefix(tok, "flow.") {
				keys[tok] = true
			}
		}
	}
	if len(keys) == 0 {
		t.Fatalf("%s: extraction matched no stage keys -- the table's shape changed in a way this parser no longer understands", skillPath)
	}
	return keys
}

// TestStageKeysMatchFlowFastSkillTable pins the /flow-fast vocabulary to
// flow-fast's own Stage keys table: every key the skill names carries
// /flow-fast in stages.Table, and no other key does. Compared as sets --
// the skill groups by phase file, the vocabulary by README row.
func TestStageKeysMatchFlowFastSkillTable(t *testing.T) {
	want := extractFlowFastSkillKeys(t)

	got := map[string]bool{}
	for _, s := range stages.Table {
		for _, c := range s.Commands {
			if c == stages.FlowFast {
				got[s.Key] = true
			}
		}
	}

	missing := map[string]bool{}
	for key := range want {
		if !got[key] {
			missing[key] = true
		}
	}
	extra := map[string]bool{}
	for key := range got {
		if !want[key] {
			extra[key] = true
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		t.Errorf("/flow-fast vocabulary disagrees with %s\n  missing from vocabulary: %v\n  in vocabulary but not the skill: %v", skillPath, sortedKeys(missing), sortedKeys(extra))
	}
}

// sortedKeys renders a key set sorted, for a deterministic failure
// message.
func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
```

Add `"sort"` to the file's import block beside `"reflect"`.

  - [ ] **Step 2: Run it to verify it fails**

Run: `cd stats && go test ./internal/stages/ -run TestStageKeysMatchFlowFastSkillTable -count=1`
Expected: FAIL — compile error on `stages.FlowFast` undefined (the
vocabulary carries no `/flow-fast` yet), and once the constant exists,
`missing from vocabulary` naming all 26 keys.

  - [ ] **Step 3: Extend the vocabulary in both halves**

`stats/internal/stages/names.go` — the const block gains the second
command, and exactly the 26 rows the SKILL table names carry it after
`Flow`:

```go verified:authored in-tree for this change
const (
	Flow     Command = "/flow"
	FlowFast Command = "/flow-fast"
)
```

Rows gain `Commands: []Command{Flow, FlowFast}`; the four rows
flow-fast never marks (`flow.design-approval`, `flow.visual-verify`,
`flow.verify-cleanup`, `flow.self-review`) keep
`Commands: []Command{Flow}`.

`README.md` — the same 26 rows' Commands cell becomes
`` `/flow`, `/flow-fast` `` and the four keep `` `/flow` ``. The sentence
"every row below is both defined and run by `/flow`" (README.md:108-110)
is reworded to name `/flow-fast` as the one other command that runs rows,
still never defining any.

  - [ ] **Step 4: Run the tests to verify they pass, then targeted lint**

Run: `cd stats && go test ./internal/stages/ -count=1 -v`
Expected: PASS — 11 tests total
<!-- predicted: go test ./internal/stages/ -count=1 -v after this task: 10 measured before, +1 new -->, both vocabulary tests green together.
Run: `cd stats && gofmt -l . && go vet ./internal/stages/`
Expected: no output, exit 0.

  - [ ] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/internal/stages/names.go stats/internal/stages/names_test.go README.md
git commit -m "feat(stats): stage vocabulary carries /flow-fast"
```
