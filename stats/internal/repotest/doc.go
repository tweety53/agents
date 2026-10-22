// Package repotest holds the Go ports of the repository-level bash harnesses
// that once lived as scripts/test-*.sh: the Python hooks under hooks/ and the
// repository's own wiring. The tests exec the real hooks and tools (python3,
// git, make, node) against fixtures in t.TempDir(), never importing them, and
// never read or write the operator's real $HOME.
package repotest
