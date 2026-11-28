package core

import (
	"strings"
	"unicode"
)

// PyIsSpace is Python's str.isspace for one code point: Go's unicode.IsSpace
// plus the four information separators Python also counts.
func PyIsSpace(r rune) bool { return unicode.IsSpace(r) || r >= 0x1C && r <= 0x1F }

func PyStrip(s string) string { return strings.TrimFunc(s, PyIsSpace) }

// PySplit is Python's str.split() with no arguments.
func PySplit(s string) []string { return strings.FieldsFunc(s, PyIsSpace) }
