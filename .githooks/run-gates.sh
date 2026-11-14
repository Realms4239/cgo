#!/bin/sh
# run-gates — build cmd/gates into a throwaway dir and exec it.
# `go run` remaps gate exit 2 to 1 on every host (observed on this one), which
# breaks the « 2 is not green » contract; the built binary keeps exit 0/1/2 exact.
# Hooks, the Makefile and CI call this, never `go run`, for the same reason.
set -eu
root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$root"
command -v go >/dev/null 2>&1 || { echo "gates: go not found on PATH (exit 2)" >&2; exit 2; }
tmp=$(mktemp -d)
bin="$tmp/gates$(go env GOEXE)"
go build -o "$bin" ./cmd/gates
rc=0
"$bin" "$@" || rc=$?
rm -rf "$tmp"
exit $rc
