package main

import (
	"net"
	"time"
)

// Sinks — la détection du banc décide des défauts : dualns (banc durci,
// puits 10.30.0.1), legacy (kit/testbed.sh, 10.200.0.1), external (lien
// réel — pas de charge implicite : jamais de note sans charge).
type Sinks struct {
	Mode     string
	Target   string
	SmallURL string
	BulkAddr string
}

// probeTCP — un listener qui répond à 300 ms près : c'est un banc vivant.
func probeTCP(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

// ResolveSinks — dualns d'abord (le chemin de production), legacy ensuite
// (reproduction du corpus gelé), external sinon (audit honnête sans charge
// implicite ; l'opérateur fournit -bulk/CGO_BULK_ADDR pour un puits terrain).
func ResolveSinks() Sinks {
	// Le puits vit dans ns_server (s-edge 10.30.0.2) : les QUATRE sauts
	// sont sur le chemin mesuré. 10.30.0.1 est e-srv (côté edge) — y
	// terminerait la mesure un saut trop tôt (correction plan 1).
	if probeTCP("10.30.0.2:5201") {
		return Sinks{Mode: "dualns", Target: "10.30.0.2", SmallURL: "http://10.30.0.2:8081/small", BulkAddr: "10.30.0.2:5201"}
	}
	if probeTCP("10.200.0.1:5201") {
		return Sinks{Mode: "legacy", Target: "10.200.0.1", SmallURL: "http://10.200.0.1:8081/small", BulkAddr: "10.200.0.1:5201"}
	}
	return Sinks{Mode: "external", Target: "8.8.8.8"}
}

// envOrBulk — l'opérateur garde la main : CGO_BULK_ADDR écrase la détection
// (liaison de terrain explicite), même en mode external.
func (s Sinks) envOrBulk() string {
	if v := envDefault("CGO_BULK_ADDR", ""); v != "" {
		return v
	}
	return s.BulkAddr
}
