# HTTP Request Log Viewer - Implementation Guide

This document provides comprehensive implementation instructions for building the HTTP Request Log Viewer from the ground up.

## Table of Contents

1. [Project Setup](#project-setup)
2. [Database Integration](#database-integration)
3. [Type Definitions](#type-definitions)
4. [Core Components](#core-components)
5. [API Implementation](#api-implementation)
6. [Message Parsing](#message-parsing)
7. [UI Components](#ui-components)
8. [Testing Strategy](#testing-strategy)
9. [Deployment](#deployment)

## Project Setup

### 1. Initialize Next.js Project

```bash
cd log-viewer
npx create-next-app@latest . --typescript --tailwind --eslint --app
```

### 2. Install Dependencies

```bash
npm install @tanstack/react-query @tanstack/react-query-devtools
npm install clsx tailwind-merge
npm install lucide-react
npm install better-sqlite3
npm install @types/better-sqlite3
```

### 3. Project Structure

Create the following directory structure:

```
log-viewer/
├── src/
│   ├── app/
│   │   ├── globals.css
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   ├── sessions/
│   │   │   ├── page.tsx
│   │   │   └── [id]/
│   │   │       ├── page.tsx
│   │   │       └── requests/
│   │   │           └── [requestId]/
│   │   │               └── page.tsx
│   │   └── api/
│   │       ├── sessions/
│   │       │   ├── route.ts
│   │       │   └── [id]/
│   │       │       └── route.ts
│   │       └── requests/
│   │           └── [id]/
│   │               └── route.ts
│   ├── components/
│   │   ├── ui/
│   │   ├── sessions/
│   │   ├── requests/
│   │   └── layout/
│   ├── lib/
│   │   ├── database.ts
│   │   ├── parsers/
│   │   ├── types/
│   │   └── utils/
│   └── hooks/
├── public/
├── .env.local
├── next.config.js
├── tailwind.config.js
└── package.json
```

### 4. Configuration Files

#### `next.config.js`
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    serverComponentsExternalPackages: ['better-sqlite3']
  }
}

module.exports = nextConfig
```

#### `tailwind.config.js`
```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        system: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          900: '#1e3a8a',
        },
        user: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          900: '#14532d',
        },
        assistant: {
          50: '#f9fafb',
          100: '#f3f4f6',
          200: '#e5e7eb',
          900: '#111827',
        },
        tool: {
          50: '#faf5ff',
          100: '#f3e8ff',
          200: '#e9d5ff',
          900: '#581c87',
        },
        toolResponse: {
          50: '#fff7ed',
          100: '#ffedd5',
          200: '#fed7aa',
          900: '#9a3412',
        },
      }
    },
  },
  plugins: [],
}
```

#### `.env.local`
```env
DATA_DIR=../data
DATABASE_URL=sqlite:../data/sessions.db
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Database Integration

### 1. Database Connection (`src/lib/database.ts`)

```typescript
import Database from 'better-sqlite3'
import { readFileSync } from 'fs'
import { join } from 'path'

let db: Database.Database | null = null

export interface SessionMetadata {
  id: string
  session_id: string
  start_time: string
  end_time: string | null
  llm_call_count: number
  tool_call_count: number
  http_call_count: number
  total_tokens: number
  total_cost: number
  has_error: boolean
}

export interface SessionLog {
  id: string
  startTime: string
  endTime?: string
  metadata: Record<string, string>
  llmCalls: LLMCallLog[]
  toolCalls: ToolCallLog[]
  httpCalls: HTTPLog[]
  commandArgs: string[]
  userId?: string
}

export interface HTTPLog {
  id: string
  sessionId: string
  method: string
  url: string
  headers: Record<string, string[]>
  body: any
  statusCode?: number
  responseBody?: any
  responseHeaders?: Record<string, string[]>
  startTime: string
  endTime?: string
  durationMs: number
  error?: string
  parentToolCall?: string
}

export interface LLMCallLog {
  id: string
  sessionId: string
  provider: string
  model: string
  startTime: string
  endTime?: string
  request: any
  response?: any
  streamEvents?: StreamEvent[]
  error?: string
  tokensUsed?: TokenUsage
  cost?: number
  durationMs: number
  parentToolCall?: string
}

export interface ToolCallLog {
  id: string
  sessionId: string
  name: string
  startTime: string
  endTime?: string
  input: any
  output?: any
  error?: string
  durationMs: number
  parentId?: string
  childIds?: string[]
  parentLLMCall?: string
}

export interface StreamEvent {
  type: string
  data: any
  timestamp: string
}

export interface TokenUsage {
  prompt: number
  completion: number
  total: number
}

export function getDatabase(): Database.Database {
  if (!db) {
    const dbPath = process.env.DATABASE_URL?.replace('sqlite:', '') || './data/sessions.db'
    db = new Database(dbPath)
    
    // Enable foreign keys
    db.pragma('foreign_keys = ON')
  }
  
  return db
}

export function getSessionMetadata(filters: {
  startTime?: string
  endTime?: string
  hasError?: boolean
  limit?: number
} = {}): SessionMetadata[] {
  const db = getDatabase()
  
  let query = `
    SELECT 
      id, session_id, start_time, end_time,
      llm_call_count, tool_call_count, http_call_count,
      total_tokens, total_cost, has_error
    FROM sessions
    WHERE 1=1
  `
  
  const params: any[] = []
  
  if (filters.startTime) {
    query += ' AND start_time >= ?'
    params.push(filters.startTime)
  }
  
  if (filters.endTime) {
    query += ' AND start_time <= ?'
    params.push(filters.endTime)
  }
  
  if (filters.hasError !== undefined) {
    query += ' AND has_error = ?'
    params.push(filters.hasError ? 1 : 0)
  }
  
  query += ' ORDER BY start_time DESC'
  
  if (filters.limit) {
    query += ' LIMIT ?'
    params.push(filters.limit)
  }
  
  const stmt = db.prepare(query)
  return stmt.all(...params) as SessionMetadata[]
}

export function getSessionDetail(sessionId: string): SessionLog | null {
  const dataDir = process.env.DATA_DIR || './data'
  const filePath = join(dataDir, `${sessionId}.json`)
  
  try {
    const data = readFileSync(filePath, 'utf8')
    return JSON.parse(data) as SessionLog
  } catch (error) {
    console.error('Error reading session file:', error)
    return null
  }
}

export function getHTTPRequest(sessionId: string, requestId: string): HTTPLog | null {
  const session = getSessionDetail(sessionId)
  if (!session) return null
  
  return session.httpCalls.find(call => call.id === requestId) || null
}
```

## Type Definitions

### 1. Core Types (`src/lib/types/index.ts`)

```typescript
export interface MessagePart {
  type: 'system' | 'user' | 'assistant' | 'tool_call' | 'tool_response'
  content: string
  metadata?: {
    toolCallId?: string
    toolName?: string
    arguments?: any
    provider?: string
    model?: string
  }
  colorClass: string
  icon: string
}

export interface ParsedRequest {
  id: string
  method: string
  url: string
  provider: string
  model?: string
  timestamp: string
  statusCode?: number
  durationMs: number
  error?: string
  messages: MessagePart[]
  usage?: {
    promptTokens: number
    completionTokens: number
    totalTokens: number
  }
  cost?: number
}

export interface SessionSummary {
  id: string
  startTime: string
  endTime?: string
  requestCount: number
  errorCount: number
  totalTokens: number
  totalCost: number
  providers: string[]
  models: string[]
}

export interface ProviderConfig {
  name: string
  baseUrl: string
  authHeader: string
  requestParser: (request: any) => MessagePart[]
  responseParser: (response: any) => MessagePart[]
}
```

### 2. Message Colors (`src/lib/utils/colors.ts`)

```typescript
export const MESSAGE_COLORS = {
  system: {
    bg: 'bg-system-50',
    border: 'border-system-200',
    text: 'text-system-900',
    icon: '🤖'
  },
  user: {
    bg: 'bg-user-50',
    border: 'border-user-200',
    text: 'text-user-900',
    icon: '👤'
  },
  assistant: {
    bg: 'bg-assistant-50',
    border: 'border-assistant-200',
    text: 'text-assistant-900',
    icon: '🤖'
  },
  tool_call: {
    bg: 'bg-tool-50',
    border: 'border-tool-200',
    text: 'text-tool-900',
    icon: '🛠️'
  },
  tool_response: {
    bg: 'bg-toolResponse-50',
    border: 'border-toolResponse-200',
    text: 'text-toolResponse-900',
    icon: '⚡'
  },
  error: {
    bg: 'bg-red-50',
    border: 'border-red-200',
    text: 'text-red-900',
    icon: '❌'
  }
} as const

export type MessageType = keyof typeof MESSAGE_COLORS
```

## Message Parsing

### 1. GitHub Copilot Parser (`src/lib/parsers/copilot.ts`)

```typescript
import { MessagePart } from '../types'
import { MESSAGE_COLORS } from '../utils/colors'

export function parseCopilotRequest(requestBody: any): MessagePart[] {
  const parts: MessagePart[] = []
  
  if (!requestBody.messages) return parts
  
  for (const message of requestBody.messages) {
    switch (message.role) {
      case 'system':
        parts.push({
          type: 'system',
          content: typeof message.content === 'string' 
            ? message.content 
            : JSON.stringify(message.content, null, 2),
          colorClass: MESSAGE_COLORS.system.bg,
          icon: MESSAGE_COLORS.system.icon
        })
        break
        
      case 'user':
        let userContent = ''
        if (typeof message.content === 'string') {
          userContent = message.content
        } else if (Array.isArray(message.content)) {
          userContent = message.content
            .map((part: any) => part.type === 'text' ? part.text : `[${part.type}]`)
            .join('\n')
        }
        
        parts.push({
          type: 'user',
          content: userContent,
          colorClass: MESSAGE_COLORS.user.bg,
          icon: MESSAGE_COLORS.user.icon
        })
        break
        
      case 'assistant':
        if (message.content) {
          parts.push({
            type: 'assistant',
            content: message.content,
            colorClass: MESSAGE_COLORS.assistant.bg,
            icon: MESSAGE_COLORS.assistant.icon
          })
        }
        
        if (message.tool_calls) {
          for (const toolCall of message.tool_calls) {
            parts.push({
              type: 'tool_call',
              content: formatToolCall(toolCall),
              metadata: {
                toolCallId: toolCall.id,
                toolName: toolCall.function.name,
                arguments: JSON.parse(toolCall.function.arguments || '{}')
              },
              colorClass: MESSAGE_COLORS.tool_call.bg,
              icon: MESSAGE_COLORS.tool_call.icon
            })
          }
        }
        break
        
      case 'tool':
        parts.push({
          type: 'tool_response',
          content: message.content,
          metadata: {
            toolCallId: message.tool_call_id
          },
          colorClass: MESSAGE_COLORS.tool_response.bg,
          icon: MESSAGE_COLORS.tool_response.icon
        })
        break
    }
  }
  
  return parts
}

export function parseCopilotResponse(responseBody: any): MessagePart[] {
  const parts: MessagePart[] = []
  
  if (!responseBody.choices || !responseBody.choices[0]) return parts
  
  const choice = responseBody.choices[0]
  const message = choice.message
  
  if (message.content) {
    parts.push({
      type: 'assistant',
      content: message.content,
      colorClass: MESSAGE_COLORS.assistant.bg,
      icon: MESSAGE_COLORS.assistant.icon
    })
  }
  
  if (message.tool_calls) {
    for (const toolCall of message.tool_calls) {
      parts.push({
        type: 'tool_call',
        content: formatToolCall(toolCall),
        metadata: {
          toolCallId: toolCall.id,
          toolName: toolCall.function.name,
          arguments: JSON.parse(toolCall.function.arguments || '{}')
        },
        colorClass: MESSAGE_COLORS.tool_call.bg,
        icon: MESSAGE_COLORS.tool_call.icon
      })
    }
  }
  
  return parts
}

function formatToolCall(toolCall: any): string {
  return `Function: ${toolCall.function.name}\n\nArguments:\n${JSON.stringify(JSON.parse(toolCall.function.arguments || '{}'), null, 2)}`
}
```

### 2. Generic Message Parser (`src/lib/parsers/index.ts`)

```typescript
import { HTTPLog, MessagePart } from '../types'
import { parseCopilotRequest, parseCopilotResponse } from './copilot'

export function parseHTTPRequest(httpLog: HTTPLog): MessagePart[] {
  const parts: MessagePart[] = []
  
  // Determine provider from URL
  const provider = getProviderFromUrl(httpLog.url)
  
  // Parse request
  let requestParts: MessagePart[] = []
  switch (provider) {
    case 'copilot':
      requestParts = parseCopilotRequest(httpLog.body)
      break
    case 'openai':
      requestParts = parseOpenAIRequest(httpLog.body)
      break
    case 'anthropic':
      requestParts = parseAnthropicRequest(httpLog.body)
      break
    default:
      requestParts = parseGenericRequest(httpLog.body)
  }
  
  parts.push(...requestParts)
  
  // Parse response if available
  if (httpLog.responseBody && httpLog.statusCode === 200) {
    let responseParts: MessagePart[] = []
    switch (provider) {
      case 'copilot':
        responseParts = parseCopilotResponse(httpLog.responseBody)
        break
      case 'openai':
        responseParts = parseOpenAIResponse(httpLog.responseBody)
        break
      case 'anthropic':
        responseParts = parseAnthropicResponse(httpLog.responseBody)
        break
      default:
        responseParts = parseGenericResponse(httpLog.responseBody)
    }
    
    parts.push(...responseParts)
  }
  
  // Add error if present
  if (httpLog.error) {
    parts.push({
      type: 'error',
      content: httpLog.error,
      colorClass: 'bg-red-50',
      icon: '❌'
    })
  }
  
  return parts
}

function getProviderFromUrl(url: string): string {
  if (url.includes('githubcopilot.com')) return 'copilot'
  if (url.includes('openai.com')) return 'openai'
  if (url.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}

function parseOpenAIRequest(body: any): MessagePart[] {
  // Similar to Copilot parser but for OpenAI format
  return parseCopilotRequest(body) // Same format
}

function parseOpenAIResponse(body: any): MessagePart[] {
  return parseCopilotResponse(body) // Same format
}

function parseAnthropicRequest(body: any): MessagePart[] {
  // Anthropic-specific parsing logic
  const parts: MessagePart[] = []
  
  if (body.system) {
    parts.push({
      type: 'system',
      content: body.system,
      colorClass: 'bg-system-50',
      icon: '🤖'
    })
  }
  
  if (body.messages) {
    for (const message of body.messages) {
      if (message.role === 'user') {
        parts.push({
          type: 'user',
          content: Array.isArray(message.content) 
            ? message.content.map((c: any) => c.text || c.type).join('\n')
            : message.content,
          colorClass: 'bg-user-50',
          icon: '👤'
        })
      }
    }
  }
  
  return parts
}

function parseAnthropicResponse(body: any): MessagePart[] {
  const parts: MessagePart[] = []
  
  if (body.content) {
    for (const content of body.content) {
      if (content.type === 'text') {
        parts.push({
          type: 'assistant',
          content: content.text,
          colorClass: 'bg-assistant-50',
          icon: '🤖'
        })
      } else if (content.type === 'tool_use') {
        parts.push({
          type: 'tool_call',
          content: `Function: ${content.name}\n\nArguments:\n${JSON.stringify(content.input, null, 2)}`,
          metadata: {
            toolCallId: content.id,
            toolName: content.name,
            arguments: content.input
          },
          colorClass: 'bg-tool-50',
          icon: '🛠️'
        })
      }
    }
  }
  
  return parts
}

function parseGenericRequest(body: any): MessagePart[] {
  return [{
    type: 'system',
    content: JSON.stringify(body, null, 2),
    colorClass: 'bg-gray-50',
    icon: '📄'
  }]
}

function parseGenericResponse(body: any): MessagePart[] {
  return [{
    type: 'assistant',
    content: JSON.stringify(body, null, 2),
    colorClass: 'bg-gray-50',
    icon: '📄'
  }]
}
```

## API Implementation

### 1. Sessions API (`src/app/api/sessions/route.ts`)

```typescript
import { NextResponse } from 'next/server'
import { getSessionMetadata } from '@/lib/database'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    
    const filters = {
      startTime: searchParams.get('startTime') || undefined,
      endTime: searchParams.get('endTime') || undefined,
      hasError: searchParams.get('hasError') ? 
        searchParams.get('hasError') === 'true' : undefined,
      limit: searchParams.get('limit') ? 
        parseInt(searchParams.get('limit')!) : 100
    }
    
    const sessions = getSessionMetadata(filters)
    
    return NextResponse.json(sessions)
  } catch (error) {
    console.error('Error fetching sessions:', error)
    return NextResponse.json(
      { error: 'Failed to fetch sessions' },
      { status: 500 }
    )
  }
}
```

### 2. Session Detail API (`src/app/api/sessions/[id]/route.ts`)

```typescript
import { NextResponse } from 'next/server'
import { getSessionDetail } from '@/lib/database'

export async function GET(
  request: Request,
  { params }: { params: { id: string } }
) {
  try {
    const session = getSessionDetail(params.id)
    
    if (!session) {
      return NextResponse.json(
        { error: 'Session not found' },
        { status: 404 }
      )
    }
    
    return NextResponse.json(session)
  } catch (error) {
    console.error('Error fetching session:', error)
    return NextResponse.json(
      { error: 'Failed to fetch session' },
      { status: 500 }
    )
  }
}
```

### 3. Request Detail API (`src/app/api/requests/[id]/route.ts`)

```typescript
import { NextResponse } from 'next/server'
import { getHTTPRequest } from '@/lib/database'
import { parseHTTPRequest } from '@/lib/parsers'

export async function GET(
  request: Request,
  { params }: { params: { id: string } }
) {
  try {
    const { searchParams } = new URL(request.url)
    const sessionId = searchParams.get('sessionId')
    
    if (!sessionId) {
      return NextResponse.json(
        { error: 'Session ID required' },
        { status: 400 }
      )
    }
    
    const httpRequest = getHTTPRequest(sessionId, params.id)
    
    if (!httpRequest) {
      return NextResponse.json(
        { error: 'Request not found' },
        { status: 404 }
      )
    }
    
    // Parse the request into human-readable messages
    const messages = parseHTTPRequest(httpRequest)
    
    return NextResponse.json({
      ...httpRequest,
      messages
    })
  } catch (error) {
    console.error('Error fetching request:', error)
    return NextResponse.json(
      { error: 'Failed to fetch request' },
      { status: 500 }
    )
  }
}
```

## Core Components

### 1. Layout Component (`src/components/layout/Layout.tsx`)

```typescript
'use client'

import { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      cacheTime: 10 * 60 * 1000, // 10 minutes
    },
  },
})

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white shadow-sm border-b">
          <div className="max-w-7xl mx-auto px-4 py-4">
            <h1 className="text-2xl font-bold text-gray-900">
              HTTP Request Log Viewer
            </h1>
          </div>
        </header>
        
        <main className="max-w-7xl mx-auto">
          {children}
        </main>
      </div>
      
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  )
}
```

### 2. Session List Component (`src/components/sessions/SessionList.tsx`)

```typescript
'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { SessionCard } from './SessionCard'
import { SearchIcon, FilterIcon } from 'lucide-react'

interface SessionListProps {
  onSessionSelect?: (sessionId: string) => void
  selectedSessionId?: string
}

export function SessionList({ onSessionSelect, selectedSessionId }: SessionListProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [filterError, setFilterError] = useState<boolean | undefined>(undefined)
  
  const { data: sessions, isLoading, error } = useQuery({
    queryKey: ['sessions', { hasError: filterError }],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filterError !== undefined) {
        params.append('hasError', filterError.toString())
      }
      
      const response = await fetch(`/api/sessions?${params}`)
      if (!response.ok) {
        throw new Error('Failed to fetch sessions')
      }
      return response.json()
    }
  })
  
  const filteredSessions = sessions?.filter((session: any) =>
    session.id.toLowerCase().includes(searchTerm.toLowerCase())
  ) || []
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }
  
  if (error) {
    return (
      <div className="text-center py-8 text-red-600">
        Error loading sessions: {error.message}
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b bg-white">
        <h2 className="text-xl font-semibold mb-3">Sessions</h2>
        
        <div className="relative mb-3">
          <SearchIcon className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search sessions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
        
        <div className="flex gap-2">
          <select
            value={filterError?.toString() || ''}
            onChange={(e) => setFilterError(
              e.target.value === '' ? undefined : e.target.value === 'true'
            )}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="">All Sessions</option>
            <option value="false">Success Only</option>
            <option value="true">With Errors</option>
          </select>
        </div>
      </div>
      
      <div className="flex-1 overflow-y-auto">
        {filteredSessions.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No sessions found
          </div>
        ) : (
          filteredSessions.map((session: any) => (
            <SessionCard
              key={session.id}
              session={session}
              isSelected={session.id === selectedSessionId}
              onClick={() => onSessionSelect?.(session.id)}
            />
          ))
        )}
      </div>
    </div>
  )
}
```

### 3. Session Card Component (`src/components/sessions/SessionCard.tsx`)

```typescript
'use client'

import { SessionMetadata } from '@/lib/database'
import { Clock, AlertCircle, Activity } from 'lucide-react'

interface SessionCardProps {
  session: SessionMetadata
  isSelected?: boolean
  onClick?: () => void
}

export function SessionCard({ session, isSelected, onClick }: SessionCardProps) {
  const startTime = new Date(session.start_time).toLocaleString()
  const endTime = session.end_time ? new Date(session.end_time).toLocaleString() : 'Running'
  
  return (
    <div
      onClick={onClick}
      className={`p-4 border-b cursor-pointer transition-colors ${
        isSelected 
          ? 'bg-blue-50 border-blue-200' 
          : 'hover:bg-gray-50 border-gray-200'
      }`}
    >
      <div className="flex justify-between items-start mb-2">
        <div className="flex-1">
          <h3 className="font-medium text-sm text-gray-900 truncate">
            {session.session_id}
          </h3>
          <div className="flex items-center gap-2 mt-1 text-xs text-gray-600">
            <Clock className="h-3 w-3" />
            <span>{startTime}</span>
          </div>
        </div>
        
        {session.has_error && (
          <AlertCircle className="h-4 w-4 text-red-500 flex-shrink-0" />
        )}
      </div>
      
      <div className="flex gap-2 text-xs">
        <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded">
          {session.llm_call_count} LLM
        </span>
        <span className="bg-green-100 text-green-800 px-2 py-1 rounded">
          {session.http_call_count} HTTP
        </span>
        <span className="bg-purple-100 text-purple-800 px-2 py-1 rounded">
          {session.tool_call_count} Tools
        </span>
      </div>
      
      <div className="flex justify-between items-center mt-2 text-xs text-gray-600">
        <span>{session.total_tokens.toLocaleString()} tokens</span>
        <span>${session.total_cost.toFixed(4)}</span>
      </div>
    </div>
  )
}
```

### 4. Request List Component (`src/components/requests/RequestList.tsx`)

```typescript
'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { RequestCard } from './RequestCard'
import { Filter } from 'lucide-react'

interface RequestListProps {
  sessionId: string
  onRequestSelect?: (requestId: string) => void
  selectedRequestId?: string
}

export function RequestList({ sessionId, onRequestSelect, selectedRequestId }: RequestListProps) {
  const [providerFilter, setProviderFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState('all')
  
  const { data: session, isLoading } = useQuery({
    queryKey: ['session', sessionId],
    queryFn: async () => {
      const response = await fetch(`/api/sessions/${sessionId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch session')
      }
      return response.json()
    },
    enabled: !!sessionId
  })
  
  const requests = session?.httpCalls || []
  
  const filteredRequests = requests.filter((request: any) => {
    const provider = getProviderFromUrl(request.url)
    const hasError = request.error || request.statusCode >= 400
    
    if (providerFilter !== 'all' && provider !== providerFilter) {
      return false
    }
    
    if (statusFilter === 'success' && hasError) {
      return false
    }
    
    if (statusFilter === 'error' && !hasError) {
      return false
    }
    
    return true
  })
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b bg-white">
        <h3 className="text-lg font-semibold mb-2">HTTP Requests</h3>
        
        <div className="flex gap-2">
          <select
            value={providerFilter}
            onChange={(e) => setProviderFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
          >
            <option value="all">All Providers</option>
            <option value="copilot">GitHub Copilot</option>
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
          </select>
          
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
          >
            <option value="all">All Status</option>
            <option value="success">Success</option>
            <option value="error">Error</option>
          </select>
        </div>
      </div>
      
      <div className="flex-1 overflow-y-auto">
        {filteredRequests.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No requests found
          </div>
        ) : (
          filteredRequests.map((request: any) => (
            <RequestCard
              key={request.id}
              request={request}
              isSelected={request.id === selectedRequestId}
              onClick={() => onRequestSelect?.(request.id)}
            />
          ))
        )}
      </div>
    </div>
  )
}

function getProviderFromUrl(url: string): string {
  if (url.includes('githubcopilot.com')) return 'copilot'
  if (url.includes('openai.com')) return 'openai'
  if (url.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}
```

### 5. Request Card Component (`src/components/requests/RequestCard.tsx`)

```typescript
'use client'

import { HTTPLog } from '@/lib/database'
import { Clock, AlertCircle, CheckCircle } from 'lucide-react'

interface RequestCardProps {
  request: HTTPLog
  isSelected?: boolean
  onClick?: () => void
}

export function RequestCard({ request, isSelected, onClick }: RequestCardProps) {
  const isSuccess = request.statusCode && request.statusCode >= 200 && request.statusCode < 300
  const hasError = request.error || (request.statusCode && request.statusCode >= 400)
  
  const provider = getProviderFromUrl(request.url)
  const providerName = getProviderName(provider)
  
  return (
    <div
      onClick={onClick}
      className={`p-4 border-b cursor-pointer transition-colors ${
        isSelected 
          ? 'bg-blue-50 border-blue-200' 
          : 'hover:bg-gray-50 border-gray-200'
      }`}
    >
      <div className="flex justify-between items-start mb-2">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
              {request.method}
            </span>
            <span className="text-sm text-gray-600 truncate">
              {providerName}
            </span>
          </div>
          
          <div className="flex items-center gap-2 text-xs text-gray-600">
            <Clock className="h-3 w-3" />
            <span>{new Date(request.startTime).toLocaleString()}</span>
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          {request.statusCode && (
            <span className={`px-2 py-1 rounded text-xs font-medium ${
              isSuccess
                ? 'bg-green-100 text-green-800'
                : hasError
                ? 'bg-red-100 text-red-800'
                : 'bg-yellow-100 text-yellow-800'
            }`}>
              {request.statusCode}
            </span>
          )}
          
          <span className="text-xs text-gray-500">
            {request.durationMs}ms
          </span>
          
          {hasError && <AlertCircle className="h-4 w-4 text-red-500" />}
          {isSuccess && <CheckCircle className="h-4 w-4 text-green-500" />}
        </div>
      </div>
      
      {request.error && (
        <div className="text-xs text-red-600 mt-1 p-2 bg-red-50 rounded">
          {request.error}
        </div>
      )}
    </div>
  )
}

function getProviderFromUrl(url: string): string {
  if (url.includes('githubcopilot.com')) return 'copilot'
  if (url.includes('openai.com')) return 'openai'
  if (url.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}

function getProviderName(provider: string): string {
  switch (provider) {
    case 'copilot': return 'GitHub Copilot'
    case 'openai': return 'OpenAI'
    case 'anthropic': return 'Anthropic'
    default: return 'Unknown Provider'
  }
}
```

## UI Components

### 1. Collapsible Message Viewer (`src/components/ui/CollapsibleMessageViewer.tsx`)

```typescript
'use client'

import { useState } from 'react'
import { ChevronRight, Copy } from 'lucide-react'
import { MessagePart } from '@/lib/types'

interface CollapsibleMessageViewerProps {
  message: MessagePart
  defaultOpen?: boolean
}

export function CollapsibleMessageViewer({ 
  message, 
  defaultOpen = false 
}: CollapsibleMessageViewerProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)
  
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(message.content)
      // TODO: Show toast notification
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }
  
  return (
    <div className={`border rounded-lg mb-3 ${message.colorClass} border-gray-200`}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full p-3 text-left flex justify-between items-center hover:bg-opacity-80 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-lg">{message.icon}</span>
          <span className="font-medium text-gray-900 capitalize">
            {message.type.replace('_', ' ')}
          </span>
          {message.metadata?.toolName && (
            <span className="text-sm text-gray-600">
              ({message.metadata.toolName})
            </span>
          )}
        </div>
        
        <div className="flex items-center gap-2">
          <button
            onClick={(e) => {
              e.stopPropagation()
              handleCopy()
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
        <div className="p-3 pt-0">
          <ContentRenderer message={message} />
        </div>
      )}
    </div>
  )
}

function ContentRenderer({ message }: { message: MessagePart }) {
  if (message.type === 'tool_call' && message.metadata?.arguments) {
    return (
      <div>
        <div className="font-mono text-sm mb-2 text-gray-700">
          Function: {message.metadata.toolName}
        </div>
        <pre className="bg-gray-100 p-3 rounded text-xs overflow-x-auto">
          {JSON.stringify(message.metadata.arguments, null, 2)}
        </pre>
      </div>
    )
  }
  
  if (message.type === 'system' || message.type === 'user') {
    return (
      <div className="whitespace-pre-wrap text-sm leading-relaxed">
        {message.content}
      </div>
    )
  }
  
  return (
    <div className="prose prose-sm max-w-none">
      <div className="whitespace-pre-wrap text-sm leading-relaxed">
        {message.content}
      </div>
    </div>
  )
}
```

### 2. Request Detail View (`src/components/requests/RequestDetail.tsx`)

```typescript
'use client'

import { useQuery } from '@tanstack/react-query'
import { CollapsibleMessageViewer } from '../ui/CollapsibleMessageViewer'
import { Clock, Server, Zap } from 'lucide-react'

interface RequestDetailProps {
  sessionId: string
  requestId: string
}

export function RequestDetail({ sessionId, requestId }: RequestDetailProps) {
  const { data: request, isLoading } = useQuery({
    queryKey: ['request', requestId, sessionId],
    queryFn: async () => {
      const response = await fetch(`/api/requests/${requestId}?sessionId=${sessionId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch request')
      }
      return response.json()
    },
    enabled: !!sessionId && !!requestId
  })
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }
  
  if (!request) {
    return (
      <div className="text-center py-8 text-gray-500">
        Select a request to view details
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b bg-white">
        <div className="flex justify-between items-start mb-2">
          <h3 className="text-lg font-semibold">Request Details</h3>
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <Clock className="h-4 w-4" />
            <span>{request.durationMs}ms</span>
          </div>
        </div>
        
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <div className="flex items-center gap-1">
            <Server className="h-4 w-4" />
            <span className="font-mono">{request.method}</span>
          </div>
          
          <div className="flex items-center gap-1">
            <Zap className="h-4 w-4" />
            <span className={`px-2 py-1 rounded text-xs ${
              request.statusCode >= 200 && request.statusCode < 300
                ? 'bg-green-100 text-green-800'
                : 'bg-red-100 text-red-800'
            }`}>
              {request.statusCode}
            </span>
          </div>
          
          <span className="text-xs text-gray-500">
            {new Date(request.startTime).toLocaleString()}
          </span>
        </div>
        
        <div className="mt-2 text-xs text-gray-600">
          {request.url}
        </div>
      </div>
      
      <div className="flex-1 overflow-y-auto p-4">
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Messages</h4>
            <div className="space-y-2">
              {request.messages?.map((message: any, index: number) => (
                <CollapsibleMessageViewer
                  key={index}
                  message={message}
                  defaultOpen={index === 0}
                />
              ))}
            </div>
          </div>
          
          {request.error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <h4 className="font-medium text-red-900 mb-1">Error</h4>
              <p className="text-sm text-red-700">{request.error}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
```

### 3. Main Page Layout (`src/app/page.tsx`)

```typescript
'use client'

import { useState } from 'react'
import { Layout } from '@/components/layout/Layout'
import { SessionList } from '@/components/sessions/SessionList'
import { RequestList } from '@/components/requests/RequestList'
import { RequestDetail } from '@/components/requests/RequestDetail'

export default function HomePage() {
  const [selectedSessionId, setSelectedSessionId] = useState<string>('')
  const [selectedRequestId, setSelectedRequestId] = useState<string>('')
  
  return (
    <Layout>
      <div className="h-screen flex">
        {/* Sessions Sidebar */}
        <div className="w-80 border-r bg-white">
          <SessionList 
            onSessionSelect={setSelectedSessionId}
            selectedSessionId={selectedSessionId}
          />
        </div>
        
        {/* Main Content */}
        <div className="flex-1 flex">
          {/* Requests List */}
          <div className="w-96 border-r bg-white">
            {selectedSessionId ? (
              <RequestList 
                sessionId={selectedSessionId}
                onRequestSelect={setSelectedRequestId}
                selectedRequestId={selectedRequestId}
              />
            ) : (
              <div className="flex items-center justify-center h-full text-gray-500">
                Select a session to view requests
              </div>
            )}
          </div>
          
          {/* Request Detail */}
          <div className="flex-1 bg-gray-50">
            {selectedSessionId && selectedRequestId ? (
              <RequestDetail 
                sessionId={selectedSessionId}
                requestId={selectedRequestId}
              />
            ) : (
              <div className="flex items-center justify-center h-full text-gray-500">
                Select a request to view details
              </div>
            )}
          </div>
        </div>
      </div>
    </Layout>
  )
}
```

## Testing Strategy

### 1. Unit Tests (`src/lib/parsers/__tests__/copilot.test.ts`)

```typescript
import { parseCopilotRequest, parseCopilotResponse } from '../copilot'

describe('Copilot Parser', () => {
  describe('parseCopilotRequest', () => {
    it('should parse system message', () => {
      const request = {
        messages: [
          {
            role: 'system',
            content: 'You are a helpful assistant.'
          }
        ]
      }
      
      const result = parseCopilotRequest(request)
      
      expect(result).toHaveLength(1)
      expect(result[0]).toEqual({
        type: 'system',
        content: 'You are a helpful assistant.',
        colorClass: 'bg-system-50',
        icon: '🤖'
      })
    })
    
    it('should parse user message with text content', () => {
      const request = {
        messages: [
          {
            role: 'user',
            content: 'Hello, how are you?'
          }
        ]
      }
      
      const result = parseCopilotRequest(request)
      
      expect(result).toHaveLength(1)
      expect(result[0]).toEqual({
        type: 'user',
        content: 'Hello, how are you?',
        colorClass: 'bg-user-50',
        icon: '👤'
      })
    })
    
    it('should parse tool calls', () => {
      const request = {
        messages: [
          {
            role: 'assistant',
            tool_calls: [
              {
                id: 'call_123',
                type: 'function',
                function: {
                  name: 'get_weather',
                  arguments: '{"location": "San Francisco"}'
                }
              }
            ]
          }
        ]
      }
      
      const result = parseCopilotRequest(request)
      
      expect(result).toHaveLength(1)
      expect(result[0]).toEqual({
        type: 'tool_call',
        content: 'Function: get_weather\n\nArguments:\n{\n  "location": "San Francisco"\n}',
        metadata: {
          toolCallId: 'call_123',
          toolName: 'get_weather',
          arguments: { location: 'San Francisco' }
        },
        colorClass: 'bg-tool-50',
        icon: '🛠️'
      })
    })
  })
})
```

### 2. Integration Tests (`src/components/__tests__/RequestDetail.test.tsx`)

```typescript
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RequestDetail } from '../requests/RequestDetail'

// Mock fetch
global.fetch = jest.fn()

const mockRequest = {
  id: 'req-123',
  method: 'POST',
  url: 'https://api.githubcopilot.com/chat/completions',
  statusCode: 200,
  durationMs: 1500,
  startTime: '2024-01-01T00:00:00Z',
  messages: [
    {
      type: 'system',
      content: 'You are a helpful assistant.',
      colorClass: 'bg-system-50',
      icon: '🤖'
    },
    {
      type: 'user',
      content: 'Hello!',
      colorClass: 'bg-user-50',
      icon: '👤'
    }
  ]
}

function renderWithQueryClient(component: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  
  return render(
    <QueryClientProvider client={queryClient}>
      {component}
    </QueryClientProvider>
  )
}

describe('RequestDetail', () => {
  beforeEach(() => {
    ;(fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockRequest,
    })
  })
  
  afterEach(() => {
    jest.clearAllMocks()
  })
  
  it('renders request details correctly', async () => {
    renderWithQueryClient(
      <RequestDetail sessionId="session-123" requestId="req-123" />
    )
    
    await waitFor(() => {
      expect(screen.getByText('Request Details')).toBeInTheDocument()
      expect(screen.getByText('POST')).toBeInTheDocument()
      expect(screen.getByText('200')).toBeInTheDocument()
      expect(screen.getByText('1500ms')).toBeInTheDocument()
    })
  })
  
  it('renders messages correctly', async () => {
    renderWithQueryClient(
      <RequestDetail sessionId="session-123" requestId="req-123" />
    )
    
    await waitFor(() => {
      expect(screen.getByText('System')).toBeInTheDocument()
      expect(screen.getByText('User')).toBeInTheDocument()
    })
  })
})
```

## Deployment

### 1. Production Build

```bash
npm run build
```

### 2. Environment Variables

```env
# Production environment
NODE_ENV=production
DATA_DIR=/var/lib/superopencode/data
DATABASE_URL=sqlite:///var/lib/superopencode/data/sessions.db
NEXT_PUBLIC_APP_URL=https://your-domain.com
```

### 3. Docker Configuration (`Dockerfile`)

```dockerfile
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000

CMD ["node", "server.js"]
```

### 4. Docker Compose (`docker-compose.yml`)

```yaml
version: '3.8'

services:
  log-viewer:
    build: .
    ports:
      - "3000:3000"
    volumes:
      - ./data:/var/lib/superopencode/data:ro
    environment:
      - NODE_ENV=production
      - DATA_DIR=/var/lib/superopencode/data
      - DATABASE_URL=sqlite:///var/lib/superopencode/data/sessions.db
      - NEXT_PUBLIC_APP_URL=http://localhost:3000
    restart: unless-stopped
```

## Implementation Checklist

### Phase 1: Core Setup
- [ ] Initialize Next.js project with TypeScript and Tailwind
- [ ] Set up database connection and types
- [ ] Create basic API routes for sessions and requests
- [ ] Implement core message parsing logic

### Phase 2: Basic UI
- [ ] Create Layout component with React Query
- [ ] Build SessionList and SessionCard components
- [ ] Implement RequestList and RequestCard components
- [ ] Add basic routing and navigation

### Phase 3: Message Parsing
- [ ] Implement Copilot request/response parsing
- [ ] Add OpenAI and Anthropic parsers
- [ ] Create CollapsibleMessageViewer component
- [ ] Add RequestDetail view with message display

### Phase 4: Advanced Features
- [ ] Add search and filtering functionality
- [ ] Implement error handling and loading states
- [ ] Add copy-to-clipboard functionality
- [ ] Create responsive design

### Phase 5: Testing & Deployment
- [ ] Write unit tests for parsers
- [ ] Add integration tests for components
- [ ] Set up production build and deployment
- [ ] Create documentation and user guide

This implementation guide provides a complete roadmap for building the HTTP Request Log Viewer. Each section includes working code examples and detailed explanations to help developers implement the feature successfully.