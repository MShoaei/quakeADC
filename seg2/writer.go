package seg2

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

var (
	fileDescriptorBlockID      = []byte{0x55, 0x3a}
	revisionNumber             = int16(1)
	m                          = int16(72 * 4)
	sizeOfStringTerminator     = int8(1)
	firstStringTerminatorChar  = '\x00'
	secondStringTerminatorChar = '\x00'
	sizeOfLineTerminator       = int8(1)
	firstLineTerminatorChar    = '\x0a'
	secondLineTerminatorChar   = '\x00'
)

var (
	traceDescriptorBlockID = []byte{0x22, 0x44}
)

type writer struct {
	dateTime time.Time
	n        int16
	note     string

	buf []byte
}

func NewWriter(t time.Time, n int16, note string) *writer {
	if len(note)%4 != 0 {
		note = note + string(make([]byte, (4-len(note)%4), (4-len(note)%4)))
	}

	return &writer{
		dateTime: t,
		n:        n,
		note:     note,

		buf: make([]byte, 0, 20*1024*1024),
	}
}

func (w *writer) Reset(t time.Time, n int16, note string) {
	w.dateTime = t
	w.n = n
	w.note = note
	w.buf = w.buf[:0]
}

func (w *writer) Write(dst io.Writer, data []*traceDescriptorBlock) error {
	w.writeFileHeader()

	if int16(len(data)) != w.n {
		return fmt.Errorf("insufficient number of data blocks. Expectd %d, got %d", w.n, len(data))
	}

	for i := 0; i < len(data); i++ {
		w.writeTraceBlockHeader(i, data[i])
		w.buf = append(w.buf, data[i].data...)
	}

	_, err := dst.Write(w.buf)
	return err
}

func (w *writer) writeFileHeader() {
	w.buf = append(w.buf, fileDescriptorBlockID...)
	temp := make([]byte, 2, 2)
	binary.LittleEndian.PutUint16(temp, uint16(revisionNumber))
	w.buf = append(w.buf, temp...)

	binary.LittleEndian.PutUint16(temp, uint16(m))
	w.buf = append(w.buf, temp...)

	binary.LittleEndian.PutUint16(temp, uint16(w.n))
	w.buf = append(w.buf, temp...)

	w.buf = append(w.buf, byte(sizeOfStringTerminator))
	w.buf = append(w.buf, byte(firstStringTerminatorChar))
	w.buf = append(w.buf, byte(secondStringTerminatorChar))

	w.buf = append(w.buf, byte(sizeOfLineTerminator))
	w.buf = append(w.buf, byte(firstLineTerminatorChar))
	w.buf = append(w.buf, byte(secondLineTerminatorChar))

	w.buf = append(w.buf, make([]byte, 18, 18)...)

	w.buf = append(w.buf, make([]byte, 4*72, 4*72)...)

	w.buf = append(w.buf, w.note...)
}

func (w *writer) writeTraceBlockHeader(i int, data *traceDescriptorBlock) {
	binary.LittleEndian.PutUint32(w.buf[32+4*i:32+4*(i+1)], uint32(len(w.buf)))

	w.buf = append(w.buf, traceDescriptorBlockID...)

	temp := make([]byte, 4, 4)
	binary.LittleEndian.PutUint16(temp[:2], data.x)
	w.buf = append(w.buf, temp[:2]...)

	binary.LittleEndian.PutUint32(temp[:4], data.y)
	w.buf = append(w.buf, temp[:4]...)

	binary.LittleEndian.PutUint32(temp[:4], data.ns)
	w.buf = append(w.buf, temp[:4]...)

	w.buf = append(w.buf, byte(data.format))

	w.buf = append(w.buf, make([]byte, 19, 19)...)

	w.buf = append(w.buf, data.info...)
}
