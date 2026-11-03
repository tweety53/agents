package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// liveStoreCopy is `flow-guard live-store-copy up <name> [<port>] | restart
// <name> | down <name>` — a design's `## Live check` against real data without
// touching the dev stack (kan-1041 self-review: the check was hand-built
// across about 15 calls).
//
//	up    copies the dev `flow` database, read-only on the source, into a
//	      scratch database flow_<name, - as _>_live (pg_dump | psql inside the
//	      flow-postgres container), builds flowd from this checkout and starts
//	      it on <port> (default 4175) against the copy, then prints the address
//	      and the FLOW_ADDR / FLOW_RECORDS_ADDR exports that point a session at it.
//	      It registers a scratch change <name> in the copy -- a STARTED state
//	      record and .superpowers/sdd/<name>/ in this checkout, refused when
//	      that directory already exists -- so `flow record -change <name>` works.
//	      A `-` in <name> becomes `_` in the database name and the state
//	      directory alike, so `a-b` and `a_b` name one copy, never two sharing
//	      a database (kan-1067 self-review, 1-E).
//	      The scratch flowd reads transcripts from FLOW_LIVE_COPY_TRANSCRIPTS
//	      when set (e.g. "$HOME/.claude/projects"), else from an empty scratch
//	      directory: a copy that is to exercise harvest or the back-fill reads
//	      the real transcripts, read-only, while every write stays in the
//	      scratch database.
//	restart  stops that scratch flowd as `down` does and launches it again as
//	      `up` does, on the port, binary and transcripts directory `up`
//	      recorded -- no copy, no build, no drop -- then prints the address
//	      and the exports. Exit 2 with no prior `up`, 1 when flowd does not
//	      answer within 60 s (kan-1051 self-review, F17).
//	down  stops that scratch flowd — only the pid its own pidfile names, and
//	      only when that pidfile names this script's binary — and drops
//	      flow_<name, - as _>_live. A second `down` is a no-op. It removes the
//	      scratch change directory `up` created, and no other.
//
// Never the protected dev stack: port 4173 (the dev flowd) and 4174 (the
// UI-test stack) are refused, and the only database dropped is
// flow_<name, - as _>_live. State lives in /tmp/flow-live-<name, - as _>/; the daemon writes
// its own pidfile at /tmp/flowd-<port>.pid (TMPDIR pinned, as ui-test-up does).
//
// Exit codes: 0 done, 1 a step failed (docker, copy, build, not answering),
// 2 usage or a refusal.
//
// "This checkout" is FLOW_GUARD_REPO_ROOT, the checkout scripts/flow-guard
// runs from. docker, make, go, curl, git and flow are found over the
// caller's PATH; a failing docker, make or go step exits with that step's
// own status, as the bash's `set -e` did.
func init() { Registry["live-store-copy"] = liveStoreCopy }

var lscName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// lscExit unwinds liveStoreCopy with its exit code: the bash's `exit`.
type lscExit int

type lsc struct {
	env            Env
	stdout, stderr io.Writer
	environ        []string
	port, dir, db  string
}

func liveStoreCopy(args []string, env Env, stdout, stderr io.Writer) (code int) {
	l := &lsc{env: env, stdout: stdout, stderr: stderr, environ: icChildEnv(env)}
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(lscExit)
			if !ok {
				panic(r)
			}
			code = int(e)
		}
	}()
	if len(args) < 2 {
		l.usage()
	}
	cmd, name := args[0], args[1]
	l.port = "4175"
	if len(args) > 2 {
		l.port = args[2]
	}
	if !lscName.MatchString(name) {
		l.die(2, "name '%s' is not [a-z0-9][a-z0-9_-]*", name)
	}
	l.db = "flow_" + strings.ReplaceAll(name, "-", "_") + "_live"
	l.dir = "/tmp/flow-live-" + strings.ReplaceAll(name, "-", "_")
	repo := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if repo == "" {
		l.die(2, "FLOW_GUARD_REPO_ROOT is unset — run flow-guard live-store-copy, which sets it")
	}
	sdd := repo + "/.superpowers/sdd/" + name
	switch cmd {
	case "up":
		if len(args) > 3 {
			l.usage()
		}
		l.portOK()
		if lscExists(l.dir) {
			l.die(2, "%s exists -- run down %s first", l.dir, name)
		}
		if lscExists(sdd) {
			l.die(2, "%s exists -- pick another name", sdd)
		}
		if l.answers() {
			l.die(2, "port %s already answers", l.port)
		}
		probe := l.pg("psql", "-U", "flow", "-d", "flow", "-tAc", "select 1 from pg_database where datname='"+l.db+"'")
		if out, _ := probe.Output(); strings.TrimRight(string(out), "\n") != "" {
			l.die(2, "database %s exists -- run down %s first", l.db, name)
		}
		for _, d := range []string{l.dir + "/transcripts", l.dir + "/state", sdd} {
			l.check(os.MkdirAll(d, 0o777))
		}
		transcripts := env.Getenv("FLOW_LIVE_COPY_TRANSCRIPTS")
		if transcripts == "" {
			transcripts = l.dir + "/transcripts"
		}
		for f, v := range map[string]string{"scratch": sdd, "port": l.port, "transcripts-dir": transcripts} {
			l.check(os.WriteFile(l.dir+"/"+f, []byte(v+"\n"), 0o666))
		}
		l.step(l.pg("createdb", "-U", "flow", l.db))
		if !l.copyStore() {
			l.die(1, "copy into %s failed", l.db)
		}
		for _, c := range [][]string{{"make", "-s", "web-dist"}, {"go", "build", "-o", l.dir + "/flowd", "./cmd/flowd"}} {
			build := l.cmd(c[0], c[1:]...)
			build.Dir = repo + "/stats"
			l.step(build)
		}
		l.launch(l.read("transcripts-dir"))
		l.printAddr()
		branch, _ := l.cmd("git", "-C", repo, "branch", "--show-current").Output()
		head, _ := l.cmd("git", "-C", repo, "rev-parse", "HEAD").Output()
		body := fmt.Sprintf(`{"state":"STARTED","branch":"%s","worktrees":{"%s":"%s"},"artifactUrl":null,"jiraIssue":null,"prUrl":null,"updatedBy":"/flow"}`,
			strings.TrimRight(string(branch), "\n"), repo, strings.TrimRight(string(head), "\n"))
		set := l.cmd("flow", "state", "set", "-C", repo, name)
		set.Env = append(set.Env, "FLOW_ADDR=http://127.0.0.1:"+l.port)
		set.Stdin = strings.NewReader(body)
		set.Stdout, set.Stderr = stdout, stderr
		if set.Run() != nil {
			l.die(1, "state record for %s failed", name)
		}
		fmt.Fprintf(stdout, "scratch change: %s\n", name)
	case "restart":
		if len(args) != 2 {
			l.usage()
		}
		for _, f := range []string{"port", "flowd", "transcripts-dir"} {
			if !lscExists(l.dir + "/" + f) {
				l.die(2, "no %s/%s -- run up %s first", l.dir, f, name)
			}
		}
		l.port = l.read("port")
		l.portOK()
		l.stopFlowd()
		l.launch(l.read("transcripts-dir"))
		l.printAddr()
	case "down":
		if len(args) != 2 {
			l.usage()
		}
		if isFile(l.dir + "/port") {
			l.port = l.read("port")
			l.portOK()
			l.stopFlowd()
		}
		l.step(l.pg("dropdb", "-U", "flow", "--if-exists", "--force", l.db))
		if isFile(l.dir + "/scratch") {
			l.check(os.RemoveAll(l.read("scratch")))
		}
		l.check(os.RemoveAll(l.dir))
	default:
		l.usage()
	}
	return 0
}

func (l *lsc) usage() {
	fmt.Fprintln(l.stderr, "usage: flow-guard live-store-copy up <name> [<port>] | restart <name> | down <name>")
	panic(lscExit(2))
}

func (l *lsc) die(code int, format string, a ...any) {
	fmt.Fprintf(l.stderr, "live-store-copy: "+format+"\n", a...)
	panic(lscExit(code))
}

// check is set -e over a file system step: its error ends the run with 1.
func (l *lsc) check(err error) {
	if err != nil {
		l.die(1, "%v", err)
	}
}

// cmd is name over the caller's PATH, run in the caller's environment.
func (l *lsc) cmd(name string, args ...string) *exec.Cmd {
	path, ok := lookPath(l.env, name)
	if !ok {
		path = name
	}
	cmd := exec.Command(path, args...)
	cmd.Dir = l.env.Dir
	cmd.Env = append([]string{}, l.environ...)
	return cmd
}

// pg is `docker exec -i flow-postgres <args>`.
func (l *lsc) pg(args ...string) *exec.Cmd {
	return l.cmd("docker", append([]string{"exec", "-i", "flow-postgres"}, args...)...)
}

// step runs cmd with its output passed through, ending the run with its own
// exit status when it fails.
func (l *lsc) step(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stderr = l.stdout, l.stderr
	if rc := exitCode(cmd.Run()); rc != 0 {
		panic(lscExit(rc))
	}
}

// copyStore is `pg pg_dump … flow | pg psql … -d <db> >/dev/null` under
// pipefail: true only when both ends succeed.
func (l *lsc) copyStore() bool {
	dump := l.pg("pg_dump", "-U", "flow", "--no-owner", "--no-acl", "flow")
	restore := l.pg("psql", "-q", "-v", "ON_ERROR_STOP=1", "-U", "flow", "-d", l.db)
	// An os.Pipe the parent closes once both ends have started, never
	// StdoutPipe: that keeps a read end open in the guard until dump.Wait, so
	// a psql that exits early leaves pg_dump blocked on a full pipe forever
	// instead of dying of SIGPIPE as the bash pipeline's did.
	r, w, err := os.Pipe()
	if err != nil {
		return false
	}
	// Both ends write the one stderr, as the bash pipeline's shared fd did;
	// a writer that is not a file is copied into by a goroutine per end.
	shared := &lscLockedWriter{w: l.stderr}
	dump.Stdout, dump.Stderr, restore.Stdin, restore.Stderr = w, shared, r, shared
	dumpStarted, restoreStarted := dump.Start() == nil, restore.Start() == nil
	_, _ = r.Close(), w.Close()
	var dumpErr, restoreErr error = errLscNotStarted, errLscNotStarted
	if dumpStarted {
		dumpErr = dump.Wait()
	}
	if restoreStarted {
		restoreErr = restore.Wait()
	}
	return restoreErr == nil && dumpErr == nil
}

var errLscNotStarted = errors.New("not started")

// lscLockedWriter serialises the writes of the processes that share it.
type lscLockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (lw *lscLockedWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}

// read is the trimmed content of the state file f, `$(cat "$dir/f")`.
func (l *lsc) read(f string) string {
	b, err := os.ReadFile(l.dir + "/" + f)
	l.check(err)
	return strings.TrimRight(string(b), "\n")
}

func (l *lsc) portOK() {
	n, err := strconv.Atoi(l.port)
	if err != nil || strings.TrimLeft(l.port, "0123456789") != "" || n < 1024 || n > 65535 || n == 4173 || n == 4174 {
		l.die(2, "port '%s' refused (1024-65535, never 4173 or 4174)", l.port)
	}
}

// answers is `curl -sf http://127.0.0.1:<port>/api/v1/changes`.
func (l *lsc) answers() bool {
	return l.cmd("curl", "-sf", "http://127.0.0.1:"+l.port+"/api/v1/changes").Run() == nil
}

// stopFlowd stops the scratch flowd on the port -- only the pid its own
// pidfile names, and only when that pidfile names this guard's binary.
func (l *lsc) stopFlowd() {
	b, err := os.ReadFile("/tmp/flowd-" + l.port + ".pid")
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 || !lscSameFile(lines[1], l.dir+"/flowd") {
		return
	}
	// A pid of 0 or below would signal a process group, never one daemon.
	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil || pid <= 0 {
		return
	}
	alive := func() bool { return syscall.Kill(pid, 0) == nil }
	_ = syscall.Kill(pid, syscall.SIGTERM)
	for i := 0; i < 30 && alive(); i++ {
		time.Sleep(time.Second)
	}
	if alive() {
		l.die(1, "flowd %d did not exit within 30s", pid)
	}
}

// launch starts the scratch flowd on the port against the copy, detached
// as nohup did, and returns once it answers; it ends the run with 1 when it
// does not within 60 s.
func (l *lsc) launch(transcripts string) {
	log, err := os.OpenFile(l.dir+"/log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	l.check(err)
	defer log.Close()
	cmd := exec.Command(l.dir + "/flowd")
	cmd.Dir = l.env.Dir
	for _, kv := range l.environ {
		switch k, _, _ := strings.Cut(kv, "="); k {
		case "FLOWD_JIRA_SITE", "FLOWD_JIRA_EMAIL", "FLOWD_JIRA_TOKEN":
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, "FLOWD_PORT="+l.port,
		"FLOWD_DSN=postgres://flow:flow@localhost:5433/"+l.db+"?sslmode=disable",
		"FLOW_TRANSCRIPTS_DIR="+transcripts, "FLOW_STATE_DIR="+l.dir+"/state", "TMPDIR=/tmp")
	cmd.Stdout, cmd.Stderr = log, log
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	l.check(cmd.Start())
	go func() { _ = cmd.Wait() }()
	for i := 0; i < 60; i++ {
		if l.answers() {
			return
		}
		time.Sleep(time.Second)
	}
	l.die(1, "flowd did not answer on :%s within 60s -- see %s/log", l.port, l.dir)
}

func (l *lsc) printAddr() {
	fmt.Fprintf(l.stdout, "http://127.0.0.1:%s\n", l.port)
	fmt.Fprintf(l.stdout, "export FLOW_ADDR=http://127.0.0.1:%s FLOW_RECORDS_ADDR=http://127.0.0.1:%s\n", l.port, l.port)
}

// lscExists is `[ -e ]`.
func lscExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// lscSameFile is `[ a -ef b ]`.
func lscSameFile(a, b string) bool {
	fa, err := os.Stat(filepath.Clean(a))
	if err != nil {
		return false
	}
	fb, err := os.Stat(b)
	return err == nil && os.SameFile(fa, fb)
}
