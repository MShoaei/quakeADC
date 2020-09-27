package driver

import (
	"fmt"

	"gobot.io/x/gobot/drivers/spi"
	"periph.io/x/periph/conn/gpio"
)

// ADC7768-4 Register Addresses
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
)

// ADC7768-4 Register Addresses
const (
	GPIOControl uint8 = 0x0E + iota
	GPIOWriteData
	GPIOReadData
	PrechargeBuffer1
	PrechargeBuffer2
	PositiveReferencePrechargeBuffer
	NegativeReferencePrechargeBuffer
)

// ADC7768-4 Register Addresses
const (
	Ch0OffsetMSB uint8 = 0x1E + iota
	Ch0OffsetMid
	Ch0OffsetLSB
	Ch1OffsetMSB
	Ch1OffsetMid
	Ch1OffsetLSB
)

// ADC7768-4 Register Addresses
const (
	Ch2OffsetMSB uint8 = 0x2A + iota
	Ch2OffsetMid
	Ch2OffsetLSB
	Ch3OffsetMSB
	Ch3OffsetMid
	Ch3OffsetLSB
)

// ADC7768-4 Register Addresses
const (
	Ch0GainMSB uint8 = 0x36 + iota
	Ch0GainMid
	Ch0GainLSB
	Ch1GainMSB
	Ch1GainMid
	Ch1GainLSB
)

// ADC7768-4 Register Addresses
const (
	Ch2GainMSB uint8 = 0x42 + iota
	Ch2GainMid
	Ch2GainLSB
	Ch3GainMSB
	Ch3GainMid
	Ch3GainLSB
)

// ADC7768-4 Register Addresses
const (
	Ch0SyncOffset uint8 = 0x4E + iota
	Ch1SyncOffset
	_
	_
	Ch2SyncOffset
	Ch3SyncOffset
	_
	_
	DiagnosticRX
	DiagnosticMuxControl
	ModulatorDelayControl
	ChopControl
)

// Adc77684 is an SPI connection to send commands and receive responses
type Adc77684 struct {
	connection *spi.SpiConnection
}

// GetSpiConnection creates a new connection to send commands on.
func GetSpiConnection(busNum, chipNum, mode, bits int, maxSpeed int64) (*Adc77684, error) {
	c, err := spi.GetSpiConnection(busNum, chipNum, mode, bits, maxSpeed)
	if err != nil {
		return nil, err
	}

	return &Adc77684{connection: c.(*spi.SpiConnection)}, nil
}

// Transmit is used to send a new command and receive last commands response
func (adc *Adc77684) Transmit(tx, rx []byte, cs uint8) error {
	if cs < 1 || cs > 9 {
		return fmt.Errorf("invalid chip select %d", cs)
	}
	chipSelectPins[cs].FastOut(gpio.Low)
	err := adc.connection.Tx(tx, rx)
	chipSelectPins[cs].FastOut(gpio.High)

	return err
}

// Write sends command and ignores previous commands response
func (adc *Adc77684) Write(tx []byte, cs uint8) error {
	if cs < 1 || cs > 9 {
		return fmt.Errorf("invalid chip select %d", cs)
	}
	chipSelectPins[cs].FastOut(gpio.Low)
	err := adc.connection.Tx(tx, nil)
	chipSelectPins[cs].FastOut(gpio.High)

	return err
}

// Read reads the response of previous command
func (adc *Adc77684) Read(rx []byte, cs uint8) error {
	if cs < 1 || cs > 9 {
		return fmt.Errorf("invalid chip select %d", cs)
	}
	chipSelectPins[cs].FastOut(gpio.Low)
	err := adc.connection.Tx(nil, rx)
	chipSelectPins[cs].FastOut(gpio.High)

	return err
}

// Close closes the connection and frees the resources
func (adc *Adc77684) Close() error {
	return adc.connection.Close()
}
