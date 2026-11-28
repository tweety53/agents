package guardtest

import (
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CopyTreeMode copies src to dst; an executable is linked to its master inode
// (WriteExec), so a copied guard or harness costs no fresh assessment.
func CopyTreeMode(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := dst + strings.TrimPrefix(p, src)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		if fi.Mode()&0o111 != 0 {
			WriteExec(t, target, string(b))
			return nil
		}
		return os.WriteFile(target, b, fi.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

func CopyTreeT(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := dst + strings.TrimPrefix(p, src)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, fi.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

// LibRes is one call's observable result, bash or Go.
type LibRes struct {
	Stdout, Stderr string
	Status         int
}
