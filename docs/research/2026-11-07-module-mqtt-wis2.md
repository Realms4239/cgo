# Module MQTT / WIS 2.0 + CAP — missing tool (primary-source research)

*Date: 2026-11-07 · Repo: `operator-cockpit` (CGO / Meteolink) · Method: primary sources only. Every claim cites a primary source (repo `file:line` or fetched URL). Anything not fetchable is marked `unverified — primary source not found` and never paraphrased as requirement. Deep research only: no code, no commit, no live broker run.*

> **Causal note (the whole file turns on it):** CGO measures links; a WIS2 Node is a separate dissemination system. WIS 2.0 is the *transport* (MQTT publish-subscribe over the public Internet) [R1]; CAP 1.2 is the *payload vocabulary* for all-hazard alerts [R2]; CGO's job is to measure whether a constrained link can carry small alert/telemetry objects on time [R4, R5]. Confusing measurement with dissemination is the overclaim this module must not make.

---

## 1. Primary-source table

| # | Topic | Primary source consulted (fetched 2026-11-07) | What it actually says | Status |
|---|---|---|---|---|
| R1 | WMO WIS 2.0 overview | https://community.wmo.int/site/knowledge-hub/programmes-and-initiatives/wmo-information-system-wis/wis2-overview | Operational 01/01/2025; unlike GTS private links, WIS 2.0 uses the public Internet + MQTT publish-subscribe; each publisher must operate a WIS2 Node; Global Broker / Global Cache / Global Discovery Catalogue / Global Monitor; topic hierarchy `origin/a/wis2/<centre-id>/data/core\|recommended`, `cache/a/wis2/<centre-id>/data/core`, `origin/a/wis2/<centre-id>/metadata`; WIS2 Notification Message (WNM) with a canonical HTTP(S) download link; WCMP2 discovery metadata; `wis2box` = FOSS reference Node; registration via PR/NFP + centre-id + GISC assessment, only for an *operational* Node with a public MQTT endpoint regularly publishing notifications | **verified (primary: WMO)** |
| R2 | CAP 1.2 (OASIS Standard, 01 July 2010) | https://docs.oasis-open.org/emergency/cap/v1.2/CAP-v1.2-os.html | §1.1 open, non-proprietary format for all types of alerts/notifications, no particular application or telecom method; §1.3 structure `<alert>` → `<info>` (urgency/severity/certainty) → `<resource>`/`<area>`; §3.2.1 namespace `urn:oasis:names:tc:emergency:cap:1.2`, required `<alert>` elements `identifier`, `sender`, `sent`, `status`, `msgType`, `scope`; §3.2.2 required `<info>` elements `category`, `event`, `urgency`, `severity`, `certainty`; §4.3/§4.4 conformance as Message Producer / Consumer | **verified (primary: OASIS)** |
| R3 | Eclipse Mosquitto `mosquitto.conf(5)` man | https://mosquitto.org/man/mosquitto-conf-5.html | Config file for the broker; `listener` + per-listener TLS (`cafile`/`certfile`/`keyfile`, `require_certificate`, `use_identity_as_username`/`use_subject_as_username`, `tls_version`); `password_file`, `acl_file` (`topic [read\|write\|readwrite\|deny] <topic>`, `user <username>`, `pattern` with `%c`/`%u`), `allow_anonymous` (defaults `false` unless no listener defined), `max_qos`, `retain_available`, `persistence`, `sys_interval` (`$SYS` hierarchy), bridge options (`topic <pattern> [[in\|out\|both] qos] [local-prefix remote-prefix]`, bridge TLS `bridge_cafile`/`bridge_certfile`/`bridge_keyfile`, bridge notification topic `$SYS/broker/connection/<id>/state`) | **verified (primary: Eclipse Mosquitto)** |
| R4 | Repo identity: what CGO is | `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:10-21` | CGO = Congestion Ground-truth Observatory: a reproducible bench for comparing AQM/BBR policies on constrained access links, ISGEI × Météo Madagascar / DGM; frozen corpus 159 runs / 528 rows, never re-measured | **verified (repo)** |
| R5 | Repo: DGM scenario + Model C (CAP SLA) | `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:134-136`, `:151-155`; `pkg/metrics/cost.go:75-112` | Download audit = bulk FROM sink through shaped `e-wan` = *"the actual DGM scenario (200 MB EUMETSAT downlink vs CAP egress)"*; Model C: loaded 16 KB completions at 220 ms, breach `< 95%`; `cost.go:75-79` states 220 ms *"is NOT a threshold imposed by the WMO"*, WIS 2.0 "real time" = "a few seconds to a few minutes" with no numeric per-bulletin threshold | **verified (repo)** |
| R6 | Repo: ledger status of the CAP objective | `docs/VERIFY.md:48` | Enterprise tariff scenarios + Model A/B/C = `RECORDED` contractual scenarios, not measurements; *"Not re-run as a measurement here"* | **verified (repo)** |
| R7 | Repo: case-study (forecasting server) posture | `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md:16-38`, `:65-78`; `docs/VERIFY.md:60` | Read-only only; never: capture/probe/load/write/root/login/scan/query; BLOCKED on the operator's written grant (a)–(d); nothing was read, sent or executed | **verified (repo)** |
| R8 | Repo: hardening plan defers WIS/Mosquitto | `docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:82-83` | §6.3 DEFERRED: *"WIS 2.0 / CAP publication; Mosquitto multi-site"* — each deferred item names its unblock condition (primary docs fetched, image procured, human grant signed) | **verified (repo)** |
| R9 | Repo: current live surface (no MQTT) | `docs/api.md:21`, `:69`; `go.mod:1-10` | Live data is SSE over HTTP: `/api/stream` 10 Hz (ring 2048, `503` > 64 subscribers), `/api/events` last 50; `go.mod` direct deps are fyne/bubbletea/lipgloss/x-sys only — no MQTT client library | **verified (repo)** |
| R10 | Repo: absence of MQTT/WIS/CAP code | grep (case-insensitive) `mqtt\|mosquitto\|wis2\|paho` in `pkg/` and `cmd/` → **no matches** (2026-11-07); repo-wide grep for `MQTT`/`Mosquitto` hits only docs/specs/plans (R8 and the 2026-11-06 hardening spec), never `pkg/` or `cmd/` | No broker, no client, no topic, no CAP XML producer/consumer exists in the product code | **verified (repo, by absence)** |
| R11 | WNM / WCMP2 / WIS2 Topic Hierarchy full schemas | Linked from R1 as WMO-No. 1060 Vol. II App. D/E and WMO-No. 1061; not fetched in this pass | Exact WNM JSON fields, WCMP2 record fields, and the full topic-hierarchy annex beyond the `origin/a/wis2/<centre-id>/data/core\|recommended` shapes shown on R1 | `unverified — primary source not found` |
| R12 | WMO `warnings.wmo.int` CAP portal / any mandated QoS, retain policy, or CAP-over-WIS2-topic mapping | Not fetched in this pass; R1 does not state a QoS value, a retain policy, or that CAP XML travels on a WIS2 data topic | Whether CAP documents are carried as WIS2 data granules, and at what QoS | `unverified — primary source not found` |
| R13 | Madagascar / DGM centre-id, GISC, registration state | Not established from primary sources in this pass | Any claim that DGM holds a centre-id, a registered Node, or a Global Broker subscription | `unverified — primary source not found` |
| R14 | `wis2box` deployment specifics (install, config keys, supported payloads) | R1 names it only as *"[WIS2 in a box] (wis2box): a Free and Open Source (FOSS) reference implementation of a WIS2 Node"* with a docs link | How to run it, what it accepts, whether it accepts CAP XML | `unverified — primary source not found` |

> Rule applied throughout: §2–§6 never extend R1–R3 beyond what was fetched. Where a proposal below goes beyond R1–R3, it is labelled **PROPOSED** and is not a standard, a WMO requirement, or a CGO commitment.

---

## 2. Responsibility — the missing tool

### 2.1 What is missing (verified)

1. **No MQTT plane at all.** There is no broker, no client, no topic, and no MQTT dependency in the product [R9, R10]. The only live distribution is HTTP SSE inside one host (`/api/stream` 10 Hz, 64-subscriber cap) [R9]; that is a dashboard feed, not a dissemination system, and it does not leave the bench host.
2. **No WIS2 Node.** A WIS2 Node is *two* public endpoints — an MQTT broker publishing WIS2 notifications *and* an HTTP server for data download — plus WCMP2 metadata records and per-granule notifications [R1]. None of these exist in CGO [R10].
3. **No CAP producer/consumer.** CGO computes a CAP *SLA* (Model C, 16 KB completions vs 220 ms / 95%) [R5] but produces no CAP XML document, validates none against the R2 schema, and implements neither §4.3 producer nor §4.4 consumer conformance [R2, R10]. "CAP" in CGO today means "the deadline objective for a small transaction", never "a CAP message" [R5, R6].
4. **The gap is already named as deferred, not forgotten.** The hardening spec lists *"WIS 2.0 / CAP publication; Mosquitto multi-site"* under DEFERRED with the unblock condition *"primary docs fetched"* [R8]. This file is the first half of that unblock condition (R1–R3 fetched); the second half — a human decision that dissemination is in scope at all — is not taken here.

### 2.2 What the missing tool's responsibility is (and is not)

| Responsibility | Owner | Primary source |
|---|---|---|
| Measure whether a constrained link delivers a 16 KB object within 220 ms under load (Model C) | **CGO (existing)** — unchanged | R5 (`pkg/metrics/cost.go:75-112`) |
| Carry small objects/notifications over the public Internet between sites | **MQTT broker layer (missing)** — Mosquitto-class, TLS + ACL per R3 | R1 (WIS 2.0 = MQTT pub-sub over public Internet), R3 |
| Publish data + metadata notifications and serve downloads | **WIS2 Node (missing, separate system)** — broker + HTTP server + WCMP2 + registration | R1 |
| Express an all-hazard alert in a standard message format | **CAP 1.2 producer (missing)** — only if the operator wants CAP alerts | R2 |
| Decide DGM's centre-id / registration / GISC path | **Human / DGM / WMO Secretariat (out of scope for any code)** | R1 (registration steps), R13 = `unverified — primary source not found` |

**Non-responsibility (causal boundary):** the probe must not turn CGO into a Node by default. Measurement evidence (sealed `audit_report.json`, `telemetry.csv`) stays in CGO's evidence model [R4]; the probe *reads* a sealed report as input and publishes *derived* notifications. It does not re-grade, re-measure, or re-seal anything.

### 2.3 Why a separate tool, not a CGO feature

- CGO's product contract is a single Go binary with embedded React and no DB, frozen CSV schemas, and a ledger discipline where `VERIFIED` requires a named command on a named host [R4, R6; `docs/VERIFY.md:7-18`]. Adding a public Internet broker + TLS + ACL + registration surface to that binary widens its trust boundary to exactly the plane the hardening spec deferred until authN/Z lands [R8].
- The case-study contract forbids CGO from connecting to operator infrastructure at all — *"CGO itself connects to nothing"* [R7]. A probe that *does* connect (MQTT client → broker) is therefore a different trust class and must live in a different repo with its own grant boundary.
- A missing tool that is also a *separate* tool can be refused, revoked, or left uninstalled without touching the measurement bench. That reversibility is the point.

---

## 3. Exported API — PROPOSED (not a standard)

Everything in §3 is a **proposal** for the future independent tool. It is not a WMO requirement, not an OASIS conformance claim, and not a CGO commitment. Standard shapes are marked **[R1]/[R2]/[R3]**; inventions are marked **PROPOSED**.

### 3.1 Topic tree

**Standard shape (from R1):**

```
origin/a/wis2/<centre-id>/data/core
origin/a/wis2/<centre-id>/data/recommended
origin/a/wis2/<centre-id>/metadata
cache/a/wis2/<centre-id>/data/core
```

- `origin/a/wis2/<centre-id>/data/core` implies free, unrestricted access (Resolution 1); the Global Cache downloads it and re-announces on `cache/a/wis2/<centre-id>/data/core` [R1].
- `.../data/recommended` may require access control and is *not* downloaded by Global Caches [R1].
- `.../metadata` carries the WCMP2 canonical link; the Global Discovery Catalogue is updated from it [R1].

**PROPOSED lab namespace (default until a centre-id exists):**

```
cgo/v1/probe/<site-id>/link-quality        # PROPOSED — lab only, never on a WIS2 Node
cgo/v1/probe/<site-id>/cap-sla              # PROPOSED — lab only
cgo/v1/probe/<site-id>/status/$SYS-like     # PROPOSED — use broker $SYS instead where possible [R3]
```

Rationale: `origin/a/wis2/...` requires a centre-id proposed by the Member and endorsed by the WMO Secretariat [R1]. Publishing under `origin/a/wis2/<invented-id>/...` before assignment would be a false WIS2 claim. `cgo/v1/...` keeps the lab tool honest and is what Mosquitto ACLs can scope (`acl_file` topic rules, `pattern` with `%u`/`%c`) [R3].

**PROPOSED WIS2 mapping (only after centre-id assignment + GISC assessment [R1]):**

| Probe output | WIS2 topic | Class |
|---|---|---|
| Sealed audit summary (derived JSON, §3.2) | `origin/a/wis2/<centre-id>/data/recommended/cgo/link-quality` | PROPOSED sub-path under the standard `.../data/recommended` prefix [R1]; `recommended` because access control applies and Global Cache must not ingest it [R1] |
| WCMP2 record for the above dataset | `origin/a/wis2/<centre-id>/metadata` | standard [R1]; record schema = `unverified — primary source not found` [R11] |
| CAP XML document (§3.3) | HTTP download URL referenced from the WNM canonical link, *not* the MQTT payload | PROPOSED separation; whether CAP travels as a WIS2 data granule = `unverified — primary source not found` [R12] |

### 3.2 Notification payload

**Standard (from R1):** a WIS2 Notification Message (WNM) contains a *"canonical" link indicating the HTTP(S) endpoint from which the data can be downloaded* [R1]. Exact WNM fields = `unverified — primary source not found` [R11].

**PROPOSED probe notification (JSON, lab namespace):**

```json
{
  "schema": "cgo.probe.notification/v1",
  "site_id": "PROPOSED — operator-assigned, never a hostname",
  "audit_id": "<id from sealed audit_report.json>",
  "seal_sha256": "<SHA-256 of telemetry.csv — copied, never recomputed differently>",
  "measured_at": "2026-11-07T00:00:00Z",
  "link_quality": {
    "cap_deadline_ms": 220,
    "cap_ok_pct": 96.5,
    "cap_breach": false,
    "basis": "loaded 16 KiB completions"
  },
  "download_url": "https://<probe-http-host>/<audit_id>/audit_report.json",
  "cap_url": "https://<probe-http-host>/<audit_id>/alert.xml"
}
```

- `cap_deadline_ms` / `cap_breach` semantics are copied from `AssessCAPSLA` (`Measured=false` ⇒ no value, never a pass) [R5]. The 220 ms / 95% figures stay labelled *"objectif DGM, cadre WIS 2.0"* — never a WMO mandate [R5].
- `seal_sha256` is a *copy* of the CGO seal [R4]; the probe never re-seals. A report without a seal is refused (fail-closed), mirroring the repo's *"unmeasured ⇒ `Unavailable`, never zero, never a pass"* invariant [R4:40-68].

### 3.3 CAP mapping

**Standard (from R2):** CAP 1.2 is XML, namespace `urn:oasis:names:tc:emergency:cap:1.2` [R2 §3.2.1]. Required `<alert>` children: `identifier`, `sender`, `sent`, `status`, `msgType`, `scope` [R2 §3.2.1]. Required `<info>` children: `category`, `event`, `urgency`, `severity`, `certainty` [R2 §3.2.2]. `status` ∈ {Actual, Exercise, System, Test, Draft}; `msgType` ∈ {Alert, Update, Cancel, Ack, Error}; `scope` ∈ {Public, Restricted, Private} [R2 §3.2.1]. `sender` must be globally unique (e.g. domain-based) [R2 §3.2.1]. Producer conformance is defined in §4.3, consumer in §4.4 [R2].

**PROPOSED mapping (probe → CAP, only when the operator asks for CAP alerts):**

| CAP element [R2] | PROPOSED value | Source |
|---|---|---|
| `<identifier>` | `<audit_id>-<sha256-prefix>` (unique per message, no spaces/commas) | PROPOSED; uniqueness rule [R2 §3.2.1] |
| `<sender>` | operator domain-based id, e.g. `cgo-probe/<site-id>.dgm.example` | PROPOSED; globally-unique rule [R2 §3.2.1] |
| `<sent>` | audit `measured_at` in DateTime format, UTC as `-00:00` | rule [R2 §3.2.1]; format rule [R2 §3.2.2 `effective` note] |
| `<status>` | `Actual` | PROPOSED; domain [R2 §3.2.1] |
| `<msgType>` | `Alert` on first breach, `Update`+`<references>` on recovery, `Cancel` to retract | PROPOSED sequencing; domain + `<references>` form `sender,identifier,sent` [R2 §3.2.1] |
| `<scope>` | `Private` (named operator addresses) or `Restricted` | PROPOSED; domain [R2 §3.2.1] — never `Public` without operator sign-off |
| `<info><category>` | `Infra` (telecom infrastructure) | PROPOSED choice from the R2 domain [R2 §3.2.2] |
| `<info><event>` | `Link quality below operator objective` | PROPOSED text; element required [R2 §3.2.2] |
| `<info><urgency>/<severity>/<certainty>` | derived from `cap_ok_pct` bands — **bands are PROPOSED and must be operator-signed** | domains [R2 §3.2.2]; no band table exists in R2 — inventing one silently would be a toy |
| `<info><parameter>` | `valueName="CGO_CAP_OK_PCT"`, `value="<cap_ok_pct>"`; `valueName="CGO_SEAL_SHA256"`, `value="<seal>"` | PROPOSED use of the optional `<parameter>` mechanism [R2 §3.2.2] |
| `<resource>` | the sealed `audit_report.json` / `telemetry.csv` via `<uri>` | PROPOSED; `<resource>` = optional reference to a digital asset [R2 §1.3.3] |

**Hard boundary:** a healthy probe (no breach) publishes **no CAP message**. CAP `msgType=Alert` carries *"initial information requiring attention"* [R2 §3.2.1]; routine telemetry is not an alert. The steady-state channel is the JSON notification in §3.2, not CAP.

### 3.4 Broker-facing contract (Mosquitto)

**Standard capabilities (from R3) the proposal relies on:**

- Separate listeners per trust class: `listener <port> [bind]` [R3]; TLS listener with `cafile`/`certfile`/`keyfile` [R3].
- Client authentication: `require_certificate true` + `use_identity_as_username true` makes the certificate CN the ACL username, password unused [R3]; or `password_file` + `allow_anonymous false` [R3].
- Per-user topic control: `acl_file` with `topic [read|write|readwrite|deny] <topic>` and `user <username>` lines; `pattern` substitution `%c` (client id) / `%u` (username) [R3].
- QoS cap per listener: `max_qos` [R3]; retained-message support: `retain_available` [R3]; offline durability: `persistence` + `persistence_location` [R3]; broker self-reporting: `sys_interval` publishes the `$SYS` hierarchy [R3].
- Multi-broker fan-out: `connection` bridges with `topic <pattern> [[in|out|both] qos-level] [local-prefix remote-prefix]`, bridge TLS (`bridge_cafile`, `bridge_certfile`, `bridge_keyfile`), and `$SYS/broker/connection/<id>/state` notifications [R3].

**PROPOSED profile (lab):**

```ini
# PROPOSED mosquitto.conf fragment — lab only, not a WIS2 Node config
listener 8883
cafile /etc/cgo-wis2/ca.crt
certfile /etc/cgo-wis2/broker.crt
keyfile /etc/cgo-wis2/broker.key
require_certificate true
use_identity_as_username true
allow_anonymous false
max_qos 1
retain_available true
persistence true
persistence_location /var/lib/cgo-wis2/
sys_interval 10
acl_file /etc/cgo-wis2/acl
```

- `max_qos 1` and the retain policy are **PROPOSED**: R1 states no QoS value and R3 only defines the `max_qos` knob [R3, R12]. QoS 1 is proposed because QoS 0 is fire-and-forget and the notification must survive a constrained link — but that reasoning is engineering choice, not a WMO mandate, and is recorded as such.
- Bind to a specific interface/address (`listener <port> <bind>` [R3]); a public-Internet listener is a *separate, later* decision gated on registration [R1], never the lab default.

---

## 4. Dependencies

| Dependency | Kind | Status / gate | Primary source |
|---|---|---|---|
| Eclipse Mosquitto broker (external binary) | Runtime, not a Go module | New external dependency — human approval required before scaffolding (repo rule: *"a new dependency"* is a stop-and-ask item) [`AGENTS.md:85-88`] | R3 (its config surface is the contract) |
| MQTT client library (Go) | New Go module dependency | **None exists today** [R9: `go.mod:5-10`]; adding one is a stop-and-ask item [`AGENTS.md:85-88`]; candidate choice is `unverified — primary source not found` (no client library was evaluated in this pass) | R9, R10 |
| CAP XML generation | `encoding/xml` (stdlib) | No new dependency proposed; schema validation against R2 §3.4 is the work, not a library | R2 §3.4 (XML Schema) |
| `wis2box` (FOSS reference Node) | Alternative runtime | **Not proposed as a dependency of the probe.** R1 names it as the reference Node implementation [R1]; deploying it is a *separate* decision by the eventual Node operator. Its install/config specifics = `unverified — primary source not found` [R14] | R1, R14 |
| CGO sealed report (`audit_report.json` + `telemetry.csv`) | Input artifact, read-only file | The probe consumes CGO output; it does not import CGO packages, re-grade, or write into `book/data/runs` (frozen corpus is untouchable) [`AGENTS.md:35`] | R4, R7 |
| Centre-id, GISC assessment, Global Broker subscription | Organisational dependency | External to any code; required before any `origin/a/wis2/...` publishing [R1]; Madagascar's state = `unverified — primary source not found` [R13] | R1, R13 |
| Operator written grant | Human dependency | Required before the probe reads *anything* operator-side, per the case-study contract's grant items (a)–(d) [R7] | R7 |

**Explicitly not a dependency:** CGO's frozen corpus, CSV schemas, and six-gate verification system [R4, R6]. The probe adds no gate to CGO and must not weaken one.

---

## 5. Primary-source requirements (what must be fetched before building)

Each item below is a *requirement on the builder*, not a claim. Status = this pass.

| # | Requirement | Why | Status |
|---|---|---|---|
| Q1 | Fetch WMO-No. 1060 Vol. II App. D (WIS2 Topic Hierarchy) and App. E (WNM, WCMP2) | R1 links them but does not reproduce the schemas; §3.2/§3.3 cannot be implemented from R1 alone | `unverified — primary source not found` [R11] |
| Q2 | Fetch the WMO-No. 1061 Guide §2.6.1.1 (Node registration/decommissioning) | R1 summarises registration; the normative steps and assessment criteria are in the Guide | `unverified — primary source not found` [R11] |
| Q3 | Establish whether CAP documents are carried as WIS2 data granules, and at what QoS/retain policy | §3.3's separation (CAP via HTTP canonical link, not MQTT payload) is a proposal until this is answered | `unverified — primary source not found` [R12] |
| Q4 | Obtain the operator's centre-id status or a written decision not to register | Determines whether §3.1's WIS2 mapping is ever used or the lab namespace is permanent | `unverified — primary source not found` [R13] |
| Q5 | Operator-signed CAP severity bands (`cap_ok_pct` → urgency/severity/certainty) | R2 defines the domains but no mapping from a link metric to them [R2 §3.2.2]; unsigned bands = invented thresholds | Not a fetch — a signature |
| Q6 | Evaluate `wis2box` docs (install, accepted payloads, CAP support) | Decides probe→`wis2box` vs probe→own broker; R1 only names it [R1] | `unverified — primary source not found` [R14] |
| Q7 | Select an MQTT client library with a primary source (release notes / API docs) | No library is in `go.mod` today [R9]; the choice must be citable, not remembered | Not started |

**Build rule:** no slice in §7 starts until its Q-items are closed. In particular, S4 (WIS2 publishing) is blocked on Q1–Q4 exactly as the hardening spec's DEFERRED condition requires [R8].

---

## 6. Gap analysis — why CGO is NOT a WIS2 Node

| WIS2 Node requirement [R1] | CGO today | Gap |
|---|---|---|
| MQTT broker exposed over the public Internet, publishing WIS2 notifications for metadata and data [R1] | No MQTT code at all [R10]; live feed is HTTP SSE on one host [R9] | **Absent.** Not a configuration gap — a whole plane |
| HTTP server enabling download of data and metadata [R1] | Dashboard on `10.99.0.2:9090` (management veth) serves the UI/API, not data granules for the WIS2 network [R4:29] | **Absent as a Node function.** The dashboard's HTTP surface is not a WIS2 download endpoint |
| WCMP2-compliant metadata records; publish a WNM on `origin/a/wis2/<centre-id>/metadata` with the canonical link [R1] | No WCMP2 records, no WNM messages [R10]; record schema itself unfetched [R11] | **Absent** |
| Topic hierarchy `origin/a/wis2/<centre-id>/data/core\|recommended` [R1] | No topics; centre-id never assigned [R10, R13] | **Absent** |
| Registration: PR/NFP request, centre-id endorsement, WIS2 Register, GISC assessment, Global Broker subscription — only for an *operational* Node with a public MQTT endpoint *regularly publishing* [R1] | No registration, no assessment, no subscription; and per R1's Note 2, registration is only considered for an already-operational endpoint | **Absent, and the order matters:** broker first, paperwork after — not the reverse |
| 24/7 operation with Global Monitor metrics/alerting [R1 Global Monitor] | Bench + audit tooling; no 24/7 dissemination duty | **Different duty.** CGO's availability model is "run an audit, seal the evidence" [R4] |

**Causal statement:** CGO's evidence chain (raw probes → freeze → manifest → `hash8` provenance) [R4] proves *what the link did*. A WIS2 Node's chain (data granule → WNM with canonical link → Global Cache/Catalogue) [R1] proves *that the network can fetch the data*. These are different claims, different operators, different failure modes. A tool that measures the link cannot, by measuring, become entitled to publish for the network — and a Node that publishes cannot, by publishing, prove the link was good. The probe sits *between* them: it reads sealed measurement evidence and emits notifications, claiming neither role.

**What CGO *does* legitimately supply** (and the probe may consume): the sealed audit report, the Model C CAP-SLA verdict with its honest *"operator objective, not a WMO mandate"* framing [R5], and the ledger discipline (`VERIFIED` only on a named command) [R6] as the standard the probe's own evidence must meet.

---

## 7. Target tool — `cgo-wis2-probe/gameplan` (independent repo)

**Name:** `cgo-wis2-probe/gameplan` — one independent repository, name to be fixed by the human at creation; not created by this research. **Not** a package, directory, or branch inside `operator-cockpit`.

**Boundary (from R7):** the probe is the first CGO-adjacent component allowed to *connect* to anything — and only to a broker the human configures, only after the operator grant [R7]. Everything else (frozen corpus, CSV schemas, CGO gates) stays untouched [R4, `AGENTS.md:35`].

**Inputs (read-only):** `data/audits/<id>/audit_report.json` + `telemetry.csv` + seals [R4]. **Outputs:** MQTT notifications (§3.2) and, on operator demand, CAP XML (§3.3). **Never:** re-grading, re-sealing, writing into CGO, or publishing under `origin/a/wis2/...` before Q4 closes [R13].

### Slices (dependency order; each gated on §5)

| Slice | Scope | Acceptance (all required) |
|---|---|---|
| **S0** | Close Q1–Q4 (fetch WNM/WCMP2/topic-hierarchy/Guide; establish centre-id status; decide CAP-carriage) | Sources table updated with fetched URLs; any item still unfetched stays `unverified — primary source not found` and blocks S4 |
| **S1** | Lab Mosquitto: TLS listener, `require_certificate` + `use_identity_as_username`, `acl_file`, `max_qos`, `persistence`, `$SYS` (`sys_interval`) [R3] | Anonymous connect refused; wrong-CN cert refused; `mosquitto_sub` with a valid cert receives only its ACL topics; broker survives restart with retained messages intact (persistence proof) |
| **S2** | Probe publisher, lab namespace only (`cgo/v1/probe/<site-id>/...`, §3.1) | Sealed report → notification round-trip on the lab broker; unsealed/unsigned report refused; `Measured=false` reports publish *nothing* (never a pass by default [R5]) |
| **S3** | CAP producer (§4.3 conformance [R2]) | Generated XML validates against the R2 §3.4 schema; `<sender>` globally unique; `Update`/`Cancel` carry correct `<references>` [R2 §3.2.1]; healthy probe emits no CAP |
| **S4** | WIS2 publishing (only if Q4 grants a centre-id and Q1–Q3 close) | WNM on `origin/a/wis2/<centre-id>/metadata` with canonical link [R1]; data on `.../data/recommended` [R1]; GISC assessment path documented — until then this slice **does not run** |
| **S5** | Bridge / multi-site fan-out (Mosquitto `connection` + `topic ... in/out/both`, bridge TLS [R3]) | Two lab brokers exchange notifications with prefixes intact; bridge state visible on `$SYS/broker/connection/<id>/state` [R3]; loop/duplicate behaviour documented |
| **S6** | Evidence + ledger (probe's own `VERIFY`-style rows) | Every claim row names command + host + date + saved output [R6]; fault proofs in §8 pass; a `BLOCKED` row names its condition and stays blocked |

**Sequencing constraint (from R8):** S1 must not precede CGO's own authN/Z decision — the hardening spec defers Mosquitto precisely because a second unauthenticated plane would compound exposure. The probe's broker is a *new* plane; the human sequences it.

---

## 8. Anti-toy criteria

A toy here = a demo that *looks* like WIS2/CAP dissemination but is a mock broker, a hardcoded payload, or a screenshot. Each criterion is falsifiable; a failure is a `REFUTED` row, not a quiet fix [R6].

1. **Real broker, real TLS.** S1 passes against a real Mosquitto binary with certificate auth [R3]. A mock/in-process fake broker in the acceptance test = toy, refused. Fault proof: replace the valid client cert with an expired one → connect refused.
2. **ACLs are enforced, not decorated.** A client with a valid cert but no ACL for `cgo/v1/probe/<other-site>/...` must be denied publish/subscribe [R3 `acl_file`]. Mutate the ACL to `deny` → the same publish must fail. (Pattern substitution `%u`/`%c` [R3] gets its own test: site A cannot write site B's prefix.)
3. **No green without a measured input.** Feed the probe a report with `Measured=false` / no loaded samples → it publishes nothing and logs the refusal [R5]. Feed it a fabricated `cap_ok_pct` with no seal → refused. This is CGO's own invariant applied to the probe [R4].
4. **CAP is schema-valid or it does not ship.** Every emitted CAP XML validates against R2 §3.4. Fault proof: hand-edit one `<urgency>` to an out-of-domain value → validation must fail and the message must not publish. §4.3 producer conformance or nothing [R2].
5. **The 220 ms framing survives the trip.** Any human-facing string derived from a notification says *"objectif DGM, cadre WIS 2.0"* and never presents 220 ms as a WMO threshold [R5]. Automated check: grep the probe's output corpus for `WMO mandate` / `seuil OMM` phrasings → must be empty.
6. **No false WIS2 identity.** Until Q4 closes, the probe must never publish to `origin/a/wis2/...` [R1]; a test asserts the lab config contains no `origin/a/wis2` topic. After Q4, only the assigned centre-id appears, and the GISC assessment record is on file [R1].
7. **Persistence is proven by killing the broker.** Stop Mosquitto mid-queue, restart, assert the retained/queued notification is delivered [R3 `persistence`]. A clean shutdown test does not count.
8. **Bridge loops are tested, not assumed.** Two brokers bridged both ways (`topic ... both` [R3]) must not amplify a message; duplicate-delivery behaviour is measured and recorded, not waved away.
9. **Evidence discipline matches CGO's.** The probe repo keeps the same ledger standard: `VERIFIED` only on a named command + host + date + saved output; `BLOCKED` rows name their condition [R6]. A claim without a row is a hypothesis and is labelled as such.
10. **The case-study grant gates everything operator-side.** No probe run reads operator configuration, names a real service/host, or publishes on behalf of DGM without the grant items (a)–(d) [R7]. Default identity is the pseudonym, exactly as the contract requires [R7].

---

## 9. What stayed unproven

- WNM / WCMP2 / full topic-hierarchy schemas [R11]; CAP-over-WIS2 carriage and mandated QoS [R12]; DGM centre-id / GISC / registration state [R13]; `wis2box` specifics [R14] — all `unverified — primary source not found` in this pass.
- No broker was installed, no client library chosen, no CAP XML generated, no notification published. This file is research; §7's slices are proposals, and their acceptance commands have not run.
- Whether WIS2 dissemination is in CGO's thesis scope at all is a human decision [R8]; this file does not take it.

---

*Sources: R1 WMO WIS2 Overview (community.wmo.int, fetched 2026-11-07); R2 OASIS CAP 1.2 OS (docs.oasis-open.org/emergency/cap/v1.2/CAP-v1.2-os.html, fetched 2026-11-07); R3 Eclipse Mosquitto `mosquitto.conf(5)` man (mosquitto.org/man/mosquitto-conf-5.html, fetched 2026-11-07); R4–R10 repo paths as cited inline. Anything else is labelled `unverified — primary source not found` where it appears.*
