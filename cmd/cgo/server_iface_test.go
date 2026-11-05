package main

import "testing"

func TestEnvSampler(t *testing.T) {
	env := map[string]string{}
	get := func(k string) string { return env[k] }

	if s := envSampler(get, "windows"); s != nil {
		t.Fatal("Windows sans CGO_IFACE doit désactiver l'échantillonneur")
	}
	env["CGO_IFACE"] = "c-cli"
	if s := envSampler(get, "windows"); s != nil {
		t.Fatal("Windows avec CGO_IFACE doit journaliser et désactiver")
	}

	s := envSampler(get, "linux")
	if s == nil {
		t.Fatal("Linux avec CGO_IFACE doit créer l'échantillonneur")
	}
	got := s.Latest()
	if got.Name != "c-cli" || got.Source != "unknown" || got.Gap != "warmup" {
		t.Fatalf("latest = %+v", got)
	}

	env["CGO_IFACE_SOURCE"] = "prod"
	if got := envSampler(get, "linux").Latest(); got.Source != "prod" {
		t.Fatalf("source = %q, want prod", got.Source)
	}
	env["CGO_IFACE_SOURCE"] = "production"
	if got := envSampler(get, "linux").Latest(); got.Source != "unknown" {
		t.Fatalf("source inconnue = %q, want unknown", got.Source)
	}
}
