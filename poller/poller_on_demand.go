package poller

import (
	"errors"
	"log"
	"lywsd03mmc-prom-expo/collector"
	"time"

	"tinygo.org/x/bluetooth"
)

type PollerOnDemand struct {
	adapter        *bluetooth.Adapter
	devicesMAC     []sixBytes
	scanTimeoutSec uint
}

func (p *PollerOnDemand) Poll() ([]collector.PollResult, error) {
	pendingDevices := make(map[sixBytes]interface{})
	for _, v := range p.devicesMAC {
		pendingDevices[v] = nil
	}
	scanResults := make([]collector.PollResult, 0, len(p.devicesMAC))

	successChan := make(chan interface{}, 1)
	scanCallErrChan := make(chan error, 1)
	timer := time.NewTimer(time.Second * time.Duration(p.scanTimeoutSec))

	go func() {
		err := p.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			scannedMAC := result.Address.MAC
			_, ok := pendingDevices[scannedMAC]
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
					scanResults = append(scanResults,
						collector.PollResult{
							MAC:       result.Address.String(),
							Temp:      temp,
							Humidity:  humidity,
							Voltage:   voltage,
							Timestamp: time.Now(),
						})

					// no wait it anymore
					delete(pendingDevices, scannedMAC)
					break
				}
			}
			if len(pendingDevices) == 0 {
				successChan <- nil
			}
		})
		if err != nil {
			scanCallErrChan <- err
		}
	}()

	// waiting what happens first of three:
	// timeout
	// scan() call return error
	// scan() call succsess
	select {
	case <-timer.C:
		p.adapter.StopScan()
		log.Println(errors.New("scan timeout"))
		return scanResults, nil
	case e := <-scanCallErrChan:
		timer.Stop()
		return nil, e
	case <-successChan:
		timer.Stop()
		p.adapter.StopScan()
		return scanResults, nil
	}
}

func NewPollerOnDemand(scanTimeoutSec uint) (*PollerOnDemand, error) {
	var adapter = bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}

	devices := []sixBytes{}
	return &PollerOnDemand{adapter, devices, scanTimeoutSec}, nil

}

func (p *PollerOnDemand) NewDevice(device MACAddr) {
	p.devicesMAC = append(p.devicesMAC, device.BytesLE())
}
