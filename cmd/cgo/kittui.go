package main

// Centre de contrôle kit — TUI interactive, zéro commande à taper.
// Un PC Windows + une VM Ubuntu : ce TUI scanne les hyperviseurs, fait
// choisir la VM au clavier, installe les dépendances, pose la clé SSH,
// pousse le binaire et sert le dashboard. Les actions lourdes tournent en
// fond avec sortie en direct (kittui_run.go) ; keysetup suspend le TUI pour
// le prompt mot de passe natif puis reprend.

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Realms4239/cgo/internal/kit"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ktStep int

const (
	ktDeps ktStep = iota
	ktVM
	ktAcces
	ktDeploy
	ktControle
)

var ktStepNames = []string{"1 Dépendances", "2 Machine", "3 Accès SSH", "4 Déployer", "5 Contrôle"}

// ktItem — une ligne actionnable : id stable pour Update, label + hint affichés.
type ktItem struct {
	id    string
	label string
	hint  string
}

type vmEntry struct {
	path string
	name string
	hyp  string // vmware | virtualbox | ?
	live bool   // allumée
}

type depRow struct {
	label string
	ok    bool
	info  string
	fixID string // "" = rien à installer
}

type modelKT struct {
	bus     *ktBus
	cfgPath string
	cfg     *kit.Config
	version string

	step    ktStep
	cursor  int
	quitArm bool
	width   int
	height  int

	deps []depRow
	vms  []vmEntry

	busy string // action en cours ("" = libre)
	log  []string

	sshState  string // "", "ok", "ko:…"
	dashState string // "", "ok <ver>", "ko"
}

func initialModelKT(cfgPath string, version string) modelKT {
	c, _ := kit.LoadConfig(cfgPath)
	return modelKT{bus: &ktBus{}, cfgPath: cfgPath, cfg: c, version: version}
}

func (m *modelKT) pushLog(s string) {
	m.log = append(m.log, s)
	if len(m.log) > 200 {
		m.log = m.log[len(m.log)-200:]
	}
}

func (m *modelKT) reloadCfg() {
	if c, err := kit.LoadConfig(m.cfgPath); err == nil {
		m.cfg = c
	}
}

func (m modelKT) lockedVM() string {
	if m.cfg == nil {
		return ""
	}
	if m.cfg.VMName != "" {
		return m.cfg.VMName
	}
	if m.cfg.VMXPath != "" {
		return filepath.Base(m.cfg.VMXPath)
	}
	return ""
}

// items — la liste actionnable de l'étape courante (curseur 0..n-1).
func (m modelKT) items() []ktItem {
	switch m.step {
	case ktDeps:
		out := []ktItem{}
		for _, d := range m.deps {
			if !d.ok && d.fixID != "" {
				out = append(out, ktItem{id: d.fixID, label: "Installer : " + d.label, hint: d.info})
			}
		}
		out = append(out,
			ktItem{id: "recheck", label: "Revérifier", hint: "relance le contrôle"},
			ktItem{id: "next", label: "Continuer → Machine", hint: ""})
		return out
	case ktVM:
		out := []ktItem{{id: "rescan", label: "↻ Rescanner tout le PC", hint: "vmware + virtualbox, profond"}}
		for i, v := range m.vms {
			state := "éteinte"
			if v.live {
				state = "allumée"
			}
			out = append(out, ktItem{id: fmt.Sprintf("vm:%d", i), label: v.name, hint: v.hyp + " · " + state})
		}
		return out
	case ktAcces:
		return []ktItem{
			{id: "ensure", label: "Démarrer la VM / vérifier SSH", hint: "boot headless + attente SSH"},
			{id: "keysetup", label: "Poser la clé SSH", hint: "mot de passe demandé une fois"},
			{id: "retry", label: "Revérifier l'accès", hint: ""},
			{id: "next", label: "Continuer → Déployer", hint: ""},
		}
	case ktDeploy:
		return []ktItem{
			{id: "deploy", label: "DÉPLOYER vers la VM", hint: "push binaire + install + health"},
			{id: "open", label: "Ouvrir le dashboard", hint: "navigateur — " + m.dashURL()},
			{id: "next", label: "Continuer → Contrôle", hint: ""},
		}
	default: // ktControle
		return []ktItem{
			{id: "svc-status", label: "État dashboard", hint: ""},
			{id: "svc-start", label: "Démarrer dashboard", hint: ""},
			{id: "svc-stop", label: "Arrêter dashboard", hint: ""},
			{id: "svc-restart", label: "Redémarrer dashboard", hint: ""},
			{id: "vm-start", label: "Démarrer la VM", hint: "headless"},
			{id: "vm-stop", label: "Arrêter la VM", hint: "ACPI puis forcé"},
			{id: "logs", label: "Journal (40 lignes)", hint: ""},
			{id: "dns", label: "Mapper meteolink.dev", hint: "admin requis"},
			{id: "tls", label: "Confiance HTTPS", hint: "admin requis"},
			{id: "verify", label: "Vérifier intégrité", hint: "SHA-256 sur la VM"},
			{id: "backup", label: "Rapatrier runs", hint: "./backup"},
			{id: "snapshot", label: "Snapshot VM", hint: "garde-fou"},
			{id: "keysetup", label: "Reposer la clé SSH", hint: ""},
			{id: "open", label: "Ouvrir le dashboard", hint: ""},
			{id: "switch", label: "Changer de VM", hint: "→ étape Machine"},
		}
	}
}

func (m modelKT) dashURL() string {
	host, port := "meteolink.dev", "9090"
	if m.cfg != nil {
		if m.cfg.DashHost != "" {
			host = m.cfg.DashHost
		}
		if m.cfg.DashPort != "" {
			port = m.cfg.DashPort
		}
	}
	return "https://" + host + ":" + port
}

// ---- vue ----

var (
	stKHi    = lipgloss.NewStyle().Foreground(lipgloss.Color("5ad3e3")).Bold(true)
	stKStep  = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("240"))
	stKStep_ = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("5ad3e3")).Bold(true).Underline(true)
	stKOK    = lipgloss.NewStyle().Foreground(lipgloss.Color("1fa348"))
	stKKO    = lipgloss.NewStyle().Foreground(lipgloss.Color("e22718"))
	stKDim   = lipgloss.NewStyle().Foreground(lipgloss.Color("8b9099"))
	stKBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("26262a")).Padding(0, 1)
)

func (m modelKT) Init() tea.Cmd {
	return func() tea.Msg { return ktBooted{} }
}

func (m modelKT) View() string {
	var b strings.Builder
	b.WriteString(stTitle.Render("Meteolink Kit — centre de contrôle") + "  " + stMuted.Render(m.version) + "\n")
	for i, name := range ktStepNames {
		if ktStep(i) == m.step {
			b.WriteString(stKStep_.Render(name))
		} else {
			b.WriteString(stKStep.Render(name))
		}
	}
	b.WriteString("\n\n")

	switch m.step {
	case ktDeps:
		b.WriteString(m.viewDeps())
	case ktVM:
		b.WriteString(m.viewVMs())
	case ktAcces:
		b.WriteString(m.viewAcces())
	case ktDeploy:
		b.WriteString(m.viewDeploy())
	default:
		b.WriteString(m.viewControle())
	}

	// journal (8 dernières lignes)
	b.WriteString("\n" + stMuted.Render("── journal ──") + "\n")
	lines := m.log
	if len(lines) > 8 {
		lines = lines[len(lines)-8:]
	}
	if len(lines) == 0 {
		b.WriteString(stMuted.Render("  (silence — choisissez une action)") + "\n")
	}
	for _, l := range lines {
		b.WriteString("  " + l + "\n")
	}
	if m.busy != "" {
		b.WriteString("\n  " + stKHi.Render("◌ "+m.busy+" …") + "\n")
	}
	b.WriteString("\n" + stMuted.Render("j/k déplacer · entrée activer · 1-5 étapes · q quitter") + "\n")
	return b.String()
}

func (m modelKT) viewItems() string {
	var b strings.Builder
	for i, it := range m.items() {
		mark := "  "
		if i == m.cursor {
			mark = stKHi.Render("▸ ")
		}
		hint := ""
		if it.hint != "" {
			hint = "  " + stKDim.Render("("+it.hint+")")
		}
		b.WriteString(mark + it.label + hint + "\n")
	}
	return b.String()
}

func (m modelKT) viewDeps() string {
	var b strings.Builder
	b.WriteString(stTitle.Render("Dépendances du poste") + "\n\n")
	if len(m.deps) == 0 {
		b.WriteString("  " + stMuted.Render("contrôle en cours…") + "\n\n")
	} else {
		for _, d := range m.deps {
			st := stKOK.Render("  [ok]")
			if !d.ok {
				st = stKKO.Render("  [KO]")
			}
			b.WriteString(fmt.Sprintf("%s %-16s %s\n", st, d.label, stKDim.Render(d.info)))
		}
		b.WriteString("\n")
	}
	b.WriteString(m.viewItems())
	return b.String()
}

func (m modelKT) viewVMs() string {
	var b strings.Builder
	locked := m.lockedVM()
	if locked != "" {
		b.WriteString("  Machine verrouillée : " + stKHi.Render(locked) + "\n\n")
	} else {
		b.WriteString("  " + stWarn.Render("Aucune machine verrouillée — choisissez dans la liste.") + "\n\n")
	}
	if len(m.vms) == 0 && !m.scanning() {
		b.WriteString("  " + stMuted.Render("aucune VM trouvée — Rescanner.") + "\n\n")
	}
	b.WriteString(m.viewItems())
	return b.String()
}

func (m *modelKT) scanning() bool { return m.busy == "scan" }

func (m modelKT) viewAcces() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  Cible : %s\n", stKHi.Render(m.sshTarget())))
	ssh := m.sshState
	var sshTxt string
	if ssh == "" {
		sshTxt = stMuted.Render("non vérifié")
	} else if ssh == "ok" {
		sshTxt = stKOK.Render("SSH actif ✓")
	} else {
		sshTxt = stKKO.Render(ssh)
	}
	b.WriteString(fmt.Sprintf("  Accès : %s\n\n", sshTxt))
	b.WriteString(m.viewItems())
	return b.String()
}

func (m modelKT) sshTarget() string {
	if m.cfg == nil {
		return "—"
	}
	return fmt.Sprintf("%s@%s:%s (clé %s)", m.cfg.SSHUser, m.cfg.SSHHost, m.cfg.SSHPort, m.cfg.SSHKey)
}

func (m modelKT) viewDeploy() string {
	var b strings.Builder
	mode := "recompile depuis les sources"
	if _, err := exec.LookPath("go"); err != nil {
		mode = "pousse le binaire précompilé"
	}
	b.WriteString(fmt.Sprintf("  Binaire : %s\n", stKHi.Render(mode)))
	b.WriteString(fmt.Sprintf("  État dashboard : %s\n\n", m.dashStateOr("-")))
	b.WriteString(m.viewItems())
	return b.String()
}

func (m modelKT) dashStateOr(fb string) string {
	if m.dashState == "" {
		return stMuted.Render(fb)
	}
	if strings.HasPrefix(m.dashState, "ok") {
		return stKOK.Render(m.dashState)
	}
	return stKKO.Render(m.dashState)
}

func (m modelKT) viewControle() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  VM : %s   SSH : %s   Dashboard : %s\n\n",
		stKHi.Render(firstNonEmpty(m.lockedVM(), "—")),
		sshShort(m.sshState), m.dashStateOr("—")))
	b.WriteString(m.viewItems())
	return b.String()
}

func firstNonEmpty(v, fb string) string {
	if v != "" {
		return v
	}
	return fb
}

func sshShort(s string) string {
	if s == "" {
		return stMuted.Render("—")
	}
	if s == "ok" {
		return stKOK.Render("actif")
	}
	return stKKO.Render("coupé")
}
