'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { RequestCard } from './RequestCard'
import { HTTPLog } from '@/lib/database'

interface RequestListProps {
  sessionId: string
  onRequestSelect?: (requestId: string) => void
  selectedRequestId?: string
}

export function RequestList({ sessionId, onRequestSelect, selectedRequestId }: RequestListProps) {
  const [providerFilter, setProviderFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState('all')
  
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
    const hasError = request.error || (request.statusCode as number) >= 400
    
    if (providerFilter !== 'all' && provider !== providerFilter) {
      return false
    }
    
    if (statusFilter === 'success' && hasError) {
      return false
    }
    
    if (statusFilter === 'error' && !hasError) {
      return false
    }
    
    return true
  })
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b bg-white">
        <h3 className="text-lg font-semibold mb-2">HTTP Requests</h3>
        
        <div className="flex gap-2">
          <select
            value={providerFilter}
            onChange={(e) => setProviderFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
          >
            <option value="all">All Providers</option>
            <option value="copilot">GitHub Copilot</option>
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
          </select>
          
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
          >
            <option value="all">All Status</option>
            <option value="success">Success</option>
            <option value="error">Error</option>
          </select>
        </div>
      </div>
      
      <div className="flex-1 overflow-y-auto">
        {filteredRequests.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No requests found
          </div>
        ) : (
          filteredRequests.map((request: Record<string, unknown>) => (
            <RequestCard
              key={request.id as string}
              request={request as unknown as HTTPLog}
              isSelected={request.id === selectedRequestId}
              onClick={() => onRequestSelect?.(request.id as string)}
            />
          ))
        )}
      </div>
    </div>
  )
}

function getProviderFromUrl(url: string): string {
  if (url.includes('githubcopilot.com')) return 'copilot'
  if (url.includes('openai.com')) return 'openai'
  if (url.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}