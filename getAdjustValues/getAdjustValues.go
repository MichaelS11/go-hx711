package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/MichaelS11/go-hx711"
)

func main() {
	err := hx711.HostInit()
	if err != nil {
		fmt.Println("Host Init error:", err)
		return
	}

	fmt.Println("Enter clock pin name (press enter for GPIO6):")
	clockPinName := getInput()
	if clockPinName == "" {
		clockPinName = "GPIO6"
	}

	fmt.Println("Enter data pin name (press enter for GPIO5):")
	dataPinName := getInput()
	if dataPinName == "" {
		dataPinName = "GPIO5"
	}

	fmt.Println("Enter weight1:")
	input := getInput()
	if input == "" {
		fmt.Println("input empty")
		return
	}
	weight1, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("ParseFloat error:", err)
		return
	}

	fmt.Println("Enter weight2:")
	input = getInput()
	if input == "" {
		fmt.Println("input empty")
		return
	}
	weight2, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("ParseFloat error:", err)
		return
	}

	hx711, err := hx711.NewHx711(clockPinName, dataPinName)
	if err != nil {
		fmt.Println("NewHx711 error:", err)
		return
	}

	fmt.Printf("\nclockPinName: %v - dataPinName: %v - weight1: %f, weight2: %f\n\n", clockPinName, dataPinName, weight1, weight2)

	hx711.GetAdjustValues(weight1, weight2)
}

func getInput() string {
	buffer := make([]byte, 512)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(buffer, 512)

	if !scanner.Scan() {
		return ""
	}
	if scanner.Err() != nil {
		return ""
	}

	return scanner.Text()
}
