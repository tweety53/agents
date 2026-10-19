package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the check-identical-baselines contract
// (scripts/check-identical-baselines.sh).
func TestCheckIdenticalBaselines(t *testing.T) {
	cases := []struct {
		name  string
		files []string // "rel=content"; "" is a blank stdin line
		rc    int
		want  string // %D% is the temp root
	}{
		{"distinct files pass", []string{"s-snapshots/up.png=a", "s-snapshots/down.png=b"}, 0, "IDENTICAL-NONE: 2 PNGs\n"},
		// kan-998: pad-up and pad-down baselines were one image.
		{"identical pair fails", []string{"s-snapshots/pin-p3-pad-down.png=x", "s-snapshots/pin-p2-pad-up.png=x"}, 1,
			"identical: %D%/s-snapshots: pin-p2-pad-up.png pin-p3-pad-down.png\n"},
		{"group of three on one line", []string{"s/c.png=x", "s/a.png=x", "s/b.png=x", "s/d.png=y"}, 1,
			"identical: %D%/s: a.png b.png c.png\n"},
		{"different directories never group", []string{"results/up.png=x", "s-snapshots/up.png=x"}, 0, "IDENTICAL-NONE: 2 PNGs\n"},
		{"blank stdin lines are skipped", []string{"s/a.png=a", "", "s/b.png=b"}, 0, "IDENTICAL-NONE: 2 PNGs\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			var stdin strings.Builder
			for _, f := range c.files {
				if f == "" {
					stdin.WriteString("\n")
					continue
				}
				rel, body, _ := strings.Cut(f, "=")
				p := filepath.Join(root, rel)
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
				stdin.WriteString(p + "\n")
			}
			r := runGuard("check-identical-baselines", nil, Env{Stdin: strings.NewReader(stdin.String())})
			want := strings.ReplaceAll(c.want, "%D%", root)
			if r.rc != c.rc || r.stdout != want {
				t.Fatalf("rc=%d stdout=%q, want rc=%d stdout=%q", r.rc, r.stdout, c.rc, want)
			}
		})
	}

	for name, tc := range map[string]struct {
		args  []string
		stdin string
	}{
		"an argument given":          {[]string{"x"}, ""},
		"empty stdin":                {nil, ""},
		"a path that does not exist": {nil, filepath.Join(t.TempDir(), "absent.png") + "\n"},
	} {
		t.Run("exit 2: "+name, func(t *testing.T) {
			r := runGuard("check-identical-baselines", tc.args, Env{Stdin: strings.NewReader(tc.stdin)})
			if r.rc != 2 || r.stdout != "" {
				t.Fatalf("rc=%d stdout=%q, want 2 and empty stdout", r.rc, r.stdout)
			}
		})
	}
}
