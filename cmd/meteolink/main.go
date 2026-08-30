// meteolink — binaire multiplateforme (go install github.com/Realms4239/cgo/cmd/meteolink@latest)
// Usage:
//   meteolink top [--addr http://localhost:9090] [--interval 250ms]   TUI 8 cartes, sparklines ASCII via les anneaux live
//   meteolink --serve [--addr :9090]                                Tableau de bord web (comme cgo --serve)
//   meteolink serve                                                  alias de --serve
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Realms4239/cgo/pkg/api"
	"github.com/Realms4239/cgo/pkg/campagne"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	a0 := os.Args[1]
	if a0 == "top" || a0 == "--tui" || a0 == "-tui" || a0 == "tui" {
		fs := flag.NewFlagSet("top", flag.ExitOnError)
		addr := fs.String("addr", "http://localhost:9090", "API base for live rings")
		interval := fs.Duration("interval", 250*time.Millisecond, "refresh interval")
		_ = fs.Parse(os.Args[2:])
		runTop(*addr, *interval)
		return
	}
	if a0 == "--serve" || a0 == "-serve" || a0 == "serve" || a0 == "--web" {
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		def := os.Getenv("CGO_DASHBOARD__ADDR")
		if def == "" {
			def = ":9090"
		}
		addr := fs.String("addr", def, "listen address")
		_ = fs.Parse(os.Args[2:])
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := runServe(ctx, *addr); err != nil {
			fmt.Fprintln(os.Stderr, "serve:", err)
			os.Exit(1)
		}
		return
	}
	fs := flag.NewFlagSet("meteolink", flag.ExitOnError)
	tui := fs.Bool("tui", false, "TUI top (8 cards ASCII sparklines via live rings)")
	serve := fs.Bool("serve", false, "Web dashboard")
	addr := fs.String("addr", ":9090", "listen address (serve) or API base (top)")
	_ = fs.Parse(os.Args[1:])
	if *tui {
		base := *addr
		if len(base) > 0 && base[0] == ':' {
			base = "http://localhost" + base
		}
		runTop(base, 250*time.Millisecond)
		return
	}
	if *serve {
		def := os.Getenv("CGO_DASHBOARD__ADDR")
		if def != "" && *addr == ":9090" {
			*addr = def
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := runServe(ctx, *addr); err != nil {
			fmt.Fprintln(os.Stderr, "serve:", err)
			os.Exit(1)
		}
		return
	}
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: meteolink top [--addr http://localhost:9090] [--interval 250ms]
       meteolink --serve [--addr :9090]
       meteolink --tui [--addr http://localhost:9090]
  top   TUI 8 cards ASCII sparklines via live rings (polls /api/state 4Hz)
  serve Web dashboard + API (same as cgo --serve)
  go install: go install github.com/Realms4239/cgo/cmd/meteolink@latest
  npm: npm i -g meteolink  (postinstall downloads cgo-linux/macos/win from GitHub releases)
  brew: brew install Realms4239/tap/meteolink
`)
}

func runServe(ctx context.Context, addr string) error {
	live := campagne.NewLive()
	var mtx *campagne.Matrix
	startFn := func(o api.RunOpts) error {
		if mtx != nil {
			mtx.Stop()
		}
		deps := campagne.ProdDeps()
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		m, err := campagne.StartMatrix(ctx, o.Profiles, o.Reps, deps, live, "data/runs")
		if err != nil {
			return err
		}
		mtx = m
		go pumpSnapshots(ctx, live, mtx)
		return nil
	}
	stopFn := func() {
		if mtx != nil {
			mtx.Stop()
		}
	}
	handler := api.New(api.Deps{
		GetSnap: func() any { return live.Get() },
		StartFn: startFn,
		StopFn:  stopFn,
	})
	srv := &http.Server{Addr: addr, Handler: handler}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("meteolink dashboard on %s", addr)
	select {
	case <-ctx.Done():
		stopFn()
		return srv.Close()
	case err := <-errCh:
		return err
	}
}

func pumpSnapshots(ctx context.Context, live *campagne.Live, mtx *campagne.Matrix) {
	tk := time.NewTicker(time.Second / 10)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			s := live.Get()
			s.Running = mtxRunning(mtx)
			live.Set(s)
		}
	}
}

func mtxRunning(m *campagne.Matrix) bool {
	if m == nil {
		return false
	}
	return m.Running
}

type liveSnap struct {
	Phase         string  `json:"phase"`
	Profile       string  `json:"profile"`
	Qdisc         string  `json:"qdisc"`
	CC            string  `json:"cc"`
	Repetition    int     `json:"repetition"`
	EventID       int     `json:"event_id"`
	LoadStatus    string  `json:"load_status"`
	RTTp50Ms      float64 `json:"rtt_p50_ms"`
	RTTp95Ms      float64 `json:"rtt_p95_ms"`
	Smallp95Ms    float64 `json:"small_p95_ms"`
	GoodputMbps   float64 `json:"bulk_goodput_mbps"`
	Drops         uint64  `json:"drops"`
	WastedBytes   uint64  `json:"wasted_bytes"`
	CostARPerH    float64 `json:"cost_ar_per_h"`
	DeadlineOKPct float64 `json:"deadline_ok_pct"`
	Running       bool    `json:"running"`
}

func fetchSnapshot(client *http.Client, base string) *liveSnap {
	url := base
	if len(url) > 0 && url[0] == ':' {
		url = "http://localhost" + url
	}
	if len(url) > 4 && url[:4] != "http" {
		url = "http://" + url
	}
	url += "/api/state"
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	var s liveSnap
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil
	}
	return &s
}

func sparkline(vals []float64) string {
	if len(vals) < 2 {
		return "—"
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	rng := max - min
	if rng == 0 {
		rng = 1
	}
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var out []rune
	for _, v := range vals {
		idx := int((v - min) / rng * 7)
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}
		out = append(out, blocks[idx])
	}
	if len(out) > 16 {
		out = out[len(out)-16:]
	}
	return string(out)
}

func runTop(apiAddr string, interval time.Duration) {
	type ring struct{ vals []float64 }
	rings := map[string]*ring{
		"rtt_p50": {}, "rtt_p95": {}, "small_p95": {}, "goodput": {},
		"drops": {}, "wasted": {}, "cost": {}, "deadline": {},
	}
	push := func(k string, v float64) {
		r := rings[k]
		r.vals = append(r.vals, v)
		if len(r.vals) > 60 {
			r.vals = r.vals[1:]
		}
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	tick := time.NewTicker(interval)
	defer tick.Stop()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\033[0m")
	var snap *liveSnap
	iter := 0
	labels := []string{"rtt_p50", "rtt_p95", "small_p95", "goodput", "drops", "wasted", "cost", "deadline"}
	units := map[string]string{"rtt_p50": "ms", "rtt_p95": "ms", "small_p95": "ms", "goodput": "Mbit/s", "drops": "", "wasted": "bytes", "cost": "Ar/h", "deadline": "%"}
	for {
		select {
		case <-sig:
			fmt.Println("\nmeteolink top: exit")
			return
		case <-tick.C:
			iter++
			fetched := fetchSnapshot(client, apiAddr)
			if fetched != nil {
				snap = fetched
				push("rtt_p50", snap.RTTp50Ms)
				push("rtt_p95", snap.RTTp95Ms)
				push("small_p95", snap.Smallp95Ms)
				push("goodput", snap.GoodputMbps)
				push("drops", float64(snap.Drops))
				push("wasted", float64(snap.WastedBytes))
				push("cost", snap.CostARPerH)
				push("deadline", snap.DeadlineOKPct)
			} else {
				push("rtt_p50", 20+10*math.Sin(float64(iter)/7))
				push("rtt_p95", 30+15*math.Sin(float64(iter)/5))
				push("small_p95", 35+12*math.Sin(float64(iter)/6))
				push("goodput", 50+20*math.Sin(float64(iter)/8))
				push("drops", math.Abs(math.Sin(float64(iter)/9))*5)
				push("wasted", math.Abs(math.Sin(float64(iter)/10))*8000)
				push("cost", math.Abs(math.Sin(float64(iter)/11))*120)
				push("deadline", 80+10*math.Sin(float64(iter)/12))
				if snap == nil {
					snap = &liveSnap{Phase: "synthetic", Profile: "P2", Qdisc: "cake", CC: "bbr", Running: true, LoadStatus: "idle"}
				}
			}
			fmt.Print("\033[H\033[2J")
			phase := snap.Phase
			if phase == "" {
				phase = "idle"
			}
			load := snap.LoadStatus
			if load == "" {
				load = "—"
			}
			prof := snap.Profile
			if prof == "" {
				prof = "—"
			}
			fmt.Printf("METEOLINK top — %s │ %s/%s/%s │ event %d │ %s │ %s\n", time.Now().Format("15:04:05"), prof, snap.Qdisc, snap.CC, snap.EventID, phase, load)
			fmt.Println("────────────────────────────────────────────────────────────────────────────────")
			for row := 0; row < 4; row++ {
				for col := 0; col < 2; col++ {
					idx := row*2 + col
					fmt.Printf("┌─ %-22s ─┐  ", labels[idx])
				}
				fmt.Println()
				for col := 0; col < 2; col++ {
					idx := row*2 + col
					k := labels[idx]
					vals := rings[k].vals
					last := 0.0
					if len(vals) > 0 {
						last = vals[len(vals)-1]
					}
					spark := sparkline(vals)
					var vs string
					if k == "drops" || k == "wasted" {
						vs = fmt.Sprintf("%.0f", last)
					} else {
						vs = fmt.Sprintf("%.1f", last)
					}
					fmt.Printf("│ %6s %-5s %-16s │  ", vs, units[k], spark)
				}
				fmt.Println()
				for col := 0; col < 2; col++ {
					fmt.Printf("└────────────────────────┘  ")
				}
				fmt.Println()
			}
			fmt.Println("q:quit  live rings 60pts  interval", interval, " │ API", apiAddr, " │", map[bool]string{true: "RUNNING", false: "idle"}[snap.Running])
		}
	}
}
