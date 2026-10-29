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
