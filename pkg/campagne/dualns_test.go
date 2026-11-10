package campagne

import (
	"context"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
	"github.com/Realms4239/cgo/qdisc"
)

func TestBurstGEOnlyP2P3(t *testing.T) {
	// P2 ≈ 0,544 % et P3 ≈ 1,08 % de perte moyenne, en salves — les cibles
	// du banc durci, VÉRIFIÉES par la perte moyenne théorique, pas déclarées.
	p2, ok := BurstGE("P2")
	if !ok {
		t.Fatal("P2 doit porter un modèle en salves")
	}
	if m := qdisc.GEMeanLoss(p2); m < 0.5 || m > 0.6 {
		t.Fatalf("P2 : %v %% (attendu ≈0.544)", m)
	}
	p3, ok := BurstGE("P3")
	if !ok {
		t.Fatal("P3 doit porter un modèle en salves")
	}
	if m := qdisc.GEMeanLoss(p3); m < 1.0 || m > 1.15 {
		t.Fatalf("P3 : %v %% (attendu ≈1.08)", m)
	}
	for _, id := range []string{"P1", "P4", "P9", "", "p2-importé"} {
		if _, ok := BurstGE(id); ok {
			t.Fatalf("%q ne doit PAS offrir de mode rafales (P2/P3 seulement)", id)
		}
	}
}

func TestRunEventBurstyValidation(t *testing.T) {
	// bursty hors P2/P3 : erreur AVANT toute mutation qdisc — jamais de
	// repli silencieux en perte uniforme, l'identité de la cellule en jeu.
	f := &qdisc.FakeRunner{}
	d := Deps{
		TC: f, TCShaper: f, CliIf: "w-edge", ShaperIf: "e-srv",
		ChannelModel: "bursty",
		Ping:         func(context.Context, string, int) []float64 { return []float64{20} },
		Small:        func(context.Context) (float64, error) { return 0, context.Canceled },
	}
	ev := model.Event{Profile: "P1", Qdisc: "cake", CC: "bbr"}
	if _, err := RunEvent(context.Background(), ev, model.Profiles["P1"], d); err == nil {
		t.Fatal("bursty+P1 doit être refusé")
	} else if !strings.Contains(err.Error(), "P1") || !strings.Contains(err.Error(), "bursty") {
		t.Fatalf("l'erreur doit nommer le profil et la règle : %v", err)
	}
	if len(f.Calls) > 0 {
		t.Fatalf("aucune mutation qdisc avant la validation : %v", f.Calls)
	}
}

func TestApplyPhysicsAllIfaces(t *testing.T) {
	// La physique d'un profil s'applique à CliIf ET à chaque iface
	// supplémentaire (w-edge + c-wan dans le banc durci : les DEUX sens
	// voient le délai/la perte, sinon un run « symétrique » mesurerait un
	// lien asymétrique en secret).
	f := &qdisc.FakeRunner{}
	d := &Deps{TC: f, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}}
	ge := burstGEOrNull("P2", "bursty")
	if err := applyPhysics(d, model.Profiles["P2"], ge); err != nil {
		t.Fatal(err)
	}
	devs := map[string]bool{}
	for _, c := range f.Calls {
		for i, a := range c {
			if a == "dev" && i+1 < len(c) {
				devs[c[i+1]] = true
			}
		}
	}
	if !devs["w-edge"] || !devs["c-wan"] {
		t.Fatalf("les deux ifaces de physique doivent être posées : %v", devs)
	}
	// bursty : chaque pose porte gemodel, jamais « loss 0.5% ».
	for _, c := range f.Calls {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, "gemodel") == false && strings.Contains(joined, "netem") {
			t.Fatalf("pose sans gemodel en mode bursty : %s", joined)
		}
		if strings.Contains(joined, "loss 0.5%") {
			t.Fatalf("perte uniforme en mode bursty : %s", joined)
		}
	}
	// clean : la même boucle en uniforme.
	f2 := &qdisc.FakeRunner{}
	d2 := &Deps{TC: f2, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}}
	if err := applyPhysics(d2, model.Profiles["P2"], nil); err != nil {
		t.Fatal(err)
	}
	firstNetem := ""
	for _, c := range f2.Calls {
		if j := strings.Join(c, " "); strings.Contains(j, "netem") {
			firstNetem = j
			break
		}
	}
	if !strings.Contains(firstNetem, "loss 0.5%") {
		t.Fatalf("clean doit rester uniforme : %s", firstNetem)
	}
}

func TestProdDepsDualNSWiring(t *testing.T) {
	t.Setenv("CGO_SHAPE_MODE", "egress")
	up := ProdDepsDualNS("")
	if up.Bench != "dualns" || up.ChannelModel != "clean" {
		t.Fatalf("étiquettes : %+v", up)
	}
	if up.CliIf != "w-edge" || len(up.PhysicsIfaces) != 1 || up.PhysicsIfaces[0] != "c-wan" {
		t.Fatalf("physique : CliIf=%s Ifaces=%v", up.CliIf, up.PhysicsIfaces)
	}
	if up.ShaperIf != "e-srv" || up.Target != "10.30.0.2" || up.BulkAddr != "10.30.0.2:5201" {
		t.Fatalf("up : shaper=%s target=%s bulk=%s", up.ShaperIf, up.Target, up.BulkAddr)
	}
	down := ProdDepsDualNS("down")
	if down.ShaperIf != "e-wan" {
		t.Fatalf("down doit façonner e-wan : %s", down.ShaperIf)
	}
	if up.BulkN == nil {
		t.Fatal("BulkN doit être câblé (multi-flux, JFI)")
	}
	if up.ShapeMode != "egress" || up.IngressPreflight != nil {
		t.Fatalf("default path must stay egress: %+v", up)
	}
}

func TestProdDepsDualNSIngressWiring(t *testing.T) {
	t.Setenv("CGO_SHAPE_MODE", "ingress")
	t.Setenv("CGO_CLIENT_NS", "ns_test_client")
	t.Setenv("CGO_INGRESS_IF", "c-test")
	t.Setenv("CGO_INGRESS_DEV", "ifb9")

	d := ProdDepsDualNS("down")
	if d.ShapeMode != "ingress" {
		t.Fatalf("shape mode lost: %q", d.ShapeMode)
	}
	if d.ShaperIf != "ifb9" {
		t.Fatalf("shaper is not on the IFB: %q", d.ShaperIf)
	}
	if d.IngressPreflight == nil {
		t.Fatal("ingress preflight not wired")
	}
	if d.StatsFn == nil {
		t.Fatal("ifb0 stats not wired")
	}
	r, ok := d.TCShaper.(qdisc.NsRunner)
	if !ok || r.Ns != "ns_test_client" {
		t.Fatalf("shaper runner is not the client namespace: %#v", d.TCShaper)
	}
}

func writeDistCalibration(t *testing.T, dir, profile string, delay float64) {
	t.Helper()
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = delay + float64(i%8)
	}
	cal, err := qdisc.BuildCalibration(model.Profile{ID: profile, DelayMs: delay}, "audit-test", rtts)
	if err != nil {
		t.Fatal(err)
	}
	cal.SourceTelemetrySHA256 = "test"
	cal.Created = "2026-10-27T00:00:00Z"
	if err := qdisc.WriteCalibration(dir, cal); err != nil {
		t.Fatal(err)
	}
}

func TestApplyPhysicsDistMeasuredHop(t *testing.T) {
	dir := t.TempDir()
	writeDistCalibration(t, dir, "P2", 100)
	f := &qdisc.FakeRunner{}
	d := Deps{TC: f, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir}
	if err := applyPhysics(&d, model.Profiles["P2"], nil); err != nil {
		t.Fatal(err)
	}
	var up, down string
	for _, c := range f.Calls {
		joined := strings.Join(c, " ")
		if !strings.Contains(joined, "netem") {
			continue
		}
		if strings.Contains(joined, "dev w-edge") {
			up = joined
		}
		if strings.Contains(joined, "dev c-wan") {
			down = joined
		}
	}
	if !strings.Contains(up, "delay 50ms") || !strings.Contains(up, "distribution P2") {
		t.Fatalf("w-edge doit porter la table : %q", up)
	}
	if !strings.Contains(down, "delay 50ms 0ms") || strings.Contains(down, "distribution") {
		t.Fatalf("c-wan doit rester au plancher sans table : %q", down)
	}
}

func TestApplyPhysicsDistDownUsesDownlinkHop(t *testing.T) {
	dir := t.TempDir()
	writeDistCalibration(t, dir, "P2", 100)
	f := &qdisc.FakeRunner{}
	d := Deps{TC: f, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir, Direction: "down"}
	if err := applyPhysics(&d, model.Profiles["P2"], nil); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Calls {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, "dev c-wan") && strings.Contains(joined, "netem") && !strings.Contains(joined, "distribution P2") {
			t.Fatalf("c-wan mesuré doit porter la table : %q", joined)
		}
		if strings.Contains(joined, "dev w-edge") && strings.Contains(joined, "distribution") {
			t.Fatalf("w-edge non mesuré ne doit pas porter la table : %q", joined)
		}
	}
}

func TestApplyPhysicsDistFailsClosed(t *testing.T) {
	dir := t.TempDir()
	writeDistCalibration(t, dir, "P2", 100)
	stale := model.Profiles["P2"]
	stale.DelayMs = 120
	cases := []struct {
		name string
		deps Deps
		prof model.Profile
	}{
		{"calibration absente", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: t.TempDir()}, model.Profiles["P2"]},
		{"direction both", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir, Direction: "both"}, model.Profiles["P2"]},
		{"banc legacy", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "legacy", DelayModel: "dist", CalibrationDir: dir}, model.Profiles["P2"]},
		{"deux ifaces", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan", "c-wan2"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir}, model.Profiles["P2"]},
		{"plancher périmé", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir}, stale},
	}
	for _, tc := range cases {
		f := tc.deps.TC.(*qdisc.FakeRunner)
		if err := applyPhysics(&tc.deps, tc.prof, nil); err == nil {
			t.Fatalf("%s : erreur attendue", tc.name)
		}
		if len(f.Calls) != 0 {
			t.Fatalf("%s : tc appelé avant l'échec : %v", tc.name, f.Calls)
		}
	}
}

func TestProdDepsDualNSDistEnv(t *testing.T) {
	t.Setenv("CGO_DELAY_MODEL", "dist")
	t.Setenv("CGO_CALIBRATION_DIR", t.TempDir())
	d := ProdDepsDualNS("up")
	if d.DelayModel != "dist" || d.CalibrationDir == "" {
		t.Fatalf("dist non câblé : %+v", d)
	}
	ns, ok := d.TC.(qdisc.NsRunner)
	if !ok || len(ns.Env) != 1 || ns.Env[0] != "TC_LIB_DIR="+d.CalibrationDir {
		t.Fatalf("NsRunner.Env = %+v (calib %q)", d.TC, d.CalibrationDir)
	}
}

func TestProdDepsLegacyDistWired(t *testing.T) {
	t.Setenv("CGO_DELAY_MODEL", "dist")
	if d := ProdDeps(); d.DelayModel != "dist" {
		t.Fatalf("CGO_DELAY_MODEL ignoré par le banc legacy : %+v", d)
	}
}
