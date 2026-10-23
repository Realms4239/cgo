package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
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
	}

	loadMeasured := res.ThroughputMbps > 0 && res.BloatGrade != ""
	rep.Measured = Measured{
		RTTIdleP50Ms:   res.RTTIdleP50,
		RTTIdleP95Ms:   res.RTTIdleP95,
		RTTLoadedP50Ms: res.RTTLoadedP50,
		RTTLoadedP95Ms: res.RTTLoadedP95,
		QDIMs:          res.RTTLoadedP95 - res.RTTLoadedP50,
		BloatDeltaMs:   res.BloatDeltaMs,
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
	fmt.Fprintf(&b, "- Site: %s\n- Link: %s (%s)\n- Measured: %s\n\n",
		r.Site, r.LinkType, r.Provider, r.Timestamp)

	b.WriteString("## Measured\n\n")
	b.WriteString("| Metric | Value |\n|---|---|\n")
	fmt.Fprintf(&b, "| Idle RTT p50 | %.1f ms |\n", r.Measured.RTTIdleP50Ms)
	fmt.Fprintf(&b, "| Idle RTT p95 | %.1f ms |\n", r.Measured.RTTIdleP95Ms)
	if r.Measured.LoadMeasured {
		fmt.Fprintf(&b, "| Loaded RTT p50 | %.1f ms |\n", r.Measured.RTTLoadedP50Ms)
		fmt.Fprintf(&b, "| Loaded RTT p95 | %.1f ms |\n", r.Measured.RTTLoadedP95Ms)
		fmt.Fprintf(&b, "| Queue delay impairment (QDI) | %.1f ms |\n", r.Measured.QDIMs)
		fmt.Fprintf(&b, "| Bufferbloat delta (p95) | %.1f ms |\n", r.Measured.BloatDeltaMs)
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
