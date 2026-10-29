// Package embedded embeds skills/flow-contracts/model-policy.md into flowd,
// so the roles table it serves is the one `make restart` built after the
// last landing (KAN-1067's roles-table-embedded decision).
//
// The file is not committed here: stats/Makefile's policy-copy target --
// a prerequisite of web-dist, build and test -- copies it beside this file
// on every make line, and .gitignore keeps the copy out of git. A //go:embed
// pattern cannot climb above its directory with "../", hence the copy, as
// internal/web/embed.go explains for dist/. A plain `go build` of a package
// importing this one without that copy fails at compile time ("pattern
// model-policy.md: no matching files found"), never serving a stale table.
//
// It is its own package, not part of internal/modelpolicy, because
// internal/guard imports internal/modelpolicy and cmd/flow-guard is built by
// a plain `go build` (scripts/lib/flow-guard.sh) that never runs make.
package embedded

import _ "embed"

// Table is model-policy.md as of the last make.
//
//go:embed model-policy.md
var Table []byte
