package aliyun

import (
	"github.com/hanzhihua/yajs/config"
)

type AliyunServerProvider struct {}

var instance config.ServerProvider

func init(){
	instance = AliyunServerProvider{}
	config.AddServerProvider(&instance)
}

func (instance AliyunServerProvider) GetAllServer() ([]*config.Server,error){
	return getAllServer()
}


func (instance AliyunServerProvider) Name() string{
	return "aliyun"
}

