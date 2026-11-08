# Runbook — observateur eBPF (`cgo-bpf-observer`) — DESIGN DIFFÉRÉ

> **État : DEFERRED.** Design seul — aucun code, aucun chargeur, aucun
> objet eBPF construit ici. L'attache live, l'overhead et le dualns
> 4/4 restent différés jusqu'à la gate de construction Linux 6.8
> (humain) + l'objet construit + le chargeur écrit (§8.12 du design).

Contexte : `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md`
§8 (eBPF/BTF/CO-RE/cilium-ebpf, R1–R11, slices 0–6). La ligne BLOCKED
« eBPF retransmit / RTT / qdisc drops » (`docs/VERIFY.md:29`) est le
goulot : tant que l'objet et le chargeur n'existent pas, chaque
revendication `source=ebpf` est jouette. Le côté Windows est vérifié
(contrat + ladder + rapport, 2026-10-30) mais ne dit jamais
« vérifié » là où la condition est une fonctionnalité noyau Linux.

## API `Session` cible (dépôt indépendant `cgo-bpf-observer`)

Un dépôt, une bibliothèque, CLI fin en sucre. Zéro import
`metrics`/`audit`/`remediate`/`model` — `go list -m all` montre
stdlib + `cilium/ebpf` seuls. `internal/` (pas `pkg/`) : les
appels externes consomment UN package public (`probe`) + les valeurs
`Summary`/`Flow`/`Globals` retournées.

```go
package probe  // stdlib + cilium/ebpf seuls
// — palier TCP_INFO —
func TCPRetransBytes(c net.Conn) (uint64, error)
var ErrTCPInfoUnavailable error
// — sélection CC —
func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error)
func SetConnCC(conn net.Conn, cc string) error
// — palier noyau —
func KernelSource() KernelSourceInfo
type KernelSourceInfo struct{ Source, Reason string }  // Source ∈ ebpf|tcpinfo|""
// — Session (nouveau, design spec §4.1) —
type Session interface {
    Lookup(local, peer netip.AddrPort) (bpf.Flow, bool)
    Snapshot() []bpf.Flow
    Globals() (bpf.Globals, error)
    Attach() string  // "fentry"|"kprobe"
    Close() error
}
func Open() (*Session, error)   // cartes fraîches par session ; load+attach, tout-ou-rien par load
func Check() error              // vérifie SHA puis load-only (sans attach) ; BuildError|UnavailableError
// — bulk avec vérité noyau —
type BulkStats struct {
    Bytes []uint64 `json:"bytes"`
    RetransBytes uint64 `json:"retrans_bytes"`
    RetransMeasured bool `json:"retrans_measured"`
    Source string `json:"source,omitempty"`          // + ebpf|tcpinfo|""
    Kernel *bpf.Summary `json:"kernel,omitempty"`    // + couverture + par-flow
}
```

```text
cgo-bpf-observer/
  go.mod                          # module github.com/Realms4239/cgo-bpf-observer ; go 1.25 ; require cilium/ebpf <épinglé>, x/sys
  README.md                       # ce que ça mesure, ce que ça refuse de revendiquer, préconditions BLOCKED
  Makefile                        # bpf / bpf-check / test / verify (même forme que Makefile:44-59)
  cmd/cgo-bpf-observer/main.go   # CLI : attach + dump Summary JSON ; sous-commande `tcpinfo`
  internal/probe/                 # pkg/probe/*.go déplacés
  internal/probe/bpf/             # pkg/probe/bpf/{types.go,status.go,errors.go,assets_*.go,bpf_probe.c,BUILD.md}
  internal/probe/bpf/session_linux.go   # chargeur : cilium/ebpf Open/Load/Attach/Tear-down
  internal/probe/bpf/session_stub.go    # //go:build !linux || !amd64 — API identique, UnavailableError typé
```

CLI : `cgo-bpf-observer capture --flows 4 --duration 10s --cc bbr --out
summary.json` ; `cgo-bpf-observer tcpinfo --conn tcp://host:port` ;
`cgo-bpf-observer doctor` (`Check()` → ok|warn|fail, nomme le fait
manquant). Non-objectifs explicites : aucune implémentation BBR (la CC
est un nom d'option socket), aucun contrôle qdisc/tc, aucun netns,
ni économie/audit/rapport, pas de clés par-flow IPv6, pas de
queue-dwell/ECN/ACK pairing — portée 6.8 héritée.

## BTF / CO-RE

BTF = API noyau + format fichier ELF (`.BTF` + `.BTF.ext` avec
sous-section `core_relo`). CO-RE = Compile Once – Run Everywhere.
La règle concrète du design :

- (a) commiter `bpf_probe.o` préconstruit + SHA-256 épinglé +
  `BUILD.md` avec version clang, en-têtes libbpf, source BTF,
  commande exacte ; (b) `vmlinux.h` généré + gitignoré ; (c) compiler
  avec `-g` et relocations CO-RE pour que les offsets
  `BPF_CORE_READ` / `bpf_core_enum_value` soient patelés au
  chargement, pas codés en dur (le C le fait déjà — `bpf_probe.c:105-108`,
  `:195-197`) ; (d) construction reproductible — même toolchain +
  même BTF + même répertoire = même SHA épinglé, sinon enquête
  (DWARF incorpore les chemins source : un autre répertoire change
  le hash, `BUILD.md:45-48`) ; (e) vérifier les sections
  `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`,
  `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb` via `readelf`
  (`BUILD.md:35-41`) et épingler le SHA via
  `TestEmbeddedObjectPinned` sans `make`.
- Génération : `bpftool btf dump file /sys/kernel/btf/vmlinux format c
  > vmlinux.h` ; inclure `vmlinux.h` élimine la dépendance aux
  en-têtes noyau système. libbpf résout les relocations BTF
  enregistrées contre la BTF runtime et patche les offsets au
  chargement.
- À ne PAS vendre : BTF est métadonnée/debug, pas un offset magique.
  Elle ne fait pas charger un programme vérifieur-unsafe, ni ne rend
  `tcpinfo` byte-exact. Un champ CO-RE manquant fait échouer tout le
  chargement de la collection et retombe sur TCP_INFO — comportement
  honnête, pas silence.

Hooks eBPF (6.8) : `raw_tp/tcp_retransmit_skb` (octets retransmis
`skb->len` payload, TSO compris) ; `fentry/tcp_rcv_established` +
fallback `kprobe/tcp_rcv_established` (échantillons `srtt_us >> 3`,
` snd_cwnd`, état, ≤20 Hz/socket) ; `raw_tp/kfree_skb` filtré sur
`SKB_DROP_REASON_QDISC_DROP` via `bpf_core_enum_value` — Linux 6.8 n'a
PAS de tracepoint `qdisc_drop` (introduit en v7.1) ; sur 7.1+ le
`qdisc_drop` natif peut remplacer `kfree_skb` (§8.12 : différé).

## Discipline capget (capacités réelles, pas d'heuristique euid)

Le chargeur DOIT vérifier les capacités (capget) et retomber avec
une raison typée, pas approximer via l'euid ni `stat` seul. Le
`kernelFacts` actuel (`kernelsource.go:19-31`) approche les caps par
`os.Geteuid() == 0` et BTF par `os.Stat("/sys/kernel/btf/vmlinux")` :
root contourne SAUF si le bounding set a retiré la capacité, donc
`euid == 0` sur-revendique ; la PRÉSENCE de BTF ≠ BTF parsable
(`CONFIG_DEBUG_INFO_BTF`). Correction design : parser BTF (via
`cilium/ebpf/btf` ou lecture `bpftool`, pas `stat`) et utiliser
`capget` pour tester `CAP_BPF` / `CAP_PERFMON` dans l'ensemble
effectif ; le détail `warn` du doctor doit nommer le fait manquant.

- `CAP_BPF` (depuis 5.8) = BPF privilégié ; `CAP_PERFMON` (depuis
  5.8) = perf/BPF perf ; `CAP_NET_ADMIN` déjà requis pour `tc` ; root
  supporté, repli `sudo setcap cap_bpf,cap_perfmon,cap_net_admin+ep
  ./cgo`. `CAP_SYS_ADMIN` est surchargé/legacy — éviter pour les
  nouvelles fonctionnalités.
- `Check()` = vérifier le hash puis load-only (pas d'attache) ; le
  docteur ne doit jamais revendiquer la capacité d'attache depuis un
  load-only : « loadable (BTF + CAP_BPF) ; l'attache est vérifiée par
  l'audit ».
- Chaque fenêtre non-`ebpf` porte `Source=tcpinfo|""` +
  `FallbackReason` nommant le fait manquant (non-Linux, BTF absent,
  pas `CAP_BPF`, vérifieur/CO-RE, attache, couverture partielle,
  carte pleine, expéditeur distant) — ladder `status.go:11` + table
  spec §5. Manquant ≠ zéro ; zéro exact exige un échantillon.

## Gate de construction 6.8

Conditions avant tout slice live (humain) :

- Hôte Linux 6.8 stable joignable ; Ubuntu 24.04, noyau 6.8.0,
  clang-18 (`BPF_CLANG=clang-18`), en-têtes libbpf-dev
  (`<bpf/*.h>`), bpftool, `/sys/kernel/btf/vmlinux` (`CONFIG_DEBUG_INFO_BTF`).
  `.BTF`/`.BTF.ext` exigent clang ≥8.0 ou pahole.
- `make bpf` génère `vmlinux.h` → compilation CO-RE → `bpf_probe.o.sha256` ;
  `make bpf-check` sort 0 ; vérification `readelf -S … | grep -E
  'raw_tp|fentry|kprobe'` ; SHA intégré == SHA épinglé
  (`TestEmbeddedObjectPinned`).
- Objet C GPL-2.0 (`bpf_probe.c:1`) — conserver les deux licences
  (MIT repo + GPL objet) dans les fichiers ; ne jamais relicencier
  l'objet.
- Windows ne peut pas construire : « Windows body only » — les lignes
  VERIFY eBPF doivent dire « Windows body only », jamais « vérifié »
  là où la condition est une fonctionnalité noyau Linux.

## Dépendance nouvelle `cilium/ebpf` — stop-and-ask

`go.mod` ne contient AUCUN chargeur (`github.com/cilium/ebpf`, ou
binding libbpf) — le code de chargeur et de lecture de cartes
n'existe pas aujourd'hui (spec §3.5 dit que `go.mod` épinglera une
release précise pour `session_linux.go` ; le pin est absent).
Ajouter `cilium/ebpf` est une dépendance nouvelle = **stop-and-ask
selon AGENTS.md** (pas de dépendance nouvelle sans approbation
humaine explicite). La slice 2 (objet + chargeur) ne démarre qu'après
cette approbation écrite + la gate 6.8 ci-dessus.

Phase chargeur à implémenter (Open/Load/Attach/Tear-down) :
parser ELF → `LoadCollectionSpec` → `NewCollection` (charge cartes +
progs, `Collection.Close()` libère) ou `LoadAndAssign` (chargement
sélectif dans une struct) ; `rlimit.RemoveMemlock()` best-effort sur
6.8 ; cartes par session (`Collection` par session, pas d'état
global) avec `Lookup`/`Snapshot` concurrents-safe ; split par build
tag : chargeur réel `//go:build linux && amd64`, stub ailleurs avec
API exportée identique ; capture du log vérifieur. Aucun slice de ce
design ne touche du code dans cette passe — slices 0–6 (§8.10) sont
tous conditionnés aux gates ci-dessus.

## Références primaires

- `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md` §8
  (hooks/hooks 6.8 vs 7.1, BTF/CO-RE, chargeur libbpf/ebpf-go,
  capget, slices 0–6, critères anti-jouet, §8.12 DEFERRED).
- `docs/superpowers/plans/2026-11-08-cgo-supplychain-gaps.md` Task 5.
