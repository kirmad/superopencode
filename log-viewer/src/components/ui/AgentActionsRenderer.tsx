'use client'

import { useState } from 'react'
import { ChevronDown, ChevronRight, Bot, Wrench, MessageSquare, Code, FileText, Terminal } from 'lucide-react'
import { JsonRenderer } from './JsonRenderer'

interface AgentAction {
  type: 'text_response' | 'tool_call' | 'tool_response' | 'thinking' | 'analysis'
  content?: string
  toolName?: string
  arguments?: Record<string, unknown>
  result?: Record<string, unknown>
  summary: string
  timestamp?: string
  metadata?: Record<string, unknown>
}

interface AgentActionsRendererProps {
  messages: Array<Record<string, unknown>>
  responseBody?: Record<string, unknown> | string
  title?: string
}

export function AgentActionsRenderer({ messages, responseBody, title = "Agent Actions" }: AgentActionsRendererProps) {
  const [isExpanded, setIsExpanded] = useState(true)
  const [selectedAction, setSelectedAction] = useState<number | null>(null)

  // Parse messages and response body to extract agent actions
  const parseAgentActions = (): AgentAction[] => {
    const actions: AgentAction[] = []
    

    // Parse messages array - data is already parsed into MessagePart format
    if (messages && Array.isArray(messages)) {
      for (let i = 0; i < messages.length; i++) {
        const message = messages[i]
        
        // Handle MessagePart objects (parsed data)
        if (message.type === 'tool_call' && (message.metadata as any)?.toolName) {
          const toolName = (message.metadata as any).toolName as string
          const args = ((message.metadata as any).arguments as Record<string, unknown>) || {}
          
          
          const action: AgentAction = {
            type: 'tool_call',
            toolName,
            arguments: args,
            summary: generateToolCallSummary(toolName, args),
            metadata: {
              id: (message.metadata as any).toolCallId as string,
              provider: (message.metadata as any).provider as string,
              model: (message.metadata as any).model as string
            }
          }
          
          actions.push(action)
        }
        
        // Handle assistant text responses
        else if (message.type === 'assistant' && message.content) {
          actions.push({
            type: 'text_response',
            content: message.content as string,
            summary: generateTextResponseSummary(message.content as string),
            metadata: message.metadata as Record<string, unknown>
          })
        }
        
        // Handle raw tool_calls format (if it exists)
        else if (message.role === 'assistant' && message.tool_calls && Array.isArray(message.tool_calls)) {
          
          // Process tool calls in pairs - first has id/name, second has arguments
          let j = 0
          while (j < message.tool_calls.length) {
            const toolCall = message.tool_calls[j]
            
            // Check if this is a tool call with ID (indicates start of a pair)
            if (toolCall.id && toolCall.function?.name) {
              const toolName = toolCall.function.name
              let parsedArgs = {}
              
              // Check if initial call has any arguments
              if (toolCall.function?.arguments) {
                const args = toolCall.function.arguments
                if (typeof args === 'string' && args !== '{}') {
                  try {
                    parsedArgs = JSON.parse(args)
                  } catch {
                    parsedArgs = { raw: args }
                  }
                } else if (typeof args === 'object' && Object.keys(args).length > 0) {
                  parsedArgs = args
                }
              }
              
              // Look for the next tool call without ID - this contains the arguments
              if (j + 1 < message.tool_calls.length) {
                const nextToolCall = message.tool_calls[j + 1]
                // Next call should not have an ID and should have arguments
                if (!nextToolCall.id && nextToolCall.function?.arguments) {
                  const additionalArgs = nextToolCall.function.arguments
                  if (typeof additionalArgs === 'object') {
                    parsedArgs = { ...parsedArgs, ...additionalArgs }
                  } else if (typeof additionalArgs === 'string') {
                    try {
                      const parsed = JSON.parse(additionalArgs)
                      parsedArgs = { ...parsedArgs, ...parsed }
                    } catch {
                      parsedArgs = { ...parsedArgs, raw: additionalArgs }
                    }
                  }
                  j++ // Skip the argument entry as we've processed it
                }
              }
              
              const action: AgentAction = {
                type: 'tool_call',
                toolName,
                arguments: parsedArgs,
                summary: generateToolCallSummary(toolName, parsedArgs),
                metadata: {
                  id: toolCall.id,
                  type: toolCall.type
                }
              }
              
              actions.push(action)
            }
            
            j++
          }
        }
        
        // Handle user messages
        else if (message.type === 'user' && message.content) {
          // Don't add user messages to actions, they're just input
        }
      }
    }

    // Parse responseBody for SSE streams
    if (responseBody && typeof responseBody === 'string' && responseBody.includes('data:')) {
      const streamActions = parseSSEForActions(responseBody)
      actions.push(...streamActions)
    }

    // Parse direct response body
    if (responseBody && typeof responseBody === 'object') {
      const responseActions = parseResponseBodyForActions(responseBody)
      actions.push(...responseActions)
    }

    return actions.filter(action => action.summary) // Remove empty actions
  }

  const generateToolCallSummary = (toolName: string, args: Record<string, unknown>): string => {
    const hasArgs = args && Object.keys(args).length > 0
    
    // Special handling for common tool patterns
    switch (toolName.toLowerCase()) {
      case 'ls':
        return hasArgs && args.path 
          ? `Listed directory: ${args.path}`
          : 'Listed current directory'
      case 'view':
      case 'read':
        return hasArgs && args.file_path 
          ? `Viewed file: ${args.file_path}${args.limit ? ` (first ${args.limit} lines)` : ''}`
          : `Called ${toolName}`
      case 'bash':
      case 'shell':
      case 'terminal':
        return hasArgs && args.command 
          ? `Executed: ${args.command}`
          : `Called ${toolName}`
      case 'write':
        return hasArgs && args.file_path 
          ? `Wrote to file: ${args.file_path}`
          : `Called ${toolName}`
      case 'edit':
        return hasArgs && args.file_path 
          ? `Edited file: ${args.file_path}`
          : `Called ${toolName}`
      default:
        const formattedArgs = hasArgs 
          ? ` with ${formatArguments(args)}`
          : ''
        return `Called ${toolName}${formattedArgs}`
    }
  }

  const generateTextResponseSummary = (content: string): string => {
    if (!content) return ''
    
    const truncated = content.length > 100 
      ? content.substring(0, 100) + '...'
      : content
    
    return `Responded: "${truncated}"`
  }

  const formatArguments = (args: Record<string, unknown>): string => {
    const keys = Object.keys(args)
    if (keys.length === 0) return 'no arguments'
    if (keys.length === 1) {
      const key = keys[0]
      const value = args[key]
      if (typeof value === 'string') {
        if (value.length < 50) {
          return `${key}: "${value}"`
        } else {
          return `${key}: "${value.substring(0, 47)}..."`
        }
      } else if (typeof value === 'number' || typeof value === 'boolean') {
        return `${key}: ${value}`
      }
      return `${key}: ${typeof value}`
    }
    if (keys.length <= 3) {
      const preview = keys.slice(0, 2).map(key => {
        const value = args[key]
        if (typeof value === 'string' && value.length < 20) {
          return `${key}: "${value}"`
        } else if (typeof value === 'number' || typeof value === 'boolean') {
          return `${key}: ${value}`
        }
        return `${key}: ${typeof value}`
      }).join(', ')
      return keys.length > 2 ? `${preview}, ...` : preview
    }
    return `${keys.length} arguments`
  }

  const parseSSEForActions = (sseData: string): AgentAction[] => {
    const actions: AgentAction[] = []
    const lines = sseData.split('\n')
    const toolCallsMap = new Map() // Track tool calls across multiple events
    
    for (const line of lines) {
      if (line.startsWith('data: ') && line !== 'data: [DONE]') {
        try {
          const data = JSON.parse(line.slice(6))
          
          // Check for tool calls in streaming
          if (data.choices?.[0]?.delta?.tool_calls) {
            const toolCalls = data.choices[0].delta.tool_calls
            for (const toolCall of toolCalls) {
              const toolCallId = toolCall.id || `tool_${toolCall.index || Date.now()}`
              
              if (toolCall.function?.name) {
                // This is a function name event
                toolCallsMap.set(toolCallId, {
                  name: toolCall.function.name,
                  arguments: {},
                  argumentsString: '',
                  id: toolCallId,
                  index: toolCall.index
                })
              } else if (toolCall.function?.arguments) {
                // This is an arguments event - accumulate with existing tool call
                const existing = toolCallsMap.get(toolCallId)
                if (existing) {
                  existing.argumentsString += toolCall.function.arguments
                  try {
                    // Try to parse the accumulated arguments
                    existing.arguments = JSON.parse(existing.argumentsString)
                  } catch {
                    // Keep accumulating if not complete JSON yet
                    existing.arguments = { raw: existing.argumentsString }
                  }
                }
              }
            }
          }
          
          // Check for content in streaming
          if (data.choices?.[0]?.delta?.content) {
            const content = data.choices[0].delta.content
            actions.push({
              type: 'text_response',
              content,
              summary: `Responded with text content`,
              metadata: { streamEvent: true, contentChunk: true }
            })
          }
        } catch {
          // Ignore parsing errors for individual events
        }
      }
    }
    
    // Convert completed tool calls to actions
    toolCallsMap.forEach((toolCall) => {
      if (toolCall.name) {
        actions.push({
          type: 'tool_call',
          toolName: toolCall.name,
          arguments: toolCall.arguments,
          summary: generateToolCallSummary(toolCall.name, toolCall.arguments),
          metadata: { 
            streamEvent: true, 
            id: toolCall.id,
            index: toolCall.index,
            argumentsString: toolCall.argumentsString
          }
        })
      }
    })
    
    return actions
  }

  const parseResponseBodyForActions = (body: any): AgentAction[] => {
    const actions: AgentAction[] = []
    
    // Check for direct tool calls
    if (body.tool_calls && Array.isArray(body.tool_calls)) {
      for (const toolCall of body.tool_calls) {
        actions.push({
          type: 'tool_call',
          toolName: toolCall.function?.name || toolCall.name,
          arguments: toolCall.function?.arguments || toolCall.arguments,
          summary: generateToolCallSummary(toolCall.function?.name || toolCall.name, toolCall.function?.arguments || toolCall.arguments || {}),
          metadata: body
        })
      }
    }
    
    // Check for content in choices
    if (body.choices?.[0]?.message?.content) {
      actions.push({
        type: 'text_response',
        content: body.choices[0].message.content,
        summary: generateTextResponseSummary(body.choices[0].message.content)
      })
    }
    
    return actions
  }

  const getActionIcon = (action: AgentAction) => {
    switch (action.type) {
      case 'tool_call':
        return <Wrench className="h-4 w-4" />
      case 'tool_response':
        return <Code className="h-4 w-4" />
      case 'text_response':
        return <MessageSquare className="h-4 w-4" />
      case 'thinking':
        return <Bot className="h-4 w-4" />
      case 'analysis':
        return <FileText className="h-4 w-4" />
      default:
        return <Terminal className="h-4 w-4" />
    }
  }

  const getActionStyle = (action: AgentAction) => {
    switch (action.type) {
      case 'tool_call':
        return {
          bgColor: 'bg-purple-50',
          borderColor: 'border-purple-200',
          textColor: 'text-purple-800',
          iconColor: 'text-purple-600'
        }
      case 'tool_response':
        return {
          bgColor: 'bg-orange-50',
          borderColor: 'border-orange-200',
          textColor: 'text-orange-800',
          iconColor: 'text-orange-600'
        }
      case 'text_response':
        return {
          bgColor: 'bg-emerald-50',
          borderColor: 'border-emerald-200',
          textColor: 'text-emerald-800',
          iconColor: 'text-emerald-600'
        }
      default:
        return {
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200',
          textColor: 'text-gray-800',
          iconColor: 'text-gray-600'
        }
    }
  }

  const actions = parseAgentActions()

  // Always show the component if we have any messages (for debugging)
  if (actions.length === 0 && (!messages || messages.length === 0)) {
    return null
  }

  return (
    <div className="bg-white rounded-lg border border-slate-200">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-3 py-1.5 flex items-center justify-between text-left hover:bg-slate-50 transition-colors border-b border-slate-200"
      >
        <div className="flex items-center gap-2">
          <Bot className="h-3 w-3 text-emerald-600" />
          <h3 className="text-sm font-medium text-slate-900">{title}</h3>
          <span className="px-1.5 py-0.5 bg-emerald-100 text-emerald-800 rounded text-xs font-medium">
            {actions.length} action{actions.length !== 1 ? 's' : ''}
            {actions.length === 0 ? ' (debug)' : ''}
          </span>
        </div>
        {isExpanded ? (
          <ChevronDown className="h-3 w-3 text-slate-600" />
        ) : (
          <ChevronRight className="h-3 w-3 text-slate-600" />
        )}
      </button>

      {isExpanded && (
        <div className="p-3">
          {actions.length === 0 && (
            <div className="text-xs text-gray-500 bg-yellow-50 p-2 rounded border">
              <strong>Debug Info:</strong>
              <br />- Messages received: {messages?.length || 0}
              <br />- Message types: {messages?.map(m => m.type || m.role).join(', ') || 'none'}
              <br />- First message: {messages?.[0] ? JSON.stringify(messages[0], null, 2).substring(0, 200) + '...' : 'none'}
            </div>
          )}
          <div className="space-y-2">
            {actions.map((action, index) => {
              const style = getActionStyle(action)
              const isSelected = selectedAction === index
              
              return (
                <div 
                  key={index}
                  className={`border rounded-lg transition-all ${style.bgColor} ${style.borderColor} ${
                    isSelected ? 'ring-2 ring-blue-200' : ''
                  }`}
                  data-action-type={action.type}
                >
                  <button
                    onClick={() => setSelectedAction(isSelected ? null : index)}
                    className="w-full p-2 text-left hover:bg-white/50 transition-colors rounded-lg"
                  >
                    <div className="flex items-center gap-2">
                      <div className={style.iconColor}>
                        {getActionIcon(action)}
                      </div>
                      <span className={`text-sm font-medium ${style.textColor}`}>
                        {action.summary}
                      </span>
                      <div className="ml-auto flex items-center gap-1">
                        {action.toolName && (
                          <span className="px-1.5 py-0.5 bg-white/80 rounded text-xs font-mono text-gray-700">
                            {action.toolName}
                          </span>
                        )}
                        {isSelected ? (
                          <ChevronDown className="h-3 w-3 text-gray-500" />
                        ) : (
                          <ChevronRight className="h-3 w-3 text-gray-500" />
                        )}
                      </div>
                    </div>
                  </button>

                  {isSelected && (
                    <div className="px-2 pb-2 space-y-2">
                      {action.content && (
                        <div>
                          <label className="text-xs font-medium text-gray-600 block mb-1">Content</label>
                          <div className="p-2 bg-white rounded border text-sm">
                            {action.content.length > 500 ? (
                              <details>
                                <summary className="cursor-pointer text-blue-600 text-xs mb-1">
                                  Show full content ({action.content.length} characters)
                                </summary>
                                <div className="whitespace-pre-wrap text-gray-900 leading-relaxed">{action.content}</div>
                              </details>
                            ) : (
                              <div className="whitespace-pre-wrap text-gray-900 leading-relaxed">{action.content}</div>
                            )}
                          </div>
                        </div>
                      )}

                      {action.arguments && Object.keys(action.arguments).length > 0 && (
                        <div>
                          <label className="text-xs font-medium text-gray-600 block mb-1">Arguments</label>
                          <JsonRenderer 
                            data={action.arguments}
                            title="Tool Arguments"
                            maxHeight="150px"
                          />
                        </div>
                      )}

                      {action.result && (
                        <div>
                          <label className="text-xs font-medium text-gray-600 block mb-1">Result</label>
                          <JsonRenderer 
                            data={action.result}
                            title="Tool Result"
                            maxHeight="150px"
                          />
                        </div>
                      )}

                      {action.metadata && Object.keys(action.metadata).length > 0 && (
                        <div>
                          <label className="text-xs font-medium text-gray-600 block mb-1">Metadata</label>
                          <JsonRenderer 
                            data={action.metadata}
                            title="Action Metadata"
                            maxHeight="100px"
                          />
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}