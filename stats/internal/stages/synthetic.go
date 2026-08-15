package stages

// SyntheticChangeUpdatedBy is the store's sentinel value for a change row
// that a stage mark's begin handler bootstraps only to give an
// otherwise-unknown mark somewhere to attach -- never a value a real
// `state set` writes, so a record carrying it is immediately recognisable
// as "nobody ever ran /myflow-start for this" (design.md, kan-174).
//
// It lives in package stages, not because both sides happen to import
// stages already, but because it exists only in service of a stage mark:
// the daemon's bootstrap path (internal/api/stages.go) writes it exactly
// when a mark's begin handler has no real change row to attach to, and the
// CLI (cmd/myflow/state.go) reads it back to decide whether `state get`
// should report `"synthetic": true`. That is a producer/consumer pair
// agreeing on one sentinel, which is what this package's constants section
// (see the package doc in names.go) is for -- keeping it here means the
// daemon's writer and the CLI's reader can never drift onto two copies of
// the same literal.
const SyntheticChangeUpdatedBy = "myflow stage begin (synthetic)"
