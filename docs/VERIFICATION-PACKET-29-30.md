# Verifier packet — rows 83/85 (slice 29/30 builder passes)

The verifier re-runs **these exact commands** at `main` HEAD and owns
both rows. A deviation on any line = refusal (per house rule). Expected
outputs are quoted from the builder's 2026-11-09 evidence; the verifier's
own run replaces them in the row entry, not here.

**Prerequisites:** Go 1.25+, Node 22, `gh` authenticated, `cosign`
binary (release zip from `sigstore/cosign`, or `go install` — big), Internet
(`proxy.golang.org` + `releases.ubuntu.com` reachable), and a working
`web/frontend/dist` (build before copying commands verbatim is **not**
needed — see commands).

## Packet A — slice 29 (row 83: frontend + build + root Go)

```bash
cd web/frontend
npm test                  # expected: 68 files, 321 tests passed
npm run build             # expected: exit 0, TOTAL gz : 357.6 KB (budget 600 KB)
git checkout -- dist/index.html   # restore the tracked stub; git status --short clean
cd ../.. 
 go test ./... -count=1  # expected: exit 0, all packages ok
 python tools/verify.py  # expected: 6 passed, 0 failed, 0 not run
```

## Packet B — slice 30 (row 85: root Go + SBOM/digest)

```bash
cd /tmp/cgo-digest && rm -rf * 2>/dev/null; mkdir -p /tmp/cgo-digest && cd /tmp/cgo-digest
gh release download v0.1.0 --repo Realms4239/cgo   # 13 assets
sha256sum --ignore-missing -c checksums.txt        # expected: 6 OK (only 6 files are hashed)
cosign verify-blob --certificate-identity "https://github.com/Realms4239/cgo/.github/workflows/release.yml@refs/tags/v0.1.0" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --bundle checksums.txt.sigstore.json checksums.txt   # expected: Verified OK (keyless OIDC sign-blob)
```

GoReleaser CI-run evidence (reuse, **re-run only if the tag moves**): Actions
run `42493531454` on tag `v0.1.0` @ `500cc18` — release, 5 per-module SBOMs,
publish artifacts, `cosign attest digest`, all 13 steps green in 9m2s.
Record the run id in the row when present; if the tag was force-retagged for
a newer release, point to the new run.

## Signature

`docs/VERIFY.md`: promote each row only if **every** command in its packet
passes with the expected output (no substitution, no -skip flags).
Append a `docs/CHECKPOINTS.md` entry *before* the row promotion
(house order) quoting command, date, host, HEAD, full output for any
mismatch. Rows 83 and 85 may be signed in one pass (same command set).
