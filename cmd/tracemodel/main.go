package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

const minSamples = 30

var (
	flowsHeader = []string{
		"kind", "window_unix", "local_label", "peer_label", "local_port", "peer_port",
		"retrans_payload_bytes", "retrans_events", "qdisc_drop_packets", "qdisc_drop_bytes",
		"srtt_us", "snd_cwnd", "snd_ssthresh", "state", "samples",
	}
	profileOrder = []string{"P1", "P2", "P3", "P4"}
	floors       = map[string]float64{"P1": 20, "P2": 100, "P3": 600, "P4": 40}
)

// readFlows — lecture stricte du header scellé. Un srtt absent ou invalide
// devient NaN : il sera ignoré, jamais transformé en 0 ms mesuré.
func readFlows(path string) ([]FlowRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = len(flowsHeader)
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("tracemodel: en-tête %s: %w", path, err)
	}
	if !slices.Equal(header, flowsHeader) {
		return nil, fmt.Errorf("tracemodel: en-tête %s = %v, attendu %v", path, header, flowsHeader)
	}
	srttIdx := slices.Index(header, "srtt_us")

	var rows []FlowRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tracemodel: ligne %s: %w", path, err)
		}
		srtt := math.NaN()
		if raw := strings.TrimSpace(rec[srttIdx]); raw != "" {
			if v, e := strconv.ParseFloat(raw, 64); e == nil {
				srtt = v
			}
		}
		rows = append(rows, FlowRow{SRTTus: srtt})
	}
	return rows, nil
}

func main() {
	var (
		flowsPath = flag.String("flows", "", "sealed flows.csv to read")
		outDir    = flag.String("out", "", "verification directory for audit tree")
		traceID   = flag.String("trace-id", "", "trace identifier for provenance")
	)
	flag.Parse()
	if *flowsPath == "" || *outDir == "" || *traceID == "" {
		fmt.Fprintln(os.Stderr, "tracemodel: --flows, --out and --trace-id are required")
		os.Exit(2)
	}

	rows, err := readFlows(*flowsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tracemodel:", err)
		os.Exit(1)
	}
	buckets := BucketFlows(rows, floors)
	for _, p := range profileOrder {
		samples := buckets[p]
		if err := ValidateBand(samples, minSamples, p); err != nil {
			fmt.Println("absent:", err)
			continue
		}
		sum, err := WriteTelemetry(*outDir, *traceID, p, samples)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tracemodel:", err)
			os.Exit(1)
		}
		fmt.Printf("present: %s has %d samples · telemetry sha256 %s\n", p, len(samples), sum)
	}
}
