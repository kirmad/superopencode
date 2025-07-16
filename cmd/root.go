package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/kirmad/superopencode/internal/app"
	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/db"
	"github.com/kirmad/superopencode/internal/format"
	"github.com/kirmad/superopencode/internal/llm/agent"
	"github.com/kirmad/superopencode/internal/logging"
	"github.com/kirmad/superopencode/internal/pubsub"
	"github.com/kirmad/superopencode/internal/tui"
	"github.com/kirmad/superopencode/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "opencode",
	Short: "Terminal-based AI assistant for software development",
	Long: `OpenCode is a powerful terminal-based AI assistant that helps with software development tasks.
It provides an interactive chat interface with AI capabilities, code analysis, and LSP integration
to assist developers in writing, debugging, and understanding code directly from the terminal.`,
	Example: `
  # Run in interactive mode
  opencode

  # Run with debug logging
  opencode -d

  # Run with debug logging in a specific directory
  opencode -d -c /path/to/project

  # Run with automatic todo completion enabled
  opencode --auto-complete-todos

  # Run with custom max todo continuations
  opencode --auto-complete-todos --max-todo-continuations 20

  # Print version
  opencode -v

  # Run a single non-interactive prompt
  opencode -p "Explain the use of context in Go"

  # Run a single non-interactive prompt with JSON output format
  opencode -p "Explain the use of context in Go" -f json
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flag("help").Changed {
			cmd.Help()
			return nil
		}
		if cmd.Flag("version").Changed {
			fmt.Println(version.Version)
			return nil
		}

		debug, _ := cmd.Flags().GetBool("debug")
		detailedLog, _ := cmd.Flags().GetBool("detailed-log")
		cwd, _ := cmd.Flags().GetString("cwd")
		prompt, _ := cmd.Flags().GetString("prompt")
		outputFormat, _ := cmd.Flags().GetString("output-format")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoCompleteTodos, _ := cmd.Flags().GetBool("auto-complete-todos")
		maxTodoContinuations, _ := cmd.Flags().GetInt("max-todo-continuations")
		dangerouslySkipPermissions, _ := cmd.Flags().GetBool("dangerously-skip-permissions")

		if !dangerouslySkipPermissions && os.Getenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS") == "true" {
			dangerouslySkipPermissions = true
		}
		if !format.IsValid(outputFormat) {
			return fmt.Errorf("invalid format option: %s\n%s", outputFormat, format.GetHelpText())
		}
		if cwd != "" {
			err := os.Chdir(cwd)
			if err != nil {
				return fmt.Errorf("failed to change directory: %v", err)
			}
		}
		if autoCompleteTodos {
			os.Setenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS", "true")
		}
		if maxTodoContinuations != 10 {
			os.Setenv("SUPEROPENCODE_MAX_TODO_CONTINUATIONS", fmt.Sprintf("%d", maxTodoContinuations))
		}
		if cwd == "" {
			c, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %v", err)
			}
			cwd = c
		}
		_, err := config.Load(cwd, debug, detailedLog)
		if err != nil {
			return err
		}
		conn, err := db.Connect()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		app, err := app.New(ctx, conn)
		if err != nil {
			logging.Error("Failed to create app: %v", err)
			return err
		}
		defer app.Shutdown()
		initMCPTools(ctx, app)
		if dangerouslySkipPermissions {
			if err := validateDangerousMode(); err != nil {
				return fmt.Errorf("dangerous mode validation failed: %w", err)
			}
			printDangerousWarning()
			app.SetDangerousMode(true)
		}
		if prompt != "" {
			return app.RunNonInteractive(ctx, prompt, outputFormat, quiet, dangerouslySkipPermissions)
		}
		zone.NewGlobal()
		program := tea.NewProgram(
			tui.New(app),
			tea.WithAltScreen(),
		)
		ch, cancelSubs := setupSubscriptions(app, ctx)
		tuiCtx, tuiCancel := context.WithCancel(ctx)
		var tuiWg sync.WaitGroup
		tuiWg.Add(1)
		go func() {
			defer tuiWg.Done()
			defer logging.RecoverPanic("TUI-message-handler", func() {
				attemptTUIRecovery(program)
			})
			for {
				select {
				case <-tuiCtx.Done():
					logging.Info("TUI message handler shutting down")
					return
				case msg, ok := <-ch:
					if !ok {
						logging.Info("TUI message channel closed")
						return
					}
					program.Send(msg)
				}
			}
		}()
		cleanup := func() {
			app.Shutdown()
			cancelSubs()
			tuiCancel()
			tuiWg.Wait()
			logging.Info("All goroutines cleaned up")
		}
		result, err := program.Run()
		cleanup()
		if err != nil {
			logging.Error("TUI error: %v", err)
			return fmt.Errorf("TUI error: %v", err)
		}
		logging.Info("TUI exited with result: %v", result)
		return nil
	},
}

func init() {
	rootCmd.Flags().BoolP("help", "h", false, "Help")
	rootCmd.Flags().BoolP("version", "v", false, "Version")
	rootCmd.Flags().BoolP("debug", "d", false, "Debug")
	rootCmd.Flags().Bool("detailed-log", false, "Enable detailed logging of copilot requests/responses")
	rootCmd.Flags().StringP("cwd", "c", "", "Current working directory")
	rootCmd.Flags().StringP("prompt", "p", "", "Prompt to run in non-interactive mode")
	rootCmd.Flags().StringP("output-format", "f", format.Text.String(),
		"Output format for non-interactive mode (text, json)")
	rootCmd.Flags().BoolP("quiet", "q", false, "Hide spinner in non-interactive mode")
	rootCmd.Flags().Bool("auto-complete-todos", false, "Enable automatic completion of todos until all are done")
	rootCmd.Flags().Int("max-todo-continuations", 10, "Maximum number of todo continuation attempts per session")
	rootCmd.Flags().Bool("dangerously-skip-permissions", false, "Skip all permission checks for this session (USE WITH EXTREME CAUTION)")
	rootCmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return format.SupportedFormats, cobra.ShellCompDirectiveNoFileComp
	})
}

func attemptTUIRecovery(program *tea.Program) {
	logging.Info("Attempting to recover TUI after panic")
	program.Quit()
}

func initMCPTools(ctx context.Context, app *app.App) {
	go func() {
		defer logging.RecoverPanic("MCP-goroutine", nil)
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		agent.GetMcpTools(ctxWithTimeout, app.Permissions)
		logging.Info("MCP message handling goroutine exiting")
	}()
}

func setupSubscriber[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	subscriber func(context.Context) <-chan pubsub.Event[T],
	outputCh chan<- tea.Msg,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer logging.RecoverPanic(fmt.Sprintf("subscription-%s", name), nil)
		subCh := subscriber(ctx)
		for {
			select {
			case event, ok := <-subCh:
				if !ok {
					logging.Info("subscription channel closed", "name", name)
					return
				}
				var msg tea.Msg = event
				select {
				case outputCh <- msg:
				case <-time.After(2 * time.Second):
					logging.Warn("message dropped due to slow consumer", "name", name)
				case <-ctx.Done():
					logging.Info("subscription cancelled", "name", name)
					return
				}
			case <-ctx.Done():
				logging.Info("subscription cancelled", "name", name)
				return
			}
		}
	}()
}

func setupSubscriptions(app *app.App, parentCtx context.Context) (chan tea.Msg, func()) {
	ch := make(chan tea.Msg, 100)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(parentCtx)
	setupSubscriber(ctx, &wg, "logging", logging.Subscribe, ch)
	setupSubscriber(ctx, &wg, "sessions", app.Sessions.Subscribe, ch)
	setupSubscriber(ctx, &wg, "messages", app.Messages.Subscribe, ch)
	setupSubscriber(ctx, &wg, "permissions", app.Permissions.Subscribe, ch)
	setupSubscriber(ctx, &wg, "coderAgent", app.CoderAgent.Subscribe, ch)
	cleanupFunc := func() {
		logging.Info("Cancelling all subscriptions")
		cancel()
		waitCh := make(chan struct{})
		go func() {
			defer logging.RecoverPanic("subscription-cleanup", nil)
			wg.Wait()
			close(waitCh)
		}()
		select {
		case <-waitCh:
			logging.Info("All subscription goroutines completed successfully")
			close(ch)
		case <-time.After(5 * time.Second):
			logging.Warn("Timed out waiting for some subscription goroutines to complete")
			close(ch)
		}
	}
	return ch, cleanupFunc
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func validateDangerousMode() error {
	if os.Getenv("PRODUCTION") == "true" {
		return fmt.Errorf("dangerous mode disabled in production")
	}
	if os.Getenv("CI") == "true" || os.Getenv("CONTINUOUS_INTEGRATION") == "true" {
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Dangerous mode enabled in CI environment\n")
	}
	if os.Geteuid() == 0 {
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Running dangerous mode as root user\n")
	}
	return nil
}

func printDangerousWarning() {
	fmt.Fprintf(os.Stderr, "⚠️  DANGEROUS MODE: All permission checks will be bypassed\n")
	fmt.Fprintf(os.Stderr, "   This session has unrestricted system access\n")
}
