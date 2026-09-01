#!/usr/bin/env bash
# Meteolink — Linux dependency bootstrap (Debian/Ubuntu).
# Idempotent: safe to re-run; every step checks before acting.
# Never switches the system default congestion control — reports only.
set -euo pipefail

msg()  { printf 'meteolink-install: %s\n' "$*"; }
warn() { printf 'meteolink-install: WARNING: %s\n' "$*" >&2; }

# --- Privilege handling: use sudo only when not already root ---------------
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    warn "not root and sudo not found — apt-get and module setup will fail"
  fi
fi

# --- Plateforme : Linux = full, Windows/macOS = observation (portable) ----
if [ "$(uname -s)" = "Linux" ]; then
  # APT packages: iproute2 (tc/ss), curl (downloads), bc (rate math)
  MISSING=""
  for pkg in iproute2 curl bc; do
    if ! dpkg-query -W -f='${Status}' "$pkg" 2>/dev/null | grep -q "install ok installed"; then
      MISSING="$MISSING $pkg"
    fi
  done
  if [ -n "$MISSING" ]; then
    msg "installing packages:$MISSING"
    $SUDO apt-get update || warn "apt-get update failed"
    # shellcheck disable=SC2086
    $SUDO apt-get install -y $MISSING || warn "apt-get install failed"
  else
    msg "apt packages already present (iproute2, curl, bc)"
  fi

  # tcp_bbr kernel module: try to load, persist across reboots
  if ! lsmod 2>/dev/null | grep -q '^tcp_bbr '; then
    if $SUDO modprobe tcp_bbr 2>/dev/null; then
      msg "tcp_bbr module loaded"
    else
      warn "could not load tcp_bbr (kernel may lack BBR support)"
    fi
  else
    msg "tcp_bbr module already loaded"
  fi
  if [ -w /etc/modules-load.d ] || [ "$(id -u)" -eq 0 ]; then
    if ! grep -qs '^tcp_bbr$' /etc/modules-load.d/meteolink.conf 2>/dev/null; then
      printf 'tcp_bbr\n' | $SUDO tee /etc/modules-load.d/meteolink.conf >/dev/null
      msg "tcp_bbr persisted via /etc/modules-load.d/meteolink.conf"
    fi
  else
    warn "cannot write /etc/modules-load.d/meteolink.conf (no root) — tcp_bbr will not autoload at boot"
  fi

  # Congestion control availability: report only, never switch
  if [ -r /proc/sys/net/ipv4/tcp_available_congestion_control ]; then
    AVAILABLE="$(cat /proc/sys/net/ipv4/tcp_available_congestion_control)"
    msg "available congestion control algorithms: $AVAILABLE"
    case " $AVAILABLE " in
      *" bbr "*) msg "bbr is available" ;;
      *) warn "bbr is NOT in the available list — check kernel version and tcp_bbr module" ;;
    esac
    CURRENT="$(cat /proc/sys/net/ipv4/tcp_congestion_control)"
    msg "current system default congestion control: $CURRENT (left unchanged)"
  else
    warn "cannot read tcp_available_congestion_control from /proc"
  fi

  # Verify tc exists
  if command -v tc >/dev/null 2>&1; then
    msg "tc found at $(command -v tc)"
  else
    warn "tc not found in PATH — netem/AQM shaping will not work"
  fi

  # Capability summary (mirrors what `cgo doctor` reports)
  CAP_NET_ADMIN=no
  if command -v capsh >/dev/null 2>&1; then
    if capsh --print 2>/dev/null | grep -q 'cap_net_admin'; then
      CAP_NET_ADMIN=yes
    fi
  elif [ "$(id -u)" -eq 0 ]; then
    CAP_NET_ADMIN=yes
  fi
  msg "CAP_NET_ADMIN: $CAP_NET_ADMIN"
  msg "kernel: $(uname -r)"
  msg "summary: run 'cgo doctor' inside the app for the authoritative check"
else
  msg "hôte $(uname -s) — mode observation (pas de tc/BBR), DNS local seul"
fi

# --- DNS local portable : meteolink.dev → 127.0.0.1 (idempotent) -------
add_hosts_entry() {
  local ip="$1" host="$2" file="$3"
  if grep -qE "^[[:space:]]*$ip[[:space:]]+.*\b$host\b" "$file" 2>/dev/null; then
    msg "$host déjà dans $file"
    return 0
  fi
  if [ -w "$file" ] || [ "$(id -u)" -eq 0 ]; then
    printf '%s %s\n' "$ip" "$host" | $SUDO tee -a "$file" >/dev/null
    msg "$host → $ip ajouté à $file"
  else
    warn "ajout $host → $ip dans $file : relancez avec --hosts en root/Admin"
  fi
}
if [ "${1:-}" = "--hosts" ] || [ "${HOSTS:-0}" = "1" ]; then
  # portable : meteolink.dev → VM si présente, sinon localhost
  _vm_ip=""
  if [ -f kit/cgo-vm.yaml ] && grep -q "host:" kit/cgo-vm.yaml 2>/dev/null; then
    _vm_ip="$(grep -E '^[[:space:]]*host:' kit/cgo-vm.yaml | head -1 | sed 's/.*host:[[:space:]]*//' | tr -d '\"' | tr -d ' ')"
    [ "$_vm_ip" = "auto" ] && _vm_ip=""
  fi
  if [ -z "$_vm_ip" ]; then _vm_ip="$(cgo kit status 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1)"; fi
  if [ -z "$_vm_ip" ]; then _vm_ip="VM-IP-auto-découverte"
  if [ -f /c/Windows/System32/drivers/etc/hosts ]; then
    add_hosts_entry "127.0.0.1" "meteolink.dev" "/c/Windows/System32/drivers/etc/hosts"
    add_hosts_entry "$_vm_ip" "meteolink.vm" "/c/Windows/System32/drivers/etc/hosts"
    msg "hosts : http://meteolink.dev:9090 (local) et http://meteolink.vm:9090 (VM $_vm_ip) — éditez en Admin si Permission denied"
  elif [ -f /etc/hosts ]; then
    add_hosts_entry "127.0.0.1" "meteolink.dev" "/etc/hosts"
    add_hosts_entry "$_vm_ip" "meteolink.vm" "/etc/hosts"
  fi
else
  msg "DNS local : lancez 'bash kit/install.sh --hosts' (Admin) pour ajouter meteolink.dev → 127.0.0.1 et meteolink.vm → VM (portable, idempotent)"
fi

# --- Note -----------------------------------------------------------------
cat <<'EOF'

meteolink-install: done.
Note: shaping (netem/AQM/BBR on the access link) requires CAP_NET_ADMIN.
Run the cgo server as root, grant the capability to the binary
  sudo setcap cap_net_admin+ep ./cgo
or add your user to a sudo-capable group. Without it, cgo runs in
observation mode only.
Portable : dashboard sur http://meteolink.dev:9090 après --hosts (sinon http://localhost:9090).
VM : http://<ip-vm>:9090 (auto-découvert, ex. via cgo kit status) ou http://meteolink.vm:9090 si hosts.
EOF
