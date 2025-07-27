package app

import (
	"context"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/logging"
	"github.com/kirmad/superopencode/internal/lsp"
	"github.com/kirmad/superopencode/internal/lsp/copilot"
	"github.com/kirmad/superopencode/internal/lsp/watcher"
)

func (app *App) initLSPClients(ctx context.Context) {
	cfg := config.Get()

	// Initialize Copilot client if enabled
	if cfg.Copilot.EnableCopilot {
		go app.createAndStartCopilotClient(ctx)
	}

	// Initialize LSP clients
	for name, clientConfig := range cfg.LSP {
		// Skip if this is gopls and Copilot is configured to replace it
		if cfg.Copilot.EnableCopilot && cfg.Copilot.ReplaceGopls && name == "gopls" {
			logging.Info("Skipping gopls initialization (replaced by Copilot)", "name", name)
			continue
		}

		// Start each client initialization in its own goroutine
		go app.createAndStartLSPClient(ctx, name, clientConfig.Command, clientConfig.Args...)
	}
	logging.Info("LSP clients initialization started in background")
}

// createAndStartLSPClient creates a new LSP client, initializes it, and starts its workspace watcher
func (app *App) createAndStartLSPClient(ctx context.Context, name string, command string, args ...string) {
	// Create a specific context for initialization with a timeout
	logging.Info("Creating LSP client", "name", name, "command", command, "args", args)

	// Create the LSP client
	lspClient, err := lsp.NewClient(ctx, command, args...)
	if err != nil {
		logging.Error("Failed to create LSP client for", name, err)
		return
	}

	// Create a longer timeout for initialization (some servers take time to start)
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Initialize with the initialization context
	_, err = lspClient.InitializeLSPClient(initCtx, config.WorkingDirectory())
	if err != nil {
		logging.Error("Initialize failed", "name", name, "error", err)
		// Clean up the client to prevent resource leaks
		lspClient.Close()
		return
	}

	// Wait for the server to be ready
	if err := lspClient.WaitForServerReady(initCtx); err != nil {
		logging.Error("Server failed to become ready", "name", name, "error", err)
		// We'll continue anyway, as some functionality might still work
		lspClient.SetServerState(lsp.StateError)
	} else {
		logging.Info("LSP server is ready", "name", name)
		lspClient.SetServerState(lsp.StateReady)
	}

	logging.Info("LSP client initialized", "name", name)

	// Create a child context that can be canceled when the app is shutting down
	watchCtx, cancelFunc := context.WithCancel(ctx)

	// Create a context with the server name for better identification
	watchCtx = context.WithValue(watchCtx, "serverName", name)

	// Create the workspace watcher
	workspaceWatcher := watcher.NewWorkspaceWatcher(lspClient)

	// Store the cancel function to be called during cleanup
	app.cancelFuncsMutex.Lock()
	app.watcherCancelFuncs = append(app.watcherCancelFuncs, cancelFunc)
	app.cancelFuncsMutex.Unlock()

	// Add the watcher to a WaitGroup to track active goroutines
	app.watcherWG.Add(1)

	// Add to map with mutex protection before starting goroutine
	app.clientsMutex.Lock()
	app.LSPClients[name] = lspClient
	app.clientsMutex.Unlock()

	go app.runWorkspaceWatcher(watchCtx, name, workspaceWatcher)
}

// runWorkspaceWatcher executes the workspace watcher for an LSP client
func (app *App) runWorkspaceWatcher(ctx context.Context, name string, workspaceWatcher *watcher.WorkspaceWatcher) {
	defer app.watcherWG.Done()
	defer logging.RecoverPanic("LSP-"+name, func() {
		// Try to restart the client
		app.restartLSPClient(ctx, name)
	})

	workspaceWatcher.WatchWorkspace(ctx, config.WorkingDirectory())
	logging.Info("Workspace watcher stopped", "client", name)
}

// restartLSPClient attempts to restart a crashed or failed LSP client
func (app *App) restartLSPClient(ctx context.Context, name string) {
	// Get the original configuration
	cfg := config.Get()
	clientConfig, exists := cfg.LSP[name]
	if !exists {
		logging.Error("Cannot restart client, configuration not found", "client", name)
		return
	}

	// Clean up the old client if it exists
	app.clientsMutex.Lock()
	oldClient, exists := app.LSPClients[name]
	if exists {
		delete(app.LSPClients, name) // Remove from map before potentially slow shutdown
	}
	app.clientsMutex.Unlock()

	if exists && oldClient != nil {
		// Try to shut it down gracefully, but don't block on errors
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = oldClient.Shutdown(shutdownCtx)
		cancel()
	}

	// Create a new client using the shared function
	app.createAndStartLSPClient(ctx, name, clientConfig.Command, clientConfig.Args...)
	logging.Info("Successfully restarted LSP client", "client", name)
}

// createAndStartCopilotClient creates and starts the GitHub Copilot LSP client
func (app *App) createAndStartCopilotClient(ctx context.Context) {
	logging.Info("Creating GitHub Copilot LSP client")

	cfg := config.Get()
	if cfg == nil {
		logging.Error("No config available for Copilot client")
		return
	}

	// Merge with defaults and environment
	copilotConfig := copilot.LoadFromEnvironment(copilot.MergeConfig(&cfg.Copilot))

	// Validate configuration
	if err := copilot.ValidateConfig(copilotConfig); err != nil {
		logging.Error("Invalid Copilot configuration", err)
		if !copilotConfig.FallbackToGopls {
			return
		}
		logging.Warn("Configuration invalid, but fallback is enabled")
	}

	// Create the Copilot client
	copilotClient, err := copilot.NewCopilotClient(ctx, copilotConfig)
	if err != nil {
		logging.Error("Failed to create Copilot client", err)
		if !copilotConfig.FallbackToGopls {
			return
		}
		logging.Warn("Copilot client creation failed, but fallback is enabled")
		return
	}

	// Create a longer timeout for initialization
	initCtx, cancel := context.WithTimeout(ctx, 45*time.Second) // Copilot may take longer
	defer cancel()

	// Initialize the Copilot client
	if err := copilotClient.Initialize(initCtx, config.WorkingDirectory()); err != nil {
		logging.Error("Failed to initialize Copilot client", err)
		if !copilotConfig.FallbackToGopls {
			return
		}
		logging.Warn("Copilot initialization failed, but fallback is enabled")
		// Continue to add to map for status reporting even if initialization failed
	}

	logging.Info("Copilot client initialized")

	// Create a child context that can be canceled when the app is shutting down
	watchCtx, cancelFunc := context.WithCancel(ctx)

	// Create a context with the server name for better identification
	watchCtx = context.WithValue(watchCtx, "serverName", "copilot")

	// Create the workspace watcher using the underlying LSP client
	workspaceWatcher := watcher.NewWorkspaceWatcher(copilotClient.Client)

	// Store the cancel function to be called during cleanup
	app.cancelFuncsMutex.Lock()
	app.watcherCancelFuncs = append(app.watcherCancelFuncs, cancelFunc)
	app.cancelFuncsMutex.Unlock()

	// Add the watcher to a WaitGroup to track active goroutines
	app.watcherWG.Add(1)

	// Add to map with mutex protection before starting goroutine
	app.clientsMutex.Lock()
	app.LSPClients["copilot"] = copilotClient.Client
	app.clientsMutex.Unlock()

	go app.runWorkspaceWatcher(watchCtx, "copilot", workspaceWatcher)

	logging.Info("GitHub Copilot LSP client setup completed")
}
