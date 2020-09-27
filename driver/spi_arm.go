package driver

import (
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/bcm283x"
)

func init() {
	if _, err := host.Init(); err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	chipSelectPins = []*bcm283x.Pin{
		bcm283x.GPIO22, // Pin 15 for XMega
		bcm283x.GPIO21, // Pin 40
		bcm283x.GPIO20, // Pin 38
		bcm283x.GPIO16, // Pin 36
		bcm283x.GPIO12, // Pin 32
		bcm283x.GPIO1,  // Pin 28
		bcm283x.GPIO2,  // Pin 3
		bcm283x.GPIO3,  // Pin 5
		bcm283x.GPIO4,  // Pin 7
		bcm283x.GPIO17, // Pin 11
	}

	for _, c := range chipSelectPins {
		if err := c.SetFunc(gpio.OUT_HIGH); err != nil {
			log.Fatal(err)
		}
	}

	err := bcm283x.GPIO7.SetFunc(gpio.OUT_HIGH) // Pin 26 for ADC Sync
	if err != nil {
		log.Fatalln(err)
	}
}
