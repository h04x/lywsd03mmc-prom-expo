package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SensorData struct {
	updated  time.Time
	temp     float64
	humidity uint
	battery  float64
}

type SensorCollector struct {
	tempMetricDesc     *prometheus.Desc
	humidityMetricDesc *prometheus.Desc
	batteryMetricDesc  *prometheus.Desc
	m                  sync.Mutex
	sensorsData        map[string]SensorData
}

func (c *SensorCollector) UpdateSensorData(mac string, temp float64, humidity uint, battery float64) {
	updated := time.Now()
	c.m.Lock()
	c.sensorsData[mac] = SensorData{updated, temp, humidity, battery}
	c.m.Unlock()
}

func (c *SensorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tempMetricDesc
	ch <- c.humidityMetricDesc
	ch <- c.batteryMetricDesc
}

func (c *SensorCollector) Collect(ch chan<- prometheus.Metric) {
	c.m.Lock()
	for mac, data := range c.sensorsData {
		m1 := prometheus.NewMetricWithTimestamp(data.updated, prometheus.MustNewConstMetric(c.tempMetricDesc, prometheus.GaugeValue, data.temp, mac))
		m2 := prometheus.NewMetricWithTimestamp(data.updated, prometheus.MustNewConstMetric(c.humidityMetricDesc, prometheus.GaugeValue, float64(data.humidity), mac))
		m3 := prometheus.NewMetricWithTimestamp(data.updated, prometheus.MustNewConstMetric(c.batteryMetricDesc, prometheus.GaugeValue, data.battery, mac))
		ch <- m1
		ch <- m2
		ch <- m3
	}
	c.m.Unlock()
}

func NewSensorCollector() *SensorCollector {
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
		sensorsData: make(map[string]SensorData),
	}
}
