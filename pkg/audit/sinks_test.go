package audit

import "testing"

// TestResolveSinksPrecedence — la sonde est simulée : dualns gagne quand son
// puits écoute, legacy prend le relais sinon, external en dernier SANS charge
// implicite (jamais de note sans charge). L'échelle est partagée CLI/API.
func TestResolveSinksPrecedence(t *testing.T) {
	cases := []struct {
		name string
		up   map[string]bool
		want Sinks
	}{
		{
			name: "dualns prioritaire",
			up:   map[string]bool{"10.30.0.2:5201": true, "10.200.0.1:5201": true},
			want: Sinks{Mode: "dualns", Target: "10.30.0.2", SmallURL: "http://10.30.0.2:8081/small", BulkAddr: "10.30.0.2:5201"},
		},
		{
			name: "legacy si dualns muet",
			up:   map[string]bool{"10.200.0.1:5201": true},
			want: Sinks{Mode: "legacy", Target: "10.200.0.1", SmallURL: "http://10.200.0.1:8081/small", BulkAddr: "10.200.0.1:5201"},
		},
		{
			name: "external sans puits",
			up:   map[string]bool{},
			want: Sinks{Mode: "external", Target: "8.8.8.8"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var probed []string
			got := resolveSinks(func(addr string) bool {
				probed = append(probed, addr)
				return tc.up[addr]
			})
			if got != tc.want {
				t.Fatalf("sinks = %+v, want %+v", got, tc.want)
			}
			if len(probed) == 0 || probed[0] != "10.30.0.2:5201" {
				t.Fatalf("ordre de sonde : %v — dualns doit être testé d'abord", probed)
			}
		})
	}
}

func TestSinksEnvOrBulkOverride(t *testing.T) {
	// L'opérateur garde la main : CGO_BULK_ADDR écrase la détection
	// (liaison de terrain explicite), même en mode external.
	t.Setenv("CGO_BULK_ADDR", "192.0.2.10:5201")
	s := Sinks{Mode: "external", Target: "8.8.8.8"}
	if got := EnvOrBulk(s); got != "192.0.2.10:5201" {
		t.Fatalf("override env ignoré : %q", got)
	}
}

func TestSinksEnvOrBulkDetected(t *testing.T) {
	// Sans variable d'environnement, le puits DÉTECTÉ passe (banc vivant).
	s2 := Sinks{Mode: "dualns", Target: "10.30.0.2", SmallURL: "http://10.30.0.2:8081/small", BulkAddr: "10.30.0.2:5201"}
	if got := EnvOrBulk(s2); got != "10.30.0.2:5201" {
		t.Fatalf("sans env, le puits détecté doit passer : %q", got)
	}
}
