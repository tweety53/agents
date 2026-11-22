package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// flow-guard check-frame-constants — two frames of one change must not draw
// one shared constant two ways (kan-1155: the frames contradicted each other
// on bar heights and greys, and every verify round re-found it).
//
// Usage: flow-guard check-frame-constants <design.md> <frames>...
//
// The constants are the names design.md lists under `## Shared constants`,
// one per list item: "- `<name>`: <value>" (skills/flow/brainstorm-planner.md).
// Each <frames> is a directory, every frame under it judged, or one frame
// file. A frame is a `.png` or `.html` file, named by its path below the
// directory less the extension, and it states a constant's value in either of
// two places:
//   - a sidecar beside it, `<frame>.constants`, one `<name>: <value>` per
//     line, `#` opening a comment line;
//   - in an HTML frame, an attribute `data-const-<name>="<value>"` on any
//     element.
//
// Every frame states every listed constant: the value it draws, or `n/a` when
// it does not draw the constant at all. A frame that states nothing is not a
// frame that agrees — a rendered PNG carries no readable constant, so a set
// of them passed unread until each had its sidecar.
//
// Prints on stderr one line per frame that leaves a listed constant unstated
// and one per conflicting constant:
//
//	FRAME-CONSTANT-UNSTATED: <frame> — <name>, …
//	FRAME-CONSTANT-CONFLICT: <name> — <value> (<frame>, …) vs <value> (<frame>, …)
//
// Exit 0 every frame states every listed constant and none is drawn two ways
// (FRAME-CONSTANTS-OK on stdout, with how many frames and frame values it
// compared); 1 a conflict or an unstated constant; 2 a usage error, an
// unreadable input, a design.md with no `## Shared constants` list, or no
// frame under the given paths.
func init() { Registry["check-frame-constants"] = checkFrameConstants }

var (
	fcListItem = regexp.MustCompile("^[-*] `([^`]+)`")
	fcDataAttr = regexp.MustCompile(`data-const-([A-Za-z0-9_.-]+)="([^"]*)"`)
)

func checkFrameConstants(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard check-frame-constants <design.md> <frames dir or frame file>...")
		return 2
	}
	design, err := os.ReadFile(pcAbs(env, args[0]))
	if err != nil {
		Reason(env, "design-unreadable")
		fmt.Fprintf(stderr, "check-frame-constants: cannot read %s: %v\n", args[0], err)
		return 2
	}
	var listed []string
	in := false
	for _, l := range strings.Split(string(design), "\n") {
		if strings.HasPrefix(l, "## ") {
			in = strings.EqualFold(strings.TrimSpace(l[3:]), "Shared constants")
		} else if m := fcListItem.FindStringSubmatch(strings.TrimSpace(l)); in && m != nil {
			listed = append(listed, m[1])
		}
	}
	if len(listed) == 0 {
		Reason(env, "no-shared-constants")
		fmt.Fprintf(stderr, "check-frame-constants: %s lists no constant under `## Shared constants` (\"- `<name>`: <value>\")\n", args[0])
		return 2
	}

	stated := map[string]map[string]string{} // frame → constant → value
	note := func(frame, name, value string) {
		name, value = strings.Trim(strings.TrimSpace(name), "`"), strings.Trim(strings.TrimSpace(value), "`")
		if value != "" {
			stated[frame][name] = value
		}
	}
	read := func(p, frame string) error {
		ext := filepath.Ext(p)
		if ext != ".png" && ext != ".html" {
			return nil
		}
		if stated[frame] == nil {
			stated[frame] = map[string]string{}
		}
		if ext == ".html" {
			body, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for _, m := range fcDataAttr.FindAllStringSubmatch(string(body), -1) {
				note(frame, m[1], m[2])
			}
		}
		body, err := os.ReadFile(strings.TrimSuffix(p, ext) + ".constants")
		if os.IsNotExist(err) {
			return nil
		}
		for _, l := range strings.Split(string(body), "\n") {
			if name, value, ok := strings.Cut(l, ":"); ok && !strings.HasPrefix(strings.TrimSpace(l), "#") {
				note(frame, name, value)
			}
		}
		return err
	}
	for _, arg := range args[1:] {
		root := pcAbs(env, arg)
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			frame := filepath.Base(p)
			if rel, _ := filepath.Rel(root, p); rel != "." {
				frame = filepath.ToSlash(rel)
			}
			return read(p, strings.TrimSuffix(frame, filepath.Ext(frame)))
		})
		if err != nil {
			Reason(env, "frames-unreadable")
			fmt.Fprintf(stderr, "check-frame-constants: cannot read %s: %v\n", arg, err)
			return 2
		}
	}
	if len(stated) == 0 {
		Reason(env, "no-frames")
		fmt.Fprintf(stderr, "check-frame-constants: no `.png` or `.html` frame under %s\n", strings.Join(args[1:], ", "))
		return 2
	}

	frames := make([]string, 0, len(stated))
	for f := range stated {
		frames = append(frames, f)
	}
	sort.Strings(frames)
	failed, compared := "", 0
	for _, f := range frames {
		var unstated []string
		for _, name := range listed {
			if stated[f][name] == "" {
				unstated = append(unstated, name)
			}
		}
		if len(unstated) > 0 {
			failed = "frame-constants-unstated"
			fmt.Fprintf(stderr, "FRAME-CONSTANT-UNSTATED: %s — %s\n", f, strings.Join(unstated, ", "))
		}
	}
	for _, name := range listed {
		byValue := map[string][]string{}
		for _, f := range frames {
			if v := stated[f][name]; v != "" && !strings.EqualFold(v, "n/a") {
				byValue[v] = append(byValue[v], f)
				compared++
			}
		}
		if len(byValue) < 2 {
			continue
		}
		var sides []string
		for value, fs := range byValue {
			sides = append(sides, fmt.Sprintf("%s (%s)", value, strings.Join(fs, ", ")))
		}
		sort.Strings(sides)
		failed = "frame-constants-disagree"
		fmt.Fprintf(stderr, "FRAME-CONSTANT-CONFLICT: %s — %s\n", name, strings.Join(sides, " vs "))
	}
	if failed != "" {
		Reason(env, failed)
		return 1
	}
	fmt.Fprintf(stdout, "FRAME-CONSTANTS-OK: %d listed constants, %d frames, %d frame values compared\n", len(listed), len(frames), compared)
	return 0
}
