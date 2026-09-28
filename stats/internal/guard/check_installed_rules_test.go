package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-check-installed-rules.sh at c5379c0a, one
// subtest per ok: label, plus a collation row (Review Focus). The harness
// grepped the combined output for a needle; each case here pins the exit
// code, stdout and stderr whole, as the bash printed them at c5379c0a.
//
// Like the harness, every case runs against this repository's own rules/
// and, unless a case overrides it, this checkout's own setup.sh — whose
// managed_files declaration names ~/.claude/CLAUDE.md and ~/.zcode/AGENTS.md,
// both of which every fixture home carries. No case
// reads or writes the operator's real $HOME: CHECK_INSTALLED_RULES_HOME (or,
// for the worktree case, HOME itself) names a t.TempDir().

// cirRoot is this checkout, the root the shim exports.
var cirRoot = sync.OnceValues(func() (string, error) { return filepath.Abs("../../..") })

// cirRules is the harness's always_on: the always-on rule basenames, found
// by running the harness's own awk over rules/*.mdc under bash's glob order
// — an oracle independent of the Go port of that awk.
var cirRules = sync.OnceValues(func() ([]string, error) {
	root, err := cirRoot()
	if err != nil {
		return nil, err
	}
	out, err := exec.Command("bash", "-c", `for f in "$1"/rules/*.mdc; do
  [ -f "$f" ] || continue
  if awk '
    { sub(/\r$/, "") }
    NR == 1 && $0 != "---" { exit }
    NR > 1  && $0 == "---" { closed = 1; exit }
    NR > 20 { exit }
    NR > 1  && /^alwaysApply:[[:space:]]*true[[:space:]]*$/ { always = 1 }
    END { exit !(always && closed) }
  ' "$f"; then basename "$f"; fi
done`, "bash", root).Output()
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(out)), nil
})

// cirHome is new_home: a complete, known-good fixture install.
func cirHome(t *testing.T, root string, rules []string) string {
	t.Helper()
	home := t.TempDir()
	mkdir(t, home+"/.claude/rules")
	mkdir(t, home+"/.zcode")
	for _, r := range rules {
		cirLink(t, root+"/rules/"+r, home+"/.claude/rules/"+strings.TrimSuffix(r, ".mdc")+".md")
	}
	cirLink(t, root+"/rules/agent-baseline.md", home+"/.claude/rules/agent-baseline.md")
	cirBlock(t, home+"/.claude/CLAUDE.md", rules)
	cirBlock(t, home+"/.zcode/AGENTS.md", rules)
	return home
}

func cirLink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// cirBlock is render_block: a managed block carrying one marker per rule.
func cirBlock(t *testing.T, path string, rules []string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("# fixture\n")
	for _, r := range rules {
		b.WriteString("\n<!-- rule: " + r + " -->\n\nbody\n")
	}
	writeFile(t, path, b.String())
}

// cirDropMarker is the harness's `sed -i.bak "/<!-- rule: X -->/d"`.
func cirDropMarker(t *testing.T, path, rule string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var keep []string
	for _, l := range strings.SplitAfter(string(b), "\n") {
		if !strings.Contains(l, "<!-- rule: "+rule+" -->") {
			keep = append(keep, l)
		}
	}
	writeFile(t, path, strings.Join(keep, ""))
}

func TestCheckInstalledRules(t *testing.T) {
	t.Parallel()
	root, err := cirRoot()
	if err != nil {
		t.Fatal(err)
	}
	rules, err := cirRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) < 2 {
		t.Fatalf("need at least two always-on rules to test with, got %v", rules)
	}
	first := rules[0]
	firstMD := strings.TrimSuffix(first, ".mdc") + ".md"
	n := strconv.Itoa(len(rules))

	const p = "check-installed-rules: "
	stale := func(home string, v int) string {
		return "INSTALLED-RULES-STALE: " + home + " — " + strconv.Itoa(v) + " violation(s); re-run ./setup.sh global from " + root + "\n"
	}
	unrendered := func(file, rule string) string {
		return p + file + " carries no managed-block marker for " + rule + " — no session reads that rule\n"
	}

	type result struct {
		code           int
		stdout, stderr string
	}
	type fixture struct {
		root, home string
		vars       map[string]string // on top of CHECK_INSTALLED_RULES_HOME=home
		noOverride bool              // CHECK_INSTALLED_RULES_HOME unset; HOME=home
		noHome     bool              // with noOverride: HOME unset too
	}
	for _, tc := range []struct {
		name string
		// build breaks a known-good install and returns what the bash printed.
		build func(t *testing.T) (fixture, result)
	}{
		{"clean: a complete install passes", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h}, result{0,
				"INSTALLED-RULES-OK: " + h + " — " + n + " always-on rule(s) installed and rendered\n", ""}
		}},
		{"missing_link: an always-on rule that was never installed", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			installed := h + "/.claude/rules/" + firstMD
			if err := os.Remove(installed); err != nil {
				t.Fatal(err)
			}
			cirDropMarker(t, h+"/.claude/CLAUDE.md", first)
			cirDropMarker(t, h+"/.zcode/AGENTS.md", first)
			return fixture{root: root, home: h}, result{1, stale(h, 3),
				p + installed + " is missing — rules/" + first + " is always-on but was never installed\n" +
					unrendered(h+"/.claude/CLAUDE.md", first) + unrendered(h+"/.zcode/AGENTS.md", first)}
		}},
		{"copy_not_link: a copied rule file", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			installed := h + "/.claude/rules/" + firstMD
			b, err := os.ReadFile(root + "/rules/" + first)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(installed); err != nil {
				t.Fatal(err)
			}
			writeFile(t, installed, string(b))
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + installed + " is not a symlink — a copy goes stale the next time the rule is edited\n"}
		}},
		{"dangling: a link whose target is gone", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			installed := h + "/.claude/rules/" + firstMD
			if err := os.Remove(installed); err != nil {
				t.Fatal(err)
			}
			cirLink(t, root+"/rules/no-such-rule.mdc", installed)
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + installed + " is a dangling symlink — it points at " + root + "/rules/no-such-rule.mdc, which does not exist\n"}
		}},
		{"foreign_checkout: a link resolving into another tree", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			other := t.TempDir()
			b, err := os.ReadFile(root + "/rules/" + first)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, other+"/rules/"+first, string(b))
			installed := h + "/.claude/rules/" + firstMD
			if err := os.Remove(installed); err != nil {
				t.Fatal(err)
			}
			cirLink(t, other+"/rules/"+first, installed)
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + installed + " points at " + other + "/rules/" + first + ", not this checkout's " + root + "/rules/" + first + "\n"}
		}},
		{"retired: a link for a rule this checkout no longer has", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			cirLink(t, root+"/rules/"+first, h+"/.claude/rules/retired-rule.md")
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + h + "/.claude/rules/retired-rule.md is installed but rules/retired-rule.mdc is not an always-on rule here — a stale link still reads as installed\n"}
		}},
		{"no_baseline: the agent baseline is not installed", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			if err := os.Remove(h + "/.claude/rules/agent-baseline.md"); err != nil {
				t.Fatal(err)
			}
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + h + "/.claude/rules/agent-baseline.md is missing or dangling — every subagent dispatch points at it\n"}
		}},
		{"unrendered_claude: linked but absent from CLAUDE.md", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			cirDropMarker(t, h+"/.claude/CLAUDE.md", first)
			return fixture{root: root, home: h}, result{1, stale(h, 1), unrendered(h+"/.claude/CLAUDE.md", first)}
		}},
		{"unrendered_zcode: linked but absent from AGENTS.md", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			cirDropMarker(t, h+"/.zcode/AGENTS.md", first)
			return fixture{root: root, home: h}, result{1, stale(h, 1), unrendered(h+"/.zcode/AGENTS.md", first)}
		}},
		{"absent_managed_file: a declared file that does not exist is noted, not scanned", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			if err := os.Remove(h + "/.zcode/AGENTS.md"); err != nil {
				t.Fatal(err)
			}
			return fixture{root: root, home: h}, result{0,
				"INSTALLED-RULES-OK: " + h + " — " + n + " always-on rule(s) installed and rendered\n",
				p + h + "/.zcode/AGENTS.md does not exist, not scanned\n"}
		}},
		{"stale_marker: a rendered rule this checkout does not have", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			f := h + "/.claude/CLAUDE.md"
			b, _ := os.ReadFile(f)
			writeFile(t, f, string(b)+"\n<!-- rule: retired-rule.mdc -->\n\nbody\n")
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + f + " renders retired-rule.mdc, which is not an always-on rule here\n"}
		}},
		{"worktree: a linked worktree is skipped, not failed", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			wt := t.TempDir()
			mkdir(t, wt+"/scripts/lib")
			entries, err := filepath.Glob(root + "/rules/*.mdc")
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range append(entries, root+"/rules/agent-baseline.md") {
				b, err := os.ReadFile(e)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, wt+"/rules/"+filepath.Base(e), string(b))
			}
			writeFile(t, wt+"/.git", "gitdir: /somewhere/.git/worktrees/x\n")
			return fixture{root: wt, home: h, noOverride: true}, result{0,
				"INSTALLED-RULES-WORKTREE: " + wt + " is a linked worktree — the install tracks the main checkout, nothing to check here\n", ""}
		}},
		// The bash's `${CHECK_INSTALLED_RULES_HOME:-$HOME}` under set -u died
		// on an unset HOME (exit 1); reading it as empty would pass a
		// worktree or an empty install instead.
		{"port: unset HOME with no override is refused", func(t *testing.T) (fixture, result) {
			return fixture{root: root, noOverride: true, noHome: true},
				result{1, "", p + "HOME is unset and CHECK_INSTALLED_RULES_HOME is not set — cannot locate the install\n"}
		}},
		{"no_install: an absent install is reported, not failed", func(t *testing.T) (fixture, result) {
			h := t.TempDir()
			return fixture{root: root, home: h}, result{0,
				"INSTALLED-RULES-NONE: " + h + " — no global install found, nothing to check\n", ""}
		}},
		{"partial: an install missing one rule is stale, not absent", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			installed := h + "/.claude/rules/" + firstMD
			if err := os.Remove(installed); err != nil {
				t.Fatal(err)
			}
			return fixture{root: root, home: h}, result{1, stale(h, 1),
				p + installed + " is missing — rules/" + first + " is always-on but was never installed\n"}
		}},
		{"served_managed_files: a file the installer added is scanned", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/setup.sh"
			writeFile(t, setup, `install_global() {
  # The managed-block targets, declared once so the preflight below and the
  # install below can never scan a different set than they write.
  local managed_files=("$home_dir/.claude/CLAUDE.md" "$home_dir/.zcode/AGENTS.md" "$home_dir/.warp/AGENTS.md")
}
`)
			h := cirHome(t, root, rules)
			cirBlock(t, h+"/.warp/AGENTS.md", rules[1:])
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, stale(h, 1), unrendered(h+"/.warp/AGENTS.md", first)}
		}},
		{"no_declaration: a declaration-free installer is refused", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/setup.sh"
			writeFile(t, setup, "#!/usr/bin/env bash\n# no managed_files here\n")
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, "", p + "no 'local managed_files=(' declaration in " + setup + " — cannot resolve the managed-block targets\n"}
		}},
		{"commented_declaration: a commented-out declaration is not parsed", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/setup.sh"
			writeFile(t, setup, "# local managed_files=(\"$home_dir/.claude/CLAUDE.md\")\n")
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, "", p + "no 'local managed_files=(' declaration in " + setup + " — cannot resolve the managed-block targets\n"}
		}},
		{"multiline_declaration: a multi-line declaration is refused, never truncated", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/setup.sh"
			writeFile(t, setup, `  local managed_files=("$home_dir/.claude/CLAUDE.md"
    "$home_dir/.zcode/AGENTS.md")
`)
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, "", p + "the managed_files declaration in " + setup + " does not close on its own line — this guard parses a single-line declaration only, and a truncated set must be refused, never silently served\n"}
		}},
		{"empty_declaration: an empty declaration is refused", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/setup.sh"
			writeFile(t, setup, "local managed_files=()\n")
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, "", p + "setup.sh's managed_files declaration is empty in " + setup + " — refusing to scan nothing and call it clean\n"}
		}},
		{"unresolvable_element: an unresolvable element is refused", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/setup.sh"
			writeFile(t, setup, "local managed_files=(\"$home_dir/.claude/CLAUDE.md\" \"$project_dir/other.md\")\n")
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, "", p + "cannot resolve managed_files element '$project_dir/other.md' in " + setup + " — expected a \"$home_dir/-relative path\n"}
		}},
		{"unreadable_setup: an absent installer is refused", func(t *testing.T) (fixture, result) {
			setup := t.TempDir() + "/absent.sh"
			h := cirHome(t, root, rules)
			return fixture{root: root, home: h, vars: map[string]string{"CHECK_INSTALLED_RULES_SETUP_SH": setup}},
				result{1, "", p + setup + " is unreadable — cannot resolve the managed-block targets\n"}
		}},
		// Review Focus: the markers' `sort -u` keeps the caller's collation —
		// en_US.UTF-8 puts alpha before Zeta, where a byte order would not.
		{"collation: stale markers are reported in the locale's sort order", func(t *testing.T) (fixture, result) {
			h := cirHome(t, root, rules)
			f := h + "/.claude/CLAUDE.md"
			b, _ := os.ReadFile(f)
			writeFile(t, f, string(b)+"<!-- rule: Zeta.mdc -->\n<!-- rule: alpha.mdc --> <!-- rule: alpha.mdc -->\n")
			return fixture{root: root, home: h, vars: map[string]string{"LC_ALL": "en_US.UTF-8"}}, result{1, stale(h, 2),
				p + f + " renders alpha.mdc, which is not an always-on rule here\n" +
					p + f + " renders Zeta.mdc, which is not an always-on rule here\n"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx, want := tc.build(t)
			vars := map[string]string{"FLOW_GUARD_REPO_ROOT": fx.root, "HOME": t.TempDir()}
			if fx.noHome {
				delete(vars, "HOME")
			} else if fx.noOverride {
				vars["HOME"] = fx.home
			} else {
				vars["CHECK_INSTALLED_RULES_HOME"] = fx.home
			}
			for k, v := range fx.vars {
				vars[k] = v
			}
			lookup := func(k string) (string, bool) {
				if v, ok := vars[k]; ok {
					return v, true
				}
				if k == "PATH" || k == "LANG" || k == "LC_ALL" || k == "LC_COLLATE" || k == "LC_CTYPE" {
					return os.LookupEnv(k)
				}
				return "", false
			}
			getenv := func(k string) string { v, _ := lookup(k); return v }
			var stdout, stderr bytes.Buffer
			code := Registry["check-installed-rules"]
			if code == nil {
				t.Fatal("check-installed-rules is not registered")
			}
			got := code(nil, Env{Getenv: getenv, LookupEnv: lookup, Dir: t.TempDir()}, &stdout, &stderr)
			if got != want.code || stdout.String() != want.stdout || stderr.String() != want.stderr {
				t.Errorf("exit %d, want %d\nstdout:\n%s\nwant:\n%s\nstderr:\n%s\nwant:\n%s",
					got, want.code, stdout.String(), want.stdout, stderr.String(), want.stderr)
			}
		})
	}
}
