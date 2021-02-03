package server

import (
	"fmt"
	"strings"
)

type usbDevice struct {
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	MountPoint string      `json:"mountpoint"`
	Size       string      `json:"size"`
	Children   []usbDevice `json:"children"`
}

type RXResponse []byte

func (r RXResponse) MarshalJSON() ([]byte, error) {
	var result string
	if r == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", r)), ",")
	}
	return []byte(result), nil
}
