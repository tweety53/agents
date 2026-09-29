package guard

// The retired scripts/test-check-workspace-isolation.sh, ported case for
// case: each subtest is named after that harness's `ok:` label. Each case
// builds a throwaway project root under t.TempDir(), writes a
// `.flow/project.md` into it, and asserts the guard's violation lines, its
// verdict and its exit status. Never reads or writes the real repository
// tree except where a case names it (1c, 1d, 15a), and never reads another
// real project on this machine.
//
// READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
// contract in skills/flow-contracts/project-configuration.md and
// skills/flow-contracts/project-configuration-isolation.md — the
// `## workspace isolation` row of the key table, the four `In a workspace`
// cell forms, "What a `url` row may reference, and what it may not", and the
// two bullets under "An isolation row resolves under the same rules this file
// applies to everything else it consumes". Never assert against observed
// output. check_plan_provenance_test.go's header records the retired
// plan-provenance harness encoding the guard's own defects as its
// specification more than once, which then made each defect look verified.
//
// WHY EVERY CASE ASSERTS ON THE REASON TEXT. The contract's remedy for a bad
// row is that the row is "reported by name and dropped", so a guard that
// answered only "this file is invalid" would satisfy every exit-status
// assertion while leaving the operator nothing to act on. The needles below
// are short and behavioural — evidence that the named row, cell and rule
// reached the report, not a transcript of its prose.
//
// WHY THIS DUPLICATES the cleanup guard's test helpers (ccNew, ccVerdict and
// their siblings in check_cleanup_complete_test.go) instead of sharing them.
// The two guards read the same section of the same file and are deliberately
// kept independent: that guard asks whether the removal happened, this one
// asks whether the declaration is well formed, and shared helpers would
// couple their suites so that a change to one guard's contract could only be
// made by editing code the other one also runs. The duplication is the
// cheaper of the two, and it is recorded here rather than left unexplained.

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

const wiRepoRoot = "../../.."

type wiResult struct {
	rc       int
	out, err string // out is stdout only: a refusal must leave it empty
}

// wiRun runs the guard in-process with vars as its whole environment. The
// two streams are captured apart, because the contract distinguishes them: a
// refusal puts its message on stderr and must leave stdout empty.
func wiRun(t *testing.T, vars map[string]string, args ...string) wiResult {
	t.Helper()
	fn := Registry["check-workspace-isolation"]
	if fn == nil {
		t.Fatal("check-workspace-isolation is not registered")
	}
	var o, e bytes.Buffer
	rc := fn(args, Env{Dir: "/", Getenv: func(k string) string { return vars[k] }}, &o, &e)
	return wiResult{rc, o.String(), e.String()}
}

// wiQ writes a fixture's backticks as `~`, which no fixture otherwise
// carries, so the rows below can be Go raw strings.
func wiQ(s string) string { return strings.ReplaceAll(s, "~", "`") }

// wiConfig is write_config: the whole of root/.flow/project.md, as
// printf '%s\n' writes it.
func wiConfig(t *testing.T, root, body string) {
	t.Helper()
	writeFile(t, root+"/.flow/project.md", body+"\n")
}

// The rows a correct declaration is built from, kept in one place so a case
// that breaks one row is visibly a case about THAT row. Every cell form the
// contract names appears here, so this set doubles as the positive case: two
// removable resources, the probed index, a port, and the three `url` shapes —
// one built from a `<value:…>` port reference, one from a `<value:…>` bucket
// reference, and one carrying no token at all.
var wiValidRes = wiQ(`| ~database~ | ~DB_URL~ | ~jdbc:postgresql://localhost:5432/appdb~ | ~jdbc:postgresql://localhost:5432/appdb_<id_underscored>~ |
| ~bucket~ | ~MEDIA_BUCKET~ | ~appdb-media~ | ~appdb-media-<id>~ |
| ~cache index~ | ~CACHE_INDEX~ | ~0~ | ~probed~ |
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| ~url~ | ~MEDIA_BASE_URL~ | ~http://localhost:9000/appdb-media~ | ~http://localhost:9000/<value:MEDIA_BUCKET>~ |
| ~url~ | ~WEB_URL~ | ~http://localhost:8080~ | ~http://localhost:<value:API_PORT>~ |
| ~url~ | ~STATIC_URL~ | ~http://localhost:9000~ | ~http://localhost:9000~ |`)

var wiValidCmd = wiQ(`| ~create~ | ~./scripts/workspace create~ |
| ~remove~ | ~./scripts/workspace remove <id>~ |
| ~survivors~ | ~./scripts/workspace survivors <id>~ |`)

// wiIsoBody is write_iso's file: a `## workspace isolation` section carrying
// the two tables with those rows, and prose around them, because prose
// beside the tables is permitted and is never read. A section written without
// any is not the shape a project actually writes, and a parser that only ever
// met bare tables would not be tested against the one it will meet.
func wiIsoBody(res, cmd string) string {
	return "# fixture project configuration\n\n## apps\n\nNothing here is read by this guard.\n\n" +
		"## workspace isolation\n\nProse above the tables, which the resolver ignores.\n\n" +
		"| Resource | Variable | Default | In a workspace |\n|----------|----------|---------|----------------|\n" +
		res + "\n\nProse between the tables, likewise ignored.\n\n" +
		"| Command | Runs |\n|---------|------|\n" + cmd +
		"\n\nProse below the tables, saying what this project has deliberately not isolated.\n\n" +
		"## lint\n\nAlso not read by this guard."
}

// wiIso is new_project plus write_iso, then the guard run on that root.
func wiIso(t *testing.T, res, cmd string) wiResult {
	t.Helper()
	root := t.TempDir()
	wiConfig(t, root, wiIsoBody(wiQ(res), wiQ(cmd)))
	return wiRun(t, nil, root)
}

// wiFile is new_project plus write_config, then the guard run on that root.
func wiFile(t *testing.T, body string) wiResult {
	t.Helper()
	root := t.TempDir()
	wiConfig(t, root, wiQ(body))
	return wiRun(t, nil, root)
}

// wiRaw writes the configuration byte for byte, with no newline appended.
func wiRaw(t *testing.T, body string) wiResult {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root+"/.flow/project.md", body)
	return wiRun(t, nil, root)
}

func wiCheck(t *testing.T, label string, ok bool, format string, a ...any) {
	t.Helper()
	t.Run(label, func(t *testing.T) {
		if !ok {
			t.Fatalf(format, a...)
		}
	})
}

// wiLines is stdout as $(…) captured it — trailing newlines stripped — split
// into lines.
func wiLines(out string) []string {
	return strings.Split(strings.TrimRight(out, "\n"), "\n")
}

func wiVerdict(t *testing.T, r wiResult, rc int, prefix, label string) {
	t.Helper()
	l := wiLines(r.out)
	last := l[len(l)-1]
	switch {
	case r.rc != rc:
		wiCheck(t, label, false, "expected exit %d, got rc=%d out=%s err=%s", rc, r.rc, r.out, r.err)
	default:
		wiCheck(t, label, strings.HasPrefix(last, prefix), "expected an %s verdict, got: %s", prefix, r.out)
	}
}

// wiOK is assert_ok: exit 0 and an ISOLATION-OK verdict on the last line.
func wiOK(t *testing.T, r wiResult, label string) {
	t.Helper()
	wiVerdict(t, r, 0, "ISOLATION-OK:", label)
}

// wiSilent is assert_silent: exit 0, and the verdict line is the ONLY line. A
// project that declares no section is the overwhelmingly common case, and a
// guard that printed a note per project would make every lint run noisier
// for every repository that has nothing to declare.
func wiSilent(t *testing.T, r wiResult, label string) {
	t.Helper()
	wiOK(t, r, label)
	n := len(wiLines(r.out))
	wiCheck(t, label+": says exactly one line", n == 1, "expected exactly one stdout line, got %d: %s", n, r.out)
}

// wiInvalid is assert_invalid: exit 1 and an ISOLATION-INVALID verdict. Exit
// 1 is "violations found", kept distinct from exit 2, "cannot answer at all".
func wiInvalid(t *testing.T, r wiResult, label string) {
	t.Helper()
	wiVerdict(t, r, 1, "ISOLATION-INVALID:", label)
}

// wiReports is assert_reports: the report names this row, cell or rule.
func wiReports(t *testing.T, r wiResult, needle, label string) {
	t.Helper()
	wiCheck(t, label, strings.Contains(r.out, needle), "the report does not name %q: %s", needle, r.out)
}

// wiNotReported is assert_not_reported. A guard that reported a row which is
// in fact well formed sends the operator to fix something that is not
// broken, and every exit-status assertion still passes while it does.
func wiNotReported(t *testing.T, r wiResult, needle, label string) {
	t.Helper()
	wiCheck(t, label, !strings.Contains(r.out, needle), "the report names a row that is well formed (%q): %s", needle, r.out)
}

// wiRefuses is assert_refuses: exit 2, nothing on stdout, and the guard's
// own name on stderr. The needle carries the colon deliberately: without it
// a shell's own "…/check-workspace-isolation.sh: No such file or directory"
// satisfies the case, so it passes while the guard does not exist.
func wiRefuses(t *testing.T, r wiResult, label string) {
	t.Helper()
	wiCheck(t, label+": exits 2", r.rc == 2, "expected exit 2, got rc=%d out=%s", r.rc, r.out)
	wiCheck(t, label+": writes nothing to stdout", r.out == "", "emitted a verdict line: %s", r.out)
	wiCheck(t, label+": names the failure on stderr", strings.Contains(r.err, "check-workspace-isolation: "),
		"no named message on stderr: %s", r.err)
}

// wiRepoAbs is the checkout this test runs in.
func wiRepoAbs(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(wiRepoRoot)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

// wiSkillTree is the scratch tree of cases 16 and 17: the script reachable at
// two depths, its real home (root/scripts/) and a skills/flow/scripts/
// symlink, mirroring how setup.sh's install carries it. Returns the root and
// the symlinked path.
func wiSkillTree(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	src, err := os.ReadFile(wiRepoRoot + "/scripts/check-workspace-isolation.sh")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root+"/scripts/check-workspace-isolation.sh", string(src))
	// The real lib/, so the shim's lib/flow-guard.sh finds stats/ to build
	// flow-guard from (it resolves its own directory physically).
	if err := os.Symlink(wiRepoAbs(t)+"/scripts/lib", root+"/scripts/lib"); err != nil {
		t.Fatal(err)
	}
	mkdir(t, root+"/skills/flow/scripts")
	link := root + "/skills/flow/scripts/check-workspace-isolation.sh"
	for l, target := range map[string]string{link: "../../../scripts/check-workspace-isolation.sh",
		root + "/skills/flow/scripts/lib": "../../../scripts/lib"} {
		if err := os.Symlink(target, l); err != nil {
			t.Fatal(err)
		}
	}
	return root, link
}

func TestCheckWorkspaceIsolation(t *testing.T) {
	t.Parallel()
	t.Run("port: the output matches the bash at c5379c0a byte for byte", wiParity)

	// 1. A project that declares nothing.

	// 1a. No .flow/project.md at all. The file is optional and its absence is
	//     a supported, ordinary case — never an error.
	wiSilent(t, wiRun(t, nil, t.TempDir()), "a project with no .flow/project.md passes silently")

	// 1b. A project.md with no `## workspace isolation` section. This is the
	//     overwhelmingly common case across projects flow is installed into.
	wiSilent(t, wiFile(t, "# fixture\n\n## test\n\n~~~bash\nscripts/test-setup.sh\n~~~\n"),
		"a project declaring no isolation section passes silently")

	// 1c. The agents repository itself, read from its real path rather than
	//     from a fixture. It declares its own `## workspace isolation`
	//     section, and this guard is on its own lint list — a guard that
	//     failed on the repository it ships in would be reverted rather than
	//     fixed. wiSilent checks a single-line verdict, not empty output, so a
	//     well-formed section still passes: it reports ISOLATION-OK on one line.
	repo := wiRepoAbs(t)
	wiSilent(t, wiRun(t, nil, repo), "the agents repository's own section validates cleanly")

	// 1d. Invoked with no argument at all, which is how the `## lint` list
	//     runs it. It resolves its own repository root — from FLOW_GUARD_SELF,
	//     the path the shim was invoked by — so a lint step is one word long.
	wiSilent(t, wiRun(t, map[string]string{"FLOW_GUARD_SELF": repo + "/scripts/check-workspace-isolation.sh"}),
		"invoked bare, the guard checks its own repository")

	// 2. Inputs it cannot answer about. Never fail open: a file it cannot
	//    read is not a project that declares nothing.

	// 2a. A path that is not a directory. Reading "declares nothing" out of a
	//     mistyped path would report every project valid.
	wiRefuses(t, wiRun(t, nil, t.TempDir()+"/no-such-project-root"), "a path that is not a directory")

	// 2b. A project.md that exists but cannot be read. Absence read out of a
	//     path that was never readable is the false pass this guard exists to
	//     prevent. THE GATE IS READABILITY OF THE FILE ITSELF, not the user
	//     id: a privileged capability, a filesystem mounted with permissions
	//     turned off, or an ACL that outranks the mode bits all leave a
	//     non-root user reading a mode-000 file — measured once in six runs
	//     on this machine before the harness's gate changed.
	unreadable := t.TempDir()
	wiConfig(t, unreadable, "# fixture")
	if err := os.Chmod(unreadable+"/.flow/project.md", 0); err != nil {
		t.Fatal(err)
	}
	if syscall.Access(unreadable+"/.flow/project.md", 4) == nil {
		t.Run("an unreadable project.md refuses", func(t *testing.T) { t.Skip("this user can read a mode-000 file") })
	} else {
		wiRefuses(t, wiRun(t, nil, unreadable), "an unreadable project.md")
	}

	// 2c. A `.flow/project.md` that is a DIRECTORY. Readability is true of a
	//     directory, and the bash guard's `grep -qiE` then failed with "Is a
	//     directory" while its exit status was indistinguishable from "no
	//     match" — so this shipped as `ISOLATION-OK` and exit 0, the guard's
	//     own NEVER FAIL OPEN invariant broken in the one script written to
	//     hold it. Anyone able to land a pull request can create the symlink.
	dirCfg := t.TempDir()
	mkdir(t, dirCfg+"/.flow/project.md")
	wiRefuses(t, wiRun(t, nil, dirCfg), "a project.md that is a directory")

	// 2d. A `.flow/project.md` that is a symlink to nowhere. A test that
	//     follows the link answers "absent", and the guard would report the
	//     ordinary, silent "no .flow/project.md" verdict — an absence
	//     manufactured out of a path someone deliberately pointed at nothing.
	dangling := t.TempDir()
	mkdir(t, dangling+"/.flow")
	if err := os.Symlink(t.TempDir()+"/no-such-configuration-target", dangling+"/.flow/project.md"); err != nil {
		t.Fatal(err)
	}
	wiRefuses(t, wiRun(t, nil, dangling), "a project.md that is a dangling symlink")

	// 2e. A `.flow/project.md` that is a symlink to a REAL file is still read:
	//     the refusals above exclude a file type rather than an indirection.
	linked := t.TempDir()
	mkdir(t, linked+"/.flow")
	target := t.TempDir() + "/linked-project.md"
	writeFile(t, target, "# fixture\n")
	if err := os.Symlink(target, linked+"/.flow/project.md"); err != nil {
		t.Fatal(err)
	}
	wiSilent(t, wiRun(t, nil, linked), "a project.md that is a symlink to a real file is read")

	// 3a. Every cell form the contract names, in one section.
	wiOK(t, wiIso(t, wiValidRes, wiValidCmd), "a declaration using every cell form passes")

	// 4. The closed `Resource` vocabulary. Cleanup selects the rows it must
	//    remove by this word, so a spelling it does not know is a resource
	//    nobody removes.

	// 4a. A word outside the five — the shape a project reaches for when it
	//     isolates something the contract does not name.
	r := wiIso(t, `| ~queue~ | ~QUEUE_NAME~ | ~appq~ | ~appq-<id>~ |
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |`, wiValidCmd)
	wiInvalid(t, r, "an unknown Resource word is a violation")
	wiReports(t, r, "QUEUE_NAME", "the unknown-Resource report names the row")
	wiReports(t, r, "queue", "the unknown-Resource report names the word it did not know")
	wiNotReported(t, r, "API_PORT", "the well-formed row beside it is not reported")

	// 4b. A near miss: a guard that matched on a prefix or a substring would
	//     pass it.
	r = wiIso(t, "| ~databases~ | ~DB_URL~ | ~appdb~ | ~appdb_<id_underscored>~ |", wiValidCmd)
	wiInvalid(t, r, "a near-miss Resource word is a violation")
	wiReports(t, r, "databases", "the near-miss report names the word")

	// 4c. `cache index` is two words, and its internal space is part of the
	//     word.
	r = wiIso(t, "| ~cacheindex~ | ~CACHE_INDEX~ | ~0~ | ~probed~ |", wiValidCmd)
	wiInvalid(t, r, "cacheindex is not the two-word cache index")
	wiReports(t, r, "cacheindex", "the report names the run-together spelling")

	// 4d. Case and surrounding whitespace are not the mistake.
	wiOK(t, wiIso(t, `| ~Database~ | ~DB_URL~ | ~appdb~ | ~appdb_<id_underscored>~ |
| ~ CACHE INDEX ~ | ~CACHE_INDEX~ | ~0~ | ~probed~ |`, wiValidCmd),
		"the Resource word is read case-insensitively and untrimmed")

	// 5. The `Variable` cell. A run EXPORTS these, so a cell that is not a
	//    legal environment-variable name is a row nothing can carry.
	r = wiIso(t, "| ~port~ | ~~ | ~8080~ | ~+<offset>~ |", wiValidCmd)
	wiInvalid(t, r, "an empty Variable is a violation")
	wiReports(t, r, "Variable", "the empty-Variable report names the cell")

	r = wiIso(t, "| ~port~ | ~API-PORT~ | ~8080~ | ~+<offset>~ |", wiValidCmd)
	wiInvalid(t, r, "a Variable that is not a legal environment-variable name")
	wiReports(t, r, "API-PORT", "the report names the illegal variable")

	r = wiIso(t, "| ~port~ | ~8080_PORT~ | ~8080~ | ~+<offset>~ |", wiValidCmd)
	wiInvalid(t, r, "a Variable starting with a digit")
	wiReports(t, r, "8080_PORT", "the report names the variable starting with a digit")

	// 5d. The contract's shape is `[A-Za-z_][A-Za-z0-9_]*`, not SCREAMING_CASE.
	wiOK(t, wiIso(t, "| ~port~ | ~_api_port~ | ~8080~ | ~+<offset>~ |", wiValidCmd),
		"a lowercase Variable with a leading underscore is legal")

	// 6. The bare-integer `Default`, which binds `port` and `cache index`.
	r = wiIso(t, "| ~port~ | ~API_PORT~ | ~localhost:8080~ | ~+<offset>~ |", wiValidCmd)
	wiInvalid(t, r, "a port Default carrying a host is a violation")
	wiReports(t, r, "API_PORT", "the port-Default report names the row")
	wiReports(t, r, "localhost:8080", "the port-Default report names the value it rejected")

	wiInvalid(t, wiIso(t, "| ~port~ | ~API_PORT~ | ~~ | ~+<offset>~ |", wiValidCmd),
		"an empty port Default is a violation")

	r = wiIso(t, "| ~cache index~ | ~CACHE_INDEX~ | ~db0~ | ~probed~ |", wiValidCmd)
	wiInvalid(t, r, "a cache index Default that is not a bare integer")
	wiReports(t, r, "CACHE_INDEX", "the cache-index-Default report names the row")
	wiReports(t, r, "db0", "the cache-index-Default report names the value it rejected")

	wiOK(t, wiIso(t, "| ~port~ | ~API_PORT~ |   ~8080~   | ~+<offset>~ |", wiValidCmd),
		"a padded bare integer is still a bare integer")

	// 6e. The rule binds `port` and `cache index` and NOTHING ELSE.
	wiOK(t, wiIso(t, `| ~database~ | ~DB_URL~ | ~jdbc:postgresql://localhost:5432/appdb~ | ~appdb_<id_underscored>~ |
| ~url~ | ~WEB_URL~ | ~http://localhost:3000~ | ~http://localhost:3000~ |`, wiValidCmd),
		"a non-integer Default on a database or url row is legal")

	// 7. The four `In a workspace` cell forms, one violated at a time. The
	//    form is decided by the ROW's Resource, so each case below is a cell
	//    that would be legal under some other form.
	r = wiIso(t, "| ~database~ | ~DB_URL~ | ~appdb~ | ~~ |", wiValidCmd)
	wiInvalid(t, r, "an empty database cell is a violation")
	wiReports(t, r, "DB_URL", "the empty-database-cell report names the row")

	r = wiIso(t, "| ~database~ | ~DB_URL~ | ~appdb~ | ~appdb_<workspace>~ |", wiValidCmd)
	wiInvalid(t, r, "an unknown token in a database cell is a violation")
	wiReports(t, r, "<workspace>", "the unknown-token report names the token")

	r = wiIso(t, `| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| ~bucket~ | ~MEDIA_BUCKET~ | ~appdb-media~ | ~appdb-media-<value:API_PORT>~ |`, wiValidCmd)
	wiInvalid(t, r, "a <value:…> reference in a bucket cell is a violation")
	wiReports(t, r, "MEDIA_BUCKET", "the misplaced-reference report names the row")

	// 7d. The cell that carries no token at all: a token-only check passes it.
	r = wiIso(t, "| ~port~ | ~API_PORT~ | ~8080~ | ~9090~ |", wiValidCmd)
	wiInvalid(t, r, "a port cell carrying a literal number is a violation")
	wiReports(t, r, "API_PORT", "the port-cell report names the row")
	wiReports(t, r, "9090", "the port-cell report names the cell it rejected")

	wiInvalid(t, wiIso(t, "| ~port~ | ~API_PORT~ | ~8080~ | ~8080+<offset>~ |", wiValidCmd),
		"a port cell with anything before the token is a violation")

	r = wiIso(t, "| ~cache index~ | ~CACHE_INDEX~ | ~0~ | ~3~ |", wiValidCmd)
	wiInvalid(t, r, "a cache index cell naming an index is a violation")
	wiReports(t, r, "CACHE_INDEX", "the cache-index-cell report names the row")

	wiInvalid(t, wiIso(t, "| ~cache index~ | ~CACHE_INDEX~ | ~0~ | ~probed at run time~ |", wiValidCmd),
		"a cache index cell with prose around the literal is a violation")

	r = wiIso(t, "| ~url~ | ~WEB_URL~ | ~http://localhost:3000~ | ~http://localhost:<port>~ |", wiValidCmd)
	wiInvalid(t, r, "an unknown token in a url cell is a violation")
	wiReports(t, r, "<port>", "the url-token report names the token")

	wiInvalid(t, wiIso(t, "| ~url~ | ~WEB_URL~ | ~http://localhost:3000~ | ~~ |", wiValidCmd),
		"an empty url cell is a violation")

	// 7j. The forms that must NOT be reported.
	wiOK(t, wiIso(t, `| ~url~ | ~STATIC_URL~ | ~http://localhost:9000~ | ~http://localhost:9000~ |
| ~url~ | ~DB_CONSOLE_URL~ | ~http://localhost:8081/?db=appdb~ | ~http://localhost:8081/?db=appdb_<id_underscored>~ |`, wiValidCmd),
		"a url row with no token, and one built from <id_underscored>, both pass")

	// 7k. A STRAY SPACE INSIDE THE BRACKETS matches the token shape nowhere,
	//     so every token rule passes over it: both rows shipped as
	//     ISOLATION-OK, exit 0, before the near-miss rule.
	r = wiIso(t, `| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| ~url~ | ~WEB_URL~ | ~http://localhost:8080~ | ~http://localhost:<value: API_PORT>~ |`, wiValidCmd)
	wiInvalid(t, r, "a space inside a <value:…> reference is a violation")
	wiReports(t, r, "<value: API_PORT>", "the spaced-token report quotes the run exactly as written")
	wiReports(t, r, "WEB_URL", "the spaced-token report names the row that carries it")

	r = wiIso(t, "| ~database~ | ~DB_URL~ | ~appdb~ | ~appdb_< id_underscored>~ |", wiValidCmd)
	wiInvalid(t, r, "a space inside <id_underscored> is a violation")
	wiReports(t, r, "< id_underscored>", "the spaced-token report quotes the run as written")

	// 7m. A bracketed run that is NOT a near miss of any contract token stays
	//     ordinary text.
	wiOK(t, wiIso(t, "| ~url~ | ~WEB_URL~ | ~http://localhost:3000/?q=a<b c>d~ | ~http://localhost:3000/?q=a<b c>d~ |", wiValidCmd),
		"a bracketed run that names no contract token is not reported")

	// 8. What a `url` row may reference, and what it may not.
	r = wiIso(t, `| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| ~url~ | ~WEB_URL~ | ~http://localhost:8080~ | ~http://localhost:<value:GATEWAY_PORT>~ |`, wiValidCmd)
	wiInvalid(t, r, "a <value:…> naming no row is a violation")
	wiReports(t, r, "GATEWAY_PORT", "the dangling-reference report names the variable it could not find")
	wiReports(t, r, "WEB_URL", "the dangling-reference report names the row that carries it")

	r = wiIso(t, `| ~url~ | ~BASE_URL~ | ~http://localhost:9000~ | ~http://localhost:9000~ |
| ~url~ | ~MEDIA_URL~ | ~http://localhost:9000/m~ | ~<value:BASE_URL>/m~ |`, wiValidCmd)
	wiInvalid(t, r, "a <value:…> naming another url row is a violation")
	wiReports(t, r, "BASE_URL", "the url-reference report names the row it pointed at")
	wiReports(t, r, "MEDIA_URL", "the url-reference report names the row that carries it")

	r = wiIso(t, `| ~cache index~ | ~CACHE_INDEX~ | ~0~ | ~probed~ |
| ~url~ | ~CACHE_URL~ | ~redis://localhost:6379/0~ | ~redis://localhost:6379/<value:CACHE_INDEX>~ |`, wiValidCmd)
	wiInvalid(t, r, "a <value:…> naming a cache index row is a violation")
	wiReports(t, r, "CACHE_INDEX", "the cache-index-reference report names the row it pointed at")
	wiReports(t, r, "CACHE_URL", "the cache-index-reference report names the row that carries it")

	// 8d. A reference that resolves to a row written after it: a guard that
	//     resolved as it read would make the table order load-bearing.
	wiOK(t, wiIso(t, `| ~url~ | ~WEB_URL~ | ~http://localhost:8080~ | ~http://localhost:<value:API_PORT>~ |
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |`, wiValidCmd), "a <value:…> resolving to a later row passes")

	wiOK(t, wiIso(t, `| ~database~ | ~DB_NAME~ | ~appdb~ | ~appdb_<id_underscored>~ |
| ~bucket~ | ~MEDIA_BUCKET~ | ~appdb-media~ | ~appdb-media-<id>~ |
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| ~url~ | ~DB_URL~ | ~x/appdb~ | ~x/<value:DB_NAME>~ |
| ~url~ | ~MEDIA_URL~ | ~y/appdb-media~ | ~y/<value:MEDIA_BUCKET>~ |
| ~url~ | ~WEB_URL~ | ~http://localhost:8080~ | ~http://localhost:<value:API_PORT>~ |`, wiValidCmd),
		"references to database, bucket and port rows all resolve")

	// 8f. Two rows holding the same `Variable`: the reference names no
	//     unique row.
	r = wiIso(t, `| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| ~port~ | ~API_PORT~ | ~8081~ | ~+<offset>~ |`, wiValidCmd)
	wiInvalid(t, r, "two rows holding the same Variable is a violation")
	wiReports(t, r, "API_PORT", "the duplicate-Variable report names the variable")

	// 9. How many times each `Resource` word may appear.
	r = wiIso(t, `| ~database~ | ~DB_URL~ | ~appdb~ | ~appdb_<id_underscored>~ |
| ~database~ | ~REPORTING_DB_URL~ | ~appreport~ | ~appreport_<id_underscored>~ |`, wiValidCmd)
	wiInvalid(t, r, "a second database row is a violation")
	wiReports(t, r, "REPORTING_DB_URL", "the second-database report names the row")

	r = wiIso(t, `| ~bucket~ | ~MEDIA_BUCKET~ | ~appdb-media~ | ~appdb-media-<id>~ |
| ~bucket~ | ~BACKUP_BUCKET~ | ~appdb-backup~ | ~appdb-backup-<id>~ |`, wiValidCmd)
	wiInvalid(t, r, "a second bucket row is a violation")
	wiReports(t, r, "BACKUP_BUCKET", "the second-bucket report names the row")

	r = wiIso(t, `| ~cache index~ | ~CACHE_INDEX~ | ~0~ | ~probed~ |
| ~cache index~ | ~SESSION_INDEX~ | ~1~ | ~probed~ |`, wiValidCmd)
	wiInvalid(t, r, "a second cache index row is a violation")
	wiReports(t, r, "SESSION_INDEX", "the second-cache-index report names the row")

	wiOK(t, wiIso(t, `| ~port~ | ~GATEWAY_PORT~ | ~8080~ | ~+<offset>~ |
| ~port~ | ~SERVER_PORT~ | ~8081~ | ~+<offset>~ |
| ~port~ | ~ADMIN_PANEL_PORT~ | ~3001~ | ~+<offset>~ |
| ~url~ | ~BACKEND_URL~ | ~http://localhost:8081~ | ~http://localhost:<value:SERVER_PORT>~ |
| ~url~ | ~FRONTEND_URL~ | ~http://localhost:3000~ | ~http://localhost:3000~ |`, wiValidCmd),
		"several port and url rows are legal")

	// 10. The command table: three verbs, two columns.
	r = wiIso(t, wiValidRes, `| ~create~ | ~./scripts/workspace create~ |
| ~verify~ | ~./scripts/workspace verify~ |`)
	wiInvalid(t, r, "a command verb outside the three is a violation")
	wiReports(t, r, "verify", "the unknown-verb report names the verb")

	r = wiIso(t, wiValidRes, `| ~remove~ | ~./scripts/workspace remove~ |
| ~remove~ | ~./scripts/other remove~ |`)
	wiInvalid(t, r, "the same command verb twice is a violation")
	wiReports(t, r, "remove", "the duplicate-verb report names the verb")

	r = wiIso(t, wiValidRes, "| ~create~ | ~~ |")
	wiInvalid(t, r, "an empty Runs cell is a violation")
	wiReports(t, r, "create", "the empty-Runs report names the verb")

	r = wiIso(t, wiValidRes, "| ~remove~ | ~docker exec <container> dropdb <id_underscored>~ |")
	wiInvalid(t, r, "an unsubstituted token in a command is a violation")
	wiReports(t, r, "<container>", "the command-token report names the token")

	r = wiIso(t, wiValidRes, "| ~remove~ | ~./scripts/workspace remove <value:API_PORT>~ |")
	wiInvalid(t, r, "a <value:…> reference in a command is a violation")
	wiReports(t, r, "<value:API_PORT>", "the command-reference report names the token")

	// 10f. `create` and `remove` with no `survivors` is a supported state.
	wiOK(t, wiIso(t, wiValidRes, `| ~create~ | ~./scripts/workspace create~ |
| ~remove~ | ~./scripts/workspace remove~ |`), "a command table with no survivors row is legal")

	wiOK(t, wiIso(t, wiValidRes, "| ~survivors~ | ~./scripts/workspace survivors < /dev/null > /tmp/out.txt~ |"),
		"a shell redirection in a command is not read as a token")

	// 10h. A `|` inside a command cell, written `\|`, is part of the command.
	wiOK(t, wiIso(t, wiValidRes, `| ~survivors~ | ~./gradlew -q survivors \| awk "/appdb/"~ |`),
		"an escaped pipe inside a command cell does not split the row")

	r = wiIso(t, wiValidRes, "| ~remove~ | ~./scripts/workspace remove < id>~ |")
	wiInvalid(t, r, "a space inside a command token is a violation")
	wiReports(t, r, "< id>", "the command spaced-token report quotes the run as written")

	// 10j. A bracketed run holding only token characters and a space is a
	//      redirection, not a near miss — the case that actually pins the
	//      rule, since 10g's slashes fall outside the token characters.
	wiOK(t, wiIso(t, wiValidRes, "| ~survivors~ | ~./scripts/workspace survivors < in.txt > out.txt~ |"),
		"a redirection whose filename is token-shaped is still not a token")

	// 11. Declare the section at most once. The two sections carry DIFFERENT
	//     verbs, so nothing else in this fixture is malformed.
	r = wiFile(t, `# fixture

## workspace isolation

| Command | Runs |
|---------|------|
| ~create~ | ~./one create~ |

## workspace isolation

| Command | Runs |
|---------|------|
| ~survivors~ | ~./two survivors~ |
`)
	wiInvalid(t, r, "a duplicated heading is a violation")
	wiReports(t, r, "2", "the duplicate-heading report names the count")
	wiReports(t, r, "workspace isolation", "the duplicate-heading report names the heading")

	r = wiFile(t, "# fixture\n\n## workspace isolation\n\n## workspace isolation\n\n## Workspace Isolation\n")
	wiInvalid(t, r, "three headings are a violation")
	wiReports(t, r, "3", "the duplicate-heading report counts headings case-insensitively")

	// 11c. A `###` heading is a different heading and is not the section.
	wiOK(t, wiFile(t, `# fixture

## workspace isolation

| Command | Runs |
|---------|------|
| ~survivors~ | ~./one survivors~ |

### workspace isolation notes

Prose that is not a second declaration.
`), "a deeper heading is not a second declaration")

	// 12. The shape of the two tables.
	r = wiFile(t, `# fixture

## workspace isolation

| Resource | Variable | In a workspace |
|----------|----------|----------------|
| ~port~ | ~API_PORT~ | ~+<offset>~ |
`)
	wiInvalid(t, r, "a resource table missing a column is a violation")
	// The report has to carry the order the columns were meant to be in.
	wiReports(t, r, "Resource | Variable | Default | In a workspace", "the header report names the order it expected")

	wiInvalid(t, wiFile(t, `# fixture

## workspace isolation

| Resource | Variable | In a workspace | Default |
|----------|----------|----------------|---------|
| ~port~ | ~API_PORT~ | ~+<offset>~ | ~8080~ |
`), "a reordered resource header is a violation")

	wiInvalid(t, wiFile(t, `# fixture

## workspace isolation

| Thing | Value |
|-------|-------|
| ~database~ | ~appdb~ |
`), "an unrecognised table header is a violation")

	// 12d. The needle is the COUNT, not the word "cell": a short row also
	//      leaves an empty `In a workspace` cell, and that rule reports the
	//      word.
	r = wiIso(t, "| ~port~ | ~API_PORT~ | ~8080~ |", wiValidCmd)
	wiInvalid(t, r, "a resource row with too few cells is a violation")
	wiReports(t, r, "expected 4", "the short-row report names the cell count it expected")

	wiInvalid(t, wiFile(t, "# fixture\n\n## workspace isolation\n\nProse describing what this project isolates, and no tables at all.\n"),
		"a section with neither table is a violation")

	wiOK(t, wiFile(t, `# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
`), "a resource table with no command table is a partial declaration, not a violation")

	// 12g. The two tables back to back: the report names the real cause, and
	//      the four misleading cell-count lines are gone.
	r = wiFile(t, `# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
| Command | Runs |
|---------|------|
| ~create~ | ~./scripts/workspace create~ |
| ~remove~ | ~./scripts/workspace remove <id>~ |
| ~survivors~ | ~./scripts/workspace survivors <id>~ |
`)
	wiInvalid(t, r, "two tables with no blank line between them is a violation")
	wiReports(t, r, "a second table header", "the merged-tables report names the real cause")
	wiReports(t, r, "blank line", "the merged-tables report says what to do about it")
	wiNotReported(t, r, "cell(s) where its header has", "the merged-tables report drops the misleading cell-count lines")

	// 13. The parser hazards. The input is tracked in the repository and
	//     editable in any pull request, so none of these may change what a
	//     row is read as.
	wiOK(t, wiIso(t, `| ~url~ | ~WEB_URL~ | ~http://localhost:3000/?a=1\|2~ | ~http://localhost:3000/?a=1\|2~ |`, wiValidCmd),
		"an escaped pipe inside a resource cell does not split the row")

	wiOK(t, wiFile(t, `# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace
|----------|----------|---------|----------------
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~
`), "rows with no trailing pipe are read correctly")

	wiOK(t, wiRaw(t, wiQ("# fixture\r\n\r\n## workspace isolation\r\n\r\n| Resource | Variable | Default | In a workspace |\r\n|---|---|---|---|\r\n| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |\r\n| ~cache index~ | ~CACHE_INDEX~ | ~0~ | ~probed~ |\r\n")),
		"a CRLF file is read the same as an LF one")

	wiOK(t, wiFile(t, `# fixture

## workspace isolation

  | Resource | Variable | Default | In a workspace |
  |----------|----------|---------|----------------|
  | ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |
`), "indented rows are read correctly")

	// 13e. Nothing in the file is EXECUTED.
	canary := t.TempDir() + "/canary"
	r = wiIso(t, wiValidRes, "| ~create~ | ~touch "+canary+"~ |\n| ~remove~ | ~$(touch "+canary+".sub)~ |")
	_, err := os.Lstat(canary)
	wiCheck(t, "a declared command is never run", errors.Is(err, os.ErrNotExist), "the guard ran a declared command: %s exists", canary)
	_, err = os.Lstat(canary + ".sub")
	wiCheck(t, "a command substitution in a cell is never evaluated", errors.Is(err, os.ErrNotExist),
		"the guard evaluated a command substitution: %s.sub exists", canary)

	// 13f. A table OUTSIDE the section is not read.
	wiOK(t, wiFile(t, `# fixture

## apps

| App | Repo root | Kind | URL |
|-----|-----------|------|-----|
| demo | ~/tmp/demo~ | Kotlin | http://localhost:8080 |

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| ~port~ | ~API_PORT~ | ~8080~ | ~+<offset>~ |

## lint

| Command | Runs |
|---------|------|
| ~nonsense~ | ~~ |
`), "tables outside the section are not read")

	// 13g. A cell carrying an ANSI escape sequence: a report whose bytes say
	//      ISOLATION-INVALID and whose rendering shows ISOLATION-OK. The
	//      assertion is on the BYTES.
	isoHead := "# fixture\n\n## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n"
	r = wiRaw(t, isoHead+wiQ("| ~queue\033[2K\033[1;32mISOLATION-OK: all good\033[0m~ | ~Q~ | ~x~ | ~x-<id>~ |\n"))
	wiInvalid(t, r, "a cell carrying an escape sequence is still reported as invalid")
	wiCheck(t, "no raw ESC byte from the configuration reaches stdout", !strings.Contains(r.out, "\033"),
		"an ESC byte from the configuration reached stdout: %s", r.out)
	wiReports(t, r, `\x1b[2K`, "the escape sequence is rendered visibly rather than stripped")
	wiReports(t, r, "ISOLATION-OK: all good", "the cell's own text is still quoted back to the operator")

	// 13g2. THE ENCODING IS INJECTIVE, driven as a collision: the two
	//       configurations differ in one byte and sit in the SAME project
	//       root, so the two reports are comparable in full.
	same := t.TempDir()
	writeFile(t, same+"/.flow/project.md", isoHead+wiQ("| ~queue\033[2K~ | ~Q~ | ~x~ | ~x-<id>~ |\n"))
	esc := wiRun(t, nil, same)
	wiInvalid(t, esc, "a cell holding a real ESC byte is reported as invalid")
	wiNotReported(t, esc, `\\x1b`, "a real ESC byte renders with a single backslash")
	writeFile(t, same+"/.flow/project.md", isoHead+wiQ(`| ~queue\x1b[2K~ | ~Q~ | ~x~ | ~x-<id>~ |`+"\n"))
	lit := wiRun(t, nil, same)
	wiInvalid(t, lit, `a cell holding the four literal characters \x1b is reported as invalid`)
	wiReports(t, lit, `\\x1b[2K`, "a literal backslash in the cell is doubled in the report")
	wiCheck(t, `a real ESC byte and the literal characters \x1b produce different reports`, esc.out != lit.out,
		"the display encoding is not injective: both produce %s", esc.out)

	// 13h. Non-ASCII is NOT escaped: a bucket carrying an accent is reported
	//      by its real name.
	r = wiIso(t, "| ~bücket~ | ~MEDIA_BUCKET~ | ~café-media~ | ~café-media-<id>~ |", wiValidCmd)
	wiInvalid(t, r, "an accented Resource word is reported")
	wiReports(t, r, "bücket", "a non-ASCII cell is reported by its real name, not escaped")

	// 15a. The validator itself failing to run. An empty report is what a
	//      crashed validator and a clean file look like from the outside. The
	//      bash's validator was an awk program, refused on when it printed no
	//      summary line; the port's validator is flow-guard, and the shim is
	//      what refuses when that cannot run. Driven for real: the shim and
	//      lib/flow-guard.sh copied into a tree with no stats/ to build from,
	//      run by bash against a well-formed project.
	wiRefuses(t, wiShimNoStats(t), "a validator that could not run")

	// 16. The no-argument default derives the repository root from the
	//     script's OWN resolved location, not from a fixed "one level up" —
	//     which only holds while it lives at <repo>/scripts/. Invoked through
	//     the skills/flow/scripts/ symlink with no argument, it must find THAT
	//     tree's own .flow/project.md — never the skill directory's, which has
	//     none and would silently read as "declares nothing".
	skillRoot, skillLink := wiSkillTree(t)
	wiConfig(t, skillRoot, wiIsoBody(wiValidRes, wiValidCmd))
	r = wiRun(t, map[string]string{"FLOW_GUARD_SELF": skillLink})
	wiOK(t, r, "case 16: no-arg default resolves through a skill-dir symlink to the real repo root")
	wiReports(t, r, "resource row(s) and", "case 16: the real .flow/project.md was read and validated")
	wiNotReported(t, r, "no .flow/project.md", "case 16: did not fall back to the skill directory, which has none")
	// The same bare run through the real shim, so its own
	// FLOW_GUARD_SELF="${BASH_SOURCE[0]}" line is what resolves the root.
	r = wiBash(t, skillLink, "/", append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t)))
	wiOK(t, r, "port: case 16 through the real shim resolves the skill-dir symlink to the real repo root")
	wiReports(t, r, "resource row(s) and", "port: case 16 through the real shim read and validated the real .flow/project.md")

	// 17. F1/F11 regression: an explicit project-root argument — the shape
	//     every real caller uses — must never depend on resolving the
	//     script's OWN location. The harness rigged `readlink` to fail; the
	//     port's rig is a FLOW_GUARD_SELF no resolution can finish, a symlink
	//     naming itself.
	loop := t.TempDir() + "/check-workspace-isolation.sh"
	if err := os.Symlink(loop, loop); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	wiConfig(t, project, wiIsoBody(wiValidRes, wiValidCmd))
	r = wiRun(t, map[string]string{"FLOW_GUARD_SELF": loop}, project)
	wiOK(t, r, "case 17 (F11): an explicit project root argument never triggers self-location resolution, so a rigged readlink cannot abort the run")
	wiReports(t, r, "resource row(s) and", "case 17 (F11): the given project root was validated, not skipped")
	wiCheck(t, "case 17 (F11): self-resolution was never attempted", !strings.Contains(r.err, "cannot resolve this script's own location"),
		"self-resolution was attempted even though an explicit root was given: %s", r.err)

	// Beyond the harness: the rig bites — the same FLOW_GUARD_SELF with no
	// argument is refused, so case 17 passing is the explicit root's doing.
	r = wiRun(t, map[string]string{"FLOW_GUARD_SELF": loop})
	wiRefuses(t, r, "port: an unresolvable FLOW_GUARD_SELF with no argument")
	wiCheck(t, "port: the refusal names the self-location failure", strings.Contains(r.err, "cannot resolve this script's own location"),
		"stderr: %s", r.err)
	wiRefuses(t, wiRun(t, nil), "port: no argument and no FLOW_GUARD_SELF")
}

// wiShimNoStats runs scripts/check-workspace-isolation.sh, copied with
// lib/flow-guard.sh into a tree with no stats/ beside it, on a well-formed
// project.
func wiShimNoStats(t *testing.T) wiResult {
	t.Helper()
	tree := t.TempDir()
	for _, rel := range []string{"scripts/check-workspace-isolation.sh", "scripts/lib/flow-guard.sh"} {
		b, err := os.ReadFile(wiRepoRoot + "/" + rel)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, tree+"/"+rel, string(b))
	}
	project := t.TempDir()
	wiConfig(t, project, wiIsoBody(wiValidRes, wiValidCmd))
	return wiBash(t, tree+"/scripts/check-workspace-isolation.sh", "/", append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+t.TempDir()), project)
}

// wiBash runs a bash script with env in dir, its two streams apart.
func wiBash(t *testing.T, script, dir string, env []string, args ...string) wiResult {
	t.Helper()
	var o, e bytes.Buffer
	cmd := exec.Command("bash", append([]string{script}, args...)...)
	cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, env, &o, &e
	err := cmd.Run()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return wiResult{cmd.ProcessState.ExitCode(), o.String(), e.String()}
}

// wiParity is the end-to-end old-vs-new comparison: the bash guard at
// c5379c0a and the port run over the same project roots, arguments,
// locale and CHECK_WORKSPACE_ISOLATION_PRINT_ROWS, their stdout, stderr and
// exit compared byte for byte. The fixtures reach what the cases above never
// name: every violation message, awk's [[:space:]] and tolower under each
// locale, the cell splitter's edges, NUL and CR bytes, the sanitizer, several
// roots in one run and a refusal after output. Two inputs are left out
// because the port reads them differently on purpose (workspaceisolation.go
// names both): a line that is not valid UTF-8 under a UTF-8 locale, and a
// heading whose space is U+00A0.
func wiParity(t *testing.T) {
	t.Parallel()
	tree := t.TempDir()
	for _, rel := range []string{"scripts/check-workspace-isolation.sh", "scripts/lib/resolve-file.sh"} {
		src, err := exec.Command(fixtureGit, "-C", wiRepoRoot, "show", "c5379c0a:"+rel).Output()
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, tree+"/"+rel, string(src))
	}
	base := t.TempDir()
	put := func(name, body string) { writeFile(t, base+"/"+name+"/.flow/project.md", body) }
	section := "# p\n\n## workspace isolation\n\n"
	put("valid", wiIsoBody(wiValidRes, wiValidCmd)+"\n")
	put("every-violation", section+wiQ(`| Resource | Variable | Default | In a workspace |
|---|:---:|---:|:---|
| ~queue~ | ~Q~ | ~x~ | ~x~ |
| ~port~ | ~~ | ~1~ | ~+<offset>~ |
| ~port~ | ~A-B~ | ~1~ | ~+<offset>~ |
| ~database~ | ~DB~ | ~d~ | ~d_<id>~ |
| ~database~ | ~DB2~ | ~d~ | ~d_<workspace>~ |
| ~bucket~ | ~B~ | ~b~ | ~b-<id>~ |
| ~bucket~ | ~B2~ | ~b~ | ~b-<id>~ |
| ~cache index~ | ~C~ | ~0~ | ~probed~ |
| ~cache index~ | ~C2~ | ~1~ | ~probed~ |
| ~port~ | ~DB~ | ~1~ | ~+<offset>~ |
| ~port~ | ~P~ | ~x1~ | ~+<offset>~ |
| ~port~ | ~P2~ | ~1~ | ~9~ |
| ~cache index~ | ~C3~ | ~0~ | ~3~ |
| ~url~ | ~U0~ | ~u~ | ~~ |
| ~database~ | ~D3~ | ~d~ | ~d_< id>~ |
| ~bucket~ | ~B3~ | ~b~ | ~b-<value:P>~ |
| ~url~ | ~U1~ | ~u~ | ~u/<foo>~ |
| ~url~ | ~U2~ | ~u~ | ~u/<value:NOPE>~ |
| ~url~ | ~U3~ | ~u~ | ~u/<value:U1>~ |
| ~url~ | ~U4~ | ~u~ | ~u/<value:C>~ |
| ~url~ | ~U5~ | ~u~ | ~u/<value:DB>/<value:B>/<value:P>/<value:>~ |
| ~url~ | ~U6~ | ~u~ | ~u/<value: P><id_ underscored>< offset ><my thing>~ |
| ~port~ | ~P4~ | ~1~ | ~+<offset>~ | extra |
|
| ~Port~ | ~P5~ | ~ 7 ~ | ~ +<offset> ~ |

| Thing | Value |
|---|---|
| x | y |

| Command | Runs |
|---|---|
| ~create~ | ~c~ |
| ~Verify~ | ~v~ |
| ~CREATE~ | ~c2~ |
| ~remove~ | ~~ |
| ~survivors~ | ~s <container> < id> <value:X>~ |
| ~x~ |

| Command | Runs |
|---|---|
| ~create~ | ~z~ |
| Resource | Variable | Default | In a workspace |
| ~port~ | ~ZZ~ | ~1~ | ~+<offset>~ |

### not the section
| ~bogus~ |
`))
	put("unicode", section+"| Resource | Variable | Default | In a workspace | Ünit ẞ \u212a |\n|---|---|---|---|---|\n\n"+
		"| RESOURCE | VARİABLE | DEFAULT | IN A WORKSPACE |\n|---|---|---|---|\n"+
		"| \u00a0`port`\u00a0 | `NB` | `1` | `+<offset>` |\n"+
		"| `cache\u00a0index` | `CI` | `0` | `probed` |\n"+
		"| `cache\t\tindex` | `CI2` | `0` | `probed` |\n"+
		"| `database` | `DBU` | `d` | `d_<\u00a0id>` |\n"+
		"| `BÜCKET` | `BU` | `b` | `b` |\n"+
		"| `bucket` | `B\u00a0U` | `b` | `b` |\n")
	put("parser", section+"| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n"+
		"| `url` | `T1` | `u` | `u\\` |\n"+
		"| `url` | `T2` | `u` | `a\\\\|b\\x` |\n"+
		"|`port`|`T3`|`1`|`+<offset>`|\n"+
		"\t| `url` | `T4` | `u` | ` a ` b ` |\n"+
		"| `url` | `T5` | `u\rv` | `w` |   \t\n"+
		"||\n|\n| `url` | `T6` | `u` | `v` |||\n"+
		"| `port` | `T7` | `+1` | `+<offset>` |\n"+
		"| `port` | `T8` | `08` | `+<offset>` |\n")
	put("nul", section+wiQ("| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n| ~port~ | ~N1~ | ~1~ | ~+<offset>~ |\x00| junk |\n| ~url~ | ~N2\x00~ | ~u~ | ~v~ |\n## x\x00\n| ~a~ |\n"))
	put("crlf", "# p\r\n\r\n## Workspace Isolation \t\r\n\r\n| Command | Runs |\r\n|---|---|\r\n| `create` | `c <id>` |\r\n| `remove` | `r <id_underscored>` |\r\n")
	put("controls", section+"| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n"+
		"| `q\x1b[1m\x7f\x01` | `V\\\\` | `x` | `x` |\n| `port` | `P\tQ` | `1` | `+<offset>` |\n")
	put("dup", "## workspace isolation\n## WORKSPACE ISOLATION\n#### workspace isolation\n##\tworkspace isolation   \n")
	put("none", "# p\n\n## workspace isolationx\n### workspace isolation\n")
	put("prose-only", "# p\n\n## lint\n\n## workspace isolation\n\nprose\n\n## run\n| Command | Runs |\n")
	put("empty-file", "")
	put("heading-last", "text\n## workspace isolation")
	mkdir(t, base+"/no-config")
	put("unterminated", section+"| Command | Runs |\n|---|---|\n| `create` | `c` |")
	mkdir(t, base+"/cfg-dir/.flow/project.md")
	put("bad-utf8", section+"| Resource \xff | Variable |\n|---|---|\n\n| Command | Runs |\n|---|---|\n| `cr\xffeate` | `c\xfe` |\n")
	all := []string{"valid", "every-violation", "unicode", "parser", "nul", "crlf", "controls", "dup", "none",
		"prose-only", "empty-file", "heading-last", "no-config", "./unterminated/", base + "/valid"}
	runs := [][]string{all, {"valid", "missing", "valid"}, {"crlf", "cfg-dir"}, {""}, {"valid"}}
	for _, loc := range []string{"C", "en_US.UTF-8"} {
		for _, rows := range []string{"", "1", "yes"} {
			for i, args := range append(runs, []string{"bad-utf8"}) {
				if loc != "C" && args[0] == "bad-utf8" {
					continue
				}
				// Each run is its own parallel subtest: they share only the
				// read-only trees above, and run one after another they held
				// one slot for ~9s, the last test of the package to finish.
				t.Run(fmt.Sprintf("run %d LC_ALL=%s PRINT_ROWS=%q", i, loc, rows), func(t *testing.T) {
					t.Parallel()
					env := []string{"LC_ALL=" + loc}
					for _, kv := range os.Environ() {
						if !strings.HasPrefix(kv, "LC_ALL=") && !strings.HasPrefix(kv, "CHECK_WORKSPACE_ISOLATION_PRINT_ROWS=") {
							env = append(env, kv)
						}
					}
					vars := map[string]string{"LC_ALL": loc}
					if rows != "" {
						env = append(env, "CHECK_WORKSPACE_ISOLATION_PRINT_ROWS="+rows)
						vars["CHECK_WORKSPACE_ISOLATION_PRINT_ROWS"] = rows
					}
					want := wiBash(t, tree+"/scripts/check-workspace-isolation.sh", base, env, args...)
					fn := Registry["check-workspace-isolation"]
					if fn == nil {
						t.Fatal("check-workspace-isolation is not registered")
					}
					var o, e bytes.Buffer
					rc := fn(args, Env{Dir: base, Getenv: func(k string) string { return vars[k] }}, &o, &e)
					if rc != want.rc || o.String() != want.out || e.String() != want.err {
						t.Errorf("run %d LC_ALL=%s PRINT_ROWS=%q args=%q:\nbash rc=%d\n%s\nstderr:\n%s\nport rc=%d\n%s\nstderr:\n%s",
							i, loc, rows, args, want.rc, want.out, want.err, rc, o.String(), e.String())
					}
				})
			}
		}
	}
}
