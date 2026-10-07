package lspmux

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestMain doubles as the fake LSP server: the mux under test starts this
// test binary with `fake-lsp` as its first argument (see fakeServer).
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "fake-lsp" {
		fakeLSP(os.Args[2:])
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fakeRecord is one line of the fake server's log: every message it
// received ("recv") or sent from its script ("sent"), with its own cwd.
type fakeRecord struct {
	Cwd  string          `json:"cwd"`
	Pid  int             `json:"pid"`
	Kind string          `json:"kind"`
	Msg  json.RawMessage `json:"msg"`
}

func (r fakeRecord) message() message {
	var m message
	_ = json.Unmarshal(r.Msg, &m)
	return m
}

// fakeLSP is the fake server. Options, each key=value:
//
//	log=<path>     append a fakeRecord per message (required)
//	script=<path>  a JSON array played after `initialized`: each entry is a
//	               message to send, or {"wait":"<file>"} — block until that
//	               file exists
//	fail=<method>  answer that method with an error
//	die=<method>   exit on that notification, as a crashing server does
//	flood=<method> on that notification, block its one reader as an lsp4j
//	               server does: send workspace/configuration, then 1 MB of
//	               window/logMessage, before reading anything else
//
// Every other request is answered [{"root":<cwd>,"method":<method>}];
// initialize with empty capabilities, shutdown with null; exit ends it.
func fakeLSP(args []string) {
	opt := map[string]string{}
	for _, a := range args {
		k, v, _ := strings.Cut(a, "=")
		opt[k] = v
	}
	cwd, _ := os.Getwd()
	logf, err := os.OpenFile(opt["log"], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		os.Exit(3)
	}
	record := func(kind string, msg []byte) {
		line, _ := json.Marshal(fakeRecord{Cwd: cwd, Pid: os.Getpid(), Kind: kind, Msg: msg})
		_, _ = logf.Write(append(line, '\n'))
	}
	var mu sync.Mutex
	send := func(b []byte) {
		mu.Lock()
		defer mu.Unlock()
		_ = writeMessage(os.Stdout, b)
	}
	reply := func(id json.RawMessage, result, rpcErr string) {
		m := message{JSONRPC: "2.0", ID: id}
		if rpcErr != "" {
			m.Error = json.RawMessage(rpcErr)
		} else {
			m.Result = json.RawMessage(result)
		}
		b, _ := json.Marshal(m)
		send(b)
	}
	r := bufio.NewReader(os.Stdin)
	for {
		b, err := readMessage(r)
		if err != nil {
			return
		}
		record("recv", b)
		var msg message
		_ = json.Unmarshal(b, &msg)
		switch {
		case msg.Method == "exit":
			return
		case msg.Method == "initialized" && opt["script"] != "":
			go fakePlay(opt["script"], send, record)
		case opt["die"] != "" && msg.Method == opt["die"]:
			os.Exit(1)
		case opt["flood"] != "" && msg.Method == opt["flood"]:
			send([]byte(`{"jsonrpc":"2.0","id":"cfg","method":"workspace/configuration","params":{"items":[{}]}}`))
			big, _ := json.Marshal(message{JSONRPC: "2.0", Method: "window/logMessage",
				Params: mustMarshal(map[string]any{"type": 4, "message": strings.Repeat("x", 1<<20)})})
			send(big)
		case msg.Method == "" || msg.ID == nil:
		case msg.Method == "initialize":
			reply(msg.ID, `{"capabilities":{}}`, "")
		case msg.Method == "shutdown":
			reply(msg.ID, "null", "")
		case msg.Method == opt["fail"]:
			reply(msg.ID, "", `{"code":-32603,"message":"fake failure"}`)
		default:
			res, _ := json.Marshal([]map[string]string{{"root": cwd, "method": msg.Method}})
			reply(msg.ID, string(res), "")
		}
	}
}

func fakePlay(path string, send func([]byte), record func(string, []byte)) {
	b, _ := os.ReadFile(path)
	var entries []json.RawMessage
	_ = json.Unmarshal(b, &entries)
	for _, e := range entries {
		var w struct{ Wait string }
		if json.Unmarshal(e, &w); w.Wait != "" {
			for {
				if _, err := os.Stat(w.Wait); err == nil {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			continue
		}
		send(e)
		record("sent", e)
	}
}

// fakeServer is the Mux.Server command that runs fakeLSP with opts.
func fakeServer(t *testing.T, opts ...string) []string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return append([]string{exe, "fake-lsp"}, opts...)
}

func readLog(t *testing.T, path string) []fakeRecord {
	t.Helper()
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var recs []fakeRecord
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		var r fakeRecord
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			t.Fatalf("fake log line %q: %v", line, err)
		}
		recs = append(recs, r)
	}
	return recs
}

// waitLog polls the fake's log until pred holds — the fake is another
// process, so its log is the only thing to wait on.
func waitLog(t *testing.T, path string, pred func([]fakeRecord) bool) []fakeRecord {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		recs := readLog(t, path)
		if pred(recs) {
			return recs
		}
		if time.Now().After(deadline) {
			t.Fatalf("fake log never satisfied the wait; it holds:\n%s", mustRead(path))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func mustRead(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

// testClient drives a Mux through pipes, as Claude Code drives it on stdio.
type testClient struct {
	t      *testing.T
	in     *io.PipeWriter
	outW   *io.PipeWriter
	msgs   chan message
	done   chan error
	nextID int
	once   sync.Once
}

// startMux runs m against a test client. A test that leaves IndexLog or
// Settle unset gets a log in its own temp directory — never the operator's
// cache — and a 1ms settle window.
func startMux(t *testing.T, m *Mux) *testClient {
	t.Helper()
	if m.IndexLog == "" {
		m.IndexLog = t.TempDir() + "/index.log"
	}
	if m.Settle == 0 {
		m.Settle = time.Millisecond
	}
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	c := &testClient{t: t, in: inW, outW: outW, msgs: make(chan message, 1000), done: make(chan error, 1)}
	go func() { c.done <- m.Run(context.Background(), inR, outW) }()
	go func() {
		r := bufio.NewReader(outR)
		for {
			b, err := readMessage(r)
			if err != nil {
				close(c.msgs)
				return
			}
			var msg message
			if err := json.Unmarshal(b, &msg); err != nil {
				t.Errorf("mux wrote invalid JSON %q: %v", b, err)
			}
			c.msgs <- msg
		}
	}()
	t.Cleanup(c.close)
	return c
}

func (c *testClient) write(m message) {
	c.t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		c.t.Fatal(err)
	}
	if err := writeMessage(c.in, b); err != nil {
		c.t.Fatal(err)
	}
}

func rawJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (c *testClient) notify(method string, params any) {
	c.t.Helper()
	c.write(message{JSONRPC: "2.0", Method: method, Params: rawJSON(c.t, params)})
}

// request sends a request and returns its id.
func (c *testClient) request(method string, params any) json.RawMessage {
	c.t.Helper()
	c.nextID++
	id := rawJSON(c.t, c.nextID)
	c.write(message{JSONRPC: "2.0", ID: id, Method: method, Params: rawJSON(c.t, params)})
	return id
}

// next returns the next message the mux wrote to the client.
func (c *testClient) next() message {
	c.t.Helper()
	select {
	case m, ok := <-c.msgs:
		if !ok {
			c.t.Fatal("the mux closed its output")
		}
		return m
	case <-time.After(10 * time.Second):
		c.t.Fatal("no message from the mux within 10s")
	}
	return message{}
}

// call sends a request and returns its response; every other message read
// on the way is returned too, in order.
func (c *testClient) call(method string, params any) (message, []message) {
	c.t.Helper()
	id := c.request(method, params)
	var other []message
	for {
		m := c.next()
		if m.Method == "" && string(m.ID) == string(id) {
			return m, other
		}
		other = append(other, m)
	}
}

// close ends the session as Claude Code's exit does — stdin closes — and
// waits for Run to return, which it does only once every child is gone.
func (c *testClient) close() {
	c.once.Do(func() {
		_ = c.in.Close()
		select {
		case err := <-c.done:
			if err != nil {
				c.t.Errorf("Run: %v", err)
			}
		case <-time.After(15 * time.Second):
			c.t.Error("Run did not return within 15s of stdin closing")
		}
		_ = c.outW.Close()
		for range c.msgs {
		}
	})
}

func fileURI(p string) string { return (&url.URL{Scheme: "file", Path: p}).String() }

func textDoc(path string) map[string]any {
	return map[string]any{"textDocument": map[string]any{"uri": fileURI(path)}}
}

func didOpen(path string) map[string]any {
	return map[string]any{"textDocument": map[string]any{"uri": fileURI(path), "languageId": "go", "version": 1, "text": "package x\n"}}
}

// fixture is a main checkout with one linked worktree, a session directory
// in no repository, and a loose directory in no repository — all physical
// paths, as git reports them.
type fixture struct{ dir, main, wt, session, loose string }

func newFixture(t *testing.T) fixture {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := fixture{dir: dir, main: dir + "/repo", wt: dir + "/repo-worktrees/x", session: dir + "/session", loose: dir + "/loose"}
	for _, d := range []string{f.session, f.loose} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"init", "-q", f.main},
		{"-C", f.main, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", "base"},
		{"-C", f.main, "worktree", "add", "-q", "-b", "x", f.wt},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return f
}

func initParams(root string) map[string]any {
	return map[string]any{
		"processId":        os.Getpid(),
		"rootUri":          fileURI(root),
		"rootPath":         root,
		"workspaceFolders": []map[string]string{{"uri": fileURI(root), "name": filepath.Base(root)}},
		"capabilities":     map[string]any{"textDocument": map[string]any{"hover": map[string]any{"contentFormat": []string{"markdown"}}}},
	}
}

// resultRoot is the root the fake put in its answer.
func resultRoot(t *testing.T, m message) string {
	t.Helper()
	var res []map[string]string
	if err := json.Unmarshal(m.Result, &res); err != nil || len(res) != 1 {
		t.Fatalf("result %s (error %s): want one fake answer", m.Result, m.Error)
	}
	return res[0]["root"]
}

func jsonEqual(t *testing.T, got json.RawMessage, want string) bool {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("%q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("%q: %v", want, err)
	}
	return reflect.DeepEqual(g, w)
}

func TestStaticInitialize(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	log := f.dir + "/fake.log"
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+log)})
	resp, other := c.call("initialize", initParams(f.main))
	want := `{"capabilities":{"textDocumentSync":{"openClose":true,"change":1},"hoverProvider":true,` +
		`"definitionProvider":true,"referencesProvider":true,"implementationProvider":true,` +
		`"documentSymbolProvider":true,"workspaceSymbolProvider":true,"callHierarchyProvider":true},` +
		`"serverInfo":{"name":"worktree-lsp"}}`
	if !jsonEqual(t, resp.Result, want) || resp.Error != nil {
		t.Errorf("initialize answered %s (error %s), want %s", resp.Result, resp.Error, want)
	}
	if len(other) != 0 {
		t.Errorf("messages before the initialize answer: %v", other)
	}
	c.notify("initialized", map[string]any{})
	if _, other := c.call("shutdown", nil); len(other) != 0 {
		t.Errorf("messages before the shutdown answer: %v", other)
	}
	c.close()
	if _, err := os.Stat(log); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a child was started: %s", mustRead(log))
	}
}

func TestRouteByWorktreeRoot(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	log := f.dir + "/fake.log"
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+log)})
	c.call("initialize", initParams(f.session))
	c.notify("initialized", map[string]any{})
	for _, tc := range []struct{ file, root string }{
		{f.main + "/a.go", f.main},
		{f.wt + "/a.go", f.wt},
		{f.loose + "/a.go", f.session},
		{f.main + "/a.go", f.main},
	} {
		c.notify("textDocument/didOpen", didOpen(tc.file))
		resp, _ := c.call("textDocument/hover", textDoc(tc.file))
		if got := resultRoot(t, resp); got != tc.root {
			t.Errorf("hover on %s answered by the child at %s, want %s", tc.file, got, tc.root)
		}
	}
	recs := readLog(t, log)
	pids := map[string]map[int]bool{}
	for _, r := range recs {
		if pids[r.Cwd] == nil {
			pids[r.Cwd] = map[int]bool{}
		}
		pids[r.Cwd][r.Pid] = true
		if m := r.message(); m.Method == "textDocument/didOpen" {
			var p struct{ TextDocument struct{ URI string } }
			_ = json.Unmarshal(m.Params, &p)
			u, _ := url.Parse(p.TextDocument.URI)
			if want := map[string]string{f.main: f.main, f.wt: f.wt, f.loose: f.session}[filepath.Dir(u.Path)]; want != r.Cwd {
				t.Errorf("didOpen of %s reached the child at %s, want %s", u.Path, r.Cwd, want)
			}
		}
	}
	if len(pids) != 3 {
		t.Errorf("children ran at %d roots, want 3: %v", len(pids), pids)
	}
	c.close()
	for root, ps := range pids {
		if len(ps) != 1 {
			t.Errorf("%d children at %s, want 1", len(ps), root)
		}
		for pid := range ps {
			if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
				t.Errorf("child %d at %s outlived Run: %v", pid, root, err)
			}
		}
	}
}

func TestLazySpawnRewritesRoot(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	log := f.dir + "/fake.log"
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+log)})
	c.call("initialize", initParams(f.main))
	c.notify("initialized", map[string]any{})
	c.notify("textDocument/didOpen", didOpen(f.wt+"/a.go"))
	recs := waitLog(t, log, func(rs []fakeRecord) bool {
		for _, r := range rs {
			if r.message().Method == "textDocument/didOpen" {
				return true
			}
		}
		return false
	})
	var methods []string
	for _, r := range recs {
		if r.Cwd != f.wt {
			t.Errorf("a child ran at %s, want only %s", r.Cwd, f.wt)
		}
		m := r.message()
		methods = append(methods, m.Method)
		if m.Method != "initialize" {
			continue
		}
		var p map[string]json.RawMessage
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		for k, want := range map[string]string{
			"rootUri":          `"` + fileURI(f.wt) + `"`,
			"rootPath":         `"` + f.wt + `"`,
			"workspaceFolders": `[{"uri":"` + fileURI(f.wt) + `","name":"x"}]`,
			"processId":        rawJSONString(t, os.Getpid()),
		} {
			if !jsonEqual(t, p[k], want) {
				t.Errorf("child initialize %s = %s, want %s", k, p[k], want)
			}
		}
		var caps struct{ TextDocument map[string]json.RawMessage }
		_ = json.Unmarshal(p["capabilities"], &caps)
		if !jsonEqual(t, caps.TextDocument["hover"], `{"contentFormat":["markdown"]}`) {
			t.Errorf("the client's capabilities were not passed on: %s", p["capabilities"])
		}
	}
	if want := []string{"initialize", "initialized", "textDocument/didOpen"}; !reflect.DeepEqual(methods, want) {
		t.Errorf("child received %v, want %v", methods, want)
	}
}

func rawJSONString(t *testing.T, v any) string { return string(rawJSON(t, v)) }

// TestBlockingServerNeverStallsTheMux: a server that writes while it is not
// reading must never stall the mux behind a write to its stdin. Before each
// child got its own writer, answering the server's request needed the lock a
// 1 MB client write held while the server, flooding its stdout, read nothing.
func TestBlockingServerNeverStallsTheMux(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+f.dir+"/fake.log", "flood=textDocument/didOpen")})
	c.call("initialize", initParams(f.session))
	c.notify("initialized", map[string]any{})
	file := f.main + "/a.go"
	big := strings.Repeat("y", 1<<20)
	c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": fileURI(file), "languageId": "go", "version": 1, "text": big}})
	c.notify("textDocument/didChange", map[string]any{"textDocument": map[string]any{"uri": fileURI(file), "version": 2},
		"contentChanges": []map[string]any{{"text": big}}})
	resp, _ := c.call("textDocument/hover", textDoc(file))
	if got := resultRoot(t, resp); got != f.main {
		t.Errorf("hover answered by %s, want %s", got, f.main)
	}
	c.close()
}

func TestServerRequestsAnsweredLocally(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	log, script := f.dir+"/fake.log", f.dir+"/script.json"
	answers := map[string]string{
		"1": `[null,null]`,
		"2": `null`,
		"3": `null`,
		"4": `null`,
		"5": `null`,
		"6": `{"applied":false}`,
	}
	if err := os.WriteFile(script, []byte(`[
		{"jsonrpc":"2.0","id":1,"method":"workspace/configuration","params":{"items":[{"section":"gopls"},{"section":"other"}]}},
		{"jsonrpc":"2.0","id":2,"method":"window/workDoneProgress/create","params":{"token":"t"}},
		{"jsonrpc":"2.0","id":3,"method":"client/registerCapability","params":{"registrations":[]}},
		{"jsonrpc":"2.0","id":4,"method":"client/unregisterCapability","params":{"unregisterations":[]}},
		{"jsonrpc":"2.0","id":5,"method":"window/showMessageRequest","params":{"type":1,"message":"m"}},
		{"jsonrpc":"2.0","id":6,"method":"workspace/applyEdit","params":{"edit":{}}},
		{"jsonrpc":"2.0","method":"textDocument/publishDiagnostics","params":{"uri":"file:///x.go","diagnostics":[]}}
	]`), 0o644); err != nil {
		t.Fatal(err)
	}
	c := startMux(t, &Mux{Server: fakeServer(t, "log="+log, "script="+script)})
	c.call("initialize", initParams(f.main))
	c.notify("initialized", map[string]any{})
	c.notify("textDocument/didOpen", didOpen(f.main+"/a.go"))
	if m := c.next(); m.Method != "textDocument/publishDiagnostics" || !jsonEqual(t, m.Params, `{"uri":"file:///x.go","diagnostics":[]}`) {
		t.Errorf("first message to the client: %+v, want the forwarded publishDiagnostics", m)
	}
	recs := waitLog(t, log, func(rs []fakeRecord) bool {
		n := 0
		for _, r := range rs {
			if m := r.message(); r.Kind == "recv" && m.Method == "" {
				n++
			}
		}
		return n == len(answers)
	})
	for _, r := range recs {
		m := r.message()
		if r.Kind != "recv" || m.Method != "" {
			continue
		}
		want, ok := answers[string(m.ID)]
		if !ok || m.Error != nil || !jsonEqual(t, m.Result, want) {
			t.Errorf("answer to server request %s: result %s error %s, want %s", m.ID, m.Result, m.Error, want)
		}
	}
	if _, other := c.call("shutdown", nil); len(other) != 0 {
		t.Errorf("server requests reached the client: %+v", other)
	}
}

func TestRespawnReopensDocuments(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	log := f.dir + "/fake.log"
	m := &Mux{Server: fakeServer(t, "log="+log, "die=textDocument/didSave")}
	c := startMux(t, m)
	c.call("initialize", initParams(f.main))
	c.notify("initialized", map[string]any{})
	a, b := f.main+"/a.go", f.main+"/b.go"
	c.notify("textDocument/didOpen", didOpen(a))
	c.notify("textDocument/didOpen", didOpen(b))
	c.notify("textDocument/didChange", map[string]any{"textDocument": map[string]any{"uri": fileURI(a), "version": 2},
		"contentChanges": []map[string]any{{"text": "package y\n"}}})
	c.notify("textDocument/didClose", textDoc(b))
	c.call("textDocument/hover", textDoc(a))
	first := readLog(t, log)[0].Pid

	c.notify("textDocument/didSave", textDoc(a)) // the server dies
	deadline := time.Now().Add(10 * time.Second)
	for len(m.running()) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("the dead server was never forgotten")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if resp, _ := c.call("textDocument/hover", textDoc(a)); resultRoot(t, resp) != f.main {
		t.Fatalf("hover after the respawn: %s", resp.Result)
	}

	var got []string
	opens := 0 // the first server's didOpens: route started it, so nothing is replayed to it
	for _, r := range readLog(t, log) {
		if r.Pid == first && r.message().Method == "textDocument/didOpen" {
			opens++
		}
		if r.Pid == first || r.Kind != "recv" {
			continue
		}
		msg := r.message()
		got = append(got, msg.Method)
		if msg.Method == "textDocument/didOpen" && !jsonEqual(t, msg.Params,
			`{"textDocument":{"uri":"`+fileURI(a)+`","languageId":"go","version":2,"text":"package y\n"}}`) {
			t.Errorf("replayed didOpen %s, want a.go at version 2 with its changed text", msg.Params)
		}
	}
	if opens != 2 {
		t.Errorf("the first server received %d didOpens, want 2 — one per document", opens)
	}
	want := []string{"initialize", "initialized", "textDocument/didOpen", "textDocument/hover"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the respawned server received %q, want %q", got, want)
	}
}
