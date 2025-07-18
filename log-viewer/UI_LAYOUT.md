# UI Layout Documentation - HTTP Request Log Viewer

Comprehensive documentation of the user interface layout, design system, and interaction patterns for the HTTP Request Log Viewer.

## Table of Contents

1. [Overall Layout Architecture](#overall-layout-architecture)
2. [Component Layout Specifications](#component-layout-specifications)
3. [Design System](#design-system)
4. [Responsive Design](#responsive-design)
5. [Interaction Patterns](#interaction-patterns)
6. [Visual States](#visual-states)
7. [Accessibility](#accessibility)

## Overall Layout Architecture

### Main Layout Structure

The application uses a **three-panel layout** optimized for log analysis workflows:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Header (64px)                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│          │                    │                                                │
│ Sessions │   HTTP Requests    │            Request Details                     │
│  Panel   │      Panel         │               Panel                            │
│ (320px)  │     (384px)        │            (remaining)                         │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
│          │                    │                                                │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Layout Dimensions

```css
/* Desktop Layout (1200px+) */
.main-layout {
  display: flex;
  height: 100vh;
}

.sessions-panel {
  width: 320px;
  min-width: 280px;
  max-width: 400px;
  border-right: 1px solid #e5e7eb;
}

.requests-panel {
  width: 384px;
  min-width: 320px;
  max-width: 480px;
  border-right: 1px solid #e5e7eb;
}

.details-panel {
  flex: 1;
  min-width: 400px;
}

/* Header */
.header {
  height: 64px;
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}
```

## Component Layout Specifications

### 1. Header Component

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  [Logo/Icon]  HTTP Request Log Viewer                        [Settings] [Help]  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Specifications:**
- Height: `64px`
- Background: `#ffffff`
- Border: `1px solid #e5e7eb` (bottom)
- Padding: `16px 24px`
- Shadow: `0 1px 3px 0 rgba(0, 0, 0, 0.1)`

```jsx
<header className="h-16 bg-white border-b border-gray-200 shadow-sm">
  <div className="flex items-center justify-between h-full px-6">
    <div className="flex items-center space-x-3">
      <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
        <span className="text-white text-lg">🔍</span>
      </div>
      <h1 className="text-xl font-semibold text-gray-900">
        HTTP Request Log Viewer
      </h1>
    </div>
    
    <div className="flex items-center space-x-2">
      <button className="p-2 hover:bg-gray-100 rounded-lg">
        <SettingsIcon className="w-5 h-5 text-gray-600" />
      </button>
      <button className="p-2 hover:bg-gray-100 rounded-lg">
        <HelpIcon className="w-5 h-5 text-gray-600" />
      </button>
    </div>
  </div>
</header>
```

### 2. Sessions Panel

```
┌─────────────────────────────────┐
│ Sessions                        │
│ ┌─────────────────────────────┐ │
│ │ 🔍 Search sessions...       │ │
│ └─────────────────────────────┘ │
│ ┌─────────────────────────────┐ │
│ │ [All] [Success] [Error]     │ │
│ └─────────────────────────────┘ │
│ ┌─────────────────────────────┐ │
│ │ 📋 session-abc123           │ │
│ │ 🕐 2024-01-15 14:30:25     │ │
│ │ [5 LLM] [8 HTTP] [12 Tools] │ │
│ │ 2,500 tokens • $0.0125     │ │
│ └─────────────────────────────┘ │
│ ┌─────────────────────────────┐ │
│ │ 📋 session-def456      ⚠️   │ │
│ │ 🕐 2024-01-15 14:25:10     │ │
│ │ [3 LLM] [4 HTTP] [6 Tools]  │ │
│ │ 1,200 tokens • $0.0060     │ │
│ └─────────────────────────────┘ │
│ ...                             │
└─────────────────────────────────┘
```

**Layout Structure:**
```jsx
<div className="w-80 h-full flex flex-col bg-white border-r border-gray-200">
  {/* Header */}
  <div className="p-4 border-b border-gray-200">
    <h2 className="text-lg font-semibold text-gray-900 mb-3">Sessions</h2>
    
    {/* Search Bar */}
    <div className="relative mb-3">
      <SearchIcon className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
      <input
        type="text"
        placeholder="Search sessions..."
        className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      />
    </div>
    
    {/* Filter Buttons */}
    <div className="flex space-x-2">
      <button className="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-lg">
        All
      </button>
      <button className="px-3 py-1 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200">
        Success
      </button>
      <button className="px-3 py-1 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200">
        Error
      </button>
    </div>
  </div>
  
  {/* Session List */}
  <div className="flex-1 overflow-y-auto">
    {/* Session items */}
  </div>
</div>
```

### 3. Session Card Component

```
┌─────────────────────────────────────────────────────────────────┐
│ 📋 session-abc123                                        [⚠️]   │
│ 🕐 2024-01-15 14:30:25                                         │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐                           │
│ │ 5 LLM   │ │ 8 HTTP  │ │ 12 Tools│                           │
│ └─────────┘ └─────────┘ └─────────┘                           │
│ 2,500 tokens                                       $0.0125    │
└─────────────────────────────────────────────────────────────────┘
```

**Visual States:**
- **Default**: `border-gray-200 hover:bg-gray-50`
- **Selected**: `border-blue-200 bg-blue-50`
- **With Error**: Red warning icon visible
- **Loading**: Skeleton animation

```jsx
<div className={`p-4 border-b cursor-pointer transition-colors ${
  isSelected 
    ? 'bg-blue-50 border-blue-200' 
    : 'hover:bg-gray-50 border-gray-200'
}`}>
  <div className="flex justify-between items-start mb-2">
    <div className="flex items-center space-x-2">
      <span className="text-lg">📋</span>
      <h3 className="font-medium text-sm text-gray-900 truncate">
        {session.id}
      </h3>
    </div>
    {session.hasError && (
      <AlertCircle className="h-4 w-4 text-red-500" />
    )}
  </div>
  
  <div className="flex items-center space-x-2 text-xs text-gray-600 mb-2">
    <Clock className="h-3 w-3" />
    <span>{formatTime(session.startTime)}</span>
  </div>
  
  <div className="flex space-x-2 mb-2">
    <Badge variant="blue">{session.llmCallCount} LLM</Badge>
    <Badge variant="green">{session.httpCallCount} HTTP</Badge>
    <Badge variant="purple">{session.toolCallCount} Tools</Badge>
  </div>
  
  <div className="flex justify-between items-center text-xs text-gray-600">
    <span>{session.totalTokens.toLocaleString()} tokens</span>
    <span>${session.totalCost.toFixed(4)}</span>
  </div>
</div>
```

### 4. HTTP Requests Panel

```
┌─────────────────────────────────────────────────────────────────┐
│ HTTP Requests                                                   │
│ ┌─────────────────┐ ┌─────────────────┐                       │
│ │ All Providers ▼ │ │ All Status    ▼ │                       │
│ └─────────────────┘ └─────────────────┘                       │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ [POST] GitHub Copilot                            [200] 1.5s │ │
│ │ 🕐 2024-01-15 14:30:25                                ✅   │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ [POST] OpenAI                                    [429] 2.1s │ │
│ │ 🕐 2024-01-15 14:30:23                                ❌   │ │
│ │ Rate limit exceeded                                         │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ...                                                             │
└─────────────────────────────────────────────────────────────────┘
```

**Layout Structure:**
```jsx
<div className="w-96 h-full flex flex-col bg-white border-r border-gray-200">
  {/* Header */}
  <div className="p-4 border-b border-gray-200">
    <h3 className="text-lg font-semibold text-gray-900 mb-3">HTTP Requests</h3>
    
    {/* Filter Controls */}
    <div className="flex space-x-2">
      <select className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500">
        <option>All Providers</option>
        <option>GitHub Copilot</option>
        <option>OpenAI</option>
        <option>Anthropic</option>
      </select>
      
      <select className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500">
        <option>All Status</option>
        <option>Success</option>
        <option>Error</option>
      </select>
    </div>
  </div>
  
  {/* Request List */}
  <div className="flex-1 overflow-y-auto">
    {/* Request items */}
  </div>
</div>
```

### 5. Request Card Component

```
┌─────────────────────────────────────────────────────────────────┐
│ [POST] GitHub Copilot                            [200] 1.5s ✅ │
│ 🕐 2024-01-15 14:30:25                                         │
│ api.githubcopilot.com/chat/completions                         │
└─────────────────────────────────────────────────────────────────┘
```

**Status Colors:**
- **Success (2xx)**: `bg-green-100 text-green-800`
- **Client Error (4xx)**: `bg-yellow-100 text-yellow-800`
- **Server Error (5xx)**: `bg-red-100 text-red-800`

```jsx
<div className={`p-4 border-b cursor-pointer transition-colors ${
  isSelected 
    ? 'bg-blue-50 border-blue-200' 
    : 'hover:bg-gray-50 border-gray-200'
}`}>
  <div className="flex justify-between items-start mb-2">
    <div className="flex items-center space-x-2">
      <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
        {request.method}
      </span>
      <span className="text-sm font-medium text-gray-900">
        {getProviderName(request.url)}
      </span>
    </div>
    
    <div className="flex items-center space-x-2">
      <StatusBadge status={request.statusCode} />
      <span className="text-xs text-gray-500">{request.durationMs}ms</span>
      <StatusIcon status={request.statusCode} />
    </div>
  </div>
  
  <div className="flex items-center space-x-2 text-xs text-gray-600 mb-1">
    <Clock className="h-3 w-3" />
    <span>{formatTime(request.startTime)}</span>
  </div>
  
  <div className="text-xs text-gray-500 truncate">
    {extractDomain(request.url)}
  </div>
  
  {request.error && (
    <div className="mt-2 text-xs text-red-600 bg-red-50 p-2 rounded">
      {request.error}
    </div>
  )}
</div>
```

### 6. Request Details Panel

```
┌─────────────────────────────────────────────────────────────────┐
│ Request Details                                         1.5s    │
│ [POST] https://api.githubcopilot.com/chat/completions  [200] ✅ │
│ 🕐 2024-01-15 14:30:25                                         │
│ ─────────────────────────────────────────────────────────────── │
│                                                                 │
│ Messages                                                        │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 🤖 SYSTEM                                              > │ │
│ │ You are a helpful assistant specialized in...              │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 👤 USER                                                > │ │
│ │ How do I implement authentication in Next.js?             │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 🛠️ TOOL CALL (Auth.verify)                            > │ │
│ │ Function: Auth.verify                                      │ │
│ │ Arguments: { "token": "abc123" }                          │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ ⚡ TOOL RESPONSE                                       > │ │
│ │ { "valid": true, "user": "john@example.com" }            │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 🤖 ASSISTANT                                           > │ │
│ │ Here's how to implement authentication in Next.js...      │ │
│ └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**Layout Structure:**
```jsx
<div className="flex-1 h-full flex flex-col bg-gray-50">
  {/* Header */}
  <div className="p-4 border-b bg-white">
    <div className="flex justify-between items-center mb-2">
      <h3 className="text-lg font-semibold text-gray-900">Request Details</h3>
      <div className="flex items-center space-x-2 text-sm text-gray-600">
        <Clock className="h-4 w-4" />
        <span>{request.durationMs}ms</span>
      </div>
    </div>
    
    <div className="flex items-center space-x-4 text-sm text-gray-600">
      <div className="flex items-center space-x-1">
        <span className="font-mono bg-gray-100 px-2 py-1 rounded">
          {request.method}
        </span>
        <span>{request.url}</span>
      </div>
      
      <StatusBadge status={request.statusCode} />
      <StatusIcon status={request.statusCode} />
    </div>
    
    <div className="mt-1 text-xs text-gray-500">
      {formatTime(request.startTime)}
    </div>
  </div>
  
  {/* Content */}
  <div className="flex-1 overflow-y-auto p-4">
    <div className="space-y-4">
      <div>
        <h4 className="font-medium text-gray-900 mb-3">Messages</h4>
        <div className="space-y-2">
          {messages.map((message, index) => (
            <CollapsibleMessageViewer
              key={index}
              message={message}
              defaultOpen={index === 0}
            />
          ))}
        </div>
      </div>
      
      {request.error && (
        <ErrorSection error={request.error} />
      )}
    </div>
  </div>
</div>
```

### 7. Collapsible Message Viewer

```
┌─────────────────────────────────────────────────────────────────┐
│ 🤖 SYSTEM                                            [📋] [>] │
│ ─────────────────────────────────────────────────────────────── │
│ You are a helpful assistant specialized in software            │
│ development. When responding to questions about code,          │
│ provide clear explanations and working examples.               │
│                                                                 │
│ Always format code blocks with proper syntax highlighting      │
│ and include comments explaining complex logic.                 │
└─────────────────────────────────────────────────────────────────┘
```

**Expanded States:**
- **Collapsed**: Shows only header with icon and title
- **Expanded**: Shows full content with copy button
- **Loading**: Skeleton animation in content area

```jsx
<div className={`border rounded-lg mb-3 ${message.colorClass} ${message.borderClass}`}>
  <button
    onClick={() => setIsOpen(!isOpen)}
    className="w-full p-3 text-left flex justify-between items-center hover:bg-opacity-80 transition-colors"
  >
    <div className="flex items-center space-x-3">
      <span className="text-lg">{message.icon}</span>
      <div>
        <span className="font-medium text-gray-900 uppercase text-sm">
          {message.type.replace('_', ' ')}
        </span>
        {message.metadata?.toolName && (
          <span className="text-sm text-gray-600 ml-2">
            ({message.metadata.toolName})
          </span>
        )}
      </div>
    </div>
    
    <div className="flex items-center space-x-2">
      <button
        onClick={(e) => {
          e.stopPropagation()
          copyToClipboard(message.content)
        }}
        className="p-1 hover:bg-gray-200 rounded"
        title="Copy content"
      >
        <Copy className="h-4 w-4" />
      </button>
      
      <ChevronRight 
        className={`h-4 w-4 transform transition-transform ${
          isOpen ? 'rotate-90' : ''
        }`}
      />
    </div>
  </button>
  
  {isOpen && (
    <div className="p-3 pt-0 border-t border-gray-200">
      <MessageContent message={message} />
    </div>
  )}
</div>
```

## Design System

### Color System

#### Message Type Colors

```css
/* System Messages */
.system-message {
  --bg-color: #eff6ff;        /* Blue 50 */
  --border-color: #bfdbfe;    /* Blue 200 */
  --text-color: #1e3a8a;      /* Blue 900 */
  --icon: '🤖';
}

/* User Messages */
.user-message {
  --bg-color: #f0fdf4;        /* Green 50 */
  --border-color: #bbf7d0;    /* Green 200 */
  --text-color: #14532d;      /* Green 900 */
  --icon: '👤';
}

/* Assistant Messages */
.assistant-message {
  --bg-color: #f9fafb;        /* Gray 50 */
  --border-color: #e5e7eb;    /* Gray 200 */
  --text-color: #111827;      /* Gray 900 */
  --icon: '🤖';
}

/* Tool Calls */
.tool-call {
  --bg-color: #faf5ff;        /* Purple 50 */
  --border-color: #e9d5ff;    /* Purple 200 */
  --text-color: #581c87;      /* Purple 900 */
  --icon: '🛠️';
}

/* Tool Responses */
.tool-response {
  --bg-color: #fff7ed;        /* Orange 50 */
  --border-color: #fed7aa;    /* Orange 200 */
  --text-color: #9a3412;      /* Orange 900 */
  --icon: '⚡';
}

/* Error Messages */
.error-message {
  --bg-color: #fef2f2;        /* Red 50 */
  --border-color: #fecaca;    /* Red 200 */
  --text-color: #991b1b;      /* Red 900 */
  --icon: '❌';
}
```

#### Status Colors

```css
/* Success States */
.status-success {
  --bg-color: #f0fdf4;        /* Green 50 */
  --border-color: #bbf7d0;    /* Green 200 */
  --text-color: #15803d;      /* Green 700 */
}

/* Warning States */
.status-warning {
  --bg-color: #fffbeb;        /* Yellow 50 */
  --border-color: #fde68a;    /* Yellow 200 */
  --text-color: #d97706;      /* Yellow 700 */
}

/* Error States */
.status-error {
  --bg-color: #fef2f2;        /* Red 50 */
  --border-color: #fecaca;    /* Red 200 */
  --text-color: #dc2626;      /* Red 700 */
}
```

#### Interactive States

```css
/* Hover States */
.hover-light {
  --hover-bg: #f9fafb;        /* Gray 50 */
}

.hover-primary {
  --hover-bg: #eff6ff;        /* Blue 50 */
}

/* Focus States */
.focus-ring {
  --focus-ring: 0 0 0 2px #3b82f6;  /* Blue 500 */
}

/* Active States */
.active-primary {
  --active-bg: #dbeafe;       /* Blue 100 */
  --active-border: #3b82f6;   /* Blue 500 */
}
```

### Typography Scale

```css
/* Headers */
.text-h1 {
  font-size: 2.25rem;         /* 36px */
  line-height: 2.5rem;        /* 40px */
  font-weight: 700;           /* Bold */
}

.text-h2 {
  font-size: 1.875rem;        /* 30px */
  line-height: 2.25rem;       /* 36px */
  font-weight: 600;           /* Semibold */
}

.text-h3 {
  font-size: 1.5rem;          /* 24px */
  line-height: 2rem;          /* 32px */
  font-weight: 600;           /* Semibold */
}

.text-h4 {
  font-size: 1.25rem;         /* 20px */
  line-height: 1.75rem;       /* 28px */
  font-weight: 600;           /* Semibold */
}

/* Body Text */
.text-body {
  font-size: 1rem;            /* 16px */
  line-height: 1.5rem;        /* 24px */
  font-weight: 400;           /* Normal */
}

.text-small {
  font-size: 0.875rem;        /* 14px */
  line-height: 1.25rem;       /* 20px */
  font-weight: 400;           /* Normal */
}

.text-xs {
  font-size: 0.75rem;         /* 12px */
  line-height: 1rem;          /* 16px */
  font-weight: 400;           /* Normal */
}

/* Code */
.text-code {
  font-family: 'SF Mono', Monaco, 'Inconsolata', 'Roboto Mono', monospace;
  font-size: 0.875rem;        /* 14px */
  line-height: 1.25rem;       /* 20px */
}
```

### Spacing System

```css
/* Padding Scale */
.p-1 { padding: 0.25rem; }    /* 4px */
.p-2 { padding: 0.5rem; }     /* 8px */
.p-3 { padding: 0.75rem; }    /* 12px */
.p-4 { padding: 1rem; }       /* 16px */
.p-5 { padding: 1.25rem; }    /* 20px */
.p-6 { padding: 1.5rem; }     /* 24px */
.p-8 { padding: 2rem; }       /* 32px */

/* Margin Scale */
.m-1 { margin: 0.25rem; }     /* 4px */
.m-2 { margin: 0.5rem; }      /* 8px */
.m-3 { margin: 0.75rem; }     /* 12px */
.m-4 { margin: 1rem; }        /* 16px */
.m-6 { margin: 1.5rem; }      /* 24px */
.m-8 { margin: 2rem; }        /* 32px */

/* Gap Scale */
.gap-1 { gap: 0.25rem; }      /* 4px */
.gap-2 { gap: 0.5rem; }       /* 8px */
.gap-3 { gap: 0.75rem; }      /* 12px */
.gap-4 { gap: 1rem; }         /* 16px */
.gap-6 { gap: 1.5rem; }       /* 24px */
```

### Border Radius

```css
.rounded-none { border-radius: 0; }
.rounded-sm { border-radius: 0.125rem; }  /* 2px */
.rounded { border-radius: 0.25rem; }      /* 4px */
.rounded-md { border-radius: 0.375rem; }  /* 6px */
.rounded-lg { border-radius: 0.5rem; }    /* 8px */
.rounded-xl { border-radius: 0.75rem; }   /* 12px */
.rounded-2xl { border-radius: 1rem; }     /* 16px */
.rounded-full { border-radius: 9999px; }
```

### Shadows

```css
.shadow-sm {
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.shadow {
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
}

.shadow-md {
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.shadow-lg {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}
```

## Responsive Design

### Breakpoints

```css
/* Mobile First Approach */
.container {
  width: 100%;
  padding: 0 1rem;
}

/* Tablet (768px+) */
@media (min-width: 768px) {
  .container {
    max-width: 768px;
    margin: 0 auto;
  }
  
  .main-layout {
    flex-direction: row;
  }
  
  .sessions-panel {
    width: 280px;
  }
  
  .requests-panel {
    width: 320px;
  }
}

/* Desktop (1024px+) */
@media (min-width: 1024px) {
  .container {
    max-width: 1024px;
  }
  
  .sessions-panel {
    width: 320px;
  }
  
  .requests-panel {
    width: 384px;
  }
}

/* Large Desktop (1280px+) */
@media (min-width: 1280px) {
  .container {
    max-width: 1280px;
  }
  
  .sessions-panel {
    width: 360px;
  }
  
  .requests-panel {
    width: 420px;
  }
}
```

### Mobile Layout

```
┌─────────────────────────────────────────────────────────────────┐
│                         Header                                  │
├─────────────────────────────────────────────────────────────────┤
│ [Sessions] [Requests] [Details]                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│              Active Tab Content                                 │
│                                                                 │
│                                                                 │
│                                                                 │
│                                                                 │
│                                                                 │
│                                                                 │
│                                                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Mobile Navigation:**
```jsx
<div className="block md:hidden">
  <div className="border-b border-gray-200">
    <nav className="flex space-x-8 px-4" aria-label="Tabs">
      <button
        className={`py-2 px-1 border-b-2 font-medium text-sm ${
          activeTab === 'sessions'
            ? 'border-blue-500 text-blue-600'
            : 'border-transparent text-gray-500 hover:text-gray-700'
        }`}
        onClick={() => setActiveTab('sessions')}
      >
        Sessions
      </button>
      <button
        className={`py-2 px-1 border-b-2 font-medium text-sm ${
          activeTab === 'requests'
            ? 'border-blue-500 text-blue-600'
            : 'border-transparent text-gray-500 hover:text-gray-700'
        }`}
        onClick={() => setActiveTab('requests')}
      >
        Requests
      </button>
      <button
        className={`py-2 px-1 border-b-2 font-medium text-sm ${
          activeTab === 'details'
            ? 'border-blue-500 text-blue-600'
            : 'border-transparent text-gray-500 hover:text-gray-700'
        }`}
        onClick={() => setActiveTab('details')}
      >
        Details
      </button>
    </nav>
  </div>
</div>
```

### Tablet Layout

```
┌─────────────────────────────────────────────────────────────────┐
│                         Header                                  │
├─────────────────────────────────────────────────────────────────┤
│              │                                                  │
│   Sessions   │              Details                             │
│              │                                                  │
│              │                                                  │
│              │                                                  │
│              │                                                  │
│              │                                                  │
│              │                                                  │
│              │                                                  │
│              │                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Two-Panel Layout:**
```jsx
<div className="hidden md:flex lg:hidden">
  <div className="w-80 border-r border-gray-200">
    <SessionList />
  </div>
  <div className="flex-1">
    {selectedRequest ? (
      <RequestDetail />
    ) : (
      <RequestList />
    )}
  </div>
</div>
```

## Interaction Patterns

### 1. Navigation Flow

```
Sessions → Select Session → View Requests → Select Request → View Details
    ↓                          ↓                               ↓
Filter/Search            Filter by Provider/Status      Expand/Collapse Messages
```

### 2. Selection States

**Visual Feedback:**
- **Hover**: Subtle background color change
- **Selected**: Border color change + background tint
- **Focus**: Keyboard focus ring
- **Active**: Pressed state with darker background

```jsx
const selectionStyles = {
  default: "border-gray-200 hover:bg-gray-50",
  selected: "border-blue-200 bg-blue-50",
  focus: "focus:ring-2 focus:ring-blue-500 focus:border-blue-500",
  active: "active:bg-gray-100"
}
```

### 3. Loading States

**Skeleton Components:**
```jsx
<div className="animate-pulse">
  <div className="h-4 bg-gray-300 rounded w-3/4 mb-2"></div>
  <div className="h-3 bg-gray-300 rounded w-1/2 mb-2"></div>
  <div className="h-3 bg-gray-300 rounded w-2/3"></div>
</div>
```

**Loading Indicators:**
```jsx
<div className="flex items-center justify-center h-64">
  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
</div>
```

### 4. Error States

**Error Boundaries:**
```jsx
<div className="text-center py-8">
  <div className="text-red-500 text-lg mb-2">⚠️</div>
  <div className="text-gray-900 font-medium mb-1">Something went wrong</div>
  <div className="text-gray-600 text-sm mb-4">{error.message}</div>
  <button
    onClick={retry}
    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
  >
    Try Again
  </button>
</div>
```

### 5. Empty States

**No Data:**
```jsx
<div className="text-center py-12">
  <div className="text-gray-400 text-6xl mb-4">📋</div>
  <div className="text-gray-900 font-medium mb-2">No sessions found</div>
  <div className="text-gray-600 text-sm">
    Start using the application to see HTTP request logs here
  </div>
</div>
```

## Visual States

### Component States

#### Session Cards
- **Default**: `border-gray-200 bg-white`
- **Hover**: `hover:bg-gray-50`
- **Selected**: `border-blue-200 bg-blue-50`
- **With Error**: Red warning icon visible
- **Loading**: Skeleton animation

#### Request Cards
- **Default**: `border-gray-200 bg-white`
- **Hover**: `hover:bg-gray-50`
- **Selected**: `border-blue-200 bg-blue-50`
- **Success**: Green checkmark icon
- **Error**: Red error icon + error message
- **Loading**: Skeleton animation

#### Message Viewers
- **Collapsed**: Header only with chevron right
- **Expanded**: Full content with chevron down
- **Copying**: Brief highlight animation
- **Error**: Red border and background

### Animation System

```css
/* Transitions */
.transition-all {
  transition: all 0.15s ease-in-out;
}

.transition-colors {
  transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out;
}

.transition-transform {
  transition: transform 0.15s ease-in-out;
}

/* Animations */
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.animate-fadeIn {
  animation: fadeIn 0.2s ease-out;
}
```

## Accessibility

### Keyboard Navigation

**Tab Order:**
1. Header actions (Settings, Help)
2. Sessions panel (Search → Filters → Session items)
3. Requests panel (Provider filter → Status filter → Request items)
4. Details panel (Message items → Copy buttons)

**Keyboard Shortcuts:**
- `Tab` / `Shift+Tab`: Navigate between focusable elements
- `Enter` / `Space`: Activate buttons and select items
- `Escape`: Close modals and collapse expanded sections
- `Arrow Keys`: Navigate within lists
- `Home` / `End`: Jump to first/last item in lists

### Screen Reader Support

**ARIA Labels:**
```jsx
<div role="main" aria-label="HTTP Request Log Viewer">
  <aside aria-label="Sessions" role="navigation">
    <input
      type="text"
      placeholder="Search sessions..."
      aria-label="Search sessions"
    />
    <div role="list" aria-label="Session list">
      <div role="listitem" aria-selected="false">
        <button aria-describedby="session-desc-1">
          session-abc123
        </button>
        <div id="session-desc-1" className="sr-only">
          5 LLM calls, 8 HTTP requests, 12 tool calls
        </div>
      </div>
    </div>
  </aside>
</div>
```

### Color Contrast

All color combinations meet WCAG AA standards:
- Text on light backgrounds: minimum 4.5:1 ratio
- Text on colored backgrounds: minimum 3:1 ratio
- Interactive elements: minimum 3:1 ratio for non-text elements

### Focus Management

```jsx
// Focus trap for modals
const focusableElements = modal.querySelectorAll(
  'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
)

const firstElement = focusableElements[0]
const lastElement = focusableElements[focusableElements.length - 1]

// Handle Tab key
if (event.key === 'Tab') {
  if (event.shiftKey) {
    if (document.activeElement === firstElement) {
      lastElement.focus()
      event.preventDefault()
    }
  } else {
    if (document.activeElement === lastElement) {
      firstElement.focus()
      event.preventDefault()
    }
  }
}
```

This comprehensive UI layout documentation provides detailed specifications for implementing a consistent, accessible, and user-friendly interface for the HTTP Request Log Viewer.