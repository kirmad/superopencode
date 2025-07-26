# SuperOpenCode Project Overview

## Project Description
SuperOpenCode is a Go-based terminal AI assistant for developers, providing intelligent coding assistance directly in the terminal. Originally known as OpenCode, it's now continuing development under a new name as it prepares for a public relaunch.

## Key Features
- **Interactive TUI**: Built with Bubble Tea for smooth terminal experience
- **Multiple AI Providers**: OpenAI, Anthropic Claude, Google Gemini, AWS Bedrock, Groq, Azure OpenAI, OpenRouter
- **Session Management**: Save and manage multiple conversation sessions
- **Tool Integration**: AI can execute commands, search files, and modify code
- **LSP Integration**: Language Server Protocol support for code intelligence
- **MCP Support**: Model Context Protocol for external tool integration
- **Log Viewer**: Next.js web application for visualizing HTTP request logs

## Architecture Overview
- **cmd/**: Command-line interface using Cobra
- **internal/app**: Core application services
- **internal/config**: Configuration management
- **internal/db**: Database operations and migrations (SQLite)
- **internal/llm**: LLM providers and tools integration
- **internal/tui**: Terminal UI components using Bubble Tea
- **internal/logging**: Logging infrastructure with detailed logging support
- **internal/lsp**: Language Server Protocol integration
- **log-viewer/**: Next.js web application for log visualization

## Technology Stack
- **Backend**: Go 1.24.0
- **UI Framework**: Charm's Bubble Tea (TUI)
- **Database**: SQLite with migrations (goose)
- **Log Viewer**: Next.js 14, TypeScript, Tailwind CSS
- **LSP**: Language Server Protocol integration
- **MCP**: Model Context Protocol support

## Development Status
⚠️ Early development - not ready for production use. Features may change, break, or be incomplete.