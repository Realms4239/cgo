# Wall-kit 4 — Kit (Build → Scan → Bootstrap → Ensure → Deploy → Tunnel → npm) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generalize ops kit — build-then-embed strict, scan C/D VMware/VBox detect-fallback, bootstrap veth, ensure SSH, deploy binary-first + Cloudflare tunnel, npm/brew TUI.

**Architecture:** kit/engine.sh single entrypoint hypervisor auto, always build before deploy, embed analysis dist sha==buildSha==integrity sha8, postinstall ponytail.

**Tech Stack:** Bash kit/engine.sh, vmrun, VBoxManage, cloudflared, Go cross-compile, bun build, npm postinstall.js

## Global Constraints

- kit/ not deploy/, hypervisor auto vmware|virtualbox
- --action scan C:/D:/ depth≤3 *.vmx/*.vbox --deep full
- --action ensure poll SSH 60×5s, --action bootstrap iproute2 curl bc tcp_bbr veth-c/veth-s parent 1:
- --action build always before deploy, TOTAL 600 echarts 350 guard
- --public tunnel via $CLOUDFLARE_TUNNEL_TOKEN not committed, npm postinstall ponytail minimal
- Inline not subagents

---

### Task 4.1: Kit Rename deploy→kit + Build-then-Embed Strict

**Files:**
- Create: `kit/engine.sh` (rename from deploy/engine.sh)
- Create: `kit/cgo-vm.yaml.example`
- Modify: `kit/config.sh`
- Test: `kit/test-kit.sh` build guard

- [ ] **Step 1: Write failing test**

```bash
test -f kit/engine.sh && grep -q "TOTAL 600" kit/engine.sh
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash kit/test-kit.sh`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```bash
# kit/engine.sh --action build: go vet && bun run typecheck && bun run build && node scripts/check-bundle.mjs || exit 1; go build -o bin/cgo-linux
# scan: vmrun list + VBoxManage list vms + C:/D:/ depth≤3 *.vmx/*.vbox
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bash kit/test-kit.sh`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add kit/
git commit -m "feat(kit): rename deploy→kit + build-then-embed strict 600/350"
```

### Task 4.2: VMware Detect-fallback VirtualBox

**Files:**
- Modify: `kit/engine.sh:scan`
- Test: `kit/test-hypervisor.sh`

- [ ] **Step 1: Write failing test**

```bash
grep -q "VBoxManage" kit/engine.sh || exit 1
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash kit/test-hypervisor.sh`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```bash
if command -v vmrun &>/dev/null; then vmrun list; elif command -v VBoxManage &>/dev/null; then VBoxManage list vms; fi
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bash kit/test-hypervisor.sh`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add kit/engine.sh
git commit -m "feat(kit): VMware detect-fallback VirtualBox"
```

### Task 4.3: Cloudflare Tunnel + npm/brew TUI

**Files:**
- Modify: `kit/engine.sh:deploy` add --public
- Create: `scripts/npm-postinstall.js`
- Modify: `package.json`
- Create: `cmd/meteolink/main.go`
- Test: `web/frontend/src/lib/tui.test.ts`

- [ ] **Step 1: Write failing test**

```ts
describe('tui',()=>{
  it('meteolink top exists',()=>{
    const s=readFileSync('cmd/meteolink/main.go','utf8')
    expect(s).toContain('top')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/tui.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/meteolink/main.go package main flag --tui vs --serve, top renders 8 cards ASCII sparklines 60×12 via live rings
// scripts/npm-postinstall.js detect linux/darwin/win32 x64/arm64 download cgo-linux/macos/win
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/tui.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/meteolink/main.go scripts/npm-postinstall.js package.json
git commit -m "feat(kit): cloudflare tunnel + npm/brew TUI meteolink top"
```
