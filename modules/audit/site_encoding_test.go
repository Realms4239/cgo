package audit

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestRunRefusesAnUndecodableSite — un nom de site illisible a déjà été scellé
// tel quel dans link_audit.csv et dans un audit_report.json (« D\uFFFDpartement X »),
// et la console l'affiche depuis. Le moteur ne peut pas savoir qui l'a mal encodé,
// mais il peut refuser d'archiver ce qu'il ne peut pas relire : la porte est ici,
// partagée par le CLI et l'API, et elle tombe AVANT la première sonde — mesurer
// pour jeter ensuite n'est pas un refus, c'est une perte de temps d'antenne.
func TestRunRefusesAnUndecodableSite(t *testing.T) {
	probed := 0
	d := Deps{
		Ping: func(context.Context, string, int) []float64 {
			probed++
			return []float64{20, 21, 22}
		},
		Small: func(context.Context) (float64, error) { probed++; return 15, nil },
		BulkLoad: func(context.Context, string, string, int) LoadOutcome {
			probed++
			return LoadOutcome{Bytes: 1_000_000, PerFlow: []uint64{1_000_000}, Elapsed: time.Second}
		},
	}
	bad := []string{
		"D\uFFFDpartement X",      // le caractère de remplacement, déjà scellé par le passé
		"D\xe9partement X",        // latin-1 lu comme UTF-8
		"D\xc3\xa9partement \xc3", // dernière séquence tronquée
	}
	for _, site := range bad {
		probed = 0
		_, err := Run(context.Background(), Params{AuditID: "audit-encodage", Site: site, LinkType: "4g", Duration: 2, Target: "10.30.0.2"}, d)
		if err == nil {
			t.Errorf("site %q accepté : il serait scellé illisible dans le CSV et le rapport", site)
			continue
		}
		if !strings.Contains(err.Error(), "non décodable") {
			t.Errorf("site %q : refus pour la mauvaise raison : %v", site, err)
		}
		if probed != 0 {
			t.Errorf("site %q : %d sondes avant le refus — un refus doit précéder la mesure", site, probed)
		}
	}

	probed = 0
	if _, err := Run(context.Background(), Params{AuditID: "audit-ok", Site: "Département X", LinkType: "4g", Duration: 2, Target: "10.30.0.2"}, d); err != nil {
		t.Errorf("un nom français valide ne doit pas être refusé : %v", err)
	}
	if probed == 0 {
		t.Error("aucune sonde sur le cas valide : le test ne prouverait rien")
	}
}

// TestRunRefusesAnUndecodableProvider — même porte pour le second champ libre.
func TestRunRefusesAnUndecodableProvider(t *testing.T) {
	d := Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{20} },
		Small: func(context.Context) (float64, error) { return 15, nil },
	}
	_, err := Run(context.Background(), Params{AuditID: "a", Site: "ok", Provider: "y\uFFFDs", Duration: 2, Target: "x"}, d)
	if err == nil || !strings.Contains(err.Error(), "non décodable") {
		t.Fatalf("fournisseur illisible : %v", err)
	}
}
