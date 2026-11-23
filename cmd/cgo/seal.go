package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/audit"
)

// qdisc_drop_packets=0 n'est PAS une mesure de zéro perte : c'est l'angle mort
// nommé du filtre de raisons du noyau. Le mot est recopié dans le sceau pour
// qu'aucun lecteur ne le prenne pour une absence de congestion.
const sealZeroDropNote = "qdisc_drop_packets=0 : angle mort nommé du filtre de raisons du noyau (kernel reason-filter blind spot) — ne pas lire comme une absence de perte (not zero loss)"

// sealAbsentDropNote — l'audit 6.8-cake est sérialisé sans le champ
// qdisc_drop_packets : l'absence est un fait, pas une mesure à zéro.
const sealAbsentDropNote = "qdisc_drop_packets absent : angle mort nommé du filtre de raisons du noyau 6.8-cake (kernel reason-filter blind spot) — absence explicite, PAS une mesure à zéro (not a measured zero)"

// tcDroppedRx — une seule forme : `dropped <N>`. Un capture ambigu (0 ou
// plusieurs occurrences) est refusé fermé : on ne devine jamais le compteur.
var tcDroppedRx = regexp.MustCompile(`\bdropped[[:space:]]+([0-9]+)\b`)

// requireJSONFields — chaque clé doit être présente et non null. json.Unmarshal
// transforme silencieusement null en zéro pour un compteur : c'est le trou
// absence-never-zero, refusé ici plutôt que lu comme une mesure à zéro.
func requireJSONFields(raw json.RawMessage, keys ...string) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("objet JSON illisible: %w", err)
	}
	for _, k := range keys {
		v, ok := m[k]
		if !ok || string(v) == "null" {
			return fmt.Errorf("%s absent ou null", k)
		}
	}
	return nil
}

// tcQdiscHeaderRx — en-tête `qdisc <kind> ... dev <iface> ...` d'un bloc de
// `tc -s qdisc show`. Les types et ifaces sont capturés tels quels, sans
// recherche partielle : on exige une correspondance exacte avec les labels.
var tcQdiscHeaderRx = regexp.MustCompile(`^qdisc[[:space:]]+([^[:space:]]+)[[:space:]].*\bdev[[:space:]]+([^[:space:]]+)`)

// parseTCDrops lit `dropped N` dans un capture texte `tc -s qdisc show` et
// vérifie que le capture confirme LUI-MÊME les labels demandés : une ligne
// `qdisc <kind> ... dev <iface>` doit exister et le compteur doit appartenir à
// ce bloc. --iface/--qdisc ne sont jamais recollés sur un capture étranger.
// Échec fermé sur absence, divergence ou ambiguïté.
func parseTCDrops(path, iface, qdisc string) (uint64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("capture tc illisible (%s): %w", path, err)
	}
	text := string(raw)

	// Garde conservée : exactement un `dropped` dans tout le capture.
	all := tcDroppedRx.FindAllStringSubmatch(text, -1)
	if len(all) != 1 {
		return 0, fmt.Errorf("capture tc (%s): dropped introuvable ou ambigu (%d correspondances)", path, len(all))
	}

	// Découpage en blocs : chaque ligne `qdisc ` ouvre un bloc ; les lignes
	// suivantes lui appartiennent jusqu'au prochain en-tête qdisc.
	type qdiscBlock struct {
		kind, dev string
		body      []string
	}
	var blocks []qdiscBlock
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "qdisc ") {
			kind, dev := "", ""
			if m := tcQdiscHeaderRx.FindStringSubmatch(line); m != nil {
				kind, dev = m[1], m[2]
			}
			blocks = append(blocks, qdiscBlock{kind: kind, dev: dev})
			continue
		}
		if len(blocks) > 0 {
			blocks[len(blocks)-1].body = append(blocks[len(blocks)-1].body, line)
		}
	}

	// Les labels demandés doivent être confirmés par le capture, pas supposés.
	var target *qdiscBlock
	for i := range blocks {
		if blocks[i].dev == iface && blocks[i].kind == qdisc {
			target = &blocks[i]
			break
		}
	}
	if target == nil {
		return 0, fmt.Errorf("capture tc (%s): labels iface=%q qdisc=%q absents ou divergents — aucune ligne `qdisc %s ... dev %s` confirmée", path, iface, qdisc, qdisc, iface)
	}

	// Le compteur doit venir du bloc cible, pas d'un autre qdisc du capture.
	m := tcDroppedRx.FindAllStringSubmatch(strings.Join(target.body, "\n"), -1)
	if len(m) != 1 {
		return 0, fmt.Errorf("capture tc (%s): dropped sur le qdisc %q dev %q introuvable ou ambigu (%d correspondances)", path, qdisc, iface, len(m))
	}
	n, err := strconv.ParseUint(m[0][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("capture tc (%s): dropped invalide: %w", path, err)
	}
	return n, nil
}

// buildSeal produit `cgo.kernel-truth-seal/1` UNIQUEMENT depuis des entrées
// réelles déjà scellées : un rapport d'audit (télémétrie ebpf) et deux captures
// tc -s. Aucun compteur n'est fabriqué, aucun chemin par défaut n'est supposé.
// Toute valeur manquante, ambiguë ou nulle est un échec fermé.
func buildSeal(auditPath, tcBeforePath, tcAfterPath, iface, qdisc, outPath string) error {
	if auditPath == "" || tcBeforePath == "" || tcAfterPath == "" || iface == "" || qdisc == "" || outPath == "" {
		return fmt.Errorf("kerneltruth seal: --audit, --tc-before, --tc-after, --iface, --qdisc et --out sont requis (aucun défaut)")
	}

	raw, err := os.ReadFile(auditPath)
	if err != nil {
		return fmt.Errorf("audit illisible (%s): %w", auditPath, err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return fmt.Errorf("audit JSON illisible (%s): %w", auditPath, err)
	}
	if err := requireJSONFields(raw, "audit_id", "kernel_telemetry"); err != nil {
		return fmt.Errorf("audit (%s): %w", auditPath, err)
	}
	telRaw := top["kernel_telemetry"]
	if err := requireJSONFields(telRaw, "source", "available", "retrans_payload_bytes", "flows"); err != nil {
		return fmt.Errorf("kernel_telemetry (%s): %w", auditPath, err)
	}

	// qdisc_drop_packets peut être ABSENT (ancienne sérialisation omitempty des
	// audits scellés). On lit le brut : jamais de coercition absent→0, jamais
	// de null accepté comme mesure (null est une absence déguisée).
	var tel map[string]json.RawMessage
	if err := json.Unmarshal(telRaw, &tel); err != nil {
		return fmt.Errorf("kernel_telemetry illisible (%s): %w", auditPath, err)
	}
	dropRaw, dropPresent := tel["qdisc_drop_packets"]
	if dropPresent && string(dropRaw) == "null" {
		return fmt.Errorf("kernel_telemetry (%s): qdisc_drop_packets est null (absence déguisée, jamais une mesure à zéro)", auditPath)
	}
	var qdiscDrops *uint64
	if dropPresent {
		var n uint64
		if err := json.Unmarshal(dropRaw, &n); err != nil {
			return fmt.Errorf("kernel_telemetry (%s): qdisc_drop_packets invalide: %w", auditPath, err)
		}
		qdiscDrops = &n
	}

	// Garde de présence par flux : un retrans_events null sommé en silence
	// donnerait un retrans_events trop faible (trou absence-never-zero).
	var flowGuard struct {
		Flows []struct {
			RetransEvents json.RawMessage `json:"retrans_events"`
		} `json:"flows"`
	}
	if err := json.Unmarshal(telRaw, &flowGuard); err != nil {
		return fmt.Errorf("kernel_telemetry.flows illisible (%s): %w", auditPath, err)
	}
	for i, f := range flowGuard.Flows {
		if len(f.RetransEvents) == 0 || string(f.RetransEvents) == "null" {
			return fmt.Errorf("flows[%d].retrans_events absent ou null", i)
		}
	}

	var rep audit.Report
	if err := json.Unmarshal(raw, &rep); err != nil {
		return fmt.Errorf("audit illisible (%s): %w", auditPath, err)
	}
	kt := rep.KernelTelemetry
	if kt == nil {
		return fmt.Errorf("kernel_telemetry absent ou invalide (%s)", auditPath)
	}
	if kt.Source != "ebpf" {
		return fmt.Errorf("kerneltruth seal: source %q dans l'audit — seul ebpf est une vérité noyau scellée", kt.Source)
	}
	if !kt.Available {
		return fmt.Errorf("kerneltruth seal: available=false — pas de vérité noyau scellée")
	}
	if rep.AuditID == "" {
		return fmt.Errorf("kerneltruth seal: audit_id absent")
	}
	if kt.ObjectSHA256 == "" {
		return fmt.Errorf("kerneltruth seal: object_sha256 absent")
	}
	if len(kt.Flows) == 0 {
		return fmt.Errorf("kerneltruth seal: flows vide — aucun flux retransmis")
	}
	var retransEvents uint64
	for _, f := range kt.Flows {
		retransEvents += f.RetransEvents
	}

	before, err := parseTCDrops(tcBeforePath, iface, qdisc)
	if err != nil {
		return err
	}
	after, err := parseTCDrops(tcAfterPath, iface, qdisc)
	if err != nil {
		return err
	}
	if after < before {
		return fmt.Errorf("kerneltruth seal: tc drops_after=%d < drops_before=%d", after, before)
	}

	seal := kernelTruthSeal{
		Schema:                 confrontSealSchema,
		AuditID:                rep.AuditID,
		Source:                 "ebpf",
		ObjectSHA256:           kt.ObjectSHA256,
		RetransPayloadBytes:    kt.RetransPayloadBytes,
		RetransEvents:          retransEvents,
		QdiscDropPackets:       qdiscDrops,
		QdiscDropPacketsAbsent: !dropPresent,
		TCInterface:            iface,
		TCQdisc:                qdisc,
		TCDropsBefore:          before,
		TCDropsAfter:           after,
	}
	if qdiscDrops != nil && *qdiscDrops == 0 {
		seal.Note = sealZeroDropNote
	} else if !dropPresent {
		seal.Note = sealAbsentDropNote
	}
	b, err := json.MarshalIndent(seal, "", "  ")
	if err != nil {
		return fmt.Errorf("kerneltruth seal: rendu JSON: %w", err)
	}
	b = append(b, '\n')

	// Écriture atomique + autovalidation : on ne publie que ce que le
	// consommateur (loadSeal) acceptera. Aucun fichier partiel.
	tmp := outPath + ".tmp"
	defer os.Remove(tmp)
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("kerneltruth seal: écriture temporaire: %w", err)
	}
	if _, err := loadSeal(tmp); err != nil {
		return fmt.Errorf("kerneltruth seal: autovalidation: %w", err)
	}
	if err := os.Rename(tmp, outPath); err != nil {
		return fmt.Errorf("kerneltruth seal: publication: %w", err)
	}
	return nil
}

func runKernelTruthSeal(auditPath, tcBeforePath, tcAfterPath, iface, qdisc, outPath string) int {
	if err := buildSeal(auditPath, tcBeforePath, tcAfterPath, iface, qdisc, outPath); err != nil {
		fmt.Fprintln(os.Stderr, "kerneltruth seal:", err)
		return 2
	}
	fmt.Printf("kerneltruth seal → %s (audit %s)\n", outPath, auditPath)
	return 0
}

func runKernelTruthSealCmd(args []string) int {
	fs := flag.NewFlagSet("kerneltruth seal", flag.ContinueOnError)
	auditPath := fs.String("audit", "", "rapport d'audit scellé (JSON, source ebpf)")
	before := fs.String("tc-before", "", "capture texte `tc -s qdisc show` avant")
	after := fs.String("tc-after", "", "capture texte `tc -s qdisc show` après")
	iface := fs.String("iface", "", "interface tc (ex. e-wan)")
	qdisc := fs.String("qdisc", "", "qdisc cible (ex. cake)")
	out := fs.String("out", "", "chemin du sceau à publier")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	return runKernelTruthSeal(*auditPath, *before, *after, *iface, *qdisc, *out)
}
