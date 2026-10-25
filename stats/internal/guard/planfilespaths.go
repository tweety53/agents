package guard

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// planFilesPathsCmd is `flow-guard plan-files-paths <tasks.md>`, run by
// check-plan-shape.sh beside plan-script-paths on every tasks.md it scans
// (decision files-path-check-own-command). It refuses a task's `**Files:**`
// path whose directory the repository does not carry while the same
// directory with one segment dropped does — a doubled or stray segment, as
// kan-1013's `src/app/app/src/main/resources/db/migration/…` for
// `src/app/src/main/resources/db/migration/` (decision
// files-path-dropped-segment). A path that exists, or a new file in an
// existing directory, passes; so does a new package or module, since no
// segment-dropped alternative of at least two segments exists for it. A
// token carrying `<`, `*`, `{` or `$` is a placeholder. The repository is
// the one whose `<spec-root>/changes/<name>/tasks.md` this is; an archived
// plan, or a file outside a `changes/<name>/` directory, is not checked.
// Exit 0 clean, 1 one `file:line: task <id> names … ` line per such path, 2 a
// usage error or a file that cannot be read.
func init() { Registry["plan-files-paths"] = planFilesPathsCmd }

func planFilesPathsCmd(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: flow-guard plan-files-paths <tasks.md>")
		return 2
	}
	file := args[0]
	if !filepath.IsAbs(file) {
		file = filepath.Join(env.Dir, file)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		return 2
	}
	lines, err := tcfPlanLines(b)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		return 2
	}
	changes := filepath.Dir(filepath.Dir(file))
	if filepath.Base(changes) != "changes" {
		return 0
	}
	root := filepath.Dir(filepath.Dir(changes))
	isDir := func(p string) bool {
		st, err := os.Stat(filepath.Join(root, p))
		return err == nil && st.IsDir()
	}
	found := 0
	for _, task := range tcfIterTasks(lines) {
		t, ok := tcfParseTask(lines, task.id)
		if !ok {
			continue
		}
		for _, p := range t.files {
			if strings.ContainsAny(p, "<*{$") {
				continue
			}
			if _, err := os.Lstat(filepath.Join(root, p)); err == nil {
				continue
			}
			dir := path.Dir(p)
			if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
				continue
			}
			segs := strings.Split(dir, "/")
			for i := 0; i < len(segs)-1; i++ {
				alt := append(append([]string{}, segs[:i]...), segs[i+1:]...)
				if len(alt) < 2 || !isDir(strings.Join(alt, "/")) {
					continue
				}
				fmt.Fprintf(stdout, "%s:%d: task %s names `%s` in a directory the repository does not carry; `%s` exists — name the path the file lands in (decision files-path-dropped-segment)\n",
					args[0], psFieldLine(task.lines, task.taskLine, "Files"), task.id, p, strings.Join(alt, "/"))
				found++
				break
			}
		}
	}
	if found > 0 {
		return 1
	}
	return 0
}
