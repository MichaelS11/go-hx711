// +build windows

package hx711

import (
	"time"
)

// HostInit calls periph.io host.Init(). This needs to be done before Hx711 can be used.
func HostInit() error {
	return nil
}

// NewHx711 creates new Hx711.
// Make sure to set clockPinName and dataPinName to the correct pins.
// https://cdn.sparkfun.com/datasheets/Sensors/ForceFlex/hx711_english.pdf
func NewHx711(clockPinName string, dataPinName string) (*Hx711, error) {
	return &Hx711{}, nil
}

// SetGain can be set to gain of 128, 64, or 32.
// Gain of 128 or 64 is input channel A, gain of 32 is input channel B.
// Default gain is 128.
// Note change only takes affect after one reading.
func (hx711 *Hx711) SetGain(gain int) {
}

// setClockHighThenLow sets clock pin high then low
func (hx711 *Hx711) setClockHighThenLow() error {
	return nil
}

// Reset starts up or resets the chip.
// The chip needs to be reset if it is not used for just about any amount of time.
func (hx711 *Hx711) Reset() error {
	return nil
}

// Shutdown puts the chip in powered down mode.
// The chip should be shutdown if it is not used for just about any amount of time.
func (hx711 *Hx711) Shutdown() error {
	return nil
}

// waitForDataReady waits for data to go to low which means chip is ready
func (hx711 *Hx711) waitForDataReady() error {
	return nil
}

// ReadDataRaw will get one raw reading from chip.
// Usually will need to call Reset before calling this and Shutdown after.
func (hx711 *Hx711) ReadDataRaw() (int, error) {
	return 0, nil
}

// readDataMedianRaw will get median of numReadings raw readings.
func (hx711 *Hx711) readDataMedianRaw(numReadings int, stop *bool) (int, error) {
	return 0, nil
}

// ReadDataMedianRaw will get median of numReadings raw readings.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianRaw(numReadings int) (int, error) {
	return 0, nil
}

// ReadDataMedian will get median of numReadings raw readings,
// then will adjust number with AdjustZero and AdjustScale.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedian(numReadings int) (float64, error) {
	return 0, nil
}

// ReadDataMedianThenAvg will get median of numReadings raw readings,
// then do that numAvgs number of time, and average those.
// then will adjust number with AdjustZero and AdjustScale.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianThenAvg(numReadings, numAvgs int) (float64, error) {
	return 0, nil
}

// ReadDataMedianThenMovingAvgs will get median of numReadings raw readings,
// then will adjust number with AdjustZero and AdjustScale. Stores data into previousReadings.
// Then returns moving average.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) ReadDataMedianThenMovingAvgs(numReadings, numAvgs int, previousReadings *[]float64) (float64, error) {
	if len(*previousReadings) < numAvgs {
		*previousReadings = append(*previousReadings, 0)
	} else {
		*previousReadings = append((*previousReadings)[1:numAvgs], 0)
	}
	return 0, nil
}

// BackgroundReadMovingAvgs it means to run in the background, run as a Goroutine.
// Will continue to get readings and update movingAvg until stop is set to true.
// After it has been stopped, the stopped chan will be closed.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) BackgroundReadMovingAvgs(numReadings, numAvgs int, movingAvg *float64, stop *bool, stopped chan struct{}) {
	*movingAvg = 0
	for !*stop {
		time.Sleep(200 * time.Millisecond)
	}
	close(stopped)
}

// GetAdjustValues will help get you the adjust values to plug in later.
// Do not call Reset before or Shutdown after.
// Reset and Shutdown are called for you.
func (hx711 *Hx711) GetAdjustValues(weight1 float64, weight2 float64) {
}
