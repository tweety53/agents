package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// checkFrameList is scripts/check-frame-list.sh: that script's header comment
// is the contract. It derives the frames design.md declares from the mockups
// directory and reconciles them against the capture specs' sidecars and the
// verifier report's per-frame and `- frames:` lines.
func init() { Registry["check-frame-list"] = checkFrameList }

var (
	flSpecLine   = regexp.MustCompile(`^\s*- spec:\s*(\S+)\s*$`)
	flFramesLine = regexp.MustCompile(`^\s*- frames:\s*(\d+)/(\d+)(.*)$`)
	// flConsidered is a decision's rejected-alternatives line: its frames are not declared.
	flConsidered = regexp.MustCompile(`(?m)^\s*\*\*Considered:\*\*.*$`)
	// flRange is `<prefix><a> … <prefix><b>`, each end optionally a path ending in `/`, `.png` and backticks.
	flRange = regexp.MustCompile("`?(?:[^\\s`]*/)?([A-Za-z0-9_-]*?)(\\d+)(?:\\.png)?`?\\s*(?:…|–|\\.\\.\\.)\\s*`?(?:[^\\s`]*/)?([A-Za-z0-9_-]*?)(\\d+)(?:\\.png)?`?")
)

// flToken matches id as a whole token: not preceded or followed by a letter,
// digit, `_` or `-`.
func flToken(id string) *regexp.Regexp {
	return regexp.MustCompile(`(^|[^A-Za-z0-9_-])` + regexp.QuoteMeta(id) + `($|[^A-Za-z0-9_-])`)
}

func checkFrameList(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-frame-list"
	if len(args) != 3 {
		fmt.Fprintf(stderr, "%s: usage: check-frame-list.sh <design.md> <mockups dir> <report file>\n", prog)
		return 2
	}
	design, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
		return 2
	}
	entries, err := os.ReadDir(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read mockups dir %s: %v\n", prog, args[1], err)
		return 2
	}
	report, err := os.ReadFile(args[2])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[2], err)
		return 2
	}

	design = flConsidered.ReplaceAll(design, nil)
	frames := map[string]bool{}
	for _, e := range entries {
		if id, ok := strings.CutSuffix(e.Name(), ".png"); ok && !e.IsDir() && id != "" {
			frames[id] = true
		}
	}
	ranges := flRange.FindAllStringSubmatch(string(design), -1)
	// A frame is in a range when it carries the range's prefix and a number between its ends:
	// one check per mockup and range, never one per integer the span covers.
	inRange := func(id string) bool {
		for _, m := range ranges {
			if m[1] != m[3] {
				continue
			}
			n, err := strconv.Atoi(strings.TrimPrefix(id, m[1]))
			if !strings.HasPrefix(id, m[1]) || err != nil || strconv.Itoa(n) != id[len(m[1]):] {
				continue
			}
			a, _ := strconv.Atoi(m[2])
			b, _ := strconv.Atoi(m[4])
			if a <= n && n <= b {
				return true
			}
		}
		return false
	}
	var declared []string
	for id := range frames {
		if inRange(id) || flToken(id).Match(design) {
			declared = append(declared, id)
		}
	}
	sort.Strings(declared)

	lines := strings.Split(string(report), "\n")
	sidecar := map[string]bool{}
	framesLine := []string(nil)
	for _, line := range lines {
		if m := flSpecLine.FindStringSubmatch(line); m != nil {
			data, err := os.ReadFile(m[1] + ".mockups")
			if err != nil {
				continue
			}
			for _, l := range strings.Split(string(data), "\n") {
				f := strings.Fields(l)
				if len(f) >= 2 && !strings.HasPrefix(f[0], "#") {
					sidecar[f[1]] = true
				}
			}
		}
		if m := flFramesLine.FindStringSubmatch(line); m != nil {
			framesLine = m
		}
	}

	var gaps []string
	if len(declared) > 0 {
		if framesLine == nil {
			gaps = append(gaps, "no frames line")
		} else if m, _ := strconv.Atoi(framesLine[2]); m < len(declared) {
			gaps = append(gaps, fmt.Sprintf("frames %s/%s below %d declared", framesLine[1], framesLine[2], len(declared)))
		}
	}
	for _, id := range declared {
		// A frame the frames line names absent with its reason had no capture, so no sidecar line either.
		// Only an entry's leading id counts — `<id>: <reason>` after the `—` or a `;`/`,` separator —
		// never an id the reason text cites.
		absent := framesLine != nil && regexp.MustCompile(`(^|[—;,])\s*`+regexp.QuoteMeta(id)+`\s*:`).MatchString(framesLine[3])
		if !sidecar[id] && !absent {
			gaps = append(gaps, "no sidecar line: "+id)
		}
		// A per-frame line may name several frames joined by ` / ` — `- LM-L9 / LM-L11 …:` names both
		// (kan-1008's self-review: LM-L11 read as unreported and blocked the dispatch).
		perFrame := regexp.MustCompile(`^\s*- ([^\s:]+ / )*` + regexp.QuoteMeta(id) + `( [^:]*)?:`)
		named := absent
		for _, line := range lines {
			if perFrame.MatchString(line) {
				named = true
				break
			}
		}
		if !named {
			gaps = append(gaps, "no report line: "+id)
		}
	}
	if len(gaps) > 0 {
		fmt.Fprintf(stdout, "FRAME-LIST-GAP: %s — %s\n", args[2], strings.Join(gaps, "; "))
		return 1
	}
	fmt.Fprintf(stdout, "FRAME-LIST-COMPLETE: %s — %d declared\n", args[2], len(declared))
	return 0
}
