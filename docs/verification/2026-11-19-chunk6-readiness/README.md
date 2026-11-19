# Chunk 6 readiness record

Chunk 6.1 step 1 (environment readiness) was run on the Linux bench on
2026-11-19 with the current HEAD binary `/home/bench/cgo-20261119`, reached
through the approved bench SSH path.

```
$ sudo -n /home/bench/cgo-20261119 doctor
mode: full
  [ok] os — linux/amd64 → mode full
  [ok] tc — /usr/sbin/tc
  [ok] cap_net_admin — root
  [ok] bbr — disponible (BBRv1, mainline)
  [warn] bbr3 — absent — optionnel : kit/install-bbr3.sh (noyau BBRv3, 7.1)
  [ok] bench-offloads — 6 interfaces de transit relues off (GRO/GSO/TSO)
  [ok] ping — présent
```

The `bbr3` warning is expected on the stock 6.8 kernel and does not reduce the
readiness result: the chunk 6 definition of done explicitly allows CUBIC plus
BBRv1 on stock 6.8 with BBRv3 runtime detection ready.

## Chunk 6 status from this checkout

| Step | Status | Evidence |
|---|---|---|
| 6.1.1 environment readiness | done | output above |
| 6.1.2 race and real-tag suites | not run in this pass | host gates run six-of-six on every commit; the Linux `-tags=real` lane remains a named gap |
| 6.1.3 eBPF-vs-tc confrontation JSON | gap in this tree | no `--confront-out` flag exists in `cmd/cgo`; the kernel truth itself is sealed at `docs/VERIFY.md` row 43 (226 eBPF retrans events matched 226 `tc -s` cake drops) |
| 6.1.4 clean dualns matrix | done earlier | `docs/verification/2026-11-16-v2-corpus/run-1794853778`, 24 cells all valid, plus the 146-cell exhaustive campaign |
| 6.2 field audit and `.dist` | done with caveats | `../2026-11-19-field-tune/`, `../2026-11-19-field-trace-p3/`, `../2026-11-19-multiprofile-trace/` |
| 6.3 RouterOS CHR | done | `../2026-11-19-chr-verifier-replay/`, VERIFY row 71 promoted by an independent verifier |
| 6.4 corpus, screenshots, thesis | human-gated | `book/` and `C:\cgo\memoire.docx` stay human-owned; no thesis build in this pass |
| 6.5 v2.0.0 tag and hard freeze | human-gated | no tag, no freeze, no force-push |

## Chunk 6.2 integrity surface

The current HEAD served locally at `127.0.0.1:19092` with `--dev-no-auth`
returned HTTP 200 on `GET /api/integrity`: `integrity: verified`, `ok: true`,
159 manifests, 208 valid rows, 74 degraded, 246 quarantined, `hash8: 4ac3e057`.
These counts match the frozen-corpus row in `docs/VERIFY.md`. The check ran on
loopback only; no production host was contacted. The server was stopped after
the read.

## Interface verification, 2026-11-19

Fresh evidence, all run in this pass after the current HEAD:

| Check | Command | Result |
|---|---|---|
| Six gates | `sh .githooks/run-gates.sh` | `6 passed, 0 failed, 0 not run` |
| Frontend unit tests | `npm test` (web/frontend) | 78 files, 360 tests passed, exit 0 |
| Frontend build | `npm run build` | exit 0, `TOTAL gz : 355.3 KB`, echarts 198.8 KB / 350 KB max |
| Embed artifact | `git diff web/frontend/dist/index.html` | stale title and hash corrected, committed separately |

The first gate run failed while `npm run build` was still rewriting `dist/`;
the re-run after the build completed passed 6/6. That failure was a concurrent
build race, not a code defect, and no gate or permit was changed.

The rendered cockpit was inspected at 1440x900 through a loopback server at
`127.0.0.1:19093` with `--dev-no-auth`. The first screen carries exactly two
navigation modes, Cockpit and Labo, one primary action (launch the 60 s audit),
one secondary action (view an archived report), and an honest empty state that
says no value is invented before measurement. The 10/40/10 audit sequence is
stated once in a compact side panel. No chart, badge or number is shown before
it is measured.

The muted palette is deliberate and measured, not a contrast defect:
`--text-faint #8a8a92` is documented in `web/frontend/src/styles/tokens.css`
at 5.37:1 on cards and 5.75:1 on canvas, above the WCAG AA threshold. No
styling change was made to "fix" a passing design. The only UI-related change
was syncing the tracked `dist/index.html` with the current source title and
build hash.

This record promotes no VERIFY row. Step 6.1.3 cannot be claimed: producing
`ebpf_vs_tc_report.json` through `--confront-out` would require new CLI code
and a verifier pass, which is a separate slice rather than evidence assembly.
