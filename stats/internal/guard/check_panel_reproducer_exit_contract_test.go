package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-panel-reproducer-exit-contract.sh at
// 0747740, one subtest per ok: label. The bash harness's stub `flow` answered
// both `state get` and `record findings`; here the findings read is
// Env.Findings (case 12 alone execs the stub, so the production path is run
// too) and `state get` stays a stub `flow` on Env's PATH. The bash stub
// runner is gone: the real in-process runner executes a fixture reproducer
// whose own exit decides the runner's verdict, and every reproducer logs
// its run to <sandbox>/runs, so "the runner never ran" is the log's absence.

const pcState = `{"state":"IN_PROGRESS","worktrees":{}}`

// pcSandbox is a worktree carrying the harness's default auditable
// reproducer, repro.sh (declaration target.txt:2), plus a stub `flow` whose
// `state get` prints bin/state.json.
type pcSandbox struct {
	wt, bin string
	env     Env
}

func pcFindings(triples ...string) []byte {
	type finding struct {
		Ref        string `json:"ref"`
		Status     string `json:"status"`
		Reproducer string `json:"reproducer"`
	}
	out := []finding{}
	for i := 0; i+2 < len(triples); i += 3 {
		out = append(out, finding{triples[i], triples[i+1], triples[i+2]})
	}
	b, _ := json.Marshal(out)
	return b
}

func newPCSandbox(t *testing.T, findings []byte) *pcSandbox {
	t.Helper()
	s := &pcSandbox{wt: rrWorktree(t), bin: t.TempDir()}
	writeFile(t, filepath.Join(s.bin, "state.json"), pcState+"\n")
	writeExec(t, filepath.Join(s.bin, "flow"), `#!/usr/bin/env bash
case "${1:-}" in
  state) cat "$(dirname -- "$0")/state.json"; exit 0 ;;
  *) echo "flow: connect: connection refused" >&2; exit 1 ;;
esac
`)
	writeFile(t, filepath.Join(s.wt, "target.txt"), "line one\ndefect present here\nline three\n")
	s.repro(t, "repro.sh", "# demonstrates: target.txt:2:defect present here\nexit 9")
	m := map[string]string{"PATH": s.bin + ":" + os.Getenv("PATH"), "TMPDIR": t.TempDir()}
	s.env = Env{Getenv: func(k string) string { return m[k] }, Dir: t.TempDir(),
		ReproducerBound: rrTestBound, ReproducerGrace: rrTestGrace,
		Findings: func(string) ([]byte, error) { return findings, nil }}
	return s
}

// repro writes an executable reproducer in the sandbox's worktree: a
// shebang, a line logging the run (its physical cwd to <wt>/runs, its
// process group to <wt>/groups for rrWorktree's cleanup), then body. Every
// declaration in body therefore sits one line lower than in the bash
// harness's fixtures, which moves no case across the 10-line window.
func (s *pcSandbox) repro(t *testing.T, rel, body string) {
	t.Helper()
	s.reproIn(t, s.wt, rel, body)
}

// A reproducer in the sandbox's own worktree names <wt> from its own path
// (fixtureRoot), so cases with the same rel and body share one assessed
// inode (writeExec); one in another tree still logs to this sandbox's <wt>.
func (s *pcSandbox) reproIn(t *testing.T, tree, rel, body string) {
	t.Helper()
	root := s.wt + "/"
	if tree == s.wt {
		root = fixtureRoot(rel)
	}
	writeExec(t, filepath.Join(tree, rel), fmt.Sprintf("#!/usr/bin/env bash\npwd -P >> \"%sruns\"; echo $$ >> \"%sgroups\"\n%s\n",
		root, root, body))
}

func (s *pcSandbox) run(t *testing.T, name string) (int, string) {
	t.Helper()
	fn := Registry["check-panel-reproducer-exit-contract"]
	if fn == nil {
		t.Fatal("check-panel-reproducer-exit-contract is not registered")
	}
	var out bytes.Buffer
	return fn([]string{s.wt, name}, s.env, &out, &out), out.String()
}

// runs is the reproducer's run log: one physical cwd per run.
func (s *pcSandbox) runs() []string {
	b, _ := os.ReadFile(filepath.Join(s.wt, "runs"))
	return strings.Fields(string(b))
}

func (s *pcSandbox) neverRan(t *testing.T, label string) {
	t.Run(label, func(t *testing.T) {
		if r := s.runs(); len(r) != 0 {
			t.Fatalf("the runner was invoked despite the audit failure: %v", r)
		}
	})
}

func TestCheckPanelReproducerExitContract(t *testing.T) {
	t.Parallel()
	// The reproducer bodies that give each runner verdict: 0 demonstrated,
	// 1 not demonstrated, 2 refused (no execute permission, found only once
	// the audit has passed), 3 unverifiable (the bound fired through the
	// bound file), 4 cannot answer (exit 127, the runner's exec heuristic).
	withVerdict := func(t *testing.T, s *pcSandbox, code int) {
		t.Helper()
		decl := "# demonstrates: target.txt:2:defect present here\n"
		switch code {
		case 0:
			s.repro(t, "repro.sh", decl+"exit 9")
		case 1:
			s.repro(t, "repro.sh", decl+"exit 0")
		case 2:
			// Rewritten as a file of its own rather than chmodded: repro.sh
			// is a hard link to a shared master (writeExec).
			p := filepath.Join(s.wt, "repro.sh")
			b, err := os.ReadFile(p)
			if err == nil {
				err = os.Remove(p)
			}
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, p, string(b))
		case 3:
			bound := filepath.Join(s.wt, "bound")
			m := s.env.Getenv
			s.env.Getenv = func(k string) string {
				if k == "RUN_REPRODUCER_BOUND_FILE" {
					return bound
				}
				return m(k)
			}
			s.repro(t, "repro.sh", decl+"touch "+bound+"; sleep 30\nexit 0")
		case 4:
			s.repro(t, "repro.sh", decl+"exit 127")
		}
	}
	expect := func(t *testing.T, got int, out string, want int, needle string) {
		t.Helper()
		rrExpect(t, got, out, want, needle)
	}

	t.Run("case 1: a demonstrated open finding exits 0", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "REPRODUCER-EXIT-CONTRACT-OK (1 runnable")
		// Bare means no --pre-fix-verdict: with one, this demonstrated
		// reproducer would have been refused as ambiguous.
		t.Run("case 1b: the runner ran bare (two arguments, no --pre-fix-verdict)", func(t *testing.T) {
			if n := len(s.runs()); n != 1 || strings.Contains(out, "ambiguous") {
				t.Fatalf("runs %d, want 1 bare run:\n%s", n, out)
			}
		})
	})

	for _, c := range []struct {
		label  string
		code   int
		want   int
		needle string
	}{
		{"case 2: a not-demonstrated verdict on an open finding exits 1", 1, 1, "F1"},
		{"case 3: a refused reproducer exits 1", 2, 1, "F1"},
		{"case 4: an unverifiable reproducer is cannot-answer", 3, 2, "F1"},
		{"case 5: a cannot-answer runner verdict is the guard exit 2", 4, 2, "F1"},
	} {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
			withVerdict(t, s, c.code)
			got, out := s.run(t, "demo")
			expect(t, got, out, c.want, c.needle)
		})
	}

	for _, st := range []string{"fixed", "deferred covered by the suite run", "withdrawn operator said so"} {
		t.Run("case 6: a "+st+" finding is skipped, runner never invoked", func(t *testing.T) {
			t.Parallel()
			s := newPCSandbox(t, pcFindings("F1", st, "repro.sh"))
			withVerdict(t, s, 2)
			got, out := s.run(t, "demo")
			expect(t, got, out, 0, "(0 runnable")
		})
	}

	for _, c := range []struct{ label, repro string }{
		{"case 7: the none — reason exemption is skipped", "none — prose-only, no runnable check"},
		{"case 8: a bare none is the lexical guard's subject, skipped here", "none"},
	} {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			s := newPCSandbox(t, pcFindings("F1", "open", c.repro))
			got, out := s.run(t, "demo")
			expect(t, got, out, 0, "(0 runnable")
		})
	}

	t.Run("case 9: zero findings exits 0", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, []byte("[]"))
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "")
	})

	t.Run("case 10: non-directory argument exits 2 and names it", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, []byte("[]"))
		s.wt = filepath.Join(s.wt, "target.txt")
		got, out := s.run(t, "demo")
		expect(t, got, out, 2, "not a directory")
	})

	for _, bad := range []string{"../../../planted/clear", "demo*", "demo/../demo", ".hidden", "demo?x"} {
		t.Run("case 11: change name '"+bad+"' is rejected", func(t *testing.T) {
			t.Parallel()
			s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
			got, out := s.run(t, bad)
			expect(t, got, out, 2, "is not a plain change name")
		})
	}
	t.Run("case 11: a missing change name is rejected", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		got, out := s.run(t, "")
		expect(t, got, out, 2, "usage:")
	})

	t.Run("case 12: an unreachable store exits 2", func(t *testing.T) {
		t.Parallel()
		// Env.Findings nil: the guard execs the stub `flow record findings`,
		// which fails the way an unreachable store does.
		s := newPCSandbox(t, nil)
		s.env.Findings = nil
		got, out := s.run(t, "demo")
		expect(t, got, out, 2, "cannot determine anything")
		expect(t, got, out, 2, "connection refused")
	})
	t.Run("case 12 (Env.Findings): an unreachable store exits 2", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, nil)
		s.env.Findings = func(string) ([]byte, error) { return []byte("dial refused"), errors.New("exit status 1") }
		got, out := s.run(t, "demo")
		expect(t, got, out, 2, "cannot determine anything: dial refused")
	})

	t.Run("case 13: a null reproducer on an open finding is cannot-answer", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", ""))
		got, out := s.run(t, "demo")
		expect(t, got, out, 2, "no reproducer field")
	})

	t.Run("case 14: mixed verdicts exit 1 naming only the contradicted ref", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro-good.sh", "F2", "open", "repro-inverted.sh", "F3", "fixed", "repro-old.sh"))
		decl := "# demonstrates: target.txt:2:defect present here\n"
		s.repro(t, "repro-good.sh", decl+"exit 9")
		s.repro(t, "repro-inverted.sh", decl+"exit 0")
		s.repro(t, "repro-old.sh", decl+"exit 9")
		got, out := s.run(t, "demo")
		expect(t, got, out, 1, "F2")
		if strings.Contains(out, "check-panel-reproducer-exit-contract: F1's") || strings.Contains(out, "F3 — running") {
			t.Fatalf("a ref other than F2 was named a violation or run:\n%s", out)
		}
	})

	t.Run("case 15: cannot-answer outranks violations", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro-a.sh", "F2", "open", "repro-b.sh"))
		decl := "# demonstrates: target.txt:2:defect present here\n"
		bound := filepath.Join(s.wt, "bound")
		m := s.env.Getenv
		s.env.Getenv = func(k string) string {
			if k == "RUN_REPRODUCER_BOUND_FILE" {
				return bound
			}
			return m(k)
		}
		s.repro(t, "repro-a.sh", decl+"exit 0")
		s.repro(t, "repro-b.sh", decl+"touch "+bound+"; sleep 30\nexit 0")
		got, out := s.run(t, "demo")
		expect(t, got, out, 2, "F2")
	})

	t.Run("case 16: real runner, non-zero reproducer, exit 0", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "demo-repro.sh"))
		s.repro(t, "demo-repro.sh", "# demonstrates: target.txt:2:defect present here\nexit 7")
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "F1 — defect demonstrated, claim holds")
	})
	t.Run("case 17: real runner, zero-exit reproducer, exit 1", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "demo-repro.sh"))
		s.repro(t, "demo-repro.sh", "# demonstrates: target.txt:2:defect present here\nexit 0")
		got, out := s.run(t, "demo")
		expect(t, got, out, 1, "F1")
	})

	// The instrument audit (KAN-606): every audit failure is exit 1 naming
	// the ref, and the runner never runs.
	for _, c := range []struct {
		n, label, body string
		setup          func(t *testing.T, s *pcSandbox)
	}{
		{"18", "no demonstrates declaration is a violation", "exit 9", nil},
		{"19", "a declaration past line 10 is a violation", strings.Repeat("# padding\n", 10) + "# demonstrates: target.txt:2:defect present here\nexit 9", nil},
		{"20", "an absolute declared path is a violation", "# demonstrates: /etc/passwd:1:root\nexit 9", nil},
		{"21", "a .. declared path is a violation", "# demonstrates: ../outside.txt:1:content\nexit 9", nil},
		{"22", "a declared file the tree does not carry is a violation", "# demonstrates: missing.txt:2:defect present here\nexit 9", nil},
		{"23", "a declared line past the end of file is a violation", "# demonstrates: target.txt:99:defect present here\nexit 9", nil},
		{"24", "content absent from the declared line is a violation", "# demonstrates: target.txt:2:some other assertion entirely\nexit 9", nil},
		{"25", "a malformed declaration is a violation", "# demonstrates: target.txt:defect present here\nexit 9", nil},
		{"29", "a symlink-escape citation is a violation", "# demonstrates: escape.txt:1:outside content\nexit 9", func(t *testing.T, s *pcSandbox) {
			outside := t.TempDir()
			writeFile(t, filepath.Join(outside, "external.txt"), "outside content\n")
			if err := os.Symlink(filepath.Join(outside, "external.txt"), filepath.Join(s.wt, "escape.txt")); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run("case "+c.n+": "+c.label, func(t *testing.T) {
			t.Parallel()
			s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
			s.repro(t, "repro.sh", c.body)
			if c.setup != nil {
				c.setup(t, s)
			}
			got, out := s.run(t, "demo")
			expect(t, got, out, 1, "F1")
			b := map[string]string{"18": "no declaration means the runner never runs", "29": "a symlink escape means the runner never runs"}[c.n]
			if b == "" {
				b = "the audit failure means the runner never runs"
			}
			s.neverRan(t, "case "+c.n+"b: "+b)
		})
	}

	t.Run("case 26: a mutation-declared reproducer skips the audit", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", "# mutation-reproducer\nexit 0")
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "")
		t.Run("case 26b: the exempt reproducer still ran", func(t *testing.T) {
			if len(s.runs()) != 1 {
				t.Fatalf("the exempt reproducer ran %d times, want 1", len(s.runs()))
			}
		})
	})

	t.Run("case 27: an unreadable reproducer script is a violation", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "absent.sh"))
		got, out := s.run(t, "demo")
		expect(t, got, out, 1, "F1")
		s.neverRan(t, "case 27b: an unreadable script means the runner never runs")
	})

	t.Run("case 28: a tab-separated reproducer reaches the runner", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh\t--strict"))
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "")
		t.Run("case 28: the tab-separated reproducer was invoked bare", func(t *testing.T) {
			if len(s.runs()) != 1 || strings.Contains(out, "ambiguous") {
				t.Fatalf("runs %v, want one bare run:\n%s", s.runs(), out)
			}
		})
	})

	t.Run("case 30: the not-demonstrated message names the mutation convention", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", "# mutation-reproducer\nexit 7")
		got, out := s.run(t, "demo")
		expect(t, got, out, 1, "mutation-reproducer")
	})
	t.Run("case 31: the not-demonstrated message names the generic convention", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		withVerdict(t, s, 1)
		got, out := s.run(t, "demo")
		expect(t, got, out, 1, "a generic one demonstrates with a non-zero exit")
	})

	// The state record (KAN-658).
	for _, c := range []struct{ label, stateArm, needle string }{
		{"case 32: a state record the store does not carry is cannot-answer", "exit 1", "no record of change"},
		{"case 33: an unreachable state read is cannot-answer", "exit 0", "cannot determine anything"},
	} {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
			writeExec(t, filepath.Join(s.bin, "flow"), "#!/usr/bin/env bash\ncase \"${1:-}\" in\n  state) "+c.stateArm+" ;;\n  *) exit 1 ;;\nesac\n")
			got, out := s.run(t, "demo")
			expect(t, got, out, 2, c.needle)
		})
	}

	// withPeers records peer worktrees in the state record's worktrees map
	// and drops the canonical tree's own repro.sh unless keep is set.
	withPeers := func(t *testing.T, s *pcSandbox, keep bool, peers ...string) {
		t.Helper()
		m := map[string]string{}
		for i, p := range peers {
			m[p] = fmt.Sprintf("%040d", i)
		}
		b, _ := json.Marshal(map[string]any{"state": "IN_PROGRESS", "worktrees": m})
		writeFile(t, filepath.Join(s.bin, "state.json"), string(b)+"\n")
		if !keep {
			if err := os.Remove(filepath.Join(s.wt, "repro.sh")); err != nil {
				t.Fatal(err)
			}
		}
	}

	t.Run("case 34: a reproducer resolving only in a recorded peer worktree runs there", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		peer := t.TempDir()
		withPeers(t, s, false, peer)
		// The canonical tree's cited line differs on purpose: an audit
		// resolved against the wrong tree fails this case.
		writeFile(t, filepath.Join(s.wt, "target.txt"), "line one\ncanonical content\nline three\n")
		writeFile(t, filepath.Join(peer, "target.txt"), "line one\npeer content\nline three\n")
		s.reproIn(t, peer, "repro.sh", "# demonstrates: target.txt:2:peer content\nexit 9")
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "")
		t.Run("case 34b: the runner was invoked with the peer worktree, not the canonical one", func(t *testing.T) {
			want, _ := filepath.EvalSymlinks(peer)
			if r := s.runs(); len(r) != 1 || r[0] != want {
				t.Fatalf("runs %v, want [%s]", r, want)
			}
		})
	})

	t.Run("case 35: a path resolving in several recorded worktrees is cannot-answer", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		p1, p2 := t.TempDir(), t.TempDir()
		withPeers(t, s, false, p1, p2)
		s.reproIn(t, p1, "repro.sh", "# demonstrates: target.txt:2:peer content\nexit 9")
		s.reproIn(t, p2, "repro.sh", "# demonstrates: target.txt:2:peer content\nexit 9")
		got, out := s.run(t, "demo")
		expect(t, got, out, 2, "F1")
		s.neverRan(t, "case 35b: an ambiguous tree means the runner never runs")
	})

	t.Run("case 36: the canonical tree is preferred when it carries the path", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		p3 := t.TempDir()
		withPeers(t, s, true, p3)
		writeFile(t, filepath.Join(p3, "target.txt"), "line one\ndefect present here\nline three\n")
		s.reproIn(t, p3, "repro.sh", "# demonstrates: target.txt:2:defect present here\nexit 9")
		got, out := s.run(t, "demo")
		expect(t, got, out, 0, "")
		t.Run("case 36b: the runner ran against the canonical worktree", func(t *testing.T) {
			want, _ := filepath.EvalSymlinks(s.wt)
			if r := s.runs(); len(r) != 1 || r[0] != want {
				t.Fatalf("runs %v, want [%s]", r, want)
			}
		})
	})
}

// TestPanelExitContractReproducerGetsLCAllC: the bash guard exported
// LC_ALL=C, so every reproducer it ran saw it.
func TestPanelExitContractReproducerGetsLCAllC(t *testing.T) {
	t.Parallel()
	s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
	s.repro(t, "repro.sh", "# demonstrates: target.txt:2:defect present here\n[ \"${LC_ALL:-}\" = C ] && exit 9\nexit 0")
	got, out := s.run(t, "demo")
	rrExpect(t, got, out, 0, "REPRODUCER-EXIT-CONTRACT-OK (1 runnable")
}

// TestPanelExitContractNullFindingsCannotAnswer: a `null` findings body is
// jq failing in the bash guard — never a verdict.
func TestPanelExitContractNullFindingsCannotAnswer(t *testing.T) {
	t.Parallel()
	s := newPCSandbox(t, []byte("null"))
	got, out := s.run(t, "demo")
	if got != 2 || strings.Contains(out, "REPRODUCER-EXIT-CONTRACT-OK") {
		t.Fatalf("exit %d, want 2 and no OK line:\n%s", got, out)
	}
}
