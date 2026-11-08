// Package cite — garde de citation pour les rapports d'audit scellés.
// Port Go de web/frontend/src/lib/citation.ts (auditCitation) ; les deux
// implémentations obéissent à la même règle : aucune citation partielle
// sans scellé — l'absence de jeton n'est jamais rendue par un zéro.
package cite

import "strings"

// AuditMeasured — miroir du sous-ensemble mesuré lu par la citation.
type AuditMeasured struct {
	Direction         string `json:"direction,omitempty"`
	LoadMeasured      bool   `json:"load_measured,omitempty"`
	LoadMeasuredDown  bool   `json:"load_measured_down,omitempty"`
}

// AuditProvenance — le scellé transporté par la citation.
type AuditProvenance struct {
	SourceSHA256 string `json:"source_sha256,omitempty"`
}

// AuditReport — miroir des champs consommés par AuditCitation.
type AuditReport struct {
	AuditID    string         `json:"audit_id"`
	Timestamp  string         `json:"timestamp,omitempty"`
	Site       string         `json:"site,omitempty"`
	LinkType   string         `json:"link_type,omitempty"`
	Provider   string         `json:"provider,omitempty"`
	Measured   AuditMeasured  `json:"measured"`
	Provenance *AuditProvenance `json:"provenance,omitempty"`
}

// directionLabel — miroir de directionLabel (web/frontend/src/lib/auditFlow.ts).
func directionLabel(direction string) string {
	switch strings.TrimSpace(direction) {
	case "down":
		return "descendant (down)"
	case "both":
		return "les deux sens (both)"
	case "up":
		return "montant (up)"
	default:
		return "—"
	}
}

// AuditCitation reproduit auditCitation : "" quand l'identifiant ou le
// scellé manque (après trim) ; jamais de citation partielle.
func AuditCitation(report AuditReport) string {
	id := strings.TrimSpace(report.AuditID)
	if id == "" {
		return ""
	}
	var seal string
	if report.Provenance != nil {
		seal = strings.TrimSpace(report.Provenance.SourceSHA256)
	}
	if seal == "" {
		return ""
	}

	site := strings.TrimSpace(report.Site)
	link := strings.TrimSpace(report.LinkType)
	provider := strings.TrimSpace(report.Provider)
	linkLabel := strings.Join(nonEmpty(link, provider), "/")
	dir := directionLabel(report.Measured.Direction)

	details := strings.Join(nonEmpty(site, linkLabel, strings.TrimSpace(report.Timestamp)), ", ")
	loadMeasured := report.Measured.LoadMeasured || report.Measured.LoadMeasuredDown
	var measurement string
	if loadMeasured {
		if dir != "—" {
			measurement = "direction " + dir
		} else {
			measurement = "charge mesurée"
		}
	} else if dir != "—" {
		measurement = "direction " + dir + " · charge non mesurée"
	} else {
		measurement = "charge non mesurée"
	}
	tail := strings.Join(nonEmpty(details, measurement), ", ")
	var out string
	if tail != "" {
		out = "Audit CGO " + id + " — " + tail + "."
	} else {
		out = "Audit CGO " + id + "."
	}
	out += " Rapport scellé : source_sha256 " + seal + "."
	return out
}

func nonEmpty(xs ...string) []string {
	var out []string
	for _, x := range xs {
		if x != "" {
			out = append(out, x)
		}
	}
	return out
}
