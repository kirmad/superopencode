import { useState } from 'react'
import { format } from 'date-fns'
import { 
  ChevronDown, 
  ChevronRight, 
  Bot, 
  User, 
  Settings, 
  AlertTriangle,
  Clock,
  DollarSign,
  Hash
} from 'lucide-react'
import { LLMCallLog } from '@/types'
import { MessageViewer } from './MessageViewer'

interface LLMCallViewerProps {
  call: LLMCallLog
}

export function LLMCallViewer({ call }: LLMCallViewerProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  
  const messages = call.request?.messages || []
  const response = call.response
  const hasError = !!call.error
  
  return (
    <div className="bg-card rounded-lg border border-border overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full p-4 text-left hover:bg-accent transition-colors"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
            <Bot className="h-5 w-5 text-blue-500" />
            <div>
              <h3 className="font-medium text-foreground">
                LLM Call - {call.provider} ({call.model})
              </h3>
              <p className="text-sm text-muted-foreground">
                {format(new Date(call.start_time), 'HH:mm:ss.SSS')}
                {call.end_time && (
                  <span className="ml-2">• {call.duration_ms}ms</span>
                )}
              </p>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            {call.tokens_used && (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <Hash className="h-4 w-4" />
                {call.tokens_used.total_tokens} tokens
              </div>
            )}
            
            {call.cost && (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <DollarSign className="h-4 w-4" />
                ${call.cost.toFixed(4)}
              </div>
            )}
            
            {hasError && (
              <div className="flex items-center gap-1 text-destructive">
                <AlertTriangle className="h-4 w-4" />
                <span className="text-sm font-medium">Error</span>
              </div>
            )}
          </div>
        </div>
      </button>

      {isExpanded && (
        <div className="border-t border-border">
          {hasError && (
            <div className="p-4 bg-destructive/10 border-b border-border">
              <div className="flex items-center gap-2 mb-2">
                <AlertTriangle className="h-4 w-4 text-destructive" />
                <span className="font-medium text-destructive">Error</span>
              </div>
              <pre className="text-sm text-destructive syntax-highlight overflow-x-auto">
                {call.error}
              </pre>
            </div>
          )}
          
          <div className="p-4 space-y-4">
            <div>
              <h4 className="text-sm font-medium text-foreground mb-3">Request Messages</h4>
              <div className="space-y-3">
                {messages.map((message: any, index: number) => (
                  <MessageViewer key={index} message={message} />
                ))}
              </div>
            </div>
            
            {response && (
              <div>
                <h4 className="text-sm font-medium text-foreground mb-3">Response</h4>
                <MessageViewer message={response} />
              </div>
            )}
            
            {call.stream_events && call.stream_events.length > 0 && (
              <div>
                <h4 className="text-sm font-medium text-foreground mb-3">Stream Events</h4>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {call.stream_events.map((event, index) => (
                    <div key={index} className="p-2 bg-background rounded text-xs">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-primary">{event.event_type}</span>
                        <span className="text-muted-foreground">
                          {format(new Date(event.timestamp), 'HH:mm:ss.SSS')}
                        </span>
                      </div>
                      <pre className="text-muted-foreground syntax-highlight overflow-x-auto">
                        {JSON.stringify(event.data, null, 2)}
                      </pre>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            {call.tokens_used && (
              <div className="p-3 bg-background rounded-lg">
                <h4 className="text-sm font-medium text-foreground mb-2">Token Usage</h4>
                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Prompt:</span>
                    <span className="ml-2 font-medium">{call.tokens_used.prompt_tokens}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Completion:</span>
                    <span className="ml-2 font-medium">{call.tokens_used.completion_tokens}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Total:</span>
                    <span className="ml-2 font-medium">{call.tokens_used.total_tokens}</span>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}