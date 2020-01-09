// +build !windows

package hx711

import (
	"fmt"
	"log"
	"sort"
	"time"
)

var ErrTimeout = fmt.Errorf("timeout")

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
