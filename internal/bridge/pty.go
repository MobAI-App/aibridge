//go:build !windows

package bridge

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type PTY struct {
	cmd      *exec.Cmd
	ptmx     *os.File
	mu       sync.Mutex
	closed   bool
	oldState *term.State

	echoMu        sync.Mutex
	echoWaiting   string
	echoBuf       string
	echoCh        chan struct{}
	injectDelayMs int
}

func NewPTY(command string, args []string, injectDelayMs int) *PTY {
	cmd := exec.Command(command, args...)
	return &PTY{
		cmd:           cmd,
		injectDelayMs: injectDelayMs,
	}
}

func (p *PTY) Start(outputCallback func(line string)) error {
	ptmx, err := pty.Start(p.cmd)
	if err != nil {
		return err
	}
	p.ptmx = ptmx

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	p.oldState = oldState

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if p.closed {
				return
			}
			_ = pty.InheritSize(os.Stdin, p.ptmx)
		}
	}()
	ch <- syscall.SIGWINCH

	go func() {
		reader := bufio.NewReader(p.ptmx)
		lineBuffer := make([]byte, 0, 1024)

		buf := make([]byte, 4096)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				return
			}

			_, _ = os.Stdout.Write(buf[:n])

			p.checkEcho(string(buf[:n]))

			for i := 0; i < n; i++ {
				b := buf[i]
				if b == '\n' || b == '\r' {
					if len(lineBuffer) > 0 {
						outputCallback(string(lineBuffer))
						lineBuffer = lineBuffer[:0]
					}
				} else {
					lineBuffer = append(lineBuffer, b)
				}
			}

			if len(lineBuffer) > 0 {
				outputCallback(string(lineBuffer))
			}
		}
	}()

	go func() {
		_, _ = io.Copy(p.ptmx, os.Stdin)
	}()

	return nil
}

func (p *PTY) InjectText(text string, sendEnter bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return io.ErrClosedPipe
	}

	if sendEnter {
		p.echoMu.Lock()
		p.echoWaiting = text
		p.echoCh = make(chan struct{})
		ch := p.echoCh
		p.echoMu.Unlock()

		_, err := p.ptmx.WriteString(text)
		if err != nil {
			p.clearEchoWait()
			return err
		}

		select {
		case <-ch:
			time.Sleep(time.Duration(p.injectDelayMs) * time.Millisecond)
		case <-time.After(2 * time.Second):
		}
		p.clearEchoWait()

		_, err = p.ptmx.Write([]byte{'\r'})
		return err
	}

	_, err := p.ptmx.WriteString(text)
	return err
}

func (p *PTY) clearEchoWait() {
	p.echoMu.Lock()
	p.echoWaiting = ""
	p.echoBuf = ""
	p.echoCh = nil
	p.echoMu.Unlock()
}

func (p *PTY) checkEcho(data string) {
	p.echoMu.Lock()
	defer p.echoMu.Unlock()

	if p.echoWaiting == "" || p.echoCh == nil {
		return
	}

	p.echoBuf += data
	needle := p.echoWaiting
	if len(needle) > 20 {
		needle = needle[len(needle)-20:]
	}
	if strings.Contains(p.echoBuf, needle) {
		close(p.echoCh)
		p.echoWaiting = ""
		p.echoBuf = ""
		p.echoCh = nil
	}
}

func (p *PTY) Wait() error {
	return p.cmd.Wait()
}

func (p *PTY) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	if p.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), p.oldState)
	}

	if p.ptmx != nil {
		return p.ptmx.Close()
	}
	return nil
}

func (p *PTY) Running() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	return p.cmd.ProcessState == nil
}
