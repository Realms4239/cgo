 # API bootstrap et clarté du cockpit — 2026-11-17

Ces sections vivaient dans `docs/CHECKPOINTS.md` jusqu'à ce que ce dernier atteigne son plafond de 384 Ko (contrôle `hygiene`). Le texte est conservé ici intégralement ; le ledger ne garde qu'un pointeur. Aucune preuve n'a été réécrite ni supprimée.

---

## 2026-11-17T11:36Z — API bootstrap and cockpit clarity (web-perf + impeccable)

- MEASURED BEFORE (real cgo server, localhost, 4173): `/api/integrity`
  0.43–1.14 s every call; `/api/results` 3.29 s cold (0.08 s warm); duplicate
  `/api/audit/list` twice on first render; bundle 354.6 KB gz; LCP 736 ms,
  CLS 0.018, DCL 306 ms (single sample, noisy).
- ROOT CAUSE: integrity re-hashed the whole frozen corpus on EVERY call, twice
  (`ProvenanceSHA` + `ProvenanceHash8` → second full pass); results was memoized
  but paid the cold scan on the first operator render.
- FIX: `api/corpus.go` memoizes the integrity payload by the cheap corpus
  fingerprint, derives `hash8` from the single SHA, exposes ETag +
  `Cache-Control: no-cache`, and `New()` warms both caches before `/api/health`.
  `NewServer` (tests) stays cold by design.
- MEASURED AFTER: `/api/integrity` **0.085–0.088 s**, `/api/results` **0.089–0.099 s**
  on the first call; ETag `0925872515d10e73` present; `/api/audit/list` now **1**
  request instead of 2 (shared single-flight `src/lib/auditList.ts`).
- Copy clarity (impeccable/clarify): `Alertes à l'heure` → `Respect de l'échéance`
  (4 labels + cockpit), `Delta sous charge` → `Allongement sous charge`, and
  "imputable à la technologie d'accès" → "dû à l'attente dans les files" —
  plain French, domain terms (small p95, JFI, hash8, AQM/CC) preserved.
- VERIFIED: 345 vitest tests (+2 single-flight) green; **25/25 Playwright e2e**
  with installed Edge via `PWCHROMIUM` (bundled Chromium absent — environment,
  not a code failure); detector `[]`; bundle 354.7 KB gz < 400 budget;
  `go test` API module + gates 6/6. Browser-level LCP was noisy (736→~1050 ms
  median) and is NOT claimed as improved — local plaintext HTTP/1.1 queues ~20
  parallel fetches; production HTTPS/HTTP2 multiplexes. Font subsetting (352 KB
  Inter) needs a new tool — out of scope, named residual.


## 2026-11-17T07:57Z — campaign complete, consolidation sealed

- All planned matrices finished: 14 runs / 146 cells (Layer 1 72 cells; Layers
  3+4 74 cells), consolidated in `docs/verification/2026-11-16-exhaustive/CONSOLIDATION.md`.
  The formerly BLOCKED `both` direction was re-run clean as `run-1794900710`
  (7 rows, 4 valid) after the guard regex was fixed — no concurrent writers.
- P2 goodput divergence resolved by `l4p2rep5` (10 cells): cubic 2.9–6.0 Mbps
  (median 4.2) brackets the sealed 2.4–4.9; BBR 18.0–19.4 in 4/5 reps with a
  single 1.3 collapse. Stochastic loss variance; no measurement defect.
- Comparator regression after the 1% rate-only tolerance (`0788835`): the local
  replay of the ORIGINAL sealed fixture export still returns 23/23 accepted —
  the fix did not weaken the earlier proof. Gates 6/6.
- Residual scope: push/freeze/tag/merge, verifier-owned VERIFY rows, field-trace
  seal and `.dist` calibration, DSI countersign, T40/T46/T48 BLOCKED with named
  conditions. No claim promoted by the builder in this campaign.


## 2026-11-17T07:26Z — REAL-audit CHR proof closes the fixture-only gap

- Fresh 45 s audit in a single-writer gap with the FIXED binary and interface
  flags (`/home/bench/eb-realrem.sh`) produced `audit-1794898684`: eBPF telemetry
  (`source=ebpf`, object `8612f578…`, 421 368 retrans bytes, flows 5/5) and
  `remediation_state=available` with a REAL RouterOS LAN script derived from the
  measured 17.52 Mbit bottleneck / 200 ms RTT (ether1/ether2).
- Live CHR 7.24.6 import under named window
  `APPROVED-2026-11-17-chr-real-audit.md` (pinned image re-deposited) →
  `import.accepted=true`, **11/11 checks**, export SHA `91d67458…`. Evidence in
  `docs/verification/2026-11-17-chr-real-audit/` (`43da827`).
- First attempt 10/11 failed only on `cake-bandwidth`: RouterOS prints
  `16.6Mbps` for expected `16648496` (0.29% one-decimal display rounding).
  Fixed with rate-only 1% tolerance + regression test (`0788835`); rebuilt
  `cgo-chr-v13` and re-ran from a fresh pinned image → 11/11. No tolerance for
  non-rate strings or genuinely different rates (19M vs 20M still fails).
- Residual, named: 11 checks (LAN/ingress only) because the real audit was
  down-direction; the earlier WAN claim still rests on the fixture seal. Local
  proof/export are transcriptions (safe text channel is not byte-exact) — the
  authoritative hashes are the proof's `export_sha256` and host `HOST-SHA256.txt`.
  No VERIFIED row promoted by the builder.

## 2026-11-17T01:05Z — CORRECTION: "disk full / two-strike" was a misdiagnosis (déplacé du ledger)

- I called a two-strike halt on `l3both` and blamed the guest disk (97%, 79 MB free).
  That was WRONG. Disk and memory are healthy (7.4 GB available, no OOM in dmesg,
  `/home/bench` only 56 MB). The real root cause: the generic launcher's guard was
  `pgrep -f 'cgo run --bench dualns'`, which does NOT match `cgo-fixed run --bench
  dualns`. So successive launches (l3both run-1794876850, run-1794877014 and l3f1)
  overlapped — up to THREE writers — and fought over the shaper, producing exactly
  the rtt 0.0 / small 0.0 signature seen before. This is the same class of defect
  as the very first contaminated campaign, now in the launcher itself.
- Fix: guard changed to `pgrep -f 'run --bench dualns'` (matches both binaries),
  verified by the launcher correctly REFUSING a launch with exit 3 while a survivor
  was alive. All surviving PIDs (135120/133161/133163/165942 etc.) stopped by exact
  PID; contaminated runs left with their partial data and NOT used as evidence.
- Also reverted the launcher's in-script `run-dualns.sh up` step (it coincided with
  the failures); topology is restored in a separate verified step instead.
- Clean relaunch l3f1 `run-1794877413` (P2/cake/{cubic,bbr}/down, flows 1) shows
  real RTT 116–123 ms in its log — contamination gone. Lesson: verify the guard
  pattern against the ACTUAL binary name before trusting single-writer claims.
