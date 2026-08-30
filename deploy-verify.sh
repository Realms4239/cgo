#!/usr/bin/env bash
set -e
cd /c/cgo/.worktrees/polish
bash kit/engine.sh --action deploy --config kit/cgo-vm.yaml
cd web/frontend
npx playwright test e2e/chart-probe.spec.ts e2e/compare-probe.spec.ts e2e/control-demo.spec.ts e2e/archives.spec.ts e2e/deep-probe.spec.ts --workers=1
