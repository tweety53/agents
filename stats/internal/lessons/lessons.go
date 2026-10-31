// Package lessons resolves a process lesson — a durable brief or the
// archived session narratives mentioning it — across every project the
// store knows. It is the resolver side of the lessons home: a ticket that
// references a practice names it once and this package finds where the
// practice is written down, instead of the caller searching repositories.
//
// Everything here is a plain filesystem and git-history read over the roots the caller
// supplies, resolved fresh on every call — the same "the daemon reads the
// repositories" shape the self-review bundle uses. An index that
// materialized briefs or narrative mentions would go stale the moment a
// brief was edited outside it, which is the exact unreachability this
// resolver exists to remove, so no lesson state is stored anywhere.
package lessons

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

// landedSubject matches the commit that removed a landed change's directory
// from the tip -- a `drop` now, an `archive` move before -- whose history
// holds the change's narrative.md (landed.File).
var landedSubject = regexp.MustCompile(`^chore\(spectre\): (?:drop|archive) (\S+)$`)

// matchLineCap bounds one narrative's printed matching lines. A narrative
// that discusses the topic at length still answers "where is it written",
// not "quote everything"; the caller can read the file, whose path is
// printed beside the lines.
const matchLineCap = 5

// Normalize folds the forms a topic and a text may take — case, hyphens,
// underscores and runs of space — into one comparable shape, so
// "flake-bisection" resolves the file "flake_bisection.md" and a line
// about "flake bisection" alike.
func Normalize(s string) string {
	replacer := strings.NewReplacer("-", " ", "_", " ")
	return strings.Join(strings.Fields(strings.ToLower(replacer.Replace(s))), " ")
}

// Resolve scans every root for briefs and archived narratives matching
// topic. A root that cannot be read at all is reported in Result's
// Unreadable, never an error: one broken checkout must not silence the
// others' answers. A single brief or narrative inside a readable root
// whose read fails is reported there the same way — a permissions edge or
// a mid-write file is environmental, not evidence nothing matches. A root
// without a docs/briefs/ or without an archive simply contributes
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

		cs, err := landedChanges(root.Path)
		if err != nil {
			res.Unreadable = append(res.Unreadable, root.Path)
		}
		for _, c := range cs {
			m, matched, err := readNarrative(root, c, needle)
			if err != nil {
				res.Unreadable = append(res.Unreadable, root.Path+": "+c.name)
				continue
			}
			if matched {
				res.Mentions = append(res.Mentions, m)
			}
		}
	}

	sort.Slice(res.Briefs, func(i, j int) bool {
		if res.Briefs[i].Repo != res.Briefs[j].Repo {
			return res.Briefs[i].Repo < res.Briefs[j].Repo
		}
		return res.Briefs[i].Slug < res.Briefs[j].Slug
	})
	sort.SliceStable(res.Mentions, func(i, j int) bool {
		if !res.Mentions[i].ModTime.Equal(res.Mentions[j].ModTime) {
			return res.Mentions[i].ModTime.After(res.Mentions[j].ModTime)
		}
		return res.Mentions[i].Change < res.Mentions[j].Change
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

// landedChange is one change whose directory a commit removed from the tip,
// with that commit's time and the tree-ish path its narrative.md is read at.
type landedChange struct {
	name string
	when time.Time
	rev  string
}

// landedChanges lists root's landed changes from one history walk, newest
// commit per name, a drop subject over a legacy archive one as landed.File
// prefers it. A root that is not a git repository has none; any other git
// failure is an error.
func landedChanges(root string) ([]landedChange, error) {
	out, err := exec.Command("git", "-C", root, "log", "--all", "--format=%H %ct %s",
		"--fixed-strings", "--grep", "chore(spectre): ").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && bytes.Contains(ee.Stderr, []byte("not a git repository")) {
			return nil, nil
		}
		return nil, fmt.Errorf("git log in %s: %w", root, err)
	}
	byName := map[string]landedChange{}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.SplitN(line, " ", 3)
		if len(f) < 3 {
			continue
		}
		m := landedSubject.FindStringSubmatch(f[2])
		if m == nil {
			continue
		}
		prev, seen := byName[m[1]]
		drop := strings.HasPrefix(f[2], "chore(spectre): drop ")
		if seen && (!drop || !strings.Contains(prev.rev, ":spectre/changes/archive/")) {
			continue
		}
		sec, _ := strconv.ParseInt(f[1], 10, 64)
		c := landedChange{name: m[1], when: time.Unix(sec, 0)}
		if drop {
			c.rev = f[0] + "^:spectre/changes/" + m[1] + "/narrative.md"
		} else {
			c.rev = f[0] + ":spectre/changes/archive/" + m[1] + "/narrative.md"
		}
		if !seen {
			names = append(names, m[1])
		}
		byName[m[1]] = c
	}
	cs := make([]landedChange, 0, len(names))
	for _, n := range names {
		cs = append(cs, byName[n])
	}
	return cs, nil
}

// readNarrative reads one landed change's narrative.md from history and
// reports whether any of its lines matches the needle. A git failure is an
// error; no narrative or no match is a nil error with matched false.
func readNarrative(root Root, c landedChange, needle string) (Mention, bool, error) {
	content, err := exec.Command("git", "-C", root.Path, "show", c.rev).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && (bytes.Contains(ee.Stderr, []byte("does not exist")) || bytes.Contains(ee.Stderr, []byte("exists on disk, but not in"))) {
			return Mention{}, false, nil
		}
		return Mention{}, false, fmt.Errorf("git show %s: %w", c.rev, err)
	}
	m := Mention{
		Repo:    root.Path,
		Change:  c.name,
		Path:    "spectre/changes/" + c.name + "/narrative.md",
		ModTime: c.when,
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(Normalize(line), needle) {
			if len(m.Lines) < matchLineCap {
				m.Lines = append(m.Lines, strings.TrimSpace(line))
			}
		}
	}
	return m, len(m.Lines) > 0, nil
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
