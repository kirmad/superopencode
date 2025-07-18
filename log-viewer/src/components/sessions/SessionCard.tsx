'use client'

import { SessionMetadata } from '@/lib/database'
import { Clock, AlertCircle } from 'lucide-react'

interface SessionCardProps {
  session: SessionMetadata
  isSelected?: boolean
  onClick?: () => void
}

export function SessionCard({ session, isSelected, onClick }: SessionCardProps) {
  const startTime = new Date(session.start_time).toLocaleString()
  
  return (
    <div
      onClick={onClick}
      data-testid="session-item"
      className={`p-3 border-b cursor-pointer transition-colors ${
        isSelected 
          ? 'bg-blue-50 border-blue-200' 
          : 'hover:bg-gray-50 border-gray-200'
      }`}
    >
      <div className="flex justify-between items-start mb-2">
        <div className="flex-1">
          <h3 className="font-medium text-xs text-gray-900 truncate">
            {session.session_id}
          </h3>
          <div className="flex items-center gap-1 mt-1 text-xs text-gray-600">
            <Clock className="h-3 w-3" />
            <span>{startTime}</span>
          </div>
        </div>
        
        {session.has_error && (
          <AlertCircle className="h-4 w-4 text-red-500 flex-shrink-0" />
        )}
      </div>
      
      <div className="flex gap-2 text-xs">
        <span className="bg-blue-100 text-blue-800 px-1.5 py-0.5 rounded text-xs">
          {session.llm_call_count} LLM
        </span>
        <span className="bg-green-100 text-green-800 px-1.5 py-0.5 rounded text-xs">
          {session.http_call_count} HTTP
        </span>
        <span className="bg-purple-100 text-purple-800 px-1.5 py-0.5 rounded text-xs">
          {session.tool_call_count} Tools
        </span>
      </div>
      
      <div className="flex justify-between items-center mt-1.5 text-xs text-gray-600">
        <span>{session.total_tokens.toLocaleString()} tokens</span>
        <span>${session.total_cost.toFixed(4)}</span>
      </div>
    </div>
  )
}