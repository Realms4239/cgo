package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Realms4239/cgo/pkg/metrics"
)

// auditView — le dernier audit scellé, réduit à ce que le TUI rend. Les onglets
// 4 et 5 (Live, Résultats) sont la campagne, pas un audit : ce miroir leur donne
// le verdict et les quatre faits du dernier gel, sans sixième onglet.
//
// OK distingue « aucun audit » de « je n'ai pas pu lire » : un rapport
// illisible ne doit jamais s'afficher comme un lien jamais audité.
type auditView struct {
	OK  bool
	Why string

	AuditID     string           `json:"audit_id"`
	Timestamp   string           `json:"timestamp"`
	Site        string           `json:"site"`
	LinkType    string           `json:"link_type"`
	Grade       string           `json:"bloat_grade"`
	Verdict     metrics.Verdict  `json:"operator_verdict"`
	Measured    measuredView     `json:"measured"`
	Provenance  provenanceView   `json:"provenance"`
	Remediation *remediationView `json:"remediation,omitempty"`
	Validation  *validationView  `json:"routeros_validation,omitempty"`
	Kernel      *kernelView      `json:"kernel_telemetry,omitempty"`
}

type measuredView struct {
	RTTIdleP50    float64  `json:"rtt_idle_p50_ms"`
	RTTIdleP95    float64  `json:"rtt_idle_p95_ms"`
	RTTLoadedP50  float64  `json:"rtt_loaded_p50_ms"`
	RTTLoadedP95  float64  `json:"rtt_loaded_p95_ms"`
	BloatDeltaMs  float64  `json:"bloat_delta_ms"`
	QDIMs         float64  `json:"qdi_ms"`
	Throughput    float64  `json:"throughput_mbps"`
	LossPct       *float64 `json:"loss_pct"`
	LoadAttempted bool     `json:"load_attempted"`
	LoadMeasured  bool     `json:"load_measured"`
	Direction     string   `json:"direction"`
	FlowCount     int      `json:"flow_count"`
}

type provenanceView struct {
	SourceSHA256    string `json:"source_sha256"`
	TelemetrySHA256 string `json:"telemetry_sha256"`
}

type remediationView struct {
	Qdisc              string   `json:"qdisc"`
	Reason             string   `json:"reason"`
	LinuxTC            []string `json:"linux_tc"`
	RouterOS           string   `json:"routeros"`
	LinuxTCIngress     []string `json:"linux_tc_ingress"`
	RouterOSIngress    string   `json:"routeros_ingress"`
	UpMeasuredMbps     float64  `json:"up_measured_mbps"`
	UpShapedMbps       float64  `json:"up_shaped_mbps"`
	DownMeasuredMbps   float64  `json:"down_measured_mbps"`
	DownShapedMbps     float64  `json:"down_shaped_mbps"`
	RouterOSLanIface   string   `json:"routeros_lan_iface"`
	RouterOSLanAssumed bool     `json:"routeros_lan_assumed"`
}

type validationView struct {
	Version string      `json:"version"`
	Level   string      `json:"level"`
	Checks  []checkView `json:"checks"`
}

type checkView struct {
	Name string `json:"name"`
	Pass bool   `json:"pass"`
}

type kernelView struct {
	Available bool   `json:"available"`
	Source    string `json:"source"`
	Reason    string `json:"reason"`
}

const (
	paneNone       = 0
	paneDeployment = 1
	paneProof      = 2

	paneWidth  = 78
	platformRO = "routeros"
	platformLX = "linux"
)

type auditMsg struct{ v auditView }

func pollAudit(api string) tea.Cmd {
	return func() tea.Msg {
		cl := &http.Client{Timeout: 3 * time.Second, Transport: apiTransport(api)}
		resp, err := cl.Get(api + "/api/audit/status")
		if err != nil {
			return auditMsg{auditView{Why: "API injoignable — " + err.Error()}}
		}
		defer resp.Body.Close()
		var doc struct {
			Report json.RawMessage `json:"report"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			return auditMsg{auditView{Why: "rapport illisible — " + err.Error()}}
		}
		if len(doc.Report) == 0 || string(doc.Report) == "null" {
			return auditMsg{auditView{Why: "aucun audit scellé — lancez un audit"}}
		}
		var v auditView
		if err := json.Unmarshal(doc.Report, &v); err != nil {
			return auditMsg{auditView{Why: "rapport illisible — " + err.Error()}}
		}
		v.OK = true
		return auditMsg{v}
	}
}

// ---- les quatre faits ----

// verdictLine — une ligne, le mot du moteur et sa phrase. Jamais un chiffre
// recalculé : la cascade vit dans pkg/metrics et son résultat est gelé.
func verdictLine(v auditView) string {
	if !v.OK {
		return "Verdict indisponible — " + firstSet(v.Why, "aucun audit scellé")
	}
	if v.Verdict.Word == "" {
		return "Verdict indisponible — le rapport ne porte pas de cascade"
	}
	if v.Verdict.Sentence == "" {
		return v.Verdict.Word
	}
	return v.Verdict.Word + " — " + v.Verdict.Sentence
}

// auditFactLines — exactement quatre lignes. Un zéro n'est jamais un stand-in
// pour « absent » : une latence de 0 ms est physiquement impossible, un débit
// nul signifie que le flux n'a rien transféré, et la perte n'est rendue que si
// le rapport la porte (pointeur), pour qu'un 0 % réel reste un 0 %.
func auditFactLines(v auditView) [4]string {
	var out [4]string
	if !v.OK {
		return [4]string{"Repos : non mesuré", "Sous charge : non mesuré", "Débit : non mesuré", "Perte : non mesuré"}
	}
	m := v.Measured

	if m.RTTIdleP50 == 0 && m.RTTIdleP95 == 0 {
		out[0] = "Repos : non mesuré"
	} else {
		out[0] = fmt.Sprintf("Repos : p50 %.1f ms · p95 %.1f ms", m.RTTIdleP50, m.RTTIdleP95)
	}

	switch {
	case !m.LoadAttempted:
		out[1] = "Sous charge : charge non tentée"
	case !m.LoadMeasured:
		out[1] = "Sous charge : charge non établie"
	case m.RTTLoadedP95 == 0 && m.RTTLoadedP50 == 0:
		out[1] = "Sous charge : non mesuré"
	default:
		out[1] = fmt.Sprintf("Sous charge : p95 %.1f ms (objectif %d ms) · +%.1f ms", m.RTTLoadedP95, metrics.CAPDeadlineMs, m.BloatDeltaMs)
	}

	if m.Throughput <= 0 {
		out[2] = "Débit : non mesuré"
	} else {
		out[2] = fmt.Sprintf("Débit : %.1f Mb/s", m.Throughput)
	}

	if m.LossPct == nil {
		out[3] = "Perte : non mesuré"
	} else {
		out[3] = fmt.Sprintf("Perte : %.2f %%", *m.LossPct)
	}
	return out
}

// ---- panes ----

func firstSet(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// wrap78 — replie un texte à 78 colonnes sans jamais les dépasser. Un mot plus
// long que la colonne (une commande `tc` l'est souvent) est coupé dur : c'est
// préférable à un défilement horizontal, que le spec interdit.
func wrap78(s string) []string {
	s = strings.ReplaceAll(s, "\t", "    ")
	if strings.TrimSpace(s) == "" {
		return []string{""}
	}
	var out []string
	var cur []rune
	flush := func() {
		out = append(out, string(cur))
		cur = cur[:0]
	}
	for _, word := range strings.Split(s, " ") {
		w := []rune(word)
		if len(w) == 0 {
			continue
		}
		for len(w) > paneWidth {
			if len(cur) > 0 {
				flush()
			}
			out = append(out, string(w[:paneWidth]))
			w = w[paneWidth:]
		}
		if len(w) == 0 {
			continue
		}
		if len(cur) == 0 {
			cur = append(cur, w...)
			continue
		}
		if len(cur)+1+len(w) > paneWidth {
			flush()
		}
		cur = append(cur, ' ')
		cur = append(cur, w...)
	}
	if len(cur) > 0 || len(out) == 0 {
		flush()
	}
	return out
}

func wrapAll(lines []string) []string {
	var out []string
	for _, l := range lines {
		out = append(out, wrap78(l)...)
	}
	return out
}

func hasPlatform(v auditView, p string) bool {
	if v.Remediation == nil {
		return false
	}
	switch p {
	case platformRO:
		return strings.TrimSpace(v.Remediation.RouterOS) != ""
	case platformLX:
		return len(v.Remediation.LinuxTC) > 0
	}
	return false
}

// platformFor — la plateforme que `c` écrira. La préférence de l'opérateur
// n'est gardée que si le rapport porte un script pour elle : on n'écrit jamais
// un fichier vide parce qu'une touche a été pressée sur un rapport sans ce
// script.
func platformFor(v auditView, want string) string {
	if hasPlatform(v, want) {
		return want
	}
	if hasPlatform(v, platformRO) {
		return platformRO
	}
	if hasPlatform(v, platformLX) {
		return platformLX
	}
	return ""
}

func otherPlatform(p string) string {
	if p == platformRO {
		return platformLX
	}
	return platformRO
}

func qdiscLine(r *remediationView) string {
	q := firstSet(r.Qdisc, "aucun")
	if strings.TrimSpace(r.Reason) == "" {
		return "qdisc : " + q
	}
	return "qdisc : " + q + " — " + r.Reason
}

func rateLine(label string, measured, shaped float64) string {
	if measured <= 0 {
		return label + " : non mesuré — aucun script de ce sens"
	}
	return fmt.Sprintf("%s : %.1f mesurés → %.1f façonnés (0,95)", label, measured, shaped)
}

// replayLine — ce que la validation CHR/banc a prouvé, jamais ce qui a été
// appliqué. « appliquée » est un mot interdit ici.
func replayLine(v auditView) string {
	switch {
	case v.Validation == nil:
		return "Rejeu RouterOS : aucune preuve"
	case len(v.Validation.Checks) == 0:
		return "Rejeu RouterOS : non validée — aucune vérification"
	}
	for _, c := range v.Validation.Checks {
		if !c.Pass {
			return "Rejeu RouterOS : non validée — écarts dans le rapport"
		}
	}
	return "Rejeu RouterOS : validée par rejeu"
}

func sensorTier(v auditView) string {
	if v.Kernel == nil {
		return "standard — non instrumentée"
	}
	switch v.Kernel.Source {
	case "ebpf":
		return "eBPF — compteurs noyau"
	case "tcpinfo":
		return "TCP_INFO — amont du noyau : " + firstSet(v.Kernel.Reason, "raison non consignée")
	}
	return "standard — " + firstSet(v.Kernel.Reason, "sonde non instrumentée")
}

// deploymentLines — le pane « Déploiement & scripts », une commande par ligne
// avant repli : le repli ne mélange donc jamais deux commandes.
func deploymentLines(v auditView, platform string) []string {
	var lines []string
	lines = append(lines, "Déploiement du script — "+firstSet(v.Site, "site inconnu")+" · "+firstSet(platform, "aucune plateforme"))
	lines = append(lines, verdictLine(v))
	if r := v.Remediation; r != nil {
		lines = append(lines, rateLine("Envoi", r.UpMeasuredMbps, r.UpShapedMbps))
		lines = append(lines, rateLine("Réception", r.DownMeasuredMbps, r.DownShapedMbps))
		lines = append(lines, qdiscLine(r))
		switch platform {
		case platformRO:
			lines = append(lines, "", "--- RouterOS v7 — envoi ---")
			lines = append(lines, strings.Split(strings.TrimRight(r.RouterOS, "\n"), "\n")...)
			if strings.TrimSpace(r.RouterOSIngress) != "" {
				lines = append(lines, "", "--- RouterOS v7 — réception (LAN) ---")
				lines = append(lines, strings.Split(strings.TrimRight(r.RouterOSIngress, "\n"), "\n")...)
			}
		default:
			lines = append(lines, "", "--- Linux tc — envoi ---")
			lines = append(lines, r.LinuxTC...)
			if len(r.LinuxTCIngress) > 0 {
				lines = append(lines, "", "--- Linux tc — réception (IFB) ---")
				lines = append(lines, r.LinuxTCIngress...)
			}
		}
		if r.RouterOSLanAssumed {
			lines = append(lines, "", "Interface LAN supposée « "+firstSet(r.RouterOSLanIface, "bridge")+" » — à vérifier avant de coller")
		}
	} else {
		lines = append(lines, "Aucune remédiation : la capacité ou le RTT n'a pas été mesuré.")
	}
	lines = append(lines, "", replayLine(v))
	lines = append(lines, cueLine(v, platform))
	return wrapAll(lines)
}

func cueLine(v auditView, platform string) string {
	if platform == "" {
		return "esc revient · aucun script dans ce rapport"
	}
	if hasPlatform(v, otherPlatform(platform)) {
		return "c écrit " + scriptName(platform) + " · tab change de plateforme · esc revient"
	}
	return "c écrit " + scriptName(platform) + " · esc revient"
}

// proofLines — le pane « Preuves & télémétrie brute ».
func proofLines(v auditView) []string {
	var lines []string
	lines = append(lines, "Preuves — "+firstSet(v.AuditID, "aucun audit"))
	lines = append(lines, verdictLine(v))
	for _, f := range auditFactLines(v) {
		lines = append(lines, f)
	}
	m := v.Measured
	lines = append(lines, fmt.Sprintf("Écart pire cas : %.1f ms · QDI %.1f ms", m.BloatDeltaMs, m.QDIMs))
	lines = append(lines, fmt.Sprintf("Sens : %s · flux : %d", firstSet(m.Direction, "non consigné"), m.FlowCount))
	lines = append(lines, "Note : "+firstSet(v.Grade, "indisponible"))
	lines = append(lines, "Télémétrie : "+sensorTier(v))
	if v.Provenance.SourceSHA256 != "" {
		lines = append(lines, "Source sha256 : "+short12(v.Provenance.SourceSHA256))
	}
	if v.Provenance.TelemetrySHA256 != "" {
		lines = append(lines, "Télémétrie sha256 : "+short12(v.Provenance.TelemetrySHA256))
	}
	if v.AuditID != "" {
		lines = append(lines, "Rapport scellé : /api/audit/report?audit_id="+v.AuditID)
	}
	lines = append(lines, "Duel 4G : /api/duel?profile=P2")
	lines = append(lines, "esc revient")
	return wrapAll(lines)
}

func short12(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12]
}

// ---- écriture du script ----

func scriptName(platform string) string {
	if platform == platformRO {
		return "remediation.routeros.rsc"
	}
	return "remediation.linux.sh"
}

// safeAuditID — l'identifiant vient du serveur et devient un chemin : la même
// règle que validRunID côté API (E11). Rien qui remonte, rien de vide.
func safeAuditID(id string) error {
	if id == "" || id == "." || id == ".." {
		return fmt.Errorf("identifiant d'audit invalide")
	}
	for _, r := range id {
		ok := r == '-' || r == '_' || r == '.' ||
			(r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		if !ok {
			return fmt.Errorf("identifiant d'audit invalide : %q", id)
		}
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("identifiant d'audit invalide : %q", id)
	}
	return nil
}

// scriptFor — le texte exact que `c` écrit. RouterOS reste verbatim (c'est un
// script collé dans le routeur) ; Linux est une suite de commandes, une par
// ligne, la réception derrière une ligne de commentaire pour qu'un copier-coller
// partiel ne mélange pas les deux sens.
func scriptFor(v auditView, platform string) (name, text string, err error) {
	if v.Remediation == nil {
		return "", "", fmt.Errorf("aucune remédiation dans ce rapport")
	}
	switch platform {
	case platformRO:
		s := strings.TrimRight(v.Remediation.RouterOS, "\n")
		if strings.TrimSpace(s) == "" {
			return "", "", fmt.Errorf("aucun script RouterOS dans ce rapport")
		}
		return scriptName(platform), s + "\n", nil
	case platformLX:
		if len(v.Remediation.LinuxTC) == 0 {
			return "", "", fmt.Errorf("aucun script Linux dans ce rapport")
		}
		lines := append([]string{}, v.Remediation.LinuxTC...)
		if len(v.Remediation.LinuxTCIngress) > 0 {
			lines = append(lines, "", "# réception (IFB)")
			lines = append(lines, v.Remediation.LinuxTCIngress...)
		}
		return scriptName(platform), strings.Join(lines, "\n") + "\n", nil
	}
	return "", "", fmt.Errorf("plateforme inconnue : %q", platform)
}

// auditScriptRoot — l'emplacement de production, à côté des rapports scellés.
// Variable pour que le test le vérifie sans écrire dans data/ ; jamais /tmp.
var auditScriptRoot = filepath.Join("data", "audits")

// writeRemediation — écrit à côté du rapport scellé, jamais dans /tmp, jamais
// hors de data/audits.
func writeRemediation(root string, v auditView, platform string) (string, error) {
	if err := safeAuditID(v.AuditID); err != nil {
		return "", err
	}
	name, text, err := scriptFor(v, platform)
	if err != nil {
		return "", err
	}
	rootClean := filepath.Clean(root)
	dir := filepath.Join(rootClean, v.AuditID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if clean := filepath.Clean(path); !strings.HasPrefix(clean, rootClean+string(os.PathSeparator)) {
		return "", fmt.Errorf("chemin hors de %s : %s", rootClean, clean)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// auditMirror — l'en-tête d'audit des onglets 4 et 5 : le verdict, les quatre
// faits, et la ligne qui dit ce que d/p/c feraient. Rien d'autre n'est ajouté à
// ces vues : la campagne garde ses propres cartes.
func (m modelTUI) auditMirror() string {
	var b strings.Builder
	style := stMuted
	switch string(m.audit.Verdict.Level) {
	case "green":
		style = stOK
	case "amber":
		style = stWarn
	case "red":
		style = stErr
	}
	b.WriteString(style.Render(verdictLine(m.audit)) + "\n")
	for _, f := range auditFactLines(m.audit) {
		b.WriteString("  " + f + "\n")
	}
	hint := "dernier audit gelé · d scripts · p preuves"
	if pf := platformFor(m.audit, m.platform); pf != "" {
		hint += " · c écrit " + scriptName(pf)
	}
	b.WriteString(stMuted.Render("  "+hint) + "\n")
	if m.status != "" {
		b.WriteString(stMuted.Render("  "+m.status) + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

// copyScript — l'action de `c` : écrire à côté du rapport scellé, puis tenter
// OSC 52. Le statut nomme le chemin et dit que la copie est tentée, pas
// réussie : aucun terminal ne renvoie d'accusé sur OSC 52.
func (m modelTUI) copyScript() (tea.Model, tea.Cmd) {
	if !m.audit.OK {
		m.status = firstSet(m.audit.Why, "aucun audit scellé — lancez un audit")
		return m, nil
	}
	platform := platformFor(m.audit, m.platform)
	if platform == "" {
		m.status = "ce rapport ne porte aucun script de remédiation"
		return m, nil
	}
	m.platform = platform
	path, err := writeRemediation(auditScriptRoot, m.audit, platform)
	if err != nil {
		m.status = "écriture impossible — " + err.Error()
		return m, nil
	}
	_, text, err := scriptFor(m.audit, platform)
	if err != nil {
		m.status = path + " · copie OSC 52 impossible — " + err.Error()
		return m, nil
	}
	m.status = path + " · copie OSC 52 tentée (le terminal peut l'ignorer)"
	return m, tea.Printf("%s", osc52(text))
}

// osc52 — le presse-papier du terminal. Aucune réponse n'atteste la copie : le
// statut dit « tentée », jamais « copiée ».
func osc52(text string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(text)) + "\a"
}
