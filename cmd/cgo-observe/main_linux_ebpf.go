//go:build linux && ebpf_object

// cgo-observe — collecteur passif borné : attache les tracepoints eBPF du
// banc épinglé, écrit des lignes pseudonymisées par fenêtre, s'arrête à
// l'échéance, détache, scelle. Aucune mesure synthétique, aucun trafic
// injecté, aucun état host persistant au-delà du processus.
//
// Compteurs cumulatifs du binaire noyau : les débits par fenêtre sont les
// deltas d'une ligne à l'autre (analyse soustrait, jamais relègue).
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Realms4239/cgo/probe/bpf"
)

const usage = `cgo-observe — trace passive bornée
Usage: cgo-observe --out DIR [--duration SEC] [--interval SEC]
Plafond dur : 172800 s (48 h). -tags ebpf_object requis.`

func main() {
	out := flag.String("out", "", "répertoire de trace (créé 0700)")
	duration := flag.Int("duration", 60, "secondes (plafond dur 172800)")
	interval := flag.Int("interval", 5, "secondes entre deux exports")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if *interval < 1 {
		fmt.Fprintln(os.Stderr, "intervalle < 1 s refusé")
		os.Exit(2)
	}
	dur := clampDuration(*duration)

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		log.Fatalf("sel: %v", err)
	}
	if err := os.MkdirAll(*out, 0o700); err != nil {
		log.Fatalf("out: %v", err)
	}
	_ = os.Chmod(*out, 0o700)
	csvPath := filepath.Join(*out, "flows.csv")
	f, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatalf("csv: %v", err)
	}
	fmt.Fprintln(f, "kind,window_unix,local_label,peer_label,retrans_payload_bytes,retrans_events,qdisc_drop_packets,qdisc_drop_bytes,srtt_us,snd_cwnd,snd_ssthresh,state,samples")

	sess, err := bpf.Attach()
	if err != nil {
		log.Fatalf("attache eBPF: %v", err)
	}
	log.Printf("attaché: %s — trace jusqu'à %s", strings.Join(sess.Attached(), ","), time.Now().Add(time.Duration(dur)*time.Second).UTC().Format(time.RFC3339))

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	deadline := time.After(time.Duration(dur) * time.Second)

	rows := uint64(0)
	for {
		select {
		case <-deadline:
			goto done
		case s := <-sig:
			log.Printf("signal %s — arrêt gracieux", s)
			goto done
		default:
		}
		sum, err := sess.Summary()
		window := time.Now().Unix()
		if err == nil {
			g(f, window, sum, &rows)
			for _, fl := range sum.Flows {
				writeFlow(f, window, fl, salt, &rows)
			}
		} else {
			log.Printf("synthèse: %v", err)
		}
		_ = f.Sync()
		time.Sleep(time.Duration(*interval) * time.Second)
	}
done:
	sess.Close()
	_ = f.Sync()
	_ = f.Close()

	data, _ := os.ReadFile(csvPath)
	sum := sha256.Sum256(data)
	manifest := map[string]any{
		"kind":             "cgo-observe/manifest",
		"started_unix":     time.Now().Unix(),
		"duration_s":       dur,
		"interval_s":       *interval,
		"rows":             rows,
		"attach":           sess.Attached(),
		"flows_csv_sha256": hex.EncodeToString(sum[:]),
	}
	if sha, err := bpf.ObjectSHA256(); err == nil {
		manifest["object_sha256"] = sha
	}
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(filepath.Join(*out, "manifest.json"), mb, 0o600)
	log.Printf("arrêt propre — lignes=%d manifest=%s", rows, filepath.Join(*out, "manifest.json"))
}

func g(f *os.File, window int64, sum bpf.Summary, rows *uint64) {
	fmt.Fprintf(f, "G,%d,-,-,%d,%d,%d,%d,-,-,-,-,-\n",
		window, sum.RetransPayload, sum.RetransEvents, sum.QdiscDrops, sum.QdiscDropBytes)
	*rows++
}

func writeFlow(f *os.File, window int64, fl bpf.Flow, salt []byte, rows *uint64) {
	fmt.Fprintf(f, "F,%d,%s,%s,%d,%d,%d,%d,%d,%d,%d,%d,%d\n",
		window, pseudonym(fl.Local, salt), pseudonym(fl.Peer, salt),
		fl.RetransPayload, fl.RetransEvents, fl.DropPackets, fl.DropBytes,
		fl.SRTTus, fl.SndCwnd, fl.SndSsthresh, fl.State, fl.Samples)
	*rows++
}
