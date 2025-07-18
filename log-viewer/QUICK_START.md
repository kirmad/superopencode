# Quick Start Guide - HTTP Request Log Viewer

A step-by-step guide to get the HTTP Request Log Viewer up and running in under 10 minutes.

## Prerequisites

- Node.js 18+ installed
- Access to SuperOpenCode project with detailed logging enabled
- Terminal/command line access

## Step 1: Setup (2 minutes)

```bash
# Navigate to the log-viewer directory
cd /path/to/superopencode/log-viewer

# Install dependencies
npm install

# Install additional required packages
npm install @tanstack/react-query @tanstack/react-query-devtools
npm install better-sqlite3 @types/better-sqlite3
npm install lucide-react clsx tailwind-merge
```

## Step 2: Configure Environment (1 minute)

Create `.env.local` file:

```env
# Point to your detailed logging data directory
DATA_DIR=../data
DATABASE_URL=sqlite:../data/sessions.db
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Step 3: Quick Implementation (5 minutes)

### Create Basic Project Structure

```bash
# Create required directories
mkdir -p src/lib/{types,parsers,utils}
mkdir -p src/components/{ui,sessions,requests,layout}
mkdir -p src/hooks
mkdir -p src/app/api/{sessions,requests}
```

### Essential Files to Create

1. **Database Connection** (`src/lib/database.ts`)
2. **Message Parser** (`src/lib/parsers/copilot.ts`)
3. **Session List** (`src/components/sessions/SessionList.tsx`)
4. **Request Detail** (`src/components/requests/RequestDetail.tsx`)
5. **Main Layout** (`src/components/layout/Layout.tsx`)

### Copy Key Configuration

Update `next.config.js`:
```javascript
const nextConfig = {
  experimental: {
    serverComponentsExternalPackages: ['better-sqlite3']
  }
}
module.exports = nextConfig
```

Update `tailwind.config.js` with color themes:
```javascript
module.exports = {
  // ... existing config
  theme: {
    extend: {
      colors: {
        system: { 50: '#eff6ff', 200: '#bfdbfe', 900: '#1e3a8a' },
        user: { 50: '#f0fdf4', 200: '#bbf7d0', 900: '#14532d' },
        assistant: { 50: '#f9fafb', 200: '#e5e7eb', 900: '#111827' },
        tool: { 50: '#faf5ff', 200: '#e9d5ff', 900: '#581c87' },
        toolResponse: { 50: '#fff7ed', 200: '#fed7aa', 900: '#9a3412' },
      }
    }
  }
}
```

## Step 4: Start Development (1 minute)

```bash
# Start the development server
npm run dev
```

Open `http://localhost:3000` in your browser.

## Step 5: Test with Sample Data (1 minute)

Ensure your SuperOpenCode application has detailed logging enabled:

```go
// In your main application (if not already enabled)
import "github.com/kirmad/superopencode/internal/detailed_logging"

// Configure detailed logging
logger := detailed_logging.NewDetailedLogger("/path/to/data")
```

Run a few LLM requests to populate the database, then refresh the log viewer.

## Expected Result

You should see:
- ✅ Session list in the left sidebar
- ✅ HTTP request list in the middle panel
- ✅ Detailed request/response view in the right panel
- ✅ Color-coded message types (system, user, assistant, tool calls)
- ✅ Collapsible sections for each message
- ✅ Provider-specific parsing (especially GitHub Copilot)

## Common Issues & Solutions

### Issue: "No sessions found"
**Solution**: 
1. Check that `DATA_DIR` points to correct directory
2. Verify detailed logging is enabled in main application
3. Ensure database file exists and has correct permissions

### Issue: "Cannot resolve module 'better-sqlite3'"
**Solution**:
```bash
npm install better-sqlite3 --save
npm install @types/better-sqlite3 --save-dev
```

### Issue: Database connection error
**Solution**:
1. Verify database path in `.env.local`
2. Check file permissions
3. Ensure SQLite database exists

### Issue: Styling not working
**Solution**:
1. Verify Tailwind configuration
2. Check that custom colors are defined
3. Restart development server

## Next Steps

1. **Customize Parsing**: Modify `src/lib/parsers/` to handle your specific providers
2. **Add Features**: Implement search, filtering, and export functionality
3. **Improve UI**: Enhance styling and add animations
4. **Add Tests**: Create unit and integration tests
5. **Deploy**: Set up production deployment

## File Structure Reference

```
log-viewer/
├── src/
│   ├── app/
│   │   ├── page.tsx              # Main page
│   │   ├── layout.tsx            # Root layout
│   │   └── api/
│   │       ├── sessions/
│   │       │   └── route.ts      # Sessions API
│   │       └── requests/
│   │           └── [id]/route.ts # Request detail API
│   ├── components/
│   │   ├── layout/
│   │   │   └── Layout.tsx        # App layout wrapper
│   │   ├── sessions/
│   │   │   ├── SessionList.tsx   # Session browser
│   │   │   └── SessionCard.tsx   # Session item
│   │   ├── requests/
│   │   │   ├── RequestList.tsx   # Request browser
│   │   │   ├── RequestCard.tsx   # Request item
│   │   │   └── RequestDetail.tsx # Request detail view
│   │   └── ui/
│   │       └── CollapsibleMessageViewer.tsx
│   ├── lib/
│   │   ├── database.ts           # Database integration
│   │   ├── parsers/
│   │   │   └── copilot.ts        # Copilot message parser
│   │   ├── types/
│   │   │   └── index.ts          # Type definitions
│   │   └── utils/
│   │       └── colors.ts         # Color theme
│   └── hooks/
│       └── useSession.ts         # Data fetching hooks
├── .env.local                    # Environment variables
├── next.config.js               # Next.js configuration
├── tailwind.config.js           # Tailwind configuration
└── package.json                 # Dependencies
```

## Performance Tips

1. **Database Indexing**: Ensure your SQLite database has proper indexes
2. **Query Optimization**: Use React Query for efficient data fetching
3. **Virtual Scrolling**: For large datasets, consider virtual scrolling
4. **Code Splitting**: Use dynamic imports for heavy components

## Support

For issues with implementation:
1. Check the detailed [IMPLEMENTATION.md](./IMPLEMENTATION.md) guide
2. Review the [README.md](./README.md) for complete documentation
3. Open an issue in the repository

---

**Total Setup Time: ~10 minutes**

This guide gets you from zero to a working HTTP Request Log Viewer in under 10 minutes. For advanced features and customization, refer to the complete implementation documentation.