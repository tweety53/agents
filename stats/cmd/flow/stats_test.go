package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestStatsBottlenecks: the table and sinks line render flowd's ranking,
// -last/-project reach the query, -json prints the body as is.
func TestStatsBottlenecks(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	const body = `{"last":5,"wallMs":100000,"dimensions":["stage"],"ranked":{"stage":[{"dimension":"stage","name":"flow.verify","runs":2,"totalMs":50000,"share":0.5,"p50Ms":20000,"p95Ms":29000}]},"sinks":[{"dimension":"stage","name":"flow.verify","runs":2,"totalMs":50000,"share":0.5,"p50Ms":20000,"p95Ms":29000}]}`
	var query string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Path + "?" + r.URL.RawQuery
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	call := func(args ...string) (int, string) {
		var out, errOut bytes.Buffer
		code := run(context.Background(), append([]string{"stats", "bottlenecks", "-addr", srv.URL}, args...), env.get, strings.NewReader(""), &out, &errOut)
		return code, out.String() + errOut.String()
	}
	code, out := call("--last", "5", "--project", "p")
	if code != 0 || query != "/api/v1/stats/bottlenecks?last=5&project=p" {
		t.Fatalf("code %d query %q out %s", code, query, out)
	}
	for _, want := range []string{"wall clock 1m40s", "flow.verify  2     50s    50.0%  20s  29s", "time sinks: stage flow.verify 50s (50.0%)"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if code, out := call("-json"); code != 0 || out != body+"\n" {
		t.Errorf("-json = %d %q", code, out)
	}
	if code, _ := call("-last", "0"); code != 2 {
		t.Errorf("-last 0 = %d, want 2", code)
	}
}
