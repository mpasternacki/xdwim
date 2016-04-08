package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/nsf/termbox-go"
)

type WMWindow struct {
	XWin     xproto.Window
	IsActive bool
	IsUrgent bool
	Name     string
}

type WMDesktop struct {
	Number    uint
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

func (desk *WMDesktop) Prev() {
	if desk.Selected > 0 {
		desk.Selected--
	}
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

type UIState struct {
	Desktops []WMDesktop
	Selected int
}

func NewUIState(desks []WMDesktop) UIState {
	seldesk := 0
	for i, desk := range desks {
		if desk.IsCurrent {
			seldesk = i
			break
		}
	}

	return UIState{
		Desktops: desks,
		Selected: seldesk,
	}
}

func (ui *UIState) Desk() *WMDesktop {
	return &ui.Desktops[ui.Selected]
}

func (ui *UIState) Prev() {
	for cur := ui.Selected - 1; cur > 0; cur-- {
		if ui.Desktops[cur].IsVisible() {
			ui.Selected = cur
			return
		}
	}
}

func (ui *UIState) Next() {
	for cur := ui.Selected + 1; cur < len(ui.Desktops); cur++ {
		if ui.Desktops[cur].IsVisible() {
			ui.Selected = cur
			return
		}
	}
}

func (ui *UIState) Draw() {
	cols, rows := termbox.Size()

	// Tab bar
	col := 0
	for i, desk := range ui.Desktops {
		if !desk.IsVisible() {
			continue
		}

		fg := termbox.ColorDefault
		extra := termbox.Attribute(0)

		if desk.IsUrgent {
			fg = termbox.ColorRed
		}

		if desk.IsCurrent {
			extra = termbox.AttrUnderline
		}

		if i == ui.Selected {
			fg = fg | termbox.AttrReverse
		}

		label := fmt.Sprintf("%d (%d)", i, len(desk.Windows))
		termbox.SetCell(col, 0, ' ', fg, termbox.ColorDefault)
		col++

		for _, ch := range label {
			termbox.SetCell(col, 0, ch, fg|extra, termbox.ColorDefault)
			col++
		}

		termbox.SetCell(col, 0, ' ', fg, termbox.ColorDefault)
		col++
	}

	// Window List
	desk := ui.Desk()
	for i, win := range desk.Windows {
		fg := termbox.ColorDefault
		extra := termbox.Attribute(0)

		if win.IsUrgent {
			fg = termbox.ColorRed
		}

		if win.IsActive {
			extra = termbox.AttrUnderline
		}

		if i == desk.Selected {
			fg = fg | termbox.AttrReverse
		}

		termbox.SetCell(0, i+1, ' ', fg, termbox.ColorDefault)
		col = 1
		for _, ch := range win.Name {
			termbox.SetCell(col, i+1, ch, fg|extra, termbox.ColorDefault)
			col++
		}

		for ; col < cols; col++ {
			termbox.SetCell(col, i+1, ' ', fg, termbox.ColorDefault)
		}
	}

	for i := 1 + len(desk.Windows); i < rows; i++ {
		for j := 0; j < cols; j++ {
			termbox.SetCell(j, i, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	termbox.Flush()
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

	// Drawing stuff
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	ui := NewUIState(desks)
	ui.Draw()
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyArrowLeft:
				ui.Prev()
			case termbox.KeyArrowRight:
				ui.Next()
			case termbox.KeyArrowUp:
				ui.Desk().Prev()
			case termbox.KeyArrowDown:
				ui.Desk().Next()
				// case termbox.KeyBackspace, termbox.KeyBackspace2:
				// 	edit_box.DeleteRuneBackward()
				// case termbox.KeyDelete, termbox.KeyCtrlD:
				// 	edit_box.DeleteRuneForward()
				// case termbox.KeyTab:
				// 	edit_box.InsertRune('\t')
				// case termbox.KeySpace:
				// 	edit_box.InsertRune(' ')
				// case termbox.KeyCtrlK:
				// 	edit_box.DeleteTheRestOfTheLine()
				// case termbox.KeyHome, termbox.KeyCtrlA:
				// 	edit_box.MoveCursorToBeginningOfTheLine()
				// case termbox.KeyEnd, termbox.KeyCtrlE:
				// 	edit_box.MoveCursorToEndOfTheLine()
				// default:
				// 	if ev.Ch != 0 {
				// 		edit_box.InsertRune(ev.Ch)
				// 	}
			}
		case termbox.EventError:
			log.Fatal(ev.Err)
		}
		ui.Draw()
	}
}
