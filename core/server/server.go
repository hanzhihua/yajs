package server

import (
	"errors"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/hanzhihua/yajs/config"
	_ "github.com/hanzhihua/yajs/config/aliyun"
	"github.com/hanzhihua/yajs/core/common"
	"github.com/hanzhihua/yajs/core/ui"
	"github.com/hanzhihua/yajs/utils"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	gossh "golang.org/x/crypto/ssh"
	"unsafe"
)


func Run(){

	addr := ":"+strconv.Itoa(utils.Port)
	hkFile,err := config.Instance.GetHostKeyFile()
	if err != nil{
		utils.Logger.Error("fail to start,err:%v",err)
		os.Exit(2)
	}
	ssh.Handle(func(sess ssh.Session) {
		defer func() {
			sess.Exit(0)
			if r := recover(); r != nil {
				utils.PrintStackTrace()
				common.GetWriter(&sess).WriteExist(true)
			}else{
				common.GetWriter(&sess).WriteExist(false)
			}
		}()

		_,err := common.NewWriter(&sess)
		if err != nil{
			utils.Logger.Errorf("it occur error :%v, when generated aduit log ",err)
			return
		}

		changeIdleTimeout(sess)

		utils.PrintBannerWithUsername(sess,sess.User())
		uiService := ui.UIService{Session:&sess}
		uiService.ShowUI()
	})

	utils.Logger.Fatal(ssh.ListenAndServe(
		addr,
		nil,
		ssh.PublicKeyAuth(publickKeyAuth),
		//ssh.WrapConn(warpConn),
		ssh.HostKeyFile(hkFile),
		//ssh.ConnCallback()
		func(srv *ssh.Server) error {
			srv.IdleTimeout = time.Duration(utils.SshIdleTimeout) * time.Second
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
		if sshuser != ""{
			ctx.SetValue(utils.SSHUSER_KEY,&sshuser)
		}

		if len(users) == 3{
			idle_t,err := strconv.Atoi(users[2])
			if err == nil{
				ctx.SetValue(utils.IDILTIMEOUT_KEY,idle_t)
			}else{
				utils.Logger.Warningf("%s is not integer",users[2])
			}
		}
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

func GetUnexportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

func SetUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
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

func changeIdleTimeout(sess ssh.Session){

	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Warningf("r:%v",r)
		}
	}()

	//idleTimeout := getIdleTimeout(sess)
	//if idleTimeout == nil{
	//	return
	//}
	//
	//idleTimeoutInt,err := strconv.Atoi(*idleTimeout)
	//if err != nil{
	//	utils.Logger.Warningf("idleTimeout %s,err:%v",*idleTimeout,err)
	//}

	ctx := sess.Context()

	idleTimeoutInt,ok := ctx.Value(utils.IDILTIMEOUT_KEY).(int)
	if !ok{
		return
	}

	serverConn,ok := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	sess.Environ()
	if ok{
		value := reflect.ValueOf(serverConn.Conn)
		value = value.Elem()
		value = value.FieldByName("sshConn")
		value = value.FieldByName("conn")
		value = GetUnexportedField(value)
		v := value.Interface()
		value = reflect.ValueOf(v).Elem()
		value = value.FieldByName("idleTimeout")
		var d = (*time.Duration)(unsafe.Pointer(value.UnsafeAddr()))
		*d = (time.Duration(idleTimeoutInt) * time.Second)
	}else{
		utils.Logger.Warningf("conn is not server conn,%v",serverConn)
	}
}

//func getIdleTimeout(sess ssh.Session) *string{
//	for _,envItem :=  range sess.Environ(){
//		strs := strings.Split(envItem,"=")
//		if len(strs) == 2{
//			if strs[0] == "idle_timeout"{
//				return &strs[1]
//			}
//		}
//	}
//	return nil
//}