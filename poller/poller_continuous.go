package poller

import (
	"log"
	"maps"
	//"slices"
	"sync"
	"time"

	"lywsd03mmc-prom-expo/collector"

	"tinygo.org/x/bluetooth"
)

type PollerContinuous struct {
	adapter          *bluetooth.Adapter
	devicesMutex     *sync.Mutex
	devices          map[sixBytes]*collector.PollResult
	scanTimeout      time.Duration
	scanRestartDelay time.Duration
}

func NewContinuousPoller(scanTimeout time.Duration, scanRestartDelay time.Duration) (*PollerContinuous, error) {
	var adapter = bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}

	devices := make(map[sixBytes]*collector.PollResult)
	var m sync.Mutex
	return &PollerContinuous{adapter, &m, devices, scanTimeout, scanRestartDelay}, nil

}

func (p *PollerContinuous) NewDevice(device MACAddr) {
	p.devicesMutex.Lock()
	p.devices[device.BytesLE()] = nil
	p.devicesMutex.Unlock()
}

func (p *PollerContinuous) Scan() {
	heartbeat := make(chan interface{})
	for {
		go func() {
			ticker := time.NewTicker(p.scanTimeout)
			for {
				select {
				case <-heartbeat:
					ticker.Reset(p.scanTimeout)
				case <-ticker.C:
					log.Printf("no bluetooth message received in %v, restart adapter.Scan()", p.scanTimeout)
					p.adapter.StopScan()
					return
				}
			}
		}()

		err := p.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			heartbeat <- nil

			scannedMAC := result.Address.MAC
			p.devicesMutex.Lock()
			_, ok := p.devices[scannedMAC]
			p.devicesMutex.Unlock()
			// if this MAC into waiting list
			if ok {
				sd := result.AdvertisementPayload.ServiceData()
				if len(sd) == 0 {
					log.Println(scannedMAC, "empty AdvertisementPayload.ServiceData()")
					return
				}

				for _, v := range sd {
					temp, humidity, voltage, err := parseCustomPVVX(
						scannedMAC, v.UUID.String(), v.Data)
					if err != nil {
						log.Println(scannedMAC, err.Error())
						continue
					}
					p.devicesMutex.Lock()
					p.devices[scannedMAC] = &collector.PollResult{
						MAC:       result.Address.String(),
						Temp:      temp,
						Humidity:  humidity,
						Voltage:   voltage,
						Timestamp: time.Now(),
					}
					p.devicesMutex.Unlock()
					break
				}
			}
		})
		if err != nil {
			log.Println("adapter.Scan() return error", err)
		}
		log.Println("scan stopped, restart in", p.scanRestartDelay)
		time.Sleep(p.scanRestartDelay)
	}
}

func (p *PollerContinuous) Poll() ([]collector.PollResult, error) {
	p.devicesMutex.Lock()
	s := make([]collector.PollResult, 0)
	for v := range maps.Values(p.devices) {
		if v != nil {
			s = append(s, *v)
		}
	}
	p.devicesMutex.Unlock()
	return s, nil
}
