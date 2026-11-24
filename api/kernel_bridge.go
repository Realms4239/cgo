package api

import (
	"github.com/Realms4239/cgo/audit"
	"github.com/Realms4239/cgo/probe/bpf"
)

// kernelTelemetryFromSummary — le pont unique entre le résumé noyau et le
// rapport. Il vit dans la couche de composition (elle voit audit et bpf) :
// pkg/audit n'importe pas le format eBPF, et pkg/probe ne peut pas importer
// audit (cycle). Le chargeur appellera cette fonction quand l'objet existera.
func kernelTelemetryFromSummary(s bpf.Summary) *audit.KernelTelemetry {
	flows := make([]audit.KernelFlow, 0, len(s.Flows))
	rawSRTT := make([]uint32, 0, len(s.Flows))
	for _, f := range s.Flows {
		flows = append(flows, audit.KernelFlow{
			Local: f.Local, Peer: f.Peer,
			RetransPayloadBytes: f.RetransPayload,
			RetransEvents:       f.RetransEvents,
			QdiscDropPackets:    f.DropPackets,
			SRTTus:              f.SRTTus,
		})
		rawSRTT = append(rawSRTT, f.SRTTus)
	}
	// Les champs de queue sont réduits ici, côté sonde, puis portés par le
	// KernelSample : KernelTelemetryFrom applique le plancher et le basis.
	sample := audit.KernelSample{
		Source: s.Source, Attach: s.Attach, Reason: s.FallbackReason,
		ObjectSHA256: s.ObjectSHA256,
		FlowsSeen:    s.FlowsSeen, FlowsTotal: s.FlowsTotal,
		RetransPayload:  s.RetransPayload,
		QdiscDrops:      s.QdiscDrops,
		ReserveFailures: s.ReserveFailures,
		MapFull:         s.MapFull,
		LatencySamples:  s.LatencySamples,
		RawSRTTus:       rawSRTT,
		Flows:           flows,
	}
	if s.LatencySamples >= 1000 {
		sample.LatencyP50us, _ = bpf.Percentile(s.LatencyHistogram, 0.50, s.LatencySamples)
		sample.LatencyP95us, _ = bpf.Percentile(s.LatencyHistogram, 0.95, s.LatencySamples)
		sample.LatencyP999us, _ = bpf.Percentile(s.LatencyHistogram, 0.999, s.LatencySamples)
		if sample.LatencyP999us > 0 {
			sample.TailBinBounds = "bucket i = [2^i, 2^(i+1)) us"
		}
	}
	tel := audit.KernelTelemetryFrom(sample)
	if tel != nil {
		// Perte d'observateur : le compteur vient de globals, hors du flux
		// d'événements. ReserveFailures et MapFull sont déjà portés par le
		// KernelSample ; ici on ne fait que nommer la borne inférieure.
		tel.LowerBound = s.ReserveFailures > 0
		// Table pleine = perte d'observateur au même titre que le ring plein :
		// dans les deux cas les queues ne sont qu'une borne inférieure.
		if s.MapFull > 0 {
			tel.LowerBound = true
		}
		// Dénominateur réel : les compteurs d'événements tentés (globals, hors
		// ring saturé). ReserveFailures reste le seul numérateur, non dupliqué.
		tel.ObserverEvents, tel.ObserverLossPct, tel.ObserverLossDenominator =
			audit.ObserverLoss(s.RetransEvents+s.QdiscDrops, s.ReserveFailures)
	}
	return tel
}
