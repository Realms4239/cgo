package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Realms4239/cgo/audit"
	"github.com/Realms4239/cgo/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/qdisc"
)

// runCalibrate — transforme une trace terrain (telemetry.csv, phase baseline
// seulement) en data/calibrations/<profil>.dist + sidecar scellé.
func runCalibrate(args []string) int {
	profile.Load()
	fs := flag.NewFlagSet("calibrate", flag.ExitOnError)
	auditID := fs.String("audit", "", "identifiant de l'audit source (data/audits/<id>)")
	profileID := fs.String("profile", "", "profil à calibrer (P2, P3, yas_4g…)")
	dataDir := fs.String("data", "data", "racine des artefacts d'audit")
	outDir := fs.String("out", "", "répertoire des .dist/.json (défaut <data>/calibrations)")
	_ = fs.Parse(args)
	if *auditID == "" || *profileID == "" {
		fmt.Fprintln(os.Stderr, "calibrate: --audit et --profile requis")
		return 2
	}
	if *outDir == "" {
		*outDir = filepath.Join(*dataDir, "calibrations")
	}

	model.ProfilesMu.RLock()
	prof, ok := model.Profiles[*profileID]
	model.ProfilesMu.RUnlock()
	if !ok {
		fmt.Fprintf(os.Stderr, "calibrate: profil %q inconnu\n", *profileID)
		return 2
	}

	rows, err := audit.ReadTelemetry(*dataDir, *auditID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	rtts := make([]float64, 0, len(rows))
	for _, r := range rows {
		if r.Phase == model.PhaseBaseline && r.RTTms > 0 {
			rtts = append(rtts, r.RTTms)
		}
	}
	cal, err := qdisc.BuildCalibration(prof, *auditID, rtts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}

	// provenance : la trace et son rapport d'audit, scellés s'ils existent.
	// L'identifiant vient de la ligne de commande : il passe par la même porte
	// ArtifactID que les écritures, sinon un -audit ../../chemin lit hors de la
	// zone d'audits et scelle un empreinte qui n'appartient à personne.
	telPath, err := audit.ArtifactFile(*dataDir, *auditID, audit.TelemetryArtifactName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	sum, err := fileSHA256(telPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	cal.SourceTelemetrySHA256 = sum

	repPath, err := audit.ArtifactFile(*dataDir, *auditID, audit.ReportArtifactName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	if raw, err := os.ReadFile(repPath); err == nil {
		reportSum := sha256.Sum256(raw)
		cal.SourceReportSHA256 = hex.EncodeToString(reportSum[:])
		var rep struct {
			Site     string `json:"site"`
			LinkType string `json:"link_type"`
		}
		if err := json.Unmarshal(raw, &rep); err != nil {
			fmt.Fprintln(os.Stderr, "calibrate: rapport illisible:", err)
			return 1
		}
		cal.SourceSite, cal.SourceLinkType = rep.Site, rep.LinkType
	} else if !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}

	cal.Created = time.Now().Format(time.RFC3339)
	if err := qdisc.WriteCalibration(*outDir, cal); err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	fmt.Printf("calibrate %s — %d échantillons baseline · plancher %.1f ms · échelle %.3f ms · p50 %.1f / p95 %.1f / p99.9 %.1f ms → %s\n",
		cal.Profile, cal.Samples, cal.FloorMs, cal.ScaleMs,
		cal.SourceP50RTTms, cal.SourceP95RTTms, cal.SourceP999RTTms,
		filepath.Join(*outDir, cal.DistFile))
	if cal.BelowFloorPct > 0 || cal.ClippedPct > 0 {
		fmt.Printf("attention: %.1f %% des RTT sous le plancher, %.1f %% de la table écrêtée — relire le sidecar\n", cal.BelowFloorPct, cal.ClippedPct)
	}
	return 0
}

func fileSHA256(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
