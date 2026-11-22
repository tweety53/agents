package guard

import (
	"strings"
	"testing"
)

// Two frames drawing one listed constant two ways is a conflict, whether a
// frame states it in a sidecar or a data attribute; an unlisted name and a
// frame that states nothing are not judged.
func TestCheckFrameConstants(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root+"/design.md", "# demo\n\n## Shared constants\n\n- `bar-height`: 56px\n- `rail-grey`: #8a8a8a\n\n## Decisions\n\n- `not-a-constant`: x\n")
	writeFile(t, root+"/frames/WU1.constants", "# from tokens.css\nbar-height: 56px\nrail-grey: #8a8a8a\nnot-a-constant: 1\n")
	writeFile(t, root+"/frames/WU2.html", `<header data-const-bar-height="56px" data-const-not-a-constant="2"></header>`)
	writeFile(t, root+"/frames/WU3.png", "png")
	env := Env{Dir: root}
	if r := runGuard("check-frame-constants", []string{"design.md", "frames"}, env); r.rc != 0 || !strings.Contains(r.out, "FRAME-CONSTANTS-OK: 2 listed constants, 3 frame values compared") {
		t.Fatalf("agreeing frames: exit %d\n%s", r.rc, r.out+r.err)
	}
	writeFile(t, root+"/frames/WU4.constants", "bar-height: 64px\n")
	r := runGuard("check-frame-constants", []string{"design.md", "frames"}, env)
	if r.rc != 1 || !strings.Contains(r.err, "FRAME-CONSTANT-CONFLICT: bar-height — 56px (WU1, WU2) vs 64px (WU4)") || strings.Contains(r.err, "not-a-constant") {
		t.Fatalf("conflicting frames: exit %d, want 1 naming bar-height alone\n%s", r.rc, r.out+r.err)
	}
	writeFile(t, root+"/bare.md", "# demo\n")
	if r := runGuard("check-frame-constants", []string{"bare.md", "frames"}, env); r.rc != 2 {
		t.Fatalf("no Shared constants list: exit %d, want 2\n%s", r.rc, r.out+r.err)
	}
}
