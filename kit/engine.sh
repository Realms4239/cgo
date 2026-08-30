#!/bin/bash
# ================================================================
#  CGO Kit engine — build / scan / boot / bootstrap / ensure / deploy / tunnel / status / logs
#  Hypervisor auto: VMware (vmrun) primary, VirtualBox (VBoxManage) fallback.
#  Scan entire disk C:/D: shallow depth≤3 for .vmx/.vbox (--deep full)
#  Build-then-embed strict before deploy, Cloudflare tunnel via $CLOUDFLARE_TUNNEL_TOKEN
#  ================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ACTION="" CFG="$SCRIPT_DIR/cgo-vm.yaml" DEEP=0 PUBLIC=0
while [ $# -gt 0 ]; do
  case "$1" in
    --action) ACTION="$2"; shift 2;;
    --config) CFG="$2"; shift 2;;
    --deep) DEEP=1; shift;;
    --public) PUBLIC=1; shift;;
    *) echo "unknown arg: $1" >&2; exit 2;;
  esac
done
[ -n "$ACTION" ] || { echo "usage: kit/engine.sh --action build|scan|boot|bootstrap|ensure|deploy|tunnel|status|logs [--config FILE] [--deep] [--public]" >&2; exit 2; }

# shellcheck disable=SC1091
source "$SCRIPT_DIR/config.sh" "$CFG"

# hypervisor detection
VMRUN=""; VBOX=""
for c in "/c/Program Files (x86)/VMware/VMware Workstation/vmrun.exe" "/c/Program Files/VMware/VMware Workstation/vmrun.exe" "$(command -v vmrun 2>/dev/null || true)"; do [ -n "$c" ] && [ -x "$c" ] && VMRUN="$c" && break; done
for c in "/c/Program Files/Oracle/VirtualBox/VBoxManage.exe" "$(command -v VBoxManage 2>/dev/null || true)"; do [ -n "$c" ] && [ -x "$c" ] && VBOX="$c" && break; done
HYP="$CFG_HYPERVISOR"
if [ "$HYP" = "auto" ]; then
  if [ -n "$VMRUN" ]; then HYP="vmware"
  elif [ -n "$VBOX" ]; then HYP="virtualbox"
  else HYP="none"
  fi
fi

ssh_vm() {
  ssh -o ConnectTimeout=4 -o BatchMode=yes -o StrictHostKeyChecking=accept-new -p "$CFG_SSH_PORT" -i "$CFG_SSH_KEY" "$CFG_SSH_USER@$CFG_SSH_HOST" "$@"
}
scp_vm() {
  scp -o ConnectTimeout=4 -o BatchMode=yes -P "$CFG_SSH_PORT" -i "$CFG_SSH_KEY" "$1" "$CFG_SSH_USER@$CFG_SSH_HOST:$2"
}
ssh_up() { ssh_vm 'true' >/dev/null 2>&1; }

scan_vmx() {
  local found="/tmp/cgo-vmx-candidates.txt"; : > "$found"
  if [ "$HYP" = "vmware" ] || [ "$HYP" = "auto" ]; then
    if [ -n "$VMRUN" ]; then "$VMRUN" list 2>/dev/null | grep -E '\.vmx|\.vbox' >> "$found" || true; fi
    local inv="$APPDATA/VMware/inventory.vmls"
    if [ -f "$inv" ]; then tr -d '\r' < "$inv" | grep -oE '"[^"]+\.(vmx|vbox)"' | tr -d '"' >> "$found" || true; fi
  fi
  if [ "$HYP" = "virtualbox" ] || [ "$HYP" = "auto" ]; then
    if [ -n "$VBOX" ]; then "$VBOX" list vms 2>/dev/null | grep -oE '"[^"]+\.(vbox|vmx)"' | tr -d '"' >> "$found" || true; fi
    local vdirs="$HOME/VirtualBox\ VMs" # fallback
    for d in "$HOME/VirtualBox VMs" "$HOME/Documents/Virtual Machines"; do [ -d "$d" ] && find "$(cygpath -u "$d" 2>/dev/null || echo "$d")" -maxdepth 3 -iname '*.vbox' -o -iname '*.vmx' 2>/dev/null >> "$found"; done
  fi
  # shallow C:/D: scan depth≤3 for both .vmx and .vbox, --deep full
  local letters; letters="$(powershell.exe -NoProfile -Command "(Get-PSDrive -PSProvider FileSystem).Name" 2>/dev/null | tr -d '\r' || echo "C D")"
  for L in $letters; do
    local root="/$(echo "$L" | tr 'A-Z' 'a-z')"
    [ -d "$root" ] || continue
    if [ "$DEEP" = "1" ]; then
      find "$root" \( -iname '*.vmx' -o -iname '*.vbox' \) -not -path '*/\$RECYCLE.BIN/*' 2>/dev/null >> "$found"
    else
      find "$root" -maxdepth 3 \( -iname '*.vmx' -o -iname '*.vbox' \) -not -path '*/\$RECYCLE.BIN/*' 2>/dev/null >> "$found"
    fi
  done
  sort -u "$found" | grep -v '^$' | while IFS= read -r p; do cygpath -u "$p" 2>/dev/null || echo "$p"; done | sort -u
}

pick_vmx() {
  local all; all="$(scan_vmx)"; [ -z "$all" ] && return 3
  if [ -n "$CFG_VM_NAME" ]; then
    local hit; hit="$(grep -i "/$CFG_VM_NAME\.\(vmx\|vbox\)$" <<<"$all" | head -1)"
    [ -n "$hit" ] && { echo "$hit"; return 0; }
  fi
  # prefer vmx_path/vbox_path if set
  if [ -n "$CFG_VMX_PATH" ] && [ -f "$CFG_VMX_PATH" ]; then echo "$CFG_VMX_PATH"; return 0; fi
  if [ -n "$CFG_VBOX_PATH" ] && [ -f "$CFG_VBOX_PATH" ]; then echo "$CFG_VBOX_PATH"; return 0; fi
  local n; n="$(wc -l <<<"$all")"
  if [ "$n" -eq 1 ]; then echo "$all"; return 0; fi
  echo "$all" >&2; return 3
}

save_vmx() {
  local p="$1" name; name="$(basename "$p" | sed 's/\.vmx$//;s/\.vbox$//')"
  touch "$CFG"
  if [[ "$p" == *.vbox ]]; then
    grep -q '^vbox_path:' "$CFG" && sed -i "s|^vbox_path:.*|vbox_path: \"$p\"|" "$CFG" || echo "vbox_path: \"$p\"" >> "$CFG"
    grep -q '^hypervisor:' "$CFG" && sed -i "s|^hypervisor:.*|hypervisor: \"virtualbox\"|" "$CFG" || echo "hypervisor: \"virtualbox\"" >> "$CFG"
  else
    grep -q '^vmx_path:' "$CFG" && sed -i "s|^vmx_path:.*|vmx_path: \"$p\"|" "$CFG" || echo "vmx_path: \"$p\"" >> "$CFG"
    grep -q '^hypervisor:' "$CFG" && sed -i "s|^hypervisor:.*|hypervisor: \"vmware\"|" "$CFG" || echo "hypervisor: \"vmware\"" >> "$CFG"
  fi
  grep -q '^vm_name:' "$CFG" && sed -i "s|^vm_name:.*|vm_name: \"$name\"|" "$CFG" || echo "vm_name: \"$name\"" >> "$CFG"
  export CFG_VMX_PATH="$p" CFG_VM_NAME="$name"
}

boot_vmx() {
  local p="$1"
  if [[ "$p" == *.vbox ]]; then
    [ -n "$VBOX" ] || { echo "VBoxManage not found — start VM manually" >&2; return 4; }
    echo "[boot] starting VirtualBox $(basename "$p") --type headless"
    "$VBOX" startvm "$(basename "$p" .vbox)" --type headless >/dev/null
  else
    [ -n "$VMRUN" ] || { echo "vmrun not found — start VM manually" >&2; return 4; }
    if [ -n "$CFG_SNAPSHOT" ]; then echo "[boot] reverting to snapshot '$CFG_SNAPSHOT'"; "$VMRUN" revertToSnapshot "$p" "$CFG_SNAPSHOT" >/dev/null; fi
    echo "[boot] starting $(basename "$p") (nogui)"; "$VMRUN" start "$p" nogui >/dev/null
  fi
}

case "$ACTION" in
  build)
    echo "[build] go vet..."
    (cd "$ROOT" && go vet ./...) || exit 2
    echo "[build] bun typecheck..."
    (cd "$ROOT/web/frontend" && C:/Users/ASUS/.bun/bin/bun.exe x tsc --noEmit) || exit 2
    echo "[build] bun build..."
    (cd "$ROOT/web/frontend" && C:/Users/ASUS/.bun/bin/bun.exe run build) || exit 2
    echo "[build] check-bundle TOTAL 600 echarts 350..."
    (cd "$ROOT/web/frontend" && node scripts/check-bundle.mjs) || exit 2
    echo "[build] vitest..."
    (cd "$ROOT/web/frontend" && C:/Users/ASUS/.bun/bin/bun.exe x vitest run) || exit 2
    echo "[build] embed analysis..."
    mkdir -p "$ROOT/kit/logs"
    sha256sum "$ROOT/web/frontend/dist/assets/"*.js 2>/dev/null | head -5 > "$ROOT/kit/logs/build.log" || true
    echo "[build] ok TOTAL <600 echarts <350"
    ;;
  scan)
    all="$(scan_vmx)"
    if [ -z "$all" ]; then echo "no .vmx/.vbox found ($( [ "$DEEP" = 1 ] && echo deep || echo shallow ) hypervisor $HYP)"; exit 3; fi
    echo "$all"
    n="$(wc -l <<<"$all")"
    if [ "$n" -eq 1 ]; then save_vmx "$(head -1 <<<"$all")"; echo "[scan] selected: $(head -1 <<<"$all") hypervisor $HYP"
    elif [ -n "$CFG_VM_NAME" ] && hit="$(grep -i "/$CFG_VM_NAME\.\(vmx\|vbox\)$" <<<"$all" | head -1)" && [ -n "$hit" ]; then save_vmx "$hit"; echo "[scan] selected: $hit hypervisor $HYP"
    else echo "[scan] ambiguous ($n candidates hypervisor $HYP) — set vm_name in $CFG or pass --deep"; exit 3; fi
    ;;
  boot)
    p="$(pick_vmx)" || { echo "pick failed ($?) hypervisor $HYP" >&2; exit 3; }
    save_vmx "$p"; boot_vmx "$p"
    ;;
  bootstrap)
    echo "[bootstrap] installing deps on VM..."
    ssh_vm "sudo apt update && sudo apt install -y iproute2 curl bc && sudo modprobe tcp_bbr || true && sudo sysctl -w net.ipv4.tcp_congestion_control=bbr || true" || exit 6
    ssh_vm "sudo ip link show veth-c >/dev/null 2>&1 || (sudo ip link add veth-c type veth peer name veth-s && sudo ip link set veth-c up && sudo ip link set veth-s up && echo 'veth-c/veth-s up')" || true
    echo "[bootstrap] done"
    ;;
  ensure)
    if ssh_up; then echo "[ensure] SSH already up at $CFG_SSH_HOST hypervisor $HYP"; exit 0; fi
    p="$(pick_vmx)" || { echo "pick failed ($?) hypervisor $HYP — run with --deep or set vmx_path/vbox_path in yaml" >&2; exit 3; }
    save_vmx "$p"; boot_vmx "$p"
    echo "[ensure] waiting for SSH (max 300 s) hypervisor $HYP"
    for i in $(seq 1 60); do ssh_up && { echo "[ensure] SSH up after ~$((i*5)) s"; exit 0; }; sleep 5; done
    echo "[ensure] timeout waiting for SSH" >&2; exit 5
    ;;
  deploy)
    "$SCRIPT_DIR/engine.sh" --action build --config "$CFG" || exit $?
    "$SCRIPT_DIR/engine.sh" --action ensure --config "$CFG" ${DEEP:+--deep} || exit $?
    if [ "$PUBLIC" = "1" ]; then
      if [ -z "${CLOUDFLARE_TUNNEL_TOKEN:-}" ] && [ -z "${CF_TUNNEL_TOKEN:-}" ]; then echo "[deploy] --public requires CLOUDFLARE_TUNNEL_TOKEN env (cfut_...)" >&2; exit 2; fi
      echo "[deploy] public tunnel via cloudflared..."
      nohup cloudflared tunnel run --token "${CLOUDFLARE_TUNNEL_TOKEN:-$CF_TUNNEL_TOKEN}" >/tmp/cgo-tunnel.log 2>&1 &
      echo "[deploy] tunnel pid $! log /tmp/cgo-tunnel.log"
    fi
    echo "[deploy] cross-compiling linux/amd64..."
    (cd "$ROOT" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kit/cgo-linux.new ./cmd/cgo) || exit 6
    echo "[deploy] pushing binary + installer..."
    ssh_vm "mkdir -p $CFG_PROJECT_DIR/kit"
    scp_vm "$ROOT/kit/cgo-linux.new" "$CFG_PROJECT_DIR/cgo-linux.new" || exit 7
    scp_vm "$SCRIPT_DIR/vm-install.sh" "$CFG_PROJECT_DIR/kit/vm-install.sh" || exit 7
    echo "[deploy] installing on VM..."
    ssh_vm "cd $CFG_PROJECT_DIR && bash kit/vm-install.sh" || exit 8
    # embed analysis: compare dist sha vs build sha vs integrity sha
    mkdir -p "$ROOT/kit/logs"
    sha256sum "$ROOT/kit/cgo-linux.new" 2>/dev/null | cut -c1-8 > "$ROOT/kit/logs/build.log" || true
    echo "[deploy] done → http://$CFG_SSH_HOST:$CFG_DASHBOARD_PORT hypervisor $HYP $( [ "$PUBLIC" = 1 ] && echo "public tunnel active" || echo "local")"
    ;;
  tunnel)
    if [ -z "${CLOUDFLARE_TUNNEL_TOKEN:-}" ] && [ -z "${CF_TUNNEL_TOKEN:-}" ]; then echo "CLOUDFLARE_TUNNEL_TOKEN required" >&2; exit 2; fi
    exec cloudflared tunnel run --token "${CLOUDFLARE_TUNNEL_TOKEN:-$CF_TUNNEL_TOKEN}"
    ;;
  status)
    if ssh_up; then
      echo "[status] SSH: up hypervisor $HYP"
      ssh_vm "pgrep -af 'cgo-linux' || echo 'cgo-linux: not running'" || true
      curl -fsS -m 4 "http://$CFG_SSH_HOST:$CFG_DASHBOARD_PORT/api/health" >/dev/null 2>&1 && echo "[status] dashboard: healthy on :$CFG_DASHBOARD_PORT" || echo "[status] dashboard: unreachable"
    else
      echo "[status] SSH: down hypervisor $HYP"
      [ -n "$VMRUN" ] && "$VMRUN" list | sed 's/^/[vmrun] /'
      [ -n "$VBOX" ] && "$VBOX" list runningvms | sed 's/^/[vbox] /'
    fi
    ;;
  logs)
    ssh_vm "tail -n 40 /tmp/cgo.log" || exit 8
    ;;
esac
