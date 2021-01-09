package seg2

type dataFormat byte

const (
	Fixed16 dataFormat = 0x01 + iota
	Fixed32
	// Float20
	_
	Float32
	Float64
)

func (d dataFormat) size() int {
	switch d {
	case Fixed16:
		return 2
	case Fixed32, Float32:
		return 4
	case Float64:
		return 8
	default:
		// this should never happen!
		return 0
	}
}

type traceDescriptorBlock struct {
	// size of block
	x uint16

	// size of data block
	y uint32

	// number of samples
	ns uint32

	// format of the data in block
	format dataFormat

	// information about aquisition
	info string

	// block data
	data []byte
}

func NewTraceDescriptor(info []string, data [][]byte, format dataFormat) []*traceDescriptorBlock {

	if len(info) != len(data) {
		return nil
	}

	result := make([]*traceDescriptorBlock, 0, len(data))
	for i := 0; i < len(data); i++ {
		if len(info[i])%4 != 0 {
			info[i] = info[i] + string(make([]byte, (4-len(info[i])%4), (4-len(info[i])%4)))
		}

		if len(data)%4 != 0 {
			data[i] = append(data[i], make([]byte, 4-len(data[i])%4, 4-len(data[i])%4)...)
		}
		result = append(result, &traceDescriptorBlock{
			x:      uint16(32 + len(info)),
			y:      uint32(len(data)),
			ns:     uint32(len(data) / format.size()),
			format: format,
			info:   info[i],
			data:   data[i],
		})
	}
	return result
}
