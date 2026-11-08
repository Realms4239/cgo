# Module Master / DGM — demonstration requirements (primary-source research)

*Date: 2026-11-07 · Repo: `operator-cockpit` (CGO / Meteolink) · Method: primary sources only. Every claim cites a primary source (repo `file:line` or fetched URL). Anything not fetchable is marked `unverified — primary source not found` and never paraphrased as requirement.*

---

## 1. Primary-source table

| # | Topic | Primary source consulted (fetched) | What it actually says | Status |
|---|---|---|---|---|
| S1 | Repo identity: single Go binary, embedded React, frozen CSV + SHA-256, no DB | `README.md:5-6`, `README.md:37`, `docs/architecture.md:1-4` | Real-time test bench + terminal/browser visualiser; `go:embed` dashboard; no DB / no service deps | **verified (repo)** |
| S2 | Repo: SSE 10 Hz, G0–G7 gates, freeze + manifest + `hash8` provenance | `docs/architecture.md:23-41`, `docs/architecture.md:60-83` | 10 Hz hub; gates G0–G7; `ProvenanceSHA` / `ProvenanceHash8` in SVG, report, `/api/integrity` | **verified (repo)** |
| S3 | Repo: API surface + observe mode 501 | `docs/api.md:1-11`, `docs/api.md:59-110` | `POST /api/shape|run/start|watch|burst` → `501` in observe; audit/consult stay functional | **verified (repo)** |
| S4 | Repo: frozen corpus + seal | `docs/VERIFY.md:24-26` | 159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells (`audit-archive.py` OK); seal `docs/frozen-manifests.sha256` = 159 manifests | **verified (repo)** |
| S5 | Repo: ledger discipline (VERIFIED only on named command+host+date) | `docs/VERIFY.md:7-18`, `docs/VERIFY.md:83-91` | `VERIFIED`, `RECORDED`, `PENDING`, `BLOCKED`, `REFUTED`, `SUPERSEDED`; row moves only on named command | **verified (repo)** |
| S6 | Repo: six gates + selftest fault proofs | `docs/VERIFY.md:42`, `docs/CHECKPOINTS.md:114-132` | `python tools/verify.py` → six PASS; `--selftest` → six `caught` with controls | **verified (repo)** |
| S7 | Repo: thesis title + honesty invariants | `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:10-21`, `:40-68` | *Conception et évaluation d'un banc d'essai reproductible pour la comparaison de politiques AQM et BBR sur des liens d'accès contraints*, ISGEI × Météo Madagascar / DGM; invariants: no-load⇒no-grade, unmeasured⇒`Unavailable`, 220 ms = DGM operator objective not WMO mandate, BBR = BBRv1 | **verified (repo)** |
| S8 | Repo: DGM download scenario + CAP SLA model | `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:134-136`, `:151-155` | Download audit = bulk FROM sink through shaped `e-wan` = *"the actual DGM scenario (200 MB EUMETSAT downlink vs CAP egress)"*; Model C: loaded 16 KB completions at 220 ms, breach `< 95%` | **verified (repo)** |
| S9 | Repo: tariffs (Yas official, 31/08/2026) | `docs/data-prices.md:23`, `:33-40`, `:140-144`, `:148-156` | Yas Net Month 4.5 Go = 25 000 Ar (5 556 Ar/Go); FTTH IN 490 Ar/Go; GOLD 1 M Ar/mo 200 Mbps CIR; PLATINE 2.5 M Ar/mo 30 TB (81.38 Ar/GB derived); retail tiers table | **verified (repo, sourced to yas.mg)** |
| S10 | Repo: case-study contract = DGM risk posture | `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md:16-27`, `:29-38`, `:65-78` | Read-only only; never: capture/probe/load/write/root/login/scan; grant items a–d; pseudonym default; nothing executed without grant | **verified (repo)** |
| S11 | Repo: BLOCKED rows (live acceptance, eBPF, `.dist`, BBRv3, CHR, case study) | `docs/VERIFY.md:28-30`, `:52-53`, `:57`, `:60`; `docs/superpowers/plans/README.md:25-58` | Each names its unblock condition (Linux 6.8 host, kernel 7.1, ≥30-sample trace, QEMU+CHR, written grant) | **verified (repo)** |
| S12 | Repo: measurement protocol (profiles, gates, provenance chain) | `docs/methodology.fr.md:47-57`, `:59-66`, `:118-142`, `:219-238` | QR1–QR4; P1–P4 profiles; 23-col `aqm_eval.csv`; freeze → manifest → `hash8` triple provenance | **verified (repo)** |
| S13 | Repo: remediation (Linux tc + RouterOS scripts) | `docs/api.md:101`; `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:224-239` | `/api/remediate` emits Linux `tc` (CAKE 95%) and RouterOS v7 `kind=cake` strings; refuses zero/NaN capacity | **verified (repo)** |
| S14 | WMO WIS 2.0 | https://community.wmo.int/site/knowledge-hub/programmes-and-initiatives/wmo-information-system-wis/wis2-overview (fetched) | Operational 01/01/2025; MQTT publish-subscribe over public Internet (vs GTS private links); WIS2 Node = MQTT broker + HTTP server; Global Broker / Cache / Catalogue / Monitor; topic hierarchy `origin/a/wis2/<centre-id>/data/core|recommended`; WNM + WCMP2; `wis2box` = FOSS reference Node; registration via PR/NFP + centre-id + GISC assessment | **verified (primary: WMO)** |
| S15 | WMO warnings portal `warnings.wmo.int` | https://warnings.wmo.int/iwp/OPCMMS/en/ (fetch attempted) | Transport error / DNS failure in this environment; no content retrieved | `unverified — primary source not found` |
| S16 | CAP 1.2 alert format | https://docs.oasis-open.org/emergency/cap/v1.2/CAP-v1.2-os.html (fetched, OASIS Standard 01 July 2010) | §1.1 open, non-proprietary format for all-hazard alerts; §1.3 structure `<alert>` → `<info>` (urgency/severity/certainty triples) → `<resource>`/`<area>`; namespace `urn:oasis:names:tc:emergency:cap:1.2` | **verified (primary: OASIS)** |
| S17 | Dublin Descriptors, 2nd cycle (Master) | https://www.univ-tech.eu/dublin-descriptors (EUT+ page reproducing JQI 2004 + EQF 2008) (fetched) | 90–120 ECTS (≥60 at 2nd cycle); advanced knowledge allowing original thinking; problem-solving in new/unfamiliar environments; multidisciplinary integration; critical/ethical analysis; high autonomy | **verified (primary-citing: EUT+)** |
| S18 | ISGEI / MSTI framework (thesis PDF, jury rubric, word count, defence duration) | Web search returned only non-primary results (Facebook page, generic portals); no fetchable PDF/rubric | — | `unverified — primary source not found` |

> Rule applied throughout: §2–§3 never extend S14–S16 beyond what was fetched. CGO is a link measurement/bench tool; dissemination stays with wis2box-class tooling [S14] and CAP is the payload vocabulary [S16].

---

## 2. What the Master needs

**Anchor: Dublin Descriptors 2nd cycle [S17] + repo honesty invariants [S7].** ISGEI/MSTI specifics are `unverified — primary source not found` [S18] — not invented below.

### 2.1 Original contribution

| Dublin 2nd-cycle outcome [S17] | CGO mapping (primary source) | What counts as "original" |
|---|---|---|
| Advanced knowledge allowing original thinking | Frozen corpus (159/528/22) + seal [S4]; G0–G7 validity gates [S2]; 23-col provenance chain: raw → freeze → manifest → `hash8` [S12] | A *reproducible evidence engine* (method + frozen proof), not a new qdisc algorithm. FQ-CoDel/CAKE themselves are established IETF/Linux standards — novelty is claimed in measurement method and honesty invariants [S7]. |
| Problem-solving in new/unfamiliar environments | Four-namespace bench + client-side audit + direction-aware `down`/`both` [S8]; enterprise economics pivot: retail → GOLD/PLATINE/APN scenarios [S7, S9] | Transfer from constrained-link bench to operator questions (CAP SLA, waste cost Ar/h). |
| Multidisciplinary integration | Kernel netem/qdisc (`tc`), AQM standards, Malagasy tariff economics [S9], service APIs [S3] | Each strand anchored to a primary spec in §1, not to marketing. |
| Critical / ethical reflection | Invariants: no-load⇒no-grade, unmeasured⇒`Unavailable`, scenario≠claim, 220 ms = operator objective not WMO mandate, BBR = BBRv1 [S7]; case-study read-only guardrails [S10] | The thesis's strongest Master-level asset: a measured refusal to overclaim, testable in code. |
| Autonomy | Single-binary, no-DB system [S1]; six gates + selftest [S6] | Sole-author design, build, and ledger discipline demonstrable without external service. |

**Explicitly missing (obtain from ISGEI secretariat before defence):** required thesis word count, jury size/composition, rubric weights (code vs writing vs demo), defence language, appendix/software rules — `unverified — primary source not found` [S18].

### 2.2 Reproducibility (the defence exhibit)

1. **Ledger standard exceeds the descriptor floor.** Every `VERIFIED` row requires named command + host + date + saved output [S5]; `BLOCKED` rows name their unblock condition and stay blocked [S11].
2. **Self-proving.** `make verify` / `python tools/verify.py` runs six gates with fault proofs (`--selftest` six `caught`) [S6]; `python book/tools/audit-archive.py` prints both `OK` lines on the frozen corpus [S4].
3. **Air-gap safe.** All proof commands are local (Python + Go stdlib + `sha256sum`); dashboard runs single-binary with embedded React and no DB [S1]. No network needed in the room — a design property, not a demo hack.
4. **What reproducibility does *not* include.** Linux 6.8 live acceptance, eBPF object, BBRv3, `.dist`, CHR, IFB ingress, case study are all `BLOCKED` rows with named conditions [S11]. The jury re-runs the *proof*, not the blocked live runs.

### 2.3 Defence demo arc with falsification (5 minutes, air-gap-safe)

| Minute | Jury sees | Falsification move (jury invites to break it) | Source of truth |
|---|---|---|---|
| 0:00–1:00 | `python tools/verify.py` → six PASS | `--selftest`: each gate must catch its injected mutation | [S6] |
| 1:00–2:00 | `python book/tools/audit-archive.py` → both OK lines (159 / 528 / 22) | Mutate one CSV row in a temp copy → audit fails | [S4] |
| 2:00–3:00 | `sha256sum -c docs/frozen-manifests.sha256` → 159 OK | Rewrite one manifest + its file consistently → audit still passes, **seal catches it** | [S4] |
| 3:00–4:00 | `cgo verify` / `cgo figures` regenerate from frozen CSV; identical `hash8` in Wall, Résultats, Provenance, `report.md` | Edit one figure's source CSV → `hash8` must change | [S2, S12] |
| 4:00–5:00 | Fixture report without measured load renders `Unavailable`, never green | Feed zero-load fixture → verdict must refuse green | [S7] |

**Falsification principle:** the demo invites the jury to break one gate/seal/hook and watch it catch the mutation [S6]. A system that cannot be falsified in the room is theatre; this one is named, recorded, and re-runnable on the jury's host.

**Stays explicitly unproven in the room:** every `BLOCKED` row in [S11] — stated in the defence, not hidden. Naming them is the credibility signal under the Dublin critical-reflection outcome [S17].

---

## 3. What DGM needs — verified vs speculation

### 3.1 Verified operational needs (repo-grounded only; no invented specifics)

| # | Need | Evidence (repo primary) | Standard neighbour (fetched primary) |
|---|---|---|---|
| V1 | **Real-time alert/telemetry protection on constrained links.** Download audit models the real case: bulk FROM sink through shaped `e-wan` = 200 MB EUMETSAT downlink vs CAP egress [S8]. | Model C: loaded 16 KB completions at 220 ms, breach `< 95%` [S8]; 220 ms is the DGM operator objective, never a WMO mandate [S7]. | WIS 2.0 is the *transport* (MQTT pub-sub over public Internet, WIS2 Node = broker + HTTP, core data via Global Cache) [S14]. CAP 1.2 is the *payload* vocabulary (`<alert>`/`<info>` urgency/severity/certainty) [S16]. |
| V2 | **Cost-bounded bandwidth.** Waste cost in Ar/h per enterprise scenario, not raw throughput. | GOLD 1 M Ar/mo 200 Mbps CIR; PLATINE 2.5 M Ar/mo 30 TB; wholesale 81.38 Ar/GB; APN 150 Ar/GB quoted; retail 5 556 Ar/GB = emergency overage [S7, S9]. Scenario shown everywhere; primary marked by declared link type; tariff provenance stated in report [S7]. | — (national tariff research [S9]) |
| V3 | **Router-grade remediation at the edge.** Generated Linux `tc` + RouterOS scripts for the identified control point, built only from measured capacity. | `/api/remediate` + §7 enterprise rebuild [S13]; refuses zero/NaN capacity [S13]. CHR live proof is `BLOCKED` [S11] — verified as *requirement*, not capability. | — |
| V4 | **Read-only evidence without operational risk.** The operator's actual risk posture is the contract: read-only inspection only, seven prohibitions, grant a–d. | Case-study contract [S10]; slice status `BLOCKED` until written grant exists (`docs/VERIFY.md:60`) [S11]. | — |

**Correct product boundary (verified):** CGO = link auditor/bench [S1]; WIS 2.0 = transport [S14]; CAP = alert format [S16]. **Do not claim CGO is a WIS2 Node or CAP producer** — `wis2box` is the FOSS reference Node [S14]; CGO measures link behaviour.

### 3.2 Speculation (NOT verified — excluded from requirements)

Everything below is `unverified — primary source not found` (no fetched primary source in §1; repo names only the pseudonym "the national forecasting server" [S10]):

- DGM Ampandrianomby site topology, host/service inventory, `sadcnwp` details.
- Actual cadences, deadlines, or "late" definitions per service class A/B/C (the contract lists these as *grant deliverables to be defined by the operator*, not as facts [S10]).
- WIS2 Node registration status: centre-id, GISC assessment state, Global Broker subscription state (registration flow is documented [S14] but Madagascar's state was not established).
- Any WMO numeric per-alert threshold: WMO defines "real time" in the thesis framing as a few seconds to minutes [S7]; no 220 ms mandate exists in fetched WMO text [S14].
- Any number copied from operator dashboards into a figure without provenance + operator approval — explicitly prohibited by the contract [S10].
- `warnings.wmo.int` portal behaviour — `unverified — primary source not found` [S15].

If the jury or operator asks about any item above, the honest answer is the contract: *it depends on the written grant a–d* [S10], and until then nothing executes [S11].

---

## 4. Gap analysis (what the demo / DGM use needs that the repo does not yet prove)

| Gap | Evidence | Grade |
|---|---|---|
| ISGEI/MSTI rubric, word count, jury, defence language | No fetchable primary source [S18] | **Blocker for defence prep** — obtain from secretariat; Dublin mapping in §2.1 is the fallback framework only. |
| Live acceptance on Linux 6.8 (bench, IFB ingress) | `BLOCKED` rows [S11] | **Blocker for any live claim** — needs stable host + human-approved run. Demo stays air-gap/frozen-only (§2.3). |
| eBPF object, `.dist` calibration, BBRv3 matrix, CHR validation | `BLOCKED` / deferred rows [S11] | **Deferred** — each names its unblock condition; none is demo-critical. |
| Case-study grant (service map, workload model, SLA candidates) | `BLOCKED` until grant a–d [S10, S11] | **Blocker for any DGM-production claim** — contract defines deliverables but nothing executed. |
| WMO `warnings.wmo.int` specifics | Fetch failed [S15] | `unverified — primary source not found` — do not cite portal behaviour in the thesis. |
| RouterOS CAKE live validation | Generator exists [S13]; CHR proof `BLOCKED` [S11] | **Requirement verified, capability not** (see V3). |
| Operator security perimeter (auth, rate limit) | Hardening slice 28 is approved direction, pending implementation (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:1-6`) | **Out of Master-demo scope** — enterprise deployment concern, not a thesis-evidence concern. |

---

## 5. Target demonstration (the exact run-of-show)

**Setup:** one laptop, repo checkout at the defended HEAD, Python + Go installed, `book/` present. No network, no VM, no gateway touch. ≈5 minutes.

1. **Prove the ledger** — `python tools/verify.py` (six PASS) then `--selftest` (six `caught`) [S6]. Invitation: *"Break one."*
2. **Prove the corpus** — `python book/tools/audit-archive.py` (both OK lines) [S4].
3. **Prove the seal** — `sha256sum -c docs/frozen-manifests.sha256` (159 OK), then consistent rewrite caught by seal [S4].
4. **Prove provenance** — regenerate figures; show one `hash8` identical across Wall / Résultats / Provenance / `report.md`; change one CSV row and show the hash moves [S2, S12].
5. **Prove honesty** — load-less fixture → `Unavailable`, never green; 220 ms presented as DGM operator objective [S7], tariffs presented as contractual scenarios with provenance [S7, S9].
6. **Name the unproven** — list the `BLOCKED` rows and their conditions [S11] on the final slide. This is the Dublin critical-reflection exhibit [S17].

**Anti-demo rules:** no live bench, no CHR, no `.dist`, no eBPF, no operator hostname/IP, no number without a §1 source. Anything asked about that is not in §1 gets the `unverified — primary source not found` answer, with the grant path [S10] for operator questions.

---

## 6. Anti-toy criteria (pass/fail for the demonstration)

| Criterion | Pass condition (primary source) | Current state |
|---|---|---|
| Re-runnable proof | Jury re-runs gates + audit + seal on their host, air-gap, exit 0 [S4, S6] | **Pass** (air-gap floor) |
| Falsifiability | Each gate carries a fault proof that catches an injected mutation [S6] | **Pass** |
| No invented numbers | Every thesis figure cites a `VERIFY` row or is labelled `RECORDED`/hypothesis [S5]; tariffs labelled contractual scenarios [S7] | **Pass** — the ledger *is* the anti-toy asset |
| No overclaimed standards | 220 ms never presented as WMO mandate [S7]; BBR never presented as BBRv3 [S7]; WIS2/CAP never claimed as CGO output [S14, S16] | **Pass** |
| DGM specifics discipline | No site/topology/cadence invented; speculated items named and quarantined in §3.2 | **Pass iff §3.2 is respected in slides and text** |
| Degree-framework honesty | Dublin mapping is a fallback; ISGEI specifics marked `unverified — primary source not found` until obtained from the secretariat [S17, S18] | **Pass with caveat** — obtain rubric before defence |
| Toy signals absent | No synthetic data, no masked invalids (quarantine counted, never hidden) [S4, S12]; no green without measured load [S7] | **Pass** |
| Live-claim discipline | Any claim needing a `BLOCKED` condition is stated as blocked, never demoed [S11] | **Pass** — the BLOCKED list is shown, not hidden |

---

*Research only — no code modified, no commit made. Every claim above cites a §1 primary source; unknowns are marked `unverified — primary source not found` explicitly.*
