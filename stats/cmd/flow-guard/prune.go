package main

import (
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// cacheKeyMaxAge is how long a key directory of scripts/flow-guard's cache
// may go unused before pruneCache removes it. Every run touches its own key,
// so a binary any checkout still calls is never old.
const cacheKeyMaxAge = 7 * 24 * time.Hour

// pruneInterval is how often pruneCache scans the cache: the scan stats every
// key directory, so it runs at most once per interval, timed by the mtime of
// <cache>/pruned, rather than on every guard call.
const pruneInterval = 24 * time.Hour

var cacheKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// pruneCache keeps scripts/flow-guard's cache from growing without bound: each
// source edit adds a key directory <cache>/<sha256>/flow-guard, and nothing
// else ever removes one. When exe is such a binary it marks its own key used
// and removes every sibling key directory unused for cacheKeyMaxAge, at most once
// per pruneInterval. Any other
// exe (go run, a test binary) is left alone, and every failure is ignored --
// pruning is housekeeping, never a reason for a guard to fail.
func pruneCache(exe string, now time.Time) {
	own := filepath.Dir(exe)
	cache := filepath.Dir(own)
	if filepath.Base(exe) != "flow-guard" || !cacheKey.MatchString(filepath.Base(own)) {
		return
	}
	_ = os.Chtimes(own, now, now)
	stamp := filepath.Join(cache, "pruned")
	if info, err := os.Stat(stamp); err == nil && now.Sub(info.ModTime()) < pruneInterval {
		return
	}
	if f, err := os.Create(stamp); err == nil {
		_ = f.Close()
	}
	_ = os.Chtimes(stamp, now, now)
	entries, err := os.ReadDir(cache)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || !cacheKey.MatchString(e.Name()) || e.Name() == filepath.Base(own) {
			continue
		}
		if info, err := e.Info(); err == nil && now.Sub(info.ModTime()) > cacheKeyMaxAge {
			_ = os.RemoveAll(filepath.Join(cache, e.Name()))
		}
	}
}
