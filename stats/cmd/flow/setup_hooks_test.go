package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// registrationCount is how many lines registerHooks prints per run besides its heading: one per
// hook entry plus the permission.
func registrationCount() int {
	n := 0
	for _, c := range hookChecks {
		if c.permission != "" {
			n++
		}
		n += len(c.regs)
	}
	return n
}

func runRegister(t *testing.T, path string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := registerHooks(path, &out)
	return out.String(), err
}

// assertAllRegistered checks that the file now satisfies every registration check, the same
// check `flow setup global` prints a snippet for when it fails.
func assertAllRegistered(t *testing.T, path string) {
	t.Helper()
	var out bytes.Buffer
	s := &installer{stdout: &out, stderr: &out}
	for _, c := range hookChecks {
		s.warnUnregisteredHook(path, c)
	}
	if out.Len() != 0 {
		t.Fatalf("after --register-hooks the checks still warn:\n%s", out.String())
	}
}

func backups(t *testing.T, dir string) []string {
	t.Helper()
	m, _ := filepath.Glob(filepath.Join(dir, "settings.json.*.bak"))
	return m
}

func TestRegisterHooksIntoMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	out, err := runRegister(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(out, "  ✓ added "); got != registrationCount() {
		t.Fatalf("want %d added lines, got %d:\n%s", registrationCount(), got, out)
	}
	assertAllRegistered(t, path)
	if b := backups(t, filepath.Dir(path)); len(b) != 0 {
		t.Fatalf("no file existed, so no backup is due; got %v", b)
	}
}

func TestRegisterHooksIntoEmptyObject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := runRegister(t, path); err != nil {
		t.Fatal(err)
	}
	assertAllRegistered(t, path)
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode not kept: %v", fi.Mode())
	}
	b := backups(t, filepath.Dir(path))
	if len(b) != 1 {
		t.Fatalf("want one backup, got %v", b)
	}
	if got, _ := os.ReadFile(b[0]); string(got) != "{}\n" {
		t.Fatalf("backup is not the old file: %q", got)
	}
}

func TestRegisterHooksKeepsForeignKeysAndHooks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	in := `{
  "model": "opus",
  "env": {"A": "1", "N": 1.50},
  "permissions": {"allow": ["Bash(ls *)"], "deny": ["Bash(rm *)"]},
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "mine.sh", "timeout": 5}]},
      {"matcher": "Agent|Task", "hooks": [{"type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook enforce-loop-dispatch || true"}]}
    ],
    "Stop": [{"hooks": [{"type": "command", "command": "say done"}]}]
  },
  "zlast": true
}`
	if err := os.WriteFile(path, []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runRegister(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `  ✓ set "timeout": 300 on PreToolUse hook enforce-loop-dispatch`) {
		t.Fatalf("the timeout-less loop hook was not fixed:\n%s", out)
	}
	assertAllRegistered(t, path)
	got, _ := os.ReadFile(path)
	s := string(got)
	for _, want := range []string{`"model": "opus"`, `"N": 1.50`, `"Bash(ls *)"`, `"deny": [`, `"command": "mine.sh"`, `"timeout": 5`, `"command": "say done"`, `"zlast": true`} {
		if !strings.Contains(s, want) {
			t.Errorf("lost %s:\n%s", want, s)
		}
	}
	if strings.Index(s, `"model"`) > strings.Index(s, `"env"`) || strings.Index(s, `"hooks"`) > strings.Index(s, `"zlast"`) {
		t.Errorf("key order not kept:\n%s", s)
	}
	if strings.Count(s, "enforce-loop-dispatch") != 1 {
		t.Errorf("the present loop hook was registered twice:\n%s", s)
	}
}

func TestRegisterHooksTwiceIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if _, err := runRegister(t, path); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)
	out, err := runRegister(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(out, " already present\n"); got != registrationCount() {
		t.Fatalf("want %d already-present lines, got %d:\n%s", registrationCount(), got, out)
	}
	second, _ := os.ReadFile(path)
	if !bytes.Equal(first, second) {
		t.Fatalf("second run changed the file:\n%s\n---\n%s", first, second)
	}
	if b := backups(t, filepath.Dir(path)); len(b) != 0 {
		t.Fatalf("an unchanged file needs no backup; got %v", b)
	}
}

func TestRegisterHooksRefusesMalformed(t *testing.T) {
	for name, body := range map[string]string{
		"syntax":     `{"hooks": {`,
		"not-object": `[]`,
		"hooks-type": `{"hooks": []}`,
		"event-type": `{"hooks": {"Stop": {}}}`,
		"allow-type": `{"permissions": {"allow": "x"}}`,
		"trailing":   `{} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			if checkSettings(path) == nil {
				t.Fatal("checkSettings accepted a file registerHooks cannot merge into")
			}
			if _, err := runRegister(t, path); err == nil {
				t.Fatal("registerHooks accepted it")
			}
			if got, _ := os.ReadFile(path); string(got) != body {
				t.Fatalf("file modified: %q", got)
			}
			ents, _ := os.ReadDir(filepath.Dir(path))
			if len(ents) != 1 {
				t.Fatalf("refusal left files behind: %v", ents)
			}
		})
	}
}

// TestRunSetupRegisterHooksRefusesMalformedBeforeInstalling: the refusal is the global preflight,
// so a malformed settings.json leaves the whole home untouched and exits non-zero.
func TestRunSetupRegisterHooksRefusesMalformedBeforeInstalling(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".claude/settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo, _ := filepath.Abs("../../..")
	env := map[string]string{"HOME": home, "FLOW_SETUP_REPO_ROOT": repo, "SETUP_REAL_HOME": "/nonexistent-real-home"}
	var out, errb bytes.Buffer
	code := runSetup([]string{"global", "--register-hooks"}, func(k string) string { return env[k] }, &out, &errb)
	if code == 0 {
		t.Fatalf("exit 0 on a malformed settings.json\n%s%s", out.String(), errb.String())
	}
	if !strings.Contains(errb.String(), "ERROR: "+path) {
		t.Fatalf("refusal does not name the file:\n%s", errb.String())
	}
	ents, _ := os.ReadDir(filepath.Join(home, ".claude"))
	if len(ents) != 1 {
		t.Fatalf("install steps ran before the refusal: %v", ents)
	}
}

// TestHookChecksMatchTheirSnippets: the merge writes from regs, the warning prints the snippet,
// so each reg's command, matcher and timeout must be the ones its snippet shows.
func TestHookChecksMatchTheirSnippets(t *testing.T) {
	for _, c := range hookChecks {
		if c.permission != "" {
			if !strings.Contains(c.snippet, `"`+c.permission+`"`) {
				t.Errorf("%s: snippet lacks %q", c.label, c.permission)
			}
			continue
		}
		if !strings.Contains(c.snippet, strings.ReplaceAll(hookCommand(c.identifier), `"`, `\"`)) {
			t.Errorf("%s: snippet lacks the command", c.label)
		}
		for _, r := range c.regs {
			if r.matcher != "" && !strings.Contains(c.snippet, `"matcher": "`+r.matcher+`"`) {
				t.Errorf("%s: snippet lacks matcher %q", c.label, r.matcher)
			}
			hasT := strings.Contains(c.snippet, `"timeout"`)
			if (r.timeout > 0) != hasT || (hasT && !strings.Contains(c.snippet, `"timeout": `+strconv.Itoa(r.timeout))) {
				t.Errorf("%s: snippet and reg disagree on timeout %d", c.label, r.timeout)
			}
		}
	}
}
