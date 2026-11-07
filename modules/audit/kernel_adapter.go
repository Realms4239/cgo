package audit

// KernelSample — la vue neutre qu'une sonde remet au rapport. Elle existe pour
// que pkg/audit n'importe jamais pkg/probe/bpf : le chargeur mappe son Summary
// sur cette structure, et le rapport reste indépendant du format noyau.
type KernelSample struct {
	Source         string
	Attach         string
	Reason         string
	ObjectSHA256   string
	FlowsSeen      int
	FlowsTotal     int
	RetransPayload uint64
	QdiscDrops     uint64
	Flows          []KernelFlow
}

// KernelTelemetryFrom — nil quand la sonde n'a rien produit : un absent ne
// devient pas « tcpinfo » par défaut, il reste absent du rapport.
func KernelTelemetryFrom(s KernelSample) *KernelTelemetry {
	if s.Source == "" {
		return nil
	}
	return &KernelTelemetry{
		Available:           s.Source == "ebpf",
		Source:              s.Source,
		Reason:              s.Reason,
		ObjectSHA256:        s.ObjectSHA256,
		RetransPayloadBytes: s.RetransPayload,
		QdiscDropPackets:    s.QdiscDrops,
		FlowsSeen:           s.FlowsSeen,
		FlowsTotal:          s.FlowsTotal,
		Flows:               s.Flows,
	}
}
