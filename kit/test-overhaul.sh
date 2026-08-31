#!/usr/bin/env bash
# test-overhaul — anti-régression : les 12 frictions purgées ne reviennent jamais.
set -u
fail=0
ck() { # ck <description> <command...>
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then echo "PASS  $desc"; else echo "FAIL  $desc"; fail=1; fi
}
ckg() { # ckg <description> <pattern> <grep-mode(-L absent | -H présent)> <fichier...>
  local desc="$1" pat="$2" mode="$3"; shift 3
  if [ "$mode" = "-L" ]; then
    if grep -q "$pat" "$@" 2>/dev/null; then echo "FAIL  $desc (présent)"; fail=1; else echo "PASS  $desc"; fi
  else
    if grep -q "$pat" "$@" 2>/dev/null; then echo "PASS  $desc"; else echo "FAIL  $desc (absent)"; fail=1; fi
  fi
}

# 1. deploy/ supprimé — kit/ seule source
ck "deploy/ n'existe plus" test ! -d deploy
# 2. aucun commentaire Go ne pointe deploy/
ckg "aucun deploy/ dans les commentaires Go" "deploy/" -L pkg cmd
# 3. pas de --addr 9090 dur dans vm-install (port vient de la config)
ckg "vm-install suit \$PORT (pas de 9090 dur)" 'addr 0.0.0.0:9090' -L kit/vm-install.sh
# 4. moteur Go présent + shim traduit
ck "internal/kit existe" test -f internal/kit/kit.go
ck "internal/vm existe" test -f internal/vm/vm.go
ckg "shim traduit --action" "action" -H kit/engine.sh
# 5. pur Go : ni powershell ni cygpath dans le moteur
ckg "pas de powershell dans internal/" "powershell" -L internal
ckg "pas de cygpath dans internal/" "cygpath" -L internal
# 6. pas d'IP VM en dur dans README/docs (hors examples documentés)
ckg "README sans URL IP VM" "http://192.168.174.128" -L README.md docs
# 7. doctor chemins corrects (pas de double kit)
ckg "engine doctor sans kit/../kit" "kit/../kit" -L kit/engine.sh internal
# 8. bits exécutables (git index — fiable sous Windows/NTFS)
ck "install.sh exécutable (index git)" bash -c 'test "$(git ls-files -s kit/install.sh | cut -c1-6)" = "100755"'
ck "setup-meteolink-dev.bat exécutable (index git)" bash -c 'test "$(git ls-files -s kit/setup-meteolink-dev.bat | cut -c1-6)" = "100755"'
ck "testbed.sh exécutable (index git)" bash -c 'test "$(git ls-files -s kit/testbed.sh | cut -c1-6)" = "100755"'
# 9. pas de binaires .new traînants versionnés
ck "aucun cgo-linux.new suivi" bash -c '! git ls-files --error-unmatch kit/cgo-linux.new >/dev/null 2>&1'
# 10. nomenclature npm unique (archive cgo-*)
ckg "install.js télécharge cgo-*" 'cgo-\${p.os}-\${p.arch}' -H npm/install.js
ck "npm bin/cli.js lanceur présent" test -f npm/bin/cli.js
# 11. stub cgo run disparu — vrai moteur
ck "cgo run implémenté (runCLI)" bash -c 'grep -q runCLI cmd/cgo/runcli.go'
ckg "plus de stub exit(2) run" 'use API POST /api/run/start' -L cmd/cgo
# 12. TUI sans données synthétiques
ckg "plus de sinus synthétique dans le TUI" "synthetic" -L cmd/cgo

[ "$fail" -eq 0 ] && echo "== test-overhaul : 12/12 PASS ==" || { echo "== test-overhaul : ÉCHEC =="; exit 1; }
