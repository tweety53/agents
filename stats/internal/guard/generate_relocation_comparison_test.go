package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestGenerateRelocationComparison ports
// scripts/test-generate-relocation-comparison.sh at bbfe685e: one subtest per
// PASS: label, each case a real git repository whose merge-base commit and
// worktree the generator reads in-process, the comparison read back from the
// file it writes. The shim's own argument handling execs the real shim.
func TestGenerateRelocationComparison(t *testing.T) {
	t.Parallel()
	type check struct{ label, fail string }
	var checks []check
	want := func(label string, ok bool, r guardResult, table string) {
		c := check{label: label}
		if !ok {
			c.fail = fmt.Sprintf("rc=%d out=%q table=%q", r.rc, r.out, table)
		}
		checks = append(checks, c)
	}
	env := Env{Getenv: os.Getenv}
	// repo commits base as the merge-base, then writes after over the
	// worktree (a nil body deletes the path) -- the harness's second commit
	// changes nothing the generator reads.
	repo := func(base, after map[string]string) (string, string) {
		dir := t.TempDir()
		gitRun(t, dir, "init", "-q")
		for p, b := range base {
			writeFile(t, dir+"/"+p, b)
		}
		gitRun(t, dir, "add", "-A")
		gitRun(t, dir, "commit", "-qm", "base")
		sha := tcfGit(t, dir, "rev-parse", "HEAD")
		for p, b := range after {
			if err := os.RemoveAll(dir + "/" + p); err != nil {
				t.Fatal(err)
			}
			if b != "\x00" {
				writeFile(t, dir+"/"+p, b)
			}
		}
		return dir, sha
	}
	plan := func(relocation string, files ...string) string {
		s := "# fixture plan\n\n> **Relocation:** " + relocation + "\n\n- [ ] 1. Only task\n\n**Files:**\n"
		for _, f := range files {
			s += "- Modify: `" + f + "`\n"
		}
		return s + "\n"
	}
	// gen runs the generator on dir with tasks as dir/tasks.md (none when
	// empty) and returns its result and the table it wrote ("" for none).
	gen := func(dir, sha, tasks string) (guardResult, string) {
		if tasks != "" {
			writeFile(t, dir+"/tasks.md", tasks)
		}
		r := runGuard("generate-relocation-comparison", []string{dir, dir, sha}, env)
		b, err := os.ReadFile(dir + "/.superpowers/sdd/relocation-comparison.md")
		if err != nil {
			return r, ""
		}
		return r, string(b)
	}
	// moved builds the harness's two-file move: text leaves src under srcH
	// and lands in dst under dstH, the other file's content trading places.
	moved := func(src, srcH, dst, dstH, text, filler string) (string, string) {
		return repo(map[string]string{src: srcH + "\n\n" + text + "\n\n", dst: dstH + "\n\n" + filler + "\n"},
			map[string]string{src: srcH + "\n\n" + filler + "\n", dst: dstH + "\n\n" + text + "\n\n"})
	}
	stays := "Unrelated content that stays put."
	ok := func(label string, r guardResult, table string, cond bool) {
		want(label, r.rc == 0 && table != "" && cond, r, table)
	}
	in := strings.Contains

	dir, sha := moved("source.md", "## Section One", "dest.md", "## Section Two", "This is the paragraph that will move verbatim to another file\nacross this change.", stays)
	r, table := gen(dir, sha, plan("yes", "source.md", "dest.md"))
	ok("case a: verbatim move classified moved", r, table, in(table, "moved") && in(table, "This is the paragraph"))
	want("pin: case a's table byte for byte", table == "| Passage | Source | Destination | Class |\n| --- | --- | --- | --- |\n"+
		"| This is the paragraph that will move verbatim… | `source.md` § Section One | `dest.md` § Section Two | moved |\n"+
		"| Unrelated content that stays put. | `dest.md` § Section Two | `source.md` § Section One | moved |\n", r, table)
	want("pin: the generator prints nothing", r.out == "", r, table)

	dir, sha = moved("alpha.md", "## Alpha", "beta.md", "## Beta", "See the rule in `old.md`, **Old Section**, for the full rule.", "Unrelated content.")
	writeFile(t, dir+"/beta.md", "## Beta\n\nSee the rule in `new.md`, **New Section**, for the full rule.\n\n")
	r, table = gen(dir, sha, plan("yes", "alpha.md", "beta.md"))
	ok("case b: repointed citation classified repointed", r, table, in(table, "repointed"))
	want("pin: case b's table byte for byte", table == "| Passage | Source | Destination | Class |\n| --- | --- | --- | --- |\n"+
		"| Unrelated content. | `beta.md` § Beta | `alpha.md` § Alpha | moved |\n"+
		"| See the rule in `old.md`, **Old Section**, for… | `alpha.md` § Alpha | `beta.md` § Beta | repointed |\n", r, table)

	dir, sha = repo(map[string]string{"gamma.md": "## Gamma\n\nOriginal content only.\n"},
		map[string]string{"gamma.md": "## Gamma\n\nOriginal content only.\n\nA brand new paragraph that never existed before this change landed.\n\n"})
	r, table = gen(dir, sha, plan("yes", "gamma.md"))
	ok("case c: new paragraph classified added", r, table, in(table, "added") && in(table, "A brand new paragraph"))

	dir, sha = repo(map[string]string{"delta.md": "## Delta\n\nContent that stays.\n\nThis paragraph is simply deleted, not moved anywhere else.\n\n"},
		map[string]string{"delta.md": "## Delta\n\nContent that stays.\n"})
	r, table = gen(dir, sha, plan("yes", "delta.md"))
	ok("case d: deleted paragraph classified removed", r, table, in(table, "removed") && in(table, "This paragraph is simply deleted"))
	want("pin: case d's table byte for byte", table == "| Passage | Source | Destination | Class |\n| --- | --- | --- | --- |\n"+
		"| This paragraph is simply deleted, not moved anywhere… | `delta.md` § Delta | — | removed |\n", r, table)

	dir, sha = repo(map[string]string{"epsilon.md": "## Epsilon\n\nSome content.\n"}, map[string]string{"epsilon.md": "## Epsilon\n\nSome different content.\n"})
	r, table = gen(dir, sha, plan("no", "epsilon.md"))
	want("case e: **Relocation:** no is a no-op, no output written", r.rc == 0 && table == "", r, table)

	dir, _ = repo(map[string]string{"zeta.md": "## Zeta\n\nSome content.\n"}, nil)
	r, table = gen(dir, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", plan("yes", "zeta.md"))
	want("case f: unresolvable merge-base exits 2 via the merge-base-resolves guard", r.rc == 2 && in(r.out, "does not resolve to a commit"), r, table)
	want("pin: case f's refusal byte for byte", r.stdout == "" && r.err == "generate-relocation-comparison: merge-base 'deadbeefdeadbeefdeadbeefdeadbeefdeadbeef' does not resolve to a commit in "+dir+"\n", r, table)

	dir, sha = repo(map[string]string{"eta.md": "## Eta\n\nSome content.\n"}, map[string]string{"eta.md": "\x00"})
	r, table = gen(dir, sha, plan("yes", "eta.md"))
	want("case g: scoped file missing on disk but still tracked exits 2", r.rc == 2, r, table)
	want("pin: case g's refusal byte for byte", r.err == "generate-relocation-comparison: scoped file missing from worktree: "+dir+"/eta.md: [Errno 2] No such file or directory: '"+dir+"/eta.md'\n", r, table)

	dir, sha = repo(map[string]string{"old.md": "## Old\n\nThis paragraph leaves a file the branch deletes outright.\n\n", "new.md": "## New\n\nFiller.\n"}, nil)
	gitRun(t, dir, "rm", "-q", "old.md")
	writeFile(t, dir+"/new.md", "## New\n\nFiller.\n\nThis paragraph leaves a file the branch deletes outright.\n\n")
	r, table = gen(dir, sha, plan("yes", "old.md", "new.md"))
	ok("case g2: a file the branch deleted reads as empty text", r, table, in(table, "| `old.md` § Old | `new.md` § New | moved |"))

	dir, sha = repo(map[string]string{"theta.md": "## Theta\n\nSome content.\n"}, nil)
	r, table = gen(dir, sha, "")
	want("case h: missing tasks.md exits 2", r.rc == 2, r, table)
	want("pin: case h's refusal byte for byte", r.err == "generate-relocation-comparison: cannot read "+dir+"/tasks.md: [Errno 2] No such file or directory: '"+dir+"/tasks.md'\n", r, table)

	dir, sha = repo(map[string]string{"iota.md": "## Iota\n\nSome content.\n"}, map[string]string{"iota.md": "## Iota\n\nSome different content.\n"})
	r, table = gen(dir, sha, "# fixture plan\n\n- [ ] 1. Only task\n\n**Files:**\n- Modify: `iota.md`\n\n")
	want("case i: absent **Relocation:** header is a no-op, no output written", r.rc == 0 && table == "", r, table)

	dir, sha = repo(map[string]string{"kappa.md": "## Kappa\n\nThis passage never changes at all across the whole revision.\n\nThis passage does change between the two revisions.\n\n"},
		map[string]string{"kappa.md": "## Kappa\n\nThis passage never changes at all across the whole revision.\n\nThis passage has now changed between the two revisions.\n\n"})
	r, table = gen(dir, sha, plan("yes", "kappa.md"))
	ok("case j: byte-identical same-location passage produces no row", r, table, !in(table, "This passage never changes"))

	dir, sha = repo(map[string]string{"mu.md": "## Kappa2\n\nSome content.\n"}, map[string]string{"mu.md": "## Kappa2\n\nSome content \xff\xfe binary garbage.\n"})
	r, table = gen(dir, sha, plan("yes", "mu.md"))
	want("case k: non-UTF-8 worktree file exits 2 (F1 regression)", r.rc == 2, r, table)
	want("pin: case k's refusal byte for byte", r.err == "generate-relocation-comparison: "+dir+"/mu.md is not valid UTF-8: 'utf-8' codec can't decode byte 0xff in position 24: invalid start byte\n", r, table)

	dir, sha = repo(map[string]string{"nu.md": "## Nu\n\nSome content.\n"}, nil)
	blob := tcfGit(t, dir, "rev-parse", sha+":nu.md")
	obj := dir + "/.git/objects/" + blob[:2] + "/" + blob[2:]
	if err := os.Chmod(obj, 0o644); err != nil {
		t.Fatal(err)
	}
	writeFile(t, obj, "corrupt")
	r, table = gen(dir, sha, plan("yes", "nu.md"))
	want("case l: corrupt merge-base blob exits 2, not misread as missing path (F2 regression)", r.rc == 2 && !in(r.out, "does not exist in"), r, table)
	want("pin: case l names the failing cat-file", strings.HasPrefix(r.err, "generate-relocation-comparison: git cat-file -t "+sha+":nu.md failed: "), r, table)

	edit := func(file, heading, before, after string) (string, string) {
		return repo(map[string]string{file: heading + "\n\n" + before + "\n\n"}, map[string]string{file: heading + "\n\n" + after + "\n\n"})
	}
	split := func(table string) bool { return !in(table, "repointed") && in(table, "removed") && in(table, "added") }

	dir, sha = edit("xi.md", "## Xi", "See `note.md` for details. Also **Alpha** far from the citation and not adjacent to it at all.",
		"See `note.md` for details. Also **Beta** far from the citation and not adjacent to it at all.")
	r, table = gen(dir, sha, plan("yes", "xi.md"))
	ok("case m: far bold-span edit not misclassified repointed (F3 regression guard)", r, table, split(table))

	dir, sha = moved("pi_src.md", "## Omicron", "pi_dst.md", "## Rho", "Value | here is piped content that moves across files verbatim.", stays)
	r, table = gen(dir, sha, plan("yes", "pi_src.md", "pi_dst.md"))
	ok("case n: literal pipe escaped in moved passage cell (F5 regression guard)", r, table, in(table, `Value \| here`))

	dir, sha = moved("alpha_o.md", "## AlphaO", "beta_o.md", "## BetaO", "See `old.md` **Old Section** for the full rule.", "Unrelated content.")
	writeFile(t, dir+"/beta_o.md", "## BetaO\n\nSee `new.md` **New Section** for the full rule.\n\n")
	r, table = gen(dir, sha, plan("yes", "alpha_o.md", "beta_o.md"))
	ok("case o: no-separator citation repoint classified repointed (F7 regression guard)", r, table, in(table, "repointed"))

	dir, sha = moved("bullet_src.md", "## SectionBulletA", "bullet_dst.md", "## SectionBulletB", "- This bullet line will move verbatim to another file.", stays)
	r, table = gen(dir, sha, plan("yes", "bullet_src.md", "bullet_dst.md"))
	ok("case p: moved bullet-list passage classified moved", r, table, in(table, "moved") && in(table, "This bullet line will move"))

	dir, sha = moved("table_src.md", "## SectionTableA", "table_dst.md", "## SectionTableB", "| MovedTableCell | second column value here |", stays)
	r, table = gen(dir, sha, plan("yes", "table_src.md", "table_dst.md"))
	ok("case q: moved table-row passage classified moved", r, table, in(table, "moved") && in(table, "MovedTableCell"))

	dir, sha = moved("heading_src.md", "## DistinctiveHeadingUpsilon", "heading_dst.md", "## HeadingDest", "This paragraph moves under a heading so its location gets a section suffix.", stays)
	r, table = gen(dir, sha, plan("yes", "heading_src.md", "heading_dst.md"))
	ok("case r: _format_location's § heading suffix appears in output", r, table, in(table, "§ HeadingDest"))

	dir, sha = moved("dup_src.md", "## SectionDup", "dup_dst.md", "## SectionDupDest", "This paragraph exists once and must not be reported twice.", stays)
	r, table = gen(dir, sha, "# fixture plan\n\n> **Relocation:** yes\n\n- [ ] 1. First task\n\n**Files:**\n- Modify: `dup_src.md`\n- Modify: `dup_dst.md`\n\n"+
		"- [ ] 2. Second task, naming dup_src.md again\n\n**Files:**\n- Modify: `dup_src.md`\n\n")
	ok("case s: scoped_files de-duplicates a file named by two tasks", r, table, strings.Count(table, "This paragraph exists once") == 1)

	dir, sha = moved("alpha_t.md", "## AlphaT", "beta_t.md", "## BetaT", "See the rule in `old.md`, **Old Section**, for the full rule.", "Unrelated content.")
	writeFile(t, dir+"/beta_t.md", "## BetaT\n\nSee the rule in `new.md` **New Section**, for the full rule.\n\n")
	r, table = gen(dir, sha, plan("yes", "alpha_t.md", "beta_t.md"))
	ok("case t: citation repoint with a changed separator style still classified repointed", r, table, in(table, "repointed"))

	dir, sha = repo(map[string]string{"phi.md": "## Phi\n\nSome unrelated content at the merge base.\n"},
		map[string]string{"new_file.md": "## Chi\n\nA brand new file this plan creates fresh, never at the merge base.\n"})
	r, table = gen(dir, sha, plan("yes", "phi.md", "new_file.md"))
	ok("case u: file on disk but absent at merge-base classified added, not exit 2 (F14 regression guard)", r, table,
		in(table, "added") && in(table, "A brand new file this plan creates fresh"))

	dir, sha = edit("sigma.md", "## Sigma", "The color `red` describes it, similar to `**Squash-with:**` field, unrelated to any move.",
		"The color `red` totally unrelated content edit here to `**Squash-with:**` field, unrelated to any move.")
	r, table = gen(dir, sha, plan("yes", "sigma.md"))
	ok("case v: cross-span splice content edit not misclassified repointed (F13 regression guard)", r, table, split(table))

	repoint := func(label, a, b, before, after string) {
		dir, sha := moved(a+".md", "## "+strings.ToUpper(a[:1])+a[1:], b+".md", "## "+strings.ToUpper(b[:1])+b[1:], before, "Unrelated content.")
		writeFile(t, dir+"/"+b+".md", "## "+strings.ToUpper(b[:1])+b[1:]+"\n\n"+after+"\n\n")
		r, table := gen(dir, sha, plan("yes", a+".md", b+".md"))
		ok(label, r, table, in(table, "repointed"))
	}
	repoint("case w: repointed citation with dropped trailing above classified repointed (F16 regression guard)", "alpha_w", "beta_w",
		"`old-file.md` **Old Heading** is the place to look above.", "`new-file.md` **New Heading** is the place to look.")

	dir, sha = repo(map[string]string{"psi/inner.md": "nested file content, irrelevant to the scoped path itself\n"},
		map[string]string{"psi": "This is now a plain file where a directory used to be.\n"})
	r, table = gen(dir, sha, plan("yes", "psi"))
	want("case x: directory-to-file scoped path exits 2, not garbage rows (F18 regression guard)", r.rc == 2, r, table)
	want("pin: case x's refusal byte for byte", r.err == "generate-relocation-comparison: "+sha+":psi is a tree, not a file (blob)\n", r, table)

	dir, sha = moved("h3_src.md", "### SectionH3A", "h3_dst.md", "### SectionH3B", "This paragraph moves verbatim under a level-three heading only.", stays)
	r, table = gen(dir, sha, plan("yes", "h3_src.md", "h3_dst.md"))
	ok("case y: moved passage under a ### heading classified moved (F19.1)", r, table, in(table, "moved"))

	dir, sha = moved("star_src.md", "## SectionStarA", "star_dst.md", "## SectionStarB", "* This star bullet line will move verbatim to another file.", stays)
	r, table = gen(dir, sha, plan("yes", "star_src.md", "star_dst.md"))
	ok("case z: moved * bullet passage classified moved (F19.2)", r, table, in(table, "moved") && in(table, "This star bullet line will move"))

	dir, sha = edit("omega.md", "## Omega", "See `note.md` here and there **Alpha** unrelated text after the gap.", "See `note.md` here and there **Beta** unrelated text after the gap.")
	r, table = gen(dir, sha, plan("yes", "omega.md"))
	ok("case aa: 4+ char gap not treated as citation adjacency (F19.3 guard)", r, table, split(table))

	dir, sha = edit("substr.md", "## Substr", "This sentence ends with the word notabove.", "This sentence ends with the word notbelow.")
	r, table = gen(dir, sha, plan("yes", "substr.md"))
	ok("case bb: word merely containing above/below as a substring not truncated (F19.4 guard)", r, table, split(table))

	dir, sha = repo(map[string]string{"cc.md": "## Cc\n\nSome content.\n"}, map[string]string{"cc.md": "## Cc\n\nSome different content.\n"})
	r, table = gen(dir, sha, "# fixture plan\n\n> **Relocation:** yes\n\n- [ ] 1. Only task, no files scoped\n\n")
	want("case cc: Relocation yes with zero scoped files is a no-op, no output written (F19.5)", r.rc == 0 && table == "", r, table)

	repoint("case dd: trailing punctuation run (ellipsis / !?) not left behind after above-strip (F20 regression guard)", "alpha_dd", "beta_dd",
		"`old-file.md` **Old Heading** is the place to look above...", "`new-file.md` **New Heading** is the place to look!?")
	repoint("case ee: punctuation-only-on-after-side and no-above-word stress cases still classified repointed", "alpha_ee", "beta_ee",
		"See the rule in `old.md`, **Old Section**, for the full rule", "See the rule in `new.md`, **New Section**, for the full rule.")

	dir = t.TempDir()
	gitRun(t, dir, "init", "-q")
	writeFile(t, dir+"/real_target.md", "the real target file content, unrelated to the symlink path itself\n")
	psLink(t, "real_target.md", dir+"/link_ff.md")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-qm", "base")
	sha = tcfGit(t, dir, "rev-parse", "HEAD")
	r, table = gen(dir, sha, plan("yes", "link_ff.md"))
	want("case ff: symlink scoped path at merge-base ref exits 2, not a spurious row (F21 regression guard)", r.rc == 2, r, table)
	want("pin: case ff's refusal names the worktree symlink first", r.err == "generate-relocation-comparison: scoped file is a symlink in the worktree, cannot compare: "+dir+"/link_ff.md\n", r, table)

	dir, sha = repo(map[string]string{"gg.md": "original content at the scoped path, as an ordinary regular file\n"},
		map[string]string{"gg.md": "\x00", "gg_target.md": "different real content the symlink target actually holds\n"})
	psLink(t, "gg_target.md", dir+"/gg.md")
	r, table = gen(dir, sha, plan("yes", "gg.md"))
	want("case gg: symlink scoped path in the worktree exits 2, not a spurious row (F21 regression guard)", r.rc == 2, r, table)

	repoint("case hh: trailing word + whitespace + punctuation not left diverging after above-strip (F22 regression guard)", "alpha_hh", "beta_hh",
		"`old-file.md` **Old Heading** is the place to look above  .", "`new-file.md` **New Heading** is the place to look .")

	dir, sha = moved("src_ii.md", "## Source", "dst_ii.md", "## Dest", "A paragraph that moves verbatim between files.", "Unrelated content.")
	r, table = gen(dir, sha, "# fixture plan\n\n> **Relocation:** yes\n\n- [ ] 1. Only task\n\n**Files:** `src_ii.md`,\n`dst_ii.md`\n**Tests:** none\n")
	ok("case ii: an inline, wrapped **Files:** field scopes its paths (KAN-860 F5)", r, table, in(table, "moved"))

	// Beyond the harness: the outputs it never pinned, as the Python and its
	// wrapper printed them at bbfe685e.
	dir, sha = repo(map[string]string{"sym.md": "## S\n\nkeep\n"}, nil)
	if err := os.Remove(dir + "/sym.md"); err != nil {
		t.Fatal(err)
	}
	psLink(t, "elsewhere.md", dir+"/sym.md")
	writeFile(t, dir+"/elsewhere.md", "## S\n\nkeep\n")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-qm", "symlink")
	sha = tcfGit(t, dir, "rev-parse", "HEAD")
	if err := os.Remove(dir + "/sym.md"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir+"/sym.md", "## S\n\nkeep\n")
	r, table = gen(dir, sha, plan("yes", "sym.md"))
	want("pin: a symlink at the merge-base only exits 2 naming it", r.rc == 2 &&
		r.err == "generate-relocation-comparison: "+sha+":sym.md is a symlink, not a regular file\n", r, table)
	dir, sha = repo(map[string]string{"bin.md": "## B\n\n\xff\n"}, map[string]string{"bin.md": "## B\n\nok\n"})
	r, table = gen(dir, sha, plan("yes", "bin.md"))
	want("pin: a merge-base blob that is not UTF-8 exits 2", r.rc == 2 &&
		r.err == "generate-relocation-comparison: "+sha+":bin.md is not valid UTF-8: 'utf-8' codec can't decode byte 0xff in position 6: invalid start byte\n", r, table)
	// Header wins: the Python let a non-UTF-8 tasks.md escape as a traceback
	// (exit 1); its header names "tasks.md cannot be read" exit 2.
	dir, sha = repo(map[string]string{"a.md": "x\n"}, nil)
	r, table = gen(dir, sha, "> **Relocation:** yes \xff\n")
	want("pin: a tasks.md that is not UTF-8 exits 2", r.rc == 2 &&
		r.err == "generate-relocation-comparison: cannot read "+dir+"/tasks.md: 'utf-8' codec can't decode byte 0xff in position 22: invalid start byte\n", r, table)
	// The header's own grammar: leading space and `>` spacing are free, the
	// first matching line decides, and `yes` must end at a word boundary.
	dir, sha = moved("m1.md", "## M1", "m2.md", "## M2", "Moving words.", "Other words.")
	r, table = gen(dir, sha, "  >   **Relocation:**  yes, it is\n> **Relocation:** no\n\n- [ ] 1. T\n\n**Files:**\n- `m1.md`\n- `m2.md`\n")
	ok("pin: the first header line decides, its spacing free", r, table, in(table, "Moving words."))
	dir, sha = moved("m1.md", "## M1", "m2.md", "## M2", "Moving words.", "Other words.")
	r, table = gen(dir, sha, "> **Relocation:** yesé\n\n- [ ] 1. T\n\n**Files:**\n- `m1.md`\n- `m2.md`\n")
	want("pin: a header value running into a Unicode letter is no yes", r.rc == 0 && table == "", r, table)
	// Eight words exactly carry no ellipsis; nine do; a heading-less passage
	// has a bare location; an empty heading is falsy too.
	dir, sha = repo(map[string]string{"w.md": "one two three four five six seven eight\n\n## \n\nnine words here one two three four five six\n"},
		map[string]string{"w.md": "\n"})
	r, table = gen(dir, sha, plan("yes", "w.md"))
	ok("pin: row cells for eight and nine words and bare locations", r, table, table == "| Passage | Source | Destination | Class |\n| --- | --- | --- | --- |\n"+
		"| one two three four five six seven eight | `w.md` | — | removed |\n"+
		"| nine words here one two three four five… | `w.md` | — | removed |\n")
	// The shim: three arguments, the plan read as <changeRoot>/tasks.md.
	dir, sha = moved("source.md", "## One", "dest.md", "## Two", "Shim-run moved text.", stays)
	writeFile(t, dir+"/plan/tasks.md", plan("yes", "source.md", "dest.md"))
	r = psExec(t, []string{"FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, grcShim, dir, dir+"/plan", sha)
	b, _ := os.ReadFile(dir + "/.superpowers/sdd/relocation-comparison.md")
	want("pin: the shim appends <changeRoot>/tasks.md", r.rc == 0 && in(string(b), "Shim-run moved text."), r, string(b))
	r = psExec(t, []string{"FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, grcShim, dir, dir)
	want("pin: the usage line", r.rc == 2 && r.err == "usage: generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>\n", r, "")

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}

var grcShim = psScripts + "/generate-relocation-comparison.sh"
