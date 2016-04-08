package main

import (
	"fmt"

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
	Name      string
	Selected  int
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

func (desk *WMDesktop) IsVisible() bool {
	return len(desk.Windows) > 0 || desk.IsCurrent
}

func (desk *WMDesktop) Next() {
	if desk.Selected < len(desk.Windows)-1 {
		desk.Selected++
	}
}

func (desk *WMDesktop) NextWrap() {
	if desk.Selected < len(desk.Windows)-1 {
		desk.Selected++
	} else {
		desk.Selected = 0
	}
}

func (desk *WMDesktop) Prev() {
	if desk.Selected > 0 {
		desk.Selected--
	}
}

func (desk *WMDesktop) PrevWrap() {
	if desk.Selected > 0 {
		desk.Selected--
	} else {
		desk.Selected = len(desk.Windows) - 1
	}
}

func (desk *WMDesktop) Window() *WMWindow {
	return &desk.Windows[desk.Selected]
}

func Desktops(xu *xgbutil.XUtil) ([]WMDesktop, error) {
	ndesk, err := ewmh.NumberOfDesktopsGet(xu)
	if err != nil {
		return nil, err
	}

	desktops := make([]WMDesktop, ndesk)

	names, err := ewmh.DesktopNamesGet(xu)
	if err != nil {
		return nil, err
	}

	for i := range desktops {
		desktops[i].Number = uint(i)
		desktops[i].Name = names[i]
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

		isActive := xw == aw
		isUrgent := hints.Flags&icccm.HintUrgency == icccm.HintUrgency

		desktops[desk].Windows = append(desktops[desk].Windows, WMWindow{
			XWin:     xw,
			IsActive: isActive,
			IsUrgent: isUrgent,
			Name:     name,
		})

		desktops[desk].IsUrgent = desktops[desk].IsUrgent || isUrgent

		if isActive {
			desktops[desk].Selected = len(desktops[desk].Windows) - 1
		}
	}

	return desktops, nil
}
