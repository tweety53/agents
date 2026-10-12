package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the plan-suite-steps contract: a task body's unfenced inline
// command that runs a module or repository test suite with no selector is
// refused, naming the task and the line the span opens on; a selector-narrowed
// run, a lint command, a fenced example and prose outside any task pass.
func TestPlanSuiteSteps(t *testing.T) {
	cases := []struct {
		name, plan string
		code       int
		want       []string
	}{
		{"gradle module suite (kan-931 Task 3)",
			"- [ ] 3. Menu\n\n**Files:** `A.kt`\n\n  - [ ] **Step 3:** run `./gradlew :shared:desktopTest`; migrate failures\n",
			1, []string{"tasks.md:5: task 3 runs the suite `./gradlew :shared:desktopTest`"}},
		{"gradle suite span wrapped across lines",
			"- [ ] 7. Rows\n\n  - [ ] **Step 2:** Then the full `./gradlew\n    :shared:desktopTest` — green.\n",
			1, []string{"tasks.md:3: task 7 runs the suite `./gradlew :shared:desktopTest`"}},
		{"bare gradle test", "- [ ] 1. A\n\n- run `./gradlew test`\n", 1, []string{"task 1 runs the suite `./gradlew test`"}},
		{"go test with no -run", "- [ ] 1. A\n\n- `cd stats && go test ./... -race -count=1`\n", 1, []string{"task 1 runs the suite"}},
		{"npm test with no -t", "- [ ] 2. A\n\n- `cd stats/web && npm test`\n", 1, []string{"task 2 runs the suite"}},
		{"vitest with no -t", "- [ ] 2. A\n\n- `npx vitest run`\n", 1, []string{"task 2 runs the suite"}},
		{"two tasks, one hit each", "- [ ] 1. A\n\n- `./gradlew test`\n- [ ] 2. B\n\n- `go test ./...`\n", 1,
			[]string{"tasks.md:3: task 1", "tasks.md:6: task 2"}},
		{"gradle with --tests", "- [ ] 1. A\n\n- `./gradlew :shared:desktopTest --tests\n  'com.x.FooTest'`\n", 0, nil},
		{"verify bundle with --tests", "- [ ] 1. A\n\n- `./gradlew verifyChange :app:test --tests 'X'`\n", 0, nil},
		{"gradle lint and dev tasks", "- [ ] 1. A\n\n- `./gradlew ktlintCheck detekt`, `./gradlew verifyChange`, `./gradlew -q devStackFresh`\n", 0, nil},
		{"go test with -run", "- [ ] 1. A\n\n- `go test -run 'TestPlanSuiteSteps' ./internal/guard/`\n", 0, nil},
		{"vitest with -t", "- [ ] 1. A\n\n- `npm test -- -t 'renders'`\n", 0, nil},
		{"npm run test:visual is not npm test", "- [ ] 1. A\n\n- `npm run test:visual`\n", 0, nil},
		{"test names in Tests field", "- [ ] 1. A\n\n**Tests:** `a destructive row reads magenta-800`, `theTest`\n", 0, nil},
		{"prose naming gradle and a test class", "- [ ] 1. A\n\n**Baseline:** `unverified: a grep count of LocalDevSeedServiceTest, not a gradle run`\n", 0, nil},
		{"fenced example", "- [ ] 1. A\n\n```bash\n`./gradlew test`\n```\n", 0, nil},
		{"preamble outside any task", "## Global\n\n`./gradlew test` runs after the last task.\n\n- [ ] 1. A\n", 0, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(tc.plan), 0o644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := Registry["plan-suite-steps"]([]string{"tasks.md"}, Env{Dir: dir}, &stdout, &stderr)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
			}
			for _, w := range tc.want {
				if !strings.Contains(stdout.String(), w) {
					t.Errorf("stdout %q lacks %q", stdout.String(), w)
				}
			}
			if tc.code == 0 && stdout.Len() != 0 {
				t.Errorf("clean plan printed %q", stdout.String())
			}
		})
	}
}

func TestPlanSuiteStepsUsageAndUnreadable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Registry["plan-suite-steps"](nil, Env{}, &stdout, &stderr); code != 2 {
		t.Errorf("no args: exit %d, want 2", code)
	}
	if code := Registry["plan-suite-steps"]([]string{"missing.md"}, Env{Dir: t.TempDir()}, &stdout, &stderr); code != 2 {
		t.Errorf("missing file: exit %d, want 2", code)
	}
}
