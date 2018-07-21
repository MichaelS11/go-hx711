// +build windows

package hx711

// Hx711 struct to interface with the hx711 chip.
// Call NewHx711 to create a new one.
type Hx711 struct {
	// AdjustZero should be set to an int that will zero out a raw reading
	AdjustZero int
	// AdjustScale should be set to a float64 that will give output units wanted
	AdjustScale float64
}
