package cmd

import (
	"io"
	"os"
	"testing"
)

func Test_convert(t *testing.T) {
	type args struct {
		reader1  io.Reader
		reader2  io.Reader
		reader3  io.Reader
		writer   io.WriteCloser
		size     int64
		channels [24]bool
	}
	i, _ := os.Open("../testdata/data1.raw")
	o, _ := os.Create("../testdata/data1.conv")
	stat, _ := i.Stat()

	tests := []struct {
		name string
		args args
	}{
		{name: "test 1", args: args{
			reader1: i,
			reader2: nil,
			reader3: nil,
			writer:  o,
			size:    stat.Size(),
			channels: [24]bool{
				true, false, false, false, true, true, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convert(tt.args.reader1, tt.args.reader2, tt.args.reader3, tt.args.writer, tt.args.size, tt.args.channels)
		})
	}
}
