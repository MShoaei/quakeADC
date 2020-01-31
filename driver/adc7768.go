package driver

import (
	"gobot.io/x/gobot/drivers/spi"
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

type Adc77684 struct {
	connection *spi.SpiConnection
}

func GetSpiConnection(busNum, chipNum, mode, bits int, maxSpeed int64) (*Adc77684, error) {
	c, err := spi.GetSpiConnection(busNum, chipNum, mode, bits, maxSpeed)
	if err != nil {
		return nil, err
	}

	return &Adc77684{connection: c.(*spi.SpiConnection)}, nil
}

func (adc *Adc77684) Transmit(tx, rx []byte) error {
	return adc.connection.Tx(tx, rx)
}

func (adc *Adc77684) Write(tx []byte) error {
	return adc.connection.Tx(tx, nil)
}

func (adc *Adc77684) Read(rx []byte) error {
	return adc.connection.Tx(nil, rx)
}

func (adc *Adc77684) Close() error {
	return adc.connection.Close()
}
