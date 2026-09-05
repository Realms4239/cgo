# Remontée 16/20 — preuve pilote + parité mémoire × interface (design validé)

Date : 2026-09-05. Approche A maintenant, B en extension même pipeline. P0 (placeholders `[[ ]]`) ignoré par décision utilisateur (by design, inchangé).

## 1. Contexte vérifié (archives, pas texte)

- `book/data/runs` : 153 dirs dont 1 `run-smoke` exclu → 152 runs, 412 lignes = 136 valid + 62 degraded + 214 invalid.
- Publié actuel mélange 62 degraded (G2/G6) aux 136 valid dans n=16/12/12 → décision : valid-only strict, n recalculés.
- 214 invalid = 199 G4-suspects (93 %, catégories non-exclusives : zero ⊂ low, G3 recoupe G4) : 157 low (goodput < 50 % cap, surtout CUBIC, Mathis sous perte) dont 20 zero (G0/G1 bring-up, rtt=0 bulk=0), 42 high (P1, bug double-compte veth fixé, `pkg/campagne/prod.go:68-70`), 12 G3 (P1). CPU≥90 : 0. Pas de bug fatal unique.
- Gaps réels : `quarantine.json` ne stocke jamais la porte fautive (`pkg/campagne/writer.go:25-31`) ; `build-stats.js` pointe vers `c:/thesis-cgo/thesis-cgo` mort ; `book/stats.json` fantôme vs `book/data/stats.json` canonique ; `data/runs` racine (3 runs) hors pipeline.

## 2. Fil rouge (recadrage validé)

« Instrument validé + preuve pilote P2 BBR, P3/P4 en bornes exploratoires », pas « CAKE sauve les alertes ». P3/P4 hors résultat central (bornes en annexe). Résumé/abstract réécrits chiffrés à n (best practice, sans « première mesure locale »).

## 3. Mémoire (Section A validée)

- `book/extract-stats.js` : filtre valid-only ; + IQR, p95, IC95 percentile-bootstrap de la médiane (10k resamples, seed 42) ; Mann-Whitney bilatéral α=0,05 pfifo vs cake (P2 BBR valid-only, deadline_ok_pct + small_p95) + Cliff delta Vargha-Delaney (<0,11 négligeable, <0,28 petit, <0,43 moyen, sinon grand) ; miroir du Go, parité vérifiée sur P2 BBR.
- Tableaux P2/P1/P4 : colonne n valid-only partout (ex-16/12/12 republiés tels que recalculés) ; VoIP/JFI sortis du central (annexe/limites, labels théorique / n.a. mono-flux).
- Quarantaine : tableau backfill non-exclusif (157 low dont 20 zero / 42 high / 12 G3) + note G4-Mathis (`BW ≈ MSS/(RTT·√p)` ; CUBIC P2 ≈ 2 Mb/s prédit, P1-CUBIC tient 66 Mb/s à perte 0 → interaction CC × profil = QR2).
- Coût : audit ligne par ligne `wasted_bytes = drops×1448` pour l'outlier P1|cake|bbr (médiane gardée, explication écrite) + tableau sensibilité 6 paliers + limites (extrapolation linéaire, retrans=0 borne basse).
- P3 : phrase 220 < 600 par construction, jugement QDI/small_p95 ; P4 : exploratoire n=4, pas de « CAKE recommandé LEO » ; P1 bloat résiduel (p95 des p95 763 ms, n=55) gardé.
- Biblio 11→~28 (RFC8033 PIE, BBRv2, Flent 2017, WIS2.0/OMM, ARTEC, tarifs) citées en 02/03. `build-stats.js` neutralisé legacy.
- Figures : figR2/R3 barres IQR/IC + n valid-only + mention CUBIC-G4, titres finding.

## 4. Interface (Section B validée)

- Check complet UI × Partie II × CSV 23 col (métriques, G0–G7, P1–P4, qdisc/CC). Écarts corrigés côté texte ou UI (additif).
- Nouveau `pkg/metrics/stats.go` (médiane/IQR/bootstrap seed 42/MW/Cliff, seuils §3), réutilisé par API ; `extract-stats.js` en miroir.
- `GET /api/results` étendu en additif (n, iqr, ic95, p, cliff ; anciens champs intacts) + précalcul cache ; `GET /api/quarantine` + `failed_gates` (futur) + synthèse backfill affichée en Provenance.
- Vue Résultats : méd [IC95] + IQR + n ; épinglage pfifo vs cake P2 BBR avec p + delta ; verdict auto étendu (même moteur Go). Coût : 6 paliers + sensibilité P2 cake|BBR. VoIP/JFI sortis du central comme mémoire.
- TUI : version compacte 80 cols, mêmes chiffres. Captures Web 3 vues + TUI recapturées 1920×1080 sur VM via `kit scan` (secrets env seuls, jamais commités).

## 5. Pipeline / exécution / vérif (Section C validée)

- Source unique `book/data/runs` → `book/data/stats.json` (documenté ; `data/runs` racine + `book/stats.json` ignorés).
- `pkg/campagne/writer.go` : champ `failed_gates` pour runs futurs (hash manifest évolutif, archives passées intactes).
- B (extension, hors prod + reprise) : 9 events P2 BBR (~45 min) + P3-1500ms 6 events 3×2 (~18 min) + `data/link_audit.csv` (site, lien, opérateur, horaire, idle vs loaded p50/p95, grade A–F ; fibre siège + Yas 4G creuse/pointe) + tableau RQ1 calibrant P1/P2.
- Worktree `overhaul1.2.2` sur HEAD, `node_modules` + `GOMODCACHE` réutilisés, zéro download.
- Vérif : `go test ./pkg/metrics/ ./pkg/campagne/ ./pkg/results/ ./pkg/api/` (1 assert/fonction + parité P2) + `node book/extract-stats.js` + `node book/figures-lien.js`.

## 6. Fichiers touchés (plan détaillé = writing-plans)

`pkg/metrics/stats.go` (+test), `pkg/campagne/writer.go` (+test), `pkg/api/server.go` (results/quarantine additifs + cache), web Résultats/Provenance, TUI Résultats, `book/extract-stats.js`, `book/figures-lien.js`, `book/content/{00-front,02-partie1,03-partie2,04-pratique,07-refs,09-annexes}.js`, `book/data/stats.json` (régénéré), `book/build-stats.js` (legacy), `data/link_audit.csv` (B), `shots/` (B), cette spec.

## 7. Hors scope

P0 placeholders inchangés (by design). Pas de traduction MikroTik testée (marquée indicative). Pas de campagne live Windows (observe 501). Aucun secret dans repo/spec/logs.
