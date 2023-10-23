package core

import (
	"os"
	"yajs/config"
	_ "yajs/config/aliyun"
	"yajs/core/server"
	"yajs/utils"
)


func Run(){

	trapSignals()

	config.ConfDir = &utils.ConfigDir
	err := config.Setup()
	if err != nil{
		utils.Logger.Errorf("fail to start,err:%v",err)
		os.Exit(2)
	}
	utils.PrintBanner(os.Stdout)
	server.Run()

}