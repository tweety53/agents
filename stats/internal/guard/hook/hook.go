package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/records"
)

// flow-guard hook <name> — run the Claude Code hook <name> on the event payload on stdin. Each
// hook is registered on ~/.claude/hooks/flow-guard-hook, the binary `setup.sh global` builds from
// stats/cmd/flow-guard-hook (install_hook_binary), as `"$HOME/.claude/hooks/flow-guard-hook" hook <name> || true`. The hooks
// and the guards share one set of patterns and one state reader. A hook's contract is the leading comment of its own file
// (hook<name>.go); every hook writes its JSON with hookJSON and exits 0: a hook fault, a panic
// included, lets the event through. flow-guard records no guard-run telemetry for a hook, so a hook
// makes no flowd call its contract does not name; with Env.HookRuns set (flow-guard's and flow-guard-hook's main set
// it to records.HookRunsPath) each run appends its wall time to that local file instead.
//
// Exit 2 on a missing or unknown hook name (usage).
func init() { core.Registry["hook"] = hook }

// hooks maps a hook's name to its body: the decoded payload (nil when stdin is no JSON object)
// and the raw stdin.
var hooks = map[string]func(env core.Env, payload map[string]any, stdout, stderr io.Writer){}

func hook(args []string, env core.Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || hooks[args[0]] == nil {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard hook <name>")
		return 2
	}
	if env.HookRuns != "" {
		start := time.Now()
		defer appendHookRun(env.HookRuns, args[0], start) // runs after the recover below
	}
	defer func() { _ = recover() }() // fail open: a hook fault never blocks work
	var raw []byte
	if env.Stdin != nil {
		raw, _ = io.ReadAll(env.Stdin)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var payload map[string]any
	var v any
	if dec.Decode(&v) == nil && !dec.More() {
		payload, _ = v.(map[string]any)
	}
	hooks[args[0]](env, payload, stdout, stderr)
	return 0
}

// appendHookRun appends hook's run, start to now, to path as one records.HookRun line
// (records.AppendJSONLine).
func appendHookRun(path, hook string, start time.Time) {
	now := time.Now()
	records.AppendJSONLine(path, records.HookRun{Hook: hook, DurationUS: now.Sub(start).Microseconds(), At: now.UTC()})
}

// hookJSON is v as one line of JSON, HTML characters unescaped so a raw transcript reads as the
// reason was written.
func hookJSON(v any) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// pyTruthy is Python's bool(v) for a decoded JSON value.
func pyTruthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case json.Number:
		f, err := x.Float64()
		return err != nil || f != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	}
	return true
}

// pyStr is Python's str(v) for a decoded JSON value's scalar.
func pyStr(v any) string {
	switch x := v.(type) {
	case nil:
		return "None"
	case bool:
		if x {
			return "True"
		}
		return "False"
	case string:
		return x
	case json.Number:
		return x.String()
	}
	return fmt.Sprint(v)
}

// pyStrOr is Python's str(v or alt).
func pyStrOr(v any, alt string) string {
	if pyTruthy(v) {
		return pyStr(v)
	}
	return alt
}

// strOf is v when it is a string, else "".
func strOf(v any) string {
	s, _ := v.(string)
	return s
}

// pyWords is Python's s.split(): the whitespace-separated words.
func pyWords(v any) []string {
	s, _ := v.(string)
	return strings.Fields(s)
}

// pyQuoteURL is Python's urllib.parse.quote(s, safe): every byte outside letters, digits, `_.-~`
// and safe percent-encoded.
func pyQuoteURL(s, safe string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte("_.-~"+safe, c) >= 0 {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// pyFirstLine is Python's (s.strip().splitlines() or [s])[0][:n], n counted in characters.
func pyFirstLine(s string, n int) string {
	t := strings.TrimSpace(s)
	line := s
	if t != "" {
		line = pySplitFirst(t)
	}
	if utf8.RuneCountInString(line) > n {
		line = string([]rune(line)[:n])
	}
	return line
}

// pySplitFirst is Python's (s.splitlines() or [""])[0]: s up to its first line boundary.
func pySplitFirst(s string) string {
	if i := strings.IndexFunc(s, func(r rune) bool {
		return strings.ContainsRune("\n\r\v\f\x1c\x1d\x1e\u0085\u2028\u2029", r)
	}); i >= 0 {
		return s[:i]
	}
	return s
}

// hookBound is a hook's wall-clock bound on a store or subprocess call: d, or 0 (none) under
// FLOW_HOOK_UNBOUNDED=1, which the hooks' tests set so a deny case never fails open because a
// starved stub outran the bound — the bound is production's fail-open, not under test.
func hookBound(env core.Env, d time.Duration) time.Duration {
	if env.Getenv("FLOW_HOOK_UNBOUNDED") == "1" {
		return 0
	}
	return d
}

// hookRun runs name (a path, else looked up on env's PATH) with args in dir ("" is the hook's
// own), returning its stdout and exit code; err when it could not start or outran bound.
func hookRun(env core.Env, bound time.Duration, dir, name string, args ...string) ([]byte, int, error) {
	path := name
	if !strings.Contains(name, "/") {
		var ok bool
		if path, ok = core.LookPath(env, name); !ok {
			return nil, 0, errors.New(name + ": not found")
		}
	}
	ctx := context.Background()
	if bound > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, bound)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	if env.Environ != nil {
		cmd.Env = env.Environ()
	}
	cmd.WaitDelay = time.Second
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil, 0, ctx.Err()
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return out.Bytes(), ee.ExitCode(), nil
	}
	return out.Bytes(), 0, err
}

// hookGetJSON GETs url and decodes its JSON body; err on a transport failure, a non-2xx status or
// a body that is no JSON.
func hookGetJSON(env core.Env, url string, bound time.Duration) (any, error) {
	resp, err := (&http.Client{Timeout: bound, Transport: &http.Transport{}}).Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	var v any
	return v, dec.Decode(&v)
}

// hookRecordIncident writes one `flow record incident` row for a deny, so a self-review pass can
// count the denials; FLOW_HOOK_INCIDENT_CMD overrides `flow` for the hooks' own harness only.
// Every failure is swallowed: the deny already written stands either way.
func hookRecordIncident(env core.Env, guardName, reason, cwd, change string) {
	cmd := env.Getenv("FLOW_HOOK_INCIDENT_CMD")
	if cmd == "" {
		cmd = "flow"
	}
	args := []string{"record", "incident", "-guard", guardName, "-symptom", pyFirstLine(reason, 300),
		"-recovery", "re-sent per the deny reason", "-minutes-lost", "0"}
	if cwd != "" {
		args = append(args, "-C", cwd)
	}
	if change != "" {
		args = append(args, "-change", change)
	}
	_, _, _ = hookRun(env, hookBound(env, 10*time.Second), "", cmd, args...)
}

// hookDenyJSON is a PreToolUse deny decision.
func hookDenyJSON(reason string) string {
	return hookJSON(map[string]any{"hookSpecificOutput": map[string]any{
		"hookEventName":            "PreToolUse",
		"permissionDecision":       "deny",
		"permissionDecisionReason": reason,
	}})
}

// hookInProcess runs guard fn with args in dir as its own `scripts/flow-guard` process did, its
// stdout, stderr and exit code returned -- bounded: past bound (0 is none) it returns
// context.DeadlineExceeded and leaves fn running, so a hung call never hangs the hook; the hook's
// process exiting ends it.
func hookInProcess(env core.Env, bound time.Duration, dir string, fn core.Func, args ...string) (stdout, stderr string, code int, err error) {
	type result struct {
		out, errb string
		code      int
	}
	if dir != "" {
		env.Dir = dir
	}
	done := make(chan result, 1)
	go func() {
		var out, errb bytes.Buffer
		c := fn(args, env, &out, &errb)
		done <- result{out.String(), errb.String(), c}
	}()
	var timeout <-chan time.Time
	if bound > 0 {
		timer := time.NewTimer(bound)
		defer timer.Stop()
		timeout = timer.C
	}
	select {
	case r := <-done:
		return r.out, r.errb, r.code, nil
	case <-timeout:
		return "", "", 0, context.DeadlineExceeded
	}
}
