package poller

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
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
