package ui

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"sort"
	"yajs/config"
	"yajs/core/client"
	"yajs/core/jumpserver"
	"yajs/utils"
)

var (
	MainMenu        *[]*MenuItem
	ListServersMenu *MenuItem
	BatchCMDMenu   *MenuItem
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

func serverShow(index int,item *MenuItem,session *ssh.Session,items []*MenuItem) bool {
	server,ok := item.Value.(*config.Server)
	if ok{
		ctx := (*session).Context()
		user := ctx.Value(utils.USER_KEY).(*config.User)
		b,err := config.CanAssessServer(user.Username,server.Name)
		if err== nil && b{
			return true
		}else{
			return false
		}
	}
	return true
}

func menuShow(index int,item *MenuItem,session *ssh.Session,items []*MenuItem) bool {
	ctx := (*session).Context()
	user := ctx.Value(utils.USER_KEY).(*config.User)
	b,err := config.CanAssessMenu(user.Username,item.Id)
	if err== nil && b{
		return true
	}else{
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
				getServerItem(servers[i],sess),
			)
		}
		return &menuItems
	}
}

func getServerItem(server *config.Server,session *ssh.Session ) *MenuItem {
	menuItem := new(MenuItem)
	menuItem.Label = fmt.Sprintf("%s %s:%d", server.Name, server.IP,server.Port)
	menuItem.Value = server
	menuItem.IsShow = serverShow
	menuItem.SelectedFunc = func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {

		theSshUser,err := config.Instance.GetSshUser(session,menuItem.Value.(*config.Server).Name)
		if err != nil{
			return err
		}
		utils.Logger.Warningf("theSshUser:%v",theSshUser)
		err = client.NewTerminal(menuItem.Value.(*config.Server), theSshUser, sess)
		return err
	}
	return menuItem
}

func init() {
	ListServersMenu = &MenuItem{
		Id:            "serverList",
		Label:         "服务器列表",
		IsShow:        defaultShow,
		GetSubMenu:    GetServersMenu(),
	}

	//BatchCMDMenu = &MenuItem{
	//	Label:         "批量运行",
	//	IsShow:        defaultShow,
	//	SelectedFunc:    batchCMDSelect(),
	//}
	//
	EnterJumpServerMenu = &MenuItem{
		Id:            "jumpserver",
		Label:         "进入跳板机",
		IsShow:        menuShow,
		SelectedFunc:    enterJumpServer(),
	}

	MainMenu = &[]*MenuItem{
		//ListServersMenu,BatchCMDMenu,EnterMenu,
		ListServersMenu,EnterJumpServerMenu,
	}
}

func enterJumpServer() func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
		return jumpserver.Enter(sess)
	}
}