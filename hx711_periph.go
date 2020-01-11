// +build !windows,!gpiomem

package hx711

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

// HostInit calls periph.io host.Init(). This needs to be done before Hx711 can be used.
func HostInit() error {
	_, err := host.Init()
	return err
}

// NewHx711 creates new Hx711.
// Make sure to set clockPinName and dataPinName to the correct pins.
// https://cdn.sparkfun.com/datasheets/Sensors/ForceFlex/hx711_english.pdf
func NewHx711(clockPinName string, dataPinName string) (*Hx711, error) {
	hx711 := &Hx711{numEndPulses: 1}

	hx711.clockPin = gpioreg.ByName(clockPinName)
	if hx711.clockPin == nil {
		return nil, fmt.Errorf("clockPin is nill")
	}

	hx711.dataPin = gpioreg.ByName(dataPinName)
	if hx711.dataPin == nil {
		return nil, fmt.Errorf("dataPin is nill")
	}

	err := hx711.dataPin.In(gpio.PullNoChange, gpio.FallingEdge)
	if err != nil {
		return nil, fmt.Errorf("dataPin setting to in error: %v", err)
	}

	return hx711, nil
}

// setClockHighThenLow sets clock pin high then low
func (hx711 *Hx711) setClockHighThenLow() error {
	err := hx711.clockPin.Out(gpio.High)
	if err != nil {
		return fmt.Errorf("set clock pin to high error: %v", err)
	}
	err = hx711.clockPin.Out(gpio.Low)
	if err != nil {
		return fmt.Errorf("set clock pin to low error: %v", err)
	}
	return nil
}

// Reset starts up or resets the chip.
// The chip needs to be reset if it is not used for just about any amount of time.
func (hx711 *Hx711) Reset() error {
	err := hx711.clockPin.Out(gpio.Low)
	if err != nil {
		return fmt.Errorf("set clock pin to low error: %v", err)
	}
	err = hx711.clockPin.Out(gpio.High)
	if err != nil {
		return fmt.Errorf("set clock pin to high error: %v", err)
	}
	time.Sleep(70 * time.Microsecond)
	err = hx711.clockPin.Out(gpio.Low)
	if err != nil {
		return fmt.Errorf("set clock pin to low error: %v", err)
	}
	return nil
}

// Shutdown puts the chip in powered down mode.
// The chip should be shutdown if it is not used for just about any amount of time.
func (hx711 *Hx711) Shutdown() error {
	err := hx711.clockPin.Out(gpio.High)
	if err != nil {
		return fmt.Errorf("set clock pin to high error: %v", err)
	}
	return nil
}

// waitForDataReady waits for data to go to low which means chip is ready
func (hx711 *Hx711) waitForDataReady() error {
	err := hx711.clockPin.Out(gpio.Low)
	if err != nil {
		return fmt.Errorf("set clock pin to low error: %v", err)
	}

	var level gpio.Level

	// looks like chip often takes 80 to 100 milliseconds to get ready
	// but somettimes it takes around 500 milliseconds to get ready
	// WaitForEdge sometimes returns right away
	// So will loop for 11, which could be more than 1 second, but usually 500 milliseconds
	for i := 0; i < 11; i++ {
		level = hx711.dataPin.Read()
		if level == gpio.Low {
			return nil
		}
		hx711.dataPin.WaitForEdge(100 * time.Millisecond)
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

	var level gpio.Level
	var data int
	for i := 0; i < 24; i++ {
		err = hx711.setClockHighThenLow()
		if err != nil {
			return 0, fmt.Errorf("setClockHighThenLow error: %v", err)
		}

		level = hx711.dataPin.Read()
		data = data << 1
		if level == gpio.High {
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
