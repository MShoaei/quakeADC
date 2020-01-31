package cmd

import (
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	flag "github.com/spf13/pflag"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var readParams struct {
	File     string `json:"file"`
	Skip     int    `json:"skip"`
	Duration int    `json:"duration"`
}

var dataFile *os.File

// CommandsList is the list of all available commands
var CommandsList = map[string]func(*flag.FlagSet) ([]byte, []byte, error){
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

// NewAPI creates a new API which receives commands and executes them
// the server should be started with:
// $ rpiCMD server
// the server will listen on port "9090" and accepts POST request to "/command".
// POST request should be a JSON object. e.g.:
//{
//	"Command": "chStandby",
//	"Flags": "--write --ch3 t --ch2 t --ch1 t --ch0 f"
//}
func NewAPI() *iris.Application {
	api := iris.Default()
	var templates string
	if templates = os.Getenv("TEMPLATES_DIR"); templates == "" {
		templates = "./templates"
	}
	api.RegisterView(iris.HTML(templates, ".html").Reload(true))

	api.Get("/", homeHandler)
	api.Post("/command", commandHandler)
	api.Get("/readlive", readLiveHandler)
	api.Post("/readlive", readLivePostHandler)
	// api.Any("/readlive", readLiveHandler)
	api.Get("/getfile", getFileHandler)

	return api
}

func commandHandler(ctx iris.Context) {
	data := &struct {
		Command string
		Flags   []struct {
			Name  string
			Value string
		}
	}{}
	set := flag.NewFlagSet("quakeADC", flag.ExitOnError)

	err := ctx.ReadJSON(data)
	if err != nil {
		ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Println(err)
		return
	}

	cmd := CommandsList[data.Command]
	if cmd == nil {
		ctx.JSON(iris.Map{
			"message": "Unknown command",
			"command": data.Command,
			//"error":   err.Error(),
		})
		log.Printf("unknown command: %s", data.Command)
		return
	}

	flags := make([]string, 0, 7)
	for _, f := range data.Flags {
		flags = append(flags, f.Name+"="+f.Value)
	}
	err = set.ParseAll(flags, nil)
	if err != nil {
		ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Printf("flag parse failed with error: %s", err)
		return
	}

	tx, rx, err := cmd(set)
	if err != nil {
		ctx.StatusCode(iris.StatusNotAcceptable)
		ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Printf("command %s failed: %s", data.Command, err)
		return
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(iris.Map{
		"message": "success",
		"tx":      tx,
		"rx":      rx,
	})
}

func readLiveHandler(ctx iris.Context) {
	log.Println("readLiveHandler called")
	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Fatal(err)
	}
	dataFile, err = os.OpenFile(readParams.File+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	rcvToSend := make(chan string)
	go read(readOptions{
		file:     dataFile,
		skip:     readParams.Skip,
		duration: readParams.Duration,
		ch:       rcvToSend,
	})
	time.Sleep(1 * time.Second)
	for {
		data, ok := <-rcvToSend
		if !ok {
			conn.WriteMessage(websocket.TextMessage, []byte("Send finished"))
			conn.Close()
			// ctx.Redirect("/getfile")
			// ctx.ResetResponseWriter(ctx.ResponseWriter())
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte(data))
		// time.Sleep(1 * time.Second)
	}
}

func readLivePostHandler(ctx iris.Context) {
	log.Println("readLivePostHandler called")
	ctx.ReadJSON(&readParams)
	ctx.JSON(iris.Map{"code": 200})
}

func getFileHandler(ctx iris.Context) {
	err := ctx.SendFile(dataFile.Name(), dataFile.Name())
	if err != nil {
		log.Println(err)
	}
	return
}

func homeHandler(ctx iris.Context) {
	ctx.View("ws.html")
}
