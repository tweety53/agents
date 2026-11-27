/*
Copyright 2012 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Vendored from github.com/mautrix/go-util/shlex/shlex.go at 68f8502489a52dd2c3ea10f90f67fc4e9edc3911
// (MPL-2.0, LICENSE-MPL-2.0), itself derived from github.com/google/shlex (Apache-2.0,
// LICENSE-APACHE-2.0), and modified here.
//
// Python compatibility: this package splits a command exactly as Python's
// shlex.shlex(s, posix=True, punctuation_chars=True) with whitespace_split = True does — the
// splitter hooks/block-unsafe-commands.py used, whose decisions the Go port must reproduce byte
// for byte. That differs from upstream in six ways, each a change to the state machine below:
//   - `();<>|&` are punctuation: a run of them is a token of its own, ending the word before it,
//     and the character that ends the run is pushed back to start the next token.
//   - `#` starts a comment anywhere outside quotes, mid-word included: it ends the word.
//   - inside double quotes a backslash escapes only `"` and `\`; before any other character it
//     is kept, so "a\$b" is a\$b.
//   - a backslash-newline outside quotes is a literal newline in the word, not a continuation.
//   - an unclosed quote is the error "No closing quotation" and a trailing backslash "No escaped
//     character", Python's messages; neither returns a partial token.
//   - whitespace is " \t\r\n" only, as upstream: Python's shlex treats no other character as space.

/*
Package shlex implements a simple lexer which splits input in to tokens using
shell-style rules for quoting and commenting.

The basic use case uses the default ASCII lexer to split a string into sub-strings:

	shlex.Split("one \"two three\" four") -> []string{"one", "two three", "four"}

To process a stream of strings:

	l := NewLexer(os.Stdin)
	for ; token, err := l.Next(); err != nil {
		// process token
	}

To access the raw token stream (which includes tokens for comments):

	  t := NewTokenizer(os.Stdin)
	  for ; token, err := t.Next(); err != nil {
		// process token
	  }
*/
package shlex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// TokenType is a top-level token classification: A word, space, comment, unknown.
type TokenType int

// runeTokenClass is the type of a UTF-8 character classification: A quote, space, escape.
type runeTokenClass int

// the internal state used by the lexer state machine
type lexerState int

// Token is a (type, value) pair representing a lexographical token.
type Token struct {
	tokenType TokenType
	value     string
}

// Equal reports whether tokens a, and b, are equal.
// Two tokens are equal if both their types and values are equal. A nil token can
// never be equal to another token.
func (a *Token) Equal(b *Token) bool {
	if a == nil || b == nil {
		return false
	}
	if a.tokenType != b.tokenType {
		return false
	}
	return a.value == b.value
}

// Named classes of UTF-8 runes
const (
	spaceRunes            = " \t\r\n"
	escapingQuoteRunes    = `"`
	nonEscapingQuoteRunes = "'"
	escapeRunes           = `\`
	commentRunes          = "#"
	punctuationRunes      = "();<>|&"
)

// Classes of rune token
const (
	unknownRuneClass runeTokenClass = iota
	spaceRuneClass
	escapingQuoteRuneClass
	nonEscapingQuoteRuneClass
	escapeRuneClass
	commentRuneClass
	punctuationRuneClass
	eofRuneClass
)

// Classes of lexographic token
const (
	UnknownToken TokenType = iota
	WordToken
	SpaceToken
	CommentToken
)

// Lexer state machine states
const (
	startState           lexerState = iota // no runes have been seen
	inWordState                            // processing regular runes in a word
	punctuationState                       // processing a run of punctuation runes
	escapingState                          // we have just consumed an escape rune; the next rune is literal
	quotingEscapingState                   // we are within a quoted string that supports escaping ("...")
	quotingState                           // we are within a string that does not support escaping ('...')
	commentState                           // we are within a comment (everything following an unquoted or unescaped #
)

// The errors Split returns, worded as Python's shlex raises them.
var (
	ErrNoClosingQuotation = errors.New("No closing quotation")
	ErrNoEscapedCharacter = errors.New("No escaped character")
)

// tokenClassifier is used for classifying rune characters.
type tokenClassifier map[rune]runeTokenClass

func (typeMap tokenClassifier) addRuneClass(runes string, tokenType runeTokenClass) {
	for _, runeChar := range runes {
		typeMap[runeChar] = tokenType
	}
}

// newDefaultClassifier creates a new classifier for ASCII characters.
func newDefaultClassifier() tokenClassifier {
	t := tokenClassifier{}
	t.addRuneClass(spaceRunes, spaceRuneClass)
	t.addRuneClass(escapingQuoteRunes, escapingQuoteRuneClass)
	t.addRuneClass(nonEscapingQuoteRunes, nonEscapingQuoteRuneClass)
	t.addRuneClass(escapeRunes, escapeRuneClass)
	t.addRuneClass(commentRunes, commentRuneClass)
	t.addRuneClass(punctuationRunes, punctuationRuneClass)
	return t
}

// ClassifyRune classifiees a rune
func (t tokenClassifier) ClassifyRune(runeVal rune) runeTokenClass {
	return t[runeVal]
}

// Lexer turns an input stream into a sequence of tokens. Whitespace and comments are skipped.
type Lexer Tokenizer

// NewLexer creates a new lexer from an input stream.
func NewLexer(r io.Reader) *Lexer {
	return (*Lexer)(NewTokenizer(r))
}

// Next returns the next word, or an error. If there are no more words,
// the error will be io.EOF.
func (l *Lexer) Next() (string, error) {
	for {
		token, err := (*Tokenizer)(l).Next()
		if err != nil {
			return "", err
		}
		switch token.tokenType {
		case WordToken:
			return token.value, nil
		case CommentToken:
			// skip comments
		default:
			return "", fmt.Errorf("unknown token type: %v", token.tokenType)
		}
	}
}

// Tokenizer turns an input stream into a sequence of typed tokens
type Tokenizer struct {
	input      bufio.Reader
	classifier tokenClassifier
	pushback   []rune // runes read past the end of a token, read again first
}

// NewTokenizer creates a new tokenizer from an input stream.
func NewTokenizer(r io.Reader) *Tokenizer {
	input := bufio.NewReader(r)
	classifier := newDefaultClassifier()
	return &Tokenizer{
		input:      *input,
		classifier: classifier,
	}
}

// readRune is the next rune and its class: a pushed-back one first, eofRuneClass at the end.
func (t *Tokenizer) readRune() (rune, runeTokenClass, error) {
	if n := len(t.pushback); n > 0 {
		r := t.pushback[n-1]
		t.pushback = t.pushback[:n-1]
		return r, t.classifier.ClassifyRune(r), nil
	}
	r, _, err := t.input.ReadRune()
	if err == io.EOF {
		return 0, eofRuneClass, nil
	}
	if err != nil {
		return 0, unknownRuneClass, err
	}
	return r, t.classifier.ClassifyRune(r), nil
}

// scanStream scans the stream for the next token using the internal state machine.
func (t *Tokenizer) scanStream() (*Token, error) {
	state := startState
	escapedFrom := inWordState // the state an escape returns to
	var tokenType TokenType
	var value []rune
	word := func() (*Token, error) { return &Token{tokenType: tokenType, value: string(value)}, nil }

	for {
		nextRune, nextRuneType, err := t.readRune()
		if err != nil {
			return nil, err
		}

		switch state {
		case startState: // no runes read yet
			switch nextRuneType {
			case eofRuneClass:
				return nil, io.EOF
			case spaceRuneClass:
			case escapingQuoteRuneClass:
				tokenType = WordToken
				state = quotingEscapingState
			case nonEscapingQuoteRuneClass:
				tokenType = WordToken
				state = quotingState
			case escapeRuneClass:
				tokenType = WordToken
				escapedFrom = inWordState
				state = escapingState
			case commentRuneClass:
				tokenType = CommentToken
				state = commentState
			case punctuationRuneClass:
				tokenType = WordToken
				value = append(value, nextRune)
				state = punctuationState
			default:
				tokenType = WordToken
				value = append(value, nextRune)
				state = inWordState
			}
		case inWordState, punctuationState: // in a word, or in a run of punctuation
			switch {
			case nextRuneType == eofRuneClass || nextRuneType == spaceRuneClass:
				return word()
			case nextRuneType == commentRuneClass:
				t.pushback = append(t.pushback, nextRune)
				return word()
			case state == punctuationState:
				if nextRuneType == punctuationRuneClass {
					value = append(value, nextRune)
				} else {
					t.pushback = append(t.pushback, nextRune)
					return word()
				}
			case nextRuneType == escapingQuoteRuneClass:
				state = quotingEscapingState
			case nextRuneType == nonEscapingQuoteRuneClass:
				state = quotingState
			case nextRuneType == escapeRuneClass:
				escapedFrom = inWordState
				state = escapingState
			case nextRuneType == punctuationRuneClass:
				t.pushback = append(t.pushback, nextRune)
				return word()
			default:
				value = append(value, nextRune)
			}
		case escapingState: // the rune after an escape character
			if nextRuneType == eofRuneClass {
				return nil, ErrNoEscapedCharacter
			}
			if escapedFrom == quotingEscapingState && nextRune != '"' && nextRune != '\\' {
				value = append(value, '\\')
			}
			value = append(value, nextRune)
			state = escapedFrom
		case quotingEscapingState: // in escaping double quotes
			switch nextRuneType {
			case eofRuneClass:
				return nil, ErrNoClosingQuotation
			case escapingQuoteRuneClass:
				state = inWordState
			case escapeRuneClass:
				escapedFrom = quotingEscapingState
				state = escapingState
			default:
				value = append(value, nextRune)
			}
		case quotingState: // in non-escaping single quotes
			switch nextRuneType {
			case eofRuneClass:
				return nil, ErrNoClosingQuotation
			case nonEscapingQuoteRuneClass:
				state = inWordState
			default:
				value = append(value, nextRune)
			}
		case commentState: // in a comment
			if nextRuneType == eofRuneClass || nextRune == '\n' {
				return word()
			}
			value = append(value, nextRune)
		default:
			return nil, fmt.Errorf("unexpected state: %v", state)
		}
	}
}

// Next returns the next token in the stream.
func (t *Tokenizer) Next() (*Token, error) {
	return t.scanStream()
}

// Split partitions a string into a slice of strings.
func Split(s string) ([]string, error) {
	l := NewLexer(strings.NewReader(s))
	subStrings := make([]string, 0)
	for {
		word, err := l.Next()
		if err != nil {
			if err == io.EOF {
				return subStrings, nil
			}
			return subStrings, err
		}
		subStrings = append(subStrings, word)
	}
}
