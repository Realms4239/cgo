#!/bin/bash
# ============================================================================
#  run-dualns.sh — le chemin de production en une commande.
#
#  banc durci (netem | CAKE découplés, offloads fermés-échec)
#    + puits dans ns_server (10.30.0.2 — les QUATRE sauts sur le chemin)
#    + tableau de bord DANS ns_client, servi sur la veth de gestion
#      (10.99.0.2:9090, HTTP brut — le navigateur hôte joint 10.99.0.2).
#
#  Usage:
#    run-dualns.sh up              banc + puits prêts (campagne CLI ensuite)
#    run-dualns.sh serve           up + dashboard dans ns_client (avant-plan)
#    run-dualns.sh down            tout retirer
#  Options: --bin PATH (défaut $CGO_BIN ou ~/cgo/cgo-linux)
# ============================================================================
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
BIN="${CGO_BIN:-$HOME/cgo/cgo-linux}"
CMD="${1:-up}"; shift 2>/dev/null || true
while [ $# -gt 0 ]; do
  case "$1" in
    --bin) BIN="$2"; shift 2;;
    *) echo "unknown option: $1" >&2; exit 2;;
  esac
done
[ -x "$BIN" ] || { echo "[!] binaire introuvable: $BIN (compilez: go build -o $BIN ./cmd/cgo)" >&2; exit 1; }
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if sudo -n true 2>/dev/null; then SUDO="sudo -n"; else echo "[!] root ou sudo NOPASSWD requis" >&2; exit 1; fi
fi

case "$CMD" in
  up|serve)
    "$HERE/cgo-bench-dualns.sh" up || exit 1
    # puits : petit objet + masse, DANS ns_server — jamais dans le routeur
    # (un puits à l'edge court-circuiterait le shaper : 193 Mbit/s à travers
    # un plafond de 20, vu en prod).
    $SUDO ip netns exec ns_server sh -c "nohup '$BIN' testbedsrv -http 10.30.0.2:8081 -bulk 10.30.0.2:5201 >/tmp/testbedsrv-dualns.log 2>&1 &"
    sleep 1
    $SUDO ip netns exec ns_client ping -c1 -W2 10.30.0.2 >/dev/null \
      || { echo "[!] puits injoignable depuis ns_client — voir /tmp/testbedsrv-dualns.log"; exit 1; }
    if [ "$CMD" = "up" ]; then
      echo "[run-dualns] banc + puits prêts — dashboard: $0 serve · campagne: $BIN run --bench dualns"
      exit 0
    fi
    # CGO_DASHBOARD__ADDR + HTTP brut : la veth de gestion ne porte AUCUN
    # trafic mesuré, et meteolink.dev/127.0.0.1 n'existe pas dans ns_client.
    echo "[run-dualns] dashboard → http://10.99.0.2:9090 (navigateur hôte)"
    exec $SUDO ip netns exec ns_client env CGO_DASHBOARD__ADDR=10.99.0.2:9090 \
      "$BIN" --serve --tls=false --http-addr=
    ;;
  down)
    pkill -f "testbedsrv -http 10.30.0.2" 2>/dev/null || true
    "$HERE/cgo-bench-dualns.sh" down
    echo "[run-dualns] down"
    ;;
  *)
    echo "usage: $0 up|serve|down [--bin PATH]" >&2; exit 2;;
esac
