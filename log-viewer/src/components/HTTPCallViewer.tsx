import { useState } from 'react'
import { format } from 'date-fns'
import { 
  ChevronDown, 
  ChevronRight, 
  Globe, 
  AlertTriangle,
  Check,
  X
} from 'lucide-react'
import { HTTPLog } from '@/types'

interface HTTPCallViewerProps {
  call: HTTPLog
}

export function HTTPCallViewer({ call }: HTTPCallViewerProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  
  const hasError = !!call.error || (call.status_code && call.status_code >= 400)
  const isSuccess = call.status_code && call.status_code >= 200 && call.status_code < 300
  
  const getStatusColor = () => {
    if (hasError) return 'text-destructive'
    if (isSuccess) return 'text-green-600'
    return 'text-muted-foreground'
  }
  
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
            <Globe className="h-5 w-5 text-purple-500" />
            <div>
              <h3 className="font-medium text-foreground">
                HTTP {call.method} - {call.url}
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
            {call.status_code && (
              <div className={`flex items-center gap-1 text-sm font-medium ${getStatusColor()}`}>
                {isSuccess ? (
                  <Check className="h-4 w-4" />
                ) : hasError ? (
                  <X className="h-4 w-4" />
                ) : null}
                {call.status_code}
              </div>
            )}
            
            {call.error && (
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
          {call.error && (
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
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <h4 className="text-sm font-medium text-foreground mb-3">Request</h4>
                
                <div className="space-y-3">
                  <div className="bg-background rounded p-3">
                    <h5 className="text-xs font-medium text-muted-foreground mb-2">URL</h5>
                    <code className="text-sm syntax-highlight break-all">{call.url}</code>
                  </div>
                  
                  {Object.keys(call.headers || {}).length > 0 && (
                    <div className="bg-background rounded p-3">
                      <h5 className="text-xs font-medium text-muted-foreground mb-2">Headers</h5>
                      <pre className="text-sm syntax-highlight overflow-x-auto">
                        {JSON.stringify(call.headers, null, 2)}
                      </pre>
                    </div>
                  )}
                  
                  {call.body && (
                    <div className="bg-background rounded p-3">
                      <h5 className="text-xs font-medium text-muted-foreground mb-2">Body</h5>
                      <pre className="text-sm syntax-highlight overflow-x-auto">
                        {typeof call.body === 'string' 
                          ? call.body 
                          : JSON.stringify(call.body, null, 2)
                        }
                      </pre>
                    </div>
                  )}
                </div>
              </div>
              
              <div>
                <h4 className="text-sm font-medium text-foreground mb-3">Response</h4>
                
                <div className="space-y-3">
                  {call.status_code && (
                    <div className="bg-background rounded p-3">
                      <h5 className="text-xs font-medium text-muted-foreground mb-2">Status</h5>
                      <span className={`text-sm font-medium ${getStatusColor()}`}>
                        {call.status_code}
                      </span>
                    </div>
                  )}
                  
                  {call.response_headers && Object.keys(call.response_headers).length > 0 && (
                    <div className="bg-background rounded p-3">
                      <h5 className="text-xs font-medium text-muted-foreground mb-2">Headers</h5>
                      <pre className="text-sm syntax-highlight overflow-x-auto">
                        {JSON.stringify(call.response_headers, null, 2)}
                      </pre>
                    </div>
                  )}
                  
                  {call.response_body && (
                    <div className="bg-background rounded p-3">
                      <h5 className="text-xs font-medium text-muted-foreground mb-2">Body</h5>
                      <pre className="text-sm syntax-highlight overflow-x-auto max-h-64">
                        {typeof call.response_body === 'string' 
                          ? call.response_body 
                          : JSON.stringify(call.response_body, null, 2)
                        }
                      </pre>
                    </div>
                  )}
                </div>
              </div>
            </div>
            
            {call.parent_tool_call && (
              <div className="p-3 bg-background rounded-lg">
                <h4 className="text-sm font-medium text-foreground mb-2">Context</h4>
                <div className="text-sm">
                  <span className="text-muted-foreground">Parent Tool Call:</span>
                  <span className="ml-2 font-mono">{call.parent_tool_call}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}