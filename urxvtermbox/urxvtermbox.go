package urxvtermbox

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kr/pty"
	"github.com/mpasternacki/termbox-go"
)

func UrxvtPty(args ...string) (*os.File, func(), error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, nil, err
	}

	cmd := exec.Command("urxvt", append([]string{"-pty-fd", "3"}, args...)...)
	fini := func() {
		slave.Close()
		if cmd.Process != nil {
			cmd.Wait()
		}
	}

	cmd.ExtraFiles = []*os.File{master}
	cmd.Start()

	// poll until urxvt is done starting
	for {
		rows, _, err := pty.Getsize(slave)
		if err != nil {
			fini()
			return nil, nil, err
		}
		if rows > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	return slave, fini, nil
}

func TermboxUrxvt(width, height int, args ...string) (func(), error) {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 25
	}

	pty, fini, err := UrxvtPty(append(
		[]string{"-geometry", fmt.Sprintf("%dx%d", width, height)},
		args...)...)

	if err != nil {
		return nil, err
	}

	origTerminalDevice := termbox.TerminalDevice
	termbox.TerminalDevice = pty.Name()
	return func() {
		termbox.TerminalDevice = origTerminalDevice
		fini()
	}, nil
}
