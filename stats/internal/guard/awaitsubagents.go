package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// flow-guard await-sub-agents — wait, at most 110 seconds (inside the Bash tool's
// default 2-minute timeout), for background agents to
// finish, so a subagent waiting on its own agents keeps its 5-minute prompt
// cache warm: each call returns inside the cache window, and the next call is
// a cache read where ending the turn and waking on the hand-back past 5
// minutes rewrites the whole context (kan-1061's planner: 206k tokens
// rewritten after a 13-minute wait on its Plan agents).
//
// Usage: flow-guard await-sub-agents <agent id>...
//
// An agent is finished when its subagents/agent-<id>.meta.json, under any
// session of $CLAUDE_PROJECTS_DIR (default ~/.claude/projects), carries
// "finished": true. An unfinished agent whose transcript, agent-<id>.jsonl
// beside its meta.json, was last written more than 30 minutes ago is taken as
// ended -- the rule and figure of hooks/await-own-agents.py's STALE_SECONDS --
// and named on a `stale: <id>...` stdout line; a missing transcript is not
// stale. Exit 0 once every named agent is finished or stale; exit 1 at the
// cap with a line naming those still running, after the `stale:` line when
// any agent went stale — call again; exit 2 on no
// argument, an agent id with no meta.json, or a cap
// ($AWAIT_SUB_AGENTS_SECONDS) above 240.

func init() {
	Registry["await-sub-agents"] = awaitSubAgents
}

var (
	asaCapRe      = regexp.MustCompile(`^[0-9]+$`)
	asaFinishedRe = regexp.MustCompile(`"finished"\s*:\s*true`)
)

// asaStaleAfter is hooks/await-own-agents.py's STALE_SECONDS.
const asaStaleAfter = 30 * time.Minute

func awaitSubAgents(args []string, env Env, stdout, stderr io.Writer) int {
	root := env.Getenv("CLAUDE_PROJECTS_DIR")
	if root == "" {
		root = filepath.Join(env.Getenv("HOME"), ".claude", "projects")
	}
	capS := env.Getenv("AWAIT_SUB_AGENTS_SECONDS")
	if capS == "" {
		capS = "110"
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: flow-guard await-sub-agents <agent id>...")
		return 2
	}
	capN, err := strconv.Atoi(capS)
	if !asaCapRe.MatchString(capS) || err != nil || capN > 240 {
		fmt.Fprintf(stderr, "await-sub-agents: AWAIT_SUB_AGENTS_SECONDS must be 0-240, got %s\n", capS)
		return 2
	}

	metas := make([]string, len(args))
	for i, id := range args {
		metas[i] = asaFindMeta(root, "agent-"+id+".meta.json")
		if metas[i] == "" {
			fmt.Fprintf(stderr, "await-sub-agents: no meta.json for agent %s under %s\n", id, root)
			return 2
		}
	}

	deadline := time.Now().Add(time.Duration(capN) * time.Second)
	for {
		var running, stale []string
		for i, m := range metas {
			b, _ := os.ReadFile(m)
			switch {
			case asaFinishedRe.Match(b):
			case asaStale(strings.TrimSuffix(m, ".meta.json") + ".jsonl"):
				stale = append(stale, args[i])
			default:
				running = append(running, args[i])
			}
		}
		done := len(running) == 0
		if done || !time.Now().Before(deadline) {
			if len(stale) > 0 {
				fmt.Fprintln(stdout, "stale: "+strings.Join(stale, " "))
			}
			if done {
				return 0
			}
			fmt.Fprintln(stdout, "still running: "+strings.Join(running, " "))
			return 1
		}
		time.Sleep(2 * time.Second)
	}
}

// asaStale reports whether transcript exists and was last written more than
// asaStaleAfter ago.
func asaStale(transcript string) bool {
	fi, err := os.Stat(transcript)
	return err == nil && time.Since(fi.ModTime()) > asaStaleAfter
}

// asaFindMeta is `find root -path "*/subagents/<name>" -print -quit`: the
// first match, or "" when none (an unreadable root included).
func asaFindMeta(root, name string) string {
	found := ""
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == name && filepath.Base(filepath.Dir(p)) == "subagents" {
			found = p
			return fs.SkipAll
		}
		return nil
	})
	return found
}
