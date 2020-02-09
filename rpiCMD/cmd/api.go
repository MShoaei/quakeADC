package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	flag "github.com/spf13/pflag"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Println(r.Header["Origin"][0])
		return true
	},
}

var readParams struct {
	File     string `json:"file"`
	Skip     int    `json:"skip"`
	Duration int    `json:"duration"`
}

var dataFile *os.File

// CommandsList is the list of all available commands
var CommandsList map[string]func(*flag.FlagSet) ([]byte, []byte, error)

var flagsList = map[string]*flag.FlagSet{}

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
	//var templates string
	//if templates = os.Getenv("TEMPLATES_DIR"); templates == "" {
	//	templates = "./templates"
	//}
	//api.RegisterView(iris.HTML("dist", ".html").Reload(true))

	api.Get("/", homeHandler)

	api.Get("/readlive", readLiveHandler)
	api.Post("/readlive", readLivePostHandler)
	// api.Any("/readlive", readLiveHandler)
	api.Use(cors.AllowAll())
	api.Options("/command", homeHandler)
	api.Post("/command", commandHandler)
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

	err := ctx.ReadJSON(data)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Println(err)
		return
	}

	cmd := CommandsList[data.Command]
	if cmd == nil {
		ctx.StatusCode(iris.StatusNotImplemented)
		ctx.JSON(iris.Map{
			"message": "Unknown command",
			"command": data.Command,
			//"error":   err.Error(),
		})
		log.Printf("unknown command: %s", data.Command)
		return
	}

	set := flagsList[data.Command]
	if set == nil { // should never happen!
		ctx.StatusCode(iris.StatusNotImplemented)
		ctx.JSON(iris.Map{
			"message": "flag set not found",
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
	err = set.Parse(flags)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
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
	log.Printf("tx: %v, rx: %v", tx, rx)
	ctx.JSON(iris.Map{
		"message": "success",
		"tx":      fmt.Sprintf("%v", tx),
		"rx":      fmt.Sprintf("%v", rx),
	})
}

func readLiveHandler(ctx iris.Context) {
	log.Println("readLiveHandler called")
	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Println(err)
		return
	}
	dataFile, err = os.OpenFile(path.Join("/", "tmp", readParams.File+".txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
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
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte(data))
		// time.Sleep(1 * time.Second)
	}
}

func readLivePostHandler(ctx iris.Context) {
	log.Println("readLivePostHandler called")
	ctx.ReadJSON(&readParams)
	log.Println(readParams)
	ctx.JSON(iris.Map{"code": 200})
}

func getFileHandler(ctx iris.Context) {
	err := ctx.SendFile(dataFile.Name(), path.Base(dataFile.Name()))
	if err != nil {
		log.Println(err)
	}
	return
}

func homeHandler(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"message": "Home api",
	})
}
