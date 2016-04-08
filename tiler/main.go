package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/mpasternacki/termbox-go"
)

var posX = 0
var posY = 0

var markX = -1
var markY = -1

func draw() {
	for i := 0; i < 12; i++ {
		ch0 := ' '
		if i >= 9 {
			ch0 = '1'
		}
		ch1 := rune('0' + (i+1)%10)
		termbox.SetCell(0, i+1, ch0, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(1, i+1, ch1, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(26, i+1, ch0, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(27, i+1, ch1, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(2*i+2, 0, ch0, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(2*i+3, 0, ch1, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(2*i+2, 13, ch0, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(2*i+3, 13, ch1, termbox.ColorDefault, termbox.ColorDefault)
	}

	for i := 0; i < 12; i++ {
		for j := 0; j < 12; j++ {
			fg := termbox.ColorBlue
			if (i+j)%2 == 1 {
				fg = fg | termbox.AttrBold
			}

			ch := '░'
			if i == posX && j == posY {
				ch = '█'
				fg = termbox.ColorCyan
			} else if markX >= 0 && markY >= 0 {
				l, r, t, b := posX, markX, posY, markY
				if l > r {
					l, r = r, l
				}
				if t > b {
					t, b = b, t
				}
				if l <= i && i <= r && t <= j && j <= b {
					ch = '▓'
				}
			}

			termbox.SetCell(2*i+2, j+1, ch, fg, termbox.ColorDefault)
			termbox.SetCell(2*i+3, j+1, ch, fg, termbox.ColorDefault)
		}
	}

	termbox.Flush()
}

func mousePos(ev termbox.Event) (x int, y int) {
	x, y = (ev.MouseX-2)/2, ev.MouseY-1
	if x < 0 {
		x = 0
	}
	if x > 11 {
		x = 11
	}
	if y < 0 {
		y = 0
	}
	if y > 11 {
		y = 11
	}
	return
}

func main() {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// get active window's center
	aw, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		log.Fatal(err)
	}

	geom, err := xwindow.New(xu, aw).Geometry()
	if err != nil {
		log.Fatal(err)
	}

	cx := geom.X() + geom.Width()/2
	cy := geom.Y() + geom.Height()/2

	heads, err := xinerama.PhysicalHeads(xu)
	if err != nil {
		log.Fatal(err)
	}

	awHead := heads[0]
	for _, head := range heads {
		if cx >= head.X() &&
			cx <= head.X()+head.Width() &&
			cy >= head.Y() &&
			cy <= head.X()+head.Height() {
			awHead = head
			break
		}
	}

	fmt.Println(heads)
	fmt.Println(geom, "→", cx, cy, "→", awHead)

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	draw()
	mouseHold := false
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return
				// return nil
			case termbox.KeyArrowUp:
				if posY > 0 {
					posY--
				}
			case termbox.KeyArrowDown:
				if posY < 11 {
					posY++
				}
			case termbox.KeyArrowLeft:
				if posX > 0 {
					posX--
				}
			case termbox.KeyArrowRight:
				if posX < 11 {
					posX++
				}
			case termbox.KeyEnter:
				if markX < 0 {
					markX, markY = posX, posY
				} else {
					markX, posX = posX, markX
					markY, posY = posY, markY
				}
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				markX = -1
				markY = -1
			default:
				switch ev.Ch {
				case 'q':
					return // nil
				}
			}
		case termbox.EventMouse:
			switch ev.Key {
			case termbox.MouseLeft:
				posX, posY = mousePos(ev)
				if !mouseHold {
					markX, markY = posX, posY
				}
				mouseHold = true
			case termbox.MouseRight:
				markX, markY = mousePos(ev)
			case termbox.MouseRelease:
				mouseHold = false
			}
		case termbox.EventError:
			log.Fatal(ev.Err)
		}
		draw()
	}
}
