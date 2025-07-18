# HTTP Request Log Viewer

A modern, standalone Next.js application for visualizing and analyzing HTTP request logs from LLM providers including GitHub Copilot, OpenAI, Anthropic, and others.

## Overview

The HTTP Request Log Viewer transforms raw HTTP request logs into a clean, navigable interface that allows you to:

- **Browse Sessions**: View all logged sessions with metadata and statistics
- **Analyze HTTP Requests**: Inspect detailed request/response pairs for each LLM call
- **Parse Chat Messages**: Extract and format system prompts, user messages, tool calls, and responses
- **Color-Coded Visualization**: Intuitive visual distinction between different message types
- **Collapsible Interface**: Compact, expandable sections for detailed inspection
- **Multi-Provider Support**: Works with GitHub Copilot, OpenAI, Anthropic, and other providers

## Features

### 🎯 Core Features
- **Session Management**: Browse, filter, and search through logged sessions
- **Request Visualization**: Clean display of HTTP requests and responses
- **Message Parsing**: Human-readable formatting of chat completion components
- **Real-time Updates**: Live updates as new sessions are logged
- **Error Tracking**: Highlight sessions and requests with errors
- **Performance Metrics**: Token usage, response times, and cost tracking

### 🎨 User Interface
- **Modern Design**: Clean, responsive interface built with Tailwind CSS
- **Color Coding**: Intuitive colors for different message types
- **Collapsible Sections**: Expandable/collapsible content for better organization
- **Search & Filter**: Powerful filtering and search capabilities
- **Responsive Layout**: Works on desktop and mobile devices

### 🔧 Technical Features
- **Next.js 14**: Built with the latest Next.js App Router
- **TypeScript**: Full type safety throughout the application
- **SQLite Integration**: Direct integration with existing logging database
- **React Query**: Efficient data fetching and caching
- **Tailwind CSS**: Utility-first CSS framework for styling

## Prerequisites

Before setting up the HTTP Request Log Viewer, ensure you have:

- **Node.js** (v18 or higher)
- **npm** or **yarn** package manager
- **Access to the SuperOpenCode database** with detailed logging enabled
- **Git** for version control

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/superopencode.git
cd superopencode/log-viewer
```

### 2. Install Dependencies

```bash
npm install
# or
yarn install
```

### 3. Environment Setup

Create a `.env.local` file in the log-viewer directory:

```env
# Database Configuration
DATA_DIR=/path/to/your/detailed-logging/data
DATABASE_URL=sqlite:///path/to/sessions.db

# Application Configuration
NEXT_PUBLIC_APP_URL=http://localhost:3000
NODE_ENV=development

# Optional: Custom port
PORT=3000
```

### 4. Database Schema

Ensure your database has the detailed logging schema. The viewer expects:

```sql
-- Sessions metadata table
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    llm_call_count INTEGER DEFAULT 0,
    tool_call_count INTEGER DEFAULT 0,
    http_call_count INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    total_cost REAL DEFAULT 0,
    has_error BOOLEAN DEFAULT 0,
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Plus JSON files for detailed session data in the format:
```
{DATA_DIR}/
├── sessions.db
├── {session-id-1}.json
├── {session-id-2}.json
└── ...
```

## Getting Started

### 1. Start the Development Server

```bash
npm run dev
# or
yarn dev
```

The application will be available at `http://localhost:3000`

### 2. Enable Detailed Logging

To populate the log viewer with data, ensure detailed logging is enabled in your main SuperOpenCode application:

```go
// In your main application
import "github.com/kirmad/superopencode/internal/detailed_logging"

// Enable detailed logging
logger := detailed_logging.NewDetailedLogger("/path/to/data/dir")
```

### 3. Access the Interface

Navigate to `http://localhost:3000` in your browser to access the log viewer interface.

## Usage Guide

### Session Browser

The main interface displays a list of all logged sessions:

- **Session List**: Left sidebar showing all available sessions
- **Session Details**: Click on a session to view its HTTP requests
- **Search & Filter**: Use the search bar to find specific sessions
- **Status Indicators**: Color-coded badges show session status and error states

### Request Viewer

When you select a session, you'll see:

- **Request List**: All HTTP requests made during that session
- **Provider Filtering**: Filter requests by provider (Copilot, OpenAI, etc.)
- **Status Filtering**: Filter by success/error status
- **Timeline View**: Chronological order of requests

### Message Inspector

Click on any HTTP request to see its detailed breakdown:

#### System Messages
- **Color**: Blue background (`bg-blue-50`)
- **Icon**: 🤖
- **Content**: System prompts and instructions

#### User Messages
- **Color**: Green background (`bg-green-50`)
- **Icon**: 👤
- **Content**: User input and queries

#### Tool Calls
- **Color**: Purple background (`bg-purple-50`)
- **Icon**: 🛠️
- **Content**: Function calls with arguments formatted as JSON

#### Tool Responses
- **Color**: Orange background (`bg-orange-50`)
- **Icon**: ⚡
- **Content**: Tool execution results

#### Assistant Responses
- **Color**: Gray background (`bg-gray-50`)
- **Icon**: 🤖
- **Content**: AI-generated responses

### Collapsible Sections

All message types are displayed in collapsible sections:

- **Click to expand/collapse** any section
- **Syntax highlighting** for JSON content
- **Copy to clipboard** functionality
- **Full-text search** within messages

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATA_DIR` | Path to detailed logging data directory | `./data` |
| `DATABASE_URL` | SQLite database connection string | `sqlite://./data/sessions.db` |
| `NEXT_PUBLIC_APP_URL` | Public URL for the application | `http://localhost:3000` |
| `PORT` | Port for the development server | `3000` |

### Customization

#### Color Themes

Modify the color scheme in `src/lib/utils/colors.ts`:

```typescript
export const MESSAGE_COLORS = {
  system: {
    bg: "bg-blue-50",
    border: "border-blue-200",
    text: "text-blue-900",
    icon: "🤖"
  },
  // ... customize other colors
}
```

#### Message Formatting

Extend message parsing in `src/lib/parsers/`:

```typescript
// Add custom provider parsing
export const parseCustomProvider = (request: HTTPLog) => {
  // Custom parsing logic
}
```

## Development

### Project Structure

```
log-viewer/
├── src/
│   ├── app/                 # Next.js App Router pages
│   ├── components/          # React components
│   ├── lib/                 # Utilities and types
│   ├── hooks/               # Custom React hooks
│   └── api/                 # API route handlers
├── public/                  # Static assets
├── .env.local              # Environment variables
├── next.config.js          # Next.js configuration
├── tailwind.config.js      # Tailwind CSS configuration
└── package.json            # Dependencies and scripts
```

### Available Scripts

```bash
# Development
npm run dev          # Start development server
npm run build        # Build for production
npm run start        # Start production server

# Code Quality
npm run lint         # Run ESLint
npm run type-check   # Run TypeScript checks
npm run format       # Format code with Prettier

# Testing
npm run test         # Run tests
npm run test:watch   # Run tests in watch mode
```

### Building for Production

```bash
npm run build
npm run start
```

## API Reference

### Sessions API

#### GET `/api/sessions`

Returns a list of all sessions with metadata.

**Response:**
```json
[
  {
    "id": "session-id",
    "sessionId": "session-id",
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "2024-01-01T00:05:00Z",
    "llmCallCount": 5,
    "toolCallCount": 12,
    "httpCallCount": 8,
    "totalTokens": 2500,
    "totalCost": 0.05,
    "hasError": false
  }
]
```

#### GET `/api/sessions/[id]`

Returns detailed session data including all HTTP calls.

**Response:**
```json
{
  "id": "session-id",
  "startTime": "2024-01-01T00:00:00Z",
  "endTime": "2024-01-01T00:05:00Z",
  "metadata": {},
  "llmCalls": [...],
  "toolCalls": [...],
  "httpCalls": [
    {
      "id": "http-call-id",
      "method": "POST",
      "url": "https://api.githubcopilot.com/chat/completions",
      "headers": {...},
      "body": {...},
      "statusCode": 200,
      "responseBody": {...},
      "startTime": "2024-01-01T00:00:00Z",
      "endTime": "2024-01-01T00:00:01Z",
      "durationMs": 1000
    }
  ]
}
```

### Request API

#### GET `/api/requests/[id]`

Returns detailed information about a specific HTTP request.

## Troubleshooting

### Common Issues

#### Database Connection Error
```
Error: failed to open database: no such file or directory
```
**Solution**: Ensure the `DATA_DIR` environment variable points to a valid directory with the sessions.db file.

#### No Sessions Displayed
**Possible causes:**
- Detailed logging is not enabled in the main application
- Database is empty or corrupted
- Incorrect `DATA_DIR` path

**Solution**: Verify detailed logging is enabled and check the database path.

#### Port Already in Use
```
Error: listen EADDRINUSE: address already in use :::3000
```
**Solution**: Change the port in `.env.local` or kill the process using port 3000.

### Performance Issues

#### Slow Loading
- **Large datasets**: Implement pagination for sessions with many HTTP calls
- **Memory usage**: Monitor memory usage with large JSON files
- **Database queries**: Add indexes for frequently queried fields

#### Browser Performance
- **Large responses**: Implement virtual scrolling for large response bodies
- **Memory leaks**: Use React.memo and useCallback for expensive renders

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Make your changes and add tests
4. Commit your changes: `git commit -m "Add new feature"`
5. Push to the branch: `git push origin feature/new-feature`
6. Submit a pull request

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Support

For issues and questions:

- **GitHub Issues**: https://github.com/your-org/superopencode/issues
- **Documentation**: https://docs.superopencode.com
- **Community**: Join our Discord server

## Changelog

### v1.0.0 (Initial Release)
- ✅ Session browsing and filtering
- ✅ HTTP request visualization
- ✅ Message parsing and formatting
- ✅ Color-coded UI with collapsible sections
- ✅ Multi-provider support (Copilot, OpenAI, Anthropic)
- ✅ Next.js 14 with App Router
- ✅ TypeScript support
- ✅ Tailwind CSS styling
- ✅ React Query data management