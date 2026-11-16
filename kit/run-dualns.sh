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
#           --ingress on|off (façonnage descendant IFB, opt-in)
#           CGO_SHAPE_MODE=ingress|egress (défaut quand --ingress absent ;
#           le drapeau gagne sur l'env, l'env résolue est ré-exportée pour la
#           campagne et la topologie posée est écrite dans $CGO_TOPOLOGY_JSON)
# ============================================================================
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
BIN="${CGO_BIN:-$HOME/cgo/cgo-linux}"
CMD="${1:-up}"; shift 2>/dev/null || true
INGRESS=""
while [ $# -gt 0 ]; do
  case "$1" in
    --bin) BIN="${2:?--bin requires PATH}"; shift 2;;
    --ingress) INGRESS="${2:?--ingress requires on|off}"; shift 2;;
    *) echo "unknown option: $1" >&2; exit 2;;
  esac
done
case "$INGRESS" in
  ""|on|off) ;;
  *) echo "ingress must be on|off" >&2; exit 2;;
esac
# I1 — mode de façonnage résolu UNE fois, partagé par le banc et la campagne :
# un mode qui diffère entre les deux fait mesurer un chemin non façonné en le
# croyant façonné — défaut de classe fraude, refusé par construction ici.
SHAPE_MODE=""
case "${CGO_SHAPE_MODE:-}" in
  ""|ingress|egress) ;;
  *) echo "[!] CGO_SHAPE_MODE must be ingress|egress (got: ${CGO_SHAPE_MODE})" >&2; exit 2;;
esac
case "$INGRESS" in
  "")   SHAPE_MODE="${CGO_SHAPE_MODE:-}";;
  on)   SHAPE_MODE="ingress";;
  off)  SHAPE_MODE="egress";;
esac
case "$SHAPE_MODE" in
  ingress) BENCH_INGRESS_ARGS="--ingress on";;
  egress)  BENCH_INGRESS_ARGS="--ingress off";;
  *)       BENCH_INGRESS_ARGS="";; # env absente : défaut du banc (egress)
esac
[ -x "$BIN" ] || { echo "[!] binaire introuvable: $BIN (compilez: CGO_ENABLED=0 go build -tags ebpf_object -o $BIN ./cmd/cgo — sans l'étiquette, la télémétrie noyau est silencieusement absente)" >&2; exit 1; }
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if sudo -n true 2>/dev/null; then SUDO="sudo -n"; else echo "[!] root ou sudo NOPASSWD requis" >&2; exit 1; fi
fi

case "$CMD" in
  up|serve)
    if [ -n "$SHAPE_MODE" ]; then
      export CGO_SHAPE_MODE="$SHAPE_MODE"
    fi
    # $BENCH_INGRESS_ARGS non quoté : découpage voulu ("--ingress on|off").
    "$HERE/cgo-bench-dualns.sh" up $BENCH_INGRESS_ARGS || exit 1
    # I1 — topologie réellement bâtie, lue en préflight par CheckBenchTopology
    # (pkg/campagne) : un désaccord ingress/egress refuse la cellule AVANT
    # toute mutation tc. Débit/délai/perte : défauts littéraux du banc
    # (cgo-bench-dualns.sh n'expose pas ces réglages en env — flags seuls).
    TOPO="${CGO_TOPOLOGY_JSON:-/tmp/cgo-dualns-topology.json}"
    IFB_JSON=null
    if [ "$SHAPE_MODE" = "ingress" ]; then IFB_JSON='"ifb0"'; fi
    # Écriture atomique : tmp puis mv (même fs), jamais un JSON à moitié
    # écrit qu'un preflight pourrait lire (M3).
    printf '{"shape_mode":"%s","rate_mbit":20,"delay_ms":100,"loss_pct":0,"namespaces":["ns_client","ns_wan_emulator","ns_edge_router","ns_server"],"ifb":%s,"written_at":"%s"}\n' \
      "${SHAPE_MODE:-egress}" "$IFB_JSON" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$TOPO.tmp"
    mv "$TOPO.tmp" "$TOPO"
    # puits : petit objet + masse, DANS ns_server — jamais dans le routeur
    # (un puits à l'edge court-circuiterait le shaper : 193 Mbit/s à travers
    # un plafond de 20, vu en prod).
    # obj16.bin épinglé par kit/testbed.sh / cgo-netctl (16 KiB urandom, root) :
    # sans lui le puits sert HTTP 500 sur /small et la campagne mesure du vide.
    $SUDO ip netns exec ns_server sh -c "CGO_TESTBED_OBJ=/var/lib/cgo-testbed/obj16.bin nohup '$BIN' testbedsrv -http 10.30.0.2:8081 -bulk 10.30.0.2:5201 >/tmp/testbedsrv-dualns.log 2>&1 &"
    sleep 1
    $SUDO ip netns exec ns_client ping -c1 -W2 10.30.0.2 >/dev/null \
      || { echo "[!] puits injoignable depuis ns_client — voir /tmp/testbedsrv-dualns.log"; exit 1; }
    # /small prouve l'objet 16 KiB épinglé : sans lui le puits répond 500 et
    # la campagne mesure du vide (G1/G2/G4 rouges — vu sur sadcnwp 2026-11-16).
    $SUDO ip netns exec ns_client curl -fsS -o /dev/null http://10.30.0.2:8081/small
    if [ "$CMD" = "up" ]; then
      echo "[run-dualns] banc + puits prêts — dashboard: $0 serve · campagne: $BIN run --bench dualns"
      exit 0
    fi
    # CGO_DASHBOARD__ADDR + HTTP brut : la veth de gestion ne porte AUCUN
    # trafic mesuré, et meteolink.dev/127.0.0.1 n'existe pas dans ns_client.
    echo "[run-dualns] dashboard → http://10.99.0.2:9090 (navigateur hôte)"
    exec $SUDO ip netns exec ns_client env CGO_DASHBOARD__ADDR=10.99.0.2:9090 \
      CGO_IFACE="${CGO_IFACE:-c-cli}" \
      CGO_IFACE_SOURCE="${CGO_IFACE_SOURCE:-twin}" \
      "$BIN" --serve --tls=false --http-addr=
    ;;
  down)
    # T1 — le dashboard (serve) et le puits survivent au `bench down` :
    # tués explicitement. Marqueur `--serve .*--http-addr=` (option HTTP
    # vide du serveur embarqué) pour ne jamais viser un autre process cgo.
    # le puits et le dashboard sont lancés par `up` sous sudo (ip netns exec
    # via root) — le pkill doit être aussi sous sudo, sinon il ne voit ni ne
    # tue ces processus (pgrep voit tout, kill non) : le down échouait fermé
    # avec « orphelins survivants » à chaque run.
    $SUDO pkill -f '.*--serve .*--http-addr=' 2>/dev/null || true
    $SUDO pkill -f "testbedsrv -http 10.30.0.2" 2>/dev/null || true
    # Échec du bench down ne doit PAS court-circuiter la preuve 0-orphan :
    # rc capturé, vérification exécutée, verdict à la fin (C1).
    bench_rc=0
    "$HERE/cgo-bench-dualns.sh" down || bench_rc=$?
    # Échec fermé : affirmer un teardown propre sans preuve = faux succès.
    orphans=""
    for ns in ns_client ns_wan_emulator ns_edge_router ns_server; do
      if ip netns list 2>/dev/null | grep -qw "$ns"; then
        orphans="$orphans netns:$ns"
      fi
    done
    if pgrep -f "testbedsrv -http 10.30.0.2" >/dev/null 2>&1; then
      orphans="$orphans proc:testbedsrv"
    fi
    if pgrep -f '.*--serve .*--http-addr=' >/dev/null 2>&1; then
      orphans="$orphans proc:dashboard"
    fi
    # Verdict unique : orphelins OU bench down en échec ⇒ pas de « 0 orphelin ».
    if [ -n "$orphans" ]; then
      echo "[!] orphelins survivants au down :$orphans" >&2
      exit 1
    fi
    if [ "$bench_rc" -ne 0 ]; then
      echo "[!] bench down en échec (rc=$bench_rc) — teardown non prouvé" >&2
      exit 1
    fi
    echo "[run-dualns] down — 0 orphelin (netns, puits, dashboard)"
    ;;
  *)
    echo "usage: $0 up|serve|down [--bin PATH] [--ingress on|off]" >&2; exit 2;;
esac
