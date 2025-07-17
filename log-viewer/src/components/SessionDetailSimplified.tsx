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
  Hash
} from 'lucide-react'
import { apiService } from '@/services/api'
import clsx from 'clsx'

interface CollapsibleProps {
  title: string
  icon: React.ReactNode
  children: React.ReactNode
  bgColor?: string
}

function Collapsible({ title, icon, children, bgColor = 'bg-gray-50' }: CollapsibleProps) {
  const [isOpen, setIsOpen] = useState(false)
  
  return (
    <div className={clsx("rounded-lg border overflow-hidden", bgColor)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full p-4 text-left hover:bg-black/5 transition-colors"
      >
        <div className="flex items-center gap-3">
          {isOpen ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
          {icon}
          <span className="font-medium">{title}</span>
        </div>
      </button>
      
      {isOpen && (
        <div className="border-t p-4 bg-white">
          {children}
        </div>
      )}
    </div>
  )
}

interface MessageProps {
  message: any
}

function Message({ message }: MessageProps) {
  const [isOpen, setIsOpen] = useState(false)
  const role = message.role || 'unknown'
  const content = message.content || ''
  
  const roleColors = {
    system: 'bg-purple-100 text-purple-800',
    user: 'bg-blue-100 text-blue-800',
    assistant: 'bg-green-100 text-green-800',
    tool: 'bg-orange-100 text-orange-800',
  }
  
  const roleIcons = {
    system: Settings,
    user: User,
    assistant: Bot,
    tool: Wrench,
  }
  
  const Icon = roleIcons[role as keyof typeof roleIcons] || Bot
  const colorClass = roleColors[role as keyof typeof roleColors] || 'bg-gray-100 text-gray-800'
  
  return (
    <div className={clsx("rounded border p-3", colorClass)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full text-left"
      >
        <div className="flex items-center gap-2">
          {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          <Icon className="h-4 w-4" />
          <span className="font-medium capitalize">{role}</span>
          {!isOpen && content && (
            <span className="text-sm truncate">
              {content.slice(0, 50)}...
            </span>
          )}
        </div>
      </button>
      
      {isOpen && content && (
        <div className="mt-3 pt-3 border-t">
          <pre className="text-sm whitespace-pre-wrap">{content}</pre>
        </div>
      )}
    </div>
  )
}

export function SessionDetailSimplified() {
  const { sessionId } = useParams<{ sessionId: string }>()
  
  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', sessionId],
    queryFn: () => apiService.getSession(sessionId!),
    enabled: !!sessionId,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  if (error || !session) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="text-center">
          <AlertTriangle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">Failed to load session</h3>
          <p className="text-gray-600">Session {sessionId} could not be found</p>
        </div>
      </div>
    )
  }

  const totalCost = session.llm_calls.reduce((sum, call) => sum + (call.cost || 0), 0)
  const hasErrors = session.llm_calls.some(call => call.error) || 
                   session.tool_calls.some(call => call.error) ||
                   session.http_calls.some(call => call.error)

  const duration = session.end_time 
    ? Math.round((new Date(session.end_time).getTime() - new Date(session.start_time).getTime()) / 1000)
    : 'Ongoing'

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Header */}
      <div className="bg-white rounded-lg border p-6">
        <div className="flex justify-between items-start mb-6">
          <div>
            <h1 className="text-2xl font-bold mb-2">Session {sessionId}</h1>
            <p className="text-gray-600">
              Started {format(new Date(session.start_time), 'PPpp')}
            </p>
          </div>
          
          <div className="flex items-center gap-2">
            {hasErrors ? (
              <div className="flex items-center gap-1 text-red-600">
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

        {/* Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
          <div className="text-center p-3 bg-gray-50 rounded">
            <Clock className="h-5 w-5 text-gray-600 mx-auto mb-1" />
            <div className="font-bold">{duration}</div>
            <div className="text-xs text-gray-600">Duration</div>
          </div>
          
          <div className="text-center p-3 bg-gray-50 rounded">
            <Activity className="h-5 w-5 text-gray-600 mx-auto mb-1" />
            <div className="font-bold">{session.llm_calls.length}</div>
            <div className="text-xs text-gray-600">LLM Calls</div>
          </div>
          
          <div className="text-center p-3 bg-gray-50 rounded">
            <Zap className="h-5 w-5 text-gray-600 mx-auto mb-1" />
            <div className="font-bold">{session.tool_calls.length}</div>
            <div className="text-xs text-gray-600">Tool Calls</div>
          </div>
          
          <div className="text-center p-3 bg-gray-50 rounded">
            <Globe className="h-5 w-5 text-gray-600 mx-auto mb-1" />
            <div className="font-bold">{session.http_calls.length}</div>
            <div className="text-xs text-gray-600">HTTP Calls</div>
          </div>
          
          <div className="text-center p-3 bg-gray-50 rounded">
            <DollarSign className="h-5 w-5 text-gray-600 mx-auto mb-1" />
            <div className="font-bold">${totalCost.toFixed(4)}</div>
            <div className="text-xs text-gray-600">Total Cost</div>
          </div>
        </div>
      </div>

      {/* Timeline */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Request Timeline</h2>
        
        {/* LLM Calls */}
        {session.llm_calls.map((call, index) => (
          <Collapsible
            key={call.id}
            title={`LLM Call ${index + 1} - ${call.provider} (${call.model})`}
            icon={<Activity className="h-5 w-5 text-blue-500" />}
            bgColor="bg-blue-50"
          >
            <div className="space-y-4">
              <div className="flex gap-4 text-sm text-gray-600">
                {call.tokens_used && (
                  <div className="flex items-center gap-1">
                    <Hash className="h-4 w-4" />
                    {call.tokens_used.total_tokens} tokens
                  </div>
                )}
                {call.cost && (
                  <div className="flex items-center gap-1">
                    <DollarSign className="h-4 w-4" />
                    ${call.cost.toFixed(4)}
                  </div>
                )}
                <div>{call.duration_ms}ms</div>
              </div>
              
              {call.error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded">
                  <div className="text-red-700 font-medium">Error:</div>
                  <pre className="text-sm text-red-600 mt-1">{call.error}</pre>
                </div>
              )}
              
              <div>
                <h4 className="font-medium mb-2">Request Messages</h4>
                <div className="space-y-2">
                  {(call.request?.messages || []).map((msg: any, msgIndex: number) => (
                    <Message key={msgIndex} message={msg} />
                  ))}
                </div>
              </div>
              
              {call.response && (
                <div>
                  <h4 className="font-medium mb-2">Response</h4>
                  <Message message={call.response} />
                </div>
              )}
            </div>
          </Collapsible>
        ))}
        
        {/* Tool Calls */}
        {session.tool_calls.map((call, index) => (
          <Collapsible
            key={call.id}
            title={`Tool Call ${index + 1} - ${call.name}`}
            icon={<Zap className="h-5 w-5 text-orange-500" />}
            bgColor="bg-orange-50"
          >
            <div className="space-y-4">
              <div className="text-sm text-gray-600">{call.duration_ms}ms</div>
              
              {call.error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded">
                  <div className="text-red-700 font-medium">Error:</div>
                  <pre className="text-sm text-red-600 mt-1">{call.error}</pre>
                </div>
              )}
              
              <div>
                <h4 className="font-medium mb-2">Input</h4>
                <pre className="text-sm bg-gray-50 p-3 rounded overflow-x-auto">
                  {JSON.stringify(call.input, null, 2)}
                </pre>
              </div>
              
              {call.output && (
                <div>
                  <h4 className="font-medium mb-2">Output</h4>
                  <pre className="text-sm bg-gray-50 p-3 rounded overflow-x-auto">
                    {typeof call.output === 'string' ? call.output : JSON.stringify(call.output, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </Collapsible>
        ))}
        
        {/* HTTP Calls */}
        {session.http_calls.map((call, index) => (
          <Collapsible
            key={call.id}
            title={`HTTP ${call.method} ${call.url}`}
            icon={<Globe className="h-5 w-5 text-purple-500" />}
            bgColor="bg-purple-50"
          >
            <div className="space-y-4">
              <div className="flex gap-4 text-sm text-gray-600">
                {call.status_code && (
                  <div className={clsx("font-medium", {
                    "text-green-600": call.status_code >= 200 && call.status_code < 300,
                    "text-red-600": call.status_code >= 400,
                    "text-gray-600": call.status_code < 200 || (call.status_code >= 300 && call.status_code < 400)
                  })}>
                    Status: {call.status_code}
                  </div>
                )}
                <div>{call.duration_ms}ms</div>
              </div>
              
              {call.error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded">
                  <div className="text-red-700 font-medium">Error:</div>
                  <pre className="text-sm text-red-600 mt-1">{call.error}</pre>
                </div>
              )}
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <h4 className="font-medium mb-2">Request</h4>
                  <div className="space-y-2">
                    {call.headers && Object.keys(call.headers).length > 0 && (
                      <div>
                        <div className="text-sm font-medium text-gray-600">Headers:</div>
                        <pre className="text-xs bg-gray-50 p-2 rounded overflow-x-auto">
                          {JSON.stringify(call.headers, null, 2)}
                        </pre>
                      </div>
                    )}
                    {call.body && (
                      <div>
                        <div className="text-sm font-medium text-gray-600">Body:</div>
                        <pre className="text-xs bg-gray-50 p-2 rounded overflow-x-auto">
                          {typeof call.body === 'string' ? call.body : JSON.stringify(call.body, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
                
                <div>
                  <h4 className="font-medium mb-2">Response</h4>
                  <div className="space-y-2">
                    {call.response_headers && Object.keys(call.response_headers).length > 0 && (
                      <div>
                        <div className="text-sm font-medium text-gray-600">Headers:</div>
                        <pre className="text-xs bg-gray-50 p-2 rounded overflow-x-auto">
                          {JSON.stringify(call.response_headers, null, 2)}
                        </pre>
                      </div>
                    )}
                    {call.response_body && (
                      <div>
                        <div className="text-sm font-medium text-gray-600">Body:</div>
                        <pre className="text-xs bg-gray-50 p-2 rounded overflow-x-auto max-h-48">
                          {typeof call.response_body === 'string' ? call.response_body : JSON.stringify(call.response_body, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </Collapsible>
        ))}
      </div>
    </div>
  )
}