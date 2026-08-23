// Package frontend embeds the built SPA (web/frontend/dist).
// The dist directory always exists via its tracked .gitkeep so the module
// builds before the first frontend build; Makefile rebuilds it first.
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// FS serves dist contents at root ("/index.html" etc.).
var FS fs.FS = func() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}()
