package core

import (
	"strings"
	"unicode"
)

// IsSpaceOrSep is unicode.IsSpace plus the four information separators
// U+001C-U+001F, which the guards' text grammars also read as space.
func IsSpaceOrSep(r rune) bool { return unicode.IsSpace(r) || r >= 0x1C && r <= 0x1F }

// TrimSpaceOrSep is strings.TrimSpace over IsSpaceOrSep.
func TrimSpaceOrSep(s string) string { return strings.TrimFunc(s, IsSpaceOrSep) }

// FieldsSpaceOrSep is strings.Fields over IsSpaceOrSep.
func FieldsSpaceOrSep(s string) []string { return strings.FieldsFunc(s, IsSpaceOrSep) }
