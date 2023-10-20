package config

import (
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/gliderlabs/ssh"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
	"sync"
	"yajs/utils"
)

var (
	Instance *Config
	ConfDir  *string
	rwMutex  sync.RWMutex
	enforcer *casbin.Enforcer
	serverProviders []*ServerProvider
)

func Setup() error {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	if Instance != nil{
		utils.Logger.Warningf("setup has already executed")
		return nil
	}
	tmpConf,err := readFromDefault()
	if err != nil{
		return err
	}
	enforcer, err = NewEnforcer()
	if err != nil{
		return err
	}
	Instance = tmpConf
	return nil
}

func GetConfigFile() (string,error){
	fileName := "config.yaml"

	str := os.Getenv("CONFIG_FILENAME")
	if str != ""{
		fileName = str
	}

	file := *ConfDir+"/"+fileName
	fileInfo,err := os.Stat(file)

	if err != nil{
		return "",err
	}

	if fileInfo.IsDir(){
		return "",errors.New(fmt.Sprintf("%s is a directory",file))
	}
	return file,nil
}

func GetPubKeyContent(username string) (string,error){
	file := *ConfDir+"/pubs/"+username+".pub"
	fileInfo,err := os.Stat(file)

	if err != nil{
		return "",err
	}

	if fileInfo.IsDir(){
		return "",errors.New(fmt.Sprintf("%s is a directory",file))
	}

	bs,err := os.ReadFile(file)
	if err != nil{
		return "",err
	}
	return string(bs),nil
}

func GetPrivateKeyContent(sshuserName string)(string,error){
	file := *ConfDir+"/pris/"+sshuserName+"_rsa"
	fileInfo,err := os.Stat(file)

	if err != nil{
		return "",err
	}

	if fileInfo.IsDir(){
		return "",errors.New(fmt.Sprintf("%s is a directory",file))
	}

	bs,err := os.ReadFile(file)
	if err != nil{
		return "",err
	}
	return string(bs),nil
}


type Config struct {
	Users   []*User   `yaml:"users"`
	Servers []*Server `yaml:"servers"`
	SshUsers []*SSHUser `yaml:"sshusers"`
	ServerProvider *string `yaml:"server_provider"`
	sshUserMap map[string]*SSHUser
}

type User struct {
	Username         string `yaml:"username"`
	PublicKeyContent string
}

type Server struct {
	Name string `yaml:"name"`
	IP   string `yaml:"ip"`
	Port int   `default:"22" yaml:"port"`

}

type SSHUser struct {
	Username          string `yaml:"username"`
	PrivateKeyContent string
}

func readFromDefault() (*Config,error) {
	if ConfDir ==nil{
		utils.Logger.Panicf("conf path is nil")
	}
	confFile,err := GetConfigFile()
	if err != nil{
		return nil,err
	}
	return readFrom(confFile)
}

func readFrom(path string) (*Config,error) {
	configBytes, err := os.ReadFile(utils.FilePath(path))
	if err != nil {
		utils.Logger.Errorf("Error reading YAML file: %s\n", err)
		return nil,err
	}
	tmpConf := &Config{sshUserMap:map[string]*SSHUser{}}
	err = yaml.Unmarshal([]byte(configBytes), tmpConf)
	if err != nil {
		utils.Logger.Warningf("Error parsing YAML file: %s\n", err)
		return nil,err
	}
	for _,user := range tmpConf.Users{
		user.PublicKeyContent,err = GetPubKeyContent(user.Username)
		if err != nil{
			return nil,err
		}
	}
	for _,sshuser := range tmpConf.SshUsers{
		sshuser.PrivateKeyContent,err = GetPrivateKeyContent(sshuser.Username)
		if err != nil{
			return nil,err
		}
	}

	for _,sshUser  := range tmpConf.SshUsers{
		tmpConf.sshUserMap[sshUser.Username] = sshUser
	}
	if tmpConf.Servers == nil && tmpConf.ServerProvider != nil{
		for _,provider := range serverProviders{
			if (*provider).Name() == *tmpConf.ServerProvider{
				tmpConf.Servers,err = (*provider).GetAllServer()
				if err != nil{
					return nil,err
				}
			}
		}
	}

	return tmpConf,nil
}

func Reload() error{
	rwMutex.Lock()
	defer rwMutex.Unlock()
	utils.Logger.Warningf("reload config")
	tmpConf,err := readFromDefault()
	if err != nil{
		utils.Logger.Errorf("reloading has error:%v",err)
		return err
	}

	tmpEnforcer, err := NewEnforcer()

	if err != nil{
		utils.Logger.Errorf("reloading has error:%v",err)
		return err
	}
	Instance = tmpConf
	enforcer = tmpEnforcer
	utils.Logger.Infof("reload config:%v finish", Instance)
	return nil
}

func (config *Config) GetUserByUsername(username *string) *User {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	if username == nil || *username == ""{
		utils.Logger.Warningf("username is nil at GetUserByUsername")
		return nil
	}
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	var tmpstr = strings.TrimSpace(*username)
	for _, user := range config.Users {
		if user.Username == tmpstr {
			return user
		}
	}
	utils.Logger.Infof("%s doesn't exist",*username)
	return nil
}

func  (config *Config) GetServerByName(name *string) *Server {
	if name == nil || *name == ""{
		utils.Logger.Warningf("name is nil at GetServerByName")
		return nil
	}
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	var tmpstr = strings.TrimSpace(*name)
	for _, server := range Instance.Servers {
		if server.Name == tmpstr {
			return server
		}
	}
	utils.Logger.Infof("%s doesn't exist",name)
	return nil
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

func (*Config) GetHostKeyFile() (string,error){
	file := *ConfDir+"/"+"yajs_hk"
	fileInfo,err := os.Stat(file)

	if err != nil{
		return "",err
	}

	if fileInfo.IsDir(){
		return "",errors.New(fmt.Sprintf("%s is a directory",file))
	}
	return file,nil
}

func (config *Config) GetSSHUser(username *string) *SSHUser{
	if username == nil{
		return nil
	}
	return config.sshUserMap[*username]
}

func (config *Config) GetJumpServerSshUser(session *ssh.Session) (*SSHUser,error){
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	ctx := (*session).Context()

	sshuserName,ok := ctx.Value(utils.SSHUSER_KEY).(*string)
	if !ok{
		return nil,nil
	}else{
		sshuser,ok := config.sshUserMap[*sshuserName]
		if ok{
			return sshuser,nil
		}else{
			return nil,errors.New(fmt.Sprintf("sshuser:%s does not exist",*sshuserName))
		}
	}
}

func (config *Config) GetSshUser(session *ssh.Session,server string) (*SSHUser,error){
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	ctx := (*session).Context()

	sshuserName,ok := ctx.Value(utils.SSHUSER_KEY).(*string)
	if ok{
		sshuser,ok := config.sshUserMap[*sshuserName]
		if ok{
			return sshuser,nil
		}else{
			return nil,errors.New(fmt.Sprintf("sshuser:%s do not exist",*sshuserName))
		}
	}

	user := ctx.Value(utils.USER_KEY).(*User)

	for _,sshuser := range config.SshUsers{
		b,err := CanAssessServerWithSshuser(user.Username,server,sshuser.Username)
		if err != nil{
			utils.Logger.Errorf("enforcer.Enforce(%s,%s,%s) has error:%v,",user.Username,server,sshuser.Username,err)
		}else{
			if b{
				return sshuser,nil
			}else{
				utils.Logger.Errorf("enforcer.Enforce(%s,%s,%s) result:false,",user.Username,server,sshuser.Username)
			}
		}
	}
	return nil,errors.New("no right to access the server")
}

type ServerProvider interface {
	GetAllServer() ([]*Server,error)
	Name() string
}

func AddServerProvider(provider *ServerProvider){
	serverProviders = append(serverProviders,provider)
}

func (sshUser *SSHUser) IsRootUser()  bool{
	if sshUser == nil{
		return false
	}
	if strings.EqualFold(sshUser.Username,"root"){
		return true
	}else{
		return false
	}
}