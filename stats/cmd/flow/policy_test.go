package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runPolicyArgs(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append([]string{"policy"}, args...), strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestPolicyGetCachesAndFallsBack: get prints the daemon's pairs and caches
// them; with the daemon gone it prints the cache and warns; with neither it
// exits 1 naming both.
func TestPolicyGetCachesAndFallsBack(t *testing.T) {
	state := t.TempDir()
	t.Setenv("FLOW_STATE_DIR", state)
	const policy = `{"planner":{"base":["opus/flow-planner-low"],"optin":["fable/flow-planner-xhigh"],"marked":[]}}`
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/policy" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		_, _ = w.Write([]byte(policy))
	}))
	code, out, errOut := runPolicyArgs("get", "-addr", srv.URL, "-timeout", "2s")
	if code != 0 || out != policy+"\n" || errOut != "" {
		t.Fatalf("get = %d %q %q, want the daemon's pairs", code, out, errOut)
	}
	cache := filepath.Join(state, "policy-cache.json")
	if b, err := os.ReadFile(cache); err != nil || string(b) != policy {
		t.Fatalf("cache = %q, %v; want the daemon's pairs", b, err)
	}
	srv.Close()

	code, out, errOut = runPolicyArgs("get", "-addr", srv.URL, "-timeout", "2s")
	if code != 0 || out != policy+"\n" || !strings.Contains(errOut, "answering from the cache at "+cache) {
		t.Fatalf("get with flowd down = %d %q %q, want the cache and one warning", code, out, errOut)
	}

	if err := os.Remove(cache); err != nil {
		t.Fatal(err)
	}
	code, out, errOut = runPolicyArgs("get", "-addr", srv.URL, "-timeout", "2s")
	if code != 1 || out != "" || !strings.Contains(errOut, "flowd unreachable") || !strings.Contains(errOut, "no usable cache at "+cache) {
		t.Fatalf("get with neither = %d %q %q, want 1 naming flowd and the cache", code, out, errOut)
	}
}

// TestPolicyGrantCommands: grant, grants and revoke send what they name and
// print their outcome; a refused revoke exits 1; usage errors exit 2.
func TestPolicyGrantCommands(t *testing.T) {
	var got []string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.URL.Path == "/api/v1/policy/grants" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"change":"` + r.URL.Query().Get("change") + `","roles":["planner"]}`))
		case r.URL.Path == "/api/v1/policy/grants":
			var dto struct {
				Change string   `json:"change"`
				Roles  []string `json:"roles"`
			}
			_ = json.Unmarshal(body, &dto)
			got = append(got, r.Method+" "+dto.Change+" "+strings.Join(dto.Roles, ","))
			if r.Method == http.MethodDelete && dto.Change == "kan-2" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"store: grant not found"}`))
				return
			}
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	}))
	defer srv.Close()
	addr := []string{"-addr", srv.URL, "-timeout", "2s"}
	if code, out, errOut := runPolicyArgs(append([]string{"grant", "-change", "kan-1", "-roles", "planner,plan-designer"}, addr...)...); code != 0 || out != "policy grant kan-1 planner,plan-designer\n" {
		t.Fatalf("grant = %d %q %q", code, out, errOut)
	}
	if code, out, _ := runPolicyArgs(append([]string{"grants", "-change", "kan-1"}, addr...)...); code != 0 || out != "[\"planner\"]\n" {
		t.Fatalf("grants = %d %q", code, out)
	}
	if code, _, _ := runPolicyArgs(append([]string{"revoke", "-change", "kan-1", "-roles", "planner"}, addr...)...); code != 0 {
		t.Fatalf("revoke = %d", code)
	}
	if code, _, errOut := runPolicyArgs(append([]string{"revoke", "-change", "kan-2", "-roles", "planner"}, addr...)...); code != 1 || !strings.Contains(errOut, "grant not found") {
		t.Fatalf("revoke unknown = %d %q, want 1", code, errOut)
	}
	if want := []string{"PUT kan-1 planner,plan-designer", "DELETE kan-1 planner", "DELETE kan-2 planner"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requests = %q, want %q", got, want)
	}
	for _, args := range [][]string{{}, {"bogus"}, {"grant", "-change", "kan-1"}, {"grants"}, {"get", "-change", "x"}, {"get", "extra"}} {
		if code, _, _ := runPolicyArgs(args...); code != 2 {
			t.Errorf("%q = %d, want 2", args, code)
		}
	}
}
