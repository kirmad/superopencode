package copilot

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kirmad/superopencode/internal/config"
)

// DefaultConfig returns a default Copilot configuration
func DefaultConfig() *config.CopilotConfig {
	return &config.CopilotConfig{
		EnableCopilot:     false,
		ChatEnabled:       true,
		CompletionEnabled: true,
		AutoInstall:       true,
		FallbackToGopls:   false,
		Timeout:           30,
		RetryAttempts:     3,
		LogLevel:          "info",
		Performance: &config.PerformanceConfig{
			MaxCompletionTime:   3000,
			DebounceDelay:       200,
			MaxParallelRequests: 5,
			CacheEnabled:        true,
			CacheSize:           100,
		},
		Security: &config.SecurityConfig{
			DisableTelemetry: false,
			PrivateMode:      false,
		},
		AgentConfig: &config.AgentConfig{
			CodingAgent:        true,
			DebuggingAgent:     true,
			DocumentationAgent: false,
		},
	}
}

// MergeConfig merges a user config with the default config
func MergeConfig(userConfig *config.CopilotConfig) *config.CopilotConfig {
	if userConfig == nil {
		return DefaultConfig()
	}
	
	result := DefaultConfig()
	
	// Merge basic settings
	if userConfig.EnableCopilot {
		result.EnableCopilot = userConfig.EnableCopilot
	}
	if userConfig.ServerPath != "" {
		result.ServerPath = userConfig.ServerPath
	}
	if userConfig.NodePath != "" {
		result.NodePath = userConfig.NodePath
	}
	if userConfig.UseNativeBinary {
		result.UseNativeBinary = userConfig.UseNativeBinary
	}
	if userConfig.ReplaceGopls {
		result.ReplaceGopls = userConfig.ReplaceGopls
	}
	
	// Merge authentication
	if userConfig.AuthToken != "" {
		result.AuthToken = userConfig.AuthToken
	}
	
	// Merge feature flags
	result.ChatEnabled = userConfig.ChatEnabled
	result.CompletionEnabled = userConfig.CompletionEnabled
	
	// Merge installation settings
	result.AutoInstall = userConfig.AutoInstall
	if len(userConfig.ServerArgs) > 0 {
		result.ServerArgs = userConfig.ServerArgs
	}
	if len(userConfig.Environment) > 0 {
		result.Environment = userConfig.Environment
	}
	
	// Merge performance settings
	if userConfig.Timeout > 0 {
		result.Timeout = userConfig.Timeout
	}
	if userConfig.RetryAttempts > 0 {
		result.RetryAttempts = userConfig.RetryAttempts
	}
	result.FallbackToGopls = userConfig.FallbackToGopls
	
	// Merge logging
	if userConfig.LogLevel != "" {
		result.LogLevel = userConfig.LogLevel
	}
	
	// Merge advanced settings
	if userConfig.Performance != nil {
		result.Performance = mergePerformanceConfig(result.Performance, userConfig.Performance)
	}
	if userConfig.Security != nil {
		result.Security = mergeSecurityConfig(result.Security, userConfig.Security)
	}
	if userConfig.AgentConfig != nil {
		result.AgentConfig = mergeAgentConfig(result.AgentConfig, userConfig.AgentConfig)
	}
	
	return result
}

// mergePerformanceConfig merges performance configurations
func mergePerformanceConfig(defaultCfg, userCfg *config.PerformanceConfig) *config.PerformanceConfig {
	result := *defaultCfg
	
	if userCfg.MaxCompletionTime > 0 {
		result.MaxCompletionTime = userCfg.MaxCompletionTime
	}
	if userCfg.DebounceDelay > 0 {
		result.DebounceDelay = userCfg.DebounceDelay
	}
	if userCfg.MaxParallelRequests > 0 {
		result.MaxParallelRequests = userCfg.MaxParallelRequests
	}
	result.CacheEnabled = userCfg.CacheEnabled
	if userCfg.CacheSize > 0 {
		result.CacheSize = userCfg.CacheSize
	}
	
	return &result
}

// mergeSecurityConfig merges security configurations
func mergeSecurityConfig(defaultCfg, userCfg *config.SecurityConfig) *config.SecurityConfig {
	result := *defaultCfg
	
	result.DisableTelemetry = userCfg.DisableTelemetry
	result.PrivateMode = userCfg.PrivateMode
	if len(userCfg.AllowedDomains) > 0 {
		result.AllowedDomains = userCfg.AllowedDomains
	}
	
	return &result
}

// mergeAgentConfig merges agent configurations
func mergeAgentConfig(defaultCfg, userCfg *config.AgentConfig) *config.AgentConfig {
	result := *defaultCfg
	
	result.CodingAgent = userCfg.CodingAgent
	result.DebuggingAgent = userCfg.DebuggingAgent
	result.DocumentationAgent = userCfg.DocumentationAgent
	
	return &result
}

// LoadFromEnvironment loads configuration from environment variables
func LoadFromEnvironment(cfg *config.CopilotConfig) *config.CopilotConfig {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	
	// Basic settings
	if env := os.Getenv("OPENCODE_COPILOT_ENABLE"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.EnableCopilot = enabled
		}
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_SERVER_PATH"); env != "" {
		cfg.ServerPath = env
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_NODE_PATH"); env != "" {
		cfg.NodePath = env
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_USE_NATIVE_BINARY"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.UseNativeBinary = enabled
		}
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_REPLACE_GOPLS"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.ReplaceGopls = enabled
		}
	}
	
	// Authentication
	if env := os.Getenv("OPENCODE_COPILOT_AUTH_TOKEN"); env != "" {
		cfg.AuthToken = env
	}
	
	// Feature flags
	if env := os.Getenv("OPENCODE_COPILOT_CHAT_ENABLED"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.ChatEnabled = enabled
		}
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_COMPLETION_ENABLED"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.CompletionEnabled = enabled
		}
	}
	
	// Installation
	if env := os.Getenv("OPENCODE_COPILOT_AUTO_INSTALL"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.AutoInstall = enabled
		}
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_SERVER_ARGS"); env != "" {
		cfg.ServerArgs = strings.Split(env, ",")
	}
	
	// Performance
	if env := os.Getenv("OPENCODE_COPILOT_TIMEOUT"); env != "" {
		if timeout, err := strconv.Atoi(env); err == nil {
			cfg.Timeout = timeout
		}
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_RETRY_ATTEMPTS"); env != "" {
		if attempts, err := strconv.Atoi(env); err == nil {
			cfg.RetryAttempts = attempts
		}
	}
	
	if env := os.Getenv("OPENCODE_COPILOT_FALLBACK_TO_GOPLS"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			cfg.FallbackToGopls = enabled
		}
	}
	
	// Logging
	if env := os.Getenv("OPENCODE_COPILOT_LOG_LEVEL"); env != "" {
		cfg.LogLevel = env
	}
	
	return cfg
}

// ValidateConfig validates a Copilot configuration
func ValidateConfig(cfg *config.CopilotConfig) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Validate server path if specified
	if cfg.ServerPath != "" {
		if !filepath.IsAbs(cfg.ServerPath) {
			return fmt.Errorf("server_path must be absolute: %s", cfg.ServerPath)
		}
		
		if _, err := os.Stat(cfg.ServerPath); err != nil {
			return fmt.Errorf("server_path does not exist: %s", cfg.ServerPath)
		}
	}
	
	// Validate node path if specified
	if cfg.NodePath != "" {
		if !filepath.IsAbs(cfg.NodePath) {
			return fmt.Errorf("node_path must be absolute: %s", cfg.NodePath)
		}
		
		if _, err := os.Stat(cfg.NodePath); err != nil {
			return fmt.Errorf("node_path does not exist: %s", cfg.NodePath)
		}
	}
	
	// Validate timeout
	if cfg.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative: %d", cfg.Timeout)
	}
	
	// Validate retry attempts
	if cfg.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative: %d", cfg.RetryAttempts)
	}
	
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if cfg.LogLevel != "" {
		valid := false
		for _, level := range validLogLevels {
			if cfg.LogLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log_level: %s (must be one of: %s)", 
				cfg.LogLevel, strings.Join(validLogLevels, ", "))
		}
	}
	
	// Validate performance config
	if cfg.Performance != nil {
		if err := validatePerformanceConfig(cfg.Performance); err != nil {
			return fmt.Errorf("performance config invalid: %w", err)
		}
	}
	
	return nil
}

// validatePerformanceConfig validates performance configuration
func validatePerformanceConfig(cfg *config.PerformanceConfig) error {
	if cfg.MaxCompletionTime < 0 {
		return fmt.Errorf("max_completion_time must be non-negative: %d", cfg.MaxCompletionTime)
	}
	
	if cfg.DebounceDelay < 0 {
		return fmt.Errorf("debounce_delay must be non-negative: %d", cfg.DebounceDelay)
	}
	
	if cfg.MaxParallelRequests < 1 {
		return fmt.Errorf("max_parallel_requests must be at least 1: %d", cfg.MaxParallelRequests)
	}
	
	if cfg.CacheSize < 0 {
		return fmt.Errorf("cache_size must be non-negative: %d", cfg.CacheSize)
	}
	
	return nil
}

// GetConfigForProfile returns a configuration for a specific profile
func GetConfigForProfile(profile string) *config.CopilotConfig {
	switch profile {
	case "development":
		cfg := DefaultConfig()
		cfg.LogLevel = "debug"
		cfg.AutoInstall = true
		return cfg
		
	case "production":
		cfg := DefaultConfig()
		cfg.LogLevel = "info"
		cfg.ReplaceGopls = true
		cfg.UseNativeBinary = true
		cfg.Security.DisableTelemetry = true
		if cfg.Performance != nil {
			cfg.Performance.CacheEnabled = true
			cfg.Performance.CacheSize = 200
		}
		return cfg
		
	case "testing":
		cfg := DefaultConfig()
		cfg.LogLevel = "warn"
		cfg.Performance.MaxCompletionTime = 1000
		return cfg
		
	default:
		return DefaultConfig()
	}
}