package guard

import (
	"os"
	"os/user"
	"strings"
	"unicode"
)

// Python's posixpath, str and re character classes as the ported hooks use them (CPython 3.14):
// each helper names the Python call it reproduces, and differs from it only where its comment
// says so.

// pyReS is the character class of Python's `\s` in a str pattern, for use inside `[...]`:
// str.isspace()'s characters, where Go's `\s` is ASCII only.
const pyReS = `\t\n\x0B\f\r\x1C-\x1F \x{85}\x{A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}`

// pyReW is the character class of Python's `\w` in a str pattern, for use inside `[...]`.
// ponytail: letters and numbers of Go's Unicode tables, which may trail Python's by a version.
const pyReW = `\p{L}\p{N}_`

// pyIsWord is one character of Python's `\w`.
func pyIsWord(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) }

// pyDigitOnly holds the characters Python's str.isdigit() accepts beyond unicode.Nd
// (superscripts, circled digits, …), from CPython 3.14's Unicode 16 database.
var pyDigitOnly = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0xB2, 0xB3, 1}, {0xB9, 0xB9, 1}, {0x1369, 0x1371, 1}, {0x19DA, 0x19DA, 1}, {0x2070, 0x2070, 1},
		{0x2074, 0x2079, 1}, {0x2080, 0x2089, 1}, {0x2460, 0x2468, 1}, {0x2474, 0x247C, 1},
		{0x2488, 0x2490, 1}, {0x24EA, 0x24EA, 1}, {0x24F5, 0x24FD, 1}, {0x24FF, 0x24FF, 1},
		{0x2776, 0x277E, 1}, {0x2780, 0x2788, 1}, {0x278A, 0x2792, 1},
	},
	R32: []unicode.Range32{
		{0x10A40, 0x10A43, 1}, {0x10E60, 0x10E68, 1}, {0x11052, 0x1105A, 1}, {0x1F100, 0x1F10A, 1},
	},
}

// pyIsDigit is Python's s.isdigit(): non-empty, every character a digit.
func pyIsDigit(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.Nd, r) && !unicode.Is(pyDigitOnly, r) {
			return false
		}
	}
	return s != ""
}

// pyIntDigits is Python's str(int(s)) for s a run of `\d` (unicode.Nd) characters: each digit's
// value, leading zeros dropped. Nd characters come in contiguous runs of ten from zero.
func pyIntDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		zero := r
		for unicode.Is(unicode.Nd, zero-1) {
			zero--
		}
		b.WriteByte(byte('0' + (r-zero)%10))
	}
	if t := strings.TrimLeft(b.String(), "0"); t != "" {
		return t
	}
	return "0"
}

// pySplitLines is Python's s.splitlines().
func pySplitLines(s string) []string {
	var out []string
	start := 0
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		switch rs[i] {
		case '\n', '\r', '\v', '\f', '\x1c', '\x1d', '\x1e', '\u0085', '\u2028', '\u2029':
			out = append(out, string(rs[start:i]))
			if rs[i] == '\r' && i+1 < len(rs) && rs[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	if start < len(rs) {
		out = append(out, string(rs[start:]))
	}
	return out
}

// pyJoin is os.path.join(a, *p).
func pyJoin(a string, p ...string) string {
	for _, b := range p {
		switch {
		case strings.HasPrefix(b, "/"):
			a = b
		case a == "" || strings.HasSuffix(a, "/"):
			a += b
		default:
			a += "/" + b
		}
	}
	return a
}

// pyDirname is os.path.dirname(p).
func pyDirname(p string) string {
	head := p[:strings.LastIndex(p, "/")+1]
	if head != "" && head != strings.Repeat("/", len(head)) {
		head = strings.TrimRight(head, "/")
	}
	return head
}

// pyBasename is os.path.basename(p).
func pyBasename(p string) string { return p[strings.LastIndex(p, "/")+1:] }

// pyNormpath is os.path.normpath(p).
func pyNormpath(p string) string {
	if p == "" {
		return "."
	}
	initial := 0
	if strings.HasPrefix(p, "/") {
		initial = 1
		if strings.HasPrefix(p, "//") && !strings.HasPrefix(p, "///") {
			initial = 2
		}
	}
	var comps []string
	for _, c := range strings.Split(p, "/") {
		if c == "" || c == "." {
			continue
		}
		if c != ".." || (initial == 0 && len(comps) == 0) || (len(comps) > 0 && comps[len(comps)-1] == "..") {
			comps = append(comps, c)
		} else if len(comps) > 0 {
			comps = comps[:len(comps)-1]
		}
	}
	out := strings.Repeat("/", initial) + strings.Join(comps, "/")
	if out == "" {
		return "."
	}
	return out
}

// pyAbspath is os.path.abspath(p), cwd standing for os.getcwd().
func pyAbspath(p, cwd string) string {
	if !strings.HasPrefix(p, "/") {
		p = pyJoin(cwd, p)
	}
	return pyNormpath(p)
}

// pyRealpath is os.path.realpath(p) (strict=False), cwd standing for os.getcwd(). A NUL in a
// path it lstats panics, as os.lstat's ValueError escapes realpath.
func pyRealpath(p, cwd string) string {
	parts := strings.Split(p, "/")
	type part struct {
		name   string
		marker bool // a resolved symlink's own path follows: record it in seen
	}
	var rest []part
	for i := len(parts) - 1; i >= 0; i-- {
		rest = append(rest, part{name: parts[i]})
	}
	count := len(rest)
	path := cwd
	if strings.HasPrefix(p, "/") {
		path = "/"
	}
	seen := map[string]*string{}
	pop := func() part {
		x := rest[len(rest)-1]
		rest = rest[:len(rest)-1]
		return x
	}
	for count > 0 {
		x := pop()
		if x.marker {
			resolved := path
			seen[pop().name] = &resolved
			continue
		}
		count--
		name := x.name
		if name == "" || name == "." {
			continue
		}
		if name == ".." {
			path = path[:strings.LastIndex(path, "/")]
			if path == "" {
				path = "/"
			}
			continue
		}
		newpath := path + "/" + name
		if path == "/" {
			newpath = path + name
		}
		if strings.ContainsRune(newpath, 0) {
			panic("realpath: embedded null byte")
		}
		fi, err := os.Lstat(newpath)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			path = newpath
			continue
		}
		if s, ok := seen[newpath]; ok {
			if s != nil {
				path = *s
				continue
			}
			path = newpath // a symlink loop
			continue
		}
		target, err := os.Readlink(newpath)
		if err != nil {
			path = newpath
			continue
		}
		if strings.HasPrefix(target, "/") {
			path = "/"
		}
		seen[newpath] = nil
		rest = append(rest, part{name: newpath}, part{marker: true})
		tparts := strings.Split(target, "/")
		for i := len(tparts) - 1; i >= 0; i-- {
			rest = append(rest, part{name: tparts[i]})
		}
		count += len(tparts)
	}
	return path
}

// pyIsDir is os.path.isdir(p).
func pyIsDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// pyExists is os.path.exists(p).
func pyExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// pyIsLink is os.path.islink(p).
func pyIsLink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// pyExpandUser is os.path.expanduser(p).
// ponytail: a `~user` lookup goes through os/user, which without cgo reads /etc/passwd only.
func pyExpandUser(env Env, p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	i := strings.Index(p[1:], "/") + 1
	if i == 0 {
		i = len(p)
	}
	var home string
	if i == 1 {
		h, ok := "", false
		if env.LookupEnv != nil {
			h, ok = env.LookupEnv("HOME")
		} else if h = env.Getenv("HOME"); h != "" {
			ok = true
		}
		if ok {
			home = h
		} else if u, err := user.Current(); err == nil {
			home = u.HomeDir
		} else {
			return p
		}
	} else {
		u, err := user.Lookup(p[1:i])
		if err != nil {
			return p
		}
		home = u.HomeDir
	}
	if out := strings.TrimRight(home, "/") + p[i:]; out != "" {
		return out
	}
	return "/"
}
