package agent

import (
	"context"

	"github.com/kirmad/superopencode/internal/history"
	"github.com/kirmad/superopencode/internal/llm/tools"
	"github.com/kirmad/superopencode/internal/lsp"
	"github.com/kirmad/superopencode/internal/message"
	"github.com/kirmad/superopencode/internal/permission"
	"github.com/kirmad/superopencode/internal/session"
)

func CoderAgentTools(
	permissions permission.Service,
	sessions session.Service,
	messages message.Service,
	history history.Service,
	lspClients map[string]*lsp.Client,
) []tools.BaseTool {
	ctx := context.Background()
	otherTools := GetMcpTools(ctx, permissions)
	if len(lspClients) > 0 {
		otherTools = append(otherTools, tools.NewDiagnosticsTool(lspClients))
	}
	return append(
		[]tools.BaseTool{
			tools.NewBashTool(permissions),
			tools.NewEditTool(lspClients, permissions, history),
			tools.NewFetchTool(permissions),
			tools.NewGlobTool(),
			tools.NewGrepTool(),
			tools.NewLsTool(),
			tools.NewSourcegraphTool(),
			tools.NewTodoReadTool(),
			tools.NewTodoWriteTool(),
			tools.NewViewTool(lspClients),
			tools.NewPatchTool(lspClients, permissions, history),
			tools.NewWriteTool(lspClients, permissions, history),
			NewAgentTool(sessions, messages, lspClients),
			NewTaskTool(sessions, messages, permissions, history, lspClients),
		}, otherTools...,
	)
}

// TaskAgentTools provides limited read-only tools for task agents
func TaskAgentTools(lspClients map[string]*lsp.Client) []tools.BaseTool {
	return []tools.BaseTool{
		tools.NewGlobTool(),
		tools.NewGrepTool(),
		tools.NewLsTool(),
		tools.NewSourcegraphTool(),
		tools.NewViewTool(lspClients),
	}
}

// ResearchAgentTools provides research-optimized tools
func ResearchAgentTools(
	permissions permission.Service,
	sessions session.Service,
	messages message.Service,
	history history.Service,
	lspClients map[string]*lsp.Client,
) []tools.BaseTool {
	ctx := context.Background()
	mcpTools := GetMcpTools(ctx, permissions)
	
	return append([]tools.BaseTool{
		tools.NewViewTool(lspClients),          // Read files
		tools.NewGrepTool(),                    // Search content
		tools.NewGlobTool(),                    // Find files
		tools.NewSourcegraphTool(),             // Advanced search
		tools.NewFetchTool(permissions),        // Web research
		tools.NewLsTool(),                      // Directory exploration
		tools.NewTodoReadTool(),                // Task tracking
		tools.NewTodoWriteTool(),               // Task management
	}, mcpTools...) // Include MCP tools for enhanced research capabilities
}

// CodingAgentTools provides coding-optimized tools
func CodingAgentTools(
	permissions permission.Service,
	sessions session.Service,
	messages message.Service,
	history history.Service,
	lspClients map[string]*lsp.Client,
) []tools.BaseTool {
	ctx := context.Background()
	mcpTools := GetMcpTools(ctx, permissions)
	
	var diagnosticTools []tools.BaseTool
	if len(lspClients) > 0 {
		diagnosticTools = append(diagnosticTools, tools.NewDiagnosticsTool(lspClients))
	}
	
	return append(append([]tools.BaseTool{
		tools.NewViewTool(lspClients),          // Read code
		tools.NewWriteTool(lspClients, permissions, history), // Create files
		tools.NewEditTool(lspClients, permissions, history),  // Edit code
		tools.NewBashTool(permissions),         // Execute commands
		tools.NewGrepTool(),                    // Search code
		tools.NewGlobTool(),                    // Find files
		tools.NewPatchTool(lspClients, permissions, history), // Apply patches
		tools.NewLsTool(),                      // Directory navigation
		tools.NewTodoReadTool(),                // Task tracking
		tools.NewTodoWriteTool(),               // Task management
	}, diagnosticTools...), mcpTools...) // Include MCP tools and diagnostics
}

// AnalysisAgentTools provides analysis-optimized tools
func AnalysisAgentTools(
	permissions permission.Service,
	sessions session.Service,
	messages message.Service,
	history history.Service,
	lspClients map[string]*lsp.Client,
) []tools.BaseTool {
	ctx := context.Background()
	mcpTools := GetMcpTools(ctx, permissions)
	
	return append([]tools.BaseTool{
		tools.NewViewTool(lspClients),          // Read data files
		tools.NewGrepTool(),                    // Pattern analysis
		tools.NewGlobTool(),                    // File discovery
		tools.NewBashTool(permissions),         // Data processing commands
		tools.NewLsTool(),                      // Directory analysis
		tools.NewSourcegraphTool(),             // Advanced search
		tools.NewTodoReadTool(),                // Task tracking
		tools.NewTodoWriteTool(),               // Task management
		tools.NewFetchTool(permissions),        // External data access
	}, mcpTools...) // Include MCP tools for enhanced analysis
}
