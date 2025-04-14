package pollerContinuous

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"maps"
	//"slices"
	"sync"
	"time"

	"lywsd03mmc-prom-expo/collector"

	"tinygo.org/x/bluetooth"
)

type sixBytes = [6]byte

type MACAddr interface {
	BytesLE() sixBytes
}

const expectedUUID string = "0000181a-0000-1000-8000-00805f9b34fb"

var ErrMismatchUUID = fmt.Errorf("mismatch UUID")

const scanRestartDelay = time.Second * 3

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

type PollerContinuous struct {
	adapter        *bluetooth.Adapter
	deviceMutex    *sync.Mutex
	devicesMAC     map[sixBytes]*collector.PollResult
	scanTimeoutSec uint
}

func New(scanTimeoutSec uint) (*PollerContinuous, error) {
	var adapter = bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}

	devices := make(map[sixBytes]*collector.PollResult)
	var m sync.Mutex
	return &PollerContinuous{adapter, &m, devices, scanTimeoutSec}, nil

}

func (p *PollerContinuous) NewDevice(device MACAddr) {
	p.deviceMutex.Lock()
	p.devicesMAC[device.BytesLE()] = nil
	p.deviceMutex.Unlock()
}

func (p *PollerContinuous) Scan() {
	heartbeat := make(chan interface{})
	for {
		go func() {
			period := time.Second * time.Duration(p.scanTimeoutSec)
			ticker := time.NewTicker(period)
			for {
				select {
				case <-heartbeat:
					ticker.Reset(period)
				case <-ticker.C:
					log.Printf("no bluetooth message received in %v, restart adapter.Scan()", period)
					p.adapter.StopScan()
					return
				}
			}
		}()

		err := p.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			heartbeat <- nil

			scannedMAC := result.Address.MAC
			p.deviceMutex.Lock()
			_, ok := p.devicesMAC[scannedMAC]
			p.deviceMutex.Unlock()
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
					p.deviceMutex.Lock()
					p.devicesMAC[scannedMAC] = &collector.PollResult{
						MAC:       result.Address.String(),
						Temp:      temp,
						Humidity:  humidity,
						Voltage:   voltage,
						Timestamp: time.Now(),
					}
					p.deviceMutex.Unlock()
					break
				}
			}
		})
		if err != nil {
			log.Println("adapter.Scan() return error", err)
		}
		log.Println("scan stopped, restart in", scanRestartDelay)
		time.Sleep(scanRestartDelay)
	}
}

func (p *PollerContinuous) Poll() ([]collector.PollResult, error) {
	p.deviceMutex.Lock()
	s := make([]collector.PollResult, 0)
	for v := range maps.Values(p.devicesMAC) {
		if v != nil {
			s = append(s, *v)
		}
	}
	p.deviceMutex.Unlock()
	return s, nil
}
