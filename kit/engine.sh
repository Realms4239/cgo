#!/bin/bash
# ================================================================
#  CGO deploy engine — scan / boot / ensure / deploy / status / logs
#  VMware Workstation (vmrun) primary; plain SSH fallback for boot.
#  See docs/PLAN.md M0.2 and docs/SPEC.md §4.
#  Usage: bash engine.sh --action <action> [--config FILE] [--deep]
# ================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ACTION="" CFG="$SCRIPT_DIR/cgo-vm.yaml" DEEP=0
while [ $# -gt 0 ]; do
  case "$1" in
    --action) ACTION="$2"; shift 2;;
    --config) CFG="$2"; shift 2;;
    --deep) DEEP=1; shift;;
    *) echo "unknown arg: $1" >&2; exit 2;;
  esac
done
[ -n "$ACTION" ] || { echo "usage: engine.sh --action scan|boot|ensure|deploy|status|logs" >&2; exit 2; }

# shellcheck disable=SC1091
source "$SCRIPT_DIR/config.sh" "$CFG"

VMRUN=""
for c in "/c/Program Files (x86)/VMware/VMware Workstation/vmrun.exe" \
         "/c/Program Files/VMware/VMware Workstation/vmrun.exe" "$(command -v vmrun)"; do
  [ -n "$c" ] && [ -x "$c" ] && VMRUN="$c" && break
done

ssh_vm() { # <remote command...>
  ssh -o ConnectTimeout=4 -o BatchMode=yes -o StrictHostKeyChecking=accept-new \
      -p "$CFG_SSH_PORT" -i "$CFG_SSH_KEY" "$CFG_SSH_USER@$CFG_SSH_HOST" "$@"
}
scp_vm() { # <local> <remote>
  scp -o ConnectTimeout=4 -o BatchMode=yes -P "$CFG_SSH_PORT" \
      -i "$CFG_SSH_KEY" "$1" "$CFG_SSH_USER@$CFG_SSH_HOST:$2"
}
ssh_up() { ssh_vm 'true' >/dev/null 2>&1; }

# ---- discovery -----------------------------------------------------
scan_vmx() {
  local found="/tmp/cgo-vmx-candidates.txt"
  : > "$found"
  if [ -n "$VMRUN" ]; then "$VMRUN" list 2>/dev/null | grep '\.vmx' >> "$found" || true; fi
  local inv="$APPDATA/VMware/inventory.vmls"
  if [ -f "$inv" ]; then
    tr -d '\r' < "$inv" | grep -oE '"[^"]+\.vmx"' | tr -d '"' >> "$found" || true
  fi
  local dirs="$HOME/Documents/Virtual Machines"
  for d in "$dirs"; do [ -d "$d" ] && find "$(cygpath -u "$d" 2>/dev/null || echo "$d")" -maxdepth 3 -iname '*.vmx' 2>/dev/null >> "$found"; done
  if [ ! -s "$found" ]; then
    local letters; letters="$(powershell.exe -NoProfile -Command "(Get-PSDrive -PSProvider FileSystem).Name" 2>/dev/null | tr -d '\r')"
    for L in $letters; do
      local root="/$(echo "$L" | tr 'A-Z' 'a-z')"
      [ -d "$root" ] || continue
      if [ "$DEEP" = "1" ]; then
        find "$root" -iname '*.vmx' -not -path '*/\$RECYCLE.BIN/*' 2>/dev/null >> "$found"
      else
        find "$root" -maxdepth 3 -iname '*.vmx' -not -path '*/\$RECYCLE.BIN/*' 2>/dev/null >> "$found"
      fi
    done
  fi
  sort -u "$found" | grep -v '^$' | while IFS= read -r p; do cygpath -u "$p" 2>/dev/null || echo "$p"; done | sort -u
}

pick_vmx() {
  [ -n "$CFG_VMX_PATH" ] && [ -f "$CFG_VMX_PATH" ] && { echo "$CFG_VMX_PATH"; return 0; }
  local all; all="$(scan_vmx)"
  [ -z "$all" ] && return 3
  if [ -n "$CFG_VM_NAME" ]; then
    local hit; hit="$(grep -i "/$CFG_VM_NAME\.vmx$" <<<"$all" | head -1)"
    [ -n "$hit" ] && { echo "$hit"; return 0; }
  fi
  local n; n="$(wc -l <<<"$all")"
  if [ "$n" -eq 1 ]; then echo "$all"; return 0; fi
  echo "$all" >&2; return 3
}

save_vmx() { # persist vm_name + vmx_path back into machine-local yaml
  local p="$1" name; name="$(basename "$p" .vmx)"
  touch "$CFG"
  grep -q '^vmx_path:' "$CFG" && sed -i "s|^vmx_path:.*|vmx_path: \"$p\"|" "$CFG" || echo "vmx_path: \"$p\"" >> "$CFG"
  grep -q '^vm_name:'  "$CFG" && sed -i "s|^vm_name:.*|vm_name: \"$name\"|"   "$CFG" || echo "vm_name: \"$name\""   >> "$CFG"
  export CFG_VMX_PATH="$p" CFG_VM_NAME="$name"
}

boot_vmx() {
  local p="$1"
  [ -n "$VMRUN" ] || { echo "vmrun not found — cannot boot; start the VM manually" >&2; return 4; }
  if [ -n "$CFG_SNAPSHOT" ]; then
    echo "[boot] reverting to snapshot '$CFG_SNAPSHOT'"
    "$VMRUN" revertToSnapshot "$p" "$CFG_SNAPSHOT" >/dev/null
  fi
  echo "[boot] starting $(basename "$p") (nogui)"
  "$VMRUN" start "$p" nogui >/dev/null
}

# ---- actions -------------------------------------------------------
case "$ACTION" in
  scan)
    all="$(scan_vmx)"
    if [ -z "$all" ]; then echo "no .vmx found ($( [ "$DEEP" = 1 ] && echo deep || echo shallow ))"; exit 3; fi
    echo "$all"
    n="$(wc -l <<<"$all")"
    if [ "$n" -eq 1 ]; then save_vmx "$(head -1 <<<"$all")"; echo "[scan] selected: $(head -1 <<<"$all")"
    elif [ -n "$CFG_VM_NAME" ] && hit="$(grep -i "/$CFG_VM_NAME\.vmx$" <<<"$all" | head -1)" && [ -n "$hit" ]; then save_vmx "$hit"; echo "[scan] selected: $hit"
    else echo "[scan] ambiguous ($n candidates) — set vm_name in $CFG or pass --deep"; exit 3; fi
    ;;
  boot)
    p="$(pick_vmx)" || { echo "pick failed ($?)" >&2; exit 3; }
    save_vmx "$p"; boot_vmx "$p"
    ;;
  ensure)
    if ssh_up; then echo "[ensure] SSH already up at $CFG_SSH_HOST"; exit 0; fi
    p="$(pick_vmx)" || { echo "pick failed ($?) — run with --deep or set vmx_path in yaml" >&2; exit 3; }
    save_vmx "$p"; boot_vmx "$p"
    echo "[ensure] waiting for SSH (max 300 s)"
    for i in $(seq 1 60); do
      ssh_up && { echo "[ensure] SSH up after ~$((i*5)) s"; exit 0; }
      sleep 5
    done
    echo "[ensure] timeout waiting for SSH" >&2; exit 5
    ;;
  deploy)
    "$0" --action ensure --config "$CFG" ${DEEP:+--deep} || exit $?
    DIST="$ROOT/web/frontend/dist"
    if [ ! -f "$DIST/index.html" ]; then echo "[deploy] building frontend…"; (cd "$ROOT/web/frontend" && bun run build) || exit 6; fi
    echo "[deploy] cross-compiling linux/amd64…"
    (cd "$ROOT" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o deploy/cgo-linux.new ./cmd/cgo) || exit 6
    echo "[deploy] pushing binary + installer…"
    ssh_vm "mkdir -p $CFG_PROJECT_DIR/deploy"
    scp_vm "$ROOT/deploy/cgo-linux.new" "$CFG_PROJECT_DIR/cgo-linux.new" || exit 7
    scp_vm "$SCRIPT_DIR/vm-install.sh"  "$CFG_PROJECT_DIR/deploy/vm-install.sh" || exit 7
    echo "[deploy] installing on VM…"
    ssh_vm "cd $CFG_PROJECT_DIR && bash deploy/vm-install.sh" || exit 8
    echo "[deploy] done → http://$CFG_SSH_HOST:$CFG_DASHBOARD_PORT"
    ;;
  status)
    if ssh_up; then
      echo "[status] SSH: up"
      ssh_vm "pgrep -af 'cgo-linux' || echo 'cgo-linux: not running'" || true
      curl -fsS -m 4 "http://$CFG_SSH_HOST:$CFG_DASHBOARD_PORT/api/health" >/dev/null 2>&1 \
        && echo "[status] dashboard: healthy on :$CFG_DASHBOARD_PORT" \
        || echo "[status] dashboard: unreachable"
    else
      echo "[status] SSH: down"
      [ -n "$VMRUN" ] && "$VMRUN" list | sed 's/^/[vmrun] /'
    fi
    ;;
  logs)
    ssh_vm "tail -n 40 /tmp/cgo.log" || exit 8
    ;;
esac
