package main

import "testing"

// TestLaunchRoleIDFixKeysAreAnchored pins the fix-key regex anchors: only a
// fix's own fix-<k>-review / fix-<k>-implementer key maps to a role, never a
// key that merely contains one (pipeline-fix-2-review).
func TestLaunchRoleIDFixKeysAreAnchored(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ role, key, want string }{
		{"reviewer", "fix-2-review", "panel-reviewer"},
		{"reviewer", "pipeline-fix-2-review", ""},
		{"reviewer", "fix-2-reviewer", ""},
		{"implementer", "fix-2-implementer", "implementer"},
		{"implementer", "pipeline-fix-2-implementer", ""},
		{"implementer", "fix-2-implementer-x", ""},
	} {
		if got := launchRoleID(c.role, c.key); got != c.want {
			t.Errorf("launchRoleID(%q, %q) = %q, want %q", c.role, c.key, got, c.want)
		}
	}
}
