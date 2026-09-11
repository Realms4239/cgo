package kit

import (
	"testing"
)

// NATForwardTarget — racine commune keysetup/Next/Diagnostiquer : l'espace
// invité 10.0.2.x part au forward SANS pilote (décision topologique pure,
// donc testable sans hyperviseur) ; le loopback explicite n'est jamais
// repris ; sans VM il n'y a pas de cas NAT.
func TestNATForwardTarget(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		natPort string
		wantH   string
		wantP   string
		wantNAT bool
	}{
		{"invitée NAT", "10.0.2.15", "", "127.0.0.1", "2222", true},
		{"invitée NAT port custom", "10.0.2.15", "2223", "127.0.0.1", "2223", true},
		{"gateway NAT", "10.0.2.1", "", "127.0.0.1", "2222", true},
		{"loopback gardé", "127.0.0.1", "", "", "", false},
		{"localhost gardé", "localhost", "", "", "", false},
		{"auto non résolu", "auto", "", "", "", false},
		{"vide", "", "", "", "", false},
		{"bridgé sans VM", "192.168.174.128", "", "", "", false},
		{"10.0.3.x pas invité", "10.0.3.5", "", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := &Config{SSHHost: c.host, NatHostPort: c.natPort}
			h, p, ok := NATForwardTarget(cfg, c.host)
			if ok != c.wantNAT || h != c.wantH || p != c.wantP {
				t.Errorf("NATForwardTarget(%q) = (%q,%q,%v), want (%q,%q,%v)",
					c.host, h, p, ok, c.wantH, c.wantP, c.wantNAT)
			}
		})
	}
}

// staleSSHHost — le scan rafraîchit auto/vide/ancienne IP, jamais un
// forward explicite, jamais l'espace invité NAT (propriété d'ensure).
func TestStaleSSHHost(t *testing.T) {
	cases := []struct {
		name, current, discovered string
		want                      bool
	}{
		{"auto vers IP", "auto", "192.168.174.129", true},
		{"vide vers IP", "", "192.168.174.129", true},
		{"ancienne IP", "192.168.174.128", "192.168.174.129", true},
		{"identique", "192.168.174.128", "192.168.174.128", false},
		{"forward gardé", "127.0.0.1", "192.168.174.129", false},
		{"localhost gardé", "localhost", "10.0.2.15", false},
		{"espace NAT jamais écrit", "192.168.174.128", "10.0.2.15", false},
		{"auto vers NAT jamais écrit", "auto", "10.0.2.15", false},
		{"découvert vide", "auto", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := staleSSHHost(c.current, c.discovered); got != c.want {
				t.Errorf("staleSSHHost(%q,%q) = %v, want %v",
					c.current, c.discovered, got, c.want)
			}
		})
	}
}

// dashStale — ORDONNÉ sur majeur.mineur (le patch ne change pas les
// paliers) : seul have < want est périmé ; have > want = kit local
// derrière (voir TestDashAhead, jamais un deploy) ; doute
// (vide/illisible) ou patch seul : jamais périmé, on ne bloque pas.
func TestDashStale(t *testing.T) {
	cases := []struct {
		name, have, want string
		stale            bool
	}{
		{"identique", "1.2.13", "1.2.13", false},
		{"patch seul", "1.2.12", "1.2.13", false},
		{"mineur antérieur", "1.2.12", "1.3.0", true},
		{"majeur antérieur", "1.9.9", "2.0.0", true},
		{"distant plus neuf (pas périmé)", "1.3.0", "1.2.13", false},
		{"majeur plus neuf (pas périmé)", "2.0.0", "1.9.9", false},
		{"mineur plus neuf (pas périmé)", "1.4.0", "1.3.9", false},
		{"préfixe v", "v1.2.12", "1.3.0", true},
		{"want vide (compat)", "1.2.12", "", false},
		{"have vide", "", "1.2.13", false},
		{"illisible", "dev", "1.2.13", false},
		{"court", "1", "1.2.13", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dashStale(c.have, c.want); got != c.stale {
				t.Errorf("dashStale(%q,%q) = %v, want %v", c.have, c.want, got, c.stale)
			}
		})
	}
}

// dashAhead — miroir ordonné de dashStale : seul have > want (majeur.mineur)
// est « ahead » ; le patch seul, l'égal et le doute ne le sont jamais.
func TestDashAhead(t *testing.T) {
	cases := []struct {
		name, have, want string
		ahead            bool
	}{
		{"mineur plus neuf", "1.3.0", "1.2.13", true},
		{"majeur plus neuf", "2.0.0", "1.9.9", true},
		{"préfixe v plus neuf", "v1.4.0", "1.3.9", true},
		{"identique", "1.2.13", "1.2.13", false},
		{"patch seul", "1.2.14", "1.2.13", false},
		{"mineur antérieur", "1.2.12", "1.3.0", false},
		{"majeur antérieur", "1.9.9", "2.0.0", false},
		{"want vide (compat)", "1.3.0", "", false},
		{"have vide", "", "1.2.13", false},
		{"illisible", "dev", "1.2.13", false},
		{"court", "1", "1.2.13", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dashAhead(c.have, c.want); got != c.ahead {
				t.Errorf("dashAhead(%q,%q) = %v, want %v", c.have, c.want, got, c.ahead)
			}
		})
	}
}

// dnsMapsTo — l'espace invité NAT ne mappe jamais (garde pré-résolution,
// déterministe sans réseau : le palier dns rend « attente » sous NAT).
func TestDNSMapsToGuestSpace(t *testing.T) {
	for _, ip := range []string{"10.0.2.15", "10.0.2.1", " 10.0.2.15 "} {
		if dnsMapsTo("meteolink.dev", ip) {
			t.Errorf("dnsMapsTo(meteolink.dev, %q) = true, want false (espace invité NAT)", ip)
		}
	}
}

// sshProbeAs — la sonde surchargée ne mute jamais la config partagée et
// propage l'erreur (codes de sortie honnêtes en aval). Cible invalide :
// échec rapide, sans réseau.
func TestSSHProbeAsNoMutate(t *testing.T) {
	c := &Config{SSHUser: "u", SSHHost: "192.168.174.128", SSHPort: "22", SSHKey: "~/.ssh/id_ed25519"}
	if _, err := sshProbeAs(c, "nonexistent.invalid", "22"); err == nil {
		t.Error("sonde vers l'invalide : want error")
	}
	if c.SSHHost != "192.168.174.128" || c.SSHPort != "22" {
		t.Errorf("config mutée : %+v", c)
	}
}
