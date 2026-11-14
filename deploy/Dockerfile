# Build multi-étape : le clone git ne contient pas bin/cgo (gitignored) —
# l'image se construit de la source, jamais depuis un artefact local.
FROM golang:1.25-bookworm AS build
WORKDIR /src
ARG VERSION=dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /cgo ./cmd/cgo

FROM scratch
ARG VERSION=dev
ARG REVISION=dev
ENV CGO_DATA_DIR=/data
COPY --from=build --chown=1000:1000 /cgo /cgo
COPY --chown=1000:1000 deploy/data /data
WORKDIR /data
VOLUME /data
LABEL org.opencontainers.image.title="cgo" \
      org.opencontainers.image.description="Binaire cgo (dashboard + API + audit) — le façonnage exige CAP_NET_ADMIN, l'observation ne demande rien" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.source="https://github.com/Realms4239/cgo" \
      org.opencontainers.image.licenses="MIT"
EXPOSE 9090

# Non-root: /data appartient à 1000:1000 et le volume nommé préserve
# l'uid. scratch ne peut binder les ports < 1024 (non-root sans NET_BIND_SERVICE) ;
# 9090 est hors privilège. Pour le façonnage (CAP_NET_ADMIN) : --cap-add au
# run, jamais dans l'image. scratch n'emporte pas de racines CA : cgo sert son
# propre TLS ; la confiance sortante (kit tls) épingle par SPKI.
USER 1000:1000
ENTRYPOINT ["/cgo", "--serve"]
