package doctor

import "testing"

const ethtoolSample = `Features for c-cli:
rx-checksumming: on
tx-checksumming: on
scatter-gather: on
tcp-segmentation-offload: off
generic-segmentation-offload: off
generic-receive-offload: off
large-receive-offload: off [fixed]
rx-gro-hw: off [fixed]
tx-checksum-ip-generic: on
`

func TestParseOffloadFeatures(t *testing.T) {
	// Le parseur doit lire la sortie ethtool -k réelle : lignes
	// « feature: state », états on/off, suffixes [fixed]/[requested on]
	// tolérés sans fausser l'état.
	m := ParseOffloadFeatures(ethtoolSample)
	if len(m) == 0 {
		t.Fatal("aucune feature parsée")
	}
	if v, ok := m["tcp-segmentation-offload"]; !ok || v {
		t.Fatalf("TSO attendu off : %v (%+v)", v, m)
	}
	if v, ok := m["generic-receive-offload"]; !ok || v {
		t.Fatalf("GRO attendu off : %v", v)
	}
	if v, ok := m["rx-checksumming"]; !ok || !v {
		t.Fatalf("une feature « on » doit être true : %+v", m)
	}
	if _, ok := m["large-receive-offload"]; !ok {
		t.Fatal("[fixed] ne doit pas faire disparaître la ligne")
	}
	// L'en-tête « Features for … » n'est pas une feature.
	if _, ok := m["features for c-cli"]; ok {
		t.Fatal("ligne d'en-tête parsée comme feature")
	}
	if got := ParseOffloadFeatures(""); len(got) != 0 {
		t.Fatalf("sortie vide ⇒ map vide : %+v", got)
	}
}

func TestReportHasBenchOffloadsCheck(t *testing.T) {
	// Quelle que soit la plateforme, la porte existe et dit son état —
	// « n/a banc absent » sur un hôte sans banc, jamais un silence.
	for _, mode := range []string{"observe", "full"} {
		_, checks := Report(mode)
		found := false
		for _, c := range checks {
			if c.Name == "bench-offloads" {
				found = true
				switch c.Status {
				case "ok", "warn", "fail":
				default:
					t.Fatalf("mode %s: statut invalide %q", mode, c.Status)
				}
			}
		}
		if !found {
			t.Fatalf("mode %s: porte bench-offloads absente du rapport", mode)
		}
	}
}
