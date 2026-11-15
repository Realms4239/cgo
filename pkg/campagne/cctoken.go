package campagne

import (
	"github.com/Realms4239/cgo/capabilities"
	"github.com/Realms4239/cgo/probe"
)

// init — pose la traduction vie entière du process : la cellule bbr3 sur
// un hôte provisionné parle « bbr » à TCP_CONGESTION (le seul token v3 de
// cet arbre), l'étiquette enregistrée reste bbr3.
func init() {
	probe.SetSocketToken(func(cc string) string {
		if cc == "bbr3" && capabilities.BBRv3Token() == "bbr" {
			return "bbr"
		}
		return cc
	})
}
