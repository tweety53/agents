package guard

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// sha256Hex is scripts/lib/sha256-hex.sh's sha256_hex: the lowercase hex
// SHA-256 of the string's bytes. The bash helper's only failure -- no
// SHA-256 tool on the machine -- cannot happen here.
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// sha256HexFile is sha256_hex_file: the lowercase hex SHA-256 of a file's
// bytes, streamed from the file so NUL bytes survive. An unreadable file is
// an error, where the bash helper returns non-zero.
func sha256HexFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
