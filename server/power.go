package server

import (
	"github.com/MShoaei/quakeADC/driver"
	"github.com/gin-gonic/gin"
	"github.com/go-cmd/cmd"
)

func (s *server) restartSequenceHandler(c *gin.Context) {
	driver.Reset()
	cmd.NewCmd("/usr/bin/sudo", "/sbin/shutdown", "-r", "now").Start()

}

func (s *server) shutdownSequenceHandler(c *gin.Context) {
	driver.Shutdown(s.adc.Connection())
}
