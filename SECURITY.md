# Politique de sécurité — Meteolink (CGO)

*Dernière révision : 2026-11-14. Correspond au binaire 1.3.5.*

## Versions supportées

| Version | Support |
|---|---|
| 1.3.x (dernier tag : v1.3.5) | ✅ correctifs de sécurité |
| ≤ 1.3.4 | ❌ plus de correctifs ; mettez à jour |

Une seule ligne de support maintient les choses honnêtes : les correctifs
visent `main` et le tag courant. Il n'y a **pas** de rétroportage. Les
releases publiées par la voie CI signent les checksums et joignent une
attestation de provenance.

## Modèle de menace (honnête)

Meteolink est un banc d'essai réseau embarqué sur le poste/serveur de
l'opérateur. Son seul rôle est de mesurer un lien que l'opérateur possède et
où il a un droit de regard.

- **Défaut = loopback.** Écoute `127.0.0.1:9090` + TLS (certificat local
  auto-signé `meteolink.dev` → `127.0.0.1`). En mode dev sans jeton, le
  loopback seul suffit ; aucun accès distant n'est ouvert.
- **Hors loopback = refusé sans jeton.** Un bind non-loopback (`--addr`
  ailleurs que 127.0.0.0/8, `::1`, `localhost`) sans `--expose` → exit 2,
  aucun socket ouvert. Avec `--expose` mais sans jeton (`CGO_API_TOKEN`
  absent **et** `/etc/cgo/auth.token` absent) → exit 2. TLS seul
  n'authentifie personne.
- **Deux adaptateurs d'authentification** (sur `/api/*`, sauf `/api/health`
  qui reste anonyme pour la vivacité) : Bearer (`Authorization: Bearer
  <token>`) pour CLI/API, ou cookie de session HttpOnly `cgo_session`
  (HMAC-SHA256 du jeton) pour navigateur/EventSource via `POST /api/login` ;
  `POST /api/logout` l'efface. Le jeton n'est jamais affiché ni journalisé.
- **En-têtes de sécurité** (nosniff, DENY, Referrer-Policy, CSP self-only,
  HSTS sous TLS réel), **timeouts de lecture** et **limite de débit par IP**
  (429 + `Retry-After: 1`) sur les POST.
- **Mode observation** (Windows ou `--mode observe`) : les mutations renvoient
  **501** — rien n'est façonné par accident.
- **Jeton de service** : `cgo service install` écrit l'unité systemd
  **loopback+TLS**, génère `/etc/cgo/auth.token` (32 octets, `0600
  root:root`) si absent et le passe via `EnvironmentFile=-` — jamais en ligne
  dans l'unité (world-readable).
- **Cookie déterministe** : la session est un HMAC du jeton ; il n'y a pas de
  rotation côté serveur sans rotation du jeton, et pas de révocation
  individuelle. La rotation du jeton invalide toutes les sessions.
- **Sessions d'audit à purge** : `cgo audit purge` (voir `docs/config.md` et
  `--help`) ; jamais de suppression du corpus gelé ni des sessions dans la
  fenêtre `--keep`.
- **Cibles privées** : par défaut, les cibles RFC1918/link-local sont
  refusées (SSRF, 403) ; le banc dualns opt-in `CGO_ALLOW_PRIVATE_TARGETS=1`
  les autorise explicitement. Ce n'est pas un interrupteur d'administration.
- **Ce qui reste hors modèle** : un attaquant déjà sur la machine (accès
  lire `/etc/cgo/auth.token`), un navigateur légitime avec le cookie
  (session = possession du jeton), et le banc lui-même (`tc` exige
  `CAP_NET_ADMIN`/root — l'hôte du banc est de confiance).

### État honnête des capacités (caps de l'audit, non dissimulées)

| Cap | Constat | État |
|---|---|---|
| 35(ii) | clair/plaintext `0.0.0.0:9090` sans authentification | **code remédié** (Task 1.1 : loopback+TLS par défaut, `--expose`+jeton, double adaptateur) ; **vérification indépendante en attente** — la levée appartient au vérificateur |
| 40 | PAT en historique public | **OUVERT / humain** — rotation/révocation et purge d'historique sont des portes humaines (voir ci-dessous « Secrets dans le dépôt ») |
| 45 | la même exposition côté transport | **réduite** par le cadenasage (loopback+TLS, en-têtes, 429) ; **vérification indépendante en attente** |

Aucune de ces lignes ne dit « résolu » : la décision de lever appartient au
vérificateur et aux portes humaines.

## Secrets dans le dépôt

- Aucun secret vivant n'est suivi par git. La voie de release CI et les
  portes de commit lancent **gitleaks** et la porte **hygiene**.
- L'allowlist n'est **pas extensible** ailleurs : 18 archives
  `backup/*.tar.gz` historiques + `results/runs-frozen.tar.gz` (embarquée,
  `//go:embed`) + un faux littéral documenté (`CGO_BENCH=dualns`) — rien
  d'autre.
- `/etc/cgo/auth.token` vit **hors du dépôt**, `0600`, root:root. Le jeton ne
  transite jamais dans une unité systemd, un log ou un rapport scellé.
- **Cap 40 (PAT en historique public) reste OUVERT et relève de l'humain** :
  rotation/révocation côté fournisseur, purge `filter-repo`, scan
  inter-historique propre, force-push approuvé. La CI ajoute un gate de scan
  sur les forks/clones, mais aucune purge d'historique n'est automatique.
- Si vous trouvez un secret dans le dépôt : signalez-le via l'issue tracker
  avec le label `security` **sans le recopier** dans le texte public — le
  mainteneur reprend depuis là.

## Signaler une vulnérabilité

1. **Issue tracker du repo** : <https://github.com/Realms4239/cgo/issues> —
   ouvrez une issue avec le label `security`, **sans détails exploitables
   publics** (parlez du principe, de la version et du fichier concerné).
2. **Contact opérateur/DGM** : aucun contact DGM n'est documenté dans ce
   dépôt. Le canal opérateur est le contrat d'étude de cas (lecture seule,
   pseudonyme par défaut) — voir
   `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md` ; le
   signalement d'un incident suit ce canal, pas un email inventé.
3. Il n'existe pas de bug bounty. Il n'existe pas de contact « presse ». S'il
   vous faut un canal privé, demandez-le dans l'issue — la réponse se fait
   sur le tracker public, jamais en clair avec des détails d'exploitation.

Réponse attendue : accusé de réception sur le tracker ; aucune garantie de
délai n'est promise au-delà de « on regarde, on corrige, on publie un tag ».

## Licence

Meteolink est publié sous **licence MIT** — voir [`LICENSE`](LICENSE).
