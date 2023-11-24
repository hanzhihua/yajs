package server

import (
	"github.com/gliderlabs/ssh"
	"net"
)

func warpConn(ctx ssh.Context, conn net.Conn) net.Conn{

	return &yajsConn{conn}
}

type yajsConn struct {
	net.Conn
}

func (yajsConn *yajsConn)Close() error{
	//s := string(debug.Stack())
	//utils.Logger.Errorf("print stack:\n%s",s)
	//utils.Logger.Warningf("yajs conn close")
	return yajsConn.Conn.Close()

}

func (yajsConn *yajsConn)Read(b []byte) (n int, err error){
	//s := string(b)
	//utils.Logger.Warningf("yajsConn read:%s",s)
	return yajsConn.Conn.Read(b)
}


func (yajsConn *yajsConn)Write(b []byte) (n int, err error){
	//s, _ := utf8.DecodeRune(b)
	//utils.Logger.Warningf("yajsConn write:%s",s)
	return yajsConn.Conn.Write(b)
}

