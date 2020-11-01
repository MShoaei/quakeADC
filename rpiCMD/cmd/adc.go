package cmd

import (
	"log"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/bcm283x"
)

var chipSelect uint8

// adcCmd represents the adc command
var adcCmd = &cobra.Command{
	Use:   "adc",
	Short: "command to control the ADC over SPI",
}

func newAdcChStandbyCommand() *cobra.Command {
	options := driver.ChStandbyOpts{}
	cmd := &cobra.Command{
		Use:   "ChStandby",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tx, rx, err := adcConnection.ChStandby(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	// true==standby
	// false==enabled
	f.BoolVar(&options.Channels[0], "ch0", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[1], "ch1", true, "channels 1 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[2], "ch2", true, "channels 2 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[3], "ch3", true, "channels 3 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[4], "ch4", true, "channels 4 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[5], "ch5", true, "channels 5 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[6], "ch6", true, "channels 6 standby mode. true/t: Standby, false/f: Enabled")
	f.BoolVar(&options.Channels[7], "ch7", true, "channels 7 standby mode. true/t: Standby, false/f: Enabled")
	return cmd
}

func newAdcChModeACommand() *cobra.Command {
	options := driver.ChModeOpts{}
	cmd := &cobra.Command{Use: "ChModeA",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tx, rx, err := adcConnection.ChModeA(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.FType, "f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	f.Uint16Var(&options.DecRate, "dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")
	//flagsList["ChModeA"] = f

	return cmd
}

func newAdcChModeBCommand() *cobra.Command {
	options := driver.ChModeOpts{}
	cmd := &cobra.Command{Use: "ChModeB",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.ChModeB(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.FType, "f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	f.Uint16Var(&options.DecRate, "dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")
	//flagsList["ChModeB"] = f

	return cmd
}

func newAdcChModeSelectCommand() *cobra.Command {
	options := driver.ChModeSelectOpts{}
	cmd := &cobra.Command{
		Use:   "ChModeSel",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.ChModeSel(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channels[0], "ch0", 0, "set channel mode for channel 0. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[1], "ch1", 0, "set channel mode for channel 1. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[2], "ch2", 0, "set channel mode for channel 2. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[3], "ch3", 0, "set channel mode for channel 3. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[4], "ch4", 0, "set channel mode for channel 4. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[5], "ch5", 0, "set channel mode for channel 5. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[6], "ch6", 0, "set channel mode for channel 6. 0:Mode A, 1:Mode B")
	f.Uint8Var(&options.Channels[7], "ch7", 0, "set channel mode for channel 7. 0:Mode A, 1:Mode B")
	//flagsList["ChModeSel"] = f

	return cmd
}

func newAdcPowerModeCommand() *cobra.Command {
	options := driver.PowerModeOpts{}
	cmd := &cobra.Command{
		Use:   "PowerMode",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.PowerMode(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Sleep, "sleep", 0, "0: Normal operation, 1: Sleep mode")
	f.Uint8Var(&options.Power, "power", 0, "0: Low power, 2: Median, 3: Fast")
	f.Uint8Var(&options.LVDSClock, "lvds-clk", 0, "0: disable LVDS clock, 1: enable LVDS clock")
	f.Uint8Var(&options.MCLKDiv, "mclk-div", 0, "0: set to MCLK/32 for low power mode, 2: set to MCLK/8 for median mode, 3: set to MCLK/4 for fast mode")
	f.SortFlags = false
	flagsList["PowerMode"] = f

	return cmd
}
func newAdcGeneralConfigurationCommand() *cobra.Command {
	options := driver.GeneralConfOpts{}
	cmd := &cobra.Command{
		Use:   "GeneralConf",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.GeneralConf(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.RETimeEnable, "retime-en", 0, "SYNC_OUT signal retime enable bit. 0: disabled, 1: enabled")
	f.Uint8Var(&options.VcmPd, "vcm-pd", 0, "VCM buffer power-down. 0: enabled, 1: disabled")
	f.Uint8Var(&options.VcmVSelect, "vcm-vsel", 0, "VCM voltage select bits. 0: (AVDD1 − AVSS)/2 V, 1: 1.65 V, 2: 2.5 V, 3: 2.14 V")
	flagsList["GeneralConf"] = f

	return cmd
}
func newAdcDataControlCommand() *cobra.Command {
	options := driver.DataControlOpts{}
	cmd := &cobra.Command{
		Use:   "DataControl",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.DataControl(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.SpiSync, "spi-sync", 1, "Software synchronization of the AD7768-4. This command has the same effect as sending a signal pulse to the START pin. 0: Change to SPI_SYNC low, 1: Change to SPI_SYNC high")
	f.Uint8Var(&options.SingleShot, "single-shot", 0, "One-shot mode. Enables one-shot mode. In one-shot mode, the AD7768-4 output a conversion result in response to a SYNC_IN rising edge. 0: Disabled, 1: Enabled")
	f.Uint8Var(&options.SpiReset, "spi-reset", 0, "Soft reset. Two successive commands must be received in the correct order to generate a reset. 0: No effect, 1: No effect, 2: Second reset command, 3: First reset command")
	flagsList["DataControl"] = f

	return cmd
}
func newAdcInterfaceConfigurationCommand() *cobra.Command {
	options := driver.InterfaceConfOpts{}
	cmd := &cobra.Command{
		Use:   "InterfaceConf",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.InterfaceConf(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.CRCSelect, "crc-sel", 0, "CRC select. These bits allow the user to implement a CRC on the data interface. 0: No CRC. Status bits with every conversion, 1: Replace the header with CRC message every 4 samples, 2: Replace the header with CRC message every 16 samples, 3: Replace the header with CRC message every 16 samples")
	f.Uint8Var(&options.DclkDiv, "dclk-div", 0, "DCLK divider. These bits control division of the DCLK clock used to clock out conversion data on the DOUTx pins. 0: Divide by 8, 1: Divide by 4, 2: Divide by 2, 3: No division")
	flagsList["InterfaceConf"] = f

	return cmd
}
func newAdcBISTControlCommand() *cobra.Command {
	options := driver.BISTControlOpts{}
	cmd := &cobra.Command{
		Use:   "BISTControl",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.BISTControl(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.RamBISTStart, "ram-bist-start", 0, "RAM BIST. 0: Off, 1: Begin RAM BIST")
	flagsList["BISTControl"] = f

	return cmd
}
func newAdcDeviceStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "DeviceStatus",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.DeviceStatus(chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err

		},
	}
	return cmd
}

func newAdcRevisionIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "RevisionID",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.RevisionID(chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err

		},
	}
	return cmd
}

func newAdcGPIOControlCommand() *cobra.Command {
	options := driver.GPIOControlOpts{}
	cmd := &cobra.Command{
		Use:   "GPIOControl",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.GPIOControl(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.UGpioEnable, "ugpio-en", 0, "User GPIO enable. 0: GPIO Disabled, 1: GPIO Enabled")
	f.Uint8Var(&options.GPIO[0], "gpio0", 0, "GPIO0 Direction. 0: Input, 1: Output")
	f.Uint8Var(&options.GPIO[1], "gpio1", 0, "GPIO0 Direction. 1: Input, 1: Output")
	f.Uint8Var(&options.GPIO[2], "gpio2", 0, "GPIO0 Direction. 2: Input, 1: Output")
	f.Uint8Var(&options.GPIO[3], "gpio3", 0, "GPIO0 Direction. 3: Input, 1: Output")
	f.Uint8Var(&options.GPIO[4], "gpio4", 0, "GPIO0 Direction. 4: Input, 1: Output")

	return cmd
}

// TODO: Need better explanation for flags
func newAdcGPIOWriteDataCommand() *cobra.Command {
	options := driver.GPIOWriteDataOpts{}
	cmd := &cobra.Command{
		Use:   "GPIOWriteData",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.GPIOWriteData(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.GPIO[0], "gpio0", 0, "GPIO0/MODE0")
	f.Uint8Var(&options.GPIO[1], "gpio1", 0, "GPIO1/MODE1")
	f.Uint8Var(&options.GPIO[2], "gpio2", 0, "GPIO2/MODE2")
	f.Uint8Var(&options.GPIO[3], "gpio3", 0, "GPIO3/MODE3")
	f.Uint8Var(&options.GPIO[4], "gpio4", 0, "GPIO4/FILTER")
	//flagsList["GPIOWriteData"] = f

	return cmd
}
func newAdcGPIOReadDataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "GPIOReadData",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.GPIOReadData(chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	return cmd
}
func newAdcPrechargeBuffer1Command() *cobra.Command {
	options := driver.PreChargeBufferOpts{}
	cmd := &cobra.Command{
		Use:   "PrechargeBuffer1",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.PrechargeBuffer1(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false

	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.ChPositiveEnable[0], "ch0-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[0], "ch0-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChPositiveEnable[1], "ch1-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[1], "ch1-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChPositiveEnable[2], "ch2-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[2], "ch2-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChPositiveEnable[3], "ch3-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[3], "ch3-neg", 1, "0: Off, 1: On (default: 1)")

	return cmd
}
func newAdcPrechargeBuffer2Command() *cobra.Command {
	options := driver.PreChargeBufferOpts{}
	cmd := &cobra.Command{
		Use:   "PrechargeBuffer2",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.PrechargeBuffer2(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.ChPositiveEnable[0], "ch4-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[0], "ch4-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChPositiveEnable[1], "ch5-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[1], "ch5-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChPositiveEnable[2], "ch6-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[2], "ch6-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChPositiveEnable[3], "ch7-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8Var(&options.ChNegativeEnable[3], "ch7-neg", 1, "0: Off, 1: On (default: 1)")

	return cmd
}

func newAdcPositiveRefPrechargeBufCommand() *cobra.Command {
	options := driver.ReferencePrechargeBufOpts{}
	cmd := &cobra.Command{Use: "PositiveRefPrechargeBuf",
		Aliases: []string{"prpb"},
		Short:   "",
		Long:    "",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.PositiveRefPrechargeBuf(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channels[0], "ch0", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[1], "ch1", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[2], "ch2", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[3], "ch3", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[4], "ch4", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[5], "ch5", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[6], "ch6", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[7], "ch7", 1, "0: Off, 1: On (default: 0)")
	//flagsList["PositiveRefPrechargeBuf"] = f

	return cmd
}
func newAdcNegativeRefPrechargeBufCommand() *cobra.Command {
	options := driver.ReferencePrechargeBufOpts{}
	cmd := &cobra.Command{Use: "NegativeRefPrechargeBuf",
		Aliases: []string{"nrpb"},
		Short:   "",
		Long:    "",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.NegativeRefPrechargeBuf(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channels[0], "ch0", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[1], "ch1", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[2], "ch2", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[3], "ch3", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[4], "ch4", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[5], "ch5", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[6], "ch6", 1, "0: Off, 1: On (default: 0)")
	f.Uint8Var(&options.Channels[7], "ch7", 1, "0: Off, 1: On (default: 0)")
	flagsList["NegativeRefPrechargeBuf"] = f

	return cmd
}
func newAdcChannelOffsetCommand() *cobra.Command {
	options := driver.ChannelOffsetOpts{}
	cmd := &cobra.Command{
		Use:   "ChOffset",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return adcConnection.ChannelOffset(options, chipSelect)
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channel, "ch", 255, "select channel to set offset: ch: [0..7]")
	f.Int8Var(&options.Offset[0], "MSB", 0, "Channel 'ch' offset MSB signed 8 bit integer (default: 0)")
	f.Int8Var(&options.Offset[1], "Mid", 0, "Channel 'ch' offset Mid signed 8 bit integer (default: 0)")
	f.Int8Var(&options.Offset[2], "LSB", 0, "Channel 'ch' offset LSB signed 8 bit integer (default: 0)")

	return cmd
}

func newAdcChGainCommandCommand() *cobra.Command {
	options := driver.ChannelGainOpts{}
	cmd := &cobra.Command{
		Use:   "ChGain",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return adcConnection.ChannelGain(options, chipSelect)
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channel, "ch", 255, "select channel to set offset: ch: [0..7]")
	_ = cmd.MarkFlagRequired("ch")
	f.Uint8Var(&options.Offset[0], "MSB", 0, "Channel 'ch' gain MSB unsigned 8 bit integer (default: 0)")
	f.Uint8Var(&options.Offset[1], "Mid", 0, "Channel 'ch' gain Mid unsigned 8 bit integer (default: 0)")
	f.Uint8Var(&options.Offset[2], "LSB", 0, "Channel 'ch' gain LSB unsigned 8 bit integer (default: 0)")

	return cmd
}

func newAdcChannelSyncOffsetCommand() *cobra.Command {
	options := driver.ChannelSyncOffsetOpts{}
	cmd := &cobra.Command{
		Use:   "ChSyncOffset",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return adcConnection.ChannelSyncOffset(options, chipSelect)
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channel, "ch", 255, "select channel to set offset: ch: [0..7]")
	_ = cmd.MarkFlagRequired("ch")
	f.Uint8Var(&options.Offset, "MSB", 0, "Channel 'ch' sync phase offset (default: 0)")

	return cmd
}

func newAdcDiagnosticRXCommand() *cobra.Command {
	options := driver.DiagnosticRXOpts{}
	cmd := &cobra.Command{
		Use:   "DiagnosticRX",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.DiagnosticRX(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.Channels[0], "ch0", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[1], "ch1", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[2], "ch2", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[3], "ch3", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[4], "ch4", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[5], "ch5", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[6], "ch6", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8Var(&options.Channels[7], "ch7", 0, "0: Not in use, 1: Receive (default: 0)")
	//flagsList["DiagnosticRX"] = f

	return cmd
}
func newAdcDiagnosticMuxControlCommand() *cobra.Command {
	options := driver.DiagnosticMuxControlOpts{}
	cmd := &cobra.Command{
		Use:   "DiagnosticMuxControl",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.DiagnosticMuxControl(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.GrpbSelect, "grpb-sel", 0, "0: Off, 3: Positive full-scale ADC check, 4: Negative full-scale ADC check, 5: Zero-scale ADC check")
	f.Uint8Var(&options.GrpaSelect, "grpa-sel", 0, "0: Off, 3: Positive full-scale ADC check, 4: Negative full-scale ADC check, 5: Zero-scale ADC check")
	//flagsList["DiagnosticMuxControl"] = f

	return cmd
}
func newAdcModulatorDelayControlCommand() *cobra.Command {
	options := driver.ModulatorDelayControlOpts{}
	cmd := &cobra.Command{
		Use:   "ModulatorDelayControl",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.ModulatorDelayControl(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.ModDelay, "mod-delay", 0, "0: Disabled delayed clock for all channels, 1: Enable delayed clock for Channel 0 and Channel 1, 2: Enable delayed clock for Channel 2 and Channel 3, 3: Enable delayed clock for all channels")
	//flagsList["ModulatorDelayControl"] = f

	return cmd
}
func newAdcChopControlCommand() *cobra.Command {
	options := driver.ChopControlOpts{}
	cmd := &cobra.Command{
		Use:   "ChopControl",
		Short: "",
		Long:  "",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			tx, rx, err = adcConnection.ChopControl(options, chipSelect)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVar(&options.Write, "write", false, "set the write bit")
	f.Uint8Var(&options.GrpaChop, "grpa-chop", 0, "1: Chop at f MOD /8, 2: Chop at f MOD /32")
	f.Uint8Var(&options.GrpbChop, "grpb-chop", 0, "1: Chop at f MOD /8, 2: Chop at f MOD /32")
	//flagsList["ChopControl"] = f

	return cmd
}
func newAdcHardResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "HardReset",
		Short: "Perform hard reset",
		Long:  `Hard reset is recommended before using the adc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				tx  = make([]byte, 2)
				rx  = make([]byte, 2)
			)

			_, _, err = adcConnection.HardReset(nil)
			if err != nil {
				return err
			}

			if debug {
				log.Println(tx, rx)
			}
			return err
		},
	}
	return cmd
}

func SendSyncSignal() {
	bcm283x.GPIO7.FastOut(gpio.Low)
	bcm283x.GPIO7.FastOut(gpio.High)
}

func init() {
	var f *flag.FlagSet
	rootCmd.AddCommand(adcCmd)
	adcCmd.AddCommand(
		newAdcChStandbyCommand(),
		newAdcChModeACommand(),
		newAdcChModeBCommand(),
		newAdcChModeSelectCommand(),
		newAdcPowerModeCommand(),
		newAdcGeneralConfigurationCommand(),
		newAdcDataControlCommand(),
		newAdcInterfaceConfigurationCommand(),
		newAdcBISTControlCommand(),
		newAdcDeviceStatusCommand(),
		newAdcRevisionIDCommand(),
		newAdcGPIOControlCommand(),
		newAdcGPIOWriteDataCommand(),
		newAdcGPIOReadDataCommand(),
		newAdcPrechargeBuffer1Command(),
		newAdcPrechargeBuffer2Command(),
		newAdcPositiveRefPrechargeBufCommand(),
		newAdcNegativeRefPrechargeBufCommand(),
		newAdcChannelOffsetCommand(),
		newAdcChGainCommandCommand(),
		newAdcChannelSyncOffsetCommand(),
		newAdcDiagnosticRXCommand(),
		newAdcDiagnosticMuxControlCommand(),
		newAdcModulatorDelayControlCommand(),
		newAdcChopControlCommand(),
		newAdcHardResetCommand(),
	)

	f = adcCmd.PersistentFlags()
	f.Uint8Var(&chipSelect, "adc", 0, "select the ADC to control: [1..9], 0: all")
	_ = adcCmd.MarkFlagRequired("adc")

}
