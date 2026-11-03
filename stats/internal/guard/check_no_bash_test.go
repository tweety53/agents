package guard

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckNoBash(t *testing.T) {
	t.Parallel()
	// run builds a git repository holding tracked (path → body) and one
	// untracked y.sh, then runs check-no-bash over it.
	run := func(t *testing.T, tracked map[string]string) (int, string) {
		t.Helper()
		root := t.TempDir()
		gitRun(t, root, "init", "-q")
		for p, body := range tracked {
			writeFile(t, filepath.Join(root, p), body)
			gitRun(t, root, "add", "--", p)
		}
		writeFile(t, filepath.Join(root, "y.sh"), "#!/usr/bin/env bash\n")
		return nbRun(root, root)
	}
	for _, c := range []struct {
		label, path, body string
		rc                int
	}{
		{"a tracked .sh fails", "x.sh", "echo\n", 1},
		{"an extensionless env-bash script fails", "tool", "#!/usr/bin/env bash\necho\n", 1},
		{"a /bin/sh script fails", "tool", "#!/bin/sh\necho\n", 1},
		{"an env -S bash script fails", "tool", "#!/usr/bin/env -S bash -e\necho\n", 1},
		{"an env -u FOO bash script fails", "tool", "#!/usr/bin/env -u FOO bash\necho\n", 1},
		{"an env -C dir -P path bash script fails", "tool", "#!/usr/bin/env -C /tmp -P /bin bash\necho\n", 1},
		{"an env -Sbash script fails", "tool", "#!/usr/bin/env -Sbash -e\necho\n", 1},
		{"an env -S python3 script passes", "tool", "#!/usr/bin/env -S python3 -u\nprint()\n", 0},
		{"an env-python3 script passes", "tool", "#!/usr/bin/env python3\nprint()\n", 0},
		{"setup.sh passes", "setup.sh", "#!/usr/bin/env bash\n", 0},
		{"scripts/flow-guard passes", "scripts/flow-guard", "#!/usr/bin/env bash\n", 0},
		{"a .sh under spectre/changes/archive/ passes", "spectre/changes/archive/a/x.sh", "#!/bin/sh\n", 0},
	} {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			rc, out := run(t, map[string]string{"README.md": "x\n", c.path: c.body})
			if rc != c.rc {
				t.Fatalf("exit %d, want %d\n%s", rc, c.rc, out)
			}
			if c.rc == 1 && !strings.Contains(out, c.path+": bash script — write it in Go in flow-guard (rules/no-bash-scripts.mdc)\n") {
				t.Fatalf("no finding naming %s:\n%s", c.path, out)
			}
			if c.rc == 0 && (!strings.Contains(out, "NO-BASH-OK") || strings.Contains(out, "y.sh")) {
				t.Fatalf("want NO-BASH-OK and the untracked y.sh ignored:\n%s", out)
			}
		})
	}
	t.Run("outside a git worktree it cannot answer", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if rc, out := nbRun(dir, dir); rc != 2 {
			t.Fatalf("exit %d, want 2\n%s", rc, out)
		}
	})
}

// nbRun runs check-no-bash with CHECK_NO_BASH_ROOT=root from dir.
func nbRun(root, dir string) (int, string) {
	env := Env{Dir: dir, Getenv: func(k string) string {
		if k == "CHECK_NO_BASH_ROOT" {
			return root
		}
		return ""
	}}
	var out, errb bytes.Buffer
	rc := Registry["check-no-bash"](nil, env, &out, &errb)
	return rc, out.String() + errb.String()
}
