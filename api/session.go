package api

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Realms4239/cgo/audit"
)

// État mutable du serveur — chaque pièce a UN propriétaire.
//
// Avant, l'état d'audit, le journal opérateur et le façonnage vivaient
// dans des variables de paquet partagées entre server.go et
// server_audit.go : « qui a le verrou ? » était une question ouverte, et
// les tests devait écrire dans ces variables pour preparer un état.
// Ici, un serveur possède son état ; les tests passent par la même seam.

// ── Journal opérateur ──────────────────────────────────────────────

// eventRing — les 50 derniers évènements, un mutex, un seul anneau.
type eventRing struct {
	mu    sync.Mutex
	items []map[string]any
}

func (r *eventRing) record(kind, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, map[string]any{"ts": time.Now().Format(time.RFC3339), "kind": kind, "msg": msg})
	if len(r.items) > 50 {
		r.items = r.items[len(r.items)-50:]
	}
}

// snapshot — une copie : le lecteur ne voit jamais l'anneau en mutation.
func (r *eventRing) snapshot() []map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]map[string]any, len(r.items))
	copy(out, r.items)
	return out
}

// journal — l'unique anneau du processus. L'hôte y écrit par RecordEvent
// (ses rappels de quarantaine, de skip, d'alerte) ; le serveur le lit.
var journal = &eventRing{}

// RecordEvent — le seul point d'entrée public du journal : l'hôte y
// écrit, le cockpit lit. L'interface reste celle d'avant.
func RecordEvent(kind, msg string) { journal.record(kind, msg) }

// recordEvent — le même anneau, vu de l'intérieur du module.
func recordEvent(kind, msg string) { journal.record(kind, msg) }

// ── Session d'audit ────────────────────────────────────────────────

// auditSession — l'état d'un audit : en cours, rapport, progression.
// Un mutex pour tout l'état : plus de « quel verrou protège ce champ ? ».
type auditSession struct {
	mu        sync.Mutex
	last      *audit.Result
	report    *audit.Report
	running   bool
	cancel    context.CancelFunc
	startedAt time.Time
	totalSec  int
	progress  atomic.Value // audit.Progress : écrit par l'audit, lu sans verrou
}

// begin — réserve la session pour un nouvel audit. false ⇒ un audit
// tourne déjà : l'appelant répond 409 et ne touche à rien.
func (s *auditSession) begin(cancel context.CancelFunc, totalSec int, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	s.cancel = cancel
	s.startedAt = now
	s.totalSec = totalSec
	s.progress.Store(audit.Progress{Phase: "repos"})
	// Le rapport en mémoire appartient à l'audit PRÉCÉDENT : le garder
	// pendant ce run ferait lire un ancien rapport comme le résultat de
	// celui-ci si le nouveau échoue. L'archive disque reste disponible.
	s.report = nil
	return true
}

// finish — l'audit est achevé (ou a échoué) : la session se libère.
func (s *auditSession) finish() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.cancel = nil
}

// clearReport — un rapport ne doit jamais outlire son audit.
func (s *auditSession) clearReport() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.report = nil
}

// setResult — le couple résultat/rapport d'un audit achevé.
func (s *auditSession) setResult(res *audit.Result, rep *audit.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last = res
	s.report = rep
}

// status — l'état lu par /api/audit/status.
func (s *auditSession) status() (bool, *audit.Result, *audit.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running, s.last, s.report
}

// reportFor — le rapport en mémoire quand aucun identifiant n'est demandé,
// ou quand l'identifiant demandé est celui de ce rapport. nil sinon.
func (s *auditSession) reportFor(id string) *audit.Report {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" || (s.report != nil && s.report.AuditID == id) {
		return s.report
	}
	return nil
}

// reportByID — le rapport en mémoire, seulement si son identifiant
// correspond. Vide ≠ « le dernier » : ici l'identifiant est exigé.
func (s *auditSession) reportByID(id string) *audit.Report {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" && s.report != nil && s.report.AuditID == id {
		return s.report
	}
	return nil
}

// lastResult — le résultat mesuré du dernier audit (pour la remédiation).
func (s *auditSession) lastResult() *audit.Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

// live — l'état lu par /api/audit/live (2 s).
func (s *auditSession) live() (bool, time.Time, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running, s.startedAt, s.totalSec
}

// stop — l'annulation opérateur. Une seule livraison : le cancel est
// retiré avant l'appel, un second arrêt ne peut donc pas re-annuler.
func (s *auditSession) stop() (context.CancelFunc, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.cancel == nil {
		return nil, false
	}
	cancel := s.cancel
	s.cancel = nil
	return cancel, true
}

func (s *auditSession) setProgress(p audit.Progress) { s.progress.Store(p) }

func (s *auditSession) progressValue() audit.Progress {
	pr, _ := s.progress.Load().(audit.Progress)
	return pr
}

// ── Façonnage du bord ──────────────────────────────────────────────

// shapeState — le dernier façonnage appliqué (observabilité du contrôle).
type shapeState struct {
	mu    sync.Mutex
	qdisc string
	cap   float64
	since time.Time
}

func (s *shapeState) set(q string, cap float64, now time.Time) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.qdisc, s.cap, s.since = q, cap, now
	return s.since.Format(time.RFC3339)
}

func (s *shapeState) get() (string, float64, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.qdisc, s.cap, s.since
}

// ── Cache de comparaison ───────────────────────────────────────────

// compareCache — le bootstrap/MW coûte ~ms par cellule : précalcul +
// cache, invalidé par le hash de provenance (tout nouveau gel change la clé).
type compareCache struct {
	sync.Mutex
	key string
	val map[string]any
}
