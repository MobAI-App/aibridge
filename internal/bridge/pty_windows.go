//go:build windows

package bridge

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/UserExistsError/conpty"
)

type PTY struct {
	cmd           *exec.Cmd
	cpty          *conpty.ConPty
	mu            sync.Mutex
	closed        bool
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
	cpty, err := conpty.Start(p.cmd.Path, conpty.ConPtyDimensions(120, 40))
	if err != nil {
		return err
	}
	p.cpty = cpty

	go func() {
		reader := bufio.NewReader(cpty)
		lineBuffer := make([]byte, 0, 1024)

		buf := make([]byte, 4096)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				return
			}

			_, _ = os.Stdout.Write(buf[:n])

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
		_, _ = io.Copy(cpty, os.Stdin)
	}()

	return nil
}

func (p *PTY) InjectText(text string, sendEnter bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return io.ErrClosedPipe
	}

	_, err := p.cpty.Write([]byte(text))
	if err != nil {
		return err
	}

	if sendEnter {
		time.Sleep(time.Duration(p.injectDelayMs) * time.Millisecond)
		_, err = p.cpty.Write([]byte{'\r'})
	}
	return err
}

func (p *PTY) Wait() error {
	if p.cpty == nil {
		return nil
	}
	_, err := p.cpty.Wait(context.Background())
	return err
}

func (p *PTY) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	if p.cpty != nil {
		return p.cpty.Close()
	}
	return nil
}

func (p *PTY) Running() bool {
	if p.cpty == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := p.cpty.Wait(ctx)
	return err == context.DeadlineExceeded
}
