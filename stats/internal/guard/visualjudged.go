package guard

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// visual-judged.txt (<worktree>/.superpowers/sdd/, beside the manifests) holds
// one `<sha-256 hex> <frame id>` line per frame a closed report judged frame
// by frame: the capture that report judged. close rewrites it from the row's
// manifest pairs (judgedAfter); a scoped round's measure leaves a frame whose
// capture still hashes to its line off its `changed:` line (vvpCarried) —
// skills/flow/visual-verify.md's **A scoped round re-judges only changed
// frames** is canonical for the rule.
const judgedName = "visual-judged.txt"

// readJudged is the judged file of sdd as frame → capture hash; an absent or
// unreadable file is none.
func readJudged(sdd string) map[string]string {
	m := map[string]string{}
	b, _ := os.ReadFile(filepath.Join(sdd, judgedName))
	for _, l := range strings.Split(string(b), "\n") {
		if f := strings.Fields(l); len(f) == 2 {
			m[f[1]] = f[0]
		}
	}
	return m
}

// writeJudged writes m as the judged file of sdd, one line per frame in id order.
func writeJudged(sdd string, m map[string]string) error {
	var lines []string
	for f, h := range m {
		lines = append(lines, h+" "+f)
	}
	sort.Strings(lines)
	body := strings.Join(lines, "\n")
	if body != "" {
		body += "\n"
	}
	return os.WriteFile(filepath.Join(sdd, judgedName), []byte(body), 0o644)
}

// pngHash is path's sha-256 hex, "" when it cannot be read.
func pngHash(path string) string {
	b, err := os.ReadFile(path)
	if path == "" || err != nil {
		return ""
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// vvpCarried is the frames of pairs whose capture hashes to its judged line.
func vvpCarried(judged map[string]string, pairs []vvmPair) []string {
	var out []string
	for _, p := range pairs {
		if h := judged[p.Frame]; h != "" && h == pngHash(p.Capture) && !slices.Contains(out, p.Frame) {
			out = append(out, p.Frame)
		}
	}
	return out
}

var vjDecision = regexp.MustCompile(`decision\s+([^\s;)]+)`)

// judgedAfter is the judged map once a report closed over pairs: a frame the
// report gave its own per-frame line (flPerFrame) and no over-tolerance row
// excused by a decision naming more than one of the pairs' frames gets its
// capture's hash; any other frame keeps its line only while its capture still
// hashes to it, and loses it otherwise.
func judgedAfter(old map[string]string, pairs []vvmPair, report, design []byte) map[string]string {
	lines := flReportLines(report)
	var frames []string
	for _, p := range pairs {
		if !slices.Contains(frames, p.Frame) {
			frames = append(frames, p.Frame)
		}
	}
	out := map[string]string{}
	for f, h := range old {
		out[f] = h
	}
	for _, p := range pairs {
		h := pngHash(p.Capture)
		switch {
		case h != "" && flPerFrame(p.Frame, lines) && !vjCollective(p.Frame, lines, frames, design):
			out[p.Frame] = h
		case h == "" || out[p.Frame] != h:
			delete(out, p.Frame)
		}
	}
	return out
}

// vjCollective reports whether an over-tolerance row of frame is excused by a
// decision whose design.md entry, its `**Considered:**` line aside, names more
// than one of frames.
func vjCollective(frame string, lines, frames []string, design []byte) bool {
	for _, l := range lines {
		m := vrOverLine.FindStringSubmatch(l)
		if m == nil || m[1] != frame {
			continue
		}
		for _, d := range vjDecision.FindAllStringSubmatch(m[2], -1) {
			entry := vjEntry(string(design), d[1])
			n := 0
			for _, f := range frames {
				if flToken(f).MatchString(entry) {
					n++
				}
			}
			if n > 1 {
				return true
			}
		}
	}
	return false
}

// vjEntry is the text of design's decision id, from its `**ID:**` line to the
// next heading or `**ID:**` line, its `**Considered:**` line dropped.
func vjEntry(design, id string) string {
	var b strings.Builder
	in := false
	for _, l := range strings.Split(design, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "**ID:**") {
			in = strings.TrimSpace(strings.TrimPrefix(t, "**ID:**")) == id
			continue
		}
		if strings.HasPrefix(t, "#") {
			in = false
		}
		if in && !flConsidered.MatchString(l) {
			b.WriteString(l + "\n")
		}
	}
	return b.String()
}
