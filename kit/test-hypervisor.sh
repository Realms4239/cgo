#!/usr/bin/env bash
# Test du kit — détection VMware avec repli VirtualBox (moteur Go).
set -u
fail=0
grep -q "virtualbox" internal/vm/vm.go || { echo "FAIL VBoxManage fallback absent"; fail=1; }
grep -q "vmware" internal/vm/vm.go || { echo "FAIL vmrun detect absent"; fail=1; }
grep -q "powershell" internal/vm/vm.go && { echo "FAIL powershell encore présent (doit être pur Go)"; fail=1; }
[ "$fail" -eq 0 ] && echo "PASS test-hypervisor" || exit 1
