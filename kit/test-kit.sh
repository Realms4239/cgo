#!/usr/bin/env bash
# Test du kit — le moteur Go porte les portes historiques du bash.
set -u
fail=0
test -f internal/kit/kit.go || { echo "FAIL internal/kit manquant"; fail=1; }
grep -q "check-bundle" internal/kit/kit.go || { echo "FAIL check-bundle guard absent"; fail=1; }
grep -q "vitest" internal/kit/kit.go || { echo "FAIL vitest gate absent"; fail=1; }
test -f internal/vm/vm.go || { echo "FAIL internal/vm manquant"; fail=1; }
[ "$fail" -eq 0 ] && echo "PASS test-kit" || exit 1

# V2 bench robustness (sadcnwp 2026-11-16) : le puits vit dans ns_server.
# Sans pkill AVANT ip netns del, l'ancien puits retient le pin et le prochain
# `up` échoue « Cannot create namespace file ... File exists ». Sans le check
# /small, la campagne enregistre une cellule « valid » qui mesure du vide.
grep -q "pkill -f 'testbedsrv -http 10.30.0.2'" kit/cgo-bench-dualns.sh || { echo "FAIL puits not killed before netns del"; fail=1; }
grep -q 'curl -fsS -o /dev/null http://10.30.0.2:8081/small' kit/run-dualns.sh || { echo "FAIL puits /small preflight absent"; fail=1; }
grep -q 'CGO_TESTBED_OBJ=/var/lib/cgo-testbed/obj16.bin' kit/run-dualns.sh || { echo "FAIL obj16 pin absent from run-dualns"; fail=1; }

# V2 launcher : sudo (netns root), renommage du lock (jamais rm), détachement.
grep -q 'sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1' kit/v2-matrix-run.sh || { echo "FAIL v2 launcher missing sudo"; fail=1; }
grep -q 'mv "$DATA/.${RUN_ID}.lock"' kit/v2-matrix-run.sh || { echo "FAIL v2 stale-lock rename absent"; fail=1; }
grep -q 'pgrep -f .cgo run --bench dualns.' kit/v2-matrix-run.sh || { echo "FAIL v2 single-writer guard absent"; fail=1; }
grep -q 'setsid nohup' kit/v2-matrix-run.sh || { echo "FAIL v2 launcher not detached"; fail=1; }

# V2 evidence checker : manifeste requis, 4 profils, statut de porte explicite.
grep -q 'manifest' kit/check-v2-run.sh || { echo "FAIL v2 checker missing manifest guard"; fail=1; }
grep -q 'valid|degraded|invalid' kit/check-v2-run.sh || { echo "FAIL v2 checker gate vocabulary absent"; fail=1; }
grep -q 'P1 P2 P3 P4' kit/check-v2-run.sh || { echo "FAIL v2 checker profiles absent"; fail=1; }

# V2 seal : exact cell count + house hash8 convention, no frozen-corpus write.
grep -q 'rows.*-eq 24' kit/seal-v2-run.sh || { echo "FAIL v2 seal cell-count guard absent"; fail=1; }
grep -q 'hash8' kit/seal-v2-run.sh || { echo "FAIL v2 seal hash8 absent"; fail=1; }
grep -q 'book/data/runs' kit/seal-v2-run.sh || { echo "FAIL v2 seal frozen-corpus boundary note absent"; fail=1; }
