package main

import (
	"bytes"
	"fmt"
	"io"
)

var (
	ErrBufferOverflow = fmt.Errorf("input buffer overflow")
)

type SizedBufferedReader struct {
	buf    []byte
	bufPos int
	bufLen int

	rd       io.Reader
	readSize int

	pendingErr error
}

const minReadSize = 16

func NewSizedBufReader(rd io.Reader, readSize int) *SizedBufferedReader {
	bufSize := readSize

	if readSize < minReadSize {
		readSize = minReadSize
	}
	if bufSize < readSize*2 {
		bufSize = readSize * 2
	}

	brd := &SizedBufferedReader{
		buf:      make([]byte, bufSize),
		rd:       rd,
		readSize: readSize,
	}

	return brd
}

func shift(arr []byte, src, dst, length int) int {
	// check for overlapped
	if src <= dst || length < 0 {
		return 0
	}
	// check for length overflow
	maxLen := len(arr) - src
	if length > maxLen {
		length = maxLen
	}
	dstSlice := arr[dst:]
	srcSlice := arr[src : src+length]
	copy(dstSlice, srcSlice)

	return length
}

func (brd *SizedBufferedReader) ReadString(delim byte) (string, error) {
	index := -1
	if brd.bufLen > 0 {
		index = bytes.IndexByte(brd.buf[brd.bufPos:brd.bufLen], delim)
	}
	if index < 0 && brd.pendingErr == nil {
		if brd.pendingErr == nil {
			if brd.bufLen+minReadSize >= len(brd.buf) {
				// shift buffer
				n := shift(brd.buf[0:brd.bufLen], brd.bufPos+index+1, 0, brd.bufLen-brd.bufPos)
				brd.bufPos = 0
				brd.bufLen = n
			}
			n, err := brd.rd.Read(brd.buf[brd.bufLen:])
			if err != nil {
				brd.pendingErr = err
			} else if n == 0 {
				// input buffer overflow
				brd.pendingErr = ErrBufferOverflow
			}
			brd.bufLen += n
		}
		index = bytes.IndexByte(brd.buf[brd.bufPos:brd.bufLen], delim)
	}
	if index == -1 {
		if brd.pendingErr == nil {
			_, brd.pendingErr = brd.rd.Read(brd.buf[brd.bufLen:brd.bufLen])
			if brd.pendingErr == nil {
				// detect EOF and input buffer overflow
				brd.pendingErr = ErrBufferOverflow
			}
		}
		s := string(brd.buf[brd.bufPos:brd.bufLen])
		if len(s) > 0 {
			brd.bufPos += len(s)
			return s, nil
		}
		return s, brd.pendingErr
	} else if brd.bufPos+index == brd.bufLen-1 {
		// delim on the end of buffer, so return pendingErr
		s := string(brd.buf[brd.bufPos:brd.bufLen])
		brd.bufPos = 0
		brd.bufLen = 0
		return s, brd.pendingErr
	} else {
		s := string(brd.buf[brd.bufPos : brd.bufPos+index+1])
		brd.bufPos += len(s)
		return s, nil
	}
}
