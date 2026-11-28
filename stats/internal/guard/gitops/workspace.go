package gitops

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/workspace.sh shim (KAN-1058).
//
// Usage:
//   flow-guard workspace create <id>
//   flow-guard workspace remove <id> [--force]

// workspace creates and removes this repository's per-workspace database,
// the `create`/`remove` pair named in .flow/project.md's `## workspace
// isolation` section per skills/flow-contracts/project-configuration.md
// (canonical for what each command runs).
//
// Usage:
//
//	flow-guard workspace create <id>
//	flow-guard workspace remove <id> [--force]
//
// `remove`'s `--force` is opt-in and additive: omitted, `remove` runs exactly
// as it always has (`dropdb --if-exists`, no --force). Passed, it adds
// Postgres' WITH (FORCE) to the drop, which succeeds against a database with
// a live connection this process did not start. It changes nothing about id
// validation or the refusal to operate on the shared "flow" database below.
//
// `--force` also refuses any id whose database name does not end in
// "_uitest" -- the same suffix rule cmd/uitest-seed/guard.go enforces on its
// own destructive path. This is defence in depth, not the only guard: the
// caller (stats/Makefile's ui-test-up) is still responsible for passing the
// right id, since this command has no way to know a caller's intent -- only
// a second, independent check on the one attribute that distinguishes the
// UI-test database from every other workspace database it can remove. A
// future caller that legitimately needs a forced drop of a non-uitest
// workspace would need this rule revisited, not silently bypassed.
//
// The database service is the flow-postgres container (superuser flow,
// database flow, host port 5433). Every subcommand reaches it through
// `docker exec`, since this machine has no host-installed psql/createdb/
// dropdb client -- one mechanism for both commands.
//
// A workspace database is named flow_<id_underscored> (every "-" in the id
// replaced by "_"), per skills/flow-contracts/workspace-isolation.md's
// "What the id derives". <id> is validated against that section's own
// output alphabet ([a-z0-9-]) before it is ever interpolated into SQL or a
// database name.
//
// `create` is a no-op when the database already exists, so a re-run of the
// pipeline in an existing worktree is not an error. `create` and `remove`
// both refuse an empty or missing id rather than operating on the shared
// "flow" database.
func init() { core.Registry["workspace"] = workspace }

const (
	wsContainer = "flow-postgres"
	wsSuperuser = "flow"
	wsUsage     = "usage: flow-guard workspace {create|remove} <id> [--force]\n" +
		"       --force is accepted only by remove\n"
)

func workspace(args []string, env core.Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	fail := func(id, msg string, rc int) int {
		core.Reason(env, id)
		fmt.Fprintln(stderr, msg)
		return rc
	}
	cmd, id, forceFlag := arg(0), arg(1), arg(2)
	if cmd == "" {
		fmt.Fprint(stderr, wsUsage)
		core.Reason(env, "usage")
		return 2
	}
	if id == "" {
		return fail("id-required", "ERROR: a workspace id is required (refusing to operate on the shared "+wsSuperuser+" database)", 2)
	}
	if strings.Trim(id, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" {
		return fail("bad-id", "ERROR: workspace id '"+id+"' is not [a-z0-9-]; refusing to build a database name from it", 2)
	}
	force := false
	if forceFlag != "" {
		if cmd != "remove" || forceFlag != "--force" {
			fmt.Fprintf(stderr, "ERROR: unrecognized argument '%s'\n%s", forceFlag, wsUsage)
			core.Reason(env, "unrecognized-argument")
			return 2
		}
		force = true
	}
	db := "flow_" + strings.ReplaceAll(id, "-", "_")
	if force && !strings.HasSuffix(db, "_uitest") {
		return fail("force-refused", "ERROR: --force refused: database '"+db+"' does not end in _uitest", 2)
	}

	docker := "docker"
	if p, ok := core.LookPath(env, "docker"); ok {
		docker = p
	}
	run := func(out io.Writer, dockerArgs ...string) int {
		c := exec.Command(docker, append([]string{"exec"}, dockerArgs...)...)
		c.Stdout, c.Stderr = out, stderr
		return wsExitCode(c.Run())
	}
	switch cmd {
	case "create":
		var out bytes.Buffer
		if rc := run(&out, wsContainer, "psql", "-U", wsSuperuser, "-d", wsSuperuser, "-tAc",
			"select 1 from pg_database where datname = '"+db+"'"); rc != 0 {
			return fail("container-unreachable", "ERROR: could not reach "+wsContainer+" to check for "+db, rc)
		}
		if strings.TrimRight(out.String(), "\n") == "1" {
			return 0
		}
		if rc := run(stdout, wsContainer, "createdb", "-U", wsSuperuser, db); rc != 0 {
			return fail("createdb-failed", "ERROR: createdb failed for "+db, rc)
		}
		fmt.Fprintln(stdout, "created "+db)
	case "remove":
		if force {
			if rc := run(stdout, wsContainer, "dropdb", "-U", wsSuperuser, "--if-exists", "--force", db); rc != 0 {
				return fail("dropdb-failed", "ERROR: dropdb --force failed for "+db, rc)
			}
		} else if rc := run(stdout, wsContainer, "dropdb", "-U", wsSuperuser, "--if-exists", db); rc != 0 {
			return fail("dropdb-failed", "ERROR: dropdb failed for "+db, rc)
		}
	default:
		fmt.Fprint(stderr, wsUsage)
		core.Reason(env, "usage")
		return 2
	}
	return 0
}

// wsExitCode is the exit status bash would report for a command that ran
// with err: the child's own code, 128+n when a signal killed it, or 127 when
// docker could not be started.
func wsExitCode(err error) int {
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &exitErr):
		return core.WaitExitCode(exitErr.ProcessState)
	default:
		return 127
	}
}
