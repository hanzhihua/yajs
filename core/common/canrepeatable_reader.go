package common

import (
	"io"
)

type CanrepeatableReader struct {
	io.Reader
	repeat bool
	lastData []byte
}

func NewCanrepeatableReader(r io.Reader) *CanrepeatableReader {
	c := &CanrepeatableReader{
		Reader:r,
	}
	return c
}

func(c *CanrepeatableReader)SetRepeat(){
	c.repeat = true
}

func (c *CanrepeatableReader) Read(b []byte) (n int, err error) {
	if c.repeat{
		copy(b, c.lastData)
		c.repeat = false
		return len(b),nil
	}
	n,err = c.Reader.Read(b)
	copy(c.lastData,b)
	return n,err
}