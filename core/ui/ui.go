package ui

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/manifoldco/promptui"
	"strings"
	"yajs/utils"
)

type UIService struct {
	Session *ssh.Session
}

func (uiService *UIService) ShowUI(){
	selectedChain := make([]*MenuItem, 0)
	uiService.ShowMenu(utils.SLOGAN, MainMenu, "退出", selectedChain)
}

func (uiService *UIService) ShowMenu(label string, menu *[]*MenuItem, BackOptionLabel string, selectedChain []*MenuItem) {
	var menuPui *promptui.Select
	var index,backIndex int
	var subMenuLabel string
	var err error
	menuLabels := make([]string, 0)
	menuItems := make([]*MenuItem, 0)
	for {
		if menuPui == nil{
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
			backIndex = len(menuLabels) - 1

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

			menuPui = &promptui.Select{
				Label:  label,
				Items:  menuLabels,
				Stdin:  *uiService.Session,
				Stdout: *uiService.Session,
				Templates: templates,
				Searcher: searcher,
				Size: 10,
			}
			index, subMenuLabel, err = menuPui.Run()
		}else{
			index, subMenuLabel, err = menuPui.RunCursorAt(index, menuPui.ScrollPosition())
		}

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
				uiService.ShowMenu(subMenuLabel, subMenu, back, selectedChain)
			} else {
				noSubMenuInfo := "["+selected.Label+"] 下面没有东西 ... "
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

func ErrorInfo(prefix string,err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s %s\n", prefix,err))
}



