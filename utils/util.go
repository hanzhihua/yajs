package utils

import (
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/fatih/color"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	NAME             = "yajs"
	SLOGAN           = "Yet another jump server"
	HOSTKEYFILE      = "yajs_hk"
	WebUserKey       = "web"
	RootUserkey      = "root"
	UpgradeFlag      = "__up"
	SshUserFlag      = "@@@"
	USER_KEY         = "user_key"
	SERVER_KEY       = "server_key"
	SSHUSER_KEY      = "sshuser_key"
	IDILTIMEOUT_KEY  = "idiltimeout_key"
	WRITER_KEY       = "writer_key"
	Default_SSH_PORT = 22
	SERVER_PREFIX    = "server:"
	MENU_PREFIX      = "menu:"
	TIME_LAYOUT      = "2006-01-02 15:04:05"
)

var (
	ConfigDir      string
	SshIdleTimeout int
	Port           int
)

var (
	ProcUser  *user.User
	blackList = []string{"rm", "mkfs", "mkfs.ext3", "make.ext2", "make.ext4", "make2fs", "shutdown", "reboot", "init", "dd"}
)

func init() {
	theUser, err := user.Current()
	if err != nil {
		fmt.Printf("fail to get system user, error is %v", err)
		os.Exit(2)
	}
	ProcUser = theUser
}

func getFilename(i uint8) string {
	_, file, _, ok := runtime.Caller(int(i))
	if !ok {
		file = "???"
	} else {
		file = filepath.Base(file)
	}
	return file
}

func printCurrentFileName(i uint8) {
	fmt.Printf("current go file name: %s \n ", getFilename(i))
}

func FilePath(path string) string {
	return strings.Replace(path, "~", ProcUser.HomeDir, 1)
}

func FileExited(path string) bool {
	info, err := os.Stat(FilePath(path))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsDirector(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func IsSaftCMDDefault(cmd string) bool {
	return IsSaftCMD(cmd, blackList)
}

func IsSaftCMD(cmd string, blacks []string) bool {
	lcmd := strings.ToLower(cmd)
	cmds := strings.Split(lcmd, " ")
	for _, ds := range cmds {
		for _, bk := range blacks {
			if ds == bk {
				return false
			}
		}
	}
	return true
}

func GetLocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips, nil
}

func ContainsStr(slice []string, element string) bool {
	if slice == nil {
		return false
	}

	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}

func PrintBanner(w io.Writer) {

	w.Write([]byte("\n"))
	templ := `{{ .AnsiColor.BrightMagenta }}{{ .Title "Yajs" "starwars" 0 }}{{ .AnsiColor.Default }}`
	banner.InitString(w, true, true, templ)
}

func PrintBannerWithUsername(w io.Writer, username string) {

	w.Write([]byte("\n"))
	templ := `{{ .AnsiColor.BrightMagenta }}{{ .Title "Yajs" "starwars" 0 }}{{ .AnsiColor.Default }}`
	banner.InitString(w, true, true, templ)
	color := color.New(color.FgMagenta)
	color.Fprint(w, fmt.Sprintf("\n当前登陆用户名: %s\n", username))
}

func PrintStackTrace() {
	s := string(debug.Stack())
	Logger.Errorf("exception stack:\n%s", s)
}



