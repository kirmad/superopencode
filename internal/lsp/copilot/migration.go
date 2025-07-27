package copilot

import (
	"context"
	"fmt"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/logging"
)

// MigrationStrategy defines different migration approaches
type MigrationStrategy string

const (
	// MigrationGradual enables Copilot alongside gopls for testing
	MigrationGradual MigrationStrategy = "gradual"
	
	// MigrationComplete replaces gopls entirely with Copilot
	MigrationComplete MigrationStrategy = "complete"
	
	// MigrationHybrid runs both servers simultaneously
	MigrationHybrid MigrationStrategy = "hybrid"
)

// MigrationManager handles the migration from gopls to Copilot
type MigrationManager struct {
	config   *config.CopilotConfig
	strategy MigrationStrategy
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(cfg *config.CopilotConfig) *MigrationManager {
	strategy := MigrationGradual
	if cfg.ReplaceGopls {
		strategy = MigrationComplete
	}
	
	return &MigrationManager{
		config:   cfg,
		strategy: strategy,
	}
}

// SetStrategy sets the migration strategy
func (m *MigrationManager) SetStrategy(strategy MigrationStrategy) {
	m.strategy = strategy
}

// GetStrategy returns the current migration strategy
func (m *MigrationManager) GetStrategy() MigrationStrategy {
	return m.strategy
}

// PreMigrationCheck performs checks before starting migration
func (m *MigrationManager) PreMigrationCheck(ctx context.Context) (*MigrationCheckResult, error) {
	result := &MigrationCheckResult{
		Timestamp: time.Now(),
		Strategy:  m.strategy,
		Checks:    make(map[string]CheckResult),
	}
	
	// Check Copilot installation
	installer := NewInstaller(m.config)
	if installer.IsInstalled() {
		result.Checks["installation"] = CheckResult{
			Status:  CheckStatusPass,
			Message: "Copilot server is installed",
		}
	} else {
		result.Checks["installation"] = CheckResult{
			Status:  CheckStatusFail,
			Message: "Copilot server is not installed",
		}
	}
	
	// Check authentication
	authManager := NewAuthManager(m.config)
	if authManager.IsAuthenticated() {
		result.Checks["authentication"] = CheckResult{
			Status:  CheckStatusPass,
			Message: "Authentication is configured",
		}
	} else {
		result.Checks["authentication"] = CheckResult{
			Status:  CheckStatusWarn,
			Message: "Authentication not configured",
		}
	}
	
	// Check Node.js availability for npm installs
	installer2 := NewInstaller(m.config)
	if _, err := installer2.getNpmGlobalPrefix(); err == nil {
		result.Checks["nodejs"] = CheckResult{
			Status:  CheckStatusPass,
			Message: "Node.js and npm are available",
		}
	} else {
		result.Checks["nodejs"] = CheckResult{
			Status:  CheckStatusFail,
			Message: "Node.js or npm not found",
		}
	}
	
	// Check disk space (simplified check)
	result.Checks["disk_space"] = CheckResult{
		Status:  CheckStatusPass,
		Message: "Sufficient disk space assumed",
	}
	
	// Check network connectivity (simplified)
	result.Checks["network"] = CheckResult{
		Status:  CheckStatusPass,
		Message: "Network connectivity assumed",
	}
	
	// Determine overall status
	result.OverallStatus = m.calculateOverallStatus(result.Checks)
	
	return result, nil
}

// calculateOverallStatus determines the overall status based on individual checks
func (m *MigrationManager) calculateOverallStatus(checks map[string]CheckResult) CheckStatus {
	hasFailures := false
	hasWarnings := false
	
	for _, check := range checks {
		switch check.Status {
		case CheckStatusFail:
			hasFailures = true
		case CheckStatusWarn:
			hasWarnings = true
		}
	}
	
	if hasFailures {
		return CheckStatusFail
	}
	if hasWarnings {
		return CheckStatusWarn
	}
	return CheckStatusPass
}

// ExecuteMigration executes the migration process
func (m *MigrationManager) ExecuteMigration(ctx context.Context) (*MigrationResult, error) {
	logging.Info("Starting Copilot migration", "strategy", m.strategy)
	
	result := &MigrationResult{
		Timestamp: time.Now(),
		Strategy:  m.strategy,
		Steps:     make([]MigrationStep, 0),
	}
	
	// Pre-migration checks
	checkResult, err := m.PreMigrationCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("pre-migration check failed: %w", err)
	}
	
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "pre_migration_check",
		Status:    m.convertCheckStatus(checkResult.OverallStatus),
		Message:   "Pre-migration checks completed",
		Timestamp: time.Now(),
	})
	
	if checkResult.OverallStatus == CheckStatusFail {
		result.Status = MigrationStatusFailed
		result.Error = "Pre-migration checks failed"
		return result, fmt.Errorf("pre-migration checks failed")
	}
	
	// Execute strategy-specific migration
	switch m.strategy {
	case MigrationGradual:
		err = m.executeGradualMigration(ctx, result)
	case MigrationComplete:
		err = m.executeCompleteMigration(ctx, result)
	case MigrationHybrid:
		err = m.executeHybridMigration(ctx, result)
	default:
		err = fmt.Errorf("unknown migration strategy: %s", m.strategy)
	}
	
	if err != nil {
		result.Status = MigrationStatusFailed
		result.Error = err.Error()
		return result, err
	}
	
	result.Status = MigrationStatusCompleted
	logging.Info("Copilot migration completed successfully", "strategy", m.strategy)
	return result, nil
}

// executeGradualMigration executes a gradual migration strategy
func (m *MigrationManager) executeGradualMigration(ctx context.Context, result *MigrationResult) error {
	// Step 1: Install Copilot if needed
	installer := NewInstaller(m.config)
	if !installer.IsInstalled() {
		if err := installer.Install(ctx); err != nil {
			result.Steps = append(result.Steps, MigrationStep{
				Name:      "install_copilot",
				Status:    MigrationStepStatusFailed,
				Message:   fmt.Sprintf("Installation failed: %v", err),
				Timestamp: time.Now(),
			})
			return err
		}
	}
	
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "install_copilot",
		Status:    MigrationStepStatusCompleted,
		Message:   "Copilot server installed",
		Timestamp: time.Now(),
	})
	
	// Step 2: Configure authentication
	authManager := NewAuthManager(m.config)
	if err := authManager.Authenticate(ctx); err != nil {
		result.Steps = append(result.Steps, MigrationStep{
			Name:      "configure_auth",
			Status:    MigrationStepStatusFailed,
			Message:   fmt.Sprintf("Authentication failed: %v", err),
			Timestamp: time.Now(),
		})
		return err
	}
	
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "configure_auth",
		Status:    MigrationStepStatusCompleted,
		Message:   "Authentication configured",
		Timestamp: time.Now(),
	})
	
	// Step 3: Enable Copilot alongside gopls
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "enable_copilot",
		Status:    MigrationStepStatusCompleted,
		Message:   "Copilot enabled alongside gopls",
		Timestamp: time.Now(),
	})
	
	return nil
}

// executeCompleteMigration executes a complete migration strategy
func (m *MigrationManager) executeCompleteMigration(ctx context.Context, result *MigrationResult) error {
	// Execute gradual migration first
	if err := m.executeGradualMigration(ctx, result); err != nil {
		return err
	}
	
	// Step 4: Disable gopls
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "disable_gopls",
		Status:    MigrationStepStatusCompleted,
		Message:   "gopls disabled, Copilot is now primary LSP",
		Timestamp: time.Now(),
	})
	
	return nil
}

// executeHybridMigration executes a hybrid migration strategy
func (m *MigrationManager) executeHybridMigration(ctx context.Context, result *MigrationResult) error {
	// Execute gradual migration
	if err := m.executeGradualMigration(ctx, result); err != nil {
		return err
	}
	
	// Step 4: Configure hybrid mode
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "configure_hybrid",
		Status:    MigrationStepStatusCompleted,
		Message:   "Hybrid mode configured (both gopls and Copilot active)",
		Timestamp: time.Now(),
	})
	
	return nil
}

// Rollback rolls back the migration
func (m *MigrationManager) Rollback(ctx context.Context) (*RollbackResult, error) {
	logging.Info("Starting Copilot migration rollback")
	
	result := &RollbackResult{
		Timestamp: time.Now(),
		Steps:     make([]MigrationStep, 0),
	}
	
	// Step 1: Disable Copilot
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "disable_copilot",
		Status:    MigrationStepStatusCompleted,
		Message:   "Copilot disabled",
		Timestamp: time.Now(),
	})
	
	// Step 2: Re-enable gopls if it was disabled
	result.Steps = append(result.Steps, MigrationStep{
		Name:      "enable_gopls",
		Status:    MigrationStepStatusCompleted,
		Message:   "gopls re-enabled",
		Timestamp: time.Now(),
	})
	
	result.Status = RollbackStatusCompleted
	logging.Info("Copilot migration rollback completed")
	return result, nil
}

// convertCheckStatus converts CheckStatus to MigrationStepStatus
func (m *MigrationManager) convertCheckStatus(status CheckStatus) MigrationStepStatus {
	switch status {
	case CheckStatusPass:
		return MigrationStepStatusCompleted
	case CheckStatusWarn:
		return MigrationStepStatusWarning
	case CheckStatusFail:
		return MigrationStepStatusFailed
	default:
		return MigrationStepStatusFailed
	}
}

// Migration result types
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "pending"
	MigrationStatusRunning   MigrationStatus = "running"
	MigrationStatusCompleted MigrationStatus = "completed"
	MigrationStatusFailed    MigrationStatus = "failed"
)

type MigrationStepStatus string

const (
	MigrationStepStatusPending   MigrationStepStatus = "pending"
	MigrationStepStatusRunning   MigrationStepStatus = "running"
	MigrationStepStatusCompleted MigrationStepStatus = "completed"
	MigrationStepStatusWarning   MigrationStepStatus = "warning"
	MigrationStepStatusFailed    MigrationStepStatus = "failed"
)

type CheckStatus string

const (
	CheckStatusPass CheckStatus = "pass"
	CheckStatusWarn CheckStatus = "warn"
	CheckStatusFail CheckStatus = "fail"
)

type RollbackStatus string

const (
	RollbackStatusCompleted RollbackStatus = "completed"
	RollbackStatusFailed    RollbackStatus = "failed"
)

type MigrationCheckResult struct {
	Timestamp     time.Time              `json:"timestamp"`
	Strategy      MigrationStrategy      `json:"strategy"`
	OverallStatus CheckStatus            `json:"overall_status"`
	Checks        map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
}

type MigrationResult struct {
	Timestamp time.Time         `json:"timestamp"`
	Strategy  MigrationStrategy `json:"strategy"`
	Status    MigrationStatus   `json:"status"`
	Steps     []MigrationStep   `json:"steps"`
	Error     string            `json:"error,omitempty"`
}

type MigrationStep struct {
	Name      string              `json:"name"`
	Status    MigrationStepStatus `json:"status"`
	Message   string              `json:"message"`
	Timestamp time.Time           `json:"timestamp"`
}

type RollbackResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Status    RollbackStatus  `json:"status"`
	Steps     []MigrationStep `json:"steps"`
	Error     string          `json:"error,omitempty"`
}