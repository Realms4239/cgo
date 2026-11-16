#!/usr/bin/env bash
# check-v2-run.sh — vérifie les engagements d'une campagne v2 scellée, sans
# rien effacer ni modifier. Le corpus gelé (book/data/runs) n'est jamais lu.
#
#   bash kit/check-v2-run.sh <run-dir> [expected-cells]
#   ex: bash kit/check-v2-run.sh /home/bench/bench-data/run-1794853778 24
set -euo pipefail
dir="${1:?usage: check-v2-run.sh RUN_DIR [EXPECTED_CELLS]}"
expected="${2:-24}"
csv="$dir/aqm_eval.csv"
manifest="$dir/manifest.json"
[ -f "$csv" ] || { echo "[v2] CSV absent: $csv" >&2; exit 1; }

rows=$(awk 'NR>1 && NF>0' "$csv" | wc -l)
if [ "$rows" -ne "$expected" ]; then
  echo "[v2] ATTENDU $expected cellules, VU $rows" >&2
  exit 1
fi
if [ ! -f "$manifest" ]; then
  echo "[v2] manifeste absent (campagne pas encore scellée): $manifest" >&2
  exit 2
fi
# toute ligne porte un statut de porte explicite, jamais vide
awk -F, 'NR>1 && $NF !~ /^(valid|degraded|invalid)$/ { print "[v2] gate_status invalide ligne " NR ": " $NF > "/dev/stderr"; bad=1 } END { exit bad }' "$csv"
# les trois profils de la matrice reduced doivent être représentés
for p in P1 P2 P3 P4; do
  awk -F, -v p="$p" 'NR>1 && $3==p {n++} END {exit !(n>0)}' "$csv" \
    || { echo "[v2] profil $p absent" >&2; exit 1; }
done
echo "[v2] OK — $rows cellules, manifeste présent, portes explicites, 4 profils"
