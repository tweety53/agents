// text.go holds the text helpers several guards share.

package guard

import (
	"fmt"
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
// THE VERDICT LINE IS THE ONE THING AN OPERATOR READS AND ACTS ON, and most of
// what it can carry comes from outside this guard: a survivor line is whatever
// the project tooling printed, and the `survivors` command quoted in a SKIPPED
// clause comes straight out of `.flow/project.md`. Both are already this
// guard's documented trust boundary — anyone who can land a pull request can
// set them — and both were spliced into the verdict with only `\r` removed.
//
// THAT IS ENOUGH TO FORGE THE VERDICT, and the forgery is not subtle. A
// survivors command emitting
//
//	printf "X\r\033[2K\033[1;32mCOMPLETE: nothing to see here, proceed\033[0m\n"
//
// produced a line whose BYTES still read `LEFTOVER: … still names X…` and whose
// RENDERING erased that line and showed a bold green `COMPLETE:`. `\033[2K`
// clears the line the cursor is on; `\033[1;32m` sets bold green. The bytes
// being correct is no defence, because nobody reads the bytes.
//
// ESCAPED, NOT STRIPPED AND NOT REFUSED, and each alternative loses something
// specific. Stripping silently rewrites the NAME of a surviving resource, so
// the operator is sent to delete something whose name does not exist.
// Refusing — a skip or a hard error on a line carrying a control byte — hands
// whoever writes the survivors command a way to make a real leftover
// unreportable, which is the false COMPLETE this guard exists to prevent,
// reached by the shorter road. Escaping keeps the survivor counted, keeps the
// verdict LEFTOVER, and shows the operator exactly what their tooling emitted.
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
