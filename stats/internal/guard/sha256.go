package guard

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// sha256Hex is the lowercase hex SHA-256 of the string's bytes -- the retired
// scripts/lib/sha256-hex.sh's sha256_hex, whose only failure (no SHA-256 tool
// on the machine) cannot happen here.
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// sha256HexFile is the retired sha256_hex_file: the lowercase hex SHA-256 of a file's
// bytes, streamed from the file so NUL bytes survive. An unreadable file is
// an error, where the bash helper returned non-zero.
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
