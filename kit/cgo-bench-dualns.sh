#!/bin/bash
# ============================================================================
#  cgo-bench-dualns.sh — decoupled bench: link physics vs queue policy.
#
#  WHY THIS EXISTS
#  Stacking a child qdisc under a root netem puts the child behind netem's own
#  internal FIFO (default 1000 packets). That queue is invisible to the child,
#  so CAKE's RTT estimator sees a delay it did not create, and any latency
#  measured through the stack is partly netem's buffer rather than the AQM's
#  behaviour. The measurement then attributes netem's queueing to the AQM.
#
#  The fix is topological, not parametric: put link physics and queue
#  management on different interfaces, in different namespaces.
#
#      ns_client          ns_wan_emulator        ns_edge_router        ns_server
#      ┌────────┐  veth  ┌──────────────┐ veth  ┌──────────────┐ veth  ┌────────┐
#      │ client │◄──────►│  netem only  │◄─────►│  CAKE only   │◄─────►│ server │
#      └────────┘        └──────────────┘       └──────────────┘       └────────┘
#                          delay + loss           rate + AQM
#
#  netem carries propagation delay and channel loss and nothing else. CAKE
#  carries bandwidth shaping and the AQM and nothing else. Neither is a child
#  of the other.
#
#  WHY THE SERVER HAS ITS OWN NAMESPACE
#  If the server listens inside ns_edge_router, an upload flow terminates at
#  the router's local address and never egresses any interface — so the CAKE
#  qdisc is bypassed entirely and the audit reports the raw veth rate (this was
#  observed: 193 Mbit/s through a 20 Mbit/s shaper). The server must sit behind
#  the shaped interface for the shaper to be on the data path.
#
#  Usage:
#    cgo-bench-dualns.sh up    [--delay MS] [--loss PCT] [--rate MBIT] [--rtt MS]
#    cgo-bench-dualns.sh down
#    cgo-bench-dualns.sh check
#    cgo-bench-dualns.sh show
#
#  Requires: root (or passwordless sudo for ip/tc/modprobe), Linux >= 5.15.
# ============================================================================
set -uo pipefail

NS_CLIENT=ns_client
NS_WAN=ns_wan_emulator
NS_EDGE=ns_edge_router
NS_SRV=ns_server

# Link physics (netem) and queue policy (CAKE) are separate knobs on purpose.
DELAY_MS=100
LOSS_PCT=0
RATE_MBIT=20
RTT_MS=100
# Hardening default: veth offloads off so GSO super-packets cannot hide drops
# from CAKE. The legacy frozen campaign (kit/testbed.sh) keeps its own ON
# recipe; this bench is the explicit offload-clean comparison.
OFFLOADS=off

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if sudo -n true 2>/dev/null; then SUDO="sudo -n"; else
    echo "[!] need root or passwordless sudo for ip/tc" >&2; exit 1
  fi
fi

log() { echo "[dualns] $*"; }
die() { echo "[!] $*" >&2; exit 1; }

set_offloads() {
  local ns="$1" iface="$2"
  case "$OFFLOADS" in
    on|off) ;;
    *) die "offloads must be on|off";;
  esac
  $SUDO ip netns exec "$ns" ethtool -K "$iface" tso "$OFFLOADS" gso "$OFFLOADS" gro "$OFFLOADS" 2>/dev/null || true
}

check_offloads() {
  local ns="$1" iface="$2" feat line
  for feat in tcp-segmentation-offload generic-segmentation-offload generic-receive-offload; do
    line=$($SUDO ip netns exec "$ns" ethtool -k "$iface" 2>/dev/null | grep -m1 "^${feat}:")
    case "$line" in
      *": $OFFLOADS"*) ;;
      *) echo "[!] $feat not $OFFLOADS on $ns/$iface ($line)"; return 1;;
    esac
  done
}

need_modules() {
  $SUDO modprobe sch_netem sch_cake sch_fq_codel tcp_bbr 2>/dev/null || true
  # BBR is the campaign's reference congestion control; without it the bulk
  # flow cannot fill a high-RTT pipe and no standing queue ever forms.
  if ! grep -qw bbr /proc/sys/net/ipv4/tcp_allowed_congestion_control 2>/dev/null; then
    die "BBR unavailable — modprobe tcp_bbr failed (install linux-modules-extra)"
  fi
}

down() {
  for ns in "$NS_CLIENT" "$NS_WAN" "$NS_EDGE" "$NS_SRV"; do
    $SUDO ip netns del "$ns" 2>/dev/null || true
  done
  log "down"
}

up() {
  need_modules
  down >/dev/null 2>&1 || true

  for ns in "$NS_CLIENT" "$NS_WAN" "$NS_EDGE" "$NS_SRV"; do
    $SUDO ip netns add "$ns" || die "cannot create $ns"
  done

  # --- the three veth pairs that chain the namespaces ---------------------
  # Each pair crosses a namespace boundary, so no two hops can share a qdisc
  # tree. This is what makes the split real rather than cosmetic.
  $SUDO ip link add c-cli netns "$NS_CLIENT" type veth peer name c-wan netns "$NS_WAN" \
    || die "cannot create client<->wan veth"
  $SUDO ip link add w-edge netns "$NS_WAN" type veth peer name e-wan netns "$NS_EDGE" \
    || die "cannot create wan<->edge veth"
  $SUDO ip link add e-srv netns "$NS_EDGE" type veth peer name s-edge netns "$NS_SRV" \
    || die "cannot create edge<->server veth"

  # --- addressing ---------------------------------------------------------
  # 10.10.0.0/30 : client <-> wan_emulator
  # 10.20.0.0/30 : wan_emulator <-> edge_router
  # 10.30.0.0/30 : edge_router <-> server
  $SUDO ip netns exec "$NS_CLIENT" ip addr add 10.10.0.1/30 dev c-cli
  $SUDO ip netns exec "$NS_WAN"    ip addr add 10.10.0.2/30 dev c-wan
  $SUDO ip netns exec "$NS_WAN"    ip addr add 10.20.0.1/30 dev w-edge
  $SUDO ip netns exec "$NS_EDGE"   ip addr add 10.20.0.2/30 dev e-wan
  $SUDO ip netns exec "$NS_EDGE"   ip addr add 10.30.0.1/30 dev e-srv
  $SUDO ip netns exec "$NS_SRV"    ip addr add 10.30.0.2/30 dev s-edge

  for ns in "$NS_CLIENT" "$NS_WAN" "$NS_EDGE" "$NS_SRV"; do
    $SUDO ip netns exec "$ns" ip link set lo up
  done
  $SUDO ip netns exec "$NS_CLIENT" ip link set c-cli up
  $SUDO ip netns exec "$NS_WAN"    ip link set c-wan up
  $SUDO ip netns exec "$NS_WAN"    ip link set w-edge up
  $SUDO ip netns exec "$NS_EDGE"   ip link set e-wan up
  $SUDO ip netns exec "$NS_EDGE"   ip link set e-srv up
  $SUDO ip netns exec "$NS_SRV"    ip link set s-edge up

  # Offload recipe is applied after the links exist (ethtool cannot tune a
  # down/missing device). Both directions are covered: c-cli/c-wan, w-edge,
  # e-wan/e-srv, s-edge.
  set_offloads "$NS_CLIENT" c-cli
  set_offloads "$NS_WAN" c-wan
  set_offloads "$NS_WAN" w-edge
  set_offloads "$NS_EDGE" e-wan
  set_offloads "$NS_EDGE" e-srv
  set_offloads "$NS_SRV" s-edge

  # --- routing ------------------------------------------------------------
  # Forwarding is what puts each hop's qdisc on the actual path.
  for ns in "$NS_WAN" "$NS_EDGE"; do
    $SUDO ip netns exec "$ns" sysctl -qw net.ipv4.ip_forward=1
  done

  $SUDO ip netns exec "$NS_CLIENT" ip route add default via 10.10.0.2
  $SUDO ip netns exec "$NS_WAN"    ip route add 10.30.0.0/30 via 10.20.0.2
  $SUDO ip netns exec "$NS_EDGE"   ip route add 10.10.0.0/30 via 10.20.0.1
  $SUDO ip netns exec "$NS_SRV"    ip route add default via 10.30.0.1

  # initcwnd 50 on both ends: a 16 KiB object is 12 segments, more than the
  # default initial window of 10, so without this the response leaves in two
  # bursts and the small-object probe measures the split, not the link.
  # Applied AFTER addressing, because the connected route must exist first.
  $SUDO ip netns exec "$NS_CLIENT" ip route replace default via 10.10.0.2 initcwnd 50
  $SUDO ip netns exec "$NS_SRV"    ip route replace default via 10.30.0.1 initcwnd 50

  # --- link physics: netem ONLY, on the wan emulator ----------------------
  # No shaping here. netem's job is delay and loss; if a rate were also set
  # here, netem's internal FIFO would become the bottleneck queue and the
  # whole point of the split would be lost.
  $SUDO ip netns exec "$NS_WAN" tc qdisc replace dev w-edge root netem \
    delay "${DELAY_MS}ms" loss "${LOSS_PCT}%"
  # Download propagation: server -> edge -> wan -> client egresses c-wan.
  # Without this second netem, only upload sees delay/loss and a "symmetric"
  # dual-ns run would silently measure an asymmetric link.
  $SUDO ip netns exec "$NS_WAN" tc qdisc replace dev c-wan root netem \
    delay "${DELAY_MS}ms" loss "${LOSS_PCT}%"

  # --- queue policy: CAKE ONLY, on the edge router ------------------------
  # e-srv shapes the upload direction (client -> server): this is the egress
  # that actually carries the bulk flow, because the server now sits behind it.
  # e-wan shapes the download direction. Both are root CAKE with no netem
  # above them.
  $SUDO ip netns exec "$NS_EDGE" tc qdisc replace dev e-srv root cake \
    bandwidth "${RATE_MBIT}mbit" rtt "${RTT_MS}ms" \
    ack-filter nat diffserv4 triple-isolate
  $SUDO ip netns exec "$NS_EDGE" tc qdisc replace dev e-wan root cake \
    bandwidth "${RATE_MBIT}mbit" rtt "${RTT_MS}ms" \
    ack-filter nat diffserv4 triple-isolate

  log "up — netem ${DELAY_MS}ms/${LOSS_PCT}% on w-edge ($NS_WAN); cake ${RATE_MBIT}mbit on e-srv+e-wan ($NS_EDGE)"
  check
}

check() {
  local ok=1
  for ns in "$NS_CLIENT" "$NS_WAN" "$NS_EDGE" "$NS_SRV"; do
    $SUDO ip netns list | grep -qw "$ns" || { echo "[!] namespace $ns missing"; ok=0; }
  done

  local wan_q up_q down_q dl_q
  wan_q=$($SUDO ip netns exec "$NS_WAN" tc qdisc show dev w-edge 2>/dev/null)
  dl_q=$($SUDO ip netns exec "$NS_WAN" tc qdisc show dev c-wan 2>/dev/null)
  up_q=$($SUDO ip netns exec "$NS_EDGE" tc qdisc show dev e-srv 2>/dev/null)
  down_q=$($SUDO ip netns exec "$NS_EDGE" tc qdisc show dev e-wan 2>/dev/null)

  echo "$wan_q" | grep -q netem || { echo "[!] no netem on w-edge (upload physics): $wan_q"; ok=0; }
  echo "$dl_q" | grep -q netem || { echo "[!] no netem on c-wan (download physics): $dl_q"; ok=0; }
  echo "$up_q" | grep -q cake || { echo "[!] no cake on e-srv (upload path unshaped): $up_q"; ok=0; }
  echo "$down_q" | grep -q cake || { echo "[!] no cake on e-wan (download path unshaped): $down_q"; ok=0; }

  if echo "$wan_q" | grep -q cake || echo "$dl_q" | grep -q cake; then
    echo "[!] cake found on a netem interface — stacking bug"; ok=0
  fi
  for q in "$up_q" "$down_q"; do
    if echo "$q" | grep -q netem; then
      echo "[!] netem found on a cake interface — stacking bug: $q"; ok=0
    fi
  done

  # Reachability end to end.
  $SUDO ip netns exec "$NS_CLIENT" ping -c1 -W2 10.30.0.2 >/dev/null 2>&1 \
    || { echo "[!] client cannot reach server"; ok=0; }

  # Routing must traverse the emulated hops, not a local/loopback shortcut.
  $SUDO ip netns exec "$NS_CLIENT" ip route get 10.30.0.2 2>/dev/null | grep -q "via 10.10.0.2 dev c-cli" \
    || { echo "[!] client route leak (not via c-cli/10.10.0.2)"; ok=0; }
  $SUDO ip netns exec "$NS_SRV" ip route get 10.10.0.1 2>/dev/null | grep -q "via 10.30.0.1 dev s-edge" \
    || { echo "[!] server route leak (not via s-edge/10.30.0.1)"; ok=0; }

  check_offloads "$NS_CLIENT" c-cli || ok=0
  check_offloads "$NS_WAN" c-wan || ok=0
  check_offloads "$NS_WAN" w-edge || ok=0
  check_offloads "$NS_EDGE" e-wan || ok=0
  check_offloads "$NS_EDGE" e-srv || ok=0
  check_offloads "$NS_SRV" s-edge || ok=0

  if [ "$ok" = 1 ]; then
    echo "[dualns] ok — link physics and queue policy are isolated"
  fi
  return $((1-ok))
}

show() {
  echo "=== $NS_WAN (w-edge) — upload link physics ==="
  $SUDO ip netns exec "$NS_WAN" tc qdisc show dev w-edge 2>/dev/null
  echo
  echo "=== $NS_WAN (c-wan) — download link physics ==="
  $SUDO ip netns exec "$NS_WAN" tc qdisc show dev c-wan 2>/dev/null
  echo
  echo "=== $NS_EDGE (e-srv) — upload queue policy ==="
  $SUDO ip netns exec "$NS_EDGE" tc qdisc show dev e-srv 2>/dev/null
  echo
  echo "=== $NS_EDGE (e-wan) — download queue policy ==="
  $SUDO ip netns exec "$NS_EDGE" tc qdisc show dev e-wan 2>/dev/null
  echo
  echo "=== routes ==="
  for ns in "$NS_CLIENT" "$NS_WAN" "$NS_EDGE" "$NS_SRV"; do
    echo "--- $ns"
    $SUDO ip netns exec "$ns" ip route show 2>/dev/null
  done
}

CMD="${1:-check}"; shift || true
while [ $# -gt 0 ]; do
  case "$1" in
    --delay) DELAY_MS="$2"; shift 2;;
    --loss)  LOSS_PCT="$2"; shift 2;;
    --rate)  RATE_MBIT="$2"; shift 2;;
    --rtt)   RTT_MS="$2"; shift 2;;
    --offloads) OFFLOADS="$2"; shift 2;;
    *) echo "unknown option: $1" >&2; exit 2;;
  esac
done

case "$CMD" in
  up) up;;
  down) down;;
  check) check;;
  show) show;;
  *) echo "usage: $0 up|down|check|show [--delay MS] [--loss PCT] [--rate MBIT] [--rtt MS] [--offloads on|off]"; exit 2;;
esac
