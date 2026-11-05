# CGO Enterprise Rebuild — Plan 2: Bench + Campaign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the four-namespace bench the production measurement path — fail-closed offloads, four-hop leak proof, management veth with serve-in-`ns_client`; fix the gemodel argument bug; wire the campaign to the dual-ns topology with `flows=4`, `channel_model=clean|bursty`, manifests + `valuation.json` sidecar; doctor offload gate; correct Plan 1's sink address.

**Architecture:** Go packages (`pkg/qdisc`, `pkg/campagne`, `pkg/doctor`, `cmd/cgo`) + two kit scripts. Linux-only behavior goes through existing seams (`TCRunner`, `NsRunner`, runtime checks) so Windows stays green; scripts are syntax-checked here and executed on the CGO-DEV VM in Plan 5.

**Tech Stack:** Go 1.25 stdlib only; bash for kit scripts; existing FakeRunner test seam.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` (§2 invariants, §3 topology, §6 campaign).
- Go 1.25 stdlib only, **no new dependencies**; comments French, identifiers English.
- From repo root `C:\cgo`: `go build ./...` and `go test ./pkg/qdisc/... ./pkg/campagne/... ./pkg/doctor/... ./pkg/metrics/... ./cmd/cgo/...` green on Windows; `bash -n` on every touched script; `GOOS=linux go build ./pkg/... ./cmd/cgo/...` green.
- Frozen-data guard: `book/data/runs`, `pkg/results/runs-frozen.tar.gz` untouched; frozen CSV schema immutable; `book/tools/audit-archive.py` output byte-identical after every task.
- Never fabricate: bursty mode requires P2/P3 (else validation error); unmeasured ⇒ stated.
- Commit after every task with the given message.

---

### Task 1: Gilbert-Elliott — documented form, GE math, profile field cleanup

**Files:**
- Modify: `pkg/qdisc/qdisc.go:62-93` (replace `ApplyNetemBurst`), `pkg/qdisc/netem_burst_test.go`
- Modify: `pkg/model/types.go:21-24` (remove `LossBurstP/R/H/K`), `pkg/campagne/campagne.go:225` (call site)
- Create: `pkg/qdisc/ge.go`, `pkg/qdisc/ge_test.go`

**Interfaces:**
- Produces: `qdisc.GEModel{GoodToBad, BadToGood, LossBad, LossGood float64}`; `qdisc.GEMeanLoss(ge GEModel) float64` (0–100, theoretical mean loss = π_bad·LossBad + (1−π_bad)·LossGood, π_bad = GoodToBad/(GoodToBad+BadToGood), guard sum=0 ⇒ 0); `qdisc.ApplyNetemGE(r TCRunner, iface string, delayMs, jitterMs float64, ge GEModel) error` emitting `qdisc replace dev IF root handle 1: netem delay Dms Jms loss gemodel <p> <r> <1-h> <1-k>` with the FOUR documented arguments (decimals, paper form: p=good→bad, r=bad→good, 1-h=loss-in-bad, 1-k=loss-in-good); `qdisc.UniformLossOrGE` helper used by campagne: `ApplyNetemPhysics(r, iface, delayMs, jitterMs, lossPct float64, ge *GEModel) error` (ge nil ⇒ uniform `loss X%`, matching today's behavior byte-for-byte).
- Consumes (removed): `ApplyNetemBurst` five-arg form and `Profile.LossBurstP/R/H/K` — deleted, all references migrated in this task.

- [ ] **Step 1: Write the failing tests**

`pkg/qdisc/ge_test.go`:

```go
package qdisc

import "testing"

func TestGEMeanLossMatchesPaperModel(t *testing.T) {
	// Exemple travaillé de l'article des implémenteurs (Carry/Briscoe) :
	// p=0.01, r=0.19, 1-h=0.81, 1-k=0.01 → 5 % de perte moyenne en salves.
	got := GEMeanLoss(GEModel{GoodToBad: 0.01, BadToGood: 0.19, LossBad: 0.81, LossGood: 0.01})
	if got < 4.9 || got > 5.1 {
		t.Fatalf("exemple article : %v %% (attendu ≈5)", got)
	}
	if GEMeanLoss(GEModel{}) != 0 {
		t.Fatal("modèle vide : 0, pas une division par zéro")
	}
}

func TestApplyNetemGEFourArguments(t *testing.T) {
	// LA correction centrale : la forme documentée est QUATRE arguments
	// « p r 1-h 1-k » — l'ancienne forme à cinq (p r h 1-r k) est invalide
	// sur Linux et disparaît.
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}
	if err := ApplyNetemGE(f, "w-edge", 100, 15, ge); err != nil {
		t.Fatal(err)
	}
	call := f.Calls[0]
	// repère « gemodel » puis les quatre valeurs dans l'ordre du man.
	idx := -1
	for i, a := range call {
		if a == "gemodel" {
			idx = i
			break
		}
	}
	if idx < 0 || len(call) != idx+5 {
		t.Fatalf("gemodel doit recevoir EXACTEMENT 4 args : %v", call)
	}
	want := []string{"0.001", "0.1", "0.5", "0.0005"}
	for i, w := range want {
		if call[idx+1+i] != w {
			t.Fatalf("arg %d = %q, want %q (ordre p, r, 1-h, 1-k)", i+1, call[idx+1+i], w)
		}
	}
}

func TestApplyNetemPhysicsUniformUnchanged(t *testing.T) {
	// ge nil ⇒ la forme uniforme historique, à l'octet près.
	f := &FakeRunner{}
	if err := ApplyNetemPhysics(f, "w-edge", 100, 15, 0.5, nil); err != nil {
		t.Fatal(err)
	}
	call := f.Calls[0]
	joined := strings.Join(call, " ")
	if !strings.Contains(joined, "loss 0.5%") {
		t.Fatalf("forme uniforme attendue : %v", call)
	}
}
```

(Add `"strings"` to imports.) Extend `netem_burst_test.go`: delete the `H=200` test (`TestApplyNetemBurstLoss` — invalid semantics), replace with a `ApplyNetemPhysics` variant passing `&GEModel{...}` asserting the gemodel branch.

- [ ] **Step 2: Run to verify failure** — `go test ./pkg/qdisc/ -run 'TestGE|TestApplyNetem' -v` → FAIL `undefined: GEModel`.

- [ ] **Step 3: Implement** `ge.go` with `GEModel`, `GEMeanLoss` (returns percent), `ApplyNetemGE`, `ApplyNetemPhysics`; delete `ApplyNetemBurst` from qdisc.go; migrate `campagne.go:225` (`qdisc.ApplyNetemBurst(d.TC, d.CliIf, prof.DelayMs, prof.JitterMs, prof.LossPct, prof.LossBurstP, ...)` → the `ApplyNetemPhysics(d.TC, d.CliIf, prof.DelayMs, prof.JitterMs, prof.LossPct, nil)` call — the campaign-level bursty wiring arrives in Task 5); delete the four `LossBurst*` fields from `model.Profile` (grep confirms campagne.go was the only reader).

- [ ] **Step 4: Verify** — `go test ./pkg/qdisc/ ./pkg/campagne/ ./pkg/model/ -count=1` PASS; `grep -rn "ApplyNetemBurst\|LossBurstP" --include="*.go" .` empty.

- [ ] **Step 5: Commit** — `git add pkg/qdisc/ pkg/model/ pkg/campagne/ && git commit -m "qdisc: gemodel — forme documentée à 4 arguments (p r 1-h 1-k), perte moyenne théorique, champs profil périmés supprimés"`

---

### Task 2: Plan 1 correction — dual-ns sink address is 10.30.0.2

**Files:**
- Modify: `cmd/cgo/auditsinks.go`, `cmd/cgo/auditsinks_test.go`

**Interfaces:** none new; `ResolveSinks()` dualns branch corrected.

- [ ] **Step 1: Update the test first** — in `TestResolveSinksPrefersLiveListener`, the dualns case becomes:

```go
	case "dualns":
		// Le puits vit DANS ns_server (s-edge 10.30.0.2) : cible, petit
		// objet et charge traversent les QUATRE sauts. 10.30.0.1 est e-srv
		// (côté edge) — y terminer mesurerait trois sauts, pas quatre.
		if s.BulkAddr != "10.30.0.2:5201" || s.Target != "10.30.0.2" || s.SmallURL != "http://10.30.0.2:8081/small" {
			t.Fatalf("dualns : %+v", s)
		}
```

- [ ] **Step 2: Verify red** — `go test ./cmd/cgo/ -run TestResolveSinks -count=1` FAIL on the old `.1` values.
- [ ] **Step 3: Fix** `ResolveSinks`'s dualns return to `Target: "10.30.0.2"`, `SmallURL: "http://10.30.0.2:8081/small"`, `BulkAddr: "10.30.0.2:5201"`; same for the test's `TestSinksEnvOrBulkDetected` fixture if it cites the address.
- [ ] **Step 4: Verify green** — `go test ./cmd/cgo/ -count=1` PASS.
- [ ] **Step 5: Commit** — `cmd/cgo: puits dualns = 10.30.0.2 (ns_server) — correction du plan 1, la mesure traverse les quatre sauts`

---

### Task 3: Bench script — fail-closed offloads, counter leak proof, management veth

**Files:**
- Modify: `kit/cgo-bench-dualns.sh`

**Interfaces (script contract other tasks rely on):** `up` dies unless all six interfaces read back offloads=off; `check` additionally proves data-plane traversal via TX counters; a fifth veth pair `m-host` (host, 10.99.0.1/30) ↔ `m-client` (ns_client, 10.99.0.2/30) exists after `up`.

- [ ] **Step 1: `set_offloads` fail-closed** — remove `2>/dev/null || true`; on failure: `die "ethtool -K $OFFLOADS failed on $ns/$iface (feature name absent ou noyau exotique — le banc ne démarre PAS avec des super-paquets possibles)"`.
- [ ] **Step 2: `up()` ends with the read-back gate** — after the CAKE block, before `log "up …"`:

```bash
  # Fail-closed : un offload resté ON fausse drops×1448 et le déficit
  # round-robin — le banc refuse de démarrer, il ne prévient pas.
  for pair in "$NS_CLIENT c-cli" "$NS_WAN c-wan" "$NS_WAN w-edge" "$NS_EDGE e-wan" "$NS_EDGE e-srv" "$NS_SRV s-edge"; do
    set -- $pair
    check_offloads "$1" "$2" || die "offload non vérifié sur $1/$2 — démarrage refusé"
  done
```

- [ ] **Step 3: management veth** — after the three transit pairs and their addressing, add the fifth pair (host side lives in the ROOT namespace):

```bash
  # --- veth de gestion : le tableau de bord sert DEPUIS ns_client -----
  # Le navigateur hôte joint 10.99.0.2:9090 ; le trafic MESURÉ ne passe
  # jamais par ici (c-cli reste la seule route par défaut de ns_client).
  $SUDO ip link add m-host type veth peer name m-client netns "$NS_CLIENT" \
    || die "cannot create mgmt veth"
  $SUDO ip addr add 10.99.0.1/30 dev m-host
  $SUDO ip link set m-host up
  $SUDO ip netns exec "$NS_CLIENT" ip addr add 10.99.0.2/30 dev m-client
  $SUDO ip netns exec "$NS_CLIENT" ip link set m-client up
  $SUDO ip netns exec "$NS_CLIENT" ethtool -K m-client tso off gso off gro off 2>/dev/null || true
```

(m-client offloads: best-effort tolerated — the management link carries NO measured traffic; the fail-closed gate covers only the six transit interfaces.) `down()` also deletes the host side: `ip link del m-host 2>/dev/null || true`.

- [ ] **Step 4: counter-based leak proof in `check()`** — after the route-get assertions:

```bash
  # Preuve data-plane : 5 pings vers le puits doivent INCRÉMENTER les TX
  # de c-cli — un raccourci loopback ne toucherait pas ce compteur.
  local tx0 tx1
  tx0=$($SUDO ip netns exec "$NS_CLIENT" cat /sys/class/net/c-cli/statistics/tx_packets)
  $SUDO ip netns exec "$NS_CLIENT" ping -c5 -W2 10.30.0.2 >/dev/null 2>&1
  tx1=$($SUDO ip netns exec "$NS_CLIENT" cat /sys/class/net/c-cli/statistics/tx_packets)
  if [ $((tx1 - tx0)) -lt 5 ]; then
    echo "[!] leak: pings vers le puits sans traverser c-cli (tx $tx0→$tx1)"; ok=0
  fi
  # La veth de gestion existe et porte son adresse.
  $SUDO ip netns exec "$NS_CLIENT" ip addr show m-client 2>/dev/null | grep -q "10.99.0.2/30" \
    || { echo "[!] veth de gestion absente (m-client/10.99.0.2)"; ok=0; }
```

- [ ] **Step 5: Verify + commit** — `bash -n kit/cgo-bench-dualns.sh` clean; VM execution is Plan 5. Commit: `kit: banc dualns fermé-échec sur les offloads, preuve de traversée par compteurs, veth de gestion 10.99.0.x`

---

### Task 4: `kit/run-dualns.sh` orchestrator + serve bind override

**Files:**
- Create: `kit/run-dualns.sh`
- Modify: `cmd/cgo/server.go` (bind override), `cmd/cgo/server_test.go` (add bind resolution test — new file may be needed)

**Interfaces:** `run-dualns.sh up|down|serve` — `up`: bench up + testbedsrv in ns_server; `serve`: `up` + `sudo ip netns exec ns_client env CGO_HTTP_ADDR=10.99.0.2:9090 <cgo binary> --serve` foreground; `down`: bench down. Go side: `resolveHTTPAddr(flagVal string) string` — flag explicit ⇒ flag; else `CGO_HTTP_ADDR` env; else current default — the dashboard binds the management address only when the operator (or run-dualns.sh) says so.

- [ ] **Step 1: Failing test** — `resolveHTTPAddr`: empty flag + no env ⇒ current default; empty flag + `CGO_HTTP_ADDR=10.99.0.2:9090` ⇒ `10.99.0.2:9090`; explicit flag wins over env. (Read `cmd/cgo/server.go:280-345` first; extract the current default resolution into `resolveHTTPAddr` and thread it through the existing listen path — the TLS/redirect logic stays untouched.)
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./cmd/cgo/ -count=1`.
- [ ] **Step 5: the script** — `run-dualns.sh` (bash, `set -uo pipefail`, SUDO detection copied from the bench script):

```bash
#!/bin/bash
# run-dualns.sh — le chemin de production en une commande : banc durci,
# puits dans ns_server, tableau de bord DANS ns_client (10.99.0.2:9090).
# Usage: run-dualns.sh up|serve|down [--bin PATH]
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
BIN="${CGO_BIN:-$HOME/cgo/cgo-linux}"
while [ $# -gt 1 ]; do case "$1" in --bin) BIN="$2"; shift 2;; *) shift;; esac; done
CMD="${1:-up}"
SUDO=""; [ "$(id -u)" -ne 0 ] && SUDO="sudo -n"
case "$CMD" in
  up|serve)
    "$HERE/cgo-bench-dualns.sh" up || exit 1
    # puits : petit objet + masse, DANS ns_server (10.30.0.2)
    $SUDO ip netns exec ns_server sh -c "nohup '$BIN' testbedsrv -http 10.30.0.2:8081 -bulk 10.30.0.2:5201 >/tmp/testbedsrv-dualns.log 2>&1 &"
    sleep 1
    $SUDO ip netns exec ns_client ping -c1 -W2 10.30.0.2 >/dev/null || { echo "[!] puits injoignable"; exit 1; }
    [ "$CMD" = up ] && { echo "[run-dualns] banc + puits prêts (dashboard: run-dualns.sh serve)"; exit 0; }
    echo "[run-dualns] dashboard → http://10.99.0.2:9090 (depuis l'hôte)"
    exec $SUDO ip netns exec ns_client env CGO_HTTP_ADDR=10.99.0.2:9090 "$BIN" --serve
    ;;
  down) "$HERE/cgo-bench-dualns.sh" down; pkill -f testbedsrv-dualns 2>/dev/null; exit 0;;
  *) echo "usage: $0 up|serve|down [--bin PATH]"; exit 2;;
esac
```

- [ ] **Step 6: Verify + commit** — `bash -n kit/run-dualns.sh`; `go build ./...`. Commit: `kit: run-dualns.sh (banc+puits+dashboard dans ns_client), addr HTTP surchargeable (CGO_HTTP_ADDR)`

---

### Task 5: Doctor — offload gate with parseable state

**Files:**
- Modify: `pkg/doctor/doctor.go`
- Create: `pkg/doctor/offload_test.go`

**Interfaces:** `ParseOffloadFeatures(ethtoolKOutput string) map[string]bool` (feature-line name → is-on; parses `tcp-segmentation-offload: off` style lines, ignores fixed/`[fixed]` suffixed values); `Report` gains a `bench-offloads` Check: bench namespaces absent ⇒ status `warn`, detail `n/a — banc absent (kit/run-dualns.sh up)`; present ⇒ run `ip netns exec <ns> ethtool -k <iface>` for the six known pairs, any of GRO/GSO/TSO on ⇒ `fail` (in any mode — a lying bench is a lying bench), all off ⇒ `ok` with the six names. Non-Linux ⇒ `warn n/a`.

- [ ] **Step 1: Failing test** — table test on `ParseOffloadFeatures` with real ethtool-shaped output (on/off/fixed/`[requested on]` variants; a feature reported `off` must be `false` in the map, absent features absent). Plus: `Report("full")` on any OS contains a check named `bench-offloads` whose status is one of ok/warn/fail (never absent).
- [ ] **Step 2: red → Step 3: implement** (runtime.GOOS gate + `ip netns list` presence check; no build tags needed) `→ Step 4: green` — `go test ./pkg/doctor/ -count=1`.
- [ ] **Step 5: Commit** — `doctor: porte offloads du banc — échec franc si GRO/GSO/TSO actif, n/a honnête si banc absent`

---

### Task 6: Campaign on the hardened topology — `LinkPhysics`, `ProdDepsDualNS`, `channel_model`

**Files:**
- Modify: `pkg/campagne/campagne.go` (Deps + RunEvent), `pkg/campagne/prod.go`, `pkg/campagne/config.go`, `pkg/campagne/matrix.go` (config record), `cmd/cgo/runcli.go`, `pkg/api/server.go` (run start params)
- Create: `pkg/campagne/dualns.go`, `pkg/campagne/dualns_test.go`

**Interfaces:**
- `Deps` gains: `PhysicsIfaces []string` (netem applies to CliIf + every extra iface — w-edge AND c-wan), `ChannelModel string` (`"clean"` default | `"bursty"`), `Bench string` (label recorded, `"legacy"` default | `"dualns"`).
- `BurstGE(profileID string) (qdisc.GEModel, bool)` — P2: `{0.001, 0.1, 0.5, 0.0005}` (≈0.544 % mean), P3: `{0.002, 0.1, 0.5, 0.001}` (≈1.08 %); false for every other profile (caller MUST return a validation error, never silently uniform).
- `ProdDepsDualNS(direction string) Deps` — `TC: qdisc.NsRunner{"ns_wan_emulator"}`, `CliIf: "w-edge"`, `PhysicsIfaces: ["c-wan"]`, `TCShaper: qdisc.NsRunner{"ns_edge_router"}`, `ShaperIf: "e-srv"` (up) / `"e-wan"` (down), `Target/SmallURL/BulkAddr` = 10.30.0.2 trio, `Bench: "dualns"`, `BulkN` wiring identical to `prodDeps`, `StatsFn` polls the shaped egress in ns_edge (single-hop rule preserved).
- `runCLI` + API: `CGO_BENCH=dualns` (or `--bench dualns`) selects `ProdDepsDualNS`; `--channel-model clean|bursty` flag; bursty + non-P2/P3 ⇒ hard error before any cell runs.

- [ ] **Step 1: Failing tests** (`dualns_test.go`):
  1. `TestBurstGEOnlyP2P3` — P2/P3 return models whose `GEMeanLoss` lands 0.5–0.6 % / 1.0–1.15 %; P1/P4/others ⇒ ok=false.
  2. `TestRunEventAppliesPhysicsToAllIfaces` — FakeRunner Deps with `PhysicsIfaces: ["c-wan"]`: RunEvent's netem calls hit BOTH `w-edge` and `c-wan` (assert two `qdisc replace … netem` calls with distinct dev args).
  3. `TestRunEventBurstyValidation` — `ChannelModel: "bursty"` with profile P1 ⇒ error naming P1 and the clean|bursty rule; with P2 ⇒ the netem call carries `gemodel` and the P2 parameters.
  4. `TestProdDepsDualNSWiring` — up: shaper iface `e-srv`, stats poll `e-srv` via ns_edge; down: `e-wan`; `Bench == "dualns"`; bulk addr `10.30.0.2:5201`.
- [ ] **Step 2: red → Step 3: implement** — RunEvent's physics block becomes: resolve `ge *qdisc.GEModel` (nil unless bursty); loop `append([]string{d.CliIf}, d.PhysicsIfaces...)` calling `ApplyNetemPhysics`; `defaults()` sets `ChannelModel="clean"`, `Bench="legacy"` when empty; the bursty validation error fires before any qdisc mutation. `prod.go` gains `ProdDepsDualNS`; `runcli.go`/server thread `--bench`/`--channel-model` + env.
- [ ] **Step 4: green** — `go test ./pkg/campagne/ ./cmd/cgo/ ./pkg/api/ -count=1`; frozen suite untouched (`go test ./pkg/campagne/ -run TestWriter -count=1`).
- [ ] **Step 5: Commit** — `campagne: LinkPhysics multi-ifaces, ProdDepsDualNS, channel_model clean|bursty (P2/P3 seulement)`

---

### Task 7: flows=4 default, config labels, `valuation.json` sidecar, snapshot economics

**Files:**
- Modify: `pkg/campagne/config.go:47-48` (default 4), `cmd/cgo/main.go` (run flag default + help), `pkg/campagne/matrix.go` (Config record + manifest), `pkg/campagne/writer.go` (sidecar), `pkg/campagne/campagne.go` (Snapshot fields), `pkg/api/server.go` (snapshot passthrough — verify JSON tags suffice)
- Create: `pkg/campagne/valuation.go`, `pkg/campagne/valuation_test.go`

**Interfaces:**
- `Config` gains `ChannelModel string \`json:"channel_model"\`` and `Bench string \`json:"bench"\`` (recorded in matrixConfig + the run manifest — frozen CSV schema untouched).
- `WriteValuation(dir string, cfg Config, ev model.Event) error` — sidecar `valuation.json` per run event: Model A (`CapacityWasteARPerHour` for GOLD+PLATINE at ratio = `Drops×1448/Bytes` when `Bytes>0`), Model B (`MeteredWasteARPerHour` for 81.38 + 150 at window = `ChargeSec`), primary per `BuildEconomics` naming, `flows/channel_model/bench` labels, and the same disclaimer string. Ratio basis `"estimated"` (drops×1448 proxy — stated), never presented as measured.
- `Snapshot` gains `CostPrimaryARPerH float64 \`json:"cost_primary_ar_per_h"\``, `CostB2BARPerH float64 \`json:"cost_b2b_ar_per_h"\``, `JFIPct float64 \`json:"jfi_pct"\`` — populated from live event data (existing `WastedBytes`/`CostARPerH` fields keep their frozen meaning).
- CLI: `--flows` default 4 (help text: « 1 = régime historique gelé »).

- [ ] **Step 1: Failing tests** — `TestWriteValuationSidecar` (write to temp dir; JSON parses; five scenarios present; primary named for the link-less default GOLD... campagne has no LinkType — primary key = `"campaign"` ⇒ no primary pointer, all scenarios flat: document this in the sidecar's own `note` field: « campagne sans type de lien déclaré — aucun primaire marqué »); `TestConfigRecordsLabels` (matrixConfig JSON contains flows=4 default, channel_model, bench); `TestSnapshotCostFields` (Snapshot serializes the three new keys).
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./pkg/campagne/ ./cmd/cgo/ ./pkg/api/ -count=1`; `python book/tools/audit-archive.py` byte-identical.
- [ ] **Step 5: Commit** — `campagne: flows=4 par défaut, étiquettes bench/channel_model au manifest, sidecar valuation.json, coût entreprise dans le snapshot live`

---

## Self-review (author's, kept in-document)

**Spec coverage (Plan 2 slice):** §3 fail-closed offloads + leak proof + mgmt veth + run-dualns (Tasks 3–4); §6 gemodel fix + LinkPhysics + flows=4 + channel model + P2/P3 gate (Tasks 1, 6); valuation sidecar + snapshot economics (Task 7); doctor offload check (Task 5); Plan-1 sink correction (Task 2). Left to Plan 5 (VM): empirical gemodel calibration, P4/G6 re-validation, regime-bridge cells, délestage/blackout, `-race`, script execution; Plan 3: UI channel-model selector + `/api/cost/enterprise`.

**Type consistency:** `qdisc.GEModel`/`GEMeanLoss`/`ApplyNetemPhysics` consumed by Task 6's `BurstGE`; `metrics` Model functions consumed by Task 7's `WriteValuation` (imports `pkg/metrics` — no cycle: metrics imports nothing from campagne).

**Executor notes:** Task 1 removes exported symbols (`ApplyNetemBurst`, `Profile.LossBurst*`) — `grep -rn` both before committing; Task 6 touches `RunEvent`'s qdisc block — re-read `campagne.go:215-260` first, the reset-ordering comments are load-bearing (documented production incidents).
