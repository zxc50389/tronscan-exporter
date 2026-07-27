# tronscan-exporter

A small [Prometheus](https://prometheus.io/) exporter, written in Go, that polls
the public [TronScan](https://tronscan.org/) API and exposes TRON network
throughput metrics for scraping.

## Metrics

By default the exporter polls `https://apilist.tronscanapi.com/api/system/tps`
every 10 seconds and publishes the following gauges on `/metrics`:

| Metric               | Description                                                        |
| -------------------- | ----------------------------------------------------------------- |
| `tron_current_tps`   | Current transactions per second on the TRON network               |
| `tron_max_tps`       | Maximum transactions per second on the TRON network               |
| `tron_block_height`  | Current TRON block height                                         |
| `tron_up`            | `1` if the last scrape of the TronScan API succeeded, else `0`    |

Use `tron_up` to detect stale data: when a scrape fails the other gauges keep
their previous values, so alert on `tron_up == 0` rather than on a metric
suddenly reading zero.

## Configuration

All settings are optional flags:

| Flag                   | Default                                             | Description                              |
| ---------------------- | --------------------------------------------------- | ---------------------------------------- |
| `-web.listen-address`  | `:8091`                                             | Address to expose metrics/landing page   |
| `-web.telemetry-path`  | `/metrics`                                          | Path under which metrics are exposed     |
| `-tron.tps-endpoint`   | `https://apilist.tronscanapi.com/api/system/tps`    | TronScan TPS API endpoint to poll        |
| `-scrape.interval`     | `10s`                                               | How often to poll the API                |
| `-scrape.timeout`      | `10s`                                               | Timeout for each API request             |

The service also serves a simple landing page at `/` linking to the metrics.

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
./tronscan-exporter -scrape.interval=15s
```

Each [GitHub Release](https://github.com/zxc50389/tronscan-exporter/releases)
attaches a prebuilt `linux/amd64` binary and a Debian package. To install via
the `.deb` (registers and starts a `systemd` service):

```sh
sudo dpkg -i tronscan-exporter_<version>_amd64.deb
systemctl status tronscan-exporter
```

## Repository layout

```
.
├── .github/workflows/    # GitHub Actions: CI (build/test) and Release (publish binary)
├── .gitlab-ci.yml        # GitLab CI: build the binary and package/deploy a .deb
├── ci/                   # Debian packaging scripts
│   ├── deb.sh            # Builds the .deb (control file + systemd unit)
│   ├── files.sh          # Records build metadata (version)
│   └── deb/DEBIAN/       # maintainer scripts (preinst/postinst/prerm) + control
└── src/                  # Go source (module github.com/zxc50389/tronscan-exporter)
    ├── main.go
    ├── main_test.go
    ├── go.mod
    └── go.sum
```

## License

Released under the [MIT License](LICENSE).
