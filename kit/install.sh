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

# --- APT packages: iproute2 (tc/ss), curl (downloads), bc (rate math) ------
MISSING=""
for pkg in iproute2 curl bc; do
  if ! dpkg-query -W -f='${Status}' "$pkg" 2>/dev/null | grep -q "install ok installed"; then
    MISSING="$MISSING $pkg"
  fi
done
if [ -n "$MISSING" ]; then
  msg "installing packages:$MISSING"
  $SUDO apt-get update
  # shellcheck disable=SC2086
  $SUDO apt-get install -y $MISSING
else
  msg "apt packages already present (iproute2, curl, bc)"
fi

# --- tcp_bbr kernel module: try to load, persist across reboots -----------
if ! lsmod 2>/dev/null | grep -q '^tcp_bbr '; then
  if $SUDO modprobe tcp_bbr 2>/dev/null; then
    msg "tcp_bbr module loaded"
  else
    warn "could not load tcp_bbr (kernel may lack BBR support)"
  fi
else
  msg "tcp_bbr module already loaded"
fi
# Persist autoload at boot; harmless if the module is built-in.
if [ -w /etc/modules-load.d ] || [ "$(id -u)" -eq 0 ]; then
  if ! grep -qs '^tcp_bbr$' /etc/modules-load.d/meteolink.conf 2>/dev/null; then
    printf 'tcp_bbr\n' | $SUDO tee /etc/modules-load.d/meteolink.conf >/dev/null
    msg "tcp_bbr persisted via /etc/modules-load.d/meteolink.conf"
  fi
else
  warn "cannot write /etc/modules-load.d/meteolink.conf (no root) — tcp_bbr will not autoload at boot"
fi

# --- Congestion control availability: report only, never switch ----------
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

# --- Verify tc exists ------------------------------------------------------
if command -v tc >/dev/null 2>&1; then
  msg "tc found at $(command -v tc)"
else
  warn "tc not found in PATH — netem/AQM shaping will not work"
fi

# --- Capability summary (mirrors what `cgo doctor` reports) ---------------
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

# --- Note -----------------------------------------------------------------
cat <<'EOF'

meteolink-install: done.
Note: shaping (netem/AQM/BBR on the access link) requires CAP_NET_ADMIN.
Run the cgo server as root, grant the capability to the binary
  sudo setcap cap_net_admin+ep ./cgo
or add your user to a sudo-capable group. Without it, cgo runs in
observation mode only.
EOF
