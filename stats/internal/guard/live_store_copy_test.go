package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// TestLiveStoreCopy ports scripts/test-live-store-copy.sh: live-store-copy
// copies only into flow_<name>_live, starts and stops only its own flowd,
// and refuses the protected ports and a malformed name before any side
// effect (kan-1041 self-review). docker, make, go, curl and flow are stubs
// on PATH; no database or daemon is touched. The cases share one name, so
// they run in order.
func TestLiveStoreCopy(t *testing.T) {
	stub := t.TempDir()
	log := stub + "/log"
	name := fmt.Sprintf("lsc-test%d", os.Getpid())
	alias := strings.ReplaceAll(name, "-", "_")
	db := "flow_" + alias + "_live"
	port := strconv.Itoa(20000 + os.Getpid()%20000)
	dir := "/tmp/flow-live-" + alias
	pidfile := "/tmp/flowd-" + port + ".pid"

	repo := t.TempDir()
	var g fxGit
	g.git("", "init", "-q", "-b", "main", repo)
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	if g.err != nil {
		t.Fatal(g.err)
	}
	mkdir(t, repo+"/stats") // up builds flowd from <repo>/stats
	sdd := repo + "/.superpowers/sdd/" + name
	t.Cleanup(func() {
		if pid, ok := lscPid(pidfile); ok && pid > 0 {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
		_ = os.RemoveAll(dir)
		_ = os.Remove(pidfile)
	})

	writeExec(t, stub+"/docker", "#!/usr/bin/env bash\necho \"docker $*\" >>\"$LOG\"\n")
	writeExec(t, stub+"/make", "#!/usr/bin/env bash\necho \"make $*\" >>\"$LOG\"\n")
	writeExec(t, stub+"/curl", "#!/usr/bin/env bash\n[ -e \"$STUB/up\" ]\n")
	writeExec(t, stub+"/go", `#!/usr/bin/env bash
echo "go $*" >>"$LOG"
out=$3
printf '#!/usr/bin/env bash\nprintf "%%s\\n%%s\\n" $$ "$0" >"$TMPDIR/flowd-$FLOWD_PORT.pid"\necho "$FLOWD_DSN" >>"$LOG"\necho "TRANSCRIPTS=$FLOW_TRANSCRIPTS_DIR" >>"$LOG"\ntouch "$STUB/up"\nexec sleep 300\n' >"$out"
chmod +x "$out"
`)
	writeExec(t, stub+"/flow", "#!/usr/bin/env bash\n{ echo \"flow $*\"; echo \"FLOW_ADDR=$FLOW_ADDR\"; cat; echo; } >>\"$LOG\"\n")

	vars := map[string]string{"STUB": stub, "LOG": log, "PATH": stub + ":" + os.Getenv("PATH"), "FLOW_GUARD_REPO_ROOT": repo}
	var out string
	lsc := func(extra map[string]string, args ...string) int {
		t.Helper()
		env := map[string]string{}
		for _, kv := range os.Environ() {
			k, v, _ := strings.Cut(kv, "=")
			env[k] = v
		}
		delete(env, "FLOW_LIVE_COPY_TRANSCRIPTS")
		for _, m := range []map[string]string{vars, extra} {
			for k, v := range m {
				env[k] = v
			}
		}
		var environ []string
		for k, v := range env {
			environ = append(environ, k+"="+v)
		}
		var b bytes.Buffer
		rc := liveStoreCopy(args, Env{Dir: repo, Getenv: func(k string) string { return env[k] }, Environ: func() []string { return environ }}, &b, &b)
		out = b.String()
		return rc
	}
	readLog := func() string { b, _ := os.ReadFile(log); return string(b) }
	resetLog := func() { writeFile(t, log, "") }
	hasLine := func(text, line string) bool {
		for _, l := range strings.Split(text, "\n") {
			if l == line {
				return true
			}
		}
		return false
	}
	alive := func(pid int) bool { return pid > 0 && syscall.Kill(pid, 0) == nil }
	sentinel := func() *exec.Cmd {
		c := exec.Command("sleep", "300")
		if err := c.Start(); err != nil {
			t.Fatal(err)
		}
		go func() { _ = c.Wait() }()
		return c
	}
	nothingRan := func(t *testing.T, rc int) {
		t.Helper()
		if rc != 2 {
			t.Errorf("exit %d, want 2: %s", rc, out)
		}
		if l := readLog(); l != "" {
			t.Errorf("ran %s", l)
		}
	}

	t.Run("an existing scratch change directory is refused before any side effect", func(t *testing.T) {
		mkdir(t, sdd)
		resetLog()
		nothingRan(t, lsc(nil, "up", name, port))
		if err := os.RemoveAll(sdd); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("a malformed name is refused before any side effect", func(t *testing.T) {
		for _, bad := range []string{"", "Flow", "../a", "a-b;rm", "-a"} {
			resetLog()
			nothingRan(t, lsc(nil, "up", bad, port))
		}
	})
	t.Run("a protected or invalid port is refused before any side effect", func(t *testing.T) {
		for _, p := range []string{"4173", "4174", "80", "x"} {
			resetLog()
			nothingRan(t, lsc(nil, "up", name, p))
		}
	})
	t.Run("restart with no up is refused before any side effect", func(t *testing.T) {
		resetLog()
		nothingRan(t, lsc(nil, "restart", name))
	})

	var pid int
	t.Run("up copies into the scratch database and starts its flowd on the copy", func(t *testing.T) {
		resetLog()
		if rc := lsc(nil, "up", name, port); rc != 0 {
			t.Fatalf("up: exit %d (%s)", rc, out)
		}
		l := readLog()
		for _, line := range []string{
			"docker exec -i flow-postgres createdb -U flow " + db,
			"docker exec -i flow-postgres pg_dump -U flow --no-owner --no-acl flow",
			"postgres://flow:flow@localhost:5433/" + db + "?sslmode=disable",
			"TRANSCRIPTS=" + dir + "/transcripts",
			"flow state set -C " + repo + " " + name,
			"FLOW_ADDR=http://127.0.0.1:" + port,
		} {
			if !hasLine(l, line) {
				t.Errorf("up: log lacks %q:\n%s", line, l)
			}
		}
		if !regexp.MustCompile(`(?m)psql -q -v ON_ERROR_STOP=1 -U flow -d ` + db + `$`).MatchString(l) {
			t.Errorf("up: no restore into %s:\n%s", db, l)
		}
		if !strings.Contains(l, `"state":"STARTED"`) {
			t.Errorf("up: state record is not STARTED:\n%s", l)
		}
		if !strings.Contains(out, "http://127.0.0.1:"+port) || !hasLine(out, "scratch change: "+name) {
			t.Errorf("up: address or scratch change not printed:\n%s", out)
		}
		pid, _ = lscPid(pidfile)
		if !alive(pid) {
			t.Error("up: scratch flowd not running")
		}
		if !isDir(sdd) {
			t.Errorf("up: no scratch change directory %s", sdd)
		}
		if rc := lsc(nil, "up", name, port); rc != 2 {
			t.Errorf("second up: exit %d, want 2", rc)
		}
	})
	t.Run("down by the _ alias stops only its own flowd and drops only its copy", func(t *testing.T) {
		s := sentinel()
		defer func() { _ = s.Process.Kill() }()
		resetLog()
		if rc := lsc(nil, "down", alias); rc != 0 {
			t.Fatalf("down: exit %d (%s)", rc, out)
		}
		if alive(pid) {
			t.Error("down: scratch flowd still running")
		}
		if !alive(s.Process.Pid) {
			t.Error("down: killed a process it did not start")
		}
		l := readLog()
		if !hasLine(l, "docker exec -i flow-postgres dropdb -U flow --if-exists --force "+db) {
			t.Errorf("down: no dropdb of %s:\n%s", db, l)
		}
		if regexp.MustCompile(`(?m)((dropdb|createdb).* flow$|-d flow$)`).MatchString(l) {
			t.Errorf("down: touched the flow database:\n%s", l)
		}
		if lscExists(dir) || lscExists(sdd) {
			t.Error("down: state dir or scratch change directory left")
		}
		if rc := lsc(nil, "down", name); rc != 0 {
			t.Errorf("second down: exit %d", rc)
		}
	})
	t.Run("FLOW_LIVE_COPY_TRANSCRIPTS points the scratch flowd at real transcripts", func(t *testing.T) {
		_ = os.Remove(stub + "/up")
		resetLog()
		if rc := lsc(map[string]string{"FLOW_LIVE_COPY_TRANSCRIPTS": stub + "/real"}, "up", name, port); rc != 0 {
			t.Fatalf("up with transcripts: exit %d (%s)", rc, out)
		}
		if !hasLine(readLog(), "TRANSCRIPTS="+stub+"/real") {
			t.Errorf("up: FLOW_LIVE_COPY_TRANSCRIPTS not passed as FLOW_TRANSCRIPTS_DIR:\n%s", readLog())
		}
	})
	var restarted int
	t.Run("restart stops only its own flowd and relaunches it with nothing copied, built or dropped", func(t *testing.T) {
		old, _ := lscPid(pidfile)
		s := sentinel()
		defer func() { _ = s.Process.Kill() }()
		_ = os.Remove(stub + "/up")
		resetLog()
		if rc := lsc(nil, "restart", name); rc != 0 {
			t.Fatalf("restart: exit %d (%s)", rc, out)
		}
		if alive(old) {
			t.Errorf("restart: old flowd %d still running", old)
		}
		restarted, _ = lscPid(pidfile)
		if restarted == old || !alive(restarted) {
			t.Error("restart: no new flowd running")
		}
		l := readLog()
		if regexp.MustCompile(`(?m)(^(docker|make|go) |createdb|pg_dump|dropdb)`).MatchString(l) {
			t.Errorf("restart: copied, built or dropped:\n%s", l)
		}
		if !hasLine(l, "TRANSCRIPTS="+stub+"/real") {
			t.Errorf("restart: recorded transcripts directory not passed:\n%s", l)
		}
		if !strings.Contains(out, "export FLOW_ADDR=http://127.0.0.1:"+port) {
			t.Errorf("restart: exports not printed:\n%s", out)
		}
		if !alive(s.Process.Pid) {
			t.Error("restart: killed a process it did not start")
		}
		if !isDir(sdd) {
			t.Error("restart: scratch change directory removed")
		}
	})
	t.Run("restart survives a dead previous flowd", func(t *testing.T) {
		if restarted <= 0 {
			t.Fatal("no flowd from the restart case to kill")
		}
		_ = syscall.Kill(restarted, syscall.SIGKILL)
		_ = os.Remove(stub + "/up")
		if rc := lsc(nil, "restart", name); rc != 0 {
			t.Fatalf("restart: dead flowd — exit %d (%s)", rc, out)
		}
		if p, _ := lscPid(pidfile); !alive(p) {
			t.Error("restart: dead flowd — no flowd running")
		}
		if rc := lsc(nil, "down", name); rc != 0 {
			t.Errorf("down after up with transcripts: exit %d", rc)
		}
	})
}

// lscPid is the pid on the first line of a flowd pidfile.
func lscPid(pidfile string) (int, bool) {
	b, err := os.ReadFile(pidfile)
	if err != nil {
		return 0, false
	}
	first, _, _ := strings.Cut(string(b), "\n")
	pid, err := strconv.Atoi(strings.TrimSpace(first))
	return pid, err == nil
}
