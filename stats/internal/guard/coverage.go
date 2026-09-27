package guard

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// coverage is scripts/lib/coverage.sh's per-member coverage for the Go
// guards: recording a per-member count (record), rendering the
// members-and-counts fragment a verdict line carries (report), and deciding
// declared vs. undeclared zero (verdict). coverage.sh stays the source of
// truth while bash guards still source it; its header carries the reasoning
// (KAN-73, KAN-197 F3, KAN-374 F9) and TestCoverageParity fails when the two
// render, decide or reject differently.
//
// Declaration is a call the guard's own source makes (declare), never
// inferred from the tree. A rejected call — an empty member, a negative
// count, a second record or declaration of one member — returns coverage.sh's
// own message as an error rather than being coerced or overwritten.
type coverage struct {
	members  []string
	counts   []int
	declared []string
	reasons  []string
}

// record is coverage_record: how many items the guard checked for member.
func (c *coverage) record(member string, count int) error {
	switch {
	case member == "":
		return fmt.Errorf("coverage_record: member name must not be empty")
	case count < 0:
		return fmt.Errorf("coverage_record: count must be a non-negative integer, got '%d'", count)
	case slices.Contains(c.members, member):
		return fmt.Errorf("coverage_record: '%s' was already recorded — call coverage_record once per member", member)
	}
	c.members = append(c.members, member)
	c.counts = append(c.counts, count)
	return nil
}

// declare is coverage_declare: member legitimately checks nothing, for
// reason (may be empty), carried into report's rendering of it.
func (c *coverage) declare(member, reason string) error {
	switch {
	case member == "":
		return fmt.Errorf("coverage_declare: member name must not be empty")
	case slices.Contains(c.declared, member):
		return fmt.Errorf("coverage_declare: '%s' was already declared — call coverage_declare once per member", member)
	}
	c.declared = append(c.declared, member)
	c.reasons = append(c.reasons, reason)
	return nil
}

// reason is coverage_declared_reason: member's declared reason, and whether
// it was declared at all.
func (c *coverage) reason(member string) (string, bool) {
	if i := slices.Index(c.declared, member); i >= 0 {
		return c.reasons[i], true
	}
	return "", false
}

// report is coverage_report without its newline: every recorded member and
// its count in record order, joined by " · ", a declared zero suffixed
// " (declared)" or " (declared: <reason>)". Empty for an empty corpus —
// verdict, not report, fails that.
func (c *coverage) report() string {
	parts := make([]string, len(c.members))
	for i, m := range c.members {
		parts[i] = m + " " + strconv.Itoa(c.counts[i])
		if reason, ok := c.reason(m); ok && c.counts[i] == 0 {
			if reason != "" {
				parts[i] += " (declared: " + reason + ")"
			} else {
				parts[i] += " (declared)"
			}
		}
	}
	return strings.Join(parts, " · ")
}

// verdict is coverage_verdict: one line per violation, none when the
// coverage holds — an empty corpus, an undeclared zero, a declared member
// now non-zero, a declared member never recorded.
func (c *coverage) verdict() []string {
	if len(c.members) == 0 {
		return []string{"coverage: no member was ever recorded — this guard checked nothing (empty corpus)"}
	}
	var lines []string
	for i, m := range c.members {
		reason, declared := c.reason(m)
		switch n := c.counts[i]; {
		case n == 0 && !declared:
			lines = append(lines, m+": 0 checked, and not declared expected-zero (coverage)")
		case n != 0 && declared && reason != "":
			lines = append(lines, fmt.Sprintf("%s: %d checked, but declared expected-zero (%s) — declaration is now false (coverage)", m, n, reason))
		case n != 0 && declared:
			lines = append(lines, fmt.Sprintf("%s: %d checked, but declared expected-zero — declaration is now false (coverage)", m, n))
		}
	}
	for _, m := range c.declared {
		if !slices.Contains(c.members, m) {
			lines = append(lines, m+": declared expected-zero but never recorded — stale declaration, or this member was never scanned (coverage)")
		}
	}
	return lines
}
