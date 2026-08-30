#!/usr/bin/env bash
# Test du kit — détection VMware avec repli VirtualBox.
set -u
fail=0
grep -q "VBoxManage" kit/engine.sh || { echo "FAIL VBoxManage fallback absent"; fail=1; }
grep -q "vmrun" kit/engine.sh || { echo "FAIL vmrun detect absent"; fail=1; }
[ "$fail" -eq 0 ] && echo "PASS test-hypervisor" || exit 1
