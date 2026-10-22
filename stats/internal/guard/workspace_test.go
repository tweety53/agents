package guard

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestWorkspace ports scripts/test-workspace.sh at e111a883: one subtest per
// ok: label, plus the refusals the harness never named. The refusal cases
// exit before any docker call and always run; the round-trip cases run
// against the real flow-postgres container and skip when it is unreachable.
func TestWorkspace(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	ws := func(args ...string) guardResult { return runGuard("workspace", args, env) }

	for _, c := range []struct {
		label string
		args  []string
		want  string
	}{
		{"create refuses an empty id", []string{"create", ""}, "ERROR: a workspace id is required (refusing to operate on the shared flow database)"},
		{"create refuses a missing id", []string{"create"}, "ERROR: a workspace id is required"},
		{"remove refuses an empty id", []string{"remove", ""}, "ERROR: a workspace id is required"},
		{"remove refuses a missing id", []string{"remove"}, "ERROR: a workspace id is required"},
		{"no command prints usage", nil, "usage: workspace.sh {create|remove|survivors} <id> [--force]\n       --force is accepted only by remove"},
		{"an id outside [a-z0-9-] is refused", []string{"create", "Bad"}, "ERROR: workspace id 'Bad' is not [a-z0-9-]; refusing to build a database name from it"},
		{"an unknown flag is refused", []string{"remove", "x", "--nope"}, "ERROR: unrecognized argument '--nope'\nusage: workspace.sh"},
		{"--force on create is refused", []string{"create", "x", "--force"}, "ERROR: unrecognized argument '--force'"},
		{"remove --force refuses a database not ending in _uitest", []string{"remove", "x", "--force"}, "ERROR: --force refused: database 'flow_x' does not end in _uitest"},
		{"an unknown command prints usage", []string{"frob", "x"}, "usage: workspace.sh"},
	} {
		t.Run(c.label, func(t *testing.T) {
			r := ws(c.args...)
			if r.rc != 2 || !strings.HasPrefix(r.err, c.want) || r.stdout != "" {
				t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
			}
		})
	}

	t.Run("round trip", func(t *testing.T) {
		if exec.Command("docker", "exec", "flow-postgres", "true").Run() != nil {
			t.Skip("SKIP: flow-postgres compose stack not reachable")
		}
		pid := strconv.Itoa(os.Getpid())
		id, id2 := "wstest-"+pid, "wstest-force-"+pid+"-uitest"
		db, db2 := "flow_wstest_"+pid, "flow_wstest_force_"+pid+"_uitest"
		psql := func(d, q string) *exec.Cmd {
			return exec.Command("docker", "exec", "flow-postgres", "psql", "-U", "flow", "-d", d, "-tAc", q)
		}
		exists := func(d string) bool {
			out, _ := psql("flow", "select 1 from pg_database where datname = '"+d+"'").Output()
			return strings.TrimSpace(string(out)) == "1"
		}
		t.Cleanup(func() {
			_ = exec.Command("docker", "exec", "flow-postgres", "dropdb", "-U", "flow", "--if-exists", db).Run()
			_ = exec.Command("docker", "exec", "flow-postgres", "dropdb", "-U", "flow", "--if-exists", "--force", db2).Run()
		})
		check := func(label string, ok bool, r guardResult) {
			t.Run(label, func(t *testing.T) {
				if !ok {
					t.Fatalf("rc=%d out=%q", r.rc, r.out)
				}
			})
		}

		r := ws("create", id)
		check("create exits 0", r.rc == 0, r)
		check("create makes "+db, exists(db), r)
		r = ws("create", id)
		check("create is a no-op on an existing database", r.rc == 0, r)
		r = ws("survivors", id)
		check("survivors exits 0 while the database survives", r.rc == 0, r)
		check("survivors prints "+db, r.stdout == db+"\n", r)
		r = ws("remove", id)
		check("remove exits 0", r.rc == 0, r)
		check("remove drops "+db, !exists(db), r)
		r = ws("survivors", id)
		check("survivors exits 0 with nothing surviving", r.rc == 0, r)
		check("survivors prints nothing after remove", r.stdout == "", r)
		for _, args := range [][]string{{"create", ""}, {"create"}, {"remove", ""}, {"remove"}} {
			r = ws(args...)
			check(strings.Join(args, " ")+" refuses an empty or missing id", r.rc != 0, r)
		}
		check("the shared flow database is untouched", exists("flow"), r)

		ws("create", id2)
		hold := psql(db2, "select pg_sleep(30)")
		if err := hold.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = hold.Process.Kill(); _ = hold.Wait() }()
		held := false
		for i := 0; i < 20 && !held; i++ {
			out, _ := psql("flow", "select count(*) from pg_stat_activity where datname = '"+db2+"' and pid <> pg_backend_pid()").Output()
			n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
			held = n > 0
			if !held {
				time.Sleep(500 * time.Millisecond)
			}
		}
		check("a live connection is held open on "+db2, held, r)
		r = ws("remove", id2)
		check("remove (no flag) fails against a database with a live connection", r.rc != 0, r)
		check(db2+" still exists after the unforced remove failed", exists(db2), r)
		r = ws("remove", id2, "--force")
		check("remove --force exits 0 against a database with a live connection", r.rc == 0, r)
		check("remove --force drops "+db2, !exists(db2), r)

		ws("create", id)
		r = ws("remove", id, "--force")
		check("remove --force refuses a database not ending in _uitest", r.rc != 0, r)
		check(db+" survives the refused --force", exists(db), r)
	})
}
