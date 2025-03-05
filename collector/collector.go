package collector

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SensorData struct {
	updated  time.Time
	temp     float64
	humidity uint
	battery  float64
}

type Poller interface {
	Poll() (temp float64, humidity uint8, vlotage float64, err error)
	Mac() string
}

type SensorCollector struct {
	tempMetricDesc     *prometheus.Desc
	humidityMetricDesc *prometheus.Desc
	batteryMetricDesc  *prometheus.Desc
	pollers            []Poller
}

func (c *SensorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tempMetricDesc
	ch <- c.humidityMetricDesc
	ch <- c.batteryMetricDesc
}

func (c *SensorCollector) Collect(ch chan<- prometheus.Metric) {
	for _, p := range c.pollers {
		temp, humidity, battery, err := p.Poll()
		if err != nil {
			log.Println(err.Error())
			continue
		}

		m1 := prometheus.MustNewConstMetric(c.tempMetricDesc,
			prometheus.GaugeValue, temp, p.Mac())
		m2 := prometheus.MustNewConstMetric(c.humidityMetricDesc,
			prometheus.GaugeValue, float64(humidity), p.Mac())
		m3 := prometheus.MustNewConstMetric(c.batteryMetricDesc,
			prometheus.GaugeValue, battery, p.Mac())

		ch <- m1
		ch <- m2
		ch <- m3
	}

}

func NewSensorCollector(pollers []Poller) *SensorCollector {
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
		pollers: pollers,
	}
}
