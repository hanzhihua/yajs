package helper

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"
	"yajs/config"
	//"yajs/core/client"
	"yajs/utils"
)

func EnterJumpServer(sess *ssh.Session) error{
	shell := os.Getenv("SHELL")
	if shell == ""{
		shell = "/bin/bash"
	}
	cmd := exec.Command(shell,"--login")
	//cmd := exec.Command(shell)
	p, winCh, isPty := (*sess).Pty()
	if isPty {
		cmd.Env = append(cmd.Env,fmt.Sprintf("TERM=%s", p.Term))
		sshUser,err := config.Instance.GetJumpServerSshUser(sess)
		if err != nil{
			return err
		}
		if sshUser == nil || (*sshUser).IsRootUser(){
			cmd.Env = append(cmd.Env, "HOME=/root")
			cmd.Dir = "/root";
		}else{
			u, err := user.Lookup((*sshUser).Username)
			if err == nil{
				uid, _ := strconv.Atoi(u.Uid)
				gid, _ := strconv.Atoi(u.Gid)
				cmd.SysProcAttr = &syscall.SysProcAttr{}
				cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
				cmd.Dir = u.HomeDir
				cmd.Env = append(cmd.Env, "HOME="+u.HomeDir)
			}else{
				return err
			}
		}

		utils.Logger.Infof("cmd is %v",cmd)
		f, err := pty.StartWithSize(cmd,&pty.Winsize{Cols: uint16(p.Window.Width), Rows: uint16(p.Window.Height)})
		if err != nil {
			utils.Logger.Errorf("has error %v",err)
			return err
		}
		go func() {
			for win := range winCh {
				utils.Logger.Errorf("window changed :%v",win)
				//setWinsize(f, win.Width, win.Height)
			}
		}()
		writer := GetWriter(sess)
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

type Writer struct {
	frontSess *ssh.Session
	backSess *gossh.Session
	file *os.File
}

func NewWriter(sess *ssh.Session) (*Writer,error){

	writer := &Writer{
		frontSess: sess,
		backSess: nil,
	}
	logDir := fmt.Sprintf("%s%s",*config.ConfDir,"/logs")
	if !utils.IsDirector(logDir){
		os.MkdirAll(logDir, os.FileMode(0755))
	}

	timestr := time.Now().Format("20060102150405")
	path :=  fmt.Sprintf("%s%s_%s_%s%s", logDir, "/r",(*sess).User(),timestr,".log")
	var err error
	writer.file,err = os.Create(path)
	if err != nil{
		return nil,err
	}else{
		(*sess).Context().SetValue(utils.WRITER_KEY,writer)
		return writer,nil
	}
}

func GetWriter(sess *ssh.Session) *Writer {
	writer,ok:=(*sess).Context().Value(utils.WRITER_KEY).(*Writer)
	if ok{
		return writer
	}else{
		return nil
	}
}

func (w *Writer) BeginWrite(serverName string) (n int, err error) {
	timestr := time.Now().Format("20060102150405")
	content := fmt.Sprintf(serverName+" begin==========================%s===========================\n",timestr)
	return w.file.Write([]byte(content));
}

func (w *Writer) WriteEnd(serverName string) (n int, err error) {
	timestr := time.Now().Format("20060102150405")
	content := fmt.Sprintf(serverName+" end==========================%s===========================\n\n\n",timestr)
	return w.file.Write([]byte(content));
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.file.Write(p);
	return (*w.frontSess).Write(p)
}

func (w *Writer) Read(p []byte) (n int, err error) {
	n, err = (*w.frontSess).Read(p)
	return n, err
}
