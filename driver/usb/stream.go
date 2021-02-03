package usb

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/google/gousb"
)

const (
	cmdStartFlagsCLK48MHZ   uint8 = 1 << 6
	cmdStartFlagsSample8Bit uint8 = 0 << 5
	delay                   uint8 = 1
)

type cmdStartAcquisition struct {
	Flags        uint8
	SampleDelayH uint8
	SampleDelayL uint8
}

type streamConnection struct {
	ctx  *gousb.Context
	devs []*gousb.Device
	cfg  *gousb.Config
	intf *gousb.Interface
	epIn *gousb.InEndpoint

	Stream *gousb.ReadStream
}

func NewReadStream() (s *streamConnection, err error) {
	var (
		ctx  *gousb.Context
		devs []*gousb.Device
		cfg  *gousb.Config
		intf *gousb.Interface
		epIn *gousb.InEndpoint

		stream *gousb.ReadStream
	)
	defer func() {
		if err == nil {
			return
		}

		if intf != nil {
			intf.Close()
		}

		if cfg != nil {
			cfg.Close()
		}
		for _, d := range devs {
			if d != nil {
				d.Close()
			}
		}
		if ctx != nil {
			ctx.Close()
		}
	}()

	ctx = gousb.NewContext()

	vid, pid := gousb.ID(0x0925), gousb.ID(0x3881)
	devs, err = ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return desc.Vendor == vid && desc.Product == pid
	})

	if err != nil {
		return nil, fmt.Errorf("OpenDevices(): %w", err)
	}
	if len(devs) == 0 {
		return nil, fmt.Errorf("no devices found matching VID %s and PID %s", vid, pid)
	}

	dev := devs[0]

	cfg, err = dev.Config(1)
	if err != nil {
		return nil, fmt.Errorf("%s.Config(1): %w", dev, err)
	}

	cmd := cmdStartAcquisition{}
	cmd.Flags = cmdStartFlagsCLK48MHZ
	cmd.Flags |= cmdStartFlagsSample8Bit
	cmd.Flags |= 0 // not using analog channels
	cmd.SampleDelayH = (delay >> 8) & 0xff
	cmd.SampleDelayL = delay & 0xff

	const sz = int(unsafe.Sizeof(cmdStartAcquisition{}))
	var asByteSlice []byte = (*(*[sz]byte)(unsafe.Pointer(&cmd)))[:]

	num, err := dev.Control(gousb.ControlVendor|gousb.ControlOut, 0xb1, 0, 0, asByteSlice)
	if num != 3 || err != nil {
		return nil, fmt.Errorf("device control failed: %w", err)
	}

	intf, err = cfg.Interface(0, 0)
	if err != nil {
		return nil, fmt.Errorf("%s.Interface(0, 0): %w", cfg, err)
	}

	epIn, err = intf.InEndpoint(2)
	if err != nil {
		return nil, fmt.Errorf("%s.InEndpoint(2): %w", intf, err)
	}
	log.Println(epIn.Desc.MaxPacketSize)

	stream, err = epIn.NewStream(512*10, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to create ReadStream: %W", err)
	}

	s = &streamConnection{
		ctx:  ctx,
		devs: devs,
		cfg:  cfg,
		intf: intf,
		epIn: epIn,

		Stream: stream,
	}
	return s, nil
}

func (s *streamConnection) Close() {
	s.intf.Close()
	s.cfg.Close()
	for _, d := range s.devs {
		if d != nil {
			d.Close()
		}
	}
	s.ctx.Close()
}
