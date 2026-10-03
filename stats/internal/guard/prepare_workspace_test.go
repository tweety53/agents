package guard

// The retired scripts/test-prepare-workspace.sh, ported case for case: each
// subtest is named after that harness's `ok:` label, followed by the refusal
// paths the harness never exercised. Every fixture is a git repository under
// t.TempDir() checked out on `spectre/<name>` — the branch every apply
// worktree is created on, and the one place prepare-workspace reads the
// change name from — beside a scripts/ directory whose
// check-workspace-isolation.sh FLOW_GUARD_WORKSPACE_ISOLATION points the guard
// at, the way the shim's own export does.
//
// The sibling check-workspace-isolation.sh is the REAL guard wherever the
// harness ran the real one: a two-line wrapper exec'ing the flow-guard binary
// this package's tests already build (guardBinary), so the validation, the
// `#ROW` lines and the violation text relayed in case 3 are the producer's
// own, never hand-built. Cases 5 and 6 replace it, as the harness did: a
// non-executable copy, and a stub reporting rows a real guard refuses.
//
// Case 2's change name is kan-15-parallel-flow-task-lanes, the worked example
// under "The workspace id" (skills/flow-contracts/workspace-isolation.md): id
// kan-15-fb13, id_underscored kan_15_fb13, offset 2760 — so the expected
// values are the contract's, not this guard's own arithmetic.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type pwResult struct {
	rc       int
	out, err string
}

// pwRepo is a fresh repository on spectre/<name>, with project.md committed
// when cfg is non-empty. Its path is named after the harness's own sandbox,
// never t.TempDir()'s: that one carries the running subtest's name, and the
// fixture runs under whichever label gets there first — case 3's label holds
// an `=`, which the relayed violation then carried onto stdout and failed
// that case's own "no KEY=value line" check.
func pwRepo(t *testing.T, name, cfg string) string {
	t.Helper()
	repo, err := os.MkdirTemp("", "prepare-workspace-test.")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(repo) })
	gitRun(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "README.md"), "seed\n")
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-q", "-m", "seed")
	gitRun(t, repo, "checkout", "-q", "-b", "spectre/"+name)
	if cfg != "" {
		writeFile(t, filepath.Join(repo, ".flow", "project.md"), cfg+"\n")
		gitRun(t, repo, "add", "-A")
		gitRun(t, repo, "commit", "-q", "-m", "configure")
	}
	return repo
}

// pwRealSibling is a repository root whose scripts/check-workspace-isolation.sh
// runs the real guard.
func pwRealSibling(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeExec(t, filepath.Join(root, "scripts", "check-workspace-isolation.sh"),
		"#!/bin/sh\nFLOW_GUARD_TELEMETRY=off exec '"+guardBinary(t)+"' check-workspace-isolation \"$@\"\n")
	return root
}

func pwRun(t *testing.T, root string, args ...string) pwResult {
	t.Helper()
	var out, errb bytes.Buffer
	env := Env{Dir: t.TempDir(), Getenv: func(k string) string {
		if k == "FLOW_GUARD_WORKSPACE_ISOLATION" {
			if root == "" {
				return ""
			}
			return filepath.Join(root, "scripts", "check-workspace-isolation.sh")
		}
		return os.Getenv(k)
	}}
	rc := prepareWorkspace(args, env, &out, &errb)
	return pwResult{rc, out.String(), errb.String()}
}

const pwCommands = `
| Command | Runs |
|---------|------|
| ` + "`create`" + ` | ` + "`./scripts/workspace create`" + ` |
| ` + "`remove`" + ` | ` + "`./scripts/workspace remove`" + ` |
| ` + "`survivors`" + ` | ` + "`./scripts/workspace survivors`" + ` |
`

func pwPortSection(def string) string {
	return "## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|----------|----------|---------|----------------|\n" +
		"| `port` | `API_PORT` | `" + def + "` | `+<offset>` |\n" + pwCommands
}

const pwDeclared = "## workspace isolation\n\n" +
	"| Resource | Variable | Default | In a workspace |\n" +
	"|----------|----------|---------|----------------|\n" +
	"| `database` | `DB_URL` | `jdbc:postgresql://localhost:5432/appdb` | `jdbc:postgresql://localhost:5432/appdb_<id_underscored>` |\n" +
	"| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<id>` |\n" +
	"| `port` | `API_PORT` | `8080` | `+<offset>` |\n" +
	"| `url` | `MEDIA_BASE_URL` | `http://localhost:9000/appdb-media` | `http://localhost:9000/<value:MEDIA_BUCKET>` |\n" +
	"| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value:API_PORT>` |\n" +
	"| `cache index` | `CACHE_INDEX` | `0` | `probed` |\n" + pwCommands

func TestPrepareWorkspace(t *testing.T) {
	t.Parallel()

	type check struct {
		label string
		ok    func(r pwResult) bool
	}
	cases := []struct {
		name   string // the harness's case, for the reader
		run    func(t *testing.T) pwResult
		checks []check
	}{
		{"1a", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), pwRepo(t, "kan-1-no-isolation", "# fixture\n\n## test\n\n```bash\nscripts/test-setup.sh\n```\n"))
		}, []check{
			{"1a: no isolation section exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"1a: no isolation section prints nothing", func(r pwResult) bool { return r.out == "" }},
		}},
		{"1b", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), pwRepo(t, "kan-2-no-config", ""))
		}, []check{
			{"1b: no .flow/project.md exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"1b: no .flow/project.md prints nothing", func(r pwResult) bool { return r.out == "" }},
		}},
		{"2", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), pwRepo(t, "kan-15-parallel-flow-task-lanes", pwDeclared))
		}, []check{
			{"2: a declared section exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"2: database row substitutes <id_underscored>", func(r pwResult) bool {
				return strings.Contains(r.out, "DB_URL=jdbc:postgresql://localhost:5432/appdb_kan_15_fb13")
			}},
			{"2: bucket row substitutes <id>", func(r pwResult) bool { return strings.Contains(r.out, "MEDIA_BUCKET=appdb-media-kan-15-fb13") }},
			{"2: port row adds the offset (8080 + 2760)", func(r pwResult) bool { return strings.Contains(r.out, "API_PORT=10840") }},
			{"2: url row resolves <value:MEDIA_BUCKET>", func(r pwResult) bool {
				return strings.Contains(r.out, "MEDIA_BASE_URL=http://localhost:9000/appdb-media-kan-15-fb13")
			}},
			{"2: url row resolves <value:API_PORT>", func(r pwResult) bool { return strings.Contains(r.out, "WEB_URL=http://localhost:10840") }},
			// The whole stdout, pinned byte for byte: declaration order, one
			// line per exported row.
			{"2: exactly one KEY=value line per declared row", func(r pwResult) bool {
				return r.out == "DB_URL=jdbc:postgresql://localhost:5432/appdb_kan_15_fb13\n"+
					"MEDIA_BUCKET=appdb-media-kan-15-fb13\nAPI_PORT=10840\n"+
					"MEDIA_BASE_URL=http://localhost:9000/appdb-media-kan-15-fb13\nWEB_URL=http://localhost:10840\n"
			}},
			{"2: the cache index row is not among the exported KEY=value lines", func(r pwResult) bool { return !strings.Contains(r.out, "CACHE_INDEX") }},
			{"2: the cache index row is reported by name on stderr", func(r pwResult) bool {
				return strings.Contains(r.err, "prepare-workspace: `CACHE_INDEX` (cache index) is claimed by probing the project's own cache, not derived from the workspace id — per the registry in skills/flow-contracts/artifacts-registry.md, `/flow` claims it, by probing, when it exports the workspace's variables. This script does not carry a client for the project's cache, so this row is reported rather than exported; claim it against the real service before anything reads `CACHE_INDEX`.\n")
			}},
		}},
		{"3", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), pwRepo(t, "kan-3-malformed", pwPortSection("notanumber")))
		}, []check{
			{"3: a malformed row is a non-zero exit", func(r pwResult) bool { return r.rc == 1 }},
			{"3: the guard's own violation is relayed", func(r pwResult) bool {
				i := strings.Index(r.out, "API_PORT")
				return i >= 0 && strings.Contains(r.out[i:], "not a bare integer")
			}},
			{"3: no KEY=value line reached stdout", func(r pwResult) bool { return !strings.Contains(r.out, "=") }},
		}},
		{"4a", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), pwRepo(t, "kan-4-leading-zero-port", pwPortSection("0070")))
		}, []check{
			{"4a: a leading-zero port default exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"4a: API_PORT is exported", func(r pwResult) bool { return strings.Contains(r.out, "API_PORT=") }},
			{"4a: 0070 is not read as octal", func(r pwResult) bool { return !strings.Contains(r.out, "API_PORT=76") }},
		}},
		{"4b", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), pwRepo(t, "kan-5-invalid-octal-port", pwPortSection("0080")))
		}, []check{
			{"4b: an invalid-octal port default does not crash", func(r pwResult) bool { return r.rc == 0 }},
		}},
		{"5", func(t *testing.T) pwResult {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "scripts", "check-workspace-isolation.sh"), "#!/bin/sh\nexit 0\n")
			return pwRun(t, root, pwRepo(t, "kan-6-guard-not-executable", pwPortSection("8080")))
		}, []check{
			{"5: a non-executable guard exits 2", func(r pwResult) bool { return r.rc == 2 }},
			{"5: a clear message names the problem", func(r pwResult) bool {
				return strings.Contains(r.err, "cannot find a runnable check-workspace-isolation.sh")
			}},
		}},
		{"6", func(t *testing.T) pwResult {
			root := t.TempDir()
			writeExec(t, filepath.Join(root, "scripts", "check-workspace-isolation.sh"), "#!/bin/sh\n"+
				"printf 'ISOLATION-OK: stub — 2 resource row(s) and 3 command row(s) validated\\n'\n"+
				"printf '#ROW\\tcache index\\tCACHE_INDEX\\t0\\tprobed\\n'\n"+
				"printf '#ROW\\turl\\tWEB_URL\\thttp://localhost/0\\thttp://localhost/<value:CACHE_INDEX>\\n'\n")
			return pwRun(t, root, pwRepo(t, "kan-7-value-ref-wrong-kind",
				"## workspace isolation\n\n(fixture only — the stub guard below supplies the rows this script reads)\n"))
		}, []check{
			{"6: a url row referencing a cache index row is a non-zero exit", func(r pwResult) bool { return r.rc == 2 }},
			{"6: no substituted value reached stdout", func(r pwResult) bool { return !strings.Contains(r.out, "CACHE_INDEX") }},
			{"6: the wrong-kind reference is reported by name on stderr", func(r pwResult) bool {
				return strings.Contains(r.err, "prepare-workspace: <value:CACHE_INDEX> resolves to `CACHE_INDEX`, a `cache index` row — a reference may only name a `database`, `bucket` or `port` row")
			}},
		}},
		// The refusals the harness never reached, one case each.
		{"usage", func(t *testing.T) pwResult { return pwRun(t, pwRealSibling(t)) }, []check{
			{"no worktree argument is exit 2 with the usage line", func(r pwResult) bool {
				return r.rc == 2 && r.err == "Usage: prepare-workspace.sh <worktree>\n" && r.out == ""
			}},
		}},
		{"notdir", func(t *testing.T) pwResult {
			return pwRun(t, pwRealSibling(t), filepath.Join(t.TempDir(), "absent"))
		}, []check{
			{"a worktree that is not a directory is exit 2", func(r pwResult) bool {
				return r.rc == 2 && strings.HasSuffix(r.err, "/absent is not a directory\n") && strings.HasPrefix(r.err, "prepare-workspace: ")
			}},
		}},
		{"emptyarg", func(t *testing.T) pwResult { return pwRun(t, pwRealSibling(t), "") }, []check{
			{"an empty worktree argument is not a directory, exit 2", func(r pwResult) bool {
				return r.rc == 2 && r.err == "prepare-workspace:  is not a directory\n" && r.out == ""
			}},
		}},
		{"noroot", func(t *testing.T) pwResult { return pwRun(t, "", pwRepo(t, "kan-8-no-root", "")) }, []check{
			{"FLOW_GUARD_WORKSPACE_ISOLATION unset is exit 2", func(r pwResult) bool {
				return r.rc == 2 && r.err == "prepare-workspace: FLOW_GUARD_WORKSPACE_ISOLATION is unset — run scripts/prepare-workspace.sh, which sets it\n"
			}},
		}},
		{"missing", func(t *testing.T) pwResult { return pwRun(t, t.TempDir(), pwRepo(t, "kan-9-missing", "")) }, []check{
			{"a missing sibling guard is exit 2 naming the directory", func(r pwResult) bool {
				return r.rc == 2 && strings.HasPrefix(r.err, "prepare-workspace: cannot find a runnable check-workspace-isolation.sh in ") &&
					strings.HasSuffix(r.err, "/scripts\n")
			}},
		}},
		{"badport", func(t *testing.T) pwResult {
			root := t.TempDir()
			writeExec(t, filepath.Join(root, "scripts", "check-workspace-isolation.sh"), "#!/bin/sh\n"+
				"printf '#ROW\\tport\\tAPI_PORT\\tx80\\t+<offset>\\n'\n")
			return pwRun(t, root, pwRepo(t, "kan-11-bad-port", "## workspace isolation\n"))
		}, []check{
			{"a port default that is not an integer is exit 1 and prints nothing", func(r pwResult) bool {
				return r.rc == 1 && r.out == "" && r.err == "prepare-workspace: `API_PORT` (port) default \"x80\" is not a base-10 integer\n"
			}},
		}},
		{"detached", func(t *testing.T) pwResult {
			repo := pwRepo(t, "kan-10-detached", pwPortSection("8080"))
			gitRun(t, repo, "checkout", "-q", "--detach")
			return pwRun(t, pwRealSibling(t), repo)
		}, []check{
			{"a detached worktree is exit 2 naming detached HEAD", func(r pwResult) bool {
				return r.rc == 2 && r.out == "" && strings.Contains(r.err, "is not on a spectre/<name> branch (got 'detached HEAD') — cannot derive a workspace id without the change name\n")
			}},
		}},
	}
	// One subtest per label, so parity is a `--- PASS` count; each fixture
	// runs once, in whichever of its labels gets there first.
	for _, c := range cases {
		var once sync.Once
		var r pwResult
		for _, ch := range c.checks {
			t.Run(ch.label, func(t *testing.T) {
				t.Parallel()
				once.Do(func() { r = c.run(t) })
				if !ch.ok(r) {
					t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", r.rc, r.out, r.err)
				}
			})
		}
	}
}

// The real shim, linked into a temporary directory not named scripts/ with
// its sibling and lib/ linked beside it, execs the sibling from beside
// itself, as the bash's $SCRIPT_DIR did — never from <parent>/scripts.
func TestPrepareWorkspaceShimSibling(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "bin")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"prepare-workspace.sh", "check-workspace-isolation.sh", "lib"} {
		symlink(t, tcfScriptsDir(t)+"/"+n, dir+"/"+n)
	}
	cmd := exec.Command("/bin/bash", dir+"/prepare-workspace.sh", pwRepo(t, "kan-12-shim-sibling", ""))
	cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	_ = cmd.Run()
	if rc := cmd.ProcessState.ExitCode(); rc != 0 {
		t.Errorf("shim beside its sibling outside scripts/: rc=%d stdout=%q stderr=%q", rc, out.String(), errb.String())
	}
}
