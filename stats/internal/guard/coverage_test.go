package guard

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// TestCoverageParity runs scripts/lib/coverage.sh and coverage.go over the
// same declarations and records and fails on any difference in the rendered
// fragment, the verdict lines, the verdict's return code, or the message a
// rejected call gives (coverage-go-twin, per helper-parity-tests).
func TestCoverageParity(t *testing.T) {
	t.Parallel()
	type call struct {
		declare       bool
		member, arg   string // arg: the reason (declare) or the count (record)
		noReasonGiven bool   // declare with one argument rather than an empty reason
	}
	rec := func(m string, n int) call { return call{member: m, arg: strconv.Itoa(n)} }
	dec := func(m, reason string) call { return call{declare: true, member: m, arg: reason} }
	for _, tc := range []struct {
		name  string
		calls []call
	}{
		{"none recorded", nil},
		{"none recorded, one declared", []call{dec("a.md", "why")}},
		{"all non-zero", []call{rec("a.md", 3), rec("b.md", 1)}},
		{"an undeclared zero", []call{rec("a.md", 2), rec("b.md", 0)}},
		{"a declared zero with a reason", []call{dec("b.md", "stub — nothing to check"), rec("a.md", 1), rec("b.md", 0)}},
		{"a declared zero without a reason", []call{{declare: true, member: "b.md", noReasonGiven: true}, rec("b.md", 0)}},
		{"a declared zero with an empty reason", []call{dec("b.md", ""), rec("b.md", 0)}},
		{"a declared member now non-zero, with and without a reason",
			[]call{dec("a.md", "why"), dec("b.md", ""), rec("a.md", 4), rec("b.md", 2)}},
		{"a declared member never recorded", []call{dec("gone.md", "stale"), rec("a.md", 1)}},
		{"several members in and out of declaration order",
			[]call{dec("d.md", "r1"), dec("b.md", "r2"), rec("c.md", 0), rec("b.md", 0), rec("a.md", 5), rec("d.md", 0), dec("e.md", "r3"), rec("f.md", 7)}},
		{"a member name carrying a colon", []call{rec("x:y.md", 0)}},
		{"a second record of one member is rejected", []call{rec("a.md", 1), rec("a.md", 2)}},
		{"a second declaration of one member is rejected", []call{dec("a.md", "x"), dec("a.md", "y"), rec("a.md", 0)}},
		{"an empty member is rejected", []call{rec("", 1), dec("", "r"), rec("a.md", 1)}},
		{"a negative count is rejected", []call{rec("a.md", -1), rec("b.md", 1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			script := `. "$1" || exit 99` + "\n"
			var c coverage
			var goOut strings.Builder
			for _, k := range tc.calls {
				var err error
				if k.declare {
					if k.noReasonGiven {
						script += "coverage_declare " + shq(k.member) + "\n"
					} else {
						script += "coverage_declare " + shq(k.member) + " " + shq(k.arg) + "\n"
					}
					err = c.declare(k.member, k.arg)
				} else {
					n, _ := strconv.Atoi(k.arg)
					script += "coverage_record " + shq(k.member) + " " + shq(k.arg) + "\n"
					err = c.record(k.member, n)
				}
				script += `echo "call rc=$?"` + "\n"
				if err != nil {
					goOut.WriteString(err.Error() + "\ncall rc=2\n")
				} else {
					goOut.WriteString("call rc=0\n")
				}
			}
			script += "coverage_report\ncoverage_verdict\necho \"verdict rc=$?\"\n"
			if frag := c.report(); frag != "" {
				goOut.WriteString(frag + "\n")
			}
			lines := c.verdict()
			for _, l := range lines {
				goOut.WriteString(l + "\n")
			}
			rc := 0
			if len(lines) > 0 {
				rc = 1
			}
			goOut.WriteString("verdict rc=" + strconv.Itoa(rc) + "\n")

			out, err := exec.Command("bash", "-c", "exec 2>&1\n"+script, "_", "../../../scripts/lib/coverage.sh").Output()
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, out)
			}
			if string(out) != goOut.String() {
				t.Errorf("coverage.go differs from coverage.sh\n go:\n%s\nbash:\n%s", goOut.String(), out)
			}
		})
	}
}

// shq single-quotes s for a bash command line.
func shq(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
