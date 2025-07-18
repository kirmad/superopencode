export interface MessagePart {
  type: 'system' | 'user' | 'assistant' | 'tool_call' | 'tool_response' | 'error'
  content: string
  metadata?: {
    toolCallId?: string
    toolName?: string
    arguments?: Record<string, unknown>
    provider?: string
    model?: string
    streamEvents?: number
    hasToolCalls?: boolean
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
  requestParser: (request: Record<string, unknown>) => MessagePart[]
  responseParser: (response: Record<string, unknown>) => MessagePart[]
}

export type MessageType = 'system' | 'user' | 'assistant' | 'tool_call' | 'tool_response' | 'error'