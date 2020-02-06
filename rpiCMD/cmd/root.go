package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/MShoaei/quakeADC/driver"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string
var adcConnection *driver.Adc77684

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rpiCMD",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var (
			err             error
			speed           int64
			bus, chip, mode int
		)

		if skip, _ := cmd.Flags().GetBool("skip"); skip {
			log.Println("Skipping")
			return nil
		}

		if adcConnection != nil {
			return nil
		}
		bus, err = cmd.Flags().GetInt("bus")
		if err != nil {
			return err
		}
		chip, err = cmd.Flags().GetInt("chip")
		if err != nil {
			return err
		}
		mode, err = cmd.Flags().GetInt("mode")
		if err != nil {
			return err
		}
		speed, err = cmd.Flags().GetInt64("speed")
		if err != nil {
			return err
		}
		if mode < 0 || mode > 3 {
			return fmt.Errorf("invalid mode! expected value [0..3], got %d", mode)
		}

		if adcConnection != nil {
			return nil
		}
		adcConnection, err = driver.GetSpiConnection(bus, chip, mode, 8, speed)
		if err != nil {
			return err
		}
		initCommands()
		log.Println("initialized")
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			log.Println(speed)
		}
		return err
	},
}

func initCommands() {
	CommandsList = map[string]func(*flag.FlagSet) ([]byte, []byte, error){
		"ChStandby":               adcConnection.ChStandby,
		"ChModeA":                 adcConnection.ChModeA,
		"ChModeB":                 adcConnection.ChModeB,
		"ChModeSel":               adcConnection.ChModeSel,
		"PowerMode":               adcConnection.PowerMode,
		"GeneralConf":             adcConnection.GeneralConf,
		"DataControl":             adcConnection.DataControl,
		"InterfaceConf":           adcConnection.InterfaceConf,
		"BISTControl":             adcConnection.BISTControl,
		"DeviceStatus":            adcConnection.DeviceStatus,
		"RevisionID":              adcConnection.RevisionID,
		"GPIOControl":             adcConnection.GPIOControl,
		"GPIOWriteData":           adcConnection.GPIOWriteData,
		"GPIOReadData":            adcConnection.GPIOReadData,
		"PrechargeBuffer1":        adcConnection.PrechargeBuffer1,
		"PrechargeBuffer2":        adcConnection.PrechargeBuffer2,
		"PositiveRefPrechargeBuf": adcConnection.PositiveRefPrechargeBuf,
		"NegativeRefPrechargeBuf": adcConnection.NegativeRefPrechargeBuf,
		"Ch0OffsetMSB":            adcConnection.Ch0OffsetMSB,
		"Ch0OffsetMid":            adcConnection.Ch0OffsetMid,
		"Ch0OffsetLSB":            adcConnection.Ch0OffsetLSB,
		"Ch1OffsetMSB":            adcConnection.Ch1OffsetMSB,
		"Ch1OffsetMid":            adcConnection.Ch1OffsetMid,
		"Ch1OffsetLSB":            adcConnection.Ch1OffsetLSB,
		"Ch2OffsetMSB":            adcConnection.Ch2OffsetMSB,
		"Ch2OffsetMid":            adcConnection.Ch2OffsetMid,
		"Ch2OffsetLSB":            adcConnection.Ch2OffsetLSB,
		"Ch3OffsetMSB":            adcConnection.Ch3OffsetMSB,
		"Ch3OffsetMid":            adcConnection.Ch3OffsetMid,
		"Ch3OffsetLSB":            adcConnection.Ch3OffsetLSB,
		"Ch0SyncOffset":           adcConnection.Ch0SyncOffset,
		"Ch1SyncOffset":           adcConnection.Ch1SyncOffset,
		"Ch2SyncOffset":           adcConnection.Ch2SyncOffset,
		"Ch3SyncOffset":           adcConnection.Ch3SyncOffset,
		"DiagnosticRX":            adcConnection.DiagnosticRX,
		"DiagnosticMuxControl":    adcConnection.DiagnosticMuxControl,
		"DiagnosticDelayControl":  adcConnection.DiagnosticDelayControl,
		"ChopControl":             adcConnection.ChopControl,
		"SoftReset":               adcConnection.SoftReset,
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if adcConnection != nil {
		_ = adcConnection.Close()
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var f *flag.FlagSet
	f = rootCmd.PersistentFlags()
	f.Int("bus", 0, "spi bus number and is usually 0")
	f.Int("chip", 0, "spi chipSelect number")
	f.Int("mode", 0, "spi mode number [0..3]")
	f.Int64("speed", 50000, "spi connection speed in Hz")
	f.BoolP("debug", "V", false, "Debug Mode. Print Sent and received values.")
	f.BoolP("skip", "S", false, "Skip initializing spi connection. ONLY FOR TEST")
	c := f.Lookup("debug")
	c.NoOptDefVal = "true"
	f.SortFlags = false

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/rpiGo/.rpiCMD.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".rpiCMD" (without extension).
		viper.AddConfigPath(path.Join(home, "rpiGo"))
		viper.SetConfigName(".rpiCMD")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
