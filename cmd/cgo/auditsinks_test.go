package main

import (
	"net"
	"testing"
)

func TestResolveSinksPrefersLiveListener(t *testing.T) {
	// Un puits qui ÉCOUTE est le banc : le mode doit le refléter, pas se
	// fier à une adresse en dur. Sur un hôte de dev sans banc, la détection
	// doit retomber proprement sur external SANS charge implicite — jamais
	// de note sans charge.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("loopback indisponible")
	}
	defer ln.Close()
	s := ResolveSinks()
	if s.Mode == "" {
		t.Fatal("mode vide — la détection doit toujours trancher (au pire external)")
	}
	switch s.Mode {
	case "dualns":
		// Le puits vit DANS ns_server (s-edge 10.30.0.2) : cible, petit
		// objet et charge traversent les QUATRE sauts. 10.30.0.1 est e-srv
		// (côté edge) — y terminer mesurerait trois sauts, pas quatre.
		if s.BulkAddr != "10.30.0.2:5201" || s.Target != "10.30.0.2" || s.SmallURL != "http://10.30.0.2:8081/small" {
			t.Fatalf("dualns : %+v", s)
		}
	case "legacy":
		if s.BulkAddr != "10.200.0.1:5201" || s.Target != "10.200.0.1" {
			t.Fatalf("legacy : %+v", s)
		}
	case "external":
		if s.BulkAddr != "" {
			t.Fatalf("mode externe : pas de charge implicite (jamais de A+ sans charge) : %+v", s)
		}
		if s.Target == "" {
			t.Fatal("mode externe : cible ping par défaut attendue")
		}
	default:
		t.Fatalf("mode inconnu : %q", s.Mode)
	}
}

func TestSinksEnvOrBulkOverride(t *testing.T) {
	// L'opérateur garde la main : CGO_BULK_ADDR écrase la détection
	// (liaison de terrain explicite), même en mode external.
	t.Setenv("CGO_BULK_ADDR", "192.0.2.10:5201")
	s := Sinks{Mode: "external", Target: "8.8.8.8"}
	if got := s.envOrBulk(); got != "192.0.2.10:5201" {
		t.Fatalf("override env ignoré : %q", got)
	}
}

func TestSinksEnvOrBulkDetected(t *testing.T) {
	// Sans variable d'environnement, le puits DÉTECTÉ passe (banc vivant).
	s2 := Sinks{Mode: "dualns", Target: "10.30.0.1", SmallURL: "http://10.30.0.1:8081/small", BulkAddr: "10.30.0.1:5201"}
	if got := s2.envOrBulk(); got != "10.30.0.1:5201" {
		t.Fatalf("sans env, le puits détecté doit passer : %q", got)
	}
}
