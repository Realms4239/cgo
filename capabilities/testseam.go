package capabilities

// TestSeam — le point d'entrée public pour les tests qui doivent simuler un
// noyau : la liste des CC enregistrés et la Release lue dans /proc. Les tests
// du module `capabilities` l'utilisent ; les tests du moteur de campagne (hors
// module) en ont besoin aussi — d'où un export explicite plutôt qu'un état
// de paquet partagé.
//
// La règle du module est inchangée : DefaultCCs liste cubic/bbr toujours,
// bbr3 seulement si le noyau l'enregistre. Restore() rend l'hôte réel.
type TestSeam struct {
	restore func()
}

// StubHost — remplace la liste CC et la release, rend une fonction de
// restauration. À utiliser avec defer.
func StubHost(ccList []string, release string) func() {
	oldCC, oldRel := readCCList, readKernelRelease
	readCCList = func() []string { return ccList }
	readKernelRelease = func() string { return release }
	return func() {
		readCCList, readKernelRelease = oldCC, oldRel
	}
}
