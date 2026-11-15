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

var cacheKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// pruneCache keeps scripts/flow-guard's cache from growing without bound: each
// source edit adds a key directory <cache>/<sha256>/flow-guard, and nothing
// else ever removes one. When exe is such a binary it marks its own key used
// and removes every sibling key directory unused for cacheKeyMaxAge. Any other
// exe (go run, a test binary) is left alone, and every failure is ignored --
// pruning is housekeeping, never a reason for a guard to fail.
func pruneCache(exe string, now time.Time) {
	own := filepath.Dir(exe)
	cache := filepath.Dir(own)
	if filepath.Base(exe) != "flow-guard" || !cacheKey.MatchString(filepath.Base(own)) {
		return
	}
	_ = os.Chtimes(own, now, now)
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
