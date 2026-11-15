package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// rofFixture is one `flow record dispatches` row; ended is how long ago the
// row closed (0 leaves endedAt null, an open row) and ctx its
// signals.sidechain.context_end ("" leaves the signal out).
func rofFixture(seq int64, key, agent string, ended time.Duration, ctx string) map[string]any {
	row := map[string]any{"seq": seq, "key": key, "agentId": agent, "endedAt": nil, "metrics": map[string]any{}}
	if ended > 0 {
		row["endedAt"] = time.Now().Add(-ended).Format(time.RFC3339Nano)
	}
	if ctx != "" {
		row["metrics"] = map[string]any{"signals": map[string]any{"sidechain": map[string]any{"context_end": ctx}}}
	}
	return row
}

func rofRun(t *testing.T, rows []byte, readErr error, args ...string) (int, string) {
	t.Helper()
	fn := Registry["resume-or-fresh"]
	if fn == nil {
		t.Fatal("resume-or-fresh is not registered")
	}
	env := Env{Getenv: pathEnv(t.TempDir()), Dir: t.TempDir(),
		Dispatches: func(string) ([]byte, error) { return rows, readErr }}
	var out, errb bytes.Buffer
	code := fn(args, env, &out, &errb)
	return code, out.String()
}

func TestResumeOrFresh(t *testing.T) {
	const key = "task-3-implementer"
	rows := func(r ...map[string]any) []byte {
		b, _ := json.Marshal(r)
		return b
	}
	for _, c := range []struct {
		name string
		rows []byte
		err  error
		code int
		out  string
	}{
		{"warm small agent resumes", rows(rofFixture(1, key, "a1", 10*time.Second, "42000")), nil, 0, "resume a1\n"},
		{"store unreachable", nil, errors.New("flow exit 1"), 1, "fresh: rows unreadable\n"},
		{"rows not json", []byte("not json"), nil, 1, "fresh: rows unreadable\n"},
		{"no row under the key", rows(rofFixture(1, "panel-1-primary", "a1", 10*time.Second, "42000")), nil, 1, "fresh: no row task-3-implementer\n"},
		{"empty agent id", rows(rofFixture(1, key, "", 10*time.Second, "42000")), nil, 1, "fresh: no agent id\n"},
		{"agent id none", rows(rofFixture(1, key, "none", 10*time.Second, "42000")), nil, 1, "fresh: no agent id\n"},
		{"agent id inline", rows(rofFixture(1, key, "inline", 10*time.Second, "42000")), nil, 1, "fresh: no agent id\n"},
		{"row still open", rows(rofFixture(1, key, "a1", 0, "42000")), nil, 1, "fresh: row open\n"},
		{"cache lapsed", rows(rofFixture(1, key, "a1", 600*time.Second, "42000")), nil, 1, "fresh: age 600s >= 240s\n"},
		{"context not harvested", rows(rofFixture(1, key, "a1", 10*time.Second, "")), nil, 1, "fresh: context unknown\n"},
		{"context at the bound", rows(rofFixture(1, key, "a1", 10*time.Second, "100000")), nil, 1, "fresh: context 100000 >= 100000\n"},
		{"highest seq wins", rows(rofFixture(1, key, "old", 600*time.Second, "42000"), rofFixture(2, key, "new", 10*time.Second, "42000")), nil, 0, "resume new\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			code, out := rofRun(t, c.rows, c.err, t.TempDir(), "demo", key)
			if code != c.code || out != c.out {
				t.Fatalf("got exit %d %q, want exit %d %q", code, out, c.code, c.out)
			}
		})
	}
}

func TestResumeOrFreshUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"wt", "demo"}, {"wt", "demo", "k", "extra"}} {
		if code, _ := rofRun(t, []byte("[]"), nil, args...); code != 2 {
			t.Errorf("args %q: got exit %d, want 2", args, code)
		}
	}
}
