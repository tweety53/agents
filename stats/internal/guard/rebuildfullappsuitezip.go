// rebuildfullappsuitezip.go is rebuild-full-app-suite-zip.sh ported to Go
// (KAN-1020).
//
// rebuild-full-app-suite-zip.sh — rebuild <screenshots root>/full-app-suite.zip
// from exactly the PNGs full-app-suite.spec.ts produced. The rebuild lived only
// as inline bash in step 8 of skills/flow/visual-verify-verifier.md, so an
// in-run recapture never ran it and kan-1003's zip went stale twice; the
// verifier, the in-run fix loop (skills/flow/verify-fix-loop.md) and the base
// sync's recapture (skills/flow/sync-onto-base.md) all call this one script.
//
// Usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>
//
// The PNG list is `resolve-visual-screenshots.sh <worktree> full-app-suite.spec.ts`.
// The old zip is removed first, so the archive never keeps a PNG the suite no
// longer produces. Every PNG's mtime is pinned to 1980-01-01 before zipping,
// since the zip stores it: a rebuild over unchanged images is byte-identical
// and commits nothing.
//
// Exit 0 the zip is written; 1 the resolver matched no PNG (nothing written);
// 2 cannot answer — a usage error, a relative <screenshots root> (it would land
// under the caller's working directory), a resolved PNG outside <screenshots
// root> (another checkout's root), the resolver's own exit 2, or a failing zip.
//
// The resolver is exec'd from FLOW_GUARD_SCRIPT_DIR, the directory the shim
// exports as its own. The archive is still Info-ZIP's `zip -q -X -j -@`, never
// archive/zip, whose headers differ: the first rebuild after the port would
// otherwise commit a different zip over unchanged images (decision zip-via-exec).

package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() { Registry["rebuild-full-app-suite-zip"] = rebuildFullAppSuiteZip }

func rebuildFullAppSuiteZip(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>")
		return 2
	}
	if !filepath.IsAbs(args[1]) {
		fmt.Fprintf(stderr, "rebuild-full-app-suite-zip: <screenshots root> %q is relative — pass it resolved the way resolve-visual-screenshots.sh resolves it\n", args[1])
		return 2
	}
	scriptDir := env.Getenv("FLOW_GUARD_SCRIPT_DIR")
	if scriptDir == "" {
		fmt.Fprintln(stderr, "rebuild-full-app-suite-zip: FLOW_GUARD_SCRIPT_DIR is unset — run scripts/rebuild-full-app-suite-zip.sh, which sets it")
		return 2
	}
	abs := func(p string) string {
		if env.Dir != "" && !filepath.IsAbs(p) {
			return filepath.Join(env.Dir, p)
		}
		return p
	}
	zipPath := args[1] + "/full-app-suite.zip"

	resolve := exec.Command(filepath.Join(scriptDir, "resolve-visual-screenshots.sh"), args[0], "full-app-suite.spec.ts")
	resolve.Dir, resolve.Env = env.Dir, icChildEnv(env)
	resolve.Stderr = stderr
	out, err := resolve.Output()
	list, rc := strings.TrimRight(string(out), "\n"), exitCode(err)
	// A resolver that cannot answer (exit 2) leaves the old zip alone: only a resolved list replaces it.
	if rc == 2 {
		return 2
	}
	// A root no resolved PNG lies under is the wrong root: zipping there
	// leaves a stray archive in another checkout (kan-1009 wrote one into
	// the backend main checkout). Refused before the old zip is touched.
	if rc == 0 {
		for _, png := range strings.Split(list, "\n") {
			if !strings.HasPrefix(abs(png), filepath.Clean(args[1])+"/") {
				fmt.Fprintf(stderr, "rebuild-full-app-suite-zip: %s lies outside <screenshots root> %s — pass `screenshots` resolved the way resolve-visual-screenshots.sh resolves it\n", png, args[1])
				return 2
			}
		}
	}
	os.Remove(abs(zipPath))
	if rc != 0 {
		return rc
	}
	epoch := time.Date(1980, 1, 1, 0, 0, 0, 0, time.Local)
	for _, png := range strings.Split(list, "\n") {
		// touch -t: creates a missing file; a failure exits as xargs did, 123.
		f, err := os.OpenFile(abs(png), os.O_CREATE|os.O_WRONLY, 0o666)
		if err == nil {
			f.Close()
			err = os.Chtimes(abs(png), epoch, epoch)
		}
		if err != nil {
			fmt.Fprintf(stderr, "rebuild-full-app-suite-zip: cannot pin the mtime of %s: %v\n", png, err)
			return 123
		}
	}
	zipBin := "zip"
	if p, ok := lookPath(env, "zip"); ok {
		zipBin = p
	}
	zip := exec.Command(zipBin, "-q", "-X", "-j", "-@", zipPath)
	zip.Dir, zip.Env = env.Dir, icChildEnv(env)
	zip.Stdin = strings.NewReader(list + "\n")
	zip.Stdout, zip.Stderr = stdout, stderr
	if zip.Run() != nil {
		return 2
	}
	return 0
}
