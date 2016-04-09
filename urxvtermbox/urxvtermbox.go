package urxvtermbox

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kr/pty"
	"github.com/mpasternacki/termbox-go"
)

var UrxvtermboxPath string

func init() {
	absbin, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	realbin, err := filepath.EvalSymlinks(absbin)
	if err != nil {
		panic(err)
	}
	UrxvtermboxPath = filepath.Join(filepath.Dir(filepath.Dir(realbin)), "urxvtermbox", "perl")
}

func UrxvtPty(args ...string) (*os.File, func(), <-chan error, error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, nil, nil, err
	}

	cmd := exec.Command("urxvt", append([]string{
		"-pty-fd", "3",
		"-title", filepath.Base(os.Args[0]),
		"--perl-lib", UrxvtermboxPath,
	}, args...)...)

	cmd.ExtraFiles = []*os.File{master}
	cmd.Start()

	errchan := make(chan error, 1)
	go func() {
		errchan <- cmd.Wait()
		slave.Close()
	}()

	// poll until urxvt is done starting
	for {
		rows, _, err := pty.Getsize(slave)
		if err != nil {
			slave.Close()
			return nil, nil, nil, err
		}
		if rows > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	return slave, func() { slave.Close() }, errchan, nil
}

func TermboxUrxvt(width, height int, args ...string) (func() error, error) {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 25
	}

	pty, fini, errch, err := UrxvtPty(append(
		[]string{
			"+sb",
			"-geometry", fmt.Sprintf("%dx%d", width, height),
		}, args...)...)

	if err != nil {
		return nil, err
	}

	errch2 := make(chan error)

	go func() {
		err := <-errch
		if err != nil {
			log.Println("ERROR in urxvt:", err)
		}
		if termbox.IsInit {
			termbox.Interrupt()
		}
		errch2 <- err
	}()

	origTerminalDevice := termbox.TerminalDevice
	termbox.TerminalDevice = pty.Name()

	if err := termbox.Init(); err != nil {
		termbox.TerminalDevice = origTerminalDevice
		return nil, err
	}

	return func() error {
		termbox.Close()
		termbox.TerminalDevice = origTerminalDevice
		fini()
		return <-errch2
	}, nil
}
