package main

import (
	"flag"
	"fmt"
	"log"

	"lywsd03mmc-prom-expo/collector"
	"lywsd03mmc-prom-expo/poller"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")
var scanTimeoutSec = flag.Uint("bt-scan-timeout", 31, "Bluetooth scan timeout in seconds.")

func main() {
	flag.Parse()

	p, err := poller.NewDevicePoller(*scanTimeoutSec, []string{
		"A4:C1:38:B4:96:0B",
		"A4:C1:38:8A:3B:DE",
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	c := collector.NewSensorCollector(p)

	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
