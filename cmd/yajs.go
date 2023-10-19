package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"yajs/core/server"
	"yajs/utils"
)


var (
	port          int
	showHelp      bool
	configDir     string
	GitCommit string
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.IntVar(&port,"p", 2222, "Port")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/yajs/")
	flag.StringVar(&configDir,"c",defaultConfigDir,"Config Directory")
	flag.IntVar(&server.SshIdleTimeout,"ssh.idletimeout",120,"Ssh idletimeout")
	flag.Parse()
}

func main() {

	if showHelp {
		flag.Usage()
		fmt.Print(releaseString())
		os.Exit(2)
	}

	server.Run(configDir,port)
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, git commit:%s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}