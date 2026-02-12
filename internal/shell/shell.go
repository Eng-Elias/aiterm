package shell

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Info holds details about the detected shell.
type Info struct {
	Path string // absolute path to shell binary
	Name string // short name (bash, zsh, powershell, etc.)
}

// Detect determines the appropriate shell for the current platform.
// If shellOverride is non-empty and not "auto", it is used directly.
func Detect(shellOverride string) (*Info, error) {
	if shellOverride != "" && shellOverride != "auto" {
		return resolveShell(shellOverride)
	}

	if runtime.GOOS == "windows" {
		return detectWindows()
	}
	return detectUnix()
}

// detectUnix checks $SHELL and falls back to bash then sh.
func detectUnix() (*Info, error) {
	shellEnv := os.Getenv("SHELL")
	if shellEnv != "" {
		name := shellName(shellEnv)
		return &Info{Path: shellEnv, Name: name}, nil
	}

	// Fallback chain: bash → sh
	for _, candidate := range []string{"bash", "sh"} {
		p, err := exec.LookPath(candidate)
		if err == nil {
			return &Info{Path: p, Name: candidate}, nil
		}
	}

	return nil, fmt.Errorf("no suitable shell found on this system")
}

// detectWindows tries PowerShell then falls back to cmd.
func detectWindows() (*Info, error) {
	// Try pwsh (PowerShell 7+) first, then powershell (Windows PowerShell)
	for _, candidate := range []string{"pwsh", "powershell", "cmd"} {
		p, err := exec.LookPath(candidate)
		if err == nil {
			return &Info{Path: p, Name: candidate}, nil
		}
	}

	return nil, fmt.Errorf("no suitable shell found on this system")
}

// resolveShell resolves a user-specified shell override.
func resolveShell(shell string) (*Info, error) {
	p, err := exec.LookPath(shell)
	if err != nil {
		return nil, fmt.Errorf("shell %q not found: %w", shell, err)
	}
	return &Info{Path: p, Name: shellName(p)}, nil
}

// shellName extracts the short name from a shell path.
func shellName(path string) string {
	parts := strings.Split(path, "/")
	if runtime.GOOS == "windows" {
		parts = strings.Split(path, "\\")
	}
	name := parts[len(parts)-1]
	// Remove .exe suffix on Windows
	name = strings.TrimSuffix(name, ".exe")
	return name
}
