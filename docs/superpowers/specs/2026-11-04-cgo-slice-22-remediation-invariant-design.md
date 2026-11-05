# Slice 22 — Réparer le domaine avant la vitrine : l'invariant de remédiation

*Spec écrite le 2026-11-04 par le builder. Statut : **projet** (brainstorming de la slice 22, HEAD `1dffe46`). Aucun code n'existe encore. Aucune mesure n'est revendiquée.*

## 1. Pourquoi

NEW.md et NEW1.md (lus à `C:\IDEA\NEW.md` et `C:\IDEA\NEW1.md`) nomment la même faille sous deux enregistrements : le script RouterOS `kind=pcq` dans les captures (n'existe plus à HEAD), et surtout le rapport qui affiche `Alertes à l'heure 96,7 %` + `Protéger ce lien — réglage calibré sur ce qui a été mesuré (95 % du goulot, Envoi 4,8 Mbit/s)` **alors que** la ligne brute dit `Bufferbloat non mesuré — aucune charge bulk exploitable` et `Débit 0,0 Mbit/s`.

Le rapport scellé Departement X du 2026-10-26 (`data/link_audit.csv` : `throughput_mbps=0`, `bloat_grade` vide sur 2 lignes) peut encore être lu comme un fait. Un jury de Master 2 ne voit pas le garde interne `tryEmit` ; il voit le fait produit.

La slice 22 fait donc disparaître la contradiction au niveau domaine, pas à l'écran : **un rapport sans charge mesurée ne peut pas porter de script**, et ce fait est nommé dans le JSON, pas inféré par l'UI. Les captures PCQ sont réfutées dans le code et figées par un test de régression ; la présentation (cockpit 3 cartes, tuiles) est explicitement hors scope — c'est la slice 23.

Human decisions (brainstorming, 2026-11-04): invariant `RTTIdleP50>0 ∧ (Up>0∨Down>0)` (NEW1's `>0.5 ∧ Δp95≥0` rejeté), scripts par direction, `RemediationAvailable` explicite, suppression frontend, PCQ gelé par test. Aucun live, aucun CHR, aucun mobile.

## 2. Ce qui existe déjà (vérifié en source)

- `pkg/remediate/remediate.go` : `RouterOS()` et `RouterOSIngress()` émettent `kind=cake|sfq|fifo` uniquement. Zéro occurrence de `pcq` dans `pkg/` (Go + tests). `RouterOSIngress` est un `<lan-queue-tree>` (`parent=<lan-ou-bridge>`), pas l'IFB ; Linux garde l'IFB (`ifb0 ingress`).
- `pkg/audit/report.go:403-490` : `BuildReport` émet `Remediation` derrière `res.RTTIdleP50 > 0 && (upCap>0 || downCap>0)`, avec scripts par direction et `Unavailable` par direction pour `both` partiel.
- `pkg/audit/report.go:208-241` : `Remediation` porte déjà `LinuxTC`, `RouterOS`, `LinuxTCIngress`, `RouterOSIngress`, `UpMeasuredMbps/UpShapedMbps`, `DownMeasuredMbps/DownShapedMbps`, `LinuxIface`, `RouterOSWanIface`, `RouterOSLanIface`, `RouterOSLanAssumed`.
- `pkg/audit/report_ingress_test.go` : `TestReportDownEmitsOnlyIngress`, `TestReportLinuxOnlyDownEmitsLinuxIngressAndNoRouterOS`, `TestReportBothMissingDownCapacityIsNamed`, `TestReportDownWithoutLinuxInterfaceStillNamesRouterOS`, etc. — la parité des directions est testée.
- `pkg/audit/routeros_compare_test.go` : 12/12 sur fixture, extract `kind=cake` — la seule validation RouterOS disponible avant un CHR.

Ce qui manque, et que NEW/NEW1 révèlent :

- Aucun champ `RemediationAvailable` → l'UI devine `nil`.
- Aucun test ne fige les six champs structurés dans le contrat JSON.
- Aucune suppression frontend : `RemediationCard.tsx` rend `Protéger ce lien` dès que `Remediation` est non-nil.
- Aucun test ne fige `kind ∈ {cake,sfq,fifo}` et l'absence de `pcq` / `/interface queue add`.

## 3. Design

### 3.1 Le domaine : `RemediationState` + `RemediationAvailable`

`Report` gagne un état nommé, plus son alias booléen :

```go
type RemediationState string

const (
    RemediationAvailableState  RemediationState = "available"
    RemediationObservationState RemediationState = "observation_only"
)

type Report struct {
    // …
    Remediation              *Remediation      `json:"remediation,omitempty"`
    RemediationState         RemediationState   `json:"remediation_state"`
    RemediationAvailable     bool               `json:"remediation_available"`
    // …
}
```

`BuildReport` :

- Le seul chemin qui pose `rep.Remediation = rem` pose aussi `rep.RemediationState = RemediationAvailableState` et `rep.RemediationAvailable = true`.
- Le chemin `remediation` absent pose `rep.RemediationState = RemediationObservationState`, `rep.RemediationAvailable = false` + l'`Unavailable` existant (`remediation` / `remediation (upload)` / `remediation (download)`).
- Invariant : `RemediationState == available ⇔ Remediation != nil ⇔ RemediationAvailable == true` ; sinon l'état est `observation_only`.
- Un rapport scellé avant ce champ décode avec `RemediationState == observation_only` et `RemediationAvailable == false` via `UnmarshalJSON` — jamais `true` par zéro Go.

L'invariant reste exactement celui du code actuel (`RTTIdleP50 > 0 ∧ (Up > 0 ∨ Down > 0)`). NEW1's `>0.5 Mbps ∧ Δp95≥0` est rejeté : un lien de champ à 0,3 Mbit/s est remédiable, et `Δp95≥0` est redondant avec `LoadMeasured` qui exige déjà un grade & un débit.

### 3.2 Contrat JSON figé

Tests nouveaux dans `pkg/audit/remediation_fields_test.go` :

- `TestRemediationAvailableMirrorsPresence` — unmeasured → `Available == false ∧ Remediation == nil`, measured → `Available == true ∧ Remediation != nil`.
- `TestRemediationStructuredFieldsRoundTrip` — `RenderJSON()` puis unmarshal : les six champs (`up/down` mesurés & façonnés, `linux_iface`, `routeros_wan_iface`, `routeros_lan_iface`, `routeros_lan_assumed`) survivent inchangés.

### 3.3 Frontend : suppression — l'absence comme état nommé

`RemediationCard.tsx` passe de `RemediationCardProps { remediation: Pcc null | undefined }` à `RemediationCardProps { remediation: Pcc | null | undefined; state: AuditRemediationState }` où :

```ts
export type AuditRemediationState = "available" | "observation_only"
```

- `state !== "available"` → `<EmptyState>Aucune configuration émise — charge non établie. Relancez un audit complet.</EmptyState>` ; jamais de carte, jamais de script.
- Une fixture `unmeasured` prouve la suppression en DOM (test `RemediationCard.dom.test.tsx`).
- `AuditView.tsx` et `DashboardView.tsx` passent `report.remediation_state` (mirror de `types.ts` via `AuditReport.remediation_state`) à la carte.




### 3.4 Garde PCQ

Nouveau `TestRouterOSEmitsOnlyValidKinds` dans `pkg/remediate/remediate_test.go` :

- Pour chaque `(qdisc ∈ {cake,fq_codel,pfifo_fast})` et chaque direction :
  - `RouterOS()` contient `kind=cake|sfq|fifo`, jamais `pcq`, jamais `/interface queue add`.
  - `RouterOSIngress()` : parent = LAN, pas WAN ; `RouterOS()` : parent = WAN, pas LAN.
  - Les deux ont leur commentaire d'inventaire disjoint (`cgo-managed` ≠ `cgo-managed-ingress`).

### 3.5 Non-objectifs (explicitement hors slice)

- Pas de cockpit 3 cartes, pas de tuiles profile, pas de `PREUVE VÉRIFIABLE` rewording, pas de deadline 220 par défaut, pas de `BBRv3` badge, pas de fix de troncature CSS : slice 23 (présentation).
- Pas de CHR boot, pas de live `audit`, pas de bench Linux 6.8, pas de mobile 390.
- Pas de CAP p95 promise : la slice ne change aucun nombre produit.
- Aucun champ `RemediationAvailable` côté RouterOS Validation — le CHR continue de valider le script, pas son absence.

## 4. Règles

- R1 : suppression visible above the fold dès qu'"unmeasured → no script" est testé.
- R2 : ce spec est le contrat ; tout changement de l'invariant rouvre le brainstorming.
- R3 : chaque tâche laisse une capture (JuryRoom, RemediationCard suppression).
- R4 : la spec nomme ce qu'elle ne prouve pas (§3.4).
- Le numérique gelé (`ShapedRate = 95 %`) ne change pas ; il est documenté par `remediate.ShapedRate`.
- Aucun live audit/bench/mesure bench sans approbation humaine explicite.

## 5. Test (vert avant code)

Le bon état fini est définissons par les tests existants mis à jour et les deux nouveaux :

```
go test ./pkg/audit -run TestReport -count=1
go test ./pkg/remediate -run TestRouterOS -count=1
cd web/frontend && npm test
python tools/verify.py
python book/tools/audit-archive.py
```

Acceptation : toutes les commandes de la section Vert existent et passent, les deux nouveaux tests existent, et une capture JuryRoom 1024/1920 montre la suppression sur une fixture non mesurée.

## 6. Séquence de commits

1. `fix(audit): fige le contrat RemediationAvailable et les champs structurés (slice 22, T0-T1)`
2. `ui(audit): casse la "Protéger ce lien" quand la charge n'est pas établie (slice 22, T2)`
3. `test(remediate): fige l'absence PCQ et /interface queue add (slice 22, T3)`
4. `docs: le paquet de vérification tranche 22, sans auto-certification (slice 22, T4)`

## Ledger après ce plan

| Task | Surfaces touchées | Capture | Lignes VERIFY qui bougent |
|---|---|---|---|
| T0 | node, spec | — | 1 |
| T1 | Go — invariant | — | 0 |
| T2 | UI — suppression | JuryRoom | 0 |
| T3 | Go — PCQ gate | — | 0 |
| T4 | packet | final set | 1 |

## 7. Non prouvé (pour la slice)

- Aucun audit live, aucun CHR, aucune bench Linux 6.8, aucun mobile : la slice ne le mesure pas et ne le revendique pas.
- Le `remediation_available: false` sur les rapports scellés avant la slice est une absence, pas une preuve qu'ils n'auraient pas rien émis.
