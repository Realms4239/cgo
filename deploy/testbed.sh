#!/bin/bash
# ================================================================
#  CGO testbed provisioner — runs ON THE VM (LIEN III.III banc).
#  Topology: main-ns veth-c (netem hop) ↔ netns cgo-srv veth-s
#  (shaping+aqm hop). One `cgo testbedsrv` inside the ns serves the
#  16 KiB small object (HTTP :8081) and the bulk sink (:5201).
#  Idempotent. Usage: testbed.sh up|down|check
# ================================================================
set -uo pipefail
NS=cgo-srv
CLI=veth-c; SRV=veth-s
CIDR=10.200.0.1/24; CIP=10.200.0.2/24
BIN="$HOME/cgo/cgo-linux"
TBD="$HOME/cgo/testbed"

up() {
  if ! ip netns list | grep -q "^$NS"; then ip netns add $NS; fi
  if ! ip link show $CLI &>/dev/null; then
    ip link add $CLI type veth peer name $SRV
    ip link set $SRV netns $NS
    ip addr add $CIP dev $CLI
    ip link set $CLI up
    ip netns exec $NS ip addr add $CIDR dev $SRV
    ip netns exec $NS ip link set $SRV up
    ip netns exec $NS ip link set lo up
  fi
  mkdir -p "$TBD"
  [ -f "$TBD/obj16.bin" ] || dd if=/dev/urandom of="$TBD/obj16.bin" bs=16384 count=1 status=none
  if ! ip netns exec $NS pgrep -f 'testbedsrv' >/dev/null; then
    ip netns exec $NS env CGO_TESTBED_OBJ="$TBD/obj16.bin" \
      nohup "$BIN" testbedsrv --http 10.200.0.1:8081 --bulk 10.200.0.1:5201 \
      >/tmp/cgo-testbedsrv.log 2>&1 &
  fi
  sleep 1
  check
}

down() {
  ip netns pids $NS 2>/dev/null | xargs -r kill 2>/dev/null || true
  ip link del $CLI 2>/dev/null || true
  ip netns del $NS 2>/dev/null || true
  echo "[testbed] down"
}

check() {
  local ok=1
  ip netns list | grep -q "^$NS" || { echo "[!] netns $NS missing"; ok=0; }
  ip link show $CLI &>/dev/null || { echo "[!] $CLI missing"; ok=0; }
  ping -c1 -W1 10.200.0.1 >/dev/null 2>&1 || { echo "[!] ping 10.200.0.1 fails"; ok=0; }
  curl -fsS -m2 http://10.200.0.1:8081/small -o /dev/null || { echo "[!] small object fails"; ok=0; }
  timeout 1 bash -c '</dev/tcp/10.200.0.1/5201' 2>/dev/null || { echo "[!] bulk sink closed"; ok=0; }
  [ $ok = 1 ] && echo "[testbed] ok — constrained path ready"
  return $((1-ok))
}

case "${1:-check}" in
  up) up;; down) down;; check) check;;
  *) echo "usage: testbed.sh up|down|check"; exit 2;;
esac
