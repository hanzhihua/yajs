package helper

import (
	"bytes"
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

var (
	// sz fmt.Sprintf("%+q", "rz\r**\x18B00000000000000\r\x8a\x11")
	//ZModemSZStart = []byte{13, 42, 42, 24, 66, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 13, 138, 17}
	ZModemSZStart = []byte{42, 42, 24, 66, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 13, 138, 17}
	// sz 结束 fmt.Sprintf("%+q", "\r**\x18B0800000000022d\r\x8a")
	//ZModemSZEnd = []byte{13, 42, 42, 24, 66, 48, 56, 48, 48, 48, 48, 48, 48, 48, 48, 48, 50, 50, 100, 13, 138}
	ZModemSZEnd = []byte{42, 42, 24, 66, 48, 56, 48, 48, 48, 48, 48, 48, 48, 48, 48, 50, 50, 100, 13, 138}
	// sz 结束后可能还会发送两个 OO，但是经过测试发现不一定每次都会发送 fmt.Sprintf("%+q", "OO")
	ZModemSZEndOO = []byte{79, 79}

	// rz fmt.Sprintf("%+q", "**\x18B0100000023be50\r\x8a\x11")
	ZModemRZStart = []byte{42, 42, 24, 66, 48, 49, 48, 48, 48, 48, 48, 48, 50, 51, 98, 101, 53, 48, 13, 138, 17}
	// rz -e fmt.Sprintf("%+q", "**\x18B0100000063f694\r\x8a\x11")
	ZModemRZEStart = []byte{42, 42, 24, 66, 48, 49, 48, 48, 48, 48, 48, 48, 54, 51, 102, 54, 57, 52, 13, 138, 17}
	// rz -S fmt.Sprintf("%+q", "**\x18B0100000223d832\r\x8a\x11")
	ZModemRZSStart = []byte{42, 42, 24, 66, 48, 49, 48, 48, 48, 48, 48, 50, 50, 51, 100, 56, 51, 50, 13, 138, 17}
	// rz -e -S fmt.Sprintf("%+q", "**\x18B010000026390f6\r\x8a\x11")
	ZModemRZESStart = []byte{42, 42, 24, 66, 48, 49, 48, 48, 48, 48, 48, 50, 54, 51, 57, 48, 102, 54, 13, 138, 17}
	// rz 结束 fmt.Sprintf("%+q", "**\x18B0800000000022d\r\x8a")
	ZModemRZEnd = []byte{42, 42, 24, 66, 48, 56, 48, 48, 48, 48, 48, 48, 48, 48, 48, 50, 50, 100, 13, 138}

	// **\x18B0
	ZModemRZCtrlStart = []byte{42, 42, 24, 66, 48}
	// \r\x8a\x11
	ZModemRZCtrlEnd1 = []byte{13, 138, 17}
	// \r\x8a
	ZModemRZCtrlEnd2 = []byte{13, 138}

	// zmodem 取消 \x18\x18\x18\x18\x18\x08\x08\x08\x08\x08
	ZModemCancel = []byte{24, 24, 24, 24, 24, 8, 8, 8, 8, 8}
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
	rz           bool
	sz           bool
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
		(*sess).Context().(ssh.Context).SetValue(utils.WRITER_KEY,writer)
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

	if !w.sz && !w.rz {
		if bytes.Contains(p,ZModemSZStart){
			w.sz = true
			w.file.Write([]byte("sz start\n"))
		}else if bytes.Contains(p,ZModemRZStart){
			w.rz = true
			w.file.Write([]byte("rz start\n"))
		}
	}

	if w.sz{
		if bytes.Contains(p,ZModemSZEnd){
			w.sz = false

			w.file.Write([]byte("sz end\n"))
		}else if bytes.Contains(p,ZModemCancel){
			w.sz = false
			w.file.Write([]byte("sz cancle\n"))
		}
	}else if w.rz{
		if bytes.Contains(p,ZModemRZEnd){
			w.rz = false
			w.file.Write(p);
			w.file.Write([]byte("rz end\n"))
		}else if bytes.Contains(p,ZModemCancel){
			w.rz = false
			w.file.Write(p);
			w.file.Write([]byte("rz cancle\n"))
		}
	}else{
		w.file.Write(p);
	}






	//s := string(p)
	//if !w.sz && !w.rz {
	//	hexData := hex.EncodeToString(p)
	//	if strings.Contains(hexData, "727a0d2a2a184230303030303030303030303030300d8a11") {
	//		w.sz = true
	//		w.file.Write([]byte("sz start\n"))
	//	} else if strings.Contains(hexData, "727a2077616974696e6720746f20726563656976652e2a2a184230313030303030303233626535300d8a11") {
	//		w.rz = true
	//		w.file.Write([]byte("rz start\n"))
	//	}
	//}
	//
	//if w.sz {
	//	// sz 会以 OO 结尾
	//	if "OO" == s {
	//		w.sz = false
	//		w.file.Write([]byte("sz finish\n"))
	//	}
	//} else if w.rz {
	//	// rz 最后会显示 Received /home/xxx
	//	w.file.Write([]byte("rz:\n"+s+"\n"))
	//	if strings.Contains(s, "Received") {
	//		w.rz = false
	//		w.file.Write(p)
	//		w.file.Write([]byte("rz finish\n"))
	//	}
	//} else {
	//	w.file.Write(p);
	//}


	//w.file.Write(p);
	return (*w.frontSess).Write(p)
}

func (w *Writer) Read(p []byte) (n int, err error) {
	n, err = (*w.frontSess).Read(p)
	return n, err
}
