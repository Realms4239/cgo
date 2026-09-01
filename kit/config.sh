#!/usr/bin/env bash
# CGO deploy config loader — machine-local YAML subset (docs/PLAN.md M0.2).
# Usage: source config.sh /path/to/cgo-vm.yaml
# Keys: ssh.{user,host,port,key,password} vm_name vmx_path snapshot
#       project_dir dashboard_port go_min_version
set -uo pipefail

CFG_FILE="${1:?usage: source config.sh <cgo-vm.yaml>}"
[ -f "$CFG_FILE" ] || { echo "config not found: $CFG_FILE" >&2; return 1 2>/dev/null || exit 1; }

# defaults
: "${CFG_SSH_USER:=altfloat}"; : "${CFG_SSH_HOST:=auto}"
: "${CFG_SSH_PORT:=22}"; : "${CFG_SSH_KEY:=$HOME/.ssh/id_ed25519}"
: "${CFG_SSH_PASSWORD:=}"; : "${CFG_VM_NAME:=}"; : "${CFG_VMX_PATH:=}"; : "${CFG_VBOX_PATH:=}"
: "${CFG_SNAPSHOT:=}"; : "${CFG_PROJECT_DIR:=/home/altfloat/cgo}"
: "${CFG_DASHBOARD_PORT:=9090}"; : "${CFG_GO_MIN_VERSION:=1.25}"; : "${CFG_HYPERVISOR:=auto}"

_expand() {
  local v="$1"
  v="${v//%USERPROFILE%/$HOME}"; v="${v//\\//}"
  case "$v" in "~"|"~"/*) v="$HOME${v#\~}";; esac
  printf '%s' "$v"
}

_section=""
while IFS= read -r line || [ -n "$line" ]; do
  line="${line%%#*}"; line="${line%$'\r'}"; line="${line//\"/}"
  [[ -z "${line//[[:space:]]/}" ]] && continue
  if [[ "$line" =~ ^([A-Za-z_]+):[[:space:]]*$ ]]; then _section="${BASH_REMATCH[1]}"; continue; fi
  if [[ "$line" =~ ^([[:space:]]*)?([A-Za-z_]+):[[:space:]]*(.*)$ ]]; then
    local_k="${BASH_REMATCH[2]}"; v="${BASH_REMATCH[3]}"
    [ -n "${BASH_REMATCH[1]}" ] || _section=""
    k="$local_k"; [ -n "$_section" ] && k="${_section}_${local_k}"
    case "$k" in
      ssh_user) CFG_SSH_USER="$v";; ssh_host) CFG_SSH_HOST="$v";;
      ssh_port) CFG_SSH_PORT="$v";; ssh_key) CFG_SSH_KEY="$(_expand "$v")";;
      ssh_password) CFG_SSH_PASSWORD="$v";; vm_name) CFG_VM_NAME="$v";;
      vmx_path) CFG_VMX_PATH="$(_expand "$v")";; vbox_path) CFG_VBOX_PATH="$(_expand "$v")";; hypervisor) CFG_HYPERVISOR="$v";; snapshot) CFG_SNAPSHOT="$v";;
      project_dir) CFG_PROJECT_DIR="$v";; dashboard_port) CFG_DASHBOARD_PORT="$v";;
      go_min_version) CFG_GO_MIN_VERSION="$v";;
    esac
  fi
done < "$CFG_FILE"

# portabilité : host auto → IP invité via vmrun (toute VM trouvée), sinon défaut générique
if [ "${CFG_SSH_HOST}" = "auto" ]; then
  _auto_ip=""
  # 1) vmx_path explicite
  if [ -n "${CFG_VMX_PATH}" ] && [ -f "${CFG_VMX_PATH}" ]; then
    for c in "/c/Program Files (x86)/VMware/VMware Workstation/vmrun.exe" "/c/Program Files/VMware/VMware Workstation/vmrun.exe" "$(command -v vmrun 2>/dev/null || true)"; do
      [ -n "$c" ] && [ -x "$c" ] && _auto_ip="$("$c" getGuestIPAddress "${CFG_VMX_PATH}" 2>/dev/null | tr -d '\r' | head -1)" && echo "$_auto_ip" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' && break || _auto_ip=""
    done
  fi
  # 2) scan générique C:/VMs, D:/VMs si vmx_path vide
  if [ -z "$_auto_ip" ]; then
    for d in "/d/VMs" "/c/VMs" "/d" "/c"; do
      [ -d "$d" ] || continue
      for vmx in "$d"/*.vmx "$d"/*/*.vmx; do
        [ -f "$vmx" ] || continue
        for c in "/c/Program Files (x86)/VMware/VMware Workstation/vmrun.exe" "$(command -v vmrun 2>/dev/null || true)"; do
          [ -n "$c" ] && [ -x "$c" ] && _auto_ip="$("$c" getGuestIPAddress "$vmx" 2>/dev/null | tr -d '\r' | head -1)" && echo "$_auto_ip" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' && break 2 || _auto_ip=""
        done
      done
    done 2>/dev/null || true
  fi
  # 3) repli : variable d'env ou laisser auto (pas d'IP en dur — le prochain ensure la découvrira)
  if [ -z "$_auto_ip" ]; then _auto_ip="${CGO_VM_IP:-}"; fi
  if [ -z "$_auto_ip" ]; then
    return 0
  fi
  CFG_SSH_HOST="$_auto_ip"
fi
# surcharge env portable : CGO_SSH_HOST, CGO_DASHBOARD_PORT, etc.
[ -n "${CGO_SSH_HOST:-}" ] && CFG_SSH_HOST="$CGO_SSH_HOST"
[ -n "${CGO_SSH_PORT:-}" ] && CFG_SSH_PORT="$CGO_SSH_PORT"
[ -n "${CGO_DASHBOARD_PORT:-}" ] && CFG_DASHBOARD_PORT="$CGO_DASHBOARD_PORT"

export CFG_SSH_USER CFG_SSH_HOST CFG_SSH_PORT CFG_SSH_KEY CFG_SSH_PASSWORD \
       CFG_VM_NAME CFG_VMX_PATH CFG_VBOX_PATH CFG_HYPERVISOR CFG_SNAPSHOT CFG_PROJECT_DIR \
       CFG_DASHBOARD_PORT CFG_GO_MIN_VERSION
