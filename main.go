package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"lywsd03mmc-prom-expo/collector"
	"lywsd03mmc-prom-expo/poller"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jessevdk/go-flags"
)

type MAC net.HardwareAddr

func (m *MAC) UnmarshalFlag(value string) error {
	mac, err := net.ParseMAC(value)
	if err != nil {
		return errors.New("MAC is invalid")
	}

	if len(mac) != 6 {
		return errors.New("MAC addr len != 6 bytes")
	}

	*m = MAC(mac)
	return nil
}

func (p MAC) MarshalFlag() (string, error) {
	return string(p), nil
}

func (m MAC) BytesLE() [6]byte {
	var b [6]byte
	b[0] = m[5]
	b[1] = m[4]
	b[2] = m[3]
	b[3] = m[2]
	b[4] = m[1]
	b[5] = m[0]
	return b
}

type MyDuration time.Duration

func (md *MyDuration) UnmarshalFlag(value string) error {
	d, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	*md = MyDuration(d)
	return nil
}

func (md MyDuration) MarshalFlag() (string, error) {
	return fmt.Sprintf("%v", md), nil
}

func (md MyDuration) Duration() time.Duration {
	return time.Duration(md)
}

type Options struct {
	ListenAddress string     `long:"listen-address" default:"127.0.0.1:8080" description:"The address to listen on for HTTP requests."`
	ScanTimeout   MyDuration `long:"bt-scan-timeout" default:"31s" description:"Bluetooth scan timeout (e.g 31s, 1h, 8m)"`
	Devices       []MAC      `long:"dev" required:"true" description:"Bluetooth device MAC address to poll. Repeatable (e.g. AA:BB:CC:DD:EE:FF)"`
}

var opt Options

var parser = flags.NewParser(&opt, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		return
	}

	p, err := poller.NewContinuousPoller(opt.ScanTimeout.Duration(), time.Second*3)
	if err != nil {
		log.Fatal("poller create failed:", err.Error())
	}

	for _, d := range opt.Devices {
		p.NewDevice(d)
	}

	go p.Scan()

	c := collector.NewSensorCollector(p)

	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	err = http.ListenAndServe(opt.ListenAddress, nil)
	if err != nil {
		log.Fatal("listen failed:", err.Error())
	}
}
