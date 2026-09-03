# Référence API — Meteolink (cgo)

> Vérifiée contre `pkg/api/server.go`, `sse.go` et `translate.go`. Tout corps
> JSON est limité à 1 Mio (`http.MaxBytesReader`).
> **Mode observation** (Windows, ou `--mode observe`) : les endpoints de
> contrôle — `POST /api/shape`, `POST /api/run/start`, `POST /api/watch` —
> répondent `501` avec un message pointant vers `docs/deploy.md`. L'audit et
> la consultation restent fonctionnels.

## Santé & état

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/health` | GET | — | `{"ok": true, "version": "1.2.0", "mode": "full\|observe\|\"\"}` | — |
| `/api/doctor` | GET | — | `{"mode", "checks": [{name, status: ok\|warn\|fail, detail}]}` — mêmes capacités que `cgo doctor` (os, tc, cap_net_admin, bbr, ping) | — |
| `/api/state` | GET | — | snapshot courant (`GetSnap`) ; `{"running": false}` si non câblé | — |
| `/api/diagnostics` | GET | — | `{"hub":"ok","time":"<RFC3339 UTC>"}` | — |
| `/api/stream` | GET (SSE) | — | flux 10 Hz : replay (`Last-Event-ID`, anneau 2048) puis live ; frame complète pour client frais **ou** client périmé (ID antérieur à l'anneau) ; event `backpressure` | `503` > 64 abonnés ; `500` flusher non supporté |
| `/api/schema` | GET | — | `{"params":[...]}` — registre unique des bornes/défauts (le client rend les champs depuis lui) | — |
| `/api/cost/tiers` | GET | — | `{"tiers":[...],"default":"yas-month-4.5gb"}` — paliers tarifaires réels | — |
| `/api/burst` | POST | `{cc: cubic\|bbr, seconds: 2–10 (défaut 4)}` | `{"ok":true,"cc","seconds"}` — sonde bulk à travers le bord façonné, bloquant | `400` JSON invalide, cc/seconds hors bornes ; `409` campagne active ; `501` mode observation ; `503` moteur non câblé ; `500` échec sonde |

## Profils

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/profiles` | GET | — | `{"profiles":[{id, capacity_mbps, delay_ms, jitter_ms, loss_pct, imported}]}` trié par id ; `imported` = id non natif (natifs : P1–P4) | — |
| `/api/profile/list` | GET | — | map `model.Profiles` (rechargée du disque) | — |
| `/api/profile/import` | POST | CSV (`text/csv` ou corps ne commençant pas par `{`) : `id,capacity_mbps,delay_ms,jitter_ms,loss_pct` (en-tête optionnel sauté) **ou** JSON `model.Profile` | `{"ok": true, "profile": {...}}` | `400` bad body / bad csv / `id` manquant ; `500` échec d'import |

## Façonnage du bord

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/shape` | GET | — | `{"applied": false}` ou `{"applied": true, "qdisc", "capacity_mbps", "since"}` | — |
| `/api/shape` | POST | `{qdisc, capacity_mbps, delay_ms, jitter_ms, loss_pct}` | `{"qdisc", "capacity_mbps", "since"}` + événement journal | `400` JSON invalide, qdisc hors `cake\|fq_codel\|pfifo_fast\|none`, capacité hors **1–1000**, conditions hors délai **0–600 ms** / gigue **0–100 ms** / perte **0–10 %** ; `409` campagne active ; `501` mode observation ; `503` moteur de shape non câblé sur cet hôte ; `500` échec d'application tc |

## Surveillance continue

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/watch` | POST | `{"on": true\|false}` | `{"watch": bool}` ; boucle non intrusive ping + petits objets (500 ms), phase `surveil` ; refusée pendant une campagne (l'inverse : la campagne auto-arrête la surveillance) | `400` JSON invalide ; `409` campagne active ; `501` mode observation ; `503` moteur non câblé ; `500` échec du moteur |

## Campagne

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/run/start` | POST | `{profiles, reps, deadline_ms, target}` | `{"started": true}` + événement journal | `400` JSON invalide ; `503` moteur de run non câblé ; `409` toute erreur de `StartFn` (campagne déjà active, profils inconnus…) |
| `/api/run/stop` | POST | — | `{"stopped": true}` + événement journal | `503` moteur non câblé |
| `/api/run/skip` | POST | — | `{"skipped": true}` — coupe la cellule en cours sans arrêter la matrice (reprise possible) | `503` moteur non câblé |
| `/api/results?run=` | GET | — | `{"available": true, "groups": [...]}` (médianes sur lignes non `invalid`, quarantaine, best) | `{"available": false, "reason": …}` si aucun gel |
| `/api/results/delta?cell=P\|q\|cc` | GET | — | `{"available":true,"previous_run","current_run","delta":{small_p95_pct,rtt_p95_pct,goodput_pct}}` — dérive entre les deux derniers runs | `400` cellule mal formée ; `{"available":false}` si < 2 runs ou cellule absente |
| `/api/run/rows?run=` | GET | — | `{"run", "rows": [lignes brutes aqm_eval.csv]}` | `400` id de run vide ou contenant `/`, `\`, `.` (anti-traversal) ; `404` run introuvable ou vide |
| `/api/events` | GET | — | `{"events": [{ts, kind, msg}, …]}` — anneau des 50 derniers | — |

## Intégrité, rapport, figures

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/integrity` | GET | — | `{available, runs, manifests, valid, quarantined, run_ids (récents d'abord), breakdown: [{run, rows, valid, quarantined}], updated: <RFC3339 mtime dernier csv>, sha256, hash8}` | `{"available": false}` si aucun run |
| `/api/quarantine?run=` | GET | — | `{"quarantines": [{run, event_id, profile, qdisc, cc, gate_status}]}` — `run` vide = tous runs (récents d'abord), `[]` si aucune | `400` id invalide (anti-traversal) ; `404` run inconnu |
| `/api/report/export?format=md\|csv\|sh` | GET | — | tableau Markdown (défaut), CSV en pièce jointe `report.csv` + ligne de provenance `hash8`, ou **script tc** en pièce jointe `aqm-recipe.sh` : par profil, la meilleure cellule gelée devient une recette rejouable (reset + netem aux conditions du profil + shaper au goodput mesuré ×0,9 — même règle que la suggestion CLI), sans `sudo` préfixé (l'opérateur lance en root) | `404` aucune donnée |
| `/api/figures/regen` | POST | — | `{"ok": true}` — régénère `data/figures` depuis `data/runs` | `500` échec de génération |
| `/api/figures/` | GET | — | fichiers statiques de `data/figures/` | — |

## Replay

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/replay/list` | GET | — | `{"runs": [ids]}` — runs avec `aqm_eval.csv` | — |
| `/api/replay/stream?run=` | GET (SSE) | — | rejoue chaque ligne de `aqm_eval.csv` en event JSON, **400 ms/ligne**, `phase: "replay"` | `400` run manquant ou contenant `/`, `\`, `.` (anti-traversal) ; `404` introuvable/vide ; `500` flusher non supporté |

## Audit de lien

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/audit/start` | POST | `{site, link_type, provider, duration, target, small_url?}` ; `duration` défaut 30 s, `target` défaut `8.8.8.8` ; sans `small_url`, le petit objet n'est pas mesuré (0 + note, jamais de valeur synthétique) | `{"started": true}` — exécution **asynchrone** (goroutine détachée), résultat appendu à `data/link_audit.csv` | `400` JSON invalide ; `409` audit déjà en cours |
| `/api/audit/status` | GET | — | `{"running": bool, "last": <Result\|null>}` | — |
| `/api/audit/list` | GET | — | `{"audits": [lignes de data/link_audit.csv]}` | — (retourne `[]` si absent) |
| `/api/audit/toprofile` | POST | — | `{"ok":true,"profile"}` — dernière ligne de `link_audit.csv` → profil rejouable (`capacity` = throughput, `delay` = rtt_idle_p95, replis 20/100 documentés) | `404` aucun audit / audit vide ; `500` échec d'import |

## Traduction matérielle

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/hardware/translate?profile=P2` | GET | — | `{"recommendation": "…"}` — qdisc recommandé : `Best` mesuré si disponible, sinon `cake` pour P1, `fq_codel` sinon | `{"available": false, "reason": "aucune donnée — lancez campagne"}` (HTTP 200) |

## Divers

- **404 JSON** pour tout `/api/*` inconnu : `{"error": "not found", "path"}` —
  enregistré avant le catch-all SPA.
- **SPA** : toute autre route sert `web/frontend.FS` (fichiers embarqués).
