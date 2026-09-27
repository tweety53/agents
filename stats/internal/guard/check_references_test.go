package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-references.sh at 3e48ecac, one subtest per
// ok: label. The bash harness ran the guard with CHECK_REFERENCES_ROOT on a
// sandboxed fixture and asserted on its exit and its combined output; here
// the guard runs in-process with the same override on Env, over a fixture in
// t.TempDir() carrying the harness's rules/ and skills/demo/ directories.

// crEnv is an Env whose variables are vars alone, plus the process's PATH
// and locale for the `sort` the guard runs; a name absent from vars is unset.
func crEnv(dir string, vars map[string]string) Env {
	lookup := func(k string) (string, bool) {
		if v, ok := vars[k]; ok {
			return v, true
		}
		switch k {
		case "PATH", "LANG", "LC_ALL", "LC_COLLATE", "LC_CTYPE":
			return os.LookupEnv(k)
		}
		return "", false
	}
	return Env{Dir: dir, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
}

// crFixture is new_fixture plus files (path relative to the fixture: body).
func crFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	mkdir(t, root+"/rules")
	mkdir(t, root+"/skills/demo")
	for rel, body := range files {
		writeFile(t, filepath.Join(root, rel), body)
	}
	return root
}

func crRun(env Env) (int, string) {
	var out bytes.Buffer
	rc := checkReferences(nil, env, &out, &out)
	return rc, out.String()
}

func TestCheckReferences(t *testing.T) {
	t.Parallel()
	const (
		ntp      = "rules/never-touch-production.mdc"
		demo     = "skills/demo/SKILL.md"
		wrr      = "skills/flow-contracts/worktree-resolution-rationale.md"
		refState = "Resolve it per **State file** in `rules/never-touch-production.mdc`.\n"
	)
	variants := "see **State file** in `rules/never-touch-production.mdc`\n" +
		"per **State file** in `rules/never-touch-production.mdc`\n" +
		"defined once under **State file** in `rules/never-touch-production.mdc`\n"
	outside := t.TempDir()
	writeFile(t, outside+"/target.md", "## Something else\n\nbody\n")
	nowhere := filepath.Join(t.TempDir(), "refs-nonexistent")

	type row struct {
		label  string
		files  map[string]string
		env    func(t *testing.T, root string) Env // nil: CHECK_REFERENCES_ROOT=<fixture>
		rc     int                                 // -1: any non-zero
		has    []string
		hasNot []string
	}
	rows := []row{
		{label: "live reference passes", files: map[string]string{ntp: "## State file\n\nbody\n", demo: refState}},
		{label: "moved section fails", files: map[string]string{ntp: "## Something else\n\nbody\n", demo: refState}, rc: -1},
		{label: "reports file:line", files: map[string]string{ntp: "## Something else\n\nbody\n", demo: refState}, rc: -1,
			has: []string{"skills/demo/SKILL.md:1"}},
		{label: "phrasing variants pass", files: map[string]string{ntp: "## State file\n\nbody\n", demo: variants}},
		{label: "phrasing variants all fail when absent", files: map[string]string{ntp: "## Gone\n\nbody\n", demo: variants}, rc: -1,
			has: []string{"SKILL.md:1", "SKILL.md:2", "SKILL.md:3"}},
		{label: "multi-section line passes", files: map[string]string{
			ntp:  "## Stage transitions\n\nbody\n\n## State file\n\nbody\n",
			demo: "Follow `rules/never-touch-production.mdc` — sections **Stage transitions**, **State file**.\n"}},
		{label: "allow marker suppresses", files: map[string]string{ntp: "## Gone\n\nbody\n",
			wrr: "**Do not** copy from `rules/never-touch-production.mdc` <!-- refs-guard:allow -->\n"}},
		{label: "fenced block skipped", files: map[string]string{ntp: "## Gone\n\nbody\n",
			wrr: "```\nsee **State file** in `rules/never-touch-production.mdc`\n```\n"}},
		{label: "unresolvable path skipped", files: map[string]string{
			wrr: "see **Whatever** in `spectre/changes/<name>/tasks.md`\n"}},
		{label: "relative path resolves", files: map[string]string{wrr: "## Panel re-runs\n\nbody\n",
			demo: "Follow **Panel re-runs** in `../flow-contracts/worktree-resolution-rationale.md`.\n"}},
		{label: "a stale reference through a relative path still fails", files: map[string]string{wrr: "## Something else\n\nbody\n",
			demo: "Follow **Panel re-runs** in `../flow-contracts/worktree-resolution-rationale.md`.\n"}, rc: -1},
		{label: "heading normalization", files: map[string]string{ntp: "## The `widget` setting\n\nbody\n",
			demo: "see **The widget setting** in `rules/never-touch-production.mdc`\n"}},
		// The bash case ran the guard with no override from inside a stale
		// fixture and relied on this repository being clean. Here the root the
		// shim hands over (FLOW_GUARD_REPO_ROOT) is a clean fixture, so the
		// case proves the same thing — cwd is never scanned — without reading
		// this repository's own tree.
		{label: "no-override run ignores cwd, scans the real repo instead",
			files: map[string]string{ntp: "## Gone\n\nbody\n", demo: refState},
			env: func(t *testing.T, root string) Env {
				clean := crFixture(t, map[string]string{ntp: "## State file\n\nbody\n", demo: refState})
				return crEnv(root, map[string]string{"FLOW_GUARD_REPO_ROOT": clean})
			}},
		{label: "in-code ** does not desync a live reference", files: map[string]string{ntp: "## State file\n\nbody\n",
			demo: "See **State file** and code `x**y` in `rules/never-touch-production.mdc`.\n"}},
		{label: "in-code ** does not mask a genuinely stale reference", files: map[string]string{ntp: "## Something else\n\nbody\n",
			demo: "See **Gone** and code `x**y` in `rules/never-touch-production.mdc`.\n"}, rc: -1},
		{label: "soft-wrapped bold span still tokenizes correctly", files: map[string]string{ntp: "## Jira integration\n\nbody\n",
			demo: "sync** in **Jira integration** (`rules/never-touch-production.mdc`) — canonical.\n"}},
		{label: "a live reference to an #-titled file passes", files: map[string]string{ntp: "# State file\n\n## Body notes\n\nbody\n", demo: refState}},
		{label: "a stale reference to an #-titled file still fails", files: map[string]string{ntp: "# Something else\n\nbody\n", demo: refState}, rc: -1},
		{label: "a fenced # comment in the referenced file does not satisfy a reference", files: map[string]string{
			ntp: "# Something else\n\n```bash\n# State file\necho hi\n```\n", demo: refState}, rc: -1},
		{label: "a real #-title still resolves alongside a fenced comment of the same text", files: map[string]string{
			ntp: "# State file\n\n## Body notes\n\n```bash\n# State file\necho hi\n```\n", demo: refState}},
		{label: "unassociated emphasis does not demand a heading", files: map[string]string{ntp: "## State file\n\nbody\n",
			wrr: "**Never** commit during apply. The contract lives in `rules/never-touch-production.mdc`.\n" +
				"Run it **after** step 2 — see the note in `rules/never-touch-production.mdc`.\n"}},
		{label: "an associated token is still checked beside emphasis", files: map[string]string{ntp: "## Something else\n\nbody\n",
			demo: "**Never** skip this — see **State file** in `rules/never-touch-production.mdc`.\n"}, rc: -1},
		{label: "a token is assigned to its nearest path only", files: map[string]string{
			ntp: "## Apps\n\nbody\n", "rules/context7.mdc": "## Panel re-runs\n\nbody\n",
			demo: "the apps in `rules/never-touch-production.mdc` (see **Panel re-runs** in `rules/context7.mdc`)\n"}},
		{label: "an absolute path outside the root fails the run", files: map[string]string{
			demo: "see **State file** in `" + outside + "/target.md`\n"}, rc: -1},
		{label: "the refusal is reported", files: map[string]string{
			demo: "see **State file** in `" + outside + "/target.md`\n"}, rc: -1, has: []string{"resolves outside the repository root"}},
		{label: "an escaping path whose target does not exist fails identically", files: map[string]string{
			demo: "see **State file** in `" + nowhere + "/target.md`\n"}, rc: -1},
		{label: "the refusal is reported for a nonexistent target", files: map[string]string{
			demo: "see **State file** in `" + nowhere + "/target.md`\n"}, rc: -1, has: []string{"resolves outside the repository root"}},
		{label: "a traversing relative path fails the run", files: map[string]string{
			demo: "see **State file** in `../../refs-escape-nowhere/target.md`\n"}, rc: -1},
		{label: "an empty CHECK_REFERENCES_ROOT exits 2", rc: 2,
			env: func(t *testing.T, root string) Env {
				return crEnv(root, map[string]string{"CHECK_REFERENCES_ROOT": ""})
			}},
		{label: "a nonexistent CHECK_REFERENCES_ROOT exits 2", rc: 2,
			env: func(t *testing.T, root string) Env {
				return crEnv(root, map[string]string{"CHECK_REFERENCES_ROOT": root + "/refs-does-not-exist"})
			}},
		{label: "a root with no Markdown exits 2 instead of reporting clean", rc: 2,
			env: func(t *testing.T, root string) Env {
				for _, d := range []string{"rules", "skills"} {
					if err := os.RemoveAll(filepath.Join(root, d)); err != nil {
						t.Fatal(err)
					}
				}
				return crEnv(root, map[string]string{"CHECK_REFERENCES_ROOT": root})
			}},
		{label: "coverage: a file with a live reference reports rc=0", files: map[string]string{ntp: "## State file\n\nbody\n", demo: refState}},
		{label: "coverage: the breakdown names the file with its checked count", files: map[string]string{ntp: "## State file\n\nbody\n", demo: refState},
			has: []string{"skills/demo/SKILL.md 1"}},
		{label: "coverage: an undeclared zero-coverage file fails", files: map[string]string{
			demo: "Just prose. No cross-reference syntax anywhere in this file.\n"}, rc: -1},
		{label: "coverage: names the file, the zero count, and that it is undeclared", files: map[string]string{
			demo: "Just prose. No cross-reference syntax anywhere in this file.\n"}, rc: -1,
			has: []string{"skills/demo/SKILL.md:0: 0 checked, and not declared expected-zero (coverage)"}},
		{label: "coverage: a declared expected-zero member passes", files: map[string]string{
			ntp: "Never access a production system directly. No file cross-references here.\n"}},
		{label: "coverage: the breakdown marks the declared zero", files: map[string]string{
			ntp: "Never access a production system directly. No file cross-references here.\n"},
			has: []string{"rules/never-touch-production.mdc 0 (declared"}},
		{label: "a prefixed citation is still checked", files: map[string]string{"agents-repo-fixture.md": "## Something else\n\nbody\n",
			demo: "Resolve it per **Some section** in `<agents repo>/agents-repo-fixture.md`.\n"}, rc: -1},
		{label: "the prefixed reference is reported by file:line", files: map[string]string{"agents-repo-fixture.md": "## Something else\n\nbody\n",
			demo: "Resolve it per **Some section** in `<agents repo>/agents-repo-fixture.md`.\n"}, rc: -1, has: []string{"skills/demo/SKILL.md:1"}},
		{label: "a project-prefixed citation is neither resolved nor refused", files: map[string]string{
			".flow/project.md": "## Something else\n\nbody\n", wrr: "see **Whatever** in `<project>/.flow/project.md`\n"}},
		{label: "project prefix: not read as a containment escape", files: map[string]string{
			".flow/project.md": "## Something else\n\nbody\n", wrr: "see **Whatever** in `<project>/.flow/project.md`\n"},
			hasNot: []string{"resolves outside the repository root"}},
		{label: "an unreadable file refuses with exit 2, naming the read and never coverage_record", files: map[string]string{
			ntp: "## State file\n\nbody\n", demo: refState},
			env: func(t *testing.T, root string) Env {
				if os.Geteuid() == 0 {
					t.Skip("chmod 000 does not block reads as root")
				}
				p := filepath.Join(root, demo)
				if err := os.Chmod(p, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(p, 0o644) })
				return crEnv(t.TempDir(), map[string]string{"CHECK_REFERENCES_ROOT": root})
			},
			rc: 2, has: []string{"check-references: cannot read skills/demo/SKILL.md: "}, hasNot: []string{"coverage_record failed"}},
	}
	for _, r := range rows {
		t.Run(r.label, func(t *testing.T) {
			t.Parallel()
			root := crFixture(t, r.files)
			env := crEnv(t.TempDir(), map[string]string{"CHECK_REFERENCES_ROOT": root})
			if r.env != nil {
				env = r.env(t, root)
			}
			rc, out := crRun(env)
			if (r.rc == -1 && rc == 0) || (r.rc != -1 && rc != r.rc) {
				t.Fatalf("rc = %d, want %d (-1: non-zero); out:\n%s", rc, r.rc, out)
			}
			for _, s := range r.has {
				if !strings.Contains(out, s) {
					t.Errorf("output lacks %q; out:\n%s", s, out)
				}
			}
			for _, s := range r.hasNot {
				if strings.Contains(out, s) {
					t.Errorf("output carries %q; out:\n%s", s, out)
				}
			}
		})
	}
}
