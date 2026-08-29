# meteolink (npm)

Binary distribution of [Meteolink](https://github.com/Realms4239/cgo) via npm —
esbuild/rollup-style platform wrapper: the `postinstall` hook downloads the
prebuilt `cgo` binary for your platform from the GitHub release (no
compilation, no dependencies).

## Installation / Installation

```sh
npm i -g meteolink
```

## Usage / Utilisation

```sh
meteolink --serve
```

- **English:** start the dashboard + API server. On **Windows** the tool runs
  in observation mode (no traffic control); full AQM/BBR shaping requires
  **Linux** with `CAP_NET_ADMIN` (see `kit/install.sh` for dependencies).
- **Français :** lance le tableau de bord et le serveur API. Sous **Windows**,
  l'outil fonctionne en mode observation ; le façonnage AQM/BBR complet
  nécessite **Linux** avec `CAP_NET_ADMIN` (voir `kit/install.sh`).

## Supported platforms / Plateformes prises en charge

| OS      | Arch   |
|---------|--------|
| darwin  | amd64, arm64 |
| linux   | amd64, arm64 |
| windows | amd64  |

If the binary is missing, run `npm rebuild meteolink` to re-trigger the
download.

## License / Licence

MIT — see [LICENSE](https://github.com/Realms4239/cgo/blob/main/LICENSE).
