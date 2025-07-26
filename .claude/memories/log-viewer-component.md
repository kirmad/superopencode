# SuperOpenCode Log Viewer

## Overview
The Log Viewer is a standalone Next.js 14 application for visualizing and analyzing HTTP request logs from LLM providers. It provides a modern web interface for debugging and understanding AI interactions.

## Technology Stack
- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript for full type safety
- **Styling**: Tailwind CSS utility-first framework
- **Database**: SQLite integration with existing logging database
- **Data Fetching**: React Query for efficient caching
- **UI**: Responsive design for desktop and mobile

## Key Features

### Session Management
- **Browse Sessions**: View all logged sessions with metadata
- **Search & Filter**: Powerful filtering capabilities
- **Status Indicators**: Color-coded badges for session states
- **Real-time Updates**: Live updates as new sessions are logged

### Request Visualization
- **HTTP Request Display**: Clean visualization of request/response pairs
- **Provider Filtering**: Filter by provider (GitHub Copilot, OpenAI, Anthropic, etc.)
- **Status Filtering**: Filter by success/error status
- **Timeline View**: Chronological display of requests

### Message Parsing & Display
- **Color-Coded Messages**: Visual distinction between message types
- **Collapsible Interface**: Expandable sections for detailed inspection
- **Syntax Highlighting**: JSON content formatting
- **Copy to Clipboard**: Easy content copying

#### Message Type Color Coding
- **System Messages**: Blue background (`bg-blue-50`) with 🤖 icon
- **User Messages**: Green background (`bg-green-50`) with 👤 icon
- **Tool Calls**: Purple background (`bg-purple-50`) with 🛠️ icon
- **Tool Responses**: Orange background (`bg-orange-50`) with ⚡ icon
- **Assistant Responses**: Gray background (`bg-gray-50`) with 🤖 icon

## Project Structure
```
log-viewer/
├── src/
│   ├── app/                 # Next.js App Router pages
│   │   ├── api/            # API route handlers
│   │   │   ├── sessions/   # Session management endpoints
│   │   │   └── requests/   # Request detail endpoints
│   │   ├── globals.css     # Global styles
│   │   ├── layout.tsx      # Root layout
│   │   └── page.tsx        # Home page
│   ├── components/         # React components
│   │   ├── layout/         # Layout components
│   │   ├── requests/       # Request-related components
│   │   ├── sessions/       # Session-related components
│   │   └── ui/             # UI utility components
│   ├── lib/                # Utilities and types
│   │   ├── database.ts     # Database connection
│   │   ├── parsers/        # Message parsing logic
│   │   ├── types/          # TypeScript type definitions
│   │   └── utils/          # Utility functions
│   └── hooks/              # Custom React hooks
├── public/                 # Static assets
├── .env.local              # Environment configuration
└── configuration files     # Next.js, Tailwind, TypeScript configs
```

## API Endpoints

### Sessions API
- **GET `/api/sessions`**: List all sessions with metadata
- **GET `/api/sessions/[id]`**: Get detailed session data with HTTP calls

### Requests API
- **GET `/api/requests/[id]`**: Get detailed information about specific HTTP request

## Configuration

### Environment Variables
- `DATA_DIR`: Path to detailed logging data directory
- `DATABASE_URL`: SQLite database connection string
- `NEXT_PUBLIC_APP_URL`: Public URL for the application
- `PORT`: Development server port (default: 3000)

### Database Schema
Expects detailed logging schema with:
- `sessions` table for metadata
- JSON files for detailed session data in `{DATA_DIR}/{session-id}.json` format

## Development Commands
- `npm run dev`: Start development server
- `npm run build`: Build for production
- `npm run start`: Start production server
- `npm run lint`: Run ESLint
- `npm run type-check`: Run TypeScript checks

## Integration with Main Application
Requires detailed logging to be enabled in the main SuperOpenCode application:
```go
import "github.com/kirmad/superopencode/internal/detailed_logging"
logger := detailed_logging.NewDetailedLogger("/path/to/data/dir")
```

## Use Cases
- **Debugging**: Analyze failed AI interactions
- **Performance Analysis**: Track response times and token usage
- **Provider Comparison**: Compare different LLM provider behaviors
- **Development**: Understand tool usage patterns and system behavior