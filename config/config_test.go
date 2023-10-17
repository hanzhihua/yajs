package config

import (
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"testing"
	"yajs/utils"
)

func TestReadFrom(t *testing.T) {
	pwd,err := os.Getwd()
	if err != nil{
		t.Error(err)
	}
	tmpstr := pwd+"/testconf"
	ConfDir = &tmpstr
	err = Setup();
	if err != nil{
		t.Error(err)
	}

	if len(Instance.SshUsers) != 2{
		t.Errorf("the count of Instance.SshUsers is not right")
	}
	if Instance.sshUserMap[utils.RootUserkey] == nil || Instance.sshUserMap[utils.WebUserKey] == nil{
		t.Errorf("sshusers has error")
	}
	tmpstr = "saas01"
	server := Instance.GetServerByName(&tmpstr)
	if server.Port != 22{
		t.Logf("ignore,%s has error ,server port:%v",server.IP,server.Port)
	}

	tmpstr = "saas02"
	server = Instance.GetServerByName(&tmpstr)
	if server.Port != 23{
		t.Errorf("%s has error ,server port:%v",server.IP,server.Port)
	}
}

func TestCanAssessServer(t *testing.T) {
	pwd,err := os.Getwd()
	if err != nil{
		t.Errorf("has error:%v",err)
	}
	tmp := filepath.Join(pwd,"config.yaml")
	utils.Logger.Infof("%v",&tmp)
	ConfDir = &tmp
	err = Setup();
	if err != nil{
		t.Error(err)
	}
	b,err:=CanAssessServer("test1", "saas01")
	if err != nil{
		t.Error(err)
	}
	if !b{
		t.Errorf("CanAssessServer(\"test1\", \"saas01\") result:%v",b)
	}

	b,err =CanAssessServerWithSshuser("test1", "saas01","web")
	if err != nil{
		t.Error(err)
	}
	if !b{
		t.Errorf("CanAssessServerWithSshuser(\"test1\", \"saas01\",\"web\") result:%v",b)
	}

	b,err =CanAssessServerWithSshuser("test1", "saas01","root")
	if err != nil{
		t.Error(err)
	}
	if b{
		t.Errorf("CanAssessServerWithSshuser(\"test1\", \"saas01\",\"web\") result:%v",b)
	}



	b,err=CanAssessServer("test2", "saas01")
	if err != nil{
		t.Error(err)
	}
	if b{
		t.Errorf("CanAssessServer(\"test2\", \"saas01\") result:%v",b)
	}

	b,err =CanAssessServerWithSshuser("test2", "saas01","web")
	if err != nil{
		t.Error(err)
	}
	if b{
		t.Errorf("CanAssessServerWithSshuser(\"test2\", \"saas01\",\"web\") result:%v",b)
	}

	b,err =CanAssessServerWithSshuser("test2", "saas01","root")
	if err != nil{
		t.Error(err)
	}
	if b{
		t.Errorf("CanAssessServerWithSshuser(\"test2\", \"saas01\",\"web\") result:%v",b)
	}


	b,err=CanAssessServer("test2", "xyz")
	if err != nil{
		t.Error(err)
	}
	if !b{
		t.Errorf("CanAssessServer(\"xyz\", \"saas01\") result:%v",b)
	}

	b,err =CanAssessServerWithSshuser("test2", "xyz","web")
	if err != nil{
		t.Error(err)
	}
	if !b{
		t.Errorf("CanAssessServerWithSshuser(\"test2\", \"xyz\",\"web\") result:%v",b)
	}

	b,err =CanAssessServerWithSshuser("test2", "xyz","root")
	if err != nil{
		t.Error(err)
	}
	if !b{
		t.Errorf("CanAssessServerWithSshuser(\"test2\", \"saas01\",\"web\") result:%v",b)
	}

}

func TestPubKey(t *testing.T) {
	pwd,err := os.Getwd()
	if err != nil{
		t.Error(err)
	}
	tmpstr := pwd+"/testconf"
	ConfDir = &tmpstr
	err = Setup();
	if err != nil{
		t.Error(err)
	}
	username := "hanzhihua"
	pub := Instance.GetUserByUsername(&username).PublicKey
	allowed, _, _, _, _ := ssh.ParseAuthorizedKey( []byte(pub))
	bs,err := os.ReadFile(*ConfDir+"/test_rsa")
	if err != nil{
		t.Error(err)
	}
	singer,err := gossh.ParsePrivateKey(bs)
	if err != nil{
		t.Error(err)
	}
	b := ssh.KeysEqual(singer.PublicKey(),allowed)
	if !b{
		t.Error("has err,result false")
	}

}

