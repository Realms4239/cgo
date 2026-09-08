#!/bin/bash
# ================================================================
#  CGO — guest-setup.sh : prépare TOUT côté invité Ubuntu, à la main.
#  Usage (console de la VM, droits sudo) :
#    bash guest-setup.sh [--tarball CHEMIN] [--port 9090] [--project ~/cgo]
#  Idempotent : chaque étape vérifie d'abord, n'installe que le manquant.
#  Hors-ligne : les étapes apt sont sautées proprement (message clair),
#  tout le reste (SSH déjà là ? binaire ? dashboard ?) continue.
#  Zéro identité en dur : utilisateur = $USER, IP = détectée, jamais tapée.
# ================================================================
set -uo pipefail
GREEN='\033[0;32m'; CYAN='\033[0;36m'; RED='\033[0;31m'; YEL='\033[0;33m'; NC='\033[0m'
step(){ echo -e "\n${CYAN}>>${NC} $1"; }
ok(){ echo -e "  ${GREEN}[OK]${NC} $1"; }
warn(){ echo -e "  ${YEL}[..]${NC} $1"; }
fail(){ echo -e "  ${RED}[X]${NC} $1"; exit 1; }

TARBALL=""
PORT="9090"
PROJECT="$HOME/cgo"
while [ $# -gt 0 ]; do
  case "$1" in
    --tarball) TARBALL="${2:-}"; shift 2;;
    --port) PORT="${2:-9090}"; shift 2;;
    --project) PROJECT="${2:-$HOME/cgo}"; shift 2;;
    -h|--help) sed -n '2,10p' "$0"; exit 0;;
    *) fail "option inconnue : $1 (voir --help)";;
  esac
done

if [ -f /etc/os-release ]; then . /etc/os-release; else ID="unknown"; fi
if [ "$ID" != "ubuntu" ] && [ "$ID" != "debian" ]; then
  warn "OS détecté : ${ID:-inconnu} — prévu pour Ubuntu/Debian, on continue prudemment"
fi

# Réseau : UNE sonde TCP = flaky (pare-feu de labo, stalls) — prouvé live :
# deux exécutions identiques, un OK un KO. Donc redondance : DNS système
# d'abord (résolveur local, fiable), puis deux TCP distincts.
have_net() {
  getent hosts archive.ubuntu.com >/dev/null 2>&1 && return 0
  timeout 3 bash -c '</dev/tcp/1.1.1.1/443' 2>/dev/null && return 0
  timeout 3 bash -c '</dev/tcp/8.8.8.8/53' 2>/dev/null && return 0
  return 1
}
need_pkg() { ! dpkg -s "$1" >/dev/null 2>&1; }
is_elf() { [ "$(head -c 4 "$1" 2>/dev/null | od -An -tx1 | tr -d ' \n')" = "7f454c46" ]; }
ME="${USER:-$(id -un 2>/dev/null || echo unknown)}"

step "0/5 — réseau et droits"
if have_net; then ok "internet joignable"; else warn "HORS-LIGNE : les installations apt seront sautées, le reste continue"; fi
if [ "$(id -u)" = "0" ]; then SUDO=""; ok "root direct"; elif sudo -n true 2>/dev/null; then SUDO="sudo -n"; ok "sudo sans mot de passe"; else SUDO="sudo"; warn "sudo demandera le mot de passe (normal)"; fi

step "1/5 — paquets (openssh-server, outils invité)"
if have_net && command -v apt-get >/dev/null 2>&1; then
  WANT="openssh-server iproute2 curl"
  if [ -d /sys/bus/pci/drivers/vmw_pvscsi ] 2>/dev/null || systemd-detect-virt 2>/dev/null | grep -qi vmware; then
    WANT="$WANT open-vm-tools"
  elif systemd-detect-virt 2>/dev/null | grep -qi oracle; then
    WANT="$WANT virtualbox-guest-utils"
  fi
  MISSING=""
  for p in $WANT; do need_pkg "$p" && MISSING="$MISSING $p"; done
  if [ -z "$MISSING" ]; then
    ok "tout est déjà installé ($WANT)"
  else
    $SUDO apt-get update -qq && $SUDO apt-get install -y -qq $MISSING \
      && ok "installé :$MISSING" \
      || fail "apt a échoué — relancez avec internet : sudo apt install -y$MISSING"
  fi
else
  warn "sauté (pas d'apt ou pas d'internet) — vérifiez à la main : sshd ? ip ?"
fi

step "2/5 — serveur SSH"
if systemctl is-active --quiet ssh 2>/dev/null || systemctl is-active --quiet sshd 2>/dev/null; then
  ok "sshd actif"
else
  if ! command -v sshd >/dev/null 2>&1; then fail "openssh-server absent ET hors-ligne — installez-le d'abord (internet requis une fois)"; fi
  $SUDO systemctl enable --now ssh || fail "démarrage sshd impossible"
  ok "sshd démarré + activé au boot"
fi
ss -ltn 2>/dev/null | grep -q ':22 ' && ok "port 22 en écoute" || warn "port 22 non vu en écoute — vérifiez le pare-feu (ufw allow ssh)"

step "3/5 — binaire"
mkdir -p "$PROJECT" || fail "création $PROJECT impossible"
cd "$PROJECT" || fail "cd $PROJECT impossible"
if [ -n "$TARBALL" ]; then
  [ -f "$TARBALL" ] || fail "tarball introuvable : $TARBALL"
  tar xzf "$TARBALL" ./cgo 2>/dev/null || tar xzf "$TARBALL" -C "$PROJECT" --strip-components=0
  ok "extrait depuis $TARBALL"
fi
if [ -f cgo-linux.new ]; then
  MAGIC=$(head -c 4 cgo-linux.new | od -An -tx1 | tr -d ' \n')
  [ "$MAGIC" = "7f454c46" ] || { rm -f cgo-linux.new; fail "cgo-linux.new n'est pas un ELF Linux"; }
  mv -f cgo-linux.new cgo-linux
  chmod +x cgo-linux
  ok "installé $(du -h cgo-linux | cut -f1) sha256:$(sha256sum cgo-linux | cut -c1-12)…"
elif [ -f ./cgo ] && is_elf ./cgo; then
  mv -f ./cgo ./cgo-linux; chmod +x ./cgo-linux
  ok "cgo-linux prêt (depuis l'archive)"
elif [ -x cgo-linux ]; then
  ok "binaire déjà en place — on le garde"
else
  fail "aucun binaire : poussez-le (scp) ou relancez avec --tarball CHEMIN"
fi
./cgo-linux version || warn "le binaire ne démarre pas (libc ? après un --tarball exotique ?)"

step "4/5 — dashboard sécurisé (TLS, port $PORT)"
cat > start.sh <<LAUNCHER
#!/bin/bash
cd "$PROJECT"
exec ./cgo-linux --serve --addr "0.0.0.0:$PORT" --http-addr= >> /tmp/cgo.log 2>&1
LAUNCHER
chmod +x start.sh
pkill -u "$ME" -f '[c]go-linux --serve' 2>/dev/null || true
for _ in $(seq 1 10); do pgrep -u "$ME" -f '[c]go-linux --serve' >/dev/null || break; sleep 1; done
( setsid nohup ./start.sh >/dev/null 2>&1 & )
ok "dashboard (re)lancé"

step "5/5 — santé + récapitulatif"
for i in $(seq 1 15); do
  if curl -fsS -m 2 -k "https://127.0.0.1:$PORT/api/health" 2>/dev/null | grep -q '"ok":true'; then
    ok "sain sur :$PORT (TLS) après ${i}s"
    break
  fi
  [ "$i" = "15" ] && { echo "---- log ----"; tail -n 15 /tmp/cgo.log 2>/dev/null || true; fail "dashboard injoignable sur :$PORT"; }
  sleep 1
done
echo ""
echo "RÉCAP — depuis le PC hôte, il reste :"
GUEST_IP=$(ip -br addr show 2>/dev/null | awk '/UP/ && !/lo/ && !/veth/ && !/docker/ {print $3}' | cut -d/ -f1 | head -n 1)
[ -z "$GUEST_IP" ] && GUEST_IP="<IP-invitée>"
echo "  IP invitée : $(ip -br addr show 2>/dev/null | awk '/UP/ && !/lo/ {print $3}' | cut -d/ -f1 | head -n 3 | tr '\n' ' ')"
echo "  1) clé SSH vers $ME@$GUEST_IP  2) https://meteolink.dev:$PORT (hosts + confiance)"
echo "  Détail copiable dans LISEZ-MOI.txt / kit tui."
