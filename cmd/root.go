package cmd

import (
	"fmt"
	"os"

	"aiterm/internal/config"
	"aiterm/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Version is set at build time via ldflags.
var Version = "dev"

var debug bool

var rootCmd = &cobra.Command{
	Use:   "aiterm",
	Short: "AI-powered terminal command generator",
	Long: `aiterm is a terminal wrapper that lets you generate shell commands
using natural language. Press Ctrl+K in the shell to activate AI mode,
describe what you want, and aiterm will generate and optionally execute the command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running in a TTY
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return cmd.Help()
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// If API token is missing, run setup wizard
		if cfg.APIToken == "" {
			fmt.Println("No API token configured. Running setup wizard...")
			return runSetup()
		}

		return runTUI(cfg)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging to ~/.aiterm/debug.log")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runTUI starts the interactive TUI mode.
func runTUI(cfg *config.Config) error {
	model, err := tui.NewModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize TUI: %w", err)
	}
	defer model.Cleanup()

	// Set raw terminal for shell pass-through
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
