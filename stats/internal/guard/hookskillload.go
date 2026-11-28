package guard

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// flow-guard hook flow-skill-load — tell flowd which instruction file a session just loaded, so
// `flow-guard prune-candidates` can list the skill files no session loads.
//
// On PostToolUse it posts one POST /api/v1/skill-loads {"path", "sessionId"} when the call loaded a
// file under skills/, rules/ or docs/ of an agents checkout — a directory holding
// stats/internal/guard, the main checkout or a worktree of it:
//
//	Read  <file_path>  the file, symlinks resolved, so ~/.claude/skills/... counts as the repo's
//	Skill <skill>      $HOME/.claude/skills/<skill>/SKILL.md, resolved the same way; a plugin's
//	                   skill (a name carrying ':') never resolves into the repo and posts nothing
//
// path is relative to that checkout's root. A Read of only part of a file counts as a load of the
// whole file: the hook does not name sections. Any other tool, path or event posts nothing.
//
// Observe-only, like flow-session-activity: it prints nothing and exits 0, and an unreachable or
// refusing flowd costs the one load row — nothing is journalled for a replay, for the reason
// cmd/flow-guard/telemetry.go gives for guard runs.
func init() { hooks["flow-skill-load"] = flowSkillLoad }

func flowSkillLoad(env Env, payload map[string]any, _, _ io.Writer) {
	in, _ := payload["tool_input"].(map[string]any)
	var path string
	switch payload["tool_name"] {
	case "Read":
		path = strOf(in["file_path"])
	case "Skill":
		if s := strOf(in["skill"]); s != "" && !strings.ContainsAny(s, ":/") {
			path = filepath.Join(env.Getenv("HOME"), ".claude", "skills", s, "SKILL.md")
		}
	}
	rel := agentsInstructionPath(path)
	if rel == "" {
		return
	}
	body := hookJSON(map[string]any{"path": rel, "sessionId": strOf(payload["session_id"])})
	resp, err := (&http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{}}).Post(
		hookFlowAddr(env)+"/api/v1/skill-loads", "application/json", strings.NewReader(body))
	if err == nil {
		_ = resp.Body.Close()
	}
}

// agentsInstructionPath is path, symlinks resolved, relative to the agents checkout it lies in when
// its first segment there is skills, rules or docs; "" for any other path or one that does not
// resolve.
func agentsInstructionPath(path string) string {
	if !filepath.IsAbs(path) {
		return ""
	}
	p, err := filepath.EvalSymlinks(path)
	if err != nil {
		return ""
	}
	parts := strings.Split(p, string(filepath.Separator))
	for i := 1; i < len(parts)-1; i++ {
		switch parts[i] {
		case "skills", "rules", "docs":
			root := string(filepath.Separator) + filepath.Join(parts[:i]...)
			if st, err := os.Stat(filepath.Join(root, "stats", "internal", "guard")); err == nil && st.IsDir() {
				return filepath.Join(parts[i:]...)
			}
		}
	}
	return ""
}
