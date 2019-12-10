package cmd

import (
	"fmt"

	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
	rpi "github.com/stianeikeland/go-rpio/v4"
)

// Not sure if this works!!
var adcReadDevCmd = &cobra.Command{
	Use:   "readDev",
	Short: "Read /dev/gpio",

	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		clk := make(chan struct{})
		err = rpi.Open()
		if err != nil {
			return err
		}

		go func(ch chan<- struct{}) {
			rpi.DetectEdge(23, rpi.FallEdge)
			for rpi.EdgeDetected(23) {
				ch <- struct{}{}
			}
		}(clk)

		func(ch <-chan struct{}) {
			var data uint32
			rpi.DetectEdge(24, rpi.FallEdge)
			for {
				fmt.Printf("%0b\n", data)
				data = 0
				select {
				case <-ch:
					if rpi.EdgeDetected(24) {
						for i := 0; i < 32; i++ {
							<-ch
							if rpi.ReadPin(22) == 1 {
								data = data << 1
								data |= 1
							} else {
								data = data << 1
							}
						}
					}

				}
			}
		}(clk)
		return nil
	},
}

// readCmd represents the read command
var adcReadCmd = &cobra.Command{
	Use:   "read",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := pcap.FindAllDevs()
		if err != nil {
			return err
		}

		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	},
}

func init() {
	adcCmd.AddCommand(adcReadCmd)
	adcCmd.AddCommand(adcReadDevCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
