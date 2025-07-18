export interface SSEEvent {
  id?: string
  event?: string
  data: string
  retry?: number
  timestamp?: number
  parsed?: any
}

export interface ParsedStreamResponse {
  events: SSEEvent[]
  totalEvents: number
  hasToolCalls: boolean
  reconstructedMessage?: any
  metadata: {
    firstEventTime?: number
    lastEventTime?: number
    streamDuration?: number
    model?: string
    finishReason?: string
    usage?: any
  }
}

/**
 * Parse Server-Sent Events (SSE) format text into individual events
 */
export function parseSSEStream(sseText: string): ParsedStreamResponse {
  if (!sseText || typeof sseText !== 'string') {
    return {
      events: [],
      totalEvents: 0,
      hasToolCalls: false,
      metadata: {}
    }
  }

  const events: SSEEvent[] = []
  const lines = sseText.split('\n')
  let currentEvent: Partial<SSEEvent> = {}
  let eventIndex = 0

  for (const line of lines) {
    const trimmed = line.trim()
    
    // Skip empty lines - they end an event
    if (!trimmed) {
      if (currentEvent.data !== undefined) {
        // Process the completed event
        const event: SSEEvent = {
          ...currentEvent as SSEEvent,
          timestamp: eventIndex++
        }
        
        // Try to parse JSON data
        if (event.data && event.data !== '[DONE]') {
          try {
            event.parsed = JSON.parse(event.data)
          } catch {
            // Not JSON, keep as string
          }
        }
        
        events.push(event)
        currentEvent = {}
      }
      continue
    }

    // Parse SSE field
    const colonIndex = trimmed.indexOf(':')
    if (colonIndex === -1) continue

    const field = trimmed.substring(0, colonIndex).trim()
    const value = trimmed.substring(colonIndex + 1).trim()

    switch (field) {
      case 'data':
        currentEvent.data = value
        break
      case 'event':
        currentEvent.event = value
        break
      case 'id':
        currentEvent.id = value
        break
      case 'retry':
        currentEvent.retry = parseInt(value, 10)
        break
    }
  }

  // Process final event if exists
  if (currentEvent.data !== undefined) {
    const event: SSEEvent = {
      ...currentEvent as SSEEvent,
      timestamp: eventIndex++
    }
    
    if (event.data && event.data !== '[DONE]') {
      try {
        event.parsed = JSON.parse(event.data)
      } catch {
        // Not JSON, keep as string
      }
    }
    
    events.push(event)
  }

  // Analyze the events to extract metadata and reconstruct messages
  const analysis = analyzeStreamEvents(events)

  return {
    events,
    totalEvents: events.length,
    hasToolCalls: analysis.hasToolCalls,
    reconstructedMessage: analysis.reconstructedMessage,
    metadata: analysis.metadata
  }
}

/**
 * Analyze parsed events to extract metadata and reconstruct streaming messages
 */
function analyzeStreamEvents(events: SSEEvent[]) {
  let hasToolCalls = false
  let model: string | undefined
  let finishReason: string | undefined
  let usage: any
  let firstEventTime: number | undefined
  let lastEventTime: number | undefined

  // For tool call reconstruction
  const toolCalls: Record<string, any> = {}
  let content = ''
  let role = ''

  for (const event of events) {
    if (!event.parsed) continue

    const data = event.parsed

    // Extract metadata
    if (data.model) model = data.model
    if (data.usage) usage = data.usage
    if (data.created && !firstEventTime) firstEventTime = data.created
    if (data.created) lastEventTime = data.created

    // Process choices for message reconstruction
    if (data.choices && Array.isArray(data.choices)) {
      for (const choice of data.choices) {
        if (choice.finish_reason) {
          finishReason = choice.finish_reason
        }

        if (choice.delta) {
          const delta = choice.delta

          // Handle role
          if (delta.role) {
            role = delta.role
          }

          // Handle content
          if (delta.content) {
            content += delta.content
          }

          // Handle tool calls
          if (delta.tool_calls && Array.isArray(delta.tool_calls)) {
            hasToolCalls = true
            
            for (const toolCall of delta.tool_calls) {
              const id = toolCall.id || `tool_${toolCall.index || 0}`
              
              if (!toolCalls[id]) {
                toolCalls[id] = {
                  id: toolCall.id,
                  type: toolCall.type || 'function',
                  function: {
                    name: '',
                    arguments: ''
                  }
                }
              }

              if (toolCall.function) {
                if (toolCall.function.name) {
                  toolCalls[id].function.name = toolCall.function.name
                }
                if (toolCall.function.arguments) {
                  toolCalls[id].function.arguments += toolCall.function.arguments
                }
              }
            }
          }
        }
      }
    }
  }

  // Build reconstructed message
  let reconstructedMessage: any = null

  if (role || content || hasToolCalls) {
    reconstructedMessage = {
      role: role || 'assistant',
      content: content || null
    }

    if (hasToolCalls && Object.keys(toolCalls).length > 0) {
      reconstructedMessage.tool_calls = Object.values(toolCalls).map(tc => {
        // Try to parse arguments as JSON
        try {
          return {
            ...tc,
            function: {
              ...tc.function,
              arguments: tc.function.arguments ? JSON.parse(tc.function.arguments) : {}
            }
          }
        } catch {
          return tc
        }
      })
    }
  }

  const streamDuration = firstEventTime && lastEventTime 
    ? (lastEventTime - firstEventTime) * 1000 
    : undefined

  return {
    hasToolCalls,
    reconstructedMessage,
    metadata: {
      firstEventTime,
      lastEventTime,
      streamDuration,
      model,
      finishReason,
      usage
    }
  }
}

/**
 * Check if response body appears to be SSE format
 */
export function isSSEResponse(responseBody: any): boolean {
  if (typeof responseBody === 'string') {
    return responseBody.includes('data:') && 
           (responseBody.includes('[DONE]') || responseBody.includes('event:'))
  }
  return false
}

/**
 * Get a summary of stream events for display
 */
export function getStreamSummary(parsedStream: ParsedStreamResponse): string {
  const { totalEvents, hasToolCalls, metadata } = parsedStream
  
  const parts = []
  
  parts.push(`${totalEvents} events`)
  
  if (metadata.model) {
    parts.push(`model: ${metadata.model}`)
  }
  
  if (hasToolCalls) {
    parts.push('tool calls')
  }
  
  if (metadata.usage?.total_tokens) {
    parts.push(`${metadata.usage.total_tokens} tokens`)
  }
  
  if (metadata.streamDuration && metadata.streamDuration > 0) {
    parts.push(`${metadata.streamDuration}ms`)
  }
  
  return parts.join(' • ')
}