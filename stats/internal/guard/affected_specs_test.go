package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAffectedSpecs pins the mapping flow-guard affected-specs prints: a
// per-screen change selects that screen's one spec, a shared change selects
// every spec whose screen imports it, and a UI file nothing reaches selects
// them all.
func TestAffectedSpecs(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	root := t.TempDir()
	files := map[string]string{
		".flow/project.md":            "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**` |\n",
		"web/package.json":            "{}\n",
		"web/src/pages/Home.tsx":      "import { Chip } from \"../shared/Chip\";\nexport function Home() { return <Chip /> }\n",
		"web/src/pages/Rooms.tsx":     "import { Chip } from \"../shared/Chip\";\nimport { PushedHeader } from \"../shared/Header\";\nexport function Rooms() { return <PushedHeader /> }\n",
		"web/src/shared/Chip.tsx":     "import { tokens } from \"./tokens\";\nexport function Chip() { return tokens.gap }\n",
		"web/src/shared/Header.tsx":   "export function PushedHeader() { return null }\n",
		"web/src/shared/tokens.ts":    "export const tokens = { gap: 4 };\n",
		"web/src/App.tsx":             "import { Home } from \"./pages/Home\";\nimport { Rooms } from \"./pages/Rooms\";\nimport { Bar } from \"./shared/Bar\";\nimport { Badge } from \"./shared/Badge\";\nexport function App() { return <Bar /> }\n",
		"web/src/shared/Bar.tsx":      "export function Bar() { return null }\n",
		"web/src/shared/Badge.tsx":    "export function Badge() { return null }\n",
		"web/src/pages/Day.tsx":       "import { Badge } from \"../shared/Badge\";\nexport function Day() { return <Badge /> }\n",
		"web/src/pages/Home.test.tsx": "import { Home } from \"./Home\";\n",
		"web/src/main.css":            "body { margin: 0 }\n",
		// Kotlin: a same-package use carries no import line.
		"web/src/kt/Palette.kt":                     "package ui\n\nobject Palette { val ink = 1 }\n",
		"web/src/kt/DayScreen.kt":                   "package ui\n\n@Composable\nfun DayScreen() { Palette.ink }\n",
		"web/tests/home.spec.ts":                    "test('home', () => {})\n",
		"web/tests/home.spec.ts.sources":            "# the home screen\nweb/src/pages/Home.tsx\n",
		"web/tests/rooms.spec.ts":                   "import { open } from \"./support\";\ntest('rooms', () => {})\n",
		"web/tests/rooms.spec.ts.sources":           "web/src/pages/Rooms.tsx\n",
		"web/tests/day.spec.ts":                     "test('day', () => {})\n",
		"web/tests/day.spec.ts.sources":             "web/src/kt/Day*.kt\n",
		"web/tests/support.ts":                      "export function open() {}\n",
		"web/tests/home.spec.ts-snapshots/home.png": "png",
		"web/tests/all.sources":                     "# what every spec renders\nseed/**\n",
		"seed/fixture.sql":                          "insert\n",
		"docs/notes.md":                             "notes\n",
	}
	for p, body := range files {
		writeFile(t, filepath.Join(root, p), body)
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")

	const home, rooms, day = "tests/home.spec.ts", "tests/rooms.spec.ts", "tests/day.spec.ts"
	all := []string{day, home, rooms}
	cases := []struct {
		name, touch string
		want        []string
		errHas      string
	}{
		{"a per-screen file selects its own spec", "web/src/pages/Home.tsx", []string{home}, ""},
		{"a shared file selects every spec importing it", "web/src/shared/Chip.tsx", []string{home, rooms}, ""},
		{"a shared file one screen imports selects that spec", "web/src/shared/Header.tsx", []string{rooms}, ""},
		{"tokens reach every importer transitively", "web/src/shared/tokens.ts", []string{home, rooms}, ""},
		{"a Kotlin name used with no import line is an import", "web/src/kt/Palette.kt", []string{day}, ""},
		{"the shell selects all", "web/src/shared/Bar.tsx", all, "all: web/src/shared/Bar.tsx"},
		{"a file one screen and the shell both draw selects all", "web/src/shared/Badge.tsx", all, "all: web/src/shared/Badge.tsx"},
		{"a unit test moves no capture", "web/src/pages/Home.test.tsx", nil, "test: web/src/pages/Home.test.tsx"},
		{"a UI file nothing reaches selects all", "web/src/main.css", all, "all: web/src/main.css"},
		{"a seed listed in all.sources selects all", "seed/fixture.sql", all, "all: seed/fixture.sql — listed in all.sources"},
		{"a non-UI file nothing reaches selects none", "docs/notes.md", nil, "unreached: docs/notes.md"},
		{"a spec selects itself", "web/tests/home.spec.ts", []string{home}, ""},
		{"a sidecar selects its spec", "web/tests/rooms.spec.ts.sources", []string{rooms}, ""},
		{"a spec helper selects its importers", "web/tests/support.ts", []string{rooms}, ""},
		{"a baseline PNG selects nothing", "web/tests/home.spec.ts-snapshots/home.png", nil, ""},
	}
	for _, c := range cases {
		writeFile(t, filepath.Join(root, c.touch), files[c.touch]+"\n/* edited */\n")
		r := runGuard("affected-specs", []string{root, "HEAD", filepath.Join(root, "web/tests")}, env)
		if got := strings.Fields(r.stdout); r.rc != 0 || strings.Join(got, " ") != strings.Join(c.want, " ") || !strings.Contains(r.err, c.errHas) {
			t.Errorf("%s: rc=%d stdout=%q err=%q, want %q", c.name, r.rc, r.stdout, r.err, c.want)
		}
		writeFile(t, filepath.Join(root, c.touch), files[c.touch])
	}

	// The task set: only a file a sidecar names selects a spec, beside the
	// specs the task's Tests and Regression name.
	plan := filepath.Join(t.TempDir(), "tasks.md")
	writeFile(t, plan, "- [ ] 7. In-run fix 1 — visual-verify: x\n\n**Tests:** none\n**Regression:** e2e/tests/rooms.spec.ts renders the seed\n")
	for touch, want := range map[string]string{
		"seed/fixture.sql":        rooms,
		"web/src/shared/Chip.tsx": rooms,
		"web/src/pages/Home.tsx":  home + " " + rooms,
		"web/tests/day.spec.ts":   day + " " + rooms,
	} {
		writeFile(t, filepath.Join(root, touch), files[touch]+"\n/* edited */\n")
		r := runGuard("affected-specs", []string{"-task", plan + ":7", root, "HEAD", filepath.Join(root, "web/tests")}, env)
		if got := strings.Join(strings.Fields(r.stdout), " "); r.rc != 0 || got != want || strings.Contains(r.err, "all:") {
			t.Errorf("-task, %s: rc=%d stdout=%q err=%q, want %q", touch, r.rc, r.stdout, r.err, want)
		}
		writeFile(t, filepath.Join(root, touch), files[touch])
	}
	if r := runGuard("affected-specs", []string{"-task", plan + ":8", root, "HEAD", filepath.Join(root, "web/tests")}, env); r.rc != 2 {
		t.Errorf("-task naming no task: rc=%d, want 2", r.rc)
	}

	// A spec with no sidecar renders an unknown screen: any UI change selects it.
	writeFile(t, filepath.Join(root, "web/tests/legacy.spec.ts"), "test('legacy', () => {})\n")
	writeFile(t, filepath.Join(root, "web/src/pages/Home.tsx"), files["web/src/pages/Home.tsx"]+"// edited\n")
	r := runGuard("affected-specs", []string{root, "HEAD", filepath.Join(root, "web/tests")}, env)
	if r.rc != 0 || r.stdout != home+"\ntests/legacy.spec.ts\n" {
		t.Errorf("unmapped spec: rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	// The same edit, committed, read as a range.
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "edit")
	r = runGuard("affected-specs", []string{root, "HEAD^..HEAD", filepath.Join(root, "web/tests")}, env)
	if r.rc != 0 || !strings.Contains(r.stdout, home) || strings.Contains(r.stdout, rooms) {
		t.Errorf("range: rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}

	for _, args := range [][]string{{root, "HEAD"}, {root, "no-such-rev", filepath.Join(root, "web/tests")}, {root, "HEAD", filepath.Join(root, "missing")}} {
		if r := runGuard("affected-specs", args, env); r.rc != 2 || r.stdout != "" {
			t.Errorf("%v: rc=%d out=%q, want exit 2", args[1:], r.rc, r.out)
		}
	}
}

// TestAffectedSpecsSpecRootEditIsNoUIChange pins kan-1099 self-review F2: with
// no `ui paths` row every file is UI, but a file inside the spec root is not,
// so a spec edit selects itself, not every unmapped spec.
func TestAffectedSpecsSpecRootEditIsNoUIChange(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	root := t.TempDir()
	files := map[string]string{
		".flow/project.md":           "# p\n",
		"web/src/pages/Home.tsx":     "export function Home() { return null }\n",
		"tests/home.spec.ts":         "test('home', () => {})\n",
		"tests/home.spec.ts.sources": "web/src/pages/Home.tsx\n",
		"tests/rooms.spec.ts":        "test('rooms', () => {})\n",
	}
	for p, body := range files {
		writeFile(t, filepath.Join(root, p), body)
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	for touch, want := range map[string]string{
		"tests/home.spec.ts":     "home.spec.ts\n",
		"web/src/pages/Home.tsx": "home.spec.ts\nrooms.spec.ts\n",
	} {
		writeFile(t, filepath.Join(root, touch), files[touch]+"// edited\n")
		r := runGuard("affected-specs", []string{root, "HEAD", filepath.Join(root, "tests")}, env)
		if r.rc != 0 || r.stdout != want {
			t.Errorf("%s: rc=%d stdout=%q err=%q, want %q", touch, r.rc, r.stdout, r.err, want)
		}
		writeFile(t, filepath.Join(root, touch), files[touch])
	}
}
