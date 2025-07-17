import { useState } from 'react'
import { format } from 'date-fns'
import { 
  ChevronDown, 
  ChevronRight, 
  Wrench, 
  AlertTriangle,
  Clock
} from 'lucide-react'
import { ToolCallLog } from '@/types'

interface ToolCallViewerProps {
  call: ToolCallLog
}

export function ToolCallViewer({ call }: ToolCallViewerProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  
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
            <Wrench className="h-5 w-5 text-orange-500" />
            <div>
              <h3 className="font-medium text-foreground">
                Tool Call - {call.name}
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
              <h4 className="text-sm font-medium text-foreground mb-3">Input</h4>
              <div className="bg-background rounded p-3">
                <pre className="text-sm syntax-highlight overflow-x-auto">
                  {JSON.stringify(call.input, null, 2)}
                </pre>
              </div>
            </div>
            
            {call.output && (
              <div>
                <h4 className="text-sm font-medium text-foreground mb-3">Output</h4>
                <div className="bg-background rounded p-3">
                  <pre className="text-sm syntax-highlight overflow-x-auto">
                    {typeof call.output === 'string' 
                      ? call.output 
                      : JSON.stringify(call.output, null, 2)
                    }
                  </pre>
                </div>
              </div>
            )}
            
            {(call.parent_id || call.child_ids?.length) && (
              <div className="p-3 bg-background rounded-lg">
                <h4 className="text-sm font-medium text-foreground mb-2">Hierarchy</h4>
                <div className="space-y-1 text-sm">
                  {call.parent_id && (
                    <div>
                      <span className="text-muted-foreground">Parent:</span>
                      <span className="ml-2 font-mono">{call.parent_id}</span>
                    </div>
                  )}
                  {call.child_ids && call.child_ids.length > 0 && (
                    <div>
                      <span className="text-muted-foreground">Children:</span>
                      <div className="ml-2 font-mono">
                        {call.child_ids.map(id => (
                          <div key={id}>{id}</div>
                        ))}
                      </div>
                    </div>
                  )}
                  {call.parent_llm_call && (
                    <div>
                      <span className="text-muted-foreground">Parent LLM Call:</span>
                      <span className="ml-2 font-mono">{call.parent_llm_call}</span>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}