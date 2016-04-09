package main

import (
	"errors"
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
)

var cmdCloseWindow = errors.New("CLOSE WINDOW")
var cmdCancel = errors.New("CANCEL")

func innerMain() error {
	xu, err := xgbutil.NewConn()
	if err != nil {
		return err
	}

	desks, err := Desktops(xu)
	if err != nil {
		return err
	}

	ui := NewUIState(desks)

	switch err := ui.Main(); err {
	case cmdCancel:
		return nil
	case cmdCloseWindow:
		xw := ui.Desk().Window().XWin
		return ewmh.CloseWindow(xu, xw)
	case nil:
		// default: choose window
		desk := ui.Desk()
		win := desk.Window()
		xw := win.XWin

		err = ewmh.CurrentDesktopReq(xu, int(desk.Number))
		if err != nil {
			return err
		}

		return ewmh.ActiveWindowReq(xu, xw)
	default:
		return err
	}
}

func main() {
	if err := innerMain(); err != nil {
		log.Fatal(err)
	}
}
