package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func init() { Registry["check-visual-verification"] = checkVisualVerification }

// checkVisualVerification is scripts/check-visual-verification.sh: validate a
// project's `## visual verification` declaration against the contract that
// defines it. The script's header is the contract — arguments, verdict lines
// and exit codes. What follows are the body's reasons, beside the code they
// explain.
//
// THE INPUT IS ATTACKER-INFLUENCED: `.flow/project.md` is tracked and editable
// in any pull request. NOTHING read here is executed — this guard never runs
// `setup`, `verify`, `capture`, `fingerprint`, `start` or `specs`, and never
// interpolates a cell into a shell. Every violation line, and every cell
// interpolated into a message afterwards, passes through sanitizeDisplay
// before it reaches the terminal.
func checkVisualVerification(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(msg string) int {
		fmt.Fprintln(stderr, "check-visual-verification: "+msg)
		return 2
	}
	if len(args) != 1 {
		return refuse("usage: check-visual-verification.sh <project root>")
	}
	root := args[0]
	if !isDir(wiAbs(env, root)) {
		return refuse(root + " is not a directory — cannot tell whether it declares a visual verification section")
	}
	cfg := root + "/.flow/project.md"
	path := wiAbs(env, cfg)

	// The file is optional. No `.flow/project.md` at all is a supported,
	// ordinary case, not an error — matching check-workspace-isolation's own
	// treatment of the identical absence. An Lstat, so a dangling symlink is
	// not "nothing here" either.
	if _, err := os.Lstat(path); err != nil {
		fmt.Fprintf(stdout, "VISUAL-OK: %s — no .flow/project.md, so nothing is declared\n", root)
		return 0
	}
	// A REGULAR FILE, TESTED BEFORE READABILITY, because readability answers
	// about a DIRECTORY too, and the bash's `grep` then failed with "Is a
	// directory" while its exit status was indistinguishable from "no match".
	if !isFile(path) {
		return refuse(cfg + " is not a regular file — cannot validate what it declares")
	}
	if syscall.Access(path, 4) != nil {
		return refuse(cfg + " exists but is not readable — cannot validate what it declares")
	}
	// A READ THAT FAILS IS A FAILURE TO LOOK, never an absence: the bash read
	// grep's three answers as three (matched, did not match, could not look),
	// and this is the third, refused with the message the bash gave it.
	body, err := os.ReadFile(path)
	if err != nil {
		return refuse("grep exited 2 while looking for the '## visual verification' heading in " + cfg + " — that is a failure to look rather than an absence, so nothing was validated")
	}
	// Every read below runs on vvLines — BOM stripped, each line ended at its
	// first NUL as grep and awk ended it.
	text := string(body)

	count := vvHeadingCount(text)
	if count == 0 {
		fmt.Fprintf(stdout, "VISUAL-OK: %s — declares no `## visual verification` section\n", cfg)
		return 0
	}
	// TWO DECLARATIONS ARE NOT A DECLARATION. This file is hand-written, so a
	// heading duplicated by a bad merge or a copied block is a realistic
	// mistake, and the two sections could name two different regression
	// repos. Neither is validated; validating the first would be this guard
	// picking the winner the contract declines to pick.
	if count > 1 {
		fmt.Fprint(stdout, sanitizeDisplay(fmt.Sprintf("%s:0: the file declares %d `## visual verification` sections — a second declaration is an ambiguous declaration rather than a merge, so neither was validated\n", cfg, count)))
		fmt.Fprintf(stdout, "VISUAL-INVALID: %s — 1 violation(s) in the `## visual verification` section\n", cfg)
		return 1
	}

	// The bash ran the table scan as an awk program and refused when it
	// printed no `#SUMMARY` sentinel (no awk, a syntax error, a signal); this
	// scan is in-process code with no such way to stop short, and the shim is
	// what refuses when flow-guard itself cannot run.
	v := &vvValidation{cfg: cfg, set: map[string]vvField{}, cmd: map[string]vvField{}}
	heading := 0
	for i, l := range vvLines(text) {
		if vvHeading.MatchString(gsASCIILower(l)) {
			heading = i + 1
			break
		}
	}
	v.scan(vvSectionLines(text))

	// The four "absent or empty" findings for the required settings and
	// commands.
	for _, req := range []struct {
		fields map[string]vvField
		key    string
	}{{v.set, "ui paths"}, {v.set, "screenshots"}, {v.cmd, "verify"}, {v.cmd, "capture"}} {
		f, ok := req.fields[req.key]
		if !ok {
			v.addv(heading, "`"+req.key+"` is absent — this setting is required")
		} else if f.val == "" {
			v.addv(f.line, "`"+req.key+"` is empty — this setting is required and non-empty")
		}
	}

	// `ui paths` YIELDING NO USABLE GLOB IS THE WORSE HALF OF THIS DEFECT, per
	// design.md's `validator-agrees-with-trigger`: the check above only
	// rejects a wholly empty cell, but a cell that is non-empty as a whole —
	// e.g. two empty backtick pairs joined by a comma — resolves, after the
	// SAME split-then-strip parse check-visual-trigger applies (split on
	// comma, then trimGlobElement each element), to zero usable elements.
	// That is exactly the condition under which the trigger guard itself
	// exits 2 ("resolved to no usable glob"), so a validator that stayed
	// silent on it would keep blessing a declaration the trigger cannot use.
	// Runs only when `ui paths` is present and non-empty — an absent or empty
	// cell already has its own finding above.
	if f, ok := v.set["ui paths"]; ok && f.val != "" {
		usable := 0
		for _, e := range strings.Split(f.val, ",") {
			if trimGlobElement(e) != "" {
				usable++
			}
		}
		if usable == 0 {
			v.addv(f.line, "`ui paths` is `"+f.val+"` — splitting on commas and stripping each element's backticks and whitespace yields no usable glob, so check-visual-trigger.sh cannot use it either")
		}
	}

	// `regression repo` present with no `regression checkout`, or the reverse
	// — design.md names no use for either half declared alone: a repo with
	// nothing to check its origin against, or a checkout nothing names as the
	// repository it must equal.
	checkout, hasCheckout := v.set["regression checkout"]
	repo, hasRepo := v.set["regression repo"]
	if hasRepo && !hasCheckout {
		v.addv(repo.line, "`regression repo` is present with no `regression checkout` — a repo with nothing to check its origin against declares nothing checkable")
	}
	if hasCheckout && !hasRepo {
		v.addv(checkout.line, "`regression checkout` is present with no `regression repo` — nothing names the repository its origin must equal")
	}

	valid := false
	if hasCheckout {
		if !isDir(wiAbs(env, checkout.val)) {
			v.addv(checkout.line, "`regression checkout` names `"+checkout.val+"`, which is not an existing directory")
		} else if vvGitClean(env, "-C", checkout.val, "rev-parse", "--git-dir").Run() != nil {
			v.addv(checkout.line, "`regression checkout` names `"+checkout.val+"`, which is not a git checkout")
		} else {
			valid = true
		}
	}

	// THE SECURITY-RELEVANT CHECK. `regression repo` is an identity
	// assertion, not an authorisation — nothing is granted by a match, per
	// design.md's `no-automatic-push` decision, which dropped this contract's
	// earlier `push to default branch` setting after a panel slot
	// demonstrated that both sides of this same equality lived in this one
	// pull-request-editable file and so authorised nothing. Run whenever
	// there is something to verify: a checkout that validated as a real git
	// checkout above, and a declared repo to compare its origin against.
	// Every other combination already has its own finding above, so running
	// this too would report the same misconfiguration twice.
	if valid && hasRepo {
		cmd := vvGitClean(env, "-C", checkout.val, "remote", "get-url", "origin")
		var out, errOut bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errOut
		rc := 0
		if err := cmd.Run(); err != nil {
			rc = 127
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				rc = exit.ExitCode()
				if ws, ok := exit.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
					rc = 128 + int(ws.Signal())
				}
			}
		}
		// `$(…)` dropped every trailing newline.
		origin := strings.TrimRight(out.String(), "\n")
		gitErr := strings.TrimRight(errOut.String(), "\n")
		switch {
		case rc == 0:
			if origin != repo.val {
				v.addv(checkout.line, "the checkout's real `origin` (`"+origin+"`) does not equal the declared `regression repo` (`"+repo.val+"`)")
			}
		case strings.Contains(gitErr, "No such remote") || strings.Contains(gitErr, "no such remote"):
			// `git remote get-url origin` FAILING FOR "NO SUCH REMOTE" IS A
			// FINDING, NOT A 2: a real, valid checkout with no `origin` remote
			// at all cannot honour the identity assertion either — a fact
			// about the declaration, not a failure to look.
			v.addv(checkout.line, "the checkout at `"+checkout.val+"` has no `origin` remote configured, so it cannot match the declared `regression repo` (`"+repo.val+"`)")
		default:
			// Any OTHER failure is this guard failing to look, and escalates
			// the whole run rather than folding into a "clean" verdict (none
			// reproduces against real git 2.50.1 once `rev-parse --git-dir`
			// accepted the checkout, which is why the test reaches it through
			// a stub git). The checkout is a cell out of `.flow/project.md`
			// and gitErr is git's own stderr against it, so both go through
			// sanitizeDisplay as every other interpolated cell does.
			fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-visual-verification: git remote get-url origin failed unexpectedly in %s (exit %d): %s — cannot determine whether the identity assertion holds\n", checkout.val, rc, gitErr)))
			return 2
		}
	}

	if len(v.viol) != 0 {
		fmt.Fprint(stdout, sanitizeDisplay(strings.Join(v.viol, "")))
		fmt.Fprintf(stdout, "VISUAL-INVALID: %s — %d violation(s) in the `## visual verification` section\n", cfg, len(v.viol))
		return 1
	}
	fmt.Fprintf(stdout, "VISUAL-OK: %s — the `## visual verification` section validated\n", cfg)
	return 0
}

// vvGitClean is the since-deleted lib/git-clean.sh's git_clean: git run with every AMBIENT
// `GIT_*` environment variable removed, so this guard's own repo-selection
// decisions (which checkout `-C` points at, and what its `origin` resolves
// to) cannot be overridden by whatever invoked it. `GIT_DIR` is the
// demonstrated vector: git honours it over `-C`'s repo selection, so
// `GIT_DIR=/anywhere/.git` makes both `rev-parse --git-dir` and `remote
// get-url origin` answer about that repository instead of the one `-C`
// named — including reporting a `regression checkout` that is not a git
// repository at all as one. Git hooks set `GIT_DIR` automatically, so a
// guard invoked from one runs pre-bypassed without this. `GIT_WORK_TREE`,
// `GIT_COMMON_DIR`, `GIT_OBJECT_DIRECTORY`, `GIT_INDEX_FILE`,
// `GIT_CEILING_DIRECTORIES` and every `GIT_CONFIG*` variable redirect repo
// or config resolution the same way, so EVERY `GIT_`-prefixed variable
// present is dropped — read from the environment rather than a fixed list
// the next git release could add to. envGit keeps the whole environment,
// which is why this guard does not run git through it bare.
func vvGitClean(env Env, args ...string) *exec.Cmd {
	cmd := envGit(env)(args...)
	cmd.Env = []string{} // never nil: nil inherits every variable
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	return cmd
}

// vvField is one recognised, non-duplicate setting or command row.
type vvField struct {
	line int
	val  string
}

// vvValidation is the port of the bash's awk table scan, whose only input
// was the file's text, plus the violations the checks after it add.
type vvValidation struct {
	cfg      string
	viol     []string // `<cfg>:<line>: <message>\n`, unsanitised
	set, cmd map[string]vvField
}

func (v *vvValidation) addv(line int, msg string) {
	v.viol = append(v.viol, v.cfg+":"+strconv.Itoa(line)+": "+msg+"\n")
}

// scan reads the section's lines as contiguous runs of table rows. A run
// ends at any non-table line and at a heading — a `###` subheading keeps the
// section open (project-configuration.md allows prose beside the two tables,
// and a subheading is exactly that) but still ends the table above it, which
// shows here as a gap in the line numbers vvSectionLines leaves out.
func (v *vvValidation) scan(lines []vvLine) {
	var blk []vvLine
	for _, l := range lines {
		isRow := strings.HasPrefix(strings.TrimLeft(strings.ReplaceAll(l.Text, "\r", ""), cSpace), "|")
		if len(blk) > 0 && (!isRow || blk[len(blk)-1].No+1 != l.No) {
			v.flushBlock(blk)
			blk = nil
		}
		if isRow {
			blk = append(blk, l)
		}
	}
	if len(blk) > 0 {
		v.flushBlock(blk)
	}
}

const (
	vvSettingHeader = "setting|value"
	vvCommandHeader = "command|runs"
)

func vvHeader(cells []string) string {
	folded := make([]string, len(cells))
	for i, c := range cells {
		folded[i] = vtFoldCell(c)
	}
	return strings.Join(folded, "|")
}

var vvDelimiterCell = regexp.MustCompile(`^:?-+:?$`)

func vvIsDelimiter(cells []string) bool {
	for _, c := range cells {
		if !vvDelimiterCell.MatchString(vtTrimCell(c)) {
			return false
		}
	}
	return len(cells) > 0
}

// flushBlock checks one contiguous run of table rows. Its first row is the
// header, and the header is what says which of the two tables this is.
func (v *vvValidation) flushBlock(blk []vvLine) {
	hdr := vvHeader(vtSplitCells(blk[0].Text))
	var kind string
	switch hdr {
	case vvSettingHeader:
		kind = "setting"
	case vvCommandHeader:
		kind = "command"
	default:
		// A HEADER THIS GUARD DOES NOT RECOGNISE IS A VIOLATION, never a table
		// it skips — skipping would let a reordered or misspelled header leave
		// every cell present and every one meaning nothing, and a section
		// whose only table was skipped would declare nothing and pass.
		v.addv(blk[0].No, "a table under `## visual verification` whose header is `"+hdr+"` — the section holds one settings table (`Setting | Value`) and one commands table (`Command | Runs`), and nothing else in it is read")
		return
	}
	for _, row := range blk[1:] {
		cells := vtSplitCells(row.Text)
		if vvIsDelimiter(cells) {
			continue
		}
		// A SECOND HEADER INSIDE ONE BLOCK IS TWO TABLES THAT MERGED — the same
		// hazard check-workspace-isolation's flushBlock guards against, and
		// the same fix: abandon the block rather than read its later rows
		// against the wrong columns.
		if other := vvHeader(cells); other == vvSettingHeader || other == vvCommandHeader {
			v.addv(row.No, "a second table header (`"+other+"`) begins here, inside the table that started at line "+strconv.Itoa(blk[0].No)+" — two tables written with no blank line or prose between them are one block to this parser, so it read only the first header and stopped here rather than reading the rows below against the wrong columns; separate the two tables with a blank line")
			return
		}
		if len(cells) != 2 {
			v.addv(row.No, "a "+kind+" row with "+strconv.Itoa(len(cells))+" cell(s) where its header has 2 (expected 2) — the row does not line up with its header, so it is dropped rather than read against the wrong columns")
			continue
		}
		if kind == "setting" {
			v.settingRow(row.No, cells)
		} else {
			v.commandRow(row.No, cells)
		}
	}
}

var vvMockupFrame = regexp.MustCompile(`^scale=[1-9][0-9]* status=[0-9]+ border=[0-9]+$`)

// settingRow checks one row of the settings table. The `Setting` vocabulary
// is closed — an unrecognised name is reported, never silently skipped.
// `push to default branch` is deliberately not in it: task 14 removed it from
// the contract, so a row naming it is an unrecognised Setting like any other.
// `mockups` is optional, and this guard validates neither that the directory
// it names exists nor any `<spec>.mockups` sidecar; that is
// compose-mockup-frames' run-time question (design.md section 6). `mockup
// frame` is optional too, and its value's shape IS checked here: a geometry
// the compose step cannot parse is a usage error that would otherwise surface
// mid-run, after the stack is up and the captures are taken. The three fields
// are pinned in one order so a declared geometry is greppable, though the
// compose step itself accepts any.
func (v *vvValidation) settingRow(line int, cells []string) {
	disp, key := vtTrimCell(cells[0]), vtFoldCell(cells[0])
	switch key {
	case "ui paths", "screenshots", "regression checkout", "regression repo", "mockups", "mockup frame":
	default:
		v.addv(line, "Setting `"+disp+"` is not one of `ui paths`, `screenshots`, `regression checkout`, `regression repo`, `mockups` or `mockup frame` — the vocabulary is closed, so the row is dropped")
		return
	}
	if f, seen := v.set[key]; seen {
		v.addv(line, "a second `"+disp+"` row, the first being at line "+strconv.Itoa(f.line)+" — the row is dropped")
		return
	}
	val := vtTrimCell(cells[1])
	if key == "mockup frame" && val != "" && !vvMockupFrame.MatchString(val) {
		v.addv(line, "Setting `mockup frame` must be `scale=<int> status=<px> border=<px>`, got `"+val+"`")
		return
	}
	v.set[key] = vvField{line, val}
}

// commandRow checks one row of the commands table. The `Command` vocabulary
// is closed for the same reason. `fingerprint`, `start` and `specs` are
// optional and nothing requires any of them (KAN-395, KAN-462, KAN-761): a
// project with no served bundle declares no `fingerprint`, a project whose
// `## run` already starts the stack declares no `start`, and a project with
// no spec map declares no `specs`; each absence is silent, as `setup`'s is.
func (v *vvValidation) commandRow(line int, cells []string) {
	disp, key := vtTrimCell(cells[0]), vtFoldCell(cells[0])
	switch key {
	case "setup", "verify", "capture", "fingerprint", "start", "specs":
	default:
		v.addv(line, "Command `"+disp+"` is not one of `setup`, `verify`, `capture`, `fingerprint`, `start` or `specs` — the vocabulary is closed, so the row is dropped")
		return
	}
	if f, seen := v.cmd[key]; seen {
		v.addv(line, "a second `"+disp+"` row, the first being at line "+strconv.Itoa(f.line)+" — the row is dropped")
		return
	}
	v.cmd[key] = vvField{line, vtTrimCell(cells[1])}
}
