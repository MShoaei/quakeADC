package seg2

import (
	"bufio"
	"testing"
	"time"
)

func Test_writer_Write(t *testing.T) {
	type fields struct {
		dateTime time.Time
		n        int16
		note     string
		w        *bufio.Writer
	}
	type args struct {
		in0 [][]byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// s := writer{
			// 	dateTime: tt.fields.dateTime,
			// 	n:        tt.fields.n,
			// 	note:     tt.fields.note,
			// 	w:        tt.fields.w,
			// }
			// s.Write(tt.args.in0)
		})
	}
}
