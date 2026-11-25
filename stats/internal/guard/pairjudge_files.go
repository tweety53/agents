package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// pjNamedPath is an absolute path in a digest: opening the text or after a
// space, a backtick, a quote or a bracket, running to the next such character.
var pjNamedPath = regexp.MustCompile("(?:^|[\\s`'\"(<\\[])(/[^\\s`'\"()<>\\[\\]]+)")

// pjPathTail is what may follow a named path without being part of it: a
// `:line[:col]` and trailing punctuation.
var pjPathTail = regexp.MustCompile(`(?::\d+)*[.,;:!?]*$`)

// pjSecretName is a file name never sent: key, credential, secret, password
// or token files and key-store extensions; .env files are refused apart.
var pjSecretName = regexp.MustCompile(`(?i)key|secret|credential|passw|token|\.(pem|p12|pfx|der|crt|jks|keystore|gpg|asc)$`)

// pjAppendNamed appends to digest each file it names that pjSendable admits,
// once, under a `--- <path> ---` header, within pjDigestCap: the digest
// comes first and is never cut for a file; the file that crosses the cap is
// cut there and followed by pjTruncated, and no file follows it.
func pjAppendNamed(env Env, digestFile string, digest []byte) []byte {
	self, _ := filepath.EvalSymlinks(digestFile)
	seen := map[string]bool{self: true}
	for _, m := range pjNamedPath.FindAllSubmatch(digest, -1) {
		p := pjPathTail.ReplaceAllString(string(m[1]), "")
		real, err := filepath.EvalSymlinks(p)
		if err != nil || seen[real] {
			continue
		}
		seen[real] = true
		body, ok := pjSendable(env, real)
		if !ok {
			continue
		}
		digest = append(digest, "\n\n--- "+p+" ---\n"...)
		if len(digest)+len(body) > pjDigestCap {
			return append(append(digest, body[:max(0, pjDigestCap-len(digest))]...), pjTruncated...)
		}
		digest = append(digest, body...)
	}
	return digest
}

// pjSendable is real's content when it may leave the machine: a regular
// UTF-8 text file with no NUL, inside a git working tree, not gitignored
// unless under .superpowers/, not a .env or secret-named file, and not under
// ~/.claude or ~/.config. real is already symlink-resolved.
func pjSendable(env Env, real string) ([]byte, bool) {
	name := filepath.Base(real)
	if strings.HasPrefix(name, ".env") || pjSecretName.MatchString(name) || !isRegular(real) {
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
		!strings.Contains(filepath.ToSlash(real[len(root):])+"/", "/.superpowers/") {
		return nil, false
	}
	body, err := os.ReadFile(real)
	if err != nil || bytes.IndexByte(body, 0) >= 0 || !utf8.Valid(body) {
		return nil, false
	}
	return body, true
}

func pjUnder(p, dir string) bool {
	return p == dir || strings.HasPrefix(p, dir+string(filepath.Separator))
}
