// +build !windows,rpio

package hx711

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

var timeoutError = fmt.Errorf("timeout")

// HostInit opens /dev/gpiomem. This needs to be done before Hx711 can be used.
func HostInit() error {
	return rpio.Open()
}

// NewHx711 creates new Hx711.
// Make sure to set clockPin and dataPin to the correct pins.
// https://cdn.sparkfun.com/datasheets/Sensors/ForceFlex/hx711_english.pdf
func NewHx711(clockPin int, dataPin int) (*Hx711, error) {
	hx711 := &Hx711{numEndPulses: 1}
	hx711.clockPin = rpio.Pin(clockPin)
	hx711.dataPin = rpio.Pin(dataPin)
	hx711.dataPin.Input()
	hx711.clockPin.Output()
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

	return timeoutError
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

// readDataMedianRaw will get median of numReadings raw readings.
func (hx711 *Hx711) readDataMedianRaw(numReadings int, stop *bool) (int, error) {
	var err error
	var data int
	datas := make([]int, 0, numReadings)

	for i := 0; i < numReadings; i++ {
		if *stop {
			return 0, fmt.Errorf("stopped")
		}

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

	if len(datas) < 1 {
		return 0, fmt.Errorf("no data, last err: %v", err)
	}

	sort.Ints(datas)

	return datas[len(datas)/2], nil
}

// ReadDataMedianRaw will get median of numReadings raw readings.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianRaw(numReadings int) (int, error) {
	var data int

	err := hx711.Reset()
	if err != nil {
		return 0, fmt.Errorf("Reset error: %v", err)
	}

	stop := false
	data, err = hx711.readDataMedianRaw(numReadings, &stop)

	hx711.Shutdown()

	return data, err
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
// then will adjust number with AdjustZero and AdjustScale. Stores data into previousReadings.
// Then returns moving average.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
// Will panic if previousReadings is nil
func (hx711 *Hx711) ReadDataMedianThenMovingAvgs(numReadings, numAvgs int, previousReadings *[]float64) (float64, error) {
	data, err := hx711.ReadDataMedian(numReadings)
	if err != nil {
		return 0, err
	}

	if len(*previousReadings) < numAvgs {
		*previousReadings = append(*previousReadings, data)
	} else {
		*previousReadings = append((*previousReadings)[1:numAvgs], data)
	}

	var result float64
	for i := range *previousReadings {
		result += (*previousReadings)[i]
	}
	return result / float64(len(*previousReadings)), nil
}

// BackgroundReadMovingAvgs it meant to be run in the background, run as a Goroutine.
// Will continue to get readings and update movingAvg until stop is set to true.
// After it has been stopped, the stopped chan will be closed.
// Note when scale errors the movingAvg value will not change.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
// Will panic if movingAvg or stop are nil
func (hx711 *Hx711) BackgroundReadMovingAvgs(numReadings, numAvgs int, movingAvg *float64, stop *bool, stopped chan struct{}) {
	var err error
	var data int
	var result float64
	previousReadings := make([]float64, 0, numAvgs)

	for {
		err = hx711.Reset()
		if err == nil {
			break
		}
		log.Print("hx711 BackgroundReadMovingAvgs Reset error:", err)
		time.Sleep(time.Second)
	}

	for !*stop {
		data, err = hx711.readDataMedianRaw(numReadings, stop)
		if err != nil && err.Error() != "stopped" {
			log.Print("hx711 BackgroundReadMovingAvgs ReadDataMedian error:", err)
			continue
		}

		result = float64(data-hx711.AdjustZero) / hx711.AdjustScale
		if len(previousReadings) < numAvgs {
			previousReadings = append(previousReadings, result)
		} else {
			previousReadings = append(previousReadings[1:numAvgs], result)
		}

		result = 0
		for i := range previousReadings {
			result += previousReadings[i]
		}

		*movingAvg = result / float64(len(previousReadings))
	}

	hx711.Shutdown()

	close(stopped)
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
