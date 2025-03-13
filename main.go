package main

import (
	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"lywsd03mmc-prom-expo/collector"
	"lywsd03mmc-prom-expo/poller"
	"net/http"
)

const scanTimeoutSec = 10

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	pollers := []collector.Poller{
		poller.NewDevicePoller("A4:C1:38:8A:3B:DE", scanTimeoutSec),
		poller.NewDevicePoller("A4:C1:38:B4:96:0B", scanTimeoutSec),
	}

	c := collector.NewSensorCollector(pollers)

	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	http.ListenAndServe(*addr, nil)
}
