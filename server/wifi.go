package server

import (
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-cmd/cmd"
)

func (s *Server) ScanNetworks(c *gin.Context) {
	aps, err := scan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"accessPoints": aps,
	})
}

func (s *Server) Connect(c *gin.Context) {
	data := struct {
		ESSID    string `json:"essid"`
		Password string `json:"password"`
	}{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := connect(data.ESSID, data.Password); err != nil {
		s.l.Errorf("failed to connect: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	status := <-cmd.NewCmd("/bin/ping", "-c", "4", "google.com").Start()
	if status.Exit != 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": status.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

type accessPoint struct {
	ID    string `json:"id,omitempty"`
	ESSID string `json:"essid"`
}

func scan() ([]accessPoint, error) {
	scanStatus := <-cmd.NewCmd("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "scan").Start()
	if scanStatus.Exit != 0 {
		return nil, scanStatus.Error
	}
	time.Sleep(500 * time.Millisecond)
	scanStatus = <-cmd.NewCmd("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "scan_results").Start()
	if scanStatus.Exit != 0 {
		return nil, scanStatus.Error
	}

	result := make([]accessPoint, 0)
	for i := 1; i < len(scanStatus.Stdout); i++ {
		line := strings.Fields(scanStatus.Stdout[i])
		result = append(result, accessPoint{ESSID: line[4]})
	}
	return result, nil
}

func knownNetworks() []accessPoint {
	result := make([]accessPoint, 0)
	all := (<-cmd.NewCmd("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "list_networks").Start()).Stdout
	for i := 1; i < len(all); i++ {
		line := strings.Fields(all[i])
		result = append(result, accessPoint{ID: line[0], ESSID: line[1]})
	}
	return result
}

func connect(essid string, password string) error {
	status := <-cmd.NewCmd("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "add_network").Start()
	id := status.Stdout[0]

	if err := exec.Command("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "set_network", id, "ssid", "\""+essid+"\"").Run(); err != nil {
		return err
	}

	if err := exec.Command("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "set_network", id, "psk", `"`+password+`"`).Run(); err != nil {
		return err
	}

	if err := exec.Command("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "select_network", id).Run(); err != nil {
		return err
	}
	if err := exec.Command("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "save_config").Run(); err != nil {
		return err
	}
	if err := exec.Command("/usr/bin/sudo", "/sbin/wpa_cli", "-i", "wlan0", "reassociate").Run(); err != nil {
		return err
	}

	return status.Error
}
