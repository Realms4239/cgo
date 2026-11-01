package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const auditFixture = `{"audit_id":"audit-ok","site":"siège","link_type":"fibre","bloat_grade":"B",
"operator_verdict":{"level":"green","word":"Lien sain","sentence":"Sous charge, la latence reste sous l'objectif"},
"measured":{"rtt_idle_p50_ms":22.0,"rtt_idle_p95_ms":24.0,"rtt_loaded_p50_ms":40.0,"rtt_loaded_p95_ms":60.0,
"bloat_delta_ms":36.0,"qdi_ms":12.0,"throughput_mbps":18.5,"loss_pct":0,"load_attempted":true,
"load_measured":true,"direction":"up","flow_count":4},
"provenance":{"source_sha256":"aaaaaaaaaaaa1111","telemetry_sha256":"bbbbbbbbbbbb2222"},
"remediation":{"qdisc":"cake","reason":"meilleur classement mesuré",
"linux_tc":["tc qdisc add dev eth0 root cake bandwidth 5mbit","tc -s qdisc show dev eth0"],
"routeros":"/queue type add name=cgo-cake kind=cake\n/queue tree add name=cgo-up parent=global max-limit=5M"}}`

func statusServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func pollOnce(t *testing.T, api string) auditView {
	t.Helper()
	msg, ok := pollAudit(api)().(auditMsg)
	if !ok {
		t.Fatal("pollAudit ne rend pas un auditMsg")
	}
	return msg.v
}

// Un rapport absent n'est pas un échec : le TUI doit dire « aucun audit »
// plutôt que d'inventer un lien vierge.
func TestAuditViewNoneIsNotAFailure(t *testing.T) {
	srv := statusServer(t, `{"running":false,"last":null,"report":null}`)
	v := pollOnce(t, srv.URL)
	if v.OK {
		t.Fatal("rapport null traité comme un audit")
	}
	if !strings.Contains(v.Why, "aucun audit") {
		t.Fatalf("raison = %q, voulu « aucun audit »", v.Why)
	}
}

// Une API injoignable n'est pas un audit absent : deux causes, deux phrases.
func TestAuditViewTransportFailureNamesTheAPI(t *testing.T) {
	srv := statusServer(t, `{}`)
	api := srv.URL
	srv.Close()
	v := pollOnce(t, api)
	if v.OK {
		t.Fatal("échec transport traité comme un audit")
	}
	if !strings.Contains(v.Why, "injoignable") {
		t.Fatalf("raison = %q, voulu « API injoignable »", v.Why)
	}
	if strings.Contains(v.Why, "aucun audit") {
		t.Fatalf("un échec réseau se lit comme une absence d'audit : %q", v.Why)
	}
}

func TestAuditViewDecodesVerdictAndFacts(t *testing.T) {
	srv := statusServer(t, `{"running":false,"report":`+auditFixture+`}`)
	v := pollOnce(t, srv.URL)
	if !v.OK {
		t.Fatalf("audit valide refusé : %q", v.Why)
	}
	if v.AuditID != "audit-ok" || v.Verdict.Word != "Lien sain" {
		t.Fatalf("décodage = %+v", v)
	}
	if !strings.Contains(verdictLine(v), "Lien sain") {
		t.Fatalf("entête = %q", verdictLine(v))
	}
	facts := auditFactLines(v)
	if !strings.Contains(facts[0], "22.0") || !strings.Contains(facts[0], "24.0") {
		t.Fatalf("repos = %q", facts[0])
	}
	if !strings.Contains(facts[1], "objectif 220 ms") || !strings.Contains(facts[1], "36.0") {
		t.Fatalf("sous charge = %q", facts[1])
	}
	if !strings.Contains(facts[2], "18.5") {
		t.Fatalf("débit = %q", facts[2])
	}
	// perte 0 % présente : c'est une mesure, pas une absence.
	if !strings.Contains(facts[3], "0.00") {
		t.Fatalf("perte = %q, voulu le 0 %% mesuré", facts[3])
	}
}

// Aucun 0 inventé : un rapport vide ne fabrique ni latence ni débit.
func TestAuditFactLinesNeverInventAValue(t *testing.T) {
	v := auditView{OK: true}
	for _, f := range auditFactLines(v) {
		if strings.Contains(f, " 0.0") || strings.Contains(f, ": 0 ") {
			t.Fatalf("un zéro fabriqué : %q", f)
		}
		if !strings.Contains(f, "non mesuré") && !strings.Contains(f, "non tentée") {
			t.Fatalf("fait sans mesure annoncée comme une valeur : %q", f)
		}
	}
}

// La charge tentée mais jamais établie a sa propre phrase (cascade §3.2).
func TestAuditFactLinesNameChargeNonEtablie(t *testing.T) {
	v := auditView{OK: true, Measured: measuredView{LoadAttempted: true, LoadMeasured: false}}
	facts := auditFactLines(v)
	if !strings.Contains(facts[1], "non établie") {
		t.Fatalf("sous charge = %q", facts[1])
	}
}

func TestVerdictHeaderUsesTheReportWord(t *testing.T) {
	v := auditView{OK: true}
	if !strings.Contains(verdictLine(v), "indisponible") {
		t.Fatalf("rapport sans cascade = %q", verdictLine(v))
	}
	v.Verdict.Word = "Lien fragile"
	if !strings.Contains(verdictLine(v), "Lien fragile") {
		t.Fatalf("entête = %q", verdictLine(v))
	}
}

func TestWrap78NeverExceedsTheColumn(t *testing.T) {
	long := strings.Repeat("abcdefghij", 20) // 200 runes, un seul mot
	lines := wrap78(long)
	for _, l := range lines {
		if len([]rune(l)) > paneWidth {
			t.Fatalf("ligne de %d colonnes : %q", len([]rune(l)), l)
		}
	}
	if strings.Join(lines, "") != long {
		t.Fatal("le repli d'un mot long perd des caractères")
	}
	phrase := "une phrase avec beaucoup de mots ordinaires " + strings.Repeat("mot ", 30)
	for _, l := range wrap78(phrase) {
		if len([]rune(l)) > paneWidth {
			t.Fatalf("phrase : ligne de %d colonnes", len([]rune(l)))
		}
	}
}

func auditWithRemediation() auditView {
	v := auditView{OK: true, AuditID: "audit-ok"}
	v.Remediation = &remediationView{
		Qdisc: "cake", Reason: "meilleur classement mesuré",
		LinuxTC:        []string{"tc qdisc add dev eth0 root cake bandwidth 5mbit", "tc -s qdisc show dev eth0"},
		RouterOS:       "/queue type add name=cgo-cake kind=cake\n/queue tree add name=cgo-up parent=global max-limit=5M",
		LinuxTCIngress: []string{"ip link add ifb0 type ifb", "tc qdisc add dev ifb0 root cake bandwidth 17mbit"},
		UpMeasuredMbps: 5.1, UpShapedMbps: 4.8,
	}
	return v
}

// Une commande par ligne avant repli : deux commandes ne partagent jamais une
// ligne de sortie.
func TestDeploymentPaneOneCommandPerLine(t *testing.T) {
	v := auditWithRemediation()
	lines := deploymentLines(v, platformLX)
	for _, l := range lines {
		if strings.Contains(l, "tc qdisc add") && strings.Contains(l, "tc -s qdisc") {
			t.Fatalf("deux commandes sur une ligne : %q", l)
		}
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "tc qdisc add dev eth0") {
		t.Fatalf("commande envoi absente du pane :\n%s", joined)
	}
	if !strings.Contains(joined, "réception (IFB)") {
		t.Fatalf("section réception absente :\n%s", joined)
	}
	if !strings.Contains(joined, "validée par rejeu") && !strings.Contains(joined, "aucune preuve") {
		t.Fatalf("ligne de rejeu absente :\n%s", joined)
	}
}

func TestDeploymentPaneNamesAssumedLan(t *testing.T) {
	v := auditWithRemediation()
	v.Remediation.RouterOSLanAssumed = true
	v.Remediation.RouterOSLanIface = "bridge"
	joined := strings.Join(deploymentLines(v, platformRO), "\n")
	if !strings.Contains(joined, "Interface LAN supposée « bridge »") {
		t.Fatalf("l'hypothèse LAN n'est pas dite :\n%s", joined)
	}
}

func TestProofPaneNamesSensorTier(t *testing.T) {
	v := auditWithRemediation()
	v.Kernel = &kernelView{Available: true, Source: "ebpf"}
	if !strings.Contains(strings.Join(proofLines(v), "\n"), "eBPF") {
		t.Fatal("tier eBPF absent du pane de preuves")
	}
	v.Kernel = &kernelView{Source: "tcpinfo", Reason: "objet eBPF absent"}
	joined := strings.Join(proofLines(v), "\n")
	if !strings.Contains(joined, "TCP_INFO") || !strings.Contains(joined, "objet eBPF absent") {
		t.Fatalf("repli TCP_INFO non expliqué :\n%s", joined)
	}
	v.Kernel = nil
	if !strings.Contains(strings.Join(proofLines(v), "\n"), "non instrumentée") {
		t.Fatal("instrumentation absente non dite")
	}
}

// L'identifiant vient du serveur et devient un chemin : il est validé avant.
// Le câblage, pas seulement les aides : le miroir apparaît sur les onglets 4 et
// 5, pas ailleurs, et un pane remplace la barre d'onglets.
func TestViewShowsAuditMirrorOnTabs4And5(t *testing.T) {
	var v auditView
	if err := json.Unmarshal([]byte(auditFixture), &v); err != nil {
		t.Fatal(err)
	}
	v.OK = true
	m := modelTUI{audit: v}
	for _, tab := range []int{3, 4} {
		m.tab = tab
		out := m.View()
		if !strings.Contains(out, "Lien sain") {
			t.Fatalf("onglet %d sans entête de verdict :\n%s", tab+1, out)
		}
		if !strings.Contains(out, "Repos : p50") || !strings.Contains(out, "Perte :") {
			t.Fatalf("onglet %d sans les quatre faits :\n%s", tab+1, out)
		}
	}
	m.tab = 2
	if strings.Contains(m.View(), "Lien sain") {
		t.Fatal("le miroir d'audit fuit sur un onglet de campagne")
	}
	m.tab = 4
	m.pane = paneProof
	out := m.View()
	if strings.Contains(out, "Setup") {
		t.Fatalf("barre d'onglets visible dans un pane :\n%s", out)
	}
	if !strings.Contains(out, "Preuves — audit-ok") {
		t.Fatalf("pane de preuves absent :\n%s", out)
	}
}

// Découvert en rendant la vue pour de vrai : le statut et l'aide poussaient des
// lignes à 100 colonnes. La vue entière tient maintenant dans 80.
func TestRenderedViewsStayWithinEightyColumns(t *testing.T) {
	var v auditView
	if err := json.Unmarshal([]byte(auditFixture), &v); err != nil {
		t.Fatal(err)
	}
	v.OK = true
	v.Kernel = &kernelView{Source: "tcpinfo", Reason: "objet eBPF absent"}
	m := modelTUI{audit: v, status: "data/audits/audit-ok/remediation.routeros.rsc · copie OSC 52 tentée, le terminal peut l'ignorer"}
	check := func(label, out string) {
		for _, l := range strings.Split(out, "\n") {
			if n := len([]rune(l)); n > 80 {
				t.Fatalf("%s : ligne de %d colonnes : %q", label, n, l)
			}
		}
	}
	for _, tab := range []int{3, 4} {
		m.tab = tab
		check("onglet", m.View())
	}
	m.pane = paneDeployment
	check("pane déploiement", m.View())
	m.pane = paneProof
	check("pane preuves", m.View())
}

func TestWriteRemediationRefusesTraversal(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"../evade", "..", "", "a/b", `a\b`, "run.1/.."} {
		v := auditWithRemediation()
		v.AuditID = id
		if _, err := writeRemediation(root, v, platformLX); err == nil {
			t.Fatalf("identifiant %q accepté", id)
		}
	}
	// rien n'a été créé hors du répertoire de base
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("écriture hors base : %v", entries)
	}
}

func TestWriteRemediationWritesBothPlatforms(t *testing.T) {
	root := t.TempDir()
	v := auditWithRemediation()

	roPath, err := writeRemediation(root, v, platformRO)
	if err != nil {
		t.Fatal(err)
	}
	roBytes, err := os.ReadFile(roPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(roBytes), "/queue type add") || !strings.HasSuffix(string(roBytes), "\n") {
		t.Fatalf("script RouterOS = %q", string(roBytes))
	}

	lxPath, err := writeRemediation(root, v, platformLX)
	if err != nil {
		t.Fatal(err)
	}
	lxBytes, _ := os.ReadFile(lxPath)
	if !strings.Contains(string(lxBytes), "tc qdisc add dev eth0") || !strings.Contains(string(lxBytes), "# réception (IFB)") {
		t.Fatalf("script Linux = %q", string(lxBytes))
	}
}

func TestWriteRemediationStaysUnderDataAudits(t *testing.T) {
	root := filepath.Join(t.TempDir(), "audits")
	v := auditWithRemediation()
	path, err := writeRemediation(root, v, platformLX)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, filepath.Clean(root)+string(os.PathSeparator)) {
		t.Fatalf("chemin %q hors de %q", path, root)
	}
	if filepath.Base(path) != "remediation.linux.sh" {
		t.Fatalf("nom de fichier = %q", filepath.Base(path))
	}
	// la racine de production est data/audits, jamais un temporaire : c'est
	// l'appelant (copyScript) qui la pose, donc c'est elle qu'on vérifie.
	if auditScriptRoot != filepath.Join("data", "audits") {
		t.Fatalf("racine de production = %q", auditScriptRoot)
	}
	if strings.HasPrefix(filepath.Clean(auditScriptRoot), filepath.Clean(os.TempDir())) {
		t.Fatalf("racine de production sous un temporaire : %q", auditScriptRoot)
	}
}

// Une plateforme absente du rapport n'est jamais écrite : `c` ne produit pas un
// fichier vide.
func TestScriptForRefusesMissingPlatform(t *testing.T) {
	v := auditView{OK: true, AuditID: "audit-ok", Remediation: &remediationView{Qdisc: "cake"}}
	if _, _, err := scriptFor(v, platformRO); err == nil {
		t.Fatal("script RouterOS vide accepté")
	}
	if _, _, err := scriptFor(v, platformLX); err == nil {
		t.Fatal("script Linux vide accepté")
	}
	if p := platformFor(v, platformRO); p != "" {
		t.Fatalf("plateforme %q inventée", p)
	}
}
