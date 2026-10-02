package lessons

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Render draws one Result as the Markdown answer the resolve route serves
// and the CLI prints verbatim — the record-render rule: the answer's shape
// is decided where it is produced, never constructed by the caller.
//
// Briefs render in full before any mention: the brief is the thing a
// caller acts on. Each mention renders its matching lines and the path
// holding them. Unreadable roots lead the notes, so a half-served answer
// says so itself.
func Render(r *Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Lessons: %s\n\n", r.Topic)
	fmt.Fprintf(&b, "found: %d briefs, %d narrative mentions\n", len(r.Briefs), len(r.Mentions))
	for _, path := range r.Unreadable {
		fmt.Fprintf(&b, "note: %s could not be read — its briefs and narratives are reported absent for that reason, not because none exist\n", path)
	}
	for _, brief := range r.Briefs {
		fmt.Fprintf(&b, "\n## brief: %s\n\n%s\n", displayPath(brief.Repo, brief.Path), brief.Content)
	}
	for _, m := range r.Mentions {
		fmt.Fprintf(&b, "\n## narrative: %s\n\n", displayPath(m.Repo, m.Path))
		for _, line := range m.Lines {
			fmt.Fprintf(&b, "- %s\n", line)
		}
	}
	return b.String()
}

// displayPath labels a source as <repository basename>/<repo-relative
// path> — enough for a caller on this machine to open it, without
// printing one checkout's absolute path as if it were the only one.
func displayPath(repo, rel string) string {
	return filepath.Join(filepath.Base(repo), rel)
}
