#!/usr/bin/bash
# Shim de compatibilité — le moteur vit dans le binaire : `cgo kit`.
# Traduit --action X en X (nouvelle signature) et cherche le binaire local.
# Retiré à v1.2 : utilisez `cgo kit <action>` directement.
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARGS=()
while [ $# -gt 0 ]; do
  case "$1" in
    --action) ARGS+=("$2"); shift 2;;
    *) ARGS+=("$1"); shift;;
  esac
done
if [ -x "$DIR/bin/cgo" ]; then
  exec "$DIR/bin/cgo" kit "${ARGS[@]}"
elif [ -x "$DIR/bin/cgo.exe" ]; then
  exec "$DIR/bin/cgo.exe" kit "${ARGS[@]}"
elif command -v cgo >/dev/null 2>&1; then
  exec cgo kit "${ARGS[@]}"
else
  (cd "$DIR" && go run ./cmd/cgo kit "${ARGS[@]}")
fi
