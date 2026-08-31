// Command flow's workspaceid.go implements `flow workspace-id <name>`, the
// CLI surface over deriveWorkspaceID -- the four-step derivation stated
// canonically under "The workspace id" in
// skills/flow-contracts/workspace-isolation.md, ported here so run 2 of the
// finish contract stops re-deriving it by hand from that contract for one
// 6-line value. It reads no file and touches no store: the id is a pure
// function of the change name and nothing else.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const workspaceIDUsage = `usage: flow workspace-id <name>

workspace-id prints the workspace id derived from a change name -- the same
deterministic id skills/flow-contracts/workspace-isolation.md's "The
workspace id" defines and scripts/prepare-workspace.sh already derives in
Bash for apply worktrees. The id is never recorded and never negotiated: the
same name always yields the same id.
`

// deriveWorkspaceID computes the workspace id for a change name, following
// workspace-isolation.md's four steps exactly, over name's UTF-8 bytes:
//
//  1. digest: the first 4 hex characters of lowercase SHA-256(name).
//  2. normalised name: each byte in A-Z is lowered by 32, each byte in
//     a-z/0-9 is kept, every other byte becomes '-'.
//  3. prefix: the longest leading run of whole '-'-separated segments of
//     the normalised name whose joined length is at most 12 characters,
//     or the first segment's first 12 characters if that segment alone
//     exceeds 12; trailing '-' are then trimmed.
//  4. id: prefix and digest joined by a single '-'.
func deriveWorkspaceID(name string) string {
	sum := sha256.Sum256([]byte(name))
	digest := hex.EncodeToString(sum[:])[:4]

	raw := []byte(name)
	normalised := make([]byte, len(raw))
	for i, b := range raw {
		switch {
		case b >= 'A' && b <= 'Z':
			normalised[i] = b + 32
		case b >= 'a' && b <= 'z', b >= '0' && b <= '9':
			normalised[i] = b
		default:
			normalised[i] = '-'
		}
	}

	prefix := prefixOf(string(normalised))
	return prefix + "-" + digest
}

// prefixOf implements step 3 over an already-normalised name: the longest
// leading run of whole '-'-separated segments totalling at most 12
// characters, or the first segment's own first 12 characters when that
// segment alone exceeds 12 -- then trailing '-' trimmed.
func prefixOf(normalised string) string {
	segments := strings.Split(normalised, "-")

	if len(segments[0]) > 12 {
		return segments[0][:12]
	}

	kept := 0
	length := -1 // -1 so the first segment's own length, plus one joining '-', lands on len(segments[0])
	for _, seg := range segments {
		next := length + 1 + len(seg)
		if next > 12 {
			break
		}
		length = next
		kept++
	}

	prefix := strings.Join(segments[:kept], "-")
	return strings.TrimRight(prefix, "-")
}

// runWorkspaceID takes no flags of its own -- its single positional argument
// is the change name, which workspace-isolation.md explicitly allows to
// begin with '-' (a name that normalises to nothing but separators is
// legal). Parsing it with flag.FlagSet would reject such a name as an
// unrecognized flag before it ever reaches deriveWorkspaceID, so args is
// read directly instead.
func runWorkspaceID(_ context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(stdout, workspaceIDUsage)
		return 0
	}
	if len(args) != 1 {
		fmt.Fprintln(stderr, "flow: expected exactly one argument, the change name")
		fmt.Fprint(stderr, workspaceIDUsage)
		return 2
	}

	fmt.Fprintln(stdout, deriveWorkspaceID(args[0]))
	return 0
}
