// changename.go holds the change-name helper several guards share.

package guard

import "strings"

// plainChangeName is the allowlist: records.Destination's Protection 1
// (stats/internal/records/render.go), character for character, and that
// function's comment is canonical for why each hazard is in it — the `/` that
// was blocked only by an accident of string concatenation, and the glob
// metacharacter that once matched and overwrote a DIFFERENT change's
// preserved record. The bash guard that still carries the rule
// (check-workspace-isolation.sh) keeps its own copy because it is
// single-file by design and copied into projects one at a time; its harness
// and TestCheckCleanupComplete/12 assert the same rejected shapes, which is what keeps the copies from drifting apart
// silently.
//
// It also closes the symlink question at these paths: the file and directory
// tests follow symlinks, so a name that cannot leave the repository is what
// makes the paths this guard tests the ones it was asked about.
//
// THE ALLOWED CHARACTERS ARE ENUMERATED RATHER THAN WRITTEN AS RANGES because
// a shell bracket range is a COLLATING range, not a byte range, so what it
// admits is whatever the ambient locale's collation puts between the two
// endpoints. MEASURED on bash 3.2 (Darwin 25.5.0) against the bash guard's
// previous spelling `[!A-Za-z0-9]* | *[!A-Za-z0-9._-]*`: the names `écho`,
// `İstanbul`, `ﬀoo`, `ⅰx`, `Ａbc` and `ⅹ` were all refused under `LC_ALL=C`
// and `ru_RU.UTF-8`, and every one of them was ADMITTED under `en_US.UTF-8`,
// `de_DE.UTF-8` and `tr_TR.UTF-8`. A containment gate whose accepted set
// changes with the operator's environment is not a containment gate.
//
// A LITERAL LIST HAS NO ENDPOINTS, so there is nothing for a collation order
// to reorder: membership is membership in every locale. That is why the bash
// fix was the enumeration and NOT `export LC_ALL=C`, which
// check-unfinished-work.sh does carry for reasons of its own — in this guard
// it would have been EXPORTED into the project's own `survivors` command,
// changing the locale a project's tooling runs under. The enumeration needs
// no environment at all. The accepted set is unchanged from what the range
// form accepted under `LC_ALL=C`. This Go port compares bytes against
// ccAllowFirst/ccAllowRest, so no locale reaches it either way; the
// enumeration is kept so the rule reads the same as its bash copies.
//
// THIS COMMENT IS CANONICAL for that measurement and that reasoning, and the
// other copies of the rule cite it. It was Protection 1 of the record-copying
// script this repository has since retired, along with the session records it
// copied; of the copies that survive, the bash check-cleanup-complete.sh was
// the one where the enumeration was the fix rather than belt and braces, so
// the reasoning moved there rather than going with the script, and moved here
// with the port (KAN-760).
//
// It is also change-plan.sh's _change_plan_name_ok, which the change-plan
// resolver (changeplan.go), gather-dispatch-context and
// check-panel-reproducer-exit-contract call: one allowlist in this package,
// so an empty name is refused here rather than indexed.
func plainChangeName(name string) bool {
	if name == "" || strings.IndexByte(ccAllowFirst, name[0]) < 0 {
		return false
	}
	for i := 0; i < len(name); i++ {
		if strings.IndexByte(ccAllowRest, name[i]) < 0 {
			return false
		}
	}
	return true
}

const (
	ccAllowFirst = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	ccAllowRest  = ccAllowFirst + "._-"
)
