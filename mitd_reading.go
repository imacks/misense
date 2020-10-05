package misense

import (
	"math"
	"fmt"
)

// THReading represents a temperature and humidity sensor reading.
// Works with LYWSD03MMC. 
type THReading struct {
	Temperature float64
	Humidity    float64
	Voltage     float64
	macaddr     string
}

// NewTHReading returns a new THReading
func NewTHReading(t, h, volt float64, macaddr string) *THReading {
	return &THReading{
		Temperature: t,
		Humidity:    h,
		Voltage:     volt,
		macaddr:     macaddr,
	}
}

// MAC returns the sensor MAC address
func (r *THReading)  MAC() string {
	return r.macaddr
}

// Battery returns the current battery level based on voltage
func (r *THReading) Battery() int {
	// 2.1 means 0%, 3.1 means full
	return int(math.Min((r.Voltage - 2.1) * 100, 100))
}

// String returns a string representation of a THReading
func (r *THReading) String() string {
	return fmt.Sprintf("[A] %s [T] %.04f [d] %.04f [E/V] %d | %.04f", r.MAC(), r.Temperature, r.Humidity, r.Battery(), r.Voltage)
}