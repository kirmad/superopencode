'use client'

import { HTTPLog } from '@/lib/database'
import { Clock, AlertCircle, CheckCircle } from 'lucide-react'

interface RequestCardProps {
  request: HTTPLog
  isSelected?: boolean
  onClick?: () => void
}

export function RequestCard({ request, isSelected, onClick }: RequestCardProps) {
  const isSuccess = request.statusCode && request.statusCode >= 200 && request.statusCode < 300
  const hasError = request.error || (request.statusCode && request.statusCode >= 400)
  
  const provider = getProviderFromUrl(request.url)
  const providerName = getProviderName(provider)
  
  return (
    <div
      onClick={onClick}
      data-testid="request-item"
      className={`p-4 border-b cursor-pointer transition-colors ${
        isSelected 
          ? 'bg-blue-50 border-blue-200' 
          : 'hover:bg-gray-50 border-gray-200'
      }`}
    >
      <div className="flex justify-between items-start mb-2">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
              {request.method}
            </span>
            <span className="text-sm text-gray-600 truncate">
              {providerName}
            </span>
          </div>
          
          <div className="flex items-center gap-2 text-xs text-gray-600">
            <Clock className="h-3 w-3" />
            <span>{new Date(request.startTime).toLocaleString()}</span>
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          {request.statusCode && (
            <span className={`px-2 py-1 rounded text-xs font-medium ${
              isSuccess
                ? 'bg-green-100 text-green-800'
                : hasError
                ? 'bg-red-100 text-red-800'
                : 'bg-yellow-100 text-yellow-800'
            }`}>
              {request.statusCode}
            </span>
          )}
          
          <span className="text-xs text-gray-500">
            {request.durationMs}ms
          </span>
          
          {hasError && <AlertCircle className="h-4 w-4 text-red-500" />}
          {isSuccess && <CheckCircle className="h-4 w-4 text-green-500" />}
        </div>
      </div>
      
      {request.error && (
        <div className="text-xs text-red-600 mt-1 p-2 bg-red-50 rounded">
          {request.error}
        </div>
      )}
    </div>
  )
}

function getProviderFromUrl(url: string): string {
  if (url.includes('githubcopilot.com')) return 'copilot'
  if (url.includes('openai.com')) return 'openai'
  if (url.includes('anthropic.com')) return 'anthropic'
  return 'unknown'
}

function getProviderName(provider: string): string {
  switch (provider) {
    case 'copilot': return 'GitHub Copilot'
    case 'openai': return 'OpenAI'
    case 'anthropic': return 'Anthropic'
    default: return 'Unknown Provider'
  }
}