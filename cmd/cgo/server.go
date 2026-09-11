package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/api"
	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/pkg/doctor"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/probe"
	"github.com/Realms4239/cgo/pkg/qdisc"
	"github.com/Realms4239/cgo/pkg/results"
)

// runServer démarre le tableau de bord + l'API avec le cœur de campagne.
// mode: "full" (Linux bench/edge) or "observe" (Windows workstation).
// tlsOn: HTTPS avec le certificat local meteolink.dev (défaut) ; false =
// HTTP brut (VM, systemd). httpAddr: redirection HTTP→HTTPS ("" = off).
func runServer(ctx context.Context, addr, mode string, tlsOn bool, httpAddr string) error {
	// runs gelés embarqués : premier lancement sans preuves → restauration
	// best-effort (jamais fatale : un seed illisible n'empêche pas de servir).
	if seeded, n, err := results.SeedRuns("."); err != nil {
		log.Printf("runs gelés : seed impossible (%v) — démarrage sans restauration", err)
	} else if seeded {
		log.Printf("runs gelés intégrés : %d fichiers", n)
	}
	live := campagne.NewLive()

	// deadline par défaut — même registre que GET /api/schema
	defaultDeadline := func() float64 {
		for _, p := range api.SchemaParams() {
			if p.Key == "deadline_ms" {
				if v, ok := p.Default.(float64); ok {
					return v
				}
				if v, ok := p.Default.(int); ok {
					return float64(v)
				}
			}
		}
		return 1000
	}

	// les cellules en quarantaine vont dans le journal opérateur
	campagne.OnQuarantine = func(runID string, eventID int, profile, qdisc, cc string) {
		api.RecordEvent("quarantaine", fmt.Sprintf("%s évènement %d %s/%s/%s — cellule invalidée par les portes", runID, eventID, profile, qdisc, cc))
	}
	// les cellules skippées aussi — un trou dans le gel doit être
	// explicable après coup (reprise possible, raison journalisée)
	campagne.OnSkip = func(runID string, eventID int, profile, qdisc, cc string) {
		api.RecordEvent("skip", fmt.Sprintf("%s évènement %d %s/%s/%s — cellule coupée par l'opérateur, rejouable à la reprise", runID, eventID, profile, qdisc, cc))
	}
	// sentinelle deadline — la surveillance alerter honnêtement quand le lien
	// tenu hors campagne dépasse l'objectif small p95 de façon soutenue
	campagne.OnWatchAlert = func(p95, deadline float64) {
		api.RecordEvent("alerte", fmt.Sprintf("deadline dépassée en production — small p95 %.1f ms > %.0f ms", p95, deadline))
	}

	if mode == "full" {
		// startup self-check: a crashed session can leave a root qdisc
		// derrière; reset-puis-application repart d'un shaper propre.
		deps := campagne.ProdDeps()
		_, _ = deps.TCShaper.Run("qdisc", "del", "dev", deps.ShaperIf, "root")
	}

	var mtx *campagne.Matrix
	var mtxMu sync.Mutex
	getMtx := func() *campagne.Matrix {
		mtxMu.Lock()
		defer mtxMu.Unlock()
		return mtx
	}
	setMtx := func(m *campagne.Matrix) {
		mtxMu.Lock()
		mtx = m
		mtxMu.Unlock()
	}

	// Boucle de diffusion unique pour toute la vie du serveur:
	// lit le mtx courant par getMtx à chaque tick, pas de fuite, pas de flottement.
	go pumpSnapshots(ctx, live, getMtx)

	// watchCtl — bascule de surveillance partagée entre le chemin campagne
	// (startFn arrête la surveillance) et le chemin API (POST /api/watch). Les
	// deux mutent le même état; un mutex, sinon course de données (un bool
	// nu courait entre les deux chemins).
	watchCtl := struct {
		mu   sync.Mutex
		stop func()
		on   bool
	}{stop: func() {}}
	watchStart := func() error {
		watchCtl.mu.Lock()
		defer watchCtl.mu.Unlock()
		if watchCtl.on {
			return nil
		}
		deps := campagne.ProdDeps()
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		// la sentinelle connaît l'objectif courant (paramètre partagé campagne/surveillance)
		deps.DeadlineMs = defaultDeadline()
		stop := campagne.StartWatch(ctx, deps)
		watchCtl.stop = stop
		watchCtl.on = true
		return nil
	}
	watchStop := func() {
		watchCtl.mu.Lock()
		defer watchCtl.mu.Unlock()
		if watchCtl.on {
			watchCtl.stop()
			watchCtl.on = false
		}
	}
	startFn := func(o api.RunOpts) error {
		watchStop() // campagne and surveillance are mutually exclusive
		if m := getMtx(); m != nil {
			m.Stop()
			// drainage — les cellules de l'ancienne matrice peuvent être en plein appel tc;
			// démarrer par-dessus fait courir le shaper et casse les nouvelles cellules
			for i := 0; i < 30 && m.IsRunning(); i++ {
				time.Sleep(100 * time.Millisecond)
			}
		}
		deps := campagne.ProdDeps()
		if o.Direction == "down" {
			deps = campagne.ProdDepsDown()
		}
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		// la deadline choisie par l'opérateur voyage avec la campagne —
		// reprise du défaut registre si non fournie (0 interdit : 0% partout)
		if o.DeadlineMs > 0 {
			deps.DeadlineMs = float64(o.DeadlineMs)
		} else {
			deps.DeadlineMs = defaultDeadline()
		}
		// porte plan de mesure — sans banc (VM fraîche : pas de veth, pas de
		// testbedsrv, pas de sudoers), chaque cellule échoue et la matrice
		// gèle un run à ZÉRO ligne (« Démarrer → idle → rien », vu en prod).
		// Refus honnête (409) plutôt que run fantôme.
		if err := campagne.PlaneReady(ctx, deps); err != nil {
			return err
		}
		// axes : vides = matrice pleine ; ciblés = sous-matrice (cellules n=1
		// rejouables sans 54 min de matrice complète)
		qdiscs, ccs := o.Qdiscs, o.CCs
		if len(qdiscs) == 0 {
			qs := make([]string, len(model.AllQdiscs))
			for i, q := range model.AllQdiscs {
				qs[i] = string(q)
			}
			qdiscs = qs
		}
		if len(ccs) == 0 {
			cs := make([]string, len(model.AllCC))
			for i, c := range model.AllCC {
				cs[i] = string(c)
			}
			ccs = cs
		}
		// Reprise : un run_id existant rejoue les cellules manquantes au lieu
		// de geler un doublon (PC/VM éteints en pleine matrice). Vide = frais.
		var m *campagne.Matrix
		var err error
		if o.RunID != "" {
			m, err = campagne.StartMatrixFilteredWithID(ctx, o.RunID, o.Profiles, qdiscs, ccs, o.Reps, deps, "data/runs")
		} else {
			m, err = campagne.StartMatrixFiltered(ctx, o.Profiles, qdiscs, ccs, o.Reps, deps, "data/runs")
		}
		if err != nil {
			return err
		}
		setMtx(m)
		// RRUL (both) : l'up gélée, enchaîner la matrice down — les deux sens
		// dans le même run pour la comparaison directe (up puis down
		// séquentiels : le banc n'a qu'un shaper actif à la fois par iface)
		if o.Direction == "both" {
			go func(up *campagne.Matrix) {
				for i := 0; i < 3600 && up.IsRunning(); i++ {
					time.Sleep(300 * time.Millisecond)
				}
				if up.IsRunning() || ctx.Err() != nil {
					return
				}
				depsDown := campagne.ProdDepsDown()
				depsDown.DeadlineMs = deps.DeadlineMs
				depsDown.OnSnap = deps.OnSnap
				if m2, err := campagne.StartMatrixFiltered(ctx, o.Profiles, qdiscs, ccs, o.Reps, depsDown, "data/runs"); err == nil {
					setMtx(m2)
				}
			}(m)
		}
		return nil
	}
	stopFn := func() {
		if getMtx() != nil {
			getMtx().Stop()
		}
		// pump will notice mtx.IsRunning()==false on next tick (≤100ms) and
		// passe live à false atomiquement; réglé aussi tout de suite pour la réactivité.
		live.SetRunning(false)
	}
	// Skip — coupe la cellule en cours sans arrêter la matrice
	skipFn := func() {
		if m := getMtx(); m != nil {
			m.Skip()
		}
	}
	// Levier de façonnage sur la sortie du bord (contrôle manuel):
	// mêmes primitives que les cellules de campagne (mode manuel pour la DSI).
	shapeFn := func(r api.ShapeReq) error {
		deps := campagne.ProdDeps()
		// reset d'abord — des handles périmés d'une cellule font échouer 'replace';
		// le levier possède un état propre déterministe
		_, _ = deps.TCShaper.Run("qdisc", "del", "dev", deps.ShaperIf, "root")
		if r.Qdisc == "none" {
			return nil
		}
		// Conditions du lien: netem à la racine, shaper empilé en enfant.
		if r.DelayMs > 0 || r.JitterMs > 0 || r.LossPct > 0 {
			if err := qdisc.ApplyNetem(deps.TCShaper, deps.ShaperIf, r.DelayMs, r.JitterMs, r.LossPct); err != nil {
				return err
			}
		}
		return qdisc.ApplyShaper(deps.TCShaper, deps.ShaperIf, model.Qdisc(r.Qdisc), r.CapMbps, 100)
	}
	handler := api.New(api.Deps{
		GetSnap:   func() any { return live.Get() },
		StartFn:   startFn,
		StopFn:    stopFn,
		SkipFn:    skipFn,
		ShapeFn:   shapeFn,
		RunningFn: func() bool { return getMtx() != nil && getMtx().IsRunning() },
		Mode:      mode,
		Version:   version,
		DoctorFn: func() any {
			m, checks := doctor.Report(mode)
			return map[string]any{"mode": m, "checks": checks}
		},
		BurstFn: func(cc string, seconds int) error {
			deps := campagne.ProdDeps()
			ctxB, cancel := context.WithTimeout(ctx, time.Duration(seconds+3)*time.Second)
			defer cancel()
			b, err := probe.BulkSendTo(ctxB, deps.BulkAddr, cc)
			if err != nil && b == 0 {
				return err
			}
			g := float64(b) * 8 / 1e6 / float64(seconds)
			if g > 2500 { // same ring clamp as the campagne cells
				g = 2500
			}
			live.Set(campagne.Snapshot{
				Phase: "burst", Profile: "burst", CC: cc,
				GoodputMbps: math.Round(g*10) / 10,
				Drops:       0, Running: false,
			})
			return nil
		},
		WatchFn: func(on bool) error {
			if on {
				if err := watchStart(); err != nil {
					return err
				}
			} else {
				watchStop()
			}
			return nil
		},
	})
	srv := &http.Server{
		Addr: addr, Handler: handler,
		// Pas de WriteTimeout (SSE + replay tiennent des connexions longues),
		// mais les en-têtes lentes ne doivent pas pendre un slot (Slowloris).
		ReadHeaderTimeout: 5 * time.Second,
	}
	var extra []*http.Server // redirection HTTP : fermés avec le principal
	errCh := make(chan error, 1)
	if tlsOn {
		certFile, keyFile, err := ensureLocalCert()
		if err != nil {
			return err
		}
		// Listen explicite (pas ListenAndServeTLS direct) : en cas d'échec
		// sur meteolink.dev, on explique le hosts au lieu d'une erreur brute.
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			hint := ""
			if host, _, herr := net.SplitHostPort(addr); herr == nil && host == "meteolink.dev" {
				hint = " — ajoutez « 127.0.0.1 meteolink.dev » au fichier hosts (cgo setup le propose)"
			}
			return fmt.Errorf("écoute %q impossible%s : %w", addr, hint, err)
		}
		if httpAddr != "" {
			redir := &http.Server{Addr: httpAddr, Handler: redirectHTTPS(addr), ReadHeaderTimeout: 5 * time.Second}
			extra = append(extra, redir)
			go func() {
				if err := redir.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()
			log.Printf("cgo redirect http://%s → https (meteolink.dev)", httpAddr)
		}
		go func() { errCh <- srv.ServeTLS(ln, certFile, keyFile) }()
		log.Printf("cgo dashboard on https://%s (certificat local meteolink.dev)", addr)
	} else {
		go func() { errCh <- srv.ListenAndServe() }()
		log.Printf("cgo dashboard on http://%s (HTTP brut, VM/systemd)", addr)
	}
	select {
	case <-ctx.Done():
		stopFn()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		for _, s := range extra {
			_ = s.Shutdown(shutCtx)
		}
		handler.CloseHub() // stop the 10 Hz broadcast ticker — Server.Close never does
		return nil
	case err := <-errCh:
		return err
	}
}

// redirectHTTPS renvoie tout vers https://meteolink.dev:<port> + chemin.
// Le nom canonique plutôt que l'hôte demandé : un seul nom, un seul cert.
func redirectHTTPS(tlsAddr string) http.Handler {
	port := "9090"
	if _, p, err := net.SplitHostPort(tlsAddr); err == nil && p != "" {
		port = p
	}
	target := "https://meteolink.dev:" + port
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target+r.URL.RequestURI(), http.StatusTemporaryRedirect)
	})
}

// pumpSnapshots reflète la progression de la matrice dans l'instantané diffus à 10 Hz.
// Uses Live.SetRunning atomically to avoid Get+Modify+Set lost-update race (H2).
// getMtx lit le pointeur de matrice sous verrou (pompe unique).
func pumpSnapshots(ctx context.Context, live *campagne.Live, getMtx func() *campagne.Matrix) {
	tk := time.NewTicker(time.Second / 10)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			live.SetRunning(mtxRunning(getMtx()))
		}
	}
}

func mtxRunning(m *campagne.Matrix) bool { return m.IsRunning() }
