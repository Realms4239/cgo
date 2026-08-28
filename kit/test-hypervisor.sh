#!/usr/bin/env bash
# kit regression test (plan 2026-08-27-wall-kit-4-kit.md Task 4.2) — VMware detect-fallback VirtualBox.
set -u
fail=0
grep -q "VBoxManage" kit/engine.sh || { echo "FAIL VBoxManage fallback absent"; fail=1; }
grep -q "vmrun" kit/engine.sh || { echo "FAIL vmrun detect absent"; fail=1; }
[ "$fail" -eq 0 ] && echo "PASS test-hypervisor" || exit 1
