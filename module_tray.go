package main

import "github.com/getlantern/systray"

func Module_tray() {
	go systray.Run(SetupTray, OnExit)
}
