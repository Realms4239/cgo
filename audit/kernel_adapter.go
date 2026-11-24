package audit

// maxRawSRTTus — plafond du tableau d'échantillons bruts. Un srtt_us par flux
// suffit à la fidélité ; au-delà, la charge utile ne justifie pas une borne
// non bornée dans un rapport scellé.
const maxRawSRTTus = 512

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
	// Pertes d'observateur lues dans globals, hors du ring saturé.
	ReserveFailures uint64
	MapFull         uint64
	// Queues (tails) — percentiles réduits par la sonde depuis l'histogramme
	// log2 du noyau. Sous 1000 échantillons, ces champs restent zéro : aucune
	// valeur n'est extrapolée pour ressembler à une mesure.
	LatencySamples uint64
	LatencyP50us   float64
	LatencyP95us   float64
	LatencyP999us  float64
	TailBinBounds  string
	// RawSRTTus — srtt_us par flux, brut et non interpolé, borné à la frontière.
	RawSRTTus []uint32
	Flows     []KernelFlow
}

// KernelTelemetryFrom — nil quand la sonde n'a rien produit : un absent ne
// devient pas « tcpinfo » par défaut, il reste absent du rapport.
func KernelTelemetryFrom(s KernelSample) *KernelTelemetry {
	if s.Source == "" {
		return nil
	}
	tel := &KernelTelemetry{
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
	tel.ReserveFailures = s.ReserveFailures
	tel.MapFull = s.MapFull
	tel.LatencySamples = s.LatencySamples
	if s.LatencySamples >= 1000 {
		tel.LatencyP50us = s.LatencyP50us
		tel.LatencyP95us = s.LatencyP95us
		tel.LatencyP999us = s.LatencyP999us
		if s.LatencyP999us > 0 {
			tel.TailBinBounds = s.TailBinBounds
		}
	}
	if len(s.RawSRTTus) > 0 {
		n := len(s.RawSRTTus)
		if n > maxRawSRTTus {
			n = maxRawSRTTus
		}
		tel.LatencySamplesSRTT = append([]uint32(nil), s.RawSRTTus[:n]...)
		if tel.Reason != "" {
			tel.Reason += "; "
		}
		tel.Reason += "latency_samples_srtt = srtt_us par flux (brut, non interpolé, <=512)"
	}
	return tel
}
