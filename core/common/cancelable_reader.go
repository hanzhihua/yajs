package common

import (
	"io"
	"sync"
	"sync/atomic"
)

type CancelableStdin struct {
	r      io.Reader
	mutex  sync.Mutex
	stop   chan struct{}
	closed int32
	notify chan struct{}
	data   []byte
	read   int
	err    error
}

func NewCancelableStdin(r io.Reader) *CancelableStdin {
	c := &CancelableStdin{
		r:      r,
		notify: make(chan struct{}),
		stop:   make(chan struct{}),
	}
	go c.ioloop()
	return c
}

func (c *CancelableStdin) ioloop() {
loop:
	for {
		select {
		case <-c.notify:
			c.read, c.err = c.r.Read(c.data)
			select {
			case c.notify <- struct{}{}:
			case <-c.stop:
				break loop
			}
		case <-c.stop:
			break loop
		}
	}
}

func (c *CancelableStdin) Read(b []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if atomic.LoadInt32(&c.closed) == 1 {
		return 0, io.EOF
	}

	c.data = b
	select {
	case c.notify <- struct{}{}:
	case <-c.stop:
		return 0, io.EOF
	}
	select {
	case <-c.notify:
		return c.read, c.err
	case <-c.stop:
		return 0, io.EOF
	}
}

func (c *CancelableStdin) Close(){
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		close(c.stop)
	}
}

func (c *CancelableStdin) LastStr() string{
	return string(c.data)
}
