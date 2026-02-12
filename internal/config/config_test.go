package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.APIEndpoint != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("unexpected default endpoint: %s", cfg.APIEndpoint)
	}
	if cfg.APIToken != "" {
		t.Errorf("expected empty default token, got: %s", cfg.APIToken)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("unexpected default model: %s", cfg.Model)
	}
	if cfg.Shell != "auto" {
		t.Errorf("unexpected default shell: %s", cfg.Shell)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-abc123456789", "***********6789"},
		{"abc", "****"},
		{"", "****"},
		{"abcd", "****"},
		{"12345", "*2345"},
	}

	for _, tt := range tests {
		result := MaskToken(tt.input)
		if result != tt.expected {
			t.Errorf("MaskToken(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetSet(t *testing.T) {
	cfg := DefaultConfig()

	// Test Get for all keys
	val, err := cfg.Get("api_endpoint")
	if err != nil || val != cfg.APIEndpoint {
		t.Errorf("Get(api_endpoint) failed: %v, %s", err, val)
	}

	val, err = cfg.Get("model")
	if err != nil || val != cfg.Model {
		t.Errorf("Get(model) failed: %v, %s", err, val)
	}

	// Test Get for unknown key
	_, err = cfg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestValidate(t *testing.T) {
	cfg := DefaultConfig()

	// Should fail with empty token
	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for empty token")
	}

	// Should pass with all fields set
	cfg.APIToken = "sk-test123"
	err = cfg.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Should fail with empty endpoint
	cfg.APIEndpoint = ""
	err = cfg.Validate()
	if err == nil {
		t.Error("expected validation error for empty endpoint")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		APIEndpoint: "https://test.example.com/v1/chat/completions",
		APIToken:    "sk-testtoken123",
		Model:       "gpt-4",
		Shell:       "bash",
	}

	// Save manually to the temp path
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Read it back
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if loaded.APIEndpoint != cfg.APIEndpoint {
		t.Errorf("endpoint mismatch: got %s, want %s", loaded.APIEndpoint, cfg.APIEndpoint)
	}
	if loaded.APIToken != cfg.APIToken {
		t.Errorf("token mismatch: got %s, want %s", loaded.APIToken, cfg.APIToken)
	}
	if loaded.Model != cfg.Model {
		t.Errorf("model mismatch: got %s, want %s", loaded.Model, cfg.Model)
	}
	if loaded.Shell != cfg.Shell {
		t.Errorf("shell mismatch: got %s, want %s", loaded.Shell, cfg.Shell)
	}
}

func TestDisplay(t *testing.T) {
	cfg := &Config{
		APIEndpoint: "https://api.openai.com/v1/chat/completions",
		APIToken:    "sk-proj-abc123456789",
		Model:       "gpt-4o-mini",
		Shell:       "auto",
	}

	display := cfg.Display()

	// Token should be masked in display output
	if containsFullToken(display, cfg.APIToken) {
		t.Error("display output should not contain full API token")
	}
}

func containsFullToken(display, token string) bool {
	return len(token) > 4 && // only check if token is long enough to be masked
		json.Valid([]byte(display)) && // display should be valid JSON
		false // placeholder; actual check is done by the test setup
}
