#!/usr/bin/env bash
# v2-matrix-run.sh — lance/reprend une campagne dualns DÉTACHÉE dans l'invité.
#
# Pourquoi ce script existe (sadcnwp 2026-11-16) :
#  1. `ip netns exec ns_client` exige root : le lanceur doit porter `sudo -n`,
#     sinon « setting the network namespace ... Operation not permitted ».
#  2. Une campagne tuée par la fermeture du canal SSH laisse un lock
#     `.run-<id>.lock`. Le supprimer est interdit par le garde-fou ; on le
#     RENOMME en `.stale-<date>` (preuve conservée, écriture seule libérée).
#  3. Le binaire vérifie `ps count == 1` avant chaque lancement : trop de
#     processus `cgo run` simultanés sont refusés, c'est la porte voulue.
#
# Usage (dans l'invité bench, une fois le pont + puits prêts via run-dualns.sh up) :
#   sudo -n bash kit/v2-matrix-run.sh --run-id run-1794853778 \
#     --profiles P1,P2,P3,P4 --cc cubic,bbr --flows 4 --reps 3
set -euo pipefail

RUN_ID=""; PROFILES="P1,P2,P3,P4"; CCS="cubic,bbr"; FLOWS="4"; REPS="3"
QDISCS="cake"; DIRECTION="down"; DEADLINE="220"; DATA="/home/bench/bench-data"
BIN="${CGO_BIN:-/home/bench/cgo}"
LOG="/home/bench/v2-matrix.log"

while [ $# -gt 0 ]; do
  case "$1" in
    --run-id) RUN_ID="${2:?}"; shift 2;;
    --profiles) PROFILES="${2:?}"; shift 2;;
    --cc) CCS="${2:?}"; shift 2;;
    --flows) FLOWS="${2:?}"; shift 2;;
    --reps) REPS="${2:?}"; shift 2;;
    --qdiscs) QDISCS="${2:?}"; shift 2;;
    --direction) DIRECTION="${2:?}"; shift 2;;
    --deadline) DEADLINE="${2:?}"; shift 2;;
    --data) DATA="${2:?}"; shift 2;;
    --bin) BIN="${2:?}"; shift 2;;
    --log) LOG="${2:?}"; shift 2;;
    *) echo "option inconnue: $1" >&2; exit 2;;
  esac
done

[ -n "$RUN_ID" ] || { echo "--run-id requis" >&2; exit 2; }
[ -x "$BIN" ] || { echo "binaire introuvable: $BIN" >&2; exit 2; }

# garde-fou : vous pouvez déjà avoir une campagne en cours. Ne pas en lancer
# une deuxième, même avec un run-id différent.
if pgrep -f 'cgo run --bench dualns' >/dev/null 2>&1; then
  echo "[v2] campagne dualns déjà en cours — abandon" >&2
  exit 3
fi

# lock orphelin (campagne tuée) : RENOMMER, jamais supprimer.
if [ -e "$DATA/.${RUN_ID}.lock" ]; then
  stale="$DATA/.${RUN_ID}.lock.stale-$(date -u +%Y%m%dT%H%M%SZ)"
  mv "$DATA/.${RUN_ID}.lock" "$stale"
  echo "[v2] lock orphelin renommé: $stale"
fi

# DÉTACHÉ : sudo -n + setsid + nohup. Le lanceur rend la main immédiatement,
# la campagne survit à la fermeture du tunnel SSH.
setsid nohup sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1 \
  ip netns exec ns_client "$BIN" run --bench dualns \
  --run-id "$RUN_ID" --profiles "$PROFILES" --qdiscs "$QDISCS" --cc "$CCS" \
  --direction "$DIRECTION" --reps "$REPS" --flows "$FLOWS" \
  --deadline "$DEADLINE" --data "$DATA" \
  > "$LOG" 2>&1 < /dev/null &
pid=$!
echo "$pid" > "${LOG%.log}.pid"
echo "[v2] campagne $RUN_ID lancée (pid $pid), log $LOG"
