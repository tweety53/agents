package guard

import (
	"strconv"
	"strings"
	"testing"
)

// TestCoverage fails on any difference in coverage.go's rendered fragment,
// verdict lines, verdict code, or the message a rejected call gives, from
// what the bash coverage library printed for the same calls at 2a3a0297 --
// each want captured from that script before its delete
// (coverage-golden-strings), covering this table's cases and those of the
// library's own bash harness. A call's line is its message, if any, then
// `call rc=<0|2>`; the report fragment and verdict lines follow, then
// `verdict rc=<0|1>`.
func TestCoverage(t *testing.T) {
	t.Parallel()
	type call struct {
		declare     bool
		member, arg string // arg: the reason (declare) or the count (record)
	}
	rec := func(m string, n int) call { return call{member: m, arg: strconv.Itoa(n)} }
	dec := func(m, reason string) call { return call{declare: true, member: m, arg: reason} }
	for _, tc := range []struct {
		name  string
		calls []call
		want  string
	}{
		{"none recorded", nil,
			"coverage: no member was ever recorded — this guard checked nothing (empty corpus)\nverdict rc=1\n"},
		{"none recorded, one declared", []call{dec("a.md", "why")},
			"call rc=0\ncoverage: no member was ever recorded — this guard checked nothing (empty corpus)\nverdict rc=1\n"},
		{"all non-zero", []call{rec("a.md", 3), rec("b.md", 1)},
			"call rc=0\ncall rc=0\na.md 3 · b.md 1\nverdict rc=0\n"},
		{"an undeclared zero", []call{rec("a.md", 2), rec("b.md", 0)},
			"call rc=0\ncall rc=0\na.md 2 · b.md 0\nb.md: 0 checked, and not declared expected-zero (coverage)\nverdict rc=1\n"},
		{"a declared zero with a reason", []call{dec("b.md", "stub — nothing to check"), rec("a.md", 1), rec("b.md", 0)},
			"call rc=0\ncall rc=0\ncall rc=0\na.md 1 · b.md 0 (declared: stub — nothing to check)\nverdict rc=0\n"},
		{"a declared zero without a reason", []call{dec("b.md", ""), rec("b.md", 0)},
			"call rc=0\ncall rc=0\nb.md 0 (declared)\nverdict rc=0\n"},
		{"a declared zero with an empty reason", []call{dec("b.md", ""), rec("b.md", 0)},
			"call rc=0\ncall rc=0\nb.md 0 (declared)\nverdict rc=0\n"},
		{"a declared member now non-zero, with and without a reason",
			[]call{dec("a.md", "why"), dec("b.md", ""), rec("a.md", 4), rec("b.md", 2)},
			"call rc=0\ncall rc=0\ncall rc=0\ncall rc=0\na.md 4 · b.md 2\na.md: 4 checked, but declared expected-zero (why) — declaration is now false (coverage)\nb.md: 2 checked, but declared expected-zero — declaration is now false (coverage)\nverdict rc=1\n"},
		{"a declared member never recorded", []call{dec("gone.md", "stale"), rec("a.md", 1)},
			"call rc=0\ncall rc=0\na.md 1\ngone.md: declared expected-zero but never recorded — stale declaration, or this member was never scanned (coverage)\nverdict rc=1\n"},
		{"several members in and out of declaration order",
			[]call{dec("d.md", "r1"), dec("b.md", "r2"), rec("c.md", 0), rec("b.md", 0), rec("a.md", 5), rec("d.md", 0), dec("e.md", "r3"), rec("f.md", 7)},
			"call rc=0\ncall rc=0\ncall rc=0\ncall rc=0\ncall rc=0\ncall rc=0\ncall rc=0\ncall rc=0\nc.md 0 · b.md 0 (declared: r2) · a.md 5 · d.md 0 (declared: r1) · f.md 7\nc.md: 0 checked, and not declared expected-zero (coverage)\ne.md: declared expected-zero but never recorded — stale declaration, or this member was never scanned (coverage)\nverdict rc=1\n"},
		{"a member name carrying a colon", []call{rec("x:y.md", 0)},
			"call rc=0\nx:y.md 0\nx:y.md: 0 checked, and not declared expected-zero (coverage)\nverdict rc=1\n"},
		{"a second record of one member is rejected", []call{rec("a.md", 1), rec("a.md", 2)},
			"call rc=0\ncoverage_record: 'a.md' was already recorded — call coverage_record once per member\ncall rc=2\na.md 1\nverdict rc=0\n"},
		{"a second declaration of one member is rejected", []call{dec("a.md", "x"), dec("a.md", "y"), rec("a.md", 0)},
			"call rc=0\ncoverage_declare: 'a.md' was already declared — call coverage_declare once per member\ncall rc=2\ncall rc=0\na.md 0 (declared: x)\nverdict rc=0\n"},
		{"an empty member is rejected", []call{rec("", 1), dec("", "r"), rec("a.md", 1)},
			"coverage_record: member name must not be empty\ncall rc=2\ncoverage_declare: member name must not be empty\ncall rc=2\ncall rc=0\na.md 1\nverdict rc=0\n"},
		{"a negative count is rejected", []call{rec("a.md", -1), rec("b.md", 1)},
			"coverage_record: count must be a non-negative integer, got '-1'\ncall rc=2\ncall rc=0\nb.md 1\nverdict rc=0\n"},
		{"a leading-dash, spaced member renders intact", []call{rec("-weird flag-like name", 3)},
			"call rc=0\n-weird flag-like name 3\nverdict rc=0\n"},
		{"a leading-dash, spaced undeclared zero is named", []call{rec("-weird flag-like name", 0)},
			"call rc=0\n-weird flag-like name 0\n-weird flag-like name: 0 checked, and not declared expected-zero (coverage)\nverdict rc=1\n"},
		{"a leading-dash, spaced declared zero is silenced", []call{rec("-weird flag-like name", 0), dec("-weird flag-like name", "test fixture")},
			"call rc=0\ncall rc=0\n-weird flag-like name 0 (declared: test fixture)\nverdict rc=0\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var c coverage
			var got strings.Builder
			for _, k := range tc.calls {
				var err error
				if k.declare {
					err = c.declare(k.member, k.arg)
				} else {
					n, _ := strconv.Atoi(k.arg)
					err = c.record(k.member, n)
				}
				if err != nil {
					got.WriteString(err.Error() + "\ncall rc=2\n")
				} else {
					got.WriteString("call rc=0\n")
				}
			}
			if frag := c.report(); frag != "" {
				got.WriteString(frag + "\n")
			}
			lines := c.verdict()
			for _, l := range lines {
				got.WriteString(l + "\n")
			}
			rc := 0
			if len(lines) > 0 {
				rc = 1
			}
			got.WriteString("verdict rc=" + strconv.Itoa(rc) + "\n")
			if got.String() != tc.want {
				t.Errorf("coverage.go output\n got: %q\nwant: %q", got.String(), tc.want)
			}
		})
	}
}
