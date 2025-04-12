package pollerOnDemand

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"lywsd03mmc-prom-expo/collector"
	"time"

	"tinygo.org/x/bluetooth"
)

type sixBytes = [6]byte

type MACAddr interface {
	BytesLE() sixBytes
}

type PollerOnDemand struct {
	adapter        *bluetooth.Adapter
	devicesMAC     []sixBytes
	scanTimeoutSec uint
}

const expectedUUID string = "0000181a-0000-1000-8000-00805f9b34fb"

var ErrMismatchUUID = fmt.Errorf("mismatch UUID")

// https://github.com/pvvx/ATC_MiThermometer?tab=readme-ov-file#custom-format-all-data-little-endian
func parseCustomPVVX(expectedMAC [6]byte, advertisedUUID string, b []byte) (temp float64, humidity float64, voltage float64, err error) {
	if advertisedUUID != expectedUUID {
		return 0, 0, 0, fmt.Errorf("%w: expected %v != adverised %v",
			ErrMismatchUUID, expectedUUID, advertisedUUID)
	}

	if len(b) < 14 {
		return 0, 0, 0, errors.New("advertised bytes len < 14")
	}

	advertisedMAC := b[:6]
	if bytes.Compare(advertisedMAC, expectedMAC[:]) != 0 {
		return 0, 0, 0, fmt.Errorf("in advertised bytes ExpectedMAC(% x) != AdvertisedMac(% x)",
			expectedMAC, advertisedMAC)
	}

	tmp := binary.LittleEndian.Uint16(b[6:8])
	temp = float64(tmp) / 100

	tmp = binary.LittleEndian.Uint16(b[8:10])
	humidity = float64(tmp) / 100

	tmp = binary.LittleEndian.Uint16(b[10:12])
	voltage = float64(tmp) / 1000

	return temp, humidity, voltage, nil
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
				if sd == nil || len(sd) == 0 {
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

func New(scanTimeoutSec uint) (*PollerOnDemand, error) {
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
