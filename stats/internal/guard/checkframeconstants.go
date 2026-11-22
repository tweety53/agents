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
// Usage: flow-guard check-frame-constants <design.md> <frames dir>
//
// The constants are the names design.md lists under `## Shared constants`,
// one per list item: "- `<name>`: <value>" (skills/flow/brainstorm-planner.md).
// A frame states a constant's value in either of two places under
// <frames dir>, searched recursively:
//   - a sidecar `<frame id>.constants`, one `<name>: <value>` per line, `#`
//     opening a comment line;
//   - an HTML frame `<frame id>.html`, as an attribute
//     `data-const-<name>="<value>"` on any element.
//
// A frame that states no value for a constant is not judged on it: a PNG
// frame with no sidecar is silent, never a conflict.
//
// Prints one line per conflicting constant on stderr:
//
//	FRAME-CONSTANT-CONFLICT: <name> — <value> (<frame>, …) vs <value> (<frame>, …)
//
// Exit 0 no listed constant is drawn two ways (FRAME-CONSTANTS-OK on stdout,
// with how many frame values it compared); 1 a conflict; 2 a usage error, an
// unreadable input, or a design.md with no `## Shared constants` list.
func init() { Registry["check-frame-constants"] = checkFrameConstants }

var (
	fcListItem = regexp.MustCompile("^[-*] `([^`]+)`")
	fcDataAttr = regexp.MustCompile(`data-const-([A-Za-z0-9_.-]+)="([^"]*)"`)
)

func checkFrameConstants(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard check-frame-constants <design.md> <frames dir>")
		return 2
	}
	design, err := os.ReadFile(pcAbs(env, args[0]))
	if err != nil {
		Reason(env, "design-unreadable")
		fmt.Fprintf(stderr, "check-frame-constants: cannot read %s: %v\n", args[0], err)
		return 2
	}
	listed := map[string]bool{}
	in := false
	for _, l := range strings.Split(string(design), "\n") {
		if strings.HasPrefix(l, "## ") {
			in = strings.EqualFold(strings.TrimSpace(l[3:]), "Shared constants")
		} else if m := fcListItem.FindStringSubmatch(strings.TrimSpace(l)); in && m != nil {
			listed[m[1]] = true
		}
	}
	if len(listed) == 0 {
		Reason(env, "no-shared-constants")
		fmt.Fprintf(stderr, "check-frame-constants: %s lists no constant under `## Shared constants` (\"- `<name>`: <value>\")\n", args[0])
		return 2
	}

	values := map[string]map[string][]string{} // constant → value → frames
	compared := 0
	note := func(frame, name, value string) {
		if name, value = strings.TrimSpace(name), strings.Trim(strings.TrimSpace(value), "`"); listed[name] && value != "" {
			if values[name] == nil {
				values[name] = map[string][]string{}
			}
			values[name][value] = append(values[name][value], frame)
			compared++
		}
	}
	root := pcAbs(env, args[1])
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(p)
		if ext != ".constants" && ext != ".html" {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		frame := strings.TrimSuffix(filepath.Base(p), ext)
		if ext == ".html" {
			for _, m := range fcDataAttr.FindAllStringSubmatch(string(body), -1) {
				note(frame, m[1], m[2])
			}
			return nil
		}
		for _, l := range strings.Split(string(body), "\n") {
			if name, value, ok := strings.Cut(l, ":"); ok && !strings.HasPrefix(strings.TrimSpace(l), "#") {
				note(frame, strings.Trim(strings.TrimSpace(name), "`"), value)
			}
		}
		return nil
	})
	if err != nil {
		Reason(env, "frames-unreadable")
		fmt.Fprintf(stderr, "check-frame-constants: cannot read %s: %v\n", args[1], err)
		return 2
	}

	var names []string
	for name, byValue := range values {
		if len(byValue) > 1 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var sides []string
		for value, frames := range values[name] {
			sides = append(sides, fmt.Sprintf("%s (%s)", value, strings.Join(frames, ", ")))
		}
		sort.Strings(sides)
		fmt.Fprintf(stderr, "FRAME-CONSTANT-CONFLICT: %s — %s\n", name, strings.Join(sides, " vs "))
	}
	if len(names) > 0 {
		Reason(env, "frame-constants-disagree")
		return 1
	}
	fmt.Fprintf(stdout, "FRAME-CONSTANTS-OK: %d listed constants, %d frame values compared\n", len(listed), compared)
	return 0
}
