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
	for _, f := range s.Flows {
		flows = append(flows, audit.KernelFlow{
			Local: f.Local, Peer: f.Peer,
			RetransPayloadBytes: f.RetransPayload,
			RetransEvents:       f.RetransEvents,
			QdiscDropPackets:    f.DropPackets,
			SRTTus:              f.SRTTus,
		})
	}
	tel := audit.KernelTelemetryFrom(audit.KernelSample{
		Source: s.Source, Attach: s.Attach, Reason: s.FallbackReason,
		ObjectSHA256: s.ObjectSHA256,
		FlowsSeen:    s.FlowsSeen, FlowsTotal: s.FlowsTotal,
		RetransPayload: s.RetransPayload, QdiscDrops: s.QdiscDrops,
		Flows: flows,
	})
	if tel != nil {
		// Perte d'observateur : le noyau n'émet aucun record de perte sur un
		// ring plein. Le compteur vient de globals, hors du flux d'événements.
		tel.ReserveFailures = s.ReserveFailures
		tel.LowerBound = s.ReserveFailures > 0
		tel.MapFull = s.MapFull
		// Table pleine = perte d'observateur au même titre que le ring plein :
		// dans les deux cas les queues ne sont qu'une borne inférieure.
		if s.MapFull > 0 {
			tel.LowerBound = true
		}
		// Dénominateur réel : les compteurs d'événements tentés (globals, hors
		// ring saturé). ReserveFailures reste le seul numérateur, non dupliqué.
		tel.ObserverEvents, tel.ObserverLossPct, tel.ObserverLossDenominator =
			audit.ObserverLoss(s.RetransEvents+s.QdiscDrops, s.ReserveFailures)
		// Queues — sous le plancher de 1000, le compteur est porté mais aucun
		// percentile n'est affirmé : l'absence reste un mot, pas un zéro.
		tel.LatencySamples = s.LatencySamples
		if s.LatencySamples >= 1000 {
			tel.LatencyP50us, _ = bpf.Percentile(s.LatencyHistogram, 0.50, s.LatencySamples)
			tel.LatencyP95us, _ = bpf.Percentile(s.LatencyHistogram, 0.95, s.LatencySamples)
			tel.LatencyP999us, _ = bpf.Percentile(s.LatencyHistogram, 0.999, s.LatencySamples)
			if tel.LatencyP999us > 0 {
				tel.TailBinBounds = "bucket i = [2^i, 2^(i+1)) us"
			}
		}
	}
	return tel
}
