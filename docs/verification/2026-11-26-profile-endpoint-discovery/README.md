# Endpoint discovery for P1/P2/P3 field capture — 2026-11-26

Read-only discovery through `C:\WEBSERVER\safe_read.py`; no host writes, no capture, no custom DNS guessing. Ratified bands from `cmd/tracemodel/main.go:17-21`: P1 `[20,40) ∪ [60,100)`, P2 `[100,200)`, P3 `>=600`.

ICMP RTT is a *proxy* for the TCP `srtt_us` the tracer records. An endpoint is a candidate only; every candidate must still be proven by a real TCP capture before any calibration.

## Observed endpoints

| Band | Endpoint | ICMP min/avg/max (ms) | Evidence | Status |
|---|---|---|---|---|
| P1 [20,40) | `apple.com` | 35.3 / 36.0 / 36.5 | `20261126-075538` sha `22377be5…` | candidate |
| P1 [20,40) | `ibm.com` | 36.1 / 36.3 / 36.5 | `20261126-080214` sha `e7b30fe0…` | candidate |
| P1 [20,40) | `amazonaws.com` | 37.2 / 37.9 / 38.3 | `20261126-075538` sha `22377be5…` | candidate |
| P1 [20,40) | `spotify.com` | 39.2 / 39.7 / 40.3 | `20261126-081806` sha `4d07acde…` | near upper edge |
| P1 [60,100) | `microsoft.com` | 63.2 / 63.6 / 63.8 | `20261126-075538` sha `22377be5…` | candidate |
| P4 (already calibrated) | `1.1.1.1`, `google.com`, `mozilla.org`, `nasa.gov` | 41-47 | `20261126-075028` sha `88c5c474…`, `20261126-075215` sha `dd280530…` | not in this scope |
| local | `192.168.10.1`, `41.207.49.27` | 0.25-0.3 | `20261126-075028` sha `88c5c474…` | below all bands |
| P2 [100,200) | `dropbox.com` | 206.4 / 207.1 / 208.0 | `20261126-080214` sha `e7b30fe0…` | above band |
| P2 [100,200) | `wikipedia.org` | 211.0 / 211.3 / 211.9 | `20261126-080214` sha `e7b30fe0…` | above band |
| P2 [100,200) | `github.com` | 229.4 / 229.7 / 230.1 | `20261126-075215` sha `dd280530…` | above band |
| P3 >=600 | `salesforce.com` | 546.7 / 558.7 / 566.8 | `20261126-080214` sha `e7b30fe0…` | below band |
| P3 >=600 | `afrinic.net` | 378.5 / 379.8 / 380.8 | `20261126-075215` sha `dd280530…` | below band |
| gap / no route | `oracle.com`, `redhat.com`, `slack.com`, `akamaized.net`, `githubusercontent.com` | 100 % packet loss | `20261126-075538` / `080214` / `081806` | unusable |
| near-local | `zoom.us`, `digitalocean.com`, `yas.mg`, `example.com`, `cloudflare.com` | 1-3 | `20261126-075215` / `081806` | below all bands |

## Disposition

- P1 has four plausible TCP targets (three in the 20-40 ms sub-band, one in 60-100 ms). They still need a scoped capture to prove real `srtt_us` samples and reach 30 per band.
- P2 and P3 are **BLOCKED**: every probed candidate falls outside its band (P2 above 200 ms, P3 below 600 ms). No endpoint was invented and no band was widened. Named unblock: a reachable TCP service whose measured SRTT actually lands in the band, or a human ruling on the profile definition.
- No `.dist`, no frozen corpus, no gate and no host state was changed by this discovery.
