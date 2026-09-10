#!/bin/bash
# ================================================================
#  CGO — guest-setup.sh : prépare TOUT côté invité Ubuntu, à la main.
#  Usage (console de la VM, droits sudo) :
#    bash guest-setup.sh [--tarball CHEMIN] [--port 9090] [--project ~/cgo]
#    bash guest-setup.sh --check   # audit lecture seule (code 1 si manque)
#  Idempotent : chaque étape vérifie d'abord, n'installe que le manquant.
#  Hors-ligne : les étapes apt sont sautées proprement (message clair),
#  tout le reste (SSH déjà là ? binaire ? dashboard ?) continue.
#  Zéro identité en dur : utilisateur = $USER, IP = détectée, jamais tapée.
#  Journal auto : /tmp/guest-setup-<date>.log (preuves envoyables).
# ================================================================
set -uo pipefail
# Couleurs seulement sur terminal (logs propres quand redirigé).
if [ -t 1 ]; then GREEN='\033[0;32m'; CYAN='\033[0;36m'; RED='\033[0;31m'; YEL='\033[0;33m'; NC='\033[0m'; else GREEN=''; CYAN=''; RED=''; YEL=''; NC=''; fi
step(){ echo -e "\n${CYAN}>>${NC} $1"; }
ok(){ echo -e "  ${GREEN}[OK]${NC} $1"; }
warn(){ echo -e "  ${YEL}[..]${NC} $1"; }
fail(){ echo -e "  ${RED}[X]${NC} $1"; exit 1; }

TARBALL=""
PORT="9090"
PROJECT="$HOME/cgo"
CHECK=0
while [ $# -gt 0 ]; do
  case "$1" in
    --tarball) TARBALL="${2:-}"; shift 2;;
    --port) PORT="${2:-9090}"; shift 2;;
    --project) PROJECT="${2:-$HOME/cgo}"; shift 2;;
    --check) CHECK=1; shift;;
    -h|--help) sed -n '2,11p' "$0"; exit 0;;
    *) fail "option inconnue : $1 (voir --help)";;
  esac
done

LOG="/tmp/guest-setup-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee -a "$LOG") 2>&1
echo "journal : $LOG"

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
MISS=0
chk() { if eval "$2"; then ok "$1"; else warn "$1 — MANQUANT"; MISS=$((MISS+1)); fi }

# --check : audit lecture seule, zéro mutation. Sortie 0 = prêt, 1 = manque.
do_check() {
  step "audit (lecture seule)"
  chk "internet" "have_net"
  chk "apt présent" "command -v apt-get >/dev/null 2>&1"
  chk "openssh-server installé" "! need_pkg openssh-server"
  chk "sshd actif" "systemctl is-active --quiet ssh 2>/dev/null || systemctl is-active --quiet sshd 2>/dev/null"
  chk "port 22 en écoute" "ss -ltn 2>/dev/null | grep -q ':22 '"
  chk "binaire $PROJECT/cgo-linux" '[ -x "$PROJECT/cgo-linux" ]'
  chk "sudoers cgo-network (tc/ip sans mot de passe)" '! sudo -n ip link show lo 2>&1 | grep -qi password'
  chk "banc de mesure (small :8081)" 'curl -fsS -m2 http://10.200.0.1:8081/small -o /dev/null 2>/dev/null'
  chk "dashboard :$PORT sain (TLS)" "curl -fsS -m 2 -k \"https://127.0.0.1:$PORT/api/health\" 2>/dev/null | grep -q '\"ok\":true'"
  if [ "$MISS" = "0" ]; then ok "PRÊT — rien à faire"; else warn "$MISS point(s) à corriger — relancez sans --check"; fi
  exit "$MISS"
}
[ "$CHECK" = "1" ] && do_check

step "0/6 — réseau et droits"
if have_net; then ok "internet joignable"; else warn "HORS-LIGNE : les installations apt seront sautées, le reste continue"; fi
if [ "$(id -u)" = "0" ]; then SUDO=""; ok "root direct"; elif sudo -n ip link show lo 2>&1 | grep -qi "password"; then SUDO="sudo"; warn "sudo demandera le mot de passe (normal)"; else SUDO="sudo -n"; ok "sudo sans mot de passe"; fi

step "1/6 — paquets (openssh-server, outils invité)"
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
    # apt transient (miroir occupé, stall) : 3 tentatives espacées
    TRY=0
    until [ "$TRY" -ge 3 ]; do
      TRY=$((TRY+1))
      if $SUDO apt-get update -qq && $SUDO apt-get install -y -qq $MISSING; then
        ok "installé :$MISSING (tentative $TRY)"
        break
      fi
      [ "$TRY" -ge 3 ] && fail "apt a échoué 3× — relancez avec internet : sudo apt install -y$MISSING"
      warn "apt tentative $TRY échouée — nouvel essai dans 5 s…"
      sleep 5
    done
  fi
else
  warn "sauté (pas d'apt ou pas d'internet) — vérifiez à la main : sshd ? ip ?"
fi

step "2/6 — serveur SSH"
if systemctl is-active --quiet ssh 2>/dev/null || systemctl is-active --quiet sshd 2>/dev/null; then
  ok "sshd actif"
else
  if ! command -v sshd >/dev/null 2>&1; then fail "openssh-server absent ET hors-ligne — installez-le d'abord (internet requis une fois)"; fi
  $SUDO systemctl enable --now ssh || fail "démarrage sshd impossible"
  ok "sshd démarré + activé au boot"
fi
ss -ltn 2>/dev/null | grep -q ':22 ' && ok "port 22 en écoute" || warn "port 22 non vu en écoute — vérifiez le pare-feu (ufw allow ssh)"
# pare-feu actif ? ouvre 22 + dashboard (sinon le host ne joindra jamais).
# $SUDO (jamais sudo nu) : en mode -n, échec silencieux → message manuel.
if command -v ufw >/dev/null 2>&1 && $SUDO ufw status 2>/dev/null | grep -q "Status: active"; then
  for p in 22 "$PORT"; do
    if $SUDO ufw status 2>/dev/null | grep -q "$p/tcp.*ALLOW"; then
      ok "ufw : $p/tcp autorisé"
    else
      $SUDO ufw allow "$p/tcp" >/dev/null 2>&1 && ok "ufw : $p/tcp ouvert à l'instant" || warn "ufw actif mais $p/tcp fermé — à la main : sudo ufw allow $p/tcp"
    fi
  done
fi

step "3/6 — binaire"
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

step "4/6 — dashboard sécurisé (TLS, port $PORT)"
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

step "5/6 — banc de mesure (campagnes)"
if [ -f "$PROJECT/kit/testbed.sh" ]; then
  bash "$PROJECT/kit/testbed.sh" up || warn "testbed up a échoué — campagnes impossibles tant que le banc est absent"
else
  warn "kit/testbed.sh absent — récupérez-le du zip (dossier kit/) puis : sudo bash kit/testbed.sh up"
fi

step "6/6 — santé + récapitulatif"
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
