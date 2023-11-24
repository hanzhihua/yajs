package common

import (
	"io"
)

type InteruptReader struct {
	r        io.Reader
	stop   chan struct{}
}

func NewInteruptReader(r io.Reader) *InteruptReader {
	return &InteruptReader{
		r:r,
		stop:   make(chan struct{}),
	}
}

func (r *InteruptReader) Read(p []byte) (n int, err error) {
	if r.r == nil {
		return 0, io.EOF
	}
	select {
	case <-r.stop:
		return 0, io.EOF
	default:
		return r.r.Read(p)

	}
}

func (r *InteruptReader) Close(){
	close(r.stop)
}