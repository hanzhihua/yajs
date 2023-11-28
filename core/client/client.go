package client

import (
	"fmt"
	"github.com/hanzhihua/yajs/config"
	"github.com/hanzhihua/yajs/core/common"
	"github.com/hanzhihua/yajs/utils"
	gossh "golang.org/x/crypto/ssh"
	"net"
)

func NewTerminal(server *config.Server, sshUser *config.SSHUser, sess *common.YajsSession) error {
	upstreamClient, err := NewSSHClient(server, sshUser)
	if err != nil {
		return err
	}

	upstreamSess, err := upstreamClient.NewSession()
	if err != nil {
		return nil
	}
	defer upstreamSess.Close()

	aduitIO := common.GetAduitIO(sess)
	if err != nil {
		return err
	}

	upstreamSess.Stdout = aduitIO
	upstreamSess.Stdin = aduitIO
	upstreamSess.Stderr = aduitIO

	pty, winCh, _ := (*sess).Pty()

	utils.Logger.Warningf("pty term:%v,window:%v,", pty.Term, pty.Window)

	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}

	if err := upstreamSess.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, modes); err != nil {
		return err
	}

	if err := upstreamSess.Shell(); err != nil {
		return err
	}

	aduitIO.BeginWrite(server.Name)
	defer func() {
		aduitIO.WriteEnd(server.Name)
	}()
	go func() {
		for win := range winCh {
			utils.Logger.Warningf("win change,height:%d,width:%d", win.Height, win.Width)
			upstreamSess.WindowChange(win.Height, win.Width)
		}
	}()

	if err := upstreamSess.Wait(); err != nil {
		return err
	}
	return nil
}

func NewSSHClient(server *config.Server, sshUser *config.SSHUser) (*gossh.Client, error) {

	signer, err := gossh.ParsePrivateKey([]byte(sshUser.PrivateKeyContent))
	if err != nil {
		utils.Logger.Error(err)
		return nil, err
	}

	config := &gossh.ClientConfig{
		User: sshUser.Username,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }),
	}

	addr := fmt.Sprintf("%s:%d", server.IP, server.Port)
	utils.Logger.Infof("dial tcp address:%s", addr)
	client, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
