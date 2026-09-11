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
  # sudoers : le banc (veth/netns via sudo -n ci-dessous) ET les cellules
  # de campagne (sudo tc / sudo ip non-interactifs) exigent le sudo sans
  # mot de passe — vu en prod : VM fraîche = chaque cellule échoue et la
  # campagne gèle ZÉRO ligne (« Démarrer → idle → rien »). Une fois pour
  # toutes ici (console : mot de passe demandé une fois, jamais après).
  # SUDO_USER quand lancé via `sudo bash` (sinon $USER=root = règle inutile).
  ME2="${SUDO_USER:-$USER}"
  # sudo 1.9.15p5 (Ubuntu 24.04) sort en 0 MÊME en refusant -n (prouvé
  # live : « a password is required », RC=0) — sonder la SORTIE, jamais
  # le code, sinon la porte est du théâtre. Règle déjà bonne = on ne
  # touche à rien (réécrire exigerait un mot de passe pour rien).
  if sudo -n ip link show lo 2>&1 | grep -qi "password"; then
    echo "[testbed] sudo sans mot de passe requis — configuration (mot de passe demandé une fois)…"
    sudo true || { echo "[!] sudo refusé"; return 1; }
    printf '%s ALL=(ALL) NOPASSWD: /usr/sbin/tc, /sbin/tc, /usr/sbin/ip, /sbin/ip, /usr/sbin/modprobe, /sbin/modprobe, /usr/sbin/ethtool, /sbin/ethtool\n' "$ME2" | sudo tee /etc/sudoers.d/cgo-network >/dev/null
    sudo chmod 0440 /etc/sudoers.d/cgo-network
    sudo visudo -c 2>/dev/null | grep -q "OK" || { echo "[!] sudoers invalide — nettoyez /etc/sudoers.d/cgo-network"; return 1; }
    sudo -n ip link show lo 2>&1 | grep -qi "password" && { echo "[!] sudo -n toujours refusé après écriture (ouvrez une autre session ?)"; return 1; }
  fi
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
  # RECETTE DES RUNS PARFAITS (prouvée camp-fin3 2026-09-11, 9/9 BBR valid) :
  # 1. qdiscs NETTOYÉS — un netem en couches/corrompu (refcnt 129, laissé par
  #    un run tué ou un double-lancement) étrangle le flux à ~1-3 Mb/s ;
  #    détruire AVANT toute campagne rend le banc à neuf.
  sudo -n tc qdisc del dev $CLI root 2>/dev/null || true
  # 2. initcwnd 50 DES DEUX CÔTÉS — l'objet 16 KiB = 12 paquets > initcwnd 10
  #    par défaut : la réponse part en 2 vols → sonde small = 3×RTT (~315 ms
  #    au lieu de ~210). Volatile (perdu au reboot) → réappliqué à chaque up.
  sudo -n ip route change 10.200.0.0/24 dev $CLI initcwnd 50 2>/dev/null || sudo -n ip route add 10.200.0.0/24 dev $CLI initcwnd 50 2>/dev/null || true
  sudo -n ip netns exec $NS ip route change 10.200.0.0/24 dev $SRV initcwnd 50 2>/dev/null || sudo -n ip netns exec $NS ip route add 10.200.0.0/24 dev $SRV initcwnd 50 2>/dev/null || true
  # 3. offloads veth ON des deux côtés — état de la bonne ère (final220,
  #    BBR 18-19 Mb/s) ; réglage volatil → réappliqué à chaque up.
  sudo -n ethtool -K $CLI tso on gso on gro on 2>/dev/null || true
  sudo -n ip netns exec $NS ethtool -K $SRV tso on gso on gro on 2>/dev/null || true
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
  # banc sain = état de la recette : TSO ON (état de la bonne ère) + route
  # initcwnd 50 présente des deux côtés + aucun qdisc résiduel sur veth-c.
  for itf in "$CLI:root" "$SRV:$NS"; do
    it="${itf%%:*}"; ns="${itf##*:}"
    if [ "$ns" = root ]; then o=$(ethtool -k $CLI 2>/dev/null | grep -m1 '^tcp-segmentation-offload'); else o=$(sudo ip netns exec $NS ethtool -k $SRV 2>/dev/null | grep -m1 '^tcp-segmentation-offload'); fi
    [ "$o" = "tcp-segmentation-offload: on" ] || { echo "[!] TSO off on $it ($o)"; ok=0; }
  done
  ip route show | grep -q "10.200.0.0/24 dev $CLI scope link initcwnd 50" || { echo "[!] initcwnd 50 route missing (client)"; ok=0; }
  sudo ip netns exec $NS ip route show | grep -q "10.200.0.0/24 dev $SRV scope link initcwnd 50" || { echo "[!] initcwnd 50 route missing (server)"; ok=0; }
  tc qdisc show dev $CLI | grep -q noqueue || { echo "[!] stale qdisc on $CLI — run: testbed.sh up"; ok=0; }
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
