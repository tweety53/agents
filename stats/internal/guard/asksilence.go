package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// checkAskSilence is scripts/check-ask-silence.sh's guard, and this comment
// is the contract's one home — the shim header defers here. It fails when a
// run-loaded ask site states none of its own silence outcome, a delegation
// of the ask, or a citation of the pipeline contract's
// **Unanswered mid-run asks** section (skills/flow-contracts/pipeline.md).
// KAN-880, deferred from KAN-772's self-review: the corpus-wide survey of
// AskUserQuestion sites was hand-run once, and nothing re-ran it, so a
// future ask site added without its own silence outcome drifted silently.
//
// THE UNIT is the section: from a Markdown heading line to the next heading
// of any level, plus the preamble before a file's first heading. A section
// that mentions AskUserQuestion outside a code fence is an ask site, and its
// whole section must pass.
//
// A section passes when its text outside code fences, case-insensitively,
// carries any of:
//
//   - the citation: "unanswered mid-run asks";
//   - the contract's own silence vocabulary — "silence", "silent",
//     "unanswered", "default", "recommended", "explicit", "without asking"
//     (substring match, so "explicitly" and "defaults" count);
//   - a delegation preposition before a bolded cross-reference — "per **",
//     "applies **", "under **", "exactly as **" — handing the ask's
//     governance to a section this guard cannot follow.
//
// LIMITS, deliberately. The detection is lexical, never semantic: a site
// whose outcome is stated in words outside the vocabulary fails, and the
// remedy is to state the outcome, delegate the ask, or cite the section —
// never to widen the
// vocabulary to hide one. A delegation's target is not checked, so a section
// can pass by pointing at prose that itself states nothing; that is the
// preposition list's price, paid for not parsing prose. Text inside a fenced
// code block is neither a heading, a mention nor an outcome — a worked
// example in a fence is not a live ask.
//
// THE CORPUS is ownedCorpus (ownedcorpus.go), shared with
// check-normative-inventory.
//
// Exit 0 when every ask section passes (and when the corpus holds none), 1
// on violation(s), 2 when the corpus cannot be answered at all —
// FLOW_GUARD_REPO_ROOT unset, a missing or unreadable scope root, a symlink
// hiding Markdown or one find -L cannot look through, or a corpus file that
// exists but cannot be read.
func init() {
	Registry["check-ask-silence"] = checkAskSilence
}

const askSilenceName = "check-ask-silence"

// asVocabulary is the contract's own words for a stated silence outcome; asDelegation
// is the prepositions a section delegates its ask's governance with.
var (
	asVocabulary  = []string{"silence", "silent", "unanswered", "default", "recommended", "explicit", "without asking"}
	asDelegation  = []string{"per **", "applies **", "under **", "exactly as **"}
	asCitation    = "unanswered mid-run asks"
	asTool        = "AskUserQuestion"
	asHeadingMark = "#"
)

type asSection struct {
	heading string // the heading line, trimmed; "(preamble)" before the first heading
	start   int    // 1-based line the section starts at, its heading included
	body    string // the section's text outside code fences, its heading line included
}

func checkAskSilence(_ []string, env Env, stdout, stderr io.Writer) int {
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		fmt.Fprintf(stderr, askSilenceName+": FLOW_GUARD_REPO_ROOT is unset — run scripts/check-ask-silence.sh, which sets it\n")
		return 2
	}
	files, code := ownedCorpus(root, askSilenceName, stderr)
	if files == nil {
		return code
	}

	violations := 0
	sections := 0
	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			// An unreadable file is the corpus not answerable in full —
			// the same refusal owned_corpus_files makes for a missing
			// scope root, never a partial answer counted as violations.
			fmt.Fprintf(stderr, askSilenceName+": %s exists but cannot be read — %v\n", path, err)
			return 2
		}
		for _, sec := range asSections(string(b)) {
			if !strings.Contains(sec.body, asTool) {
				continue
			}
			sections++
			if asPasses(sec.body) {
				continue
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			fmt.Fprintf(stderr, "%s:%d: ASK-SILENCE: section %q mentions %s and states no silence outcome, no delegation of the ask, and no citation of Unanswered mid-run asks (skills/flow-contracts/pipeline.md)\n",
				rel, sec.start, sec.heading, asTool)
			violations++
		}
	}

	if violations != 0 {
		fmt.Fprintf(stdout, "ASK-SILENCE-FAIL: %s — %d ask section(s) of %d state no silence outcome; state the outcome, delegate the ask, or cite Unanswered mid-run asks (skills/flow-contracts/pipeline.md)\n",
			root, violations, sections)
		return 1
	}
	if sections == 0 {
		fmt.Fprintf(stdout, "ASK-SILENCE-OK: %s — no AskUserQuestion site in the corpus\n", root)
		return 0
	}
	fmt.Fprintf(stdout, "ASK-SILENCE-OK: %s — %d ask section(s), each citing, delegating, or stating its silence outcome\n", root, sections)
	return 0
}

// asPasses reports whether one section's fence-free body states a silence
// outcome, cites the contract, or delegates the ask — the header's three
// passes, in that order.
func asPasses(body string) bool {
	low := strings.ToLower(body)
	if strings.Contains(low, asCitation) {
		return true
	}
	for _, w := range asVocabulary {
		if strings.Contains(low, w) {
			return true
		}
	}
	for _, p := range asDelegation {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}

// asSections splits one file into sections at Markdown headings, outside
// code fences. The heading line opens the next section and is part of the
// body it opens — a section naming the tool only in its heading is still an
// ask site; the preamble before a file's first heading is its own section.
func asSections(text string) []asSection {
	var secs []asSection
	cur := asSection{heading: "(preamble)", start: 1}
	flush := func() {
		if cur.body != "" || cur.heading != "(preamble)" {
			secs = append(secs, cur)
		}
	}
	inFence := false
	line := 1
	for _, raw := range strings.SplitAfter(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~"):
			inFence = !inFence
		case !inFence && strings.HasPrefix(trimmed, asHeadingMark) && isHeadingLine(trimmed):
			flush()
			// The heading opens the section and is part of its body: a
			// section naming the tool only in its heading is still an ask
			// site, and the violation line anchors at the heading itself.
			cur = asSection{heading: trimmed, start: line, body: raw}
		default:
			if !inFence {
				cur.body += raw
			}
		}
		line++
	}
	flush()
	return secs
}

// isHeadingLine distinguishes a `# …` heading from a non-heading line that
// merely starts with the mark: one to six marks, then a space or the end of
// the line.
func isHeadingLine(trimmed string) bool {
	n := 0
	for n < len(trimmed) && trimmed[n] == '#' {
		n++
	}
	return n >= 1 && n <= 6 && (n == len(trimmed) || trimmed[n] == ' ')
}
