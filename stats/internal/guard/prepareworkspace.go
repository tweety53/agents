package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func init() { Registry["prepare-workspace"] = prepareWorkspace }

// pwHeading is the bash's `grep -qiE '^##[[:space:]]+workspace
// isolation[[:space:]]*$'`, matched against the ASCII-lowered line.
var pwHeading = regexp.MustCompile(`^##[ \t\n\v\f\r]+workspace isolation[ \t\n\v\f\r]*$`)

// pwRow is one resource row of the `## workspace isolation` table, as the
// guard's `#ROW` line reported it, plus the value this guard derives for it.
type pwRow struct{ res, variable, def, cell, value string }

// prepareWorkspace is scripts/prepare-workspace.sh's body. Its contract —
// arguments, the KEY=value lines, the cache-index exception and the exit
// codes — is the script's header; the reasoning behind each step is here.
func prepareWorkspace(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprint(stderr, "Usage: prepare-workspace.sh <worktree>\n")
		return 2
	}
	worktree := args[0]
	abs := worktree
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(env.Dir, abs)
	}
	// `[ -d "" ]` is false: an empty argument is never the cwd.
	if st, err := os.Stat(abs); worktree == "" || err != nil || !st.IsDir() {
		fmt.Fprintf(stderr, "prepare-workspace: %s is not a directory\n", worktree)
		return 2
	}

	// The sibling guard sits beside the shim: the bash took it from its own
	// $SCRIPT_DIR, and the shim exports that path as
	// FLOW_GUARD_WORKSPACE_ISOLATION, whatever the directory is named.
	guard := env.Getenv("FLOW_GUARD_WORKSPACE_ISOLATION")
	if guard == "" {
		fmt.Fprint(stderr, "prepare-workspace: FLOW_GUARD_WORKSPACE_ISOLATION is unset — run scripts/prepare-workspace.sh, which sets it\n")
		return 2
	}
	scriptDir := filepath.Dir(guard)
	// Executable as well as a regular file: a present-but-non-executable
	// guard is not runnable either, and letting the exec below hit it
	// directly would fail with a raw exit 126 ("Permission denied") — outside
	// this guard's own 0/1/2 exit contract and undocumented as such. Treated
	// the same as "cannot find it at all", since neither case leaves this
	// guard able to validate anything.
	if st, err := os.Stat(guard); err != nil || !st.Mode().IsRegular() || syscall.Access(guard, 1) != nil {
		fmt.Fprintf(stderr, "prepare-workspace: cannot find a runnable check-workspace-isolation.sh in %s\n", scriptDir)
		return 2
	}

	// Validate first, with the guard, not by eye — and before resolving a
	// single row. Its stdout is captured so a violation report can be
	// relayed after this guard's own exit code is decided; its stderr is
	// inherited, so a refusal it prints is already visible.
	//
	// CHECK_WORKSPACE_ISOLATION_PRINT_ROWS=1 asks the guard to append each
	// validated resource row to its own report, as a `#ROW` line, once it has
	// finished validating — see workspaceisolation.go's comment on that
	// variable. This is the ONE parse of `.flow/project.md`'s resource table
	// this guard needs: the guard already walked the file's cells to validate
	// them, and the rows below are extracted from that same walk rather than
	// a second one this guard would otherwise have to run itself.
	cmd := exec.Command(guard, worktree)
	cmd.Dir, cmd.Stderr = env.Dir, stderr
	cmd.Env = append(os.Environ(), "CHECK_WORKSPACE_ISOLATION_PRINT_ROWS=1")
	raw, err := cmd.Output()
	// $(...) strips every trailing newline; the relay below adds one back.
	guardOut := strings.TrimRight(string(raw), "\n")
	if rc := pwExitCode(err); rc != 0 {
		if guardOut != "" {
			fmt.Fprintf(stdout, "%s\n", guardOut)
		}
		return rc
	}

	// No `.flow/project.md`, or one declaring no `## workspace isolation`
	// section, is a no-op: nothing exported, nothing printed. The guard above
	// already confirmed this file — if present — is well formed — so the only
	// question left here is whether the section exists at all, exactly
	// mirroring the guard's own no-op case for a project with no such
	// section. An unreadable file reads as no section, as the bash's
	// `grep -q` failing did.
	cfg, err := os.ReadFile(filepath.Join(abs, ".flow", "project.md"))
	if err != nil {
		return 0
	}
	found := false
	for _, l := range strings.Split(string(cfg), "\n") {
		if pwHeading.MatchString(ccASCIILower(l)) {
			found = true
			break
		}
	}
	if !found {
		return 0
	}

	// The change name, read from the worktree's own branch. Every apply
	// worktree is created on `spectre/<name>` per skills/flow/implement.md,
	// and the id is derived from the change name and nothing else, per **The
	// workspace id** (skills/flow-contracts/workspace-isolation.md).
	branch, _ := gitExec(env.Dir, nil, io.Discard, "-C", worktree, "symbolic-ref", "--quiet", "--short", "HEAD")
	branch = strings.TrimRight(branch, "\n")
	name, ok := strings.CutPrefix(branch, "spectre/")
	if !ok {
		got := branch
		if got == "" {
			got = "detached HEAD"
		}
		fmt.Fprintf(stderr, "prepare-workspace: %s is not on a spectre/<name> branch (got '%s') — cannot derive a workspace id without the change name\n", worktree, got)
		return 2
	}

	// The workspace id — prefix and digest joined by '-' — derived exactly
	// as **The workspace id** (skills/flow-contracts/workspace-isolation.md)
	// states: ccWorkspaceID, the derivation check-cleanup-complete already
	// carries, locale-independent by construction.
	id := ccWorkspaceID(name)
	idUnderscored := strings.ReplaceAll(id, "-", "_")
	digest, _ := strconv.ParseUint(id[len(id)-4:], 16, 64)
	offset := int64(digest%400+1) * 10

	// The resource table's rows, already parsed and validated — extracted
	// from the guard's own report rather than re-reading `.flow/project.md`.
	// Every line of its stdout whose first tab-separated field is `#ROW` is
	// one resource row the guard walked while validating it; the four cells
	// after the marker are `Resource`, `Variable`, `Default` and `In a
	// workspace`. The bash read them with `IFS=$'\t' read`, under which a run
	// of tabs is one separator, so an empty cell shifts the ones after it —
	// strings.FieldsFunc splits the same way. A row with no variable is
	// skipped.
	var rows []pwRow
	for _, line := range strings.Split(guardOut, "\n") {
		f := strings.Split(line, "\t")
		if f[0] != "#ROW" {
			continue
		}
		cells := strings.FieldsFunc(strings.Join(f[1:min(5, len(f))], "\t"), func(r rune) bool { return r == '\t' })
		cells = append(cells, "", "", "", "")
		if cells[1] == "" {
			continue
		}
		rows = append(rows, pwRow{res: cells[0], variable: cells[1], def: cells[2], cell: cells[3]})
	}

	// substituteID — <id> and <id_underscored>, the two tokens every
	// `database`, `bucket` and `url` cell may carry, per **What the id
	// derives** (skills/flow-contracts/workspace-isolation.md). The two
	// tokens never overlap as literal substrings, so the order they are
	// replaced in does not matter.
	substituteID := func(cell string) string {
		cell = strings.ReplaceAll(cell, "<id_underscored>", idUnderscored)
		return strings.ReplaceAll(cell, "<id>", id)
	}

	// Pass 1 — every row a `url` row may reference: `database`, `bucket`,
	// `port`. `cache index` is resolved in this same pass in the sense that
	// it is recognised and reported; it never receives a value, and no `url`
	// row may reference it (the guard already refuses one that tries).
	for i := range rows {
		r := &rows[i]
		switch r.res {
		case "database", "bucket":
			r.value = substituteID(r.cell)
		case "port":
			// Parsed base 10. The bash had to force `10#` because its
			// arithmetic reads a leading-zero literal as octal — a declared
			// default of `0070` would silently become the wrong port (`0070 +
			// 20` is `76`, not `90`), and `0080` is not even valid octal. The
			// guard's own `Default` rule for a `port` row is only "a bare
			// integer" (`^[0-9]+$`), which admits a leading zero, so this
			// guard — not the table's author — is the one place that has to
			// fix the base.
			n, err := strconv.ParseInt(r.def, 10, 64)
			if err != nil {
				fmt.Fprintf(stderr, "prepare-workspace: `%s` (port) default %q is not a base-10 integer\n", r.variable, r.def)
				return 1
			}
			r.value = strconv.FormatInt(n+offset, 10)
		case "cache index":
			fmt.Fprintf(stderr, "prepare-workspace: `%s` (cache index) is claimed by probing the project's own cache, not derived from the workspace id — per the registry in skills/flow-contracts/artifacts-registry.md, `/flow` claims it, by probing, when it exports the workspace's variables. This script does not carry a client for the project's cache, so this row is reported rather than exported; claim it against the real service before anything reads `%s`.\n", r.variable, r.variable)
		}
	}

	// Pass 2 — `url` rows, which may reference the values pass 1 just
	// resolved.
	for i := range rows {
		if rows[i].res != "url" {
			continue
		}
		v, err := pwResolveValueRefs(substituteID(rows[i].cell), rows)
		if err != nil {
			fmt.Fprintf(stderr, "prepare-workspace: %v\n", err)
			return 2
		}
		rows[i].value = v
	}

	// Print, in declaration order — the same order the project wrote the
	// table in, so a KEY=value line's position in the section is the same
	// position it prints in. The bash also exported each one, into a process
	// that exited straight after; the printed lines were always the whole of
	// what a caller received.
	for _, r := range rows {
		if r.res == "cache index" {
			continue
		}
		fmt.Fprintf(stdout, "%s=%s\n", r.variable, r.value)
	}
	return 0
}

// pwResolveValueRefs resolves every `<value:VARIABLE>` reference in a `url`
// cell against the already-computed value of the row named VARIABLE. The
// guard has already confirmed every reference names a `database`, `bucket`
// or `port` row in this same table, so this is resolution, not
// re-validation — but it is checked again here rather than trusted blindly,
// because a resolution bug that only ever checked the NAME would be silent
// rather than loud: `database`, `bucket` and `port` rows are the only ones
// pass 1 gives a value before this runs, so a reference resolving to any
// other kind — a `url` row (forbidden per **What a `url` row may reference,
// and what it may not**, skills/flow-contracts/project-configuration-isolation.md
// — a `url` row may never reference another `url` row, so a table read
// cleanly by definition never produces one) or a `cache index` row (never
// given a value at all, per **The cache index**) — would still find the
// VARIABLE by name and substitute whatever value it currently holds: the
// empty string every row starts with. That is a wrong value printed with no
// error at all, so the kind is checked here too, and a reference this guard
// cannot resolve stops the run rather than printing an empty substitution.
func pwResolveValueRefs(cell string, rows []pwRow) (string, error) {
	for {
		start := strings.Index(cell, "<value:")
		if start < 0 || !strings.Contains(cell[start+len("<value:"):], ">") {
			return cell, nil
		}
		rest := cell[start+len("<value:"):]
		name := rest[:strings.Index(rest, ">")]
		token := "<value:" + name + ">"
		val := ""
		for _, r := range rows {
			if r.variable != name {
				continue
			}
			switch r.res {
			case "database", "bucket", "port":
			default:
				return "", fmt.Errorf("<value:%s> resolves to `%s`, a `%s` row — a reference may only name a `database`, `bucket` or `port` row, per **What a `url` row may reference, and what it may not** (skills/flow-contracts/project-configuration-isolation.md) — refusing rather than substituting an empty or stale value", name, r.variable, r.res)
			}
			val = r.value
			break
		}
		// A reference naming no row leaves val empty: the guard already
		// refuses one, so that empty substitution is unreachable on a
		// validated file — kept anyway rather than looping forever on a token
		// that never resolves.
		cell = strings.Replace(cell, token, val, 1)
	}
}

// pwExitCode is the `$?` bash saw from `$(guard ...)`: the guard's status,
// 128+n when signal n killed it, 126 when it could not be executed and 127
// when it could not be found — the two exec failures racing the runnable
// check above.
func pwExitCode(err error) int {
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return rrExitCode(ee.ProcessState)
	case errors.Is(err, os.ErrNotExist):
		return 127
	default:
		return 126
	}
}
