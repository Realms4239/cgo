#!/usr/bin/env bash
# test-fresh-clone — simule un NOUVEL utilisateur : clone vierge → cgo setup
# --yes --no-vm → binaire → doctor. La porte "easy to use for anyone".
# Usage : bash kit/test-fresh-clone.sh [dossier-temp]
set -euo pipefail
DIR="${1:-$(mktemp -d)}"
echo "[fresh] clone dans $DIR"
git clone -q --depth 1 . "$DIR/cgo" 2>/dev/null || { git clone -q . "$DIR/cgo"; }
cd "$DIR/cgo"
echo "[fresh] setup headless (--yes --no-vm, sans build binaire local)"
# pas de build lourd ici : on vérifie le wizard déterministe + doctor
printf '' | go run ./cmd/cgo setup --yes --no-vm >/tmp/fresh-setup.log 2>&1 || {
  tail -n 5 /tmp/fresh-setup.log
  echo "FAIL setup headless"
  exit 1
}
grep -q "setup terminé" /tmp/fresh-setup.log || { echo "FAIL récap manquant"; tail /tmp/fresh-setup.log; exit 1; }
echo "[fresh] doctor"
go run ./cmd/cgo doctor >/tmp/fresh-doctor.log 2>&1 || { echo "FAIL doctor"; cat /tmp/fresh-doctor.log; exit 1; }
echo "[fresh] kit doctor (moteur Go sur clone frais)"
go run ./cmd/cgo kit doctor >/tmp/fresh-kit.log 2>&1 || { echo "FAIL kit doctor"; cat /tmp/fresh-kit.log; exit 1; }
grep -q "\[doctor\] hôte" /tmp/fresh-kit.log || { echo "FAIL sortie kit doctor"; cat /tmp/fresh-kit.log; exit 1; }
echo "[fresh] test-overhaul sur clone frais"
bash kit/test-overhaul.sh >/dev/null 2>&1 || { echo "FAIL test-overhaul sur le clone"; exit 1; }
echo "== PASS fresh-clone : clone → setup --yes → kit doctor =="