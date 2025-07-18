# HTTP Request Log Viewer - Setup Guide

## ✅ Implementation Complete

The HTTP Request Log Viewer has been successfully implemented with all core features:

- ✅ Next.js 15.4.1 with TypeScript and Tailwind CSS
- ✅ Three-panel layout (Sessions, Requests, Request Details)
- ✅ Database integration with SQLite and JSON file support
- ✅ Message parsers for GitHub Copilot, OpenAI, and Anthropic
- ✅ Color-coded collapsible message viewer
- ✅ React Query for efficient data fetching
- ✅ Responsive design with proper error handling

## Quick Start

### 1. Install Dependencies
```bash
npm install
```

### 2. Environment Configuration
Copy the example environment file:
```bash
cp .env.local.example .env.local
```

Edit `.env.local` to point to your data directory:
```env
DATA_DIR=../data
DATABASE_URL=sqlite:../data/sessions.db
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### 3. Start Development Server
```bash
npm run dev
```

The application will be available at `http://localhost:3000`

## Project Structure

```
log-viewer/
├── src/
│   ├── app/
│   │   ├── api/                    # API routes
│   │   │   ├── sessions/          # Sessions API
│   │   │   └── requests/          # Request details API
│   │   ├── layout.tsx             # Root layout
│   │   ├── page.tsx               # Main application page
│   │   └── globals.css            # Global styles with custom colors
│   ├── components/
│   │   ├── layout/
│   │   │   └── Layout.tsx         # App wrapper with React Query
│   │   ├── sessions/
│   │   │   ├── SessionList.tsx    # Session browser
│   │   │   └── SessionCard.tsx    # Individual session item
│   │   ├── requests/
│   │   │   ├── RequestList.tsx    # HTTP request browser
│   │   │   ├── RequestCard.tsx    # Individual request item
│   │   │   └── RequestDetail.tsx  # Request detail viewer
│   │   └── ui/
│   │       └── CollapsibleMessageViewer.tsx # Message display component
│   ├── lib/
│   │   ├── database.ts            # Database connection and queries
│   │   ├── types/
│   │   │   └── index.ts           # TypeScript interfaces
│   │   ├── parsers/
│   │   │   ├── index.ts           # Main parser dispatcher
│   │   │   └── copilot.ts         # GitHub Copilot parser
│   │   └── utils/
│   │       └── colors.ts          # Color theme definitions
│   └── hooks/                     # Custom React hooks (future)
├── .env.local.example             # Environment template
├── next.config.ts                 # Next.js configuration
├── tailwind.config.ts             # Tailwind CSS configuration
└── package.json                   # Dependencies and scripts
```

## Features Implemented

### 🎯 Core Features
- **Session Management**: Browse, filter, and search through logged sessions
- **Request Visualization**: Clean display of HTTP requests and responses
- **Message Parsing**: Human-readable formatting for GitHub Copilot, OpenAI, and Anthropic
- **Real-time UI**: Responsive interface with loading states and error handling
- **Color Coding**: Intuitive colors for different message types:
  - 🤖 **System** (Blue): System prompts and instructions
  - 👤 **User** (Green): User input and queries
  - 🤖 **Assistant** (Gray): AI-generated responses
  - 🛠️ **Tool Calls** (Purple): Function calls with formatted arguments
  - ⚡ **Tool Responses** (Orange): Tool execution results
  - ❌ **Errors** (Red): Error messages and failed requests

### 🎨 User Interface
- **Three-Panel Layout**: Sessions → Requests → Details
- **Collapsible Sections**: Expandable/collapsible message content
- **Search & Filter**: Filter sessions by status, requests by provider
- **Responsive Design**: Works on desktop and mobile
- **Copy Functionality**: Click to copy message content

### 🔧 Technical Features
- **TypeScript**: Full type safety throughout
- **React Query**: Efficient data fetching and caching
- **Database Integration**: SQLite with JSON session files
- **Provider Support**: GitHub Copilot, OpenAI, Anthropic, with extensible parser system

## Development Commands

```bash
# Development
npm run dev          # Start development server
npm run build        # Build for production
npm run start        # Start production server

# Code Quality
npm run lint         # Run ESLint
npm run type-check   # TypeScript type checking
```

## Database Schema

The viewer expects:
1. **SQLite Database** (`sessions.db`) with session metadata
2. **JSON Files** for detailed session data (`{session-id}.json`)

### Session Metadata Table
```sql
CREATE TABLE sessions (
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

### JSON File Structure
```json
{
  "id": "session-id",
  "startTime": "2024-01-01T00:00:00Z",
  "endTime": "2024-01-01T00:05:00Z",
  "httpCalls": [
    {
      "id": "request-id",
      "method": "POST",
      "url": "https://api.githubcopilot.com/chat/completions",
      "headers": {},
      "body": {},
      "statusCode": 200,
      "responseBody": {},
      "startTime": "2024-01-01T00:00:00Z",
      "endTime": "2024-01-01T00:00:01Z",
      "durationMs": 1000
    }
  ]
}
```

## Integration with SuperOpenCode

To populate the log viewer with data, ensure detailed logging is enabled in your main SuperOpenCode application:

```go
import "github.com/kirmad/superopencode/internal/detailed_logging"

// Configure detailed logging
logger := detailed_logging.NewDetailedLogger("/path/to/data")
```

## Troubleshooting

### Common Issues

**No sessions found:**
- Check `DATA_DIR` path in `.env.local`
- Verify `sessions.db` exists and has correct permissions
- Ensure detailed logging is enabled in main application

**Build errors:**
- Run `npm install` to ensure all dependencies are installed
- Check TypeScript errors with `npm run type-check`

**Database connection issues:**
- Verify database path in environment variables
- Check file permissions for database and JSON files

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Performance Notes

- Optimized for sessions with 100+ HTTP requests
- Lazy loading for large message content
- Efficient React Query caching (5min stale time)
- Virtual scrolling recommended for 1000+ sessions

---

**🎉 The HTTP Request Log Viewer is ready for use!**

Start your development server with `npm run dev` and navigate to `http://localhost:3000` to begin exploring your HTTP request logs.