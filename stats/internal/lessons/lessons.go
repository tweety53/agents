// Package lessons resolves a process lesson — a durable brief or the
// archived session narratives mentioning it — across every project the
// store knows. It is the resolver side of the lessons home: a ticket that
// references a practice names it once and this package finds where the
// practice is written down, instead of the caller searching repositories.
//
// Everything here is a plain filesystem read of docs/briefs/ over the roots
// the caller supplies -- never git history, since a change's directory is
// never committed (KAN-1073) -- resolved fresh on every call — the same "the daemon reads the
// repositories" shape the self-review bundle uses. An index that
// materialized briefs or narrative mentions would go stale the moment a
// brief was edited outside it, which is the exact unreachability this
// resolver exists to remove, so no lesson state is stored anywhere.
package lessons

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Root is one project's main checkout the scan reads. It mirrors the
// store's ProjectRoot without importing it — the caller (the api handler)
// holds both types and translates, so this package stays a pure scanner.
// Only the path is carried: the scan matches and ranks on paths alone.
type Root struct {
	Path string
}

// Brief is one durable brief: a Markdown file in a project's docs/briefs/
// directory, the lessons home. Slug is the file's stem — the name a ticket
// references ("the flake-bisection brief"); Title is the file's first
// heading line with its marker stripped.
type Brief struct {
	Repo    string
	Path    string
	Slug    string
	Title   string
	Content string
}

// Mention is one archived narrative whose content matches the topic.
// Change is the archived change's directory name; Lines are the matching
// lines, capped per narrative at matchLineCap, in file order.
type Mention struct {
	Repo    string
	Change  string
	Path    string
	Lines   []string
	ModTime time.Time
}

// Result is one Resolve call's answer: the briefs and narrative mentions
// matching the topic, ranked — briefs first, then mentions newest-first —
// plus the roots, or the individual files inside a readable root, that
// could not be read, which are an environmental failure a caller must see
// rather than an absence a caller could mistake for "no lesson exists".
type Result struct {
	Topic      string
	Briefs     []Brief
	Mentions   []Mention
	Unreadable []string
}

// briefsDir is the lessons home, relative to a project root — the one
// known location the convention names. Docs are repo content, so a brief
// lands with the change that promotes it and survives every cleanup.
const briefsDir = "docs/briefs"

// Normalize folds the forms a topic and a text may take — case, hyphens,
// underscores and runs of space — into one comparable shape, so
// "flake-bisection" resolves the file "flake_bisection.md" and a line
// about "flake bisection" alike.
func Normalize(s string) string {
	replacer := strings.NewReplacer("-", " ", "_", " ")
	return strings.Join(strings.Fields(strings.ToLower(replacer.Replace(s))), " ")
}

// Resolve scans every root's docs/briefs/ for briefs matching topic;
// Mentions stays empty, since no narrative is committed to read. A root that cannot be read at all is reported in Result's
// Unreadable, never an error: one broken checkout must not silence the
// others' answers. A single brief inside a readable root
// whose read fails is reported there the same way — a permissions edge or
// a mid-write file is environmental, not evidence nothing matches. A root
// without a docs/briefs/ simply contributes
// nothing — legitimate absence, the same rule the self-review bundle's
// sources follow.
func Resolve(roots []Root, topic string) (*Result, error) {
	needle := Normalize(topic)
	if needle == "" {
		return nil, fmt.Errorf("lessons: topic is empty after normalization")
	}

	res := &Result{Topic: topic}
	for _, root := range roots {
		if _, err := os.Stat(root.Path); err != nil {
			res.Unreadable = append(res.Unreadable, root.Path)
			continue
		}

		briefs, _ := filepath.Glob(filepath.Join(root.Path, briefsDir, "*.md"))
		for _, path := range briefs {
			b, matched, err := readBrief(root, path, needle)
			if err != nil {
				res.Unreadable = append(res.Unreadable, path)
				continue
			}
			if matched {
				res.Briefs = append(res.Briefs, b)
			}
		}
	}

	sort.Slice(res.Briefs, func(i, j int) bool {
		if res.Briefs[i].Repo != res.Briefs[j].Repo {
			return res.Briefs[i].Repo < res.Briefs[j].Repo
		}
		return res.Briefs[i].Slug < res.Briefs[j].Slug
	})
	return res, nil
}

// readBrief reads one brief file and reports whether it matches the
// needle. The whole file is the match surface — slug, title and content
// alike — because a brief is short and a topic may name what a brief only
// discusses in its body. A read failure is an error; no match is a nil
// error with matched false.
func readBrief(root Root, path, needle string) (Brief, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Brief{}, false, err
	}
	b := Brief{
		Repo:    root.Path,
		Path:    relTo(root.Path, path),
		Slug:    strings.TrimSuffix(filepath.Base(path), ".md"),
		Title:   firstHeading(string(content)),
		Content: strings.TrimRight(string(content), "\n"),
	}
	if !strings.Contains(Normalize(b.Slug+" "+b.Title+" "+b.Content), needle) {
		return Brief{}, false, nil
	}
	return b, true, nil
}

// firstHeading returns the file's first Markdown heading line with its
// marker and space stripped, or "" when the file opens with none.
func firstHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	return ""
}

// relTo renders path relative to root, falling back to the path itself
// when it is not beneath root (a symlinked root, say) — the label stays
// readable either way.
func relTo(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}
