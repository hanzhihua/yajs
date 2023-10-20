package main

import (
	"flag"
	"fmt"
	"runtime"
	"yajs/core/server"
	"yajs/utils"
)


var (
	showHelp  bool
	GitCommit string
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.IntVar(&server.Port,"p", 2222, "Port")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/yajs/")
	flag.StringVar(&server.ConfigDir,"c",defaultConfigDir,"Config Directory")
	flag.IntVar(&server.SshIdleTimeout,"ssh.idletimeout",120,"Ssh idletimeout")
	flag.Parse()
}

func main() {

	if showHelp {
		flag.Usage()
		fmt.Print(releaseString())
		return
	}

	server.Run()
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, git commit:%s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}