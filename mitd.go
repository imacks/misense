package misense

import (
	"fmt"
	"context"
	"time"
	"encoding/binary"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
)

// MiTDSensor is a sensor from Mijia2
type MiTDSensor struct {
	macaddr string
	host *linux.Device
	client ble.Client
	profile *ble.Profile
}

// NewMiTDSensor creates a new Mi temperature and humidity sensor interface
func NewMiTDSensor(host *linux.Device, macaddr string) *MiTDSensor {
	return &MiTDSensor{
		host: host,
		macaddr: macaddr,
	}
}

// Connect will connect the host to the sensor
func (tds *MiTDSensor) Connect(timeout time.Duration) error {
	var err error

	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), timeout))
	tds.client, err = tds.host.Dial(ctx, ble.NewAddr(tds.macaddr))
	if err != nil {
		return err
	}

	tds.profile, err = tds.client.DiscoverProfile(true)
	if err != nil {
		return err
	}

	return nil
}

// Disconnect will disconnect host from the sensor
func (tds *MiTDSensor) Disconnect() error {
	err := tds.client.CancelConnection()
	if err != nil {
		return err
	}
	return nil
}

// Version reports the supported sensor version 
func (tds *MiTDSensor) Version() string {
	return "LYWSD03MMC"
}

// MAC returns the sensor MAC address
func (tds *MiTDSensor) MAC() string {
	return tds.macaddr
}

// Subscribe adds a hook to the notification API
func (tds *MiTDSensor) Subscribe(handler func(r *THReading)) error {
	c := getCharacteristicByValueHandle(tds.profile, 0x36)
	if c == nil {
		return fmt.Errorf("characteristic 0x36 not found")
	}

	err := tds.client.Subscribe(c, false, func(req []byte) {
		// https://github.com/AnthonyKNorman/Xiaomi_LYWSD03MMC_for_HA/blob/master/ble.py
		// 00 01 02 03 04
		// T2 T1 HX V1 V2
		// * T1-T2 is the temperature as signed INT16 in little endian. divide by 100 to get the temp in degree c
		// * HX    is the humidity. only integer output!
		// * V1-V2 is battery voltage in millivolts in little endian

		if len(req) != 5 {
			fmt.Printf("expect 5 bytes but got %d\n", len(req))
			return
		}

		t := float64(int(binary.LittleEndian.Uint16(req[0:2]))) / 100.0
		h := float64(req[2]) / 100.0
		volt := float64(int(binary.LittleEndian.Uint16(req[3:]))) / 1000.0

		handler(NewTHReading(t, h, volt, tds.MAC()))
	})
	if err != nil {
		return err
	}

	return nil
}

// Time returns the sensor RTC time
func (tds *MiTDSensor) Time() (time.Time, error) {
	c := getCharacteristicByValueHandle(tds.profile, 0x23)
	if c == nil {
		return time.Unix(0, 0), fmt.Errorf("characteristic 0x23 not found")
	}

	data, err := tds.client.ReadCharacteristic(c)
	if err != nil {
		return time.Unix(0, 0), err
	}

	if len(data) != 4 {
		return time.Unix(0, 0), fmt.Errorf("expect 4 bytes but got %d", len(data))
	}

	v := binary.LittleEndian.Uint32(data)

	return time.Unix(int64(v), 0), nil
}

// SetTime sets the sensor RTC time
func (tds *MiTDSensor) SetTime(t time.Time) error {
	c := getCharacteristicByValueHandle(tds.profile, 0x23)
	if c == nil {
		return fmt.Errorf("characteristic 0x23 not found")
	}

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(t.Unix()))

	err := tds.client.WriteCharacteristic(c, data, true)
	if err != nil {
		return err
	}

	return nil
}

// EnableNotify makes the sensor send sensor data notifications
func (tds *MiTDSensor) EnableNotify() error {
	d := getDescriptorByHandle(tds.profile, 0x38)
	if d == nil {
		return fmt.Errorf("descriptor %d not found", 0x38)
	}

	err := tds.client.WriteDescriptor(d, []byte{0x01, 0x00})
	if err != nil {
		return err
	}

	return nil
}

// SavePower reduces sensor power consumption by reducing report interval.
// See https://github.com/JsBergbau/MiTemperature2/issues/18#issuecomment-590986874
func (tds *MiTDSensor) SavePower() error {
	c := getCharacteristicByValueHandle(tds.profile, 0x46)
	if c == nil {
		return fmt.Errorf("characteristic 0x46 not found")
	}

	err := tds.client.WriteCharacteristic(c, []byte{0xf4, 0x01, 0x00}, true)
	if err != nil {
		return err
	}

	return nil	
}

// Comfortable reports the comfortable temperature and humidity ranges
func (tds *MiTDSensor) Comfortable() (float64, float64, int, int, error) {
	c := getCharacteristicByValueHandle(tds.profile, 0x43)
	if c == nil {
		return -1, -1, -1, -1, fmt.Errorf("characteristic 0x43 not found")
	}

	data, err := tds.client.ReadCharacteristic(c)
	if err != nil {
		return -1, -1, -1, -1, err
	}

	if len(data) != 6 {
		return -1, -1, -1, -1, fmt.Errorf("expect 6 bytes but got %d", len(data))
	}

	// T T t t M m
	
	maxH := uint8(data[4])
	minH := uint8(data[5])
	maxT := float64(binary.LittleEndian.Uint16(data[0:2])) / 100
	minT := float64(binary.LittleEndian.Uint16(data[2:4])) / 100

	return maxT, minT, int(maxH), int(minH), nil
}

// SetComfortable sets the comfortable temperature and humidity ranges.
// Emoji on sensor will change according to actual value within range (adapted from manufacturer manual):
// ```
// |-----------|--------|-----------|-------|
// | Humidity  | < minT | minT~maxT | >maxT |
// |-----------|--------|-----------|-------|
// | <minH     | (-^-)  | (-^-)     | (-^-) |
// | minH~maxH | (-^-)  | (^_^)     | (-^-) |
// | >maxH     | (-^-)  | (-^-)     | (-^-) |
// ```
func (tds *MiTDSensor) SetComfortable(maxT, minT float64, maxH, minH int) error {
	if maxT < 0 || minT < 0 || maxT > 99 || minT > 99 {
		return fmt.Errorf("temperature range out of bounds")
	} else if minT > maxT {
		return fmt.Errorf("min temperature must be lower than max temperature")
	}

	if maxH < 0 || minH < 0 || maxH > 255 || minH > 255 {
		return fmt.Errorf("humidity range out of bounds")
	} else if minH > maxH {
		return fmt.Errorf("min humidity must be lower than max humidity")
	}

	c := getCharacteristicByValueHandle(tds.profile, 0x43)
	if c == nil {
		return fmt.Errorf("characteristic 0x43 not found")
	}

	// T T t t M m

	data := make([]byte, 6)

	data[4] = byte(maxH)
	data[5] = byte(minH)
	
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(maxT * 100))
	data[0] = buf[0]
	data[1] = buf[1]

	binary.LittleEndian.PutUint16(buf, uint16(minT * 100))
	data[2] = buf[0]
	data[3] = buf[1]

	err := tds.client.WriteCharacteristic(c, data, true)
	if err != nil {
		return err
	}

	return nil
}