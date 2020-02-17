package cmd

import (
	"log"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// adcCmd represents the adc command
var adcCmd = &cobra.Command{
	Use:   "adc",
	Short: "command to control the ADC over SPI",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

var adcChStandby = &cobra.Command{
	Use:   "ChStandby",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.ChStandby(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err

	},
}

var adcChModeA = &cobra.Command{
	Use:   "ChModeA",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.ChModeA(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcChModeB = &cobra.Command{
	Use:   "ChModeB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.ChModeB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcChModeSelect = &cobra.Command{
	Use:   "ChModeSel",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.ChModeSel(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcPowerMode = &cobra.Command{
	Use:   "PowerMode",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.PowerMode(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcGeneralConfiguration = &cobra.Command{
	Use:   "GeneralConf",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.GeneralConf(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcDataControl = &cobra.Command{
	Use:   "DataControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.DataControl(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcInterfaceConfiguration = &cobra.Command{
	Use:   "InterfaceConf",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.InterfaceConf(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcBISTControl = &cobra.Command{
	Use:   "BISTControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.BISTControl(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcDeviceStatus = &cobra.Command{
	Use:   "DeviceStatus",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.DeviceStatus(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err

	},
}

var adcRevisionID = &cobra.Command{
	Use:   "RevisionID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.RevisionID(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err

	},
}

var adcGPIOControl = &cobra.Command{
	Use:   "GPIOControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.GPIOControl(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

// TODO: Need better explanation for flags
var adcGPIOWriteData = &cobra.Command{
	Use:   "GPIOWriteData",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.GPIOWriteData(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcGPIOReadData = &cobra.Command{
	Use:   "GPIOReadData",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.GPIOReadData(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcPrechargeBuffer1 = &cobra.Command{
	Use:   "PrechargeBuffer1",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.PrechargeBuffer1(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcPrechargeBuffer2 = &cobra.Command{
	Use:   "PrechargeBuffer2",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.PrechargeBuffer2(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcPositiveRefPrechargeBuf = &cobra.Command{
	Use:     "PositiveRefPrechargeBuf",
	Aliases: []string{"prpb"},
	Short:   "",
	Long:    "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.PositiveRefPrechargeBuf(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcNegativeRefPrechargeBuf = &cobra.Command{
	Use:     "NegativeRefPrechargeBuf",
	Aliases: []string{"nrpb"},
	Short:   "",
	Long:    "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.NegativeRefPrechargeBuf(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0OffsetMSB = &cobra.Command{
	Use:   "Ch0OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0OffsetMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0OffsetMid = &cobra.Command{
	Use:   "Ch0OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0OffsetMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0OffsetLSB = &cobra.Command{
	Use:   "Ch0OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0OffsetLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1OffsetMSB = &cobra.Command{
	Use:   "Ch1OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1OffsetMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1OffsetMid = &cobra.Command{
	Use:   "Ch1OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1OffsetMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1OffsetLSB = &cobra.Command{
	Use:   "Ch1OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1OffsetLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2OffsetMSB = &cobra.Command{
	Use:   "Ch2OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2OffsetMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2OffsetMid = &cobra.Command{
	Use:   "Ch2OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2OffsetMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2OffsetLSB = &cobra.Command{
	Use:   "Ch2OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2OffsetLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3OffsetMSB = &cobra.Command{
	Use:   "Ch3OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3OffsetMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3OffsetMid = &cobra.Command{
	Use:   "Ch3OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3OffsetMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3OffsetLSB = &cobra.Command{
	Use:   "Ch3OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3OffsetLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0GainMSB = &cobra.Command{
	Use:   "Ch0GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0GainMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0GainMid = &cobra.Command{
	Use:   "Ch0GainMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0GainMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0GainLSB = &cobra.Command{
	Use:   "Ch0GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0GainLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1GainMSB = &cobra.Command{
	Use:   "Ch1GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1GainMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1GainMid = &cobra.Command{
	Use:   "Ch1GainMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1GainMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1GainLSB = &cobra.Command{
	Use:   "Ch1GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1GainLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2GainMSB = &cobra.Command{
	Use:   "Ch2GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2GainMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2GainMid = &cobra.Command{
	Use:   "Ch2GainMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2GainMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2GainLSB = &cobra.Command{
	Use:   "Ch2GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2GainLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3GainMSB = &cobra.Command{
	Use:   "Ch3GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3GainMSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3GainMid = &cobra.Command{
	Use:   "Ch3GainMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3GainMid(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3GainLSB = &cobra.Command{
	Use:   "Ch3GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3GainLSB(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh0SyncOffset = &cobra.Command{
	Use:   "Ch0SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch0SyncOffset(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh1SyncOffset = &cobra.Command{
	Use:   "Ch1SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch1SyncOffset(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh2SyncOffset = &cobra.Command{
	Use:   "Ch2SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch2SyncOffset(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcCh3SyncOffset = &cobra.Command{
	Use:   "Ch3SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.Ch3SyncOffset(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcDiagnosticRX = &cobra.Command{
	Use:   "DiagnosticRX",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.DiagnosticRX(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcDiagnosticMuxControl = &cobra.Command{
	Use:   "DiagnosticMuxControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.DiagnosticMuxControl(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcDiagnosticDelayControl = &cobra.Command{
	Use:   "DiagnosticDelayControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.DiagnosticDelayControl(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcChopControl = &cobra.Command{
	Use:   "ChopControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		tx, rx, err = adcConnection.ChopControl(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

var adcHardReset = &cobra.Command{
	Use:   "HardReset",
	Short: "Perform hard reset",
	Long:  `Hard reset is recommended before using the adc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
			tx  = make([]byte, 2)
			rx  = make([]byte, 2)
		)

		_, _, err = adcConnection.HardReset(cmd.Flags())
		if err != nil {
			return err
		}

		if debug, _ := adcCmd.Flags().GetBool("debug"); debug {
			log.Println(tx, rx)
		}
		return err
	},
}

func init() {
	var f *flag.FlagSet
	rootCmd.AddCommand(adcCmd)
	adcCmd.AddCommand(adcChStandby, adcChModeA, adcChModeB, adcChModeSelect,
		adcPowerMode, adcGeneralConfiguration, adcDataControl, adcInterfaceConfiguration, adcBISTControl, adcDeviceStatus,
		adcRevisionID, adcGPIOControl, adcGPIOWriteData, adcGPIOReadData, adcPrechargeBuffer1, adcPrechargeBuffer2,
		adcPositiveRefPrechargeBuf, adcNegativeRefPrechargeBuf,
		adcCh0OffsetMSB, adcCh0OffsetMid, adcCh0OffsetLSB, adcCh1OffsetMSB, adcCh1OffsetMid, adcCh1OffsetLSB,
		adcCh2OffsetMSB, adcCh2OffsetMid, adcCh2OffsetLSB, adcCh3OffsetMSB, adcCh3OffsetMid, adcCh3OffsetLSB,
		adcCh0GainMSB, adcCh0GainMid, adcCh0GainLSB, adcCh1GainMSB, adcCh1GainMid, adcCh1GainLSB,
		adcCh2GainMSB, adcCh2GainMid, adcCh2GainLSB, adcCh3GainMSB, adcCh3GainMid, adcCh3GainLSB,
		adcCh0SyncOffset, adcCh1SyncOffset, adcCh2SyncOffset, adcCh3SyncOffset,
		adcDiagnosticRX, adcDiagnosticMuxControl, adcDiagnosticDelayControl, adcChopControl,
		adcHardReset)

	// ------------------------

	f = adcChStandby.Flags()
	f.Bool("write", false, "set the write bit")
	// true==standby
	// false==enabled
	f.Bool("ch3", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.Bool("ch2", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.Bool("ch1", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.Bool("ch0", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.SortFlags = false
	flagsList["ChStandby"] = f
	// ------------------------

	f = adcChModeA.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	f.Uint16("dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")
	f.SortFlags = false
	flagsList["ChModeA"] = f

	f = adcChModeB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	f.Uint16("dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")
	f.SortFlags = false
	flagsList["ChModeB"] = f

	// ------------------------

	f = adcChModeSelect.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch0", 0, "set channel mode for channels 0. 0:Mode A, 1:Mode B")
	f.Uint8("ch1", 0, "set channel mode for channels 0. 0:Mode A, 1:Mode B")
	f.Uint8("ch2", 0, "set channel mode for channels 0. 0:Mode A, 1:Mode B")
	f.Uint8("ch3", 0, "set channel mode for channels 0. 0:Mode A, 1:Mode B")
	f.SortFlags = false
	flagsList["ChModeSel"] = f

	// ------------------------

	f = adcPowerMode.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("sleep", 0, "0: Normal operation, 1: Sleep mode")
	f.Uint8("power", 0, "0: Low power, 2: Median, 3: Fast")
	f.Uint8("lvds-clk", 0, "0: disable LVDS clock, 1: enable LVDS clock")
	f.Uint8("mclk-div", 0, "0: set to MCLK/32 for low power mode, 2: set to MCLK/8 for median mode, 3: set to MCLK/4 for fast mode")
	f.SortFlags = false
	flagsList["PowerMode"] = f

	// ------------------------

	f = adcGeneralConfiguration.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("retime-en", 0, "SYNC_OUT signal retime enable bit. 0: disabled, 1: enabled")
	f.Uint8("vcm-pd", 0, "VCM buffer power-down. 0: enabled, 1: disabled")
	f.Uint8("vcm-vsel", 0, "VCM voltage select bits. 0: (AVDD1 − AVSS)/2 V, 1: 1.65 V, 2: 2.5 V, 3: 2.14 V")
	f.SortFlags = false
	flagsList["GeneralConf"] = f

	// ------------------------

	f = adcDataControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("spi-sync", 1, "Software synchronization of the AD7768-4. This command has the same effect as sending a signal pulse to the START pin. 0: Change to SPI_SYNC low, 1: Change to SPI_SYNC high")
	f.Uint8("single-shot", 0, "One-shot mode. Enables one-shot mode. In one-shot mode, the AD7768-4 output a conversion result in response to a SYNC_IN rising edge. 0: Disabled, 1: Enabled")
	f.Uint8("spi-reset", 0, "Soft reset. Two successive commands must be received in the correct order to generate a reset. 0: No effect, 1: No effect, 2: Second reset command, 3: First reset command")
	f.SortFlags = false
	flagsList["DataControl"] = f

	// ------------------------

	f = adcInterfaceConfiguration.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("crc-sel", 0, "CRC select. These bits allow the user to implement a CRC on the data interface. 0: No CRC. Status bits with every conversion, 1: Replace the header with CRC message every 4 samples, 2: Replace the header with CRC message every 16 samples, 3: Replace the header with CRC message every 16 samples")
	f.Uint8("dclk-div", 0, "DCLK divider. These bits control division of the DCLK clock used to clock out conversion data on the DOUTx pins. 0: Divide by 8, 1: Divide by 4, 2: Divide by 2, 3: No division")
	f.SortFlags = false
	flagsList["InterfaceConf"] = f

	// ------------------------

	f = adcBISTControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ram-bist-start", 0, "RAM BIST. 0: Off, 1: Begin RAM BIST")
	f.SortFlags = false
	flagsList["BISTControl"] = f

	// ------------------------

	f = adcDeviceStatus.Flags()
	// adcDeviceStatus
	flagsList["DeviceStatus"] = f

	// ------------------------

	f = adcRevisionID.Flags()
	// adcRevisionID
	flagsList["RevisionID"] = f

	// ------------------------

	f = adcGPIOControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ugpio-en", 0, "User GPIO enable. 0: GPIO Disabled, 1: GPIO Enabled")
	f.Uint8("gpio4", 0, "GPIO4 Direction. 0: Input, 1: Output")
	f.Uint8("gpio3", 0, "GPIO3 Direction. 0: Input, 1: Output")
	f.Uint8("gpio2", 0, "GPIO2 Direction. 0: Input, 1: Output")
	f.Uint8("gpio1", 0, "GPIO1 Direction. 0: Input, 1: Output")
	f.Uint8("gpio0", 0, "GPIO0 Direction. 0: Input, 1: Output")
	f.SortFlags = false
	flagsList["GPIOControl"] = f

	// ------------------------

	f = adcGPIOWriteData.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("gpio4", 0, "GPIO4/FILTER")
	f.Uint8("gpio3", 0, "GPIO3/MODE3")
	f.Uint8("gpio2", 0, "GPIO2/MODE2")
	f.Uint8("gpio1", 0, "GPIO1/MODE1")
	f.Uint8("gpio0", 0, "GPIO0/MODE0")
	f.SortFlags = false
	flagsList["GPIOWriteData"] = f

	// ------------------------

	f = adcGPIOReadData.Flags()
	// adcGPIOReadData
	flagsList["GPIOReadData"] = f

	// ------------------------

	f = adcPrechargeBuffer1.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch1-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch1-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch0-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch0-pos", 1, "0: Off, 1: On (default: 1)")
	f.SortFlags = false
	flagsList["PrechargeBuffer1"] = f

	// ------------------------

	f = adcPrechargeBuffer2.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch3-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch2-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch2-pos", 1, "0: Off, 1: On (default: 1)")
	f.SortFlags = false
	flagsList["PrechargeBuffer2"] = f

	// ------------------------

	f = adcPositiveRefPrechargeBuf.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch2", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch1", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch0", 1, "0: Off, 1: On (default: 0)")
	f.SortFlags = false
	flagsList["PositiveRefPrechargeBuf"] = f

	// ------------------------

	f = adcNegativeRefPrechargeBuf.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch2", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch1", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch0", 1, "0: Off, 1: On (default: 0)")
	f.SortFlags = false
	flagsList["NegativeRefPrechargeBuf"] = f

	// ------------------------

	f = adcCh0OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 offset MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0OffsetMSB"] = f

	// ------------------------

	f = adcCh0OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 offset Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0OffsetMid"] = f

	// ------------------------

	f = adcCh0OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 offset LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0OffsetLSB"] = f

	// ------------------------

	f = adcCh1OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 offset MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1OffsetMSB"] = f

	// ------------------------

	f = adcCh1OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 offset Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1OffsetMid"] = f

	// ------------------------

	f = adcCh1OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 offset LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1OffsetLSB"] = f
	// ------------------------

	f = adcCh2OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 offset MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2OffsetMSB"] = f

	// ------------------------

	f = adcCh2OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 offset Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2OffsetMid"] = f

	// ------------------------

	f = adcCh2OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 offset LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2OffsetLSB"] = f
	// ------------------------

	f = adcCh3OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 offset MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3OffsetMSB"] = f

	// ------------------------

	f = adcCh3OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 offset Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3OffsetMid"] = f

	// ------------------------

	f = adcCh3OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 offset LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3OffsetLSB"] = f

	// ------------------------

	f = adcCh0GainMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0GainMSB"] = f

	// ------------------------

	f = adcCh0GainMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0GainMid"] = f

	// ------------------------

	f = adcCh0GainLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0GainLSB"] = f

	// ------------------------

	f = adcCh1GainMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1GainMSB"] = f

	// ------------------------

	f = adcCh1GainMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1GainMid"] = f

	// ------------------------

	f = adcCh1GainLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1GainLSB"] = f

	// ------------------------

	f = adcCh2GainMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2GainMSB"] = f

	// ------------------------

	f = adcCh2GainMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2GainMid"] = f

	// ------------------------

	f = adcCh2GainLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2GainLSB"] = f

	// ------------------------

	f = adcCh3GainMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3GainMSB"] = f

	// ------------------------

	f = adcCh3GainMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3GainMid"] = f

	// ------------------------

	f = adcCh3GainLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3GainLSB"] = f

	// ------------------------

	f = adcCh0SyncOffset.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("offset", 0, "Channel 0 sync phase offset signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch0SyncOffset"] = f

	// ------------------------

	f = adcCh1SyncOffset.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("offset", 0, "Channel 1 sync phase offset signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch1SyncOffset"] = f

	// ------------------------

	f = adcCh2SyncOffset.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("offset", 0, "Channel 2 sync phase offset signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch2SyncOffset"] = f

	// ------------------------

	f = adcCh3SyncOffset.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("offset", 0, "Channel 3 sync phase offset signed 8 bit integer (default: 0)")
	f.SortFlags = false
	flagsList["Ch3SyncOffset"] = f

	// ------------------------

	f = adcDiagnosticRX.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8("ch2", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8("ch1", 0, "0: Not in use, 1: Receive (default: 0)")
	f.Uint8("ch0", 0, "0: Not in use, 1: Receive (default: 0)")
	flagsList["DiagnosticRX"] = f

	// ------------------------

	f = adcDiagnosticMuxControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("grpb-sel", 0, "0: Off, 3: Positive full-scale ADC check, 4: Negative full-scale ADC check, 5: Zero-scale ADC check")
	f.Uint8("grpa-sel", 0, "0: Off, 3: Positive full-scale ADC check, 4: Negative full-scale ADC check, 5: Zero-scale ADC check")
	flagsList["DiagnosticMuxControl"] = f

	// ------------------------

	f = adcDiagnosticDelayControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("mod-delay", 0, "0: Disabled delayed clock for all channels, 1: Enable delayed clock for Channel 0 and Channel 1, 2: Enable delayed clock for Channel 2 and Channel 3, 3: Enable delayed clock for all channels")
	flagsList["DiagnosticDelayControl"] = f

	// ------------------------

	f = adcChopControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("grpa-chop", 0, "1: Chop at f MOD /8, 2: Chop at f MOD /32")
	f.Uint8("grpb-chop", 0, "1: Chop at f MOD /8, 2: Chop at f MOD /32")
	flagsList["ChopControl"] = f

	// ------------------------

	f = adcHardReset.Flags()
	flagsList["HardReset"] = f
}
