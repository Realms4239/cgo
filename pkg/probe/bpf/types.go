package bpf

// Flow — un flux côté émetteur, lu dans la table noyau.
type Flow struct {
	Local          string `json:"local"`
	Peer           string `json:"peer"`
	RetransPayload uint64 `json:"retrans_payload_bytes"`
	RetransEvents  uint64 `json:"retrans_events"`
	DropPackets    uint64 `json:"qdisc_drop_packets"`
	DropBytes      uint64 `json:"qdisc_drop_bytes"`
	SRTTus         uint32 `json:"srtt_us"`
	SndCwnd        uint32 `json:"snd_cwnd"`
	SndSsthresh    uint32 `json:"snd_ssthresh"`
	State          uint32 `json:"state"`
	Samples        uint32 `json:"samples"`
}

// Summary — une fenêtre de télémétrie. Source dit quel ratio en est tiré :
// ebpf | tcpinfo | "" ; Available dit si la session a été attachée.
type Summary struct {
	Available      bool   `json:"available"`
	Source         string `json:"source,omitempty"`
	Attach         string `json:"attach,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ObjectSHA256   string `json:"object_sha256,omitempty"`
	FlowsSeen      int    `json:"flows_seen"`
	FlowsTotal     int    `json:"flows_total"`
	RetransPayload uint64 `json:"retrans_payload_bytes"`
	RetransEvents  uint64 `json:"retrans_events"`
	QdiscDrops     uint64 `json:"qdisc_drop_packets"`
	QdiscDropBytes uint64 `json:"qdisc_drop_bytes"`
	QdiscAvailable bool   `json:"qdisc_available"`
	MapFull        uint64 `json:"map_full,omitempty"`
	Flows          []Flow `json:"flows,omitempty"`
}

// Globals — diagnostics hôte, jamais une affirmation par flux. Retrans
// porte le repli global (IPv6/non-IPv4 et table des flux pleine) : sans ces
// champs, ces octets disparaîtraient au chargement.
type Globals struct {
	RetransPayload uint64
	RetransEvents  uint64
	DropPackets    uint64
	DropBytes      uint64
	MapFull        uint64
	QdiscAvailable bool
}
