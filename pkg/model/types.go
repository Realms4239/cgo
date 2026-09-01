// Package model — types du domaine.
// L'ordre des colonnes CSV est défini par csvFieldOrder.
package model

// Profil — valeurs par défaut, remplacées par les profils importés.
type Profile struct {
	ID           string  `json:"id"`
	CapacityMbps float64 `json:"capacity_mbps"`
	DelayMs      float64 `json:"delay_ms"`
	JitterMs     float64 `json:"jitter_ms"`
	LossPct      float64 `json:"loss_pct"`
}

var Profiles = map[string]Profile{
	"P1": {ID: "P1", CapacityMbps: 80, DelayMs: 20, JitterMs: 2, LossPct: 0},
	"P2": {ID: "P2", CapacityMbps: 20, DelayMs: 100, JitterMs: 15, LossPct: 0.5},
	"P3": {ID: "P3", CapacityMbps: 5, DelayMs: 600, JitterMs: 30, LossPct: 1},
	// P4 — Starlink LEO (recherche 2026 : 226 000 Ar/mois à Madagascar) :
	// débit élevé, délai bas par rapport au VSAT géostationnaire, mais gigue
	// marquée (handovers satellites) et perte ponctuelle — la classe LEO.
	"P4": {ID: "P4", CapacityMbps: 100, DelayMs: 40, JitterMs: 20, LossPct: 0.3},
}

type Qdisc string

const (
	PfifoFast Qdisc = "pfifo_fast"
	FqCodel   Qdisc = "fq_codel"
	Cake      Qdisc = "cake"
)

var AllQdiscs = []Qdisc{PfifoFast, FqCodel, Cake}

type CC string

const (
	Cubic CC = "cubic"
	BBR   CC = "bbr"
)

var AllCC = []CC{Cubic, BBR}

// Gate G0–G7 per SPEC §2.2.
type Gate int

const (
	G0TargetReachable Gate = iota
	G1BulkStarted
	G2ProbesProducing
	G3LatencyPlausible
	G4ThroughputCoherent
	G5NoDuplicateRows
	G6BaselineStable
	G7CPUNotSaturated
)

var GateCount = 8 // number of gates

// gate_status values (aqm_eval.csv).
const (
	GatePass     = "valid"
	GateDegraded = "degraded"
	GateInvalid  = "invalid"
)

// Phases d'un événement et durées (secondes).
const (
	PhaseBaseline = "baseline"
	PhaseCharge   = "charge"
	PhaseRecup    = "recup"

	BaselineSec = 30
	ChargeSec   = 120
	RecupSec    = 30
)

// Event is one matrix cell execution.
type Event struct {
	RunID      string  `json:"run_id"`
	EventID    int     `json:"event_id"`
	Profile    string  `json:"profile"`
	Qdisc      Qdisc   `json:"qdisc"`
	CC         CC      `json:"cc"`
	Repetition int     `json:"repetition"`

	RTTp50Ms       float64 `csv:"rtt_p50_ms"        json:"rtt_p50_ms"`
	RTTp95Ms       float64 `csv:"rtt_p95_ms"        json:"rtt_p95_ms"`
	QDIPctMs       float64 `csv:"qdi_ms"            json:"qdi_ms"`
	VoIPR          float64 `csv:"voip_r"            json:"voip_r"`
	Smallp95Ms     float64 `csv:"small_p95_ms"      json:"small_p95_ms"`
	DeadlineOKPct  float64 `csv:"deadline_ok_pct"   json:"deadline_ok_pct"`
	BulkGoodputMbps float64 `csv:"bulk_goodput_mbps" json:"bulk_goodput_mbps"`
	Drops          uint64  `csv:"drops"             json:"drops"`
	Retransmissions uint64 `csv:"retransmissions"   json:"retransmissions"`
	WastedBytes    uint64  `csv:"wasted_bytes"      json:"wasted_bytes"`
	CostARPerH     float64 `csv:"cost_ar_per_h"     json:"cost_ar_per_h"`
	CPUPct         float64 `csv:"cpu_pct"           json:"cpu_pct"`
	GateStatus     string  `csv:"gate_status"       json:"gate_status"`
}

// En-têtes aqm_eval.csv (après les colonnes d'identité).
var AQMEvalHeader = []string{
	"run_id", "event_id", "profile", "qdisc", "cc", "repetition",
	"rtt_p50_ms", "rtt_p95_ms", "qdi_ms", "voip_r", "small_p95_ms", "deadline_ok_pct",
	"bulk_goodput_mbps", "drops", "retransmissions", "wasted_bytes",
	"cost_ar_per_h", "cpu_pct", "gate_status",
}

// En-têtes link_audit.csv.
var LinkAuditHeader = []string{
	"audit_id", "timestamp", "site", "link_type", "provider",
	"rtt_idle_p50_ms", "rtt_idle_p95_ms", "rtt_loaded_p50_ms", "rtt_loaded_p95_ms",
	"throughput_mbps", "loss_pct", "http_small_p95_ms", "data_used_mb", "notes",
}
