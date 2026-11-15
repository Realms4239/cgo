# bpf_probe.o — artefact eBPF embarqué

Objet CO-RE précompilé et embarqué : la VM et `go build` ne demandent ni clang,
ni gcc, ni llvm, ni linux-headers.

## Statut

**Le 2026-11-12/13, l'objet a été construit sur le rehearsal VM 6.8.0-149**
(`make bpf` → `sha256(b6de79ac365ae6697f784ace1971cd75f0b89cd067e6ebf8b2c5b45b70d2f7e3)`),
**le chargeur Go est écrit (`loader_linux.go`, cilium/ebpf v0.19.0)**, et une
**audit eBPF réelle a été scellée** : `audit-1794517790` (`kernel_telemetry
source=ebpf, object_sha256 b6de79ac…`, 10/10 flux échantillonnés, 2.69 Mo
réémis). Le 2026-11-12/13, un **audit eBPF live a été scellé** avec la
télémétrie noyau dans le rapport (preuves :
`docs/verification/2026-11-12-ebpf-probe-bringup/README.md`).

Piège de build : sur Linux, compiler le CLI requiert `CGO_ENABLED=0`
(sinon le `.c` de la source de l'objet est rejeté par le toolchain Go —
comportement voulu, la .c n'est pas du Go). La Windows, le défaut est déjà
`CGO_ENABLED=0`.

Ne pas committer l'objet avant que l'acceptation live 6.8 soit vérifiée
**et** que l'humain décide. bpf_probe.o reste hors du dépôt pour l'instant ;
son SHA-256 consigné est le seul qui fasse foi en attendant.

## Outillage épinglé

- Ubuntu 24.04 LTS, noyau 6.8.0 (BTF `/sys/kernel/btf/vmlinux`)
- clang 18 (`clang-18`)
- libbpf-dev (en-têtes `<bpf/*.h>`)
- bpftool (génération de `vmlinux.h`)

## Construction

```sh
sudo apt-get install -y clang-18 libbpf-dev linux-tools-generic
make bpf
make bpf-check
```

`make bpf` génère `vmlinux.h` depuis le BTF du noyau courant, compile
`probe/bpf/src/bpf_probe.c` en CO-RE, puis écrit `bpf_probe.o.sha256`.
`vmlinux.h` est gitignoré : seul l'objet est le livrable.

## Vérification

```sh
readelf -S probe/bpf/bpf_probe.o | grep -E 'raw_tp|fentry|kprobe'
```

Les sections `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`,
`kprobe/tcp_rcv_established` et `raw_tp/kfree_skb` doivent être présentes.
Le futur test Go `TestEmbeddedObjectPinned` vérifiera le SHA-256 sans make.

## Reproductibilité

Un rebuild sur le même outillage, le même BTF et le même chemin de travail
doit reproduire le SHA-256 épinglé. Le DWARF embarque le chemin des sources :
un autre répertoire change le hash. Dans ce cas, régénérer le `.sha256` et le
SHA-256 de ce fichier, jamais accepter un objet dont le hash ne correspond pas.
