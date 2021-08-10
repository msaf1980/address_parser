package main

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_shift(t *testing.T) {
	tests := []struct {
		name string

		arr []byte
		src int
		dst int
		len int

		result []byte
		want   int
	}{
		{
			name: "shift must success",
			arr:  []byte{0, 1, 2, 3, 4, 5, 6},
			src:  3,
			dst:  1,
			len:  3,

			result: []byte{0, 3, 4, 5, 4, 5, 6},
			want:   3,
		},
		{
			name: "shift must success (length truncated)",
			arr:  []byte{0, 1, 2, 3, 4, 5, 6},
			src:  3,
			dst:  1,
			len:  5,

			result: []byte{0, 3, 4, 5, 6, 5, 6},
			want:   4,
		},
		{
			name: "shift zero len",
			arr:  []byte{0, 1, 2, 3, 4, 5, 6},
			src:  3,
			dst:  1,
			len:  0,

			result: []byte{0, 1, 2, 3, 4, 5, 6},
			want:   0,
		},
		{
			name: "shift (src == dst)",
			arr:  []byte{0, 1, 2, 3, 4, 5, 6},
			src:  1,
			dst:  1,
			len:  0,

			result: []byte{0, 1, 2, 3, 4, 5, 6},
			want:   0,
		},
		{
			name: "shift (src < dst)",
			arr:  []byte{0, 1, 2, 3, 4, 5, 6},
			src:  1,
			dst:  2,
			len:  0,

			result: []byte{0, 1, 2, 3, 4, 5, 6},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shift(tt.arr, tt.src, tt.dst, tt.len)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBufferedReader_ReadString(t *testing.T) {
	tests := []struct {
		name string

		readSize int

		src string

		wantResult []string
		wantErr    error
	}{
		{
			name:       "one read to buffer",
			src:        "Test_string__ ed",
			readSize:   16,
			wantResult: []string{"Test_", "string_", "_", " ed"},
			wantErr:    io.EOF,
		},
		{
			name:       "two reads to buffer",
			src:        "Test_string__ ended 2 fill",
			readSize:   16,
			wantResult: []string{"Test_", "string_", "_", " ended 2 fill"},
			wantErr:    io.EOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := strings.NewReader(tt.src)
			brd := NewSizedBufReader(rd, tt.readSize)
			var result []string
			var err error
			for {
				var s string
				if s, err = brd.ReadString('_'); err != nil {
					break
				} else {
					result = append(result, s)
				}
			}
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
