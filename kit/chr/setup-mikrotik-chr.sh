#!/usr/bin/env bash
# Wrapper portable (Git Bash et Linux) autour de cmd/cgo-chr. Il ne juge rien :
# il localise le dépôt et transmet les arguments. Toute la logique est en Go
# pour rester testable sans VM.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
cd "$root"
exec go run ./cmd/cgo-chr "$@"
