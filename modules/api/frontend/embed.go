// Package frontend — stub SPA pour le module extrait cgo/api.
// Le vrai SPA (web/frontend/dist) reste construit côté web/frontend ;
// ce stub garde le module autonome jusqu'à la bascule (cutover).
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed index.html
var stub embed.FS

// FS sert le stub à la racine ; le handler api attache le repli SPA sur "/".
var FS fs.FS = stub
