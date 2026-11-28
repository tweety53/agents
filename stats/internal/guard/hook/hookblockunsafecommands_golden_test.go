package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The block-unsafe-commands golden corpus: every payload below was fed to hooks/block-unsafe-commands.py
// as it stood before its port, against the fixture bucFixture builds, and what it printed — stdout,
// stderr and exit code — is testdata/block-unsafe-commands-golden.json, the fixture's root written
// {ROOT} and stdout's JSON document re-encoded by bucReencode, since the Go hook writes its JSON with
// encoding/json rather than Python's json.dumps spacing and ASCII escaping.
// TestHookBlockUnsafeCommandsGolden requires the Go hook to print the same bytes on every case.
//
// To capture it again (a case added), point BLOCK_UNSAFE_GOLDEN_PY at the Python original, recovered
// from git history:
//
//	git show e75c4fdab:hooks/block-unsafe-commands.py > /tmp/buc.py
//	BLOCK_UNSAFE_GOLDEN_PY=/tmp/buc.py go test ./internal/guard/ -run TestHookBlockUnsafeCommandsGolden
//
// which rewrites the JSON from Python instead of comparing, save the cases whose entry carries a
// `note`: those bytes are the Go hook's own — a deliberate departure from Python, the note saying
// why — and a capture keeps them as they stand. A case the corpus has no entry for fails.

const bucGoldenFile = "testdata/block-unsafe-commands-golden.json"

type bucCase struct{ name, stdin string }

type bucOut struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Exit   int    `json:"exit"`
	// Note marks a case whose bytes are the Go hook's own, never Python's, and says why.
	Note string `json:"note,omitempty"`
}

// bucPayload is a PreToolUse payload as Claude Code sends it; a nil cwd leaves the key out.
func bucPayload(tool string, input any, cwd any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	p := map[string]any{"tool_name": tool, "tool_input": input}
	if cwd != nil {
		p["cwd"] = cwd
	}
	if err := enc.Encode(p); err != nil {
		panic(err)
	}
	return buf.String()
}

func bucCases() []bucCase {
	var cases []bucCase
	seen := map[string]int{}
	add := func(name, stdin string) {
		seen[name]++
		if seen[name] > 1 {
			name = fmt.Sprintf("%s #%d", name, seen[name])
		}
		cases = append(cases, bucCase{name, stdin})
	}
	b := func(name, cwd, command string) {
		add(name, bucPayload("Bash", map[string]any{"command": command}, cwd))
	}
	w := func(name, tool, path string) {
		add(name, bucPayload(tool, map[string]any{"file_path": path, "content": "x"}, "{ROOT}"))
	}
	plan := "C={SIB}; cat > $C/t.md <<'EOF'\n> **Execution:** a > b; rm x\nEOF\n"
	xsh := "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n"
	rec := `{SIB}/scripts/flow-guard run-verify-list {SIB} lint kan-1 verify mf-abc123 haiku low`
	expansion := "export X=$Y\ncat > {SIB}/r.md <<'EOF'\nwrite f<i>.txt here\nEOF"

	// The repotest harness's cases (stats/internal/repotest/block_unsafe_commands_test.go), on this
	// corpus's one fixture.
	w("1 Write inside main checkout on main", "Write", "{MAIN}/new.txt")
	w("2 Edit inside linked worktree", "Edit", "{WT}/f.txt")
	w("13 Write outside any repo", "Write", "{ROOT}/loose.txt")
	w("15 no-origin repo on develop", "Write", "{ROOT}/noorigin/g.txt")
	w("31 loose file directly in .worktrees", "Write", "{MAIN}/.worktrees/loose.txt")
	w("44 sibling layout: Edit in an existing sibling worktree", "Edit", "{SIB}/f.txt")
	w("45 sibling layout: Write into a not-yet-existing sibling worktree", "Write", "{ROOT}/noorigin/inner-worktrees/new/x")
	w("47 sibling layout: <X>-worktrees beside a plain directory is main-checkout content", "Write", "{ROOT}/noorigin/lib-worktrees/new/x")
	w("3 Write in main checkout on a feature branch", "Write", "{ROOT}/off/new.txt")
	add("14 malformed JSON passes through", "not json")
	b("redirect after a cd group lands in main", "{MAIN}", `(cd {WT} && true) > c56.out 2> c56.out.fix`)
	b("redirect after a cd command substitution lands in main", "{MAIN}", `x=$(cd {WT} && pwd) && echo "$x" > f.txt`)
	b("a cd in a quoted substitution ends at its close", "{MAIN}", "echo \"$(cd {WT}; pwd)\"\necho x > f.txt")
	b("redirect inside the cd group lands in the worktree", "{MAIN}", `(cd {WT} && echo x > f.txt)`)
	b("redirect glued to a group's close lands in main", "{MAIN}", `(cd {WT} && true)>c.out`)
	b("append glued to a group's close lands in main", "{MAIN}", `(cd {WT} && true)>>c.out`)
	b("&> glued to a group's close lands in main", "{MAIN}", `(cd {WT} && true)&>c.out`)
	b("&& glued to a group's close ends the cd", "{MAIN}", `(cd {WT} && true)&& echo > c.out`)
	b("| glued to a group's close ends the cd", "{MAIN}", `(cd {WT} && true)| tee c.out`)
	b(";) closes the group", "{MAIN}", `(cd {WT}; true;) > c.out`)
	b("a cd before a group still holds after it", "{ROOT}", `cd {WT} && (true) > f.txt`)
	b("trash of a tracked file in main", "{ROOT}", `trash {MAIN}/f.txt`)
	b("trash of an untracked file in main", "{ROOT}", `trash {MAIN}/scratch.out`)
	b("4 Bash git commit with cwd main", "{MAIN}", `git commit -m x`)
	b("5 Bash git -C main add from elsewhere", "{ROOT}", `git -C {MAIN} add .`)
	b("6 Bash cd main && git reset", "{ROOT}", `cd {MAIN} && git reset HEAD~1`)
	b("7 Bash git -C main pull --ff-only", "{ROOT}", `git -C {MAIN} pull --ff-only`)
	b("8 Bash git -C main log", "{ROOT}", `git -C {MAIN} log --oneline -3`)
	b("9 Bash git -C worktree commit", "{ROOT}", `git -C {WT} commit -m x`)
	b("10 Bash sed -i into main", "{ROOT}", `sed -i '' s/a/b/ {MAIN}/f.txt`)
	b("11 Bash redirect into main", "{ROOT}", `echo x > {MAIN}/f.txt`)
	b("12 Bash cat from main", "{ROOT}", `cat {MAIN}/f.txt`)
	b("16 Bash git worktree add from main", "{ROOT}", `git -C {MAIN} worktree add {MAIN}/.worktrees/other -b other origin/main`)
	b("17 Bash git -C main stash", "{ROOT}", `git -C {MAIN} stash`)
	b("18 Bash tee into main", "{ROOT}", `echo x | tee {MAIN}/f.txt`)
	b("19 Bash rm in main", "{ROOT}", `rm {MAIN}/f.txt`)
	b("20 landing script named, not a git verb", "{ROOT}", `land-self-review-report.sh {MAIN} main 'docs: x' docs/x.md --push main`)
	b("21 tilde destination is not cwd-relative", "{MAIN}", `mv {ROOT}/elsewhere/x ~/.Trash/`)
	b("22 mv out of main checkout", "{ROOT}", `mv {MAIN}/f.txt {ROOT}/`)
	b("23 cd worktree; git commit", "{ROOT}", `cd {WT}; git commit -m x`)
	b("24 cd worktree; redirect to an unexpanded var", "{MAIN}", `S=/tmp/x; cd {WT}; npx playwright test a.spec.ts >$S/cap.log 2>&1; echo cap=$?`)
	b("25 cd main;git reset with separators touching", "{ROOT}", `cd {MAIN};git reset HEAD~1`)
	b("26 multi-line cp cannot swallow a later line's target", "{ROOT}", "cp notes.md draft.md\necho done\ngit add {MAIN}/README.md")
	b("27 multi-line second line judged on its own", "{ROOT}", "echo start\ncd {MAIN} && git reset HEAD~1")
	b("28 assignment-set redirect into main resolves and denies", "{ROOT}", `C={MAIN}; echo x > $C/f.txt`)
	b("29 nested variable stays let-through", "{ROOT}", `B={MAIN}; A=$B; echo x > $A/f.txt`)
	b("30 cp into a not-yet-existing worktree path", "{ROOT}", `cp /tmp/a.md {MAIN}/.worktrees/new-landing/spectre/foo.md`)
	b("32 cd state threads across lines without a separator", "{ROOT}", "cd {MAIN}\ngit reset HEAD~1")
	b("33 longest-name expansion wins the collision", "{ROOT}", `CA={MAIN}; C=/tmp; echo x > $CA/f.txt`)
	b("34 braced expansion resolves the variable", "{ROOT}", `C={MAIN}; echo x > ${C}/f.txt`)
	b("35 rm of the .worktrees root itself stays denied", "{ROOT}", `rm -rf {MAIN}/.worktrees`)
	b("36 unset near-name variable is not expanded greedily", "{ROOT}", `D={MAIN}/docs; echo x > $Dy/f.txt`)
	b("37 use before set is not expanded", "{ROOT}", "echo x > $C/f.txt\nC={MAIN}")
	b("43 BSD sed -i with an empty suffix, file outside", "{MAIN}", `sed -i '' 's/x/y/' {ROOT}/outside.txt`)
	b("48 heredoc append into a sibling worktree, body naming rm and >", "{MAIN}", "cat >> {SIB}/f.txt <<'EOF'\nthen rm the stale file and cp it > there\nEOF")
	b("49 rm of an absolute sibling-worktree path from main", "{MAIN}", `rm {SIB}/tmp`)
	b("50 relative rm with cwd main", "{MAIN}", `rm f.txt`)
	b("51 heredoc write into main", "{ROOT}", "cat > {MAIN}/f.txt <<EOF\nx\nEOF")
	b("52 a line after the heredoc is scanned again", "{MAIN}", "cat >> {SIB}/f.txt <<-EOF\n\tx\n\tEOF\nrm f.txt")
	b("53 heredoc fed to a shell stays scanned", "{ROOT}", "bash <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("54 a quoted << opens no heredoc", "{MAIN}", "echo \"<<EOF\"\nrm f.txt\nEOF")
	b("55 a commented << opens no heredoc", "{MAIN}", "true # <<EOF\nrm f.txt\nEOF")
	b("56 an arithmetic shift opens no heredoc", "{MAIN}", "echo $((1<<N))\nrm f.txt")
	b("57 a partly quoted delimiter is unquoted whole", "{MAIN}", "cat <<E\"OF\" >/dev/null\nx\nEOF\nrm f.txt")
	b("58 a shell named by path keeps its body scanned", "{MAIN}", "/bin/bash <<EOF\nrm f.txt\nEOF")
	b("59 a body sourced by . stays scanned", "{MAIN}", ". /dev/stdin <<EOF\nrm f.txt\nEOF")
	b("60 dash fed a heredoc keeps its body scanned", "{MAIN}", "dash <<EOF\nrm f.txt\nEOF")
	b("61 a heredoc piped into a shell keeps its body scanned", "{MAIN}", "cat <<EOF | dash\nrm f.txt\nEOF")
	b("62 $SHELL fed a heredoc keeps its body scanned", "{MAIN}", "$SHELL <<EOF\nrm f.txt\nEOF")
	b("63 a heredoc captured then eval'd keeps its body scanned", "{MAIN}", "x=$(cat <<EOF\nrm f.txt\nEOF\n)\neval \"$x\"")
	b("64 a heredoc read then eval'd keeps its body scanned", "{MAIN}", "read -r -d '' c <<EOF\nrm f.txt\nEOF\neval \"$c\"")
	b("65 a heredoc written to a script then run keeps its body scanned", "{MAIN}", "cat <<EOF > {SIB}/s.sh\nrm f.txt\nEOF\nsh {SIB}/s.sh")
	b("66 a CRLF closing line does not end the scan", "{MAIN}", "cat >> {SIB}/f.txt <<EOF\nx\nEOF\r\nrm f.txt")
	b("67 a << inside a parameter expansion opens no heredoc", "{MAIN}", "echo ${x:-<<EOF}\nrm f.txt\nEOF")
	b("68 tee -a heredoc into a sibling worktree, body naming rm", "{MAIN}", "tee -a {SIB}/f.txt <<'EOF'\nthen rm the stale file\nEOF")
	b("69 heredoc append into main", "{ROOT}", "cat >> {MAIN}/f.txt <<EOF\nx\nEOF")
	b("70 an unquoted heredoc body's command substitution stays scanned", "{MAIN}", "cat >> {SIB}/f.txt <<EOF\n$(rm f.txt)\nEOF")
	b("71 a report heredoc in a longer command keeps its body as data", "{MAIN}", "D={SIB}; cat > $D/r.md <<'EOF'\nthe row reads > the old one\nEOF\necho written")
	b("72 a report heredoc in a longer command still judges its target", "{ROOT}", "D={MAIN}; cat > $D/r.md <<'EOF'\nx\nEOF\necho written")
	b("73 a heredoc into a process substitution keeps its body scanned", "{ROOT}", "echo start; cat > >(sh) <<'EOF'\nrm {MAIN}/f.txt\nEOF\necho done")
	b("74 a tee heredoc feeding a process substitution keeps its body scanned", "{ROOT}", "echo start; tee /dev/null >(bash) <<'EOF'\nrm {MAIN}/f.txt\nEOF\necho done")
	b("75 a heredoc inside bash <( keeps its body scanned", "{ROOT}", "bash <(\ncat > /dev/stdout <<'EOF'\nrm {MAIN}/f.txt\nEOF\n)")
	b("76 a heredoc written then run under another spelling keeps its body scanned", "{ROOT}", "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\nsh \"{SIB}\"/x.sh")
	b("102 a symlink into the checkout, then .., resolves before judging", "{ROOT}", `cat > {ROOT}/elsewhere/lsub/../f.txt`)
	b("103 rm and mv of a symlink into the checkout act on the link, outside it", "{ROOT}", `rm {ROOT}/elsewhere/lsub2 && mv {ROOT}/elsewhere/lsub {ROOT}/elsewhere/l3`)
	b("110 an unquoted data heredoc's literal redirect beside a backtick span is data", "{MAIN}", "cat > {SIB}/r.md <<EOF\nsee `a` > b\nEOF")
	b("111 a backtick span in an unquoted data heredoc body is still scanned", "{MAIN}", "cat > {SIB}/r.md <<EOF\nsee `rm {MAIN}/x` here\nEOF")
	b("112 a backtick span is scanned as a command", "{ROOT}", "echo `rm {MAIN}/x`")
	b("113 a command substitution inside a double-quoted word is scanned", "{ROOT}", `echo "$(rm {MAIN}/x)"`)
	b("114 a backtick cp in an unquoted data heredoc body is scanned", "{ROOT}", "cat > {SIB}/r.md <<EOF\nsee `cp a {MAIN}/y` here\nEOF")
	b("115 a cp inside an unquoted $(...) is scanned", "{ROOT}", `echo $(cp a {MAIN}/y)`)
	b("116 a cp inside a subshell is scanned", "{ROOT}", `(cp a {MAIN}/y)`)
	b("117 an escaped backtick does not pair with a real span", "{MAIN}", "cat > {SIB}/r.md <<EOF\n\\` `rm {MAIN}/x` \\`\nEOF")
	b("118 a data heredoc's target named again as an argument keeps its body as data", "{MAIN}", plan+"check-references.sh $C/t.md")
	b("119 a path-named command naming another file keeps the body as data", "{MAIN}", plan+"~/.claude/skills/flow/scripts/lint.sh /x/o.md")
	b("120 a data heredoc's target run as a command keeps its body scanned", "{MAIN}", plan+"chmod +x $C/t.md && $C/t.md")
	b("121 a heredoc's file name run under another path keeps its body scanned", "{ROOT}", xsh+"/elsewhere/x.sh")
	for _, c := range []struct{ n, run string }{
		{"122", "/bin/bash {SIB}/x.sh"}, {"123", "/bin/sh {SIB}/x.sh"}, {"124", "cd {SIB} && x.sh"},
		{"125", "awk -f {SIB}/x.sh"}, {"126", "make -f {SIB}/x.sh"}, {"127", "watch {SIB}/x.sh"},
		{"128", "crontab {SIB}/x.sh"}, {"129", "php {SIB}/x.sh"}, {"130", "setsid {SIB}/x.sh"},
		{"131", "find {SIB} -name x.sh -exec {} ;"}, {"132", "ssh localhost {SIB}/x.sh"},
		{"133", "psql -f {SIB}/x.sh"}, {"134", "git -c core.pager=x show {SIB}/x.sh"},
		{"135", "rg --pre {SIB}/x.sh x f"},
	} {
		b(c.n+" a heredoc body then '"+c.run+"'", "{ROOT}", xsh+c.run)
	}
	b("136 a written check script run afterwards keeps its body scanned", "{ROOT}", "cat > {SIB}/check-a.sh <<EOF\nrm {MAIN}/f.txt\nEOF\n{SIB}/check-a.sh")
	b("137 a heredoc body then 'GIT_EXTERNAL_DIFF={SIB}/x.sh git diff' keeps its body scanned", "{ROOT}", xsh+"GIT_EXTERNAL_DIFF={SIB}/x.sh git diff --ext-diff a b")
	b("110 go test over the whole guard package", "{ROOT}", `cd {SIB} && go test ./internal/guard/`)
	b("111 go test over ./... with -count", "{ROOT}", `cd {SIB}/stats && go test -count=1 ./...`)
	b("112 go test of the guard package naming its tests", "{ROOT}", `cd {SIB}/stats && go test ./internal/guard/ -run 'PairJudge|PjSample'`)
	b("113 go test of a light package", "{ROOT}", `cd {SIB}/stats && go test ./internal/modelpolicy/`)
	b("143 run-verify-list recording form from the main session", "{ROOT}", rec)
	b("144 run-verify-list bare form from the main session", "{ROOT}", `{SIB}/scripts/flow-guard run-verify-list {SIB} lint`)
	add("145 recording form from a subagent (agent_id)", strings.Replace(bucPayload("Bash", map[string]any{"command": rec}, "{ROOT}"), "{", `{"agent_id":"a1b2c3",`, 1))
	add("146 recording form from a subagent (transcript)", strings.Replace(bucPayload("Bash", map[string]any{"command": rec}, "{ROOT}"), "{", `{"transcript_path":"/x/s/subagents/agent-a1.jsonl",`, 1))
	b("138 a quoted '>Save<' argument, cwd = main", "{MAIN}", `python3 grab.py "{SIB}/a.html" '>Save<' 1 1200`)
	b("139 a python heredoc body finding '>Save<', cwd = main", "{MAIN}", "python3 - <<'EOF'\nprint(open('{SIB}/a.html').read().find('>Save<'))\nEOF")
	b("140 '>|' into the main checkout", "{ROOT}", `echo x >| {MAIN}/f.txt`)
	b("141 '&>' into the main checkout", "{ROOT}", `echo x &> {MAIN}/f.txt`)
	b("142 an unbalanced quote, then '>' glued to a main-checkout path", "{ROOT}", `echo it's >{MAIN}/f.txt`)
	b("149 '>&' into the main checkout", "{ROOT}", `echo x >& {MAIN}/f.txt`)
	b("163 a python heredoc body comparing d>0, cwd = main", "{MAIN}", "python3 -I - x.png <<'EOF'\nif v<230 or d>0: print(y)\nEOF")
	b("164 a python heredoc body running git commit in main", "{ROOT}", "python3 - <<'EOF'\nsubprocess.run([\"git\", \"-C\", \"{MAIN}\", \"commit\", \"-m\", \"x\"])\nEOF")
	b("165 a report heredoc then echo of an unassigned name, cwd = main", "{MAIN}", "D={SIB}; cat > $D/r.md <<'EOF'\ncrop from `x-<n>.png` here\nEOF\necho $N")
	b("167 a python file written then run by python3, cwd = main", "{MAIN}", "O={SIB}; cat > $O/t.py <<'EOF'\nif a > 0: open(p, 'w')\nEOF\npython3 $O/t.py {SIB}")
	b("168 a spec written then run by npx playwright test, cwd = main", "{MAIN}", "cat > {SIB}/a.spec.ts <<'EOF'\nconst k = (x) => x > 0\nEOF\ncd {SIB} && U=http://x npx playwright test a.spec.ts")
	b("169 a python file written then run by bash keeps its body scanned", "{ROOT}", "cat > {SIB}/t.py <<'EOF'\nrm {MAIN}/f.txt\nEOF\nbash {SIB}/t.py")
	b("170 a report heredoc then a substitution running it keeps its body scanned", "{ROOT}", "cat > {SIB}/r.md <<'EOF'\nrm {MAIN}/f.txt\nEOF\nX=$({SIB}/r.md)")
	b("151 checkout -- of a path with unstaged edits", "{ROOT}", `git -C {WT} checkout -- f.txt`)
	b("152 restore of a path with unstaged edits", "{ROOT}", `git -C {WT} restore f.txt`)
	b("153 checkout HEAD -- after a cd", "{ROOT}", `cd {WT} && git checkout HEAD -- f.txt`)
	b("154 restore --staged touches the index only", "{ROOT}", `git -C {WT} restore --staged f.txt`)
	b("156 restore with stderr sent to /dev/null", "{ROOT}", `git -C {WT} restore f.txt 2>/dev/null`)
	b("157 checkout -- with stdout sent to /dev/null", "{ROOT}", `git -C {WT} checkout -- f.txt >/dev/null`)
	b("158 checkout HEAD -- of a staged-only edit", "{ROOT}", `git -C {WT} checkout HEAD -- g.txt`)
	b("159 restore --source=HEAD --staged --worktree of a staged-only edit", "{ROOT}", `git -C {WT} restore --source=HEAD --staged --worktree g.txt`)
	b("160 restore in a worktree the change state lists", "{ROOT}", `git -C {TRK} restore f.txt`)
	b("161 restore in the main checkout", "{ROOT}", `git -C {MAIN} restore f.txt`)
	b("162 restore of a drifted PNG baseline", "{ROOT}", `git -C {WT} restore b.png`)
	b("155 checkout -- of a clean path", "{ROOT}", `git -C {WT} checkout -- h.txt`)
	for _, c := range []string{"cat > /tmp/x.go", "cat >> f", "cat >f; echo", "a && cat > f", "false || cat > f", "(cat > f)",
		"cat > f <<'EOF'\nbody\nEOF", "cat > f < in", "x | cat > f", "cat a > f", "cat -n f > g", "echo cat > f", "concat > f"} {
		b("bare cat: "+c, "{ROOT}", c)
	}
	b("data body from main", "{MAIN}", "flow record summary -change x -file - <<'EOF'\nstep 1 -> step 2; rm {MAIN}/f.txt > {MAIN}/g.txt\nEOF")
	b("flow opener names a shell", "{MAIN}", "flow run csh <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("bash body", "{ROOT}", "bash <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("sudo sh body", "{ROOT}", "sudo sh <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("substituted sh body", "{ROOT}", "x=$(sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n)")
	b("unquoted body substitution", "{ROOT}", "flow record summary -change x -file - <<EOF\n$(rm {MAIN}/f.txt)\nEOF")
	b("data opener then bash script", "{ROOT}", "flow record summary -change x -file - <<'EOF'\nrm {MAIN}/f.txt\nEOF\nbash {SIB}/x.sh")
	b("quoted << opens no heredoc", "{ROOT}", "echo '<<EOF'\nrm {MAIN}/f.txt\nEOF")
	b("data opener piped to sh", "{ROOT}", "flow record summary -change x -file - <<'EOF' | sh\nrm {MAIN}/f.txt\nEOF")
	b("expanded shell opener", "{ROOT}", "$SHELL <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("commented << opens no heredoc", "{ROOT}", "x=1 # <<EOF\nrm {MAIN}/f.txt\nEOF")
	b("<< inside ${} opens no heredoc", "{ROOT}", "echo \"${x:-<<EOF}\"\nrm {MAIN}/f.txt\nEOF")
	b("escaped bash opener", "{ROOT}", "\\bash <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("csh opener", "{ROOT}", "csh <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("written then run by path", "{ROOT}", "envsubst <<'EOF' > /tmp/zz.sh && /tmp/zz.sh\nrm {MAIN}/f.txt\nEOF")
	b("flow output run by path", "{ROOT}", "flow x <<'EOF' > /tmp/zz.sh && /tmp/zz.sh\nrm {MAIN}/f.txt\nEOF")
	b("lines run as commands", "{ROOT}", "while read l; do $l; done <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	b("continued flow opener", "{ROOT}", "flow x <<'EOF' \\\n  | sh\nrm {MAIN}/f.txt\nEOF")
	b("unbalanced ) in body", "{ROOT}", "flow x <<EOF\n$(case a in a) rm {MAIN}/f.txt;; esac)\nEOF")
	b("two report heredocs, a lone dot and an unpaired backtick", "{MAIN}", "cat > {SIB}/a.md <<'E'\n.\nE\ncat > {SIB}/b.md <<'E'\n`> .\nE")
	b("a second body that runs the first's file keeps it scanned", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/y <<B\n$(bash {SIB}/x.sh)\nB")
	b("a second heredoc fed to a shell keeps the first scanned", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\nbash <<'B'\n{SIB}/x.sh\nB")
	b("a second quoted body running the first's file keeps it scanned", "{MAIN}", "cat > {SIB}/a.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/b.sh <<'B'\n{SIB}/a.sh\nB\nchmod +x {SIB}/*.sh; {SIB}/b.sh")
	b("a second quoted body substituted back keeps the first scanned", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/y <<'B'\n{SIB}/x.sh\nB\n$(cat {SIB}/y)")
	b("single-quoted backtick after a data heredoc", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho '`'")
	b("ANSI-quoted backtick after a data heredoc", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho $'\\'`'")
	b("unquoted backtick keeps the body scanned", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho `true`")
	b("double-quoted backtick keeps the body scanned", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho \"`true`\"")
	b("an apostrophe in double quotes opens no span", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho \"it's `true`\"")
	b("an unclosed quote reads as before", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho '`")
	b("an apostrophe in a comment opens no span", "{MAIN}", "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n: # it's\necho `bash {SIB}/x.sh` # '")
	b("an apostrophe in another heredoc body opens no span", "{MAIN}", "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\ncat > {SIB}/n <<'EOF2'\nit's\nEOF2\necho `bash {SIB}/x.sh` > {SIB}/o # '")
	b("a quote span crossing a newline reads as before", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\n# don't\necho `true` '")
	b("rm -rf of a whole local clone", "{ROOT}", `rm -rf {ROOT}/clone`)
	b("rm of a file inside a local clone", "{ROOT}", `rm {ROOT}/clone/f.txt`)
	b("rm -rf of a clone of a network origin", "{ROOT}", `rm -rf {ROOT}/netclone`)
	b("rm -rf of a checkout cloned from a local bare repository", "{ROOT}", `rm -rf {MAIN}`)
	b("export literal", "{MAIN}", "CR={SIB}\nexport FLOWD_PORT=4174\ncat > $CR/live-verification.md <<'EOF'\nwrite f<i>.txt here\nEOF")
	b("export names target", "{MAIN}", "export PAGER={SIB}/r.md\ncat > {SIB}/r.md <<'EOF'\nwrite f<i>.txt here\nEOF")
	b("export expansion", "{MAIN}", expansion)
	b("export without value", "{MAIN}", "export PAGER\ncat > {SIB}/r.md <<'EOF'\nwrite f<i>.txt here\nEOF")
	for _, c := range []string{"git stash pop", "git stash apply", "git stash drop", "git stash pop --index", "git stash push -- x && git stash pop",
		"git stash pop stash@{0}", "git stash apply stash@{0}", "git stash drop stash@{0}", "git stash branch nb stash@{0}", "git stash pop stash@{9}",
		"git stash pop 0", "git stash pop stash@{1}", "git stash apply stash@{1}", "git stash drop stash@{1}", "git stash branch nb stash@{1}",
		"git stash list", "git stash push -- f.txt", "git stash show stash@{0}"} {
		b("stash in WT: "+c, "{WT}", c)
	}

	// Payload shapes: each the hook lets through or judges by its own rule.
	add("empty stdin", "")
	add("a JSON list", `[1,2]`)
	add("no tool_input", `{"tool_name":"Bash","cwd":"{ROOT}"}`)
	add("tool_input a string", `{"tool_name":"Bash","tool_input":"rm {MAIN}/f.txt","cwd":"{ROOT}"}`)
	add("command not a string", bucPayload("Bash", map[string]any{"command": 5}, "{ROOT}"))
	add("empty command", bucPayload("Bash", map[string]any{"command": ""}, "{ROOT}"))
	add("unknown tool", bucPayload("Read", map[string]any{"file_path": "{MAIN}/f.txt"}, "{ROOT}"))
	add("MultiEdit in main", bucPayload("MultiEdit", map[string]any{"file_path": "{MAIN}/f.txt"}, "{ROOT}"))
	add("NotebookEdit notebook_path in main", bucPayload("NotebookEdit", map[string]any{"notebook_path": "{MAIN}/n.ipynb"}, "{ROOT}"))
	add("Edit empty file_path falls to notebook_path", bucPayload("Edit", map[string]any{"file_path": "", "notebook_path": "{MAIN}/n.ipynb"}, "{ROOT}"))
	add("Edit file_path a number", bucPayload("Edit", map[string]any{"file_path": 7}, "{ROOT}"))
	add("Write relative path, cwd main", bucPayload("Write", map[string]any{"file_path": "new.txt"}, "{MAIN}"))
	add("Write relative path, cwd worktree", bucPayload("Write", map[string]any{"file_path": "new.txt"}, "{WT}"))
	add("Write relative path, no cwd (process cwd is ROOT)", bucPayload("Write", map[string]any{"file_path": "main/new.txt"}, nil))
	add("Write relative path, empty cwd", bucPayload("Write", map[string]any{"file_path": "main/new.txt"}, ""))
	add("Write absolute path, cwd a number", bucPayload("Write", map[string]any{"file_path": "{MAIN}/new.txt"}, 12))
	add("Write relative path, cwd a number", bucPayload("Write", map[string]any{"file_path": "new.txt"}, 12))
	add("Bash rm, cwd a number", bucPayload("Bash", map[string]any{"command": "rm {MAIN}/f.txt"}, 12))
	add("Bash relative rm, cwd a number", bucPayload("Bash", map[string]any{"command": "rm f.txt"}, 12))
	add("Bash git commit, cwd a number", bucPayload("Bash", map[string]any{"command": "git commit -m x"}, 12))
	add("Bash bare stash pop, cwd a number", bucPayload("Bash", map[string]any{"command": "git stash pop"}, 12))
	add("Bash stash pop named, cwd a number", bucPayload("Bash", map[string]any{"command": "git stash pop stash@{1}"}, 12))
	add("Bash cat > file, cwd false", bucPayload("Bash", map[string]any{"command": "echo x > main/f.txt"}, false))
	w("Write a tilde path", "Write", "~/x.txt")
	w("Write a $VAR path", "Write", "$HOME/x.txt")
	w("Write into main's .git", "Write", "{MAIN}/.git/x")
	w("Write into main's subdirectory", "Write", "{MAIN}/sub/new.txt")
	w("Write into a missing main subtree", "Write", "{MAIN}/a/b/c.txt")
	w("Write into the sibling worktree root itself", "Write", "{ROOT}/main-worktrees")
	w("Write into a future sibling worktree", "Write", "{ROOT}/main-worktrees/new/x")
	w("Write into a future .worktrees worktree", "Write", "{MAIN}/.worktrees/new/x")
	w("Write in a no-head clone on main", "Write", "{ROOT}/nohead/x")
	w("Write in a master repo with no origin", "Write", "{ROOT}/mst/x")
	w("Write in a detached checkout", "Write", "{ROOT}/det/x")
	w("Write in a non-ASCII main checkout", "Write", "{ROOT}/mäin/x")
	w("Write through a symlink into main", "Write", "{ROOT}/elsewhere/lmain/x")
	w("Write through a symlink then ..", "Write", "{ROOT}/elsewhere/lsub/../x")
	w("Write with a .. into main", "Write", "{ROOT}/off/../main/x")
	w("Write with a double slash", "Write", "{MAIN}//x")
	w("Write to the root dir", "Write", "/")
	w("Write in the worktree's subdir", "Write", "{WT}/sub/x")

	// Each git verb in the main checkout, and the spellings around a git verb.
	for _, v := range []string{"add", "am", "apply", "cherry-pick", "clean", "commit", "merge", "mv", "rebase", "reset", "restore", "revert", "rm",
		"stash", "update-index", "update-ref", "checkout -b x", "switch -c x", "fetch", "push", "status", "diff", "pull --ff-only", "worktree add x"} {
		b("git "+v+" in main", "{MAIN}", "git "+v)
	}
	b("git -c then commit in main", "{ROOT}", "git -c user.name=x -C {MAIN} commit -m x")
	b("git --no-pager commit in main", "{MAIN}", "git --no-pager commit")
	b("git $OPTS commit", "{MAIN}", "git $OPTS commit")
	b("git -C $D commit", "{ROOT}", "git -C $D commit")
	b("$GIT push", "{MAIN}", "$GIT push")
	b("`which git` push", "{MAIN}", "`which git` push origin main")
	b("git -C relative commit", "{ROOT}", "git -C main commit -m x")
	b("git -C relative worktree commit", "{ROOT}", "git -C main/.worktrees/change commit -m x")
	b("git -C ~ commit", "{ROOT}", "git -C ~ commit")
	b("git -C main without verb", "{ROOT}", "git -C {MAIN}")
	b("git -C at end", "{ROOT}", "git -C")
	b("git commit in a quoted string", "{MAIN}", `echo "git commit -m x"`)
	b("git commit after && in main", "{ROOT}", "cd {MAIN} && echo hi && git add -A")
	b("git commit after cd to worktree in a subshell", "{MAIN}", "(cd {WT} && git commit -m x) && git status")
	b("git commit after subshell cd, cwd main", "{MAIN}", "(cd {WT}) ; git commit -m x")
	b("cd then cd - then commit", "{ROOT}", "cd {WT}; cd -; git commit")
	b("cd with no argument", "{MAIN}", "cd")
	b("cd ~ then commit", "{MAIN}", "cd ~ && git commit")
	b("cd $VAR then commit", "{MAIN}", "cd $W && git commit")
	b("cd relative into worktree then commit", "{MAIN}", "cd .worktrees/change && git commit")
	b("cd relative into main then commit", "{ROOT}", "cd main && git commit")
	b("cd to a missing dir then commit", "{ROOT}", "cd nowhere && git commit")
	b("git commit in a nested substitution", "{ROOT}", "echo $(echo $(git -C {MAIN} commit))")
	b("git commit in backticks", "{ROOT}", "echo `git -C {MAIN} commit`")
	b("an escaped backtick", "{ROOT}", "echo \\`git -C {MAIN} commit\\`")
	b("git commit after a comment", "{MAIN}", "echo x # git commit")
	b("git commit on a line after a comment", "{MAIN}", "echo x # note\ngit commit")
	b("a hash glued to a word", "{MAIN}", "echo x#y > f.txt")
	b("a hash inside a word then git", "{MAIN}", "echo a#b; git commit")
	b("continuation joins a git verb", "{MAIN}", "git \\\ncommit -m x")
	b("a quoted newline inside a line", "{MAIN}", "echo 'a\nb' > f.txt")
	b("a double-quoted escape", "{MAIN}", "echo \"a\\\"b\" > f.txt")
	b("a double-quoted backslash-dollar", "{MAIN}", "echo \"a\\$b\" > \"f.txt\"")
	b("an escaped space in a target", "{MAIN}", "echo x > my\\ file.txt")
	b("a trailing backslash", "{MAIN}", "echo x > f.txt \\")
	b("a quoted backslash at the end", "{MAIN}", "echo \"x\\")
	b("an unbalanced double quote", "{MAIN}", "echo \"x > f.txt")
	b("an unbalanced quote with 2>> glued", "{MAIN}", "echo it's 2>>f.txt")
	b("an unbalanced quote with &> glued", "{MAIN}", "echo it's &>f.txt")
	b("an unbalanced quote with >&1", "{MAIN}", "echo it's >&1")
	b("an unbalanced quote with >| glued", "{MAIN}", "echo it's >|f.txt")
	b("a quoted > is data", "{MAIN}", "echo '>' f.txt")
	b("a redirect glued to a word", "{MAIN}", "echo x>f.txt")
	b("a redirect with fd", "{MAIN}", "echo x 2>f.txt")
	b("a 2>&1 dup", "{MAIN}", "echo x 2>&1")
	b(">& 2 dups", "{MAIN}", "echo x >& 2")
	b(">& - closes", "{MAIN}", "echo x >& -")
	b(">& f writes", "{MAIN}", "echo x >& f.txt")
	b("a redirect at end with no target", "{MAIN}", "echo x >")
	b("an input redirect", "{MAIN}", "cat < f.txt")
	b("a here-string", "{MAIN}", "cat <<< f.txt")
	b("redirect to /dev/null in main", "{MAIN}", "echo x > /dev/null")
	b("redirect into the worktree from main", "{MAIN}", "echo x > {WT}/f.txt")
	b("redirect with $(...) target", "{MAIN}", "echo x > $(pwd)/f.txt")
	b("redirect glued after paren chars", "{MAIN}", "(echo x;)>>f.txt")
	b("a punctuation run ;;", "{MAIN}", "case x in x) echo;; esac > f.txt")
	b("a pipe-and", "{MAIN}", "echo x |& tee f.txt")
	b("a background job", "{MAIN}", "sleep 1 & echo x > f.txt")
	b("an or-list", "{MAIN}", "false || rm f.txt")
	b("sed -i in main", "{MAIN}", "sed -i 's/a/b/' f.txt")
	b("sed -i.bak in main", "{MAIN}", "sed -i.bak 's/a/b/' f.txt")
	b("sed -e then -i", "{MAIN}", "sed -e s/a/b/ -i f.txt")
	b("sed -i of a missing file", "{MAIN}", "sed -i s/a/b/ nope.txt")
	b("sed without -i", "{MAIN}", "sed s/a/b/ f.txt > /tmp/o")
	b("sed -i far after", "{MAIN}", "sed -e a -e b -e c -i f.txt")
	b("tee -a in main", "{MAIN}", "echo x | tee -a f.txt")
	b("tee two files", "{ROOT}", "echo x | tee /tmp/a {MAIN}/f.txt")
	b("tee then pipe", "{MAIN}", "echo x | tee /tmp/a | cat")
	b("cp into main", "{ROOT}", "cp a b {MAIN}/")
	b("cp -r into worktree", "{MAIN}", "cp -r a {WT}/")
	b("cp from main", "{MAIN}", "cp f.txt /tmp/f.txt")
	b("cp with no args", "{MAIN}", "cp")
	b("mv inside worktree, cwd main", "{MAIN}", "mv {WT}/a {WT}/b")
	b("rm -f relative in worktree", "{WT}", "rm -f f.txt")
	b("rm of a glob in main", "{MAIN}", "rm *.txt")
	b("rm $VAR in main", "{MAIN}", "rm $X")
	b("rm of a missing file in main", "{MAIN}", "rm nope.txt")
	b("rm of the main checkout's parent dir", "{ROOT}", "rm -rf {ROOT}/elsewhere")
	b("rm of a symlink to main", "{ROOT}", "rm {ROOT}/elsewhere/lmain")
	b("mv a symlink to main", "{ROOT}", "mv {ROOT}/elsewhere/lmain {ROOT}/elsewhere/lm2")
	b("rm of a whole local clone relative", "{ROOT}", "rm -rf clone")
	b("trash -v of a tracked file", "{MAIN}", "trash -v f.txt")
	b("trash of a missing file", "{MAIN}", "trash nope")
	b("trash in a subshell", "{ROOT}", "(trash {MAIN}/f.txt)")
	b("assignment prefix not expanded in its own command", "{ROOT}", "D={MAIN} rm $D/f.txt")
	b("assignment then separator expands", "{ROOT}", "D={MAIN}; rm $D/f.txt")
	b("assignment at end of line", "{ROOT}", "D={MAIN}\nrm $D/f.txt")
	b("two assignments then command", "{ROOT}", "A=1 D={MAIN} rm $D/f.txt")
	b("assignment with quotes", "{ROOT}", "D=\"{MAIN}\"; rm \"$D\"/f.txt")
	b("assignment referencing another", "{ROOT}", "A={MAIN}; B=$A/f.txt; rm $B")
	b("braced unknown var", "{ROOT}", "rm ${NOPE}/f.txt")
	b("export assignment", "{ROOT}", "export D={MAIN}; rm $D/f.txt")
	b("full go test ./internal/guard", "{ROOT}", "go test ./internal/guard")
	b("full go test ./cmd/flow", "{ROOT}", "go test ./cmd/flow/ -v")
	b("full go test -run=X ./...", "{ROOT}", "go test -run=X ./...")
	b("full go test -run X", "{ROOT}", "go test -run X ./...")
	b("full go test -runx", "{ROOT}", "go test -runx ./...")
	b("full go test with tabs", "{ROOT}", "go\ttest\t./...")
	b("full go test two spaces", "{ROOT}", "go  test ./...")
	b("gotest", "{ROOT}", "gotest ./...")
	b("xgo test", "{ROOT}", "xgo test ./...")
	b("go test of a subpackage", "{ROOT}", "go test ./internal/guard/x")
	b("go test quoted pattern", "{ROOT}", `go test "./..."`)
	b("go test then echo", "{ROOT}", "go test ./...;echo")
	b("go vet then go test -run", "{ROOT}", "go vet ./... && go test ./internal/guard -run X")
	b("go test across pipes", "{ROOT}", "go test -count=1 | grep ./...")
	b("go test across lines", "{ROOT}", "go test -v\n./...")
	b("go test of guard in a heredoc body", "{ROOT}", "cat > /tmp/x <<'EOF'\ngo test ./...\nEOF")
	b("go test in a comment", "{ROOT}", "echo # go test ./...")
	b("go test after é", "{ROOT}", "égo test ./...")
	b("go test after a digit", "{ROOT}", "1go test ./...")
	b("bare cat at line start of second line", "{ROOT}", "echo x\ncat > f")
	b("bare cat glued", "{ROOT}", "cat>f")
	b("bare cat with semicolon after", "{ROOT}", "cat > f;")
	b("bare cat before )", "{ROOT}", "(cat > f )")
	b("bare cat before ||", "{ROOT}", "cat > f || true")
	b("bare cat before |", "{ROOT}", "cat > f | x")
	b("bare cat before &&", "{ROOT}", "cat >> f && y")
	b("bare cat padded", "{ROOT}", "  cat > f  ")
	b("bare cat quoted target", "{ROOT}", `echo; cat > "f"`)
	b("bare cat with 2>", "{ROOT}", "cat > f 2>/dev/null")
	b("bare cat with trailing tab", "{ROOT}", "cat > f\t")
	b("bare cat in a heredoc body", "{ROOT}", "cat > /tmp/a <<'EOF'\ncat > f\nEOF")
	b("bare cat after &", "{ROOT}", "x & cat > f")
	b("bare cat with CR", "{ROOT}", "cat > f\r")
	// Go-only cases (a `note` in the corpus): the bare-cat rule reads only command text.
	b("bare cat in an unquoted heredoc body", "{ROOT}", "cat > /tmp/a <<EOF\ncat > f\nEOF")
	b("bare cat in a <<- heredoc body", "{ROOT}", "cat > /tmp/a <<-\"EOF\"\n\tcat > f\n\tEOF")
	b("bare cat after ; in a flow heredoc body", "{ROOT}", "flow record summary -file - <<'EOF'\nthen; cat > f\nEOF")
	b("bare cat in an unquoted heredoc body's substitution", "{ROOT}", "cat > /tmp/a <<EOF\n$(cat > f)\nEOF")
	b("bare cat in double quotes", "{ROOT}", `echo "cat > f"`)
	b("bare cat after ; in double quotes", "{ROOT}", `echo "a; cat > f"`)
	b("bare cat after ; in single quotes", "{ROOT}", `echo 'a; cat > f'`)
	b("bare cat on a line of a multi-line quote", "{ROOT}", "git commit -m 'a\ncat > f\n'")
	b("bare cat in a double-quoted substitution", "{ROOT}", `echo "$(cat > f)"`)
	b("bare cat in a backtick substitution", "{ROOT}", "echo `cat > f`")
	b("bare cat in bash -c", "{ROOT}", `bash -c 'cat > f'`)
	b("bare cat in sh -c", "{ROOT}", `sh -c "cat > f"`)
	b("bare cat after ;", "{ROOT}", "echo; cat > f")
	b("bare cat after a quoted word", "{ROOT}", `echo "x" && cat >> f`)
	b("cat with a heredoc after &&", "{ROOT}", "true && cat > f <<'EOF'\nx\nEOF")
	b("data write tee -a >/dev/null", "{ROOT}", "tee -a {MAIN}/f.txt >/dev/null <<'EOF'\nx\nEOF")
	b("data write into worktree", "{MAIN}", "cat > {WT}/f.txt <<'EOF'\nrm f.txt\nEOF")
	b("data write with trailing blank lines", "{MAIN}", "cat > {WT}/f.txt <<'EOF'\nrm f.txt\nEOF\n\n  \n")
	b("data write with trailing command", "{MAIN}", "cat > {WT}/f.txt <<'EOF'\nrm f.txt\nEOF\nrm f.txt")
	b("data write unclosed", "{MAIN}", "cat > {WT}/f.txt <<'EOF'\nrm f.txt")
	b("data write unquoted with $(", "{MAIN}", "cat > {WT}/f.txt <<EOF\n$(rm f.txt)\nEOF")
	b("data write double-quoted delim", "{MAIN}", "cat > {WT}/f.txt <<\"EOF\"\nrm f.txt\nEOF")
	b("data write with CRLF opener", "{MAIN}", "cat > {WT}/f.txt <<'EOF'\r\nrm f.txt\nEOF")
	b("data write with tabbed <<-", "{MAIN}", "cat > {WT}/f.txt <<-EOF\nrm f.txt\n\tEOF")
	b("heredoc <<- body into main", "{ROOT}", "cat > {MAIN}/f.txt <<-'EOF'\n\tx\n\tEOF")
	b("heredoc with relative target", "{MAIN}", "cat > f.txt <<'EOF'\nx\nEOF")
	b("heredoc tee with options then data", "{MAIN}", "echo a; tee -a {WT}/f.txt <<'EOF'\nrm f.txt\nEOF")
	b("heredoc tee without target", "{MAIN}", "echo a; tee <<'EOF'\nrm f.txt\nEOF")
	b("heredoc cat without target", "{MAIN}", "echo a; cat <<'EOF'\nrm f.txt\nEOF")
	b("heredoc cat to /dev/stderr", "{MAIN}", "echo a; cat > /dev/stderr <<'EOF'\nrm f.txt\nEOF")
	b("heredoc node program", "{MAIN}", "node <<'EOF'\nif (a > b) rm(x)\nEOF")
	b("heredoc python program with $( unquoted", "{MAIN}", "python3 <<EOF\nprint('$(rm f.txt)')\nEOF")
	b("heredoc python program unclosed", "{MAIN}", "python3 <<'EOF'\nprint(1) > x")
	b("heredoc python program naming git", "{ROOT}", "python3 - <<'EOF'\nimport os\nos.system('git -C {MAIN} reset --hard')\nEOF")
	b("heredoc python program written then run by python3 with git", "{ROOT}", "cat > {SIB}/t.py <<'EOF'\nos.system('git -C {MAIN} reset --hard')\nEOF\npython3 {SIB}/t.py")
	b("flow heredoc with assignment prefix", "{MAIN}", "X=1 flow record summary -file - <<'EOF'\nrm f.txt\nEOF")
	b("flow heredoc then a reader", "{MAIN}", "flow record summary -file - <<'EOF'\nrm f.txt\nEOF\ncat /tmp/x")
	b("flow heredoc with a redirect", "{MAIN}", "flow record summary -file - > /tmp/o <<'EOF'\nrm f.txt\nEOF")
	b("flow heredoc unquoted with backtick", "{MAIN}", "flow record summary -file - <<EOF\n`rm f.txt`\nEOF")
	b("flow heredoc with a comment before", "{MAIN}", "flow x # c <<'EOF'\nrm f.txt\nEOF")
	b("flow heredoc with path word", "{MAIN}", "flow x /tmp/a <<'EOF'\nrm f.txt\nEOF")
	b("two heredocs on one line", "{MAIN}", "cat > {WT}/a <<A; cat > {WT}/b <<B\nrm f.txt\nA\nrm f.txt\nB")
	b("heredoc after && in a longer command", "{MAIN}", "true && cat > {WT}/a <<'EOF'\nrm f.txt\nEOF")
	b("heredoc with an env-prefixed reader after", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nX=1 cat {WT}/a")
	b("heredoc then git diff naming target", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\ngit diff {WT}/a")
	b("heredoc then git apply naming target", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\ngit -C {WT} apply {WT}/a")
	b("heredoc then spectre validate", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nspectre validate {WT}/a")
	b("heredoc then flow-guard check", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nscripts/flow-guard check-references {WT}/a")
	b("heredoc then flow-guard non-check", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nflow-guard run {WT}/a")
	b("heredoc then a check script named as the target", "{MAIN}", "cat > {WT}/check-x.sh <<'EOF'\nrm f.txt\nEOF\ncheck-x.sh {WT}/check-x.sh")
	b("heredoc then do-loop reading target", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nfor x in 1; do cat {WT}/a; done")
	b("heredoc then export of a path", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nexport P=/x/y")
	b("heredoc then export of a literal", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nexport P=1 Q=2")
	b("heredoc then env", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\nenv")
	b("heredoc then process substitution", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\ndiff <(cat x) y")
	b("heredoc then nested $(", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\necho $(echo $(ls))")
	b("heredoc then $'...' ANSI", "{MAIN}", "cat > {WT}/a <<'EOF'\nrm f.txt\nEOF\necho $'a\\'b'")
	b("heredoc then npx playwright test then bash", "{MAIN}", "cat > {WT}/a.spec.ts <<'EOF'\nrm f.txt\nEOF\nnpx playwright test {WT}/a.spec.ts bash")
	b("heredoc unquoted span unclosed", "{MAIN}", "cat > {WT}/a <<EOF\n$(rm f.txt\nEOF")
	b("heredoc unquoted odd backticks", "{MAIN}", "cat > {WT}/a <<EOF\n` rm f.txt\nEOF")
	b("heredoc with ${ before <<", "{MAIN}", "echo ${a} && cat > {WT}/a <<'EOF'\nrm f.txt\nEOF")
	b("cd inside heredoc body", "{ROOT}", "cat > {WT}/a <<'EOF'\ncd {MAIN}\nEOF\ngit commit")
	b("logical line continuation across rm", "{ROOT}", "rm \\\n{MAIN}/f.txt")
	b("unicode: non-ASCII word then redirect", "{MAIN}", "echo é > f.txt")
	b("unicode: NBSP before redirect", "{MAIN}", "echo x\u00a0> f.txt")
	b("unicode: ideographic space in a git verb", "{MAIN}", "git\u3000commit")
	b("unicode: target with non-ASCII", "{ROOT}", "echo x > {ROOT}/mäin/f.txt")
	b("unicode: unit separator", "{MAIN}", "echo x\x1f> f.txt")
	b("unicode: stash ref with an Arabic digit", "{WT}", "git stash pop stash@{١}")
	b("unicode: Arabic digit as a bare ref", "{WT}", "git stash pop ١")
	b("unicode: superscript digit before a redirect", "{MAIN}", "echo x ²> f.txt")
	// The checks that replace Python's lookarounds and `\b`, at their edges.
	b("post-check: stash verb running into a word", "{WT}", "git stash popx")
	b("post-check: stash verb then a dash", "{WT}", "git stash pop-x")
	b("post-check: git after a letter", "{WT}", "égit stash pop")
	b("post-check: git after an underscore", "{WT}", "_git stash pop")
	b("post-check: stash with NBSP separators", "{WT}", "git stash pop")
	b("post-check: go test with an ideographic space", "{ROOT}", "go　test ./...")
	b("post-check: go test then a word char", "{ROOT}", "go testx ./...")
	b("post-check: heavy package then NBSP", "{ROOT}", "go test ./... -v")
	b("post-check: bare cat then NBSP and a word", "{ROOT}", "cat > f x")
	b("post-check: bare cat then a line separator", "{ROOT}", "cat > f ")
	b("post-check: bare cat then spaces and a newline", "{ROOT}", "cat > f  \n x")
	b("post-check: bare cat twice, first fails", "{ROOT}", "cat > a b; cat > c")
	add("tool_name a list", `{"tool_name":["Bash"],"tool_input":{"command":"git commit"},"cwd":"{MAIN}"}`)
	b("stash: -C main checkout", "{ROOT}", "git -C {MAIN} stash drop stash@{1}")
	b("stash: -c flag", "{WT}", "git -c a=b stash pop stash@{1}")
	b("stash: two commands", "{WT}", "git stash pop stash@{1}; git stash pop")
	b("stash: apply by number", "{WT}", "git stash apply 1")
	b("stash: branch with number", "{WT}", "git stash branch nb 1")
	b("stash: in a subshell", "{WT}", "(git stash pop stash@{1})")
	b("stash: no repo", "{ROOT}", "git stash pop stash@{0}")
	b("stash: -C $VAR", "{WT}", "git -C $W stash pop stash@{1}")
	b("stash: other worktree pops its own", "{ROOT}/main/.worktrees/other", "git stash pop stash@{0}")
	b("stash: words after ref", "{WT}", "git stash pop --index stash@{1} x")
	b("discard: restore -s HEAD", "{ROOT}", "git -C {WT} restore -s HEAD f.txt")
	b("discard: restore --source HEAD", "{ROOT}", "git -C {WT} restore --source HEAD f.txt")
	b("discard: restore -W --staged", "{ROOT}", "git -C {WT} restore -W --staged f.txt")
	b("discard: restore -S", "{ROOT}", "git -C {WT} restore -S f.txt")
	b("discard: restore -- path", "{ROOT}", "git -C {WT} restore -- f.txt")
	b("discard: restore with no paths", "{ROOT}", "git -C {WT} restore")
	b("discard: checkout without --", "{ROOT}", "git -C {WT} checkout f.txt")
	b("discard: checkout -- of two paths", "{ROOT}", "git -C {WT} checkout -- f.txt b.png")
	b("discard: checkout -- of the png", "{ROOT}", "git -C {WT} checkout -- b.png")
	b("discard: checkout in a tracked worktree", "{ROOT}", "git -C {TRK} checkout -- f.txt")
	b("discard: restore then && in the same command", "{ROOT}", "git -C {WT} restore f.txt && echo ok")
	b("discard: restore in $W", "{ROOT}", "git -C $W restore f.txt")
	b("discard: restore with 2>&1", "{ROOT}", "git -C {WT} restore f.txt 2>&1")
	b("discard: restore in a sibling worktree", "{ROOT}", "git -C {SIB} restore f.txt")
	b("discard: restore relative cwd", "{WT}", "git restore f.txt")
	b("discard and main write together", "{ROOT}", "git -C {WT} restore f.txt; rm {MAIN}/f.txt")
	b("bare cat and stash together", "{WT}", "cat > f; git stash pop")
	b("stash and full go test together", "{WT}", "go test ./...; git stash pop")
	b("refresh-main-checkout landing script", "{ROOT}", "flow-guard refresh-main-checkout {MAIN}")
	b("a heredoc mention outside heredoc deny", "{MAIN}", "echo x > f.txt # <<EOF")
	b("<< inside quotes and a write", "{MAIN}", "echo '<<EOF' > f.txt")
	return cases
}

// bucFixture builds the corpus's repositories under one root and returns the root and the
// environment both hooks run with.
func bucFixture(t *testing.T) (string, []string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := root + "/home"
	env := []string{"HOME=" + home, "PATH=" + root + "/bin:" + os.Getenv("PATH"), "GIT_CONFIG_NOSYSTEM=1", "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8"}
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if k != "HOME" && k != "PATH" && k != "LANG" && k != "LC_ALL" && !strings.HasPrefix(k, "GIT_") {
			env = append(env, kv)
		}
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(p, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mkdir := func(p string) {
		t.Helper()
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	main := root + "/main"
	wt := main + "/.worktrees/change"
	mkdir(home)
	write(root+"/bin/flow", "#!/bin/sh\necho '{\"worktrees\":{\""+root+"/main-worktrees/tracked\":null}}'\n")
	git("init", "-q", "--bare", root+"/origin.git")
	git("clone", "-q", root+"/origin.git", main)
	write(main+"/f.txt", "a\n")
	write(main+"/g.txt", "g\n")
	write(main+"/h.txt", "h\n")
	write(main+"/b.png", "base\n")
	git("-C", main, "add", ".")
	git("-C", main, "commit", "-q", "-m", "init")
	git("-C", main, "branch", "-M", "main")
	git("-C", main, "push", "-q", "-u", "origin", "main")
	git("-C", main, "remote", "set-head", "origin", "main")
	mkdir(main + "/sub")
	write(main+"/scratch.out", "x\n")
	git("-C", main, "worktree", "add", "-q", wt, "-b", "change")
	git("-C", main, "worktree", "add", "-q", main+"/.worktrees/other", "-b", "other")
	git("-C", main, "worktree", "add", "-q", root+"/main-worktrees/sib", "-b", "sib")
	git("-C", main, "worktree", "add", "-q", root+"/main-worktrees/tracked", "-b", "tracked")
	// stash@{0} is the one made on `other`, stash@{1} the one made on `change`.
	write(wt+"/f.txt", "change wip\n")
	git("-C", wt, "stash", "push", "-q", "-m", "mine")
	write(main+"/.worktrees/other/f.txt", "other wip\n")
	git("-C", main+"/.worktrees/other", "stash", "push", "-q", "-m", "theirs")
	write(wt+"/f.txt", "flipped\n")
	write(wt+"/b.png", "drifted\n")
	write(wt+"/g.txt", "staged\n")
	git("-C", wt, "add", "g.txt")
	mkdir(wt + "/sub")
	write(root+"/main-worktrees/tracked/f.txt", "flipped\n")
	write(root+"/main-worktrees/sib/f.txt", "flipped\n")
	git("init", "-q", "-b", "develop", root+"/noorigin")
	write(root+"/noorigin/f.txt", "a\n")
	git("-C", root+"/noorigin", "add", "f.txt")
	git("-C", root+"/noorigin", "commit", "-q", "-m", "init")
	git("init", "-q", "-b", "develop", root+"/noorigin/inner")
	for _, d := range []string{"/noorigin/inner-worktrees", "/noorigin/lib", "/noorigin/lib-worktrees", "/other-worktrees/sib"} {
		mkdir(root + d)
	}
	write(root+"/elsewhere/x", "")
	for _, l := range []string{"/elsewhere/lsub", "/elsewhere/lsub2"} {
		if err := os.Symlink(main+"/sub", root+l); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(main, root+"/elsewhere/lmain"); err != nil {
		t.Fatal(err)
	}
	write(root+"/outside.txt", "x\n")
	git("clone", "-q", "--local", main, root+"/clone")
	git("clone", "-q", "--local", main, root+"/netclone")
	git("-C", root+"/netclone", "remote", "set-url", "origin", "https://example.invalid/x.git")
	git("clone", "-q", root+"/origin.git", root+"/off")
	git("-C", root+"/off", "checkout", "-q", "-b", "feature")
	git("clone", "-q", root+"/origin.git", root+"/mäin")
	git("clone", "-q", root+"/origin.git", root+"/nohead")
	git("-C", root+"/nohead", "remote", "set-head", "origin", "-d")
	git("init", "-q", "-b", "master", root+"/mst")
	write(root+"/mst/f.txt", "a\n")
	git("-C", root+"/mst", "add", "f.txt")
	git("-C", root+"/mst", "commit", "-q", "-m", "init")
	git("clone", "-q", root+"/origin.git", root+"/det")
	git("-C", root+"/det", "checkout", "-q", "--detach")
	return root, env
}

func bucExpand(s, root string) string {
	return strings.NewReplacer("{ROOT}", root, "{MAIN}", root+"/main", "{WT}", root+"/main/.worktrees/change",
		"{SIB}", root+"/main-worktrees/sib", "{TRK}", root+"/main-worktrees/tracked").Replace(s)
}

// bucRunGo runs the Go hook in-process with env as its whole environment and root as its cwd.
func bucRunGo(stdin, root string, env []string) bucOut {
	m := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	var out, errb bytes.Buffer
	e := core.Env{
		Getenv:    func(k string) string { return m[k] },
		LookupEnv: func(k string) (string, bool) { v, ok := m[k]; return v, ok },
		Environ:   func() []string { return env },
		Dir:       root,
		Stdin:     strings.NewReader(stdin),
	}
	rc := core.Registry["hook"]([]string{"block-unsafe-commands"}, e, &out, &errb)
	return bucOut{Stdout: out.String(), Stderr: errb.String(), Exit: rc}
}

// bucReencode is a Python-captured stdout's JSON document as hookJSON encodes the same value — the
// Go hook's spacing and escaping, never Python's; stdout that is no JSON document stays as it is.
func bucReencode(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return s
	}
	return hookJSON(v)
}

func bucRunPython(t *testing.T, py, stdin, root string, env []string) bucOut {
	t.Helper()
	cmd := exec.Command("python3", py)
	cmd.Dir = root
	cmd.Env = env
	cmd.Stdin = strings.NewReader(stdin)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	rc := 0
	if ee, ok := err.(*exec.ExitError); ok {
		rc = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return bucOut{Stdout: out.String(), Stderr: errb.String(), Exit: rc}
}

// TestHookBlockUnsafeCommandsGolden: the Go hook prints, on every corpus case, the bytes the Python
// hook printed (see the file comment).
func TestHookBlockUnsafeCommandsGolden(t *testing.T) {
	t.Parallel()
	root, env := bucFixture(t)
	cases := bucCases()
	unroot := func(o bucOut) bucOut {
		o.Stdout = strings.ReplaceAll(o.Stdout, root, "{ROOT}")
		o.Stderr = strings.ReplaceAll(o.Stderr, root, "{ROOT}")
		return o
	}
	if py := os.Getenv("BLOCK_UNSAFE_GOLDEN_PY"); py != "" {
		golden := map[string]bucOut{}
		if data, err := os.ReadFile(bucGoldenFile); err == nil {
			if err := json.Unmarshal(data, &golden); err != nil {
				t.Fatal(err)
			}
		}
		for _, c := range cases {
			if golden[c.name].Note != "" {
				continue // the Go hook's own bytes, never Python's
			}
			o := unroot(bucRunPython(t, py, bucExpand(c.stdin, root), root, env))
			o.Stdout = bucReencode(o.Stdout)
			golden[c.name] = o
		}
		data, err := json.MarshalIndent(golden, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(bucGoldenFile, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	data, err := os.ReadFile(bucGoldenFile)
	if err != nil {
		t.Fatal(err)
	}
	var golden map[string]bucOut
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for _, c := range cases {
		want, ok := golden[c.name]
		if !ok {
			t.Errorf("%s: no golden entry — capture the corpus again (file comment)", c.name)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			got := unroot(bucRunGo(bucExpand(c.stdin, root), root, env))
			want.Note = ""
			if got != want {
				mu.Lock()
				t.Errorf("%s\n got: %#v\nwant: %#v", c.name, got, want)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
}
