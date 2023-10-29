package ui

import (
	"fmt"
	"github.com/hanzhihua/yajs/config"
	"github.com/hanzhihua/yajs/core/batchcmd"
	"github.com/hanzhihua/yajs/core/client"
	"github.com/hanzhihua/yajs/core/jumpserver"
	"github.com/hanzhihua/yajs/utils"
	"io"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/manifoldco/promptui"
)

var (
	MainMenu            *[]*MenuItem
	ListServersMenu     *MenuItem
	BatchCMDMenu        *MenuItem
	EnterJumpServerMenu *MenuItem
)

type MenuItem struct {
	Id                string
	Label             string
	IsShow            func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) bool
	GetSubMenu        func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem
	SelectedFunc      func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error
	BackAfterSelected bool
	Value             any
}

func defaultShow(int, *MenuItem, *ssh.Session, []*MenuItem) bool { return true }

func serverShow(index int, item *MenuItem, session *ssh.Session, items []*MenuItem) bool {
	server, ok := item.Value.(*config.Server)
	if ok {
		ctx := (*session).Context()
		user := ctx.Value(utils.USER_KEY).(*config.User)
		b, err := config.CanAssessServer(user.Username, server.Name)
		if err == nil && b {
			return true
		} else {
			return false
		}
	}
	return true
}

func menuShow(index int, item *MenuItem, session *ssh.Session, items []*MenuItem) bool {
	ctx := (*session).Context()
	user := ctx.Value(utils.USER_KEY).(*config.User)
	b, err := config.CanAssessMenu(user.Username, item.Id)
	if err == nil && b {
		return true
	} else {
		return false
	}
}

func GetServersMenu() func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem {
		menuItems := make([]*MenuItem, 0)
		servers := config.Instance.Servers
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].Name < servers[j].Name
		})
		for i, _ := range servers {
			menuItems = append(
				menuItems,
				getServerItem(servers[i], sess),
			)
		}
		return &menuItems
	}
}

func batchCMDSelect() func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
		cmdFilePui := cmdFilePrompt("", *sess)
		cmdFile, err := cmdFilePui.Run()
		if err != nil {
			return err
		}
		err = batchcmd.BatchRunCMD(sess, cmdFile)
		if err == nil {
			green := color.New(color.FgGreen)
			green.Fprint(*sess, fmt.Sprintf("%s execute successfully", cmdFile))
		}
		return err
	}
}

func getServerItem(server *config.Server, session *ssh.Session) *MenuItem {
	menuItem := new(MenuItem)
	menuItem.Label = fmt.Sprintf("%s %s:%d", server.Name, server.IP, server.Port)
	menuItem.Value = server
	menuItem.IsShow = serverShow
	menuItem.SelectedFunc = func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {

		theSshUser, err := config.Instance.GetSshUser(session, menuItem.Value.(*config.Server).Name)
		if err != nil {
			return err
		}
		utils.Logger.Warningf("theSshUser:%v", theSshUser.Username)
		err = client.NewTerminal(menuItem.Value.(*config.Server), theSshUser, sess)
		return err
	}
	return menuItem
}

func init() {
	ListServersMenu = &MenuItem{
		Id:         "serverList",
		Label:      "服务器列表",
		IsShow:     defaultShow,
		GetSubMenu: GetServersMenu(),
	}

	BatchCMDMenu = &MenuItem{
		Id:           "batchCmd",
		Label:        "批量运行",
		IsShow:       menuShow,
		SelectedFunc: batchCMDSelect(),
	}
	//
	EnterJumpServerMenu = &MenuItem{
		Id:           "jumpserver",
		Label:        "进入跳板机",
		IsShow:       menuShow,
		SelectedFunc: enterJumpServer(),
	}

	MainMenu = &[]*MenuItem{
		ListServersMenu, BatchCMDMenu, EnterJumpServerMenu,
	}
}

func enterJumpServer() func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
		return jumpserver.Enter(sess)
	}
}

var cmdFilePrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label:    "请输入命令文件",
		Validate: FileRequired("命令文件"),
		Default:  defaultShow,
		Stdin:    stdio,
		Stdout:   stdio,
	}
}

func FileRequired(field string) func(string) error {
	return func(input string) error {
		i := strings.ReplaceAll(input, " ", "")
		if i == "" {
			return fmt.Errorf("Please input %s", field)
		}
		if !utils.FileExited(utils.FilePath(i)) {
			return fmt.Errorf("%s  not existed", i)
		}
		return nil
	}
}
