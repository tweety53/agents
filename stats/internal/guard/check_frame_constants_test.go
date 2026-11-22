package guard

import (
	"strings"
	"testing"
)

// Two frames drawing one listed constant two ways is a conflict, whether a
// frame states it in a sidecar or a data attribute; an unlisted name is not
// judged, and `n/a` states a constant the frame does not draw.
func TestCheckFrameConstants(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root+"/design.md", "# demo\n\n## Shared constants\n\n- `bar-height`: 56px\n- `rail-grey`: #8a8a8a\n\n## Decisions\n\n- `not-a-constant`: x\n")
	writeFile(t, root+"/frames/WU1.png", "png")
	writeFile(t, root+"/frames/WU1.constants", "# measured on the frame\nbar-height: 56px\nrail-grey: #8a8a8a\nnot-a-constant: 1\n")
	writeFile(t, root+"/frames/tag/WU2.html", `<header data-const-bar-height="56px" data-const-rail-grey="n/a" data-const-not-a-constant="2"></header>`)
	env := Env{Dir: root}
	if r := runGuard("check-frame-constants", []string{"design.md", "frames"}, env); r.rc != 0 || !strings.Contains(r.out, "FRAME-CONSTANTS-OK: 2 listed constants, 2 frames, 3 frame values compared") {
		t.Fatalf("agreeing frames: exit %d\n%s", r.rc, r.out+r.err)
	}
	writeFile(t, root+"/more/WU4.png", "png")
	writeFile(t, root+"/more/WU4.constants", "bar-height: 64px\nrail-grey: n/a\n")
	r := runGuard("check-frame-constants", []string{"design.md", "frames", "more/WU4.png"}, env)
	if r.rc != 1 || !strings.Contains(r.err, "FRAME-CONSTANT-CONFLICT: bar-height — 56px (WU1, tag/WU2) vs 64px (WU4)") || strings.Contains(r.err, "not-a-constant") {
		t.Fatalf("conflicting frames: exit %d, want 1 naming bar-height alone\n%s", r.rc, r.out+r.err)
	}
	writeFile(t, root+"/bare.md", "# demo\n")
	if r := runGuard("check-frame-constants", []string{"bare.md", "frames"}, env); r.rc != 2 {
		t.Fatalf("no Shared constants list: exit %d, want 2\n%s", r.rc, r.out+r.err)
	}
}

// A frame that states nothing does not agree by silence (gymie's frames are
// rendered PNGs, which carry no readable constant): each frame leaving a
// listed constant unstated is named, and a path holding no frame is refused.
func TestCheckFrameConstantsRefusesAnUnstatedFrame(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root+"/design.md", "## Shared constants\n\n- `bar-height`: 56px\n- `rail-grey`: #8a8a8a\n")
	writeFile(t, root+"/frames/WU1.png", "png")
	writeFile(t, root+"/frames/WU2.png", "png")
	writeFile(t, root+"/frames/WU2.constants", "bar-height: 56px\n")
	env := Env{Dir: root}
	r := runGuard("check-frame-constants", []string{"design.md", "frames"}, env)
	if r.rc != 1 || !strings.Contains(r.err, "FRAME-CONSTANT-UNSTATED: WU1 — bar-height, rail-grey\n") || !strings.Contains(r.err, "FRAME-CONSTANT-UNSTATED: WU2 — rail-grey\n") {
		t.Fatalf("unstated frames: exit %d, want 1 naming WU1 and WU2\n%s", r.rc, r.out+r.err)
	}
	writeFile(t, root+"/empty/notes.txt", "x")
	if r := runGuard("check-frame-constants", []string{"design.md", "empty"}, env); r.rc != 2 || !strings.Contains(r.err, "no `.png` or `.html` frame") {
		t.Fatalf("no frames: exit %d, want 2\n%s", r.rc, r.out+r.err)
	}
}
