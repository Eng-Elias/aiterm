# aiterm

**AI-powered terminal command generator** — describe what you want in plain English, get the shell command instantly.

`aiterm` wraps your shell with an AI overlay. Press `Ctrl+K` at any time, type a natural language description, and aiterm generates and (optionally) executes the corresponding shell command.

---

## Features

- **Interactive AI Mode** — Press `Ctrl+K` to activate, type a description, get a command
- **Shell Wrapper** — Transparent PTY pass-through; your shell works exactly as normal
- **Command Confirmation** — Always see the generated command before it runs
- **Headless Mode** — `aiterm generate "..."` for scripting and piping
- **OpenAI-Compatible** — Works with OpenAI, LiteLLM, Ollama, and any compatible endpoint
- **Cross-Platform** — Linux, macOS, and Windows support
- **Configurable** — JSON config with setup wizard
- **Secure** — API tokens masked in output, config files with restricted permissions

---

## Installation

### Download Binary

Download the latest release for your platform from the [Releases](https://github.com/yourusername/aiterm/releases) page:

| Platform       | Binary                      |
|----------------|-----------------------------|
| Linux (AMD64)  | `aiterm-linux-amd64`        |
| Linux (ARM64)  | `aiterm-linux-arm64`        |
| macOS (Intel)  | `aiterm-darwin-amd64`       |
| macOS (Apple)  | `aiterm-darwin-arm64`       |
| Windows        | `aiterm-windows-amd64.exe`  |

### From Source

```bash
git clone https://github.com/yourusername/aiterm.git
cd aiterm
make build
```

### Install System-Wide

```bash
make install
```

This copies the binary to `/usr/local/bin` (Unix) or `%PROGRAMFILES%` (Windows).

---

## Quick Start

### 1. Run Setup

```bash
aiterm setup
```

You will be prompted for:
- **API Endpoint** (default: OpenAI)
- **API Token** (your OpenAI API key)
- **Model** (default: `gpt-4o-mini`)

The wizard will test your connection and save the configuration.

### 2. Launch aiterm

```bash
aiterm
```

This opens your default shell inside aiterm. Use it normally — it's a transparent wrapper.

### 3. Generate a Command

1. Press **Ctrl+K** to activate AI mode
2. Type a description: `find all PDF files modified in the last 7 days`
3. Press **Enter** to send to the AI
4. Review the generated command
5. Press **Enter** to execute, or **Esc** to cancel

---

## Usage

### Interactive TUI Mode

```bash
aiterm
```

Launches the interactive shell wrapper. Use your shell normally; press `Ctrl+K` to enter AI mode.

### Headless Mode

```bash
aiterm generate "list all docker containers"
# Output: docker ps -a
```

Generates a command and prints it to stdout. Useful for scripting:

```bash
$(aiterm generate "count lines in all Python files")
```

### Configuration

```bash
# Show current config
aiterm config

# Get a specific value
aiterm config get model

# Set a value
aiterm config set model gpt-4

# Run setup wizard again
aiterm setup
```

### Version

```bash
aiterm version
```

---

## Keyboard Shortcuts

| Shortcut  | Context       | Action                                    |
|-----------|---------------|-------------------------------------------|
| `Ctrl+K`  | Shell mode    | Activate AI command generation mode       |
| `Escape`  | AI mode       | Cancel and return to normal shell         |
| `Enter`   | AI input      | Send description to AI                    |
| `Enter`   | Confirmation  | Execute the generated command             |
| `Escape`  | Confirmation  | Discard command and return to shell       |
| `Ctrl+C`  | Shell mode    | Send SIGINT to shell (normal behavior)    |
| `Ctrl+D`  | Shell mode    | Send EOF to shell (normal behavior)       |

---

## Configuration Reference

Configuration is stored in `~/.aiterm/config.json` (or `%USERPROFILE%\.aiterm\config.json` on Windows).

```json
{
  "api_endpoint": "https://api.openai.com/v1/chat/completions",
  "api_token": "sk-...",
  "model": "gpt-4o-mini",
  "shell": "auto"
}
```

| Key            | Description                                          | Default                                          |
|----------------|------------------------------------------------------|--------------------------------------------------|
| `api_endpoint` | OpenAI-compatible chat completions URL               | `https://api.openai.com/v1/chat/completions`     |
| `api_token`    | API bearer token                                     | *(required)*                                     |
| `model`        | Model name to use                                    | `gpt-4o-mini`                                    |
| `shell`        | Shell to use (`auto`, `bash`, `zsh`, `powershell`)   | `auto`                                           |

### Shell Auto-Detection

When `shell` is set to `auto`:
- **Linux/macOS**: Uses `$SHELL`, falls back to `bash` → `sh`
- **Windows**: Uses `pwsh` (PowerShell 7), falls back to `powershell` → `cmd`

---

## API Compatibility

aiterm works with any OpenAI-compatible chat completions endpoint:

| Provider   | Endpoint Example                                    |
|------------|-----------------------------------------------------|
| OpenAI     | `https://api.openai.com/v1/chat/completions`        |
| LiteLLM    | `http://localhost:4000/v1/chat/completions`          |
| Ollama     | `http://localhost:11434/v1/chat/completions`         |
| Azure      | `https://<resource>.openai.azure.com/openai/deployments/<model>/chat/completions?api-version=2024-02-01` |

---

## Troubleshooting

### "No API token configured"

Run `aiterm setup` to configure your API credentials.

### "Authentication failed"

Your API token is invalid or expired. Update it:

```bash
aiterm config set api_token sk-your-new-token
```

### "API request failed"

- Check your internet connection
- Verify the API endpoint is correct: `aiterm config get api_endpoint`
- For local endpoints (LiteLLM, Ollama), ensure the server is running

### "Rate limit exceeded"

Wait a moment and try again. Consider upgrading your API plan for higher limits.

### Shell not detected

Override the shell manually:

```bash
aiterm config set shell /usr/bin/bash
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
aiterm --debug
```

Logs are written to `~/.aiterm/debug.log`.

---

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build targets)

### Build

```bash
make build          # Build for current platform
make build-all      # Cross-compile for all platforms
make test           # Run tests
make dev            # Live reload with air
```

### Project Structure

```
aiterm/
├── main.go                    # Entry point
├── cmd/
│   ├── root.go                # Root command & TUI launcher
│   ├── config.go              # Config subcommands
│   ├── generate.go            # Headless generation
│   ├── setup.go               # Setup wizard
│   └── version.go             # Version command
├── internal/
│   ├── ai/
│   │   ├── client.go          # OpenAI API client
│   │   └── client_test.go     # API client tests
│   ├── config/
│   │   ├── config.go          # Configuration management
│   │   └── config_test.go     # Config tests
│   ├── shell/
│   │   ├── shell.go           # Shell detection
│   │   ├── pty_unix.go        # Unix PTY handling
│   │   ├── pty_windows.go     # Windows pipe handling
│   │   └── shell_test.go      # Shell tests
│   └── tui/
│       └── model.go           # Bubbletea TUI model
├── Makefile                   # Build targets
├── .air.toml                  # Live reload config
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Run tests: `make test`
5. Commit: `git commit -m "feat: add my feature"`
6. Push: `git push origin feature/my-feature`
7. Open a Pull Request

Please follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

---

## License

MIT License — see [LICENSE](LICENSE) for details.
