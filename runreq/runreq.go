// Package runreq — la règle unique de validation d'une campagne.
//
// Avant ce module, la même règle vivait en trois exemplaires (CLI, API,
// moteur) avec trois listes de messages différents. Ici, c'est UN seul
// endroit qui décide : chaque appelant rend son verdict, le module rend la
// même cause. Les messages sont en français — c'est la langue opérateur du
// produit, pas un hasard de l'histoire — et chaque refus porte un code de
// règle en ASCII (`Rule`) pour les appelants machine.
package runreq

import (
	"fmt"
	"strings"

	"github.com/Realms4239/cgo/model"
)

// Bench — le banc visé pour la campagne.
const (
	BenchLegacy = "legacy"
	BenchDualNS = "dualns"
)

// Direction — le sens de la charge (RRUL séquentiel).
const (
	DirectionUp   = "up"
	DirectionDown = "down"
	DirectionBoth = "both"
)

// ChannelModel — la physique du canal simulé.
const (
	Clean  = "clean"
	Bursty = "bursty"
)

// CampaignRequest — ce qu'un opérateur (CLI, HTTP) demande au banc.
// Zéro = la valeur n'est pas posée (défaut laissé au moteur : reps 3,
// flows 4, aucun sens, aucun bench, canal clean).
type CampaignRequest struct {
	RunID        string
	Profiles     []string
	Qdiscs       []string
	CCs          []string
	Reps         int
	Flows        int
	DeadlineMs   int
	Bench        string
	Direction    string
	ChannelModel string
	Target       string
}

// Refusal — le refus typé. Rule est la cause machine (stable, ASCII) ;
// Msg est ce que l'opérateur lit (français, descriptif).
type Refusal struct {
	Rule string
	Msg  string
}

func (r *Refusal) Error() string { return r.Msg }

// ValidBursty — P2/P3 uniquement : la physique rafales radio est documentée
// par profil, pas tolérée pour un autre canal.
func ValidBursty(profile string) bool {
	return profile == "P2" || profile == "P3"
}

// ValidBench — le banc n'existe pas tant que le moteur n'est pas câblé ;
// l'API le refuse avant de laisser passer un 200 mensonger.
func ValidBench(b string) bool {
	return b == "" || b == BenchLegacy || b == BenchDualNS
}

// ValidChannelModel — clean toujours disponible ; bursty restreint à P2/P3
// (vérifié ailleurs, ici c'est l'enum).
func ValidChannelModel(m string) bool { return m == "" || m == Clean || m == Bursty }

// ValidDirection — up | down | both.
func ValidDirection(d string) bool {
	return d == "" || d == DirectionUp || d == DirectionDown || d == DirectionBoth
}

var burstyProfiles = "P2/P3"

// Validate — retourne le premier refus, nil si la requête est valide.
// Une refus = une cause ; les appelants qui veulent la liste complète
// appellent ValidateAll.
func (c CampaignRequest) Validate() error {
	return first(c.ValidateAll())
}

// ValidateAll — chaque refus détecté, jamais d'arrêt prématuré au premier.
// L'UI liste toutes les erreurs à la fois plutôt que de laisser l'opérateur
// coller ses corrections une à une.
func (c CampaignRequest) ValidateAll() []error {
	var out []error
	add := func(rule, format string, a ...any) {
		out = append(out, &Refusal{Rule: rule, Msg: fmt.Sprintf(format, a...)})
	}

	if len(c.Profiles) == 0 {
		add("profiles", "aucun profil sélectionné")
	} else {
		_, missing := knownProfiles(c.Profiles)
		for _, id := range missing {
			add("profiles", "profil inconnu: %s", id)
		}
	}
	if c.Reps < 0 || c.Reps > 5 {
		add("reps", "répétitions hors bornes (1–5)")
	}
	if c.Flows != 0 && (c.Flows < 1 || c.Flows > 8) {
		add("flows", "flux hors bornes (1–8)")
	}
	if c.DeadlineMs != 0 && (c.DeadlineMs < 200 || c.DeadlineMs > 5000) {
		add("deadline_ms", "deadline small p95 hors bornes (200–5000 ms)")
	}
	if len(c.Target) > 64 {
		add("target", "cible trop longue (≤ 64 caractères)")
	}
	if !ValidBench(c.Bench) {
		add("bench", "banc doit être legacy | dualns")
	}
	if !ValidChannelModel(c.ChannelModel) {
		add("channel_model", "canal doit être clean | bursty")
	}
	if c.ChannelModel == Bursty {
		for _, id := range c.Profiles {
			if !ValidBursty(id) {
				add("channel_model", "canal bursty invalide pour %s — rafales sur %s uniquement (choisir clean)", id, burstyProfiles)
			}
		}
	}
	if !ValidDirection(c.Direction) {
		add("direction", "sens de charge doit être up | down | both")
	}
	if (len(c.Qdiscs) == 0) != (len(c.CCs) == 0) {
		add("axes", "qdiscs et ccs se filtrent ensemble")
	}
	for _, q := range UnknownQdiscs(c.Qdiscs) {
		add("qdiscs", "qdisc inconnu: %s", q)
	}
	for _, cc := range UnknownCCs(c.CCs) {
		add("ccs", "cc inconnu: %s", cc)
	}
	return out
}

// UnknownQdiscs — les valeurs posées que le banc ne connaît pas.
func UnknownQdiscs(qs []string) []string {
	valid := map[string]bool{}
	for _, q := range model.AllQdiscs {
		valid[string(q)] = true
	}
	var out []string
	for _, q := range qs {
		if !valid[q] {
			out = append(out, q)
		}
	}
	return out
}

// HostCCAdditions — le noyau conditionne la matrice
// (cubic/bbr attendus de base, le reste ci-dessous).
func UnknownCCs(ccs []string) []string {
	valid := map[string]bool{}
	for _, c := range model.KnownCC {
		valid[string(c)] = true
	}
	var out []string
	for _, cc := range ccs {
		if !valid[cc] {
			out = append(out, cc)
		}
	}
	return out
}

func knownProfiles(ids []string) ([]string, []string) {
	model.ProfilesMu.RLock()
	defer model.ProfilesMu.RUnlock()
	var ok, bad []string
	for _, id := range ids {
		if _, known := model.Profiles[id]; known {
			ok = append(ok, id)
		} else {
			bad = append(bad, id)
		}
	}
	return ok, bad
}

func first(items []error) error {
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

// ContainsP3 / ContainsBursty — les règles bursty se partagent
// à la fois l'API (avant porte moteur) et le moteur (before physics).
func BurstyViolators(profiles []string) []string {
	var out []string
	for _, id := range profiles {
		if !ValidBursty(id) {
			out = append(out, id)
		}
	}
	return out
}

// BurstyViolationMsg — le message exact produit par l'API et le moteur.
// Un seul exemplaire, un seul libellé, jamais deux copies.
func BurstyViolationMsg(profile string) string {
	return fmt.Sprintf(
		"canal bursty invalide pour %s — rafales sur %s uniquement (choisir clean)",
		idToLabel(profile), burstyProfiles)
}

func idToLabel(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	if id == "" {
		return "?"
	}
	return id
}
