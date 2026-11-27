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

// Vendored from github.com/mautrix/go-util/shlex/shlex_test.go at 68f8502489a52dd2c3ea10f90f67fc4e9edc3911
// (MPL-2.0, LICENSE-MPL-2.0), itself derived from github.com/google/shlex (Apache-2.0,
// LICENSE-APACHE-2.0), and modified here: the expected tokens follow Python's shlex (shlex.go's
// Python-compatibility comment), each changed case commented; TestSplitMatchesPython is added.

package shlex

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var (
	// one two "three four" "five \"six\"" seven#eight # nine # ten
	// eleven 'twelve\' thirteen=13 fourteen/14 fif\
	// teen \
	//  sixteen seven\ teen
	testString = "one two \"three four\" \"five \\\"six\\\"\" seven#eight # nine # ten\n eleven 'twelve\\' thirteen=13 fourteen/14 fif\\\nteen \\\n sixteen seven\\ teen"
)

func TestClassifier(t *testing.T) {
	classifier := newDefaultClassifier()
	tests := map[rune]runeTokenClass{
		' ':  spaceRuneClass,
		'"':  escapingQuoteRuneClass,
		'\'': nonEscapingQuoteRuneClass,
		'#':  commentRuneClass,
		// Python compatibility: punctuation is a class of its own.
		';': punctuationRuneClass,
		'(': punctuationRuneClass,
	}
	for runeChar, want := range tests {
		got := classifier.ClassifyRune(runeChar)
		if got != want {
			t.Errorf("ClassifyRune(%v) -> %v. Want: %v", runeChar, got, want)
		}
	}
}

func TestTokenizer(t *testing.T) {
	testInput := strings.NewReader(testString)
	expectedTokens := []*Token{
		{WordToken, "one"},
		{WordToken, "two"},
		{WordToken, "three four"},
		{WordToken, "five \"six\""},
		// Python compatibility: a `#` mid-word starts a comment (upstream: "seven#eight", then
		// the comment " nine # ten").
		{WordToken, "seven"},
		{CommentToken, "eight # nine # ten"},
		{WordToken, "eleven"},
		{WordToken, "twelve\\"},
		{WordToken, "thirteen=13"},
		{WordToken, "fourteen/14"},
		// Python compatibility: a backslash-newline is a literal newline, not a continuation
		// (upstream: "fifteen", then nothing for the lone `\` + newline).
		{WordToken, "fif\nteen"},
		{WordToken, "\n"},
		{WordToken, "sixteen"},
		{WordToken, "seven teen"},
	}

	tokenizer := NewTokenizer(testInput)
	for i, want := range expectedTokens {
		got, err := tokenizer.Next()
		if err != nil {
			t.Error(err)
		}
		if !got.Equal(want) {
			t.Errorf("Tokenizer.Next()[%v] of %q -> %v. Want: %v", i, testString, got, want)
		}
	}
}

// Python compatibility: "seven" for upstream's "seven#eight", "fif\nteen" and "\n" for upstream's
// "fifteen" (TestTokenizer's comments).
var expectedSplit = []string{"one", "two", "three four", "five \"six\"", "seven", "eleven", "twelve\\", "thirteen=13", "fourteen/14", "fif\nteen", "\n", "sixteen", "seven teen"}

func TestLexer(t *testing.T) {
	testInput := strings.NewReader(testString)

	lexer := NewLexer(testInput)
	for i, want := range expectedSplit {
		got, err := lexer.Next()
		if err != nil {
			t.Error(err)
		}
		if got != want {
			t.Errorf("Lexer.Next()[%v] of %q -> %v. Want: %v", i, testString, got, want)
		}
	}
}

func TestSplit(t *testing.T) {
	got, err := Split(testString)
	if err != nil {
		t.Error(err)
	}
	if len(expectedSplit) != len(got) {
		t.Errorf("Split(%q) -> %v. Want: %v", testString, got, expectedSplit)
	}
	for i := range got {
		if got[i] != expectedSplit[i] {
			t.Errorf("Split(%q)[%v] -> %v. Want: %v", testString, i, got[i], expectedSplit[i])
		}
	}
}

// pythonSplit is the program that captured testdata/python-split.json: each input's
// shlex.shlex(s, posix=True, punctuation_chars=True) words with whitespace_split = True, or
// "error: <message>".
const pythonSplit = `
import json, shlex, sys
rows = []
for s, _ in json.load(open(sys.argv[1])):
    lx = shlex.shlex(s, posix=True, punctuation_chars=True)
    lx.whitespace_split = True
    try:
        rows.append([s, list(lx)])
    except ValueError as e:
        rows.append([s, "error: " + str(e)])
open(sys.argv[1], "w").write(json.dumps(rows, indent=1, ensure_ascii=False) + "\n")
`

// TestSplitMatchesPython: Split returns Python's words, or Python's error, on every input of
// testdata/python-split.json. SHLEX_GOLDEN_PY=1 recaptures the expected column from python3
// (an input added: append [input, null] and run it).
func TestSplitMatchesPython(t *testing.T) {
	const file = "testdata/python-split.json"
	if os.Getenv("SHLEX_GOLDEN_PY") == "1" {
		if out, err := exec.Command("python3", "-c", pythonSplit, file).CombinedOutput(); err != nil {
			t.Fatalf("python3: %v\n%s", err, out)
		}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var rows [][2]any
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		in := row[0].(string)
		got, err := Split(in)
		var gotAny any
		if err != nil {
			gotAny = "error: " + err.Error()
		} else {
			words := make([]any, len(got))
			for i, w := range got {
				words[i] = w
			}
			gotAny = words
		}
		g, _ := json.Marshal(gotAny)
		w, _ := json.Marshal(row[1])
		if string(g) != string(w) {
			t.Errorf("Split(%q) = %s, Python: %s", in, g, w)
		}
	}
}
