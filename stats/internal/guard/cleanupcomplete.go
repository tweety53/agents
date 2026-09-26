package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// check-cleanup-complete: scripts/check-cleanup-complete.sh's header is the
// contract; the reasoning for each branch, moved here from the bash body it
// replaced (0747740), sits beside the code it explains, updated where the port
// changed the mechanism. Two scans the bash ran as separate grep/awk
// processes over .flow/project.md -- the heading count and the survivors-row
// extraction -- read the one buffer here, so their "could not look" branches
// collapse into the single read's failure, reported as the bash reports a
// grep that could not look.
func init() { Registry["check-cleanup-complete"] = checkCleanupComplete }

const (
	// ccDefaultTimeout is the bound on the project's `survivors` command — the
	// same 60 seconds the `## stop` command gets under **Worktree cleanup**
	// (`skills/flow-contracts/pipeline.md`). One number for every
	// project-supplied command run unattended is one number an operator has
	// to know; a second one would be a second rule to keep from drifting, and
	// the case for tightening it here is weaker than it looks. A bound that is
	// too generous costs wall clock only in the rare hang. A bound that is too
	// tight turns a slow-but-working report into a skip, and a skip leaves the
	// row UNVERIFIED — the worst outcome this guard can reach short of a false
	// COMPLETE.
	//
	// ccDefaultGrace is the grace between the SIGTERM and the SIGKILL
	// ccRunSurvivors sends. Named beside the bound and justified to the same
	// depth, because a second unexplained number in a function whose every
	// other number is argued for is the one a later reader changes without
	// knowing what it was holding.
	//
	// IT IS NOT A SECOND BOUND, and that is why it is short. It is spent AFTER
	// the bound has already fired, so every second of it is a second past the
	// promise the bound made to a run nobody is watching — the effective
	// ceiling is the bound plus the grace, and a grace generous enough for a
	// JVM shutdown hook would make the second term a meaningful fraction of
	// the first.
	//
	// IT BUYS NOTHING FOR THE VERDICT, WHICH IS ALREADY DECIDED. By the time
	// the SIGTERM goes out, timedOut is true and the row is a skip: a
	// terminated command has said nothing about survivors, and nothing it does
	// during the grace can change that. What the grace is for is the COMMAND's
	// own housekeeping — closing a connection rather than leaving the service
	// to reap it — a courtesy whose cost when it is not finished in time is
	// real but bounded: a server-side connection left to its own idle timeout.
	// Two seconds is enough for a client that closes a socket on SIGTERM and
	// not enough for a `./gradlew` target running a shutdown hook, and that
	// asymmetry is accepted rather than papered over. The remedy for a command
	// that needs longer is a command that answers faster, which is the same
	// remedy the bound itself points a project author at under "What
	// `survivors` prints, and what its exit code means" in
	// skills/flow-contracts/project-configuration.md — and a command reaching
	// a service through a container runtime is not helped by ANY grace here,
	// for the reason ccRunSurvivors gives.
	//
	// THERE IS NO ENVIRONMENT OVERRIDE, and the absence is a decision rather
	// than an omission. CHECK_CLEANUP_SURVIVORS_TIMEOUT exists because a
	// harness would otherwise spend a minute of wall clock per timeout case;
	// the grace costs two seconds, so that argument buys nothing here — and a
	// knob on the interface every /flow integrate and archive run reads has to
	// be paid for by more than symmetry. (The Go tests shorten it in-process,
	// through Env.SurvivorsKillGrace, which no caller of the shim can reach.)
	ccDefaultTimeout = 60 * time.Second
	ccDefaultGrace   = 2 * time.Second
	ccAllowFirst     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	ccAllowRest      = ccAllowFirst + "._-"
)

// ccHeading is the heading rule, written once and used by both the presence
// count and the survivors-row extraction in ccWorkspaceRow. Two spellings of
// it would be two answers to "is this project isolated?", and the
// extraction's would win silently.
var (
	ccUnknownToken = regexp.MustCompile(`<[A-Za-z0-9_:.-]+>`)
	ccHeading      = regexp.MustCompile(`^##[ \t\n\v\f\r]+workspace isolation[ \t\n\v\f\r]*$`)
	ccAnyHeading   = regexp.MustCompile(`^#+[ \t\n\v\f\r]`)
)

func checkCleanupComplete(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	repo, name, stateDir := arg(0), arg(1), arg(2)
	if repo == "" || name == "" || stateDir == "" {
		fmt.Fprintln(stderr, "usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>")
		return 2
	}
	// CONTAINMENT: the change name arrives from a pull-request-editable state
	// file and is concatenated into a path, a ref name and a directory test
	// below. Without this check `../../nonexistent-decoy` makes every row answer
	// "already gone" and the guard reports COMPLETE while the real change
	// directory and artifact source are still sitting there — a confirmed
	// cleanup that removed nothing. plainChangeName carries the rule and its
	// reasoning. The refusal is sanitized for the same reason the verdict below
	// is: the name arrives from a pull-request-editable state file, and this is
	// the one message that quotes it back before the allowlist has vouched for
	// anything.
	if !plainChangeName(name) {
		fmt.Fprint(stderr, ccSanitize(fmt.Sprintf("check-cleanup-complete: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)))
		return 2
	}
	git := "git"
	if p, ok := lookPath(env, "git"); ok {
		git = p
	}
	gitCmd := func(a ...string) *exec.Cmd { return exec.Command(git, append([]string{"-C", repo}, a...)...) }
	if !isDir(repo) || gitCmd("rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "check-cleanup-complete: %s is not a git repository — cannot determine anything\n", repo)
		return 2
	}
	// A state directory that is not there cannot answer whether the artifact
	// source is gone. A mistyped path would otherwise read as "the artifact is
	// absent" and contribute to a COMPLETE verdict nobody checked.
	if !isDir(stateDir) {
		fmt.Fprintf(stderr, "check-cleanup-complete: %s is not a directory — cannot tell whether the proposal artifact source remains\n", stateDir)
		return 2
	}

	var left, notes []string
	add := func(s string) { left = append(left, s) }
	// notes carries what the verdict line must say ABOUT a row without that being
	// a leftover: the workspace row verified clean, or its verification skipped.
	// It is separate from left because the two are read differently — a leftover
	// stops run 2 and a note does not — and a skip written into the breakdown
	// would be acted on as something to go and remove.
	note := func(s string) { notes = append(notes, s) }

	// Row one — the worktree, found by BRANCH rather than by path, exactly as
	// the Worktree cleanup contract finds worktrees when the state file's map
	// is absent or empty: a path is never guessed from a conventional layout.
	//
	// Read into memory, never through a scratch file: a verifier that writes
	// into the tree it is verifying can leave behind exactly the class of
	// leftover it exists to report.
	//
	// git's failure is not read as an empty list. Doing so would turn an
	// unreadable repository into a COMPLETE verdict — the reassuring one —
	// which is the same silence this guard was added to break.
	porcelain, err := gitCmd("worktree", "list", "--porcelain").Output()
	if err != nil {
		fmt.Fprintf(stderr, "check-cleanup-complete: cannot list the worktrees of %s — cannot determine anything\n", repo)
		return 2
	}
	var wts []string
	wt := ""
	// The path is the whole rest of the `worktree ` line, never a
	// whitespace-split field: `worktree list --porcelain` emits the path raw, so
	// a worktree under a TMPDIR or a home directory containing a space is
	// truncated at the first one by a field reference, and the operator is then
	// sent to a path that does not exist. The branch is a ref name and cannot
	// contain a space, so its first field is right for it — and the comparison
	// is for EQUALITY, so a neighbouring change's spectre/<name>-something is
	// not reported as this change's leftover.
	//
	// A worktree still listed after `git worktree prune` should have run is a
	// leftover whether or not its directory survives: the registration is what
	// run 2 is required to remove.
	for _, l := range strings.Split(string(porcelain), "\n") {
		if p, ok := strings.CutPrefix(l, "worktree "); ok {
			wt = p
		} else if b, ok := strings.CutPrefix(l, "branch "); ok {
			if f := strings.Fields(b); len(f) > 0 && f[0] == "refs/heads/spectre/"+name {
				wts = append(wts, wt)
			}
		}
	}
	if len(wts) > 0 {
		add(fmt.Sprintf("worktree(s) still registered for spectre/%s at %s", name, strings.Join(wts, ", ")))
	}

	// Rows two and three — the local branch and the remote branch. The remote
	// one is read through its tracking ref, which is what this repository can
	// see without a network call; run 2 deletes the remote branch and prunes
	// the ref together, so a surviving ref is the observable half of that step
	// failing.
	//
	// refState <ref> — `present`, `absent` or `unreadable`, and never collapses
	// the last two. This is the invariant above applied to git's ref store,
	// and it takes two commands because ONE cannot say it:
	//
	//   `show-ref --verify --quiet` exits 1 both for a ref that is not there
	//   and for a ref whose loose file cannot be read or does not parse.
	//   Dropping --quiet does not help — both become the same
	//   `fatal: '<ref>' - not a valid ref` and the same exit 128. So the
	//   positive answer is taken from show-ref, which matches the FULL ref
	//   name and never a prefix, and only its exit 1 is sent on to be
	//   disambiguated.
	//
	//   `for-each-ref` is the command that distinguishes them, and it does so
	//   on stderr rather than in its exit status: a ref that is absent
	//   produces nothing at all, while one that exists and cannot be resolved
	//   produces `warning: ignoring broken ref <ref>` and exit 0. Measured on
	//   git 2.50.1 against a mode-000 loose ref and against a loose ref
	//   holding garbage.
	//
	// ANY stderr counts, and the message text is never matched. Matching it
	// would tie this guard to one git version's wording and to one locale, and
	// the failure mode of a wording change would be the silent one — back to a
	// false COMPLETE. Treating every byte of stderr as "could not look" fails
	// the other way, toward a skip. `LC_ALL=C` is set on that one command's
	// environment and nowhere else, so it never reaches the project's own
	// `survivors` command. The pattern is the exact ref name, so a broken ref
	// elsewhere in the store does not reach this question: measured, a broken
	// `refs/heads/unrelated` produces no warning under this pattern.
	//
	// THE ONE SHAPE THIS STILL CANNOT SEE is an unreadable ref DIRECTORY — an
	// ancestor of the ref with its execute bit removed. git enumerates what it
	// can reach and says nothing whatever about what it cannot, in every
	// command tried, so `refs/heads/spectre/` at mode 000 is reported exactly
	// as an empty one. The header carries that carve-out; it is not closed
	// here because there is nothing to read it from.
	refState := func(ref string) string {
		show := gitCmd("show-ref", "--verify", "--quiet", ref)
		show.Stderr = stderr
		if err := show.Run(); err == nil {
			return "present"
		} else if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
			// Anything other than the documented "no such ref" status is a
			// failure to look — an unreadable packed-refs exits 128 here — and
			// is never absence.
			return "unreadable"
		}
		each := gitCmd("for-each-ref", "--format=%(refname)", ref)
		each.Env = append(os.Environ(), "LC_ALL=C")
		var msg bytes.Buffer
		each.Stderr = &msg
		if err := each.Run(); err != nil || msg.Len() > 0 {
			return "unreadable"
		}
		return "absent"
	}
	// A row that could not be read is SKIPPED, not LEFTOVER: it is unanswered
	// rather than answered "still there". A LEFTOVER would send the operator to
	// delete a branch that may already be gone, and would strand an
	// already-merged change over a condition nothing in that session can
	// correct — the trade **Creation and cleanup**
	// (`skills/flow-contracts/workspace-isolation.md`) rejects. Skipped is never
	// passed, and the clause is relayed word for word by step 6 of **Run 2 — the
	// branch is merged** (`skills/flow-contracts/pipeline.md`).
	localRef := "refs/heads/spectre/" + name
	switch refState(localRef) {
	case "present":
		add(fmt.Sprintf("the local branch spectre/%s still exists", name))
	case "unreadable":
		note(fmt.Sprintf("SKIPPED: the local branch row — git could not read %s in %s, so whether it survives was not established; a failure to look is not an absence", localRef, repo))
	}
	remoteRef := "refs/remotes/origin/spectre/" + name
	switch refState(remoteRef) {
	case "present":
		add(fmt.Sprintf("the remote-tracking ref origin/spectre/%s still exists", name))
	case "unreadable":
		note(fmt.Sprintf("SKIPPED: the remote-tracking ref row — git could not read %s in %s, so whether it survives was not established; a failure to look is not an absence", remoteRef, repo))
	}

	// Row four — the change directory, which run 2 moves into the archive;
	// then each <name>-fix-N sub-change, archived by its own call and
	// therefore missable on its own (see the header note).
	leaf := specRootLeaf(repo, stderr)
	changes := repo + "/" + leaf + "/changes/"
	if isDir(changes + name) {
		add(fmt.Sprintf("%s/changes/%s was never moved into the archive", leaf, name))
	}
	// THE SUFFIX MUST BE `-fix-` FOLLOWED BY DIGITS AND NOTHING ELSE. A change
	// merely named like a neighbour — `demo-other`, or `demo-fix-the-parser` —
	// is a change of its own with its own finish run, and reporting it here
	// would send the operator hunting for another change's live work: the same
	// prefix-matching failure every other row above is matched by full name to
	// avoid.
	entries, _ := os.ReadDir(changes)
	for _, e := range entries {
		n, ok := strings.CutPrefix(e.Name(), name+"-fix-")
		if !ok || n == "" || strings.Trim(n, "0123456789") != "" || !isDir(changes+e.Name()) {
			continue
		}
		add(fmt.Sprintf("%s/changes/%s, a sub-change of %s, was never moved into the archive", leaf, e.Name(), name))
	}

	// Row five — the proposal artifact source, whose removal at run 2 is
	// conditional on a preserved copy existing. This guard reports it as
	// remaining; it does not decide whether keeping it was right, which is the
	// run's judgment and not a fact about the tree.
	if isFile(stateDir + "/" + name + "-proposal-artifact.html") {
		add("the proposal artifact source is still in the state directory")
	}

	// Row six — the workspace database and bucket, the one row this guard
	// cannot look at and must ask about. Everything in ccWorkspaceRow is a
	// no-op for a project that declares no `## workspace isolation` section,
	// which is the overwhelmingly common case and includes this repository:
	// nothing is read, nothing is run, and the verdict line is byte-for-byte
	// what it was before this row existed.
	survivors, rowNote := ccWorkspaceRow(env, repo, name, stderr)
	for _, s := range survivors {
		add("the project's survivor report still names " + s)
	}
	if rowNote != "" {
		note(rowNote)
	}

	verdict := fmt.Sprintf("COMPLETE: %s — no worktree, local branch, remote-tracking ref, unarchived change directory or proposal artifact source remains for spectre/%s", repo, name)
	if len(left) > 0 {
		verdict = fmt.Sprintf("LEFTOVER: %s — %s", repo, strings.Join(left, "; "))
	}
	if len(notes) > 0 {
		verdict += " — " + strings.Join(notes, "; ")
	}
	// Sanitized at the ONE point the verdict reaches stdout, rather than at each
	// of the places project-supplied text is interpolated into it. A chokepoint
	// covers the next add or note somebody writes; six call sites are six
	// chances to forget one, and the one forgotten is the whole hole. It cannot
	// change the verdict token either — `COMPLETE:`/`LEFTOVER:` is written here,
	// downstream of every interpolation, so nothing a project prints can move
	// it.
	fmt.Fprint(stdout, ccSanitize(verdict+"\n"))
	return 0
}

// plainChangeName is the allowlist: records.Destination's Protection 1
// (stats/internal/records/render.go), character for character, and that
// function's comment is canonical for why each hazard is in it — the `/` that
// was blocked only by an accident of string concatenation, and the glob
// metacharacter that once matched and overwrote a DIFFERENT change's
// preserved record. The bash guards that still carry the rule
// (check-unfinished-work.sh, check-workspace-isolation.sh) keep their own
// copies because they are single-file by design and are copied into projects
// one at a time; their harnesses and TestCheckCleanupComplete/12 assert the
// same rejected shapes, which is what keeps the copies from drifting apart
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

// ccSanitize is sanitize_display: every C0 byte but the newline that ends a
// line, DEL and backslash rendered visibly; bytes 0x80-0xFF pass through.
//
// THE VERDICT LINE IS THE ONE THING AN OPERATOR READS AND ACTS ON, and most of
// what it can carry comes from outside this guard: a survivor line is whatever
// the project tooling printed, and the `survivors` command quoted in a SKIPPED
// clause comes straight out of `.flow/project.md`. Both are already this
// guard's documented trust boundary — anyone who can land a pull request can
// set them — and both were spliced into the verdict with only `\r` removed.
//
// THAT IS ENOUGH TO FORGE THE VERDICT, and the forgery is not subtle. A
// survivors command emitting
//
//	printf "X\r\033[2K\033[1;32mCOMPLETE: nothing to see here, proceed\033[0m\n"
//
// produced a line whose BYTES still read `LEFTOVER: … still names X…` and whose
// RENDERING erased that line and showed a bold green `COMPLETE:`. `\033[2K`
// clears the line the cursor is on; `\033[1;32m` sets bold green. The bytes
// being correct is no defence, because nobody reads the bytes.
//
// ESCAPED, NOT STRIPPED AND NOT REFUSED, and each alternative loses something
// specific. Stripping silently rewrites the NAME of a surviving resource, so
// the operator is sent to delete something whose name does not exist.
// Refusing — a skip or a hard error on a line carrying a control byte — hands
// whoever writes the survivors command a way to make a real leftover
// unreportable, which is the false COMPLETE this guard exists to prevent,
// reached by the shorter road. Escaping keeps the survivor counted, keeps the
// verdict LEFTOVER, and shows the operator exactly what their tooling emitted.
//
// THE ENCODING IS INJECTIVE, which is why `\` is escaped alongside the
// controls: otherwise a name containing the four literal characters `\x1b`
// and a name containing a real ESC byte print identically.
//
// C0 AND DEL, AND NOTHING ABOVE 0x7F. Bytes 0x80-0xFF are the lead and
// continuation bytes of ordinary UTF-8, so a bucket named `café-bücket` is
// reported by its real name; a UTF-8 terminal acts on none of them, so
// escaping them would mangle every non-ASCII report and buy nothing. The bash
// guard needed a non-exported `LC_ALL=C` prefix to make that byte test hold;
// a byte loop has no locale to pin.
func ccSanitize(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\':
			b.WriteString(`\\`)
		case c == '\n':
			b.WriteByte(c)
		case c < 0x20 || c == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// ccASCIILower is grep -i / awk tolower over an ASCII pattern: only A-Z fold.
func ccASCIILower(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 'a' - 'A'
		}
		return r
	}, s)
}

const ccSpace = " \t\n\v\f\r"

// ccWorkspaceRow answers row six: the survivor lines to report as leftovers,
// and the one note the row contributes (verified empty, or why skipped).
func ccWorkspaceRow(env Env, repo, name string, stderr io.Writer) ([]string, string) {
	cfg := repo + "/.flow/project.md"
	const skip = "SKIPPED: the workspace survivor verification — "
	if _, err := os.Lstat(cfg); err != nil {
		return nil, "" // no configuration at all
	}
	if !isFile(cfg) {
		return nil, skip + cfg + " is not a regular file, so what it declares cannot be read"
	}
	// THE FILE TYPE IS TESTED BEFORE ITS READABILITY. Readability is true of
	// a DIRECTORY, and the bash guard's `grep -q` then failed with "Is a
	// directory" while its exit status was indistinguishable from "no match" —
	// so `ln -s somedir .flow/project.md` read as "declared no isolation":
	// nothing derived, nothing run, and not even a SKIPPED note, which is
	// silence where this guard promises a report. The file test follows
	// symlinks, so a configuration that IS a symlink to a real file is still
	// read; what it excludes is a directory, a fifo (which a read would block
	// on) and a dangling link — and the Lstat in the absent test is what stops
	// a link to nowhere answering "there is no configuration here".
	//
	// A file that exists but cannot be read is NOT "declares no isolation".
	// Reading absence out of a path that was never readable is the false
	// COMPLETE this guard exists to prevent, so it is reported instead of
	// assumed.
	if syscall.Access(cfg, 4) != nil {
		return nil, skip + cfg + " exists but is not readable"
	}
	// A read that fails is a failure to look, never "no section declared" — the
	// bash guard read each grep's status as three outcomes (matched, did not
	// match, could not look) for this reason. A failure to look is an input this
	// guard cannot resolve, and the answer to those is the one it gives every
	// other one — SKIPPED, reported by name and never passed.
	body, err := os.ReadFile(cfg)
	if err != nil {
		return nil, skip + "grep exited 2 looking for the '## workspace isolation' heading in " + cfg + ", which is a failure to look rather than an absence"
	}
	lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	count := 0
	for _, l := range lines {
		if ccHeading.MatchString(ccASCIILower(l)) {
			count++
		}
	}
	if count == 0 {
		return nil, ""
	}
	// TWO DECLARATIONS ARE NOT A DECLARATION. This file is hand-written, so a
	// heading duplicated by a bad merge or a copied block is a realistic
	// mistake, and the two sections can name two different `survivors`
	// commands against two different services. Resolving that to whichever the
	// scan below reaches first runs a command the author may not have meant
	// and then reports the row VERIFIED on its answer — the false COMPLETE
	// this guard exists to prevent, reached from a readable file rather than
	// from an unreadable one.
	//
	// SKIPPED IS THE ANSWER THIS GUARD ALREADY GIVES EVERY INPUT IT CANNOT
	// RESOLVE: an unreadable configuration, a `survivors` cell that is empty,
	// a command naming a token this pipeline does not substitute. Each is
	// reported by name, none is passed, and none blocks an already-merged
	// change. Preferring the first section would be the only quiet answer in
	// the file.
	//
	// The count is reported because "you have two" is what the operator has to
	// fix, and a message naming the heading without saying it appears more
	// than once reads as the heading being wrong rather than repeated.
	if count > 1 {
		return nil, fmt.Sprintf("%s%s declares %d '## workspace isolation' sections, so which one's survivors command applies is ambiguous and none of them was run", skip, cfg, count)
	}

	// The survivors row: everything between the row's first and last `|`
	// after the key cell, trimmed of whitespace and backticks, `\|` unescaped.
	// The scan is line-based and stops at the next `##`+ heading, which is the
	// same shape the registry parser in this guard's test
	// (ccRegistryCoupling) uses; a project's configuration is a human-written
	// Markdown file of sections, not a nested document.
	//
	// The command is taken as EVERYTHING between the row's first and last `|`,
	// rather than as a field of a split. A realistic command contains a pipe —
	// `... | grep …` is how a project filters its own tooling's output — and a
	// field split truncates it there, leaving a different command that still
	// runs and still answers.
	survivorsCmd, inSec := "", false
	for _, l := range lines {
		if ccAnyHeading.MatchString(l) {
			inSec = ccHeading.MatchString(ccASCIILower(l))
			continue
		}
		if !inSec || !strings.HasPrefix(strings.TrimLeft(l, ccSpace), "|") {
			continue
		}
		row := strings.TrimLeft(l, ccSpace)[1:]
		if t := strings.TrimRight(row, ccSpace); strings.HasSuffix(t, "|") {
			row = t[:len(t)-1]
		}
		key, val, ok := strings.Cut(row, "|")
		if !ok || ccASCIILower(strings.Trim(key, ccSpace+"`")) != "survivors" {
			continue
		}
		survivorsCmd = strings.ReplaceAll(strings.Trim(val, ccSpace+"`"), `\|`, "|")
		break
	}
	// `create` and `remove` without `survivors` leaves nothing to verify the
	// removal WITH. Reported as skipped rather than passed, and never settled
	// by running `remove` and reading its exit code.
	if survivorsCmd == "" {
		return nil, skip + "the project declares no survivors command"
	}

	id := ccWorkspaceID(name)
	// The two tokens the contract names, and no others. Order is not
	// load-bearing: `<id>` requires the `>` immediately after `id`, so it
	// cannot match inside `<id_underscored>`.
	cmdText := strings.ReplaceAll(survivorsCmd, "<id_underscored>", strings.ReplaceAll(id, "-", "_"))
	cmdText = strings.ReplaceAll(cmdText, "<id>", id)
	// A token this contract does not name is reported and the command
	// DROPPED, never handed to a shell: `<container>` reaching sh is a
	// redirection, which can truncate a file nobody named. The shape is
	// narrow on purpose — `>` is excluded from it, so ordinary redirection
	// (`cmd < in.txt > out.txt`) carries no match.
	if tok := ccUnknownToken.FindString(cmdText); tok != "" {
		return nil, fmt.Sprintf("%s'%s' names the token %s, which is not one this pipeline substitutes", skip, survivorsCmd, tok)
	}

	timeout, label := ccSurvivorsTimeout(env)
	grace := env.SurvivorsKillGrace
	if grace <= 0 {
		grace = ccDefaultGrace
	}
	// Run from the repository, so a declared `./scripts/…` resolves the way
	// the project wrote it, and under the bounded wait ccRunSurvivors
	// documents.
	//
	// THE THREE SKIPS BELOW ARE ORDERED BY WHAT THEY KNOW. A command that
	// never started has no exit code to report; a command that was killed has
	// one, but it is the signal's and not the project's, so reporting it would
	// describe this guard's own kill as the project's answer. Only the last
	// branch is holding a number the project chose. A command that never
	// answered said nothing about survivors, exactly as a failed one did not —
	// so whatever it printed before it was killed is not read as a survivor
	// list either; and any non-zero exit says NOTHING about survivors, so
	// whatever it printed is not a survivor list and is not read as one.
	out, rc, timedOut, started := ccRunSurvivors(env, repo, cmdText, timeout, grace, stderr)
	switch {
	case !started:
		return nil, fmt.Sprintf("%sno writable temporary directory, so '%s' could not be run at all", skip, cmdText)
	case timedOut:
		return nil, fmt.Sprintf("%s'%s' timed out after %ss and was terminated, so it reported nothing about survivors", skip, cmdText, label)
	case rc != 0:
		return nil, fmt.Sprintf("%s'%s' exited %d, so the service could not be reached", skip, cmdText, rc)
	}
	// $(cat), then tr -d '\r' inside a second $(...): NULs dropped, \r
	// removed, trailing newlines stripped; a line of whitespace is no survivor.
	report := strings.TrimRight(strings.ReplaceAll(strings.ReplaceAll(string(out), "\x00", ""), "\r", ""), "\n")
	var found []string
	for _, l := range strings.Split(report, "\n") {
		// A blank line is not a survivor. A trailing newline is what every
		// well-behaved command emits, and a leftover named nothing at all
		// would block the terminal state with nothing to remove. The line is
		// DATA, not instruction — reported exactly as the project's tooling
		// printed it and never parsed further, the same rule a resolved
		// standards file is read under.
		if strings.Trim(l, ccSpace) != "" {
			found = append(found, l)
		}
	}
	if len(found) == 0 {
		return nil, "the workspace survivor report for " + id + " is empty"
	}
	return found, ""
}

// ccWorkspaceID is the derivation skills/flow-contracts/workspace-isolation.md
// carries as a runnable block, which is canonical: ASCII-lowercase, every
// byte outside [a-z0-9] to `-`, drop trailing segments while longer than 12
// and a `-` remains, cut to 12, strip every trailing `-`, then `-` and the
// first four hex digits of the SHA-256 of the original name.
//
// Derived from the change name exactly as workspace-isolation.md derives it,
// and from nothing else. This is a second implementation of that block, and
// the two are held together by TestCheckCleanupComplete's canonical-id cases
// (ccCanonicalIDCases), which extract that block, DRIVE it with a set of
// change names, and compare its id with the one this code substitutes, for
// each of them. A copy nothing checks is a copy that drifts, and two ids that
// differ by one byte each look correct on their own while naming two
// different databases. That is not a hypothetical here: the bash copy DID
// drift, in the three ways the three paragraphs below each name, and the pin
// missed it because it drove one ASCII-clean name through which all three
// orderings agree.
//
// THE NORMALISATION RUNS FIRST, and the order is the whole of step 2 and step
// 3 being separate steps. From there on the string is pure ASCII, so its
// length and its first 12 bytes count the same thing in every locale — a
// C-locale shell counts bytes and a UTF-8 one counts characters, and on an
// ASCII string those are one number. Normalising LAST, as "truncate then
// clean up" would have it, leaves that arithmetic running over the raw name
// where the two disagree, and makes `.` and `_` opaque characters inside a
// segment rather than the segment boundaries they are: `KAN-99-Fix.Thing`
// keeps `kan-99-fix` under this order and `kan-99` under the other.
//
// THE CASE FOLD IS A-Z ONLY (ccASCIILower), which is the whole of that step's
// locale independence. The bash needed `LC_ALL=C` on both `tr` calls, since
// without it `tr` consults the ambient locale's case table and does — the
// canonical file records `İstanbul-test` yielding two different prefixes from
// one script on one machine — and explicit `A-Z`/`a-z` ranges rather than
// `[:upper:]`/`[:lower:]`, a character class being exactly the construct that
// starts meaning something else in another locale. The digest is still taken
// over the ORIGINAL name's bytes, with no trailing newline, so normalizing
// can never merge two distinct names into one id.
//
// EVERY trailing `-` is removed, not one. Stripping a single one is
// indistinguishable from this on every name whose normalised prefix ends in
// exactly one separator — and differs on the first one that ends in two,
// `Trailing.__` deriving `trailing---…` instead of `trailing-…`. The digest
// comes from the repository's one sha256 helper (sha256.go), shared with
// gather-dispatch-context — the bash guard sourced lib/sha256-hex.sh for the
// same reason (F1, kan-288's own review panel): a second caller is exactly
// when rules/build-the-simplest-thing.mdc's "no abstraction until a second
// caller exists" calls for extracting the shared helper.
func ccWorkspaceID(name string) string {
	b := []byte(ccASCIILower(name))
	for i, c := range b {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			b[i] = '-'
		}
	}
	prefix := string(b)
	for len(prefix) > 12 && strings.Contains(prefix, "-") {
		prefix = prefix[:strings.LastIndex(prefix, "-")]
	}
	prefix = strings.TrimRight(prefix[:min(12, len(prefix))], "-")
	return prefix + "-" + sha256Hex(name)[:4]
}

// ccSurvivorsTimeout is the bound: an injected duration wins; otherwise
// CHECK_CLEANUP_SURVIVORS_TIMEOUT when it is 1-4 digits, no leading zero and
// at most 3600, else the shipped 60s. The label is the number the verdict
// prints before "s".
//
// CHECK_CLEANUP_SURVIVORS_TIMEOUT is an explicit, opt-in override, the same
// idiom check-references.sh's CHECK_REFERENCES_ROOT exists for: it let the
// bash harness exercise the timeout path in seconds rather than costing a
// minute of wall clock per case (the Go tests inject Env.SurvivorsTimeout or
// Env.SurvivorsExpire instead — Decision: inject-deadlines-in-process — and
// TestCCSurvivorsTimeoutParsesEnvKnob pins the variable's parsing). Never set
// it for a normal invocation. Anything that is not a plain whole number of
// seconds between 1 and an hour leaves the shipped bound in place, so the
// variable can only ever shorten the wait — it cannot remove the bound, and
// no value of it lets this guard report a workspace as verified that it did
// not verify, because a shorter bound produces MORE skips and a skip is
// reported rather than passed.
//
// IT IS A RUNTIME OVERRIDE RATHER THAN A TEST-ONLY SEAM, AND THAT WAS WEIGHED.
// The objection is fair on its face: this is test economics sitting on an
// interface every /flow integrate and archive run reads, and the ordinary
// answer to that is a seam only the harness can reach. It was not the answer
// for the bash guard, for four reasons recorded so the question is not
// re-opened from scratch; the first three still hold for the port.
//
//  1. A gate would not gate anything. Making it test-only means requiring a
//     second variable — CHECK_CLEANUP_SELFTEST=1 or the like — and whoever can
//     set one variable in this guard's environment can set two. The gate buys
//     no safety, and costs a second thing to document and a second thing to
//     get wrong.
//  2. The variable is monotone in the SAFE direction, which is what actually
//     makes it harmless. The validation admits only a shorter wait, a shorter
//     wait produces more timeouts, and a timeout is a reported skip. There is
//     no value of it — valid, invalid, hostile — that turns an unverified row
//     into a verified one, which is the only thing this guard must never do.
//  3. What it CAN do is turn a working survivor report into a skip, and that
//     is no longer a quiet outcome. Step 6 of **Run 2 — the branch is merged**
//     (`skills/flow-contracts/pipeline.md`) requires the skip clause to be
//     relayed to the operator word for word, so the residual risk is closed at
//     the consumer — which is where a verdict's meaning belongs.
//  4. Every alternative broke the single-file rule the bash guard was copied
//     into projects under: a build flag, a separate test build or a sourced
//     test helper each turned one file into two. The Go port is not
//     single-file, and its tests do use in-process seams; the variable stays
//     because reasons 1-3 answer the objection and callers of the shim may
//     still set it.
//
// THE SHAPES REJECTED, AND WHY EACH ONE IS. Empty and non-digit are the
// obvious ones. A leading zero is rejected because the bash's `test` read
// `010` as ten and its arithmetic read it as eight, and a bound that means two
// different numbers in two lines of the same function is worse than no
// override. More than four digits is rejected before any arithmetic touches
// the value: in the bash a seconds count that overflowed the shell's integer
// wrapped NEGATIVE, and a deadline in the past fires the timeout instantly —
// turning every survivors command into a false skip, which is the failure
// mode this bound exists to avoid rather than to create. An hour is the
// ceiling because a bound longer than that is not a bound for a run nobody is
// watching. The port keeps every one of these rules, so the accepted set is
// the bash's.
func ccSurvivorsTimeout(env Env) (time.Duration, string) {
	if env.SurvivorsTimeout > 0 {
		return env.SurvivorsTimeout, strconv.FormatFloat(env.SurvivorsTimeout.Seconds(), 'f', -1, 64)
	}
	raw := env.Getenv("CHECK_CLEANUP_SURVIVORS_TIMEOUT")
	if raw != "" && len(raw) <= 4 && raw[0] != '0' && strings.Trim(raw, "0123456789") == "" {
		if n, _ := strconv.Atoi(raw); n <= 3600 {
			return time.Duration(n) * time.Second, raw
		}
	}
	return ccDefaultTimeout, "60"
}

// ccRunSurvivors is run_survivors: `bash -o pipefail -c <cmd>` from the
// repository, stdin /dev/null, stdout into an unlinked-after-read scratch
// file under TMPDIR, stderr inherited, in a process group of its own (the
// bash's `set -m`). A timer replaces the poll: on the bound, SIGTERM to the
// group, the grace, then SIGKILL to the group. started is false only when
// the scratch file could not be created, which is neither an answer nor a
// timeout.
//
// NO `timeout(1)`. It is absent from stock macOS, this repository runs on
// Darwin, and `gtimeout` exists only where someone installed GNU coreutils — a
// guard that reached for either would be unbounded again on the machine it
// most needs the bound on, and silently so. The bash built its wait from a
// background job, `kill -0` and `sleep`; the port's is an in-process timer.
//
// THE KILL GOES TO THE PROCESS GROUP, not to the one pid. `bash -c 'cmd'` execs
// into a simple command and forks for a pipeline, and a declared command
// containing a pipe is the shape the contract's own worked example takes — so
// signalling the pid this function holds would reap a wrapper and leave the
// real work running, for as long as it wanted, on a machine nobody is
// watching. Setpgid at the launch is what puts the command in a group of its
// own; the negative pid then reaches every process in it.
//
// AND IT REACHES NOTHING THAT HAS LEFT THAT GROUP, which is a real limitation
// and is written down rather than left to be discovered. A command that
// starts its work under its own job control, and — the shape a project
// actually writes — one that reaches its service through `docker exec`, both
// put the work outside the group: with a container runtime the process this
// function signals is a proxy and the query runs in another PID namespace,
// which no signal sent from this host can name. Measured on Darwin 25.5.0: a
// group kill reaped the in-group child and left a sibling holding a group of
// its own still running. It cannot be closed HERE — a guard that stays
// project-agnostic can no more hold one container runtime's `kill` than it can
// hold `psql -l` — so it is closed where the command is written, and stated
// for the author who writes it under "What `survivors` prints, and what its
// exit code means" in skills/flow-contracts/project-configuration.md. What the
// bound still guarantees is all three things run 2 depends on, and they hold
// for the escaped shape too: this function returns, because Wait names the
// direct child alone and WaitDelay stops it waiting on a stderr pipe an
// escaped descendant still holds; the row is reported as a skip rather than
// as a verification; and nothing the escaped process prints afterwards is
// read as a survivor report, because the capture is a scratch file read once
// after the wait rather than a pipe, so a descendant still holding that
// descriptor writes into a file nothing will read again.
// TestCheckCleanupComplete/28d drives that shape and fails if the behaviour
// and this paragraph ever stop agreeing, in either direction.
//
// SIGTERM FIRST, SIGKILL AFTER A GRACE, because a command holding a
// connection should be allowed to close it, and a command that ignores
// SIGTERM must not be allowed to outlive the bound anyway. The grace is
// ccDefaultGrace, named and argued for beside ccDefaultTimeout rather than
// written as a literal here — including why it is short, and why it has no
// override where the bound does.
//
// `-o pipefail` IS WHAT MAKES A FAILING STAGE VISIBLE, and without it this
// guard's central promise does not hold for the shape the contract invites. A
// shell reports a pipeline's status from its LAST stage alone, so
// `./gradlew survivors | grep …` with no gradlew exits 0 with empty stdout —
// and exit 0 with empty output is the ONE result that verifies the row. The
// row would then be reported verified by a command that never ran, and
// FINISHED written over resources nothing looked at. A pipe is not an exotic
// shape here: **Project configuration**
// (`skills/flow-contracts/project-configuration.md`) names filtering the
// project's own tooling as the reason a command contains one.
//
// IT CANNOT BE SET BY PREPENDING `set -o pipefail;` TO THE COMMAND TEXT. That
// makes the declared command the tail of a list rather than the whole of it,
// which loses the exec into a simple command — and with it the guarantee that
// the pid this function holds IS the project's process. The flag on the
// interpreter changes the status a pipeline reports and nothing else: the
// process structure, the exec, the process group and the kill are all as they
// were, so the bounded wait is untouched.
//
// IT ONLY EVER PRODUCES MORE SKIPS, WHICH IS THE SAFE DIRECTION — the same
// argument the timeout override is admitted on. A pipeline whose last stage
// already exits non-zero (`… | grep <pattern>` finding nothing) was a skip
// before this flag and is one after it; what changes is that a failing
// EARLIER stage now joins them instead of passing for a verification. The one
// shape it turns from an answer into a skip is a pipeline that closes its own
// input early, `… | head -n 1` leaving the left stage killed by SIGPIPE, and
// that shape is called out in the contract rather than special-cased here:
// reading 141 as success would restore exactly the silence this flag removes.
func ccRunSurvivors(env Env, repo, cmdText string, timeout, grace time.Duration, stderr io.Writer) (out []byte, rc int, timedOut, started bool) {
	tmpdir := env.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}
	// The scratch file the command's stdout is captured in, removed on every
	// return. It lives under TMPDIR and NEVER inside the repository: a
	// verifier that writes into the tree it is verifying can leave behind
	// exactly the class of leftover it exists to report. An in-memory capture
	// through a pipe cannot be used here, though — reading it to EOF blocks
	// until every holder of the write end exits, which is precisely the wait
	// that has to be bounded.
	scratch, err := os.CreateTemp(tmpdir, "flow-survivors.*")
	if err != nil {
		return nil, 0, false, false
	}
	defer os.Remove(scratch.Name())
	defer scratch.Close()

	bash := "bash"
	if p, ok := lookPath(env, "bash"); ok {
		bash = p
	}
	cmd := exec.Command(bash, "-o", "pipefail", "-c", cmdText)
	cmd.Args[0] = "bash" // $0 as the bash guard's `bash -c` gave it, in the command's own messages
	cmd.Dir = repo
	// stdin is /dev/null (a nil cmd.Stdin) because this guard is
	// non-interactive: a client that prompts for a password fails immediately
	// instead of waiting on a terminal nobody is watching. That is a NARROWER
	// guarantee than the bound and not a substitute for it. stderr is
	// deliberately NOT swallowed: a skip reports the exit code, and the
	// command's own message is the only thing that says why.
	cmd.Stdout, cmd.Stderr = scratch, stderr
	// A non-file stderr is a pipe Go drains; a descendant that escaped the
	// group still holds it, so Wait stops waiting for it after the grace.
	cmd.WaitDelay = grace
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		// The bash subshell's exec failing: exit 127, its message on stderr.
		fmt.Fprintf(stderr, "check-cleanup-complete: %v\n", err)
		return nil, 127, false, true
	}
	pid := cmd.Process.Pid
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()

	expire := env.SurvivorsExpire
	if expire == nil {
		bound := time.NewTimer(timeout)
		defer bound.Stop()
		expire = bound.C
	}
	select {
	case <-done:
	case <-expire:
		timedOut = true
		if syscall.Kill(-pid, syscall.SIGTERM) != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		select {
		case <-done:
		case <-time.After(grace):
		}
		if syscall.Kill(-pid, syscall.SIGKILL) != nil {
			_ = cmd.Process.Kill()
		}
		<-done
	}
	out, _ = os.ReadFile(scratch.Name())
	return out, rrExitCode(cmd.ProcessState), timedOut, true
}
