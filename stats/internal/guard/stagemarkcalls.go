package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The contract below moved here from the header of the deleted
// scripts/check-stage-mark-calls.sh shim (KAN-1058).
//
// flow-guard check-stage-mark-calls — every `flow stage begin` in a skill must carry
// a session token and a harness, the session token must be a literal, and the harness must
// NOT be — the two required flags are wrong in opposite directions, and this
// script rejects both. Every `flow record dispatch` must carry a literal
// session token too, for the same reason and by the same two checks.
// An `flow-guard integrate-change` call, which makes /flow's integrate and cleanup
// marks with the --session-token and --harness it is passed, is held to
// `stage begin`'s token and harness rules.
// A `flow-guard visual-verify-pass prepare` call, which begins flow.visual-verify and
// each verifier row with the --session-token and --harness it is passed, is
// held to `flow-guard integrate-change`'s rules; a `flow-guard visual-verify-pass close`
// call, which ends them, to `flow record dispatch`'s token rules alone (it
// takes no harness). A `flow-guard visual-verify-pass measure` call marks nothing and
// is not scanned.
//
// Usage: flow-guard check-stage-mark-calls [path ...]          # dirs or files
//
// With no arguments it scans `skills/`, resolved relative to the repo root —
// the one tree the marking skills live under. A caller may pass explicit
// paths instead (its Go tests do, against t.TempDir() fixtures).
//
// WHAT THIS CATCHES. KAN-172 task 1 gave `flow stage begin` two required
// flags: `-session-token`, a literal token the daemon later finds in the calling
// session's own transcript to bind that stage run's session_id, and
// `-harness`, the harness recording the mark. Neither is enforced by prose:
// four skills carried a `flow stage begin` invocation with neither flag for
// the whole of KAN-16's life before this guard existed (tasks.md, "This guard
// is the deliverable of the task, not the edits"). This script makes that
// defect loud rather than silent, forever.
//
// It also enforces design.md's second decision, "the session token is a literal,
// never a shell substitution": `-session-token "mf-$(date +%s)-$$"` is recorded by
// the transcript *before* the shell expands it, so every session would record
// the identical string and it would identify nothing. A `-session-token` value
// containing `$(`, a backtick, or `$` immediately followed by a shell
// variable-name character is rejected here, in the skill source, before a
// reader ever has the chance to run it — the same three shapes
// `stats/cmd/flow/stage.go`'s `validateSessionToken` rejects at the CLI, and
// `internal/api/stages.go`'s `validateSessionTokenShape` rejects again at the store,
// as defence in depth.
//
// `flow record dispatch` IS SCANNED FOR THE SESSION-TOKEN RULES TOO
// (KAN-258). The record verb takes `-session-token` for precisely the reason
// `stage begin` does: the daemon finds that literal token in the calling
// session's own transcript and binds the dispatch row to the session that
// wrote it. A substituted value there lands in every transcript as the same
// unexpanded string and so discriminates between no two sessions — the whole
// mechanism, silently gone. A guard that enforced the literal on one verb and
// not the other would be a guard with a hole, and the record verb's call
// sites are numerous (four in `skills/flow/review-panel.md` alone). So both
// session-token rules — the flag is present, and its value is a literal —
// are applied to `flow stage begin` and `flow record dispatch` alike.
//
// The other rules stay with the verb that owns them. `-harness` is a `stage
// begin` flag and `record dispatch` has none (a dispatch row inherits its
// harness from the stage run it belongs to), so requiring one on a record
// call would demand a flag the CLI would reject. The guessed-change-name
// rule likewise stays on `stage begin`, whose begin handler is the one that
// bootstraps a change row from a name it has never seen.
//
// Only `flow record dispatch` is scanned: `record finding`, `record
// status`, `record render` and `record journal-count` take no session token
// at all, so there is nothing on them for these rules to say.
//
// `-harness` FAILS THE OPPOSITE WAY: a value that IS a literal is the defect.
// These skill files are one source installed into `~/.claude/skills/` and
// `~/.zcode/skills/` alike (CLAUDE.md's "installed alongside the skills in
// every harness"), so a hardcoded `-harness claude-code` in the shared source
// mislabels every ZCode run as Claude Code — masking the very thing the
// harness field exists to record: which harness ran the stage. The skill source must instead carry a
// placeholder the agent fills in at call time, exactly as `-session-token`'s literal
// token is filled in at call time — this guard requires the `-harness` value
// to be written as a bracketed placeholder (`<harness>`) and rejects any bare
// value, the same way it rejects a substituted session token.
//
// `flow stage end` is deliberately NOT checked: pipeline.md's "Stage marks"
// section states plainly that `stage end` carries no harness of its own (the
// harness is recorded once at `begin` and is immutable), and it takes no
// `-session-token` either — attribution happens once, at `begin`. Only lines that
// invoke `stage begin` are examined.
//
// `/flow-status` marks nothing, per README.md's Level 1 section
// so a `flow stage begin` appearing in its SKILL.md would itself be
// a defect this guard would (correctly) catch — there is no exemption for it.
//
// It also enforces that a `stage begin`'s change argument is never a guess.
// The guard cannot know whether `<name>` is resolved at a given call site,
// but it can know that a bracketed placeholder whose own text says "guess"
// is not — `<name-or-best-guess>` and any restatement of it. KAN-182: on
// 2026-08-15 `/flow-fast`'s state gate marked its stage with
// `<name-or-best-guess>` before its change name was resolved; the daemon's
// begin handler bootstraps a change row for any name it has never seen
// (`ApplyBeginStageMark`, `stats/internal/api/stages.go`), so the guess got
// a row of its own — `kan-175` and `kan-175-more-ui-ux-fixes` appeared as
// two open changes, 29 seconds apart. A mark writes, so a guessed name
// bootstraps a change row that outlives the run.
//
// The check does not try to identify "the change argument" at all. Finding
// the last token on a shell command line requires parsing shell syntax, and
// this guard kept getting that parsing wrong: KAN-182 round 1 added
// quote-awareness to a naive last-token extraction so a quoted or
// commented-out placeholder would still be caught, and that parsing
// introduced three more bypasses (round 2) — a quoted value with an internal
// space that a space-only split truncated before the placeholder, an escaped
// quote inside a session token that desynced the quote-tracker so a later
// `#` was read as an unquoted comment and discarded the rest of the line,
// and a tab before the change argument that a space-only split never saw as
// a token boundary. Extraction is deleted rather than patched a fourth time.
// Instead the check searches the WHOLE assembled command text for a
// bracketed placeholder whose own text says "guess" (`<...guess...>`,
// case-insensitive) — strictly stronger than extracting a token, and immune
// to quoting, spacing and comments because it never tries to reason about
// any of them. One behaviour is worth stating plainly: a guessed placeholder
// written inside a trailing comment is reported too — a `stage begin` line
// that mentions `<name-or-best-guess>` anywhere is a call site to fix.
//
// MULTI-LINE INVOCATIONS. A real call is frequently wrapped across several
// lines with a trailing `\` continuation, exactly as an ordinary shell
// command would be. This guard joins a `stage begin` line with every
// following line while the previous physical line ends in `\`, then checks
// the assembled logical command as one string — a check against the raw
// physical lines alone would miss `-session-token`/`-harness` written on a
// continuation line, which is precisely how these calls are written today.
//
// STAGE KEYS ARE SERVED, NOT TRANSCRIBED. A `stage begin`'s `-stage <key>`
// names a key of the documented stage vocabulary, which this guard does not
// keep a copy of: the key set is served by `go run ./cmd/flow stage keys`
// from $REPO_ROOT's own stats module (internal/stages' table), so a key the
// checked-out tree serves is exactly a key the tree's Go side accepts --
// one source, no third transcription that could drift from it (KAN-533).
// A key the served vocabulary does not list -- a typo, a rename that missed
// a call site -- is a violation here rather than a caller-mistake exit at
// run time, when the mark silently prints one line and the stage goes
// unrecorded. A serve that fails or yields no key at all is a guard that
// cannot answer (exit 2).
//
// Exit codes: 0 every scanned call is compliant; 1 at least one
// violation was found; 2 the guard cannot answer at all (a bad path, a
// rejected coverage record, or no stage key served by flow stage keys — never
// read as "no findings").
//
// Ported to Go (KAN-841): the body's reasoning for every step is in
// stats/internal/guard/stagemarkcalls.go. The binary does not live in this
// checkout, so this shim exports FLOW_GUARD_REPO_ROOT — this script's parent
// directory, derived as the bash guard derived its REPO_ROOT — as the root
// whose skills/ is scanned by default and whose stats module serves the keys.

// checkStageMarkCalls is `flow-guard check-stage-mark-calls`: that script's
// header comment is the contract -- every `flow stage begin` carries a
// literal -session-token, a placeholder -harness, a served -stage key and no
// guessed change name; every `flow record dispatch` a literal
// -session-token; a `flow-guard integrate-change` or `flow-guard visual-verify-pass prepare`
// call the begin token and harness rules, a `flow-guard visual-verify-pass close`
// call the token rules alone; exit 0 compliant, 1 violations (coverage included), 2
// cannot answer. The reasoning for each step, moved here from the bash body
// it replaced (d71a2327), sits beside the code it explains.
func init() {
	Registry["check-stage-mark-calls"] = checkStageMarkCalls
}

// EXPECTED-ZERO FILES — after narrowing, only two SKILL.md files under
// skills/ genuinely carry no checked call site of their own — neither a
// `stage begin` nor a `record dispatch` (measured 2026-08-18 for `stage
// begin` and re-measured 2026-08-22 for `record dispatch`, by grepping each
// directly against this guard's own call-detection pattern), plus
// flow-status/SKILL.md declared separately below with its own by-contract
// reason. Each gets its OWN one-line justification for WHY its zero is
// legitimate by design (KAN-197 F2), not a shared string attesting only to
// how the zero was measured.
var smcExpectedZero = []struct{ file, reason string }{
	{"skills/flow-contracts/SKILL.md", "the contracts index — shared prose loaded by several command skills; it is never itself run as a command, so it marks no stage and dispatches no subagent of its own"},
	{"skills/flow-self-review/SKILL.md", "a standalone command with no per-change state, no implementation or verification stage to mark, and no `flow record dispatch` call — its one-shot dispatches are recorded by none"},
	{"skills/flow-prompt-tuning/SKILL.md", "a standalone command with no per-change state, no pipeline stage to mark, and no `flow record dispatch` call — its replay dispatches are recorded by none"},
	{"skills/flow-settings/SKILL.md", "a standalone settings command with no per-change state, no implementation or verification stage to mark, and no subagent to dispatch — the same reason check-guard-symlinks declared it expected-zero"},
	{"skills/spectre-planning/SKILL.md", "the generic planning prompts spectre vendors and flow loads — shared prose with no change of its own, so the stage marks and dispatch records stay in flow's brainstorm-planner.md"},
	{"skills/flow-experiment-close/SKILL.md", "a standalone closing skill, not a pipeline stage: it only calls the flow CLI and hands a positive verdict to a new /flow run, so it has no stage to mark and no subagent to dispatch"},
	{"skills/flow-learn-quiz/SKILL.md", "a standalone teaching skill run only on request, not a pipeline stage: it has no stage to mark and no subagent to dispatch"},
	{"skills/call-efficiency-audit/SKILL.md", "a standalone measuring skill, not a pipeline stage: it runs on no change, so it has no stage to mark and no subagent to dispatch"},
	{"skills/flow/SKILL.md", "a legitimate zero-mark router file — it resolves state and dispatches into the topic file (brainstorm.md, implement.md, review-panel.md, verify-and-handoff.md, integrate.md, cleanup.md) that owns the phase in force; every flow.* mark lives in one of those phase files, which this guard's corpus now scans directly, never in this router itself"},
	// flow-status marks nothing BY CONTRACT, not merely as a measured fact
	// like the files above: it is a read-only status report, and a stage
	// mark or a dispatch record it wrote would record work nobody did.
	// Declared with its own reason rather than folded into the generic
	// reason, per this task's own note.
	{"skills/flow-status/SKILL.md", "read-only status report; writes no stage marks by contract — a mark here would record work nobody did"},
}

// KAN-197 FIX ROUND, F1 — the corpus is narrowed to the members CAPABLE of
// carrying a `stage begin` call, not every Markdown file under skills/.
// Before this fix the scan glob was '*.md', which reached 33 files while only
// a SKILL.md or pipeline.md could then ever hold a checked invocation (a
// contract doc, a rationale file or a reviewer-prompt file structurally
// cannot) — so 28 of 33 declared entries were declared not because their
// zero was meaningful but because they were never candidates, and every new
// rationale or reviewer-prompt file added under skills/ would need a line
// here purely to stay green. Narrowing the corpus removes that whole
// non-candidate class rather than declaring around it.
//
// KAN-374 — the six `skills/flow/` PHASE FILES (brainstorm.md,
// implement.md, review-panel.md, verify-and-handoff.md, integrate.md,
// cleanup.md) are candidates alongside SKILL.md/pipeline.md:
// `skills/flow/SKILL.md` is a zero-mark router (see its reason above) and
// every `flow.*` mark and `flow record dispatch` call actually lives in
// these six phase files, so a guard meant to catch a missing
// `-session-token`/`-harness` or a substituted token at its true call site
// has to scan them directly rather than only their router. Detection is
// preserved for every existing candidate: a SKILL.md or pipeline.md that
// loses all its marks is still a member of this corpus and still fires.
// KAN-490 widened it again to skills/flow-fast/'s own phase files
// (review.md, finish.md); /flow-fast has since collapsed into its single
// SKILL.md, which carries every one of its marks itself.
// KAN-856 added skills/flow/brainstorm-planner.md: once brainstorm.md stopped
// restating its marks, the planner holds the only flow.design-approval,
// flow.create-artifacts and flow.writing-plans `stage begin` lines.
// KAN-1060 added skills/flow/visual-verify.md, whose one call is
// `flow-guard visual-verify-pass close`.
var smcCandidates = map[string]bool{
	"SKILL.md": true, "pipeline.md": true, "brainstorm.md": true, "brainstorm-planner.md": true, "implement.md": true, "document-fix.md": true,
	"review-panel.md": true, "verify-and-handoff.md": true, "integrate.md": true,
	"cleanup.md": true, "review.md": true, "finish.md": true, "visual-verify.md": true,
}

const (
	smcBegin    = "flow stage begin"
	smcMark     = "flow stage mark"
	smcDispatch = "flow record dispatch"
	// flow-guard integrate-change makes /flow's integrate and cleanup marks itself,
	// with the --session-token and --harness its caller passes: the literal
	// a skill writes there is the one the daemon binds, so its call is held
	// to the same token and harness rules (its stage keys are the script's).
	smcIntegrate = "flow-guard integrate-change"
	// flow-guard visual-verify-pass prepare begins flow.visual-verify and each
	// verifier row with the --session-token and --harness it is passed, so it
	// takes flow-guard integrate-change's rules; close ends them and records with the
	// token alone, so it takes `record dispatch`'s. measure marks nothing.
	smcVVPrepare = "flow-guard visual-verify-pass prepare"
	smcVVClose   = "flow-guard visual-verify-pass close"
)

// The ERE patterns the bash ran through grep/awk. The value extractions are
// compiled POSIX: grep -o takes the leftmost-LONGEST match, so a bare token
// that runs past a closing quote (`"mf-\"abc#tok"`) wins over the quoted
// alternative, as it did in the bash.
var (
	smcBeginLine     = regexp.MustCompile(`^[[:space:]]*flow stage begin([[:space:]]|\\|$)`)
	smcMarkLine      = regexp.MustCompile(`^[[:space:]]*flow stage mark([[:space:]]|\\|$)`)
	smcDispatchLine  = regexp.MustCompile(`^[[:space:]]*flow record dispatch([[:space:]]|\\|$)`)
	smcIntegrateLine = regexp.MustCompile(`^[[:space:]]*flow-guard integrate-change([[:space:]]|\\|$)`)
	smcVVPrepareLine = regexp.MustCompile(`^[[:space:]]*flow-guard visual-verify-pass prepare([[:space:]]|\\|$)`)
	smcVVCloseLine   = regexp.MustCompile(`^[[:space:]]*flow-guard visual-verify-pass close([[:space:]]|\\|$)`)
	smcContinued     = regexp.MustCompile(`\\[[:space:]]*$`)
	smcHasToken      = regexp.MustCompile(`(^|[[:space:]])-session-token([[:space:]]|=)`)
	smcHasHarness    = regexp.MustCompile(`(^|[[:space:]])-harness([[:space:]]|=)`)
	smcTokenValue    = smcValueRe("session-token")
	smcHarnessValue  = smcValueRe("harness")
	smcStageValue    = smcValueRe("stage")
	smcStagesValue   = smcValueRe("stages")
	smcVarRef        = regexp.MustCompile(`\$[A-Za-z_]`)
	smcPlaceholder   = regexp.MustCompile(`^<[^<>]+>$`)
	// guess_placeholder — the first bracketed placeholder whose own text
	// says "guess" (`<...guess...>`, case-insensitive), found ANYWHERE in the
	// assembled command text. No shell parsing at all: no token extraction,
	// no quote-tracking, no comment-stripping — the regex is applied to the
	// whole line, so quoting, internal whitespace, tabs and trailing comments
	// cannot hide a placeholder from it (KAN-182 F6/F7/F8).
	smcGuess = regexp.MustCompilePOSIX(`<[^<>]*[Gg][Uu][Ee][Ss][Ss][^<>]*>`)
)

func smcValueRe(flag string) *regexp.Regexp {
	return regexp.MustCompilePOSIX(`-` + flag + `[[:space:]]+('[^']*'|"[^"]*"|[^[:space:]]+)`)
}

// smcValue is stage_value / session_token_value / harness_value: the first
// `-<flag>`'s argument, single- or double-quoted or bare, one quote pair
// stripped by each of the bash's two sed expressions in turn, or "" when the
// flag does not appear. One extraction per flag keeps the presence check and
// the shape check consistent.
func smcValue(re *regexp.Regexp, text string) string {
	m := re.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	v := m[1]
	for _, q := range []string{"'", `"`} {
		if len(v) >= 2 && strings.HasPrefix(v, q) && strings.HasSuffix(v, q) {
			v = v[1 : len(v)-1]
		}
	}
	return v
}

type smcCall struct {
	line int
	verb string
	cmd  string
}

// smcAssemble is assemble_calls: one call per checked invocation in body,
// continuation lines folded in. Two verbs are checked, `stage begin` and
// `record dispatch`; the verb is carried through so the rules are applied
// per verb rather than re-derived from the command text.
//
// MULTI-LINE INVOCATIONS. A real call is frequently wrapped across several
// lines with a trailing `\` continuation. A `stage begin` line is joined with
// every following line while the previous physical line ends in `\`, and the
// assembled logical command is checked as one string — a check against the
// raw physical lines alone would miss `-session-token`/`-harness` written on
// a continuation line, which is precisely how these calls are written today.
//
// One deliberate difference from the bash: under a UTF-8 locale macOS awk
// aborts on a line that is not valid UTF-8 (`towc: multibyte conversion
// failure`), so the bash counted such a file as zero calls. This reads bytes,
// so the calls in that file are still checked. Under LC_ALL=C the two agree.
func smcAssemble(body string) []smcCall {
	lines := strings.Split(body, "\n")
	if strings.HasSuffix(body, "\n") {
		lines = lines[:len(lines)-1]
	}
	var calls []smcCall
	var cur *smcCall
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if cur != nil {
			cur.cmd += " " + line
			if smcContinued.MatchString(line) {
				cur.cmd = smcContinued.ReplaceAllString(cur.cmd, "")
				continue
			}
			calls = append(calls, *cur)
			cur = nil
			continue
		}
		verb := ""
		switch {
		case smcBeginLine.MatchString(line):
			verb = smcBegin
		case smcMarkLine.MatchString(line):
			verb = smcMark
		case smcDispatchLine.MatchString(line):
			verb = smcDispatch
		case smcIntegrateLine.MatchString(line):
			verb = smcIntegrate
		case smcVVPrepareLine.MatchString(line):
			verb = smcVVPrepare
		case smcVVCloseLine.MatchString(line):
			verb = smcVVClose
		}
		if verb == "" {
			continue
		}
		if smcContinued.MatchString(line) {
			cur = &smcCall{i + 1, verb, smcContinued.ReplaceAllString(line, "")}
			continue
		}
		calls = append(calls, smcCall{i + 1, verb, line})
	}
	// A file ending mid-continuation is a malformed fixture/skill, not a
	// silently dropped call -- what was gathered is checked rather than
	// swallowed.
	if cur != nil {
		calls = append(calls, *cur)
	}
	// The bash read each call back with `IFS=$'\t' read -r`, which strips
	// tabs from both ends of the command text; the rules see what it saw.
	for i := range calls {
		calls[i].cmd = strings.Trim(calls[i].cmd, "\t")
		// The script spells the flags with two dashes; the rules read one.
		if v := calls[i].verb; v == smcIntegrate || v == smcVVPrepare || v == smcVVClose {
			calls[i].cmd = strings.NewReplacer(" --session-token", " -session-token", " --harness", " -harness").Replace(calls[i].cmd)
		}
	}
	return calls
}

type smcScan struct {
	root, keys string
	utf8       bool
	stdout     io.Writer
	violations int
	checked    int
}

func (s *smcScan) finding(format string, a ...any) {
	fmt.Fprintf(s.stdout, format+"\n", a...)
	s.violations++
}

func checkStageMarkCalls(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-stage-mark-calls: "+format+"\n", a...)
		return 2
	}
	// REPO_ROOT is the guard script's own checkout. The bash derived it from
	// BASH_SOURCE; this binary lives in a cache, so its shim exports
	// FLOW_GUARD_REPO_ROOT, derived the same way.
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		return die("FLOW_GUARD_REPO_ROOT is unset — run flow-guard check-stage-mark-calls, which sets it")
	}
	targets := args
	if len(targets) == 0 {
		targets = []string{root + "/skills"}
	}
	for _, t := range targets {
		if _, err := os.Stat(smcAbs(env, t)); t == "" || err != nil {
			return die("no such file or directory: %s", t)
		}
	}

	// STAGE KEYS ARE SERVED, NOT TRANSCRIBED. A `stage begin`'s `-stage
	// <key>` names a key of the documented stage vocabulary, which this guard
	// does not keep a copy of: the key set is served by `go run ./cmd/flow
	// stage keys` from REPO_ROOT's own stats module (internal/stages' table),
	// so a key the checked-out tree serves is exactly a key the tree's Go
	// side accepts -- one source, no third transcription that could drift
	// from it (KAN-533). Never an installed binary's, and never a README
	// transcription, which this guard used to keep and which could drift
	// from the Go side. Read once; a failed serve or an empty answer is a
	// refusal (exit 2).
	serve := env.StageKeys
	if serve == nil {
		serve = func() (string, error) { return smcServeKeys(env, root, stderr) }
	}
	keys, err := serve()
	if err != nil {
		return die("the served stage-key source failed: (cd %s/stats && go run ./cmd/flow stage keys)", root)
	}
	if keys = strings.TrimRight(keys, "\n"); keys == "" {
		return die("the served stage-key source (flow stage keys) printed no keys")
	}

	s := &smcScan{root: root, keys: keys, utf8: smcUTF8(env), stdout: stdout}
	var cov coverage
	// The declared list is a hardcoded set of THIS REPOSITORY's own paths,
	// declared ONLY for the guard's own default, full-corpus scan (no
	// explicit targets). A caller-supplied target scans a wholly different,
	// smaller tree where none of these paths exist; declaring them there
	// would make every one a KAN-197 F3 "declared but never recorded"
	// violation even though they are simply not part of that run's corpus —
	// not a real staleness. Gating on "no arguments" keeps F3's protection
	// exactly where it means something: this guard's own real, default
	// invocation.
	if len(args) == 0 {
		for _, z := range smcExpectedZero {
			if err := cov.declare(z.file, z.reason); err != nil {
				fmt.Fprintln(stderr, err)
				return die("coverage_declare failed for '%s' (see stderr above)", z.file)
			}
		}
	}

	for _, target := range targets {
		files := []string{target}
		if isDir(smcAbs(env, target)) {
			files = smcFind(env, target, stderr)
			if len(files) == 0 {
				// KAN-197 F7: a directory target that enumerates to ZERO
				// candidate files would otherwise vanish from coverage
				// entirely rather than being caught — the exact vacuous-pass
				// shape KAN-197 closes. The TARGET itself is recorded with a
				// zero count so it becomes a corpus member like any other --
				// undeclared, it is now a violation naming itself.
				if err := cov.record(strings.TrimPrefix(target, root+"/"), 0); err != nil {
					fmt.Fprintln(stderr, err)
					return die("coverage_record failed for target '%s' (see stderr above)", target)
				}
				continue
			}
		}
		for _, f := range files {
			n := s.checkFile(env, f, stderr)
			// KAN-197 coverage: how many checked calls — `stage begin` and
			// `record dispatch` together — this guard actually found IN
			// THIS FILE; distinct from checked, the corpus-wide total. A
			// file the corpus reaches but which carries no checked call at
			// all is "nothing was checked here," visible even on a run that
			// goes on to report clean.
			//
			// KAN-197 F8: a rejected record (e.g. the same file enumerated
			// twice) refuses to answer rather than going on to report clean
			// against its own contradictory breakdown.
			if err := cov.record(strings.TrimPrefix(f, root+"/"), n); err != nil {
				fmt.Fprintln(stderr, err)
				return die("coverage_record failed for '%s' (see stderr above)", f)
			}
		}
	}

	// COVERAGE — a file whose coverage is zero and is not declared above is
	// folded into the ordinary violation count, exactly as
	// check-references and check-guard-calls do it — never a
	// separate exit status: "<member>: <text>" becomes "<member>:0: <text>".
	for _, line := range cov.verdict() {
		member, _, _ := strings.Cut(line, ":")
		_, text, _ := strings.Cut(line, ": ")
		s.finding("%s:0: %s", member, text)
	}
	if s.violations > 0 {
		fmt.Fprintf(stderr, "check-stage-mark-calls: %d violation(s) across %d call site(s) checked\n", s.violations, s.checked)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Stage-mark-calls guard: clean (%d call site(s) checked)\n", s.checked)
	if frag := cov.report(); frag != "" {
		fmt.Fprintf(stdout, "  %s\n", frag)
	}
	return 0
}

// checkFile applies every rule to each checked call in f and returns how
// many it found. An unreadable file is awk's error and zero calls, as the
// bash's process substitution left it.
func (s *smcScan) checkFile(env Env, f string, stderr io.Writer) int {
	body, err := os.ReadFile(smcAbs(env, f))
	if err != nil {
		// awk counts the program text's opening newline, so BEGIN is line 2.
		fmt.Fprintf(stderr, "awk: can't open file %s\n source line number 2\n", f)
		return 0
	}
	calls := smcAssemble(string(body))
	for _, c := range calls {
		s.checked++
		s.checkCall(f, c)
	}
	return len(calls)
}

func (s *smcScan) checkCall(f string, c smcCall) {
	hasToken := smcHasToken.MatchString(c.cmd)
	if !hasToken {
		s.finding("%s:%d: `%s` carries no -session-token -- required so the daemon can later bind this record to the session that wrote it (design.md, \"bind after the fact, by a correlator the caller writes\")", f, c.line, c.verb)
	}

	// `-harness` is a `stage begin` flag and `flow record dispatch` has none:
	// a dispatch row inherits its harness from the stage run it belongs to,
	// so requiring one here would demand a flag the CLI would reject. The
	// rule is applied to the verb that takes it, not to every scanned call.
	if c.verb != smcDispatch && c.verb != smcVVClose {
		if !smcHasHarness.MatchString(c.cmd) {
			s.finding("%s:%d: `%s` carries no -harness -- required so a recorded run states which harness marked it", f, c.line, c.verb)
		} else if h := smcValue(smcHarnessValue, c.cmd); !smcPlaceholder.MatchString(h) {
			s.finding("%s:%d: -harness %s is a hardcoded literal, not a placeholder -- this skill source is one file installed into `~/.claude/skills/` and `~/.zcode/skills/` alike, so a fixed value mislabels every harness but the one it names; write a bracketed placeholder (e.g. -harness <harness>) that the agent fills in with the harness actually running the command", f, c.line, smcQuote(h, s.utf8))
		}
	}

	if hasToken {
		switch tok := smcValue(smcTokenValue, c.cmd); {
		case strings.Contains(tok, "$("):
			s.finding("%s:%d: -session-token %s contains a command substitution \"$(...)\" -- the transcript records the command text before the shell expands it, so this value would be recorded identically by every caller and identify nothing; write a literal token instead", f, c.line, smcQuote(tok, s.utf8))
		case strings.Contains(tok, "`"):
			s.finding("%s:%d: -session-token %s contains a backtick -- backtick command substitution is expanded by the shell after the transcript already recorded the unexpanded text; write a literal token instead", f, c.line, smcQuote(tok, s.utf8))
		case smcVarRef.MatchString(tok):
			s.finding("%s:%d: -session-token %s contains a shell variable reference ($VAR) -- the transcript records the command text before the shell expands it, so this value would be recorded identically by every caller and identify nothing; write a literal token instead", f, c.line, smcQuote(tok, s.utf8))
		}
	}

	if c.verb == smcDispatch || c.verb == smcIntegrate || c.verb == smcVVPrepare || c.verb == smcVVClose {
		return
	}
	short := strings.TrimPrefix(c.verb, "flow ")
	if guess := smcGuess.FindString(c.cmd); guess != "" {
		s.finding("%s:%d: `%s` names a guess, not a resolved change (%s) -- a mark writes, so a guessed name bootstraps a change row that outlives the run; wait until the change name is resolved before marking", f, c.line, short, guess)
	}
	// A `stage mark` names its keys comma-separated in -stages; each is
	// checked exactly as a `stage begin`'s one -stage key is.
	if c.verb == smcMark {
		v := smcValue(smcStagesValue, c.cmd)
		if v == "" {
			s.finding("%s:%d: `stage mark` carries no -stages key", f, c.line)
			return
		}
		for _, key := range strings.Split(v, ",") {
			s.checkKey(f, c, key)
		}
		return
	}
	// An unlisted key -- a typo, a rename that missed a call site -- is a
	// violation here rather than a caller-mistake exit at run time, when the
	// mark silently prints one line and the stage goes unrecorded. The
	// membership test is an exact whole-line, fixed-string match (grep -qxF).
	key := smcValue(smcStageValue, c.cmd)
	if key == "" {
		s.finding("%s:%d: `stage begin` carries no -stage key", f, c.line)
		return
	}
	s.checkKey(f, c, key)
}

func (s *smcScan) checkKey(f string, c smcCall, key string) {
	if !smcListed(s.keys, key) {
		s.finding("%s:%d: -stage %s is not a key of the served stage-key vocabulary (flow stage keys) -- a mark under an unknown key is refused by the daemon as a caller mistake and the stage goes unrecorded; use a listed key or add the row first", f, c.line, smcQuote(key, s.utf8))
	}
}

func smcListed(keys, key string) bool {
	for _, k := range strings.Split(keys, "\n") {
		if k == key {
			return true
		}
	}
	return false
}

// smcServeKeys runs `cd <root>/stats && go run ./cmd/flow stage keys` with
// go from env's PATH, its stderr passed through as the bash's was, and
// returns its stdout.
func smcServeKeys(env Env, root string, stderr io.Writer) (string, error) {
	dir := root + "/stats"
	if !isDir(dir) {
		fmt.Fprintf(stderr, "check-stage-mark-calls: cd: %s: No such file or directory\n", dir)
		return "", errors.New("no stats directory")
	}
	goBin, ok := lookPath(env, "go")
	if !ok {
		fmt.Fprint(stderr, "check-stage-mark-calls: go: command not found\n")
		return "", errors.New("no go on PATH")
	}
	cmd := exec.Command(goBin, "run", "./cmd/flow", "stage", "keys")
	cmd.Dir = dir
	cmd.Stderr = stderr
	out, err := cmd.Output()
	return string(out), err
}

// smcFind is `find <target> -type f \( -name SKILL.md -o ... \)`: candidate
// basenames only, regular files only, no symlink followed (a symlinked
// target operand included), in find's own order -- directory order,
// preorder, never sorted -- which the findings and the coverage breakdown
// carry. Each path is spelled as find prints it.
func smcFind(env Env, target string, stderr io.Writer) []string {
	var files []string
	var walk func(disp string)
	walk = func(disp string) {
		d, err := os.Open(smcAbs(env, disp))
		var ents []os.DirEntry
		if err == nil {
			ents, err = d.ReadDir(-1)
			d.Close()
		}
		if err != nil {
			fmt.Fprintf(stderr, "find: %s: %s\n", disp, ppStrerror(err))
		}
		for _, e := range ents {
			p := disp + "/" + e.Name()
			if strings.HasSuffix(disp, "/") {
				p = disp + e.Name()
			}
			switch {
			case e.IsDir():
				walk(p)
			case e.Type().IsRegular() && smcCandidates[e.Name()]:
				files = append(files, p)
			}
		}
	}
	if fi, err := os.Lstat(smcAbs(env, target)); err == nil && fi.IsDir() {
		walk(target)
	}
	return files
}

// smcBackslashed is the set bash's printf %q backslash-escapes anywhere in a
// word; `#` and `~` are escaped only as its first byte.
const smcBackslashed = " !\"$&'()*,;<>?[\\]^`{|}"

// smcQuote is bash's `printf %q`: ” for an empty word, $'...' ANSI-C
// quoting when any character is unprintable in the locale, else each shell
// metacharacter backslash-escaped.
func smcQuote(s string, utf8Locale bool) string {
	if s == "" {
		return "''"
	}
	if smcUnprintable(s, utf8Locale) {
		return smcANSIC(s, utf8Locale)
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if strings.IndexByte(smcBackslashed, c) >= 0 || (i == 0 && (c == '#' || c == '~')) {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

func smcUnprintable(s string, utf8Locale bool) bool {
	if !utf8Locale {
		for i := 0; i < len(s); i++ {
			if s[i] < 0x20 || s[i] >= 0x7f {
				return true
			}
		}
		return false
	}
	for _, r := range s {
		if r == utf8.RuneError || !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}

var smcEscapes = map[byte]string{'\a': `\a`, '\b': `\b`, 0x1b: `\E`, '\f': `\f`, '\n': `\n`,
	'\r': `\r`, '\t': `\t`, '\v': `\v`, '\\': `\\`, '\'': `\'`}

// smcANSIC is bash's ansic_quote: $'...' with the C escapes, \' and \\, a
// printable character as itself and any other byte as \ooo.
func smcANSIC(s string, utf8Locale bool) string {
	var b strings.Builder
	b.WriteString("$'")
	for i := 0; i < len(s); {
		c := s[i]
		if esc, ok := smcEscapes[c]; ok {
			b.WriteString(esc)
			i++
			continue
		}
		if c >= 0x20 && c < 0x7f {
			b.WriteByte(c)
			i++
			continue
		}
		if utf8Locale && c >= 0x80 {
			if r, n := utf8.DecodeRuneInString(s[i:]); r != utf8.RuneError && unicode.IsPrint(r) {
				b.WriteString(s[i : i+n])
				i += n
				continue
			}
		}
		fmt.Fprintf(&b, `\%03o`, c)
		i++
	}
	b.WriteString("'")
	return b.String()
}
