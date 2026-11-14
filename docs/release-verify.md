# Vérification du release sur hôte vierge

Les commandes ci-dessous vérifient une archive publiée pour un tag `vX.Y.Z`,
sur une machine fraîche (sans le dépôt, sans cache). Remplacez `vX.Y.Z` par
le tag réel.

```
gh release download vX.Y.Z -R Realms4239/cgo -p 'checksums.txt*'
sha256sum -c checksums.txt
cosign verify-blob --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'github\.com/Realms4239/cgo' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com checksums.txt
gh attestation verify checksums.txt -R Realms4239/cgo
```

Ce que cela prouve. `sha256sum -c` confirme que les archives téléchargées
correspondent aux sommes publiées. Le `cosign verify-blob` confirme que ces
sommes ont été signées par le workflow GitHub Actions du dépôt (keyless OIDC)
au moment du tag. `gh attestation verify` confirme la provenance SLSA : les
artefacts attestés proviennent bien du build du tag, par ce dépôt. Un échec à
n'importe laquelle de ces étapes signifie que l'archive n'est PAS la release
signée — ne l'installez pas, et signalez-le au mainteneur. Un succès complet
signifie l'intégrité des octets, l'identité du signataire et la provenance
du build. Le tag `v*` est créé par l'humain, jamais par CI : la CI construit,
signe et atteste ce que le tag pointe, elle ne décide jamais de le créer.
