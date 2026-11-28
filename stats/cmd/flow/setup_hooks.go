package main

// The hook half of `flow setup global`: the hook binary build, the registration checks and the
// snippets they print, and `--register-hooks`, which merges those registrations into
// ~/.claude/settings.json instead.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// hookReg is one registration a check's snippet describes: a command entry under event, in a
// group carrying matcher ("" for none), with timeout (0 for none).
type hookReg struct {
	event, matcher string
	timeout        int
}

// hookCheck is one registration check: identifier is what a registered entry's command (or, for
// the permission, the allow list) contains; warning lines name the settings file as {settings}.
type hookCheck struct {
	identifier, label string
	warning           []string
	snippet           string
	permission        string    // a permissions.allow entry, for the one check that is not a hook
	regs              []hookReg // the hook entries the snippet registers
}

func hookCommand(name string) string {
	return `"$HOME/.claude/hooks/flow-guard-hook" hook ` + name + " || true"
}

var sessionActivityEvents = []string{"PreToolUse", "PostToolUse", "Notification", "UserPromptSubmit", "Stop", "SessionEnd"}

var hookChecks = func() []hookCheck {
	var activity []hookReg
	for _, e := range sessionActivityEvents {
		activity = append(activity, hookReg{event: e})
	}
	return []hookCheck{
		{identifier: "flow-guard sync-base-behind", label: "sync-base-behind-permission",
			warning: []string{
				"  ⚠ The base-sync permission is not allowed, so the auto-mode classifier can stall /flow at",
				"    implementation start. Add this to \"permissions.allow\" in {settings}:"},
			snippet: `
    "Bash(flow-guard sync-base-behind *)"`,
			permission: "Bash(flow-guard sync-base-behind *)"},
		{identifier: "enforce-agent-baseline", label: "agent-baseline",
			warning: []string{
				"  ⚠ The agent-baseline hook is installed but NOT registered, so nothing yet enforces",
				"    that subagent dispatches carry the rules. Add this to \"hooks\" in {settings}:"},
			snippet: `
    "PreToolUse": [
      {
        "matcher": "Agent|Task",
        "hooks": [
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook enforce-agent-baseline || true" }
        ]
      }
    ]`,
			regs: []hookReg{{event: "PreToolUse", matcher: "Agent|Task"}}},
		{identifier: "enforce-loop-dispatch", label: "enforce-loop-dispatch",
			warning: []string{
				"  ⚠ The enforce-loop-dispatch hook is installed but NOT registered, so nothing yet admits",
				"    a pipeline-fix or self-review loop dispatch off its pairs row, or records it.",
				"    Add this entry to \"PreToolUse\" in {settings}:"},
			snippet: `
      {
        "matcher": "Agent|Task",
        "hooks": [
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook enforce-loop-dispatch || true", "timeout": 300 }
        ]
      }`,
			regs: []hookReg{{event: "PreToolUse", matcher: "Agent|Task", timeout: 300}}},
		{identifier: "flow-active-change", label: "flow-active-change",
			warning: []string{
				"  ⚠ The flow-active-change hook is installed but NOT registered, so a plain problem",
				"    report is never named as a fix of the open change. Add this to \"hooks\" in {settings}:"},
			snippet: `
    "UserPromptSubmit": [
      {
        "hooks": [
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook flow-active-change || true" }
        ]
      }
    ]`,
			regs: []hookReg{{event: "UserPromptSubmit"}}},
		{identifier: "flow-session-activity", label: "flow-session-activity",
			warning: []string{
				"  ⚠ The flow-session-activity hook is installed but NOT registered, so the Live state",
				"    board never shows a session waiting on you. Add one entry per event to \"hooks\" in",
				"    {settings} — PreToolUse, PostToolUse,",
				"    Notification, UserPromptSubmit, Stop and SessionEnd without a matcher — each running:"},
			snippet: `
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook flow-session-activity || true" }`,
			regs: activity},
		{identifier: "flow-skill-load", label: "flow-skill-load",
			warning: []string{
				"  ⚠ The flow-skill-load hook is installed but NOT registered, so flowd records no skill",
				"    file a session loads and prune-candidates cannot list the ones never loaded.",
				"    Add this entry to \"PostToolUse\" in {settings}:"},
			snippet: `
      {
        "matcher": "Read|Skill",
        "hooks": [
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook flow-skill-load || true" }
        ]
      }`,
			regs: []hookReg{{event: "PostToolUse", matcher: "Read|Skill"}}},
		{identifier: "await-own-agents", label: "await-own-agents",
			warning: []string{
				"  ⚠ The await-own-agents hook is installed but NOT registered, so a subagent may still",
				"    end its turn while its own agents run and lose its prompt cache waiting on them.",
				"    Add this to \"hooks\" in {settings}:"},
			snippet: `
    "SubagentStop": [
      {
        "hooks": [
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook await-own-agents || true" }
        ]
      }
    ]`,
			regs: []hookReg{{event: "SubagentStop"}}},
		{identifier: "block-unsafe-commands", label: "block-unsafe-commands",
			warning: []string{
				"  ⚠ The block-unsafe-commands hook is installed but NOT registered, so nothing yet stops",
				"    an agent editing, staging or committing in a main checkout on its default branch.",
				"    Add this entry to \"PreToolUse\" in {settings}:"},
			snippet: `
      {
        "matcher": "Edit|Write|MultiEdit|NotebookEdit|Bash",
        "hooks": [
          { "type": "command", "command": "\"$HOME/.claude/hooks/flow-guard-hook\" hook block-unsafe-commands || true", "timeout": 300 }
        ]
      }`,
			regs: []hookReg{{event: "PreToolUse", matcher: "Edit|Write|MultiEdit|NotebookEdit|Bash", timeout: 300}}},
	}
}()

// installHooks builds the hook binary into <claudeDir>/hooks and checks each registration in
// settings.json, printing the snippet for any it lacks. settings.json is the user's own file, so it
// is edited only on --register-hooks, which merges the registrations in first (and the checks then
// print nothing).
func (s *installer) installHooks(claudeDir string, register bool) {
	s.info("Installing hooks into %s/hooks", claudeDir)
	s.mkdirAll(filepath.Join(claudeDir, "hooks"))
	s.pruneStaleLinks(filepath.Join(claudeDir, "hooks"))
	s.installHookBinary(filepath.Join(claudeDir, "hooks/flow-guard-hook"))
	settings := filepath.Join(claudeDir, "settings.json")
	if register {
		if err := registerHooks(settings, s.stdout); err != nil {
			s.die("%v", err)
		}
	}
	for _, c := range hookChecks {
		s.warnUnregisteredHook(settings, c)
	}
}

// installHookBinary builds cmd/flow-guard-hook from this checkout to dest, the binary every
// registration runs directly as `"<dest>" hook <name> || true` — one Go start per tool call, where
// scripts/flow-guard would add the bash bootstrap and its source hashing. It is built only from the
// main checkout this installer runs in, and every landing there is followed by this reinstall, so it
// needs no staleness check. The checkout is baked in with -ldflags -X (builtRepoRoot,
// stats/cmd/flow-guard-hook/main.go), standing in for the bootstrap's FLOW_GUARD_REPO_ROOT. The build
// goes to a pid-named file renamed into place, so a hook never execs a half-written binary; GOOS,
// GOARCH and GOFLAGS are removed, as the bootstrap removes them. `|| true` keeps a registration
// fail-open when dest is missing or dies; a failed build warns and leaves any earlier binary.
func (s *installer) installHookBinary(dest string) {
	if _, err := exec.LookPath("go"); err != nil {
		s.warn("no go on PATH — %s not built, so every hook registered on it lets its event through", dest)
		s.skipped++
		return
	}
	tmp := fmt.Sprintf("%s.%d", dest, os.Getpid())
	cmd := exec.Command("go", "build", "-ldflags", "-X 'main.builtRepoRoot="+s.repo+"'", "-o", tmp, "./cmd/flow-guard-hook")
	cmd.Dir = filepath.Join(s.repo, "stats")
	cmd.Env = withoutEnv(os.Environ(), "GOOS", "GOARCH", "GOFLAGS")
	cmd.Stdout, cmd.Stderr = s.stdout, s.stderr
	if err := cmd.Run(); err == nil {
		if err = os.Rename(tmp, dest); err == nil {
			s.info("Built %s", dest)
			return
		}
	}
	_ = os.Remove(tmp)
	s.warn("could not build %s from %s/stats — every hook registered on it lets its event through", dest, s.repo)
	s.skipped++
}

func withoutEnv(env []string, names ...string) []string {
	var out []string
next:
	for _, kv := range env {
		for _, n := range names {
			if strings.HasPrefix(kv, n+"=") {
				continue next
			}
		}
		out = append(out, kv)
	}
	return out
}

var snippetTimeout = regexp.MustCompile(`"timeout": *([0-9]+)`)

// warnUnregisteredHook prints a blank line, the check's warning lines, its snippet and a blank line
// when config does not name its identifier (a missing config included). A snippet pinning
// `"timeout": N` also pins the registered entry's timeout: when no command entry naming the
// identifier carries exactly N, the timeout warning and the snippet are printed instead, since a
// harness default below the hook's own budget kills it mid-check. A registered hook prints nothing,
// and "unregistered" never fails the install: the printed snippet is the user's whole remedy.
func (s *installer) warnUnregisteredHook(config string, c hookCheck) {
	warning := c.warning
	if isRegular(config) {
		b, err := os.ReadFile(config)
		if err != nil {
			s.die("cannot read %s while checking it for the %s hook (%v).\n"+
				"  Its answer decides whether the hook is registered, and guessing 'no' or 'yes'\n"+
				"  would print the wrong remedy.", config, c.label, err)
		}
		if bytes.Contains(b, []byte(c.identifier)) {
			m := snippetTimeout.FindStringSubmatch(c.snippet)
			if m == nil {
				return
			}
			want, _ := strconv.ParseFloat(m[1], 64)
			if v, err := decodeJSON(b); err == nil && hasTimeout(v, c.identifier, want) {
				return
			}
			warning = []string{
				fmt.Sprintf("  ⚠ The %s hook is registered in %s without \"timeout\": %s,", c.label, config, m[1]),
				"    so the harness may kill it mid-check. Register it as:"}
		}
	}
	fmt.Fprintln(s.stdout)
	for _, l := range warning {
		fmt.Fprintln(s.stdout, strings.ReplaceAll(l, "{settings}", config))
	}
	fmt.Fprintf(s.stdout, "%s\n\n", c.snippet)
}

// hasTimeout reports whether any object anywhere in v whose "command" names identifier carries
// "timeout" equal to want.
func hasTimeout(v any, identifier string, want float64) bool {
	switch v := v.(type) {
	case *jsonObject:
		if cmd, ok := v.vals["command"]; ok && strings.Contains(fmt.Sprint(cmd), identifier) {
			if n, ok := v.vals["timeout"].(json.Number); ok {
				if f, err := n.Float64(); err == nil && f == want {
					return true
				}
			}
		}
		for _, k := range v.keys {
			if hasTimeout(v.vals[k], identifier, want) {
				return true
			}
		}
	case []any:
		for _, e := range v {
			if hasTimeout(e, identifier, want) {
				return true
			}
		}
	}
	return false
}

// --- --register-hooks ---

// checkSettings refuses a settings.json registerHooks could not merge into — before any install
// step runs, so a refusal leaves the whole home untouched. A missing file is fine.
func checkSettings(path string) error {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot read %s: %v — --register-hooks refuses to merge into it", path, err)
	}
	if _, err := settingsRoot(b); err != nil {
		return fmt.Errorf("%s %v — --register-hooks refuses to rewrite it. Fix the file and re-run; it was not modified", path, err)
	}
	return nil
}

// settingsRoot decodes settings.json, requiring an object whose "hooks" (an object of arrays) and
// "permissions.allow" (an array) have the shapes a merge appends to.
func settingsRoot(b []byte) (*jsonObject, error) {
	v, err := decodeJSON(b)
	if err != nil {
		return nil, fmt.Errorf("is not valid JSON (%v)", err)
	}
	root, ok := v.(*jsonObject)
	if !ok {
		return nil, errors.New("is not a JSON object")
	}
	if h, ok := root.vals["hooks"]; ok {
		hooks, ok := h.(*jsonObject)
		if !ok {
			return nil, errors.New(`has a "hooks" that is not an object`)
		}
		for _, k := range hooks.keys {
			if _, ok := hooks.vals[k].([]any); !ok {
				return nil, fmt.Errorf(`has a "hooks.%s" that is not an array`, k)
			}
		}
	}
	if p, ok := root.vals["permissions"]; ok {
		perms, ok := p.(*jsonObject)
		if !ok {
			return nil, errors.New(`has a "permissions" that is not an object`)
		}
		if a, ok := perms.vals["allow"]; ok {
			if _, ok := a.([]any); !ok {
				return nil, errors.New(`has a "permissions.allow" that is not an array`)
			}
		}
	}
	return root, nil
}

// registerHooks merges every registration of hookChecks into the settings file at path, printing
// one line per registration. A hook entry counts as present when an object under its event names
// the hook in its "command" (its timeout is set when the registration pins one it lacks); the
// permission when permissions.allow holds it. Everything else in the file — every foreign key and
// hook, and their order — is kept. Nothing changed: nothing written. A change: a timestamped copy
// of the file first, then a temp file in the same directory, re-read and checked, renamed into
// place. A symlinked settings.json is written through to its target.
func registerHooks(path string, out io.Writer) error {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	}
	b, err := os.ReadFile(path)
	existed := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot read %s: %v — --register-hooks refuses to merge into it", path, err)
	}
	root := &jsonObject{vals: map[string]any{}}
	if existed {
		if root, err = settingsRoot(b); err != nil {
			return fmt.Errorf("%s %v — --register-hooks refuses to rewrite it. Fix the file and re-run; it was not modified", path, err)
		}
	}
	fmt.Fprintf(out, "→ Registering hooks in %s\n", path)
	changed := mergeRegistrations(root, out)
	if !changed {
		return nil
	}
	mode := os.FileMode(0o644)
	if existed {
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		mode = fi.Mode().Perm()
		backup := path + "." + time.Now().Format("20060102-150405") + ".bak"
		if err := os.WriteFile(backup, b, mode); err != nil {
			return fmt.Errorf("cannot back %s up to %s: %v — not modified", path, backup, err)
		}
		fmt.Fprintf(out, "  ✓ previous copy kept at %s\n", backup)
	}
	var enc bytes.Buffer
	encodeJSON(&enc, root, "")
	enc.WriteString("\n")
	check, err := settingsRoot(enc.Bytes())
	if err != nil || mergeRegistrations(check, io.Discard) {
		return fmt.Errorf("the merged %s did not read back as written (%v) — not modified", path, err)
	}
	tmp := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s.tmp.%d", filepath.Base(path), os.Getpid()))
	defer os.Remove(tmp)
	if err := os.WriteFile(tmp, enc.Bytes(), mode); err != nil {
		return fmt.Errorf("cannot write %s: %v — %s not modified", tmp, err, path)
	}
	if err := os.Chmod(tmp, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("cannot rename %s onto %s: %v — not modified", tmp, path, err)
	}
	return nil
}

// mergeRegistrations applies every registration to root, reporting each, and says whether any
// changed it.
func mergeRegistrations(root *jsonObject, out io.Writer) bool {
	changed := false
	for _, c := range hookChecks {
		if c.permission != "" {
			perms := root.object("permissions")
			allow, _ := perms.vals["allow"].([]any)
			present := false
			for _, a := range allow {
				if a == c.permission {
					present = true
				}
			}
			if present {
				fmt.Fprintf(out, "  – permissions.allow %s already present\n", c.permission)
				continue
			}
			perms.set("allow", append(allow, c.permission))
			fmt.Fprintf(out, "  ✓ added permissions.allow %s\n", c.permission)
			changed = true
			continue
		}
		for _, r := range c.regs {
			hooks := root.object("hooks")
			groups, _ := hooks.vals[r.event].([]any)
			found := commandObjects(groups, c.identifier)
			switch {
			case len(found) == 0:
				entry := &jsonObject{}
				entry.set("type", "command")
				entry.set("command", hookCommand(c.identifier))
				if r.timeout > 0 {
					entry.set("timeout", json.Number(strconv.Itoa(r.timeout)))
				}
				group := &jsonObject{}
				if r.matcher != "" {
					group.set("matcher", r.matcher)
				}
				group.set("hooks", []any{entry})
				hooks.set(r.event, append(groups, group))
				fmt.Fprintf(out, "  ✓ added %s hook %s\n", r.event, c.identifier)
				changed = true
			case r.timeout > 0 && !hasTimeout(groups, c.identifier, float64(r.timeout)):
				for _, o := range found {
					o.set("timeout", json.Number(strconv.Itoa(r.timeout)))
				}
				fmt.Fprintf(out, "  ✓ set \"timeout\": %d on %s hook %s\n", r.timeout, r.event, c.identifier)
				changed = true
			default:
				fmt.Fprintf(out, "  – %s hook %s already present\n", r.event, c.identifier)
			}
		}
	}
	return changed
}

// commandObjects lists every object under v whose "command" names identifier.
func commandObjects(v any, identifier string) []*jsonObject {
	var out []*jsonObject
	switch v := v.(type) {
	case *jsonObject:
		if cmd, ok := v.vals["command"].(string); ok && strings.Contains(cmd, identifier) {
			out = append(out, v)
		}
		for _, k := range v.keys {
			out = append(out, commandObjects(v.vals[k], identifier)...)
		}
	case []any:
		for _, e := range v {
			out = append(out, commandObjects(e, identifier)...)
		}
	}
	return out
}

// --- order-keeping JSON: settings.json is the user's file, so a merge keeps its key order and
// writes every number exactly as it was read ---

type jsonObject struct {
	keys []string
	vals map[string]any
}

func (o *jsonObject) set(k string, v any) {
	if o.vals == nil {
		o.vals = map[string]any{}
	}
	if _, ok := o.vals[k]; !ok {
		o.keys = append(o.keys, k)
	}
	o.vals[k] = v
}

// object returns the object at key k, adding an empty one when absent.
func (o *jsonObject) object(k string) *jsonObject {
	if v, ok := o.vals[k].(*jsonObject); ok {
		return v
	}
	v := &jsonObject{vals: map[string]any{}}
	o.set(k, v)
	return v
}

func decodeJSON(b []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	v, err := decodeValue(d)
	if err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errors.New("trailing data after the top-level value")
	}
	return v, nil
}

func decodeValue(d *json.Decoder) (any, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}
	switch t {
	case json.Delim('{'):
		o := &jsonObject{vals: map[string]any{}}
		for d.More() {
			kt, err := d.Token()
			if err != nil {
				return nil, err
			}
			v, err := decodeValue(d)
			if err != nil {
				return nil, err
			}
			o.set(kt.(string), v)
		}
		_, err := d.Token()
		return o, err
	case json.Delim('['):
		a := []any{}
		for d.More() {
			v, err := decodeValue(d)
			if err != nil {
				return nil, err
			}
			a = append(a, v)
		}
		_, err := d.Token()
		return a, err
	}
	return t, nil
}

// encodeJSON writes v indented by two spaces, as Claude Code writes settings.json.
func encodeJSON(w *bytes.Buffer, v any, indent string) {
	inner := indent + "  "
	switch v := v.(type) {
	case *jsonObject:
		if len(v.keys) == 0 {
			w.WriteString("{}")
			return
		}
		w.WriteString("{\n")
		for i, k := range v.keys {
			w.WriteString(inner)
			encodeString(w, k)
			w.WriteString(": ")
			encodeJSON(w, v.vals[k], inner)
			if i < len(v.keys)-1 {
				w.WriteString(",")
			}
			w.WriteString("\n")
		}
		w.WriteString(indent + "}")
	case []any:
		if len(v) == 0 {
			w.WriteString("[]")
			return
		}
		w.WriteString("[\n")
		for i, e := range v {
			w.WriteString(inner)
			encodeJSON(w, e, inner)
			if i < len(v)-1 {
				w.WriteString(",")
			}
			w.WriteString("\n")
		}
		w.WriteString(indent + "]")
	case string:
		encodeString(w, v)
	case json.Number:
		w.WriteString(v.String())
	case bool:
		w.WriteString(strconv.FormatBool(v))
	case nil:
		w.WriteString("null")
	}
}

func encodeString(w *bytes.Buffer, s string) {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	_ = e.Encode(s)
	w.Truncate(w.Len() - 1) // Encode's newline
}
