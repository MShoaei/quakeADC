package cmd

import (
	"log"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var dataFS afero.Fs

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "launch a server to execute command",
	Long:  `launch a server which listens on port 9090 and executes commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := NewAPI()
		port := "9090"
		if os.Getenv("PORT") != "" {
			port = os.Getenv("PORT")
		}
		api.Listen(":" + port)
	},
}

func init() {
	wd, _ := os.Getwd()
	if err := os.MkdirAll(path.Join(wd, "data"), os.ModeDir|0755); err != nil {
		log.Fatalf("failed to create directory: %v", err)
	}
	dataFS = afero.NewBasePathFs(afero.NewOsFs(), path.Join(wd, "data"))
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
