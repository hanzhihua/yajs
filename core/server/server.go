package server

import (
	"errors"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"yajs/config"
	_ "yajs/config/aliyun"
	"yajs/core/helper"
	"yajs/core/ui"
	"yajs/utils"
)

var (
	ConfigDir string
	SshIdleTimeout int
	Port int
)

func Run(){
	trapSignals()
	config.ConfDir = &ConfigDir
	err := config.Setup()
	if err != nil{
		utils.Logger.Errorf("fail to start,err:%v",err)
		os.Exit(2)
	}
	addr := ":"+strconv.Itoa(Port)


	hkFile,err := config.Instance.GetHostKeyFile()
	if err != nil{
		utils.Logger.Error("fail to start,err:%v",err)
		os.Exit(2)
	}
	ssh.Handle(func(sess ssh.Session) {
		defer func() {
			if r := recover(); r != nil {
				s := string(debug.Stack())
				utils.Logger.Errorf("exception stack:\n%s",s)
				helper.GetWriter(&sess).WriteExist(true)
			}
		}()
		_,err := helper.NewWriter(&sess)
		if err != nil{
			utils.Logger.Panic(err)
		}
		printBanner(&sess)
		uiService := ui.UIService{Session:&sess}
		uiService.ShowUI()
		helper.GetWriter(&sess).WriteExist(false)
	})

	utils.Logger.Fatal(ssh.ListenAndServe(
		addr,
		nil,
		ssh.PublicKeyAuth(publickKeyAuth),
		//ssh.WrapConn(warpConn),
		ssh.HostKeyFile(hkFile),
		//ssh.ConnCallback()
		func(srv *ssh.Server) error {
			srv.IdleTimeout = time.Duration(SshIdleTimeout) * time.Second
			return nil
		},
	))
}


func publickKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	username,err := GetRealName(ctx)
	if err != nil{
		utils.Logger.Error(err)
		return false
	}

	user := config.Instance.GetUserByUsername(username)
	if user == nil{
		utils.Logger.Warningf("%s does not exist",*username)
		return false
	}

	if isLocal(ctx){
		utils.Logger.Warningf("test yajs in local")
		ctx.SetValue(utils.USER_KEY,user)
		return true
	}

	pubkey, _, _, _, err := ssh.ParseAuthorizedKey( []byte(user.PublicKeyContent))
	if err != nil{
		utils.Logger.Errorf("%s login fail, because occurs err:%v",username,err)
		return false
	}
	b := ssh.KeysEqual(key, pubkey)
	if b{
		utils.Logger.Warningf("%s login successful,remote_ip:%s,local_ip:%s",*username,ctx.Value(ssh.ContextKeyRemoteAddr),ctx.Value(ssh.ContextKeyLocalAddr))
		ctx.SetValue(utils.USER_KEY,user)
	}else{
		utils.Logger.Warningf("%s login failed,remote_ip:%s,local_ip:%s",*username,ctx.Value(ssh.ContextKeyRemoteAddr),ctx.Value(ssh.ContextKeyLocalAddr))
	}
	return b
}


func GetRealName(ctx ssh.Context) (*string,error){
	username := ctx.User()
	if len(strings.TrimSpace(username)) == 0{
		return nil,errors.New("username is blank")
	}

	if strings.Contains(username,utils.SshUserFlag){
		users :=strings.Split(username,utils.SshUserFlag)
		username = users[0]
		sshuser := users[1]
		ctx.SetValue(utils.SSHUSER_KEY,&sshuser)
		return &username,nil
	}else{
		return &username,nil
	}
}

func isLocal(ctx ssh.Context) bool{
	if addr, ok := ctx.Value(ssh.ContextKeyRemoteAddr).(*net.TCPAddr); ok {
		if addr.IP.IsLoopback(){
			return true
		}else{
			return false
		}
	}else{
		return false
	}
}

func printBanner(sess *ssh.Session){

	(*sess).Write([]byte("\n"))
	templ := `{{ .AnsiColor.BrightMagenta }}{{ .Title "Yajs" "starwars" 0 }}{{ .AnsiColor.Default }}`
	banner.InitString((*sess), true, true, templ)
	color := color.New(color.FgMagenta)
	color.Fprint((*sess), fmt.Sprintf("\n当前登陆用户名: %s\n", (*sess).User()))
}



//func warpConn(ctx ssh.Context, conn net.Conn) net.Conn{
//
//	return &yajsConn{conn}
//}
//
//type yajsConn struct {
//	net.Conn
//}
//
//func (yajsConn *yajsConn)Close() error{
//	s := string(debug.Stack())
//	utils.Logger.Errorf("print stack:\n%s",s)
//	utils.Logger.Warningf("yajs conn close")
//	return yajsConn.Conn.Close()
//
//}