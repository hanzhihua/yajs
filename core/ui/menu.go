package ui

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/hanzhihua/yajs/config"
	"github.com/hanzhihua/yajs/core/batchcmd"
	"github.com/hanzhihua/yajs/core/client"
	"github.com/hanzhihua/yajs/core/common"
	"github.com/hanzhihua/yajs/core/jumpserver"
	"github.com/hanzhihua/yajs/utils"
	"github.com/manifoldco/promptui"
	_ "golang.org/x/crypto/ssh"
	"reflect"
	"sort"
	"strings"
	"unsafe"
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
	IsShow            func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) bool
	GetSubMenu        func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) *[]*MenuItem
	SelectedFunc      func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) error
	BackAfterSelected bool
	Value             any
}

func defaultShow(int, *MenuItem, *common.YajsSession, []*MenuItem) bool { return true }

func serverShow(index int, item *MenuItem, session *common.YajsSession, items []*MenuItem) bool {
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

func menuShow(index int, item *MenuItem, session *common.YajsSession, items []*MenuItem) bool {
	ctx := (*session).Context()
	user := ctx.Value(utils.USER_KEY).(*config.User)
	b, err := config.CanAssessMenu(user.Username, item.Id)
	if err == nil && b {
		return true
	} else {
		return false
	}
}

func GetServersMenu() func(int, *MenuItem, *common.YajsSession, []*MenuItem) *[]*MenuItem {
	return func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) *[]*MenuItem {
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

func batchCMDSelect() func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) error {
	return func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) error {
		cmdFilePui := cmdFilePrompt("", sess)
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

func getServerItem(server *config.Server, session *common.YajsSession) *MenuItem {
	menuItem := new(MenuItem)
	menuItem.Label = fmt.Sprintf("%s %s:%d", server.Name, server.IP, server.Port)
	menuItem.Value = server
	menuItem.IsShow = serverShow
	menuItem.SelectedFunc = func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) error {

		theSshUser, err := config.Instance.GetSshUser(&session.Session, menuItem.Value.(*config.Server).Name)
		if err != nil {
			return err
		}
		utils.Logger.Warningf("theSshUser:%v", theSshUser.Username)
		err = client.NewTerminal(menuItem.Value.(*config.Server), theSshUser, sess)
		//signal(sess)
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

func enterJumpServer() func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) error {
	return func(index int, menuItem *MenuItem, sess *common.YajsSession, selectedChain []*MenuItem) error {
		return jumpserver.Enter(sess)
	}
}

var cmdFilePrompt = func(defaultShow string, sess *common.YajsSession) promptui.Prompt {
	return promptui.Prompt{
		Label:    "请输入命令文件",
		Validate: FileRequired("命令文件"),
		Default:  defaultShow,
		Stdin:    sess,
		Stdout:   sess,
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

func signal(sess *ssh.Session){
	value := reflect.ValueOf(sess)
	value = value.Elem()
	v := value.Interface()
	value = reflect.ValueOf(v).Elem()
	value = value.FieldByName("Channel")
	value = reflect.ValueOf(value.Interface()).Elem()
	value = value.FieldByName("pending")
	value = GetUnexportedField(value)
	//value = value.Elem()
	private_write(value.Interface(),[]byte(" "))
	//value = value.MethodByName("write")
	//value.Call([]reflect.Value{reflect.ValueOf([]byte(" "))})
	//value = value.FieldByName("Cond")
	//value = GetUnexportedField(value)
	//value.Interface().(*sync.Cond).Signal()
}

func GetUnexportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

//go:linkname private_write golang.org/x/crypto/ssh.(*buffer).write
func private_write(b interface{},buf []byte)