package main

import (
	"flag"
	"fmt"
	"os"
	"yajs/core/server"
	"yajs/utils"
)


var (
	port          int
	showHelp      bool
	showVersion   bool
	configDir     string
	version       = "dev"
)

func init() {
	flag.IntVar(&port,"p", 2222, "Port")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	defaultConfigDir := fmt.Sprintf("%s%s", utils.ProcUser.HomeDir, "/yajs/")
	flag.StringVar(&configDir,"c",defaultConfigDir,"Config Directory")
	flag.Parse()
}

func main() {

	if showHelp {
		doHelp()
		os.Exit(2)
	} else if showVersion {
		fmt.Println(version)
		return
	}

	server.Run(configDir,port)
}

func doHelp() {
	fmt.Println("Usage: deving")
}
