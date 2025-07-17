import { useState } from 'react'
import { ChevronDown, ChevronRight, User, Bot, Settings, Wrench } from 'lucide-react'
import clsx from 'clsx'

interface MessageViewerProps {
  message: any
}

export function MessageViewer({ message }: MessageViewerProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  
  const role = message.role || 'unknown'
  const content = message.content || message.message || ''
  const toolCalls = message.tool_calls || []
  
  const roleIcons = {
    system: Settings,
    user: User,
    assistant: Bot,
    tool: Wrench,
    function: Wrench,
  }
  
  const Icon = roleIcons[role as keyof typeof roleIcons] || Bot
  
  return (
    <div className={clsx(
      "rounded-lg border overflow-hidden",
      `role-${role}`
    )}>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full p-3 text-left hover:bg-opacity-20 hover:bg-current transition-colors"
      >
        <div className="flex items-center gap-3">
          {isExpanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
          <Icon className="h-4 w-4" />
          <span className="font-medium capitalize">{role}</span>
          {content && (
            <span className="text-sm opacity-75 truncate">
              {typeof content === 'string' 
                ? content.slice(0, 100) + (content.length > 100 ? '...' : '')
                : 'Complex content'
              }
            </span>
          )}
        </div>
      </button>

      {isExpanded && (
        <div className="border-t p-4 bg-background/50">
          {content && (
            <div className="mb-4">
              <h5 className="text-sm font-medium mb-2">Content</h5>
              <div className="bg-background rounded p-3">
                {typeof content === 'string' ? (
                  <pre className="text-sm syntax-highlight whitespace-pre-wrap overflow-x-auto">
                    {content}
                  </pre>
                ) : (
                  <pre className="text-sm syntax-highlight overflow-x-auto">
                    {JSON.stringify(content, null, 2)}
                  </pre>
                )}
              </div>
            </div>
          )}
          
          {toolCalls.length > 0 && (
            <div>
              <h5 className="text-sm font-medium mb-2">Tool Calls</h5>
              <div className="space-y-2">
                {toolCalls.map((toolCall: any, index: number) => (
                  <div key={index} className="bg-background rounded p-3">
                    <div className="flex items-center gap-2 mb-2">
                      <Wrench className="h-4 w-4 text-orange-500" />
                      <span className="font-medium">{toolCall.function?.name || toolCall.name || 'Unknown'}</span>
                    </div>
                    <pre className="text-sm syntax-highlight overflow-x-auto">
                      {JSON.stringify(toolCall.function?.arguments || toolCall.arguments || toolCall, null, 2)}
                    </pre>
                  </div>
                ))}
              </div>
            </div>
          )}
          
          {message.function_call && (
            <div>
              <h5 className="text-sm font-medium mb-2">Function Call</h5>
              <div className="bg-background rounded p-3">
                <div className="flex items-center gap-2 mb-2">
                  <Wrench className="h-4 w-4 text-orange-500" />
                  <span className="font-medium">{message.function_call.name}</span>
                </div>
                <pre className="text-sm syntax-highlight overflow-x-auto">
                  {JSON.stringify(message.function_call.arguments, null, 2)}
                </pre>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}