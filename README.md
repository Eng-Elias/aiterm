# aiterm

**AI-powered terminal command generator** — describe what you want in plain English, get the shell command instantly.

```bash
aiterm "find all PDF files modified in the last 7 days"
```

---

## Features

- **Simple CLI** — Just run `aiterm "your description"` and get a command
- **Auto-Detect OS** — Automatically generates commands for your current platform
- **Target Any OS** — Use `-t win`, `-t linux`, or `-t mac` to generate for other platforms
- **Command Confirmation** — Review the generated command before it runs
- **Headless Mode** — `aiterm generate "..."` for scripting and piping
- **OpenAI-Compatible** — Works with OpenAI, LiteLLM, Ollama, and any compatible endpoint
- **Cross-Platform** — Linux, macOS, and Windows support
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

### 2. Generate a Command

```bash
aiterm "list all files larger than 100MB"
```

aiterm will:
1. Send your description to the AI
2. Display the generated command
3. Ask for confirmation (`Y/n`)
4. Execute the command if you confirm

---

## Usage

### Basic Usage

```bash
# Auto-detects your OS and generates the right command
aiterm "show disk usage sorted by size"

# Target a specific OS
aiterm "find all log files" -t linux
aiterm "list running processes" -t mac
aiterm "check open ports" -t win
```

### Target OS Flag (`-t`)

| Flag Value       | Target         | Shell      |
|------------------|----------------|------------|
| `win`, `windows` | Windows        | PowerShell |
| `linux`          | Linux          | bash       |
| `mac`, `macos`   | macOS          | zsh        |
| *(omitted)*      | Auto-detected  | Auto       |

### Headless Mode

```bash
aiterm generate "list all docker containers"
# Output: docker ps -a
```

Generates a command and prints it to stdout without confirmation. Useful for scripting:

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
| `shell`        | Shell hint for prompt context                        | `auto`                                           |

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

### "AI not configured"

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

### Debug Mode

Enable debug logging for troubleshooting:

```bash
aiterm "your prompt" --debug
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
│   ├── root.go                # Root command & CLI logic
│   ├── config.go              # Config subcommands
│   ├── generate.go            # Headless generation
│   ├── setup.go               # Setup wizard
│   └── version.go             # Version command
├── internal/
│   ├── ai/
│   │   ├── client.go          # OpenAI API client
│   │   └── client_test.go     # API client tests
│   └── config/
│       ├── config.go          # Configuration management
│       └── config_test.go     # Config tests
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
