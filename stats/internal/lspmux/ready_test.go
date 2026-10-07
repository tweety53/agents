package lspmux

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// progressScript is a fake-server script that opens progress token "tok",
// holds it until release exists, then ends it.
func progressScript(t *testing.T, dir string) (script, release string) {
	t.Helper()
	script, release = dir+"/script.json", dir+"/release"
	body := `[
		{"jsonrpc":"2.0","id":"c1","method":"window/workDoneProgress/create","params":{"token":"tok"}},
		{"jsonrpc":"2.0","method":"$/progress","params":{"token":"tok","value":{"kind":"begin","title":"Indexing"}}},
		{"wait":"` + release + `"},
		{"jsonrpc":"2.0","method":"$/progress","params":{"token":"tok","value":{"kind":"end"}}}
	]`
	if err := os.WriteFile(script, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return script, release
}

// awaitAnswer reads until the answer to id, returning the messages before it.
func (c *testClient) awaitAnswer(id json.RawMessage) (message, []message) {
	c.t.Helper()
	var other []message
	for {
		m := c.next()
		if m.Method == "" && string(m.ID) == string(id) {
			return m, other
		}
		other = append(other, m)
	}
}

func TestReadinessWaitsForProgressEnd(t *testing.T) {
	t.Parallel()

	t.Run("progress begin…end, then the settle window", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		log := f.dir + "/fake.log"
		script, release := progressScript(t, f.dir)
		settle := 200 * time.Millisecond
		c := startMux(t, &Mux{Server: fakeServer(t, "log="+log, "script="+script), Settle: settle})
		c.call("initialize", initParams(f.main))
		c.notify("textDocument/didOpen", didOpen(f.main+"/a.go"))
		id := c.request("textDocument/hover", textDoc(f.main+"/a.go"))
		waitLog(t, log, func(rs []fakeRecord) bool {
			return slices.ContainsFunc(rs, func(r fakeRecord) bool { return r.Kind == "sent" && strings.Contains(string(r.Msg), `"begin"`) })
		})
		start := time.Now()
		if err := os.WriteFile(release, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		resp, other := c.awaitAnswer(id)
		if d := time.Since(start); d < settle {
			t.Errorf("answered %v after the progress could end, want at least the %v settle window", d, settle)
		}
		if got := resultRoot(t, resp); got != f.main {
			t.Errorf("answered by %s, want %s", got, f.main)
		}
		if len(other) != 0 {
			t.Errorf("the client got %+v; $/progress for tokens it never created must not reach it", other)
		}
		end, hover := -1, -1
		for i, r := range readLog(t, log) {
			m := r.message()
			switch {
			case r.Kind == "sent" && strings.Contains(string(r.Msg), `"end"`):
				end = i
			case r.Kind == "recv" && m.Method == "textDocument/hover":
				hover = i
			case r.Kind == "recv" && m.Method == "initialize":
				var p struct {
					Capabilities struct {
						Window struct{ WorkDoneProgress bool } `json:"window"`
					} `json:"capabilities"`
				}
				_ = json.Unmarshal(m.Params, &p)
				if !p.Capabilities.Window.WorkDoneProgress {
					t.Errorf("the server's initialize does not ask for progress: %s", m.Params)
				}
			}
		}
		if end < 0 || hover < end {
			t.Errorf("hover reached the server at record %d, progress ended at %d: want after the end\n%s", hover, end, mustRead(log))
		}
	})

	t.Run("no progress: ready after the settle window", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		settle := 300 * time.Millisecond
		c := startMux(t, &Mux{Server: fakeServer(t, "log="+f.dir+"/fake.log"), Settle: settle})
		c.call("initialize", initParams(f.main))
		start := time.Now()
		resp, _ := c.call("textDocument/hover", textDoc(f.main+"/a.go"))
		if d := time.Since(start); d < settle {
			t.Errorf("answered after %v, want at least the %v settle window", d, settle)
		}
		if got := resultRoot(t, resp); got != f.main {
			t.Errorf("answered by %s, want %s", got, f.main)
		}
	})
}

func TestIndexLogLine(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	path := f.dir + "/cache/worktree-lsp/index.log"
	server := fakeServer(t, "log="+f.dir+"/fake.log")
	settle := 50 * time.Millisecond
	c := startMux(t, &Mux{Server: server, Settle: settle, IndexLog: path})
	c.call("initialize", initParams(f.main))
	c.call("textDocument/hover", textDoc(f.main+"/a.go"))
	c.call("textDocument/hover", textDoc(f.main+"/a.go"))
	c.close()
	lines := strings.Split(strings.TrimRight(mustRead(path), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("index log holds %q, want one ready line", lines)
	}
	m := regexp.MustCompile(`^(\S+) (\S+) (\S+) ready (\d+)$`).FindStringSubmatch(lines[0])
	if m == nil {
		t.Fatalf("index log line %q, want `<RFC3339> <server> <root> ready <ms>`", lines[0])
	}
	if _, err := time.Parse(time.RFC3339, m[1]); err != nil {
		t.Errorf("timestamp %q: %v", m[1], err)
	}
	if m[2] != filepath.Base(server[0]) || m[3] != f.main {
		t.Errorf("server %q root %q, want %q %q", m[2], m[3], filepath.Base(server[0]), f.main)
	}
	if ms, _ := strconv.Atoi(m[4]); ms < int(settle.Milliseconds()) {
		t.Errorf("ready after %dms, want at least the %v settle window", ms, settle)
	}
}

func TestWorkspaceSymbolFanOut(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	exe := fakeServer(t)[0]
	log := f.dir + "/fake.log"
	// One command for every root, as Mux.Server is; the worktree's server
	// fails workspace/symbol.
	server := []string{"/bin/sh", "-c", `if [ "$(pwd -P)" = "$1" ]; then exec "$2" fake-lsp log="$3" fail=workspace/symbol; fi; exec "$2" fake-lsp log="$3"`, "sh", f.wt, exe, log}
	c := startMux(t, &Mux{Server: server})
	c.call("initialize", initParams(f.session))
	for _, p := range []string{f.main + "/a.go", f.wt + "/a.go", f.loose + "/a.go"} {
		c.notify("textDocument/didOpen", didOpen(p))
	}
	resp, _ := c.call("workspace/symbol", map[string]string{"query": "x"})
	var res []map[string]string
	if err := json.Unmarshal(resp.Result, &res); err != nil || resp.Error != nil {
		t.Fatalf("workspace/symbol answered %s (error %s)", resp.Result, resp.Error)
	}
	var roots []string
	for _, r := range res {
		roots = append(roots, r["root"])
	}
	slices.Sort(roots)
	if want := []string{f.main, f.session}; !slices.Equal(roots, want) {
		t.Errorf("results from %v, want %v — every running server's, the failing one's dropped", roots, want)
	}
	n := 0
	for _, r := range readLog(t, log) {
		if r.message().Method == "workspace/symbol" {
			n++
		}
	}
	if n != 3 {
		t.Errorf("workspace/symbol reached %d servers, want 3", n)
	}
}

func TestCallHierarchyRoutesByItem(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+f.dir+"/fake.log")})
	c.call("initialize", initParams(f.main))
	for _, tc := range []struct{ method, file, root string }{
		{"callHierarchy/incomingCalls", f.wt + "/a.go", f.wt},
		{"callHierarchy/outgoingCalls", f.wt + "/a.go", f.wt},
		{"callHierarchy/incomingCalls", f.main + "/a.go", f.main},
	} {
		item := map[string]any{"item": map[string]any{"name": "F", "kind": 12, "uri": fileURI(tc.file)}}
		resp, _ := c.call(tc.method, item)
		if got := resultRoot(t, resp); got != tc.root {
			t.Errorf("%s on %s answered by %s, want %s", tc.method, tc.file, got, tc.root)
		}
	}
}

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

// waitGone polls until pid is gone — it is not this process's child, so
// there is nothing to wait on.
func waitGone(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for alive(pid) {
		if time.Now().After(deadline) {
			t.Fatalf("process %d still running", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRemovedRootStopsChild(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	log := f.dir + "/fake.log"
	m := &Mux{Server: fakeServer(t, "log="+log), WatchInterval: 20 * time.Millisecond}
	c := startMux(t, m)
	c.call("initialize", initParams(f.main))
	c.call("textDocument/hover", textDoc(f.main+"/a.go"))
	c.notify("textDocument/didOpen", didOpen(f.wt+"/a.go"))
	c.call("textDocument/hover", textDoc(f.wt+"/a.go"))
	pids := map[string]int{}
	for _, r := range readLog(t, log) {
		pids[r.Cwd] = r.Pid
	}
	if err := os.RemoveAll(f.wt); err != nil {
		t.Fatal(err)
	}
	waitGone(t, pids[f.wt])
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(10 * time.Millisecond) {
		m.mu.Lock()
		_, held := m.docs[f.wt]
		m.mu.Unlock()
		if !held {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the removed root's open documents are still remembered")
		}
	}
	if !alive(pids[f.main]) {
		t.Error("the server of a root that still exists was stopped")
	}
	resp, _ := c.call("textDocument/hover", textDoc(f.main+"/a.go"))
	if got := resultRoot(t, resp); got != f.main {
		t.Errorf("hover after the removal answered by %s, want %s", got, f.main)
	}
}

// startWrapperChildren starts a process named worktree-lsp, cwd parentDir,
// with one `sleep` child per dir in dirs, and returns the children's pids.
func startWrapperChildren(t *testing.T, parentDir string, dirs ...string) []int {
	t.Helper()
	script := ""
	for i := range dirs {
		script += `(cd "$` + strconv.Itoa(i+1) + `" && exec sleep 60) & echo $!; `
	}
	cmd := exec.Command("/bin/bash", append([]string{"-c", script + "wait", "bash"}, dirs...)...)
	cmd.Args[0] = "worktree-lsp" // ps reports argv[0], as for the built wrapper the plugin execs
	cmd.Dir = parentDir
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var pids []int
	r := bufio.NewReader(out)
	for range dirs {
		line, err := r.ReadString('\n')
		pid, _ := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || pid == 0 {
			t.Fatalf("reading a child pid: %q %v", line, err)
		}
		pids = append(pids, pid)
	}
	t.Cleanup(func() {
		for _, p := range pids {
			_ = syscall.Kill(p, syscall.SIGKILL)
		}
		_ = cmd.Wait()
	})
	return pids
}

func scriptsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "..", "..", "..", "scripts")
}

func TestStopUnderKillsChildren(t *testing.T) {
	t.Parallel()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wt, outside := dir+"/wt", dir+"/other"
	for _, d := range []string{wt + "/sub", outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pids := startWrapperChildren(t, dir, wt+"/sub", outside)
	// A process whose parent is not a worktree-lsp is not ours to stop.
	other := dir + "/plain"
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	plain := exec.Command("sleep", "60")
	plain.Dir = other
	if err := plain.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = plain.Process.Kill(); _ = plain.Wait() })

	if err := StopUnder(wt); err != nil {
		t.Fatalf("StopUnder: %v", err)
	}
	if alive(pids[0]) {
		t.Error("the server under the worktree survived")
	}
	if !alive(pids[1]) {
		t.Error("the server outside the worktree was stopped")
	}
	out, err := exec.Command(filepath.Join(scriptsDir(t), "check-worktree-processes.sh"), wt).Output()
	if err != nil || !strings.HasPrefix(string(out), "CLEAR: ") {
		t.Errorf("check-worktree-processes.sh %s: %q %v, want CLEAR", wt, out, err)
	}

	if err := StopUnder(other); err != nil {
		t.Fatalf("StopUnder: %v", err)
	}
	if !alive(plain.Process.Pid) {
		t.Error("StopUnder stopped a process whose parent is not a worktree-lsp")
	}
	if err := StopUnder(dir + "/missing"); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("StopUnder on a missing directory: %v, want a not-exist error", err)
	}
}

// TestQueuedRequestFailsWhenServerDiesBeforeReady: a request held for a
// server that dies before it is ready gets an error answer, never silence.
func TestQueuedRequestFailsWhenServerDiesBeforeReady(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+f.dir+"/fake.log", "die=initialized"), Settle: time.Hour})
	c.call("initialize", initParams(f.main))
	c.notify("textDocument/didOpen", didOpen(f.main+"/a.go"))
	resp, _ := c.call("textDocument/hover", textDoc(f.main+"/a.go"))
	if resp.Error == nil || !strings.Contains(string(resp.Error), "exited") {
		t.Errorf("the held hover got %s / %s, want an exited error", resp.Result, resp.Error)
	}
}
