package records

import (
	"encoding/json"
	"maps"
	"os"
	"testing"
)

type keyFixture struct {
	Key      string `json:"key"`
	Role     string `json:"role"`
	Opener   string `json:"opener"`
	Category string `json:"category"`
}

func TestCategoryTableClassesEveryFixtureKey(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("testdata/dispatch-keys.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []keyFixture
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("empty fixture")
	}
	for _, r := range rows {
		id, label := Category(r.Key, r.Role, r.Opener)
		if id != r.Category {
			t.Errorf("Category(%q, %q, %q) = %q, want %q", r.Key, r.Role, r.Opener, id, r.Category)
		}
		if label == "" {
			t.Errorf("Category(%q): empty label", r.Key)
		}
	}
}

func TestCategoryReadsTheFirstTaskOpener(t *testing.T) {
	t.Parallel()
	openers := map[string]string{"13": "verify-gymie", "14": "visual-verify-fix-1-rest-gymie"}
	key := "task-13+14-implementer"
	if got := FirstTaskID(key); got != "13" {
		t.Fatalf("FirstTaskID(%q) = %q, want 13", key, got)
	}
	if id, _ := Category(key, "implementer", openers[FirstTaskID(key)]); id != "in-run-fix-verify" {
		t.Fatalf("Category = %q, want in-run-fix-verify (task 13's opener)", id)
	}
	if got := FirstTaskID("panel-1-merger"); got != "" {
		t.Fatalf("FirstTaskID(panel-1-merger) = %q, want empty", got)
	}
}

func TestTaskOpenersReadsEachTasksOpener(t *testing.T) {
	t.Parallel()
	const plan = "# p\n\n" +
		"- [ ] 1. One\n\n**Files:** `a.go`\n**Commit:** one\n**Opener:** verify-gymie\n\n" +
		"- [ ] 2. Two\n  - [ ] **Step 1: x**\n\n```markdown\n**Opener:** fenced-not-read\n```\n\n**Files:** `b.go`\n\n" +
		"- [x] 3. Three\n  - [x] **Step 1: y**\n\n**Opener:** `visual-verify-fix-1-rest-gymie`\n**Files:** `c.go`\n"
	got := TaskOpeners([]byte(plan))
	want := map[string]string{"1": "verify-gymie", "3": "visual-verify-fix-1-rest-gymie"}
	if !maps.Equal(got, want) {
		t.Fatalf("TaskOpeners = %v, want %v", got, want)
	}
}
