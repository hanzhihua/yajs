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
	version   bool
	configDir     string
)

var (
	// nolint
	appVersion = "(untracked dev build)" // inferred at startup
	devBuild   = true                    // inferred at startup

	buildDate        string // date -u
	gitTag           string // git describe --exact-match HEAD 2> /dev/null
	gitNearestTag    string // git describe --abbrev=0 --tags HEAD
	gitShortStat     string // git diff-index --shortstat
	gitFilesModified string // git diff-index --name-only HEAD

	// Gitcommit contains the commit where we built CoreDNS from.
	GitCommit string
)

func init() {
	flag.IntVar(&port,"p", 2222, "Port")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&version, "version", false, "Show version")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/yajs/")
	flag.StringVar(&configDir,"c",defaultConfigDir,"Config Directory")
	flag.IntVar(&server.SshIdleTimeout,"ssh.idletimeout",120,"Ssh idletimeout")
	flag.Parse()
}

func main() {

	if showHelp {
		doHelp()
		os.Exit(2)
	} else if version {
		fmt.Print(releaseString())
		return
	}

	server.Run(configDir,port)
}

func doHelp() {
	fmt.Println("Usage: deving")
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}