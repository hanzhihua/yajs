package main

import (
	"flag"
	"fmt"
	"github.com/hanzhihua/yajs/core"
	"github.com/hanzhihua/yajs/utils"
	"io"
	"runtime"
	"strings"
)

var (
	BuildTags = "unknown"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = "unknown"
	showHelp  bool
	signal    string
)

func init() {
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.IntVar(&utils.Port, "p", 2222, "Port")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/github.com/hanzhihua/yajs/")
	flag.StringVar(&utils.ConfigDir, "cd", defaultConfigDir, "Config Directory")
	flag.IntVar(&utils.SshIdleTimeout, "ssh.idletimeout", 120, "Ssh idletimeout")
	flag.StringVar(&signal, "s", "", " send signal to a yajs process: reload")
	flag.Parse()
}

func main() {

	if showHelp {
		printReleaseString(flag.CommandLine.Output())
		flag.Usage()
		return
	}

	if strings.EqualFold(signal, "reload") {
		err := Reload()
		if err != nil {
			utils.Logger.Errorf("fail to send single for reload,err:%v", err)
		}
		return
	}

	lock, err := CreatePidFile()
	if err != nil {
		panic(err)
	}
	defer removePidFile(lock)

	core.Run()
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, git commit:%s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}

func printReleaseString(w io.Writer) {
	fmt.Fprintln(w, "Version:   "+BuildTags)
	fmt.Fprintln(w, "Built:     "+BuildTime)
	fmt.Fprintln(w, "GitCommit: "+GitCommit)
	fmt.Fprintln(w, "GoVersion: "+GoVersion)
}