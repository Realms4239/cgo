# Référence API — Meteolink (cgo)

> Vérifiée contre `pkg/api/server.go`, `sse.go` et `translate.go`. Tout corps
> JSON est limité à 1 Mio (`http.MaxBytesReader`).
> Les mutations (`POST`) refusent `403` une requête navigateur cross-site
> (`Sec-Fetch-Site: cross-site`) ; les clients directs (curl, CLI) ne sont pas
> concernés.
> **Mode observation** (Windows, ou `--mode observe`) : les endpoints de
> contrôle — `POST /api/shape`, `POST /api/run/start`, `POST /api/watch`,
> `POST /api/burst` — répondent `501` avec un message pointant vers
> `docs/deploy.md`. L'audit et la consultation restent fonctionnels.

## Santé & état

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/health` | GET | — | `{"ok": true, "version": "1.2.3", "mode": "full\|observe\|\"\"}` | — |
| `/api/doctor` | GET | — | `{"mode", "checks": [{name, status: ok\|warn\|fail, detail}]}` — mêmes capacités que `cgo doctor` (os, tc, cap_net_admin, bbr, ping) | — |
| `/api/state` | GET | — | snapshot courant (`GetSnap`) ; `{"running": false}` si non câblé | — |
| `/api/diagnostics` | GET | — | `{"hub":"ok","time":"<RFC3339 UTC>"}` | — |
| `/api/stream` | GET (SSE) | — | flux 10 Hz : replay (`Last-Event-ID`, anneau 2048) puis live ; frame complète pour client frais **ou** client périmé (ID antérieur à l'anneau) ; event `backpressure` | `503` > 64 abonnés ; `500` flusher non supporté |
| `/api/schema` | GET | — | `{"params":[...]}` — registre unique des bornes/défauts (le client rend les champs depuis lui) | — |
| `/api/cost/tiers` | GET | — | `{"tiers":[...],"default":"<palier actif>"}` — paliers tarifaires réels | — |
| `/api/cost/tier` | POST | `{tier: nom exact}` | `{"ok":true,"tier"}` — le palier actif suit (tous les coûts affichés/calculés ; en-mémoire, redémarrage = défaut) | `400` JSON invalide ou palier inconnu |
| `/api/cost/enterprise` | GET | — | `{"plans":[{name, monthly_opex_mga, cir_mbps?\|fup_bytes?}], "tariffs":[{name, ar_per_gb, quoted?}], "emergency_ar_per_gb":5556, "primary_mapping":{fiber,other,4g,5g,vsat}, "disclaimer"}` — **tout est sérialisé depuis `pkg/metrics`** : un tarif changé au paquet met à jour l'API et les coûts calculés côté serveur pour l'UI, aucun nombre recopié ; un plan CIR n'a pas de FUP et inversement (champ absent, jamais un zéro qui lit « mesuré ») ; `quoted:true` = devis opérateur (tarif APN), à présenter comme tel | — |
| `/api/burst` | POST | `{cc: cubic\|bbr, seconds: 2–10 (défaut 4)}` | `{"ok":true,"cc","seconds"}` — sonde bulk à travers le bord façonné, bloquant | `400` JSON invalide, cc/seconds hors bornes ; `409` campagne active ; `501` mode observation ; `503` moteur non câblé ; `500` échec sonde |

**Champs d'interface (additifs).** Quand `CGO_IFACE` est défini côté serveur,
chaque frame `/api/state` et `/api/stream` porte en plus :
`iface_name`, `iface_source` (`twin` | `prod` | `unknown`), `iface_rx_mbps`,
`iface_tx_mbps`, `iface_rx_drops`, `iface_tx_drops`, `iface_window_ms`,
`iface_sample_ts`, `iface_valid`, `iface_gap`. `iface_valid:false` invalide
tous les champs numériques ; `iface_gap` vaut `warmup`, `absent`, `reset`,
`slow`, `read` ou `clock`. Aucun champ `iface_*` n'est présent quand
`CGO_IFACE` est vide. Le hub 10 Hz lit un échantillon frais ; `/api/state`
renvoie le dernier échantillon sans avancer la fenêtre.

## Profils

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/profiles` | GET | — | `{"profiles":[{id, capacity_mbps, capacity_up_mbps?, delay_ms, jitter_ms, loss_pct, loss_burst_{p,r,h,k}?, imported}]}` trié par id ; `imported` = id non natif (natifs : P1–P4) ; champs `?` omis quand à zéro | — |
| `/api/profile/list` | GET | — | map `model.Profiles` complète (rechargée du disque) | — |
| `/api/profile/import` | POST | CSV (`text/csv` ou corps ne commençant pas par `{`) : `id,capacity_mbps,delay_ms,jitter_ms,loss_pct[,capacity_up_mbps,loss_burst_p,loss_burst_r,loss_burst_h,loss_burst_k]` (en-tête optionnel sauté ; positions 5–9 absentes = symétrique + perte uniforme) **ou** JSON `model.Profile` | `{"ok": true, "profile": {...}}` | `400` bad body / bad csv / `id` manquant ; `500` échec d'import |

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
| `/api/run/start` | POST | `{profiles, qdiscs?, ccs?, reps, deadline_ms, target, direction?, flows?, bench?, channel_model?}` — `direction` : `up` (défaut) upload client→sink · `down` download source→client (shaper veth-s via netns) · `both` RRUL séquentiel (matrice up puis matrice down dans la même campagne). `flows` : absent/0 ⇒ **4** (défaut du banc ; 1 = régime historique, 2–8 flux concurrents) — **changement de comportement** : l'ancien défaut API était 1. `bench` : `""` ⇒ `CGO_BENCH` → `legacy` (reproduction du corpus) \| `dualns` (banc durci). `channel_model` : `""`/`clean` \| `bursty` (rafales Gilbert-Elliott, **P2/P3 uniquement**) | `{"started": true}` + événement journal | `400` JSON invalide, direction hors bornes, qdisc/cc inconnu, axes déséquilibrés, `flows` hors 1–8, `bench` ou `channel_model` inconnu, `bursty` avec un profil hors {P2,P3} (les profils fautifs sont nommés) ; `503` moteur non câblé — **la validation précède le test `StartFn == nil`** (un corps invalide répond 400 même sur un hôte sans moteur ; le mode observation, lui, répond 501 avant lecture) ; `409` toute erreur de `StartFn` |
| `/api/run/stop` | POST | — | `{"stopped": true}` + événement journal | `503` moteur non câblé |
| `/api/run/skip` | POST | — | `{"skipped": true}` — coupe la cellule en cours sans arrêter la matrice (reprise possible) | `503` moteur non câblé |
| `/api/results?run=` | GET | — | `{"available": true, "groups": [...]}` (médianes sur lignes non `invalid`, IQR small/RTT, `direction: up\|down`, quarantaine, best) | `{"available": false, "reason": …}` si aucun gel |
| `/api/results/delta?cell=P\|q\|cc[\|dir]` | GET | — | `{"available":true,"cell":"P\|q\|cc\|dir","previous_run","current_run","delta":{small_p95_pct,rtt_p95_pct,goodput_pct}}` — dérive entre les deux derniers runs ; `dir` = `down` explicite, défaut `up` (les sens ne se comparent jamais) | `400` cellule mal formée ; `{"available":false}` si < 2 runs ou cellule absente |
| `/api/run/rows?run=` | GET | — | `{"run", "rows": [lignes brutes aqm_eval.csv]}` | `400` id de run vide ou contenant `/`, `\`, `.` (anti-traversal) ; `404` run introuvable ou vide |
| `/api/events` | GET | — | `{"events": [{ts, kind, msg}, …]}` — anneau des 50 derniers | — |

`cc=bbr3` n'est accepté que si `/proc/sys/net/ipv4/tcp_available_congestion_control`
expose exactement le token `bbr3` ; sinon `400` avec le chemin
`kit/install-bbr3.sh`, jamais un repli silencieux. `bbr` reste BBRv1.

## Intégrité, rapport, figures

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/integrity` | GET | — | `{available, runs, manifests, valid, degraded, quarantined, run_ids (récents d'abord), breakdown: [{run, rows, valid, degraded, quarantined}], updated: <RFC3339 mtime dernier csv>, sha256, hash8}` — `valid` est strict (les lignes `degraded` sont comptées à part) | `{"available": false}` si aucun run |
| `/api/quarantine?run=` | GET | — | `{"quarantines": [{run, event_id, profile, qdisc, cc, gate_status}]}` — `run` vide = tous runs (récents d'abord), `[]` si aucune | `400` id invalide (anti-traversal) ; `404` run inconnu |
| `/api/report/export?format=md\|csv\|sh` | GET | — | tableau Markdown (défaut), CSV en pièce jointe `report.csv` (en-tête + lignes ; la ligne de provenance `hash8` n'existe que dans le Markdown, pas dans le CSV), ou **script tc** en pièce jointe `aqm-recipe.sh` : par profil, la meilleure cellule gelée devient une recette rejouable (reset + netem aux conditions du profil + shaper au goodput mesuré ×0,9 — même règle que la suggestion CLI), sans `sudo` préfixé (l'opérateur lance en root) | `404` aucune donnée |
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
| `/api/audit/start` | POST | `{site, link_type, provider, duration?, target?, direction?, flows?, small_url?, bulk_addr?, iface?, routeros_iface?, routeros_lan?, qdisc?, asymmetric?, cpu_constrained?}` — `duration` défaut 30 s (borné 10–600) ; `direction` défaut `up` (`down\|both`) ; `flows` ≤ 0 ⇒ 4 (dont absent), > 8 ⇒ 400 ; `target` vide ⇒ puits du banc détecté (`external` : `8.8.8.8`) ; `small_url` vide ⇒ env `CGO_SMALL_URL` puis puits — **reste vide en external** (jamais de charge implicite, un A+ sans charge serait un mensonge) ; `bulk_addr` vide ⇒ `EnvOrBulk` ; l'événement d'ouverture nomme le banc détecté | `{"started": true}` — exécution **asynchrone** (goroutine détachée), résultat appendu à `data/link_audit.csv` | `400` JSON invalide, `target` trop longue, direction inconnue, `flows` > 8 ; `409` audit déjà en cours |
| `/api/audit/status` | GET | — | `{"running": bool, "last": <Result\|null>, "report": <Report\|null>}` — `report` est le dernier rapport en mémoire ; **effacé au démarrage d'un nouvel audit** (un run annulé ou en échec ne peut pas afficher l'ancien rapport comme son résultat) | — |
| `/api/audit/live` | GET | — | `{"running": true, "phase", "elapsed_sec", "total_sec", "phase_remain_sec", "last_rtt_ms", "last_small_ms", "samples":{"rtt_ms":[],"small_ms":[]}}` — `phase` ∈ `repos\|charge\|charge-down\|drain` (`charge-down` = seconde fenêtre d'un run `both`) ; **non running ⇒ `{"running":false}` seul** : aucun zéro ne se lit comme mesuré | — |
| `/api/audit/stop` | POST | — | `{"stopped": true}` si un audit tournait, `{"stopped": false}` sinon — annule le contexte du run ; la goroutine journalise « annulé par l'opérateur » (distinguable d'un échec) et **n'append aucune ligne CSV partielle** | — |
| `/api/audit/report?audit_id=&format=json\|md` | GET | — | `<Report>` JSON (défaut) ou Markdown (`format=md`) — sans `audit_id`, le dernier rapport en mémoire, sinon `data/audit_report.json` ; avec `audit_id`, la copie `data/audits/<id>/` ; `unavailable[]` nomme chaque métrique non mesurable et sa raison (ex. « économie entreprise — fenêtre de charge vide »). En `both`, quand la fenêtre headline est muette, les miroirs mesurés de l'autre sens restent rendus ; seule l'économie non calculable devient une entrée `Unavailable`, jamais une absence silencieuse ni un zéro fabriqué. `invocation` gèle en plus les réglages **effectifs** de l'audit (durée, flux, sens, cible, petit objet, puits, mode de banc détecté, origine `cli`/`http`) et la commande CLI équivalente, après normalisation par le moteur : c'est la borne « donnée brute + commande exacte » de la console. Le bloc est **absent** (et non zéro) pour un rapport scellé avant ce champ, et le Markdown rend alors `Reproduction: non enregistrée` | `404` rapport introuvable |
| `/api/remediate?audit_id=&iface=&routeros_iface=&routeros_lan=&qdisc=&asymmetric=&cpu_constrained=` | GET | — | `{audit_id, remediation, unavailable}` — reprend le rapport en mémoire (ou `data/audits/<id>/`) ; `iface` fourni reconstruit la remédiation depuis le résultat mesuré, sinon renvoie celle du rapport archivé ; `asymmetric` n'est faux que sur `=false` | `404` aucun rapport d'audit |
| `/api/audit/presets` | GET | — | `{presets: [{id, label, site, link_type, provider, duration, target, notes}]}` — gabarits du protocole RQ1 (fibre siège, 4G heure creuse, 4G heure pointe) ; aucun preset n'est une mesure exécutée | — |
| `/api/audit/list` | GET | — | `{"audits": [lignes de data/link_audit.csv + `has_report`]}` ; quand le rapport archivé existe et scelle un verdict, `verdict_level` et `verdict_word` s'ajoutent à la ligne — jamais dérivés du CSV, dont la note ne dit pas si la charge a été tenté : une ligne sans artefact, ou un archive antérieur au champ, ne reçoit **aucun** verdict plutôt qu'un « neutre » inventé | — (retourne `[]` si absent) |
| `/api/audit/toprofile` | POST | — | `{"ok":true,"profile"}` — dernière ligne de `link_audit.csv` → profil rejouable (`capacity` = throughput, `delay` = rtt_idle_p95, replis 20/100 documentés) | `404` aucun audit / audit vide ; `500` échec d'import |

## Traduction matérielle

| Endpoint | Méthode | Corps | Réponse | Erreurs |
|---|---|---|---|---|
| `/api/hardware/translate?profile=P2` | GET | — | `{"recommendation": "…"}` — qdisc recommandé : `Best` mesuré si disponible, sinon `cake` pour P1, `fq_codel` sinon | `{"available": false, "reason": "aucune donnée — lancez campagne"}` (HTTP 200) |

## Divers

- **404 JSON** pour tout `/api/*` inconnu : `{"error": "not found", "path"}` —
  enregistré avant le catch-all SPA.
- **SPA** : toute autre route sert `web/frontend.FS` (fichiers embarqués).
