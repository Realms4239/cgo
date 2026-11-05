package campagne

import "github.com/Realms4239/cgo/pkg/model"

// Capabilities — la vue hôte rendue au client : files supportées, contrôles
// réellement exécutables, libellés connus et release du noyau. Le client ne
// recopie plus ces listes ; une seule source les compose (API et banc).
type Capabilities struct {
	Qdiscs   []string `json:"qdiscs"`
	CCs      []string `json:"ccs"`
	KnownCCs []string `json:"known_ccs"`
	Kernel   string   `json:"kernel,omitempty"`
	BBRv3    bool     `json:"bbr3"`
}

// CapabilitiesNow — instantané honnête : CCs ne contient que ce que le noyau
// expose, KnownCCs liste les étiquettes acceptées même indisponibles.
func CapabilitiesNow() Capabilities {
	qdiscs := make([]string, 0, len(model.AllQdiscs))
	for _, q := range model.AllQdiscs {
		qdiscs = append(qdiscs, string(q))
	}
	known := make([]string, 0, len(model.KnownCC))
	for _, c := range model.KnownCC {
		known = append(known, string(c))
	}
	return Capabilities{
		Qdiscs:   qdiscs,
		CCs:      DefaultCCs(),
		KnownCCs: known,
		Kernel:   readKernelRelease(),
		BBRv3:    ccPresent(string(model.BBRv3)),
	}
}
