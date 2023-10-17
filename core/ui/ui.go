package ui

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/manifoldco/promptui"
	"sort"
	"strings"
	"yajs/config"
	"yajs/core/client"
	"yajs/core/helper"
	"yajs/utils"
)

type UIService struct {
	Session *ssh.Session
}

type MenuItem struct {
	Id                string
	Label             string
	Info              map[string]interface{}
	IsShow            func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) bool
	SubMenuTitle      string
	GetSubMenu        func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem
	SelectedFunc      func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error
	NoSubMenuInfo     string
	BackAfterSelected bool
	BackOptionLabel   string
	Value             interface{}
}

var (
	MainMenu        *[]*MenuItem
	ListServersMenu *MenuItem
	BatchCMDMenu   *MenuItem
	EnterJumpServerMenu *MenuItem
)

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
		NoSubMenuInfo: "目前没有服务器，请在server manager加入服务器",
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

const logo = `
   __        _
  / /__   __(_)__ _      __
 / __/ | / / / _ \ | /| / /
/ /_ | |/ / /  __/ |/ |/ /
\__/ |___/_/\___/|__/|__/

`

func (uiService *UIService) ShowUI(){
	defer func() {
		(*uiService.Session).Exit(0)
	}()
	selectedChain := make([]*MenuItem, 0)
	//templ := `{{ .Title "Yet another jump server" "" 0 }}`
	//var buffer bytes.Buffer
	//banner.Init(&buffer, true, true, bytes.NewBufferString(templ))
	//label := buffer.String()
	//fmt.Println(label)
	uiService.ShowMenu(utils.SLOGAN, MainMenu, "退出", selectedChain)
}

func (uiService *UIService) ShowMenu(label string, menu *[]*MenuItem, BackOptionLabel string, selectedChain []*MenuItem) {
	for {
		menuLabels := make([]string, 0)
		menuItems := make([]*MenuItem, 0)

		if menu == nil {
			utils.Logger.Warningf("menu is nil")
			break
		}

		for index, menuItem := range *menu {
			if menuItem.IsShow == nil || menuItem.IsShow(index, menuItem, uiService.Session, selectedChain) {
				menuLabels = append(menuLabels, menuItem.Label)
				menuItems = append(menuItems, menuItem)
			}
		}

		menuLabels = append(menuLabels, BackOptionLabel)
		backIndex := len(menuLabels) - 1

		templates := &promptui.SelectTemplates{
			Label:    "{{ . | green }}",
			Active:   "\U0001F527 \033[4m{{ . | cyan }}\033[0m",
			Inactive: "  {{ . | cyan }}",
			Selected: "\U0001F527 {{ . | green }}",
		}

		searcher := func(input string, index int) bool {
			name := strings.Replace(strings.ToLower(menuLabels[index]), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		menuPui := promptui.Select{
			Label:  label,
			Items:  menuLabels,
			Stdin:  *uiService.Session,
			Stdout: *uiService.Session,
			Templates: templates,
			Searcher: searcher,
			Size: 10,
		}

		index, subMenuLabel, err := menuPui.Run()

		if err != nil {
			utils.Logger.Infof("Select menu error %s\n", err)
			break
		}

		if index == backIndex {
			break
		}

		selected := menuItems[index]

		if selected.GetSubMenu != nil {
			selectedChain = append(selectedChain, selected)
			subMenu := selected.GetSubMenu(index, selected, uiService.Session, selectedChain)

			if subMenu != nil && len(*subMenu) > 0 {
				back := "返回上一级"
				if selected.BackOptionLabel != "" {
					back = selected.BackOptionLabel
				}

				if selected.SubMenuTitle != "" {
					subMenuLabel = selected.SubMenuTitle
				}
				uiService.ShowMenu(subMenuLabel, subMenu, back, selectedChain)
			} else {
				noSubMenuInfo := "["+selected.Label+"] 下面没有东西 ... "
				if selected.NoSubMenuInfo != "" {
					noSubMenuInfo = selected.NoSubMenuInfo
				}
				ErrorInfo(selected.Label+" has error:",errors.New(noSubMenuInfo), uiService.Session)
			}
		}else if selected.SelectedFunc != nil {
			selectedChain = append(selectedChain, selected)
			err := selected.SelectedFunc(index, selected, uiService.Session, selectedChain)
			if err != nil {
				utils.Logger.Errorf("Run selected func err: %v", err)
				ErrorInfo(selected.Label+" has error:",err, uiService.Session)
			}
		}
	}
}

func ErrorInfoWithStr(str string, sess *ssh.Session) {
	err := errors.New(str)
	ErrorInfo("",err,sess)
}

func ErrorInfo(prefix string,err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s %s\n", prefix,err))
}

// Info Info
func Info(msg string, sess *ssh.Session) {
	green := color.New(color.FgGreen)
	green.Fprint(*sess, msg)
}


func enterJumpServer() func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
		return helper.EnterJumpServer(sess)
	}
}
