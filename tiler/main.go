package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/mpasternacki/termbox-go"

	"../urxvtermbox"
)

var posX = 0
var posY = 0

var markX = -1
var markY = -1
var prefix = 1

func draw() {
	// axes
	for i := 0; i < 12; i++ {
		ch0 := ' '
		if i >= 9 {
			ch0 = '1'
		}
		ch1 := rune('0' + (i+1)%10)
		fgX := termbox.ColorDefault
		fgY := termbox.ColorDefault
		if i == posX {
			fgX = termbox.ColorWhite | termbox.AttrBold
		}
		if i == posY {
			fgY = termbox.ColorWhite | termbox.AttrBold
		}
		termbox.SetCell(0, i+1, ch0, fgY, termbox.ColorDefault)
		termbox.SetCell(1, i+1, ch1, fgY, termbox.ColorDefault)
		termbox.SetCell(26, i+1, ch0, fgY, termbox.ColorDefault)
		termbox.SetCell(27, i+1, ch1, fgY, termbox.ColorDefault)
		termbox.SetCell(2*i+2, 0, ch0, fgX, termbox.ColorDefault)
		termbox.SetCell(2*i+3, 0, ch1, fgX, termbox.ColorDefault)
		termbox.SetCell(2*i+2, 13, ch0, fgX, termbox.ColorDefault)
		termbox.SetCell(2*i+3, 13, ch1, fgX, termbox.ColorDefault)
	}

	// grid
	for i := 0; i < 12; i++ {
		for j := 0; j < 12; j++ {
			fg := termbox.ColorBlue
			if (i+j)%2 == 1 {
				fg = fg | termbox.AttrBold
			}

			ch := '░'
			if i == posX && j == posY {
				ch = '█'
				fg = termbox.ColorYellow | (fg & termbox.AttrBold)
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

	// prefix
	prfg := termbox.ColorGreen | termbox.AttrBold
	if prefix == 1 {
		prfg = termbox.ColorBlack | termbox.AttrBold
	}
	pr0 := ' '
	if prefix >= 10 {
		pr0 = '1'
	}
	pr1 := rune('0' + prefix%10)
	termbox.SetCell(0, 0, pr0, prfg, termbox.ColorDefault)
	termbox.SetCell(1, 0, pr1, prfg, termbox.ColorDefault)

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

func doMove(dx, dy int) {
	posX += dx
	if posX < 0 {
		posX = 0
	}
	if posX > 11 {
		posX = 11
	}
	posY += dy
	if posY < 0 {
		posY = 0
	}
	if posY > 11 {
		posY = 11
	}
}

func uiMain() error {
	if fini, err := urxvtermbox.TermboxUrxvt(28, 14); err != nil {
		return err
	} else {
		defer fini()
	}

	if err := termbox.Init(); err != nil {
		return err
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
				markX = -1
				markY = -1
				return nil
			case termbox.KeyArrowUp:
				doMove(0, -prefix)
				prefix = 1
			case termbox.KeyArrowDown:
				doMove(0, prefix)
				prefix = 1
			case termbox.KeyArrowLeft:
				doMove(-prefix, 0)
				prefix = 1
			case termbox.KeyArrowRight:
				doMove(prefix, 0)
				prefix = 1
			case termbox.KeyEnter:
				if markX >= 0 {
					return nil
				}
				fallthrough
			case termbox.KeySpace:
				markX, markY = posX, posY
				prefix = 1
			case termbox.KeyTab:
				if markX >= 0 {
					markX, posX = posX, markX
					markY, posY = posY, markY
				}
				prefix = 1
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				markX = -1
				markY = -1
				prefix = 1
			default:
				switch ev.Ch {
				case 'q':
					markX = -1
					markY = -1
					return nil
				case 'e':
					markX = -1
					markY = -1
					prefix = 1
				case 'w':
					doMove(0, -prefix)
					prefix = 1
				case 's':
					doMove(0, prefix)
					prefix = 1
				case 'a':
					doMove(-prefix, 0)
					prefix = 1
				case 'd':
					doMove(prefix, 0)
					prefix = 1
				case 'W':
					posY = 0
					prefix = 1
				case 'S':
					posY = 11
					prefix = 1
				case 'A':
					posX = 0
					prefix = 1
				case 'D':
					posX = 11
					prefix = 1
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					prefix = int(ev.Ch - '0')
				case '0':
					prefix = 10
				case '-':
					prefix = 11
				case '=':
					prefix = 12
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
			return ev.Err
		}
		draw()
	}

	return errors.New("CAN'T HAPPEN")
}

func main() {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// get active window's center
	axw, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		log.Fatal(err)
	}

	aw := xwindow.New(xu, axw)
	geom, err := aw.Geometry()
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

	err = uiMain()
	if err != nil {
		log.Fatal(err)
	}

	if markX < 0 {
		return
	}

	if posX > markX {
		posX, markX = markX, posX
	}

	if posY > markY {
		posY, markY = markY, posY
	}

	markX++
	markY++

	stepX := awHead.Width() / 12
	stepY := awHead.Height() / 12
	x := posX * stepX
	y := posY * stepY
	w := (markX - posX) * stepX
	h := (markY - posY) * stepY

	fmt.Println(posX, posY, markX, markY, "*", stepX, stepY, "→", x, y, w, h)

	// TODO: check if ewmh.MoveresizeWindow(xu, aw, x, y, w, h) is supported (not in cwm)
	aw.MoveResize(x, y, w, h)
	err = ewmh.ActiveWindowReq(xu, axw)
	if err != nil {
		log.Fatal(err)
	}
}
