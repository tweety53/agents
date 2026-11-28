package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPruneCache(t *testing.T) {
	cache := t.TempDir()
	now := time.Now()
	key := func(c byte) string { return strings.Repeat(string(c), 64) }
	mk := func(name string, age time.Duration) {
		d := filepath.Join(cache, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "flow-guard"), nil, 0o755); err != nil {
			t.Fatal(err)
		}
		old := now.Add(-age)
		if err := os.Chtimes(d, old, old); err != nil {
			t.Fatal(err)
		}
	}
	mk(key('a'), 30*24*time.Hour) // own key, stale: touched, kept
	mk(key('b'), 30*24*time.Hour) // stale sibling: removed
	mk(key('c'), time.Hour)       // fresh sibling: kept
	mk("not-a-key", 30*24*time.Hour)

	pruneCache(filepath.Join(cache, key('a'), "flow-guard"), now)

	for name, want := range map[string]bool{key('a'): true, key('b'): false, key('c'): true, "not-a-key": true} {
		_, err := os.Stat(filepath.Join(cache, name))
		if (err == nil) != want {
			t.Errorf("%.8s kept=%v, want %v", name, err == nil, want)
		}
	}
}

func TestPruneCacheIgnoresOtherBinaries(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, strings.Repeat("b", 64))
	if err := os.Mkdir(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-30 * 24 * time.Hour)
	_ = os.Chtimes(stale, old, old)
	pruneCache(filepath.Join(dir, "flow-guard.test"), time.Now())
	if _, err := os.Stat(stale); err != nil {
		t.Errorf("a binary outside the cache layout pruned its neighbours: %v", err)
	}
}

func TestPruneCacheRunsOncePerWindow(t *testing.T) {
	cache := t.TempDir()
	now := time.Now()
	own := filepath.Join(cache, strings.Repeat("a", 64))
	stale := filepath.Join(cache, strings.Repeat("b", 64))
	for _, d := range []string{own, stale} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pruneCache(filepath.Join(own, "flow-guard"), now)
	old := now.Add(-30 * 24 * time.Hour)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}
	pruneCache(filepath.Join(own, "flow-guard"), now.Add(time.Hour))
	if _, err := os.Stat(stale); err != nil {
		t.Fatalf("a second run within the window pruned: %v", err)
	}
	pruneCache(filepath.Join(own, "flow-guard"), now.Add(pruneInterval+time.Hour))
	if _, err := os.Stat(stale); err == nil {
		t.Fatal("a run past the window did not prune")
	}
}
