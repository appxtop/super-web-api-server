package main

import (
	"fmt"
	"os"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

func SetupTray() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("可托盘应用")
	systray.SetTooltip("这是一个示例托盘应用")

	// 添加一个菜单项
	item1 := systray.AddMenuItem("打开界面", "这是菜单项1")
	go func() {
		for {
			<-item1.ClickedCh
			menuItemClick()
		}
	}()

	// 添加一个退出菜单项
	systray.AddSeparator() // 添加一个分割线
	exitItem := systray.AddMenuItem("退出", "退出应用程序")
	go func() {
		<-exitItem.ClickedCh
		onExit()
	}()

}

func onExit() {
	// Cleanup code here
	fmt.Println("程序退出")
	// 退出程序
	os.Exit(0)
}

func menuItemClick() {
	fmt.Println("菜单项被点击")
	err := OpenBrowser("https://www.baidu.com")
	if err != nil {
		fmt.Println("打开浏览器失败：", err)
	}
}
