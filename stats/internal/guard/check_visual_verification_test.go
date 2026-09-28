package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The cases of the retired scripts/test-check-visual-verification.sh, one
// subtest per `ok:` label, so parity is a `--- PASS` count. Every subtest
// builds its own fixture tree under t.TempDir() and runs the guard against
// REAL files on disk. The origin-mismatch and no-origin cases use a REAL
// `git init` checkout with a REAL `origin` remote set via `git remote add`:
// the shape the guard's `git remote get-url origin` call gets from a real
// checkout is not the shape a hand-written fixture would produce, so only a
// real checkout proves the finding fires on the real boundary.

// vvCheck is one `ok:` label and the assertion behind it.
type vvCheck struct {
	label string
	ok    func(out string, rc int) bool
}

func vvExit(c string, want int) vvCheck {
	return vvCheck{fmt.Sprintf("%s: exit %d", c, want), func(_ string, rc int) bool { return rc == want }}
}

func vvNames(c, needle string) vvCheck {
	return vvCheck{fmt.Sprintf("%s: output names '%s'", c, needle), func(out string, _ int) bool { return strings.Contains(out, needle) }}
}

func vvOmits(c, needle string) vvCheck {
	return vvCheck{fmt.Sprintf("%s: output correctly omits '%s'", c, needle), func(out string, _ int) bool { return !strings.Contains(out, needle) }}
}

// vvCase is one harness case: run builds its fixture and runs the guard
// (stdout and stderr in one stream, as the harness's `2>&1`), and every
// check is a subtest of its own that runs it afresh.
type vvCase struct {
	run    func(t *testing.T) (string, int)
	checks []vvCheck
}

const (
	vvUIRow      = "| `ui paths` | `stats/web/src/**` |"
	vvShotsRow   = "| `screenshots` | `stats/web/tests/visual/baseline.spec.ts-snapshots` |"
	vvShotsDot   = "| `screenshots` | `.` |"
	vvShotsDir   = "| `screenshots` | `stats/web/tests/visual` |"
	vvRepoRow    = "| `regression repo` | `git@github.com:tweety53/gymie-playwright.git` |"
	vvVerifyRow  = "| `verify` | `npm run test:visual` |"
	vvCaptureRow = "| `capture` | `npx playwright test <spec>` |"
	vvRepoURL    = "git@github.com:tweety53/gymie-playwright.git"
)

// vvSection is a `## visual verification` section with one settings table
// and one commands table, separated by a blank line.
func vvSection(settings, commands []string) string {
	return "## visual verification\n\n| Setting | Value |\n|---------|-------|\n" +
		strings.Join(settings, "\n") + "\n\n| Command | Runs |\n|---------|------|\n" +
		strings.Join(commands, "\n")
}

var vvMinimal = vvSection([]string{vvUIRow, vvShotsRow}, []string{vvVerifyRow, vvCaptureRow})

func vvCheckoutRow(dir string) string { return "| `regression checkout` | `" + dir + "` |" }

// vvRoot is new_root: a fresh project root carrying .flow/.
func vvRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mkdir(t, root+"/.flow")
	return root
}

// vvWriteCfg is write_cfg: body plus the newline printf '%s\n' adds.
func vvWriteCfg(t *testing.T, root, body string) {
	t.Helper()
	writeFile(t, root+"/.flow/project.md", body+"\n")
}

// vvGitCheckout is make_git_checkout: a REAL `git init` checkout, with a REAL
// `origin` remote set via `git remote add` when a URL is given.
func vvGitCheckout(t *testing.T, dir, origin string) {
	t.Helper()
	mkdir(t, dir)
	gitRun(t, dir, "init", "-q")
	if origin != "" {
		gitRun(t, dir, "remote", "add", "origin", origin)
	}
}

// vvGuard runs the registered guard in-process.
func vvGuard(t *testing.T, env Env, args ...string) (string, int) {
	t.Helper()
	fn, ok := Registry["check-visual-verification"]
	if !ok {
		t.Fatal("check-visual-verification is not registered")
	}
	var out bytes.Buffer
	rc := fn(args, env, &out, &out)
	return out.String(), rc
}

func vvEnv(dir string) Env {
	return Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: dir}
}

// vvCfg is a case whose whole fixture is the configuration: cfg gets the
// fresh root and returns the file's body.
func vvCfg(cfg func(t *testing.T, root string) string) func(t *testing.T) (string, int) {
	return func(t *testing.T) (string, int) {
		root := vvRoot(t)
		vvWriteCfg(t, root, cfg(t, root))
		return vvGuard(t, vvEnv(root), root)
	}
}

func vvText(body string) func(t *testing.T) (string, int) {
	return vvCfg(func(*testing.T, string) string { return body })
}

// vvWithCheckout is a case whose `regression checkout` is a real checkout
// at <root>/checkout with origin (none when ""), declared beside
// `regression repo` and extra settings rows.
func vvWithCheckout(origin string, extra ...string) func(t *testing.T) (string, int) {
	return vvCfg(func(t *testing.T, root string) string {
		dir := root + "/checkout"
		vvGitCheckout(t, dir, origin)
		settings := append([]string{vvUIRow, vvShotsDot, vvCheckoutRow(dir), vvRepoRow}, extra...)
		return vvSection(settings, []string{vvVerifyRow, vvCaptureRow})
	})
}

func vvCases() map[string]vvCase {
	cases := map[string]vvCase{
		// Case 1: no .flow/project.md at all — exit 0, clean.
		"case 1": {func(t *testing.T) (string, int) {
			root := vvRoot(t)
			return vvGuard(t, vvEnv(root), root)
		}, []vvCheck{vvExit("case 1", 0), vvNames("case 1", "no .flow/project.md")}},

		// Case 2: .flow/project.md exists but declares no `## visual
		// verification` section — exit 0, clean.
		"case 2": {vvText("# Project\n\n## run\n\necho hi\n"),
			[]vvCheck{vvExit("case 2", 0), vvNames("case 2", "declares no")}},

		// Case 3: a minimal, fully valid section — exit 0.
		"case 3": {vvText(vvMinimal), []vvCheck{vvExit("case 3", 0), vvNames("case 3", "VISUAL-OK")}},

		// Case 4: `ui paths` absent — exit 1, names the row.
		"case 4": {vvText(vvSection([]string{vvShotsRow}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 4", 1), vvNames("case 4", "ui paths")}},

		// Case 5: `screenshots` present but empty — exit 1, names the row.
		"case 5": {vvText(vvSection([]string{vvUIRow, "| `screenshots` | |"}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 5", 1), vvNames("case 5", "screenshots")}},

		// Case 6: `verify` absent — exit 1, names the row.
		"case 6": {vvText(vvSection([]string{vvUIRow, vvShotsRow}, []string{vvCaptureRow})),
			[]vvCheck{vvExit("case 6", 1), vvNames("case 6", "verify")}},

		// Case 7: `capture` present but empty — exit 1, names the row.
		"case 7": {vvText(vvSection([]string{vvUIRow, vvShotsRow}, []string{vvVerifyRow, "| `capture` | |"})),
			[]vvCheck{vvExit("case 7", 1), vvNames("case 7", "capture")}},

		// Case 8: `regression repo` present with no `regression checkout` —
		// exit 1.
		"case 8": {vvText(vvSection([]string{vvUIRow, vvShotsRow, vvRepoRow}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 8", 1), vvNames("case 8", "regression repo"), vvNames("case 8", "regression checkout")}},

		// Case 9: `regression checkout` present with no `regression repo` —
		// exit 1 (the reverse pairing).
		"case 9": {vvCfg(func(t *testing.T, root string) string {
			dir := root + "/checkout"
			vvGitCheckout(t, dir, "")
			return vvSection([]string{vvUIRow, vvShotsDot, vvCheckoutRow(dir)}, []string{vvVerifyRow, vvCaptureRow})
		}), []vvCheck{vvExit("case 9", 1), vvNames("case 9", "regression checkout")}},

		// Case 10: `regression checkout` names a path that does not exist —
		// exit 1.
		"case 10": {vvCfg(func(t *testing.T, root string) string {
			return vvSection([]string{vvUIRow, vvShotsDot, vvCheckoutRow(root + "/does-not-exist"), vvRepoRow},
				[]string{vvVerifyRow, vvCaptureRow})
		}), []vvCheck{vvExit("case 10", 1), vvNames("case 10", "not an existing directory")}},

		// Case 11: `regression checkout` names an existing directory that is
		// not a git checkout — exit 1.
		"case 11": {vvCfg(func(t *testing.T, root string) string {
			mkdir(t, root+"/plain-dir")
			return vvSection([]string{vvUIRow, vvShotsDot, vvCheckoutRow(root + "/plain-dir"), vvRepoRow},
				[]string{vvVerifyRow, vvCaptureRow})
		}), []vvCheck{vvExit("case 11", 1), vvNames("case 11", "not a git checkout")}},

		// Case 12: `push to default branch` is no longer part of the contract
		// — task 14 dropped it once a panel slot demonstrated that both sides
		// of its origin-equality check live in this same pull-request-editable
		// file. A row naming it is an unrecognised Setting, reported rather
		// than silently honoured, even when the checkout it names is real and
		// its origin matches `regression repo`.
		"case 12": {vvWithCheckout(vvRepoURL, "| `push to default branch` | `allowed` |"),
			[]vvCheck{vvExit("case 12", 1), vvNames("case 12", "push to default branch"), vvNames("case 12", "vocabulary is closed")}},

		// Case 13 (REPRODUCE, DON'T READ): `regression repo` is an identity
		// assertion, checked whenever `regression checkout` and `regression
		// repo` are both declared. A REAL checkout's REAL origin does not
		// equal the declared `regression repo` — exit 1, reports BOTH URLs.
		"case 13": {vvWithCheckout("git@github.com:someone-else/decoy.git"),
			[]vvCheck{vvExit("case 13", 1), vvNames("case 13", "git@github.com:someone-else/decoy.git"), vvNames("case 13", vvRepoURL)}},

		// Case 14 (REPRODUCE, DON'T READ): a REAL checkout's REAL origin DOES
		// equal the declared `regression repo` — exit 0, clean.
		"case 14": {vvWithCheckout(vvRepoURL), []vvCheck{vvExit("case 14", 0), vvNames("case 14", "VISUAL-OK")}},

		// Case 15 (REPRODUCE, DON'T READ): both declared and the REAL checkout
		// carries no `origin` remote — exit 1, a finding rather than a silent
		// pass.
		"case 15": {vvWithCheckout(""), []vvCheck{vvExit("case 15", 1), vvNames("case 15", "origin")}},

		// Case 16: the project root does not exist — exit 2.
		"case 16": {func(t *testing.T) (string, int) {
			return vvGuard(t, vvEnv(t.TempDir()), fmt.Sprintf("/nonexistent/%d", os.Getpid()))
		}, []vvCheck{vvExit("case 16", 2)}},

		// Case 17: `.flow/project.md` is a directory, not a regular file —
		// exit 2.
		"case 17": {func(t *testing.T) (string, int) {
			root := vvRoot(t)
			mkdir(t, root+"/.flow/project.md")
			return vvGuard(t, vvEnv(root), root)
		}, []vvCheck{vvExit("case 17", 2)}},

		// Case 18: the file declares `## visual verification` twice — exit 1,
		// a single violation naming the duplication, neither section
		// validated.
		"case 18": {vvText(vvMinimal + "\n\n" + vvMinimal), []vvCheck{vvExit("case 18", 1), vvNames("case 18", "2")}},

		// Case 19: a duplicate `ui paths` row inside the same settings table —
		// exit 1.
		"case 19": {vvText(vvSection([]string{vvUIRow, "| `ui paths` | `stats/web/other/**` |", vvShotsDot},
			[]string{vvVerifyRow, vvCaptureRow})), []vvCheck{vvExit("case 19", 1), vvNames("case 19", "ui paths")}},

		// Case 20: an unrecognised table header under the section — exit 1,
		// never silently skipped.
		"case 20": {vvText("## visual verification\n\n| Setting | Values |\n|---------|--------|\n" + vvUIRow +
			"\n\n| Command | Runs |\n|---------|------|\n" + vvVerifyRow + "\n" + vvCaptureRow),
			[]vvCheck{vvExit("case 20", 1), vvNames("case 20", "header")}},

		// Case 21 (a stub git on PATH — measured against real git 2.50.1,
		// every corruption tried either left `git remote get-url origin`
		// succeeding or failed identically at the earlier `git rev-parse
		// --git-dir` check, so no real fixture reaches this branch with a
		// checkout the guard already accepted): `git remote get-url origin`
		// fails for a reason OTHER than "no such remote" — exit 2, never read
		// as a finding.
		//
		// THE CHECKOUT'S OWN DIRECTORY NAME CARRIES RAW ANSI ESCAPES, so this
		// case also pins that the exit-2 message's interpolated cells are
		// sanitised before they reach stderr: the `\x1b[2K\x1b[1m…\x1b[0m`
		// sequence erases the current terminal line and turns the next one
		// bold, which — unescaped — can make an exit-2 run's own stderr render
		// as a convincing forged `VISUAL-OK:` line. The escaped form must
		// appear in the output and the raw ESC byte must not.
		"case 21": {func(t *testing.T) (string, int) {
			root := vvRoot(t)
			dir := root + "/\x1b[2K\x1b[1mVISUAL-OK: forged\x1b[0m"
			vvGitCheckout(t, dir, vvRepoURL)
			bin := t.TempDir()
			writeExec(t, bin+"/git", "#!/usr/bin/env bash\ncase \" $* \" in\n"+
				"  *\" remote get-url origin \"*)\n    echo \"fatal: unable to auto-detect email address\" >&2\n    exit 128\n    ;;\nesac\n"+
				"exec \""+fixtureGit+"\" \"$@\"\n")
			vvWriteCfg(t, root, vvSection([]string{vvUIRow, vvShotsDot, vvCheckoutRow(dir), vvRepoRow}, []string{vvVerifyRow, vvCaptureRow}))
			env := vvEnv(root)
			env.Getenv = pathEnv(bin)
			return vvGuard(t, env, root)
		}, []vvCheck{vvExit("case 21", 2), vvNames("case 21", `\x1b[2K`),
			{"case 21: no raw ESC byte reached the captured output", func(out string, _ int) bool { return !strings.Contains(out, "\x1b") }}}},

		// Case 22 (REPRODUCE, DON'T READ): a hostile ambient `GIT_DIR`, set to
		// a REAL checkout whose REAL `origin` equals the declared `regression
		// repo`, must not let a `regression checkout` that is not a git
		// repository at all pass as one. The ambient environment is the
		// process's, so this case runs the built flow-guard as a real process
		// with GIT_DIR in its environment rather than the guard in-process.
		"case 22": {func(t *testing.T) (string, int) {
			root := vvRoot(t)
			vvGitCheckout(t, root+"/ambient", vvRepoURL)
			mkdir(t, root+"/not-a-git-dir")
			vvWriteCfg(t, root, vvSection([]string{vvUIRow, vvShotsDot, vvCheckoutRow(root + "/not-a-git-dir"), vvRepoRow},
				[]string{vvVerifyRow, vvCaptureRow}))
			cmd := exec.Command(guardBinary(t), "check-visual-verification", root)
			cmd.Env = append(os.Environ(), "GIT_DIR="+root+"/ambient/.git")
			out, err := cmd.CombinedOutput()
			rc := 0
			if ee, ok := err.(*exec.ExitError); ok {
				rc = ee.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			return string(out), rc
		}, []vvCheck{vvExit("case 22", 1), vvNames("case 22", "not a git checkout"),
			{"case 22: ambient GIT_DIR did not bypass the checkout-validity check", func(out string, _ int) bool { return !strings.Contains(out, "VISUAL-OK") }}}},

		// Case 23: a `###` subheading inside the section does not end it — a
		// second settings-table block declared below it, and the commands
		// table below that, are still read. exit 0, clean.
		"case 23": {vvText("## visual verification\n\n| Setting | Value |\n|---------|-------|\n" + vvUIRow +
			"\n\n### Notes\n\n| Setting | Value |\n|---------|-------|\n" + vvShotsRow +
			"\n\n| Command | Runs |\n|---------|------|\n" + vvVerifyRow + "\n" + vvCaptureRow),
			[]vvCheck{vvExit("case 23", 0), vvNames("case 23", "VISUAL-OK")}},

		// Case 24 (REPRODUCE, DON'T READ): the file begins with a UTF-8 BOM,
		// with `## visual verification` as its first line. A BOM must not make
		// a present section read as absent.
		"case 24": {vvText("\xef\xbb\xbf" + vvMinimal), []vvCheck{vvExit("case 24", 0),
			vvNames("case 24", "the `## visual verification` section validated"), vvOmits("case 24", "declares no")}},

		// Case 25: `push to default branch: allowed` with NEITHER `regression
		// checkout` nor `regression repo` declared — an unrecognised Setting
		// on its own, before any pairing logic is reached.
		"case 25": {vvText(vvSection([]string{vvUIRow, vvShotsDot, "| `push to default branch` | `allowed` |"},
			[]string{vvVerifyRow, vvCaptureRow})), []vvCheck{vvExit("case 25", 1),
			vvNames("case 25", "push to default branch"), vvNames("case 25", "vocabulary is closed")}},

		// Case 26 (KAN-359, validator-agrees-with-trigger): `ui paths` is two
		// empty backtick pairs joined by a comma — non-empty as a cell, zero
		// usable globs after check-visual-trigger's split-then-strip parse,
		// the condition under which that guard exits 2. The two guards must
		// not disagree: exit 1, naming `ui paths`.
		"case 26": {vvText(vvSection([]string{"| `ui paths` | ``, `` |", vvShotsDot}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 26", 1), vvNames("case 26", "VISUAL-INVALID"), vvNames("case 26", "ui paths"),
				vvNames("case 26", "no usable glob"), vvOmits("case 26", "VISUAL-OK")}},

		// Case 27 (KAN-359 task 8): case 26's empty elements sit at the list
		// boundaries, where the whole-cell trim already removes them before
		// trimGlobElement runs. This element is INTERIOR and whitespace-and-
		// backtick only, so it is non-vacuous by construction: the real trim
		// strips it to nothing (VISUAL-INVALID), an identity trim leaves
		// " ` ` " and counts it usable (VISUAL-OK).
		"case 27": {vvText(vvSection([]string{"| `ui paths` | , ` ` , |", vvShotsDot}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 27", 1), vvNames("case 27", "VISUAL-INVALID"), vvNames("case 27", "ui paths"),
				vvNames("case 27", "no usable glob"), vvOmits("case 27", "VISUAL-OK")}},

		// Case 28 (KAN-395): `fingerprint` is in the closed Command vocabulary
		// and optional.
		"case 28": {vvText(vvSection([]string{vvUIRow, vvShotsDir}, []string{vvVerifyRow, vvCaptureRow,
			"| `fingerprint` | `curl -sf http://127.0.0.1:4174/ \\| cmp -s - internal/web/dist/index.html` |"})),
			[]vvCheck{vvExit("case 28", 0), vvNames("case 28", "VISUAL-OK"), vvOmits("case 28", "vocabulary is closed"), vvOmits("case 28", "fingerprint")}},

		// Case 29 (KAN-395 panel round 0, mutation finding): an unrecognised
		// Command row's message names the closed vocabulary word by word, not
		// just that it is closed.
		"case 29": {vvText(vvSection([]string{vvUIRow, vvShotsDot}, []string{"| `verify` | `true` |", "| `capture` | `true` |",
			"| `deploy` | `npm run deploy` |"})), []vvCheck{vvExit("case 29", 1), vvNames("case 29", "vocabulary is closed"),
			vvNames("case 29", "`setup`"), vvNames("case 29", "`verify`"), vvNames("case 29", "`capture`"), vvNames("case 29", "`fingerprint`")}},

		// Case 30 (KAN-462 task 4): a `start` row is accepted.
		"case 30": {vvText(vvSection([]string{vvUIRow, vvShotsDir}, []string{vvVerifyRow, vvCaptureRow, "| `start` | `./gradlew devStart` |"})),
			[]vvCheck{vvExit("case 30", 0), vvNames("case 30", "VISUAL-OK"), vvOmits("case 30", "vocabulary is closed"), vvOmits("case 30", "start")}},

		// Case 31 (KAN-462 task 4): a `start` row is optional.
		"case 31": {vvText(vvSection([]string{vvUIRow, vvShotsDir}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 31", 0), vvNames("case 31", "VISUAL-OK"), vvOmits("case 31", "start")}},

		// Case 32 (KAN-449 task 3): a `mockups` row is accepted.
		"case 32": {vvText(vvSection([]string{vvUIRow, vvShotsDir, "| `mockups` | `docs/design/screens` |"}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 32", 0), vvNames("case 32", "VISUAL-OK"), vvOmits("case 32", "vocabulary is closed")}},

		// Case 33 (KAN-449 task 3): the unrecognised-Setting message names the
		// closed vocabulary word by word; case 12's fixture.
		"case 33": {vvWithCheckout(vvRepoURL, "| `push to default branch` | `allowed` |"),
			[]vvCheck{vvExit("case 33", 1), vvNames("case 33", "vocabulary is closed"), vvNames("case 33", "`ui paths`"),
				vvNames("case 33", "`screenshots`"), vvNames("case 33", "`regression checkout`"),
				vvNames("case 33", "`regression repo`"), vvNames("case 33", "`mockups`")}},

		// Case 34 (KAN-515 task 2): a well-formed `mockup frame` row — exit 0.
		"case 34": {vvText(vvSection([]string{vvUIRow, vvShotsRow, "| `mockup frame` | `scale=2 status=26 border=1` |"},
			[]string{vvVerifyRow, vvCaptureRow})), []vvCheck{vvExit("case 34", 0), vvNames("case 34", "VISUAL-OK")}},

		// Case 36 (KAN-515 task 2): `mockup frame` declared twice — the
		// existing duplicate-setting violation.
		"case 36": {vvText(vvSection([]string{vvUIRow, vvShotsRow, "| `mockup frame` | `scale=2 status=26 border=1` |",
			"| `mockup frame` | `scale=3 status=20 border=1` |"}, []string{vvVerifyRow, vvCaptureRow})),
			[]vvCheck{vvExit("case 36", 1), vvNames("case 36", "a second `mockup frame` row")}},

		// Case 37 (KAN-761 task 15): a `specs` row is accepted and optional.
		"case 37": {vvText(vvSection([]string{vvUIRow, vvShotsDir}, []string{vvVerifyRow, vvCaptureRow,
			"| `specs` | `node scripts/specs-for-diff.mjs <frontend-root> <merge-base>` |"})),
			[]vvCheck{vvExit("case 37", 0), vvNames("case 37", "VISUAL-OK"), vvOmits("case 37", "vocabulary is closed")}},
	}

	// Case 35 (KAN-515 task 2): a `mockup frame` value that is not
	// `scale=<int> status=<px> border=<px>` — exit 1, reported here rather
	// than reaching the compose step as a usage error mid-run.
	for _, bad := range []string{"2x", "scale=2 status=26", "status=26 border=1 scale=2", "scale=0 status=26 border=1"} {
		c := "case 35 (" + bad + ")"
		cases[c] = vvCase{vvText(vvSection([]string{vvUIRow, vvShotsRow, "| `mockup frame` | `" + bad + "` |"},
			[]string{vvVerifyRow, vvCaptureRow})), []vvCheck{vvExit(c, 1), vvNames(c, "mockup frame"),
			vvNames(c, "scale=<int> status=<px> border=<px>")}}
	}
	return cases
}

func TestCheckVisualVerification(t *testing.T) {
	t.Parallel()
	for _, c := range vvCases() {
		for _, chk := range c.checks {
			t.Run(chk.label, func(t *testing.T) {
				t.Parallel()
				out, rc := c.run(t)
				if !chk.ok(out, rc) {
					t.Errorf("rc=%d out=%q", rc, out)
				}
			})
		}
	}

	// Beyond the harness: the output's exact bytes. The root is printed as
	// given — a relative one relative, resolved against the working
	// directory only to be read — and every violation line is
	// `<cfg>:<line>: <message>` through the sanitize twin, in the script's
	// order, before one verdict line.
	t.Run("a relative root is printed as given", func(t *testing.T) {
		t.Parallel()
		root := vvRoot(t)
		vvWriteCfg(t, root, vvMinimal)
		out, rc := vvGuard(t, vvEnv(root), ".")
		want := "VISUAL-OK: ./.flow/project.md — the `## visual verification` section validated\n"
		if rc != 0 || out != want {
			t.Errorf("rc=%d out=%q, want 0 %q", rc, out, want)
		}
	})
	t.Run("violation lines are sanitized, in order, before the verdict", func(t *testing.T) {
		t.Parallel()
		root := vvRoot(t)
		vvWriteCfg(t, root, "## visual verification\r\n\r\n| Setting | Value |\r\n|---|---|\r\n"+
			"| `ui paths` | `a\\|b` |\r\n| `bogus\x1b[1m` | x |\r\n| `ui paths` | c |\r\n| `regression repo` | r |\r\n| `x` |\r\n"+
			"| Command | Runs |\r\n| `verify` | v |\r\n")
		out, rc := vvGuard(t, vvEnv(root), ".")
		c := "./.flow/project.md"
		want := c + ":6: Setting `bogus\\x1b[1m` is not one of `ui paths`, `screenshots`, `regression checkout`, `regression repo`, `mockups` or `mockup frame` — the vocabulary is closed, so the row is dropped\n" +
			c + ":7: a second `ui paths` row, the first being at line 5 — the row is dropped\n" +
			c + ":9: a setting row with 1 cell(s) where its header has 2 (expected 2) — the row does not line up with its header, so it is dropped rather than read against the wrong columns\n" +
			c + ":10: a second table header (`command|runs`) begins here, inside the table that started at line 3 — two tables written with no blank line or prose between them are one block to this parser, so it read only the first header and stopped here rather than reading the rows below against the wrong columns; separate the two tables with a blank line\n" +
			c + ":1: `screenshots` is absent — this setting is required\n" +
			c + ":1: `verify` is absent — this setting is required\n" +
			c + ":1: `capture` is absent — this setting is required\n" +
			c + ":8: `regression repo` is present with no `regression checkout` — a repo with nothing to check its origin against declares nothing checkable\n" +
			"VISUAL-INVALID: " + c + " — 8 violation(s) in the `## visual verification` section\n"
		if rc != 1 || out != want {
			t.Errorf("rc=%d out:\n%s\nwant:\n%s", rc, out, want)
		}
	})
}
