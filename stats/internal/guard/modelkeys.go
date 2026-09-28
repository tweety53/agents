package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func init() {
	Registry["check-model-keys"] = checkModelKeys
}

// checkModelKeys is scripts/check-model-keys.sh: that script's header
// comment carries the contract.
//
// Divergences from the bash, all on input the bash itself mishandled: under
// a UTF-8 locale macOS awk aborts ("towc: multibyte conversion failure") on
// a project.md or settings.go line that is not valid UTF-8, tr refuses such
// bytes in the CLI's stderr and sed aborts on a NUL there; this port reads
// all three as bytes. The CLI's stderr is buffered in memory, never a mktemp
// file, and the CLI runs with no stdin rather than the guard's.
func checkModelKeys(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-model-keys: "+format+"\n", a...)
		return 2
	}
	// REPO_ROOT is the guard script's own checkout. The bash derived it from
	// BASH_SOURCE; this binary lives in a cache, so its shim exports
	// FLOW_GUARD_REPO_ROOT, derived the same way.
	repo := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if repo == "" {
		return die("FLOW_GUARD_REPO_ROOT is unset — run scripts/check-model-keys.sh, which sets it")
	}
	settings := repo + "/stats/internal/store/settings.go"
	hasSource := gsReadable(settings)
	var source []string
	if hasSource {
		source = mkSourceSet(settings)
	}
	utf8 := smcUTF8(env)

	valid, note := mkAsk(env, utf8, stderr)
	switch {
	case len(valid) == 0:
		fmt.Fprintf(stderr, "check-model-keys: no answer from flow settings models%s — falling back to parsing %s\n", note, settings)
		if !hasSource {
			return die("cannot read %s", settings)
		}
		if len(source) == 0 {
			return die("no ValidModels members found in %s", settings)
		}
		valid = source
	case len(source) > 0:
		// The CLI answered and the source is readable: compare the two.
		// Drift means the installed binary predates the checkout — the
		// check-installed-rules.sh class of failure — and is named with each
		// side's unique members. On drift the source set governs the verdict
		// below, so a stale install cannot fail a value the checkout
		// declares; the announcement is what tells the operator to rebuild
		// flow.
		drift, err := mkDrift(env, valid, source)
		if err != nil {
			return die("cannot compare the CLI's ValidModels set with %s: %v", settings, err)
		}
		if len(drift) > 0 {
			fmt.Fprintf(stderr, "check-model-keys: installed flow's ValidModels set disagrees with %s — rebuild or reinstall flow; left column only-in-CLI, right column only-in-source:\n", settings)
			fmt.Fprintf(stderr, "%s\n", strings.Join(drift, "\n"))
			fmt.Fprintln(stderr, "check-model-keys: the source set governs this run's verdict while the install is stale")
			valid = source
		}
	case hasSource:
		fmt.Fprintf(stderr, "check-model-keys: note — %s yielded no ValidModels members; the source cross-check is off (did the map literal's shape change?)\n", settings)
	}

	roots := args
	if len(roots) == 0 {
		roots = []string{repo}
	}
	violations, checked := 0, 0
	for _, root := range roots {
		abs := smcAbs(env, root)
		if root == "" || !isDir(abs) {
			return die("not a directory: %s", root)
		}
		if !gsReadable(abs) {
			return die("cannot read directory: %s", root)
		}
		pf := root + "/.flow/project.md"
		if _, err := os.Stat(abs + "/.flow/project.md"); err != nil {
			fmt.Fprintf(stdout, "MODEL-KEYS-OK: %s — no .flow/project.md (nothing to check)\n", root)
			checked++
			continue
		}
		if !isFile(abs + "/.flow/project.md") {
			return die("not a regular file: %s", pf)
		}
		if !gsReadable(abs + "/.flow/project.md") {
			return die("cannot read: %s", pf)
		}
		const key = "self review model"
		// The value is the body's head (KAN-797): mkSectionBody has already
		// trimmed each line and stripped its surrounding backticks, so the
		// head is the first line that is not empty after that — lines below
		// it are documentation, never read.
		if body := mkSectionBody(abs+"/.flow/project.md", key, utf8); body != "" {
			head := ""
			for _, line := range strings.Split(body, "\n") {
				if line != "" {
					head = line
					break
				}
			}
			if head != "" && !mkMember(valid, head) {
				fmt.Fprintf(stdout, "%s: `## %s` value %s is not a ValidModels member\n", pf, key, smcQuote(head, utf8))
				violations++
			}
		}
		checked++
	}
	if violations > 0 {
		fmt.Fprintf(stderr, "check-model-keys: %d violation(s) across %d project(s) checked\n", violations, checked)
		return 1
	}
	fmt.Fprintf(stdout, "MODEL-KEYS-OK: %d project(s) checked\n", checked)
	return 0
}

func mkMember(set []string, s string) bool {
	for _, m := range set {
		if m == s {
			return true
		}
	}
	return false
}

// mkQuoted is awk's match($0, /"[^"]+"/): the first quoted run on a line.
var mkQuoted = regexp.MustCompile(`"[^"]+"`)

// mkSourceSet is source_set, the map-literal extraction shared by the
// offline fallback and the cross-check: the quoted keys between `var
// ValidModels = map[string]bool{` and its closing `}`, each of shape
// `"name": true,`. As awk read it, a line ends at its first NUL, and a file
// that cannot be read (a directory) yields nothing, silently.
func mkSourceSet(file string) []string {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var set []string
	grabbing := false
	for _, line := range lines(b) {
		line, _, _ = strings.Cut(line, "\x00")
		switch {
		case strings.HasPrefix(line, "var ValidModels = map[string]bool{"):
			grabbing = true
		case grabbing && strings.HasPrefix(line, "}"):
			return set
		case grabbing:
			if m := mkQuoted.FindString(line); m != "" {
				set = append(set, m[1:len(m)-1])
			}
		}
	}
	return set
}

// mkAsk asks the installed `flow` CLI first — its `settings models`
// subcommand prints the compiled-in set with no daemon needed. Its stderr is
// relayed to this guard's stderr rather than discarded, so a warn-but-answer
// ask is still visible, and the fallback announcement can say why the ask
// failed, not only that it did. It returns the answer and that
// announcement's note.
func mkAsk(env Env, utf8 bool, stderr io.Writer) ([]string, string) {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return nil, " — no flow on PATH"
	}
	var out, errb bytes.Buffer
	cmd := exec.Command(flow, "settings", "models")
	cmd.Dir, cmd.Stdout, cmd.Stderr = env.Dir, &out, &errb
	// The CLI's exit status is not read; a CLI that cannot start reports
	// why the way a failed exec would, on its stderr.
	if err := cmd.Run(); err != nil {
		if _, exited := err.(*exec.ExitError); !exited {
			fmt.Fprintf(&errb, "%v\n", err)
		}
	}
	// `while IFS= read -r m`: NULs dropped, an unterminated last line never
	// read, empty lines skipped, nothing else trimmed.
	var valid []string
	answer := strings.Split(strings.ReplaceAll(out.String(), "\x00", ""), "\n")
	for _, m := range answer[:len(answer)-1] {
		if m != "" {
			valid = append(valid, m)
		}
	}
	switch {
	case errb.Len() > 0:
		// `tr '\n' ' ' | sed 's/[[:space:]]*$//'`.
		note := " — " + strings.TrimRightFunc(strings.ReplaceAll(errb.String(), "\n", " "), mkSpace(utf8))
		// Relay only when the ask answered: on a failed ask the fallback
		// announcement already embeds the same diagnostic, and one emission
		// per stderr stream is the difference between a warning and noise.
		if len(valid) > 0 {
			fmt.Fprintf(stderr, "check-model-keys: flow settings models reported:%s\n", note)
		}
		return valid, note
	case len(valid) == 0:
		return nil, " — empty answer"
	}
	return valid, ""
}

// mkDrift is `comm -3 <(… | sort -u) <(… | sort -u)`: the members only the
// CLI answered, then — tab-prefixed — those only the source declares, in the
// caller's collation. Both sort -u runs are exec'd (crSort) so duplicates
// fold as the locale folds them; comm's merge of two sorted lists is the
// collation order of their symmetric difference, which a third exec'd sort
// gives.
func mkDrift(env Env, cli, source []string) ([]string, error) {
	a, err := crSort(env, cli, true)
	if err != nil {
		return nil, err
	}
	b, err := crSort(env, source, true)
	if err != nil {
		return nil, err
	}
	onlySource := map[string]bool{}
	var diff []string
	for _, m := range a {
		if !mkMember(b, m) {
			diff = append(diff, m)
		}
	}
	for _, m := range b {
		if !mkMember(a, m) {
			diff = append(diff, m)
			onlySource[m] = true
		}
	}
	if diff, err = crSort(env, diff, false); err != nil {
		return nil, err
	}
	for i, m := range diff {
		if onlySource[m] {
			diff[i] = "\t" + m
		}
	}
	return diff, nil
}

// mkSpace is awk's and sed's [[:space:]]: the ASCII whitespace but newline,
// plus U+00A0 under a UTF-8 locale (measured on macOS; U+0085, U+2000–U+200A
// and U+3000 are not matched).
func mkSpace(utf8 bool) func(rune) bool {
	return func(r rune) bool { return strings.ContainsRune(gdcSpace, r) || (utf8 && r == '\u00a0') }
}

// mkSectionBody is extract_section_body: the trimmed body of the first
// `## <heading>` section (everything up to the next `^## ` heading or EOF),
// or "" if the heading is absent. Leading/trailing blank lines are stripped,
// matching project-configuration.md's own "trimmed" rule for these keys'
// single-line-literal bodies. Surrounding backticks are also stripped,
// matching every actual `.flow/project.md` in this repository (`## jira`,
// `## default landing route`, `## self review model` all write their
// single-line-literal value as a markdown code span) and jira-followups.md's
// own "strip surrounding backticks" rule for the same kind of body. As awk
// read it, a line ends at its first NUL; trailing lines a stripped code span
// leaves empty are cut, as the command substitution cut them.
func mkSectionBody(file, heading string, utf8 bool) string {
	section := projectSection(file, heading)
	if section == "" {
		return ""
	}
	space := mkSpace(utf8)
	body := strings.Split(section, "\n")
	for i, line := range body {
		body[i], _, _ = strings.Cut(line, "\x00")
	}
	for len(body) > 0 && strings.TrimFunc(body[0], space) == "" {
		body = body[1:]
	}
	for len(body) > 0 && strings.TrimFunc(body[len(body)-1], space) == "" {
		body = body[:len(body)-1]
	}
	for i, line := range body {
		line = strings.TrimFunc(line, space)
		if len(line) >= 2 && line[0] == '`' && line[len(line)-1] == '`' {
			line = line[1 : len(line)-1]
		}
		body[i] = line
	}
	return strings.TrimRight(strings.Join(body, "\n"), "\n")
}
