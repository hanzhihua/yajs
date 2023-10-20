package config

import (
	"github.com/casbin/casbin/v2"
	"strings"
	"yajs/utils"
)

func NewEnforcer() (*casbin.Enforcer,error){
	enforcer, err := casbin.NewEnforcer(*ConfDir+"/acl_model.conf", *ConfDir+"/acl_policy.csv")
	if err != nil{
		return nil,err
	}
	enforcer.AddFunction("rMatch",func(args ...interface{}) (interface{}, error){
		if len(args) != 2{
			utils.Logger.Panicf("args:%v len isn't two",args)
		}
		for _, p := range args {
			err, ok := p.(string)
			if !ok {
				utils.Logger.Panicf("has error:%v",err)
			}
		}
		left := strings.Split(args[0].(string),"|")
		right := strings.Split(args[1].(string),"|")

		if len(right) == 1 && right[0] == "*"{
			return true,nil
		}
		if len(left) == 1 && len(right) == 1{
			server := left[0]
			policyServer := right[0]
			return internalMatch(server,policyServer),nil
		}else if len(left) == 2 && len(right) == 2{
			server := left[0]
			sshuser := left[1]
			policyServer := right[0]
			policySshuser := right[1]

			return internalMatch(server,policyServer) && internalMatch(sshuser,policySshuser),nil
		}else{
			utils.Logger.Debugf("left:%v,right:%v dose not match",left,right)
			return false, nil
		}
	})
	return enforcer,nil
}


func internalMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}