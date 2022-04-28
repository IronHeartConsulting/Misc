package main

import (
	"fmt"
	"log"

	"github.com/lukechannings/max31855"

	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	s, err := spireg.Open("")

	if err != nil {
		log.Fatal("Couldn't open SPI port! " + err.Error())
	}

	dev, err := max31855.New(s)

	if err != nil {
		log.Fatal("Couldn't open device! " + err.Error())
	}

	temp, err := dev.GetTemp()

  if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Current temperature: %v °C (%v °F)", temp.Thermocouple.Celsius(), temp.Thermocouple.Fahrenheit())
}
