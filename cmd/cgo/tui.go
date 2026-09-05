package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---- styles ----

var (
	stTab      = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("240"))
	stTabOn    = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("5ad3e3")).Bold(true)
	stTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("f2f2f4"))
	stMuted    = lipgloss.NewStyle().Foreground(lipgloss.Color("767b84"))
	stOK       = lipgloss.NewStyle().Foreground(lipgloss.Color("1fa348"))
	stWarn     = lipgloss.NewStyle().Foreground(lipgloss.Color("f4b400"))
	stErr      = lipgloss.NewStyle().Foreground(lipgloss.Color("e22718"))
	stCell     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("26262a")).Padding(0, 1)
	stCellOn   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("5ad3e3")).Padding(0, 1)
	stBadgeRun = lipgloss.NewStyle().Foreground(lipgloss.Color("5ad3e3")).Bold(true)
)

const tabsUI = "[1 Setup] [2 Kit] [3 Campagne] [4 Live] [5 Résultats]  (h/l ou 1-5, ? aide, q quitter)"

// ---- modèles ----

type frameTUI struct {
	Phase       string  `json:"phase"`
	Profile     string  `json:"profile"`
	Qdisc       string  `json:"qdisc"`
	CC          string  `json:"cc"`
	EventID     int     `json:"event_id"`
	TotalEvents int     `json:"total_events"`
	SmallP95    float64 `json:"small_p95_ms"`
	RTTp95      float64 `json:"rtt_p95_ms"`
	Goodput     float64 `json:"bulk_goodput_mbps"`
	Drops       uint64  `json:"drops"`
	Running     bool    `json:"running"`
	OK          bool    `json:"ok"`
	Mode        string  `json:"mode"`
	Version     string  `json:"version"`
}

type resultRow struct {
	Profile string  `json:"profile"`
	Qdisc   string  `json:"qdisc"`
	CC      string  `json:"cc"`
	Count   int     `json:"count"`
	Small   float64 `json:"small_p95_median"`
	RTT     float64 `json:"rtt_p95_median"`
	Good    float64 `json:"goodput_median"`
	Best    bool    `json:"best"`
	// valid-only (additif, absent des vieux serveurs → 0 = repli count).
	ValidN  int        `json:"small_p95_valid_n"`
	SmallCI [2]float64 `json:"small_p95_valid_ci95"`
}

type modelTUI struct {
	tab       int // 0..4
	kitSel    int
	kitAction string
	kitDone   string
	apiURL    string
	frame     frameTUI
	rows      []resultRow
	resultsOK bool
	err       string
	quitting  bool
	lastTUI   time.Time
}

func initialModelTUI() modelTUI {
	api := os.Getenv("CGO_API")
	if api == "" {
		api = "https://meteolink.dev:9090"
	}
	return modelTUI{apiURL: api}
}

// apiTransport — HTTPS : racines système + certificat local meteolink.dev
// (auto-signé accepté, pas de skip-verify). HTTP : transport standard.
var (
	httpsTransportOnce sync.Once
	httpsTransport     *http.Transport
)

func apiTransport(api string) http.RoundTripper {
	if !strings.HasPrefix(api, "https://") {
		return http.DefaultTransport
	}
	httpsTransportOnce.Do(func() {
		pool, err := localCertPool()
		if err != nil {
			pool = nil // racines système seules (le TLS échouera proprement)
		}
		httpsTransport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}
	})
	if httpsTransport == nil {
		return http.DefaultTransport
	}
	return httpsTransport
}

// ---- messages de polling ----

type tickMsg time.Time
type frameMsg frameTUI
type resultsMsg []resultRow

func pollFrames(api string) tea.Cmd {
	return tea.Tick(1*time.Second, func(time.Time) tea.Msg {
		cl := &http.Client{Timeout: 2 * time.Second, Transport: apiTransport(api)}
		resp, err := cl.Get(api + "/api/state")
		if err != nil {
			return frameMsg{}
		}
		defer resp.Body.Close()
		var f frameTUI
		_ = json.NewDecoder(resp.Body).Decode(&f)
		return frameMsg(f)
	})
}

func pollResults(api string) tea.Cmd {
	return func() tea.Msg {
		cl := &http.Client{Timeout: 3 * time.Second, Transport: apiTransport(api)}
		resp, err := cl.Get(api + "/api/results")
		if err != nil {
			return resultsMsg(nil)
		}
		defer resp.Body.Close()
		var doc struct {
			Available bool        `json:"available"`
			Groups    []resultRow `json:"groups"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&doc)
		if doc.Available {
			return resultsMsg(doc.Groups)
		}
		return resultsMsg(nil)
	}
}

func kitExec(action string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(os.Args[0], "kit", action)
		out, _ := cmd.CombinedOutput()
		return kitDoneMsg{action: action, out: string(out)}
	}
}

type kitDoneMsg struct {
	action string
	out    string
}

// ---- Update ----

func (m modelTUI) Init() tea.Cmd { return tea.Batch(pollFrames(m.apiURL)) }

func (m modelTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "1", "2", "3", "4", "5":
			m.tab = int(msg.String()[0] - '1')
		case "h", "left":
			if m.tab > 0 {
				m.tab--
			}
		case "l", "right":
			if m.tab < 4 {
				m.tab++
			}
		case "j", "down":
			if m.tab == 1 && m.kitSel < 8 {
				m.kitSel++
			}
		case "k", "up":
			if m.tab == 1 && m.kitSel > 0 {
				m.kitSel--
			}
		case "enter", "o":
			if m.tab == 1 {
				actions := []string{"doctor", "scan", "ensure", "align", "build", "deploy", "bootstrap", "status", "logs"}
				if m.kitSel < len(actions) {
					return m, kitExec(actions[m.kitSel])
				}
			}
			if m.tab == 3 {
				// campagne : démarrer/arrêter selon l'état
				var cmd tea.Cmd
				if m.frame.Running {
					cmd = httpPost(m.apiURL, "/api/run/stop", "")
				} else {
					cmd = httpPost(m.apiURL, "/api/run/start", `{"profiles":["P2"],"reps":1}`)
				}
				return m, cmd
			}
		}
	case tickMsg:
		return m, pollFrames(m.apiURL)
	case frameMsg:
		m.frame = frameTUI(msg)
		return m, pollFrames(m.apiURL)
	case resultsMsg:
		m.rows = msg
		m.resultsOK = msg != nil
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return pollResults(m.apiURL)() })
	case kitDoneMsg:
		m.kitDone = strings.TrimSpace(msg.out)
		return m, nil
	}
	return m, nil
}

func httpPost(api, path, body string) tea.Cmd {
	return func() tea.Msg {
		cl := &http.Client{Timeout: 5 * time.Second}
		var resp *http.Response
		var err error
		if body == "" {
			resp, err = cl.Post(api+path, "", nil)
		} else {
			resp, err = cl.Post(api+path, "application/json", strings.NewReader(body))
		}
		if err != nil {
			return frameMsg{}
		}
		defer resp.Body.Close()
		var f frameTUI
		_ = json.NewDecoder(resp.Body).Decode(&f)
		return frameMsg(f)
	}
}

// ---- View ----

func (m modelTUI) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder

	// barre d'onglets
	for i, name := range []string{"Setup", "Kit", "Campagne", "Live", "Résultats"} {
		if i == m.tab {
			b.WriteString(stTabOn.Render(name))
		} else {
			b.WriteString(stTab.Render(name))
		}
	}
	b.WriteString("  " + stMuted.Render("h/l · q quitter") + "\n\n")

	switch m.tab {
	case 0: // Setup
		b.WriteString(stTitle.Render("Setup — première installation") + "\n\n")
		b.WriteString("  Le wizard complet : " + stBadgeRun.Render("cgo setup") + "\n")
		b.WriteString("  Modes : interactif · --yes (CI) · --no-vm (sans hyperviseur)\n\n")
		b.WriteString("  " + stMuted.Render("Étapes : détection → dépendances → build → contexte →\n           config VM → DNS local → doctor + récap") + "\n")
	case 1: // Kit
		b.WriteString(stTitle.Render("Kit — moteur de déploiement") + "\n\n")
		actions := []string{"doctor", "scan", "ensure", "align", "build", "deploy", "bootstrap", "status", "logs"}
		for i, a := range actions {
			style := stCell
			if i == m.kitSel {
				style = stCellOn
			}
			b.WriteString(style.Render(a))
			if i%3 == 2 {
				b.WriteString("\n")
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n\n" + stMuted.Render("j/k choisir · entrée exécuter") + "\n")
		if m.kitDone != "" {
			b.WriteString("\n" + stMuted.Render("--- dernière sortie ---") + "\n" + m.kitDone + "\n")
		}
	case 2: // Campagne
		b.WriteString(stTitle.Render("Campagne — pilotez la mesure") + "\n\n")
		if m.frame.Running {
			b.WriteString("  " + stBadgeRun.Render(fmt.Sprintf("Event %d/%d — %s/%s/%s — %s",
				m.frame.EventID, m.frame.TotalEvents, m.frame.Profile, m.frame.Qdisc, m.frame.CC, m.frame.Phase)) + "\n")
			b.WriteString("  small p95 : " + fmt.Sprintf("%.1f", m.frame.SmallP95) + "ms · rtt : " + fmt.Sprintf("%.1f", m.frame.RTTp95) + "ms\n")
			b.WriteString("\n  " + stErr.Render("ENTRÉE = arrêter la campagne (gel gracieux)"))
		} else {
			b.WriteString("  " + stOK.Render("ENTRÉE = démarrer P2 ×1 (deadline 1000ms)") + "\n")
			b.WriteString("  " + stMuted.Render("matrice : 3 qdiscs × 2 CC × reps = 6 events") + "\n")
			b.WriteString("  " + stMuted.Render("CLI équivalent : cgo run --profiles P2 --reps 1"))
		}
	case 3: // Live
		b.WriteString(stTitle.Render("Tableau live — "+m.frame.Phase) + "\n\n")
		cards := []struct {
			label string
			val   string
		}{
			{"rtt p95", fmt.Sprintf("%.1f ms", m.frame.RTTp95)},
			{"small p95", fmt.Sprintf("%.1f ms", m.frame.SmallP95)},
			{"goodput", fmt.Sprintf("%.1f Mb/s", m.frame.Goodput)},
			{"drops", fmt.Sprintf("%d", m.frame.Drops)},
			{"event", fmt.Sprintf("%d/%d", m.frame.EventID, m.frame.TotalEvents)},
		}
		for i, c := range cards {
			b.WriteString(stCell.Render(stMuted.Render(c.label) + "\n" + c.val))
			if i%2 == 1 {
				b.WriteString("\n")
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
		if !m.frame.Running {
			b.WriteString("\n  " + stMuted.Render("idle — démarrez depuis Campagne"))
		}
	case 4: // Résultats
		b.WriteString(stTitle.Render("Résultats — classement complet") + "\n\n")
		if !m.resultsOK || len(m.rows) == 0 {
			b.WriteString("  " + stMuted.Render("aucun résultat gelé — lancez une campagne"))
			return b.String() + "\n"
		}
		b.WriteString(fmt.Sprintf("  %-6s %-10s %-5s %10s %10s %10s\n", "profil", "qdisc", "cc", "small p95", "rtt p95", "goodput"))
		for _, r := range m.rows {
			mark := ""
			if r.Best {
				mark = " ★"
			}
			nv := r.ValidN
			if nv <= 0 {
				nv = r.Count // vieux serveur sans champ valid : repli honnête
			}
			ci := ""
			if r.SmallCI[0] > 0 || r.SmallCI[1] > 0 {
				ci = fmt.Sprintf(" [%.0f-%.0f]", r.SmallCI[0], r.SmallCI[1])
			}
			b.WriteString(fmt.Sprintf("  %-6s %-10s %-5s %9.1fms %9.1fms %8.1fMb n=%dv%s%s\n",
				r.Profile, r.Qdisc, r.CC, r.Small, r.RTT, r.Good, nv, ci, mark))
		}
	}
	return b.String() + "\n"
}

// runTUI — tableau de bord terminal complet : 5 onglets, pilotage réel via
// l'API locale + exécution kit en sous-process. Honest : sans API, Live
// affiche l'état déconnecté, jamais de données fabriquées.
func runTUI(args []string) int {
	p := tea.NewProgram(initialModelTUI(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui:", err)
		return 1
	}
	return 0
}
