package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-frame-list.sh shim (KAN-1058).
//
// The folded check-frame-list guard (no longer a flow-guard subcommand; `close` runs it
// in-process) — reconcile the mockup frames design.md declares against
// the capture specs' `.mockups` sidecars and a flow.visual-verify verifier's
// report, so a declared frame no capture reached is named rather than lost
// (KAN-549: design.md declared 22 frames, two rounds reported 15/17 composed,
// and nothing compared the two).
//
// Arguments: <design.md> <mockups dir> <report file> [--frames <frame>[,<frame>…]]
//
// <design.md> is the change's <changeRoot>/design.md; <mockups dir> the
// resolved `mockups` root (skills/flow-contracts/project-configuration-visual.md);
// <report file> the verifier's
// <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
// (skills/flow/visual-verify.md).
//
// THE DECLARED LIST is derived, never written down twice: every
// `<frame id>.png` directly inside <mockups dir> whose <frame id> design.md
// names as a whole token — bounded on each side by the start or end of a line
// or by a character that is not a letter, digit, `_` or `-`. A range
// `<prefix><a> … <prefix><b>` (joined by `…`, `–` or `...`, each end optionally
// a path ending in `/`, a `.png` suffix and backticks) names every
// `<prefix><n>` from <a> to <b>; ends with different prefixes name only
// themselves. A design line opening with `**Considered:**` names no frame:
// its alternatives were rejected.
//
// A sub-directory `<sub>` of <mockups dir> that design.md names as `<sub>/`
// (left-bounded like a token) is cited, and its `*.png` frames join the
// candidates as `<sub>/<id>`. Such a frame is declared when design.md names
// `<sub>/<id>` as a token, or when exactly one cited sub-directory holds
// `<id>` and design.md names bare `<id>` (a token not preceded by `/`) or
// covers it with a range; an `<id>` two or more cited sub-directories hold is
// declared only by its explicit `<sub>/<id>`. A root frame is never declared
// when a cited sub-directory holds the same id — the sub-directory shadows it
// — and a root id preceded by `/` matches only when the path segment before it
// is <mockups dir>'s own base name. While design.md cites a sub-directory, a
// root frame is declared only by `<id>.png` or a path ending in
// `<mockups base>/<id>`, never by a bare token or a range (kan-1087
// self-review, F5: "rules R1–R49" declared another change's root `R1.png`). No `dk-<id>` twin is inferred: it is
// declared only when named itself (kan-1084 self-review, F12: a design citing
// `kan-1084/` and naming P3 bare declared the root P3.png of another change,
// while its own eleven frames went undeclared).
//
// An id design.md writes in answer position — directly followed by `=` and a
// letter or digit, `Q1=A` — is a question label, and the design declares no
// frame by that id anywhere. So is an id in citation position — letters then
// digits written directly after `round <n>, `, `answer-<n> ` or the word `question `, any case
// (`round 1, Q1`, `answer-2 Q1`, `question Q4`) — kan-1102 self-review, F1.
//
// With --frames, a shard's slice (skills/flow/visual-verify.md), the declared
// list is the declared frames in the slice, and the frames line's <m> is
// compared with that count (decision manifest-per-shard).
//
// Every report line below is read from the report's `## Report` heading to
// the next `## ` heading alone: a `## Verdict` departure line never stands in
// for a per-frame line (decision verdict-only-handback).
//
// THE SIDECARS are `<spec>.mockups` beside every spec the report's
// `- spec: <path>` lines name; a missing sidecar contributes no frame. A
// sidecar line is `<screenshot name> <frame id>`, `#` comments and blank lines
// ignored.
//
// A GAP is, for a declared frame:
//   no sidecar line: <frame id>   no read sidecar maps a screenshot to it
//   no report line: <frame id>    no `- <frame id>:` or `- <frame id> <…>:`
//                                 line — the id may follow other ids joined
//                                 by ` / `, `- <id> / <frame id> <…>:` — and
//                                 the `- frames:` line does not name it
// and, for the report:
//   no frames line                the report carries no `- frames: <n>/<m>`
//                                 line while design.md declares a frame
//   frames <n>/<m> below <k> declared   <m> is smaller than the declared count
//
// Prints ONE verdict line to stdout:
//   FRAME-LIST-COMPLETE: <report file> — <k> declared
//   FRAME-LIST-GAP: <report file> — <each gap, `; `-separated>
//
// Exit 0 no gap, 1 at least one gap, 2 cannot answer (a usage error, an
// unreadable design.md or report, a mockups dir that is not a directory).
//
// Go, in flow-guard (stats/internal/guard/framelist.go).

const frameListUsage = "%s: usage: frame-list check, run in-process by visual-verify-pass close: <design.md> <mockups dir> <report file> [--frames <frame>[,<frame>…]]\n"

// checkFrameList is the frame-list check `visual-verify-pass close` runs in-process: the comment above is the contract. It derives the frames design.md declares from the mockups
// directory and reconciles them against the capture specs' sidecars and the
// verifier report's per-frame and `- frames:` lines.

var (
	flSpecLine   = regexp.MustCompile(`^\s*- spec:\s*(\S+)\s*$`)
	flFramesLine = regexp.MustCompile(`^\s*- frames:\s*(\d+)/(\d+)(.*)$`)
	// flConsidered is a decision's rejected-alternatives line: its frames are not declared.
	flConsidered = regexp.MustCompile(`(?m)^\s*\*\*Considered:\*\*.*$`)
	// flAnswered is an id in answer position, `Q1=A`: a question label, never a frame.
	flAnswered = regexp.MustCompile(`(?m)(^|[^A-Za-z0-9_-])([A-Za-z0-9_-]+)=[A-Za-z0-9]`)
	// flCited is a question label in citation position, `round 1, Q1` or `question Q4`: never a frame.
	flCited = regexp.MustCompile(`(?im)(?:\bround \d+, |\banswer-\d+ |\bquestion )([A-Za-z]+\d+)(?:[^A-Za-z0-9_-]|$)`)
	// flRange is `<prefix><a> … <prefix><b>`, each end optionally a path ending in `/`, `.png` and backticks.
	flRange = regexp.MustCompile("`?(?:[^\\s`]*/)?([A-Za-z0-9_-]*?)(\\d+)(?:\\.png)?`?\\s*(?:…|–|\\.\\.\\.)\\s*`?(?:[^\\s`]*/)?([A-Za-z0-9_-]*?)(\\d+)(?:\\.png)?`?")
)

// flBare matches id as a token never preceded by `/`, save after prefix
// (the mockups root's own `<base>/`) when prefix is not empty.
func flBare(id, prefix string) *regexp.Regexp {
	left := `(^|[^A-Za-z0-9_/-]`
	if prefix != "" {
		left += `|` + regexp.QuoteMeta(prefix)
	}
	return regexp.MustCompile(left + `)` + regexp.QuoteMeta(id) + `($|[^A-Za-z0-9_-])`)
}

// flToken matches id as a whole token: not preceded or followed by a letter,
// digit, `_` or `-`.
func flToken(id string) *regexp.Regexp {
	return regexp.MustCompile(`(^|[^A-Za-z0-9_-])` + regexp.QuoteMeta(id) + `($|[^A-Za-z0-9_-])`)
}

func checkFrameList(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-frame-list"
	var slice []string // --frames: the shard's slice, nil for every frame
	if i := slices.Index(args, "--frames"); i >= 0 {
		if i+1 == len(args) || args[i+1] == "" {
			fmt.Fprintf(stderr, frameListUsage, prog)
			Reason(env, "usage")
			return 2
		}
		slice = strings.Split(args[i+1], ",")
		args = slices.Delete(slices.Clone(args), i, i+2)
	}
	if len(args) != 3 {
		fmt.Fprintf(stderr, frameListUsage, prog)
		Reason(env, "usage")
		return 2
	}
	design, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
		Reason(env, "cannot-read")
		return 2
	}
	entries, err := os.ReadDir(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read mockups dir %s: %v\n", prog, args[1], err)
		Reason(env, "cannot-read-mockups-dir")
		return 2
	}
	report, err := os.ReadFile(args[2])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[2], err)
		Reason(env, "cannot-read")
		return 2
	}

	declared := flDeclared(design, args[1], entries)
	if slice != nil {
		declared = slices.DeleteFunc(declared, func(id string) bool { return !slices.Contains(slice, id) })
	}

	lines := flReportLines(report)
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
				if len(f) >= 2 && !strings.HasPrefix(f[0], "#") && f[0] != "frame" {
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
		if !absent && !flPerFrame(id, lines) {
			gaps = append(gaps, "no report line: "+id)
		}
	}
	if len(gaps) > 0 {
		fmt.Fprintf(stdout, "FRAME-LIST-GAP: %s — %s\n", args[2], strings.Join(gaps, "; "))
		Reason(env, "frame-list-gap")
		return 1
	}
	fmt.Fprintf(stdout, "FRAME-LIST-COMPLETE: %s — %d declared\n", args[2], len(declared))
	return 0
}

// flDeclared is the frames design declares among the mockups in dir (its
// entries): the answered and cited sets, the `**Considered:**` strip, the
// ranges, the sub-directory holders and the declared list, sorted.
func flDeclared(design []byte, dir string, entries []os.DirEntry) []string {
	answered := map[string]bool{}
	for _, m := range flAnswered.FindAllSubmatch(design, -1) {
		answered[string(m[2])] = true
	}
	for _, m := range flCited.FindAllSubmatch(design, -1) {
		answered[string(m[1])] = true
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
	// holders maps a frame id to the cited sub-directories holding it.
	holders := map[string][]string{}
	subCited := false
	for _, e := range entries {
		sub := e.Name()
		if !e.IsDir() || !regexp.MustCompile(`(^|[^A-Za-z0-9_-])`+regexp.QuoteMeta(sub+"/")).Match(design) {
			continue
		}
		subCited = true
		pngs, _ := filepath.Glob(filepath.Join(dir, sub, "*.png"))
		for _, p := range pngs {
			if id := strings.TrimSuffix(filepath.Base(p), ".png"); id != "" {
				holders[id] = append(holders[id], sub)
			}
		}
	}
	var declared []string
	for id, subs := range holders {
		for _, sub := range subs {
			full := sub + "/" + id
			if !answered[full] && (flToken(full).Match(design) ||
				len(subs) == 1 && !answered[id] && (inRange(id) || flBare(id, "").Match(design))) {
				declared = append(declared, full)
			}
		}
	}
	base := filepath.Base(dir) + "/"
	for id := range frames {
		if len(holders[id]) > 0 || answered[id] {
			continue
		}
		if subCited && (flBare(id+".png", base).Match(design) || flToken(base+id).Match(design)) ||
			!subCited && (inRange(id) || flBare(id, base).Match(design)) {
			declared = append(declared, id)
		}
	}
	sort.Strings(declared)
	return declared
}

// flReportLines is report's lines from `## Report` to the next `## ` heading.
func flReportLines(report []byte) []string {
	// Every line is read from `## Report` to the next `## ` heading alone: a
	// `## Verdict` departure line never stands in for a per-frame line.
	var lines []string
	inReport := false
	for _, line := range strings.Split(string(report), "\n") {
		if strings.HasPrefix(line, "## ") {
			inReport = strings.TrimSpace(line) == "## Report"
		}
		if inReport {
			lines = append(lines, line)
		}
	}
	return lines
}

// flPerFrame reports whether one of lines is id's own per-frame line. A
// per-frame line may name several frames joined by ` / ` — `- LM-L9 / LM-L11 …:`
// names both (kan-1008's self-review: LM-L11 read as unreported and blocked
// the dispatch).
func flPerFrame(id string, lines []string) bool {
	perFrame := regexp.MustCompile(`^\s*- ([^\s:]+ / )*` + regexp.QuoteMeta(id) + `( [^:]*)?:`)
	for _, line := range lines {
		if perFrame.MatchString(line) {
			return true
		}
	}
	return false
}
