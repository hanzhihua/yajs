package client

import (
	"errors"
	"fmt"
	"github.com/gliderlabs/ssh"
	"net"
	"strings"
	"yajs/config"
	"yajs/core/common"
	"yajs/utils"

	"github.com/fatih/color"
	gossh "golang.org/x/crypto/ssh"
)

func NewTerminal(server *config.Server, sshUser *config.SSHUser, sess *ssh.Session) error {
	upstreamClient, err := NewSSHClient(server, sshUser)
	if err != nil {
		return err
	}

	upstreamSess, err := upstreamClient.NewSession()
	if err != nil {
		return nil
	}
	defer upstreamSess.Close()

	writer := common.GetWriter(sess)
	if err != nil{
		return err
	}
	upstreamSess.Stdout = writer
	upstreamSess.Stdin = writer
	upstreamSess.Stderr = writer

	pty, winCh, _ := (*sess).Pty()

	utils.Logger.Warningf("pty term:%v,window:%v,",pty.Term, pty.Window)

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

	writer.BeginWrite(server.Name)
	defer func() {
		writer.WriteEnd(server.Name)
	}()
	go func () {
		for win := range winCh {
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
	utils.Logger.Infof("dial tcp address:%s",addr)
	client, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ParseRawCommand ParseRawCommand
func ParseRawCommand(command string) (string, []string, error) {
	parts := strings.Split(command, " ")

	if len(parts) < 1 {
		return "", nil, errors.New("No command in payload: " + command)
	}

	if len(parts) < 2 {
		return parts[0], []string{}, nil
	}

	return parts[0], parts[1:], nil
}

// ErrorInfo ErrorInfo
func ErrorInfo(err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s\n", err))
}

// Info Info
func Info(msg string, sess *ssh.Session) {
	green := color.New(color.FgGreen)
	green.Fprint(*sess, msg)
}
