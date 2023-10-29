package batchcmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/hanzhihua/yajs/config"
	"github.com/hanzhihua/yajs/core/client"
	"github.com/hanzhihua/yajs/utils"
	"os"
	"strings"
)

func BatchRunCMD(sess *ssh.Session,cmdFile string) error{
	readFile, err := os.Open(cmdFile)
	if err != nil {
		return err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string
	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	if len(fileLines) != 2 {
		return errors.New(fmt.Sprintf("the line of %s content is not two or three lines",readFile))
	}
	hosts := strings.Split(fileLines[0],",")
	cmd := fileLines[1]
	if !utils.IsSaftCMDDefault(cmd){
		return errors.New("no saft cmd")
	}
	for _,host := range hosts{
		err = runCMD(sess,host,cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCMD(sess *ssh.Session,host,cmd string) error{
	server := config.Instance.GetServerByName(&host)
	if server == nil{
		return errors.New(fmt.Sprintf("%s is not a valid host",host))
	}
	sshuer,err := config.Instance.GetSshUser(sess,host)
	if err != nil{
		return err
	}
	client, err := client.NewSSHClient(server,  sshuer)
	defer client.Close()
	if err != nil{
		return err
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	bs,err := session.CombinedOutput(cmd)
	if err != nil {
		return err
	}
	utils.Logger.Infof("%s in %s run was sucessful,result:%v",cmd,host,bs)
	return nil
}
