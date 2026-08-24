package qdisc

import (
	"os/exec"
	"strconv"
	"strings"
)

// Stats holds per-qdisc lifetime counters parsed from `tc -s qdisc show`.
type Stats struct {
	Kind       string
	Handle     string
	Bytes      uint64
	Packets    uint64
	Drops      uint64
	Overlimits uint64
	Backlog    uint64
}

// PollStats parses `tc -s qdisc show dev <iface>` for all non-default qdiscs.
// Works on both main-ns (veth-c) and netns (veth-s) via the TCRunner seam.
func PollStats(r TCRunner, iface string) ([]Stats, error) {
	out, err := r.Run("-s", "qdisc", "show", "dev", iface)
	if err != nil {
		return nil, err
	}
	return parseStats(string(out)), nil
}

// parseStats extracts qdisc kind, handle, bytes, packets, drops, overlimits.
// `tc -s` output format:
//
//	qdisc netem 1: root refcnt 129 limit 1000 delay 20ms 2ms
//	 Sent 9596383791 bytes 6351949 pkt (dropped 281, overlimits 0 requeues 0)
//	 backlog 0b 0p requeues 0
func parseStats(out string) []Stats {
	var stats []Stats
	var cur *Stats
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "qdisc ") {
			if cur != nil {
				stats = append(stats, *cur)
			}
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			cur = &Stats{Kind: fields[1], Handle: fields[2]}
		} else if cur != nil && strings.HasPrefix(line, "Sent ") {
			cur.Bytes = parseU64(line, "Sent ", " bytes")
			cur.Packets = parseU64(line, "bytes ", " pkt")
			cur.Drops = parseU64(line, "dropped ", ",")
			cur.Overlimits = parseU64(line, "overlimits ", " requeues")
		} else if cur != nil && strings.HasPrefix(line, "backlog ") {
			bl := strings.Fields(line)
			if len(bl) >= 2 {
				cur.Backlog = parseU64("backlog " + bl[1], "backlog ", "b")
			}
		}
	}
	if cur != nil {
		stats = append(stats, *cur)
	}
	return stats
}

func parseU64(line, prefix, suffix string) uint64 {
	i := strings.Index(line, prefix)
	if i < 0 {
		return 0
	}
	rest := line[i+len(prefix):]
	j := strings.Index(rest, suffix)
	if j < 0 {
		return 0
	}
	v, err := strconv.ParseUint(strings.TrimSpace(rest[:j]), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// SumDrops sums drops across all non-root qdiscs for a delta measurement.
func SumDrops(stats []Stats) uint64 {
	var total uint64
	for _, s := range stats {
		total += s.Drops
	}
	return total
}

// SumBytes sums bytes across all qdiscs (for goodput delta measurement).
func SumBytes(stats []Stats) uint64 {
	var total uint64
	for _, s := range stats {
		total += s.Bytes
	}
	return total
}

// execCmdRunner is a helper for tests.
type execCmdRunner struct{}

func (execCmdRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("tc", args...).CombinedOutput()
}
