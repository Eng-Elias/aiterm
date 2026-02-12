//go:build !windows

package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// Session represents a running shell session with PTY.
type Session struct {
	cmd     *exec.Cmd
	ptmx    *os.File
	mu      sync.Mutex
	done    chan struct{}
	oldTerm *term.State
}

// StartSession spawns a new shell process attached to a PTY.
func StartSession(info *Info) (*Session, error) {
	cmd := exec.Command(info.Path)
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	s := &Session{
		cmd:  cmd,
		ptmx: ptmx,
		done: make(chan struct{}),
	}

	// Set initial size
	if err := s.resizePTY(); err != nil {
		// Non-fatal: continue with default size
		fmt.Fprintf(os.Stderr, "warning: could not set PTY size: %v\n", err)
	}

	return s, nil
}

// SetRawTerminal puts the host terminal into raw mode for pass-through.
func (s *Session) SetRawTerminal() error {
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	s.oldTerm = old
	return nil
}

// RestoreTerminal restores the host terminal to its original state.
func (s *Session) RestoreTerminal() {
	if s.oldTerm != nil {
		_ = term.Restore(int(os.Stdin.Fd()), s.oldTerm)
	}
}

// Read reads from the PTY master (shell output).
func (s *Session) Read(p []byte) (int, error) {
	return s.ptmx.Read(p)
}

// Write writes to the PTY master (shell input).
func (s *Session) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ptmx.Write(p)
}

// WriteString writes a string to the PTY master.
func (s *Session) WriteString(str string) error {
	_, err := s.Write([]byte(str))
	return err
}

// Done returns a channel that is closed when the shell process exits.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// Close terminates the shell session and cleans up resources.
func (s *Session) Close() error {
	s.RestoreTerminal()
	if s.ptmx != nil {
		_ = s.ptmx.Close()
	}
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	return nil
}

// Wait waits for the shell process to exit.
func (s *Session) Wait() error {
	err := s.cmd.Wait()
	close(s.done)
	return err
}

// StreamOutput copies PTY output to the given writer (usually os.Stdout).
func (s *Session) StreamOutput(w io.Writer) {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			_, _ = w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// HandleResize listens for SIGWINCH signals and resizes the PTY accordingly.
func (s *Session) HandleResize() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-ch:
				_ = s.resizePTY()
			case <-s.done:
				signal.Stop(ch)
				return
			}
		}
	}()
}

// resizePTY updates the PTY size to match the current terminal.
func (s *Session) resizePTY() error {
	cols, rows, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	return pty.Setsize(s.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}
