package repotest

// TestGuardShimsNameTheirSiblings pins that every guard's scripts/<id>.sh shim
// spells each sibling script its Go source execs, so check-guard-symlinks rule
// 2 derives it and setup.sh installs it beside the shim (kan-1048:
// run-capped.sh was never installed, and every suite failed to start;
// kan-1048 self-review, 1-B: the check covered runverifylist.go alone).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	gsSiblingExec = []*regexp.Regexp{
		regexp.MustCompile(`scriptDir, "([^"/]+)\.sh"`),
		regexp.MustCompile(`scriptDir\+"/([^"/]+)\.sh"`),
		regexp.MustCompile(`sibling\("([^"]+)"\)`),
		regexp.MustCompile(`c\.run\("([^"]+)"`),
	}
	gsRegistry = regexp.MustCompile(`Registry\["([^"]+)"\] =`)
	gsSuffix   = regexp.MustCompile(`_[a-z0-9]+\.go$`)
)

func TestGuardShimsNameTheirSiblings(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "stats/internal/guard")
	srcs, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	read := func(p string) string {
		b, err := os.ReadFile(p)
		if err != nil {
			return ""
		}
		return string(b)
	}
	seen := 0
	for _, p := range srcs {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		src := read(p)
		var sibs []string
		for _, re := range gsSiblingExec {
			for _, m := range re.FindAllStringSubmatch(src, -1) {
				sibs = append(sibs, m[1])
			}
		}
		if len(sibs) == 0 {
			continue
		}
		seen += len(sibs)
		name := filepath.Base(p)
		id := gsRegistry.FindStringSubmatch(src)
		if id == nil {
			id = gsRegistry.FindStringSubmatch(read(filepath.Join(dir, gsSuffix.ReplaceAllString(name, ".go"))))
		}
		if id == nil {
			t.Errorf("%s execs siblings %v but maps to no guard id", name, sibs)
			continue
		}
		shim := read(filepath.Join(root, "scripts", id[1]+".sh"))
		if shim == "" {
			t.Errorf("%s execs siblings %v but scripts/%s.sh is missing", name, sibs, id[1])
			continue
		}
		for _, s := range sibs {
			if !strings.Contains(shim, "$SCRIPT_DIR/"+s+".sh") {
				t.Errorf("scripts/%s.sh does not name $SCRIPT_DIR/%s.sh, which %s execs", id[1], s, name)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no sibling exec found in stats/internal/guard")
	}
}
