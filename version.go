// Package cgo — uniquement la source de version du produit.
// La chaîne de version vient du fichier VERSION embarqué : une seule autorité.
package cgo

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// Version — version du produit en clair (« 1.3.5 »), sans saut de ligne.
var Version = strings.TrimSpace(versionFile)
