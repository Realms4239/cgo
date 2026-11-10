FROM scratch
ARG VERSION=dev
ARG REVISION=dev
COPY bin/cgo /cgo
# Étiquettes OCI de provenance (SLSA L1 : quoi a construit quoi, depuis
# quel commit) — le digest de l'image publié doit correspondre à ces
# valeurs. Jamais de tag mutable :latest ; l'image se tire par digest
# (voir deploy/docker-compose.yml + docs/deploy.md § Chaîne
# d'approvisionnement).
# Le corpus gelé vit ici : CGO_DATA_DIR épingle la racine pour TOUT le flux
# (runs, figures, link_audit.csv, artefacts d'audit, profils, calibrations).
ENV CGO_DATA_DIR=/data
WORKDIR /data
VOLUME /data
LABEL org.opencontainers.image.title="cgo" \
      org.opencontainers.image.description="Binaire cgo (dashboard + API + audit) — le façonnage exige CAP_NET_ADMIN, l'observation ne demande rien" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.source="https://github.com/Realms4239/cgo" \
      org.opencontainers.image.licenses="MIT"
EXPOSE 9090
ENTRYPOINT ["/cgo", "--serve"]
