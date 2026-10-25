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

	// Provenance seals the raw telemetry this report was built from.
	Provenance Provenance `json:"provenance"`

	// Remediation is present only when capacity and RTT were both measured.
	Remediation *Remediation `json:"remediation,omitempty"`

	// Waste is the financial projection, present only when a load was measured.
	Waste *Waste `json:"waste,omitempty"`
}

// Waste is the retransmission cost projection in Malagasy Ariary.
//
// It is derived from the measured loss and throughput, not from a guess: the
// audit knows how many bytes it moved and what fraction was lost, so the wasted
// volume is measured. The tariff is the operator's declared tier.
type Waste struct {
	// Tier is the prepaid tariff the projection uses.
	Tier string `json:"tier"`
	// ARPerGB is that tariff's rate.
	ARPerGB float64 `json:"ar_per_gb"`
	// WastedMB is the measured volume lost to retransmission.
	WastedMB float64 `json:"wasted_mb"`
	// ARPerHour is the projected cost of that waste per hour of the same load.
	ARPerHour float64 `json:"ar_per_hour"`
	// Basis states how the hourly figure was extrapolated, so it is auditable.
	Basis string `json:"basis"`
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
}

// Unavailable records a metric the audit could not produce, and why.
type Unavailable struct {
	Metric string `json:"metric"`
	Reason string `json:"reason"`
}

// Provenance seals the evidence chain.
type Provenance struct {
	// SourceFile is the CSV the report was rendered from.
	SourceFile string `json:"source_file"`
	// SourceSHA256 is the SHA-256 of that file's bytes at render time.
	SourceSHA256 string `json:"source_sha256"`
	// RowCount is the number of data rows in the source file.
	RowCount int `json:"row_count"`
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

	loadMeasured := res.ThroughputMbps > 0 && res.BloatGrade != ""
	rep.Measured = Measured{
		RTTIdleP50Ms:   res.RTTIdleP50,
		RTTIdleP95Ms:   res.RTTIdleP95,
		RTTLoadedP50Ms: res.RTTLoadedP50,
		RTTLoadedP95Ms: res.RTTLoadedP95,
		QDIMs:          res.RTTLoadedP95 - res.RTTLoadedP50,
		BloatDeltaMs:   res.BloatDeltaMs,
		DrainMs:        res.DrainMs,
		DrainMeasured:  res.DrainMeasured,
		LoadMeasured:   loadMeasured,
		ThroughputMbps: res.ThroughputMbps,
		LossPct:        res.LossPct,
		HTTPSmallP95Ms: res.HTTPSmallP95,
		SmallMeasured:  res.HTTPSmallP95 > 0,
		DataUsedMB:     res.DataUsedMB,
	}

	// VoIP R from the loaded conditions: one-way delay is half the RTT, jitter
	// is the p95-p50 spread the jitter buffer has to absorb.
	if loadMeasured {
		rep.Measured.VoIPR = metrics.VoIPR(res.RTTLoadedP50/2, res.RTTLoadedP95-res.RTTLoadedP50, res.LossPct)
	}

	// Financial projection. The audit measured how much data it moved and what
	// fraction was lost, so the wasted volume is a measurement, not an estimate.
	// The hourly figure extrapolates that waste rate to one hour of the same
	// load, which is the number an operator can act on.
	if loadMeasured && res.DataUsedMB > 0 && res.LossPct > 0 && res.LoadSeconds > 0 {
		tier := metrics.ActiveTier()
		wastedMB := res.DataUsedMB * res.LossPct / 100
		// Extrapolate the measured waste rate over the measured load window to
		// one hour. Both inputs are measurements, so the hourly figure is
		// traceable rather than assumed.
		perHour := wastedMB / res.LoadSeconds * 3600
		rep.Waste = &Waste{
			Tier:      tier.Name,
			ARPerGB:   tier.ARGB,
			WastedMB:  wastedMB,
			ARPerHour: perHour / 1024 * tier.ARGB,
			Basis: fmt.Sprintf("%.2f MB lost over a measured %.1f s load, extrapolated to one hour at %s (%.0f Ar/GB)",
				wastedMB, res.LoadSeconds, tier.Name, tier.ARGB),
		}
	} else if loadMeasured {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "waste_ar_per_h",
			Reason: "no measured loss or no measured volume; a zero here would mean 'no waste', which is a different claim from 'not measured'",
		})
	}

	// Say plainly what is missing. These are the metrics a jury or an operator
	// will ask for, so each absence is named rather than silently zeroed.
	if !loadMeasured {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "bufferbloat grade",
			Reason: "no bulk load ran, so the loaded window was never under load; a grade here would be fabricated",
		})
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "throughput",
			Reason: "no bulk sink reachable; supply -bulk host:port or install iperf3",
		})
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "voip_r",
			Reason: "requires loaded conditions; not computed without a measured load",
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
	// Deadline compliance needs per-probe completion samples, which a link
	// audit does not retain. The campaign does; this report does not pretend to.
	rep.Unavailable = append(rep.Unavailable, Unavailable{
		Metric: "deadline_ok_pct",
		Reason: "link audit retains percentiles only; per-probe completions are recorded by the campaign, not the audit",
	})

	// Remediation only from a real measurement. This is the guard that keeps a
	// guessed bandwidth out of a production router config.
	if res.ThroughputMbps > 0 && res.RTTIdleP50 > 0 {
		link.CapacityMbps = res.ThroughputMbps
		link.RTTms = res.RTTIdleP50
		rec := remediate.Recommend(link, asymmetric, cpuConstrained)
		useQdisc := qdisc
		if useQdisc == "" {
			useQdisc = rec.Qdisc
		}
		tcCmds, err := remediate.LinuxTC(link, useQdisc)
		if err == nil {
			rem := &Remediation{Qdisc: useQdisc, Reason: rec.Reason, LinuxTC: tcCmds}
			if link.RouterOSInterface != "" {
				if script, err := remediate.RouterOS(link, useQdisc); err == nil {
					rem.RouterOS = script
				}
			}
			rep.Remediation = rem
		}
	} else {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "remediation",
			Reason: "capacity or idle RTT unmeasured; a queue config sized from a guess would be deployed blindly",
		})
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
	r.Provenance = Provenance{
		SourceFile:   path,
		SourceSHA256: hex.EncodeToString(sum[:]),
		RowCount:     countCSVRows(data),
	}
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
// It is written to be pasted into a ticket or a thesis appendix without
// editing: measured values first, absences stated, remediation last.
func (r *Report) RenderMarkdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Link audit — %s\n\n", r.AuditID)
	fmt.Fprintf(&b, "- Site: %s\n- Link: %s (%s)\n- Measured: %s\n", r.Site, r.LinkType, r.Provider, r.Timestamp)
	fmt.Fprintf(&b, "- Phases: %.0f s baseline / %.0f s load / %.0f s drain\n\n",
		r.Phases.BaselineSeconds, r.Phases.LoadSeconds, r.Phases.RecoverySeconds)

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
		fmt.Fprintf(&b, "- Source: `%s`\n", r.Provenance.SourceFile)
		fmt.Fprintf(&b, "- Rows: %d\n", r.Provenance.RowCount)
		fmt.Fprintf(&b, "- SHA-256: `%s`\n\n", r.Provenance.SourceSHA256)
	}

	if r.Waste != nil {
		b.WriteString("## Financial waste\n\n")
		fmt.Fprintf(&b, "- Tariff: %s (%.0f Ar/GB)\n", r.Waste.Tier, r.Waste.ARPerGB)
		fmt.Fprintf(&b, "- Volume lost to retransmission: %.2f MB\n", r.Waste.WastedMB)
		fmt.Fprintf(&b, "- Projected cost: **%.0f Ar/hour**\n\n", r.Waste.ARPerHour)
		fmt.Fprintf(&b, "Basis: %s\n\n", r.Waste.Basis)
	}

	if r.Remediation != nil {
		b.WriteString("## Remediation\n\n")
		fmt.Fprintf(&b, "Recommended qdisc: **%s** — %s\n\n", r.Remediation.Qdisc, r.Remediation.Reason)
		b.WriteString("### Linux\n\n```sh\n")
		for _, c := range r.Remediation.LinuxTC {
			b.WriteString(c + "\n")
		}
		b.WriteString("```\n\n")
		if r.Remediation.RouterOS != "" {
			b.WriteString("### MikroTik RouterOS v7+\n\n```\n")
			b.WriteString(r.Remediation.RouterOS)
			b.WriteString("```\n\n")
		}
	}

	return b.String()
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
	id := filepath.Base(strings.TrimSpace(rep.AuditID))
	if id == "" || id == "." || id == string(filepath.Separator) {
		return "", "", fmt.Errorf("audit: invalid audit id %q", rep.AuditID)
	}
	dir := filepath.Join(baseDir, "audits", id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	js, err := rep.RenderJSON()
	if err != nil {
		return "", "", err
	}
	md := []byte(rep.RenderMarkdown())
	jsonPath = filepath.Join(dir, "audit_report.json")
	mdPath = filepath.Join(dir, "audit_report.md")
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
