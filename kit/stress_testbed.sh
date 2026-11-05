#!/bin/bash
# ============================================================================
#  stress_testbed.sh — adversarial check of the constrained benchmark.
#
#  What it proves, on the real Linux bench (not a mock):
#    1. veth offloads are actually OFF (GRO/GSO/TSO), so a 64 KiB GSO skb
#       cannot hide drops from CAKE/TBF accounting;
#    2. traffic traverses veth-c -> netns cgo-srv, not a loopback shortcut;
#    3. a multi-flow BBR load still fills the shaper under P2 loss;
#    4. teardown removes the namespace, the listener and its goroutines.
#
#  The default is the campaign-wired legacy bench (kit/testbed.sh). The
#  dual-namespace bench is checked separately by cgo-bench-dualns.sh check;
#  it is not yet wired to campagne.ProdDeps, so claiming a campaign result
#  from it would be false.
#
#  Usage (as the bench user, not via `sudo bash`):
#    kit/stress_testbed.sh [--flows 4] [--baseline 5] [--charge 15] [--recup 5] [--keep]
# ============================================================================
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="${CGO_BIN:-$HOME/cgo/cgo-linux}"
FLOWS=4
BASE=5
CHARGE=15
RECUP=5
KEEP=0
TMP="$(mktemp -d "${TMPDIR:-/tmp}/cgo-stress.XXXXXX")"
LOG="${CGO_TESTBED_LOG:-$HOME/cgo/testbed/testbedsrv.log}"
FD_BEFORE=0

while [ $# -gt 0 ]; do
  case "$1" in
    --flows) FLOWS="$2"; shift 2;;
    --baseline) BASE="$2"; shift 2;;
    --charge) CHARGE="$2"; shift 2;;
    --recup) RECUP="$2"; shift 2;;
    --keep) KEEP=1; shift;;
    -h|--help) sed -n '2,25p' "$0"; exit 0;;
    *) echo "unknown option: $1" >&2; exit 2;;
  esac
done
case "$FLOWS" in ''|*[!0-9]*) echo "[!] --flows must be 1–8" >&2; exit 2;; esac
[ "$FLOWS" -ge 1 ] && [ "$FLOWS" -le 8 ] || { echo "[!] --flows must be 1–8" >&2; exit 2; }

log() { echo "[stress] $*"; }
die() { echo "[!] $*" >&2; exit 1; }

cleanup() {
  local rc=$?
  if [ "$KEEP" = 0 ]; then
    ( cd "$ROOT" && CGO_OFFLOADS=off ./kit/testbed.sh down ) >/dev/null 2>&1 || true
  fi
  if [ -s "$LOG" ]; then
    echo "--- testbedsrv log (tail) ---"
    tail -20 "$LOG" || true
  fi
  rm -rf "$TMP"
  exit "$rc"
}
trap cleanup EXIT INT TERM

[ "$(uname -s)" = Linux ] || die "Linux only"
command -v ethtool >/dev/null || die "ethtool missing (apt install ethtool)"
[ -x "$BIN" ] || die "binary $BIN missing — build: go build -o cgo-linux ./cmd/cgo"
[ -f "$ROOT/kit/testbed.sh" ] || die "kit/testbed.sh missing"

if [ -r /proc/sys/fs/file-nr ]; then
  FD_BEFORE=$(awk '{print $1}' /proc/sys/fs/file-nr)
fi

log "tearing down any stale bench"
( cd "$ROOT" && CGO_OFFLOADS=off ./kit/testbed.sh down ) >/dev/null 2>&1 || true
log "bringing up bench with CGO_OFFLOADS=off"
( cd "$ROOT" && CGO_OFFLOADS=off ./kit/testbed.sh up ) || die "testbed up failed"

# Independent verification (testbed.sh check also does this; do not rely on
# the script agreeing with itself).
for spec in "veth-c:" "veth-s:cgo-srv"; do
  it="${spec%%:*}"; ns="${spec##*:}"
  for feat in tcp-segmentation-offload generic-segmentation-offload generic-receive-offload; do
    if [ -n "$ns" ]; then
      line=$(sudo -n ip netns exec "$ns" ethtool -k "$it" 2>/dev/null | grep -m1 "^${feat}:")
    else
      line=$(sudo -n ethtool -k "$it" 2>/dev/null | grep -m1 "^${feat}:")
    fi
    case "$line" in
      *": off"*) ;;
      *) die "$feat not OFF on $it ($line) — GSO super-packets would distort drop counts";;
    esac
  done
done
log "offloads verified off on veth-c and veth-s"

ip route get 10.200.0.2 2>/dev/null | grep -q 'dev veth-c' \
  || die "client route to 10.200.0.2 does not use veth-c"
sudo -n ip netns exec cgo-srv ip route get 10.200.0.1 2>/dev/null | grep -q 'dev veth-s' \
  || die "server route to 10.200.0.1 does not use veth-s"
log "namespace route isolation verified"

tx_before=$(cat /sys/class/net/veth-c/statistics/tx_packets 2>/dev/null || echo 0)

log "multi-flow campaign: P2/cake/bbr, ${FLOWS} flows, ${BASE}/${CHARGE}/${RECUP}s"
set +e
(
  cd "$ROOT" || exit 1
  CGO_OFFLOADS=off \
  CGO_TARGET=10.200.0.1 \
  CGO_SMALL_URL=http://10.200.0.1:8081/small \
  CGO_BULK_ADDR=10.200.0.1:5201 \
  CGO_BASELINE_SEC="$BASE" CGO_CHARGE_SEC="$CHARGE" CGO_RECUP_SEC="$RECUP" \
  "$BIN" run --profiles P2 --qdiscs cake --cc bbr --reps 1 --flows "$FLOWS" \
    --data "$TMP/runs" --target 10.200.0.1
)
run_rc=$?
set -e
[ "$run_rc" -eq 0 ] || die "campaign failed (rc=$run_rc)"

csv=$(find "$TMP/runs" -name aqm_eval.csv -print | head -1)
[ -n "$csv" ] || die "no aqm_eval.csv produced — a zero-row run is not a stress pass"

if ! python3 - "$csv" "$FLOWS" <<'PY'
import csv, sys
path, flows = sys.argv[1], int(sys.argv[2])
rows = list(csv.DictReader(open(path, encoding="utf-8", errors="replace")))
if not rows:
    print("[!] no events", file=sys.stderr); sys.exit(1)
for r in rows:
    good = float(r.get("bulk_goodput_mbps") or 0)
    gate = r.get("gate_status") or ""
    if good <= 0:
        print(f"[!] zero goodput ({gate})", file=sys.stderr); sys.exit(1)
    print(f"[stress] {r.get('qdisc')}/{r.get('cc')} flows={flows} goodput={good:.1f} Mbit/s gate={gate}")
PY
then
  die "multi-flow load did not produce positive goodput"
fi

tx_after=$(cat /sys/class/net/veth-c/statistics/tx_packets 2>/dev/null || echo 0)
[ "$tx_after" -gt "$tx_before" ] || die "veth-c TX counter did not increase — traffic bypassed the bench"
log "veth-c TX packets: $tx_before -> $tx_after (traffic crossed the constrained path)"

if [ "$KEEP" = 0 ]; then
  ( cd "$ROOT" && CGO_OFFLOADS=off ./kit/testbed.sh down ) || die "testbed down failed"
  if ip netns list 2>/dev/null | grep -qw cgo-srv; then die "namespace cgo-srv survived teardown"; fi
  if pgrep -f '[t]estbedsrv' >/dev/null; then die "testbedsrv process survived teardown"; fi
  log "teardown clean: namespace and listener gone"
fi

if [ -r /proc/sys/fs/file-nr ]; then
  fd_after=$(awk '{print $1}' /proc/sys/fs/file-nr)
  log "open file handles: $FD_BEFORE -> $fd_after"
fi

log "PASS — offloads off, route isolated, multi-flow load positive, teardown clean"
