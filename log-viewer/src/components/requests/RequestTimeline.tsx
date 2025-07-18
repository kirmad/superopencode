'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Clock, AlertTriangle, CheckCircle, XCircle, Zap } from 'lucide-react'

interface RequestTimelineProps {
  sessionId: string
  onRequestSelect?: (requestId: string) => void
  selectedRequestId?: string
}

export function RequestTimeline({ sessionId, onRequestSelect, selectedRequestId }: RequestTimelineProps) {
  const [providerFilter, setProviderFilter] = useState('all')
  
  const { data: session, isLoading } = useQuery({
    queryKey: ['session', sessionId],
    queryFn: async () => {
      const response = await fetch(`/api/sessions/${sessionId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch session')
      }
      return response.json()
    },
    enabled: !!sessionId
  })
  
  const requests = session?.httpCalls || []
  
  const filteredRequests = requests.filter((request: Record<string, unknown>) => {
    const provider = getProviderFromUrl(request.url as string)
    if (providerFilter !== 'all' && provider !== providerFilter) {
      return false
    }
    return true
  })
  
  // Sort by startTime
  const sortedRequests = [...filteredRequests].sort((a, b) => {
    const timeA = new Date(a.startTime as string).getTime()
    const timeB = new Date(b.startTime as string).getTime()
    return timeA - timeB
  })
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      {/* Filter Header */}
      <div className="p-2 border-b">
        <select
          value={providerFilter}
          onChange={(e) => setProviderFilter(e.target.value)}
          className="w-full px-2 py-1 border border-gray-300 rounded text-xs focus:ring-1 focus:ring-blue-500 focus:border-transparent"
        >
          <option value="all">All Providers</option>
          <option value="copilot">GitHub Copilot</option>
          <option value="openai">OpenAI</option>
          <option value="anthropic">Anthropic</option>
        </select>
      </div>
      
      {/* Timeline */}
      <div className="flex-1 overflow-y-auto">
        {sortedRequests.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <div className="text-2xl mb-2">📭</div>
            <p className="text-xs">No requests found</p>
          </div>
        ) : (
          <div className="relative">
            {/* Timeline line */}
            <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-gray-200"></div>
            
            {sortedRequests.map((request: Record<string, unknown>, index: number) => {
              const isSelected = request.id === selectedRequestId
              const isError = request.error || (request.statusCode as number) >= 400
              const isSuccess = (request.statusCode as number) >= 200 && (request.statusCode as number) < 300
              const provider = getProviderFromUrl(request.url as string)
              const time = new Date(request.startTime as string)
              const previousRequest = index > 0 ? sortedRequests[index - 1] : null
              const hasNewInteraction = checkForNewInteraction(request, previousRequest)
              
              return (
                <div
                  key={request.id as string}
                  className={`relative pl-12 pr-3 py-2 cursor-pointer transition-all duration-200 ${
                    isSelected 
                      ? 'bg-blue-50 border-r-2 border-blue-500' 
                      : 'hover:bg-gray-50'
                  } ${hasNewInteraction ? 'bg-yellow-50' : ''}`}
                  onClick={() => onRequestSelect?.(request.id as string)}
                >
                  {/* Timeline dot */}
                  <div className={`absolute left-5 top-3 w-2 h-2 rounded-full border-2 ${
                    isSelected 
                      ? 'bg-blue-500 border-blue-500' 
                      : isError 
                        ? 'bg-red-500 border-red-500'
                        : isSuccess
                          ? 'bg-green-500 border-green-500'
                          : 'bg-gray-400 border-gray-400'
                  }`}></div>
                  
                  {/* New interaction indicator */}
                  {hasNewInteraction && (
                    <div className="absolute left-2 top-1 w-1 h-1 rounded-full bg-yellow-500 animate-pulse"></div>
                  )}
                  
                  {/* Content */}
                  <div className="space-y-1">
                    {/* Time and Status */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1 text-xs text-gray-500">
                        <Clock className="h-3 w-3" />
                        <span>{time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        {isError ? (
                          <XCircle className="h-3 w-3 text-red-500" />
                        ) : isSuccess ? (
                          <CheckCircle className="h-3 w-3 text-green-500" />
                        ) : (
                          <AlertTriangle className="h-3 w-3 text-yellow-500" />
                        )}
                        <span className={`text-xs font-mono ${
                          isError ? 'text-red-600' : isSuccess ? 'text-green-600' : 'text-yellow-600'
                        }`}>
                          {request.statusCode ? String(request.statusCode) : '—'}
                        </span>
                      </div>
                    </div>
                    
                    {/* Provider and Method */}
                    <div className="flex items-center justify-between">
                      <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${
                        provider === 'copilot' ? 'bg-purple-100 text-purple-700' :
                        provider === 'openai' ? 'bg-green-100 text-green-700' :
                        provider === 'anthropic' ? 'bg-orange-100 text-orange-700' :
                        'bg-gray-100 text-gray-700'
                      }`}>
                        {provider === 'copilot' ? 'Copilot' :
                         provider === 'openai' ? 'OpenAI' :
                         provider === 'anthropic' ? 'Anthropic' :
                         'Unknown'}
                      </span>
                      <span className="text-xs font-mono text-gray-600">{String(request.method)}</span>
                    </div>
                    
                    {/* Duration */}
                    <div className="flex items-center gap-1 text-xs text-gray-500">
                      <Zap className="h-3 w-3" />
                      <span>{request.durationMs ? `${request.durationMs}ms` : '—'}</span>
                    </div>
                    
                    {/* Model if available */}
                    {(request.body as any)?.model && (
                      <div className="text-xs text-gray-400 truncate">
                        {(request.body as any).model}
                      </div>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

function getProviderFromUrl(url: string): string {
  if (url.includes('githubcopilot.com') || url.includes('github.com')) return 'copilot'
  if (url.includes('openai.com')) return 'openai'
  if (url.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}

function checkForNewInteraction(current: Record<string, unknown>, previous: Record<string, unknown> | null): boolean {
  if (!previous) return false
  
  // Check if there are new messages or different content
  const currentMessages = (current.messages as any[]) || []
  const previousMessages = (previous.messages as any[]) || []
  
  // Simple heuristic: if current request has more messages than previous, it's likely a new interaction
  if (currentMessages.length > previousMessages.length) {
    return true
  }
  
  // Check if the last message content is different
  if (currentMessages.length > 0 && previousMessages.length > 0) {
    const currentLastMessage = currentMessages[currentMessages.length - 1]
    const previousLastMessage = previousMessages[previousMessages.length - 1]
    
    return JSON.stringify(currentLastMessage) !== JSON.stringify(previousLastMessage)
  }
  
  return false
}