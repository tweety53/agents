package guard

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// flow-guard hook enforce-agent-baseline — the PreToolUse hook that blocks subagent dispatches
// omitting the agent-baseline pointer (hooks/enforce-agent-baseline.py runs it).
//
// Rules in ~/.claude/CLAUDE.md are injected into a session's system prompt and are NOT inherited
// by dispatched agents, nor by agents those agents dispatch. The prompt is the only channel, so
// every dispatch must carry a pointer to ~/.claude/rules/agent-baseline.md plus the sentence that
// makes the receiving agent pass it on in turn.
//
// The PreToolUse contract has no way to rewrite tool input, so this hook cannot inject the line
// itself. It denies the call instead and returns the exact text to add; the model then
// re-dispatches with it.
//
// It denies every fork in a flow session -- one whose transcript records a /flow,
// /flow-interactive, /flow-plan or /flow-self-review command, or whose prompt names a flow skill: a
// flow dispatch is a fresh agent whose prompt carries the paths it needs (KAN-1000: a parent forked
// itself into a self-review loop and dragged ~300k tokens of its own context into it).
//
// It denies, in a flow session, a dispatch whose description opens with a loop key
// (`pipeline-fix-`, `self-review-`) and whose prompt carries no `Key:` line: the
// enforce-loop-dispatch hook records the row under the loop's change off that line, so the fix is
// the line, never a begin under the run's change (kan-1032's self-review: that begin left a twin
// row).
//
// It begins, best-effort, in a flow session, a dispatch's ledger row when none is open (KAN-1115,
// decision hook-same-file): the session's change and session token are read off the store as its
// latest open stage run, else its latest, of any command, skipping a loop's own runs -- a change
// named `pipeline-fix-...` or `self-review-...`, whose stage runs check-loop-dispatch's begins open
// in the parent's session; `flow record dispatches` lists its rows, and when none is open (empty
// `outcome`) under the dispatch's key the hook runs `flow record dispatch begin` itself -- the key
// the prompt's column-0 `Key:` line, else the description's first word; the role the prompt's
// `Role:` line mapped back as launchRoleID (stats/cmd/flow/record.go) maps it; the call's model and
// effort. The begin is written only after every deny check has passed, so a launch the hook denies
// leaves no row. The row's end is the await-own-agents hook's, through flow-guard
// end-dispatch-row. The ledger never denies a dispatch: a dispatch with no key or no mapped role, a
// begin `flow` refuses, an unreachable store, no stage begin for the session, a stage run with no
// session token, an absent or failing `flow` CLI, or an unparseable answer lets the dispatch
// through with no begin. Nothing is begun once the session's change is FINISHED in the store.
//
// A planning agent (Explore, Plan) is begun under its `explore-<n>` or `design-<n>` key; one
// carrying a `Role:` line only while the session's latest stage run is `flow.plan` (decision
// planning-agent-begins-in-brainstorm), and none whose prompt names only another worktree of the
// main checkout: a loop planner's agents keep the planner's explicit begin. An `Explore` or
// general-purpose agent with no `Role:` line is begun as `explorer` or `general`. A planning or
// general agent naming no key is begun under `agent-<tool_use_id>`.
//
// It denies, in a flow session, a `flow-*` dispatch whose prompt has no `Role: <id>` line, whose id
// is not a role of `flow policy get`, or whose `<model>/<subagent_type>` pair is in neither the
// role's base nor its optin list: the deny names the role's admitted pairs, or the roles. A `flow
// policy get` that fails lets the dispatch through (skills/flow-contracts/model-policy.md,
// **Allowed pairs**). FLOW_HOOK_ROLE_CHECK=off skips that check and the pair-judge pick check (the
// hooks' own harness, for cases not about them).
//
// Each deny, once written, also records one `flow record incident` row, with the session's change
// where the store names one; that write fails open, leaving the deny as it stands.
//
// Fails open: any unexpected input, parse error, or internal fault prints nothing and lets the
// dispatch through. A guardrail on a convention must never be the reason work stops.
func init() { hooks["enforce-agent-baseline"] = enforceAgentBaseline }

const eabRequiredText = "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for " +
	"this whole task. Include this instruction verbatim in any prompt you write for " +
	"another agent."

const eabDenyReason = "Dispatch blocked: the subagent prompt does not carry the agent-baseline pointer.\n\n" +
	"A dispatched agent starts from an empty context. It has never read ~/.claude/CLAUDE.md, " +
	"and neither will anything it dispatches. Without the pointer it operates under no rules " +
	"at all — production access, lint suppression, pushes to main, dependency versions.\n\n" +
	"Re-send this call with the following two sentences included verbatim in `prompt`:\n\n" +
	eabRequiredText + "\n\n" +
	"The second sentence is not optional: it is what carries the rules past depth 1.\n" +
	"Exempt: subagent_type \"fork\" outside a flow session, which inherits this session's " +
	"context already."

const eabForkReason = "Dispatch blocked: no flow phase, and no loop a phase or the self-review runs, dispatches " +
	"subagent_type \"fork\" (skills/flow/implement.md, KAN-1000).\n\n" +
	"Re-send this call as a fresh agent on the pair its contract names, its prompt carrying " +
	"what it needs -- paths to the run's artifacts, reports and records -- never this " +
	"session's context."

var (
	// eabFlowCommand is a flow run's slash command as the harness writes it into the transcript.
	eabFlowCommand = regexp.MustCompile(`<command-name>/flow(?:-fast|-interactive|-plan|-self-review)?</command-name>`)
	eabFlowSkills  = []string{"flow", "flow-interactive", "flow-plan", "flow-self-review"}
	// eabLoopKeyLine is the loop key line exactly as the harvester reads it (loopKeyPattern,
	// stats/internal/harvest/transcript.go).
	eabLoopKeyLine   = regexp.MustCompile(`(?m)^[ \t]*Key:[ \t]*(?:pipeline-fix|self-review)-\S+[ \t]*$`)
	eabAdHocKey      = regexp.MustCompile(`^[a-z0-9][a-z0-9+-]*$`)
	eabRoleLine      = regexp.MustCompile(`(?m)^Role:[ \t]*(\S+)[ \t]*$`)
	eabSingleTaskKey = regexp.MustCompile(`^task-(\d+)-`)
	eabFlowEffort    = regexp.MustCompile(`^flow-(?:planner-)?([a-z]+)$`)
)

// eabLoopKeyPrefixes open a description naming a row the enforce-loop-dispatch hook records.
var eabLoopKeyPrefixes = []string{"pipeline-fix-", "self-review-"}

// eabRowRole maps a prompt's `Role:` id to the row role `flow record dispatch begin -role` takes:
// the inverse of launchRoleID (stats/cmd/flow/record.go).
var eabRowRole = map[string]string{
	"implementer": "implementer", "operator-asked": "implementer",
	"gated-reviewer": "reviewer", "gated-re-reviewer": "reviewer",
	"panel-reviewer": "reviewer", "panel-rerun": "reviewer",
	"panel-fix": "panel-fix", "minor-fix": "panel-fix", "no-progress-fixer": "panel-fix",
	"visual-verifier": "verifier",
	"planner":         "planner", "plan-designer": "planner", "designer": "planner",
	"explorer": "explorer", "general": "general",
}

// eabUnroled is the role of an agent type dispatched with no `Role:` line ("" is a call naming no
// type).
var eabUnroled = map[string]string{"Explore": "explorer", "general-purpose": "general", "claude": "general", "": "general"}

func hasAnyPrefix(s string, prefixes []string) bool {
	return slices.ContainsFunc(prefixes, func(p string) bool { return strings.HasPrefix(s, p) })
}

// eab is one PreToolUse event's state: its payload, tool input and prompt, and the flow-session
// answers each check shares.
type eab struct {
	env                 Env
	payload, in         map[string]any
	prompt              string
	session, run        *bool
	sessionID, cwd, typ string
}

func enforceAgentBaseline(env Env, payload map[string]any, stdout, _ io.Writer) {
	if payload == nil {
		return
	}
	if t := strOf(payload["tool_name"]); t != "Agent" && t != "Task" {
		return
	}
	in, ok := payload["tool_input"].(map[string]any)
	if !ok {
		return
	}
	e := &eab{env: env, payload: payload, in: in, sessionID: strOf(payload["session_id"]), cwd: strOf(payload["cwd"])}
	e.typ = strOf(in["subagent_type"])
	deny := func(reason string) {
		_, _ = io.WriteString(stdout, hookDenyJSON(reason))
		change := ""
		if e.sessionID != "" {
			change, _, _, _ = eabSessionChange(env, e.sessionID)
		}
		hookRecordIncident(env, "enforce-agent-baseline", reason, e.cwd, change)
	}
	if e.typ == "fork" {
		if strings.Contains(strOf(in["prompt"]), "skills/flow") || e.inFlowSession() {
			deny(eabForkReason)
		}
		return
	}
	prompt, ok := in["prompt"].(string)
	if !ok {
		return
	}
	e.prompt = prompt
	lowered := strings.ToLower(prompt)
	if !strings.Contains(lowered, "agent-baseline.md") || !strings.Contains(lowered, "verbatim") {
		deny(eabDenyReason)
		return
	}
	for _, check := range []func() string{e.loopKeyLineReason, e.effortReason, e.rolePairReason} {
		if r := check(); r != "" {
			deny(r)
			return
		}
	}
	begin, reason := e.beginRow()
	switch {
	case reason != "":
		deny(reason)
	case begin != nil:
		_, _, _ = hookRun(env, hookBound(env, 10*time.Second), "", "flow", begin...)
	}
}

func (e *eab) loopWord() string {
	if w := pyWords(e.in["description"]); len(w) > 0 && hasAnyPrefix(w[0], eabLoopKeyPrefixes) {
		return w[0]
	}
	return ""
}

func (e *eab) planning() bool { return e.typ == "Explore" || e.typ == "Plan" }

func (e *eab) loopKeyLineReason() string {
	word := e.loopWord()
	if e.planning() || word == "" || eabLoopKeyLine.MatchString(e.prompt) || !e.inFlowRun() {
		return ""
	}
	return "Dispatch blocked: `" + word + "` is a loop key — its prompt carries `Key: " + word + "` as a line of " +
		"its own, and hooks/enforce-loop-dispatch.py records its row under the loop's change off " +
		"that line. Add the line and re-send; never `flow record dispatch begin` it under the " +
		"run's change."
}

// effortReason denies a flow-session dispatch that would run at the model's built-in effort: a
// subagent_type that is not `flow-<effort>` / `flow-planner-<effort>`, with no `effort` on the
// call. Every flow dispatch runs on a pair its role admits, and every pair names an effort.
func (e *eab) effortReason() string {
	agent := pyStrOr(e.in["subagent_type"], "")
	if eabFlowEffort.MatchString(agent) || pyTruthy(e.in["effort"]) || !e.inFlowSession() {
		return ""
	}
	if agent == "" {
		agent = "<none>"
	}
	return "Dispatch blocked: `subagent_type: " + agent + "` carries no effort and the call " +
		"sets none, so it would run at the model's built-in effort.\n\n" +
		"Run `flow-guard pair-judge <role> <prompt file>` and re-send on the pair it prints " +
		"(`skills/flow-contracts/model-policy.md`, **Model and effort**): a `flow-<effort>` type, or " +
		"an `Explore`/`Plan` type with the pair's `effort` set on the call."
}

// policy is `flow policy get`'s answer, nil when it cannot be read.
func (e *eab) policy() map[string]any {
	out, rc, err := hookRun(e.env, hookBound(e.env, 10*time.Second), e.cwd, "flow", "policy", "get")
	if err != nil || rc != 0 {
		return nil
	}
	var v any
	if json.Unmarshal(out, &v) != nil {
		return nil
	}
	m, _ := v.(map[string]any)
	return m
}

func (e *eab) rolePairReason() string {
	if e.env.Getenv("FLOW_HOOK_ROLE_CHECK") == "off" || !strings.HasPrefix(e.typ, "flow-") || !e.inFlowSession() {
		return ""
	}
	policy := e.policy()
	if len(policy) == 0 {
		return ""
	}
	names := make([]string, 0, len(policy))
	for k := range policy {
		names = append(names, k)
	}
	sort.Strings(names)
	roles := strings.Join(names, ", ")
	line := eabRoleLine.FindStringSubmatch(e.prompt)
	if line == nil {
		return "Dispatch blocked: a flow-* dispatch's prompt carries a `Role: <id>` line " +
			"(skills/flow-contracts/model-policy.md, **Allowed pairs**).\n\n" +
			"Roles: " + roles + ".\nAn unusual task no role fits takes `general`."
	}
	role := line[1]
	entry, ok := policy[role].(map[string]any)
	if !ok {
		return "Dispatch blocked: `Role: " + role + "` is not a role of the policy; never invent one.\n\n" +
			"Roles: " + roles + ".\nAn unusual task no role fits takes `general`."
	}
	var admitted []string
	for _, k := range []string{"base", "optin"} {
		list, _ := entry[k].([]any)
		for _, p := range list {
			admitted = append(admitted, pyStr(p))
		}
	}
	pair := pyStrOr(e.in["model"], "<no model>") + "/" + e.typ
	if slices.Contains(admitted, pair) {
		return ""
	}
	list := strings.Join(admitted, ", ")
	if list == "" {
		list = "none"
	}
	return "Dispatch blocked: `" + role + "` does not admit `" + pair + "`.\n\n" +
		"Admitted pairs: " + list + ".\n" +
		"Re-send with `model` and `subagent_type` set to one of them."
}

// beginKey is the row key a dispatch is begun under: the prompt's column-0 `Key:` line of any
// slug, else the description's first word, when it is a key the hook begins (edrBeginKey).
func (e *eab) beginKey() string {
	if m := edrKey.FindStringSubmatch(e.prompt); m != nil && eabAdHocKey.MatchString(m[1]) {
		return m[1]
	}
	if w := pyWords(e.in["description"]); len(w) > 0 && edrBeginKey.MatchString(w[0]) {
		return w[0]
	}
	return ""
}

// rowKey is beginKey, save for a planning or general agent: its Key line, else its
// explore-/design- word, else `agent-<tool_use_id>` -- never a planner's `plan-<n>` prefix, the
// planner's own row. end-dispatch-row (edrMetaKey) and the harvester derive the same key.
func (e *eab) rowKey(roleID string) string {
	if !e.planning() && roleID != "explorer" && roleID != "general" {
		return e.beginKey()
	}
	if m := edrKey.FindStringSubmatch(e.prompt); m != nil && eabAdHocKey.MatchString(m[1]) {
		return m[1]
	}
	if m := edrPlanningKey.FindStringSubmatch(strOf(e.in["description"])); m != nil {
		return m[1]
	}
	if !e.planning() {
		if k := e.beginKey(); k != "" {
			return k
		}
	}
	if use := strOf(e.payload["tool_use_id"]); use != "" {
		return "agent-" + use
	}
	return ""
}

// beginRow is the begin argv to write once every deny check has passed, or a deny reason when a
// judged role's dispatch has no saved pair-judge pick and its role more than one base pair, or
// neither: a row already open, a dispatch with no key or no mapped role, or an unanswered store.
func (e *eab) beginRow() (args []string, reason string) {
	if eabLoopKeyLine.MatchString(e.prompt) || e.loopWord() != "" {
		return nil, ""
	}
	planning := e.planning()
	roleLine := eabRoleLine.FindStringSubmatch(e.prompt)
	var roleID string
	if roleLine != nil {
		roleID = roleLine[1]
	} else {
		roleID = eabUnroled[pyStrOr(e.in["subagent_type"], "")]
	}
	role := eabRowRole[roleID]
	key := ""
	if role != "" {
		key = e.rowKey(roleID)
	}
	_, cwdOK := e.payload["cwd"].(string)
	if key == "" || role == "" || e.sessionID == "" || !cwdOK || !e.inFlowRun() {
		return nil, ""
	}
	change, project, token, stage := eabSessionChange(e.env, e.sessionID)
	if change == "" {
		return nil, ""
	}
	main := eabMainCheckout(e.env, e.cwd)
	if project != "" && eabProjectKey(main) != project {
		return nil, ""
	}
	if token == "" || main == "" {
		return nil, ""
	}
	worktree := main + "-worktrees/" + change
	ours := regexp.MustCompile(regexp.QuoteMeta(worktree) + `(?:/|\s|$)`).MatchString(e.prompt)
	if (planning || roleLine == nil) && !ours && strings.Contains(e.prompt, main+"-worktrees/") {
		return nil, ""
	}
	if planning && roleLine != nil && stage != "flow.plan" {
		return nil, ""
	}
	out, rc, err := hookRun(e.env, hookBound(e.env, 5*time.Second), "", "flow", "record", "dispatches", "-change", change, "-C", e.cwd)
	if err != nil || rc != 0 {
		return nil, ""
	}
	var rows []any
	if json.Unmarshal(out, &rows) != nil || rows == nil {
		return nil, ""
	}
	for _, r := range rows {
		if m, ok := r.(map[string]any); ok && !pyTruthy(m["outcome"]) && m["key"] == key {
			return nil, ""
		}
	}
	effort := e.in["effort"]
	if m := eabFlowEffort.FindStringSubmatch(pyStrOr(e.in["subagent_type"], "")); m != nil {
		effort = m[1]
	}
	args = []string{"record", "dispatch", "begin", "-C", e.cwd, "-change", change, "-role", role,
		"-model", pyStrOr(e.in["model"], "unknown (agent-defined)"), "-effort", pyStr(effort)}
	if m := eabSingleTaskKey.FindStringSubmatch(key); m != nil {
		args = append(args, "-task", m[1])
	}
	args = append(args, "-key", key, "-session-token", token, "-policy-role", roleID)
	if pick := eabReadPick(filepath.Join(worktree, ".superpowers", "sdd", "pair-judge", key+".pick")); pick != "" {
		return append(args, "-pair-pick", pick), ""
	}
	if roleLine != nil && (ours || !planning) && role != "general" && e.env.Getenv("FLOW_HOOK_ROLE_CHECK") != "off" {
		if e.singleBasePair(roleID) {
			// The text flow-guard pair-judge prints for a one-candidate role.
			return append(args, "-pair-pick", "single: the only candidate"), ""
		}
		return nil, "Dispatch blocked: `" + roleID + "` dispatches take the pair judge's pick, and " +
			"`.superpowers/sdd/pair-judge/" + key + ".pick` is not there.\n\n" +
			"Write the dispatch's prompt to `" + worktree + "/.superpowers/sdd/pair-judge/" + key + ".md`, run\n" +
			"flow-guard pair-judge " + roleID + " " + worktree + "/.superpowers/sdd/pair-judge/" + key + ".md\n" +
			"(it never blocks: a judge that cannot answer prints the role's first pair), then send the " +
			"dispatch on the pair it prints (`skills/flow-contracts/model-policy.md`, **Model and effort**)."
	}
	return args, ""
}

// singleBasePair is whether `flow policy get` gives the role exactly one base pair: the judge has
// nothing to pick, so no .pick is owed. A policy that cannot be read is not single.
func (e *eab) singleBasePair(roleID string) bool {
	entry, ok := e.policy()[roleID].(map[string]any)
	if !ok {
		return false
	}
	base, _ := entry["base"].([]any)
	return len(base) == 1
}

// eabReadPick is the `<source>: <reason>` of the `<model> <effort> <source> <reason…>` line
// flow-guard pair-judge saved, or "" when the file is absent or its first line does not parse.
func eabReadPick(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	line, _ := bufio.NewReader(f).ReadString('\n')
	if !utf8.ValidString(line) {
		return ""
	}
	var fields []string
	rest := line
	for len(fields) < 3 {
		rest = strings.TrimLeftFunc(rest, isPySpace)
		i := strings.IndexFunc(rest, isPySpace)
		if rest == "" || i < 0 {
			return ""
		}
		fields, rest = append(fields, rest[:i]), rest[i:]
	}
	reason := strings.TrimSpace(rest)
	if reason == "" {
		return ""
	}
	return fields[2] + ": " + reason
}

func isPySpace(r rune) bool {
	return strings.ContainsRune(" \t\n\r\v\f\x1c\x1d\x1e\x1f\u0085   ", r)
}

// inFlowRun is a flow session whose change is not yet FINISHED -- the scope of the row checks.
func (e *eab) inFlowRun() bool {
	if e.run == nil {
		r := e.inFlowSession() && !e.changeFinished()
		e.run = &r
	}
	return *e.run
}

// inFlowSession is whether the transcript records a flow command.
func (e *eab) inFlowSession() bool {
	if e.session == nil {
		s := eabTranscriptHasFlowCommand(strOf(e.payload["transcript_path"]), e.payload["transcript_path"])
		e.session = &s
	}
	return *e.session
}

func eabTranscriptHasFlowCommand(path string, raw any) bool {
	if _, ok := raw.(string); !ok {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if eabIsFlowCommand(line) {
			return true
		}
		if err != nil {
			return false
		}
	}
}

// eabIsFlowCommand: a typed slash command is a user turn with string content; a tool result
// quoting the tag carries list content and does not count. A run the model started is an
// assistant turn calling the Skill tool on a flow skill (kan-1131: a flow run started that way
// began no dispatch row). The substring test spares a json parse of every line of a long
// transcript that holds neither.
func eabIsFlowCommand(line []byte) bool {
	if !bytes.Contains(line, []byte("<command-name>/flow")) && !bytes.Contains(line, []byte(`"Skill"`)) {
		return false
	}
	var entry map[string]any
	if json.Unmarshal(line, &entry) != nil {
		return false
	}
	msg, _ := entry["message"].(map[string]any)
	content := msg["content"]
	if entry["type"] == "user" {
		s, ok := content.(string)
		return ok && eabFlowCommand.MatchString(s)
	}
	blocks, ok := content.([]any)
	if entry["type"] != "assistant" || !ok {
		return false
	}
	for _, b := range blocks {
		m, _ := b.(map[string]any)
		input, _ := m["input"].(map[string]any)
		skill, _ := input["skill"].(string)
		if m["type"] == "tool_use" && m["name"] == "Skill" && input != nil && slices.Contains(eabFlowSkills, skill) {
			return true
		}
	}
	return false
}

// changeFinished is whether the session's change is already FINISHED in the store: then the
// session runs no flow phase any more. Unanswered: false.
func (e *eab) changeFinished() bool {
	if e.sessionID == "" {
		return false
	}
	change, key, _, _ := eabSessionChange(e.env, e.sessionID)
	if key == "" {
		if _, ok := e.payload["cwd"].(string); ok {
			key = eabProjectKey(eabMainCheckout(e.env, e.cwd))
		}
	}
	if change == "" || key == "" {
		return false
	}
	v, err := hookGetJSON(e.env, hookFlowAddr(e.env)+"/api/v1/changes/"+pyQuoteURL(key)+"/"+pyQuoteURL(change), hookBound(e.env, 2*time.Second))
	m, _ := v.(map[string]any)
	return err == nil && m["state"] == "FINISHED"
}

// eabSessionChange is the session's latest open stage run, else its latest stage run, of any
// command, skipping a loop's own runs -- a change named `pipeline-fix-...` or `self-review-...`,
// whose stage runs check-loop-dispatch's begins open in the parent's session: its change, project
// key, session token and stage. An open run wins over a later one already ended, so a
// pipeline-fix run that began and ended inside an open sdd-tdd run does not name the session's
// change from then on (kan-1099 self-review, F12: row 139 was never begun). All "" when the store
// does not answer.
func eabSessionChange(env Env, sessionID string) (change, project, token, stage string) {
	// ponytail: an open run older than the 50 latest is missed; filter ended_at server-side if that bites.
	q := url.Values{"session_id": {sessionID}, "sort": {"-started_at"}, "limit": {"50"}}
	v, err := hookGetJSON(env, hookFlowAddr(env)+"/api/v1/stage-runs?"+q.Encode(), hookBound(env, 2*time.Second))
	body, _ := v.(map[string]any)
	if err != nil || body == nil {
		return "", "", "", ""
	}
	list, _ := body["stageRuns"].([]any)
	var runs []map[string]any
	for _, r := range list {
		if m, ok := r.(map[string]any); ok && !hasAnyPrefix(pyStrOr(m["changeName"], ""), eabLoopKeyPrefixes) {
			runs = append(runs, m)
		}
	}
	var run map[string]any
	for _, r := range runs {
		if !pyTruthy(r["endedAt"]) {
			run = r
			break
		}
	}
	if run == nil && len(runs) > 0 {
		run = runs[0]
	}
	return strOf(run["changeName"]), strOf(run["projectKey"]), strOf(run["sessionToken"]), strOf(run["stage"])
}

// eabMainCheckout is cwd's main checkout, from its git common dir; "" outside a repository.
func eabMainCheckout(env Env, cwd string) string {
	out, rc, err := hookRun(env, hookBound(env, 5*time.Second), "", "git", "-C", cwd, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil || rc != 0 {
		return ""
	}
	dir := filepath.Dir(filepath.Clean(strings.TrimSpace(string(out))))
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		return real
	}
	return dir
}

// eabProjectKey is the store's project key for a main checkout: stats/internal/fallback/statefile.go
// ProjectKey. "" for "".
func eabProjectKey(main string) string {
	if main == "" {
		return ""
	}
	sum := sha1.Sum([]byte(main))
	return filepath.Base(main) + "-" + hex.EncodeToString(sum[:])[:8]
}
