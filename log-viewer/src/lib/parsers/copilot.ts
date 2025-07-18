import { MessagePart } from '../types'
import { MESSAGE_COLORS } from '../utils/colors'

export function parseCopilotRequest(requestBody: Record<string, unknown>): MessagePart[] {
  const parts: MessagePart[] = []
  
  if (!requestBody.messages || !Array.isArray(requestBody.messages)) return parts
  
  for (const message of requestBody.messages as Record<string, unknown>[]) {
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
            .map((part: Record<string, unknown>) => part.type === 'text' ? part.text : `[${part.type}]`)
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
            content: message.content as string,
            colorClass: MESSAGE_COLORS.assistant.bg,
            icon: MESSAGE_COLORS.assistant.icon
          })
        }
        
        if (message.tool_calls && Array.isArray(message.tool_calls)) {
          // Handle paired tool call format: first element has id/name, second has arguments
          let i = 0
          while (i < (message.tool_calls as Record<string, unknown>[]).length) {
            const toolCall = (message.tool_calls as Record<string, unknown>[])[i]
            
            // Check if this is a tool call with ID (indicates start of a pair)
            if (toolCall.id && toolCall.function && (toolCall.function as Record<string, unknown>).name) {
              const func = toolCall.function as Record<string, unknown>
              const toolName = func.name as string
              let args = {}
              
              // Check if this tool call has arguments
              if (func.arguments) {
                if (typeof func.arguments === 'string') {
                  try {
                    args = JSON.parse(func.arguments)
                  } catch {
                    args = { raw: func.arguments }
                  }
                } else if (typeof func.arguments === 'object') {
                  args = func.arguments as Record<string, unknown>
                }
              }
              
              // Look for the next tool call without ID - this contains the arguments
              if (i + 1 < (message.tool_calls as Record<string, unknown>[]).length) {
                const nextToolCall = (message.tool_calls as Record<string, unknown>[])[i + 1]
                if (!nextToolCall.id && nextToolCall.function && (nextToolCall.function as Record<string, unknown>).arguments) {
                  const nextArgs = (nextToolCall.function as Record<string, unknown>).arguments
                  if (typeof nextArgs === 'object') {
                    args = { ...args, ...nextArgs as Record<string, unknown> }
                  }
                  i++ // Skip the argument entry as we've processed it
                }
              }
              
              parts.push({
                type: 'tool_call',
                content: `Function: ${toolName}\n\nArguments:\n${JSON.stringify(args, null, 2)}`,
                metadata: {
                  toolCallId: toolCall.id as string,
                  toolName,
                  arguments: args
                },
                colorClass: MESSAGE_COLORS.tool_call.bg,
                icon: MESSAGE_COLORS.tool_call.icon
              })
            }
            
            i++
          }
        }
        break
        
      case 'tool':
        parts.push({
          type: 'tool_response',
          content: message.content as string,
          metadata: {
            toolCallId: message.tool_call_id as string
          },
          colorClass: MESSAGE_COLORS.tool_response.bg,
          icon: MESSAGE_COLORS.tool_response.icon
        })
        break
    }
  }
  
  return parts
}

export function parseCopilotResponse(responseBody: Record<string, unknown>): MessagePart[] {
  const parts: MessagePart[] = []
  
  if (!responseBody.choices || !Array.isArray(responseBody.choices) || !responseBody.choices[0]) return parts
  
  const choice = responseBody.choices[0] as Record<string, unknown>
  const message = choice.message as Record<string, unknown>
  
  if (message.content) {
    parts.push({
      type: 'assistant',
      content: message.content as string,
      colorClass: MESSAGE_COLORS.assistant.bg,
      icon: MESSAGE_COLORS.assistant.icon
    })
  }
  
  if (message.tool_calls && Array.isArray(message.tool_calls)) {
    // Handle paired tool call format: first element has id/name, second has arguments
    let i = 0
    while (i < (message.tool_calls as Record<string, unknown>[]).length) {
      const toolCall = (message.tool_calls as Record<string, unknown>[])[i]
      
      // Check if this is a tool call with ID (indicates start of a pair)
      if (toolCall.id && toolCall.function && (toolCall.function as Record<string, unknown>).name) {
        const func = toolCall.function as Record<string, unknown>
        const toolName = func.name as string
        let args = {}
        
        // Check if this tool call has arguments
        if (func.arguments) {
          if (typeof func.arguments === 'string') {
            try {
              args = JSON.parse(func.arguments)
            } catch {
              args = { raw: func.arguments }
            }
          } else if (typeof func.arguments === 'object') {
            args = func.arguments as Record<string, unknown>
          }
        }
        
        // Look for the next tool call without ID - this contains the arguments
        if (i + 1 < (message.tool_calls as Record<string, unknown>[]).length) {
          const nextToolCall = (message.tool_calls as Record<string, unknown>[])[i + 1]
          if (!nextToolCall.id && nextToolCall.function && (nextToolCall.function as Record<string, unknown>).arguments) {
            const nextArgs = (nextToolCall.function as Record<string, unknown>).arguments
            if (typeof nextArgs === 'object') {
              args = { ...args, ...nextArgs as Record<string, unknown> }
            }
            i++ // Skip the argument entry as we've processed it
          }
        }
        
        parts.push({
          type: 'tool_call',
          content: `Function: ${toolName}\n\nArguments:\n${JSON.stringify(args, null, 2)}`,
          metadata: {
            toolCallId: toolCall.id as string,
            toolName,
            arguments: args
          },
          colorClass: MESSAGE_COLORS.tool_call.bg,
          icon: MESSAGE_COLORS.tool_call.icon
        })
      }
      
      i++
    }
  }
  
  return parts
}

