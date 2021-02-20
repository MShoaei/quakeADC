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

func (s *Server) NewAPI() *gin.Engine {
	api := gin.Default()
	if s.Debug {
		api.Any("/api/:path", func(c *gin.Context) {
			r := c.Request
			r.URL.Path = strings.Replace(r.URL.Path, "/api", "", 1)
			//c.Application().ServeHTTPC(c)
		})
	}
	api.Use(cors.Default())

	api.GET("/status", s.SamplingStatusHandler)

	api.GET("/tree/*dir", s.TreeHandler)
	api.DELETE("/tree/*path", s.TreeDeleteHandler)
	api.PATCH("/tree/*path", s.TreePatchHandler)
	api.POST("/tree", s.CreateNewProject)
	api.GET("/project/active", s.GetActiveProjectPath)
	api.PATCH("/project/active", s.SetActiveProjectPath)

	api.GET("/wifi/scan", s.ScanNetworks)
	api.POST("/wifi/connect", s.Connect)

	api.GET("/plot", s.ReadDataHandler)
	api.POST("/plot", s.ReadDataPostHandler)

	api.GET("/dl/*path", func(c *gin.Context) {
		c.Header("cache-control", "no-store, max-age=0")
	}, s.DownloadSampleHandler)

	api.POST("/setup", s.SetupHandler)
	api.POST("/command/:cmd/:adc", s.CommandHandler)
	api.GET("/getfile", s.GetFileHandler)

	api.GET("/usb", s.GetAllUSBHandler)

	api.POST("/rpi/shutdown", s.ShutdownSequenceHandler)
	api.POST("/rpi/restart", s.RestartSequenceHandler)
	api.GET("/channels", s.GetChannelsHandler)
	api.POST("/channels", s.SetChannelsHandler)
	api.GET("/gains", s.GetGainsHandler)
	api.POST("/gains", s.SetGainsHandler)
	api.GET("/info", s.BoardInfoHandler)
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
