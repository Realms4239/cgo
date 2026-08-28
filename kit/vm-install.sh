#!/bin/bash
# ================================================================
#  CGO — VM-side installer (runs INSIDE the Ubuntu VM).
#  Binary-first: host pushed cgo-linux.new; we validate, install,
#  regenerate start.sh, restart, and health-check. Idempotent.
#  Adapted from donor thesis-cgo/deploy/vm-install.sh (LIEN scope:
#  no Cloudflare, no route-pin/negctl envs).
# ================================================================
set -uo pipefail
GREEN='\033[0;32m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
step(){ echo -e "\n${CYAN}>>${NC} $1"; }
ok(){ echo -e "  ${GREEN}[OK]${NC} $1"; }
fail(){ echo -e "  ${RED}[X]${NC} $1"; exit 1; }

PROJECT_DIR="${CFG_PROJECT_DIR:-$HOME/cgo}"
PORT="${CFG_DASHBOARD_PORT:-9090}"

step "1/4 — binary"
cd "$PROJECT_DIR" || fail "project dir missing: $PROJECT_DIR"
if [ -f cgo-linux.new ]; then
  MAGIC=$(head -c 4 cgo-linux.new | od -An -tx1 | tr -d ' \n')
  [ "$MAGIC" = "7f454c46" ] || { rm -f cgo-linux.new; fail "cgo-linux.new is not a Linux ELF"; }
  mv -f cgo-linux.new cgo-linux
  chmod +x cgo-linux
  ok "installed $(du -h cgo-linux | cut -f1) sha256:$(sha256sum cgo-linux | cut -c1-12)…"
elif [ -x cgo-linux ]; then
  ok "no new binary — restarting existing"
else
  fail "no binary available"
fi

step "2/4 — launcher"
cat > start.sh <<LAUNCHER
#!/bin/bash
cd "$PROJECT_DIR"
export CGO_DASHBOARD__ADDR=":$PORT"
exec ./cgo-linux --serve >> /tmp/cgo.log 2>&1
LAUNCHER
chmod +x start.sh
ok "start.sh written (port $PORT)"

step "3/4 — restart"
pkill -u "$USER" -f '[c]go-linux --serve' 2>/dev/null || true
for _ in $(seq 1 10); do pgrep -u "$USER" -f '[c]go-linux --serve' >/dev/null || break; sleep 1; done
( setsid nohup ./start.sh >/dev/null 2>&1 & )
ok "launcher started"

step "4/4 — health"
for i in $(seq 1 15); do
  curl -fsS -m 2 "http://127.0.0.1:$PORT/api/health" >/dev/null 2>&1 && { ok "healthy on :$PORT after ${i}s"; exit 0; }
  sleep 1
done
echo "---- last log ----"; tail -n 20 /tmp/cgo.log || true
fail "health check failed on :$PORT"
