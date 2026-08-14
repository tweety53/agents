// Package web embeds the built SPA (stats/web, built by Vite) into the
// myflowd binary and serves it: static assets as-is, and index.html for
// any path the built app doesn't recognise as one of its own files, so
// client-side routing can take over.
//
// The embed directive below is the whole mechanism behind this task's
// "a build with no dist/ must fail loudly at compile time" requirement: a
// //go:embed pattern that names a directory which does not exist -- or
// exists but matches no files -- is a compile-time error ("pattern
// all:dist: no matching files found"), not a runtime one. There is no
// fallback path here that would let this package compile against a
// missing or empty dist/ and serve an empty page instead; the compiler
// refuses first. That is also why dist/ is Vite's own build.outDir
// (stats/web/vite.config.ts), not a copy this package's build step makes:
// a //go:embed pattern cannot climb above the directory holding the
// directive with "../", so the only way to embed Vite's output from here
// is to have Vite write it directly into this package.
package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// FS returns the embedded SPA build, rooted so that "index.html" (not
// "dist/index.html") names the app's entry point. It also verifies that
// index.html is actually present: //go:embed's own compile-time check
// guarantees dist/ existed and matched *some* file, but a build that
// produced an empty or malformed dist/ (a stray .gitkeep, a partially
// failed `vite build`) would still satisfy that and only fail at request
// time, silently, as a 404 on "/" -- exactly the runtime failure mode
// this task's requirement exists to rule out. Checking here, once, at
// startup, turns that into a loud startup error instead.
func FS() (fs.FS, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, fmt.Errorf("web: dist subtree: %w", err)
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, fmt.Errorf("web: embedded build has no index.html (was `vite build` run before `go build`?): %w", err)
	}
	return sub, nil
}

// Handler serves fsys (as returned by FS) as the SPA: a request naming a
// file that exists in fsys is served as that file; any other request --
// "/", "/changes/kan-16-myflow-stats-app", a path the client-side router
// owns and the build never produced a file for -- is answered with
// index.html so the SPA's own router can take over.
//
// Callers must never mount this handler under an API prefix. It has no
// notion of "/api/" at all -- it will happily "serve" index.html for
// "/api/v1/typo-ed-path" if asked to, which is precisely the swallowing
// this task's own non-negotiable requirement forbids. Guarding that is
// the caller's job: internal/api/server.go registers its own catch-all
// for unmatched "/api/" routes, and Go's http.ServeMux resolves the more
// specific "/api/" pattern before this handler (mounted at "/") is ever
// reached for a request under that prefix.
func Handler(fsys fs.FS) (http.Handler, error) {
	indexHTML, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		return nil, fmt.Errorf("web: reading embedded index.html: %w", err)
	}

	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "." {
			name = "index.html"
		}

		if _, err := fs.Stat(fsys, name); err != nil {
			// Not a file the build produced -- hand it to the SPA's own
			// router rather than returning a bare 404: everything under
			// "/" that isn't a real static asset is a client-side route
			// by definition, on this daemon's own routing split.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexHTML)
			return
		}
		fileServer.ServeHTTP(w, r)
	}), nil
}
