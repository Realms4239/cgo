// Package frontend embeds the built SPA (web/frontend/dist).
// The dist directory always exists via its tracked .gitkeep so the module
// builds before the first frontend build; Makefile rebuilds it first.
package frontend

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// ValidateEmbeddedAssets refuses a placeholder embed: dist/index.html alone
// serves a blank dashboard. The binary must carry the hashed JS bundle.
func ValidateEmbeddedAssets() error {
	if _, err := fs.Stat(embedded, "dist/index.html"); err != nil {
		return fmt.Errorf("embedded SPA: dist/index.html missing: %w", err)
	}
	js, err := fs.Glob(embedded, "dist/assets/*.js")
	if err != nil {
		return fmt.Errorf("embedded SPA: dist/assets unreadable: %w", err)
	}
	if len(js) == 0 {
		return fmt.Errorf("embedded SPA: no dist/assets/*.js — dist/index.html is a placeholder, run the frontend build")
	}
	return nil
}

// FS serves dist contents at root ("/index.html" etc.).
var FS fs.FS = func() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}()
