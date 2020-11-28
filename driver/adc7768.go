package driver

import (
	"fmt"

	"gobot.io/x/gobot/drivers/spi"
	"periph.io/x/periph/conn/gpio"
)

// ADC7768 Register Addresses
const (
	ChannelStandby uint8 = 0x00 + iota
	ChannelModeA
	ChannelModeB
	ChannelModeSelect
	PowerMode
	GeneralConfiguration
	DataControl
	InterfaceConfiguration
	BISTControl
	DeviceStatus
	RevisionID
	_ // reserved
	_ // reserved
	_ // reserved
	GPIOControl
	GPIOWriteData
	GPIOReadData
	PrechargeBuffer1
	PrechargeBuffer2
	PositiveReferencePrechargeBuffer
	NegativeReferencePrechargeBuffer
)

const (
	Ch0OffsetMSB uint8 = 0x1e + iota
	Ch0OffsetMid
	Ch0OffsetLSB
	Ch1OffsetMSB
	Ch1OffsetMid
	Ch1OffsetLSB
	Ch2OffsetMSB
	Ch2OffsetMid
	Ch2OffsetLSB
	Ch3OffsetMSB
	Ch3OffsetMid
	Ch3OffsetLSB
	Ch4OffsetMSB
	Ch4OffsetMid
	Ch4OffsetLSB
	Ch5OffsetMSB
	Ch5OffsetMid
	Ch5OffsetLSB
	Ch6OffsetMSB
	Ch6OffsetMid
	Ch6OffsetLSB
	Ch7OffsetMSB
	Ch7OffsetMid
	Ch7OffsetLSB
	Ch0GainMSB
	Ch0GainMid
	Ch0GainLSB
	Ch1GainMSB
	Ch1GainMid
	Ch1GainLSB
	Ch2GainMSB
	Ch2GainMid
	Ch2GainLSB
	Ch3GainMSB
	Ch3GainMid
	Ch3GainLSB
	Ch4GainMSB
	Ch4GainMid
	Ch4GainLSB
	Ch5GainMSB
	Ch5GainMid
	Ch5GainLSB
	Ch6GainMSB
	Ch6GainMid
	Ch6GainLSB
	Ch7GainMSB
	Ch7GainMid
	Ch7GainLSB
	Ch0SyncOffset
	Ch1SyncOffset
	Ch2SyncOffset
	Ch3SyncOffset
	Ch4SyncOffset
	Ch5SyncOffset
	Ch6SyncOffset
	Ch7SyncOffset
	DiagnosticRX
	DiagnosticMuxControl
	ModulatorDelayControl
	ChopControl
)

// Adc7768 is an SPI connection to send commands and receive responses
type Adc7768 struct {
	connection *spi.SpiConnection
}

// GetSpiConnection creates a new connection to send commands on.
func GetSpiConnection(busNum, chipNum, mode, bits int, maxSpeed int64) (*Adc7768, error) {
	c, err := spi.GetSpiConnection(busNum, chipNum, mode, bits, maxSpeed)
	if err != nil {
		return nil, err
	}

	return &Adc7768{connection: c.(*spi.SpiConnection)}, nil
}

func (adc Adc7768) Connection() spi.Connection {
	return adc.connection
}

// Transmit is used to send a new command and receive last commands response
func (adc *Adc7768) Transmit(tx, rx []byte, cs uint8) error {
	if cs < 1 || cs > 9 {
		return fmt.Errorf("invalid chip select %d", cs)
	}
	chipSelectPins[cs].FastOut(gpio.Low)
	err := adc.connection.Tx(tx, rx)
	chipSelectPins[cs].FastOut(gpio.High)

	return err
}

// Write sends command and ignores previous commands response
func (adc *Adc7768) Write(tx []byte, cs uint8) error {
	if cs < 1 || cs > 9 {
		return fmt.Errorf("invalid chip select %d", cs)
	}
	chipSelectPins[cs].FastOut(gpio.Low)
	err := adc.connection.Tx(tx, nil)
	chipSelectPins[cs].FastOut(gpio.High)

	return err
}

// Read reads the response of previous command
func (adc *Adc7768) Read(rx []byte, cs uint8) error {
	if cs < 1 || cs > 9 {
		return fmt.Errorf("invalid chip select %d", cs)
	}
	chipSelectPins[cs].FastOut(gpio.Low)
	err := adc.connection.Tx([]byte{0x8a, 0x00}, rx)
	chipSelectPins[cs].FastOut(gpio.High)

	return err
}

// Close closes the connection and frees the resources
func (adc *Adc7768) Close() error {
	return adc.connection.Close()
}
