package audit

// Presets RQ1 — le protocole d'audit terrain du mémoire (3 mesures opérateur
// à Madagascar), matérialisé pour que l'opérateur clique au lieu de recopier
// des commandes depuis un document. Mêmes paramètres que le plan d'audits :
// fibre siège hors production, Yas 4G heure creuse puis pointe.
type Preset struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Site     string `json:"site"`
	LinkType string `json:"link_type"`
	Provider string `json:"provider"`
	Duration int    `json:"duration"`
	Target   string `json:"target"`
	Notes    string `json:"notes"`
}

func Presets() []Preset {
	return []Preset{
		{
			ID: "fibre-siege", Label: "Fibre siège — hors production",
			Site: "siege-ampandrianomby", LinkType: "fiber", Provider: "yas",
			Duration: 60, Target: "8.8.8.8",
			Notes: "hors heures de production, poste autorisé ; câble, pas de partage",
		},
		{
			ID: "4g-creuse", Label: "Yas 4G — heure creuse",
			Site: "tana-partage-creuse", LinkType: "4g", Provider: "yas",
			Duration: 60, Target: "8.8.8.8",
			Notes: "partage de connexion du téléphone autorisé ; noter l'heure exacte en commentaire opérateur",
		},
		{
			ID: "4g-pointe", Label: "Yas 4G — heure pointe",
			Site: "tana-partage-pointe", LinkType: "4g", Provider: "yas",
			Duration: 60, Target: "8.8.8.8",
			Notes: "même poste, même emplacement que l'heure creuse ; la paire creuse/pointe fait le tableau RQ1",
		},
	}
}
