// Package exporter — collecteur Prometheus honnête (stdlib seulement).
// Contrat : # HELP / # TYPE avant le premier échantillon, Content-Type
// « text/plain; version=0.0.4 », et l'absent n'est jamais un zéro.
package exporter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Sample — un échantillon mesuré. Value est déjà formatée ; l'appelant ne
// pousse que les valeurs mesurées, jamais un zéro par défaut.
type Sample struct {
	Labels map[string]string
	Value  string
}

// Render construit le corps d'exposition Prometheus. Les séries absentes
// (aucun échantillon) n'émettent que HELP/TYPE, jamais « nom{ } 0 ».
func Render(samples map[string][]Sample, help map[string]string) string {
	var b strings.Builder
	names := make([]string, 0, len(samples))
	for n := range samples {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, help[name])
		typ := "gauge"
		if strings.HasSuffix(name, "_total") {
			typ = "counter"
		}
		fmt.Fprintf(&b, "# TYPE %s %s\n", name, typ)
		for _, s := range samples[name] {
			parts := make([]string, 0, len(s.Labels))
			for k, v := range s.Labels {
				parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(parts)
			fmt.Fprintf(&b, "%s{%s} %s\n", name, strings.Join(parts, ","), s.Value)
		}
	}
	return b.String()
}

// Handler sert l'exposition avec le Content-Type Prometheus contractuel.
// collect ne renvoie que les séries mesurées ; l'absence = aucun échantillon.
func Handler(collect func() (map[string][]Sample, map[string]string)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		samples, help := collect()
		_, _ = w.Write([]byte(Render(samples, help)))
	})
}
