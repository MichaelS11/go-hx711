// +build !windows

package hx711

import (
	"fmt"
	"sort"
	"time"

	// waiting for pull request to switch to:
	// "github.com/mrmorphic/hwio"
	"github.com/MichaelS11/hwio"
)

// HwioCloseAll closes all opened hwio
func HwioCloseAll() {
	hwio.CloseAll()
}

// NewHx711 creates new Hx711.
// Make sure to set clockPinName and dataPinName to the correct pins.
// https://cdn.sparkfun.com/datasheets/Sensors/ForceFlex/hx711_english.pdf
func NewHx711(clockPinName string, dataPinName string) (*Hx711, error) {
	var err error
	hx711 := &Hx711{numEndPulses: 1}

	hx711.clockPin, err = hwio.GetPinWithMode(clockPinName, hwio.OUTPUT)
	if err != nil {
		return nil, fmt.Errorf("clockPin GetPinWithMode error: %v", err)
	}

	hx711.dataPin, err = hwio.GetPinWithMode(dataPinName, hwio.INPUT)
	if err != nil {
		return nil, fmt.Errorf("dataPin GetPinWithMode error: %v", err)
	}

	return hx711, nil
}

// SetGain can be set to gain of 128, 64, or 32.
// Gain of 128 or 64 is input channel A, gain of 32 is input channel B.
// Default gain is 128.
// Note change only takes affect after one reading.
func (hx711 *Hx711) SetGain(gain int) {
	switch gain {
	case 128:
		hx711.numEndPulses = 1
	case 64:
		hx711.numEndPulses = 3
	case 32:
		hx711.numEndPulses = 2
	default:
		hx711.numEndPulses = 1
	}
}

func (hx711 *Hx711) setClockHigh() error {
	return hwio.DigitalWrite(hx711.clockPin, hwio.HIGH)
}

func (hx711 *Hx711) setClockLow() error {
	return hwio.DigitalWrite(hx711.clockPin, hwio.LOW)
}

func (hx711 *Hx711) readDataBit() (int, error) {
	return hwio.DigitalRead(hx711.dataPin)
}

// Reset starts up or resets the chip.
// The chip needs to be reset if it is not used for just about any amount of time.
func (hx711 *Hx711) Reset() error {
	err := hx711.setClockLow()
	if err != nil {
		return fmt.Errorf("setClockLow error: %v", err)
	}
	err = hx711.setClockHigh()
	if err != nil {
		return fmt.Errorf("setClockHigh error: %v", err)
	}
	time.Sleep(70 * time.Microsecond)
	err = hx711.setClockLow()
	if err != nil {
		return fmt.Errorf("setClockLow error: %v", err)
	}
	return nil
}

// Shutdown puts the chip in powered down mode.
// The chip should be shutdown if it is not used for just about any amount of time.
func (hx711 *Hx711) Shutdown() error {
	err := hx711.setClockHigh()
	if err != nil {
		return fmt.Errorf("setClockHigh error: %v", err)
	}
	return nil
}

func (hx711 *Hx711) waitForDataReady() error {
	err := hx711.setClockLow()
	if err != nil {
		return fmt.Errorf("setClockLow error: %v", err)
	}

	// wait for at least a second for the chip to be ready
	// 50 * 20 * millisecond = 1 second
	// looks like chip often takes 80 to 100 milliseconds to get ready
	// but somettimes it takes around 500 milliseconds to get ready
	for i := 0; i < 100; i++ {
		data, err := hx711.readDataBit()
		if err != nil {
			return fmt.Errorf("readDataBit error: %v", err)
		}
		if data == hwio.LOW {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("timeout")
}

// ReadDataRaw will get one raw reading from chip.
// Usually will need to call Reset before calling this and Shutdown after.
func (hx711 *Hx711) ReadDataRaw() (int, error) {
	err := hx711.waitForDataReady()
	if err != nil {
		return 0, fmt.Errorf("waitForDataReady error: %v", err)
	}

	var bit int
	var data int
	for i := 0; i < 24; i++ {
		err = hx711.setClockHigh()
		if err != nil {
			return 0, fmt.Errorf("setClockHigh error: %v", err)
		}
		err = hx711.setClockLow()
		if err != nil {
			return 0, fmt.Errorf("setClockLow error: %v", err)
		}
		bit, err = hx711.readDataBit()
		if err != nil {
			return 0, fmt.Errorf("readDataBit error: %v", err)
		}
		data = data << 1
		if bit == 1 {
			data++
		}
	}

	for i := 0; i < hx711.numEndPulses; i++ {
		err = hx711.setClockHigh()
		if err != nil {
			return 0, fmt.Errorf("setClockHigh error: %v", err)
		}
		err = hx711.setClockLow()
		if err != nil {
			return 0, fmt.Errorf("setClockLow error: %v", err)
		}
	}

	// if high 23 bit is set, value is negtive
	if (data & 0x800000) > 0 {
		// flip bits higher than 23 to get negtive number for int
		data |= ^0xffffff
	}

	return data, nil
}

// ReadDataMedianRaw will get median of numReadings raw readings.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianRaw(numReadings int) (int, error) {
	var err error
	var data int
	datas := make([]int, 0, numReadings)

	err = hx711.Reset()
	if err != nil {
		return 0, fmt.Errorf("Reset error: %v", err)
	}

	for i := 0; i < numReadings; i++ {
		data, err = hx711.ReadDataRaw()
		if err != nil {
			continue
		}
		// reading of -1 seems to be some kind of error
		if data == -1 {
			continue

		}
		datas = append(datas, data)
	}

	hx711.Shutdown()

	if len(datas) < 1 {
		return 0, fmt.Errorf("no data, last err: %v", err)
	}

	sort.Ints(datas)

	return datas[len(datas)/2], nil
}

// ReadDataMedian will get median of numReadings raw readings,
// then will adjust number with AdjustZero and AdjustScale.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedian(numReadings int) (float64, error) {
	data, err := hx711.ReadDataMedianRaw(numReadings)
	if err != nil {
		return 0, err
	}
	return float64(data-hx711.AdjustZero) / hx711.AdjustScale, nil
}

// ReadDataMedianThenAvg will get median of numReadings raw readings,
// then do that numAvgs number of time, and average those.
// then will adjust number with AdjustZero and AdjustScale.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianThenAvg(numReadings, numAvgs int) (float64, error) {
	var sum int
	for i := 0; i < numAvgs; i++ {
		data, err := hx711.ReadDataMedianRaw(numReadings)
		if err != nil {
			return 0, err
		}
		sum += data - hx711.AdjustZero
	}
	return (float64(sum) / float64(numAvgs)) / hx711.AdjustScale, nil
}

// ReadDataMedianThenMovingAvgs will get median of numReadings raw readings,
// then will adjust number with AdjustZero and AdjustScale
// then get the moving avgs.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianThenMovingAvgs(numReadings, numAvgs int, movingAvgs *[]float64) error {
	data, err := hx711.ReadDataMedian(numReadings)
	if err != nil {
		return err
	}
	var result float64
	for i := range *movingAvgs {
		result += (*movingAvgs)[i]
	}
	result = (result + float64(data)) / float64(len(*movingAvgs)+1)
	if len(*movingAvgs) < numAvgs {
		*movingAvgs = append(*movingAvgs, result)
		return nil
	}
	*movingAvgs = append((*movingAvgs)[1:], result)
	return nil
}

// GetAdjustValues will help get you the adjust values to plug in later.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) GetAdjustValues(weight1 float64, weight2 float64) {
	var err error
	var adjustZero int
	var scale1 int
	var scale2 int

	fmt.Println("Make sure scale is working and empty, getting weight in 5 seconds...")
	time.Sleep(5 * time.Second)
	fmt.Println("Getting weight...")
	adjustZero, err = hx711.ReadDataMedianRaw(11)
	if err != nil {
		fmt.Println("ReadDataMedianRaw error:", err)
		return
	}
	fmt.Println("Raw weight is:", adjustZero)
	fmt.Println("")

	fmt.Printf("Put first weight of %.2f on scale, getting weight in 15 seconds...\n", weight1)
	time.Sleep(15 * time.Second)
	fmt.Println("Getting weight...")
	scale1, err = hx711.ReadDataMedianRaw(11)
	if err != nil {
		fmt.Println("ReadDataMedianRaw error:", err)
		return
	}
	fmt.Println("Raw weight is:", scale1)
	fmt.Println("")

	fmt.Printf("Put second weight of %.2f on scale, getting weight in 15 seconds...\n", weight2)
	time.Sleep(15 * time.Second)
	fmt.Println("Getting weight...")
	scale2, err = hx711.ReadDataMedianRaw(11)
	if err != nil {
		fmt.Println("ReadDataMedianRaw error:", err)
		return
	}
	fmt.Println("Raw weight is ", scale2)
	fmt.Println("")

	adjust1 := float64(scale1-adjustZero) / weight1
	adjust2 := float64(scale2-adjustZero) / weight2

	fmt.Println("AdjustZero should be set to:", adjustZero)
	fmt.Printf("AdjustScale should be set to a value between %f and %f\n", adjust1, adjust2)
	fmt.Println("")
}
