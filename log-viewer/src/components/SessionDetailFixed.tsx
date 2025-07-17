import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { 
  Clock, 
  AlertTriangle, 
  CheckCircle,
  Activity,
  DollarSign,
  Zap,
  Globe
} from 'lucide-react'
import { apiService } from '@/services/api'

export function SessionDetailFixed() {
  const { sessionId } = useParams<{ sessionId: string }>()
  
  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', sessionId],
    queryFn: () => apiService.getSession(sessionId!),
    enabled: !!sessionId,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error || !session) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="text-center">
          <AlertTriangle className="h-12 w-12 text-destructive mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-foreground mb-2">Failed to load session</h3>
          <p className="text-muted-foreground">Session {sessionId} could not be found</p>
        </div>
      </div>
    )
  }

  const totalTokens = session.llm_calls.reduce((sum, call) => 
    sum + (call.tokens_used?.total_tokens || 0), 0
  )
  
  const totalCost = session.llm_calls.reduce((sum, call) => 
    sum + (call.cost || 0), 0
  )

  const hasErrors = session.llm_calls.some(call => call.error) || 
                   session.tool_calls.some(call => call.error) ||
                   session.http_calls.some(call => call.error)

  const duration = session.end_time 
    ? Math.round((new Date(session.end_time).getTime() - new Date(session.start_time).getTime()) / 1000)
    : 'Ongoing'

  return (
    <div className="space-y-6">
      <div className="bg-card rounded-lg border border-border p-6">
        <div className="flex items-start justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-foreground mb-2">
              Session {sessionId}
            </h1>
            <p className="text-muted-foreground">
              Started {format(new Date(session.start_time), 'PPpp')}
              {session.end_time && (
                <span className="ml-2">
                  • Ended {format(new Date(session.end_time), 'PPpp')}
                </span>
              )}
            </p>
          </div>
          
          <div className="flex items-center gap-2">
            {hasErrors ? (
              <div className="flex items-center gap-1 text-destructive">
                <AlertTriangle className="h-5 w-5" />
                <span className="font-medium">Has Errors</span>
              </div>
            ) : (
              <div className="flex items-center gap-1 text-green-600">
                <CheckCircle className="h-5 w-5" />
                <span className="font-medium">Success</span>
              </div>
            )}
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-5 gap-6">
          <div className="text-center p-4 bg-background rounded-lg">
            <Clock className="h-6 w-6 text-muted-foreground mx-auto mb-2" />
            <div className="text-2xl font-bold text-foreground">{duration}</div>
            <div className="text-sm text-muted-foreground">Duration</div>
          </div>
          
          <div className="text-center p-4 bg-background rounded-lg">
            <Activity className="h-6 w-6 text-muted-foreground mx-auto mb-2" />
            <div className="text-2xl font-bold text-foreground">{session.llm_calls.length}</div>
            <div className="text-sm text-muted-foreground">LLM Calls</div>
          </div>
          
          <div className="text-center p-4 bg-background rounded-lg">
            <Zap className="h-6 w-6 text-muted-foreground mx-auto mb-2" />
            <div className="text-2xl font-bold text-foreground">{session.tool_calls.length}</div>
            <div className="text-sm text-muted-foreground">Tool Calls</div>
          </div>
          
          <div className="text-center p-4 bg-background rounded-lg">
            <Globe className="h-6 w-6 text-muted-foreground mx-auto mb-2" />
            <div className="text-2xl font-bold text-foreground">{session.http_calls.length}</div>
            <div className="text-sm text-muted-foreground">HTTP Calls</div>
          </div>
          
          <div className="text-center p-4 bg-background rounded-lg">
            <DollarSign className="h-6 w-6 text-muted-foreground mx-auto mb-2" />
            <div className="text-2xl font-bold text-foreground">${totalCost.toFixed(4)}</div>
            <div className="text-sm text-muted-foreground">Total Cost</div>
          </div>
        </div>

        {session.command_args.length > 0 && (
          <div className="mt-6 p-4 bg-background rounded-lg">
            <h3 className="text-sm font-medium text-foreground mb-2">Command Arguments</h3>
            <code className="text-sm text-muted-foreground syntax-highlight">
              {session.command_args.join(' ')}
            </code>
          </div>
        )}
      </div>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-foreground">Request Timeline</h2>
        
        <div className="space-y-4">
          {session.llm_calls.map((call) => (
            <div key={call.id} className="bg-card rounded-lg border border-border p-4">
              <div className="flex items-center gap-3 mb-3">
                <Activity className="h-5 w-5 text-blue-500" />
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
              
              <div className="text-sm text-muted-foreground">
                {call.tokens_used && (
                  <span>Tokens: {call.tokens_used.total_tokens}</span>
                )}
                {call.cost && (
                  <span className="ml-4">Cost: ${call.cost.toFixed(4)}</span>
                )}
                {call.error && (
                  <span className="ml-4 text-destructive">Error: {call.error}</span>
                )}
              </div>
            </div>
          ))}
          
          {session.tool_calls.map((call) => (
            <div key={call.id} className="bg-card rounded-lg border border-border p-4">
              <div className="flex items-center gap-3 mb-3">
                <Zap className="h-5 w-5 text-orange-500" />
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
              
              <div className="text-sm text-muted-foreground">
                {call.error && (
                  <span className="text-destructive">Error: {call.error}</span>
                )}
              </div>
            </div>
          ))}
          
          {session.http_calls.map((call) => (
            <div key={call.id} className="bg-card rounded-lg border border-border p-4">
              <div className="flex items-center gap-3 mb-3">
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
              
              <div className="text-sm text-muted-foreground">
                {call.status_code && (
                  <span>Status: {call.status_code}</span>
                )}
                {call.error && (
                  <span className="ml-4 text-destructive">Error: {call.error}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}