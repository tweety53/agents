package records

import (
	"regexp"
	"strings"
)

// A dispatch row's stored role is a role of flowd's roles table (the allowed
// pairs, skills/flow-contracts/model-policy.md): panel-reviewer,
// visual-verifier. A dispatcher may still type one of the
// coarse legacy names, and StoredRole turns it into the roles-table role its
// key names. Class folds a roles-table role back to the coarse class the
// consumers that only need "a reviewer" or "an implementer" test.

// legacyRoles are the coarse names `-role` accepted before the roles table
// named every role.
var legacyRoles = map[string]bool{
	"implementer": true, "reviewer": true, "panel-fix": true, "red-partner": true,
	"planner": true, "conductor": true, "verifier": true, "explorer": true, "merger": true,
}

// roleClass maps a roles-table role to its coarse class; a role absent from
// it is its own class. dispatch_class (migrations/0094) is its SQL twin, and
// a test holds the two equal.
var roleClass = map[string]string{
	"panel-reviewer": "reviewer", "panel-rerun": "reviewer", "gated-reviewer": "reviewer",
	"gated-re-reviewer": "reviewer", "self-review-pass": "reviewer",
	"loop-review-1": "reviewer", "loop-review-r": "reviewer",
	"minor-fix":      "panel-fix",
	"operator-asked": "implementer", "loop-fix": "implementer",
	"no-progress-fixer": "implementer",
	"plan-designer":     "planner", "designer": "planner", "loop-plan": "planner",
	// Retired roles (panel-merger, planner-short, and the self-review-* and
	// pipeline-fix-* loop roles 0109 merged into loop-*): kept so their
	// history rows fold as dispatch_class does.
	"panel-merger": "merger", "planner-short": "planner",
	"self-review-review-1": "reviewer", "self-review-review-r": "reviewer",
	"pipeline-fix-review-1": "reviewer", "pipeline-fix-review-r": "reviewer",
	"self-review-fix": "implementer", "pipeline-fix-fix": "implementer",
	"self-review-plan": "planner", "pipeline-fix-plan": "planner",
	// runner, retired for run-verify-list's own re-run (0110).
	"runner": "verifier", "visual-verifier": "verifier",
}

// Class is role's coarse class: reviewer, implementer, panel-fix, merger,
// planner, verifier, or the role itself where it has none.
func Class(role string) string {
	if c, ok := roleClass[role]; ok {
		return c
	}
	return role
}

// RoleClasses returns every roles-table role Class folds, for the parity test
// against dispatch_class.
func RoleClasses() map[string]string {
	out := make(map[string]string, len(roleClass))
	for k, v := range roleClass {
		out[k] = v
	}
	return out
}

// IsLegacyRole is whether role is one of the coarse names rather than a
// roles-table role.
func IsLegacyRole(role string) bool { return legacyRoles[role] }

var (
	// launchTaskKey is a gated task row's key: task-<ids>-implementer… or
	// task-<ids>-reviewer…, ids +-joined.
	launchTaskKey       = regexp.MustCompile(`^task-([0-9]+(?:\+[0-9]+)*)-(implementer|reviewer)(.*)$`)
	fixReviewKeyRE      = regexp.MustCompile(`^fix-[1-9][0-9]*-review$`)
	fixImplementerKeyRE = regexp.MustCompile(`^fix-[1-9][0-9]*-implementer$`)
	panelReviewKeyRE    = regexp.MustCompile(`^panel-[0-9]+-`)
)

// LaunchRoleID is the roles-table role a dispatch's role and key derive, or ""
// where they derive none -- a self-review or pipeline-fix key among them. A
// task's gated-review fixer and its re-commit run on the panel-fix pair. A
// role that is already a roles-table role is its own.
func LaunchRoleID(role, key string) string {
	if role != "" && !IsLegacyRole(role) {
		return role
	}
	if m := launchTaskKey.FindStringSubmatch(key); m != nil {
		switch {
		case m[2] == "implementer" && (strings.HasPrefix(m[3], "-fix-") || strings.HasPrefix(m[3], "-recommit-")):
			return "panel-fix"
		case m[2] == "implementer":
			return "implementer"
		case strings.HasPrefix(m[3], "-fix-"):
			return "gated-re-reviewer"
		default:
			return "gated-reviewer"
		}
	}
	switch {
	case role == "panel-fix" && strings.HasPrefix(key, "panel-") && strings.HasSuffix(key, "-minor-fix"):
		return "minor-fix"
	case role == "panel-fix":
		return "panel-fix"
	case role == "reviewer" && fixReviewKeyRE.MatchString(key):
		return "panel-reviewer"
	case role == "implementer" && fixImplementerKeyRE.MatchString(key):
		return "implementer"
	case role == "planner" && strings.HasPrefix(key, "design-"):
		return "plan-designer"
	case role == "planner" && strings.HasPrefix(key, "designer-"):
		return "designer"
	case role == "verifier" && strings.HasPrefix(key, "visual-verify"):
		return "visual-verifier"
	case role == "implementer" && strings.HasPrefix(key, "operator-"):
		return "operator-asked"
	}
	return ""
}

// StoredRole is the role a dispatch row is written under: the roles-table role
// its prompt's `Role:` line named (policyRole), else the one its role and key
// derive, else role as typed.
func StoredRole(role, key, policyRole string) string {
	if policyRole != "" {
		return policyRole
	}
	if id := LaunchRoleID(role, key); id != "" {
		return id
	}
	// A panel round's reviewer is panel-reviewer unless its prompt said it was a re-run (panel-rerun).
	if role == "reviewer" && panelReviewKeyRE.MatchString(key) {
		return "panel-reviewer"
	}
	return role
}
