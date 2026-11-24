//go:build linux && ebpf_object

package bpf

// loader_linux — la moitié qui attache. Conçue pour le banc Linux 6.8
// (BTF présent), avec repli kprobe si l'attache fentry refuse.

import (
	"bytes"
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

// Registre brut des maps et sections de l'objet — les tailles doivent
// refléter struct flow_key / flow_stat / global_stat telles que compilées.
type flowKey struct {
	Saddr  uint32
	Daddr  uint32
	Sport  uint16
	Dport  uint16
	Family uint16
	Pad    uint16
}

type flowStat struct {
	RetransPayload uint64
	RetransEvents  uint64
	DropPackets    uint64
	DropBytes      uint64
	SRTTUsec       uint32
	SndCWnd        uint32
	SndSthresh     uint32
	State          uint32
	LastSampleNs   uint64
	Samples        uint32
	Pad            uint32
}

type globalStat struct {
	RetransPayload  uint64
	RetransEvents   uint64
	DropPackets     uint64
	DropBytes       uint64
	MapFull         uint64
	ReserveFailures uint64
}

func init() {
	// L'objet CO-RE a été compilé pour une struct flow_stat de 64 octets —
	// la vérification vite-stupide ici ferait paniquer l'ancienne planter.
	if sz := unsafe.Sizeof(flowStat{}); sz != 64 {
		panic(fmt.Sprintf("flow_stat = %d octets, want 64 — Layout désynchronisée", sz))
	}
	if sz := unsafe.Sizeof(flowKey{}); sz != 16 {
		panic(fmt.Sprintf("flow_key = %d octets, want 16", sz))
	}
	if sz := unsafe.Sizeof(globalStat{}); sz != 48 {
		panic(fmt.Sprintf("global_stat = %d octets, want 48", sz))
	}
}

// Session — l'attache vivante : fermer dans le bon ordre, jamais fuir.
type Session struct {
	Coll     *ebpf.Collection
	links    []link.Link
	flows    *ebpf.Map
	global   *ebpf.Map
	hist     *ebpf.Map
	events   *ebpf.Map
	attached []string
}

// Attach charge l'objet embarqué, attache les trois tracepoints, et rend
// la main. En cas d'échec en cours de route, tout ce qui a été attaché est
// libéré — jamais à moitié attachée.
func Attach() (*Session, error) {
	obj, err := ObjectBytes()
	if err != nil {
		return nil, err
	}
	_ = rlimit.RemoveMemlock()

	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(obj))
	if err != nil {
		return nil, fmt.Errorf("spécification eBPF illisible : %w", err)
	}
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return nil, fmt.Errorf("chargement du collection eBPF : %w", err)
	}
	s := &Session{Coll: coll, flows: coll.Maps["flows"], global: coll.Maps["globals"], hist: coll.Maps["latency_hist"], events: coll.Maps["events"]}
	cleanup := func(e error) error {
		s.Close()
		return e
	}

	attach := func(name string, fn func() (link.Link, error)) error {
		l, err := fn()
		if err != nil {
			return fmt.Errorf("attache %s : %w", name, err)
		}
		s.links = append(s.links, l)
		s.attached = append(s.attached, name)
		return nil
	}
	if err := attach("raw_tp/tcp_retransmit_skb", func() (link.Link, error) {
		return link.AttachRawTracepoint(link.RawTracepointOptions{
			Name:    "tcp_retransmit_skb",
			Program: coll.Programs["on_retransmit"],
		})
	}); err != nil {
		return nil, cleanup(err)
	}
	// fentry (BTF) en premier, kprobe en repli — même sémantique, visée différente.
	if err := attach("fentry/tcp_rcv_established", func() (link.Link, error) {
		return link.AttachTracing(link.TracingOptions{
			Program: coll.Programs["on_ack_fentry"],
		})
	}); err != nil {
		if err2 := attach("kprobe/tcp_rcv_established", func() (link.Link, error) {
			return link.Kprobe("tcp_rcv_established", coll.Programs["on_ack_kprobe"], nil)
		}); err2 != nil {
			return nil, cleanup(fmt.Errorf("attache fentry refusée (%v) et kprobe aussi (%v)", err, err2))
		}
	}
	if err := attach("raw_tp/kfree_skb", func() (link.Link, error) {
		return link.AttachRawTracepoint(link.RawTracepointOptions{
			Name:    "kfree_skb",
			Program: coll.Programs["on_kfree_skb"],
		})
	}); err != nil {
		return nil, cleanup(err)
	}
	return s, nil
}

// Close — détache dans l'ordre inverse d'attache, libère la collection.
func (s *Session) Close() error {
	var first error
	for i := len(s.links) - 1; i >= 0; i-- {
		if err := s.links[i].Close(); err != nil && first == nil {
			first = err
		}
	}
	if s.Coll != nil {
		s.Coll.Close()
	}
	s.links = nil
	s.flows = nil
	s.global = nil
	s.hist = nil
	s.events = nil
	return first
}

// Attached — la liste des tracepoints effectivement attachés, pour le sceau.
func (s *Session) Attached() []string {
	out := make([]string, len(s.attached))
	copy(out, s.attached)
	return out
}

// drainEvents vide les événements déjà disponibles sans bloquer la fenêtre.
// Voie observationnelle : un échec de vidage n'invalide pas les agrégats.
func (s *Session) drainEvents() {
	if s.events == nil {
		return // objet précompilé sans ring : rien à vider
	}
	rd, err := ringbuf.NewReader(s.events)
	if err != nil {
		return
	}
	defer rd.Close()
	if rd.SetDeadline(time.Now()) == nil {
		_, _ = DrainEvents(rd, nil)
	}
}

// Summary — fige la fenêtre : flux + compteurs globaux, jamais une
// affirmation non mesurée. Un flux n'est « seen » que s'il a été échantillonné.
func (s *Session) Summary() (Summary, error) {
	s.drainEvents()
	summary := Summary{Available: true, Source: "ebpf", Attach: ""}
	for i, a := range s.attached {
		if i > 0 {
			summary.Attach += ","
		}
		summary.Attach += a
	}
	if sha, err := ObjectSHA256(); err == nil {
		summary.ObjectSHA256 = sha
	}
	key := flowKey{}
	stat := flowStat{}
	it := s.flows.Iterate()
	for it.Next(&key, &stat) {
		summary.FlowsTotal++
		if stat.Samples > 0 {
			summary.FlowsSeen++
		}
		summary.RetransPayload += stat.RetransPayload
		summary.RetransEvents += stat.RetransEvents
		summary.QdiscDrops += stat.DropPackets
		summary.QdiscDropBytes += stat.DropBytes
		flow := Flow{
			Local:          net.IPv4(byte(key.Saddr), byte(key.Saddr>>8), byte(key.Saddr>>16), byte(key.Saddr>>24)).String(),
			Peer:           net.IPv4(byte(key.Daddr), byte(key.Daddr>>8), byte(key.Daddr>>16), byte(key.Daddr>>24)).String(),
			LocalPort:      key.Sport,
			PeerPort:       key.Dport,
			RetransPayload: stat.RetransPayload,
			RetransEvents:  stat.RetransEvents,
			DropPackets:    stat.DropPackets,
			DropBytes:      stat.DropBytes,
			SRTTus:         stat.SRTTUsec,
			SndCwnd:        stat.SndCWnd,
			SndSsthresh:    stat.SndSthresh,
			State:          stat.State,
			Samples:        stat.Samples,
		}
		if len(summary.Flows) < 50 {
			summary.Flows = append(summary.Flows, flow)
		}
	}
	if err := it.Err(); err != nil {
		return Summary{Available: false}, fmt.Errorf("iteration flux : %w", err)
	}
	// globals est PERCPU_ARRAY : un tableau par CPU, sommé ici.
	numCPU := 4
	if n, err := ebpf.PossibleCPU(); err == nil && n > 0 {
		numCPU = n
	}
	cpuVals := make([]globalStat, numCPU)
	var zero uint32
	if err := s.global.Lookup(&zero, &cpuVals); err == nil {
		for _, g := range cpuVals {
			summary.RetransPayload += g.RetransPayload
			summary.RetransEvents += g.RetransEvents
			summary.QdiscDrops += g.DropPackets
			summary.QdiscDropBytes += g.DropBytes
			summary.MapFull += g.MapFull
			summary.ReserveFailures += g.ReserveFailures
			// Le drop qdisc n'est compté par flux que si la table avalait le
			// flux ; la voie globale porte ceux-ci quand même.
		}
	}
	// histogramme latency_hist : PERCPU_ARRAY de 64 buckets, sommé par CPU.
	// Un objet précompilé sans cette map laisse s.hist nil : l'absence reste
	// un compteur à zéro, jamais un histogramme fabriqué. Échec de lecture d'un
	// bucket : ce bucket est omis, les autres restent mesurés.
	if s.hist != nil {
		cpuVals := make([]uint64, numCPU)
		for i := range summary.LatencyHistogram {
			key := uint32(i)
			if err := s.hist.Lookup(&key, &cpuVals); err != nil {
				continue
			}
			var bucket uint64
			for _, v := range cpuVals {
				bucket += v
			}
			summary.LatencyHistogram[i] = bucket
			summary.LatencySamples += bucket
		}
	}
	summary.QdiscAvailable = true
	return summary, nil
}
