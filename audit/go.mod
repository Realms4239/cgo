module github.com/Realms4239/cgo/audit

go 1.25

require (
	github.com/Realms4239/cgo/metrics v0.1.0
	github.com/Realms4239/cgo/model v0.1.1
	github.com/Realms4239/cgo/probe v0.1.0
	github.com/Realms4239/cgo/remediate v0.1.0
)

// Même remplacement local que le module racine et api : le metrics du dépôt
// porte DistributionEMD (tâche 3), absent du v0.1.0 du cache. Aucun changement
// de version, aucune nouvelle dépendance.
replace github.com/Realms4239/cgo/metrics => ../metrics
