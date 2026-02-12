package shell

import (
	"runtime"
	"testing"
)

func TestDetectAuto(t *testing.T) {
	info, err := Detect("auto")
	if err != nil {
		t.Fatalf("Detect(auto) failed: %v", err)
	}

	if info.Path == "" {
		t.Error("expected non-empty shell path")
	}
	if info.Name == "" {
		t.Error("expected non-empty shell name")
	}

	t.Logf("Detected shell: %s (%s)", info.Name, info.Path)
}

func TestDetectEmpty(t *testing.T) {
	info, err := Detect("")
	if err != nil {
		t.Fatalf("Detect('') failed: %v", err)
	}

	if info.Path == "" {
		t.Error("expected non-empty shell path")
	}
}

func TestDetectInvalidShell(t *testing.T) {
	_, err := Detect("nonexistent_shell_xyz")
	if err == nil {
		t.Error("expected error for nonexistent shell")
	}
}

func TestShellName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/bin/bash", "bash"},
		{"/usr/bin/zsh", "zsh"},
		{"/bin/sh", "sh"},
	}

	if runtime.GOOS == "windows" {
		tests = []struct {
			path     string
			expected string
		}{
			{`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, "powershell"},
			{`C:\Program Files\PowerShell\7\pwsh.exe`, "pwsh"},
		}
	}

	for _, tt := range tests {
		result := shellName(tt.path)
		if result != tt.expected {
			t.Errorf("shellName(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestDetectPlatformDefault(t *testing.T) {
	info, err := Detect("auto")
	if err != nil {
		t.Skipf("skipping: no shell available: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		// Should find powershell or cmd
		validNames := map[string]bool{"pwsh": true, "powershell": true, "cmd": true}
		if !validNames[info.Name] {
			t.Errorf("unexpected Windows shell: %s", info.Name)
		}
	default:
		// Should find bash, zsh, or sh
		validNames := map[string]bool{"bash": true, "zsh": true, "sh": true}
		if !validNames[info.Name] {
			t.Errorf("unexpected Unix shell: %s", info.Name)
		}
	}
}
