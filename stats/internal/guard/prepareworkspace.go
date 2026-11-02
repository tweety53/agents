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

// The contract below moved here from the header of the deleted
// scripts/prepare-workspace.sh shim (KAN-1058).
//
// flow-guard prepare-workspace — validate a worktree's `## workspace isolation`
// declaration and export the workspace variables it derives.
//
// Usage: flow-guard prepare-workspace <worktree>
//
// This is exactly what `/flow`'s implement phase used to do by hand, in prose:
// run flow-guard check-workspace-isolation against the worktree first — fail loudly
// and stop if it fails — then derive and export the workspace variables the
// project's `## workspace isolation` section declares, resolved against the
// workspace id. Every rule this script applies is stated once, and canonical,
// elsewhere: the id derivation and what it derives under **The workspace id**
// and **What the id derives** (skills/flow-contracts/workspace-isolation.md),
// and the `## workspace isolation` row shapes — the four `In a workspace`
// forms, the `<id>`, `<id_underscored>` and `<value:VARIABLE>` tokens — under
// **Project configuration — workspace isolation** (skills/flow-contracts/project-configuration-isolation.md).
// This script re-derives none of that reasoning; it applies the rule.
//
// Prints one `KEY=value` line per exported variable to stdout, one per
// declared row that carries a value (see the cache-index exception below), so
// a caller reads exactly what was exported without re-deriving anything
// itself.
//
// THE WORKSPACE ID NEEDS THE CHANGE NAME, AND THIS SCRIPT TAKES ONLY A
// WORKTREE. The id is derived from the change name and from nothing else, per
// **The workspace id**, so this script reads the name from the worktree's own
// branch: every apply worktree is created on `spectre/<name>`, per
// skills/flow/implement.md, and that branch is the one place the name is
// already recorded where a script can read it without being handed it
// separately. A self-review or pipeline-fix loop worktree is its own change,
// named after its whole branch — `self-review-<name>`, `pipeline-fix-<name>-<k>`
// — by flow-guard pipeline-fix, so that whole branch is its name (KAN-997). A
// worktree on none of these — detached HEAD, or checked out on something else —
// cannot be resolved, and this script refuses rather than guessing at a name.
//
// THE CACHE INDEX IS THE ONE ROW THIS SCRIPT DOES NOT EXPORT. Per **The cache
// index** (skills/flow-contracts/workspace-isolation.md), a `cache index`
// row is claimed by probing the project's own cache, not derived from the id —
// and the registry in skills/flow-contracts/artifacts-registry.md names `/flow`,
// by probing, as what claims it "when it exports the workspace's variables".
// Probing means holding a client for whatever cache technology the project
// actually runs, which is exactly the kind of project-specific knowledge a
// script shipped into every project alike cannot carry. So a declared
// `cache index` row is reported by name on stderr — never silently dropped —
// and exports nothing; claiming it stays the operator's or the agent's own
// step, done against the real service, same as it was before this script
// existed.
//
// Exit 0 when the section is absent (no-op: nothing exported, nothing
// printed) or when isolation resolved and exported cleanly. When
// flow-guard check-workspace-isolation itself fails, this script's exit code is the
// guard's own — 1 (a dropped row) or 2 (the guard could not answer at all) —
// and the guard's stdout is relayed verbatim before exiting; its stderr is
// inherited, not captured, so it is already in front of the operator by the
// time this script's own exit code lands. Exit 2 additionally covers a
// worktree this script cannot derive a change name from, and one whose
// flow-guard check-workspace-isolation cannot be found.
//
// A project declaring no `## workspace isolation` section is the ordinary
// case for a repository with no runnable application, and is never reported
// as a misconfiguration, per **The empty id**
// (skills/flow-contracts/workspace-isolation.md).
//
// Ported to Go (KAN-873): the body's reasoning for every step is in
// stats/internal/guard/prepareworkspace.go, beside the code. The guard still
// execs flow-guard check-workspace-isolation from beside this script, as the bash
// did: FLOW_GUARD_WORKSPACE_ISOLATION is exported as its path, so a missing or
// non-executable sibling keeps its exit 2.

func init() { Registry["prepare-workspace"] = prepareWorkspace }

// pwHeading is the bash's `grep -qiE '^##[[:space:]]+workspace
// isolation[[:space:]]*$'`, matched against the ASCII-lowered line.
var pwHeading = regexp.MustCompile(`^##[ \t\n\v\f\r]+workspace isolation[ \t\n\v\f\r]*$`)

// pwRow is one resource row of the `## workspace isolation` table, as the
// guard's `#ROW` line reported it, plus the value this guard derives for it.
type pwRow struct{ res, variable, def, cell, value string }

// prepareWorkspace is `flow-guard prepare-workspace`. Its contract —
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

	// The sibling guard, check-workspace-isolation, runs in this same binary
	// (runSibling); FLOW_GUARD_WORKSPACE_ISOLATION, when set, names an
	// executable to run instead. Executable as well as a regular file: a
	// present-but-non-executable override is not runnable either, and letting
	// the exec below hit it directly would fail with a raw exit 126
	// ("Permission denied") — outside this guard's own 0/1/2 exit contract and
	// undocumented as such. Treated the same as "cannot find it at all", since
	// neither case leaves this guard able to validate anything.
	if guard := env.Getenv("FLOW_GUARD_WORKSPACE_ISOLATION"); guard != "" {
		if st, err := os.Stat(guard); err != nil || !st.Mode().IsRegular() || syscall.Access(guard, 1) != nil {
			fmt.Fprintf(stderr, "prepare-workspace: cannot find a runnable check-workspace-isolation.sh in %s\n", filepath.Dir(guard))
			return 2
		}
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
	cmd := siblingCmd(env, "FLOW_GUARD_WORKSPACE_ISOLATION", "check-workspace-isolation", worktree)
	cmd.Stderr = stderr
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
	// worktree is created on `spectre/<name>` per skills/flow/implement.md; a
	// self-review or pipeline-fix loop worktree is its own change, named after
	// its whole branch — `self-review-<name>`, `pipeline-fix-<name>-<k>` — by
	// flow-guard pipeline-fix (KAN-997). The id is derived from the change name
	// and nothing else, per **The workspace id**
	// (skills/flow-contracts/workspace-isolation.md).
	branch, _ := gitExec(env.Dir, nil, io.Discard, "-C", worktree, "symbolic-ref", "--quiet", "--short", "HEAD")
	branch = strings.TrimRight(branch, "\n")
	name, ok := strings.CutPrefix(branch, "spectre/")
	if !ok && (strings.HasPrefix(branch, "self-review-") || strings.HasPrefix(branch, "pipeline-fix-")) {
		name, ok = branch, true
	}
	if !ok {
		got := branch
		if got == "" {
			got = "detached HEAD"
		}
		fmt.Fprintf(stderr, "prepare-workspace: %s is not on a spectre/<name>, self-review-<name> or pipeline-fix-<name>-<k> branch (got '%s') — cannot derive a workspace id without the change name\n", worktree, got)
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
