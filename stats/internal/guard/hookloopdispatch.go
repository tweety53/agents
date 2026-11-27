package guard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// flow-guard hook enforce-loop-dispatch — PreToolUse: record a pipeline-fix or self-review
// fix-loop dispatch's row. It never denies.
//
// A loop dispatch is an Agent/Task call whose prompt carries a `Key: pipeline-fix-<k>-...` or
// `Key: self-review-...` line (skills/flow-contracts/pipeline.md **Pipeline defects found
// mid-run**, skills/flow-self-review/SKILL.md step 3). This hook runs check-loop-dispatch in
// process for every such call, against the agents checkout it is installed from; it opens the
// dispatch's row in flowd -- the row the per-role cost stats read. The checker runs in the
// payload's `cwd`, so a self-review on a change's own branch in another project files its rows in
// that project's flowd project, not the agents checkout's.
//
// Recording is best-effort and never blocks: a checker that cannot answer, cannot run or times out
// is logged on stderr and the Agent call proceeds. A loop dispatch whose prompt lacks the
// agent-baseline pointer is not recorded, since enforce-agent-baseline denies it after, which
// would leave its row open. Anything that is not a loop dispatch -- unparseable input, another
// tool, a prompt with no valid loop key -- passes untouched.
//
// The checker's wait stays well under the 300s timeout setup.sh registers.
func init() { hooks["enforce-loop-dispatch"] = enforceLoopDispatch }

var (
	// Any Key line whose value starts as a loop key; the value must then be a whole eldValidKey.
	eldKey      = regexp.MustCompile(`(?m)^[ \t]*Key:[ \t]*((?:pipeline-fix|self-review)-.*?)[ \t]*$`)
	eldValidKey = regexp.MustCompile(`^(?:pipeline-fix-[0-9]+-\S+|self-review-\S+-\S+)$`)
	// A prompt's `Role:` line: the allowed-pairs role the row is counted under in Role costs.
	eldRole = regexp.MustCompile(`(?m)^Role:[ \t]*(\S+)[ \t]*$`)
)

// Under setup.sh's registered 300s.
const eldCheckTimeout = 240 * time.Second

func enforceLoopDispatch(env Env, payload map[string]any, _, stderr io.Writer) {
	if tool := payload["tool_name"]; tool != "Agent" && tool != "Task" {
		return
	}
	in, _ := payload["tool_input"].(map[string]any)
	prompt, ok := in["prompt"].(string)
	if !ok {
		return
	}
	cwd := strOf(payload["cwd"])
	if st, err := os.Stat(cwd); cwd == "" || err != nil || !st.IsDir() {
		cwd = ""
	}
	m := eldKey.FindStringSubmatch(prompt)
	if m == nil || !eldValidKey.MatchString(m[1]) || !strings.Contains(prompt, "~/.claude/rules/agent-baseline.md") {
		return
	}
	key := m[1]
	role := ""
	if r := eldRole.FindStringSubmatch(prompt); r != nil {
		role = r[1]
	}
	repo := env.Getenv("FLOW_GUARD_REPO_ROOT")
	args := []string{repo, key, pyStrOr(in["subagent_type"], ""), pyStrOr(in["model"], ""), role}
	// LOOP_DISPATCH_CHECK swaps in a checker process for the hook's own harness only.
	if c := env.Getenv("LOOP_DISPATCH_CHECK"); c != "" {
		eldRunCheck(env, c, cwd, key, args, stderr)
		return
	}
	out, errb, code, err := hookInProcess(env, hookBound(env, eldCheckTimeout), cwd, checkLoopDispatch, args...)
	switch {
	case err != nil:
		fmt.Fprintf(stderr, "enforce-loop-dispatch: %s not recorded: flow-guard check-loop-dispatch could not run (%v)\n", key, err)
	case code != 0:
		fmt.Fprintf(stderr, "enforce-loop-dispatch: %s not recorded (flow-guard check-loop-dispatch exit %d):\n%s%s\n", key, code, out, errb)
	}
}

// eldRunCheck runs the harness's checker process check with args in cwd, logging as the
// in-process call does.
func eldRunCheck(env Env, check, cwd, key string, args []string, stderr io.Writer) {
	ctx := context.Background()
	if b := hookBound(env, eldCheckTimeout); b > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, check, args...)
	cmd.Dir = cwd
	if env.Environ != nil {
		cmd.Env = env.Environ()
	}
	cmd.WaitDelay = time.Second
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case ctx.Err() != nil:
		fmt.Fprintf(stderr, "enforce-loop-dispatch: %s not recorded: flow-guard check-loop-dispatch could not run (%v)\n", key, ctx.Err())
	case err == nil:
	case errors.As(err, &ee):
		fmt.Fprintf(stderr, "enforce-loop-dispatch: %s not recorded (flow-guard check-loop-dispatch exit %d):\n%s%s\n", key, ee.ExitCode(), out.String(), errb.String())
	default:
		fmt.Fprintf(stderr, "enforce-loop-dispatch: %s not recorded: flow-guard check-loop-dispatch could not run (%v)\n", key, err)
	}
}
