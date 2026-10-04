package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// shTok is a shell word, quotes removed, or a control operator, with the
// 1-based line it starts on.
type shTok struct {
	s    string
	op   bool
	line int
}

var (
	heredocDelim = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	assignment   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
	grepName     = regexp.MustCompile(`^[ef]?grep$`)
)

// earlyExitOpt matches a grep option that stops reading at the first match:
// a short-option cluster carrying q, m, l or L before any e or f, whose
// remainder is their argument, or a long form.
var earlyExitOpt = regexp.MustCompile(`^(?:-[A-Za-dg-z0-9]*[qmlL]|--quiet$|--silent$|--max-count(?:=|$)|--files-with(?:-matches|out-match)$)`)

// argOpt matches a grep option whose argument is the next token: `-e`/`-f`,
// alone or ending a short-option cluster, and their long forms.
var argOpt = regexp.MustCompile(`^(?:-[A-Za-z0-9]*[ef]|--regexp|--file)$`)

// shellTokens splits src into words and control operators as bash does, as
// far as this guard needs: quotes group a word and are removed, a backslash
// escapes the next byte and joins a continued line, a `#` opening a word
// starts a comment, `$(` inside double quotes nests a command, and a heredoc
// body is skipped. Redirections stay inside words; `&&`, `;;` and `>&` lex
// as single-byte operators, which end a command all the same.
func shellTokens(src string) []shTok {
	var (
		toks           []shTok
		word           strings.Builder
		inWord, dq     bool
		line, wordLine = 1, 1
		nest           []bool   // per open `(`: whether its `)` resumes a double-quoted word
		delims         []string // heredocs whose bodies follow the next newline
		pendingDelim   bool     // a bare `<<` was the last word; the next one names the heredoc
	)
	start := func() {
		if !inWord {
			inWord, wordLine = true, line
		}
	}
	add := func(c byte) {
		start()
		word.WriteByte(c)
		if c == '\n' {
			line++
		}
	}
	flush := func() {
		if !inWord {
			return
		}
		w := word.String()
		d, isHeredoc := strings.CutPrefix(w, "<<")
		switch {
		case pendingDelim:
			pendingDelim, d, isHeredoc = false, w, true
		case isHeredoc && strings.HasPrefix(d, "<"):
			isHeredoc = false
		case isHeredoc:
			d = strings.TrimPrefix(d, "-")
			pendingDelim = d == ""
		}
		if isHeredoc && heredocDelim.MatchString(d) {
			delims = append(delims, d)
		}
		toks = append(toks, shTok{s: w, line: wordLine})
		word.Reset()
		inWord = false
	}
	op := func(s string) {
		flush()
		toks = append(toks, shTok{s: s, op: true, line: line})
	}
	next := func(i int) byte {
		if i+1 < len(src) {
			return src[i+1]
		}
		return 0
	}
	for i := 0; i < len(src); i++ {
		c := src[i]
		if dq {
			switch {
			case c == '"':
				dq = false
			case c == '\\' && next(i) == '\n':
				i++
				line++
			case c == '\\' && next(i) != 0 && strings.IndexByte("\"\\$`", next(i)) >= 0:
				i++
				add(src[i])
			case c == '$' && next(i) == '(':
				i++
				dq = false
				nest = append(nest, true)
				op("(")
			default:
				add(c)
			}
			continue
		}
		switch c {
		case '\'':
			start()
			end := strings.IndexByte(src[i+1:], '\'')
			if end < 0 {
				end = len(src) - i - 1
			}
			for _, b := range []byte(src[i+1 : i+1+end]) {
				add(b)
			}
			i += end + 1
		case '"':
			start()
			dq = true
		case '\\':
			if next(i) == '\n' {
				line++
			} else if next(i) != 0 {
				add(next(i))
			}
			i++
		case '#':
			if inWord {
				add(c)
				continue
			}
			for i+1 < len(src) && src[i+1] != '\n' {
				i++
			}
		case ' ', '\t':
			flush()
		case '\n':
			op("\n")
			line++
			for _, d := range delims {
				for i+1 < len(src) {
					body := src[i+1:]
					end := strings.IndexByte(body, '\n')
					if end < 0 {
						i = len(src)
						break
					}
					i += end + 1
					line++
					if strings.TrimSpace(body[:end]) == d {
						break
					}
				}
			}
			delims = nil
		case '|':
			switch next(i) {
			case '|', '&':
				op(src[i : i+2])
				i++
			default:
				op("|")
			}
		case '&', ';':
			op(string(c))
		case '(':
			nest = append(nest, false)
			op("(")
		case ')':
			op(")")
			if n := len(nest); n > 0 {
				dq = nest[n-1]
				nest = nest[:n-1]
			}
		default:
			add(c)
		}
	}
	flush()
	return toks
}

// grepPipeHits returns the 1-based lines of src where a pipe feeds a grep
// that can exit before reading all of its input.
func grepPipeHits(src string) []int {
	toks := shellTokens(src)
	var hits []int
	for i, t := range toks {
		if !t.op || (t.s != "|" && t.s != "|&") {
			continue
		}
		j := i + 1
		for j < len(toks) && toks[j].op && toks[j].s == "\n" {
			j++
		}
		for j < len(toks) && !toks[j].op && (toks[j].s == "command" || toks[j].s == "env" || assignment.MatchString(toks[j].s)) {
			j++
		}
		if j == len(toks) || toks[j].op || !grepName.MatchString(toks[j].s) {
			continue
		}
		for k := j + 1; k < len(toks) && !toks[k].op && toks[k].s != "--"; k++ {
			if earlyExitOpt.MatchString(toks[k].s) {
				hits = append(hits, toks[j].line)
				break
			}
			if argOpt.MatchString(toks[k].s) && k+1 < len(toks) && !toks[k+1].op {
				k++
			}
		}
	}
	return hits
}
func TestGrepPipeHits(t *testing.T) {
	t.Parallel()
	cases := []struct {
		src  string
		want int
	}{
		{`printf '%s\n' "$OUT" | grep -Eq '^ok:'`, 1},
		{`echo "$ERR" | grep -q "x"`, 1},
		{`lsof -Fn | grep -qx -- "n$2"`, 1},
		{`cmd | grep -E -m1 x`, 1},
		{`cmd | grep --quiet x`, 1},
		{"cmd |\\\n  grep -q x", 1},
		{`grep -Eq '^ok:' <<<"$OUT"`, 0},
		{`cmd | grep -c x`, 0},
		{`[ "$(cmd | grep -c x)" -eq 1 ]`, 0},
		{`cmd | grep x -- -q`, 0},
		{`a || grep -q x f`, 0},
		{`write_cfg 'printf a\n \| grep -q b'`, 0},
		{`# never printf ... | grep -q`, 0},
		{`cmd | grep -v foo | grep -q bar`, 1},
		{`cmd | grep foo | grep -q bar`, 1},
		{`cmd | grep -e -q x`, 0},
		{`cmd | grep -f -m x`, 0},
		{`cmd | grep -qe x`, 1},
		{`cmd | grep -eq x`, 0},
		{`cmd | egrep -q x`, 1},
		{`cmd | command grep -q x`, 1},
		{`cmd | LC_ALL=C grep -q x`, 1},
		{`cmd | grep -E 'a|b' -q`, 1},
		{`x=1 # a | grep -q b`, 0},
		{`echo "# not a comment" | grep -q x`, 1},
		{"cmd |\n  grep -q x", 1},
		{"cmd |\n  # why\n\n  grep -q x", 1},
		{"a ||\n  grep -q x f", 0},
		{`cmd | grep 'foo -q bar'`, 0},
		{`echo 'a | grep -q b'`, 0},
		{`cmd | grep '-q' x`, 1},
		{`cmd |& grep -q x`, 1},
		{`cmd | grep -l x`, 1},
		{`cmd | grep --files-without-match x`, 1},
		{`x="$(cmd | grep -m1 y)"`, 1},
		{`cmd | grep -e; grep -q x f`, 0},
		{"cat <<'EOF'\nit's | grep -q x\nEOF\ncmd | grep -q y", 1},
		{"awk '\n  # it | grep -q x\n' f", 0},
	}
	for _, c := range cases {
		if got := len(grepPipeHits(c.src)); got != c.want {
			t.Errorf("grepPipeHits(%q) = %d hits, want %d", c.src, got, c.want)
		}
	}
}

// TestNoPipeIntoEarlyExitGrep fails any script — setup.sh, scripts/,
// scripts/lib/ — that pipes into `grep -q`/`-m`/`-l`: grep exits on its first
// match, the writer dies of SIGPIPE on its next write, and pipefail turns the
// pipeline's status into 141 — a match read as a miss, intermittently. Feed
// grep a here-string (`grep -q re <<<"$OUT"`), which has no writer process to
// kill. Every file is scanned, pipefail or not: a sourced lib runs under its
// caller's pipefail, and a script can turn pipefail on later.
func TestNoPipeIntoEarlyExitGrep(t *testing.T) {
	t.Parallel()
	var paths []string
	for _, g := range []string{"../../../setup.sh", "../../../scripts/*.sh", "../../../scripts/lib/*.sh"} {
		m, err := filepath.Glob(g)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, m...)
	}
	if len(paths) == 0 {
		t.Fatal("no scripts found under ../../../scripts")
	}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range grepPipeHits(string(src)) {
			t.Errorf("%s:%d: pipe into an early-exit grep; use a here-string", filepath.Base(path), line)
		}
	}
}
