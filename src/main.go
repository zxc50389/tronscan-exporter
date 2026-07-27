package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultTPSEndpoint = "https://apilist.tronscanapi.com/api/system/tps"

var (
	currentTPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tron_current_tps",
		Help: "Current transactions per second on the TRON network.",
	})
	maxTPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tron_max_tps",
		Help: "Maximum transactions per second on the TRON network.",
	})
	blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tron_block_height",
		Help: "Current TRON block height.",
	})
	up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tron_up",
		Help: "Whether the last scrape of the TronScan API succeeded (1) or failed (0).",
	})
)

func init() {
	prometheus.MustRegister(currentTPS, maxTPS, blockHeight, up)
}

// TPSData mirrors the relevant fields of the TronScan /api/system/tps response.
type TPSData struct {
	Data struct {
		BlockHeight float64 `json:"blockHeight"`
		CurrentTps  float64 `json:"currentTps"`
		MaxTps      float64 `json:"maxTps"`
	} `json:"data"`
	Type string `json:"type"`
}

// fetchTPS retrieves the current TPS snapshot from the TronScan API. It returns
// an error on transport failures, non-200 responses, or invalid JSON.
func fetchTPS(ctx context.Context, client *http.Client, endpoint string) (*TPSData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data TPSData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

// updateMetrics polls the TronScan API every interval until ctx is cancelled,
// updating the exported gauges. On any failure it sets tron_up to 0 and keeps
// the previous readings so stale data is detectable via tron_up.
func updateMetrics(ctx context.Context, client *http.Client, endpoint string, interval time.Duration) {
	scrape := func() {
		data, err := fetchTPS(ctx, client, endpoint)
		if err != nil {
			up.Set(0)
			log.Printf("failed to fetch TPS: %v", err)
			return
		}
		up.Set(1)
		currentTPS.Set(data.Data.CurrentTps)
		maxTPS.Set(data.Data.MaxTps)
		blockHeight.Set(data.Data.BlockHeight)
	}

	scrape() // scrape once at startup so metrics are populated immediately

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scrape()
		}
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":8091", "Address on which to expose metrics and the landing page.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		endpoint      = flag.String("tron.tps-endpoint", defaultTPSEndpoint, "TronScan TPS API endpoint to poll.")
		interval      = flag.Duration("scrape.interval", 10*time.Second, "How often to poll the TronScan API.")
		timeout       = flag.Duration("scrape.timeout", 10*time.Second, "Timeout for each TronScan API request.")
	)
	flag.Parse()

	client := &http.Client{Timeout: *timeout}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go updateMetrics(ctx, client, *endpoint, *interval)

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html>
<head><title>TRON Exporter</title></head>
<body>
<h1>TRON Exporter</h1>
<p><a href='%s'>Metrics</a></p>
</body>
</html>`, *metricsPath)
	})

	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", *listenAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
