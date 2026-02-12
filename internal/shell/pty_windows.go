//go:build windows

package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"golang.org/x/term"
)

// Session represents a running shell session on Windows.
// Windows does not have native PTY support via creack/pty,
// so we use stdin/stdout pipes as a fallback.
type Session struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	mu      sync.Mutex
	done    chan struct{}
	oldTerm *term.State
}

// StartSession spawns a new shell process with piped I/O on Windows.
func StartSession(info *Info) (*Session, error) {
	cmd := exec.Command(info.Path)
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	s := &Session{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		done:   make(chan struct{}),
	}

	return s, nil
}

// SetRawTerminal puts the host terminal into raw mode for pass-through.
func (s *Session) SetRawTerminal() error {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		// On Windows, raw mode may not be available for all terminal types.
		// Continue without raw mode.
		return nil
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

// Read reads from the shell's stdout.
func (s *Session) Read(p []byte) (int, error) {
	return s.stdout.Read(p)
}

// Write writes to the shell's stdin.
func (s *Session) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stdin.Write(p)
}

// WriteString writes a string to the shell's stdin.
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
	if s.stdin != nil {
		_ = s.stdin.Close()
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

// StreamOutput copies shell output to the given writer (usually os.Stdout).
func (s *Session) StreamOutput(w io.Writer) {
	buf := make([]byte, 4096)
	for {
		n, err := s.stdout.Read(buf)
		if n > 0 {
			_, _ = w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// HandleResize is a no-op on Windows (no SIGWINCH).
func (s *Session) HandleResize() {
	// Windows does not support SIGWINCH.
	// Terminal resize is handled by the Windows console subsystem.
}
