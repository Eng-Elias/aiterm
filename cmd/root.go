package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"aiterm/internal/ai"
	"aiterm/internal/config"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var (
	debug      bool
	targetType string
)

var rootCmd = &cobra.Command{
	Use:   "aiterm [prompt]",
	Short: "AI-powered terminal command generator",
	Long: `aiterm generates shell commands from natural language descriptions.

Examples:
  aiterm "list all files larger than 100MB"
  aiterm "find all PDFs modified in the last 7 days" -t linux
  aiterm "show disk usage sorted by size" -t mac`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		prompt := strings.Join(args, " ")
		return runGenerate(prompt, targetType)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging to ~/.aiterm/debug.log")
	rootCmd.Flags().StringVarP(&targetType, "type", "t", "", "Target OS type: win, linux, mac (auto-detected if omitted)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runGenerate handles the main flow: generate command → confirm → execute.
func runGenerate(prompt, target string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Println("AI not configured. Run 'aiterm setup' first.")
		return err
	}

	// Resolve target for display
	osName, shellType := ai.ResolveTargetOS(target)
	fmt.Printf("\033[90mTarget: %s (%s)\033[0m\n", osName, shellType)
	fmt.Printf("\033[90mGenerating command...\033[0m\n")

	client := ai.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	command, err := client.GenerateCommand(ctx, prompt, target)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Display the generated command
	fmt.Printf("\n\033[1;34m❯ %s\033[0m\n\n", command)

	// Ask for confirmation
	fmt.Print("\033[33mRun this command? [Y/n]: \033[0m")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "" && answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Execute the command
	return executeCommand(command)
}

// executeCommand runs a shell command and streams output to stdout/stderr.
func executeCommand(command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "sh"
		}
		cmd = exec.Command(shell, "-c", command)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
