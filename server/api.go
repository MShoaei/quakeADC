package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Println(r.Header["Origin"][0])
		return true
	},
}

func (s *server) newAPI() *gin.Engine {
	api := gin.Default()
	if s.Debug {
		api.Any("/api/:path", func(c *gin.Context) {
			r := c.Request
			r.URL.Path = strings.Replace(r.URL.Path, "/api", "", 1)
			//c.Application().ServeHTTPC(c)
		})
	}
	api.Use(cors.Default())

	api.GET("/status", s.samplingStatusHandler)

	api.GET("/tree/*dir", s.treeHandler)
	api.DELETE("/tree/*path", s.treeDeleteHandler)
	api.PATCH("/tree/*path", s.treePatchHandler)

	api.GET("/plot", s.readDataHandler)
	api.POST("/plot", s.readDataPostHandler)

	api.GET("/dl/*path", func(c *gin.Context) {
		c.Header("cache-control", "no-store, max-age=0")
	}, s.DownloadSampleHandler)

	api.POST("/setup", s.setupHandler)
	api.POST("/command/:cmd/:adc", s.commandHandler)
	api.GET("/getfile", s.getFileHandler)

	api.GET("/usb", s.GetAllUSBHandler)

	api.POST("/rpi/shutdown", s.shutdownSequenceHandler)
	api.POST("/rpi/restart", s.restartSequenceHandler)
	api.GET("/channels", s.getChannelsHandler)
	api.POST("/channels", s.setChannelsHandler)
	api.GET("/gains", s.getGainsHandler)
	api.POST("/gains", s.setGainsHandler)
	api.GET("/info", s.boardInfoHandler)
	api.POST("/calibrate", func(c *gin.Context) {
		s.adc.CilabrateChOffset(s.logics[0], s.Debug)
		for i := 0; i < len(s.hd.EnabledChannels); i++ {
			s.hd.EnabledChannels[i] = true
			s.hd.Gains[i] = 1000
		}
		c.Status(http.StatusOK)
	})

	api.POST("/save/project", s.SaveProjectFolder)
	api.POST("/save/sample", s.SaveSampleFile)

	api.PATCH("/multiplier", func(c *gin.Context) {
		val, err := strconv.Atoi(c.Query("val"))
		if err != nil {
			c.String(http.StatusBadRequest, "%v", err)
		}
		s.GainMultiply = uint32(val)
		c.String(http.StatusOK, "%d", s.GainMultiply)
	})
	return api
}
