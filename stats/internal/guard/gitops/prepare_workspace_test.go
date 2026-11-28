package gitops

// The retired scripts/test-prepare-workspace.sh, ported case for case: each
// subtest is named after that harness's `ok:` label, followed by the refusal
// paths the harness never exercised. Every fixture is a git repository under
// t.TempDir() checked out on `spectre/<name>` — the branch every apply
// worktree is created on, and the one place prepare-workspace reads the
// change name from. The isolation validation runs in-process, so the
// violation text relayed in case 3 is the validator's own.
//
// Case 2's change name is kan-15-parallel-flow-task-lanes, the worked example
// under "The workspace id" (skills/flow-contracts/workspace-isolation.md): id
// kan-15-fb13, id_underscored kan_15_fb13, offset 2760 — so the expected
// values are the contract's, not this guard's own arithmetic.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

type pwResult struct {
	rc       int
	out, err string
}

// pwLoopRun runs prepare-workspace on a worktree checked out on the loop branch
// [branch], and answers rc 99 when its output differs from a spectre/<branch>
// worktree's — the whole branch is the change name.
func pwLoopRun(t *testing.T, branch string) pwResult {
	t.Helper()
	reference := pwRun(t, guardtest.ProjectRepo(t, branch, pwPortSection("8080")))
	repo := guardtest.ProjectRepo(t, branch, pwPortSection("8080"))
	guardtest.GitRun(t, repo, "checkout", "-q", "-b", branch)
	r := pwRun(t, repo)
	if r.out != reference.out {
		r.rc = 99
	}
	return r
}

func pwRun(t *testing.T, args ...string) pwResult {
	t.Helper()
	var out, errb bytes.Buffer
	env := core.Env{Dir: t.TempDir(), Getenv: os.Getenv}
	rc := prepareWorkspace(args, env, &out, &errb)
	return pwResult{rc, out.String(), errb.String()}
}

func pwPortSection(def string) string {
	return "## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|----------|----------|---------|----------------|\n" +
		"| `port` | `API_PORT` | `" + def + "` | `+<offset>` |\n" + guardtest.ProjectCommands
}

const pwDeclared = "## workspace isolation\n\n" +
	"| Resource | Variable | Default | In a workspace |\n" +
	"|----------|----------|---------|----------------|\n" +
	"| `database` | `DB_URL` | `jdbc:postgresql://localhost:5432/appdb` | `jdbc:postgresql://localhost:5432/appdb_<id_underscored>` |\n" +
	"| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<id>` |\n" +
	"| `port` | `API_PORT` | `8080` | `+<offset>` |\n" +
	"| `url` | `MEDIA_BASE_URL` | `http://localhost:9000/appdb-media` | `http://localhost:9000/<value:MEDIA_BUCKET>` |\n" +
	"| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value:API_PORT>` |\n" +
	"| `cache index` | `CACHE_INDEX` | `0` | `probed` |\n" + guardtest.ProjectCommands

func TestPrepareWorkspace(t *testing.T) {
	smoke.Heavy(t)
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
			return pwRun(t, guardtest.ProjectRepo(t, "kan-1-no-isolation", "# fixture\n\n## test\n\n```bash\nscripts/test-setup.sh\n```\n"))
		}, []check{
			{"1a: no isolation section exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"1a: no isolation section prints nothing", func(r pwResult) bool { return r.out == "" }},
		}},
		{"1b", func(t *testing.T) pwResult {
			return pwRun(t, guardtest.ProjectRepo(t, "kan-2-no-config", ""))
		}, []check{
			{"1b: no .flow/project.md exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"1b: no .flow/project.md prints nothing", func(r pwResult) bool { return r.out == "" }},
		}},
		{"2", func(t *testing.T) pwResult {
			return pwRun(t, guardtest.ProjectRepo(t, "kan-15-parallel-flow-task-lanes", pwDeclared))
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
			return pwRun(t, guardtest.ProjectRepo(t, "kan-3-malformed", pwPortSection("notanumber")))
		}, []check{
			{"3: a malformed row is a non-zero exit", func(r pwResult) bool { return r.rc == 1 }},
			{"3: the guard's own violation is relayed", func(r pwResult) bool {
				i := strings.Index(r.out, "API_PORT")
				return i >= 0 && strings.Contains(r.out[i:], "not a bare integer")
			}},
			{"3: no KEY=value line reached stdout", func(r pwResult) bool { return !strings.Contains(r.out, "=") }},
		}},
		{"4a", func(t *testing.T) pwResult {
			return pwRun(t, guardtest.ProjectRepo(t, "kan-4-leading-zero-port", pwPortSection("0070")))
		}, []check{
			{"4a: a leading-zero port default exits 0", func(r pwResult) bool { return r.rc == 0 }},
			{"4a: API_PORT is exported", func(r pwResult) bool { return strings.Contains(r.out, "API_PORT=") }},
			{"4a: 0070 is not read as octal", func(r pwResult) bool { return !strings.Contains(r.out, "API_PORT=76") }},
		}},
		{"4b", func(t *testing.T) pwResult {
			return pwRun(t, guardtest.ProjectRepo(t, "kan-5-invalid-octal-port", pwPortSection("0080")))
		}, []check{
			{"4b: an invalid-octal port default does not crash", func(r pwResult) bool { return r.rc == 0 }},
		}},
		// The refusals the harness never reached, one case each.
		{"usage", func(t *testing.T) pwResult { return pwRun(t) }, []check{
			{"no worktree argument is exit 2 with the usage line", func(r pwResult) bool {
				return r.rc == 2 && r.err == "Usage: flow-guard prepare-workspace <worktree>\n" && r.out == ""
			}},
		}},
		{"notdir", func(t *testing.T) pwResult {
			return pwRun(t, filepath.Join(t.TempDir(), "absent"))
		}, []check{
			{"a worktree that is not a directory is exit 2", func(r pwResult) bool {
				return r.rc == 2 && strings.HasSuffix(r.err, "/absent is not a directory\n") && strings.HasPrefix(r.err, "prepare-workspace: ")
			}},
		}},
		{"emptyarg", func(t *testing.T) pwResult { return pwRun(t, "") }, []check{
			{"an empty worktree argument is not a directory, exit 2", func(r pwResult) bool {
				return r.rc == 2 && r.err == "prepare-workspace:  is not a directory\n" && r.out == ""
			}},
		}},
		// KAN-997: a self-review or pipeline-fix loop worktree is its own change, named
		// after its whole branch (scripts/pipeline-fix.sh), so its id derives from that
		// name — the same output a spectre/<that name> worktree gets.
		{"self-review", func(t *testing.T) pwResult {
			return pwLoopRun(t, "self-review-kan-1-x")
		}, []check{
			{"self-review branch derives the id from the branch", func(r pwResult) bool {
				return r.rc == 0 && strings.Contains(r.out, "API_PORT=")
			}},
		}},
		{"pipeline-fix", func(t *testing.T) pwResult {
			return pwLoopRun(t, "pipeline-fix-kan-1-x-2")
		}, []check{
			{"pipeline-fix branch derives the id from the branch", func(r pwResult) bool {
				return r.rc == 0 && strings.Contains(r.out, "API_PORT=")
			}},
		}},
		// A parallel executor's worktree, branch <name>-r<n>, is its own workspace:
		// the whole branch is the id's name, so -r1 and -r2 never share one.
		{"executor-r1", func(t *testing.T) pwResult {
			return pwLoopRun(t, "kan-1-x-r1")
		}, []check{
			{"<name>-r1 branch derives the id from the branch", func(r pwResult) bool {
				return r.rc == 0 && r.out != pwLoopRun(t, "kan-1-x-r2").out
			}},
		}},
		{"other", func(t *testing.T) pwResult {
			repo := guardtest.ProjectRepo(t, "kan-12-other", pwPortSection("8080"))
			// A loop prefix anywhere but the branch's start is no loop branch.
			guardtest.GitRun(t, repo, "checkout", "-q", "-b", "feature/self-review-kan-12")
			return pwRun(t, repo)
		}, []check{
			{"other branch still refused", func(r pwResult) bool {
				return r.rc == 2 && r.out == "" && strings.Contains(r.err, "(got 'feature/self-review-kan-12')")
			}},
		}},
		{"detached", func(t *testing.T) pwResult {
			repo := guardtest.ProjectRepo(t, "kan-10-detached", pwPortSection("8080"))
			guardtest.GitRun(t, repo, "checkout", "-q", "--detach")
			return pwRun(t, repo)
		}, []check{
			{"a detached worktree is exit 2 naming detached HEAD", func(r pwResult) bool {
				return r.rc == 2 && r.out == "" && strings.Contains(r.err, "is not on a spectre/<name>, self-review-<name>, pipeline-fix-<name>-<k> or <name>-r<n> branch (got 'detached HEAD') — cannot derive a workspace id without the change name\n")
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

// TestPwChangeNameExecutorBranches: a session branch <name> names no workspace,
// and each <name>-r<n> executor branch names its own, whole.
func TestPwChangeNameExecutorBranches(t *testing.T) {
	t.Parallel()
	for branch, want := range map[string]bool{"kan-1-x": false, "kan-1-x-r1": true, "kan-1-x-r2": true, "kan-1-x-r": false, "kan-1-x-rx": false} {
		name, ok := ChangeName(branch)
		if ok != want || (ok && name != branch) {
			t.Errorf("pwChangeName(%q) = %q, %v; want %v", branch, name, ok, want)
		}
	}
}
