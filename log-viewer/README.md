# OpenCode Log Viewer

A standalone client for visualizing detailed HTTP logs from OpenCode's database. This application provides a modern, clean interface to explore sessions, HTTP requests/responses, LLM calls, and tool executions with color-coded, collapsible sections.

## Features

- **Session Management**: Browse and filter sessions with error status and search
- **Request Timeline**: Visualize the complete flow of LLM calls, tool executions, and HTTP requests
- **Color-Coded Components**: 
  - 🟦 LLM Calls (Blue)
  - 🟠 Tool Calls (Orange) 
  - 🟣 HTTP Calls (Purple)
- **Collapsible Sections**: Expand/collapse any section to focus on relevant data
- **Human-Readable Format**: View system prompts, agent responses, user inputs, tool calls/responses in structured, readable format
- **Real-time Data**: Direct connection to OpenCode's SQLite database and JSON session files

## Prerequisites

1. **OpenCode with detailed logging enabled**:
   ```bash
   opencode --detailed-logs
   ```
   This creates the database and session files in `~/.opencode/detailed_logs/`

## Installation

1. **Install dependencies**:
   ```bash
   # Frontend dependencies
   npm install
   
   # Backend dependencies
   cd server && npm install
   ```

2. **Start the application**:
   ```bash
   # Start both frontend and backend
   npm run dev:all
   
   # Or start separately:
   # Backend (port 3001)
   npm run dev:server
   
   # Frontend (port 3000)
   npm run dev
   ```

3. **Open in browser**:
   Navigate to `http://localhost:3000`

## Usage

### Session List
- View all sessions with key metrics (LLM calls, tool calls, HTTP calls, cost)
- Filter by success/error status
- Search by session ID or metadata
- Click any session to view details

### Session Detail View
- **Overview**: Session duration, total metrics, command arguments
- **Request Timeline**: Chronological view of all requests/responses
- **LLM Calls**: View complete request/response cycles including:
  - System prompts
  - User messages  
  - Assistant responses
  - Tool calls and function calls
  - Token usage and costs
  - Streaming events
- **Tool Calls**: View tool execution details including:
  - Input parameters
  - Output results
  - Execution duration
  - Error messages
  - Parent-child relationships
- **HTTP Calls**: View HTTP request/response pairs including:
  - Full URL and method
  - Request/response headers and bodies
  - Status codes
  - Error details

### Color Coding
- **Purple**: System prompts and configuration
- **Blue**: User inputs and messages
- **Green**: Assistant/agent responses
- **Orange**: Tool calls and function executions
- **Red**: Errors and function calls

## Architecture

### Frontend (React + TypeScript)
- **Framework**: Vite + React 18 + TypeScript
- **Styling**: Tailwind CSS with custom design system
- **State Management**: React Query for server state
- **Routing**: React Router for navigation
- **UI Components**: Custom components with Lucide icons

### Backend (Node.js + Express)
- **Database**: Direct SQLite connection to OpenCode's database
- **API**: RESTful endpoints for sessions and session details
- **File System**: Direct access to JSON session files
- **CORS**: Enabled for frontend communication

### Data Flow
1. OpenCode generates detailed logs in `~/.opencode/detailed_logs/`
2. Backend server connects to SQLite database and reads JSON files
3. Frontend queries backend API for session data
4. UI renders interactive, collapsible components for exploration

## API Endpoints

- `GET /api/sessions` - List sessions with filtering
- `GET /api/sessions/:sessionId` - Get complete session details
- `GET /health` - Health check and database status

## Development

### Project Structure
```
log-viewer/
├── src/
│   ├── components/     # React components
│   ├── services/       # API and database services
│   ├── types/          # TypeScript type definitions
│   └── main.tsx        # Application entry point
├── server/
│   ├── server.js       # Express backend server
│   └── package.json    # Backend dependencies
└── package.json        # Frontend dependencies
```

### Key Components
- **SessionList**: Paginated session browser with filters
- **SessionDetail**: Complete session view with metrics
- **LLMCallViewer**: Interactive LLM request/response viewer
- **ToolCallViewer**: Tool execution details with hierarchy
- **HTTPCallViewer**: HTTP request/response pair viewer
- **MessageViewer**: Color-coded message content viewer

## Troubleshooting

### "Database not found"
Ensure OpenCode has been run with `--detailed-logs` flag to generate the database.

### "Failed to load sessions"
Check that the backend server is running on port 3001 and can access `~/.opencode/detailed_logs/`.

### Empty session list
Verify that OpenCode has generated session data in the logs directory.

## License

This project is part of the OpenCode ecosystem.