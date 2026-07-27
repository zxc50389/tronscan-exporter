# tronscan-exporter

A small [Prometheus](https://prometheus.io/) exporter, written in Go, that polls
the public [TronScan](https://tronscan.org/) API and exposes TRON network
throughput metrics for scraping.

## Metrics

The exporter fetches `https://apilist.tronscanapi.com/api/system/tps` every 10
seconds and publishes the following gauges on `/metrics`:

| Metric              | Description                                            |
| ------------------- | ------------------------------------------------------ |
| `tron_current_tps`  | Current transactions per second on the TRON network   |
| `tron_max_tps`      | Maximum transactions per second on the TRON network   |
| `tron_blockHeight`  | Current TRON block height                              |

The HTTP server listens on `:8091`:

- `/`        – simple landing page with a link to the metrics
- `/metrics` – Prometheus metrics endpoint

## Running locally

```sh
cd src
go mod tidy
go run .
# then browse to http://localhost:8091/metrics
```

Or build a static binary:

```sh
cd src
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tronscan-exporter
./tronscan-exporter
```

## Repository layout

```
.
├── .gitlab-ci.yml        # GitLab CI: build the binary and package/deploy a .deb
├── ci/                   # Debian packaging scripts
│   ├── deb.sh            # Builds the .deb (control file + systemd unit)
│   ├── files.sh          # Stages runtime assets before build
│   └── deb/DEBIAN/       # maintainer scripts (preinst/postinst/prerm) + control
├── src/                  # Go source (module first_exporter)
│   ├── main.go
│   ├── go.mod
│   └── go.sum
└── tron_exporter/        # Earlier copy of the same source (kept for reference)
```

## CI / packaging

`.gitlab-ci.yml` defines two stages:

1. **build** – compiles a static `linux/amd64` binary with the `golang:1.20-alpine`
   image and publishes it as an artifact.
2. **deploy-prd** – on tags ending in `-prd`, wraps the binary in a Debian package
   (via `ci/deb.sh`, which also generates a systemd service unit) and copies it to
   the production Debian repo host over SSH.

The generated systemd unit runs the service from `/becreator/tronscan-exporter`.

## Notes

- `src/` and `tron_exporter/` contain the same program; `tron_exporter/` is an
  older copy retained from the original archive. The compiled binary that shipped
  in that folder is a build artifact and is intentionally not committed (see
  `.gitignore`).
- `ci/files.sh` references runtime assets (`config/`, `html/`, `lang/`,
  `GeoLite2-City.mmdb`) that are not part of this exporter's source tree; they are
  leftovers from the template this project was derived from and are not required to
  build or run the exporter.
