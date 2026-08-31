#!/usr/bin/env bash
# Test du kit — le moteur Go porte les portes historiques du bash.
set -u
fail=0
test -f internal/kit/kit.go || { echo "FAIL internal/kit manquant"; fail=1; }
grep -q "check-bundle" internal/kit/kit.go || { echo "FAIL check-bundle guard absent"; fail=1; }
grep -q "vitest" internal/kit/kit.go || { echo "FAIL vitest gate absent"; fail=1; }
test -f internal/vm/vm.go || { echo "FAIL internal/vm manquant"; fail=1; }
[ "$fail" -eq 0 ] && echo "PASS test-kit" || exit 1
