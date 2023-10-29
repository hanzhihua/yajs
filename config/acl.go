package config

import (
	"github.com/casbin/casbin/v2"
	"github.com/hanzhihua/yajs/utils"
	"strings"
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

func CanAssessMenu(user,menuId string) (bool,error){
	return enforcer.Enforce(user,withMenuPrefix(menuId))
}

func CanAssessServer(user,server string) (bool,error){
	return enforcer.Enforce(user,withServerPrefix(server))
}

func CanAssessServerWithSshuser(user,server,sshuser string) (bool,error){
	b,err := enforcer.Enforce(user,withServerPrefix(server))
	if err != nil{
		utils.Logger.Errorf("CanAssessServerWithSshuser user:%s,server:%s,err:%v",user,server,err)
		return false,err
	}
	if !b{
		utils.Logger.Warningf("CanAssessServerWithSshuser user:%s,server:%s,result:%v",user,server,b)
		return false,nil
	}

	resource := withServerPrefix(server+"|"+sshuser)
	b,err = enforcer.Enforce(user,resource)
	if err != nil{
		utils.Logger.Errorf("CanAssessServerWithSshuser user:%s,server:%s,err:%v",user,resource,err)
		return false,err
	}
	if !b{
		utils.Logger.Warningf("CanAssessServerWithSshuser user:%s,server:%s,result:%v",user,resource,b)
		return false,nil
	}
	return true,nil
}

func withServerPrefix(reousrce string) string{
	return utils.SERVER_PREFIX+reousrce
}

func withMenuPrefix(reousrce string) string{
	return utils.MENU_PREFIX+reousrce
}
