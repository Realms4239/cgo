# WIS2 Gates Q1–Q7 — Primary-source research

> Scope: close Q1–Q7 from `docs/runbooks/wis2-probe.md` §Gates, `design §7.7` (R11–R14).
> Method: **primary sources only** — WMO publications and WMO-hosted standards, `wis2box` official docs, OASIS CAP standard, official MQTT client docs/releases. No third-party architecture diagrams or blogs.
> All URLs below were fetched **2026-11-09** (UTC). Version/date stamps are those published at the source.
> Finding flags (runbook vocabulary):
> - **PROVEN-a** — claim directly supported by a fetched primary source on 2026-11-09.
> - **unverified** — no primary support found in this pass; blocks S4/WIS2 use until resolved.
> - **CONFLICT** — fetched primaries disagree with each other or with a prior R-claim in the design/runbook; builder must present both before choosing.

No code, no broker, no client was loaded or installed in this pass. This file is the sole output.

---

## Q1 — WMO-No. 1060 Vol. II App. D (Topic Hierarchy) + App. E (WNM, WCMP2)

**Requirement** (`wis2-probe.md:163`): retrieve hierarchy, WNM, WCMP2 schemas — R1 links them but does not reproduce them. Status was `unverified [R11]`.

**Primary sources fetched**

| Item | URL | Retrieval | Version stamp at source |
|---|---|---|---|
| WIS2 Topic Hierarchy (STABLE) | `https://wmo-im.github.io/wis2-topic-hierarchy/standard/wis2-topic-hierarchy-STABLE.html` | 2026-11-09 | v1.3.0, 2026-06-15, STABLE. Header: *"Note: This section … is the same as Appendix D in the Manual on the WMO Information System (WMO-No. 1060), Volume II as approved by EC-78."* Revision table: 1.0.0 Jun 2024 EC-78 → 1.1.0 Nov 2024 FT2024-2 → 1.2.0 Feb 2026 FT2025-2 → 1.3.0 Jun 2026 FT2026-1 |
| WIS2 Notification Message (WNM, STABLE) | `https://wmo-im.github.io/wis2-notification-message/standard/wis2-notification-message-STABLE.html` | 2026-11-09 | v1.3.0, 2026-06-15, STABLE. Header: *"same as Appendix E … as approved by EC-78."* |
| WMO Core Metadata Profile 2 (WCMP2, STABLE) | `https://wmo-im.github.io/wcmp2/standard/wcmp2-STABLE.html` | 2026-11-09 | v2.3.0, 2026-06-15, STABLE. Header: *"same as Appendix F … as approved by EC-78."* (Note: Appendix F, not E, for WCMP2 in current edition — see CONFLICT note below) |
| WIS2 Overview (Knowledge Hub) | `https://community.wmo.int/site/knowledge-hub/programmes-and-initiatives/wmo-information-system-wis/wis2-overview` | 2026-11-09 | Last update 27 Jun 2025 / page footer 29 Oct 2026; operational phase 01/01/2025 |
| Codes Registry — earth-system-discipline | `https://codes.wmo.int/wis/topic-hierarchy/earth-system-discipline` | 2026-11-09 | stable, changed 06 Jan 2026, version info 10 |
| Codes Registry — centre-id | `https://codes.wmo.int/wis/topic-hierarchy/centre-id` | 2026-11-09 | stable, changed 06 Jan 2026, version info 185 |

**Key excerpts / cited claims**

- **PROVEN-a — Hierarchy shape (7 levels).** WTH §7 Table 2: Level 1 channel (`origin` for data providers, `cache` for Global Services), 2 version (currently `a`), 3 system (`wis2`), 4 centre-id, 5 notification-type (`data` or `metadata`), 6 data-policy (`core` or `recommended`), 7 earth-system-discipline (one of `atmospheric-composition | climate | cryosphere | hydrology | ocean | space-weather | weather`). Levels 8+ are sub-discipline specific, proposed by domain communities, hierarchical. Examples: `origin/a/wis2/ca-eccc-msc/data/core/weather/surface-based-observations/synop`, `cache/a/wis2/ca-eccc-msc/data/core/weather/prediction/forecast/medium-range/probabilistic/global`.
- **PROVEN-a — Publishing rules.** Requirement /req/core/publishing: (A) Data SHALL NOT be published with a topic not defined in the WTH; (B) Data SHALL be published to at least level 8 (sub-discipline or beyond); (C) Metadata SHALL be published to exactly `.../metadata` (level 5 = `metadata`, no levels 6–7). Recommendation: SHOULD publish to a leaf topic; `experimental` SHOULD be used temporarily until approved. Permission: MAY publish with `experimental` + unapproved sub-topics, MAY publish to non-leaf (but discouraged: *"pollution of data messages"*).
- **PROVEN-a — centre-id format.** §1.6 /req/core/centre-id: SHALL be `tld-centre-name` where `tld` is IANA TLD for country/organisation, `centre-name` based on centre name; SHALL NOT be used by more than one Node/Global Service; `test` TLD only for internal testing. Conventions: lowercase, IRA T.50, no dots, dash-separated.
- **PROVEN-a — WNM core.** WNM §7 Table 2: required `id` (UUID), `type="Feature"`, `conformsTo` (or deprecated `version="v04"`), `geometry` (Point|Polygon|null, WGS84, deg), `properties.pubtime` (RFC3339 UTC, set by originator, SHALL NOT be modified by intermediaries), `properties.data_id` (unique in dataset scope), `links` (≥1 link, exactly one `rel=canonical|update|deletion`, href scheme http/https/ftp/sftp). Optional: `metadata_id`, `producer`, `global-cache`, `datetime` or `start_datetime+end_datetime`, `cache` (default `true`), `integrity` (sha256/sha384/sha512/sha3-*), `content` (inline data), plus arbitrary additional properties allowed.
- **PROVEN-a — Size limits.** WNM §1.2 /req/core/message_size: WNM SHALL NOT exceed **8192 bytes**. §1.16 /req/core/content: inline `content.value` SHALL NOT be provided if encoded size > **4096 bytes**; MAY be provided if <4096 (encoding `utf-8|base64|gzip`, `size` = original bytes).
- **PROVEN-a — WCMP2 core.** WCMP2 §1.1 Table 2: required `id` as `urn:wmo:md:{centre_id}:{local_identifier}`, `type="Feature"`, `conformsTo` = `http://wis.wmo.int/spec/wcmp/2/conf/core`, `properties.type` (code from `codes.wmo.int/wis/resource-type`, usually `dataset`), `properties.title`, `properties.description`, `properties.themes` (≥1 theme, each with `scheme` + ≥1 concept; dataset SHALL list all applicable Earth-system disciplines from WTH), `geometry` (WGS84 Polygon|Point|null), `time` (date|timestamp|interval|null, ISO8601), `properties.contacts` (≥1, SHALL include `organization`), `properties.created`, `properties.wmo:dataPolicy` (core|recommended, REQUIRED for datasets), `links` (≥1). Granularity: §1.6: *"A **dataset** is the primary granularity of a WCMP record describing a collection of data granules"*; §6.1.3: granules described via STAC/API, not separate WCMP records.
- **PROVEN-a — Overview cross-check.** WIS2 Overview: `origin/a/wis2/<centre-id>/data/core` implies free unrestricted access (Resolution 1), downloaded by Global Cache and re-announced as `cache/a/wis2/<centre-id>/data/core`; `.../data/recommended` may require access control, NOT downloaded by Global Caches; `origin/a/wis2/<centre-id>/metadata` carries WCMP2 canonical link, Global Discovery Catalogue updated from it; WNM contains canonical HTTP(S) link.

**Open question that remains (Q1)**

- Which exact pinned file+version does the builder treat as "WMO-No. 1060 Vol. II App. D/E" for implementation — the printed Manual edition or the online STABLE files that claim *"same as Appendix D/E"* but include later fast-track amendments (FT2024-2, FT2025-2, FT2026-1)? Implementation needs a frozen version pin (WTH v1.3.0 + WNM v1.3.0 + WCMP2 v2.3.0, all 2026-06-15) + schema bundle URLs (`schemas.wmo.int/wth/a`, `/wnm/1.1.0/schemas/wis2-notification-message-bundled.json`, `/wcmp/2.0.0/schemas/wcmp2-bundled.json`), not a "latest" link.

**Recommended answer**

- Treat WTH v1.3.0 / WNM v1.3.0 / WCMP2 v2.3.0 STABLE as the citable Q1 closure (PROVEN-a). Implement topic construction strictly as 7 primary levels + L8 leaf (or `experimental`), `metadata` exactly at notification-type, WNM validation against bundled JSON Schema with 8192/4096 checks, WCMP2 `urn:wmo:md:` id. Keep `cgo/v1/...` lab namespace separate — it violates WTH (no valid centre-id/discipline leaf) and therefore **must not** be presented as WIS2 (PROVEN-a: would fail /req/core/publishing A+B).
- CONFLICT to flag: design/runbook cite "App. E (WNM, WCMP2)" as one appendix. Fetched primaries place WNM in App. E and WCMP2 in **App. F** (both "same as … as approved by EC-78" headers). Cite Appendix F for WCMP2 going forward; note the runbook's "App. E" shorthand is outdated.
- CONFLICT to flag: "same as Appendix D/E as approved by EC-78" vs current STABLE v1.3.0 includes 2024–2026 fast-track revisions. The online files are the operative approved versions, not a verbatim copy of the first EC-78 print. Pin by URL + version + date in code/README.

---

## Q2 — WMO-No. 1061 §2.6.1.1 (Node registration / decommissioning)

**Requirement**: normative registration steps + assessment + decommissioning (R1 only summarises). Status was `unverified [R11]`.

**Primary sources fetched**

| Item | URL | Retrieval | Version stamp |
|---|---|---|---|
| Guide WMO-No. 1061 Vol. II (APPROVED) | `https://wmo-im.github.io/wis2-guide/guide/wis2-guide-APPROVED.html` | 2026-11-09 | v1.3.0, 2026-06-15, APPROVED; WMO publication `https://library.wmo.int/idurl/4/69130` |
| WIS2 Overview — Registration section | `https://community.wmo.int/site/knowledge-hub/programmes-and-initiatives/wmo-information-system-wis/wis2-overview` | 2026-11-09 | § "WIS2 Node Registration" + Notes 1–6 |

**Key excerpts / cited claims**

- **PROVEN-a — Approval authority.** Guide §2.6.1.1: *"The registration and decommissioning of WIS2 Nodes must be approved by the Permanent Representative (PR) with WMO of the country or territory where the WIS centre is located. The WIS National Focal Point (NFP) can register a WIS2 Node on behalf of the PR for an official NC or DCPC listed in the Manual on WIS, Volume I. Where the WIS2 Node is part of a DCPC, the sponsoring WMO programme or regional association shall be consulted."* Project/campaign Nodes for limited time may be registered by NFP in coordination with Secretariat. A Node may publish on behalf of other centres (DCPC role), using producers' centre-ids.
- **PROVEN-a — 7-step registration sequence.** (1) *Request to host*: by NFP (or PR / president of international org for DCPCs); (2) *Assign centre-id*: acronym proposed by Member, endorsed by WMO Secretariat (`tld-centre-name`, see §2.6.1.2); (3) *Complete the WIS2 register*: NFP completes register maintained by Secretariat; (4) *Provide Global Service details*: Secretariat provides IP/connection details so Node can be configured; (5) *WIS2 Node assessment*: principal GISC verifies (a) topics comply with WTH, (b) notifications comply with WNM, (c) data server correctly configured/functioning, (d) Message Broker correctly configured/functioning; (6) *Add new centre*: GISC notifies Secretariat that centre can be added; (7) *Communicate to Global Services*: Secretariat provides Node details to Global Brokers to subscribe. Dataset registration follows only after Node registration: publish discovery metadata → validation by Global Discovery Catalogues → Brokers auto-subscribe to listed topics → then exchange data.
- **PROVEN-a — Operational precondition (Notes).** Overview Notes: Note 1 — ONLY PR/NFP submissions considered; Note 2 — *"ONLY an operational WIS2 Nodes can be registered … To be considered operational, your WIS2 Node needs [to] have a publicly available MQTT-endpoint that is regularly publishing WIS2 Notifications to pass the initial WIS2 Node assessment"*; Note 3 — if not yet operational, contact principal GISC before submitting form; Note 4 — centre-id may be pre-registered in WMO Codes Registry before Node is operational (contact wis2-support@wmo.int); Note 5 — Secretariat informs GISC to start assessment; Note 6 — on pass, Secretariat notifies Global Brokers, Node listed in Operational WIS2 Nodes.
- **PROVEN-a — centre-id guidance (§2.6.1.2).** Two dash-separated tokens (Token 1 IANA TLD with overseas-territory rule, Token 2 recognisable centre name, e.g. `uk-metoffice` over `ukmo`, avoid GTS CCCC); multiple centres SHOULD have separate ids (`uk-metoffice-nmc`, `-vaac`, `-global-cache`); test Nodes add `-test`; system names discouraged (functional designations durable).
- **PROVEN-a — Decommissioning.** Guide §2.6.1.1 final para: *"When decommissioning a WIS2 Node, operators must ensure that obligations relating to data sharing within WIS continue to be met after the WIS2 Node is decommissioned, for example, by migrating the data sharing obligations to another WIS2 Node. In the case of a DCPC, this may mean transferring the responsibilities to another Member."* No separate "unregistration form" is described in §2.6.1.1 beyond this migration obligation + PR approval.
- **PROVEN-a — Auth for registration.** §2.6.1.3: Global Brokers authenticate to Node broker with username/password; credentials provided to Secretariat at registration, shared with Global Service operators, stored in WIS register, *"should not be considered confidential or secret"*; brokers on fixed IPs (IP-filtering possible); data-server access may be restricted to Global Caches if core-only.

**Open question that remains (Q2)**

- None at the primary-source level. The procedure text is fully fetched. What remains is organisational: who is the Madagascar NFP/PR, which GISC is principal, and whether the project will even seek registration — see Q4. No additional WMO text is needed to close Q2 as a research gate.

**Recommended answer**

- Close Q2 as PROVEN-a on the Guide §2.6.1.1 text + Overview Notes 1–6. Builder citation: `Guide §2.6.1.1 + §2.6.1.2 + §2.6.1.3 (APPROVED v1.3.0, 2026-06-15)`. Unregistration = PR-approved decommission + migration of obligations to another Node/Member, not a silent disconnect.
- CONFLICT to flag: runbook design §7.8 says "only for an *operational* Node with public MQTT regularly publishing" — primaries confirm this ordering (Note 2) but add that centre-id **can** be pre-reserved in the Codes Registry before operational. "Broker first, paperwork after" is therefore true for full registration, false for centre-id reservation. Keep both statements distinct in future slices.

---

## Q3 — CAP 1.2 granularity as WIS2 topic (QoS / retain)

**Requirement**: are CAP documents carried as WIS2 data granules, and at what QoS/retain? Runbook separation (CAP via HTTP canonical link, not MQTT payload) was `PROPOSED`, `unverified [R12]`.

**Primary sources fetched**

| Item | URL | Retrieval |
|---|---|---|
| Guide §2.5.1 MQTT protocol config + §2.3 Table 1 granularity + §1.3.3.4 + §2.6.2.1 | `https://wmo-im.github.io/wis2-guide/guide/wis2-guide-APPROVED.html` | 2026-11-09 |
| WNM §1.16 content + §1.17 links | `https://wmo-im.github.io/wis2-notification-message/standard/wis2-notification-message-STABLE.html` | 2026-11-09 |
| wis2box — Publishing CAP Alerts | `https://docs.wis2box.wis.wmo.int/en/latest/user/cap-alerts.html` | 2026-11-09 |
| wis2box — Data pipeline plugins (`CAPMessageData`) | `https://docs.wis2box.wis.wmo.int/en/latest/reference/running/data-pipeline-plugins.html` | 2026-11-09 |
| wis2box — Adding datasets (template `weather/advisories-warnings`) | `https://docs.wis2box.wis.wmo.int/en/latest/user/setup-datasets.html` | 2026-11-09 |
| CAP 1.2 OASIS Standard | `https://docs.oasis-open.org/emergency/cap/v1.2/CAP-v1.2-os.html` | 2026-11-09 |

**Key excerpts / cited claims**

- **PROVEN-a — Granularity table.** Guide §2.3 Table 1: WCMP2 granularity = **Datasets** (Global Discovery Catalogue), WTH granularity = **Dataset granules** (Global Broker, Nodes), WNM granularity = **Dataset metadata, dataset granules** (Global Broker, Nodes). A CAP document is therefore a *data granule* (one file/object), not a dataset, not a separate WIS2 message type.
- **PROVEN-a — MQTT parameters (normative).** Guide §2.5.1: *"MQTT v3.1.1 and v5.0 are the chosen protocols … MQTT v5.0 is preferred for connecting to Global Brokers as it provides additional features such as shared subscriptions. The following parameters are to be used for all MQTT client/server connections and subscriptions: Message retention: false; Quality of Service (QoS) of 1; A maximum of 2000 messages to be held in a queue per client."* Plus: username/password auth, secure MQTT preferred (valid cert), standard ports 8883 (MQTTS) / 443 (WSS). §2.5.2: HTTPS port 443 for downloads.
- **PROVEN-a — Two carrier options for CAP XML, both as granules.** wis2box CAP docs: dataset template **`weather/advisories-warnings`** selects `CAPMessageData` plugin (*"validates it against the CAP v1.2 schema … optionally verifies the digital signature … using capvalidator"*); ingest via `wis2box-incoming` bucket upload OR via WMO CAP Composer → MQTT forward to wis2box (broker host/port/user/pass + dataset ID), then signed/published to public storage. Resulting WNM(s) carry a canonical HTTP(S) link to the CAP XML file in public storage (`WIS2BOX_URL/data/...`); subscribers download the XML over HTTP. This matches the runbook's "CAP via canonical link, not as WNM payload" separation.
- **PROVEN-a — Optional inline embedding (≤4096 B).** Guide §1.3.3.4 + WNM §1.16: data MAY be embedded in `properties.content` (utf-8/base64/gzip) if encoded size ≤4096 B, in addition to being published via data server; URL in notification SHALL refer to the newly added granule (not whole dataset). Guide §2.6.2.1 explicitly lists *"Publishing data embedded in a message (for example, Common Alerting Protocol (CAP) warnings)"* as a required Node performance capability. So a small CAP XML may travel **both** as inline `content` and as canonical-link file; a large CAP XML **must** be link-only.
- **PROVEN-a — Topic discipline for CAP.** Guide §1.3.3.4: each dataset mapped to **one and only one** topic; if no approved sub-topic fits, use `.../data/{earth-system-discipline}/experimental/...` (level 8 `experimental` per WTH). wis2box default for CAP is `weather/advisories-warnings` — i.e. discipline `weather`, sub-discipline `advisories-warnings` (L8). Whether that L8 value is in the current approved codelist bundle is determined by the WTH bundle (`schemas.wmo.int/wth/a`); fallback is `weather/experimental/...` until approved.
- **PROVEN-a — Retention vs QoS distinction.** Global **Cache** retention (§2.4.3): *"shall retain a copy of the core data … not less than 24 hours"* (data at rest on HTTP cache). MQTT **retain flag** (§2.5.1) is `false` (no retained MQTT message). WNM `properties.cache` (STABLE): default `true` (should be cached); `false` means Global Caches do not cache the canonical link and republish the notification with the origin link unmodified. These are three different "retention" knobs — do not conflate.

**Open question that remains (Q3)**

- Does the operator want `core` or `recommended` data-policy for the CAP dataset? Primaries do not fix it: wis2box template selection exposes `WMO Data Policy can be selected … If "recommended" is selected, a license link is required` (setup-datasets). Core implies Global Cache ingestion (24 h) + free unrestricted; recommended implies access-control description + no Global Cache ingestion. This is a signed operator choice, not a fetchable fact.
- Is the L8 `advisories-warnings` value approved in WTH v1.3.0 bundle, or must the probe use `weather/experimental/...`? Requires bundle check at implementation time (`codes.wmo.int/wis/topic-hierarchy/...` + `schemas.wmo.int/wth/a/wth.zip`); trivially closable, not a research blocker.

**Recommended answer**

- Close Q3 as PROVEN-a: CAP = WIS2 **data granule** in a `weather/advisories-warnings` (or `weather/experimental/...`) dataset, announced by WNM with canonical HTTP(S) link (primary carrier) + optional 4096-B inline `content`, published at **MQTT QoS 1, retain=false, max 2000 queued/client** per Guide §2.5.1.
- **CONFLICT (runbook vs primary):** runbook §Mosquitto profile proposed `retain_available true` + `max_qos 1` (engineer's choice, *"R1 states no QoS value"*). The fetched Guide §2.5.1 **does** state QoS 1 and retention false. The runbook's `retain_available true` should be reversed for any WIS2-facing broker; it may remain only for the lab-only `cgo/v1/...` broker (not WIS2, different guarantee).
- CONFLICT to flag: runbook mapping table listed CAP XML as "HTTP download URL referenced from WNM canonical link, PAS un payload MQTT" (PROPOSED separation). Primaries confirm the canonical-link separation as the default, **but** explicitly allow ≤4096 B inline embedding. The runbook position is therefore *mostly* right, *partly* incomplete — update the separation note to "link-primary, inline-optional-if-small".
- Data-policy (core vs recommended) stays **unverified** until operator signs Q5-adjacent dataset policy.

---

## Q4 — Operator centre-id status / written non-registration decision

**Requirement**: obtain operator centre-id status or a written decision not to register. Determines whether the design's `origin/a/wis2/...` mapping is ever used or the lab namespace is permanent. Status was `unverified [R13]`.

**Primary sources fetched**

| Item | URL | Retrieval | Version |
|---|---|---|---|
| Codes Registry — centre-id register (full list) | `https://codes.wmo.int/wis/topic-hierarchy/centre-id` | 2026-11-09 | stable, changed 06 Jan 2026, version info 185 |
| Guidance on assigning centre-id | `https://community.wmo.int/site/knowledge-hub/guidance-assigning-centre-identifier-wis2-node` | 2026-11-09 | Excerpt of Guide §2.6.1.2, last update 07 Mar 2025 |
| Operational WIS2 Nodes (Global Registry dashboard) | `https://community.wmo.int/site/knowledge-hub/programmes-and-initiatives/wmo-information-system-wis/operational-wis2-nodes` | 2026-11-09 | Last updated 29 Oct 2026; table/map *"automatically updated whenever a new Operational WIS2 Node is recorded"* |
| Guide §2.6.1.2 (centre-id) | `https://wmo-im.github.io/wis2-guide/guide/wis2-guide-APPROVED.html` | 2026-11-09 | v1.3.0 APPROVED |

**Key excerpts / cited claims**

- **PROVEN-a — Registry contents on 2026-11-09.** The centre-id register lists ~180+ stable entries (ae-ncm, ag-antiguamet, … za-weathersa, zm-zmd, plus Global Service suffixes like `int-ecmwf`, `eu-eumetnet-warnings`). On a scrape of the full Contents table, **no `mg-*` entry** (Madagascar ccTLD `mg`) and no `cv-*` neighbour confusion was found. Entries around the `m` range are e.g. `ml-malimeteo`, `mn-namem`, `mr-onm`, `ms-metservice`, `mv-mms`, `mw-dccms`, `mx-smn`, `my-metmalaysia`, `mz-inam` — with `mg-` absent between them.
- **PROVEN-a — What absence means (and does not).** Per Guide §2.6.1.2 / Overview Note 4, a centre-id **may be pre-registered** in the Codes Registry before the Node is operational, and per §2.6.1.1 full registration requires PR approval + GISC assessment + operational MQTT publishing. Absence from the Codes Registry therefore proves *"no published centre-id as of 06 Jan 2026 snapshot"* — it does **not** prove no application is in flight, no approved-but-unpublished id, and does **not** prove the operator has decided not to register.
- **PROVEN-a — Operational Nodes list is dynamic.** The Operational WIS2 Nodes page table/map is JS-rendered and auto-updated from the Global Registry; the markdown fetch returns only the shell (no row data). It therefore cannot be cited as positive/negative evidence for any single centre in this pass.
- **PROVEN-a — Format a future id would take.** `mg-{name}` (lowercase, IANA `mg` TLD, recognisable name; dashes allowed in Token 2; `-test` suffix for test Nodes). Examples in guidance: `dz-meteoalgerie`, `sg-mss`, `int-ecmwf`.

**Open question that remains (Q4)**

- **Everything substantive.** No primary source fetched in this pass can answer "does DGM/Madagascar hold or want a centre-id?" That answer lives with the PR/NFP and WMO Secretariat, not in a spec. Q4 needs a human artefact: either (a) a centre-id string + Codes Registry URL + GISC assessment record, or (b) a dated written non-registration decision signed by the operator/humain. Neither can be produced by a fetch.
- Sub-question: if a centre-id **is** in flight but unpublished, the Codes Registry scrape is insufficient — need Secretariat confirmation.

**Recommended answer**

- Record Q4 as **unverified** (primary-source negative finding only, not closure). Positive finding that **is** PROVEN-a: *no `mg-*` centre-id was published in the Codes Registry snapshot of 06 Jan 2026 (fetched 2026-11-09)*. Do not promote this to "Madagascar has no centre-id" — that would be a toy inference.
- Builder implication (unchanged from runbook, now PROVEN-a backed): keep `cgo/v1/probe/<site-id>/...` as the only publishable namespace in any slice; `origin/a/wis2/...` stays forbidden until Q4 closes with a citable centre-id + GISC record. Pre-register path (Codes Registry before operational) exists, but still requires PR/NFP action.
- Blocker: needs human/operator input. Research cannot close Q4.

---

## Q5 — Operator-signed severity bands (`cap_ok_pct` → urgency / severity / certainty)

**Requirement**: operator-signed mapping from link metric to CAP `<urgency>`/`<severity>`/`<certainty>`. R2 defines domains but no mapping; unsigned bands = invented thresholds. Status was *"signature, pas un fetch"*.

**Primary sources fetched**

| Item | URL | Retrieval |
|---|---|---|
| CAP 1.2 OASIS Standard (01 Jul 2010) | `https://docs.oasis-open.org/emergency/cap/v1.2/CAP-v1.2-os.html` | 2026-11-09 |
| Guide §2.6.2.1 (CAP as embedded-message example) | `https://wmo-im.github.io/wis2-guide/guide/wis2-guide-APPROVED.html` | 2026-11-09 |
| wis2box — Publishing CAP Alerts + CAPMessageData validation | `https://docs.wis2box.wis.wmo.int/en/latest/user/cap-alerts.html` + `https://docs.wis2box.wis.wmo.int/en/latest/reference/running/data-pipeline-plugins.html` | 2026-11-09 |

**Key excerpts / cited claims**

- **PROVEN-a — Value domains (CAP 1.2 §3.2.2, required `<info>` elements).**
  - `<urgency>` ∈ {`Immediate` (respond immediately), `Expected` (respond soon, within next hour), `Future` (respond in near future), `Past` (no longer required), `Unknown`}.
  - `<severity>` ∈ {`Extreme` (extraordinary threat to life/property), `Severe` (significant threat), `Moderate` (possible threat), `Minor` (minimal to no known threat), `Unknown`}.
  - `<certainty>` ∈ {`Observed` (determined to have occurred/ongoing), `Likely` (p > ~50%), `Possible` (p ≤ ~50%), `Unlikely` (p ~ 0), `Unknown`}. Deprecated `Very Likely` treated as `Likely`.
  - The three *"collectively distinguish less emphatic from more emphatic messages."* Related: `<category>` includes `Infra` (utility, telecom, other non-transport infrastructure); `<status>` ∈ {Actual, Exercise, System, Test, Draft}; `<msgType>` ∈ {Alert, Update, Cancel, Ack, Error}; `<scope>` ∈ {Public, Restricted, Private}; `<sender>` globally unique (may be domain-based); `<identifier>` unique, no spaces/commas; `<sent>` DateTime with `-00:00` for UTC. Producer conformance §4.3, consumer §4.4.
- **PROVEN-a — No numeric mapping exists in any fetched primary.** Neither CAP 1.2 nor the WIS2 Guide nor the wis2box CAP docs define thresholds from a link-quality percentage to urgency/severity/certainty. wis2box only validates XML against the CAP v1.2 schema (via `capvalidator`) and optionally checks the signature (`CHECK_CAP_SIGNATURE=True`); it does not generate the bands. The Guide mentions CAP only as an example of embeddable content. Any `cap_ok_pct` → band table would therefore be new (operator) policy, not a WMO/OASIS rule.
- **PROVEN-a — CAP `Alert` semantics.** `<msgType>Alert` = *"Initial information requiring attention by targeted recipients."* Routine telemetry with no breach therefore still has no CAP message (runbook hard boundary stands). Healthy probe → no CAP is consistent with CAP semantics, now PROVEN-a.
- **PROVEN-a — Authoring tool named by WMO.** wis2box CAP docs name the **WMO CAP Composer** (`https://github.com/World-Meteorological-Organization/cap-composer`, docs `https://cap-composer.readthedocs.io`) as the tool to create/sign CAP XML, forwardable to wis2box over MQTT per dataset. (Informative for the builder's choice-presenting role; the probe is not the Composer.)

**Open question that remains (Q5)**

- Which numeric band boundaries (in `cap_ok_pct` space, plus any dwell/hysteresis rule) does the operator sign? Example shape requiring signature: `≥95%` → no CAP (healthy, JSON only); `<95%` → `Alert` with operator-chosen triple per sub-band; recovery → `Update`+`<references>`, retraction → `Cancel`. Also: `<category>` (`Infra` proposed), `<scope>` (`Private`/`Restricted`, never `Public` without sign-off), `<sender>` domain, and whether `<certainty>` is always `Observed` (measured breach) vs `Likely`.
- No fetch can sign this. It is a human signature gate by definition.

**Recommended answer**

- Record domains as PROVEN-a (cite CAP 1.2 §3.2.1/§3.2.2 above). Record the **mapping itself** as **unverified** — deliberately not proposed as a citable rule.
- Provide the builder a *presentation* table to get the signature, with every cell marked PROPOSED—UNSIGNED, e.g. (illustrative only, **not a recommendation to implement**): breach depth (e.g. 85–95 / 70–85 / <70 %) → (`Expected` vs `Immediate`, `Moderate` vs `Severe` vs `Extreme`, `Observed` constant beyond first measured breach). State explicitly that shipping any such table without the operator's dated signature repeats the toy the anti-toy criteria forbid.
- Blocker: needs operator signature. Research cannot close Q5.

---

## Q6 — `wis2box` docs evaluation (install, payloads, CAP support)

**Requirement**: evaluate `wis2box` docs to decide probe → `wis2box` vs probe → own broker. R1 names it only. Status was `unverified [R14]`.

**Primary sources fetched**

| Item | URL | Retrieval |
|---|---|---|
| Landing / What is wis2box | `https://docs.wis2box.wis.wmo.int/` (via index) | 2026-11-09 |
| Getting started (requirements, network, system) | `https://docs.wis2box.wis.wmo.int/en/latest/user/getting-started.html` | 2026-11-09 |
| Installation and configuration (1.3.0) | `https://docs.wis2box.wis.wmo.int/en/latest/user/setup.html` | 2026-11-09 |
| How wis2box works (container topology) | `https://docs.wis2box.wis.wmo.int/en/latest/reference/how-wis2box-works.html` | 2026-11-09 |
| Adding datasets (WCMP2, topics, templates) | `https://docs.wis2box.wis.wmo.int/en/latest/user/setup-datasets.html` | 2026-11-09 |
| Ingesting data | `https://docs.wis2box.wis.wmo.int/en/latest/user/data-ingest.html` | 2026-11-09 |
| Data pipeline plugins (incl. `CAPMessageData`) | `https://docs.wis2box.wis.wmo.int/en/latest/reference/running/data-pipeline-plugins.html` | 2026-11-09 |
| Publishing CAP Alerts | `https://docs.wis2box.wis.wmo.int/en/latest/user/cap-alerts.html` | 2026-11-09 |
| WIS2 (framework pointer) | `https://docs.wis2box.wis.wmo.int/en/latest/reference/wis2.html` | 2026-11-09 |
| Guide §2.6.3 (wis2box as reference) | `https://wmo-im.github.io/wis2-guide/guide/wis2-guide-APPROVED.html` | 2026-11-09 |

**Key excerpts / cited claims**

- **PROVEN-a — Identity + version.** *"WIS2 in a box (wis2box) is a Reference Implementation of a WIS2 Node"* / *"FOSS Reference Implementation … designed to help accelerate international data and metadata exchange on WIS 2.0"* (Guide §2.6.3 + landing). Fetched docs are **v1.3.0** (`wis2box-setup-1.3.0.zip` from `github.com/World-Meteorological-Organization/wis2box-release/releases/tag/1.3.0`); Guide §2.6.3 lists functions: config/generation/publication of data+metadata compliant to Node requirements, MQTT broker + notifications, HTTP object storage + raw access, station/metadata curation UIs, OGC API server, extensible pipelines, metrics, access control for recommended, subscribe/download from Global Services, modular design. Docs: `https://docs.wis2box.wis.wmo.int`.
- **PROVEN-a — Install shape.** Linux (Ubuntu 20.04/22.04/24.04 recommended; other OS *"not officially supported"*), Python ≥3.8, Docker Engine ≥20.10.14, Docker Compose ≥2.0; download `wis2box-setup-1.3.0.zip`, unzip, `python3 wis2box-create-config.py` (produces `wis2box.env` in an absolute path dir mapped to `/data/wis2box` in `wis2box-management`), then `python3 wis2box-ctl.py start` (first run creates `docker-compose.images-1.3.0.yml`, pulls images) + `python3 wis2box-ctl.py status`. Verify via web-proxy `WIS2BOX_URL/oapi`, `WIS2BOX_URL/`, `WIS2BOX_URL/data/`.
- **PROVEN-a — Runtime topology (Docker Compose stack, 11 services on `status`).** Storage: `wis2box-minio` (S3, buckets `wis2box-incoming` → `wis2box-public` proxied at `/data/` over HTTP(S)) + Elasticsearch (9200/9300); Brokers: **internal** Mosquitto + **public** Mosquitto (host `1883`, `8884` for public; internal `1883` inside network); Data Management: `wis2box` + `wis2box-api` (ES entrypoint), `wis2box-auth`, `wis2box-ui` (nginx), `wis2box-webapp` (4173); Observability: Grafana (3000), Prometheus (9090), Loki (3100), `mqtt_metrics_collector` (8000/8001). Web-proxy: nginx on 80/443 (reverse-proxy, HTTPS/WS termination; MQTT via WS at `/mqtt`). Tech table (how-wis2box-works): Mosquitto (MQTT), MinIO+Elasticsearch (S3), ecCodes/csv2bufr/bufr2geojson/synop2bufr/OWSLib/pygeometa (WCMP/WMDR), pygeoapi (OGC API), Vue+Leaflet (UI). Follows Twelve-Factor App methodology.
- **PROVEN-a — Network / sizing / public-Internet posture.** Getting-started: needs GitHub/GHCR/Docker Hub + `iana.org`, `codes.wmo.int`, `oscar.wmo.int` reachable; host ports `80/tcp (HTTP)`, `1883/tcp (MQTT)`, `3000/tcp (Grafana)`, `9000/9001/tcp (MinIO/Console)` — **public-route only `80` and `1883`** (recommend NGINX reverse proxy + HTTPS, ACLs, DMZ/isolation, *"not recommended to use a host within your internal network"*). Minimum **2 vCPUs, 4 GB RAM, 24 GB local storage** (e.g. AWS `t3a.medium` 0–2000 obs/day ~US$35/mo, `t3a.large` 2000–10000 obs/day ~US$70/mo). Two pre-configured MQTT users: `everyone/everyone` read-only `origin/a/wis2/#` (for Global Broker/public) and `wis2box/<WIS2BOX_BROKER_PASSWORD>` read/write `origin/a/wis2/#` (internal publish). Public-services setup adds SSL.
- **PROVEN-a — Payloads accepted (dataset-driven).** Each dataset = WCMP2 record (via webapp Dataset Editor or MCF YAML in `metadata/discovery`) + data-mappings plugins. Metadata notification → `origin/a/wis2/CENTRE_ID/metadata` → GDC downloads/caches the record; data ingest → MinIO `wis2box-incoming` → `wis2box-management` processes per mapping → public storage → WNM with canonical link. Built-in plugins: `csv2bufr.ObservationDataCSV2BUFR` (CSV→BUFR, needs template), `bufr2geojson.ObservationDataBUFR2GeoJSON`, `synop2bufr` (SYNOP→BUFR, filename must contain YYYYMM groups), `bufr4.ObservationDataBUFR` (splits bulletins/subsets, WIGOS/location QA), `geojson.ObservationDataGeoJSON` (API/GeoJSON publish), `cap_message.CAPMessageData` (CAP XML, see below), `universal.UniversalData` (passthrough, `properties.datatime` = regex group 1, geometry null). Templates prefill mappings (e.g. `weather/surface-based-observations/synop|temp`, `weather/advisories-warnings`); `other` allows custom Discipline + Sub-discipline topics and WMO Data Policy (core caches by default; recommended needs license link); `Publish metadata without WIS2 data notifications` allowed (then no TH defined, download link required).
- **PROVEN-a — CAP support (yes, first-class).** CAP Alerts doc: dataset with **`weather/advisories-warnings`** template → `CAPMessageData` plugin validates XML against **CAP v1.2 schema** via `capvalidator`, optional signature verification (`CHECK_CAP_SIGNATURE=True`); authoring via **WMO CAP Composer** (separate web service) → MQTT forward (host/port/user/pass/dataset ID) *or* direct upload of CAP XML to `wis2box-incoming` (any producer). After publish: visible in webapp Monitoring → Notifications/Published Data → View Alert. Plugin config example: `xml: - plugin: wis2box.data.cap_message.CAPMessageData / notify: true / buckets: [${WIS2BOX_STORAGE_INCOMING}] / file-pattern: '^.*\.xml$'`.
- **PROVEN-a — Guide endorsement (with restraint).** Guide §2.6.3: *"Members may use whichever software components they consider most appropriate … wis2box is available [and] implements the requirements for a WIS2 Node"* with the function list above. It is a reference, **not** a mandate to run it (no "Node SHALL run wis2box" in §2.6.3).

**Open question that remains (Q6)**

- Which parts of `wis2box` would a future Node operator run (full stack vs only broker/storage/API subset for link-quality JSON)? Primaries describe the full stack; there is no primary guidance on a minimal "probe-only" subset. That decision belongs to the Node operator + GISC, not to this research file.
- Operational data volume / cost for the DGM link (beyond EC2 examples) remains site-specific and unverified for this project.

**Recommended answer**

- Close Q6 at the research level as PROVEN-a: `wis2box` **is** the WMO FOSS reference Node (v1.3.0 fetched), it **does** accept file ingest + MQTT CAP-Composer forward, it **does** support CAP via `CAPMessageData`/`capvalidator`, and it **is** a full Docker Compose stack (MinIO + double Mosquitto + ES/pygeoapi/Vue) requiring public 80/1883 exposure + DMZ posture + 2 vCPU/4 GB/24 GB minimum.
- Builder implication (present as choice, do not pre-decide): (A) *Probe → own broker (lab)* — probe publishes `cgo/v1/...` (or later `origin/a/...` if Q4 closes) directly to the lab Mosquitto; wis2box not involved; (B) *Probe → wis2box* — probe drops sealed-report JSON/CAP XML into a wis2box dataset's incoming path (or MQTT-forwards CAP via the Composer broker config) and lets wis2box emit the WNM + serve HTTP + publish metadata. Option (B) is the only WIS2-conformant production path (it gets WTH/WNM/WCMP2/GISC validation for free); option (A) stays lab-honest and never WIS2.
- No conflict between primaries. Flag as **not a recommendation to deploy wis2box** — the stack's public-Internet requirement + internal-network warning make it a Node-operator decision, not a probe dependency.

---

## Q7 — Citable MQTT client library (primary source: docs / releases)

**Requirement**: choose a Go MQTT client with citable primary source. No library in `go.mod` today. Status was *non commencé*.

**Primary sources fetched**

| Item | URL | Retrieval | Pin |
|---|---|---|---|
| Eclipse Paho Go MQTT 5 client (`paho.golang`) — README/docs via pkg.go.dev | `https://pkg.go.dev/github.com/eclipse/paho.golang` | 2026-11-09 | **v0.23.0, published 06 Sep 2025**, EPL-2.0, valid go.mod, tagged, not-stable badge (pre-1.0) |
| Eclipse Paho Go MQTT 3 client (`paho.mqtt.golang`) — README via pkg.go.dev | `https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang` | 2026-11-09 | **v1.5.1, 16 Sep 2025** (versions index `?tab=versions`) |
| Eclipse Paho Go MQTT 3 repo README (GitHub) | `https://github.com/eclipse-paho/paho.mqtt.golang` | 2026-11-09 | 3.1k stars, 574 forks, EPL-2.0, 582 commits |
| Eclipse Paho project site — Go Client overview | `https://eclipse.dev/paho/clients/golang/` | 2026-11-09 | Lists MQTT 3.1/3.1.1/5.0, SSL/TLS, persistence, TCP/WebSocket, Non-Blocking API, High Availability |

**Key excerpts / cited claims**

- **PROVEN-a — Two libraries, two generations.** `paho.mqtt.golang` README: *"This repository contains the source code for the Eclipse Paho MQTT 3.1/3.11 Go client library … This library supports a fully asynchronous mode … A client supporting MQTT V5 is also available [paho.golang]."* `paho.golang` README: *"This repository contains the source code for the Eclipse Paho MQTT V5 Go client library … [v3 client] is an older project, and its API is very different to this one."* V5 README adds: release v0.20 introduced **full QoS 1/2 support** (session persistence gaps in ≤0.12 closed), `autopaho` recommended as the simple starting point (auto-reconnect wrapper over `paho`), `paho` is the low-level library, `autopaho/examples/docker` runs publisher+subscriber against Mosquitto.
- **PROVEN-a — API surface relevant to WIS2 (QoS 1, TLS, WS, persistence).** v3 (1.5.1): `ClientOptions` (broker URI `tcp://|tls://|ssl://|tcps://|ws://|wss://`), `NewClient`, `Publish(topic, qos, retained, payload)` + `Wait/Error`, `Subscribe`, `SetAutoReconnect(true)` default, `SetKeepAlive`, `SetOrderMatters(false)` recommended unless ordered delivery + `max_inflight_messages=1`, `DefaultPublishHandler`/`AddRoute` required to avoid inflight-limit deadlock when `CleanSession=false` + retained messages, file/memory stores, logging endpoints (`ERROR/CRITICAL/WARN/DEBUG`). Dependencies: `github.com/gorilla/websocket`, `golang.org/x/net/proxy` (both via `go get`). Good Mosquitto defaults notes: `listener` loopback-only in v2+, `persistence` defaults false, `max_keepalive` 65535.
- **PROVEN-a — V5 API + caveats.** v0.23.0 README: packages `github.com/eclipse/paho.golang/paho` + `github.com/eclipse/paho.golang/autopaho`; install `go get github.com/eclipse/paho.golang` + `go mod tidy`. Known issues documented: topic aliases are **not** part of session state (queued aliased messages may misfire after reconnect); multi-server/`CleanStart=false` requires shared session state (clustered servers) or messages may be lost; `SessionExpiryInterval` effectively ignored (server `Session Present` flag carries expiry); `Receive Maximum` not enforced on inbound (honoured on outbound). v0.20 breaking changes: `ClientConfig` private, `Pinger` interface changed, `ClientConfig.Router` → `OnPublishReceived`, `autopaho` `CleanSession` default now `false` (with `SessionExpiryInterval` default 0, so usually no impact).
- **PROVEN-a — WIS2 protocol fit.** Guide §2.5.1 requires **both** MQTT v3.1.1 **and** v5.0 for publication/subscription; v5 preferred for Global Brokers (shared subscriptions) — code in §2.5.1. Both Paho lines therefore satisfy the protocol floor; only the V5 line gives shared subscriptions natively. QoS 1 + retain=false + username/password + TLS (8883) + WS/443 are all expressible in both libraries.
- **PROVEN-a — Licence + maintenance.** Both EPL-2.0 (redistributable); pkg.go.dev badges: `paho.golang` v0.23.0 *"Stable version [unchecked]"* (pre-1.0), `paho.mqtt.golang` v1.5.1 latest with 2025 fix for oversized UTF-8 encoding noted in downstream release notes. Both are actively published in 2025 (not abandoned). Bug reporting requires pinned version in `go.mod` + full `ClientOption` config + server details — citable reproducibility path.

**Open question that remains (Q7)**

- None at the citability level. The remaining choice is human (stop-and-ask for a new Go module dep): V5-preferred (`paho.golang` v0.23.0) vs battle-tested V3 (`paho.mqtt.golang` v1.5.1) vs both (V3 for Node, V5 for Global Broker). The research file presents the trade-off; it does not add the dep.

**Recommended answer**

- Recommend **`github.com/eclipse/paho.golang` v0.23.0** (V5, `autopaho` for probe publishing + `paho` for advanced subscribe) as the primary citable candidate, because WIS2 §2.5.1 explicitly prefers V5 for Global Broker flows and the V5 line is the maintained future line (full QoS 1/2 since v0.20, active Sep 2025 release). Keep **`github.com/eclipse/paho.mqtt.golang` v1.5.1** as the citable fallback if the lab Mosquitto must stay on 3.1.1-only — both are PROVEN-a compatible with WIS2 QoS 1/retain=false/TLS.
- CONFLICT to flag: GitHub org paths differ from `pkg.go.dev` import paths (`github.com/eclipse-paho/paho.mqtt.golang` vs import `github.com/eclipse/paho.mqtt.golang`; `github.com/eclipse-paho/paho.golang` vs import `github.com/eclipse/paho.golang`). Pin the **import path** (`github.com/eclipse/...`), not the GitHub org, in `go.mod` + README. Both EPL-2.0, so licence path is clean.
- CONFLICT to note: any prior "no MQTT library" baseline (R9/R10) is still true of the repo checkout; Q7 selection does not yet change `go.mod`. Adding either line needs the stop-and-ask approval already named in design §7.6.
- Blocker: needs human dep approval + broker choice (own vs wis2box) — not a fetch.

---

## Cross-gate synthesis for the builder

| Q | Research verdict | Slices it gates | Human artefact still needed |
|---|---|---|---|
| Q1 | **PROVEN-a** (WTH v1.3.0 + WNM v1.3.0 + WCMP2 v2.3.0 pinned) | S0, S2, S3, S4 | Version pin decision only |
| Q2 | **PROVEN-a** (Guide §2.6.1.1–§2.6.1.3 + Overview Notes) | S0, S4 | NFP/PR + GISC identity (Q4) |
| Q3 | **PROVEN-a** (granule + QoS 1/retain=false + link-primary/inline-optional) | S0, S3, S4 | Data-policy (core vs recommended) choice |
| Q4 | **unverified** (negative Codes Registry finding only) | S4 blocked | Signed centre-id record **or** written non-registration decision |
| Q5 | **unverified** (domains PROVEN-a, mapping absent by design) | S3 blocked | Operator-signed band table + scope/sender/category |
| Q6 | **PROVEN-a** (wis2box v1.3.0 = reference Node, CAP first-class) | S0 decision, S4 alt path | Own-broker vs wis2box + public-Internet stance |
| Q7 | **PROVEN-a** (two citable candidates, V5 recommended) | S1/S2 dep | Stop-and-ask dep approval + version pin |

**What this does NOT do:** propose a broker config, pick a Mosquitto listener, add a `go.mod` entry, or invent Q4/Q5. The lab `cgo/v1/...` namespace and the `origin/a/...` WIS2 namespace remain strictly separated until Q4 closes (PROVEN-a: GISC assessment checks topics + notifications + data server + broker). Any slice built before Q4/Q5 signatures must assert (test) that its config contains zero `origin/a/wis2` topics and zero CAP `Alert` emissions without a signed band table.

*End of research — builder picks presentation order, human signs Q4/Q5.*
