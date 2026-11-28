package main

import (
	"bytes"
	"strings"
	"testing"
)

// The binary answers `hook <name>` alone: any other first argument, or none, is a usage error
// (exit 2), while `hook <name>` reaches guard/hook's own dispatch, which fails open on a
// garbage payload and exits 2 on an unknown hook name.
func TestRunAnswersHookAlone(t *testing.T) {
	t.Setenv("FLOW_STATE_DIR", t.TempDir())
	for _, tc := range []struct {
		args  []string
		stdin string
		code  int
		err   string
	}{
		{nil, "", 2, "usage: flow-guard-hook hook <name>"},
		{[]string{"check-guard-calls"}, "", 2, "usage: flow-guard-hook hook <name>"},
		{[]string{"hook", "no-such-hook"}, "", 2, "usage: flow-guard hook <name>"},
		{[]string{"hook", "block-unsafe-commands"}, "not json {", 0, ""},
	} {
		var stdout, stderr bytes.Buffer
		code := run(tc.args, strings.NewReader(tc.stdin), &stdout, &stderr)
		if code != tc.code || !strings.Contains(stderr.String(), tc.err) || (tc.code == 0 && stdout.Len() != 0) {
			t.Errorf("run(%q) = %d, stdout %q, stderr %q; want %d, stderr containing %q", tc.args, code, stdout.String(), stderr.String(), tc.code, tc.err)
		}
	}
}
