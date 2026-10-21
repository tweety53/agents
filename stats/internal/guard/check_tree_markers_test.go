package guard

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-check-tree-markers.sh at c9b8fde8, one subtest
// per ok: label, plus the empty-path refusal the harness never reached. The
// harness built throwaway git repositories and ran the guard with stdout and
// stderr merged; here the guard runs in-process over a copy of one master
// repository built once, so the KAN-579 shape (case 5: never-committed
// content destroyed by a real `git checkout`) still crosses real git.

// tmMaster is make_repo's tree: one committed artifact holding two known
// strings, and a notes file whose uncommitted edit carries a string of its own.
var tmMaster = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "check-tree-markers-master")
	if err != nil {
		return "", err
	}
	g := &fxGit{}
	g.git(dir, "init", "-q")
	if g.err == nil {
		g.err = os.MkdirAll(dir+"/spectre/changes/demo", 0o755)
	}
	if g.err == nil {
		g.err = os.WriteFile(dir+"/spectre/changes/demo/design.md", []byte("## Decisions\n\n**ID:** the-decision\n"), 0o644)
	}
	g.write(dir+"/notes.md", "working notes: committed baseline")
	g.git(dir, "add", "-A")
	g.git(dir, "commit", "-q", "-m", "original")
	g.appendLine(dir+"/notes.md", "working notes: uncommitted")
	return dir, g.err
})

// tmMarkersBody is make_markers' file: two spec lines on the committed
// artifact, the notes marker that exists only in the uncommitted edit, and one
// line whose marker matches nothing.
const tmMarkersBody = "spectre/changes/demo/design.md\tthe-decision\n" +
	"spectre/changes/demo/design.md\t## Decisions\n" +
	"notes.md\tuncommitted\n" +
	"notes.md\tno-such-marker\n"

type tmFx struct {
	t                       *testing.T
	wt, markers, snap, root string
}

// tmNew copies the master into a fresh directory and lays out a markers file
// and a scratch snapshot path beside it.
func tmNew(t *testing.T) *tmFx {
	t.Helper()
	master, err := tmMaster()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if out, err := exec.Command("cp", "-R", master, root+"/wt").CombinedOutput(); err != nil {
		t.Fatalf("cp: %v\n%s", err, out)
	}
	writeFile(t, root+"/m/markers", tmMarkersBody)
	mkdir(t, root+"/s")
	return &tmFx{t, root + "/wt", root + "/m/markers", root + "/s/snapshot", root}
}

func (fx *tmFx) run(args ...string) guardResult {
	return runGuard("check-tree-markers", args, Env{Getenv: os.Getenv, Dir: fx.root})
}

type tmCheck struct{ label, fail string }

func tmExit(label string, r guardResult, want int) tmCheck {
	if r.rc != want {
		return tmCheck{label, "expected exit " + strconv.Itoa(want) + ", got " + strconv.Itoa(r.rc) + "\n" + r.out}
	}
	return tmCheck{label: label}
}

func tmNames(label string, r guardResult, needle string) tmCheck {
	if !strings.Contains(r.out, needle) {
		return tmCheck{label, "expected " + needle + " named in output, got:\n" + r.out}
	}
	return tmCheck{label: label}
}

func tmFileLine(label, file, needle string) tmCheck {
	b, _ := os.ReadFile(file)
	if !strings.Contains(string(b), needle) {
		return tmCheck{label, "expected " + needle + " in " + file + ", got:\n" + string(b)}
	}
	return tmCheck{label: label}
}

func tmExact(label, got, want string) tmCheck {
	if got != want {
		return tmCheck{label, "got:\n" + strconv.Quote(got) + "\nwant:\n" + strconv.Quote(want)}
	}
	return tmCheck{label: label}
}

func TestCheckTreeMarkers(t *testing.T) {
	t.Parallel()
	var checks []tmCheck
	add := func(c ...tmCheck) { checks = append(checks, c...) }

	// Case 1: snapshot records counts.
	fx := tmNew(t)
	add(tmExit("case 1a: snapshot exits 0", fx.run("snapshot", fx.wt, fx.markers, fx.snap), 0),
		tmFileLine("case 1b: records the decision-id marker count", fx.snap, "spectre/changes/demo/design.md\tthe-decision\t1"),
		tmFileLine("case 1c: records the heading marker count", fx.snap, "spectre/changes/demo/design.md\t## Decisions\t1"),
		tmFileLine("case 1d: records a zero count for an absent marker", fx.snap, "notes.md\tno-such-marker\t0"))
	// Case 2: verify on an untouched tree.
	add(tmExit("case 2: verify clean on an untouched tree exits 0", fx.run("verify", fx.wt, fx.markers, fx.snap), 0))

	// Case 3: an artifact rewritten without its known string.
	fx = tmNew(t)
	add(tmExit("case 3a: snapshot exits 0", fx.run("snapshot", fx.wt, fx.markers, fx.snap), 0))
	writeFile(t, fx.wt+"/spectre/changes/demo/design.md", "## Decisions\n\n**ID:** another-decision\n")
	r := fx.run("verify", fx.wt, fx.markers, fx.snap)
	add(tmExit("case 3b: verify after a marker-losing rewrite exits 1", r, 1),
		tmNames("case 3c: names the rewritten artifact", r, "design.md"),
		tmNames("case 3d: says the clean-state claim is not evidence", r, "not evidence"))

	// Case 4: an artifact deleted after the snapshot.
	fx = tmNew(t)
	add(tmExit("case 4a: snapshot exits 0", fx.run("snapshot", fx.wt, fx.markers, fx.snap), 0))
	if err := os.Remove(fx.wt + "/spectre/changes/demo/design.md"); err != nil {
		t.Fatal(err)
	}
	r = fx.run("verify", fx.wt, fx.markers, fx.snap)
	add(tmExit("case 4b: verify after a deletion exits 1", r, 1),
		tmNames("case 4c: names the deleted artifact", r, "design.md"))

	// Case 5: THE KAN-579 SHAPE -- never-committed content destroyed by a
	// real `git checkout` after the snapshot.
	fx = tmNew(t)
	add(tmExit("case 5a: snapshot exits 0", fx.run("snapshot", fx.wt, fx.markers, fx.snap), 0))
	gitRun(t, fx.wt, "checkout", "--", ".")
	r = fx.run("verify", fx.wt, fx.markers, fx.snap)
	add(tmExit("case 5b: verify after the checkout destroys the uncommitted edit exits 1", r, 1),
		tmNames("case 5c: names the destroyed notes file", r, "notes.md"))

	// Case 6: churn outside the marker list is clean.
	fx = tmNew(t)
	add(tmExit("case 6a: snapshot exits 0", fx.run("snapshot", fx.wt, fx.markers, fx.snap), 0))
	writeFile(t, fx.wt+"/src.txt", "source: edited\n")
	writeFile(t, fx.wt+"/planted.md", "planted\n")
	gitRun(t, fx.wt, "add", "-A")
	gitRun(t, fx.wt, "commit", "-q", "-m", "implementer-work")
	add(tmExit("case 6: verify ignores churn outside the marker list exits 0", fx.run("verify", fx.wt, fx.markers, fx.snap), 0))

	// Case 7: exit 2 refusals.
	fx = tmNew(t)
	bad := fx.root + "/bad"
	writeFile(t, bad+"/absolute", "/etc/passwd\troot\n")
	writeFile(t, bad+"/dotdot", "../escape\troot\n")
	writeFile(t, bad+"/notab", "no-tab-line\n")
	writeFile(t, bad+"/emptymarker", "notes.md\t\n")
	writeFile(t, bad+"/emptypath", "\troot\n")
	add(tmExit("case 7a: no arguments exits 2", fx.run(), 2),
		tmExit("case 7b: a bad mode exits 2", fx.run("check", fx.wt, fx.snap, fx.snap), 2),
		tmExit("case 7c: a non-directory worktree exits 2", fx.run("snapshot", fx.wt+"/no-such-dir", fx.snap, fx.snap), 2),
		tmExit("case 7d: an unreadable markers file exits 2", fx.run("snapshot", fx.wt, fx.snap+"/no-such-markers", fx.snap), 2),
		tmExit("case 7e: an absolute path in the markers file exits 2", fx.run("snapshot", fx.wt, bad+"/absolute", fx.snap), 2),
		tmExit("case 7f: a .. path in the markers file exits 2", fx.run("snapshot", fx.wt, bad+"/dotdot", fx.snap), 2),
		tmExit("case 7g: a line without a tab exits 2", fx.run("snapshot", fx.wt, bad+"/notab", fx.snap), 2),
		tmExit("case 7h: an empty marker exits 2", fx.run("snapshot", fx.wt, bad+"/emptymarker", fx.snap), 2),
		tmExit("case 7i: snapshot to an unwritable path exits 2", fx.run("snapshot", fx.wt, fx.markers, fx.snap+"/no-dir/snapshot"), 2),
		tmExit("case 7j: an empty path in the markers file exits 2", fx.run("snapshot", fx.wt, bad+"/emptypath", fx.snap), 2))
	// 7k: an existing regular file as the worktree is refused as a
	// non-directory before any marker is read.
	r = fx.run("snapshot", fx.markers, fx.markers, fx.snap)
	add(tmExit("case 7k: a regular-file worktree exits 2", r, 2),
		tmExact("case 7k: a regular-file worktree is named as not a directory", r.err, "check-tree-markers: not a directory: "+fx.markers+"\n"))

	// Case 8: verify with no readable snapshot file.
	fx = tmNew(t)
	add(tmExit("case 8: verify without a snapshot file exits 2", fx.run("verify", fx.wt, fx.markers, fx.root+"/s/absent"), 2))

	// Case 9: F1 -- an existing directory as the snapshot path.
	fx = tmNew(t)
	add(tmExit("case 9: snapshot to an existing directory exits 2", fx.run("snapshot", fx.wt, fx.markers, fx.root+"/s"), 2))

	// Case 9b/9c: the snapshot-path checks run before any marker is
	// recomputed, as the bash ran them, so an unusable snapshot path is
	// named even when a marker's ERE would also fail.
	fx = tmNew(t)
	writeFile(t, fx.root+"/badere/markers", "notes.md\t(\n")
	r = fx.run("snapshot", fx.wt, fx.root+"/badere/markers", fx.root+"/s")
	add(tmExact("case 9b: a directory snapshot path is named before a failing grep", r.err,
		"check-tree-markers: snapshot path is an existing directory: "+fx.root+"/s\n"))
	r = fx.run("verify", fx.wt, fx.root+"/badere/markers", fx.root+"/s/absent")
	add(tmExact("case 9c: a missing snapshot is named before a failing grep", r.err,
		"check-tree-markers: no readable snapshot at "+fx.root+"/s/absent -- cannot answer; snapshot runs before the dispatch, never after it\n"))

	// Case 10: F2 -- a markers file with zero valid lines.
	fx = tmNew(t)
	writeFile(t, fx.root+"/empty/markers", "")
	add(tmExit("case 10a: snapshot over an empty markers file exits 2", fx.run("snapshot", fx.wt, fx.root+"/empty/markers", fx.snap), 2),
		tmExit("case 10b: verify over an empty markers file exits 2", fx.run("verify", fx.wt, fx.root+"/empty/markers", fx.snap), 2))

	// Case 11: the bytes the bash wrote and printed at c9b8fde8, captured
	// by running it over this case's tree with a relative snapshot path --
	// so a snapshot written before the port verifies after it.
	fx = tmNew(t)
	r = fx.run("snapshot", "wt", fx.markers, "s/snapshot")
	add(tmExact("case 11a: snapshot prints the bash's OK line", r.stdout, "TREE-MARKERS-SNAPSHOT-OK: markers pinned at s/snapshot\n"))
	snapBytes, _ := os.ReadFile(fx.snap)
	add(tmExact("case 11b: the snapshot file carries the bash's bytes", string(snapBytes),
		"spectre/changes/demo/design.md\tthe-decision\t1\n"+
			"spectre/changes/demo/design.md\t## Decisions\t1\n"+
			"notes.md\tuncommitted\t1\n"+
			"notes.md\tno-such-marker\t0\n"))
	if fi, err := os.Stat(fx.snap); err == nil {
		add(tmExact("case 11c: the snapshot file is mode 0600 as mktemp made it", fi.Mode().Perm().String(), "-rw-------"))
	} else {
		add(tmCheck{"case 11c: the snapshot file is mode 0600 as mktemp made it", err.Error()})
	}
	r = fx.run("verify", "wt", fx.markers, "s/snapshot")
	add(tmExact("case 11d: verify prints the bash's OK line", r.stdout, "TREE-MARKERS-OK: every recorded marker matches its pre-dispatch snapshot\n"))
	if err := os.Remove(fx.wt + "/notes.md"); err != nil {
		t.Fatal(err)
	}
	r = fx.run("verify", "wt", fx.markers, "s/snapshot")
	add(tmExact("case 11e: a failed verify writes the bash's stderr", r.err,
		"check-tree-markers: the tree's recorded markers changed across the dispatch -- the dispatch's own clean-or-restored claim is not evidence; stop and recover by hand from before the dispatch\n"+
			"check-tree-markers: 3,4c3,4\n"+
			"check-tree-markers: < notes.md\tuncommitted\t1\n"+
			"check-tree-markers: < notes.md\tno-such-marker\t0\n"+
			"check-tree-markers: ---\n"+
			"check-tree-markers: > notes.md\tuncommitted\tmissing\n"+
			"check-tree-markers: > notes.md\tno-such-marker\tmissing\n"),
		tmExact("case 11f: a failed verify prints nothing on stdout", r.stdout, ""))
	r = fx.run()
	add(tmExact("case 11g: the usage line is the bash's", r.err,
		"check-tree-markers: usage: check-tree-markers.sh snapshot|verify <worktree> <markers-file> <snapshot-file>\n"))

	// Case 12: the ERE dialect, and a blank markers-file line skipped.
	// `uncommit+ed` and `the-decision|nothing` match only as EREs; a BRE grep
	// would count both 0.
	fx = tmNew(t)
	writeFile(t, fx.root+"/ere/markers", "notes.md\tuncommit+ed\n\nspectre/changes/demo/design.md\tthe-decision|nothing\n")
	add(tmExit("case 12a: snapshot over a markers file with a blank line exits 0", fx.run("snapshot", fx.wt, fx.root+"/ere/markers", fx.snap), 0),
		tmFileLine("case 12b: an ERE `+` marker is counted", fx.snap, "notes.md\tuncommit+ed\t1"),
		tmFileLine("case 12c: an ERE `|` marker is counted", fx.snap, "spectre/changes/demo/design.md\tthe-decision|nothing\t1"))

	// Case 13: grep failing (an invalid ERE, exit 2) cannot answer.
	fx = tmNew(t)
	writeFile(t, fx.root+"/badere/markers", "notes.md\t(\n")
	r = fx.run("snapshot", fx.wt, fx.root+"/badere/markers", fx.snap)
	add(tmExit("case 13a: snapshot whose grep fails exits 2", r, 2),
		tmNames("case 13b: names the grep failure", r, "grep failed"))

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}
