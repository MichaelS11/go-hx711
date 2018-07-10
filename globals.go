// +build !windows

package hx711

import (
	// waiting for pull request to switch to:
	// "github.com/mrmorphic/hwio"
	"github.com/MichaelS11/hwio"
)

// Hx711 struct to interface with the hx711 chip
// Call NewHx711 to create a new one
type Hx711 struct {
	clockPin     hwio.Pin
	dataPin      hwio.Pin
	numEndPulses int
	// NumReadings sets the number of readings to get for median reads.
	// Defaults to 21.
	// Do not set below 1.
	NumReadings int
	// AdjustZero should be set to an int that will zero out a raw reading
	AdjustZero int
	// AdjustScale should be set to a float64 that will give output units wanted
	AdjustScale float64
}
