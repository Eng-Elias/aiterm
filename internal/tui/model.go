package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"aiterm/internal/ai"
	"aiterm/internal/config"
	"aiterm/internal/shell"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents the current input mode of the TUI.
type Mode int

const (
	ModeShell   Mode = iota // Normal shell pass-through
	ModeAIInput             // Accepting AI prompt input
	ModeLoading             // Waiting for AI response
	ModeConfirm             // Showing generated command, awaiting confirmation
)

// Model is the bubbletea model for the TUI.
type Model struct {
	cfg       *config.Config
	client    *ai.Client
	session   *shell.Session
	shellInfo *shell.Info

	mode         Mode
	aiInput      string // user's natural language input
	generatedCmd string // AI-generated command
	errMsg       string // error message to display
	shellOutput  string // accumulated shell output for display
	width        int
	height       int
	quitting     bool
}

// Styles
var (
	aiInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("42")). // green
			Padding(0, 1).
			MarginTop(1)

	generatedCmdStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("17")). // dark blue
				Foreground(lipgloss.Color("15")). // white
				Padding(0, 1).
				Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")). // red
			Foreground(lipgloss.Color("196")).
			Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	modeIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("42")).
				Bold(true).
				Padding(0, 1)
)

// shellOutputMsg carries new output from the shell.
type shellOutputMsg struct {
	data string
}

// shellExitMsg signals the shell has exited.
type shellExitMsg struct{}

// aiResponseMsg carries the AI-generated command.
type aiResponseMsg struct {
	command string
	err     error
}

// NewModel creates and initializes the TUI model.
func NewModel(cfg *config.Config) (*Model, error) {
	shellInfo, err := shell.Detect(cfg.Shell)
	if err != nil {
		return nil, fmt.Errorf("shell detection failed: %w", err)
	}

	session, err := shell.StartSession(shellInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to start shell session: %w", err)
	}

	m := &Model{
		cfg:       cfg,
		client:    ai.NewClient(cfg),
		session:   session,
		shellInfo: shellInfo,
		mode:      ModeShell,
		width:     80,
		height:    24,
	}

	return m, nil
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.readShellOutput(),
		m.waitForShellExit(),
	)
}

// readShellOutput returns a command that reads from the shell PTY.
func (m *Model) readShellOutput() tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		n, err := m.session.Read(buf)
		if err != nil {
			if err == io.EOF {
				return shellExitMsg{}
			}
			return shellExitMsg{}
		}
		return shellOutputMsg{data: string(buf[:n])}
	}
}

// waitForShellExit returns a command that waits for the shell to exit.
func (m *Model) waitForShellExit() tea.Cmd {
	return func() tea.Msg {
		_ = m.session.Wait()
		return shellExitMsg{}
	}
}

// generateCommand returns a command that calls the AI API.
func (m *Model) generateCommand(description string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*1000*1000*1000) // 30s
		defer cancel()

		cmd, err := m.client.GenerateCommand(ctx, description)
		return aiResponseMsg{command: cmd, err: err}
	}
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case shellOutputMsg:
		// Write shell output directly to stdout for real-time display
		_, _ = os.Stdout.WriteString(msg.data)
		// Keep reading
		return m, m.readShellOutput()

	case shellExitMsg:
		m.quitting = true
		return m, tea.Quit

	case aiResponseMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.mode = ModeShell
			return m, nil
		}
		m.generatedCmd = msg.command
		m.mode = ModeConfirm
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keyboard input based on the current mode.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {

	case ModeShell:
		return m.handleShellKey(msg)

	case ModeAIInput:
		return m.handleAIInputKey(msg)

	case ModeLoading:
		// Only allow Escape to cancel during loading
		if msg.Type == tea.KeyEscape {
			m.mode = ModeShell
			m.aiInput = ""
			return m, nil
		}
		return m, nil

	case ModeConfirm:
		return m.handleConfirmKey(msg)
	}

	return m, nil
}

// handleShellKey processes keys in normal shell mode.
func (m *Model) handleShellKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+K: activate AI mode
	if msg.Type == tea.KeyCtrlK {
		if err := m.cfg.Validate(); err != nil {
			m.errMsg = "AI not configured — run 'aiterm setup' to set your API token"
			return m, nil
		}
		m.mode = ModeAIInput
		m.aiInput = ""
		m.errMsg = ""
		m.generatedCmd = ""
		return m, nil
	}

	// Ctrl+C: pass through to shell (it handles SIGINT)
	if msg.Type == tea.KeyCtrlC {
		_ = m.session.WriteString("\x03")
		return m, nil
	}

	// Ctrl+D: exit
	if msg.Type == tea.KeyCtrlD {
		_ = m.session.WriteString("\x04")
		return m, nil
	}

	// Pass all other keys through to the shell
	var input string
	switch msg.Type {
	case tea.KeyEnter:
		input = "\n"
	case tea.KeyTab:
		input = "\t"
	case tea.KeyBackspace:
		input = "\x7f"
	case tea.KeySpace:
		input = " "
	case tea.KeyUp:
		input = "\x1b[A"
	case tea.KeyDown:
		input = "\x1b[B"
	case tea.KeyRight:
		input = "\x1b[C"
	case tea.KeyLeft:
		input = "\x1b[D"
	default:
		if msg.Type == tea.KeyRunes {
			input = string(msg.Runes)
		}
	}

	if input != "" {
		_ = m.session.WriteString(input)
	}

	return m, nil
}

// handleAIInputKey processes keys in AI input mode.
func (m *Model) handleAIInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = ModeShell
		m.aiInput = ""
		return m, nil

	case tea.KeyEnter:
		if strings.TrimSpace(m.aiInput) == "" {
			return m, nil
		}
		m.mode = ModeLoading
		return m, m.generateCommand(m.aiInput)

	case tea.KeyBackspace:
		if len(m.aiInput) > 0 {
			m.aiInput = m.aiInput[:len(m.aiInput)-1]
		}
		return m, nil

	case tea.KeyRunes:
		m.aiInput += string(msg.Runes)
		return m, nil

	case tea.KeySpace:
		m.aiInput += " "
		return m, nil
	}

	return m, nil
}

// handleConfirmKey processes keys in command confirmation mode.
func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Execute the generated command in the shell
		_ = m.session.WriteString(m.generatedCmd + "\n")
		m.mode = ModeShell
		m.generatedCmd = ""
		m.aiInput = ""
		return m, nil

	case tea.KeyEscape:
		// Discard command
		m.mode = ModeShell
		m.generatedCmd = ""
		m.aiInput = ""
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var sections []string

	// Error message (if any)
	if m.errMsg != "" {
		errBox := errorStyle.Render("Error: " + m.errMsg)
		sections = append(sections, errBox)
	}

	switch m.mode {
	case ModeAIInput:
		sections = append(sections, m.aiInputView())

	case ModeLoading:
		sections = append(sections, m.loadingView())

	case ModeConfirm:
		sections = append(sections, m.confirmView())
	}

	if len(sections) == 0 {
		return ""
	}

	return "\r\n" + strings.Join(sections, "\r\n")
}

// aiInputView renders the AI input prompt.
func (m *Model) aiInputView() string {
	indicator := modeIndicatorStyle.Render(" AI Mode - Press Esc to cancel ")

	inputText := m.aiInput
	if inputText == "" {
		inputText = hintStyle.Render("Describe the command you want...")
	}

	inputBox := aiInputStyle.Render(inputText + "█")

	return indicator + "\r\n" + inputBox
}

// loadingView renders the loading indicator.
func (m *Model) loadingView() string {
	indicator := modeIndicatorStyle.Render(" AI Mode ")
	loading := hintStyle.Render("⣾ Generating command...")
	return indicator + "\r\n" + loading
}

// confirmView renders the generated command and confirmation prompt.
func (m *Model) confirmView() string {
	indicator := modeIndicatorStyle.Render(" AI Mode ")
	label := labelStyle.Render("Generated Command:")
	cmd := generatedCmdStyle.Render(m.generatedCmd)
	hint := hintStyle.Render("Press Enter to execute, Esc to cancel")

	return indicator + "\r\n" + label + "\r\n" + cmd + "\r\n" + hint
}

// Cleanup releases all resources held by the model.
func (m *Model) Cleanup() {
	if m.session != nil {
		_ = m.session.Close()
	}
}
