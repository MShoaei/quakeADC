package cmd

import (
	"log"
	"strings"

	"github.com/kataras/iris/v12"
)

// NewAPI creates a new API which recieves commands and executes them
// the server should be started with:
// $ rpiCMD server
// the server will listen on port "9091" and accepts POST request to "/command".
// POST request should be a JSON object. e.g.:
//{
//	"Command": "chStandby",
//	"Flags": "--write --ch3 t --ch2 t --ch1 t --ch0 f"
//}
func NewAPI() *iris.Application {
	api := iris.Default()

	api.Post("/command", commandHandler)

	return api
}

func commandHandler(ctx iris.Context) {
	data := &struct {
		Command string
		Flags   string
	}{}
	err := ctx.ReadJSON(data)
	if err != nil {
		log.Println(err)
	}
	log.Println(data)

	a := []string{"adc", data.Command}
	a = append(a, strings.Split(data.Flags, " ")...)

	// cmd := cmdMap[data.Command]
	rootCmd.SetArgs(a)
	if err := rootCmd.Execute(); err != nil {
		ctx.StatusCode(iris.StatusNotAcceptable)
		ctx.JSON(map[string]string{
			"message": "failed with error",
		})
		log.Printf("command %s failed: %s", data.Command, err)
		return
	}

	ctx.JSON(map[string]string{
		"message": "success",
	})
	ctx.StatusCode(iris.StatusOK)
}
