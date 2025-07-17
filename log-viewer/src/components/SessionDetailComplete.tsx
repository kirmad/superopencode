import { useState } from 'react'
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
  Globe,
  ChevronDown,
  ChevronRight,
  User,
  Bot,
  Settings,
  Wrench,
  Hash,
  Play,
  Pause,
  Info
} from 'lucide-react'
import { apiService } from '@/services/api'
import { LLMCallLog, ToolCallLog, HTTPLog } from '@/types'
import clsx from 'clsx'

interface CollapsibleSectionProps {
  title: string
  icon: React.ReactNode
  children: React.ReactNode
  defaultOpen?: boolean
  className?: string
}

function CollapsibleSection({ title, icon, children, defaultOpen = false, className = '' }: CollapsibleSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)
  
  return (
    <div className={clsx("border border-border rounded-lg overflow-hidden", className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full p-4 bg-card hover:bg-accent transition-colors text-left"
      >
        <div className="flex items-center gap-3">
          {isOpen ? (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          )}
          {icon}
          <span className="font-medium text-foreground">{title}</span>
        </div>
      </button>
      
      {isOpen && (
        <div className="border-t border-border bg-background">
          {children}
        </div>
      )}
    </div>
  )
}

interface MessageItemProps {
  message: any
  index: number
}

function MessageItem({ message }: MessageItemProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  
  const role = message.role || 'unknown'
  const content = message.content || message.message || ''
  const toolCalls = message.tool_calls || []
  
  const roleColors = {
    system: 'bg-purple-50 text-purple-700 border-purple-200',
    user: 'bg-blue-50 text-blue-700 border-blue-200',
    assistant: 'bg-green-50 text-green-700 border-green-200',
    tool: 'bg-orange-50 text-orange-700 border-orange-200',
    function: 'bg-red-50 text-red-700 border-red-200',
  }
  
  const roleIcons = {
    system: Settings,
    user: User,
    assistant: Bot,
    tool: Wrench,
    function: Wrench,
  }
  
  const Icon = roleIcons[role as keyof typeof roleIcons] || Bot
  const colorClass = roleColors[role as keyof typeof roleColors] || 'bg-gray-50 text-gray-700 border-gray-200'
  
  return (
    <div className={clsx("rounded-lg border overflow-hidden", colorClass)}>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full p-3 text-left hover:bg-black/5 transition-colors"
      >
        <div className="flex items-center gap-3">
          {isExpanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
          <Icon className="h-4 w-4" />
          <span className="font-medium capitalize">{role}</span>
          {!isExpanded && content && (
            <span className="text-sm opacity-75 truncate flex-1">
              {typeof content === 'string' 
                ? content.slice(0, 80) + (content.length > 80 ? '...' : '')
                : 'Complex content'
              }
            </span>
          )}
        </div>
      </button>

      {isExpanded && (
        <div className="border-t p-4 bg-white/50">
          {content && (
            <div className="mb-4">
              <h5 className="text-sm font-medium mb-2 flex items-center gap-2">
                <Info className="h-4 w-4" />
                Content
              </h5>
              <div className="bg-white rounded p-3 border">
                <pre className="text-sm whitespace-pre-wrap overflow-x-auto">
                  {typeof content === 'string' ? content : JSON.stringify(content, null, 2)}
                </pre>
              </div>
            </div>
          )}
          
          {toolCalls.length > 0 && (
            <div>
              <h5 className="text-sm font-medium mb-2 flex items-center gap-2">
                <Wrench className="h-4 w-4" />
                Tool Calls
              </h5>
              <div className="space-y-2">
                {toolCalls.map((toolCall: any, tcIndex: number) => (
                  <div key={tcIndex} className="bg-white rounded p-3 border">
                    <div className="flex items-center gap-2 mb-2">
                      <Play className="h-4 w-4 text-orange-500" />
                      <span className="font-medium">{toolCall.function?.name || toolCall.name || 'Unknown'}</span>
                    </div>
                    <pre className="text-sm overflow-x-auto">
                      {JSON.stringify(toolCall.function?.arguments || toolCall.arguments || toolCall, null, 2)}
                    </pre>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

interface LLMCallCompactProps {
  call: LLMCallLog
  index: number
}

function LLMCallCompact({ call, index }: LLMCallCompactProps) {
  const messages = call.request?.messages || []
  const response = call.response
  const hasError = !!call.error
  
  const title = `LLM Call ${index + 1} - ${call.provider} (${call.model})`
  const subtitle = `${format(new Date(call.start_time), 'HH:mm:ss.SSS')} • ${call.duration_ms}ms`
  
  return (
    <CollapsibleSection
      title={title}
      icon={<Activity className="h-5 w-5 text-blue-500" />}
      className="bg-blue-50/30"
    >
      <div className="p-4 space-y-4">
        <div className="flex items-center gap-6 text-sm">
          {call.tokens_used && (
            <div className="flex items-center gap-1">
              <Hash className="h-4 w-4 text-muted-foreground" />
              <span>{call.tokens_used.total_tokens} tokens</span>
            </div>
          )}
          {call.cost && (
            <div className="flex items-center gap-1">
              <DollarSign className="h-4 w-4 text-muted-foreground" />
              <span>${call.cost.toFixed(4)}</span>
            </div>
          )}
          <div className="text-muted-foreground">{subtitle}</div>
        </div>
        
        {hasError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle className="h-4 w-4 text-red-600" />
              <span className="font-medium text-red-600">Error</span>
            </div>
            <pre className="text-sm text-red-700 overflow-x-auto">
              {call.error}
            </pre>
          </div>
        )}
        
        <div>
          <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
            <User className="h-4 w-4" />
            Request Messages ({messages.length})
          </h4>
          <div className="space-y-2">
            {messages.map((message: any, msgIndex: number) => (
              <MessageItem key={msgIndex} message={message} index={msgIndex} />
            ))}
          </div>
        </div>
        
        {response && (
          <div>
            <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
              <Bot className="h-4 w-4" />
              Response
            </h4>
            <MessageItem message={response} index={0} />
          </div>
        )}
        
        {call.tokens_used && (
          <div className="p-3 bg-gray-50 rounded border">
            <h4 className="text-sm font-medium mb-2">Token Usage</h4>
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
    </CollapsibleSection>
  )
}

interface ToolCallCompactProps {
  call: ToolCallLog
  index: number
}

function ToolCallCompact({ call, index }: ToolCallCompactProps) {
  const hasError = !!call.error
  
  const title = `Tool Call ${index + 1} - ${call.name}`
  const subtitle = `${format(new Date(call.start_time), 'HH:mm:ss.SSS')} • ${call.duration_ms}ms`
  
  return (
    <CollapsibleSection
      title={title}
      icon={<Zap className="h-5 w-5 text-orange-500" />}
      className="bg-orange-50/30"
    >
      <div className="p-4 space-y-4">
        <div className="text-sm text-muted-foreground">{subtitle}</div>
        
        {hasError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle className="h-4 w-4 text-red-600" />
              <span className="font-medium text-red-600">Error</span>
            </div>
            <pre className="text-sm text-red-700 overflow-x-auto">
              {call.error}
            </pre>
          </div>
        )}
        
        <div>
          <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
            <Play className="h-4 w-4" />
            Input
          </h4>
          <div className="bg-white rounded p-3 border">
            <pre className="text-sm overflow-x-auto">
              {JSON.stringify(call.input, null, 2)}
            </pre>
          </div>
        </div>
        
        {call.output && (
          <div>
            <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
              <Pause className="h-4 w-4" />
              Output
            </h4>
            <div className="bg-white rounded p-3 border">
              <pre className="text-sm overflow-x-auto">
                {typeof call.output === 'string' 
                  ? call.output 
                  : JSON.stringify(call.output, null, 2)
                }
              </pre>
            </div>
          </div>
        )}
        
        {(call.parent_id || call.child_ids?.length || call.parent_llm_call) && (
          <div className="p-3 bg-gray-50 rounded border">
            <h4 className="text-sm font-medium mb-2">Hierarchy</h4>
            <div className="space-y-1 text-sm">
              {call.parent_id && (
                <div>
                  <span className="text-muted-foreground">Parent:</span>
                  <span className="ml-2 font-mono text-xs">{call.parent_id}</span>
                </div>
              )}
              {call.parent_llm_call && (
                <div>
                  <span className="text-muted-foreground">Parent LLM Call:</span>
                  <span className="ml-2 font-mono text-xs">{call.parent_llm_call}</span>
                </div>
              )}
              {call.child_ids && call.child_ids.length > 0 && (
                <div>
                  <span className="text-muted-foreground">Children:</span>
                  <div className="ml-2 mt-1 space-y-1">
                    {call.child_ids.map(id => (
                      <div key={id} className="font-mono text-xs">{id}</div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </CollapsibleSection>
  )
}

interface HTTPCallCompactProps {
  call: HTTPLog
  index: number
}

function HTTPCallCompact({ call, index }: HTTPCallCompactProps) {
  const hasError = !!call.error || (call.status_code && call.status_code >= 400)
  const isSuccess = call.status_code && call.status_code >= 200 && call.status_code < 300
  
  const getStatusColor = () => {
    if (hasError) return 'text-red-600'
    if (isSuccess) return 'text-green-600'
    return 'text-gray-600'
  }
  
  const title = `HTTP Call ${index + 1} - ${call.method} ${call.url}`
  const subtitle = `${format(new Date(call.start_time), 'HH:mm:ss.SSS')} • ${call.duration_ms}ms`
  
  return (
    <CollapsibleSection
      title={title}
      icon={<Globe className="h-5 w-5 text-purple-500" />}
      className="bg-purple-50/30"
    >
      <div className="p-4 space-y-4">
        <div className="flex items-center gap-6 text-sm">
          {call.status_code && (
            <div className={clsx("flex items-center gap-1 font-medium", getStatusColor())}>
              <span>Status: {call.status_code}</span>
            </div>
          )}
          <div className="text-muted-foreground">{subtitle}</div>
        </div>
        
        {call.error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle className="h-4 w-4 text-red-600" />
              <span className="font-medium text-red-600">Error</span>
            </div>
            <pre className="text-sm text-red-700 overflow-x-auto">
              {call.error}
            </pre>
          </div>
        )}
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
              <Play className="h-4 w-4" />
              Request
            </h4>
            
            <div className="space-y-3">
              <div className="bg-white rounded p-3 border">
                <h5 className="text-xs font-medium text-gray-600 mb-2">URL</h5>
                <code className="text-sm break-all">{call.url}</code>
              </div>
              
              {call.headers && Object.keys(call.headers).length > 0 && (
                <div className="bg-white rounded p-3 border">
                  <h5 className="text-xs font-medium text-gray-600 mb-2">Headers</h5>
                  <pre className="text-sm overflow-x-auto">
                    {JSON.stringify(call.headers, null, 2)}
                  </pre>
                </div>
              )}
              
              {call.body && (
                <div className="bg-white rounded p-3 border">
                  <h5 className="text-xs font-medium text-gray-600 mb-2">Body</h5>
                  <pre className="text-sm overflow-x-auto">
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
            <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
              <Pause className="h-4 w-4" />
              Response
            </h4>
            
            <div className="space-y-3">
              {call.status_code && (
                <div className="bg-white rounded p-3 border">
                  <h5 className="text-xs font-medium text-gray-600 mb-2">Status</h5>
                  <span className={clsx("text-sm font-medium", getStatusColor())}>
                    {call.status_code}
                  </span>
                </div>
              )}
              
              {call.response_headers && Object.keys(call.response_headers).length > 0 && (
                <div className="bg-white rounded p-3 border">
                  <h5 className="text-xs font-medium text-gray-600 mb-2">Headers</h5>
                  <pre className="text-sm overflow-x-auto">
                    {JSON.stringify(call.response_headers, null, 2)}
                  </pre>
                </div>
              )}
              
              {call.response_body && (
                <div className="bg-white rounded p-3 border">
                  <h5 className="text-xs font-medium text-gray-600 mb-2">Body</h5>
                  <pre className="text-sm overflow-x-auto max-h-64">
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
          <div className="p-3 bg-gray-50 rounded border">
            <h4 className="text-sm font-medium mb-2">Context</h4>
            <div className="text-sm">
              <span className="text-muted-foreground">Parent Tool Call:</span>
              <span className="ml-2 font-mono text-xs">{call.parent_tool_call}</span>
            </div>
          </div>
        )}
      </div>
    </CollapsibleSection>
  )
}

export function SessionDetailComplete() {
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
          <p className="text-muted-foreground">
            Session {sessionId} could not be found
            {error && <span className="block mt-1 text-sm">Error: {error.toString()}</span>}
          </p>
        </div>
      </div>
    )
  }

  // const totalTokens = session.llm_calls.reduce((sum, call) => 
  //   sum + (call.tokens_used?.total_tokens || 0), 0
  // )
  
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

        <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
          <div className="text-center p-3 bg-background rounded-lg">
            <Clock className="h-5 w-5 text-muted-foreground mx-auto mb-2" />
            <div className="text-lg font-bold text-foreground">{duration}</div>
            <div className="text-xs text-muted-foreground">Duration</div>
          </div>
          
          <div className="text-center p-3 bg-background rounded-lg">
            <Activity className="h-5 w-5 text-muted-foreground mx-auto mb-2" />
            <div className="text-lg font-bold text-foreground">{session.llm_calls.length}</div>
            <div className="text-xs text-muted-foreground">LLM Calls</div>
          </div>
          
          <div className="text-center p-3 bg-background rounded-lg">
            <Zap className="h-5 w-5 text-muted-foreground mx-auto mb-2" />
            <div className="text-lg font-bold text-foreground">{session.tool_calls.length}</div>
            <div className="text-xs text-muted-foreground">Tool Calls</div>
          </div>
          
          <div className="text-center p-3 bg-background rounded-lg">
            <Globe className="h-5 w-5 text-muted-foreground mx-auto mb-2" />
            <div className="text-lg font-bold text-foreground">{session.http_calls.length}</div>
            <div className="text-xs text-muted-foreground">HTTP Calls</div>
          </div>
          
          <div className="text-center p-3 bg-background rounded-lg">
            <DollarSign className="h-5 w-5 text-muted-foreground mx-auto mb-2" />
            <div className="text-lg font-bold text-foreground">${totalCost.toFixed(4)}</div>
            <div className="text-xs text-muted-foreground">Total Cost</div>
          </div>
        </div>

        {session.command_args.length > 0 && (
          <div className="mt-6 p-4 bg-background rounded-lg">
            <h3 className="text-sm font-medium text-foreground mb-2">Command Arguments</h3>
            <code className="text-sm text-muted-foreground">
              {session.command_args.join(' ')}
            </code>
          </div>
        )}
      </div>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-foreground">Request Timeline</h2>
        
        <div className="space-y-4">
          {session.llm_calls.map((call, index) => (
            <LLMCallCompact key={call.id} call={call} index={index} />
          ))}
          
          {session.tool_calls.map((call, index) => (
            <ToolCallCompact key={call.id} call={call} index={index} />
          ))}
          
          {session.http_calls.map((call, index) => (
            <HTTPCallCompact key={call.id} call={call} index={index} />
          ))}
        </div>
      </div>
    </div>
  )
}