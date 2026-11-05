# CGO eBPF Ground-Truth Telemetry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the TCP_INFO retransmission proxy with kernel-exact sender-side retransmit payload counters on `linux/amd64`, keep TCP_INFO as an explicit fallback, and carry the source through the audit, report, and doctor — without touching the frozen corpus or CSV schemas.

**Architecture:** A prebuilt CO-RE object (`pkg/probe/bpf/bpf_probe.o`) is embedded in the Go binary. A raw tracepoint on `tcp_retransmit_skb` counts submitted retransmit payload per IPv4 flow; a fentry/kprobe on `tcp_rcv_established` samples socket state; a raw tracepoint on `kfree_skb` counts qdisc drops. `pkg/probe` starts one session per bulk load, queries the sender-side flow before each socket closes, and falls back to TCP_INFO for the whole window if any flow is missing or unobserved. Sources are never mixed inside one window.

**Tech Stack:** Go 1.25 module (`github.com/Realms4239/cgo`), `github.com/cilium/ebpf v0.22.0`, clang 18 + libbpf headers + bpftool on the maintainer build host, stdlib `testing`. The runtime host needs no clang, gcc, llvm, or linux-headers.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-27-cgo-ebpf-ground-truth-design.md` (approved 2026-10-27). Every task is bound by it.
- Branch: `rebuild/plan1-economics-audit`. Do not merge to `main`.
- Reference bench: Ubuntu 24.04 LTS, Linux 6.8, amd64. The committed object is built against 6.8 BTF; 7.x portability is a later gate.
- Runtime artifact: `pkg/probe/bpf/bpf_probe.o` is committed with `bpf_probe.o.sha256` and `BUILD.md`. `vmlinux.h` is generated and gitignored. No runtime clang.
- `payload = skb->len` at the `tcp_retransmit_skb` tracepoint. No header subtraction, no wire counter, no `PayloadKnown`.
- Success-only accounting: failed retransmit attempts do not fire the tracepoint. The report says “réémissions transmises”, never “ACKed” or “delivered”.
- Qdisc drops use `skb:kfree_skb` filtered to `SKB_DROP_REASON_QDISC_DROP`. Drop counters are host-wide diagnostics; they never enter the retransmission ratio.
- All-or-nothing per window: a missing or unobserved flow makes the whole window use TCP_INFO. Sources are never summed. A missing flow is never zero.
- `basis: measured|estimated` keeps its meaning. `retrans_ratio` gains an optional `source: ebpf|tcpinfo` (`""` omitted). Additive JSON only.
- `telemetry.csv`, `link_audit.csv`, their headers, `pkg/audit/telemetry.go`, `pkg/audit/writer.go`, and the frozen corpus stay untouched.
- French comments and operator-facing text; English identifiers.
- Never `git add -A`. Stage only the files named by the task. Leave `epbfBBR.md`, `data/runs/`, and user artifacts untracked.
- The frozen guard's authoritative count is **159 runs / 208 valid lines**; never restate it as “208 runs”.
- Do not call the feature “verified” until the opt-in Linux integration test and the manual `dualns` acceptance run on a stable 6.8 host.
- Known blocker: the CGO-DEV VM is unstable (soft lockups, unresponsive SSH). Task 1 Step 4 and the live acceptance are blocked until a stable Linux 6.8 host builds and runs the object. The Windows gates cannot run until the object is committed.

---

## File Structure

**Create**
- `pkg/probe/bpf/bpf_probe.c` — CO-RE source, four programs, two maps.
- `pkg/probe/bpf/bpf_probe.o` — prebuilt CO-RE ELF (Linux host, committed).
- `pkg/probe/bpf/bpf_probe.o.sha256` — pinned lowercase hex.
- `pkg/probe/bpf/BUILD.md` — toolchain, exact commands, hash.
- `pkg/probe/bpf/types.go` — `Flow`, `Summary`, `Globals`, platform-neutral.
- `pkg/probe/bpf/errors.go` — `UnavailableError`, `BuildError`.
- `pkg/probe/bpf/assets.go` — embed, `ObjectSHA256`, `checkObject`.
- `pkg/probe/bpf/wire.go` — C-compatible key/stat structs and host-order helpers.
- `pkg/probe/bpf/session.go` — `Session` wrapper, `Start`, `Check`.
- `pkg/probe/bpf/session_linux.go` — cilium/ebpf loader and attach lifecycle.
- `pkg/probe/bpf/session_stub.go` — typed unavailable stub for other targets.
- `pkg/probe/bpf/wire_test.go` — key encoding and struct sizes.
- `pkg/probe/bpf/assets_test.go` — embedded hash and `BuildError`.
- `pkg/probe/bpf/session_stub_test.go` — `Check` unavailable off `linux/amd64`.
- `pkg/probe/bpf/integration_linux_test.go` — opt-in `ebpf_integration` loopback test.
- `pkg/probe/kernel.go` — `KernelTelemetry = bpf.Summary` alias.
- `pkg/doctor/ebpf_test.go` — ok/warn/fail mapping and `Report` presence.

**Modify**
- `pkg/probe/bulkstats_test.go` — append the complete/partial/zero/map-full/
  start-failure ladder; keep the existing tests.
- `Makefile` — `bpf`, `bpf-check` maintainer targets.
- `.gitignore` — `pkg/probe/bpf/vmlinux.h`.
- `go.mod`, `go.sum` — `github.com/cilium/ebpf v0.22.0`.
- `pkg/probe/bulkstats.go` — session lifecycle, sender keys, source, summary.
- `pkg/metrics/cost.go` — `RetransRatio.Source`.
- `pkg/audit/audit.go` — `LoadOutcome`, `Result` fields, ratio sources, pointers.
- `pkg/audit/report.go` — `kernel_telemetry`, Markdown section.
- `pkg/doctor/doctor.go` — `ebpf` check.
- `pkg/audit/audit_test.go`, `pkg/audit/report_test.go`, `pkg/audit/writer_test.go`, `pkg/metrics/cost_test.go` — new assertions.
- `docs/api.md` — doctor capability list.

**Never touch**
- `pkg/audit/telemetry.go`, `pkg/audit/writer.go`, `data/runs/`, `book/data/runs`, `pkg/results/runs-frozen.tar.gz`, `web/frontend/`.

---

### Task 1: CO-RE object, loader, and typed fallback

**Files:**
- Create: `pkg/probe/bpf/bpf_probe.c`, `pkg/probe/bpf/BUILD.md`
- Generate: `pkg/probe/bpf/bpf_probe.o`, `pkg/probe/bpf/bpf_probe.o.sha256`
- Create: `pkg/probe/bpf/types.go`, `errors.go`, `assets.go`, `wire.go`, `session.go`, `session_linux.go`, `session_stub.go`, `wire_test.go`, `assets_test.go`, `session_stub_test.go`
- Modify: `Makefile`, `.gitignore`, `go.mod`, `go.sum`

**Interfaces:**
- Consumes: the approved spec §3 (object, maps, loader) and §4.1 (API).
- Produces: `bpf.Start() (*Session, error)`, `(*Session).Lookup/Snapshot/Globals/Attach/Close`, `bpf.Check() error`, `bpf.ObjectSHA256() string`, `bpf.UnavailableError`, `bpf.BuildError`, `bpf.Flow`, `bpf.Summary`, `bpf.Globals`.

- [ ] **Step 1: Write the CO-RE source**

Create `pkg/probe/bpf/bpf_probe.c`:

```c
// SPDX-License-Identifier: GPL-2.0
//
// bpf_probe.c — télémétrie noyau exacte du banc CGO.
// Objet CO-RE précompilé et embarqué : la VM ne compile jamais de BPF.
//
// Réémissions : raw_tp/tcp_retransmit_skb, succès seulement, skb->len est
//   déjà le payload TCP du skb de file (pas le clone qui porte les en-têtes).
// Échantillon socket : fentry/tcp_rcv_established, repli kprobe.
// Drops qdisc : raw_tp/kfree_skb filtré sur SKB_DROP_REASON_QDISC_DROP.

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#ifndef AF_INET
#define AF_INET 2
#endif
#ifndef IPPROTO_TCP
#define IPPROTO_TCP 6
#endif

#define MAX_FLOWS 4096
#define SAMPLE_NS 50000000ULL /* 50 ms */

struct flow_key {
    __u32 saddr;   /* local IPv4, ordre hôte */
    __u32 daddr;   /* pair IPv4, ordre hôte */
    __u16 sport;   /* ordre hôte */
    __u16 dport;   /* ordre hôte */
    __u16 family;  /* AF_INET seulement dans cette tranche */
    __u16 pad;
};

struct flow_stat {
    __u64 retrans_payload;
    __u64 retrans_events;
    __u64 drop_packets;
    __u64 drop_bytes;
    __u32 srtt_us;
    __u32 snd_cwnd;
    __u32 snd_ssthresh;
    __u32 state;
    __u64 last_sample_ns;
    __u32 samples;
    __u32 pad;     /* taille stable : 64 octets, comme bpfFlowStat en Go */
};

struct global_stat {
    __u64 retrans_payload; /* IPv6/non-IPv4 : global seulement */
    __u64 retrans_events;
    __u64 drop_packets;
    __u64 drop_bytes;
    __u64 map_full;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_FLOWS);
    __type(key, struct flow_key);
    __type(value, struct flow_stat);
} flows SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct global_stat);
} globals SEC(".maps");

static __always_inline struct global_stat *globals_get(void)
{
    __u32 zero = 0;
    return bpf_map_lookup_elem(&globals, &zero);
}

static __always_inline struct flow_stat *
flow_get_or_create(const struct flow_key *key)
{
    struct flow_stat *st = bpf_map_lookup_elem(&flows, key);
    if (st)
        return st;

    struct flow_stat zero = {};
    if (bpf_map_update_elem(&flows, key, &zero, BPF_NOEXIST) == 0)
        return bpf_map_lookup_elem(&flows, key);

    /* Course avec un autre CPU : l'entrée peut exister malgré l'échec. */
    st = bpf_map_lookup_elem(&flows, key);
    if (!st) {
        struct global_stat *g = globals_get();
        if (g)
            __sync_fetch_and_add(&g->map_full, 1);
    }
    return st;
}

static __always_inline void
socket_key(struct sock *sk, struct flow_key *key)
{
    key->family = AF_INET;
    key->saddr = bpf_ntohl(BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr));
    key->daddr = bpf_ntohl(BPF_CORE_READ(sk, __sk_common.skc_daddr));
    key->sport = BPF_CORE_READ(sk, __sk_common.skc_num);
    key->dport = bpf_ntohs(BPF_CORE_READ(sk, __sk_common.skc_dport));
}

SEC("raw_tp/tcp_retransmit_skb")
int on_retransmit(struct bpf_raw_tracepoint_args *ctx)
{
    struct sock *sk = (struct sock *)ctx->args[0];
    struct sk_buff *skb = (struct sk_buff *)ctx->args[1];
    if (!sk || !skb)
        return 0;

    __u64 payload = BPF_CORE_READ(skb, len);

    if (BPF_CORE_READ(sk, __sk_common.skc_family) != AF_INET) {
        struct global_stat *g = globals_get();
        if (g) {
            __sync_fetch_and_add(&g->retrans_payload, payload);
            __sync_fetch_and_add(&g->retrans_events, 1);
        }
        return 0;
    }

    struct flow_key key = {};
    socket_key(sk, &key);
    struct flow_stat *st = flow_get_or_create(&key);
    if (!st)
        return 0;

    __sync_fetch_and_add(&st->retrans_payload, payload);
    __sync_fetch_and_add(&st->retrans_events, 1);
    return 0;
}

static __always_inline int sample_sock(struct sock *sk)
{
    if (!sk || BPF_CORE_READ(sk, __sk_common.skc_family) != AF_INET)
        return 0;

    struct flow_key key = {};
    socket_key(sk, &key);
    struct flow_stat *st = flow_get_or_create(&key);
    if (!st)
        return 0;

    __u64 now = bpf_ktime_get_ns();
    __u64 last = st->last_sample_ns;
    if (last && now - last < SAMPLE_NS)
        return 0;
    st->last_sample_ns = now;

    struct tcp_sock *tp = (struct tcp_sock *)sk;
    st->srtt_us = BPF_CORE_READ(tp, srtt_us) >> 3;
    st->snd_cwnd = BPF_CORE_READ(tp, snd_cwnd);
    st->snd_ssthresh = BPF_CORE_READ(tp, snd_ssthresh);
    st->state = BPF_CORE_READ(sk, __sk_common.skc_state);
    __sync_fetch_and_add(&st->samples, 1);
    return 0;
}

SEC("fentry/tcp_rcv_established")
int BPF_PROG(on_ack_fentry, struct sock *sk, struct sk_buff *skb)
{
    return sample_sock(sk);
}

SEC("kprobe/tcp_rcv_established")
int BPF_KPROBE(on_ack_kprobe, struct sock *sk, struct sk_buff *skb)
{
    return sample_sock(sk);
}

SEC("raw_tp/kfree_skb")
int on_kfree_skb(struct bpf_raw_tracepoint_args *ctx)
{
    struct sk_buff *skb = (struct sk_buff *)ctx->args[0];
    if (!skb)
        return 0;

    __u32 reason = (__u32)ctx->args[2];
    if (reason != (__u32)bpf_core_enum_value(enum skb_drop_reason,
                                            SKB_DROP_REASON_QDISC_DROP))
        return 0;

    __u32 total_len = BPF_CORE_READ(skb, len);
    __u32 data_len = BPF_CORE_READ(skb, data_len);
    if (total_len < data_len)
        return 0;
    __u32 headlen = total_len - data_len;

    struct global_stat *g = globals_get();
    if (g) {
        __sync_fetch_and_add(&g->drop_packets, 1);
        __sync_fetch_and_add(&g->drop_bytes, total_len);
    }

    __u16 nh = BPF_CORE_READ(skb, network_header);
    if ((__u32)nh + sizeof(struct iphdr) + sizeof(struct tcphdr) > headlen)
        return 0;

    unsigned char *head = (unsigned char *)BPF_CORE_READ(skb, head);
    if (!head)
        return 0;

    struct iphdr iph;
    if (bpf_probe_read_kernel(&iph, sizeof(iph), head + nh) != 0)
        return 0;
    if (iph.version != 4 || iph.protocol != IPPROTO_TCP)
        return 0;

    __u32 ihl = (__u32)iph.ihl * 4;
    if (ihl < sizeof(struct iphdr))
        return 0;
    if ((__u32)nh + ihl + sizeof(struct tcphdr) > headlen)
        return 0;
    if ((bpf_ntohs(iph.frag_off) & 0x3fff) != 0)
        return 0;

    struct tcphdr th;
    if (bpf_probe_read_kernel(&th, sizeof(th), head + nh + ihl) != 0)
        return 0;

    struct flow_key key = {};
    key.family = AF_INET;
    key.saddr = bpf_ntohl(iph.saddr);
    key.daddr = bpf_ntohl(iph.daddr);
    key.sport = bpf_ntohs(th.source);
    key.dport = bpf_ntohs(th.dest);

    /* Attribution par flux seulement si l'entrée existe déjà ; sinon global. */
    struct flow_stat *st = bpf_map_lookup_elem(&flows, &key);
    if (st) {
        __sync_fetch_and_add(&st->drop_packets, 1);
        __sync_fetch_and_add(&st->drop_bytes, total_len);
    }
    return 0;
}

char __license[] SEC("license") = "GPL";
```

- [ ] **Step 2: Write the build documentation**

Create `pkg/probe/bpf/BUILD.md`:

```markdown
# bpf_probe.o — artefact eBPF embarqué

Objet CO-RE précompilé, commité. La VM et `go build` ne demandent ni clang,
ni gcc, ni llvm, ni linux-headers.

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
Le test Go `TestEmbeddedObjectPinned` vérifie le SHA-256 sans make.

## Reproductibilité

Un rebuild sur le même outillage, le même BTF et le même chemin de travail
doit reproduire le SHA-256 épinglé. Le DWARF embarque le chemin des sources :
un autre répertoire change le hash. Dans ce cas, régénérer le `.sha256` et le
SHA-256 de ce fichier, jamais accepter un objet dont le hash ne correspond pas.
```

- [ ] **Step 3: Add the maintainer targets and ignore the generated header**

Replace `Makefile` with (the existing targets are unchanged):

```make
GO ?= go
FRONTEND := web/frontend
BPF_DIR := pkg/probe/bpf
BPF_CLANG ?= clang-18
BPF_CFLAGS := -O2 -g -target bpf -D__TARGET_ARCH_x86 -Wall -Werror -I$(BPF_DIR)

.PHONY: build test vet test-real figures frontend clean bpf bpf-check

build: frontend
	$(GO) build -o bin/cgo$(shell go env GOEXE) ./cmd/cgo

test:
	$(GO) vet ./...
	$(GO) test ./...
	cd $(FRONTEND) && bun run typecheck && bun run build && bunx vitest run && npx playwright test

vet:
	$(GO) vet ./...

# Real gates (tc/netem/BBR) — run INSIDE the VM only.
test-real:
	$(GO) test -tags=real ./pkg/qdisc/... ./pkg/campagne/...

figures: build
	./bin/cgo figures

frontend:
	cd $(FRONTEND) && bun run build

clean:
	rm -rf bin $(FRONTEND)/dist

$(BPF_DIR)/vmlinux.h:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $@

bpf: $(BPF_DIR)/vmlinux.h
	$(BPF_CLANG) $(BPF_CFLAGS) -c $(BPF_DIR)/bpf_probe.c -o $(BPF_DIR)/bpf_probe.o
	sha256sum $(BPF_DIR)/bpf_probe.o | cut -d' ' -f1 > $(BPF_DIR)/bpf_probe.o.sha256

bpf-check:
	@cd $(BPF_DIR) && test "$$(sha256sum bpf_probe.o | cut -d' ' -f1)" = "$$(cat bpf_probe.o.sha256)"
```

Add to `.gitignore` under `# deps/build`:

```gitignore
pkg/probe/bpf/vmlinux.h
```

- [ ] **Step 4: Build the object on Linux 6.8 and pin its hash**

Run on the Linux 6.8 host, from the repo root:

```sh
sudo apt-get update
sudo apt-get install -y clang-18 libbpf-dev bpftool
make bpf
make bpf-check
readelf -S pkg/probe/bpf/bpf_probe.o | grep -E 'raw_tp|fentry|kprobe'
sha256sum pkg/probe/bpf/bpf_probe.o
```

Expected: `make bpf-check` exits 0; the `readelf` output names the four
sections above; `sha256sum` prints one lowercase hex digest.

Append the digest to `BUILD.md` (replace nothing else):

```markdown
## SHA-256

`<digest from sha256sum>`
```

Do not fabricate the object, the digest, or a successful load. If the VM is
unstable, stop here; the remaining steps wait for a stable host.

- [ ] **Step 5: Add the Go dependency**

Run:

```sh
go get github.com/cilium/ebpf@v0.22.0
go mod tidy
```

Expected: `go.mod` contains `github.com/cilium/ebpf v0.22.0`; `go.sum` gains
its checksums; the `go` directive stays `1.25` (the dependency requires
1.25.0 and the local toolchain is 1.26.5). `go mod tidy` also raises
`golang.org/x/sys` from v0.36.0 to at least v0.43.0 (cilium's own minimum):
that is an expected transitive bump, not a code change. Re-run
`go test ./...` after it.

- [ ] **Step 6: Write the failing wire and object tests**

Create `pkg/probe/bpf/wire_test.go`:

```go
package bpf

import (
	"encoding/binary"
	"net/netip"
	"testing"
	"unsafe"
)

// La clé noyau est en ordre hôte des deux côtés. Un mélange d'ordres
// manquerait silencieusement tous les flux et fabriquerait un faux zéro.
func TestFlowKeyHostOrder(t *testing.T) {
	local := netip.MustParseAddrPort("192.168.1.10:40000")
	peer := netip.MustParseAddrPort("10.0.0.2:5201")
	k, ok := flowKeyFrom(local, peer)
	if !ok {
		t.Fatal("IPv4 refusé")
	}
	if k.Saddr != 0xC0A8010A {
		t.Fatalf("saddr = %#x, want 0xC0A8010A", k.Saddr)
	}
	if k.Daddr != 0x0A000002 {
		t.Fatalf("daddr = %#x, want 0x0A000002", k.Daddr)
	}
	if k.Sport != 40000 || k.Dport != 5201 || k.Family != afINET {
		t.Fatalf("ports/famille = %d/%d/%d", k.Sport, k.Dport, k.Family)
	}
}

func TestFlowKeyRejectsIPv6(t *testing.T) {
	local := netip.MustParseAddrPort("[::1]:40000")
	peer := netip.MustParseAddrPort("[::1]:5201")
	if _, ok := flowKeyFrom(local, peer); ok {
		t.Fatal("IPv6 ne doit pas produire de clé par flux")
	}
}

// net.TCPAddr.AddrPort() rend les IPv4 en forme mappée : la clé doit les
// démaper, sinon aucun flux réel n'est jamais trouvé.
func TestFlowKeyUnmapsFourInSix(t *testing.T) {
	local := netip.AddrPortFrom(netip.MustParseAddr("::ffff:192.168.1.10"), 40000)
	peer := netip.AddrPortFrom(netip.MustParseAddr("::ffff:10.0.0.2"), 5201)
	k, ok := flowKeyFrom(local, peer)
	if !ok {
		t.Fatal("IPv4 mappée refusée")
	}
	if k.Saddr != 0xC0A8010A || k.Daddr != 0x0A000002 {
		t.Fatalf("saddr/daddr = %#x/%#x", k.Saddr, k.Daddr)
	}
}

func TestFlowFromRestoresHostOrder(t *testing.T) {
	f := flowFrom(flowKey{
		Saddr: 0xC0A8010A, Daddr: 0x0A000002,
		Sport: 40000, Dport: 5201, Family: afINET,
	}, bpfFlowStat{RetransPayload: 7, Samples: 2})
	if f.Local != "192.168.1.10:40000" || f.Peer != "10.0.0.2:5201" {
		t.Fatalf("flux = %s -> %s", f.Local, f.Peer)
	}
	if f.RetransPayload != 7 || f.Samples != 2 {
		t.Fatalf("compteurs = %+v", f)
	}
}

// Les tailles doivent correspondre à struct flow_key / struct flow_stat :
// cilium/ebpf encode par reflection et refuse un décalage de padding.
func TestWireSizes(t *testing.T) {
	if got := binary.Size(flowKey{}); got != 16 {
		t.Fatalf("flowKey = %d octets, want 16", got)
	}
	if got := binary.Size(bpfFlowStat{}); got != 64 {
		t.Fatalf("bpfFlowStat = %d octets, want 64", got)
	}
	if got := int(unsafe.Sizeof(bpfFlowStat{})); got != 64 {
		t.Fatalf("unsafe.Sizeof(bpfFlowStat) = %d, want 64", got)
	}
}
```

Create `pkg/probe/bpf/assets_test.go`:

```go
package bpf

import (
	"errors"
	"strings"
	"testing"
)

func TestEmbeddedObjectPinned(t *testing.T) {
	if len(objectBytes) == 0 {
		t.Fatal("objet eBPF vide")
	}
	if err := checkObject(); err != nil {
		t.Fatalf("objet épinglé : %v", err)
	}
	if got := ObjectSHA256(); !strings.EqualFold(got, strings.TrimSpace(pinnedHash)) {
		t.Fatalf("ObjectSHA256 = %s, épinglé = %s", got, strings.TrimSpace(pinnedHash))
	}
}

func TestObjectHashMismatchIsBuildError(t *testing.T) {
	err := checkObjectBytes(objectBytes, strings.Repeat("0", 64))
	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("BuildError attendu, reçu %v", err)
	}
}

func TestEmptyObjectIsBuildError(t *testing.T) {
	var be *BuildError
	if !errors.As(checkObjectBytes(nil, "x"), &be) {
		t.Fatal("objet vide : BuildError attendu")
	}
}
```

Create `pkg/probe/bpf/session_stub_test.go`:

```go
//go:build !linux || !amd64

package bpf

import (
	"errors"
	"testing"
)

func TestCheckUnavailableOffLinuxAMD64(t *testing.T) {
	err := Check()
	var ue *UnavailableError
	if !errors.As(err, &ue) {
		t.Fatalf("UnavailableError attendu, reçu %v", err)
	}
}
```

- [ ] **Step 7: Run the tests to verify they fail**

Run:

```sh
go test ./pkg/probe/bpf/ -count=1
```

Expected: FAIL to compile — `flowKeyFrom`, `flowFrom`, `flowKey`, `bpfFlowStat`,
`objectBytes`, `pinnedHash`, `BuildError`, `UnavailableError` are undefined.

- [ ] **Step 8: Implement the platform-neutral types and key helpers**

Create `pkg/probe/bpf/types.go`:

```go
// Package bpf — objet eBPF CO-RE embarqué : réémissions exactes par flux,
// échantillon socket et drops qdisc. Le chargeur réel vit dans
// session_linux.go ; partout ailleurs un stub rend une erreur typée.
package bpf

// Flow — un flux côté émetteur, lu dans la table noyau.
type Flow struct {
	Local          string `json:"local"`
	Peer           string `json:"peer"`
	RetransPayload uint64 `json:"retrans_payload_bytes"`
	RetransEvents  uint64 `json:"retrans_events"`
	DropPackets    uint64 `json:"qdisc_drop_packets"`
	DropBytes      uint64 `json:"qdisc_drop_bytes"`
	SRTTus         uint32 `json:"srtt_us"`
	SndCwnd        uint32 `json:"snd_cwnd"`
	SndSsthresh    uint32 `json:"snd_ssthresh"`
	State          uint32 `json:"state"`
	Samples        uint32 `json:"samples"`
}

// Summary — une fenêtre de télémétrie. Source dit quel ratio en est tiré :
// ebpf | tcpinfo | "" ; Available dit si la session a été attachée.
type Summary struct {
	Available      bool   `json:"available"`
	Source         string `json:"source,omitempty"`
	Attach         string `json:"attach,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ObjectSHA256   string `json:"object_sha256,omitempty"`
	FlowsSeen      int    `json:"flows_seen"`
	FlowsTotal     int    `json:"flows_total"`
	RetransPayload uint64 `json:"retrans_payload_bytes"`
	RetransEvents  uint64 `json:"retrans_events"`
	QdiscDrops     uint64 `json:"qdisc_drop_packets"`
	QdiscDropBytes uint64 `json:"qdisc_drop_bytes"`
	QdiscAvailable bool   `json:"qdisc_available"`
	MapFull        uint64 `json:"map_full,omitempty"`
	Flows          []Flow `json:"flows,omitempty"`
}

// Globals — diagnostics hôte, jamais une affirmation par flux.
type Globals struct {
	DropPackets    uint64
	DropBytes      uint64
	MapFull        uint64
	QdiscAvailable bool
}
```

Create `pkg/probe/bpf/errors.go`:

```go
package bpf

// UnavailableError — limitation de l'hôte (BTF, capacités, OS, vérificateur).
type UnavailableError struct{ Reason string }

func (e *UnavailableError) Error() string { return "ebpf indisponible : " + e.Reason }

// BuildError — défaut de l'objet embarqué (hash, ELF, programme absent).
type BuildError struct{ Reason string }

func (e *BuildError) Error() string { return "ebpf : objet invalide : " + e.Reason }
```

Create `pkg/probe/bpf/assets.go`:

```go
package bpf

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
)

//go:embed bpf_probe.o
var objectBytes []byte

//go:embed bpf_probe.o.sha256
var pinnedHash string

// ObjectSHA256 rend le SHA-256 de l'objet embarqué (hex minuscule).
func ObjectSHA256() string {
	sum := sha256.Sum256(objectBytes)
	return hex.EncodeToString(sum[:])
}

func checkObjectBytes(obj []byte, pinned string) error {
	if len(obj) == 0 {
		return &BuildError{Reason: "objet eBPF embarqué vide"}
	}
	sum := sha256.Sum256(obj)
	got := hex.EncodeToString(sum[:])
	want := strings.ToLower(strings.TrimSpace(pinned))
	if got != want {
		return &BuildError{Reason: fmt.Sprintf("sha256 de bpf_probe.o = %s, épinglé = %s", got, want)}
	}
	return nil
}

func checkObject() error { return checkObjectBytes(objectBytes, pinnedHash) }
```

Create `pkg/probe/bpf/wire.go`:

```go
package bpf

import (
	"encoding/binary"
	"net/netip"
)

const afINET = 2

// flowKey — disposition mémoire de struct flow_key (16 octets).
// Champs exportés : cilium/ebpf encode les structs par reflection.
type flowKey struct {
	Saddr  uint32
	Daddr  uint32
	Sport  uint16
	Dport  uint16
	Family uint16
	Pad    uint16
}

// bpfFlowStat — disposition mémoire de struct flow_stat (64 octets).
type bpfFlowStat struct {
	RetransPayload uint64
	RetransEvents  uint64
	DropPackets    uint64
	DropBytes      uint64
	SRTTus         uint32
	SndCwnd        uint32
	SndSsthresh    uint32
	State          uint32
	LastSampleNs   uint64
	Samples        uint32
	Pad            uint32
}

// flowKeyFrom — la paire locale/pair côté ÉMETTEUR, en ordre hôte.
// IPv6 n'a pas de clé par flux dans cette tranche.
func flowKeyFrom(local, peer netip.AddrPort) (flowKey, bool) {
	// Unmap : net.TCPAddr.AddrPort() rend une IPv4 mappée en 16 octets
	// (::ffff:127.0.0.1) dont Is4() est faux — sans Unmap, toute connexion
	// réelle manquerait la table et le ratio retomberait en TCP_INFO.
	l, p := local.Addr().Unmap(), peer.Addr().Unmap()
	if !l.Is4() || !p.Is4() {
		return flowKey{}, false
	}
	la, pa := l.As4(), p.As4()
	return flowKey{
		Saddr:  binary.BigEndian.Uint32(la[:]),
		Daddr:  binary.BigEndian.Uint32(pa[:]),
		Sport:  local.Port(),
		Dport:  peer.Port(),
		Family: afINET,
	}, true
}

func flowFrom(key flowKey, st bpfFlowStat) Flow {
	local := netip.AddrPortFrom(netip.AddrFrom4(u32To4(key.Saddr)), key.Sport)
	peer := netip.AddrPortFrom(netip.AddrFrom4(u32To4(key.Daddr)), key.Dport)
	return Flow{
		Local:          local.String(),
		Peer:           peer.String(),
		RetransPayload: st.RetransPayload,
		RetransEvents:  st.RetransEvents,
		DropPackets:    st.DropPackets,
		DropBytes:      st.DropBytes,
		SRTTus:         st.SRTTus,
		SndCwnd:        st.SndCwnd,
		SndSsthresh:    st.SndSsthresh,
		State:          st.State,
		Samples:        st.Samples,
	}
}

// u32To4 — nombre hôte (comme bpf_ntohl) vers les quatre octets de l'adresse.
func u32To4(v uint32) [4]byte {
	return [4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}
```

- [ ] **Step 9: Implement the session wrapper and the non-Linux stub**

Create `pkg/probe/bpf/session.go`:

```go
package bpf

import "net/netip"

// backend — l'implémentation réelle (session_linux.go) ou le stub.
type backend interface {
	Lookup(local, peer netip.AddrPort) (Flow, bool)
	Snapshot() []Flow
	Globals() (Globals, error)
	Attach() string
	Close() error
}

// Session — une collection chargée avec ses cartes et ses sondes propres.
// Chaque Start crée sa session : deux audits concurrents ne se réinitialisent
// jamais.
type Session struct{ b backend }

// Start vérifie l'objet épinglé, le charge et attache les sondes.
func Start() (*Session, error) {
	if err := checkObject(); err != nil {
		return nil, err
	}
	b, err := start()
	if err != nil {
		return nil, err
	}
	return &Session{b: b}, nil
}

func (s *Session) Lookup(local, peer netip.AddrPort) (Flow, bool) {
	if s == nil || s.b == nil {
		return Flow{}, false
	}
	return s.b.Lookup(local, peer)
}

func (s *Session) Snapshot() []Flow {
	if s == nil || s.b == nil {
		return nil
	}
	return s.b.Snapshot()
}

func (s *Session) Globals() (Globals, error) {
	if s == nil || s.b == nil {
		return Globals{}, nil
	}
	return s.b.Globals()
}

func (s *Session) Attach() string {
	if s == nil || s.b == nil {
		return ""
	}
	return s.b.Attach()
}

func (s *Session) Close() error {
	if s == nil || s.b == nil {
		return nil
	}
	return s.b.Close()
}

// Check vérifie le SHA épinglé puis charge l'objet sans l'attacher. Un défaut
// d'objet est un BuildError ; une limitation de l'hôte est un
// UnavailableError.
func Check() error {
	if err := checkObject(); err != nil {
		return err
	}
	return check()
}
```

Create `pkg/probe/bpf/session_stub.go`:

```go
//go:build !linux || !amd64

package bpf

func start() (backend, error) {
	return nil, &UnavailableError{Reason: "linux/amd64 requis"}
}

func check() error {
	return &UnavailableError{Reason: "linux/amd64 requis"}
}
```

- [ ] **Step 10: Run the neutral tests**

Run:

```sh
go test ./pkg/probe/bpf/ -count=1
```

Expected: PASS on Windows (the stub path). This requires the object from
Step 4 to exist.

- [ ] **Step 11: Implement the Linux loader**

Create `pkg/probe/bpf/session_linux.go`:

```go
//go:build linux && amd64

package bpf

import (
	"bytes"
	"errors"
	"net/netip"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

// bpfGlobal — disposition mémoire de struct global_stat (40 octets).
type bpfGlobal struct {
	RetransPayload uint64
	RetransEvents  uint64
	DropPackets    uint64
	DropBytes      uint64
	MapFull        uint64
}

type linuxSession struct {
	coll       *ebpf.Collection
	flows      *ebpf.Map
	globals    *ebpf.Map
	links      []link.Link
	attachMode string
	qdisc      bool
	cpus       int
}

func loadSpec() (*ebpf.CollectionSpec, error) {
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(objectBytes))
	if err != nil {
		return nil, &BuildError{Reason: "lecture ELF : " + err.Error()}
	}
	return spec, nil
}

// check charge sans attacher : le doctor ne prétend jamais que l'attache
// fonctionne.
func check() error {
	spec, err := loadSpec()
	if err != nil {
		return err
	}
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return &UnavailableError{Reason: "chargement refusé (BTF/CO-RE/vérificateur/capacités) : " + err.Error()}
	}
	return coll.Close()
}

func start() (backend, error) {
	_ = rlimit.RemoveMemlock() // best-effort : comptabilité memcg sur 6.8

	spec, err := loadSpec()
	if err != nil {
		return nil, err
	}
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return nil, &UnavailableError{Reason: "chargement refusé (BTF/CO-RE/vérificateur/capacités) : " + err.Error()}
	}

	s := &linuxSession{coll: coll, flows: coll.Maps["flows"], globals: coll.Maps["globals"]}
	fail := func(err error) (backend, error) {
		_ = s.Close()
		return nil, err
	}

	if s.flows == nil || s.globals == nil {
		return fail(&BuildError{Reason: "cartes flows/globals absentes de l'objet"})
	}
	for _, name := range []string{"on_retransmit", "on_ack_fentry", "on_ack_kprobe", "on_kfree_skb"} {
		if coll.Programs[name] == nil {
			return fail(&BuildError{Reason: "programme " + name + " absent de l'objet"})
		}
	}
	cpus, err := ebpf.PossibleCPU()
	if err != nil {
		return fail(&UnavailableError{Reason: "CPUs possibles : " + err.Error()})
	}
	s.cpus = cpus

	lnk, err := link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "tcp_retransmit_skb",
		Program: coll.Programs["on_retransmit"],
	})
	if err != nil {
		return fail(&UnavailableError{Reason: "attache tcp_retransmit_skb : " + err.Error()})
	}
	s.links = append(s.links, lnk)

	// fentry d'abord, kprobe en repli ; l'attache est nommée, jamais mélangée.
	lnk, err = link.AttachTracing(link.TracingOptions{
		Program:    coll.Programs["on_ack_fentry"],
		AttachType: ebpf.AttachTraceFEntry,
	})
	if err == nil {
		s.links = append(s.links, lnk)
		s.attachMode = "fentry"
	} else {
		lnk, err = link.Kprobe("tcp_rcv_established", coll.Programs["on_ack_kprobe"], nil)
		if err != nil {
			return fail(&UnavailableError{Reason: "attache tcp_rcv_established (fentry et kprobe) : " + err.Error()})
		}
		s.links = append(s.links, lnk)
		s.attachMode = "kprobe"
	}

	// Sonde indépendante : un échec laisse les réémissions actives.
	lnk, err = link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "kfree_skb",
		Program: coll.Programs["on_kfree_skb"],
	})
	if err == nil {
		s.links = append(s.links, lnk)
		s.qdisc = true
	}
	return s, nil
}

func (s *linuxSession) Lookup(local, peer netip.AddrPort) (Flow, bool) {
	key, ok := flowKeyFrom(local, peer)
	if !ok {
		return Flow{}, false
	}
	var st bpfFlowStat
	if err := s.flows.Lookup(&key, &st); err != nil {
		return Flow{}, false
	}
	return flowFrom(key, st), true
}

func (s *linuxSession) Snapshot() []Flow {
	it := s.flows.Iterate()
	var key flowKey
	var st bpfFlowStat
	out := make([]Flow, 0, 16)
	for it.Next(&key, &st) {
		out = append(out, flowFrom(key, st))
	}
	return out
}

func (s *linuxSession) Globals() (Globals, error) {
	per := make([]bpfGlobal, s.cpus)
	if err := s.globals.Lookup(uint32(0), &per); err != nil {
		return Globals{}, err
	}
	out := Globals{QdiscAvailable: s.qdisc}
	for _, g := range per {
		out.DropPackets += g.DropPackets
		out.DropBytes += g.DropBytes
		out.MapFull += g.MapFull
	}
	return out, nil
}

func (s *linuxSession) Attach() string { return s.attachMode }

func (s *linuxSession) Close() error {
	var errs []error
	for _, l := range s.links {
		if err := l.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	s.links = nil
	if s.coll != nil {
		if err := s.coll.Close(); err != nil {
			errs = append(errs, err)
		}
		s.coll = nil
	}
	return errors.Join(errs...)
}
```

- [ ] **Step 12: Cross-check the Linux build**

Run:

```sh
GOOS=linux GOARCH=amd64 go build ./...
GOOS=linux GOARCH=amd64 go vet ./pkg/probe/bpf/
```

Expected: both exit 0. If the cilium/ebpf API differs from this plan,
fix the call site to the installed version — do not change the object.

- [ ] **Step 13: Commit**

```sh
git add pkg/probe/bpf/bpf_probe.c pkg/probe/bpf/bpf_probe.o \
  pkg/probe/bpf/bpf_probe.o.sha256 pkg/probe/bpf/BUILD.md \
  pkg/probe/bpf/types.go pkg/probe/bpf/errors.go pkg/probe/bpf/assets.go \
  pkg/probe/bpf/wire.go pkg/probe/bpf/session.go \
  pkg/probe/bpf/session_linux.go pkg/probe/bpf/session_stub.go \
  pkg/probe/bpf/wire_test.go pkg/probe/bpf/assets_test.go \
  pkg/probe/bpf/session_stub_test.go Makefile .gitignore go.mod go.sum
git commit -m "bpf: objet CO-RE, chargeur et repli TCP_INFO"
```

---

### Task 2: Bulk telemetry, metrics source, and audit wiring

**Files:**
- Create: `pkg/probe/kernel.go`
- Modify: `pkg/probe/bulkstats_test.go` (append; keep the existing tests)
- Modify: `pkg/probe/bulkstats.go`
- Modify: `pkg/metrics/cost.go`, `pkg/metrics/cost_test.go`
- Modify: `pkg/audit/audit.go`, `pkg/audit/audit_test.go`, `pkg/audit/writer_test.go`

**Interfaces:**
- Consumes: Task 1's `bpf.Start`, `(*Session).Lookup/Snapshot/Globals/Attach/Close`, `bpf.Flow`, `bpf.Summary`, `bpf.Globals`.
- Produces: `probe.KernelTelemetry` (alias of `bpf.Summary`), `probe.BulkStats.Source` and `.Kernel`, `metrics.RetransRatio.Source`, `audit.LoadOutcome.RetransSource` and `.Kernel`, `audit.Result.KernelUp/Down` and `.RetransRatioUp/Down`.

- [ ] **Step 1: Write the failing bulk tests**

Extend the existing `pkg/probe/bulkstats_test.go` — keep
`TestBulkStatsMultiFlowLocal` and `TestTCPRetransBytesErrorOnNonTCP`. Add
`errors`, `net/netip`, `strings`, and
`github.com/Realms4239/cgo/pkg/probe/bpf` to its imports (`context`, `net`,
`testing`, `time` are already there), then append:

```go
type fakeKernel struct {
	flows   map[string]bpf.Flow
	globals bpf.Globals
	gerr    error
	attach  string
	closed  bool
}

func (f *fakeKernel) Lookup(local, peer netip.AddrPort) (bpf.Flow, bool) {
	fl, ok := f.flows[local.String()+" "+peer.String()]
	return fl, ok
}

func (f *fakeKernel) Snapshot() []bpf.Flow { return nil }
func (f *fakeKernel) Globals() (bpf.Globals, error) { return f.globals, f.gerr }
func (f *fakeKernel) Attach() string { return f.attach }
func (f *fakeKernel) Close() error { f.closed = true; return nil }

type fakeConn struct {
	net.Conn
	local, remote *net.TCPAddr
}

func (c *fakeConn) LocalAddr() net.Addr  { return c.local }
func (c *fakeConn) RemoteAddr() net.Addr { return c.remote }

func tcpAddr(ip string, port int) *net.TCPAddr {
	return &net.TCPAddr{IP: net.ParseIP(ip), Port: port}
}

func fakeDialer(limit int) func(context.Context) (net.Conn, error) {
	port := 40000
	calls := 0
	return func(context.Context) (net.Conn, error) {
		calls++
		if limit > 0 && calls > limit {
			return nil, errors.New("dial refusé")
		}
		a, b := net.Pipe()
		_ = b
		c := &fakeConn{
			Conn:   a,
			local:  tcpAddr("127.0.0.1", port),
			remote: tcpAddr("127.0.0.1", 5201),
		}
		port++
		return c, nil
	}
}

func sendBytes(n uint64) func(context.Context, net.Conn) (uint64, error) {
	return func(context.Context, net.Conn) (uint64, error) { return n, nil }
}

func withKernel(t *testing.T, k kernelSession, startErr error) {
	t.Helper()
	oldStart, oldTCP := startKernel, tcpRetransBytes
	startKernel = func() (kernelSession, error) { return k, startErr }
	tcpRetransBytes = func(net.Conn) (uint64, error) { return 0, errors.New("TCP_INFO absent") }
	t.Cleanup(func() {
		startKernel, tcpRetransBytes = oldStart, oldTCP
	})
}

func TestBulkNStatsKernelComplete(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {
				Local: "127.0.0.1:40000", Peer: "127.0.0.1:5201",
				RetransPayload: 1000, RetransEvents: 2, Samples: 3,
			},
			"127.0.0.1:40001 127.0.0.1:5201": {
				Local: "127.0.0.1:40001", Peer: "127.0.0.1:5201",
				RetransPayload: 500, RetransEvents: 1, Samples: 1,
			},
		},
		globals: bpf.Globals{DropPackets: 2, DropBytes: 3000, QdiscAvailable: true},
		attach:  "fentry",
	}
	withKernel(t, k, nil)

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 2)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "ebpf" || !st.RetransMeasured || st.RetransBytes != 1500 {
		t.Fatalf("stats = %+v", st)
	}
	if st.Kernel == nil || !st.Kernel.Available || st.Kernel.FlowsSeen != 2 || st.Kernel.FlowsTotal != 2 {
		t.Fatalf("kernel = %+v", st.Kernel)
	}
	if st.Kernel.Attach != "fentry" || !st.Kernel.QdiscAvailable || st.Kernel.QdiscDrops != 2 || st.Kernel.QdiscDropBytes != 3000 {
		t.Fatalf("kernel global = %+v", st.Kernel)
	}
	if !k.closed {
		t.Fatal("session non fermée")
	}
}

// Un flux à zéro réémission mais avec un échantillon est un zéro EXACT.
func TestBulkNStatsZeroRetransmitWithSample(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {Samples: 4},
		},
	}
	withKernel(t, k, nil)

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "ebpf" || !st.RetransMeasured || st.RetransBytes != 0 {
		t.Fatalf("zéro mesuré attendu : %+v", st)
	}
}

// Un flux présent seulement par un drop qdisc n'est pas observé : repli total.
func TestBulkNStatsDropOnlyDoesNotObserve(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {DropPackets: 1, DropBytes: 100},
		},
	}
	withKernel(t, k, nil)
	tcpRetransBytes = func(net.Conn) (uint64, error) { return 4242, nil }

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "tcpinfo" || !st.RetransMeasured || st.RetransBytes != 4242 {
		t.Fatalf("repli TCP_INFO attendu : %+v", st)
	}
	if st.Kernel == nil || !strings.Contains(st.Kernel.FallbackReason, "flux non observés") {
		t.Fatalf("raison absente : %+v", st.Kernel)
	}
}

// Couverture partielle : jamais de somme mixte eBPF/TCP_INFO.
func TestBulkNStatsPartialFallsBackWholeWindow(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {RetransPayload: 1000, RetransEvents: 1, Samples: 2},
		},
	}
	withKernel(t, k, nil)
	tcpRetransBytes = func(net.Conn) (uint64, error) { return 4242, nil }

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 2)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "tcpinfo" || st.RetransBytes != 8484 {
		t.Fatalf("stats = %+v", st)
	}
	if st.Kernel == nil || st.Kernel.FlowsSeen != 1 || st.Kernel.FlowsTotal != 2 || st.Kernel.Source != "tcpinfo" {
		t.Fatalf("couverture = %+v", st.Kernel)
	}
}

func TestBulkNStatsMapFullFallsBack(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {RetransPayload: 1000, RetransEvents: 1, Samples: 2},
		},
		globals: bpf.Globals{MapFull: 1},
	}
	withKernel(t, k, nil)
	tcpRetransBytes = func(net.Conn) (uint64, error) { return 10, nil }

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "tcpinfo" || st.Kernel == nil || st.Kernel.MapFull != 1 {
		t.Fatalf("map full : %+v", st)
	}
	if !strings.Contains(st.Kernel.FallbackReason, "table de flux pleine") {
		t.Fatalf("raison = %q", st.Kernel.FallbackReason)
	}
}

func TestBulkNStatsGlobalsErrorKeepsEBPF(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {RetransPayload: 1000, RetransEvents: 1, Samples: 2},
		},
		gerr: errors.New("lookup globals"),
	}
	withKernel(t, k, nil)

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "ebpf" || st.RetransBytes != 1000 || st.Kernel == nil || st.Kernel.FallbackReason != "" {
		t.Fatalf("stats = %+v", st)
	}
	if st.Kernel.QdiscAvailable {
		t.Fatal("drops annoncés disponibles malgré l'échec des globaux")
	}
}

func TestBulkNStatsSessionUnavailableKeepsTCPINFO(t *testing.T) {
	withKernel(t, nil, &bpf.UnavailableError{Reason: "CAP_BPF absent"})
	tcpRetransBytes = func(net.Conn) (uint64, error) { return 1000, nil }

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "tcpinfo" || !st.RetransMeasured || st.RetransBytes != 1000 {
		t.Fatalf("stats = %+v", st)
	}
	if st.Kernel == nil || st.Kernel.Available || !strings.Contains(st.Kernel.FallbackReason, "CAP_BPF") {
		t.Fatalf("kernel = %+v", st.Kernel)
	}
}

func TestBulkNStatsNoKernelNoTCPINFO(t *testing.T) {
	withKernel(t, nil, errors.New("boom"))

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), upSenderKey, "up", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "" || st.RetransMeasured || st.RetransBytes != 0 {
		t.Fatalf("aucun compteur attendu : %+v", st)
	}
	if st.Kernel == nil || !strings.Contains(st.Kernel.FallbackReason, "boom") {
		t.Fatalf("raison = %+v", st.Kernel)
	}
}

func TestBulkNStatsFailedDialsNotCounted(t *testing.T) {
	k := &fakeKernel{
		flows: map[string]bpf.Flow{
			"127.0.0.1:40000 127.0.0.1:5201": {Samples: 1},
		},
	}
	withKernel(t, k, nil)

	st, err := bulkNStats(context.Background(), fakeDialer(1), sendBytes(10e6), upSenderKey, "up", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Bytes) != 1 || st.Kernel.FlowsTotal != 1 || st.Kernel.FlowsSeen != 1 {
		t.Fatalf("stats = %+v", st)
	}
}

// Un download depuis un serveur distant : aucune clé locale ne le trouve ;
// la raison le dit au lieu de prétendre un émetteur local.
func TestBulkNStatsRemoteDownloadSender(t *testing.T) {
	withKernel(t, &fakeKernel{flows: map[string]bpf.Flow{}}, nil)
	tcpRetransBytes = func(net.Conn) (uint64, error) { return 4242, nil }

	st, err := bulkNStats(context.Background(), fakeDialer(0), sendBytes(10e6), downSenderKey, "down", 1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != "tcpinfo" || !st.RetransMeasured || st.RetransBytes != 4242 {
		t.Fatalf("stats = %+v", st)
	}
	if st.Kernel == nil || !strings.Contains(st.Kernel.FallbackReason, "émetteur distant") {
		t.Fatalf("raison = %+v", st.Kernel)
	}
}

func TestSenderKeys(t *testing.T) {
	c := &fakeConn{local: tcpAddr("127.0.0.1", 40000), remote: tcpAddr("10.0.0.2", 5201)}

	local, peer, err := upSenderKey(c)
	if err != nil || local.Port() != 40000 || peer.Port() != 5201 {
		t.Fatalf("montée = %v %v %v", local, peer, err)
	}
	local, peer, err = downSenderKey(c)
	if err != nil || local.Port() != 5201 || peer.Port() != 40000 {
		t.Fatalf("descente = %v %v %v", local, peer, err)
	}

	v6 := &fakeConn{local: tcpAddr("::1", 40000), remote: tcpAddr("::1", 5201)}
	if _, _, err := upSenderKey(v6); err == nil {
		t.Fatal("IPv6 doit être refusé")
	}
}
```

- [ ] **Step 2: Run the bulk tests to verify they fail**

Run:

```sh
go test ./pkg/probe/ -run 'TestBulkNStats|TestSenderKeys' -count=1
```

Expected: FAIL to compile — `startKernel`, `tcpRetransBytes`, the new
`bulkNStats` arity, `upSenderKey`, `downSenderKey`, `BulkStats.Source`, and
`BulkStats.Kernel` are undefined.

- [ ] **Step 3: Implement the probe integration**

Create `pkg/probe/kernel.go`:

```go
package probe

import "github.com/Realms4239/cgo/pkg/probe/bpf"

// KernelTelemetry — le résumé eBPF d'une fenêtre, exposé au rapport sans que
// pkg/audit importe le chargeur.
type KernelTelemetry = bpf.Summary
```

Replace `pkg/probe/bulkstats.go` with:

```go
package probe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"

	"github.com/Realms4239/cgo/pkg/probe/bpf"
)

// BulkStats — charge multi-flux avec la vérité par flux ET le ratio de
// réémission : JFI se calcule sur Bytes, le modèle A sur RetransBytes.
type BulkStats struct {
	Bytes           []uint64          `json:"bytes"`
	RetransBytes    uint64            `json:"retrans_bytes"`
	RetransMeasured bool              `json:"retrans_measured"`
	Source          string            `json:"source,omitempty"` // ebpf | tcpinfo | ""
	Kernel          *KernelTelemetry  `json:"kernel,omitempty"`
}

// kernelSession — seam de test : le chargeur réel vit sur linux/amd64.
type kernelSession interface {
	Lookup(local, peer netip.AddrPort) (bpf.Flow, bool)
	Snapshot() []bpf.Flow
	Globals() (bpf.Globals, error)
	Attach() string
	Close() error
}

var (
	startKernel     = func() (kernelSession, error) { return bpf.Start() }
	tcpRetransBytes = TCPRetransBytes
)

// keepConn — enveloppe qui retient Close : les primitives de charge ferment
// la connexion en rendant la main, mais l'échantillon TCP_INFO et la lecture
// de la table noyau exigent un socket VIVANT. bulkNStats neutralise donc la
// fermeture pendant la charge, échantillonne, puis ferme la vraie connexion.
type keepConn struct{ net.Conn }

func (keepConn) Close() error { return nil }

func addrPort(a net.Addr) (netip.AddrPort, error) {
	ta, ok := a.(*net.TCPAddr)
	if !ok {
		return netip.AddrPort{}, fmt.Errorf("adresse TCP attendue, reçu %T", a)
	}
	ap := ta.AddrPort()
	// net.ParseIP et LocalAddr rendent parfois une IPv4 mappée en 16 octets :
	// Unmap la ramène à la forme 4 octets attendue par flowKeyFrom.
	ip := ap.Addr().Unmap()
	if !ip.Is4() {
		return netip.AddrPort{}, fmt.Errorf("IPv4 requis pour la télémétrie eBPF")
	}
	return netip.AddrPortFrom(ip, ap.Port()), nil
}

// upSenderKey — l'émetteur d'un upload est le client local.
func upSenderKey(c net.Conn) (netip.AddrPort, netip.AddrPort, error) {
	local, err := addrPort(c.LocalAddr())
	if err != nil {
		return netip.AddrPort{}, netip.AddrPort{}, err
	}
	peer, err := addrPort(c.RemoteAddr())
	if err != nil {
		return netip.AddrPort{}, netip.AddrPort{}, err
	}
	return local, peer, nil
}

// downSenderKey — l'émetteur d'un download est le serveur, vu par le client :
// local = RemoteAddr, pair = LocalAddr. Aucun fd serveur n'est requis.
func downSenderKey(c net.Conn) (netip.AddrPort, netip.AddrPort, error) {
	local, err := addrPort(c.RemoteAddr())
	if err != nil {
		return netip.AddrPort{}, netip.AddrPort{}, err
	}
	peer, err := addrPort(c.LocalAddr())
	if err != nil {
		return netip.AddrPort{}, netip.AddrPort{}, err
	}
	return local, peer, nil
}

// bulkNStats — moteur commun : N connexions, une session noyau par charge,
// lecture de l'émetteur AVANT la vraie fermeture. dir ne sert qu'à nommer le
// repli « émetteur distant » d'un download, jamais à changer la mesure.
func bulkNStats(ctx context.Context, dial func(context.Context) (net.Conn, error), send func(context.Context, net.Conn) (uint64, error), key func(net.Conn) (netip.AddrPort, netip.AddrPort, error), dir string, n int) (BulkStats, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	st := BulkStats{Bytes: make([]uint64, 0, n)}
	oneTCP := false
	total := 0
	var flows []bpf.Flow
	fallback := ""
	keyErrReason := ""

	ks, kerr := startKernel()
	if kerr != nil {
		ks = nil // jamais une interface non-nil sur un pointeur nil
		fallback = kerr.Error()
	}
	if ks != nil {
		defer ks.Close()
	}

	for i := 0; i < n; i++ {
		c, err := dial(ctx)
		if err != nil {
			continue // lien en défaut : les flux vivants comptent
		}
		total++
		local, peer, keyErr := key(c)
		wg.Add(1)
		go func(c net.Conn, local, peer netip.AddrPort, keyErr error) {
			defer wg.Done()
			b, _ := send(ctx, keepConn{c}) // fermeture retenue, pas perdue
			rb, rerr := tcpRetransBytes(c)
			var flow bpf.Flow
			observed := false
			if ks != nil && keyErr == nil {
				flow, observed = ks.Lookup(local, peer)
				observed = observed && (flow.Samples > 0 || flow.RetransEvents > 0)
			}
			c.Close()
			mu.Lock()
			defer mu.Unlock()
			st.Bytes = append(st.Bytes, b)
			st.RetransBytes += rb
			if rerr == nil {
				oneTCP = true
			}
			if observed {
				flows = append(flows, flow)
			}
			if keyErr != nil && keyErrReason == "" {
				keyErrReason = keyErr.Error()
			}
		}(c, local, peer, keyErr)
	}
	wg.Wait()

	observed := len(flows)
	st.RetransMeasured = oneTCP
	if oneTCP {
		st.Source = "tcpinfo"
	}

	if ks == nil {
		st.Kernel = &KernelTelemetry{
			Available:      false,
			Source:         st.Source,
			FallbackReason: fallback,
			ObjectSHA256:   bpf.ObjectSHA256(),
			FlowsSeen:      observed,
			FlowsTotal:     total,
			Flows:          flows,
		}
		return st, nil
	}

	g, gerr := ks.Globals()
	sum := &KernelTelemetry{
		Available:    true,
		Attach:       ks.Attach(),
		Source:       st.Source,
		ObjectSHA256: bpf.ObjectSHA256(),
		FlowsSeen:    observed,
		FlowsTotal:   total,
		Flows:        flows,
	}
	for _, f := range flows {
		sum.RetransPayload += f.RetransPayload
		sum.RetransEvents += f.RetransEvents
	}
	if gerr == nil {
		sum.QdiscAvailable = g.QdiscAvailable
		sum.QdiscDrops = g.DropPackets
		sum.QdiscDropBytes = g.DropBytes
		sum.MapFull = g.MapFull
	}
	// Les globaux sont des diagnostics : leur échec ne jette pas une mesure
	// par flux complète. Seul map_full invalide la couverture.
	switch {
	case gerr == nil && g.MapFull > 0:
		fallback = "table de flux pleine (map_full)"
	case total == 0:
		fallback = "aucun flux établi"
	case observed == 0 && dir == "down" && keyErrReason == "":
		// Aucun socket local ne porte la clé du serveur : download distant.
		fallback = "émetteur distant (aucun flux émetteur local observé)"
	case observed < total && keyErrReason != "":
		fallback = fmt.Sprintf("flux non observés (%d/%d) — %s", observed, total, keyErrReason)
	case observed < total:
		fallback = fmt.Sprintf("flux non observés (%d/%d)", observed, total)
	}

	if fallback == "" && total > 0 {
		// Fenêtre complète : le payload noyau remplace TCP_INFO, jamais additionné.
		st.RetransBytes = sum.RetransPayload
		st.RetransMeasured = true
		st.Source = "ebpf"
		sum.Source = "ebpf"
	} else {
		sum.FallbackReason = fallback
	}
	st.Kernel = sum
	return st, nil
}

// BulkSendNStats — charge MONTANTE multi-flux avec statistiques.
func BulkSendNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error) {
	return bulkNStats(ctx,
		func(ctx context.Context) (net.Conn, error) { return DialWithCC(ctx, addr, cc) },
		func(ctx context.Context, c net.Conn) (uint64, error) { return BulkSend(ctx, c) },
		upSenderKey, "up", n)
}

// BulkDownloadNStats — charge DESCENDANTE multi-flux avec statistiques.
func BulkDownloadNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error) {
	return bulkNStats(ctx,
		func(ctx context.Context) (net.Conn, error) { return DialWithCC(ctx, addr, cc) },
		func(ctx context.Context, c net.Conn) (uint64, error) { return BulkRecvSource(ctx, c, cc) },
		downSenderKey, "down", n)
}
```

- [ ] **Step 4: Run the probe tests**

Run:

```sh
go test ./pkg/probe/ -count=1
```

Expected: PASS on Windows.

- [ ] **Step 5: Write the failing metrics and audit tests**

Add to `pkg/metrics/cost_test.go`:

```go
func TestRetransRatioSourceJSON(t *testing.T) {
	b, err := json.Marshal(RetransRatio{Ratio: 0.01, Basis: "measured", Source: "ebpf"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"source":"ebpf"`) {
		t.Fatalf("source absente : %s", b)
	}
	b, err = json.Marshal(RetransRatio{Ratio: 0.01, Basis: "measured"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"source"`) {
		t.Fatalf("source vide doit être omise : %s", b)
	}
}
```

Change the import line to:

```go
import (
	"encoding/json"
	"strings"
	"testing"
)
```

Add to `pkg/audit/audit_test.go`:

```go
// La télémétrie noyau et les deux ratios de fenêtre sont recopiés sans
// changer le ratio headline.
func TestRunKernelTelemetryAndSource(t *testing.T) {
	d := Deps{
		Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{20, 21, 22} },
		Small: func(context.Context) (float64, error) { return 0, fmt.Errorf("no small URL") },
		BulkLoad: func(_ context.Context, _ string, dir string, _ int) LoadOutcome {
			return LoadOutcome{
				Bytes: 10e6, PerFlow: []uint64{10e6},
				RetransBytes: 0, RetransMeasured: true, RetransSource: "ebpf",
				Kernel: &probe.KernelTelemetry{
					Available: true, Source: "ebpf", Attach: "fentry",
					FlowsSeen: 1, FlowsTotal: 1,
				},
				Elapsed: 5 * time.Second,
			}
		},
	}
	res, err := Run(context.Background(), Params{Duration: 20, Direction: "up", Target: "x"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if res.RetransRatio.Source != "ebpf" || res.RetransRatio.Basis != "measured" || res.RetransRatio.Ratio != 0 {
		t.Fatalf("zéro mesuré eBPF attendu : %+v", res.RetransRatio)
	}
	if res.KernelUp == nil || res.KernelDown != nil {
		t.Fatalf("fenêtres = %+v / %+v", res.KernelUp, res.KernelDown)
	}
	if res.RetransRatioUp == nil || res.RetransRatioUp.Source != "ebpf" || res.RetransRatioDown != nil {
		t.Fatalf("ratios = %+v / %+v", res.RetransRatioUp, res.RetransRatioDown)
	}
}

func TestWindowRatioSources(t *testing.T) {
	oc := LoadOutcome{Bytes: 1000, RetransBytes: 10, RetransMeasured: true, RetransSource: "tcpinfo"}
	if got := windowRatio(oc, true, 0); got.Basis != "measured" || got.Source != "tcpinfo" {
		t.Fatalf("TCP_INFO = %+v", got)
	}
	oc.RetransSource = "ebpf"
	if got := windowRatio(oc, true, 0); got.Basis != "measured" || got.Source != "ebpf" {
		t.Fatalf("eBPF = %+v", got)
	}
	if got := windowRatio(LoadOutcome{Bytes: 1000}, true, 4); got.Basis != "estimated" || got.Source != "" {
		t.Fatalf("proxy = %+v", got)
	}
}

func TestPickRatioRanksEBPFAboveTCPINFO(t *testing.T) {
	ebpf := metrics.RetransRatio{Ratio: 0.001, Basis: "measured", Source: "ebpf"}
	tcp := metrics.RetransRatio{Ratio: 0.5, Basis: "measured", Source: "tcpinfo"}
	if got := pickRatio(tcp, ebpf); got.Source != "ebpf" {
		t.Fatalf("eBPF doit gagner : %+v", got)
	}
	est := metrics.RetransRatio{Ratio: 0.9, Basis: "estimated"}
	if got := pickRatio(est, tcp); got.Source != "tcpinfo" {
		t.Fatalf("mesuré doit gagner sur estimé : %+v", got)
	}
	low := metrics.RetransRatio{Ratio: 0.01, Basis: "measured", Source: "ebpf"}
	high := metrics.RetransRatio{Ratio: 0.02, Basis: "measured", Source: "ebpf"}
	if got := pickRatio(low, high); got.Ratio != 0.02 {
		t.Fatalf("à rang égal, la plus grande gagne : %+v", got)
	}
}
```

Add `"github.com/Realms4239/cgo/pkg/metrics"` and
`"github.com/Realms4239/cgo/pkg/probe"` to the `audit_test.go` imports.

Add to `pkg/audit/writer_test.go`:

```go
func TestAppendLinkAuditIgnoresKernelFields(t *testing.T) {
	dir := t.TempDir()
	res := &Result{
		AuditID: "audit-k", Timestamp: "t", Site: "S", LinkType: "4g",
		KernelUp: &probe.KernelTelemetry{Available: true, Source: "ebpf", FlowsSeen: 4, FlowsTotal: 4},
	}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "link_audit.csv"))
	if strings.Contains(string(raw), "kernel") {
		t.Fatalf("les champs noyau ne doivent pas entrer dans le CSV :\n%s", raw)
	}
}
```

Add `"github.com/Realms4239/cgo/pkg/probe"` to the `writer_test.go` imports.

- [ ] **Step 6: Run the metrics and audit tests to verify they fail**

Run:

```sh
go test ./pkg/metrics/ ./pkg/audit/ -count=1
```

Expected: FAIL to compile — `RetransRatio.Source`, `LoadOutcome.RetransSource`,
`LoadOutcome.Kernel`, `Result.KernelUp/Down`, `Result.RetransRatioUp/Down`,
and the `probe.KernelTelemetry` alias field wiring are missing.

- [ ] **Step 7: Implement the metrics source**

In `pkg/metrics/cost.go`, replace the `RetransRatio` struct:

```go
// RetransRatio — part des octets réémis. Basis dit d'où elle vient :
// "measured" (compteur noyau ou TCP_INFO sur nos sockets), "estimated" (proxy
// perte sonde × volume), "" (non mesuré — alors aucun scénario n'est chiffré).
// Source nomme le compteur mesuré : "ebpf", "tcpinfo", ou "" si non mesuré.
type RetransRatio struct {
	Ratio  float64 `json:"ratio"`
	Basis  string  `json:"basis"`
	Source string  `json:"source,omitempty"`
}
```

- [ ] **Step 8: Implement the audit wiring**

In `pkg/audit/audit.go` (also refresh the two stale comments: the `upRatio`
comment says “TCP_INFO d'abord”, and `Result.RetransRatio` says “TCP_INFO sur
nos sockets” — both now name eBPF first):

Add to `LoadOutcome`:

```go
	RetransSource string
	Kernel        *probe.KernelTelemetry
```

Add to `Result` after `RetransRatio`:

```go
	// Télémétrie noyau et ratios par fenêtre — JSON seulement, jamais CSV.
	KernelUp         *probe.KernelTelemetry `json:"kernel_up,omitempty"`
	KernelDown       *probe.KernelTelemetry `json:"kernel_down,omitempty"`
	RetransRatioUp   *metrics.RetransRatio  `json:"retrans_ratio_up,omitempty"`
	RetransRatioDown *metrics.RetransRatio  `json:"retrans_ratio_down,omitempty"`
```

In the default `BulkLoad` closure, replace the `out := LoadOutcome{...}` block:

```go
			out := LoadOutcome{Bytes: sumU64(st.Bytes), PerFlow: st.Bytes,
				RetransBytes: st.RetransBytes, RetransMeasured: st.RetransMeasured,
				RetransSource: st.Source, Kernel: st.Kernel,
				Elapsed: time.Since(t0)}
```

In the `res := &Result{...}` literal, add:

```go
		KernelUp:         upOc.Kernel,
		KernelDown:       downOc.Kernel,
		RetransRatioUp:   ratioPtr(upRatio),
		RetransRatioDown: ratioPtr(downRatio),
```

Replace `windowRatio` and `pickRatio`:

```go
// windowRatio — ratio de réémission d'UNE fenêtre : mesure noyau exacte
// d'abord, sinon TCP_INFO, sinon le proxy documenté perte sonde × volume.
// Les sources ne sont jamais additionnées : un appel = une seule source.
func windowRatio(oc LoadOutcome, measured bool, lossPct float64) metrics.RetransRatio {
	if !measured {
		return metrics.RetransRatio{}
	}
	total := float64(oc.Bytes)
	switch {
	case oc.RetransMeasured && total > 0:
		return metrics.RetransRatio{
			Ratio:  float64(oc.RetransBytes) / total,
			Basis:  "measured",
			Source: oc.RetransSource,
		}
	case lossPct > 0 && total > 0:
		return metrics.RetransRatio{Ratio: lossPct / 100, Basis: "estimated"}
	}
	return metrics.RetransRatio{}
}

// ratioPtr — nil plutôt qu'un « 0, base vide » qui se lirait comme mesuré.
func ratioPtr(r metrics.RetransRatio) *metrics.RetransRatio {
	if r.Basis == "" {
		return nil
	}
	return &r
}

// pickRatio — pour "both", une seule ratio headline : eBPF d'abord, puis
// TCP_INFO, puis le proxy ; à rang égal, la plus grande des deux fenêtres.
func pickRatio(a, b metrics.RetransRatio) metrics.RetransRatio {
	rank := func(r metrics.RetransRatio) int {
		switch {
		case r.Source == "ebpf":
			return 3
		case r.Basis == "measured":
			return 2
		case r.Basis == "estimated":
			return 1
		}
		return 0
	}
	if rank(a) != rank(b) {
		if rank(a) > rank(b) {
			return a
		}
		return b
	}
	if b.Ratio > a.Ratio {
		return b
	}
	return a
}
```

- [ ] **Step 9: Run the probe, metrics, and audit tests**

Run:

```sh
go test ./pkg/probe/ ./pkg/metrics/ ./pkg/audit/ -count=1
```

Expected: PASS.

- [ ] **Step 10: Commit**

```sh
git add pkg/probe/kernel.go pkg/probe/bulkstats.go pkg/probe/bulkstats_test.go \
  pkg/metrics/cost.go pkg/metrics/cost_test.go \
  pkg/audit/audit.go pkg/audit/audit_test.go pkg/audit/writer_test.go
git commit -m "probe/audit: télémétrie noyau dans le ratio de réémission"
```

---

### Task 3: Report block, Markdown section, doctor, and API docs

**Files:**
- Modify: `pkg/audit/report.go`, `pkg/audit/report_test.go`
- Modify: `pkg/doctor/doctor.go`
- Create: `pkg/doctor/ebpf_test.go`
- Modify: `docs/api.md`

**Interfaces:**
- Consumes: `probe.KernelTelemetry` (alias of `bpf.Summary`), `metrics.RetransRatio.Source`, `audit.Result.KernelUp/Down` and `.RetransRatioUp/Down`.
- Produces: `audit.Report.KernelTelemetry`, `audit.KernelTelemetry`, `audit.KernelWindow`, `doctor.ebpfCheckWith`.

- [ ] **Step 1: Write the failing report tests**

Add to `pkg/audit/report_test.go`:

```go
func TestBuildReportKernelTelemetry(t *testing.T) {
	res := loadedResult()
	res.LoadSeconds = 40
	res.DataUsedMB = 50
	res.RetransRatio = metrics.RetransRatio{Ratio: 0.01, Basis: "measured", Source: "ebpf"}
	res.KernelUp = &probe.KernelTelemetry{
		Available: true, Source: "ebpf", Attach: "fentry", ObjectSHA256: "abc123",
		FlowsSeen: 4, FlowsTotal: 4, RetransPayload: 12345, RetransEvents: 9,
		QdiscDrops: 3, QdiscDropBytes: 4344, QdiscAvailable: true,
	}
	res.RetransRatioUp = &metrics.RetransRatio{Ratio: 0.01, Basis: "measured", Source: "ebpf"}

	rep, err := BuildReport(res, testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.KernelTelemetry == nil || rep.KernelTelemetry.Up == nil {
		t.Fatal("kernel_telemetry.up absent")
	}

	js, err := rep.RenderJSON()
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(js, &got); err != nil {
		t.Fatal(err)
	}
	kt, ok := got["kernel_telemetry"].(map[string]any)
	if !ok {
		t.Fatal("kernel_telemetry absent du JSON")
	}
	up, ok := kt["up"].(map[string]any)
	if !ok {
		t.Fatalf("kernel_telemetry.up = %+v", kt["up"])
	}
	if up["source"] != "ebpf" || up["retrans_payload_bytes"].(float64) != 12345 {
		t.Fatalf("up = %+v", up)
	}
	ratio, ok := up["retrans_ratio"].(map[string]any)
	if !ok || ratio["source"] != "ebpf" {
		t.Fatalf("retrans_ratio up = %+v", ratio)
	}

	md := rep.RenderMarkdown()
	for _, want := range []string{
		"## Télémétrie noyau",
		"eBPF (échantillon socket fentry)",
		"4/4 flux",
		"12345 octets de réémission transmis (payload)",
		"9 événements",
		"Drops qdisc (hôte, tous flux) : 3 paquets / 4344 octets",
		"sha256:abc123",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("Markdown sans %q :\n%s", want, md)
		}
	}
}

func TestBuildReportNoKernelTelemetrySection(t *testing.T) {
	rep, err := BuildReport(loadedResult(), testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.KernelTelemetry != nil {
		t.Fatalf("bloc inattendu : %+v", rep.KernelTelemetry)
	}
	if strings.Contains(rep.RenderMarkdown(), "Télémétrie noyau") {
		t.Fatal("section affichée sans résumé noyau")
	}
}

func TestBuildReportKernelFallbackLines(t *testing.T) {
	res := loadedResult()
	res.KernelUp = &probe.KernelTelemetry{
		Available: true, Source: "tcpinfo", FallbackReason: "CAP_BPF absent",
	}
	rep, err := BuildReport(res, testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "repli TCP_INFO") || !strings.Contains(md, "CAP_BPF absent") {
		t.Fatalf("repli absent :\n%s", md)
	}
}

func TestBuildReportKernelUnmeasuredLine(t *testing.T) {
	res := loadedResult()
	res.KernelUp = &probe.KernelTelemetry{
		Available: false, FallbackReason: "linux/amd64 requis",
	}
	rep, err := BuildReport(res, testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "non mesuré") || !strings.Contains(md, "linux/amd64 requis") {
		t.Fatalf("absence non dite :\n%s", md)
	}
}
```

Add `"github.com/Realms4239/cgo/pkg/probe"` to the `report_test.go` imports.

- [ ] **Step 2: Run the report tests to verify they fail**

Run:

```sh
go test ./pkg/audit/ -run 'TestBuildReportKernel' -count=1
```

Expected: FAIL to compile — `Report.KernelTelemetry`, `KernelTelemetry`,
`KernelWindow` are undefined.

- [ ] **Step 3: Implement the report block and Markdown section**

In `pkg/audit/report.go`:

Add the `probe` import.

Add to `Report` after `CAPSLA`:

```go
	// KernelTelemetry — télémétrie noyau par fenêtre ; absent quand aucune
	// session n'a été ouverte. Les drops sont des diagnostics hôte.
	KernelTelemetry *KernelTelemetry `json:"kernel_telemetry,omitempty"`
```

Add the types after `Report`:

```go
// KernelTelemetry — résumés up/down d'un audit. Une fenêtre non exécutée est
// nil, jamais un bloc à zéro.
type KernelTelemetry struct {
	Up   *KernelWindow `json:"up,omitempty"`
	Down *KernelWindow `json:"down,omitempty"`
}

// KernelWindow — le résumé eBPF et le ratio qui en a été tiré.
type KernelWindow struct {
	*probe.KernelTelemetry
	RetransRatio *metrics.RetransRatio `json:"retrans_ratio,omitempty"`
}
```

In `BuildReport`, after the `CAPSLA` assignment:

```go
	if res.KernelUp != nil || res.KernelDown != nil {
		kt := &KernelTelemetry{}
		if res.KernelUp != nil {
			kt.Up = &KernelWindow{KernelTelemetry: res.KernelUp, RetransRatio: res.RetransRatioUp}
		}
		if res.KernelDown != nil {
			kt.Down = &KernelWindow{KernelTelemetry: res.KernelDown, RetransRatio: res.RetransRatioDown}
		}
		rep.KernelTelemetry = kt
	}
```

In `RenderMarkdown`, after the CAP SLA block, call:

```go
	r.renderKernelTelemetry(&b)
```

Add the renderer after `RenderMarkdown`:

```go
// renderKernelTelemetry — une section, un mode par fenêtre exécutée.
func (r *Report) renderKernelTelemetry(b *strings.Builder) {
	if r.KernelTelemetry == nil {
		return
	}
	up, down := r.KernelTelemetry.Up, r.KernelTelemetry.Down
	if up == nil && down == nil {
		return
	}
	b.WriteString("## Télémétrie noyau\n\n")

	sha := ""
	if up != nil {
		sha = up.ObjectSHA256
	}
	if sha == "" && down != nil {
		sha = down.ObjectSHA256
	}

	rows := []struct {
		label string
		win   *KernelWindow
	}{{"Montée", up}, {"Descente", down}}
	both := up != nil && down != nil
	for _, row := range rows {
		if row.win == nil {
			continue
		}
		prefix := ""
		if both {
			prefix = row.label + " : "
		}
		switch {
		case row.win.Source == "ebpf":
			fmt.Fprintf(b, "- %sMode : eBPF (échantillon socket %s) — %d/%d flux · %d octets de réémission transmis (payload) · %d événements.\n",
				prefix, reasonOr(row.win.Attach, "inconnu"),
				row.win.FlowsSeen, row.win.FlowsTotal,
				row.win.RetransPayload, row.win.RetransEvents)
		case row.win.Source == "tcpinfo":
			fmt.Fprintf(b, "- %sMode : repli TCP_INFO — %s ; le ratio reste mesuré par TCP_INFO.\n",
				prefix, reasonOr(row.win.FallbackReason, "raison non rapportée"))
		default:
			fmt.Fprintf(b, "- %sMode : non mesuré — %s ; aucun ratio noyau.\n",
				prefix, reasonOr(row.win.FallbackReason, "linux/amd64 requis"))
		}
		if row.win.QdiscAvailable {
			fmt.Fprintf(b, "- %sDrops qdisc (hôte, tous flux) : %d paquets / %d octets (raison QDISC_DROP).\n",
				prefix, row.win.QdiscDrops, row.win.QdiscDropBytes)
		} else if row.win.Available {
			fmt.Fprintf(b, "- %sDrops qdisc : non disponibles (sonde non attachée ou compteurs illisibles).\n", prefix)
		}
	}
	if sha != "" {
		fmt.Fprintf(b, "- Objet : sha256:%s\n", sha)
	}
	b.WriteString("\n")
}

func reasonOr(reason, fallback string) string {
	if strings.TrimSpace(reason) == "" {
		return fallback
	}
	return reason
}
```

- [ ] **Step 4: Run the report tests**

Run:

```sh
go test ./pkg/audit/ -run 'TestBuildReportKernel|TestBuildReportNoKernel' -count=1
```

Expected: PASS.

- [ ] **Step 5: Write the failing doctor test**

Create `pkg/doctor/ebpf_test.go`:

```go
package doctor

import (
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/probe/bpf"
)

func TestEBPFCheckStatuses(t *testing.T) {
	ok := ebpfCheckWith(func() error { return nil })
	if ok.Name != "ebpf" || ok.Status != "ok" {
		t.Fatalf("ok = %+v", ok)
	}
	warn := ebpfCheckWith(func() error {
		return &bpf.UnavailableError{Reason: "BTF absent"}
	})
	if warn.Status != "warn" || !strings.Contains(warn.Detail, "BTF absent") {
		t.Fatalf("warn = %+v", warn)
	}
	fail := ebpfCheckWith(func() error {
		return &bpf.BuildError{Reason: "sha256"}
	})
	if fail.Status != "fail" || !strings.Contains(fail.Detail, "sha256") {
		t.Fatalf("fail = %+v", fail)
	}
}

func TestReportIncludesEBPFCheck(t *testing.T) {
	_, checks := Report(Mode("auto"))
	for _, c := range checks {
		if c.Name == "ebpf" {
			return
		}
	}
	t.Fatal("check ebpf absent du rapport")
}
```

- [ ] **Step 6: Run the doctor test to verify it fails**

Run:

```sh
go test ./pkg/doctor/ -run 'TestEBPF' -count=1
```

Expected: FAIL to compile — `ebpfCheckWith` is undefined.

- [ ] **Step 7: Implement the doctor check and update the API doc**

In `pkg/doctor/doctor.go`, add to the imports:

```go
	"errors"

	"github.com/Realms4239/cgo/pkg/probe/bpf"
```

In `Report`, after the `bbr` check and before `benchOffloadCheck()`:

```go
	checks = append(checks, ebpfCheck())
```

Add after `Print`:

```go
// ebpfCheck — read-only: SHA + charge sans attache. Un BuildError est un
// défaut d'objet ; toute limitation d'hôte est un avertissement avec repli.
func ebpfCheck() Check { return ebpfCheckWith(bpf.Check) }

func ebpfCheckWith(check func() error) Check {
	err := check()
	if err == nil {
		return Check{Name: "ebpf", Status: "ok", Detail: "objet chargeable (BTF + CAP_BPF) ; l'attache est vérifiée par l'audit"}
	}
	var be *bpf.BuildError
	if errors.As(err, &be) {
		return Check{Name: "ebpf", Status: "fail", Detail: "objet eBPF invalide — " + be.Reason}
	}
	return Check{Name: "ebpf", Status: "warn", Detail: err.Error() + " — repli TCP_INFO"}
}
```

In `docs/api.md`, replace the `/api/doctor` row's capability list:

```markdown
| `/api/doctor` | GET | — | `{"mode", "checks": [{name, status: ok\|warn\|fail, detail}]}` — mêmes capacités que `cgo doctor` (os, tc, cap_net_admin, bbr, ebpf, bench-offloads, ping) | — |
```

- [ ] **Step 8: Run the report and doctor tests**

Run:

```sh
go test ./pkg/audit/ ./pkg/doctor/ -count=1
```

Expected: PASS.

- [ ] **Step 9: Commit**

```sh
git add pkg/audit/report.go pkg/audit/report_test.go \
  pkg/doctor/doctor.go pkg/doctor/ebpf_test.go docs/api.md
git commit -m "audit/report: bloc kernel_telemetry et provenance"
```

---

### Task 4: Opt-in integration test, docs, and final gates

**Files:**
- Create: `pkg/probe/bpf/integration_linux_test.go`
- Modify: `pkg/probe/bpf/BUILD.md` only if the build steps changed
- No production code changes expected; fix regressions in the owning task, not here.

**Interfaces:**
- Consumes: Task 1–3 APIs.
- Produces: the `ebpf_integration` build-tag test; no exported symbols.

- [ ] **Step 1: Write the opt-in integration test**

Create `pkg/probe/bpf/integration_linux_test.go`:

```go
//go:build linux && amd64 && ebpf_integration

package bpf

import (
	"io"
	"net"
	"net/netip"
	"os"
	"testing"
	"time"
)

// TestLoopbackSession — preuve d'attache réelle, jamais exécutée par les
// portes Windows. Root + BTF requis ; le forçage d'une réémission reste une
// étape manuelle du banc (dualns + netem loss).
func TestLoopbackSession(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("root requis")
	}
	if _, err := os.Stat("/sys/kernel/btf/vmlinux"); err != nil {
		t.Skip("BTF absent")
	}
	if err := Check(); err != nil {
		t.Fatalf("Check: %v", err)
	}

	s, err := Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Close()
	if mode := s.Attach(); mode != "fentry" && mode != "kprobe" {
		t.Fatalf("attache = %q", mode)
	}

	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = io.Copy(io.Discard, c)
	}()

	c, err := net.Dial("tcp4", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.Write(make([]byte, 64*1024)); err != nil {
		t.Fatal(err)
	}

	local := netip.MustParseAddrPort(c.LocalAddr().String())
	peer := netip.MustParseAddrPort(c.RemoteAddr().String())
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if f, ok := s.Lookup(local, peer); ok && f.Samples > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("flux loopback non observé (samples=0)")
}
```

- [ ] **Step 2: Compile the integration test for Linux**

Run:

```sh
GOOS=linux GOARCH=amd64 go test -c -tags=ebpf_integration -o "$TEMP/bpf-integration.test" ./pkg/probe/bpf
```

Expected: exit 0. Do not run the binary on Windows.

- [ ] **Step 3: Run the full Windows gates**

Run from `C:\cgo`:

```sh
go test ./... -count=1
go vet ./...
GOOS=linux GOARCH=amd64 go build ./...
GOOS=linux GOARCH=amd64 go vet ./pkg/probe/bpf/ ./pkg/probe/ ./pkg/audit/ ./pkg/doctor/
```

Expected: all exit 0.

Run the frontend gates (unchanged by this slice):

```sh
cd web/frontend
bun run typecheck
bunx vitest run
bun run build
```

Expected: typecheck OK; 37 files / 139 tests PASS; build OK.

- [ ] **Step 4: Run the frozen-corpus guard**

Run from `C:\cgo`:

```sh
python book/tools/audit-archive.py
```

Expected: the same output as before this slice; authoritative count
**159 runs / 208 valid lines**. Do not describe it as “208 runs”.

- [ ] **Step 5: Commit**

```sh
git add pkg/probe/bpf/integration_linux_test.go pkg/probe/bpf/BUILD.md
git commit -m "docs/tests: preuves eBPF et garde-fous"
```

- [ ] **Step 6: Live acceptance (manual, blocked until a stable 6.8 host)**

Do not run this on the unstable VM and do not claim it before it runs.

1. On the `dualns` bench, run a 60 s upload audit and a 60 s download audit
   with `tc netem loss` on the WAN emulator.
2. Expect `source=ebpf`, coverage `4/4`, `retrans_payload_bytes > 0`, and
   qdisc drops when the queue overflows.
3. Run an audit from a remote download sender; expect `source=tcpinfo` or
   empty with “émetteur distant”, never a local sender claim.
4. Build the pre-eBPF binary from `05715f3` and the slice binary; run the
   same 60 s audit alternately (A/B/A), three runs each. Accept only a
   median goodput delta under 2% with all phases executed. There is no
   runtime eBPF switch in this slice; this A/B against the parent commit is
   the honest comparison.

---

## Self-Review

**Spec coverage**
- §0.1/§0.2 scope: Tasks 1–4 only touch telemetry; no BBRv3, `.dist`, IFB, UI.
- §2.1/§2.2 hooks: Task 1 C source uses the raw `tcp_retransmit_skb`
  tracepoint, fentry/kprobe socket sample, and `kfree_skb` qdisc filter.
- §2.3 byte accounting: `payload = skb->len`; no wire counter, no header parse.
- §2.4 sampling: 50 ms `last_sample_ns`.
- §2.5 byte order: `wire.go` C-compatible structs and host-order tests.
- §2.6 no dwell: no timestamp programs, no skb correlation.
- §3 object/maps/loader: Task 1 C source, Makefile, loader, stub, hash.
- §4.1 API: `Session` includes `Globals`/`Attach` (added to the spec to close
  the host-wide-counter gap), `types.go` JSON, `kernel.go` alias.
- §4.2 bulk: one session per load, sender keys, query-before-close,
  all-or-nothing fallback, `flows_total` from successful dials. `Unmap` is
  mandatory because `net.TCPAddr.AddrPort()` renders IPv4 in mapped form;
  a download with zero observed and no key error says “émetteur distant”.
- §4.3 metrics: `Source`, measured zero, `pickRatio` rank.
- §4.4 audit/report: `KernelUp/Down`, `RetransRatioUp/Down`, report block,
  French Markdown.
- §4.5 doctor: `ebpf` ok/warn/fail, read-only. The check is mode-independent:
  `Check` loads without attaching and the audit itself has no mode gate, so a
  Linux `--mode observe` host truthfully reports `ok` (the attach is still
  possible and is exercised by the audit). On Windows the stub reports `warn`
  with the os/arch reason. This is a deliberate divergence from the spec's
  “observe → warn” row, which would otherwise claim a TCP_INFO fallback the
  audit does not perform.
- §5 failure ladder: bulk tests cover no session, partial, zero, drop-only,
  map full, remote download sender, TCP_INFO failure.
- §6 tests: Windows stub, cross-build, unit tests, opt-in integration,
  acceptance notes.
- §8 commits: four commits, exact file lists, no frozen files.

**Placeholder scan:** No “TBD”, “TODO”, or “implement later”. The object hash
is a build output written by `make bpf`, not a placeholder.

**Type consistency:** `Flow`, `Summary`, `Globals` are defined once and reused
through the alias; `LoadOutcome.RetransSource`/`.Kernel` feed
`windowRatio`/`Result`; `Result.KernelUp/Down` feed `BuildReport`; the
Markdown renderer reads `KernelWindow`'s embedded `*probe.KernelTelemetry`
and `RetransRatio`.
