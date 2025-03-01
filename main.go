package main

import (
	"flag"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"lywsd03mmc-prom-expo/collector"
	"lywsd03mmc-prom-expo/poller"
	"net/http"
	"time"
)

const scanTimeoutSec = 10
const pollPeriodMin = 15

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	c := collector.NewSensorCollector()

	go func() {
		pollers := []*poller.Poller{
			poller.NewDevicePoller("A4:C1:38:8A:3B:DE", scanTimeoutSec),
			poller.NewDevicePoller("A4:C1:38:B4:96:0B", scanTimeoutSec),
		}

		for {
			for _, p := range pollers {
				temp, humidity, battery, err := p.Poll()
				if err != nil {
					fmt.Println(err.Error())
					time.Sleep(time.Second)
					continue
				}
				c.UpdateSensorData(p.Mac(), temp, uint(humidity), battery)
			}

			time.Sleep(time.Minute * pollPeriodMin)
		}
	}()

	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	http.ListenAndServe(*addr, nil)
}
