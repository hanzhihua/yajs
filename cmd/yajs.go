package main

import (
	"flag"
	"fmt"
	"github.com/hanzhihua/yajs/core"
	"github.com/hanzhihua/yajs/utils"
	"runtime"
)


var (
	showHelp  bool
	GitCommit string
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.IntVar(&utils.Port,"p", 2222, "Port")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/github.com/hanzhihua/yajs/")
	flag.StringVar(&utils.ConfigDir,"cd",defaultConfigDir,"Config Directory")
	flag.IntVar(&utils.SshIdleTimeout,"ssh.idletimeout",120,"Ssh idletimeout")
	flag.Parse()
}

func main() {

	if showHelp {
		flag.Usage()
		fmt.Print(releaseString())
		return
	}

	core.Run()
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, git commit:%s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}

