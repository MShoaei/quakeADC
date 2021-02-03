package server

import (
	"net/http"
	"strconv"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/gin-gonic/gin"
)

func (s *server) commandHandler(c *gin.Context) {
	adc, err := strconv.ParseUint(c.Param("adc"), 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	switch c.Param("cmd") {
	case "ChStandby":
		opts := driver.ChStandbyOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.ChStandby(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.ChStandby(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeA":
		opts := driver.ChModeOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.ChModeA(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.ChModeA(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeB":
		opts := driver.ChModeOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.ChModeB(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.ChModeB(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeSel":
		opts := driver.ChModeSelectOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.ChModeSel(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.ChModeSel(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PowerMode":
		opts := driver.PowerModeOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.PowerMode(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.PowerMode(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GeneralConf":
		opts := driver.GeneralConfOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.GeneralConf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.GeneralConf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DataControl":
		opts := driver.DataControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.DataControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.DataControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "InterfaceConf":
		opts := driver.InterfaceConfOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.InterfaceConf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.InterfaceConf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "BISTControl":
		opts := driver.BISTControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.BISTControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.BISTControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DeviceStatus":
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.DeviceStatus(uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.DeviceStatus(i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "RevisionID":
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.RevisionID(uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.RevisionID(i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOControl":
		opts := driver.GPIOControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.GPIOControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.GPIOControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOWriteData":
		opts := driver.GPIOWriteDataOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.GPIOWriteData(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.GPIOWriteData(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOReadData":
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.GPIOReadData(uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.GPIOReadData(i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PrechargeBuffer1":
		opts := driver.PreChargeBufferOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.PrechargeBuffer1(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.PrechargeBuffer1(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PrechargeBuffer2":
		opts := driver.PreChargeBufferOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.PrechargeBuffer2(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.PrechargeBuffer2(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PositiveRefPrechargeBuf":
		opts := driver.ReferencePrechargeBufOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.PositiveRefPrechargeBuf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.PositiveRefPrechargeBuf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "NegativeRefPrechargeBuf":
		opts := driver.ReferencePrechargeBufOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.NegativeRefPrechargeBuf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.NegativeRefPrechargeBuf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChannelOffset":
		opts := driver.ChannelOffsetOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			err := s.adc.ChannelOffset(opts, uint8(adc), s.Debug)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, nil)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := s.adc.ChannelOffset(opts, i, s.Debug)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusOK, nil)
		return

	case "ChannelGain":
		opts := driver.ChannelGainOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			_, err := s.adc.ChannelGain(opts, uint8(adc), s.Debug)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, nil)
			return
		}
		for i := uint8(1); i < 10; i++ {
			_, err := s.adc.ChannelGain(opts, i, s.Debug)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
	case "ChannelSyncOffset":
		opts := driver.ChannelSyncOffsetOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			err := s.adc.ChannelSyncOffset(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			c.JSON(http.StatusOK, nil)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := s.adc.ChannelSyncOffset(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
	case "DiagnosticRX":
		opts := driver.DiagnosticRXOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.DiagnosticRX(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.DiagnosticRX(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DiagnosticMuxControl":
		opts := driver.DiagnosticMuxControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.DiagnosticMuxControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.DiagnosticMuxControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ModulatorDelayControl":
		opts := driver.ModulatorDelayControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.ModulatorDelayControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.ModulatorDelayControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChopControl":
		opts := driver.ChopControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := s.adc.ChopControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := s.adc.ChopControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "HardReset":
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "command not found",
		})
		return
	}
}
