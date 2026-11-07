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
	return audit.KernelTelemetryFrom(audit.KernelSample{
		Source: s.Source, Attach: s.Attach, Reason: s.FallbackReason,
		ObjectSHA256: s.ObjectSHA256,
		FlowsSeen:    s.FlowsSeen, FlowsTotal: s.FlowsTotal,
		RetransPayload: s.RetransPayload, QdiscDrops: s.QdiscDrops,
		Flows: flows,
	})
}
