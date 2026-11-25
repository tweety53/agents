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

// pjAppendNamed appends to digest each file it names that pjSendable admits,
// once, under a `--- <path> ---` header, each cut to its pjHead, within
// pjDigestCap: the digest comes first and is never cut for a file; the file
// that crosses that total is cut there and followed by pjTruncated, and no
// file follows it.
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
		body = pjHead(body)
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
