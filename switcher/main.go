package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
)

func main() {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	desks, err := Desktops(xu)
	if err != nil {
		log.Fatal(err)
	}

	ui := NewUIState(desks)

	err = ui.Main()
	if err != nil {
		log.Fatal(err)
	}

	if desk := ui.Desk(); desk != nil {
		win := desk.Window()
		xw := win.XWin
		log.Println(win, xw, desk.Number)

		err = ewmh.CurrentDesktopReq(xu, int(desk.Number))
		if err != nil {
			log.Fatal(err)
		}

		err = ewmh.ActiveWindowReq(xu, xw)
		if err != nil {
			log.Fatal(err)
		}
	}
}
