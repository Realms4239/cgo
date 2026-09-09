package kit

import "testing"

func TestIsVBoxGuestSpace(t *testing.T) {
	for in, want := range map[string]bool{
		"10.0.2.15":         true,
		" 10.0.2.1 ":        true,
		"10.0.2.1":          true,
		"192.168.174.128":   false,
		"127.0.0.1":         false,
		"10.0.3.5":          false,
		"":                  false,
		"meteolink.dev":     false,
	} {
		if got := isVBoxGuestSpace(in); got != want {
			t.Errorf("isVBoxGuestSpace(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestWantAdapterDeterministic(t *testing.T) {
	// Chemins qui ne dépendent d'aucune interface de la machine.
	if got := wantAdapter("bogus-hyp", "192.168.1."); got != "" {
		t.Errorf("hyp inconnu → \"\", got %q", got)
	}
	if got := wantAdapter("virtualbox", "192.168.1."); got != "" {
		t.Errorf("vbox hors 192.168.56 → diagnostic seul, got %q", got)
	}
	if got := wantAdapter("", "192.168.1."); got != "" {
		t.Errorf("hyp vide → \"\", got %q", got)
	}
}

func TestAdapterHasOtherSubnetAbsent(t *testing.T) {
	if adapterHasOtherSubnet("no-such-adapter-xyz", "192.168.99.") {
		t.Error("alias inexistant → false")
	}
}

func TestResolveHypNoVM(t *testing.T) {
	c := &Config{}
	if got := resolveHyp(c); got != "" {
		t.Errorf("sans VM verrouillée → \"\", got %q", got)
	}
}

func TestVnetSubnet(t *testing.T) {
	for in, want := range map[string]string{
		"192.168.174.128": "192.168.174.",
		"10.0.2.15":       "10.0.2.",
		"abc":             "abc",
	} {
		if got := vnetSubnet(in); got != want {
			t.Errorf("vnetSubnet(%q) = %q, want %q", in, got, want)
		}
	}
}
