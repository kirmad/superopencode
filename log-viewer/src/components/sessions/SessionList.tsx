'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { SessionCard } from './SessionCard'
import { Search } from 'lucide-react'
import { SessionMetadata } from '@/lib/database'

interface SessionListProps {
  onSessionSelect?: (sessionId: string) => void
  selectedSessionId?: string
}

export function SessionList({ onSessionSelect, selectedSessionId }: SessionListProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [filterError, setFilterError] = useState<boolean | undefined>(undefined)
  
  const { data: sessions, isLoading, error } = useQuery({
    queryKey: ['sessions', { hasError: filterError }],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filterError !== undefined) {
        params.append('hasError', filterError.toString())
      }
      
      const response = await fetch(`/api/sessions?${params}`)
      if (!response.ok) {
        throw new Error('Failed to fetch sessions')
      }
      return response.json()
    }
  })
  
  const filteredSessions = sessions?.filter((session: Record<string, unknown>) =>
    (session.session_id as string).toLowerCase().includes(searchTerm.toLowerCase())
  ) || []
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }
  
  if (error) {
    return (
      <div className="text-center py-8 text-red-600">
        Error loading sessions: {(error as Error).message}
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b bg-white">
        <h2 className="text-lg font-semibold mb-2">Sessions</h2>
        
        <div className="relative mb-3">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search sessions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-8 pr-3 py-1.5 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
        
        <div className="flex gap-2">
          <select
            value={filterError?.toString() || ''}
            onChange={(e) => setFilterError(
              e.target.value === '' ? undefined : e.target.value === 'true'
            )}
            className="px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="">All Sessions</option>
            <option value="false">Success Only</option>
            <option value="true">With Errors</option>
          </select>
        </div>
      </div>
      
      <div className="flex-1 overflow-y-auto">
        {filteredSessions.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No sessions found
          </div>
        ) : (
          filteredSessions.map((session: Record<string, unknown>) => (
            <SessionCard
              key={session.id as string}
              session={session as unknown as SessionMetadata}
              isSelected={session.session_id === selectedSessionId}
              onClick={() => onSessionSelect?.(session.session_id as string)}
            />
          ))
        )}
      </div>
    </div>
  )
}