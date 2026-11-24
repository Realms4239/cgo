// topocheck compare un contrat de topologie déclaré à des faits FOURNIS.
// Lecture seule : aucun ip/tc, aucun shell, aucune mutation du banc. Un fait
// absent n'est jamais supposé conforme — il est signalé comme drift explicite.
// Codes de sortie : 0 = accord, 1 = drift, 2 = contrôle impossible.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Contract struct {
	Namespaces []string `json:"namespaces"`
	Ifaces     []string `json:"ifaces"`
	Qdisc      string   `json:"qdisc"`
	CapMbps    float64  `json:"cap_mbps"`
	DelayMs    float64  `json:"delay_ms"`
	ShapeMode  string   `json:"shape_mode"`
}

// Facts — faits de topologie tels que fournis à la vérification. Aucune valeur
// par défaut n'est inventée : un champ absent reste zéro/vide et devient un
// drift explicite dans Check.
type Facts struct {
	Namespaces []string `json:"namespaces"`
	Ifaces     []string `json:"ifaces"`
	Qdisc      string   `json:"qdisc"`
	CapMbps    float64  `json:"cap_mbps"`
	DelayMs    float64  `json:"delay_ms"`
	ShapeMode  string   `json:"shape_mode"`
}

type Drift struct {
	Field    string `json:"field"`
	Declared string `json:"declared"`
	Observed string `json:"observed"`
}

func (d Drift) String() string {
	return fmt.Sprintf("%s: declared %s, observed %s", d.Field, d.Declared, d.Observed)
}

// rawFields — pointeurs pour distinguer un champ absent d'un champ à zéro.
type rawFields struct {
	Namespaces *[]string `json:"namespaces"`
	Ifaces     *[]string `json:"ifaces"`
	Qdisc      *string   `json:"qdisc"`
	CapMbps    *float64  `json:"cap_mbps"`
	DelayMs    *float64  `json:"delay_ms"`
	ShapeMode  *string   `json:"shape_mode"`
}

func decode(raw *rawFields, path string, strictFields bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if strictFields {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(raw); err != nil {
		return err
	}
	// Une seconde valeur JSON n'est pas un contrat/fait valide.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("valeurs JSON multiples")
		}
		return err
	}
	return nil
}

func validate(raw rawFields) error {
	if raw.Namespaces == nil || raw.Ifaces == nil || raw.Qdisc == nil || raw.CapMbps == nil || raw.DelayMs == nil || raw.ShapeMode == nil {
		return fmt.Errorf("champ requis absent")
	}
	if len(*raw.Namespaces) == 0 || len(*raw.Ifaces) == 0 {
		return fmt.Errorf("namespaces et ifaces doivent être non vides")
	}
	for _, list := range [][]string{*raw.Namespaces, *raw.Ifaces} {
		for _, v := range list {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("entrée vide dans namespaces/ifaces")
			}
		}
	}
	if strings.TrimSpace(*raw.Qdisc) == "" {
		return fmt.Errorf("qdisc vide")
	}
	if math.IsNaN(*raw.CapMbps) || math.IsInf(*raw.CapMbps, 0) || *raw.CapMbps <= 0 {
		return fmt.Errorf("cap_mbps doit être > 0")
	}
	if math.IsNaN(*raw.DelayMs) || math.IsInf(*raw.DelayMs, 0) || *raw.DelayMs <= 0 {
		return fmt.Errorf("delay_ms doit être > 0")
	}
	switch *raw.ShapeMode {
	case "ingress", "egress":
	default:
		return fmt.Errorf("shape_mode %q invalide (ingress|egress requis)", *raw.ShapeMode)
	}
	return nil
}

// Load refuse un contrat absent, invalide ou incomplet (échec fermé).
func Load(path string) (Contract, error) {
	var raw rawFields
	if err := decode(&raw, path, true); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Contract{}, fmt.Errorf("contrat absent (%s): fail closed: %w", path, err)
		}
		return Contract{}, fmt.Errorf("contrat illisible/invalide (%s): %w", path, err)
	}
	if err := validate(raw); err != nil {
		return Contract{}, fmt.Errorf("contrat invalide (%s): %w", path, err)
	}
	return Contract{
		Namespaces: *raw.Namespaces,
		Ifaces:     *raw.Ifaces,
		Qdisc:      *raw.Qdisc,
		CapMbps:    *raw.CapMbps,
		DelayMs:    *raw.DelayMs,
		ShapeMode:  *raw.ShapeMode,
	}, nil
}

// loadFacts accepte des clés supplémentaires (le JSON écrit par
// kit/run-dualns.sh en contient) mais n'invente aucune valeur manquante.
func loadFacts(path string) (Facts, error) {
	var raw rawFields
	if err := decode(&raw, path, false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Facts{}, fmt.Errorf("faits absents (%s): fail closed", path)
		}
		return Facts{}, fmt.Errorf("faits illisibles/invalides (%s): %w", path, err)
	}
	facts := Facts{}
	if raw.Namespaces != nil {
		facts.Namespaces = *raw.Namespaces
	}
	if raw.Ifaces != nil {
		facts.Ifaces = *raw.Ifaces
	}
	if raw.Qdisc != nil {
		facts.Qdisc = *raw.Qdisc
	}
	if raw.CapMbps != nil {
		facts.CapMbps = *raw.CapMbps
	}
	if raw.DelayMs != nil {
		facts.DelayMs = *raw.DelayMs
	}
	if raw.ShapeMode != nil {
		facts.ShapeMode = *raw.ShapeMode
	}
	return facts, nil
}

func absent(v string) string {
	if v == "" {
		return "(absent)"
	}
	return v
}

func showList(v []string) string {
	if len(v) == 0 {
		return "(absent)"
	}
	return strings.Join(v, ",")
}

func showFloat(v float64) string {
	if v == 0 {
		return "(absent)"
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// Check compare les six champs déclarés aux faits fournis. Aucune tolérance
// implicite : toute divergence, y compris un fait absent, est un drift.
func Check(c Contract, facts Facts) []Drift {
	var out []Drift
	if !slices.Equal(c.Namespaces, facts.Namespaces) {
		out = append(out, Drift{"namespaces", showList(c.Namespaces), showList(facts.Namespaces)})
	}
	if !slices.Equal(c.Ifaces, facts.Ifaces) {
		out = append(out, Drift{"ifaces", showList(c.Ifaces), showList(facts.Ifaces)})
	}
	if c.Qdisc != facts.Qdisc {
		out = append(out, Drift{"qdisc", absent(c.Qdisc), absent(facts.Qdisc)})
	}
	if c.CapMbps != facts.CapMbps {
		out = append(out, Drift{"cap_mbps", showFloat(c.CapMbps), showFloat(facts.CapMbps)})
	}
	if c.DelayMs != facts.DelayMs {
		out = append(out, Drift{"delay_ms", showFloat(c.DelayMs), showFloat(facts.DelayMs)})
	}
	if c.ShapeMode != facts.ShapeMode {
		out = append(out, Drift{"shape_mode", absent(c.ShapeMode), absent(facts.ShapeMode)})
	}
	return out
}

// Fingerprint — empreinte versionnée des OCTETS du contrat déclaré. C'est la
// même valeur que pkg/campagne écrit dans topology_contract_sha256.
func Fingerprint(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func main() {
	contractPath := flag.String("contract", "kit/topology.contract.json", "declared topology contract (JSON)")
	factsPath := flag.String("facts", "", "supplied topology facts (JSON), read-only")
	fingerprint := flag.Bool("fingerprint", false, "print the contract SHA-256 and exit")
	flag.Parse()

	if *fingerprint {
		sum, err := Fingerprint(*contractPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "topocheck:", err)
			os.Exit(2)
		}
		fmt.Println(sum)
		return
	}
	if *factsPath == "" {
		fmt.Fprintln(os.Stderr, "topocheck: --facts requis (aucun fait inventé)")
		os.Exit(2)
	}
	c, err := Load(*contractPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "topocheck:", err)
		os.Exit(2)
	}
	facts, err := loadFacts(*factsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "topocheck:", err)
		os.Exit(2)
	}
	drifts := Check(c, facts)
	for _, d := range drifts {
		fmt.Println(d)
	}
	if len(drifts) > 0 {
		os.Exit(1)
	}
	fmt.Println("topology ok")
}
