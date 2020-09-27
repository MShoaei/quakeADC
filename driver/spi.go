package driver

import (
	"fmt"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/bcm283x"
)

var chipSelectPins []*bcm283x.Pin

func EnableChipSelect(chip uint8) error {
	if chip < 0 || chip > 9 {
		return fmt.Errorf("invalid chip value %d", chip)
	}
	chipSelectPins[chip].FastOut(gpio.Low)
	return nil
}

func DisableChipSelect(chip uint8) error {
	if chip < 0 || chip > 9 {
		return fmt.Errorf("invalid chip value %d", chip)
	}
	chipSelectPins[chip].FastOut(gpio.High)
	return nil
}
