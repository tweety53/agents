package repotest

import (
	"path/filepath"
	"testing"
)

// TestProtectMainCheckoutRecognisesFlowGuardSpelling: a guard invoked as
// `flow-guard check-…` — bare, or as `…/scripts/flow-guard check-…` — is read
// as a guard invocation, as its `check-….sh` spelling is: a command that only
// names a data heredoc's target as that guard's argument keeps the body as
// data, so the body's `rm` and `>` text is not scanned as shell.
func TestProtectMainCheckoutRecognisesFlowGuardSpelling(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	main := root + "/main"
	f.vars["MAIN"], f.vars["SIB"] = main, root+"/main-worktrees/sib"
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "remote", "set-head", "origin", "main")

	body := "C={SIB}; cat > $C/t.md <<'EOF'\n> **Execution:** a > b; rm x\nEOF\n"
	f.bash("bare flow-guard check-untasked-commits", "allow", "{MAIN}", body+"flow-guard check-untasked-commits $C/t.md")
	f.bash("path-named flow-guard check-loop-landing", "allow", "{MAIN}", body+"{ROOT}/agents/scripts/flow-guard check-loop-landing $C/t.md")
	f.bash("check-untasked-commits.sh", "allow", "{MAIN}", body+"check-untasked-commits.sh $C/t.md")
	f.bash("path-named check-loop-landing.sh", "allow", "{MAIN}", body+"{ROOT}/agents/scripts/check-loop-landing.sh $C/t.md")
	f.bash("flow-guard run of a non-guard word keeps the body scanned", "deny", "{MAIN}", body+"flow-guard sh $C/t.md")
}
