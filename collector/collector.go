package collector

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PollResult struct {
	MAC       string
	Temp      float64
	Humidity  float64
	Voltage   float64
	Timestamp time.Time
}

type Poller interface {
	Poll() ([]PollResult, error)
}

type SensorCollector struct {
	tempMetricDesc     *prometheus.Desc
	humidityMetricDesc *prometheus.Desc
	batteryMetricDesc  *prometheus.Desc
	poller             Poller
}

func (c *SensorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tempMetricDesc
	ch <- c.humidityMetricDesc
	ch <- c.batteryMetricDesc
}

func (c *SensorCollector) Collect(ch chan<- prometheus.Metric) {
	scanResult, err := c.poller.Poll()
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, v := range scanResult {
		m1 := prometheus.NewMetricWithTimestamp(v.Timestamp, prometheus.MustNewConstMetric(c.tempMetricDesc,
			prometheus.GaugeValue, v.Temp, v.MAC))
		m2 := prometheus.NewMetricWithTimestamp(v.Timestamp, prometheus.MustNewConstMetric(c.humidityMetricDesc,
			prometheus.GaugeValue, float64(v.Humidity), v.MAC))
		m3 := prometheus.NewMetricWithTimestamp(v.Timestamp, prometheus.MustNewConstMetric(c.batteryMetricDesc,
			prometheus.GaugeValue, v.Voltage, v.MAC))
		ch <- m1
		ch <- m2
		ch <- m3
	}
}

func NewSensorCollector(poller Poller) *SensorCollector {
	return &SensorCollector{
		tempMetricDesc: prometheus.NewDesc(
			"sensor_temp_celsius",
			"Temperature in celsius",
			[]string{"mac"},
			nil,
		),
		humidityMetricDesc: prometheus.NewDesc(
			"sensor_humidity_percent",
			"Humidity in percent",
			[]string{"mac"},
			nil,
		),
		batteryMetricDesc: prometheus.NewDesc(
			"sensor_battery_volts",
			"Battery voltage",
			[]string{"mac"},
			nil,
		),
		poller: poller,
	}
}
