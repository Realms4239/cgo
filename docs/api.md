# Référence API — Meteolink (cgo)

> Vérifiée contre `pkg/api/server.go`, `sse.go` et `translate.go` (worktree
> certify). Tout corps JSON est limité à 1 Mio (`http.MaxBytesReader`).
> Les endpoints `/api/doctor` et `/api/watch` **n'existent pas** dans cette
> version du handler ; le health check ne renvoie que `{"ok": true}`.

## Santé & état

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/health` | GET | — | `{"ok": true}` | — |
| `/api/state` | GET | — | snapshot courant (`GetSnap`) ; `{"running": false}` si non câblé | — |
| `/api/diagnostics` | GET | — | `{"hub":"ok","time":"<RFC3339 UTC>"}` | — |
| `/api/stream` | GET (SSE) | — | flux 10 Hz : replay (`Last-Event-ID`, anneau 2048) puis live ; frame complète pour client frais ; event `backpressure` | `503` > 64 abonnés ; `500` flusher non supporté |

## Profils

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/profiles` | GET | — | `{"profiles":[{id, capacity_mbps, delay_ms, jitter_ms, loss_pct, imported}]}` trié par id ; `imported` = tout id hors P1/P2/P3 | — |
| `/api/profile/list` | GET | — | map `model.Profiles` (rechargée du disque) | — |
| `/api/profile/import` | POST | CSV (`text/csv` ou corps ne commençant pas par `{`) : `id,capacity_mbps,delay_ms,jitter_ms,loss_pct` (en-tête optionnel sauté) **ou** JSON `model.Profile` | `{"ok": true, "profile": {...}}` | `400` bad body / bad csv / `id` manquant ; `500` échec d'import |

## Façonnage du bord

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/shape` | GET | — | `{"applied": false}` ou `{"applied": true, "qdisc", "capacity_mbps", "since"}` | — |
| `/api/shape` | POST | `{qdisc, capacity_mbps, delay_ms, jitter_ms, loss_pct}` | `{"qdisc", "capacity_mbps", "since"}` + événement journal | `400` JSON invalide, qdisc hors `cake\|fq_codel\|pfifo_fast\|none`, capacité hors **1–1000**, conditions hors délai **0–600 ms** / gigue **0–100 ms** / perte **0–10 %** ; `409` campagne active ; `503` moteur de shape non câblé sur cet hôte ; `500` échec d'application tc |

## Campagne

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/run/start` | POST | `{profiles, reps, deadline_ms, target}` | `{"started": true}` + événement journal | `400` JSON invalide ; `503` moteur de run non câblé ; `409` toute erreur de `StartFn` (campagne déjà active, profils inconnus…) |
| `/api/run/stop` | POST | — | `{"stopped": true}` + événement journal | `503` moteur non câblé |
| `/api/results?run=` | GET | — | `{"available": true, "groups": [...]}` (médianes, quarantaine, best) | `{"available": false, "reason": …}` si aucun gel |
| `/api/run/rows?run=` | GET | — | `{"run", "rows": [lignes brutes aqm_eval.csv]}` | `400` id de run vide ou contenant `/`, `\`, `.` (anti-traversal) ; `404` run introuvable ou vide |
| `/api/events` | GET | — | `{"events": [{ts, kind, msg}, …]}` — anneau des 50 derniers | — |

## Intégrité, rapport, figures

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/integrity` | GET | — | `{available, runs, manifests, valid, quarantined, run_ids, sha256, hash8}` | `{"available": false}` si aucun run |
| `/api/report/export?format=md\|csv` | GET | — | tableau Markdown (défaut) ou CSV en pièce jointe `report.csv` + ligne de provenance `hash8` | `404` aucune donnée |
| `/api/figures/regen` | POST | — | `{"ok": true}` — régénère `data/figures` depuis `data/runs` | `500` échec de génération |
| `/api/figures/` | GET | — | fichiers statiques de `data/figures/` | — |

## Replay

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/replay/list` | GET | — | `{"runs": [ids]}` — runs avec `aqm_eval.csv` | — |
| `/api/replay/stream?run=` | GET (SSE) | — | rejoue chaque ligne de `aqm_eval.csv` en event JSON, **400 ms/ligne**, `phase: "replay"` | `400` run manquant ; `404` introuvable/vide ; `500` flusher non supporté |

## Audit de lien

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/audit/start` | POST | `{site, link_type, provider, duration, target}` ; `duration` défaut 30 s, `target` défaut `8.8.8.8` | `{"started": true}` — exécution **asynchrone** (goroutine détachée), résultat appendu à `data/link_audit.csv` | `400` JSON invalide ; `409` audit déjà en cours |
| `/api/audit/status` | GET | — | `{"running": bool, "last": <Result\|null>}` | — |
| `/api/audit/list` | GET | — | `{"audits": [lignes de data/link_audit.csv]}` | — (retourne `[]` si absent) |

## Traduction matérielle

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/hardware/translate?profile=P2` | GET | — | `{"recommendation": "…"}` — qdisc recommandé : `Best` mesuré si disponible, sinon `cake` pour P1, `fq_codel` sinon | `{"available": false, "reason": "aucune donnée — lancez campagne"}` (HTTP 200) |

## Divers

- **404 JSON** pour tout `/api/*` inconnu : `{"error": "not found", "path"}` —
  enregistré avant le catch-all SPA.
- **SPA** : toute autre route sert `web/frontend.FS` (fichiers embarqués).
