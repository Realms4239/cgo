#!/usr/bin/env bash
# test-fresh-clone — clone vierge → make build → le binaire sert le VRAI SPA.
# Porte "easy to use for anyone" : /, chaque asset JS/CSS référencé et
# /api/health répondent 200 — jamais le stub.
# Usage : bash kit/test-fresh-clone.sh [dossier-temp]
set -euo pipefail
DIR="${1:-$(mktemp -d)}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${PORT:-19090}"

echo "[fresh] clone dans $DIR"
git clone -q "$ROOT" "$DIR/cgo"
cd "$DIR/cgo"
# clone = HEAD du dépôt courant, pas la tête d'une autre branche
git checkout -q "$(git -C "$ROOT" rev-parse HEAD)"

echo "[fresh] build (make build)"
make build
BIN="bin/cgo"
[ -x "$BIN" ] || BIN="bin/cgo.exe"
[ -x "$BIN" ] || { echo "FAIL binaire absent après make build"; exit 1; }

BASE="http://127.0.0.1:$PORT"
LOG="$DIR/serve.log"
"$BIN" --serve --addr "127.0.0.1:$PORT" --tls=false --http-addr= --mode observe >"$LOG" 2>&1 &
PID=$!
trap 'kill "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 150); do
  curl -fsS "$BASE/api/health" >/dev/null 2>&1 && break
  sleep 0.2
done
curl -fsS "$BASE/api/health" >/dev/null || { echo "FAIL /api/health"; cat "$LOG"; exit 1; }
echo "[fresh] /api/health 200"

HTML="$(curl -fsS "$BASE/")" || { echo "FAIL GET /"; cat "$LOG"; exit 1; }
printf '%s' "$HTML" | grep -q 'id="root"' || { echo "FAIL / ne sert pas le SPA réel"; exit 1; }
echo "[fresh] / 200"

ASSETS="$(printf '%s' "$HTML" | grep -oE '(src|href)="/[^"]+\.(js|css)"' | sed -E 's/.*="([^"]+)"/\1/' | sort -u)"
[ -n "$ASSETS" ] || { echo "FAIL aucun asset JS/CSS référencé par /"; exit 1; }
while IFS= read -r a; do
  code="$(curl -s -o /dev/null -w '%{http_code}' "$BASE$a")"
  [ "$code" = "200" ] || { echo "FAIL $a → $code"; exit 1; }
  echo "[fresh] $a 200"
done <<< "$ASSETS"

echo "== PASS fresh-clone : clone → make build → SPA réel servi =="
