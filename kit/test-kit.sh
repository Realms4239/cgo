#!/usr/bin/env bash
# kit regression test (plan 2026-08-27-wall-kit-4-kit.md Task 4.1) — build-then-embed strict gate present.
set -u
fail=0
test -f kit/engine.sh || { echo "FAIL kit/engine.sh missing"; fail=1; }
grep -q "TOTAL 600" kit/engine.sh || { echo "FAIL build gate TOTAL 600 absent"; fail=1; }
grep -q "check-bundle" kit/engine.sh || { echo "FAIL check-bundle guard absent"; fail=1; }
[ "$fail" -eq 0 ] && echo "PASS test-kit" || exit 1
