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
  # disciplines du banc : l'autoload échoue parfois dans le netns (vu en prod :
  # "Invalid qdisc name" sur la 1re cellule cake après boot) — preload explicite.
  sudo -n modprobe sch_netem sch_fq_codel sch_cake 2>/dev/null || true
  if ! sudo ip netns list | grep -q "^$NS"; then sudo ip netns add $NS; fi
  # paire veth : la créer si absente, RÉPARER si à moitié présente (reboot :
  # veth-c existe sans veth-s dans le ns — l'ancien test tout-ou-rien
  # laissait le banc cassé sans un mot).
  if ! sudo ip link show $CLI &>/dev/null; then
    # veth-s orphelin dans la racine ? le supprimer avant de recréer la paire
    if sudo ip link show $SRV &>/dev/null; then sudo ip link del $SRV 2>/dev/null || true; fi
    sudo ip link add $CLI type veth peer name $SRV
  fi
  if ! sudo ip netns exec $NS ip link show $SRV &>/dev/null; then
    # veth-s coincé dans la racine : le rentrer dans le ns (down d'abord)
    sudo ip link set $SRV down 2>/dev/null || true
    sudo ip link set $SRV netns $NS
  fi
  # adresses idempotentes : n'ajouter que si absentes (l'ajout double
  # échoue et laissait le script croire le banc prêt)
  sudo ip addr show dev $CLI | grep -q "${CIP%/*}" || sudo ip addr add $CIP dev $CLI
  sudo ip link set $CLI up
  sudo ip netns exec $NS ip addr show dev $SRV | grep -q "${CIDR%/*}" || sudo ip netns exec $NS ip addr add $CIDR dev $SRV
  sudo ip netns exec $NS ip link set $SRV up
  sudo ip netns exec $NS ip link set lo up
  mkdir -p "$TBD"
  [ -f "$TBD/obj16.bin" ] || dd if=/dev/urandom of="$TBD/obj16.bin" bs=16384 count=1 status=none
  sudo rm -f "$TBD/testbedsrv.log" /tmp/cgo-testbedsrv.log 2>/dev/null || true
  sudo ip netns exec $NS pkill -f '[t]estbedsrv' 2>/dev/null || true
  sleep 1
  sudo ip netns exec $NS bash -c "CGO_TESTBED_OBJ=\"$TBD/obj16.bin\" nohup \"$BIN\" testbedsrv --http 10.200.0.1:8081 --bulk 10.200.0.1:5201 >\"$TBD/testbedsrv.log\" 2>&1 &"
  sleep 1
  check
}

down() {
  sudo ip netns exec $NS pkill -f '[t]estbedsrv' 2>/dev/null || true
  sleep 1
  sudo ip link del $CLI 2>/dev/null || true
  sudo ip netns del $NS 2>/dev/null || true
  sudo rm -f "$TBD/testbedsrv.log" /tmp/cgo-testbedsrv.log 2>/dev/null || true
  echo "[testbed] down"
}

check() {
  local ok=1
  sudo ip netns list | grep -q "^$NS" || { echo "[!] netns $NS missing"; ok=0; }
  sudo ip link show $CLI &>/dev/null || { echo "[!] $CLI missing"; ok=0; }
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
