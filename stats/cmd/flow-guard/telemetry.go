package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// telemetry.go records every guard this binary dispatches into flowd's
// guard_runs table (POST /api/v1/guard-runs/{project}), so the flow-health
// dashboard can answer which guards ever fire -- see
// stats/internal/store/migrations/0032_guard_runs.sql.
//
// It is strictly best-effort and never changes a guard's answer: a daemon
// that is down, slow, or refuses the write (an unknown project -- a guard
// run inside a test fixture's throwaway repository) is ignored, bounded by
// telemetryTimeout. Nothing is journalled for a later replay: a lost guard
// run costs one row of statistics, and a replay queue for it would be a
// second failure mode for every guard.
//
// It uses only the standard library, not internal/client or
// internal/fallback: scripts/lib/flow-guard.sh keys the cached binary on
// the sources of this package and internal/guard alone, so importing
// another package of this module would let the cache serve a binary built
// from stale copies of it.

// telemetryTimeout bounds the whole recording -- the git call resolving the
// project and the POST -- so a guard never waits on its own telemetry for
// longer than this.
const telemetryTimeout = 500 * time.Millisecond

// telemetryOff names FLOW_GUARD_TELEMETRY's opt-out values.
var telemetryOff = map[string]bool{"0": true, "off": true, "false": true, "no": true}

// guardRun is records.GuardRun's wire shape, spelled out here for the
// reason this file's header gives.
type guardRun struct {
	Guard      string    `json:"guard"`
	Worktree   string    `json:"worktree"`
	ExitCode   int       `json:"exitCode"`
	Outcome    string    `json:"outcome"`
	DurationMS int       `json:"durationMs"`
	RecordedAt time.Time `json:"recordedAt"`
}

// outcomeFor classifies guard name's exit code: 0 is clear, the guard's own
// could-not-answer code is cannot-answer, and anything else -- a violation,
// a refusal -- is the guard firing.
func outcomeFor(name string, code int) string {
	switch code {
	case 0:
		return "clear"
	case cannotAnswer(name):
		return "cannot-answer"
	default:
		return "fired"
	}
}

// recordsAddr is the daemon the run is recorded against: FLOW_RECORDS_ADDR,
// then FLOW_ADDR, then the dev daemon -- the order the flow CLI's record
// family resolves (cmd/flow/state.go's resolveRecordsAddr), so a guard in
// an isolated worktree records where that worktree's run records go, never
// into a workspace database its cleanup drops.
func recordsAddr(getenv func(string) string) string {
	for _, k := range []string{"FLOW_RECORDS_ADDR", "FLOW_ADDR"} {
		if v := getenv(k); v != "" {
			return strings.TrimRight(v, "/")
		}
	}
	return "http://127.0.0.1:4173"
}

// projectKey derives dir's project key exactly as
// stats/internal/fallback.ProjectKey does -- the main checkout's basename,
// a dash, and the first eight hex digits of the SHA-1 of its resolved path.
// telemetry_test.go pins the two together.
func projectKey(ctx context.Context, dir string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", err
	}
	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(dir, commonDir)
	}
	mainCheckout := filepath.Clean(filepath.Dir(commonDir))
	if resolved, err := filepath.EvalSymlinks(mainCheckout); err == nil {
		mainCheckout = resolved
	}
	sum := sha1.Sum([]byte(mainCheckout))
	return filepath.Base(mainCheckout) + "-" + hex.EncodeToString(sum[:])[:8], nil
}

// recordRun records one dispatched guard's run, and reports nothing: every
// failure is swallowed, per this file's header.
func recordRun(getenv func(string) string, dir, name string, code int, elapsed time.Duration, client *http.Client) {
	if telemetryOff[strings.ToLower(getenv("FLOW_GUARD_TELEMETRY"))] {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), telemetryTimeout)
	defer cancel()

	key, err := projectKey(ctx, dir)
	if err != nil {
		return
	}
	body, err := json.Marshal(guardRun{
		Guard: name, Worktree: dir, ExitCode: code, Outcome: outcomeFor(name, code),
		DurationMS: int(elapsed.Milliseconds()), RecordedAt: time.Now().UTC(),
	})
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		recordsAddr(getenv)+"/api/v1/guard-runs/"+url.PathEscape(key), bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// dispatched reports whether args name a registered guard -- the only runs
// recorded; a usage error is not a guard's answer.
func dispatched(args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	if _, ok := registry[args[0]]; !ok {
		return "", false
	}
	return args[0], true
}

// workingDir is the directory the guard ran in, or "" when it cannot be
// read (run itself then answered cannot-answer, and there is no project to
// record against).
func workingDir() string {
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	return d
}
