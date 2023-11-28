package jumpserver

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/hanzhihua/yajs/config"
	"github.com/hanzhihua/yajs/core/common"
	"github.com/hanzhihua/yajs/utils"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func Enter(sess *common.YajsSession) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	cmd := exec.Command(shell, "--login")
	//cmd := exec.Command(shell)
	p, winCh, isPty := (*sess).Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", p.Term))
		sshUser, err := config.Instance.GetJumpServerSshUser(&sess.Session)
		if err != nil {
			return err
		}
		if sshUser == nil || (*sshUser).IsRootUser() {
			cmd.Env = append(cmd.Env, "HOME=/root")
			cmd.Dir = "/root"
		} else {
			u, err := user.Lookup((*sshUser).Username)
			if err == nil {
				uid, _ := strconv.Atoi(u.Uid)
				gid, _ := strconv.Atoi(u.Gid)
				cmd.SysProcAttr = &syscall.SysProcAttr{}
				cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
				cmd.Dir = u.HomeDir
				cmd.Env = append(cmd.Env, "HOME="+u.HomeDir)
			} else {
				return err
			}
		}

		utils.Logger.Infof("cmd is %v", cmd)
		f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(p.Window.Width), Rows: uint16(p.Window.Height)})
		if err != nil {
			utils.Logger.Errorf("has error %v", err)
			return err
		}
		go func() {
			for win := range winCh {
				utils.Logger.Errorf("window changed :%v", win)
				//setWinsize(f, win.Width, win.Height)
			}
		}()
		writer := common.GetAduitIO(sess)
		writer.BeginWrite("jumpserver")
		defer func() {
			writer.WriteEnd("jumpserver")
			_ = f.Close()
		}()
		go func() {
			io.Copy(f, writer)
		}()
		io.Copy(writer, f)
		cmd.Wait()
	} else {
		io.WriteString(*sess, "No PTY requested.\n")
	}

	return nil
}
