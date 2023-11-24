package common

import (
	"github.com/gliderlabs/ssh"
)

type YajsSession struct {
	ssh.Session
	lastData []byte
	readCount int
	notify chan struct{}
}

func NewYajsSession(session *ssh.Session) *YajsSession {
	ys := &YajsSession{
		Session:*session,
		notify: make(chan struct{}),
	}
	return ys
}

func (ys *YajsSession) Read(b []byte) (n int, err error) {
	ys.readCount++
	defer func() {
		ys.readCount--
	}()

	if ys.readCount > 1{
		select {
			case <-ys.notify:
				i := len(ys.lastData)
				if len(b) < i {
					i = len(b)
				}
				if i > 0 {
					copy(b, ys.lastData)
					ys.lastData = ys.lastData[:0]
					return i, nil
				}else{
					return 0,nil
				}
		}
	}

	n,err = (*ys).Session.Read(b)
	if ys.readCount > 1 {
		if ys.lastData == nil || len(ys.lastData) == 0  {
			ys.lastData = make([]byte,n)
		}
		copy(ys.lastData, b)
		ys.notify <- struct{}{}
	}
	return n,err
}

