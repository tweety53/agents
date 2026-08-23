package harvest

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestUnattributedReasonStringsMatchTheLedgersCopy pins internal/records'
// duplicated copy of the three "unattributed" reason strings this package
// declares -- reasonSessionNeverBound, reasonSessionAmbiguous and
// reasonDispatchAmbiguous -- against the actual constants, rather than
// against a second hand-typed copy of them.
//
// WHY THE DUPLICATION EXISTS, AND WHY THIS TEST LIVES HERE RATHER THAN
// REMOVING IT. internal/records/render.go restates these three literals
// because internal/records must not import internal/harvest -- it is the
// wire shape every layer depends on, and importing the harvester into it
// to reuse four field names would invert that dependency. The tradeoff is
// named in render.go's own comment: a renamed producer reason falls to
// tokenLine's verbatim `default:` branch rather than being mislabelled as
// a different state, so drift is visible rather than silent, but nothing
// signals that the drift happened. This test is that signal.
//
// WHY THIS PACKAGE, NOT internal/records. The three constants this test
// needs are unexported, and reaching them from outside package harvest
// would mean exporting them from production code purely so a test could
// read them -- widening the production API to serve a test is the
// cheaper alternative this task asks to avoid. package harvest already
// has them in scope. The dependency this test file adds runs the other
// way: internal/harvest's TEST BINARY importing internal/records, which
// is not the direction render.go's comment forbids (only
// internal/records -> internal/harvest, in production code, is
// forbidden), and it is a test-only edge regardless -- confirmed below by
// `go list -deps` naming this package's own PRODUCTION build, which does
// not import internal/records at all.
//
//	cd stats && go list -deps ./internal/harvest | grep internal/records
//	(no output: the production package internal/harvest does not depend
//	on internal/records; only this _test.go file does, and test files are
//	excluded from `go list -deps`'s default package)
//
// An external test package (`package records_test`, living in
// internal/records) was the other candidate the task named, but it could
// not do this job: it can reach harvest's exported API, but not these
// three unexported constants, so it would still be pinning a second
// hand-typed copy of the literals rather than the constants themselves.
func TestUnattributedReasonStringsMatchTheLedgersCopy(t *testing.T) {
	cases := []struct {
		name   string
		reason string
		bag    string
		want   string
	}{
		{
			name:   "reasonSessionNeverBound",
			reason: reasonSessionNeverBound,
			bag:    `{"unattributed":{"reason":` + jsonString(reasonSessionNeverBound) + `}}`,
			want:   "cost unattributed — session never bound",
		},
		{
			name:   "reasonSessionAmbiguous",
			reason: reasonSessionAmbiguous,
			bag:    `{"unattributed":{"reason":` + jsonString(reasonSessionAmbiguous) + `,"candidates":2}}`,
			want:   "cost unattributed — the session token matched 2 sessions",
		},
		{
			name:   "reasonDispatchAmbiguous",
			reason: reasonDispatchAmbiguous,
			bag:    `{"unattributed":{"reason":` + jsonString(reasonDispatchAmbiguous) + `,"candidates":3}}`,
			want:   "cost unattributed — indistinguishable from 3 concurrent dispatches",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := records.RenderLedger(records.Run{
				Change: "demo",
				Dispatches: []records.Dispatch{{
					Seq: 1, Role: "implementer", Model: "opus",
					StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
					Metrics:   json.RawMessage(tc.bag),
				}},
			})

			wantLine := "- Tokens: " + tc.want + "\n"
			if !strings.Contains(out, wantLine) {
				t.Errorf(
					"internal/harvest's %s constant (%q) no longer matches internal/records' copy of it -- "+
						"the ledger fell through to tokenLine's verbatim default branch instead of rendering %q:\n%s",
					tc.name, tc.reason, tc.want, out,
				)
			}
		})
	}
}

// jsonString renders s as a JSON string literal, so the test cases above
// embed the package's actual reason constants rather than a second
// hand-typed copy of their text.
func jsonString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}
