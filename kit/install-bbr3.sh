#!/bin/bash
# ============================================================================
#  install-bbr3.sh — build and install Google's BBRv3 kernel, opt-in.
#
#  BBRv3 is not a congestion-control module: it ships as a full kernel tree
#  (google/bbr) with core TCP changes. This script builds the pinned branch
#  on this VM, installs the image, and leaves the reboot to the operator.
#  It is never called by `cgo run`, `cgo deploy` or `cgo kit bootstrap`.
#
#  Usage: sudo bash kit/install-bbr3.sh [--yes] [--reboot] [--src DIR] [--jobs N]
#
#  Requires: Debian/Ubuntu, network, ~20 GiB free, Secure Boot off.
#  Time: 30–90 min depending on CPU.
# ============================================================================
set -euo pipefail

BRANCH="bbr-v3-2026-09-16-01"
COMMIT="674859761ded9f32690c7bdaf22ef585531452ed"
REPO="https://github.com/google/bbr.git"
SRC="${HOME}/bbr3-src"
JOBS="$(nproc 2>/dev/null || echo 4)"
ASSUME_YES=0
REBOOT=0

usage() {
  echo "usage: $0 [--yes] [--reboot] [--src DIR] [--jobs N]" >&2
}

while [ $# -gt 0 ]; do
  case "$1" in
    --yes) ASSUME_YES=1; shift ;;
    --reboot) REBOOT=1; shift ;;
    --src) SRC="$2"; shift 2 ;;
    --jobs) JOBS="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage; exit 2 ;;
  esac
done

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if sudo -n true 2>/dev/null; then SUDO="sudo -n"; else
    echo "[!] root requis (ou sudo NOPASSWD) — console : sudo bash kit/install-bbr3.sh" >&2
    exit 1
  fi
fi

log() { echo "[bbr3] $*"; }
die() { echo "[bbr3] ERREUR : $*" >&2; exit 1; }

if grep -qw bbr3 /proc/sys/net/ipv4/tcp_available_congestion_control 2>/dev/null; then
  log "bbr3 déjà exposé par $(uname -r) — rien à faire"
  exit 0
fi

command -v apt-get >/dev/null || die "apt-get absent — script prévu pour Debian/Ubuntu"
command -v git >/dev/null || die "git absent — apt-get install -y git"
[ -r /proc/sys/kernel/osrelease ] || die "/proc/sys/kernel/osrelease illisible"

mkdir -p "$(dirname "$SRC")"
AVAIL_KB=$(df -Pk "$(dirname "$SRC")" | awk 'NR==2{print $4}')
[ "${AVAIL_KB:-0}" -ge 20971520 ] || \
  die "espace insuffisant sous $(dirname "$SRC") : $(( ${AVAIL_KB:-0} / 1024 / 1024 )) GiB libres, 20 GiB requis"

if [ "$ASSUME_YES" -ne 1 ]; then
  echo "[bbr3] cette opération installe des paquets de compilation, télécharge"
  echo "        google/bbr ($BRANCH), construit un noyau 7.1 et l'installe."
  echo "        Le noyau courant reste dans GRUB ; le redémarrage est manuel."
  read -r -p "[bbr3] continuer ? [o/N] " ans
  case "$ans" in o|O|y|Y) ;; *) echo "annulé"; exit 1 ;; esac
fi

log "paquets de compilation…"
$SUDO apt-get update -y
$SUDO env DEBIAN_FRONTEND=noninteractive apt-get install -y \
  build-essential bc bison flex libssl-dev libelf-dev dwarves \
  git rsync cpio zstd kmod mokutil debhelper fakeroot dpkg-dev

if command -v mokutil >/dev/null && mokutil --sb-state 2>/dev/null | grep -qi enabled; then
  die "Secure Boot actif — désactivez-le avant d'installer un noyau non signé"
fi

if [ ! -d "$SRC/.git" ]; then
  log "clone $REPO…"
  mkdir -p "$SRC"
  git -C "$SRC" init -q
  git -C "$SRC" remote add origin "$REPO" 2>/dev/null || git -C "$SRC" remote set-url origin "$REPO"
else
  log "source existante : $SRC"
fi
log "fetch $BRANCH…"
git -C "$SRC" fetch --depth 1 origin "$BRANCH"
git -C "$SRC" checkout -q --detach FETCH_HEAD
[ "$(git -C "$SRC" rev-parse HEAD)" = "$COMMIT" ] || \
  die "la branche a bougé : HEAD=$(git -C "$SRC" rev-parse --short HEAD), épinglé=$COMMIT — mettez le pin à jour sciemment"

cd "$SRC"
log "configuration (base : /boot/config-$(uname -r))…"
if [ -r "/boot/config-$(uname -r)" ]; then
  cp "/boot/config-$(uname -r)" .config
else
  log "/boot/config-$(uname -r) absent — make defconfig"
  make -s defconfig
fi
make -s olddefconfig
# BBR v1 ET v3 intégrés (pas des modules) : /proc les liste dès le boot.
scripts/config --enable TCP_CONG_BBR
scripts/config --enable TCP_CONG_BBR3
scripts/config --enable TCP_CONG_CUBIC
scripts/config --set-str DEFAULT_TCP_CONG "cubic"
scripts/config --set-str LOCALVERSION "-cgo-bbr3"
# Clés absentes dans l'arborescence upstream : le build échoue sinon.
scripts/config --disable SYSTEM_TRUSTED_KEYS
scripts/config --disable SYSTEM_REVOCATION_KEYS
# Le banc : netem, CAKE, fq_codel, et la plomberie IFB du slice 6.
scripts/config --enable NET_SCH_NETEM
scripts/config --enable NET_SCH_CAKE
scripts/config --enable NET_SCH_FQ_CODEL
scripts/config --enable NET_CLS_MATCHALL
scripts/config --enable NET_ACT_MIRRED
scripts/config --enable IFB
make -s olddefconfig

for opt in \
  CONFIG_TCP_CONG_BBR=y CONFIG_TCP_CONG_BBR3=y CONFIG_TCP_CONG_CUBIC=y \
  CONFIG_NET_SCH_NETEM=y CONFIG_NET_SCH_CAKE=y CONFIG_NET_SCH_FQ_CODEL=y \
  CONFIG_NET_CLS_MATCHALL=y CONFIG_NET_ACT_MIRRED=y CONFIG_IFB=y; do
  grep -qx "$opt" .config || die "option manquante après olddefconfig : $opt"
done

KREL=$(make -s kernelrelease)
LOG="$SRC/build-$(date +%Y%m%d-%H%M%S).log"
log "construction $KREL avec $JOBS jobs (30–90 min) — log : $LOG"
make -j"$JOBS" bindeb-pkg 2>&1 | tee "$LOG"
grep -qx 'CONFIG_TCP_CONG_BBR3=y' .config || die "BBRv3 non intégré au noyau construit"

DEB=$(ls -t "$(dirname "$SRC")"/linux-image-*.deb 2>/dev/null | head -1)
[ -n "$DEB" ] || die "paquet linux-image introuvable — voir $LOG"
dpkg-deb -f "$DEB" Package | grep -q "$KREL" || \
  die "le paquet le plus récent n'est pas $KREL : $DEB"

log "installation $DEB…"
$SUDO dpkg -i "$DEB" || $SUDO apt-get -f install -y

$SUDO mkdir -p /var/lib/cgo
$SUDO tee /var/lib/cgo/bbr3-kernel.json >/dev/null <<EOF
{"branch":"$BRANCH","commit":"$COMMIT","kernel":"$KREL","built":"$(date -Is)"}
EOF

log "installé : $KREL (BBRv1 = bbr, BBRv3 = bbr3)"
log "après redémarrage : cgo doctor | grep -E 'bbr\\b|bbr3'"
if [ "$REBOOT" -eq 1 ]; then
  log "redémarrage…"
  $SUDO systemctl reboot
else
  log "redémarrez quand vous êtes prêt : sudo reboot"
fi
