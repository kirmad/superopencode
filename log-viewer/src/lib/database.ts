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

// Real data structure from superopencode (snake_case)
export interface RealSessionLog {
  id: string
  start_time: string
  end_time?: string
  metadata: Record<string, unknown>
  llm_calls: RealLLMCallLog[]
  tool_calls: RealToolCallLog[]
  http_calls: RealHTTPLog[]
  command_args?: string[]
}

export interface RealHTTPLog {
  id: string
  session_id: string
  method: string
  url: string
  headers: Record<string, string[]>
  body?: Record<string, unknown>
  status_code?: number
  response_body?: Record<string, unknown>
  response_headers?: Record<string, string[]>
  start_time: string
  end_time?: string
  duration_ms?: number
  error?: string
}

export interface RealLLMCallLog {
  id: string
  session_id: string
  provider: string
  model: string
  start_time: string
  end_time?: string
  request: Record<string, unknown>
  response?: Record<string, unknown>
  tokens_used?: {
    prompt: number
    completion: number
    total: number
  }
  duration_ms: number
}

export interface RealToolCallLog {
  id: string
  session_id: string
  name: string
  start_time: string
  end_time?: string
  input: Record<string, unknown>
  output?: Record<string, unknown>
  error?: string
  duration_ms: number
}

export interface HTTPLog {
  id: string
  sessionId: string
  method: string
  url: string
  headers: Record<string, string[]>
  body: Record<string, unknown>
  statusCode?: number
  responseBody?: Record<string, unknown>
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
  request: Record<string, unknown>
  response?: Record<string, unknown>
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
  input: Record<string, unknown>
  output?: Record<string, unknown>
  error?: string
  durationMs: number
  parentId?: string
  childIds?: string[]
  parentLLMCall?: string
}

export interface StreamEvent {
  type: string
  data: Record<string, unknown>
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
  
  const params: (string | number)[] = []
  
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
    const realData = JSON.parse(data) as RealSessionLog
    
    // Convert real data structure to expected format
    return convertRealSessionToSessionLog(realData)
  } catch (error) {
    console.error('Error reading session file:', error)
    return null
  }
}

function convertRealSessionToSessionLog(realSession: RealSessionLog): SessionLog {
  return {
    id: realSession.id,
    startTime: realSession.start_time,
    endTime: realSession.end_time,
    metadata: realSession.metadata as Record<string, string>,
    llmCalls: realSession.llm_calls.map(convertRealLLMCall),
    toolCalls: realSession.tool_calls.map(convertRealToolCall),
    httpCalls: realSession.http_calls.map(convertRealHTTPCall),
    commandArgs: realSession.command_args || []
  }
}

function convertRealHTTPCall(realCall: RealHTTPLog): HTTPLog {
  return {
    id: realCall.id,
    sessionId: realCall.session_id,
    method: realCall.method,
    url: realCall.url,
    headers: realCall.headers,
    body: realCall.body || {},
    statusCode: realCall.status_code,
    responseBody: realCall.response_body,
    responseHeaders: realCall.response_headers,
    startTime: realCall.start_time,
    endTime: realCall.end_time,
    durationMs: realCall.duration_ms || 0,
    error: realCall.error
  }
}

function convertRealLLMCall(realCall: RealLLMCallLog): LLMCallLog {
  return {
    id: realCall.id,
    sessionId: realCall.session_id,
    provider: realCall.provider,
    model: realCall.model,
    startTime: realCall.start_time,
    endTime: realCall.end_time,
    request: realCall.request,
    response: realCall.response,
    tokensUsed: realCall.tokens_used ? {
      prompt: realCall.tokens_used.prompt,
      completion: realCall.tokens_used.completion,
      total: realCall.tokens_used.total
    } : undefined,
    durationMs: realCall.duration_ms
  }
}

function convertRealToolCall(realCall: RealToolCallLog): ToolCallLog {
  return {
    id: realCall.id,
    sessionId: realCall.session_id,
    name: realCall.name,
    startTime: realCall.start_time,
    endTime: realCall.end_time,
    input: realCall.input,
    output: realCall.output,
    error: realCall.error,
    durationMs: realCall.duration_ms
  }
}

export function getHTTPRequest(sessionId: string, requestId: string): HTTPLog | null {
  const session = getSessionDetail(sessionId)
  if (!session) return null
  
  return session.httpCalls.find(call => call.id === requestId) || null
}