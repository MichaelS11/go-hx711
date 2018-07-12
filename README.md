# Go HX711 interface

Golang HX711 interface

[![GoDoc Reference](https://godoc.org/github.com/MichaelS11/go-hx711?status.svg)](http://godoc.org/github.com/MichaelS11/go-hx711)
[![Go Report Card](https://goreportcard.com/badge/github.com/MichaelS11/go-hx711)](https://goreportcard.com/report/github.com/MichaelS11/go-hx711)

## Please note

Please make sure to setup your HX711 correctly. Do a search on the internet to find guide. Here is an example of a guide:

https://learn.sparkfun.com/tutorials/load-cell-amplifier-hx711-breakout-hookup-guide

The examples below are from using a Raspberry Pi 3 with GPIO 6 for clock and GPIO 5 for data. Your setup may be different, if so, your pin names would need to change in each example.

If you need to read from channel B, make sure to call hx711.SetGain(32)

Side note, in my testing using 3V input had better consistency then using a 5V input.


## Get

go get github.com/MichaelS11/go-cql-driver


## Simple test to make sure scale is working

Run the following program to test your scale. Add and remove weight. Make sure there are no errors. Also make sure that the values go up when you add weight and go down when you remove weight. Don't worry about if the values match the weight, just that they go up and down in value at the correct time.

```go
package main

import (
	"fmt"
	"time"

	"github.com/MichaelS11/go-hx711"
)

func main() {
	defer hx711.HwioCloseAll()

	hx711, err := hx711.NewHx711("gpio6", "gpio5")
	if err != nil {
		fmt.Println("NewHx711 error:", err)
		return
	}

	defer hx711.Shutdown()

	err = hx711.Reset()
	if err != nil {
		fmt.Println("Reset error:", err)
		return
	}

	var data int
	for i := 0; i < 10000; i++ {
		time.Sleep(200 * time.Microsecond)

		data, err = hx711.ReadDataRaw()
		if err != nil {
			fmt.Println("ReadDataRaw error:", err)
			continue
		}

		fmt.Println(data)
	}

}
```


## Calibrate the readings / get AdjustZero & AdjustScale values

To get the values needed to calibrate the scale's readings will need at least one weight of known value. Having two weights is preferred. In the below program change weight1 and weight2 to the known weight values. Make sure to set it in the unit of measurement that you prefer (pounds, ounces, kg, g, etc.). To start, make sure there are no weight on the scale. Run the program. When asked, put the first weight on the scale. Then when asked, put the second weight on the scale. It will print out the AdjustZero and AdjustScale values. Those are using in the next example.

Please note that temperature affects the readings. This means if you are planning on reading the weight often, should do that for about 45 minutes before calibration.

```go
package main

import (
	"fmt"

	"github.com/MichaelS11/go-hx711"
)

func main() {
	defer hx711.HwioCloseAll()

	hx711, err := hx711.NewHx711("gpio6", "gpio5")
	if err != nil {
		fmt.Println("NewHx711 error:", err)
		return
	}
  
	// SetGain default is 128
	// Gain of 128 or 64 is input channel A, gain of 32 is input channel B
	// hx711.SetGain(128)

	var weight1 float64
	var weight2 float64

	weight1 = 100
	weight2 = 200

	hx711.GetAdjustValues(weight1, weight2)
}
```


## Simple program to get weight

Take the AdjustZero and AdjustScale values from the above program and plug them into the below program. Run program. Put weight on the scale and check the values. The AdjustZero and AdjustScale may need to be adjusted to your liking.

```go
package main

import (
	"fmt"
	"time"

	"github.com/MichaelS11/go-hx711"
)

func main() {
	defer hx711.HwioCloseAll()

	hx711, err := hx711.NewHx711("gpio6", "gpio5")
	if err != nil {
		fmt.Println("NewHx711 error:", err)
		return
	}

	// SetGain default is 128
	// Gain of 128 or 64 is input channel A, gain of 32 is input channel B
	// hx711.SetGain(128)

	// make sure to use your values from calibration above
	hx711.AdjustZero = -123
	hx711.AdjustScale = 456

	var data float64
	for i := 0; i < 10000; i++ {
		time.Sleep(200 * time.Microsecond)

		data, err = hx711.ReadDataMedian()
		if err != nil {
			fmt.Println("ReadDataMedian error:", err)
			continue
		}

		fmt.Println(data)
	}
}
```
