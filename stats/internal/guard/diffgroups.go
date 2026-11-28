package guard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// diff-groups <abs-worktree> <diff-file> writes <diff-file minus .diff>.groups.md beside the diff,
// the diff's files grouped mechanically so no reviewer regroups them: one `## <dir>` group per
// directory, a test file joining the directory of the non-test file sharing its stem, docs and
// config files joining the group of their directory (else `## docs and config`). Each file
// is listed with its absolute path (the worktree its `# worktree:` header names, then the
// diff's relative path) and added/removed line counts.
//
// Exit 0 written; 2 usage, an unreadable diff, an unwritable output.
func init() { Registry["diff-groups"] = diffGroups }

type dgFile struct {
	abs, dir, stem string
	add, del       int
	test, doc      bool
}

var (
	dgTestSuffix = regexp.MustCompile(`(_test|\.test|\.spec|Test|Tests)$`)
	dgDocExt     = map[string]bool{".md": true, ".mdc": true, ".yml": true, ".yaml": true, ".json": true, ".toml": true, ".txt": true, ".ini": true, ".conf": true}
)

func diffGroups(args []string, env Env, stdout, stderr io.Writer) int {
	fail := func(id, msg string) int {
		Reason(env, id)
		fmt.Fprintln(stderr, "diff-groups: "+msg)
		return 2
	}
	if len(args) != 2 {
		return fail("usage", "usage: flow-guard diff-groups <abs-worktree> <diff-file>")
	}
	f, err := os.Open(args[1])
	if err != nil {
		return fail("diff-unreadable", err.Error())
	}
	defer f.Close()
	var files []*dgFile
	var cur *dgFile
	wt := ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		l := sc.Text()
		switch {
		case strings.HasPrefix(l, "# worktree: "):
			wt = strings.TrimPrefix(l, "# worktree: ")
			if i := strings.Index(wt, " — "); i >= 0 {
				wt = wt[:i]
			}
			cur = nil
		case strings.HasPrefix(l, "diff --git "):
			i := strings.LastIndex(l, " b/")
			if i < 0 {
				cur = nil
				continue
			}
			rel := l[i+3:]
			base := path.Base(rel)
			ext := path.Ext(base)
			stem := dgTestSuffix.ReplaceAllString(strings.TrimSuffix(base, ext), "")
			cur = &dgFile{abs: filepath.Join(wt, rel), dir: path.Dir(rel), stem: stem,
				test: stem != strings.TrimSuffix(base, ext), doc: dgDocExt[ext]}
			files = append(files, cur)
		case cur != nil && strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++"):
			cur.add++
		case cur != nil && strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---"):
			cur.del++
		}
	}
	if err := sc.Err(); err != nil {
		return fail("diff-unreadable", err.Error())
	}
	stemDir := map[string]string{}
	codeDirs := map[string]bool{}
	for _, x := range files {
		if !x.test && !x.doc {
			codeDirs[x.dir] = true
			if _, ok := stemDir[x.stem]; !ok {
				stemDir[x.stem] = x.dir
			}
		}
	}
	groups := map[string][]*dgFile{}
	for _, x := range files {
		g := x.dir
		switch {
		case x.test:
			if d, ok := stemDir[x.stem]; ok {
				g = d
			}
		case x.doc && !codeDirs[x.dir]:
			g = "docs and config"
		}
		groups[g] = append(groups[g], x)
	}
	names := make([]string, 0, len(groups))
	for g := range groups {
		names = append(names, g)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, g := range names {
		fmt.Fprintf(&b, "## %s\n\n", g)
		sort.Slice(groups[g], func(i, j int) bool { return groups[g][i].abs < groups[g][j].abs })
		for _, x := range groups[g] {
			fmt.Fprintf(&b, "- %s (+%d -%d)\n", x.abs, x.add, x.del)
		}
		b.WriteString("\n")
	}
	dst := bspGroupsPath(args[1])
	if err := os.WriteFile(dst, []byte(b.String()), 0o644); err != nil {
		return fail("cannot-write-groups", err.Error())
	}
	fmt.Fprintln(stdout, dst)
	return 0
}
