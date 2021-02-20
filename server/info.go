package server

import (
	"net/http"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/gin-gonic/gin"
)

func (s *Server) BoardInfoHandler(c *gin.Context) {
	type message struct {
		Voltage []int16 `json:"voltage"`
		Current []int16 `json:"current"`
	}
	voltage := driver.GetVoltage(s.adc.Connection())
	current := driver.GetCurrent(s.adc.Connection())
	m := message{
		Voltage: voltage,
		Current: current,
	}
	c.JSON(http.StatusOK, &m)
}

func (s *Server) SamplingStatusHandler(c *gin.Context) {
	if s.sigrokRunning {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "running",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "not running",
	})
}
