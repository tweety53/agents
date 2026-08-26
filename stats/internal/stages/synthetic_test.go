package stages

import "testing"

// TestSyntheticChangeUpdatedByIsPinnedToTheStoredLiteral pins
// SyntheticChangeUpdatedBy to its exact byte sequence, deliberately spelling
// the literal out rather than comparing the constant to itself.
//
// WHY A TEST THAT LOOKS TAUTOLOGICAL IS THE ONLY ONE THAT WORKS HERE. This
// constant is not vocabulary -- it is a value the live `changes` table already
// holds in its `updated_by` column, written by every synthetic change row the
// store has ever created. Renaming it does not migrate those rows; it makes
// the code stop recognising them, and the row is then invisible to synthetic
// detection forever.
//
// Every other test that touches this constant refers to it SYMBOLICALLY --
// `stages.SyntheticChangeUpdatedBy` in cmd/flow/state_test.go and
// internal/api/stages_test.go -- so producer and consumer read the same symbol
// and agree with each other no matter what it says. That is exactly the shape
// that let KAN-289's original four defects through: a fixture and the code it
// covers renamed together, green, and both wrong about the world.
//
// Measured, not assumed: during the myflow->flow rename this constant WAS
// renamed to "flow stage begin (synthetic)", and the entire suite -- go vet,
// gofmt and `go test ./... -race` across all sixteen packages -- passed with
// the mutant in place. The rename was caught only by querying the live
// database. This test is what makes the next attempt fail loudly instead.
//
// If a future change genuinely intends to move this value, it must migrate
// `changes.updated_by` in the same change and update the literal here. Editing
// this literal alone to make a build pass reintroduces the defect it exists to
// prevent.
func TestSyntheticChangeUpdatedByIsPinnedToTheStoredLiteral(t *testing.T) {
	const storedLiteral = "myflow stage begin (synthetic)"

	if SyntheticChangeUpdatedBy != storedLiteral {
		t.Fatalf(
			"SyntheticChangeUpdatedBy = %q, want %q.\n"+
				"This value is stored in changes.updated_by in the live database. "+
				"Changing the constant without migrating those rows makes every "+
				"existing synthetic change row undetectable. See this test's doc "+
				"comment before editing the literal.",
			SyntheticChangeUpdatedBy, storedLiteral,
		)
	}
}
