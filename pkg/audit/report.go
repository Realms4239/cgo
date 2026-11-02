package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/remediate"
)

// Report is the operator-facing artifact for one audit.
//
// Every field is either measured or explicitly marked unavailable. There is no
// field that gets a plausible default: a report that invents a deadline
// compliance rate is worse than one that says "not measured", because the
// invented number travels into a thesis table and a procurement decision.
type Report struct {
	AuditID   string `json:"audit_id"`
	Timestamp string `json:"timestamp"`
	Site      string `json:"site"`
	LinkType  string `json:"link_type"`
	Provider  string `json:"provider"`

	// Measured holds the values this audit actually observed.
	Measured Measured `json:"measured"`
	// Phases records the executed windows: baseline, load, drain.
	Phases Phases `json:"phases"`
	// Unavailable names what could not be measured, with the reason. This is
	// the honest half of the report and is never empty in practice.
	Unavailable []Unavailable `json:"unavailable"`

	// Grade is empty when bufferbloat was not measured.
	Grade   string `json:"bloat_grade"`
	Verdict string `json:"bloat_verdict"`

	// OperatorVerdict — le mot que lit un opérateur qui ne connaît pas la
	// méthode. Dérivé des faits déjà gelés, jamais d'une moyenne inventée.
	OperatorVerdict metrics.Verdict `json:"operator_verdict"`

	// Provenance seals the raw telemetry this report was built from.
	Provenance Provenance `json:"provenance"`

	// Invocation — les réglages effectifs et la commande CLI équivalente. Absent
	// pour un rapport scellé avant ce champ : rendu avec des zéros, un
	// « flows: 0 » se lirait comme « un flux demandé ». La console affiche alors
	// une absence, jamais une reproduction inventée.
	Invocation *Invocation `json:"invocation,omitempty"`

	// Artifact — où vivre la preuve brute de ce rapport, en chemin relatif au
	// répertoire de travail. Scellé par le moteur pour que la console n'ait pas à
	// reconstruire un chemin depuis un identifiant : une console qui devine un
	// dossier montre un jour le mauvais.
	Artifact string `json:"artifact,omitempty"`

	// Remediation is present only when capacity and RTT were both measured.
	Remediation *Remediation `json:"remediation,omitempty"`

	// Economics is the enterprise valuation, present only when a load was
	// measured AND a retransmission ratio exists (measured or estimated).
	// All contractual scenarios are shown with the primary marked by link
	// type — a scenario match, never a claim about the site's real contract.
	Economics *metrics.Economics `json:"economics,omitempty"`

	// CAPSLA is the 220 ms operator-objective assessment of the loaded 16 KB
	// completions (Model C). Measured=false means unavailable, never a pass.
	CAPSLA metrics.CAPSLA `json:"cap_sla"`

	// RouterOSValidation is the CHR proof for this report's remediation
	// script. Absent means "not validated", never "validated by default".
	RouterOSValidation *RouterOSValidation `json:"routeros_validation,omitempty"`

	// KernelTelemetry is the eBPF or TCP_INFO tier for this audit. Absent
	// means the build did not attempt it — never a silent zero.
	KernelTelemetry *KernelTelemetry `json:"kernel_telemetry,omitempty"`
}

// RouterOSValidation records what a CHR run proved about this report's
// RouterOS script: the device version, the evidence level, the export hash
// and the field-level checks. It never claims the configuration was applied.
type RouterOSValidation struct {
	Version      string          `json:"version"`
	Level        string          `json:"level"`
	ExportSHA256 string          `json:"export_sha256"`
	ProofPath    string          `json:"proof_path,omitempty"`
	Checks       []RouterOSCheck `json:"checks"`
}

// KernelTelemetry is what the kernel tier could say for this audit, and why,
// when it could not say more. Source is ebpf|tcpinfo|unavailable; a reason
// always accompanies a fallback.
type KernelTelemetry struct {
	Available           bool         `json:"available"`
	Source              string       `json:"source,omitempty"`
	Reason              string       `json:"reason,omitempty"`
	ObjectSHA256        string       `json:"object_sha256,omitempty"`
	RetransPayloadBytes uint64       `json:"retrans_payload_bytes,omitempty"`
	QdiscDropPackets    uint64       `json:"qdisc_drop_packets,omitempty"`
	FlowsSeen           int          `json:"flows_seen,omitempty"`
	FlowsTotal          int          `json:"flows_total,omitempty"`
	Flows               []KernelFlow `json:"flows,omitempty"`
}

// KernelFlow is one sender-side flow as the kernel tier saw it. Present only
// when the eBPF object ran; TCP_INFO has no per-flow rows.
type KernelFlow struct {
	Local               string `json:"local"`
	Peer                string `json:"peer"`
	RetransPayloadBytes uint64 `json:"retrans_payload_bytes"`
	RetransEvents       uint64 `json:"retrans_events"`
	QdiscDropPackets    uint64 `json:"qdisc_drop_packets"`
	SRTTus              uint32 `json:"srtt_us"`
}

// Phases records the audit schedule actually executed. A 60 s audit is
// 10 s baseline, 40 s load and 10 s drain; the numbers prove the loaded
// window was long enough to fill a queue rather than a token 8 s burst.
type Phases struct {
	BaselineSeconds float64 `json:"baseline_seconds"`
	LoadSeconds     float64 `json:"load_seconds"`
	RecoverySeconds float64 `json:"recovery_seconds"`
}

// Measured is the set of numbers the audit actually produced.
type Measured struct {
	RTTIdleP50Ms   float64 `json:"rtt_idle_p50_ms"`
	RTTIdleP95Ms   float64 `json:"rtt_idle_p95_ms"`
	RTTLoadedP50Ms float64 `json:"rtt_loaded_p50_ms"`
	RTTLoadedP95Ms float64 `json:"rtt_loaded_p95_ms"`
	// QDIMs is the queue delay impairment: loaded p95 minus loaded p50. It
	// isolates standing-queue delay from propagation delay.
	QDIMs float64 `json:"qdi_ms"`
	// BloatDeltaMs is loaded p95 minus idle p95, the Waveform-style measure.
	// Only meaningful when LoadMeasured is true.
	BloatDeltaMs float64 `json:"bloat_delta_ms"`
	// DrainMs is how far above the rest threshold the p95 still sat after the
	// load stopped. 0 with DrainMeasured=false means "not measured".
	DrainMs       float64 `json:"drain_ms"`
	DrainMeasured bool    `json:"drain_measured"`
	// LoadAttempted — la fenêtre de charge était-elle prévue ? Sans ce fait,
	// un audit sans puits se lirait « jamais testé ».
	LoadAttempted bool `json:"load_attempted"`
	// LoadMeasured is false when no bulk transfer ran, in which case
	// BloatDeltaMs and Grade carry no meaning.
	LoadMeasured bool `json:"load_measured"`
	// ThroughputMbps is measured goodput of the bulk flow, 0 when unmeasured.
	ThroughputMbps float64 `json:"throughput_mbps"`
	LossPct        float64 `json:"loss_pct"`
	// HTTPSmallP95Ms is the 95th percentile of small-object completion, 0 when
	// no small-object URL was supplied.
	HTTPSmallP95Ms float64 `json:"http_small_p95_ms"`
	SmallMeasured  bool    `json:"small_measured"`
	// VoIPR is the ITU-T G.107 R-score derived from the measured conditions.
	VoIPR float64 `json:"voip_r"`
	// DataUsedMB is the volume the audit itself consumed.
	DataUsedMB float64 `json:"data_used_mb"`
	// Direction — sens de la fenêtre headline ("up", "down", "both") et
	// flux de charge exercés ; les miroirs ci-dessous portent la fenêtre
	// descendante quand elle a tourné.
	Direction string  `json:"direction"`
	FlowCount int     `json:"flow_count"`
	JFIPct    float64 `json:"jfi_pct"`
	// JFIMeasured — faux quand l'index de Jain n'avait pas de contributions par
	// flux à comparer. Sans ce fait, jfi_pct = 0 se lit « 0 % d'équité ».
	JFIMeasured bool `json:"jfi_measured"`
	// Miroirs descendants — zéro tant que la fenêtre down n'a pas tourné.
	RTTLoadedDownP95Ms float64 `json:"rtt_loaded_down_p95_ms"`
	ThroughputDownMbps float64 `json:"throughput_down_mbps"`
	BloatDeltaDownMs   float64 `json:"bloat_delta_down_ms"`
	JFIDownPct         float64 `json:"jfi_down_pct"`
	// LoadMeasuredDown — la fenêtre descendante a mesuré une charge, même
	// quand la fenêtre headline (up) est muette ; ses miroirs restent rendus.
	LoadMeasuredDown bool `json:"load_measured_down"`
	// GradeDown — note bufferbloat de la fenêtre descendante, vide tant
	// qu'elle n'a pas tourné.
	GradeDown string `json:"grade_down"`
}

// Unavailable records a metric the audit could not produce, and why.
type Unavailable struct {
	Metric string `json:"metric"`
	Reason string `json:"reason"`
}

// Provenance seals the evidence chain — TWO seals: the raw telemetry
// timeseries this report's percentiles were computed from, and the appended
// link_audit.csv row file. One artifact, two independently checkable links.
type Provenance struct {
	// SourceFile is the CSV the report was rendered from.
	SourceFile string `json:"source_file"`
	// SourceSHA256 is the SHA-256 of that file's bytes at render time.
	SourceSHA256 string `json:"source_sha256"`
	// RowCount is the number of data rows in the source file.
	RowCount int `json:"row_count"`
	// TelemetryFile / TelemetrySHA256 seal the raw per-probe timeseries.
	TelemetryFile   string `json:"telemetry_file,omitempty"`
	TelemetrySHA256 string `json:"telemetry_sha256,omitempty"`
	// RowFileSHA256 is the row-file seal under its explicit name (same bytes
	// as SourceSHA256; kept so both links are named, not implied).
	RowFileSHA256 string `json:"row_file_sha256,omitempty"`
}

// Remediation is the production configuration derived from the measurement.
type Remediation struct {
	Qdisc  string `json:"qdisc"`
	Reason string `json:"reason"`
	// LinuxTC is the ordered command set, one command per element.
	LinuxTC []string `json:"linux_tc"`
	// RouterOS is the RouterOS v7+ script, empty when no RouterOS interface
	// was named.
	RouterOS string `json:"routeros,omitempty"`
	// LinuxTCIngress is the ordered download command set for a down (or both)
	// measurement. It stays an empty array for an upload-only report.
	LinuxTCIngress []string `json:"linux_tc_ingress,omitempty"`
	// RouterOSIngress is the LAN queue-tree script for the download direction.
	// It is emitted for any measured down window on a link that names a
	// RouterOS interface; an undeclared LAN interface is assumed and said so in
	// the script, and routeros_lan_iface/routeros_lan_assumed record that
	// assumption — only here, when this script exists. A report that carries no
	// download RouterOS script carries no labelled assumption about one.
	RouterOSIngress string `json:"routeros_ingress,omitempty"`

	// Rates and interfaces named by the measurement that produced them. The
	// UI reads these instead of parsing a command string — un taux recopié
	// depuis le texte d'un script dérive dès qu'on change la formule. Le nom
	// d'interface LAN et son drapeau « supposé » ne sont posés que quand le
	// script qui les justifie existe : pas de script, pas d'hypothèse étiquetée.
	UpMeasuredMbps     float64 `json:"up_measured_mbps,omitempty"`
	UpShapedMbps       float64 `json:"up_shaped_mbps,omitempty"`
	DownMeasuredMbps   float64 `json:"down_measured_mbps,omitempty"`
	DownShapedMbps     float64 `json:"down_shaped_mbps,omitempty"`
	LinuxIface         string  `json:"linux_iface,omitempty"`
	RouterOSWanIface   string  `json:"routeros_wan_iface,omitempty"`
	RouterOSLanIface   string  `json:"routeros_lan_iface,omitempty"`
	RouterOSLanAssumed bool    `json:"routeros_lan_assumed,omitempty"`
}

// BuildReport assembles a report from an audit result.
//
// link supplies the interface names for remediation. Its CapacityMbps and
// RTTms are overwritten from the measurement: the caller does not get to
// declare the link's capacity, the audit measures it.
func BuildReport(res *Result, link remediate.Link, qdisc string, asymmetric, cpuConstrained bool) (*Report, error) {
	if res == nil {
		return nil, fmt.Errorf("audit: nil result")
	}
	rep := &Report{
		AuditID:   res.AuditID,
		Timestamp: res.Timestamp,
		Site:      res.Site,
		LinkType:  res.LinkType,
		Provider:  res.Provider,
		Grade:     res.BloatGrade,
		Verdict:   res.BloatVerdict,
		Phases: Phases{
			BaselineSeconds: res.BaselineSeconds,
			LoadSeconds:     res.LoadSeconds,
			RecoverySeconds: res.RecoverySeconds,
		},
	}
	// -ros-lan ne prend que l'interface NOMMÉE : link.RouterOSLan() rend le
	// "bridge" supposé, et l'installer dans la commande écrirait un réglage que
	// personne n'a demandé. Une invocation sans commande (résultat scellé avant
	// le champ) reste absente du rapport, plutôt que zérotée.
	if in := res.Invocation.avecRemediation(qdisc, link.Interface,
		link.RouterOSInterface, link.RouterOSLanInterface, asymmetric, cpuConstrained); in.Command != "" {
		rep.Invocation = &in
	}
	// Un identifiant refusé par ArtifactID laisse le pointeur vide : le rapport
	// ne publie pas un chemin approximatif.
	rep.Artifact = ArtifactPointer(res.AuditID)

	loadMeasured := res.ThroughputMbps > 0 && res.BloatGrade != ""
	loadMeasuredDown := res.ThroughputDownMbps > 0
	rep.Measured = Measured{
		RTTIdleP50Ms:       res.RTTIdleP50,
		RTTIdleP95Ms:       res.RTTIdleP95,
		RTTLoadedP50Ms:     res.RTTLoadedP50,
		RTTLoadedP95Ms:     res.RTTLoadedP95,
		QDIMs:              res.RTTLoadedP95 - res.RTTLoadedP50,
		BloatDeltaMs:       res.BloatDeltaMs,
		DrainMs:            res.DrainMs,
		DrainMeasured:      res.DrainMeasured,
		LoadAttempted:      res.LoadAttempted,
		LoadMeasured:       loadMeasured,
		ThroughputMbps:     res.ThroughputMbps,
		LossPct:            res.LossPct,
		HTTPSmallP95Ms:     res.HTTPSmallP95,
		SmallMeasured:      res.HTTPSmallP95 > 0,
		DataUsedMB:         res.DataUsedMB,
		Direction:          res.Direction,
		FlowCount:          res.FlowCount,
		JFIPct:             res.JFIPct,
		JFIMeasured:        res.JFIMeasured,
		RTTLoadedDownP95Ms: res.RTTLoadedDownP95,
		ThroughputDownMbps: res.ThroughputDownMbps,
		BloatDeltaDownMs:   res.BloatDeltaDownMs,
		JFIDownPct:         res.JFIDownPct,
		LoadMeasuredDown:   loadMeasuredDown,
		GradeDown:          res.GradeDown,
	}

	// VoIP R from the loaded conditions: one-way delay is half the RTT, jitter
	// is the p95-p50 spread the jitter buffer has to absorb.
	if loadMeasured {
		rep.Measured.VoIPR = metrics.VoIPR(res.RTTLoadedP50/2, res.RTTLoadedP95-res.RTTLoadedP50, res.LossPct)
	}

	// Valorisation entreprise : TOUS les scénarios, primaire MARQUÉ pour le
	// type de lien déclaré. Le gaspillage vient de la mesure : octets
	// réémis (TCP_INFO, base "measured") ou part de perte sonde appliquée
	// au volume transporté (base "estimated", proxy documenté). Sans ratio,
	// aucune valorisation chiffrée — jamais un zéro qui se lirait « aucun
	// gaspillage ».
	if loadMeasured && res.RetransRatio.Basis != "" && res.DataUsedMB > 0 && res.LoadSeconds > 0 {
		transported := res.DataUsedMB * 1e6
		wasted := res.DataUsedMB * res.LossPct / 100 * 1e6
		if res.RetransRatio.Basis == "measured" {
			wasted = res.RetransRatio.Ratio * transported
		}
		rep.Economics = metrics.BuildEconomics(res.LinkType, res.RetransRatio, transported, wasted, res.LoadSeconds)
	}
	// Une valorisation absente doit être dite, jamais silencieuse : sans
	// entrée Unavailable, le lecteur ne distingue pas un oubli d'une mesure
	// impossible. La fenêtre headline reste la base du calcul (durée et
	// volume chargés) ; un « both » avec up muet garde ses miroirs down mais
	// n'extrapole pas une durée à moitié chargée.
	if rep.Economics == nil {
		reason := "économie entreprise — fenêtre de charge vide ; aucun scénario n'est chiffré plutôt qu'un zéro fabriqué"
		switch {
		case loadMeasured:
			reason = "ratio de réémission non mesuré ou volume de charge vide ; aucun scénario n'est chiffré plutôt qu'un zéro fabriqué"
		case loadMeasuredDown:
			reason = "économie entreprise — fenêtre headline (up) muette ; les miroirs descendants restent mesurés mais la valorisation n'extrapole pas une durée à moitié chargée"
		}
		rep.Unavailable = append(rep.Unavailable, Unavailable{Metric: "economics", Reason: reason})
	}
	// SLA CAP (modèle C) : l'objectif opérateur DGM 220 ms, cadre WIS 2.0 —
	// jamais présenté comme un seuil imposé par l'OMM.
	rep.CAPSLA = metrics.CAPSLA{
		Measured: res.CAPSLAMeasured, OKPct: res.CAPDeadlineOKPct, Breach: res.CAPSLABreach, Label: res.CAPSLALabel,
		TargetPct: metrics.CAPBreachThresholdPct, DeadlineMs: metrics.CAPDeadlineMs,
	}

	// Verdict opérateur : uniquement des faits déjà gelés dans ce rapport.
	// La note vient de rep.Grade (donc res.BloatGrade), jamais d'une chaîne
	// reparsée ni d'un défaut inventé ; LoadEstablished relit LoadMeasured,
	// dérivé plus haut, pour que le mot et le fait ne se contredisent pas.
	rep.OperatorVerdict = metrics.AssessVerdict(metrics.VerdictInput{
		LoadAttempted:   res.LoadAttempted,
		LoadEstablished: rep.Measured.LoadMeasured,
		Grade:           rep.Grade,
		SLA:             rep.CAPSLA,
	})

	// Say plainly what is missing. These are the metrics a jury or an operator
	// will ask for, so each absence is named rather than silently zeroed.
	if !loadMeasured {
		gradeReason := "no bulk load ran, so the loaded window was never under load; a grade here would be fabricated"
		throughputReason := "no bulk sink reachable; supply -bulk host:port or install iperf3"
		if loadMeasuredDown {
			gradeReason = "the headline (up) window was not measured; the down window's grade is reported as grade_down, never promoted to the headline grade"
			throughputReason = "the headline (up) window was not measured; the down window's throughput is reported under its own label"
		}
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "bufferbloat grade",
			Reason: gradeReason,
		})
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "throughput",
			Reason: throughputReason,
		})
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "voip_r",
			Reason: "requires a measured headline load window; not computed from an empty headline window",
		})
	}
	if !rep.Measured.SmallMeasured {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "http_small_p95",
			Reason: "no small-object URL supplied; ping RTT is not a substitute for object completion time",
		})
	}
	if !rep.Measured.DrainMeasured {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "drain_ms",
			Reason: "audit shorter than 20 s, so no recovery window was collected; drain is not the same as zero",
		})
	}
	// Deadline compliance of the CAMPAIGN's operator-defined objective is a
	// separate metric this link audit cannot compute; the audit's own CAP
	// assessment at the fixed 220 ms operator objective is cap_sla above.
	rep.Unavailable = append(rep.Unavailable, Unavailable{
		Metric: "deadline_ok_pct",
		Reason: "cap_deadline_ok_pct covers the fixed 220 ms operator objective; the campaign's operator-defined deadline_ok_pct stays a separate, campaign-only metric",
	})

	// Remediation only from a real measurement. This is the guard that keeps a
	// guessed bandwidth out of a production router config. A `down` headline
	// already lives in ThroughputMbps; a `both` result has separate mirrors.
	upCap, downCap := 0.0, 0.0
	switch res.Direction {
	case "down":
		downCap = res.ThroughputMbps
	case "both":
		upCap = res.ThroughputMbps
		downCap = res.ThroughputDownMbps
	default: // "", "up"
		upCap = res.ThroughputMbps
	}
	if res.RTTIdleP50 <= 0 || (upCap <= 0 && downCap <= 0) {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "remediation",
			Reason: "capacity or idle RTT unmeasured; a queue config sized from a guess would be deployed blindly",
		})
	} else {
		link.RTTms = res.RTTIdleP50
		rec := remediate.Recommend(link, asymmetric, cpuConstrained)
		useQdisc := qdisc
		if useQdisc == "" {
			useQdisc = rec.Qdisc
		}
		rem := &Remediation{Qdisc: useQdisc, Reason: rec.Reason, LinuxTC: []string{}}
		rem.LinuxIface = link.Interface
		rem.RouterOSWanIface = link.RouterOSInterface
		if upCap > 0 {
			rem.UpMeasuredMbps = upCap
			rem.UpShapedMbps = remediate.ShapedRate(upCap)
			link.CapacityMbps = upCap
			if cmds, err := remediate.LinuxTC(link, useQdisc); err == nil {
				rem.LinuxTC = cmds
			}
			if link.RouterOSInterface != "" {
				if script, err := remediate.RouterOS(link, useQdisc); err == nil {
					rem.RouterOS = script
				}
			}
		} else if res.Direction == "both" {
			rep.Unavailable = append(rep.Unavailable, Unavailable{
				Metric: "remediation (upload)",
				Reason: "upload throughput unmeasured; the download remediation is not promoted to an upload config",
			})
		}
		if downCap > 0 {
			rem.DownMeasuredMbps = downCap
			rem.DownShapedMbps = remediate.ShapedRate(downCap)
			link.CapacityMbps = downCap
			if cmds, err := remediate.LinuxTCIngress(link, useQdisc); err == nil {
				rem.LinuxTCIngress = cmds
			}
			// Le script descendant est émis dès que la fenêtre descendante est
			// mesurée : une interface LAN non déclarée est supposée (« bridge »)
			// et le script le dit. Un routeur MikroTik jamais nommé reste une
			// absence, jamais une hypothèse : un audit purement Linux ne reçoit
			// pas de script RouterOS.
			if link.RouterOSInterface != "" || link.RouterOSLanInterface != "" {
				if script, err := remediate.RouterOSIngress(link, useQdisc); err == nil {
					rem.RouterOSIngress = script
					// La paire d'interface supposée n'est déclarée que là où le
					// script existe : un rapport sans script descendant ne doit
					// rien affirmer sur une remédiation qu'il ne porte pas, faute
					// de quoi l'UI rendrait « interface supposée » sur une carte
					// sans routeur.
					lanIface, lanAssumed := link.RouterOSLan()
					rem.RouterOSLanIface, rem.RouterOSLanAssumed = lanIface, lanAssumed
				}
			}
		} else if res.Direction == "both" {
			rep.Unavailable = append(rep.Unavailable, Unavailable{
				Metric: "remediation (download)",
				Reason: "download throughput unmeasured; the upload remediation is not promoted to a download config",
			})
		}
		if res.Direction == "down" {
			rem.Reason = "download path: the shaper queues packets after they cross the link (Linux IFB ingress or RouterOS LAN queue tree); the upload ack-filter does not apply to a download data stream"
		}
		if len(rem.LinuxTC) > 0 || len(rem.LinuxTCIngress) > 0 || rem.RouterOS != "" || rem.RouterOSIngress != "" {
			rep.Remediation = rem
		} else {
			rep.Unavailable = append(rep.Unavailable, Unavailable{
				Metric: "remediation",
				Reason: "no interface was named, so no command set could be generated",
			})
		}
	}

	return rep, nil
}

// SealSource records the SHA-256 of the CSV the report was built from, so the
// report and its evidence can be checked against each other later.
func (r *Report) SealSource(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("audit: seal source %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	r.Provenance.SourceFile = path
	r.Provenance.SourceSHA256 = hex.EncodeToString(sum[:])
	r.Provenance.RowFileSHA256 = r.Provenance.SourceSHA256
	r.Provenance.RowCount = countCSVRows(data)
	return nil
}

// SealTelemetry scelle la série brute (telemetry.csv) : c'est la preuve
// première, avant toute agrégation en percentiles.
func (r *Report) SealTelemetry(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("audit: seal telemetry %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	r.Provenance.TelemetryFile = path
	r.Provenance.TelemetrySHA256 = hex.EncodeToString(sum[:])
	return nil
}

// countCSVRows counts data rows, excluding the header and blank lines.
func countCSVRows(data []byte) int {
	lines := strings.Split(strings.TrimRight(string(data), "\r\n"), "\n")
	n := 0
	for i, ln := range lines {
		if i == 0 || strings.TrimSpace(ln) == "" {
			continue
		}
		n++
	}
	return n
}

// RenderJSON returns the machine-readable report.
func (r *Report) RenderJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// RenderMarkdown returns the operator-facing report.
//
// codeSpan — délimiteur Markdown qui s'élargit si le contenu porte un accent
// grave. Le nom du site est une saisie libre : une commande qui casse son span
// se colle mal dans un ticket ou une annexe du mémoire.
func codeSpan(s string) string {
	tick := "`"
	for strings.Contains(s, tick) {
		tick += "`"
	}
	return tick + " " + s + " " + tick
}

// It is written to be pasted into a ticket or a thesis appendix without
// editing: measured values first, absences stated, remediation last.
func (r *Report) RenderMarkdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Link audit — %s\n\n", r.AuditID)
	fmt.Fprintf(&b, "- Site: %s\n- Link: %s (%s)\n- Measured: %s\n", r.Site, r.LinkType, r.Provider, r.Timestamp)
	fmt.Fprintf(&b, "- Sens: %s (charge: %d flux)\n", sensLabel(r.Measured.Direction), r.Measured.FlowCount)
	fmt.Fprintf(&b, "- Phases: %.0f s baseline / %.0f s load / %.0f s drain\n",
		r.Phases.BaselineSeconds, r.Phases.LoadSeconds, r.Phases.RecoverySeconds)
	// Reproduction — la commande qui rejoue cet audit avec ses réglages effectifs.
	// Un archive scellé avant ce champ le dit absent : jamais de commande inventée
	// pour remplir la ligne.
	if r.Invocation == nil || r.Invocation.Command == "" {
		b.WriteString("- Reproduction: non enregistrée (rapport scellé avant ce champ)\n")
	} else {
		fmt.Fprintf(&b, "- Reproduction: %s\n", codeSpan(r.Invocation.Command))
	}
	// Preuve — l'endroit où la donnée brute est archivée, scellé avec le rapport.
	if r.Artifact != "" {
		fmt.Fprintf(&b, "- Preuve: %s\n", codeSpan(r.Artifact))
	} else {
		b.WriteString("- Preuve: non enregistrée (aucun chemin d'artefact scellé)\n")
	}
	b.WriteString("\n")

	b.WriteString("## Measured\n\n")
	b.WriteString("| Metric | Value |\n|---|---|\n")
	fmt.Fprintf(&b, "| Idle RTT p50 | %.1f ms |\n", r.Measured.RTTIdleP50Ms)
	fmt.Fprintf(&b, "| Idle RTT p95 | %.1f ms |\n", r.Measured.RTTIdleP95Ms)
	if r.Measured.LoadMeasured {
		fmt.Fprintf(&b, "| Loaded RTT p50 | %.1f ms |\n", r.Measured.RTTLoadedP50Ms)
		fmt.Fprintf(&b, "| Loaded RTT p95 | %.1f ms |\n", r.Measured.RTTLoadedP95Ms)
		fmt.Fprintf(&b, "| Queue delay impairment (QDI) | %.1f ms |\n", r.Measured.QDIMs)
		fmt.Fprintf(&b, "| Bufferbloat delta (p95) | %.1f ms |\n", r.Measured.BloatDeltaMs)
		if r.Measured.DrainMeasured {
			fmt.Fprintf(&b, "| Queue drain after load | %.1f ms above rest |\n", r.Measured.DrainMs)
		}
		fmt.Fprintf(&b, "| Throughput | %.2f Mbit/s |\n", r.Measured.ThroughputMbps)
		fmt.Fprintf(&b, "| VoIP R-score | %.1f |\n", r.Measured.VoIPR)
		if r.Measured.FlowCount > 1 {
			if r.Measured.JFIMeasured {
				fmt.Fprintf(&b, "| Équité inter-flux (JFI) | %.1f %% |\n", r.Measured.JFIPct)
			} else {
				// L'index n'avait rien à comparer. Un « 0,0 % » se lit comme une
				// injustice totale : une absence nommée est le seul rendu honnête.
				b.WriteString("| Équité inter-flux (JFI) | non mesuré (aucune contribution par flux) |\n")
			}
		}
	}
	// Miroirs descendants — sous leur propre condition : un « both » avec
	// fenêtre up muette rend quand même la fenêtre down mesurée. Sous
	// LoadMeasured, un down muet afficherait des zéros déguisés en mesure.
	// Pour un audit « down », le headline porte déjà ces valeurs.
	if r.Measured.LoadMeasuredDown && r.Measured.Direction != "down" {
		if r.Measured.RTTLoadedDownP95Ms > 0 {
			fmt.Fprintf(&b, "| Chargé p95 (descendant) | %.1f ms |\n", r.Measured.RTTLoadedDownP95Ms)
			fmt.Fprintf(&b, "| Écart bufferbloat (descendant) | %.1f ms |\n", r.Measured.BloatDeltaDownMs)
		}
		fmt.Fprintf(&b, "| Throughput (descendant) | %.2f Mbit/s |\n", r.Measured.ThroughputDownMbps)
		if r.Measured.FlowCount > 1 {
			fmt.Fprintf(&b, "| JFI (descendant) | %.1f %% |\n", r.Measured.JFIDownPct)
		}
		if r.Measured.GradeDown != "" {
			fmt.Fprintf(&b, "| Bufferbloat grade (descendant, grade_down) | %s |\n", r.Measured.GradeDown)
		}
	}
	fmt.Fprintf(&b, "| Loss | %.2f %% |\n", r.Measured.LossPct)
	if r.Measured.SmallMeasured {
		fmt.Fprintf(&b, "| Small-object p95 | %.1f ms |\n", r.Measured.HTTPSmallP95Ms)
	}
	fmt.Fprintf(&b, "| Data consumed by audit | %.2f MB |\n\n", r.Measured.DataUsedMB)

	if r.Grade != "" {
		fmt.Fprintf(&b, "**Bufferbloat grade: %s** — %s\n\n", r.Grade, r.Verdict)
	} else {
		fmt.Fprintf(&b, "**Bufferbloat grade: not measured.** %s\n\n", r.Verdict)
	}

	// Le mot opérateur suit la note, jamais l'inverse. Un rapport scellé avant
	// cette tranche ne porte pas d'operator_verdict : l'absence se dit, au lieu
	// d'imprimer une ligne au verdict vide qui se lirait « rien à signaler ».
	if r.OperatorVerdict.Word != "" {
		b.WriteString("**Verdict opérateur : " + r.OperatorVerdict.Word + "** — " + r.OperatorVerdict.Sentence + "\n\n")
	} else {
		b.WriteString("**Verdict opérateur : non calculé** — le rapport ne porte pas d'`operator_verdict` : aucune cascade n'a été évaluée sur ses faits, et une absence n'est pas un feu vert.\n\n")
	}

	// SLA CAP — l'objectif opérateur DGM 220 ms dans le cadre WIS 2.0 :
	// jamais un succès par défaut, jamais présenté comme un seuil OMM.
	if r.CAPSLA.Measured {
		state := "CONFORME"
		if r.CAPSLA.Breach {
			state = r.CAPSLA.Label
		}
		fmt.Fprintf(&b, "**SLA CAP (objectif DGM 220 ms, cadre WIS 2.0) : %.1f %% sous échéance — %s**\n\n", r.CAPSLA.OKPct, state)
	} else {
		fmt.Fprintf(&b, "**SLA CAP : non mesuré** — aucune complétion 16 KiB chargée ; jamais un succès par défaut.\n\n")
	}

	if len(r.Unavailable) > 0 {
		b.WriteString("## Not measured\n\n")
		b.WriteString("Each of these is absent for a stated reason. None is zero-filled.\n\n")
		for _, u := range r.Unavailable {
			fmt.Fprintf(&b, "- **%s** — %s\n", u.Metric, u.Reason)
		}
		b.WriteString("\n")
	}

	if r.Provenance.SourceSHA256 != "" {
		b.WriteString("## Provenance\n\n")
		if r.Provenance.TelemetrySHA256 != "" {
			fmt.Fprintf(&b, "- Série brute (sonde par sonde): `%s`\n", r.Provenance.TelemetryFile)
			fmt.Fprintf(&b, "- SHA-256 télémétrie: `%s`\n", r.Provenance.TelemetrySHA256)
		}
		fmt.Fprintf(&b, "- Source: `%s`\n", r.Provenance.SourceFile)
		fmt.Fprintf(&b, "- Rows: %d\n", r.Provenance.RowCount)
		fmt.Fprintf(&b, "- SHA-256: `%s`\n\n", r.Provenance.SourceSHA256)
	}

	if r.Economics != nil {
		b.WriteString("## Impact économique (scénarios contractuels)\n\n")
		fmt.Fprintf(&b, "Ratio de réémission : %.4f (base %s) · %.2f Mo gaspillés sur %.2f Mo transportés en %.0f s de charge.\n\n",
			r.Economics.RetransRatio.Ratio, r.Economics.RetransRatio.Basis,
			r.Economics.WastedBytes/1e6, r.Economics.TransportedBytes/1e6, r.Economics.LoadSeconds)
		b.WriteString("| Scénario | Ar/h | Base |\n|---|---|---|\n")
		for _, s := range r.Economics.Scenarios {
			mark := ""
			if r.Economics.Primary != nil && s.Name == r.Economics.Primary.Name {
				mark = " ← PRIMAIRE (" + r.Economics.PrimaryLinkType + ")"
			}
			fmt.Fprintf(&b, "| %s (%s)%s | %.0f | %s |\n", s.Name, s.Model, mark, s.ARPerHour, s.Basis)
		}
		fmt.Fprintf(&b, "\n%s\n\n", r.Economics.Disclaimer)
	}

	if r.Remediation != nil {
		b.WriteString("## Remediation\n\n")
		fmt.Fprintf(&b, "Recommended qdisc: **%s** — %s\n\n", r.Remediation.Qdisc, r.Remediation.Reason)
		if len(r.Remediation.LinuxTC) > 0 {
			b.WriteString("### Linux — upload (egress)\n\n```sh\n")
			for _, c := range r.Remediation.LinuxTC {
				b.WriteString(c + "\n")
			}
			b.WriteString("```\n\n")
		}
		if len(r.Remediation.LinuxTCIngress) > 0 {
			b.WriteString("### Linux — download (ingress IFB)\n\n```sh\n")
			for _, c := range r.Remediation.LinuxTCIngress {
				b.WriteString(c + "\n")
			}
			b.WriteString("```\n\n")
			if r.Remediation.Qdisc == remediate.QdiscCake {
				b.WriteString("CAKE runs in `ingress` mode on `ifb0`: the packet has already crossed the link when it is dropped, and `ack-filter` does not apply to a download data stream.\n\n")
			} else {
				b.WriteString("The download is shaped on `ifb0` after the packet has crossed the link; `ingress` is CAKE-only and is not used with this qdisc, and `ack-filter` does not apply to a download data stream.\n\n")
			}
		}
		if r.Remediation.RouterOS != "" {
			b.WriteString("### MikroTik RouterOS v7+ — upload (WAN)\n\n```\n")
			b.WriteString(r.Remediation.RouterOS)
			b.WriteString("```\n\n")
		}
		if r.Remediation.RouterOSIngress != "" {
			b.WriteString("### MikroTik RouterOS v7+ — download (LAN queue tree)\n\n```\n")
			b.WriteString(r.Remediation.RouterOSIngress)
			b.WriteString("```\n\n")
			b.WriteString("RouterOS has no IFB: this is a LAN-interface egress queue tree, not byte-equivalent to the Linux IFB.\n\n")
		}
	}
	if v := r.RouterOSValidation; v != nil {
		// Zéro vérification n'est pas une validation : un tableau vide
		// passait pour un succès dans le Markdown comme dans la carte.
		ok := len(v.Checks) > 0
		for _, c := range v.Checks {
			if !c.Pass {
				ok = false
			}
		}
		level := "CHR"
		if v.Level == "linux-namespace" {
			level = "CHR + banc"
		}
		switch {
		case len(v.Checks) == 0:
			fmt.Fprintf(&b, "### Validation RouterOS\n\nConfiguration RouterOS non validée (%s, %s) — aucune vérification.\n\n", v.Version, level)
		case ok:
			fmt.Fprintf(&b, "### Validation RouterOS\n\nConfiguration RouterOS validée (%s, %s) · export %s\n\n", v.Version, level, short8(v.ExportSHA256))
		default:
			fmt.Fprintf(&b, "### Validation RouterOS\n\nConfiguration RouterOS non validée (%s, %s) — écarts dans le rapport.\n\n", v.Version, level)
		}
	}
	if k := r.KernelTelemetry; k != nil {
		switch k.Source {
		case "ebpf":
			fmt.Fprintf(&b, "### Télémétrie noyau\n\nTélémétrie noyau : eBPF — %d octets réémis, %d paquets écartés.\n\n", k.RetransPayloadBytes, k.QdiscDropPackets)
		case "tcpinfo":
			fmt.Fprintf(&b, "### Télémétrie noyau\n\nRepli TCP_INFO — %s.\n\n", k.Reason)
		default:
			fmt.Fprintf(&b, "### Télémétrie noyau\n\nTélémétrie noyau non instrumentée%s.\n\n", reasonClause(k.Reason))
		}
	}

	return b.String()
}

func reasonClause(reason string) string {
	if reason == "" {
		return ""
	}
	return " — " + reason
}

func short8(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// SortUnavailable orders the unavailable list for stable output.
func (r *Report) SortUnavailable() {
	sort.Slice(r.Unavailable, func(i, j int) bool {
		return r.Unavailable[i].Metric < r.Unavailable[j].Metric
	})
}

// WriteArtifacts writes the JSON and Markdown reports. Each audit gets its own
// directory (data/audits/<id>/audit_report.*) and the latest copy is mirrored
// at data/audit_report.* for operators and the dashboard, which do not need to
// know the audit id ahead of time.
func WriteArtifacts(rep *Report, baseDir string) (jsonPath, mdPath string, err error) {
	if rep == nil {
		return "", "", fmt.Errorf("audit: nil report")
	}
	if baseDir == "" {
		baseDir = "data"
	}
	dir, err := ArtifactDir(baseDir, rep.AuditID)
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	js, err := rep.RenderJSON()
	if err != nil {
		return "", "", err
	}
	md := []byte(rep.RenderMarkdown())
	jsonPath = filepath.Join(dir, ReportArtifactName)
	mdPath = filepath.Join(dir, ReportMarkdownName)
	if err := os.WriteFile(jsonPath, js, 0644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(mdPath, md, 0644); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(filepath.Join(baseDir, "audit_report.json"), js, 0644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(filepath.Join(baseDir, "audit_report.md"), md, 0644); err != nil {
		return "", "", err
	}
	return jsonPath, mdPath, nil
}

// sensLabel — libellé français du sens pour le rendu opérateur.
func sensLabel(dir string) string {
	switch dir {
	case "down":
		return "descendant (down)"
	case "both":
		return "les deux sens (both)"
	case "up":
		return "montant (up)"
	}
	return dir
}
