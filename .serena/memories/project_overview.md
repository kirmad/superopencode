# SuperOpenCode Project Overview

## Project Purpose
OpenCode (SuperOpenCode) is a powerful terminal-based AI assistant for developers that brings AI assistance directly to your terminal. It provides a sophisticated Terminal User Interface (TUI) for interacting with various AI models to help with coding tasks, debugging, and development workflows.

**Key Features:**
- Interactive TUI built with Bubble Tea framework
- Support for 10+ AI providers (Anthropic Claude, OpenAI, GitHub Copilot, Google Gemini, Groq, etc.)
- Session management with persistent conversation storage
- Tool integration - AI can execute commands, search files, and modify code
- LSP integration for real-time code diagnostics
- MCP (Model Context Protocol) support for extensible tool system
- File versioning and change tracking
- Permission system for secure tool execution

## Current Status
⚠️ **Early Development Notice:** This project is in early development and is not yet ready for production use. Features may change, break, or be incomplete.

## Repository Information
- **GitHub:** https://github.com/kirmad/superopencode
- **License:** MIT
- **Go Version:** 1.24.0+
- **Original:** This is a fork/continuation of OpenCode, now being developed by Charm with the original creator

## Entry Points
1. **Interactive Mode:** `opencode` - Full TUI with all features
2. **CLI Mode:** `opencode -p "prompt"` - Single prompt execution  
3. **Debug Mode:** `opencode -d` - Enable debug logging and diagnostics

## Architecture
The application follows a modular, service-oriented architecture with clear separation of concerns across multiple packages within the `internal/` directory.