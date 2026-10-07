package lspmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Readiness (design.md, readiness-progress-settle): a server is ready once
// it answered initialize, every $/progress token it began has ended, and no
// token began for a settle window. Until then everything the client sends
// it waits in its queue. Readiness latches: later progress — kotlin-lsp
// re-indexes in short bursts for minutes after its import — holds nothing.

const (
	// defaultSettle covers the gap between a server's initialize answer and
	// its first progress token, and between tokens: measured ~0.03s for
	// gopls and kotlin-lsp on 2026-10-07.
	defaultSettle        = 2 * time.Second
	defaultWatchInterval = 5 * time.Second
)

// defaultIndexLog is ${XDG_CACHE_HOME:-$HOME/.cache}/worktree-lsp/index.log.
func defaultIndexLog() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "worktree-lsp", "index.log")
}

func orDefault(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

// withProgress is caps with window.workDoneProgress set, so a server
// reports its indexing as $/progress whatever the client declared.
func withProgress(caps json.RawMessage) json.RawMessage {
	c := map[string]json.RawMessage{}
	_ = json.Unmarshal(caps, &c)
	w := map[string]json.RawMessage{}
	_ = json.Unmarshal(c["window"], &w)
	w["workDoneProgress"] = json.RawMessage("true")
	c["window"] = mustMarshal(w)
	return mustMarshal(c)
}

// initDone records that c answered initialize.
func (m *Mux) initDone(c *child) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.initDone = true
	if len(c.tokens) == 0 {
		m.armLocked(c)
	}
}

// progress tracks one $/progress notification of c's.
func (m *Mux) progress(c *child, params json.RawMessage) {
	var p struct {
		Token json.RawMessage `json:"token"`
		Value struct {
			Kind string `json:"kind"`
		} `json:"value"`
	}
	if json.Unmarshal(params, &p) != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ready {
		return
	}
	switch p.Value.Kind {
	case "begin":
		c.tokens[string(p.Token)] = true
		c.gen++ // a settle timer already running no longer counts
	case "end":
		delete(c.tokens, string(p.Token))
		if c.initDone && len(c.tokens) == 0 {
			m.armLocked(c)
		}
	}
}

// armLocked (re)starts c's settle timer; c.mu is held.
func (m *Mux) armLocked(c *child) {
	c.gen++
	gen := c.gen
	if c.timer != nil {
		c.timer.Stop()
	}
	c.timer = time.AfterFunc(orDefault(m.Settle, defaultSettle), func() { m.settled(c, gen) })
}

// settled makes c ready if nothing began since timer gen was armed. The
// ready line is written before the queue is released, so whoever got an
// answer from c can read it.
func (m *Mux) settled(c *child, gen int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if gen != c.gen || c.ready || c.pending == nil {
		return
	}
	m.logReady(c.root, time.Since(c.started).Milliseconds())
	c.readyLocked()
}

// logReady appends `<RFC3339> <server> <root> ready <ms>` to the index log.
func (m *Mux) logReady(root string, ms int64) {
	path := m.IndexLog
	if path == "" {
		path = defaultIndexLog()
	}
	line := fmt.Sprintf("%s %s %s ready %d\n", time.Now().Format(time.RFC3339), filepath.Base(m.Server[0]), root, ms)
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err == nil {
		var f *os.File
		if f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			_, err = f.WriteString(line)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
	}
	if err != nil {
		m.logf("index log %s: %v", path, err)
	}
}

// fanOut sends workspace/symbol to every running server and answers with
// their results concatenated (workspace-symbol-fanout); a server's error is
// logged and drops only its own part.
func (m *Mux) fanOut(msg message) {
	cs := m.running()
	if len(cs) == 0 {
		m.reply(msg.ID, json.RawMessage("[]"))
		return
	}
	var mu sync.Mutex
	left, all := len(cs), []json.RawMessage{}
	for _, c := range cs {
		c.request(msg.Method, msg.Params, func(r message) {
			var part []json.RawMessage
			if r.Error != nil {
				m.logf("%s: %s: %s", c.root, msg.Method, r.Error)
			} else if err := json.Unmarshal(r.Result, &part); err != nil {
				m.logf("%s: %s: unreadable result: %v", c.root, msg.Method, err)
			}
			mu.Lock()
			all = append(all, part...)
			left--
			done := left == 0
			mu.Unlock()
			if done {
				m.reply(msg.ID, mustMarshal(all))
			}
		})
	}
}

// watch stops the server of every root that no longer exists, each
// interval, until quit closes (stop-at-cleanup-and-removal).
func (m *Mux) watch(quit <-chan struct{}) {
	t := time.NewTicker(orDefault(m.WatchInterval, defaultWatchInterval))
	defer t.Stop()
	for {
		select {
		case <-quit:
			return
		case <-t.C:
			for _, c := range m.running() {
				if _, err := os.Stat(c.root); errors.Is(err, fs.ErrNotExist) {
					m.logf("%s is gone — stopping its server", c.root)
					_ = c.cmd.Process.Kill()
					m.mu.Lock()
					delete(m.docs, c.root) // nothing left to reopen them in
					m.mu.Unlock()
				}
			}
		}
	}
}
