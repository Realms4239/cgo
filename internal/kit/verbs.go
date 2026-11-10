package kit

// SuiteVerbs — la liste canonique des verbes que « Suite » et les exécutions
// (Win32 runKind, Fyne dispatchKind) savent déclencher. UN maillon par verbe.
//
// Pourquoi ici : la whitelist vivait dans le GUI (27 verbes) pendant que les
// exécutants en connaissaient 30 — « Suite » refusait deux pas que les
// boutons exécutaient (console:dns, console:tls). Le banc de cette dérive,
// c'est une seule source de vérité qui compile des deux côtés du tag build.
var SuiteVerbs = []string{
	"direct:lock", "direct:saveuser", "direct:open", "direct:suite",
	"bg:scan", "bg:diag", "bg:dns", "bg:tls", "bg:mkkey", "bg:console",
	"bg:ensure", "bg:deploy", "bg:svc", "bg:testbed", "bg:logs",
	"bg:verify", "bg:backup", "bg:snapshot", "bg:netinfo", "bg:vmon",
	"bg:vmoff", "bg:nic-toggle", "bg:nic", "bg:hosttun", "bg:guest",
	"bg:vnet", "bg:guide",
	"console:keysetup", "console:dns", "console:tls",
}

// IsKnownSuiteVerb — le verbe appartient-il à la suite (bouton ET dispatch) ?
// La Suite gui fatale sans verbe connu affiche le remède ; elle ne l'exécute pas.
func IsKnownSuiteVerb(verb string) bool {
	for _, v := range SuiteVerbs {
		if v == verb {
			return true
		}
	}
	return false
}
