// flowcli.go holds the flow CLI helpers several guards share.

package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// pcFlow runs the flow CLI from env's PATH in env.Dir: stdout (with stderr
// folded in when combined, as `2>&1` does), stderr, and the exit code --
// 127 with bash's "command not found" when there is no flow to run.
func pcFlow(env Env, combined bool, args ...string) ([]byte, []byte, int) {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return []byte(pcNotFound), []byte(pcNotFound + "\n"), pcNotFoundC
	}
	cmd := exec.Command(flow, args...)
	cmd.Dir = env.Dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if combined {
		cmd.Stderr = &out
	}
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		return out.Bytes(), errb.Bytes(), 0
	case errors.As(err, &ee):
		return out.Bytes(), errb.Bytes(), rrExitCode(ee.ProcessState)
	default:
		return out.Bytes(), []byte(err.Error() + "\n"), 126
	}
}

// pcRaw is `jq -r` of one value: a string's text, null (or absent) as
// "null", anything else as compact JSON.
func pcRaw(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "null"
	}
	var s string
	if pcIsString(raw) && json.Unmarshal(raw, &s) == nil {
		return s
	}
	var c bytes.Buffer
	if json.Compact(&c, raw) != nil {
		return string(raw)
	}
	return c.String()
}

// pfdRead is one `flow record <verb> -change <name> -C <worktree>` read:
// the hook when set, else the CLI on PATH, its stderr discarded as the
// bash's 2>/dev/null did.
func pfdRead(env Env, hook func(string) ([]byte, error), verb, name, worktree string) ([]byte, error) {
	if hook != nil {
		return hook(name)
	}
	out, _, rc := pcFlow(env, false, "record", verb, "-change", name, "-C", worktree)
	if rc != 0 {
		return nil, fmt.Errorf("flow exit %d", rc)
	}
	return out, nil
}

// pfdStream decodes a stream of JSON values, jq's input: any number of
// them, none included.
func pfdStream(raw []byte) ([]any, bool) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var vals []any
	for {
		var v any
		if err := dec.Decode(&v); err == io.EOF {
			return vals, true
		} else if err != nil {
			return nil, false
		}
		vals = append(vals, v)
	}
}

const (
	pcNotFound  = "flow: command not found"
	pcNotFoundC = 127
)
