#!/usr/bin/env bash
# seal-v2-run.sh — assemble the seal manifest for a completed v2 run.
# Reads only; never writes book/data/runs (the frozen corpus). The operator still has to
# copy the guest artifacts into docs/verification/2026-11-16-v2-corpus/ via the
# safe channel and commit them explicitly.
set -euo pipefail
dir="${1:?usage: seal-v2-run.sh RUN_DIR}"
csv="$dir/aqm_eval.csv"
manifest="$dir/manifest.json"
[ -f "$csv" ] && [ -f "$manifest" ] || { echo "[seal] CSV/manifest missing" >&2; exit 1; }
rows=$(awk 'NR>1 && NF>0' "$csv" | wc -l)
[ "$rows" -eq 24 ] || { echo "[seal] want 24 cells, got $rows" >&2; exit 1; }
# hash8 = SHA-256 of sorted per-run CSV hashes (house convention)
csv_sha=$(sha256sum "$csv" | awk '{print $1}')
m_sha=$(sha256sum "$manifest" | awk '{print $1}')
q="$dir/quarantine.json"
if [ -f "$q" ]; then
  q_sha=$(sha256sum "$q" | awk '{print $1}')
  q_note="present"
else
  q_sha="-"
  q_note="absent — no invalid cell sealed"
fi
printf 'run: %s\ncells: %s\naqm_eval_sha256: %s\nmanifest_sha256: %s\nquarantine: %s (%s)\nhash8: %s\n' \
  "$(basename "$dir")" "$rows" "$csv_sha" "$m_sha" "$q_note" "$q_sha" "${csv_sha:0:8}"
