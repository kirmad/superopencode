import { MessagePart } from '../types'
import { HTTPLog } from '../database'
import { parseCopilotRequest, parseCopilotResponse } from './copilot'
import { MESSAGE_COLORS } from '../utils/colors'
import { parseSSEStream, isSSEResponse } from './sse'

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
      colorClass: MESSAGE_COLORS.error.bg,
      icon: MESSAGE_COLORS.error.icon
    })
  }
  
  return parts
}

function getProviderFromUrl(url: string): string {
  const urlLower = url.toLowerCase()
  if (urlLower.includes('githubcopilot.com') || urlLower.includes('github.com/copilot') || urlLower.includes('copilot')) return 'copilot'
  if (urlLower.includes('openai.com')) return 'openai'  
  if (urlLower.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}

function parseOpenAIRequest(body: Record<string, unknown>): MessagePart[] {
  // OpenAI uses same format as Copilot
  return parseCopilotRequest(body)
}

function parseOpenAIResponse(body: Record<string, unknown> | string): MessagePart[] {
  // Check if this is a streaming response
  if (isSSEResponse(body)) {
    return parseStreamingResponse(body as string, 'openai')
  }
  
  return parseCopilotResponse(body as Record<string, unknown>)
}

function parseAnthropicRequest(body: Record<string, unknown>): MessagePart[] {
  const parts: MessagePart[] = []
  
  if (body.system) {
    parts.push({
      type: 'system',
      content: body.system as string,
      colorClass: MESSAGE_COLORS.system.bg,
      icon: MESSAGE_COLORS.system.icon
    })
  }
  
  if (body.messages && Array.isArray(body.messages)) {
    for (const message of body.messages as Record<string, unknown>[]) {
      if (message.role === 'user') {
        parts.push({
          type: 'user',
          content: Array.isArray(message.content) 
            ? message.content.map((c: Record<string, unknown>) => c.text || c.type).join('\n')
            : message.content as string,
          colorClass: MESSAGE_COLORS.user.bg,
          icon: MESSAGE_COLORS.user.icon
        })
      }
    }
  }
  
  return parts
}

function parseAnthropicResponse(body: Record<string, unknown> | string): MessagePart[] {
  // Check if this is a streaming response
  if (isSSEResponse(body)) {
    return parseStreamingResponse(body as string, 'anthropic')
  }
  
  // Handle regular JSON response
  const jsonBody = body as Record<string, unknown>
  const parts: MessagePart[] = []
  
  if (jsonBody.content && Array.isArray(jsonBody.content)) {
    for (const content of jsonBody.content as Record<string, unknown>[]) {
      if (content.type === 'text') {
        parts.push({
          type: 'assistant',
          content: content.text as string,
          colorClass: MESSAGE_COLORS.assistant.bg,
          icon: MESSAGE_COLORS.assistant.icon
        })
      } else if (content.type === 'tool_use') {
        parts.push({
          type: 'tool_call',
          content: `Function: ${content.name}\n\nArguments:\n${JSON.stringify(content.input, null, 2)}`,
          metadata: {
            toolCallId: content.id as string,
            toolName: content.name as string,
            arguments: content.input as Record<string, unknown>
          },
          colorClass: MESSAGE_COLORS.tool_call.bg,
          icon: MESSAGE_COLORS.tool_call.icon
        })
      }
    }
  }
  
  return parts
}

function parseGenericRequest(body: Record<string, unknown>): MessagePart[] {
  return [{
    type: 'system',
    content: JSON.stringify(body, null, 2),
    colorClass: MESSAGE_COLORS.assistant.bg,
    icon: '📄'
  }]
}

function parseGenericResponse(body: Record<string, unknown>): MessagePart[] {
  return [{
    type: 'assistant',
    content: JSON.stringify(body, null, 2),
    colorClass: MESSAGE_COLORS.assistant.bg,
    icon: '📄'
  }]
}

function parseStreamingResponse(sseText: string, provider: string): MessagePart[] {
  const parsedStream = parseSSEStream(sseText)
  const parts: MessagePart[] = []

  // Add reconstructed message if available
  if (parsedStream.reconstructedMessage) {
    const message = parsedStream.reconstructedMessage

    // Add main content if present
    if (message.content) {
      parts.push({
        type: 'assistant',
        content: message.content,
        metadata: {
          provider,
          model: parsedStream.metadata.model,
          streamEvents: parsedStream.totalEvents
        },
        colorClass: MESSAGE_COLORS.assistant.bg,
        icon: MESSAGE_COLORS.assistant.icon
      })
    }

    // Add tool calls if present
    if (message.tool_calls && Array.isArray(message.tool_calls)) {
      for (const toolCall of message.tool_calls) {
        const argumentsText = typeof toolCall.function.arguments === 'object' 
          ? JSON.stringify(toolCall.function.arguments, null, 2)
          : toolCall.function.arguments

        parts.push({
          type: 'tool_call',
          content: `Function: ${toolCall.function.name}\n\nArguments:\n${argumentsText}`,
          metadata: {
            toolCallId: toolCall.id,
            toolName: toolCall.function.name,
            arguments: toolCall.function.arguments,
            provider,
            model: parsedStream.metadata.model,
            streamEvents: parsedStream.totalEvents
          },
          colorClass: MESSAGE_COLORS.tool_call.bg,
          icon: MESSAGE_COLORS.tool_call.icon
        })
      }
    }
  }

  // If no reconstructed message, show a summary of the stream
  if (parts.length === 0) {
    parts.push({
      type: 'assistant',
      content: `Streaming response with ${parsedStream.totalEvents} events`,
      metadata: {
        provider,
        model: parsedStream.metadata.model,
        streamEvents: parsedStream.totalEvents,
        hasToolCalls: parsedStream.hasToolCalls
      },
      colorClass: MESSAGE_COLORS.assistant.bg,
      icon: '🌊'
    })
  }

  return parts
}