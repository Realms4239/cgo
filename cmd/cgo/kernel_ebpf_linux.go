//go:build linux && ebpf_object

package main

// kernel_ebpf_linux — la moitié eBPF du CLI de terrain ; la moitié Windows
// (kernel_ebpf_other.go) reste stummiente.

import (
	"github.com/Realms4239/cgo/audit"
	"github.com/Realms4239/cgo/probe/bpf"
)

// attachKernelProbe câble le harnais eBPF dans les deps CLI. Sur tout refus
// (objet absent, BTF absent, CAP_BPF manquant, attach refusée), l'audit
// continue en tcpinfo et le rapport le dit — jamais un audit qui échoue
// parce que la télémétrie auxiliaire a refusé.
func attachKernelProbe(d *audit.Deps) {
	d.KernelAttach = func() func() audit.KernelSample {
		sess, err := bpf.Attach()
		if err != nil {
			return func() audit.KernelSample {
				return audit.KernelSample{Source: "tcpinfo", Reason: "eBPF attach refusé : " + err.Error()}
			}
		}
		return func() audit.KernelSample {
			defer sess.Close()
			sum, err := sess.Summary()
			if err != nil {
				return audit.KernelSample{Source: "tcpinfo", Reason: "eBPF résumé indisponible : " + err.Error()}
			}
			s := audit.KernelSample{
				Source:          "ebpf",
				Attach:          sum.Attach,
				ObjectSHA256:    sum.ObjectSHA256,
				RetransPayload:  sum.RetransPayload,
				QdiscDrops:      sum.QdiscDrops,
				FlowsSeen:       sum.FlowsSeen,
				FlowsTotal:      sum.FlowsTotal,
				ReserveFailures: sum.ReserveFailures,
				MapFull:         sum.MapFull,
				LatencySamples:  sum.LatencySamples,
			}
			// Sans carte d'histogramme, LatencySamples vaut 0 : aucun
			// percentile n'est affirmé, jamais un zéro fabriqué.
			if sum.LatencySamples >= 1000 {
				s.LatencyP50us, _ = bpf.Percentile(sum.LatencyHistogram, 0.50, sum.LatencySamples)
				s.LatencyP95us, _ = bpf.Percentile(sum.LatencyHistogram, 0.95, sum.LatencySamples)
				s.LatencyP999us, _ = bpf.Percentile(sum.LatencyHistogram, 0.999, sum.LatencySamples)
				if s.LatencyP999us > 0 {
					s.TailBinBounds = "bucket i = [2^i, 2^(i+1)) us"
				}
			}
			for _, f := range sum.Flows {
				s.Flows = append(s.Flows, audit.KernelFlow{
					Local:               f.Local,
					Peer:                f.Peer,
					RetransPayloadBytes: f.RetransPayload,
					RetransEvents:       f.RetransEvents,
					QdiscDropPackets:    f.DropPackets,
					SRTTus:              f.SRTTus,
				})
				s.RawSRTTus = append(s.RawSRTTus, f.SRTTus)
			}
			return s
		}
	}
}
