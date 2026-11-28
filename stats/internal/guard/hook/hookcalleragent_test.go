package hook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

func TestCallerAgentCommand(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ in, want string }{
		{"flow stage begin -command '/f' -stage flow.plan x",
			"flow -caller-agent a7 stage begin -command '/f' -stage flow.plan x"},
		{"cd /w && flow record dispatch end -key k; /a/scripts/flow-guard integrate-change prepare /m n",
			"cd /w && flow -caller-agent a7 record dispatch end -key k; /a/scripts/flow-guard integrate-change -caller-agent a7 prepare /m n"},
		{"PATH=/s/bin:$PATH flow-guard verify-change /m n | tail -1",
			"PATH=/s/bin:$PATH flow-guard verify-change -caller-agent a7 /m n | tail -1"},
		{"(flow state list)\nflow x", "(flow -caller-agent a7 state list)\nflow -caller-agent a7 x"},
		// Not a call: an argument, a quoted string, a heredoc body, another program.
		{"echo flow x; grep 'flow stage' f; which flow-guard", "echo flow x; grep 'flow stage' f; which flow-guard"},
		{"flow record summary -file - <<'EOF'\nflow is data\nEOF\nflow y",
			"flow -caller-agent a7 record summary -file - <<'EOF'\nflow is data\nEOF\nflow -caller-agent a7 y"},
		{"flowd -x; ./flow-guardian", "flowd -x; ./flow-guardian"},
	} {
		if got := callerAgentCommand(c.in, "a7"); got != c.want {
			t.Errorf("callerAgentCommand(%q)\n got %q\nwant %q", c.in, got, c.want)
		}
	}
}

// block-unsafe-commands rewrites a subagent's flow call and nothing else: the main thread's call,
// a call carrying no flow, and an agent id that is no plain word pass untouched.
func TestHookBlockUnsafeCommandsTagsCallerAgent(t *testing.T) {
	t.Parallel()
	run := func(stdin string) string {
		var out, errb bytes.Buffer
		core.Registry["hook"]([]string{"block-unsafe-commands"}, core.Env{Getenv: func(string) string { return "" }, Dir: t.TempDir(), Stdin: strings.NewReader(stdin)}, &out, &errb)
		return out.String()
	}
	got := run(`{"agent_id":"a7","tool_name":"Bash","tool_input":{"command":"flow stage end x","description":"d"}}`)
	var v struct {
		H struct {
			Event    string         `json:"hookEventName"`
			Decision *string        `json:"permissionDecision"`
			Input    map[string]any `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Fatalf("output %q: %v", got, err)
	}
	if v.H.Event != "PreToolUse" || v.H.Decision != nil || v.H.Input["command"] != "flow -caller-agent a7 stage end x" || v.H.Input["description"] != "d" {
		t.Errorf("output = %s", got)
	}
	for _, in := range []string{
		`{"tool_name":"Bash","tool_input":{"command":"flow stage end x"}}`,
		`{"agent_id":"a7","tool_name":"Bash","tool_input":{"command":"ls"}}`,
		`{"agent_id":"a7; rm","tool_name":"Bash","tool_input":{"command":"flow stage end x"}}`,
	} {
		if got := run(in); got != "" {
			t.Errorf("%s: output %q, want none", in, got)
		}
	}
}
