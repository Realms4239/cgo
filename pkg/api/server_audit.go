package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/pkg/remediate"
)

func handleAuditStart(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Site           string `json:"site"`
			LinkType       string `json:"link_type"`
			Provider       string `json:"provider"`
			Duration       int    `json:"duration"`
			Target         string `json:"target"`
			Direction      string `json:"direction"`
			Flows          int    `json:"flows"`
			SmallURL       string `json:"small_url"`
			BulkAddr       string `json:"bulk_addr"`
			Iface          string `json:"iface"`
			RouterOSIface  string `json:"routeros_iface"`
			RouterOSLan    string `json:"routeros_lan"`
			Qdisc          string `json:"qdisc"`
			Asymmetric     *bool  `json:"asymmetric"`
			CPUConstrained bool   `json:"cpu_constrained"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if body.Duration <= 0 {
			body.Duration = 30
		}
		// prévention — bornes imposées; un audit débridé coûte cher
		if body.Duration < 10 {
			body.Duration = 10
		}
		if body.Duration > 600 {
			body.Duration = 600
		}
		if len(body.Site) > 80 {
			body.Site = truncateUTF8(body.Site, 80)
		}
		// une cible en blanc ferait échouer le ping sans dire pourquoi :
		// elle retombe sur le banc détecté, comme une cible absente.
		body.Target = strings.TrimSpace(body.Target)
		if len(body.Target) > 64 {
			writeErr(w, r, "cible trop longue", http.StatusBadRequest)
			return
		}
		// puits décidés par le banc détecté (M4) — cible, petit objet et
		// charge par défaut viennent du banc vivant, pas d'un 10.200.0.1
		// en dur. En external, le petit objet reste VIDE : jamais de charge
		// implicite (un A+ sans charge est un mensonge). L'opérateur garde
		// la main : champs explicites > env (CGO_SMALL_URL/CGO_BULK_ADDR) > banc.
		sinks := audit.Sinks{Mode: "external", Target: "8.8.8.8"}
		if d.SinksFn != nil {
			sinks = d.SinksFn()
		} else {
			sinks = audit.ResolveSinks()
		}
		if body.Target == "" {
			body.Target = sinks.Target
		}
		// Sens et flux : défauts up/4, bornes imposées — un « both » débridé
		// double la charge, un « flows » hors 1–8 sort du régime documenté.
		switch body.Direction {
		case "", "up":
			body.Direction = "up"
		case "down", "both":
		default:
			writeErr(w, r, "direction invalide (up|down|both)", http.StatusBadRequest)
			return
		}
		if body.Flows <= 0 {
			body.Flows = 4
		}
		if body.Flows > 8 {
			writeErr(w, r, "flows hors bornes (1–8)", http.StatusBadRequest)
			return
		}
		if body.SmallURL == "" {
			// external laisse le petit objet vide (pas de charge implicite) ;
			// l'env garde la main sur les liaisons de terrain.
			body.SmallURL = envDefault("CGO_SMALL_URL", sinks.SmallURL)
		}
		if body.BulkAddr == "" {
			body.BulkAddr = audit.EnvOrBulk(sinks)
		}
		if body.Iface == "" {
			body.Iface = os.Getenv("CGO_AUDIT_IFACE")
		}
		if body.RouterOSIface == "" {
			body.RouterOSIface = os.Getenv("CGO_AUDIT_ROUTEROS_IFACE")
		}
		auditMu.Lock()
		if auditRunning {
			auditMu.Unlock()
			writeErr(w, r, "audit already running", http.StatusConflict)
			return
		}
		// contexte détaché de la requête HTTP mais annulable par l'opérateur
		runCtx, cancel := context.WithCancel(context.Background())
		auditRunning = true
		auditCancel = cancel
		auditStartedAt = time.Now()
		auditTotalSec = body.Duration
		auditProgress.Store(audit.Progress{Phase: "repos"})
		// Le rapport en mémoire appartient à l'audit PRÉCÉDENT : le garder
		// pendant ce run ferait lire un ancien rapport comme le résultat de
		// celui-ci si le nouveau échoue. L'archive disque reste disponible.
		lastAuditReport = nil
		auditMu.Unlock()
		asymmetric := true
		if body.Asymmetric != nil {
			asymmetric = *body.Asymmetric
		}
		go func() {
			defer func() {
				// Une panne de dépendance ne doit pas tuer le dashboard :
				// net/http ne récupère pas les goroutines détachées. On
				// journalise la pile, on efface le rapport de ce run et on
				// rend la main (auditRunning=false) pour le suivant.
				if rec := recover(); rec != nil {
					log.Printf("audit: panne interne : %v\n%s", rec, debug.Stack())
					recordEvent("audit", fmt.Sprintf("panne interne — %s (%s) : %v", body.Site, body.LinkType, rec))
					auditMu.Lock()
					lastAuditReport = nil
					auditMu.Unlock()
				}
				cancel()
				auditMu.Lock()
				auditRunning = false
				auditCancel = nil
				auditMu.Unlock()
			}()
			p := auditParamsFromRequest(map[string]any{
				"site": body.Site, "link_type": body.LinkType, "provider": body.Provider,
				"duration": float64(body.Duration), "target": body.Target,
				"direction": body.Direction, "flows": float64(body.Flows),
				"small_url": body.SmallURL, "bulk_addr": body.BulkAddr,
			})
			deps := audit.Deps{}
			if d.AuditDepsFn != nil {
				deps = d.AuditDepsFn()
			}
			deps.OnProgress = func(pr audit.Progress) { auditProgress.Store(pr) }
			res, err := audit.Run(runCtx, p, deps)
			// Annulation opérateur : Run rend context.Canceled sans résultat
			// partiel. Le rapport précédent a été effacé au démarrage : un
			// arrêt ne peut donc pas être maquillé en audit terminé. Un audit
			// achevé juste avant l'annulation garde son résultat (err == nil).
			if errors.Is(err, context.Canceled) {
				recordEvent("audit", fmt.Sprintf("annulé par l'opérateur — %s (%s)", p.Site, p.LinkType))
				return
			}
			if err != nil {
				recordEvent("audit", fmt.Sprintf("échec — %s (%s) : %s", p.Site, p.LinkType, err))
				return
			}
			// Evidence first: the report is sealed from the row this audit
			// appended, so the artifact and the CSV cannot drift apart.
			if err := audit.AppendLinkAudit("data", res); err != nil {
				recordEvent("audit", fmt.Sprintf("échec écriture CSV — %s : %s", p.Site, err))
			}
			link := remediate.Link{
				Interface:            body.Iface,
				RouterOSInterface:    body.RouterOSIface,
				RouterOSLanInterface: body.RouterOSLan,
			}
			rep, rerr := audit.BuildReport(res, link, body.Qdisc, asymmetric, body.CPUConstrained)
			if rerr != nil {
				recordEvent("audit", fmt.Sprintf("rapport impossible — %s : %s", p.Site, rerr))
			} else {
				if err := rep.SealSource(filepath.Join("data", "link_audit.csv")); err != nil {
					recordEvent("audit", fmt.Sprintf("scellement impossible — %s : %s", p.Site, err))
				}
				if _, _, err := audit.WriteArtifacts(rep, "data"); err != nil {
					recordEvent("audit", fmt.Sprintf("artefacts non écrits — %s : %s", p.Site, err))
				}
			}
			auditMu.Lock()
			lastAudit = res
			lastAuditReport = rep
			auditMu.Unlock()
			grade := res.BloatGrade
			if grade == "" {
				grade = "non mesuré"
			}
			recordEvent("audit", fmt.Sprintf("terminé — %s (%s), p95 %g ms, bufferbloat %s", p.Site, p.LinkType, res.RTTIdleP95, grade))
		}()
		// la ligne d'ouverture nomme le banc détecté (M4) : l'opérateur voit
		// le chemin qui va mesurer AVANT de lire un verdict.
		smallNote, bulkNote := body.SmallURL, body.BulkAddr
		if smallNote == "" {
			smallNote = "—"
		}
		if bulkNote == "" {
			bulkNote = "—"
		}
		recordEvent("audit", fmt.Sprintf("démarré — %s (%s) %ds, banc %s, cible %s, charge %s, small %s",
			body.Site, body.LinkType, body.Duration, sinks.Mode, body.Target, bulkNote, smallNote))
		writeJSON(w, map[string]any{"started": true})
	}
}

func handleAuditStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		defer auditMu.Unlock()
		writeJSON(w, map[string]any{"running": auditRunning, "last": lastAudit, "report": lastAuditReport})
	}
}

// Audit vivant — la carte « Observer » sonde ceci toutes les 2 s. Pas
// d'audit ⇒ running:false seul : aucun zéro ne se lit comme mesuré.
func handleAuditLive() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		running, started, total := auditRunning, auditStartedAt, auditTotalSec
		auditMu.Unlock()
		if !running {
			writeJSON(w, map[string]any{"running": false})
			return
		}
		pr, _ := auditProgress.Load().(audit.Progress)
		rttSamples, smallSamples := pr.SamplesRTT, pr.SamplesSmall
		if rttSamples == nil {
			rttSamples = []float64{}
		}
		if smallSamples == nil {
			smallSamples = []float64{}
		}
		writeJSON(w, map[string]any{
			"running":          true,
			"phase":            pr.Phase,
			"elapsed_sec":      int(time.Since(started).Seconds()),
			"total_sec":        total,
			"phase_remain_sec": pr.PhaseRemainSec,
			"last_rtt_ms":      pr.LastRTTms,
			"last_small_ms":    pr.LastSmallMs,
			"samples":          map[string]any{"rtt_ms": rttSamples, "small_ms": smallSamples},
			"timeline":         pr.Timeline,
		})
	}
}

// Arrêt opérateur — la goroutine d'audit voit son contexte annulé,
// n'écrit rien de partiel et journalise « annulé par l'opérateur ».
func handleAuditStop() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		cancel, running := auditCancel, auditRunning
		if !running || cancel == nil {
			auditMu.Unlock()
			writeJSON(w, map[string]any{"stopped": false})
			return
		}
		// Une seule livraison : la goroutine efface encore ces champs à sa
		// sortie, sans effet. Sans ce nil, deux arrêts rapprochés répondaient
		// « stopped » deux fois alors qu'un seul avait annulé quelque chose.
		auditCancel = nil
		auditMu.Unlock()
		cancel()
		writeJSON(w, map[string]any{"stopped": true})
	}
}

// Rapport d'audit — artefact persistant, pas seulement l'état en mémoire.
// Sans audit_id, le dernier rapport ; avec, la copie archivée.
func handleAuditReport() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.URL.Query().Get("audit_id"))
		var rep *audit.Report
		auditMu.Lock()
		if id == "" || (lastAuditReport != nil && lastAuditReport.AuditID == id) {
			rep = lastAuditReport
		}
		auditMu.Unlock()
		if rep == nil {
			var err error
			rep, err = readAuditReport(id)
			if err != nil {
				writeErr(w, r, "rapport introuvable : "+id, http.StatusNotFound)
				return
			}
		}
		if r.URL.Query().Get("format") == "md" || r.URL.Query().Get("format") == "markdown" {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			_, _ = io.WriteString(w, rep.RenderMarkdown())
			return
		}
		writeJSON(w, rep)
	}
}

// Remédiation — le script est déjà dans le rapport scellé ; l'endpoint
// existe pour que le dashboard copie sans parser le Markdown. Un iface
// fourni ici ne réécrit pas les mesures : BuildReport repart du Result
// mesuré et ne fait que nommer l'équipement.
func handleRemediate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := strings.TrimSpace(q.Get("audit_id"))
		var rep *audit.Report
		auditMu.Lock()
		if id == "" || (lastAuditReport != nil && lastAuditReport.AuditID == id) {
			rep = lastAuditReport
		}
		auditMu.Unlock()
		if rep == nil {
			var err error
			rep, err = readAuditReport(id)
			if err != nil {
				writeErr(w, r, "aucun rapport d'audit — lancez cgo audit", http.StatusNotFound)
				return
			}
		}
		// Rebuild only when the operator names an interface and the measured
		// result is still in memory; otherwise return the archived report's
		// remediation verbatim.
		if iface := strings.TrimSpace(q.Get("iface")); iface != "" {
			auditMu.Lock()
			res := lastAudit
			auditMu.Unlock()
			if res != nil && res.AuditID == rep.AuditID {
				link := remediate.Link{
					Interface:            iface,
					RouterOSInterface:    strings.TrimSpace(q.Get("routeros_iface")),
					RouterOSLanInterface: strings.TrimSpace(q.Get("routeros_lan")),
				}
				asym := q.Get("asymmetric") != "false"
				rebuilt, err := audit.BuildReport(res, link, q.Get("qdisc"), asym, q.Get("cpu_constrained") == "true")
				if err == nil {
					rep = rebuilt
				}
			}
		}
		writeJSON(w, map[string]any{"audit_id": rep.AuditID, "remediation": rep.Remediation, "unavailable": rep.Unavailable})
	}
}

// protocole RQ1 matérialisé : l'opérateur choisit au lieu de recopier.
func handleAuditPresets() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"presets": audit.Presets()})
	}
}

func handleAuditList() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// read link_audit.csv
		f, err := os.Open("data/link_audit.csv")
		if err != nil {
			writeJSON(w, map[string]any{"audits": []any{}})
			return
		}
		defer f.Close()
		rd := csv.NewReader(f)
		rows, _ := rd.ReadAll()
		if len(rows) <= 1 {
			writeJSON(w, map[string]any{"audits": []any{}})
			return
		}
		var out []map[string]string
		hdr := rows[0]
		for _, row := range rows[1:] {
			m := map[string]string{}
			for i, h := range hdr {
				if i < len(row) {
					m[h] = row[i]
				}
			}
			m["has_report"] = strconv.FormatBool(auditReportExists(m["audit_id"]))
			out = append(out, m)
		}
		writeJSON(w, map[string]any{"audits": out})
	}
}

// boucle RQ1→RQ2 : l'audit réel devient profil rejouable sur le banc.
// capacity = throughput, delay = rtt_idle_p95, loss = loss_pct mesurés.
// ?audit_id= choisit la ligne (défaut : dernière) — l'opérateur importe
// la fibre sans relancer la pointe.
func handleAuditToProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("data/link_audit.csv")
		if err != nil {
			writeErr(w, r, "aucun audit — lancez d'abord cgo audit", http.StatusNotFound)
			return
		}
		defer f.Close()
		rd := csv.NewReader(f)
		rows, _ := rd.ReadAll()
		if len(rows) < 2 {
			writeErr(w, r, "audit vide", http.StatusNotFound)
			return
		}
		hdr := rows[0]
		last := rows[len(rows)-1]
		// ?audit_id= : la ligne choisie, sinon la dernière
		if want := r.URL.Query().Get("audit_id"); want != "" {
			found := false
			for _, row := range rows[1:] {
				for i, h := range hdr {
					if h == "audit_id" && i < len(row) && row[i] == want {
						last, found = row, true
					}
				}
				if found {
					break
				}
			}
			if !found {
				writeErr(w, r, "audit introuvable : "+want, http.StatusNotFound)
				return
			}
		}
		get := func(col string) string {
			for i, h := range hdr {
				if h == col && i < len(last) {
					return last[i]
				}
			}
			return ""
		}
		cap, _ := strconv.ParseFloat(get("throughput_mbps"), 64)
		delay, _ := strconv.ParseFloat(get("rtt_idle_p95_ms"), 64)
		loss, _ := strconv.ParseFloat(get("loss_pct"), 64)
		id := "P-audit"
		if v := get("audit_id"); v != "" {
			id = "P-" + v
		}
		p := model.Profile{ID: id, CapacityMbps: cap, DelayMs: delay, LossPct: loss}
		// Un profil importé est rejoué comme un lien réel : une capacité ou un
		// délai non mesuré ne peut pas devenir 20 Mbit/s / 100 ms. L'ancien
		// repli fabriquait un lien de laboratoire et l'appelait « audit ».
		if p.CapacityMbps <= 0 || p.DelayMs <= 0 {
			writeErr(w, r, "audit sans capacité ou RTT idle mesurés — profil non importable (relancez l'audit avec charge bulk et cible joignable)", http.StatusUnprocessableEntity)
			return
		}
		if err := profile.Import(p); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		recordEvent("profil", fmt.Sprintf("%s importé depuis audit — %g Mbit/s, %g ms, perte %g %%", id, p.CapacityMbps, p.DelayMs, p.LossPct))
		writeJSON(w, map[string]any{"ok": true, "profile": p})
	}
}
