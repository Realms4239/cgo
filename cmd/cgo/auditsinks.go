package main

import "github.com/Realms4239/cgo/audit"

// La détection des puits vit dans pkg/audit depuis le plan 3 : le CLI et
// l'API tranchent avec la MÊME échelle (dualns → legacy → external). Ces
// symboles restent des délégations pour les sites CLI historiques.
type Sinks = audit.Sinks

// ResolveSinks — délégation au paquet partagé.
func ResolveSinks() Sinks { return audit.ResolveSinks() }

// envOrBulk — délégation : CGO_BULK_ADDR écrase la détection.
func envOrBulk(s Sinks) string { return audit.EnvOrBulk(s) }
