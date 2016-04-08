package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
)

type WMWindow struct {
	XWin     xproto.Window
	IsActive bool
	IsUrgent bool
	Name     string
}

type WMDesktop struct {
	Number    uint
	IsCurrent bool
	IsUrgent  bool
	Windows   []WMWindow
}

func (wmw WMWindow) String() string {
	activeFlag := ""
	if wmw.IsActive {
		activeFlag = "*"
	}

	urgentFlag := ""
	if wmw.IsUrgent {
		urgentFlag = "!"
	}

	return fmt.Sprintf("%v%v%#v", activeFlag, urgentFlag, wmw.Name)
}

func (wmd WMDesktop) String() string {
	currentFlag := ""
	if wmd.IsCurrent {
		currentFlag = "*"
	}

	urgentFlag := ""
	if wmd.IsUrgent {
		urgentFlag = "!"
	}

	wins := make([]string, len(wmd.Windows))
	for i, wmw := range wmd.Windows {
		wins[i] = wmw.String()
	}

	return fmt.Sprintf("%s%s%d%v", currentFlag, urgentFlag, wmd.Number, wins)
}

func Desktops(xu *xgbutil.XUtil) ([]WMDesktop, error) {
	ndesk, err := ewmh.NumberOfDesktopsGet(xu)
	if err != nil {
		return nil, err
	}

	desktops := make([]WMDesktop, ndesk)

	for i := range desktops {
		desktops[i].Number = uint(i)
	}

	cdesk, err := ewmh.CurrentDesktopGet(xu)
	if err != nil {
		return nil, err
	}
	desktops[cdesk].IsCurrent = true

	aw, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		return nil, err
	}

	// TODO: check whether ClientListStackingGet is supported (cwm
	// doesn't support it)
	xws, err := ewmh.ClientListGet(xu)
	if err != nil {
		return nil, err
	}

	for _, xw := range xws {
		name, err := ewmh.WmNameGet(xu, xw)
		if err != nil {
			return nil, err
		}

		hints, err := icccm.WmHintsGet(xu, xw)
		if err != nil {
			return nil, err
		}

		desk, err := ewmh.WmDesktopGet(xu, xw)
		if err != nil {
			return nil, err
		}

		isUrgent := hints.Flags&icccm.HintUrgency == icccm.HintUrgency

		desktops[desk].Windows = append(desktops[desk].Windows, WMWindow{
			XWin:     xw,
			IsActive: xw == aw,
			IsUrgent: isUrgent,
			Name:     name,
		})

		desktops[desk].IsUrgent = desktops[desk].IsUrgent || isUrgent
	}

	return desktops, nil
}

func main() {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	desks, err := Desktops(xu)
	if err != nil {
		log.Fatal(err)
	}

	for _, desk := range desks {
		fmt.Println(desk)
	}
}
