# `/model` Command Implementation Guide

## Overview

This document provides detailed implementation specifications for the `/model` slash command, enabling users to list available models and switch between them directly from the chat interface.

## Features

### Core Functionality
- **Model Listing**: Display all available models grouped by provider
- **Model Switching**: Switch active model for current session
- **Provider Filtering**: Show models for specific providers
- **Current Model Status**: Display active model information
- **Search Capability**: Find models by name or capability

### Command Syntax
```bash
/model                          # Show all available models
/model <model-name>            # Switch to specified model
/model --provider <provider>   # Show models for specific provider
/model --current              # Show current active model
/model --search <term>        # Search models by name
/model --help                 # Show command help
```

## Architecture

### Component Overview
```
Chat Interface → Slash Command Processor → Model Command Handler → Model Manager → Configuration
```

### Integration Points
1. **Slash Command System**: Extends existing command processor
2. **Model Management**: Leverages current model discovery and validation
3. **Configuration**: Enhances existing agent configuration system
4. **UI Display**: Reuses model dialog display components

## Implementation Plan

### Phase 1: Command Infrastructure

#### 1.1 Extend Slash Command Processor
**File**: `internal/tui/components/dialog/slash_commands.go`

**Location**: Line 87 (in `IsSlashCommand` method)
```go
// Add model command detection
if strings.HasPrefix(trimmed, "model") {
    return true
}
```

**Location**: Line 120 (in `ProcessSlashCommand` method)
```go
// Add model command routing
case "model":
    return scp.handleModelCommand(args[1:])
```

#### 1.2 Model Command Recognition
**New Method in SlashCommandProcessor**:
```go
func (scp *SlashCommandProcessor) handleModelCommand(args []string) *SlashCommandResult {
    modelHandler := NewModelCommandHandler(scp.configManager, scp.modelService)
    return modelHandler.Execute(args)
}
```

### Phase 2: Model Command Handler

#### 2.1 Create Model Command Handler
**New File**: `internal/tui/components/dialog/model_commands.go`

```go
package dialog

import (
    "fmt"
    "strings"
    "sort"
    
    "github.com/opencodedev/opencode/internal/config"
    "github.com/opencodedev/opencode/internal/llm/models"
    "github.com/opencodedev/opencode/internal/llm/agent"
)

type ModelCommandHandler struct {
    configManager *config.Manager
    modelService  *models.Service
    agentManager  *agent.Manager
}

func NewModelCommandHandler(cm *config.Manager, ms *models.Service) *ModelCommandHandler {
    return &ModelCommandHandler{
        configManager: cm,
        modelService:  ms,
        agentManager:  agent.NewManager(cm),
    }
}

func (mch *ModelCommandHandler) Execute(args []string) *SlashCommandResult {
    if len(args) == 0 {
        return mch.showAvailableModels()
    }
    
    switch args[0] {
    case "--current":
        return mch.showCurrentModel()
    case "--provider":
        if len(args) < 2 {
            return &SlashCommandResult{
                Error: fmt.Errorf("--provider requires provider name"),
            }
        }
        return mch.showProviderModels(args[1])
    case "--search":
        if len(args) < 2 {
            return &SlashCommandResult{
                Error: fmt.Errorf("--search requires search term"),
            }
        }
        return mch.searchModels(args[1])
    case "--help":
        return mch.showHelp()
    default:
        return mch.switchToModel(args[0])
    }
}
```

#### 2.2 Core Handler Methods

**Show Available Models**:
```go
func (mch *ModelCommandHandler) showAvailableModels() *SlashCommandResult {
    providers := mch.modelService.GetEnabledProviders()
    
    var content strings.Builder
    content.WriteString("## Available Models\n\n")
    
    // Show current model first
    currentModel := mch.getCurrentModelInfo()
    content.WriteString(fmt.Sprintf("**Current**: %s (%s)\n\n", 
        currentModel.Name, currentModel.Provider))
    
    // Group by provider
    for _, provider := range providers {
        models := mch.modelService.GetModelsForProvider(provider)
        if len(models) == 0 {
            continue
        }
        
        content.WriteString(fmt.Sprintf("### %s\n", provider))
        for _, model := range models {
            status := ""
            if model.ID == currentModel.ID {
                status = " ← **active**"
            }
            
            content.WriteString(fmt.Sprintf("- `%s` - %s%s\n", 
                model.ID, model.Name, status))
            content.WriteString(fmt.Sprintf("  - Context: %s tokens | Cost: $%.2f/1M in\n", 
                formatNumber(model.ContextWindow), model.CostPer1MIn))
        }
        content.WriteString("\n")
    }
    
    content.WriteString("**Usage**: `/model <model-id>` to switch models\n")
    
    return &SlashCommandResult{
        ProcessedCommand: &ProcessedCommand{
            Command: &Command{Content: content.String()},
        },
    }
}
```

**Switch Model**:
```go
func (mch *ModelCommandHandler) switchToModel(modelID string) *SlashCommandResult {
    // Validate model exists
    model, err := mch.modelService.GetModel(models.ModelID(modelID))
    if err != nil {
        return mch.suggestSimilarModels(modelID)
    }
    
    // Validate provider is enabled
    if !mch.modelService.IsProviderEnabled(model.Provider) {
        return &SlashCommandResult{
            Error: fmt.Errorf("provider %s is not enabled. Configure API key in settings.", 
                model.Provider),
        }
    }
    
    // Update agent configuration
    err = mch.agentManager.UpdateModel(config.AgentCoder, model.ID)
    if err != nil {
        return &SlashCommandResult{
            Error: fmt.Errorf("failed to switch model: %w", err),
        }
    }
    
    // Save configuration
    err = mch.configManager.Save()
    if err != nil {
        return &SlashCommandResult{
            Error: fmt.Errorf("failed to save configuration: %w", err),
        }
    }
    
    content := fmt.Sprintf("✅ **Switched to %s**\n\n", model.Name)
    content += fmt.Sprintf("- **Provider**: %s\n", model.Provider)
    content += fmt.Sprintf("- **Context Window**: %s tokens\n", formatNumber(model.ContextWindow))
    content += fmt.Sprintf("- **Cost**: $%.2f per 1M input tokens\n", model.CostPer1MIn)
    
    if model.CanReason {
        content += "- **Reasoning**: Supported\n"
    }
    if model.SupportsAttachments {
        content += "- **Attachments**: Supported\n"
    }
    
    return &SlashCommandResult{
        ProcessedCommand: &ProcessedCommand{
            Command: &Command{Content: content},
        },
    }
}
```

**Show Current Model**:
```go
func (mch *ModelCommandHandler) showCurrentModel() *SlashCommandResult {
    model := mch.getCurrentModelInfo()
    
    content := fmt.Sprintf("## Current Model: %s\n\n", model.Name)
    content += fmt.Sprintf("- **ID**: `%s`\n", model.ID)
    content += fmt.Sprintf("- **Provider**: %s\n", model.Provider)
    content += fmt.Sprintf("- **Context Window**: %s tokens\n", formatNumber(model.ContextWindow))
    content += fmt.Sprintf("- **Max Output**: %s tokens\n", formatNumber(model.DefaultMaxTokens))
    content += fmt.Sprintf("- **Cost**: $%.2f/$%.2f per 1M tokens (in/out)\n", 
        model.CostPer1MIn, model.CostPer1MOut)
    
    if model.CanReason {
        content += "- **Reasoning**: ✅ Supported\n"
    }
    if model.SupportsAttachments {
        content += "- **Attachments**: ✅ Supported\n"
    }
    
    return &SlashCommandResult{
        ProcessedCommand: &ProcessedCommand{
            Command: &Command{Content: content},
        },
    }
}
```

### Phase 3: Model Manager Enhancement

#### 3.1 Enhanced Model Management
**New File**: `internal/llm/models/manager.go`

```go
package models

import (
    "fmt"
    "strings"
    "sort"
    
    "github.com/opencodedev/opencode/internal/config"
)

type Manager struct {
    service *Service
    config  *config.Manager
}

func NewManager(service *Service, config *config.Manager) *Manager {
    return &Manager{
        service: service,
        config:  config,
    }
}

func (m *Manager) SwitchModel(modelID ModelID, agentType config.AgentType) error {
    // Validate model exists and is available
    model, err := m.service.GetModel(modelID)
    if err != nil {
        return fmt.Errorf("model not found: %w", err)
    }
    
    if !m.service.IsProviderEnabled(model.Provider) {
        return fmt.Errorf("provider %s is not enabled", model.Provider)
    }
    
    // Update agent configuration
    agent := m.config.GetAgent(agentType)
    agent.Model = modelID
    agent.LastSwitched = time.Now()
    agent.SwitchCount++
    
    m.config.SetAgent(agentType, agent)
    
    // Update model preferences
    prefs := m.config.GetModelPreferences()
    prefs.AddRecent(modelID)
    m.config.SetModelPreferences(prefs)
    
    return m.config.Save()
}

func (m *Manager) SearchModels(query string) []Model {
    allModels := m.service.GetAllModels()
    query = strings.ToLower(query)
    
    var matches []Model
    for _, model := range allModels {
        if strings.Contains(strings.ToLower(string(model.ID)), query) ||
           strings.Contains(strings.ToLower(model.Name), query) ||
           strings.Contains(strings.ToLower(string(model.Provider)), query) {
            matches = append(matches, model)
        }
    }
    
    // Sort by relevance (exact matches first, then partial)
    sort.Slice(matches, func(i, j int) bool {
        scoreI := m.getRelevanceScore(matches[i], query)
        scoreJ := m.getRelevanceScore(matches[j], query)
        return scoreI > scoreJ
    })
    
    return matches
}

func (m *Manager) SuggestSimilarModels(input string) []Model {
    // Implement fuzzy matching logic
    allModels := m.service.GetAllModels()
    var suggestions []Model
    
    for _, model := range allModels {
        if m.calculateSimilarity(strings.ToLower(input), strings.ToLower(string(model.ID))) > 0.6 {
            suggestions = append(suggestions, model)
        }
    }
    
    return suggestions
}
```

### Phase 4: Configuration Enhancement

#### 4.1 Enhanced Agent Configuration
**File**: `internal/config/config.go`

**Add to Agent struct** (around line 67):
```go
type Agent struct {
    Model           models.ModelID `json:"model"`
    MaxTokens       int64          `json:"maxTokens"`
    ReasoningEffort string         `json:"reasoningEffort"`
    LastSwitched    time.Time      `json:"lastSwitched,omitempty"`
    SwitchCount     int            `json:"switchCount,omitempty"`
}
```

**Add ModelPreferences struct**:
```go
type ModelPreferences struct {
    Recent       []models.ModelID            `json:"recent,omitempty"`
    Favorites    []models.ModelID            `json:"favorites,omitempty"`
    DefaultAgent map[AgentType]models.ModelID `json:"defaultAgent,omitempty"`
    MaxRecent    int                         `json:"maxRecent,omitempty"`
}

func (mp *ModelPreferences) AddRecent(modelID models.ModelID) {
    if mp.MaxRecent == 0 {
        mp.MaxRecent = 10
    }
    
    // Remove if already exists
    for i, id := range mp.Recent {
        if id == modelID {
            mp.Recent = append(mp.Recent[:i], mp.Recent[i+1:]...)
            break
        }
    }
    
    // Add to front
    mp.Recent = append([]models.ModelID{modelID}, mp.Recent...)
    
    // Trim to max
    if len(mp.Recent) > mp.MaxRecent {
        mp.Recent = mp.Recent[:mp.MaxRecent]
    }
}
```

**Add to Config struct**:
```go
type Config struct {
    // ... existing fields ...
    ModelPreferences ModelPreferences `json:"modelPreferences,omitempty"`
}
```

#### 4.2 Configuration Methods
```go
func (c *Config) GetModelPreferences() ModelPreferences {
    return c.ModelPreferences
}

func (c *Config) SetModelPreferences(prefs ModelPreferences) {
    c.ModelPreferences = prefs
}
```

### Phase 5: Command Registration

#### 5.1 Register Built-in Command
**File**: `internal/tui/tui.go`

**Add to initialization** (around line 89):
```go
func (model *Model) initializeCommands() {
    // Register built-in commands
    model.RegisterCommand(&dialog.Command{
        Name:    "model",
        Content: "", // Built-in command, no content needed
        IsBuiltIn: true,
    })
}
```

#### 5.2 Chat Integration
**File**: `internal/tui/page/chat.go`

**Modify handleSlashCommand method** (around line 234):
```go
func (p *ChatPage) handleSlashCommand(text string, attachments []attachment.Attachment) tea.Cmd {
    result := p.slashProcessor.ProcessSlashCommand(text)
    if result.Error != nil {
        return p.showError(result.Error)
    }
    
    if result.ProcessedCommand != nil {
        // Handle built-in commands that don't send messages
        if result.ProcessedCommand.Command.IsBuiltIn {
            return p.showCommandResult(result.ProcessedCommand.Command.Content)
        }
        
        // Regular command processing
        return p.sendMessage(result.ProcessedCommand.Content, attachments)
    }
    
    return nil
}
```

## Testing Strategy

### Unit Tests

#### Model Command Handler Tests
```go
func TestModelCommandHandler_ShowAvailableModels(t *testing.T) {
    // Test model listing functionality
}

func TestModelCommandHandler_SwitchToModel(t *testing.T) {
    // Test successful model switching
    // Test invalid model handling
    // Test disabled provider handling
}

func TestModelCommandHandler_SearchModels(t *testing.T) {
    // Test search functionality
    // Test fuzzy matching
}
```

#### Model Manager Tests
```go
func TestManager_SwitchModel(t *testing.T) {
    // Test model switching logic
    // Test configuration persistence
    // Test error conditions
}
```

### Integration Tests

#### End-to-End Command Flow
```go
func TestSlashCommandIntegration(t *testing.T) {
    // Test complete flow from command input to model switch
    // Test error handling and user feedback
}
```

## Error Handling

### Common Error Scenarios

1. **Model Not Found**
   - Show fuzzy match suggestions
   - List available models
   - Provide clear error message

2. **Provider Disabled**
   - Explain how to enable provider
   - Show provider configuration guide
   - Suggest alternative models

3. **Configuration Save Failure**
   - Preserve current state
   - Show clear error message
   - Provide recovery instructions

4. **Invalid Arguments**
   - Show command help
   - Provide usage examples
   - Suggest corrections

### Error Message Examples

```bash
# Model not found
❌ Model 'gpt-5' not found. Did you mean:
- gpt-4o
- gpt-4.1
- gpt-4o-mini

# Provider disabled
❌ Provider 'openai' is not enabled. 
Configure your OpenAI API key in settings to use GPT models.

# Invalid syntax
❌ Invalid command syntax.
Usage: /model [model-name] [--provider provider] [--current] [--search term]
```

## Documentation Updates

### User Documentation
- Add `/model` command to slash commands documentation
- Update model switching workflows
- Add troubleshooting guide

### Developer Documentation
- Document model command handler architecture
- Add configuration schema changes
- Update API documentation

## Deployment Considerations

### Backward Compatibility
- Existing `Ctrl+O` model dialog continues to work
- Configuration migration for new fields
- Graceful handling of missing configuration

### Performance
- Reuse existing model discovery logic
- Cache model information for quick access
- Minimal impact on startup time

### Security
- Validate all user inputs
- Sanitize model names and provider names
- Prevent configuration corruption

## Success Metrics

### Functionality
- ✅ All command variations work correctly
- ✅ Model switching persists across sessions
- ✅ Error handling provides helpful feedback
- ✅ Performance impact is minimal

### User Experience
- ✅ Commands are intuitive and discoverable
- ✅ Feedback is immediate and clear
- ✅ Integration feels seamless

### Technical
- ✅ Code follows existing patterns
- ✅ Test coverage is comprehensive
- ✅ Documentation is complete

## Future Enhancements

### Advanced Features
- Model favorites and bookmarks
- Usage statistics and recommendations
- Model performance metrics
- Automatic model suggestions based on task type

### Integration Opportunities
- Integration with model comparison tools
- Cost tracking and optimization
- Performance monitoring
- A/B testing capabilities

---

*This implementation guide provides a comprehensive roadmap for adding the `/model` command to SuperOpenCode, maintaining consistency with existing architecture while adding powerful new functionality.*