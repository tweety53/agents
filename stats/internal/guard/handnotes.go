package guard

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// checkHandNotesInStep is scripts/check-hand-notes-in-step.sh: that script's
// header comment is the contract — fail when the hand-maintained content of
// the installed harness files (everything outside their `<!-- flow:begin -->
// … <!-- flow:end -->` managed block) has drifted apart. The managed block is
// re-rendered from one source on every install, so it cannot drift; the hand
// appends around it are written by hand twice, once per harness file, and
// nothing else asserts the copies stay identical. KAN-808: an identical
// 8-line pointer was appended to both files by hand, and a one-sided edit of
// such an append would diverge silently forever after.
//
// Exit 0 on the OK, NONE and SINGLE verdicts, 1 on the DRIFT verdict and on
// every refusal: this guard has no exit 2, like check-installed-rules.
func init() {
	Registry["check-hand-notes-in-step"] = checkHandNotesInStep
}

const hnhName = "check-hand-notes-in-step"

func checkHandNotesInStep(_ []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, hnhName+": "+format+"\n", a...)
		return 1
	}
	violations := 0
	violation := func(format string, a ...any) {
		fmt.Fprintf(stderr, hnhName+": "+format+"\n", a...)
		violations++
	}

	// The file set is setup.sh's own `managed_files` declaration, parsed
	// live (kan-585) — the same parser and refusal contract rule 4 of
	// check-installed-rules uses, so a harness file the installer gains
	// moves this guard first too. CHECK_HAND_NOTES_SETUP_SH defaults to
	// this checkout's setup.sh.
	setupSh := env.Getenv("CHECK_HAND_NOTES_SETUP_SH")
	if setupSh == "" {
		repoRoot := env.Getenv("FLOW_GUARD_REPO_ROOT")
		if repoRoot == "" {
			return refuse("FLOW_GUARD_REPO_ROOT is unset and CHECK_HAND_NOTES_SETUP_SH is not set — run scripts/check-hand-notes-in-step.sh, which sets it")
		}
		setupSh = repoRoot + "/setup.sh"
	}
	setup, err := os.ReadFile(setupSh)
	if err != nil {
		return refuse("%s is unreadable — cannot resolve the managed-block targets", setupSh)
	}
	harnessFiles, code := cirManagedFiles(string(setup), setupSh, refuse)
	if harnessFiles == nil {
		return code
	}

	homeDir := env.Getenv("CHECK_HAND_NOTES_HOME")
	if homeDir == "" {
		h, ok := env.LookupEnv("HOME")
		if !ok {
			return refuse("HOME is unset and CHECK_HAND_NOTES_HOME is not set — cannot locate the install")
		}
		homeDir = h
	}

	type hnhFile struct {
		path, hand string
	}
	var existing []hnhFile
	for _, rel := range harnessFiles {
		path := homeDir + "/" + rel
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		hand, terminated := hnhHandContent(string(b))
		if !terminated {
			violation("%s ends inside an unterminated managed block — repair the file", path)
		}
		existing = append(existing, hnhFile{path: path, hand: hand})
	}

	if len(existing) == 0 {
		fmt.Fprintf(stdout, "HAND-NOTES-NONE: %s — no global install found, nothing to check\n", homeDir)
		return 0
	}
	if len(existing) == 1 && violations == 0 {
		fmt.Fprintf(stdout, "HAND-NOTES-SINGLE: %s — one harness file present, no pair to compare\n", homeDir)
		return 0
	}

	// Byte-compare every file's hand content against the first file's,
	// naming the first line where the copies part. The comparison is over
	// the lines the hand content keeps, newline endings included, so a
	// lost final newline is drift like any other one-sided edit. A
	// content position past a file's end — a missing line, or the empty
	// tail piece a final newline leaves — is named "<EOF>", never quoted
	// as an empty line.
	lineAt := func(lines []string, i int) string {
		if i >= len(lines) || (i == len(lines)-1 && lines[i] == "") {
			return "<EOF>"
		}
		return lines[i]
	}
	for _, f := range existing[1:] {
		base := strings.SplitAfter(existing[0].hand, "\n")
		this := strings.SplitAfter(f.hand, "\n")
		for i := 0; i <= max(len(base), len(this)); i++ {
			if baseLine, thisLine := lineAt(base, i), lineAt(this, i); baseLine != thisLine {
				violation("%s differs from %s at hand-section line %d: %s vs %s",
					f.path, existing[0].path, i+1, strconv.Quote(thisLine), strconv.Quote(baseLine))
				break
			}
		}
	}

	if violations != 0 {
		fmt.Fprintf(stdout, "HAND-NOTES-DRIFT: %s — %d violation(s); make every harness file's hand-maintained sections identical\n", homeDir, violations)
		return 1
	}
	fmt.Fprintf(stdout, "HAND-NOTES-OK: %s — %d harness file(s), hand-maintained sections in step\n", homeDir, len(existing))
	return 0
}

// hnhHandContent strips every `<!-- flow:begin -->` … `<!-- flow:end -->`
// pair, delimiters included, and reports whether the body ended outside any
// block. Delimiters match on the trimmed whole line, the way the renderer
// writes them.
func hnhHandContent(body string) (string, bool) {
	var b strings.Builder
	inBlock := false
	for _, line := range strings.SplitAfter(body, "\n") {
		switch {
		case inBlock:
			if strings.TrimSpace(line) == "<!-- flow:end -->" {
				inBlock = false
			}
		case strings.TrimSpace(line) == "<!-- flow:begin -->":
			inBlock = true
		default:
			b.WriteString(line)
		}
	}
	return b.String(), !inBlock
}
