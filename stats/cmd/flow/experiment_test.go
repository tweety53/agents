package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func runExperimentArgs(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append([]string{"experiment"}, args...), strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestFlowExperiment drives every `flow experiment` subcommand against a
// stand-in daemon: what each sends, prints and exits with.
func TestFlowExperiment(t *testing.T) {
	var opened map[string]any
	var closeBody map[string]string
	overrides := `{}`
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/experiments":
			_ = json.Unmarshal(body, &opened)
			if opened["name"] == "dup" {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error":"store: experiment conflict: role \"a\" is held by open experiment 3"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":12,"name":"t","cells":{},"startedAt":"2026-10-29T10:00:00Z","summary":"","status":"open"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/experiments":
			if r.URL.Query().Get("open") != "1" {
				_, _ = w.Write([]byte(`[{"id":4,"name":"old run","cells":{"x":["opus/flow-low"]},"startedAt":"2026-10-01T10:00:00Z","closedAt":"2026-10-08T10:00:00Z","verdict":"negative","summary":"s","status":"closed"},` +
					`{"id":5,"name":"t","cells":{"b":["x/z"],"a":["x/y@low","x/y@medium"]},"startedAt":"2026-10-29T10:00:00Z","summary":"","status":"open"}]`))
				return
			}
			_, _ = w.Write([]byte(`[{"id":5,"name":"t","cells":{"a":["x/y@low"]},"startedAt":"2026-10-29T10:00:00Z","summary":"","status":"open"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/experiments/overrides":
			_, _ = w.Write([]byte(overrides))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/experiments/5/close":
			_ = json.Unmarshal(body, &closeBody)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	defer srv.Close()
	addr := []string{"-addr", srv.URL, "-timeout", "2s"}

	code, out, errOut := runExperimentArgs(append([]string{"open"}, append(addr, "-name", "t", "-cell", "a=x/y@low", "-cell", "b=x/z")...)...)
	if code != 0 || out != "experiment 12 open\n" {
		t.Fatalf("open = %d %q %q, want 0 \"experiment 12 open\"", code, out, errOut)
	}
	wantCells := map[string]any{"a": []any{"x/y@low"}, "b": []any{"x/z"}}
	if !reflect.DeepEqual(opened["cells"], wantCells) {
		t.Fatalf("open posted cells %v, want %v", opened["cells"], wantCells)
	}
	if code, _, _ = runExperimentArgs(append([]string{"open"}, append(addr, "-name", "t", "-cell", "a=x/y@low,x/y@medium")...)...); code != 0 ||
		!reflect.DeepEqual(opened["cells"], map[string]any{"a": []any{"x/y@low", "x/y@medium"}}) {
		t.Fatalf("open with a comma list = %d, posted %v; want both pairs in one cell", code, opened["cells"])
	}
	code, _, errOut = runExperimentArgs(append([]string{"open"}, append(addr, "-name", "dup", "-cell", "a=x/y")...)...)
	if code != 1 || !strings.Contains(errOut, "held by open experiment 3") {
		t.Fatalf("open conflict = %d %q, want 1 naming the holder", code, errOut)
	}
	for _, bad := range [][]string{{"-name", "t"}, {"-cell", "a=x/y"}, {"-name", "t", "-cell", "noequals"}, {"-name", "t", "-cell", "a=x/y", "extra"}} {
		if code, _, _ = runExperimentArgs(append([]string{"open"}, append(addr, bad...)...)...); code != 2 {
			t.Errorf("open %v = %d, want 2", bad, code)
		}
	}

	code, out, _ = runExperimentArgs(append([]string{"list"}, addr...)...)
	wantList := "4 closed 2026-10-01 old run x=opus/flow-low\n5 open 2026-10-29 t a=x/y@low,x/y@medium;b=x/z\n"
	if code != 0 || out != wantList {
		t.Fatalf("list = %d %q, want %q", code, out, wantList)
	}
	if code, out, _ = runExperimentArgs(append([]string{"list", "-open"}, addr...)...); code != 0 || out != "5 open 2026-10-29 t a=x/y@low\n" {
		t.Fatalf("list -open = %d %q", code, out)
	}

	if code, out, _ = runExperimentArgs(append([]string{"overrides"}, addr...)...); code != 0 || out != "{}\n" {
		t.Fatalf("overrides (none) = %d %q, want {}", code, out)
	}
	overrides = `{"b": ["x/z"], "a": ["x/y@low"]}`
	if code, out, _ = runExperimentArgs(append([]string{"overrides"}, addr...)...); code != 0 || out != `{"a":["x/y@low"],"b":["x/z"]}`+"\n" {
		t.Fatalf("overrides = %d %q, want compact JSON", code, out)
	}

	if code, _, errOut = runExperimentArgs(append([]string{"close", "5"}, append(addr, "-verdict", "positive", "-summary", "s")...)...); code != 0 ||
		closeBody["verdict"] != "positive" || closeBody["summary"] != "s" {
		t.Fatalf("close = %d %q, posted %v", code, errOut, closeBody)
	}
	if code, _, _ = runExperimentArgs(append([]string{"close", "5"}, append(addr, "-verdict", "great", "-summary", "s")...)...); code != 2 {
		t.Fatalf("close bad verdict = %d, want 2", code)
	}
	if code, _, _ = runExperimentArgs(append([]string{"close", "x"}, append(addr, "-verdict", "positive")...)...); code != 2 {
		t.Fatalf("close bad id = %d, want 2", code)
	}
	if code, _, _ = runExperimentArgs("bogus"); code != 2 {
		t.Fatalf("unknown subcommand = %d, want 2", code)
	}
	if code, _, _ = runExperimentArgs(); code != 2 {
		t.Fatalf("no subcommand = %d, want 2", code)
	}

	dead := []string{"-addr", "http://127.0.0.1:1", "-timeout", "500ms"}
	for _, args := range [][]string{
		append([]string{"open", "-name", "t", "-cell", "a=x/y"}, dead...),
		append([]string{"list"}, dead...),
		append([]string{"overrides"}, dead...),
		append([]string{"close", "5", "-verdict", "negative"}, dead...),
	} {
		if code, _, errOut = runExperimentArgs(args...); code != 1 || errOut == "" {
			t.Errorf("%v on an unreachable store = %d %q, want 1 with a message", args, code, errOut)
		}
	}
}

// TestFlowExperimentStats drives `flow experiment stats <id>`: it prints
// the daemon's JSON as-is, exits 1 on an unknown id or an unreachable
// store and 2 on a malformed id.
func TestFlowExperimentStats(t *testing.T) {
	const body = `{"experiment":{"id":5},"runs":{"before":3,"after":1},"metrics":[]}`
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/experiments/5/stats":
			_, _ = w.Write([]byte(body))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/experiments/9/stats":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"store: experiment not found: 9"}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	defer srv.Close()
	addr := []string{"-addr", srv.URL, "-timeout", "2s"}

	if code, out, errOut := runExperimentArgs(append([]string{"stats", "5"}, addr...)...); code != 0 || out != body+"\n" {
		t.Fatalf("stats = %d %q %q, want 0 printing the daemon's JSON", code, out, errOut)
	}
	if code, _, errOut := runExperimentArgs(append([]string{"stats", "9"}, addr...)...); code != 1 || !strings.Contains(errOut, "experiment not found: 9") {
		t.Fatalf("stats unknown = %d %q, want 1 naming the id", code, errOut)
	}
	for _, args := range [][]string{{"stats"}, append([]string{"stats", "x"}, addr...), append([]string{"stats", "5", "extra"}, addr...)} {
		if code, _, _ := runExperimentArgs(args...); code != 2 {
			t.Errorf("%v = %d, want 2", args, code)
		}
	}
	if code, _, errOut := runExperimentArgs("stats", "5", "-addr", "http://127.0.0.1:1", "-timeout", "500ms"); code != 1 || errOut == "" {
		t.Errorf("stats on an unreachable store = %d %q, want 1 with a message", code, errOut)
	}
}
