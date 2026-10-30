#!/usr/bin/env bash
# Diagnostic lecture seule d'un hôte de banc. Ne modifie rien, n'exige pas
# root : il dit ce qui manque et sort en 1 si une ligne requise manque.
# Windows : les lignes Linux sont n/a, le script reste utilisable.
set -uo pipefail
require_linux=0
[ "${1:-}" = "--require-linux" ] && require_linux=1
os="$(uname -s)"
if [ "$require_linux" = 1 ] && [ "$os" != "Linux" ]; then
  echo "hôte non-Linux ($os) — l'acceptation live exige Linux 6.8" >&2
  exit 2
fi
missing=0
row() { printf '%-22s %-8s %s\n' "$1" "$2" "$3"; }
need() {
  local name="$1" status="$2" detail="$3"
  row "$name" "$status" "$detail"
  if [ "$status" = "MISSING" ] && [ "$os" = "Linux" ]; then missing=1; fi
}
if [ "$os" != "Linux" ]; then
  for c in "noyau >= 6.8" "BTF vmlinux" "CAP_BPF/PERFMON" "ip/tc/ethtool" "bpftool+clang" "KVM" "qemu" "image CHR"; do
    row "$c" "n/a" "hôte non-Linux"
  done
  row "bench dualns" "OK" "$( [ -f kit/cgo-bench-dualns.sh ] && echo présent || echo ABSENT )"
  echo
  echo "résumé : hôte non-Linux, acceptation live impossible ici"
  exit 0
fi
kver="$(uname -r)"; major="${kver%%.*}"; minor="$(echo "$kver" | cut -d. -f2)"
if [ "$major" -gt 6 ] || { [ "$major" -eq 6 ] && [ "$minor" -ge 8 ]; }; then
  need "noyau >= 6.8" OK "$kver"
else
  need "noyau >= 6.8" MISSING "$kver"
fi
[ -f /sys/kernel/btf/vmlinux ] && need "BTF vmlinux" OK présent || need "BTF vmlinux" MISSING absent
if [ "$(id -u)" = 0 ]; then need "CAP_BPF/PERFMON" OK root; else need "CAP_BPF/PERFMON" MISSING "root requis (approximation)"; fi
ok=1; for t in ip tc ethtool; do command -v "$t" >/dev/null || ok=0; done
[ "$ok" = 1 ] && need "ip/tc/ethtool" OK présents || need "ip/tc/ethtool" MISSING manquant
ok=1; for t in bpftool clang; do command -v "$t" >/dev/null || ok=0; done
[ "$ok" = 1 ] && need "bpftool+clang" OK "build seulement" || need "bpftool+clang" MISSING "build de l'objet impossible"
[ -e /dev/kvm ] && need "KVM" OK présent || need "KVM" MISSING absent
command -v qemu-system-x86_64 >/dev/null && need "qemu" OK présent || need "qemu" MISSING absent
ls kit/chr/cache/*.img >/dev/null 2>&1 && need "image CHR" OK présente || need "image CHR" MISSING "voir kit/chr/README.md"
echo
if [ "$missing" = 0 ]; then
  echo "résumé : hôte prêt pour l'acceptation"
  exit 0
fi
echo "résumé : des lignes requises manquent"
exit 1
