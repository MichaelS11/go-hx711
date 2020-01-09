// +build !windows,rpio

package hx711

import (
	"fmt"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

// HostInit opens /dev/gpiomem. This needs to be done before Hx711 can be used.
func HostInit() error {
	return rpio.Open()
}

// NewHx711 creates new Hx711.
// Make sure to set clockPin and dataPin to the correct pins.
// https://cdn.sparkfun.com/datasheets/Sensors/ForceFlex/hx711_english.pdf
func NewHx711(clockPinStr string, dataPinStr string) (*Hx711, error) {
	clockPin, err := strconv.ParseInt(clockPinStr, 10, 32)
	if err != nil {
		return nil, err
	}
	dataPin, err := strconv.ParseInt(dataPinStr, 10, 32)
	if err != nil {
		return nil, err
	}
	hx711 := &Hx711{numEndPulses: 1}
	hx711.clockPin = rpio.Pin(int(clockPin))
	hx711.dataPin = rpio.Pin(int(dataPin))
	hx711.dataPin.Input()
	hx711.clockPin.Output()
	return hx711, nil
}

// setClockHighThenLow sets clock pin high then low
func (hx711 *Hx711) setClockHighThenLow() error {
	hx711.clockPin.Write(rpio.High)
	hx711.clockPin.Write(rpio.Low)
	return nil
}

// Reset starts up or resets the chip.
// The chip needs to be reset if it is not used for just about any amount of time.
func (hx711 *Hx711) Reset() error {
	hx711.clockPin.Write(rpio.Low)
	hx711.clockPin.Write(rpio.High)
	time.Sleep(70 * time.Microsecond)
	hx711.clockPin.Write(rpio.Low)
	return nil
}

// Shutdown puts the chip in powered down mode.
// The chip should be shutdown if it is not used for just about any amount of time.
func (hx711 *Hx711) Shutdown() error {
	hx711.clockPin.Write(rpio.High)
	return nil
}

// waitForDataReady waits for data to go to low which means chip is ready
func (hx711 *Hx711) waitForDataReady() error {
	var level rpio.State

	// looks like chip often takes 80 to 100 milliseconds to get ready
	// but somettimes it takes around 500 milliseconds to get ready
	// WaitForEdge sometimes returns right away
	// So will loop for N times, which could be more than 1 second, but usually 500 milliseconds

	hx711.dataPin.Detect(rpio.FallEdge)
	hx711.clockPin.Write(rpio.Low)
	defer hx711.dataPin.Detect(rpio.NoEdge)
	level = hx711.dataPin.Read()
	if level == rpio.Low {
		return nil
	}
	for i := 0; i < 200000; i++ {
		if !hx711.dataPin.EdgeDetected() {
			time.Sleep(5 * time.Microsecond)
		} else {
			return nil
		}
	}

	return ErrTimeout
}

// ReadDataRaw will get one raw reading from chip.
// Usually will need to call Reset before calling this and Shutdown after.
func (hx711 *Hx711) ReadDataRaw() (int, error) {
	err := hx711.waitForDataReady()
	if err != nil {
		return 0, fmt.Errorf("waitForDataReady error: %v", err)
	}

	var level rpio.State
	var data int
	for i := 0; i < 24; i++ {
		err = hx711.setClockHighThenLow()
		if err != nil {
			return 0, fmt.Errorf("setClockHighThenLow error: %v", err)
		}

		level = hx711.dataPin.Read()
		data = data << 1
		if level == rpio.High {
			data++
		}
	}

	for i := 0; i < hx711.numEndPulses; i++ {
		err = hx711.setClockHighThenLow()
		if err != nil {
			return 0, fmt.Errorf("setClockHighThenLow error: %v", err)
		}
	}

	// if high 24 bit is set, value is negtive
	// 100000000000000000000000
	if (data & 0x800000) > 0 {
		// flip bits 24 and lower to get negtive number for int
		// 111111111111111111111111
		data |= ^0xffffff
	}

	return data, nil
}
