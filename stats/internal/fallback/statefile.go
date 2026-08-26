// Package fallback implements the CLI's non-blocking escape hatch: the
// on-disk state file and its journal, and the project-key resolution that
// makes both address the same record from a worktree and from its main
// checkout. Nothing here talks HTTP -- internal/client owns that -- so this
// package is exactly what a store outage leaves the CLI with.
package fallback

import (
	"crypto/sha1" // reproduces skills/myflow-contracts/state-file.md's own key-derivation formula verbatim (`shasum`'s default algorithm) -- not a security use.
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultStateRoot is the root of every project's state directory, exactly
// as skills/myflow-contracts/state-file.md names it.
const DefaultStateRoot = "/Users/tweety53/Agents/myflow/state"

// StateRoot returns the root of every project's state directory:
// FLOW_STATE_DIR when set (a test's own isolated root), DefaultStateRoot
// otherwise.
func StateRoot() string {
	if v := os.Getenv("FLOW_STATE_DIR"); v != "" {
		return v
	}
	return DefaultStateRoot
}

// StateFilePath returns the on-disk path for the state file recording
// projectKey/name, per skills/myflow-contracts/state-file.md.
func StateFilePath(projectKey, name string) string {
	return filepath.Join(StateRoot(), projectKey, name+".json")
}

// JournalFilePath returns the path of the journal beside the state file
// recording projectKey/name. "Beside it" is design.md's own phrase for the
// journal's placement (see "Availability and reconciliation").
func JournalFilePath(projectKey, name string) string {
	return filepath.Join(StateRoot(), projectKey, name+".journal")
}

// ProjectKey resolves the project key and the main checkout's absolute path
// for the git repository containing dir, using exactly the algorithm
// skills/myflow-contracts/state-file.md specifies:
//
//	MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)"
//	PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
//
// `--git-common-dir` (never `--show-toplevel`) is what makes this resolve
// to the *main* checkout from inside a worktree too -- show-toplevel would
// return the worktree's own root, giving a worktree and its main checkout
// two different records for the same change. git-common-dir's output is
// relative to dir when dir is itself the main checkout (git prints ".git"),
// and absolute when dir is a worktree (git prints the main repo's .git
// path directly) -- both cases are handled by joining a relative result
// onto dir before taking its parent.
//
// The result is then resolved through EvalSymlinks -- the Go-side
// equivalent of the contract's own `pwd -P` (physical, symlink-resolved)
// rather than a bare `pwd`. This matters in practice, not just in theory:
// git itself stores a worktree's gitdir pointer as an already-symlink-
// resolved absolute path, while the main checkout side of this function
// only ever gets as far as filepath.Abs on whatever path dir was called
// with. On a machine where the relevant directory is reached through a
// symlink (macOS's /var -> /private/var is the everyday example -- it is
// exactly what a t.TempDir() path crosses), the two sides would otherwise
// disagree on spelling for the identical directory, and two different
// project keys would result for one project. EvalSymlinks makes both
// sides agree on the same canonical spelling, exactly as `pwd -P` does for
// the shell recipe above. A path that cannot be resolved (already removed,
// or a permissions error) falls back to the unresolved, cleaned path
// rather than failing the whole call -- callers need a key even when this
// best-effort canonicalisation cannot run.
func ProjectKey(dir string) (key, mainCheckout string, err error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("fallback: resolve %q: %w", dir, err)
	}

	out, err := exec.Command("git", "-C", absDir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", "", fmt.Errorf("fallback: git rev-parse --git-common-dir in %s: %w", absDir, err)
	}

	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(absDir, commonDir)
	}
	mainCheckout = filepath.Clean(filepath.Dir(commonDir))
	if resolved, evalErr := filepath.EvalSymlinks(mainCheckout); evalErr == nil {
		mainCheckout = resolved
	}

	sum := sha1.Sum([]byte(mainCheckout))
	hash := hex.EncodeToString(sum[:])[:8]
	key = filepath.Base(mainCheckout) + "-" + hash
	return key, mainCheckout, nil
}

// ListStateFileNames returns the basenames (minus ".json") of every state
// file directly under projectKey's state directory -- the on-disk
// fallback records a failed `state set` left behind (state-file.md, "The
// store starts empty"), never a second live source. A project with no
// directory yet -- nothing has ever fallen back for it -- is not an
// error: it returns a nil slice, exactly like a project with nothing
// recorded, so a caller does not have to special-case "never written" from
// "written and empty".
func ListStateFileNames(projectKey string) ([]string, error) {
	dir := filepath.Join(StateRoot(), projectKey)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("fallback: list state directory %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if name, ok := strings.CutSuffix(e.Name(), ".json"); ok {
			names = append(names, name)
		}
	}
	return names, nil
}

// ReadStateFile returns the raw bytes of the on-disk state file at path,
// exactly as last written -- the CLI never parses or reshapes this content,
// per the state-file contract's own "never rebuilt by inference" rule.
func ReadStateFile(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fallback: read state file %s: %w", path, err)
	}
	return body, nil
}

// WriteStateFile writes body to path verbatim, creating its parent
// directory if needed. body is the whole record exactly as the caller
// supplied it to `state set` -- this function performs no field-level
// merge, matching the contract's own whole-object write rule.
func WriteStateFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("fallback: create state directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("fallback: write state file %s: %w", path, err)
	}
	return nil
}
