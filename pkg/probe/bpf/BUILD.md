# bpf_probe.o — artefact eBPF embarqué

Objet CO-RE précompilé, commité. La VM et `go build` ne demandent ni clang,
ni gcc, ni llvm, ni linux-headers.

## Statut

**Source et cible de build posées le 2026-10-29. L'objet n'est pas encore
construit : aucun hôte Linux 6.8 stable n'était joignable.** Ne pas committer
un objet sans le `sha256sum` produit sur la cible 6.8 ; ne pas fabriquer
l'objet ni le hash. Le reste du plan eBPF (chargeur Go, tests d'attache,
rapport) attend cet objet.

## Outillage épinglé

- Ubuntu 24.04 LTS, noyau 6.8.0 (BTF `/sys/kernel/btf/vmlinux`)
- clang 18 (`clang-18`)
- libbpf-dev (en-têtes `<bpf/*.h>`)
- bpftool (génération de `vmlinux.h`)

## Construction

```sh
sudo apt-get install -y clang-18 libbpf-dev bpftool
make bpf
make bpf-check
```

`make bpf` génère `vmlinux.h` depuis le BTF du noyau courant, compile
`bpf_probe.c` en CO-RE, puis écrit `bpf_probe.o.sha256`. `vmlinux.h` est
gitignoré : seul l'objet est le livrable.

## Vérification

```sh
readelf -S pkg/probe/bpf/bpf_probe.o | grep -E 'raw_tp|fentry|kprobe'
```

Les sections `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`,
`kprobe/tcp_rcv_established` et `raw_tp/kfree_skb` doivent être présentes.
Le futur test Go `TestEmbeddedObjectPinned` vérifiera le SHA-256 sans make.

## Reproductibilité

Un rebuild sur le même outillage, le même BTF et le même chemin de travail
doit reproduire le SHA-256 épinglé. Le DWARF embarque le chemin des sources :
un autre répertoire change le hash. Dans ce cas, régénérer le `.sha256` et le
SHA-256 de ce fichier, jamais accepter un objet dont le hash ne correspond pas.
