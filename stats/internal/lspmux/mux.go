// Package lspmux is the worktree-lsp wrapper's core: it speaks LSP to one
// client (Claude Code) and runs one real language server per git worktree,
// started the first time a document of that worktree is opened or queried
// and rooted there (design.md, wrapper-per-session). The client's initialize
// is answered here with a fixed capability set (static-initialize); requests
// the servers send the client are answered here too, so the client only ever
// sees the answers to its own requests and the servers' notifications.
package lspmux

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Mux multiplexes one client over a server per git worktree. Set its fields,
// then call Run once.
type Mux struct {
	// Server is the real server's command and arguments.
	Server []string
	// Log receives the mux's own log lines and every server's stderr; nil
	// discards them. It must be safe for concurrent use, as an *os.File is.
	Log io.Writer
	// Settle is how long a server must report no progress, once it answered
	// initialize, before it counts as ready; zero means defaultSettle.
	Settle time.Duration
	// WatchInterval is how often each root is checked for removal; zero
	// means defaultWatchInterval.
	WatchInterval time.Duration
	// IndexLog is the file each ready line is appended to; "" means
	// defaultIndexLog().
	IndexLog string

	out   io.Writer
	outMu sync.Mutex

	mu       sync.Mutex
	init     json.RawMessage           // the client's initialize params
	session  string                    // the session root: a request naming no file goes there
	children map[string]*child         // running servers, by root
	roots    map[string]string         // directory -> its git toplevel, "" for none
	docs     map[string]map[string]doc // open documents, by root then uri
	wg       sync.WaitGroup            // one per child's reader goroutine
}

// message is any JSON-RPC message: a request has Method and ID, a
// notification Method only, a response ID and Result or Error.
type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

// initializeResult is the fixed answer to the client's initialize: full text
// sync and the operations Claude Code's LSP tool issues.
const initializeResult = `{"capabilities":{"textDocumentSync":{"openClose":true,"change":1},` +
	`"hoverProvider":true,"definitionProvider":true,"referencesProvider":true,` +
	`"implementationProvider":true,"documentSymbolProvider":true,"workspaceSymbolProvider":true,` +
	`"callHierarchyProvider":true},"serverInfo":{"name":"worktree-lsp"}}`

// stopGrace is how long a server gets to exit after `exit` before it is
// killed.
const stopGrace = 5 * time.Second

const (
	shutdownRequest = `{"jsonrpc":"2.0","id":"worktree-lsp-shutdown","method":"shutdown"}`
	exitNotice      = `{"jsonrpc":"2.0","method":"exit"}`
)

// Run serves the client on in/out until it sends exit, in ends, or ctx is
// cancelled, then stops every server and returns once each has exited.
// Meanwhile the server of a root that no longer exists is stopped. A
// cancelled ctx leaves one goroutine blocked reading in until in is closed.
func (m *Mux) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	m.out = out
	m.children, m.roots, m.docs = map[string]*child{}, map[string]string{}, map[string]map[string]doc{}
	msgs, readErr, quit := make(chan []byte), make(chan error, 1), make(chan struct{})
	go func() {
		r := bufio.NewReader(in)
		for {
			b, err := readMessage(r)
			if err != nil {
				readErr <- err
				return
			}
			select {
			case msgs <- b:
			case <-quit:
				return
			}
		}
	}()
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		m.watch(quit)
	}()
	defer func() {
		close(quit)
		<-watched
		m.stopAll()
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-readErr:
			if err == io.EOF {
				return nil
			}
			return err
		case b := <-msgs:
			if m.handle(b) {
				return nil
			}
		}
	}
}

// handle acts on one client message and reports whether it was exit.
func (m *Mux) handle(b []byte) bool {
	var msg message
	if err := json.Unmarshal(b, &msg); err != nil {
		m.logf("dropping an unreadable client message: %v", err)
		return false
	}
	switch msg.Method {
	case "":
		// A response: this mux never sends the client a request.
	case "initialize":
		m.mu.Lock()
		m.init = msg.Params
		m.mu.Unlock()
		m.session = m.sessionRoot(msg.Params)
		m.reply(msg.ID, json.RawMessage(initializeResult))
	case "initialized":
		// Each server gets its own, once it answers its initialize.
	case "$/cancelRequest":
		// ponytail: dropped — the client ignores the late answer; forward it
		// with the id remapped if a server's wasted work ever matters.
	case "shutdown":
		for _, c := range m.running() {
			c.write([]byte(shutdownRequest))
		}
		m.reply(msg.ID, json.RawMessage("null"))
	case "exit":
		return true
	case "workspace/symbol":
		m.fanOut(msg)
	default:
		m.route(msg, b)
		m.track(msg)
	}
	return false
}

// doc is an open document as the client last sent it: initializeResult
// declares full sync, so every didChange carries the whole text.
type doc struct {
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// track remembers the documents the client holds open, so a server started
// after its root's last one died is handed them again (child). It runs after
// route, so a server that route itself started never gets a document twice.
func (m *Mux) track(msg message) {
	switch msg.Method {
	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose":
	default:
		return
	}
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
			doc
		} `json:"textDocument"`
		ContentChanges []struct {
			Text string `json:"text"`
		} `json:"contentChanges"`
	}
	if json.Unmarshal(msg.Params, &p) != nil || p.TextDocument.URI == "" {
		return
	}
	uri := p.TextDocument.URI
	root := m.rootOf(uri)
	m.mu.Lock()
	defer m.mu.Unlock()
	docs := m.docs[root]
	switch msg.Method {
	case "textDocument/didOpen":
		if docs == nil {
			docs = map[string]doc{}
			m.docs[root] = docs
		}
		docs[uri] = p.TextDocument.doc
	case "textDocument/didChange":
		d, ok := docs[uri]
		if !ok || len(p.ContentChanges) == 0 {
			return
		}
		d.Version, d.Text = p.TextDocument.Version, p.ContentChanges[len(p.ContentChanges)-1].Text
		docs[uri] = d
	case "textDocument/didClose":
		delete(docs, uri)
	}
}

// route sends a document message to the server of the document's worktree,
// a request naming no document to the session root's server, and a
// notification naming none to every running server.
func (m *Mux) route(msg message, raw []byte) {
	uri := documentURI(msg.Params)
	if uri == "" && msg.ID == nil {
		for _, c := range m.running() {
			c.send(raw)
		}
		return
	}
	root := m.session
	if uri != "" {
		root = m.rootOf(uri)
	}
	c, err := m.child(root)
	if err != nil {
		m.logf("%v", err)
		if msg.ID != nil {
			m.replyError(msg.ID, err.Error())
		}
		return
	}
	if msg.ID == nil {
		c.send(raw)
		return
	}
	m.forward(c, msg)
}

// forward sends a client request to c and its answer back under the
// client's id.
func (m *Mux) forward(c *child, msg message) {
	id := msg.ID
	c.request(msg.Method, msg.Params, func(r message) {
		r.ID = id
		m.toClient(r)
	})
}

// documentURI is the document a message is about — its textDocument, or a
// call hierarchy item's — "" for none.
func documentURI(params json.RawMessage) string {
	var p struct {
		TextDocument struct{ URI string } `json:"textDocument"`
		Item         struct{ URI string } `json:"item"`
	}
	_ = json.Unmarshal(params, &p)
	if p.TextDocument.URI != "" {
		return p.TextDocument.URI
	}
	return p.Item.URI
}

// sessionRoot is the root the client's initialize names — its git toplevel
// when it has one — else the working directory.
func (m *Mux) sessionRoot(params json.RawMessage) string {
	var p struct {
		RootURI          string `json:"rootUri"`
		RootPath         string `json:"rootPath"`
		WorkspaceFolders []struct {
			URI string `json:"uri"`
		} `json:"workspaceFolders"`
	}
	_ = json.Unmarshal(params, &p)
	dir := ""
	switch {
	case p.RootURI != "":
		dir = uriPath(p.RootURI)
	case p.RootPath != "":
		dir = p.RootPath
	case len(p.WorkspaceFolders) > 0:
		dir = uriPath(p.WorkspaceFolders[0].URI)
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if r := m.rootFor(dir); r != "" {
		return r
	}
	return dir
}

// rootOf is the root whose server owns uri: its directory's git toplevel,
// else the session root.
func (m *Mux) rootOf(uri string) string {
	if p := uriPath(uri); p != "" {
		if r := m.rootFor(filepath.Dir(p)); r != "" {
			return r
		}
	}
	return m.session
}

// rootFor is dir's git toplevel, "" when dir is in no repository; cached per
// directory.
func (m *Mux) rootFor(dir string) string {
	m.mu.Lock()
	r, ok := m.roots[dir]
	m.mu.Unlock()
	if ok {
		return r
	}
	if out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
		r = strings.TrimSpace(string(out))
	}
	m.mu.Lock()
	m.roots[dir] = r
	m.mu.Unlock()
	return r
}

// uriPath is a file URI's path, "" for any other URI.
func uriPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	return u.Path
}

func uriOf(path string) string { return (&url.URL{Scheme: "file", Path: path}).String() }

// running is a snapshot of the running servers.
func (m *Mux) running() []*child {
	m.mu.Lock()
	defer m.mu.Unlock()
	cs := make([]*child, 0, len(m.children))
	for _, c := range m.children {
		cs = append(cs, c)
	}
	return cs
}

// child is root's server, started on first use.
func (m *Mux) child(root string) (*child, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c := m.children[root]; c != nil {
		return c, nil
	}
	cmd := exec.Command(m.Server[0], m.Server[1:]...)
	cmd.Dir = root
	if m.Log != nil {
		cmd.Stderr = m.Log
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("start %s in %s: %w", m.Server[0], root, err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("start %s in %s: %w", m.Server[0], root, err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s in %s: %w", m.Server[0], root, err)
	}
	c := &child{root: root, cmd: cmd, started: time.Now(), stdin: stdin, done: make(chan struct{}),
		wake: make(chan struct{}, 1), pending: map[int64]func(message){}, tokens: map[string]bool{}}
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		c.writer()
	}()
	// id 0 is initialize's; it is written past the queue every other
	// message waits in until the server is ready.
	c.pending[0] = func(r message) { m.initialized(c, r) }
	// A server replacing one that died gets the documents the client still
	// holds open in its root; a first server's list is empty.
	for uri, d := range m.docs[root] {
		c.reopen = append(c.reopen, mustMarshal(message{JSONRPC: "2.0", Method: "textDocument/didOpen",
			Params: mustMarshal(map[string]any{"textDocument": map[string]any{
				"uri": uri, "languageId": d.LanguageID, "version": d.Version, "text": d.Text}})}))
	}
	c.write(mustMarshal(message{JSONRPC: "2.0", ID: json.RawMessage("0"), Method: "initialize", Params: childInitParams(m.init, root)}))
	m.children[root] = c
	m.wg.Add(1)
	go m.readChild(c, stdout)
	return c, nil
}

// childInitParams is the client's initialize params with every root field
// naming root, asking for progress reports.
func childInitParams(init json.RawMessage, root string) json.RawMessage {
	p := map[string]json.RawMessage{}
	_ = json.Unmarshal(init, &p)
	p["capabilities"] = withProgress(p["capabilities"])
	uri := uriOf(root)
	p["rootUri"] = mustMarshal(uri)
	p["rootPath"] = mustMarshal(root)
	p["workspaceFolders"] = mustMarshal([]map[string]string{{"uri": uri, "name": filepath.Base(root)}})
	return mustMarshal(p)
}

// initialized is c's answer to initialize: on success the server gets
// `initialized`, and the messages queued for it once it is ready.
func (m *Mux) initialized(c *child, r message) {
	if r.Error != nil {
		m.logf("%s: initialize failed: %s", c.root, r.Error)
		_ = c.cmd.Process.Kill()
		return
	}
	c.write([]byte(`{"jsonrpc":"2.0","method":"initialized","params":{}}`))
	for _, b := range c.reopen {
		c.write(b)
	}
	m.initDone(c)
}

// readChild reads c's messages until it exits, then fails what it still
// owed and forgets it.
func (m *Mux) readChild(c *child, stdout io.Reader) {
	defer m.wg.Done()
	r := bufio.NewReader(stdout)
	for {
		b, err := readMessage(r)
		if err != nil {
			if err != io.EOF {
				m.logf("%s: %v — stopping its server", c.root, err)
				_ = c.cmd.Process.Kill()
			}
			break
		}
		m.fromChild(c, b)
	}
	_ = c.cmd.Wait()
	m.mu.Lock()
	if m.children[c.root] == c {
		delete(m.children, c.root)
	}
	m.mu.Unlock()
	c.die()
	close(c.done)
}

// fromChild acts on one server message.
func (m *Mux) fromChild(c *child, b []byte) {
	var msg message
	if err := json.Unmarshal(b, &msg); err != nil {
		m.logf("%s: dropping an unreadable server message: %v", c.root, err)
		return
	}
	switch {
	case msg.Method == "":
		c.answered(msg)
	case msg.ID != nil:
		a := localAnswer(msg)
		c.write(mustMarshal(a))
	case msg.Method == "$/progress":
		// Readiness input; the client never created these tokens.
		m.progress(c, msg.Params)
	default:
		m.toClientRaw(b)
	}
}

// localAnswer answers a request a server sent the client.
func localAnswer(msg message) message {
	a := message{JSONRPC: "2.0", ID: msg.ID, Result: json.RawMessage("null")}
	switch msg.Method {
	case "workspace/configuration":
		var p struct{ Items []json.RawMessage }
		_ = json.Unmarshal(msg.Params, &p)
		a.Result = mustMarshal(make([]any, len(p.Items)))
	case "window/workDoneProgress/create", "client/registerCapability", "client/unregisterCapability", "window/showMessageRequest":
	case "workspace/applyEdit":
		a.Result = json.RawMessage(`{"applied":false}`)
	default:
		a.Result = nil
		a.Error = mustMarshal(map[string]any{"code": -32601, "message": "worktree-lsp does not answer " + msg.Method})
	}
	return a
}

// stopAll sends every server exit, kills one still running after
// stopGrace, and returns once every reader goroutine has finished.
func (m *Mux) stopAll() {
	var wg sync.WaitGroup
	for _, c := range m.running() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.stop()
		}()
	}
	wg.Wait()
	m.wg.Wait()
}

func (m *Mux) reply(id, result json.RawMessage) {
	m.toClient(message{JSONRPC: "2.0", ID: id, Result: result})
}

func (m *Mux) replyError(id json.RawMessage, text string) {
	m.toClient(message{JSONRPC: "2.0", ID: id, Error: mustMarshal(map[string]any{"code": -32603, "message": text})})
}

func (m *Mux) toClient(msg message) { m.toClientRaw(mustMarshal(msg)) }

func (m *Mux) toClientRaw(b []byte) {
	m.outMu.Lock()
	defer m.outMu.Unlock()
	if err := writeMessage(m.out, b); err != nil {
		m.logf("writing to the client: %v", err)
	}
}

func (m *Mux) logf(format string, a ...any) {
	if m.Log != nil {
		fmt.Fprintf(m.Log, "worktree-lsp: "+format+"\n", a...)
	}
}

// mustMarshal marshals values this package builds, which always marshal.
func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// child is one running server and the requests it still owes.
type child struct {
	root    string
	cmd     *exec.Cmd
	started time.Time
	reopen  [][]byte      // didOpen per document to replay after initialized
	done    chan struct{} // closed once the server exited and its reader finished

	stdin io.WriteCloser // written by writer() alone
	wake  chan struct{}  // cap 1: the outbox gained a message, or closing was set

	mu      sync.Mutex // guards the fields below
	outbox  [][]byte   // messages for writer(), in order
	closing bool       // writer() closes stdin once the outbox is drained
	nextID  int64
	pending map[int64]func(message) // nil once the server died
	ready   bool
	queue   [][]byte // client traffic held until ready, in order

	// readiness, ready.go
	initDone bool
	tokens   map[string]bool // progress tokens begun and not ended
	gen      int             // the settle timer that may make it ready
	timer    *time.Timer
}

// write sends b now, past the ready queue. It never blocks on the server:
// writer() does the writing, so a server that writes while it is not reading
// can never stall a caller holding c.mu — the reader answering its requests
// among them.
func (c *child) write(b []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.postLocked(b)
}

// postLocked appends b to the outbox; c.mu is held.
func (c *child) postLocked(b []byte) {
	c.outbox = append(c.outbox, b)
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

// writer writes the outbox to the server's stdin in order, then closes stdin
// once stop asked it to and the outbox is drained. A dead server's write
// fails; its reader reports the death, and writer ends with it.
func (c *child) writer() {
	for {
		c.mu.Lock()
		batch, closing := c.outbox, c.closing
		c.outbox = nil
		c.mu.Unlock()
		for _, b := range batch {
			_ = writeMessage(c.stdin, b)
		}
		if len(batch) > 0 {
			continue
		}
		if closing {
			_ = c.stdin.Close()
			return
		}
		select {
		case <-c.wake:
		case <-c.done:
			return
		}
	}
}

// send sends b once the server is ready, after everything queued before it.
func (c *child) send(b []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ready {
		c.postLocked(b)
		return
	}
	c.queue = append(c.queue, b)
}

// request sends a request under the next free id, as send does; cb gets the
// answer, or an error answer if the server dies first.
func (c *child) request(method string, params json.RawMessage, cb func(message)) {
	c.mu.Lock()
	if c.pending == nil {
		c.mu.Unlock()
		cb(c.exited())
		return
	}
	c.nextID++
	c.pending[c.nextID] = cb
	b := mustMarshal(message{JSONRPC: "2.0", ID: json.RawMessage(strconv.FormatInt(c.nextID, 10)), Method: method, Params: params})
	if c.ready {
		c.postLocked(b)
	} else {
		c.queue = append(c.queue, b)
	}
	c.mu.Unlock()
}

// answered hands a response to whoever is waiting for it.
func (c *child) answered(msg message) {
	var id int64
	if json.Unmarshal(msg.ID, &id) != nil {
		return // an id this mux never issued, e.g. shutdownRequest's
	}
	c.mu.Lock()
	cb := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if cb != nil {
		cb(msg)
	}
}

// readyLocked flushes the queue and lets later traffic straight through;
// c.mu is held.
func (c *child) readyLocked() {
	for _, b := range c.queue {
		c.postLocked(b)
	}
	c.queue, c.ready = nil, true
}

// die fails every request the server still owes.
func (c *child) die() {
	c.mu.Lock()
	pending := c.pending
	c.pending, c.queue = nil, nil
	if c.timer != nil {
		c.timer.Stop()
	}
	c.mu.Unlock()
	for _, cb := range pending {
		cb(c.exited())
	}
}

func (c *child) exited() message {
	return message{JSONRPC: "2.0", Error: mustMarshal(map[string]any{"code": -32603, "message": "worktree-lsp: the server for " + c.root + " exited"})}
}

// stop sends exit, closes stdin, and kills the server if it is still
// running after stopGrace; it returns once the server's reader finished.
func (c *child) stop() {
	c.mu.Lock()
	c.postLocked([]byte(exitNotice))
	c.closing = true
	c.mu.Unlock()
	select {
	case <-c.done:
		return
	case <-time.After(stopGrace):
	}
	_ = c.cmd.Process.Kill()
	<-c.done
}
