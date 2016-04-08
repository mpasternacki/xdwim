package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/nsf/termbox-go"
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
			case termbox.KeyTab:
				ui.Desk().NextWrap()
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
