package main

import (
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"runtime"
	"strings"
	"syscall"
	"yajs/core/server"
	"yajs/utils"
)


var (
	showHelp  bool
	GitCommit string
	signal string
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.IntVar(&server.Port,"p", 2222, "Port")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/yajs/")
	flag.StringVar(&server.ConfigDir,"c",defaultConfigDir,"Config Directory")
	flag.StringVar(&signal,"s","","")
	flag.IntVar(&server.SshIdleTimeout,"ssh.idletimeout",120,"Ssh idletimeout")
	flag.Parse()
}

func main() {

	if showHelp {
		flag.Usage()
		fmt.Print(releaseString())
		return
	}



		if strings.EqualFold(signal,"reload"){
			sendSignal(utils.NAME)
			return
		}


	server.Run()
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, git commit:%s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}

func sendSignal(name string) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}
	for _, p := range processes {
		n, err := p.Name()
		utils.Logger.Infof("pid:%d,%s\n",p.Pid,n)
		if err != nil {
			return err
		}
		if n == name {
			utils.Logger.Infof("process:%v",p)
			return p.SendSignal(syscall.SIGUSR2)
		}
	}
	return fmt.Errorf("process not found")
}