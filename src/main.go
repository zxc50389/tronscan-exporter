package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	tronTPSEndpoint = "https://apilist.tronscanapi.com/api/system/tps"
)

var (
	currentTPS = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tron_current_tps",
			Help: "Current Transactions per second on the TRON network",
		},
	)
	maxTPS = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tron_max_tps",
			Help: "Max Transactions per second on the TRON network",
		},
	)
	blockHeight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tron_blockHeight",
			Help: "tron BlockHeight on the TRON network",
		},
	)
)

func init() {
	prometheus.MustRegister(currentTPS)
	prometheus.MustRegister(maxTPS)
	prometheus.MustRegister(blockHeight)
}

type TPSData struct {
	Data struct {
		BlockHeight float64 `json:"blockHeight"`
		CurrentTps  float64 `json:"currentTps"`
		MaxTps      float64 `json:"maxTps"`
	} `json:"data"`
	Type string `json:"type"`
}

func fetchTPS() (*TPSData, error) {
	resp, err := http.Get(tronTPSEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data TPSData

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func updateMetrics() {
	for {
		tpsData, err := fetchTPS()
		if err == nil {
			currentTPS.Set(tpsData.Data.CurrentTps)
			maxTPS.Set(tpsData.Data.MaxTps)
			blockHeight.Set(tpsData.Data.BlockHeight)
		} else {
			fmt.Printf("Failed to fetch TPS: %v\n", err)
		}
		time.Sleep(10 * time.Second) // Update every 30 seconds
	}
}

func main() {
	go updateMetrics()
	metricsPath := "/metrics"
	listenAddress := ":8091"
	http.Handle(metricsPath, promhttp.Handler())
	//http.ListenAndServe(listenAddress, nil)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
            <html>
            <head><title>tron Exporter Metrics</title></head>
            <body>
            <h1>tron metrics</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>
        `))
	})
	fmt.Println(http.ListenAndServe(listenAddress, nil))
}
