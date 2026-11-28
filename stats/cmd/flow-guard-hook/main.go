// Command flow-guard-hook runs the Claude Code hooks, called as
// `flow-guard-hook hook <name>` — the one call every registration `setup.sh global`
// prints makes, `"$HOME/.claude/hooks/flow-guard-hook" hook <name> || true`
// (install_hook_binary). It is cmd/flow-guard's `hook` dispatch alone, linking
// guard/core and guard/hook and no other guard family, so the start-up every
// tool call pays carries no guard a hook never runs
// (TestFlowGuardHookLinksOnlyTheHookFamily, stats/internal/repotest). The hook
// contract — the payload on stdin, the JSON on stdout, exit 0 on any hook
// fault, exit 2 on a missing or unknown hook name — is the leading comment of
// stats/internal/guard/hook/hook.go.
//
// Exit 2 on any first argument but `hook`, or none: a usage error.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tweety53/agents/stats/internal/guard/core"
	_ "github.com/tweety53/agents/stats/internal/guard/hook" // registers core.Registry["hook"]
	"github.com/tweety53/agents/stats/internal/records"
)

// builtRepoRoot is the checkout setup.sh global built this binary from, set
// with -ldflags -X; it stands in for scripts/flow-guard's export of
// FLOW_GUARD_REPO_ROOT, over any inherited value as there.
var builtRepoRoot string

func main() {
	if builtRepoRoot != "" {
		_ = os.Setenv("FLOW_GUARD_REPO_ROOT", builtRepoRoot)
	}
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run dispatches args, `hook <name>`, to guard/hook with stdin as the payload.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "hook" {
		fmt.Fprint(stderr, "usage: flow-guard-hook hook <name>\n")
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "flow-guard-hook: cannot read the working directory: %v\n", err)
		return core.CannotAnswer("hook")
	}
	env := core.Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: cwd, Stdin: stdin,
		Reasons: &core.Reasons{}, SlotGate: true, HookRuns: records.HookRunsPath(os.Getenv)}
	return core.Registry["hook"](args[1:], env, stdout, stderr)
}
