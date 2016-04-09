package main

import (
	"errors"
	"log"
	"strconv"
	"unicode/utf8"

	"../urxvtermbox"
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

var indexDigits = []rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹'}

func (ui *UIState) Draw() {
	cols, rows := termbox.Size()
	fgFrame := termbox.ColorYellow
	fgTitle := termbox.ColorGreen

	if rows < ui.Height+4 {
		panic("Too little rows!")
	}
	if cols < ui.Width+2 {
		panic("Too little cols!")
	}

	// Tab bar
	termbox.SetCell(0, 2, '╭', fgFrame|termbox.AttrBold, termbox.ColorDefault)
	col := 1
	for i, desk := range ui.Desktops {
		if !desk.IsVisible() {
			continue
		}

		fg := fgTitle // termbox.ColorDefault
		extra := termbox.Attribute(0)

		if i == ui.Selected {
			// fg = termbox.ColorWhite
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

		if i < ui.Selected {
			termbox.SetCell(col, 0, '╭', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, '│', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		} else if i == ui.Selected {
			termbox.SetCell(col, 0, '╭', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			termbox.SetCell(col, 1, '│', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			termbox.SetCell(col, 2, '╯', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		}
		col++

		index := ' '
		if i < len(indexDigits) {
			index = indexDigits[i]
		}
		if i == ui.Selected {
			termbox.SetCell(col, 0, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			termbox.SetCell(col, 1, index, fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, index, fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		}
		col++

		for _, ch := range desk.Name {
			termbox.SetCell(col, 1, ch, fg|extra, termbox.ColorDefault)
			if i == ui.Selected {
				termbox.SetCell(col, 0, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
				termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
			} else {
				termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
				termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			}
			col++
		}

		if i == ui.Selected {
			termbox.SetCell(col, 0, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		}
		col++

		for _, ch := range strconv.Itoa(len(desk.Windows)) {
			termbox.SetCell(col, 1, ch, fgFrame, termbox.ColorDefault)
			if i == ui.Selected {
				termbox.SetCell(col, 0, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
				termbox.SetCell(col, 2, ' ', fgFrame, termbox.ColorDefault)
			} else {
				termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
				termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			}
			col++
		}

		if i > ui.Selected {
			termbox.SetCell(col, 0, '╮', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, '│', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		} else if i == ui.Selected {
			termbox.SetCell(col, 0, '╮', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			termbox.SetCell(col, 1, '│', fgFrame|termbox.AttrBold, termbox.ColorDefault)
			termbox.SetCell(col, 2, '╰', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, 0, '─', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 1, ' ', fgFrame, termbox.ColorDefault)
			termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		}
		col++
	}

	if col > ui.Width {
		ui.Width = col
	}

	for ; col < ui.Width+1; col++ {
		termbox.SetCell(col, 2, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
	}
	termbox.SetCell(ui.Width+1, 2, '╮', fgFrame|termbox.AttrBold, termbox.ColorDefault)

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

		termbox.SetCell(0, i+3, '│', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		col = 1
		for _, ch := range win.Name {
			termbox.SetCell(col, i+3, ch, fg|extra, termbox.ColorDefault)
			col++
		}

		for ; col < ui.Width+1; col++ {
			termbox.SetCell(col, i+3, ' ', fg, termbox.ColorDefault)
		}
		termbox.SetCell(ui.Width+1, i+3, '│', fgFrame|termbox.AttrBold, termbox.ColorDefault)
	}

	for i := len(desk.Windows); i < ui.Height; i++ {
		termbox.SetCell(0, i+3, '│', fgFrame|termbox.AttrBold, termbox.ColorDefault)
		for j := 1; j < ui.Width+1; j++ {
			termbox.SetCell(j, i+3, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
		termbox.SetCell(ui.Width+1, i+3, '│', fgFrame|termbox.AttrBold, termbox.ColorDefault)
	}

	termbox.SetCell(0, ui.Height+3, '╰', fgFrame|termbox.AttrBold, termbox.ColorDefault)
	for j := 1; j < ui.Width+1; j++ {
		termbox.SetCell(j, ui.Height+3, '─', fgFrame|termbox.AttrBold, termbox.ColorDefault)
	}
	termbox.SetCell(ui.Width+1, ui.Height+3, '╯', fgFrame|termbox.AttrBold, termbox.ColorDefault)

	termbox.Flush()
}

func (ui *UIState) Main() (erv error) {
	if fini, err := urxvtermbox.TermboxUrxvt(ui.Width+2, ui.Height+4, "-pe", "destroy_on_focus_out"); err != nil {
		return err
	} else {
		defer func() {
			if err := fini(); err != nil {
				if erv == nil {
					erv = err
					// It's been logged anyway
				}
			}
		}()
	}

	termbox.SetInputMode(termbox.InputEsc)

	ui.Draw()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return cmdCancel
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
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				return cmdCloseWindow
			default:
				switch ev.Ch {
				case 'w': // ↑
					ui.Prev()
				case 's': // ↓
					ui.Next()
				case 'a': // ←
					ui.Desk().Prev()
				case 'd': // →
					ui.Desk().Next()
				case 'q': // Esc
					return cmdCancel
				case 'e': // move to active window
					for i, desk := range ui.Desktops {
						if desk.IsCurrent {
							ui.Selected = i
							for j, win := range desk.Windows {
								if win.IsActive {
									// desk is not a pointer
									ui.Desktops[i].Selected = j
									break
								}
							}
							break
						}
					}

				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					if desk := int(ev.Ch - '0'); desk < len(ui.Desktops) && len(ui.Desktops[desk].Windows) > 0 {
						ui.Selected = desk
					}
				case ' ': // same as Enter
					return nil
				case '!':
					// Find next urgent window
					sxw := ui.Desk().Window().XWin
					d := ui.Selected
					w := ui.Desk().Selected
					for {
						w++                                   // next window
						if w >= len(ui.Desktops[d].Windows) { // next desktop
							w = 0
							for {
								d++
								if d >= len(ui.Desktops) {
									d = 0
								}
								if len(ui.Desktops[d].Windows) > 0 {
									break
								}
							}
						}
						win := ui.Desktops[d].Windows[w]
						if win.XWin == sxw {
							// We have wrapped around, let's break
							break
						}
						if win.IsUrgent {
							ui.Selected = d
							ui.Desk().Selected = w
							break
						}
					}
				}
			}
		case termbox.EventInterrupt:
			return cmdCancel
		case termbox.EventError:
			return ev.Err
		default:
			log.Println("EVENT:", ev)
		}
		ui.Draw()
	}

	return errors.New("CAN'T HAPPEN")
}
