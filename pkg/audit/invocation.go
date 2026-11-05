package audit

import (
	"strconv"
	"strings"
)

// Invocation — la reproduction gelée d'un audit. Le rapport scellait des mesures
// et la provenance du CSV, mais rien sur les réglages qui avaient produit la
// mesure : le tiroir de la console n'avait donc aucune commande à montrer, et un
// vérificateur ne pouvait pas relancer le même audit. Ce bloc est construit là où
// les défauts sont déjà résolus (Params après normalisation par Run), jamais là où
// la demande est entrée — sinon on écrirait une commande qui ne rejoue pas ce qui
// a tourné.
type Invocation struct {
	// Origin — "cli" ou "http" : ce que l'opérateur a réellement utilisé.
	// Command est l'équivalent CLI, pas la transcription de sa frappe.
	Origin   string `json:"origin"`
	SinkMode string `json:"sink_mode,omitempty"`
	Duration int    `json:"duration_sec"`
	Flows    int    `json:"flows"`
	// Direction — le sens effectif de la charge, après normalisation : une
	// demande vide devient "up", et la commande doit dire "up".
	Direction string `json:"direction"`
	Site      string `json:"site,omitempty"`
	LinkType  string `json:"link_type,omitempty"`
	Provider  string `json:"provider,omitempty"`
	Target    string `json:"target,omitempty"`
	SmallURL  string `json:"small_url,omitempty"`
	BulkAddr  string `json:"bulk_addr,omitempty"`
	// Command — chaque réglage explicite, en syntaxe POSIX. Vide = rapport scellé
	// avant ce champ ; jamais une commande inventée pour combler l'absence.
	Command string `json:"replay_command"`
}

// shellQuote — guillemet simple POSIX, l'apostrophe interne devenant '\”. Le
// moteur n'exécute pas cette chaîne : elle est écrite pour être relue puis collée
// dans un shell, où un nom de site contenant une apostrophe finirait sinon hors
// des guillemets.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// arg — un drapeau "-nom='valeur'" ; les entiers se passent de guillemets.
func arg(name, value string) string {
	return " -" + name + "=" + shellQuote(value)
}

func argInt(name string, value int) string {
	return " -" + name + "=" + strconv.Itoa(value)
}

// invocationFrom — les cinq drapeaux qui définissent la mesure sont toujours
// publiés, même égaux au défaut : une reproduction qui omet -flows re-détecterait
// le banc au lieu de rejouer la campagne. Les cibles ne sortent que si elles ont
// servi : -bulk=” demanderait explicitement « pas de charge », ce qui n'est pas
// ce qui a tourné quand le champ était vide.
func invocationFrom(p Params) Invocation {
	in := Invocation{
		Origin: p.Origin, SinkMode: p.SinkMode, Duration: p.Duration,
		Flows: p.Flows, Direction: p.Direction, Site: p.Site,
		LinkType: p.LinkType, Provider: p.Provider,
		Target: p.Target, SmallURL: p.SmallURL, BulkAddr: p.BulkAddr,
	}
	if in.Direction == "" {
		in.Direction = "up"
	}
	var b strings.Builder
	b.WriteString("cgo audit")
	b.WriteString(arg("site", in.Site))
	b.WriteString(arg("link-type", in.LinkType))
	b.WriteString(argInt("duration", in.Duration))
	b.WriteString(arg("direction", in.Direction))
	b.WriteString(argInt("flows", in.Flows))
	for _, opt := range []struct {
		name, value string
	}{
		{"provider", in.Provider},
		{"target", in.Target},
		{"small", in.SmallURL},
		{"bulk", in.BulkAddr},
	} {
		if opt.value != "" {
			b.WriteString(arg(opt.name, opt.value))
		}
	}
	in.Command = b.String()
	return in
}

// avecRemediation — complète la commande avec ce que seul le rapport connaît :
// le qdisc retenu, les interfaces façonquées et les deux bascules de
// recommandation. Une invocation vide (rapport archivé) reste vide : il n'y a
// rien à compléter, et publier "-qdisc=”" ferait d'une absence un réglage.
func (in Invocation) avecRemediation(qdisc, iface, rosWan, rosLan string, asymmetric, cpuConstrained bool) Invocation {
	if in.Command == "" {
		return in
	}
	var b strings.Builder
	b.WriteString(in.Command)
	// sink_mode n'est PAS un drapeau du CLI : la détection se rejoue à chaque
	// audit. Le consigner sert au tiroir, pas à la ligne de commande.
	for _, opt := range []struct {
		name, value string
	}{
		{"qdisc", qdisc},
		{"iface", iface},
		{"routeros-iface", rosWan},
		{"ros-lan", rosLan},
	} {
		if opt.value != "" {
			b.WriteString(arg(opt.name, opt.value))
		}
	}
	if !asymmetric {
		b.WriteString(" -asymmetric=false")
	}
	if cpuConstrained {
		b.WriteString(" -cpu-constrained")
	}
	in.Command = b.String()
	return in
}
