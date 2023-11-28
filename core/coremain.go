package core

import (
	"github.com/hanzhihua/yajs/config"
	_ "github.com/hanzhihua/yajs/config/aliyun"
	"github.com/hanzhihua/yajs/core/server"
	"github.com/hanzhihua/yajs/utils"
	"os"
)

func Run() {

	trapSignals()

	config.ConfDir = &utils.ConfigDir
	err := config.Setup()
	if err != nil {
		utils.Logger.Errorf("fail to start,err:%v", err)
		os.Exit(2)
	}
	utils.PrintBanner(os.Stdout)
	server.Run()

}
