# CGO — BBRv3 evaluation matrix — implementation plan

*2026-10-27 · Branch: `rebuild/plan1-economics-audit`*
*Spec: `docs/superpowers/specs/2026-10-27-cgo-bbrv3-matrix-design.md`*

## Goal

Add BBRv3 as an additive, kernel-gated congestion-control label, freeze the
kernel release with every new matrix, fail closed when a download socket
refuses its CC, and ship an opt-in kernel builder for the 7.1 BBRv3 branch.

## Constraints

- `bbr` stays BBRv1. `bbr3` maps 1:1 to the socket name `bbr3`.
- Stock 6.8 keeps the two-CC default; no BBRv3 is promised there.
- No frozen corpus edits. New runs under `data/runs`, never `book/data/runs`.
- No CSV schema or column-order change.
- No runtime compilation by the cgo binary. The kernel build is an explicit
  operator action.
- Windows proves the engine and gates only. Live acceptance needs the VM.
- French comments and user-facing messages; English identifiers.
- Commit per logical group with explicit paths. Never `git add -A`.

## Proof boundary

Windows can prove: parsing, default selection, explicit gating, config
provenance, doctor output, and command construction. It cannot prove that a
7.1 kernel boots, that `/proc` lists `bbr3`, or that a socket uses it. The
slice stays **Windows-green** until Task 8 runs on the VM.

---

## Task 1: model — `bbr3` label and CC list parser

**Files:**
- Modify: `pkg/model/types.go`
- Create: `pkg/model/cc.go`
- Create: `pkg/model/cc_test.go`

**Interfaces:**
- Produces: `model.BBRv3`, `model.KnownCC`, `model.ParseCCList(string) []string`.
- Keeps: `model.AllCC` unchanged.

- [ ] **Step 1: Add the label**

In `pkg/model/types.go`, replace the CC block:

```go
type CC string

const (
	Cubic CC = "cubic"
	BBR   CC = "bbr"  // BBRv1 — décision #8 : « bbr » ne veut jamais dire v3.
	BBRv3 CC = "bbr3" // BBRv3 — seulement si le noyau l'expose.
)

// AllCC — défaut historique de la matrice pleine ; bbr3 n'y entre que si le
// noyau l'enregistre (campagne.DefaultCCs).
var AllCC = []CC{Cubic, BBR}

// KnownCC — étiquettes acceptées. bbr3 reste gatée à l'exécution : une
// demande explicite sur un noyau sans bbr3 est refusée, jamais repliée.
var KnownCC = []CC{Cubic, BBR, BBRv3}
```

- [ ] **Step 2: Add the parser**

Create `pkg/model/cc.go`:

```go
package model

import "strings"

// ParseCCList tokenise la liste du noyau
// (/proc/sys/net/ipv4/tcp_available_congestion_control).
// L'ordre du noyau est conservé ; vides et doublons sont écartés.
func ParseCCList(raw string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range strings.Fields(raw) {
		if seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}
```

- [ ] **Step 3: Test it**

Create `pkg/model/cc_test.go`:

```go
package model

import (
	"reflect"
	"testing"
)

func TestParseCCList(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"reno cubic bbr\n", []string{"reno", "cubic", "bbr"}},
		{"  bbr3   bbr bbr3 ", []string{"bbr3", "bbr"}},
		{"", nil},
	}
	for _, c := range cases {
		if got := ParseCCList(c.in); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("ParseCCList(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBBRv3IsKnownButNotDefault(t *testing.T) {
	known := false
	for _, c := range KnownCC {
		if c == BBRv3 {
			known = true
		}
	}
	if !known {
		t.Fatal("KnownCC sans bbr3")
	}
	for _, c := range AllCC {
		if c == BBRv3 {
			t.Fatal("AllCC ne doit pas inclure bbr3 par défaut")
		}
	}
}
```

- [ ] **Step 4: Run the focused tests**

```bash
go test ./pkg/model -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/model/types.go pkg/model/cc.go pkg/model/cc_test.go
git commit -m "model: étiquette bbr3 et lecture de la liste CC du noyau"
```

---

## Task 2: doctor — separate `bbr` and `bbr3`

**Files:**
- Modify: `pkg/doctor/doctor.go`
- Create: `pkg/doctor/cc_test.go`

**Interfaces:**
- Consumes: `model.ParseCCList`.
- Produces: `ccChecks(mode string) []Check`, `ccChecksFrom(mode string, list []string, readable bool) []Check`.
- Removes the substring test `strings.Contains(raw, "bbr")`.

- [ ] **Step 1: Replace the bbr block**

In `pkg/doctor/doctor.go`, replace the current `bbr := "indéterminé"` block
through `checks = append(checks, Check{Name: "bbr", ...})` with:

```go
	checks = append(checks, ccChecks(mode)...)
```

Add the import:

```go
	"github.com/Realms4239/cgo/pkg/model"
```

Add the helpers at the end of the file:

```go
// ccChecks — deux lignes distinctes : « bbr » (v1, mainline) et « bbr3 »
// (noyau BBRv3). Le token exact, jamais le préfixe : contains("bbr")
// déclarait bbr disponible sur un noyau qui n'expose que bbr3.
func ccChecks(mode string) []Check {
	raw, err := os.ReadFile("/proc/sys/net/ipv4/tcp_available_congestion_control")
	if err != nil {
		return ccChecksFrom(mode, nil, false)
	}
	return ccChecksFrom(mode, model.ParseCCList(string(raw)), true)
}

func ccChecksFrom(mode string, list []string, readable bool) []Check {
	if !readable {
		return []Check{
			{Name: "bbr", Status: "warn", Detail: "indéterminé — /proc/sys/net/ipv4/tcp_available_congestion_control illisible"},
			{Name: "bbr3", Status: "warn", Detail: "indéterminé — même fichier illisible"},
		}
	}
	present := map[string]bool{}
	for _, c := range list {
		present[c] = true
	}
	bbr := Check{Name: "bbr", Status: "ok", Detail: "disponible (BBRv1, mainline)"}
	if !present["bbr"] {
		bbr = Check{Name: "bbr", Status: "warn", Detail: "absent — modprobe tcp_bbr (kit/install.sh)"}
		if mode == "full" {
			bbr.Status = "fail"
		}
	}
	bbr3 := Check{Name: "bbr3", Status: "warn", Detail: "absent — optionnel : kit/install-bbr3.sh (noyau BBRv3, 7.1)"}
	if present["bbr3"] {
		bbr3 = Check{Name: "bbr3", Status: "ok", Detail: "disponible (BBRv3)"}
	}
	return []Check{bbr, bbr3}
}
```

`strings` remains used elsewhere in the file.

- [ ] **Step 2: Test the separation**

Create `pkg/doctor/cc_test.go`:

```go
package doctor

import "testing"

func TestCCChecksSeparatesBBRVersions(t *testing.T) {
	got := ccChecksFrom("full", []string{"reno", "cubic", "bbr"}, true)
	if got[0].Name != "bbr" || got[0].Status != "ok" {
		t.Fatalf("bbr v1 : %+v", got[0])
	}
	if got[1].Name != "bbr3" || got[1].Status != "warn" {
		t.Fatalf("bbr3 absent doit rester warn : %+v", got[1])
	}

	got = ccChecksFrom("full", []string{"reno", "cubic", "bbr3"}, true)
	if got[0].Status != "fail" {
		t.Fatalf("bbr v1 absent en mode full doit échouer : %+v", got[0])
	}
	if got[1].Status != "ok" {
		t.Fatalf("bbr3 présent : %+v", got[1])
	}

	got = ccChecksFrom("full", nil, false)
	if got[0].Status != "warn" || got[1].Status != "warn" {
		t.Fatalf("fichier illisible : %+v", got)
	}
}
```

- [ ] **Step 3: Run the focused tests**

```bash
go test ./pkg/doctor -count=1
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add pkg/doctor/doctor.go pkg/doctor/cc_test.go
git commit -m "doctor: séparer bbr (v1) et bbr3 par token exact"
```

---

## Task 3: campagne — availability, defaults, gating, kernel provenance

**Files:**
- Create: `pkg/campagne/ccavail.go`
- Create: `pkg/campagne/ccavail_linux.go`
- Create: `pkg/campagne/ccavail_other.go`
- Create: `pkg/campagne/main_test.go`
- Create: `pkg/campagne/ccavail_test.go`
- Modify: `pkg/campagne/matrix.go`
- Modify: `pkg/campagne/config.go`

**Interfaces:**
- Produces: `campagne.DefaultCCs() []string`, `campagne.ccPresent(string) bool`.
- Vars for tests: `readCCList`, `readKernelRelease`.
- `matrixConfig.Kernel string`.

- [ ] **Step 1: Add the availability layer**

Create `pkg/campagne/ccavail.go`:

```go
package campagne

import "github.com/Realms4239/cgo/pkg/model"

// readCCList et readKernelRelease sont des variables : les tests les
// remplacent pour rester indépendants de l'hôte (un noyau BBRv3 ne doit pas
// changer le nombre de lignes attendu par les tests).
var (
	readCCList        = readHostCCList
	readKernelRelease = readHostKernelRelease
)

// ccPresent — le noyau expose-t-il exactement ce nom ? Le token, pas le
// préfixe : « bbr » ne doit jamais se lire dans « bbr3 ».
func ccPresent(cc string) bool {
	for _, c := range readCCList() {
		if c == cc {
			return true
		}
	}
	return false
}

// DefaultCCs — matrice pleine : cubic et bbr (v1) toujours, bbr3 seulement
// si le noyau l'enregistre. Une demande explicite de bbr3 reste gatée dans
// startMatrix : l'omission silencieuse n'est permise que pour le défaut.
func DefaultCCs() []string {
	out := make([]string, 0, len(model.AllCC)+1)
	for _, c := range model.AllCC {
		out = append(out, string(c))
	}
	if ccPresent(string(model.BBRv3)) {
		out = append(out, string(model.BBRv3))
	}
	return out
}
```

Create `pkg/campagne/ccavail_linux.go`:

```go
//go:build linux

package campagne

import (
	"os"
	"strings"

	"github.com/Realms4239/cgo/pkg/model"
)

func readHostCCList() []string {
	raw, err := os.ReadFile("/proc/sys/net/ipv4/tcp_available_congestion_control")
	if err != nil {
		return nil
	}
	return model.ParseCCList(string(raw))
}

func readHostKernelRelease() string {
	raw, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}
```

Create `pkg/campagne/ccavail_other.go`:

```go
//go:build !linux

package campagne

// Hors Linux, ni /proc ni les qdiscs : la campagne dualns n'y tourne pas.
func readHostCCList() []string      { return nil }
func readHostKernelRelease() string { return "" }
```

- [ ] **Step 2: Use the shared default in the matrix**

In `pkg/campagne/matrix.go`:

Replace both `allCCStrings()` calls with `DefaultCCs()` and delete
`allCCStrings`:

```go
	return startMatrix(base, newRunID(), profiles, allQdiscStrings(), DefaultCCs(), reps, deps, dataDir)
```

```go
	return startMatrix(base, runID, profiles, allQdiscStrings(), DefaultCCs(), reps, deps, dataDir)
```

Replace the CC validation block with:

```go
	validCC := map[string]bool{}
	for _, c := range model.KnownCC {
		validCC[string(c)] = true
	}
	for _, c := range ccs {
		if !validCC[c] {
			return nil, fmt.Errorf("cc inconnu: %s", c)
		}
		if c == string(model.BBRv3) && !ccPresent(c) {
			return nil, fmt.Errorf("cc bbr3 indisponible sur cet hôte : le noyau n'expose pas bbr3 (kit/install-bbr3.sh) — refus plutôt qu'exécuter un autre contrôle sous cette étiquette")
		}
	}
```

- [ ] **Step 3: Freeze the kernel release**

In `pkg/campagne/config.go`, add the field:

```go
	Kernel string `json:"kernel,omitempty"`
```

and set it in `buildMatrixConfig`:

```go
		Kernel: readKernelRelease(),
```

- [ ] **Step 4: Make the test suite host-independent**

Create `pkg/campagne/main_test.go`:

```go
package campagne

import (
	"os"
	"testing"
)

// TestMain — les tests de matrice comptent des lignes ; un hôte BBRv3 ne
// doit pas changer le défaut. Les tests qui veulent bbr3 le stubbent.
func TestMain(m *testing.M) {
	readCCList = func() []string { return []string{"cubic", "bbr"} }
	os.Exit(m.Run())
}
```

- [ ] **Step 5: Test availability and gating**

Create `pkg/campagne/ccavail_test.go`:

```go
package campagne

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/model"
)

func stubCCs(t *testing.T, list ...string) {
	t.Helper()
	old := readCCList
	readCCList = func() []string { return list }
	t.Cleanup(func() { readCCList = old })
}

func TestDefaultCCsWithoutBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	got := DefaultCCs()
	if len(got) != 2 || got[0] != "cubic" || got[1] != "bbr" {
		t.Fatalf("défaut sans bbr3 : %v", got)
	}
}

func TestDefaultCCsWithBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3")
	got := DefaultCCs()
	if len(got) != 3 || got[2] != "bbr3" {
		t.Fatalf("défaut avec bbr3 : %v", got)
	}
}

func TestExplicitBBRv3RefusedWithoutKernel(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	_, err := StartMatrixFiltered(context.Background(), []string{"P2"},
		[]string{"cake"}, []string{"bbr3"}, 1, fastDeps(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "install-bbr3") {
		t.Fatalf("bbr3 explicite sans noyau doit refuser : %v", err)
	}
}

func TestExplicitBBRv3AcceptedWhenPresent(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3")
	m, err := StartMatrixFiltered(context.Background(), []string{"P2"},
		[]string{"cake"}, []string{"bbr3"}, 1, fastDeps(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40 && m.Done != m.Total; i++ {
		time.Sleep(50 * time.Millisecond)
	}
	if m.Total != 1 {
		t.Fatalf("total = %d, want 1", m.Total)
	}
}

func TestKernelReleaseIsFrozen(t *testing.T) {
	old := readKernelRelease
	readKernelRelease = func() string { return "7.1.0-rc5-cgo-bbr3-test" }
	t.Cleanup(func() { readKernelRelease = old })
	cfg, err := buildMatrixConfig([]string{"P1"}, []string{"cake"}, []string{string(model.BBRv3)}, 1, Deps{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), `"kernel":"7.1.0-rc5-cgo-bbr3-test"`) {
		t.Fatalf("kernel absent du gel : %s", cfg)
	}
}
```

- [ ] **Step 6: Run the focused tests**

```bash
go test ./pkg/campagne -count=1
```

Expected: PASS. The existing matrix tests keep their 18-row expectation
because `TestMain` stubs the host list.

- [ ] **Step 7: Commit**

```bash
git add pkg/campagne/ccavail.go pkg/campagne/ccavail_linux.go pkg/campagne/ccavail_other.go \
  pkg/campagne/main_test.go pkg/campagne/ccavail_test.go pkg/campagne/matrix.go pkg/campagne/config.go
git commit -m "campagne: gater bbr3 sur le noyau et geler sa version"
```

---

## Task 4: cmd — shared default and fail-closed download CC

**Files:**
- Modify: `cmd/cgo/runcli.go`
- Modify: `cmd/cgo/server.go`
- Modify: `cmd/cgo/testbedsrv.go`

**Interfaces:**
- Consumes: `campagne.DefaultCCs`.
- Produces: no new exported surface.

- [ ] **Step 1: Share the default in the CLI**

In `cmd/cgo/runcli.go`, replace:

```go
	if ccs == nil {
		cs := make([]string, len(model.AllCC))
		for i, c := range model.AllCC {
			cs[i] = string(c)
		}
		ccs = cs
	}
```

with:

```go
	if ccs == nil {
		ccs = campagne.DefaultCCs()
	}
```

`model` remains used for the qdisc axis.

- [ ] **Step 2: Share the default in the dashboard start**

In `cmd/cgo/server.go`, replace the same CC block with:

```go
		if len(ccs) == 0 {
			ccs = campagne.DefaultCCs()
		}
```

- [ ] **Step 3: Fail closed when the source refuses its CC**

In `cmd/cgo/testbedsrv.go`, replace:

```go
					if err := probe.SetConnCC(c, cc); err != nil {
						log.Printf("[src] CC %q refusée: %v (défaut hôte)", cc, err)
					}
```

with:

```go
					if err := probe.SetConnCC(c, cc); err != nil {
						log.Printf("[src] CC %q refusée: %v — source fermée, cellule vouée à l'échec (pas de repli cubic silencieux)", cc, err)
						return
					}
```

The deferred `c.Close()` runs; the client reads zero bytes and its gates
fail.

- [ ] **Step 4: Build and test the touched packages**

```bash
go build ./cmd/cgo
go test ./cmd/... -count=1
```

Expected: build OK; `cmd/cgo` has no unit tests, so the test command may
report `no test files`.

- [ ] **Step 5: Commit**

```bash
git add cmd/cgo/runcli.go cmd/cgo/server.go cmd/cgo/testbedsrv.go
git commit -m "cmd: défaut CC partagé et source fermée si la CC est refusée"
```

---

## Task 5: `kit/install-bbr3.sh` — opt-in kernel builder

**Files:**
- Create: `kit/install-bbr3.sh`

**Interfaces:**
- Consumes: Debian/Ubuntu, `apt-get`, `git`, root or `sudo -n`.
- Produces: a bootable kernel with `bbr` (v1) and `bbr3` (v3) built in;
  `/var/lib/cgo/bbr3-kernel.json`.

- [ ] **Step 1: Write the script**

Create `kit/install-bbr3.sh`:

```bash
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
```

- [ ] **Step 2: Syntax-check**

```bash
bash -n kit/install-bbr3.sh
```

Expected: no output.

- [ ] **Step 3: Static inspection**

```bash
grep -n "bbr3\|BRANCH\|COMMIT\|CONFIG_TCP_CONG\|CONFIG_LOCALVERSION\|reboot" kit/install-bbr3.sh
```

Expected: the pinned commit, the forced options, the manual reboot, and no
automatic call to `cgo run`.

- [ ] **Step 4: Commit**

```bash
git add kit/install-bbr3.sh
git commit -m "kit: constructeur de noyau BBRv3 opt-in"
```

---

## Task 6: help text, methodology, roadmap

**Files:**
- Modify: `cmd/cgo/main.go`
- Modify: `docs/methodology.fr.md`
- Modify: `docs/superpowers/plans/README.md`

**Interfaces:**
- Consumes: Tasks 1–5.
- Produces: operator-facing documentation and the roadmap status.

- [ ] **Step 1: Update the `--cc` help**

In `cmd/cgo/main.go`, change:

```go
		ccs := fs.String("cc", "", "contrôles de congestion (cubic,bbr) — vide = tous")
```

to:

```go
		ccs := fs.String("cc", "", "contrôles de congestion (cubic,bbr,bbr3) — vide = tous ; bbr3 si le noyau l'expose")
```

- [ ] **Step 2: Add the methodology section**

Append to `docs/methodology.fr.md`:

```markdown
## 10. BBRv3 : noyau, étiquette et matrice 3×3

La section 9 est réservée au façonnage descendant IFB (plan distinct). Cette
section décrit la troisième ligne de contrôle du plan factoriel.

### 10.1 Pourquoi un noyau

BBRv3 n'est pas un module de contrôle de congestion : Google le publie comme
une arborescence de noyau complète (`google/bbr`), avec des changements dans
le cœur TCP. Le noyau 6.8 d'Ubuntu n'expose que `bbr`, qui est **BBRv1**.
La décision #8 du projet reste donc vraie : `bbr` signifie BBRv1. BBRv3
entre sous une étiquette additives, `bbr3`, et seulement sur un hôte dont le
noyau enregistre ce nom.

### 10.2 Provisionnement opt-in

`kit/install-bbr3.sh` construit la branche `bbr-v3-2026-09-16-01` (Linux 7.1)
sur la VM, avec `CONFIG_TCP_CONG_BBR=y` et `CONFIG_TCP_CONG_BBR3=y`. Les deux
contrôles vivent alors dans le même noyau : la comparaison v1/v3 ne change
qu'une variable. Le script n'est jamais appelé par `cgo run`, `cgo deploy` ou
`cgo kit bootstrap` ; il s'exécute sur la console ou via `cgo kit ssh`. Le
redémarrage est manuel par défaut. Le noyau précédent reste dans GRUB.

### 10.3 Matrice

```bash
./kit/run-dualns.sh up
sudo ip netns exec ns_client env CGO_BENCH=dualns ~/cgo/cgo-linux run \
  --bench dualns --direction up --profiles P2 \
  --qdiscs pfifo_fast,fq_codel,cake \
  --cc cubic,bbr,bbr3 \
  --reps 3 --flows 4 --target 10.30.0.2
```

Neuf cellules, trois répétitions, 27 événements. Le sens montant reproduit
la campagne du mémoire. Ce jeu de runs est **autonome** : il ne se fusionne
pas numériquement avec les cellules du corpus 6.8 ni avec le Tableau 15,
qui viennent d'un autre banc et d'un autre régime de flux. Chaque manifeste
gele le noyau (`kernel`) avec la configuration ; une ligne `cc=bbr3` n'est
citable qu'avec ce noyau.

### 10.4 Portée

La matrice n'active ni ECN ni L4S : les cellules comparent le comportement
sous perte et délai, pas les signaux ECN de BBRv3. Le façonnage descendant
reste celui du slice 6 ; la matrice ne dépend pas de l'IFB.
```

- [ ] **Step 3: Update the roadmap**

In `docs/superpowers/plans/README.md`:

1. Change slice 7's row to:

```markdown
| 7 | BBRv3 evaluation matrix | `2026-10-27-cgo-bbrv3-matrix-design.md` | `2026-10-27-cgo-bbrv3-matrix.md` | **Proposed** | opt-in 7.1 BBRv3 kernel on the VM | User approval, then Windows-green Tasks 1–7; Task 8 blocked until the kernel is built and booted |
```

2. Replace the "Slices 7–8 — not planned yet" BBRv3 bullet with a scope
   paragraph: additive `bbr3`, exact-token gating, kernel frozen in the
   manifest, `kit/install-bbr3.sh`, no ECN axis, no UI change.

3. Update the dependency graph:

```
BBRv3 (slice 7) ◄── opt-in 7.1 kernel (kit/install-bbr3.sh)
```

4. Add under standing constraints:

```markdown
- `bbr` means BBRv1; `bbr3` is additive and only exists on a kernel that
  registers it. Never alias the two.
```

- [ ] **Step 4: Commit**

```bash
git add cmd/cgo/main.go docs/methodology.fr.md docs/superpowers/plans/README.md
git commit -m "docs: matrice BBRv3, provisionnement et feuille de route"
```

---

## Task 7: Windows gates

**Files:** none (verification only).

- [ ] **Step 1: Format and vet**

```bash
gofmt -l pkg/model pkg/doctor pkg/campagne cmd/cgo
go vet ./...
```

Expected: `gofmt -l` prints nothing; vet exits 0.

- [ ] **Step 2: Full Go tests**

```bash
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Linux cross-build**

```bash
GOOS=linux go build ./...
```

Expected: exit 0. This compiles the `ccavail_linux.go` branch that Windows
tests do not.

- [ ] **Step 4: Shell syntax**

```bash
bash -n kit/install-bbr3.sh
```

Expected: no output.

- [ ] **Step 5: Frozen corpus guard**

```bash
python book/tools/audit-archive.py
```

Expected: 159 runs, 528 rows, 208 valid, 74 degraded, 246 invalid. Say
"208 valid lines", never "208 runs".

- [ ] **Step 6: Record the pass**

Update the roadmap status to **Windows-green** with the exact commands and
results. Do not write **Verified**.

- [ ] **Step 7: Commit the status**

```bash
git add docs/superpowers/plans/README.md
git commit -m "docs: portes Windows vertes pour la matrice BBRv3"
```

---

## Task 8: live acceptance — BLOCKED

**Files:** none in the repo; evidence goes under `data/runs/`.

**Blocker:** no BBRv3 kernel has been built on the bench VM yet, and the
VM must be stable for the build and reboot. Windows cannot run this task.

- [ ] **Step 1: Snapshot the VM**

```bash
cgo kit snapshot bbr3-pre-kernel
```

- [ ] **Step 2: Build and install the kernel on the VM**

```bash
cgo kit ssh
# dans l'invitée :
cd ~/cgo
sudo bash kit/install-bbr3.sh
sudo reboot
```

- [ ] **Step 3: Verify the kernel and both controls**

```bash
uname -r                       # contains -cgo-bbr3
cat /proc/sys/net/ipv4/tcp_available_congestion_control   # bbr ... bbr3
cgo doctor
```

Expected: `bbr` and `bbr3` both `ok`. If `bbr3` is absent, stop; the 3×3
would be a lie.

- [ ] **Step 4: Bring the bench up and run the matrix**

```bash
cd ~/cgo
./kit/run-dualns.sh up
sudo ip netns exec ns_client env CGO_BENCH=dualns ~/cgo/cgo-linux run \
  --bench dualns --direction up --profiles P2 \
  --qdiscs pfifo_fast,fq_codel,cake \
  --cc cubic,bbr,bbr3 \
  --reps 3 --flows 4 --target 10.30.0.2
```

Expected: 27 events, no preflight refusal. The run directory is printed by
the summary.

- [ ] **Step 5: Verify the freeze**

```bash
python - <<'PY'
import csv, glob, json, os
run = sorted(glob.glob('data/runs/*'), key=os.path.getmtime)[-1]
doc = json.load(open(os.path.join(run, 'manifest.json')))
cfg = json.loads(doc['config'])
rows = list(csv.DictReader(open(os.path.join(run, 'aqm_eval.csv'))))
print("run:", os.path.basename(run))
print("kernel:", cfg.get('kernel'))
print("ccs:", cfg.get('ccs'))
print("rows:", len(rows), "cc labels:", sorted({r['cc'] for r in rows}))
per = {}
for r in rows:
    per.setdefault(r['qdisc'], set()).add(r['cc'])
print("per qdisc:", {k: sorted(v) for k, v in per.items()})
PY
```

Expected: `kernel` ends in `-cgo-bbr3`; `ccs` is `["cubic","bbr","bbr3"]`;
27 rows; every qdisc has all three labels.

- [ ] **Step 6: Inspect one bbr3 row against its bbr twin**

```bash
grep ',bbr3,' data/runs/<run-id>/aqm_eval.csv | head
grep ',bbr,'  data/runs/<run-id>/aqm_eval.csv | head
```

Record the exact values and gate statuses. Do not claim BBRv3 reduces latency
until the gates and the repetitions say so.

- [ ] **Step 7: Record the evidence**

Add the commands, the run id, the kernel, the 27-row check, and the gate
summary to the roadmap. Status stays **Windows-green + live 3×3 produced**;
only the thesis chapter may call the comparison "measured".

---

## Self-review

- **Spec coverage:** D1 (Task 1), D2 (Tasks 1, 3), D3 (Tasks 1, 2, 3),
  D4 (Task 3), D5 (Task 3), D6 (Task 4), D7 (Task 3), D8 (Tasks 6, 8),
  D9 (Task 6), D10 (Task 5), D11 (Tasks 5, 6), D12 (Task 5), D13 (Task 2),
  D14 (no task — no UI change), D15 (Tasks 7, 8).
- **No placeholder:** the only blocked step is Task 8, named with its
  blocker.
- **Consistency:** `bbr` is never redefined; `bbr3` is always exact-token;
  the kernel is frozen before any cell; Windows never claims the live path.
- **Scope:** no ECN axis, no IFB, no `.dist`, no eBPF, no UI, no frozen
  corpus edit.
