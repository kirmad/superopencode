package copilot

import (
	"context"
	"testing"
	"time"

	"github.com/kirmad/superopencode/internal/config"
)

func TestNewMigrationManager(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}

	migrator := NewMigrationManager(cfg)
	if migrator == nil {
		t.Error("NewMigrationManager() returned nil")
	}

	if migrator.config != cfg {
		t.Error("NewMigrationManager() did not store config correctly")
	}
}

func TestMigrationManager_GetStrategy(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ReplaceGopls:  false, // Should default to gradual
	}
	migrator := NewMigrationManager(cfg)

	strategy := migrator.GetStrategy()
	if strategy != MigrationGradual {
		t.Errorf("GetStrategy() = %v, want %v", strategy, MigrationGradual)
	}

	// Test with ReplaceGopls true (should default to complete)
	cfg.ReplaceGopls = true
	migrator2 := NewMigrationManager(cfg)
	strategy2 := migrator2.GetStrategy()
	if strategy2 != MigrationComplete {
		t.Errorf("GetStrategy() with ReplaceGopls = %v, want %v", strategy2, MigrationComplete)
	}
}

func TestMigrationManager_SetStrategy(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}
	migrator := NewMigrationManager(cfg)

	// Test setting different strategies
	testStrategies := []MigrationStrategy{
		MigrationGradual,
		MigrationComplete,
		MigrationHybrid,
	}
	
	for _, strategy := range testStrategies {
		migrator.SetStrategy(strategy)
		currentStrategy := migrator.GetStrategy()
		
		if currentStrategy != strategy {
			t.Errorf("SetStrategy(%v) -> GetStrategy() = %v, want %v", strategy, currentStrategy, strategy)
		}
	}
}

func TestMigrationManager_PreMigrationCheck(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}
	migrator := NewMigrationManager(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test pre-migration checks
	result, err := migrator.PreMigrationCheck(ctx)
	if err != nil {
		t.Errorf("PreMigrationCheck() returned error: %v", err)
		return
	}
	
	// Pre-checks should return a valid result
	if result == nil {
		t.Error("PreMigrationCheck() returned nil result")
		return
	}

	if result.Strategy != migrator.GetStrategy() {
		t.Errorf("PreMigrationCheck() strategy = %v, want %v", result.Strategy, migrator.GetStrategy())
	}

	// Should have some checks
	if len(result.Checks) == 0 {
		t.Error("PreMigrationCheck() returned no checks")
	}

	// Should have an overall status
	if result.OverallStatus == "" {
		t.Error("PreMigrationCheck() returned empty overall status")
	}
}

func TestMigrationManager_ExecuteMigration(t *testing.T) {
	tests := []struct {
		name     string
		strategy MigrationStrategy
		wantErr  bool
	}{
		{
			name:     "gradual strategy",
			strategy: MigrationGradual,
			wantErr:  false,
		},
		{
			name:     "complete strategy",
			strategy: MigrationComplete,
			wantErr:  false,
		},
		{
			name:     "hybrid strategy",
			strategy: MigrationHybrid,
			wantErr:  false,
		},
	}

	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := NewMigrationManager(cfg)
			migrator.SetStrategy(tt.strategy)
			
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := migrator.ExecuteMigration(ctx)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteMigration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Errorf("ExecuteMigration() returned nil result without error")
				} else {
					// Validate result structure
					if result.Strategy != tt.strategy {
						t.Errorf("ExecuteMigration() result strategy = %v, want %v", result.Strategy, tt.strategy)
					}
					
					// Should have some status
					if result.Status == "" {
						t.Error("ExecuteMigration() returned empty status")
					}
				}
			}
		})
	}
}

func TestMigrationManager_Rollback(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}
	migrator := NewMigrationManager(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test rollback (should work even without prior migration)
	result, err := migrator.Rollback(ctx)
	if err != nil {
		t.Errorf("Rollback() should not return error: %v", err)
		return
	}

	if result == nil {
		t.Error("Rollback() returned nil result")
		return
	}

	// Should have completed status
	if result.Status != RollbackStatusCompleted {
		t.Errorf("Rollback() status = %v, want %v", result.Status, RollbackStatusCompleted)
	}

	// Should have some steps
	if len(result.Steps) == 0 {
		t.Error("Rollback() returned no steps")
	}
}

func TestMigrationStrategy_Constants(t *testing.T) {
	// Test that the migration strategy constants are defined correctly
	if MigrationGradual != "gradual" {
		t.Errorf("MigrationGradual = %v, want %v", MigrationGradual, "gradual")
	}
	
	if MigrationComplete != "complete" {
		t.Errorf("MigrationComplete = %v, want %v", MigrationComplete, "complete")
	}
	
	if MigrationHybrid != "hybrid" {
		t.Errorf("MigrationHybrid = %v, want %v", MigrationHybrid, "hybrid")
	}
}

func TestMigrationStatus_Constants(t *testing.T) {
	// Test migration status constants
	if MigrationStatusPending != "pending" {
		t.Errorf("MigrationStatusPending = %v, want %v", MigrationStatusPending, "pending")
	}
	
	if MigrationStatusCompleted != "completed" {
		t.Errorf("MigrationStatusCompleted = %v, want %v", MigrationStatusCompleted, "completed")
	}
	
	if MigrationStatusFailed != "failed" {
		t.Errorf("MigrationStatusFailed = %v, want %v", MigrationStatusFailed, "failed")
	}
}