package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
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

// NewAPI creates a new API which recieves commands and executes them
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
	api.RegisterView(iris.HTML(templates, ".html"))

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

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(map[string]string{
		"message": "success",
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
		log.Fatal(err)
	}
	return
}

func homeHandler(ctx iris.Context) {
	ctx.View("ws.html")
}
