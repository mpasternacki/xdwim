package main

import (
	"errors"
	"strconv"
	"unicode/utf8"

	"github.com/mpasternacki/termbox-go"
)

type UIState struct {
	Desktops []WMDesktop
	Selected int
	Height   int
	Width    int
}

func NewUIState(desks []WMDesktop) UIState {
	st := UIState{
		Desktops: desks,
	}

	for i, desk := range desks {
		if desk.IsCurrent {
			st.Selected = i
		}
		if nw := len(desk.Windows); nw > st.Height {
			st.Height = nw
		}
		for _, win := range desk.Windows {
			if nw := utf8.RuneCountInString(win.Name); nw > st.Width {
				st.Width = nw
			}
		}
	}

	return st
}

func (ui *UIState) Desk() *WMDesktop {
	if ui.Selected < 0 {
		return nil
	}
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

var indexDigits = []rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹', '⁻', '⁼'}

func (ui *UIState) Draw() {
	cols, rows := termbox.Size()
	fgFrame := termbox.ColorWhite | termbox.AttrBold

	if rows < ui.Height+4 {
		panic("Too little rows!")
	}
	if cols < ui.Width+2 {
		panic("Too little cols!")
	}

	// Tab bar
	termbox.SetCell(0, 2, '╭', fgFrame, termbox.ColorDefault)
	col := 1
	for i, desk := range ui.Desktops {
		if !desk.IsVisible() {
			continue
		}

		fg := termbox.ColorDefault
		extra := termbox.Attribute(0)

		if i == ui.Selected {
			fg = termbox.ColorWhite
		}

		if desk.IsUrgent {
			fg = termbox.ColorRed
		}

		if desk.IsCurrent {
			extra = termbox.AttrUnderline
		}

		if i == ui.Selected {
			extra = extra | termbox.AttrBold
		}

		if i == ui.Selected {
			termbox.SetCell(col, 0, '╭', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, '│', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '╯', fgFrame, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
		}
		col++

		// TODO: desktop # for keyboard shortcut here
		index := ' '
		if i < len(indexDigits) {
			index = indexDigits[i]
		}
		if i == ui.Selected {
			termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, index, termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, index, termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
		}
		col++

		for _, ch := range desk.Name {
			termbox.SetCell(col, 1, ch, fg|extra, termbox.ColorDefault)
			if i == ui.Selected {
				termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
				termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
			} else {
				termbox.SetCell(col, 0, ' ', fgFrame, termbox.ColorDefault)
				termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
			}
			col++
		}

		if i == ui.Selected {
			termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
		}
		col++

		for _, ch := range strconv.Itoa(len(desk.Windows)) {
			termbox.SetCell(col, 1, ch, termbox.ColorDefault, termbox.ColorDefault)
			if i == ui.Selected {
				termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
				termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
			} else {
				termbox.SetCell(col, 0, ' ', fgFrame, termbox.ColorDefault)
				termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
			}
			col++
		}

		if i == ui.Selected {
			termbox.SetCell(col, 0, '╮', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, '│', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '╰', fgFrame, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
		}
		col++
	}

	if col > ui.Width {
		ui.Width = col
	}

	for ; col < ui.Width+1; col++ {
		termbox.SetCell(col, 2, '─', fgFrame, termbox.ColorDefault)
	}
	termbox.SetCell(ui.Width+1, 2, '╮', fgFrame, termbox.ColorDefault)

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

		termbox.SetCell(0, i+3, '│', fgFrame, termbox.ColorDefault)
		col = 1
		for _, ch := range win.Name {
			termbox.SetCell(col, i+3, ch, fg|extra, termbox.ColorDefault)
			col++
		}

		for ; col < ui.Width+1; col++ {
			termbox.SetCell(col, i+3, ' ', fg, termbox.ColorDefault)
		}
		termbox.SetCell(ui.Width+1, i+3, '│', fgFrame, termbox.ColorDefault)
	}

	for i := len(desk.Windows); i < ui.Height; i++ {
		termbox.SetCell(0, i+3, '│', fgFrame, termbox.ColorDefault)
		for j := 1; j < ui.Width+1; j++ {
			termbox.SetCell(j, i+3, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
		termbox.SetCell(ui.Width+1, i+3, '│', fgFrame, termbox.ColorDefault)
	}

	termbox.SetCell(0, ui.Height+3, '╰', fgFrame, termbox.ColorDefault)
	for j := 1; j < ui.Width+1; j++ {
		termbox.SetCell(j, ui.Height+3, '─', termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.SetCell(ui.Width+1, ui.Height+3, '╯', fgFrame, termbox.ColorDefault)

	termbox.Flush()
}

func (ui *UIState) Main() error {
	if fini, err := TermboxUrxvt(ui.Width+2, ui.Height+4); err != nil {
		return err
	} else {
		defer fini()
	}

	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	ui.Draw()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				ui.Selected = -1
				return nil
			case termbox.KeyArrowLeft:
				ui.Prev()
			case termbox.KeyArrowRight:
				ui.Next()
			case termbox.KeyArrowUp:
				ui.Desk().Prev()
			case termbox.KeyArrowDown:
				ui.Desk().Next()
			case termbox.KeyTab:
				ui.Desk().NextWrap()
			case termbox.KeyEnter:
				return nil
			}
		case termbox.EventError:
			return ev.Err
		}
		ui.Draw()
	}

	return errors.New("CAN'T HAPPEN")
}
