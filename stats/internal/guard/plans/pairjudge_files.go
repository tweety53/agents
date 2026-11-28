package plans

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// pjNamedPath is a file reference in a digest: a backticked
// `<abs-path>#<heading text>` (submatches 1 and 2), or an absolute path
// (submatch 3) opening the text or after a space, a backtick, a quote or a
// bracket, running to the next such character.
var pjNamedPath = regexp.MustCompile("`(/[^`#\\n]+)#([^`\\n]+)`|(?:^|[\\s`'\"(<\\[])(/[^\\s`'\"()<>\\[\\]]+)")

// pjPathLines splits a bare reference into its path and the `:a` or `:a-b`
// lines it names; a `:col` and trailing punctuation belong to neither.
var pjPathLines = regexp.MustCompile(`^(.*?)(?::(\d+)(?:-(\d+))?)?(?::\d+)*[.,;:!?]*$`)

// pjHeading is a markdown heading line: its level and text.
var pjHeading = regexp.MustCompile(`^(#{1,6})[ \t]+(.*?)(?:[ \t]+#+)?[ \t]*$`)

// pjSecretName is a file name never sent, obvious secrets only: private keys
// and key stores, and files named secret, secrets or credentials before the
// extension; .env files are refused apart.
var pjSecretName = regexp.MustCompile(`(?i)\.(key|pem|p12|pfx|jks|keystore)$|^id_(rsa|dsa|ecdsa|ed25519)$|^(secret|secrets|credentials)(\.[^.]*)?$`)

// pjHead is body when it fits pjFileCap, else its head -- cut after the last
// whole line, or else the last whole rune -- followed by pjFileTruncated.
// body is valid UTF-8.
func pjHead(body []byte) []byte {
	if len(body) <= pjFileCap {
		return body
	}
	n := bytes.LastIndexByte(body[:pjFileCap], '\n') + 1
	if n == 0 {
		for n = pjFileCap; !utf8.RuneStart(body[n]); n-- {
		}
	}
	head := body[:n:n]
	if head[n-1] != '\n' {
		head = append(head, '\n')
	}
	return append(head, pjFileTruncated...)
}

// pjBriefTasks finds the task numbers a brief names: an implementer brief's
// `Tasks: 7, 8, 9 of 20` line and a reviewer brief's `PASS task-7:` lines.
var pjBriefTasks = regexp.MustCompile(`(?m)^Tasks:[ \t]*([0-9][0-9, \t]*)|^PASS task-([0-9]+):`)

// pjNumber is one task number in a pjBriefTasks match.
var pjNumber = regexp.MustCompile(`[0-9]+`)

// pjTaskIDs is the task numbers digest names, as a set.
func pjTaskIDs(digest []byte) map[string]bool {
	ids := map[string]bool{}
	for _, m := range pjBriefTasks.FindAllSubmatch(digest, -1) {
		for _, id := range pjNumber.FindAll(append(m[1], m[2]...), -1) {
			ids[string(id)] = true
		}
	}
	return ids
}

// pjTasks is, for a file real named tasks.md, the sections of its body (IterTasks' grammar: the task
// line to the next task or `##`/`###` heading) whose number ids holds, each
// cut to its own pjHead, in file order; nil when ids is empty or the file
// holds none of them.
func pjTasks(real string, body []byte, ids map[string]bool) []byte {
	if len(ids) == 0 || filepath.Base(real) != "tasks.md" {
		return nil
	}
	lines, err := Lines(body)
	if err != nil {
		return nil
	}
	var out []byte
	for _, t := range IterTasks(lines) {
		if ids[t.Id] {
			sec := strings.Join(append([]string{lines[t.taskLine-1]}, t.lines...), "\n") + "\n"
			out = append(out, pjHead([]byte(sec))...)
		}
	}
	return out
}

// pjSection is the markdown section of body whose heading text is want, up to
// the next heading of the same or a higher level; backticks are ignored on
// both sides, and a line inside a code fence is no heading. nil when absent.
func pjSection(body []byte, want string) []byte {
	plain := func(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "`", "")) }
	var out []byte
	level, fence := 0, ""
	for _, line := range bytes.SplitAfter(body, []byte("\n")) {
		text := strings.TrimRight(string(line), "\r\n")
		if mark := strings.TrimSpace(text); fence != "" {
			if strings.HasPrefix(mark, fence) {
				fence = ""
			}
		} else if strings.HasPrefix(mark, "```") || strings.HasPrefix(mark, "~~~") {
			fence = mark[:3]
		} else if m := pjHeading.FindStringSubmatch(text); m != nil {
			if level > 0 && len(m[1]) <= level {
				break
			}
			if level == 0 && plain(m[2]) == plain(want) {
				level = len(m[1])
			}
		}
		if level > 0 {
			out = append(out, line...)
		}
	}
	return out
}

// pjLines is lines from through to (empty: from alone) of body, 1-based, to
// clamped to the last line; nil when the range holds none.
func pjLines(body []byte, from, to string) []byte {
	if to == "" {
		to = from
	}
	a, errA := strconv.Atoi(from)
	b, errB := strconv.Atoi(to)
	lines := bytes.SplitAfter(body, []byte("\n"))
	if errA != nil || errB != nil || a < 1 || b < a || a > len(lines) {
		return nil
	}
	return bytes.Join(lines[a-1:min(b, len(lines))], nil)
}

// pjAppendNamed appends to digest, for each reference in it to a file
// pjSendable admits, the part it names -- a pjSection or pjLines, cut to its
// pjHead, under a `--- <reference> ---` header; for a file named tasks.md with
// no part, the pjTasks the digest names, under `--- <path> ---` -- each once;
// a file named without a part, and a part it lacks, add nothing. All within
// pjDigestCap: the digest comes first and is never cut for a file; the file
// that crosses that total is cut there and followed by pjTruncated, and no
// file follows it.
func pjAppendNamed(env core.Env, digestFile string, digest []byte) []byte {
	self, _ := filepath.EvalSymlinks(digestFile)
	seen, ids := map[string]bool{}, pjTaskIDs(digest)
	for _, m := range pjNamedPath.FindAllSubmatch(digest, -1) {
		p, heading := string(m[1]), string(m[2])
		var from, to, ref string
		if heading != "" {
			ref = p + "#" + heading
		} else {
			l := pjPathLines.FindStringSubmatch(string(m[3]))
			p, from, to, ref = l[1], l[2], l[3], l[1]
			if from != "" {
				ref += ":" + from
			}
			if to != "" {
				ref += "-" + to
			}
		}
		real, err := filepath.EvalSymlinks(p)
		key := real + ref[len(p):]
		if err != nil || real == self || seen[key] {
			continue
		}
		seen[key] = true
		body, ok := pjSendable(env, real)
		if !ok {
			continue
		}
		switch {
		case heading != "":
			body = pjHead(pjSection(body, heading))
		case from != "":
			body = pjHead(pjLines(body, from, to))
		default:
			body = pjTasks(real, body, ids)
		}
		if len(body) == 0 {
			continue
		}
		digest = append(digest, "\n\n--- "+ref+" ---\n"...)
		if len(digest)+len(body) > pjDigestCap {
			return append(append(digest, body[:max(0, pjDigestCap-len(digest))]...), pjTruncated...)
		}
		digest = append(digest, body...)
	}
	return digest
}

// pjSendable is real's content when it may leave the machine: a regular
// UTF-8 text file with no NUL, inside a git working tree, not gitignored
// unless under .superpowers/ or spectre/changes/, not a .env or secret-named file, and not under
// ~/.claude or ~/.config. real is already symlink-resolved.
func pjSendable(env core.Env, real string) ([]byte, bool) {
	name := filepath.Base(real)
	if strings.HasPrefix(name, ".env") || pjSecretName.MatchString(name) || !core.IsRegular(real) {
		return nil, false
	}
	if home, err := filepath.EvalSymlinks(env.Getenv("HOME")); err == nil && home != "" {
		for _, d := range []string{".claude", ".config"} {
			if pjUnder(real, filepath.Join(home, d)) {
				return nil, false
			}
		}
	}
	dir := filepath.Dir(real)
	top, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, false
	}
	root, err := filepath.EvalSymlinks(strings.TrimSpace(string(top)))
	if err != nil || !pjUnder(real, root) {
		return nil, false
	}
	if exec.Command("git", "-C", dir, "check-ignore", "-q", real).Run() == nil &&
		!pjIgnoredSendable(filepath.ToSlash(real[len(root):])) {
		return nil, false
	}
	body, err := os.ReadFile(real)
	if err != nil || bytes.IndexByte(body, 0) >= 0 || !utf8.Valid(body) {
		return nil, false
	}
	return body, true
}

// pjIgnoredSendable is whether a gitignored path, relative to its tree's root
// with a leading slash, is still sent: run artifacts under .superpowers/ and
// spectre change artifacts under spectre/changes/.
func pjIgnoredSendable(rel string) bool {
	return strings.Contains(rel+"/", "/.superpowers/") || strings.HasPrefix(rel, "/spectre/changes/")
}

func pjUnder(p, dir string) bool {
	return p == dir || strings.HasPrefix(p, dir+string(filepath.Separator))
}
