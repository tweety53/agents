// text.go holds the text helpers several guards share.

package guard

import (
	"fmt"
	"os/exec"
	"strings"
)

// lines splits a file's content the way `while IFS= read -r line || [ -n
// "$line" ]` reads it: a trailing newline does not add an empty last line.
func lines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

// crSort is `sort` (or `sort -u`) over lines, exec'd rather than
// reimplemented: its order is the locale's collation (en_US.UTF-8 on macOS
// sorts "SKILL-rationale.md" after "session-records-rationale.md"), which
// no byte comparison reproduces, and the guard's output order is its.
func crSort(env Env, lines []string, unique bool) ([]string, error) {
	if unique {
		return crSortSep(env, lines, "\n", "-u")
	}
	return crSortSep(env, lines, "\n")
}

// ccSanitize is sanitize_display: every C0 byte but the newline that ends a
// line, DEL and backslash rendered visibly; bytes 0x80-0xFF pass through.
//
// A line an operator reads and acts on can carry text from outside the
// guard printing it — git output, a path, a `.flow/project.md` cell — and a
// control byte in it can rewrite what the terminal shows (`\r\033[2K` clears
// the line). ESCAPED, NOT STRIPPED: stripping silently rewrites a name, so
// the operator is sent to something that does not exist.
//
// THE ENCODING IS INJECTIVE, which is why `\` is escaped alongside the
// controls: otherwise a name containing the four literal characters `\x1b`
// and a name containing a real ESC byte print identically.
//
// C0 AND DEL, AND NOTHING ABOVE 0x7F. Bytes 0x80-0xFF are the lead and
// continuation bytes of ordinary UTF-8, so a bucket named `café-bücket` is
// reported by its real name; a UTF-8 terminal acts on none of them, so
// escaping them would mangle every non-ASCII report and buy nothing. The bash
// guard needed a non-exported `LC_ALL=C` prefix to make that byte test hold;
// a byte loop has no locale to pin.
func ccSanitize(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\':
			b.WriteString(`\\`)
		case c == '\n':
			b.WriteByte(c)
		case c < 0x20 || c == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// ccASCIILower is grep -i / awk tolower over an ASCII pattern: only A-Z fold.
func ccASCIILower(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 'a' - 'A'
		}
		return r
	}, s)
}

// smcUTF8 is whether bash's printf %q runs under a UTF-8 character type:
// LC_ALL, else LC_CTYPE, else LANG, the first one set and non-empty.
func smcUTF8(env Env) bool {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := env.Getenv(k); v != "" {
			v = strings.ToLower(v)
			return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
		}
	}
	return false
}

// crSortSep is crSort over lines joined and terminated by sep, sort run with
// args (`-z` for a NUL separator); the caller's collation variables are the
// only environment sort sees.
func crSortSep(env Env, lines []string, sep string, args ...string) ([]string, error) {
	if len(lines) < 2 {
		return lines, nil
	}
	bin, ok := lookPath(env, "sort")
	if !ok {
		return nil, fmt.Errorf("no sort on PATH")
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = []string{}
	for _, k := range []string{"LANG", "LC_ALL", "LC_COLLATE", "LC_CTYPE"} {
		if v, set := env.LookupEnv(k); set {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	cmd.Stdin = strings.NewReader(strings.Join(lines, sep) + sep)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sort: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(out), sep), sep), nil
}
