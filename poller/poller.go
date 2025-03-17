package poller

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"lywsd03mmc-prom-expo-adv/collector"
	"time"
	"tinygo.org/x/bluetooth"
)

type Poller struct {
	adapter        *bluetooth.Adapter
	devicesMAC     []string
	scanTimeoutSec uint
}

var ErrDataParseErr = errors.New("Err while parsing payload")
var ErrWrongUUID = errors.New("Wrong UUID")
var ErrShortLen = errors.New("data len too short")
var ErrMismatchMAC = errors.New("Mismatch MAC")

// https://github.com/pvvx/ATC_MiThermometer?tab=readme-ov-file#custom-format-all-data-little-endian
func parseCustomPVVX(MustMAC [6]byte, UUID string, b []byte) (temp float64, humidity float64, voltage float64, err error) {
	if UUID != "0000181a-0000-1000-8000-00805f9b34fb" {
		return 0, 0, 0, ErrWrongUUID
	}

	if len(b) < 14 {
		return 0, 0, 0, fmt.Errorf("%w: data too short", ErrDataParseErr)
	}

	if bytes.Compare(b[:6], MustMAC[:]) != 0 {
		return 0, 0, 0, fmt.Errorf("%w: Mismatch MAC", ErrDataParseErr)
	}

	tmp := binary.LittleEndian.Uint16(b[6:8])
	temp = float64(tmp) / 100

	tmp = binary.LittleEndian.Uint16(b[8:10])
	humidity = float64(tmp) / 100

	tmp = binary.LittleEndian.Uint16(b[10:12])
	voltage = float64(tmp) / 1000

	return temp, humidity, voltage, nil
}

func (p *Poller) Poll() ([]collector.PollResult, error) {
	waitList := make(map[string]interface{})
	for _, v := range p.devicesMAC {
		waitList[v] = nil
	}
	scanResults := make([]collector.PollResult, 0, len(p.devicesMAC))

	succsessChan := make(chan interface{}, 1)
	scanCallErrChan := make(chan error, 1)
	timer := time.NewTimer(time.Second * time.Duration(p.scanTimeoutSec))

	go func() {
		err := p.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			scannedMAC := result.Address.String()

			_, ok := waitList[scannedMAC]
			// if this MAC into waiting list
			if ok {
				sd := result.AdvertisementPayload.ServiceData()
				if sd == nil || len(sd) == 0 {
					log.Println(scannedMAC, "empty AdvertisementPayload.ServiceData()")
					return
				}

				for _, v := range sd {
					temp, humidity, voltage, err := parseCustomPVVX(
						result.Address.MACAddress.MAC, v.UUID.String(), v.Data)
					if err != nil {
						// log parse errors
						// ignore uuid errors
						if errors.Is(err, ErrDataParseErr) {
							log.Println(scannedMAC, err.Error())
						}
						continue
					}
					scanResults = append(scanResults,
						collector.PollResult{
							MAC:      scannedMAC,
							Temp:     temp,
							Humidity: humidity,
							Voltage:  voltage,
						})

					// no wait it anymore
					delete(waitList, scannedMAC)
					break
				}
				if len(waitList) == 0 {
					succsessChan <- nil
				}
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
	case <-succsessChan:
		timer.Stop()
		p.adapter.StopScan()
		return scanResults, nil
	}
}

func NewDevicePoller(scanTimeoutSec uint, devicesMAC []string) (*Poller, error) {
	var adapter = bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}

	return &Poller{adapter, devicesMAC, scanTimeoutSec}, nil
}
