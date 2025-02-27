package poller

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

type Poller struct {
	char           *bluetooth.DeviceCharacteristic
	devMAC         string
	scanTimeoutSec uint
}

// char data must be 5 bytes
// for example 71 09 1d 11 0b
// 0971 -> 2417 -> 24.17°C temp
// 1d           -> 29% humidity
// 0b11 -> 2833 -> 2.833V battery voltage
func (p *Poller) parse(b []byte) (temp float32, humidity uint8, voltage float32, err error) {
	if len(b) != 5 {
		return 0, 0, 0, fmt.Errorf("len(rawBytes) %v != 5", len(b))
	}
	tmp := binary.LittleEndian.Uint16(b[:2])
	temp = float32(tmp) / 100

	humidity = b[2]

	tmp2 := binary.LittleEndian.Uint16(b[3:5])
	voltage = float32(tmp2) / 1000

	return temp, humidity, voltage, nil
}

// this method is not thread safe
// coz adapter shared and adapter.scan() restrictions
func (p *Poller) Poll() (temp float32, humidity uint8, vlotage float32, err error) {
	easyerr := func(e error) (float32, uint8, float32, error) {
		return 0, 0, 0, e
	}

	if p.char == nil {
		var adapter = bluetooth.DefaultAdapter
		err := adapter.Enable()
		if err != nil {
			return easyerr(err)
		}

		succsessChan := make(chan bluetooth.ScanResult, 1)
		scanCallErrChan := make(chan error, 1)
		timer := time.NewTimer(time.Second * time.Duration(p.scanTimeoutSec))

		go func() {
			err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
				if result.Address.String() == p.devMAC {
					succsessChan <- result
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
		var result bluetooth.ScanResult
		select {
		case <-timer.C:
			adapter.StopScan()
			return easyerr(errors.New("scan timeout"))
		case e := <-scanCallErrChan:
			timer.Stop()
			return easyerr(e)
		case result = <-succsessChan:
			timer.Stop()
			adapter.StopScan()
		}

		dev, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			fmt.Println("connect err:", err.Error())
			return easyerr(err)
		}

		// ebe0ccb0-7a0a-4b0c-8a1a-6ff2997da3a6
		service, err := dev.DiscoverServices([]bluetooth.UUID{bluetooth.NewUUID([16]byte{
			0xeb, 0xe0, 0xcc, 0xb0, 0x7a, 0x0a, 0x4b, 0x0c,
			0x8a, 0x1a, 0x6f, 0xf2, 0x99, 0x7d, 0xa3, 0xa6,
		})})
		if err != nil {
			return easyerr(err)
		}
		if len(service) < 1 {
			return easyerr(errors.New("dev.DiscoverService() return empty array"))
		}

		// ebe0ccc1-7a0a-4b0c-8a1a-6ff2997da3a6
		char, err := service[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.NewUUID([16]byte{
			0xeb, 0xe0, 0xcc, 0xc1, 0x7a, 0x0a, 0x4b, 0x0c,
			0x8a, 0x1a, 0x6f, 0xf2, 0x99, 0x7d, 0xa3, 0xa6})})
		if err != nil {
			return easyerr(err)
		}
		if len(char) < 1 {
			return easyerr(errors.New("service.DiscoverCharacteristics() return empty array"))
		}

		p.char = &char[0]
	}

	rawBytes := make([]byte, 5)
	_, err = p.char.Read(rawBytes)
	if err != nil {
		// clear char to force scan() and discovery() on next poll
		p.char = nil
		return easyerr(err)
	}

	// if no call disconnect() we can avoid scan() and discovery() on next poll
	//p.char = nil
	//err = dev.Disconnect()

	return p.parse(rawBytes)
}

func NewDevicePoller(devMAC string, scanTimeoutSec uint) *Poller {
	return &Poller{devMAC: devMAC, scanTimeoutSec: scanTimeoutSec}
}
