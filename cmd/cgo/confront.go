package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Confrontation eBPF vs tc -s. Le binaire NE fabrique aucun compteur : il lit
// un sceau noyau déjà scellé (CGO_CONFRONT_INPUT), le valide, puis recopie ses
// valeurs exactes dans le rapport. Sceau absent ou invalide → échec fermé.
const (
	confrontSealSchema   = "cgo.kernel-truth-seal/1"
	confrontReportSchema = "cgo.ebpf-vs-tc/1"
)

// Rôles fixes des deux plans : surfaces complémentaires, sans classement.
const (
	confrontRoleTC   = "accounting surface (kernel counters: drops, overlimits, backlog, delay)"
	confrontRoleEBPF = "temporal surface (per-packet retrans, srtt, bytes at near-wire speed)"
)

type kernelTruthSeal struct {
	Schema                 string  `json:"schema"`
	AuditID                string  `json:"audit_id"`
	Source                 string  `json:"source"`
	ObjectSHA256           string  `json:"object_sha256,omitempty"`
	RetransPayloadBytes    uint64  `json:"retrans_payload_bytes,omitempty"`
	RetransEvents          uint64  `json:"retrans_events"`
	QdiscDropPackets       *uint64 `json:"qdisc_drop_packets,omitempty"`
	QdiscDropPacketsAbsent bool    `json:"qdisc_drop_packets_absent,omitempty"`
	TCInterface            string  `json:"tc_interface,omitempty"`
	TCQdisc                string  `json:"tc_qdisc,omitempty"`
	TCDropsBefore          uint64  `json:"tc_drops_before"`
	TCDropsAfter           uint64  `json:"tc_drops_after"`
	Note                   string  `json:"note,omitempty"`
}

type confrontReport struct {
	Schema        string              `json:"schema"`
	AuditID       string              `json:"audit_id"`
	Source        string              `json:"source"`
	KernelTruth   confrontKernel      `json:"kernel_truth"`
	TCs           confrontTC          `json:"tc_s"`
	Confrontation confrontVerdict     `json:"confrontation"`
	TruthPlanes   confrontTruthPlanes `json:"truth_planes"`
	Note          string              `json:"note,omitempty"`
}

// confrontTruthPlanes — les deux surfaces de vérité, sans hiérarchie : compteurs
// noyau côté tc/qdisc, per-packet near-wire-speed côté eBPF.
type confrontTruthPlanes struct {
	TCQdisc string `json:"tc_qdisc"`
	EBPF    string `json:"ebpf"`
}

type confrontKernel struct {
	ObjectSHA256           string  `json:"object_sha256,omitempty"`
	RetransPayloadBytes    uint64  `json:"retrans_payload_bytes,omitempty"`
	RetransEvents          uint64  `json:"retrans_events"`
	QdiscDropPackets       *uint64 `json:"qdisc_drop_packets,omitempty"`
	QdiscDropPacketsAbsent bool    `json:"qdisc_drop_packets_absent,omitempty"`
	Note                   string  `json:"note,omitempty"`
}

type confrontTC struct {
	Interface   string `json:"interface,omitempty"`
	Qdisc       string `json:"qdisc,omitempty"`
	DropsBefore uint64 `json:"drops_before"`
	DropsAfter  uint64 `json:"drops_after"`
	DropsDelta  uint64 `json:"drops_delta"`
}

type confrontVerdict struct {
	EBPFRetransEvents uint64 `json:"ebpf_retrans_events"`
	TCDropsDelta      uint64 `json:"tc_drops_delta"`
	Match             bool   `json:"match"`
}

// loadSeal lit et valide le sceau noyau. Tout manque est une erreur : jamais un
// zéro par défaut (absence-never-zero).
func loadSeal(path string) (kernelTruthSeal, error) {
	var s kernelTruthSeal
	raw, err := os.ReadFile(path)
	if err != nil {
		return s, fmt.Errorf("sceau noyau absent ou illisible (%s): %w", path, err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		return s, fmt.Errorf("sceau noyau illisible (%s): %w", path, err)
	}
	for _, k := range []string{"schema", "audit_id", "source", "retrans_events", "tc_drops_before", "tc_drops_after"} {
		v, ok := keys[k]
		if !ok {
			return s, fmt.Errorf("sceau noyau incomplet (%s): %s absent", path, k)
		}
		// null est une absence déguisée : json.Unmarshal le transformerait
		// en zéro pour les compteurs (trou absence-never-zero). Refus fermé.
		if string(v) == "null" {
			return s, fmt.Errorf("sceau noyau incomplet (%s): %s est null", path, k)
		}
	}

	// qdisc_drop_packets : compteur numérique non-null OU absence explicite
	// (drapeau true). Jamais ni l'un ni l'autre, jamais null.
	dropRaw, hasDrop := keys["qdisc_drop_packets"]
	if hasDrop && string(dropRaw) == "null" {
		return s, fmt.Errorf("sceau noyau incomplet (%s): qdisc_drop_packets est null", path)
	}
	absentRaw, hasAbsent := keys["qdisc_drop_packets_absent"]
	var absentFlag bool
	if hasAbsent {
		if string(absentRaw) == "null" {
			return s, fmt.Errorf("sceau noyau incomplet (%s): qdisc_drop_packets_absent est null", path)
		}
		if err := json.Unmarshal(absentRaw, &absentFlag); err != nil {
			return s, fmt.Errorf("sceau noyau: qdisc_drop_packets_absent invalide (%s): %w", path, err)
		}
	}
	if !hasDrop && !absentFlag {
		return s, fmt.Errorf("sceau noyau incomplet (%s): qdisc_drop_packets absent sans qdisc_drop_packets_absent:true", path)
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return s, fmt.Errorf("sceau noyau invalide (%s): %w", path, err)
	}
	if s.Schema != confrontSealSchema {
		return s, fmt.Errorf("sceau noyau: schema %q attendu %q", s.Schema, confrontSealSchema)
	}
	if s.Source != "ebpf" {
		return s, fmt.Errorf("sceau noyau: source %q — seul ebpf est une vérité noyau scellée", s.Source)
	}
	if s.AuditID == "" {
		return s, fmt.Errorf("sceau noyau: audit_id absent")
	}
	if s.TCDropsAfter < s.TCDropsBefore {
		return s, fmt.Errorf("sceau noyau: tc_drops_after < tc_drops_before")
	}
	return s, nil
}

// writeConfrontationReport écrit le rapport de confrontation depuis le sceau.
// Écriture atomique (fichier temporaire + rename) : aucun fichier partiel.
func writeConfrontationReport(inPath, outPath string) error {
	s, err := loadSeal(inPath)
	if err != nil {
		return err
	}
	delta := s.TCDropsAfter - s.TCDropsBefore
	rep := confrontReport{
		Schema:  confrontReportSchema,
		AuditID: s.AuditID,
		Source:  s.Source,
		KernelTruth: confrontKernel{
			ObjectSHA256:           s.ObjectSHA256,
			RetransPayloadBytes:    s.RetransPayloadBytes,
			RetransEvents:          s.RetransEvents,
			QdiscDropPackets:       s.QdiscDropPackets,
			QdiscDropPacketsAbsent: s.QdiscDropPacketsAbsent,
			Note:                   s.Note,
		},
		TCs: confrontTC{
			Interface:   s.TCInterface,
			Qdisc:       s.TCQdisc,
			DropsBefore: s.TCDropsBefore,
			DropsAfter:  s.TCDropsAfter,
			DropsDelta:  delta,
		},
		Confrontation: confrontVerdict{
			EBPFRetransEvents: s.RetransEvents,
			TCDropsDelta:      delta,
			Match:             s.RetransEvents == delta,
		},
		TruthPlanes: confrontTruthPlanes{
			TCQdisc: confrontRoleTC,
			EBPF:    confrontRoleEBPF,
		},
		Note: s.Note,
	}
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return fmt.Errorf("confrontation: rendu JSON: %w", err)
	}
	b = append(b, '\n')

	tmp := outPath + ".tmp"
	defer os.Remove(tmp)
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("confrontation: écriture temporaire: %w", err)
	}
	if err := os.Rename(tmp, outPath); err != nil {
		return fmt.Errorf("confrontation: publication: %w", err)
	}
	return nil
}

// runConfrontOut est le point d'entrée testable du drapeau --confront-out.
// outPath vide (= drapeau absent) est un no-op : aucun fichier n'est écrit.
func runConfrontOut(outPath, inPath string) int {
	if outPath == "" {
		return 0
	}
	if err := writeConfrontationReport(inPath, outPath); err != nil {
		fmt.Fprintln(os.Stderr, "confront:", err)
		return 2 // échec fermé : le sceau n'a pas pu être lu/validé
	}
	fmt.Printf("confront → %s (sceau %s)\n", outPath, inPath)
	return 0
}
